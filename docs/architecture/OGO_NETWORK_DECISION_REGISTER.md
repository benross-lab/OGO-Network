# OGO Network Decision Register

Status: Living register. Planning only. Decisions are recorded as they are made; nothing in this register is decided by the register itself.
Date: 2026-11-06
Updated: 2026-11-08 (hard maximum supply approved; see section 2)
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (section 11)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network launches with its own native blockchain and native OGO Coin.

## 1. Purpose

This register distinguishes five categories, and every item in OGO Network belongs to exactly one of them:

1. **Decisions explicitly made by Benross.** Decided. Recorded with the date and the words used where available. These are not revisited unless Benross revisits them.
2. **Technical recommendations.** Recommended by the research and the plan, awaiting Benross's approval. A recommendation is not a decision.
3. **Open decisions.** Not yet decided. Listed with what blocks them and what they depend on.
4. **Parameters awaiting approval.** Numbers not yet selected. Each needs Benross's explicit approval, in the decision order in section 6.
5. **Items requiring external legal or security review.** Not decidable by the team alone. They need an external review, and the review must be received before the dependent decision is made.

A decision moves between categories only by an explicit action: Benross makes it, approves a recommendation, or commissions a review. No item moves categories silently, and no item is removed from the register. A superseded item is marked superseded, with the replacement recorded, not deleted.

## 2. Decisions explicitly made by Benross

| Decision | Date | Notes |
|---|---|---|
| OGO Network is an independent Layer 1 blockchain; OGO Coin is its native currency, not an ERC-20 or SPL token | 2026-11-06 | Feasibility report option B. The conditional Base recommendation is superseded for the native coin. Guiding principle: OGO owns its monetary infrastructure, while users own their coins and private keys. |
| Option B confirmed and locked | 2026-11-06 | Do not reopen Base, Ethereum, or Solana token-first alternatives unless new evidence creates a compelling reason. |
| Blockchain lives in separate repositories: OGO-Network, OGO-Wallet, OGO-Explorer, OGO-Network-Docs | 2026-11-06 | Resolves the repo tension. OGO-Network created 2026-11-06 (public, benross-lab/OGO-Network) by Benross's explicit authorization. OGO-Wallet, OGO-Explorer, and OGO-Network-Docs are not created; each creation still needs explicit approval. |
| Fixed maximum coin supply, community governance | 2026-11-05 | Founding charter. "OGO Systems creates the initial technology. The OGO community ultimately governs the network." |
| No OGO account needed to own, transfer, or use OGO Coin | 2026-11-05 | Usable via independent wallets, exchanges, developers, and third-party services. |
| OGO Coins and Credits remain separate from OGO Coin; no automatic conversion | 2026-11-05 | Accounting must never be merged. |
| Hard maximum supply: 100,000,000,000 OGO | 2026-11-08 | APPROVED by Benross. OGO Network has an approved hard maximum supply direction of 100,000,000,000 OGO. Genesis allocation, initial circulating supply, issuance schedule, and production genesis configuration remain separate decisions. Maximum supply is not the same as initial genesis supply and is not the same as circulating supply. Genesis circulating supply G and total scheduled issuance S_actual remain undecided and must satisfy G + S_actual = 100,000,000,000 OGO. The monetary system must never permit supply to exceed this ceiling. The ceiling remains a mandatory genesis parameter with no default. Genesis configuration change and monetary protocol implementation are NOT authorized by this approval. |
| Zero reserved founder allocation; zero reserved OGO Systems allocation; no preferential private presale | 2026-11-05 | Founders may acquire OGO later under the same publicly available conditions as everyone else. OGO Systems is the founding technology developer, not the owner of the currency. |
| IBC transfers must preserve global supply accounting | 2026-11-05 | A native token escrowed on one chain and represented as a voucher on another must not be counted as two independent units of economic supply. |
| Fixed-supply protections cannot be bypassed through privileged module accounts, governance upgrades, or smart-contract interfaces | 2026-11-05 | The build and test report must include actual failures, not just a successful-looking scaffold. |
| Independent reproducibility: someone outside OGO Systems must be able to obtain the source, build the same software, run the tests, and independently verify the supply rules | 2026-11-05 | Monetary rules must be enforced by the protocol, not by trusting OGO Systems. |
| Naming: Network OGO Network, native coin OGO, ticker OGO, Wallet OGO Wallet, Explorer OGO Scan | 2026-11-05 | Locked. |
| Go 1.25.7 locked as prototype build toolchain; Gaia v28.3.1 locked as dependency reference | 2026-11-05 | Entire matrix as one tested unit. Go 1.23 kept as documented original minimum only, not a verified build target. |
| A fixed-supply network must design its eventual end of issuance and its treasury economics together | 2026-11-06 | Benross's reframing of the Decision Report. Cardano is the most important warning; Polkadot shows monetary changes affect the entire ecosystem; Solana shows the appeal of a floor, with OGO's variation making the floor temporary with a predetermined end; Ethereum shows burning is not a substitute for a funding model. |
| Validator count is THREE for the authorized local prototype | 2026-11-06 | Phase 2 Execution Directive. Distinction: 3 is the local-prototype execution requirement; 4 is the minimum operational set (smallest set tolerating one faulty validator); 50 or more is the decentralization target for a public network. Supersedes the reconciliation pass resolution of 4 for M1. |
| Separate four-validator framework-selection proof of concept (Option 1) | 2026-11-06 | Benross decided Option 1: a separate, isolated four-validator framework-selection proof of concept, in addition to the completed three-validator Phase 2 prototype, which stays unchanged. Purpose is fault tolerance, distinct from the prototype. Technical distinction: three equally weighted validators losing one leaves 2/3 voting power, insufficient for CometBFT's greater-than-two-thirds quorum; four losing one leaves 3/4, consensus continues. The three-validator halt-and-resume result is NOT fault tolerance and must not be redefined as such. The PoC executed P1, S5, M2, and O1 on 2026-11-06: S5 liveness demonstrated (blocks continued with 1 of 4 offline, restart and sync verified), P1 partially executed (24-hour duration not completed in one session), M2 not executed (no Cosmovisor or upgrade-handler support in the scaffold), O1 partially executed (runbook and snapshot/restore executed; no sentry topology on a single machine). Full results in `ogo-network-poc-4validator/ogo-network-poc-4validator-report.md`. This closes decision blocker D1. Distinctions maintained: 3 is the approved Phase 2 local prototype; 4 is the separate framework-selection fault-tolerance test; 50 or more is a proposed decentralization target, not an approved launch parameter. |
| Four-validator fault-tolerance result accepted as established; overall framework selection remains open pending remaining mandatory criteria | 2026-11-06 | Benross accepted the four-validator fault-tolerance result as demonstrated evidence, recorded as ESTABLISHED, not provisional: with one of four equally weighted validators offline, the remaining three maintained consensus at 30/40 voting power; block production continued; the stopped validator restarted, synchronized, and rejoined with matching application hashes. Five runs demonstrated the liveness result, two complete end to end. The distinction is locked in: a successful experiment is not a failure because the larger checklist has more work, and the larger checklist is not complete prematurely. Dispositions: S5 liveness ACCEPTED as passed; downtime jailing, slashing, and unjailed rejoin remain open for Milestone C. P1 OPEN until the full 24-hour continuous observation is completed. M2 OPEN and FRAMEWORK-SELECTION BLOCKING until a genuine Cosmovisor upgrade and state migration are executed successfully; this is the deciding issue and the requirement is not waived because the consensus test succeeded. O1 start, stop, snapshot, and restore procedures ACCEPTED as partial evidence; sentry topology, independent validator operation, and the upgrade procedure remain open. Phase boundary: the four-validator PoC is complete as a local fault-tolerance experiment and is not rebuilt; the three-validator prototype is not modified and its results are not retroactively changed. Remaining acceptance work carries into Milestone C planning with explicit owners, test procedures, and closure requirements. No public testnet, GitHub remote, push, deployment, token issuance, spending, or production changes are authorized by this decision. |
| Keplr and IBC are separated from framework-selection tests and are NOT prerequisites for the initial monetary prototype unless a specific technical dependency demonstrates necessity | 2026-11-06 | Phase 2 Execution Directive. O3 and O4 removed from the framework-selection criteria (21 MUST criteria, not 23). Evaluated separately when a dependency demonstrates necessity. |
| The first monetary prototype is independent of smart contracts; CosmWasm is evaluated separately before adoption, with the decision and its implications documented | 2026-11-06 | Phase 2 Execution Directive. Confirms the reconciliation treatment: S7 and P3 remain framework-selection capability checks; M2 must pass with CosmWasm disabled or unused by the monetary code. |
| The epoch hook-fires test is a critical-path requirement for the monetary prototype, with five conditions: epoch module initializes, monetary hook is registered, hook actually fires at an epoch transition, exactly one scheduled issuance occurs, duplicate hook execution cannot issue twice | 2026-11-06 | Phase 2 Execution Directive and the v1.0 review. Compilation alone does not satisfy it. Release-blocking for the monetary prototype. |
| Prototype reporting uses four exact categories per feature: Implemented (code exists), Executed successfully (actually ran successfully), Tested under failure (failure behavior deliberately tested), Not implemented (incomplete) | 2026-11-06 | Phase 2 Execution Directive. Passing compilation must never be reported as a working feature. |
| Two parallel workstreams: Track A local prototype build, Track B architecture reconciliation; neither pauses, replaces, or unnecessarily blocks the other | 2026-11-06 | Phase 2 Execution Directive. Two independent final reports: Report A (prototype results) and Report B (documentation reconciliation). |
| No em dashes in OGO copy and chat | Standing | Use commas, colons, periods, and bullets. Hyphens inside words are kept. |

### 2.1 Superseded items

| Item | Superseded by | Date |
|---|---|---|
| Conditional Base recommendation for the native coin (feasibility report) | Architecture direction decision (Option B, L1 native coin) | 2026-11-06 |
| Maximum supply is TBD, but is a mandatory genesis parameter with no default (2026-11-05) | Hard maximum supply approved: 100,000,000,000 OGO (2026-11-08). The TBD part is superseded; the mandatory-genesis-parameter-with-no-default requirement is preserved by the new decision. | 2026-11-08 |
| Zero-inflation / mint-disabled-after-genesis rule (implementation report section 3) | Monetary Constitution v0.1 (fixed max plus scheduled issuance) | 2026-11-05/06 |
| Repo-first rule for the network (all work in benross-lab/OGO only) | Separate repositories decision | 2026-11-06 |
| Go 1.23 as the prototype toolchain (blueprint proposal) | Go 1.25.7 locked (Milestone B authorization, actually used and verified) | 2026-11-05 |
| Block-based periods as the schedule architecture (scaffold) | Epoch-based periods recommendation (pending confirmation in Milestone C) | 2026-11-06 |
| M1 validator count resolved to 4 (reconciliation pass) | Benross's decision of THREE for the local prototype (Phase 2 Execution Directive) | 2026-11-06 |
| O3 and O4 as framework-selection compatibility checks (reconciliation pass) | Benross's decision: Keplr and IBC separated from framework-selection tests, DEFERRED until a specific technical dependency demonstrates necessity | 2026-11-06 |
| Proof-of-concept MUST count 23 (reconciliation pass) | MUST count 21, after O3 and O4 were removed by Benross's decision | 2026-11-06 |

## 3. Technical recommendations (awaiting Benross approval)

| Recommendation | Document | Status |
|---|---|---|
| Cosmos SDK plus CometBFT, continuing from the verified Milestone B scaffold | `OGO_NETWORK_ARCHITECTURE.md` | Recommended, not final. Final selection requires the proof of concept to pass. |
| Substrate as the credible fallback, not the default | Plan section 2.5 | If the proof of concept fails for a structural reason, the next step is a time-boxed Substrate spike, not a full rewrite. Not authorized. |
| Epoch-based periods via x/epochs | `OGO_CONSENSUS_AND_EPOCH_SPEC.md` | Recommended. x/epochs verified as a first-class SDK module from v0.53.0. Period length D is not recommended. |
| Three custom modules only: x/monetary, x/treasury, x/rewards. No x/governance, no x/fees | `OGO_NETWORK_ARCHITECTURE.md` | Recommended. x/monetary is the production name for the scaffold's x/ogocoin. |
| Three-number accounting: M = G + S_actual, computed at genesis by exact integer summation; dust left permanently unissued | `OGO_MONETARY_ENGINE_SPEC.md` | Recommended, model-independent. |
| Gentle geometric taper (factor about 0.95 over the final 20 periods) as a hybrid candidate | `OGO_MONETARY_ENGINE_SPEC.md` | Candidate if hybrid is chosen. Short linear tapers are rejected as worse than abrupt. |
| R-F treasury rule: spend = min(D, max(Trail_4, floor(0.05 x B)), max(0, B - F)) | `OGO_TREASURY_AND_SUSTAINABILITY.md` | Recommended default. B is the treasury balance at the end of the previous period. Transition keyed to the genesis-fixed final issuance period N. |
| Fixed final period in genesis for geometric, so the end date is predictable | `OGO_MONETARY_ENGINE_SPEC.md` | Recommended, not a decision. |
| CosmWasm evaluated after M3, not a dependency of the first monetary prototype | Plan section 9.4 | Recommended. Eight pass/fail criteria must pass before adoption. |
| Validator tiers: lean 75 USD, standard 200 USD, professional 500 USD per validator per month, infrastructure only | `OGO_VALIDATOR_AND_STAKING_SPEC.md` | Recommended as the cost basis for planning. Labour is separate (1,500 to 4,000 USD per validator per month at the professional tier). |
| Hub precedents as proposals: 21-day unbonding, 5 percent commission floor, downtime 0.01 percent plus 10 minute jail, double-sign 5 percent plus tombstoning | `OGO_VALIDATOR_AND_STAKING_SPEC.md` | Proposals, not decisions. Each needs approval. |
| Fee market hybrid starting point: static base fee plus congestion-adjusted priority fee | `OGO_FEE_MARKET_SPEC.md` | Recommended for evaluation, not a decision. Must be tested under load before it is chosen. |
| Protocol revenue kept separate from OGO company revenue | `OGO_TREASURY_AND_SUSTAINABILITY.md` | Recorded principle from the continuation commission. Company revenue must not be counted as protocol revenue unless an explicit, implemented, disclosed mechanism routes a portion into the network. |
| Mandatory depinject hook-fires test for the monetary prototype: epoch module initializes, monetary hook is registered, hook actually fires at an epoch transition, exactly one scheduled issuance occurs, duplicate hook execution cannot issue twice | `OGO_NETWORK_ACCEPTANCE_CRITERIA.md` (M2-5), `OGO_CONSENSUS_AND_EPOCH_SPEC.md` | Release-blocking. x/epochs is a first-class Cosmos SDK module from v0.53.0, upstreamed through PR #23815. Compilation alone is not an acceptable test. |
| Treasury reporting separates desired spending from sustainable spending: desired spending, actual permitted spending, available inflows, reserve balance, unfunded spending, and number of periods with a shortfall, reported separately | `OGO_TREASURY_AND_SUSTAINABILITY.md` | A treasury spending formula is not proof of treasury sustainability. Under low fees, raising desired spend from 1,500 to 2,500 changes actual spending by 0 units out of 114,511 (both runs spend 114,511; resolved 2026-11-06, see section 8). |
| Agent micropayment viability is an explicit protocol design requirement; agent-native batching is a core protocol consideration; the fee architecture distinguishes ordinary transactions from agent micropayments; fee sponsorship evaluated as an additional mechanism | `OGO_FEE_MARKET_SPEC.md` | Evidence: a 1,000 uogo fee makes a 0.001 OGO payment cost 100 percent in fees, while batching 100 payments makes it 0.10 percent. About 1,000,000 transactions per day would be required to cover the standard validator need at the modeled fee level. |
| Hybrid issuance valley: every scenario table identifies the fee-growth start assumption. 26 periods / 116,065 shortfall assumes fee growth begins only after issuance ends; 0 periods when fee growth begins at genesis | `OGO_TREASURY_AND_SUSTAINABILITY.md` | The 26-period figure without the assumption stated is incorrect. Historical reports state it under the post-issuance-ramp assumption and are left intact as historical records. |

## 4. Open decisions

| Decision | Blocks | Notes |
|---|---|---|
| Issuance model family: geometric or hybrid | Parameters, split, fee analysis | Geometric and hybrid remain the two candidates. Epoch-based is the legibility fallback. Halving is a reference only. Maximum supply ceiling approved 2026-11-08 (100,000,000,000 OGO); the model family decision now constrains G and S_actual within that ceiling. |
| Period length D and nominal year | Issuance schedule, treasury projections, fee transition timing | A sustainability decision. Geometric ends at period 110 (9.2 years monthly, 27.5 years quarterly); hybrid ends at period 40 (3.3 years monthly, 10.0 years quarterly). |
| Validator minimum: Definition A (4 validators) or Definition B (at least 50, Nakamoto at least 10, top 10 under 33 percent, no country or provider over 33 percent) | Validator set, fee need, decentralization claims | Definition B targets are targets, not launch requirements. A mechanism for a flat stake distribution still needs to be named. |
| Fee policy: fee level, fee distribution, and how agent transactions are priced | Post-issuance funding, validator pay | Only the High fee scenario covers the minimum need, and only at period 38 (or 48 at standard costs). A flat fee cannot serve both security and agent viability. The fee market must not make small AI-agent payments impractical. |
| Fee structure: static, congestion-adjusted, or hybrid | Fee predictability, congestion response | No option selected. Must be tested under load before it is chosen. |
| Treasury spending level D and emergency reserve floor F | Treasury sustainability | No spending policy fixes a level problem. Maximum sustainable constant spend is 909 per period geometric, 2,500 hybrid (prior research). Under Low fees in the scenario model, actual R-F spending saturates at about 954 units per period averaged over 120 periods for both models, and the treasury still ends at the floor. |
| Treasury policy: R-F (recommended) or P4 (candidate) | Treasury spending rule | P4 remains a candidate, not an approved rule. P4 is the most conservative in the scenario model (spends least, ends highest under Low fees) but has a known post-issuance freeze defect. |
| Governance model: voting thresholds, deposit, voting period, quorum, expedited proposal path | Governance security | All TBD. The Hub's 72-hour expedited path is a precedent worth evaluating, not adopted. |
| Network name and chain id | Genesis, chain registry, wallet integration | The network name is OGO Network (locked). The chain id is not yet chosen. |
| Denomination and decimal precision | Genesis, wallet, explorer, chain registry | Proposed: `uogo`, 6 decimals. Proposals, not decisions. Cannot be changed after genesis without a hard fork. |
| Wallet address format: bech32 prefix and derivation path | Wallet, chain registry, external wallet integration | Proposed: prefix `ogo`, path `44'/118'/0'/0/0`. Both provisional. Must be fixed before the first wallet is released. |
| Commission floor, notice period, and maximum change per period | Validator economics, delegator protection | Proposed floor 5 percent (Hub precedent). Notice period and maximum change are TBD. |
| Unbonding period | Staking, slashing effectiveness | Proposed 21 days (Hub precedent). |
| Slashing parameters: downtime window, missed-block threshold, slash fractions | Validator reliability, delegator risk | Proposed: Hub precedent. Window and threshold are TBD. |
| Wallet technology stack and framework | M4 | Decided in M4, not in these specifications. Must support Keplr, Leap, and Cosmostation integration and the hardware signing seam. |
| Explorer and indexer technology | M6 | Decided in M6, not in these specifications. Must be usable without an OGO account. |
| Post-issuance funding mechanism and fee share | Treasury and validator funding after period N | Must be named before genesis, with revenue modelled, not assumed. Mechanisms: fee share, blockspace or coretime sales, endowment, contributions. Treasury staking is circular and is not a funding mechanism. |
| Validator participation funding: who funds it, in what currency, how much, for how long, what happens when it ends | Validator set growth, decentralization | No OGO/USD price is assumed. Infrastructure is budgeted in USD; stake incentives are stated in OGO units with the USD value labelled unknown. OGO Systems has no reserved allocation to fund participation from. |
| Liquid staking | Later phase | Not part of the first monetary prototype. Must be evaluated against the same supply and slashing rules before adoption. No liquid staking token is created by these specifications. |
| Bridge to a non-IBC chain | Interoperability | No bridge is adopted. Any bridge is third-party trust and must be evaluated under the same supply-accounting rules as IBC. |
| Epoch identifier name | Genesis | Working name `monetary`. Fixed in genesis. |
| SDK 0.55 upgrade timing | After Milestone C | Evaluated as a separate decision after Milestone C passes, with a documented effort estimate (M4). Not folded into Milestone C. |

## 5. Parameters awaiting approval

One parameter in this table is selected: maximum supply M (approved 2026-11-08). Each remaining parameter needs Benross's explicit approval, in the decision order in section 6.

Maximum supply M is APPROVED as 100,000,000,000 OGO (2026-11-08). Genesis circulating supply G and total scheduled issuance S_actual remain undecided; they must satisfy G + S_actual = 100,000,000,000 OGO. The remaining parameters awaiting approval are: genesis circulating supply G; issuance duration and period count N; reduction factor r; floor and base share (hybrid); initial per-period amount R0; taper factor and taper periods (hybrid candidate); validator, delegator, community, and treasury split; fee level (base fee and gas price) and fee distribution; treasury spending level D and emergency reserve floor F; minimum validator count and active set size N; governance deposit, voting period, quorum, and threshold; epoch duration D and nominal year; denom and decimals; bech32 prefix and derivation path; commission floor and change limits; unbonding period; slashing fractions and windows; chain id; self-delegation minimum.

## 6. Items requiring external legal or security review

| Item | Status | Notes |
|---|---|---|
| Legal analysis: securities, money transmission, tax, international distribution, exchange admission | Not yet commissioned | Must be commissioned in writing before genesis distribution. Longest-lead item. Nothing in these specifications is legal advice. |
| Security audits 1 to 5 (monetary core, consensus and governance, contracts and IBC, wallet and RPC, build and supply chain) | Not yet commissioned | Ordered in `OGO_NETWORK_SECURITY_MODEL.md` section 7. No unresolved critical or high findings before mainnet. |
| Economic review of the fee market | Not yet commissioned | The fee transition is the binding problem. An independent economic review of the fee scenarios, the agent pricing question, and the post-issuance funding gap is needed before genesis. |
| Fee-market evidence for agent transactions | Open | Agents are 45 percent of transactions but 7.7 percent of fee revenue in the High scenario. The fee policy must state explicitly how agent transactions are priced. |

## 7. Decision order

Preserved from the Decision Preparation report, with the architecture decision (made) removed from the sequence:

1. Legal analysis, commissioned in writing, before genesis distribution (longest-lead item).
2. Period architecture: epoch-based, recommended. Confirm the epochs module is available and maintained in the chosen SDK version (verified: x/epochs is a first-class SDK module from v0.53.0).
3. Period length D and nominal year.
4. Public cryptocurrency identity: the native coin is OGO Coin on OGO Network (decided). The remaining identity question is whether the same asset appears in any other form, and if so, under what accounting. One issuance authority or the cap breaks.
5. Minimum validator set and cost basis.
6. Treasury spend level and rule.
7. Post-issuance funding mechanism and fee share.
8. Fee policy evidence.
9. Model family (geometric or hybrid). Then parameters (reduction factor, floor, period count, initial amount). Then the split (validator, delegator, community, treasury). Then fee analysis. Then genesis circulating supply G and total scheduled issuance S_actual, chosen so that G + S_actual = 100,000,000,000 OGO, the approved hard maximum supply (2026-11-08).

## 8. Open questions from the repository inspection

| Question | Status | Notes |
|---|---|---|
| `src/OgoEconomy.tsx` exists on live main but no file imports it; the wallet UI appears unmounted | Open | Confirmed against live main 2026-11-06. Needs one live UI confirmation: does the wallet UI render anywhere, or is the component dead code? |
| Worker queries `forge_symbols` and calls `forge_search`, and the Builder Engine writes to `user_apps`, but no repository migration defines those tables | Open | Likely dashboard SQL. Any integration depending on Forge or `user_apps` must confirm the live schema in the dashboard before relying on it. |
| Local working copy contains `supabase/migrations/20261028_connect_phase0.sql`, not on live main | Open | The working copy is stale and is not a git repository. Any M6 implementation work must start from a fresh clone. |
| Naming collision: OGO Token vs OGO Coin vs OGO Coins | Open | The native coin is OGO, ticker OGO (locked). OGO Coins is the existing Supabase account-based balance. Unresolved as a naming and communication question, not a technical one. |
| Treasury low-inflow finding: 4 units or 0 units? | Resolved 2026-11-06 | 0 units. Recomputed from the stated model (see `docs/discrepancy-resolutions-2026-11-06.md`): both runs spend 114,511 over 120 periods. The 4-unit figure is withdrawn; it is not reproducible from the stated model, and the Decision Preparation report does not contain it (the attribution there was unverifiable). The substance is unchanged: the spending rule cannot fix a low-inflow level problem. |

## 9. What these specifications do not decide

These specifications select one monetary parameter: the hard maximum supply, 100,000,000,000 OGO, approved 2026-11-08. They do not select a monetary model and do not select any other monetary parameter; they do not approve a repository creation, do not approve a commit or a push, do not approve a publication or a deployment, do not approve a testnet or a mainnet, do not approve any spending, do not commission legal analysis or an audit, and do not authorise token issuance or distribution. Each of those needs a separate, explicit approval from Benross.
