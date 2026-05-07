#!/bin/sh
# L2 genesis init for the linea-stack public quickstart.
#
# Renders genesis-maru.json + genesis-besu.json from their .template files by
# substituting:
#   - %FORK_TIME%                       — boot-time UNIX timestamp + 60s
#   - __L2_MESSAGE_SERVICE_ADDRESS__    — pre-computed by account-setup;
#                                         read from /shared/addresses-precomputed.json
#
# Idempotent guard lives in the docker-compose entrypoint (l2-genesis-init), not
# in this file — the compose layer checks for existing artifacts and skips
# invoking this script. Re-running this script unconditionally would change the
# fork timestamp and break any existing on-disk chaindata.
set -e

echo "[l2-genesis-init] rendering Maru + Besu L2 genesis"
date
cd /initialization

# ----- Pre-flight: read L2MessageService address from precomputed JSON --------
PRECOMPUTED="/shared/addresses-precomputed.json"
if [ ! -f "$PRECOMPUTED" ]; then
  echo "[l2-genesis-init] FATAL: $PRECOMPUTED not found — account-setup must run first" >&2
  exit 1
fi

# POSIX sed extraction. The JSON is written by account-setup.sh in a controlled
# shape: indented two spaces, key/value on same line, double-quoted address.
L2_MS_ADDR=$(sed -nE 's/.*"L2MessageService":[[:space:]]*"(0x[a-fA-F0-9]{40})".*/\1/p' "$PRECOMPUTED" | head -1)
if ! echo "$L2_MS_ADDR" | grep -qE '^0x[a-fA-F0-9]{40}$'; then
  echo "[l2-genesis-init] FATAL: could not extract L2MessageService address from $PRECOMPUTED" >&2
  echo "[l2-genesis-init] (got: '$L2_MS_ADDR')" >&2
  exit 1
fi
echo "[l2-genesis-init] L2_MESSAGE_SERVICE_ADDRESS=$L2_MS_ADDR"

# ----- Copy templates to working files ----------------------------------------
cp -T "genesis-maru.json.template" "genesis-maru.json"
cp -T "genesis-besu.json.template" "genesis-besu.json"

# ----- Substitute fork timestamp (both files) ---------------------------------
fork_timestamp=$(($(date +%s) + 60))
echo "[l2-genesis-init] FORK_TIME=$fork_timestamp"
sed -i "s/%FORK_TIME%/$fork_timestamp/g" genesis-maru.json
sed -i "s/%FORK_TIME%/$fork_timestamp/g" genesis-besu.json

# ----- Substitute L2MessageService address (besu genesis only) ----------------
sed -i "s|__L2_MESSAGE_SERVICE_ADDRESS__|$L2_MS_ADDR|g" genesis-besu.json

# Sanity: ensure no placeholder remains in the rendered genesis.
if grep -q "__L2_MESSAGE_SERVICE_ADDRESS__" genesis-besu.json; then
  echo "[l2-genesis-init] FATAL: placeholder __L2_MESSAGE_SERVICE_ADDRESS__ still present after substitution" >&2
  exit 1
fi

echo "$fork_timestamp" > /initialization/fork-timestamp.txt
echo "[l2-genesis-init] done"
