# OGO Network: Documentation Reconciliation Report (Report B)

Status: Documentation reconciliation only. No repository created on GitHub, no code committed or pushed, no deployment, no testnet launched, no token issued, no spending, no production change.
Date: 2026-11-06
Architecture workspace: `ogo-network-architecture/`
Companion documents: plan v1.0 (`ogo-network-native-coin-architecture-implementation-plan-v1.0.md`, kept intact as the reviewed version), plan v1.1 (`ogo-network-native-coin-architecture-implementation-plan-v1.1.md`, new file), `OGO_NETWORK_RECONCILIATION_LOG.md`, `OGO_NETWORK_DECISION_REGISTER.md`, `OGO_NETWORK_ACCEPTANCE_CRITERIA.md`.
Prototype results are in the separate Report A: `ogo-network-prototype/ogo-network-prototype-build-report.md`. The two reports are independent; neither depends on the other.
No em dashes are used in this report.

## 1. Summary

Benross reviewed plan v1.0 and found it strong overall, ready for prototype preparation with corrections required before the implementation specification is locked. This report records the reconciliation: the four research findings incorporated, the three inconsistency resolutions, and the superseding decisions in his Phase 2 Execution Directive (2026-11-06). The directive formalized two parallel workstreams, Track A (local prototype build) and Track B (this reconciliation), neither pausing, replacing, or unnecessarily blocking the other.

Two corrections from the reconciliation pass were themselves superseded by Benross's decisions, and both are recorded as superseded rather than silently overwritten: the validator count resolved to 4 was decided back to 3, and O3/O4 as framework-selection compatibility checks were removed from the framework-selection criteria entirely. The reconciliation log entries 17 to 20 record the supersession.

## 2. Documents updated

| Document | Change |
|---|---|
| `ogo-network-native-coin-architecture-implementation-plan-v1.1.md` (new file) | Updated overview. v1.0 kept intact. Revision log at the top lists all ten changes. Incorporates the four findings, the three inconsistency resolutions, and Benross's directive decisions. |
| `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md` | Validator count corrected to 3 (section 3, item 9, and the dated revision note). O3/O4 removed from the proof-of-concept criteria (21 MUST, not 23). Capability-checks paragraph corrected (part b: Keplr/IBC deferred; part c: M1 owns the 3-validator proof). Hook-fires wording in section 3, item 8 strengthened to name the five confirmations. |
| `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` | M1 stage corrected to 3 validators (header, M1-1, M1-3, M1-5). Section 1.4 corrected (O3/O4 no longer blocking gates). Section 1.5 corrected (identical balances from all three nodes). Section 3 header corrected (21 MUST criteria). O3/O4 rows in section 3.4 marked DEFERRED with the criterion texts retained as definitions. Gate map in section 8 corrected (O3/O4 deferred). Note added on the 4-validator PoC criteria (P1, S5, M2, O1). M3-7 parenthetical corrected. |
| `OGO_NETWORK_ARCHITECTURE.md` | Section 9 corrected: 26 to 23 to 21 MUST criteria, with the count history recorded. |
| `OGO_VALIDATOR_AND_STAKING_SPEC.md` | Section 2, Phase 1 corrected to 3 validators, with Benross's decision and the distinction recorded. Phase 2 (4 minimum, 7 target) and Definition A/B unchanged; they describe later networks, not the local prototype. |
| `OGO_NETWORK_DECISION_REGISTER.md` | Seven Benross decisions added (section 2). Three superseded items added (section 2.1). Four recommendation rows added (section 3). One open decision added (section 4: PoC validator count). One open question added (section 8: treasury 4 versus 0 units). |
| `OGO_NETWORK_RECONCILIATION_LOG.md` | Entries 17 to 20 added: the superseding decisions (validator count back to 3, MUST count 23 to 21, Keplr/IBC removed from framework selection, PoC validator-count open question). |
| `OGO_NATIVE_WALLET_SPEC.md` | Section 16 corrected: O4 is DEFERRED by Benross's decision, no longer a framework-selection gate or a MUST criterion for M0 to M3. |
| `OGO_NETWORK_API_SPEC.md` | Section 9 corrected (O4 deferred). Section 10 corrected (O3 deferred). The supply rule is unchanged. |
| `OGO_NETWORK_PROTOTYPE_SPEC.md` (new file) | Local prototype technical specification, describing the prototype as built. |
| `OGO_NETWORK_PROTOTYPE_TASKS.md` (new file) | Local prototype task breakdown, with status in the four exact categories. |
| `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md` (new file) | Automated test requirements: the 90 passing tests, the requirements satisfied, the requirements Milestone C must add, and the known test-level divergences. |
| `OGO_NETWORK_PROTOTYPE_BLOCKERS.md` (new file) | Explicit list of remaining blockers: 12 technical (B1 to B12), 6 decision (D1 to D6), 10 approval (A1 to A10). |
| `OGO_CONSENSUS_AND_EPOCH_SPEC.md` | Verified: the four findings were already incorporated by the findings agent (hook-fires test with five ordered confirmations in section 13, cross-reference fixed). No further edit needed. |
| `OGO_TREASURY_AND_SUSTAINABILITY.md` | Verified: incorporated (desired versus sustainable spending separated, low-fee finding of 4 units out of 114,511, six-figure reporting requirement, hybrid valley assumption labelling). No further edit needed. |
| `OGO_FEE_MARKET_SPEC.md` | Verified: incorporated (requirement 8 agent micropayment viability, 1,000 uogo fee example, one-million-transactions-per-day finding, fee sponsorship evaluation). No further edit needed. |
| `OGO_MONETARY_ENGINE_SPEC.md` | Verified: incorporated (treasury finding cross-reference, taper candidate, exact arithmetic rules). No further edit needed. |

## 3. Corrections made (each with the affected document)

### 3.1 The four research findings

(a) **Epoch infrastructure.** x/epochs is a first-class Cosmos SDK module from v0.53.0, upstreamed through PR #23815. A mandatory depinject hook-fires test is now a release-blocking proof-of-concept criterion with five ordered confirmations: the epoch module initializes, the monetary hook is registered, the hook actually fires at an epoch transition, exactly one scheduled issuance occurs, and a duplicate hook execution cannot issue twice. Compilation alone is not an acceptable test. Affected documents: `OGO_CONSENSUS_AND_EPOCH_SPEC.md` (section 13), `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (M2-5), `OGO_NETWORK_IMPLEMENTATION_ROADMAP.md` (section 3, item 8), plan v1.1 (S3).

(b) **Treasury sustainability.** Under low fees, raising desired spend from 1,500 to 2,500 changes actual spending by only 4 units out of 114,511. Desired spending and sustainable spending are separated throughout. Every treasury report, scenario table, and dashboard must report six figures separately: desired treasury spending, actual permitted spending, available inflows, reserve balance, unfunded spending, and number of periods with a shortfall. A treasury spending formula is not proof of treasury sustainability. Affected documents: `OGO_TREASURY_AND_SUSTAINABILITY.md` (section 2.1), `OGO_MONETARY_ENGINE_SPEC.md` (cross-reference), `OGO_NETWORK_DECISION_REGISTER.md` (section 3), plan v1.1.

(c) **Agent fee market.** The 1,000 uogo fee example is recorded: a 0.001 OGO payment carries a 100 percent fee burden; 100 payments batched into one transaction carry 0.10 percent. Approximately one million transactions per day would be required to cover the standard validator need at the modeled fee level (about 8,200,000 per day at the 100 uogo fee level that keeps small payments viable). Agent micropayment viability is an explicit protocol design requirement. Agent-native batching is a core protocol consideration. The fee architecture must explicitly distinguish ordinary transactions from agent micropayments. Fee sponsorship is evaluated as an additional mechanism (it keeps the fee and changes who pays it; it needs caps, revocation, and abuse controls). Affected documents: `OGO_FEE_MARKET_SPEC.md` (requirement 8, sections 5 and 7), `OGO_NETWORK_DECISION_REGISTER.md` (section 3), plan v1.1.

(d) **Hybrid issuance valley.** The 26-period result is corrected: it assumes fee growth begins only after issuance ends. When fee growth begins at genesis, the valley is 0 periods. Every scenario table must identify its fee-growth start assumption. Historical reports (Decision Preparation v0.1, Treasury Sustainability v0.2) state the 26-period valley under the post-issuance-ramp assumption and are left intact as historical records. Affected documents: `OGO_TREASURY_AND_SUSTAINABILITY.md` (section 7), `OGO_NETWORK_DECISION_REGISTER.md` (section 3), plan v1.1.

### 3.2 The three inconsistency resolutions

(1) **Validator count.** The reconciliation pass resolved M1 to 4 validators with BFT reasoning. Benross decided THREE for the authorized local prototype, superseding that resolution. The corrected documents state 3 for M1, with the distinction: 3 is the decided local-prototype execution requirement, 4 is the minimum operational set, 50 or more is the decentralization target. Affected documents: roadmap, acceptance matrix, validator spec, decision register, reconciliation log, plan v1.1.

(2) **CosmWasm sequencing.** S7 and P3 are framework-selection capability checks (the stack CAN execute CosmWasm safely). The first monetary prototype (M0 to M3) does not depend on CosmWasm: M2 acceptance must pass with CosmWasm disabled or unused by the monetary code. Adoption is evaluated separately after M3 against the eight criteria in plan section 9.4. This applies Benross's preference and his directive decision. Affected documents: roadmap (section 3), acceptance matrix (section 3.4, gate map), plan v1.1.

(3) **Keplr and IBC sequencing.** The reconciliation pass made O3 and O4 framework-selection compatibility checks. Benross decided they are separated from framework-selection tests and are NOT prerequisites for the initial monetary prototype unless a specific technical dependency demonstrates necessity. O3 and O4 are removed from the framework-selection criteria and marked DEFERRED, with their criterion texts retained as definitions. The proof of concept has 21 MUST criteria (P1 to P5, S1 to S8, M1 to M5, O1, O2, O5), not 23. Affected documents: roadmap, acceptance matrix, architecture spec, wallet spec, API spec, decision register, reconciliation log, plan v1.1.

### 3.3 Other corrections

- **MUST count.** 26 was a count error (corrected to 23 in the reconciliation pass). 23 was correct while O3 and O4 were included. 21 is correct now that they are removed. The count history is recorded in every file that states the count.
- **Stale delivery language.** Plan v1.0 section 9.1 said the eleven specifications were being prepared and the directory did not yet exist. Plan v1.1 states they were delivered 2026-11-06. The roadmap was already corrected in the reconciliation pass.
- **Framework status.** Cosmos SDK plus CometBFT remains a candidate, not a finalized selection, pending the proof of concept. Stated explicitly in the roadmap, the architecture spec, and plan v1.1.
- **Per-milestone ownership.** Each proof-of-concept criterion has one owning milestone, so shared tests cannot produce contradictory gates. M1 owns the 3-validator localnet proof; M2 owns S2, S3, S4 and the 11 supply invariants; M3 owns P1, S5, S6, S8, M2, M3, M5; the remaining criteria are framework-selection criteria exercised once on the proof-of-concept localnet. Recorded in the roadmap and the acceptance matrix gate map.
- **Roadmap section 3, item 8 hook-fires wording.** Strengthened to name the five confirmations.
- **Architecture spec section 9.** Corrected from 26 to 21 MUST criteria.
- **Treasury 4 versus 0 units.** The Decision Preparation report states a 4-unit change; the treasury specification's own tables show a 0-unit difference in that model run. Both are recorded, with the substance identical (the spending rule cannot fix a low-inflow level problem). Reconcile to one number if the recomputation script is available. Recorded as an open question in the decision register, section 8.

## 4. Conflicts resolved

| Conflict | Resolution | Decided by |
|---|---|---|
| M1 validator count: plan v1.0 said 3, proof-of-concept spec said minimum 4 | M1 runs 3 (Benross decided). The PoC criteria that need 4 (P1, S5, M2, O1) are a separate question, open (section 5). | Benross, 2026-11-06 |
| CosmWasm: section 2.6 made live execution a MUST criterion, sections 3.16 and 9.4 said it is not a dependency of M0 to M3 | Capability check versus prototype dependency: S7 and P3 verify the stack CAN run CosmWasm; M2 must pass with it disabled or unused; adoption evaluated separately after M3. | Benross's preference, confirmed in the directive |
| Keplr/IBC: proof of concept required them, roadmap placed wallet and infrastructure later | Removed from framework selection entirely (DEFERRED). Evaluated separately when a technical dependency demonstrates necessity. Wallet product stays in M4; live IBC to real chains stays M7 onward. | Benross, 2026-11-06 |
| Stale delivery language: specs described as being prepared in one section, delivered in another | Aligned to delivered, 2026-11-06. | Reconciliation pass, confirmed in v1.1 |
| MUST count: 26 in plan and architecture spec, 23 in the criterion tables | 26 was a count error; 23 was correct with O3/O4 included; 21 is correct now. | Reconciliation pass, then Benross's decision |
| Hybrid valley: 26 periods quoted without its assumption | Corrected: 26 periods only if fees ramp after issuance ends; 0 periods if fees ramp from genesis. Assumption must be stated in every table. | Research finding (d) |
| Treasury: desired spending presented as if it were sustainable spending | Separated, with the six-figure reporting requirement and the low-fee binding finding. | Research finding (b) |
| Agent fees: batching treated as an application convenience | Made a core protocol consideration and an explicit design requirement. | Research finding (c) |

## 5. Decisions still open

| Decision | Blocks | Notes |
|---|---|---|
| Proof-of-concept validator count for P1, S5, M2, O1 | Framework selection, M3 | P1, S5, M2, and O1 are specified for a 4-validator network. S5 cannot run on a 3-validator network. Options: run the proof of concept on a separate 4-validator network, or redefine S5 for a 3-validator network as a halt-and-resume demonstration. Needs Benross's call. |
| Issuance model family: geometric or hybrid | Parameters, split, fee analysis, max supply | Geometric and hybrid remain the two candidates. Epoch-based is the legibility fallback. Halving is a reference only. |
| Period length D and nominal year | Issuance schedule, treasury projections, fee transition timing | A sustainability decision. Geometric ends at period 110 (9.2 years monthly, 27.5 years quarterly); hybrid ends at period 40 (3.3 years monthly, 10.0 years quarterly). |
| Validator minimum for launch: Definition A (4) or Definition B (at least 50) | Validator set, fee need, decentralization claims | Definition B targets are targets, not launch requirements. A mechanism for a flat stake distribution still needs to be named. |
| Fee policy: fee level, fee distribution, how agent transactions are priced | Post-issuance funding, validator pay | Only the High fee scenario covers the minimum need, and only at period 38 (or 48 at standard costs). A flat fee cannot serve both security and agent viability. |
| Treasury spending level D and emergency reserve floor F | Treasury sustainability | No spending policy fixes a level problem. Maximum sustainable constant spend is 909 per period geometric, 2,500 hybrid. |
| Treasury policy: R-F (recommended) or P4 (candidate) | Treasury spending rule | P4 remains a candidate, not an approved rule. P4 has a known post-issuance freeze defect. |
| Post-issuance funding mechanism and fee share | Treasury and validator funding after period N | Must be named before genesis, with revenue modelled, not assumed. Treasury staking is circular and is not a funding mechanism. |
| Public-crypto identity | Distribution, exchange admission | Whether the same asset appears in any other form, and if so, under what accounting. One issuance authority or the cap breaks. |
| Keplr and IBC evaluation trigger | O3, O4, M4 (product-level O4) | A specific technical dependency must be named and documented, or the evaluation stays deferred. |
| CosmWasm adoption | M7 testnet (CosmWasm enabled) | Evaluated separately after M3 against the eight criteria in plan section 9.4. |
| Governance model, chain id, denomination, address format, commission, unbonding, slashing parameters, wallet and explorer technology, liquid staking, bridges, epoch identifier name, SDK 0.55 upgrade timing | Various | All listed in the decision register, section 4, with their blockers. None is decided by this report. |

## 6. Recommendations awaiting approval

| Recommendation | Document | Status |
|---|---|---|
| Cosmos SDK plus CometBFT, continuing from the verified Milestone B scaffold | `OGO_NETWORK_ARCHITECTURE.md` | Recommended, not final. Final selection requires the proof of concept to pass (21 MUST criteria). |
| Substrate as the credible fallback, not the default | Plan section 2.5 | If the proof of concept fails for a structural reason, the next step is a time-boxed Substrate spike, not a full rewrite. Not authorized. |
| Epoch-based periods via x/epochs | `OGO_CONSENSUS_AND_EPOCH_SPEC.md` | Recommended. x/epochs verified as a first-class SDK module from v0.53.0. Period length D is not recommended. |
| Three custom modules only: x/monetary, x/treasury, x/rewards | `OGO_NETWORK_ARCHITECTURE.md` | Recommended. x/monetary is the production name for the scaffold's x/ogocoin. |
| Three-number accounting: M = G + S_actual, computed at genesis by exact integer summation; dust left permanently unissued | `OGO_MONETARY_ENGINE_SPEC.md` | Recommended, model-independent. |
| Gentle geometric taper (factor about 0.95 over the final 20 periods) as a hybrid candidate | `OGO_MONETARY_ENGINE_SPEC.md` | Candidate if hybrid is chosen. Short linear tapers are rejected as worse than abrupt. Note: the prototype measured a renormalized final payment of 4,360 uogo with a 9.28 percent maximum drop, not the spec's rounded claim of about 5,000 with a 6.5 percent maximum drop. The measured numbers supersede the rounded claim. |
| R-F treasury rule | `OGO_TREASURY_AND_SUSTAINABILITY.md` | Recommended default. Transition keyed to the genesis-fixed final issuance period N. |
| Fixed final period in genesis for geometric, so the end date is predictable | `OGO_MONETARY_ENGINE_SPEC.md` | Recommended, not a decision. |
| CosmWasm evaluated after M3, not a dependency of the first monetary prototype | Plan section 9.4 | Recommended. Eight pass/fail criteria must pass before adoption. |
| Validator tiers: lean 75 USD, standard 200 USD, professional 500 USD per validator per month, infrastructure only | `OGO_VALIDATOR_AND_STAKING_SPEC.md` | Recommended as the cost basis for planning. Labour is separate. |
| Fee market hybrid starting point: static base fee plus congestion-adjusted priority fee | `OGO_FEE_MARKET_SPEC.md` | Recommended for evaluation, not a decision. Must be tested under load before it is chosen. |
| Mandatory depinject hook-fires test (five confirmations) | `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (M2-5) | Release-blocking for the monetary prototype. |
| Treasury six-figure reporting requirement | `OGO_TREASURY_AND_SUSTAINABILITY.md` (section 2.1) | Recommended as a reporting rule. |
| Agent micropayment viability as an explicit protocol design requirement; agent-native batching as a core protocol consideration | `OGO_FEE_MARKET_SPEC.md` (requirement 8) | Recommended as a design requirement. |

## 7. Outstanding technical, legal, and security questions

### Technical

- **PoC validator count (D1).** See section 5. The only new open decision created by this reconciliation.
- **Exact versus stepwise floor.** The spec formula floors the exact power; the legacy `AmountForPeriod` floors at every step. They diverge from period 7 onward (period 7: 478 versus 477; period 9: 387 versus 386). Milestone C must pick one rule and remove the other.
- **Taper numbers.** Measured 4,360 uogo and 9.28 percent versus the spec's rounded about 5,000 and 6.5 percent. Correct the spec to the measured numbers.
- **Treasury 4 versus 0 units.** Reconcile to one number if the recomputation script is available.
- **Hybrid through the epoch path.** The app's epoch hook still mints the legacy geometric schedule only. Wiring `MintScheduledModel` into the hook and adding CLI flags for a hybrid genesis is a Milestone C item.
- **Strict M = G + S_actual as the default.** `ValidateGenesis` still allows headroom; only `ValidateGenesisExact` enforces it. Flipping the default is a Milestone C item.
- **Real x/epochs wiring.** The prototype uses an epoch timer mirror. The depinject hook-registration bug cannot be exercised by the mirror. The Milestone C hook-fires test is release-blocking.
- **Signature verification, wasm VM, live IBC, gas metering, standard keepers.** All not implemented in the prototype. Each is a Milestone C item, listed in `OGO_NETWORK_PROTOTYPE_BLOCKERS.md`.

### Legal

- **Legal analysis not yet commissioned.** Securities, money transmission, tax, international distribution, exchange admission. Must be commissioned in writing before genesis distribution. Longest-lead item. Nothing in these specifications is legal advice.
- **Naming collision: OGO Token versus OGO Coin versus OGO Coins.** The native coin is OGO, ticker OGO (locked). OGO Coins is the existing Supabase account-based balance. Unresolved as a naming and communication question, not a technical one.
- **Public-crypto identity.** Whether the same asset appears in any other form, and if so, under what accounting. Must be decided before distribution. One issuance authority or the cap breaks.
- **SEC Covered User Interface statement (13 Apr 2026).** Staff-level, not a rule; 12 conditions; 5-year validity to about 13 Apr 2031; covers crypto asset securities only; does not cover issuance, sales, or swaps.
- **FinCEN FIN-2019-G001.** Administrators and exchangers are money transmitters; activity matters, not labels.

### Security

- **Security audits 1 to 5 not yet commissioned.** Ordered in `OGO_NETWORK_SECURITY_MODEL.md` section 7. No unresolved critical or high findings before mainnet.
- **Economic review of the fee market not yet commissioned.** The fee transition is the binding problem. Needed before genesis.
- **Fee-market evidence for agent transactions.** Agents are 45 percent of transactions but 7.7 percent of fee revenue in the High scenario. The fee policy must state explicitly how agent transactions are priced.
- **Bypass suite re-run on the wired chain (S2).** The 17 Milestone B bypass tests pass on the scaffold-derived prototype, but the standard SDK keepers are not wired, so the suite has not been re-run against the wired keepers.
- **Bridges.** No bridge is adopted. Any bridge is third-party trust and must be evaluated under the same supply-accounting rules as IBC. Bridges are the most exploited component class in the industry (over 2.8 billion USD lost since 2022). No canonical Base-Cosmos bridge exists.

## 8. What this reconciliation does not decide

This report does not select a monetary model, does not select any monetary parameter, does not approve a repository creation, does not approve a commit or a push, does not approve a publication or a deployment, does not approve a testnet or a mainnet, does not approve any spending, does not commission legal analysis or an audit, and does not authorise token issuance or distribution. Each of those needs a separate, explicit approval from Benross.

## 9. Verification

- The four findings were verified as incorporated by reading the four specifications (`OGO_CONSENSUS_AND_EPOCH_SPEC.md` section 13, `OGO_TREASURY_AND_SUSTAINABILITY.md` sections 2.1 and 7, `OGO_FEE_MARKET_SPEC.md` requirement 8 and sections 5 and 7, `OGO_MONETARY_ENGINE_SPEC.md` cross-reference) and confirming the key markers exist in each file.
- The validator-count corrections were verified by grep across the roadmap, acceptance matrix, validator spec, architecture spec, and plan v1.1: no remaining live "four nodes" or "all four" M1 references, no remaining live "23 MUST" references, and no remaining "O3/O4 as framework-selection gates" except in the explicitly marked DEFERRED rows. The reconciliation log entries 1, 2, 3, and 16 still record the earlier pass's resolutions (4 validators, 23 MUST criteria) as history; entries 17 to 20 supersede them.
- Plan v1.0 was not edited. It stays intact as the reviewed version. Plan v1.1 is a new file.
- No em dashes in any file written or edited in this pass (verified by grep).

## 10. Correction note, 2026-11-06 (GitHub publication)

This report predates the GitHub publication and two of its statements are superseded by it, recorded here rather than silently overwritten.

- **The 4-unit low-fee figure is superseded.** Sections 2 and 4 of this report record the low-fee finding as "raising desired spend from 1,500 to 2,500 changes actual spending by only 4 units out of 114,511". Recomputed 2026-11-06 from the stated model, the change is 0 units: both runs spend 114,511 over 120 periods. The 4-unit figure is withdrawn as not reproducible from the stated model. See `docs/discrepancy-resolutions-2026-11-06.md`.
- **The "no repository created" status line is superseded.** The header of this report states no repository was created on GitHub and no code was committed or pushed. Benross authorized the creation of the public benross-lab/OGO-Network repository and the push of the authorized scope on 2026-11-06. The rest of the header stands: no deployment, no testnet launched, no token issued, no spending, no production change.
