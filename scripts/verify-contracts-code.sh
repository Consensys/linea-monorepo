#!/usr/bin/env bash
set -euo pipefail

echo "Verifying deployed contracts have non-empty code..."

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Helper to call eth_getCode without jq
eth_get_code() {
  local url="$1" addr="$2"
  curl -s -X POST "$url" -H 'Content-Type: application/json' \
    --data "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"eth_getCode\",\"params\":[\"$addr\",\"latest\"]}" \
    | sed -n 's/.*"result":"\([^"]*\)".*/\1/p'
}

fail=false

# Linea Protocol addresses (from artifacts files)
L1_ROLLUP_FILE="$ROOT_DIR/contracts/local-deployments-artifacts/L1RollupAddress.txt"
L2_MSG_FILE="$ROOT_DIR/contracts/local-deployments-artifacts/L2MessageServiceAddress.txt"

L1_ROLLUP_ADDR=$(test -f "$L1_ROLLUP_FILE" && cat "$L1_ROLLUP_FILE" || echo "")
L2_MSG_ADDR=$(test -f "$L2_MSG_FILE" && cat "$L2_MSG_FILE" || echo "")

if [[ -n "$L1_ROLLUP_ADDR" ]]; then
  code=$(eth_get_code http://localhost:8445 "$L1_ROLLUP_ADDR")
  echo "L1 Rollup: $L1_ROLLUP_ADDR -> ${#code} bytes"
  [[ "$code" == "0x" || -z "$code" ]] && fail=true
else
  echo "L1 Rollup: address file missing"; fail=true
fi

if [[ -n "$L2_MSG_ADDR" ]]; then
  code=$(eth_get_code http://localhost:9045 "$L2_MSG_ADDR")
  echo "L2 MessageService: $L2_MSG_ADDR -> ${#code} bytes"
  [[ "$code" == "0x" || -z "$code" ]] && fail=true
else
  echo "L2 MessageService: address file missing"; fail=true
fi

# Status Network addresses (from helper)
get_addr() {
  (cd "$ROOT_DIR/status-network-contracts" && ./scripts/get-deployed-address.sh "$1" "$2" 2>/dev/null || true)
}

KARMA_TIERS=$(get_addr DeployKarmaTiers.s.sol KarmaTiers)
STAKE_MANAGER=$(get_addr DeployStakeManager.s.sol StakeManager)
KARMA=$(get_addr DeployKarma.s.sol Karma)
RLN=$(get_addr RLN.s.sol RLN)
KARMA_NFT=$(get_addr DeployKarmaNFT.s.sol KarmaNFT)

for row in \
  "KarmaTiers $KARMA_TIERS" \
  "StakeManager $STAKE_MANAGER" \
  "Karma $KARMA" \
  "RLN $RLN" \
  "KarmaNFT $KARMA_NFT"; do
  name=${row%% *}
  addr=${row#* }
  if [[ -n "$addr" ]]; then
    code=$(eth_get_code http://localhost:8545 "$addr")
    echo "$name: $addr -> ${#code} bytes"
    [[ "$code" == "0x" || -z "$code" ]] && fail=true
  else
    echo "$name: address not found"; fail=true
  fi
done

if [[ "$fail" == true ]]; then
  echo "One or more contracts missing code. Verification FAILED." >&2
  exit 1
fi

echo "All Linea and Status Network contracts deployed and code verified."
exit 0


