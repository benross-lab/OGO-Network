# OGO Network: Geometric vs. Hybrid Deep-Dive v0.1

Status: Research and design only. No monetary model has been approved. No maximum supply, no percentages, no intervals selected.
Date: 2026-11-06
Source: Benross handoff, OGO Network Issuance Mathematics v0.1 (2026-11-06).

## 0. Read this first: every number below is illustrative

All simulations normalize total scheduled issuance to 1,000,000 illustrative units so the two models can be compared on shape. The 1,000,000 figure is a normalization device, not a proposed supply. Reduction factors (17/20, 9/10, 19/20), the 40-period hybrid horizon, base shares (0%, 20%, 40%, 60% of total), and the 70/20/10 split used in one illustration are hypothetical inputs chosen to show sensitivity. Labelled as such everywhere they appear. Nothing here selects a maximum supply, a genesis circulating supply, an issuance duration, a reduction factor, a base rate, a bonus size, a reward split, a treasury percentage, or a fee policy.

## 1. The two models, stated exactly

Periods are 0-indexed. Period n covers block heights [start + n*L, start + (n+1)*L - 1], where start is the genesis start height and L is the period length in blocks (TBD). Issuance for period n is minted once, at the first block of period n+1 or the last block of period n, per the implementation rule in section 3. R_n is issuance in period n, in integer uogo. S is total scheduled issuance. G is genesis circulating supply. M is maximum supply. The cap constraint is G + S <= M, and issuance in any period is min(R_n, M - total_supply).

### Geometric decay

R_n = floor(R0 * r^n), with r = p/q a rational number, 0 < r < 1.

Analytic total (real numbers, before flooring): S = R0 / (1 - r). R0 is therefore determined by S and r: R0 = S * (1 - r). One tuning parameter, r, controls the whole shape.

### Hybrid

R_n = base + floor(B0 * r^n) for n < N, and 0 for n >= N.

base is a constant integer per period. B0 is the initial bonus. N is the number of periods. Analytic total: S = base * N + B0 * (1 - r^N) / (1 - r). Four parameters: base, B0, r, N. The base share f = (base * N) / S is the fraction of total issuance that arrives as a steady floor.

Relationship: hybrid with base share 0% is geometric decay truncated at N periods. Geometric decay is the limit of hybrid as base goes to 0 and N goes to infinity. The models are one family with a floor dial, not two unrelated designs.

## 2. Exact rational arithmetic and determinism

### Why exact rational, not floating point

The schedule must be computable identically by every independent validator. Floating point cannot guarantee that. The binary double 0.9 is not 9/10; it is 0.9000000000000000222 in decimal, and 0.9**100 in float differs from the exact rational (9/10)^100 in its last digits (float 2.6561398887587544e-05, exact 2.6561398887587476e-05). Summation order changes float results: summing the 110 float period amounts forward gives 999990.7386128694, summing the same values in reverse gives 999990.7386128692. Two validators that sum in different orders would disagree by a fraction of a uogo, and a consensus protocol cannot tolerate a disagreement of any size in supply accounting.

Exact rational arithmetic removes the whole class of error. All schedule parameters are integers or rational numbers p/q. Every per-period amount is computed as an exact Fraction and then floored once to integer uogo. Integer addition is associative and exact on every machine. Two validators with the same genesis parameters produce the same R_n for every n, byte for byte.

Honest note on the demonstration: in the r = 9/10, R0 = 100,000 case, the float per-period floors happened to agree with the exact floors in all 110 periods. That agreement is luck, not a guarantee. It does not hold in general, and the float total already diverges by summation order in the same case. The rule is exact rational arithmetic in consensus code, with float allowed only in off-chain analysis and clearly labelled as approximate.

### Rounding rule

Per-period issuance is floor(R0 * r^n) or base + floor(B0 * r^n). Flooring is the only rounding rule. The fractional dust from flooring is never minted by any other path. It accumulates as permanently unissued headroom. In the simulations below, dust is 45 to 118 uogo on a 1,000,000-unit normalization (0.005% to 0.012%). See section 9 for the open question of whether the final period tops up the dust or leaves it unissued.

### Period indexing

Period 0 is the first issuance period, starting at the genesis start height. Period n covers heights [start + n*L, start + (n+1)*L - 1]. The issuance for period n is claimed once, tracked by an issued-by-period set, and a second claim for the same period fails. Off-by-one risks to test: a period boundary at the wrong height issues the wrong period's amount; a 1-indexed implementation against a 0-indexed schedule shifts every amount by one period. The Milestone B scaffold already implements this with PeriodForHeight and IssuedByPeriod, and the test suite must include a boundary test that mints at the last block of period n and the first block of period n+1 and asserts different amounts.

### Terminal issuance

Geometric decay has no natural end in real numbers; in integer uogo it ends when floor(R0 * r^n) reaches 0. With R0 = 100,000 and r = 9/10, that is period 110 (R_109 = 1, R_110 would be 0). The design must still state an explicit end rule, because the terminal period depends on R0 and r and is not visible from the formula alone. Two candidate end rules, both cap-safe: (a) stop when the computed per-period amount is zero; (b) a fixed final period N in genesis, with the final period's issuance set to min(schedule amount, M - total_supply). Rule (b) is recommended because the end date is then known at genesis instead of derived. Hybrid ends by construction at period N; its terminal rule is the same min() cap check on the final period.

## 3. Sensitivity simulations (exact rational, floored to uogo)

All figures computed with Python Fractions, floored to integers. Total scheduled issuance normalized to 1,000,000 illustrative units.

### Geometric decay, reduction factor sweep

| r | R0 | Periods until floor is 0 | 50% by | 90% by | 99% by | R20 as % of R0 | R40 | Dust |
|---|---|---|---|---|---|---|---|---|
| 17/20 (0.85) | 150,000 | 74 | period 5 | period 15 | period 29 | 3.9% | 225 | 45 |
| 9/10 (0.90) | 100,000 | 110 | period 7 | period 22 | period 44 | 12.2% | 1,478 | 65 |
| 19/20 (0.95) | 50,000 | 211 | period 14 | period 45 | period 90 | 35.8% | 6,425 | 118 |

Reading: r is the whole design. Moving r from 0.85 to 0.95 triples the terminal horizon (74 to 211 periods), moves 90% issuance from period 15 to period 45, and multiplies period-20 issuance by three. A higher r pays validators longer at a lower initial rate for the same total. A lower r bootstraps harder and ends sooner.

### Hybrid, base-share sweep (N = 40 periods, r = 9/10)

| Base share of total | Base per period | R0 | 50% by | 90% by | 99% by | R20 as % of R0 | Last period R >= 10% of R0 |
|---|---|---|---|---|---|---|---|
| 0% (pure geometric, truncated) | 0 | 101,500 | period 7 | period 21 | period 36 | 12.2% | period 21 |
| 20% | 5,000 | 86,200 | period 8 | period 28 | period 39 | 17.3% | period 29 |
| 40% | 10,000 | 70,900 | period 11 | period 32 | period 40 | 24.5% | period 39 |
| 60% | 15,000 | 55,600 | period 14 | period 34 | period 40 | 35.9% | period 39 |

Reading: the base share is the floor dial. At 40% base, issuance never falls below 14.1% of its initial level while the schedule runs, and the last period at or above 10% of R0 is period 39, the final period. At 0% base the floor does not exist and the model collapses to truncated geometric. Every 20 points of base share moves 90% issuance roughly 4 to 7 periods later and roughly doubles the period-20 issuance share.

Note on the earlier v0.1 figures: the hybrid illustration there (R0 70,626) was float-normalized; the exact-rational value here is 70,900. The difference is the normalization method, not a model change, and it is under 0.4%.

## 4. Validator and delegator reward sustainability

Measure sustainability as the last period in which issuance is at least 10% (and 5%) of its initial level, on the assumption that validator operating costs do not fall with issuance. Under the illustrative split, validator/delegator rewards are a fixed share of R_n each period, so these horizons apply directly to reward income.

| Model | Last period R >= 10% of R0 | Last period R >= 5% of R0 |
|---|---|---|
| Geometric r = 0.85 | period 14 | period 18 |
| Geometric r = 0.90 | period 21 | period 28 |
| Geometric r = 0.95 | period 44 | period 58 |
| Hybrid base 20% | period 29 | period 39 |
| Hybrid base 40% | period 39 | period 39 |
| Hybrid base 60% | period 39 | period 39 |

Geometric sustainability is entirely a function of r: a 0.05 change in r more than doubles the 10% horizon. Hybrid sustainability is a function of the floor: once the base exceeds 10% of R0, the 10% horizon is the full schedule by construction. That is the hybrid's core claim. It buys sustainability with total issuance (the base consumes share that would otherwise go to early periods) rather than with a longer tail.

Unresolved: what validator income level is actually sustainable depends on validator costs, stake size, and fee income, none of which are known yet. These horizons are relative measures, not viability findings.

## 5. Early versus late participant distribution

Share of total scheduled issuance in the first 10 periods (early participants) and concentration in the top 10 periods (identical to the first 10 here, since both series decline monotonically):

| Model | First-10-periods share | Issued by period 20 | Issued by period 40 |
|---|---|---|---|
| Geometric r = 0.85 | 80.3% | 96.1% | 99.9% |
| Geometric r = 0.90 | 65.1% | 87.8% | 98.5% |
| Geometric r = 0.95 | 40.1% | 64.2% | 87.2% |
| Hybrid base 0% | 66.1% | 89.2% | 100.0% |
| Hybrid base 20% | 57.9% | 81.3% | 100.0% |
| Hybrid base 40% | 49.7% | 73.5% | 100.0% |
| Hybrid base 60% | 41.4% | 65.7% | 100.0% |

The community-first question is how much of the money goes to whoever is present earliest. Geometric r = 0.85 gives the first 10 periods 80.3% of everything ever issued. Hybrid base 60% gives them 41.4%. The floor is the most effective tool in this comparison for shifting issuance toward later participants, more effective than raising r, because it redirects issuance from early periods to every later period rather than stretching the same early-heavy curve.

## 6. Long-term security budget

In Proof of Stake, issuance is the security budget until fees replace it. The budget in period n is R_n (times the validator share of the split). The budget falls monotonically in both models. The difference is the floor: geometric's budget decays toward zero asymptotically, so security funding in the late schedule is whatever fees provide plus a negligible trickle. Hybrid's budget decays toward the base floor and stays there until period N, so the network has a known minimum security budget for the whole schedule.

After the schedule ends, both models have zero issuance and the security budget is fees alone. Neither model avoids the fee transition; hybrid only postpones its hardest phase to a known date and keeps a floor under it until then. The long-term security question is therefore not which model avoids fees, but which model gives the fee market time to develop: geometric r = 0.95 and hybrid base 40% both keep issuance above 10% of initial for roughly 40 periods, while geometric r = 0.85 drops below 10% by period 14.

## 7. Fee-transition requirements

Required fee revenue to hold validator pay at its period-0 level is R0 - R_n, expressed as a percentage of R0. It is the fee market's job description at each period.

| Model | Fees must replace by period 20 | Fees must replace by period 40 |
|---|---|---|
| Geometric r = 0.85 | 96.1% | 99.9% |
| Geometric r = 0.90 | 87.8% | 98.5% |
| Geometric r = 0.95 | 64.2% | 87.2% |
| Hybrid base 20% | 82.7% | 100.0% (schedule ended) |
| Hybrid base 40% | 75.5% | 100.0% (schedule ended) |
| Hybrid base 60% | 64.1% | 100.0% (schedule ended) |

Note the hybrid paradox: its fee-replacement percentage at period 20 is lower (gentler transition), but at period 40 it is 100% in one step because the schedule ends. The gentleness is real during the schedule; the end is still a cliff to zero fees-only operation, just a cliff from a floor rather than from a trickle. If the end cliff is a concern, the mitigation is a longer N or a final taper of the base itself, which is a further design option, not a property of either model as defined here.

## 8. Treasury and community funding continuity

Under the illustrative split of 70% validators/delegators, 20% community programs, 10% treasury, each bucket receives its share of R_n per period. The split is illustrative only; the continuity pattern is what matters.

Geometric r = 0.90: treasury receives 10,000 per period at period 0, 1,215 at period 20, 147 at period 40. Cumulative treasury by period 20 is 87,833 of its 99,993 total. The treasury is funded overwhelmingly early and then starves.

Hybrid base 40%: treasury receives 7,090 at period 0, 1,740 at period 20, and 0 at period 40 because the schedule has ended. Cumulative treasury by period 20 is 73,488 of its 99,998 total. Less front-loaded, steadier while it runs, then stops at the known end date.

Continuity conclusion: neither model funds a treasury forever; both fund it for the schedule's life. Hybrid funds it more evenly across the schedule; geometric funds it mostly at the start. If community programs need funding in the later schedule, the floor matters more than the total. If the treasury is meant to be built early and then spent down, geometric's front-loading is a feature.

## 9. Total issuance and time-to-cap

Time-to-cap has two meanings and they must not be confused. Time to 99% of scheduled issuance is 29 to 90 periods for geometric (depending on r) and 36 to 40 periods for hybrid at N = 40. Time to terminal issuance (floor reaches zero, or period N) is 74 to 211 periods for geometric and exactly N = 40 for hybrid in these illustrations. The cap M is reached only when genesis supply plus all scheduled issuance is minted, and burning reduces total outstanding supply without restoring issuance headroom. Total issuance S is fixed at genesis in both models; the schedule shape changes when the issuance arrives, never how much arrives in total.

## 10. Technical requirements check

Each requirement from the handoff, and how each model preserves it:

- Fixed maximum supply: both. S is bounded and known at genesis; G + S <= M is validated at genesis.
- Mandatory genesis maximum-supply parameter, no default: both. The schedule parameters are also mandatory genesis parameters, with no defaults.
- Deterministic scheduled issuance: both, provided the implementation uses exact rational arithmetic (section 2) and a fixed rounding rule.
- No discretionary, administrative, founder, or governance minting: both. The only mint entrypoint is the scheduled one; all other mint paths fail. Verified for geometric in Milestone B bypass tests; hybrid adds the same entrypoint with a base term and needs the same bypass tests.
- No ordinary governance modification of the schedule: both. Schedule changes through ordinary governance are rejected; a schedule change requires the separately documented constitutional process.
- Supply cap enforced every block: both. Issuance per period is min(schedule amount, M - total_supply).
- Permanent issuance cessation at the cap: both. Once total_supply reaches M, issuance is zero forever.
- One-way burning: both. Burn reduces total supply with no remint credit.
- IBC escrow/voucher accounting without double-counting: both. Unaffected by the issuance curve; escrow plus voucher accounting is ledger-level and was tested in Milestone B.
- Integer-based deterministic calculations: both, if implemented with exact rational arithmetic and floor rounding as specified in section 2. This is a requirement on the implementation, not a property that follows automatically from choosing a model.
- Reproducible independent verification: both. Any outsider with the genesis file can recompute every R_n exactly.

Implementation note: the Milestone B scaffold implements reducing-per-period (geometric) only. A hybrid schedule needs a new schedule type (base plus bonus) in x/ogocoin, with its own validation, bypass tests, and terminal-period handling. That is a small extension, not a new module.

## 11. Assumptions

1. Total scheduled issuance normalized to 1,000,000 illustrative units in all simulations.
2. Periods are uniform in length (L blocks each). Period length in blocks or time is TBD.
3. The validator/delegator/community/treasury split is a fixed share of R_n each period. A time-varying split is possible but was not simulated.
4. Validator costs are not modelled. Sustainability horizons are relative to initial issuance, not to any cost figure.
5. Fee revenue is not modelled. Fee-transition percentages describe what fees would need to replace, not what fees will provide.
6. Stake size and validator count are not modelled. Reward per validator depends on the split among validators and is out of scope here.
7. The hybrid horizon N = 40 periods is illustrative. The geometric terminal periods (74, 110, 211) are derived from the illustrative R0 and r, not chosen.
8. All arithmetic in the simulations is exact rational, floored once to integer uogo.

## 12. Unresolved questions

1. Floor or no floor: is a minimum ongoing issuance floor (hybrid) wanted, or is a smooth taper to zero (geometric) sufficient? This is the decision that separates the two models.
2. Tail length: if geometric, what r? The sensitivity table shows r is the whole design; a 0.05 change more than doubles the sustainability horizon.
3. Dust: the 45 to 118 uogo of flooring dust per 1,000,000 normalized units. Leave it permanently unissued (S as a cap, not a promise), or top up the final period so total issued equals S exactly? Recommendation: leave it unissued; it is simpler and cap-preserving. Needs Benross's confirmation.
4. End rule for geometric: stop when the per-period amount floors to zero, or a fixed final period N in genesis? Recommendation: fixed final period, so the end date is known at genesis.
5. Period length: how many blocks (or how much time) is one period? This converts every period count in this document into calendar time and must be decided with the schedule parameters.
6. Split: validator/delegator, community, and treasury shares, and whether the split is fixed per period or varies over time. Not simulated here.
7. Fee policy: what fee market is expected to replace issuance, and when is it expected to be sufficient? Requires its own analysis (decision order step 4).
8. Hybrid base justification: the base rate is a policy choice about minimum ongoing issuance. If hybrid is chosen, the base level needs its own justification, not just a share of total.
9. Final taper: should the hybrid base itself taper in the last periods to soften the end cliff, or end at full floor value at period N?

## 13. Recommendation

Recommendation, not a decision. The decision stays with Benross.

Both models preserve every technical requirement. The choice turns on two questions only: do you want a floor, and how long a tail do you want?

- If no floor is wanted, choose geometric decay. It is simpler, already implemented, and its one parameter (r) is the entire design. The open parameter is r, and the sensitivity table is the guide.
- If a floor is wanted so validators, community programs, and the treasury keep a known minimum through the schedule, choose hybrid. It costs four parameters instead of two and needs a new schedule type in x/ogocoin, but it is the only model that sustains late-schedule funding.

The next step after the family is chosen is parameter selection for that family, then the split, then fee analysis, then the maximum supply derived as G + S. That order is the decision order Benross set.
