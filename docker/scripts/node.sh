#!/usr/bin/env bash
set -Eeu

if [ $# -lt 4 ]; then
    echo "Usage: $0 networkid genesisfile gasprice gaslimit"
    echo Example: $0 12345 /genesis.json
    exit
fi

source ./scripts/functions.sh
networkid=$1
genesisfile=$2
gasprice=$3
gaslimit=$4

: ${BOOTNODES:=""} # default to empty
: ${ETHSTATS_URL:=""} #default to empty

echo DATA_DIR=$DATA_DIR
echo networkid=$networkid
echo genesisfile=$genesisfile
echo gasprice=$gasprice
echo gaslimit=$gaslimit
echo BOOTNODES=$BOOTNODES
echo ETHSTATS_URL=$ETHSTATS_URL
echo NETRESTRICT=$NETRESTRICT

prepareDatadir $DATA_DIR $genesisfile

if [ ${#BOOTNODES} -ge 1 ]; then
#    bootnode=$(getent hosts $bootnodehost | awk '{print $1}')
    echo "Starting geth connecting to $BOOTNODES"
    geth --datadir $DATA_DIR \
      --networkid $networkid \
	    --miner.gasprice $gasprice \
      --miner.gaslimit $gaslimit \
      --http --http.addr '0.0.0.0' --http.port 8545 --http.corsdomain '*' --http.api 'admin,eth,miner,net,web3,personal,txpool,debug' --http.vhosts="*" \
      --ws --ws.addr '0.0.0.0' --ws.port 8546 --ws.origins '*' --ws.api 'admin,eth,miner,net,web3,personal,txpool,debug' \
      --bootnodes $BOOTNODES \
      --ethstats "$ETHSTATS_URL" \
      --ipcdisable \
      --netrestrict "$NETRESTRICT" \
      --verbosity 3 \
      --syncmode "full"
else
    echo "Starting geth with BOOTNODE enabled on port '$BOOTNODE_PORT' with '/boot.key'"
    geth --datadir $DATA_DIR \
     --networkid $networkid \
	 --miner.gasprice $gasprice \
     --miner.gaslimit $gaslimit \
     --http --http.addr '0.0.0.0' --http.port 8545 --http.corsdomain '*' --http.api 'admin,eth,miner,net,web3,personal,txpool,debug' --http.vhosts="*" \
     --ws --ws.addr '0.0.0.0' --ws.port 8546 --ws.origins '*' --ws.api 'admin,eth,miner,net,web3,personal,txpool,debug' \
     --nodekey "/boot.key" \
     --port $BOOTNODE_PORT \
     --ethstats "$ETHSTATS_URL" \
     --ipcdisable \
     --netrestrict "$NETRESTRICT" \
     --verbosity 3 \
     --txpool.nolocals  \
     --syncmode "full"
fi
