package monetary

import (
	"testing"

	"cosmossdk.io/math"
)

// Milestone 3 model tests. All expected values below were computed
// INDEPENDENTLY with Python Fractions (exact rational arithmetic),
// not by running the Go code and copying its output. The Python used:
//   floor(R0 * (num/den)^n) per period, summed exactly.
// Milestone C decision D-B4: the single rounding rule is exact
// floor(R0 * r^n). The former stepwise variant gave 477 at period 7
// and 386 at period 9; it is removed. See TestSingleRoundingRuleExact.
// All parameters are TEST-ONLY.

func mustModel(t *testing.T, m IssuanceModel) IssuanceModel {
	t.Helper()
	if err := m.Validate(); err != nil {
		t.Fatalf("model validation failed: %v", err)
	}
	return m
}

func TestGeometricScheduleHandComputed(t *testing.T) {
	// R0=1000, r=9/10, N=10. Python Fractions:
	// [1000, 900, 810, 729, 656, 590, 531, 478, 430, 387], sum 6511.
	// Exact series total floor(1000*(1-0.9^10)/0.1) = 6513, dust 2.
	g := mustModel(t, &GeometricSchedule{
		InitialPerPeriod:     math.NewInt(1000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           10,
		Target:               math.NewInt(6513),
	})
	want := []int64{1000, 900, 810, 729, 656, 590, 531, 478, 430, 387}
	for n, w := range want {
		got, err := g.AmountForEpoch(int64(n))
		if err != nil {
			t.Fatalf("period %d: %v", n, err)
		}
		if !got.Equal(math.NewInt(w)) {
			t.Fatalf("period %d: got %s, want %d (Python Fractions floor(1000*0.9^%d))", n, got, w, n)
		}
	}
	if got, _ := g.AmountForEpoch(10); !got.IsZero() {
		t.Fatalf("period 10 (past end): got %s, want 0", got)
	}
	if got, _ := g.TotalScheduled(); !got.Equal(math.NewInt(6511)) {
		t.Fatalf("total scheduled: got %s, want 6511", got)
	}
	if got, _ := g.TotalTarget(); !got.Equal(math.NewInt(6513)) {
		t.Fatalf("total target: got %s, want 6513", got)
	}
	if got, _ := g.Dust(); !got.Equal(math.NewInt(2)) {
		t.Fatalf("dust: got %s, want 2 (6513-6511)", got)
	}
	if g.Model() != "geometric" || g.Periods() != 10 {
		t.Fatalf("model/periods: got %s/%d", g.Model(), g.Periods())
	}
}

func TestGeometricScheduleSpecResearchParameters(t *testing.T) {
	// Spec section 5.1 research figures, R0=100000, r=9/10, N=110,
	// S_target=1000000: total scheduled 999935, dust 65, R_109 = 1.
	// Python Fractions confirms sum 999935 and period 109 = 1.
	// The former stepwise variant gave 999572 and 0 at period 93
	// for these parameters, so this test also proves the exact
	// formula is the one the spec describes.
	g := mustModel(t, &GeometricSchedule{
		InitialPerPeriod:     math.NewInt(100000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           110,
		Target:               math.NewInt(1000000),
	})
	if got, _ := g.AmountForEpoch(109); !got.Equal(math.NewInt(1)) {
		t.Fatalf("period 109: got %s, want 1", got)
	}
	if got, _ := g.TotalScheduled(); !got.Equal(math.NewInt(999935)) {
		t.Fatalf("total scheduled: got %s, want 999935", got)
	}
	if got, _ := g.Dust(); !got.Equal(math.NewInt(65)) {
		t.Fatalf("dust: got %s, want 65 (1000000-999935)", got)
	}
}

func TestSingleRoundingRuleExact(t *testing.T) {
	// Milestone C decision D-B4: ONE rounding rule, exact
	// floor(R0 * r^n). The former stepwise variant (floor at every
	// multiplication step) gave 477 at period 7 and 386 at period 9;
	// it is removed. The schedule AmountForPeriod and the model
	// AmountForEpoch must now agree exactly at every period.
	legacy := testOnlySchedule()
	g := &GeometricSchedule{
		InitialPerPeriod:     math.NewInt(1000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           10,
		Target:               math.NewInt(6513),
	}
	for _, n := range []int64{0, 1, 7, 9} {
		exact, _ := g.AmountForEpoch(n)
		fromSchedule := legacy.AmountForPeriod(uint64(n))
		if !exact.Equal(fromSchedule) {
			t.Fatalf("period %d: model %s != schedule %s (single rule violated)", n, exact, fromSchedule)
		}
	}
	if got := legacy.AmountForPeriod(7); !got.Equal(math.NewInt(478)) {
		t.Fatalf("schedule period 7: got %s, want 478 (exact rule)", got)
	}
	if got := legacy.AmountForPeriod(9); !got.Equal(math.NewInt(387)) {
		t.Fatalf("schedule period 9: got %s, want 387 (exact rule)", got)
	}
	if got := legacy.TotalIssuance(); !got.Equal(math.NewInt(6511)) {
		t.Fatalf("schedule total: got %s, want 6511", got)
	}
}

func TestHybridScheduleHandComputed(t *testing.T) {
	// base=2000, B0=8000, r=9/10, N=10. Python Fractions:
	// bonus floor(8000*0.9^n): [8000,7200,6480,5832,5248,4723,4251,3826,3443,3099]
	// R_n = 2000 + bonus: [10000,9200,8480,7832,7248,6723,6251,5826,5443,5099]
	// sum 72102. Series floor total = 20000 + floor(sum 8000*0.9^n)
	// = 72105, dust 3.
	h := mustModel(t, &HybridSchedule{
		BasePerPeriod:        math.NewInt(2000),
		BonusInitial:         math.NewInt(8000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           10,
		Target:               math.NewInt(72105),
	})
	want := []int64{10000, 9200, 8480, 7832, 7248, 6723, 6251, 5826, 5443, 5099}
	for n, w := range want {
		got, err := h.AmountForEpoch(int64(n))
		if err != nil {
			t.Fatalf("period %d: %v", n, err)
		}
		if !got.Equal(math.NewInt(w)) {
			t.Fatalf("period %d: got %s, want %d", n, got, w)
		}
	}
	if got, _ := h.TotalScheduled(); !got.Equal(math.NewInt(72102)) {
		t.Fatalf("total scheduled: got %s, want 72102", got)
	}
	if got, _ := h.Dust(); !got.Equal(math.NewInt(3)) {
		t.Fatalf("dust: got %s, want 3 (72105-72102)", got)
	}
	// Hybrid never drops below its base while running.
	for n := int64(0); n < 10; n++ {
		amt, _ := h.AmountForEpoch(n)
		if amt.LT(math.NewInt(2000)) {
			t.Fatalf("period %d amount %s below base 2000", n, amt)
		}
	}
}

func TestHybridScheduleSpecResearchParameters(t *testing.T) {
	// Spec section 5.2, base 40 percent: base=10000, B0=60900,
	// r=9/10, N=40, S_target=1000000. Python Fractions: first
	// 70900, last 11000, sum 999978, dust 22.
	h := mustModel(t, &HybridSchedule{
		BasePerPeriod:        math.NewInt(10000),
		BonusInitial:         math.NewInt(60900),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           40,
		Target:               math.NewInt(1000000),
	})
	if got, _ := h.AmountForEpoch(0); !got.Equal(math.NewInt(70900)) {
		t.Fatalf("period 0: got %s, want 70900", got)
	}
	if got, _ := h.AmountForEpoch(39); !got.Equal(math.NewInt(11000)) {
		t.Fatalf("period 39: got %s, want 11000", got)
	}
	if got, _ := h.TotalScheduled(); !got.Equal(math.NewInt(999978)) {
		t.Fatalf("total scheduled: got %s, want 999978", got)
	}
	if got, _ := h.Dust(); !got.Equal(math.NewInt(22)) {
		t.Fatalf("dust: got %s, want 22 (1000000-999978)", got)
	}
}

func TestTaperedHybridScheduleHandComputed(t *testing.T) {
	// Same hybrid as TestHybridScheduleHandComputed, taper 95/100
	// over the final 5 periods (K=5), renormalized to the same total
	// 72102 by the definition in models.go (scale all raw amounts by
	// S/S_raw, floored, remainder 1 uogo to the earliest periods).
	// Python Fractions gives the renormalized amounts:
	// [10580, 9733, 8972, 8286, 7668, 6755, 5967, 5283, 4687, 4171]
	// sum 72102, final 4171 vs untapered final 5099.
	th := mustModel(t, &TaperedHybridSchedule{
		Hybrid: HybridSchedule{
			BasePerPeriod:        math.NewInt(2000),
			BonusInitial:         math.NewInt(8000),
			ReductionNumerator:   9,
			ReductionDenominator: 10,
			NumPeriods:           10,
			Target:               math.NewInt(72105),
		},
		TaperNumerator:   95,
		TaperDenominator: 100,
		TaperPeriods:     5,
	})
	want := []int64{10580, 9733, 8972, 8286, 7668, 6755, 5967, 5283, 4687, 4171}
	for n, w := range want {
		got, err := th.AmountForEpoch(int64(n))
		if err != nil {
			t.Fatalf("period %d: %v", n, err)
		}
		if !got.Equal(math.NewInt(w)) {
			t.Fatalf("period %d: got %s, want %d", n, got, w)
		}
	}
	if got, _ := th.TotalScheduled(); !got.Equal(math.NewInt(72102)) {
		t.Fatalf("total scheduled: got %s, want 72102 (same as untapered hybrid)", got)
	}
	if got, _ := th.Dust(); !got.Equal(math.NewInt(3)) {
		t.Fatalf("dust: got %s, want 3", got)
	}
	if th.Model() != "tapered-hybrid" {
		t.Fatalf("model: got %q", th.Model())
	}
}

func TestTaperedHybridSpecResearchParametersMeasured(t *testing.T) {
	// Spec research parameters, taper 95/100 over final 20 of 40.
	// MEASURED values from Python Fractions, not the spec's rounded
	// claim: renormalized final payment 4360 (untapered 11000),
	// total 999978 (same as untapered), max single-period drop
	// 928 basis points (9.28 percent), at the first tapered period.
	// The spec's "about 5000, max drop 6.5 percent" is NOT
	// reproduced for these parameters; this test asserts the
	// measured values and the report says so.
	th := mustModel(t, &TaperedHybridSchedule{
		Hybrid: HybridSchedule{
			BasePerPeriod:        math.NewInt(10000),
			BonusInitial:         math.NewInt(60900),
			ReductionNumerator:   9,
			ReductionDenominator: 10,
			NumPeriods:           40,
			Target:               math.NewInt(1000000),
		},
		TaperNumerator:   95,
		TaperDenominator: 100,
		TaperPeriods:     20,
	})
	if got, _ := th.TotalScheduled(); !got.Equal(math.NewInt(999978)) {
		t.Fatalf("total scheduled: got %s, want 999978", got)
	}
	if got, _ := th.AmountForEpoch(39); !got.Equal(math.NewInt(4360)) {
		t.Fatalf("final payment: got %s, want 4360 (measured, not spec's about 5000)", got)
	}
	prev, _ := th.AmountForEpoch(0)
	maxDropBP := math.NewInt(0)
	for n := int64(1); n < 40; n++ {
		cur, _ := th.AmountForEpoch(n)
		if cur.GT(prev) {
			t.Fatalf("period %d amount %s exceeds previous %s: tapered amounts must decrease", n, cur, prev)
		}
		dropBP := prev.Sub(cur).MulRaw(10000).Quo(prev)
		if dropBP.GT(maxDropBP) {
			maxDropBP = dropBP
		}
		prev = cur
	}
	if !maxDropBP.Equal(math.NewInt(928)) {
		t.Fatalf("max single-period drop: got %s bp, want 928 bp (measured 9.28 percent, not spec's 6.5 percent)", maxDropBP)
	}
}

func TestScheduleValidationRejectsBadModels(t *testing.T) {
	bad := []IssuanceModel{
		&GeometricSchedule{InitialPerPeriod: math.NewInt(0), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(100)},
		&GeometricSchedule{InitialPerPeriod: math.NewInt(1000), ReductionNumerator: 10, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(10000)},
		&GeometricSchedule{InitialPerPeriod: math.NewInt(1000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 0, Target: math.NewInt(10000)},
		&GeometricSchedule{InitialPerPeriod: math.NewInt(1000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(100)},
		&HybridSchedule{BasePerPeriod: math.NewInt(0), BonusInitial: math.NewInt(8000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(72105)},
		&HybridSchedule{BasePerPeriod: math.NewInt(2000), BonusInitial: math.NewInt(8000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(10000)},
		&TaperedHybridSchedule{
			Hybrid:         HybridSchedule{BasePerPeriod: math.NewInt(2000), BonusInitial: math.NewInt(8000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(72105)},
			TaperNumerator: 100, TaperDenominator: 100, TaperPeriods: 5,
		},
		&TaperedHybridSchedule{
			Hybrid:         HybridSchedule{BasePerPeriod: math.NewInt(2000), BonusInitial: math.NewInt(8000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(72105)},
			TaperNumerator: 95, TaperDenominator: 100, TaperPeriods: 11,
		},
		&TaperedHybridSchedule{
			Hybrid:         HybridSchedule{BasePerPeriod: math.NewInt(2000), BonusInitial: math.NewInt(8000), ReductionNumerator: 9, ReductionDenominator: 10, NumPeriods: 10, Target: math.NewInt(72105)},
			TaperNumerator: 95, TaperDenominator: 100, TaperPeriods: 0,
		},
	}
	for i, m := range bad {
		if err := m.Validate(); err == nil {
			t.Fatalf("bad model %d (%s) must fail validation", i, m.Model())
		}
	}
}

func TestGenesisModelSelection(t *testing.T) {
	// Geometric via genesis: empty issuance_model selects geometric.
	max := math.NewInt(16511) // TEST-ONLY: G 10000 + exact S_actual 6511
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10000), Purpose: "public"},
	})
	m, err := gs.Params.SelectedModel()
	if err != nil {
		t.Fatalf("geometric selection failed: %v", err)
	}
	if m.Model() != "geometric" {
		t.Fatalf("empty issuance_model must select geometric, got %q", m.Model())
	}
	if got, _ := m.TotalScheduled(); !got.Equal(math.NewInt(6511)) {
		t.Fatalf("selected geometric total: got %s, want 6511 (exact rule)", got)
	}
	// Unknown model is rejected by genesis validation.
	gs.Params.IssuanceModel = "halving"
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("unknown issuance_model must fail genesis validation")
	}
	// Hybrid via genesis: base 200, R0 1000, N 10. Bonus 800.
	// Exact hybrid total: 200*10 + floor-sum(800*0.9^n)
	// = 2000 + 5207 = 7207. Bonus floors (Python Fractions):
	// [800,720,648,583,524,472,425,382,344,309]. G 10000, M = 17207.
	gs2 := testGenesis(math.NewInt(17207), []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10000), Purpose: "public"},
	})
	gs2.Params.IssuanceModel = ModelHybrid
	gs2.Params.IssuanceSchedule.BasePerPeriod = math.NewInt(200)
	if err := ValidateGenesis(gs2); err != nil {
		t.Fatalf("hybrid genesis must validate: %v", err)
	}
	m2, err := gs2.Params.SelectedModel()
	if err != nil {
		t.Fatalf("hybrid selection failed: %v", err)
	}
	if got, _ := m2.TotalScheduled(); !got.Equal(math.NewInt(7207)) {
		t.Fatalf("selected hybrid total: got %s, want 7207", got)
	}
	if got, _ := m2.AmountForEpoch(0); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("hybrid period 0: got %s, want 1000 (base 200 + bonus 800)", got)
	}
	// Tapered via genesis, same hybrid params, taper 95/100 over 5.
	gs3 := testGenesis(math.NewInt(17207), []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10000), Purpose: "public"},
	})
	gs3.Params.IssuanceModel = ModelTaperedHybrid
	gs3.Params.IssuanceSchedule.BasePerPeriod = math.NewInt(200)
	gs3.Params.IssuanceSchedule.TaperEnabled = true
	gs3.Params.IssuanceSchedule.TaperNumerator = 95
	gs3.Params.IssuanceSchedule.TaperDenominator = 100
	gs3.Params.IssuanceSchedule.TaperPeriods = 5
	if err := ValidateGenesis(gs3); err != nil {
		t.Fatalf("tapered genesis must validate: %v", err)
	}
	m3, err := gs3.Params.SelectedModel()
	if err != nil {
		t.Fatalf("tapered selection failed: %v", err)
	}
	if m3.Model() != "tapered-hybrid" {
		t.Fatalf("tapered model: got %q", m3.Model())
	}
	if got, _ := m3.TotalScheduled(); !got.Equal(math.NewInt(7207)) {
		t.Fatalf("tapered total must equal untapered hybrid total 7207, got %s", got)
	}
	// Tapered without taper_enabled is rejected.
	gs3.Params.IssuanceSchedule.TaperEnabled = false
	if err := ValidateGenesis(gs3); err == nil {
		t.Fatal("tapered-hybrid without taper_enabled must fail genesis validation")
	}
}

func TestGenesisExactMaxSupply(t *testing.T) {
	// M = G + S_actual by exact integer summation (spec section 3.1).
	// Geometric TEST-ONLY schedule: G 10000, exact S_actual 6511,
	// so M must be exactly 16511.
	gs := testGenesis(math.NewInt(16511), []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10000), Purpose: "public"},
	})
	if got, err := ComputeMaxSupply(gs); err != nil || !got.Equal(math.NewInt(16511)) {
		t.Fatalf("ComputeMaxSupply: got %s, err %v, want 16511", got, err)
	}
	if err := ValidateGenesisExact(gs); err != nil {
		t.Fatalf("exact max 16511 must pass strict validation: %v", err)
	}
	// The former headroom fixture (16508 = G + stepwise 6508) FAILS
	// strict validation: it is 3 uogo short of G + exact S_actual.
	gs.Params.MaxSupply = func() *math.Int { v := math.NewInt(16508); return &v }()
	if err := ValidateGenesisExact(gs); err == nil {
		t.Fatal("max 16508 must fail strict validation (16511 required)")
	}
	// Any other max fails strict validation, even though the loose
	// ValidateGenesis accepts max above G + S_actual.
	gs.Params.MaxSupply = func() *math.Int { v := math.NewInt(1000000); return &v }()
	if err := ValidateGenesisExact(gs); err == nil {
		t.Fatal("max 1000000 must fail strict validation")
	}
	if err := ValidateGenesis(gs); err != nil {
		t.Fatalf("loose ValidateGenesis must still accept max 1000000 (legacy rule): %v", err)
	}
}
