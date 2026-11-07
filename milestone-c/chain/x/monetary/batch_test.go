package monetary

import (
	"testing"

	"cosmossdk.io/math"
)

// Milestone 4 batching tests. The fee is the prototype's flat fee
// per transaction (100 uogo, TEST-ONLY, the Milestone B localnet
// fee). There is no gas metering in this scaffold, so the measured
// numbers are measurements of the flat-fee model as implemented.
// See chain/BATCHING_RESULTS.md and batch.go for the honesty note.

func batchLedger(t *testing.T) *Ledger {
	t.Helper()
	max := math.NewInt(1006511) // TEST-ONLY: G 1_000_000 + exact S_actual 6_511
	gs := testGenesis(max, []Balance{
		{Address: "ogo1agent", Denom: Denom, Amount: math.NewInt(1000000), Purpose: "public"},
	})
	// Milestone C strict genesis: max is exactly G + S_actual.
	l, err := NewLedger(gs)
	if err != nil {
		t.Fatalf("ledger setup failed: %v", err)
	}
	return l
}

func hundredPayments() []Payment {
	payments := make([]Payment, 100)
	for i := range payments {
		payments[i] = Payment{To: "ogo1recipient", Amount: math.NewInt(1000)}
	}
	return payments
}

func TestBatchSendPreservesSupplyAndChargesOneFee(t *testing.T) {
	l := batchLedger(t)
	before := l.TotalSupply()
	payments := hundredPayments()
	if err := l.BatchSend("ogo1agent", payments, FlatFeePerTx); err != nil {
		t.Fatalf("batch send failed: %v", err)
	}
	if !l.TotalSupply().Equal(before) {
		t.Fatalf("batch changed total supply: before %s after %s", before, l.TotalSupply())
	}
	// 100 payments of 1000 to the same recipient: 100000 received.
	if got := l.BalanceOf("ogo1recipient"); !got.Equal(math.NewInt(100000)) {
		t.Fatalf("recipient balance: got %s, want 100000", got)
	}
	// One fee only: fee collector holds exactly 100.
	if got := l.BalanceOf(FeeCollector); !got.Equal(math.NewInt(100)) {
		t.Fatalf("fee collector: got %s, want 100 (one fee for 100 payments)", got)
	}
	// Sender: 1000000 - 100000 payments - 100 fee = 899900.
	if got := l.BalanceOf("ogo1agent"); !got.Equal(math.NewInt(899900)) {
		t.Fatalf("sender balance: got %s, want 899900", got)
	}
}

func TestBatchSendAtomicOnFailure(t *testing.T) {
	l := batchLedger(t)
	before := l.Clone()
	// One payment is zero: the whole batch must fail, nothing changes.
	payments := hundredPayments()
	payments[50].Amount = math.NewInt(0)
	if err := l.BatchSend("ogo1agent", payments, FlatFeePerTx); err == nil {
		t.Fatal("batch with a zero payment must fail")
	}
	if !l.Equal(before) {
		t.Fatal("failed batch changed state")
	}
	// Insufficient sender balance for payments + fee: whole batch fails.
	payments = []Payment{{To: "ogo1recipient", Amount: math.NewInt(2000000)}}
	if err := l.BatchSend("ogo1agent", payments, FlatFeePerTx); err == nil {
		t.Fatal("batch exceeding sender balance must fail")
	}
	if !l.Equal(before) {
		t.Fatal("failed overspend batch changed state")
	}
	// Empty batch fails.
	if err := l.BatchSend("ogo1agent", nil, FlatFeePerTx); err == nil {
		t.Fatal("empty batch must fail")
	}
}

func TestMeasureBatchHundredPayments(t *testing.T) {
	// 100 agent payments of 0.001 OGO (1000 uogo) each, fee 100 uogo
	// per transaction. Measured through the ledger's own paths.
	payments := hundredPayments()
	unbatched, batched, err := MeasureBatch(batchLedger(t), batchLedger(t), "ogo1agent", payments, FlatFeePerTx)
	if err != nil {
		t.Fatalf("measurement failed: %v", err)
	}
	// Unbatched: 100 transactions, 100 fees of 100 = 10000 total,
	// 100 per payment, 1000 basis points = 10.00 percent of a
	// 1000 uogo payment.
	if !unbatched.TotalFee.Equal(math.NewInt(10000)) {
		t.Fatalf("unbatched total fee: got %s, want 10000", unbatched.TotalFee)
	}
	if !unbatched.FeePerPayment.Equal(math.NewInt(100)) {
		t.Fatalf("unbatched fee per payment: got %s, want 100", unbatched.FeePerPayment)
	}
	if !unbatched.FeeBasisPoints.Equal(math.NewInt(1000)) {
		t.Fatalf("unbatched fee: got %s bp, want 1000 bp (10.00 percent)", unbatched.FeeBasisPoints)
	}
	// Batched: 1 transaction, 1 fee of 100, 1 per payment,
	// 10 basis points = 0.10 percent of a 1000 uogo payment.
	if !batched.TotalFee.Equal(math.NewInt(100)) {
		t.Fatalf("batched total fee: got %s, want 100", batched.TotalFee)
	}
	if !batched.FeePerPayment.Equal(math.NewInt(1)) {
		t.Fatalf("batched fee per payment: got %s, want 1", batched.FeePerPayment)
	}
	if !batched.FeeBasisPoints.Equal(math.NewInt(10)) {
		t.Fatalf("batched fee: got %s bp, want 10 bp (0.10 percent)", batched.FeeBasisPoints)
	}
	if batched.Transactions != 1 || unbatched.Transactions != 100 {
		t.Fatalf("transaction counts: unbatched %d, batched %d", unbatched.Transactions, batched.Transactions)
	}
}
