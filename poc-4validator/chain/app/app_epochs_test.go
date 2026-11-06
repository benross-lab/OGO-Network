// Tests for the epoch timer mirror as wired into the application.
//
// The critical test here is TestEpochHookFiresInWiredApp: it proves
// the x/monetary AfterEpochEnd hook actually executes in the wired
// application (issuance minted, epoch event emitted) when the chain
// advances past an epoch boundary. A compile check is not sufficient
// for this: the depinject hook-registration bug (SDK issues #24770,
// #25425) lets a module compile and run while its hook never fires.
// This test advances the wired app past epoch boundaries and asserts
// on ledger state and on FinalizeBlock events, not on wiring.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"

	ogocoin "github.com/benross-lab/ogo-network/x/ogocoin"
)

// testGenesisTime is the fixed genesis time for epoch tests, so epoch
// boundaries are exact and the tests never depend on wall-clock time.
var testGenesisTime = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

// testAppState builds genesis app_state JSON with the TEST-ONLY
// schedule (4 periods, initial 1000, reduction 9/10, total 3439) and,
// unless epochSeconds is 0, an epochs config with the given duration.
func testAppState(t *testing.T, epochSeconds int64) []byte {
	t.Helper()
	max := math.NewInt(1000010000)
	gs := ogocoin.GenesisState{
		Params: ogocoin.Params{
			MaxSupply:         &max,
			Denom:             ogocoin.Denom,
			Decimals:          ogocoin.Decimals,
			InflationRate:     "0",
			MintingEnabled:    false,
			SupplyCapEnforced: true,
			IssuanceSchedule: &ogocoin.IssuanceSchedule{
				ScheduleType:         ogocoin.ScheduleTypeReducingPerPeriod,
				StartHeight:          1,
				PeriodLength:         100,
				Periods:              4,
				InitialPerPeriod:     math.NewInt(1000),
				ReductionNumerator:   9,
				ReductionDenominator: 10,
			},
		},
		Balances: []ogocoin.Balance{
			{Address: "ogo1alice", Denom: ogocoin.Denom, Amount: math.NewInt(600000000), Purpose: "public"},
			{Address: "ogo1bob", Denom: ogocoin.Denom, Amount: math.NewInt(400000000), Purpose: "public"},
		},
	}
	appState := map[string]any{"ogocoin": gs}
	if epochSeconds > 0 {
		appState["epochs"] = ogocoin.EpochConfig{Identifier: ogocoin.EpochIdentifierMonetary, DurationSeconds: epochSeconds}
	}
	b, err := json.Marshal(appState)
	if err != nil {
		t.Fatalf("marshal app_state: %v", err)
	}
	return b
}

// testInitApp creates an app in a temp home and runs InitChain.
func testInitApp(t *testing.T, epochSeconds int64) *OgoApp {
	t.Helper()
	a := NewOgoApp(log.NewNopLogger(), t.TempDir())
	_, err := a.InitChain(context.Background(), &abcitypes.RequestInitChain{
		ChainId:       ogocoin.ChainID,
		Time:          testGenesisTime,
		AppStateBytes: testAppState(t, epochSeconds),
	})
	if err != nil {
		t.Fatalf("InitChain: %v", err)
	}
	return a
}

// finalize runs one block at the given height and offset from genesis
// time, and returns the FinalizeBlock response.
func finalize(t *testing.T, a *OgoApp, height int64, offset time.Duration) *abcitypes.ResponseFinalizeBlock {
	t.Helper()
	res, err := a.FinalizeBlock(context.Background(), &abcitypes.RequestFinalizeBlock{
		Height: height,
		Time:   testGenesisTime.Add(offset),
	})
	if err != nil {
		t.Fatalf("FinalizeBlock height %d: %v", height, err)
	}
	return res
}

// findEvent returns the first event of the given type, if any.
func findEvent(res *abcitypes.ResponseFinalizeBlock, typ string) (abcitypes.Event, bool) {
	for _, e := range res.Events {
		if e.Type == typ {
			return e, true
		}
	}
	return abcitypes.Event{}, false
}

// eventAttr reads one attribute value from an event.
func eventAttr(e abcitypes.Event, key string) string {
	for _, attr := range e.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

// TestEpochHookFiresInWiredApp is the CRITICAL-PATH test: advance the
// wired application past epoch boundaries and prove the monetary
// AfterEpochEnd hook executes (issuance minted, epoch events
// emitted). Epoch duration is 30s (TEST-ONLY).
func TestEpochHookFiresInWiredApp(t *testing.T) {
	a := testInitApp(t, 30)

	// Height 1, 5s after genesis: epoch 1 starts. No issuance yet:
	// issuance happens at epoch END, not at epoch start.
	res := finalize(t, a, 1, 5*time.Second)
	if e, ok := findEvent(res, "epoch_start"); !ok || eventAttr(e, "epoch_number") != "1" {
		t.Fatalf("height 1 should emit epoch_start for epoch 1, events: %+v", res.Events)
	}
	if _, ok := findEvent(res, "epoch_end"); ok {
		t.Fatal("height 1 must not end any epoch")
	}
	if got := a.ledger.TotalIssuedAmount(); !got.IsZero() {
		t.Fatalf("no issuance before the first epoch ends, got %s", got)
	}

	// Heights 2-6 (10s-30s): the epoch end time is genesis+30s, and
	// the tick needs block time STRICTLY after it, so nothing ticks.
	for h := int64(2); h <= 6; h++ {
		res := finalize(t, a, h, time.Duration(h*5)*time.Second)
		if _, ok := findEvent(res, "epoch_end"); ok {
			t.Fatalf("height %d must not end any epoch", h)
		}
	}

	// Height 7, 35s after genesis: epoch 1 ends and epoch 2 starts.
	// The hook MUST have fired: 1000 uogo minted into the rewards
	// pool, and the epoch events emitted.
	res = finalize(t, a, 7, 35*time.Second)
	endEv, ok := findEvent(res, "epoch_end")
	if !ok || eventAttr(endEv, "epoch_number") != "1" {
		t.Fatalf("height 7 should emit epoch_end for epoch 1, events: %+v", res.Events)
	}
	startEv, ok := findEvent(res, "epoch_start")
	if !ok || eventAttr(startEv, "epoch_number") != "2" {
		t.Fatalf("height 7 should emit epoch_start for epoch 2, events: %+v", res.Events)
	}
	issEv, ok := findEvent(res, "epoch_issuance")
	if !ok || eventAttr(issEv, "amount") != "1000" || eventAttr(issEv, "epoch_number") != "1" {
		t.Fatalf("height 7 should emit epoch_issuance of 1000 for epoch 1, events: %+v", res.Events)
	}
	if got := a.ledger.BalanceOf(ogocoin.RewardsPool); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("rewards pool after epoch 1: got %s want 1000", got)
	}
	if got := a.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("total issued after epoch 1: got %s want 1000", got)
	}
	if got := a.epochTimer.State().HookFireCount; got != 1 {
		t.Fatalf("hook fire count after epoch 1: got %d want 1", got)
	}
	if got := a.epochTimer.Info().CurrentEpoch; got != 2 {
		t.Fatalf("current epoch after height 7: got %d want 2", got)
	}

	// Height 13, 65s after genesis: epoch 2 ends (900), epoch 3 starts.
	for h := int64(8); h <= 12; h++ {
		finalize(t, a, h, time.Duration(h*5)*time.Second)
	}
	res = finalize(t, a, 13, 65*time.Second)
	issEv, ok = findEvent(res, "epoch_issuance")
	if !ok || eventAttr(issEv, "amount") != "900" || eventAttr(issEv, "epoch_number") != "2" {
		t.Fatalf("height 13 should emit epoch_issuance of 900 for epoch 2, events: %+v", res.Events)
	}
	if got := a.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1900)) {
		t.Fatalf("total issued after epoch 2: got %s want 1900", got)
	}
}

// TestEpochHaltCatchUpInWiredApp proves the S4 rule in the wired app:
// after a halt, missed epochs tick one per block, in epoch order,
// with no lump mint, no skipped epoch, and no double mint.
func TestEpochHaltCatchUpInWiredApp(t *testing.T) {
	a := testInitApp(t, 30)

	finalize(t, a, 1, 5*time.Second)   // epoch 1 starts
	finalize(t, a, 7, 35*time.Second)  // epoch 1 ends (1000), epoch 2 starts
	finalize(t, a, 13, 65*time.Second) // epoch 2 ends (900), epoch 3 starts

	// Halt: jump 135 seconds (4.5 epoch durations) in a single block.
	// Only ONE epoch may tick in this block: epoch 3 ends (810),
	// epoch 4 starts. No lump mint of several epochs at once.
	res := finalize(t, a, 14, 200*time.Second)
	endEv, ok := findEvent(res, "epoch_end")
	if !ok || eventAttr(endEv, "epoch_number") != "3" {
		t.Fatalf("halted block must end exactly epoch 3, events: %+v", res.Events)
	}
	issEv, ok := findEvent(res, "epoch_issuance")
	if !ok || eventAttr(issEv, "amount") != "810" {
		t.Fatalf("halted block must mint exactly epoch 3 (810), events: %+v", res.Events)
	}
	if got := a.epochTimer.Info().CurrentEpoch; got != 4 {
		t.Fatalf("current epoch after halted block: got %d want 4", got)
	}

	// Catch-up: one epoch per block, in order.
	res = finalize(t, a, 15, 205*time.Second) // epoch 4 ends (729)
	if e, ok := findEvent(res, "epoch_end"); !ok || eventAttr(e, "epoch_number") != "4" {
		t.Fatalf("catch-up block 15 must end epoch 4, events: %+v", res.Events)
	}
	res = finalize(t, a, 16, 210*time.Second) // epoch 5 ends: exhausted, no mint
	if e, ok := findEvent(res, "epoch_end"); !ok || eventAttr(e, "epoch_number") != "5" {
		t.Fatalf("catch-up block 16 must end epoch 5, events: %+v", res.Events)
	}
	if _, ok := findEvent(res, "epoch_issuance"); ok {
		t.Fatal("epoch 5 is past the last period and must mint nothing")
	}
	res = finalize(t, a, 17, 215*time.Second) // epoch 6 ends: exhausted, no mint
	if e, ok := findEvent(res, "epoch_end"); !ok || eventAttr(e, "epoch_number") != "6" {
		t.Fatalf("catch-up block 17 must end epoch 6, events: %+v", res.Events)
	}
	if got := a.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(3439)) {
		t.Fatalf("total issued after catch-up: got %s want 3439 (1000+900+810+729)", got)
	}
	if got := a.epochTimer.Info().CurrentEpoch; got != 7 {
		t.Fatalf("current epoch after catch-up: got %d want 7", got)
	}
	// The issued-by-epoch set has exactly periods 0-3, each once.
	issued := a.ledger.IssuedByEpochCopy()
	if len(issued) != 4 {
		t.Fatalf("issued-by-epoch set should have 4 entries, got %v", issued)
	}
	// A re-claim of an already issued epoch fails.
	if err := a.ledger.MintScheduledForEpoch(1); !errors.Is(err, ogocoin.ErrDoubleIssuance) {
		t.Fatalf("re-claim of epoch 1 should fail with ErrDoubleIssuance, got %v", err)
	}
}

// TestEpochsQueryPath checks the "epochs" ABCI query returns the
// timer state and the issued-by-epoch set.
func TestEpochsQueryPath(t *testing.T) {
	a := testInitApp(t, 30)
	finalize(t, a, 1, 5*time.Second)
	finalize(t, a, 7, 35*time.Second) // epoch 1 ends, 1000 minted

	res, err := a.Query(context.Background(), &abcitypes.RequestQuery{Path: "epochs"})
	if err != nil {
		t.Fatalf("Query epochs: %v", err)
	}
	if res.Code != 0 {
		t.Fatalf("Query epochs failed: %s", res.Log)
	}
	var out map[string]string
	if err := json.Unmarshal(res.Value, &out); err != nil {
		t.Fatalf("epochs query is not flat JSON: %v", err)
	}
	if out["enabled"] != "true" || out["identifier"] != "monetary" || out["current_epoch"] != "2" {
		t.Fatalf("unexpected epochs query output: %v", out)
	}
	if out["total_issued"] != "1000" || out["issued_by_period"] != "0:1000" {
		t.Fatalf("unexpected issuance in epochs query: %v", out)
	}
	if out["implementation"] == "" {
		t.Fatal("epochs query must state its implementation honestly")
	}
}

// TestEpochDisabledFallsBackToBlockSchedule proves the legacy path
// still works when genesis carries no epochs config: issuance runs
// per block period exactly as in Milestone B.
func TestEpochDisabledFallsBackToBlockSchedule(t *testing.T) {
	a := testInitApp(t, 0)
	if a.epochTimer != nil {
		t.Fatal("epoch timer must be nil when genesis has no epochs config")
	}
	finalize(t, a, 1, 5*time.Second) // period 0 (100 blocks): mints 1000
	if got := a.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("legacy path should mint 1000 at height 1, got %s", got)
	}
	finalize(t, a, 2, 10*time.Second) // same period: no more mint
	if got := a.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("legacy path must not re-mint in the same period, got %s", got)
	}
	finalize(t, a, 101, 505*time.Second) // period 1: mints 900
	if got := a.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1900)) {
		t.Fatalf("legacy path should mint 900 at height 101, got %s", got)
	}
	// The epochs query reports disabled, honestly.
	res, err := a.Query(context.Background(), &abcitypes.RequestQuery{Path: "epochs"})
	if err != nil {
		t.Fatalf("Query epochs: %v", err)
	}
	var out map[string]string
	if err := json.Unmarshal(res.Value, &out); err != nil {
		t.Fatalf("epochs query is not flat JSON: %v", err)
	}
	if out["enabled"] != "false" {
		t.Fatalf("epochs query should report disabled, got %v", out)
	}
}

// TestAppRestartRestoresHeightAndEpochs proves a node restart does
// not replay blocks or double-mint: height and epoch state persist,
// and finalizing continues exactly where the old instance stopped.
func TestAppRestartRestoresHeightAndEpochs(t *testing.T) {
	home := t.TempDir()
	a1 := NewOgoApp(log.NewNopLogger(), home)
	if _, err := a1.InitChain(context.Background(), &abcitypes.RequestInitChain{
		ChainId:       ogocoin.ChainID,
		Time:          testGenesisTime,
		AppStateBytes: testAppState(t, 30),
	}); err != nil {
		t.Fatalf("InitChain: %v", err)
	}
	finalize(t, a1, 1, 5*time.Second)
	finalize(t, a1, 7, 35*time.Second) // epoch 1 ends, 1000 minted

	// A fresh app instance on the same home must restore height 7,
	// epoch 2 current, and 1000 issued, from disk.
	a2 := NewOgoApp(log.NewNopLogger(), home)
	info, err := a2.Info(context.Background(), &abcitypes.RequestInfo{})
	if err != nil {
		t.Fatalf("Info after restart: %v", err)
	}
	if info.LastBlockHeight != 7 {
		t.Fatalf("restored height: got %d want 7", info.LastBlockHeight)
	}
	if info.LastBlockAppHash == nil || len(info.LastBlockAppHash) == 0 {
		t.Fatal("restored app hash must not be empty")
	}
	if fmt.Sprintf("%X", info.LastBlockAppHash) != fmt.Sprintf("%X", a1.appHash) {
		t.Fatal("restored app hash must match the pre-restart app hash")
	}
	if got := a2.epochTimer.Info().CurrentEpoch; got != 2 {
		t.Fatalf("restored current epoch: got %d want 2", got)
	}
	if got := a2.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1000)) {
		t.Fatalf("restored total issued: got %s want 1000", got)
	}

	// Continuing must not re-mint epoch 1: height 13 ends epoch 2
	// (900) and only epoch 2.
	finalize(t, a2, 13, 65*time.Second)
	if got := a2.ledger.TotalIssuedAmount(); !got.Equal(math.NewInt(1900)) {
		t.Fatalf("total issued after restart and epoch 2: got %s want 1900", got)
	}
	info, err = a2.Info(context.Background(), &abcitypes.RequestInfo{})
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.LastBlockHeight != 13 {
		t.Fatalf("height after continued blocks: got %d want 13", info.LastBlockHeight)
	}
}
