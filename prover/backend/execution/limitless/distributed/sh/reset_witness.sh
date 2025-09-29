#!/usr/bin/env bash
set -euo pipefail

# Usage: ./reset_witnesses.sh [--phase=gl|lpp]
# ./reset_witnesses.sh --phase=gl
# ./reset_witnesses.sh --phase=lpp


PHASE="gl"
if [[ "${1:-}" =~ ^--phase= ]]; then
    PHASE="${1#--phase=}"
    shift
fi

WITNESS_DIR="/tmp/exec-limitless/witness/${PHASE^^}"

echo "[INFO] Resetting witness markers in $WITNESS_DIR (phase=${PHASE^^})"

# Reset .success -> original
find "$WITNESS_DIR" -type f -name "*.success" | while read -r f; do
    echo "[INFO] Reverting $f -> ${f%.success}"
    mv "$f" "${f%.success}"
done

# Reset .failed -> original
find "$WITNESS_DIR" -type f -name "*.failed" | while read -r f; do
    echo "[INFO] Reverting $f -> ${f%.failed}"
    mv "$f" "${f%.failed}"
done

# Reset .INPROGRESS -> original
find "$WITNESS_DIR" -type f -name "*.INPROGRESS" | while read -r f; do
    echo "[INFO] Reverting $f -> ${f%.INPROGRESS}"
    mv "$f" "${f%.INPROGRESS}"
done

echo "[OK] Reset complete for ${PHASE^^} phase."
