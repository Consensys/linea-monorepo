#!/usr/bin/env bash
#
# deploy-contracts.sh — Linea stack v0 contract deployment wrapper (Sepolia L1).
#
# Faithfully mirrors the six deploy targets from `contracts/makefile-contracts.mk`
# in serial:
#
#   1. deploy-linea-rollup-v8       (L1)  → emits LINEA_ROLLUP_ADDRESS
#   2. deploy-l2messageservice      (L2)  → emits L2_MESSAGE_SERVICE_ADDRESS
#   3. deploy-token-bridge-l1       (L1)  ← consumes both forwarded addresses
#   4. deploy-token-bridge-l2       (L2)  ← consumes both forwarded addresses
#   5. deploy-l1-test-erc20         (L1)
#   6. deploy-l2-test-erc20         (L2)
#
# Validium variant (LINEA_COORDINATOR_DATA_AVAILABILITY=VALIDIUM) replaces
# step 1 with deploy-validium-v2.
#
# After all 6 steps, this script:
#   - aggregates addresses into /shared/addresses.json
#   - patches /rendered/coordinator-config.toml + /rendered/maru-config.toml
#     with the discovered LINEA_ROLLUP_ADDRESS, L2_MESSAGE_SERVICE_ADDRESS,
#     L2 genesis state root, computed shnarf, and L2 deploy block
#   - patches the postman env file with L1+L2 contract addresses
#
# A `post-deploy-restart` compose service then restarts maru (which booted with
# zero-default placeholders) so it picks up the patched config. Coordinator and
# postman start *after* this script, so they read the patched values on first
# boot — no restart needed.
#
# See scripts/DEPLOY-ENV-CONTRACT.md for the full env-var contract per step.
#
# Args:
#   $1 — L1 RPC endpoint (Sepolia HTTPS RPC; e.g. https://sepolia.infura.io/v3/...)
#   $2 — L2 RPC endpoint (internal: http://l2-node-besu:8545)
#   $3 — output path for addresses.json (e.g. /shared/addresses.json)
#
# Required env (compose injects from .env):
#   L1_RPC_URL                            — same as $1, REQUIRED, no default.
#   L1_DEPLOYER_PRIVATE_KEY               — Sepolia-funded; used for ALL L1 roles
#                                           (Option A: single key for security
#                                           council, rollup operators, deployer).
#
# Optional env:
#   LINEA_COORDINATOR_DATA_AVAILABILITY   — ROLLUP (default) | VALIDIUM
#   L2_DEPLOYER_PRIVATE_KEY               — overrides pre-baked L2 dev key
#   L2_CHAIN_ID                           — overrides default 1337
#
# L2 keys remain pre-baked (we own L2 genesis); L1 keys are user-supplied.
#
set -euo pipefail

# ─────────────────────────────────────────────────────────────────────────────
# Inputs and constants
# ─────────────────────────────────────────────────────────────────────────────

# L1_RPC_URL has no fallback — local L1 services are gone (Phase 1 Sepolia migration).
L1_RPC_URL="${1:-${L1_RPC_URL:?L1_RPC_URL must be set (Sepolia HTTPS RPC URL)}}"
L2_RPC_URL="${2:-${L2_RPC_URL:-http://l2-node-besu:8545}}"
OUT_PATH="${3:-${ADDRESSES_OUTPUT_PATH:-/shared/addresses.json}}"

# Required: user-supplied L1 deployer key (must be Sepolia-funded).
: "${L1_DEPLOYER_PRIVATE_KEY:?L1_DEPLOYER_PRIVATE_KEY must be set (Sepolia-funded)}"
# L2 deployer is pre-baked dev (we own L2 genesis); override only for testing.
: "${L2_DEPLOYER_PRIVATE_KEY:=0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae}"
: "${LINEA_COORDINATOR_DATA_AVAILABILITY:=ROLLUP}"
: "${L1_CONTRACT_VERSION:=8}"

CONTRACTS_DIR="/workspace/contracts"
ART_DIR="${CONTRACTS_DIR}/local-deployments-artifacts"
LOG_DIR="/shared/deploy-logs"
RENDERED_DIR="/rendered"   # mounted RW from compose; coordinator + maru read RO
mkdir -p "$LOG_DIR" "$(dirname "$OUT_PATH")"

# L2 chain ID is fixed by L2 genesis (we own it). L1 chain ID is detected at
# runtime via eth_chainId — see "Detect L1 / L2 genesis values" section below.
L2_CHAIN_ID="${L2_CHAIN_ID:-1337}"
L1_CHAIN_ID=""   # populated post-Foundry-install

log()  { printf "[deploy-contracts] %s\n" "$*"; }
die()  { printf "[deploy-contracts] ERROR: %s\n" "$*" >&2; exit 1; }
step() { printf "\n[deploy-contracts] ===== %s =====\n" "$*"; }

# ─────────────────────────────────────────────────────────────────────────────
# Tooling: pnpm + Foundry (idempotent)
# ─────────────────────────────────────────────────────────────────────────────

[[ -d "$CONTRACTS_DIR" ]] || die "contracts dir not bind-mounted at $CONTRACTS_DIR"

step "Tooling"
log "Enabling corepack + pnpm"
corepack enable >/dev/null 2>&1 || true
corepack prepare pnpm@latest --activate >/dev/null 2>&1 || true

# Hardhat-foundry plugin calls `forge` at config-load time; foundry isn't in
# the node image. Idempotent install.
if [[ ! -x "$HOME/.foundry/bin/forge" ]]; then
  log "Installing Foundry"
  curl -fsSL https://foundry.paradigm.xyz | bash
  "$HOME/.foundry/bin/foundryup"
fi
export PATH="$HOME/.foundry/bin:$PATH"
forge --version | head -1

cd "$CONTRACTS_DIR"

# ─────────────────────────────────────────────────────────────────────────────
# Pre-flight: RPCs reachable, L2 fork timestamp readable
# ─────────────────────────────────────────────────────────────────────────────

step "Pre-flight"

wait_rpc() {
  local url="$1" name="$2"
  for _ in $(seq 1 60); do
    if curl -fsS --max-time 2 \
         -H "Content-Type: application/json" \
         -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
         "$url" >/dev/null 2>&1; then
      log "$name reachable: $url"
      return 0
    fi
    sleep 2
  done
  die "$name not reachable after 120s: $url"
}
wait_rpc "$L1_RPC_URL" L1
wait_rpc "$L2_RPC_URL" L2

# L2 genesis-init writes /initialization/fork-timestamp.txt. The l2-genesis-init
# bind-mount is at /initialization in the genesis-init container; in our
# deploy-contracts container it's only reachable via /workspace/docs/...
# We try both common paths and fall back to the makefile-contracts.mk default.
FORK_TIMESTAMP=""
for f in \
    "/workspace/docs/getting-started/linea-stack/config/l2/genesis-init/fork-timestamp.txt" \
    "/workspace/docker/config/l2-genesis-initialization/fork-timestamp.txt"; do
  if [[ -f "$f" ]]; then
    FORK_TIMESTAMP="$(cat "$f")"
    log "Read FORK_TIMESTAMP=$FORK_TIMESTAMP from $f"
    break
  fi
done
: "${FORK_TIMESTAMP:=1683325137}"

# ─────────────────────────────────────────────────────────────────────────────
# Detect L1 / L2 genesis values  (uses cast — Foundry was installed above)
# ─────────────────────────────────────────────────────────────────────────────

step "Detect L1 chain ID + L1 deployer address + L2 genesis state root + shnarf"

# L1 chain ID via eth_chainId. Sepolia = 11155111.
L1_CHAIN_ID="$(cast chain-id --rpc-url "$L1_RPC_URL")" \
  || die "Failed to query L1 chain ID from $L1_RPC_URL"
log "L1_CHAIN_ID=$L1_CHAIN_ID"

# L1 deployer address — Option A: single L1 key drives all roles
# (security council, rollup operators, deployer). Phase-3 will also point
# the web3signer keystore + postman L1 signer at this same key.
L1_DEPLOYER_ADDRESS="$(cast wallet address --private-key "$L1_DEPLOYER_PRIVATE_KEY")" \
  || die "Failed to derive L1 deployer address from L1_DEPLOYER_PRIVATE_KEY"
log "L1_DEPLOYER_ADDRESS=$L1_DEPLOYER_ADDRESS"

# L2 genesis state root — must match what gets baked into LineaRollup at
# initialization and what the coordinator advertises in [protocol.genesis].
L2_GENESIS_STATE_ROOT="$(cast block 0 --rpc-url "$L2_RPC_URL" --field stateRoot)" \
  || die "Failed to read L2 genesis state root from $L2_RPC_URL"
[[ "$L2_GENESIS_STATE_ROOT" =~ ^0x[0-9a-fA-F]{64}$ ]] \
  || die "L2 genesis state root malformed: $L2_GENESIS_STATE_ROOT"
log "L2_GENESIS_STATE_ROOT=$L2_GENESIS_STATE_ROOT"

# Genesis shnarf: keccak256(parentShnarf || snarkHash || parentStateRootHash
#                            || evalClaim || evalPoint), each 32 bytes.
# Only parentStateRootHash is non-zero. Formula is the V6 shape per the
# coordinator-config comment; verify against V8 at first-boot validation.
ZERO_32_HEX="0000000000000000000000000000000000000000000000000000000000000000"
STATE_ROOT_NO_PREFIX="${L2_GENESIS_STATE_ROOT#0x}"
SHNARF_INPUT="0x${ZERO_32_HEX}${ZERO_32_HEX}${STATE_ROOT_NO_PREFIX}${ZERO_32_HEX}${ZERO_32_HEX}"
L2_GENESIS_SHNARF="$(cast keccak "$SHNARF_INPUT")" \
  || die "Failed to compute genesis shnarf"
log "L2_GENESIS_SHNARF=$L2_GENESIS_SHNARF"

export L1_CHAIN_ID L1_DEPLOYER_ADDRESS L2_GENESIS_STATE_ROOT L2_GENESIS_SHNARF

# ─────────────────────────────────────────────────────────────────────────────
# Install + compile (idempotent against host bind mount)
# ─────────────────────────────────────────────────────────────────────────────

step "pnpm install + hardhat compile"

if [[ ! -d /workspace/node_modules ]]; then
  log "Installing workspace dependencies (pnpm install --no-frozen-lockfile)"
  ( cd /workspace && pnpm install --no-frozen-lockfile )
fi

if [[ ! -d artifacts ]]; then
  log "Compiling contracts (pnpm exec hardhat compile)"
  pnpm exec hardhat compile
fi

# ─────────────────────────────────────────────────────────────────────────────
# Capture pre-deploy nonces (the parent prelude in makefile-contracts.mk:158)
# ─────────────────────────────────────────────────────────────────────────────

step "Capture pre-deploy nonces"

L1_NONCE=$(pnpm -s exec ts-node "$ART_DIR/get-wallet-nonce.ts" \
  --wallet-priv-key "$L1_DEPLOYER_PRIVATE_KEY" --rpc-url "$L1_RPC_URL")
L2_NONCE=$(pnpm -s exec ts-node "$ART_DIR/get-wallet-nonce.ts" \
  --wallet-priv-key "$L2_DEPLOYER_PRIVATE_KEY" --rpc-url "$L2_RPC_URL")
log "L1_NONCE=$L1_NONCE  L2_NONCE=$L2_NONCE"
export L1_NONCE L2_NONCE

# ─────────────────────────────────────────────────────────────────────────────
# Helpers — extract a deployed contract's address from a step's stdout
# ─────────────────────────────────────────────────────────────────────────────
#
# Deploy scripts emit lines like:
#   contract=LineaRollupV8 deployed: address=0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9 blockNumber=72 chainId=31648428
# We grep for `contract=NAME deployed:` and pull out the address.

extract_address() {
  local logfile="$1" contract_name="$2"
  local line addr
  line="$(grep -E "^contract=${contract_name} deployed: " "$logfile" | tail -1 || true)"
  if [[ -z "$line" ]]; then
    return 1
  fi
  addr="$(echo "$line" | sed -nE 's/.*address=(0x[a-fA-F0-9]{40}).*/\1/p')"
  if [[ -z "$addr" ]]; then
    return 1
  fi
  printf "%s" "$addr"
}

require_address() {
  local logfile="$1" contract_name="$2" out
  out="$(extract_address "$logfile" "$contract_name")" || \
    die "failed to extract $contract_name address from $logfile"
  printf "%s" "$out"
}

# ─────────────────────────────────────────────────────────────────────────────
# Step functions — one per Make target
# ─────────────────────────────────────────────────────────────────────────────

# Step 1 — deploy-linea-rollup-v$L1_CONTRACT_VERSION (or deploy-validium-v2)
step1_l1_rollup() {
  step "Step 1: deploy L1 Verifier + LineaRollup (or Validium)"
  local script logfile

  if [[ "$LINEA_COORDINATOR_DATA_AVAILABILITY" == "VALIDIUM" ]]; then
    script="$ART_DIR/deployPlonkVerifierAndValidiumV2.ts"
    logfile="$LOG_DIR/step1-validium.log"
  else
    script="$ART_DIR/deployPlonkVerifierAndLineaRollupV${L1_CONTRACT_VERSION}.ts"
    logfile="$LOG_DIR/step1-linea-rollup.log"
  fi

  [[ -f "$script" ]] || die "missing deploy script: $script"

  # Option A: L1 deployer key drives every L1 role (security council, operators,
  # security-council-private-key). On Sepolia the user only needs to fund a
  # single address. Phase-3 will align the web3signer keystore the same way.
  if [[ "$LINEA_COORDINATOR_DATA_AVAILABILITY" == "VALIDIUM" ]]; then
    DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
    RPC_URL="$L1_RPC_URL" \
    VERIFIER_CONTRACT_NAME="IntegrationTestTrueVerifier" \
    INITIAL_L2_STATE_ROOT_HASH="$L2_GENESIS_STATE_ROOT" \
    INITIAL_L2_BLOCK_NUMBER="0" \
    L2_GENESIS_TIMESTAMP="$FORK_TIMESTAMP" \
    L1_SECURITY_COUNCIL="$L1_DEPLOYER_ADDRESS" \
    VALIDIUM_OPERATORS="${VALIDIUM_OPERATORS:-$L1_DEPLOYER_ADDRESS}" \
    VALIDIUM_RATE_LIMIT_PERIOD="86400" \
    VALIDIUM_RATE_LIMIT_AMOUNT="1000000000000000000000" \
      pnpm -s exec ts-node "$script" 2>&1 | tee "$logfile"
  else
    DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
    RPC_URL="$L1_RPC_URL" \
    VERIFIER_CONTRACT_NAME="IntegrationTestTrueVerifier" \
    INITIAL_L2_STATE_ROOT_HASH="$L2_GENESIS_STATE_ROOT" \
    INITIAL_L2_BLOCK_NUMBER="0" \
    L2_GENESIS_TIMESTAMP="$FORK_TIMESTAMP" \
    L1_SECURITY_COUNCIL="$L1_DEPLOYER_ADDRESS" \
    LINEA_ROLLUP_OPERATORS="${LINEA_ROLLUP_OPERATORS:-$L1_DEPLOYER_ADDRESS}" \
    LINEA_ROLLUP_RATE_LIMIT_PERIOD="86400" \
    LINEA_ROLLUP_RATE_LIMIT_AMOUNT="1000000000000000000000" \
    FORCED_TRANSACTION_GATEWAY_L2_CHAIN_ID="$L2_CHAIN_ID" \
    FORCED_TRANSACTION_GATEWAY_L2_BLOCK_BUFFER="2000" \
    FORCED_TRANSACTION_GATEWAY_MAX_GAS_LIMIT="300000" \
    FORCED_TRANSACTION_GATEWAY_MAX_INPUT_LENGTH_BUFFER="1000" \
    FORCED_TRANSACTION_L2_BLOCK_DURATION_SECONDS="2" \
    FORCED_TRANSACTION_BLOCK_NUMBER_DEADLINE_BUFFER="10" \
    SECURITY_COUNCIL_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
    YIELD_MANAGER_ADDRESS="0x000000000000000000000000000000000000dEaD" \
      pnpm -s exec ts-node "$script" 2>&1 | tee "$logfile"
  fi

  # Forward the rollup address into the global env for steps 3 + 4.
  if [[ "$LINEA_COORDINATOR_DATA_AVAILABILITY" == "VALIDIUM" ]]; then
    LINEA_ROLLUP_ADDRESS="$(require_address "$logfile" "ValidiumV2")"
  else
    LINEA_ROLLUP_ADDRESS="$(require_address "$logfile" "LineaRollupV${L1_CONTRACT_VERSION}")"
  fi
  export LINEA_ROLLUP_ADDRESS
  log "Forwarding LINEA_ROLLUP_ADDRESS=$LINEA_ROLLUP_ADDRESS"
}

# Step 2 — deploy-l2messageservice
step2_l2_message_service() {
  step "Step 2: deploy L2 MessageService"
  local logfile="$LOG_DIR/step2-l2-message-service.log"

  L2_MESSAGE_SERVICE_CONTRACT_NAME="L2MessageService" \
  DEPLOYER_PRIVATE_KEY="$L2_DEPLOYER_PRIVATE_KEY" \
  RPC_URL="$L2_RPC_URL" \
  L2_SECURITY_COUNCIL="0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266" \
  L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER="${L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER:-0xd42e308fc964b71e18126df469c21b0d7bcb86cc}" \
  L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD="86400" \
  L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT="1000000000000000000000" \
    pnpm -s exec ts-node "$ART_DIR/deployL2MessageServiceV1.ts" 2>&1 | tee "$logfile"

  L2_MESSAGE_SERVICE_ADDRESS="$(require_address "$logfile" "L2MessageService")"
  export L2_MESSAGE_SERVICE_ADDRESS
  log "Forwarding L2_MESSAGE_SERVICE_ADDRESS=$L2_MESSAGE_SERVICE_ADDRESS"
}

# Step 3 — deploy-token-bridge-l1
step3_token_bridge_l1() {
  step "Step 3: deploy L1 TokenBridge + L1 BridgedToken"
  local logfile="$LOG_DIR/step3-token-bridge-l1.log"

  : "${LINEA_ROLLUP_ADDRESS:?step1 must run first}"
  : "${L2_MESSAGE_SERVICE_ADDRESS:?step2 must run first}"

  # REMOTE_DEPLOYER_ADDRESS is the L2 deployer address (deterministic via CREATE
  # from the pre-baked L2 deployer at nonce 0). Stays hardcoded — we own L2.
  # L1_SECURITY_COUNCIL: Option A — same as L1 deployer.
  DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
  REMOTE_DEPLOYER_ADDRESS="0x1B9AbEeC3215D8AdE8a33607f2cF0f4F60e5F0D0" \
  RPC_URL="$L1_RPC_URL" \
  REMOTE_CHAIN_ID="$L2_CHAIN_ID" \
  TOKEN_BRIDGE_L1="true" \
  L1_SECURITY_COUNCIL="$L1_DEPLOYER_ADDRESS" \
  L2_MESSAGE_SERVICE_ADDRESS="$L2_MESSAGE_SERVICE_ADDRESS" \
  LINEA_ROLLUP_ADDRESS="$LINEA_ROLLUP_ADDRESS" \
    pnpm -s exec ts-node "$ART_DIR/deployBridgedTokenAndTokenBridgeV1_1.ts" 2>&1 | tee "$logfile"
}

# Step 4 — deploy-token-bridge-l2
step4_token_bridge_l2() {
  step "Step 4: deploy L2 TokenBridge + L2 BridgedToken"
  local logfile="$LOG_DIR/step4-token-bridge-l2.log"

  : "${LINEA_ROLLUP_ADDRESS:?step1 must run first}"
  : "${L2_MESSAGE_SERVICE_ADDRESS:?step2 must run first}"

  DEPLOYER_PRIVATE_KEY="$L2_DEPLOYER_PRIVATE_KEY" \
  REMOTE_DEPLOYER_ADDRESS="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" \
  RPC_URL="$L2_RPC_URL" \
  REMOTE_CHAIN_ID="$L1_CHAIN_ID" \
  TOKEN_BRIDGE_L1="false" \
  L2_SECURITY_COUNCIL="0xf17f52151EbEF6C7334FAD080c5704D77216b732" \
  L2_MESSAGE_SERVICE_ADDRESS="$L2_MESSAGE_SERVICE_ADDRESS" \
  LINEA_ROLLUP_ADDRESS="$LINEA_ROLLUP_ADDRESS" \
    pnpm -s exec ts-node "$ART_DIR/deployBridgedTokenAndTokenBridgeV1_1.ts" 2>&1 | tee "$logfile"
}

# Step 5 — deploy-l1-test-erc20
step5_l1_test_erc20() {
  step "Step 5: deploy L1 TestERC20"
  local logfile="$LOG_DIR/step5-l1-test-erc20.log"

  DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
  RPC_URL="$L1_RPC_URL" \
  TEST_ERC20_L1="true" \
  TEST_ERC20_NAME="TestERC20" \
  TEST_ERC20_SYMBOL="TERC20" \
  TEST_ERC20_INITIAL_SUPPLY="100000" \
    pnpm -s exec ts-node "$ART_DIR/deployTestERC20.ts" 2>&1 | tee "$logfile"
}

# Step 6 — deploy-l2-test-erc20
step6_l2_test_erc20() {
  step "Step 6: deploy L2 TestERC20"
  local logfile="$LOG_DIR/step6-l2-test-erc20.log"

  DEPLOYER_PRIVATE_KEY="$L2_DEPLOYER_PRIVATE_KEY" \
  RPC_URL="$L2_RPC_URL" \
  TEST_ERC20_L1="false" \
  TEST_ERC20_NAME="TestERC20" \
  TEST_ERC20_SYMBOL="TERC20" \
  TEST_ERC20_INITIAL_SUPPLY="100000" \
    pnpm -s exec ts-node "$ART_DIR/deployTestERC20.ts" 2>&1 | tee "$logfile"
}

# ─────────────────────────────────────────────────────────────────────────────
# Run all steps in order
# ─────────────────────────────────────────────────────────────────────────────

step1_l1_rollup
step2_l2_message_service
step3_token_bridge_l1
step4_token_bridge_l2
step5_l1_test_erc20
step6_l2_test_erc20

# ─────────────────────────────────────────────────────────────────────────────
# Aggregate addresses → addresses.json
# ─────────────────────────────────────────────────────────────────────────────

step "Aggregate addresses → $OUT_PATH"

node - <<'NODE_EOF' "$LOG_DIR" "$OUT_PATH" "$L1_CHAIN_ID" "$L2_CHAIN_ID" "$L1_RPC_URL" "$L2_RPC_URL"
const fs = require("fs");
const path = require("path");
const [, , logDir, outPath, l1ChainId, l2ChainId, l1Url, l2Url] = process.argv;

// Format emitted by every deploy script (see contracts/scripts/hardhat/utils.ts:146):
//   contract=NAME deployed: address=0xADDR blockNumber=N chainId=Z
const re = /^contract=(\S+)\s+deployed:\s+address=(0x[a-fA-F0-9]{40})\s+blockNumber=(\d+)\s+chainId=(\d+)/;

const result = {
  _meta: {
    l1ChainId,
    l2ChainId,
    l1RpcUrl: l1Url,
    l2RpcUrl: l2Url,
    generatedAt: new Date().toISOString(),
  },
  l1: {},
  l2: {},
};

for (const entry of fs.readdirSync(logDir).sort()) {
  if (!entry.endsWith(".log")) continue;
  const lines = fs.readFileSync(path.join(logDir, entry), "utf8").split("\n");
  for (const line of lines) {
    const m = line.match(re);
    if (!m) continue;
    const [, name, addr, , chainId] = m;
    const lane = chainId === l1ChainId ? "l1" : "l2";
    // Last write wins (a contract redeployed in a later step overrides earlier).
    result[lane][name] = addr;
  }
}

fs.writeFileSync(outPath, JSON.stringify(result, null, 2));
console.log("[deploy-contracts] wrote", outPath);
console.log("[deploy-contracts] L1 contracts:", Object.keys(result.l1).join(", ") || "(none)");
console.log("[deploy-contracts] L2 contracts:", Object.keys(result.l2).join(", ") || "(none)");
NODE_EOF

# ─────────────────────────────────────────────────────────────────────────────
# Patch postman env file with deployed addresses
# ─────────────────────────────────────────────────────────────────────────────
#
# Postman reads its env_file at container start. depends_on:
# `deploy-contracts (service_completed_successfully)` gates postman until
# this patch lands.

step "Patch postman env"

POSTMAN_ENV="${POSTMAN_ENV:-/workspace/docs/getting-started/linea-stack/config/l2/postman/env}"
if [[ -f "$POSTMAN_ENV" && -f "$OUT_PATH" ]]; then
  L1_ADDR="$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));console.log(j.l1.LineaRollupV8||j.l1.ValidiumV2||"")' "$OUT_PATH")"
  L2_ADDR="$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));console.log(j.l2.L2MessageService||"")' "$OUT_PATH")"
  if [[ -n "$L1_ADDR" && -n "$L2_ADDR" ]]; then
    log "Patching postman env: L1_CONTRACT_ADDRESS=$L1_ADDR L2_CONTRACT_ADDRESS=$L2_ADDR"
    sed -i "s|^L1_CONTRACT_ADDRESS=.*|L1_CONTRACT_ADDRESS=${L1_ADDR}|" "$POSTMAN_ENV"
    sed -i "s|^L2_CONTRACT_ADDRESS=.*|L2_CONTRACT_ADDRESS=${L2_ADDR}|" "$POSTMAN_ENV"
  else
    log "WARNING: could not extract L1 LineaRollupV8/L2 L2MessageService from $OUT_PATH; postman env not patched"
  fi
else
  log "skipping postman env patch (POSTMAN_ENV=$POSTMAN_ENV exists=$(test -f "$POSTMAN_ENV" && echo yes || echo no))"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Patch rendered coordinator + maru configs with discovered values
# ─────────────────────────────────────────────────────────────────────────────
#
# config-render seeds these placeholders with safe defaults at boot
# (zero-address / zero-hash) so maru can come up before deploy-contracts runs.
# Now that we have the real values, we re-write the rendered files in place.
#
# - coordinator + postman start AFTER this script (depends_on: completed_
#   successfully), so they read the patched values on first boot.
# - maru started before this script, so a `post-deploy-restart` compose service
#   will restart it once we exit cleanly.

step "Patch rendered coordinator + maru configs"

if [[ ! -d "$RENDERED_DIR" ]]; then
  log "WARNING: $RENDERED_DIR not mounted; skipping rendered-config patch"
elif [[ ! -f "$RENDERED_DIR/coordinator-config.toml" || ! -f "$RENDERED_DIR/maru-config.toml" ]]; then
  log "WARNING: rendered configs missing under $RENDERED_DIR; skipping patch"
else
  # Pull the contracts from addresses.json (single source of truth).
  LINEA_ROLLUP_ADDR_FROM_JSON="$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));console.log(j.l1.LineaRollupV8||j.l1.ValidiumV2||"")' "$OUT_PATH")"
  L2_MS_ADDR_FROM_JSON="$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));console.log(j.l2.L2MessageService||"")' "$OUT_PATH")"

  # L2 block where L2MessageService deployed — extracted from step 2 log.
  L2_MS_DEPLOY_BLOCK=""
  if [[ -f "$LOG_DIR/step2-l2-message-service.log" ]]; then
    L2_MS_DEPLOY_BLOCK="$(grep -E "^contract=L2MessageService deployed: " "$LOG_DIR/step2-l2-message-service.log" | tail -1 | sed -nE 's/.*blockNumber=([0-9]+).*/\1/p')"
  fi
  : "${L2_MS_DEPLOY_BLOCK:=0}"

  if [[ -z "$LINEA_ROLLUP_ADDR_FROM_JSON" || -z "$L2_MS_ADDR_FROM_JSON" ]]; then
    log "WARNING: addresses.json missing LineaRollup or L2MessageService; skipping rendered-config patch"
  else
    log "Patching $RENDERED_DIR/maru-config.toml with contract-address=$LINEA_ROLLUP_ADDR_FROM_JSON"
    # Maru: only one `contract-address` line (under [linea]). Line-anchored sed safe.
    sed -i 's|^contract-address = ".*"|contract-address = "'"$LINEA_ROLLUP_ADDR_FROM_JSON"'"|' \
      "$RENDERED_DIR/maru-config.toml"

    log "Patching $RENDERED_DIR/coordinator-config.toml: L1=$LINEA_ROLLUP_ADDR_FROM_JSON L2=$L2_MS_ADDR_FROM_JSON deploy_block=$L2_MS_DEPLOY_BLOCK"
    # Coordinator: section-aware patch via awk. Two `contract-address` lines —
    # one under [protocol.l1] and one under [protocol.l2] — must be discriminated.
    awk \
      -v l1_addr="$LINEA_ROLLUP_ADDR_FROM_JSON" \
      -v l2_addr="$L2_MS_ADDR_FROM_JSON" \
      -v state_root="$L2_GENESIS_STATE_ROOT" \
      -v shnarf="$L2_GENESIS_SHNARF" \
      -v deploy_block="$L2_MS_DEPLOY_BLOCK" \
    '
      /^\[protocol\.l1\]/      { section="l1"; print; next }
      /^\[protocol\.l2\]/      { section="l2"; print; next }
      /^\[/                    { section=""; print; next }
      /^contract-address[[:space:]]*=/ {
        if (section == "l1")      { sub(/".*"/, "\"" l1_addr "\""); }
        else if (section == "l2") { sub(/".*"/, "\"" l2_addr "\""); }
      }
      /^genesis-state-root-hash[[:space:]]*=/ { sub(/".*"/, "\"" state_root "\""); }
      /^genesis-shnarf[[:space:]]*=/          { sub(/".*"/, "\"" shnarf "\""); }
      /^contract-deployment-block-number[[:space:]]*=/ { sub(/=.*/, "= " deploy_block); }
      { print }
    ' "$RENDERED_DIR/coordinator-config.toml" > "$RENDERED_DIR/coordinator-config.toml.new" \
      && mv "$RENDERED_DIR/coordinator-config.toml.new" "$RENDERED_DIR/coordinator-config.toml"
    log "Patched coordinator-config.toml"
  fi
fi

step "Done"
log "addresses.json at $OUT_PATH"
log "step logs at $LOG_DIR/"
