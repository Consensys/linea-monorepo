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
#   - patches /rendered/coordinator-config.toml with deploy-time-only values:
#     L2 genesis state root, computed shnarf, and L2 deploy block.
#
# Coordinator and postman start *after* this script. Coordinator reads the
# patched rendered TOML. Postman reads /shared/addresses.json at startup and
# exports the deployed contract addresses into its process environment.
#
# See scripts/DEPLOY-ENV-CONTRACT.md for the full env-var contract per step.
#
# Required env (compose injects from .env):
#   L1_RPC_URL                            — Sepolia HTTPS RPC, REQUIRED, no default.
#   L1_DEPLOYER_PRIVATE_KEY               — Sepolia-funded deployer key. account-setup
#                                           derives separate L1 submitter keys for
#                                           coordinator blob/finalization roles.
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
  local url="$1" name="$2" resp
  for _ in $(seq 1 120); do
    resp="$(curl -sS --max-time 10 \
         -H "Content-Type: application/json" \
         -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
         "$url" 2>/dev/null || true)"
    if echo "$resp" | grep -qE '"result"[[:space:]]*:[[:space:]]*"0x[0-9a-fA-F]+"'; then
      log "$name RPC reachable"
      return 0
    fi
    sleep 2
  done
  die "$name RPC did not return eth_blockNumber after 240s"
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
# Read precomputed addresses + L1 chain ID + deployer from account-setup output
# ─────────────────────────────────────────────────────────────────────────────
#
# Replaces the cast-based detection of Phase 2. account-setup wrote everything
# to /shared/addresses-precomputed.json before any service booted; we now just
# load those values. All extracted addresses become the source-of-truth for the
# verify-or-die step after each contract deploy below.

step "Read precomputed addresses from /shared/addresses-precomputed.json"

PRECOMPUTED="${PRECOMPUTED:-/shared/addresses-precomputed.json}"
[[ -f "$PRECOMPUTED" ]] || die "$PRECOMPUTED not found — account-setup must run first"

# POSIX-y JSON extraction (account-setup writes a controlled shape).
json_field()  { sed -nE "s/.*\"$1\":[[:space:]]*\"([^\"]+)\".*/\1/p" "$PRECOMPUTED" | head -1; }
json_int()    { sed -nE "s/.*\"$1\":[[:space:]]*([0-9]+).*/\1/p" "$PRECOMPUTED" | head -1; }

L1_CHAIN_ID="$(json_field l1ChainId)"
L1_DEPLOYER_ADDRESS="$(json_field l1)"   # deployers.l1 is the first "l1": "0x..." in the file
# Disambiguate: deployers.l1 vs l1.{...} — fall back to grepping deployers block
# specifically if the order ever changes.
if ! [[ "$L1_DEPLOYER_ADDRESS" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
  L1_DEPLOYER_ADDRESS="$(awk '/\"deployers\":/{f=1} f && /\"l1\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
fi

PRECOMPUTED_LINEA_ROLLUP="$(awk '/\"l1\":[[:space:]]*\{/{f=1} f && /\"LineaRollupV8\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_FORCED_TX_GW="$(awk '/\"l1\":[[:space:]]*\{/{f=1} f && /\"ForcedTransactionGateway\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_L1_TOKEN_BRIDGE="$(awk '/\"l1\":[[:space:]]*\{/{f=1} f && /\"TokenBridge\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_L1_TEST_ERC20="$(awk '/\"l1\":[[:space:]]*\{/{f=1} f && /\"TestERC20\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_L2_MS="$(awk '/\"l2\":[[:space:]]*\{/{f=1} f && /\"L2MessageService\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_L2_TOKEN_BRIDGE="$(awk '/\"l2\":[[:space:]]*\{/{f=1} f && /\"TokenBridge\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_L2_TEST_ERC20="$(awk '/\"l2\":[[:space:]]*\{/{f=1} f && /\"TestERC20\":/{gsub(/[\",]/,""); print $2; exit}' "$PRECOMPUTED")"
PRECOMPUTED_L1_BLOB_SUBMITTER="$(json_field l1BlobSubmitterAddress)"
PRECOMPUTED_L1_FINALIZATION_SUBMITTER="$(json_field l1FinalizationSubmitterAddress)"
PRECOMPUTED_L2_MESSAGE_ANCHORING="$(json_field l2MessageAnchoringAddress)"


[[ "$L1_CHAIN_ID" =~ ^[0-9]+$ ]] || die "Could not extract l1ChainId from $PRECOMPUTED"
[[ "$L1_DEPLOYER_ADDRESS" =~ ^0x[a-fA-F0-9]{40}$ ]] || die "Could not extract deployers.l1 from $PRECOMPUTED"
[[ "$PRECOMPUTED_LINEA_ROLLUP" =~ ^0x[a-fA-F0-9]{40}$ ]] || die "Could not extract l1.LineaRollupV8 from $PRECOMPUTED"
[[ "$PRECOMPUTED_L2_MS" =~ ^0x[a-fA-F0-9]{40}$ ]] || die "Could not extract l2.L2MessageService from $PRECOMPUTED"
[[ "$PRECOMPUTED_L1_BLOB_SUBMITTER" =~ ^0x[a-fA-F0-9]{40}$ ]] || die "Could not extract signers.l1BlobSubmitterAddress from $PRECOMPUTED"
[[ "$PRECOMPUTED_L1_FINALIZATION_SUBMITTER" =~ ^0x[a-fA-F0-9]{40}$ ]] || die "Could not extract signers.l1FinalizationSubmitterAddress from $PRECOMPUTED"
[[ "$PRECOMPUTED_L2_MESSAGE_ANCHORING" =~ ^0x[a-fA-F0-9]{40}$ ]] || die "Could not extract signers.l2MessageAnchoringAddress from $PRECOMPUTED"

log "L1_CHAIN_ID=$L1_CHAIN_ID"
log "L1_DEPLOYER_ADDRESS=$L1_DEPLOYER_ADDRESS"
log "L1 blob submitter=$PRECOMPUTED_L1_BLOB_SUBMITTER"
log "L1 finalization submitter=$PRECOMPUTED_L1_FINALIZATION_SUBMITTER"
log "L2 message anchorer=$PRECOMPUTED_L2_MESSAGE_ANCHORING"
log "Precomputed LineaRollupV8: $PRECOMPUTED_LINEA_ROLLUP"
log "Precomputed L2MessageService: $PRECOMPUTED_L2_MS"

# ─────────────────────────────────────────────────────────────────────────────
# Query Shomei for the L2 genesis ZK state root + compute genesis shnarf
# ─────────────────────────────────────────────────────────────────────────────
#
# `eth_getBlockByNumber(0).stateRoot` is the Merkle-Patricia root, which is
# wrong for ZK proof verification — the L1 LineaRollup verifies against
# Shomei's ZK state root, accessible via rollup_getZkEVMStateMerkleProofV0.
# Shomei must be up; it depends on l2-node-besu so by the time this script
# runs (after sequencer + l2-node-besu healthy) Shomei is reachable.

step "Query Shomei for genesis state root + compute shnarf"

SHOMEI_URL="${SHOMEI_URL:-http://shomei:8888}"

# Wait for Shomei to be reachable + return data on block 0.
log "Waiting for Shomei at $SHOMEI_URL"
SHOMEI_RESP=""
for _ in $(seq 1 60); do
  SHOMEI_RESP=$(curl -fsS --max-time 5 -X POST -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"rollup_getZkEVMStateMerkleProofV0","params":[{"startBlockNumber":0,"endBlockNumber":0,"zkStateManagerVersion":"2.3.0"}]}' \
    "$SHOMEI_URL" 2>/dev/null || true)
  if [[ -n "$SHOMEI_RESP" ]] && echo "$SHOMEI_RESP" | grep -q "zkEndStateRootHash"; then
    break
  fi
  sleep 2
done
echo "$SHOMEI_RESP" | grep -q "zkEndStateRootHash" \
  || die "Shomei did not return zkEndStateRootHash within 120s. Response: $SHOMEI_RESP"

# Extract zkEndStateRootHash. POSIX sed -E (no jq in node:24-bookworm by default).
L2_GENESIS_STATE_ROOT=$(echo "$SHOMEI_RESP" | sed -nE 's/.*"zkEndStateRootHash":[[:space:]]*"(0x[0-9a-fA-F]{64})".*/\1/p' | head -1)
[[ "$L2_GENESIS_STATE_ROOT" =~ ^0x[0-9a-fA-F]{64}$ ]] \
  || die "Could not extract zkEndStateRootHash from Shomei response: $SHOMEI_RESP"
log "L2_GENESIS_STATE_ROOT (from Shomei): $L2_GENESIS_STATE_ROOT"

# Genesis shnarf — keccak256(parentShnarf || snarkHash || parentStateRootHash
#                              || evalClaim || evalPoint), each 32 bytes. Only
# parentStateRootHash is non-zero. Formula is the V6 shape per the original
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

DEFAULT_L1_OPERATOR_ADDRESSES="$L1_DEPLOYER_ADDRESS,$PRECOMPUTED_L1_BLOB_SUBMITTER,$PRECOMPUTED_L1_FINALIZATION_SUBMITTER"
L1_ROLE_MIN_BALANCE_WEI="${L1_ROLE_MIN_BALANCE_WEI:-10000000000000000}"      # 0.01 ETH
L1_ROLE_TOP_UP_WEI="${L1_ROLE_TOP_UP_WEI:-30000000000000000}"                # 0.03 ETH

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

# Verify-or-die: assert deployed address matches the precomputed one. Address
# comparison is case-insensitive (cast outputs checksummed; some scripts emit
# lowercase). Empty $expected (precomputed value missing) means the contract
# isn't tracked in addresses-precomputed.json — that's intentional for
# intermediate/library contracts (Mimc, AddressFilter, ProxyAdmin, etc).
verify_address() {
  local actual="$1" expected="$2" label="$3"
  if [[ -z "$expected" ]]; then
    log "verify $label: skipped (not tracked in precomputed JSON)"
    return 0
  fi
  if [[ "$(echo "$actual" | tr 'A-F' 'a-f')" != "$(echo "$expected" | tr 'A-F' 'a-f')" ]]; then
    die "ADDRESS MISMATCH: $label deployed=$actual  expected=$expected. \
The deployer's nonce sequence may have drifted from the offsets baked into account-setup.sh. \
Re-run with a clean Sepolia deployer or adjust scripts/account-setup.sh to match the current deploy script's nonce usage."
  fi
  log "verify $label: OK ($actual)"
}

# ─────────────────────────────────────────────────────────────────────────────
# Step functions — one per Make target
# ─────────────────────────────────────────────────────────────────────────────
#
# Idempotency: each step writes its log to /shared/deploy-logs/stepN-*.log on
# the linea-shared-config volume. If this script is re-run after a partial
# failure (e.g. step 5 failed mid-deploy), already-completed steps detect
# their prior log and skip re-deploying — otherwise the wallet's advanced
# nonce would deploy NEW contracts at NEW addresses, breaking the
# precomputed-address verify chain. To force a full re-deploy, run with a
# fresh shared volume (`down -v`).

# Returns 0 if the step's log file shows a successfully-deployed contract,
# 1 otherwise. Used by each step to decide whether to skip the deploy and
# just re-extract the address from the existing log.
step_already_done() {
  local logfile="$1" contract_name="$2"
  [[ -f "$logfile" ]] && grep -qE "^contract=${contract_name} deployed: " "$logfile"
}

# Step 1 — deploy-linea-rollup-v$L1_CONTRACT_VERSION (or deploy-validium-v2)
step1_l1_rollup() {
  step "Step 1: deploy L1 Verifier + LineaRollup (or Validium)"
  local script logfile primary_contract

  if [[ "$LINEA_COORDINATOR_DATA_AVAILABILITY" == "VALIDIUM" ]]; then
    script="$ART_DIR/deployPlonkVerifierAndValidiumV2.ts"
    logfile="$LOG_DIR/step1-validium.log"
    primary_contract="ValidiumV2"
  else
    script="$ART_DIR/deployPlonkVerifierAndLineaRollupV${L1_CONTRACT_VERSION}.ts"
    logfile="$LOG_DIR/step1-linea-rollup.log"
    primary_contract="LineaRollupV${L1_CONTRACT_VERSION}"
  fi

  if step_already_done "$logfile" "$primary_contract"; then
    log "Step 1: $logfile present — skipping deploy, re-using prior addresses"
    LINEA_ROLLUP_ADDRESS="$(require_address "$logfile" "$primary_contract")"
    export LINEA_ROLLUP_ADDRESS
    log "Forwarding LINEA_ROLLUP_ADDRESS=$LINEA_ROLLUP_ADDRESS"
    return 0
  fi

  [[ -f "$script" ]] || die "missing deploy script: $script"

  # The L1 deployer remains the admin/security-council account. Coordinator's
  # blob and finalization senders are separate derived operator addresses so
  # their independent nonce managers do not contend for one L1 account.
  if [[ "$LINEA_COORDINATOR_DATA_AVAILABILITY" == "VALIDIUM" ]]; then
    DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
    RPC_URL="$L1_RPC_URL" \
    VERIFIER_CONTRACT_NAME="IntegrationTestTrueVerifier" \
    INITIAL_L2_STATE_ROOT_HASH="$L2_GENESIS_STATE_ROOT" \
    INITIAL_L2_BLOCK_NUMBER="0" \
    L2_GENESIS_TIMESTAMP="$FORK_TIMESTAMP" \
    L1_SECURITY_COUNCIL="$L1_DEPLOYER_ADDRESS" \
    VALIDIUM_OPERATORS="${VALIDIUM_OPERATORS:-$DEFAULT_L1_OPERATOR_ADDRESSES}" \
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
    LINEA_ROLLUP_OPERATORS="${LINEA_ROLLUP_OPERATORS:-$DEFAULT_L1_OPERATOR_ADDRESSES}" \
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
    # Validium isn't tracked in addresses-precomputed.json (precomputed targets
    # the ROLLUP variant); skip verify in validium mode.
    log "verify ValidiumV2: skipped (validium variant not precomputed)"
  else
    LINEA_ROLLUP_ADDRESS="$(require_address "$logfile" "LineaRollupV${L1_CONTRACT_VERSION}")"
    verify_address "$LINEA_ROLLUP_ADDRESS" "$PRECOMPUTED_LINEA_ROLLUP" "LineaRollupV${L1_CONTRACT_VERSION}"
    # Also verify ForcedTransactionGateway — it's deployed in this same step.
    FORCED_TX_GW_ADDRESS="$(extract_address "$logfile" "ForcedTransactionGateway")" || FORCED_TX_GW_ADDRESS=""
    if [[ -n "$FORCED_TX_GW_ADDRESS" ]]; then
      verify_address "$FORCED_TX_GW_ADDRESS" "$PRECOMPUTED_FORCED_TX_GW" "ForcedTransactionGateway"
    fi
  fi
  export LINEA_ROLLUP_ADDRESS
  log "Forwarding LINEA_ROLLUP_ADDRESS=$LINEA_ROLLUP_ADDRESS"
}

# Step 2 — deploy-l2messageservice
step2_l2_message_service() {
  step "Step 2: deploy L2 MessageService"
  local logfile="$LOG_DIR/step2-l2-message-service.log"

  if step_already_done "$logfile" "L2MessageService"; then
    log "Step 2: $logfile present — skipping deploy"
    L2_MESSAGE_SERVICE_ADDRESS="$(require_address "$logfile" "L2MessageService")"
    export L2_MESSAGE_SERVICE_ADDRESS
    log "Forwarding L2_MESSAGE_SERVICE_ADDRESS=$L2_MESSAGE_SERVICE_ADDRESS"
    return 0
  fi

  L2_MESSAGE_SERVICE_CONTRACT_NAME="L2MessageService" \
  DEPLOYER_PRIVATE_KEY="$L2_DEPLOYER_PRIVATE_KEY" \
  RPC_URL="$L2_RPC_URL" \
  L2_SECURITY_COUNCIL="0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266" \
  L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER="${L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER:-$PRECOMPUTED_L2_MESSAGE_ANCHORING}" \
  L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD="86400" \
  L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT="1000000000000000000000" \
    pnpm -s exec ts-node "$ART_DIR/deployL2MessageServiceV1.ts" 2>&1 | tee "$logfile"

  L2_MESSAGE_SERVICE_ADDRESS="$(require_address "$logfile" "L2MessageService")"
  verify_address "$L2_MESSAGE_SERVICE_ADDRESS" "$PRECOMPUTED_L2_MS" "L2MessageService"
  export L2_MESSAGE_SERVICE_ADDRESS
  log "Forwarding L2_MESSAGE_SERVICE_ADDRESS=$L2_MESSAGE_SERVICE_ADDRESS"
}

# Step 3 — deploy-token-bridge-l1
step3_token_bridge_l1() {
  step "Step 3: deploy L1 TokenBridge + L1 BridgedToken"
  local logfile="$LOG_DIR/step3-token-bridge-l1.log"

  if step_already_done "$logfile" "TokenBridge"; then
    log "Step 3: $logfile present — skipping deploy"
    L1_TOKEN_BRIDGE_ADDRESS="$(require_address "$logfile" "TokenBridge")"
    export L1_TOKEN_BRIDGE_ADDRESS
    return 0
  fi

  : "${LINEA_ROLLUP_ADDRESS:?step1 must run first}"
  : "${L2_MESSAGE_SERVICE_ADDRESS:?step2 must run first}"

  # REMOTE_DEPLOYER_ADDRESS is the L2 deployer address (deterministic via CREATE
  # from the pre-baked L2 deployer at nonce 0). Stays hardcoded — we own L2.
  # L1_SECURITY_COUNCIL stays the deployer/admin account.
  # REMOTE_TOKEN_BRIDGE_ADDRESS is consumed by the FORKED deploy script
  # (scaffold's deployBridgedTokenAndTokenBridgeV1_1.ts, bind-mounted over the
  # upstream path). Replaces the upstream's stale-offset-based remoteSender
  # derivation with the precomputed L2 TokenBridge from account-setup.sh.
  DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
  REMOTE_DEPLOYER_ADDRESS="0x1B9AbEeC3215D8AdE8a33607f2cF0f4F60e5F0D0" \
  REMOTE_TOKEN_BRIDGE_ADDRESS="$PRECOMPUTED_L2_TOKEN_BRIDGE" \
  RPC_URL="$L1_RPC_URL" \
  REMOTE_CHAIN_ID="$L2_CHAIN_ID" \
  TOKEN_BRIDGE_L1="true" \
  L1_SECURITY_COUNCIL="$L1_DEPLOYER_ADDRESS" \
  L2_MESSAGE_SERVICE_ADDRESS="$L2_MESSAGE_SERVICE_ADDRESS" \
  LINEA_ROLLUP_ADDRESS="$LINEA_ROLLUP_ADDRESS" \
    pnpm -s exec ts-node "$ART_DIR/deployBridgedTokenAndTokenBridgeV1_1.ts" 2>&1 | tee "$logfile"

  L1_TOKEN_BRIDGE_ADDRESS="$(require_address "$logfile" "TokenBridge")"
  verify_address "$L1_TOKEN_BRIDGE_ADDRESS" "$PRECOMPUTED_L1_TOKEN_BRIDGE" "L1 TokenBridge"
  export L1_TOKEN_BRIDGE_ADDRESS
}

# Step 4 — deploy-token-bridge-l2
step4_token_bridge_l2() {
  step "Step 4: deploy L2 TokenBridge + L2 BridgedToken"
  local logfile="$LOG_DIR/step4-token-bridge-l2.log"

  if step_already_done "$logfile" "TokenBridge"; then
    log "Step 4: $logfile present — skipping deploy"
    return 0
  fi

  : "${LINEA_ROLLUP_ADDRESS:?step1 must run first}"
  : "${L2_MESSAGE_SERVICE_ADDRESS:?step2 must run first}"

  : "${L1_TOKEN_BRIDGE_ADDRESS:?step3 must run first}"

  # REMOTE_TOKEN_BRIDGE_ADDRESS = the L1 TokenBridge proxy we deployed in
  # step 3 (forwarded as the L2 TokenBridge's remoteSender during init).
  DEPLOYER_PRIVATE_KEY="$L2_DEPLOYER_PRIVATE_KEY" \
  REMOTE_DEPLOYER_ADDRESS="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" \
  REMOTE_TOKEN_BRIDGE_ADDRESS="$L1_TOKEN_BRIDGE_ADDRESS" \
  RPC_URL="$L2_RPC_URL" \
  REMOTE_CHAIN_ID="$L1_CHAIN_ID" \
  TOKEN_BRIDGE_L1="false" \
  L2_SECURITY_COUNCIL="0xf17f52151EbEF6C7334FAD080c5704D77216b732" \
  L2_MESSAGE_SERVICE_ADDRESS="$L2_MESSAGE_SERVICE_ADDRESS" \
  LINEA_ROLLUP_ADDRESS="$LINEA_ROLLUP_ADDRESS" \
    pnpm -s exec ts-node "$ART_DIR/deployBridgedTokenAndTokenBridgeV1_1.ts" 2>&1 | tee "$logfile"

  L2_TOKEN_BRIDGE_ADDRESS="$(require_address "$logfile" "TokenBridge")"
  verify_address "$L2_TOKEN_BRIDGE_ADDRESS" "$PRECOMPUTED_L2_TOKEN_BRIDGE" "L2 TokenBridge"
}

# Step 5 — deploy-l1-test-erc20
step5_l1_test_erc20() {
  step "Step 5: deploy L1 TestERC20"
  local logfile="$LOG_DIR/step5-l1-test-erc20.log"

  if step_already_done "$logfile" "TestERC20"; then
    log "Step 5: $logfile present — skipping deploy"
    L1_TEST_ERC20_ADDRESS="$(require_address "$logfile" "TestERC20")"
    verify_address "$L1_TEST_ERC20_ADDRESS" "$PRECOMPUTED_L1_TEST_ERC20" "L1 TestERC20"
    return 0
  fi

  # env -u L1_NONCE: deployTestERC20.ts has the same stale offset bug as the
  # upstream token-bridge script (ORDERED_NONCE_POST_LINEAROLLUP=7 vs actual 8).
  # With L1_NONCE unset, the script falls through to `await wallet.getNonce()`
  # which queries Sepolia for the live nonce.
  env -u L1_NONCE \
  DEPLOYER_PRIVATE_KEY="$L1_DEPLOYER_PRIVATE_KEY" \
  RPC_URL="$L1_RPC_URL" \
  TEST_ERC20_L1="true" \
  TEST_ERC20_NAME="TestERC20" \
  TEST_ERC20_SYMBOL="TERC20" \
  TEST_ERC20_INITIAL_SUPPLY="100000" \
    pnpm -s exec ts-node "$ART_DIR/deployTestERC20.ts" 2>&1 | tee "$logfile"

  L1_TEST_ERC20_ADDRESS="$(require_address "$logfile" "TestERC20")"
  verify_address "$L1_TEST_ERC20_ADDRESS" "$PRECOMPUTED_L1_TEST_ERC20" "L1 TestERC20"
}

# Step 6 — deploy-l2-test-erc20
step6_l2_test_erc20() {
  step "Step 6: deploy L2 TestERC20"
  local logfile="$LOG_DIR/step6-l2-test-erc20.log"

  if step_already_done "$logfile" "TestERC20"; then
    log "Step 6: $logfile present — skipping deploy"
    L2_TEST_ERC20_ADDRESS="$(require_address "$logfile" "TestERC20")"
    verify_address "$L2_TEST_ERC20_ADDRESS" "$PRECOMPUTED_L2_TEST_ERC20" "L2 TestERC20"
    return 0
  fi

  # env -u L2_NONCE for the same reason as step 5 — defensive: even though
  # ORDERED_NONCE_POST_L2MESSAGESERVICE=3 is correct upstream, ride the
  # `wallet.getNonce()` fallback for parity with step 5.
  env -u L2_NONCE \
  DEPLOYER_PRIVATE_KEY="$L2_DEPLOYER_PRIVATE_KEY" \
  RPC_URL="$L2_RPC_URL" \
  TEST_ERC20_L1="false" \
  TEST_ERC20_NAME="TestERC20" \
  TEST_ERC20_SYMBOL="TERC20" \
  TEST_ERC20_INITIAL_SUPPLY="100000" \
    pnpm -s exec ts-node "$ART_DIR/deployTestERC20.ts" 2>&1 | tee "$logfile"

  L2_TEST_ERC20_ADDRESS="$(require_address "$logfile" "TestERC20")"
  verify_address "$L2_TEST_ERC20_ADDRESS" "$PRECOMPUTED_L2_TEST_ERC20" "L2 TestERC20"
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
# Fund derived L1 submitter accounts
# ─────────────────────────────────────────────────────────────────────────────

wei_lt() {
  node -e 'process.exit(BigInt(process.argv[1]) < BigInt(process.argv[2]) ? 0 : 1)' "$1" "$2"
}

fund_l1_submitter() {
  local name="$1" addr="$2" balance
  if [[ "${addr,,}" == "${L1_DEPLOYER_ADDRESS,,}" ]]; then
    log "Funding $name skipped: address is the L1 deployer"
    return 0
  fi

  balance="$(cast balance "$addr" --rpc-url "$L1_RPC_URL" | awk '{print $NF}')"
  [[ "$balance" =~ ^[0-9]+$ ]] || die "Could not read $name balance for $addr: $balance"
  log "$name balance: $balance wei at $addr"

  if wei_lt "$balance" "$L1_ROLE_MIN_BALANCE_WEI"; then
    log "Funding $name with $L1_ROLE_TOP_UP_WEI wei from L1 deployer"
    cast send "$addr" --value "$L1_ROLE_TOP_UP_WEI" --private-key "$L1_DEPLOYER_PRIVATE_KEY" --rpc-url "$L1_RPC_URL" >/dev/null
    balance="$(cast balance "$addr" --rpc-url "$L1_RPC_URL" | awk '{print $NF}')"
    log "$name balance after funding: $balance wei"
  else
    log "Funding $name skipped: balance already >= $L1_ROLE_MIN_BALANCE_WEI wei"
  fi
}

step "Fund derived L1 submitter accounts"
fund_l1_submitter "L1 blob submitter" "$PRECOMPUTED_L1_BLOB_SUBMITTER"
fund_l1_submitter "L1 finalization submitter" "$PRECOMPUTED_L1_FINALIZATION_SUBMITTER"

# ─────────────────────────────────────────────────────────────────────────────
# Aggregate addresses → addresses.json
# ─────────────────────────────────────────────────────────────────────────────

step "Aggregate addresses → $OUT_PATH"

node - <<'NODE_EOF' "$LOG_DIR" "$OUT_PATH" "$L1_CHAIN_ID" "$L2_CHAIN_ID" "$L2_RPC_URL"
const fs = require("fs");
const path = require("path");
const [, , logDir, outPath, l1ChainId, l2ChainId, l2Url] = process.argv;

// Format emitted by every deploy script (see contracts/scripts/hardhat/utils.ts:146):
//   contract=NAME deployed: address=0xADDR blockNumber=N chainId=Z
const re = /^contract=(\S+)\s+deployed:\s+address=(0x[a-fA-F0-9]{40})\s+blockNumber=(\d+)\s+chainId=(\d+)/;

const result = {
  _meta: {
    l1ChainId,
    l2ChainId,
    l1RpcUrl: "<redacted>",
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
# Patch rendered coordinator-config with deploy-time-only values
# ─────────────────────────────────────────────────────────────────────────────
#
# After Phase 2.3, config-render fills LINEA_ROLLUP_ADDRESS + L2_MESSAGE_SERVICE_ADDRESS
# at boot from /shared/addresses-precomputed.json. Maru's contract-address line
# is also already correct at boot. So we ONLY patch the values that genuinely
# can't be known until deploy time:
#
#   - genesis-state-root-hash  ← from Shomei (queried above)
#   - genesis-shnarf           ← computed from state root above
#   - contract-deployment-block-number  ← step-2 log (L2 block where
#                                          L2MessageService deployed)
#
# coordinator depends_on `deploy-contracts:service_completed_successfully`, so
# coordinator starts AFTER these patches land — first-boot reads patched values
# without needing a restart.

step "Patch rendered coordinator-config.toml with deploy-time values"

if [[ ! -d "$RENDERED_DIR" ]]; then
  log "WARNING: $RENDERED_DIR not mounted; skipping rendered-config patch"
elif [[ ! -f "$RENDERED_DIR/coordinator-config.toml" ]]; then
  log "WARNING: $RENDERED_DIR/coordinator-config.toml missing; skipping patch"
else
  # L2 block where L2MessageService deployed — extracted from step 2 log.
  L2_MS_DEPLOY_BLOCK=""
  if [[ -f "$LOG_DIR/step2-l2-message-service.log" ]]; then
    L2_MS_DEPLOY_BLOCK="$(grep -E "^contract=L2MessageService deployed: " "$LOG_DIR/step2-l2-message-service.log" | tail -1 | sed -nE 's/.*blockNumber=([0-9]+).*/\1/p')"
  fi
  : "${L2_MS_DEPLOY_BLOCK:=0}"

  log "Patching $RENDERED_DIR/coordinator-config.toml: state_root=$L2_GENESIS_STATE_ROOT shnarf=$L2_GENESIS_SHNARF deploy_block=$L2_MS_DEPLOY_BLOCK"
  awk \
    -v state_root="$L2_GENESIS_STATE_ROOT" \
    -v shnarf="$L2_GENESIS_SHNARF" \
    -v deploy_block="$L2_MS_DEPLOY_BLOCK" \
  '
    /^genesis-state-root-hash[[:space:]]*=/ { sub(/".*"/, "\"" state_root "\""); }
    /^genesis-shnarf[[:space:]]*=/          { sub(/".*"/, "\"" shnarf "\""); }
    /^contract-deployment-block-number[[:space:]]*=/ { sub(/=.*/, "= " deploy_block); }
    { print }
  ' "$RENDERED_DIR/coordinator-config.toml" > "$RENDERED_DIR/coordinator-config.toml.new" \
    && mv "$RENDERED_DIR/coordinator-config.toml.new" "$RENDERED_DIR/coordinator-config.toml"
  log "Patched coordinator-config.toml"
fi

step "Done"
log "addresses.json at $OUT_PATH"
log "step logs at $LOG_DIR/"
