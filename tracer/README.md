# Besu Tracing Plugin

Tracing plugin for Besu

* Hyperledger Besu Docs https://besu.hyperledger.org
* Hyperledger Besu repo https://github.com/hyperledger/besu/

### Services Used
- **RpcEndpointService**
    * To add a custom RPC endpoint

### Plugin Lifecycle
- **Register**
    * Add the RPC method
- **Start**
    * Connect to the Besu events
- **Stop**
    * Disconnect from the Besu events

## Build Instructions

### Install Prerequisites

* Java 17

## To Execute the Demo

Build the plugin jar
```
./gradlew build
```

Install the plugin into `$BESU_HOME`

```
mkdir $BESU_HOME/plugins
cp build/libs/*.jar $BESU_HOME/plugins
```

Enable the additional RPC API group (unless using an existing one)
eg if namespace = "tests", add "TESTS" to the rpc-http-api group:

`rpc-http-api=["ADMIN","ETH","NET","WEB3","PERM","DEBUG","MINER","EEA","PRIV","TXPOOL","TRACE","TESTS"]`

Run the Besu node 
```
$BESU_HOME/bin/besu 
```

Test with curl commands eg

`curl -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"tests_setValue","params":["bob"],"id":53}' http://localhost:8545`

`curl -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"tests_getValue","params":[],"id":53}' http://localhost:8545`

