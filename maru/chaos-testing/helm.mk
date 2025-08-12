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

helm-redeploy-maru:
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-0
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-2
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-3
	@sleep 2 # Wait for a second to ensure the previous release is fully uninstalled
	@echo "Deploying Maru Nodes"
	@helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-0 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-0.yaml
	@$(MAKE) wait_pods pod_name=maru-follower-0 pod_count=1
	@$(MAKE) wait-for-log-entry pod_name=maru-follower-0-0 log_entry="enr"
	BOOTNODE_ENR=$$(kubectl logs maru-follower-0-0 | grep -Ev '0.0.0.0|127.0.0.1' | grep -o 'enr=[^ ]*' | head -1 | cut -d= -f2); \
	echo "Bootnode ENR: $$BOOTNODE_ENR"; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-validator ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-validator.yaml --set bootnodes=$$BOOTNODE_ENR ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-1 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-1.yaml --set bootnodes=$$BOOTNODE_ENR ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-2 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-2.yaml --set bootnodes=$$BOOTNODE_ENR ; \
	helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-3 ./helm/charts/maru --force -f ./helm/charts/maru/values.yaml -f ./helm/values/maru-local-dev-follower-3.yaml --set bootnodes=$$BOOTNODE_ENR

	@$(MAKE) wait_pods pod_name=maru-validator pod_count=1

helm-redeploy-maru-and-besu:
	@echo "Redeploying Besu and Maru"
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-clean-releases
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-deploy-besu
	# Wait for Besu to be fully deployed,
	# otherwise Maru will fail to start because it cannot connect to Besu
	# then will miss P2P messages from validator
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) helm-redeploy-maru

wait-maru-follower-is-syncing:
	@echo "Waiting for Maru follower to be ready..."
	@until kubectl get pods -n default -l app.kubernetes.io/instance=maru-follower-0 | grep -q '1/1'; do \
		sleep 1; \
	done
	@echo "Maru follower is ready."
	@echo "Waiting for 'block received' message in maru-follower-0 pod..."
	@until kubectl logs -n default -l app.kubernetes.io/instance=maru-follower-0 | grep -q 'block received'; do \
		sleep 1; \
	done
