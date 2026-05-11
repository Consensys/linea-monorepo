#!/usr/bin/env sh
#
# account-setup.sh — Linea stack v0 pre-boot account/key generation + address
# pre-computation.
#
# POSIX sh, not bash — runs in the foundry-rs/foundry alpine image.
#
# The user supplies exactly one secret: a Sepolia-funded L1 deployer key. This
# script generates fresh runtime keys for the remaining roles, persists them in
# the shared Docker volume for retry-safe restarts, derives addresses/pubkeys,
# and precomputes the contract addresses needed by the first boot flow.
#
# Outputs:
#   1. /shared/addresses-precomputed.json — canonical pre-boot address map.
#   2. /shared/runtime-keys.env — generated private keys for deploy/postman.
#   3. /shared/web3signer-keys/*.yaml — web3signer file-raw keystores for
#      coordinator signers: L1 blob/data-submission, L1 finalization, and L2
#      message anchoring.
#
# Address pre-computation is via CREATE rules: address = keccak(rlp(sender, nonce))[12:].
# The nonce offsets below mirror deploy-contracts.sh's current deploy order.
# If that order changes, update this file and deploy-contracts.sh's verify step.
#
set -eu

# pipefail is not POSIX; validate command outputs explicitly instead.

: "${L1_RPC_URL:?L1_RPC_URL must be set (Sepolia HTTPS RPC URL)}"
: "${L1_DEPLOYER_PRIVATE_KEY:?L1_DEPLOYER_PRIVATE_KEY must be set (Sepolia-funded)}"

OUT_JSON="/shared/addresses-precomputed.json"
OUT_RUNTIME_KEYS_ENV="/shared/runtime-keys.env"
OUT_KEYS_DIR="/shared/web3signer-keys"
mkdir -p "$(dirname "$OUT_JSON")" "$OUT_KEYS_DIR"

log() { printf "[account-setup] %s\n" "$*"; }
die() { printf "[account-setup] ERROR: %s\n" "$*" >&2; exit 1; }

is_addr()    { printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{40}$'; }
is_pubkey()  { printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{128}$'; }
is_privkey() { printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{64}$'; }

require_file() {
  [ -f "$1" ] || die "$1 missing; remove the shared volume with 'docker compose ... down -v' and retry"
}

require_json_field() {
  grep -q "\"$1\"" "$OUT_JSON" || die "$OUT_JSON missing required field $1; remove the shared volume with 'docker compose ... down -v' and retry"
}

# Idempotency: if the precomputed JSON already exists, every downstream address
# and private key must be reused. Recomputing after partial deploys would use the
# deployer's advanced nonce and drift away from already-created contracts.
if [ -f "$OUT_JSON" ]; then
  for field in \
    l1BlobSubmitterAddress l1FinalizationSubmitterAddress l1PostmanAddress \
    l2DeployerAddress l2MessageAnchoringAddress l2PostmanAddress \
    l1BlobSubmitterPubkey l1FinalizationSubmitterPubkey l2MessageAnchoringPubkey; do
    require_json_field "$field"
  done
  require_file "$OUT_RUNTIME_KEYS_ENV"
  chmod 0644 "$OUT_RUNTIME_KEYS_ENV" || die "failed to chmod $OUT_RUNTIME_KEYS_ENV"
  require_file "$OUT_KEYS_DIR/anchoring-signer.yaml"
  require_file "$OUT_KEYS_DIR/data-submission-signer.yaml"
  require_file "$OUT_KEYS_DIR/finalization-signer.yaml"
  chmod 0644 "$OUT_KEYS_DIR/anchoring-signer.yaml" "$OUT_KEYS_DIR/data-submission-signer.yaml" "$OUT_KEYS_DIR/finalization-signer.yaml" || die "failed to chmod web3signer key files"
  log "$OUT_JSON already exists — reusing generated runtime keys (use 'down -v' to force fresh keys)"
  exit 0
fi

new_private_key() {
  cast wallet new --json | sed -nE 's/.*"private_key"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/p' | head -1
}

load_or_generate_runtime_keys() {
  if [ -f "$OUT_RUNTIME_KEYS_ENV" ]; then
    log "Reusing generated runtime keys from $OUT_RUNTIME_KEYS_ENV"
    # shellcheck disable=SC1090
    . "$OUT_RUNTIME_KEYS_ENV"
  else
    log "Generating fresh runtime keys"
    L1_BLOB_SUBMITTER_PRIVATE_KEY="$(new_private_key)"
    L1_FINALIZATION_SUBMITTER_PRIVATE_KEY="$(new_private_key)"
    L1_POSTMAN_PRIVATE_KEY="$(new_private_key)"
    L2_DEPLOYER_PRIVATE_KEY="$(new_private_key)"
    L2_MESSAGE_ANCHORING_PRIVATE_KEY="$(new_private_key)"
    L2_POSTMAN_PRIVATE_KEY="$(new_private_key)"
  fi

  for item in \
    L1_BLOB_SUBMITTER_PRIVATE_KEY \
    L1_FINALIZATION_SUBMITTER_PRIVATE_KEY \
    L1_POSTMAN_PRIVATE_KEY \
    L2_DEPLOYER_PRIVATE_KEY \
    L2_MESSAGE_ANCHORING_PRIVATE_KEY \
    L2_POSTMAN_PRIVATE_KEY; do
    eval "_value=\${$item:-}"
    is_privkey "$_value" || die "$item malformed or missing"
  done

  if [ ! -f "$OUT_RUNTIME_KEYS_ENV" ]; then
    umask 077
    tmp="$OUT_RUNTIME_KEYS_ENV.tmp"
    {
      printf "L1_BLOB_SUBMITTER_PRIVATE_KEY='%s'\n" "$L1_BLOB_SUBMITTER_PRIVATE_KEY"
      printf "L1_FINALIZATION_SUBMITTER_PRIVATE_KEY='%s'\n" "$L1_FINALIZATION_SUBMITTER_PRIVATE_KEY"
      printf "L1_POSTMAN_PRIVATE_KEY='%s'\n" "$L1_POSTMAN_PRIVATE_KEY"
      printf "L2_DEPLOYER_PRIVATE_KEY='%s'\n" "$L2_DEPLOYER_PRIVATE_KEY"
      printf "L2_MESSAGE_ANCHORING_PRIVATE_KEY='%s'\n" "$L2_MESSAGE_ANCHORING_PRIVATE_KEY"
      printf "L2_POSTMAN_PRIVATE_KEY='%s'\n" "$L2_POSTMAN_PRIVATE_KEY"
    } > "$tmp"
    mv "$tmp" "$OUT_RUNTIME_KEYS_ENV"
    log "Wrote $OUT_RUNTIME_KEYS_ENV"
  fi

  chmod 0644 "$OUT_RUNTIME_KEYS_ENV" || die "failed to chmod $OUT_RUNTIME_KEYS_ENV"
}

load_or_generate_runtime_keys

wallet_address() {
  cast wallet address --private-key "$1" | awk '{print $NF}'
}

wallet_pubkey() {
  cast wallet public-key --private-key "$1" | awk '{print $NF}'
}

L1_DEPLOYER_ADDR="$(wallet_address "$L1_DEPLOYER_PRIVATE_KEY")"
L2_DEPLOYER_ADDR="$(wallet_address "$L2_DEPLOYER_PRIVATE_KEY")"

L1_DEPLOYER_PUBKEY="$(wallet_pubkey "$L1_DEPLOYER_PRIVATE_KEY")"
L1_BLOB_SUBMITTER_ADDR="$(wallet_address "$L1_BLOB_SUBMITTER_PRIVATE_KEY")"
L1_BLOB_SUBMITTER_PUBKEY="$(wallet_pubkey "$L1_BLOB_SUBMITTER_PRIVATE_KEY")"
L1_FINALIZATION_SUBMITTER_ADDR="$(wallet_address "$L1_FINALIZATION_SUBMITTER_PRIVATE_KEY")"
L1_FINALIZATION_SUBMITTER_PUBKEY="$(wallet_pubkey "$L1_FINALIZATION_SUBMITTER_PRIVATE_KEY")"
L1_POSTMAN_ADDR="$(wallet_address "$L1_POSTMAN_PRIVATE_KEY")"
L1_POSTMAN_PUBKEY="$(wallet_pubkey "$L1_POSTMAN_PRIVATE_KEY")"
L2_DEPLOYER_PUBKEY="$(wallet_pubkey "$L2_DEPLOYER_PRIVATE_KEY")"
L2_MESSAGE_ANCHORING_ADDR="$(wallet_address "$L2_MESSAGE_ANCHORING_PRIVATE_KEY")"
L2_MESSAGE_ANCHORING_PUBKEY="$(wallet_pubkey "$L2_MESSAGE_ANCHORING_PRIVATE_KEY")"
L2_POSTMAN_ADDR="$(wallet_address "$L2_POSTMAN_PRIVATE_KEY")"
L2_POSTMAN_PUBKEY="$(wallet_pubkey "$L2_POSTMAN_PRIVATE_KEY")"

for addr in \
  "$L1_DEPLOYER_ADDR" "$L2_DEPLOYER_ADDR" "$L1_BLOB_SUBMITTER_ADDR" \
  "$L1_FINALIZATION_SUBMITTER_ADDR" "$L1_POSTMAN_ADDR" \
  "$L2_MESSAGE_ANCHORING_ADDR" "$L2_POSTMAN_ADDR"; do
  is_addr "$addr" || die "derived address malformed: $addr"
done

for pubkey in \
  "$L1_DEPLOYER_PUBKEY" "$L1_BLOB_SUBMITTER_PUBKEY" \
  "$L1_FINALIZATION_SUBMITTER_PUBKEY" "$L1_POSTMAN_PUBKEY" \
  "$L2_DEPLOYER_PUBKEY" "$L2_MESSAGE_ANCHORING_PUBKEY" \
  "$L2_POSTMAN_PUBKEY"; do
  is_pubkey "$pubkey" || die "derived pubkey malformed: $pubkey"
done

log "L1 deployer: $L1_DEPLOYER_ADDR"
log "L1 blob submitter: $L1_BLOB_SUBMITTER_ADDR"
log "L1 finalization submitter: $L1_FINALIZATION_SUBMITTER_ADDR"
log "L1 postman: $L1_POSTMAN_ADDR"
log "L2 deployer: $L2_DEPLOYER_ADDR"
log "L2 message anchorer: $L2_MESSAGE_ANCHORING_ADDR"
log "L2 postman: $L2_POSTMAN_ADDR"

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
printf '%s\n' "$L1_DEPLOYER_NONCE" | grep -qE '^[0-9]+$' || die "L1 deployer nonce malformed: $L1_DEPLOYER_NONCE"

compute() {
  cast compute-address "$1" --nonce "$2" | awk '{print $NF}'
}

# L1 offsets mirror the current deploy-contracts.sh sequence. We still track
# Only these two contract addresses are required before L2 boot/config render.
# Other contracts are captured after deployment in /shared/addresses.json.
L1_LINEA_ROLLUP="$(compute "$L1_DEPLOYER_ADDR" "$((L1_DEPLOYER_NONCE + 4))")"
L2_MESSAGE_SERVICE="$(compute "$L2_DEPLOYER_ADDR" 2)"

for addr in "$L1_LINEA_ROLLUP" "$L2_MESSAGE_SERVICE"; do
  is_addr "$addr" || die "Pre-computed address malformed: $addr"
done

log "Pre-computed L1 LineaRollupV8 (proxy): $L1_LINEA_ROLLUP"
log "Pre-computed L2 MessageService: $L2_MESSAGE_SERVICE"

cat > "$OUT_JSON" <<JSON_EOF
{
  "_meta": {
    "generatedAt": "$(date -u +%FT%TZ)",
    "l1RpcUrl": "<redacted>",
    "l1ChainId": "$L1_CHAIN_ID",
    "l1DeployerStartNonce": $L1_DEPLOYER_NONCE,
    "l2DeployerStartNonce": $L2_DEPLOYER_NONCE,
    "deployScriptVersion": "v0-runtime-keys"
  },
  "deployers": {
    "l1": "$L1_DEPLOYER_ADDR",
    "l2": "$L2_DEPLOYER_ADDR"
  },
  "signers": {
    "l1Pubkey": "$L1_DEPLOYER_PUBKEY",
    "l1BlobSubmitterAddress": "$L1_BLOB_SUBMITTER_ADDR",
    "l1BlobSubmitterPubkey": "$L1_BLOB_SUBMITTER_PUBKEY",
    "l1FinalizationSubmitterAddress": "$L1_FINALIZATION_SUBMITTER_ADDR",
    "l1FinalizationSubmitterPubkey": "$L1_FINALIZATION_SUBMITTER_PUBKEY",
    "l1PostmanAddress": "$L1_POSTMAN_ADDR",
    "l1PostmanPubkey": "$L1_POSTMAN_PUBKEY",
    "l2DeployerAddress": "$L2_DEPLOYER_ADDR",
    "l2DeployerPubkey": "$L2_DEPLOYER_PUBKEY",
    "l2MessageAnchoringAddress": "$L2_MESSAGE_ANCHORING_ADDR",
    "l2MessageAnchoringPubkey": "$L2_MESSAGE_ANCHORING_PUBKEY",
    "l2PostmanAddress": "$L2_POSTMAN_ADDR",
    "l2PostmanPubkey": "$L2_POSTMAN_PUBKEY"
  },
  "l1": {
    "LineaRollupV8": "$L1_LINEA_ROLLUP"
  },
  "l2": {
    "L2MessageService": "$L2_MESSAGE_SERVICE"
  }
}
JSON_EOF

log "Wrote $OUT_JSON"

write_keystore() {
  _ks_file="$OUT_KEYS_DIR/$1.yaml"
  umask 077
  cat > "$_ks_file" <<KEY_EOF
# ============================================================
# DEV ONLY — generated at boot by account-setup.sh.
# Stored in the Docker shared volume for retry-safe restarts.
# Do NOT commit these generated files.
# ============================================================
type: "file-raw"
keyType: "SECP256K1"
# $2
privateKey: "$3"
KEY_EOF
  chmod 0644 "$_ks_file" || die "failed to chmod $_ks_file"
  log "Wrote $_ks_file"
}

write_keystore "anchoring-signer"       "L2 message-anchoring signer (generated runtime key, L2 only)" "$L2_MESSAGE_ANCHORING_PRIVATE_KEY"
write_keystore "data-submission-signer" "L1 blob/data-submission signer (generated runtime key)" "$L1_BLOB_SUBMITTER_PRIVATE_KEY"
write_keystore "finalization-signer"    "L1 aggregation/finalization signer (generated runtime key)" "$L1_FINALIZATION_SUBMITTER_PRIVATE_KEY"

log "Done."
