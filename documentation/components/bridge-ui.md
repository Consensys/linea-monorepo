# Bridge UI

> Next.js web application for bridging assets between Ethereum and Linea.

> **Diagram:** [Bridge UI Architecture](../diagrams/bridge-ui-architecture.mmd) (Mermaid source)

## Overview

The Bridge UI provides:
- Native bridge for ETH and ERC-20 tokens
- Bridge aggregator (LiFi) for multi-chain bridging
- CEX integration (LayerSwap) for direct transfers
- On-ramp (OnRamper) for fiat purchases
- Transaction history and claiming

## Architecture

```
┌────────────────────────────────────────────────────────────────────────┐
│                            BRIDGE UI                                   │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       Next.js App Router                         │  │
│  │                                                                  │  │
│  │ ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │  │
│  │ │    /         │  │ /native-     │  │ /bridge-aggregator       │ │  │
│  │ │   Landing    │  │   bridge     │  │  (LiFi Widget)           │ │  │
│  │ └──────────────┘  └──────────────┘  └──────────────────────────┘ │  │
│  │                                                                  │  │
│  │ ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │  │
│  │ │ /centralized │  │    /buy      │  │      /faq                │ │  │
│  │ │  -exchange   │  │ (OnRamper)   │  │                          │ │  │
│  │ │ (LayerSwap)  │  │              │  │                          │ │  │
│  │ └──────────────┘  └──────────────┘  └──────────────────────────┘ │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       Components Layer                           │  │
│  │                                                                  │  │
│  │ ┌────────────────────────────────────────────────────────────┐   │  │
│  │ │                   Bridge Components                        │   │  │
│  │ │                                                            │   │  │
│  │ │ ┌────────────┐  ┌────────────┐  ┌────────────┐             │   │  │
│  │ │ │ From Chain │  │ To Chain   │  │   Amount   │             │   │  │
│  │ │ │  Selector  │  │  Selector  │  │   Input    │             │   │  │
│  │ │ └────────────┘  └────────────┘  └────────────┘             │   │  │
│  │ │                                                            │   │  │
│  │ │ ┌────────────┐  ┌────────────┐  ┌────────────┐             │   │  │
│  │ │ │   Token    │  │  Claiming  │  │   Submit   │             │   │  │
│  │ │ │   List     │  │   Mode     │  │   Button   │             │   │  │
│  │ │ └────────────┘  └────────────┘  └────────────┘             │   │  │
│  │ │                                                            │   │  │
│  │ │ ┌─────────────────────────────────────────────────────┐    │   │  │
│  │ │ │             Transaction History                     │    │   │  │
│  │ │ └─────────────────────────────────────────────────────┘    │   │  │
│  │ │                                                            │   │  │
│  │ └────────────────────────────────────────────────────────────┘   │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       State & Services                           │  │
│  │                                                                  │  │
│  │ ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │  │
│  │ │   Zustand    │  │   Wagmi +    │  │      Web3Auth            │ │  │
│  │ │   Stores     │  │    Viem      │  │       Modal              │ │  │
│  │ └──────────────┘  └──────────────┘  └──────────────────────────┘ │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
bridge-ui/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── (main)/
│   │   │   ├── page.tsx        # Landing page
│   │   │   ├── native-bridge/
│   │   │   ├── bridge-aggregator/
│   │   │   ├── centralized-exchange/
│   │   │   ├── buy/
│   │   │   └── faq/
│   │   ├── layout.tsx
│   │   └── providers.tsx
│   │
│   ├── components/
│   │   ├── bridge/
│   │   │   ├── form/           # Bridge form
│   │   │   ├── from-chain/     # Source chain selector
│   │   │   ├── to-chain/       # Destination chain selector
│   │   │   ├── amount/         # Amount input
│   │   │   ├── token-list/     # Token dropdown
│   │   │   ├── claiming/       # Claim mode (auto/manual)
│   │   │   ├── submit/         # Submit button
│   │   │   └── transaction-history/
│   │   ├── header/
│   │   ├── lifi/               # LiFi widget
│   │   ├── layerswap/          # LayerSwap widget
│   │   └── modal/
│   │
│   ├── contexts/
│   │   ├── Web3Provider.tsx    # Web3Auth + Wagmi
│   │   └── WalletDetectionProvider.tsx
│   │
│   ├── stores/                 # Zustand stores
│   │   ├── bridgeStore.ts
│   │   ├── chainStore.ts
│   │   ├── tokenStore.ts
│   │   └── historyStore.ts
│   │
│   ├── hooks/
│   │   ├── useEthBridgeTxArgs.ts
│   │   ├── useERC20BridgeTxArgs.ts
│   │   ├── useBridgingFees.ts
│   │   └── useMessageStatus.ts
│   │
│   ├── abis/                   # Contract ABIs
│   │   ├── TokenBridge.json
│   │   ├── MessageService.json
│   │   └── USDCBridge.json
│   │
│   └── config/
│       └── config.ts           # Chain and contract config
│
├── public/
│   └── assets/
│
└── package.json
```

## Tech Stack

| Technology | Purpose |
|------------|---------|
| Next.js 15 | React framework (App Router) |
| React 19 | UI library |
| Wagmi 2 | Ethereum interactions |
| Viem | Low-level Ethereum client |
| Web3Auth | Wallet connection modal |
| Zustand | State management |
| SCSS Modules | Styling |
| Motion | Animations |
| Playwright | E2E testing |

## Pages

### Native Bridge (`/native-bridge`)

Main bridging interface for ETH and ERC-20 tokens.

```tsx
export default function NativeBridgePage() {
  return (
    <BridgeLayout>
      <BridgeForm>
        <FromChainSelector />
        <ToChainSelector />
        <AmountInput />
        <TokenSelector />
        <ClaimingMode />
        <SubmitButton />
      </BridgeForm>
      <TransactionHistory />
    </BridgeLayout>
  );
}
```

### Bridge Aggregator (`/bridge-aggregator`)

LiFi widget for multi-chain bridging.

```tsx
export default function BridgeAggregatorPage() {
  return (
    <LiFiWidget
      config={{
        integrator: 'linea-bridge',
        appearance: 'dark',
        chains: supportedChains,
      }}
    />
  );
}
```

### Centralized Exchange (`/centralized-exchange`)

LayerSwap integration for CEX transfers.

```tsx
export default function CentralizedExchangePage() {
  return (
    <LayerSwapWidget
      config={{
        sourceChains: ['ethereum', 'binance', 'coinbase'],
        destinationChain: 'linea',
      }}
    />
  );
}
```

## Wallet Integration

### Web3Auth Setup

```tsx
// contexts/Web3Provider.tsx
import { Web3AuthConnector } from '@web3auth/modal-connector';
import { WagmiProvider, createConfig } from 'wagmi';

const config = createConfig({
  chains: [mainnet, linea, sepolia, lineaSepolia],
  connectors: [
    web3AuthConnector({
      web3AuthNetwork: 'mainnet',
      clientId: process.env.NEXT_PUBLIC_WEB3AUTH_CLIENT_ID,
    }),
    coinbaseWallet({
      appName: 'Linea Bridge',
    }),
    walletConnect({
      projectId: process.env.NEXT_PUBLIC_WC_PROJECT_ID,
    }),
  ],
  transports: {
    [mainnet.id]: http(),
    [linea.id]: http(),
  },
});

export function Web3Provider({ children }) {
  return (
    <WagmiProvider config={config}>
      {children}
    </WagmiProvider>
  );
}
```

## State Management

### Bridge Store

```typescript
// stores/bridgeStore.ts
import { create } from 'zustand';

interface BridgeState {
  sourceChain: Chain | null;
  destinationChain: Chain | null;
  token: Token | null;
  amount: string;
  recipient: string;
  claimMode: 'auto' | 'manual';
  
  setSourceChain: (chain: Chain) => void;
  setDestinationChain: (chain: Chain) => void;
  setToken: (token: Token) => void;
  setAmount: (amount: string) => void;
  setRecipient: (recipient: string) => void;
  setClaimMode: (mode: 'auto' | 'manual') => void;
  swap: () => void;
  reset: () => void;
}

export const useBridgeStore = create<BridgeState>((set) => ({
  sourceChain: null,
  destinationChain: null,
  token: null,
  amount: '',
  recipient: '',
  claimMode: 'auto',
  
  setSourceChain: (chain) => set({ sourceChain: chain }),
  setDestinationChain: (chain) => set({ destinationChain: chain }),
  // ... other setters
  
  swap: () => set((state) => ({
    sourceChain: state.destinationChain,
    destinationChain: state.sourceChain,
  })),
  
  reset: () => set({
    token: null,
    amount: '',
    recipient: '',
  }),
}));
```

## Contract Interactions

### ETH Bridging

```typescript
// hooks/useEthBridgeTxArgs.ts
export function useEthBridgeTxArgs() {
  const { sourceChain, destinationChain, amount, recipient } = useBridgeStore();
  
  const txArgs = useMemo(() => {
    if (!amount || !recipient) return null;
    
    const contract = sourceChain.id === mainnet.id
      ? MESSAGE_SERVICE_L1
      : MESSAGE_SERVICE_L2;
    
    return {
      address: contract,
      abi: MessageServiceAbi,
      functionName: 'sendMessage',
      args: [recipient, fee, '0x'],
      value: parseEther(amount) + fee,
    };
  }, [sourceChain, amount, recipient]);
  
  return txArgs;
}
```

### ERC-20 Bridging

```typescript
// hooks/useERC20BridgeTxArgs.ts
export function useERC20BridgeTxArgs() {
  const { sourceChain, token, amount, recipient } = useBridgeStore();
  
  // Approval transaction
  const approvalArgs = useMemo(() => ({
    address: token.address,
    abi: erc20Abi,
    functionName: 'approve',
    args: [TOKEN_BRIDGE_ADDRESS, parseUnits(amount, token.decimals)],
  }), [token, amount]);
  
  // Bridge transaction
  const bridgeArgs = useMemo(() => ({
    address: TOKEN_BRIDGE_ADDRESS,
    abi: TokenBridgeAbi,
    functionName: 'bridgeToken',
    args: [
      token.address,
      parseUnits(amount, token.decimals),
      recipient,
    ],
    value: fee,
  }), [token, amount, recipient]);
  
  return { approvalArgs, bridgeArgs };
}
```

## Configuration

### Supported Chains

```typescript
// config/config.ts
export const chains = {
  mainnet: {
    l1: mainnet,
    l2: linea,
    contracts: {
      messageService: '0xd19d4B5d358258f05D7B411E21A1460D11B0876F',
      tokenBridge: '0x051F1D88f0aF5763fB888eC78378F1109b52Cd01',
      l2MessageService: '0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec',
      l2TokenBridge: '0x353012dc4a9A6cF55c941bADC267f82004A8ceB9',
    },
  },
  sepolia: {
    l1: sepolia,
    l2: lineaSepolia,
    contracts: {
      // Sepolia addresses...
    },
  },
};
```

## Building

```bash
cd bridge-ui

# Install dependencies
pnpm install

# Development
pnpm run dev

# Build
pnpm run build

# Start production
pnpm run start

# E2E tests
pnpm run test:e2e
```

## Environment Variables

```bash
# Web3Auth
NEXT_PUBLIC_WEB3AUTH_CLIENT_ID=...

# WalletConnect
NEXT_PUBLIC_WC_PROJECT_ID=...

# Feature Flags
NEXT_PUBLIC_ENABLE_CCTP=false
NEXT_PUBLIC_ENABLE_AGGREGATOR=true

# Analytics
NEXT_PUBLIC_GTM_ID=...
```

## Features

### Transaction History

- Fetches on-chain events for user's address
- Caches locally in browser storage
- Supports restoration from blockchain
- Shows claiming status

### Auto/Manual Claiming

- **Auto**: Postman service claims automatically
- **Manual**: User claims themselves (saves on fees)

### CCTP Support (Feature Flag)

Cross-Chain Transfer Protocol for USDC bridging.

## Supported Networks

| Network | Chain ID | Type |
|---------|----------|------|
| Ethereum Mainnet | 1 | L1 |
| Linea Mainnet | 59144 | L2 |
| Sepolia | 11155111 | L1 Testnet |
| Linea Sepolia | 59141 | L2 Testnet |
