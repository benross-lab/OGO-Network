# OGO Network: Treasury Sustainability v0.2

Status: Research and mathematical modeling only. No monetary model has been approved. No maximum supply, no percentages, no intervals selected.
Date: 2026-11-06
Builds on: Monetary Design Decision Report v0.1, Geometric vs Hybrid Deep-Dive v0.1, Hybrid Terminal Transition and Treasury Sustainability v0.1.

Purpose: Model what happens to OGO over 10, 20, and 30 years, including what happens when issuance reaches zero. Four investigations: (1) treasury spending policy, (2) post-issuance revenue, (3) network security after issuance ends, (4) hybrid versus geometric across growth and low-adoption scenarios.

Every figure in this report is illustrative, computed with exact rational arithmetic (Python Fractions, floored to integer uogo) for issuance, normalized to 1,000,000 illustrative units of total scheduled issuance unless stated otherwise. The normalization is not a proposed supply. The two models carried forward are geometric decay r = 9/10, R0 = 100,000, and hybrid base 40% (10,000 per period), N = 40 periods, r = 9/10, R0 = 70,900, under the illustrative 70/20/10 validator, community, treasury split. All of these are hypothetical inputs chosen to show sensitivity. Nothing here selects a parameter.

## 1. The calendar map: periods are not years

Period length is TBD, and it is not a cosmetic conversion. It determines whether the issuance schedule covers the 10, 20, and 30-year horizons at all. This report shows both illustrative mappings and states which one each figure uses.

| Horizon | Monthly periods | Quarterly periods |
|---|---|---|
| 10 years | 120 periods | 40 periods |
| 20 years | 240 periods | 80 periods |
| 30 years | 360 periods | 120 periods |

| Model | Issuance ends | Monthly | Quarterly |
|---|---|---|---|
| Geometric r = 0.90 | period 110 | 9.2 years | 27.5 years |
| Hybrid base 40% | period 40 | 3.3 years | 10.0 years |

Reading: under monthly periods, both models finish issuing before year 10, so the 10, 20, and 30-year horizons are entirely post-issuance. Under quarterly periods, hybrid ends exactly at year 10 and geometric ends at year 27.5, so the 10-year horizon is still partly issuance-funded under geometric. The period-length decision IS a sustainability decision: a shorter period compresses the whole issuance schedule into fewer calendar years and brings the fee transition forward. Any schedule parameter chosen without a period length is meaningless in calendar terms.

A second interaction matters for every growth figure in this report. A growth rate stated per period means something very different per year depending on period length:

| Growth per period | Annual equivalent, monthly periods | Annual equivalent, quarterly periods |
|---|---|---|
| 1% | 12.7% per year | 4.1% per year |
| 2% | 26.8% per year | 8.2% per year |
| 5% | 79.6% per year | 21.6% per year |
| 10% | 213.8% per year | 46.4% per year |

Every growth assumption in sections 3 and 4 is stated per period, with its annual equivalent given under both mappings. A scenario that looks modest per period can be aggressive per year under monthly periods.

## 2. Treasury spending policy

### The setup

Treasury balance evolves per period as balance = previous balance + inflow - spend. Inflow during issuance is 10% of R_n under the illustrative split: total treasury inflow is 99,993 illustrative units under geometric and 99,998 under hybrid. Desired spend is an illustrative 1,500 units per period, carried from v0.1 so the policies can be compared on the same need. Desired spend is illustrative only; no spend rate is proposed.

Five policies, simulated over 360 periods (30 years under monthly periods):

- P0, fixed spend: spend 1,500 every period, regardless of inflow or balance. The baseline. Known failure mode from v0.1.
- P1, trailing inflow cap: spend = min(desired, average treasury inflow over the trailing 4 periods). Spend never exceeds recent inflow.
- P2, balance cap: spend = min(desired, 5% of the current balance). Endowment-style. Spend scales with the balance.
- P3, reserve floor: spend = min(desired, balance above a 20,000-unit reserve floor). The treasury never spends below the floor.
- P4, combined: spend = min(desired, trailing 4-period average inflow, 5% of balance, balance above the 20,000 reserve floor). All three constraints at once.

### Results, issuance-only inflow

| Policy | Geometric: balance at issuance end (t=109) | Geometric: balance at t=360 | Geometric: total spent | Hybrid: balance at issuance end (t=39) | Hybrid: balance at t=360 | Hybrid: total spent |
|---|---|---|---|---|---|---|
| P0 fixed | -65,053 (went negative at t=66) | -440,053 | 540,000 | 39,984 at t=39, then negative at t=66 | -440,016 | 540,000 |
| P1 trailing inflow | 55,682 | 55,682 | 44,265 | 41,757 | 41,757 | 58,227 |
| P2 balance 5% | 1,328 | 19 | 99,928 | 44,554 | 19 | 99,965 |
| P3 reserve 20,000 | 20,000 | 20,000 | 79,947 | 20,000 | 20,000 | 79,984 |
| P4 combined | 60,327 | 60,327 | 39,620 | 48,076 | 48,076 | 51,908 |

Shortfall periods (spend below the desired 1,500) out of 360: P1 geometric 340, hybrid 334; P2 geometric 315, hybrid 316; P3 geometric 307, hybrid 307; P4 geometric 344, hybrid 340. Every policy that keeps the treasury solvent spends below the desired level in most periods. That is the price of solvency against a decaying inflow.

### What each policy does, and its failure mode

P0 fixed spend fails in both models, at the same period number for different reasons. Geometric goes negative at period 66 during issuance: the long thin tail cannot sustain a fixed spend. Hybrid ends issuance with 39,984 at period 39, survives 26 further periods, then goes negative at period 66. The fixed rate is not a policy; it is a promise the inflow cannot keep.

P1 trailing inflow never goes negative, but it guarantees solvency by guaranteeing irrelevance after issuance ends. Spend follows inflow down and reaches zero within 4 periods of the end of issuance (geometric spend at t=39 is 193, at t=44 is 113, at t=110 is 0; hybrid spend at t=39 is 1,117, at t=40 is 833, at t=44 is 0). The balance then freezes forever: geometric 55,682, hybrid 41,757, never spent again. A treasury that can never spend after issuance ends is solvent and useless. The trailing cap is a good guardrail inside the schedule and a bad sole rule after it.

P2 balance cap also never goes negative, and it is the only policy that keeps spending after issuance ends, at declining levels. But it has two flaws. First, it starves the treasury at genesis: spend is 0 at period 0 because the balance is 0, and only reaches the desired 1,500 around period 20, exactly when the network may need funding most. Second, under geometric it spends the treasury down during issuance anyway: balance 1,328 at issuance end, decaying to 19 by period 240, because a 5% of balance rule spends the full desired 1,500 whenever the balance is large, and the balance is largest early when inflow is front-loaded. The balance cap prevents debt; it does not preserve a balance.

P3 reserve floor preserves exactly 20,000 forever and spends 79,947 to 79,984 in total. Its failure mode is the mirror of P2: once the balance reaches the floor, spending stops entirely and permanently. The floor is preserved, but the treasury becomes a locked vault holding 20,000 units it will never use. A reserve floor is a statement about what must never be spent; it is not a spending policy on its own.

P4 combined is the most conservative: highest end balance in both models (geometric 60,327, hybrid 48,076), lowest total spent (39,620 and 51,908), most shortfall periods. It inherits P1's post-issuance freeze: spend reaches zero after issuance ends and the balance never moves again.

### The level problem no policy can fix

Spending policy chooses HOW a shortfall is distributed. It cannot fix a shortfall that is built into the level of desired spend. The arithmetic:

- Maximum sustainable constant spend over the schedule: total inflow divided by schedule length. Geometric: 99,993 / 110 = 909 per period. Hybrid: 99,998 / 40 = 2,500 per period.
- The illustrative desired spend of 1,500 per period exceeds geometric's maximum by 65%. No spending policy can make geometric sustain 1,500 per period through its schedule; the money does not exist.
- Over a 360-period horizon, desired spend totals 540,000 units against total treasury inflow of about 100,000. The 30-year need is 5.4 times the entire issuance-funded treasury. Post-issuance revenue is not optional for any long horizon; it is required by arithmetic.

The policy recommendation follows in section 6. The level question, what spend rate is actually affordable, must be decided separately from the policy question, and for geometric the answer starts at 909 per period, not 1,500.

## 3. Post-issuance revenue

### What has to be true

After issuance ends, every unit of treasury spending and validator pay must come from somewhere else. This section states the break-even requirements for each candidate source. It does not forecast fee demand. Fee demand is unknown, and every figure here is a requirement, not a prediction.

### Transaction fees: the break-even requirement

Define the minimum post-issuance need as the sum of three illustrative requirements: minimum validator operating budget, delegator yield, and treasury spend.

Validator operating budget: illustrative assumptions, stated explicitly. Validator count N_v in {50, 100, 200}. Operating cost per validator per period c in {50, 100, 200} illustrative units. These cost figures are arbitrary placeholders chosen to show sensitivity; no validator cost study exists yet. Minimum total validator budget = N_v * c.

| Validators x cost per period | Total per period | As % of geometric R0 (100,000) | As % of hybrid R0 (70,900) |
|---|---|---|---|
| 50 x 50 | 2,500 | 2.5% | 3.5% |
| 50 x 100 | 5,000 | 5.0% | 7.1% |
| 50 x 200 | 10,000 | 10.0% | 14.1% |
| 100 x 100 | 10,000 | 10.0% | 14.1% |
| 100 x 200 | 20,000 | 20.0% | 28.2% |
| 200 x 100 | 20,000 | 20.0% | 28.2% |
| 200 x 200 | 40,000 | 40.0% | 56.4% |

Delegator yield: illustrative assumptions. Genesis supply G = 500,000 (the v0.1 example value, illustrative only). Maximum supply M = G + S_actual: geometric 1,499,935, hybrid 1,499,982. Stake fraction in {30%, 50%} of M. Required yield in {3%, 5%} per year. Yield is an opportunity-cost assumption, not a measured market rate.

| Stake fraction, yield | Per year | Per period, monthly | Per period, quarterly |
|---|---|---|---|
| 30%, 3% | 13,499 | 1,125 (1.1% of geometric R0, 1.6% of hybrid R0) | 3,375 (3.4% / 4.8%) |
| 30%, 5% | 22,499 | 1,875 (1.9% / 2.6%) | 5,625 (5.6% / 7.9%) |
| 50%, 3% | 22,499 | 1,875 (1.9% / 2.6%) | 5,625 (5.6% / 7.9%) |
| 50%, 5% | 37,498 | 3,125 (3.1% / 4.4%) | 9,375 (9.4% / 13.2%) |

Total required fee revenue per period, monthly periods, at 100 validators x 100 cost, 50% stake at 5% yield, treasury spend 1,500:

| Validator count | Total required fee revenue per period | As % of geometric R0 | As % of hybrid R0 |
|---|---|---|---|
| 50 | 9,625 | 9.6% | 13.6% |
| 100 | 14,625 | 14.6% | 20.6% |
| 200 | 24,625 | 24.6% | 34.7% |

Two readings. First, the required fee revenue is a substantial fraction of period-0 issuance: 10 to 35% depending on validator count and model. Fees must replace a tenth to a third of the network's initial issuance rate, every period, forever. Second, hybrid's requirement is larger as a share of its own R0 because its R0 is smaller (70,900 against 100,000); the absolute requirement is the same in both models, since validator costs and delegator yield do not depend on the issuance curve. The model choice does not change what fees must provide. It changes when fees must provide it.

### The evidence anchor: how far fees are today

Cardano collected about 3.3m ADA in transaction fees against 493.7m ADA in staking rewards over the 73 epochs to 1 September 2026: fees covered about 0.67% of rewards. Ethereum's EIP-1559 burn offset about 2.4% of issuance in 2026. Neither network is close to replacing issuance, and both have had years of live fee markets.

Applied to OGO's illustrative numbers, a Cardano-like fee level would produce fee revenue of 469 per period under geometric (0.67% of period-0 validator pay of 70,000) and 333 per period under hybrid (0.67% of 49,630). Against the minimum need of 14,625 per period (100 validators), that covers 3.2% under geometric and 2.3% under hybrid.

How long would fee growth take to close the gap, at constant per-period growth rates (assumptions, not forecasts)?

| Fee growth per period | Geometric: periods to reach 14,625 | Monthly years | Quarterly years |
|---|---|---|---|
| 2% | 174 | 14.5 | 43.4 |
| 5% | 71 | 5.9 | 17.6 |
| 10% | 36 | 3.0 | 9.0 |

Hybrid needs 78 periods at 5% growth (6.5 years monthly, 19.4 years quarterly). The growth rate that closes the gap is not a small one, and under quarterly periods even 5% per period growth takes nearly 18 years to reach the minimum need. The fee transition cannot be assumed; it must be given a horizon at least this long, or the minimum need must be smaller.

### Transaction volume implied

Fee revenue = average fee per transaction x transactions per period. For the minimum need of 14,625 per period:

| Average fee per transaction | Transactions per period required |
|---|---|
| 0.01 units | 1,462,500 |
| 0.1 units | 146,250 |
| 1 unit | 14,625 |
| 10 units | 1,462 |

Equivalently, as a share of transacted value: at a 0.1% fee rate, 14,625,000 units of value must move per period; at 1%, 1,462,500 units; at 0.01%, 146,250,000 units. These are break-even volumes, not usage forecasts. They show the trade-off plainly: low fees need high volume, high fees need less volume but suppress usage. The fee policy decision must pick a point on this trade-off with evidence, not by assumption.

### Infrastructure revenue

Blockspace or coretime sales, in the Polkadot style: selling bulk blockspace and routing the revenue to the treasury or to validators. What has to be true: sustained demand for OGO blockspace at prices that produce the required revenue per period. The evidence is discouraging as a primary source: Polkadot's coretime revenue was initially burned, then routed to its Dynamic Allocation Pool, and demand remains below the level where it offsets issuance. Coretime is a real mechanism and a weak revenue source so far. It should be investigated as a supplement, not counted as the plan.

Hosted developer infrastructure, premium tools, and AI services are OGO Systems revenue, not protocol revenue, unless a defined share is routed on-chain to the treasury. That routing is a design decision with its own governance and accounting questions, and it must be disclosed before genesis if it is counted in the sustainability model. It cannot be assumed silently.

### Endowment

The treasury holds a principal balance and spends only its yield. What has to be true: the balance must be large enough that a plausible yield covers spending. At an illustrative yield of 0.5% per period (6.2% per year under monthly periods), sustaining 1,500 per period requires a balance of 300,000 units: three times the entire issuance-funded treasury inflow of about 100,000. The actual end-of-issuance balances under the section 2 policies are 41,757 to 60,327 units, which would need yields of 2.5% to 3.6% per period (35% to 54% per year under monthly periods) to sustain 1,500 per period. No plausible yield does that. The endowment model fails on arithmetic unless the treasury is funded far beyond its issuance share, which means the split, not the yield, is doing the work. Treasury staking of its own holdings is circular and does not help: it earns rewards from issuance while issuance runs, and there is no issuance to earn from after issuance ends, as v0.1 already noted.

### Other sources, stated honestly

Community contributions and grants from outside the protocol are real but not protocol-enforced and not predictable. They cannot be counted in a sustainability model. A protocol-level fee share to the treasury, disclosed before genesis, is the only mechanism in this list that is already implied by the treasury's design. Everything else is a supplement.

## 4. Network security after issuance ends

### Required fee revenue, restated as security

The section 3 requirement is also the security budget. Once issuance is zero, fee revenue is the entire security budget. The required fee revenue per period, relative to period-0 issuance, at the minimum validator budget (100 validators x 100 cost), 50% stake at 5% yield, treasury 1,500:

- Geometric: 14,625 per period = 14.6% of R0 (monthly periods). Under quarterly periods the per-period requirement is 9,375 delegator yield + 10,000 validators + 1,500 treasury = 20,875, or 20.9% of R0, because the same annual yield must be paid in fewer, longer periods.
- Hybrid: 14,625 per period = 20.6% of R0 (monthly). Quarterly: 20,875, or 29.4% of R0.

To hold validator pay at its period-0 level (not the minimum), fee revenue would need to be 70,000 per period under geometric (70% of R0) or 49,630 under hybrid (70% of hybrid R0). That is the full-replacement figure from the Decision Report. The minimum-budget figure is the survival figure. The gap between them, 70,000 against 14,625, is the difference between keeping validators as well paid as at launch and keeping them paid enough to stay.

### How many validators can fees sustain

At a fee revenue level F per period, with an illustrative 80% of fees to validators and 20% to the treasury, and illustrative validator cost of 100 per period:

| Fee revenue per period | Validator share (80%) | Validators sustained at 100 per period |
|---|---|---|
| 300 | 240 | 2 |
| 1,000 | 800 | 8 |
| 5,000 | 4,000 | 40 |
| 10,000 | 8,000 | 80 |
| 25,000 | 20,000 | 200 |

This is the security failure mode in its starkest form. At the Cardano-like fee level of about 469 per period, the network sustains 3 validators at the illustrative cost. At 1,000 per period, 8 validators. A validator set that small is effectively centralized, whatever the nominal design. Delegator yield is a secondary question in this regime: if fee revenue cannot cover validator operating costs, validators exit before delegators face a yield decision. The exit cascade is: fee revenue below validator costs, marginal validators exit, stake concentrates on the survivors, the Nakamoto coefficient falls, and the chain's security degrades while its nominal issuance schedule is still being followed perfectly. The schedule can be exactly right and the network insecure at the same time.

### What varies the requirement

Validator count is the largest lever: doubling validators from 100 to 200 doubles the minimum budget from 10,000 to 20,000 per period, and raises the total required fee revenue from 14.6% to 24.6% of geometric R0. Stake fraction and required yield are the second lever: 50% stake at 5% yield needs 3,125 per period (monthly); 30% stake at 3% needs 1,125. Period length is the third: quarterly periods concentrate the same annual yield into fewer payments, raising the per-period requirement by about 43% against the monthly figure. The model family is not a lever on the requirement at all: the absolute fee revenue needed is identical under geometric and hybrid. What differs is the date by which it is needed, which is section 5.

## 5. Hybrid versus geometric across scenarios

### Scenario definitions

Two scenarios, defined in the model (transaction volume, fee revenue, validator participation), not by adjectives.

Growth scenario (illustrative): fee revenue starts at 500 per period and grows 5% per period (79.6% per year under monthly periods, 21.6% under quarterly), saturating at a ceiling of 25% of R0 per period (25,000 under geometric, 17,725 under hybrid). The ceiling is an explicit saturation assumption: fee revenue cannot compound forever, and any growth scenario without a ceiling is meaningless over 30 years. Fee split: 80% to validators, 20% to the treasury (illustrative). Validator count grows from 50 toward 100. Treasury uses the balance-cap policy (5% of balance) plus fee inflow.

Low-adoption scenario (illustrative): fee revenue starts at 100 per period and grows 1% per period (12.7% per year under monthly periods, 4.1% under quarterly), saturating at a ceiling of 5% of R0 per period (5,000 under geometric, 3,545 under hybrid). Validator count stays at 50. Same fee split and treasury policy.

### Growth scenario results

| Model | Fee revenue reaches minimum need (11,500) | Fee revenue reaches period-0 validator pay | Funding gap at issuance end | Treasury balance at t=359 (30y monthly) |
|---|---|---|---|---|
| Geometric | period 65 (5.4y monthly, 16.2y quarterly) | never (ceiling 25,000 < 70,000) | surplus of 13,500 at t=109 | 1,060,250 |
| Hybrid | period 65 (same) | never (ceiling 17,725 < 49,630) | shortfall of 8,148 at t=39 | 650,280 |

The crossover period is identical in both models because fee growth is the same and the minimum need is the same. What differs is the issuance end date, and that difference is decisive. Geometric issuance ends at period 109, by which time fee revenue has already reached its 25,000 ceiling and validator pay is 20,000 per period from fees alone: no funding valley after issuance ends. Hybrid issuance ends at period 39, when fee revenue is only 3,352 per period: a funding valley opens immediately and lasts 26 periods, from period 40 to period 65, with a total shortfall of 116,065 units against the minimum validator budget and a worst single-period shortfall of 7,184. The hybrid treasury holds 56,541 at period 39, which covers about half the valley. The remainder must come from validator exit, faster fee growth, or a reduced minimum budget. The hybrid floor protects validators during the schedule and then hands them a valley at a known date; whether the valley is survivable depends on the treasury balance and the fee growth rate, not on the floor.

Neither model reaches period-0 validator pay from fees under the growth ceiling. Full replacement of launch-level pay requires fee revenue of 70,000 or 49,630 per period, against a ceiling of 25,000 or 17,725. The growth scenario sustains the minimum, not the launch level. That distinction must be carried into any claim about the fee transition succeeding.

### Low-adoption scenario results

| Model | Fee revenue reaches minimum need | Funding gap at issuance end | Validator pay at issuance end | Validator pay at t=359 |
|---|---|---|---|---|
| Geometric | never | shortfall of 11,204 at t=109 | 236 per period (2 validators) | 2,847 per period |
| Hybrid | never | shortfall of 11,353 at t=39 | 7,816 per period (issuance still paying) | 2,836 per period |

Under low adoption, neither model recovers within the 30-year horizon. Fee revenue at t=359 is 3,559 per period, still below the 11,500 minimum need. Treasury balance under the balance policy falls to about 2,345 at issuance end under geometric and 2,390 under hybrid, and validator pay at issuance end under geometric is 236 per period, sustaining 2 validators at the illustrative cost. The low-adoption scenario is a security failure under both models, and it is a failure of fee demand, not of the issuance curve. No issuance schedule can compensate for a fee market that never develops. The honest conclusion is that low adoption is an existential scenario for a fixed-max chain, and the mitigation is not a different curve but a different plan: a longer schedule, a lower minimum validator set, or an explicit decision that the chain accepts a smaller validator set.

### Scenario summary

| Dimension | Geometric | Hybrid |
|---|---|---|
| Issuance-funded years (monthly / quarterly) | 9.2 / 27.5 | 3.3 / 10.0 |
| Time for fee market to develop before the cliff | 110 periods | 40 periods |
| Funding valley in growth scenario | none | 26 periods, 116,065 total shortfall |
| Funding valley in low-adoption scenario | permanent within 30y | permanent within 30y |
| Treasury at issuance end, growth scenario | 185,250 | 56,541 |
| Post-issuance fee requirement | identical to hybrid in absolute terms | identical to geometric in absolute terms |

The comparison turns on one trade. Hybrid buys a floor during the schedule and pays for it with an early, concentrated transition: the fee market gets 40 periods to develop, then faces a 26-period valley in the growth scenario. Geometric buys time and pays for it with a declining tail: the fee market gets 110 periods, and in the growth scenario there is no valley at all. Under low adoption both fail. The floor is insurance during the schedule; the tail is insurance at the end. They are different kinds of insurance, and the choice between them is a choice about which risk is worse: underfunding during the schedule, or a concentrated transition at its end.

A note on the taper variant from v0.1: the gentle geometric taper (factor 0.95 over the final 20 periods, renormalized) halves the hybrid final payment from 11,000 to about 5,000 and smooths the last periods, but it does not change any post-issuance figure in this report. The valley is caused by the end date, not by the size of the final payment.

## 6. Recommendation

Recommendation, not a decision. The decision stays with Benross.

1. Adopt the combined spending policy (P4: trailing inflow cap, balance cap, reserve floor) as the treasury's default rule, with the parameters (trailing window, balance percentage, reserve floor) set as genesis parameters and disclosed before genesis. It is the only policy tested that never goes negative and preserves the largest balance. But adopt it with the understanding in section 2: after issuance ends, the trailing-inflow constraint freezes spending. The combined policy needs a post-issuance clause, either a fee-revenue inflow that replaces issuance inflow in the trailing average, or a separate post-issuance spending rule based on the balance cap alone. A policy that freezes the treasury forever after issuance ends is not a sustainability policy.
2. Separate the level question from the policy question. The affordable constant spend rate is 909 per period under geometric and 2,500 under hybrid, against total treasury inflow of about 100,000. Any desired spend above those rates must be justified by post-issuance revenue, not by the issuance schedule. Do not approve a spend rate the schedule cannot fund.
3. Name the post-issuance funding mechanism and its share before genesis, as v0.1 already recommended. The mechanism with the strongest claim is a disclosed share of transaction fees to the treasury. Coretime and blockspace sales are supplements, not the plan. Endowment yield cannot sustain the illustrative spend at any plausible yield. Treasury staking is circular.
4. Set the minimum validator set explicitly, and state what happens when fee revenue cannot sustain it. The validator-count assumption is the largest lever on the fee requirement, and an unstated minimum validator set is an unstated security assumption. The low-adoption scenario shows what happens without one: the validator set collapses to 2, and the chain is effectively centralized.
5. Carry forward both models, with the scenario evidence now attached. Geometric's case is the long tail: it gives the fee market 110 periods and avoids a funding valley in the growth scenario. Hybrid's case is the floor: it sustains validators, community programs, and the treasury at a known minimum through the schedule, at the cost of a 26-period funding valley in the growth scenario that the treasury can only half-bridge. The choice is between insurance during the schedule and insurance at its end.

## 7. Assumptions

1. Total scheduled issuance normalized to 1,000,000 illustrative units in all simulations. Genesis supply G = 500,000 illustrative, carried from the v0.1 example.
2. Geometric r = 9/10, R0 = 100,000. Hybrid base 40% (10,000 per period), N = 40, r = 9/10, R0 = 70,900. Illustrative split 70/20/10.
3. Periods are uniform in length. Period length in blocks or time is TBD. Calendar mappings use illustrative monthly and quarterly assumptions.
4. Desired treasury spend 1,500 per period, reserve floor 20,000, balance cap 5% per period, trailing window 4 periods. All illustrative.
5. Validator operating cost per validator per period in {50, 100, 200} illustrative units. No validator cost study exists yet.
6. Stake fraction in {30%, 50%} of maximum supply. Required delegator yield in {3%, 5%} per year. Both are assumptions, not measured rates.
7. Fee split 80% validators, 20% treasury, illustrative. Fee growth rates and ceilings in section 5 are assumptions, not forecasts. Fee revenue growth is saturating, not exponential without limit.
8. All issuance arithmetic is exact rational, floored once to integer uogo. Fee and treasury simulations use integer arithmetic where possible; growth figures use float and are labelled approximate.
9. Cardano and Ethereum figures are point-in-time snapshots from research dated 6 November 2026, as recorded in the Decision Report.

## 8. Unresolved questions

1. Period length: how many blocks or how much time is one period? This converts every period count into calendar time and changes the annual equivalent of every per-period growth rate.
2. Treasury spend level: what spend rate is actually needed, and is it affordable under the chosen model? The affordable constant rates are 909 per period (geometric) and 2,500 per period (hybrid) against issuance inflow alone.
3. Post-issuance spending rule: the combined policy freezes spending after issuance ends because the trailing-inflow constraint goes to zero. What rule governs spending once issuance inflow is replaced by fee inflow?
4. Fee policy: what fee market is expected, at what average fee and volume, and when is it expected to be sufficient? Requires its own analysis with evidence, not assumption.
5. Minimum validator set: how many validators must the network sustain, at what cost, and what happens when fee revenue cannot sustain them? Includes the bridge question: can the treasury subsidize validator pay during a transition valley, and under what rules?
6. Validator cost study: real operating costs per validator are unknown. The 50 to 200 unit range in this report is illustrative.
7. Delegator yield: what yield do delegators require, and what is the opportunity cost they compare against? Unknown.
8. Split: validator, delegator, community, and treasury shares, and whether the split is fixed per period or varies over time. The split determines both the treasury inflow total and the validator pay level.
9. Fee share to the treasury: what share of transaction fees goes to the treasury, and is it disclosed before genesis?
10. Coretime and blockspace: is there real demand for OGO blockspace, at what price? No evidence base yet.
11. Dust convention and taper parameters: carried from v0.1, still needing Benross's confirmation.
12. Legal analysis: written legal analysis is required before genesis distribution. Not yet commissioned.

## 9. What must be decided before a monetary model is selected

In order:

1. Period length. Without it, no schedule parameter has a calendar meaning, and no growth rate has an annual meaning.
2. Minimum validator set and validator cost basis. This sets the minimum fee requirement, which is the survival threshold for the whole design.
3. Treasury spend level and spending policy, including the post-issuance rule. The level must be affordable under the chosen model or justified by named post-issuance revenue.
4. Post-issuance funding mechanism and fee share, disclosed before genesis.
5. Model family (geometric or hybrid), with the scenario evidence in section 5 attached.
6. Schedule parameters for the chosen family, then the split, then the maximum supply derived as G + S_actual.

That order is the decision order Benross set, extended by one step at the front: period length first, because it converts everything else into time.
