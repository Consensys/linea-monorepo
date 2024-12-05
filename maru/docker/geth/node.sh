#!/usr/bin/env bash
set -Eeu

if [ $# -lt 4 ]; then
    echo "Usage: $0 networkid genesisfile gasprice gaslimit datadir"
    echo Example: $0 12345 /genesis.json
    exit
fi

source ./scripts/functions.sh
networkid=$1
genesisfile=$2
gasprice=$3
gaslimit=$4
datadir=$5

: ${ETHSTATS_URL:=""} #default to empty

echo datadir=$datadir
echo networkid=$networkid
echo genesisfile=$genesisfile
echo gasprice=$gasprice
echo gaslimit=$gaslimit

prepareDatadir $datadir $genesisfile

echo "Starting geth with BOOTNODE enabled on port '30301'"
exec geth --datadir $datadir \
 --networkid $networkid \
 --miner.gasprice $gasprice \
 --miner.gaslimit $gaslimit \
 --http --http.addr '0.0.0.0' --http.port 8545 --http.corsdomain '*' --http.api 'admin,engine,eth,miner,net,web3,personal,txpool,debug' --http.vhosts="*" \
 --ws --ws.addr '0.0.0.0' --ws.port 8546 --ws.origins '*' --ws.api 'admin,eth,miner,net,web3,personal,txpool,debug' \
 --port 30301 \
 --ethstats "$ETHSTATS_URL" \
 --bootnodes $BOOTNODES \
 --netrestrict "11.11.11.0/24" \
 --log.vmodule eth/*=5 \
 --txpool.nolocals  \
 --authrpc.port 8551 \
 --authrpc.addr=0.0.0.0 \
 --authrpc.vhosts=* \
 --authrpc.jwtsecret /jwt \
 --syncmode "snap"
