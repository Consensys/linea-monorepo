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

# Contract Deployment for Dev and Production

Truffle is still used to deploy contracts to prod/dev blockchain and generate "data/".

```bash
make deploy-local
```

# Useful scripts

Most of the scripts need to know address of Ethereum RPC. This is controlled via `BLOCKCHAIN_NODE` environment variable,

for example run this before all the scripts in the same terminal:

```bash
export BLOCKCHAIN_NODE=http://localhost:5000
```

## Print balances of all test addresses:

```bash
make print-all-test-balances
```

## Register operator in smart contract

```bash
ts-node scripts/registerOperator.ts data/rollup.json ../node-data/test/keys/contract_owner.acc ../node-data/test/keys/operator_1.acc
```

or

```bash
make register-operator-in-smc-local
```

## Send inbound transfer to smart contract

in ETH:

Params are: <path-to-rollup.json> <path-to-credentials-file> <account-id-in-rollup> <amount>
please note, that amount is in ETH, not in wei.

```bash
export BLOCKCHAIN_NODE="http://localhost:8545"
ts-node scripts/registerInboundEth.ts \
    data/rollup.json \
    ../node-data/test/keys/eth_account_3.acc \
    0 \
    100.3333
```

ERC 20
Params are: <path-to-rollup.json> <path-to-credentials-file> <account-id-in-rollup> <token-id> <amount>

amount is in the smallest units

```bash
export BLOCKCHAIN_NODE="http://localhost:8545"
ts-node scripts/registerInboundErc20.ts \
data/rollup.json \
../node-data/test/keys/eth_account_3.acc \
0 \
0 \
10000000
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
