package ogocoin

import (
	"errors"
	"testing"
	"time"

	"cosmossdk.io/math"
)

// recordingHook records every hook invocation for timer tests.
type recordingHook struct {
	after     []int64
	before    []int64
	failAfter map[int64]error
}

func (h *recordingHook) AfterEpochEnd(identifier string, epochNumber int64) error {
	h.after = append(h.after, epochNumber)
	if err, ok := h.failAfter[epochNumber]; ok {
		return err
	}
	return nil
}

func (h *recordingHook) BeforeEpochStart(identifier string, epochNumber int64) error {
	h.before = append(h.before, epochNumber)
	return nil
}

// testEpochLedger builds a ledger with the TEST-ONLY schedule used by
// the epoch tests: 4 periods, initial 1000, reduction 9/10, so the
// amounts are 1000, 900, 810, 729 and the total is 3439.
func testEpochLedger(t *testing.T) *Ledger {
	t.Helper()
	max := math.NewInt(1000010000)
	gs := GenesisState{
		Params: Params{
			MaxSupply:         &max,
			Denom:             Denom,
			Decimals:          Decimals,
			InflationRate:     "0",
			MintingEnabled:    false,
			SupplyCapEnforced: true,
			IssuanceSchedule: &IssuanceSchedule{
				ScheduleType:         ScheduleTypeReducingPerPeriod,
				StartHeight:          1,
				PeriodLength:         100,
				Periods:              4,
				InitialPerPeriod:     math.NewInt(1000),
				ReductionNumerator:   9,
				ReductionDenominator: 10,
			},
		},
		Balances: []Balance{
			{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(600000000), Purpose: "public"},
			{Address: "ogo1bob", Denom: Denom, Amount: math.NewInt(400000000), Purpose: "public"},
		},
	}
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("NewLedger: %v", err)
	}
	return l
}

func TestEpochTimerFirstEpochStartsAtOne(t *testing.T) {
	genesis := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	timer := NewEpochTimer(EpochConfig{Identifier: EpochIdentifierMonetary, DurationSeconds: 30}, genesis)
	hook := &recordingHook{}
	timer.RegisterHook(hook)

	// Before the start time, nothing happens.
	if r := timer.Tick(genesis.Add(-time.Second), 1); r.StartedEpoch != 0 || r.EndedEpoch != 0 {
		t.Fatalf("tick before start time should do nothing, got %+v", r)
	}
	// The first block at or after the start time starts epoch 1. No
	// AfterEpochEnd fires: there is no epoch 0 to end.
	r := timer.Tick(genesis, 1)
	if r.StartedEpoch != 1 || r.EndedEpoch != 0 || r.HookFired {
		t.Fatalf("first tick should start epoch 1 only, got %+v", r)
	}
	if len(hook.before) != 1 || hook.before[0] != 1 {
		t.Fatalf("BeforeEpochStart should have fired once for epoch 1, got %v", hook.before)
	}
	if len(hook.after) != 0 {
		t.Fatalf("AfterEpochEnd must not fire for the initial start, got %v", hook.after)
	}
	if got := timer.Info().CurrentEpochStartTime; !got.Equal(genesis) {
		t.Fatalf("epoch 1 start time should be the genesis time, got %v", got)
	}
}

func TestEpochTimerGridAnchoredAndOnePerBlock(t *testing.T) {
	genesis := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	timer := NewEpochTimer(EpochConfig{Identifier: EpochIdentifierMonetary, DurationSeconds: 30}, genesis)
	hook := &recordingHook{}
	timer.RegisterHook(hook)

	timer.Tick(genesis, 1)
	// Block time exactly at the end time does NOT end the epoch: the
	// tick requires block time strictly after the end time.
	if r := timer.Tick(genesis.Add(30*time.Second), 2); r.EndedEpoch != 0 {
		t.Fatalf("block at exactly the end time must not tick, got %+v", r)
	}
	// One second past the end time ends epoch 1 and starts epoch 2.
	// The new epoch starts at the prior end time (genesis + 30s),
	// not at the ticking block time.
	r := timer.Tick(genesis.Add(31*time.Second), 3)
	if r.EndedEpoch != 1 || r.StartedEpoch != 2 || !r.HookFired {
		t.Fatalf("expected epoch 1 to end and epoch 2 to start, got %+v", r)
	}
	if got := timer.Info().CurrentEpochStartTime; !got.Equal(genesis.Add(30 * time.Second)) {
		t.Fatalf("epoch 2 must start at the prior end time, got %v", got)
	}
	if len(hook.after) != 1 || hook.after[0] != 1 {
		t.Fatalf("AfterEpochEnd should have fired once for epoch 1, got %v", hook.after)
	}
}

func TestEpochTimerHaltCatchUpOnePerBlock(t *testing.T) {
	genesis := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	timer := NewEpochTimer(EpochConfig{Identifier: EpochIdentifierMonetary, DurationSeconds: 30}, genesis)
	hook := &recordingHook{}
	timer.RegisterHook(hook)

	timer.Tick(genesis, 1)
	timer.Tick(genesis.Add(31*time.Second), 2) // epoch 1 ends, epoch 2 starts
	// Halt: jump 150 seconds (5 epoch durations) in one block. Only
	// ONE epoch may tick in this block.
	r := timer.Tick(genesis.Add(181*time.Second), 3)
	if r.EndedEpoch != 2 || r.StartedEpoch != 3 {
		t.Fatalf("halted jump must tick exactly one epoch (2 ends, 3 starts), got %+v", r)
	}
	// Subsequent blocks catch up one epoch each, in order.
	for i, wantEnd := range []int64{3, 4, 5, 6} {
		r := timer.Tick(genesis.Add(time.Duration(182+i)*time.Second), int64(4+i))
		if r.EndedEpoch != wantEnd || r.StartedEpoch != wantEnd+1 {
			t.Fatalf("catch-up block %d: expected epoch %d to end, got %+v", i, wantEnd, r)
		}
	}
	if len(hook.after) != 6 {
		t.Fatalf("AfterEpochEnd should have fired 6 times total, got %v", hook.after)
	}
	for i, want := range []int64{1, 2, 3, 4, 5, 6} {
		if hook.after[i] != want {
			t.Fatalf("hook order wrong at %d: got %d want %d", i, hook.after[i], want)
		}
	}
}

func TestEpochTimerHookErrorDoesNotHaltEpoch(t *testing.T) {
	genesis := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	timer := NewEpochTimer(EpochConfig{Identifier: EpochIdentifierMonetary, DurationSeconds: 30}, genesis)
	hook := &recordingHook{failAfter: map[int64]error{1: errors.New("boom")}}
	timer.RegisterHook(hook)

	timer.Tick(genesis, 1)
	r := timer.Tick(genesis.Add(31*time.Second), 2)
	if r.HookErr == nil {
		t.Fatal("expected the hook error to be reported in the tick result")
	}
	// Upstream behaviour: the epoch still advances past a failed hook.
	if r.EndedEpoch != 1 || r.StartedEpoch != 2 {
		t.Fatalf("epoch must advance past a failed hook, got %+v", r)
	}
	if timer.State().LastHookError == "" {
		t.Fatal("failed hook must be recorded in LastHookError, not swallowed silently")
	}
}

func TestMintScheduledForEpochExactOnce(t *testing.T) {
	l := testEpochLedger(t)
	// Epoch 1 mints period 0: 1000. Epoch 2 mints period 1: 900.
	if err := l.MintScheduledForEpoch(1); err != nil {
		t.Fatalf("epoch 1 mint: %v", err)
	}
	if got := l.BalanceOf(RewardsPool); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("rewards pool after epoch 1: got %s want 1000", got)
	}
	if err := l.MintScheduledForEpoch(2); err != nil {
		t.Fatalf("epoch 2 mint: %v", err)
	}
	if got := l.TotalIssuedAmount(); !got.Equal(math.NewInt(1900)) {
		t.Fatalf("total issued after epochs 1-2: got %s want 1900", got)
	}
	// Re-claiming epoch 1 fails and changes nothing.
	before := l.Clone()
	if err := l.MintScheduledForEpoch(1); !errors.Is(err, ErrDoubleIssuance) {
		t.Fatalf("re-claim of epoch 1 should fail with ErrDoubleIssuance, got %v", err)
	}
	if !l.Equal(before) {
		t.Fatal("failed re-claim changed ledger state")
	}
	// Epoch 0 does not exist.
	if err := l.MintScheduledForEpoch(0); err == nil {
		t.Fatal("epoch 0 mint should fail")
	}
	// Epochs 3 and 4 mint 810 and 729; epoch 5 is past the last period.
	if err := l.MintScheduledForEpoch(3); err != nil {
		t.Fatalf("epoch 3 mint: %v", err)
	}
	if err := l.MintScheduledForEpoch(4); err != nil {
		t.Fatalf("epoch 4 mint: %v", err)
	}
	if got := l.TotalIssuedAmount(); !got.Equal(math.NewInt(3439)) {
		t.Fatalf("total issued after all epochs: got %s want 3439", got)
	}
	if err := l.MintScheduledForEpoch(5); !errors.Is(err, ErrScheduleExhausted) {
		t.Fatalf("epoch 5 should fail with ErrScheduleExhausted, got %v", err)
	}
}

func TestMintScheduledForEpochSharesSetWithBlockPath(t *testing.T) {
	l := testEpochLedger(t)
	// The block-based path claims period 0 (height inside period 0).
	if err := l.MintScheduled(1); err != nil {
		t.Fatalf("block path mint: %v", err)
	}
	// The epoch path for the same period must see the claim and fail.
	if err := l.MintScheduledForEpoch(1); !errors.Is(err, ErrDoubleIssuance) {
		t.Fatalf("epoch claim after block claim should fail with ErrDoubleIssuance, got %v", err)
	}
	if got := l.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("total issued should be exactly 1000, got %s", got)
	}
}

func TestMonetaryEpochHookFiltersIdentifier(t *testing.T) {
	l := testEpochLedger(t)
	hook := &MonetaryEpochHook{Ledger: l}
	// A different identifier does nothing.
	if err := hook.AfterEpochEnd("daily", 1); err != nil {
		t.Fatalf("other identifier should be ignored, got %v", err)
	}
	if got := l.TotalIssuedAmount(); !got.IsZero() {
		t.Fatalf("other identifier must not mint, total issued %s", got)
	}
	// The monetary identifier mints.
	if err := hook.AfterEpochEnd(EpochIdentifierMonetary, 1); err != nil {
		t.Fatalf("monetary identifier should mint, got %v", err)
	}
	if got := l.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("total issued after monetary epoch 1: got %s want 1000", got)
	}
	// An exhausted epoch succeeds without minting (epochs keep ticking
	// after issuance ends).
	if err := hook.AfterEpochEnd(EpochIdentifierMonetary, 99); err != nil {
		t.Fatalf("exhausted epoch should succeed without minting, got %v", err)
	}
	if got := l.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("exhausted epoch must not mint, total issued %s", got)
	}
}

func TestEpochConfigValidate(t *testing.T) {
	if err := (EpochConfig{Identifier: "monetary", DurationSeconds: 30}).Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	if err := (EpochConfig{Identifier: "", DurationSeconds: 30}).Validate(); err == nil {
		t.Fatal("empty identifier should fail")
	}
	if err := (EpochConfig{Identifier: "monetary", DurationSeconds: 0}).Validate(); err == nil {
		t.Fatal("zero duration should fail")
	}
	if !(EpochConfig{DurationSeconds: 0}).Disabled() {
		t.Fatal("zero duration should mean disabled")
	}
}
