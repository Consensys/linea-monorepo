#!/bin/zsh
echo "Initialization of timestamp in genesis files for Maru, Besu, and config file for coordinator"
date
cd initialization || exit
cp -T "genesis-maru.json.template" "genesis-maru.json"
cp -T "genesis-besu.json.template" "genesis-besu.json"
cp -T "/coordinator/coordinator-config-v2.toml" "coordinator-config-v2-hardforks.toml"

fork_timestamp=$(($(date +%s) + 60))
fork_timestamp_hex="0x$(printf '%x' $fork_timestamp)"
echo "Fork Timestamp: $fork_timestamp ($fork_timestamp_hex)"
sed -i "s/%FORK_TIME%/$fork_timestamp/g" genesis-maru.json
sed -i "s/%FORK_TIME%/$fork_timestamp/g" genesis-besu.json
sed -i "s/%FORK_TIME_HEX%/$fork_timestamp_hex/g" genesis-besu.json

sed -i'' "s/^\(timestamp-based-hard-forks[ ]*=[ ]*\).*/\1[${fork_timestamp}]/" coordinator-config-v2-hardforks.toml

echo $fork_timestamp > /initialization/fork-timestamp.txt
