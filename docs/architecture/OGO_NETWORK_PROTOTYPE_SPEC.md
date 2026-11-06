# OGO Network Local Prototype: Technical Specification

Status: Specification of the prototype as built, 2026-11-06. Local staging only. No repository created on GitHub, no blockchain deployed, no token issued, no testnet launched, no spending authorized, no production change authorized.
Date: 2026-11-06
Prototype directory: `ogo-network-prototype/`
Companion documents: `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (stages M1 to M5), `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, `OGO_NETWORK_DECISION_REGISTER.md`, and the prototype build report `ogo-network-prototype/ogo-network-prototype-build-report.md`.
All values in this specification are TEST-ONLY and labelled so.

## 1. Purpose

This specification describes the local OGO Network prototype that was built in Phase 2, Milestones 1 to 5, as it actually exists, not as the plan hoped it would exist. It states what the prototype is, what it proves, and what it does not prove. Where the prototype differs from the plan or the specifications, the difference is stated here, with the reason and the Milestone C item that closes it.

The prototype's job is narrow: prove that three local validators can produce blocks and agree on the same ledger, that the monetary epoch hook demonstrably fires, that native OGO Coin transfers reconcile across validators, that geometric and hybrid issuance are both selectable from genesis configuration, that agent batching measurably reduces fee burden, and that missed epochs recover sequentially without duplicate issuance. It is not a testnet, not a wallet, not a production chain, and not evidence about any monetary parameter.

## 2. What the prototype is

- A Cosmos SDK application built from the verified Milestone B scaffold. The dependency matrix is LOCKED as one tested unit: Go 1.25.7, Cosmos SDK v0.53.6 required with Gaia's replace to v0.53.4 preserved, CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0. `go.mod` and `go.sum` are byte-identical to the Milestone B scaffold. No dependency was independently upgraded.
- A local repository initialized in staging only (`git init`, no remote, no push). Remote creation awaits Benross's explicit approval.
- A single binary, `ogod`, built from `chain/cmd/ogod/main.go`, with subcommands for genesis setup, node start, account and validator queries, transfers, burns, and balance and supply queries.
- Three local validators (val1, val2, val3), each on a separate home directory and a separate key, on distinct RPC and P2P ports, with persistent peers, started by `chain/scripts/localnet.sh` from one shared genesis.
- One custom module, `x/ogocoin`, the scaffold module that Milestone C renames to `x/monetary`. Standard SDK keepers (x/bank, x/auth, x/staking, x/slashing, x/distribution, x/gov, x/upgrade) are pinned and compiled only; they are not wired. The module boundaries are documented in `chain/MODULE_BOUNDARIES.md`.

## 3. Validator count

Three validators, decided by Benross, 2026-11-06 (Phase 2 Execution Directive). The distinction, stated in the prototype README, the demo script, the localnet script, and the module boundaries document:

- 3 is the decided local-prototype execution requirement.
- 4 is the minimum operational set (Definition A), the smallest set tolerating one faulty validator. It is the target for the slashing and liveness test (S5) on a later network, and for the M3 validator prototype.
- 50 or more is the decentralization target (Definition B) for a public network.

Consequence of 3: taking one validator offline halts the chain, because 2 of 3 is exactly two thirds and does not reach the more-than-two-thirds quorum. The demo proves this deliberately (halt-and-resume step): the chain halts with val3 down and resumes when val3 restarts, with no replay and no double mint. Continued block production with one validator down is NOT demonstrated by this prototype; it needs the 4-validator minimum set and is a later-network target.

## 4. Consensus and epoch execution

### 4.1 Consensus

- The three validators produce blocks on CometBFT consensus and agree on the same ledger state. Agreement is verified by comparing, at the same height, the app hash, the alice and bob balances, the total supply, and the total issued, on all three nodes. The demo verifies at heights 5, 15, 16, 17, and 18 in the verified runs.
- Block time is the CometBFT default in this scaffold. No block-time tuning was done; none is claimed.
- Restart persistence is implemented: height and epoch state survive a full restart of all three nodes, and the chain continues from the persisted height rather than replaying.

### 4.2 Epoch execution: the mirror, not the module

The prototype does NOT wire the real SDK x/epochs module. It implements an epoch timer mirror in `chain/x/ogocoin/epochs.go`, and `chain/EPOCHS_GAP.md` records exactly what is mirrored and what is not. This is the most important limitation in this specification, and it must not be misread.

What the mirror does:

- A fixed epoch grid anchored at genesis time, 1-indexed epochs, one tick per block, driven by block time. Determinism comes from epoch numbers, never from wall-clock time.
- `AfterEpochEnd` and `BeforeEpochStart` hooks, `epoch_end`, `epoch_start`, and `epoch_issuance` events, and `MintScheduledForEpoch`.
- The app calls the tick at the top of FinalizeBlock. Epoch state is persisted in `data/ogocoin/app.json` alongside height.
- The epoch duration in the demo is 20 seconds (TEST-ONLY).

What the mirror does NOT do (Milestone C work):

- No x/epochs keeper, no x/epochs store, no depinject wiring, no gRPC queries for epochs, no app-hash coverage of epoch state.
- The depinject hook-registration bug (issues #24770, #25425) cannot be exercised by the mirror, because the mirror is not registered through depinject. The hook-fires test in the prototype proves the mirror hook fires in the wired app; it does not prove the real x/epochs module registers its hook through depinject. That proof is the Milestone C hook-fires test, specified in `OGO_CONSENSUS_AND_EPOCH_SPEC.md` section 13 and `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` M2-5, and it is release-blocking for the monetary prototype.
- The StoreUpgrade path for adding x/epochs is documented in `EPOCHS_GAP.md` as six steps. It is not executed in this prototype.

### 4.3 Hook-fires test

The prototype's hook-fires test is `TestEpochHookFiresInWiredApp` in `chain/app/app_epochs_test.go`. It advances the wired app past epoch boundaries and asserts that issuance was minted and events were emitted. It covers the five conditions in the form the mirror allows: the epoch module (mirror) initializes, the monetary hook is registered in the app's hook set, the hook actually fires at an epoch transition, exactly one scheduled issuance occurs, and a duplicate hook execution cannot issue twice. The duplicate-claim failure is also covered by `TestEpochHaltCatchUpInWiredApp` (missed epochs tick one per block in order, no lump, no skip, no double mint, total 3,439 exact for the demo schedule) and by the ledger-level duplicate tests.

The test does not cover depinject registration of the real x/epochs module. That gap is stated in the test file, in `EPOCHS_GAP.md`, and in the build report. No claim in this specification presents the mirror test as the Milestone C hook-fires test.

## 5. Native OGO Coin accounting

- Denomination: `uogo`, 6 decimals (proposed, not decided). Chain id in the demo: `ogo-local-1`. All values TEST-ONLY.
- Genesis accounts: alice 600,000,000 uogo, bob 400,000,000 uogo, total genesis supply 1,000,000,000, max supply 1,000,010,000 (TEST-ONLY). The max supply in the demo is a test parameter, not a chosen supply.
- Transfers: `tx send` moves uogo between accounts with a flat 100 uogo fee (TEST-ONLY, no gas metering). Broadcast time measured in the demo: 102 to 114 ms. A transfer of 123,456 uogo from alice to bob commits with code 0 and reconciles on all three nodes: alice 599,876,444, bob 400,123,456.
- Burning: `tx burn` reduces supply by exactly the burned amount, one way. Milestone B proved a burn of 1,000 uogo; the prototype retains that behaviour.
- Supply invariants: total supply equals genesis plus issued minus burned, exactly, at every block. The every-block invariant is checked in tests and in the demo verify step.
- Signature verification: NOT implemented. Transactions are JSON, not signature-checked. Real x/auth arrives with Milestone C depinject wiring. This is a known scaffold limitation, carried forward honestly.
- Gas metering: NOT implemented. The fee is flat per transaction. CheckTx reports a hardcoded GasWanted of 1,000. No congestion adjustment exists.

### 5.1 Issuance models

The monetary interface is `IssuanceModel` in `chain/x/ogocoin/models.go` (named IssuanceModel, not IssuanceSchedule, because that name is already the Milestone B genesis configuration struct). It supports three models, all selected through genesis configuration:

- Geometric decay: exact rational arithmetic, `floor(R0 * r^n)` with exact powers, no floating point.
- Hybrid: base floor plus decaying bonus, `base + floor(B0 * r^n)`.
- Hybrid with gentle geometric terminal taper: factor about 0.95 over the final half of the schedule, renormalized to the same total.

Exact rational arithmetic only in consensus-critical monetary code; no floating point. The implementation uses `math/big.Rat` for exact powers; an earlier `math.Int` 256-bit overflow panic was found and fixed during the build.

Spec research parameters reproduce exactly in tests: geometric R0 = 100,000, r = 9/10, N = 110 gives total 999,935, dust 65, R_109 = 1. Hybrid base 10,000, B0 = 60,900, N = 40 gives total 999,978, dust 22, first payment 70,900, last payment 11,000.

Two known divergences, recorded honestly:

1. **Exact versus stepwise floor.** The spec formula is floor of the exact power. The legacy `AmountForPeriod` floors at every step. They diverge from period 7 onward (period 7: 478 exact versus 477 stepwise; period 9: 387 versus 386). For the demo's 4-period schedule they are identical. The spec parameters give 999,935 exact versus 999,572 stepwise, with stepwise hitting 0 at period 93. Milestone C must pick one rule and remove the other.
2. **Taper numbers.** The measured renormalized final payment is 4,360 uogo with a maximum single-period drop of 9.28 percent, not the spec's rounded claim of about 5,000 with a 6.5 percent maximum drop. The spec's rounded claim is superseded by the measured numbers.
3. **Hybrid through the epoch path.** The hybrid and tapered-hybrid models exist as code and are tested at the ledger level, but the app's epoch hook still mints the legacy geometric schedule only, and `ogod init` has no flags to express a hybrid genesis. Geometric is the only model proven live through the epoch path. Wiring the model-aware entrypoint (`MintScheduledModel`) into the hook, plus CLI flags for a hybrid genesis, is a Milestone C item. The seam is documented in `EPOCHS_GAP.md`.
4. **Strict M = G + S_actual is not the default genesis path.** `ValidateGenesis` still allows headroom above G + S_actual. The strict rule exists as `ValidateGenesisExact` and `ComputeMaxSupply`. Flipping the default is a Milestone C item.

### 5.2 Treasury

The R-F rule is implemented in exact integers in `chain/x/ogocoin/treasury.go`: spend_t = min(D, max(Trail_4(t), floor(0.05 * B_(t-1))), max(0, B_(t-1) - F)). The low-inflow binding is proved in tests: at inflow 100 per period, D = 1,500 and D = 2,500 both spend exactly 1,900 over 20 periods, balance ends 10,100, never below the 10,000 floor. The treasury is a ledger-level module in this prototype; there is no on-chain treasury governance, no multisig, and no treasury spending proposal flow. Those are later milestones.

## 6. Agent transaction economics

The batching demonstration is `chain/batchdemo/main.go`, built as `chain/batchdemo-bin` and called by `demo.sh` step 7. It measures from the prototype ledger, not from the spec table.

Measured results (verified by the coordinator, 2026-11-06):

- Unbatched: 100 payments of 1,000 uogo (0.001 OGO each), 100 transactions, one 100 uogo fee each. Total fee 10,000 uogo. Fee per payment 100 uogo. Fee as share of payment 10.00 percent (1,000 basis points).
- Batched: 100 payments in 1 transaction, one 100 uogo fee. Total fee 100 uogo. Fee per payment 1 uogo. Fee as share of payment 0.10 percent (10 basis points).
- Fee saving from batching: 9,900 uogo, 99 percent of the unbatched fee.

This reproduces the spec's finding exactly, but the numbers are measured from the prototype ledger, not copied from the spec. The fee model is flat per transaction with no gas metering, so the measurement proves the batching arithmetic under a flat fee, not under a gas-priced fee market. A gas-priced fee market is not implemented and not claimed. The throughput and latency measurements in the prototype are limited to single-transfer broadcast time (102 to 114 ms) and the batching demo; the proof-of-concept throughput criteria (P2, P3, P4) are not run by this prototype.

## 7. Network independence from the OGO application

The localnet runs with zero OGO-application involvement: no Supabase, no Worker, no benross-lab/OGO code path exists in the chain. The demo runs end to end with the application untouched. The chain is the authority for balances and supply; the application is a client of the network, not the authority controlling it. This is stated in the prototype README, `MODULE_BOUNDARIES.md`, the localnet script, and the demo script, and it is demonstrated by the demo running with no application process present.

## 8. Demonstration script

`demo.sh` in the prototype root runs the whole proof end to end:

1. Build `ogod` and start 3 validators.
2. Wait for the first epoch boundary and show the epoch state, including `hook_fire_count` and `total_issued`.
3. Transfer 123,456 uogo from alice to bob (fee 100 uogo), measuring broadcast time.
4. Show the same balances from all three nodes.
5. Verify the same ledger state on all three nodes (app hash, balances, supply, issuance) at the same height.
6. Halt-and-resume: stop val3, confirm the chain halts (no quorum with 2 of 3), restart val3, confirm blocks resume and nodes re-agree.
7. Run the agent batching demo (`chain/batchdemo-bin`).
8. Stop the localnet and leave no processes running.

Exit code 0 means every step passed. The demo is repeatable: it was run twice by the coordinator with identical results (balances, supply, issuance, and app-hash agreement at every verify step). The demo log is saved in `logs/demo-run-20261106.log`. Node logs are saved in `logs/node-val1-20261106.log`, `logs/node-val2-20261106.log`, and `logs/node-val3-20261106.log`. Test output is saved in `logs/test-output-20261106.log`.

## 9. Automated tests

90 top-level tests pass, 0 fail (`go test -count=1 ./...`, verified by the coordinator). Breakdown: 54 Milestone B tests (16 genesis, 7 schedule-math, 8 ledger, 6 IBC, 17 bypass), 13 new Milestone 2 epoch tests (including the hook-fires test and the halt catch-up test), and 23 new Milestone 3, 4, and 5 tests (issuance models, treasury, batching, monetary invariants). `go vet ./...` exits 0. The full test requirements are in `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md`.

## 10. What is NOT built

Stated plainly, so no reader mistakes this prototype for a chain:

- Real SDK x/epochs wiring (keeper, store, depinject, gRPC, app-hash coverage of epoch state). The mirror is a stand-in.
- Signature verification (JSON transactions, not signature-checked).
- Wasm VM instantiation and live CosmWasm execution.
- Live IBC relaying (escrow accounting is ledger-level only).
- Standard SDK keepers wired (x/bank, x/auth, x/staking, x/slashing, x/distribution, x/gov, x/upgrade are pinned and compiled only).
- Gas metering and a gas-priced fee market (flat fee only).
- Hybrid issuance through the app's epoch path (ledger-level only).
- Strict M = G + S_actual as the default genesis path (exists as `ValidateGenesisExact`, not the default).
- Staking, delegation, unbonding, slashing, remote signing.
- Keplr or Leap wallet integration.
- A rename of x/ogocoin to x/monetary (planned, outstanding).

Each item above is a Milestone C item, listed in `OGO_NETWORK_PROTOTYPE_BLOCKERS.md` with its owner and its unblock condition.

## 11. Reproducibility

- Toolchain: Go 1.25.7 at `~/workspace/tools/go-1.25.7/bin/go`. Go commands need `TMPDIR=~/workspace/.go-tmp` and `GOCACHE=~/workspace/.go-build-cache` (`/tmp` is a 512 MB tmpfs that fills up).
- Build: `cd chain && go build ./...`, `go vet ./...`, `go test -count=1 ./...`. Makefile targets `make vet` and `make test` are proven.
- Demo: `cd ogo-network-prototype && bash demo.sh`. No other setup is needed; the script builds the binary, initializes the localnet, runs the proof, and stops the nodes.
- The reproducible build from a clean copy (Milestone B, byte-identical under the canonical GOMODCACHE condition) is documented in `ogo-network-milestone-b/REPRODUCIBLE_BUILD.md` and was not re-run for this prototype; re-running it for the prototype is a Milestone C item, recorded in the blockers list.
