include makefile-contracts.mk

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

docker-pull-all-images:
		COMPOSE_PROFILES:=l1,l2 docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml pull

docker-pull-develop:
	L1_GENESIS_TIME=$(get_future_time) docker compose -f docker/compose.yml pull

docker-pull-images-external-to-monorepo:
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile external-to-monorepo pull

clean-local-folders:
		make clean-smc-folders
		rm -rf tmp/local/*

clean-testnet-folders:
		make clean-smc-folders
		rm -rf tmp/testnet/*

clean-environment:
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile l1 --profile l2 --profile debug --profile staterecovery kill -s 9 || true
		docker compose -f docker/compose.yml -f docker/compose-local-dev-traces-v2.overrides.yml --profile l1 --profile l2 --profile debug --profile staterecovery down || true
		make clean-local-folders
		docker volume rm linea-local-dev linea-logs || true # ignore failure if volumes do not exist already
		docker system prune -f || true



start-l1:
		L1_GENESIS_TIME=$(get_future_time) docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l1 up -d

start-l2:
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l2 up -d

start-l2-blockchain-only:
		docker compose -f docker/compose.yml -f docker/compose-local-dev.overrides.yml --profile l2-bc up -d


fresh-start-l2-blockchain-only:
		make clean-environment
		make start-l2-blockchain-only


##
## Creating new targets to avoid conflicts with existing targets
## Redundant targets above will cleanup once this get's merged
##
start-env: COMPOSE_PROFILES:=l1,l2
start-env: COMPOSE_FILE:=docker/compose-tracing-v2.yml
start-env: L1_CONTRACT_VERSION:=6
start-env: LINEA_PROTOCOL_CONTRACTS_ONLY:=false
start-env:
	L1_GENESIS_TIME=$(get_future_time) COMPOSE_PROFILES=$(COMPOSE_PROFILES) docker compose -f $(COMPOSE_FILE) up -d
	make deploy-contracts L1_CONTRACT_VERSION=$(L1_CONTRACT_VERSION) LINEA_PROTOCOL_CONTRACTS_ONLY=$(LINEA_PROTOCOL_CONTRACTS_ONLY)

start-env-with-tracing-v1:
	make start-env COMPOSE_FILE=docker/compose-tracing-v1.yml LINEA_PROTOCOL_CONTRACTS_ONLY=true

start-env-with-tracing-v1-ci:
	make start-env COMPOSE_FILE=docker/compose-tracing-v1-ci-extension.yml

start-env-with-tracing-v2:
	make start-env COMPOSE_FILE=docker/compose-tracing-v2.yml LINEA_PROTOCOL_CONTRACTS_ONLY=true

start-env-with-tracing-v2-ci:
	make start-env COMPOSE_FILE=docker/compose-tracing-v2-ci-extension.yml

fresh-start-all: COMPOSE_PROFILES:="l1,l2"
fresh-start-all: L1_CONTRACT_VERSION:=6
fresh-start-all:
		make clean-environment
		make start-all L1_CONTRACT_VERSION=$(L1_CONTRACT_VERSION) COMPOSE_PROFILES=$(COMPOSE_PROFILES)

fresh-start-all-traces-v2: COMPOSE_PROFILES:="l1,l2"
fresh-start-all-traces-v2: L1_CONTRACT_VERSION:=6
fresh-start-all-traces-v2:
		make clean-environment
		$(MAKE) start-all-traces-v2 L1_CONTRACT_VERSION=$(L1_CONTRACT_VERSION) COMPOSE_PROFILES=$(COMPOSE_PROFILES)

start-all: COMPOSE_PROFILES:=l1,l2
start-all: L1_CONTRACT_VERSION:=6
start-all:
		L1_GENESIS_TIME=$(get_future_time) make start-whole-environment COMPOSE_PROFILES=$(COMPOSE_PROFILES)
		make deploy-contracts L1_CONTRACT_VERSION=$(L1_CONTRACT_VERSION)

start-all-traces-v2: COMPOSE_PROFILES:="l1,l2"
start-all-traces-v2: L1_CONTRACT_VERSION:=6
start-all-traces-v2:
		L1_GENESIS_TIME=$(get_future_time) make start-whole-environment-traces-v2 COMPOSE_PROFILES=$(COMPOSE_PROFILES)
		$(MAKE) deploy-contracts L1_CONTRACT_VERSION=$(L1_CONTRACT_VERSION)

fresh-start-all-staterecovery: COMPOSE_PROFILES:=l1,l2,staterecovery
fresh-start-all-staterecovery: L1_CONTRACT_VERSION:=6
fresh-start-all-staterecovery:
	make clean-environment
	make start-env COMPOSE_FILE=docker/compose-tracing-v2-staterecovery-extension.yml LINEA_PROTOCOL_CONTRACTS_ONLY=true L1_CONTRACT_VERSION=$(L1_CONTRACT_VERSION)

staterecovery-replay-from-block: L1_ROLLUP_CONTRACT_ADDRESS:=0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9
staterecovery-replay-from-block: STATERECOVERY_OVERRIDE_START_BLOCK_NUMBER:=1
staterecovery-replay-from-block:
	docker compose -f docker/compose-tracing-v2-staterecovery-extension.yml down zkbesu-shomei-sr shomei-sr
	L1_ROLLUP_CONTRACT_ADDRESS=$(L1_ROLLUP_CONTRACT_ADDRESS) STATERECOVERY_OVERRIDE_START_BLOCK_NUMBER=$(STATERECOVERY_OVERRIDE_START_BLOCK_NUMBER) docker compose -f docker/compose-tracing-v2-staterecovery-extension.yml up zkbesu-shomei-sr shomei-sr -d

##
# Testnet
##
testnet-start-l2:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overrides.yml --profile l2 up -d

testnet-start-l2-traces-node-only:
		docker compose -f docker/compose.yml -f docker/compose-testnet-sync.overries.yml up traces-node -d

testnet-start: start-l1 deploy-linea-rollup-v6 testnet-start-l2
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


