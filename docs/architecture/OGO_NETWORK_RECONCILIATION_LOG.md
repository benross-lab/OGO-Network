# OGO Network Reconciliation Log

Status: Planning only. OGO-Network repository created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. No blockchain deployed, no token issued, no testnet launched, no spending authorized, no production change authorized.
Date: 2026-11-06
Scope: reconciliation pass over plan v1.0 (`ogo-network-native-coin-architecture-implementation-plan-v1.0.md`) and the eleven specifications in `ogo-network-architecture/`. Plan v1.0 was not edited (it stays intact as the reviewed version; the coordinator writes v1.1). `OGO_CONSENSUS_AND_EPOCH_SPEC.md`, `OGO_TREASURY_AND_SUSTAINABILITY.md`, `OGO_FEE_MARKET_SPEC.md`, `OGO_MONETARY_ENGINE_SPEC.md`, and `OGO_NETWORK_ARCHITECTURE.md` were not edited (another agent owns them for findings incorporation). `ogo-network-prototype/` was not touched. No em dashes are used in this log or in any file changed in this pass.

## Corrections made in this pass

### 1. Validator count for the M1 localnet proof: 3 corrected to 4
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3, item 9, plus the dated revision note at the top of the file.
- Correction: the first proof (M1) now specifies four independent local validator nodes, with the same balances verified from all four nodes, replacing three nodes.
- Reasoning: BFT consensus needs 3f+1 validators to tolerate f faulty validators. With 3 validators, f equals 0, so the network tolerates zero faulty validators and halts if any one goes down; it cannot demonstrate liveness or fault tolerance, and it cannot run the slashing and liveness test (S5), which takes 1 of 4 validators offline and requires the remaining 3 to keep producing blocks. With 4 validators, f equals 1, so one validator can go offline and the remaining 3 (more than two thirds) keep producing blocks. The epoch specification already states the minimum operational set is 4. This resolution is about the M1 proof count only; it does not decide the launch validator minimum, which remains the open decision between Definition A (4 validators) and Definition B (at least 50 validators) in the decision register.

### 2. Validator count in the staged validator strategy: 3 corrected to 4
- Affected document: `OGO_VALIDATOR_AND_STAKING_SPEC.md`, section 2, Phase 1 (Local development validators, M0 to M1).
- Correction: Phase 1 now specifies 4 independent local validator nodes, with acceptance as balances identical across all four nodes, replacing 3 nodes and all three nodes. The same BFT reasoning as entry 1 is stated in the file. Phase 2 (4 validators minimum, 7 as the target), Phase 4 (minimum operational set 4), and section 6.3 (Definition A, 4 validators) were already consistent and were not changed.

### 3. Proof-of-concept MUST criterion count: 26 corrected to 23
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 2 (deliverables) and the dated revision note.
- Correction: the proof of concept has 23 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1 to O5: 5 plus 8 plus 5 plus 5), not 26. The earlier figure of 26 was a count error. The criterion lists themselves were already correct (P1 to P5, S1 to S8, M1 to M5, O1 to O5) and were not changed.

### 4. Stale delivery language: specifications described as being prepared, corrected to delivered
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 2 (Phase 1 deliverables).
- Correction: the deliverables now state that the eleven standalone specifications have been delivered 2026-11-06 to the architecture workspace, and name all eleven. The plan's section 9.1 language saying the eleven specifications were still being prepared is superseded by delivery; the plan's companion index already describes them as delivered. The plan text itself is for the coordinator to correct in v1.1 (see entry 13).

### 5. Framework status: candidate status made explicit in the roadmap
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 1 (new framework-status paragraph), section 3 item 2 (candidate-stack qualifier), and the dated revision note.
- Correction: the roadmap now states that Cosmos SDK plus CometBFT is the recommended framework candidate, not a finalized selection, that final selection requires the proof of concept to pass, and that the dependency matrix named in the roadmap is the candidate stack under evaluation. No roadmap language now presents the framework as finalized. The architecture direction (independent Layer 1, Option B) remains confirmed and locked; only the framework choice is a candidate.

### 6. CosmWasm sequencing: capability check distinguished from prototype dependency
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3 (new capability-checks paragraph after the Phase 2 acceptance criteria) and section 5, item 4 (CosmWasm evaluation).
- Correction: criteria S7 (live wasm sandbox test) and P3 (mixed load including one CosmWasm contract call per block) are now identified as framework-selection capability checks that verify the candidate stack CAN execute CosmWasm safely. They do not make CosmWasm a dependency of the first monetary prototype: M2 acceptance (S2, S3, S4 and the 11 supply invariants) must pass with CosmWasm disabled or unused by the monetary code. The adoption decision is separate and later: after M3, CosmWasm is evaluated in staging against the eight criteria in plan section 9.4, and only if they pass is it enabled for the testnet (M7). A failed S7 blocks framework selection; a failed post-M3 adoption evaluation blocks CosmWasm adoption, not M0 to M3. This applies Benross's preference: keep the first monetary prototype independent of smart contracts, then evaluate CosmWasm separately before adopting it.

### 7. Keplr and IBC sequencing: compatibility evaluation distinguished from product implementation
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3 (capability-checks paragraph).
- Correction: criteria O4 (Keplr or Leap wallet smoke test) and O3 (live IBC smoke test to a second local chain) are now identified as framework-selection compatibility checks that verify the stack supports external wallets and IBC relaying. They are not the wallet product or production IBC: the wallet is built in M4, and live IBC to real external chains is later work (M7 testnet onward). An O4 or O3 failure blocks framework selection until fixed, and is recorded with the blocker; it does not move wallet development earlier than M4.

### 8. Per-milestone ownership of proof-of-concept criteria stated, so shared tests cannot contradict
- Affected document: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3 (capability-checks paragraph).
- Correction: ownership is now stated explicitly. M1 owns the 4-validator localnet proof; M2 owns S2, S3, S4 and the 11 supply invariants; M3 owns P1, S5, S6, S8, M2, M3, M5; the remaining criteria (P2, P3, P4, P5, S1, S7, M1, M4, O1, O2, O3, O4, O5) are framework-selection criteria exercised on the proof-of-concept localnet, recorded once and referenced by the milestones, not re-gated per milestone. A failure blocks exactly its owning stage plus framework selection.

### 9. Keplr integration in the wallet prototype scope: compatibility smoke test distinguished from the wallet product
- Affected document: `OGO_NATIVE_WALLET_SPEC.md`, section 16 (Prototype scope, M0 to M3).
- Correction: a clarification is added that O4 in M0 to M3 is a framework-selection compatibility smoke test only (chain-registry metadata plus one signed transfer from Keplr or Leap that commits with correct balances), that its failure blocks framework selection and not the monetary prototype, and that the full wallet product, including Keplr, Leap, and Cosmostation integration, is built in M4. The prototype scope itself is unchanged.

### 10. External wallet acceptance test: O4 identified as a compatibility check, not the wallet product
- Affected document: `OGO_NETWORK_API_SPEC.md`, section 9 (External wallets).
- Correction: a clarification is added that O4 is a framework-selection compatibility check and the full wallet, including Keplr, Leap, and Cosmostation integration, is built in M4. The acceptance test itself (transfers in both directions committing with correct balances) is unchanged.

### 11. Live IBC: O3 identified as a local-chain compatibility check, production IBC placed later
- Affected document: `OGO_NETWORK_API_SPEC.md`, section 10 (External chains, IBC).
- Correction: a clarification is added that O3 is a framework-selection compatibility check (live relaying between two local chains, with exact escrow release) and that live IBC to real external chains is later work, from the M7 testnet onward, not M0 to M3 product scope. The supply rule (global economic supply equals local supply minus escrow plus vouchers) is unchanged.

### 12. New file: reconciled acceptance-criteria matrix
- Affected document: `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (new file in `ogo-network-architecture/`).
- Correction: created as a new file. It gives six separate stages, (1) M1 localnet proof, (2) framework-selection proof of concept, (3) M2 monetary prototype, (4) M3 validator prototype, (5) M4 wallet prototype, (6) M5 application integration, each with its own acceptance criteria, evidence requirements, and release-blocking status. It includes the mandatory depinject hook-fires test for the monetary prototype (M2-5, release-blocking), the four-state status vocabulary (not implemented, implemented, executed successfully, tested under failure), the rule that passing compilation must never be reported as a working feature, the gate map giving each shared criterion one owner, and the mapping of Benross's non-negotiables (consensus, epoch execution, native coin, configurable economics, agent economics with measured costs including batching, recovery, independence from the OGO application) to the criteria that prove them.

## Items recorded for other owners (not edited in this pass)

### 13. Plan v1.0, section 9.2 (M1): three validators, for the coordinator to correct in v1.1
- Affected document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md`, section 9.2, M1 Localnet Proof.
- Issue: M1 still states three independent local validator nodes, a 3-validator localnet, and acceptance across all three nodes. This contradicts the resolved count of 4 (entries 1 and 2) and the proof-of-concept specification in the same plan (section 2.6), which requires a minimum of 4 validators. The plan was not edited in this pass. Recommended v1.1 change: four validators throughout M1, with the BFT reasoning from entry 1.

### 14. Plan v1.0, section 12.3 (infrastructure requirements): 3 to 4 local validator nodes, for the coordinator to correct in v1.1
- Affected document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md`, section 12.3.
- Issue: "M0 to M1 (localnet): a development machine ... with the ability to run 3 to 4 local validator nodes." The lower end of the range contradicts the resolved count of 4. Recommended v1.1 change: 4 local validator nodes.

### 15. Plan v1.0, section 9.1: eleven specifications described as being prepared, for the coordinator to correct in v1.1
- Affected document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md`, section 9.1 (Related continuation work).
- Issue: section 9.1 says the 11 standalone specification documents are being prepared and that the directory did not yet exist when the plan was written, while the plan's companion index (appendix) describes the same eleven specifications as delivered 2026-11-06. The specifications have been delivered. The roadmap was corrected in this pass (entry 4). Recommended v1.1 change: align section 9.1 with the delivered state.

### 16. Architecture specification, section 9: 26 MUST criteria, for the other agent to correct
- Affected document: `OGO_NETWORK_ARCHITECTURE.md`, section 9 (What is not finalized).
- Issue: section 9 states the proof of concept has 26 MUST criteria. The plan's section 2.6 tables list 23 (P1 to P5, S1 to S8, M1 to M5, O1 to O5). The roadmap was corrected in this pass (entry 3). This file was not edited in this pass because another agent owns it for findings incorporation. Recommended change: 23 MUST criteria.

## Consistency check results (verified, no correction needed)

The following were checked across plan v1.0 and all eleven specifications and found consistent. No change was made for any of them.

- **Native Layer 1 decision:** independent Layer 1 with native OGO Coin, Option B confirmed and locked, in the plan (ADR-1, section 9.2, decision register section 2) and in every specification header. No specification proposes a token on another chain for the native coin.
- **Separate repository architecture:** OGO-Network, OGO-Wallet, OGO-Explorer, OGO-Network-Docs, separate from benross-lab/OGO, in the plan (ADR-2) and in the architecture, roadmap, and decision register. No repository creation is authorized by any of these documents.
- **Cosmos SDK plus CometBFT candidate status:** recommended, not final, pending the proof of concept, in the plan (ADR-3, section 2.5), the architecture specification (section 9), and the decision register (section 3). The roadmap now states this explicitly as well (entry 5).
- **Epoch-based recommendation:** epoch-based periods via x/epochs, with the period identifier, duration D, and nominal year fixed in genesis, in the plan (ADR-4, sections 3.5, 9.2) and in the consensus and epoch specification and the architecture specification (section 6). Block-based periods are rejected for production in all of them; the scaffold's block-based schedule is a migration item, not a design.
- **Three custom modules:** x/monetary, x/treasury, x/rewards only, with no x/governance and no x/fees, in the plan (section 3) and in the architecture, monetary engine, and decision register. x/monetary is the production name for the scaffold's x/ogocoin everywhere.
- **Noncustodial wallet:** users own their keys, no seed or private key on any OGO server, signing only in the user's wallet, in the plan (section 6) and in the wallet, API, and security specifications.
- **Separation of platform credits and native OGO Coin:** Supabase Coins and Credits remain a separate application-level ledger, no automatic conversion, no combined balance, in the plan (ADR-5) and in the architecture, wallet, API, and decision register.
- **Geometric and hybrid issuance alternatives:** both remain candidates, no model or parameter selected, both available in the simulator and selectable through genesis configuration, in the plan (ADR-6, section 4) and in the monetary engine specification and the decision register.
- **R-F treasury recommendation:** spend equals min(D, max(Trail_4, floor(0.05 x B)), max(0, B minus F)), recommended default, transition keyed to the genesis-fixed final issuance period, in the plan (ADR-6, section 4) and in the monetary engine specification (section 7), the treasury specification (section 2), and the decision register (section 3). P4 remains a candidate, not an approved rule, in all of them.
- **Protocol revenue versus OGO company revenue:** kept separate, company revenue counted as protocol revenue only through an explicit, implemented, disclosed mechanism, in the plan, the architecture specification (section 7, boundary 5), the treasury specification (section 10), the fee market specification (section 9), and the validator specification (section 9).
- **Security and legal approval gates:** five audits with no unresolved critical or high findings, legal analysis commissioned in writing and received, genesis parameters approved by Benross in writing, launch as a separate approval, in the plan (sections 10, 11) and in the roadmap (sections 5 to 7), the security model (sections 6 and 7), and the decision register (section 6). No document claims a gate is passed.

## What could not be reconciled

- Nothing in the assigned scope was left unreconciled. The three items that could not be fixed in this pass (entries 13, 14, 15 in plan v1.0 and entry 16 in the architecture specification) are recorded above with the exact recommended change, because those files belong to the coordinator (v1.1) and to the other agent respectively. No other inconsistency was found between the plan and the eleven specifications in the items checked.

## Superseding decisions by Benross, 2026-11-06 (Phase 2 Execution Directive)

Benross issued the Phase 2 Execution Directive after the reconciliation pass above. It supersedes parts of entries 1, 2, 3, 7, 8, 13, 14, and 16. The earlier entries are kept as the record of what was resolved in that pass; the corrections below state what Benross decided and what was changed as a result. Nothing was silently overwritten: each superseded entry is named here and in the corrected files.

### 17. Validator count for the M1 localnet proof: 4 corrected back to 3 (Benross decided)
- Affected documents: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3, item 9 and the dated revision note; `OGO_VALIDATOR_AND_STAKING_SPEC.md`, section 2, Phase 1; `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, section 1.5 and section 2 (Stage 1, M1-1, M1-3, M1-5). This supersedes entries 1 and 2 above and the recommended changes in entries 13 and 14.
- Correction: the M1 localnet proof runs THREE independent local validator nodes, with the same balances verified from all three nodes. The BFT reasoning in entries 1 and 2 (a 3-validator network tolerates zero faulty validators and halts if any one goes down) is correct and is retained in the corrected text as the reason 3 is not a liveness claim, but Benross decided 3 for the authorized local prototype anyway.
- Distinction, stated in every corrected file: 3 is the decided local-prototype execution requirement; 4 is the minimum operational set (Definition A), the smallest set tolerating one faulty validator; 50 or more is the decentralization target (Definition B) for a public network.
- Plan v1.1 records 3 throughout M1 and section 12.3, restoring plan v1.0's original figure of 3, which the reconciliation pass had been about to change to 4. Entries 13 and 14 are therefore resolved in the opposite direction from their recommendations.

### 18. Proof-of-concept MUST criterion count: 23 corrected to 21 (O3 and O4 removed)
- Affected documents: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, sections 2 and 3 and the dated revision note; `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, section 3 header, section 3.4, section 8; `OGO_NETWORK_ARCHITECTURE.md`, section 9. This supersedes entry 3 above and the recommended change in entry 16.
- Correction: the proof of concept has 21 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1, O2, O5), not 23. O3 (IBC smoke test) and O4 (Keplr or Leap wallet smoke test) were removed from the framework-selection criteria by Benross's decision. The count history is: 26 was a count error; 23 was correct while O3 and O4 were included; 21 is correct now that they are removed.
- The O3 and O4 criterion texts are retained in the matrix as the definitions of what those tests must show when they happen, marked DEFERRED, not deleted.

### 19. Keplr and IBC removed from framework-selection tests (Benross decided)
- Affected documents: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3 (capability-checks paragraph, part b); `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, sections 1.4, 3.4, and 8. This supersedes entry 7 above and qualifies entries 9, 10, and 11, which described O3 and O4 as framework-selection compatibility checks.
- Correction: Keplr and IBC are separated from framework-selection tests and are NOT prerequisites for the initial monetary prototype unless a specific technical dependency demonstrates necessity. They are evaluated separately when that dependency arises. The wallet product is built in M4, where O4 is re-tested at product level (M4-10); live IBC to real external chains is later work from the M7 testnet onward. Entries 9, 10, and 11 in the wallet and API specifications still describe O4 and O3 as compatibility checks; those passages are now qualified by this entry and must be read with the deferred status, and are flagged for correction in the next specification pass (see the decision register, open item).
- The CosmWasm treatment in entry 6 is NOT superseded: S7 and P3 remain framework-selection capability checks, and the first monetary prototype remains independent of smart contracts, which is consistent with Benross's directive.

### 20. Proof-of-concept validator-dependent criteria (P1, S5, M2, O1): open question, not decided
- Affected documents: `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3; `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, section 3.4 note and section 6 (M3-7).
- Issue: P1, S5, M2, and O1 are specified for a 4-validator network, because S5 (take 1 of 4 offline, chain continues with the remaining 3) cannot be executed on a 3-validator network, where taking one validator offline halts the chain for lack of quorum. The M1 localnet proof runs 3 validators by Benross's decision. Whether the proof-of-concept run uses a separate 4-validator network or S5 is redefined for a 3-validator network as a halt-and-resume demonstration is an open question for Benross. The criteria remain specified for 4 validators and are not claimed as passed by the 3-validator local prototype. Recorded in the decision register as an open decision.

### 21. Proof-of-concept validator count: decided, Option 1 (Benross, 2026-11-06)
- Affected documents: `OGO_NETWORK_DECISION_REGISTER.md`, section 2 (new decision row) and section 4 (open-decision row removed); `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`, D1 (closed); `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, section 3.4 note and new section 3.6; `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3 (capability-checks paragraph and new item 11); `ogo-network-native-coin-architecture-implementation-plan-v1.1.md`, change summary item 3 and section 2.6 revision note. This supersedes entry 20 above.
- Decision: Benross chose Option 1, a separate, isolated four-validator framework-selection proof of concept, in addition to the completed three-validator Phase 2 prototype, which stays unchanged. The technical distinction, in his words: with three equally weighted validators, losing one leaves 2/3 voting power, insufficient for CometBFT's greater-than-two-thirds quorum; with four equally weighted validators, losing one leaves 3/4 voting power, allowing consensus to continue. The three-validator halt-and-resume result is NOT fault tolerance and must not be redefined as such.
- Distinctions, stated in every corrected file: 3 is the approved Phase 2 local prototype; 4 is the separate framework-selection fault-tolerance test; 50 or more is a proposed decentralization target, not an approved launch parameter.
- This decision does not authorize a public testnet, GitHub remote, push, deployment, spending, or token issuance.

### 22. Four-validator proof-of-concept results recorded (2026-11-06)
- Affected documents: `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, section 3.6 (new); `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3, item 11 (new); `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`, D1 (outcome recorded). New files: `ogo-network-poc-4validator/` (PoC directory with `run-poc.sh`, `README.md`, `RUNBOOK.md`, `ogo-network-poc-4validator-report.md`, `logs/`).
- Results: S5 liveness demonstrated (voting power 10 per validator, total 40; blocks continued with val4 down, height 18 to 30 over 30 s in run 5; val4 restarted and caught up to height 31; all four nodes agreeing on app hash, balances, supply 1,000,002,710, issued 2,710). S5 jailing, slashing, and unjailed rejoin NOT executed and NOT claimed (x/staking and x/slashing pinned and compiled only, not wired, blocker B6); full S5 as written remains open for Milestone C. P1 partially executed (observed windows 149 s and 144 s; 159 block-time intervals, 100 percent within 1 to 10 s; the 24-hour duration remains open). M2 not executed (no Cosmovisor binary, no upgrade command, no upgrade handler in the scaffold; no fake upgrade built). O1 partially executed (runbook written; start, stop, snapshot, and restore executed for real; no sentry topology on a single-machine localnet; upgrade step depends on M2). Tests re-verified by the coordinator: `go vet` exit 0, 90 tests pass, 0 fail; `run-poc.sh` re-run by the coordinator. Prototype directory verified untouched (git status clean, commit b29f830). PoC `go.mod` and `go.sum` byte-identical to the prototype's.

### 23. Four-validator fault-tolerance result accepted; framework selection remains open (Benross, 2026-11-06)
- Affected documents: `OGO_NETWORK_DECISION_REGISTER.md`, section 2 (new decision row); `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`, section 3.6 (statuses and dispositions for P1, S5, M2, O1); `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md`, section 3, item 11 (disposition added); `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`, D1 (disposition added). New file: `ogo-network-milestone-c-plan-v0.1.md` (Milestone C planning).
- Decision: Benross accepted the four-validator fault-tolerance result as demonstrated evidence, recorded as ESTABLISHED, not provisional. A successful experiment is not a failure because the larger checklist has more work, and the larger checklist is not complete prematurely. Dispositions: S5 liveness ACCEPTED as passed (five runs demonstrated liveness, two complete end to end); downtime jailing, slashing, and unjailed rejoin remain open for Milestone C. P1 OPEN until the full 24-hour continuous observation is completed. M2 OPEN and FRAMEWORK-SELECTION BLOCKING until a genuine Cosmovisor upgrade and state migration are executed successfully; this is the deciding issue and the requirement is not waived because the consensus test succeeded. O1 start, stop, snapshot, and restore ACCEPTED as partial evidence; sentry topology, independent validator operation, and the upgrade procedure remain open. Phase boundary: the four-validator PoC is complete as a local fault-tolerance experiment and is not rebuilt; the three-validator prototype is not modified and its results are not retroactively changed. No public testnet, GitHub remote, push, deployment, token issuance, spending, or production changes are authorized by this decision.
