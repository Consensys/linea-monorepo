#!/usr/bin/env bash
set -euo pipefail

# Usage: ./run_gl_proofs.sh [max_parallel]
MAX_PARALLEL=${1:-4}

# Set your absolute paths
WITNESS_DIR="/tmp/witness/GL"
RESPONSES_DIR="/tmp/responses"
LOG_BASE="/home/ubuntu/linea-monorepo/prover/backend/execution/limitless/logs/mainnet/gl"
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

    logdir="$LOG_BASE"
    mkdir -p "$logdir"

    seg_mod=$(echo "$base" | grep -oE "seg-[0-9]+-mod-[0-9]+" || true)
    seg_mod="${seg_mod//-/}"
    logfile="$logdir/${seg_mod:-unknown}.log"

    echo "[INFO] Claimed $infile -> processing $inprogress"

    # Build the command
    local out_file="$RESPONSES_DIR/${base}-subProof.json"
    local cmd=(
        "$PROVER" prove
        --phase=gl
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
export WITNESS_DIR RESPONSES_DIR LOG_BASE CONFIG PROVER

# Collect candidate witness files (exclude .success/.failed/.INPROGRESS)
mapfile -t files < <(find "$WITNESS_DIR" -type f -name "*-gl-wit.bin" ! -name "*.success" ! -name "*.failed" ! -name "*.INPROGRESS")

if [ "${#files[@]}" -eq 0 ]; then
    echo "[INFO] All GL witness files already processed. Nothing to do."
    exit 0
fi

# Process with parallelism
printf '%s\0' "${files[@]}" \
    | xargs -0 -n1 -P"$MAX_PARALLEL" bash -c 'run_proof "$@"' _
