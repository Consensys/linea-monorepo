# Besu Plugins related to tracer and sequencer functionality

A set of Linea plugins for the sequencer and RPC nodes.

## Quickstart - Running [Linea Besu](https://github.com/ConsenSys/linea-besu) with plugins

- compile linea-plugins `gradlew installDist`
- copy jar file to besu runtime plugins/ directory (where you will run besu from, not where you're building besu)
- add `LINEA` to besu config to enable the plugin RPC methods
  - rpc-http-api=\["ADMIN","ETH","NET","WEB3","LINEA"\]
- start besu (command line or from IDE) and you should see plugins registered at startup
- call the RPC endpoint eg:

```shell
  curl --location --request POST 'http://localhost:8545' --data-raw '{
    "jsonrpc": "2.0",
    "method": "linea_estimateGas",
    "params": [
      "from": "0x73b2e0E54510239E22cC936F0b4a6dE1acf0AbdE",
      "to": "0xBb977B2EE8a111D788B3477D242078d0B837E72b",
      "value": "0x123"
    ],
    "id": 1
  }'
```

## Development Setup

### Install Java 21

### Native Lib Prerequisites

Linux/MacOs
* Install the relevant CGo compiler for your platform
* Install the Go toolchain

Windows
* Requirement [Docker Desktop WSL 2 backend on Windows](https://docs.docker.com/desktop/wsl/)

On release native libs are built for all the supported platforms,
if you want to test this process locally run `./gradlew -PreleaseNativeLibs jar`,
jar is generated in `sequencer/build/libs`.

### Run tests

```shell
# Run all tests
./gradlew clean test

# Run only acceptance tests
./gradlew clean acceptanceTests
```

## IntelliJ IDEA Setup

### Enable Annotation Processing

- Go to `Settings | Build, Execution, Deployment | Compiler | Annotation Processors` and tick the following
  checkbox:

  ![idea_enable_annotation_processing_setting.png](images/idea_enable_annotation_processing_setting.png)

______________________________________________________________________

NOTE

> This setting is required to avoid IDE compilation errors because of the [Lombok](https://projectlombok.org/features/)
> library used for code generation of boilerplate Java code such as:
>
> - Getters/Setters (via [`@Getter/@Setter`](https://projectlombok.org/features/GetterSetter))
> - Class log instances (via [`@Slf4j`](https://projectlombok.org/features/log))
> - Builder classes (via [`@Builder`](https://projectlombok.org/features/Builder))
> - Constructors (
>   via [`@NoArgsConstructor/@RequiredArgsConstructor/@AllArgsConstructor`](https://projectlombok.org/features/constructor))
> - etc.
>
> Learn more about how Java annotation processing
> works [here](https://www.baeldung.com/java-annotation-processing-builder).

______________________________________________________________________

### Install Optional Plugins

- Install [Spotless Gradle](https://plugins.jetbrains.com/plugin/18321-spotless-gradle) plugin to re-format through
  the IDE according to spotless configuration.

## Plugins

Plugins are documented [here](PLUGINS.md).

## Release Process
Here are the steps for releasing a new version of the plugins:
  1. Create a tag with the release version number in the format vX.Y.Z (e.g., v0.2.0 creates a release version 0.2.0).
  2. Push the tag to the repository.
  3. GitHub Actions will automatically create a draft release for the release tag.
  4. Once the release workflow completes, update the release notes, uncheck "Draft", and publish the release.

Note: Release tags (of the form v*) are protected and can only be pushed by organization and/or repository owners.
