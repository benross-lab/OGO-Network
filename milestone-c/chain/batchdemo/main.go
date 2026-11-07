// Command batchdemo measures agent transaction economics on the
// prototype's own code paths (Milestone 4).
//
// It executes 100 agent payments of 0.001 OGO (1,000 uogo) each,
// once unbatched (100 separate transactions, one fee each) and once
// batched (one transaction carrying 100 payments, one fee), at the
// prototype's fee level, using the prototype ledger's Send,
// DeductFee, and BatchSend paths, and prints the measured fee cost
// per payment and fee as a percentage of payment for both cases.
//
// The numbers are measured from the code, not copied from the spec
// table. The spec table (fee market spec section 5) says the same
// thing at a 100 uogo fee; this program proves the prototype
// actually produces those numbers.
//
// Fee honesty: the prototype has no gas metering, no gas price, and
// no congestion adjustment. The fee is a flat 100 uogo per
// transaction (TEST-ONLY, the Milestone B localnet fee). What is
// measured here is the flat-fee-per-transaction model as implemented.
// A gas-priced fee market is NOT implemented and is not claimed.
// Run: go run ./batchdemo (from the chain directory).
package main

import (
	"fmt"
	"os"

	"cosmossdk.io/math"

	"github.com/benross-lab/ogo-network/x/monetary"
)

func main() {
	const paymentsCount = 100
	paymentAmount := math.NewInt(1000) // 0.001 OGO, in uogo
	fee := monetary.FlatFeePerTx        // 100 uogo per transaction, TEST-ONLY

	payments := make([]monetary.Payment, paymentsCount)
	for i := range payments {
		payments[i] = monetary.Payment{To: "ogo1recipient", Amount: paymentAmount}
	}

	unbatched, batched, err := monetary.MeasureBatch(fundingLedger(), fundingLedger(), "ogo1agent", payments, fee)
	if err != nil {
		fmt.Fprintf(os.Stderr, "batchdemo failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OGO Network batching demonstration (Milestone 4)")
	fmt.Println("Measured from the prototype ledger, not from the spec table.")
	fmt.Println("All parameters TEST-ONLY. Fee model: flat fee per transaction,")
	fmt.Println("no gas metering in this scaffold (see BATCHING_RESULTS.md).")
	fmt.Printf("Payments: %d x %s uogo (0.001 OGO each), fee %s uogo per transaction\n",
		paymentsCount, paymentAmount, fee)
	fmt.Println()
	printResult("Unbatched (100 transactions, one fee each)", unbatched)
	printResult("Batched (1 transaction, 100 payments, one fee)", batched)
	fmt.Println()
	fmt.Printf("Fee saving from batching: %s uogo (%s of the unbatched fee)\n",
		unbatched.TotalFee.Sub(batched.TotalFee), percentSaved(unbatched, batched))
}

func printResult(label string, r monetary.BatchResult) {
	// Basis points to a percentage with two decimals, in integers:
	// 1000 bp = 10.00 percent, 10 bp = 0.10 percent.
	bp := r.FeeBasisPoints.Int64()
	fmt.Printf("%s\n", label)
	fmt.Printf("  transactions: %d\n", r.Transactions)
	fmt.Printf("  total fee: %s uogo\n", r.TotalFee)
	fmt.Printf("  fee per payment: %s uogo\n", r.FeePerPayment)
	fmt.Printf("  fee as share of payment: %d.%02d percent (%d basis points)\n", bp/100, bp%100, bp)
}

func percentSaved(unbatched, batched monetary.BatchResult) string {
	saved := unbatched.TotalFee.Sub(batched.TotalFee)
	// saved / unbatched * 100, floored, in integers.
	pct := saved.MulRaw(100).Quo(unbatched.TotalFee)
	return pct.String() + "%"
}

func fundingLedger() *monetary.Ledger {
	// A funding ledger is built through the same genesis validation
	// the chain uses. The max is deliberately far above genesis +
	// schedule total, so the legacy loose ValidateGenesis passes;
	// this is a measurement harness, not a production genesis.
	// TEST-ONLY values only.
	max := math.NewInt(10000000)
	params := monetary.DefaultParams()
	params.MaxSupply = &max
	params.IssuanceSchedule = &monetary.IssuanceSchedule{
		ScheduleType:         monetary.ScheduleTypeReducingPerPeriod,
		StartHeight:          1,
		PeriodLength:         100,
		Periods:              10,
		InitialPerPeriod:     math.NewInt(1000),
		ReductionNumerator:   9,
		ReductionDenominator: 10,
	}
	gs := monetary.GenesisState{
		Params: params,
		Balances: []monetary.Balance{
			{Address: "ogo1agent", Denom: monetary.Denom, Amount: math.NewInt(1000000), Purpose: "public"},
		},
	}
	l, err := monetary.NewLedger(gs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "batchdemo ledger setup failed: %v\n", err)
		os.Exit(1)
	}
	return l
}
