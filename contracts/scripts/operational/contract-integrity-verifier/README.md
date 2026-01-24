# Contract Integrity Verifier

A TypeScript tool to verify deployed smart contract integrity against local artifacts. Validates bytecode, ABI, and on-chain state.

Inspired by [diffyscan](https://github.com/lidofinance/diffyscan).

## Quick Start

```bash
# From contracts/ directory
cd contracts

# Set RPC URL
export ETHEREUM_SEPOLIA_RPC_URL="https://sepolia.infura.io/v3/YOUR_KEY"

# Run verification
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts \
  -c scripts/operational/contract-integrity-verifier/examples/configs/example.json \
  -v
```

## Project Structure

```
contract-integrity-verifier/
├── src/                       # Source code
│   ├── cli.ts                 # CLI entry point
│   ├── config.ts              # Config loading (JSON + Markdown)
│   ├── verifier.ts            # Main verification logic
│   ├── types.ts               # TypeScript types
│   ├── index.ts               # Package exports
│   └── utils/                 # Utility modules
│       ├── abi.ts             # ABI verification
│       ├── bytecode.ts        # Bytecode comparison
│       ├── index.ts           # Utility exports
│       ├── markdown-config.ts # Markdown config parser
│       ├── state.ts           # State verification orchestration
│       └── storage-path.ts    # ERC-7201 storage path resolution
├── tools/                     # CLI tools
│   └── generate-schema.ts     # Storage schema generator
├── tests/
│   └── run-tests.ts
├── examples/                  # Example configs and schemas
│   ├── configs/
│   │   ├── chains.json
│   │   ├── example.json
│   │   ├── sepolia-linea-rollup-v7.config.json
│   │   └── sepolia-linea-rollup-v7.config.md
│   └── schemas/
│       ├── linea-rollup.json
│       └── yield-manager.json
└── README.md
```

## CLI Options

| Option | Alias | Description |
|--------|-------|-------------|
| `--config` | `-c` | Path to configuration file (required) |
| `--verbose` | `-v` | Enable verbose output |
| `--contract` | | Filter to specific contract name |
| `--chain` | | Filter to specific chain |
| `--skip-bytecode` | | Skip bytecode comparison |
| `--skip-abi` | | Skip ABI comparison |
| `--skip-state` | | Skip state verification |

## Configuration

Supports two formats: **JSON** and **Markdown**.

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

### Contract Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Display name |
| `chain` | Yes | Chain key from `chains` |
| `address` | Yes | Deployed address |
| `artifactFile` | Yes | Path to Hardhat/Foundry artifact |
| `isProxy` | No | If `true`, fetches implementation via EIP-1967 |
| `constructorArgs` | No | Constructor args for immutable validation |
| `stateVerification` | No | State checks (see below) |

### Markdown Configuration

Use markdown as both documentation AND config:

````markdown
## Contract: MyContract-Proxy

```verifier
name: MyContract-Proxy
address: 0x1234567890123456789012345678901234567890
chain: ethereum-sepolia
artifact: ../path/to/MyContract.json
isProxy: true
ozVersion: v4
schema: ../schemas/my-schema.json
```

### State Verification

| Type | Description | Check | Params | Expected |
|------|-------------|-------|--------|----------|
| viewCall | Get owner | `owner` | | `0xOwnerAddress` |
| slot | Initialized | `0x0` | uint8 | `1` |
| storagePath | Config | `MyStorage:config` | | `100` |
````

**Verifier block fields:**

| Field | Alias | Description |
|-------|-------|-------------|
| `name` | | Contract display name |
| `address` | | Deployed contract address |
| `chain` | | Chain key (default: `ethereum-mainnet`) |
| `artifact` | `artifactFile` | Path to artifact JSON |
| `isProxy` | | `true` if proxy contract |
| `ozVersion` | | `v4`, `v5`, or `auto` |
| `schema` | `schemaFile` | Path to storage schema JSON |

Default chains are included automatically: `ethereum-mainnet`, `ethereum-sepolia`, `linea-mainnet`, `linea-sepolia`.

---

## State Verification

State verification is **optional** and validates on-chain state after deployment/upgrade. Four methods are available:

### 1. View Calls (`viewCalls`)

Call public view/pure functions and check return values.

```json
{
  "stateVerification": {
    "viewCalls": [
      {
        "function": "owner",
        "expected": "0xOwnerAddress"
      },
      {
        "function": "hasRole",
        "params": ["0xRoleHash", "0xAccountAddress"],
        "expected": true
      },
      {
        "function": "balanceOf",
        "params": ["0xUserAddress"],
        "expected": "1000000000000000000",
        "comparison": "gte"
      }
    ]
  }
}
```

**View call fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `function` | Yes | Function name from ABI |
| `params` | No | Array of parameters |
| `expected` | Yes | Expected return value |
| `comparison` | No | Comparison mode (default: `"eq"`) |

**Comparison modes:** `eq` (equal), `gt`, `gte`, `lt`, `lte`, `contains`

**When to use**: Public getters, role checks, balance queries, any `view`/`pure` function.

### 2. Explicit Slots (`slots`)

Read raw storage slots directly via `eth_getStorageAt`.

```json
{
  "stateVerification": {
    "slots": [
      {
        "slot": "0x0",
        "type": "uint8",
        "name": "_initialized",
        "expected": "7"
      },
      {
        "slot": "0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc",
        "type": "address",
        "name": "implementation",
        "expected": "0xImplementationAddress"
      }
    ]
  }
}
```

**Supported types:**
- **Unsigned integers**: `uint8`, `uint16`, `uint32`, `uint64`, `uint96`, `uint128`, `uint256`
- **Signed integers**: `int8`, `int16`, `int32`, `int64`, `int96`, `int128`, `int256`
- **Other**: `address`, `bool`, `bytes32`

**When to use**:
- OpenZeppelin v4.x `_initialized` (slot 0)
- Known slot positions (e.g., EIP-1967 slots)
- Private variables with known slots

### 3. Storage Paths (`storagePaths`)

Human-readable paths for ERC-7201 namespaced storage. Requires a **schema file**.

```json
{
  "stateVerification": {
    "schemaFile": "../schemas/my-contract.json",
    "storagePaths": [
      {
        "path": "MyStorage:owner",
        "expected": "0xOwnerAddress"
      },
      {
        "path": "MyStorage:config.maxAmount",
        "expected": "1000000",
        "comparison": "gte"
      }
    ]
  }
}
```

**Storage path fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `path` | Yes | Storage path (see syntax below) |
| `expected` | Yes | Expected value |
| `comparison` | No | Comparison mode (default: `"eq"`) |

**Comparison modes:** `eq` (equal), `gt`, `gte`, `lt`, `lte`

**Path syntax**:

| Pattern | Example | Description |
|---------|---------|-------------|
| `Struct:field` | `MyStorage:owner` | Simple field |
| `Struct:a.b` | `MyStorage:config.value` | Nested field |
| `Struct:arr[0]` | `MyStorage:items[0]` | Array element |
| `Struct:arr[length]` | `MyStorage:items[length]` | Array length |
| `Struct:map[key]` | `MyStorage:balances[0x...]` | Mapping lookup |

**When to use**: ERC-7201 namespaced storage (OpenZeppelin v5.x pattern), complex storage layouts.

### 4. Namespaces (`namespaces`)

Direct ERC-7201 namespace verification without a schema file. Useful for quick checks or when schema is unavailable.

```json
{
  "stateVerification": {
    "namespaces": [
      {
        "id": "linea.storage.MyStorage",
        "variables": [
          {
            "offset": 0,
            "type": "address",
            "name": "owner",
            "expected": "0xOwnerAddress"
          },
          {
            "offset": 1,
            "type": "uint256",
            "name": "totalSupply",
            "expected": "1000000"
          }
        ]
      }
    ]
  }
}
```

**Namespace fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | ERC-7201 namespace ID (e.g., `"linea.storage.MyStorage"`) |
| `variables` | Yes | Array of variables to verify |

**Variable fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `offset` | Yes | Slot offset from namespace base (0, 1, 2, ...) |
| `type` | Yes | Solidity type for decoding |
| `name` | Yes | Display name |
| `expected` | Yes | Expected value |

**When to use**: Quick namespace verification, testing, or when you know the exact slot offsets.

---

## Storage Schemas

Schemas define the storage layout for `storagePaths` verification.

### Schema Structure

```json
{
  "structs": {
    "MyStorage": {
      "namespace": "my.namespace.MyStorage",
      "baseSlot": "0x...",
      "fields": {
        "owner": {
          "slot": 0,
          "type": "address"
        },
        "config": {
          "slot": 1,
          "type": "uint256"
        },
        "isPaused": {
          "slot": 2,
          "type": "bool",
          "byteOffset": 0
        },
        "version": {
          "slot": 2,
          "type": "uint8",
          "byteOffset": 1
        }
      }
    }
  }
}
```

### Field Properties

| Property | Required | Description |
|----------|----------|-------------|
| `slot` | Yes | Offset from struct base slot |
| `type` | Yes | Solidity type |
| `byteOffset` | No | Byte offset for packed storage (0-31) |

**Supported schema types:**
- **Unsigned integers**: `uint8`, `uint16`, `uint32`, `uint64`, `uint96`, `uint128`, `uint256`
- **Signed integers**: `int8`, `int16`, `int32`, `int64`, `int96`, `int128`, `int256`
- **Other**: `address`, `bool`, `bytes4`, `bytes32`
- **Complex**: `address[]`, `uint256[]`, `mapping(keyType => valueType)`

### Generating Schemas

Auto-generate from Solidity files:

```bash
npx ts-node scripts/operational/contract-integrity-verifier/tools/generate-schema.ts \
  --input src/storage/MyStorageLayout.sol \
  --output scripts/operational/contract-integrity-verifier/examples/schemas/my-storage.json \
  --verbose
```

The generator extracts:
- Struct definitions and field types
- ERC-7201 namespaces from `@custom:storage-location` NatSpec
- Packed storage byte offsets
- Base slot constants

**Post-generation review**: Verify mapping value types and namespace strings match your Solidity code.

---

## OpenZeppelin Version Support

| Version | Recommended Method | Notes |
|---------|-------------------|-------|
| v4.x | `slots` | `_initialized` at slot 0 |
| v5.x | `storagePaths` + schema | ERC-7201 namespaced storage |
| v5.x | `namespaces` | ERC-7201 without schema file |
| Both | `viewCalls` | Public getters work everywhere |

### Example: OZ v4 Contract

```json
{
  "stateVerification": {
    "ozVersion": "v4",
    "viewCalls": [
      { "function": "owner", "expected": "0x..." }
    ],
    "slots": [
      { "slot": "0x0", "type": "uint8", "name": "_initialized", "expected": "1" }
    ]
  }
}
```

### Example: OZ v5 Contract with ERC-7201

```json
{
  "stateVerification": {
    "ozVersion": "v5",
    "schemaFile": "../schemas/my-contract.json",
    "viewCalls": [
      { "function": "owner", "expected": "0x..." }
    ],
    "storagePaths": [
      { "path": "MyStorage:config", "expected": "100" }
    ]
  }
}
```

---

## Bytecode Verification

### How It Works

1. Fetches deployed bytecode from chain
2. For proxies (`isProxy: true`), reads EIP-1967 implementation slot
3. Strips CBOR metadata (varies between compilations)
4. Compares against local artifact

### Immutable Values

Contracts with immutables will have bytecode differences at deployment positions. The verifier:

1. Detects immutable positions (Foundry: precise, Hardhat: heuristic)
2. Reports positions and extracted values
3. If `constructorArgs` provided, validates they match

```json
{
  "constructorArgs": ["0xMessageServiceAddress", 1000000]
}
```

Output:
```
Immutable differences detected: 2
  Position 1234: address = 0xMessageServiceAddress
  Position 5678: uint256 = 1000000
Bytecode: ✓ Matches (2 immutable value(s) differ as expected)
```

---

## Artifact Formats

### Hardhat

```json
{
  "_format": "hh-sol-artifact-1",
  "abi": [...],
  "bytecode": "0x...",
  "deployedBytecode": "0x..."
}
```

### Foundry

```json
{
  "abi": [...],
  "bytecode": { "object": "0x..." },
  "deployedBytecode": {
    "object": "0x...",
    "immutableReferences": { ... }
  },
  "methodIdentifiers": { "owner()": "8da5cb5b" }
}
```

**Foundry benefits**: Precise immutable detection via `immutableReferences`, pre-computed selectors.

---

## Environment Variables

Use `${VAR_NAME}` syntax for sensitive values:

```json
{
  "rpcUrl": "${ETHEREUM_MAINNET_RPC_URL}"
}
```

```bash
export ETHEREUM_MAINNET_RPC_URL="https://mainnet.infura.io/v3/YOUR_KEY"
```

---

## Examples

### Verify a Proxy Contract

```bash
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts \
  -c scripts/operational/contract-integrity-verifier/examples/configs/sepolia-linea-rollup-v7.config.json \
  -v
```

### Verify Specific Contract Only

```bash
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts \
  -c examples/configs/example.json \
  --contract "MyContract-Proxy" \
  -v
```

### Skip State Verification

```bash
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts \
  -c config.json \
  --skip-state
```

### Using Markdown Config

```bash
npx ts-node scripts/operational/contract-integrity-verifier/src/cli.ts \
  -c scripts/operational/contract-integrity-verifier/examples/configs/sepolia-linea-rollup-v7.config.md \
  -v
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All verifications passed |
| 1 | One or more verifications failed |
| 2 | Configuration or runtime error |

---

## Troubleshooting

### "No bytecode found at address"
- Verify contract is deployed on the specified chain
- Check RPC endpoint is accessible

### "Environment variable not set"
- Set the required env var: `export VAR_NAME="value"`
- Check for typos in `${VAR_NAME}` syntax

### Bytecode mismatch with high match percentage
- May be immutable values - add `constructorArgs` to config
- Check compiler settings (optimizer, EVM version)
- Verify correct artifact version

### Storage path not found
- Verify schema file path is correct
- Check struct and field names match schema exactly
- Ensure namespace in schema matches Solidity `@custom:storage-location`

---

## Limitations

- **Immutables (Hardhat)**: Heuristic detection. Use Foundry for precision.
- **Libraries**: Must be embedded in artifact bytecode.
- **Schema generator**: Complex mapping types may need manual adjustment.

### Runtime vs Deployment Bytecode

The verifier compares **runtime bytecode only** (what's stored on-chain), not deployment bytecode:

```
Deployment tx:  [Constructor code | Runtime bytecode | Constructor args]
                                          │
                                          ▼ (stored on-chain)
Comparison:     artifact.deployedBytecode ←→ eth_getCode(address)
```

Constructor args are consumed during deployment and discarded. The `constructorArgs` config field validates that immutable value differences match expected args—it does not compare the original constructor calldata.
