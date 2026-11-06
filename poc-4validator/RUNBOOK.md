# OGO Network 4-Validator PoC: Operator Runbook

Local staging only. All values TEST-ONLY. Chain: ogo-poc-4val-1.
Validators: val1, val2, val3, val4. RPC ports: 26757, 26767, 26777,
26787. P2P ports: 26756, 26766, 26776, 26786. Data:
`chain/localnet-data/<node>/`. Logs: `chain/localnet-data/<node>/node.log`.
Snapshots: `chain/snapshots/`. Scripts: `chain/scripts/localnet.sh`,
`run-poc.sh`. Toolchain: Go 1.25.7 at
`~/workspace/tools/go-1.25.7/bin/go`.

## Scope and gaps (read first)

- Each validator runs from a separate home directory with its own
  keys (`chain/localnet-data/val1` through `val4`, separate
  `priv_validator_key.json` and `node_key.json` per node). On this
  single-machine localnet the four nodes share one host, one
  operator, and one monitoring view (`localnet.sh status` and the
  per-node logs); that is NOT independent operation in the production
  sense. True independence (separate hosts, operators, monitoring)
  is a deployment property this localnet cannot demonstrate.
- No validator runs behind sentries here: all four nodes peer
  directly with each other on localhost. A sentry topology cannot be
  demonstrated on a single-machine localnet. That part of O1 is a
  gap, not a pass.
- The upgrade step depends on M2 (governance upgrade through
  Cosmovisor), which is NOT EXECUTED on this scaffold (see the PoC
  report). The upgrade procedure below is documented but not
  executed; do not mistake the written steps for a tested upgrade.

## 1. Start

Procedure:

```sh
cd ogo-network-poc-4validator
chain/scripts/localnet.sh build   # builds chain/ogod
chain/scripts/localnet.sh init    # fresh genesis, 4 validators, TEST-ONLY values
chain/scripts/localnet.sh start   # starts all 4 nodes, waits for height 5
chain/scripts/localnet.sh status  # height and app hash per node
chain/scripts/localnet.sh verify  # same ledger state on all 4 nodes
```

Or all at once: `chain/scripts/localnet.sh up` (build, init, start,
verify). `./run-poc.sh` runs the full narrated sequence.

Execution record: executed in every PoC run (runs 1, 2, 3, 4, 5; see
the PoC report for per-run numbers). All 4 nodes reach height 5 and
`verify` reports VERIFY OK at height 5 with identical app hash,
balances (alice 600,000,000, bob 400,000,000), supply 1,000,000,000,
and issuance 0 on all four nodes. `init` refuses to run while any
node is alive or any RPC port answers, so a stale node cannot survive
a re-init. Run 3 started normally and completed through S5, then was
cut short by an aborted exec call during the block-time sampling;
its start step had already executed successfully.

## 2. Stop

Procedure:

```sh
chain/scripts/localnet.sh stop
```

Stops all four managed nodes (SIGTERM, waits up to 15s for the
processes to exit, removes pid files) and reports each stop. A node
started by hand without a pid file is left alone.

Execution record: executed at the end of every completed PoC run by
the `run-poc.sh` exit trap (runs 1, 2, 4, 5). No node was left
running after any completed run (verified by `ps` and by RPC checks
returning no answer on 26757, 26767, 26777, 26787). Run 3 did not
reach its trap: the run was cut short by an aborted exec call, and
its nodes were stopped by that abort; afterwards `ps` showed no
ogod processes and val1 RPC answered nothing.

## 3. Upgrade

Procedure (documented, NOT executed):

1. Build the new binary and place it where Cosmovisor expects it
   (Cosmovisor is NOT installed in this PoC and no upgrade handler
   exists in the scaffold; see the M2 assessment in the PoC report).
2. Submit a governance upgrade proposal naming the upgrade height,
   vote, and pass it.
3. At the upgrade height each validator stops, Cosmovisor swaps in
   the new binary, the store migration adds the one new state field,
   and the chain resumes.
4. Verify balances and total supply are identical before and after
   on all four nodes, and record total downtime (criterion: under
   30 minutes).

Execution record: NOT EXECUTED. Reason: the scaffold has no
Cosmovisor binary, no upgrade handler, and no store migration to run
(x/upgrade is pinned and compiled only; the only upgrade code is the
monetary rejection rule in `x/ogocoin/bypass.go`, which rejects
monetary changes, it does not perform an upgrade). No fake Cosmovisor
and no pretend upgrade were implemented. The full M2 upgrade remains
open for Milestone C when the keepers are wired.

## 4. Snapshot

Procedure:

```sh
chain/scripts/localnet.sh snapshot val4
```

Stops the named node cleanly, archives its whole home directory to
`chain/snapshots/<node>-<timestamp>.tar.gz`, restarts the node, and
waits for it to catch up to val1's height.

Execution record (executed for real):

- PoC run 1 (2026-11-06): val4 snapshot
  `chain/snapshots/val4-20261106-163840.tar.gz`, size 772K, stop plus
  snapshot plus restart took 6s, val4 height 77 after restart.
- PoC run 2 (2026-11-06): val4 snapshot
  `chain/snapshots/val4-20261106-164220.tar.gz`, size 796K, stop plus
  snapshot plus restart took 5s, val4 height 78 after restart.
- PoC run 4 (2026-11-06): val4 snapshot
  `chain/snapshots/val4-20261106-164840.tar.gz`, size 748K, stop plus
  snapshot plus restart took 6s, val4 height 75 after restart.
- PoC run 5 (2026-11-06): val4 snapshot
  `chain/snapshots/val4-20261106-165224.tar.gz`, size 756K, stop plus
  snapshot plus restart took 3s, val4 height 74 after restart.

## 5. Restore

Procedure:

```sh
chain/scripts/localnet.sh restore val4 chain/snapshots/<snapshot-file>
```

Stops the named node cleanly, deletes its data directory, extracts
the snapshot in its place, restarts the node, and waits for it to
catch up to val1's height. Extraction uses `tar --no-same-owner`:
this environment forbids chown, so a plain extraction fails with
"Cannot change ownership" and aborts the restore under `set -e`
(caught in PoC run 1, fixed in `localnet.sh`).

Execution record (executed for real):

- PoC run 1 (2026-11-06): RESTORE FAILED. The snapshot step had
  succeeded, but extraction aborted on the tar ownership error
  (non-zero tar exit code under `set -e`), leaving val4 down and
  ending the run early. Fixed by adding `--no-same-owner` to the
  restore extraction. Recorded in the PoC report failure section.
- PoC run 2 (2026-11-06): RESTORE OK. val4 restored from
  `chain/snapshots/val4-20261106-164220.tar.gz`, caught up to height
  78 (val1: 79), restore plus catch-up took 5s. The final verify
  immediately after then failed on a measurement race (val4 had not
  yet committed height 79 when asked for its hash there); the verify
  step was fixed to wait for all nodes to reach the comparison
  height before comparing. See the PoC report failure section.
- PoC run 4 (2026-11-06): RESTORE OK. val4 restored from
  `chain/snapshots/val4-20261106-164840.tar.gz`, caught up to height
  79 (val1: 79), restore plus catch-up took 7s. Final verify after
  restore: VERIFY OK at height 79, all four nodes identical.
- PoC run 5 (2026-11-06): RESTORE OK. val4 restored from
  `chain/snapshots/val4-20261106-165224.tar.gz`, caught up to height
  76 (val1: 76), restore plus catch-up took 5s. Final verify after
  restore: VERIFY OK at height 76, all four nodes identical.

## 6. Monitoring (what to watch)

```sh
chain/scripts/localnet.sh status    # height and app hash per node
chain/scripts/localnet.sh voting    # voting power per validator (RPC and genesis)
chain/scripts/localnet.sh blocktimes 60  # block-time distribution over 60s
```

Height, app hash, balances, supply, and issuance are visible through
`status` and `verify`; there is no dashboard and no alerting in this
PoC (O2 is out of scope for this run). Block times observed: median
about 1.04s, p95 about 2.08s to 2.11s, 100 percent within 1 to 10
seconds over the measured 60s windows (see the PoC report; the 1s
sampling resolution quantizes intervals to about 1s or 2s).

## 7. Failure handling

- One validator down: with 4 validators the chain continues (S5
  liveness, demonstrated: height advanced with val4 down, val4
  restarted and caught up). Restart the stopped node with
  `localnet.sh start` (it restarts only stopped nodes) or by
  re-running the s5 action's restart path; then run `verify`.
- All validators down: restart with `localnet.sh start`; the chain
  resumes from persisted state (restart persistence is a unit test,
  `TestAppRestartRestoresHeightAndEpochs`, in the prototype suite).
- A node that will not catch up: snapshot and restore it from a
  known-good snapshot (section 5), then run `verify`.
- Never re-run `init` while nodes are alive: `init` refuses, and
  wiping the data directory under live nodes corrupts the run.
