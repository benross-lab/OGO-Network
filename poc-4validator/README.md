# OGO Network 4-Validator Proof of Concept

Local staging only. No GitHub remote, no push, no publishing or
deployment, no public testnet, no token issuance or distribution, no
spending, no changes to OGO production infrastructure, no changes to
`~/workspace/ogo/`. All monetary values are TEST-ONLY and labelled so
everywhere.

## What this is

A separate, isolated four-validator framework-selection proof of
concept for OGO Network, decided by Benross (2026-11-06, Option 1).
It sits beside the completed three-validator Phase 2 prototype in
`ogo-network-prototype/`, which stays unchanged and is a read-only
reference for this PoC. This PoC answers one question: does the chain
keep producing blocks when one of four validators goes offline?

## How it differs from the three-validator prototype

| | Three-validator prototype | This PoC |
|---|---|---|
| Validators | 3 (val1, val2, val3) | 4 (val1, val2, val3, val4) |
| Chain ID | ogo-local-1 | ogo-poc-4val-1 |
| RPC ports | 26657, 26667, 26677 | 26757, 26767, 26777, 26787 |
| P2P ports | 26656, 26666, 26676 | 26756, 26766, 26776, 26786 |
| Fault tolerance | None: losing one leaves 2/3 voting power, below CometBFT's greater-than-two-thirds quorum, so the chain halts (halt-and-resume only) | Tolerates one faulty validator: losing one leaves 3/4 voting power, so consensus continues (S5 liveness) |
| Extra actions | liveness (halt-and-resume) | s5 (fault tolerance), blocktimes, voting, snapshot, restore |
| Run script | demo.sh | run-poc.sh |

Everything else is the same so results are comparable: the same
chain code copied from the prototype (go.mod and go.sum byte-identical
to the prototype's, dependency matrix LOCKED), the same TEST-ONLY
parameters (max supply 1,000,010,000; alice 600,000,000; bob
400,000,000; 4-period schedule, initial 1,000, reduction 9/10; epoch
duration 20s), the same transfer (123,456 uogo, fee 100 uogo,
TEST-ONLY), the same epoch timer mirror (not real SDK x/epochs wiring).

## How to repeat the run

Requirements: Go 1.25.7 at `~/workspace/tools/go-1.25.7/bin/go`, `jq`,
`curl`, `python3`. Nothing else needs installing.

```sh
cd ogo-network-poc-4validator
./run-poc.sh
```

The script builds, initializes 4 validators, starts them, verifies
the same ledger state on all four nodes, waits for the first epoch
boundary, runs one transfer and verifies balances on all four nodes,
records voting power, runs the S5 fault-tolerance test, measures
block times over 60 seconds, executes a snapshot and restore of val4,
and stops the nodes cleanly on exit. All output goes to `logs/` with
timestamps in filenames (for example `logs/poc-run-20261106.log`).

Individual actions (all run from this directory):

```sh
chain/scripts/localnet.sh build
chain/scripts/localnet.sh init
chain/scripts/localnet.sh start
chain/scripts/localnet.sh verify
chain/scripts/localnet.sh voting
chain/scripts/localnet.sh s5
chain/scripts/localnet.sh blocktimes 60
chain/scripts/localnet.sh snapshot val4
chain/scripts/localnet.sh restore val4 <snapshot-file>
chain/scripts/localnet.sh status
chain/scripts/localnet.sh stop
```

Tests:

```sh
cd chain
~/workspace/tools/go-1.25.7/bin/go vet ./...
~/workspace/tools/go-1.25.7/bin/go test -count=1 ./...
```

Expected: vet exits 0; 90 tests pass, 0 fail (same as the prototype).

## What it proves and what it does not

See `ogo-network-poc-4validator-report.md` for the full report,
including the measured S5 result, the P1 partial result, and the M2
and O1 assessments. In short: this PoC demonstrates S5 liveness
(blocks continue with 1 of 4 validators offline) and measures block
times over the observed window. It does NOT demonstrate downtime
jailing, slashing, or unjailing (x/staking and x/slashing are pinned
and compiled only, not wired), it does NOT complete the 24-hour P1
run, and it does NOT execute a Cosmovisor governance upgrade (M2).
It is a framework-selection fault-tolerance test, not a mainnet
readiness claim. The epoch hook here is the same timer mirror as the
prototype, not real SDK x/epochs wiring.

## Files

- `run-poc.sh`: narrated end-to-end run (build, start, verify,
  transfer, voting, S5, blocktimes, snapshot/restore, stop).
- `chain/scripts/localnet.sh`: 4-validator localnet (build, init,
  start, stop, status, verify, voting, s5, blocktimes, snapshot,
  restore, clean).
- `chain/`: chain code copied from the prototype (go.mod/go.sum
  byte-identical; localnet-data/, the built ogod binary, and
  batchdemo-bin excluded and rebuilt fresh).
- `RUNBOOK.md`: operator runbook for start, stop, upgrade, snapshot,
  and restore, with execution records.
- `ogo-network-poc-4validator-report.md`: the PoC report.
- `logs/`: run logs, test output, vet output.
- `chain/snapshots/`: validator snapshots created by the run.
