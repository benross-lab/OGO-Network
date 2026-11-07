package monetary

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
)

// testLedger builds a ledger with TEST-ONLY max 1_000_006_511 uogo:
// two funded accounts totalling 1_000_000_000 plus the exact TEST-ONLY
// schedule total (6_511). Milestone C strict genesis: max is exactly
// G + S_actual, no headroom. The numbers are test parameters only,
// not the production supply (TBD).
func testLedger(t *testing.T) *Ledger {
	t.Helper()
	max := math.NewInt(1_000_006_511) // TEST-ONLY
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(600_000_000), Purpose: "public"},
		{Address: "ogo1bob", Denom: Denom, Amount: math.NewInt(400_000_000), Purpose: "public"},
	})
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("test ledger setup failed: %v", err)
	}
	return l
}

// mustLedger builds a small ledger (TEST-ONLY max, 10_000 uogo genesis
// supply) for schedule tests.
func mustLedger(t *testing.T, max math.Int) *Ledger {
	t.Helper()
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10_000), Purpose: "public"},
	})
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("ledger setup failed: %v", err)
	}
	return l
}

func TestDiscretionaryMintProhibited(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.Mint("ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("discretionary Mint must always fail")
	}
	if err := l.Mint("ogo1alice", math.NewInt(1_000_000)); err == nil {
		t.Fatal("large discretionary Mint must always fail")
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("failed mint changed supply: before %s after %s", before, l.TotalSupply())
	}
}

func TestTransferPreservesSupply(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.Send("ogo1alice", "ogo1bob", math.NewInt(123_456)); err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("transfer changed total supply: before %s after %s", before, l.TotalSupply())
	}
	if !l.BalanceOf("ogo1alice").Equal(math.NewInt(600_000_000 - 123_456)) {
		t.Fatalf("sender balance wrong: %s", l.BalanceOf("ogo1alice"))
	}
	if !l.BalanceOf("ogo1bob").Equal(math.NewInt(400_000_000 + 123_456)) {
		t.Fatalf("receiver balance wrong: %s", l.BalanceOf("ogo1bob"))
	}
}

func TestTransferInsufficientFailsAndChangesNothing(t *testing.T) {
	l := testLedger(t)
	before := l.Clone()
	if err := l.Send("ogo1alice", "ogo1bob", math.NewInt(600_000_001)); err == nil {
		t.Fatal("overspend transfer must fail")
	}
	if !l.Equal(before) {
		t.Fatal("failed transfer changed state")
	}
	if err := l.Send("ogo1alice", "ogo1bob", math.NewInt(0)); err == nil {
		t.Fatal("zero transfer must fail")
	}
	if err := l.Send("ogo1alice", "ogo1bob", math.NewInt(-1)); err == nil {
		t.Fatal("negative transfer must fail")
	}
}

func TestBurnIsOneWay(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.Burn("ogo1alice", math.NewInt(1_000_000)); err != nil {
		t.Fatalf("burn failed: %v", err)
	}
	if !l.TotalSupply().Equal(before.Sub(math.NewInt(1_000_000))) {
		t.Fatalf("burn must reduce supply by exactly the burned amount: %s", l.TotalSupply())
	}
	if !l.Burned.Equal(math.NewInt(1_000_000)) {
		t.Fatalf("burned counter wrong: %s", l.Burned)
	}
	// No remint: discretionary Mint still fails after burning, and the
	// burned amount is gone. (Scheduled issuance is unaffected by burns;
	// it is governed only by the schedule and the cap.)
	if err := l.Mint("ogo1alice", math.NewInt(1_000_000)); err == nil {
		t.Fatal("remint after burn must fail")
	}
	if !l.TotalSupply().Equal(before.Sub(math.NewInt(1_000_000))) {
		t.Fatal("remint attempt changed supply")
	}
}

func TestFeesDoNotMint(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.DeductFee("ogo1alice", math.NewInt(5_000)); err != nil {
		t.Fatalf("fee deduction failed: %v", err)
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("fee changed total supply: before %s after %s", before, l.TotalSupply())
	}
	if !l.BalanceOf(FeeCollector).Equal(math.NewInt(5_000)) {
		t.Fatalf("fee collector balance wrong: %s", l.BalanceOf(FeeCollector))
	}
}

func TestHeadroomOnlyViaSchedule(t *testing.T) {
	l := testLedger(t)
	// Headroom = max - genesis = 6_511 uogo, exactly the schedule total.
	// No one may spend it directly: discretionary mint of the full
	// headroom fails.
	if !l.Headroom().Equal(math.NewInt(6_511)) {
		t.Fatalf("headroom wrong: %s", l.Headroom())
	}
	if err := l.Mint("ogo1alice", l.Headroom()); err == nil {
		t.Fatal("minting the full headroom must fail")
	}
	// Only the schedule may draw on headroom, and only its exact
	// amounts: total scheduled issuance is 6_511, all of the headroom.
	for p := uint64(0); p < 10; p++ {
		if err := l.MintScheduled(int64(1) + int64(p)*100); err != nil {
			t.Fatalf("period %d scheduled issuance failed: %v", p, err)
		}
	}
	if got := l.TotalIssued; !got.Equal(math.NewInt(6_511)) {
		t.Fatalf("total issued: got %s, want 6511", got)
	}
	if !l.Headroom().Equal(math.NewInt(0)) {
		t.Fatalf("headroom after full schedule: got %s, want 0", l.Headroom())
	}
}

func TestRandomTransfersNeverExceedMax(t *testing.T) {
	l := testLedger(t)
	rng := rand.New(rand.NewSource(42))
	addrs := []string{"ogo1alice", "ogo1bob", "ogo1carol", "ogo1dave"}
	for i := 0; i < 2000; i++ {
		from := addrs[rng.Intn(len(addrs))]
		to := addrs[rng.Intn(len(addrs))]
		amt := math.NewInt(int64(rng.Intn(10_000)))
		if amt.IsZero() {
			continue
		}
		_ = l.Send(from, to, amt) // failures (insufficient) are expected and fine
		if err := l.CheckInvariants(); err != nil {
			t.Fatalf("invariant violated after %d random transfers: %v", i, err)
		}
	}
	if l.TotalSupply().GT(l.MaxSupply) {
		t.Fatalf("total %s exceeds max %s", l.TotalSupply(), l.MaxSupply)
	}
}

func TestRandomBurnsNeverCreateSupply(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	rng := rand.New(rand.NewSource(7))
	addrs := []string{"ogo1alice", "ogo1bob"}
	for i := 0; i < 500; i++ {
		_ = l.Burn(addrs[rng.Intn(2)], math.NewInt(int64(rng.Intn(1000)+1)))
		if l.TotalSupply().GT(before) {
			t.Fatalf("burn increased supply at step %d", i)
		}
	}
}
