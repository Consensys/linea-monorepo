#!/usr/bin/env sh
set -eu

ROOT=$(git rev-parse --show-toplevel)
STACK_REL=docs/getting-started/linea-stack
STACK="$ROOT/$STACK_REL"
FAILURES=0

fail() {
  printf '[quickstart-static] FAIL: %s\n' "$*" >&2
  FAILURES=$((FAILURES + 1))
}

pass() {
  printf '[quickstart-static] OK: %s\n' "$*"
}

check_no_tracked_generated_genesis() {
  tracked=$(git -C "$ROOT" ls-files "$STACK_REL/config/l2/genesis-init" \
    | grep -E '/(genesis-(besu|maru)\.json|fork-timestamp\.txt)$' || true)
  if [ -n "$tracked" ]; then
    fail "generated L2 genesis artifacts are tracked: $(printf '%s' "$tracked" | tr '\n' ' ')"
  else
    pass "generated L2 genesis artifacts are not tracked"
  fi

  for template in genesis-besu.json.template genesis-maru.json.template; do
    if git -C "$ROOT" ls-files --error-unmatch "$STACK_REL/config/l2/genesis-init/$template" >/dev/null 2>&1; then
      pass "$template is tracked"
    else
      fail "$template must remain tracked"
    fi
  done
}

check_generated_l2_deployer_genesis() {
  genesis_template="$STACK/config/l2/genesis-init/genesis-besu.json.template"
  genesis_init="$STACK/config/l2/genesis-init/init.sh"

  if grep -q '__L2_DEPLOYER_ADDRESS__' "$genesis_template" && grep -q '__L2_DEPLOYER_ADDRESS__' "$genesis_init"; then
    pass "generated L2 deployer is injected into L2 genesis"
  else
    fail "generated L2 deployer must be injected into L2 genesis"
  fi

  if node - "$genesis_template" <<'NODE'
const fs = require("fs");
const genesis = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const funded = Object.entries(genesis.alloc)
  .filter(([, entry]) => BigInt(entry.balance || "0") > 0n)
  .map(([addr]) => addr)
  .sort();
const expected = ["__L2_DEPLOYER_ADDRESS__", "__L2_MESSAGE_SERVICE_ADDRESS__"].sort();
if (JSON.stringify(funded) !== JSON.stringify(expected)) {
  console.error(`funded=${funded.join(",")}`);
  process.exit(1);
}
for (const [addr, entry] of Object.entries(genesis.alloc)) {
  if (Object.prototype.hasOwnProperty.call(entry, "privateKey")) {
    console.error(`privateKey still present in genesis alloc for ${addr}`);
    process.exit(1);
  }
}
NODE
  then
    pass "genesis only funds generated L2 deployer and L2MessageService"
  else
    fail "genesis must only fund generated L2 deployer and L2MessageService, with no privateKey entries"
  fi
}

check_account_setup_key_model() {
  account_setup="$STACK/scripts/account-setup.sh"

  for field in \
    l1BlobSubmitterAddress \
    l1FinalizationSubmitterAddress \
    l1PostmanAddress \
    l2DeployerAddress \
    l2MessageAnchoringAddress \
    l2PostmanAddress; do
    if grep -q "\"$field\"" "$account_setup"; then
      pass "addresses-precomputed.json includes signers.$field"
    else
      fail "addresses-precomputed.json must include signers.$field"
    fi
  done

  for contract in LineaRollupV8 L2MessageService; do
    if grep -q "\"$contract\"" "$account_setup"; then
      pass "addresses-precomputed.json includes boot-critical $contract"
    else
      fail "addresses-precomputed.json must include boot-critical $contract"
    fi
  done

  for contract in Verifier ForcedTransactionGateway BridgedToken TokenBridge TestERC20; do
    if grep -q "\"$contract\"" "$account_setup"; then
      fail "addresses-precomputed.json should not precompute non-boot contract $contract"
    else
      pass "addresses-precomputed.json does not precompute non-boot contract $contract"
    fi
  done

  if grep -q 'L2_DEPLOYER_PRIVATE_KEY:=0x' "$account_setup"; then
    fail "account-setup still defaults to a pre-baked L2 deployer private key"
  else
    pass "account-setup does not default to a pre-baked L2 deployer private key"
  fi

  if grep -q 'L2_MESSAGE_ANCHORING_PRIVATE_KEY:=0x' "$account_setup"; then
    fail "account-setup still defaults to a pre-baked L2 anchorer private key"
  else
    pass "account-setup does not default to a pre-baked L2 anchorer private key"
  fi

  if grep -q 'L2_LIVENESS_SIGNER_PRIVATE_KEY\|write_keystore "liveness-signer"' "$account_setup"; then
    fail "account-setup still generates or persists a liveness signer"
  else
    pass "account-setup does not generate a liveness signer"
  fi

  if grep -q '/shared/runtime-keys.env\|OUT_RUNTIME_KEYS_ENV' "$account_setup"; then
    pass "account-setup persists generated private keys for retry-safe consumers"
  else
    fail "account-setup must persist generated private keys in /shared/runtime-keys.env"
  fi

  if grep -q 'chmod 0644 "$OUT_RUNTIME_KEYS_ENV"' "$account_setup"; then
    pass "runtime key env is readable by non-root service containers"
  else
    fail "account-setup must chmod runtime-keys.env so postman can read generated keys"
  fi

  if grep -q 'chmod 0644 "$_ks_file"' "$account_setup" && grep -q 'chmod 0644 "$OUT_KEYS_DIR/anchoring-signer.yaml"' "$account_setup"; then
    pass "web3signer key files are readable by the web3signer container"
  else
    fail "account-setup must chmod web3signer key files so web3signer can load generated signers"
  fi
}

check_postman_key_model() {
  postman_env="$STACK/config/l2/postman/env"
  compose="$STACK/docker-compose.yml"
  sequencer_template="$STACK/config/l2/sequencer/sequencer.config.toml.template"

  if grep -q '^L2_SIGNER_PRIVATE_KEY=0x' "$postman_env"; then
    fail "postman env still contains a pre-baked L2 signer private key"
  else
    pass "postman env does not contain a pre-baked L2 signer private key"
  fi

  if grep -q 'L1_POSTMAN_PRIVATE_KEY' "$compose"; then
    pass "docker compose wires generated L1 postman key"
  else
    fail "docker compose must wire L1_POSTMAN_PRIVATE_KEY into postman"
  fi

  if grep -q 'L2_POSTMAN_PRIVATE_KEY' "$compose"; then
    pass "docker compose wires generated L2 postman key"
  else
    fail "docker compose must wire L2_POSTMAN_PRIVATE_KEY into postman"
  fi

  if grep -q '^plugin-linea-liveness-enabled=false' "$sequencer_template"; then
    pass "sequencer liveness is disabled for quickstart"
  else
    fail "sequencer liveness must be disabled for quickstart"
  fi

  extra_liveness_config="$(grep '^plugin-linea-liveness-' "$sequencer_template" | grep -v '^plugin-linea-liveness-enabled=false$' || true)"
  if [ -n "$extra_liveness_config" ]; then
    fail "sequencer template must not carry inactive liveness signer/contract settings"
  else
    pass "sequencer liveness config contains only enabled=false"
  fi

  if grep -q 'net.consensys.linea.sequencer.liveness' "$STACK/config/l2/sequencer/log4j.xml"; then
    fail "sequencer log4j should not enable liveness-specific logging when liveness is disabled"
  else
    pass "sequencer log4j has no liveness-specific logger"
  fi
}

check_incremental_typescript_helpers() {
  deploy_contracts="$STACK/scripts/deploy-contracts.sh"
  aggregate_ts="$STACK/scripts/aggregate-addresses.ts"

  if [ -f "$aggregate_ts" ] && grep -q 'ts-node /scripts/aggregate-addresses.ts' "$deploy_contracts"; then
    pass "addresses.json aggregation is implemented as an incremental TypeScript helper"
  else
    fail "deploy-contracts should call scripts/aggregate-addresses.ts for addresses.json aggregation"
  fi
}

check_smoke_and_traffic_scripts() {
  for script in send-l2-test-tx.sh send-l2-erc20-transfer.sh smoke-bridge-message.sh; do
    script_path="$STACK/scripts/$script"
    if [ -x "$script_path" ]; then
      pass "$script is executable"
    else
      fail "$script must exist and be executable"
    fi

    if [ -f "$script_path" ] && sh -n "$script_path"; then
      pass "$script has valid shell syntax"
    elif [ -f "$script_path" ]; then
      fail "$script has invalid shell syntax"
    fi
  done

  if [ -f "$STACK/scripts/send-l2-test-tx.sh" ] && grep -q 'L2_DEPLOYER_PRIVATE_KEY' "$STACK/scripts/send-l2-test-tx.sh" && grep -q 'cast send' "$STACK/scripts/send-l2-test-tx.sh"; then
    pass "send-l2-test-tx.sh sends a simple L2 transaction from the generated L2 deployer"
  else
    fail "send-l2-test-tx.sh must send a simple L2 transaction from the generated L2 deployer"
  fi

  if [ -f "$STACK/scripts/send-l2-erc20-transfer.sh" ] && grep -q 'ERC20Example' "$STACK/scripts/send-l2-erc20-transfer.sh" && grep -q 'transfer(address,uint256)' "$STACK/scripts/send-l2-erc20-transfer.sh"; then
    pass "send-l2-erc20-transfer.sh transfers ERC20Example on L2"
  else
    fail "send-l2-erc20-transfer.sh must transfer ERC20Example on L2"
  fi

  if [ -f "$STACK/scripts/smoke-bridge-message.sh" ] && grep -q 'BRIDGE_SMOKE_SEND' "$STACK/scripts/smoke-bridge-message.sh" && grep -q 'not a pass/fail bridge smoke test yet' "$STACK/scripts/smoke-bridge-message.sh"; then
    pass "smoke-bridge-message.sh is guarded against fake bridge success"
  else
    fail "smoke-bridge-message.sh must be guarded against fake bridge success"
  fi
}

check_no_tracked_generated_genesis
check_generated_l2_deployer_genesis
check_account_setup_key_model
check_postman_key_model
check_incremental_typescript_helpers
check_smoke_and_traffic_scripts

if [ "$FAILURES" -ne 0 ]; then
  printf '[quickstart-static] %s failure(s)\n' "$FAILURES" >&2
  exit 1
fi

printf "[quickstart-static] all checks passed\n"
