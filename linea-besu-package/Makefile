include versions.env

TAG ?= local
PLATFORM ?=
LOCAL_SEQUENCER_ZIP_FOLDER ?= ../besu-plugins/linea-sequencer/sequencer/build/distributions
LOCAL_TRACER_ZIP_FOLDER ?= ../tracer/plugins/build/distributions
ARITHMETIZATION_VERSION ?= $(shell grep -E '^\s*arithmetization\s*=' ../gradle/releases.versions.toml | sed -n 's/.*"\([^"]*\)".*/\1/p')
SEQUENCER_VERSION ?= v$(shell grep -E '^\s*sequencer\s*=' ../gradle/releases.versions.toml | sed -n 's/.*"\([^"]*\)".*/\1/p')

build-tracer-and-sequencer-local:
	cd ..; \
	echo "Build versions: ARITHMETIZATION_VERSION=$(ARITHMETIZATION_VERSION) SEQUENCER_VERSION=$(SEQUENCER_VERSION)" ; \
	echo "Clean and build zip distribution for tracer" ; \
	./gradlew :tracer:clean && ./gradlew :tracer:artifacts -PreleaseVersion=$(ARITHMETIZATION_VERSION) ; \
	echo "Clean and build zip distribution for sequencer" ; \
	./gradlew besu-plugins:linea-sequencer:sequencer:clean && ./gradlew besu-plugins:linea-sequencer:sequencer:artifacts -PreleaseVersion=$(SEQUENCER_VERSION)

assemble:
	./scripts/assemble-packages.sh

assemble-local:
	$(MAKE) clean
	$(MAKE) build-tracer-and-sequencer-local
	LOCAL_SEQUENCER_ZIP_PATH=$(LOCAL_SEQUENCER_ZIP_FOLDER)/linea-sequencer-$(SEQUENCER_VERSION).zip LOCAL_TRACER_ZIP_PATH=$(LOCAL_TRACER_ZIP_FOLDER)/linea-tracer-$(ARITHMETIZATION_VERSION).zip ./scripts/assemble-packages.sh

build:
	$(MAKE) assemble
	$(MAKE) build-image

build-local:
	$(MAKE) assemble-local
	$(MAKE) build-image

build-image:
	echo "IMAGE TAG=$(TAG)"
	echo "PLATFORM=$(PLATFORM)"
	if [ "$(PLATFORM)" = "" ]; then \
		echo "Building image for default platform"; \
		docker buildx build --file linea-besu/Dockerfile --tag consensys/linea-besu-package:$(TAG) ./tmp --progress=plain --no-cache --load; \
	else \
		echo "Building image for $(PLATFORM)"; \
		docker buildx build --file linea-besu/Dockerfile --platform $(PLATFORM) --tag consensys/linea-besu-package:$(TAG) ./tmp --progress=plain --no-cache --load; \
	fi

run-e2e-test:
	echo "EXPECTED_TRACES_API_VERSION=${LINEA_TRACER_PLUGIN_VERSION}"
	if [ "$(shell uname)" = "Linux" ]; then \
		sed -i'' 's/^\(expected-traces-api-version[ ]*=[ ]*\).*/\1"${LINEA_TRACER_PLUGIN_VERSION}"/' ../config/coordinator/coordinator-config-v2.toml; \
	elif [ "$(shell uname)" = "Darwin" ]; then \
		sed -i '' 's/^\(expected-traces-api-version[ ]*=[ ]*\).*/\1"${LINEA_TRACER_PLUGIN_VERSION}"/' ../config/coordinator/coordinator-config-v2.toml; \
	fi
	cd .. && BESU_PACKAGE_TAG=$(TAG) make start-env-with-tracing-v2-ci && pnpm run -F e2e test:e2e:local

run-e2e-test-cleanup:
	cd .. && make clean-environment

clean:
	rm -fr tmp || true
