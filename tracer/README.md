# Besu zkBesu-tracer Plugin

A zk-evm tracing implementation for [Hyperledger Besu](https://github.com/hyperledger/besu) based on
an [existing implementation in Go](https://github.com/ConsenSys/zk-evm/).

## Development Setup

**Step 1.** Install Java 17:

```
brew install openjdk@17
```

**Step 2.** Install Go:

```
brew install go
```

**Step 3.** Install Rust:

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Use local git executable to fetch from repos (needed for private repos)
echo "net.git-fetch-with-cli=true" >> .cargo/config.toml
```

**Step 4.** Install Corset:

```shell
cargo install --git ssh://git@github.com/ConsenSys/corset
```

**Step 5.** Clone [zk-geth](https://github.com/Consensys/zk-geth) & compile `zkevm.bin`:

```shell
git clone git@github.com:ConsenSys/zk-geth.git --recursive

cd zk-geth/zk-evm
make zkevm.bin
```

**Step 6.** Set environment with path to `zkevm.bin`:

```shell
export ZK_EVM_BIN=PATH_TO_ZK_GETH/zk-evm/zkevm.bin
```

**Step 7.** Install [pre-commit](https://pre-commit.com/):

```shell
pip install pre-commit

# For macOS users.
brew install pre-commit
```

Then run `pre-commit install` to set up git hook scripts.
Used hooks can be found [here](.pre-commit-config.yaml).

______________________________________________________________________

NOTE

> `pre-commit` aids in running checks (end of file fixing,
> markdown linting, linting, runs ests, json validation, etc.)
> before you perform your git commits.

______________________________________________________________________

**Step 8.** Run tests

```shell
# Run all tests
./gradlew clean test

# Run only unit tests
./gradlew clean unitTests

# Run only acceptance tests
./gradlew clean corsetTests
```

## IntelliJ IDEA Setup

**Step 1.** Enable annotation processing setting:

- Go to `Settings | Build, Execution, Deployment | Compiler | Annotation Processors` and tick the following
  checkbox:

![idea_enable_annotation_processing_setting.png](images%2Fidea_enable_annotation_processing_setting.png)

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

**Step 2.** Install [Checkstyle](https://plugins.jetbrains.com/plugin/1065-checkstyle-idea) plugin and set IDE code
reformatting to comply with the project's Checkstyle configuration:

- Go to `Settings | Editor | Code Style | Java | <hamburger menu> | Import Scheme | Checkstyle configuration`:

  ![idea_checkstyle_reformat.png](images%2Fidea_checkstyle_reformat.png)

  and select `<project_root>/config/checkstyle.xml`.

**Step 3.** OPTIONAL: Install [Spotless Gradle](https://plugins.jetbrains.com/plugin/18321-spotless-gradle) plugins
for code linting capabilities
within the IDE.

## Debugging Traces

- JSON files can be debugged with the following command:

```shell
corset check -T JSON_FILE -v $ZK_EVM_BIN
```
