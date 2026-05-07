#!/bin/sh
# L2 genesis init for the linea-stack public quickstart.
#
# Lifted from docker/config/l2-genesis-initialization/init.sh in the
# linea-monorepo, with two adaptations:
#   1. Shebang: /bin/sh (busybox-compatible) instead of /bin/zsh.
#   2. The internal version copies coordinator-config-v2.toml from a
#      mounted /coordinator/ path to render coordinator-config-v2-hardforks.toml.
#      The public quickstart's coordinator service mounts its config directly
#      from config/l2/coordinator/, so that copy step is not needed here.
#
# Idempotent guard lives in the docker-compose entrypoint, not in this file —
# the compose layer checks for existing artifacts and skips invoking this
# script. Re-running this script unconditionally would change the fork
# timestamp and break any existing on-disk chaindata.
set -e

echo "Initialization of timestamp in genesis files for Maru and Besu"
date
cd /initialization

cp -T "genesis-maru.json.template" "genesis-maru.json"
cp -T "genesis-besu.json.template" "genesis-besu.json"

fork_timestamp=$(($(date +%s) + 60))
echo "Fork Timestamp: $fork_timestamp"
sed -i "s/%FORK_TIME%/$fork_timestamp/g" genesis-maru.json
sed -i "s/%FORK_TIME%/$fork_timestamp/g" genesis-besu.json

echo "$fork_timestamp" > /initialization/fork-timestamp.txt
