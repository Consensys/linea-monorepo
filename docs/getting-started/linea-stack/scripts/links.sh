#!/usr/bin/env sh
# Print local quickstart links and L1/L2 explorer URLs from the shared runtime volume.
set -eu

section() { printf '\n[linea-links] %s\n' "$*"; }

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

HOST_PORT_L2_RPC="$(with_default "${HOST_PORT_L2_RPC:-$(env_value HOST_PORT_L2_RPC || true)}" 8745)"
HOST_PORT_L2_BLOCKSCOUT="$(with_default "${HOST_PORT_L2_BLOCKSCOUT:-$(env_value HOST_PORT_L2_BLOCKSCOUT || true)}" 4000)"
HOST_PORT_L2_BLOCKSCOUT_FRONTEND="$(with_default "${HOST_PORT_L2_BLOCKSCOUT_FRONTEND:-$(env_value HOST_PORT_L2_BLOCKSCOUT_FRONTEND || true)}" 4001)"
HOST_PORT_POSTMAN="$(with_default "${HOST_PORT_POSTMAN:-$(env_value HOST_PORT_POSTMAN || true)}" 9090)"
HOST_PORT_COORDINATOR="$(with_default "${HOST_PORT_COORDINATOR:-$(env_value HOST_PORT_COORDINATOR || true)}" 9545)"

if ! docker info >/dev/null 2>&1; then
  echo "[linea-links] ERROR: Docker daemon is not reachable." >&2
  exit 1
fi

if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  echo "[linea-links] ERROR: linea-stack-shared-config volume not found. Boot the stack first." >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/addresses-precomputed.json 2>/dev/null || true' > "$TMP_DIR/addresses-precomputed.json"
docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/addresses.json 2>/dev/null || true' > "$TMP_DIR/addresses.json"

PRE="$TMP_DIR/addresses-precomputed.json"
ADDR="$TMP_DIR/addresses.json"

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

print_l1() {
  label="$1"
  addr="$2"
  [ -n "$addr" ] || return 0
  printf '%-28s %s\n' "$label" "https://sepolia.etherscan.io/address/$addr"
}

print_l2() {
  label="$1"
  addr="$2"
  [ -n "$addr" ] || return 0
  printf '%-28s %s\n' "$label" "http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/address/$addr"
}

section "local services"
printf '%-28s %s\n' "L2 RPC" "http://localhost:$HOST_PORT_L2_RPC"
printf '%-28s %s\n' "L2 Blockscout UI" "http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND"
printf '%-28s %s\n' "L2 Blockscout API" "http://localhost:$HOST_PORT_L2_BLOCKSCOUT"
printf '%-28s %s\n' "Postman API" "http://localhost:$HOST_PORT_POSTMAN"
printf '%-28s %s\n' "Coordinator observability" "http://localhost:$HOST_PORT_COORDINATOR"

if [ -s "$PRE" ]; then
  section "pre-boot contract links"
  print_l1 "LineaRollupV8" "$(json_addr "$PRE" l1 LineaRollupV8)"
  print_l2 "L2MessageService" "$(json_addr "$PRE" l2 L2MessageService)"

  section "runtime signer addresses"
  printf '%-28s %s\n' "L1 blob submitter" "$(json_field "$PRE" l1BlobSubmitterAddress)"
  printf '%-28s %s\n' "L1 finalization" "$(json_field "$PRE" l1FinalizationSubmitterAddress)"
  printf '%-28s %s\n' "L1 postman" "$(json_field "$PRE" l1PostmanAddress)"
  printf '%-28s %s\n' "L2 deployer" "$(json_field "$PRE" l2DeployerAddress)"
  printf '%-28s %s\n' "L2 anchorer" "$(json_field "$PRE" l2MessageAnchoringAddress)"
  printf '%-28s %s\n' "L2 postman" "$(json_field "$PRE" l2PostmanAddress)"
else
  section "pre-boot contract links"
  echo "addresses-precomputed.json missing"
fi

if [ -s "$ADDR" ]; then
  section "deployed L1 contracts"
  print_l1 "LineaRollupV8" "$(json_addr "$ADDR" l1 LineaRollupV8)"
  print_l1 "TokenBridge" "$(json_addr "$ADDR" l1 TokenBridge)"
  print_l1 "ERC20Example" "$(json_addr "$ADDR" l1 ERC20Example)"
  print_l1 "ForcedTxGateway unused" "$(json_addr "$ADDR" l1 ForcedTransactionGateway)"

  section "deployed L2 contracts"
  print_l2 "L2MessageService" "$(json_addr "$ADDR" l2 L2MessageService)"
  print_l2 "TokenBridge" "$(json_addr "$ADDR" l2 TokenBridge)"
  print_l2 "ERC20Example" "$(json_addr "$ADDR" l2 ERC20Example)"
else
  section "deployed contracts"
  echo "addresses.json missing; deploy-contracts has not completed yet"
fi
