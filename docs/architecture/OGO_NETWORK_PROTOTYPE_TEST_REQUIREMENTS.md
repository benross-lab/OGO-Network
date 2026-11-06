# OGO Network Local Prototype: Automated Test Requirements

Status: Test requirements for the prototype as built, 2026-11-06, plus the requirements Milestone C must add. Local staging only.
Date: 2026-11-06
Prototype directory: `ogo-network-prototype/`
Companion documents: `OGO_NETWORK_PROTOTYPE_SPEC.md`, `OGO_NETWORK_PROTOTYPE_TASKS.md`, `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`, and the prototype build report `ogo-network-prototype/ogo-network-prototype-build-report.md`.
All values are TEST-ONLY and labelled so.

## 1. How tests are run

- Command: `cd chain && go test -count=1 ./...`
- Toolchain: Go 1.25.7 at `~/workspace/tools/go-1.25.7/bin/go`, with `TMPDIR=~/workspace/.go-tmp` and `GOCACHE=~/workspace/.go-build-cache`.
- Vet: `go vet ./...` must exit 0. Makefile targets `make vet` and `make test` are proven.
- Current result, verified by the coordinator 2026-11-06: 90 top-level tests pass, 0 fail (85 in `x/ogocoin`, 5 in `app`). Full output saved in `ogo-network-prototype/logs/test-output-20261106.log`.
- Rule: a test that passes on the scaffold is not a pass on the wired chain. Tests must run against the prototype code in this directory. Compilation passing is not a working feature.

## 2. Test inventory (as built)

### 2.1 Milestone B tests (54, all passing)

| Group | Count | What they prove |
|---|---|---|
| Genesis validation | 16 | Valid genesis accepted; invalid genesis (bad schedule, founder or reserved purpose labels, missing mandatory labels) rejected |
| Schedule math | 7 | Per-period amounts, totals, and boundaries for the legacy schedule |
| Ledger | 8 | Balances, transfers, fees, burning, supply accounting |
| IBC (ledger level) | 6 | Escrow and voucher accounting, no double-counting of economic supply |
| Bypass | 17 | No discretionary mint through bank, staking, distribution, gov, or wasm; no governance change to max supply or schedule; no wasm mint; zero bypasses succeed, state unchanged after each attempt |

### 2.2 Milestone 2 epoch tests (13, all passing)

| Test | What it proves |
|---|---|
| `TestEpochHookFiresInWiredApp` | The hook-fires test in mirror form: advances the wired app past epoch boundaries and asserts issuance minted and events emitted. Covers the five conditions in the form the mirror allows: module initializes, hook registered, hook actually fires, exactly one issuance, duplicate cannot issue twice. Does NOT prove depinject registration of the real x/epochs module. |
| `TestEpochHaltCatchUpInWiredApp` | Missed epochs tick one per block in order, no lump, no skip, no double mint; total 3,439 exact for the demo schedule. |
| `TestAppRestartRestoresHeightAndEpochs` | Height and epoch state survive a restart; the chain continues from the persisted height, no replay. |
| Duplicate epoch claim | A duplicate claim fails with `ErrDoubleIssuance` and changes nothing. |
| Hook failure handling | A hook failure does not halt the chain; it is recorded in `LastHookError`. |
| Epoch-disabled fallback | With epochs disabled, the app falls back to the legacy block path. |
| Remaining epoch tests (8) | Epoch grid, 1-indexed epochs, one tick per block, epoch state persistence, event emission, `MintScheduledForEpoch` entrypoint. |

### 2.3 Milestone 3, 4, and 5 tests (23, all passing)

| Group | What they prove |
|---|---|
| Issuance models (`models_test.go`) | `IssuanceModel` interface; geometric, hybrid, and tapered-hybrid schedules; exact rational arithmetic; spec research parameters reproduce exactly (geometric R0 = 100,000, r = 9/10, N = 110, total 999,935, dust 65, R_109 = 1; hybrid base 10,000, B0 = 60,900, N = 40, total 999,978, dust 22, first 70,900, last 11,000). |
| Treasury (`treasury_test.go`) | R-F rule in exact integers; low-inflow binding (D = 1,500 and D = 2,500 both spend exactly 1,900 over 20 periods at inflow 100, balance ends 10,100, never below the 10,000 floor); parameter validation rejects bad parameters. |
| Batching (`batch_test.go`) | Batch fee arithmetic: 100 payments unbatched cost 10,000 uogo total fee; batched cost 100 uogo total fee; 99 percent saving. |
| Monetary invariants (`monetary_test.go`) | M = G + S_actual computed by summing the floored schedule; dust left permanently unissued; supply invariants hold. |
| Halt recovery (`monetary_test.go`) | After epochs 0 to 1, a simulated halt to epoch 5 catches up one per block (810, 729, 656, 590), no lump sum, no skip, total 4,685 equals the sum of R_0 through R_5. |
| Overflow | `math.Int` 256-bit overflow panic found and fixed with `math/big.Rat`. |

## 3. Requirements the prototype satisfies

Each requirement states the evidence, not just the test name.

1. **Duplicate issuance prevention.** The same epoch cannot mint twice. Evidence: ledger-level duplicate tests, epoch duplicate-claim test (`ErrDoubleIssuance`), and `TestEpochHookFiresInWiredApp` condition 5. A duplicate hook execution mints nothing and leaves total issued unchanged.
2. **Halt recovery.** Missed epochs are processed sequentially. Evidence: `TestEpochHaltCatchUpInWiredApp` (unit level) and the live demo halt-and-resume step (chain level). No lump-sum mint, no silent skipping, no double mint.
3. **Treasury accounting.** The R-F rule is implemented in exact integers and its low-inflow binding is proved. Evidence: `treasury_test.go`, including the D = 1,500 versus D = 2,500 comparison.
4. **Both issuance models end to end.** Geometric and hybrid both run from genesis configuration at the ledger level and reproduce the spec research parameters exactly. Evidence: `models_test.go`. Limitation: hybrid is not yet wired through the app's epoch path; only geometric is proven live. See section 5.
5. **Supply invariants.** M = G + S_actual, computed by summing the floored schedule, with dust left permanently unissued. Evidence: `monetary_test.go` and the demo verify step (supply matches exactly on all 3 nodes).
6. **Balance reconciliation.** The same balances are verified from all three nodes at the same height. Evidence: the demo verify step (heights 5, 15, 16, 17, 18) and the transfer test (alice 599,876,444, bob 400,123,456 on all 3 nodes).
7. **Epoch execution ordering.** Epochs tick one per block in epoch order after a halt. Evidence: `TestEpochHaltCatchUpInWiredApp` and the live demo epoch state (`hook_fire_count`, `total_issued`).
8. **Chain halt and recovery behaviour.** The chain halts with one validator down (no quorum with 2 of 3) and resumes when it restarts, with no replay and no double mint. Evidence: the demo halt-and-resume step and `TestAppRestartRestoresHeightAndEpochs`.

## 4. Requirements the prototype does NOT satisfy (Milestone C)

These are not gaps in the test files; they are gaps in the prototype the tests cannot close. Each is a Milestone C item.

1. **Real x/epochs hook-fires test.** The Milestone C test must prove, in the wired application with depinject, all five confirmations in `OGO_CONSENSUS_AND_EPOCH_SPEC.md` section 13: the epoch module initializes, the monetary hook is registered in the wired hook set, the hook actually fires at an epoch transition, exactly one scheduled issuance occurs, and a duplicate hook execution cannot issue twice. The mirror test does not prove depinject registration. This is release-blocking for the monetary prototype.
2. **Bypass suite re-run on the wired chain (S2).** The 17 Milestone B bypass tests pass on the scaffold-derived prototype, but the standard SDK keepers are not wired, so the bypass suite has not been re-run against the wired keepers. S2 requires the re-run on the wired chain with zero bypasses succeeding.
3. **Signature verification (S1).** A transaction naming a sender it is not signed by must be rejected before execution, with the reason logged. Not testable until x/auth is wired.
4. **Halt catch-up at production scale (S4).** Halt the chain for 30 minutes, restart, and observe missed epochs ticking one per block in epoch order with no lump, no skip, and no double mint. The prototype proves the mechanism at small scale; the 30-minute S4 test is not run.
5. **Hybrid through the epoch path.** The hybrid and tapered-hybrid models must be proven live through the app's epoch hook, not only at the ledger level. Requires wiring `MintScheduledModel` into the hook and CLI flags for a hybrid genesis.
6. **Strict M = G + S_actual as the default.** A test must prove the default genesis path rejects headroom above G + S_actual. Currently only `ValidateGenesisExact` enforces it; the default `ValidateGenesis` does not.
7. **Exact versus stepwise floor.** A test must prove the single chosen rounding rule. Currently the exact and stepwise rules diverge from period 7 onward and both exist in the code.
8. **Throughput and latency (P2, P3, P4).** Sustained 200 TPS simple transfers, 50 TPS mixed load, p95 finality under 6 seconds. Not run by this prototype.
9. **Slashing and liveness (S5).** Take 1 of 4 validators offline, observe downtime jailing and slashing, chain continues with the remaining 3. Cannot run on a 3-validator network. See the open question in the decision register.
10. **Wasm sandbox (S7).** A test CosmWasm contract attempting a mint, an overspend, and an unknown method call must fail all three with state unchanged. Requires the wasm VM to be instantiated.
11. **Reproducible build from a clean copy.** Re-run the Milestone B reproducibility proof (clean copy, fresh caches, public proxy only) for the prototype and record the binary hash comparison.

## 5. Known test-level divergences (recorded, not hidden)

1. **Exact versus stepwise floor.** Spec formula is floor of the exact power; legacy `AmountForPeriod` floors at every step. Period 7: 478 exact versus 477 stepwise. Period 9: 387 versus 386. Spec parameters: 999,935 exact versus 999,572 stepwise, with stepwise hitting 0 at period 93. Identical for the demo's 4-period schedule. Milestone C must pick one rule.
2. **Taper numbers.** Measured renormalized final payment 4,360 uogo with a maximum single-period drop of 9.28 percent, not the spec's rounded claim of about 5,000 with a 6.5 percent maximum drop. The measured numbers supersede the rounded claim.
3. **Treasury 4 versus 0 units. RESOLVED 2026-11-06: 0 units.** Recomputed from the stated model: both runs spend 114,511 over 120 periods. The 4-unit figure is withdrawn as not reproducible from the stated model. The substance is identical (the spending rule cannot fix a low-inflow level problem).
