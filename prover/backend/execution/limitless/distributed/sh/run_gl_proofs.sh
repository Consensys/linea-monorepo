#!/usr/bin/env bash
set -euo pipefail

# Usage: ./backend/execution/limitless/distributed/sh/run_gl_proofs.sh 4
# Default concurrency = 4
MAX_PARALLEL=${1:-4}

# Set your paths here
WITNESS_DIR="/tmp/witness/GL"
OUTPUT_BASE="/tmp/subproof/GL"
LOG_BASE="/home/ubuntu/linea-monorepo/prover/backend/execution/limitless/logs/mainnet/gl"
CONFIG="/home/ubuntu/linea-monorepo/prover/config/config-mainnet-limitless.toml"
PROVER="/home/ubuntu/linea-monorepo/prover/bin/prover"

mkdir -p "$OUTPUT_BASE" "$LOG_BASE"

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

    outdir="$OUTPUT_BASE/$subdir"
    mkdir -p "$outdir"

    logdir="$LOG_BASE"
    mkdir -p "$logdir"

    seg_mod=$(echo "$base" | grep -oE "seg-[0-9]+-mod-[0-9]+" || true)
    seg_mod="${seg_mod//-/}"
    logfile="$logdir/${seg_mod:-unknown}.log"

    echo "[INFO] Claimed $infile -> processing $inprogress"
    echo "[INFO] Running proof for $inprogress -> $outdir/${base}-proof.bin (log: $logfile)"

    if "$PROVER" prove \
        --phase=gl \
        --config "$CONFIG" \
        --in "$inprogress" \
        --out "$outdir/${base}-proof.bin" \
        >"$logfile" 2>&1; then
        mv -- "$inprogress" "${infile}.success" || \
            echo "[WARN] Completed but could not rename $inprogress -> ${infile}.success"
        echo "[OK] Completed $infile"
    else
        mv -- "$inprogress" "${infile}.failed" || \
            echo "[ERROR] Failed proof and could not rename $inprogress -> ${infile}.failed"
        echo "[ERROR] Failed proof for $infile (log: $logfile)"
    fi
}

export -f run_proof
export WITNESS_DIR OUTPUT_BASE LOG_BASE CONFIG PROVER

# Collect candidate witness files (those not already marked .success/.failed/.INPROGRESS)
mapfile -t files < <(find "$WITNESS_DIR" -type f -name "*-gl-wit.bin" ! -name "*.success" ! -name "*.failed" ! -name "*.INPROGRESS")

if [ "${#files[@]}" -eq 0 ]; then
    echo "[INFO] All witness files already processed. Nothing to do."
    exit 0
fi

# Process with parallelism
printf '%s\0' "${files[@]}" \
    | xargs -0 -n1 -P"$MAX_PARALLEL" bash -c 'run_proof "$@"' _
