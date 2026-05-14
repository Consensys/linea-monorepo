#!/usr/bin/env sh
# Transfer a tiny amount of the deployed L2 ERC20Example token.
set -eu

section() { printf '\n[l2-erc20-transfer] %s\n' "$*"; }
die() { printf '[l2-erc20-transfer] ERROR: %s\n' "$*" >&2; exit 1; }

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

case "${AMOUNT_WEI:-1}" in
  ''|*[!0-9]*) die "AMOUNT_WEI must be a positive integer" ;;
esac
[ "${AMOUNT_WEI:-1}" -gt 0 ] || die "AMOUNT_WEI must be greater than zero"

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
L2_TRAFFIC_ETH_MIN_BALANCE_WEI="${L2_TRAFFIC_ETH_MIN_BALANCE_WEI:-100000000000000000}"
L2_TRAFFIC_ETH_TOP_UP_WEI="${L2_TRAFFIC_ETH_TOP_UP_WEI:-1000000000000000000}"
L2_TRAFFIC_ERC20_MIN_BALANCE_WEI="${L2_TRAFFIC_ERC20_MIN_BALANCE_WEI:-100}"
L2_TRAFFIC_ERC20_TOP_UP_WEI="${L2_TRAFFIC_ERC20_TOP_UP_WEI:-10000}"

section "sending L2 ERC20Example transfer"
docker run --rm \
  --user 0:0 \
  --entrypoint sh \
  --network linea-stack_linea \
  -v linea-stack-shared-config:/shared:rw \
  -e AMOUNT_WEI="${AMOUNT_WEI:-1}" \
  -e TO="${TO:-}" \
  -e L2_TRAFFIC_PRIVATE_KEY="${L2_TRAFFIC_PRIVATE_KEY:-}" \
  -e L2_TRAFFIC_ETH_MIN_BALANCE_WEI="$L2_TRAFFIC_ETH_MIN_BALANCE_WEI" \
  -e L2_TRAFFIC_ETH_TOP_UP_WEI="$L2_TRAFFIC_ETH_TOP_UP_WEI" \
  -e L2_TRAFFIC_ERC20_MIN_BALANCE_WEI="$L2_TRAFFIC_ERC20_MIN_BALANCE_WEI" \
  -e L2_TRAFFIC_ERC20_TOP_UP_WEI="$L2_TRAFFIC_ERC20_TOP_UP_WEI" \
  -e L2_RPC_URL="$L2_RPC_URL" \
  -e BLOCKSCOUT_BASE_URL="$BLOCKSCOUT_BASE_URL" \
  "$FOUNDRY_IMAGE" \
  -lc '
    set -eu

    [ -f /shared/runtime-keys.env ] || { echo "[l2-erc20-transfer] ERROR: /shared/runtime-keys.env missing" >&2; exit 1; }
    [ -f /shared/addresses-precomputed.json ] || { echo "[l2-erc20-transfer] ERROR: /shared/addresses-precomputed.json missing" >&2; exit 1; }
    [ -f /shared/addresses.json ] || { echo "[l2-erc20-transfer] ERROR: /shared/addresses.json missing; deploy-contracts has not completed" >&2; exit 1; }

    . /shared/runtime-keys.env
    : "${L2_DEPLOYER_PRIVATE_KEY:?L2_DEPLOYER_PRIVATE_KEY missing from runtime-keys.env}"
    DEMO_TRAFFIC_ENV="/shared/demo-traffic.env"

    is_privkey() { printf "%s\n" "$1" | grep -qE "^0x[a-fA-F0-9]{64}$"; }
    is_uint() { printf "%s\n" "$1" | grep -qE "^[0-9]+$"; }

    for item in \
      L2_TRAFFIC_ETH_MIN_BALANCE_WEI \
      L2_TRAFFIC_ETH_TOP_UP_WEI \
      L2_TRAFFIC_ERC20_MIN_BALANCE_WEI \
      L2_TRAFFIC_ERC20_TOP_UP_WEI; do
      eval "value=\${$item:-}"
      is_uint "$value" || { echo "[l2-erc20-transfer] ERROR: $item must be a decimal wei value" >&2; exit 1; }
    done

    if [ -n "${L2_TRAFFIC_PRIVATE_KEY:-}" ]; then
      traffic_key="$L2_TRAFFIC_PRIVATE_KEY"
      echo "[l2-erc20-transfer] using L2_TRAFFIC_PRIVATE_KEY from environment"
    elif [ -f "$DEMO_TRAFFIC_ENV" ]; then
      . "$DEMO_TRAFFIC_ENV"
      traffic_key="${L2_TRAFFIC_PRIVATE_KEY:-}"
      echo "[l2-erc20-transfer] reusing disposable traffic account from $DEMO_TRAFFIC_ENV"
    else
      traffic_key=$(cast wallet new --json | sed -nE "s/.*\"private_key\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
      is_privkey "$traffic_key" || { echo "[l2-erc20-transfer] ERROR: failed to generate traffic private key" >&2; exit 1; }
      umask 077
      tmp="$DEMO_TRAFFIC_ENV.tmp"
      printf "L2_TRAFFIC_PRIVATE_KEY=%s\n" "$traffic_key" > "$tmp"
      mv "$tmp" "$DEMO_TRAFFIC_ENV"
      chmod 0644 "$DEMO_TRAFFIC_ENV"
      echo "[l2-erc20-transfer] created disposable traffic account in $DEMO_TRAFFIC_ENV"
    fi
    is_privkey "$traffic_key" || { echo "[l2-erc20-transfer] ERROR: L2 traffic private key malformed" >&2; exit 1; }

    erc20=$(sed -nE "/\"l2\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"ERC20Example\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" /shared/addresses.json | head -1)
    echo "$erc20" | grep -qE "^0x[a-fA-F0-9]{40}$" || { echo "[l2-erc20-transfer] ERROR: L2 ERC20Example missing from /shared/addresses.json" >&2; exit 1; }

    if [ -n "${TO:-}" ]; then
      recipient="$TO"
    else
      recipient="0x1000000000000000000000000000000000000001"
    fi
    echo "$recipient" | grep -qE "^0x[a-fA-F0-9]{40}$" || { echo "[l2-erc20-transfer] ERROR: recipient address invalid: $recipient" >&2; exit 1; }

    deployer=$(cast wallet address --private-key "$L2_DEPLOYER_PRIVATE_KEY")
    sender=$(cast wallet address --private-key "$traffic_key")

    eth_balance=$(cast balance "$sender" --rpc-url "$L2_RPC_URL" | awk "{print \$1}")
    is_uint "$eth_balance" || { echo "[l2-erc20-transfer] ERROR: could not read traffic account ETH balance" >&2; exit 1; }
    if [ "$eth_balance" -lt "$L2_TRAFFIC_ETH_MIN_BALANCE_WEI" ]; then
      echo "[l2-erc20-transfer] funding traffic account ETH from L2 deployer"
      cast send "$sender" --value "$L2_TRAFFIC_ETH_TOP_UP_WEI" \
        --private-key "$L2_DEPLOYER_PRIVATE_KEY" \
        --rpc-url "$L2_RPC_URL" >/dev/null
    fi

    token_balance=$(cast call "$erc20" "balanceOf(address)(uint256)" "$sender" --rpc-url "$L2_RPC_URL" | awk "{print \$1}")
    is_uint "$token_balance" || { echo "[l2-erc20-transfer] ERROR: could not read traffic account ERC20 balance" >&2; exit 1; }
    if [ "$token_balance" -lt "$L2_TRAFFIC_ERC20_MIN_BALANCE_WEI" ]; then
      echo "[l2-erc20-transfer] funding traffic account ERC20Example from L2 deployer"
      cast send "$erc20" "transfer(address,uint256)" "$sender" "$L2_TRAFFIC_ERC20_TOP_UP_WEI" \
        --private-key "$L2_DEPLOYER_PRIVATE_KEY" \
        --rpc-url "$L2_RPC_URL" >/dev/null
    fi

    receipt=$(cast send "$erc20" "transfer(address,uint256)" "$recipient" "$AMOUNT_WEI" \
      --private-key "$traffic_key" \
      --rpc-url "$L2_RPC_URL" \
      --json)

    tx_hash=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"transactionHash\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
    block_number=$(printf "%s\n" "$receipt" | sed -nE "s/.*\"blockNumber\"[[:space:]]*:[[:space:]]*\"?([^\",}]+)\"?.*/\1/p" | head -1)
    [ -n "$tx_hash" ] || { echo "[l2-erc20-transfer] ERROR: cast receipt did not include transactionHash" >&2; printf "%s\n" "$receipt" >&2; exit 1; }
    [ -n "$block_number" ] || block_number="unknown"

    printf "[l2-erc20-transfer] token=%s\n" "$erc20"
    printf "[l2-erc20-transfer] deployer=%s trafficAccount=%s\n" "$deployer" "$sender"
    printf "[l2-erc20-transfer] from=%s to=%s amountWei=%s\n" "$sender" "$recipient" "$AMOUNT_WEI"
    printf "[l2-erc20-transfer] tx=%s block=%s\n" "$tx_hash" "$block_number"
    printf "[l2-erc20-transfer] blockscout=%s/tx/%s\n" "$BLOCKSCOUT_BASE_URL" "$tx_hash"
    printf "[l2-erc20-transfer] tokenUrl=%s/token/%s\n" "$BLOCKSCOUT_BASE_URL" "$erc20"
  '

section "links"
printf '[l2-erc20-transfer] Blockscout UI: %s\n' "$BLOCKSCOUT_BASE_URL"
