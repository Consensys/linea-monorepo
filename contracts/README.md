# Smart Contracts

Contains Ethereum smart contract code for the Linea Rollup and Message Service.

## LineaRollup (L1MessageService)
The Linea Rollup, which contains the L1MessageService, is the smart contract that is responsible for:

- Submitting messages to be sent to Linea (L2) for later claiming.
- Anchoring of L2 message Merkle roots to allow later claiming.
- Claiming of messages sent from L2 to Ethereum mainnet (L1).
- Submission of L2 compressed data using EIP-4844 blobs or via calldata.
- Finalization of L2 state on L1 using a Zero Knowledge Proof.

## L2MessageService
The L2MessageService is the L2 smart contract that is responsible for:

- Submitting messages to be sent to L1 for later claiming.
- Anchoring of L1 to L2 Message hashes for later claiming.
- Claiming of messages sent from L1 to L2.

## Linea Canonical Token Bridge

The Canonical Token Bridge (TokenBridge) is a canonical ERC20 token brige between Ethereum and Linea networks.

The TokenBridge utilises the L1MessageService and the L2MessageService for the transmission of messages between each layer's TokenBridge.

Documentation: [Token Bridge](./docs/linea-token-bridge.md)

# Style Guide
Please see the [Smart Contract Style Guide](./docs/contract-style-guide.md) for in depth smart contract layout and styling.

# Audit reports
Please see [Audits](./docs/audits.md) for a historical list of all the smart contract audits.

# Development & Testing

Please see [Testing guidelines](./test/README.md) for in depth testing layout and styling.

This project uses following libraries
- [PNPM](https://pnpm.io/) as the Package Manager
- [Ethers](https://github.com/ethers-io/ethers.js/) as Ethereum library
- [Hardhat](https://hardhat.org/getting-started/) as development environment
- [Chai](https://www.chaijs.com/) for assertions
- [GoLang](https://go.dev/) for the compilation of code to autogenerate data for L2 data and proofs (not strictly required)
- [Docker](https://www.docker.com/) for the local stack to run in

If you already have an understanding of the tech stack, use our [Get Started](../docs/get-started.md) guide.

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

make fresh-start-all-traces-v2
```

### To deploy all the contracts

If the stack is *already running*, to redeploy the contracts, the following commands can be used:

Note: The addresses change per deployment due to nonce increments, so be sure to validate the correct ones are being used.

**NB:** The end to end tests run against a fresh stack and deploy with predetermined addresses.

**Deploying the L1 contracts**
```
# This will deploy the Linea Rollup that is currently deployed on Mainnet - the current version is the LineaRollupV5.
# Some end to end tests will test future upgrades to validate the stack remains functional.

# Note: By default a test/placeholder verifier contract is used `IntegrationTestTrueVerifier` if you wish to use a proper verifier, adjust the
# PLONKVERIFIER_NAME=IntegrationTestTrueVerifier in the make command to be something like PLONKVERIFIER_NAME=PlonkVerifierForDataAggregation .

# Be sure to check the parameter values in the Makefile before executing the command.

make deploy-linea-rollup

make deploy-token-bridge-l1
```

**Deploying the L2 contracts**
```
# This will deploy the current L2 Message Service.
# Some end to end tests will test future upgrades to validate the stack remains functional.

make deploy-l2messageservice

make deploy-token-bridge-l2
```

**Deploying L1 and L2 together**
```
make deploy-contracts

# This will trigger the following:
# Note: the deploy-l1-test-erc20 and deploy-l1-test-erc20 commands are executed for use in the end to end tests.

cd contracts/; \
	export L1_NONCE=$$(npx ts-node local-deployments-artifacts/get-wallet-nonce.ts --wallet-priv-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 --rpc-url http://localhost:8445) && \
	export L2_NONCE=$$(npx ts-node local-deployments-artifacts/get-wallet-nonce.ts --wallet-priv-key 0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae --rpc-url http://localhost:8545) && \
	cd .. && \
	$(MAKE) -j6 deploy-linea-rollup-v5 deploy-token-bridge-l1 deploy-l1-test-erc20 deploy-l2messageservice deploy-token-bridge-l2 deploy-l2-test-erc20
```