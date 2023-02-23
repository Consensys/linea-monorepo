# besu-tracing-plugin

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

Run the Besu node 
```
$BESU_HOME/bin/besu 
```

## Build Instructions

### Install Prerequisites

* Java 17
