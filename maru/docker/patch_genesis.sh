#!/bin/zsh
merge_timestamp=$(date -v+1M +"%s")
sed -i '' "s/^    \"shanghaiTime\": .*/    "\"shanghaiTime\"": $merge_timestamp,/" "$1"
sed -i '' "s/^    \"cancunTime\": .*/    "\"cancunTime\"": $merge_timestamp,/" "$1"
sed -i '' "s/^    \"shanghaiTime\": .*/    "\"shanghaiTime\"": $merge_timestamp,/" "$2"
sed -i '' "s/^    \"cancunTime\": .*/    "\"cancunTime\"": $merge_timestamp,/" "$2"