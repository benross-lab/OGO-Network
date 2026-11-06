package ogocoin

import (
	"fmt"

	"cosmossdk.io/math"
)

// ScheduleTypeReducingPerPeriod is the only supported schedule type in
// this framework.
const ScheduleTypeReducingPerPeriod = "reducing-per-period"

// Validate checks a schedule's structure. All VALUES are TBD in
// production; this validates the shape only. A missing schedule is a
// separate error (ErrScheduleMissing) raised by ValidateGenesis.
func (s *IssuanceSchedule) Validate() error {
	if s == nil {
		return ErrScheduleMissing
	}
	if s.ScheduleType != ScheduleTypeReducingPerPeriod {
		return fmt.Errorf("%w: schedule_type must be %q, got %q", ErrScheduleInvalid, ScheduleTypeReducingPerPeriod, s.ScheduleType)
	}
	if s.StartHeight < 0 {
		return fmt.Errorf("%w: start_height must be >= 0, got %d", ErrScheduleInvalid, s.StartHeight)
	}
	if s.PeriodLength <= 0 {
		return fmt.Errorf("%w: period_length_blocks must be > 0, got %d", ErrScheduleInvalid, s.PeriodLength)
	}
	if s.Periods == 0 {
		return fmt.Errorf("%w: periods must be > 0", ErrScheduleInvalid)
	}
	if s.InitialPerPeriod.IsNil() || !s.InitialPerPeriod.IsPositive() {
		return fmt.Errorf("%w: initial_per_period must be > 0", ErrScheduleInvalid)
	}
	if s.ReductionDenominator <= 0 {
		return fmt.Errorf("%w: reduction_denominator must be > 0", ErrScheduleInvalid)
	}
	if s.ReductionNumerator <= 0 {
		return fmt.Errorf("%w: reduction_numerator must be > 0", ErrScheduleInvalid)
	}
	// Issuance must gradually decrease (Constitution §1): the reduction
	// factor must be strictly below 1.
	if s.ReductionNumerator >= s.ReductionDenominator {
		return fmt.Errorf("%w: reduction factor %d/%d must be below 1 (issuance must decrease)", ErrScheduleInvalid, s.ReductionNumerator, s.ReductionDenominator)
	}
	return nil
}

// AmountForPeriod returns the exact issuance amount for period p:
// InitialPerPeriod * (num/den)^p, with integer floor division at each
// step. It is a pure function of the schedule: deterministic, no state.
func (s *IssuanceSchedule) AmountForPeriod(p uint64) math.Int {
	amt := s.InitialPerPeriod
	for i := uint64(0); i < p; i++ {
		amt = amt.MulRaw(s.ReductionNumerator).QuoRaw(s.ReductionDenominator)
		if amt.IsZero() {
			return math.NewInt(0)
		}
	}
	return amt
}

// TotalIssuance returns the sum of AmountForPeriod over all periods.
// Genesis validation requires genesis supply + TotalIssuance <= max.
func (s *IssuanceSchedule) TotalIssuance() math.Int {
	total := math.NewInt(0)
	for p := uint64(0); p < s.Periods; p++ {
		total = total.Add(s.AmountForPeriod(p))
	}
	return total
}

// PeriodForHeight maps a block height to its issuance period. It reports
// false when the height is before StartHeight (not started) or at/past
// the end of the last period (exhausted).
func (s *IssuanceSchedule) PeriodForHeight(height int64) (uint64, bool) {
	if height < s.StartHeight {
		return 0, false
	}
	elapsed := height - s.StartHeight
	period := uint64(elapsed / s.PeriodLength)
	if period >= s.Periods {
		return 0, false
	}
	return period, true
}

// EndHeight returns the first height at which issuance has stopped.
func (s *IssuanceSchedule) EndHeight() int64 {
	return s.StartHeight + int64(s.Periods)*s.PeriodLength
}
