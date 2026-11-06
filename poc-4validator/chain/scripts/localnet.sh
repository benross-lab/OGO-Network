#!/usr/bin/env bash
# OGO Network 4-validator proof of concept localnet: 4 validators on
# one machine, epoch mode on, local staging only.
#
# Why 4 validators: Benross decided (2026-11-06, Option 1) on a
# separate, isolated four-validator framework-selection proof of
# concept, in addition to the completed three-validator Phase 2
# prototype, which stays unchanged. Purpose: fault tolerance.
# Technical distinction: with three equally weighted validators,
# losing one leaves 2/3 voting power, insufficient for CometBFT's
# greater-than-two-thirds quorum, so the chain halts. With four
# equally weighted validators, losing one leaves 3/4 voting power,
# so consensus continues. The three-validator halt-and-resume result
# is NOT fault tolerance and must not be redefined as such.
#
# This PoC is isolated from the prototype: chain-id ogo-poc-4val-1
# (prototype: ogo-local-1), RPC ports 26757, 26767, 26777, 26787 and
# P2P ports 26756, 26766, 26776, 26786 (prototype: 2665x, 2666x,
# 2667x), separate data directory chain/localnet-data/ in this PoC
# directory. Never point this script at the prototype directory.
#
# Keplr and live IBC are NOT prerequisites for this PoC, and the
# monetary prototype is independent of smart contracts: CosmWasm is
# evaluated separately before adoption. This localnet demonstrates
# consensus, epoch execution, native-coin transfers, and fault
# tolerance with one validator down, with no wallet, no IBC relaying,
# and no contracts.
#
# What this script does NOT do: no deployment, no token issuance, no
# production change, no spending. All values are TEST-ONLY and are
# labelled so everywhere they appear. The network runs entirely on
# this machine and never touches the OGO application: the app is a
# client of the network, not its authority. Nothing here reads from
# or writes to Supabase, the OGO Worker, or benross-lab/OGO.
#
# Usage:
#   scripts/localnet.sh build      build the ogod binary
#   scripts/localnet.sh init       (re)create the 4 validator homes and shared genesis
#   scripts/localnet.sh start      start all 4 nodes in the background
#   scripts/localnet.sh stop       stop all 4 nodes
#   scripts/localnet.sh status     show height and app hash per node
#   scripts/localnet.sh verify     verify the same ledger state on all 4 nodes
#   scripts/localnet.sh voting     show voting power per validator (RPC and genesis)
#   scripts/localnet.sh s5         fault tolerance: stop val4, confirm blocks
#                                  continue with 3 of 4 validators, restart
#                                  val4, confirm it catches up and nodes agree
#   scripts/localnet.sh liveness   alias for s5
#   scripts/localnet.sh blocktimes [seconds]  sample block heights over a
#                                  window (default 60s) and report the
#                                  block-time distribution
#   scripts/localnet.sh snapshot <node>   stop one node, snapshot its home,
#                                  restart it
#   scripts/localnet.sh restore <node> <snapshot-file>  stop one node, delete
#                                  its data, restore from snapshot, restart it
#   scripts/localnet.sh up         build + init + start + verify
#   scripts/localnet.sh down       stop (alias)
#   scripts/localnet.sh clean      stop and delete all localnet data
#
# S5 scope note: this script demonstrates the LIVENESS half of S5
# (chain continues with 3 of 4 validators, restart and rejoin). The
# jailing, slashing, and unjail half of S5 as written cannot be
# executed on this scaffold: x/staking and x/slashing are pinned and
# compiled only, not wired (blocker B6), so there is no downtime
# jailing or slashing machinery to observe. Restarting the validator
# and verifying it syncs and rejoins consensus is the executable
# part; unjailing as a staking operation is not. The full S5 as
# written remains open for Milestone C when the keepers are wired.
#
# Data lives in chain/localnet-data/ (override with LOCALNET_DIR).
# Logs live in chain/localnet-data/<node>/node.log.
set -euo pipefail

CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${LOCALNET_DIR:-$CHAIN_DIR/localnet-data}"
SNAP_DIR="${SNAPSHOT_DIR:-$CHAIN_DIR/snapshots}"
OGOD="$CHAIN_DIR/ogod"
GO_BIN="${GO_BIN:-$HOME/workspace/tools/go-1.25.7/bin/go}"
export TMPDIR="${TMPDIR:-$HOME/workspace/.go-tmp}"
export GOCACHE="${GOCACHE:-$HOME/workspace/.go-build-cache}"

NODES=(val1 val2 val3 val4)
RPC_PORTS=(26757 26767 26777 26787)
P2P_PORTS=(26756 26766 26776 26786)

# TEST-ONLY parameters. The production max supply, schedule, and epoch
# duration are all TBD and have no defaults anywhere. These values
# match the three-validator prototype exactly so results are comparable.
MAX_SUPPLY=1000010000
SCHED_START=1
SCHED_PERIOD_LEN=100
SCHED_PERIODS=4
SCHED_INITIAL=1000
SCHED_REDUCTION="9/10"
EPOCH_SECONDS=20
ALICE_AMOUNT=600000000
BOB_AMOUNT=400000000

mkdir -p "$TMPDIR" "$GOCACHE"

node_home() { echo "$DATA_DIR/$1"; }
node_rpc() {
  local i=0
  for n in "${NODES[@]}"; do
    if [ "$n" = "$1" ]; then echo "${RPC_PORTS[$i]}"; return; fi
    i=$((i + 1))
  done
}
node_p2p() {
  local i=0
  for n in "${NODES[@]}"; do
    if [ "$n" = "$1" ]; then echo "${P2P_PORTS[$i]}"; return; fi
    i=$((i + 1))
  done
}

rpc_height() {
  local port="$1"
  curl -s --max-time 5 "http://127.0.0.1:$port/status" 2>/dev/null | jq -r '.result.sync_info.latest_block_height // empty' 2>/dev/null || true
}

rpc_app_hash_at() {
  local port="$1" height="$2"
  curl -s --max-time 5 "http://127.0.0.1:$port/block?height=$height" 2>/dev/null | jq -r '.result.block.header.app_hash // empty' 2>/dev/null || true
}

rpc_validators() {
  local port="$1"
  curl -s --max-time 5 "http://127.0.0.1:$port/validators" 2>/dev/null || true
}

cmd_build() {
  echo "building ogod..."
  (cd "$CHAIN_DIR" && "$GO_BIN" build -o "$OGOD" ./cmd/ogod)
  echo "built $OGOD"
}

cmd_init() {
  cmd_build
  # Safety: never wipe the data directory while nodes are running.
  # Wiping under live nodes leaves them writing state into the new
  # homes and the new nodes failing to bind the same ports (caught
  # 2026-11-06 in the prototype: a stale val3 survived a re-init and
  # answered RPC while its replacement died on the bind).
  for n in "${NODES[@]}"; do
    local pidfile="$DATA_DIR/$n/node.pid"
    if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
      echo "ERROR: $n is still running (pid $(cat "$pidfile")); run: $0 stop" >&2
      exit 1
    fi
    if curl -s --max-time 2 "http://127.0.0.1:$(node_rpc "$n")/status" >/dev/null 2>&1; then
      echo "ERROR: something is already answering on $n's RPC port 127.0.0.1:$(node_rpc "$n") (stale node?); stop it first" >&2
      exit 1
    fi
  done
  echo "initializing 4-validator localnet in $DATA_DIR (TEST-ONLY values)"
  rm -rf "$DATA_DIR"
  mkdir -p "$DATA_DIR"
  local homes=""
  for n in "${NODES[@]}"; do
    local home
    home="$(node_home "$n")"
    "$OGOD" init --home "$home" --chain-id ogo-poc-4val-1 \
      --max-supply "$MAX_SUPPLY" \
      --schedule-start "$SCHED_START" \
      --schedule-period-length "$SCHED_PERIOD_LEN" \
      --schedule-periods "$SCHED_PERIODS" \
      --schedule-initial-per-period "$SCHED_INITIAL" \
      --schedule-reduction "$SCHED_REDUCTION" \
      --epoch-duration-seconds "$EPOCH_SECONDS" \
      --epoch-identifier monetary >/dev/null
    echo "  init $n"
    homes="${homes:+$homes,}$home"
  done
  # Genesis accounts live in val1's genesis (the shared one).
  local alice bob
  alice="$("$OGOD" keys add alice --home "$(node_home val1)" | sed -n 's/^key "alice" created: //p')"
  bob="$("$OGOD" keys add bob --home "$(node_home val1)" | sed -n 's/^key "bob" created: //p')"
  echo "  alice: $alice"
  echo "  bob:   $bob"
  "$OGOD" add-genesis-account "$alice" "$ALICE_AMOUNT" --purpose public --home "$(node_home val1)"
  "$OGOD" add-genesis-account "$bob" "$BOB_AMOUNT" --purpose public --home "$(node_home val1)"
  # One shared genesis with all 4 validators, copied to every home.
  "$OGOD" set-genesis-validators --home "$(node_home val1)" --homes "$homes"
  for n in val2 val3 val4; do
    cp "$(node_home val1)/config/genesis.json" "$(node_home "$n")/config/genesis.json"
  done
  "$OGOD" validate-genesis --home "$(node_home val1)"
  echo "localnet initialized: 4 validators, epoch duration ${EPOCH_SECONDS}s (TEST-ONLY)"
  echo "genesis accounts: alice $ALICE_AMOUNT, bob $BOB_AMOUNT (TEST-ONLY)"
}

peer_list_for() {
  local self="$1" peers=""
  for n in "${NODES[@]}"; do
    if [ "$n" = "$self" ]; then continue; fi
    local id
    id="$("$OGOD" show-node-id --home "$(node_home "$n")")"
    peers="${peers:+$peers,}${id}@127.0.0.1:$(node_p2p "$n")"
  done
  echo "$peers"
}

cmd_start() {
  if [ ! -f "$(node_home val1)/config/genesis.json" ]; then
    echo "localnet not initialized; run: $0 init" >&2
    exit 1
  fi
  if [ ! -x "$OGOD" ]; then cmd_build; fi
  echo "starting 4 validators..."
  for n in "${NODES[@]}"; do
    local home log pidfile
    home="$(node_home "$n")"
    log="$home/node.log"
    pidfile="$home/node.pid"
    if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
      echo "  $n already running (pid $(cat "$pidfile"))"
      continue
    fi
    nohup "$OGOD" start --home "$home" --moniker "$n" \
      --rpc-laddr "tcp://127.0.0.1:$(node_rpc "$n")" \
      --p2p-laddr "tcp://127.0.0.1:$(node_p2p "$n")" \
      --persistent-peers "$(peer_list_for "$n")" \
      >"$log" 2>&1 &
    echo $! >"$pidfile"
    echo "  $n started (pid $(cat "$pidfile"), rpc 127.0.0.1:$(node_rpc "$n"), log $log)"
  done
  echo "waiting for all 4 nodes to reach height 5..."
  local deadline=$((SECONDS + 120))
  while [ "$SECONDS" -lt "$deadline" ]; do
    local ok=1
    for n in "${NODES[@]}"; do
      local h
      h="$(rpc_height "$(node_rpc "$n")")"
      if [ -z "$h" ] || [ "$h" -lt 5 ]; then ok=0; break; fi
    done
    if [ "$ok" = 1 ]; then
      echo "all 4 nodes producing blocks"
      return 0
    fi
    sleep 2
  done
  echo "ERROR: nodes did not reach height 5 within 120s" >&2
  echo "check logs in $DATA_DIR/*/node.log" >&2
  exit 1
}

cmd_stop() {
  local stopped=0
  local pids=""
  for n in "${NODES[@]}"; do
    local pidfile="$DATA_DIR/$n/node.pid"
    if [ -f "$pidfile" ]; then
      local pid
      pid="$(cat "$pidfile")"
      if kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
        echo "  stopped $n (pid $pid)"
        stopped=1
        pids="$pids $pid"
      fi
      rm -f "$pidfile"
    fi
  done
  # CometBFT takes a few seconds to shut down after SIGTERM; wait
  # until the processes are actually gone so a caller does not find
  # live nodes right after this returns.
  if [ -n "$pids" ]; then
    local deadline=$((SECONDS + 15))
    while [ "$SECONDS" -lt "$deadline" ]; do
      local alive=0
      for pid in $pids; do
        if kill -0 "$pid" 2>/dev/null; then alive=1; break; fi
      done
      if [ "$alive" = 0 ]; then break; fi
      sleep 1
    done
  fi
  # A node started by hand (no pid file) is left alone; only the four
  # managed homes are stopped here.
  if [ "$stopped" = 0 ]; then echo "no managed nodes running"; fi
}

# stop_one / start_one: used by the s5 test and by snapshot/restore.
stop_one() {
  local n="$1"
  local pidfile="$DATA_DIR/$n/node.pid"
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
    local pid
    pid="$(cat "$pidfile")"
    kill "$pid"
    rm -f "$pidfile"
    # Wait for the process to actually exit (CometBFT shutdown takes
    # a few seconds); the replacement node cannot bind the same ports
    # until the old one is gone.
    local deadline=$((SECONDS + 10))
    while [ "$SECONDS" -lt "$deadline" ] && kill -0 "$pid" 2>/dev/null; do
      sleep 1
    done
    echo "  stopped $n"
  fi
}

start_one() {
  local n="$1"
  local home log pidfile
  home="$(node_home "$n")"
  log="$home/node.log"
  pidfile="$home/node.pid"
  nohup "$OGOD" start --home "$home" --moniker "$n" \
    --rpc-laddr "tcp://127.0.0.1:$(node_rpc "$n")" \
    --p2p-laddr "tcp://127.0.0.1:$(node_p2p "$n")" \
    --persistent-peers "$(peer_list_for "$n")" \
    >>"$log" 2>&1 &
  echo $! >"$pidfile"
  echo "  $n restarted (pid $(cat "$pidfile"))"
}

wait_height() {
  # wait_height <node> <min-height> <timeout-secs>
  local n="$1"
  local want="$2"
  local timeout="$3"
  local deadline=$((SECONDS + timeout))
  while [ "$SECONDS" -lt "$deadline" ]; do
    local h
    h="$(rpc_height "$(node_rpc "$n")")"
    if [ -n "$h" ] && [ "$h" -ge "$want" ]; then return 0; fi
    sleep 2
  done
  return 1
}

cmd_status() {
  for n in "${NODES[@]}"; do
    local port h
    port="$(node_rpc "$n")"
    h="$(rpc_height "$port")"
    if [ -z "$h" ]; then
      echo "$n: DOWN (rpc 127.0.0.1:$port)"
    else
      local hash
      hash="$(rpc_app_hash_at "$port" "$h")"
      echo "$n: height=$h app_hash=$hash rpc=127.0.0.1:$port"
    fi
  done
}

# query_one <node> <query...> prints the query output for one node.
query_one() {
  local n="$1"; shift
  "$OGOD" query "$@" --node "tcp://127.0.0.1:$(node_rpc "$n")"
}

cmd_voting() {
  echo "voting power per validator (CometBFT /validators RPC, TEST-ONLY network)"
  echo "genesis validator set (val1 genesis.json):"
  jq -r '.validators[] | "  power=\(.power) address=\(.address)"' "$(node_home val1)/config/genesis.json"
  local total=0
  for n in "${NODES[@]}"; do
    local vals
    vals="$(rpc_validators "$(node_rpc "$n")")"
    if [ -z "$vals" ]; then
      echo "$n: RPC DOWN, cannot read voting power"
      continue
    fi
    echo "$n (rpc 127.0.0.1:$(node_rpc "$n")):"
    echo "$vals" | jq -r '.result.validators[] | "  power=\(.voting_power) address=\(.address)"'
    local sum
    sum="$(echo "$vals" | jq -r '[.result.validators[].voting_power | tonumber] | add')"
    echo "  total voting power reported by $n: $sum"
    total="$sum"
  done
  echo "genesis total voting power: $(jq -r '[.validators[].power | tonumber] | add' "$(node_home val1)/config/genesis.json")"
}

cmd_verify() {
  echo "verifying the same ledger state on all 4 nodes..."
  local alice bob
  alice="$(jq -r .address "$DATA_DIR/val1/keys/alice.json")"
  bob="$(jq -r .address "$DATA_DIR/val1/keys/bob.json")"
  # Pick one height and compare the app hash at that height on all
  # nodes: the state machine is deterministic, so the same height
  # must give the same app hash everywhere. The block header's app
  # hash is the ledger state after executing that block. The live
  # balance, supply, and issuance queries below are read a moment
  # later, so they may reflect a block or two past the compared
  # height; that is fine, they are compared across nodes at the same
  # moment.
  local height
  height="$(rpc_height "$(node_rpc val1)")"
  if [ -z "$height" ]; then
    echo "ERROR: val1 is not responding" >&2
    exit 1
  fi
  echo "comparing at height $height"
  # Wait for every node to actually reach the comparison height
  # before asking for its app hash there. A node one block behind
  # (normal right after a restart and catch-up) returns an empty hash
  # for a height it has not committed yet, which is a missing block,
  # not a disagreement. Comparing before waiting misreads that as a
  # mismatch (caught 2026-11-06 in PoC run 2: final verify at height
  # 79 failed with an empty val4 hash one second after a restore
  # that had caught val4 up to 78 while val1 stood at 79).
  for n in val2 val3 val4; do
    if ! wait_height "$n" "$height" 30; then
      echo "VERIFY FAILED: $n did not reach height $height within 30s" >&2
      exit 1
    fi
  done
  local fail=0
  local ref_hash ref_alice ref_bob ref_supply ref_epochs
  ref_hash="$(rpc_app_hash_at "$(node_rpc val1)" "$height")"
  echo "val1: app_hash=$ref_hash"
  for n in val2 val3 val4; do
    local hash
    hash="$(rpc_app_hash_at "$(node_rpc "$n")" "$height")"
    echo "$n: app_hash=$hash"
    if [ "$hash" != "$ref_hash" ]; then echo "MISMATCH: $n app_hash differs" >&2; fail=1; fi
  done
  if [ "$fail" = 1 ]; then
    echo "VERIFY FAILED: nodes disagree on app hash at height $height" >&2
    exit 1
  fi
  # Live balances, supply, and issuance are compared in a retry loop.
  # The header hash above is the same-height proof; the live queries
  # are read sequentially across nodes, and an epoch boundary can fall
  # between two reads (the rewards pool and supply change at the
  # boundary while the other nodes have not been read yet). That is a
  # sampling race, not a disagreement: all nodes converge to the same
  # values once the boundary has passed, so retry until they agree.
  local attempt=0
  while [ "$attempt" -lt 10 ]; do
    attempt=$((attempt + 1))
    ref_alice="$(query_one val1 balance "$alice" | jq -r .balance)"
    ref_bob="$(query_one val1 balance "$bob" | jq -r .balance)"
    ref_supply="$(query_one val1 supply | jq -r .total_supply)"
    ref_epochs="$(query_one val1 epochs | jq -r .total_issued)"
    local live_fail=0
    for n in val2 val3 val4; do
      local bal_a bal_b supply issued
      bal_a="$(query_one "$n" balance "$alice" | jq -r .balance)"
      bal_b="$(query_one "$n" balance "$bob" | jq -r .balance)"
      supply="$(query_one "$n" supply | jq -r .total_supply)"
      issued="$(query_one "$n" epochs | jq -r .total_issued)"
      if [ "$bal_a" != "$ref_alice" ] || [ "$bal_b" != "$ref_bob" ] || [ "$supply" != "$ref_supply" ] || [ "$issued" != "$ref_epochs" ]; then
        live_fail=1
      fi
    done
    if [ "$live_fail" = 0 ]; then
      echo "val1: alice=$ref_alice bob=$ref_bob supply=$ref_supply issued=$ref_epochs"
      echo "val2: alice=$ref_alice bob=$ref_bob supply=$ref_supply issued=$ref_epochs"
      echo "val3: alice=$ref_alice bob=$ref_bob supply=$ref_supply issued=$ref_epochs"
      echo "val4: alice=$ref_alice bob=$ref_bob supply=$ref_supply issued=$ref_epochs"
      echo "VERIFY OK: all 4 nodes agree on balances, supply, issuance, and app hash at height $height"
      return 0
    fi
    sleep 2
  done
  echo "VERIFY FAILED: live balances/supply/issuance did not converge across nodes after 10 attempts" >&2
  exit 1
}

cmd_s5() {
  # S5 fault-tolerance test, liveness half (see the scope note at the
  # top of this script for the jailing/slashing half, which is not
  # executable on this scaffold). With 4 equally weighted validators,
  # stopping one leaves 3/4 voting power, above CometBFT's
  # greater-than-two-thirds quorum, so blocks MUST keep being
  # produced by the remaining three validators. This is the behaviour
  # the 3-validator prototype cannot show: there, 2/3 is not a
  # quorum and the chain halts.
  echo "S5 fault-tolerance test: 1 of 4 validators offline, chain must continue"
  echo "--- voting power before the outage ---"
  cmd_voting
  local h_before
  h_before="$(rpc_height "$(node_rpc val1)")"
  if [ -z "$h_before" ]; then echo "ERROR: val1 is not responding" >&2; exit 1; fi
  echo "height on val1 before the outage: $h_before"
  echo "stopping val4 cleanly..."
  stop_one val4
  echo "val4 stopped at val1 height $h_before; sampling val1 height at 0s, 10s, 20s, 30s after the stop..."
  local h0 h10 h20 h30
  h0="$(rpc_height "$(node_rpc val1)")"
  echo "  t+0s:  val1 height=$h0"
  sleep 10
  h10="$(rpc_height "$(node_rpc val1)")"
  echo "  t+10s: val1 height=$h10"
  sleep 10
  h20="$(rpc_height "$(node_rpc val1)")"
  echo "  t+20s: val1 height=$h20"
  sleep 10
  h30="$(rpc_height "$(node_rpc val1)")"
  echo "  t+30s: val1 height=$h30"
  # Confirm val4's RPC is down during the outage.
  if [ -n "$(rpc_height "$(node_rpc val4)")" ]; then
    echo "S5 FAILED: val4 RPC is still answering during the outage" >&2
    start_one val4
    exit 1
  fi
  echo "val4 RPC confirmed DOWN during the outage"
  if [ -z "$h30" ]; then
    echo "S5 FAILED: val1 stopped responding during the outage" >&2
    start_one val4
    exit 1
  fi
  if [ "$h30" -le "$h_before" ]; then
    echo "S5 FAILED: height did not advance with val4 down ($h_before -> $h30); chain halted instead of continuing" >&2
    start_one val4
    exit 1
  fi
  echo "S5 LIVENESS OK: height advanced $h_before -> $h30 with val4 down (samples: $h0, $h10, $h20, $h30)"
  echo "restarting val4 and waiting for it to catch up..."
  start_one val4
  local h_target
  h_target="$(rpc_height "$(node_rpc val1)")"
  if ! wait_height val4 "$h_target" 120; then
    echo "S5 FAILED: val4 did not catch up to height $h_target within 120s" >&2
    exit 1
  fi
  echo "val4 caught up to height $(rpc_height "$(node_rpc val4)") (val1: $(rpc_height "$(node_rpc val1)"))"
  echo "--- voting power after the outage ---"
  cmd_voting
  echo "verifying all 4 nodes agree after the outage..."
  cmd_verify
  echo "S5 COMPLETE: blocks continued with 1 of 4 validators offline, val4 restarted and synced, all 4 nodes agree"
}

cmd_blocktimes() {
  # blocktimes [seconds]: sample block heights over a time window and
  # report the block-time distribution. Samples val1 height once per
  # second, records the first-seen time of each height, and treats the
  # gaps between successive first-seen times as block intervals.
  local window="${1:-60}"
  local stamp
  stamp="$(date +%Y%m%d-%H%M%S)"
  local csv="$DATA_DIR/blocktimes-$stamp.csv"
  echo "sampling block heights on val1 for ${window}s (one sample per second)..."
  local start_s end_s
  start_s="$(date +%s)"
  end_s=$((start_s + window))
  : >"$csv"
  while [ "$(date +%s)" -lt "$end_s" ]; do
    local h now
    h="$(rpc_height "$(node_rpc val1)")"
    now="$(date +%s.%N)"
    if [ -n "$h" ]; then echo "$now,$h" >>"$csv"; fi
    sleep 1
  done
  echo "samples saved to $csv"
  python3 - "$csv" <<'PYEOF'
import statistics, sys
path = sys.argv[1]
seen = {}
with open(path) as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        ts, h = line.split(",")
        h = int(h)
        if h not in seen:
            seen[h] = float(ts)
heights = sorted(seen)
intervals = [seen[b] - seen[a] for a, b in zip(heights, heights[1:])]
print("blocktimes: heights observed: %d (first %d, last %d)" % (len(heights), heights[0], heights[-1]))
print("blocktimes: intervals measured: %d" % len(intervals))
if not intervals:
    print("blocktimes: no intervals measured; window too short or chain halted")
    sys.exit(1)
intervals_sorted = sorted(intervals)
n = len(intervals_sorted)
median = statistics.median(intervals_sorted)
p95 = intervals_sorted[min(n - 1, int(0.95 * n))]
within = sum(1 for x in intervals if 1.0 <= x <= 10.0)
print("blocktimes: count=%d min=%.3fs median=%.3fs p95=%.3fs max=%.3fs" % (n, min(intervals), median, p95, max(intervals)))
print("blocktimes: within 1 to 10 seconds: %d of %d (%.1f percent)" % (within, n, 100.0 * within / n))
print("blocktimes: all intervals (seconds): %s" % ", ".join("%.3f" % x for x in intervals))
PYEOF
}

cmd_snapshot() {
  # snapshot <node>: stop the node, tar its home, restart it.
  local n="${1:-}"
  if [ -z "$n" ]; then echo "usage: $0 snapshot <node>" >&2; exit 1; fi
  local stamp size file
  stamp="$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$SNAP_DIR"
  file="$SNAP_DIR/$n-$stamp.tar.gz"
  echo "snapshotting $n to $file ..."
  local t0 t1
  t0="$(date +%s)"
  stop_one "$n"
  tar -czf "$file" -C "$DATA_DIR" "$n"
  size="$(du -h "$file" | cut -f1)"
  echo "  snapshot size: $size"
  start_one "$n"
  if ! wait_height "$n" "$(rpc_height "$(node_rpc val1)")" 120; then
    echo "SNAPSHOT FAILED: $n did not catch up within 120s of restart" >&2
    exit 1
  fi
  t1="$(date +%s)"
  echo "SNAPSHOT OK: $n snapshot $file size $size, stop plus snapshot plus restart took $((t1 - t0))s, $n height $(rpc_height "$(node_rpc "$n")")"
  echo "$file"
}

cmd_restore() {
  # restore <node> <snapshot-file>: stop the node, delete its data
  # directory, restore from the snapshot, restart it, verify catch-up.
  local n="${1:-}" file="${2:-}"
  if [ -z "$n" ] || [ -z "$file" ]; then echo "usage: $0 restore <node> <snapshot-file>" >&2; exit 1; fi
  if [ ! -f "$file" ]; then echo "ERROR: snapshot file not found: $file" >&2; exit 1; fi
  echo "restoring $n from snapshot $file ..."
  local t0 t1
  t0="$(date +%s)"
  stop_one "$n"
  rm -rf "$(node_home "$n")"
  # --no-same-owner: this environment forbids chown (the prototype
  # copy showed the same: cp -a cannot preserve ownership here), so
  # a plain extraction fails with "Cannot change ownership" even
  # though the files extract. Without the flag the tar exit code is
  # non-zero and, under set -e, aborts the restore before the node
  # restarts (caught 2026-11-06 in PoC run 1: snapshot OK, restore
  # aborted on the tar exit code, val4 left down).
  tar --no-same-owner -xzf "$file" -C "$DATA_DIR"
  rm -f "$(node_home "$n")/node.pid"
  start_one "$n"
  local h_target
  h_target="$(rpc_height "$(node_rpc val1)")"
  if ! wait_height "$n" "$h_target" 120; then
    echo "RESTORE FAILED: $n did not catch up to height $h_target within 120s" >&2
    exit 1
  fi
  t1="$(date +%s)"
  echo "RESTORE OK: $n restored from $file, caught up to height $(rpc_height "$(node_rpc "$n")") (val1: $(rpc_height "$(node_rpc val1)")), restore plus catch-up took $((t1 - t0))s"
}

cmd_clean() {
  cmd_stop || true
  rm -rf "$DATA_DIR"
  echo "cleaned $DATA_DIR"
}

usage() {
  sed -n '2,60p' "$0"
}

case "${1:-up}" in
  build) cmd_build ;;
  init) cmd_init ;;
  start) cmd_start ;;
  stop|down) cmd_stop ;;
  status) cmd_status ;;
  verify) cmd_verify ;;
  voting) cmd_voting ;;
  s5|liveness) cmd_s5 ;;
  blocktimes) cmd_blocktimes "${2:-60}" ;;
  snapshot) cmd_snapshot "${2:-}" ;;
  restore) cmd_restore "${2:-}" "${3:-}" ;;
  up) cmd_init; cmd_start; cmd_verify ;;
  clean) cmd_clean ;;
  *) usage; exit 1 ;;
esac
