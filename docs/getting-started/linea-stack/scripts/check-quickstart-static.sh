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

check_generated_genesis_is_volume_scoped() {
  compose="$STACK/docker-compose.yml"
  deploy_contracts="$STACK/scripts/deploy-contracts.sh"

  if grep -q 'linea-l2-genesis:' "$compose" \
    && grep -q 'name: "linea-stack-l2-genesis"' "$compose"; then
    pass "generated L2 genesis has a named Docker volume"
  else
    fail "generated L2 genesis must live in a Docker volume, not the repo tree"
  fi

  if grep -q './config/l2/genesis-init:/templates:ro' "$compose" \
    && grep -q 'linea-l2-genesis:/initialization:rw' "$compose"; then
    pass "l2-genesis-init reads templates from repo and writes generated genesis to volume"
  else
    fail "l2-genesis-init must separate read-only templates from generated genesis output"
  fi

  if grep -q './config/l2/genesis-init:/initialization' "$compose" \
    || grep -q './config/l2/genesis-init/genesis-besu.json' "$compose"; then
    fail "services must not bind-mount generated genesis files from the repo tree"
  else
    pass "services do not bind-mount generated genesis files from the repo tree"
  fi

  if grep -q 'linea-l2-genesis:/generated-genesis:ro' "$compose" \
    && grep -q '/generated-genesis/fork-timestamp.txt' "$deploy_contracts"; then
    pass "deploy-contracts reads fork timestamp from generated genesis volume"
  else
    fail "deploy-contracts must read fork timestamp from the generated genesis volume"
  fi
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
const template = fs.readFileSync(process.argv[2], "utf8")
  .replace(/__L2_CHAIN_ID__/g, "1337");
const genesis = JSON.parse(template);
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

check_l2_chain_id_wiring() {
  besu_genesis="$STACK/config/l2/genesis-init/genesis-besu.json.template"
  maru_genesis="$STACK/config/l2/genesis-init/genesis-maru.json.template"
  genesis_init="$STACK/config/l2/genesis-init/init.sh"
  prover_template="$STACK/config/l2/prover/prover-config-partial.toml.template"
  compose="$STACK/docker-compose.yml"
  status_script="$STACK/scripts/status.sh"

  if grep -q '"chainId": __L2_CHAIN_ID__' "$besu_genesis" \
    && grep -q '"chainId": __L2_CHAIN_ID__' "$maru_genesis" \
    && grep -qF 's/__L2_CHAIN_ID__/$L2_CHAIN_ID/g' "$genesis_init"; then
    pass "L2_CHAIN_ID is templated into Besu and Maru genesis"
  else
    fail "L2_CHAIN_ID must be templated into both L2 genesis files"
  fi

  if grep -q '^chain_id = __L2_CHAIN_ID__' "$prover_template" \
    && grep -qF 's|__L2_CHAIN_ID__|$$L2_CHAIN_ID|g' "$compose"; then
    pass "L2_CHAIN_ID is templated into prover public input config"
  else
    fail "L2_CHAIN_ID must feed prover-config-partial.toml.template"
  fi

  if grep -qF 'L2_CHAIN_ID: ${L2_CHAIN_ID:-1337}' "$compose" \
    && grep -qF 'NEXT_PUBLIC_NETWORK_ID: ${L2_CHAIN_ID:-1337}' "$compose" \
    && grep -qF 'CHAIN_ID: ${L2_CHAIN_ID:-1337}' "$compose"; then
    pass "docker compose propagates L2_CHAIN_ID to init, deploy, and explorer"
  else
    fail "docker compose must propagate L2_CHAIN_ID to init/deploy/explorer"
  fi

  if grep -q 'eth_chainId' "$status_script" \
    && grep -qF '^chain_id|^prover_mode|^is_allowed_circuit_id' "$status_script"; then
    pass "status.sh reports L2 chain ID and rendered prover mode"
  else
    fail "status.sh must report L2 chain ID and rendered prover mode"
  fi

  if grep -q 'l1 data availability vs finalization' "$status_script" \
    && grep -q 'latest blob tx (DA only)' "$status_script" \
    && grep -q 'finalizeBlocks(bytes,uint256,tuple)' "$status_script" \
    && grep -q 'currentL2BlockNumber' "$status_script" \
    && grep -q 'DataFinalizedV3' "$status_script" \
    && grep -q 'FinalizedStateUpdated' "$status_script"; then
    pass "status.sh separates blob submission from rollup finalization"
  else
    fail "status.sh must distinguish blob submission from finalizeBlocks finalization"
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

check_partial_prover_guardrails() {
  compose="$STACK/docker-compose.yml"
  prover_template="$STACK/config/l2/prover/prover-config-partial.toml.template"
  coordinator_template="$STACK/config/l2/coordinator/coordinator-config.toml.template"
  env_example="$STACK/.env.example"
  readme="$STACK/README.md"

  if grep -q 'PROVER_GOMEMLIMIT must be set explicitly when PROVER_DEV_OVERRIDE=false' "$compose"; then
    pass "partial prover mode requires explicit PROVER_GOMEMLIMIT"
  else
    fail "PROVER_DEV_OVERRIDE=false must require explicit PROVER_GOMEMLIMIT"
  fi

  if grep -q '^prover_mode = "partial"' "$prover_template" \
    && grep -q '^is_allowed_circuit_id = 483' "$prover_template" \
    && grep -q 'partial render verified: 2× partial, 2× dev, is_allowed_circuit_id=483' "$compose"; then
    pass "partial prover render is verified before prover starts"
  else
    fail "config-render must verify partial prover modes and bitmask"
  fi

  if grep -q '__L1_BLOB_MAX_FEE_PER_BLOB_GAS_CAP_WEI__' "$coordinator_template" \
    && grep -q '__L1_FINALIZATION_MAX_FEE_PER_GAS_CAP_WEI__' "$coordinator_template" \
    && grep -q 'L1_BLOB_MAX_FEE_PER_BLOB_GAS_CAP_WEI' "$compose" \
    && grep -q 'L1_FINALIZATION_MAX_FEE_PER_GAS_CAP_WEI' "$compose" \
    && grep -q 'L1_BLOB_MAX_FEE_PER_BLOB_GAS_CAP_WEI' "$env_example" \
    && grep -q 'L1_FINALIZATION_MAX_FEE_PER_GAS_CAP_WEI' "$env_example" \
    && grep -q 'L1 gas caps and Sepolia congestion' "$readme"; then
    pass "L1 blob/finalization gas caps are documented and configurable"
  else
    fail "L1 blob/finalization gas caps must be documented and configurable"
  fi
}

check_reuse_guardrails() {
  deploy_contracts="$STACK/scripts/deploy-contracts.sh"
  status_script="$STACK/scripts/status.sh"

  if grep -q 'step_already_done_with_code' "$deploy_contracts" \
    && grep -q 'cast code' "$deploy_contracts" \
    && grep -q 'present but no code' "$deploy_contracts" \
    && grep -q 'step_already_done_with_code "$logfile" "L2MessageService"' "$deploy_contracts" \
    && grep -q 'step_already_done_with_code "$logfile" "TokenBridge"' "$deploy_contracts" \
    && grep -q 'step_already_done_with_code "$logfile" "TestERC20"' "$deploy_contracts"; then
    pass "deploy-contracts verifies L2 code before trusting prior deploy logs"
  else
    fail "deploy-contracts must verify L2 on-chain code before skipping from prior deploy logs"
  fi

  if grep -q 'l2 rpc latest block' "$status_script" \
    && grep -q 'rollup finalized block is ahead of local L2 latest block' "$status_script" \
    && grep -q 'local chain state does not match the preserved L1 rollup state' "$status_script"; then
    pass "status.sh warns when preserved L1 rollup state is ahead of local L2"
  else
    fail "status.sh must warn when preserved L1 rollup state is ahead of local L2"
  fi
}

check_smoke_and_traffic_scripts() {
  for script in check-ports.sh send-l2-test-tx.sh send-l2-erc20-transfer.sh generate-l2-erc20-traffic.sh smoke-bridge-message.sh smoke-bridge-erc20-l1-to-l2.sh smoke-bridge-message-l2-to-l1.sh smoke-bridge-erc20-l2-to-l1.sh; do
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

  if [ -f "$STACK/scripts/check-ports.sh" ] \
    && grep -q 'HOST_PORT_L2_RPC' "$STACK/scripts/check-ports.sh" \
    && grep -q 'HOST_PORT_L2_BLOCKSCOUT_FRONTEND' "$STACK/scripts/check-ports.sh" \
    && grep -q 'HOST_PORT_COORDINATOR' "$STACK/scripts/check-ports.sh" \
    && grep -q './scripts/check-ports.sh' "$STACK/README.md"; then
    pass "check-ports.sh preflights expected host ports"
  else
    fail "check-ports.sh must preflight expected host ports and be documented"
  fi

  if [ -f "$STACK/scripts/send-l2-test-tx.sh" ] && grep -q 'L2_DEPLOYER_PRIVATE_KEY' "$STACK/scripts/send-l2-test-tx.sh" && grep -q 'cast send' "$STACK/scripts/send-l2-test-tx.sh"; then
    pass "send-l2-test-tx.sh sends a simple L2 transaction from the generated L2 deployer"
  else
    fail "send-l2-test-tx.sh must send a simple L2 transaction from the generated L2 deployer"
  fi

  if [ -f "$STACK/scripts/send-l2-erc20-transfer.sh" ] \
    && grep -q 'ERC20Example' "$STACK/scripts/send-l2-erc20-transfer.sh" \
    && grep -q 'transfer(address,uint256)' "$STACK/scripts/send-l2-erc20-transfer.sh" \
    && grep -q 'DEMO_TRAFFIC_ENV="/shared/demo-traffic.env"' "$STACK/scripts/send-l2-erc20-transfer.sh" \
    && grep -q 'L2_TRAFFIC_PRIVATE_KEY' "$STACK/scripts/send-l2-erc20-transfer.sh" \
    && grep -q 'L2_TRAFFIC_ETH_TOP_UP_WEI' "$STACK/scripts/send-l2-erc20-transfer.sh" \
    && grep -q 'L2_TRAFFIC_ERC20_TOP_UP_WEI' "$STACK/scripts/send-l2-erc20-transfer.sh"; then
    pass "send-l2-erc20-transfer.sh uses a funded disposable L2 traffic account"
  else
    fail "send-l2-erc20-transfer.sh must fund and transfer from a disposable L2 traffic account"
  fi

  if [ -f "$STACK/scripts/generate-l2-erc20-traffic.sh" ] \
    && grep -q 'docker run -d' "$STACK/scripts/generate-l2-erc20-traffic.sh" \
    && grep -q 'while \[ "$MAX_TXS" -eq 0 \]' "$STACK/scripts/generate-l2-erc20-traffic.sh" \
    && grep -q 'transfer(address,uint256)' "$STACK/scripts/generate-l2-erc20-traffic.sh" \
    && grep -q 'DEMO_TRAFFIC_ENV="/shared/demo-traffic.env"' "$STACK/scripts/generate-l2-erc20-traffic.sh" \
    && grep -q 'L2_TRAFFIC_PRIVATE_KEY' "$STACK/scripts/generate-l2-erc20-traffic.sh" \
    && grep -q 'L2_TRAFFIC_ETH_TOP_UP_WEI' "$STACK/scripts/generate-l2-erc20-traffic.sh" \
    && grep -q 'L2_TRAFFIC_ERC20_TOP_UP_WEI' "$STACK/scripts/generate-l2-erc20-traffic.sh"; then
    pass "generate-l2-erc20-traffic.sh runs continuous traffic from a disposable L2 traffic account"
  else
    fail "generate-l2-erc20-traffic.sh must run continuous traffic from a disposable L2 traffic account"
  fi

  if [ -f "$STACK/scripts/smoke-bridge-message.sh" ] \
    && grep -q 'CLAIMED_SUCCESS' "$STACK/scripts/smoke-bridge-message.sh" \
    && grep -q 'claim_tx_hash' "$STACK/scripts/smoke-bridge-message.sh" \
    && grep -q 'MessageClaimed' "$STACK/scripts/smoke-bridge-message.sh" \
    && ! grep -q 'not a pass/fail bridge smoke test yet' "$STACK/scripts/smoke-bridge-message.sh"; then
    pass "smoke-bridge-message.sh verifies a real L1-to-L2 claim"
  else
    fail "smoke-bridge-message.sh must send and verify a real L1-to-L2 claim"
  fi

  if [ -f "$STACK/scripts/smoke-bridge-erc20-l1-to-l2.sh" ] \
    && grep -q 'bridgeToken(address,uint256,address)' "$STACK/scripts/smoke-bridge-erc20-l1-to-l2.sh" \
    && grep -q 'approve(address,uint256)' "$STACK/scripts/smoke-bridge-erc20-l1-to-l2.sh" \
    && grep -q 'nativeToBridgedToken(uint256,address)' "$STACK/scripts/smoke-bridge-erc20-l1-to-l2.sh" \
    && grep -q 'balanceOf(address)(uint256)' "$STACK/scripts/smoke-bridge-erc20-l1-to-l2.sh" \
    && grep -q 'CLAIMED_SUCCESS' "$STACK/scripts/smoke-bridge-erc20-l1-to-l2.sh" \
    && grep -q 'smoke-bridge-erc20-l1-to-l2.sh' "$STACK/README.md"; then
    pass "smoke-bridge-erc20-l1-to-l2.sh verifies a real ERC20 TokenBridge L1-to-L2 transfer"
  else
    fail "smoke-bridge-erc20-l1-to-l2.sh must bridge ERC20 through TokenBridge and verify the L2 balance"
  fi

  if [ -f "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" ] \
    && grep -q 'bridgeToken(address,uint256,address)' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'approve(address,uint256)' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'L2_SEND_RPC_URL' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'claimOnL1' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'getMessageProof' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'balanceOf(address)(uint256)' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'BridgingFinalizedV2' "$STACK/scripts/smoke-bridge-erc20-l2-to-l1.sh" \
    && grep -q 'smoke-bridge-erc20-l2-to-l1.sh' "$STACK/README.md"; then
    pass "smoke-bridge-erc20-l2-to-l1.sh verifies a real ERC20 TokenBridge L2-to-L1 withdrawal"
  else
    fail "smoke-bridge-erc20-l2-to-l1.sh must bridge ERC20 through TokenBridge, claim on L1, and verify the L1 balance"
  fi

  if [ -f "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" ] \
    && grep -q 'L2_TO_L1' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'L2MessageService' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'sendMessage(address,uint256,bytes)' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'L2_SEND_RPC_URL' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'claimOnL1' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'getMessageProof' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'CLAIMED_SUCCESS' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'MessageClaimed' "$STACK/scripts/smoke-bridge-message-l2-to-l1.sh" \
    && grep -q 'smoke-bridge-message-l2-to-l1.sh' "$STACK/README.md"; then
    pass "smoke-bridge-message-l2-to-l1.sh verifies a real L2-to-L1 claim"
  else
    fail "smoke-bridge-message-l2-to-l1.sh must send, finalize, claim, and verify a real L2-to-L1 message"
  fi
}

check_pinned_utility_images_and_docs() {
  versions="$STACK/versions.env"
  readme="$STACK/README.md"

  if grep -q '^FOUNDRY_TAG=' "$versions" && ! grep -q '^FOUNDRY_TAG=latest$' "$versions"; then
    pass "Foundry image tag is pinned"
  else
    fail "FOUNDRY_TAG must be pinned, not latest"
  fi

  if grep -q 'Accounts and funding model' "$readme" \
    && grep -q 'generated L2 deployer' "$readme" \
    && grep -q 'L2MessageService' "$readme" \
    && grep -q 'ERC20Example' "$readme" \
    && grep -q 'disposable demo traffic account' "$readme" \
    && grep -q 'bootstrap/admin' "$readme"; then
    pass "README documents L2 ETH and ERC20Example funding model"
  else
    fail "README must document L2 ETH, ERC20Example, and disposable traffic-account funding model"
  fi
}

check_no_tracked_generated_genesis
check_generated_genesis_is_volume_scoped
check_generated_l2_deployer_genesis
check_l2_chain_id_wiring
check_account_setup_key_model
check_postman_key_model
check_incremental_typescript_helpers
check_partial_prover_guardrails
check_reuse_guardrails
check_smoke_and_traffic_scripts
check_pinned_utility_images_and_docs

if [ "$FAILURES" -ne 0 ]; then
  printf '[quickstart-static] %s failure(s)\n' "$FAILURES" >&2
  exit 1
fi

printf "[quickstart-static] all checks passed\n"
