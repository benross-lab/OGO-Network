package monetary

import (
	"testing"

	"cosmossdk.io/math"
)

// TEST-ONLY values. These numbers are test parameters only. They are NOT
// the chosen OGO max supply or issuance schedule, which remain TBD. Tests
// run across several values to prove validation is parameterised, not
// hard-coded to one number. Never present any of these as production
// values.
var testOnlyMaxSupplies = []math.Int{
	math.NewInt(1_000_000),         // TEST-ONLY
	math.NewInt(100_000_000),       // TEST-ONLY
	math.NewInt(1_000_000_000),     // TEST-ONLY: 1_000 OGO at 6 decimals
	math.NewInt(1_000_000_000_000), // TEST-ONLY
}

// testOnlySchedule returns a small TEST-ONLY schedule: 10 periods of
// 100 blocks, starting at height 1, 1_000 uogo in period 0, reduction
// 9/10 per period. Total scheduled issuance under the exact rule is
// 6_511 uogo, far below every TEST-ONLY max above, so schedule-fit
// checks pass unless a test deliberately breaks them.
func testOnlySchedule() *IssuanceSchedule {
	return &IssuanceSchedule{
		ScheduleType:         ScheduleTypeReducingPerPeriod,
		StartHeight:          1,
		PeriodLength:         100,
		Periods:              10,
		InitialPerPeriod:     math.NewInt(1_000), // TEST-ONLY
		ReductionNumerator:   9,
		ReductionDenominator: 10,
	}
}

func testGenesis(max math.Int, balances []Balance) GenesisState {
	p := DefaultParams()
	p.MaxSupply = &max
	p.IssuanceSchedule = testOnlySchedule()
	return GenesisState{Params: p, Balances: balances}
}

func TestValidateGenesisAcceptsExactMax(t *testing.T) {
	for _, max := range testOnlyMaxSupplies {
		gs := testGenesis(max, []Balance{
			{Address: "ogo1alice", Denom: Denom, Amount: max, Purpose: "public"},
		})
		// Exact-max genesis would exceed the cap once scheduled issuance
		// is added, so it must FAIL under the Constitution framework:
		// genesis supply + total scheduled issuance must fit under max.
		// This test asserts that failure is enforced.
		if err := ValidateGenesis(gs); err == nil {
			t.Fatalf("TEST-ONLY max %s: genesis at exact max + schedule must fail (schedule would exceed cap)", max)
		}
	}
}

func TestValidateGenesisAcceptsBelowMax(t *testing.T) {
	for _, max := range testOnlyMaxSupplies {
		gs := testGenesis(max, []Balance{
			{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(1), Purpose: "public"},
		})
		if err := ValidateGenesis(gs); err != nil {
			t.Fatalf("TEST-ONLY max %s: below-max genesis should pass: %v", max, err)
		}
	}
}

func TestValidateGenesisRejectsMissingMax(t *testing.T) {
	gs := GenesisState{Params: DefaultParams(), Balances: nil}
	gs.Params.IssuanceSchedule = testOnlySchedule()
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("missing max_supply must fail genesis validation")
	} else {
		t.Logf("missing max correctly rejected: %v", err)
	}
}

func TestValidateGenesisRejectsZeroAndNegativeMax(t *testing.T) {
	for _, v := range []int64{0, -1, -1_000_000} {
		max := math.NewInt(v)
		gs := testGenesis(max, nil)
		if err := ValidateGenesis(gs); err == nil {
			t.Fatalf("max_supply %d must fail genesis validation", v)
		}
	}
}

func TestValidateGenesisRejectsSupplyOverMax(t *testing.T) {
	for _, max := range testOnlyMaxSupplies {
		over := max.Add(math.NewInt(1)) // max + 1 uogo
		gs := testGenesis(max, []Balance{
			{Address: "ogo1alice", Denom: Denom, Amount: over, Purpose: "public"},
		})
		if err := ValidateGenesis(gs); err == nil {
			t.Fatalf("TEST-ONLY max %s: allocation of max+1 must fail", max)
		}
		// split over-max across two accounts must also fail
		gs2 := testGenesis(max, []Balance{
			{Address: "ogo1alice", Denom: Denom, Amount: max, Purpose: "public"},
			{Address: "ogo1bob", Denom: Denom, Amount: math.NewInt(1), Purpose: "public"},
		})
		if err := ValidateGenesis(gs2); err == nil {
			t.Fatalf("TEST-ONLY max %s: split allocation over max must fail", max)
		}
	}
}

func TestValidateGenesisRejectsScheduleOverMax(t *testing.T) {
	// Genesis supply small, but a huge schedule that would exceed max.
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(1), Purpose: "public"},
	})
	gs.Params.IssuanceSchedule.InitialPerPeriod = math.NewInt(1_000_000) // TEST-ONLY: schedule total >> max
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("schedule whose total exceeds max must fail genesis validation")
	}
}

func TestValidateGenesisRejectsMissingSchedule(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, nil)
	gs.Params.IssuanceSchedule = nil
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("missing issuance schedule must fail genesis validation")
	}
}

func TestValidateGenesisRejectsInvalidSchedule(t *testing.T) {
	max := math.NewInt(1_000_000_000) // TEST-ONLY
	bad := []func(*IssuanceSchedule){
		func(s *IssuanceSchedule) { s.ScheduleType = "flat" },
		func(s *IssuanceSchedule) { s.PeriodLength = 0 },
		func(s *IssuanceSchedule) { s.Periods = 0 },
		func(s *IssuanceSchedule) { s.InitialPerPeriod = math.NewInt(0) },
		func(s *IssuanceSchedule) { s.ReductionNumerator = 10; s.ReductionDenominator = 10 }, // ratio 1, not decreasing
		func(s *IssuanceSchedule) { s.ReductionNumerator = 11; s.ReductionDenominator = 10 }, // ratio > 1
		func(s *IssuanceSchedule) { s.StartHeight = -1 },
	}
	for i, mutate := range bad {
		gs := testGenesis(max, nil)
		mutate(gs.Params.IssuanceSchedule)
		if err := ValidateGenesis(gs); err == nil {
			t.Fatalf("invalid schedule case %d must fail genesis validation", i)
		}
	}
}

func TestValidateGenesisRejectsWrongDenom(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: "uatom", Amount: math.NewInt(100), Purpose: "public"},
	})
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("wrong-denom allocation must fail")
	}
	gs2 := GenesisState{Params: func() Params {
		p := DefaultParams()
		p.MaxSupply = &max
		p.IssuanceSchedule = testOnlySchedule()
		p.Denom = "uatom"
		return p
	}(), Balances: nil}
	if err := ValidateGenesis(gs2); err == nil {
		t.Fatal("wrong-denom params must fail")
	}
}

func TestValidateGenesisRejectsNegativeAllocation(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(-5), Purpose: "public"},
	})
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("negative allocation must fail")
	}
}

func TestValidateGenesisRejectsMintingEnabled(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, nil)
	gs.Params.MintingEnabled = true
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("minting_enabled=true (discretionary) in genesis must fail")
	}
}

func TestValidateGenesisRejectsCapNotEnforced(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, nil)
	gs.Params.SupplyCapEnforced = false
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("supply_cap_enforced=false in genesis must fail")
	}
}

func TestValidateGenesisRejectsDuplicateAddress(t *testing.T) {
	max := math.NewInt(1_000_000) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10), Purpose: "public"},
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10), Purpose: "public"},
	})
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("duplicate genesis allocation must fail")
	}
}

func TestValidateGenesisRejectsFounderAllocation(t *testing.T) {
	max := math.NewInt(1_000_000_000) // TEST-ONLY
	for _, purpose := range []string{"founder", "founder_reserve", "ogo_systems", "team_reserve", "insider", "presale", "private_sale"} {
		gs := testGenesis(max, []Balance{
			{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(1), Purpose: purpose},
		})
		if err := ValidateGenesis(gs); err == nil {
			t.Fatalf("founder/reserved purpose %q must fail genesis validation", purpose)
		}
	}
}

func TestValidateGenesisRejectsMissingPurpose(t *testing.T) {
	max := math.NewInt(1_000_000_000) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(1), Purpose: ""},
	})
	if err := ValidateGenesis(gs); err == nil {
		t.Fatal("allocation without a public purpose must fail")
	}
}

func TestNoDefaultMaxSupply(t *testing.T) {
	if DefaultParams().MaxSupply != nil {
		t.Fatal("DefaultParams must not contain a default max_supply")
	}
	if DefaultParams().IssuanceSchedule != nil {
		t.Fatal("DefaultParams must not contain a default issuance schedule")
	}
}

func TestStrictGenesisIsDefaultPath(t *testing.T) {
	// Milestone C: ValidateGenesisExact is the default genesis path
	// (NewLedger uses it). Headroom above G + S_actual is rejected.
	// TEST-ONLY schedule: G 10_000, exact S_actual 6_511, M 16_511.
	exact := testGenesis(math.NewInt(16_511), []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10_000), Purpose: "public"},
	})
	if err := ValidateGenesisExact(exact); err != nil {
		t.Fatalf("exact M = G + S_actual must pass strict validation: %v", err)
	}
	if _, err := NewLedger(exact); err != nil {
		t.Fatalf("NewLedger with exact M must succeed: %v", err)
	}
	headroom := testGenesis(math.NewInt(16_512), []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10_000), Purpose: "public"},
	})
	if err := ValidateGenesisExact(headroom); err == nil {
		t.Fatal("max 1 uogo above G + S_actual must fail strict validation")
	}
	if _, err := NewLedger(headroom); err == nil {
		t.Fatal("NewLedger with headroom above G + S_actual must fail (strict is the default path)")
	}
	// The lenient ValidateGenesis still passes the headroom case by
	// design (it is a building block, not a genesis path); the test
	// pins that distinction so a future reader does not mistake it
	// for the default.
	if err := ValidateGenesis(headroom); err != nil {
		t.Fatalf("lenient ValidateGenesis should still pass headroom (it is not the default path): %v", err)
	}
}
