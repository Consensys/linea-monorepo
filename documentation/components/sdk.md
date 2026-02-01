# SDK

> TypeScript SDK for bridging and cross-chain messaging between Ethereum and Linea.

## Overview

The Linea SDK enables developers to:
- Bridge ETH and ERC20 tokens between L1 and L2
- Send cross-chain messages
- Track message and bridge transaction status
- Claim messages on destination chains

## Package Structure

```
sdk/
├── sdk-core/           # @consensys/linea-sdk-core
│   │                   # Shared types, utilities, constants
│   └── src/
│       ├── types/
│       ├── utils/
│       └── constants/
│
├── sdk-viem/           # @consensys/linea-sdk-viem
│   │                   # Viem-based implementation
│   └── src/
│       ├── actions/
│       └── decorators/
│
└── sdk-ethers/         # @consensys/linea-sdk
                        # Ethers.js v6 implementation
```

## Architecture

```
┌────────────────────────────────────────────────────────────────────────┐
│                            SDK ARCHITECTURE                            │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       User Application                           │  │
│  │                                                                  │  │
│  │  import { LineaSDK } from "@consensys/linea-sdk"                 │  │
│  │  // OR                                                           │  │
│  │  import { publicActionsL1 } from "@consensys/linea-sdk-viem"     │  │
│  │                                                                  │  │
│  └──────────────────────────────┬───────────────────────────────────┘  │
│                                 │                                      │
│                 ┌───────────────┴───────────────┐                      │
│                 │                               │                      │
│                 ▼                               ▼                      │
│  ┌────────────────────────────┐  ┌────────────────────────────┐        │
│  │       sdk-ethers           │  │        sdk-viem            │        │
│  │                            │  │                            │        │
│  │  ┌──────────────────────┐  │  │  ┌──────────────────────┐  │        │
│  │  │     LineaSDK         │  │  │  │   Wallet Actions     │  │        │
│  │  │                      │  │  │  │                      │  │        │
│  │  │ - getL1Contract()    │  │  │  │ - deposit()          │  │        │
│  │  │ - getL2Contract()    │  │  │  │ - withdraw()         │  │        │
│  │  └──────────────────────┘  │  │  │ - claimOnL1()        │  │        │
│  │                            │  │  │ - claimOnL2()        │  │        │
│  │  ┌──────────────────────┐  │  │  └──────────────────────┘  │        │
│  │  │ LineaRollupClient    │  │  │                            │        │
│  │  │ L2MessageService     │  │  │  ┌──────────────────────┐  │        │
│  │  │    Client            │  │  │  │   Public Actions     │  │        │
│  │  └──────────────────────┘  │  │  │                      │  │        │
│  │                            │  │  │ - getMessageStatus() │  │        │
│  └────────────────────────────┘  │  │ - getMessageProof()  │  │        │
│                                  │  │ - getBlockExtraData()│  │        │
│                 │                │  └──────────────────────┘  │        │
│                 │                │                            │        │
│                 │                └────────────────────────────┘        │
│                 │                               │                      │
│                 └───────────────┬───────────────┘                      │
│                                 │                                      │
│                                 ▼                                      │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                          sdk-core                                │  │
│  │                                                                  │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐   │  │
│  │  │   Types     │  │ SparseMerkle│  │   Contract Addresses    │   │  │
│  │  │             │  │    Tree     │  │                         │   │  │
│  │  │ - Message   │  │             │  │ - Mainnet               │   │  │
│  │  │ - Proof     │  │ - Generate  │  │ - Sepolia               │   │  │
│  │  │ - Status    │  │   proofs    │  │                         │   │  │
│  │  └─────────────┘  └─────────────┘  └─────────────────────────┘   │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

## Installation

```bash
# Viem-based (recommended)
npm install @consensys/linea-sdk-viem

# Ethers.js v6 based
npm install @consensys/linea-sdk

# Core only (types/utilities)
npm install @consensys/linea-sdk-core
```

## Usage: sdk-viem

### Setup

```typescript
import { createPublicClient, createWalletClient, http } from 'viem';
import { mainnet, linea } from 'viem/chains';
import {
  publicActionsL1,
  publicActionsL2,
  walletActionsL1,
  walletActionsL2,
} from '@consensys/linea-sdk-viem';

// L1 clients
const l1PublicClient = createPublicClient({
  chain: mainnet,
  transport: http(),
}).extend(publicActionsL1());

const l1WalletClient = createWalletClient({
  chain: mainnet,
  transport: http(),
  account: privateKeyToAccount('0x...'),
}).extend(walletActionsL1());

// L2 clients
const l2PublicClient = createPublicClient({
  chain: linea,
  transport: http(),
}).extend(publicActionsL2());

const l2WalletClient = createWalletClient({
  chain: linea,
  transport: http(),
  account: privateKeyToAccount('0x...'),
}).extend(walletActionsL2());
```

### Bridge ETH L1 → L2

```typescript
// Deposit ETH to L2
const hash = await l1WalletClient.deposit({
  to: recipientAddress,
  value: parseEther('0.1'),
  fee: parseEther('0.001'),
});

// Wait for L2 availability
const status = await l1PublicClient.getL1ToL2MessageStatus({
  transactionHash: hash,
});
// status: 'UNKNOWN' | 'CLAIMABLE' | 'CLAIMED'

// Claim on L2 (if needed)
if (status === 'CLAIMABLE') {
  await l2WalletClient.claimOnL2({
    transactionHash: hash,
  });
}
```

### Bridge ETH L2 → L1

```typescript
// Withdraw ETH to L1
const hash = await l2WalletClient.withdraw({
  to: recipientAddress,
  value: parseEther('0.1'),
  fee: parseEther('0.001'),
});

// Wait for L1 finalization
const status = await l2PublicClient.getL2ToL1MessageStatus({
  transactionHash: hash,
});

// Get proof and claim on L1
if (status === 'CLAIMABLE') {
  const proof = await l2PublicClient.getMessageProof({
    transactionHash: hash,
  });
  
  await l1WalletClient.claimOnL1({
    transactionHash: hash,
    proof,
  });
}
```

### Query Message Status

```typescript
// Get L1→L2 message status
const l1ToL2Status = await l1PublicClient.getL1ToL2MessageStatus({
  transactionHash: '0x...',
});

// Get L2→L1 message status
const l2ToL1Status = await l2PublicClient.getL2ToL1MessageStatus({
  transactionHash: '0x...',
});

// Get message by hash
const message = await l1PublicClient.getMessageByMessageHash({
  messageHash: '0x...',
});

// Get messages from transaction
const messages = await l1PublicClient.getMessagesByTransactionHash({
  transactionHash: '0x...',
});
```

## Usage: sdk-ethers

### Setup

```typescript
import { LineaSDK, LineaSDKOptions } from '@consensys/linea-sdk';
import { JsonRpcProvider, Wallet } from 'ethers';

const options: LineaSDKOptions = {
  l1: {
    rpcUrl: 'https://mainnet.infura.io/v3/...',
    contractAddress: '0xd19d4B5d358258f05D7B411E21A1460D11B0876F',
  },
  l2: {
    rpcUrl: 'https://rpc.linea.build',
    contractAddress: '0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec',
  },
  mode: 'read-write',
  l1Signer: new Wallet(privateKey, l1Provider),
  l2Signer: new Wallet(privateKey, l2Provider),
};

const sdk = new LineaSDK(options);
```

### Bridge Operations

```typescript
// Get contract clients
const l1Contract = sdk.getL1Contract();
const l2Contract = sdk.getL2Contract();

// Send message L1 → L2
const tx = await l1Contract.sendMessage({
  to: recipientAddress,
  fee: parseEther('0.001'),
  calldata: '0x',
  value: parseEther('0.1'),
});

// Claim on L2
const message = await l2Contract.getMessageByTxHash(tx.hash);
if (message.status === 'CLAIMABLE') {
  await l2Contract.claimMessage(message);
}
```

## Core Types

```typescript
// Message structure
interface Message {
  from: Address;
  to: Address;
  fee: bigint;
  value: bigint;
  nonce: bigint;
  calldata: Hex;
  messageHash: Hex;
}

// Extended message with metadata
interface ExtendedMessage extends Message {
  blockNumber: bigint;
  transactionHash: Hex;
  status: OnChainMessageStatus;
}

// Message status
enum OnChainMessageStatus {
  UNKNOWN = 0,
  CLAIMABLE = 1,
  CLAIMED = 2,
}

// Message proof for L2→L1 claims
interface MessageProof {
  proof: Hex[];
  root: Hex;
  leafIndex: number;
}
```

## Contract Addresses

### Mainnet

```typescript
const mainnetAddresses = {
  l1: {
    lineaRollup: '0xd19d4B5d358258f05D7B411E21A1460D11B0876F',
    tokenBridge: '0x051F1D88f0aF5763fB888eC78378F1109b52Cd01',
  },
  l2: {
    messageService: '0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec',
    tokenBridge: '0x353012dc4a9A6cF55c941bADC267f82004A8ceB9',
  },
};
```

### Sepolia (Testnet)

```typescript
const sepoliaAddresses = {
  l1: {
    lineaRollup: '0xb218f8a4bc926cf1ca7b3423c154a0d627bdb7e5',
    tokenBridge: '0x5506A3805fB6C857D16e3ce28e8D13fCB12F6433',
  },
  l2: {
    messageService: '0x971e727e956690b9957be6d51ec16e73acac83a7',
    tokenBridge: '0x3E2Ea8BfA28b1e8EE8cf00B7b1A6B38DAb6c3ECe',
  },
};
```

## Bridging Flow Diagrams

### L1 → L2 Deposit

```
┌────────────────────────────────────────────────────────────────────────┐
│                       L1 → L2 DEPOSIT FLOW                             │
│                                                                        │
│  User                L1                   Coordinator        L2        │
│   │                   │                        │             │         │
│   │  deposit()        │                        │             │         │
│   │──────────────────▶│                        │             │         │
│   │                   │                        │             │         │
│   │                   │  MessageSent event     │             │         │
│   │                   │───────────────────────▶│             │         │
│   │                   │                        │             │         │
│   │                   │                        │ anchor hash │         │
│   │                   │                        │────────────▶│         │
│   │                   │                        │             │         │
│   │  getL1ToL2MessageStatus() ─────────────────────────────▶│          │
│   │◀─── CLAIMABLE ──────────────────────────────────────────│          │
│   │                   │                        │             │         │
│   │  claimOnL2()      │                        │             │         │
│   │─────────────────────────────────────────────────────────▶│         │
│   │                   │                        │             │         │
│   │◀─── ETH received ───────────────────────────────────────│          │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

### L2 → L1 Withdrawal

```
┌────────────────────────────────────────────────────────────────────────┐
│                     L2 → L1 WITHDRAWAL FLOW                            │
│                                                                        │
│  User                L2                   Coordinator        L1        │
│   │                   │                        │             │         │
│   │  withdraw()       │                        │             │         │
│   │──────────────────▶│                        │             │         │
│   │                   │                        │             │         │
│   │                   │  MessageSent event     │             │         │
│   │                   │  Rolling hash update   │             │         │
│   │                   │───────────────────────▶│             │         │
│   │                   │                        │             │         │
│   │                   │            finalization with proof   │         │
│   │                   │                        │────────────▶│         │
│   │                   │                        │             │         │
│   │  getL2ToL1MessageStatus() ◀─────────────────────────────────────── │
│   │◀─── CLAIMABLE                                                      │
│   │                   │                        │             │         │
│   │  getMessageProof()                                                 │
│   │──────────────────▶│                        │             │         │
│   │◀─── proof ────────│                        │             │         │
│   │                   │                        │             │         │
│   │  claimOnL1(proof)                                       │          │
│   │─────────────────────────────────────────────────────────▶│         │
│   │                   │                        │             │         │
│   │◀─── ETH received ───────────────────────────────────────│          │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

## Building

```bash
cd sdk

# Install dependencies
pnpm install

# Build all packages
pnpm run build

# Run tests
pnpm run test

# Lint
pnpm run lint
```

## Dependencies

- **sdk-core**: No external dependencies (pure TypeScript)
- **sdk-viem**: viem >= 2.22.0
- **sdk-ethers**: ethers >= 6.0.0
