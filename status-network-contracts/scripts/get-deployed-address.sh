#!/bin/bash

# Script to extract deployed contract addresses from forge broadcast files
# Usage: ./get-deployed-address.sh <script-name> <contract-name> [chain-id]

SCRIPT_NAME=$1
CONTRACT_NAME=$2
CHAIN_ID=${3:-1337}

if [ -z "$SCRIPT_NAME" ] || [ -z "$CONTRACT_NAME" ]; then
    echo "Usage: $0 <script-name> <contract-name> [chain-id]"
    echo "Example: $0 DeployKarma.s.sol Karma 1337"
    exit 1
fi

BROADCAST_FILE="broadcast/${SCRIPT_NAME}/${CHAIN_ID}/run-latest.json"

if [ ! -f "$BROADCAST_FILE" ]; then
    echo "Error: Broadcast file not found: $BROADCAST_FILE"
    exit 1
fi

# Extract the contract address using jq
ADDRESS=$(cat "$BROADCAST_FILE" | jq -r --arg contract "$CONTRACT_NAME" '.transactions[] | select(.contractName == $contract) | .contractAddress' 2>/dev/null | head -1)

if [ -z "$ADDRESS" ] || [ "$ADDRESS" = "null" ]; then
    echo "Error: Contract $CONTRACT_NAME not found in $BROADCAST_FILE"
    exit 1
fi

echo "$ADDRESS"