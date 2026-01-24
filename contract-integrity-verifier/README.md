# Contract Integrity Verifier

A TypeScript tool to verify deployed smart contract integrity against local artifacts. Validates bytecode, ABI, and on-chain state.

Supports **multiple web3 libraries** (ethers, viem) via an adapter pattern.

## Package Structure

| Package | Description | Dependencies |
|---------|-------------|--------------|
| `@consensys/linea-contract-integrity-verifier` | Core library with adapter interface | None (pure TypeScript) |
| `@consensys/linea-contract-integrity-verifier-ethers` | Ethers v6 adapter + CLI | peer: `ethers >=6.0.0` |
| `@consensys/linea-contract-integrity-verifier-viem` | Viem adapter + CLI | peer: `viem >=2.22.0` |

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

for (const contract of config.contracts) {
  const chain = config.chains[contract.chain];
  const result = await verifier.verifyContract(contract, chain, { verbose: true });
  console.log(result);
}
```

### With Viem

```typescript
import { Verifier, loadConfig } from "@consensys/linea-contract-integrity-verifier";
import { ViemAdapter } from "@consensys/linea-contract-integrity-verifier-viem";

const adapter = new ViemAdapter({ rpcUrl: "https://rpc.linea.build" });
const verifier = new Verifier(adapter);
const config = loadConfig("./config.json");

for (const contract of config.contracts) {
  const chain = config.chains[contract.chain];
  const result = await verifier.verifyContract(contract, chain, { verbose: true });
  console.log(result);
}
```

## CLI

Each adapter package includes a CLI tool:

```bash
# Using ethers adapter
npx verify-contract-ethers -c ./config.json -v

# Using viem adapter
npx verify-contract-viem -c ./config.json -v

# Or after building locally
cd contract-integrity-verifier/verifier-ethers
pnpm build
node dist/cli.mjs -c ../verifier-core/examples/configs/example.json -v
```

CLI Options:
- `-c, --config <PATH>` - Path to configuration file (required)
- `-v, --verbose` - Show detailed output
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
      "stateVerification": {
        "viewCalls": [...],
        "slots": [...],
        "storagePaths": [...]
      }
    }
  ]
}
```

### Markdown Configuration

See `verifier-core/examples/configs/` for Markdown configuration examples.

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
│   │   ├── config.ts                # Config loading (JSON + Markdown)
│   │   ├── index.ts                 # Public exports
│   │   ├── types.ts                 # All TypeScript types
│   │   ├── verifier.ts              # Main Verifier class
│   │   └── utils/
│   │       ├── abi.ts               # ABI utilities
│   │       ├── bytecode.ts          # Bytecode comparison
│   │       ├── markdown-config.ts   # Markdown config parser
│   │       └── storage.ts           # ERC-7201 storage utilities
│   ├── tests/
│   │   └── run-tests.ts             # Test suite
│   ├── tools/
│   │   └── generate-schema.ts       # Storage schema generator
│   └── examples/                    # Example configs and schemas
├── verifier-ethers/                  # @consensys/linea-contract-integrity-verifier-ethers
│   └── src/
│       ├── index.ts                 # EthersAdapter
│       └── cli.ts                   # CLI using ethers
└── verifier-viem/                    # @consensys/linea-contract-integrity-verifier-viem
    └── src/
        ├── index.ts                 # ViemAdapter
        └── cli.ts                   # CLI using viem
```

## Development

```bash
# Build all packages (order matters - core first)
cd contract-integrity-verifier/verifier-core && pnpm build
cd ../verifier-ethers && pnpm build
cd ../verifier-viem && pnpm build

# Or build all at once
pnpm --filter "@consensys/linea-contract-integrity-verifier" build
pnpm --filter "@consensys/linea-contract-integrity-verifier-ethers" build
pnpm --filter "@consensys/linea-contract-integrity-verifier-viem" build

# Typecheck
cd verifier-core && npx tsc --noEmit

# Lint
cd verifier-core && pnpm lint:fix
```

## Testing

### Unit Tests (Mock Adapter)

Unit tests use mock adapters and don't require network access:

```bash
# Core package - unit tests
cd verifier-core && pnpm test

# Adapter packages - integration tests with mock RPC
cd verifier-ethers && pnpm test:integration
cd verifier-viem && pnpm test:integration
```

### Live Integration Tests

Live tests connect to real networks and verify against deployed contracts.
Requires `ETHEREUM_SEPOLIA_RPC_URL` environment variable:

```bash
# Set RPC URL
export ETHEREUM_SEPOLIA_RPC_URL="https://sepolia.infura.io/v3/YOUR_KEY"

# Run live tests with ethers adapter
cd verifier-ethers && pnpm test:live

# Run live tests with viem adapter
cd verifier-viem && pnpm test:live
```

Live tests will skip gracefully if the environment variable is not set.

### Test Artifacts

The packages use real Hardhat artifacts from `contracts/deployments/bytecode/` for live tests.
Mock artifacts in `tests/fixtures/artifacts/` are used for offline unit tests.

## Features

- **Bytecode Verification**: Compare deployed bytecode against local artifacts
- **Immutable Detection**: Automatically detect and validate immutable values
- **ABI Verification**: Validate function selectors match artifact ABI
- **State Verification**: Verify on-chain state (storage slots, view calls)
- **ERC-7201 Support**: Compute and verify namespaced storage slots
- **Artifact Support**: Works with both Hardhat and Foundry artifacts
- **Markdown Config**: Human-readable configuration files
- **Multiple Web3 Libraries**: Use ethers or viem via adapter pattern

## Security Considerations

- **Input Validation**: All user inputs (addresses, paths, config values) are validated
- **Path Traversal**: File paths are resolved relative to config file location
- **Environment Variables**: Sensitive values (RPC URLs) should be passed via environment variables
- **No Secrets in Code**: Never commit RPC URLs or private keys to configuration files
