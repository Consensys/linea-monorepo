IMAGE ?= consensys/linea-erigon
TAG ?= v3.3.0-linea-patched
PLATFORM ?= linux/amd64

.PHONY: build push buildx-info

buildx-info:
	@docker buildx version
	@docker buildx ls

# Build locally and load into your Docker daemon (for testing)
build:
	docker buildx build \
		--platform $(PLATFORM) \
		--file Dockerfile \
		--tag $(IMAGE):$(TAG) \
		--progress=plain \
		--no-cache \
		--load \
		.

# Build and push to Docker Hub (requires `docker login` with permissions to the ConsenSys org)
push:
	docker buildx build \
		--platform $(PLATFORM) \
		--file Dockerfile \
		--tag $(IMAGE):$(TAG) \
		--progress=plain \
		--no-cache \
		--push \
		.

