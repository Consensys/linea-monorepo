# Maru
A consensus layer client implementing QBFT protocol adhering to Eth 2.0 CL / EL separation and API

## Requirements

- Java 21+
- Make 3.81

## Quick Start

```sh
docker-run-stack
```

## Build from sources

To build Maru from source code:

```sh
# Create a distribution ready to run
./gradlew :app:installDist
```

After building, you can run Maru using:

```sh
./app/build/install/app/bin/app [options]
```

The distribution will be created in `app/build/install/app/` with all necessary dependencies included.

### Build Docker Image Locally

```sh
docker-build-local-image
MARU_TAG=local make docker-run-stack
```
