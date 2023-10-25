#!/usr/bin/env bash
set -Eeu

if [ $# -lt 4 ]; then
    echo "Usage: $0 networkid genesisfile gasprice gaslimit [etherbase]"
    echo "Example: $0 12345 /genesis.json 0xa 0x21312 0x6d976c9b8ceee705d4fe8699b44e5eb58242f484"
    exit
fi

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
# echo NETRESTRICT=$NETRESTRICT
echo etherbase=$etherbase

mkdir -p /data/traces/raw
mkdir -p $DATA_DIR
ls -lah $DATA_DIR
# prepareDatadir $DATA_DIR $genesisfile

#cat /jwt-secret.hex

find $DATA_DIR

if [ ${#BOOTNODES} -ge 1 ]; then
#    bootnode=$(getent hosts $bootnodehost | awk '{print $1}')
    echo "Starting besu connecting to $BOOTNODES"
    besu --data-dir $DATA_DIR \
      --genesis-file $genesisfile \
      --network-id $networkid \
      --miner-gas-price $gasprice \
      --target-gas-price $gaslimit \
      --bootnodes $BOOTNODES \
      --rpc-http-enabled \
      --rpc-http-host '0.0.0.0' \
      --rpc-http-port 8545 \
      --rpc-http-cors-origins '*' \
      --rpc-http-apis  'ADMIN,ETH,MINER,NET,WEB3,PERSONAL,TXPOOL,DEBUG' \
      --rpc-ws-enabled \
      --rpc-ws-host '0.0.0.0' \
      --rpc-ws-port 8546 \
      --rpc-ws-apis 'ADMIN,ETH,MINER,NET,WEB3,PERSONAL,TXPOOL,DEBUG' \
      --ethstats "$ETHSTATS_URL" \
      --host-allowlist '*' \ # netrestrict
      --logging 'DEBUG' \
      --syncmode "full"
else
    echo "Starting besu VALIDATOR with BOOTNODE enabled on port '$BOOTNODE_PORT' with '/boot.key'"
    # mkdir -p $DATA_DIR/keystore
    # cp /keystore/$etherbase.json $DATA_DIR/keystore
    besu --data-path $DATA_DIR \
     --genesis-file $genesisfile \
     --network-id $networkid \
     --miner-coinbase $etherbase \
     --min-gas-price $gasprice \
     --target-gas-limit $gaslimit \
     --rpc-http-enabled \
     --rpc-http-host '11.11.11.101' \
     --rpc-http-port 8545 \
     --rpc-http-cors-origins '*' \
     --rpc-http-apis  'ADMIN,TXPOOL,WEB3,ETH,NET,PERM' \
     --rpc-http-max-active-connections 200 \
     --rpc-ws-enabled \
     --rpc-ws-host '0.0.0.0' \
     --rpc-ws-port 8546 \
     --rpc-ws-apis 'ADMIN,TXPOOL,WEB3,ETH,NET,PERM' \
     --rpc-ws-max-active-connections 200 \
     --metrics-enabled \
     --metrics-host '0.0.0.0' \
     --metrics-port 9545 \
     --engine-rpc-port 8550 \
     --engine-host-allowlist '*' \
     --node-private-key-file "/boot.key" \
     --p2p-host $(hostname -i) \
     --p2p-port $BOOTNODE_PORT \
     --ethstats "$ETHSTATS_URL" \
     --host-allowlist '*' \
     --logging 'INFO' \
     --miner-enabled \
     --sync-mode "full"
fi
