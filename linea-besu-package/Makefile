TAG ?= local
PLATFORM ?= linux/arm64

assemble:
	./scripts/assemble-packages.sh

build:
	$(MAKE) assemble
	echo "IMAGE TAG: $(TAG)"
	echo "IMAGE PLATFORM: $(PLATFORM)"
	docker buildx build --file linea-besu/Dockerfile --platform $(PLATFORM) --tag consensys/linea-besu-package:$(TAG) ./tmp --progress=plain --no-cache

clean:
	rm -fr tmp || true
