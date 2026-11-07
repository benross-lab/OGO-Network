# OGO Network Milestone C: Transaction Authentication Security Note

Date: 2026-11-06. Scope: the Milestone C Cosmos SDK chain in the local staging repository, `ogo-network-milestone-c/chain/`. Effective SDK version: v0.53.4 (Gaia v28.3.1 matrix, pinned by replace in `go.mod`). Method: code inspection plus implementation tests run in process against the wired application on a memory database. This note is not an independent security audit. The tests below were written by the implementer, not by an independent auditor, and they prove the stated behaviours in the test setup only.

## Implemented and tested

### Authentication architecture

* The application uses the standard SDK ante handler, and nothing else. `chain/app/app.go`, function `setAnteHandler`, calls `ante.NewAnteHandler` with `HandlerOptions` containing the account keeper, the bank keeper, the transaction config sign mode handler, the fee grant keeper, and `DefaultSigVerificationGasConsumer`. No decorator was added, removed, or modified for this milestone, because the tests below showed no gap in the standard chain.
* The standard chain, in order, is: `SetUpContextDecorator`, `ExtensionOptionsDecorator`, `ValidateBasicDecorator`, `TxTimeoutHeightDecorator`, `ValidateMemoDecorator`, `ConsumeGasForTxSizeDecorator`, `DeductFeeDecorator`, `SetPubKeyDecorator`, `ValidateSigCountDecorator`, `SigGasConsumeDecorator`, `SigVerificationDecorator`, `IncrementSequenceDecorator`. These are the SDK v0.53.4 defaults from `x/auth/ante/ante.go`.
* `chain/app/app_config.go` sets `SkipAnteHandler: true` on the transaction config, with the comment that the custom ante handler is set in `app.go`. The handler in `app.go` is the standard SDK handler, so the skip only avoids registering the config default twice. It does not disable authentication.
* `chain/app/app_config.go` sets `EnableUnorderedTransactions: true` on the auth module. All transactions in the tests are ordered, so sequence and replay semantics apply to them. Unordered transactions were not tested, see the next section.
* `x/mint` is deliberately excluded from the application, so there is no discretionary minting path in the wired app. `x/monetary` is registered through the existing hybrid hook. The `x/epochs` hook workaround described in `EPOCHS_GAP.md` (the depinject copy versus pointer defect) was not touched.
* No custom signature scheme was created, and no unsigned JSON transaction pathway exists in this build. Every transaction must arrive as a signed protobuf transaction through the SDK transaction decoder and the ante handler.

### Signature verification

* Signature verification is done by the SDK `SigVerificationDecorator` using the sign mode handler from the application transaction config, with gas charged by `DefaultSigVerificationGasConsumer`. Keys in the tests are secp256k1, and the sign mode is `SIGN_MODE_DIRECT`.
* Tested and passing: a valid signed transaction is accepted on both the CheckTx path and the FinalizeBlock path. A transaction signed by a different key than the account holds is rejected. A transaction whose message was changed after signing is rejected. A transaction whose signature bytes were corrupted after signing is rejected. Observed rejection code for signature failures is 4, `unauthorized`, with the log naming the expected account number and chain ID.
* An unsigned transaction, with no signatures at all, is rejected on both paths. Observed CheckTx code is 15, `no signatures supplied`.

### Signer and public key validation

* `SetPubKeyDecorator` stores the public key on the account at first use, before the signature verification decorators run. Later transactions are verified against the stored key.
* Tested and passing: a forged sender, where the message claims alice sends but the transaction is signed by a different key, is rejected on both paths. Observed CheckTx code is 8, `invalid pubkey`, reporting that the public key does not match the signer address.
* Tested and passing: a public key and address mismatch, where the signature is valid for alice's key but the signer info carries a different public key, is rejected on both paths.

### Chain ID and account number validation

* The signature commits to the chain ID and the account number, because both are part of the signed document. A mismatch therefore fails signature verification rather than being checked separately.
* Tested and passing: a transaction signed with chain ID `ogo-local-999` instead of `ogo-local-1` is rejected on both paths, code 4.
* Tested and passing: a transaction signed with account number 0 while the stored account number is 1 is rejected on the FinalizeBlock path, code 4, and on the CheckTx path once a block has been committed (height 1 or later), code 4. The rejection log names the stored number: `please verify account number (1)`.
* Height 0 CheckTx behaviour is a documented BaseApp property, not an OGO gap. At height 0, CheckTx verifies with account number 0 regardless of the stored number, because the check state header height is 0. The mirror case is tested: a transaction signed with account number 1 at height 0 CheckTx is rejected, code 4, naming account number (0). A live chain verifies CheckTx at the current height, so the stored number applies there. The FinalizeBlock path always verifies against the stored number.

### Sequence and replay protection

* The sequence is checked by `SigVerificationDecorator` and advanced by `IncrementSequenceDecorator`, one increment per accepted transaction.
* Tested and passing: replaying identical transaction bytes after delivery is rejected on both paths, code 32, `account sequence mismatch, expected 1, got 0`.
* Tested and passing: submitting the same transaction twice in one block executes it once. The first copy returns code 0, the second returns code 32, the sequence advances exactly once, the sender balance falls by exactly the amount plus the fee, and supply and total issued do not change.
* Tested and passing: submitting the same transaction twice through CheckTx admits the first and rejects the second, code 32, because the first admission consumes the sequence in mempool state. CheckTx does not commit state: after a successful CheckTx, the committed sequence and balances are unchanged.

### Valid signed transaction

* Tested and passing on both paths. On the FinalizeBlock path, a valid 1000 uogo send with a 100 uogo fee moves alice from 1,000,000 to 998,900, moves bob from 0 to 1000, advances alice's sequence from 0 to exactly 1, leaves total supply at 1,000,000, and leaves total issued at 0. On the CheckTx path, the same transaction returns code 0 and changes no committed state.

### Fees

* Tested and passing: a fee in a denomination the sender does not hold (`ustake`) is rejected on both paths, code 5, `insufficient funds`. A fee larger than the sender's whole balance (2,000,000 uogo against 1,000,000) is rejected on both paths, code 5.
* Zero fee behaviour is documented by a test, not assumed. No minimum gas price is set in the application code. The node takes it from its `--minimum-gas-prices` flag at runtime (`cmd/ogod/cmd/root.go`), and the test application sets none. A zero fee transaction with a valid signature is therefore accepted on both paths in this test setup, and the test pins that behaviour so a later minimum gas price change shows up as a failure. See Open risks for the launch implication.

### Malformed transactions

* Tested and passing on both paths: random bytes are rejected, code 2, `tx parse error`. A truncated valid transaction (first 10 bytes, cutting the embedded body mid message) is rejected, code 2. Empty bytes are rejected, code 18, `must contain at least one message`. No state changes in any malformed case.

### CheckTx and block execution behaviour

* The two paths were compared explicitly in `TestAuthPathsAgree`, 13 cases, each run through CheckTx (after an empty block, so the stored account number applies) and through FinalizeBlock on separate fresh applications. Every authentication failure is rejected on both paths. No transaction passes one path and bypasses the other.
* One documented split exists, and it is standard SDK design, not a bypass. CheckTx never executes messages: `runMsgs` in the SDK `baseapp` runs message handlers only in FinalizeBlock and Simulate modes, and returns immediately in CheckTx mode. A transaction with a valid signature but an unaffordable message is therefore admitted by CheckTx and rejected at block execution, code 5, `failed to execute message`. Admission is not execution, and the transaction can never change state through CheckTx alone. The test asserts the split explicitly, with different expected verdicts per path.
* Failed execution still charges the fee and consumes the sequence. This is also standard SDK design: `runTx` writes the ante handler branch (fee deduction, sequence increment) before running messages, and message execution runs on a further branch that is discarded on failure. The tests assert the exact effect for an insufficient funds transaction delivered in a block: alice falls by exactly the 100 uogo fee, the fee collector module account gains exactly 100 uogo, the sequence advances by exactly one, the recipient gains nothing, total supply stays 1,000,000, and total issued stays 0. A failed transaction is therefore not free, and tooling must account for that.

### Monetary bypass protection

* The existing monetary bypass suite was re-run: 17 tests matching `TestBypass`, `TestGov`, and `TestWasm` in `x/monetary/bypass_test.go`, all passing, 0 failures. Those tests prove the ledger model rejects module account minting, governance changes to max supply, minting flags, issuance schedule, and mint paths, and CosmWasm minting and forbidden methods, with state unchanged after each attempt.
* The bypass suite tests the standalone ledger model, not the wired keepers. The wired keeper proof is the new test `TestAuthRejectedTxCannotAlterSupply`: a block carrying seven rejected transactions (unsigned, forged sender, tampered, wrong chain ID, wrong account number, unknown fee denomination, malformed bytes) leaves alice balance, bob balance, total supply, total issued, and alice sequence unchanged. A separate leg proves the same for an insufficient funds transaction that fails at message execution, with only the fee and sequence effects described above. Supply stays 1,000,000 and total issued stays 0 in every case.
* The epoch hook regression tests pass unchanged: `TestHookFiresInWiredApp` (issuance fires exactly once per epoch transition, 1000 then 900 uogo to the rewards pool) and `TestDuplicateHookExecutionCannotIssueTwice` (a second execution for the same epoch fails and issues nothing). The hook workaround was not modified.

### Tests added and actual results

* New file: `chain/app/auth_test.go`. No application code was changed, because the tests found no gap in the ante handler or the app wiring. Existing file `chain/app/app_test.go` was not rewritten; the new file reuses its `testApp`, `ctxAt`, `testAlicePriv`, and `testAliceAddr` helpers.
* New tests, all in package `app`, all passing:
  * `TestAuthForgedSenderRejected`
  * `TestAuthUnsignedRejected`
  * `TestAuthInvalidSignatureRejected` (tampered message and corrupted signature, both paths)
  * `TestAuthPubkeyMismatchRejected`
  * `TestAuthWrongChainIDRejected`
  * `TestAuthWrongAccountNumberRejected` (FinalizeBlock, CheckTx at height 1, and the height 0 mirror case)
  * `TestAuthReplayedSequenceRejected`
  * `TestAuthDuplicateTransactionRejected` (duplicate CheckTx and duplicate in one block)
  * `TestAuthValidTransactionAccepted`
  * `TestAuthInsufficientFundsRejected` (CheckTx admission plus FinalizeBlock rejection with exact fee and sequence effects)
  * `TestAuthInvalidFeeRejected` (unknown denomination and excessive fee, both paths)
  * `TestAuthMalformedTransactionRejected` (random bytes, truncated transaction, empty bytes, both paths)
  * `TestAuthZeroFeeBehaviorDocumented` (zero fee accepted on both paths with no minimum gas price set)
  * `TestAuthPathsAgree` (13 subtests: valid, forged sender, unsigned, tampered message, corrupted signature, pubkey mismatch, wrong chain ID, wrong account number, insufficient funds with the documented split, unknown fee denomination, excessive fee, malformed random bytes, malformed empty bytes)
  * `TestAuthRejectedTxCannotAlterSupply` (rejected transaction block and CheckTx batch against the wired keepers)
* Final verification run, 2026-11-06, Go toolchain `~/workspace/tools/go-1.25.7/bin/go`, `TMPDIR` set to a directory under the chain tree because `/tmp` is a 512M tmpfs. Exact commands and exit codes, all run from `chain/`:
  * `go build ./...`, exit 0.
  * `go vet ./...`, exit 0.
  * `go test -count=1 ./...`, exit 0. Package results: `app` ok, `wasmcheck` ok, `x/monetary` ok. Packages with no test files: `batchdemo`, `cmd/ogod`, `cmd/ogod/cmd`, `loadgen`.
  * Verbose counts from the same tree state: `app` 21 passed, 0 failed (6 pre-existing tests plus the 15 new tests above). `x/monetary` 78 passed, 0 failed. No failures anywhere in the run.
* Test setup facts, for reproduction: each test builds a fresh wired application on a memory database. Alice is funded 1,000,000 uogo, account number 1, sequence 0, chain ID `ogo-local-1`. Monetary genesis is strict, M = G + S_actual = 1,003,439. All FinalizeBlock calls use the fixed genesis time, so no monetary epoch ends during the authentication tests and no issuance changes the supply. The FinalizeBlock drive was verified empirically before the suite was written: FinalizeBlock at height 1 followed by Commit, then reads through uncached contexts, with height validation requiring each next block to be exactly one higher than the last committed height.
* No integration or localnet test was run against the existing localnet data. There is no localnet test suite in this repository, and `chain/localnet-data/` was preserved exactly: its pre-existing runtime modifications were not committed and not altered by this work. All tests above used memory databases and temporary directories.

## Implemented but not adequately tested

* Only one message type was exercised: `MsgSend` from the bank module. Staking, governance, distribution, slashing, authz, feegrant, upgrade, vesting, circuit, evidence, and consensus messages all pass through the same ante handler, but none was executed through it in these tests. Their message level authorization (for example, who may submit a given governance or staking message) is not proven here.
* Only secp256k1 keys and `SIGN_MODE_DIRECT` were tested. Other sign modes supported by the transaction config sign mode handler were not tested, and no multisig transaction was built or tested.
* Only single signer transactions were tested. Multi signature accounts, nested multisig, and threshold signing are unproven.
* The fee grant keeper is wired into the ante handler options, but no test exercises a fee grant paying a fee, and no authz grant execution was tested. A misconfigured grant could pay fees or authorize messages in ways these tests would not catch.
* Unordered transactions are enabled in the application config but were never built or tested. Their timeout timestamps and replay window are unproven.
* Mempool behaviour beyond single CheckTx calls was not tested. There is no test for recheck after commit, mempool eviction under load, transaction priority ordering, or duplicate handling across a recheck.
* Gas metering was only observed (for example, a failed signature verification consumed 30,808 gas in FinalizeBlock, and a valid transaction consumed 79,272). No test asserts gas limits, out of gas behaviour, or block gas limit enforcement.
* No live network test was performed. All results are in process against a memory database with one funded account and no validators, no real CometBFT consensus, no network partitions, and no byzantine proposer.
* The public key setting path is only exercised implicitly. `SetPubKeyDecorator` stores the key on first use, and the tests prove a wrong key is rejected, but no test asserts the stored key value after a first transaction, or the behaviour of an account whose key was set by an earlier transaction on a live chain.
* Fee deduction for a valid transaction was verified only for the fee collector path in the insufficient funds case. The valid transaction test asserts the sender balance and the recipient balance, but does not assert the fee collector balance after a successful transaction.

## Designed but not implemented

* IBC and CosmWasm are not wired in this build. Both remain deferred, so there is no cross chain or smart contract authentication surface in this application, and no test covers one.
* No custom signature scheme was designed or implemented. This is deliberate: authentication rests entirely on the standard SDK ante handler and the secp256k1 signatures it supports.
* No unsigned JSON transaction pathway was designed or implemented. This is deliberate: there is no way to submit a transaction without going through the SDK transaction decoder and the ante handler.
* No minimum gas price is enforced in the application code. The only control is the node `--minimum-gas-prices` flag, read at runtime. No default minimum is set by the application itself.
* No mempool fee decorator is configured in the ante handler chain. The SDK default `NewAnteHandler` used here does not include one, so mempool fee filtering depends on the base application mempool configuration and the node minimum gas price flag.
* No per account transaction rate limiting, no transaction allow list, and no circuit breaker on transaction submission are implemented in this milestone.
* No independent security audit has been performed on this chain or on these tests. Implementation tests are not an independent audit, and this note does not claim otherwise.
* Key management, validator key handling, and node operator procedures are out of scope for this note and were not reviewed.

## Open risks

* Zero fee acceptance. With no minimum gas price set in the application code, a zero fee transaction with a valid signature is accepted, as pinned by `TestAuthZeroFeeBehaviorDocumented`. On a public network, spam resistance depends on every validator operator setting `--minimum-gas-prices` in node configuration. If an operator runs with no minimum, their node accepts zero fee transactions. A launch checklist should set and verify a minimum gas price, and the decision should be recorded rather than assumed.
* Height 0 CheckTx account number quirk. At height 0, CheckTx verifies with account number 0 regardless of the stored number. The exposure window is the period before the first block is committed, when no live chain serves user traffic, but any tooling that runs CheckTx against a height 0 node would see different verification than a live chain. Operators and client developers should not treat height 0 CheckTx results as representative.
* CheckTx admission is not execution. By SDK design, CheckTx never executes messages, so it admits transactions that will fail at block execution, including transactions with unaffordable amounts. Any service that treats CheckTx acceptance as proof that a transaction will execute is mistaken. The failure is safe (the transaction cannot change state), but the signal is weak.
* Failed execution still costs a fee and a sequence. A transaction that passes the ante handler and then fails at message execution pays its fee and consumes its sequence. Users who retry a failed transaction with the same sequence will be rejected, and users who do not expect the fee will be surprised. Client software must surface this.
* Unordered transactions are enabled but untested. If unordered transactions are used before dedicated tests exist for their timeout and replay behaviour, a replay or expiry defect in that path would not have been caught by this milestone.
* Single message, single signer, single key type coverage. The authentication proof covers bank sends by one secp256k1 signer in direct mode. Other messages, multisig, and other sign modes are unproven, so a defect in one of those paths would not have been caught here.
* In process testing only. No live testnet evidence exists for this milestone. CometBFT level behaviour (mempool propagation, proposer selection, recheck after commit, vote extensions) is outside the tested surface.
* Dependency risk. Transaction authentication rests on the pinned SDK v0.53.4, including its ante handler, signature verification decorator, and mempool code. Any advisory affecting those components applies directly to this chain, and no separate dependency audit was performed for this milestone.
* Monetary parameters and supply rules were not changed by this work, and no monetary parameter was selected here. That boundary was respected: the only monetary assertions are that rejected and valid transactions leave supply and total issued as the genesis and epoch schedule define them.
* The `x/epochs` hook workaround was not touched. Its correctness rests on the existing hook tests, which pass, but the workaround itself (rebuilding the epochs module from the hooked keeper) is a manual fix for an upstream depinject defect, and an SDK upgrade could change the behaviour it compensates for.
