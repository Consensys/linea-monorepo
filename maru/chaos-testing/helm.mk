.SILENT:

helm-clean-pvcs:
	@echo "Cleaning up Persistent Volumes Claims and correspondent PV"
	-@KUBECONFIG=$(KUBECONFIG) kubectl delete pvc -l app.kubernetes.io/component=$(component) >/dev/null 2>&1
	-@KUBECONFIG=$(KUBECONFIG) kubectl get pv --no-headers 2>/dev/null | awk '$$5=="Available" {print $$1}' | xargs -r -I {} env KUBECONFIG=$(KUBECONFIG) kubectl delete pv {} >/dev/null 2>&1

helm-clean-besu-releases:
	@echo "Cleaning up Besu Helm releases"
	-@helm --kubeconfig $(KUBECONFIG) uninstall besu-sequencer >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall besu-follower >/dev/null 2>&1
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-pvcs component=besu

helm-clean-maru-releases:
	@echo "Cleaning up MARU releases"
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-bootnode-0 >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1 >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-2 >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-3 >/dev/null 2>&1
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-pvcs component=maru

helm-clean-linea-releases:
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-besu-releases
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-maru-releases

wait_pods:
	@if [ -z "$(pod_name)" ] || [ -z "$(pod_count)" ]; then \
		echo "Usage: make wait_pods pod_name=<pod_name> pod_count=<pod_count>"; \
		exit 1; \
	fi; \
	echo "Waiting for $(pod_count) pods with label app.kubernetes.io/pod-name=$(pod_name) to be up and running..."; \
	while true; do \
		current_count=$$(kubectl get pods -l app.kubernetes.io/name=$(pod_name) --field-selector=status.phase=Running -o json | jq '.items | length'); \
		if [ "$$current_count" -ge "$(pod_count)" ]; then \
			echo "$$current_count $$pod_name pods are up and running."; \
			break; \
		fi; \
		echo "$$current_count $$pod_name pods are running. Waiting..."; \
		sleep 5; \
	done

wait-for-log-entry:
	@until kubectl logs $(pod_name) | grep -q "$(log_entry)"; do \
		sleep 1; \
	done

helm-deploy-besu:
		@echo "Deploying Besu"
		@kubectl config current-context
		@helm --kubeconfig $(KUBECONFIG) upgrade --install besu-sequencer ./helm/charts/besu --force -f ./helm/charts/besu/values.yaml -f ./helm/values/besu-local-dev-sequencer.yaml --namespace default
		@$(MAKE) wait-for-log-entry pod_name=besu-sequencer-0 log_entry="enode"
		@BOOTNODE_IP=$$(kubectl describe pod besu-sequencer-0 | grep -E '^IP:' | awk '{print $$2}'); \
		echo "Pod IP: $$BOOTNODE_IP"; \
		BOOTNODE=$$(kubectl logs besu-sequencer-0 | grep -o 'enode://[^[:space:]]*' | sed "s/127\.0\.0\.1/$$BOOTNODE_IP/"); \
		echo "Bootnode: $$BOOTNODE"; \
		helm --kubeconfig $(KUBECONFIG) upgrade --install besu-follower ./helm/charts/besu --force -f ./helm/charts/besu/values.yaml -f ./helm/values/besu-local-dev-follower.yaml --namespace default --set bootnodes=$$BOOTNODE
		@$(MAKE) wait_pods pod_name=besu-sequencer pod_count=1
		@$(MAKE) wait_pods pod_name=besu-follower pod_count=3

helm-redeploy-besu:
		-@helm --kubeconfig $(KUBECONFIG) uninstall besu-sequencer >/dev/null 2>&1
		-@helm --kubeconfig $(KUBECONFIG) uninstall besu-follower >/dev/null 2>&1
		@sleep 3 # Wait for a second to ensure the previous release is fully uninstalled
		@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-deploy-besu

IMAGE_ARG=$(if $(maru_image),--set image.name=$(maru_image),)
helm-redeploy-maru:
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-0 >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1 >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-2 >/dev/null 2>&1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-3 >/dev/null 2>&1
	@sleep 2 # Wait for a second to ensure the previous release is fully uninstalled
	@echo "Deploying Maru Nodes: IMAGE_ARG='$(IMAGE_ARG)'"
	@helm --kubeconfig $(KUBECONFIG) upgrade --install maru-bootnode-0 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-bootnode-0.yaml --namespace default $(IMAGE_ARG);
	@$(MAKE) wait_pods pod_name=maru-bootnode-0 pod_count=1
	@$(MAKE) wait-for-log-entry pod_name=maru-bootnode-0-0 log_entry="enr"
	@BOOTNODE_ENR=$$(kubectl logs maru-bootnode-0-0 | grep -Ev '0.0.0.0|127.0.0.1' | grep -o 'enr=[^ ]*' | head -1 | cut -d= -f2); \
	echo "Bootnode ENR: $$BOOTNODE_ENR"; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-validator ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-validator.yaml   --namespace default --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG) ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-1 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-1.yaml --namespace default --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG) ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-2 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-2.yaml --namespace default --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG) ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-3 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-3.yaml --namespace default --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG)
	@$(MAKE) wait_pods pod_name=maru-validator pod_count=1

helm-redeploy-linea:
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-linea-releases
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
	$(MAKE) helm-redeploy-besu & pid2=$$!; \
	if [ -n "$$pid1" ]; then wait $$pid1 || exit 1; fi; \
	wait $$pid2 || exit 1; \
	$(MAKE) helm-redeploy-maru maru_image=$(if $(maru_image),$(maru_image))

wait-maru-follower-is-syncing:
	@echo "Waiting for Maru follower to be ready..."
	@until kubectl get pods -n default -l app.kubernetes.io/instance=maru-bootnode-0 | grep -q '1/1'; do \
		sleep 1; \
	done
	@echo "Maru follower is ready."
	@echo "Waiting for 'block received' message in maru-bootnode-0 pod..."
	@until kubectl logs -n default -l app.kubernetes.io/instance=maru-bootnode-0 | grep -q 'block received'; do \
		sleep 1; \
	done

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

# Port-forward component pods exposing each pod's <port> on incremental local ports
# Usage:
#   make port-forward-all component=besu port=8545          -> 1000, 1001, ...
#   make port-forward-all component=maru port=8550          -> 1000, 1001, ...
# Optional vars: local_port_start_number (default 1000)
# Backward compatibility: besu-port-forward-all still works (defaults component=besu remote_port=8545)
component ?= besu
port ?= 8545
local_port_start_number ?= 1000
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TMP_DIR := $(MAKEFILE_DIR)/tmp

port-forward-component:
	@echo "Using tmp dir: $(TMP_DIR)"; \
	mkdir -p $(TMP_DIR); \
	summary_file="$(TMP_DIR)/port-forward-$(component)-$(port).txt"; \
	: > "$$summary_file"; \
	echo "Discovering pods (label app.kubernetes.io/component=$(component))..."; \
	pods=$$(kubectl get pods -l app.kubernetes.io/component=$(component) -o jsonpath='{.items[*].metadata.name}'); \
	if [ -z "$$pods" ]; then echo "No pods found for component $(component)"; exit 1; fi; \
	current_port=$(local_port_start_number); \
	for pod in $$pods; do \
		while lsof -i TCP:$$current_port -sTCP:LISTEN >/dev/null 2>&1; do \
			echo "Local port $$current_port in use, trying next..."; \
			current_port=$$((current_port + 1)); \
		done; \
		log_file="$(TMP_DIR)/port-forward-$$pod.log"; \
		pid_file="$(TMP_DIR)/port-forward-$$pod.pid"; \
		echo "Port-forwarding $$pod :$(port) -> 127.0.0.1:$$current_port"; \
		kubectl port-forward $$pod $$current_port:$(port) > "$$log_file" 2>&1 & \
		pf_pid=$$!; \
		echo $$pf_pid > "$$pid_file"; \
		url="$$pod = http://127.0.0.1:$$current_port"; \
		echo "$$url" >> "$$summary_file"; \
		echo "Started pid $$pf_pid (log: $$log_file, url: $$url)"; \
		current_port=$$((current_port + 1)); \
	done; \
	echo "Active forwards:"; \
	pids_files=$$(ls $(TMP_DIR)/port-forward-*.pid 2>/dev/null || true); \
	[ -n "$$pids_files" ] && ps -o pid,command -p $$(cat $$pids_files | tr '\n' ' ') 2>/dev/null || true; \
	echo "URL list written to $$summary_file";

port-forward-linea:
	$(MAKE) port-forward-component component=maru port=5060 local_port_start_number=1100
	$(MAKE) port-forward-component component=besu port=8545 local_port_start_number=1200

port-forward-stop-component:
	@echo "Scanning for all kubectl port-forward processes..."; \
	pids=$$(ps -o pid= -o command= -ax | grep 'kubectl port-forward' | grep -v grep | grep $(component) | awk '{print $$1}'); \
	if [ -z "$$pids" ]; then \
		echo "No kubectl port-forward processes found."; \
	else \
		for pid in $$pids; do \
			if kill -0 $$pid >/dev/null 2>&1; then \
				echo "Killing PID $$pid"; \
				kill $$pid 2>/dev/null || echo "Failed to kill $$pid with SIGTERM"; \
			fi; \
		done; \
		sleep 1; \
		for pid in $$pids; do \
			if kill -0 $$pid >/dev/null 2>&1; then \
				echo "Force killing PID $$pid"; \
				kill -9 $$pid 2>/dev/null || echo "Failed to force kill $$pid"; \
			fi; \
		done; \
	fi; \
	echo "Removing tracked PID files in $(TMP_DIR)"; \
	rm -f $(TMP_DIR)/port-forward-*.pid 2>/dev/null || true; \
	echo "Done.";

port-forward-stop-all-linea:
	$(MAKE) port-forward-stop-component component=maru;
	$(MAKE) port-forward-stop-component component=besu;

port-forward-restart-all-linea:
	-$(MAKE) port-forward-stop-all-linea
	$(MAKE) port-forward-linea
