define get_future_time
$(shell \
    OS=$$(uname); \
    if [ "$$OS" = "Linux" ]; then \
        date -d '+3 seconds' +%s; \
    elif [ "$$OS" = "Darwin" ]; then \
        date -v +3S +%s; \
    fi \
)
endef

pnpm-install:
		pnpm install

docker-pull-develop:
		L1_GENESIS_TIME=$(get_future_time) docker compose -f docker/compose.yml pull

clean-smc-folders:
		rm -f contracts/.openzeppelin/unknown-31648428.json
		rm -f contracts/.openzeppelin/unknown-1337.json

clean-local-folders:
		make clean-smc-folders
		rm -rf tmp/local/*

clean-testnet-folders:
		make clean-smc-folders
		rm -rf tmp/testnet/*

clean-environment:
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile l1 --profile l2 --profile debug down || true
		make clean-local-folders
		docker network prune -f
		docker volume rm linea-local-dev linea-logs || true # ignore failure if volumes do not exist already
		# Commented out because it's quite time consuming to download the plugin, but it's useful to remember about it
		#rm -rf tmp/linea-besu-sequencer/plugins/

start-l1:
		L1_GENESIS_TIME=$(get_future_time) docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l1 up -d

start-l2:
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l2 up -d

start-l2-blockchain-only:
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l2-bc up -d

start-whole-environment:
		# docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml build prover
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l1 --profile l2 up -d

start-whole-environment-traces-v2:
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile l1 --profile l2 up -d

pull-all-images:
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile l1 --profile l2 pull

pull-images-external-to-monorepo:
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile external-to-monorepo pull

compile-contracts:
		cd contracts; \
		make compile

compile-contracts-no-cache:
		cd contracts/; \
		make force-compile

deploy-linea-rollup:
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		cd contracts/; \
		PRIVATE_KEY=$${DEPLOYMENT_PRIVATE_KEY:-0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80} \
		BLOCKCHAIN_NODE=http:\\localhost:8445/ \
		PLONKVERIFIER_NAME=IntegrationTestTrueVerifier \
		LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH=0x072ead6777750dc20232d1cee8dc9a395c2d350df4bbaa5096c6f59b214dcecd \
		LINEA_ROLLUP_INITIAL_L2_BLOCK_NUMBER=0 \
		LINEA_ROLLUP_SECURITY_COUNCIL=0x90F79bf6EB2c4f870365E785982E1f101E93b906 \
		LINEA_ROLLUP_OPERATORS=$${LINEA_ROLLUP_OPERATORS:-0x70997970C51812dc3A010C7d01b50e0d17dc79C8,0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC} \
		LINEA_ROLLUP_RATE_LIMIT_PERIOD=86400 \
		LINEA_ROLLUP_RATE_LIMIT_AMOUNT=1000000000000000000000 \
		LINEA_ROLLUP_GENESIS_TIMESTAMP=1683325137 \
		npx hardhat deploy --no-compile --network zkevm_dev --tags PlonkVerifier,LineaRollupV5

deploy-linea-rollup-v6:
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		cd contracts/; \
		PRIVATE_KEY=$${DEPLOYMENT_PRIVATE_KEY:-0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80} \
		BLOCKCHAIN_NODE=http:\\localhost:8445/ \
		PLONKVERIFIER_NAME=IntegrationTestTrueVerifier \
		LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH=0x072ead6777750dc20232d1cee8dc9a395c2d350df4bbaa5096c6f59b214dcecd \
		LINEA_ROLLUP_INITIAL_L2_BLOCK_NUMBER=0 \
		LINEA_ROLLUP_SECURITY_COUNCIL=0x90F79bf6EB2c4f870365E785982E1f101E93b906 \
		LINEA_ROLLUP_OPERATORS=$${LINEA_ROLLUP_OPERATORS:-0x70997970C51812dc3A010C7d01b50e0d17dc79C8,0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC} \
		LINEA_ROLLUP_RATE_LIMIT_PERIOD=86400 \
		LINEA_ROLLUP_RATE_LIMIT_AMOUNT=1000000000000000000000 \
		LINEA_ROLLUP_GENESIS_TIMESTAMP=1683325137 \
		npx hardhat deploy --no-compile --network zkevm_dev --tags PlonkVerifier,LineaRollup

deploy-l2messageservice:
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		cd contracts/; \
		PRIVATE_KEY=$${DEPLOYMENT_PRIVATE_KEY:-0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae} \
		BLOCKCHAIN_NODE=http:\\localhost:8545/ \
		L2MSGSERVICE_SECURITY_COUNCIL=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266 \
		L2MSGSERVICE_L1L2_MESSAGE_SETTER=$${L2MSGSERVICE_L1L2_MESSAGE_SETTER:-0xd42e308fc964b71e18126df469c21b0d7bcb86cc} \
		L2MSGSERVICE_RATE_LIMIT_PERIOD=86400 \
		L2MSGSERVICE_RATE_LIMIT_AMOUNT=1000000000000000000000 \
		npx hardhat deploy --no-compile  --network zkevm_dev --tags L2MessageService

upgrade-linea-rollup-on-uat:
		cd contracts/; \
		rm -f .openzeppelin/goerli.json; \
		sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=https:\/\/goerli.infura.io\/v3\/${INFURA_KEY}/" .env.template.uat > .env; \
		sed -i~ "s/PRIVATE_KEY=.*/PRIVATE_KEY=${PRIVATE_KEY}/" .env; \
		npx hardhat run ./scripts/upgrades/upgradeZkEVM.ts --network zkevm_dev

fresh-start-l2-blockchain-only:
		make clean-environment
		make start-l2-blockchain-only

restart-shomei:
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml rm zkbesu-shomei shomei
		rm -rf tmp/local/shomei/*
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml up zkbesu-shomei shomei -d

fresh-start-all-smc-v4:
		make clean-environment
		make start-all-smc-v4

fresh-start-all:
		make clean-environment
		make start-all

fresh-start-all-traces-v2:
		make clean-environment
		make start-all-traces-v2

start-all-smc-v4:
		L1_GENESIS_TIME=$(get_future_time) make start-whole-environment
		make deploy-contracts-v4

start-all:
		L1_GENESIS_TIME=$(get_future_time) make start-whole-environment
		make deploy-contracts

start-all-traces-v2:
		L1_GENESIS_TIME=$(get_future_time) make start-whole-environment-traces-v2
		make deploy-contracts

deploy-contracts-v4:
	make compile-contracts
	$(MAKE) -j2 deploy-linea-rollup-v4 deploy-l2messageservice

deploy-contracts:
	make compile-contracts
	$(MAKE) -j2 deploy-linea-rollup deploy-l2messageservice

testnet-start-l2:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml --profile l2 up -d

testnet-start-l2-traces-node-only:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overries.yml up traces-node -d

testnet-start: start-l1 deploy-linea-rollup testnet-start-l2
testnet-restart-l2-keep-state:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml rm -f -s -v sequencer traces-node coordinator
		make testnet-start-l2

testnet-restart-l2-clean-state:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml rm -f -s -v sequencer traces-node coordinator
		docker volume rm testnet-data
		make clean-testnet-folders
		make testnet-start-l2

testnet-down:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml --profile l1 --profile l2 down -v
		make clean-testnet-folders

stop_pid:
		if [ -f $(PID_FILE) ]; then \
			kill `cat $(PID_FILE)`; \
			echo "Stopped process with PID `cat $(PID_FILE)`"; \
			rm $(PID_FILE); \
		else \
			echo "$(PID_FILE) does not exist. No process to stop."; \
		fi

restart-l2-minimal-stack-local:
		make stop-coordinator
		make stop-traces-api
		make stop_pid PID_FILE=tmp/local/traces-app.pid
		make clean-environment
		make start-l2
		make deploy-l2messageservice-to-local
		make start-traces-api
		make start-coordinator
		# TODO: use locally built prover for faster feedback loop

stop-l2-minimal-stack-local:
		make stop-coordinator
		make stop-traces-api
		make clean-environment

start-coordinator:
		mkdir -p  tmp/local/logs
		./gradlew coordinator:app:run \
			-Dconfig.override.testL1Disabled=true \
			-Dconfig.override.traces.counters.endpoints="http://127.0.0.1:8081" \
			-Dconfig.override.traces.conflation.endpoints="http://127.0.0.1:8081" \
			-Dconfig.override.dynamic-gas-price-service.miner-gas-price-update-recipients="http://127.0.0.1:8545/,http://127.0.0.1:8645/" > tmp/local/logs/coordinator.log & echo "$$!" > tmp/local/coordinator.pid

stop-coordinator:
		make stop_pid PID_FILE=tmp/local/coordinator.pid

restart-coordinator:
		make stop-coordinator
		make start-coordinator

start-traces-api:
		mkdir -p  tmp/local/logs
		mkdir -p  tmp/local/traces/raw
		./gradlew traces-api:app:run > tmp/local/logs/traces-app.log & echo "$$!" > tmp/local/traces-app.pid

stop-traces-api:
		make stop_pid PID_FILE=tmp/local/traces-app.pid

restart-traces-api:
		make stop-traces-api
		make start-traces-api
