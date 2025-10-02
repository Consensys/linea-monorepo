#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./reset.sh -witness --phase=gl
#   ./reset.sh -subproofs --phase=lpp

PHASE=""
TARGET=""

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        -witness)
            TARGET="witness"
            shift
            ;;
        -subproofs)
            TARGET="subproofs"
            shift
            ;;
        --phase=*)
            PHASE="${1#--phase=}"
            shift
            ;;
        *)
            echo "Unknown arg: $1"
            echo "Usage: $0 [-witness|-subproofs] --phase=gl|lpp"
            exit 1
            ;;
    esac
done

if [[ -z "$TARGET" || -z "$PHASE" ]]; then
    echo "Error: must specify target (-witness|-subproofs) and --phase"
    exit 1
fi

TARGET_DIR="/tmp/exec-limitless/${TARGET}/${PHASE^^}"

reset_markers() {
    local DIR="$1"
    if [[ ! -d "$DIR" ]]; then
        echo "[WARN] Skipping $DIR (does not exist)"
        return
    fi

    echo "[INFO] Resetting markers in $DIR (phase=${PHASE^^})"

    for suffix in success failed INPROGRESS; do
        find "$DIR" -type f -name "*.${suffix}" | while read -r f; do
            echo "[INFO] Reverting $f -> ${f%.${suffix}}"
            mv "$f" "${f%.${suffix}}"
        done
    done

    echo "[OK] Reset complete for $DIR"
}

reset_markers "$TARGET_DIR"
