# OGO Network: Monetary Architecture Decision Preparation v0.1

Status: Research and modeling only. No monetary model has been approved. Maximum supply approved 2026-11-08: 100,000,000,000 OGO. No percentages, no intervals selected. No token created, no contract deployed, no funds spent, no mainnet launch.
Date: 2026-11-06
Project: OGO Network

Purpose: prepare OGO for an informed monetary architecture decision, building on the Public Cryptocurrency Feasibility Report v0.1, the Monetary Design Decision Report v0.1, Hybrid Terminal Transition and Treasury Sustainability v0.1, and Treasury Sustainability v0.2. Five workstreams were run in parallel and their drafts verified against the prior reports before consolidation: period architecture, validator operating costs and decentralization, two-mode treasury specification, fee-market adoption model, and public cryptocurrency architecture comparison. This report is the consolidated record. The workstream drafts are staging material in /tmp/decision-prep/; this file is the deliverable.

## 0. How to read this report, and three corrections to prior work

Every figure in this report is illustrative, computed with exact rational arithmetic (Python Fractions, floored to integer uogo) for issuance, normalized to 1,000,000 illustrative units of total scheduled issuance unless stated otherwise. The normalization is not a proposed supply. The two models carried forward are geometric decay r = 9/10, R0 = 100,000, terminal period 110, and hybrid base 40% (10,000 per period), N = 40 periods, r = 9/10, R0 = 70,900, with a gentle geometric base taper (factor about 0.95 over the final 20 periods, renormalized) as the hybrid terminal variant. The illustrative split is 70/20/10 validator, community, treasury. Nothing here selects a parameter. A recommendation in this report is not a decision; the monetary decision stays with Benross.

Verification during consolidation found three corrections to Treasury Sustainability v0.2 and one convention clarification. They are recorded here, not smoothed over, because the decision record must be reproducible.

1. P2 full spend. v0.2 prose says the 5% balance cap "only reaches the desired 1,500 around period 20". Under the stated rule it reaches full spend at period 4 under geometric and period 6 under hybrid (1-indexed counting), because front-loaded inflow pushes the balance above 30,000 within a few periods. The genesis starvation in P2 is real but short: spend is 0 at period 0 and below desired for the first 4 to 6 periods, not the first 20. Verified by independent recomputation.
2. Hybrid "balance at issuance end". v0.2 reports the hybrid balance at issuance end for P1 (41,757), P3 (20,000), and P4 (48,076) as equal to the later post-drain balance. Under the stated rules the balances at period 39 are P1 44,532, P3 45,982, P4 49,351, and they then drain for several periods after issuance ends (P1 to 41,755, P3 to 20,000, P4 to 46,574). The v0.2 totals and shortfall counts are unaffected. This report uses the balance at the issuance end period itself.
3. P4 trailing window. v0.2 reproduces its P4 figures only if the trailing window includes the current period, while its P1 figures reproduce only if the window excludes the current period. The two policies therefore used inconsistent windows. Under one consistent convention (prior periods only, the implementable on-chain rule, since caps must read settled state), P4 is 58,827 at issuance end under geometric and 49,351 under hybrid, with totals 41,120 and 53,408. This report uses the consistent convention throughout.
4. Convention clarification, not an error: per-period costs. v0.2 defined validator cost, treasury spend, and delegator yield per period and rescaled only delegator yield across the monthly and quarterly mappings, giving a quarterly minimum need of 20,875. The fee-market workstream defines all three per month (a validator server bill is monthly, not per period) and converts to per period, giving a quarterly minimum need of 43,875, three times the monthly 14,625. Both conventions are internally consistent; the calendar-consistent one is adopted going forward, because a quarterly period contains three months of validator cost. The crossover period for the minimum need is identical in periods under both mappings, but three times longer in calendar years under quarterly periods.

A second convention note: total issuance-funded treasury inflow is 99,947 illustrative units under geometric and 99,982 under hybrid when computed as the sum of floored per-period 10% shares. v0.2 reported 99,993 and 99,998, which are 10% of the analytic totals. The simulation-grounded figures are used in this report.

Correction note, 2026-11-06 (GitHub publication): section 6.2 stated hybrid dust as 18 uogo per 1,000,000. Corrected to 22 uogo. The 18 figure came from solving B0 as an exact rational from S_target = 1,000,000; the implemented definition uses integer B0 = 60,900 (base 40%), giving S_actual 999,978 and dust 22. Recomputed 2026-11-06 with exact rational arithmetic; see `docs/discrepancy-resolutions-2026-11-06.md`.

## 1. Period architecture research

### 1.1 The three definitions

(a) Block-based periods: a period is L blocks. Period n covers heights [start + n*L, start + (n+1)*L - 1]. This is the definition the deep-dive assumed and the Milestone B scaffold implements. (b) Time-based periods: a period is a fixed wall-clock duration, enforced via block timestamps. (c) Epoch-based periods: a period is one epoch of a named epochs-module timer (Cosmos SDK x/epochs style), with a fixed duration D in seconds, fixed in genesis, and issuance claimed once in the epoch hook.

### 1.2 What real chains do

Cosmos SDK x/epochs (Osmosis) defines epochs as timers that tick when block time passes the epoch end; Osmosis x/mint mints once per epoch, weekly in the classic design and daily from v20, with a 156-epoch reduction period. The current Cosmos SDK x/mint supports epoch-based minting through MintFn as the primary path. Cosmos Hub mints per block, with inflation between 7% and 10% and a 67% bonded goal. Cardano uses 5-day epochs (432,000 one-second slots), with rho 0.3% of the remaining reserve per epoch. Solana uses 432,000-slot epochs (about 2 to 3 days). Ethereum uses 32 slots of 12 seconds (6.4 minutes). No major chain uses a variable-length calendar month as an issuance period.

### 1.3 Comparison

| Dimension | (a) Block-based | (b) Time-based, calendar | (c) Epoch-based, x/epochs |
|---|---|---|---|
| Issuance predictability in count | High. Period n is known from height alone. | Medium. Period n is known only once a timestamp crosses a boundary. | High. Epoch number is on-chain state. |
| Issuance predictability in calendar date | Low. Date depends on future average block time. | High in intent, Medium in execution. The mint block is the first block after the boundary, and the timestamp is validator-influenced. | High nominal, Medium actual. Nominal date is genesis start plus n times D. Actual mint is the first block after that date. |
| Validator rewards, amount certainty | High. Amount is R_n, fixed by n. No timestamp enters the amount. | High if amount is a function of n only. Low if any pro-rata rule is added. | High if amount is a function of epoch number only. The epochs module does not pro-rate. |
| Treasury funding, timing | Inflow at period boundaries; calendar timing drifts with block time. | Inflow on calendar boundaries, matching calendar budgeting. | Inflow on epoch boundaries, nominal calendar boundaries, matching budgeting to within the tick delay. |
| Fee revenue, accounting | Fees accrue per block; period definition decides attribution and settlement frequency. | Fees attributed to calendar periods, matching external reporting. | Fees attributed to epochs, settled at epoch ticks. This is the Osmosis pattern. |
| Participation, staking alignment | Low. Block boundaries have no relation to staking or unbonding periods, which are time-based in Cosmos (for example, 21-day unbonding on the Hub). | High. Calendar periods align with how participants think about time. | High. Epochs are the Cosmos-native unit for staking-adjacent scheduling. |
| Calendar-based monetary commitments | Low. "Issuance ends in year Y" is an estimate, not a protocol guarantee. | High in statement, fragile in enforcement. Enforcement depends on block timestamps, which are not wall-clock time. | High, with a stated tolerance. The promise is exact as a nominal date, and the actual mint is the first block after that date. |
| Timestamp manipulation risk | None. Timestamps are not consulted. | High if naive. A proposer or coalition can shift the trigger block within BFT bounds. | Medium. Same dependence, but mitigated: amount is a function of epoch number only, the anchor resets to the prior end time, and a long duration dilutes a small shift. |
| Halt and stall behavior | Clean pause. No blocks means no new periods, no issuance accrues, end date slips by the halt duration. | Ambiguous without a catch-up rule. Time passes during a halt, so periods accrue in wall-clock terms. Must decide: lump, sequential catch-up, or skip. | Defined by the module, but must be decided for minting. Missed epochs tick one per block on restart, in order, preserving S_actual and per-period cap order. |
| Implementation risk, off-by-one | Medium. One boundary convention must be chosen and tested. The scaffold already implements this. | High. Calendar boundaries, inclusivity, time zones, leap years, and missing days all need rules. | Medium. Two hooks exist (AfterEpochEnd and BeforeEpochStart). Subscribing to the wrong one, or both, double-mints or misattributes. Identifier filtering is required. |
| Determinism and independent verification | High. Any outsider with the genesis file and a height can recompute n and R_n. | Medium. Verification needs block timestamps and an exact timestamp rule, including boundary inclusivity and calendar rules. | High. Verification needs the epoch number from state and the genesis duration. The tick rule is in the module, not in OGO-specific code. |

### 1.4 Block-time drift, quantified

CometBFT does not produce blocks on a fixed clock. Actual block time varies with proposer liveness, round changes, network latency, block size, and validator performance. For a block-based period, this variability accumulates: if average block time is 10% slower than the target assumed when L was chosen, every period is 10% longer in calendar time, and the issuance end date slips by 10% of the whole schedule. Using the v0.2 nominal mappings as the base, the illustrative slip is:

| Average block time vs target | Geometric end, monthly nominal 9.2y | Geometric end, quarterly nominal 27.5y | Hybrid end, monthly nominal 3.3y | Hybrid end, quarterly nominal 10.0y |
|---|---|---|---|---|
| 10% slower | 10.1y, slip about 11 months | 30.3y, slip about 2.8y | 3.6y, slip about 4 months | 11.0y, slip about 1.0y |
| 20% slower | 11.0y, slip about 1.8y | 33.0y, slip about 5.5y | 4.0y, slip about 8 months | 12.0y, slip about 2.0y |
| 10% faster | 8.3y, slip about 11 months early | 24.8y, slip about 2.8y early | 3.0y, slip about 4 months early | 9.0y, slip about 1.0y early |

Solana is the production evidence for this scale: its epoch is a fixed slot count, its inflation schedule assumed 180 epochs per year at 2 days per epoch, and by June 2026 it was 276 days behind its intended timeline because slot times ran long. OGO block-based periods would have the same property, with height in place of slots. For epoch-based periods, block-time variability does not accumulate: a slow block delays a single tick, but the next epoch start is anchored to the prior end time, so the following tick is due sooner, not later by the full delay. The nominal schedule stays anchored, and the delay is bounded in normal operation by the block interval, not by the whole schedule length.

### 1.5 Timestamp risk

Block timestamps in CometBFT are not wall-clock time read from a trusted clock. Under the default BFT time rule, the timestamp of a block is the weighted median, by voting power, of the timestamps validators sent in their precommit messages for the previous block. If more than one half of voting power is faulty, time is under total control of the faulty validators, and the proposer has a degree of freedom in choosing the median. A newer proposer-based time design is discussed in the CometBFT specification as a proposal, not as the default in the versions OGO is building on. The February 2026 Tachyon vulnerability (GHSA-c32p-wcqj-j677, fixed in CometBFT v0.38.21) shows this is not theoretical: a bug in median-time computation allowed a malicious proposer to skew the weighted median to a value of its choosing. The risk is timing, not double minting, if the issued-by-period set is correct: an early mint pays validators, community, and treasury early, and a delayed mint pays them late. For a long period (a month or a quarter), a shift of minutes or hours is negligible as a share of the period. For a short period (a day or less), the same shift is a material share.

### 1.6 Halt behavior

Block-based pauses cleanly with no accrual, but slips every calendar date by the halt duration. Pure time-based accrues owed periods during a halt and needs an explicit catch-up rule: lump minting, sequential catch-up, or skip. Skipping would break S_actual and M and is not acceptable. Epoch-based queues missed epochs and clears one per block on restart, in order, preserving S_actual and per-period cap order, with no lump mint and no lost periods.

### 1.7 Recommendation

Recommendation, not a decision. Adopt epoch-based periods as the period architecture for OGO Network issuance. A period is one epoch of a named epochs-module timer, with a fixed duration D in seconds, fixed in genesis, and issuance for epoch n is R_n, claimed once, in the epoch hook, with the issued-by-period set keyed by epoch number and the cap check applied per epoch, in epoch order. Reasons, in order of weight: it makes the v0.2 calendar mappings well-defined as nominal dates (genesis start plus n times D, with a stated tolerance of at most one block interval in normal operation plus halt duration), instead of estimates that drift with block time; it is the Cosmos-native path, using maintained, tested, documented code for the timer, anchoring, ticking, and hooks instead of OGO-specific code; it handles halts correctly by construction; it aligns with participation, since epochs are the unit Cosmos participants already use for staking-adjacent scheduling; and timestamp risk is bounded and diluted by a long D and an amount that is a function of epoch number only.

This recommendation does not recommend a period length. Under it, a period called monthly is an epoch whose D is chosen so that 12 epochs span one nominal year, and a period called quarterly is an epoch whose D is chosen so that 4 epochs span one nominal year. A nominal year is defined in genesis as a fixed number of seconds (for example, a Julian year of 365.25 days, 31,557,600 seconds, which gives a monthly D of 2,629,800 seconds and a quarterly D of 7,889,400 seconds, illustrative arithmetic only, not a recommended length). Growth rates per period then convert to annual rates exactly for nominal years, because the number of periods per nominal year is fixed. The length decision remains separate.

Period definition, length, start, and identifier must be fixed in genesis, changeable only via the constitutional process, not ordinary governance. Changing L remaps heights to periods and breaks the issued set; changing D preserves epoch count but breaks calendar promises; switching architectures mid-schedule is effectively a new schedule and needs S_actual and M recomputation plus issued-set migration.

Evidence that would change the recommendation: if the epochs module is unavailable or unmaintained in the chosen SDK version; if a security review finds an unmitigable hook vulnerability; if Benross decides calendar promises are not wanted at all (then block-based suffices); if D is chosen short (a day or less), so timestamp shift becomes material; or if the epochs module changes its tick semantics.

## 2. Validator operating-cost analysis

Treasury Sustainability v0.2 modeled validator economics with illustrative placeholder costs only (50, 100, or 200 units per validator per period) and explicitly stated no validator cost study existed yet. This section replaces those placeholders with researched real-world evidence. Research dated 6 November 2026. Prices move; Hetzner prices moved sharply in 2026, so costs need rechecking before any budget.

### 2.1 Hardware requirements

Cosmos Hub official specification, checked 15 October 2026: 4+ CPU cores, 32 GB RAM (64 GB recommended by operators for upgrade migrations), 500 GB NVMe storage for a pruned validator, plus 3 to 5 or more sentry nodes. Comparable chains are similar, with requirements growing with chain state and throughput. Storage grows with history; a validator that keeps more history or serves snapshots needs substantially more disk. Bandwidth is dominated by consensus traffic and sentry relaying, not by transaction volume at OGO's expected scale, but it rises with validator count and block size.

### 2.2 Cost tiers, researched

| Tier | What is included | USD per validator per month |
|---|---|---|
| Lean solo, infrastructure only | One Hetzner AX41 class validator plus 2 small sentry VPS, file-based signing, self-built monitoring | 75 |
| Standard, infrastructure only | One Hetzner AX52 class validator at 64 GB plus 3 sentries across 2 providers plus monitoring and backup storage | 200 |
| Professional, infrastructure only | Standard plus standby validator host, 2 to 3 signer hosts, 5 or more multi-provider sentries, HSM or YubiHSM, snapshots | 500 |
| Professional plus part-time ops | Professional infrastructure plus 0.1 full-time equivalent of operations labour at 130,000 USD per year (about 1,080 USD per month), rounded | 1,500 |
| Enterprise, infrastructure plus 24/7 ops share | Professional infrastructure plus 0.3 to 0.4 full-time equivalent at 150,000 USD per year (about 3,750 to 5,000 USD per month), rounded down | 4,000 |

Cloud is 2x to 4x dedicated-server pricing before egress (AWS m6i.xlarge about 140 USD per month plus about 40 USD per month per 500 GB EBS plus 0.09 USD per GB egress). Node operations salaries in 2026 run 110,000 to 160,000 USD per year. A single-chain operator staffing 24/7 for OGO alone would pay about 45,000 to 55,000 USD per month in labour, which is not modeled as a tier because no fee model in this report can fund it; the labour tiers assume the operator spreads staff across about 10 chains (part-time) or about 3 to 4 chains (enterprise share).

### 2.3 What the v0.2 placeholder meant

100 OGO per validator per period equals 100 USD per month only at an illustrative price of 1 USD per OGO, which sits between lean and standard infrastructure-only cost. At 0.50 USD it equals 50 USD per month (below lean); at 0.10 USD it equals 10 USD per month (buys nothing). With any labour counted it understates cost at every illustrative price below about 15 USD. The placeholder was not wrong as a sensitivity input; it was ungrounded. The table below grounds it.

### 2.4 Recomputed minimum fee need

Minimum need, monthly periods, at 100 validators, 50% stake at 5% yield (3,125 per period), treasury spend 1,500 per period. The fee need varies 31x across the 0.10 to 10 USD illustrative price range. The model family does not change the absolute need, only the date it must be met (period 40 hybrid vs period 110 geometric). Quarterly periods raise the per-period need by about 43% under the v0.2 per-period convention, and by 3x under the calendar-consistent convention in section 0 (43,875).

| Validator tier | USD per validator per month | Total fee need per period at 1 USD | At 0.50 USD | At 0.10 USD |
|---|---|---|---|---|
| Lean, infrastructure only, 100 validators | 75 | 12,125 | 19,625 | 79,625 |
| Standard, infrastructure only, 100 validators | 200 | 24,625 | 44,625 | 204,625 |
| Professional, infrastructure only, 100 validators | 500 | 54,625 | 104,625 | 504,625 |
| Professional plus part-time ops, 100 validators | 1,500 | 154,625 | 304,625 | 1,504,625 |
| Enterprise, 100 validators | 4,000 | 404,625 | 804,625 | 4,004,625 |

For scale: at a Cardano-like fee level (469 per period, 80% to validators), the network sustains 5 lean or 1 standard validator at 1 USD, 0 at any tier at 0.10 USD, and 25 lean, 9 standard, or 3 professional at 5 USD. v0.2's finding that Cardano-like fees sustain about 3 validators stands, and sharpens: it is 3 validators gross, 0 net of delegator yield.

### 2.5 Redundancy, staffing, and security

Sentry node architecture is the standard: the validator never exposes its public address; sentries relay consensus traffic, and the validator connects only to its own sentries. Backup validators and failover, monitoring with alerting, and key management complete the baseline. Horcrux, the widely used remote signer, is archived, and a Horcrux v3.3.1 race condition tombstoned an Osmosis validator on 6 March 2025; Cosmos KMS v0.1.0 (July 2026) is now the maintained remote signer. 24/7 operations requires at least two people for coverage and upgrade response; a one-person operation can run a validator but cannot guarantee response to a 3 AM upgrade or incident. Slashing on the Hub is 0.01% plus 10 minutes jail for downtime (signing less than 5% of the last 10,000 blocks) versus 5% plus permanent tombstoning for double-signing, and the consensus key cannot be rotated, so a leak means retiring the validator. Commission floors are 5% on the Hub, dYdX, and Worrell; Hub staking APR is about 15.5% at the 10% inflation ceiling. Delegators choose on uptime, commission, reputation, and self-bond; what yield they compare against is the opportunity cost of staking elsewhere, which is why the v0.2 delegator yield assumption (3% to 5% per year) needs its own evidence before it is relied on.

### 2.6 Geographic distribution

Validators concentrate in a small number of providers and jurisdictions. The current Cosmos default (Hetzner Germany and Finland plus AWS Frankfurt) would fail any serious geographic diversity threshold. Jurisdictional concentration is a correlated-failure risk: a single country action, provider outage, or cloud region failure can remove a large share of stake weight at once. Geographic spread must be measured and reported, not assumed.

## 3. Minimum decentralization and security analysis

A set can be technically operational and still be effectively centralized. The two definitions must not be conflated.

### 3.1 Definition A: minimum technically operational validator set

The smallest set that keeps the chain live and safe, under CometBFT fault tolerance math, assuming honest and independent operators with adequate stake each.

- Count: at least 4 validators (3f plus 1 with f = 1). Below 4, the chain has zero Byzantine tolerance by count. At 4, it tolerates 1 faulty validator by count.
- Stake distribution: no single validator holds one third or more of voting power (otherwise that validator alone can halt the chain), and no two validators together hold two thirds or more.
- Liveness: more than two thirds of voting power online and signing. At n = 4, at least 3 of 4 signing. The loss of 2 validators halts the chain.
- Operations: each validator meets the standard hardware and sentry baseline, runs a remote signer or equivalent key protection, and can respond to upgrades. A set of 4 lean solo operators can meet this, at about 75 USD per validator per month in infrastructure, or 300 USD per month for the whole set.
- Cost per period, monthly periods, infrastructure only: 4 validators at 75 USD is 300 USD per month; at 200 USD is 800 USD per month; at 500 USD is 2,000 USD per month. In OGO units at an illustrative price of 1 USD, that is 300, 800, and 2,000 OGO per period respectively, before delegator yield and treasury spend.
- What it does not provide: no meaningful decentralization, no resistance to coordination among a small group, no geographic or provider diversity unless explicitly arranged, no cover for the loss of more than 1 validator, and no credibility with delegators or exchanges that the chain is not controlled by its founders or its largest operators. A 4-validator set is a testnet that happens to be on mainnet, or a permissioned consortium, not a public decentralized network. Small-set practice shows chains do operate at 17 to 38 validators, which is above the 4-validator floor but still concentrated. The floor is a liveness floor, not a target.

### 3.2 Definition B: sufficiently decentralized validator set

The set size and stake distribution needed for credible decentralization, as a research proposal for Benross to accept, tighten, or replace. No live Cosmos chain in the evidence meets all four thresholds at once. That is itself a finding: the bar is set above current practice, deliberately, so the gap is visible.

| Threshold | Proposed value | Source and reasoning |
|---|---|---|
| Nakamoto coefficient (one-third, liveness) | At least 10 | Polli treats below 5 as high risk. Osmosis at 12 was presented as an improvement over 10. Solana at 18 to 20 is the healthiest in the evidence. 10 is the bottom of the credible band, not the top. |
| Top-10 stake share | Below 33% | One third is the BFT halt threshold. If the top 10 together exceed one third, the 10 largest validators can halt the chain. Solana at 24.5% is the only chain in the evidence below this line. |
| Geographic, jurisdictional, and provider concentration | No single country, jurisdiction, or cloud provider holds more than 33% of stake weight | Same one-third BFT logic applied to correlated failure: a single country outage, jurisdiction action, or provider outage affecting more than one third of stake weight can halt the chain. This is a proposed extension of the BFT threshold to geography and infrastructure, not a measured rule from any chain. |
| Active validator count | At least 50 | Below about 30, the evidence shows Nakamoto coefficients of 1 to 5. At 50 to 70, coefficients of 3 to 5 are still common, so 50 is a floor, not a guarantee. The count matters only if the distribution also meets the first two thresholds. |

Requirements to meet Definition B: at least 50 active validators, each meeting the standard hardware and sentry baseline, with at least 10 of them holding enough stake each that the top 10 together stay below one third. This implies a deliberately flat distribution, which does not emerge from stake-rank entry alone. It needs a mechanism: delegation incentives, a saturation cap, a per-validator subsidy, or genesis selection. Validators across at least 10 countries and at least 5 providers, with no country, jurisdiction, or provider above 33%. This must be measured and reported, not assumed.

Cost per period, monthly periods, infrastructure only: 50 validators at 200 USD is 10,000 USD per month; 100 validators at 200 USD is 20,000 USD per month; 100 validators at 500 USD is 50,000 USD per month. In OGO units at an illustrative price of 1 USD, that is 10,000, 20,000, and 50,000 OGO per period respectively, before delegator yield and treasury spend.

What Definition B provides that Definition A does not: halting requires coordinating at least 10 independent validators, the top 10 cannot halt the chain alone, no single country or provider outage halts the chain, and the loss of any 1 or 2 validators does not move the Nakamoto coefficient materially. What it still does not provide: protection against a majority of stake being controlled by a smaller number of entities behind multiple validators (Sybil validators), protection against exchange or custodian concentration (Coinbase alone holds 20.56% on the Hub), or protection against correlated software failure (all validators running the same client version).

### 3.3 The gap between the definitions, in cost terms

At an illustrative price of 1 USD, monthly periods, infrastructure only, including delegator yield (3,125) and treasury spend (1,500) in the fee need:

| Set | Validators | Monthly USD | Monthly OGO at 1 USD | Fee need including yield and treasury |
|---|---|---|---|---|
| Minimum operational, lean | 4 | 300 | 300 | 4,925 |
| Minimum operational, standard | 4 | 800 | 800 | 5,425 |
| Sufficiently decentralized, standard, 50 validators | 50 | 10,000 | 10,000 | 14,625 |
| Sufficiently decentralized, standard, 100 validators | 100 | 20,000 | 20,000 | 24,625 |
| Sufficiently decentralized, professional, 100 validators | 100 | 50,000 | 50,000 | 54,625 |

Note: the 50-validator standard row at 1 USD reproduces the v0.2 baseline total of 14,625 exactly, because 50 validators at 200 OGO plus 100 validators at 100 OGO both equal 10,000 OGO of validator budget. The coincidence is arithmetic, not a finding. The substantive point is that credible decentralization costs about 2x to 5x the v0.2 validator budget in infrastructure alone at the same illustrative price, and 10x to 50x once labour is counted.

### 3.4 Stake concentration, point-in-time snapshots, research dated 6 November 2026

Cosmos Hub: Nakamoto coefficient 3 (down from 6 in a year), top 10 about 56%, largest validator Coinbase 20.56%, Gini 0.73 rising to 0.80. Osmosis: Nakamoto coefficient 4 at 30 validators, top 10 60.27% (it was 12 at 120 validators); the set was being cut 150 to 120 to 100 to 70 to proposed 30 during 2025 to 2026, so its figures are date-sensitive. Solana: Nakamoto coefficient 18 to 20, top 10 24.5%, validator count down about 68% from 2,560 (March 2023) to 659 to 795, driven by fixed about 1 to 1.1 SOL per day vote costs squeezing small validators (break-even about 50,000 to 75,000 SOL). Cardano: k = 500, saturation about 68M ADA, minimum pool cost 340 ADA per epoch with the 170 ADA halving unresolved (tipping point 10M ADA pledge). Validator count alone does not predict decentralization: Osmosis cut its set while its Nakamoto coefficient fell, and Solana lost two thirds of its validators while keeping the healthiest coefficient in the evidence, because the validators that left were small and the stake distribution stayed flat.

Caveat: Cosmos Hub active set is reported as both 180 (docs, October 2026) and capped at 200 (audit, July 2026), unresolved.

## 4. Two-mode treasury specification

### 4.1 Definitions and simulation convention

Let B_t be the treasury balance at the end of period t, with B_{-1} = 0. Let I_t be total treasury inflow in period t. Spend in period t is computed from settled state only: the balance at the start of period t (B_{t-1}, before period t inflow is added) and inflows of prior periods only. Then B_t = B_{t-1} - spend_t + I_t. A trailing average over w periods means the average of I_{t-w} through I_{t-1} (prior periods only; fewer than w at genesis; empty means 0). Inflow during issuance is floor(R_t / 10) under the illustrative 70/20/10 split, so total issuance-funded treasury inflow is 99,947 illustrative units under geometric and 99,982 under hybrid. Issuance uses exact rational arithmetic, floored once to integer uogo. Treasury arithmetic is integer throughout. Desired spend D = 1,500 per period, horizon 360 periods, reserve floor F = 20,000, balance cap rate 5% per period, trailing window 4 periods, unless a rule states otherwise. This settled-state convention is also the implementable on-chain rule: caps read state that is already final at the start of the period, never inflow that arrives in the same period.

### 4.2 Mode A: during issuance

Rules, stated as formulas. P0 fixed: spend_t = D. P1 trailing inflow cap: spend_t = min(D, Trail_4(t)). P2 balance cap: spend_t = min(D, floor(0.05 * B_{t-1})). P3 reserve floor: spend_t = min(D, max(0, B_{t-1} - F)). P4 combined AND: spend_t = min(D, Trail_4(t), floor(0.05 * B_{t-1}), max(0, B_{t-1} - F)). P5 trailing 12 inflow cap: spend_t = min(D, Trail_12(t)). P6 Cardano Net Change Limit style ceiling: spend_t = min(D, floor(0.05 * B_{t-1}), max(0, 40,000 - sum of spend over the prior 39 periods)), a cumulative ceiling (1,000 per period on average), illustrative. P7 Polkadot Dynamic Allocation Pool style inflow-dependent budget: spend_t = min(D, floor(1.5 * Trail_4(t)) + floor(0.02 * B_{t-1})). P8 spend-only-yield with a principal floor: spend_t = min(D, floor(0.005 * B_{t-1})). P9 time-locked tranches: inflow in period t is locked for 12 periods and unlocks at t + 12; spend draws only from unlocked tranches. P10 dual guardrail, OR: spend_t = min(D, max(Trail_4(t), floor(0.05 * B_{t-1}))). P11 OR with trailing 12 and a soft floor: spend_t = min(D, max(Trail_12(t), floor(0.05 * B_{t-1}))), except when B_{t-1} <= F, when spend_t = min(D, max(Trail_12(t), floor(0.05 * B_{t-1})), floor(0.02 * B_{t-1})).

Recommended Mode A rule, R-F (OR cap plus hard floor): spend_t = min(D, max(Trail_4(t), floor(0.05 * B_{t-1})), max(0, B_{t-1} - F)). This is P10 with one added constraint: spend may never take the settled balance below F. The floor is a ceiling on spend, not a separate policy.

Mode A comparison, same format as v0.2 section 2, consistent convention throughout:

| Policy | Geometric: balance at issuance end | Geometric: balance at t=360 | Geometric: total spent | Hybrid: balance at issuance end | Hybrid: balance at t=360 | Hybrid: total spent |
|---|---|---|---|---|---|---|
| P0 fixed | -65,053 | -440,053 | 540,000 | 39,982 | -440,018 | 540,000 |
| P1 trailing 4 | 55,682 | 55,682 | 44,265 | 44,532 | 41,755 | 58,227 |
| P2 balance 5% | 1,328 | 19 | 99,928 | 44,552 | 19 | 99,963 |
| P3 floor 20,000 | 20,000 | 20,000 | 79,947 | 45,982 | 20,000 | 79,984 |
| P4 combined AND | 58,827 | 58,827 | 41,120 | 49,351 | 46,574 | 53,408 |
| P5 trailing 12 | 48,846 | 48,846 | 51,101 | 42,975 | 35,476 | 64,506 |
| P6 NCL 40,000/40 + balance 5% | 4,645 | 19 | 99,928 | 59,982 | 19 | 99,963 |
| P7 DAP 150% trail + 2% balance | 9,602 | 88 | 99,859 | 41,482 | 88 | 99,894 |
| P8 yield 0.5% | 60,648 | 17,391 | 82,556 | 87,849 | 17,742 | 82,240 |
| P9 tranche 12 | 0 | 0 | 99,947 | 57,982 | 0 | 99,982 |
| P10 OR trail 4 / balance 5% | 1,253 | 19 | 99,928 | 41,482 | 19 | 99,963 |
| P11 OR trail 12 / balance 5%, soft floor | 6,771 | 71 | 99,876 | 45,192 | 71 | 99,911 |
| R-F (recommended) | 20,000 | 20,000 | 79,947 | 45,982 | 20,000 | 79,982 |

Reading: trailing-12 only delays the freeze. NCL creates a hard-stop-then-resume cliff. DAP spends nearly everything (ends at 88). Yield-only (0.5%) never once meets desired spend and preserves a balance by never funding the need. Tranches delay but end at 0. OR without a floor (P10) spends the treasury down to 19. R-F is the only rule tested that reaches full desired spend within 4 periods of genesis (t=3 geometric, t=4 hybrid), spends exactly the P3 total (79,947 geometric, 79,982 hybrid), and preserves exactly the 20,000 floor at t=360 in both models. Under hybrid it spends the full 1,500 in the issuance end period and then drains to the floor over the following periods; under geometric the balance reaches the floor before issuance ends and spend stops there. Its shortfall count is 313 in both models, against 307 for P3: R-F falls short 6 periods more often than P3, the price of spending at the OR cap early while still preserving the floor. A percentage rule in integer uogo floors to zero once the balance falls below 1 divided by the rate (below 20 units for the 5% rule), so P2, P6, and P10 show spend 0 at t=360 with balance 19; the residual dust balance is not spendable under the rule, and any specification relying on a balance-percentage rule must state this explicitly.

### 4.3 Mode B: after issuance

The freeze is cured by redefining the inflow stream, not by a new formula. The trailing average in R-F reads the total disclosed inflow stream (issuance plus fee share plus coretime plus other, each defaulting to 0), so after issuance ends the same formula keeps working on fee inflow. Under flat Cardano-like fees (93 per period to the treasury under geometric, 66 under hybrid), P1, P4, and R-B all spend at exactly the fee inflow rate forever, with the balance pinned just above the floor (20,093 geometric, 20,066 hybrid). Under growth fees the treasury becomes fee-funded (1.50M geometric, 1.09M hybrid fee inflow versus about 100k issuance inflow over 360 periods) and desired spend is met in all but 3 to 4 periods. The Mode B rule is therefore the same formula with a new stream, not a second rule: spend_t = min(D, max(Trail_4(t), floor(0.05 * B_{t-1})), max(0, B_{t-1} - F)), where Trail_4(t) averages total disclosed inflow over the prior 4 periods. What changes at the transition is the composition of I_t, not the rule.

### 4.4 Transition rule

The transition is keyed to the genesis-fixed final issuance period (geometric t = 109, hybrid t = 39), with no vote and no threshold trigger. A threshold trigger was considered and rejected: dust-level inflow near the geometric tail makes any threshold ambiguous. Mode A limits cannot reopen because issuance inflow is zero by the schedule itself and the rule reads on-chain accumulators, not labels. The floor survives as the emergency reserve sub-account; the operating balance carries over unchanged. No party can reclassify spending or reopen Mode A limits after the transition, because there is no reclassification step: the rule is one formula over one inflow stream, and the stream's composition changes by the schedule, not by any decision.

### 4.5 Validator subsidy during the transition valley

Yes, but it cannot bridge the valley. The hybrid valley is reproduced at 116,089 total shortfall over periods 40 to 65 (v0.2: 116,065), worst period 7,185. A segregated valley reserve capped at 25,000 covers only 25,000 (uncovered 91,089); even spending the entire transition balance covers at most 53,544 (uncovered 62,545). The remainder must close via fee growth, validator exit, or a lower minimum budget. Geometric has no valley. The subsidy rule, stated as formulas with a segregated valley reserve, recipient list, and cap, is in the workstream draft. The honest conclusion is that a treasury subsidy is a bridge for part of the valley, not a solution to it.

### 4.6 Emergency reserve policy

Triggers are on-chain state only: validator count below 50 for 3 periods, chain halt by block gap, fee collapse. A declared "security incident" alone is explicitly not a trigger. Caps: 1,000 per period, 6,000 per incident, 6 periods maximum, exhaustive recipient list, every draw a public on-chain event. Replenishment is automatic: regular spend is suspended until the floor rebuilds (about 91 periods under flat hybrid fees, 65 under flat geometric). Governance: the rule structure is immutable in genesis; parameters changeable only within genesis-set bounds; the floor ratchets upward only; desired spend raisable only by constitutional threshold and never above trailing-12 inflow; all changes timelocked 12 periods and never retroactive.

## 5. Fee-market adoption model

### 5.1 Evidence anchors: what real networks actually collect

Research dated 6 November 2026. Figures are point-in-time snapshots, not live chain reads. The lesson is stark: high transaction counts do not imply high fee revenue, and Cosmos appchains collect very little.

| Network | Daily transactions | Typical fee per transaction | Daily chain fee revenue | What it shows |
|---|---|---|---|---|
| Cosmos Hub | About 4,959 daily active users (altFINS, 30 Aug 2026) | About 0.0014 USD normal, 0.0029 USD fast; floor 0.005 uatom per gas, transfer about 80,000 to 100,000 gas | 38.49 USD in 24h chain fees, 0 USD chain revenue (DefiLlama, 6 Nov 2026); about 50.67 USD on 18 Sep 2026 | A chain with a market capitalization near 800M USD collects tens of dollars a day in chain fees. Fee revenue is effectively zero relative to issuance. |
| Osmosis | No reliable daily count found | About 0.002 to 0.004 USD per transaction on average; 0.1% taker fee on trades | 372 USD in 24h chain fees; DEX fees 3,333 USD in 24h, app fees 18,562 USD in 24h, DEX volume 464,675 USD in 24h; 30-day DEX fees 215,881 USD | Application fees (taker fees) exceed chain gas fees by an order of magnitude. TVL fell from 1.832B USD (Mar 2022) to about 10M USD (Nov 2026). |
| Solana | 112.6M average daily non-vote transactions in Q1 2026 (Messari); 167M per day in Aug 2026 including failed | 0.00046 USD median non-vote fee year to date to Aug 2026; base fee 5,000 lamports per signature, 50% burned, 50% to validator; priority fees 100% to validator | Network fees only about 6% of the 14.4M USD per day users paid to use Solana and its applications in the 30 days to 24 Oct 2026 (about 0.86M USD per day network fees); issuance about 62,415 SOL per day | Even at 100M plus transactions a day, sub-cent fees produce network fee revenue far below issuance. Application revenue is the larger pool, and it does not accrue to validators by default. |
| Base | 12.89M daily transactions in May 2026; 8.95M daily average in Aug 2026 | About 0.01 to 0.02 USD per transaction post-Dencun; median USDC transfer about 0.02 USD | Average daily priority fees about 156,138 USD, about 86.1% of daily sequencer revenue (30-day snapshot to 7 Sep 2026) | The highest-volume L2 still runs on one to two cent fees and a few hundred thousand daily active addresses. Most activity is low-value or automated micro-transactions. |

Osmosis is running the live experiment OGO would need to survive. Governance Proposal 1053 ended OSMO inflation on 26 October 2026 (daily minting from about 91,324 OSMO to zero), and Proposal 1052 raised the stakers share of non-OSMO taker fees from 22.5% to 47.5%. Staking rewards now depend entirely on protocol revenue. No outcome data exists yet in this report. This is the closest live test of a fee-funded transition, and it should be tracked as evidence before any OGO fee assumption is treated as plausible.

Cardano fees covered about 0.67% of staking rewards over the year to 1 September 2026. Ethereum EIP-1559 burn offset about 2.4% of issuance in 2026. A Cardano-like fee level would produce only 469 illustrative units per period under geometric (0.67% of period-0 validator pay of 70,000), covering 3.2% of the 14,625 minimum need.

Implication for OGO: a new Cosmos SDK chain should expect Cosmos-scale fee revenue at launch, not Solana-scale or Base-scale revenue. The scenarios below start there.

### 5.2 Method and conventions

Fee revenue is anchored per day, then converted to per period: monthly period = 30 days, quarterly period = 90 days. Fee revenue per period = daily fee revenue times days per period. Validator operating cost, treasury spend, and delegator yield are defined per month (calendar), then converted to per period, per the convention clarification in section 0. Growth follows F(n) = days times min(F_day0 times (1 plus g) to the power n, ceiling_day), saturating, not exponential without limit. Fee simulations use floating-point arithmetic and are labeled approximate; issuance figures use exact rational arithmetic where issuance is computed.

Two benchmarks are used throughout, both carried from v0.2: minimum need 14,625 illustrative units per month (10,000 validator operations, 3,125 delegator yield, 1,500 treasury spend), the survival figure; and period-0 validator pay 70,000 per period under geometric and 49,630 under hybrid, the full-replacement figure. A third figure is used for the valley analysis: validator operations plus treasury spend only, without delegator yield, is 11,500 per month. Fee distribution is illustrative, not selected: 80% to validators, 20% to the treasury, no burn in the base scenarios.

### 5.3 Scenario definitions

| Scenario | Starting daily fee revenue | Growth per period | Annual equivalent, monthly | Annual equivalent, quarterly | Daily ceiling | Blended average fee | Ceiling transactions per day |
|---|---|---|---|---|---|---|---|
| Low | 5 units per day | 1% | 12.7% per year | 4.1% per year | 50 per day | 0.04625 | 1,081 |
| Moderate | 20 per day | 3% | 42.6% per year | 12.6% per year | 600 per day | 0.04000 | 15,000 |
| High | 80 per day | 5% | 79.6% per year | 21.6% per year | 1,000 per day | 0.02925 | 34,188 |

Demand classes at maturity (ceiling): human transfers and payments, developer transactions, marketplace transactions, infrastructure usage, and AI agent transactions as a distinct class (many small payments, low value per transaction, fee 0.005). The scenarios are assumptions, not forecasts.

### 5.4 Results

Monthly periods. Fee revenue is gross, before the 80/20 split.

| Scenario | Year 10, period 120 | Year 20, period 240 | Year 30, period 360 | Minimum need per period | Period-0 validator pay, geometric | Period-0 validator pay, hybrid |
|---|---|---|---|---|---|---|
| Low | 495 | 1,500 (ceiling) | 1,500 (ceiling) | 14,625 | 70,000 | 49,630 |
| Moderate | 18,000 (ceiling) | 18,000 (ceiling) | 18,000 (ceiling) | 14,625 | 70,000 | 49,630 |
| High | 30,000 (ceiling) | 30,000 (ceiling) | 30,000 (ceiling) | 14,625 | 70,000 | 49,630 |

Quarterly periods (90 days). Fee revenue per period is 3 times the monthly figure at the same daily rate, and the minimum need per period is also 3 times, at 43,875. Period-0 validator pay per period is unchanged at 70,000 and 49,630, because issuance is per period.

| Scenario | Year 10, period 40 | Year 20, period 80 | Year 30, period 120 | Minimum need per period | Period-0 validator pay, geometric | Period-0 validator pay, hybrid |
|---|---|---|---|---|---|---|
| Low | 670 | 998 | 1,485 | 43,875 | 70,000 | 49,630 |
| Moderate | 5,872 | 19,154 | 54,000 (ceiling) | 43,875 | 70,000 | 49,630 |
| High | 50,688 | 90,000 (ceiling) | 90,000 (ceiling) | 43,875 | 70,000 | 49,630 |

When fee revenue covers the benchmarks:

| Scenario and mapping | Covers minimum need | Covers period-0 validator pay, geometric (70,000 per period) | Covers period-0 validator pay, hybrid (49,630 per period) |
|---|---|---|---|
| Low, monthly | Never (ceiling 1,500) | Never | Never |
| Low, quarterly | Never (ceiling 4,500) | Never | Never |
| Moderate, monthly | Period 109, 9.1 years | Never (ceiling 18,000) | Never |
| Moderate, quarterly | Period 109, 27.3 years | Never (ceiling 54,000) | Period 113, 28.3 years |
| High, monthly | Period 38, 3.2 years | Never (ceiling 30,000) | Never |
| High, quarterly | Period 38, 9.5 years | Period 47, 11.8 years | Period 40, 10.0 years |

Findings: only High covers the minimum need quickly (period 38 in both mappings, 3.2 years monthly, 9.5 years quarterly). Moderate covers it only at period 109, which is 9.1 years monthly but 27.3 years quarterly. Low never covers it. No scenario covers period-0 validator pay under monthly periods. The ceilings (1,500, 18,000, 30,000) are all below 49,630, let alone 70,000. Full replacement of launch-level validator pay from fees is out of reach in every monthly scenario modeled. Under quarterly periods, High covers both, and Moderate covers hybrid only at period 113, but that quarterly result is flattered by the period definition: 70,000 per 90 days is 23,333 per month, not 70,000 per month. On a per-month basis, High quarterly at its ceiling collects 1,000 units per day, or 30,000 per month, which is still below 70,000 per month. The honest statement is that no scenario replaces launch-level monthly validator pay.

Validators sustained, computed as (80% of fee revenue minus delegator yield per period) divided by validator cost per period, floored at zero. Monthly periods, year 10: Low sustains 0 net of yield; Moderate sustains 113; High sustains 209. Quarterly periods: Low sustains 0 in all 30 years; Moderate sustains 0 at year 10, 20 at year 20, 113 at year 30; High sustains 104 at year 10, 209 at year 20 and 30. At the v0.2 Cardano-like fee level (469 per period), the network sustains about 3 validators gross, and 0 net of delegator yield. The treasury share covers its 1,500 per month spend only in Moderate and High at maturity under monthly periods, and only in High at year 10 under quarterly periods.

Funding valley at hybrid issuance end (period 40): Moderate monthly has a 5.8-year valley totaling 572,847 units (17.2 years quarterly, 1,718,540 units); geometric has no valley in Moderate or High. High has no valley under either model. Low never closes within 30 years. The valley is extremely sensitive to growth: v0.2 growth at 5% per period gave a 26-period valley of 116,065 units, while Moderate at 3% per period gives a 69-period valley of 572,847 units monthly. The difference is the growth assumption, not a contradiction.

### 5.5 AI agent transactions as a distinct demand class

Agents are 45% of transactions but only 7.7% of fee revenue in High (0.5% in Low). Evidence: x402 averaged about 0.05 USD per payment in Q3 and October 2026, 87% of payments 0.0001 to 0.10 USD, 90.8% under 0.01 USD in one sample; TRM attributes only 0.6% to 7.5% of commerce-looking volume to agents; one wallet made 64% of Base x402 payments in six months. A fee of 0.005 is already 10% of a 0.05-value payment; the security-needed blended fee of 0.04 to 0.05 would be 80% to 100% of such a payment. Funding the minimum need from agent fees alone would need 97,500 agent transactions per day; High produces 15,385 per day. A flat fee cannot serve both security funding and agent viability; differentiation is open design work. Transaction counts must never be read as demand without a value distribution attached.

### 5.6 Protocol revenue versus OGO Systems corporate revenue

Pro subscriptions, AI inference, developer infrastructure, hosted applications, premium tools, and marketplace service fees are corporate revenue, not protocol income. They are not counted as protocol income anywhere in this report. One illustrative routing mechanism (M1, does not exist, not counted) is specified in the workstream draft: 10% of the marketplace service fee line only, converted to OGO Coin by open-market purchase, transferred monthly by company multisig to the treasury address with a memo and quarterly Stripe reconciliation, revocable, needing design, governance, and legal review plus pre-genesis disclosure before it could be counted. Break-even: at 10%, the corporate line must be 15,000 per month to cover the 1,500 treasury spend.

### 5.7 What happens if fee revenue does not cover the need

Low never sustains a single validator net of yield in 30 years under either mapping. The exit cascade from v0.2 applies: fee revenue below validator costs, marginal validators exit, stake concentrates on the survivors, the Nakamoto coefficient falls, and the chain's security degrades while its nominal issuance schedule is still being followed perfectly. The schedule can be exactly right and the network insecure at the same time. The mitigations are not a different curve but a different plan: a longer schedule, a lower minimum validator set, or an explicit decision that the chain accepts a smaller validator set.

## 6. Updated geometric versus hybrid comparison

This section uses the improved evidence: researched validator costs (section 2), the fee-market scenarios (section 5), the two-mode treasury rule (section 4), and the epoch-based period architecture (section 1). The gentle geometric taper (factor about 0.95 over the final 20 periods, renormalized) is preserved as the hybrid terminal variant. The taper halves the hybrid final payment from 11,000 to about 5,000 illustrative units and keeps every per-period drop under 7%, but it does not change any post-issuance figure in this section: the valley is caused by the end date, not by the size of the final payment.

### 6.1 The need is higher than v0.2 stated

At researched standard costs (200 USD per validator per month, 100 validators, illustrative price 1 USD), the minimum fee need is 24,625 per period, not 14,625. Recomputing the fee scenarios against 24,625, monthly periods, independently verified in consolidation:

| Scenario | Covers 14,625 (v0.2 baseline) | Covers 24,625 (standard-cost need) | Hybrid valley at 24,625 | Geometric at 24,625 |
|---|---|---|---|---|
| Low | Never (ceiling 1,500) | Never | Never closes within 360 periods | Never covered |
| Moderate | Period 109, 9.1 years | Never (ceiling 18,000) | Never closes within 360 periods | Never covered |
| High | Period 38, 3.2 years | Period 48, 4.0 years | 8 periods, total shortfall 35,659 | Covered by period 48, before issuance ends at period 110, no valley |

At researched costs, only the High scenario covers the minimum need, at period 48, and Moderate never does. The v0.2 finding that geometric has no funding valley in the growth scenario was against the lower need; at the standard-cost need, Moderate leaves both models with a permanent level gap, not a valley with a recovery date. A valley is a shortfall that ends; a level gap does not.

### 6.2 Comparison across the eight dimensions

| Dimension | Geometric r = 0.90 | Hybrid base 40%, N = 40, gentle taper |
|---|---|---|
| 10-year sustainability, monthly periods | Issuance ends period 110 (9.2y). Year 10 is post-issuance. Fee revenue must cover the need by then: only High does (period 38 at baseline need, period 48 at standard-cost need). | Issuance ends period 40 (3.3y). Year 10 is 6.7 years post-issuance. Fee revenue must cover the need by period 40: only High does (period 38 at baseline need; at standard-cost need, period 48, leaving an 8-period valley of 35,659). |
| 20-year sustainability, monthly periods | Post-issuance for the whole second decade. Moderate covers the baseline need at period 109, just before issuance ends; at standard-cost need it never covers. | Post-issuance for 16.7 years. Same fee coverage as at year 10; the model contributes nothing after period 40 except the treasury balance it left behind. |
| 30-year sustainability, monthly periods | Post-issuance for 20.8 years. Low never covers the need in 30 years under either model. | Same as geometric after period 40, except the treasury balance at issuance end differs (R-F preserves exactly 20,000 under both models; under the v0.2 P4 rule, geometric preserved 58,827 and hybrid 49,351). |
| Validator participation | Rewards decay smoothly; last period at or above 10% of R0 is period 21. Validators face a long, predictable decline and must plan the fee transition over 110 periods. | Rewards stay at or above 10% of R0 through period 39 by construction, then stop. Validators face a known minimum for 40 periods, then a step to fees-only. Participation is steadier during the schedule and more abrupt after it. |
| Treasury solvency | Under R-F, the floor of 20,000 is preserved exactly, and spending continues on fee inflow after issuance ends. Under v0.2 P4, geometric preserved 58,827 at issuance end. Maximum sustainable constant spend from issuance inflow alone is 909 per period. | Under R-F, the floor of 20,000 is preserved exactly. Under v0.2 P4, hybrid preserved 49,351 at issuance end. Maximum sustainable constant spend from issuance inflow alone is 2,500 per period, because the same inflow is spread over 40 periods instead of 110. |
| Fee-market transition | The fee market gets 110 periods to develop. In the growth scenario there is no valley at the baseline need. At the standard-cost need, Moderate never covers, so the transition fails on level, not on timing. | The fee market gets 40 periods. In the growth scenario at the baseline need there is a 26-period valley (116,065 total shortfall, treasury covering about half); at Moderate growth (3% per period) the valley is 69 periods and 572,847. The transition is concentrated at a known date. |
| Supply predictability | One parameter (r). Terminal period derived from R0 and r; prior reports recommend a fixed final period in genesis so the end date is known. Dust 65 uogo per 1,000,000. | Four parameters (base, B0, r, N) plus taper parameters if the taper is adopted. Finite by construction at period N. Dust 22 uogo per 1,000,000. |
| Distribution over time | 50% issued by period 7, 90% by period 22, first 10 periods take 65.1% of total. Most front-loaded of the two. | 50% issued by period 11, 90% by period 32, first 10 periods take 49.7% of total. The floor shifts issuance toward later participants. |
| Low-adoption survivability | Fails. Fee revenue never covers the need within 30 years; validator pay at issuance end is 236 per period under v0.2 low adoption, sustaining 2 validators at the illustrative cost. At researched standard costs the failure is worse: the same fee revenue sustains fewer validators. | Fails. Fee revenue never covers the need within 30 years; validator pay at issuance end is 7,816 per period only because issuance is still paying, then falls to the same fee level as geometric. The floor delays the failure; it does not prevent it. |
| Terminal issuance behavior | Ends when the per-period amount floors to zero (period 110), or at a fixed final period in genesis (recommended). No cliff in amount, but a long thin tail that is mostly symbolic late in the schedule. | Ends at period N by construction. Abrupt end has a final payment of 11,000 that disappears in one step. Gentle geometric taper (0.95 over final 20 periods, renormalized) halves the final payment to about 5,000 with max drop 6.5%, and is the only taper variant tested that actually smooths. A short linear taper is worse than abrupt (35.7% drops). |

### 6.3 The trade, restated with the new evidence

The comparison still turns on one trade, and the new evidence sharpens both sides of it. Hybrid buys a floor during the schedule and pays for it with an early, concentrated transition: the fee market gets 40 periods to develop, then faces a valley that is 26 periods and 116,065 units at 5% fee growth, or 69 periods and 572,847 units at 3% fee growth, at the baseline need. Geometric buys time and pays for it with a declining tail: the fee market gets 110 periods, and in the growth scenario there is no valley at all at the baseline need. Under low adoption both fail. The floor is insurance during the schedule; the tail is insurance at its end. They are different kinds of insurance, and the choice between them is a choice about which risk is worse: underfunding during the schedule, or a concentrated transition at its end.

What the new evidence adds: at researched validator costs, the fee transition is harder than v0.2 suggested for both models, because the minimum need is 24,625 rather than 14,625 at standard costs and 1 USD, and only the High fee scenario covers it. The model choice does not change what fees must provide in absolute terms. It changes when fees must provide it, and how much treasury is left at that date. No model is chosen here, and no parameter is selected.

## 7. Public cryptocurrency architecture comparison

### 7.1 The identity question first

Do not assume two similarly named assets are interchangeable. A shared name or ticker does not prove that two assets are interchangeable. The answer differs by option, and it determines everything else in this section.

- Option A: ambiguous by design, and the ambiguity is the risk. At launch, there is one asset: the public token on the existing chain. After a move, there are two possible states. Either (A1) the token and OGO Coin are the same asset in two forms, linked by a bridge or a burn-and-mint mechanism, so one unit on Base corresponds to one unit on OGO Network and total supply is conserved across both homes; or (A2) they are different assets, the old token is retired and replaced by OGO Coin, so holders must convert and the old token ceases to be the live asset. A1 and A2 are not the same promise, and they cannot both be true at once. The choice between them must be made and published before the first token is sold or distributed, because holders buy on the basis of which one they believe they hold.
- Option B: one asset, one home, no identity problem. The public asset is OGO Coin, native to OGO Network. There is no second public asset to confuse it with. The only remaining naming task is distinguishing OGO Coin from OGO Coins (platform ledger), which is a branding problem, not a cross-chain identity problem.
- Option C: one public asset, one home, no identity problem for the public asset. The public token is the asset, permanently. OGO Coin, if the name is still used for anything on OGO Network, would then be a different asset or must not be created at all. If OGO Network later issues its own native asset under a similar name, option C has silently become option A without the migration plan, which is the worst outcome. Option C therefore requires an explicit rule: either OGO Network has no native monetary asset, or its asset has a distinct name, distinct ticker, and distinct supply from the public token, stated in writing before launch.

What would have to be true for "same asset in two forms" to hold, under A1: (1) a single, published total supply rule that covers both chains, with one maximum that is enforced, not merely asserted, on both sides; (2) a single, auditable conservation invariant: every unit minted on the destination chain is backed by exactly one unit locked or burned on the source chain, and every unit released or re-minted on the source chain corresponds to exactly one unit burned on the destination chain, with no path that creates a unit without destroying or locking its counterpart; (3) a single, named canonical representation on each chain, with one contract address per chain published as canonical, and all other representations labeled wrapped, voucher, or non-canonical; (4) a single, named verification mechanism whose failure mode is stated in advance, and a published answer to what happens to the cap if that mechanism is compromised; (5) a single, named authority that can pause, upgrade, or wind down the link, with its powers, limits, and sunset stated before launch, or a proof that the link is safe without one. If any of these five is missing, the honest description is two different assets that happen to share a name, not one asset in two forms.

### 7.2 Comparison table

| Dimension | A: Existing chain now, OGO Network later | B: OGO Network from inception | C: Existing chain permanent, OGO Network repurposed |
|---|---|---|---|
| Public asset at launch | Public token on existing chain (conditionally Base ERC-20) | None until OGO Network genesis; then native OGO Coin | Public token on existing chain (conditionally Base ERC-20) |
| Are public token and OGO Coin the same asset? | Ambiguous until A1 vs A2 is chosen | Yes, trivially: one asset, OGO Coin, one home | No, or OGO Coin must not exist as a monetary asset; if it does, it is a different asset |
| Migration path | Required, either bridge (permanent link) or migration (one-way move) | None; genesis is the launch | None for the token; OGO Network purpose must be redefined separately |
| Bridging needed? | Yes under A1 (permanent), or temporarily under A2 (during a migration window), or not at all if a clean snapshot cutover is chosen | Only IBC to other Cosmos chains after launch, for OGO Coin as a Cosmos asset; no EVM bridge is inherent | No OGO-to-OGO-Network bridge for the token; ordinary third-party bridges to other chains are a separate, optional question |
| Whose consensus secures the asset? | At launch, the host chain (Base inherits Ethereum security via fraud proofs, with a 7-day challenge window). After a move, OGO Network consensus, plus the bridge's own trust assumptions for as long as the link lives | OGO Network consensus from day one: its own validator set, CometBFT finality, and governance | The host chain, permanently. OGO never controls the settlement layer |
| Security failure mode | Bridge compromise can create unbacked units on the destination chain; host-chain halt or rule change can freeze, reorg, or strand the token; OGO Network halt can strand the migrated side | OGO Network halt, validator failure, or governance capture directly threatens the asset; no host chain to fall back on; IBC counterparty failures can strand vouchers abroad | Host-chain halt, sequencer failure, fee spike, or rule change directly threatens the asset; OGO cannot fix the host chain, only wait, fork around it by migrating (which reopens option A), or accept the loss |
| Supply continuity proof | Hardest. Requires cross-chain conservation accounting, a single cap enforced on two ledgers, and a stated response to bridge compromise | Simplest. One ledger, one cap, enforced every block per Constitution Article VII, with IBC escrow and voucher reconciliation as the only cross-chain extension | Simple for the token itself: one contract, one totalSupply, one cap. Complexity moves to the OGO Network side, where any second asset must be kept strictly separate |
| Naming implications | Highest confusion risk. Same name on two chains, plus platform OGO Coins. Requires distinct labels for each form and a published canonical address per chain | Lowest confusion risk for the public asset. One name, one home. Platform ledger collision remains | Low for the public asset, but only if OGO Network never issues a similarly named asset. If it does, the confusion returns with no migration plan to justify it |
| Custody reintroduced? | Yes, under any lock-and-mint bridge: the escrow is a custodian in code. Under burn-and-mint with a multisig attester, the attester set is the custodian. Only a natively verified link (canonical rollup bridge, IBC light client) avoids a separate custodian, and even then the proving system is trusted | No bridge custodian for the asset itself. Custody questions move to validators, staking, and any custodial on-ramps, which are separate decisions | No OGO bridge custodian for the token. Custody is the host chain's design, not OGO's. OGO's interface must still satisfy the SEC 12-condition pattern if it prepares transactions |
| Monetary schedule applies to | Must be decided explicitly. It cannot silently apply to both homes independently | OGO Coin only, per the Decision Report schedule (geometric or hybrid, parameters TBD) | The token only, per whatever token schedule is chosen (Model A fixed supply or Model B ecosystem schedule, per feasibility section 4). OGO Network, if it has its own asset, needs its own separate schedule |
| Time to a public economy | Soonest, tied with C. Token can ship after Phase 2 wallet work and Phase 3 contract work, without waiting for the L1 | Slowest. No public economy until the L1 is built, audited, and launched, including validator set, genesis, and IBC readiness | Soonest, tied with A. Same Phase 2 and Phase 3 path, with no L1 dependency at all |
| What it forecloses | Nothing yet, but it creates a future migration liability that must be honored. Every holder acquired on the first chain becomes a migration obligation | Forecloses an early public economy on an existing chain, and forecloses the distribution, liquidity, and wallet ecosystem that comes with it. It does not foreclose later IBC expansion | Forecloses OGO Network as the token's home, permanently, unless the decision is reversed (which is option A with a costlier, later migration). Forecloses native-asset sovereignty for the public token |
| Reversibility | Reversible only by paying the migration cost. A bridge can be wound down; a migration cannot be un-migrated without a second migration | Reversible only by launching a token on an existing chain later, which is option A entered late, with the L1 already live | Reversible only by migrating, which is option A entered late, after holders, exchanges, and integrations have settled on the first chain |

### 7.3 Migration path, under option A

If option A is chosen and a move to OGO Network occurs, holders on the old chain must receive one of four treatments. The choice must be published before the first token is distributed, because it determines what holders actually own.

1. Snapshot and airdrop. A block height is fixed in advance. Balances on the old chain at that height are recorded. The same amounts are minted or allocated on OGO Network to corresponding addresses. Precedent: EOS, June 2018. EOS ERC-20 tokens were frozen on 1 June 2018, registration of EOS public keys closed 2 June 2018, and a snapshot of registered wallets determined mainnet distribution. The failure mode is documented: about 65% of users had not registered shortly before the deadline, and unregistered holders risked holding a frozen token with no guaranteed claim. A snapshot also strands anyone who buys on the old chain after the snapshot height, and it requires an address mapping when the two chains use different address formats (EVM hex addresses do not map automatically to Cosmos bech32 addresses).
2. Voluntary migration window (claim portal or swap contract). Holders send old tokens to a swap contract or burn address and receive new coins at a fixed ratio (illustrative 1:1, to be decided, not assumed) on OGO Network. The window is open for a stated period, for example an illustrative 12 months, then closes or stays open indefinitely. Precedent: TRON, June 2018, holders had until 26 June 2018 to swap ERC-20 TRX for mainnet TRX, with exchanges handling the swap for custodial holders. The advantage is that holders act when ready, and the old token's supply falls as migration proceeds, which keeps conservation auditable. The disadvantage is the tail: some holders never migrate, through inattention, lost keys, exchange custody, or death. An indefinite window leaves a permanent liability and a permanent old token. A closing window creates a deadline after which unmigrated tokens are worthless, which is a confiscation in effect even if it is not called one. The treatment of the tail must be decided in advance.
3. Forced migration (freeze and replace). The old token contract is frozen or paused at a fixed date, as EOS did, and only the new chain's asset is recognized thereafter. This is the cleanest supply outcome and the harshest holder outcome. It requires that the freeze mechanism exists in the old contract from deployment (an ERC-20 with a pause or freeze function, or a proxy upgrade path), which itself weakens the "fixed supply, no admin mint" story, because the same admin power that can freeze can do other things. It also requires exchange coordination.
4. Permanent dual existence (bridge, no migration). The old token stays live, and OGO Coin on OGO Network is its bridged counterpart (A1). There is no migration event and no deadline. Holders never have to act. The cost is permanent: the bridge, its escrow, its verifier set, and its supply accounting must be maintained forever, and the bridge remains the single most exploited component class in the industry. Dual existence also fragments liquidity and composability.

Recommendation for the decision record, not a decision: if option A is chosen, the migration treatment must be named in the same decision that chooses option A. A token launched with "we will figure out migration later" has already chosen treatment 4 by default, because holders, exchanges, and integrations will entrench the first chain, and any later forced move will be read as a breach of the implied promise.

Migration mechanics that must be specified, whichever treatment is chosen: address mapping (EVM to bech32, with a fallback for holders who never register); ratio and rounding (the Decision Report dust finding applies by analogy: any per-holder flooring leaves dust, and its treatment must be stated); supply reconciliation at cutover (total old-chain supply, total new-chain genesis allocation, and total unmigrated remainder as three numbers that sum correctly; if the new chain mints the full old supply at genesis and the old chain is not frozen, total supply has doubled on paper); exchange and custodian coordination; tax and legal treatment (a migration can be a taxable event depending on jurisdiction, a question for counsel before announcement); and communication and phishing (the official swap contract address and claim portal published through channels holders already trust, before the migration opens; OGO will never ask for seed phrases).

### 7.4 Bridging: lock-and-mint vs burn-and-mint vs native issuance, and bridge security as a risk class

Three mechanisms: lock-and-mint (tokens locked in escrow on the source chain, vouchers minted on the destination chain); burn-and-mint (tokens burned on the source chain, new tokens minted on the destination chain on attestation); native issuance (each chain issues its own supply independently, with no link, which is two assets, not one asset in two forms). Canonical vs third-party is the distinction that matters most: a canonical bridge exists only between a rollup and its settlement layer (7-day challenge window for Optimism and Arbitrum). There is no canonical bridge between Base and a Cosmos chain. Any Base to OGO Network link is third-party, or an IBC extension with a new trust assumption. IBC is natively verified only between Cosmos chains.

Bridges are the most exploited component class in crypto. Cumulative losses exceed 2.8B USD since 2022 (another aggregation: over 4B USD across about 49 incidents, June 2021 to September 2024). Largest: Ronin 624M USD (5-of-9 keys, undetected 5 to 6 days), Poly 612M USD, BNB Bridge 568M USD (proof flaw), Wormhole 320 to 326M USD (signature bypass), Nomad 190M USD (trusted root 0x00), Harmony 100M USD (2-of-5 keys). Failure classes: key compromise, verification bug, upgrade or initialization bug, single-operator failure.

Custody returns via the bridge: lock-and-mint escrow is custody in code, and burn-and-mint with an attester makes the attester custodian of the link. Only a natively verified link preserves non-custodial status. If OGO operates the bridge, "OGO is non-custodial" is false for the bridged supply.

### 7.5 Supply continuity

Accounting rule: under lock-and-mint, economic supply equals S_base plus S_ogo minus S_locked (summing both totalSupply values double counts the locked units). Under burn-and-mint, economic supply equals S_base plus S_ogo. A bridge compromise that mints unbacked vouchers violates the cap in economic terms even if source-chain code never minted beyond M. Incident response (pause, blacklist, haircut, reimbursement, relaunch) must be written before launch, plus a cap on the share of supply bridged at any time. "Proven" means: the conservation invariant is checkable by any outsider from on-chain state on both chains, the cap is enforced by code on the issuance side, and the response to a violation is pre-written rather than improvised.

### 7.6 Interaction with the monetary schedule and the Base recommendation

The schedule must apply to exactly one issuance authority. If the Decision Report schedule applies independently on Base and on OGO Network, issuance doubles and the cap breaks. It must be assigned to exactly one chain in writing, with the other chain's mint authority disabled or subordinated in code. The feasibility report's Base recommendation is carried forward, conditional on the architecture choice: Base is primary if option A or C is chosen (EVM, sub-cent fees, about 2 second confirmations, Coinbase integration, OGO's TypeScript stack fits directly); Solana is the leading alternative if high-frequency micropayments dominate; Polygon is the fallback. If option B is chosen, the chain comparison is moot. Sequencing: A and C tie for soonest (Phase 2 and 3 token path, no L1 wait); B is slowest (no public economy until the L1 launches). A forecloses a costless move (migration liability grows with every holder and listing); C forecloses OGO Network as token home; B forecloses early economy and Base distribution.

### 7.7 Naming treatment

The three-asset collision stands: OGO Token (public token), OGO Coin (OGO Network native asset), OGO Coins (platform ledger). Proposed rules, for decision, not decided here: one public name and ticker distinct from the platform ledger; distinct chain-form labels with contract address shown wherever the asset is displayed; the platform ledger never called a token, coin, or wallet in user copy. Naming decisions that must be taken before launch: the public token's name and ticker; OGO Coin's name and ticker, if it exists; what each chain-form is called; and what happens to the names if the home changes.

## 8. Revised decision dependency map

v0.2 section 9 gave the decision order as: period length first, then minimum validator set and validator cost basis, treasury spend level and post-issuance rule, post-issuance funding mechanism and fee share, then model family, schedule parameters, split, and max supply as G + S_actual. The new evidence revises and extends that order. Changes from v0.2: period architecture is now a decision in its own right, before period length; the validator cost basis is now researched rather than placeholder, which makes the minimum validator set decision concrete; the treasury rule is now specified (R-F) rather than open; the fee policy decision now has scenario evidence and a live experiment to track (Osmosis); and the public cryptocurrency architecture option is now a decision that interacts with the monetary schedule, not a separate track.

1. Period architecture. Epoch-based periods via an epochs module, per section 1. This is a recommendation, not a decision. It determines how every later number is expressed in calendar time.
2. Period length. How many seconds is one epoch (D), or equivalently how many epochs per nominal year (12 monthly or 4 quarterly). Without it, no schedule parameter has a calendar meaning, and no growth rate has an annual meaning. A nominal year must be defined in genesis as a fixed number of seconds.
3. Public cryptocurrency architecture option (A, B, or C) and, under A, the identity choice (A1 same asset in two forms, or A2 different assets with conversion) and the migration treatment. This decision must come before any token distribution, because it determines what holders own and whether the monetary schedule applies to the token, to OGO Coin, or to both. If option B is chosen, the chain comparison is moot. If option A or C is chosen, the Base recommendation is conditional on it.
4. Minimum validator set and validator cost basis. Definition A (4 validators, about 300 to 800 USD per month) or Definition B (at least 50 validators, Nakamoto coefficient at least 10, top-10 share below 33%, no country or provider above 33%, about 10,000 to 50,000 USD per month infrastructure only), or a different threshold Benross sets. This sets the minimum fee requirement, which is the survival threshold for the whole design. At researched standard costs the minimum need is 24,625 per period, not the v0.2 baseline of 14,625.
5. Treasury spend level and two-mode spending rule, including the post-issuance clause. The recommended default rule is R-F (section 4), with its parameters (trailing window, balance percentage, reserve floor) set as genesis parameters and disclosed before genesis. The level must be affordable under the chosen model or justified by named post-issuance revenue. The affordable constant rates from issuance inflow alone are 909 per period under geometric and 2,500 under hybrid.
6. Post-issuance funding mechanism and fee share, disclosed before genesis. The mechanism with the strongest claim is a disclosed share of transaction fees to the treasury. Coretime and blockspace sales are supplements, not the plan. Endowment yield cannot sustain the illustrative spend at any plausible yield. Treasury staking is circular. A corporate revenue routing mechanism (M1) does not exist and cannot be counted until designed, governed, and disclosed.
7. Fee policy evidence. What fee market is expected, at what average fee and volume, and when is it expected to be sufficient. The scenarios in section 5 are the evidence base; the Osmosis fee-funded experiment (Proposals 1052 and 1053, inflation ended 26 October 2026) should be tracked before any fee assumption is treated as plausible. The agent-transaction fee tension (a flat fee cannot serve both security funding and agent viability) needs a design answer before a fee policy is set.
8. Model family (geometric or hybrid), with the section 6 evidence attached. The choice is between insurance during the schedule (hybrid floor) and insurance at its end (geometric tail).
9. Schedule parameters for the chosen family, then the split, then G and S_actual chosen so that G + S_actual = 100,000,000,000 OGO (approved hard maximum supply, 2026-11-08), with S_actual computed at genesis by exact integer summation and the dust left permanently unissued.
10. Naming and branding for the three OGO assets, before any public launch.
11. Written legal analysis (securities, money transmission, state licensing, tax, international, exchange admission) before genesis distribution or any token distribution. Not yet commissioned.

## 9. Remaining unresolved decisions

1. Period architecture: epoch-based is recommended; needs Benross's confirmation. Period length (D in seconds, or epochs per nominal year) is undecided.
2. Public cryptocurrency architecture: option A, B, or C is undecided. Under A, A1 vs A2 and the migration treatment are undecided. The schedule's assignment to exactly one issuance authority is undecided.
3. Economic model for the public token (feasibility Model A fixed-supply asset vs Model B ecosystem utility token) is undecided, and must be consistent with the OGO Network monetary philosophy or explicitly different.
4. Monetary model family: geometric or hybrid is undecided. If hybrid, the base level needs its own justification, and the taper parameters (factor, window, renormalized or not) need confirmation.
5. Minimum validator set: Definition A or Definition B, or a different threshold, is undecided. The mechanism that produces a flat distribution (delegation incentives, saturation cap, per-validator subsidy, genesis selection) is undecided.
6. Validator cost basis at scale: researched costs are for single validators; the cost of a 50 to 100 validator set with geographic spread, and the labour model, need confirmation before any budget.
7. Delegator yield: what yield delegators require, and what opportunity cost they compare against, is unknown.
8. Treasury spend level: what spend rate is actually needed, and is it affordable under the chosen model? The affordable constant rates are 909 per period (geometric) and 2,500 per period (hybrid) against issuance inflow alone.
9. Treasury rule parameters: trailing window, balance percentage, reserve floor, emergency reserve triggers and caps, and valley subsidy rules are specified as formulas in section 4 but their parameter values are undecided.
10. Post-issuance funding mechanism and fee share: which mechanism or combination, and what share, is undecided. Fee demand estimates are needed before any of them can be relied on.
11. Fee policy: average fee, fee distribution (validator vs treasury vs burn), and whether agent transactions are priced differently, are undecided. The flat-fee tension in section 5.5 needs a design answer.
12. Split: validator, delegator, community, and treasury shares, and whether the split is fixed per period or varies over time, are undecided.
13. Dust convention (M = G + S_actual, dust left permanently unissued) is recommended and needs Benross's confirmation.
14. Naming: the public token's name and ticker, OGO Coin's name and ticker if it exists, chain-form labels, and the platform ledger's naming treatment are undecided.
15. Conversion mechanism between Platform Credits and the public token, if any, with legal review, is undecided.
16. Legal analysis: written legal analysis is required before genesis distribution or any token distribution. Not yet commissioned.
17. Open research questions carried from the comparative research: Solana SIMD-0550 activation date (passed, not yet live as of October 2026); Polkadot exact post-cut issuance (55 vs 56.9m DOT per year); Polkadot Dynamic Allocation Pool coretime timing; Cardano Net Change Limit (350m vs 500m ADA); Cardano minimum pool cost (340 vs 170 ADA per epoch); Osmosis fee-funded staking outcome (Proposals 1052 and 1053, inflation ended 26 October 2026).

## 10. Executive recommendation

Recommendation, not a decision. The decision stays with Benross.

### 10.1 Decisions that are ready

These have enough evidence to be decided now, with the evidence cited. Ready does not mean decided; each still needs Benross's explicit approval.

1. Three-number accounting. Adopt S_target, S_actual, and M = G + S_actual, computed at genesis by exact integer summation, with dust left permanently unissued. This is model-independent and should be adopted regardless of any other decision. The v0.2 errata in section 0 should be recorded alongside it.
2. Period architecture. Adopt epoch-based periods via an epochs module, per section 1. The evidence is the drift quantification (10% slower blocks move the geometric end by 11 months to 2.8 years) and the halt behavior (missed epochs queue in order, preserving S_actual). The length decision remains separate.
3. Treasury spending rule. Adopt R-F as the default rule (section 4), with parameters set as genesis parameters and disclosed before genesis. It is the only rule tested that reaches full desired spend within 4 periods of genesis, spends exactly the P3 total, preserves exactly the 20,000 floor at t=360 in both models, and never freezes once any fee inflow exists. The emergency reserve policy and the transition rule (keyed to the genesis-fixed final issuance period, no vote) are part of the same adoption.
4. Gentle geometric taper, if hybrid is chosen. Adopt the gentle geometric base taper (factor about 0.95 over the final 20 periods, renormalized). Do not adopt a short linear taper, which is worse than abrupt.
5. Naming rules. Adopt the section 7.7 naming rules (one public name and ticker distinct from the platform ledger; distinct chain-form labels with contract address shown; the platform ledger never called a token, coin, or wallet in user copy) as a standing rule before any public launch.
6. Chain recommendation, conditional. Base remains the primary recommendation if option A or C is chosen, per the feasibility report. If option B is chosen, the chain comparison is moot.

### 10.2 Decisions that still need evidence

These cannot be decided responsibly on the current evidence. Each names what evidence is missing.

1. Period length. Needs the schedule parameters it will be chosen with. It is first in the decision order (section 8) precisely because everything else converts through it, but it cannot be chosen alone.
2. Public cryptocurrency architecture (A, B, or C) and identity (A1 vs A2). Needs the migration and bridging analysis in section 7 to be weighed against OGO's timing: A and C ship soonest, B is slowest. The identity choice must be published before the first token is distributed.
3. Minimum validator set. Needs Benross's view on whether Definition B's thresholds (Nakamoto coefficient at least 10, top-10 below 33%, at least 50 validators) are the right bar, and what mechanism would produce a flat distribution. No live Cosmos chain meets all four thresholds, so the bar is aspirational until a mechanism is named.
4. Treasury spend level. Needs a real spending plan: what the treasury must fund, at what rate, and whether that rate is affordable under the chosen model or justified by named post-issuance revenue. The affordable constant rates (909 geometric, 2,500 hybrid) are ceilings from issuance inflow alone, not budgets.
5. Post-issuance funding mechanism and fee share. Needs fee-demand evidence. The scenarios in section 5 are models, not forecasts, and the Osmosis experiment has no outcome data yet. No mechanism should be counted in the sustainability model until its share is disclosed and its demand is evidenced.
6. Fee policy. Needs the agent-transaction tension resolved (a flat fee cannot serve both security funding and agent viability) and a decision on fee distribution. The fee scenarios show what must be true; they do not show what will be true.
7. Model family (geometric or hybrid). Needs the decisions above it in the order: period length, validator set, spend level, funding mechanism, and fee evidence. The section 6 comparison is the evidence base, and it is complete. What is missing is not more comparison, but the upstream decisions that determine which model's advantages matter.
8. Split and schedule parameters. These follow the model family, in that order. The maximum supply number is approved as the ceiling (100,000,000,000 OGO, 2026-11-08); G and S_actual are chosen so that G + S_actual equals the approved ceiling.
9. Legal analysis. Written legal analysis must be commissioned before genesis distribution or any token distribution. This is the longest-lead item and should be started before the other decisions are finalized, not after.

### 10.3 The one-sentence summary

The monetary architecture is now specified enough to decide the accounting, the period definition, and the treasury rule, and honest enough to show that the fee transition is the binding constraint: at researched validator costs only the High fee scenario covers the minimum need, and no scenario replaces launch-level validator pay, so the decisions that matter most are the ones that determine fee demand, validator count, and spend level, not the choice between two issuance curves.

## Assumptions

1. Total scheduled issuance normalized to 1,000,000 illustrative units in all simulations. Genesis supply G = 500,000 illustrative, carried from the v0.1 example.
2. Geometric r = 9/10, R0 = 100,000. Hybrid base 40% (10,000 per period), N = 40, r = 9/10, R0 = 70,900. Illustrative split 70/20/10. Gentle geometric taper factor about 0.95 over the final 20 periods, renormalized, where stated.
3. Periods are uniform in length. Period length in blocks or time is TBD. Calendar mappings use illustrative monthly (12 per nominal year) and quarterly (4 per nominal year) assumptions. A nominal year is defined in genesis as a fixed number of seconds.
4. Desired treasury spend 1,500 per period, reserve floor 20,000, balance cap 5% per period, trailing window 4 periods, unless a rule states otherwise. All illustrative.
5. Validator operating costs are researched USD per month (section 2). OGO prices in section 2 are illustrative sensitivities, not facts. Validator count, stake fraction, and delegator yield assumptions are stated where used.
6. Fee scenarios (Low, Moderate, High) are assumptions, not forecasts. Fee revenue growth is saturating, not exponential without limit. Fee simulations use floating-point arithmetic and are labeled approximate; issuance figures use exact rational arithmetic, floored once to integer uogo.
7. Cardano, Ethereum, Solana, Polkadot, Cosmos Hub, Osmosis, and Base figures are point-in-time snapshots from research dated 6 November 2026, as recorded in the Decision Report and the workstream drafts.
8. Bridge exploit figures are cumulative losses reported by public aggregations; individual incident figures are as reported in the workstream draft's sources.

## Sources

Prior OGO reports: Monetary Design Decision Report v0.1; Geometric vs Hybrid Deep-Dive v0.1; Issuance Mathematics v0.1; Hybrid Terminal Transition and Treasury Sustainability v0.1; Treasury Sustainability v0.2; Public Cryptocurrency Feasibility v0.1; Monetary Constitution v0.1. Workstream drafts and computation scripts: /tmp/decision-prep/period-architecture.md (calc_period_drift.py); validator-costs-and-decentralization.md (recompute_fee_need.py); two-mode-treasury.md (simulate.py, simulate_modeB.py, simulate_rec.py, results_modeA.json, results_modeB.json); fee-market-adoption.md (compute_fee_model.py); public-crypto-architecture.md. Consolidation verification script: /tmp/verify_prep.py and /tmp/synth_item6.py. External research is as cited in the workstream drafts, dated 6 November 2026.
