clean-releases:
	@echo "Cleaning up Helm releases"
	-@helm --kubeconfig $(KUBECONFIG) uninstall besu-sequencer
	-@helm --kubeconfig $(KUBECONFIG) uninstall besu-follower
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-0
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1
	KUBECONFIG=$(KUBECONFIG) kubectl delete pvc --all
	KUBECONFIG=$(KUBECONFIG) kubectl delete pv --all

redeploy-besu:
		@echo "Redeploying Besu"
		-@helm --kubeconfig $(KUBECONFIG) uninstall besu-sequencer
		-@helm --kubeconfig $(KUBECONFIG) uninstall besu-follower
		@sleep 3 # Wait for a second to ensure the previous release is fully uninstalled
		@helm --kubeconfig $(KUBECONFIG) upgrade --install besu-sequencer ./charts/besu --force -f ./charts/besu/values.yaml -f ./values/besu-local-dev-sequencer.yaml
		@helm --kubeconfig $(KUBECONFIG) upgrade --install besu-follower ./charts/besu --force -f ./charts/besu/values.yaml -f ./values/besu-local-dev-follower.yaml

redeploy-maru:
	@echo "Redeploying Maru"
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-validator
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-0
	-@helm --kubeconfig $(KUBECONFIG) uninstall maru-follower-1
	@sleep 2 # Wait for a second to ensure the previous release is fully uninstalled
	@echo "Deploying Maru Followers"
	@helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-0 ./charts/maru --force -f ./charts/maru/values.yaml -f ./values/maru-local-dev-follower-0.yaml
	@helm --kubeconfig $(KUBECONFIG) upgrade --install maru-follower-1 ./charts/maru --force -f ./charts/maru/values.yaml -f ./values/maru-local-dev-follower-1.yaml
	@sleep 3 # wait for followers to start
	@echo "Deploying Maru Validator"
	@helm --kubeconfig $(KUBECONFIG) upgrade --install maru-validator ./charts/maru --force -f ./charts/maru/values.yaml -f ./values/maru-local-dev-validator.yaml

redeploy:
	@echo "Redeploying Besu and Maru"
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) clean-releases
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) redeploy-besu
	@sleep 10
	# Wait for Besu to be fully deployed,
	# otherwise Maru will fail to start because it cannot connect to Besu
	# then will miss P2P messages from validator
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) redeploy-maru
