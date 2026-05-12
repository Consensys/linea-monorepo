#!/usr/bin/env sh
# Run continuous L2 ERC20Example transfer traffic for Blockscout/prover demos.
set -eu

section() { printf '\n[l2-erc20-traffic] %s\n' "$*"; }
log() { printf '[l2-erc20-traffic] %s\n' "$*"; }
die() { printf '[l2-erc20-traffic] ERROR: %s\n' "$*" >&2; exit 1; }

env_value() {
  key="$1"
  [ -f .env ] || return 1
  sed -nE "s/^${key}=([^#[:space:]].*)$/\1/p" .env | tail -1
}

with_default() {
  value="$1"
  fallback="$2"
  if [ -n "$value" ]; then printf '%s' "$value"; else printf '%s' "$fallback"; fi
}

container_exists() {
  docker inspect "$CONTAINER_NAME" >/dev/null 2>&1
}

container_running() {
  [ "$(docker inspect -f '{{.State.Running}}' "$CONTAINER_NAME" 2>/dev/null || true)" = "true" ]
}

command="${1:-start}"
CONTAINER_NAME="${TRAFFIC_CONTAINER_NAME:-linea-l2-erc20-traffic}"

if ! docker info >/dev/null 2>&1; then
  die "Docker daemon is not reachable"
fi

case "$command" in
  start|stop|status|logs) ;;
  *)
    die "usage: $0 [start|stop|status|logs]"
    ;;
esac

if [ "$command" = "status" ]; then
  if container_running; then
    docker ps --filter "name=^/${CONTAINER_NAME}$" --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}'
  elif container_exists; then
    log "$CONTAINER_NAME exists but is not running"
  else
    log "$CONTAINER_NAME is not running"
  fi
  exit 0
fi

if [ "$command" = "logs" ]; then
  if ! container_exists; then
    die "$CONTAINER_NAME does not exist"
  fi
  docker logs -f --tail "${TAIL:-80}" "$CONTAINER_NAME"
  exit 0
fi

if [ "$command" = "stop" ]; then
  if container_exists; then
    docker rm -f "$CONTAINER_NAME" >/dev/null
    log "stopped $CONTAINER_NAME"
  else
    log "$CONTAINER_NAME was not running"
  fi
  exit 0
fi

if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  die "linea-stack-shared-config volume not found. Boot the stack first."
fi

case "${AMOUNT_WEI:-1}" in
  ''|*[!0-9]*) die "AMOUNT_WEI must be a positive integer" ;;
esac
[ "${AMOUNT_WEI:-1}" -gt 0 ] || die "AMOUNT_WEI must be greater than zero"

case "${INTERVAL_SECONDS:-2}" in
  ''|*[!0-9]*) die "INTERVAL_SECONDS must be a positive integer" ;;
esac
[ "${INTERVAL_SECONDS:-2}" -gt 0 ] || die "INTERVAL_SECONDS must be greater than zero"

case "${MAX_TXS:-0}" in
  ''|*[!0-9]*) die "MAX_TXS must be a non-negative integer" ;;
esac

if container_running; then
  log "$CONTAINER_NAME is already running"
  log "tail it with: $0 logs"
  log "stop it with: $0 stop"
  exit 0
fi

if container_exists; then
  docker rm "$CONTAINER_NAME" >/dev/null
fi

if [ -f versions.env ]; then
  # shellcheck disable=SC1091
  . ./versions.env
fi

FOUNDRY_IMAGE="${FOUNDRY_IMAGE:-ghcr.io/foundry-rs/foundry:${FOUNDRY_TAG:-latest}}"
L2_RPC_URL="${L2_RPC_URL:-http://sequencer:8545}"
HOST_PORT_L2_BLOCKSCOUT_FRONTEND="$(with_default "${HOST_PORT_L2_BLOCKSCOUT_FRONTEND:-$(env_value HOST_PORT_L2_BLOCKSCOUT_FRONTEND || true)}" 4001)"
BLOCKSCOUT_BASE_URL="${BLOCKSCOUT_BASE_URL:-http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND}"

section "starting continuous ERC20Example traffic"
docker run -d \
  --name "$CONTAINER_NAME" \
  --user 0:0 \
  --entrypoint sh \
  --network linea-stack_linea \
  -v linea-stack-shared-config:/shared:ro \
  -e AMOUNT_WEI="${AMOUNT_WEI:-1}" \
  -e INTERVAL_SECONDS="${INTERVAL_SECONDS:-2}" \
  -e MAX_TXS="${MAX_TXS:-0}" \
  -e TO="${TO:-}" \
  -e L2_RPC_URL="$L2_RPC_URL" \
  -e BLOCKSCOUT_BASE_URL="$BLOCKSCOUT_BASE_URL" \
  "$FOUNDRY_IMAGE" \
  -lc '
    set -eu

    [ -f /shared/runtime-keys.env ] || { echo "[l2-erc20-traffic] ERROR: /shared/runtime-keys.env missing" >&2; exit 1; }
    [ -f /shared/addresses-precomputed.json ] || { echo "[l2-erc20-traffic] ERROR: /shared/addresses-precomputed.json missing" >&2; exit 1; }
    [ -f /shared/addresses.json ] || { echo "[l2-erc20-traffic] ERROR: /shared/addresses.json missing; deploy-contracts has not completed" >&2; exit 1; }

    . /shared/runtime-keys.env
    : "${L2_DEPLOYER_PRIVATE_KEY:?L2_DEPLOYER_PRIVATE_KEY missing from runtime-keys.env}"

    erc20=$(sed -nE "/\"l2\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"ERC20Example\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" /shared/addresses.json | head -1)
    echo "$erc20" | grep -qE "^0x[a-fA-F0-9]{40}$" || { echo "[l2-erc20-traffic] ERROR: L2 ERC20Example missing from /shared/addresses.json" >&2; exit 1; }

    if [ -n "${TO:-}" ]; then
      recipient="$TO"
    else
      recipient=$(sed -nE "s/.*\"l2PostmanAddress\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" /shared/addresses-precomputed.json | head -1)
    fi
    echo "$recipient" | grep -qE "^0x[a-fA-F0-9]{40}$" || { echo "[l2-erc20-traffic] ERROR: recipient address invalid: $recipient" >&2; exit 1; }

    sender=$(cast wallet address --private-key "$L2_DEPLOYER_PRIVATE_KEY")
    echo "[l2-erc20-traffic] start token=$erc20 from=$sender to=$recipient amountWei=$AMOUNT_WEI interval=${INTERVAL_SECONDS}s maxTxs=$MAX_TXS"
    echo "[l2-erc20-traffic] blockscout token=$BLOCKSCOUT_BASE_URL/token/$erc20"

    i=1
    while [ "$MAX_TXS" -eq 0 ] || [ "$i" -le "$MAX_TXS" ]; do
      receipt=$(cast send "$erc20" "transfer(address,uint256)" "$recipient" "$AMOUNT_WEI" \
        --private-key "$L2_DEPLOYER_PRIVATE_KEY" \
        --rpc-url "$L2_RPC_URL" \
        --json)

      tx_hash=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"transactionHash\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
      block_number=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"blockNumber\"[[:space:]]*:[[:space:]]*\"?([^\",}]+)\"?.*/\1/p" | head -1)
      [ -n "$tx_hash" ] || { echo "[l2-erc20-traffic] ERROR: cast receipt did not include transactionHash" >&2; printf "%s\n" "$receipt" >&2; exit 1; }
      [ -n "$block_number" ] || block_number="unknown"

      echo "[l2-erc20-traffic] $i tx=$tx_hash block=$block_number"
      echo "[l2-erc20-traffic] blockscout=$BLOCKSCOUT_BASE_URL/tx/$tx_hash"
      i=$((i + 1))
      sleep "$INTERVAL_SECONDS"
    done

    echo "[l2-erc20-traffic] completed maxTxs=$MAX_TXS"
  '

log "started $CONTAINER_NAME"
log "Blockscout UI: $BLOCKSCOUT_BASE_URL"
log "tail it with: $0 logs"
log "stop it with: $0 stop"
