#!/usr/bin/env sh
# Preflight host port collisions before starting the quickstart stack.
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
LINETH_LOG_CONTEXT="check-ports"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"

section() { lineth_section "$*"; }

lineth_banner "port preflight · local services"

failures=0
seen_ports=" "

env_value() {
  key="$1"
  [ -f .env ] || return 1
  sed -nE "s/^${key}=([^#[:space:]].*)$/\1/p" .env | tail -1
}

with_default() {
  value="$1"
  fallback="$2"
  if [ -n "$value" ]; then
    printf '%s' "$value"
  else
    printf '%s' "$fallback"
  fi
}

port_owner() {
  port="$1"
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | awk 'NR == 2 { print $1 " pid=" $2; exit }'
    return 0
  fi
  if command -v nc >/dev/null 2>&1 && nc -z 127.0.0.1 "$port" >/dev/null 2>&1; then
    printf 'listener detected'
  fi
}

check_port() {
  name="$1"
  env_name="$2"
  default_port="$3"
  configured="$(env_value "$env_name" || true)"
  port="$(with_default "$configured" "$default_port")"

  case "$port" in
    ''|*[!0-9]*)
      lineth_error "$env_name must be a decimal port, got '$port'"
      failures=$((failures + 1))
      return
      ;;
  esac

  case "$seen_ports" in
    *" $port "*)
      lineth_error "duplicate quickstart host port $port at $env_name ($name)"
      failures=$((failures + 1))
      return
      ;;
  esac
  seen_ports="${seen_ports}${port} "

  owner="$(port_owner "$port" || true)"
  if [ -n "$owner" ]; then
    lineth_error "$env_name=$port ($name) -> $owner"
    failures=$((failures + 1))
  else
    lineth_ok "$env_name=$port ($name)"
  fi
}

section "checking expected host ports"
check_port "Sequencer RPC" HOST_PORT_L2_SEQUENCER_RPC 8645
check_port "Maru" HOST_PORT_MARU 8080
check_port "L2 RPC HTTP" HOST_PORT_L2_RPC 8745
check_port "L2 RPC WebSocket" HOST_PORT_L2_WS 8746
check_port "L2 engine API" HOST_PORT_L2_ENGINE 8748
check_port "Shomei" HOST_PORT_SHOMEI 8998
check_port "Web3signer" HOST_PORT_WEB3SIGNER 9000
check_port "Coordinator Postgres" HOST_PORT_COORDINATOR_PG 5432
check_port "Postman Postgres" HOST_PORT_POSTMAN_PG 5433
check_port "Blockscout Postgres" HOST_PORT_BLOCKSCOUT_L2_PG 5435
check_port "Coordinator observability" HOST_PORT_COORDINATOR 9545
check_port "Postman" HOST_PORT_POSTMAN 9090
check_port "Blockscout API" HOST_PORT_L2_BLOCKSCOUT 4000
check_port "Blockscout frontend" HOST_PORT_L2_BLOCKSCOUT_FRONTEND 4001

if [ "$failures" -gt 0 ]; then
  section "found $failures port issue(s)"
  lineth_info "Stop the conflicting service or override the matching HOST_PORT_* in .env."
  exit 1
fi

section "all checked quickstart host ports are free"
