# OGO Treasury and Sustainability

Status: Technical specification and scenario model. Planning only. No treasury funded, no spending approved, no monetary parameter selected, no production change authorized.
Date: 2026-11-06
Companion documents: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (sections 4.4, 4.5, 5.4), `OGO_MONETARY_ENGINE_SPEC.md` (section 7)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.
All figures in this document are ILLUSTRATIVE comparison inputs, not selected parameters. No OGO/USD price is assumed anywhere.

## 1. Purpose

This document continues the evaluation of the R-F treasury model, compares it with alternative reserve and spending approaches, and tests it under low, moderate, and high network adoption, low, moderate, and high fee revenue, different validator counts, different operating costs, geometric and hybrid issuance, post-issuance sustainability, emergency reserve requirements, and treasury depletion scenarios. It keeps the corrected November 6 accounting conventions, and it clearly separates protocol treasury income from OGO company revenue.

## 2. The R-F model

Recommended default, not an approved rule. One formula for both modes (issuance and post-issuance), which is why it is recommended.

Definition: let B be the treasury balance at the end of the previous period, D be the desired spend for the period (from approved proposals), Trail_4 be the trailing four-period average treasury inflow, and F be the emergency reserve floor (genesis-fixed, TBD). Then:

spend_t = min(D, max(Trail_4(t), floor(0.05 * B_(t-1))), max(0, B_(t-1) - F))

In words: spend no more than the desired amount, and no more than the larger of (the trailing inflow average, or 5 percent of the balance), and never spend below the emergency reserve floor. The transition from issuance mode to post-issuance mode is keyed to the genesis-fixed final issuance period N, not to a governance vote.

Why R-F is recommended: it is one formula for both modes, so there is no mode-switch discontinuity; it keeps spending after issuance ends (unlike trailing-inflow caps, which freeze spending to zero after issuance ends); and it is bounded by the balance and the floor, so it cannot bankrupt the treasury in a single period.

### 2.1 Desired spending is not sustainable spending

Desired spending D and actual permitted spending are different numbers, and this document keeps them separate throughout. D is what approved proposals ask to spend in a period. Actual permitted spending is what the R-F rule allows given the settled balance, the trailing inflow, and the reserve floor: spend_t = min(D, max(Trail_4(t), floor(0.05 * B_(t-1))), max(0, B_(t-1) - F)). Sustainable spending is the level of actual permitted spending that can be maintained over the horizon without driving the balance to the floor; it is bounded by available inflows, not by desire. A spending rule rations a shortfall, it does not create the inflow that would remove it. Stated plainly: a spending rule cannot fix a level problem. If inflow is too low, no rule creates money, and presenting a spending formula as proof of treasury sustainability is wrong.

The low-fee finding, recorded 2026-11-06, corrected 2026-11-06 (GitHub publication): under Low fees, raising desired spend D from 1,500 to 2,500 per period changes actual R-F spending by 0 units over 120 periods (hybrid model). Both runs spend 114,511 in total (114,511.16 before rounding). The rule binds, not the desire. The scenario tables in this document show this directly: section 5 records 114,511 total spent at D = 1,500, and section 6 records 114,511 at D = 2,500. An earlier revision note in this document stated a 4-unit delta from a later recomputation; that figure is not reproducible from the stated model and is withdrawn. Recomputed 2026-11-06 from the stated model (post-inflow balance, trailing 4-period average inflow including the current period, float 10 percent treasury share of issuance, fee revenue terminal 2,462.5 ramping as 1 minus exp(-t/20), treasury fee share 10 percent, floor 10,000, horizon 120 periods); the recomputation reproduces every table total in sections 5 and 6 to rounding. See `docs/discrepancy-resolutions-2026-11-06.md`.

Treasury reporting requirement, added 2026-11-06: every treasury report, scenario table, and dashboard that presents treasury spending must report these six figures separately, for the same horizon, so a shortfall can never be hidden inside a single spend number: (1) desired treasury spending, (2) actual permitted spending, (3) available inflows, (4) reserve balance, (5) unfunded spending (desired minus actual permitted, per period and cumulative), and (6) number of periods with a shortfall. A report that shows only desired spending, or only actual spending, is incomplete.

## 3. Corrected November 6 accounting conventions

These conventions are kept throughout this document. They correct earlier errors, and they are not optional.

- **Three-number accounting:** S_target (theoretical target), S_actual (actual cumulative scheduled issuance, by exact integer summation), and M (maximum supply, M = G + S_actual). M is computed at genesis, not chosen first. Dust (S_target minus S_actual) is 45 to 118 uogo per 1,000,000 normalized units, depending on the model, and is left permanently unissued.
- **Decision Preparation v0.1 errata, recorded 2026-11-06:** P2 full spend at period 4/6, not about 20; hybrid period-39 balances P1 44,532, P3 45,982, P4 49,351; P4 trailing window inconsistent, consistent convention gives 58,827 geometric and 49,351 hybrid; quarterly minimum need 43,875, not 20,875, under calendar-consistent per-month costs.
- **Minimum need:** 14,625 fee revenue per period at lean costs, 24,625 at standard costs, for 100 validators illustrative, infrastructure only. If labour is included, the need is higher.
- **Fee level:** at a Cardano-like fee level, fee revenue is 469 per period against a minimum need of 14,625 (3.2 percent of need). Fees are 0.67 percent of rewards in Cardano over 73 epochs to 1 September 2026 (3.3 million ADA fees vs 493.7 million ADA rewards).
- **Endowment arithmetic:** a 300,000 balance is needed at 0.5 percent yield per period to fund the minimum need, which exceeds total treasury inflow. An endowment cannot be the primary post-issuance mechanism.
- **Hybrid valley, prior finding, corrected 2026-11-06:** 26 periods with a shortfall of 116,065 at 5 percent fee growth, treasury covering about half; 69 periods with a shortfall of 572,847 at 3 percent growth. The 26-period result assumes fee growth begins only after issuance ends. When fee growth begins at genesis instead, the valley is 0 periods (section 7). The valley is extremely growth-sensitive, and its length depends entirely on when fee growth starts relative to issuance end. Every scenario table in this document that quotes a valley length must identify its fee-growth start assumption.

## 4. Scenario model: inputs and assumptions

The scenario model in this section was computed for this document on 2026-11-06 (script `/tmp/treasury_v03.py`). It is an off-chain scenario model, not consensus code. Issuance uses exact integer arithmetic. Fee revenue and spending use plain floats, which is acceptable in a model and forbidden in consensus code.

Inputs, all illustrative:

- Total scheduled issuance normalized to 1,000,000 units. Geometric: r = 0.90, 110 periods, sum 999,935, dust 65. Hybrid: base 40 percent (10,000 per period), N = 40, r = 0.90, sum 999,978, dust 22.
- Treasury share of issuance: 10 percent (illustrative 70/20/10 split: validator, delegator, treasury). Treasury share of fee revenue: 10 percent (illustrative).
- Desired spend D: 1,500 per period (the fixed spend from the prior research), and 2,500 per period as a binding case.
- Emergency reserve floor F: 10,000, with sensitivity at 0 and 50,000.
- Fee revenue: an S-curve ramp to a terminal level, with time constant 20 periods, starting at genesis. Terminal levels are expressed as multiples of the minimum need (100 validators, standard cost, 24,625 per period): Low 0.1x (2,462), Moderate 0.5x (12,312), High 1.5x (36,938). The S-curve assumption matters: fees ramp from genesis, so by the time hybrid issuance ends at period 40, fee revenue is already at 86 percent of its terminal level. Section 7 states how the results change if fees ramp only after issuance ends.
- Horizon: 120 periods. Period length is not selected; the model uses periods, not years or months.

## 5. Results: R-F under low, moderate, and high fee revenue

Desired spend 1,500 per period, floor F = 10,000. Balance at issuance end, balance at period 120, post-issuance survival (periods the balance stays above the floor after issuance ends), and total spent over 120 periods. Full desired spend over 120 periods would be 178,500.

| Fee scenario | Model | Balance at issuance end | Balance at period 120 | Post-issuance survival | Total spent |
|---|---|---|---|---|---|
| Low (0.1x need) | Geometric | 10,000 (at floor) | 10,000 (at floor) | 0 periods | 114,507 |
| Low (0.1x need) | Hybrid | 46,982 | 10,000 (at floor) | 43 periods | 114,511 |
| Moderate (0.5x need) | Geometric | 46,788 | 44,060 | 10 periods (full window) | 178,500 |
| Moderate (0.5x need) | Hybrid | 68,919 | 44,065 | 80 periods (full window) | 178,500 |
| High (1.5x need) | Geometric | 267,378 | 289,194 | 10 periods (full window) | 178,500 |
| High (1.5x need) | Hybrid | 123,760 | 289,198 | 80 periods (full window) | 178,500 |

Findings:

- **Under Low fees, the treasury ends at the floor under both models.** Geometric reaches the floor during issuance (its balance is already 10,000 at period 110). Hybrid reaches the floor 43 periods after issuance ends. The R-F rule prevents the balance going below the floor, so the treasury is never negative, but at the floor no further spending is possible. Low fees are a depletion scenario, not a sustainable one.
- **Under Moderate fees, the full desired spend is paid (178,500) and the balance is stable.** Both models end near 44,000. The treasury survives the full 120-period horizon above the floor.
- **Under High fees, the balance grows strongly.** Both models end near 289,000. The treasury is more than sustainable; the question becomes whether the desired spend should be higher, which is a governance decision, not a model outcome.
- **The binding finding:** under Low fees, raising the desired spend does not raise actual spend. At D = 2,500 (section 6), the R-F rule spends 114,511 over 120 periods for hybrid, essentially identical to the 114,511 spent at D = 1,500. The rule binds, not the desire. A spending rule cannot fix a level problem: if inflow is too low, no rule creates money.

Revision note, 2026-11-06, corrected 2026-11-06 (GitHub publication): the binding finding above is confirmed by recomputation, which puts the exact change at 0 units when D rises from 1,500 to 2,500 under Low fees: both runs spend 114,511 over 120 periods. An earlier version of this note stated 4 units; that figure is withdrawn as not reproducible from the stated model. Desired spending and actual permitted spending are now separated throughout this document (section 2.1), and the six-figure treasury reporting requirement in section 2.1 applies to every scenario table here: desired spending, actual permitted spending, available inflows, reserve balance, unfunded spending, and number of periods with a shortfall.
- **Model difference:** geometric ends issuance with more treasury balance than hybrid under Moderate and High fees (267,378 vs 123,760 under High), because its issuance runs longer (110 vs 40 periods) and its treasury accumulates over more periods. But geometric also reaches the floor earlier under Low fees. The model choice changes the shape of the treasury curve, not whether Low fees are sustainable. They are not, under either model.

## 6. Results: R-F compared with alternative reserve and spending approaches

Binding case: hybrid, desired spend 2,500 per period, floor F = 10,000. Rules compared: fixed spend (spend D, bounded only by the floor); trailing-inflow cap only (spend the smaller of D and Trail_4); balance-5-percent cap only (spend the smaller of D and 5 percent of B); R-F (spend the smaller of D and the larger of Trail_4 and 5 percent of B); P4 combined (spend the smaller of D, Trail_4, and 5 percent of B).

High fees (terminal 1.5x need):

| Rule | Balance at issuance end | Balance at period 120 | Total spent |
|---|---|---|---|
| Fixed | 84,760 | 170,198 | 297,500 (full desired) |
| Trailing-only | 84,760 | 170,198 | 297,500 |
| Balance-5-percent-only | 93,323 | 178,760 | 288,938 |
| R-F | 84,760 | 170,198 | 297,500 |
| P4 | 93,323 | 178,760 | 288,938 |

Under High fees, inflow is high enough that the desired spend is paid in full under fixed, trailing-only, and R-F. The balance cap binds early (when the balance is small), so balance-5-percent-only and P4 spend less and end higher. The rules differ only in how much they throttle early spending.

Low fees (terminal 0.1x need):

| Rule | Balance at issuance end | Balance at period 120 | Total spent |
|---|---|---|---|
| Fixed | 10,000 (at floor) | 10,000 (at floor) | 114,511 |
| Trailing-only | 27,538 | 25,930 | 98,582 |
| Balance-5-percent-only | 32,544 | 10,000 (at floor) | 114,511 |
| R-F | 27,023 | 10,000 (at floor) | 114,511 |
| P4 | 39,917 | 38,309 | 86,202 |

Under Low fees, the rules diverge sharply. P4 is the most conservative: it spends the least (86,202) and ends with the highest balance (38,309). Trailing-only preserves the balance better than R-F (25,930 vs 10,000 at period 120) but spends less (98,582 vs 114,511). R-F spends the full amount the inflow allows and ends at the floor. Fixed spends the same total as R-F but reaches the floor earliest.

Reading: P4 is the most conservative option, and it is the only rule in this comparison that keeps the treasury above the floor at period 120 under Low fees. That is why P4 remains a candidate. But P4 has a known defect: it freezes spending to zero after issuance ends if the trailing inflow is zero, which is exactly the post-issuance problem it is meant to solve. R-F is recommended because it is one formula for both modes and keeps spending after issuance ends, but under Low fees it spends the treasury down to the floor. Neither rule fixes Low fees. The spend level and the fee revenue must be addressed directly.

## 7. Post-issuance sustainability and the hybrid valley

Post-issuance sustainability means: after period N, when issuance stops permanently, can the treasury and the validators be funded from fee revenue and other disclosed mechanisms?

- **Treasury post-issuance:** under Moderate and High fees, the treasury survives the full horizon above the floor under both models (section 5). Under Low fees, it ends at the floor: geometric during issuance, hybrid 43 periods after issuance ends. Post-issuance treasury survival is therefore a fee-revenue question, not a rule question.
- **Validator post-issuance funding:** fee revenue at terminal level, against the minimum need (100 validators):

| Fee scenario | Terminal fee revenue | Vs lean need (14,625) | Vs standard need (24,625) | Vs professional need (61,562, derived) |
|---|---|---|---|---|
| Low | 2,462 | 16.8 percent | 10.0 percent | 4.0 percent |
| Moderate | 12,312 | 84.2 percent | 50.0 percent | 20.0 percent |
| High | 36,938 | 252.6 percent | 150.0 percent | 60.0 percent |

Only High fee revenue covers the standard minimum need (150 percent). Moderate covers half of it. Low covers a tenth. Even High covers only 60 percent of the professional need. If the validator set is 100 validators at professional cost, no fee scenario in this model covers the need. The validator count and the cost tier must be decided before the fee policy can be judged adequate.

- **The hybrid valley:** the period after issuance ends when fee revenue has not yet grown to cover the minimum need. In this model, fees ramp from genesis (S-curve, time constant 20 periods), so by period 40 fee revenue is already at 86 percent of terminal. Under High fees, fee revenue already covers the need at period 40, so the valley is 0 periods. Under Low and Moderate fees, fee revenue never covers the need within 120 periods, so there is no valley to recover from: it is a permanent shortfall, not a valley. The prior finding (26-period valley, 116,065 shortfall at 5 percent growth; 69 periods, 572,847 at 3 percent) assumed fees ramp only after issuance ends. Both findings are correct under their assumptions. The assumption must be stated whenever the valley is quoted: the valley length depends entirely on when fee growth starts relative to issuance end. A fee policy that assumes fees will be mature at issuance end must say so, and must say what evidence supports it.

Revision note, 2026-11-06: the 26-period valley result is corrected here, not removed. It is the result under the assumption that fee growth begins only after issuance ends. Under the assumption that fee growth begins at genesis, the valley is 0 periods: fee revenue is already mature enough at issuance end to cover the need under High fees. The two results answer different questions, and quoting the 26-period figure without naming its assumption is the error this correction fixes. The fee-growth start assumption is now identified wherever the valley is quoted in this document: in section 3 for the prior finding, and in this section for the model in sections 5 and 6, where fees ramp from genesis (section 4).

## 8. Emergency reserve requirements

The emergency reserve floor F is the balance below which the treasury cannot spend except through a separate emergency process, documented before genesis. Sensitivity, hybrid, High fees, desired spend 1,500:

| Floor F | Balance at issuance end | Balance at period 120 | Total spent |
|---|---|---|---|
| 0 | 122,260 | 287,698 | 180,000 (full desired) |
| 10,000 | 123,760 | 289,198 | 178,500 |
| 50,000 | 134,260 | 299,698 | 168,000 |

A higher floor preserves more balance and spends less, because spending stops earlier. The floor is insurance, not income: it does not create funds, it reserves them. The value of F is TBD. Setting F requires answering: what emergency is the reserve for (a fee collapse, a halt, a governance failure, a security incident), how much would that emergency cost, and who decides to spend below the floor. If those questions have no answer, F is an arbitrary number, and an arbitrary floor is worse than a reasoned one, because it constrains spending without a purpose.

## 9. Treasury depletion scenarios

Depletion means the treasury balance reaches the floor F and no further spending is possible. The R-F rule prevents a negative balance, so depletion in this model is reaching the floor, not going below zero. Scenarios:

- **Low fees, either model:** depletion within the horizon. Geometric reaches the floor during issuance. Hybrid reaches it 43 periods after issuance ends. The treasury cannot fund the desired spend for the full horizon.
- **Moderate fees, either model:** no depletion within 120 periods. The balance is stable near 44,000.
- **High fees, either model:** no depletion. The balance grows.
- **Desired spend above inflow:** raising D does not raise actual spend (section 5). The treasury is not depleted faster; the spending simply does not happen. This is the R-F rule working as designed, but it means the desired spend is not delivered, which is a different failure: the network's programs are underfunded, not the treasury overdrawn.
- **Fee collapse after issuance:** if fee revenue falls after issuance ends (for example, a competing chain takes volume), the treasury spends down to the floor at the R-F rate, then stops. The time from the collapse to the floor is the treasury's runway, and it depends on the balance at the collapse and the desired spend. The runway must be computed for the approved parameters before genesis, not discovered after a collapse.

## 10. Protocol treasury income vs OGO company revenue

Clearly separated. Protocol treasury income and OGO company revenue are different money, in different accounts, under different control, and they must never be confused.

Protocol treasury income (on-chain, in x/treasury, visible in the explorer):

- A share of scheduled issuance, per the approved split.
- A share of transaction fees, per the approved fee distribution.
- Any other explicitly disclosed on-chain mechanism named before genesis (for example, blockspace or coretime sales, if designed and modelled).

OGO company revenue (off-chain, in OGO Systems' accounts, not on-chain):

- OGO Pro subscriptions.
- AI inference charges.
- Hosted apps revenue.
- Marketplace revenue (application fees, not protocol fees).

Company revenue must not be counted as protocol revenue unless an explicit, implemented, disclosed mechanism routes a portion into the network. Such a mechanism would need: a written specification of what portion is routed, how it is routed (a transaction, a contract, a manual transfer), how often, who authorizes it, and how it is verified on-chain. Until such a mechanism exists, is implemented, and is disclosed, company revenue contributes zero to protocol revenue, to the treasury, and to validator funding. Any document, dashboard, or projection that adds company revenue to protocol income without naming the mechanism is wrong, and it must be corrected, not footnoted.

The same separation applies in the other direction: protocol treasury funds must not be presented as company revenue, and treasury spending must not be used to fund company operations unless a governance-approved treasury proposal explicitly authorizes it, within the R-F rule, for a stated network purpose.

## 11. Post-issuance funding mechanisms

After period N, issuance stops permanently. The treasury receives only its share of transaction fees plus any other explicitly disclosed funding mechanism named before genesis:

- **Fee share:** the primary post-issuance mechanism. Its adequacy depends on the fee market (see `OGO_FEE_MARKET_SPEC.md`). At a Cardano-like fee level it produces 469 per period against a minimum need of 14,625 (3.2 percent). Fee revenue must be modelled, not assumed.
- **Blockspace or coretime sales:** selling blockspace or execution capacity as a separate revenue line. A named mechanism, but no design exists yet. It must be designed and its revenue modelled before genesis if it is to be relied on.
- **Endowment:** investing treasury funds and spending the yield. Fails on arithmetic: a 300,000 balance is needed at 0.5 percent yield per period to fund the minimum need, which exceeds total treasury inflow. An endowment cannot be the primary post-issuance mechanism unless the treasury balance is far larger than the research normalisation implies.
- **Contributions:** voluntary contributions to the treasury. Real but unpredictable, and cannot be budgeted.
- **Treasury staking:** noted as circular. Staking treasury funds to earn rewards that fund the treasury is staking the treasury's own coins to pay itself. It is not a funding mechanism and is not relied on in these specifications.

Decision order, preserved: name the post-issuance funding mechanism and the fee share before genesis. Do not launch a fixed-supply network without a written answer to what pays validators and the treasury after issuance ends.

## 12. What this document does not decide

Treasury spend level D; emergency reserve floor F; treasury rule (R-F recommended, P4 candidate); fee share and fee policy; validator count and cost tier; post-issuance funding mechanism; and the split. The spend level must be set at a level the inflow can support. No spending policy fixes a level problem. All are listed in `OGO_NETWORK_DECISION_REGISTER.md`.
