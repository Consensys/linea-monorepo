#!/usr/bin/env bash
set -euo pipefail

find /tmp/witness/GL -type f -name "*.success" | while read -r f; do
    mv "$f" "${f%.success}"
done

find /tmp/witness/GL -type f -name "*.failed" | while read -r f; do
    mv "$f" "${f%.failed}"
done
