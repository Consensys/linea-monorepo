chaos-mesh-install:
	#TODO: refactor to run helm once
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
	@kubectl get pods -n chaos-mesh -l app.kubernetes.io/instance=chaos-mesh

chaos-mesh-uninstall:
	@kubectl delete ns chaos-mesh

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

.PHONY: chaos-mesh-install-with-curl \
	chaos-mesh-install \
	chaos-mesh-uninstall \
	chaos-mesh-run-network-experiment \
	chaos-mesh-reload
