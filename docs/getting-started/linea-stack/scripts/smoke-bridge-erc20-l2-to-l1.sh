#!/usr/bin/env sh
# Real TokenBridge ERC20 L2->L1 withdrawal smoke test.
#
# Withdraws the L2 bridged token created by smoke-bridge-erc20-l1-to-l2.sh back
# to Sepolia through the TokenBridge. The script approves the L2 TokenBridge,
# calls bridgeToken(...), waits for L1 finality/anchoring, claims on L1 when
# needed, and verifies the L1 ERC20Example recipient balance increased.
set -eu

section() { printf '\n[erc20-bridge-l2-to-l1] %s\n' "$*"; }
log() { printf '[erc20-bridge-l2-to-l1] %s\n' "$*"; }
die() { printf '[erc20-bridge-l2-to-l1] ERROR: %s\n' "$*" >&2; exit 1; }

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

json_meta_uint() {
  file="$1"
  key="$2"
  sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"?([0-9]+)\"?.*/\1/p" "$file" | head -1
}

json_string_field() {
  field="$1"
  sed -nE "s/.*\"$field\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1
}

json_number_field() {
  field="$1"
  sed -nE "s/.*\"$field\"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p" | head -1
}

require_address() {
  label="$1"
  value="$2"
  echo "$value" | grep -qE '^0x[a-fA-F0-9]{40}$' || die "$label missing or invalid"
}

require_hash() {
  label="$1"
  value="$2"
  echo "$value" | grep -qE '^0x[a-fA-F0-9]{64}$' || die "$label missing or invalid"
}

require_uint() {
  label="$1"
  value="$2"
  case "$value" in
    ''|*[!0-9]*) die "$label must be a non-negative integer" ;;
  esac
}

rpc_l1() {
  method="$1"
  params="$2"
  curl -fsS "$L1_RPC_URL" \
    -H 'content-type: application/json' \
    --data "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"$method\",\"params\":$params}"
}

psql_value() {
  docker exec postman-pg psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-postman}" -At -F '|' -c "$1" \
    | tr -d '\r'
}

cast_wallet_address() {
  private_key="$1"
  docker run --rm \
    --entrypoint cast \
    "$FOUNDRY_IMAGE" wallet address --private-key "$private_key"
}

cast_event_topic() {
  signature="$1"
  docker run --rm \
    --entrypoint cast \
    "$FOUNDRY_IMAGE" sig-event "$signature"
}

cast_l1_call() {
  docker run --rm \
    --network linea-stack_linea \
    --entrypoint cast \
    "$FOUNDRY_IMAGE" call "$@" --rpc-url "$L1_RPC_URL"
}

cast_l2_call() {
  docker run --rm \
    --network linea-stack_linea \
    --entrypoint cast \
    "$FOUNDRY_IMAGE" call "$@" --rpc-url "$L2_READ_RPC_URL"
}

if ! docker info >/dev/null 2>&1; then
  die "Docker daemon is not reachable"
fi

if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  die "linea-stack-shared-config volume not found. Boot the stack first."
fi

if ! docker ps --format '{{.Names}}' | grep -qx 'postman-pg'; then
  die "postman-pg is not running. Boot the stack first."
fi

if ! docker ps --format '{{.Names}}' | grep -qx 'postman'; then
  die "postman is not running. Boot the stack first."
fi

if [ -f versions.env ]; then
  # shellcheck disable=SC1091
  . ./versions.env
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/addresses-precomputed.json 2>/dev/null || true' > "$TMP_DIR/addresses-precomputed.json"
docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/addresses.json 2>/dev/null || true' > "$TMP_DIR/addresses.json"
docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -c 'cat /shared/demo-traffic.env 2>/dev/null || true' > "$TMP_DIR/demo-traffic.env"

PRE="$TMP_DIR/addresses-precomputed.json"
ADDR="$TMP_DIR/addresses.json"
[ -s "$PRE" ] || die "addresses-precomputed.json missing"
[ -s "$ADDR" ] || die "addresses.json missing; deploy-contracts has not completed"

LINEA_ROLLUP="$(json_addr "$ADDR" l1 LineaRollupV8)"
L1_TOKEN_BRIDGE="$(json_addr "$ADDR" l1 TokenBridge)"
L2_TOKEN_BRIDGE="$(json_addr "$ADDR" l2 TokenBridge)"
L1_ERC20="$(json_addr "$ADDR" l1 ERC20Example)"
L2_MESSAGE_SERVICE="$(json_addr "$ADDR" l2 L2MessageService)"
L1_DEPLOYER_ADDRESS="$(json_addr "$PRE" deployers l1)"
L1_CHAIN_ID="$(json_meta_uint "$ADDR" l1ChainId)"
L2_CHAIN_ID="$(json_meta_uint "$ADDR" l2ChainId)"

require_address "L1 LineaRollupV8" "$LINEA_ROLLUP"
require_address "L1 TokenBridge" "$L1_TOKEN_BRIDGE"
require_address "L2 TokenBridge" "$L2_TOKEN_BRIDGE"
require_address "L1 ERC20Example" "$L1_ERC20"
require_address "L2 L2MessageService" "$L2_MESSAGE_SERVICE"
require_address "L1 deployer" "$L1_DEPLOYER_ADDRESS"
require_uint "l1ChainId" "$L1_CHAIN_ID"
require_uint "l2ChainId" "$L2_CHAIN_ID"

L1_RPC_URL="${L1_RPC_URL:-$(env_value L1_RPC_URL || true)}"
[ -n "$L1_RPC_URL" ] || die "L1_RPC_URL missing from env/.env"

HOST_PORT_L2_BLOCKSCOUT_FRONTEND="${HOST_PORT_L2_BLOCKSCOUT_FRONTEND:-$(env_value HOST_PORT_L2_BLOCKSCOUT_FRONTEND || true)}"
[ -n "$HOST_PORT_L2_BLOCKSCOUT_FRONTEND" ] || HOST_PORT_L2_BLOCKSCOUT_FRONTEND=4001

FOUNDRY_IMAGE="${FOUNDRY_IMAGE:-ghcr.io/foundry-rs/foundry:${FOUNDRY_TAG:-latest}}"
L2_READ_RPC_URL="${L2_READ_RPC_URL:-${L2_RPC_URL:-http://l2-node-besu:8545}}"
L2_SEND_RPC_URL="${L2_SEND_RPC_URL:-http://sequencer:8545}"
BRIDGE_AMOUNT_WEI="${BRIDGE_AMOUNT_WEI:-1000000000000000000}"
BRIDGE_MESSAGE_FEE_WEI="${BRIDGE_MESSAGE_FEE_WEI:-10000000000000000}"
BRIDGE_SMOKE_TIMEOUT_SECONDS="${BRIDGE_SMOKE_TIMEOUT_SECONDS:-7200}"
BRIDGE_SMOKE_POLL_SECONDS="${BRIDGE_SMOKE_POLL_SECONDS:-10}"
L1_RECEIPT_TIMEOUT_SECONDS="${L1_RECEIPT_TIMEOUT_SECONDS:-240}"
L2_TRAFFIC_ETH_MIN_BALANCE_WEI="${L2_TRAFFIC_ETH_MIN_BALANCE_WEI:-100000000000000000}"
L2_TRAFFIC_ETH_TOP_UP_WEI="${L2_TRAFFIC_ETH_TOP_UP_WEI:-1000000000000000000}"
L2_APPROVE_GAS_LIMIT="${L2_APPROVE_GAS_LIMIT:-100000}"
L2_BRIDGE_GAS_LIMIT="${L2_BRIDGE_GAS_LIMIT:-3000000}"
L2_GAS_PRICE_WEI="${L2_GAS_PRICE_WEI:-1000}"
L2_PRIORITY_GAS_PRICE_WEI="${L2_PRIORITY_GAS_PRICE_WEI:-1}"
MESSAGE_CLAIMED_TOPIC="0xa4c827e719e911e8f19393ccdb85b5102f08f0910604d340ba38390b7ff2ab0e"
BRIDGING_FINALIZED_V2_TOPIC="$(cast_event_topic 'BridgingFinalizedV2(address,address,uint256,address)')"

require_uint "BRIDGE_AMOUNT_WEI" "$BRIDGE_AMOUNT_WEI"
require_uint "BRIDGE_MESSAGE_FEE_WEI" "$BRIDGE_MESSAGE_FEE_WEI"
require_uint "BRIDGE_SMOKE_TIMEOUT_SECONDS" "$BRIDGE_SMOKE_TIMEOUT_SECONDS"
require_uint "BRIDGE_SMOKE_POLL_SECONDS" "$BRIDGE_SMOKE_POLL_SECONDS"
require_uint "L1_RECEIPT_TIMEOUT_SECONDS" "$L1_RECEIPT_TIMEOUT_SECONDS"
require_uint "L2_TRAFFIC_ETH_MIN_BALANCE_WEI" "$L2_TRAFFIC_ETH_MIN_BALANCE_WEI"
require_uint "L2_TRAFFIC_ETH_TOP_UP_WEI" "$L2_TRAFFIC_ETH_TOP_UP_WEI"
require_uint "L2_APPROVE_GAS_LIMIT" "$L2_APPROVE_GAS_LIMIT"
require_uint "L2_BRIDGE_GAS_LIMIT" "$L2_BRIDGE_GAS_LIMIT"
require_uint "L2_GAS_PRICE_WEI" "$L2_GAS_PRICE_WEI"
require_uint "L2_PRIORITY_GAS_PRICE_WEI" "$L2_PRIORITY_GAS_PRICE_WEI"
[ "$BRIDGE_AMOUNT_WEI" -gt 0 ] || die "BRIDGE_AMOUNT_WEI must be greater than zero"
[ "$BRIDGE_MESSAGE_FEE_WEI" -gt 0 ] || die "BRIDGE_MESSAGE_FEE_WEI must be greater than zero for Postman L2->L1 claiming"
require_hash "BridgingFinalizedV2 topic" "$BRIDGING_FINALIZED_V2_TOPIC"

if [ -n "${RECIPIENT:-}" ]; then
  L1_RECIPIENT="$RECIPIENT"
else
  L1_RECIPIENT="$L1_DEPLOYER_ADDRESS"
fi
require_address "L1 recipient" "$L1_RECIPIENT"

if [ -n "${L2_WITHDRAW_PRIVATE_KEY:-}" ]; then
  L2_SENDER="$(cast_wallet_address "$L2_WITHDRAW_PRIVATE_KEY")"
elif [ -s "$TMP_DIR/demo-traffic.env" ]; then
  # shellcheck disable=SC1090
  . "$TMP_DIR/demo-traffic.env"
  [ -n "${L2_TRAFFIC_PRIVATE_KEY:-}" ] || die "demo-traffic.env exists but L2_TRAFFIC_PRIVATE_KEY is missing"
  L2_SENDER="$(cast_wallet_address "$L2_TRAFFIC_PRIVATE_KEY")"
else
  die "no disposable traffic account found; run ./scripts/smoke-bridge-erc20-l1-to-l2.sh first"
fi
require_address "L2 sender" "$L2_SENDER"

L2_BRIDGED_TOKEN="$(cast_l2_call "$L2_TOKEN_BRIDGE" 'nativeToBridgedToken(uint256,address)(address)' "$L1_CHAIN_ID" "$L1_ERC20" | tr -d '[:space:]')"
require_address "L2 bridged token" "$L2_BRIDGED_TOKEN"
case "$L2_BRIDGED_TOKEN" in
  0x0000000000000000000000000000000000000000|0x0000000000000000000000000000000000000111|0x0000000000000000000000000000000000000222|0x0000000000000000000000000000000000000333)
    die "L2 bridged token is not deployed yet; run ./scripts/smoke-bridge-erc20-l1-to-l2.sh first"
    ;;
esac

L2_BALANCE_RAW="$(cast_l2_call "$L2_BRIDGED_TOKEN" 'balanceOf(address)(uint256)' "$L2_SENDER")"
L2_BALANCE="$(printf '%s\n' "$L2_BALANCE_RAW" | awk '{print $1}')"
require_uint "L2 sender bridged token balance" "$L2_BALANCE"
[ "$L2_BALANCE" -ge "$BRIDGE_AMOUNT_WEI" ] || {
  log "l2Sender: $L2_SENDER"
  log "l2BridgedToken: $L2_BRIDGED_TOKEN"
  log "l2BalanceWei: $L2_BALANCE"
  die "L2 sender does not have enough bridged ERC20; run ./scripts/smoke-bridge-erc20-l1-to-l2.sh first"
}

L1_BALANCE_BEFORE_RAW="$(cast_l1_call "$L1_ERC20" 'balanceOf(address)(uint256)' "$L1_RECIPIENT")"
L1_BALANCE_BEFORE="$(printf '%s\n' "$L1_BALANCE_BEFORE_RAW" | awk '{print $1}')"
require_uint "L1 recipient ERC20 balance before" "$L1_BALANCE_BEFORE"

START_MESSAGE_ID="$(psql_value "select coalesce(max(id),0) from message;")"
require_uint "postman max message id" "$START_MESSAGE_ID"

section "preflight"
log "L1 ERC20Example: https://sepolia.etherscan.io/address/$L1_ERC20"
log "L1 TokenBridge: https://sepolia.etherscan.io/address/$L1_TOKEN_BRIDGE"
log "L2 TokenBridge: http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/address/$L2_TOKEN_BRIDGE"
log "L2 bridged token: http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/token/$L2_BRIDGED_TOKEN"
log "l2Sender: $L2_SENDER"
log "l1Recipient: $L1_RECIPIENT"
log "amountWei: $BRIDGE_AMOUNT_WEI"
log "messageFeeWei: $BRIDGE_MESSAGE_FEE_WEI"
log "l2ReadRpc: $L2_READ_RPC_URL"
log "l2SendRpc: $L2_SEND_RPC_URL"
log "l1RecipientBalanceBeforeWei: $L1_BALANCE_BEFORE"

section "approve and withdraw on L2"
SEND_OUTPUT="$(
  docker run --rm \
    --user 0:0 \
    --entrypoint sh \
    --network linea-stack_linea \
    -v linea-stack-shared-config:/shared:rw \
    -e L2_WITHDRAW_PRIVATE_KEY="${L2_WITHDRAW_PRIVATE_KEY:-}" \
    -e L2_BRIDGED_TOKEN="$L2_BRIDGED_TOKEN" \
    -e L2_TOKEN_BRIDGE="$L2_TOKEN_BRIDGE" \
    -e L1_RECIPIENT="$L1_RECIPIENT" \
    -e BRIDGE_AMOUNT_WEI="$BRIDGE_AMOUNT_WEI" \
    -e BRIDGE_MESSAGE_FEE_WEI="$BRIDGE_MESSAGE_FEE_WEI" \
    -e L2_TRAFFIC_ETH_MIN_BALANCE_WEI="$L2_TRAFFIC_ETH_MIN_BALANCE_WEI" \
    -e L2_TRAFFIC_ETH_TOP_UP_WEI="$L2_TRAFFIC_ETH_TOP_UP_WEI" \
    -e L2_READ_RPC_URL="$L2_READ_RPC_URL" \
    -e L2_SEND_RPC_URL="$L2_SEND_RPC_URL" \
    -e L2_APPROVE_GAS_LIMIT="$L2_APPROVE_GAS_LIMIT" \
    -e L2_BRIDGE_GAS_LIMIT="$L2_BRIDGE_GAS_LIMIT" \
    -e L2_GAS_PRICE_WEI="$L2_GAS_PRICE_WEI" \
    -e L2_PRIORITY_GAS_PRICE_WEI="$L2_PRIORITY_GAS_PRICE_WEI" \
    "$FOUNDRY_IMAGE" \
    -lc '
      set -eu

      [ -f /shared/runtime-keys.env ] || { echo "[erc20-bridge-l2-to-l1] ERROR: /shared/runtime-keys.env missing" >&2; exit 1; }
      . /shared/runtime-keys.env
      : "${L2_DEPLOYER_PRIVATE_KEY:?L2_DEPLOYER_PRIVATE_KEY missing from runtime-keys.env}"
      DEMO_TRAFFIC_ENV="/shared/demo-traffic.env"

      is_privkey() { printf "%s\n" "$1" | grep -qE "^0x[a-fA-F0-9]{64}$"; }
      is_uint() { printf "%s\n" "$1" | grep -qE "^[0-9]+$"; }

      if [ -n "${L2_WITHDRAW_PRIVATE_KEY:-}" ]; then
        withdraw_key="$L2_WITHDRAW_PRIVATE_KEY"
        echo "[erc20-bridge-l2-to-l1] using L2_WITHDRAW_PRIVATE_KEY from environment"
      elif [ -f "$DEMO_TRAFFIC_ENV" ]; then
        . "$DEMO_TRAFFIC_ENV"
        withdraw_key="${L2_TRAFFIC_PRIVATE_KEY:-}"
        echo "[erc20-bridge-l2-to-l1] reusing disposable traffic account from $DEMO_TRAFFIC_ENV"
      else
        echo "[erc20-bridge-l2-to-l1] ERROR: no demo traffic account; run smoke-bridge-erc20-l1-to-l2.sh first" >&2
        exit 1
      fi
      is_privkey "$withdraw_key" || { echo "[erc20-bridge-l2-to-l1] ERROR: withdraw private key malformed" >&2; exit 1; }

      sender=$(cast wallet address --private-key "$withdraw_key")
      eth_balance=$(cast balance "$sender" --rpc-url "$L2_READ_RPC_URL" | awk "{print \$1}")
      is_uint "$eth_balance" || { echo "[erc20-bridge-l2-to-l1] ERROR: could not read withdraw account ETH balance" >&2; exit 1; }
      required_eth=$((L2_TRAFFIC_ETH_MIN_BALANCE_WEI + BRIDGE_MESSAGE_FEE_WEI))
      if [ "$eth_balance" -lt "$required_eth" ]; then
        echo "[erc20-bridge-l2-to-l1] funding withdraw account ETH from L2 deployer"
        cast send "$sender" --value "$L2_TRAFFIC_ETH_TOP_UP_WEI" \
          --private-key "$L2_DEPLOYER_PRIVATE_KEY" \
          --rpc-url "$L2_SEND_RPC_URL" >/dev/null
      fi

      approve_receipt=$(cast send "$L2_BRIDGED_TOKEN" "approve(address,uint256)" "$L2_TOKEN_BRIDGE" "$BRIDGE_AMOUNT_WEI" \
        --private-key "$withdraw_key" \
        --rpc-url "$L2_SEND_RPC_URL" \
        --gas-limit "$L2_APPROVE_GAS_LIMIT" \
        --gas-price "$L2_GAS_PRICE_WEI" \
        --priority-gas-price "$L2_PRIORITY_GAS_PRICE_WEI" \
        --json)
      approve_tx=$(printf "%s\n" "$approve_receipt" | sed -nE "s/.*\"transactionHash\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
      echo "$approve_tx" | grep -qE "^0x[a-fA-F0-9]{64}$" || { echo "[erc20-bridge-l2-to-l1] ERROR: approve receipt missing transactionHash" >&2; printf "%s\n" "$approve_receipt" >&2; exit 1; }

      bridge_receipt=$(cast send "$L2_TOKEN_BRIDGE" "bridgeToken(address,uint256,address)" "$L2_BRIDGED_TOKEN" "$BRIDGE_AMOUNT_WEI" "$L1_RECIPIENT" \
        --value "$BRIDGE_MESSAGE_FEE_WEI" \
        --private-key "$withdraw_key" \
        --rpc-url "$L2_SEND_RPC_URL" \
        --gas-limit "$L2_BRIDGE_GAS_LIMIT" \
        --gas-price "$L2_GAS_PRICE_WEI" \
        --priority-gas-price "$L2_PRIORITY_GAS_PRICE_WEI" \
        --json)
      bridge_tx=$(printf "%s\n" "$bridge_receipt" | sed -nE "s/.*\"transactionHash\"[[:space:]]*:[[:space:]]*\"([^\"]+)\".*/\1/p" | head -1)
      bridge_block=$(printf "%s\n" "$bridge_receipt" | sed -nE "s/.*\"blockNumber\"[[:space:]]*:[[:space:]]*\"?([^\",}]+)\"?.*/\1/p" | head -1)
      echo "$bridge_tx" | grep -qE "^0x[a-fA-F0-9]{64}$" || { echo "[erc20-bridge-l2-to-l1] ERROR: bridge receipt missing transactionHash" >&2; printf "%s\n" "$bridge_receipt" >&2; exit 1; }
      [ -n "$bridge_block" ] || bridge_block="unknown"

      printf "[erc20-bridge-l2-to-l1] sender=%s\n" "$sender"
      printf "[erc20-bridge-l2-to-l1] approveTx=%s\n" "$approve_tx"
      printf "[erc20-bridge-l2-to-l1] bridgeTx=%s\n" "$bridge_tx"
      printf "[erc20-bridge-l2-to-l1] bridgeBlock=%s\n" "$bridge_block"
    '
)"
printf '%s\n' "$SEND_OUTPUT"

APPROVE_TX_HASH="$(printf '%s\n' "$SEND_OUTPUT" | sed -nE 's/.*approveTx=(0x[a-fA-F0-9]{64}).*/\1/p' | tail -1)"
BRIDGE_TX_HASH="$(printf '%s\n' "$SEND_OUTPUT" | sed -nE 's/.*bridgeTx=(0x[a-fA-F0-9]{64}).*/\1/p' | tail -1)"
BRIDGE_BLOCK="$(printf '%s\n' "$SEND_OUTPUT" | sed -nE 's/.*bridgeBlock=([^[:space:]]+).*/\1/p' | tail -1)"
require_hash "L2 approve tx hash" "$APPROVE_TX_HASH"
require_hash "L2 bridge tx hash" "$BRIDGE_TX_HASH"
log "l2ApproveExplorer: http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/tx/$APPROVE_TX_HASH"
log "l2BridgeExplorer: http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/tx/$BRIDGE_TX_HASH"

section "wait for L1 finality/anchoring"
DEADLINE=$(( $(date +%s) + BRIDGE_SMOKE_TIMEOUT_SECONDS ))
ROW=""
READY=0
while [ "$(date +%s)" -le "$DEADLINE" ]; do
  ROW="$(psql_value "select id,status,message_hash,message_sender,destination,fee,value,message_nonce,calldata,sent_block_number,coalesce(claim_tx_hash,'') from message where id > $START_MESSAGE_ID and direction='L2_TO_L1' and lower(message_sender)=lower('$L2_TOKEN_BRIDGE') and lower(destination)=lower('$L1_TOKEN_BRIDGE') order by id desc limit 1;")"
  if [ -n "$ROW" ]; then
    STATUS="$(printf '%s' "$ROW" | cut -d '|' -f 2)"
    case "$STATUS" in
      CLAIMED_SUCCESS|ANCHORED|ZERO_FEE)
        READY=1
        break
        ;;
      PENDING)
        log "postmanStatus: PENDING; waiting for claim receipt"
        ;;
      *)
        log "postmanStatus: $STATUS"
        ;;
    esac
    case "$STATUS" in
      NON_EXECUTABLE|CLAIMED_REVERTED|FEE_UNDERPRICED|NEEDS_MANUAL_INTERVENTION)
        printf '%s\n' "$ROW" >&2
        die "postman moved ERC20 L2->L1 message to terminal/problem status: $STATUS"
        ;;
    esac
  else
    log "postmanStatus: pending ingest/finality"
  fi
  sleep "$BRIDGE_SMOKE_POLL_SECONDS"
done

[ -n "$ROW" ] || die "timed out waiting for postman to ingest the ERC20 L2->L1 message"
[ "$READY" -eq 1 ] || {
  printf '%s\n' "$ROW" >&2
  die "timed out waiting for ERC20 L2->L1 message to become claimable"
}

MESSAGE_ID="$(printf '%s' "$ROW" | cut -d '|' -f 1)"
STATUS="$(printf '%s' "$ROW" | cut -d '|' -f 2)"
MESSAGE_HASH="$(printf '%s' "$ROW" | cut -d '|' -f 3)"
MESSAGE_SENDER="$(printf '%s' "$ROW" | cut -d '|' -f 4)"
DESTINATION="$(printf '%s' "$ROW" | cut -d '|' -f 5)"
MESSAGE_FEE="$(printf '%s' "$ROW" | cut -d '|' -f 6)"
MESSAGE_VALUE="$(printf '%s' "$ROW" | cut -d '|' -f 7)"
MESSAGE_NONCE="$(printf '%s' "$ROW" | cut -d '|' -f 8)"
MESSAGE_CALLDATA="$(printf '%s' "$ROW" | cut -d '|' -f 9)"
SENT_BLOCK_NUMBER="$(printf '%s' "$ROW" | cut -d '|' -f 10)"
CLAIM_TX_HASH="$(printf '%s' "$ROW" | cut -d '|' -f 11)"

require_hash "messageHash" "$MESSAGE_HASH"
require_address "messageSender" "$MESSAGE_SENDER"
require_address "destination" "$DESTINATION"
require_uint "message fee" "$MESSAGE_FEE"
require_uint "message value" "$MESSAGE_VALUE"
require_uint "message nonce" "$MESSAGE_NONCE"
require_uint "sent block number" "$SENT_BLOCK_NUMBER"
echo "$MESSAGE_CALLDATA" | grep -qE '^0x([a-fA-F0-9]{2})*$' || die "message calldata malformed"

log "messageId: $MESSAGE_ID"
log "postmanStatus: $STATUS"
log "messageHash: $MESSAGE_HASH"
log "sentBlockNumber: $SENT_BLOCK_NUMBER"

if [ "$STATUS" != "CLAIMED_SUCCESS" ]; then
  section "claim on L1"
  CLAIM_OUTPUT=""
  if ! CLAIM_OUTPUT="$(
    docker exec \
      -e SMOKE_L1_CHAIN_ID="$L1_CHAIN_ID" \
      -e SMOKE_L2_CHAIN_ID="$L2_CHAIN_ID" \
      -e SMOKE_LINEA_ROLLUP_ADDRESS="$LINEA_ROLLUP" \
      -e SMOKE_L2_MESSAGE_SERVICE_ADDRESS="$L2_MESSAGE_SERVICE" \
      -e SMOKE_MESSAGE_HASH="$MESSAGE_HASH" \
      -e SMOKE_MESSAGE_SENDER="$MESSAGE_SENDER" \
      -e SMOKE_DESTINATION="$DESTINATION" \
      -e SMOKE_FEE="$MESSAGE_FEE" \
      -e SMOKE_VALUE="$MESSAGE_VALUE" \
      -e SMOKE_MESSAGE_NONCE="$MESSAGE_NONCE" \
      -e SMOKE_CALLDATA="$MESSAGE_CALLDATA" \
      -e SMOKE_SENT_BLOCK_NUMBER="$SENT_BLOCK_NUMBER" \
      postman \
      sh -lc 'set -eu; . /shared/runtime-keys.env; export L1_SIGNER_PRIVATE_KEY="$L1_POSTMAN_PRIVATE_KEY"; cd /usr/src/app/postman && node --input-type=module' <<'NODE'
const required = (name) => {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
};

const asBigInt = (name) => BigInt(required(name));
const chain = (id, name, rpcUrl) => ({
  id,
  name,
  nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
  rpcUrls: { default: { http: [rpcUrl] } },
});

const { createPublicClient, createWalletClient, http, zeroAddress } = await import("viem");
const { privateKeyToAccount } = await import("viem/accounts");
const { claimOnL1, getL2ToL1MessageStatus, getMessageProof } = await import("@consensys/linea-sdk-viem");

const l1RpcUrl = required("L1_RPC_URL");
const l2RpcUrl = required("L2_RPC_URL");
const l1Chain = chain(Number(required("SMOKE_L1_CHAIN_ID")), "sepolia", l1RpcUrl);
const l2Chain = chain(Number(required("SMOKE_L2_CHAIN_ID")), "local-linea", l2RpcUrl);
const account = privateKeyToAccount(required("L1_SIGNER_PRIVATE_KEY"));
const l1PublicClient = createPublicClient({ chain: l1Chain, transport: http(l1RpcUrl) });
const l1WalletClient = createWalletClient({ account, chain: l1Chain, transport: http(l1RpcUrl) });
const l2PublicClient = createPublicClient({ chain: l2Chain, transport: http(l2RpcUrl) });
const l2LogsBlockRange = {
  fromBlock: asBigInt("SMOKE_SENT_BLOCK_NUMBER"),
  toBlock: asBigInt("SMOKE_SENT_BLOCK_NUMBER"),
};

const common = {
  l2Client: l2PublicClient,
  messageHash: required("SMOKE_MESSAGE_HASH"),
  lineaRollupAddress: required("SMOKE_LINEA_ROLLUP_ADDRESS"),
  l2MessageServiceAddress: required("SMOKE_L2_MESSAGE_SERVICE_ADDRESS"),
  l2LogsBlockRange,
};

const status = await getL2ToL1MessageStatus(l1PublicClient, common);
if (status !== "CLAIMABLE") {
  throw new Error(`L2->L1 message is ${status}, not CLAIMABLE`);
}

const messageProof = await getMessageProof(l1PublicClient, common);
const claimTxHash = await claimOnL1(l1WalletClient, {
  from: required("SMOKE_MESSAGE_SENDER"),
  to: required("SMOKE_DESTINATION"),
  fee: asBigInt("SMOKE_FEE"),
  value: asBigInt("SMOKE_VALUE"),
  messageNonce: asBigInt("SMOKE_MESSAGE_NONCE"),
  calldata: required("SMOKE_CALLDATA"),
  feeRecipient: zeroAddress,
  messageProof,
  lineaRollupAddress: required("SMOKE_LINEA_ROLLUP_ADDRESS"),
});

console.log(
  JSON.stringify({
    status,
    claimTxHash,
    proofRoot: messageProof.root,
    proofLeafIndex: messageProof.leafIndex,
    proofLength: messageProof.proof.length,
    claimant: account.address,
  }),
);
NODE
  )"; then
    printf '%s\n' "$CLAIM_OUTPUT" >&2
    die "L1 SDK claim failed"
  fi

  CLAIM_TX_HASH="$(printf '%s\n' "$CLAIM_OUTPUT" | json_string_field claimTxHash)"
  PROOF_ROOT="$(printf '%s\n' "$CLAIM_OUTPUT" | json_string_field proofRoot)"
  PROOF_LEAF_INDEX="$(printf '%s\n' "$CLAIM_OUTPUT" | json_number_field proofLeafIndex)"
  PROOF_LENGTH="$(printf '%s\n' "$CLAIM_OUTPUT" | json_number_field proofLength)"
  CLAIMANT="$(printf '%s\n' "$CLAIM_OUTPUT" | json_string_field claimant)"
  require_hash "claim tx hash" "$CLAIM_TX_HASH"
  require_hash "proof root" "$PROOF_ROOT"
  require_uint "proof leaf index" "$PROOF_LEAF_INDEX"
  require_uint "proof length" "$PROOF_LENGTH"
  require_address "claimant" "$CLAIMANT"

  log "claimant: $CLAIMANT"
  log "proofRoot: $PROOF_ROOT"
  log "proofLeafIndex: $PROOF_LEAF_INDEX"
  log "proofLength: $PROOF_LENGTH"
  log "l1ClaimTx: $CLAIM_TX_HASH"
fi

require_hash "L1 claim tx" "$CLAIM_TX_HASH"
log "etherscan: https://sepolia.etherscan.io/tx/$CLAIM_TX_HASH"

section "verify L1 receipt and ERC20 balance"
RECEIPT_DEADLINE=$(( $(date +%s) + L1_RECEIPT_TIMEOUT_SECONDS ))
CLAIM_RECEIPT=""
while [ "$(date +%s)" -le "$RECEIPT_DEADLINE" ]; do
  CLAIM_RECEIPT="$(rpc_l1 eth_getTransactionReceipt "[\"$CLAIM_TX_HASH\"]")"
  if printf '%s\n' "$CLAIM_RECEIPT" | grep -q '"result":[[:space:]]*{'; then
    break
  fi
  sleep "$BRIDGE_SMOKE_POLL_SECONDS"
done

printf '%s\n' "$CLAIM_RECEIPT" | grep -q '"status":"0x1"' || {
  printf '%s\n' "$CLAIM_RECEIPT" >&2
  die "L1 claim receipt missing or failed"
}
printf '%s\n' "$CLAIM_RECEIPT" | grep -qi "$MESSAGE_CLAIMED_TOPIC" || {
  printf '%s\n' "$CLAIM_RECEIPT" >&2
  die "L1 claim receipt did not emit MessageClaimed"
}
printf '%s\n' "$CLAIM_RECEIPT" | grep -qi "$MESSAGE_HASH" || {
  printf '%s\n' "$CLAIM_RECEIPT" >&2
  die "L1 claim receipt MessageClaimed hash does not match postman message hash"
}
printf '%s\n' "$CLAIM_RECEIPT" | grep -qi "$BRIDGING_FINALIZED_V2_TOPIC" || {
  printf '%s\n' "$CLAIM_RECEIPT" >&2
  die "L1 claim receipt did not emit BridgingFinalizedV2"
}

L1_BALANCE_AFTER_RAW="$(cast_l1_call "$L1_ERC20" 'balanceOf(address)(uint256)' "$L1_RECIPIENT")"
L1_BALANCE_AFTER="$(printf '%s\n' "$L1_BALANCE_AFTER_RAW" | awk '{print $1}')"
require_uint "L1 recipient ERC20 balance after" "$L1_BALANCE_AFTER"

DELTA=$((L1_BALANCE_AFTER - L1_BALANCE_BEFORE))
[ "$DELTA" -ge "$BRIDGE_AMOUNT_WEI" ] || {
  log "l1RecipientBalanceBeforeWei: $L1_BALANCE_BEFORE"
  log "l1RecipientBalanceAfterWei: $L1_BALANCE_AFTER"
  die "L1 recipient ERC20 balance did not increase by at least $BRIDGE_AMOUNT_WEI"
}

section "success"
log "L2 bridge tx: http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/tx/$BRIDGE_TX_HASH"
[ -n "$BRIDGE_BLOCK" ] && log "L2 bridge block: $BRIDGE_BLOCK"
log "Sepolia claim tx: https://sepolia.etherscan.io/tx/$CLAIM_TX_HASH"
log "l1RecipientBalanceAfterWei: $L1_BALANCE_AFTER"
log "l1RecipientBalanceDeltaWei: $DELTA"
log "OK: ERC20 TokenBridge L2->L1 withdrawal verified"
