# OGO Network Milestone C Implementation Plan v0.1

Status: Implementation plan. Planning only in this document; execution is a separate, authorized step.
Date: 2026-11-06
Authoritative decision: Benross, 2026-11-06, sequencing change. Milestone C is an IMPLEMENTATION milestone, not another documentation milestone. Planning had become the bottleneck.
Working directory (new, created at execution): `ogo-network-milestone-c/`, copied from the Milestone B scaffold, own local git init, no remote.
Companion documents: `OGO_NETWORK_PROTOTYPE_BLOCKERS.md` (B1 to B12, D1 to D6, A1 to A10), `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (sections 3.6, 4, 5, 8), `ogo-network-poc-4validator/ogo-network-poc-4validator-report.md`, plan v1.1.
All values are TEST-ONLY and labelled so.

## 1. Phase boundary

Milestone C starts from the existing scaffold. It does not rebuild what is already proven:

- The three-validator Phase 2 prototype (`ogo-network-prototype/`, commit b29f830) is complete and preserved unchanged. Its results stand as recorded. Do not modify it.
- The four-validator framework PoC (`ogo-network-poc-4validator/`) is complete as a local fault-tolerance experiment and is not rebuilt. Its fault-tolerance result is ACCEPTED as established, not provisional: S5 liveness passed. Do not modify it or retroactively change its results.
- Milestone C works in a NEW directory (`ogo-network-milestone-c/`), copied from the scaffold, with its own local git init and no remote. The prototype and PoC directories are references only.

Framework selection is NOT declared closed by Milestone C work. It stays formally open until M2 and the other mandatory criteria are satisfied. Each criterion is reported individually on completion. M2 is the deciding, framework-selection blocking criterion.

## 2. Authorization boundary

Local staging only. This plan authorizes nothing beyond planning until Benross approves Milestone C execution, and execution itself authorizes only local staging development, code changes, local execution, demonstration scripts, and automated tests. NOT authorized: GitHub remote creation, any push, publishing or deployment, public testnet, token issuance or distribution, spending or purchasing infrastructure, changes to OGO production infrastructure, changes to `~/workspace/ogo/` (the wallet track is a separate coordinator; do not touch the app).

Locked dependency matrix, unchanged: Go 1.25.7 (`~/workspace/tools/go-1.25.7/bin/go`), Gaia v28.3.1 matrix as one tested unit (Cosmos SDK v0.53.6 requirement with the replace to v0.53.4 preserved, CometBFT v0.38.23, wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0). Do not independently upgrade individual dependencies. Do NOT adopt SDK v0.53.8 or Gaia v29. SDK 0.55 and CometBFT 0.40 remain evaluation tracks only, decided separately after Milestone C.

No monetary parameter is selected in Milestone C: no max supply, genesis supply, reduction factor, floor, period length, split, fee level, treasury spend level, or reserve floor. No model family is selected (geometric vs hybrid). All such values in code and tests are TEST-ONLY and labelled so.

No em dashes in any OGO document or chat copy (commas, colons, periods, bullets; hyphens inside words are fine).

## 3. Objective

Replace the prototype scaffold with a wired Cosmos SDK application that passes the remaining acceptance criteria, in this order (Benross's order):

1. Real SDK integration: wire actual Cosmos SDK modules and depinject.
2. Epoch engine: replace the timer mirror with real x/epochs.
3. Coin security: transaction signature verification.
4. Staking: validator operations and delegation.
5. Slashing: downtime tracking, jailing, and recovery (the remaining part of S5).
6. Upgrade infrastructure: x/upgrade, Cosmovisor, and state migration (M2, framework-selection blocking).
7. Remaining acceptance work: P1, O1, hybrid issuance, strict supply accounting, the single rounding rule, gas metering, reproducible build, module rename.

## 4. Workstreams

Owners use the established OGO private agent roles. Builder implements. Tester executes the test procedure and owns the evidence. Reviewer gates every closure: no criterion is marked passed without the evidence the closure requirement names. Architect owns design decisions and the wiring plan. Researcher owns re-derivation and evidence analysis where named. No new people are invented.

### C1. Real SDK integration

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| Wire standard keepers with depinject: x/bank, x/auth, x/staking, x/slashing, x/distribution, x/gov, x/upgrade | B6 | Builder; Tester verifies; Reviewer gates | `go vet` exit 0; `go test -count=1 ./...` passes; app boots on a 4-validator localnet; every keeper reachable through its query path | All seven keepers wired and queryable; no keeper pinned-and-compiled-only; build report records the wiring |
| Re-run the 17-test bypass suite on the wired chain | B6, S2 | Tester; Reviewer gates | Execute every bypass attempt from the Milestone B suite (module accounts, governance, CosmWasm binding surface, founder mint, off-schedule mint) against the wired chain; record each attempt and the resulting state | Zero bypasses succeed; state unchanged after each attempt; bypass log filed |
| Rename x/ogocoin to x/monetary everywhere | B12 | Builder; Tester verifies | Rename module, package, store keys, and references; full test suite passes; no `ogocoin` reference remains outside historical documents | `grep` shows no live `ogocoin` reference; tests pass |

### C2. Epoch engine

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| Wire real x/epochs with depinject; execute the 6-step StoreUpgrade path in `chain/EPOCHS_GAP.md` | B1 | Builder; Architect owns the wiring plan; Tester verifies; Reviewer gates | Wire x/epochs as a first-class module; run the StoreUpgrade on the localnet; epoch state covered by app hash | x/epochs keeper, store, depinject registration, and gRPC queries all live; StoreUpgrade log filed; epoch state in app hash |
| Release-blocking depinject hook-fires test against the REAL module | B1, M2-5 | Tester; Reviewer gates | In the wired application, advance the chain past an epoch boundary and observe the x/monetary `AfterEpochEnd` hook executing: issuance minted, event emitted. Five confirmations in order: (1) epoch module initializes, (2) monetary hook is registered, (3) hook actually fires at an epoch transition, (4) exactly one scheduled issuance occurs, (5) duplicate hook execution cannot issue twice | Wired-application run log showing the hook firing, the issuance minted, and the event emitted at the epoch boundary; all five confirmations evidenced. Compilation alone is not sufficient; a unit test of the keeper alone is not sufficient |
| Halt and catch-up on the wired chain (S4) | S4 | Tester; Reviewer gates | Halt the chain 30 minutes, restart, missed epochs tick one per block in epoch order, no lump mint, no skipped epoch, no double mint | Halt and restart log; per-block issuance record; total issued after catch-up equals the sum of R_n for all elapsed epochs |
| Single rounding rule: exact vs stepwise floor | B4 | Architect recommends; Benross approves the rule, or the team decides it as an implementation detail and records it | Recompute the spec parameters under both rules; document the divergence (period 7: 478 exact vs 477 stepwise; period 9: 387 vs 386); pick one rule, remove the other | One rule in code and spec; decision recorded in the decision register; spec parameters re-verified under the chosen rule |
| Hybrid and tapered-hybrid issuance through the epoch path | B3 | Builder; Tester verifies | Wire `MintScheduledModel` into the hook; CLI flags for a hybrid genesis; two genesis-configured runs (geometric, hybrid) with per-period amounts matching the simulator vectors exactly | Both models run from genesis configuration with no engine code change between runs; golden-test vectors pass |
| Strict M = G + S_actual as the default genesis path | B5 | Builder; Tester verifies | Make `ValidateGenesisExact` the default; test proving the default path rejects headroom above G + S_actual | Default genesis rejects headroom; dust left permanently unissued; test filed |
| Taper numbers corrected to measured values | B10 | Researcher re-derives; Architect reviews | Correct the spec to the measured numbers (renormalized final payment 4,360 uogo, 9.28 percent maximum drop) or re-derive the taper with the exact parameters stated | Spec and code agree on one set of numbers; correction recorded in the decision register |

### C3. Coin security

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| Transaction signature verification (x/auth) | B2, S1 | Builder; Tester verifies; Reviewer gates | Wire x/auth with depinject; signed, sequence-checked transactions replace the scaffold unsigned JSON transactions; submit an unsigned transaction and a wrongly-signed transaction | Both rejected before execution, with the reason logged and state unchanged; valid signed transfer commits with correct balances |

### C4. Staking

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| Validator operations and delegation (x/staking) | B6, M3 | Builder; Tester verifies | Create validator, delegate, redelegate, unbond on the localnet; verify voting power changes and unbonding behavior | Operations execute against the wired keepers; voting power and delegation state queryable and correct |

### C5. Slashing (S5 remainder)

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| Downtime tracking, jailing, and recovery | B6, S5 | Builder; Tester verifies; Reviewer gates | Take 1 of 4 validators offline for longer than the downtime window; observe downtime jailing and the documented slashing; chain continues with the remaining 3 validators; then unjail and rejoin | Downtime, jailing, slashing, and rejoin log; chain continued with 3 of 4; unjailed validator rejoins and syncs; S5 marked passed only when this full procedure, not just liveness, is evidenced |

### C6. Upgrade infrastructure (M2, framework-selection blocking)

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| x/upgrade wired; Cosmovisor upgrade with store migration | M2 | Builder; Tester executes; Reviewer gates | Install Cosmovisor; run a governance upgrade through Cosmovisor on all 4 validators, with a store migration adding one new state field: upgrade height reached, binaries swapped, migration applied, chain resumes, balances and total supply identical before and after, total downtime under 30 minutes | Upgrade log; balances and supply before and after identical; downtime under 30 minutes. No fake Cosmovisor or pretend upgrade counts |
| Post-upgrade bypass re-run (M3) | M3 | Tester; Reviewer gates | After the M2 upgrade, re-run the S2 bypass suite; attempt a governance upgrade changing max supply or schedule (must be rejected) and a non-monetary upgrade (must be accepted) | Post-upgrade bypass run output; monetary-change upgrade rejected; non-monetary upgrade accepted |

### C7. Remaining acceptance work

| Item | Blocker | Owner | Test procedure | Closure requirement |
|---|---|---|---|---|
| P1 24-hour continuous observation | P1 | Tester; Reviewer gates | 4 validators produce blocks continuously for 24 hours, no halt longer than 5 minutes except a planned halt test, block time within 1 to 10 seconds for at least 95 percent of blocks | 24-hour run log and block-time distribution; 95 percent threshold met on measured data, not on the observed short windows |
| O1 sentry topology and independent operation | O1 | Tester; Architect owns the topology design; Reviewer gates | 4 validators operated as if independent (separate homes, keys, monitoring), at least one validator behind sentries, written runbook for start, stop, upgrade, snapshot, and restore, each step executed at least once | Runbook with execution records. Open question: true validator independence needs more than one host; see section 6 |
| Gas metering and a gas-priced fee market | B9 | Builder; Tester verifies | Implement gas metering; replace the hardcoded GasWanted of 1,000; run the load tests | Metered gas on transactions; fee market evidence filed; no hardcoded gas value remains |
| Load tests P2, P3, P4 | P2, P3, P4 | Tester; Builder supports | P2: 200 TPS simple transfers for 10 minutes, p95 finality under 6 seconds. P3: 50 TPS mixed load (transfers, staking operations, governance votes, one CosmWasm contract call per block on average), p95 finality under 10 seconds, after the wasm VM is instantiated. P4: p95 finality measurement from the same runs | Measured throughput, finality, and fee per transaction in test logs. P3 depends on C1 wasm work (B7) |
| Double-sign protection (S6) | S6 | Tester; Reviewer gates | Double-sign test on the wired chain with slashing wired | Double-sign detected and handled per the documented rule; log filed |
| Live wasm sandbox test (S7) | S7, B7 | Builder instantiates the wasm VM; Tester executes | Instantiate the wasm VM; run the S7 sandbox test (contract attempting a mint, a bank send above its balance, and an unknown method call fails all three, state unchanged; a legitimate bank send succeeds and preserves total supply) | Contract deployment and call log; state before and after. S7 is a capability check for framework selection; it does not make M2 depend on CosmWasm |
| Remote signer (S8) | S8 | Builder; Tester verifies | Remote signer test against the wired chain | Signer operates remotely with keys outside the validator process; test log filed |
| Reproducible build re-proven (M1) | M1, B11 | Tester | Clean-copy build with fresh caches and public proxy only; record the binary hash comparison (canonical GOMODCACHE; the wasmvm CGO absolute-path condition from Milestone B applies) | Two clean-copy builds byte-identical; hash comparison filed |
| SDK 0.55 dependency upgrade dry run (M4) | M4 | Architect; Builder supports | Dry run to SDK 0.55 with a written effort estimate | Written effort estimate; dry-run log. This is a dry run only; no adoption decision |
| Rollback (M5) | M5 | Tester | Rollback test on the localnet | Rollback log; chain resumes with correct state |
| Monitoring and alerting (O2) | O2 | Builder; Tester verifies | Monitoring and alerting demonstrated on the localnet | Alerting evidence filed |
| Snapshot recovery (O5) | O5 | Tester | Snapshot and restore executed on the wired chain (the PoC demonstrated the procedure on the scaffold; re-execute on the wired chain) | Snapshot and restore log; restored node catches up and agrees |

## 5. Sequencing and dependencies

Order of work, because later items depend on earlier ones:

1. C1 keeper wiring (B6) first: S5 remainder, S6, S2 re-run, staking, and M2 all need wired keepers.
2. C2 epoch engine (B1) with C1: the hook-fires test and S3/S4 on the wired chain need real x/epochs.
3. B4 rounding rule before B3 hybrid wiring: the schedule totals depend on the rule.
4. C3 signature verification (B2) with C1: x/auth is part of the wiring.
5. C4 staking, then C5 slashing: slashing needs staking state.
6. C6 upgrade infrastructure (M2) after C1: x/upgrade and x/gov must be wired; M3 follows M2.
7. C7 in parallel where independent: B9 gas metering after C1; P2/P3/P4 after B9 (and B7 for P3); P1 24-hour run last, on the wired chain; B11 reproducible build after the wired chain is stable; B12 rename any time tests are green.
8. B7 wasm VM: instantiate for the S7 and P3 capability checks. CosmWasm adoption itself is evaluated separately after M3 against the eight criteria in plan section 9.4 (decision D3); wiring the VM does not adopt CosmWasm.

## 6. Open questions needing Benross's call

- O1 true validator independence: the PoC demonstrated single-host operation only. Genuine independence needs validators on separate hosts, which needs infrastructure Benross has not authorized (spending, A6). Options: (a) approve multi-host staging infrastructure when Milestone C reaches O1, or (b) record O1 independence as single-host only with the limitation stated. This is not decided by this plan.
- B4 rounding rule: Architect recommends; Benross approves the rule, or the team decides it as an implementation detail and records it in the decision register.
- B10 taper numbers: correct the spec to the measured numbers, or re-derive the taper with the exact parameters stated. Researcher re-derives; the correction is recorded in the decision register.

## 7. Reporting rules for Milestone C

- Four exact categories per feature: Implemented (code exists), Executed successfully (actually ran successfully), Tested under failure (failure behavior deliberately tested), Not implemented (incomplete). Passing compilation is never reported as a working feature.
- A failure section in the build report lists every failure, however small, and what was done about it.
- Each criterion is reported individually: passed, partially executed, not executed, or open, with the evidence named. Framework selection stays formally open until M2 and the other mandatory criteria are satisfied; this plan does not close it.
- The build report records actual test output, logs, block heights, voting power, and app hashes. No claim without the log that proves it.
- Verification is independent: the coordinator re-runs tests and the localnet rather than taking any builder's report on faith. The prototype and PoC directories are verified untouched at the end.

## 8. What is NOT in Milestone C

- Monetary parameter selection (blocker D4): max supply, genesis supply, reduction factor, floor, period length, split, fee level, treasury spend level, reserve floor. Each needs Benross's explicit approval in the decision order in the decision register, section 7.
- Model family selection (geometric vs hybrid) and period length. Open decisions, not Milestone C work.
- Validator minimum for launch (D5, Definition A vs Definition B) and public-crypto identity (D6). Benross decisions.
- Legal analysis (A8, not yet commissioned, longest-lead item), security audits 1 to 5 (A9), economic review of the fee market (A10). External reviews, not Milestone C work.
- Keplr and IBC evaluation (D2, deferred until a specific technical dependency demonstrates necessity). Live IBC relaying (B8) stays deferred; no work until that dependency is named.
- CosmWasm adoption (D3, evaluated separately after M3). SDK 0.55 upgrade timing (decided separately after Milestone C passes).
- Any GitHub remote, push, publication, deployment, public testnet, token issuance, spending, or production change. Each needs separate explicit approval.
