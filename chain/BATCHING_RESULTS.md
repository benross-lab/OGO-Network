# Batching Results (Milestone 4)

Status: measured on the local prototype, 2026-11-06. Not a deploy,
not a fee policy decision. All parameters are TEST-ONLY.

## How the numbers were produced

Program: `chain/batchdemo` (package main, `go run ./batchdemo`
from the `chain/` directory). It uses the prototype's own code
paths: `Ledger.Send` plus `Ledger.DeductFee` for the unbatched run,
and `Ledger.BatchSend` (new, `x/ogocoin/batch.go`) for the batched
run. Both runs use identically funded ledgers. The numbers below
are the program's actual output, copied verbatim, not the spec
table.

## Measured results

100 agent payments of 0.001 OGO (1,000 uogo) each, fee 100 uogo
per transaction (TEST-ONLY, the Milestone B localnet fee).

| Case | Transactions | Total fee | Fee per payment | Fee as share of payment |
|---|---|---|---|---|
| Unbatched (100 transactions, one fee each) | 100 | 10,000 uogo | 100 uogo | 10.00 percent (1,000 basis points) |
| Batched (1 transaction, 100 payments, one fee) | 1 | 100 uogo | 1 uogo | 0.10 percent (10 basis points) |

Fee saving from batching: 9,900 uogo, 99 percent of the unbatched fee.

## Comparison with the spec table

Fee market spec section 5, at a 100 uogo fee and a 0.001 OGO
payment, predicts 10.00 percent unbatched and 0.10 percent
batched for 100 payments per transaction. The prototype
reproduces both numbers exactly. The spec's 1,000 uogo fee case
(a 0.001 OGO payment costing 100 percent in fees) was not
measured here, because the prototype's fee is the flat 100 uogo
localnet fee, not a configurable gas-priced fee; measuring a
1,000 uogo fee would only have meant passing a different flat
number in, which proves nothing about the fee market.

## What the fee handling actually is (honest limitation)

The scaffold's fee handling is thin, and this file says so
explicitly, as the task requires. There is no gas metering, no
gas price, no minimum gas price, and no congestion adjustment
in this prototype. The fee for a transaction is a flat amount
supplied by the sender: the CLI default is 0, the localnet proof
used 100 uogo. `DeductFee` moves that flat fee from payer to the
fee collector; it never mints. `FeeForTransaction` returns the
flat fee regardless of how many payments are inside the
transaction, which is the whole mechanism batching exploits.

What batching therefore proves here is the arithmetic that
matters: one fee per transaction instead of one fee per payment.
It does NOT prove anything about a gas-priced fee market, a
dynamic fee, or spam resistance at a given gas price. Those are
spec sections 3.2 and 3.3 and are not implemented in this
prototype. A gas-priced fee market, tiered agent fees, and a
separate agent transaction type remain unimplemented.

## Batch semantics decided here

`BatchSend` is atomic: if any payment in the batch would fail
(zero amount, empty recipient, insufficient sender balance for
payments plus the single fee), the whole batch fails and nothing
changes. There is no skip-a-payment mode in this prototype. The
spec section 7 open question (whole batch fails vs failing
payment skipped) is answered here as whole-batch-fails, to be
revisited in developer documentation before batching is
recommended to agents.

## Reproduction

From `chain/`, with the locked toolchain:

```
TMPDIR=~/workspace/.go-tmp GOCACHE=~/workspace/.go-build-cache \
  ~/workspace/tools/go-1.25.7/bin/go run ./batchdemo
```

Expected output (verbatim, 2026-11-06 run):

```
OGO Network batching demonstration (Milestone 4)
Measured from the prototype ledger, not from the spec table.
All parameters TEST-ONLY. Fee model: flat fee per transaction,
no gas metering in this scaffold (see BATCHING_RESULTS.md).
Payments: 100 x 1000 uogo (0.001 OGO each), fee 100 uogo per transaction

Unbatched (100 transactions, one fee each)
  transactions: 100
  total fee: 10000 uogo
  fee per payment: 100 uogo
  fee as share of payment: 10.00 percent (1000 basis points)
Batched (1 transaction, 100 payments, one fee)
  transactions: 1
  total fee: 100 uogo
  fee per payment: 1 uogo
  fee as share of payment: 0.10 percent (10 basis points)

Fee saving from batching: 9900 uogo (99% of the unbatched fee)
```

Tests covering the same measurement: `TestMeasureBatchHundredPayments`,
`TestBatchSendPreservesSupplyAndChargesOneFee`, and
`TestBatchSendAtomicOnFailure` in `x/ogocoin/batch_test.go`.
