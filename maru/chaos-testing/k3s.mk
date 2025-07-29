# Helper Makefile for managing a k3s server inside a local Docker container.
# Handy to quickly validate Helm charts, run local tests and Chaos experiments
# without need to setup a full Kubernetes cluster locally.

k3s-clean:
	-@docker stop k3s-server
	-@docker rm k3s-server

k3s-run:
	@docker run -d --name k3s-server \
		--privileged \
		-p 6443:6443 \
		-p 80:80 \
		-p 443:443 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		rancher/k3s:v1.33.2-k3s1 \
		server --token a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p

k3s-setup-kubeconfig:
	@docker exec k3s-server sh -c 'while [ ! -f /etc/rancher/k3s/k3s.yaml ]; do sleep 1; done'
	@docker cp k3s-server:/etc/rancher/k3s/k3s.yaml ~/.kube/k3s-server
	#@docker exec k3s-server sh -c "cat /etc/rancher/k3s/k3s.yaml" > ~/.kube/k3s-server
	# export KUBECONFIG=~/.kube/k3s-server

k3s-wait:
	@echo "Waiting for k3s server to be up..."
	@until docker exec k3s-server sh -c "telnet localhost 6443 </dev/null 2>/dev/null | grep Connected" ; do \
		sleep 1; \
	done
	@echo "k3s server is up and running."

k3s-reload:
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) k3s-clean
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) k3s-run
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) k3s-setup-kubeconfig
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) k3s-wait
