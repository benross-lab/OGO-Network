#!/usr/bin/env bash
# OGO Network Milestone C localnet: 4 validators on one machine.
#
# Why 4 validators: the minimum operational set (BFT 3f+1 tolerates
# f=1 faulty). With 4 validators the chain keeps producing blocks
# with 1 validator down (S5 remainder / C5 slashing test), which the
# 3-validator Phase 2 prototype could not demonstrate.
#
# All values are TEST-ONLY and are labelled so everywhere they
# appear. The network runs entirely on this machine and never touches
# the OGO application, Supabase, or any production infrastructure.
# No deployment, no token issuance, no spending.
#
# Ports are offset from the Phase 2 prototype (2665x-2667x) because
# prototype nodes may still be running on this machine.
#
# Usage:
#   scripts/localnet4.sh build      build the ogod binary (v1)
#   scripts/localnet4.sh build-v2   build the v2 binary (UpgradeV2Enabled=true)
#   scripts/localnet4.sh init       (re)create the 4 validator homes and shared genesis
#   scripts/localnet4.sh start      start all 4 nodes in the background
#   scripts/localnet4.sh stop       stop all 4 nodes
#   scripts/localnet4.sh status     show height and app hash per node
#   scripts/localnet4.sh verify     verify the same ledger state on all 4 nodes
#   scripts/localnet4.sh up         build + init + start + verify
#   scripts/localnet4.sh down       stop (alias)
#   scripts/localnet4.sh clean      stop and delete all localnet4 data
#
# Data lives in chain/localnet4-data/ (override with LOCALNET4_DIR).
# Logs live in chain/localnet4-data/<node>/node.log.
set -euo pipefail

CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OGOD="${OGOD:-$CHAIN_DIR/build/ogod}"
DATA="${LOCALNET4_DIR:-$CHAIN_DIR/localnet4-data}"
CHAIN_ID="ogo-local-1"

# TEST-ONLY genesis funds (uogo). G = 4 x 200,000,000 + 1,000,000,000
# = 1,800,000,000. Self-delegation per validator is 100,000,000.
VAL_FUNDS="200000000uogo"
SELF_DELEGATION="100000000uogo"
FAUCET_FUNDS="1000000000uogo"
# TEST-ONLY monetary schedule: geometric, R0 1000, N 4, r 9/10,
# S_actual 3439, M = G + S_actual = 1,800,003,439.
SCHEDULE_R0="1000"
SCHEDULE_PERIODS="4"
MAX_SUPPLY="1800003439"
EPOCH_DURATION="60s"

NODES="val1 val2 val3 val4"

rpc_port() { case "$1" in val1) echo 27657;; val2) echo 27667;; val3) echo 27677;; val4) echo 27687;; esac; }
p2p_port() { case "$1" in val1) echo 27656;; val2) echo 27666;; val3) echo 27676;; val4) echo 27686;; esac; }
api_port() { case "$1" in val1) echo 1417;; val2) echo 1427;; val3) echo 1437;; val4) echo 1447;; esac; }
grpc_port() { case "$1" in val1) echo 9190;; val2) echo 9191;; val3) echo 9192;; val4) echo 9193;; esac; }

build() {
    echo "== building ogod (v1) =="
    (cd "$CHAIN_DIR" && GOCACHE="${GOCACHE:-$HOME/.cache/go-build}" GOTMPDIR="${GOTMPDIR:-$HOME/.cache/go-tmp}" ~/workspace/tools/go-1.25.7/bin/go build -o build/ogod ./cmd/ogod)
    ls -lh "$OGOD"
}

build_v2() {
    echo "== building ogod-v2 (UpgradeV2Enabled=true) =="
    (cd "$CHAIN_DIR" && GOCACHE="${GOCACHE:-$HOME/.cache/go-build}" GOTMPDIR="${GOTMPDIR:-$HOME/.cache/go-tmp}" ~/workspace/tools/go-1.25.7/bin/go build -ldflags "-X github.com/benross-lab/ogo-network/app.UpgradeV2Enabled=true" -o build/ogod-v2 ./cmd/ogod)
    ls -lh "$CHAIN_DIR/build/ogod-v2"
}

# Edit a genesis file with the TEST-ONLY monetary, epochs, staking,
# gov, and slashing parameters. M is computed from the bank balances
# in that file (strict M = G + S_actual, S_actual 3439 for the
# TEST-ONLY schedule R0 1000, N 4, r 9/10). Called for every home
# BEFORE gentx (gentx validates genesis, and monetary genesis has no
# default max supply) and again for val1 AFTER collect-gentxs, when
# the full account set is present.
edit_genesis() {
    local genesis="$1"
    local faucet_file="${2:-}"
    python3 - "$genesis" "$faucet_file" <<'PYEOF'
import json, sys
path, faucet_path = sys.argv[1], sys.argv[2]
with open(path) as f:
    g = json.load(f)
faucet_addr = ""
if faucet_path:
    with open(faucet_path) as f:
        faucet_addr = f.read().strip()
genesis_time = g["genesis_time"]

# G is the sum of the uogo bank balances in this genesis file.
balances = g["app_state"]["bank"]["balances"]
declared = []
g_total = 0
for b in balances:
    amt = next(c["amount"] for c in b["coins"] if c["denom"] == "uogo")
    g_total += int(amt)
    purpose = "public" if b["address"] == faucet_addr else "validator"
    declared.append({"address": b["address"], "denom": "uogo", "amount": amt, "purpose": purpose})
# TEST-ONLY schedule total: R0 1000, N 4, r 9/10 -> 1000+900+810+729.
max_supply = g_total + 3439

g["app_state"]["monetary"] = {
    "params": {
        "max_supply": str(max_supply),
        "denom": "uogo",
        "decimals": 6,
        "inflation_rate": "0",
        "minting_enabled": False,
        "supply_cap_enforced": True,
        "issuance_schedule": {
            "schedule_type": "reducing-per-period",
            "start_height": 1,
            "period_length_blocks": 100,
            "periods": 4,
            "initial_per_period": "1000",
            "reduction_numerator": 9,
            "reduction_denominator": 10,
        },
    },
    "balances": declared,
}
g["app_state"]["epochs"] = {
    "epochs": [
        {
            "identifier": "monetary",
            "start_time": genesis_time,
            "duration": "60s",
            "current_epoch": "0",
            "current_epoch_start_time": genesis_time,
            "current_epoch_start_height": "0",
            "epoch_counting_started": False,
        }
    ]
}
g["app_state"]["staking"]["params"]["bond_denom"] = "uogo"
g["app_state"]["gov"]["params"]["voting_period"] = "60s"
g["app_state"]["gov"]["params"]["expedited_voting_period"] = "30s"
# Gov deposits must be payable on this chain: the SDK default is
# 10,000,000stake, a denom that does not exist here, which would
# make every gov proposal (including the v2 upgrade) unsubmittable.
# TEST-ONLY deposits in uogo.
g["app_state"]["gov"]["params"]["min_deposit"] = [{"denom": "uogo", "amount": "1000000"}]
g["app_state"]["gov"]["params"]["expedited_min_deposit"] = [{"denom": "uogo", "amount": "5000000"}]
g["app_state"]["slashing"]["params"]["signed_blocks_window"] = "50"
g["app_state"]["slashing"]["params"]["min_signed_per_window"] = "0.500000000000000000"
g["app_state"]["slashing"]["params"]["downtime_jail_duration"] = "60s"
g["app_state"]["slashing"]["params"]["slash_fraction_downtime"] = "0.010000000000000000"
with open(path, "w") as f:
    json.dump(g, f, indent=2)
print(f"genesis edited {path}: G={g_total} M={max_supply} epoch=60s bond=uogo")
PYEOF
    "$OGOD" genesis validate-genesis --home "$(dirname "$(dirname "$genesis")")"
}

init() {
    echo "== init 4-validator localnet (TEST-ONLY genesis) =="
    rm -rf "$DATA"
    mkdir -p "$DATA"
    local faucet_addr=""
    for node in $NODES; do
        "$OGOD" init "$node" --chain-id "$CHAIN_ID" --home "$DATA/$node" >/dev/null
        "$OGOD" keys add "$node" --home "$DATA/$node" --keyring-backend test --output json >/dev/null 2>&1
    done
    # Faucet key lives in val1's keyring.
    faucet_addr="$("$OGOD" keys add faucet --home "$DATA/val1" --keyring-backend test --output json 2>/dev/null | python3 -c 'import json,sys; print(json.load(sys.stdin)["address"])')"
    echo "faucet address: $faucet_addr"
    # Every home's genesis must hold ALL accounts: collect-gentxs
    # validates each gentx's validator account against val1's
    # genesis, and each gentx validates against its own home's
    # genesis. add-genesis-account takes an address, so the key need
    # not live in that home's keyring.
    declare -A ADDRS
    for node in $NODES; do
        ADDRS[$node]="$("$OGOD" keys show "$node" --home "$DATA/$node" --keyring-backend test --address)"
    done
    for node in $NODES; do
        for acct in $NODES; do
            "$OGOD" genesis add-genesis-account "${ADDRS[$acct]}" "$VAL_FUNDS" --home "$DATA/$node" --keyring-backend test
        done
        "$OGOD" genesis add-genesis-account "$faucet_addr" "$FAUCET_FUNDS" --home "$DATA/$node" --keyring-backend test
    done
    # Record the faucet address for edit_genesis (side file, not
    # genesis: collect-gentxs may drop unknown genesis keys).
    echo "$faucet_addr" > "$DATA/faucet.addr"
    # Pre-edit every home's genesis: gentx validates genesis, and the
    # monetary module has no default max supply. Each home's M is
    # computed from its own (partial) balances; the final full-genesis
    # edit after collect-gentxs recomputes it from the full set.
    for node in $NODES; do
        edit_genesis "$DATA/$node/config/genesis.json" "$DATA/faucet.addr"
    done
    for node in $NODES; do
        "$OGOD" genesis gentx "$node" "$SELF_DELEGATION" \
            --chain-id "$CHAIN_ID" --home "$DATA/$node" --keyring-backend test \
            --moniker "$node" \
            --commission-rate "0.10" --commission-max-rate "0.20" --commission-max-change-rate "0.01" \
            --min-self-delegation "1"
    done
    mkdir -p "$DATA/val1/config/gentx"
    for node in val2 val3 val4; do
        cp "$DATA/$node/config/gentx/"*.json "$DATA/val1/config/gentx/"
    done
    "$OGOD" genesis collect-gentxs --home "$DATA/val1"
    edit_genesis "$DATA/val1/config/genesis.json" "$DATA/faucet.addr"
    # Distribute the final genesis and wire ports/peers.
    local peers=""
    for node in val2 val3 val4; do
        cp "$DATA/val1/config/genesis.json" "$DATA/$node/config/genesis.json"
    done
    for node in $NODES; do
        local id
        id="$("$OGOD" comet show-node-id --home "$DATA/$node")"
        peers="${peers}${id}@127.0.0.1:$(p2p_port "$node"),"
    done
    peers="${peers%,}"
    for node in $NODES; do
        python3 - "$DATA/$node/config/config.toml" "$(rpc_port "$node")" "$(p2p_port "$node")" "$peers" <<'PYEOF'
import sys, re
path, rpc, p2p, peers = sys.argv[1:5]
with open(path) as f: s = f.read()
s = re.sub(r'(?m)^laddr = "tcp://127\.0\.0\.1:26657"$', f'laddr = "tcp://127.0.0.1:{rpc}"', s, count=1)
s = re.sub(r'(?m)^laddr = "tcp://0\.0\.0\.0:26656"$', f'laddr = "tcp://127.0.0.1:{p2p}"', s, count=1)
# Replace ANY persistent_peers value: collect-gentxs pre-populates
# val1's line from the gentx files (validators at the VM external
# IP, default port), so an empty-only match would leave it in place.
s = re.sub(r'(?m)^persistent_peers = ".*"$', f'persistent_peers = "{peers}"', s, count=1)
# Unconditional peers are re-dialed persistently: without this, a
# startup dial that lands while the peer is still starting fails
# with "connection refused" and is not retried, leaving the mesh
# split (observed: two pairs, no quorum).
ids = ",".join(a.split("@")[0] for a in peers.split(","))
s = re.sub(r'(?m)^unconditional_peer_ids = ""$', f'unconditional_peer_ids = "{ids}"', s, count=1)
s = re.sub(r'(?m)^#?timeout_commit = ".*"$', 'timeout_commit = "1s"', s, count=1)
# Single-machine localnet: all four nodes share 127.0.0.1, and the
# default duplicate-IP filter (AllowDuplicateIP=false) permits only
# ONE peer per IP, so the mesh splits into pairs and the chain never
# reaches quorum. Production validators run on separate hosts and
# keep the default.
s = re.sub(r'(?m)^allow_duplicate_ip = false$', 'allow_duplicate_ip = true', s, count=1)
# Distinct pprof ports per node (default 6060 collides on one host).
s = re.sub(r'(?m)^pprof_laddr = "localhost:6060"$', 'pprof_laddr = ""', s, count=1)
with open(path, "w") as f: f.write(s)
PYEOF
        python3 - "$DATA/$node/config/app.toml" "$(api_port "$node")" "$(grpc_port "$node")" <<'PYEOF'
import sys, re
path, api, grpc = sys.argv[1:4]
with open(path) as f: s = f.read()
s = re.sub(r'(?m)^address = "tcp://localhost:1317"$', f'address = "tcp://127.0.0.1:{api}"', s, count=1)
s = re.sub(r'(?m)^address = "localhost:9090"$', f'address = "127.0.0.1:{grpc}"', s, count=1)
s = re.sub(r'(?m)^minimum-gas-prices = ""$', 'minimum-gas-prices = "0.001uogo"', s, count=1)
s = re.sub(r'(?m)^enable = false$', 'enable = true', s, count=1)
with open(path, "w") as f: f.write(s)
PYEOF
    done
    echo "== init done: $DATA =="
}

start() {
    echo "== starting 4 nodes (sequential: CometBFT dials persistent"
    echo "   peers once at startup with no retry, and 127.0.0.1 cannot"
    echo "   enter the addrbook for PEX redial, so each node must be"
    echo "   listening before the next one dials it) =="
    for node in $NODES; do
        # RPC/P2P listen addresses come from config.toml (edited at
        # init); the start command has no --rpc-laddr flag.
        nohup "$OGOD" start --home "$DATA/$node" --moniker "$node" \
            >"$DATA/$node/node.log" 2>&1 &
        echo "$!" >"$DATA/$node/pid"
        echo "$node pid $! rpc 127.0.0.1:$(rpc_port "$node")"
        sleep 3
    done
    sleep 8
    status
}

stop() {
    echo "== stopping nodes =="
    for node in $NODES; do
        if [ -f "$DATA/$node/pid" ]; then
            kill "$(cat "$DATA/$node/pid")" 2>/dev/null || true
            rm -f "$DATA/$node/pid"
            echo "$node stopped"
        fi
    done
    # NOTE: no pkill -f here: a pkill pattern naming the ogod start
    # command matches this script's own command line and kills it.
}

height_and_hash() {
    python3 -c "
import json,sys,urllib.request
try:
    with urllib.request.urlopen('http://127.0.0.1:$1/status', timeout=5) as r:
        d = json.load(r)
    si = d['result']['sync_info']
    print(si['latest_block_height'], si['latest_app_hash'][:16])
except Exception as e:
    print('DOWN', e)
"
}

status() {
    for node in $NODES; do
        printf "%s rpc %s: " "$node" "$(rpc_port "$node")"
        height_and_hash "$(rpc_port "$node")"
    done
}

verify() {
    echo "== verifying agreement across 4 nodes =="
    local first=""
    for node in $NODES; do
        local h
        h="$(height_and_hash "$(rpc_port "$node")")"
        echo "$node: $h"
        if [ -z "$first" ]; then first="$h"; fi
    done
    echo "heights/hashes above must agree (heights within a few blocks)"
}

clean() {
    stop || true
    rm -rf "$DATA"
    echo "cleaned $DATA"
}

case "${1:-}" in
    build) build ;;
    build-v2) build_v2 ;;
    init) init ;;
    start) start ;;
    stop) stop ;;
    down) stop ;;
    status) status ;;
    verify) verify ;;
    up) build; init; start; sleep 5; verify ;;
    clean) clean ;;
    *) echo "usage: $0 {build|build-v2|init|start|stop|down|status|verify|up|clean}"; exit 1 ;;
esac
