# Contract Integrity Verifier

A TypeScript tool to verify deployed smart contract integrity against local artifacts. Validates bytecode, ABI, and on-chain state.

Supports **multiple web3 libraries** (ethers, viem) via an adapter pattern.

## Package Structure

| Package | Description | Dependencies |
|---------|-------------|--------------|
| `@consensys/linea-contract-integrity-verifier` | Core library with adapter interface | None (pure TypeScript) |
| `@consensys/linea-contract-integrity-verifier-ethers` | Ethers v6 adapter | peer: `ethers >=6.0.0` |
| `@consensys/linea-contract-integrity-verifier-viem` | Viem adapter | peer: `viem >=2.22.0` |

## Installation

```bash
# Core + ethers adapter
pnpm add @consensys/linea-contract-integrity-verifier @consensys/linea-contract-integrity-verifier-ethers ethers

# Core + viem adapter
pnpm add @consensys/linea-contract-integrity-verifier @consensys/linea-contract-integrity-verifier-viem viem
```

## Usage

### With Ethers

```typescript
import { Verifier, loadConfig } from "@consensys/linea-contract-integrity-verifier";
import { EthersAdapter } from "@consensys/linea-contract-integrity-verifier-ethers";

const adapter = new EthersAdapter({ rpcUrl: "https://rpc.linea.build" });
const verifier = new Verifier(adapter);
const config = loadConfig("./config.json");

const result = await verifier.verify(config);
```

### With Viem

```typescript
import { Verifier, loadConfig } from "@consensys/linea-contract-integrity-verifier";
import { ViemAdapter } from "@consensys/linea-contract-integrity-verifier-viem";

const adapter = new ViemAdapter({ rpcUrl: "https://rpc.linea.build" });
const verifier = new Verifier(adapter);
const config = loadConfig("./config.json");

const result = await verifier.verify(config);
```

## CLI

The core package includes a CLI tool:

```bash
# After building
npx contract-integrity-verifier \
  -c ./config.json \
  --verbose
```

CLI Options:
- `-c, --config <PATH>` - Path to configuration file (required)
- `--verbose, -v` - Show detailed output
- `--contract <NAME>` - Filter to specific contract
- `--chain <NAME>` - Filter to specific chain
- `--skip-bytecode` - Skip bytecode verification
- `--skip-abi` - Skip ABI verification
- `--skip-state` - Skip state verification

## Configuration

Supports JSON and Markdown configuration formats.

### JSON Configuration

```json
{
  "chains": {
    "ethereum-sepolia": {
      "chainId": 11155111,
      "rpcUrl": "${ETHEREUM_SEPOLIA_RPC_URL}",
      "explorerUrl": "https://sepolia.etherscan.io"
    }
  },
  "contracts": [
    {
      "name": "MyContract-Proxy",
      "chain": "ethereum-sepolia",
      "address": "0x1234567890123456789012345678901234567890",
      "artifactFile": "../path/to/MyContract.json",
      "isProxy": true,
      "constructorArgs": ["0xMessageServiceAddress"],
      "stateVerification": { ... }
    }
  ]
}
```

### Markdown Configuration

See `examples/configs/` for Markdown configuration examples.

## Web3Adapter Interface

The adapter pattern allows the core library to work with any web3 library:

```typescript
interface Web3Adapter {
  // Crypto (synchronous)
  keccak256(value: string | Uint8Array): string;
  checksumAddress(address: string): string;
  readonly zeroAddress: string;

  // ABI (synchronous)
  encodeAbiParameters(types: readonly string[], values: readonly unknown[]): string;
  encodeFunctionData(abi: readonly AbiElement[], functionName: string, args?: readonly unknown[]): string;
  decodeFunctionResult(abi: readonly AbiElement[], functionName: string, data: string): readonly unknown[];

  // RPC (asynchronous)
  getCode(address: string): Promise<string>;
  getStorageAt(address: string, slot: string): Promise<string>;
  call(to: string, data: string): Promise<string>;
}
```

## Project Structure

```
contract-integrity-verifier/
├── verifier-core/                    # @consensys/linea-contract-integrity-verifier
│   ├── src/
│   │   ├── adapter.ts               # Web3Adapter interface
│   │   ├── cli.ts                   # CLI entry point
│   │   ├── config.ts                # Config loading (JSON + Markdown)
│   │   ├── index.ts                 # Public exports
│   │   ├── types.ts                 # All TypeScript types
│   │   ├── verifier.ts              # Main Verifier class
│   │   └── utils/
│   │       ├── abi.ts               # ABI utilities
│   │       ├── bytecode.ts          # Bytecode comparison
│   │       ├── markdown-config.ts   # Markdown config parser
│   │       ├── state.ts             # State verification
│   │       ├── storage.ts           # Storage utilities
│   │       └── storage-path.ts      # ERC-7201 storage path resolution
│   ├── tests/
│   │   └── run-tests.ts             # Test suite
│   ├── tools/
│   │   └── generate-schema.ts       # Storage schema generator
│   └── examples/                    # Example configs and schemas
├── verifier-ethers/                  # @consensys/linea-contract-integrity-verifier-ethers
│   └── src/
│       └── index.ts                 # EthersAdapter
└── verifier-viem/                    # @consensys/linea-contract-integrity-verifier-viem
    └── src/
        └── index.ts                 # ViemAdapter
```

## Development

```bash
# Build all packages
pnpm --filter "@consensys/linea-contract-integrity-verifier*" build

# Run tests
pnpm --filter "@consensys/linea-contract-integrity-verifier" test

# Typecheck
pnpm --filter "@consensys/linea-contract-integrity-verifier*" tsc --noEmit
```

## Features

- **Bytecode Verification**: Compare deployed bytecode against local artifacts
- **Immutable Detection**: Automatically detect and validate immutable values
- **ABI Verification**: Validate function selectors match artifact ABI
- **State Verification**: Verify on-chain state (storage slots, view calls)
- **ERC-7201 Support**: Compute and verify namespaced storage slots
- **Artifact Support**: Works with both Hardhat and Foundry artifacts
- **Markdown Config**: Human-readable configuration files
