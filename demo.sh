#!/usr/bin/env bash
# OGO Network prototype demo (Milestones 1-2).
#
# What this demonstrates, end to end, on one machine:
#   1. Three local validators start from one shared genesis and
#      produce blocks together (consensus on the same ledger).
#   2. Genesis accounts: alice and bob hold TEST-ONLY OGO Coin.
#   3. A transfer of TEST-ONLY OGO Coin between wallets, with the
#      fee and the elapsed broadcast time measured and printed.
#   4. Epoch issuance executing: the monetary AfterEpochEnd hook
#      mints the schedule amount into the rewards pool at each epoch
#      boundary, and the demo shows it happening live.
#   5. The same balances verified from all three nodes.
#   6. Halt-and-resume: one validator stops, the chain halts
#      (expected with 3 validators, which tolerate zero faults),
#      the validator restarts, blocks resume, nodes agree.
#   7. The batching demo, if its binary is present (another agent
#      builds it; this script only calls it).
#
# What this does NOT do: no deployment, no token issuance, no
# production change, no spending. All values are TEST-ONLY. The
# network runs entirely on this machine and never touches the OGO
# application: the application is a client of the network, not its
# authority. Nothing here reads from or writes to Supabase, the OGO
# Worker, or benross-lab/OGO.
#
# Validator count note: THREE validators is the decided count for
# the local prototype. It is distinct from the minimum operational
# set of FOUR validators (the smallest set that tolerates one faulty
# validator and can demonstrate liveness with one down) and from the
# decentralization target of FIFTY or more validators for a public
# network. Keplr and live IBC are not prerequisites for this
# prototype, and the monetary prototype is independent of smart
# contracts; CosmWasm is evaluated separately before adoption.
#
# Usage: ./demo.sh
# The script stops the nodes when it finishes (or on failure),
# so it is safe to re-run.
set -euo pipefail

PROTO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$PROTO_DIR/chain"
LOCALNET="$CHAIN_DIR/scripts/localnet.sh"
OGOD="$CHAIN_DIR/ogod"
export TMPDIR="${TMPDIR:-$HOME/workspace/.go-tmp}"
export GOCACHE="${GOCACHE:-$HOME/workspace/.go-build-cache}"

NODES=(val1 val2 val3)
RPC_PORTS=(26657 26667 26677)

say() { echo; echo "=== $1 ==="; }

cleanup() {
  echo
  echo "stopping the localnet..."
  "$LOCALNET" stop >/dev/null 2>&1 || true
}
trap cleanup EXIT

say "OGO Network prototype demo (Milestones 1-2), all values TEST-ONLY"
echo "Three validators, one shared genesis, epoch mode on (20s epochs)."
echo "Chain: ogo-local-1. Denom: uogo (proposed, 6 decimals)."

say "Step 1: build and start 3 validators"
"$LOCALNET" up

ALICE="$(jq -r .address "$CHAIN_DIR/localnet-data/val1/keys/alice.json")"
BOB="$(jq -r .address "$CHAIN_DIR/localnet-data/val1/keys/bob.json")"
echo "alice: $ALICE"
echo "bob:   $BOB"

say "Step 2: wait for the first epoch boundary (epoch issuance executing)"
echo "Epoch duration is 20s (TEST-ONLY). Polling until the hook mints..."
deadline=$((SECONDS + 90))
while [ "$SECONDS" -lt "$deadline" ]; do
  issued="$("$OGOD" query epochs --node tcp://127.0.0.1:26657 | jq -r .total_issued)"
  if [ -n "$issued" ] && [ "$issued" != "0" ]; then break; fi
  sleep 2
done
echo "Epoch state on val1:"
"$OGOD" query epochs --node tcp://127.0.0.1:26657
echo "Rewards pool balance (the mint recipient):"
"$OGOD" query balance ogo1rewardspool --node tcp://127.0.0.1:26657
echo "Supply:"
"$OGOD" query supply --node tcp://127.0.0.1:26657

say "Step 3: transfer 123456 uogo from alice to bob (fee 100 uogo, TEST-ONLY)"
echo "Measuring broadcast time (fee is flat per transaction in this scaffold; there is no gas metering yet)."
start_ns="$(date +%s%N)"
"$OGOD" tx send --from-key alice --to "$BOB" --amount 123456 --fee 100 \
  --home "$CHAIN_DIR/localnet-data/val1" --node tcp://127.0.0.1:26657
end_ns="$(date +%s%N)"
echo "broadcast took $(( (end_ns - start_ns) / 1000000 )) ms"
echo "Waiting 3s for the transfer to commit..."
sleep 3

say "Step 4: same balances from all three nodes"
for i in 0 1 2; do
  n="${NODES[$i]}"
  port="${RPC_PORTS[$i]}"
  echo "--- $n (rpc 127.0.0.1:$port) ---"
  echo -n "alice: "; "$OGOD" query balance "$ALICE" --node "tcp://127.0.0.1:$port" | jq -r .balance
  echo -n "bob:   "; "$OGOD" query balance "$BOB" --node "tcp://127.0.0.1:$port" | jq -r .balance
done
echo "Expected: alice 599876444 (600000000 - 123456 - 100 fee), bob 400123456."

say "Step 5: verify the same ledger state on all three nodes"
"$LOCALNET" verify

say "Step 6: halt-and-resume (one validator stops, chain halts, restarts, resumes)"
"$LOCALNET" liveness

say "Step 7: agent batching demo (separate agent builds it; this script only calls it)"
if [ -f "$CHAIN_DIR/batchdemo-bin" ] && [ -x "$CHAIN_DIR/batchdemo-bin" ]; then
  "$CHAIN_DIR/batchdemo-bin" || echo "NOTE: batchdemo exited non-zero; see its output above."
else
  echo "batching demo binary not present yet (chain/batchdemo-bin). Build it with:"
  echo "  cd chain && go build -o batchdemo-bin ./batchdemo"
  echo "Skipping for now."
fi

say "Demo complete"
echo "Demonstrated: 3-validator consensus on the same ledger, a wallet"
echo "transfer reconciled on all 3 nodes, epoch issuance executing through"
echo "the monetary hook, and halt-and-resume after a validator stop."
echo "All values TEST-ONLY. No deployment, no token issuance, no spending."
echo "The network ran independently of the OGO application throughout."
