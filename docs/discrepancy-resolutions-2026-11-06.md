# OGO Network: Discrepancy Resolutions, 2026-11-06

Status: Reconciliation record. No monetary model has been approved. No maximum supply, no percentages, no intervals selected. All figures below are illustrative, normalized to 1,000,000 illustrative units of total scheduled issuance unless stated otherwise. The normalization is not a proposed supply.

Date: 2026-11-06

Purpose: resolve the two discrepancies flagged in the first monthly OGO Exchange Readiness Report (2026-11-06) before the authorized scope is published to the public benross-lab/OGO-Network repository: (1) hybrid dust 18 vs 22 uogo, (2) the low-fee R-F binding delta 4 vs 0 units. Both are resolved here by recomputation from the source numbers. Neither resolution selects a monetary parameter.

Recomputation method: exact rational arithmetic (Python Fractions, floored to integer uogo) for issuance; plain floats for the off-chain treasury scenario model, as the scenario model itself uses floats. The recomputation scripts are kept with this note in `docs/discrepancy-resolutions/` (`dust_recompute.py`, `rf_recompute.py`) so anyone can re-run them.

## 1. Hybrid dust: 18 vs 22 uogo

### Resolution: 22 is correct. 18 is withdrawn.

Under the corpus definition of the hybrid schedule, as implemented in the prototype, dust is 22 uogo per 1,000,000 normalized units (S_actual 999,978).

### The recomputation

Hybrid schedule: R_n = base + floor(B0 * r^n) for n in [0, N), r = 9/10, N = 40. Two parameterizations were tested.

Variant A, integer B0 (the implemented definition):
- Base 40%: base 10,000, B0 = 60,900. R_0 = 70,900, R_39 = 11,000. S_actual = 999,978. Dust vs the 1,000,000 normalization = 22. Analytic S_target for these parameters = 999,998.44 (the normalization to 1,000,000 is approximate by 1.56 units).
- Base 0%: B0 = 101,500. R_0 = 101,500, R_39 = 1,666. S_actual = 999,978. Dust = 22.

Variant B, exact rational B0 (the source of the 18 figure):
- Base 40%: B0 = 600,000 / ((1 - r^N) / (1 - r)) = about 60,900.158. S_actual = 999,982. Dust = 18. S_target is exactly 1,000,000 by construction.
- Base 0%: B0 = about 101,500.264. S_actual = 999,982. Dust = 18.

Both variants are arithmetically correct under their own B0. They differ only in the B0 convention. The 18 figure is withdrawn because a rational B0 is not an implementable genesis parameter: B0 must be an integer uogo amount, and no other document uses the rational variant. The Hybrid Terminal report's table stated S_target = 1,000,000 for both hybrid rows without stating B0, so a reader using the corpus definition (integer B0) gets 999,978, not the table's 999,982.

### Ground truth

- The prototype Go code (`x/ogocoin/models.go`, `HybridSchedule` with integer `BonusInitial`) implements variant A, and its tests (`models_test.go`) assert total 999,978 and dust 22. The code is the executable ground truth.
- The treasury specification section 4 uses sum 999,978, dust 22.
- The Decision Preparation report's treasury inflow figures (99,982 hybrid, 99,947 geometric) equal the sum of floored per-period 10% shares under variant A (variant B gives 99,984).
- The monetary engine specification section 5.2 states the hybrid research figures with B0 = 60,900 (R0 70,900) and B0 = 101,500 for base 0% (R0 101,500).

### Files corrected

- `ogo-network-hybrid-terminal-and-treasury-v0.1.md`: both hybrid table rows (999,982 to 999,978, 18 to 22, 0.0018% to 0.0022%), the abrupt-end total in section 2 (999,982 to 999,978), and the comparison table (abrupt end 999,982 to 999,978). A revision note was added at the top. The taper-variant totals in section 2 were computed from the pre-correction 999,982 base; under the corrected base they shift by a few units, and the taper conclusions are unaffected. The per-period figures in section 2 (12,581 at period 30, 11,524 at period 35, 11,000 at period 39) are identical under both variants and are unchanged.
- `ogo-network-monetary-architecture-decision-preparation-v0.1.md`: section 6.2, hybrid dust 18 to 22, with a dated correction note in section 0.

### Two related notes, not corrections

- The deep-dive and several other documents state dust as "45 to 118 uogo per 1,000,000, depending on the model". That range covers the geometric models only (r = 0.85, 0.90, 0.95 give dust 45, 65, 118). Hybrid dust of 22 sits outside it. The range statements were left in place because they describe the geometric simulations in their own context; future editions should state the full range as 22 to 118.
- The deep-dive's hybrid treasury series (cumulative 73,488 by period 20) matches the rational-B0 variant exactly and the integer-B0 variant to within 2 units. The difference is below display precision for the series and does not affect any conclusion.

## 2. Low-fee R-F binding delta: 4 vs 0 units

### Resolution: 0 is correct. 4 is withdrawn.

Under Low fees, raising desired spend D from 1,500 to 2,500 per period changes actual R-F spending by exactly 0 units over 120 periods (hybrid model). Both runs spend 114,511 in total (114,511.16 before rounding).

### The recomputation

The model is the treasury specification section 4 scenario model (script `/tmp/treasury_v03.py`, no longer on disk; reconstructed here from the specification):

- Horizon 120 periods, t = 0 to 119. Hybrid issuance, integer B0 = 60,900 (variant A above). Treasury issuance inflow is 10% of issuance per period. Fee revenue follows an S-curve ramp to terminal, time constant 20 periods, starting at genesis: fee(t) = terminal * (1 - exp(-t/20)). Terminal Low = 2,462.5 (exact 0.1 times the 24,625 standard-cost minimum need; the specification prints the rounded 2,462). Treasury fee share is 10% of fee revenue. Reserve floor F = 10,000.
- R-F rule: spend_t = min(D, max(Trail_4(t), floor(0.05 * B)), max(0, B - F)), where B is the treasury balance after the period inflow settles, and Trail_4(t) is the trailing 4-period average treasury inflow including the current period. The reconstruction was validated against the specification tables: it reproduces the section 5 and 6 totals to rounding, including 114,511 at both D values, 114,507 geometric, 178,500 under Moderate and High fees, and 98,582 for the trailing-only rule. Residuals on individual balance figures are 0 to 8 units, consistent with display rounding in the original tables.
- Results (Low fees, hybrid, 120 periods): D = 1,500 spends 114,511.16; D = 2,500 spends 114,511.16; delta 0.000000. Balance at issuance end: 46,981.98 at D = 1,500; 27,025.91 at D = 2,500. Final balance: exactly 10,000.00 in both runs.

### Why the delta must be 0

By accounting identity, total spent = total inflow - final balance + initial balance. Both runs share the same inflow series (inflow does not depend on D) and both end at exactly the floor (10,000). So their totals are identical by construction. A nonzero delta would require the two runs to end at different final balances, which does not happen under the stated rule: the room term pins the balance to the floor once inflow falls below the cap. The delta is 0 in every variant of the reconstruction tested (pre-inflow and post-inflow timing, trailing window including or excluding the current period, floored or float issuance share, fee terminal 2,462 or 2,462.5, integer or rational B0 schedule).

### Where the 4 came from

Unknown, and the attribution is not verifiable: the Decision Preparation report contains no 4-unit figure and no 114,511 figure (verified by grep), so the decision register's claim that "The Decision Preparation report states..." the 4-unit delta is itself incorrect. Candidates for the origin are a different model run (the Decision Preparation analysis uses horizon 360, floor 20,000, and a different fee model) or a comparison between model variants (for example, floored vs float issuance share totals differ by about 5 units). Neither candidate was reconstructable from the surviving material. The 4-unit figure is therefore withdrawn without a confirmed origin, and this note records that the origin is unknown rather than inventing one.

### Files corrected

- `OGO_TREASURY_AND_SUSTAINABILITY.md` (specification): section 2.1 finding paragraph and the section 5 revision note rewritten to state 0 units, with the recomputation conventions recorded. The scenario tables already showed 114,511 at both D values and needed no change.
- `OGO_MONETARY_ENGINE_SPEC.md`: revision note corrected to 0 units.
- `OGO_NETWORK_DECISION_REGISTER.md`: the section 3 rationale cell corrected to 0 units; the section 8 open question ("4 units or 0 units?") marked Resolved 2026-11-06.
- `OGO_NETWORK_PROTOTYPE_TEST_REQUIREMENTS.md`: the open item marked resolved.
- `ogo-network-native-coin-architecture-implementation-plan-v1.1.md`: revision log item 5, the section on finding b, and the decision register copy corrected to 0 units; revision log item 11 added for the GitHub publication. Plan v1.0 was not edited (kept intact as the reviewed version).

### Not changed

- The first monthly OGO Exchange Readiness Report flags both discrepancies for reconciliation; it is a published first edition and is not silently rewritten. Its next monthly edition should pick up these resolutions.
- The whitepaper draft (`ogo-network-whitepaper-draft-v0.1.md`) states the 4-unit figure and is not published with the network scope. It should be corrected in its next revision.
- The substance is unchanged in both resolutions: the spending rule cannot fix a low-inflow level problem, and dust stays permanently unissued by rule.

## 3. What these resolutions do not decide

Neither resolution selects a monetary model, a schedule parameter, a split, a fee policy, or a maximum supply. The hybrid dust figure and the R-F delta are research arithmetic on illustrative normalizations, not decisions. The monetary decision stays with Benross.
