# OGO Network: Issuance Mathematics v0.1

Status: Comparison for review. Not a decision. Maximum supply approved 2026-11-08: 100,000,000,000 OGO. No percentages, no intervals selected. All figures in this document are illustrative normalizations, not the approved parameters.
Date: 2026-11-06
Source: Benross approved starting the comparison (2026-11-06).

## 0. Read this first: every number below is illustrative

All four models are normalized to the same total scheduled issuance of 1,000,000 illustrative units so their SHAPES can be compared. The 1,000,000 figure is a normalization device, not a proposed supply. Period lengths, initial rates, reduction factors, and the 70/20/10 split used in one illustration are arbitrary inputs chosen to make the curves comparable. Nothing in this document selects a maximum supply, a genesis circulating supply, a reward percentage, a treasury percentage, or an issuance duration. The point of this stage is to compare economic models before committing OGO to an arbitrary number, per Benross's direction.

## 1. Common framework

Let M be the maximum supply (approved: 100,000,000,000 OGO, 2026-11-08; mandatory genesis parameter, no default). Let G be genesis circulating supply (TBD). Let S be total scheduled issuance. The cap constraint is:

G + S <= M

Per-period issuance is R_n for period n = 0, 1, 2, ... . The protocol may mint only R_n in period n, exactly as the schedule dictates. The cap is enforced every block: issuance in any period is min(R_n, M - total_supply). Once total_supply reaches M, issuance stops permanently.

Every model below must satisfy: deterministic (anyone can recompute R_n from genesis parameters), finite total (S is bounded and known at genesis), no discretionary minting, no ordinary-governance change, and integer arithmetic in uogo (see section 7).

The split question is orthogonal to the curve: whatever the split between validator/delegator rewards, community programs, and treasury, each bucket receives its share of R_n in each period, so each bucket follows the same curve shape. The curve decides WHEN money arrives; the split decides WHO gets it. Split percentages are TBD and are not chosen here.

## 2. Model A: Bitcoin-inspired halving

Definition: issuance is constant within an epoch of H periods, then halves at each epoch boundary.

R_n = R0 * (1/2)^floor(n / H)

Total over an infinite horizon: S = 2 * H * R0. Finite in practice because integer division in uogo eventually rounds the per-period amount to zero.

Illustrative parameters used: H = 4 periods, R0 = 125,000 (so S = 1,000,000).

Computed shape (normalized):
- 50% of all issuance is out by period 4. 90% by period 14. 99% by period 27.
- By period 10, 81.2% issued. By period 20, 96.9%. By period 40, 99.9%.
- Issuance at period 20 is 3.1% of period 0 issuance.
- Largest single-period drop: 50%, at every halving boundary.

Strengths: simplest to explain in one sentence; Bitcoin pedigree, widely understood; strongly front-loaded, which bootstraps early validators; no parameters beyond R0 and H.

Weaknesses: the 50% cliff. In Proof of Work, miners absorb halvings through price and fees; in Proof of Stake, validator rewards ARE the security budget, so revenue halving overnight is a security event, repeated at every boundary. Extremely front-loaded: half of all issuance ever goes to the earliest participants, which sits uneasily with a community-first distribution philosophy. The long tail is thin to the point of irrelevance (period 40 issuance is 122 units on this normalization), so the fee transition must be complete long before issuance ends.

OGO fit: good as philosophy (predictable, finite), weak as mechanics for a PoS network. The halving cliff is the core problem.

## 3. Model B: Geometric decay

Definition: each period issues a fixed fraction r (0 < r < 1) of the previous period.

R_n = R0 * r^n

Total: S = R0 / (1 - r). One tuning parameter, r, controls the whole shape.

Illustrative parameters used: r = 0.90, R0 = 100,000 (S = 1,000,000). Sensitivity case r = 0.95, R0 = 50,000.

Computed shape, r = 0.90 (normalized):
- 50% issued by period 7. 90% by period 22. 99% by period 44.
- By period 10, 65.1% issued. By period 20, 87.8%. By period 40, 98.5%.
- Issuance at period 20 is 12.2% of period 0. At period 40, 1.5%.
- Largest single-period drop: 10%, every period. No cliff.

Sensitivity, r = 0.95: 50% by period 14, 90% by period 45, 99% by period 90. Issuance at period 20 is 35.8% of period 0. A smaller reduction (r closer to 1) stretches the tail and pays validators longer, at the cost of a lower initial rate for the same total.

Strengths: smooth, no cliff anywhere; one parameter makes the tradeoff legible (r near 1 means long tail, r smaller means fast bootstrap); easy to verify independently (one formula); already implemented in the Milestone B scaffold, whose IssuanceSchedule is reducing-per-period.

Weaknesses: infinite horizon in theory, so the design must define an end rule (stop when per-period issuance rounds to zero in uogo, or at a fixed final period, with the final period truncated to the cap exactly). The tail is a long thin trickle, which is honest but means late-period rewards are mostly symbolic. r is a single number that carries a lot of weight, so its selection deserves its own review.

OGO fit: strong. The smooth decline matches a PoS security budget that must taper, not cliff, into fee revenue.

## 4. Model C: Epoch-based issuance

Definition: the schedule is an explicit table. Issuance is constant within each epoch and steps to a new level at epoch boundaries.

R_n = rate[epoch(n)], where epoch(n) = floor(n / E) and rate[] is a fixed list.

Total: S = E * sum(rate). Finite by construction: after the last epoch, issuance is zero.

Illustrative parameters used: 4 epochs of 10 periods, rates 40,000 / 30,000 / 20,000 / 10,000 (S = 1,000,000).

Computed shape (normalized):
- 50% issued by period 14. 90% by period 30. 99% by period 39. 100% at period 40.
- By period 10, 40.0% issued. By period 20, 70.0%.
- Issuance at period 20 is 50.0% of period 0 (the highest late-period level of any model).
- Largest single-period drop: 50% at the last step (20,000 to 10,000), 33% and 25% at earlier steps. The steps are cliffs, but they are visible in advance in a table anyone can read.

Strengths: most legible model. The schedule is a table, no formula to verify, and it ends cleanly at a known period. Highest sustained mid-schedule issuance, which keeps validators and community programs funded through the middle of the schedule. Any shape can be expressed (halving is the special case rate[k+1] = rate[k] / 2).

Weaknesses: every epoch rate is a separate number to agree on, so the design has the most parameters and the most surface for dispute. Steps are cliffs, even if smaller and fully disclosed. A flat rate inside an epoch means no smooth response to anything.

OGO fit: strong on transparency and auditability, the most governance-legible option. Best if the community wants the schedule readable as a table rather than a formula.

## 5. Model D: Hybrid issuance

Definition: a base rate that runs for a fixed horizon, plus a decaying bonus that bootstraps the early network.

R_n = base + B0 * r^n, for n < N; 0 after period N.

Total: S = base * N + B0 * (1 - r^N) / (1 - r).

Illustrative parameters used: base 10,000 per period, N = 40 periods, bonus B0 scaled so S = 1,000,000 (effective period 0 issuance 70,626).

Computed shape (normalized):
- 50% issued by period 11. 90% by period 32. 99% by period 40. 100% at period 40.
- By period 10, 49.5% issued. By period 20, 73.4%.
- Issuance at period 20 is 24.7% of period 0, and it never falls below the base floor while the schedule runs.
- Largest single-period drop: 8.6%. No cliff during the schedule; the only stop is the clean end at period N, when issuance is already at its floor.

Strengths: the only model with a floor. Validators, community programs, and treasury keep a steady minimum through the whole schedule, which directly answers the fee-transition problem: the transition is to a known floor first, then to fees at a known date. Strong early bootstrap plus a clean finite end.

Weaknesses: most parameters (base, B0, r, N), hardest to explain in one sentence, most surface for dispute. The floor must be justified: a base rate is a policy choice about minimum ongoing issuance, and it concentrates design arguments into one number.

OGO fit: strongest on substance for OGO's actual needs (sustained validator pay, steady community/treasury funding), weakest on simplicity. Worth carrying forward as the leading alternative.

## 6. Comparison

Normalized to total scheduled issuance 1,000,000 illustrative units. Periods counted from period 0.

| Property | A: Halving | B: Geometric | C: Epoch-based | D: Hybrid |
|---|---|---|---|---|
| 50% issued by period | 4 | 7 (r=0.90) | 14 | 11 |
| 90% issued by period | 14 | 22 | 30 | 32 |
| 99% issued by period | 27 | 44 | 39 | 40 |
| Issued by period 20 | 96.9% | 87.8% | 70.0% | 73.4% |
| Issuance at period 20 vs period 0 | 3.1% | 12.2% | 50.0% | 24.7% |
| Largest single-period drop | 50% | 10% | 50% | 8.6% |
| Finite by construction | No (rounds to zero) | No (needs end rule) | Yes | Yes |
| Parameters to agree | 2 (R0, H) | 2 (R0, r) | One per epoch | 4 (base, B0, r, N) |
| Legibility | One sentence | One formula | A table | A formula plus a floor |
| Fee-transition shape | Repeated cliffs | Smooth taper | Known steps | Floor, then clean stop |
| Front-loading | Extreme | High | Moderate | Moderate |

Split illustration (70% validators / 20% community / 10% treasury, ILLUSTRATIVE ONLY, same split applied to each model): period 0 validator pay is 87,500 under halving, 70,000 under geometric, 28,000 under epoch-based, 49,438 under hybrid. By period 20 the same bucket pays 2,734 / 8,510 / 14,000 / 12,215. The split does not change the ranking; the curve does. Any treasury funded as a share of issuance inherits the curve shape, so a front-loaded curve front-loads the treasury and a floored curve funds it steadily.

Fee transition: every model ends with issuance at or near zero, so validators must eventually live on fees. Halving forces that transition in four abrupt steps. Geometric and hybrid taper into it. Epoch-based schedules it in disclosed steps. The hybrid floor is the only design that keeps a minimum issuance alive until a chosen end date, which is the gentlest transition.

## 7. Integer arithmetic, rounding, and the final period

All four models are defined here in real numbers; the chain works in integer uogo. Rules the implementation must fix:

- Per-period amounts are computed with integer floor division.
- The dust from flooring accumulates as unissued headroom; it must never be mintable by any other path.
- The final period tops up or truncates so that total issued equals S exactly and total supply never exceeds M. Concretely, final period issuance = min(schedule amount, M - total_supply).
- Halving and geometric decay need an explicit end rule: stop when the computed per-period amount is zero, or at the fixed final period, whichever the genesis schedule states.
- The Milestone B scaffold already implements reducing-per-period (geometric) with integer-floor decrease; halving, epoch-table, and hybrid schedule types would be new schedule types in x/ogocoin, not new modules.

## 8. Recommendation

Recommendation, not a decision. The decision stays with Benross.

Carry forward geometric decay (Model B) as the default family for the next design step: smoothest taper, one tuning parameter, already implemented in the scaffold, no cliff. Carry hybrid (Model D) as the leading alternative, to be adopted if the fee-transition review concludes validators and the treasury need a minimum issuance floor through the schedule. Keep epoch-based (Model C) as the legibility fallback if the community prefers the schedule as a readable table. Do not carry pure halving (Model A) forward as the primary model for a Proof of Stake network: the 50% cliffs and the extreme front-loading (50% of all issuance in the first epoch) are the wrong shape for validator security and for community-first distribution, even though the Bitcoin philosophy behind it is the right inspiration.

## 9. What this does NOT decide

- Maximum OGO supply (TBD, mandatory genesis parameter).
- Genesis circulating supply (TBD).
- Issuance duration in blocks or time (TBD; period length is a parameter, not chosen here).
- Reduction factor, halving interval, epoch rates, base rate, or bonus size (TBD).
- Validator/delegator reward percentages, community program percentages, treasury funding percentage (TBD).
- Transaction fee policy (TBD; fees are the eventual replacement for issuance, and their design is a separate decision).
- Legal classification (written legal analysis required before genesis distribution).

## 10. Next step

Once Benross picks a model family (or asks for a different comparison), the next design step is parameter selection for that family: duration and reduction shape first, then the split between rewards, community, and treasury, and only then the maximum supply number, derived as G + S with the chosen schedule. That order, schedule math before supply number, is the direction he set.
