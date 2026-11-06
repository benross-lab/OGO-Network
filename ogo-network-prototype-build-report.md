# OGO Network Local Prototype: Build Report

Status: Local staging only. No repository created on GitHub, no blockchain deployed, no token issued, no testnet launched, no spending authorized, no production change authorized.
Date: 2026-11-06
Prototype directory: `ogo-network-prototype/`
Built on: the Milestone B scaffold (`ogo-network-milestone-b/chain/`, 54 tests passing, single-node localnet proven to height 68).
Toolchain: Go 1.25.7 at `~/workspace/tools/go-1.25.7/bin/go`. Dependency matrix LOCKED as one tested unit (Cosmos SDK v0.53.6 required with Gaia's replace to v0.53.4 preserved, CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0). `go.mod` and `go.sum` are byte-identical to the Milestone B scaffold, verified by the coordinator with diff.
All values in this report are TEST-ONLY and labelled so.

## 1. Summary

The Phase 2 local prototype is built and verified. Three local validators produce blocks and agree on the same ledger. The monetary epoch hook demonstrably fires. Native OGO Coin transfers reconcile across all three validators. Geometric and hybrid issuance are both selectable from genesis configuration at the ledger level; only geometric is proven live through the epoch path. Agent batching is measured from the prototype ledger: 100 payments of 1,000 uogo cost 10,000 uogo in fees unbatched (10.00 percent per payment) and 100 uogo batched (0.10 percent per payment), a 99 percent saving. Missed epochs recover sequentially with no lump-sum mint, no silent skipping, and no duplicate issuance.

90 top-level tests pass, 0 fail. `go vet ./...` exits 0. The demonstration script runs end to end with exit 0, twice, with identical results.

This report uses Benross's four exact categories for every major feature. Implemented means code exists. Executed successfully means functionality actually ran successfully. Tested under failure means failure behaviour was deliberately tested. Not implemented means functionality remains incomplete. Passing compilation is never reported as a working feature.

## 2. Verification performed by the coordinator

The coordinator did not rely on the build agents' reports. These commands were run by the coordinator, in the prototype directory, on 2026-11-06:

- `go vet ./...` in `chain/`: exit 0.
- `go test -count=1 ./...` in `chain/`: `ok app`, `ok x/ogocoin`, 90 top-level tests pass, 0 fail. Full output saved in `logs/test-output-20261106.log`.
- `go test -count=1 -v ./x/ogocoin/`: 85 pass, 0 fail. `go test -count=1 -v ./...` total: 90 pass, 0 fail.
- `go build -o batchdemo-bin ./batchdemo` in `chain/`: built, ran, measured the batching numbers in section 6.
- `bash demo.sh` in the prototype root: exit 0, twice. Full output saved in `logs/demo-run-20261106.log` (run 2) and `logs/demo-run-20261106-run1.log` (run 1).
- `diff go.mod` against the Milestone B scaffold: identical.
- `make vet`, `make test`: pass.

## 3. Feature report (four exact categories)

### 3.1 Three-validator consensus on the same ledger

- **Implemented:** `scripts/localnet.sh` builds one shared genesis (3 validators, power 10 each) and starts 3 nodes on distinct RPC/P2P ports with persistent peers. `ogod` gained `set-genesis-validators`, `show-node-id`, `show-validator`, and `start --rpc-laddr/--p2p-laddr/--moniker/--persistent-peers`.
- **Executed successfully:** live, multiple runs. Verified demo run: VERIFY OK at heights 5, 15, 16, 17, 18. Identical app hash on all 3 nodes at the same height (for example `AAAA853A41BA4B8B0A96D801074AFE345B18B59BE93F705F1FD97BEFC3782193` at height 15), identical balances, identical supply, identical issuance. Transfer reconciled: alice 599,876,444 (600,000,000 minus 123,456 minus 100 fee), bob 400,123,456, on all 3 nodes.
- **Tested under failure:** halt-and-resume. Stopping val3 halts the chain (height stayed at 16 for 15 seconds, expected with 3 validators: 2 of 3 is exactly two thirds, no quorum). Restarting val3 resumes blocks and nodes re-agree. Full restart: all 3 nodes down 25 seconds, height continued 6 to 10/11 (not replayed), all agree, no double mint. Restart persistence is also a unit test (`TestAppRestartRestoresHeightAndEpochs`).
- **Not implemented:** staking, delegation, unbonding, slashing (keepers pinned and compiled only). Continued block production with one validator down (needs the 4-validator minimum set, a later-network target).

### 3.2 Epoch execution (the hook demonstrably fires)

- **Implemented:** epoch timer mirror in `x/ogocoin/epochs.go` (new file): fixed grid anchored at genesis, 1-indexed epochs, one tick per block, `AfterEpochEnd` and `BeforeEpochStart` hooks, `epoch_end`, `epoch_start`, and `epoch_issuance` events, `MintScheduledForEpoch`. App integration in `app/app.go` (tick at top of FinalizeBlock, `query epochs`, epoch state persisted in `data/ogocoin/app.json` alongside height).
- **Executed successfully:** live. Epoch 1 minted 1,000 uogo into `ogo1rewardspool` (supply 1,000,001,000). By the end of the demo, 2 epochs, 1,900 issued, supply 1,000,001,900. The demo epoch state shows `hook_fire_count: 1`, `total_issued: 1000` after the first boundary. Hook-fires test `TestEpochHookFiresInWiredApp` passes: advances the wired app past boundaries and asserts issuance minted and events emitted.
- **Tested under failure:** `TestEpochHaltCatchUpInWiredApp` (missed epochs tick one per block in order, no lump, no skip, no double mint, total 3,439 exact). Duplicate epoch claim fails with `ErrDoubleIssuance` and changes nothing. Hook failure does not halt the chain and is recorded in `LastHookError`. Epoch-disabled mode falls back to the legacy block path (tested).
- **Not implemented:** real SDK x/epochs wiring. This is the mirror, not the module. `chain/EPOCHS_GAP.md` records exactly what is mirrored, what is not (no keeper, store, depinject, gRPC, app-hash coverage of epoch state), the depinject bug mechanics (issues #24770, #25425, `NewAppModule` by value in v0.53.4), and the 6-step StoreUpgrade path for Milestone C. The mirror hook-fires test does not prove depinject registration of the real module; that proof is the Milestone C test, release-blocking.

### 3.3 Native coin transfers and reconciliation

- **Implemented / Executed successfully:** `tx send` and `tx burn`, `query balance/supply/params/schedule/epochs/global-supply`. Live transfer reconciled on all 3 nodes (section 3.1). Fee 100 uogo measured, broadcast 102 to 114 ms measured.
- **Tested under failure:** insufficient-balance and bad-amount rejections (Milestone B ledger tests, still passing). Restart does not re-apply transactions (height persistence fix, tested).
- **Not implemented:** signature verification (JSON transactions, not signature-checked; real x/auth arrives with Milestone C depinject wiring). Gas metering (flat fee; CheckTx reports hardcoded GasWanted 1,000). Keplr or Leap wallet integration. Live IBC relaying (escrow accounting is ledger-level only).

### 3.4 Configurable economics (geometric and hybrid)

- **Implemented:** the monetary interface `IssuanceModel` in `x/ogocoin/models.go` supports geometric decay, hybrid (base floor plus decaying bonus), and hybrid with the gentle geometric terminal taper (factor about 0.95 over the final half of the schedule), all selected through genesis configuration. Exact rational arithmetic only in consensus-critical monetary code; no floating point. Spec research parameters reproduce exactly in tests: geometric R0 = 100,000, r = 9/10, N = 110, total 999,935, dust 65, R_109 = 1; hybrid base 10,000, B0 = 60,900, N = 40, total 999,978, dust 22, first 70,900, last 11,000.
- **Executed successfully:** geometric only, live through the epoch path.
- **Not implemented:** hybrid through the app's epoch path. The epoch hook still mints the legacy geometric schedule only, and `ogod init` has no flags to express a hybrid genesis. The seam is documented in `EPOCHS_GAP.md` (hook calls `MintScheduledModel(model, epochNumber - 1)` plus CLI flags), plus two blockers that are design decisions, not wiring: the legacy schedule floors at every step while the models use exact floor(R0 times r^n) (period 9: 386 versus 387; identical for the demo's 4-period schedule, divergent from period 7 onward), and the two epoch mint entrypoints must be unified (mirror 1-indexed SDK-style, models 0-indexed model-aware; they coincide at the ledger level today, only the mirror is called by the app).

### 3.5 Agent economics (measured costs, batching)

- **Implemented:** `chain/batchdemo/main.go`, `chain/x/ogocoin/batch.go`, `BATCHING_RESULTS.md`, built as `chain/batchdemo-bin`, called by `demo.sh` step 7.
- **Executed successfully:** measured from the prototype ledger, not from the spec table. Unbatched: 100 payments of 1,000 uogo (0.001 OGO each), 100 transactions, one 100 uogo fee each, total fee 10,000 uogo, fee per payment 100 uogo, fee as share of payment 10.00 percent (1,000 basis points). Batched: 100 payments in 1 transaction, one 100 uogo fee, total fee 100 uogo, fee per payment 1 uogo, fee as share of payment 0.10 percent (10 basis points). Fee saving from batching: 9,900 uogo, 99 percent of the unbatched fee. Transfer fee and broadcast time measured live (100 uogo fee, 102 to 114 ms).
- **Tested under failure:** not applicable at this scope; the batching demo is a measurement, not a failure test. The fee model is flat per transaction with no gas metering, so the measurement proves the batching arithmetic under a flat fee, not under a gas-priced fee market.
- **Not implemented:** a gas-priced fee market. Fee sponsorship. Throughput and latency criteria P2, P3, P4 (not run by this prototype).

### 3.6 Treasury accounting

- **Implemented:** the R-F rule in exact integers in `chain/x/ogocoin/treasury.go`: spend_t = min(D, max(Trail_4(t), floor(0.05 * B_(t-1))), max(0, B_(t-1) - F)).
- **Executed successfully:** treasury tests pass, including the low-inflow binding proof: at inflow 100 per period, D = 1,500 and D = 2,500 both spend exactly 1,900 over 20 periods, balance ends 10,100, never below the 10,000 floor.
- **Tested under failure:** the low-inflow binding is itself the failure-mode proof: raising desired spend does not raise actual spend when inflow is low. Parameter validation rejects bad treasury parameters.
- **Not implemented:** on-chain treasury governance, multisig, treasury spending proposal flow. Those are later milestones.

### 3.7 Network independence from the OGO application

- **Implemented / Executed successfully:** the localnet runs with zero OGO-application involvement (no Supabase, Worker, or benross-lab/OGO code path exists in the chain). The demo ran end to end with the application untouched. The chain is the authority for balances and supply; the application is a client of the network, not the authority controlling it. Stated in the README, `MODULE_BOUNDARIES.md`, the localnet script, and the demo script.

### 3.8 Automated monetary tests

- **Implemented:** 90 top-level tests, 0 fail (54 Milestone B, 13 new Milestone 2 epoch tests, 23 new Milestone 3, 4, and 5 tests). `go vet ./...` exits 0.
- **Executed successfully:** full suite run by the coordinator with output saved in `logs/test-output-20261106.log`.
- **Tested under failure:** duplicate issuance prevention, halt recovery (sequential missed epochs, no lump, no skip, no double mint), chain halt and recovery, treasury low-inflow binding, insufficient-balance and bad-amount rejections, restart persistence.
- **Not implemented:** the Milestone C test requirements in `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md` section 4 (real x/epochs hook-fires test, S2 bypass re-run on the wired chain, S1 signature verification, 30-minute S4 halt catch-up, hybrid through the epoch path, strict M = G + S_actual as the default, the single rounding rule, P2/P3/P4 throughput and latency, S5 slashing and liveness, S7 wasm sandbox, clean-copy reproducible build for the prototype).

## 4. Demonstration results

`demo.sh` run by the coordinator, twice, exit 0 both times. Full output in `logs/demo-run-20261106.log`.

- Step 1: build and start 3 validators. Genesis valid: max_supply 1,000,010,000, total 1,000,000,000, denom `uogo`. Genesis accounts: alice 600,000,000, bob 400,000,000 (TEST-ONLY). All 3 nodes producing blocks. VERIFY OK at height 5: identical app hash, balances, supply, issuance on all 3 nodes.
- Step 2: first epoch boundary. Epoch state on val1: `current_epoch: 2`, `hook_fire_count: 1`, `total_issued: 1000`, `issued_by_period: 0:1000`, implementation honestly labelled "epoch-timer-mirror (not SDK x/epochs); see EPOCHS_GAP.md".
- Step 3: transfer 123,456 uogo from alice to bob, fee 100 uogo. Transaction hash `0EBE96AD3FF1A482EDD769485472E48DD015AEC809C7B57BA11F61353E67C748`, code 0. Broadcast took 102 ms.
- Step 4: same balances from all three nodes: alice 599,876,444, bob 400,123,456.
- Step 5: verify the same ledger state on all three nodes at height 15: identical app hash `AAAA853A41BA4B8B0A96D801074AFE345B18B59BE93F705F1FD97BEFC3782193`, alice 599,876,444, bob 400,123,456, supply 1,000,001,000, issued 1,000. VERIFY OK.
- Step 6: halt-and-resume. Val3 stopped at height 16; height stayed at 16 for 15 seconds (HALT OK, no quorum with 2 of 3). Val3 restarted; height advanced past 16 (RESUME OK); val3 caught up to height 17. VERIFY OK at height 17: identical app hash, balances, supply, issuance.
- Step 7: agent batching demo. Unbatched 10,000 uogo total fee (10.00 percent per payment); batched 100 uogo total fee (0.10 percent per payment); saving 9,900 uogo (99 percent).
- Demo complete. Localnet stopped, no processes left running.

## 5. Failure report

Actual failures encountered during the build, not just successes. Nothing in this section was smoothed over.

1. **Vet failure on integration (coordinator-found).** After the monetary-interface agent finished, `go vet ./...` failed: `cmd/ogod/main.go:365` called `writeGenesis` with 2 arguments, but the consensus agent had changed the signature at line 468 to 3 arguments (adding `ogocoin.GenesisState`). The call at line 505 was updated; the call at line 365 was not. Fixed by the consensus agent before handoff; the coordinator verified vet exit 0 on the final tree. Lesson: two agents editing the same file need a merge check before either claims vet passes.
2. **Liveness script bug (agent A).** `local n="$1" pidfile=...` in one `local` statement expands before assignment under `set -u`, which also broke `wait_height`'s deadline. Fixed by splitting the declarations; liveness then passed.
3. **Verify sampling race, real finding (agent A).** Live queries read sequentially across nodes can straddle an epoch boundary (val1 read issued 0, val2 and val3 read issued 1,000 a second later). The nodes agreed; the comparison method was flawed. Fixed with a retry-until-agreement loop for live values; the same-height header app-hash check stays strict.
4. **Stale-node re-init failure, agent A's own mistake (agent A).** The agent ran `stop` from the wrong directory, leaving an old val3 running while `init` wiped the data directory. The stale node kept answering RPC while its replacement died on the bind, and the demo failed with connection-refused. Fixed with an init guard (refuses if any pid is alive or any RPC port answers) and wait-for-exit in `stop` and `stop_one`. The demo was re-run clean twice after.
5. **App-hash misconception, corrected by measurement (agent A).** The agent first wrote that the block header app hash lags one block. A controlled test (transaction committing in block 10; header hash at 10 changed and equaled `abci_info` hash at 10) proved the header hash is the post-block state. The comment was corrected; the verify comparison remains valid.
6. **Mid-work validator-count reversal, 4 to 3 (agent A).** The 4-validator localnet, transfer, S5-style liveness (height 56 to 60 with val4 down), and full restart had all passed when Benross decided 3. The agent rebuilt everything for 3 and re-proved it all (verify, transfer, halt-and-resume, full restart, demo). The 4-validator evidence is superseded and is not claimed.
7. **`math.Int` 256-bit overflow panic (agent B).** Found in the exact-power computation and fixed with `math/big.Rat`. The panic is recorded as a found failure, not hidden.
8. **Strict M = G + S_actual is NOT the default genesis path (agent B).** `ValidateGenesis` still allows headroom above G + S_actual. The strict rule exists as `ValidateGenesisExact` and `ComputeMaxSupply`. Flipping the default is a Milestone C item. Recorded as a deviation, not fixed by silently changing the default.
9. **Exact versus stepwise floor diverge (agent B).** The spec formula floors the exact power; the legacy `AmountForPeriod` floors at every step. Period 7: 478 exact versus 477 stepwise. Period 9: 387 versus 386. Spec parameters: 999,935 exact versus 999,572 stepwise, with stepwise hitting 0 at period 93. Identical for the demo's 4-period schedule. Milestone C must pick one rule.
10. **Taper numbers do not match the spec's rounded claim (agent B).** Measured renormalized final payment 4,360 uogo (not about 5,000), maximum single-period drop 9.28 percent (not 6.5 percent). The measured numbers supersede the rounded claim.
11. **Misleading ledger comment (coordinator-found).** `ledger.go` said "hook-fires is NOT implemented in this prototype". That was true of the model-aware entrypoint (not called by the app) but misleading overall, since the mirror hook is wired and proven to fire. Fixed by the coordinator: the comment now states the mirror hook is wired and proven by `TestEpochHookFiresInWiredApp`, and that real SDK x/epochs wiring remains Milestone C.
12. **Batching demo binary path mismatch (coordinator-found).** `demo.sh` step 7 checked for `chain/batchdemo` as a file, but that path is the Go package directory; the binary was built at `chain/batchdemo/batchdemo`. Step 7 skipped in the first verified demo run. Fixed by the coordinator: the binary is built as `chain/batchdemo-bin` and `demo.sh` calls that path. Step 7 ran in the second verified demo run.
13. **Em dash in `chain/DEPENDENCIES.md` (coordinator-found).** The title carried an em dash, violating the OGO copy rule. Fixed. Pre-existing em dashes remain in `x/ogocoin/bypass.go` (code comments, not documentation), left alone.
14. **Reproducibility hash mismatch (Milestone B, carried forward).** The first clean-copy hash comparison mismatched; the proven cause is that wasmvm CGO embeds the absolute GOMODCACHE path, which `-trimpath` does not rewrite. A rebuild with the canonical (default) GOMODCACHE was byte-identical (`5794dc37...`). The reproducible build was not re-run for this prototype; that is a Milestone C item.
15. **First demo run step 7 skipped.** The batching demo binary did not exist yet when the first verified demo ran, so step 7 printed that it was not present and skipped. This is recorded here rather than re-dated: the first run proved steps 1 to 6; the second run proved all seven steps.

## 6. What is NOT yet built

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
- A clean-copy reproducible build re-run for the prototype.

Each item is a Milestone C item, listed with its owner and unblock condition in `ogo-network-architecture/OGO_NETWORK_PROTOTYPE_BLOCKERS.md`.

## 7. Evidence

- Test output: `logs/test-output-20261106.log` (90 pass, 0 fail).
- Demo output: `logs/demo-run-20261106.log` (run 2, all 7 steps), `logs/demo-run-20261106-run1.log` (run 1, steps 1 to 6, step 7 skipped).
- Node logs: `logs/node-val1-20261106.log` (5,971 lines), `logs/node-val2-20261106.log` (5,934 lines), `logs/node-val3-20261106.log` (6,107 lines).
- Batching results: `chain/BATCHING_RESULTS.md`.
- Epoch gap: `chain/EPOCHS_GAP.md`.
- Module boundaries: `chain/MODULE_BOUNDARIES.md`.
- Dependency matrix: `chain/DEPENDENCIES.md` (matrix byte-identical to Milestone B).

## 8. Authorization boundaries

- Built in local staging only. The repository is initialized locally (`git init`, no remote, no push). Remote creation awaits Benross's explicit approval.
- No commit or push, no publishing or deployment, no public testnet, no token issuance or distribution, no spending, no change to production infrastructure. `~/workspace/ogo/` is untouched.
- No monetary parameter was selected. All values are TEST-ONLY.
