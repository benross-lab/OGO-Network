#!/usr/bin/env bash
# OGO Network localnet: 3 validators on one machine, epoch mode on.
#
# Why 3 validators: Benross decided the authorized local prototype
# uses THREE validators. This is the execution requirement for the
# local prototype only, and it is distinct from the larger targets
# proposed for later networks: the minimum operational set is FOUR
# validators (BFT consensus needs 3f+1 to tolerate f faulty
# validators, so 4 tolerates 1 faulty and is the smallest set that
# can demonstrate liveness with one validator down), and the
# decentralization target for a public network is FIFTY or more
# validators. With 3 validators f is 0: the network halts if any one
# goes down, and that halt is expected behaviour, not a failure.
# The liveness action below therefore tests halt-and-resume (stop
# one validator, confirm the chain halts, restart it, confirm blocks
# resume), NOT continued block production with a validator down.
# Continued production with 1 of 4 offline is the S5 test for the
# 4-validator minimum operational set, a later-network target.
#
# Keplr and live IBC are NOT prerequisites for this prototype, and
# the monetary prototype is independent of smart contracts:
# CosmWasm is evaluated separately before adoption. This localnet
# demonstrates consensus, epoch execution, and native-coin transfers
# with no wallet, no IBC relaying, and no contracts.
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
#   scripts/localnet.sh init       (re)create the 3 validator homes and shared genesis
#   scripts/localnet.sh start      start all 3 nodes in the background
#   scripts/localnet.sh stop       stop all 3 nodes
#   scripts/localnet.sh status     show height and app hash per node
#   scripts/localnet.sh verify     verify the same ledger state on all 3 nodes
#   scripts/localnet.sh liveness   halt-and-resume: stop val3, confirm the chain
#                                  halts (expected with 3 validators), restart it,
#                                  confirm blocks resume and nodes agree
#   scripts/localnet.sh up         build + init + start + verify
#   scripts/localnet.sh down       stop (alias)
#   scripts/localnet.sh clean      stop and delete all localnet data
#
# Data lives in chain/localnet-data/ (override with LOCALNET_DIR).
# Logs live in chain/localnet-data/<node>/node.log.
set -euo pipefail

CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${LOCALNET_DIR:-$CHAIN_DIR/localnet-data}"
OGOD="$CHAIN_DIR/ogod"
GO_BIN="${GO_BIN:-$HOME/workspace/tools/go-1.25.7/bin/go}"
export TMPDIR="${TMPDIR:-$HOME/workspace/.go-tmp}"
export GOCACHE="${GOCACHE:-$HOME/workspace/.go-build-cache}"

NODES=(val1 val2 val3)
RPC_PORTS=(26657 26667 26677)
P2P_PORTS=(26656 26666 26676)

# TEST-ONLY parameters. The production max supply, schedule, and epoch
# duration are all TBD and have no defaults anywhere.
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
  # 2026-11-06: a stale val3 survived a re-init and answered RPC
  # while its replacement died on the bind).
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
  echo "initializing 3-validator localnet in $DATA_DIR (TEST-ONLY values)"
  rm -rf "$DATA_DIR"
  mkdir -p "$DATA_DIR"
  local homes=""
  for n in "${NODES[@]}"; do
    local home
    home="$(node_home "$n")"
    "$OGOD" init --home "$home" --chain-id ogo-local-1 \
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
  # One shared genesis with all 3 validators, copied to every home.
  "$OGOD" set-genesis-validators --home "$(node_home val1)" --homes "$homes"
  for n in val2 val3; do
    cp "$(node_home val1)/config/genesis.json" "$(node_home "$n")/config/genesis.json"
  done
  "$OGOD" validate-genesis --home "$(node_home val1)"
  echo "localnet initialized: 3 validators, epoch duration ${EPOCH_SECONDS}s (TEST-ONLY)"
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
  echo "starting 3 validators..."
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
  echo "waiting for all 3 nodes to reach height 5..."
  local deadline=$((SECONDS + 120))
  while [ "$SECONDS" -lt "$deadline" ]; do
    local ok=1
    for n in "${NODES[@]}"; do
      local h
      h="$(rpc_height "$(node_rpc "$n")")"
      if [ -z "$h" ] || [ "$h" -lt 5 ]; then ok=0; break; fi
    done
    if [ "$ok" = 1 ]; then
      echo "all 3 nodes producing blocks"
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
  # A node started by hand (no pid file) is left alone; only the three
  # managed homes are stopped here.
  if [ "$stopped" = 0 ]; then echo "no managed nodes running"; fi
}

# stop_one / start_one: used by the liveness test for a single node.
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

cmd_verify() {
  echo "verifying the same ledger state on all 3 nodes..."
  local alice bob
  alice="$(jq -r .address "$DATA_DIR/val1/keys/alice.json")"
  bob="$(jq -r .address "$DATA_DIR/val1/keys/bob.json")"
  # Pick one height and compare the app hash at that height on all
  # nodes: the state machine is deterministic, so the same height
  # must give the same app hash everywhere. The block header's app
  # hash is the ledger state after executing that block (verified
  # live: it equals the abci_info hash at the same height). The live
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
  local fail=0
  local ref_hash ref_alice ref_bob ref_supply ref_epochs
  ref_hash="$(rpc_app_hash_at "$(node_rpc val1)" "$height")"
  echo "val1: app_hash=$ref_hash"
  for n in val2 val3; do
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
    for n in val2 val3; do
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
      echo "VERIFY OK: all 3 nodes agree on balances, supply, issuance, and app hash at height $height"
      return 0
    fi
    sleep 2
  done
  echo "VERIFY FAILED: live balances/supply/issuance did not converge across nodes after 10 attempts" >&2
  exit 1
}

cmd_liveness() {
  # Halt-and-resume test for the 3-validator local prototype.
  # With 3 validators the network tolerates ZERO faulty validators:
  # stopping one leaves 2 of 3 (exactly two thirds, not more), so no
  # quorum and block production must HALT. That halt is the expected
  # BFT behaviour for this validator count, not a failure. Restarting
  # the stopped validator must resume block production, with all 3
  # nodes agreeing afterwards. Continued block production with one
  # validator down is the S5 test for the FOUR-validator minimum
  # operational set, a later-network target, not this prototype.
  echo "halt-and-resume test: stopping val3..."
  local h_before h_after
  h_before="$(rpc_height "$(node_rpc val1)")"
  if [ -z "$h_before" ]; then echo "ERROR: val1 is not responding" >&2; exit 1; fi
  stop_one val3
  echo "val3 stopped at val1 height $h_before; waiting 15s: with 3 validators the chain must HALT..."
  sleep 15
  h_after="$(rpc_height "$(node_rpc val1)")"
  if [ -z "$h_after" ]; then
    echo "ERROR: val1 stopped responding" >&2
    start_one val3
    exit 1
  fi
  if [ "$h_after" -gt "$h_before" ]; then
    echo "UNEXPECTED: height advanced $h_before -> $h_after with val3 down; a 3-validator network must halt without quorum" >&2
    start_one val3
    exit 1
  fi
  echo "HALT OK: height stayed at $h_before with val3 down (no quorum with 2 of 3, as expected for 3 validators)"
  echo "restarting val3 and waiting for block production to resume..."
  start_one val3
  if ! wait_height val1 $((h_before + 1)) 120; then
    echo "RESUME FAILED: blocks did not resume within 120s of restarting val3" >&2
    exit 1
  fi
  echo "RESUME OK: height advanced past $h_before after val3 restarted"
  if ! wait_height val3 $((h_before + 1)) 120; then
    echo "RESUME FAILED: val3 did not catch up within 120s" >&2
    exit 1
  fi
  echo "val3 caught up to height $(rpc_height "$(node_rpc val3)")"
  cmd_verify
}

cmd_clean() {
  cmd_stop || true
  rm -rf "$DATA_DIR"
  echo "cleaned $DATA_DIR"
}

usage() {
  sed -n '2,30p' "$0"
}

case "${1:-up}" in
  build) cmd_build ;;
  init) cmd_init ;;
  start) cmd_start ;;
  stop|down) cmd_stop ;;
  status) cmd_status ;;
  verify) cmd_verify ;;
  liveness) cmd_liveness ;;
  up) cmd_init; cmd_start; cmd_verify ;;
  clean) cmd_clean ;;
  *) usage; exit 1 ;;
esac
