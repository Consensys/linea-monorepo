#!/usr/bin/env bash
set -Eeu

if [ $# -ne 5 ]; then
    echo "Usage: $0 etherbase networkid genesis_file gasprice gaslimit"
    echo "Example: $0 0x6d976c9b8ceee705d4fe8699b44e5eb58242f484 24709958 genesis.json 1 0x1C9C380"
    exit
fi

source ./scripts/functions.sh

etherbase=$1
networkid=$2
genesisfile=$3
gasprice=$4
gaslimit=$5
: ${ETHSTATS_URL:=""} #default to empty

echo etherbase=$etherbase
echo networkid=$networkid
echo genesisfile=$genesisfile
echo gasprice=$gasprice
echo gaslimit=$gaslimit
echo ETHSTATS_URL=$ETHSTATS_URL
echo BOOTNODES=$BOOTNODES

mkdir -p $DATA_DIR/keystore
cp /keystore/$etherbase.json $DATA_DIR/keystore
prepareDatadir $DATA_DIR $genesisfile

geth --datadir $DATA_DIR \
     --networkid $networkid \
     --miner.etherbase $etherbase \
     --miner.gasprice $gasprice \
     --miner.gaslimit $gaslimit \
     --bootnodes $BOOTNODES \
     --allow-insecure-unlock \
     --http --http.addr '0.0.0.0' --http.port 8545 --http.corsdomain '*' --http.api 'admin,eth,miner,net,web3,personal,txpool,debug' --http.vhosts="*" \
     --ethstats=$ETHSTATS_URL \
     --netrestrict $NETRESTRICT \
     --unlock $etherbase \
     --password "/dev/null" \
     --ipcdisable \
     --verbosity 3 \
     --mine \
     --syncmode "full"
