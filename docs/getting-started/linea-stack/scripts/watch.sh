#!/usr/bin/env sh
# Guided progress view for the Lineth quickstart boot/finality pipeline.
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
LINETH_LOG_CONTEXT="watch"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"

WATCH_INTERVAL_SECONDS="${WATCH_INTERVAL_SECONDS:-10}"
WATCH_TIMEOUT_SECONDS="${WATCH_TIMEOUT_SECONDS:-1800}"
WATCH_SINCE="${WATCH_SINCE:-2h}"
WATCH_MODE="until-finalized"

usage() {
  cat <<'EOF'
Usage: ./scripts/watch.sh [--once|--tail]

  --once   print one progress snapshot and exit
  --tail   keep printing progress snapshots until Ctrl-C

Environment:
  WATCH_INTERVAL_SECONDS   seconds between snapshots in follow mode (default: 10)
  WATCH_TIMEOUT_SECONDS    max seconds before watch exits in default mode (default: 1800)
  WATCH_SINCE              docker log window to inspect (default: 2h)
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --once)
      WATCH_MODE="once"
      ;;
    --tail)
      WATCH_MODE="tail"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      lineth_die "unknown argument: $1"
      ;;
  esac
  shift
done

is_uint() {
  case "$1" in
    ''|*[!0-9]*) return 1 ;;
    *) return 0 ;;
  esac
}

container_state() {
  name="$1"
  docker inspect -f '{{.State.Status}} {{.State.ExitCode}}' "$name" 2>/dev/null || printf 'missing\n'
}

shared_file_status() {
  file="$1"
  if ! docker volume inspect linea-stack-shared-config >/dev/null 2>&1; then
    printf 'missing volume'
    return
  fi

  if docker run --rm -v linea-stack-shared-config:/shared:ro busybox test -f "/shared/$file" >/dev/null 2>&1; then
    printf 'present'
  else
    printf 'missing'
  fi
}

latest_finalized_block() {
  docker logs --since "$WATCH_SINCE" coordinator 2>&1 \
    | sed -nE 's/.*finalization update: previousFinalizedBlock=[0-9]+ newFinalizedBlock=FinalizationUpdate\(blockNumber=([0-9]+),.*/\1/p' \
    | tail -1
}

latest_finalization_tx() {
  docker logs --since "$WATCH_SINCE" coordinator 2>&1 \
    | sed -nE 's/.*submitted aggregation=\[([0-9]+)\.\.([0-9]+)\][^ ]* transactionHash=(0x[a-fA-F0-9]{64}).*/range=[\1..\2] tx=\3/p' \
    | tail -1
}

latest_blob_tx() {
  docker logs --since "$WATCH_SINCE" coordinator 2>&1 \
    | sed -nE 's/.*blobs submitted: blobs=\[\[([0-9]+)\.\.([0-9]+)\][^ ]* transactionHash=(0x[a-fA-F0-9]{64}).*/range=[\1..\2] tx=\3/p' \
    | tail -1
}

deploy_summary() {
  state="$(container_state deploy-contracts)"
  case "$state" in
    "missing")
      lineth_warn "deploy-contracts container not found"
      ;;
    "running "*)
      lineth_kv "deploy-contracts" "running"
      ;;
    "exited 0")
      lineth_kv "deploy-contracts" "exited 0"
      ;;
    "exited "*)
      lineth_error "deploy-contracts $state"
      ;;
    *)
      lineth_kv "deploy-contracts" "$state"
      ;;
  esac

  lineth_kv "addresses-precomputed.json" "$(shared_file_status addresses-precomputed.json)"
  lineth_kv "addresses.json" "$(shared_file_status addresses.json)"

  if docker inspect deploy-contracts >/dev/null 2>&1; then
    docker logs --tail 120 deploy-contracts 2>&1 \
      | sed -nE '
          s/.*\[deploy-contracts\] ===== ([^=].*[^[:space:]]) =====.*/phase: \1/p
          s/^contract=([^ ]+) deployed: address=(0x[a-fA-F0-9]{40}).*blockNumber=([0-9]+).*/contract: \1 block=\3 address=\2/p
          s/.*\[deploy-contracts\] (verify .*: OK.*)/\1/p
          s/.*\[deploy-contracts\] (wrote \/shared\/addresses.json|Patched coordinator-config.toml|Done).*/\1/p
        ' \
      | tail -12 \
      | lineth_indent
  fi
}

pipeline_timeline() {
  if ! docker inspect coordinator >/dev/null 2>&1; then
    lineth_warn "coordinator container not found"
    return
  fi

  events="$(docker logs --since "$WATCH_SINCE" coordinator 2>&1 \
    | sed -nE '
        s/^time=([^ ]+) .*message=Started :\).*/\1 coordinator ready/p
        s/^time=([^ ]+) .*message=new batch: batch=\[([0-9]+)\.\.([0-9]+)\].*/\1 batch             [\2..\3]/p
        s/^time=([^ ]+) .*message=new blob: blob=\[([0-9]+)\.\.([0-9]+)\].*/\1 blob candidate    [\2..\3]/p
        s/^time=([^ ]+) .*message=execution proof request generated: proofIndex=ExecutionProofIndex\(startBlockNumber=([0-9]+), endBlockNumber=([0-9]+).*/\1 proof request     [\2..\3]/p
        s/^time=([^ ]+) .*message=execution proof generated: batch=Batch\(startBlockNumber=([0-9]+), endBlockNumber=([0-9]+)\).*/\1 execution proof   [\2..\3]/p
        s/^time=([^ ]+) .*message=blob compression proof generated: blob=\[([0-9]+)\.\.([0-9]+)\].*/\1 compression proof [\2..\3]/p
        s/^time=([^ ]+) .*message=aggregation proof generated: aggregation=\[([0-9]+)\.\.([0-9]+)\].*/\1 aggregation proof [\2..\3]/p
        s/^time=([^ ]+) .*message=eth_call for blob submission failed: blob=.*\[([0-9]+)\.\.([0-9]+)\].*reason: '\''max fee per gas less than block base fee'\''.*logger=BlobSubmissionCoordinator.*/\1 blob gas cap     [\2..\3] max fee < base fee/p
        s/^time=([^ ]+) .*message=blobs submitted: blobs=\[\[([0-9]+)\.\.([0-9]+)\][^ ]* transactionHash=(0x[a-fA-F0-9]{64}).*/\1 blob submitted    [\2..\3] tx=\4/p
        s/^time=([^ ]+) .*message=submitted aggregation=\[([0-9]+)\.\.([0-9]+)\][^ ]* transactionHash=(0x[a-fA-F0-9]{64}).*/\1 finalize tx       [\2..\3] tx=\4/p
        s/^time=([^ ]+) .*message=finalization update: previousFinalizedBlock=([0-9]+) newFinalizedBlock=FinalizationUpdate\(blockNumber=([0-9]+),.*/\1 finalized        L2 \2 -> \3/p
      ' \
    | tail -45)"

  if [ -n "$events" ]; then
    printf '%s\n' "$events" | lineth_indent
  else
    lineth_info "no coordinator pipeline events yet"
  fi
}

retry_summary() {
  if ! docker inspect coordinator >/dev/null 2>&1; then
    return
  fi

  root_mismatch_count="$(docker logs --since "$WATCH_SINCE" coordinator 2>&1 | grep -c 'StartingRootHashDoesNotMatch' || true)"
  already_known_count="$(docker logs --since "$WATCH_SINCE" coordinator 2>&1 | grep -c 'already known' || true)"
  nonce_low_count="$(docker logs --since "$WATCH_SINCE" coordinator 2>&1 | grep -c 'nonce too low' || true)"
  fee_history_count="$(docker logs --since "$WATCH_SINCE" coordinator 2>&1 | grep -c 'Not enough fee history data' || true)"

  if [ "$root_mismatch_count" -gt 0 ] || [ "$already_known_count" -gt 0 ] || [ "$nonce_low_count" -gt 0 ] || [ "$fee_history_count" -gt 0 ]; then
    lineth_kv "root mismatch retries" "$root_mismatch_count"
    lineth_kv "already-known retries" "$already_known_count"
    lineth_kv "nonce-too-low retries" "$nonce_low_count"
    lineth_kv "fee-history warmup warnings" "$fee_history_count"
    lineth_info "retry noise is only a blocker if finalized L2 block stops advancing"
  else
    lineth_info "no coordinator retry noise seen in the selected log window"
  fi
}

finality_summary() {
  finalized="$(latest_finalized_block || true)"
  blob="$(latest_blob_tx || true)"
  finalization="$(latest_finalization_tx || true)"

  if [ -n "$blob" ]; then
    lineth_kv "latest blob tx" "$blob"
  else
    lineth_kv "latest blob tx" "none seen yet"
  fi

  if [ -n "$finalization" ]; then
    lineth_kv "latest finalization tx" "$finalization"
  else
    lineth_kv "latest finalization tx" "none seen yet"
  fi

  if [ -n "$finalized" ]; then
    lineth_kv "finalized L2 block" "$finalized"
  else
    lineth_kv "finalized L2 block" "0 or unavailable"
  fi
}

snapshot() {
  LINETH_SECTION_INDEX=0
  lineth_info "snapshot $(date '+%H:%M:%S')"

  lineth_section "deploy"
  deploy_summary

  lineth_section "pipeline timeline"
  pipeline_timeline

  lineth_section "retry noise"
  retry_summary

  lineth_section "finality"
  finality_summary

  lineth_section "next command"
  finalized="$(latest_finalized_block || true)"
  if [ -n "$finalized" ] && is_uint "$finalized" && [ "$finalized" -gt 0 ]; then
    lineth_ok "first L1 finalization observed; run ./scripts/export-output.sh when ready"
  else
    lineth_info "waiting for first L1 finalization; keep this watcher running"
  fi
}

if ! docker info >/dev/null 2>&1; then
  lineth_die "Docker daemon is not reachable from this shell."
fi

case "$WATCH_INTERVAL_SECONDS" in
  ''|*[!0-9]*) lineth_die "WATCH_INTERVAL_SECONDS must be a positive integer" ;;
esac
case "$WATCH_TIMEOUT_SECONDS" in
  ''|*[!0-9]*) lineth_die "WATCH_TIMEOUT_SECONDS must be a positive integer" ;;
esac

lineth_banner "watch · deploy + proof/finality timeline"

start_ts="$(date +%s)"
while :; do
  snapshot

  finalized="$(latest_finalized_block || true)"
  if [ "$WATCH_MODE" = "once" ]; then
    exit 0
  fi
  if [ "$WATCH_MODE" = "until-finalized" ] && [ -n "$finalized" ] && is_uint "$finalized" && [ "$finalized" -gt 0 ]; then
    exit 0
  fi

  elapsed=$(( $(date +%s) - start_ts ))
  if [ "$WATCH_MODE" = "until-finalized" ] && [ "$elapsed" -ge "$WATCH_TIMEOUT_SECONDS" ]; then
    lineth_warn "watch timed out before first finalization; inspect ./scripts/status.sh and coordinator/prover logs"
    exit 1
  fi

  sleep "$WATCH_INTERVAL_SECONDS"
done
