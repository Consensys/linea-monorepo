dokcer-pull-develop:
		docker compose -f docker/compose.yml --profile l2 pull

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
		docker-compose -f docker/compose.yml --profile l1 --profile l2 --profile debug --profile mev down || true
		make clean-local-folders
		docker network prune -f
		docker volume rm linea-local-dev linea-logs || true # ignore failure if volumes do not exist already

clean-environment-ci:
		docker-compose -f docker/compose.yml -f docker/compose-ci.overrides.yml --profile l1 --profile l2 down || true
		make clean-smc-folders
		docker network prune -f
		docker volume rm linea-local-dev linea-logs || true # ignore failure if volumes do not exist already

start-l1:
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l1 up -d

start-l1-ci:
		docker-compose -f docker/compose.yml --profile l1 up -d

start-l2:
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l2 up -d

start-l2-ci:
		docker-compose -f docker/compose.yml -f docker/compose-ci.overrides.yml --profile l2 up -d

start-l2-blockchain-only:
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l2-bc up -d

start-whole-environment:
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l1 --profile l2 --profile debug up -d

start-whole-environment-ci:
		docker-compose -f docker/compose.yml -f docker/compose-ci.overrides.yml --profile l1 --profile l2 up -d

start-l2-mev:
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l1 --profile l2 --profile mev up -d

pull-all-images-ci:
		docker-compose -f docker/compose.yml -f docker/compose-ci.overrides.yml --profile l1 --profile l2 pull

deploy-zkevm2-to-local:
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		cd contracts/; \
		sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8445/" .env.template > .env; \
		sed -i~ 's/PRIVATE_KEY=.*/PRIVATE_KEY=0xae99052fbea8f9c092f6d7c5132d585edc81aa1f11e4b1d18ac8fce0db44a078/' .env; \
		npx hardhat run ./scripts/deployment/deployZkEVM.ts --network zkevm_dev

deploy-zkevm2-to-ci:
		cd contracts/; \
		sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8445/" .env.template.ci > .env; \
		npx hardhat run ./scripts/deployment/deployZkEVM.ts --network zkevm_dev

deploy-l2messageservice-to-local:
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		cd contracts/; \
		sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8545/" .env.template > .env; \
		sed -i~ 's/L2MSGSERVICE_L1L2_MESSAGE_SETTER=.*/L2MSGSERVICE_L1L2_MESSAGE_SETTER="0x1B9AbEeC3215D8AdE8a33607f2cF0f4F60e5F0D0"/' .env; \
		sed -i~ 's/PRIVATE_KEY=.*/PRIVATE_KEY=0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae/' .env; \
		npx hardhat run ./scripts/deployment/deployL2MessageService.ts --network zkevm_dev

deploy-l2messageservice-to-ci:
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		cd contracts/; \
		sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8545/" .env.template.ci > .env; \
		sed -i~ 's/L2MSGSERVICE_L1L2_MESSAGE_SETTER=.*/L2MSGSERVICE_L1L2_MESSAGE_SETTER="0x1B9AbEeC3215D8AdE8a33607f2cF0f4F60e5F0D0"/' .env; \
		sed -i~ 's/PRIVATE_KEY=.*/PRIVATE_KEY=0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae/' .env; \
		npx hardhat run ./scripts/deployment/deployL2MessageService.ts --network zkevm_dev

upgrade-zkevm2-on-uat:
		cd contracts/; \
		rm -f .openzeppelin/goerli.json; \
		sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=https:\/\/goerli.infura.io\/v3\/${INFURA_KEY}/" .env.template.uat > .env; \
		sed -i~ "s/PRIVATE_KEY=.*/PRIVATE_KEY=${PRIVATE_KEY}/" .env; \
		npx hardhat run ./scripts/upgrades/upgradeZkEVMv2.ts --network zkevm_dev

fresh-start-l2-blockchain-only:
		make clean-environment
		make start-l2-blockchain-only

restart-shomei:
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml rm zkbesu-shomei shomei
		rm -rf tmp/local/shomei/*
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml up zkbesu-shomei shomei -d


fresh-start-all:
		make clean-environment
		make start-l1
		make deploy-zkevm2-to-local
		make start-whole-environment
		make deploy-l2messageservice-to-local

fresh-start-all-ci:
		make clean-environment-ci
		make start-l1-ci
		make deploy-zkevm2-to-ci
		make start-whole-environment-ci
		make deploy-l2messageservice-to-ci

start-all-ci:
		make start-whole-environment-ci
		make start-all-deploy-ci

start-all-deploy-ci: start-deploy-l1-ci start-deploy-l2-ci

start-deploy-l1-ci:
		make deploy-zkevm2-to-ci

start-deploy-l2-ci:
		make deploy-l2messageservice-to-ci

send-some-transactions-to-l2:
		cd ../zkevm-deployment/smart_contract/scripts; \
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		./txHelper.js transfer --blockchainNode http://localhost:8845 --privKey 0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae --wei 1 --transfers 1 --mode WAITEXEC --maxFeePerGas 1000000000 --maxPriorityFeePerGas 1000000000

send-some-transactions-to-builder:
		cd ../zkevm-deployment/smart_contract/scripts; \
		# WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
		./txHelper.js transfer --blockchainNode http://localhost:8580 --privKey 0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae --wei 1 --transfers 1 --mode WAITEXEC --maxFeePerGas 1000000000 --maxPriorityFeePerGas 1000000000

testnet-start-l2:
		docker-compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml --profile l2 up -d

testnet-start-l2-traces-node-only:
		docker-compose -f docker/compose.yml -f docker/compose-testnet-sync.overries.yml up traces-node -d

testnet-start: start-l1 deploy-zkevm2-to-local testnet-start-l2
testnet-restart-l2-keep-state:
		docker-compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml rm -f -s -v sequencer traces-node coordinator
		make testnet-start-l2

testnet-restart-l2-clean-state:
		docker-compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml rm -f -s -v sequencer traces-node coordinator
		docker volume rm testnet-data
		make clean-testnet-folders
		make testnet-start-l2

testnet-down:
		docker-compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml --profile l1 --profile l2 down -v
		make clean-testnet-folders

zkgeth-sequencer-smoke-test:
		make clean-environment
		docker-compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml up sequencer -d
		make deploy-l2messageservice-to-local-tmp

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
