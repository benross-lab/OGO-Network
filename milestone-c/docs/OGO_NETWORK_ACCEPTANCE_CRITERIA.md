# OGO Network Acceptance Criteria Matrix

Status: Planning only. OGO-Network repository created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. No blockchain deployed, no token issued, no testnet launched, no spending authorized, no production change authorized.
Date: 2026-11-06
Companion documents: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (sections 2.6, 9.2, 9.4), `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, `OGO_CONSENSUS_AND_EPOCH_SPEC.md`, `OGO_MONETARY_ENGINE_SPEC.md`, `OGO_NATIVE_WALLET_SPEC.md`, `OGO_NETWORK_API_SPEC.md`, `OGO_NETWORK_SECURITY_MODEL.md`
Reconciliation: this matrix reconciles plan v1.0 section 9.2 with the proof-of-concept criteria in plan section 2.6 and the roadmap. Where the plan and the roadmap disagreed, the resolution is recorded in `OGO_NETWORK_RECONCILIATION_LOG.md`. The matrix itself states the reconciled gates.

## 1. How to read this matrix

### 1.1 Status vocabulary

Every criterion carries exactly one status, from this vocabulary, and no other wording counts as a pass:

- **Not implemented:** the code, test, or procedure does not exist yet. Nothing is claimed.
- **Implemented:** the code exists and compiles, or the procedure is written. This is not a pass. Implemented alone never satisfies a criterion.
- **Executed successfully:** the test or procedure ran on the wired chain or localnet, with recorded evidence (log, transaction hash, block height, query output, or report), and the observed result matched the pass criterion exactly.
- **Tested under failure:** the failure-mode test ran with recorded evidence and the system behaved as specified under the failure (rejection, jailing, halt, rollback, or catch-up, as the criterion states). A criterion that names a failure mode is only satisfied by this status, never by a success-path run alone.

### 1.2 Evidence rule

Passing compilation must never be reported as a working feature. A criterion is satisfied only by recorded evidence from an executed run: the command run, the block height, the transaction hash, the balances or supply figures observed, and the file or log where the evidence lives. A claim without evidence is recorded as Not implemented. A test that ran but whose evidence was not recorded is recorded as Implemented, not as Executed successfully.

### 1.3 Release-blocking

Release-blocking means the named next step does not start until the criterion is satisfied with recorded evidence. A blocked step is not started early on the basis of partial results. SHOULD criteria are never release-blocking; they are recorded as gaps with an owner and a date.

### 1.4 Shared tests, one owner, no contradictory gates

Stages share tests, but each criterion has exactly one owning stage. A failure blocks its owning stage and the framework-selection decision, and nothing else. In particular: a failure of a CosmWasm capability check (S7, P3) blocks framework selection, not the M2 monetary prototype, because M2 must not depend on CosmWasm. Correction note, 2026-11-06: O3 (IBC) and O4 (Keplr or Leap) were removed from the framework-selection criteria by Benross's decision (Phase 2 Execution Directive) and are no longer blocking gates of any stage; they are evaluated separately if a specific technical dependency demonstrates necessity. A failure of S2, S3, or S4 blocks M2 and therefore everything after it. The gate map in section 8 states the ownership for every shared criterion, so no two stages can disagree about whether a gate passed.

### 1.5 Benross non-negotiables

These apply to every stage, and each is mapped to the criteria that prove it:

| Non-negotiable | Proved by |
|---|---|
| Consensus: all validators agree on the same ledger state | M1 identical balances from all three nodes; M3 continued block production with 3 of 4 validators; every stage, the every-block invariant |
| Epoch execution: the hook demonstrably fires | M2 hook-fires test (mandatory, section 4); S3 exact once-only issuance; S4 halt and catch-up |
| Native coin: transfers work, balances reconcile across validators | M1 transfer and balance agreement; M2 invariant 2; M4 balance verification from chain state; M5 verified on-chain payment |
| Configurable economics: geometric and hybrid both run from genesis configuration | M2 geometric and hybrid runs with simulator test vectors as golden tests; no engine redesign to switch models |
| Agent economics: actual measured transaction costs, including batching | Framework proof of concept P2 and P3 measured costs; M5 measured agent-economics check (section 7) |
| Recovery: missed epochs processed sequentially, without duplicate issuance | M2 S4 halt and catch-up; M2 duplicate-claim test; M3 rollback (M5) |
| Independence from the OGO application: the application is a client, not the authority | M1 balances verifiable by independent queries; M4 wallet works when the OGO website is unavailable; M5 payment verification against a public RPC endpoint, never a client claim |

## 2. Stage 1: M1 localnet proof

Owner: M1. Depends on: M0 approval (Benross approves the framework recommendation and the proof-of-concept specification). All numbers TEST-ONLY and labelled so. Validator count: 3 (decided by Benross, 2026-11-06, Phase 2 Execution Directive; the reconciliation pass had resolved it to 4, which Benross's decision supersedes for the local prototype; see the reconciliation log).

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| M1-1 | Three independent local validator nodes, each on a separate home directory and a separate key, plus genesis accounts | Node list with home directories, keys, and genesis accounts; all three nodes produce blocks | Yes, blocks M2 |
| M1-2 | A test transfer of OGO Coin commits with the correct balances | Transaction hash, sender and recipient balances before and after, fee charged | Yes, blocks M2 |
| M1-3 | The same balances are verified from all three nodes by independent queries | Balance query output from each of the three nodes, all identical | Yes, blocks M2 |
| M1-4 | Total supply matches the expected value exactly | Supply query output compared with the expected value, exact integer match | Yes, blocks M2 |
| M1-5 | A proof log records the block height, the transaction hash, and the balances from each node | The proof log file, with all three nodes represented | Yes, blocks M2 |
| M1-6 | Signed, sequence-checked transactions through x/auth replace the scaffold unsigned JSON transactions; an invalid signature or a wrong sequence is rejected before execution | Rejected transaction attempts with the reason logged, state unchanged | Yes, blocks M2 |
| M1-7 | The chain is usable without the OGO application: balances are verifiable by independent queries, not only through the application | Independent query output with no application involvement | Yes, blocks M2 |

## 3. Stage 2: Framework-selection proof of concept

Owner: the framework decision. Runs on the candidate stack (Cosmos SDK plus CometBFT, continuing from the Milestone B scaffold). The proof of concept has 21 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1, O2, O5). Correction note, 2026-11-06: O3 and O4 were removed from the framework-selection criteria by Benross's decision (Phase 2 Execution Directive); Keplr and IBC are not prerequisites for the initial monetary prototype unless a specific technical dependency demonstrates necessity. Each MUST criterion is listed with its owning milestone in section 8. A MUST failure blocks final framework selection until the failure is fixed or the framework is re-evaluated. SHOULD criteria are recorded as gaps with owners and dates, and do not by themselves block selection if the MUST criteria pass.

### 3.1 MUST criteria, performance

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| P1 | 4 validators produce blocks continuously for 24 hours, no halt longer than 5 minutes except a planned halt test, block time within 1 to 10 seconds for at least 95 percent of blocks | 24-hour run log, block time distribution | Yes, blocks framework selection (owned by M3) |
| P2 | Sustained at least 200 simple bank transfers per second for 10 minutes, end to end from submission to commit, zero failed transactions attributable to the chain | Load test log with measured throughput and measured fee per transaction | Yes, blocks framework selection |
| P3 | Sustained at least 50 transactions per second for 10 minutes under mixed load (transfers, staking operations, governance votes, one CosmWasm contract call per block on average), p95 finality under 10 seconds | Mixed-load test log with measured throughput, finality, and fee per transaction. Capability check: verifies the stack CAN run CosmWasm under load; it does not make M2 depend on CosmWasm | Yes, blocks framework selection |
| P4 | p95 time from submission to final commit under 6 seconds for simple transfers during the P2 run | Latency measurements from the P2 run | Yes, blocks framework selection |
| P5 | A new node joins the 4-validator localnet at height 10,000 or later, syncs from genesis or a snapshot, and catches up to tip within 2 hours on the standard hardware baseline (4 plus CPU cores, 32 GB RAM, 500 GB NVMe) | Join and sync log with heights and times | Yes, blocks framework selection |

### 3.2 MUST criteria, security

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| S1 | A transaction naming a sender it is not signed by is rejected before execution, with the reason logged | Rejected attempt with reason logged, state unchanged | Yes, blocks framework selection |
| S2 | All 54 Milestone B tests pass on the wired chain, plus new tests for the wired keepers; no discretionary mint through bank, staking, distribution, gov, or wasm; no governance change to max supply or schedule; no wasm mint; zero bypasses succeed, state unchanged after each attempt | Full test run output, bypass attempt log | Yes, blocks framework selection (owned by M2) |
| S3 | Issuance for each epoch equals exactly R_n, claimed exactly once per epoch, total issued never exceeds total scheduled; re-claim of the same epoch fails | Per-epoch issuance log, issued-by-epoch set query, failed re-claim | Yes, blocks framework selection (owned by M2) |
| S4 | Halt the chain for 30 minutes, restart, missed epochs tick one per block in epoch order, no lump mint, no skipped epoch, no double mint; total issued after catch-up equals the sum of R_n for all elapsed epochs | Halt and restart log, per-block issuance record | Yes, blocks framework selection (owned by M2) |
| S5 | Take 1 of 4 validators offline for longer than the downtime window, observe downtime jailing and the documented slashing, chain continues with the remaining 3 validators, then unjail and rejoin | Downtime, jailing, slashing, and rejoin log | Yes, blocks framework selection (owned by M3) |
| S6 | A double-sign attempt (or simulated equivocation evidence path) is punished by tombstoning or the configured double-sign penalty, with the evidence recorded | Evidence record and penalty log | Yes, blocks framework selection (owned by M3) |
| S7 | A test CosmWasm contract attempting a mint, a bank send above its balance, and an unknown method call fails all three, state unchanged; a legitimate bank send from the contract succeeds and preserves total supply. Capability check: verifies the stack CAN execute CosmWasm safely; it does not make M2 depend on CosmWasm | Contract deployment and call log, state before and after | Yes, blocks framework selection |
| S8 | At least one validator runs with a remote signer (cosmos/kms v0.1.0 or the documented equivalent), consensus key never present as a plaintext file on the validator host | Remote signer configuration and key-location check | Yes, blocks framework selection (owned by M3) |

### 3.3 MUST criteria, maintenance and upgrades

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| M1 | Clean-copy reproducible build from a fresh module cache exits 0, vet 0, tests pass, binary hash matches the documented hash under the canonical GOMODCACHE condition; any difference investigated and explained | Build log and hash comparison | Yes, blocks framework selection |
| M2 | Governance upgrade through Cosmovisor on all 4 validators, with a store migration adding one new state field: upgrade height reached, binaries swapped, migration applied, chain resumes, balances and total supply identical before and after, total downtime under 30 minutes | Upgrade log, balances and supply before and after | Yes, blocks framework selection (owned by M3) |
| M3 | After the M2 upgrade, the S2 bypass suite is re-run; a governance upgrade attempting to change max supply or schedule is still rejected, a non-monetary upgrade is still accepted | Post-upgrade bypass run output | Yes, blocks framework selection (owned by M3) |
| M4 | Documented effort to move the matrix from the 0.53 family to the 0.55 family in a branch, with build, vet, and test results recorded; pass is a completed dry run with a written effort estimate, not a successful upgrade | Branch, build and test output, effort estimate | Yes, blocks framework selection |
| M5 | After a failed upgrade attempt on one validator (wrong binary or bad migration, in staging), roll that validator back to the prior binary and state, rejoin, and catch up to tip; exact steps and time recorded | Rollback procedure log with steps and time | Yes, blocks framework selection (owned by M3) |

### 3.4 MUST criteria, operating requirements

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| O1 | 4 validators operated as if independent (separate homes, keys, monitoring), at least one validator behind sentries, written runbook for start, stop, upgrade, snapshot, and restore, each step executed at least once | Runbook with execution records | Yes, blocks framework selection |
| O2 | Block height, block time, missed blocks per validator, peer count, mempool size, and total supply visible on a dashboard or log query for the full 24-hour run, with an alert fired and recorded for at least one missed-block event | Dashboard or query output, alert record | Yes, blocks framework selection |
| O3 | DEFERRED by Benross's decision, 2026-11-06 (Phase 2 Execution Directive). The criterion as written (IBC channel opened to a second local Cosmos chain, TEST-ONLY transfer, escrow on the source equals the transferred amount, voucher on the destination equals the transferred amount, global economic supply unchanged, return transfer releases the escrow exactly) is retained as the definition of what a live IBC test must show, but it is NOT a framework-selection gate and NOT a prerequisite for the initial monetary prototype. IBC is evaluated separately when a specific technical dependency demonstrates necessity. | Channel, transfer, escrow, voucher, and return-transfer records, when the evaluation happens | No, deferred |
| O4 | DEFERRED by Benross's decision, 2026-11-06 (Phase 2 Execution Directive). The criterion as written (a Keplr or Leap wallet, or the chain-registry metadata for one, connects to the localnet, displays a TEST-ONLY balance, signs a transfer, and the transfer commits with correct balances) is retained as the definition of what an external-wallet compatibility test must show, but it is NOT a framework-selection gate and NOT a prerequisite for the initial monetary prototype. Keplr is evaluated separately when a specific technical dependency demonstrates necessity. The wallet product is built in M4, where O4 is re-tested at product level (M4-10). | Wallet connection and transfer record, when the evaluation happens | No, deferred |
| O5 | Stop one validator, delete its data directory, restore from a snapshot, rejoin and catch up to tip within 2 hours; snapshot size and time recorded | Recovery log with size and time | Yes, blocks framework selection |

Note on validator counts in this section: P1, S5, M2, and O1 are specified for a 4-validator network, because S5 (take 1 of 4 offline, chain continues with the remaining 3) cannot be executed on a 3-validator network, where taking one validator offline halts the chain for lack of quorum. The M1 localnet proof runs 3 validators by Benross's decision. DECIDED by Benross, 2026-11-06 (Option 1): the proof-of-concept run uses a separate 4-validator network, in addition to the unchanged 3-validator local prototype; S5 is NOT redefined as a halt-and-resume demonstration, because the 3-validator halt-and-resume result is a different behavior and is not fault tolerance. The separate 4-validator proof of concept ran on 2026-11-06; its results are recorded in section 3.6. Distinctions: 3 is the approved Phase 2 local prototype; 4 is the separate framework-selection fault-tolerance test; 50 or more is a proposed decentralization target, not an approved launch parameter.

### 3.5 SHOULD criteria (recorded, not blocking)

7 validators instead of 4, with stake deliberately uneven (largest validator under one third of voting power); a 7-day continuous run; a load test with 1,000 TPS simple transfers to find the actual ceiling; a second IBC counterparty; a governance proposal executed end to end (submit, vote, pass, execute a non-monetary parameter change); a treasury multisig simulation if the treasury module exists by then; an upgrade dry run to SDK 0.55 actually completed if M4 shows the effort is small. Each SHOULD result is recorded with an owner and a date whether it passes or not.

### 3.6 Four-validator proof-of-concept results, 2026-11-06

Owner: the framework decision. Run: a separate, isolated four-validator network (`ogo-network-poc-4validator/`, chain-id `ogo-poc-4val-1`, RPC ports 26757, 26767, 26777, 26787), built from the same Milestone B scaffold and the same locked dependency matrix as the three-validator prototype, which was left unchanged. All values TEST-ONLY. Full report: `ogo-network-poc-4validator/ogo-network-poc-4validator-report.md`. Statuses use the vocabulary in section 1.1.

| ID | Status | Result |
|---|---|---|
| P1 | Partially executed, OPEN | 24-hour continuous run not completed in one session. Observed windows: 149 s and 144 s continuous, with block-time distributions over 60 s windows: 159 intervals total, 100 percent within 1 to 10 seconds (run 4: 41 intervals, min 1.015 s, median 1.019 s, p95 2.048 s, max 2.056 s; run 5: 40 intervals, min 1.015 s, median 1.023 s, p95 2.067 s, max 2.096 s). Caveat: 1 s sampling quantizes intervals to about 1 s or 2 s, so p95 and max near 2.1 s are measurement artifacts. The 24-hour run remains open; the run script supports repeated `blocktimes` sampling over a longer observation. Disposition, decided by Benross 2026-11-06: OPEN until the full 24-hour continuous observation is completed. |
| S5 | Liveness PASSED (established); jailing and slashing open for Milestone C | Voting power 10 per validator, total 40, equal weight, confirmed from genesis and from the CometBFT `/validators` RPC on all four nodes, before and after the outage. With val4 down, remaining power 30/40 (3/4), above the greater-than-two-thirds quorum. Blocks continued with val4 down in every run: run 5, height 18 before, samples 18, 22, 26, 30 at 0, 10, 20, 30 s; val4 RPC confirmed DOWN during the outage. Restart: val4 caught up to height 31 (val1: 31); post-outage verify OK at height 31 on all four nodes, identical app hash, alice 599,876,444, bob 400,123,456, supply 1,000,002,710, issued 2,710; voting power unchanged after the outage. Jailing, slashing, and unjailed rejoin were NOT executed and NOT claimed: the scaffold has no such machinery (x/staking and x/slashing pinned and compiled only, not wired, blocker B6). Restart-and-rejoin is the executable part; unjailing as a staking operation is not. Full S5 as written remains open for Milestone C. Disposition, decided by Benross 2026-11-06: liveness ACCEPTED as passed and recorded as ESTABLISHED, not provisional. Five runs demonstrated the liveness result, two complete end to end. Downtime jailing, slashing, and unjailed rejoin remain open for Milestone C. |
| M2 | Not executed, OPEN and framework-selection blocking | No Cosmovisor binary installed, no upgrade command in `ogod`, no `SetUpgradeHandler`, `UpgradeHandler`, or `StoreUpgrades` anywhere in the chain code, x/upgrade not among the pinned imports, x/gov pinned-compiled only. No fake Cosmovisor or pretend upgrade was built. The only genuine partial path is the ledger-level `ApplyUpgrade` monetary-rejection rule in `x/ogocoin/bypass.go`, which is unit-tested; it proves a rule about what an upgrade may not do, not that an upgrade can run. M2 as written requires the wired SDK app and remains open for Milestone C. Disposition, decided by Benross 2026-11-06: OPEN and FRAMEWORK-SELECTION BLOCKING until a genuine Cosmovisor upgrade and state migration are executed successfully. This is the deciding issue; the requirement is not waived because the consensus test succeeded. |
| O1 | Partially executed, partial evidence accepted; remainder open | Runbook written (`ogo-network-poc-4validator/RUNBOOK.md`); start, stop, snapshot, and restore executed for real (run 4: snapshot 748K in 6 s, restore OK in 7 s, caught up to height 79; run 5: snapshot 756K in 3 s, restore OK in 5 s, caught up to height 76; final verifies OK). Gaps stated: no sentry topology demonstrated (not demonstrable on a single-machine localnet), validator independence is single-host only, and the upgrade step was not executed (depends on M2). Disposition, decided by Benross 2026-11-06: the demonstrated start, stop, snapshot, and restore procedures are ACCEPTED as partial evidence. Sentry topology, independent validator operation, and the upgrade procedure remain open. |

What the PoC does and does not prove: it demonstrates framework-selection fault tolerance on the candidate stack (a 4-validator network continues producing blocks with one validator offline and rejoins it cleanly). It is not a mainnet readiness claim, and it does not prove the jailing, slashing, upgrade, or 24-hour criteria, which remain open as stated above. Decision, 2026-11-06: Benross accepted the fault-tolerance result as ESTABLISHED, not provisional. A successful experiment is not a failure because the larger checklist has more work, and the larger checklist is not complete prematurely. Overall framework selection remains open, with M2 the deciding blocking criterion.

## 4. Stage 3: M2 monetary prototype

Owner: M2. Depends on: M1 passed. CosmWasm independence is a gate, not an assumption: every criterion below must pass with CosmWasm disabled or unused by the monetary code.

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| M2-1 | All 54 Milestone B tests pass on the wired chain, plus new tests for the wired keepers | Full test run output | Yes, blocks M3 |
| M2-2 | Proof-of-concept criteria S2, S3, and S4 pass (see section 3.2) | Test and run logs per S2, S3, S4 | Yes, blocks M3 |
| M2-3 | The 11 supply invariants in `OGO_MONETARY_ENGINE_SPEC.md` section 10 pass on the wired chain | Invariant test output, all 11 | Yes, blocks M3 |
| M2-4 | No discretionary mint exists through any path (bank, staking, distribution, gov, wasm, founder, module account, off-schedule, dust top-up) | Bypass attempt log, zero successes, state unchanged | Yes, blocks M3 |
| M2-5 | **Hook-fires test (mandatory, release-blocking).** In the wired application, with depinject, advance the chain past an epoch boundary and observe the x/monetary `AfterEpochEnd` hook executing: issuance minted, event emitted. A compile check is not sufficient, and a unit test of the keeper alone is not sufficient. This test exists because of the depinject hook-registration bug (SDK issues 24770 and 25425): a module wired but whose hook never fires is a silent monetary failure, with zero issuance looking like the schedule working correctly | Wired-application run log showing the hook firing, the issuance minted, and the event emitted at the epoch boundary | Yes, blocks M3 |
| M2-6 | StoreUpgrade test: x/epochs added to the scaffold through a store upgrade, migration tested on the proof-of-concept chain before use for any real schedule | StoreUpgrade test log | Yes, blocks M3 |
| M2-7 | Duplicate-claim test: claim epoch n, then attempt to claim epoch n again in the same block and in a later block; both fail, total issued unchanged | Duplicate-claim attempt log | Yes, blocks M3 |
| M2-8 | Configurable economics: geometric and hybrid both run from genesis configuration, with the optional gentle terminal taper available for hybrid; switching models changes a genesis value, not the engine; simulator test vectors used as golden tests for x/monetary | Two genesis-configured runs (one geometric, one hybrid) with per-period amounts matching the simulator vectors exactly; no engine code change between runs | Yes, blocks M3 |
| M2-9 | Three-number accounting: M equals G plus S_actual, computed at genesis by exact integer summation; dust left permanently unissued | Genesis validation output, dust query | Yes, blocks M3 |
| M2-10 | CosmWasm independence: all M2 criteria above pass with CosmWasm disabled or unused by the monetary code; no contract is required for any issuance, balance, or supply result | M2 run record stating the CosmWasm state during the run, and the same criteria passing without any contract deployed | Yes, blocks M3 |

## 5. Stage 4: M3 validator prototype

Owner: M3. Depends on: M2 passed.

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| M3-1 | A 4-validator localnet (7 as the SHOULD target) with staking, delegation, unbonding, slashing, and remote signing | Localnet configuration and wired-module list | Yes, blocks M4 and the post-M3 CosmWasm adoption evaluation |
| M3-2 | Proof-of-concept criteria P1, S5, S6, S8, M2, M3, and M5 pass (see sections 3.1 to 3.3) | Test and run logs per criterion | Yes, blocks M4 |
| M3-3 | A validator can be jailed for downtime and unjailed, and rejoins | Jailing, unjailing, and rejoin log | Yes, blocks M4 |
| M3-4 | A double-sign is punished (tombstoning or the configured penalty) | Evidence and penalty log | Yes, blocks M4 |
| M3-5 | A Cosmovisor upgrade completes with identical balances and total supply before and after | Balances and supply before and after, upgrade log | Yes, blocks M4 |
| M3-6 | A failed upgrade on one validator is rolled back and the validator rejoins and catches up | Rollback log with steps and time | Yes, blocks M4 |
| M3-7 | Continued block production with 3 of 4 validators while one is offline (liveness with one faulty validator; 4 is the minimum operational set for M3, not the M1 count, which is 3 by Benross's decision) | Block production log during the outage | Yes, blocks M4 |

## 6. Stage 5: M4 wallet prototype

Owner: M4. Depends on: M3 passed. Acceptance is the wallet verification checklist in `OGO_NATIVE_WALLET_SPEC.md` section 17, in full, plus O4 at product level.

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| M4-1 | A mnemonic created in the OGO wallet can be imported into Keplr and recovers the same address and balance, and vice versa | Two-direction import test record | Yes, blocks M5 |
| M4-2 | A transaction signed in the wallet commits with the correct balances, and the fee is exactly as displayed before signing | Signed transaction record, fee displayed versus fee charged | Yes, blocks M5 |
| M4-3 | A transaction with a wrong signature, a wrong sequence, or an insufficient balance is rejected before execution | Rejection log, state unchanged | Yes, blocks M5 |
| M4-4 | The wallet displays a balance only from chain state, and flags a discrepancy between two RPC endpoints rather than silently displaying one value | Multi-source balance check record, discrepancy flag demonstrated | Yes, blocks M5 |
| M4-5 | A transaction is not displayed as completed until it is committed in a block | Pending versus committed display record | Yes, blocks M5 |
| M4-6 | No seed phrase or private key is transmitted to any server, in any log, or in any error report, verified by inspecting network traffic and logs during wallet operations | Traffic and log inspection record | Yes, blocks M5 |
| M4-7 | An agent authorization with a spending limit, a per-transaction limit, an approved recipient list, and an expiry is enforced: a transaction above the limit, to an unapproved recipient, or after expiry is refused, and the refusal is logged | Refusal log for each of the three violation types | Yes, blocks M5 |
| M4-8 | Revocation takes effect immediately: after revocation, an agent-initiated transaction is refused even if the authorization has not expired | Revocation and refusal record | Yes, blocks M5 |
| M4-9 | The wallet works when the OGO website is unavailable: a user can recover from a mnemonic and view a balance and send a transaction using only a public RPC endpoint and an independent wallet, with no OGO server involved | Offline-website test record | Yes, blocks M5 |
| M4-10 | Proof-of-concept criterion O4 passes at product level: chain-registry metadata correct, Keplr and Leap transfers in both directions commit with correct balances, staking and governance voting work from the external wallet or are documented as not yet supported, with a date | O4 product-level test record | Yes, blocks M5 |

## 7. Stage 6: M5 application integration

Owner: M5. Depends on: M4 passed. The OGO application is a client of the network, not its authority, in every criterion below.

| ID | Criterion | Evidence required | Release-blocking |
|---|---|---|---|
| M5-1 | An end-to-end payment request flow is verified on the localnet: request, sign, commit, verify, mark paid | Flow record with request id, transaction hash, block height, and paid marking | Yes, blocks M6 |
| M5-2 | The Worker must-never list is tested and no violation exists (Worker never holds keys, signs, writes chain state, broadcasts, trusts client claims, writes to the platform ledger as a side effect of a chain operation, acts as a validator, or moves treasury funds) | Must-never test record | Yes, blocks M6 |
| M5-3 | A duplicate webhook or duplicate request cannot fulfil twice (idempotency tested, keyed on the transaction hash or the payment request id) | Duplicate delivery test record, single fulfilment | Yes, blocks M6 |
| M5-4 | The economy dashboard shows three separate sections (Coins, Credits, native OGO Coin), each labelled with its source, with no combined balance shown | Dashboard record, three sections labelled | Yes, blocks M6 |
| M5-5 | Wallet linking is verified by signature (nonce, account, address, chain id), and unlinking removes the link without affecting the on-chain balance | Linking and unlinking test record | Yes, blocks M6 |
| M5-6 | Agent payment permissions are enforced: a payment request from an agent is approved or declined by the user, approval prepares an unsigned transaction only, and no funds move before the user signs in their own wallet | Agent payment flow record | Yes, blocks M6 |
| M5-7 | Marketplace settlement is gated on verified on-chain payment: an order is not marked as paid until the payment is verified as committed on-chain against a public RPC endpoint, never on a client claim; a failed, expired, or wrong-amount payment does not mark the order paid | Settlement test record, including a wrong-amount case refused | Yes, blocks M6 |
| M5-8 | Agent economics, measured: actual transaction costs are recorded for agent payments at the fee levels in `OGO_FEE_MARKET_SPEC.md` section 5 (100, 1,000, and 10,000 uogo), including a batched case (100 payments of 0.001 OGO in one transaction, one fee), with the fee as a share of each payment stated; no fee level is assumed, and the result states whether small agent payments are viable at each level | Measured cost table with batched and unbatched cases | Yes, blocks M6 |
| M5-9 | Developer payment APIs work as specified: create a payment request, check payment status, receive a webhook only after on-chain verification; the API holds no keys, signs nothing, and broadcasts nothing | API test record | Yes, blocks M6 |

## 8. Gate map: shared criteria, one owner each

| Criterion | Owner stage | Also referenced by | Failure blocks |
|---|---|---|---|
| P1 | M3 | Framework selection | M3, framework selection |
| P2, P4, P5 | Framework selection | None | Framework selection |
| P3, S7 | Framework selection (capability check) | Post-M3 CosmWasm adoption evaluation (which re-tests the eight criteria in plan section 9.4) | Framework selection; does not block M2 |
| S1 | Framework selection | None | Framework selection |
| S2, S3, S4 | M2 | Framework selection | M2, framework selection |
| S5, S6, S8 | M3 | Framework selection | M3, framework selection |
| M1, M4 | Framework selection | None | Framework selection |
| M2, M3, M5 (maintenance) | M3 | Framework selection | M3, framework selection |
| O1, O2, O5 | Framework selection | None | Framework selection |
| O3 | DEFERRED by Benross's decision, 2026-11-06. Not a framework-selection gate, not a prerequisite for the initial monetary prototype. Evaluated separately if a specific technical dependency demonstrates necessity. | M7 testnet (live IBC to real external chains) | Nothing, until the evaluation is triggered |
| O4 | DEFERRED by Benross's decision, 2026-11-06. Not a framework-selection gate, not a prerequisite for the initial monetary prototype. Evaluated separately if a specific technical dependency demonstrates necessity. | M4 (product-level external wallet integration) | Nothing, until the evaluation is triggered |
| Hook-fires test | M2 (mandatory) | None | M2, and therefore M3 and everything after it |
| 11 supply invariants | M2 | Security gates G3, G4 | M2 |
| Wallet checklist items 1 to 9 | M4 | Security gate G7 | M4 |
| Worker must-never list, idempotency | M5 | Security gates (Phase 3) | M5 |

## 9. What is not acceptance

- Compilation passing, `go vet` passing, or a module being wired is not a working feature. It is the Implemented status only.
- A test passing on the scaffold is not a pass on the wired chain. S2, the bypass suite, and the 11 invariants must be re-run on the wired chain.
- A unit test of a keeper is not a hook-fires test. The hook must be observed firing in the wired application.
- A wallet connecting to a localnet is not wallet support. O4 at product level (M4-10) is the support claim.
- An IBC escrow test at ledger level is not live IBC. O3 requires a live channel, a transfer, and a return transfer with exact escrow release.
- A fee level assumed in a model is not measured agent economics. M5-8 requires measured costs, including batching.
- A phase is complete when its acceptance criteria pass with recorded evidence. No milestone is claimed as complete on the basis of code written.
