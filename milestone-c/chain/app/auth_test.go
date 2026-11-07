// Transaction authentication tests for the Milestone C wired application.
//
// Every scenario is exercised on BOTH authentication paths, because a
// transaction must not pass one path and bypass the other:
//   - CheckTx (mempool admission, ante handler only, no commit), and
//   - FinalizeBlock (block execution, ante handler plus message
//     execution, followed by Commit).
//
// All values are TEST-ONLY. Each test builds a fresh app from testApp
// (app_test.go): alice is funded 1,000,000 uogo, account number 1,
// sequence 0, chain ID "ogo-local-1", monetary genesis strict
// M = G + S_actual = 1,003,439.
//
// Height-0 CheckTx quirk (see TestSignatureVerification in
// app_test.go): BaseApp.CheckTx runs at height 0, where the ante
// handler verifies with account number 0 regardless of the stored
// number (1). A live chain verifies CheckTx at the current height, so
// the stored number applies there. Tests that need the stored account
// number on the CheckTx path therefore first commit an empty block at
// height 1 (advanceEmptyBlock) and sign with account number 1.
// FinalizeBlock always verifies against the stored account number.
//
// All blocks are finalised at testGenesisTime so no monetary epoch
// ends during these tests and no issuance changes the supply.
//
// Two standard SDK semantics shape these tests, and neither is an
// OGO gap:
//   - CheckTx never executes messages (baseapp runMsgs: "Messages
//     will only be executed during simulation and DeliverTx"). A
//     transaction with a valid signature but an unaffordable message
//     is therefore admitted by CheckTx and rejected at block
//     execution. That is mempool admission, not a bypass: the
//     transaction can never execute or change state. The
//     insufficient-funds case documents this split explicitly.
//   - When the ante handler succeeds but message execution fails,
//     the ante handler's writes (fee deduction, sequence increment)
//     still commit (baseapp runTx writes the ante branch before
//     running messages). A failed-execution transaction therefore
//     pays its fee and consumes its sequence, but moves no other
//     funds and changes neither total supply (the fee is a transfer
//     to the fee collector, not a mint or burn) nor total issued.
//     The insufficient-funds FinalizeBlock leg asserts exactly that.
package app

import (
	"context"
	"strings"
	"testing"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	monetary "github.com/benross-lab/ogo-network/x/monetary"
)

// testBobPriv is a second TEST-ONLY key, used as the recipient and as
// the "other key" in forgery and mismatch cases.
var testBobPriv = secp256k1.GenPrivKey()

func testBobAddr() sdk.AccAddress {
	return sdk.AccAddress(testBobPriv.PubKey().Address())
}

// txOpts describes one transaction to build for the auth tests.
// Zero values select the valid-transaction defaults.
type txOpts struct {
	signer     *secp256k1.PrivKey // key that produces the signature
	pubKey     cryptotypes.PubKey // pubkey placed in the signer info
	sender     sdk.AccAddress     // MsgSend sender (and signer address)
	recipient  sdk.AccAddress
	acctNum    uint64
	seq        uint64
	chainID    string
	amount     int64
	feeAmt     int64
	feeDenom   string
	zeroFee    bool // fee of exactly zero (feeAmt 0 means the default 100)
	gasLimit   uint64
	noSig      bool // build with no signatures at all
	tamperMsg  bool // change the message after signing
	corruptSig bool // flip a byte of the signature after signing
}

func (o *txOpts) defaults() {
	if o.signer == nil {
		o.signer = testAlicePriv
	}
	if o.pubKey == nil {
		o.pubKey = o.signer.PubKey()
	}
	if o.sender == nil {
		o.sender = testAliceAddr()
	}
	if o.recipient == nil {
		o.recipient = testBobAddr()
	}
	if o.chainID == "" {
		o.chainID = "ogo-local-1"
	}
	if o.amount == 0 {
		o.amount = 1000
	}
	if o.feeAmt == 0 && !o.zeroFee {
		o.feeAmt = 100
	}
	if o.feeDenom == "" {
		o.feeDenom = monetary.Denom
	}
	if o.gasLimit == 0 {
		o.gasLimit = 200_000
	}
}

// buildTx signs (unless noSig) and encodes a MsgSend transaction.
// acctNum and seq are the values the signature commits to; they need
// not match the stored account, which is exactly what the wrong
// account number and replay cases exercise.
func buildTx(t *testing.T, app *OgoApp, o txOpts) []byte {
	t.Helper()
	o.defaults()
	txConfig := authtx.NewTxConfig(app.AppCodec(), authtx.DefaultSignModes)
	msg := banktypes.NewMsgSend(o.sender, o.recipient, sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(o.amount))))
	builder := txConfig.NewTxBuilder()
	if err := builder.SetMsgs(msg); err != nil {
		t.Fatalf("SetMsgs: %v", err)
	}
	builder.SetGasLimit(o.gasLimit)
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(o.feeDenom, math.NewInt(o.feeAmt))))
	if o.noSig {
		bz, err := txConfig.TxEncoder()(builder.GetTx())
		if err != nil {
			t.Fatalf("TxEncoder unsigned: %v", err)
		}
		return bz
	}
	sigV2 := signing.SignatureV2{
		PubKey:   o.pubKey,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT, Signature: nil},
		Sequence: o.seq,
	}
	if err := builder.SetSignatures(sigV2); err != nil {
		t.Fatalf("SetSignatures: %v", err)
	}
	signerData := authsigning.SignerData{
		Address:       o.sender.String(),
		ChainID:       o.chainID,
		AccountNumber: o.acctNum,
		Sequence:      o.seq,
		PubKey:        o.pubKey,
	}
	sigV2, err := clienttx.SignWithPrivKey(
		context.Background(), signing.SignMode_SIGN_MODE_DIRECT, signerData,
		builder, o.signer, txConfig, o.seq,
	)
	if err != nil {
		t.Fatalf("SignWithPrivKey: %v", err)
	}
	if o.corruptSig {
		data := sigV2.Data.(*signing.SingleSignatureData)
		if len(data.Signature) == 0 {
			t.Fatal("cannot corrupt an empty signature")
		}
		data.Signature[0] ^= 0xFF
	}
	if err := builder.SetSignatures(sigV2); err != nil {
		t.Fatalf("SetSignatures: %v", err)
	}
	if o.tamperMsg {
		// Change the message after signing: the signature must no
		// longer verify.
		msg2 := banktypes.NewMsgSend(o.sender, o.recipient, sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(999_999))))
		if err := builder.SetMsgs(msg2); err != nil {
			t.Fatalf("SetMsgs tamper: %v", err)
		}
	}
	bz, err := txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		t.Fatalf("TxEncoder: %v", err)
	}
	return bz
}

// checkTx runs one CheckTx and returns its code and log.
func checkTx(t *testing.T, app *OgoApp, bz []byte) (uint32, string) {
	t.Helper()
	res, err := app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	if err != nil {
		t.Fatalf("CheckTx: %v", err)
	}
	return res.Code, res.Log
}

// finalize runs FinalizeBlock at the given height with the given txs
// (all at testGenesisTime, so no epoch ends) and then commits. It
// returns one ExecTxResult per tx, in order.
func finalize(t *testing.T, app *OgoApp, height int64, txs [][]byte) []*abci.ExecTxResult {
	t.Helper()
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Txs:               txs,
		Height:            height,
		Time:              testGenesisTime,
		Hash:              []byte("auth-test-block"),
		DecidedLastCommit: abci.CommitInfo{Round: 0},
	})
	if err != nil {
		t.Fatalf("FinalizeBlock height %d: %v", height, err)
	}
	if len(res.TxResults) != len(txs) {
		t.Fatalf("FinalizeBlock height %d returned %d tx results for %d txs", height, len(res.TxResults), len(txs))
	}
	if _, err := app.Commit(); err != nil {
		t.Fatalf("Commit height %d: %v", height, err)
	}
	return res.TxResults
}

// advanceEmptyBlock finalises and commits an empty block at the given
// height, so later CheckTx calls run at a non-zero height and verify
// against the stored account number.
func advanceEmptyBlock(t *testing.T, app *OgoApp, height int64) {
	t.Helper()
	finalize(t, app, height, nil)
}

// stateSnapshot is the observable monetary and account state used to
// prove a rejected transaction changed nothing, and a valid one
// changed exactly what it should.
type stateSnapshot struct {
	aliceBal  math.Int
	bobBal    math.Int
	supply    math.Int
	issued    math.Int
	aliceSeq  uint64
	aliceAcct uint64
}

func snapshot(t *testing.T, app *OgoApp, height int64) stateSnapshot {
	t.Helper()
	ctx := ctxAt(app, height, testGenesisTime)
	acc := app.AccountKeeper.GetAccount(ctx, testAliceAddr())
	if acc == nil {
		t.Fatal("alice account missing")
	}
	issued, err := app.MonetaryKeeper.GetTotalIssued(ctx)
	if err != nil {
		t.Fatalf("GetTotalIssued: %v", err)
	}
	return stateSnapshot{
		aliceBal:  app.BankKeeper.GetBalance(ctx, testAliceAddr(), monetary.Denom).Amount,
		bobBal:    app.BankKeeper.GetBalance(ctx, testBobAddr(), monetary.Denom).Amount,
		supply:    app.BankKeeper.GetSupply(ctx, monetary.Denom).Amount,
		issued:    issued,
		aliceSeq:  acc.GetSequence(),
		aliceAcct: acc.GetAccountNumber(),
	}
}

// assertUnchanged fails if any part of the snapshot moved.
func assertUnchanged(t *testing.T, what string, before, after stateSnapshot) {
	t.Helper()
	if !before.aliceBal.Equal(after.aliceBal) {
		t.Errorf("%s: alice balance changed %s -> %s", what, before.aliceBal, after.aliceBal)
	}
	if !before.bobBal.Equal(after.bobBal) {
		t.Errorf("%s: bob balance changed %s -> %s", what, before.bobBal, after.bobBal)
	}
	if !before.supply.Equal(after.supply) {
		t.Errorf("%s: total supply changed %s -> %s", what, before.supply, after.supply)
	}
	if !before.issued.Equal(after.issued) {
		t.Errorf("%s: total issued changed %s -> %s", what, before.issued, after.issued)
	}
	if before.aliceSeq != after.aliceSeq {
		t.Errorf("%s: alice sequence changed %d -> %d", what, before.aliceSeq, after.aliceSeq)
	}
}

// expectRejected is the shared assertion for a rejected transaction:
// non-zero code on the path that ran, and no state change.
func expectRejected(t *testing.T, what string, code uint32, logMsg string, before, after stateSnapshot) {
	t.Helper()
	if code == 0 {
		t.Fatalf("%s: transaction was accepted, must be rejected", what)
	}
	t.Logf("%s: rejected as expected (code %d): %s", what, code, logMsg)
	assertUnchanged(t, what, before, after)
}

// ---------------------------------------------------------------------------
// Forged sender
// ---------------------------------------------------------------------------

// The message claims alice sends, but the transaction is signed by a
// different key (bob's). The signer address derived from that key has
// no account, so the ante handler rejects it before execution.
func TestAuthForgedSenderRejected(t *testing.T) {
	// CheckTx path (height 0: sign with account number 0).
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{signer: testBobPriv, acctNum: 0, seq: 0})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "forged sender (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// FinalizeBlock path: the stored account number is 1.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{signer: testBobPriv, acctNum: 1, seq: 0})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "forged sender (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
}

// ---------------------------------------------------------------------------
// Unsigned transaction
// ---------------------------------------------------------------------------

func TestAuthUnsignedRejected(t *testing.T) {
	// CheckTx path.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{noSig: true})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "unsigned tx (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// FinalizeBlock path.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{noSig: true})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "unsigned tx (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
}

// ---------------------------------------------------------------------------
// Invalid signature
// ---------------------------------------------------------------------------

// Two distinct invalid-signature constructions, one per path, plus
// the tampered-message construction on both paths.
func TestAuthInvalidSignatureRejected(t *testing.T) {
	// CheckTx path: message changed after signing.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0, tamperMsg: true})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "tampered tx (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// CheckTx path: signature bytes corrupted after signing.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 0, seq: 0, corruptSig: true})
	code, logMsg = checkTx(t, app, bz)
	expectRejected(t, "corrupted signature (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// FinalizeBlock path: message changed after signing.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, tamperMsg: true})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "tampered tx (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))

	// FinalizeBlock path: signature bytes corrupted after signing.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, corruptSig: true})
	results = finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "corrupted signature (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
}

// ---------------------------------------------------------------------------
// Public key / address mismatch
// ---------------------------------------------------------------------------

// The signature is valid for alice's key, but the signer info in the
// transaction carries a different public key (bob's). The ante
// handler derives the signer address from the stated key, so the
// transaction presents as signed by an account that does not exist,
// and the signature does not verify against the stated key either.
func TestAuthPubkeyMismatchRejected(t *testing.T) {
	// CheckTx path.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{signer: testAlicePriv, pubKey: testBobPriv.PubKey(), acctNum: 0, seq: 0})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "pubkey mismatch (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// FinalizeBlock path.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{signer: testAlicePriv, pubKey: testBobPriv.PubKey(), acctNum: 1, seq: 0})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "pubkey mismatch (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
}

// ---------------------------------------------------------------------------
// Wrong chain ID
// ---------------------------------------------------------------------------

func TestAuthWrongChainIDRejected(t *testing.T) {
	// CheckTx path (height 0: account number 0, wrong chain ID).
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0, chainID: "ogo-local-999"})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "wrong chain ID (CheckTx)", code, logMsg, before, snapshot(t, app, 1))
	if !strings.Contains(logMsg, "chain-id") {
		t.Logf("note: wrong-chain-ID CheckTx rejection did not name the chain ID in its log: %s", logMsg)
	}

	// FinalizeBlock path.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, chainID: "ogo-local-999"})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "wrong chain ID (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
}

// ---------------------------------------------------------------------------
// Wrong account number
// ---------------------------------------------------------------------------

// The stored account number is 1. A transaction signed with account
// number 0 must be rejected wherever the stored number applies: on
// the FinalizeBlock path always, and on the CheckTx path only once
// the chain has a committed block (height >= 1). At height 0 CheckTx
// verifies with account number 0 by BaseApp design (see the package
// doc), so the height-0 case is asserted separately: signing with
// account number 1 at height 0 is rejected, which is the same
// verification, just with the height-0 expected number.
func TestAuthWrongAccountNumberRejected(t *testing.T) {
	// FinalizeBlock path: signed with account number 0, stored is 1.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "wrong account number (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
	if !strings.Contains(results[0].Log, "account number (1)") {
		t.Errorf("wrong-account-number rejection did not name the stored number: %s", results[0].Log)
	}

	// CheckTx path at height >= 1: commit an empty block first, then
	// a transaction signed with account number 0 must be rejected.
	app, _ = testApp(t)
	advanceEmptyBlock(t, app, 1)
	before = snapshot(t, app, 2)
	bz = buildTx(t, app, txOpts{acctNum: 0, seq: 0})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "wrong account number (CheckTx height 1)", code, logMsg, before, snapshot(t, app, 2))
	if !strings.Contains(logMsg, "account number (1)") {
		t.Errorf("wrong-account-number CheckTx rejection did not name the stored number: %s", logMsg)
	}

	// CheckTx path at height 0 verifies with account number 0, so
	// the mirror image (signed with 1) is rejected there.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0})
	code, logMsg = checkTx(t, app, bz)
	expectRejected(t, "account number 1 at CheckTx height 0", code, logMsg, before, snapshot(t, app, 1))
}

// ---------------------------------------------------------------------------
// Replayed sequence
// ---------------------------------------------------------------------------

// A valid transaction is delivered once; the identical bytes
// submitted again (same sequence) must be rejected on both paths.
func TestAuthReplayedSequenceRejected(t *testing.T) {
	// CheckTx path: deliver at height 1, then re-check the same bytes.
	app, _ := testApp(t)
	bz := buildTx(t, app, txOpts{acctNum: 1, seq: 0})
	results := finalize(t, app, 1, [][]byte{bz})
	if results[0].Code != 0 {
		t.Fatalf("setup delivery must succeed, got code %d: %s", results[0].Code, results[0].Log)
	}
	before := snapshot(t, app, 2)
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "replayed sequence (CheckTx)", code, logMsg, before, snapshot(t, app, 2))
	if !strings.Contains(logMsg, "account sequence mismatch") {
		t.Errorf("replay rejection did not report a sequence mismatch: %s", logMsg)
	}

	// FinalizeBlock path: the same bytes in the next block.
	before = snapshot(t, app, 2)
	results = finalize(t, app, 2, [][]byte{bz})
	expectRejected(t, "replayed sequence (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 3))
}

// ---------------------------------------------------------------------------
// Duplicate transaction
// ---------------------------------------------------------------------------

// The same transaction submitted twice must execute at most once:
// the second submission is rejected, and the sequence advances
// exactly once.
func TestAuthDuplicateTransactionRejected(t *testing.T) {
	// CheckTx path: the first check succeeds and consumes the
	// sequence in mempool state; the second check is rejected.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0})
	if code, logMsg := checkTx(t, app, bz); code != 0 {
		t.Fatalf("first CheckTx must succeed, got code %d: %s", code, logMsg)
	}
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "duplicate tx (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// FinalizeBlock path: the same bytes twice in one block. The
	// first executes, the second is rejected, and the sequence
	// advances exactly once.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0})
	results := finalize(t, app, 1, [][]byte{bz, bz})
	if results[0].Code != 0 {
		t.Fatalf("first copy in the block must execute, got code %d: %s", results[0].Code, results[0].Log)
	}
	if results[1].Code == 0 {
		t.Fatal("second copy of the same tx in one block must be rejected")
	}
	t.Logf("duplicate tx (FinalizeBlock): second copy rejected (code %d): %s", results[1].Code, results[1].Log)
	after := snapshot(t, app, 2)
	if after.aliceSeq != before.aliceSeq+1 {
		t.Errorf("sequence advanced by %d, want exactly 1", after.aliceSeq-before.aliceSeq)
	}
	if !after.aliceBal.Equal(before.aliceBal.Sub(math.NewInt(1100))) {
		t.Errorf("alice balance after one execution: got %s, want %s", after.aliceBal, before.aliceBal.Sub(math.NewInt(1100)))
	}
	if !after.supply.Equal(before.supply) {
		t.Errorf("supply changed by duplicate tx: %s -> %s", before.supply, after.supply)
	}
	if !after.issued.Equal(before.issued) {
		t.Errorf("total issued changed by duplicate tx: %s -> %s", before.issued, after.issued)
	}
}

// ---------------------------------------------------------------------------
// Valid signed transaction
// ---------------------------------------------------------------------------

func TestAuthValidTransactionAccepted(t *testing.T) {
	// CheckTx path: accepted, and it does not commit state.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0})
	if code, logMsg := checkTx(t, app, bz); code != 0 {
		t.Fatalf("valid tx rejected in CheckTx (code %d): %s", code, logMsg)
	}
	after := snapshot(t, app, 1)
	if after.aliceSeq != before.aliceSeq {
		t.Errorf("CheckTx changed the committed sequence: %d -> %d", before.aliceSeq, after.aliceSeq)
	}
	assertUnchanged(t, "valid tx (CheckTx)", before, after)

	// CheckTx path at height >= 1 with the stored account number.
	app, _ = testApp(t)
	advanceEmptyBlock(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0})
	if code, logMsg := checkTx(t, app, bz); code != 0 {
		t.Fatalf("valid tx rejected in CheckTx at height 1 (code %d): %s", code, logMsg)
	}

	// FinalizeBlock path: accepted, sequence advances exactly once,
	// balances move by amount plus fee, supply and issuance do not.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0})
	results := finalize(t, app, 1, [][]byte{bz})
	if results[0].Code != 0 {
		t.Fatalf("valid tx rejected in FinalizeBlock (code %d): %s", results[0].Code, results[0].Log)
	}
	after = snapshot(t, app, 2)
	if after.aliceSeq != before.aliceSeq+1 {
		t.Errorf("sequence advanced by %d, want exactly 1", after.aliceSeq-before.aliceSeq)
	}
	if !after.aliceBal.Equal(math.NewInt(998_900)) {
		t.Errorf("alice balance after valid tx: got %s, want 998900 (1000 sent + 100 fee)", after.aliceBal)
	}
	if !after.bobBal.Equal(math.NewInt(1000)) {
		t.Errorf("bob balance after valid tx: got %s, want 1000", after.bobBal)
	}
	if !after.supply.Equal(before.supply) {
		t.Errorf("supply changed by valid tx: %s -> %s", before.supply, after.supply)
	}
	if !after.issued.Equal(before.issued) {
		t.Errorf("total issued changed by valid tx: %s -> %s", before.issued, after.issued)
	}
}

// ---------------------------------------------------------------------------
// Insufficient funds
// ---------------------------------------------------------------------------

// The signature is valid, but the amount exceeds the balance. This is
// a message-execution failure, not an authentication failure, and the
// two paths treat it differently by standard SDK design (see the
// package doc): CheckTx never executes messages, so it admits the
// transaction; FinalizeBlock executes the message, rejects the
// transaction, and charges the fee and consumes the sequence exactly
// as the SDK specifies for a failed execution. Supply and total
// issued are unchanged on both paths, and the recipient gets nothing.
func TestAuthInsufficientFundsRejected(t *testing.T) {
	// CheckTx path: admitted, because CheckTx runs only the ante
	// handler and never executes the message. This is expected, not
	// a bypass: the transaction can never execute. The admission
	// consumes the sequence in mempool state, so repeating the same
	// transaction is rejected, and a valid transaction with the next
	// sequence still passes.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0, amount: 2_000_000})
	if code, logMsg := checkTx(t, app, bz); code != 0 {
		t.Fatalf("insufficient-funds tx must be admitted by CheckTx (no message execution there), got code %d: %s", code, logMsg)
	}
	t.Logf("insufficient funds (CheckTx): admitted as expected (ante only, no message execution)")
	assertUnchanged(t, "insufficient funds (CheckTx)", before, snapshot(t, app, 1))
	if code, logMsg := checkTx(t, app, bz); code == 0 {
		t.Fatal("repeat CheckTx of the admitted tx must be rejected (sequence consumed in mempool state)")
	} else {
		t.Logf("insufficient funds (CheckTx repeat): rejected (code %d): %s", code, logMsg)
	}
	if code, logMsg := checkTx(t, app, buildTx(t, app, txOpts{acctNum: 0, seq: 1})); code != 0 {
		t.Fatalf("valid tx with the next sequence must still pass CheckTx (code %d): %s", code, logMsg)
	}

	// FinalizeBlock path: rejected at message execution. The fee is
	// still charged and the sequence still consumed (standard SDK
	// failed-execution semantics); the amount never moves, the
	// supply never changes, and nothing is issued.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, amount: 2_000_000})
	results := finalize(t, app, 1, [][]byte{bz})
	if results[0].Code == 0 {
		t.Fatal("insufficient-funds tx must be rejected in FinalizeBlock")
	}
	t.Logf("insufficient funds (FinalizeBlock): rejected (code %d): %s", results[0].Code, results[0].Log)
	after := snapshot(t, app, 2)
	if !after.aliceBal.Equal(before.aliceBal.Sub(math.NewInt(100))) {
		t.Errorf("alice balance after failed execution: got %s, want exactly the 100 uogo fee deducted", after.aliceBal)
	}
	if after.aliceSeq != before.aliceSeq+1 {
		t.Errorf("sequence after failed execution: got %d, want exactly one consumed", after.aliceSeq)
	}
	if !after.bobBal.Equal(before.bobBal) {
		t.Errorf("recipient balance changed by failed execution: %s -> %s", before.bobBal, after.bobBal)
	}
	if !after.supply.Equal(before.supply) {
		t.Errorf("supply changed by failed execution: %s -> %s", before.supply, after.supply)
	}
	if !after.issued.Equal(before.issued) {
		t.Errorf("total issued changed by failed execution: %s -> %s", before.issued, after.issued)
	}
	ctx := ctxAt(app, 2, testGenesisTime)
	feeCollector := app.BankKeeper.GetBalance(ctx, authtypes.NewModuleAddress("fee_collector"), monetary.Denom)
	if !feeCollector.Amount.Equal(math.NewInt(100)) {
		t.Errorf("fee collector balance: got %s, want exactly the 100 uogo fee (transfer, not a mint or burn)", feeCollector.Amount)
	}
}

// ---------------------------------------------------------------------------
// Invalid fee
// ---------------------------------------------------------------------------

// Two invalid-fee constructions: a fee in a denomination the sender
// does not hold, and a fee larger than the sender's whole balance.
func TestAuthInvalidFeeRejected(t *testing.T) {
	// CheckTx path: unknown fee denomination.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0, feeDenom: "ustake", feeAmt: 100})
	code, logMsg := checkTx(t, app, bz)
	expectRejected(t, "unknown fee denom (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// CheckTx path: fee larger than the balance.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 0, seq: 0, feeAmt: 2_000_000})
	code, logMsg = checkTx(t, app, bz)
	expectRejected(t, "excessive fee (CheckTx)", code, logMsg, before, snapshot(t, app, 1))

	// FinalizeBlock path: unknown fee denomination.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, feeDenom: "ustake", feeAmt: 100})
	results := finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "unknown fee denom (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))

	// FinalizeBlock path: fee larger than the balance.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, feeAmt: 2_000_000})
	results = finalize(t, app, 1, [][]byte{bz})
	expectRejected(t, "excessive fee (FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
}

// ---------------------------------------------------------------------------
// Malformed transaction
// ---------------------------------------------------------------------------

func TestAuthMalformedTransactionRejected(t *testing.T) {
	malformed := map[string][]byte{
		"random bytes": {0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		"empty bytes":  {},
	}
	// Truncated form of a valid transaction: the first 10 bytes cut
	// off the embedded body mid-message, so decoding must fail.
	app, _ := testApp(t)
	valid := buildTx(t, app, txOpts{acctNum: 0, seq: 0})
	malformed["truncated tx"] = valid[:10]

	for name, bz := range malformed {
		// CheckTx path.
		app, _ := testApp(t)
		before := snapshot(t, app, 1)
		code, logMsg := checkTx(t, app, bz)
		expectRejected(t, "malformed ("+name+", CheckTx)", code, logMsg, before, snapshot(t, app, 1))

		// FinalizeBlock path: undecodable bytes get the default
		// ErrTxDecode result instead of being executed.
		app, _ = testApp(t)
		before = snapshot(t, app, 1)
		results := finalize(t, app, 1, [][]byte{bz})
		expectRejected(t, "malformed ("+name+", FinalizeBlock)", results[0].Code, results[0].Log, before, snapshot(t, app, 2))
	}
}

// ---------------------------------------------------------------------------
// Zero fee (documents current staging behaviour, not a vulnerability)
// ---------------------------------------------------------------------------

// No minimum gas price is set in the app code: the node takes it from
// its --minimum-gas-prices flag at runtime (cmd/ogod/cmd/root.go),
// and testApp sets none. A zero-fee transaction with a valid
// signature is therefore accepted on both paths in this test setup.
// This test pins that behaviour so a later minimum-gas-price change
// shows up as a failure here; it does not claim zero-fee acceptance
// is the right launch policy. Spam resistance on a public network
// depends on the launch node config setting a minimum gas price.
func TestAuthZeroFeeBehaviorDocumented(t *testing.T) {
	// CheckTx path.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 0, seq: 0, zeroFee: true})
	if code, logMsg := checkTx(t, app, bz); code != 0 {
		t.Fatalf("zero-fee valid tx rejected in CheckTx (code %d): %s", code, logMsg)
	}
	assertUnchanged(t, "zero-fee tx (CheckTx)", before, snapshot(t, app, 1))

	// FinalizeBlock path: accepted, fee deducted is zero.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz = buildTx(t, app, txOpts{acctNum: 1, seq: 0, zeroFee: true})
	results := finalize(t, app, 1, [][]byte{bz})
	if results[0].Code != 0 {
		t.Fatalf("zero-fee valid tx rejected in FinalizeBlock (code %d): %s", results[0].Code, results[0].Log)
	}
	after := snapshot(t, app, 2)
	if !after.aliceBal.Equal(math.NewInt(999_000)) {
		t.Errorf("alice balance after zero-fee tx: got %s, want 999000 (1000 sent, 0 fee)", after.aliceBal)
	}
	if after.aliceSeq != before.aliceSeq+1 {
		t.Errorf("sequence after zero-fee tx: got %d, want exactly one consumed", after.aliceSeq)
	}
	if !after.supply.Equal(before.supply) || !after.issued.Equal(before.issued) {
		t.Errorf("supply or issued changed by zero-fee tx: %+v -> %+v", before, after)
	}
}

// ---------------------------------------------------------------------------
// The two paths must agree
// ---------------------------------------------------------------------------

// TestAuthPathsAgree runs the same set of transactions through
// CheckTx (after an empty block, so the stored account number
// applies) and through FinalizeBlock, on separate fresh apps, and
// requires the accept/reject verdict to match the per-path
// expectation. Every authentication failure must be rejected on both
// paths: a transaction must never pass one path and bypass the other.
// The one documented split is insufficient funds, which is a
// message-execution failure, not an authentication failure: CheckTx
// never executes messages (baseapp runMsgs), so it admits the
// transaction, while FinalizeBlock rejects it. That split is asserted
// explicitly here, with the two different expectations.
func TestAuthPathsAgree(t *testing.T) {
	cases := []struct {
		name            string
		build           func(t *testing.T, app *OgoApp) []byte
		wantCheckAccept bool
		wantFinAccept   bool
	}{
		{"valid", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0})
		}, true, true},
		{"forged sender", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{signer: testBobPriv, acctNum: 1, seq: 0})
		}, false, false},
		{"unsigned", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{noSig: true})
		}, false, false},
		{"tampered message", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0, tamperMsg: true})
		}, false, false},
		{"corrupted signature", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0, corruptSig: true})
		}, false, false},
		{"pubkey mismatch", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{signer: testAlicePriv, pubKey: testBobPriv.PubKey(), acctNum: 1, seq: 0})
		}, false, false},
		{"wrong chain ID", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0, chainID: "ogo-local-999"})
		}, false, false},
		{"wrong account number", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 0, seq: 0})
		}, false, false},
		// Documented split: CheckTx admits (no message execution),
		// FinalizeBlock rejects (message execution fails).
		{"insufficient funds", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0, amount: 2_000_000})
		}, true, false},
		{"unknown fee denom", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0, feeDenom: "ustake", feeAmt: 100})
		}, false, false},
		{"excessive fee", func(t *testing.T, app *OgoApp) []byte {
			return buildTx(t, app, txOpts{acctNum: 1, seq: 0, feeAmt: 2_000_000})
		}, false, false},
		{"malformed random bytes", func(t *testing.T, app *OgoApp) []byte {
			return []byte{0xde, 0xad, 0xbe, 0xef}
		}, false, false},
		{"malformed empty bytes", func(t *testing.T, app *OgoApp) []byte {
			return []byte{}
		}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// CheckTx leg, at height 1 so the stored account
			// number applies.
			checkApp, _ := testApp(t)
			advanceEmptyBlock(t, checkApp, 1)
			checkBefore := snapshot(t, checkApp, 2)
			checkCode, checkLog := checkTx(t, checkApp, tc.build(t, checkApp))
			checkAccepted := checkCode == 0

			// FinalizeBlock leg, fresh app.
			finApp, _ := testApp(t)
			finBefore := snapshot(t, finApp, 1)
			finResults := finalize(t, finApp, 1, [][]byte{tc.build(t, finApp)})
			finAccepted := finResults[0].Code == 0

			if checkAccepted != tc.wantCheckAccept {
				t.Fatalf("case %q: CheckTx accepted=%v, want %v (code %d: %s)",
					tc.name, checkAccepted, tc.wantCheckAccept, checkCode, checkLog)
			}
			if finAccepted != tc.wantFinAccept {
				t.Fatalf("case %q: FinalizeBlock accepted=%v, want %v (code %d: %s)",
					tc.name, finAccepted, tc.wantFinAccept, finResults[0].Code, finResults[0].Log)
			}
			if !tc.wantCheckAccept {
				assertUnchanged(t, tc.name+" (CheckTx)", checkBefore, snapshot(t, checkApp, 2))
			}
			if !tc.wantFinAccept {
				// A FinalizeBlock rejection after a successful
				// ante handler still charges the fee and consumes
				// the sequence (standard failed-execution
				// semantics); that only applies to the
				// insufficient-funds case, whose ante handler
				// succeeds. Every other case here fails in the
				// ante handler, so nothing may change at all.
				if tc.name == "insufficient funds" {
					after := snapshot(t, finApp, 2)
					if !after.aliceBal.Equal(finBefore.aliceBal.Sub(math.NewInt(100))) {
						t.Errorf("insufficient funds (FinalizeBlock): alice balance %s, want exactly the 100 uogo fee deducted", after.aliceBal)
					}
					if after.aliceSeq != finBefore.aliceSeq+1 {
						t.Errorf("insufficient funds (FinalizeBlock): sequence %d, want exactly one consumed", after.aliceSeq)
					}
					if !after.supply.Equal(finBefore.supply) || !after.issued.Equal(finBefore.issued) {
						t.Errorf("insufficient funds (FinalizeBlock): supply or issued changed: %+v -> %+v", finBefore, after)
					}
				} else {
					assertUnchanged(t, tc.name+" (FinalizeBlock)", finBefore, snapshot(t, finApp, 2))
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Rejected transactions cannot create or alter supply (wired keepers)
// ---------------------------------------------------------------------------

// The monetary bypass suite (x/monetary/bypass_test.go) proves the
// ledger model rejects mint paths. This test proves the same property
// against the REAL wired keepers: a block full of rejected
// transactions, and a mempool full of rejected CheckTx calls, leave
// total supply and total issued byte-identical.
func TestAuthRejectedTxCannotAlterSupply(t *testing.T) {
	// FinalizeBlock leg: one block carrying every ante-level
	// rejection class. The insufficient-funds case is NOT in this
	// batch: its ante handler succeeds (valid signature, affordable
	// fee), so it legitimately pays its fee and consumes its
	// sequence when message execution fails; it is asserted
	// separately below.
	app, _ := testApp(t)
	before := snapshot(t, app, 1)
	bad := [][]byte{
		buildTx(t, app, txOpts{noSig: true}),
		buildTx(t, app, txOpts{signer: testBobPriv, acctNum: 1, seq: 0}),
		buildTx(t, app, txOpts{acctNum: 1, seq: 0, tamperMsg: true}),
		buildTx(t, app, txOpts{acctNum: 1, seq: 0, chainID: "ogo-local-999"}),
		buildTx(t, app, txOpts{acctNum: 0, seq: 0}),
		buildTx(t, app, txOpts{acctNum: 1, seq: 0, feeDenom: "ustake", feeAmt: 100}),
		{0x01, 0x02, 0x03},
	}
	results := finalize(t, app, 1, bad)
	for i, r := range results {
		if r.Code == 0 {
			t.Errorf("rejected-tx block: tx %d was accepted, must be rejected", i)
		}
	}
	after := snapshot(t, app, 2)
	assertUnchanged(t, "rejected-tx block", before, after)
	if !after.supply.Equal(math.NewInt(1_000_000)) {
		t.Errorf("supply after rejected-tx block: got %s, want 1000000", after.supply)
	}
	if !after.issued.IsZero() {
		t.Errorf("total issued after rejected-tx block: got %s, want 0", after.issued)
	}

	// Separate leg: the insufficient-funds transaction fails at
	// message execution, so it pays its fee and consumes its
	// sequence, but it must not create or alter supply, issue
	// anything, or move the amount.
	app, _ = testApp(t)
	before = snapshot(t, app, 1)
	bz := buildTx(t, app, txOpts{acctNum: 1, seq: 0, amount: 2_000_000})
	results = finalize(t, app, 1, [][]byte{bz})
	if results[0].Code == 0 {
		t.Fatal("insufficient-funds tx must be rejected in FinalizeBlock")
	}
	after = snapshot(t, app, 2)
	if !after.aliceBal.Equal(before.aliceBal.Sub(math.NewInt(100))) {
		t.Errorf("alice balance after failed execution: got %s, want exactly the 100 uogo fee deducted", after.aliceBal)
	}
	if after.aliceSeq != before.aliceSeq+1 {
		t.Errorf("sequence after failed execution: got %d, want exactly one consumed", after.aliceSeq)
	}
	if !after.bobBal.Equal(before.bobBal) {
		t.Errorf("recipient balance changed by failed execution: %s -> %s", before.bobBal, after.bobBal)
	}
	if !after.supply.Equal(math.NewInt(1_000_000)) {
		t.Errorf("supply after failed execution: got %s, want 1000000", after.supply)
	}
	if !after.issued.IsZero() {
		t.Errorf("total issued after failed execution: got %s, want 0", after.issued)
	}

	// CheckTx leg: the same classes through the mempool path. The
	// block is advanced to height 1 first, so CheckTx verifies
	// against the stored account number (1); at height 0 the
	// account-number-0 transaction in this batch would verify
	// (BaseApp height-0 behaviour) and the batch would not be
	// uniformly rejected.
	app, _ = testApp(t)
	advanceEmptyBlock(t, app, 1)
	before = snapshot(t, app, 2)
	for i, bz := range bad {
		if code, logMsg := checkTx(t, app, bz); code == 0 {
			t.Errorf("rejected-tx CheckTx: tx %d was accepted, must be rejected", i)
		} else {
			t.Logf("rejected-tx CheckTx: tx %d rejected (code %d): %s", i, code, logMsg)
		}
	}
	assertUnchanged(t, "rejected-tx CheckTx batch", before, snapshot(t, app, 2))
}
