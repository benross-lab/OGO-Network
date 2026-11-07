#!/usr/bin/env bash
# Local-only OGO Network development helper.
# It never touches Cloudflare, DNS, production routes, or the preserved PoC.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHAIN="$ROOT/milestone-c/chain"
GO_BIN="${GO_BIN:-/home/ubuntu/work/ogo-network-dev/tools/go/bin/go}"
DATA="${LOCALNET4_DIR:-/home/ubuntu/work/ogo-network-dev/localnet4-dev}"
SNAPSHOTS="${OGO_SNAPSHOT_DIR:-/home/ubuntu/work/ogo-network-dev/localnet4-snapshots}"
API_PORT="${OGO_API_PORT:-18787}"
RPC_PORT="${OGO_RPC_PORT:-27657}"
API_PID_FILE="$DATA/api.pid"

localnet() {
  LOCALNET4_DIR="$DATA" OGOD="$CHAIN/build/ogod" bash "$CHAIN/scripts/localnet4.sh" "$@"
}

build_chain() {
  mkdir -p "$CHAIN/build"
  (cd "$CHAIN" && "$GO_BIN" build -o "$CHAIN/build/ogod" ./cmd/ogod)
}

api_start() {
  if [ -f "$API_PID_FILE" ] && kill -0 "$(cat "$API_PID_FILE")" 2>/dev/null; then
    echo "API already running (pid $(cat "$API_PID_FILE"))"
    return
  fi
  mkdir -p "$DATA"
  (cd "$ROOT/network-api" && \
    OGO_API_RPC_URL="http://127.0.0.1:$RPC_PORT" \
    OGO_API_LCD_URL="http://127.0.0.1:1417" \
    OGO_API_LISTEN="127.0.0.1:$API_PORT" \
    OGO_API_ALLOW_BROADCAST="${OGO_API_ALLOW_BROADCAST:-false}" \
    "$GO_BIN" run . >"$DATA/api.log" 2>&1 & echo $! >"$API_PID_FILE")
  sleep 2
  curl -fsS "http://127.0.0.1:$API_PORT/healthz"
  echo
}

api_stop() {
  if [ -f "$API_PID_FILE" ]; then
    kill "$(cat "$API_PID_FILE")" 2>/dev/null || true
    rm -f "$API_PID_FILE"
  fi
}

snapshot() {
  mkdir -p "$SNAPSHOTS"
  localnet stop || true
  api_stop
  local stamp
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  tar -C "$(dirname "$DATA")" -czf "$SNAPSHOTS/localnet4-$stamp.tar.gz" "$(basename "$DATA")"
  echo "$SNAPSHOTS/localnet4-$stamp.tar.gz"
}

restore() {
  local archive="${1:?usage: $0 restore SNAPSHOT_TAR_GZ}"
  [ -f "$archive" ] || { echo "snapshot not found: $archive" >&2; exit 1; }
  localnet stop || true
  api_stop
  rm -rf "$DATA"
  tar -C "$(dirname "$DATA")" -xzf "$archive"
  localnet start
  localnet verify
}

case "${1:-}" in
  init) localnet init ;;
  start) localnet start; api_start ;;
  stop) api_stop; localnet stop ;;
  status) localnet status; curl -fsS "http://127.0.0.1:$API_PORT/v1/network" || true; echo ;;
  verify) localnet verify; curl -fsS "http://127.0.0.1:$API_PORT/v1/validators"; echo ;;
  snapshot) snapshot ;;
  restore) restore "$2" ;;
  test)
    build_chain
    localnet init
    localnet start
    api_start
    localnet verify
    curl -fsS "http://127.0.0.1:$API_PORT/v1/network" >/dev/null
    curl -fsS "http://127.0.0.1:$API_PORT/v1/validators" >/dev/null
    echo "local development smoke test passed"
    ;;
  clean) api_stop; localnet clean ;;
  *) echo "usage: $0 {init|start|stop|status|verify|snapshot|restore FILE|test|clean}"; exit 1 ;;
esac
