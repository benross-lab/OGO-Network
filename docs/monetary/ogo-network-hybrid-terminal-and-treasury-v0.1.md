# OGO Network: Hybrid Terminal Transition and Treasury Sustainability v0.1

Status: Research and design only. No monetary model has been approved. No maximum supply, no percentages, no intervals selected.
Date: 2026-11-06
Purpose: Tighten three items in the Monetary Design Decision Report v0.1 before it is used as the basis for a monetary decision, as Benross directed: (1) issuance totals, (2) hybrid terminal behavior, (3) treasury sustainability after issuance ends. All figures are illustrative, computed with exact rational arithmetic floored to integer uogo, normalized to 1,000,000 illustrative units of total scheduled issuance unless stated otherwise. The normalization is not a proposed supply.

Revision note, 2026-11-06 (GitHub publication): the hybrid rows of the table in section 1 originally read S_actual 999,982, dust 18. Corrected to S_actual 999,978, dust 22. The 18 figure came from solving B0 as an exact rational (about 60,900.158 for base 40%, about 101,500.264 for base 0%) from S_target = 1,000,000; a rational B0 is not an implementable genesis parameter (B0 must be an integer uogo amount), and no other document uses it. The corpus definition, as implemented in the prototype (`x/ogocoin/models.go`, integer B0 = 60,900 for base 40%, B0 = 101,500 for base 0%) and asserted by its tests, gives S_actual 999,978 and dust 22 for both hybrid rows. Recomputed 2026-11-06 with exact rational arithmetic (Python Fractions, floor to uogo); see `docs/discrepancy-resolutions-2026-11-06.md`. The taper-variant totals in section 2 were computed from the pre-correction 999,982 base; under the corrected base they shift by a few units, and the taper conclusions are unaffected. The per-period figures in section 2 (12,581 at period 30, 11,524 at period 35, 11,000 at period 39) are identical under both variants and are unchanged.

## 1. Issuance totals: three different numbers, not one

The Decision Report used a single normalized total. Integer flooring means three distinct numbers must be tracked separately. Confusing them is a real accounting error, not a rounding nicety.

- S_target, the theoretical target: the analytic total from the schedule formula. For geometric decay, S_target = R0 / (1 - r). For hybrid, S_target = base * N + B0 * (1 - r^N) / (1 - r). This number exists only in real-number arithmetic. It is the design intent.
- S_actual, the actual cumulative scheduled issuance: the integer sum of the floored per-period amounts, sum over n of floor(R_n). This is the number of coins the protocol will actually mint if the schedule runs to completion. It is always less than S_target, by the dust.
- M, the maximum supply: M = G + S, where G is genesis circulating supply (TBD). Which S goes into M is a decision, and the two conventions differ.

Computed dust, exact rational arithmetic:

| Model | S_target | S_actual | Dust | Dust as % of target |
|---|---|---|---|---|
| Geometric r = 17/20 | 1,000,000 | 999,955 | 45 | 0.0045% |
| Geometric r = 9/10 | 1,000,000 | 999,935 | 65 | 0.0065% |
| Geometric r = 19/20 | 1,000,000 | 999,882 | 118 | 0.0118% |
| Hybrid base 0% | 1,000,000 | 999,978 | 22 | 0.0022% |
| Hybrid base 40% | 1,000,000 | 999,978 | 22 | 0.0022% |

Dust is tiny in percentage terms and large in absolute principle: 65 uogo on a million-unit normalization scales to 65,000 uogo on a billion-unit supply. The amount does not matter as much as the rule. Two conventions for M, with illustrative G = 500,000 (illustrative only, for the example):

- Convention A, M = G + S_target = 1,500,000. The cap then contains 65 units of dead headroom that can never be issued, because the floored schedule can never mint them. The cap overstates the mintable supply by the dust.
- Convention B, M = G + S_actual = 1,499,935. The cap is exact. Every unit of headroom is mintable by the schedule, and no unit is dead.

Recommendation: adopt Convention B. Genesis validation must compute S_actual by summing the floored schedule with exact integer arithmetic, and set M = G + S_actual. The analytic formula must never be trusted as the total. This makes all three numbers consistent and independently verifiable: anyone with the genesis file can recompute S_target from the formula, S_actual by summing the floored schedule, and M from G + S_actual, and confirm they agree. The dust stays permanently unissued, by rule, and must never become an alternative minting allowance.

## 2. Hybrid terminal behavior: taper versus abrupt end

The hybrid model as defined in the deep-dive ends abruptly: the base floor runs at full value through period 39, then issuance is zero from period 40. The final payment is 11,000 illustrative units (base 10,000 plus a bonus of about 1,000), and it disappears in one step. The question Benross asked is whether a final taper is better than that abrupt end.

The honest answer first: no taper removes the end. Every finite schedule ends with issuance going to zero, so the final drop is 100% in every variant. What a taper changes is the SIZE of the final payment that disappears, and the size of the per-period drops leading up to it. Those two things pull in opposite directions, and a naive taper makes one of them worse.

All variants below use hybrid base 40%, N = 40 periods, r = 9/10, normalized to the same total where stated.

### Abrupt end (baseline)

Total 999,978. Period 30 issuance 12,581; period 35 issuance 11,524; period 39 (final payment) 11,000; period 40 zero. Largest single-period drop inside the last 10 periods: 2.1%. Then a 100% drop of 11,000 units at the end. Smooth until the end, then a large final cliff.

### Short linear taper (last 10 periods, base declining linearly to zero)

Unrenormalized total 954,982, a loss of 45,000 units against the abrupt version, because the taper removes base issuance without replacing it. Final payment 2,000 instead of 11,000. But the largest single-period drop inside the last 10 periods is 35.7%. The taper steps are fixed chunks of 1,000 units taken from an already small total, so each step is a large percentage cut. The final cliff is smaller, but the taper itself is rougher than the abrupt version it replaces. Do not use a short linear taper.

### Short geometric base taper (last 10 periods, base multiplied by 0.8 each period)

Unrenormalized total 944,611, a loss of 55,371. Final payment 2,342. Largest single-period drop inside the last 10 periods: 17.9%. Better than linear, still much rougher than the abrupt version's 2.1%.

### Gentle geometric base taper (last 20 periods, base multiplied by 0.95 each period)

Unrenormalized total 928,277, a loss of 71,705. Final payment 4,773. Largest single-period drop inside the last 20 periods: 7.1%; inside the last 10: 6.5%. Renormalized to the same total (999,960), the initial issuance rises to 76,376 (from 70,900) and the final payment is 5,141, with the same 6.5% maximum drop. This is the only taper variant tested that actually smooths: it roughly halves the final payment (11,000 to about 5,000) while keeping every per-period drop under 7%.

### Comparison

| Variant | Total (unrenormalized) | Final payment | Largest drop in taper window | 90% issued by |
|---|---|---|---|---|
| Abrupt end | 999,978 | 11,000 | 2.1% in-schedule, then 100% of 11,000 | period 32 |
| Linear taper, last 10 | 954,982 | 2,000 | 35.7% | period 32 |
| Geometric taper 0.8, last 10 | 944,611 | 2,342 | 17.9% | period 32 |
| Gentle geometric taper 0.95, last 20 | 928,277 | 4,773 | 7.1% | period 28 (renormalized) |

Reading: a taper is not automatically gentler. A short taper concentrates the decline into a few large steps. A taper only smooths if it is geometric, gentle (factor close to 1), and long (about half the schedule). And a taper is not free: unrenormalized it reduces total issuance by 45,000 to 95,000 units on a million-unit normalization; renormalized it raises early issuance instead (R0 from 70,900 to 76,376 in the gentle case). Either the schedule issues less in total, or it issues more early. There is no taper that only removes the cliff.

Recommendation: if hybrid is chosen, use a gentle geometric base taper (factor about 0.95) over the final half of the schedule, renormalized to the same total, with the taper parameters fixed in genesis. Do not use a short linear taper. And state the end plainly in the design: the taper shrinks the final cliff from about 11,000 to about 5,000 illustrative units; it does not eliminate the end. The end is still a date when issuance stops, and validators and the treasury must be prepared for it.

## 3. Treasury sustainability after issuance ends

A treasury funded exclusively through issuance eventually stops receiving new funds. That is arithmetic, not a forecast. This section models what happens then, under clearly labelled illustrative assumptions. Fee demand is unknown, so every post-issuance figure here is a scenario, not a prediction.

### The model

Treasury balance evolves per period as balance = previous balance + inflow - spend. Inflow during issuance is 10% of R_n under the illustrative 70/20/10 split. Spend is an illustrative fixed amount per period, shown at 1,500 and at 750 units, to show sensitivity. Spend is illustrative only; no spend rate is proposed.

### What happens with no post-issuance income

| Model | Spend 1,500 per period | Spend 750 per period |
|---|---|---|
| Geometric r = 0.90 | Balance goes NEGATIVE at period 66, during issuance. Total inflow 99,993 cannot sustain 1,500 per period over a 110-period tail. | Balance at end of issuance 17,447. Survives 23 further periods with zero inflow. |
| Hybrid base 40% | Balance at end of issuance 39,984. Survives 26 further periods with zero inflow. | Balance at end of issuance 69,984. Survives 93 further periods with zero inflow. |

Two findings. First, a fixed spend rate that the early treasury can easily afford becomes unaffordable in geometric decay's long thin tail: the treasury goes broke at period 66 while issuance is still running. A treasury cannot spend at a fixed rate against a decaying inflow without a spending rule that follows the inflow. Second, the hybrid floor is what keeps the treasury solvent: the same 1,500 spend that bankrupts the geometric treasury leaves the hybrid treasury with 39,984 at the end of issuance. The floor protects the treasury exactly as it protects validators.

### Fee break-even

To sustain a 1,500 per period treasury spend after issuance ends, fee revenue to the treasury must be 1,500 per period. That is 1.5% of geometric initial issuance (R0 100,000) or 2.1% of hybrid initial issuance (R0 70,900). As a growth illustration only: fee revenue to the treasury starting at 500 per period and growing 5% per period would cover a 1,500 spend at period 23. That growth rate is an assumption, not a forecast. The honest statement is that post-issuance sustainability cannot be proven from issuance mathematics alone; it requires fee-demand assumptions that must be stated as assumptions.

For scale, the established networks show how far fees are from covering security today. Cardano's fees covered about 0.67% of staking rewards over the year to 1 September 2026. Ethereum's fee burn offset about 2.4% of issuance in 2026. Neither is close to replacing issuance, and OGO should not assume its fee market will behave differently without evidence.

### Funding mechanisms after issuance ends (to investigate, not assumed)

1. A fixed, explicitly disclosed share of transaction fees to the treasury. This is the most direct mechanism. It requires a fee policy decision and a realistic estimate of transaction demand. It is the only mechanism in this list that is already implied by the treasury's design.
2. Blockspace or coretime sales, in the Polkadot style: selling bulk blockspace and routing the revenue to the treasury. Polkadot's experience is instructive: coretime revenue was initially burned, then routed to its Dynamic Allocation Pool, and demand remains below the level where it offsets issuance. Investigate, do not assume revenue.
3. An endowment model: the treasury holds a principal balance and spends only its yield. This requires the principal to be large enough that its yield covers spending, which is a stricter condition than the balance arithmetic above. It also requires a yield source, which brings the question back to staking or fees.
4. Treasury staking of its own holdings: the treasury stakes OGO it holds and earns validator rewards. This is circular while issuance runs, because the rewards come from issuance, and it cannot fund the treasury after issuance ends, because there is no issuance to earn from. It also concentrates governance influence in the treasury's hands. Note the circularity and treat this as a weak mechanism, not a solution.
5. Community contributions and grants from outside the protocol: real, but not protocol-enforced and not predictable. They cannot be counted in a sustainability model.

Recommendation: the treasury's post-issuance plan must name at least one mechanism from this list, with its funding share disclosed before genesis, and must include a spending rule that follows inflow (for example, spending capped as a share of the treasury balance or of trailing inflow) rather than a fixed rate. A fixed spend rate against a decaying inflow is the failure mode section 3 demonstrates. Polkadot's Net Change Limit and Cardano's spending ceiling are the closest working precedents for a spending rule: a community-set ceiling independent of the balance.

## 4. What this changes in the Decision Report

The three tightenings are addressed here, and the Decision Report should be read with this report alongside it.

1. Issuance totals: section 1 distinguishes S_target, S_actual, and M, and recommends M = G + S_actual computed at genesis by exact integer summation. The Decision Report's normalized totals should be read as S_target values; the S_actual values are in section 1.
2. Hybrid terminal behavior: section 2 compares abrupt end against three taper variants. The Decision Report's statement that hybrid ends cleanly at period N is now qualified: the end is a cliff of the final payment's size, and only a gentle geometric taper over the final half of the schedule smooths it. A short linear taper is worse than abrupt.
3. Treasury sustainability: section 3 models post-issuance treasury dynamics and finds that a fixed spend rate bankrupts a geometric treasury during issuance, while a hybrid floor keeps it solvent. The Decision Report's treasury projections cover funding during issuance only; this report covers what happens after.

## 5. Recommendation

Recommendation, not a decision. The decision stays with Benross.

1. Adopt the three-number accounting rule (S_target, S_actual, M = G + S_actual) for whichever model family is chosen. This is model-independent and should be adopted regardless.
2. If hybrid is chosen, adopt the gentle geometric base taper (factor about 0.95) over the final half of the schedule, renormalized to the same total. Do not adopt a short linear taper.
3. Adopt a treasury spending rule that follows inflow, and name at least one post-issuance funding mechanism with its share disclosed before genesis. Do not approve a monetary design whose treasury has no post-issuance plan.
4. Geometric and hybrid remain the two candidates. Nothing in this report selects a model or a parameter.

## 6. Unresolved questions

1. Dust convention: Convention B (M = G + S_actual) is recommended; needs Benross's confirmation.
2. Taper parameters: taper factor, taper window length, and whether the taper is renormalized to the same total or allowed to reduce total issuance.
3. Treasury spend rate and spending rule: the spend rates in section 3 are illustrative. A real spending rule (cap as a share of balance or trailing inflow) needs its own design, with Cardano's Net Change Limit and Polkadot's Dynamic Allocation Pool as references.
4. Post-issuance funding mechanism: which mechanism or combination from section 3, and what share? Fee demand estimates are needed before any of them can be relied on.
5. Period length: how many blocks or how much time is one period? This converts every period count in this report into calendar time.
6. Split: validator/delegator, community, and treasury shares, and whether the split is fixed per period or varies over time.
7. Fee policy: what fee market is expected to replace issuance, and when is it expected to be sufficient? Requires its own analysis.
8. Legal analysis: written legal analysis is required before genesis distribution. Not yet commissioned.
