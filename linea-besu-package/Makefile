include versions.env

TAG ?= local
PLATFORM ?=

assemble:
	./scripts/assemble-packages.sh

build:
	$(MAKE) assemble
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
	sed -i '' 's/^\(expected-traces-api-version-v2=\).*/\1"${LINEA_TRACER_PLUGIN_VERSION}"/' ../config/coordinator/coordinator-docker.config.toml
	sed -i '' 's/^\(expected-traces-api-version-v2=\).*/\1"${LINEA_TRACER_PLUGIN_VERSION}"/' ../config/coordinator/coordinator-docker-traces-v2-override.config.toml
	cd .. && BESU_PACKAGE_TAG=$(TAG) make start-env-with-tracing-v2-ci && pnpm run -F e2e test:e2e:local

clean:
	rm -fr tmp || true
