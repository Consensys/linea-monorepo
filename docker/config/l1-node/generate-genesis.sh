#!/bin/bash
set -euo pipefail

genesis_time=""
l1_genesis=""
network_config=""
mnemonics=""
output_dir=""

usage() {
  echo "Usage: $0 --genesis-time <timestamp> --l1-genesis <path to l1 genesis file> --network-config <path to network config file> --mnemonics <path to mnemonics file> --output-dir <output directory>"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --genesis-time)
      genesis_time="$2"
      shift 2
      ;;
    --l1-genesis)
      l1_genesis="$2"
      shift 2
      ;;
    --network-config)
      network_config="$2"
      shift 2
      ;;
    --mnemonics)
      mnemonics="$2"
      shift 2
      ;;
    --output-dir)
      output_dir="$2"
      shift 2
      ;;
    *)
      echo "Error: Unknown option: $1"
      usage
      ;;
  esac
done

if [ -z "$genesis_time" ] || [ -z "$l1_genesis" ] || [ -z "$network_config" ] || [ -z "$mnemonics" ] || [ -z "$output_dir" ]; then
  echo "Error: Missing required argument."
  usage
fi

echo "Genesis time set to: $genesis_time"
# Hacky workaround for limitation that Teku cannot begin with genesis state of Electra (Prague)
prague_time=$((genesis_time + 32))
echo "Prague time set to: $prague_time"

mkdir -p $output_dir
cp $l1_genesis $output_dir/genesis.json
cp $network_config $output_dir/$(basename -- $network_config)

sed -i -E 's/"timestamp": "[0-9]+"/"timestamp": "'"$genesis_time"'"/' $output_dir/genesis.json
# sed -i -E 's/"pragueTime": 0/"pragueTime": '"$prague_time"'/' $output_dir/genesis.json
sed -i 's/\$GENESIS_TIME/'"$genesis_time"'/g' $output_dir/$(basename -- $network_config)

/usr/local/bin/eth2-testnet-genesis deneb --config $output_dir/$(basename -- $network_config) --mnemonics $mnemonics --tranches-dir $output_dir/tranches --state-output $output_dir/genesis.ssz --eth1-config $output_dir/genesis.json
