# OGO Network Local Prototype: Task Breakdown

Status: Task breakdown for the prototype as built, 2026-11-06, plus the remaining Milestone C tasks. Local staging only. No repository creation on GitHub, no push, no deployment, no token issuance, no spending, no production change.
Date: 2026-11-06
Prototype directory: `ogo-network-prototype/`
Companion documents: `OGO_NETWORK_PROTOTYPE_SPEC.md`, `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md`, `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`, and the prototype build report `ogo-network-prototype/ogo-network-prototype-build-report.md`.
All values are TEST-ONLY and labelled so.

## How to read this document

Each task states its owner, its acceptance, and its status. Status uses Benross's four exact categories: Implemented (code exists), Executed successfully (actually ran successfully), Tested under failure (failure behaviour deliberately tested), Not implemented (incomplete). Compilation alone is never reported as working. Milestone C tasks are the remaining work; they are not claimed as done anywhere.

## Milestone 1: Repository initialization

| Task | Owner | Acceptance | Status |
|---|---|---|---|
| Copy the Milestone B scaffold as the starting point, do not restart from scratch | Build agent A | Scaffold copied, 54 Milestone B tests still passing | Executed successfully |
| Initialize the repository locally in staging | Build agent A | `git init` in the prototype directory, no remote, no push | Implemented (initialized, no commits yet at handoff; the coordinator commits in the consolidation step) |
| Chain application, module boundaries, configuration, build pipeline | Build agent A | `chain/MODULE_BOUNDARIES.md`, `chain/Makefile`, `chain/README.md`; `go build` and `go vet` pass | Executed successfully |
| Dependency matrix LOCKED as one tested unit | Build agent A | `go.mod` and `go.sum` byte-identical to Milestone B | Executed successfully (verified by the coordinator with diff) |
| Document the validator-count distinction (3 local prototype, 4 minimum operational set, 50+ decentralization target) | Build agent A | Stated in `chain/scripts/localnet.sh`, `chain/README.md`, `chain/MODULE_BOUNDARIES.md`, `README.md`, `demo.sh` | Executed successfully |

## Milestone 2: Consensus and epoch execution

| Task | Owner | Acceptance | Status |
|---|---|---|---|
| 3-validator localnet that produces blocks and reaches consensus on the same ledger | Build agent A | `scripts/localnet.sh` builds one shared genesis (3 validators, power 10 each), starts 3 nodes on distinct RPC/P2P ports with persistent peers; `ogod` gains `set-genesis-validators`, `show-node-id`, `show-validator`, `start --rpc-laddr/--p2p-laddr/--moniker/--persistent-peers` | Executed successfully |
| Identical ledger state verified on all 3 nodes | Build agent A | Same app hash, balances, supply, and issuance at the same height, on all 3 nodes, at heights 5, 15, 16, 17, 18 in the verified demo runs | Executed successfully |
| Epoch timer mirror with `AfterEpochEnd` and `BeforeEpochStart` hooks | Build agent A | `chain/x/ogocoin/epochs.go`, app integration in `chain/app/app.go` (tick at top of FinalizeBlock), `query epochs`, epoch state persisted in `data/ogocoin/app.json` | Executed successfully |
| Hook-fires test (mirror form) | Build agent A | `TestEpochHookFiresInWiredApp` passes: advances the wired app past boundaries, asserts issuance minted and events emitted | Executed successfully |
| Halt catch-up test (mirror form) | Build agent A | `TestEpochHaltCatchUpInWiredApp` passes: missed epochs tick one per block in order, no lump, no skip, no double mint, total 3,439 exact | Tested under failure |
| Duplicate epoch claim rejected | Build agent A | Duplicate claim fails with `ErrDoubleIssuance` and changes nothing | Tested under failure |
| Hook failure does not halt the chain | Build agent A | Failure recorded in `LastHookError`, chain continues | Tested under failure |
| Halt-and-resume on the live localnet | Build agent A | Stopping val3 halts the chain (height stays, no quorum with 2 of 3); restarting resumes blocks and nodes re-agree; full restart continues from persisted height, no replay, no double mint | Tested under failure |
| Real SDK x/epochs wiring (keeper, store, depinject, gRPC, app-hash coverage) | Milestone C | `EPOCHS_GAP.md` records the gap and the 6-step StoreUpgrade path | Not implemented |
| Milestone C hook-fires test (depinject registration of the real module) | Milestone C | Five confirmations in `OGO_CONSENSUS_AND_EPOCH_SPEC.md` section 13, release-blocking | Not implemented |

## Milestone 3: Native OGO Coin accounting

| Task | Owner | Acceptance | Status |
|---|---|---|---|
| Denomination, balances, transfers, deterministic issuance interfaces | Build agent B | `uogo`, 6 decimals (proposed); `tx send`, `tx burn`, `query balance/supply/params/schedule/epochs/global-supply` | Executed successfully |
| Transfer reconciled across all 3 validators | Build agent A | alice 599,876,444, bob 400,123,456 on all 3 nodes; fee 100 uogo measured; broadcast 102 to 114 ms measured | Executed successfully |
| Common issuance interface with geometric, hybrid, and tapered-hybrid models | Build agent B | `IssuanceModel` in `chain/x/ogocoin/models.go`; model enum on Params; genesis validation; exact rational arithmetic only, no floating point | Implemented |
| Geometric issuance live through the epoch path | Build agent A and B | Proven live in the demo (epoch 1 minted 1,000 uogo; 2 epochs, 1,900 issued) and in tests | Executed successfully |
| Hybrid and tapered-hybrid through the epoch path | Milestone C | `MintScheduledModel` wired into the hook; CLI flags for a hybrid genesis; one mint entrypoint (resolve the exact versus stepwise floor divergence first) | Not implemented |
| Exact versus stepwise floor divergence resolved | Milestone C | One rounding rule chosen; the other removed; spec parameters re-verified | Not implemented |
| Strict M = G + S_actual as the default genesis path | Milestone C | `ValidateGenesisExact` becomes the default; `ValidateGenesis` headroom removed | Not implemented |
| Signature-verified transactions through x/auth | Milestone C | Invalid signature or wrong sequence rejected before execution (M1-6) | Not implemented |
| Rename x/ogocoin to x/monetary | Milestone C | Rename applied everywhere, tests still passing | Not implemented |

## Milestone 4: Agent transaction economics

| Task | Owner | Acceptance | Status |
|---|---|---|---|
| Batching demo with measured numbers from the prototype ledger | Build agent B | `chain/batchdemo/main.go`, `chain/x/ogocoin/batch.go`, `BATCHING_RESULTS.md`; unbatched 10,000 uogo total fee (10.00 percent per payment), batched 100 uogo total (0.10 percent per payment), 99 percent saving | Executed successfully (verified by the coordinator running the binary) |
| Batching demo binary built and called by `demo.sh` step 7 | Coordinator | `chain/batchdemo-bin` built; `demo.sh` calls it; step 7 ran in the verified demo run 2 | Executed successfully |
| Transfer fee and broadcast time measured live | Build agent A | Fee 100 uogo, broadcast 102 to 114 ms, in the demo log | Executed successfully |
| Gas-priced fee market | Milestone C | Gas metering, gas price, congestion adjustment; P2, P3, P4 throughput and latency criteria run | Not implemented |
| Fee sponsorship mechanism | Milestone C | Caps, revocation, abuse controls, sponsored-fee accounting | Not implemented |

## Milestone 5: Automated monetary tests

| Task | Owner | Acceptance | Status |
|---|---|---|---|
| 90 top-level tests pass, 0 fail | Build agents A and B | `go test -count=1 ./...`, verified by the coordinator (85 in x/ogocoin, 5 in app) | Executed successfully |
| Duplicate issuance prevention | Build agent B | Same epoch cannot mint twice (ledger tests, epoch tests, app tests) | Tested under failure |
| Halt recovery, sequential missed epochs | Build agents A and B | Missed epochs processed sequentially, no lump-sum mint, no silent skipping, no double mint | Tested under failure |
| Treasury accounting (R-F rule) | Build agent B | Exact-integer R-F implementation; low-inflow binding proved (D = 1,500 and D = 2,500 both spend exactly 1,900 over 20 periods at inflow 100) | Tested under failure |
| Both issuance models end to end | Build agent B | Geometric and hybrid reproduce the spec research parameters exactly in tests | Executed successfully (ledger level; hybrid not yet through the epoch path) |
| Supply invariants | Build agent B | M = G + S_actual computed by summing the floored schedule; dust left unissued; invariants checked every block in tests | Executed successfully |
| Chain halt and recovery behaviour | Build agent A | Live halt-and-resume and full restart on the 3-validator localnet | Tested under failure |
| Balance reconciliation across validators | Build agent A | Same balances from all 3 nodes at the same height, in the demo and in tests | Executed successfully |
| Epoch execution ordering | Build agent A | Epochs tick one per block in epoch order after a halt | Tested under failure |

## Consolidation tasks (coordinator, after the build agents finished)

| Task | Status |
|---|---|
| Verify the build, tests, and demo by running them (not by reading the agents' reports) | Executed successfully |
| Fix the `writeGenesis` vet failure (2-arg call at `cmd/ogod/main.go:365` against a 3-arg signature) | Executed successfully (fixed by agent A; vet exit 0 verified) |
| Fix the misleading `ledger.go` comment ("hook-fires is NOT implemented in this prototype") | Executed successfully |
| Build the batching demo binary and wire it into `demo.sh` step 7 | Executed successfully |
| Remove the em dash from `chain/DEPENDENCIES.md` (OGO copy rule) | Executed successfully |
| Save demo logs, node logs, and test output as evidence in `logs/` | Executed successfully |
| Reconcile the architecture documents (v1.1, decision register, roadmap, acceptance matrix, reconciliation log, wallet and API specs) | Executed successfully |
| Write the prototype build report (Report A) and the reconciliation report (Report B) | Executed successfully |
| Commit the prototype repository locally (no remote) | See the build report |

## Milestone C tasks (remaining, not authorized by this phase)

These are the tasks the prototype leaves open. Each needs its own authorization before it starts. The full list, with owners and unblock conditions, is in `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`.

1. Real SDK x/epochs wiring with the depinject hook-fires test (release-blocking).
2. Signature-verified transactions through x/auth with depinject.
3. Hybrid issuance through the epoch path, with the floor-divergence resolution and CLI flags.
4. Strict M = G + S_actual as the default genesis path.
5. Wasm VM instantiation and live CosmWasm execution (capability check S7).
6. Live IBC relaying (O3, deferred until a technical dependency demonstrates necessity).
7. Standard SDK keepers wired (bank, auth, staking, slashing, distribution, gov, upgrade).
8. Gas metering and a gas-priced fee market (P2, P3, P4).
9. Rename x/ogocoin to x/monetary.
10. Reproducible build re-proven from a clean copy for the prototype.
11. Staking, delegation, unbonding, slashing, remote signing (M3).
12. Keplr or Leap wallet integration (O4, deferred until a technical dependency demonstrates necessity).
