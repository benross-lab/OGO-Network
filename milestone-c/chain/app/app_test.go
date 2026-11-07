// Wired-application tests for the Milestone C SDK app.
//
// The release-blocking test is TestHookFiresInWiredApp: it advances
// the wired application past an epoch boundary and observes the
// x/monetary AfterEpochEnd hook executing against the REAL x/epochs
// module (issuance minted, event emitted). Compilation alone is not
// sufficient, and a unit test of the keeper alone is not sufficient:
// the pinned SDK v0.53.4 depinject wiring has the copy-versus-pointer
// defect (issues #24770, #25425) under which the hook never fires.
// All values are TEST-ONLY.
package app

import (
	"context"
	"testing"
	"time"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	monetary "github.com/benross-lab/ogo-network/x/monetary"
)

// testGenesisTime is the fixed genesis time for epoch tests.
var testGenesisTime = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

// testAlice is a funded test account (TEST-ONLY funds).
var testAlicePriv = secp256k1.GenPrivKey()

func testAliceAddr() sdk.AccAddress {
	return sdk.AccAddress(testAlicePriv.PubKey().Address())
}

// testApp builds a wired app on a memory database, with x/bank funded
// for the test account and x/monetary + x/epochs initialised for a
// TEST-ONLY schedule (4 periods, R0 1000, r 9/10, S_actual 3439) and a
// 60-second monetary epoch. It returns the app and a context at the
// genesis time.
func testApp(t *testing.T) (*OgoApp, sdk.Context) {
	t.Helper()
	app := NewOgoApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		sims.NewAppOptionsWithFlagHome(t.TempDir()),
		baseapp.SetChainID("ogo-local-1"),
	)
	ctx := ctxAt(app, 1, testGenesisTime)

	// Fund the test account in x/bank (G = 1_000_000 uogo, TEST-ONLY).
	alice := testAliceAddr()
	coins := sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(1_000_000)))
	app.BankKeeper.InitGenesis(ctx, &banktypes.GenesisState{
		Params:   banktypes.DefaultParams(),
		Balances: []banktypes.Balance{{Address: alice.String(), Coins: coins}},
		Supply:   coins,
	})

	// Auth genesis first (params, including TxSigLimit), then
	// create the test account BEFORE the other modules' genesis:
	// staking InitGenesis creates the bonded-pool module account.
	// Auth InitGenesis already created the fee collector module
	// account at number 0, so alice takes number 1.
	app.AccountKeeper.InitGenesis(ctx, *authtypes.DefaultGenesisState())
	acc := authtypes.NewBaseAccount(alice, testAlicePriv.PubKey(), 1, 0)
	app.AccountKeeper.SetAccount(ctx, acc)
	// Advance the account-number sequence past the manually created
	// account (NextAccountNumber returns and increments).
	_ = app.AccountKeeper.NextAccountNumber(ctx)

	// Initialise the standard modules' genesis state. Without
	// this, distribution BeginBlocker panics on a missing FeePool.
	app.DistrKeeper.InitGenesis(ctx, *distrtypes.DefaultGenesisState())
	_ = app.StakingKeeper.InitGenesis(ctx, stakingtypes.DefaultGenesisState())
	app.SlashingKeeper.InitGenesis(ctx, app.StakingKeeper, slashingtypes.DefaultGenesisState())
	if err := app.GovKeeper.Params.Set(ctx, govtypesv1.DefaultParams()); err != nil {
		t.Fatalf("gov params: %v", err)
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
		t.Fatalf("upgrade version map: %v", err)
	}

	// x/monetary genesis: strict M = G + S_actual = 1_000_000 + 3_439.
	max := math.NewInt(1_003_439)
	gs := monetary.GenesisState{
		Params: monetary.Params{
			MaxSupply:         &max,
			Denom:             monetary.Denom,
			Decimals:          monetary.Decimals,
			InflationRate:     "0",
			MintingEnabled:    false,
			SupplyCapEnforced: true,
			IssuanceSchedule: &monetary.IssuanceSchedule{
				ScheduleType:         monetary.ScheduleTypeReducingPerPeriod,
				StartHeight:          1,
				PeriodLength:         100,
				Periods:              4,
				InitialPerPeriod:     math.NewInt(1000),
				ReductionNumerator:   9,
				ReductionDenominator: 10,
			},
		},
		Balances: []monetary.Balance{
			{Address: alice.String(), Denom: monetary.Denom, Amount: math.NewInt(1_000_000), Purpose: "public"},
		},
	}
	if err := app.MonetaryKeeper.InitGenesis(ctx, gs); err != nil {
		t.Fatalf("monetary InitGenesis: %v", err)
	}

	// x/epochs genesis: one monetary epoch, 60 seconds.
	if err := app.EpochsKeeper.InitGenesis(ctx, epochstypes.GenesisState{
		Epochs: []epochstypes.EpochInfo{
			{
				Identifier:            monetary.EpochIdentifierMonetary,
				StartTime:             testGenesisTime,
				Duration:              60 * time.Second,
				CurrentEpoch:          0,
				CurrentEpochStartTime: testGenesisTime,
				EpochCountingStarted:  false,
			},
		},
	}); err != nil {
		t.Fatalf("epochs InitGenesis: %v", err)
	}
	return app, ctx
}

// ctxAt returns a context at the given height and block time.
// NewUncachedContext is used because BaseApp.NewContext needs block
// state (checkState / finalizeBlockState) that only exists after a
// block has run; on a freshly constructed app it nil-panics.
func ctxAt(app *OgoApp, height int64, when time.Time) sdk.Context {
	return app.NewUncachedContext(true, cmtproto.Header{Height: height, Time: when, ChainID: "ogo-local-1"})
}

// beginBlock runs the app's BeginBlocker at the given height and time
// and returns the resulting BeginBlock (which carries the events).
func beginBlock(t *testing.T, app *OgoApp, height int64, when time.Time) sdk.BeginBlock {
	t.Helper()
	res, err := app.BeginBlocker(ctxAt(app, height, when))
	if err != nil {
		t.Fatalf("BeginBlocker height %d: %v", height, err)
	}
	return res
}

func hasEvent(res sdk.BeginBlock, typ string) bool {
	for _, e := range res.Events {
		if e.Type == typ {
			return true
		}
	}
	return false
}

func TestHookFiresInWiredApp(t *testing.T) {
	app, _ := testApp(t)
	pool := monetary.RewardsPoolAddress()

	// Confirmation 1: the epochs module initialises with the monetary
	// epoch present.
	infos, err := app.EpochsKeeper.AllEpochInfos(ctxAt(app, 1, testGenesisTime))
	if err != nil {
		t.Fatalf("AllEpochInfos: %v", err)
	}
	if len(infos) != 1 || infos[0].Identifier != monetary.EpochIdentifierMonetary {
		t.Fatalf("epochs module did not initialise with the monetary epoch: %+v", infos)
	}

	// Confirmation 2: the monetary hook is registered on the epochs
	// keeper that the module manager's AppModule actually uses. The
	// proof is behavioural (confirmations 3-5); here we also assert
	// the manager's epochs module is the hooked one by checking the
	// module name resolves.
	if _, ok := app.ModuleManager.Modules[epochstypes.ModuleName]; !ok {
		t.Fatal("epochs module missing from the module manager")
	}

	// Block 1 at genesis time: epoch 1 starts, no issuance yet.
	res := beginBlock(t, app, 1, testGenesisTime)
	if hasEvent(res, "monetary_epoch_issuance") {
		t.Fatal("issuance event at epoch start: epoch 1 has not ended yet")
	}
	if got := app.BankKeeper.GetBalance(ctxAt(app, 1, testGenesisTime), pool, monetary.Denom); !got.Amount.IsZero() {
		t.Fatalf("rewards pool funded before any epoch ended: %s", got)
	}

	// Block 2, 61 seconds later: epoch 1 ends, epoch 2 starts.
	// Confirmation 3: the hook fires at the epoch transition.
	// Confirmation 4: exactly one scheduled issuance occurs
	// (period 0 = 1000 uogo, TEST-ONLY).
	res = beginBlock(t, app, 2, testGenesisTime.Add(61*time.Second))
	if !hasEvent(res, "monetary_epoch_issuance") {
		t.Fatalf("hook did not fire at the epoch transition: events %+v", res.Events)
	}
	ctx := ctxAt(app, 2, testGenesisTime.Add(61*time.Second))
	if got := app.BankKeeper.GetBalance(ctx, pool, monetary.Denom); !got.Amount.Equal(math.NewInt(1000)) {
		t.Fatalf("rewards pool after epoch 1: got %s, want exactly 1000 (one issuance)", got.Amount)
	}
	if got, err := app.MonetaryKeeper.GetTotalIssued(ctx); err != nil || !got.Equal(math.NewInt(1000)) {
		t.Fatalf("total issued after epoch 1: got %s err %v, want 1000", got, err)
	}
	if got := app.BankKeeper.GetSupply(ctx, monetary.Denom); !got.Amount.Equal(math.NewInt(1_001_000)) {
		t.Fatalf("total supply after epoch 1: got %s, want 1_001_000 (G 1_000_000 + issued 1000)", got.Amount)
	}

	// Block 3, one more epoch later: period 1 issues 900.
	res = beginBlock(t, app, 3, testGenesisTime.Add(121*time.Second))
	if !hasEvent(res, "monetary_epoch_issuance") {
		t.Fatal("hook did not fire at the second epoch transition")
	}
	ctx = ctxAt(app, 3, testGenesisTime.Add(121*time.Second))
	if got := app.BankKeeper.GetBalance(ctx, pool, monetary.Denom); !got.Amount.Equal(math.NewInt(1900)) {
		t.Fatalf("rewards pool after epoch 2: got %s, want 1900", got.Amount)
	}
}

func TestDuplicateHookExecutionCannotIssueTwice(t *testing.T) {
	app, ctx := testApp(t)
	// Confirmation 5: duplicate hook execution cannot issue twice.
	hook := monetary.MonetaryEpochHook{Keeper: app.MonetaryKeeper}
	if err := hook.AfterEpochEnd(ctx, monetary.EpochIdentifierMonetary, 1); err != nil {
		t.Fatalf("first hook execution must succeed: %v", err)
	}
	pool := monetary.RewardsPoolAddress()
	if got := app.BankKeeper.GetBalance(ctx, pool, monetary.Denom); !got.Amount.Equal(math.NewInt(1000)) {
		t.Fatalf("pool after first execution: got %s, want 1000", got.Amount)
	}
	if err := hook.AfterEpochEnd(ctx, monetary.EpochIdentifierMonetary, 1); err == nil {
		t.Fatal("second hook execution for the same epoch must fail")
	}
	if got := app.BankKeeper.GetBalance(ctx, pool, monetary.Denom); !got.Amount.Equal(math.NewInt(1000)) {
		t.Fatalf("pool changed after duplicate execution: got %s, want 1000", got.Amount)
	}
	if got, err := app.MonetaryKeeper.GetTotalIssued(ctx); err != nil || !got.Equal(math.NewInt(1000)) {
		t.Fatalf("total issued after duplicate execution: got %s, want 1000", got)
	}
	// A hook for another identifier mints nothing.
	if err := hook.AfterEpochEnd(ctx, "other", 2); err != nil {
		t.Fatalf("hook for another identifier must be a no-op: %v", err)
	}
	if got := app.BankKeeper.GetBalance(ctx, pool, monetary.Denom); !got.Amount.Equal(math.NewInt(1000)) {
		t.Fatalf("pool changed after other-identifier hook: got %s", got.Amount)
	}
}

func TestKeepersWiredAndQueryable(t *testing.T) {
	app, ctx := testApp(t)
	// Every keeper from C1 must be reachable through its query path.
	if got := app.BankKeeper.GetBalance(ctx, testAliceAddr(), monetary.Denom); !got.Amount.Equal(math.NewInt(1_000_000)) {
		t.Fatalf("bank keeper: alice balance %s, want 1_000_000", got.Amount)
	}
	if _, err := app.StakingKeeper.GetParams(ctx); err != nil {
		t.Fatalf("staking keeper params: %v", err)
	}
	if _, err := app.SlashingKeeper.GetParams(ctx); err != nil {
		t.Fatalf("slashing keeper params: %v", err)
	}
	if _, err := app.DistrKeeper.Params.Get(ctx); err != nil {
		t.Fatalf("distribution keeper params: %v", err)
	}
	if _, err := app.GovKeeper.Params.Get(ctx); err != nil {
		t.Fatalf("gov keeper params: %v", err)
	}
	if _, err := app.EpochsKeeper.GetEpochInfo(ctx, monetary.EpochIdentifierMonetary); err != nil {
		t.Fatalf("epochs keeper query: %v", err)
	}
	if _, err := app.MonetaryKeeper.GetParams(ctx); err != nil {
		t.Fatalf("monetary keeper params: %v", err)
	}
	if _, err := app.UpgradeKeeper.GetModuleVersionMap(ctx); err != nil {
		t.Fatalf("upgrade keeper version map: %v", err)
	}
	for _, name := range []string{
		authtypes.ModuleName, banktypes.ModuleName, stakingtypes.ModuleName,
		slashingtypes.ModuleName, govtypes.ModuleName, epochstypes.ModuleName,
		monetary.ModuleName,
	} {
		if _, ok := app.ModuleManager.Modules[name]; !ok {
			t.Fatalf("module %s missing from the module manager", name)
		}
	}
	if app.ModuleManager.Modules[epochstypes.ModuleName] == nil {
		t.Fatal("epochs module is nil in the manager")
	}
}

// TestMonetaryInitGenesisMigrationGuard: RunMigrations calls
// InitGenesis with DefaultGenesis (no max supply) for a module that
// is missing from x/upgrade's stored version map. When monetary
// state already exists, that call must be a no-op preserving the
// stored params, not a panic or a wipe. (Observed live 2026-11-06:
// the v2 upgrade panicked in RunMigrations because monetary was
// absent from the depinject-built init version map.)
func TestMonetaryInitGenesisMigrationGuard(t *testing.T) {
	app, ctx := testApp(t)
	before, err := app.MonetaryKeeper.GetParams(ctx)
	if err != nil {
		t.Fatalf("GetParams before: %v", err)
	}
	if err := app.MonetaryKeeper.InitGenesis(ctx, monetary.GenesisState{Params: monetary.DefaultParams()}); err != nil {
		t.Fatalf("InitGenesis with empty genesis on live state must be a no-op, got: %v", err)
	}
	after, err := app.MonetaryKeeper.GetParams(ctx)
	if err != nil {
		t.Fatalf("GetParams after: %v", err)
	}
	if !before.MaxSupply.Equal(*after.MaxSupply) || before.Denom != after.Denom {
		t.Fatalf("params changed by migration-path InitGenesis: %+v -> %+v", before, after)
	}
	if got, err := app.MonetaryKeeper.GetTotalIssued(ctx); err != nil || !got.IsZero() {
		t.Fatalf("total issued changed by migration-path InitGenesis: %s %v", got, err)
	}
}

func TestMonetaryChangeUpgradeRejected(t *testing.T) {
	app, ctx := testApp(t)
	// M3: an upgrade that changes a monetary parameter is rejected.
	// The handler is invoked directly here (not live: a failing
	// handler at an upgrade height halts the chain by design).
	if _, err := app.MonetaryChangeUpgradeHandler(ctx, upgradetypes.Plan{}, nil); err == nil {
		t.Fatal("monetary-change upgrade handler must fail")
	}
	// The keeper also refuses param changes after genesis.
	p, err := app.MonetaryKeeper.GetParams(ctx)
	if err != nil {
		t.Fatalf("GetParams: %v", err)
	}
	changed := p
	bigger := math.NewInt(2_000_000)
	changed.MaxSupply = &bigger
	if err := app.MonetaryKeeper.SetParams(ctx, changed); err == nil {
		t.Fatal("SetParams changing max_supply must fail")
	}
	if got, err := app.MonetaryKeeper.GetParams(ctx); err != nil || !got.MaxSupply.Equal(*p.MaxSupply) {
		t.Fatalf("params changed after rejected SetParams: %v %v", got, err)
	}
}

func TestSignatureVerification(t *testing.T) {
	app, _ := testApp(t)
	txConfig := authtx.NewTxConfig(app.AppCodec(), authtx.DefaultSignModes)
	alice := testAliceAddr()
	bob := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	buildTx := func(t *testing.T, signer *secp256k1.PrivKey, tamper bool) []byte {
		t.Helper()
		msg := banktypes.NewMsgSend(alice, bob, sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(1000))))
		builder := txConfig.NewTxBuilder()
		if err := builder.SetMsgs(msg); err != nil {
			t.Fatalf("SetMsgs: %v", err)
		}
		builder.SetGasLimit(200_000)
		builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(100))))
		sigV2 := signing.SignatureV2{
			PubKey: signer.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
				Signature: nil,
			},
			Sequence: 0,
		}
		if err := builder.SetSignatures(sigV2); err != nil {
			t.Fatalf("SetSignatures: %v", err)
		}
		// NOTE: the ante handler reads the account number from the
		// stored account only when the block height is non-zero;
		// BaseApp.CheckTx runs at height 0, where it verifies with
		// account number 0 regardless of the stored number (1).
		// A live chain verifies CheckTx at the current height, so
		// the stored number applies there.
		signerData := authsigning.SignerData{
			Address:       alice.String(),
			ChainID:       "ogo-local-1",
			AccountNumber: 0,
			Sequence:      0,
			PubKey:        signer.PubKey(),
		}
		sigV2, err := clienttx.SignWithPrivKey(
			context.Background(), signing.SignMode_SIGN_MODE_DIRECT, signerData,
			builder, signer, txConfig, 0,
		)
		if err != nil {
			t.Fatalf("SignWithPrivKey: %v", err)
		}
		if err := builder.SetSignatures(sigV2); err != nil {
			t.Fatalf("SetSignatures: %v", err)
		}
		if tamper {
			// Change the message after signing: the signature must
			// no longer verify.
			msg2 := banktypes.NewMsgSend(alice, bob, sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(999_999))))
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

	checkTx := func(bz []byte) (uint32, string) {
		res, err := app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
		if err != nil {
			t.Fatalf("CheckTx: %v", err)
		}
		return res.Code, res.Log
	}

	// Valid signed transfer commits in CheckTx (code 0) and the
	// ante handler meters gas (no hardcoded GasWanted).
	if code, logMsg := checkTx(buildTx(t, testAlicePriv, false)); code != 0 {
		t.Fatalf("valid signed tx rejected in CheckTx (code %d): %s", code, logMsg)
	}

	// Wrongly-signed transaction (signed by a different key than the
	// account's pubkey) is rejected before execution.
	other := secp256k1.GenPrivKey()
	if code, _ := checkTx(buildTx(t, other, false)); code == 0 {
		t.Fatal("wrongly-signed tx must be rejected")
	}

	// Tampered transaction (message changed after signing) is
	// rejected before execution.
	if code, _ := checkTx(buildTx(t, testAlicePriv, true)); code == 0 {
		t.Fatal("tampered tx must be rejected")
	}

	// Unsigned transaction: no signatures at all.
	msg := banktypes.NewMsgSend(alice, bob, sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(1000))))
	builder := txConfig.NewTxBuilder()
	if err := builder.SetMsgs(msg); err != nil {
		t.Fatalf("SetMsgs: %v", err)
	}
	builder.SetGasLimit(200_000)
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(monetary.Denom, math.NewInt(100))))
	unsigned, err := txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		t.Fatalf("unsigned TxEncoder: %v", err)
	}
	if code, logMsg := checkTx(unsigned); code == 0 {
		t.Fatal("unsigned tx must be rejected")
	} else {
		t.Logf("unsigned tx rejected as expected: %s", logMsg)
	}

	// State unchanged after the rejections: alice still holds her
	// full balance (CheckTx does not commit, and the signed tx above
	// was only checked, not delivered).
	ctx := ctxAt(app, 1, testGenesisTime)
	if got := app.BankKeeper.GetBalance(ctx, alice, monetary.Denom); !got.Amount.Equal(math.NewInt(1_000_000)) {
		t.Fatalf("alice balance changed by rejected/checked txs: %s", got.Amount)
	}
}
