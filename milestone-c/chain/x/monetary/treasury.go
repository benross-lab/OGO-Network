// Treasury spending rule (Milestone 5).
//
// R-F rule, OGO_TREASURY_AND_SUSTAINABILITY.md section 2 (recommended
// default, not an approved rule). Let B be the treasury balance at
// the end of the previous period, D the desired spend for the period
// (from approved proposals), Trail_4 the trailing four-period average
// treasury inflow, and F the emergency reserve floor (genesis-fixed).
// Then:
//
//	spend_t = min(D, max(Trail_4(t), floor(0.05 * B_(t-1))),
//	              max(0, B_(t-1) - F))
//
// In words: spend no more than the desired amount, and no more than
// the larger of (the trailing inflow average, or 5 percent of the
// balance), and never spend below the emergency reserve floor.
//
// Exact integer arithmetic only: floor(0.05 * B) is B*5/100 by integer
// division, Trail_4 is a floored integer average. No floating point.
//
// Trail_4 definition used here (recorded, because the spec does not
// pin it down): the average of the treasury inflows of the previous
// four periods, or of however many previous periods exist if fewer
// than four have elapsed. The current period's inflow is NOT included:
// Trail_4(t) looks backwards only, so a period's spend is a function
// of history, not of money that has just arrived in the same period.
//
// The binding finding (spec section 5): under low fees, raising the
// desired spend does not raise actual spend. At D = 2500 the R-F rule
// spends 114511 over 120 periods for hybrid, essentially identical
// to the 114511 spent at D = 1500. A spending rule cannot fix a level
// problem: if inflow is too low, no rule creates money. The tests in
// treasury_test.go prove that property at small TEST-ONLY scale.
//
// All parameters here are TEST-ONLY. No spend level D, no floor F,
// and no treasury rule is selected for production.
package monetary

import (
	"fmt"

	"cosmossdk.io/math"
)

// ErrTreasuryInvalid is returned for invalid treasury parameters.
var ErrTreasuryInvalid = fmt.Errorf("monetary: treasury parameters are invalid")

// TreasuryParams are the R-F rule parameters. They are genesis
// parameters of x/treasury in the production design; in this
// prototype they live here so the rule can be tested without the
// full x/treasury module (Milestone C work).
type TreasuryParams struct {
	// DesiredSpend (D) is the desired spend for the period, in uogo,
	// from approved proposals.
	DesiredSpend math.Int `json:"desired_spend"`
	// ReserveFloor (F) is the emergency reserve floor, in uogo.
	// Spending never takes the balance below F.
	ReserveFloor math.Int `json:"reserve_floor"`
	// TrailingWindow is the trailing-average window in periods.
	// The spec fixes it at 4 (Trail_4); any other value is rejected
	// by Validate so the tested rule is exactly the spec rule.
	TrailingWindow int `json:"trailing_window"`
}

// Validate checks the treasury parameters.
func (p *TreasuryParams) Validate() error {
	if p == nil {
		return ErrTreasuryInvalid
	}
	if p.DesiredSpend.IsNil() || p.DesiredSpend.IsNegative() {
		return fmt.Errorf("%w: desired_spend must be >= 0", ErrTreasuryInvalid)
	}
	if p.ReserveFloor.IsNil() || p.ReserveFloor.IsNegative() {
		return fmt.Errorf("%w: reserve_floor must be >= 0", ErrTreasuryInvalid)
	}
	if p.TrailingWindow != 4 {
		return fmt.Errorf("%w: trailing window must be 4 (Trail_4), got %d", ErrTreasuryInvalid, p.TrailingWindow)
	}
	return nil
}

// TrailingAverage returns the floored average of the last
// window inflows in history (at most window, fewer if fewer have
// elapsed). History is in period order, most recent last. An empty
// history averages to 0. Exact integer arithmetic only.
func TrailingAverage(history []math.Int, window int) math.Int {
	if window <= 0 || len(history) == 0 {
		return math.NewInt(0)
	}
	start := len(history) - window
	if start < 0 {
		start = 0
	}
	sum := math.NewInt(0)
	count := 0
	for _, v := range history[start:] {
		sum = sum.Add(v)
		count++
	}
	if count == 0 {
		return math.NewInt(0)
	}
	return sum.QuoRaw(int64(count))
}

// SpendForPeriod computes the R-F spend for one period from its
// inputs directly: desired D, previous balance B, trailing average
// Trail_4, and reserve floor F. Pure function, exact integer
// arithmetic, no state. This is the function the tests prove,
// including the low-inflow binding case.
func SpendForPeriod(desired, balancePrev, trail4, floor math.Int) math.Int {
	// capByInflow = max(Trail_4, floor(0.05 * B))
	fivePercent := balancePrev.MulRaw(5).QuoRaw(100)
	capByInflow := trail4
	if fivePercent.GT(capByInflow) {
		capByInflow = fivePercent
	}
	// capByFloor = max(0, B - F)
	capByFloor := balancePrev.Sub(floor)
	if capByFloor.IsNegative() {
		capByFloor = math.NewInt(0)
	}
	spend := desired
	if capByInflow.LT(spend) {
		spend = capByInflow
	}
	if capByFloor.LT(spend) {
		spend = capByFloor
	}
	if spend.IsNegative() {
		return math.NewInt(0)
	}
	return spend
}

// Treasury is the stateful R-F treasury for the prototype: a balance,
// the inflow history that feeds Trail_4, and the rule parameters.
// It is ledger-adjacent, not the ledger: it tracks treasury money
// only, and it never mints.
type Treasury struct {
	Params  TreasuryParams `json:"params"`
	Balance math.Int       `json:"balance"`
	// Inflows is the per-period treasury inflow history, in period
	// order, most recent last. Trail_4(t) is the average of the last
	// four entries before period t is spent.
	Inflows []math.Int `json:"inflows"`
	// TotalSpent is the cumulative actual spend.
	TotalSpent math.Int `json:"total_spent"`
	// TotalDesired is the cumulative desired spend, so the shortfall
	// (desired minus actual) is always visible. The spec's binding
	// finding is stated in these terms: report desired vs actual
	// vs inflows vs reserve vs unfunded vs shortfall periods
	// separately, never as a single spend number.
	TotalDesired math.Int `json:"total_desired"`
}

// NewTreasury builds a treasury with a starting balance and the
// given rule parameters.
func NewTreasury(params TreasuryParams, startingBalance math.Int) (*Treasury, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if startingBalance.IsNil() || startingBalance.IsNegative() {
		return nil, fmt.Errorf("%w: starting balance must be >= 0", ErrTreasuryInvalid)
	}
	return &Treasury{
		Params:       params,
		Balance:      startingBalance,
		Inflows:      []math.Int{},
		TotalSpent:   math.NewInt(0),
		TotalDesired: math.NewInt(0),
	}, nil
}

// ApplyPeriod records one period: inflow is added to the balance,
// then the R-F spend for that period is computed from the trailing
// average of inflows BEFORE this period's inflow (see the Trail_4
// definition above) and deducted. It returns the actual spend.
// Desired spend D is taken from Params for this prototype; a
// production treasury takes D per period from approved proposals.
func (t *Treasury) ApplyPeriod(inflow math.Int) (math.Int, error) {
	if inflow.IsNil() || inflow.IsNegative() {
		return math.NewInt(0), fmt.Errorf("%w: inflow must be >= 0", ErrTreasuryInvalid)
	}
	trail := TrailingAverage(t.Inflows, t.Params.TrailingWindow)
	// B_(t-1) is the balance at the end of the previous period, i.e.
	// before this period's inflow is added. The spend is computed
	// on that balance, then deducted from the post-inflow balance.
	balancePrev := t.Balance
	spend := SpendForPeriod(t.Params.DesiredSpend, balancePrev, trail, t.Params.ReserveFloor)
	t.Balance = t.Balance.Add(inflow)
	t.Inflows = append(t.Inflows, inflow)
	if spend.GT(t.Balance) {
		spend = t.Balance
	}
	t.Balance = t.Balance.Sub(spend)
	t.TotalSpent = t.TotalSpent.Add(spend)
	t.TotalDesired = t.TotalDesired.Add(t.Params.DesiredSpend)
	return spend, nil
}

// Shortfall returns TotalDesired minus TotalSpent: the approved
// spending that the rule did not deliver because inflow or the
// reserve floor bound it.
func (t *Treasury) Shortfall() math.Int {
	return t.TotalDesired.Sub(t.TotalSpent)
}
