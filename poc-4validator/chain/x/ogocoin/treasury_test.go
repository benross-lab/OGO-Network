package ogocoin

import (
	"testing"

	"cosmossdk.io/math"
)

// Milestone 5 treasury tests. R-F rule:
// spend = min(D, max(Trail_4, floor(0.05*B)), max(0, B-F))
// All values TEST-ONLY, hand computed in the test comments.

func TestTreasuryTrailingAverage(t *testing.T) {
	// Last four of [100, 200, 300, 400, 500] are 200,300,400,500:
	// average (1400)/4 = 350 exactly.
	history := []math.Int{math.NewInt(100), math.NewInt(200), math.NewInt(300), math.NewInt(400), math.NewInt(500)}
	if got := TrailingAverage(history, 4); !got.Equal(math.NewInt(350)) {
		t.Fatalf("trailing average: got %s, want 350", got)
	}
	// Fewer than four: average over what exists: (100+200)/2 = 150.
	if got := TrailingAverage(history[:2], 4); !got.Equal(math.NewInt(150)) {
		t.Fatalf("short trailing average: got %s, want 150", got)
	}
	// Flooring: (100+101+102)/3 = 303/3 = 101; (100+101)/2 = 100 (100.5 floored).
	if got := TrailingAverage([]math.Int{math.NewInt(100), math.NewInt(101)}, 4); !got.Equal(math.NewInt(100)) {
		t.Fatalf("floored trailing average: got %s, want 100", got)
	}
	if got := TrailingAverage(nil, 4); !got.IsZero() {
		t.Fatalf("empty history: got %s, want 0", got)
	}
}

func TestTreasurySpendFormula(t *testing.T) {
	// B=10000, Trail_4=400, F=1000, D=1500.
	// floor(0.05*10000)=500; max(400,500)=500; max(0,10000-1000)=9000.
	// min(1500, 500, 9000) = 500. The 5 percent cap binds, not D.
	if got := SpendForPeriod(math.NewInt(1500), math.NewInt(10000), math.NewInt(400), math.NewInt(1000)); !got.Equal(math.NewInt(500)) {
		t.Fatalf("spend: got %s, want 500 (5 percent cap binds)", got)
	}
	// Trail binds: B=100000, Trail_4=2000, F=1000, D=1500.
	// floor(0.05*100000)=5000; max(2000,5000)=5000; min(1500,5000,99000)=1500.
	// Desired spend is paid in full.
	if got := SpendForPeriod(math.NewInt(1500), math.NewInt(100000), math.NewInt(2000), math.NewInt(1000)); !got.Equal(math.NewInt(1500)) {
		t.Fatalf("spend: got %s, want 1500 (desired paid in full)", got)
	}
	// Floor binds: B=10500, Trail_4=100, F=10000, D=1500.
	// max(100,525)=525; max(0,500)=500; min(1500,525,500)=500.
	// Spending stops 500 above the floor; the balance never goes below F.
	if got := SpendForPeriod(math.NewInt(1500), math.NewInt(10500), math.NewInt(100), math.NewInt(10000)); !got.Equal(math.NewInt(500)) {
		t.Fatalf("spend: got %s, want 500 (floor cap binds)", got)
	}
	// At the floor: B=10000, F=10000: max(0,0)=0, spend 0.
	if got := SpendForPeriod(math.NewInt(1500), math.NewInt(10000), math.NewInt(400), math.NewInt(10000)); !got.IsZero() {
		t.Fatalf("spend at floor: got %s, want 0", got)
	}
	// Zero desired spend: spend 0 regardless of caps.
	if got := SpendForPeriod(math.NewInt(0), math.NewInt(100000), math.NewInt(2000), math.NewInt(1000)); !got.IsZero() {
		t.Fatalf("zero desired: got %s, want 0", got)
	}
}

func TestTreasuryLowInflowBinding(t *testing.T) {
	// The spec section 5 binding finding at small TEST-ONLY scale:
	// low, steady inflow of 100 per period, starting balance 10000,
	// floor 10000, trailing window 4. Run 20 periods at D=1500 and
	// at D=2500. The rule binds, not the desire: actual spend must
	// be IDENTICAL in both runs, and the treasury must end at the
	// floor, never below it.
	run := func(desired int64) (*Treasury, error) {
		tr, err := NewTreasury(TreasuryParams{
			DesiredSpend:   math.NewInt(desired),
			ReserveFloor:   math.NewInt(10000),
			TrailingWindow: 4,
		}, math.NewInt(10000))
		if err != nil {
			return nil, err
		}
		for i := 0; i < 20; i++ {
			if _, err := tr.ApplyPeriod(math.NewInt(100)); err != nil {
				return nil, err
			}
		}
		return tr, nil
	}
	low, err := run(1500)
	if err != nil {
		t.Fatalf("D=1500 run failed: %v", err)
	}
	high, err := run(2500)
	if err != nil {
		t.Fatalf("D=2500 run failed: %v", err)
	}
	if !low.TotalSpent.Equal(high.TotalSpent) {
		t.Fatalf("binding violated: D=1500 spent %s, D=2500 spent %s; raising D must not raise actual spend under low inflow", low.TotalSpent, high.TotalSpent)
	}
	// Hand-computed: period 1 spends 0 (balance is exactly at the
	// floor, so B-F = 0), then each of the 19 remaining periods
	// spends 100 (the floor cap: balance 10100, floor 10000, so
	// at most 100 can be spent, and the 100 inflow replaces it).
	// Total spent 1900, balance 10100, never below the floor.
	if !low.TotalSpent.Equal(math.NewInt(1900)) {
		t.Fatalf("low run total spent: got %s, want 1900 (0 + 19*100)", low.TotalSpent)
	}
	if !low.Balance.Equal(math.NewInt(10100)) {
		t.Fatalf("low run must end at 10100 (one inflow above the floor), got %s", low.Balance)
	}
	if low.Balance.LT(math.NewInt(10000)) {
		t.Fatalf("balance below floor: %s", low.Balance)
	}
	// Desired was 1500*20=30000 (and 2500*20=50000); actual is far
	// below both. The shortfall is the unfunded desired spend, and
	// it must be visible, not hidden: report it, do not smooth it.
	if low.Shortfall().IsZero() || low.Shortfall().IsNegative() {
		t.Fatalf("low run must have a positive shortfall, got %s", low.Shortfall())
	}
	t.Logf("low-inflow binding: D=1500 spent %s, D=2500 spent %s, shortfall at D=1500 is %s, balance %s",
		low.TotalSpent, high.TotalSpent, low.Shortfall(), low.Balance)
}

func TestTreasuryModerateInflowPaysDesiredInFull(t *testing.T) {
	// Contrast case: inflow 2000 per period, starting balance 10000,
	// floor 1000, D=1500, 10 periods. After the first periods the
	// trailing average (2000) exceeds D, the 5 percent cap grows
	// past D, and the floor cap never binds, so the full desired
	// spend is paid every period after the ramp. Total desired
	// 15000; actual must be close to it (the first period or two
	// may be capped while Trail_4 is still 0 and the balance is
	// small: period 1 Trail_4=0, 5 percent of 10000=500, so period 1
	// spends 500, period 2 Trail_4=2000, spend 1500, and so on).
	tr, err := NewTreasury(TreasuryParams{
		DesiredSpend:   math.NewInt(1500),
		ReserveFloor:   math.NewInt(1000),
		TrailingWindow: 4,
	}, math.NewInt(10000))
	if err != nil {
		t.Fatalf("treasury setup failed: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := tr.ApplyPeriod(math.NewInt(2000)); err != nil {
			t.Fatalf("period %d failed: %v", i, err)
		}
	}
	// Period 1 spends 500, periods 2-10 spend 1500 each: total 14000.
	if !tr.TotalSpent.Equal(math.NewInt(14000)) {
		t.Fatalf("total spent: got %s, want 14000 (500 + 9*1500)", tr.TotalSpent)
	}
	// Balance: 10000 + 10*2000 inflow - 14000 spent = 16000.
	if !tr.Balance.Equal(math.NewInt(16000)) {
		t.Fatalf("balance: got %s, want 16000", tr.Balance)
	}
	if got := tr.Shortfall(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("shortfall: got %s, want 1000 (15000 desired - 14000 spent)", got)
	}
}

func TestTreasuryValidationRejectsBadParams(t *testing.T) {
	if _, err := NewTreasury(TreasuryParams{DesiredSpend: math.NewInt(-1), ReserveFloor: math.NewInt(0), TrailingWindow: 4}, math.NewInt(0)); err == nil {
		t.Fatal("negative desired spend must fail")
	}
	if _, err := NewTreasury(TreasuryParams{DesiredSpend: math.NewInt(100), ReserveFloor: math.NewInt(-1), TrailingWindow: 4}, math.NewInt(0)); err == nil {
		t.Fatal("negative reserve floor must fail")
	}
	if _, err := NewTreasury(TreasuryParams{DesiredSpend: math.NewInt(100), ReserveFloor: math.NewInt(0), TrailingWindow: 3}, math.NewInt(0)); err == nil {
		t.Fatal("trailing window other than 4 must fail (the spec rule is Trail_4)")
	}
	if _, err := NewTreasury(TreasuryParams{DesiredSpend: math.NewInt(100), ReserveFloor: math.NewInt(0), TrailingWindow: 4}, math.NewInt(-5)); err == nil {
		t.Fatal("negative starting balance must fail")
	}
}
