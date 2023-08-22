# Linea Token Bridge

## Documentation

Token Bridge is a canonical brige between Ethereum and Linea networks.

## Install

### Packages

To install packages, execute:

```shell
npm i
```

### Config

To setup config, copy the `.env.template` to `.env`, for example:

```shell
cp .env.template .env
```

Edit `.env` and add your configuration values.

| Var                         | Description                 | Default                                    |
| --------------------------- | --------------------------- | ------------------------------------------ |
| L1_RESERVED_TOKEN_ADDRESSES | Reserved L1 token addresses | 0x07865c6E87B9F70255377e024ace6630C1Eaa37F |
| L2_RESERVED_TOKEN_ADDRESSES | Reserved L2 token addresses | 0xf56dc6695cF1f5c364eDEbC7Dc7077ac9B586068 |
| ETHERSCAN_API_KEY           | Etherscan API key           |                                            |

## Deploy

### On a Local Testnet network with mocked messaging service

In a first terminal, run:

```shell
npx hardhat node
```

In a second terminal, run:

```shell
npx hardhat run --network localhost scripts/test/tokenBridge/deployMock.ts
```

### On a Goerli Testnet network with mocked messaging service

In a terminal, run:

```shell
npx hardhat run --network goerli scripts/test/tokenBridge/deployMock.ts
```

### On Goerli Testnet and Goerli Linea Testnet

In a terminal, run:

```shell
make deploy-testnet-token-bridge
```

Or

```shell
npx hardhat run --network zkevm_dev scripts/tokenBridge/deploy-1.ts
npx hardhat run --network l2 scripts/tokenBridge/deploy-1.ts
npx hardhat run --network zkevm_dev scripts/tokenBridge/deploy-2.ts
npx hardhat run --network l2 scripts/tokenBridge/deploy-2.ts
```

All addresses created will be stored in the deployments.json file at the root of this project

## Development

### Testing

To run tests, execute:

```shell
npm run test
```

or

```shell
npx hardhat test
```

To run tests on only one file, execute:

```shell
npx hardhat test test/tokenBridge/TokenBridge.ts
```

### Test coverage

This project uses the Hardhat plugin [solidity-coverage](https://github.com/sc-forks/solidity-coverage/blob/master/HARDHAT_README.md) to assess the overall coverage of the unit tests.
To generate a boilerplate report, use the following command:

```shell
npm run coverage
```

or

```shell
npx hardhat coverage --solcoverjs ./.solcover.js
```

The report will be generated in the `coverage` folder at the root of the repository. To visualize it in your web browser, you can use the `coverage/index.html` file.
Note: the second command line might not work if the folder `coverage` already exists. If you encounter an issue, please delete the whole `coverage` folder and let hardhat-coverage regenerate a new one.

### Contract verification on Etherscan

To verify the contract on Etherscan.

```shell
 npx hardhat verify --network NETWORK DEPLOYED_CONTRACT_ADDRESS "Constructor argument 1" "Constructor argument 2"
```

### Gas Estimation

You can estimate the contracts gas costs.

- On a terminal start a local node:

```shell
npx hardhat node
```

- On another terminal, execute the gas estimation script:

```shell
npx hardhat run --network localhost scripts/tokenBridge/gasEstimation/gasEstimation.ts
```

It should return gas estimation:

```shell
basic bridgeToken:                            162080
bridgeToken with permit:                      243910
bridgeToken after confirmDeploy:              126453
bridgeToken with permit after confirmDeploy:  202909
```

## Formatting Commands

### Lint Solidity

```bash
npm run lint:sol
```

### Lint TypeScript

```bash
npm run lint:ts
```

### Prettier

Check format code:

```bash
npm run prettier:check
```

Format code:

```bash
npm run prettier
```
