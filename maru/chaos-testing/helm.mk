helm-clean-releases:
	@echo "Cleaning up Helm releases"
	-@helm --kubeconfig $(KUBECONFIG) uninstall besu-sequencer
	-@helm --kubeconfig $(KUBECONFIG) uninstall besu-follower
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-0
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-2
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-3
	KUBECONFIG=$(KUBECONFIG) kubectl delete pvc --all
	KUBECONFIG=$(KUBECONFIG) kubectl delete pv --all

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
		@helm --kubeconfig $(KUBECONFIG) upgrade --install besu-sequencer ./helm/charts/besu --force -f ./helm/charts/besu/values.yaml -f ./helm/values/besu-local-dev-sequencer.yaml
		@helm --kubeconfig $(KUBECONFIG) upgrade --install besu-follower ./helm/charts/besu --force -f ./helm/charts/besu/values.yaml -f ./helm/values/besu-local-dev-follower.yaml
		@echo "Waiting for Besu to be ready..."
		@$(MAKE) wait_pods pod_name=besu-sequencer pod_count=1
		@$(MAKE) wait_pods pod_name=besu-follower pod_count=3

helm-redeploy-besu:
		-@helm --kubeconfig $(KUBECONFIG) uninstall besu-sequencer
		-@helm --kubeconfig $(KUBECONFIG) uninstall besu-follower
		@sleep 3 # Wait for a second to ensure the previous release is fully uninstalled
		@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-deploy-besu

IMAGE_ARG=$(if $(maru_image),--set image.name=$(maru_image),)
helm-redeploy-maru:
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-0
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-2
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-3
	@sleep 2 # Wait for a second to ensure the previous release is fully uninstalled
	@echo "Deploying Maru Nodes: IMAGE_ARG='$(IMAGE_ARG)'"
	@helm --kubeconfig $(KUBECONFIG) upgrade --install maru-bootnode-0 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-bootnode-0.yaml $(IMAGE_ARG);
	@$(MAKE) wait_pods pod_name=maru-bootnode-0 pod_count=1
	@$(MAKE) wait-for-log-entry pod_name=maru-bootnode-0-0 log_entry="enr"
	@BOOTNODE_ENR=$$(kubectl logs maru-bootnode-0-0 | grep -Ev '0.0.0.0|127.0.0.1' | grep -o 'enr=[^ ]*' | head -1 | cut -d= -f2); \
	echo "Bootnode ENR: $$BOOTNODE_ENR"; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-validator ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-validator.yaml --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG) ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-1 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-1.yaml --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG) ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-2 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-2.yaml --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG) ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-3 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-3.yaml --set bootnodes=$$BOOTNODE_ENR $(IMAGE_ARG)
	@$(MAKE) wait_pods pod_name=maru-validator pod_count=1

helm-redeploy-maru-and-besu:
	@echo "Redeploying Besu and Maru (maru_image=$(maru_image))"
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-releases
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-deploy-besu
	# Wait for Besu to be fully deployed,
	# otherwise Maru will fail to start because it cannot connect to Besu
	# then will miss P2P messages from validator
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-redeploy-maru $(if $(maru_image),maru_image=$(maru_image))

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

# Port-forward component pods exposing each pod's <port> on incremental local ports
# Usage:
#   make port-forward-all component=besu port=8545          -> 18545, 28545, ...
#   make port-forward-all component=maru port=8550          -> 18550, 28550, ...
# Optional vars: start_index (default 1)
# Backward compatibility: besu-port-forward-all still works (defaults component=besu remote_port=8545)
component ?= besu
port ?= 8545
start_index ?= 1
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TMP_DIR := $(MAKEFILE_DIR)/tmp

port-forward-all:
	@echo "Using tmp dir: $(TMP_DIR)"; \
	mkdir -p $(TMP_DIR); \
	summary_file="$(TMP_DIR)/port-forward-$(component)-$(port).txt"; \
	: > "$$summary_file"; \
	echo "Discovering pods (label app.kubernetes.io/component=$(component))..."; \
	pods=$$(kubectl get pods -l app.kubernetes.io/component=$(component) -o jsonpath='{.items[*].metadata.name}'); \
	if [ -z "$$pods" ]; then echo "No pods found for component $(component)"; exit 1; fi; \
	idx=$(start_index); \
	for pod in $$pods; do \
		local_port="$${idx}$(port)"; \
		if [ -z "$$local_port" ]; then echo "[ERROR] local_port empty (idx=$$idx port=$(port))"; idx=$$((idx+1)); continue; fi; \
		if lsof -i TCP:$$local_port -sTCP:LISTEN >/dev/null 2>&1; then \
			echo "Local port $$local_port in use, skipping $$pod"; \
			idx=$$((idx+1)); \
			continue; \
		fi; \
		log_file="$(TMP_DIR)/port-forward-$$pod.log"; \
		pid_file="$(TMP_DIR)/port-forward-$$pod.pid"; \
		echo "Port-forwarding $$pod :$(port) -> 127.0.0.1:$$local_port"; \
		kubectl port-forward $$pod $$local_port:$(port) > "$$log_file" 2>&1 & \
		pf_pid=$$!; \
		echo $$pf_pid > "$$pid_file"; \
		url="$$pod = http://127.0.0.1:$$local_port"; \
		echo "$$url" >> "$$summary_file"; \
		echo "Started pid $$pf_pid (log: $$log_file, url: $$url)"; \
		idx=$$((idx+1)); \
	done; \
	echo "Active forwards:"; \
	pids_files=$$(ls $(TMP_DIR)/port-forward-*.pid 2>/dev/null || true); \
	[ -n "$$pids_files" ] && ps -o pid,command -p $$(cat $$pids_files | tr '\n' ' ') 2>/dev/null || true; \
	echo "URL list written to $$summary_file";

port-forward-stop:
	@echo "Stopping pod port-forwards (tracked in $(TMP_DIR))..."; \
	for f in $(TMP_DIR)/port-forward-*.pid; do \
		[ -f "$$f" ] || continue; \
		pid=$$(cat $$f); \
		if kill -0 $$pid >/dev/null 2>&1; then \
			echo "Killing $$pid ($$f)"; \
			kill $$pid || true; \
		else \
			echo "Process $$pid already exited"; \
		fi; \
		rm -f $$f; \
	done; \
	echo "Done.";

# Stop ALL kubectl port-forward processes (regardless of PID files) and clean up stored PID/log files
# Usage: make port-forward-stop-all
port-forward-stop-all:
	@echo "Scanning for all kubectl port-forward processes..."; \
	pids=$$(ps -o pid= -o command= -ax | grep '[k]ubectl port-forward' | awk '{print $$1}'); \
	if [ -z "$$pids" ]; then \
		echo "No kubectl port-forward processes found."; \
	else \
		echo "Killing PIDs: $$pids"; \
		kill $$pids || true; \
		sleep 1; \
		for p in $$pids; do kill -0 $$p 2>/dev/null && echo "Force killing $$p" && kill -9 $$p || true; done; \
	fi; \
	echo "Removing tracked PID files in $(TMP_DIR)"; \
	rm -f $(TMP_DIR)/port-forward-*.pid 2>/dev/null || true; \
	echo "Done.";
