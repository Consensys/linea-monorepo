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

clean:
	./gradlew clean

build-local-image:
	./gradlew :app:installDist
	docker build app --build-context=libs=./app/build/install/app/lib/ --build-context=maru=./app/build/libs/ -t local/maru:latest

run-local-image:
	docker compose -f docker/compose.yaml -f docker/compose.dev.yaml up -d