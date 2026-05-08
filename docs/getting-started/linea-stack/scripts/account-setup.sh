#!/usr/bin/env sh
#
# account-setup.sh — Linea stack v0 pre-boot account derivation + address
# pre-computation (Phase 2.1 of the Sepolia migration).
#
# POSIX sh, not bash — runs in the foundry-rs/foundry alpine image which
# doesn't ship bash by default.
#
# Runs as the very first init container, BEFORE config-render and
# l2-genesis-init. Its outputs feed every later step:
#
#   1. /shared/addresses-precomputed.json — the canonical map of
#      "what-address-will-end-up-where" once deploy-contracts runs. Genesis
#      pre-funding, coordinator/maru/postman config rendering, and
#      deploy-contracts' verify-or-die check all read from this file.
#
#   2. /shared/web3signer-keys/*.yaml — four `file-raw` keystore configs
#      generated from L1_DEPLOYER_PRIVATE_KEY (3 L1 signers: anchoring,
#      data-submission, finalization) plus a static L2 dev key for the
#      liveness signer. Web3signer mounts this directory at /shared and
#      reads keystores from /shared/web3signer-keys/.
#
# Address pre-computation is via CREATE rules: address = keccak(rlp(sender, nonce))[12:].
# The nonce offsets below are derived from the EXISTING deploy-contracts.sh
# sequence (NOT the playbook's preferred Timelock-first sequence). If you
# change the deploy script's contract order, update both this file and the
# verify-or-die check in deploy-contracts.sh.
#
# Per Option A: a single user-supplied L1 key (L1_DEPLOYER_PRIVATE_KEY) drives
# every L1 role — deployer, security council, rollup operators, all three
# web3signer-backed signers (anchoring submits L1 txs to LineaRollup; the
# misleading "for anchoring on L2" comment in the original keystore was
# wrong about which chain it signs for).
#
# DEV ONLY notes:
#   - L2 deployer key + liveness signer key are pre-baked, public-knowledge
#     dev keys. Same as L2 genesis pre-funded accounts. NEVER reuse on mainnet.
#
set -eu
# pipefail isn't POSIX; avoid it. We check exit codes after every cast call
# below by validating the output shape with grep -qE.

# ─────────────────────────────────────────────────────────────────────────────
# Inputs (compose passes these in via environment)
# ─────────────────────────────────────────────────────────────────────────────

: "${L1_RPC_URL:?L1_RPC_URL must be set (Sepolia HTTPS RPC URL)}"
: "${L1_DEPLOYER_PRIVATE_KEY:?L1_DEPLOYER_PRIVATE_KEY must be set (Sepolia-funded)}"
# L2 deployer is pre-baked dev (we own L2 genesis); override only for testing.
: "${L2_DEPLOYER_PRIVATE_KEY:=0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae}"
# L2 liveness signer key is pre-baked dev (signs L2 txs only — never L1).
: "${L2_LIVENESS_SIGNER_PRIVATE_KEY:=0x234d87442cf7d43841fbe280febcdfabfb646added67bc19f7e42a5483f614c4}"

OUT_JSON="/shared/addresses-precomputed.json"
OUT_KEYS_DIR="/shared/web3signer-keys"
mkdir -p "$(dirname "$OUT_JSON")" "$OUT_KEYS_DIR"

log()  { printf "[account-setup] %s\n" "$*"; }
die()  { printf "[account-setup] ERROR: %s\n" "$*" >&2; exit 1; }

# Idempotency: if the precomputed JSON exists, skip re-derivation. Critical
# for retry cycles — if deploy-contracts is restarted (e.g. step N failed and
# we want to retry from N), compose's dependency chain re-fires account-setup
# too. With a fresh L1 deployer nonce (now advanced by prior partial deploys)
# the regenerated precomputed addresses would no longer match the contracts
# already deployed at the OLD baseline. A future `down -v` clears the volume
# and forces a fresh derivation.
if [ -f "$OUT_JSON" ]; then
  log "$OUT_JSON already exists — skipping (use 'down -v' to force fresh derivation)"
  exit 0
fi

# ─────────────────────────────────────────────────────────────────────────────
# Derive deployer addresses
# ─────────────────────────────────────────────────────────────────────────────
#
# `cast wallet address` outputs the checksummed address as a single line, no
# label. We pipe through awk '{print $NF}' defensively — some Foundry versions
# decorate output with a "Address:" prefix.

L1_DEPLOYER_ADDR="$(cast wallet address --private-key "$L1_DEPLOYER_PRIVATE_KEY" | awk '{print $NF}')"
L2_DEPLOYER_ADDR="$(cast wallet address --private-key "$L2_DEPLOYER_PRIVATE_KEY" | awk '{print $NF}')"

# Uncompressed secp256k1 public key (0x + 128 hex). Coordinator config
# pins specific public keys for each web3signer-backed signer slot
# (data-submission, finalization, anchoring); under Option A all three
# resolve to the user's L1 deployer key, so we substitute this single value
# into all three pubkey lines in the rendered coordinator-config.toml.
L1_DEPLOYER_PUBKEY="$(cast wallet public-key --private-key "$L1_DEPLOYER_PRIVATE_KEY" | awk '{print $NF}')"

is_addr()   { printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{40}$'; }
is_pubkey() { printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{128}$'; }
is_addr "$L1_DEPLOYER_ADDR" || die "L1 deployer address malformed: $L1_DEPLOYER_ADDR"
is_addr "$L2_DEPLOYER_ADDR" || die "L2 deployer address malformed: $L2_DEPLOYER_ADDR"
is_pubkey "$L1_DEPLOYER_PUBKEY" || die "L1 deployer pubkey malformed: $L1_DEPLOYER_PUBKEY"

log "L1 deployer: $L1_DEPLOYER_ADDR"
log "L2 deployer: $L2_DEPLOYER_ADDR"
log "L1 signer pubkey: $L1_DEPLOYER_PUBKEY"

# ─────────────────────────────────────────────────────────────────────────────
# Query L1 deployer nonce + L1 chain ID
# ─────────────────────────────────────────────────────────────────────────────
#
# Sepolia nonce can be non-zero if this deployer has prior history. L2 nonce
# is always 0 on a fresh chain (we control L2 genesis). If the user re-runs
# the stack against the same Sepolia deployer, the nonce will have advanced;
# pre-computed addresses will differ from the previous run. That's expected.

# Wait for Sepolia RPC to be reachable. Public RPCs sometimes 503 under load.
log "Waiting for L1 RPC (redacted)"
for _ in $(seq 1 30); do
  if cast chain-id --rpc-url "$L1_RPC_URL" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
cast chain-id --rpc-url "$L1_RPC_URL" >/dev/null 2>&1 || die "L1 RPC not reachable after 60s"

L1_CHAIN_ID="$(cast chain-id --rpc-url "$L1_RPC_URL" | awk '{print $NF}')"
L1_DEPLOYER_NONCE="$(cast nonce "$L1_DEPLOYER_ADDR" --rpc-url "$L1_RPC_URL" | awk '{print $NF}')"
L2_DEPLOYER_NONCE=0

log "L1 chain ID: $L1_CHAIN_ID"
log "L1 deployer start nonce: $L1_DEPLOYER_NONCE"

# Sanity: nonce must be a non-negative integer.
printf '%s\n' "$L1_DEPLOYER_NONCE" | grep -qE '^[0-9]+$' \
  || die "L1 deployer nonce malformed: $L1_DEPLOYER_NONCE"

# ─────────────────────────────────────────────────────────────────────────────
# Pre-compute contract addresses (CREATE rules)
# ─────────────────────────────────────────────────────────────────────────────
#
# Nonce sequence is derived from the EXISTING deploy-contracts.sh:
#
#   Step 1 (L1 — deployPlonkVerifierAndLineaRollupV8.ts) consumes 8 nonces:
#   7 contract deploys (Verifier, LineaRollupV8Implementation, ProxyAdmin,
#   AddressFilter, Mimc, LineaRollupV8 proxy, ForcedTransactionGateway) plus
#   1 role-grant tx for FORCED_TRANSACTION_SENDER_ROLE. Verified empirically
#   against Sepolia at 2026-05-07: with offset +7 the next deploy hits
#   "nonce too low: next nonce N+8, tx nonce N+7" (bringup-notes #29).
#
#   The on-chain nonce order ALSO differs from the script's emission order —
#   the proxy lands at nonce N+4 and the Mimc library at nonce N+5 (the deploy
#   script reorders tx broadcast vs log emission).
#
#   Step 2 (L2 — deployL2MessageServiceV1.ts) deploys 3:
#     L2MessageServiceImplementation, ProxyAdmin, L2MessageService (proxy)
#
#   Step 3 (L1 — deployBridgedTokenAndTokenBridgeV1_1.ts) deploys 5:
#     BridgedToken, tokenBridgeContractImplementation, ProxyAdmin,
#     UpgradeableBeacon, TokenBridge (proxy)
#
#   Step 4 (L2 — same script) deploys 5 more L2 contracts.
#   Step 5 (L1 — deployTestERC20.ts) deploys L1 TestERC20.
#   Step 6 (L2 — same) deploys L2 TestERC20.
#
# CRITICAL: deploy-contracts.sh's verify-or-die step (Phase 2.4) checks each
# deployed address against the values in this JSON. If the deploy script
# changes contract order or adds intermediate txs (approvals, etc), update
# the offsets here AND the verify check.

compute() {
  # $1 = deployer address, $2 = nonce
  cast compute-address "$1" --nonce "$2" | awk '{print $NF}'
}

# L1 contracts (offsets from $L1_DEPLOYER_NONCE).
# +0..+3:  Verifier, LineaRollupV8Implementation, ProxyAdmin, AddressFilter
# +4:      LineaRollupV8 (proxy)  ← script EMITS this 6th but on-chain it's nonce+4
# +5:      Mimc (library)         ← emitted 5th but on-chain it's nonce+5
# +6:      ForcedTransactionGateway
# +7:      grantRole(FORCED_TRANSACTION_SENDER_ROLE) — no contract deployed
# +8..+12: step 3 (BridgedToken, tokenBridgeContractImplementation, ProxyAdmin,
#                  UpgradeableBeacon, TokenBridge proxy)
# +13:     step 5 TestERC20
L1_VERIFIER="$(compute "$L1_DEPLOYER_ADDR" "$L1_DEPLOYER_NONCE")"
L1_LINEA_ROLLUP="$(compute "$L1_DEPLOYER_ADDR" "$((L1_DEPLOYER_NONCE + 4))")"
L1_FORCED_TX_GATEWAY="$(compute "$L1_DEPLOYER_ADDR" "$((L1_DEPLOYER_NONCE + 6))")"
L1_BRIDGED_TOKEN="$(compute "$L1_DEPLOYER_ADDR" "$((L1_DEPLOYER_NONCE + 8))")"
L1_TOKEN_BRIDGE="$(compute "$L1_DEPLOYER_ADDR" "$((L1_DEPLOYER_NONCE + 12))")"
L1_TEST_ERC20="$(compute "$L1_DEPLOYER_ADDR" "$((L1_DEPLOYER_NONCE + 13))")"

# L2 contracts (offsets from 0; fresh chain).
# Step 2: 3 contracts (impl @0, ProxyAdmin @1, proxy @2).
# Step 4: 5 contracts (BridgedToken @3, impl @4, ProxyAdmin @5,
#                      UpgradeableBeacon @6, TokenBridge proxy @7).
# Step 6: TestERC20 @8.
L2_MESSAGE_SERVICE="$(compute "$L2_DEPLOYER_ADDR" 2)"
L2_BRIDGED_TOKEN="$(compute "$L2_DEPLOYER_ADDR" 3)"
L2_TOKEN_BRIDGE="$(compute "$L2_DEPLOYER_ADDR" 7)"
L2_TEST_ERC20="$(compute "$L2_DEPLOYER_ADDR" 8)"

# Sanity: every output must be a 0x-prefixed 40-hex-char address.
for addr in "$L1_VERIFIER" "$L1_LINEA_ROLLUP" "$L1_FORCED_TX_GATEWAY" \
            "$L1_BRIDGED_TOKEN" "$L1_TOKEN_BRIDGE" "$L1_TEST_ERC20" \
            "$L2_MESSAGE_SERVICE" "$L2_BRIDGED_TOKEN" "$L2_TOKEN_BRIDGE" "$L2_TEST_ERC20"; do
  is_addr "$addr" || die "Pre-computed address malformed: $addr"
done

log "Pre-computed L1 LineaRollupV8 (proxy): $L1_LINEA_ROLLUP"
log "Pre-computed L2 MessageService: $L2_MESSAGE_SERVICE"

# ─────────────────────────────────────────────────────────────────────────────
# Write addresses-precomputed.json
# ─────────────────────────────────────────────────────────────────────────────

cat > "$OUT_JSON" <<EOF
{
  "_meta": {
    "generatedAt": "$(date -u +%FT%TZ)",
    "l1RpcUrl": "<redacted>",
    "l1ChainId": "$L1_CHAIN_ID",
    "l1DeployerStartNonce": $L1_DEPLOYER_NONCE,
    "l2DeployerStartNonce": $L2_DEPLOYER_NONCE,
    "deployScriptVersion": "v0-phase2.1"
  },
  "deployers": {
    "l1": "$L1_DEPLOYER_ADDR",
    "l2": "$L2_DEPLOYER_ADDR"
  },
  "signers": {
    "l1Pubkey": "$L1_DEPLOYER_PUBKEY"
  },
  "l1": {
    "Verifier": "$L1_VERIFIER",
    "LineaRollupV8": "$L1_LINEA_ROLLUP",
    "ForcedTransactionGateway": "$L1_FORCED_TX_GATEWAY",
    "BridgedToken": "$L1_BRIDGED_TOKEN",
    "TokenBridge": "$L1_TOKEN_BRIDGE",
    "TestERC20": "$L1_TEST_ERC20"
  },
  "l2": {
    "L2MessageService": "$L2_MESSAGE_SERVICE",
    "BridgedToken": "$L2_BRIDGED_TOKEN",
    "TokenBridge": "$L2_TOKEN_BRIDGE",
    "TestERC20": "$L2_TEST_ERC20"
  }
}
EOF

log "Wrote $OUT_JSON"

# ─────────────────────────────────────────────────────────────────────────────
# Generate web3signer file-raw keystore YAMLs
# ─────────────────────────────────────────────────────────────────────────────
#
# 3 L1 signers use L1_DEPLOYER_PRIVATE_KEY (Option A — single key per Phase
# 2.1). The 4th (liveness) signs L2 txs only and stays on the pre-baked dev
# key. Format: file-raw YAML — see Web3signer docs for `eth1` signer config.

write_keystore() {
  # $1 = filename without .yaml, $2 = role description, $3 = private key
  # POSIX sh — no `local`; we use a globally-named scratch var.
  _ks_file="$OUT_KEYS_DIR/$1.yaml"
  cat > "$_ks_file" <<EOF
# ============================================================
# DEV ONLY — generated at boot by account-setup.sh.
# Re-rendered every boot from .env / pre-baked dev keys; do NOT edit.
# ============================================================
type: "file-raw"
keyType: "SECP256K1"
# $2
privateKey: "$3"
EOF
  log "Wrote $_ks_file"
}

write_keystore "anchoring-signer"        "L1 anchoring signer (Option A: same as L1 deployer)"        "$L1_DEPLOYER_PRIVATE_KEY"
write_keystore "data-submission-signer"  "L1 blob/data-submission signer (Option A: same as L1 deployer)" "$L1_DEPLOYER_PRIVATE_KEY"
write_keystore "finalization-signer"     "L1 aggregation/finalization signer (Option A: same as L1 deployer)" "$L1_DEPLOYER_PRIVATE_KEY"
write_keystore "liveness-signer"         "L2 sequencer-liveness signer (pre-baked dev key, L2 only)"   "$L2_LIVENESS_SIGNER_PRIVATE_KEY"

log "Done."
