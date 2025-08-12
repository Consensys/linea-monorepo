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

chaos-full-reload:
	@$(MAKE) k3s-reload
	@sleep 5
	@$(MAKE) helm-redeploy-maru-and-besu
	@$(MAKE) chaos-mesh-install
	@$(MAKE) wait-maru-follower-is-syncing
	@sleep 30 # wait for some time to ensure nodes are cruising and ready for chaos experiments
	@$(MAKE) chaos-experiment-podkill-besu-nodes
	@echo "Chaos testing environment is ready. You can now run your chaos experiments."

chaos-full-reload-with-local-maru-image:
	@$(MAKE) k3s-reload
	@sleep 5
	@$(MAKE) build-and-import-maru-image
	@$(MAKE) helm-redeploy-maru-and-besu

chaos-redeploy-and-run-experiment:
	@$(MAKE) helm-redeploy-maru-and-besu
	@$(MAKE) wait-maru-follower-is-syncing
	@sleep 20 # wait for some time to ensure nodes are cruising and ready for chaos experiments
	@$(MAKE) chaos-experiment-podkill-besu-nodes

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
