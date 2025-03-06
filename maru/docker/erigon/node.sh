#!/usr/bin/env bash
set -Eeu

if [ $# -lt 4 ]; then
    echo "Usage: $0 networkid genesisfile gaslimit datadir"
    echo Example: "$0" "1337", "/initialization/genesis-geth.json", "0x1C9C380"
    exit
fi

source /scripts/functions.sh
networkid=$1
genesisfile=$2
gaslimit=$3
datadir=$4

: ${ETHSTATS_URL:=""} #default to empty

echo datadir=$datadir
echo networkid=$networkid
echo genesisfile=$genesisfile
echo gaslimit=$gaslimit

prepareDatadir $datadir $genesisfile

echo "Starting erigon with BOOTNODE enabled on port '30301'"
exec erigon --datadir $datadir \
 --networkid $networkid \
 --miner.gaslimit $gaslimit \
 --http --http.addr '0.0.0.0' --http.port 8545 --http.corsdomain '*' --http.api 'admin,engine,eth,net,web3,txpool,debug' --http.vhosts="*" \
 --port 30301 \
 --ethstats "$ETHSTATS_URL" \
 --bootnodes $BOOTNODES \
 --netrestrict "11.11.11.0/24" \
 --authrpc.port 8551 \
 --authrpc.addr=0.0.0.0 \
 --authrpc.vhosts=* \
 --authrpc.jwtsecret /jwt
