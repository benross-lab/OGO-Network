# OGO Network: Monetary Design Decision Report v0.1

Status: Research and design only. No monetary model has been approved. No maximum supply, no percentages, no intervals selected.
Date: 2026-11-06
Project: OGO Network

## 0. Two corrections to earlier premises

Before the comparison, two characterizations in the briefing handoff were checked against primary sources and found wrong. They are corrected here, not carried forward.

1. Polkadot was NOT a fixed-maximum-supply network historically. DOT was uncapped with roughly 10% annual inflation until November 2024, when Referendum 1139 fixed issuance at 120m DOT per year with no cap. The 2.1 billion DOT cap came only with Referendum 1710, passed in September 2025 with about 81% support and implemented in January 2026. Any description of DOT as fixed-maximum-supply before September 2025 is wrong.
2. Ethereum did NOT cut its staking issuance curve in 2024. Dencun EIP-7514 (13 March 2024) capped validator activation churn at 8 per epoch, which brakes the growth of issuance, not the issuance rate. The April 2024 Electra issuance-curve proposal (about a 30% real-yield cut) was debated and not adopted. The successor proposal, EIP-8363, was drafted in July 2026 and withdrawn from Hegota consideration on 1 October 2026.

All network figures below are point-in-time snapshots from research dated 6 November 2026, drawn from primary documentation and reporting read that day. Where sources disagree, the disagreement is stated, not resolved silently. Figures move daily and by epoch; none should be treated as a live chain read.

## 1. Geometric versus hybrid comparison

Both candidates are one family with a floor dial. Hybrid at a 0% base is truncated geometric decay. The central decision is geometric issuance with no floor, or geometric issuance with a predetermined floor and a fixed end date.

All OGO figures in this report are normalized to 1,000,000 illustrative units of total scheduled issuance so shapes can be compared. The normalization is not a proposed supply. Reduction factors, base shares, and the 70/20/10 split used in illustrations are hypothetical inputs.

Geometric decay: R_n = floor(R0 * r^n), r rational, 0 < r < 1. Analytic total S = R0 / (1 - r). One tuning parameter.

Hybrid: R_n = base + floor(B0 * r^n) for n < N, 0 after period N. Analytic total S = base * N + B0 * (1 - r^N) / (1 - r). Four parameters: base, B0, r, N.

| Property | Geometric (r = 0.90) | Hybrid (40% base, N = 40, r = 0.90) |
|---|---|---|
| 50% of scheduled issuance by | period 7 | period 11 |
| 90% by | period 22 | period 32 |
| 99% by | period 44 | period 40 |
| Issued by period 20 | 87.8% | 73.5% |
| Issuance at period 20 vs period 0 | 12.2% | 24.5% |
| Largest single-period drop | 10% | 8.6% |
| Finite by construction | No, needs an end rule | Yes, ends at period N |
| Parameters to agree | 2 (R0, r) | 4 (base, B0, r, N) |
| Floor | None | Base floor for the whole schedule |

## 2. Reduction-factor sensitivity analysis

Exact rational arithmetic, floored to integer uogo, total scheduled issuance normalized to 1,000,000 illustrative units.

| r | R0 | Terminal period (floor reaches 0) | 50% by | 90% by | 99% by | R20 as % of R0 | Dust (uogo) |
|---|---|---|---|---|---|---|---|
| 17/20 (0.85) | 150,000 | 74 | period 5 | period 15 | period 29 | 3.9% | 45 |
| 9/10 (0.90) | 100,000 | 110 | period 7 | period 22 | period 44 | 12.2% | 65 |
| 19/20 (0.95) | 50,000 | 211 | period 14 | period 45 | period 90 | 35.8% | 118 |

A change of only 0.05 in r more than doubles the sustainability horizon and roughly triples the terminal horizon. r is the whole geometric design, and its selection needs its own review. A higher r pays validators longer at a lower initial rate for the same total. A lower r bootstraps harder and ends sooner.

## 3. Floor sensitivity analysis

Hybrid, N = 40 periods, r = 9/10, base share of total scheduled issuance varied.

| Base share of total | Base per period | R0 | 50% by | 90% by | Last period R >= 10% of R0 | First-10-periods share |
|---|---|---|---|---|---|---|
| 0% | 0 | 101,500 | period 7 | period 21 | period 21 | 66.1% |
| 20% | 5,000 | 86,200 | period 8 | period 28 | period 29 | 57.9% |
| 40% | 10,000 | 70,900 | period 11 | period 32 | period 39 | 49.7% |
| 60% | 15,000 | 55,600 | period 14 | period 34 | period 39 | 41.4% |

The base share is the floor dial. At 40% base, issuance never falls below 14.1% of its initial level while the schedule runs. Every 20 points of base share moves 90% issuance roughly 4 to 7 periods later and roughly doubles the period-20 issuance share. The floor is the most effective tool in this comparison for shifting issuance toward later participants, more effective than raising r, because it redirects issuance from early periods to every later period rather than stretching the same early-heavy curve.

## 4. Validator and delegator sustainability

Sustainability is measured as the last period in which issuance is at least 10% (and 5%) of its initial level, on the assumption that validator operating costs do not fall with issuance. Validator and delegator rewards are a fixed share of R_n each period under the illustrative split, so these horizons apply directly to reward income.

| Model | Last period R >= 10% of R0 | Last period R >= 5% of R0 |
|---|---|---|
| Geometric r = 0.85 | period 14 | period 18 |
| Geometric r = 0.90 | period 21 | period 28 |
| Geometric r = 0.95 | period 44 | period 58 |
| Hybrid base 20% | period 29 | period 39 |
| Hybrid base 40% | period 39 | period 39 |
| Hybrid base 60% | period 39 | period 39 |

What validator income level is actually sustainable depends on validator costs, stake size, and fee income, none of which are known yet. These horizons are relative measures, not viability findings. Further analysis must examine validator operating costs, staking participation, transaction demand, fee distribution, and network security.

## 5. Treasury funding projections

The treasury inherits the issuance curve. Under the illustrative split of 70% validators/delegators, 20% community programs, 10% treasury, each bucket receives its share of R_n per period. The split is illustrative only.

Geometric r = 0.90: treasury receives 10,000 per period at period 0, 1,215 at period 20, 147 at period 40. Cumulative treasury by period 20 is 87,833 of its 99,993 total. The treasury is funded overwhelmingly early and then starves.

Hybrid base 40%: treasury receives 7,090 at period 0, 1,740 at period 20, and 0 at period 40 because the schedule has ended. Cumulative treasury by period 20 is 73,488 of its 99,998 total. Less front-loaded, steadier while it runs, then stops at the known end date.

Neither model funds a treasury forever. If community programs need funding in the later schedule, the floor matters more than the total. If the treasury is meant to be built early and then spent down, geometric front-loading is a feature.

## 6. Fee-transition scenarios

Required fee revenue to hold validator pay at its period-0 level is R0 - R_n, as a percentage of R0.

| Model | Fees must replace by period 20 | Fees must replace by period 40 |
|---|---|---|
| Geometric r = 0.85 | 96.1% | 99.9% |
| Geometric r = 0.90 | 87.8% | 98.5% |
| Geometric r = 0.95 | 64.2% | 87.2% |
| Hybrid base 20% | 82.7% | 100.0% (schedule ended) |
| Hybrid base 40% | 75.5% | 100.0% (schedule ended) |
| Hybrid base 60% | 64.1% | 100.0% (schedule ended) |

Hybrid makes the transition gentler during the schedule, but its end is still a step to 100% fee-funded operation at period N, from a floor rather than from a trickle. If the end step is a concern, the mitigation is a longer N or a final taper of the base itself, which is a further design option.

The established networks show how hard this transition is in practice. All four were still issuance-funded in November 2026; none has completed a fee-funded transition. Cardano collected 3.3m ADA in transaction fees against 493.7m ADA in staking rewards over the 73 epochs to 1 September 2026, so fees covered about 0.67% of rewards. Ethereum's EIP-1559 burn offset only about 2.4% of issuance in 2026, with the burn collapsing from 8,844 ETH per day in 2021 to about 57 per day in 2026. Solana's base-fee burn is far below its roughly 22.8m SOL per year issuance. OGO must plan on the assumption that fee revenue will not replace issuance quickly, and must state in advance what happens when issuance ends.

## 7. Exact arithmetic and rounding specifications

- All schedule parameters are integers or rational numbers p/q. No floating-point arithmetic in consensus-critical monetary calculations.
- Per-period issuance is computed as an exact rational and floored once to integer uogo.
- The fractional dust from flooring is never minted by any other path. Illustrative dust is 45 to 118 uogo per 1,000,000 normalized units. Recommendation: leave dust permanently unissued. Never permit rounding dust to become an alternative minting allowance.
- Periods are 0-indexed. Period n covers block heights [start + n*L, start + (n+1)*L - 1], where start is the genesis start height and L is the period length in blocks (TBD). Issuance for period n is claimed once, tracked by an issued-by-period set, and a second claim for the same period fails.
- Cumulative issuance is computed with integer addition, which is associative and exact on every machine. Independent validators with the same genesis parameters produce the same R_n for every n.

Why this matters, demonstrated: the binary double 0.9 is not 9/10, and 0.9 to the 100th power in float differs from the exact rational (9/10)^100 in its last digits. Summing the same 110 float period amounts forward gives 999990.7386128694 and in reverse gives 999990.7386128692. Two validators summing in different orders would disagree. Exact rational arithmetic removes the whole class of error.

## 8. Terminal issuance rules

- Geometric decay has no natural end in real numbers. In integer uogo it ends when floor(R0 * r^n) reaches 0. With R0 = 100,000 and r = 9/10, that is period 110 (R_109 = 1, R_110 would be 0). Recommendation: specify a fixed final period in genesis, so the end date is known at genesis instead of derived.
- Hybrid ends by construction at period N. Its terminal rule is the same cap check on the final period.
- The final period's issuance is min(schedule amount, M - total_supply). Once total_supply reaches M, issuance is zero forever.
- A floor defined as a fixed absolute amount per period cannot be paid from a finite reserve past the end date. A floor defined as a share of the remaining reserve never ends. The definition of the floor must be settled before the hybrid model can be finalized: OGO's hybrid floor is a fixed absolute base rate for a fixed number of periods, ending at period N. That is the definition used throughout this report.
- Polkadot's Pi Schedule has the same asymptotic property OGO must avoid by this rule: its cap is only approached asymptotically, with one source projecting it is reached around 2160. A true fixed end date needs a final tranche rule, which Polkadot does not have.

## 9. Comparison with Solana, Cardano, Polkadot, and Ethereum

### Solana: declining inflation with a permanent minimum

No maximum supply. Issuance is minted each epoch to validators and delegators pro-rata to stake. Schedule: 8% per year from 10 February 2021, declining 15% per year, terminal rate 1.5% per year, so rate(t) = max(1.5%, 8% x 0.85^t). Rate was about 3.58 to 3.62% in early November 2026, about 62,415 SOL per day (about 22.8m per year). SIMD-0228, the market-based inflation proposal targeting a 50% staking ratio, FAILED in March 2025 with 61.4% yes against a 66.67% supermajority threshold, on 74% turnout. SIMD-0550 (doubling the taper to 30%, terminal 1.5% unchanged) PASSED on 28 August 2026 with 67.001%, but was NOT YET LIVE as of October 2026, with no public activation date; it would reach 1.5% around the first half of 2029 instead of around the first half of 2032. There is no protocol treasury: 0% of issuance or fees funds development; ecosystem funding comes from Solana Foundation grants. Fees: base fee 5,000 lamports per signature, 50% burned and 50% to the validator; priority fees 100% to the validator since SIMD-0096. No planned end to issuance: the 1.5% floor is permanent. Criticisms: validators vote on their own pay; small validators pay about 1 to 1.6 SOL per day in fixed vote costs; Nakamoto coefficient about 18; no protocol slashing. Lesson for OGO: a fixed geometric schedule is predictable and survived a contested vote, but the terminal floor means disinflation never becomes no inflation, and letting issuance recipients vote on issuance produced a near-miss followed by a razor-thin pass.

### Cardano: declining reserve-based issuance and treasury funding (focus)

YES, hard cap: 45,000,000,000 ADA, in the Constitution, Tenet 10, ratified February 2025. Genesis: 25.927bn public, 5.185bn founders, 13.888bn reserve. Issuance is a release from the finite reserve, not on-demand minting: each 5-day epoch, rho = 0.3% of the remaining reserve plus all transaction fees go into a reward pot; tau = 20% of the pot goes to the Treasury and 80% to stake pools and delegators. Do not use the old 0.22% rho figure from a 2021 forum post; 0.3% is the current documented value. Smooth exponential decay with no floor and no fixed end date. Reserve half-life 4 to 5 years per the docs. Reserve about 6.13bn ADA at epoch 655 (13.62% of max), down from about 13.3bn at epoch 209. Fee transition is the explicit design assumption but has not happened: fees covered about 0.67% of rewards over the year to 1 September 2026. Treasury about 1.38 to 1.49bn ADA. Spending is by on-chain governance (CIP-1694) and bounded by a Net Change Limit. Unresolved conflict: Blockworks reports the 2026-27 NCL as 350m ADA; October 2026 dashboards report 500m ADA for epochs 613-713 with 457.4m committed (91.5%) and 42.6m remaining. Do not quote a single NCL figure as settled. Criticisms: the fee-revenue gap is the central risk; no slashing; pool-economics gaming around k = 500 and the minimum pool cost (340 vs 170 ADA per epoch, unresolved); governance friction, with the spending allowance nearly exhausted in October 2026 while the treasury balance remained large. Lesson for OGO: Cardano is the closest fixed-max analogue and the cautionary tale. Its reserve model IS geometric decay with no floor, and the consequence visible after about six years is a slow, predictable decline in nominal rewards with no scheduled landing point. Adding a predetermined floor plus a fixed end date is precisely the part Cardano does not have.

### Polkadot: fixed maximum supply with scheduled issuance reductions (focus, corrected timeline)

YES, capped since 2026: 2.1bn DOT. The timeline matters. 2020 to November 2024: NO cap, about 10% per year inflation with a dynamic ideal staking rate of 45 to 75%. November 2024: Referendums 1139 and 1271 fixed issuance at 120m DOT per year, 85% to stakers and 15% to treasury, still NO cap. September 2025: Referendum 1710 passed with about 81% support, implemented January 2026, capping total supply at 2.1bn DOT. From 14 March 2026, the Pi Schedule: every two years, annual issuance is reset to 13.14% of the remaining gap to the cap. The first step cut annual issuance by 53.6%, from 120m to about 55 to 57m DOT per year (sources differ on the last digit; report as a range). Next step: 14 March 2028. The cap is approached only asymptotically; one source projects it is reached around 2160. The cap is a governance parameter, not an immutable genesis constant, and can in principle be changed by referendum. Treasury: from 12 March 2026 (runtime 2.1.0, Referendum 1827) the Dynamic Allocation Pool collects new issuance and slashes, the 1% per 24-day treasury burn was halted, and fees and coretime revenue are phased in later. Treasury spending was about $133m in 2024 (about $87m in the first half of 2024, including about $37m on advertising), widely criticised as excessive. Treasury inflow is projected to fall from 18m DOT per year to about 8.3m, then about 6.2m, then about 4.5m DOT per year under the Pi Schedule (Referendum 1781 proposal projections, not an enacted budget). Lesson for OGO: Polkadot is the most instructive case for the hybrid candidate, because it actually converted an uncapped chain into a capped, scheduled one by governance, and the conversion required a referendum plus a runtime upgrade, acceptance that the cap is only approached asymptotically, and a simultaneous redesign of treasury funding (the Dynamic Allocation Pool) because the old treasury maths broke under the new issuance. OGO should note that a capped chain funding its treasury from issuance faces a shrinking real budget unless fee or coretime revenue grows.

### Ethereum: dynamic staking issuance and fee burning, without a fixed maximum supply

No maximum supply. Issuance goes only to validators. Total issuance is proportional to the square root of total ETH staked (base reward factor 64), so per-validator yield falls as 1 over the square root of stake, with about a 1.5% floor. Current (November 2026): about 42 to 44m ETH staked (about 34 to 36%), about 1.08m ETH per year issuance, about 2,970 per day, base yield about 2.6 to 2.7%, plus tips and MEV. EIP-1559 (August 2021) burns the base fee: about 4.64m ETH burned cumulatively, but the burn collapsed from 8,844 ETH per day in 2021 to about 57 per day in 2026, offsetting only about 2.4% of issuance. ETH has been net inflationary for 28 straight months, last deflationary month March 2024, supply about 122.14m ETH. No issuance-curve cut was enacted in 2024 (see section 0). EIP-8363, the tapered issuance burn, was drafted July 2026 and withdrawn from Hegota consideration on 1 October 2026. There is no protocol treasury; the Ethereum Foundation is funded from its own holdings under a 2.5-year buffer and 15% annual spend cap. Criticisms: the issuance curve never switches off; Lido controls about 24% of staked ETH; the burn is not a reliable substitute for a scheduled issuance end because it disappears precisely when the chain succeeds at scaling. Lesson for OGO: a participation-linked curve elegantly auto-limits per-validator yield, but without a saturation point or a cap it can over-issue indefinitely, and a burn funded by congestion is not a funding source.

### What transfers to a fixed-max chain, and what does not

1. Cardano is the only true fixed-max precedent here, and it has no end date. OGO's hybrid (floor plus fixed end date) must define what the floor is a share of; section 8 settles this as a fixed absolute base for a fixed number of periods.
2. Perpetual terminal inflation (Solana 1.5%, Ethereum's square-root floor about 1.5%) is the incumbents' answer to fee insufficiency, and it is unavailable to a fixed-max chain. Any fixed-max design must state in advance what happens when issuance ends.
3. Treasury funding from issuance shrinks exactly when the chain needs it most. Cardano's 20% of a shrinking pot and Polkadot's 15% of a stepping-down issuance both mean the development budget decays with the security budget.
4. Governance design is monetary design. Solana's issuance is voted on by its recipients under a 66.67% supermajority. Cardano's spending ceiling bound even though the treasury was full. Polkadot's OpenGov changed the money supply twice in 16 months. Ethereum's fork-based rough consensus has not changed its curve since the Merge. For OGO's community-governance commitment, who votes, on what threshold, and whether issuance parameters are changeable at all will determine whether a predetermined schedule stays predetermined.
5. Burns are not a funding source. Ethereum's burn collapsed when scaling succeeded; Solana burns only half the base fee; Polkadot halted its treasury burn in 2026 because it destroyed funds the capped system now needs.
6. Do not import founder allocations from these networks as a template. All four had substantial founder or foundation allocations at genesis (Cardano founders 11.5%; Solana founding team and Foundation about 25% combined in secondary sources; Polkadot Web3 Foundation about 11% or more; Ethereum's pre-mine). OGO's stated zero reserved founder allocation is a deliberate break from all four.

## 10. Unresolved decisions and trade-offs

1. Floor or no floor: is a minimum ongoing issuance floor (hybrid) wanted, or is a smooth taper to zero (geometric) sufficient? This is the decision that separates the two models.
2. Tail length: if geometric, what r? A 0.05 change more than doubles the sustainability horizon.
3. Dust: 45 to 118 uogo per 1,000,000 normalized units. Recommendation: leave it permanently unissued. Needs Benross's confirmation.
4. End rule for geometric: stop when the per-period amount floors to zero, or a fixed final period in genesis? Recommendation: fixed final period, so the end date is known at genesis.
5. Period length: how many blocks (or how much time) is one period? This converts every period count in this report into calendar time.
6. Split: validator/delegator, community, and treasury shares, and whether the split is fixed per period or varies over time.
7. Fee policy: what fee market is expected to replace issuance, and when is it expected to be sufficient? Requires its own analysis (decision order step 4).
8. Hybrid base justification: the base rate is a policy choice about minimum ongoing issuance. If hybrid is chosen, the base level needs its own justification, not just a share of total.
9. Final taper: should the hybrid base itself taper in the last periods to soften the end step, or end at full floor value at period N?
10. Governance of the schedule: who may propose a schedule change, on what threshold, and is the schedule changeable at all outside the separately documented constitutional process? The established networks show this choice determines whether a predetermined schedule stays predetermined.
11. Legal analysis: written legal analysis (securities, money transmission, tax, international, exchange admission) is required before genesis distribution. Not yet commissioned.
12. Open research questions carried from the comparative research: Solana SIMD-0550 activation date (passed, not yet live as of October 2026); Polkadot exact post-cut issuance (55 vs 56.9m DOT per year); Polkadot Dynamic Allocation Pool coretime timing; Cardano Net Change Limit (350m vs 500m ADA); Cardano minimum pool cost (340 vs 170 ADA per epoch).

## Recommendation

Recommendation, not a decision. The decision stays with Benross.

Carry forward geometric decay as the primary candidate and hybrid as the leading alternative, exactly as the handoff directs. The choice turns on two questions only: should OGO maintain a minimum issuance floor, and how long should the issuance tail last?

- If no floor is wanted, choose geometric decay. It is simpler, already implemented in the Milestone B scaffold, and its one parameter (r) is the entire design. The open parameter is r, and section 2 is the guide.
- If a floor is wanted so validators, community programs, and the treasury keep a known minimum through the schedule, choose hybrid. It costs four parameters instead of two and needs a new schedule type in x/ogocoin, but it is the only model that sustains late-schedule funding, and Polkadot's conversion shows a capped schedule plus a treasury redesign must be designed together, not separately.

The next step after the family is chosen is parameter selection for that family, then the split, then fee analysis, then the maximum supply derived as genesis supply plus scheduled issuance. That order is the decision order Benross set.

## Addendum: Tightenings addressed (2026-11-06)

Benross directed three tightenings before this report is used as the basis for a monetary decision. They are addressed in full in `ogo-network-hybrid-terminal-and-treasury-v0.1.md`:

1. Issuance totals: S_target (analytic), S_actual (integer sum of floored per-period amounts), and M are now distinguished, with a recommendation that M = G + S_actual be computed at genesis by exact integer summation.
2. Hybrid terminal behavior: abrupt end compared against linear, short geometric, and gentle geometric tapers. A short linear taper is worse than abrupt; only a gentle geometric taper over the final half of the schedule smooths the end.
3. Treasury sustainability: post-issuance treasury dynamics modelled, including the finding that a fixed spend rate bankrupts a geometric treasury during issuance, and a menu of post-issuance funding mechanisms.

## Sources

Research dated 6 November 2026. Primary documentation: Cardano Docs Monetary Policy; Cardano Docs protocol parameters; Cardano Constitution (Tenet 10); Polkadot Wiki DOT Token; Polkadot Wiki inflation model; Solana Docs fee structure; Solana Forum SIMD-0550; Solana Forum SIMD-0411; EIP-8363; Ethereum.org fork timeline; Electra issuance curve proposal (April 2024); Upgrading Ethereum issuance; Polkadot Forum digest 7 November 2024; Subsquare Referendum 1781. Reporting: BSC News (Cardano fee gap, Polkadot treasury/DAP); CryptoNews; InsideCrypto; Blockworks Cardano disclosures; CoinStats (Cardano NCL, Polkadot fundamentals); Galaxy Research (SIMD-0411); TokenInsight (SIMD-0228 fails); Bitrue; Datawallet; Hubra; CryptoCompass; Solana Compass / Fidelity FCAT; MEXC; BlockchainReporter; Coingabbar; CryptoWiki; CryptoNews (Polkadot treasury); GrowYourEthereum; IOSG EIP-8363 review via TechFlowPost. Full URLs are in the comparative research report in the research notes. OGO-side figures are from the Geometric vs Hybrid Deep-Dive v0.1 and Issuance Mathematics v0.1, computed with exact rational arithmetic.
