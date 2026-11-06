// Batch transfers (Milestone 4): agent transaction economics.
//
// Fee market spec section 7: batching is the single most effective
// micropayment approach and requires no protocol change, because a
// multi-send already exists in x/bank. This prototype implements the
// batch as a ledger-level multi-send: one transaction, one fee,
// many payments.
//
// Fee honesty, recorded here and in BATCHING_RESULTS.md: the
// prototype's fee handling is THIN. There is no gas metering, no gas
// price, and no congestion adjustment in this scaffold. The fee for a
// transaction is a flat amount supplied by the sender (the CLI
// default is 0; the localnet proof used 100 uogo per transaction,
// TEST-ONLY). DeductFee moves that flat fee from payer to the fee
// collector; it never mints. So the measured batching numbers are
// measurements of the flat-fee-per-transaction model as implemented,
// not of a gas-priced fee market. What batching proves here is the
// arithmetic that matters: one fee per transaction instead of one
// fee per payment. A gas-priced fee market (spec sections 3.2-3.3)
// is NOT implemented and is not claimed.
//
// BatchSend is atomic: every payment is validated up front (positive
// amounts, sender balance covers payments plus the single fee), and
// if any validation fails nothing changes. If one payment in the
// batch would fail, the whole batch fails; there is no skip-a-payment
// mode in this prototype, and that open question from the spec
// (whole batch fails vs failing payment skipped) is answered here
// as whole-batch-fails, to be revisited in developer documentation.
package ogocoin

import (
	"fmt"

	"cosmossdk.io/math"
)

// Payment is one payment inside a batch transfer.
type Payment struct {
	To     string   `json:"to"`
	Amount math.Int `json:"amount"`
}

// FlatFeePerTx is the TEST-ONLY flat fee per transaction used by the
// batching demonstration (chain/batchdemo) and its tests. It matches
// the fee used in the Milestone B localnet proof (100 uogo). It is
// NOT a production fee level, which remains TBD.
var FlatFeePerTx = math.NewInt(100)

// FeeForTransaction returns the fee for one transaction under the
// prototype's flat-fee model: exactly the flat fee, regardless of
// the number of payments inside it. That is the whole point of
// batching: the fee does not grow with the payment count.
func FeeForTransaction(flatFee math.Int) math.Int {
	if flatFee.IsNil() || flatFee.IsNegative() {
		return math.NewInt(0)
	}
	return flatFee
}

// BatchSend executes a multi-send in one transaction: one fee, many
// payments, all from the same sender. Total supply is exactly
// preserved (payments move coins, the fee moves coins to the fee
// collector). Atomic: any validation failure changes nothing.
func (l *Ledger) BatchSend(from string, payments []Payment, fee math.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(payments) == 0 {
		return fmt.Errorf("%w: batch must contain at least one payment", ErrZeroAmount)
	}
	if fee.IsNil() || fee.IsNegative() {
		return fmt.Errorf("%w: batch fee %s", ErrNegativeAmount, fee.String())
	}
	totalPayments := math.NewInt(0)
	for i, p := range payments {
		if p.To == "" {
			return fmt.Errorf("ogocoin: batch payment %d has an empty recipient", i)
		}
		if p.Amount.IsNil() || !p.Amount.IsPositive() {
			return fmt.Errorf("%w: batch payment %d amount %s", ErrZeroAmount, i, p.Amount.String())
		}
		totalPayments = totalPayments.Add(p.Amount)
	}
	fromBal := l.Balances[from]
	if fromBal.IsNil() {
		fromBal = math.NewInt(0)
	}
	need := totalPayments.Add(fee)
	if fromBal.LT(need) {
		return fmt.Errorf("%w: %s holds %s, batch needs %s (payments %s + fee %s)",
			ErrInsufficient, from, fromBal.String(), need.String(), totalPayments.String(), fee.String())
	}
	before := l.totalLocked()
	l.Balances[from] = fromBal.Sub(need)
	for _, p := range payments {
		toBal := l.Balances[p.To]
		if toBal.IsNil() {
			toBal = math.NewInt(0)
		}
		l.Balances[p.To] = toBal.Add(p.Amount)
	}
	if !fee.IsZero() {
		collector := l.Balances[FeeCollector]
		if collector.IsNil() {
			collector = math.NewInt(0)
		}
		l.Balances[FeeCollector] = collector.Add(fee)
	}
	if !l.totalLocked().Equal(before) {
		return fmt.Errorf("ogocoin: internal error: batch transfer changed total supply")
	}
	return l.CheckInvariants()
}

// BatchResult summarises a measured batching run, in exact integers.
// FeePerPayment is floored; FeeBasisPoints is the fee per payment in
// basis points of the payment amount (1 bp = 0.01 percent), which
// expresses the percentage without floating point.
type BatchResult struct {
	Payments      int
	PaymentAmount math.Int
	Transactions  int
	TotalFee      math.Int
	FeePerPayment math.Int
	// FeeBasisPoints is fee per payment / payment amount * 10000.
	FeeBasisPoints math.Int
}

// MeasureBatch runs a batching measurement on a ledger: it executes
// the given payments once as separate transactions (one fee each)
// and once as a single batch (one fee), using the ledger's own Send
// / DeductFee / BatchSend paths, and returns both measured results.
// The caller supplies two identically funded ledgers (one per run)
// so the two runs do not interfere; this function does not fund them.
func MeasureBatch(unbatched, batched *Ledger, from string, payments []Payment, fee math.Int) (BatchResult, BatchResult, error) {
	for _, p := range payments {
		if err := unbatched.Send(from, p.To, p.Amount); err != nil {
			return BatchResult{}, BatchResult{}, fmt.Errorf("unbatched payment to %s failed: %w", p.To, err)
		}
		if err := unbatched.DeductFee(from, fee); err != nil {
			return BatchResult{}, BatchResult{}, fmt.Errorf("unbatched fee failed: %w", err)
		}
	}
	if err := batched.BatchSend(from, payments, fee); err != nil {
		return BatchResult{}, BatchResult{}, fmt.Errorf("batched send failed: %w", err)
	}
	n := math.NewInt(int64(len(payments)))
	paymentAmount := payments[0].Amount
	unbatchedResult := BatchResult{
		Payments:       len(payments),
		PaymentAmount:  paymentAmount,
		Transactions:   len(payments),
		TotalFee:       fee.Mul(n),
		FeePerPayment:  fee,
		FeeBasisPoints: fee.MulRaw(10000).Quo(paymentAmount),
	}
	batchedResult := BatchResult{
		Payments:       len(payments),
		PaymentAmount:  paymentAmount,
		Transactions:   1,
		TotalFee:       fee,
		FeePerPayment:  fee.Quo(n),
		FeeBasisPoints: fee.MulRaw(10000).Quo(paymentAmount.Mul(n)),
	}
	return unbatchedResult, batchedResult, nil
}
