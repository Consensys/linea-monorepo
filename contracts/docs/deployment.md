# Linea Deployment Scripts
<br />

## Environment Variables Naming Convention

Environment variables follow a consistent naming pattern:

| Prefix | Usage |
|--------|-------|
| `L1_` | Ethereum L1 (e.g. `L1_SECURITY_COUNCIL`, `L1_RPC_URL`, `L1_SAFE_ADDRESS`) |
| `L2_` | Linea L2 (e.g. `L2_SECURITY_COUNCIL`, `L2_RPC_URL`, `L2_MESSAGE_SERVICE_ADDRESS`) |
| `LINEA_ROLLUP_*` | Linea Rollup contract (L1) — product-specific |

**Shared per layer:**

- `L1_SECURITY_COUNCIL` — shared across all L1 contracts (Linea Rollup, Validium, Token Bridge L1, Rollup Revenue Vault, RecoverFunds, Yield Manager)
- `L2_SECURITY_COUNCIL` — shared across all L2 contracts (L2 Message Service, Token Bridge L2)

**Shared L1 base (Linea Rollup & Validium):** `INITIAL_L2_STATE_ROOT_HASH`, `INITIAL_L2_BLOCK_NUMBER`, `L2_GENESIS_TIMESTAMP` — common to both products via shared contract base

**Product-specific:** `LINEA_ROLLUP_OPERATORS`, `LINEA_ROLLUP_RATE_LIMIT_*`, `VALIDIUM_OPERATORS`, `VALIDIUM_RATE_LIMIT_*`

**RPC endpoints:** `L1_RPC_URL`, `L2_RPC_URL`, `CUSTOM_RPC_URL` (replaces legacy `BLOCKCHAIN_NODE`, `L2_BLOCKCHAIN_NODE`, `CUSTOM_BLOCKCHAIN_URL`)

<br />

This document aims to explain how to get started with deploying the Linea deployment scripts. There are several ways the scripts can be executed dependent on: 
- If you're storing deployment variables in an environment file (.env)
- If you plan to deploy an individual script which will deploy a single contract.
- If you plan to deploy a chained deployment script that will include multiple contracts.

<br />
Running the script with an .env file set, you will need to make sure that the correct variables are set in the .env file, considering the network that you're deploying on. In this way when the script is being run, it will take the variables it needs to execute the script from that .env file. <br />
<br />

Running the script without an .env file will require you to place the variables as command-line arguments.
The command-line arguments will create or replace existing .env (only in memory) environment variables. If the variables are provided in the terminal as command-line arguments, they will have priority over the same variables if they are defined in the .env file. These need not exist in the .env file.

Furthermore, you can also specify a general set of variables in the .env file (VERIFY_CONTRACT, SEPOLIA_PRIVATE_KEY, LINEA_SEPOLIA_PRIVATE_KEY, MAINNET_PRIVATE_KEY, LINEA_MAINNET_PRIVATE_KEY, ETHERSCAN_API_KEY, INFURA_API_KEY) and provide only the script-specific variables as command-line arguments, when you run each script.

Setting `VERIFY_CONTRACT=true` will start the verifying stage after the contract is deployed, provided that there is a `ETHERSCAN_API_KEY` available in the .env or provided as CLI argument.

<br />

## Network Specific Variables

Dependent on which network you are using, a specific network private key needs to be used, as well as the corresponding API Key or RPC URL.  Also, dependent on which network you choose, the block explorer used could be different, so the block explorer parameter name might need to be adjusted.  The following table highlights which private key variable will be used per network. Please use the variable that pertains to the network. e.g. for `linea_sepolia` use `LINEA_SEPOLIA_PRIVATE_KEY` (`LINEA_SEPOLIA_PRIVATE_KEY=<key> INFURA_API_KEY=<key>`)  

| Network       | Private key parameter name   | API Key / RPC URL | Block explorer parameter name |
| ------------- | ----------------- | ---- | ----------------- | 
| sepolia    | SEPOLIA_PRIVATE_KEY    | INFURA_API_KEY  | ETHERSCAN_API_KEY |
| linea_sepolia | LINEA_SEPOLIA_PRIVATE_KEY   | INFURA_API_KEY  | ETHERSCAN_API_KEY |
| mainnet   | MAINNET_PRIVATE_KEY | INFURA_API_KEY | ETHERSCAN_API_KEY |
| linea_mainnet | LINEA_MAINNET_PRIVATE_KEY |  INFURA_API_KEY  | ETHERSCAN_API_KEY |
| custom    | CUSTOM_PRIVATE_KEY | CUSTOM_RPC_URL | ETHERSCAN_API_KEY |
| zkevm_dev | PRIVATE_KEY | L1_RPC_URL or L2_RPC_URL | n/a |

<br />

## Generalized Command Format

```shell
<possible CLI environment arguments> npx hardhat deploy --network sepolia --tags <contract tags, comma delimitted list>
```

<br />
<br />

## Order of Precedence

 When deploying, if required variables such as deployed contract addresses are not defined in the .env or provided as CLI arguments, the script will look and check if it can use the addresses stored in the deployments/<network_name>/ folder. 
 <br />
 The order of priority (unless specified otherwise) will be:
 - CLI arguments, 
 - .env variables ,
 - deployments/<network_name>/

## Deployments and their parameters
### Verifier
<br />

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name             | Required | Input Value | Description |
| -------------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| VERIFIER_CONTRACT_NAME | true  | string | The name of the PlonkVerifier contract that should be deployed |
| VERIFIER_PROOF_TYPE | true  | string | The proof type that the verifier should be mapped to |
| VERIFIER_CHAIN_ID | true  | uint256 | Chain ID passed to the verifier constructor |
| VERIFIER_BASE_FEE | true  | uint256 | Base fee passed to the verifier constructor |
| VERIFIER_COINBASE | true  | address | Coinbase address passed to the verifier constructor |
| L2_MESSAGE_SERVICE_ADDRESS | true  | address | L2 Message Service address passed to the verifier constructor |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags PlonkVerifier
```

Base command with cli arguments:

```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> VERIFIER_CONTRACT_NAME=PlonkVerifierDev npx hardhat deploy --network sepolia --tags PlonkVerifier
```

(make sure to replace `<key>` with actual values)
<br />
<br />

### EIP2935SystemContract
<br />

The EIP2935SystemContract is a system contract for historical block hashes according to [EIP-2935](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2935.md). This contract deploys to a predetermined address using a specific deployment transaction format. The deployment script automatically funds the required sender address if needed.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name             | Required | Input Value | Description |
| -------------------------- | -------- | ---------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| L1_RPC_URL     | true     | key | RPC URL endpoint` |

<br />

**Prerequisites:**
- The deployment script automatically funds the predetermined sender address (0x3462413Af4609098e1E27A490f554f260213D685) if needed
- The contract deploys to the fixed address: 0x0000F90827F1C53a10cb7A02335B175320002935

Base command:
```shell
npx hardhat deploy --network sepolia --tags EIP2935SystemContract
```

Base command with cli arguments:

```shell
PRIVATE_KEY=<key> L1_RPC_URL=<node_rpc_url> npx hardhat deploy --network sepolia --tags EIP2935SystemContract
```

(make sure to replace `<key>` with actual values)
<br />
<br />

### UpgradeableWithdrawalQueuePredeploy
<br />

The UpgradeableWithdrawalQueuePredeploy is an upgradeable predeploy contract that implements EIP-7002 execution layer triggerable withdrawals. This deploys as a placeholder implementation that can be upgraded later to provide full functionality.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name             | Required | Input Value | Description |
| -------------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags UpgradeableWithdrawalQueuePredeploy
```

Base command with cli arguments:

```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> npx hardhat deploy --network sepolia --tags UpgradeableWithdrawalQueuePredeploy
```

(make sure to replace `<key>` with actual values)
<br />
<br />

### UpgradeableConsolidationQueuePredeploy
<br />

The UpgradeableConsolidationQueuePredeploy is an upgradeable predeploy contract that implements EIP-7251 execution layer consolidation requests. This deploys as a placeholder implementation that can be upgraded later to provide full functionality.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name             | Required | Input Value | Description |
| -------------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags UpgradeableConsolidationQueuePredeploy
```

Base command with cli arguments:

```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> npx hardhat deploy --network sepolia --tags UpgradeableConsolidationQueuePredeploy
```

(make sure to replace `<key>` with actual values)
<br />
<br />

### LineaRollup
<br />

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network.|
| INITIAL_L2_STATE_ROOT_HASH   | true      | bytes | Initial State Root Hash (shared base) |
| INITIAL_L2_BLOCK_NUMBER   | true      | uint256 | Initial L2 Block Number (shared base) |
| L2_GENESIS_TIMESTAMP | true | uint256 | Genesis timestamp (shared base) |
| L1_SECURITY_COUNCIL  | true      | address | L1 Security Council Address |
| LINEA_ROLLUP_OPERATORS     | true      | address | L1 Operators Addresses (comma-delimited if multiple) |
| LINEA_ROLLUP_RATE_LIMIT_PERIOD     | true  | uint256   | L1 Rate Limit Period |
| LINEA_ROLLUP_RATE_LIMIT_AMOUNT     | true  | uint256   | L1 Rate Limit Amount |
| PLONKVERIFIER_ADDRESS | true | address | PlonkVerifier contract address (set automatically when deploying Verifier in same chain) |
| YIELD_MANAGER_ADDRESS | true | address | Yield Manager contract address |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags LineaRollup
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> INITIAL_L2_STATE_ROOT_HASH=<bytes> INITIAL_L2_BLOCK_NUMBER=<value> L2_GENESIS_TIMESTAMP=<value> L1_SECURITY_COUNCIL=<address> LINEA_ROLLUP_OPERATORS=<address> LINEA_ROLLUP_RATE_LIMIT_PERIOD=<value> LINEA_ROLLUP_RATE_LIMIT_AMOUNT=<value> YIELD_MANAGER_ADDRESS=<address> npx hardhat deploy --network sepolia --tags LineaRollup
```

(make sure to replace `<value>` `<key>` `<bytes>` `<address>` with actual values).

<br />
<br />

### Timelock
<br />


Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input Value | Description |
| ------------------ | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| TIMELOCK_PROPOSERS | true     | address | Timelock Proposers address |
| TIMELOCK_EXECUTORS | true     | address | Timelock Executors address |
| TIMELOCK_ADMIN_ADDRESS | true     | address | Timelock Admin address |
| MIN_DELAY | true      | uint256 | Timelock Minimum Delay |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags Timelock
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> TIMELOCK_PROPOSERS=<address> TIMELOCK_EXECUTORS=<address> TIMELOCK_ADMIN_ADDRESS=<address> MIN_DELAY=<value> npx hardhat deploy --network sepolia --tags Timelock
```

(make sure to replace `<value>` `<key>` `<address>` with actual values)

<br />
<br />

### L2MessageService
<br />

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input Value | Description |
| ------------------ | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| L2_SECURITY_COUNCIL | true   | address | L2 Security council address |
| L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER  | true  |  address | L1L2 Message Setter address on L2 |
| L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD    | true  |  uint256 | L2 Rate Limit Period |
| L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT    | true  |  uint256 | L2 Rate Limit Amount |

<br />

Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags L2MessageService
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true LINEA_SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> L2_SECURITY_COUNCIL=<address> L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER=<address> L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD=<value> L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT=<value> npx hardhat deploy --network linea_sepolia --tags L2MessageService
```

(make sure to replace `<value>` `<key>` `<address>` with actual values)

<br />
<br />

### BridgedToken
<br />

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input Value | Description |
| --------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT       | false    |true\|false| Verifies the deployed contract. |
| \**PRIVATE_KEY*       | true     | key | Network-specific private key used when deploying the contract. |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY         | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |

<br />

Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags BridgedToken
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true ETHERSCAN_API_KEY=<key> LINEA_SEPOLIA_PRIVATE_KEY=<key> INFURA_API_KEY=<key> npx hardhat deploy --network linea_sepolia --tags BridgedToken
```

(make sure to replace `<value>` `<key>` `<address>` with actual values)

<br />
<br />

### CustomBridgedToken
<br />

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input Value | Description |
| --------------------- | -------- | ---------- | ----------- |
| CUSTOMTOKENBRIDGE_NAME | true    |string| Token's name |
| CUSTOMTOKENBRIDGE_SYMBOL | true    |string| Token's symbol |
| CUSTOMTOKENBRIDGE_DECIMALS | true    |uint256| Token's decimals |
| CUSTOMTOKENBRIDGE_BRIDGE_ADDRESS | true    |address| Token bridge's address|
| VERIFY_CONTRACT       | false    |true\|false| Verifies the deployed contract. |
| \**PRIVATE_KEY*       | true     | key | Network-specific private key used when deploying the contract. |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY         | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |

<br />

Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags CustomBridgedToken
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true ETHERSCAN_API_KEY=<key> LINEA_SEPOLIA_PRIVATE_KEY=<key> INFURA_API_KEY=<key> CUSTOMTOKENBRIDGE_NAME=<name> CUSTOMTOKENBRIDGE_SYMBOL=<symbol> CUSTOMTOKENBRIDGE_DECIMALS=<decimals> CUSTOMTOKENBRIDGE_BRIDGE_ADDRESS=<address> npx hardhat deploy --network linea_sepolia --tags CustomBridgedToken
```

(make sure to replace `<key>` `<address>` `<name>` `<symbol>` `<decimals>` with actual values)

<br />
<br />

### TokenBridge
<br />

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input Value | Description |
| --------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT       | false    |true\|false| Verifies the deployed contract. |
| \**PRIVATE_KEY*       | true     | key | Network-specific private key used when deploying the contract. |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY         | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| L2_MESSAGE_SERVICE_ADDRESS    | true  | address   | L2 Message Service address used when deploying TokenBridge.    |
| LINEA_ROLLUP_ADDRESS         | true    | address       | L1 Rollup address used when deploying Token Bridge.   |
| REMOTE_CHAIN_ID       | true      |   uint256     | ChainID of the remote (target) network |
| REMOTE_SENDER_ADDRESS | true | address | Remote sender address (the TokenBridge on the other chain) |
| BRIDGED_TOKEN_ADDRESS | true | address | BridgedToken beacon address (deploy BridgedToken first) |
| TOKEN_BRIDGE_L1       | false     |true\|false| If Token Bridge is deployed on L1, TOKEN_BRIDGE_L1 should be set to `true`. Otherwise it should be `false`|
| L1_SECURITY_COUNCIL | conditional | address | L1 Security Council address. Required when `TOKEN_BRIDGE_L1=true` |
| L2_SECURITY_COUNCIL | conditional | address | L2 Security Council address. Required when `TOKEN_BRIDGE_L1=false` |
| L1_RESERVED_TOKEN_ADDRESSES | false   | address   | If TOKEN_BRIDGE_L1=true, then L1_RESERVED_TOKEN_ADDRESSES should be defined. If multiple addresses, provide them in a comma-delimited array.|
| L2_RESERVED_TOKEN_ADDRESSES | false   | address   | If TOKEN_BRIDGE_L1=false, then L2_RESERVED_TOKEN_ADDRESSES should be defined. If multiple addresses, provide them in a comma-delimited array.|

<br />

Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags TokenBridge
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true ETHERSCAN_API_KEY=<key> LINEA_SEPOLIA_PRIVATE_KEY=<key> INFURA_API_KEY=<key> REMOTE_CHAIN_ID=<uint256> TOKEN_BRIDGE_L1=true L1_SECURITY_COUNCIL=<address> L1_RESERVED_TOKEN_ADDRESSES=<address> L2_MESSAGE_SERVICE_ADDRESS=<address> LINEA_ROLLUP_ADDRESS=<address> REMOTE_SENDER_ADDRESS=<address> BRIDGED_TOKEN_ADDRESS=<address> npx hardhat deploy --network linea_sepolia --tags TokenBridge
```

(make sure to replace `<value>` `<key>` `<address>` with actual values)

<br />
<br />

### Validium
<br />

The Validium contract is a permutation of LineaRollup that uses off-chain data availability. It shares the same base contract (same initial state root, block number, and genesis timestamp) but has its own operators and rate limits.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network.|
| PLONKVERIFIER_ADDRESS | true | address | PlonkVerifier contract address |
| INITIAL_L2_STATE_ROOT_HASH   | true      | bytes | Initial State Root Hash (shared base) |
| INITIAL_L2_BLOCK_NUMBER   | true      | uint256 | Initial L2 Block Number (shared base) |
| L2_GENESIS_TIMESTAMP | true | uint256 | Genesis timestamp (shared base) |
| L1_SECURITY_COUNCIL  | true      | address | L1 Security Council Address |
| VALIDIUM_OPERATORS     | true      | address | Validium Operators Addresses (comma-delimited if multiple) |
| VALIDIUM_RATE_LIMIT_PERIOD     | true  | uint256   | Validium Rate Limit Period |
| VALIDIUM_RATE_LIMIT_AMOUNT     | true  | uint256   | Validium Rate Limit Amount |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags Validium
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> INITIAL_L2_STATE_ROOT_HASH=<bytes> INITIAL_L2_BLOCK_NUMBER=<value> L2_GENESIS_TIMESTAMP=<value> L1_SECURITY_COUNCIL=<address> VALIDIUM_OPERATORS=<address> VALIDIUM_RATE_LIMIT_PERIOD=<value> VALIDIUM_RATE_LIMIT_AMOUNT=<value> npx hardhat deploy --network sepolia --tags Validium
```

(make sure to replace `<value>` `<key>` `<bytes>` `<address>` with actual values).

<br />
<br />

### RecoverFunds
<br />

Deploys the RecoverFunds contract as an upgradeable proxy.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network.|
| L1_SECURITY_COUNCIL  | true      | address | L1 Security Council Address |
| RECOVERFUNDS_EXECUTOR_ADDRESS | true | address | Executor address for the RecoverFunds contract |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags RecoverFunds
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> L1_SECURITY_COUNCIL=<address> RECOVERFUNDS_EXECUTOR_ADDRESS=<address> npx hardhat deploy --network sepolia --tags RecoverFunds
```

(make sure to replace `<key>` `<address>` with actual values)

<br />
<br />

### CallForwardingProxy
<br />

Deploys a CallForwardingProxy that forwards calls to the LineaRollup contract.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| LINEA_ROLLUP_ADDRESS | true | address | LineaRollup contract address to forward calls to |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags CallForwardingProxy
```

<br />
<br />

### LineaSequencerUptimeFeed
<br />

Deploys the Linea Sequencer Uptime Feed contract (Chainlink-compatible).

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| LINEA_SEQUENCER_UPTIME_FEED_INITIAL_STATUS | true | uint256 | Initial feed status |
| LINEA_SEQUENCER_UPTIME_FEED_ADMIN | true | address | Admin address |
| LINEA_SEQUENCER_UPTIME_FEED_UPDATER | true | address | Updater address |

<br />

Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags LineaSequencerUptimeFeed
```

<br />
<br />

### EIP4788SystemContract
<br />

The EIP4788SystemContract is a system contract for the beacon block root according to [EIP-4788](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-4788.md). Like EIP2935, this contract deploys to a predetermined address using a specific deployment transaction format.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name             | Required | Input Value | Description |
| -------------------------- | -------- | ---------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| L1_RPC_URL     | true     | key | RPC URL endpoint |

<br />

**Prerequisites:**
- The deployment script automatically funds the predetermined sender address if needed

Base command:
```shell
npx hardhat deploy --network sepolia --tags EIP4788SystemContract
```

Base command with cli arguments:

```shell
PRIVATE_KEY=<key> L1_RPC_URL=<node_rpc_url> npx hardhat deploy --network sepolia --tags EIP4788SystemContract
```

(make sure to replace `<key>` with actual values)
<br />
<br />

### UpgradeableBeaconChainDepositPredeploy
<br />

The UpgradeableBeaconChainDepositPredeploy is an upgradeable predeploy contract that implements EIP-6110 execution layer deposits. This deploys as a placeholder implementation that can be upgraded later.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name             | Required | Input Value | Description |
| -------------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags UpgradeableBeaconChainDepositPredeploy
```

Base command with cli arguments:

```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> npx hardhat deploy --network sepolia --tags UpgradeableBeaconChainDepositPredeploy
```

(make sure to replace `<key>` with actual values)
<br />
<br />

### RollupRevenueVault
<br />

Deploys the RollupRevenueVault contract as an upgradeable proxy.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| L1_SECURITY_COUNCIL  | true      | address | L1 Security Council Address |
| ROLLUP_REVENUE_VAULT_LAST_INVOICE_DATE | true | uint256 | Last invoice date timestamp |
| ROLLUP_REVENUE_VAULT_INVOICE_SUBMITTER | true | address | Invoice submitter address |
| ROLLUP_REVENUE_VAULT_BURNER | true | address | Burner address |
| ROLLUP_REVENUE_VAULT_INVOICE_PAYMENT_RECEIVER | true | address | Invoice payment receiver address |
| ROLLUP_REVENUE_VAULT_TOKEN_BRIDGE | true | address | Token bridge address |
| L2_MESSAGE_SERVICE_ADDRESS | true | address | L2 Message Service address |
| ROLLUP_REVENUE_VAULT_L1_LINEA_TOKEN_BURNER | true | address | L1 Linea token burner address |
| ROLLUP_REVENUE_VAULT_LINEA_TOKEN | true | address | Linea token address |
| ROLLUP_REVENUE_VAULT_DEX_SWAP_ADAPTER | true | address | DEX swap adapter address |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags RollupRevenueVault
```

(make sure to replace `<key>` `<address>` `<value>` with actual values)
<br />
<br />

### L1LineaTokenBurner
<br />

Deploys the L1 Linea Token Burner contract.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| LINEA_ROLLUP_ADDRESS | true | address | LineaRollup address (L1 message service) |
| LINEA_TOKEN_BURNER_LINEA_TOKEN | true | address | Linea token address |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags L1LineaTokenBurner
```

<br />
<br />

### V3DexSwapAdapter
<br />

Deploys the V3 DEX Swap Adapter contract for WETH deposits.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| V3_DEX_SWAP_ADAPTER_ROUTER | true | address | Uniswap V3 router address |
| V3_DEX_SWAP_ADAPTER_WETH_TOKEN | true | address | WETH token address |
| V3_DEX_SWAP_ADAPTER_LINEA_TOKEN | true | address | Linea token address |
| V3_DEX_SWAP_ADAPTER_POOL_TICK_SPACING | true | int24 | Pool tick spacing |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags V3DexSwapAdapter
```

<br />
<br />

### YieldManager
<br />

Deploys YieldManager, ValidatorContainerProofVerifier and LidoStVaultYieldProviderFactory contracts.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| LINEA_ROLLUP_ADDRESS | true | address | LineaRollup contract address |
| L1_SECURITY_COUNCIL  | true      | address | L1 Security Council Address |
| NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS | true | address | Automation service address |
| VAULT_HUB | true | address | Lido Vault Hub address |
| VAULT_FACTORY | true | address | Lido Vault Factory address |
| STETH | true | address | stETH token address |
| MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS | true | uint256 | Minimum withdrawal reserve percentage in basis points |
| TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS | true | uint256 | Target withdrawal reserve percentage in basis points |
| MINIMUM_WITHDRAWAL_RESERVE_AMOUNT | true | uint256 | Minimum withdrawal reserve amount |
| TARGET_WITHDRAWAL_RESERVE_AMOUNT | true | uint256 | Target withdrawal reserve amount |
| GI_FIRST_VALIDATOR | false | bytes32 | Generalized index for first validator |
| GI_PENDING_PARTIAL_WITHDRAWALS_ROOT | false | bytes32 | Generalized index for pending partial withdrawals root |
| VALIDATOR_CONTAINER_PROOF_VERIFIER_ADMIN | false | address | Admin address for the ValidatorContainerProofVerifier |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags YieldManager
```

<br />
<br />

### PaymentSplitterWrapper
<br />

Deploys a PaymentSplitterWrapper contract for splitting payments among multiple payees.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| VERIFY_CONTRACT    | false    | true\|false | Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| PAYMENT_SPLITTER_PAYEES | true | address | Comma-separated list of payee addresses |
| PAYMENT_SPLITTER_SHARES | true | uint256 | Comma-separated list of shares per payee |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags PaymentSplitterWrapper
```

<br />
<br />

### TestERC20
<br />

Deploys a test ERC20 token (for testing purposes only).

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| TEST_ERC20_NAME | true | string | Token name |
| TEST_ERC20_SYMBOL | true | string | Token symbol |
| TEST_ERC20_INITIAL_SUPPLY | true | uint256 | Initial token supply |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags TestERC20
```

<br />
<br />

### ImplementationForProxy
<br />

Deploys a new implementation contract for an existing upgradeable proxy. Outputs encoded calldata for upgrading the proxy via the Security Council Safe.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| CONTRACT_NAME | true | string | Name of the contract to deploy as new implementation |
| PROXY_ADDRESS | true | address | Address of the existing proxy contract |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags ImplementationForProxy
```

Base command with cli arguments:
```shell
SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> CONTRACT_NAME=<string> PROXY_ADDRESS=<address> npx hardhat deploy --network sepolia --tags ImplementationForProxy
```

(make sure to replace `<key>` `<address>` `<string>` with actual values)

<br />
<br />

### L2MessageServiceLineaMainnet
<br />

Deploys L2MessageService using the V1-deployed ABI and bytecode (for Linea mainnet upgrade compatibility). Uses the same env vars as L2MessageService.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| L2_SECURITY_COUNCIL | true | address | L2 Security council address |
| L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER | true | address | L1L2 Message Setter address on L2 |
| L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD | true | uint256 | L2 Rate Limit Period |
| L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT | true | uint256 | L2 Rate Limit Amount |

<br />

Base command:
```shell
npx hardhat deploy --network linea_mainnet --tags L2MessageServiceLineaMainnet
```

<br />
<br />

### YieldManagerArtifacts
<br />

Deploys YieldManager, ValidatorContainerProofVerifier and LidoStVaultYieldProviderFactory from audited artifacts. Use this instead of `YieldManager` when deploying from pre-built, audit-signed bytecode.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| LINEA_ROLLUP_ADDRESS | true | address | LineaRollup contract address |
| L1_SECURITY_COUNCIL  | true      | address | L1 Security Council Address |
| NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS | true | address | Automation service address |
| VAULT_HUB | true | address | Lido Vault Hub address |
| VAULT_FACTORY | true | address | Lido Vault Factory address |
| STETH | true | address | stETH token address |
| MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS | true | uint256 | Minimum withdrawal reserve percentage in basis points |
| TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS | true | uint256 | Target withdrawal reserve percentage in basis points |
| MINIMUM_WITHDRAWAL_RESERVE_AMOUNT | true | uint256 | Minimum withdrawal reserve amount |
| TARGET_WITHDRAWAL_RESERVE_AMOUNT | true | uint256 | Target withdrawal reserve amount |
| GI_FIRST_VALIDATOR | false | bytes32 | Generalized index for first validator |
| GI_PENDING_PARTIAL_WITHDRAWALS_ROOT | false | bytes32 | Generalized index for pending partial withdrawals root |
| VALIDATOR_CONTAINER_PROOF_VERIFIER_ADMIN | false | address | Admin address for the ValidatorContainerProofVerifier |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags YieldManagerArtifacts
```

<br />
<br />

### YieldManagerImplementation
<br />

Deploys a new YieldManager implementation contract (without proxy). Use this when upgrading an existing YieldManager proxy.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input value | Description |
| --------------------- | -------- | -------------- | ----------- |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. |
| LINEA_ROLLUP_ADDRESS | true | address | LineaRollup contract address (constructor argument) |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags YieldManagerImplementation
```

<br />
<br />

## Upgrade Deployments
<br />

These scripts deploy new implementation contracts and output encoded calldata for upgrading existing proxy contracts via the Security Council Safe. They do **not** execute the upgrade — the calldata must be submitted through the Safe.

<br />

### LineaRollupWithReinitialization

Deploys a new LineaRollup implementation and generates encoded upgrade calldata with `reinitializeV8`.

| Parameter name | Required | Input value | Description |
|---|---|---|---|
| \**PRIVATE_KEY* | true | key | Network-specific private key |
| L1_SECURITY_COUNCIL | true | address | Security Council address |
| LINEA_ROLLUP_ADDRESS | true | address | Existing LineaRollup proxy address |

```shell
npx hardhat deploy --network sepolia --tags LineaRollupWithReinitialization
```

<br />

### LineaRollupV7WithReinitialization

Deploys LineaRollupV7 from audited artifacts and generates encoded upgrade calldata with `reinitializeLineaRollupV7`.

| Parameter name | Required | Input value | Description |
|---|---|---|---|
| \**PRIVATE_KEY* | true | key | Network-specific private key |
| L1_SECURITY_COUNCIL | true | address | Security Council address |
| LINEA_ROLLUP_ADDRESS | true | address | Existing LineaRollup proxy address |
| YIELD_MANAGER_ADDRESS | true | address | Yield Manager contract address |
| NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS | true | address | Automation service address |

```shell
npx hardhat deploy --network sepolia --tags LineaRollupV7WithReinitialization
```

<br />

### L2MessageServiceWithReinitialization

Deploys a new L2MessageService implementation and generates encoded upgrade calldata with `reinitializeV3`.

| Parameter name | Required | Input value | Description |
|---|---|---|---|
| \**PRIVATE_KEY* | true | key | Network-specific private key |
| L2_MESSAGE_SERVICE_ADDRESS | true | address | Existing L2MessageService proxy address |

```shell
npx hardhat deploy --network linea_sepolia --tags L2MessageServiceWithReinitialization
```

<br />

### TokenBridgeWithReinitialization

Deploys a new TokenBridge implementation and generates encoded upgrade calldata with `reinitializeV2`.

| Parameter name | Required | Input value | Description |
|---|---|---|---|
| \**PRIVATE_KEY* | true | key | Network-specific private key |
| TOKEN_BRIDGE_ADDRESS | true | address | Existing TokenBridge proxy address |

```shell
npx hardhat deploy --network sepolia --tags TokenBridgeWithReinitialization
```

<br />

### RollupRevenueVaultWithReinitialization

Deploys a new RollupRevenueVault implementation and generates encoded upgrade calldata with `initializeRolesAndStorageVariables`.

| Parameter name | Required | Input value | Description |
|---|---|---|---|
| \**PRIVATE_KEY* | true | key | Network-specific private key |
| ROLLUP_REVENUE_VAULT_ADDRESS | true | address | Existing RollupRevenueVault proxy address |
| ROLLUP_REVENUE_VAULT_LAST_INVOICE_DATE | true | uint256 | Last invoice date timestamp |
| L1_SECURITY_COUNCIL | true | address | L1 Security Council Address |
| ROLLUP_REVENUE_VAULT_INVOICE_SUBMITTER | true | address | Invoice submitter address |
| ROLLUP_REVENUE_VAULT_BURNER | true | address | Burner address |
| ROLLUP_REVENUE_VAULT_INVOICE_PAYMENT_RECEIVER | true | address | Invoice payment receiver address |
| ROLLUP_REVENUE_VAULT_TOKEN_BRIDGE | true | address | Token bridge address |
| L2_MESSAGE_SERVICE_ADDRESS | true | address | L2 Message Service address |
| ROLLUP_REVENUE_VAULT_L1_LINEA_TOKEN_BURNER | true | address | L1 Linea token burner address |
| ROLLUP_REVENUE_VAULT_LINEA_TOKEN | true | address | Linea token address |
| ROLLUP_REVENUE_VAULT_DEX_SWAP_ADAPTER | true | address | DEX swap adapter address |

```shell
npx hardhat deploy --network sepolia --tags RollupRevenueVaultWithReinitialization
```

<br />
<br />

## Chained Deployments
<br />

This section describes the scripts that can be run to deploy multiple contracts in a sequence.

<br />


### L1MessageService Chained Deployments
<br />

This will run the script that deploys PlonkVerifier, LineaRollup , Timelock contracts.

Parameters that should be filled either in .env or passed as CLI arguments:

| Parameter name        | Required | Input Value | Description |
| ------------------ | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| INITIAL_L2_STATE_ROOT_HASH   | true      | bytes | Initial State Root Hash (shared base) |
| INITIAL_L2_BLOCK_NUMBER   | true      | uint256 | Initial L2 Block Number (shared base) |
| L2_GENESIS_TIMESTAMP | true | uint256 | Genesis timestamp (shared base) |
| L1_SECURITY_COUNCIL  | true      | address | Security Council Address |
| LINEA_ROLLUP_OPERATORS     | true      | address | Operators Addresses (comma-delimited if multiple) |
| LINEA_ROLLUP_RATE_LIMIT_PERIOD     | true  | uint256   | L1 Rate Limit Period |
| LINEA_ROLLUP_RATE_LIMIT_AMOUNT     | true  | uint256   | L1 Rate Limit Amount |
| TIMELOCK_PROPOSERS | true     | address | Timelock Proposers address |
| TIMELOCK_EXECUTORS | true     | address | Timelock Executors address |
| TIMELOCK_ADMIN_ADDRESS | true     | address | Timelock Admin address |
| MIN_DELAY | true      | uint256 | Timelock Minimum Delay |
| VERIFIER_CONTRACT_NAME | true | string | PlonkVerifier contract name that should be deployed |
| VERIFIER_PROOF_TYPE | true | string | The proof type that the verifier should be mapped to |
| VERIFIER_CHAIN_ID | true | uint256 | Chain ID passed to the verifier constructor |
| VERIFIER_BASE_FEE | true | uint256 | Base fee passed to the verifier constructor |
| VERIFIER_COINBASE | true | address | Coinbase address passed to the verifier constructor |
| L2_MESSAGE_SERVICE_ADDRESS | true | address | L2 Message Service address passed to the verifier constructor |
| YIELD_MANAGER_ADDRESS | true | address | Yield Manager contract address |

<br />

Base command:
```shell
npx hardhat deploy --network sepolia --tags PlonkVerifier,LineaRollup,Timelock
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> INITIAL_L2_STATE_ROOT_HASH=<bytes> INITIAL_L2_BLOCK_NUMBER=<value> L2_GENESIS_TIMESTAMP=<value> L1_SECURITY_COUNCIL=<address> LINEA_ROLLUP_OPERATORS=<address> LINEA_ROLLUP_RATE_LIMIT_PERIOD=<value> LINEA_ROLLUP_RATE_LIMIT_AMOUNT=<value> YIELD_MANAGER_ADDRESS=<address> TIMELOCK_PROPOSERS=<address> TIMELOCK_EXECUTORS=<address> TIMELOCK_ADMIN_ADDRESS=<address> MIN_DELAY=<value> VERIFIER_CONTRACT_NAME=PlonkVerifierForMultiTypeDataAggregation npx hardhat deploy --network sepolia --tags PlonkVerifier,LineaRollup,Timelock
```

(make sure to replace `<value>` `<bytes>` `<key>` `<address>` with actual values)

<br />
<br />

### L2MessageService Chained Deployments
<br />

This will run the script that deploys Timelock, L2MessageService contracts.

| Parameter name        | Required | Input Value | Description |
| ------------------ | -------- | ---------- | ----------- |
| VERIFY_CONTRACT    | false    |true\|false| Verifies the deployed contract |
| \**PRIVATE_KEY* | true     | key | Network-specific private key used when deploying the contract |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY     | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| L2_SECURITY_COUNCIL | true   | address | L2 Security council address |
| L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER  | true  |  address | L1L2 Message Setter address on L2 |
| L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD    | true  |  uint256 | L2 Rate Limit Period |
| L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT    | true  |  uint256 | L2 Rate Limit Amount |
| TIMELOCK_PROPOSERS | true     | address | Timelock Proposers address |
| TIMELOCK_EXECUTORS | true     | address | Timelock Executors address |
| TIMELOCK_ADMIN_ADDRESS | true     | address | Timelock Admin address |
| MIN_DELAY | true      | uint256 | Timelock Minimum Delay |

<br />

Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags L2MessageService,Timelock
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true LINEA_SEPOLIA_PRIVATE_KEY=<key> ETHERSCAN_API_KEY=<key> INFURA_API_KEY=<key> L2_SECURITY_COUNCIL=<address> L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER=<address> L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD=<value> L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT=<value> TIMELOCK_PROPOSERS=<address> TIMELOCK_EXECUTORS=<address> TIMELOCK_ADMIN_ADDRESS=<address> MIN_DELAY=<value> npx hardhat deploy --network linea_sepolia --tags L2MessageService,Timelock
```

(make sure to replace `<value>` `<key>` `<address>` with actual values)

<br />
<br />

### TokenBridge & BridgedToken Chained Deployments

This will run the script that deploys the TokenBridge and BridgedToken contracts.

| Parameter name        | Required | Input Value | Description |
| --------------------- | -------- | ---------- | ----------- |
| VERIFY_CONTRACT       | false    |true\|false| Verifies the deployed contract. |
| \**PRIVATE_KEY*       | true     | key | Network-specific private key used when deploying the contract. |
| \**BLOCK_EXPLORER_API_KEY*  | false     | key | Network-specific Block Explorer API Key used for verifying deployed contracts. |
| INFURA_API_KEY         | true     | key | Infura API Key. This is required only when deploying contracts to a live network, not required when deploying on a local dev network. |
| L2_MESSAGE_SERVICE_ADDRESS    | true  | address   | L2 Message Service address used when deploying TokenBridge.    |
| LINEA_ROLLUP_ADDRESS         | true    | address       | L1 Rollup address used when deploying Token Bridge.   |
| REMOTE_CHAIN_ID       | true      |   uint256     | ChainID of the remote (target) network |
| REMOTE_SENDER_ADDRESS | true | address | Remote sender address (the TokenBridge on the other chain) |
| TOKEN_BRIDGE_L1       | false     |true\|false| If Token Bridge is deployed on L1, TOKEN_BRIDGE_L1 should be set to `true`. Otherwise it should be `false`|
| L1_SECURITY_COUNCIL | conditional | address | L1 Security Council address. Required when `TOKEN_BRIDGE_L1=true` |
| L2_SECURITY_COUNCIL | conditional | address | L2 Security Council address. Required when `TOKEN_BRIDGE_L1=false` |
| L1_RESERVED_TOKEN_ADDRESSES | false   | address   | If TOKEN_BRIDGE_L1=true, then L1_RESERVED_TOKEN_ADDRESSES should be defined. If multiple addresses, provide them in a comma-delimited array.|
| L2_RESERVED_TOKEN_ADDRESSES | false   | address   | If TOKEN_BRIDGE_L1=false, then L2_RESERVED_TOKEN_ADDRESSES should be defined. If multiple addresses, provide them in a comma-delimited array.|


Base command:
```shell
npx hardhat deploy --network linea_sepolia --tags BridgedToken,TokenBridge
```

Base command with cli arguments:
```shell
VERIFY_CONTRACT=true ETHERSCAN_API_KEY=<key> LINEA_SEPOLIA_PRIVATE_KEY=<key> INFURA_API_KEY=<key> REMOTE_CHAIN_ID=<uint256> TOKEN_BRIDGE_L1=true L1_SECURITY_COUNCIL=<address> L1_RESERVED_TOKEN_ADDRESSES=<address> L2_MESSAGE_SERVICE_ADDRESS=<address> LINEA_ROLLUP_ADDRESS=<address> REMOTE_SENDER_ADDRESS=<address> npx hardhat deploy --network linea_sepolia --tags BridgedToken,TokenBridge
```
(make sure to replace `<value>` `<key>` `<address>` with actual values)

<br />
<br />