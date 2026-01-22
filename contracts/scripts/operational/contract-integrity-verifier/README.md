# Contract Integrity Verifier

A TypeScript tool to verify deployed smart contract integrity (bytecode, ABI, and state) against local artifact files. Supports multiple chains via configuration.

Inspired by [diffyscan](https://github.com/lidofinance/diffyscan).

## Usage

```bash
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts --config <config.json>
```

### Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--config` | `-c` | Path to configuration file (required) |
| `--verbose` | `-v` | Enable verbose output |
| `--contract` | | Filter to specific contract name |
| `--chain` | | Filter to specific chain |
| `--skip-bytecode` | | Skip bytecode comparison |
| `--skip-abi` | | Skip ABI comparison |
| `--skip-state` | | Skip state verification |
| `--help` | `-h` | Show help |

### Examples

```bash
# Verify all contracts in config
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts -c config.json

# Verbose output
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts -c config.json -v

# Filter to specific contract
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts -c config.json --contract LineaRollup

# Filter to specific chain
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts -c config.json --chain mainnet

# Skip ABI comparison (bytecode only)
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts -c config.json --skip-abi
```

## Project Structure

```
contract-integrity-verifier/
├── src/                      # Source code
│   ├── index.ts              # Package entry point (exports public API)
│   ├── cli.ts                # CLI entry point
│   ├── types.ts              # TypeScript types
│   ├── config.ts             # Config loading
│   ├── verifier.ts           # Main verification logic
│   └── utils/                # Utility modules
│       ├── index.ts          # Utils barrel export
│       ├── abi.ts            # ABI parsing
│       ├── bytecode.ts       # Bytecode comparison
│       ├── state.ts          # State verification
│       └── storage-path.ts   # ERC-7201 storage paths
├── tools/                    # Standalone CLI tools
│   └── generate-schema.ts    # Schema generator
├── tests/                    # Test files
│   └── run-tests.ts
├── configs/                  # Example configurations
│   ├── chains.json
│   ├── example.json
│   └── sepolia-linea-rollup-v7.json
├── schemas/                  # Example storage schemas
│   └── linea-rollup.json
└── README.md
```

## Configuration

Create a JSON configuration file. See `configs/example.json` for a template.

### Configuration Schema

```json
{
  "chains": {
    "<chain-name>": {
      "chainId": 1,
      "rpcUrl": "${ENV_VAR_NAME}",
      "explorerUrl": "https://etherscan.io"
    }
  },
  "contracts": [
    {
      "name": "ContractName",
      "chain": "<chain-name>",
      "address": "0x...",
      "artifactFile": "path/to/artifact.json",
      "isProxy": true
    }
  ]
}
```

### Configuration Fields

#### Chains

| Field | Required | Description |
|-------|----------|-------------|
| `chainId` | Yes | Chain ID (e.g., 1 for mainnet) |
| `rpcUrl` | Yes | RPC endpoint URL. Supports `${ENV_VAR}` syntax |
| `explorerUrl` | No | Block explorer URL for reference |

#### Contracts

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Contract name for display |
| `chain` | Yes | Chain name (must match a key in `chains`) |
| `address` | Yes | Deployed contract address (proxy or implementation) |
| `artifactFile` | Yes | Path to Hardhat artifact `.json` file |
| `isProxy` | No | If `true`, fetches implementation bytecode via EIP-1967 |
| `constructorArgs` | No | Constructor arguments for immutable validation (see below) |

### Constructor Arguments

For contracts with immutable values set via constructor, provide `constructorArgs` to validate that bytecode differences are only due to these values:

```json
{
  "name": "MyContract",
  "chain": "mainnet",
  "address": "0x...",
  "artifactFile": "path/to/artifact.json",
  "constructorArgs": ["0x1234...5678", 1000000, true]
}
```

Supported argument formats:
- **Addresses**: `"0x1234567890123456789012345678901234567890"`
- **Numbers**: `1000000` or `"1000000000000000000"` (for large numbers as strings)
- **Booleans**: `true` or `false`
- **Pre-encoded hex**: `"0x000000000000000000000000..."` (full 32-byte encoded)

The tool will:
1. Compare bytecode and identify difference regions
2. Determine if differences are at immutable positions (32-byte aligned)
3. Validate that immutable values match provided constructor args
4. Report pass/fail with details

### Environment Variables

RPC URLs often contain API keys. Use `${ENV_VAR}` syntax in the config:

```json
{
  "rpcUrl": "${MAINNET_RPC_URL}"
}
```

Then set the environment variable:

```bash
export MAINNET_RPC_URL="https://mainnet.infura.io/v3/YOUR_KEY"
```

## Artifact File Format

Supports both **Hardhat** and **Foundry** artifact formats. The tool auto-detects the format.

### Hardhat Artifacts

Standard Hardhat format with `abi`, `bytecode`, and `deployedBytecode` as strings:

```json
{
  "_format": "hh-sol-artifact-1",
  "contractName": "MyContract",
  "abi": [...],
  "bytecode": "0x...",
  "deployedBytecode": "0x..."
}
```

### Foundry Artifacts

Foundry (`forge build`) format with bytecode as objects:

```json
{
  "abi": [...],
  "bytecode": { "object": "0x..." },
  "deployedBytecode": {
    "object": "0x...",
    "immutableReferences": {
      "123": [{ "start": 456, "length": 32 }]
    }
  },
  "methodIdentifiers": {
    "owner()": "8da5cb5b"
  }
}
```

### Foundry Benefits

When using Foundry artifacts, the tool leverages:

| Feature | Benefit |
|---------|---------|
| `immutableReferences` | Precise immutable position detection (no heuristics) |
| `methodIdentifiers` | Pre-computed selectors (faster ABI comparison) |

Example location: `out/MyContract.sol/MyContract.json`

## How Verification Works

### Bytecode Comparison

1. Fetches bytecode from the chain at the specified address
2. For proxy contracts (`isProxy: true`), reads the EIP-1967 implementation slot and fetches that bytecode
3. Strips CBOR metadata from both local and remote bytecode (metadata hash varies between compilations)
4. Compares the core bytecode

### ABI Comparison

1. Extracts function selectors from the ABI (4-byte keccak256 of function signature)
2. Scans the deployed bytecode for PUSH4 opcodes (function dispatcher pattern)
3. Reports any ABI selectors not found in the bytecode

### State Verification (Optional)

For upgradeable contracts, verify state set by initializers:

```json
{
  "name": "MyContract-Proxy",
  "address": "0x...",
  "artifactFile": "...",
  "isProxy": true,
  "stateVerification": {
    "ozVersion": "v5",
    "viewCalls": [
      { "function": "owner", "expected": "0x..." },
      { "function": "hasRole", "params": ["0x00...00", "0x..."], "expected": true }
    ],
    "slots": [
      { "slot": "0x0", "type": "uint8", "name": "_initialized", "expected": "6" }
    ],
    "namespaces": [
      {
        "id": "linea.storage.YieldManager",
        "variables": [
          { "offset": 0, "type": "address", "name": "messageService", "expected": "0x..." }
        ]
      }
    ]
  }
}
```

#### State Verification Methods

| Method | Use Case |
|--------|----------|
| `viewCalls` | Call public view functions (with optional params) |
| `slots` | Read explicit storage slots (OZ v4, private vars) |
| `namespaces` | Read ERC-7201 namespaced storage (OZ v5) |
| `storagePaths` | Schema-based storage paths with auto slot computation |

#### View Calls with Parameters

```json
{
  "viewCalls": [
    { "function": "owner", "expected": "0x..." },
    { "function": "hasRole", "params": ["0x00...00", "0xAdminAddress"], "expected": true },
    { "function": "balanceOf", "params": ["0xUserAddress"], "expected": "1000000000000000000" }
  ]
}
```

#### Storage Types

Supported types for slots and namespaces:
- `address`, `bool`
- `uint8`, `uint32`, `uint64`, `uint128`, `uint256`
- `bytes32`

#### Storage Paths (Schema-based)

For complex ERC-7201 storage structures, define a schema and use human-readable paths:

**Schema file** (`schemas/my-contract.json`):

```json
{
  "structs": {
    "MyStorage": {
      "namespace": "my.namespace.MyStorage",
      "fields": {
        "yieldManager": { "slot": 0, "type": "address" },
        "config": { "slot": 1, "type": "uint256" },
        "isPaused": { "slot": 2, "type": "bool", "byteOffset": 0 }
      }
    }
  }
}
```

**Config usage**:

```json
{
  "stateVerification": {
    "schemaFile": "../schemas/my-contract.json",
    "storagePaths": [
      { "path": "MyStorage:yieldManager", "expected": "0x..." },
      { "path": "MyStorage:config", "expected": "1000" }
    ]
  }
}
```

**Path syntax**:

| Pattern | Example | Description |
|---------|---------|-------------|
| `Struct:field` | `MyStorage:owner` | Simple field access |
| `Struct:a.b` | `MyStorage:config.value` | Nested field |
| `Struct:arr[0]` | `MyStorage:items[0]` | Array element |
| `Struct:arr.length` | `MyStorage:items.length` | Array length |
| `Struct:map[key]` | `MyStorage:balances[0x...]` | Mapping lookup |

Storage paths auto-compute slots from ERC-7201 namespace and field offsets.

#### OpenZeppelin Version Support

- **v4.x**: Use explicit `slots` with known slot positions
- **v5.x**: Use `namespaces` with ERC-7201 namespace IDs
- **Both**: Use `storagePaths` with schema file for readable configs

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All verifications passed |
| 1 | One or more verifications failed |
| 2 | Configuration or runtime error |

## Schema Generation

Generate storage schemas automatically from Solidity files containing ERC-7201 storage layouts:

```bash
npx ts-node scripts/operational/contract-integrity-verifier/tools/generate-schema.ts \
  --input src/yield/YieldManagerStorageLayout.sol \
  --output scripts/operational/contract-integrity-verifier/schemas/yield-manager.json \
  --verbose
```

### Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--input` | `-i` | Path to Solidity file with storage structs |
| `--output` | `-o` | Output path for JSON schema |
| `--verbose` | `-v` | Show detailed field information |

### What It Parses

The generator extracts:

1. **Struct definitions** - Field names, types, and slot positions
2. **ERC-7201 namespaces** - From `@custom:storage-location erc7201:...` NatSpec comments
3. **Packed storage** - Calculates `byteOffset` for fields sharing a slot
4. **Type normalization** - Enums → `uint8`, custom types preserved

### Example Output

```json
{
  "structs": {
    "YieldManagerStorage": {
      "namespace": "linea.storage.YieldManagerStorage",
      "baseSlot": "0xdc1272075efdca0b85fb2d76cbb5f26d954dc18e040d6d0b67071bd5cbd04300",
      "fields": {
        "minimumWithdrawalReservePercentageBps": { "slot": 0, "type": "uint16" },
        "targetWithdrawalReservePercentageBps": { "slot": 0, "type": "uint16", "byteOffset": 2 },
        "minimumWithdrawalReserveAmount": { "slot": 1, "type": "uint256" }
      }
    }
  }
}
```

### Post-Generation Review

After generating, review the schema for:

- **Mapping value types** - Complex types (e.g., `mapping(address => YieldProviderStorage)`) may need manual correction
- **Namespace accuracy** - Verify the namespace matches the Solidity `@custom:storage-location` exactly
- **Field names** - Ensure underscore prefixes (e.g., `_yieldManager`) are preserved

## Limitations

- **Immutables (Hardhat)**: Uses heuristic detection for immutable positions. Use Foundry artifacts for precise detection.
- **Immutables (Foundry)**: Exact positions from `immutableReferences` - no heuristics needed.
- **Constructor arguments**: Not compared (deployment bytecode vs runtime bytecode)
- **Libraries**: Linked libraries should be embedded in the artifact bytecode.
- **ABI heuristics**: Function selector extraction from bytecode is heuristic-based (Foundry artifacts use pre-computed `methodIdentifiers`).
- **Schema generator**: Mapping value types and complex custom types may need manual adjustment after generation.

## Troubleshooting

### "No bytecode found at address"

- Verify the contract is deployed on the specified chain
- Check that the RPC endpoint is correct and accessible

### "Environment variable not set"

- Set the required environment variable before running
- Check for typos in the `${VAR_NAME}` syntax

### Bytecode mismatch with high match percentage

- May be due to immutable variables
- Check compiler settings (optimizer, EVM version)
- Verify you're comparing the correct version
