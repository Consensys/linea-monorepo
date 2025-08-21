include chaos-mesh.mk
include helm.mk
include k3s.mk

.PHONY: chaos-full-reload \
	chaos-mesh-install-with-curl \
	chaos-mesh-install \
	chaos-mesh-uninstall \
	chaos-mesh-run-network-experiment \
	chaos-mesh-reload \
	helm-clean-releases \
	helm-redeploy-besu \
	helm-redeploy-maru \

chaos-deploy-chaos-mesh-and-linea:
	@set -e; \
	pid1=""; \
	if [ "$(maru_image)" ]; then \
		case "$(maru_image)" in \
			*local) \
				echo "maru_image ends with 'local' -> build/import local image"; \
				$(MAKE) build-and-import-maru-image & pid1=$$!; \
				;; \
			*) \
				echo "Using provided maru_image=$(maru_image)"; \
				;; \
		esac; \
	fi; \
	$(MAKE) chaos-mesh-install & pid2=$$!; \
	$(MAKE) helm-redeploy-besu & pid3=$$!; \
	if [ -n "$$pid1" ]; then wait $$pid1 || exit 1; fi; \
	wait $$pid2 || exit 1; \
	wait $$pid3 || exit 1; \
	$(MAKE) helm-redeploy-maru maru_image=$(if $(maru_image),$(maru_image))

chaos-full-reload:
	@$(MAKE) k3s-reload
	@$(MAKE) chaos-deploy-chaos-mesh-and-linea maru_image=$(if $(maru_image),$(maru_image))
	@$(MAKE) chaos-experiment-workflow-and-wait # runs workflow-linea-resilience.yaml and waits for it to finish
	@$(MAKE) chaos-health-check

chaos-full-reload-with-local-maru-image:
	@$(MAKE) k3s-reload
	@$(MAKE) chaos-deploy-chaos-mesh-and-linea maru_image=consensys/maru:local
	@$(MAKE) chaos-experiment-workflow-and-wait # runs workflow-linea-resilience.yaml and waits for it to finish
	@$(MAKE) chaos-health-check

chaos-health-check:
	@echo "Running health check for chaos testing environment"
	@echo "Creating port-forward besu and maru nodes"
	-@$(MAKE) port-forward-stop-all
	@$(MAKE) port-forward-all component=besu port=8545
	@$(MAKE) port-forward-all component=maru port=5060
	@cd .. && ./gradlew :chaos-testing:health-check:acceptanceTest

chaos-redeploy-and-run-experiment:
	@$(MAKE) helm-redeploy-maru-and-besu
	@$(MAKE) wait-maru-follower-is-syncing
	@sleep 20 # wait for some time to ensure nodes are cruising and ready for chaos experiments
	@$(MAKE) chaos-experiment-podkill-besu-nodes

chaos-experiment-workflow-and-health-check:
	@$(MAKE) chaos-experiment-workflow-and-wait
	@$(MAKE) chaos-health-check

build-maru-image:
	@echo "Building Maru image"
	cd .. && ./gradlew :app:installDist
	cd .. && docker build app --build-context=libs=./app/build/install/app/lib/ --build-context=maru=./app/build/libs/ -t consensys/maru:local

build-and-import-maru-image:
	@$(MAKE) build-maru-image
	@$(MAKE) k3s-import-local-maru-image

build-and-redeploy-maru:
	@$(MAKE) build-and-import-maru-image
	@$(MAKE) helm-redeploy-maru
