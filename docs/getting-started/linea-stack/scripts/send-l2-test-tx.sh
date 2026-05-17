#!/usr/bin/env sh
# Send one or more tiny L2 ETH transfers to make the local L2 and Blockscout move.
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
LINETH_LOG_CONTEXT="l2-test-tx"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"

section() { lineth_section "$*"; }
die() { lineth_die "$*"; }

lineth_banner "L2 ETH transfer · local sequencer"

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

case "${COUNT:-1}" in
  ''|*[!0-9]*) die "COUNT must be a positive integer" ;;
esac
[ "${COUNT:-1}" -gt 0 ] || die "COUNT must be greater than zero"

case "${VALUE_WEI:-1}" in
  ''|*[!0-9]*) die "VALUE_WEI must be a non-negative integer" ;;
esac

if ! docker info >/dev/null 2>&1; then
  die "Docker daemon is not reachable"
fi

if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  die "linea-stack-shared-config volume not found. Boot the stack first."
fi

if [ -f versions.env ]; then
  # shellcheck disable=SC1091
  . ./versions.env
fi

FOUNDRY_IMAGE="${FOUNDRY_IMAGE:-ghcr.io/foundry-rs/foundry:${FOUNDRY_TAG:-latest}}"
L2_RPC_URL="${L2_RPC_URL:-http://sequencer:8545}"
HOST_PORT_L2_BLOCKSCOUT_FRONTEND="$(with_default "${HOST_PORT_L2_BLOCKSCOUT_FRONTEND:-$(env_value HOST_PORT_L2_BLOCKSCOUT_FRONTEND || true)}" 4001)"
BLOCKSCOUT_BASE_URL="${BLOCKSCOUT_BASE_URL:-http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND}"

section "sending L2 ETH transfer(s)"
if ! lineth_run_stream docker run --rm \
  --user 0:0 \
  --entrypoint sh \
  --network linea-stack_linea \
  -v linea-stack-shared-config:/shared:ro \
  -e COUNT="${COUNT:-1}" \
  -e VALUE_WEI="${VALUE_WEI:-1}" \
  -e TO="${TO:-}" \
  -e L2_RPC_URL="$L2_RPC_URL" \
  -e BLOCKSCOUT_BASE_URL="$BLOCKSCOUT_BASE_URL" \
  "$FOUNDRY_IMAGE" \
  -lc '
    set -eu

    [ -f /shared/runtime-keys.env ] || { echo "[l2-test-tx] ERROR: /shared/runtime-keys.env missing" >&2; exit 1; }
    [ -f /shared/addresses-precomputed.json ] || { echo "[l2-test-tx] ERROR: /shared/addresses-precomputed.json missing" >&2; exit 1; }

    . /shared/runtime-keys.env
    : "${L2_DEPLOYER_PRIVATE_KEY:?L2_DEPLOYER_PRIVATE_KEY missing from runtime-keys.env}"

    if [ -n "${TO:-}" ]; then
      recipient="$TO"
    else
      recipient=$(sed -nE "s/.*\"l2PostmanAddress\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" /shared/addresses-precomputed.json | head -1)
    fi
    echo "$recipient" | grep -qE "^0x[a-fA-F0-9]{40}$" || { echo "[l2-test-tx] ERROR: recipient address invalid: $recipient" >&2; exit 1; }

    i=1
    while [ "$i" -le "$COUNT" ]; do
      receipt=$(cast send "$recipient" \
        --value "$VALUE_WEI" \
        --private-key "$L2_DEPLOYER_PRIVATE_KEY" \
        --rpc-url "$L2_RPC_URL" \
        --json)

      tx_hash=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"transactionHash\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
      block_number=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"blockNumber\"[[:space:]]*:[[:space:]]*\"?([^\",}]+)\"?.*/\1/p" | head -1)
      [ -n "$tx_hash" ] || { echo "[l2-test-tx] ERROR: cast receipt did not include transactionHash" >&2; printf "%s\n" "$receipt" >&2; exit 1; }
      [ -n "$block_number" ] || block_number="unknown"

      printf "[l2-test-tx] %s/%s tx=%s block=%s to=%s valueWei=%s\n" "$i" "$COUNT" "$tx_hash" "$block_number" "$recipient" "$VALUE_WEI"
      printf "[l2-test-tx] blockscout=%s/tx/%s\n" "$BLOCKSCOUT_BASE_URL" "$tx_hash"
      i=$((i + 1))
    done

    current_block=$(cast block-number --rpc-url "$L2_RPC_URL")
    printf "[l2-test-tx] currentL2Block=%s\n" "$current_block"
  '
then
  exit 1
fi

section "links"
lineth_kv "Blockscout UI" "$BLOCKSCOUT_BASE_URL"
