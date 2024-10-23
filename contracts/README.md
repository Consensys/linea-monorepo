Readme.md

# Smart Contracts

Contains Ethereum smart contract code for Consensys Rollups.

## LineaRollup (L1MessageService)
The Linea Rollup, which contains the L1MessageService, is the smart contract that is responsible for:

- Submitting messages to be sent to Linea (L2) for later claiming.
- Anchoring of L2 message Merkle roots to allow later claiming.
- Claiming of messages sent from L2 to Ethereum mainnet (L1).
- Submission of L2 compressed data using EIP-4844 blobs.
- Finalization of L2 state on L1 using a Zero Knowledge Proof.

## L2MessageService
The L2MessageService is the L2 smart contract that is responsible for:

- Submitting messages to be sent to L1 for later claiming.
- Claiming of messages sent from L1 to L2.
- Anchoring of L1 to L2 Message hashes.

## Linea Canonical Token Bridge

The Canonical Token Bridge (TokenBridge) is a canonical ERC20 token brige between Ethereum and Linea networks.

The TokenBridge utilises the L1MessageService and the L2MessageService for the transmission of messages between each layer's TokenBridge.

Documentation: [./docs/linea-token-bridge.md](./docs/linea-token-bridge.md)

# Development & Testing

This project uses following libraries
- [PNPM](https://pnpm.io/) as the Package Manager
- [Ethers](https://github.com/ethers-io/ethers.js/) as Ethereum library
- [Hardhat](https://hardhat.org/getting-started/) as development environment
- [Chai](https://www.chaijs.com/) for assertions
- [GoLang](https://go.dev/) for the compilation of code to autogenerate data for L2 data and proofs (not strictly required)
- [Docker](https://www.docker.com/) for the local stack to run in

To run the tests:

## Testing without coverage

```bash
cd contracts # from the root folder
pnpm install
npx hardhat test
```

## Testing with coverage

```bash
cd contracts # from the root folder
pnpm install

npx hardhat coverage
```
## Deploying the contracts to the local stack
Prerequisites: 
- Be sure Docker is running.

Some caveats:
- The L2 chain will not produce empty blocks if there are no transactions, so it would be useful to execute a script to keep the chain "moving". [PlaceHolder]()
- For blob submission and finalization, there needs to be sufficient blocks to trigger it. Keeping the chain moving on L2 is vital for this to take place.

From the root of the repository:

Please read the MakeFile: [MakeFile](../Makefile)

```
# This will deploy all the relevant services and smart contracts

make fresh-start-all
```



### To deploy ZkEvm to local docker compose

Run in ./contracts with running docker-compose stack.

```shell
sed "s/BLOCKCHAIN_NODE=.*/BLOCKCHAIN_NODE=http:\/\/localhost:8445/" .env.template > .env
npx hardhat run ./scripts/deployment/deployZkEVM.ts --network zkevm_dev
```

