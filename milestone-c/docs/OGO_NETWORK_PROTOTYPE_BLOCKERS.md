# OGO Network Local Prototype: Remaining Blockers

Status: Explicit list of what blocks the next step, 2026-11-06. Local staging only. No blocker in this list is authorized to be worked on by this document; each needs Benross's explicit approval before it starts.
Date: 2026-11-06
Prototype directory: `ogo-network-prototype/`
Companion documents: `OGO_NETWORK_PROTOTYPE_SPEC.md`, `OGO_NETWORK_PROTOTYPE_TASKS.md`, `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md`, and the prototype build report `ogo-network-prototype/ogo-network-prototype-build-report.md`.
All values are TEST-ONLY and labelled so.

## How to read this document

A blocker is something that must happen, or be decided, before a later milestone can pass. Each blocker states what is blocked, what unblocks it, and who decides. Blockers are not failures of the prototype; the prototype was scoped to Milestones 1 to 5, and these are the items that scope deliberately left out.

## Technical blockers (Milestone C)

| # | Blocker | Blocks | Unblocked by | Decided by |
|---|---|---|---|---|
| B1 | Real SDK x/epochs wiring (keeper, store, depinject, gRPC, app-hash coverage of epoch state) is not implemented; the prototype uses an epoch timer mirror | M2 monetary prototype, S3, S4, framework selection | Wire x/epochs with depinject; execute the 6-step StoreUpgrade path in `EPOCHS_GAP.md`; pass the Milestone C hook-fires test (five confirmations, release-blocking) | Benross approves Milestone C |
| B2 | Signature verification is not implemented; transactions are JSON, not signature-checked | M1-6, S1, any claim of authenticated transfers | Wire x/auth with depinject; prove an unsigned or wrongly-signed transaction is rejected before execution | Benross approves Milestone C |
| B3 | Hybrid and tapered-hybrid issuance are not wired through the app's epoch path; only geometric is proven live | M2 configurable-economics non-negotiable (both models from genesis configuration) | Wire `MintScheduledModel` into the hook; add CLI flags for a hybrid genesis; resolve the exact versus stepwise floor divergence (B4) first | Benross approves Milestone C |
| B4 | Exact versus stepwise floor divergence: the spec formula floors the exact power, the legacy `AmountForPeriod` floors at every step; they diverge from period 7 onward | B3, any claim about schedule totals | Pick one rounding rule; remove the other; re-verify the spec parameters | Benross approves the rule, or the team decides it as an implementation detail and records it |
| B5 | Strict M = G + S_actual is not the default genesis path; `ValidateGenesis` still allows headroom | M2 supply-invariant non-negotiable | Make `ValidateGenesisExact` the default; add a test proving the default path rejects headroom | Benross approves Milestone C |
| B6 | Standard SDK keepers are pinned and compiled only, not wired (x/bank, x/auth, x/staking, x/slashing, x/distribution, x/gov, x/upgrade) | S2 bypass re-run on the wired chain, M2, M3 | Wire the keepers with depinject; re-run the 17-test bypass suite on the wired chain with zero bypasses succeeding | Benross approves Milestone C |
| B7 | Wasm VM is not instantiated; no live CosmWasm execution | S7, P3 (mixed load with a contract call per block) | Instantiate the wasm VM; run the S7 sandbox test | Benross approves Milestone C; CosmWasm adoption itself is evaluated separately after M3 |
| B8 | Live IBC relaying is not wired; escrow accounting is ledger-level only | O3 (deferred) | O3 is DEFERRED by Benross's decision until a specific technical dependency demonstrates necessity. No work until that dependency is named. | Benross, when a dependency demonstrates necessity |
| B9 | Gas metering is not implemented; the fee is flat per transaction, CheckTx reports a hardcoded GasWanted of 1,000 | P2, P3, P4 (throughput and latency criteria), any fee-market claim | Implement gas metering and a gas-priced fee market; run the load tests | Benross approves Milestone C |
| B10 | Taper numbers do not match the spec's rounded claim: measured renormalized final payment 4,360 uogo with a 9.28 percent maximum drop, versus about 5,000 with a 6.5 percent maximum drop | Any claim citing the spec's taper numbers | Correct the spec to the measured numbers, or re-derive the taper with the exact parameters stated | Team correction, recorded in the decision register |
| B11 | Reproducible build from a clean copy has not been re-run for the prototype | M1 (maintenance) for the prototype | Clean-copy build with fresh caches and public proxy only; record the binary hash comparison | Team task, no approval needed beyond Milestone C |
| B12 | x/ogocoin has not been renamed to x/monetary | M2 (production module naming) | Rename everywhere; tests still passing | Team task, no approval needed beyond Milestone C |

## Decision blockers (need Benross)

| # | Blocker | Blocks | Options | Decided by |
|---|---|---|---|---|
| D1 | Proof-of-concept validator count for P1, S5, M2, O1: those criteria are specified for a 4-validator network, but the local prototype runs 3 | Framework selection, M3 | CLOSED, decided by Benross 2026-11-06 (Option 1): separate four-validator framework-selection proof of concept, in addition to the unchanged three-validator prototype. S5 liveness demonstrated 2026-11-06 (blocks continued with 1 of 4 offline, restart and sync verified); P1 partially executed (24-hour duration not completed in one session); M2 not executed (no Cosmovisor or upgrade-handler support in the scaffold); O1 partially executed (runbook and snapshot/restore executed; no sentry topology on a single machine). Full results in `ogo-network-poc-4validator/ogo-network-poc-4validator-report.md`. Disposition, decided by Benross 2026-11-06: the fault-tolerance result is ACCEPTED as established, not provisional (S5 liveness passed). S5 jailing, slashing, and unjailed rejoin remain open for Milestone C; P1 open until the full 24-hour observation is completed; M2 open and FRAMEWORK-SELECTION BLOCKING until a genuine Cosmovisor upgrade and state migration are executed successfully; O1 start, stop, snapshot, and restore accepted as partial evidence, sentry topology, independent operation, and upgrade procedure open. Overall framework selection remains open pending the remaining mandatory criteria. | Benross, decided 2026-11-06 |
| D2 | Keplr and IBC evaluation trigger: O3 and O4 are deferred until a specific technical dependency demonstrates necessity | O3, O4, M4 (product-level O4) | A technical dependency is named and documented, or the evaluation stays deferred | Benross, when a dependency is demonstrated |
| D3 | CosmWasm adoption: evaluated separately after M3 against the eight criteria in plan section 9.4 | M7 testnet (CosmWasm enabled) | Adopt if the eight criteria pass; do not adopt if they fail | Benross, after M3 |
| D4 | No monetary parameter is selected: max supply, genesis supply, reduction factor, floor, period length, split, fee level, treasury spend level, reserve floor | Genesis, fee analysis, max supply as M = G + S_actual | Each parameter needs Benross's explicit approval in the decision order in the decision register, section 7 | Benross |
| D5 | Validator minimum for launch: Definition A (4 validators) or Definition B (at least 50, Nakamoto at least 10, top 10 under 33 percent) | Validator set, fee need, decentralization claims | Definition A or Definition B, with a mechanism for a flat stake distribution named | Benross |
| D6 | Public-crypto identity: whether the same asset appears in any other form, and if so, under what accounting | Distribution, exchange admission | Same asset in two forms with one issuance authority, or different assets | Benross |

## Approval blockers (need explicit approval before they start)

| # | Blocker | Notes |
|---|---|---|
| A1 | GitHub remote repository creation for OGO-Network, OGO-Wallet, OGO-Explorer, OGO-Network-Docs | OGO-Network created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. OGO-Wallet, OGO-Explorer, and OGO-Network-Docs are not created; each creation still needs explicit approval. |
| A2 | Any commit or push | The prototype has local commit b29f830 in staging. Push to benross-lab/OGO-Network was authorized 2026-11-06 for the authorized scope only. Any other push still needs explicit approval. |
| A3 | Publishing or deploying any documentation or code | The authorized OGO Network scope was published to benross-lab/OGO-Network on 2026-11-06. Nothing else is published or deployed. |
| A4 | Public testnet or mainnet launch | Not authorized. M7 and M8 are readiness stages, not launches. |
| A5 | Issuing or distributing native OGO Coin | Not authorized. All values in the prototype are TEST-ONLY. |
| A6 | Spending money or purchasing infrastructure | Not authorized. |
| A7 | Changing existing OGO production infrastructure | Not authorized. `~/workspace/ogo/` is untouched. |
| A8 | Legal analysis (securities, money transmission, tax, international distribution, exchange admission) | Not yet commissioned. Must be commissioned in writing before genesis distribution. Longest-lead item. |
| A9 | Security audits 1 to 5 | Not yet commissioned. No unresolved critical or high findings before mainnet. |
| A10 | Economic review of the fee market | Not yet commissioned. Needed before genesis. |

## What is NOT a blocker

- The 3-validator localnet halting when one validator goes down. That is expected CometBFT behaviour with 3 validators (2 of 3 is exactly two thirds, no quorum), and it is demonstrated deliberately in the demo. It blocks nothing in this phase; continued production with one validator down is a later-network target (D1, D5).
- The epoch mirror. It is a documented stand-in for Milestone C, not a defect in the prototype's scope. It becomes a blocker only when Milestone C starts (B1).
- The treasury 4-versus-0-units discrepancy. Both figures support the same conclusion (the spending rule cannot fix a low-inflow level problem). It needs a recomputation to pick one number, not a decision.
