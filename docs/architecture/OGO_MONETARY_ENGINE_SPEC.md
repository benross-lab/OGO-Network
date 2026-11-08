# OGO Monetary Engine Specification

Status: Technical specification. Planning only. No blockchain deployed, no token issued, one monetary parameter selected (hard maximum supply, 100,000,000,000 OGO, approved 2026-11-08), no production change authorized.
Date: 2026-11-06
Updated: 2026-11-08 (hard maximum supply approved; see sections 2 and 3)
Companion document: `ogo-network-native-coin-architecture-implementation-plan-v1.0.md` (section 4)
Architecture direction: CONFIRMED AND LOCKED, Option B. OGO Network is an independent Layer 1 blockchain with native OGO Coin.
No model selected. One parameter selected: hard maximum supply (see sections 2 and 3). Both models are available in the simulator.

## 1. Purpose

This document specifies the monetary engine: the component that owns the supply cap, the issuance schedule, the mint operation, and supply accounting. It is the production design for the scaffold's x/ogocoin module, renamed x/monetary. The design goal is that every monetary choice (initial supply, denomination, issuance model, schedule, epoch duration, allocations, fee distribution, reserve floor, spending limits, maximum issuance, governance-controlled upgrades) is a genesis parameter, and the model is selected through genesis configuration, not by redesigning the engine.

## 2. Design principles

1. **The engine executes a schedule; it does not choose one.** The schedule is fixed in genesis. The engine computes R_n for epoch n and mints exactly that amount, once. It has no discretion.
2. **One issuance authority.** x/monetary is the only module that can mint. There is no second mint path, no admin mint, no governance mint, no contract mint, and no founder mint.
3. **Configuration, not code, selects the model.** Geometric and hybrid are two implementations of one issuance interface. The model is selected by a genesis parameter. Changing the model before genesis means changing a genesis value, not rewriting the engine.
4. **Exact integer arithmetic.** No floating point in any consensus-critical monetary code. One integer-floor rounding rule. Deterministic period indexing. Independently reproducible test vectors.
5. **Three-number accounting.** S_target (theoretical target), S_actual (actual cumulative scheduled issuance, by exact integer summation), and M (maximum supply, M = G + S_actual) are distinguished at all times. M is approved as 100,000,000,000 OGO (2026-11-08); at genesis M is verified as G + S_actual by exact integer summation, and G and S_actual remain undecided and must sum to the approved ceiling.
6. **Dust is left unissued.** Integer flooring leaves dust. Dust is never an alternative minting allowance, a rounding buffer, or a discretionary top-up.

## 3. Genesis parameters

All of the following are genesis parameters. All are TBD except max_supply, which is approved as 100,000,000,000 OGO (2026-11-08). A genesis file that omits any mandatory parameter is rejected. A genesis file that contains a default for max supply is rejected: max supply has no default; the approved ceiling must appear explicitly in genesis.

### 3.1 Mandatory parameters

| Parameter | Type | Notes |
|---|---|---|
| `max_supply` | integer (uogo) | Mandatory, no default. Approved value: 100,000,000,000 OGO (Benross, 2026-11-08). Must equal G + S_actual, computed at genesis by exact integer summation; G and S_actual remain TBD and must sum to the approved ceiling. Genesis validation rejects a max supply that does not equal G + S_actual. |
| `denom` | string | Proposed `uogo`. Not final. |
| `decimals` | integer | Proposed 6. Not final. 1 OGO equals 1,000,000 uogo under the proposal. |
| `genesis_supply` (G) | integer (uogo) | Genesis circulating supply. TBD. |
| `issuance_model` | enum | `geometric` or `hybrid`. Selects the schedule implementation. No default that favours one model; the value must be present. |
| `epoch_duration` (D) | integer (seconds) | Fixed duration of one epoch. Not selected. |
| `epoch_identifier` | string | Working name `monetary`. Fixed in genesis. |
| `first_epoch_start` | timestamp and height | Start of the first epoch. Fixed in genesis. |
| `final_issuance_period` (N) | integer | Genesis-fixed final issuance period. Issuance stops permanently after N. The treasury mode transition is keyed to N. |
| `initial_per_period_amount` (R0) | integer (uogo) | Initial per-period issuance. Derived from the model parameters and S_target, not chosen independently of them. |
| `reduction_factor` (r) | rational | Geometric reduction factor, less than 1. Stored as an exact rational (numerator and denominator), never as a float. |
| `periods` | integer | Number of periods in the schedule. For geometric, the terminal period. For hybrid, N. |
| `validator_allocation` | rational | Share of each epoch's issuance to validators. Part of the split. TBD. |
| `delegator_allocation` | rational | Share of each epoch's issuance to delegators. Part of the split. TBD. |
| `community_allocation` | rational | Share of each epoch's issuance to the community pool. Part of the split. TBD. |
| `treasury_allocation` | rational | Share of each epoch's issuance to the treasury. Part of the split. TBD. |
| `fee_distribution` | struct | Shares of transaction fees to validators, delegators, community, and treasury. TBD. |
| `treasury_spending_rule` | enum and params | Recommended default R-F (see section 7). The rule and its parameters (desired spend D, trailing window, balance cap, emergency reserve floor F) are genesis parameters. |
| `reserve_floor` (F) | integer (uogo) | Emergency reserve floor. Genesis-fixed. TBD. |
| `schedule_purpose_labels` | list of strings | Mandatory public purpose labels for the schedule. Genesis validation rejects a schedule without them, and rejects founder or reserved purposes. |

### 3.2 Hybrid-only parameters

| Parameter | Type | Notes |
|---|---|---|
| `base_rate` | integer (uogo) | Fixed absolute base rate per period (the floor). Paid for N periods. A floor as a share of remaining reserve never ends; a fixed absolute floor cannot be paid past the end date. |
| `base_share` | rational | Base rate as a share of total scheduled issuance. Used to derive the base rate and the geometric bonus. TBD. |
| `taper_enabled` | boolean | Whether the gentle geometric terminal taper applies. Candidate, not a decision. |
| `taper_factor` | rational | Gentle taper factor, about 0.95 over the final half of the schedule. Candidate. |
| `taper_periods` | integer | Number of final periods the taper applies to. Candidate, about 20. |

### 3.3 Validation rules

Genesis validation rejects:

- A missing max supply, or a max supply with a default value.
- A max supply that does not equal G + S_actual, computed by exact integer summation of the floored per-period amounts.
- A schedule with no purpose labels, or with a founder or reserved purpose label.
- A reduction factor greater than or equal to 1, or less than or equal to 0.
- A split whose shares do not sum to exactly 1, in rational arithmetic.
- A fee distribution whose shares do not sum to exactly 1.
- An epoch duration less than or equal to 0, or a final issuance period less than or equal to 0.
- A hybrid schedule whose base rate times N exceeds S_target.
- Any parameter that is not a valid integer or rational (for example, a float).

## 4. Common issuance interface

Both models implement one interface. The engine calls the interface; it does not know which model is behind it. This is what makes the model selectable through genesis configuration.

```go
// IssuanceSchedule is the common interface for all issuance models.
// Every method is deterministic: same inputs, same outputs, on every node.
// No method uses floating point. All amounts are integer uogo.
type IssuanceSchedule interface {
    // Model returns the model identifier: "geometric" or "hybrid".
    Model() string

    // AmountForEpoch returns R_n, the issuance for epoch n (0-indexed),
    // floored to an integer. Returns 0 for n >= Periods().
    AmountForEpoch(n int64) (math.Int, error)

    // Periods returns the number of periods in the schedule (N).
    Periods() int64

    // TotalScheduled returns S_actual: the exact integer sum of
    // AmountForEpoch(n) for n in [0, Periods()).
    TotalScheduled() (math.Int, error)

    // TotalTarget returns S_target: the theoretical target total.
    TotalTarget() (math.Int, error)

    // Dust returns S_target minus S_actual: the amount left unissued
    // by integer flooring. Dust is never issuable.
    Dust() (math.Int, error)

    // Validate checks the schedule parameters (see section 3.3).
    // It is called during genesis validation.
    Validate() error
}
```

Implementations:

- `GeometricSchedule`: R_n = floor(R0 * r^n), for n in [0, Periods()). R0 and r are genesis parameters. Periods is the terminal period, genesis-fixed so the end date is predictable. Without a fixed final period, geometric decay has no fixed end date.
- `HybridSchedule`: R_n = base + floor(B0 * r^n), for n in [0, N). Base is the fixed absolute base rate. B0 is the initial geometric bonus, derived from the base share and the total. The floor is a fixed absolute base rate for a fixed number of periods ending at period N.
- `TaperedHybridSchedule` (candidate): the hybrid schedule with a gentle geometric taper applied over the final periods. Taper factor about 0.95 over the final 20 periods, renormalized to the same total. Short linear tapers are rejected as worse than abrupt (drops up to 35.7 percent). The taper never removes the 100 percent end; it shrinks it.

The simulator (section 8) implements the same interface off-chain, in exact rational arithmetic, so the on-chain schedule and the simulator can be compared directly.

## 5. The two models, as implemented

No model is selected. Geometric and hybrid remain the two candidates. The figures below are from the research, normalized to total scheduled issuance 1,000,000 illustrative units. They are comparison inputs, not proposed parameters.

### 5.1 Geometric decay (primary candidate)

R_n = floor(R0 * r^n). Research figures, r = 0.90, R0 = 100,000:

- 50 percent issued by period 7, 90 percent by period 22, 99 percent by period 44.
- Maximum single-period drop 10 percent. Period-20 issuance 12.2 percent of R0.
- Terminal period 110 (R_109 = 1), dust 65.
- Sensitivity: r = 0.85 gives terminal period 74 and dust 45. r = 0.95 gives terminal period 211 and dust 118. A 0.05 change in r more than doubles the horizon. r is the most sensitive parameter in the design.
- Already implemented in the scaffold as the reducing-per-period schedule type.
- End rule recommendation (not a decision): specify a fixed final period in genesis so the end date is predictable. Without it, geometric decay has no fixed end date, which is the Cardano warning: a declining reserve with no landing point.

### 5.2 Hybrid, base plus geometric bonus (leading alternative)

R_n = base + floor(B0 * r^n), for N = 40 periods, r = 0.90. Research figures:

- Base 0 percent (truncated geometric): R0 101,500, first-10-periods share 66.1 percent, 10 percent horizon period 21.
- Base 20 percent (5,000 per period): R0 86,200, 90 percent by period 28, 10 percent horizon period 29.
- Base 40 percent (10,000 per period): R0 70,900, 50 percent by period 11, 90 percent by period 32, first-10 share 49.7 percent, issuance never below 14.1 percent of R0 while running, 10 percent horizon period 39.
- Base 60 percent (15,000 per period): R0 55,600, first-10 share 41.4 percent.
- The hybrid floor is a fixed absolute base rate for a fixed number of periods ending at period N. A floor as a share of remaining reserve never ends; a fixed absolute floor cannot be paid past the end date.
- Terminal behaviour: an abrupt end (base floor directly to zero) is a reward cliff. The final payment in the research is 11,000. A short linear taper is worse than abrupt (drops up to 35.7 percent). A gentle geometric taper (factor about 0.95 over the final 20 periods) halves the final payment to about 5,000, with a maximum drop of 6.5 percent. The taper never removes the 100 percent end; it shrinks it. The gentle taper is a candidate if hybrid is chosen, not a decision.

### 5.3 Epoch-based (legibility fallback) and halving (reference only)

- Epoch-based: four epochs of 10 periods, amounts 40,000, 30,000, 20,000, 10,000 per period. 50 percent by period 14, 100 percent at period 40. Most legible as a table. Retained as a fallback.
- Halving: 50 percent by period 4, 90 percent by period 14, 99 percent by period 27, with a 50 percent drop at every boundary. Retained as a philosophical reference, not a primary candidate for a proof-of-stake network.

## 6. Mint operation

The mint operation is the only way new OGO Coin enters supply, and it happens only through the schedule.

- Entry point: `MintScheduled(epochNumber)`, called by x/monetary from the x/epochs `AfterEpochEnd` hook. It takes no amount argument.
- Steps: check the epoch identifier; check the issued-by-epoch set (if the epoch is already claimed, return without minting); compute R_n through the issuance interface; mint exactly R_n into the rewards pool module account; record the epoch in the issued-by-epoch set with the amount and the block height; hand the minted amount to x/rewards for splitting.
- The mint fails if: the epoch is already claimed; the epoch number is greater than the current epoch; the computed amount is 0 and the epoch is not the final epoch (a zero amount outside the terminal period indicates a schedule error, and it is logged, not silently skipped); total issued plus R_n would exceed total scheduled; total supply plus R_n would exceed max supply M.
- There is no other mint entrypoint. Bank, staking, distribution, gov, and wasm cannot mint. A contract cannot mint. A governance proposal cannot mint. A founder cannot mint.

## 7. Treasury integration

The monetary engine hands the treasury its share of each epoch's issuance through x/rewards, per the approved split. The treasury itself is owned by x/treasury (see `OGO_TREASURY_AND_SUSTAINABILITY.md`). The engine's responsibilities to the treasury are:

- Deliver the treasury share of R_n, exactly, in integer arithmetic, every epoch, with the split conserving (the parts sum to R_n).
- Expose the treasury share, total treasury inflow, and total issued as queries.
- Never mint directly into the treasury. The treasury receives its share through x/rewards, like every other recipient.
- The treasury spending rule (R-F, recommended default) and its parameters (desired spend D, trailing window, balance cap, emergency reserve floor F) are genesis parameters of x/treasury, not of x/monetary. The engine does not enforce the spending rule; it supplies the inflow the rule operates on.

R-F rule, for reference: spend = min(D, max(Trail_4, floor(0.05 * B)), max(0, B - F)), where B is the treasury balance at the end of the previous period, Trail_4 is the trailing four-period average treasury inflow, and F is the emergency reserve floor. Transition from issuance mode to post-issuance mode is keyed to the genesis-fixed final issuance period N.

Revision note, 2026-11-06, corrected 2026-11-06 (GitHub publication): desired spending D and actual permitted spending are separate numbers. The R-F rule rations a shortfall, it does not create inflow, and a spending rule cannot fix a level problem: under Low fees, raising D from 1,500 to 2,500 changes actual spending by 0 units over 120 periods: both runs spend 114,511. An earlier version of this note stated 4 units; that figure is withdrawn as not reproducible from the stated model. The full separation and the six-figure treasury reporting requirement are specified in `OGO_TREASURY_AND_SUSTAINABILITY.md` section 2.1, which governs. This engine supplies the inflow the rule operates on; it does not certify sustainability.

## 8. Simulator

The simulator is an off-chain tool that implements the same issuance interface as the chain, in exact rational arithmetic, so model choices can be compared before genesis without touching the chain.

- Both models are available in the simulator: geometric and hybrid, with the optional gentle terminal taper. The model is selected by a configuration value, the same way genesis selects it.
- All parameters are configurable: initial supply, denomination, issuance model, schedule parameters, epoch duration, validator, community, and treasury allocations, fee distribution, reserve floor, treasury spending limits, maximum issuance, and governance-controlled upgrades. The simulator reads the same parameter names as the genesis file, so a simulator run and a genesis file can be compared directly.
- The simulator computes, for each model and parameter set: the per-period amounts R_n, S_target, S_actual, dust, M = G + S_actual, the cumulative issuance curve, the 50 percent, 90 percent, and 99 percent horizons, the maximum single-period drop, the final payment, the treasury inflow curve, and the fee-transition requirement (what share of period-0 validator pay fees must replace by each period).
- The simulator uses exact rational arithmetic (fractions, floored to integers), the same as the chain. It does not use floating point for schedule computation. A simulator result that disagrees with the chain's computation for the same parameters is a bug, and the chain wins until the bug is found, because the chain is the implementation and the simulator is the model.
- Test vectors: the simulator generates test vectors (parameter set, per-period amounts, totals, dust) that are checked into the repository and used as golden tests for x/monetary. If the chain's computation ever diverges from the test vectors, the tests fail.
- The simulator is not consensus code. It is a design tool. Its outputs are comparison inputs, not selected parameters. No simulator output is a decision.

## 9. Determinism and exact arithmetic

- No floating point in any consensus-critical monetary code. The float summation order changes the total (forward 999990.7386128694 vs reverse 999990.7386128692 for the same series in the research), and 0.9 is not exactly representable in binary. Float per-period floors happened to agree with exact floors in all 110 periods in one tested case; that was luck, not a guarantee.
- One integer-floor rounding rule, applied everywhere: floor to the integer, always, in the same place in the computation.
- Deterministic period indexing: period n is epoch n, 0-indexed. Period n covers the epoch whose number is n. There is no off-by-one between the schedule and the epoch number.
- Deterministic cumulative calculations: cumulative issued is the sum of the floored per-period amounts, computed in epoch order, not the floor of a cumulative float.
- Independently reproducible test vectors: for each model and parameter set, the simulator generates vectors that any third party can recompute from the genesis parameters alone, without trusting OGO Systems.

## 10. Supply invariants and minimum test list

Eleven invariants, each with a minimum test. These are the acceptance tests for x/monetary. They must pass on the wired chain, not just on the scaffold.

1. Total supply never exceeds max supply M. Test: attempt to mint above the cap; mint fails, supply unchanged.
2. Total supply equals genesis supply plus total scheduled issued minus total burned, exactly, in integer arithmetic. Test: after a transfer and a burn, query total supply and compare to the computed value.
3. Scheduled issuance for epoch n equals exactly R_n. Test: for each epoch in a test schedule, claim and compare.
4. Each epoch is claimed exactly once. Test: re-claim the same epoch; second claim fails, issued total unchanged.
5. Total issued never exceeds total scheduled. Test: claim all epochs, then attempt one more; it fails.
6. Dust is never issued. Test: after all epochs, total issued equals S_actual, and the difference S_target minus S_actual is exactly the dust, which remains unissued.
7. No discretionary mint exists. Test: attempt to mint through bank, staking, distribution, gov, and wasm; all fail, state unchanged.
8. Governance cannot change max supply, the schedule, or the final issuance period. Test: submit a governance proposal to change each; each is rejected or has no effect, and the schedule query returns the original values.
9. IBC escrow and voucher accounting conserves global economic supply. Test: transfer out, verify escrow equals transferred amount and voucher equals transferred amount; transfer back, verify escrow released exactly.
10. The split conserves. Test: for a given epoch, validator plus delegator plus community plus treasury shares equal exactly the amount minted, with no remainder and no shortfall, in integer arithmetic.
11. The every-block invariant holds. Test: at every block, the invariant check passes; if it fails, the block is rejected or the chain halts, per the configured behaviour, and the failure is logged.

## 11. Consensus-critical vs governance-controlled

Consensus-critical rules (deterministic, identical across nodes, changeable only through the constitutional process, never by ordinary governance): denomination and decimal precision; issuance arithmetic (schedule formula, reduction factor, floor, period count, per-period amounts, integer-floor rounding rule, deterministic period indexing, deterministic cumulative calculations); supply limits (max supply M, cap enforcement on every mint, issued-by-epoch set, permanent issuance cessation at the cap); fee calculation formula; validator reward calculation for a given split; transaction validity (signature verification, sequence checking, balance checking).

Governance-controlled rules (explicit permissions, voting thresholds, safeguards, changeable through x/gov): approved upgrades (through x/upgrade, subject to the denylist gate); treasury spending proposals (within the R-F rule); selected network parameters (the validator, delegator, community, and treasury split; the fee level and fee distribution; governance parameters themselves). Each must have an explicit permission, a voting threshold, and a safeguard (a maximum change per proposal, a timelock, or both) documented before genesis.

## 12. What this document does not decide

No model is selected. One parameter is selected: hard maximum supply, 100,000,000,000 OGO (approved 2026-11-08). The decision order is preserved for the remaining decisions: (1) select model family, (2) schedule parameters, (3) validator, delegator, community, and treasury split, (4) fees and long-term security analysis, (5) choose G and S_actual so that G + S_actual = 100,000,000,000 OGO. The legal analysis (securities, money transmission, tax, international distribution, exchange admission) must be commissioned in writing before genesis distribution. It is the longest-lead item and is not yet commissioned.
