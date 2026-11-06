# OGO Network 4-Validator Proof of Concept: Report

Status: Local staging only. No repository created on GitHub, no
blockchain deployed, no token issued, no testnet launched, no spending
authorized, no production change authorized, no change to
`~/workspace/ogo/`.
Date: 2026-11-06
PoC directory: `ogo-network-poc-4validator/`
Built on: the completed three-validator Phase 2 prototype
(`ogo-network-prototype/`, local git commit b29f830, no remote),
which stays unchanged and is a read-only reference for this PoC.
Toolchain: Go 1.25.7 at `~/workspace/tools/go-1.25.7/bin/go`.
Dependency matrix LOCKED as one tested unit (Cosmos SDK v0.53.6
required with Gaia's replace to v0.53.4 preserved, CometBFT v0.38.23,
wasmd v0.60.9, wasmvm/v2 v2.3.5, ibc-go/v10 v10.7.0). `go.mod` and
`go.sum` are byte-identical to the prototype's, verified with `cmp`.
All values in this report are TEST-ONLY and labelled so.

This report uses Benross's four exact categories for every major
feature. Implemented means code exists. Executed successfully means
functionality actually ran successfully. Tested under failure means
failure behaviour was deliberately tested. Not implemented means
functionality remains incomplete. Passing compilation is never
reported as a working feature.

## 1. Summary

The four-validator proof of concept is built and verified for its
one purpose: fault tolerance. With four equally weighted validators,
losing one leaves 3/4 voting power, above CometBFT's
greater-than-two-thirds quorum, so the chain keeps producing blocks
with the remaining three validators. That was demonstrated live in
four runs, with measured heights, and the stopped validator restarted,
caught up, and rejoined consensus with all four nodes agreeing. The
same ledger state, transfer, and epoch issuance behaviour as the
three-validator prototype were reproduced on four nodes.

S5 is PARTIALLY EXECUTED: the liveness half is demonstrated with
measured evidence; the jailing, slashing, and unjail half is NOT
EXECUTED, because the scaffold has no downtime jailing or slashing
machinery to observe (see section 4). P1 is PARTIALLY EXECUTED: block
times were measured over the observed windows, but a 24-hour run
cannot be completed in one session (see section 5). M2 is NOT
EXECUTED: no Cosmovisor upgrade can run on this scaffold (see section
6). O1 is PARTIALLY EXECUTED: start, stop, snapshot, and restore
were executed for real, but no sentry topology was demonstrated and
the upgrade step depends on M2 (see section 7).

90 top-level tests pass, 0 fail. `go vet ./...` exits 0. The run
script `run-poc.sh` ran end to end cleanly twice (runs 4 and 5), with
identical outcomes in kind (same balances, same voting power, same
block-time profile; heights and transaction hashes differ run to run
because each run starts from a fresh genesis with fresh keys).

## 2. Isolation from the prototype

- **Implemented:** the PoC lives in its own directory
  `ogo-network-poc-4validator/`. Only the prototype's `chain/`
  directory was copied, excluding `localnet-data/`, the built `ogod`
  binary, and `batchdemo-bin`, which were rebuilt fresh. Nothing
  else was copied from the prototype directory.
- **Executed successfully:** the PoC uses chain-id `ogo-poc-4val-1`
  (prototype: `ogo-local-1`), RPC ports 26757, 26767, 26777, 26787
  (prototype: 26657, 26667, 26677), P2P ports 26756, 26766, 26776,
  26786 (prototype: 26656, 26666, 26676), and its own data directory
  `chain/localnet-data/` inside the PoC directory. The prototype's
  running nodes and data directory were never reused.
- **Tested under failure:** not applicable; isolation is a
  configuration property, verified by inspection and by the port and
  chain-id values in every run log.
- **Not implemented:** nothing missing. The prototype directory's
  `git status --porcelain` was empty before the work started and is
  empty after it finished (verified in the final self-check).

## 3. Feature report (four exact categories)

### 3.1 Four-validator consensus on the same ledger

- **Implemented:** `chain/scripts/localnet.sh` builds one shared
  genesis (4 validators, power 10 each) and starts 4 nodes on
  distinct RPC/P2P ports with persistent peers. `run-poc.sh` runs
  the narrated end-to-end sequence. Every 3-node loop and message
  from the prototype script was generalized to 4 nodes; `verify`
  compares all four nodes.
- **Executed successfully:** live, in runs 4 and 5 (and in runs 1, 2,
  and 3 for the steps they reached). Run 4: VERIFY OK at heights 5,
  16, 31, and 79. Run 5: VERIFY OK at heights 5, 17, 31, and 76.
  Identical app hash on all 4 nodes at the same height (run 4 final:
  `B67EBC765DB8AF43BE9328C486C750C74E0507C9A09109136D6B95A6F78E4F2F`
  at height 79; run 4 final status: all four nodes at height 80 with
  that same app hash). Identical balances, identical supply,
  identical issuance on all four nodes at every verify.
- **Tested under failure:** S5 (section 4) and the snapshot/restore
  test (section 7). A lagging node after restore is handled by the
  verify step waiting for all nodes to reach the comparison height
  before comparing (fix recorded in section 9, item 2).
- **Not implemented:** staking, delegation, unbonding, slashing
  (keepers pinned and compiled only). No change from the prototype.

### 3.2 Epoch execution (the hook demonstrably fires)

- **Implemented:** the same epoch timer mirror as the prototype
  (`x/ogocoin/epochs.go`, `AfterEpochEnd` hook, app integration in
  `app/app.go`). No code in this area was changed for the PoC.
- **Executed successfully:** live, in every run. Run 4: after the
  first boundary, epoch state on val1 showed `current_epoch: 2`,
  `hook_fire_count: 1`, `total_issued: 1000`,
  `issued_by_period: 0:1000`, rewards pool balance 1000, supply
  1,000,001,000. By the S5 verify, 3 epochs had fired: issued 2710,
  supply 1,000,002,710 (1000 plus 900 plus 810).
- **Tested under failure:** the prototype's halt catch-up and
  duplicate-claim tests still pass in the suite (90 tests, 0 fail);
  no new epoch failure test was added in this PoC.
- **Not implemented:** real SDK x/epochs wiring. The mirror is a
  stand-in, exactly as in the prototype. The mirror hook-fires test
  does not prove depinject registration of the real module; that
  proof is Milestone C work, release-blocking, and is not claimed
  here.

### 3.3 Native coin transfers and reconciliation

- **Implemented / Executed successfully:** `tx send`, `query
  balance/supply/params/schedule/epochs/global-supply`, as in the
  prototype. Run 4: transfer of 123,456 uogo from alice to bob, fee
  100 uogo, transaction hash
  `E0A821BDD9B63EEF1BF57EC44E9761D6A6159A435C747A1E77A2CF13B5A0B338`,
  code 0, broadcast took 98 ms. Run 5: transaction hash
  `F2AAD930EC8F978AE7F6D13A7D4909EF65C1AA59D37E8685DD4C7E66EA6FF470`,
  code 0, broadcast took 80 ms. Both runs: alice 599,876,444
  (600,000,000 minus 123,456 minus 100 fee), bob 400,123,456, on all
  4 nodes. (Transaction hashes and wallet addresses differ between
  runs because each run starts from a fresh genesis with fresh keys;
  the balances are identical.)
- **Tested under failure:** insufficient-balance and bad-amount
  rejections (prototype ledger tests, still passing in the suite).
- **Not implemented:** signature verification (JSON transactions,
  not signature-checked; real x/auth arrives with Milestone C
  depinject wiring). Gas metering (flat fee; CheckTx reports
  hardcoded GasWanted 1,000). Keplr or Leap wallet integration. Live
  IBC relaying (escrow accounting is ledger-level only).

### 3.4 Configurable economics (geometric and hybrid)

- **Implemented:** unchanged from the prototype: the monetary
  interface supports geometric decay, hybrid, and hybrid with the
  gentle terminal taper, selected through genesis configuration.
- **Executed successfully:** geometric only, live through the epoch
  path (section 3.2). The PoC run uses the same TEST-ONLY parameters
  as the prototype so results are comparable: max supply
  1,000,010,000; alice 600,000,000; bob 400,000,000; 4-period
  schedule, initial 1,000, reduction 9/10; epoch duration 20s.
- **Not implemented:** hybrid through the app's epoch path, and no
  monetary parameters were selected in this PoC (no max supply,
  genesis supply, reduction factor, floor, period length, split, fee
  level, treasury spend level, or reserve floor was chosen, and no
  model family was selected). The prototype's documented deviations
  carry over unchanged (strict M equals G plus S_actual is not the
  default genesis path; exact versus stepwise floor diverge from
  period 7 onward; taper numbers are the measured ones).

### 3.5 S5 fault tolerance (liveness half)

- **Implemented:** the `s5` action in `chain/scripts/localnet.sh`
  and step 7 of `run-poc.sh`. It records voting power and height
  before the outage, stops val4 cleanly, samples val1 height at 0s,
  10s, 20s, 30s after the stop, confirms val4's RPC is down during
  the outage, restarts val4, waits for it to catch up, records
  voting power after the outage, and verifies all four nodes agree.
- **Executed successfully:** live, in runs 1, 2, 3, 4, and 5. The
  measured result is in section 4. Blocks continued with 1 of 4
  validators offline in every run; val4 restarted, caught up, and
  all four nodes agreed afterwards in every run.
- **Tested under failure:** this is itself the failure test. The
  deliberate failure is one validator offline; the system behaved as
  specified for the liveness half (continued block production).
- **Not implemented:** the jailing, slashing, and unjail half of S5
  as written (section 4). Restarting the validator and verifying it
  syncs and rejoins consensus is the executable part; unjailing as a
  staking operation is not.

### 3.6 Block-time measurement

- **Implemented:** the `blocktimes` action in
  `chain/scripts/localnet.sh` and step 8 of `run-poc.sh`. It samples
  val1 height once per second over a window (default 60s), records
  the first-seen time of each height, and reports count, min,
  median, p95, max, and the percentage of intervals within 1 to 10
  seconds.
- **Executed successfully:** live, in runs 1, 2, 4, and 5 (run 3 was
  cut short during this step). The measured distributions are in
  section 5. Across the four measured windows, 159 intervals were
  measured and 100 percent fell within 1 to 10 seconds.
- **Tested under failure:** not applicable; this is a measurement,
  not a failure test.
- **Not implemented:** a 24-hour continuous observation (section 5),
  a dashboard, and alerting (O2, out of scope for this run).

### 3.7 Snapshot and restore (O1 / O5 behaviour)

- **Implemented:** the `snapshot` and `restore` actions in
  `chain/scripts/localnet.sh` and step 9 of `run-poc.sh`.
- **Executed successfully:** live, in runs 2, 4, and 5 (run 1's
  restore failed; see section 9, item 1). Run 4: snapshot 748K, 6s,
  val4 height 75; restore OK, caught up to height 79 (val1: 79),
  7s; final verify OK at height 79. Run 5: snapshot 756K, 3s, val4
  height 74; restore OK, caught up to height 76 (val1: 76), 5s;
  final verify OK at height 76.
- **Tested under failure:** the restore itself is the failure
  recovery test: a validator's data directory is deleted and rebuilt
  from a snapshot, and the node rejoins and catches up.
- **Not implemented:** restore from a snapshot taken on another
  machine, and restore within a 2-hour bound measured at scale (O5
  as written wants catch-up within 2 hours; the local restore took
  seconds, which satisfies the bound trivially at this scale but
  does not prove it at production scale). O5 is not claimed as
  passed by this PoC; the restore behaviour is recorded as executed
  evidence for it.

### 3.8 Automated monetary tests

- **Implemented:** the prototype's 90 top-level tests, unchanged
  (54 Milestone B, 13 Milestone 2 epoch tests, 23 Milestone 3, 4,
  and 5 tests). No test was added, removed, or modified in this PoC.
- **Executed successfully:** full suite run by this PoC with output
  saved in `logs/test-output-20261106.log` (see section 8).
- **Tested under failure:** duplicate issuance prevention, halt
  recovery, chain halt and recovery, treasury low-inflow binding,
  insufficient-balance and bad-amount rejections, restart
  persistence (all in the suite, all passing).
- **Not implemented:** the Milestone C test requirements (real
  x/epochs hook-fires test, S2 bypass re-run on the wired chain, S1
  signature verification, 30-minute S4 halt catch-up, hybrid through
  the epoch path, strict M equals G plus S_actual as the default,
  the single rounding rule, P2/P3/P4 throughput and latency, S5
  slashing and liveness as written, S7 wasm sandbox, clean-copy
  reproducible build for the prototype).

## 4. S5 measured result: PARTIALLY EXECUTED

Criterion (acceptance matrix section 3.2, S5): take 1 of 4 validators
offline for longer than the downtime window, observe downtime jailing
and the documented slashing, chain continues with the remaining 3
validators, then unjail and rejoin.

### 4.1 Voting power (equal weight, exact numbers)

Genesis validator set, run 4 (each run generates fresh keys, so
addresses differ per run; powers are identical in every run):

- val1: power 10, address B18A5C2A1B64A21DAA49EAF82961EA20A475C075
- val2: power 10, address 608B9430CF448B2BF68BD8EDA1C1501810A470D0
- val3: power 10, address 78F84A0726CA70DD8C454AA7B77717753E0E8C0B
- val4: power 10, address C1750740618336DFFA7AB88F90F06E574F24AE6F

Genesis total voting power: 40. The CometBFT `/validators` RPC
endpoint on each of the four nodes reported the same four validators
at power 10 each, total 40, before the outage and again after the
outage, in every run. Run 5 genesis set, for cross-check:

- val1: power 10, address 7AAEB0021F25C37815E600629835198068BED23A
- val2: power 10, address D306C72B871E2EF9EB2F276512621575299FE2B2
- val3: power 10, address C47D833B5F870F3A926EB7AFD72C29060A091EA3
- val4: power 10, address FDF088167CC4212C979E312432979556B1A432C4

With one validator down, the remaining voting power is 30 of 40
(3/4), above CometBFT's greater-than-two-thirds quorum (26.67 of 40).
With three validators, the same outage would leave 20 of 30 (2/3),
not a quorum, and the chain halts. That is the technical distinction
this PoC exists to demonstrate.

### 4.2 Liveness: blocks continued with val4 down: YES

| Run | Height before outage (val1) | t+0s | t+10s | t+20s | t+30s | Advanced | val4 RPC during outage |
|---|---|---|---|---|---|---|---|
| 1 | 19 | 20 | 24 | 27 | 30 | 19 to 30 | DOWN (confirmed) |
| 2 | 18 | 18 | 22 | 26 | 30 | 18 to 30 | DOWN (confirmed) |
| 3 | 18 | 19 | 23 | 26 | 30 | 18 to 30 | DOWN (confirmed) |
| 4 | 17 | 18 | 22 | 25 | 29 | 17 to 29 | DOWN (confirmed) |
| 5 | 18 | 18 | 22 | 26 | 30 | 18 to 30 | DOWN (confirmed) |

In every run the height advanced at every sample while val4 was
down. The chain did not halt. This is the opposite of the
three-validator prototype, where stopping one validator held the
height fixed for 15 seconds.

### 4.3 Restart and sync verification: YES

| Run | val4 caught up to | val1 at that moment | Verify after outage | Balances / supply / issued at that verify |
|---|---|---|---|---|
| 1 | 33 | 33 | VERIFY OK at height 33 | alice 599,876,444; bob 400,123,456; supply 1,000,002,710; issued 2,710 |
| 2 | 31 | 31 | VERIFY OK at height 31 | same values as run 1 |
| 3 | 33 | 33 | VERIFY OK at height 33 | same values as run 1 |
| 4 | 31 | 31 | VERIFY OK at height 31 | same values as run 1 |
| 5 | 31 | 31 | VERIFY OK at height 31 | same values as run 1 |

After the outage, voting power was re-recorded on all four nodes:
power 10 each, total 40, identical to before the outage. The
post-outage verify compared the app hash at a common height on all
four nodes and the live balances, supply, and issuance: all identical.
No double mint occurred across the outage (issued 2,710 is exactly
1000 plus 900 plus 810, three epochs, no lump).

### 4.4 Jailing, slashing, unjail: NOT EXECUTED

No downtime jailing was observed. No slashing was observed. No
unjailing as a staking operation was performed. Reason: the scaffold
has no downtime jailing or slashing machinery to observe. The
standard keepers (x/staking, x/slashing) are pinned and compiled
only, not wired (blocker B6 in
`OGO_NETWORK_PROTOTYPE_BLOCKERS.md`). Evidence checked in this PoC:

- `chain/app/modules.go` lists x/staking as "pinned-compiled",
  keeper wiring pending, and does not list x/slashing at all.
- The string "slashing" appears nowhere in `chain/app/` or
  `chain/cmd/` (grep, 2026-11-06).
- No `SetUpgradeHandler`, `UpgradeHandler`, `StoreUpgrades`, or
  upgrade command exists anywhere in the chain code.
- The only upgrade-related code is the ledger-level monetary
  rejection rule in `x/ogocoin/bypass.go` (`ApplyUpgrade` rejects
  changes to max supply, schedule, or mint paths), which is a rule
  about what an upgrade may not do, not an upgrade mechanism.

Restarting val4 and verifying it syncs and rejoins consensus is the
executable part, and it was executed. Unjailing as a staking
operation is not executable here. The full S5 as written remains open
for Milestone C, when the keepers are wired. Nothing in this report
claims jailing or slashing was observed.

## 5. P1: PARTIALLY EXECUTED

Criterion (acceptance matrix section 3.1, P1): 4 validators produce
blocks continuously for 24 hours, no halt longer than 5 minutes
except a planned halt test, block time within 1 to 10 seconds for at
least 95 percent of blocks.

A 24-hour run cannot be completed in this session. P1 is therefore
PARTIALLY EXECUTED, not passed. What was executed: the network ran
continuously for the whole duration of each PoC run (run 4: 149
seconds wall clock, 16:46:27Z to 16:48:56Z; run 5: 144 seconds,
16:50:11Z to 16:52:35Z), and the block-time distribution was measured
over a 60-second window inside each run with the `blocktimes` action.

Measured block-time distributions (60s window per run, val1, one
sample per second):

| Run | Intervals | Min | Median | p95 | Max | Within 1 to 10s |
|---|---|---|---|---|---|---|
| 1 | 39 | 1.024s | 1.040s | 2.080s | 2.139s | 39 of 39 (100.0 percent) |
| 2 | 39 | 1.023s | 1.049s | 2.108s | 2.175s | 39 of 39 (100.0 percent) |
| 4 | 41 | 1.015s | 1.019s | 2.048s | 2.056s | 41 of 41 (100.0 percent) |
| 5 | 40 | 1.015s | 1.023s | 2.067s | 2.096s | 40 of 40 (100.0 percent) |

Total across the four measured windows: 159 intervals, 159 within 1
to 10 seconds (100.0 percent), which exceeds the 95 percent bar for
the observed windows only. No halt longer than 5 minutes occurred in
any run; the longest planned halt was the 30-second S5 outage of one
validator, during which the other three kept producing blocks.

Measurement caveat, stated honestly: the 1-second sampling
resolution quantizes the intervals. True block time is about 1
second, so consecutive first-seen times land about 1 second or about
2 seconds apart, and the intervals alternate between those values.
The p95 and max near 2.1s are artifacts of that quantization, not
evidence of 2-second blocks. The distribution still establishes that
blocks arrive steadily at about 1 per second with no gaps above 10
seconds in the observed windows. It does not establish 24-hour
behaviour.

Path to complete P1: run the same 4-validator network continuously
for 24 hours and sample block times over the full period. The run
script as written stops the network after about 2.5 minutes, so
completing P1 needs a long-run mode (start the network with
`localnet.sh start` and sample with `blocktimes` repeatedly, or run
`run-poc.sh` with its stop step deferred) plus the same distribution
report over the 24-hour window. That run was not attempted here.

## 6. M2 (maintenance): NOT EXECUTED

Criterion (acceptance matrix section 3.3, M2): governance upgrade
through Cosmovisor on all 4 validators, with a store migration adding
one new state field: upgrade height reached, binaries swapped,
migration applied, chain resumes, balances and total supply identical
before and after, total downtime under 30 minutes.

M2 was NOT EXECUTED. Reason: it cannot be executed as specified on
this scaffold. Checked in this PoC (2026-11-06):

- No Cosmovisor binary is installed (`which cosmovisor` finds
  nothing). `go.mod` mentions Cosmovisor v1.7.3 only as a separate
  binary, not a Go module dependency.
- The `ogod` binary has no upgrade command (commands: version, init,
  show-node-id, show-validator, set-genesis-validators, keys,
  add-genesis-account, validate-genesis, start, tx, query).
- No upgrade handler exists in the code: no `SetUpgradeHandler`,
  `UpgradeHandler`, `StoreUpgrades`, or `RegisterUpgrade` anywhere
  in `chain/` Go code.
- x/upgrade is not even among the blank imports in
  `chain/app/modules.go` (only auth, bank, distribution, gov,
  staking, wasm, and ibc-go are pinned and compiled). x/gov is
  pinned and compiled only, so there is no on-chain governance
  proposal path to submit the upgrade through either.

No fake Cosmovisor and no pretend upgrade were implemented. The
genuine partial path that does exist, stated exactly: the ledger-level
upgrade rule in `x/ogocoin/bypass.go` (`ApplyUpgrade`) rejects
monetary changes through an upgrade proposal, and that rule is unit
tested (including `TestGovNonMonetaryUpgradeAllowed`, which allows a
non-monetary upgrade). That proves a ledger rule about what an
upgrade may not do. It does not prove a Cosmovisor upgrade can run,
a store migration can apply, or a chain can resume after one, and it
is not claimed as M2 progress. The full M2 remains open for
Milestone C, when the keepers are wired. M2 blocks framework
selection until it is executed; nothing in this PoC removes that
block.

## 7. O1 (operating requirements): PARTIALLY EXECUTED

Criterion (acceptance matrix section 3.4, O1): 4 validators operated
as if independent (separate homes, keys, monitoring), at least one
validator behind sentries, written runbook for start, stop, upgrade,
snapshot, and restore, each step executed at least once.

O1 is PARTIALLY EXECUTED. The runbook exists (`RUNBOOK.md` in the
PoC directory) and the executable steps were executed for real:

- Start: executed in every PoC run (runs 1, 2, 3, 4, 5). All 4 nodes
  reach height 5; verify OK at height 5.
- Stop: executed at the end of runs 1, 2, 4, and 5 by the
  `run-poc.sh` exit trap; no node left running afterwards (verified
  by `ps` and RPC checks). Run 3's nodes were stopped by the aborted
  exec call that ended that run (see section 9, item 3).
- Snapshot: executed for real (run 1: 772K, 6s, val4 height 77;
  run 2: 796K, 5s, val4 height 78; run 4: 748K, 6s, val4 height 75;
  run 5: 756K, 3s, val4 height 74).
- Restore: executed for real (run 2: OK, 5s, caught up to 78;
  run 4: OK, 7s, caught up to 79; run 5: OK, 5s, caught up to 76;
  run 1 failed and is recorded in section 9, item 1).
- Upgrade: NOT executed (depends on M2, section 6).

Gaps, stated explicitly:

- Sentry topology: NOT demonstrated. All four validators peer
  directly with each other on localhost. A sentry topology cannot be
  demonstrated on a single-machine localnet.
- Independence: the four validators have separate homes, separate
  keys, and per-node logs, but they share one host, one operator,
  and one monitoring view. That is not independent operation in the
  production sense; true independence is a deployment property this
  localnet cannot demonstrate.
- Upgrade step: not executed, because M2 is not executable here.

## 8. Test summary

Commands run in `chain/` with Go 1.25.7:

- `go vet ./...`: exit 0. Output in `logs/vet-20261106.log`.
- `go test -count=1 ./...`: exit 0. Full output in
  `logs/test-output-20261106.log`:
  - `ok github.com/benross-lab/ogo-network/app`
  - `ok github.com/benross-lab/ogo-network/x/ogocoin`
  - `batchdemo` and `cmd/ogod`: no test files.
- Verbose count: 90 top-level tests pass, 0 fail. Same as the
  prototype (90 pass, 0 fail).
- `go.mod` and `go.sum` byte-identical to the prototype's
  (verified with `cmp` before and after the test runs).

## 9. Failure report

Actual failures encountered during this PoC, not just successes.
Nothing in this section was smoothed over.

1. **Restore failed in PoC run 1 (tar ownership).** The snapshot
   step succeeded (val4 snapshot 772K, 6s), but the restore step
   aborted: extracting the snapshot failed with `tar: ... Cannot
   change ownership to uid 0, gid 65534: Operation not permitted`
   for every file, and tar exited non-zero, which under `set -e`
   aborted the restore before the node restarted, leaving val4 down
   and ending the run early (run 1 wall clock 162s, steps 1 to 8
   passed, step 9 failed). Cause: this environment forbids chown
   (the prototype copy showed the same: `cp -a` could not preserve
   ownership). The files had extracted despite the errors; only the
   exit code was wrong. Fixed by adding `--no-same-owner` to the
   restore extraction in `localnet.sh`, with a comment recording the
   cause. Restores succeeded in runs 2, 4, and 5 after the fix.
2. **Final verify failed in PoC run 2 (measurement race).** The
   restore had caught val4 up to height 78 while val1 stood at 79.
   The verify step immediately after picked height 79 from val1 and
   asked val4 for its app hash there; val4 had not yet committed
   block 79, so it returned an empty hash, which the script read as
   a mismatch (`VERIFY FAILED: nodes disagree on app hash at height
   79`). The nodes did not disagree; one was one block behind and
   the comparison asked for a block it did not have yet. Fixed by
   making `verify` wait up to 30s for every node to reach the
   comparison height before comparing app hashes. Verifies passed
   in runs 4 and 5 after the fix, including the post-restore verify.
3. **PoC run 3 cut short by an aborted exec call.** Run 3 had
   completed through S5 (S5 liveness OK, 18 to 30, val4 caught up to
   33, verify OK at height 33) and was in the block-time sampling
   step when the foreground exec session returned "aborted" (the
   delivery of an earlier background run's completion interrupted
   it). The run-poc.sh process was killed, its nodes were stopped by
   that abort, and no cleanup line was written to the log. No data
   was lost (each run starts from a fresh genesis), and run 3 does
   not count as a full end-to-end run. The two full clean runs
   counted for this report are runs 4 and 5.
4. **Ownership warnings on the initial copy.** `cp -a` of the
   prototype's `chain/` reported "failed to preserve ownership:
   Operation not permitted" for five directories (same chown
   restriction as item 1). Cosmetic only: the copy completed, and
   `go.mod` and `go.sum` were verified byte-identical with `cmp`.
5. **Block-time sampling quantization.** The `blocktimes` action
   samples once per second, so measured intervals quantize to about
   1s or about 2s. The p95 and max near 2.1s are artifacts of that
   resolution, not evidence of 2-second blocks (see section 5).
6. **Snapshot archives include node.log.** Each snapshot tars the
   whole validator home, including its node.log (about 8 MB
   uncompressed at snapshot time). Compressed snapshot sizes were
   748K to 796K. Not a failure, recorded so a future snapshot
   procedure can exclude the log if size matters.
7. **One pre-existing em dash in copied code.** The copied
   `chain/x/ogocoin/bypass.go` contains one em dash in a code
   comment, inherited byte-identical from the prototype (the
   prototype report notes pre-existing em dashes remain in that
   file). No file authored for this PoC contains an em dash; the
   copied file was not edited. Verified with grep across every file
   written for this PoC (localnet.sh, run-poc.sh, README.md,
   RUNBOOK.md, this report, Makefile changes).

## 10. Demonstration results (runs 4 and 5, the two full clean runs)

Run 4 (2026-11-06, 16:46:27Z to 16:48:56Z, 149 seconds):

- Step 1: build and start 4 validators. Genesis valid:
  max_supply 1,000,010,000, total 1,000,000,000, denom `uogo`.
  Genesis accounts: alice 600,000,000, bob 400,000,000 (TEST-ONLY).
  All 4 nodes producing blocks. VERIFY OK at height 5.
- Step 2: first epoch boundary. Epoch state on val1:
  `current_epoch: 2`, `hook_fire_count: 1`, `total_issued: 1000`,
  rewards pool 1000, supply 1,000,001,000.
- Step 3: transfer 123,456 uogo from alice to bob, fee 100 uogo.
  Transaction hash
  `E0A821BDD9B63EEF1BF57EC44E9761D6A6159A435C747A1E77A2CF13B5A0B338`,
  code 0. Broadcast took 98 ms.
- Step 4: same balances from all four nodes: alice 599,876,444,
  bob 400,123,456.
- Step 5: verify the same ledger state on all four nodes at height
  16: VERIFY OK.
- Step 6: voting power: 10 per validator, total 40, on every node
  and in genesis.
- Step 7: S5. Height 17 before the outage; samples 18, 22, 25, 29;
  advanced 17 to 29 with val4 down; val4 RPC down during the outage;
  val4 restarted and caught up to 31; voting power 10 each, total 40,
  after the outage; VERIFY OK at height 31.
- Step 8: block times over 60s: 41 intervals, min 1.015s, median
  1.019s, p95 2.048s, max 2.056s, 41 of 41 within 1 to 10 seconds.
- Step 9: snapshot val4 (748K, 6s, height 75); restore val4 (OK, 7s,
  caught up to 79); final verify OK at height 79.
- Step 10: final status: all four nodes at height 80 with identical
  app hash `B67EBC765DB8AF43BE9328C486C750C74E0507C9A09109136D6B95A6F78E4F2F`.
  Nodes stopped cleanly on exit.

Run 5 (2026-11-06, 16:50:11Z to 16:52:35Z, 144 seconds): identical
outcomes in kind. Transaction hash
`F2AAD930EC8F978AE7F6D13A7D4909EF65C1AA59D37E8685DD4C7E66EA6FF470`,
broadcast took 80 ms. VERIFY OK at heights 5, 17, 31, and 76. S5:
height 18 before the outage; samples 18, 22, 26, 30; advanced 18 to
30; val4 caught up to 31; VERIFY OK at height 31. Block times over
60s: 40 intervals, min 1.015s, median 1.023s, p95 2.067s, max 2.096s,
40 of 40 within 1 to 10 seconds. Snapshot val4 (756K, 3s, height 74);
restore val4 (OK, 5s, caught up to 76); final verify OK at height 76.
Nodes stopped cleanly on exit.

Runs 1, 2, and 3 reached the same S5 result (section 4.2) before
their respective endings (run 1: restore failure, section 9 item 1;
run 2: verify race, section 9 item 2; run 3: aborted exec call,
section 9 item 3). The fixes for items 1 and 2 are in the shipped
scripts and were exercised by runs 4 and 5.

## 11. What this PoC does and does not prove

This PoC proves:

- With four equally weighted validators, the chain keeps producing
  blocks when one validator goes offline (3/4 voting power is a
  quorum), and the stopped validator restarts, catches up, and
  rejoins consensus with all four nodes agreeing on the same ledger.
  That is the framework-selection fault-tolerance question it was
  built to answer, and it answers it yes, with measured heights.
- The same ledger, transfer, and epoch behaviour as the
  three-validator prototype reproduces on four nodes with the same
  TEST-ONLY parameters.
- Block times over the observed 60-second windows sit at about 1
  second per block, 100 percent within 1 to 10 seconds.
- Snapshot and restore of a validator work on this machine and take
  seconds.

This PoC does NOT prove:

- That the chain survives 24 hours of continuous operation (P1 is
  partial; the observed windows are about 2.5 minutes per run).
- That downtime jailing, slashing, or unjailing work (S5 is partial;
  no such machinery exists in this scaffold).
- That a governance upgrade through Cosmovisor works, or that a
  store migration can add a state field (M2 is not executed).
- That validators can operate independently, or that a sentry
  topology works (O1 is partial; everything ran on one machine).
- That the chain is ready for mainnet, testnet, or any real value.
  It is a framework-selection fault-tolerance test in local staging
  only. No token was issued, no value moved, no deployment happened.
- That real SDK x/epochs wiring works. The epoch hook here is the
  same timer mirror as the prototype, not real SDK x/epochs wiring
  (no keeper, store, depinject, gRPC, or app-hash coverage of epoch
  state). The mirror hook-fires test does not prove depinject
  registration of the real module; that proof is Milestone C work.
- Throughput, latency, or load behaviour (P2, P3, P4 were not run),
  remote signing (S8 was not run), double-sign punishment (S6 was
  not run), or CosmWasm safety under load (S7, P3 were not run).

## 12. Files

- `run-poc.sh`: narrated end-to-end run.
- `chain/scripts/localnet.sh`: 4-validator localnet (build, init,
  start, stop, status, verify, voting, s5, blocktimes, snapshot,
  restore, clean).
- `chain/`: chain code copied from the prototype (go.mod/go.sum
  byte-identical; localnet-data/, the built ogod binary, and
  batchdemo-bin excluded and rebuilt fresh).
- `README.md`: what the PoC is, how it differs from the prototype,
  how to repeat the run.
- `RUNBOOK.md`: operator runbook for start, stop, upgrade, snapshot,
  and restore, with execution records.
- `logs/poc-run-20261106-164626.log`: full run 4 log.
- `logs/poc-run-20261106-165011.log`: full run 5 log.
- `logs/poc-run-20261106-163606.log`: run 1 log (restore failure).
- `logs/poc-run-20261106-163958.log`: run 2 log (verify race).
- `logs/poc-run-20261106-164400.log`: run 3 log (cut short).
- `logs/test-output-20261106.log`: full test output (90 pass, 0 fail).
- `logs/vet-20261106.log`: `go vet` output (exit 0).
- `chain/snapshots/`: validator snapshots from the runs.
