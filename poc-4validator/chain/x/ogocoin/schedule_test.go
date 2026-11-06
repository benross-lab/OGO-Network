package ogocoin

import (
	"testing"

	"cosmossdk.io/math"
)

// Schedule math is deterministic and pure: AmountForPeriod and
// TotalIssuance depend only on the schedule, never on chain state.
// All values below are TEST-ONLY.

func TestScheduleAmountDecreases(t *testing.T) {
	s := testOnlySchedule()
	prev := s.AmountForPeriod(0)
	if !prev.Equal(math.NewInt(1_000)) {
		t.Fatalf("period 0 must issue exactly initial_per_period 1000, got %s", prev)
	}
	for p := uint64(1); p < s.Periods; p++ {
		amt := s.AmountForPeriod(p)
		if amt.GT(prev) {
			t.Fatalf("period %d amount %s exceeds previous period %s: issuance must decrease", p, amt, prev)
		}
		prev = amt
	}
	// 1000 * 0.9^9 floored stepwise = 386
	if got := s.AmountForPeriod(9); !got.Equal(math.NewInt(386)) {
		t.Fatalf("period 9 amount: got %s, want 386 (1000 * 0.9^9 floored stepwise)", got)
	}
}

func TestScheduleTotalIssuance(t *testing.T) {
	s := testOnlySchedule()
	// 1000 + 900 + 810 + 729 + 656 + 590 + 531 + 477 + 429 + 386 = 6508
	if got := s.TotalIssuance(); !got.Equal(math.NewInt(6_508)) {
		t.Fatalf("total scheduled issuance: got %s, want 6508", got)
	}
}

func TestSchedulePeriodForHeight(t *testing.T) {
	s := testOnlySchedule() // start 1, period length 100, 10 periods
	cases := []struct {
		height int64
		period uint64
		ok     bool
	}{
		{0, 0, false},    // before start
		{1, 0, true},     // period 0 start
		{100, 0, true},   // period 0 end
		{101, 1, true},   // period 1 start
		{1000, 9, true},  // last period start
		{1001, 0, false}, // past end: issuance stops
		{100_000, 0, false},
	}
	for _, c := range cases {
		p, ok := s.PeriodForHeight(c.height)
		if ok != c.ok || (ok && p != c.period) {
			t.Fatalf("height %d: got period %d ok=%v, want period %d ok=%v", c.height, p, ok, c.period, c.ok)
		}
	}
	if got := s.EndHeight(); got != 1001 {
		t.Fatalf("end height: got %d, want 1001", got)
	}
}

func TestScheduledIssuanceExactAmountOncePerPeriod(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10_000), Purpose: "public"},
	})
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("ledger setup failed: %v", err)
	}
	before := l.TotalSupply()
	if err := l.MintScheduled(1); err != nil {
		t.Fatalf("scheduled issuance at height 1 (period 0) must succeed: %v", err)
	}
	if got := l.BalanceOf(RewardsPool); !got.Equal(math.NewInt(1_000)) {
		t.Fatalf("rewards pool after period 0: got %s, want exactly 1000", got)
	}
	if got := l.TotalSupply(); !got.Equal(before.Add(math.NewInt(1_000))) {
		t.Fatalf("total supply after period 0: got %s, want %s", got, before.Add(math.NewInt(1_000)))
	}
	// Same period, later height: double issuance must fail.
	if err := l.MintScheduled(50); err == nil {
		t.Fatal("second issuance in the same period must fail")
	}
	if got := l.TotalSupply(); !got.Equal(before.Add(math.NewInt(1_000))) {
		t.Fatalf("total supply changed after failed double issuance: %s", got)
	}
	// Next period issues exactly period 1 amount (900).
	if err := l.MintScheduled(101); err != nil {
		t.Fatalf("scheduled issuance at height 101 (period 1) must succeed: %v", err)
	}
	if got := l.BalanceOf(RewardsPool); !got.Equal(math.NewInt(1_900)) {
		t.Fatalf("rewards pool after period 1: got %s, want exactly 1900", got)
	}
}

func TestScheduledIssuanceBeforeStartFails(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	l := mustLedger(t, max)
	if err := l.MintScheduled(0); err == nil {
		t.Fatal("issuance before schedule start must fail")
	}
}

func TestScheduledIssuanceAfterEndFailsForever(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	l := mustLedger(t, max)
	// Claim every period.
	for p := uint64(0); p < 10; p++ {
		h := int64(1) + int64(p)*100
		if err := l.MintScheduled(h); err != nil {
			t.Fatalf("period %d issuance failed: %v", p, err)
		}
	}
	if got := l.TotalIssued; !got.Equal(math.NewInt(6_508)) {
		t.Fatalf("total issued after all periods: got %s, want 6508", got)
	}
	// After the last period, issuance is permanently over.
	for _, h := range []int64{1001, 2000, 1_000_000} {
		if err := l.MintScheduled(h); err == nil {
			t.Fatalf("issuance at height %d after schedule end must fail", h)
		}
	}
	if got := l.TotalIssued; !got.Equal(math.NewInt(6_508)) {
		t.Fatalf("total issued changed after failed post-schedule attempts: %s", got)
	}
}

func TestScheduledIssuanceNeverExceedsCap(t *testing.T) {
	// Max set exactly at genesis supply + total scheduled issuance:
	// 10_000 + 6_508 = 16_508. Every period issues, and the final total
	// lands exactly on max. Nothing may push it over.
	max := math.NewInt(16_508) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10_000), Purpose: "public"},
	})
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("ledger setup failed: %v", err)
	}
	for p := uint64(0); p < 10; p++ {
		if err := l.MintScheduled(int64(1) + int64(p)*100); err != nil {
			t.Fatalf("period %d issuance failed: %v", p, err)
		}
	}
	if got := l.TotalSupply(); !got.Equal(max) {
		t.Fatalf("total supply after full schedule: got %s, want exactly max %s", got, max)
	}
	// Discretionary mint of even 1 uogo over the cap must fail.
	if err := l.Mint("ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("discretionary mint must fail at the cap")
	}
	if got := l.TotalSupply(); !got.Equal(max) {
		t.Fatalf("total supply changed after failed mint: %s", got)
	}
}
