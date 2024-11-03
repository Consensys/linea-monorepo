Readme.md

# Smart Contract

Contains Ethereum smart contract code for ConsenSys Rollups.

# Development & Testing

This project uses following libraries

- [Ethers](https://github.com/ethers-io/ethers.js/) as Ethereum library
- [Hardhat](https://hardhat.org/getting-started/) as development environment
- [Chai](https://www.chaijs.com/) for assertions

To run the tests:

```bash
make test
```

# Useful scripts

Most of the scripts need to know address of Ethereum RPC. This is controlled via `BLOCKCHAIN_NODE` environment variable,

for example run this before all the scripts in the same terminal:

```bash
export BLOCKCHAIN_NODE=http://localhost:5000
```

## Balance of ERC 20 token

```
export BLOCKCHAIN_NODE="http://localhost:8545"
ts-node scripts/balanceOf.ts \
data/rollup.json \
../node-data/test/keys/eth_account_3.acc \
1 \
../node-data/test/keys/eth_account_3.acc \
../node-data/test/keys/eth_account_4.acc \
../node-data/test/keys/eth_account_5.acc
```

### To deploy ZkEvm to local docker compose

Run in ./contracts with running docker-compose stack.

```shell
sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8445/" .env.template > .env
npx hardhat run ./scripts/deployment/deployZkEVM.ts --network zkevm_dev
```

### To deploy ZkEvm to local docker compose

Run in ./contracts with running docker-compose stack.

```shell
sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8445/" .env.template > .env
npx hardhat run ./scripts/deployment/deployZkEVM.ts --network zkevm_dev
```

## Linea Token Bridge

Token Bridge is a canonical brige between Ethereum and Linea networks.

Documentation: [./docs/linea-token-bridge.md](./docs/linea-token-bridge.md)

## Foundry Setup & Testing

This project supports Foundry, in addition to Hardhat.

To install Foundry, follow the [Foundry installation guide](https://book.getfoundry.sh/getting-started/installation.html).

Make sure the submodules are initialized:

```bash
git submodule update --init --recursive
```

### Running Foundry Tests

Foundry tests have been added to a specific directory (`contracts/test/foundry`) to coexist with Hardhat tests.

To compile and run Foundry tests:

```bash
forge build
forge test --match-path "test/foundry/*.t.sol" -vvv
```

This command builds and runs Foundry tests, ensuring compatibility with the project's bytecode and enabling further test expansion.