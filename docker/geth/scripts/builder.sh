#!/usr/bin/env bash
set -Eeu

if [ $# -lt 4 ]; then
    echo "Usage: $0 networkid genesisfile gasprice gaslimit [etherbase]"
    echo "Example: $0 12345 /genesis.json "
    exit
fi

source ./scripts/functions.sh
networkid=$1
genesisfile=$2
gasprice=$3
gaslimit=$4
etherbase=${5-""}

: ${BOOTNODES:=""} # default to empty
: ${ETHSTATS_URL:=""} #default to empty

echo DATA_DIR=$DATA_DIR
echo networkid=$networkid
echo genesisfile=$genesisfile
echo gasprice=$gasprice
echo gaslimit=$gaslimit
echo BOOTNODES=$BOOTNODES
echo ETHSTATS_URL=$ETHSTATS_URL
echo NAT=$NAT

mkdir -p /data/traces/raw
prepareDatadir $DATA_DIR $genesisfile

#cat /jwt-secret.hex

if [ ${#BOOTNODES} -ge 1 ]; then
#    bootnode=$(getent hosts $bootnodehost | awk '{print $1}')
    echo "Starting geth connecting to $BOOTNODES"
    geth --datadir $DATA_DIR \
      --networkid $networkid \
      --miner.gasprice $gasprice \
      --miner.gaslimit $gaslimit \
      --bootnodes $BOOTNODES \
      --http --http.addr '0.0.0.0' --http.port 8545 --http.corsdomain '*' --http.api 'admin,eth,miner,net,web3,personal,txpool,debug' --http.vhosts="*" \
      --ws --ws.addr '0.0.0.0' --ws.port 8546 --ws.origins '*' --ws.api 'admin,eth,miner,net,web3,personal,txpool,debug' \
      --ethstats "$ETHSTATS_URL" \
      --netrestrict "$NETRESTRICT" \
      --ipcdisable \
      --verbosity 4 \
      --txpool.nolocals \
      --syncmode "full" \
      --gcmode "archive" \
      --nat "$NAT" \
      --builder \
      --builder.local_relay \
      --builder.seconds_in_slot 12 \
      --builder.submission_offset=6s \
      --builder.algotype greedy
else
    echo "Builder needs at least one bootnodes to start"
fi
