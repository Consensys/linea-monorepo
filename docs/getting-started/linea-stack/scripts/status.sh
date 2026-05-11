#!/usr/bin/env sh
# Print a redacted, milestone-oriented status report for the Linea quickstart.
# Run from docs/getting-started/linea-stack while the stack is booting.
set -eu

section() { printf '\n[linea-status] %s\n' "$*"; }

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

section "coordinator"
if docker ps --format '{{.Names}}' | grep -qx coordinator; then
  docker exec coordinator sh -eu -c '
    if grep -qi ":2549" /proc/net/tcp /proc/net/tcp6 2>/dev/null; then echo "observability 9545: listening"; else echo "observability 9545: not listening"; fi
    if grep -qi ":254A" /proc/net/tcp /proc/net/tcp6 2>/dev/null; then echo "json-rpc 9546: listening"; else echo "json-rpc 9546: not listening"; fi
    for lane in execution compression aggregation; do
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
  docker logs --tail 120 prover 2>&1 \
    | grep -E 'ERROR|WARN|proof|request|response|done|completed|failed' \
    | tail -25 \
    | awk 'length($0) > 260 { $0 = substr($0, 1, 260) "..." } { print }' || true
else
  echo "prover: not running"
fi
