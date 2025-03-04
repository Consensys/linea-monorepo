#!/bin/zsh
echo "Initialization of timestamp in genesis files for Besu and Geth."
date
cd initialization
cp -T "genesis-besu.json.template" "genesis-besu.json"
cp -T "genesis-geth.json.template" "genesis-geth.json"
cp -T "genesis-nethermind.json.template" "genesis-nethermind.json"

merge_timestamp=$(($(date +%s) + 60))
sed -i "s/^    \"shanghaiTime\": .*/    "\"shanghaiTime\"": $merge_timestamp,/" genesis-besu.json
sed -i "s/^    \"cancunTime\": .*/    "\"cancunTime\"": $merge_timestamp,/" genesis-besu.json
sed -i "s/^    \"shanghaiTime\": .*/    "\"shanghaiTime\"": $merge_timestamp,/" genesis-geth.json
sed -i "s/^    \"cancunTime\": .*/    "\"cancunTime\"": $merge_timestamp,/" genesis-geth.json
merge_timestamp_hex=$(printf "0x%x" $merge_timestamp)
sed -i "s/%SWITCH_TIME%/$merge_timestamp_hex/g" genesis-nethermind.json
