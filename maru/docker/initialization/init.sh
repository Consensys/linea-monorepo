#!/bin/zsh
echo "Checking if genesis files already exist."

if [[ ! -f "initialization/genesis-besu.json" && ! -f "initialization/genesis-geth.json" && ! -f "initialization/genesis-nethermind.json"  && ! -f "initialization/genesis-maru.json" ]]; then
    echo "Initialization of timestamp in genesis files for Maru, Besu, Geth, and Nethermind."
    date
    cd initialization || exit
    cp -T "genesis-maru.json.template" "genesis-maru.json"
    cp -T "genesis-besu.json.template" "genesis-besu.json"
    cp -T "genesis-geth.json.template" "genesis-geth.json"
    cp -T "genesis-nethermind.json.template" "genesis-nethermind.json"

    merge_timestamp=$(($(date +%s) + 60))
    echo "Timestamp: $merge_timestamp"
    sed -i "s/%PRAGUE_TIME%/$merge_timestamp/g" genesis-maru.json
    sed -i "s/%SWITCH_TIME%/$merge_timestamp/g" genesis-besu.json
    sed -i "s/%SWITCH_TIME%/$merge_timestamp/g" genesis-geth.json
    merge_timestamp_hex=$(printf "0x%x" $merge_timestamp)
    sed -i "s/%SWITCH_TIME%/$merge_timestamp_hex/g" genesis-nethermind.json
else
    echo "Genesis files already exist. Initialization skipped."
fi
