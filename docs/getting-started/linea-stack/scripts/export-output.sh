#!/usr/bin/env sh
# Export a local, shareable evidence bundle from the Lineth quickstart volumes.
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
STACK_DIR="$(CDPATH= cd "$SCRIPT_DIR/.." && pwd -P)"
LINETH_LOG_CONTEXT="export-output"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"

section() { lineth_section "$*"; }

lineth_banner "export · evidence bundle"

OUTPUT_DIR="${LINETH_OUTPUT_DIR:-$STACK_DIR/lineth-output}"
SHARED_VOLUME="linea-stack-shared-config"

env_value() {
  key="$1"
  if [ -f "$STACK_DIR/.env" ]; then
    sed -nE "s/^${key}=([^#[:space:]].*)$/\1/p" "$STACK_DIR/.env" | tail -1
  fi
}

with_default() {
  value="$1"
  fallback="$2"
  if [ -n "$value" ]; then printf '%s' "$value"; else printf '%s' "$fallback"; fi
}

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

json_value() {
  value="$1"
  if [ -n "$value" ]; then
    printf '"%s"' "$(json_escape "$value")"
  else
    printf 'null'
  fi
}

json_bool() {
  case "$1" in
    true) printf 'true' ;;
    *) printf 'false' ;;
  esac
}

json_string_field() {
  key="$1"
  sed -nE "s/.*\"${key}\"[[:space:]]*:[[:space:]]*\"([^\"]*)\".*/\1/p" | head -1
}

json_addr() {
  file="$1"
  section_name="$2"
  key="$3"
  [ -s "$file" ] || return 0
  sed -nE "/\"$section_name\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" "$file" | head -1
}

json_root_field() {
  file="$1"
  key="$2"
  [ -s "$file" ] || return 0
  sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"?([^\",}]*)\"?.*/\1/p" "$file" | head -1
}

hex_to_dec_small() {
  hex="$1"
  hex="${hex#0x}"
  [ -n "$hex" ] || { echo ""; return; }
  printf '%d\n' "$((16#$hex))" 2>/dev/null || printf '0x%s\n' "$hex"
}

rpc() {
  url="$1"
  method="$2"
  params="$3"
  [ -n "$url" ] || return 0
  curl -sS -X POST -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"$method\",\"params\":$params}" \
    "$url" 2>/dev/null || true
}

latest_log_hash() {
  pattern="$1"
  docker logs --tail 4000 coordinator 2>&1 \
    | sed -nE "$pattern" \
    | tail -1 || true
}

psql_json() {
  query="$1"
  if docker ps --format '{{.Names}}' | grep -qx postman-pg; then
    docker exec postman-pg psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-postman}" -At -c "$query" 2>/dev/null || printf '[]'
  else
    printf '[]'
  fi
}

if ! docker info >/dev/null 2>&1; then
  lineth_die "Docker daemon is not reachable."
fi

mkdir -p "$OUTPUT_DIR/deploy-logs"

section "copy shared runtime files"
if docker volume inspect "$SHARED_VOLUME" >/dev/null 2>&1; then
  docker run --rm \
    -v "$SHARED_VOLUME:/shared:ro" \
    -v "$OUTPUT_DIR:/out:rw" \
    busybox sh -eu -c '
      cp /shared/addresses-precomputed.json /out/addresses-precomputed.json 2>/dev/null || true
      cp /shared/addresses.json /out/addresses.json 2>/dev/null || true
      mkdir -p /out/deploy-logs
      cp /shared/deploy-logs/*.log /out/deploy-logs/ 2>/dev/null || true
    '
  lineth_ok "copied shared files into $OUTPUT_DIR"
else
  lineth_warn "$SHARED_VOLUME volume missing; boot the stack before exporting addresses"
fi

PRE="$OUTPUT_DIR/addresses-precomputed.json"
ADDR="$OUTPUT_DIR/addresses.json"

HOST_PORT_L2_RPC="$(with_default "${HOST_PORT_L2_RPC:-$(env_value HOST_PORT_L2_RPC || true)}" 8745)"
HOST_PORT_L2_BLOCKSCOUT="$(with_default "${HOST_PORT_L2_BLOCKSCOUT:-$(env_value HOST_PORT_L2_BLOCKSCOUT || true)}" 4000)"
HOST_PORT_L2_BLOCKSCOUT_FRONTEND="$(with_default "${HOST_PORT_L2_BLOCKSCOUT_FRONTEND:-$(env_value HOST_PORT_L2_BLOCKSCOUT_FRONTEND || true)}" 4001)"
HOST_PORT_POSTMAN="$(with_default "${HOST_PORT_POSTMAN:-$(env_value HOST_PORT_POSTMAN || true)}" 9090)"
HOST_PORT_COORDINATOR="$(with_default "${HOST_PORT_COORDINATOR:-$(env_value HOST_PORT_COORDINATOR || true)}" 9545)"
L2_RPC_URL="${L2_RPC_URL:-http://localhost:$HOST_PORT_L2_RPC}"
L1_RPC_URL="${L1_RPC_URL:-$(env_value L1_RPC_URL || true)}"

LINEA_ROLLUP="$(json_addr "$ADDR" l1 LineaRollupV8 || true)"
L1_TOKEN_BRIDGE="$(json_addr "$ADDR" l1 TokenBridge || true)"
L1_ERC20="$(json_addr "$ADDR" l1 ERC20Example || true)"
L2_MESSAGE_SERVICE="$(json_addr "$ADDR" l2 L2MessageService || true)"
L2_TOKEN_BRIDGE="$(json_addr "$ADDR" l2 TokenBridge || true)"
L2_ERC20="$(json_addr "$ADDR" l2 ERC20Example || true)"

section "write links.json"
cat > "$OUTPUT_DIR/links.json" <<EOF
{
  "local": {
    "l2Rpc": "http://localhost:$HOST_PORT_L2_RPC",
    "l2BlockscoutUi": "http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND",
    "l2BlockscoutApi": "http://localhost:$HOST_PORT_L2_BLOCKSCOUT",
    "postmanApi": "http://localhost:$HOST_PORT_POSTMAN",
    "coordinatorObservability": "http://localhost:$HOST_PORT_COORDINATOR"
  },
  "l1": {
    "LineaRollupV8": $(json_value "${LINEA_ROLLUP:+https://sepolia.etherscan.io/address/$LINEA_ROLLUP}"),
    "TokenBridge": $(json_value "${L1_TOKEN_BRIDGE:+https://sepolia.etherscan.io/address/$L1_TOKEN_BRIDGE}"),
    "ERC20Example": $(json_value "${L1_ERC20:+https://sepolia.etherscan.io/address/$L1_ERC20}")
  },
  "l2": {
    "L2MessageService": $(json_value "${L2_MESSAGE_SERVICE:+http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/address/$L2_MESSAGE_SERVICE}"),
    "TokenBridge": $(json_value "${L2_TOKEN_BRIDGE:+http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/address/$L2_TOKEN_BRIDGE}"),
    "ERC20Example": $(json_value "${L2_ERC20:+http://localhost:$HOST_PORT_L2_BLOCKSCOUT_FRONTEND/address/$L2_ERC20}")
  }
}
EOF
lineth_ok "$OUTPUT_DIR/links.json"

section "write finality-report.json"
local_chain_resp="$(rpc "$L2_RPC_URL" eth_chainId '[]')"
local_chain_hex="$(printf '%s\n' "$local_chain_resp" | json_string_field result)"
local_block_resp="$(rpc "$L2_RPC_URL" eth_blockNumber '[]')"
local_block_hex="$(printf '%s\n' "$local_block_resp" | json_string_field result)"
local_block_dec="$(hex_to_dec_small "$local_block_hex")"
addresses_chain_id="$(json_root_field "$ADDR" l2ChainId || true)"

latest_blob_tx=""
latest_finalization_tx=""
latest_finalization_window=""
if docker ps -a --format '{{.Names}}' | grep -qx coordinator; then
  latest_blob_tx="$(latest_log_hash 's/.*blobs submitted:.*transactionHash=(0x[a-fA-F0-9]{64}).*/\1/p')"
  latest_finalization_tx="$(latest_log_hash 's/.*submitted aggregation=[^ ]+ transactionHash=(0x[a-fA-F0-9]{64}).*/\1/p')"
  latest_finalization_window="$(latest_log_hash 's/.*submitted aggregation=([^ ]+) transactionHash=0x[a-fA-F0-9]{64}.*/\1/p')"
fi

rollup_block_hex=""
rollup_block_dec=""
finalization_selector=""
finalization_receipt_status=""
data_finalized="false"
state_updated="false"
state_mismatch="false"

if [ -n "$L1_RPC_URL" ] && [ -n "$LINEA_ROLLUP" ]; then
  rollup_resp="$(rpc "$L1_RPC_URL" eth_call "[{\"to\":\"$LINEA_ROLLUP\",\"data\":\"0x695378f5\"},\"latest\"]")"
  rollup_block_hex="$(printf '%s\n' "$rollup_resp" | json_string_field result)"
  rollup_block_dec="$(hex_to_dec_small "$rollup_block_hex")"

  if [ -n "$latest_finalization_tx" ]; then
    tx_resp="$(rpc "$L1_RPC_URL" eth_getTransactionByHash "[\"$latest_finalization_tx\"]")"
    finalization_input="$(printf '%s\n' "$tx_resp" | json_string_field input)"
    finalization_selector="$(printf '%.10s' "$finalization_input")"

    receipt_resp="$(rpc "$L1_RPC_URL" eth_getTransactionReceipt "[\"$latest_finalization_tx\"]")"
    finalization_receipt_status="$(printf '%s\n' "$receipt_resp" | json_string_field status)"
    case "$receipt_resp" in
      *a0262dc79e4ccb71ceac8574ae906311ae338aa4a2044fd4ec4b99fad5ab60cb*) data_finalized="true" ;;
    esac
    case "$receipt_resp" in
      *32e016ccc5c33419c35caa94023fdeb75143da613fb2ac738ab736404c09fc5d*) state_updated="true" ;;
    esac
  fi
fi

case "$rollup_block_dec:$local_block_dec" in
  *[!0-9]*:*) ;;
  *:*[!0-9]*) ;;
  :*) ;;
  *:) ;;
  *)
    if [ "$rollup_block_dec" -gt "$local_block_dec" ]; then
      state_mismatch="true"
    fi
    ;;
esac

cat > "$OUTPUT_DIR/finality-report.json" <<EOF
{
  "localL2": {
    "rpcUrl": $(json_value "$L2_RPC_URL"),
    "chainIdHex": $(json_value "$local_chain_hex"),
    "latestBlockHex": $(json_value "$local_block_hex"),
    "latestBlockNumber": $(json_value "$local_block_dec")
  },
  "addresses": {
    "l2ChainId": $(json_value "$addresses_chain_id"),
    "lineaRollupV8": $(json_value "$LINEA_ROLLUP")
  },
  "l1": {
    "latestBlobTxHash": $(json_value "$latest_blob_tx"),
    "latestFinalizationTxHash": $(json_value "$latest_finalization_tx"),
    "latestFinalizationWindow": $(json_value "$latest_finalization_window"),
    "rollupCurrentL2BlockHex": $(json_value "$rollup_block_hex"),
    "rollupCurrentL2BlockNumber": $(json_value "$rollup_block_dec"),
    "latestFinalizationSelector": $(json_value "$finalization_selector"),
    "latestFinalizationReceiptStatus": $(json_value "$finalization_receipt_status"),
    "dataFinalizedV3": $(json_bool "$data_finalized"),
    "finalizedStateUpdated": $(json_bool "$state_updated")
  },
  "stateMismatch": $(json_bool "$state_mismatch")
}
EOF
lineth_ok "$OUTPUT_DIR/finality-report.json"

section "write smoke-report.json"
POSTMAN_SUMMARY="$(psql_json "select coalesce(json_agg(row_to_json(t))::text,'[]') from (select direction,status,count(*)::int as count from message group by direction,status order by direction,status) t;")"
POSTMAN_LATEST="$(psql_json "select coalesce(json_agg(row_to_json(t))::text,'[]') from (select id,direction,status,message_hash,coalesce(claim_tx_hash,'') as claim_tx_hash,message_sender,destination,value,message_nonce from message order by id desc limit 10) t;")"
[ -n "$POSTMAN_SUMMARY" ] || POSTMAN_SUMMARY="[]"
[ -n "$POSTMAN_LATEST" ] || POSTMAN_LATEST="[]"
GENERATED_AT="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

cat > "$OUTPUT_DIR/smoke-report.json" <<EOF
{
  "generatedAt": $(json_value "$GENERATED_AT"),
  "source": "postman database",
  "postmanMessageSummary": $POSTMAN_SUMMARY,
  "latestPostmanMessages": $POSTMAN_LATEST,
  "smokeCommands": [
    "./scripts/smoke-bridge-message.sh",
    "./scripts/smoke-bridge-message-l2-to-l1.sh",
    "./scripts/smoke-bridge-erc20-l1-to-l2.sh",
    "./scripts/smoke-bridge-erc20-l2-to-l1.sh"
  ]
}
EOF
lineth_ok "$OUTPUT_DIR/smoke-report.json"

section "done"
lineth_kv "output" "$OUTPUT_DIR"
