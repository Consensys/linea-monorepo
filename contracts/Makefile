.PHONY: test coverage

compile:
	npm run build
	npx hardhat run scripts/hardhat/postCompile.ts

clean:
	rm -rf data/
	rm -rf coverage/
	rm -rf build/
	rm -rf cache/
	mkdir -p data
	rm -f contracts/verifiers/*.sol

lint:
	npm run solhint:check

test: export SKIP_DEPLOY_LOG := true
test:
	npm test

fmt:
	npm run lint:check

fmt-fix:
	npm run lint:fix

coverage: export SKIP_DEPLOY_LOG := true
coverage:
	npm run coverage

deploy:
	npx hardhat run --network besu scripts/hardhat/deploy.ts

deploy-local: export BLOCKCHAIN_NODE := http://127.0.0.1:8545
deploy-local: export ROLLUP_JSON_PATH := $(shell pwd)/data/rollup.json
deploy-local: export SMC_CONFIG_PATH := $(shell pwd)/data/smc_config.json
deploy-local: export ERC20_ADDR_PATH := $(shell pwd)/data/erc20_addr.addr
deploy-local: deploy

deploy-zkevm:
	npx hardhat run --network zkevm_dev scripts/hardhat/deploy.ts

deploy-local-zkevm: export BLOCKCHAIN_NODE := http://127.0.0.1:8445
deploy-local-zkevm: export ROLLUP_JSON_PATH := $(shell pwd)/data/rollup.json
deploy-local-zkevm: export SMC_CONFIG_PATH := $(shell pwd)/data/smc_config.json
deploy-local-zkevm: export TX_GAS_LIMIT := 10000000
deploy-local-zkevm: export BLOCKCHAIN_GAS_PRICE := 1400000000
deploy-local-zkevm: export BLOCKCHAIN_TIMEOUT_MS := 30000
deploy-local-zkevm: deploy-zkevm

register-operator-in-smc-local: export BLOCKCHAIN_NODE := http://127.0.0.1:8545
register-operator-in-smc-local: export SMC_CONFIG_PATH := $(shell pwd)/data/smc_config.json
register-operator-in-smc-local:
	ts-node scripts/registerOperator.ts data/rollup.json ../node-data/test/keys/contract_owner.acc ../node-data/test/keys/operator_1.acc

register-operator-2-in-smc-local: export BLOCKCHAIN_NODE := http://127.0.0.1:8545
register-operator-2-in-smc-local: export SMC_CONFIG_PATH := $(shell pwd)/data/smc_config.json
register-operator-2-in-smc-local:
	ts-node scripts/registerOperator.ts data/rollup.json ../node-data/test/keys/contract_owner.acc ../node-data/test/keys/operator_2.acc

register-erc20-local: export BLOCKCHAIN_NODE := http://127.0.0.1:8545
register-erc20-local: export ROLLUP_JSON_PATH := $(shell pwd)/data/rollup.json
register-erc20-local: export ERC_20_MAPPING_PATH := $(shell pwd)/data/erc_20_mapping.json
register-erc20-local: export ERC20_ADDR_PATH := $(shell pwd)/data/erc20_addr.addr
register-erc20-local: export ERC_20_NUMBER := 2
register-erc20-local:
	npx hardhat run --network besu scripts/hardhat/deployAndRegisterErc20.ts

load-funds-to-smc: export BLOCKCHAIN_NODE := http://127.0.0.1:8545
load-funds-to-smc:
	ts-node scripts/sendEthToRollup.ts data/rollup.json ../node-data/test/keys/contract_owner.acc "90.10"

print-all-test-balances:
	npm run balance -- \
		../node-data/test/keys/contract_owner.acc \
		../node-data/test/keys/operator_1.acc \
		../node-data/test/keys/operator_2.acc \
		../node-data/test/keys/eth_account_3.acc \
		../node-data/test/keys/eth_account_4.acc \
		../node-data/test/keys/eth_account_5.acc \
		../node-data/test/keys/eth_account_6.acc \
		../node-data/test/keys/eth_account_7.acc \
		../node-data/test/keys/eth_account_8.acc \
		../node-data/test/keys/eth_account_9.acc \
		../node-data/test/keys/eth_account_10.acc \
		data/rollup.json

enable-zk-proof-verification:
	ts-node scripts/changeZkVerification.ts data/rollup.json ../node-data/test/keys/contract_owner.acc 1

disable-zk-proof-verification:
	ts-node scripts/changeZkVerification.ts data/rollup.json ../node-data/test/keys/contract_owner.acc 0

enable-signature-verification:
	ts-node scripts/changeSignatureVerificaiton.ts data/rollup.json ../node-data/test/keys/contract_owner.acc 1

print-zk-verification-status:
	ts-node scripts/printZkVerification.ts data/rollup.json ../node-data/test/keys/contract_owner.acc

listen-to-events: export DISABLE_BLOCKCHAIN_LOG := true
listen-to-events:
	ts-node scripts/printEvents.ts data/rollup.json ../node-data/test/keys/contract_owner.acc

print-nonces: export DISABLE_BLOCKCHAIN_LOG := true
print-nonces:
	ts-node scripts/printNonces.ts ../node-data/test/keys/operator_1.acc

run-ganache:
	ganache-cli --allowUnlimitedContractSize --callGasLimit="0x1fffffffffffff" -l="0x1fffffffffffff" -g="0x0" --verbose --account=0xb17202c37cce9498e6f7dcdc1abd207802d09b5eee96677ea219ac867a198b91,90021000000000000000000000  --account=0x202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d,9000000000000000000000000

.PHONY: deploy-testnet-token-bridge
deploy-testnet-token-bridge:
	npx hardhat run --network zkevm_dev scripts/tokenBridge/deploy-1.ts
	npx hardhat run --network l2 scripts/tokenBridge/deploy-1.ts
	npx hardhat run --network zkevm_dev scripts/tokenBridge/deploy-2.ts
	npx hardhat run --network l2 scripts/tokenBridge/deploy-2.ts