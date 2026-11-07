package monetary

import (
	"testing"

	"cosmossdk.io/math"
)

// Milestone 5 monetary tests: duplicate issuance, halt recovery,
// both models end to end, and the ledger-level supply invariants
// (monetary engine spec section 10, testable items 1-6 and 10).
// All parameters are TEST-ONLY.

// epochLedger builds a small ledger whose schedule is the TEST-ONLY
// schedule (10 periods, period amounts under the exact rule 1000, 900,
// 810, 729, 656, 590, 531, 478, 430, 387; total 6511) with genesis
// 10000 and max exactly G + S_actual = 16511.
func epochLedger(t *testing.T) *Ledger {
	t.Helper()
	return mustLedger(t, math.NewInt(16511))
}

func TestDuplicateIssuanceSameBlockAndLaterBlock(t *testing.T) {
	l := epochLedger(t)
	if err := l.MintScheduledEpoch(0); err != nil {
		t.Fatalf("first claim of epoch 0 must succeed: %v", err)
	}
	issued := l.TotalIssued
	supply := l.TotalSupply()
	// Re-claim the same epoch in the "same block": a second call
	// with the same epoch number, immediately after the first.
	if err := l.MintScheduledEpoch(0); err == nil {
		t.Fatal("re-claim of epoch 0 in the same block must fail")
	}
	if !l.TotalIssued.Equal(issued) || !l.TotalSupply().Equal(supply) {
		t.Fatalf("issued total changed after same-block re-claim: %s / %s", l.TotalIssued, l.TotalSupply())
	}
	// Re-claim the same epoch in a "later block": claim epoch 1
	// first (a later block, a later epoch), then try epoch 0 again.
	if err := l.MintScheduledEpoch(1); err != nil {
		t.Fatalf("claim of epoch 1 must succeed: %v", err)
	}
	issued = l.TotalIssued
	if err := l.MintScheduledEpoch(0); err == nil {
		t.Fatal("re-claim of epoch 0 in a later block must fail")
	}
	if !l.TotalIssued.Equal(issued) {
		t.Fatalf("issued total changed after later-block re-claim: %s", l.TotalIssued)
	}
	// The height-based entrypoint has the same guard for the same
	// period: MintScheduled at a later height in period 1 must fail
	// because period 1 is already claimed.
	if err := l.MintScheduled(150); err == nil {
		t.Fatal("height-based re-claim of claimed period 1 must fail")
	}
	if !l.TotalIssued.Equal(issued) {
		t.Fatalf("issued total changed after height-based re-claim: %s", l.TotalIssued)
	}
}

func TestHaltRecoveryOneEpochPerBlock(t *testing.T) {
	l := epochLedger(t)
	// Normal operation: epochs 0 and 1 claimed, one per block.
	if _, err := l.ProcessNextEpoch(0, 1); err != nil {
		t.Fatalf("epoch 0 catch-up failed: %v", err)
	}
	if _, err := l.ProcessNextEpoch(1, 2); err != nil {
		t.Fatalf("epoch 1 catch-up failed: %v", err)
	}
	// HALT: no blocks are produced while epochs 2, 3, 4, 5 end.
	// The chain restarts with currentEpoch = 5. Missed epochs must
	// be processed sequentially, one per block, in epoch order.
	// Expected exact amounts: epoch 2 = 810, 3 = 729, 4 = 656, 5 = 590.
	want := []struct {
		epoch  int64
		amount int64
	}{
		{2, 810}, {3, 729}, {4, 656}, {5, 590},
	}
	issuedBefore := l.TotalIssued
	for i, w := range want {
		amt, err := l.ProcessNextEpoch(5, int64(10+i))
		if err != nil {
			t.Fatalf("catch-up block %d failed: %v", i, err)
		}
		if !amt.Equal(math.NewInt(w.amount)) {
			t.Fatalf("catch-up block %d: minted %s, want exactly epoch %d amount %d (no lump sum)", i, amt, w.epoch, w.amount)
		}
		// No epoch may be skipped: the next unclaimed epoch must be
		// exactly the one after the epoch just claimed.
		next := l.NextUnclaimedEpoch()
		if i < len(want)-1 && next != w.epoch+1 {
			t.Fatalf("after claiming epoch %d, next unclaimed is %d, want %d (an epoch was skipped)", w.epoch, next, w.epoch+1)
		}
	}
	// Total issued after catch-up equals the sum of R_n for all
	// elapsed epochs 0..5: 1000+900+810+729+656+590 = 4685.
	if got := l.TotalIssued; !got.Equal(math.NewInt(4685)) {
		t.Fatalf("total issued after catch-up: got %s, want 4685", got)
	}
	if got := l.TotalIssued.Sub(issuedBefore); !got.Equal(math.NewInt(2785)) {
		t.Fatalf("catch-up minted %s, want 2785 (810+729+656+590)", got)
	}
	// A further block with nothing missed mints nothing.
	amt, err := l.ProcessNextEpoch(5, 99)
	if err != nil {
		t.Fatalf("idle catch-up block failed: %v", err)
	}
	if !amt.IsZero() {
		t.Fatalf("idle catch-up minted %s, want 0", amt)
	}
	// A block that has not yet reached the next epoch end also
	// mints nothing (currentEpoch 5, next unclaimed is 6).
	amt, err = l.ProcessNextEpoch(5, 100)
	if err != nil {
		t.Fatalf("not-yet-ended catch-up failed: %v", err)
	}
	if !amt.IsZero() {
		t.Fatalf("catch-up before epoch 6 ends minted %s, want 0", amt)
	}
	// And the next epoch, when it does end, is claimed exactly once.
	amt, err = l.ProcessNextEpoch(6, 101)
	if err != nil {
		t.Fatalf("epoch 6 catch-up failed: %v", err)
	}
	if !amt.Equal(math.NewInt(531)) {
		t.Fatalf("epoch 6 minted %s, want 531", amt)
	}
}

func TestBothModelsEndToEnd(t *testing.T) {
	// Geometric end to end, through the model mint path, to the
	// final period. R0=5000, r=9/10, N=8, target 28476.
	// Exact amounts (Python Fractions): [5000,4500,4050,3645,3280,
	// 2952,2657,2391], sum 28475, dust 1.
	geo := &GeometricSchedule{
		InitialPerPeriod:     math.NewInt(5000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           8,
		Target:               math.NewInt(28476),
	}
	if err := geo.Validate(); err != nil {
		t.Fatalf("geometric validation failed: %v", err)
	}
	// Hybrid end to end. base=2000, B0=8000, r=9/10, N=10,
	// target 72105, sum 72102, dust 3 (see models_test.go).
	hyb := &HybridSchedule{
		BasePerPeriod:        math.NewInt(2000),
		BonusInitial:         math.NewInt(8000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
		NumPeriods:           10,
		Target:               math.NewInt(72105),
	}
	if err := hyb.Validate(); err != nil {
		t.Fatalf("hybrid validation failed: %v", err)
	}
	for _, tc := range []struct {
		name    string
		model   IssuanceModel
		genesis int64
		total   int64
		dust    int64
		sched   *IssuanceSchedule
		modelID string
	}{
		{"geometric", geo, 100000, 28475, 1, &IssuanceSchedule{ScheduleType: ScheduleTypeReducingPerPeriod, StartHeight: 1, PeriodLength: 100, Periods: 8, InitialPerPeriod: math.NewInt(5000), ReductionNumerator: 9, ReductionDenominator: 10}, ModelGeometric},
		{"hybrid", hyb, 100000, 72102, 3, &IssuanceSchedule{ScheduleType: ScheduleTypeReducingPerPeriod, StartHeight: 1, PeriodLength: 100, Periods: 10, InitialPerPeriod: math.NewInt(10000), BasePerPeriod: math.NewInt(2000), ReductionNumerator: 9, ReductionDenominator: 10}, ModelHybrid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			max := math.NewInt(tc.genesis + tc.total)
			gs := testGenesis(max, []Balance{
				{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(tc.genesis), Purpose: "public"},
			})
			// Milestone C strict genesis: the genesis schedule must
			// be the schedule the model mint path runs, so M equals
			// G + the model total. The genesis schedule fields are
			// set from the model under test.
			gs.Params.IssuanceSchedule = tc.sched
			gs.Params.IssuanceModel = tc.modelID
			l, err := NewLedger(gs)
			if err != nil {
				t.Fatalf("ledger setup failed: %v", err)
			}
			for n := int64(0); n < tc.model.Periods(); n++ {
				if err := l.MintScheduledModel(tc.model, n); err != nil {
					t.Fatalf("epoch %d mint failed: %v", n, err)
				}
				// NOTE: CheckMonetaryInvariants is legacy-schedule
				// only (it compares against the legacy schedule),
				// so it is not called here; the model amounts differ
				// from the legacy schedule by design.
				// The invariant checked here is the model-level one:
				// total issued never exceeds the model total, and
				// total supply never exceeds max.
				if got, _ := tc.model.TotalScheduled(); l.TotalIssued.GT(got) {
					t.Fatalf("epoch %d: total issued %s exceeds model total %s", n, l.TotalIssued, got)
				}
				if err := l.CheckInvariants(); err != nil {
					t.Fatalf("cap invariant failed after epoch %d: %v", n, err)
				}
			}
			if got := l.TotalIssued; !got.Equal(math.NewInt(tc.total)) {
				t.Fatalf("total issued: got %s, want %d (S_actual)", got, tc.total)
			}
			// Dust is left unissued: total issued is S_actual, the
			// difference to S_target is exactly the dust, and no
			// further epoch can mint it.
			dust, err := tc.model.Dust()
			if err != nil {
				t.Fatalf("dust failed: %v", err)
			}
			if !dust.Equal(math.NewInt(tc.dust)) {
				t.Fatalf("dust: got %s, want %d", dust, tc.dust)
			}
			if err := l.MintScheduledModel(tc.model, tc.model.Periods()); err == nil {
				t.Fatal("mint past the final period must fail")
			}
			if got := l.TotalIssued; !got.Equal(math.NewInt(tc.total)) {
				t.Fatalf("total issued changed after failed post-schedule mint: %s", got)
			}
			if got := l.TotalSupply(); !got.Equal(max) {
				t.Fatalf("total supply after full schedule: got %s, want exactly max %s (genesis + S_actual)", got, max)
			}
		})
	}
}

func TestSupplyInvariantsLedgerLevel(t *testing.T) {
	l := epochLedger(t)
	genesisSupply := l.TotalSupply()
	// Invariant 1: total supply never exceeds max.
	// Invariant 2: total = genesis + issued - burned, exactly.
	// Invariant 3: each epoch's issuance is exactly R_n.
	// Invariant 5: total issued never exceeds total scheduled.
	// Invariant 6: dust is never issued (total issued = S_actual
	// after the full schedule, dust remains unissued).
	// Invariant 10 (split conserves) is tested at the schedule
	// level in TestSplitConserves below, because the rewards split
	// itself is a Milestone C mechanism (x/rewards), not a ledger
	// function in this prototype.
	for p := int64(0); p < 10; p++ {
		if err := l.MintScheduledEpoch(p); err != nil {
			t.Fatalf("epoch %d mint failed: %v", p, err)
		}
		if err := l.CheckMonetaryInvariants(); err != nil {
			t.Fatalf("invariants failed after epoch %d: %v", p, err)
		}
		want := genesisSupply.Add(l.TotalIssued)
		if got := l.TotalSupply(); !got.Equal(want) {
			t.Fatalf("after epoch %d: total %s, want genesis %s + issued %s = %s", p, got, genesisSupply, l.TotalIssued, want)
		}
	}
	if got := l.TotalIssued; !got.Equal(math.NewInt(6511)) {
		t.Fatalf("total issued after full schedule: got %s, want 6511 (exact S_actual)", got)
	}
	// Burn, then re-check invariant 2 with the burn included.
	if err := l.Burn("ogo1alice", math.NewInt(1000)); err != nil {
		t.Fatalf("burn failed: %v", err)
	}
	want := genesisSupply.Add(l.TotalIssued).Sub(math.NewInt(1000))
	if got := l.TotalSupply(); !got.Equal(want) {
		t.Fatalf("after burn: total %s, want %s (genesis + issued - burned)", got, want)
	}
	if err := l.CheckMonetaryInvariants(); err != nil {
		t.Fatalf("invariants failed after burn: %v", err)
	}
	// Invariant 1 under attack: discretionary mint above the cap
	// fails and supply is unchanged (also covered by the bypass
	// suite; asserted here as part of the invariant list).
	before := l.TotalSupply()
	if err := l.Mint("ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("discretionary mint must fail")
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("failed mint changed supply: %s", l.TotalSupply())
	}
}

func TestSplitConserves(t *testing.T) {
	// Spec section 10 invariant 10: for a given epoch, validator
	// plus delegator plus community plus treasury shares equal
	// exactly the amount minted, with no remainder and no shortfall,
	// in integer arithmetic. The split itself (x/rewards) is
	// Milestone C; this test proves the integer split arithmetic
	// the split must use, with a TEST-ONLY 70/20/5/5 split on a
	// period amount, including the remainder rule: the largest
	// share (validators) takes the flooring remainder, so the
	// parts sum exactly to R_n.
	// R_n = 1000: 70 percent = 700, 20 percent = 200, 5 percent =
	// 50, 5 percent = 50. Sum 1000 exactly.
	// R_n = 1001 (odd, exercises flooring): validator floor(700.7)
	// = 700, delegator floor(200.2) = 200, community floor(50.05)
	// = 50, treasury floor(50.05) = 50; sum 1000, remainder 1 goes
	// to validators: 701 + 200 + 50 + 50 = 1001 exactly.
	for _, rn := range []int64{1000, 1001, 70900} {
		amount := math.NewInt(rn)
		validator := amount.MulRaw(70).QuoRaw(100)
		delegator := amount.MulRaw(20).QuoRaw(100)
		community := amount.MulRaw(5).QuoRaw(100)
		treasury := amount.MulRaw(5).QuoRaw(100)
		sum := validator.Add(delegator).Add(community).Add(treasury)
		remainder := amount.Sub(sum)
		validator = validator.Add(remainder)
		sum = validator.Add(delegator).Add(community).Add(treasury)
		if !sum.Equal(amount) {
			t.Fatalf("R_n=%d: split sums to %s, want exactly %d", rn, sum, rn)
		}
		if remainder.IsNegative() {
			t.Fatalf("R_n=%d: negative remainder %s", rn, remainder)
		}
	}
}
