# Linea Besu Distribution

This project uses Gradle to manage dependencies, build tasks, and create distributions for linea-besu with 
all the necessary plugins to run a node for operators. The build process will also create a Docker image that can be 
used to run a node with a specific profile.

## Run with Docker

### Step 1. Download configuration files

You can start with the Docker Compose files located in the [docker-compose](https://github.com/Consensys/linea-besu-package/tree/main/docker) directory.

### Step 2. Update the Docker Compose file
In the docker-compose.yaml file, update the --p2p-host command to include your public IP address. For example:
```sh
--p2p-host=103.10.10.10
```

Update plugin-linea-l1-rpc-endpoint config with your L1 RPC endpoint.
You should replace YOUR_L1_RPC_ENDPOINT with your endpoint, like:
```sh
--plugin-linea-l1-rpc-endpoint=https://mainnet.infura.io/v3/PROJECT_ID
```

This is required to enable RPC queries using "FINALIZED" block tag.
Linea Finalization status is based on L1 RPC endpoint's response

### Step 2. Start the Besu node
```sh
docker compose -f ./docker/docker-compose-basic-mainnet.yaml up
```
Alternatively, to run a node with a specific profile, set the `BESU_PROFILE` environment variable to the desired profile name:

```sh
docker run -e BESU_PROFILE=basic-mainnet consensys/linea-besu-package:mainnet-latest
```
Or use the sepolia image
```sh
docker run -e BESU_PROFILE=basic-mainnet consensys/linea-besu-package:sepolia-latest
```
## Run with a binary distribution

### Step 1. Install Linea Besu from packaged binaries
*  Download the [linea-besu-package&lt;release&gt;.tar.gz](https://github.com/Consensys/linea-besu-package/releases) 
from the assets.
* Unpack the downloaded files and change into the besu-linea-package-&lt;release&gt;
directory.

Display Besu command line help to confirm installation:
```sh
bin/besu --help
```

### Step 2. Start the Besu client

Note the YOUR_L1_RPC_ENDPOINT. You should replace this with your L1 RPC endpoint. 

This is required to enable RPC queries using "FINALIZED" tag. 
Linea Finalization status is based on L1 RPC endpoint's response
```sh
bin/besu --profile=advanced-mainnet --plugin-linea-l1-rpc-endpoint=YOUR_L1_RPC_ENDPOINT
```

## Build from source

1. Make a branch with changes to `versions/linea-*.env` as needed
2. Go to the [actions tab](https://github.com/Consensys/linea-besu-package/actions) and click on the appropriate workflow and select your branch - docker image will get published

## How-To Release

Releases are automated using GitHub Actions and are triggered by pushing a tag that matches the
pattern `'v[0-9]+.[0-9]+.[0-9]+`. (e.g., `v1.0.0`, `v2.1.3`)

The tag creation will create a release and include the distribution artifact uploaded as an asset.
   ```sh
   git tag v0.0.1
   git push upstream v0.0.1
   ```

Additionally, the `latest` tag will be updated to match this release.


## Profiles

This project leverages [Besu Profiles](https://besu.hyperledger.org/public-networks/how-to/use-configuration-file/profile) to enable multiple startup configurations for different node types.

During the build process, all TOML files located in the [linea-besu/profiles](https://github.com/Consensys/linea-besu-package/tree/main/linea-besu/profiles) directory will be incorporated into the package. These profiles are crucial for configuring the node, as each one specifies the necessary plugins and CLI options to ensure Besu operates correctly.

Each profile is a TOML file that outlines the plugins and CLI options to be used when starting the node. For example:

```toml
# required plugins to run a sequencer node
plugins=["LineaExtraDataPlugin","LineaEndpointServicePlugin","LineaTransactionPoolValidatorPlugin","LineaTransactionSelectorPlugin"]

# required options to configure the plugins above
plugin-linea-module-limit-file-path="config/trace-limits.mainnet.toml"
# Other required plugin options
# ...

# Other Besu options
# ...
```

Currently, the following profiles are available:

| Profile Name                                                                                                              | Description                                                               | Network | 
|---------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|---------|
| [`basic-mainnet`](https://github.com/Consensys/linea-besu-package/blob/main/linea-besu/profiles/basic-mainnet.toml)       | Creates a basic Linea Node.                                               | Mainnet |
| [`advanced-mainnet`](https://github.com/Consensys/linea-besu-package/blob/main/linea-besu/profiles/advanced-mainnet.toml) | Creates a Linea Node with `linea_estimateGas` and `finalized tag plugin`. | Mainnet |
| [`basic-sepolia`](https://github.com/Consensys/linea-besu-package/blob/main/linea-besu/profiles/basic-sepolia.toml)       | Creates a basic Linea Node.                                               | Sepolia |
| [`advanced-sepolia`](https://github.com/Consensys/linea-besu-package/blob/main/linea-besu/profiles/advanced-mainnet.toml) | Creates a Linea Node with `linea_estimateGas` and `finalized tag plugin`. | Sepolia |
