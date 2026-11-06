package ogocoin

import (
	"testing"

	"cosmossdk.io/math"
)

const testChannel = "channel-0" // TEST-ONLY channel id

func TestIBCEscrowDoesNotChangeLocalSupply(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.EscrowForIBC("ogo1alice", testChannel, math.NewInt(50_000_000)); err != nil {
		t.Fatalf("escrow failed: %v", err)
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("escrow changed local supply: before %s after %s", before, l.TotalSupply())
	}
	if !l.Escrowed(testChannel).Equal(math.NewInt(50_000_000)) {
		t.Fatalf("escrowed amount wrong: %s", l.Escrowed(testChannel))
	}
}

func TestIBCVoucherIsNotNewSupply(t *testing.T) {
	l := testLedger(t)
	localBefore := l.TotalSupply()
	if err := l.EscrowForIBC("ogo1alice", testChannel, math.NewInt(50_000_000)); err != nil {
		t.Fatalf("escrow failed: %v", err)
	}
	if err := l.RegisterVoucher(testChannel, math.NewInt(50_000_000)); err != nil {
		t.Fatalf("voucher registration failed: %v", err)
	}
	// Global economic supply must equal local supply, NOT local + vouchers.
	global := GlobalEconomicSupply(l.TotalSupply(), l.Escrowed(testChannel), l.VouchersOutstanding(testChannel))
	if !global.Equal(localBefore) {
		t.Fatalf("global economic supply %s != local supply %s: double-counted", global, localBefore)
	}
	naive := NaiveDoubleCount(l.TotalSupply(), l.VouchersOutstanding(testChannel))
	if naive.Equal(global) {
		t.Fatal("naive sum unexpectedly equals global supply; test is not discriminating")
	}
	if !naive.Equal(localBefore.Add(math.NewInt(50_000_000))) {
		t.Fatalf("naive sum should be local+50M, got %s", naive)
	}
	if err := l.CheckGlobalSupply(testChannel); err != nil {
		t.Fatalf("global supply check failed: %v", err)
	}
}

func TestIBCVoucherCannotExceedEscrow(t *testing.T) {
	l := testLedger(t)
	if err := l.EscrowForIBC("ogo1alice", testChannel, math.NewInt(10_000)); err != nil {
		t.Fatalf("escrow failed: %v", err)
	}
	if err := l.RegisterVoucher(testChannel, math.NewInt(10_001)); err == nil {
		t.Fatal("voucher above escrow must fail")
	}
	if err := l.RegisterVoucher(testChannel, math.NewInt(10_000)); err != nil {
		t.Fatalf("voucher equal to escrow must succeed: %v", err)
	}
	// Direct registration with no escrow at all must fail for any positive amount.
	if err := l.RegisterVoucher("channel-99", math.NewInt(1)); err == nil {
		t.Fatal("voucher with zero escrow must fail")
	}
}

func TestIBCReleaseReturnsEscrow(t *testing.T) {
	l := testLedger(t)
	before := l.TotalSupply()
	if err := l.EscrowForIBC("ogo1alice", testChannel, math.NewInt(20_000_000)); err != nil {
		t.Fatalf("escrow failed: %v", err)
	}
	if err := l.RegisterVoucher(testChannel, math.NewInt(20_000_000)); err != nil {
		t.Fatalf("voucher failed: %v", err)
	}
	if err := l.ReleaseEscrow("ogo1bob", testChannel, math.NewInt(20_000_000)); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("release changed local supply: before %s after %s", before, l.TotalSupply())
	}
	if !l.BalanceOf("ogo1bob").Equal(math.NewInt(400_000_000 + 20_000_000)) {
		t.Fatalf("bob balance wrong after release: %s", l.BalanceOf("ogo1bob"))
	}
	if err := l.CheckGlobalSupply(testChannel); err != nil {
		t.Fatalf("global check after release failed: %v", err)
	}
}

func TestIBCReleaseCannotExceedVouchers(t *testing.T) {
	l := testLedger(t)
	if err := l.EscrowForIBC("ogo1alice", testChannel, math.NewInt(10_000)); err != nil {
		t.Fatalf("escrow failed: %v", err)
	}
	if err := l.RegisterVoucher(testChannel, math.NewInt(10_000)); err != nil {
		t.Fatalf("voucher failed: %v", err)
	}
	before := l.Clone()
	if err := l.ReleaseEscrow("ogo1bob", testChannel, math.NewInt(10_001)); err == nil {
		t.Fatal("release above vouchers must fail")
	}
	if !l.Equal(before) {
		t.Fatal("failed release changed state")
	}
}

func TestIBCInboundCappedByMax(t *testing.T) {
	// Ledger at exactly TEST-ONLY max (genesis + full schedule issued):
	// any inbound native OGO must fail.
	max := math.NewInt(16_508) // TEST-ONLY: 10_000 genesis + 6_508 scheduled
	gs := testGenesis(max, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(10_000), Purpose: "public"},
	})
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	for p := uint64(0); p < 10; p++ {
		if err := l.MintScheduled(int64(1) + int64(p)*100); err != nil {
			t.Fatalf("period %d issuance failed: %v", p, err)
		}
	}
	if !l.TotalSupply().Equal(max) {
		t.Fatalf("setup: total %s should equal max %s", l.TotalSupply(), max)
	}
	if err := l.InboundTransfer("ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("inbound transfer over max must fail")
	}
	// Ledger below max: inbound up to headroom succeeds, beyond fails.
	max2 := math.NewInt(1_000_000_000) // TEST-ONLY
	gs = testGenesis(max2, []Balance{
		{Address: "ogo1alice", Denom: Denom, Amount: math.NewInt(100), Purpose: "public"},
	})
	l2, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := l2.InboundTransfer("ogo1alice", math.NewInt(1_000_000_000-100)); err != nil {
		t.Fatalf("inbound up to headroom must succeed: %v", err)
	}
	if err := l2.InboundTransfer("ogo1alice", math.NewInt(1)); err == nil {
		t.Fatal("inbound past max must fail")
	}
}
