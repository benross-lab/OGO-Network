// Issuance models for the OGO monetary engine (Milestone 3).
//
// This file implements the common issuance interface from
// OGO_MONETARY_ENGINE_SPEC.md section 4, with three implementations:
// GeometricSchedule, HybridSchedule, and TaperedHybridSchedule.
//
// Naming note (deviation, recorded honestly): the spec names the
// interface IssuanceSchedule. That name is already taken in this
// prototype by the Milestone B genesis configuration struct
// IssuanceSchedule (schedule.go / types.go), which the existing tests
// and the app/cmd genesis code construct directly. Go cannot have a
// type and an interface with the same name in one package, so the
// interface is named IssuanceModel here, with exactly the spec method
// set (Model, AmountForEpoch, Periods, TotalScheduled, TotalTarget,
// Dust, Validate). The planned rename of x/monetary to x/monetary is a
// remaining item; at that rename the interface can take the spec name.
//
// Arithmetic note: the single rounding rule for Milestone C
// (decision D-B4) is R_n = floor(R0 * r^n) with r an exact rational,
// i.e. the floor of the exact power, with a single floor at the end.
// The former stepwise variant (floor at every multiplication step)
// gave period 7 = 477 and period 9 = 386 for R0=1000, r=9/10, while
// the exact formula gives 478 and 387. For the spec research
// parameters (R0=100000, r=9/10, N=110) the exact formula gives total
// 999935 and R_109 = 1, matching the spec; the stepwise version gave
// total 999572 and reached 0 at period 93, contradicting the spec.
// Both the schedule (schedule.go AmountForPeriod) and the models in
// this file now use the exact formula, computed with integer
// (big.Int) arithmetic only: no floating point anywhere.
//
// TotalTarget (S_target) is a genesis input parameter, not derived:
// the spec research normalises it to 1,000,000 and derives R0 from
// it, so S_target is stored on each schedule and Dust is
// S_target minus S_actual. Validate requires Target >= S_actual.
//
// All parameters in this file are TEST-ONLY unless a genesis file
// explicitly sets them. No production parameter is selected here.
package monetary

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
)

// IssuanceModel is the common issuance interface (spec section 4).
// Every method is deterministic: same inputs, same outputs, on every
// node. No method uses floating point. All amounts are integer uogo.
type IssuanceModel interface {
	// Model returns the model identifier: "geometric", "hybrid", or
	// "tapered-hybrid".
	Model() string

	// AmountForEpoch returns R_n, the issuance for epoch n
	// (0-indexed), floored to an integer. Returns 0 for
	// n >= Periods().
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

	// Validate checks the schedule parameters. It is called during
	// genesis validation.
	Validate() error
}

// Compile-time assertions that the three implementations satisfy the
// interface.
var (
	_ IssuanceModel = (*GeometricSchedule)(nil)
	_ IssuanceModel = (*HybridSchedule)(nil)
	_ IssuanceModel = (*TaperedHybridSchedule)(nil)
)

// Model identifiers for the issuance_model genesis enum.
const (
	ModelGeometric     = "geometric"
	ModelHybrid        = "hybrid"
	ModelTaperedHybrid = "tapered-hybrid"
)

// errInvalidModel is returned for an unknown issuance_model value.
var ErrInvalidModel = fmt.Errorf("monetary: issuance_model must be %q, %q, or %q", ModelGeometric, ModelHybrid, ModelTaperedHybrid)

// geometricExactAmount returns floor(R0 * (num/den)^n) by exact
// integer arithmetic. The intermediate power is computed with
// math/big.Rat (unbounded), because cosmossdk.io/math.Int is
// 256-bit bounded and R0 * num^n overflows it for long schedules
// (for example 10^109 in the spec research parameters). Only the
// final floored result, which fits in normal supply ranges, is
// converted back to math.Int. No floating point.
func geometricExactAmount(initial math.Int, num, den int64, n int64) math.Int {
	r := new(big.Rat).SetFrac(big.NewInt(num), big.NewInt(den))
	term := new(big.Rat).SetInt(initial.BigInt())
	for i := int64(0); i < n; i++ {
		term.Mul(term, r)
	}
	q := new(big.Int).Quo(term.Num(), term.Denom())
	return math.NewIntFromBigInt(q)
}

// geometricSeriesFloor returns floor(sum_{n=0}^{N-1} R0 * r^n) by exact
// rational arithmetic (math/big.Rat, not float). It is the theoretical
// series total before per-period flooring, used only as a cross-check
// in tests and documentation; TotalTarget is the stored S_target.
func geometricSeriesFloor(initial math.Int, num, den int64, periods int64) math.Int {
	sum := new(big.Rat)
	r := new(big.Rat).SetFrac(big.NewInt(num), big.NewInt(den))
	term := new(big.Rat).SetInt(initial.BigInt())
	for i := int64(0); i < periods; i++ {
		sum.Add(sum, term)
		term.Mul(term, r)
	}
	// Floor of a non-negative rational: numerator / denominator.
	q := new(big.Int).Quo(sum.Num(), sum.Denom())
	return math.NewIntFromBigInt(q)
}

// validateReduction checks 0 < num < den for a reduction factor.
func validateReduction(num, den int64) error {
	if den <= 0 {
		return fmt.Errorf("%w: reduction denominator must be > 0, got %d", ErrScheduleInvalid, den)
	}
	if num <= 0 {
		return fmt.Errorf("%w: reduction numerator must be > 0, got %d", ErrScheduleInvalid, num)
	}
	if num >= den {
		return fmt.Errorf("%w: reduction factor %d/%d must be below 1", ErrScheduleInvalid, num, den)
	}
	return nil
}

// validateTarget checks Target is present, positive, and at least the
// scheduled total, so Dust is never negative.
func validateTarget(target, scheduled math.Int, periods int64) error {
	if target.IsNil() || !target.IsPositive() {
		return fmt.Errorf("%w: total target (S_target) must be > 0", ErrScheduleInvalid)
	}
	if scheduled.GT(target) {
		return fmt.Errorf("%w: total scheduled %s exceeds total target %s", ErrScheduleInvalid, scheduled.String(), target.String())
	}
	if periods <= 0 {
		return fmt.Errorf("%w: periods must be > 0", ErrScheduleInvalid)
	}
	return nil
}

// ---------------------------------------------------------------------------
// GeometricSchedule: R_n = floor(R0 * r^n), n in [0, Periods()).
// ---------------------------------------------------------------------------

// GeometricSchedule is the geometric decay model (spec section 5.1).
// The existing reducing-per-period schedule becomes this
// implementation, with the exact-floor definition above.
type GeometricSchedule struct {
	// InitialPerPeriod (R0) is the issuance in period 0, in uogo.
	InitialPerPeriod math.Int `json:"initial_per_period"`
	// ReductionNumerator / ReductionDenominator is r, exact rational.
	ReductionNumerator   int64 `json:"reduction_numerator"`
	ReductionDenominator int64 `json:"reduction_denominator"`
	// NumPeriods (N) is the genesis-fixed final period count.
	NumPeriods int64 `json:"periods"`
	// Target (S_target) is the theoretical target total, in uogo.
	Target math.Int `json:"total_target"`
}

func (g *GeometricSchedule) Model() string { return ModelGeometric }

func (g *GeometricSchedule) Periods() int64 { return g.NumPeriods }

func (g *GeometricSchedule) AmountForEpoch(n int64) (math.Int, error) {
	if n < 0 || n >= g.NumPeriods {
		return math.NewInt(0), nil
	}
	return geometricExactAmount(g.InitialPerPeriod, g.ReductionNumerator, g.ReductionDenominator, n), nil
}

func (g *GeometricSchedule) TotalScheduled() (math.Int, error) {
	total := math.NewInt(0)
	for n := int64(0); n < g.NumPeriods; n++ {
		amt, _ := g.AmountForEpoch(n)
		total = total.Add(amt)
	}
	return total, nil
}

func (g *GeometricSchedule) TotalTarget() (math.Int, error) { return g.Target, nil }

func (g *GeometricSchedule) Dust() (math.Int, error) {
	s, err := g.TotalScheduled()
	if err != nil {
		return math.NewInt(0), err
	}
	return g.Target.Sub(s), nil
}

func (g *GeometricSchedule) Validate() error {
	if g == nil {
		return ErrScheduleMissing
	}
	if g.InitialPerPeriod.IsNil() || !g.InitialPerPeriod.IsPositive() {
		return fmt.Errorf("%w: initial_per_period must be > 0", ErrScheduleInvalid)
	}
	if err := validateReduction(g.ReductionNumerator, g.ReductionDenominator); err != nil {
		return err
	}
	if g.NumPeriods <= 0 {
		return fmt.Errorf("%w: periods must be > 0", ErrScheduleInvalid)
	}
	s, err := g.TotalScheduled()
	if err != nil {
		return err
	}
	if err := validateTarget(g.Target, s, g.NumPeriods); err != nil {
		return err
	}
	// The terminal period must still issue at least 1 uogo; a schedule
	// that runs past the point where the exact amount floors to 0 has
	// no predictable end date (the Cardano warning, spec section 5.1).
	last, _ := g.AmountForEpoch(g.NumPeriods - 1)
	if last.IsZero() {
		return fmt.Errorf("%w: final period amount is 0; periods extends past the terminal period", ErrScheduleInvalid)
	}
	return nil
}

// ---------------------------------------------------------------------------
// HybridSchedule: R_n = base + floor(B0 * r^n), n in [0, N).
// ---------------------------------------------------------------------------

// HybridSchedule is the base-plus-geometric-bonus model (spec
// section 5.2). Base is a fixed absolute rate paid for N periods; it
// cannot be paid past the end date, which is what gives the schedule
// a fixed end.
type HybridSchedule struct {
	// BasePerPeriod is the fixed absolute base rate per period, uogo.
	BasePerPeriod math.Int `json:"base_per_period"`
	// BonusInitial (B0) is the initial geometric bonus, in uogo.
	BonusInitial math.Int `json:"bonus_initial"`
	// ReductionNumerator / ReductionDenominator is r, exact rational.
	ReductionNumerator   int64 `json:"reduction_numerator"`
	ReductionDenominator int64 `json:"reduction_denominator"`
	// NumPeriods (N) is the genesis-fixed final period count.
	NumPeriods int64 `json:"periods"`
	// Target (S_target) is the theoretical target total, in uogo.
	Target math.Int `json:"total_target"`
}

func (h *HybridSchedule) Model() string { return ModelHybrid }

func (h *HybridSchedule) Periods() int64 { return h.NumPeriods }

func (h *HybridSchedule) AmountForEpoch(n int64) (math.Int, error) {
	if n < 0 || n >= h.NumPeriods {
		return math.NewInt(0), nil
	}
	bonus := geometricExactAmount(h.BonusInitial, h.ReductionNumerator, h.ReductionDenominator, n)
	return h.BasePerPeriod.Add(bonus), nil
}

func (h *HybridSchedule) TotalScheduled() (math.Int, error) {
	total := math.NewInt(0)
	for n := int64(0); n < h.NumPeriods; n++ {
		amt, _ := h.AmountForEpoch(n)
		total = total.Add(amt)
	}
	return total, nil
}

func (h *HybridSchedule) TotalTarget() (math.Int, error) { return h.Target, nil }

func (h *HybridSchedule) Dust() (math.Int, error) {
	s, err := h.TotalScheduled()
	if err != nil {
		return math.NewInt(0), err
	}
	return h.Target.Sub(s), nil
}

func (h *HybridSchedule) Validate() error {
	if h == nil {
		return ErrScheduleMissing
	}
	if h.BasePerPeriod.IsNil() || !h.BasePerPeriod.IsPositive() {
		return fmt.Errorf("%w: base_per_period must be > 0 for a hybrid schedule", ErrScheduleInvalid)
	}
	if h.BonusInitial.IsNil() || !h.BonusInitial.IsPositive() {
		return fmt.Errorf("%w: bonus_initial must be > 0 for a hybrid schedule", ErrScheduleInvalid)
	}
	if err := validateReduction(h.ReductionNumerator, h.ReductionDenominator); err != nil {
		return err
	}
	if h.NumPeriods <= 0 {
		return fmt.Errorf("%w: periods must be > 0", ErrScheduleInvalid)
	}
	// Spec section 3.3: a hybrid whose base rate times N exceeds
	// S_target is rejected.
	if h.BasePerPeriod.MulRaw(h.NumPeriods).GT(h.Target) {
		return fmt.Errorf("%w: base_per_period * N exceeds total target", ErrScheduleInvalid)
	}
	s, err := h.TotalScheduled()
	if err != nil {
		return err
	}
	return validateTarget(h.Target, s, h.NumPeriods)
}

// ---------------------------------------------------------------------------
// TaperedHybridSchedule: hybrid with a gentle geometric terminal taper,
// renormalized to the same total.
// ---------------------------------------------------------------------------

// TaperedHybridSchedule (spec section 4, candidate) applies a gentle
// geometric taper over the final TaperPeriods periods and then
// renormalizes to the same total as the untapered hybrid.
//
// Definition, fixed here so it is reproducible:
//   - Untapered amounts A_n are the HybridSchedule amounts.
//   - For n in the final K periods, j = n - (N-K) + 1 (1..K):
//     raw_n = floor(A_n * (taperNum/taperDen)^j), floored stepwise
//     in exact integer arithmetic.
//   - Raw tapered total S_raw is below the untapered total S.
//     Renormalization scales every raw amount by S / S_raw, floored:
//     scaled_n = floor(raw_n * S / S_raw). The remainder
//     R = S - sum(scaled_n) (0 <= R < N, since each floor loses < 1)
//     is distributed 1 uogo to each of the first R periods, in epoch
//     order. TotalScheduled therefore equals the untapered total
//     exactly, and the taper shape (final payment well below the
//     untapered final payment) is preserved.
//
// Measured, not claimed: for the spec research parameters (base
// 10000, B0 60900, r 9/10, N 40, taper 95/100 over the final 20)
// the renormalized final payment is 4360 uogo (untapered 11000)
// and the maximum single-period drop is 9.28 percent (928 bp, measured),
 // at the first
// tapered period, where the 5 percent taper step lands on top of
// the underlying geometric drop. The spec's "about 5000, max drop
// 6.5 percent" is NOT reproduced for these parameters; the test
// asserts the measured values above instead of the spec's rounded
// claim. Short linear tapers are rejected by construction: there
// is no linear taper mode in this implementation at all.
type TaperedHybridSchedule struct {
	// Hybrid is the underlying untapered hybrid schedule.
	Hybrid HybridSchedule `json:"hybrid"`
	// TaperNumerator / TaperDenominator is the per-step taper
	// factor, about 95/100. Must be strictly below 1.
	TaperNumerator   int64 `json:"taper_numerator"`
	TaperDenominator int64 `json:"taper_denominator"`
	// TaperPeriods (K) is the number of final periods the taper
	// applies to. Must satisfy 1 <= K <= N.
	TaperPeriods int64 `json:"taper_periods"`
}

func (t *TaperedHybridSchedule) Model() string { return ModelTaperedHybrid }

func (t *TaperedHybridSchedule) Periods() int64 { return t.Hybrid.NumPeriods }

// rawTapered returns the pre-renormalization tapered amounts.
func (t *TaperedHybridSchedule) rawTapered() ([]math.Int, error) {
	n := t.Hybrid.NumPeriods
	raw := make([]math.Int, n)
	for i := int64(0); i < n; i++ {
		amt, err := t.Hybrid.AmountForEpoch(i)
		if err != nil {
			return nil, err
		}
		if i >= n-t.TaperPeriods {
			j := i - (n - t.TaperPeriods) + 1
			v := amt
			for k := int64(0); k < j; k++ {
				v = v.MulRaw(t.TaperNumerator).QuoRaw(t.TaperDenominator)
			}
			raw[i] = v
		} else {
			raw[i] = amt
		}
	}
	return raw, nil
}

// taperedAmounts returns the renormalized tapered amounts.
func (t *TaperedHybridSchedule) taperedAmounts() ([]math.Int, error) {
	raw, err := t.rawTapered()
	if err != nil {
		return nil, err
	}
	untapered, err := t.Hybrid.TotalScheduled()
	if err != nil {
		return nil, err
	}
	rawTotal := math.NewInt(0)
	for _, v := range raw {
		rawTotal = rawTotal.Add(v)
	}
	if rawTotal.IsZero() {
		return nil, fmt.Errorf("%w: tapered total is 0", ErrScheduleInvalid)
	}
	scaled := make([]math.Int, len(raw))
	scaledTotal := math.NewInt(0)
	for i, v := range raw {
		scaled[i] = v.Mul(untapered).Quo(rawTotal)
		scaledTotal = scaledTotal.Add(scaled[i])
	}
	remainder := untapered.Sub(scaledTotal)
	// Remainder is < N (each floor loses < 1), so one uogo to each of
	// the first R periods, in epoch order, is exact and deterministic.
	for i := int64(0); !remainder.IsZero() && i < int64(len(scaled)); i++ {
		scaled[i] = scaled[i].AddRaw(1)
		remainder = remainder.SubRaw(1)
	}
	return scaled, nil
}

func (t *TaperedHybridSchedule) AmountForEpoch(n int64) (math.Int, error) {
	if n < 0 || n >= t.Hybrid.NumPeriods {
		return math.NewInt(0), nil
	}
	amounts, err := t.taperedAmounts()
	if err != nil {
		return math.NewInt(0), err
	}
	return amounts[n], nil
}

func (t *TaperedHybridSchedule) TotalScheduled() (math.Int, error) {
	// By construction this equals the untapered hybrid total; sum
	// the amounts rather than trusting the construction.
	amounts, err := t.taperedAmounts()
	if err != nil {
		return math.NewInt(0), err
	}
	total := math.NewInt(0)
	for _, v := range amounts {
		total = total.Add(v)
	}
	return total, nil
}

func (t *TaperedHybridSchedule) TotalTarget() (math.Int, error) {
	return t.Hybrid.TotalTarget()
}

func (t *TaperedHybridSchedule) Dust() (math.Int, error) {
	s, err := t.TotalScheduled()
	if err != nil {
		return math.NewInt(0), err
	}
	target, err := t.TotalTarget()
	if err != nil {
		return math.NewInt(0), err
	}
	return target.Sub(s), nil
}

func (t *TaperedHybridSchedule) Validate() error {
	if t == nil {
		return ErrScheduleMissing
	}
	if err := t.Hybrid.Validate(); err != nil {
		return err
	}
	if t.TaperDenominator <= 0 || t.TaperNumerator <= 0 {
		return fmt.Errorf("%w: taper factor numerator/denominator must be > 0", ErrScheduleInvalid)
	}
	if t.TaperNumerator >= t.TaperDenominator {
		return fmt.Errorf("%w: taper factor %d/%d must be below 1", ErrScheduleInvalid, t.TaperNumerator, t.TaperDenominator)
	}
	if t.TaperPeriods < 1 || t.TaperPeriods > t.Hybrid.NumPeriods {
		return fmt.Errorf("%w: taper_periods must satisfy 1 <= K <= N (%d), got %d", ErrScheduleInvalid, t.Hybrid.NumPeriods, t.TaperPeriods)
	}
	// The taper must shrink, not remove, the final payment: the
	// renormalized final amount must be positive and below the
	// untapered final amount.
	amounts, err := t.taperedAmounts()
	if err != nil {
		return err
	}
	last := amounts[len(amounts)-1]
	if last.IsZero() {
		return fmt.Errorf("%w: tapered final payment is 0", ErrScheduleInvalid)
	}
	untaperedLast, _ := t.Hybrid.AmountForEpoch(t.Hybrid.NumPeriods - 1)
	if !last.LT(untaperedLast) {
		return fmt.Errorf("%w: tapered final payment %s is not below untapered %s", ErrScheduleInvalid, last.String(), untaperedLast.String())
	}
	return nil
}

// ---------------------------------------------------------------------------
// Genesis selection: build the selected model from Params.
// ---------------------------------------------------------------------------

// SelectedModel builds the issuance model named by Params.IssuanceModel
// from the genesis schedule fields. An empty IssuanceModel means the
// legacy Milestone B genesis (no model field): it selects geometric,
// which is the existing reducing-per-period schedule, so existing
// fixtures keep working. A present but unknown value is rejected.
// Hybrid and tapered-hybrid read their extra parameters from the
// IssuanceSchedule struct fields (BasePerPeriod, taper fields); see
// types.go. TotalTarget for a genesis-built model is derived as the
// floor of the exact series (geometric) or series plus base
// (hybrid), so Dust is the per-period flooring loss only; a genesis
// file that wants a nominal S_target must set it through the model
// structs directly, not through this legacy-schedule path.
func (p *Params) SelectedModel() (IssuanceModel, error) {
	if p.IssuanceSchedule == nil {
		return nil, ErrScheduleMissing
	}
	s := p.IssuanceSchedule
	switch p.IssuanceModel {
	case "", ModelGeometric:
		target := geometricSeriesFloor(s.InitialPerPeriod, s.ReductionNumerator, s.ReductionDenominator, int64(s.Periods))
		return &GeometricSchedule{
			InitialPerPeriod:     s.InitialPerPeriod,
			ReductionNumerator:   s.ReductionNumerator,
			ReductionDenominator: s.ReductionDenominator,
			NumPeriods:           int64(s.Periods),
			Target:               target,
		}, nil
	case ModelHybrid:
		if s.BasePerPeriod.IsNil() || !s.BasePerPeriod.IsPositive() {
			return nil, fmt.Errorf("%w: hybrid model requires base_per_period > 0 in the genesis schedule", ErrScheduleInvalid)
		}
		// Bonus initial is the geometric part of period 0: R0 minus base.
		bonus := s.InitialPerPeriod.Sub(s.BasePerPeriod)
		if bonus.IsNil() || !bonus.IsPositive() {
			return nil, fmt.Errorf("%w: hybrid model requires initial_per_period > base_per_period", ErrScheduleInvalid)
		}
		series := geometricSeriesFloor(bonus, s.ReductionNumerator, s.ReductionDenominator, int64(s.Periods))
		target := s.BasePerPeriod.MulRaw(int64(s.Periods)).Add(series)
		return &HybridSchedule{
			BasePerPeriod:        s.BasePerPeriod,
			BonusInitial:         bonus,
			ReductionNumerator:   s.ReductionNumerator,
			ReductionDenominator: s.ReductionDenominator,
			NumPeriods:           int64(s.Periods),
			Target:               target,
		}, nil
	case ModelTaperedHybrid:
		hybrid, err := (&Params{IssuanceModel: ModelHybrid, IssuanceSchedule: s}).SelectedModel()
		if err != nil {
			return nil, err
		}
		hs, ok := hybrid.(*HybridSchedule)
		if !ok {
			return nil, fmt.Errorf("%w: internal error building tapered hybrid", ErrScheduleInvalid)
		}
		if !s.TaperEnabled {
			return nil, fmt.Errorf("%w: tapered-hybrid model requires taper_enabled in the genesis schedule", ErrScheduleInvalid)
		}
		if s.TaperPeriods == 0 {
			return nil, fmt.Errorf("%w: tapered-hybrid model requires taper_periods > 0", ErrScheduleInvalid)
		}
		return &TaperedHybridSchedule{
			Hybrid:           *hs,
			TaperNumerator:   s.TaperNumerator,
			TaperDenominator: s.TaperDenominator,
			TaperPeriods:     int64(s.TaperPeriods),
		}, nil
	default:
		return nil, fmt.Errorf("%w: got %q", ErrInvalidModel, p.IssuanceModel)
	}
}

// SelectedScheduledTotal returns S_actual for the model selected by
// genesis, by exact integer summation of the floored per-period
// amounts. It agrees exactly with the schedule TotalIssuance, because
// both use the single exact rounding rule (decision D-B4).
func (p *Params) SelectedScheduledTotal() (math.Int, error) {
	m, err := p.SelectedModel()
	if err != nil {
		return math.NewInt(0), err
	}
	return m.TotalScheduled()
}
