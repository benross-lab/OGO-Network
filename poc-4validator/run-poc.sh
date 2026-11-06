#!/usr/bin/env bash
# OGO Network 4-validator proof of concept: narrated end-to-end run.
#
# What this run demonstrates, end to end, on one machine, local
# staging only:
#   1. Four local validators start from one shared genesis and
#      produce blocks together (consensus on the same ledger).
#   2. Genesis accounts: alice and bob hold TEST-ONLY OGO Coin.
#   3. Epoch issuance executing: the monetary AfterEpochEnd hook
#      mints the schedule amount into the rewards pool at each epoch
#      boundary, and the run shows it happening live.
#   4. A transfer of TEST-ONLY OGO Coin between wallets, with the
#      fee and the elapsed broadcast time measured and printed, and
#      the same balances verified from all four nodes.
#   5. Voting power recorded for each validator (equal weight).
#   6. S5 fault tolerance (liveness half): val4 stops, blocks keep
#      being produced by the remaining three validators, val4
#      restarts, catches up, and all four nodes agree.
#   7. Block-time distribution measured over a 60 second window.
#   8. Snapshot and restore of val4 executed for real (O1 runbook).
#
# What this does NOT do: no deployment, no token issuance, no
# production change, no spending. All values are TEST-ONLY. The
# network runs entirely on this machine and never touches the OGO
# application: the application is a client of the network, not its
# authority. Nothing here reads from or writes to Supabase, the OGO
# Worker, or benross-lab/OGO.
#
# S5 scope: the liveness half is demonstrated here. Downtime jailing,
# slashing, and unjailing as a staking operation cannot be executed
# on this scaffold (x/staking and x/slashing are pinned and compiled
# only, not wired; blocker B6). See the PoC report for the full
# S5, P1, M2, and O1 assessments.
#
# P1 scope: a 24-hour continuous run cannot be completed in one
# session. This run measures the block-time distribution over the
# observed window (from first start to final stop) and records P1 as
# PARTIALLY EXECUTED. Re-running this script for a 24-hour
# observation is the path to complete P1.
#
# Usage: ./run-poc.sh
# The script stops the nodes when it finishes (or on failure),
# so it is safe to re-run. All output goes to logs/ with timestamps
# in filenames.
set -euo pipefail

POC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$POC_DIR/chain"
LOCALNET="$CHAIN_DIR/scripts/localnet.sh"
OGOD="$CHAIN_DIR/ogod"
LOG_DIR="$POC_DIR/logs"
export TMPDIR="${TMPDIR:-$HOME/workspace/.go-tmp}"
export GOCACHE="${GOCACHE:-$HOME/workspace/.go-build-cache}"

STAMP="$(date +%Y%m%d-%H%M%S)"
DAY="$(date +%Y%m%d)"
LOG="$LOG_DIR/poc-run-$DAY.log"
RUN_LOG="$LOG_DIR/poc-run-$STAMP.log"
NODES=(val1 val2 val3 val4)
RPC_PORTS=(26757 26767 26777 26787)

mkdir -p "$LOG_DIR"
# Both a dated log (poc-run-YYYYMMDD.log, overwritten per day) and a
# per-run timestamped log exist; the per-run log is the full record.
exec > >(tee -a "$RUN_LOG" "$LOG") 2>&1

RUN_START_S="$(date +%s)"
RUN_START_HUMAN="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

say() { echo; echo "=== $1 ==="; }

cleanup() {
  echo
  echo "stopping the localnet..."
  "$LOCALNET" stop >/dev/null 2>&1 || true
  local end_s
  end_s="$(date +%s)"
  echo "PoC run wall-clock observed window: $((end_s - RUN_START_S)) seconds (started $RUN_START_HUMAN UTC, ended $(date -u +%Y-%m-%dT%H:%M:%SZ) UTC)"
  echo "full log: $RUN_LOG"
}
trap cleanup EXIT

say "OGO Network 4-validator proof of concept run, all values TEST-ONLY"
echo "Run started: $RUN_START_HUMAN UTC"
echo "Four validators, one shared genesis, epoch mode on (20s epochs)."
echo "Chain: ogo-poc-4val-1. Denom: uogo (proposed, 6 decimals)."
echo "RPC ports: 26757, 26767, 26777, 26787. P2P ports: 26756, 26766, 26776, 26786."
echo "Log: $RUN_LOG"

say "Step 1: build and start 4 validators"
"$LOCALNET" up

ALICE="$(jq -r .address "$CHAIN_DIR/localnet-data/val1/keys/alice.json")"
BOB="$(jq -r .address "$CHAIN_DIR/localnet-data/val1/keys/bob.json")"
echo "alice: $ALICE"
echo "bob:   $BOB"

say "Step 2: wait for the first epoch boundary (epoch issuance executing)"
echo "Epoch duration is 20s (TEST-ONLY). Polling until the hook mints..."
deadline=$((SECONDS + 90))
while [ "$SECONDS" -lt "$deadline" ]; do
  issued="$("$OGOD" query epochs --node tcp://127.0.0.1:26757 | jq -r .total_issued)"
  if [ -n "$issued" ] && [ "$issued" != "0" ]; then break; fi
  sleep 2
done
echo "Epoch state on val1:"
"$OGOD" query epochs --node tcp://127.0.0.1:26757
echo "Rewards pool balance (the mint recipient):"
"$OGOD" query balance ogo1rewardspool --node tcp://127.0.0.1:26757
echo "Supply:"
"$OGOD" query supply --node tcp://127.0.0.1:26757

say "Step 3: transfer 123456 uogo from alice to bob (fee 100 uogo, TEST-ONLY)"
echo "Measuring broadcast time (fee is flat per transaction in this scaffold; there is no gas metering yet)."
start_ns="$(date +%s%N)"
"$OGOD" tx send --from-key alice --to "$BOB" --amount 123456 --fee 100 \
  --home "$CHAIN_DIR/localnet-data/val1" --node tcp://127.0.0.1:26757
end_ns="$(date +%s%N)"
echo "broadcast took $(( (end_ns - start_ns) / 1000000 )) ms"
echo "Waiting 3s for the transfer to commit..."
sleep 3

say "Step 4: same balances from all four nodes"
for i in 0 1 2 3; do
  n="${NODES[$i]}"
  port="${RPC_PORTS[$i]}"
  echo "--- $n (rpc 127.0.0.1:$port) ---"
  echo -n "alice: "; "$OGOD" query balance "$ALICE" --node "tcp://127.0.0.1:$port" | jq -r .balance
  echo -n "bob:   "; "$OGOD" query balance "$BOB" --node "tcp://127.0.0.1:$port" | jq -r .balance
done
echo "Expected: alice 599876444 (600000000 - 123456 - 100 fee), bob 400123456."

say "Step 5: verify the same ledger state on all four nodes"
"$LOCALNET" verify

say "Step 6: voting power per validator"
"$LOCALNET" voting

say "Step 7: S5 fault tolerance (liveness half): val4 stops, blocks continue, val4 rejoins"
"$LOCALNET" s5

say "Step 8: block-time distribution over 60 seconds"
"$LOCALNET" blocktimes 60

say "Step 9: snapshot and restore val4 (O1 runbook steps, executed for real)"
SNAP_OUT="$("$LOCALNET" snapshot val4)"
echo "$SNAP_OUT"
SNAP_FILE="$(echo "$SNAP_OUT" | tail -n 1)"
echo "snapshot file: $SNAP_FILE"
"$LOCALNET" restore val4 "$SNAP_FILE"
echo "final verification after restore:"
"$LOCALNET" verify

say "Step 10: final status"
"$LOCALNET" status

say "PoC run complete"
echo "Demonstrated: 4-validator consensus on the same ledger, a wallet"
echo "transfer reconciled on all 4 nodes, epoch issuance executing through"
echo "the monetary hook, S5 liveness with 1 of 4 validators offline,"
echo "block-time distribution measured, and snapshot plus restore of a"
echo "validator executed for real."
echo "All values TEST-ONLY. No deployment, no token issuance, no spending."
echo "The network ran independently of the OGO application throughout."
