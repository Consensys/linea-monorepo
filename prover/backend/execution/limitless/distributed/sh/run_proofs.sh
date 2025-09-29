#!/usr/bin/env bash
set -euo pipefail

# Usage: ./run_proofs.sh [--phase=gl|lpp] [max_parallel]
# ./run_proofs.sh --phase=gl 4
# ./run_proofs.sh --phase=lpp 2


PHASE="gl"
if [[ "${1:-}" =~ ^--phase= ]]; then
    PHASE="${1#--phase=}"
    shift
fi

MAX_PARALLEL=${1:-4}

# Paths (adjust as needed)
WITNESS_DIR="/tmp/exec-limitless/witness/${PHASE^^}" # uppercase phase in path
RESPONSES_DIR="/tmp/responses"
LOG_BASE="/home/ubuntu/linea-monorepo/prover/backend/execution/limitless/logs/mainnet/${PHASE}"
CONFIG="/home/ubuntu/linea-monorepo/prover/config/config-mainnet-limitless.toml"
PROVER="/home/ubuntu/linea-monorepo/prover/bin/prover"

mkdir -p "$RESPONSES_DIR" "$LOG_BASE"

run_proof() {
    local infile="$1"
    local inprogress="${infile}.INPROGRESS"

    if ! mv -- "$infile" "$inprogress" 2>/dev/null; then
        echo "[WARN] Could not claim $infile, skipping."
        return 0
    fi

    relpath="${inprogress#$WITNESS_DIR/}"
    subdir="$(dirname "$relpath")"
    name="$(basename "$inprogress")"
    name_no_inprog="${name%.INPROGRESS}"
    base="${name_no_inprog%-wit.bin}"

    mkdir -p "$LOG_BASE"
    seg_mod=$(echo "$base" | grep -oE "seg-[0-9]+-mod-[0-9]+" || true)
    seg_mod="${seg_mod//-/}"
    logfile="$LOG_BASE/${seg_mod:-unknown}.log"

    echo "[INFO] Claimed $infile -> processing $inprogress"

    # Output file (currently discarded, since orchestration writes .json later)
    local out_file="/dev/null"

    # Build the command
    local cmd=(
        "$PROVER" prove
        --phase="$PHASE"
        --config "$CONFIG"
        --in "$inprogress"
        --out "$out_file"
    )

    echo "[INFO] Command: ${cmd[*]}"
    echo "[INFO] Command: ${cmd[*]}" >>"$logfile"

    local start_ts end_ts duration
    start_ts=$(date +%s)

    if "${cmd[@]}" >"$logfile" 2>&1; then
        mv -- "$inprogress" "${infile}.success" || \
            echo "[WARN] Completed but could not rename $inprogress -> ${infile}.success"
        end_ts=$(date +%s)
        duration=$((end_ts - start_ts))
        echo "[OK] Completed $infile in ${duration}s"
        echo "[OK] Completed $infile in ${duration}s" >>"$logfile"
    else
        mv -- "$inprogress" "${infile}.failed" || \
            echo "[ERROR] Failed proof and could not rename $inprogress -> ${infile}.failed"
        end_ts=$(date +%s)
        duration=$((end_ts - start_ts))
        echo "[ERROR] Failed proof for $infile in ${duration}s (log: $logfile)"
        echo "[ERROR] Failed proof for $infile in ${duration}s" >>"$logfile"
        return 1
    fi
}

export -f run_proof
export WITNESS_DIR RESPONSES_DIR LOG_BASE CONFIG PROVER PHASE

# Collect candidate witness files
mapfile -t files < <(find "$WITNESS_DIR" -type f -name "*-wit.bin" ! -name "*.success" ! -name "*.failed" ! -name "*.INPROGRESS")

if [ "${#files[@]}" -eq 0 ]; then
    echo "[INFO] All ${PHASE^^} witness files already processed. Nothing to do."
    exit 0
fi

# Process with parallelism
printf '%s\0' "${files[@]}" \
    | xargs -0 -n1 -P"$MAX_PARALLEL" bash -c 'run_proof "$@"' _
