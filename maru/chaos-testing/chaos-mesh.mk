chaos-mesh-install:
	@helm repo add chaos-mesh https://charts.chaos-mesh.org
	@helm repo update
	@helm install chaos-mesh chaos-mesh/chaos-mesh \
		--namespace chaos-mesh --create-namespace \
		--version 2.7.2 \
		--set controllerManager.create=true \
		--set controllerManager.clusterScoped=true \
		--set chaosDaemon.create=true \
		--set chaosDaemon.runtime=containerd \
		--set chaosDaemon.socketPath=/run/k3s/containerd/containerd.sock \
		--set dashboard.create=true \
		--set dashboard.securityMode=false \
		--set dashboard.service.type=NodePort

chaos-mesh-uninstall:
	-@kubectl delete ns chaos-mesh >/dev/null 2>&1

chaos-mesh-reinstall:
	-@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) chaos-mesh-uninstall
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) chaos-mesh-install

chaos-experiment-http-engine-api-error:
	-@kubectl delete -f experiments/http-engine-api-error.yaml
	@kubectl apply -f experiments/http-engine-api-error.yaml

chaos-experiment-podkill-validator:
	@kubectl apply -f experiments/podkill-validator.yaml
	@kubectl get pods -l app.kubernetes.io/instance=maru-validator

chaos-experiment-podkill-besu-nodes:
	-@kubectl delete -f experiments/podkill-besu-once.yaml
	@kubectl apply -f experiments/podkill-besu-once.yaml
	@kubectl get pods -l app.kubernetes.io/component=besu

chaos-experiment-workflow:
	-@kubectl delete -f experiments/workflow-linea-resilience.yaml
	@kubectl apply -f experiments/workflow-linea-resilience.yaml

experiment_name ?= linea-resilience
wait-experiment-done:
	@echo "Waiting for workflow $(experiment_name) to be accomplished..."
	@timeout=$${WAIT_TIMEOUT_SECONDS:-3000}; interval=$${WAIT_INTERVAL_SECONDS:-10}; elapsed=0; \
	while [ $$elapsed -lt $$timeout ]; do \
	  out=$$(kubectl describe workflows.chaos-mesh.org $(experiment_name) -n chaos-mesh 2>/dev/null || true); \
	  echo "--- status check (elapsed $$elapsed s) ---"; \
	  echo "$$out" | grep -qi 'WorkflowAccomplished' && { echo "Workflow $(experiment_name) accomplished."; exit 0; }; \
	  echo "$$out" | grep -qi 'Failed' && { echo "Workflow $(experiment_name) failed."; exit 1; }; \
	  sleep $$interval; elapsed=$$((elapsed+interval)); \
	done; \
	echo "Timeout ($$timeout s) waiting for workflow $(experiment_name)."; exit 1

chaos-experiment-workflow-and-wait:
	@$(MAKE) chaos-experiment-workflow
	@$(MAKE) wait-experiment-done experiment_name=linea-resilience

.PHONY: chaos-mesh-install-with-curl \
	chaos-mesh-install \
	chaos-mesh-uninstall \
	chaos-mesh-run-network-experiment \
	chaos-mesh-reload \
	wait-experiment-done
