.PHONY: build

build:
	./gradlew build

spotless-happy:
	./gradlew spotlessApply

pre-commit:
	$(MAKE) spotless-happy
	$(MAKE) build

run-e2e-test:
	./gradlew acceptanceTest

run-e2e-test-use-maru-container:
	./gradlew acceptanceTest -PuseMaruContainer=true

clean:
	./gradlew clean

build-local-image:
	./gradlew :app:installDist
	docker build app --build-context=libs=./app/build/install/app/lib/ --build-context=maru=./app/build/libs/ -t consensys/maru:local

run-local-image:
	CREATE_EMPTY_BLOCKS=true docker compose -f docker/compose.yaml -f docker/compose.dev.yaml up -d

run-local-image-partial:
	CREATE_EMPTY_BLOCKS=true docker compose -f docker/compose.yaml -f docker/compose.dev.yaml up -d maru

docker-clean-environment:
	docker compose -f docker/compose.yaml -f docker/compose.dev.yaml down || true
	docker volume rm maru-local-dev maru-logs || true # ignore failure if volumes do not exist already
	docker system prune -f || true
	rm docker/initialization/*.json || true # ignore failure if files do not exist already
