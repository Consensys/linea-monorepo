#!/usr/bin/env bash
set -euo pipefail

WITNESS_GL_DIR="/tmp/exec-limitless/witness/GL"

find ${WITNESS_GL_DIR} -type f -name "*.success" | while read -r f; do
    mv "$f" "${f%.success}"
done

find ${WITNESS_GL_DIR} -type f -name "*.failed" | while read -r f; do
    mv "$f" "${f%.failed}"
done
