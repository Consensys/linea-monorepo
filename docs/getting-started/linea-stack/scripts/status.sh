#!/usr/bin/env sh
# Print a redacted, milestone-oriented status report for the Linea quickstart.
# Run from docs/getting-started/linea-stack while the stack is booting.
set -eu

section() { printf '\n[linea-status] %s\n' "$*"; }

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

shared_addr() {
  file="$1"
  section_name="$2"
  key="$3"
  docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c "
    [ -f /shared/$file ] || exit 0
    sed -nE '/\"$section_name\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p' /shared/$file | head -1
  " 2>/dev/null || true
}

if ! docker info >/dev/null 2>&1; then
  echo "[linea-status] ERROR: Docker daemon is not reachable from this shell." >&2
  exit 1
fi

section "containers"
docker ps -a --format 'table {{.Names}}\t{{.Status}}' \
  | awk 'NR == 1 || /^(account-setup|config-render|l2-genesis-init|sequencer|l2-node-besu|shomei|deploy-contracts|coordinator|prover|postman|web3signer|maru|l2-blockscout)[[:space:]]/'

section "deploy milestones"
if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  echo "shared volume missing: deploy has not started"
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
  '
fi

section "chain / prover config"
HOST_PORT_L2_RPC="${HOST_PORT_L2_RPC:-$(env_value HOST_PORT_L2_RPC || true)}"
[ -n "$HOST_PORT_L2_RPC" ] || HOST_PORT_L2_RPC=8745
L2_RPC_URL="${L2_RPC_URL:-http://localhost:$HOST_PORT_L2_RPC}"
chain_response="$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  "$L2_RPC_URL" 2>/dev/null || true)"
if [ -n "$chain_response" ]; then
  echo "l2 rpc eth_chainId: $chain_response"
else
  echo "l2 rpc eth_chainId: unavailable at $L2_RPC_URL"
fi

if docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  docker run --rm -v linea-stack-shared-config:/shared:ro busybox sh -eu -c '
    if [ -f /shared/addresses.json ]; then
      sed -nE "s/.*\"l2ChainId\":[[:space:]]*\"?([0-9]+)\"?.*/addresses.json l2ChainId: \1/p" /shared/addresses.json | head -1
    fi
  '
fi

if docker volume inspect linea-stack-rendered-config >/dev/null 2>&1; then
  docker run --rm -v linea-stack-rendered-config:/rendered:ro busybox sh -eu -c '
    if [ -f /rendered/prover-config-partial.toml ]; then
      grep -E "^chain_id|^prover_mode|^is_allowed_circuit_id" /rendered/prover-config-partial.toml
    else
      echo "rendered prover config: missing"
    fi
  '
else
  echo "rendered config volume missing"
fi

section "l1 data availability vs finalization"
LINEA_ROLLUP_ADDRESS=""
if docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
  LINEA_ROLLUP_ADDRESS="$(shared_addr addresses.json l1 LineaRollupV8)"
fi

if [ -n "$LINEA_ROLLUP_ADDRESS" ]; then
  echo "LineaRollupV8: $LINEA_ROLLUP_ADDRESS"
else
  echo "LineaRollupV8: unavailable until addresses.json exists"
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
  echo "latest blob tx (DA only): $latest_blob_tx"
else
  echo "latest blob tx (DA only): none seen in coordinator logs yet"
fi

if [ -n "$latest_finalization_tx" ]; then
  if [ -n "$latest_finalization_window" ]; then
    echo "latest finalization tx: $latest_finalization_tx aggregation=$latest_finalization_window"
  else
    echo "latest finalization tx: $latest_finalization_tx"
  fi
else
  echo "latest finalization tx: none seen in coordinator logs yet"
fi

L1_RPC_URL="${L1_RPC_URL:-$(env_value L1_RPC_URL || true)}"
if [ -n "$L1_RPC_URL" ] && [ -n "$LINEA_ROLLUP_ADDRESS" ]; then
  l2_finalized_resp="$(curl -sS -X POST -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"eth_call\",\"params\":[{\"to\":\"$LINEA_ROLLUP_ADDRESS\",\"data\":\"0x695378f5\"},\"latest\"]}" \
    "$L1_RPC_URL" 2>/dev/null || true)"
  l2_finalized_hex="$(printf '%s\n' "$l2_finalized_resp" | json_string_field result)"
  if [ -n "$l2_finalized_hex" ] && [ "$l2_finalized_hex" != "0x" ]; then
    echo "rollup currentL2BlockNumber: $(hex_to_dec_small "$l2_finalized_hex") ($l2_finalized_hex)"
  else
    echo "rollup currentL2BlockNumber: unavailable"
  fi

  if [ -n "$latest_finalization_tx" ]; then
    finalization_tx_resp="$(curl -sS -X POST -H "Content-Type: application/json" \
      -d "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"eth_getTransactionByHash\",\"params\":[\"$latest_finalization_tx\"]}" \
      "$L1_RPC_URL" 2>/dev/null || true)"
    finalization_input="$(printf '%s\n' "$finalization_tx_resp" | json_string_field input)"
    finalization_selector="$(printf '%.10s' "$finalization_input")"
    if [ "$finalization_selector" = "0x755bc62f" ]; then
      echo "latest finalization method: finalizeBlocks(bytes,uint256,tuple) selector=$finalization_selector"
    elif [ -n "$finalization_selector" ]; then
      echo "latest finalization method: unexpected selector=$finalization_selector"
    else
      echo "latest finalization method: unavailable"
    fi

    finalization_receipt_resp="$(curl -sS -X POST -H "Content-Type: application/json" \
      -d "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"eth_getTransactionReceipt\",\"params\":[\"$latest_finalization_tx\"]}" \
      "$L1_RPC_URL" 2>/dev/null || true)"
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
      echo "latest finalization receipt: status=$finalization_status DataFinalizedV3=$data_finalized FinalizedStateUpdated=$state_updated"
    else
      echo "latest finalization receipt: unavailable"
    fi
  fi
else
  echo "Sepolia rollup state check: skipped (L1_RPC_URL or LineaRollupV8 unavailable)"
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
  '
  docker logs --tail 200 coordinator 2>&1 \
    | grep -E 'Rollup finalized block updated|execution proof request generated|blob compression proof generated|blobs to submit|aggregation proof|submitted|finalized|ERROR|WARN' \
    | tail -25 \
    | awk 'length($0) > 260 { $0 = substr($0, 1, 260) "..." } { print }' || true
else
  echo "coordinator: not running"
fi

section "prover"
if docker ps --format '{{.Names}}' | grep -qx prover; then
  docker stats --no-stream --format 'prover resources: mem={{.MemUsage}} cpu={{.CPUPerc}}' prover 2>/dev/null || true
  docker logs --tail 120 prover 2>&1 \
    | grep -E 'ERROR|WARN|Running the|Chain config|IsAllowedCircuitID|proof|request|response|done|completed|failed' \
    | tail -25 \
    | awk 'length($0) > 260 { $0 = substr($0, 1, 260) "..." } { print }' || true
else
  echo "prover: not running"
fi
