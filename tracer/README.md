# Besu zkBesu-tracer Plugin

A zk-evm tracing implementation for [Hyperledger Besu](https://github.com/hyperledger/besu) based on an [existing implementation in Go](https://github.com/ConsenSys/zk-evm/).

## Prerequisites

* Java 17
```
brew install openjdk@17
```
* Install Go
```
brew install go
```
* Install Rust
```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Use local git executable to fetch from repos (needed for private repos)
echo "net.git-fetch-with-cli=true" >> .cargo/config.toml
```
* Install Corset
```
```cargo install --git ssh://git@github.com/ConsenSys/corset

* Clone zk-geth & compile zkevm.bin
```
git clone git@github.com:ConsenSys/zk-geth.git --recursive

cd zk-geth/zk-evm
make zkevm.bin 
```
* Set environment with path to zkevm.bin
```
export ZK_EVM_BIN=PATH_TO_ZK_GETH/zk-evm/zkevm.bin
```

## Run tests
```
./gradlew test
```

## Debugging traces

JSON files can be debugged with the following command:
```
corset check -T JSON_FILE -v $ZK_EVM_BIN
```