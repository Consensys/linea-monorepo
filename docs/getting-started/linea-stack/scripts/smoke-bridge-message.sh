#!/usr/bin/env sh
# Bridge/message smoke scaffold. This is not a pass/fail bridge smoke test yet.
set -eu

section() { printf '\n[bridge-smoke] %s\n' "$*"; }
die() { printf '[bridge-smoke] ERROR: %s\n' "$*" >&2; exit 1; }

env_value() {
  key="$1"
  [ -f .env ] || return 1
  sed -nE "s/^${key}=([^#[:space:]].*)$/\1/p" .env | tail -1
}

json_addr() {
  file="$1"
  section="$2"
  key="$3"
  sed -nE "/\"$section\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" "$file" | head -1
}

json_field() {
  file="$1"
  key="$2"
  sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" "$file" | head -1
}

require_address() {
  label="$1"
  value="$2"
  echo "$value" | grep -qE '^0x[a-fA-F0-9]{40}$' || die "$label missing or invalid"
}

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

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/addresses-precomputed.json 2>/dev/null || true' > "$TMP_DIR/addresses-precomputed.json"
docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/addresses.json 2>/dev/null || true' > "$TMP_DIR/addresses.json"

PRE="$TMP_DIR/addresses-precomputed.json"
ADDR="$TMP_DIR/addresses.json"
[ -s "$PRE" ] || die "addresses-precomputed.json missing"
[ -s "$ADDR" ] || die "addresses.json missing; deploy-contracts has not completed"

LINEA_ROLLUP="$(json_addr "$ADDR" l1 LineaRollupV8)"
L2_MESSAGE_SERVICE="$(json_addr "$ADDR" l2 L2MessageService)"
L1_TOKEN_BRIDGE="$(json_addr "$ADDR" l1 TokenBridge)"
L2_TOKEN_BRIDGE="$(json_addr "$ADDR" l2 TokenBridge)"
DEFAULT_RECIPIENT="$(json_field "$PRE" l2DeployerAddress)"
RECIPIENT="${RECIPIENT:-$DEFAULT_RECIPIENT}"

require_address "L1 LineaRollupV8" "$LINEA_ROLLUP"
require_address "L2 L2MessageService" "$L2_MESSAGE_SERVICE"
require_address "RECIPIENT" "$RECIPIENT"

HOST_PORT_L2_BLOCKSCOUT_FRONTEND="${HOST_PORT_L2_BLOCKSCOUT_FRONTEND:-$(env_value HOST_PORT_L2_BLOCKSCOUT_FRONTEND || true)}"
[ -n "$HOST_PORT_L2_BLOCKSCOUT_FRONTEND" ] || HOST_PORT_L2_BLOCKSCOUT_FRONTEND=4001

section "preflight"
printf '[bridge-smoke] LineaRollupV8: https://sepolia.etherscan.io/address/%s\n' "$LINEA_ROLLUP"
printf '[bridge-smoke] L2MessageService: http://localhost:%s/address/%s\n' "$HOST_PORT_L2_BLOCKSCOUT_FRONTEND" "$L2_MESSAGE_SERVICE"
[ -n "$L1_TOKEN_BRIDGE" ] && printf '[bridge-smoke] L1 TokenBridge: https://sepolia.etherscan.io/address/%s\n' "$L1_TOKEN_BRIDGE"
[ -n "$L2_TOKEN_BRIDGE" ] && printf '[bridge-smoke] L2 TokenBridge: http://localhost:%s/address/%s\n' "$HOST_PORT_L2_BLOCKSCOUT_FRONTEND" "$L2_TOKEN_BRIDGE"
printf '[bridge-smoke] recipient: %s\n' "$RECIPIENT"

if [ "${BRIDGE_SMOKE_SEND:-0}" != "1" ]; then
  cat <<'EOF'
[bridge-smoke] This is not a pass/fail bridge smoke test yet.
[bridge-smoke] It only preflights the deployed bridge/message addresses by default.
[bridge-smoke] To submit an experimental L1->L2 message, run with:
[bridge-smoke]   BRIDGE_SMOKE_SEND=1 L1_MESSAGE_VALUE_WEI=<wei> ./scripts/smoke-bridge-message.sh
[bridge-smoke] The script will submit the L1 transaction but does not yet verify L2 claim/final delivery.
EOF
  exit 2
fi

L1_RPC_URL="${L1_RPC_URL:-$(env_value L1_RPC_URL || true)}"
L1_DEPLOYER_PRIVATE_KEY="${L1_DEPLOYER_PRIVATE_KEY:-$(env_value L1_DEPLOYER_PRIVATE_KEY || true)}"
[ -n "$L1_RPC_URL" ] || die "L1_RPC_URL missing from env/.env"
[ -n "$L1_DEPLOYER_PRIVATE_KEY" ] || die "L1_DEPLOYER_PRIVATE_KEY missing from env/.env"
[ -n "${L1_MESSAGE_VALUE_WEI:-}" ] || die "set L1_MESSAGE_VALUE_WEI explicitly before sending an experimental bridge message"
case "$L1_MESSAGE_VALUE_WEI" in
  ''|*[!0-9]*) die "L1_MESSAGE_VALUE_WEI must be a non-negative integer" ;;
esac

CALLDATA="${CALLDATA:-0x}"
FOUNDRY_IMAGE="${FOUNDRY_IMAGE:-ghcr.io/foundry-rs/foundry:${FOUNDRY_TAG:-latest}}"

section "experimental send"
docker run --rm \
  --entrypoint sh \
  -e L1_RPC_URL="$L1_RPC_URL" \
  -e L1_DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
  -e LINEA_ROLLUP="$LINEA_ROLLUP" \
  -e RECIPIENT="$RECIPIENT" \
  -e L1_MESSAGE_VALUE_WEI="$L1_MESSAGE_VALUE_WEI" \
  -e CALLDATA="$CALLDATA" \
  "$FOUNDRY_IMAGE" \
  -lc '
    set -eu
    receipt=$(cast send "$LINEA_ROLLUP" "sendMessage(address,uint256,bytes)" "$RECIPIENT" 0 "$CALLDATA" \
      --value "$L1_MESSAGE_VALUE_WEI" \
      --rpc-url "$L1_RPC_URL" \
      --private-key "$L1_DEPLOYER_PRIVATE_KEY" \
      --json)
    tx_hash=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"transactionHash\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
    [ -n "$tx_hash" ] || { echo "[bridge-smoke] ERROR: cast receipt did not include transactionHash" >&2; printf "%s\n" "$receipt" >&2; exit 1; }
    printf "[bridge-smoke] l1Tx=%s\n" "$tx_hash"
    printf "[bridge-smoke] etherscan=https://sepolia.etherscan.io/tx/%s\n" "$tx_hash"
  '

cat <<'EOF'
[bridge-smoke] L1 message submission completed.
[bridge-smoke] Full L2 claim/final-delivery verification is still pending; do not treat this as final bridge success.
EOF
