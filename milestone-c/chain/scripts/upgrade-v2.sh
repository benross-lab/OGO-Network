#!/usr/bin/env bash
# OGO Network Milestone C: genuine Cosmovisor upgrade to v2 (M2).
#
# What v2 does (non-monetary): adds one new state field, the upgrade
# marker in the new `monetary-meta` store, and records it in the
# monetary store. It changes no monetary parameter, schedule, or
# balance. The v2 binary is the same source built with
# -ldflags "-X .../app.UpgradeV2Enabled=true"; only the v2 binary
# mounts the monetary-meta store (store loader at the upgrade height).
#
# All values are TEST-ONLY. Local staging only.
#
# Usage:
#   scripts/upgrade-v2.sh setup      stop nodes, lay out Cosmovisor dirs,
#                                    restart all 4 nodes under Cosmovisor
#   scripts/upgrade-v2.sh propose    submit the v2 software-upgrade gov
#                                    proposal (height = current + 40) and
#                                    vote yes from all 4 validators
#   scripts/upgrade-v2.sh watch      wait for the upgrade height, then
#                                    verify post-upgrade state
#   scripts/upgrade-v2.sh run        setup + propose + watch (full run)
#   scripts/upgrade-v2.sh prestate   record pre-upgrade state snapshot
#   scripts/upgrade-v2.sh verify     verify post-upgrade state vs snapshot
#
# State snapshots live in chain/upgrade-v2/ (pre.json, post.json).
set -euo pipefail

CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OGOD="$CHAIN_DIR/build/ogod"
OGOD_V2="$CHAIN_DIR/build/ogod-v2"
COSMOVISOR="${COSMOVISOR:-$CHAIN_DIR/build/cosmovisor}"
DATA="$CHAIN_DIR/localnet4-data"
CHAIN_ID="ogo-local-1"
UPGRADE_DIR="$CHAIN_DIR/upgrade-v2"
NODES="val1 val2 val3 val4"
MONETARY_ACCT="ogo158r4scgvqzrgtc8cz6f076x37meg5thc2zzfz4"

rpc_port() { case "$1" in val1) echo 27657;; val2) echo 27667;; val3) echo 27677;; val4) echo 27687;; esac; }

height() {
    curl -s --max-time 5 "http://127.0.0.1:$(rpc_port val1)/status" | python3 -c "import json,sys; print(json.load(sys.stdin)['result']['sync_info']['latest_block_height'])"
}

apphash() {
    curl -s --max-time 5 "http://127.0.0.1:$(rpc_port val1)/status" | python3 -c "import json,sys; print(json.load(sys.stdin)['result']['sync_info']['latest_app_hash'])"
}

# Query a balance via the CLI (node flag goes after the subcommand).
balance() {
    "$OGOD" query bank balances "$1" --node "tcp://127.0.0.1:27657" --output json 2>/dev/null | python3 -c "import json,sys; b=json.load(sys.stdin)['balances']; print(b[0]['amount'] if b else '0')"
}

prestate() {
    mkdir -p "$UPGRADE_DIR"
    local h ah supply faucet
    h="$(height)"; ah="$(apphash)"
    supply="$("$OGOD" query bank total-supply --node tcp://127.0.0.1:27657 --output json 2>/dev/null | python3 -c "import json,sys; print(json.load(sys.stdin)['supply'][0]['amount'])")"
    faucet="$(balance "$(cat "$DATA/faucet.addr")")"
    python3 - "$UPGRADE_DIR/pre.json" "$h" "$ah" "$supply" "$faucet" <<'PYEOF'
import json, sys
path, h, ah, supply, faucet = sys.argv[1:6]
addrs = {}
with open(path, "w") as f:
    json.dump({"height": int(h), "app_hash": ah, "total_supply_uogo": supply,
               "faucet_uogo": faucet}, f, indent=2)
print(f"pre-upgrade state recorded: height={h} app_hash={ah[:16]} supply={supply} faucet={faucet}")
PYEOF
    for node in $NODES; do
        local addr bal
        addr="$("$OGOD" keys show "$node" --home "$DATA/$node" --keyring-backend test --address)"
        bal="$(balance "$addr")"
        echo "$node $addr liquid=$bal"
        python3 - "$UPGRADE_DIR/pre.json" "$node" "$bal" <<'PYEOF'
import json, sys
path, node, bal = sys.argv[1:4]
with open(path) as f: d = json.load(f)
d[f"{node}_liquid_uogo"] = bal
with open(path, "w") as f: json.dump(d, f, indent=2)
PYEOF
    done
    echo "monetary module account (rewards pool): $(balance "$MONETARY_ACCT") uogo"
    python3 - "$UPGRADE_DIR/pre.json" "$(balance "$MONETARY_ACCT")" <<'PYEOF'
import json, sys
path, bal = sys.argv[1:3]
with open(path) as f: d = json.load(f)
d["monetary_pool_uogo"] = bal
with open(path, "w") as f: json.dump(d, f, indent=2)
PYEOF
}

setup() {
    echo "== Cosmovisor setup: v1 under genesis, v2 staged =="
    if [ ! -x "$COSMOVISOR" ]; then echo "cosmovisor not found at $COSMOVISOR" >&2; exit 1; fi
    if [ ! -x "$OGOD_V2" ]; then echo "v2 binary not found at $OGOD_V2 (run: scripts/localnet4.sh build-v2)" >&2; exit 1; fi
    "$CHAIN_DIR/scripts/localnet4.sh" stop || true
    sleep 2
    for node in $NODES; do
        local home="$DATA/$node"
        mkdir -p "$home/cosmovisor/genesis/bin" "$home/cosmovisor/upgrades/v2/bin"
        cp "$OGOD" "$home/cosmovisor/genesis/bin/ogod"
        cp "$OGOD_V2" "$home/cosmovisor/upgrades/v2/bin/ogod"
        echo "$node: cosmovisor dirs ready"
    done
    echo "== restarting all 4 nodes under Cosmovisor =="
    for node in $NODES; do
        DAEMON_NAME=ogod DAEMON_HOME="$DATA/$node" DAEMON_ALLOW_DOWNLOAD_BINARIES=false DAEMON_RESTART_AFTER_UPGRADE=true \
            nohup "$COSMOVISOR" run start --home "$DATA/$node" --moniker "$node" \
            >"$DATA/$node/node.log" 2>&1 &
        echo "$!" >"$DATA/$node/pid"
        echo "$node cosmovisor pid $!"
        sleep 3
    done
    sleep 10
    "$CHAIN_DIR/scripts/localnet4.sh" status
}

propose() {
    local cur target
    cur="$(height)"
    target=$((cur + 40))
    echo "current height $cur; upgrade height $target"
    prestate
    local gov_acct faucet
    gov_acct="ogo10d07y265gmmuvt4z0w9aw880jnsr700jvdnngz"
    faucet="$(cat "$DATA/faucet.addr")"
    cat > "$UPGRADE_DIR/proposal.json" <<EOF
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "$gov_acct",
      "plan": {
        "name": "v2",
        "height": "$target",
        "info": "Milestone C v2: non-monetary store migration adding the monetary-meta upgrade marker (TEST-ONLY local staging)"
      }
    }
  ],
  "metadata": "ipfs://test-only-milestone-c-v2",
  "deposit": "1000000uogo",
  "title": "Milestone C v2 upgrade (TEST-ONLY)",
  "summary": "Non-monetary upgrade: adds the monetary-meta store with the v2 upgrade marker. No monetary parameter, schedule, or balance changes."
}
EOF
    echo "== submitting software-upgrade proposal (deposit 1,000,000uogo from faucet) =="
    "$OGOD" tx gov submit-proposal "$UPGRADE_DIR/proposal.json" \
        --from faucet --chain-id "$CHAIN_ID" --home "$DATA/val1" --keyring-backend test \
        --node tcp://127.0.0.1:27657 --gas 400000 --fees 1000uogo --yes --output json | tail -n 3
    sleep 4
    local pid
    pid="$("$OGOD" query gov proposals --node tcp://127.0.0.1:27657 --output json | python3 -c "import json,sys; ps=json.load(sys.stdin)['proposals']; print(ps[-1]['id'])")"
    echo "proposal id: $pid"
    echo "$pid" > "$UPGRADE_DIR/proposal.id"
    echo "$target" > "$UPGRADE_DIR/upgrade.height"
    echo "== voting yes from all 4 validators (60s voting period) =="
    for node in $NODES; do
        "$OGOD" tx gov vote "$pid" yes --from "$node" --chain-id "$CHAIN_ID" \
            --home "$DATA/$node" --keyring-backend test --node tcp://127.0.0.1:27657 \
            --gas 200000 --fees 1000uogo --yes --output json 2>&1 | tail -n 1
    done
    sleep 65
    "$OGOD" query gov proposal "$pid" --node tcp://127.0.0.1:27657 --output json | python3 -c "import json,sys; p=json.load(sys.stdin)['proposal']; print('proposal status:', p['status'])"
    "$OGOD" query upgrade plan --node tcp://127.0.0.1:27657 --output json 2>/dev/null || echo "(upgrade plan query returned nothing yet)"
}

# Read the upgrade marker from a store via the ABCI store query path.
# Usage: store_marker <store-key> -> value or empty.
store_marker() {
    local store="$1"
    local key_hex
    key_hex="$(python3 -c "print('upgrade_marker'.encode().hex())")"
    curl -s --max-time 5 "http://127.0.0.1:27657/abci_query?path=%22/store/$store/key%22&data=0x$key_hex" | python3 -c "
import json,sys,base64
d = json.load(sys.stdin)
try:
    v = d['result']['response']['value']
    print(base64.b64decode(v).decode() if v else '(empty)')
except Exception:
    print('(no value)')
"
}

verify() {
    echo "== post-upgrade verification =="
    local h ah supply faucet pool
    h="$(height)"; ah="$(apphash)"
    supply="$("$OGOD" query bank total-supply --node tcp://127.0.0.1:27657 --output json 2>/dev/null | python3 -c "import json,sys; print(json.load(sys.stdin)['supply'][0]['amount'])")"
    faucet="$(balance "$(cat "$DATA/faucet.addr")")"
    pool="$(balance "$MONETARY_ACCT")"
    python3 - "$UPGRADE_DIR/post.json" "$h" "$ah" "$supply" "$faucet" "$pool" <<'PYEOF'
import json, sys
path, h, ah, supply, faucet, pool = sys.argv[1:7]
with open(path, "w") as f:
    json.dump({"height": int(h), "app_hash": ah, "total_supply_uogo": supply,
               "faucet_uogo": faucet, "monetary_pool_uogo": pool}, f, indent=2)
print(f"post-upgrade state: height={h} app_hash={ah[:16]} supply={supply} faucet={faucet} pool={pool}")
PYEOF
    python3 - "$UPGRADE_DIR/pre.json" "$UPGRADE_DIR/post.json" <<'PYEOF'
import json, sys
pre = json.load(open(sys.argv[1])); post = json.load(open(sys.argv[2]))
for k in ("total_supply_uogo", "faucet_uogo", "monetary_pool_uogo"):
    same = "IDENTICAL" if pre[k] == post[k] else f"CHANGED {pre[k]} -> {post[k]}"
    print(f"{k}: pre={pre[k]} post={post[k]} {same}")
PYEOF
    echo "monetary store upgrade_marker: $(store_marker monetary)"
    echo "monetary-meta store upgrade_marker: $(store_marker monetary-meta)"
    echo "v2 binary running? (exe of each node process)"
    for node in $NODES; do
        local pid exe
        pid="$(cat "$DATA/$node/pid" 2>/dev/null || echo '?')"
        exe="$(readlink "/proc/$pid/exe" 2>/dev/null || echo '?')"
        echo "$node pid=$pid exe=$exe"
    done
    "$CHAIN_DIR/scripts/localnet4.sh" status
}

watch() {
    local target
    target="$(cat "$UPGRADE_DIR/upgrade.height")"
    local t0 t1
    t0="$(date +%s)"
    echo "waiting for upgrade height $target (current $(height))..."
    while [ "$(height)" -lt "$target" ]; do sleep 2; done
    echo "upgrade height $target reached at $(date -u +%H:%M:%S); waiting for the chain to resume..."
    sleep 15
    # The chain halts at the upgrade height while Cosmovisor swaps
    # binaries; resume is visible as blocks past the target height.
    local deadline=$((SECONDS + 300))
    while [ "$(height)" -le "$target" ] && [ "$SECONDS" -lt "$deadline" ]; do sleep 2; done
    t1="$(date +%s)"
    echo "chain resumed at height $(height) after $((t1 - t0))s from watch start (upgrade downtime measured from the last pre-upgrade block separately in the report)"
    verify
}

case "${1:-}" in
    setup) setup ;;
    propose) propose ;;
    watch) watch ;;
    prestate) prestate ;;
    verify) verify ;;
    run) setup; propose; watch ;;
    *) echo "usage: $0 {setup|propose|watch|prestate|verify|run}"; exit 1 ;;
esac
