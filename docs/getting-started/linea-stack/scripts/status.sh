#!/usr/bin/env sh
# Print a redacted, milestone-oriented status report for the Lineth quickstart.
# Run from docs/getting-started/linea-stack while the stack is booting.
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
LINETH_LOG_CONTEXT="status"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"

section() { lineth_section "$*"; }

lineth_banner "status · local L2 + Sepolia finality"

env_value() {
  key="$1"
  if [ -f .env ]; then
    sed -nE "s/^${key}=([^#[:space:]].*)$/\1/p" .env | tail -1
  fi
}

json_string_field() {
  key="$1"
  sed -nE "s/.*\"${key}\"[[:space:]]*:[[:space:]]*\"([^\"]*)\".*/\1/p" | head -1
}

hex_to_dec_small() {
  hex="$1"
  hex="${hex#0x}"
  [ -n "$hex" ] || { echo "0"; return; }
  # currentL2BlockNumber is small in the quickstart; fall back to hex if a
  # shell cannot fit the value in its native integer range.
  printf '%d\n' "$((16#$hex))" 2>/dev/null || printf '0x%s\n' "$hex"
}

is_uint() {
  case "$1" in
    ''|*[!0-9]*) return 1 ;;
    *) return 0 ;;
  esac
}

shared_addr() {
  file="$1"
  section_name="$2"
  key="$3"
  docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c "
    [ -f /shared/$file ] || exit 0
    sed -nE '/\"$section_name\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p' /shared/$file | head -1
  " 2>/dev/null || true
}

rpc_json() {
  url="$1"
  method="$2"
  params="$3"
  [ -n "$url" ] || return 0
  curl -sS -X POST -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"$method\",\"params\":$params}" \
    "$url" 2>/dev/null || true
}

check_contract_code() {
  scope="$1"
  url="$2"
  label="$3"
  address="$4"

  [ -n "$address" ] || return 0
  if [ -z "$url" ]; then
    lineth_warn "$scope code check skipped for $label: RPC unavailable"
    STATE_MISMATCH=1
    return 0
  fi

  code_resp="$(rpc_json "$url" eth_getCode "[\"$address\",\"latest\"]")"
  code="$(printf '%s\n' "$code_resp" | json_string_field result)"
  if [ -z "$code_resp" ] || [ -z "$code" ]; then
    lineth_warn "$scope $label code check unavailable at $address"
    return 0
  fi
  case "$code" in
    "0x")
      lineth_error "$scope $label has no code at $address"
      STATE_MISMATCH=1
      ;;
    *)
      lineth_kv "$scope $label code" "ok ($address)"
      ;;
  esac
}

if ! docker info >/dev/null 2>&1; then
  lineth_die "Docker daemon is not reachable from this shell."
fi

STATE_MISMATCH=0

section "containers"
docker ps -a --format 'table {{.Names}}\t{{.Status}}' \
  | awk 'NR == 1 || /^(account-setup|config-render|l2-genesis-init|sequencer|l2-node-besu|shomei|deploy-contracts|coordinator|prover|postman|web3signer|maru|l2-blockscout)[[:space:]]/' \
  | lineth_indent

section "deploy milestones"
if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  lineth_warn "shared volume missing: deploy has not started"
else
  docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c '
    if [ -f /shared/addresses-precomputed.json ]; then
      echo "addresses-precomputed.json: present"
    else
      echo "addresses-precomputed.json: missing"
    fi

    if [ -f /shared/addresses.json ]; then
      echo "addresses.json: present"
      sed -nE "s/.*\"LineaRollupV8\":[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/L1 LineaRollupV8: \1/p" /shared/addresses.json | head -1
      sed -nE "s/.*\"TokenBridge\":[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/TokenBridge: \1/p" /shared/addresses.json | head -2
      sed -nE "s/.*\"ERC20Example\":[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/ERC20Example: \1/p" /shared/addresses.json | head -2
    else
      echo "addresses.json: missing"
    fi

    for f in /shared/deploy-logs/*.log; do
      [ -f "$f" ] || continue
      name="${f##*/}"
      count=$(grep -c "^contract=.* deployed:" "$f" || true)
      if [ "$count" -gt 0 ]; then
        echo "$name: $count contract deploy log(s)"
        grep -E "^contract=.* deployed:|\[deploy-contracts\] verify .*: OK|ERROR|ADDRESS MISMATCH" "$f" | tail -12
      else
        echo "$name: present, no deploy marker yet"
      fi
    done
  ' | lineth_indent
fi

section "chain / prover config"
HOST_PORT_L2_RPC="${HOST_PORT_L2_RPC:-$(env_value HOST_PORT_L2_RPC || true)}"
[ -n "$HOST_PORT_L2_RPC" ] || HOST_PORT_L2_RPC=8745
L2_RPC_URL="${L2_RPC_URL:-http://localhost:$HOST_PORT_L2_RPC}"
chain_response="$(rpc_json "$L2_RPC_URL" eth_chainId '[]')"
local_l2_chain_hex="$(printf '%s\n' "$chain_response" | json_string_field result)"
local_l2_chain_dec=""
if [ -n "$chain_response" ]; then
  lineth_kv "l2 rpc eth_chainId" "$chain_response"
  if [ -n "$local_l2_chain_hex" ]; then
    local_l2_chain_dec="$(hex_to_dec_small "$local_l2_chain_hex")"
  fi
else
  lineth_warn "l2 rpc eth_chainId unavailable at $L2_RPC_URL"
fi

local_l2_latest_block_dec=""
block_response="$(rpc_json "$L2_RPC_URL" eth_blockNumber '[]')"
local_l2_latest_block_hex="$(printf '%s\n' "$block_response" | json_string_field result)"
if [ -n "$local_l2_latest_block_hex" ] && [ "$local_l2_latest_block_hex" != "0x" ]; then
  local_l2_latest_block_dec="$(hex_to_dec_small "$local_l2_latest_block_hex")"
  lineth_kv "l2 rpc latest block" "$local_l2_latest_block_dec ($local_l2_latest_block_hex)"
else
  lineth_warn "l2 rpc latest block unavailable at $L2_RPC_URL"
fi

if docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c '
    if [ -f /shared/addresses.json ]; then
      sed -nE "s/.*\"l2ChainId\":[[:space:]]*\"?([0-9]+)\"?.*/addresses.json l2ChainId: \1/p" /shared/addresses.json | head -1
    fi
  ' | lineth_indent
fi

if docker volume inspect linea-stack-rendered-config >/dev/null 2>&1; then
  docker run --rm -v linea-stack-rendered-config:/rendered:ro busybox sh -eu -c '
    if [ -f /rendered/prover-config-partial.toml ]; then
      awk "
        /^\\[[^]]+\\]/ {
          section = \$0
          gsub(/^\\[/, \"\", section)
          gsub(/\\]$/, \"\", section)
          next
        }
        /^prover_mode[[:space:]]*=/ ||
        /^is_allowed_circuit_id[[:space:]]*=/ ||
        /^chain_id[[:space:]]*=/ {
          printf \"%s.%s\\n\", section, \$0
        }
      " /rendered/prover-config-partial.toml
    else
      echo "rendered prover config: missing"
    fi
  ' | lineth_indent
else
  lineth_warn "rendered config volume missing"
fi

section "state mismatch guardrails"
L2_MESSAGE_SERVICE_ADDRESS=""
L2_TOKEN_BRIDGE_ADDRESS=""
L2_ERC20_ADDRESS=""
L1_TOKEN_BRIDGE_ADDRESS=""
L1_ERC20_ADDRESS=""
ADDRESSES_L2_CHAIN_ID=""
L2_DEPLOY_MAX_BLOCK=""
if docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  L2_MESSAGE_SERVICE_ADDRESS="$(shared_addr addresses.json l2 L2MessageService)"
  L2_TOKEN_BRIDGE_ADDRESS="$(shared_addr addresses.json l2 TokenBridge)"
  L2_ERC20_ADDRESS="$(shared_addr addresses.json l2 ERC20Example)"
  L1_TOKEN_BRIDGE_ADDRESS="$(shared_addr addresses.json l1 TokenBridge)"
  L1_ERC20_ADDRESS="$(shared_addr addresses.json l1 ERC20Example)"
  ADDRESSES_L2_CHAIN_ID="$(docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c '
    [ -f /shared/addresses.json ] || exit 0
    sed -nE "s/.*\"l2ChainId\"[[:space:]]*:[[:space:]]*\"?([0-9]+)\"?.*/\1/p" /shared/addresses.json | head -1
  ' 2>/dev/null || true)"
  L2_DEPLOY_MAX_BLOCK="$(docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c '
    max=0
    for f in /shared/deploy-logs/*.log; do
      [ -f "$f" ] || continue
      awk "/contract=.* deployed:.*chainId=1337/ {
        if (match(\$0, /blockNumber=[0-9]+/)) {
          v=substr(\$0, RSTART + 12, RLENGTH - 12)
          if (v > max) max = v
        }
      } END { if (max > 0) print max }" "$f"
    done | awk "BEGIN { max=0 } { if (\$1 > max) max=\$1 } END { if (max > 0) print max }"
  ' 2>/dev/null || true)"
fi

if [ -n "$ADDRESSES_L2_CHAIN_ID" ] && [ -n "$local_l2_chain_dec" ]; then
  if [ "$ADDRESSES_L2_CHAIN_ID" = "$local_l2_chain_dec" ]; then
    lineth_kv "addresses.json vs local chainId" "ok ($ADDRESSES_L2_CHAIN_ID)"
  else
    lineth_error "addresses.json l2ChainId=$ADDRESSES_L2_CHAIN_ID but local eth_chainId=$local_l2_chain_dec"
    STATE_MISMATCH=1
  fi
else
  lineth_info "addresses.json/local chainId comparison unavailable"
fi

if [ -n "$L2_DEPLOY_MAX_BLOCK" ] && is_uint "$L2_DEPLOY_MAX_BLOCK" && is_uint "$local_l2_latest_block_dec"; then
  if [ "$L2_DEPLOY_MAX_BLOCK" -le "$local_l2_latest_block_dec" ]; then
    lineth_kv "L2 deploy-log block floor" "ok (max deployed block $L2_DEPLOY_MAX_BLOCK <= local latest $local_l2_latest_block_dec)"
  else
    lineth_error "deploy logs reference L2 block $L2_DEPLOY_MAX_BLOCK but local latest block is $local_l2_latest_block_dec"
    STATE_MISMATCH=1
  fi
else
  lineth_info "L2 deploy-log block floor unavailable"
fi

check_contract_code "L2" "$L2_RPC_URL" "L2MessageService" "$L2_MESSAGE_SERVICE_ADDRESS"
check_contract_code "L2" "$L2_RPC_URL" "TokenBridge" "$L2_TOKEN_BRIDGE_ADDRESS"
check_contract_code "L2" "$L2_RPC_URL" "ERC20Example" "$L2_ERC20_ADDRESS"

section "l1 data availability vs finalization"
LINEA_ROLLUP_ADDRESS=""
if docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  LINEA_ROLLUP_ADDRESS="$(shared_addr addresses.json l1 LineaRollupV8)"
fi

if [ -n "$LINEA_ROLLUP_ADDRESS" ]; then
  lineth_kv "LineaRollupV8" "$LINEA_ROLLUP_ADDRESS"
else
  lineth_warn "LineaRollupV8 unavailable until addresses.json exists"
fi

latest_blob_tx=""
latest_finalization_tx=""
latest_finalization_window=""
if docker ps -a --format '{{.Names}}' | grep -qx coordinator; then
  latest_blob_tx="$(docker logs --tail 4000 coordinator 2>&1 \
    | sed -nE 's/.*blobs submitted:.*transactionHash=(0x[a-fA-F0-9]{64}).*/\1/p' \
    | tail -1 || true)"
  latest_finalization_tx="$(docker logs --tail 4000 coordinator 2>&1 \
    | sed -nE 's/.*submitted aggregation=[^ ]+ transactionHash=(0x[a-fA-F0-9]{64}).*/\1/p' \
    | tail -1 || true)"
  latest_finalization_window="$(docker logs --tail 4000 coordinator 2>&1 \
    | sed -nE 's/.*submitted aggregation=([^ ]+) transactionHash=0x[a-fA-F0-9]{64}.*/\1/p' \
    | tail -1 || true)"
fi

if [ -n "$latest_blob_tx" ]; then
  lineth_kv "latest blob tx (DA only)" "$latest_blob_tx"
else
  lineth_info "latest blob tx (DA only): none seen in coordinator logs yet"
fi

if [ -n "$latest_finalization_tx" ]; then
  if [ -n "$latest_finalization_window" ]; then
    lineth_kv "latest finalization tx" "$latest_finalization_tx aggregation=$latest_finalization_window"
  else
    lineth_kv "latest finalization tx" "$latest_finalization_tx"
  fi
else
  lineth_info "latest finalization tx: none seen in coordinator logs yet"
fi

L1_RPC_URL="${L1_RPC_URL:-$(env_value L1_RPC_URL || true)}"
if [ -n "$L1_RPC_URL" ] && [ -n "$LINEA_ROLLUP_ADDRESS" ]; then
  check_contract_code "L1" "$L1_RPC_URL" "LineaRollupV8" "$LINEA_ROLLUP_ADDRESS"
  check_contract_code "L1" "$L1_RPC_URL" "TokenBridge" "$L1_TOKEN_BRIDGE_ADDRESS"
  check_contract_code "L1" "$L1_RPC_URL" "ERC20Example" "$L1_ERC20_ADDRESS"

  l2_finalized_resp="$(rpc_json "$L1_RPC_URL" eth_call "[{\"to\":\"$LINEA_ROLLUP_ADDRESS\",\"data\":\"0x695378f5\"},\"latest\"]")"
  l2_finalized_hex="$(printf '%s\n' "$l2_finalized_resp" | json_string_field result)"
  if [ -n "$l2_finalized_hex" ] && [ "$l2_finalized_hex" != "0x" ]; then
    rollup_current_l2_block_dec="$(hex_to_dec_small "$l2_finalized_hex")"
    lineth_kv "rollup currentL2BlockNumber" "$rollup_current_l2_block_dec ($l2_finalized_hex)"
    if is_uint "$rollup_current_l2_block_dec" \
      && is_uint "$local_l2_latest_block_dec" \
      && [ "$rollup_current_l2_block_dec" -gt "$local_l2_latest_block_dec" ]; then
      lineth_warn "rollup finalized block is ahead of local L2 latest block."
      lineth_warn "local chain state does not match the preserved L1 rollup state; run docker compose down -v for a clean boot."
      STATE_MISMATCH=1
    fi
  else
    lineth_warn "rollup currentL2BlockNumber unavailable"
  fi

  if [ -n "$latest_finalization_tx" ]; then
    finalization_tx_resp="$(rpc_json "$L1_RPC_URL" eth_getTransactionByHash "[\"$latest_finalization_tx\"]")"
    finalization_input="$(printf '%s\n' "$finalization_tx_resp" | json_string_field input)"
    finalization_selector="$(printf '%.10s' "$finalization_input")"
    if [ "$finalization_selector" = "0x755bc62f" ]; then
      lineth_kv "latest finalization method" "finalizeBlocks(bytes,uint256,tuple) selector=$finalization_selector"
    elif [ -n "$finalization_selector" ]; then
      lineth_warn "latest finalization method unexpected selector=$finalization_selector"
    else
      lineth_warn "latest finalization method unavailable"
    fi

    finalization_receipt_resp="$(rpc_json "$L1_RPC_URL" eth_getTransactionReceipt "[\"$latest_finalization_tx\"]")"
    finalization_status="$(printf '%s\n' "$finalization_receipt_resp" | json_string_field status)"
    case "$finalization_receipt_resp" in
      *a0262dc79e4ccb71ceac8574ae906311ae338aa4a2044fd4ec4b99fad5ab60cb*) data_finalized="yes" ;;
      *) data_finalized="no" ;;
    esac
    case "$finalization_receipt_resp" in
      *32e016ccc5c33419c35caa94023fdeb75143da613fb2ac738ab736404c09fc5d*) state_updated="yes" ;;
      *) state_updated="no" ;;
    esac
    if [ -n "$finalization_status" ]; then
      lineth_kv "latest finalization receipt" "status=$finalization_status DataFinalizedV3=$data_finalized FinalizedStateUpdated=$state_updated"
    else
      lineth_warn "latest finalization receipt unavailable"
    fi
  fi
else
  lineth_info "Sepolia rollup state check skipped: L1_RPC_URL or LineaRollupV8 unavailable"
fi

if [ "$STATE_MISMATCH" -ne 0 ]; then
  section "state mismatch action"
  lineth_error "Local L2 state and preserved Sepolia/shared state do not belong together."
  lineth_info "Stop debugging this boot; reset with:"
  lineth_info "docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v --remove-orphans"
elif [ -n "$LINEA_ROLLUP_ADDRESS" ] || [ -n "$ADDRESSES_L2_CHAIN_ID" ]; then
  section "state mismatch action"
  lineth_ok "no local/L1 state mismatch detected"
else
  section "state mismatch action"
  lineth_info "guardrails pending until addresses.json exists"
fi

section "coordinator"
if docker ps --format '{{.Names}}' | grep -qx coordinator; then
  docker exec coordinator sh -eu -c '
    if grep -qi ":2549" /proc/net/tcp /proc/net/tcp6 2>/dev/null; then echo "observability 9545: listening"; else echo "observability 9545: not listening"; fi
    if grep -qi ":254A" /proc/net/tcp /proc/net/tcp6 2>/dev/null; then echo "json-rpc 9546: listening"; else echo "json-rpc 9546: not listening"; fi
    for lane in execution compression invalidity aggregation; do
      req_dir="/data/prover/v3/$lane/requests"
      resp_dir="/data/prover/v3/$lane/responses"
      if [ -d "$req_dir" ]; then
        requests=$(find "$req_dir" -type f 2>/dev/null | wc -l | tr -d " ")
        failed=$(find "$req_dir" -type f -name "*.failure.*" 2>/dev/null | wc -l | tr -d " ")
        inprogress=$(find "$req_dir" -type f -name "*.inprogress.*" 2>/dev/null | wc -l | tr -d " ")
      else
        requests="missing"; failed="?"; inprogress="?"
      fi
      if [ -d "$resp_dir" ]; then
        responses=$(find "$resp_dir" -type f 2>/dev/null | wc -l | tr -d " ")
      else
        responses="missing"
      fi
      echo "prover $lane: requests=$requests responses=$responses failed=$failed inprogress=$inprogress"
    done
  ' | lineth_indent
  docker logs --tail 200 coordinator 2>&1 \
    | grep -E 'Rollup finalized block updated|execution proof request generated|blob compression proof generated|blobs to submit|blob submission failed|max fee per gas less than block base fee|aggregation proof|submitted|finalized|ERROR|WARN' \
    | tail -25 \
    | awk 'length($0) > 260 { $0 = substr($0, 1, 260) "..." } { print }' \
    | lineth_indent || true
else
  lineth_warn "coordinator not running"
fi

section "prover"
if docker ps --format '{{.Names}}' | grep -qx prover; then
  docker stats --no-stream --format 'prover resources: mem={{.MemUsage}} cpu={{.CPUPerc}}' prover 2>/dev/null \
    | lineth_indent || true
  docker logs --tail 120 prover 2>&1 \
    | grep -E 'ERROR|WARN|Running the|Chain config|IsAllowedCircuitID|proof|request|response|done|completed|failed' \
    | tail -25 \
    | awk 'length($0) > 260 { $0 = substr($0, 1, 260) "..." } { print }' \
    | lineth_indent || true
else
  lineth_warn "prover not running"
fi
