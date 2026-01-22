# Bytecode Verifier

A TypeScript tool to verify deployed smart contract bytecode and ABI against local artifact files. Supports multiple chains via configuration.

Inspired by [diffyscan](https://github.com/lidofinance/diffyscan).

## Usage

```bash
npx ts-node scripts/operational/bytecode-verifier/index.ts --config <config.json>
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
| `--help` | `-h` | Show help |

### Examples

```bash
# Verify all contracts in config
npx ts-node scripts/operational/bytecode-verifier/index.ts -c config.json

# Verbose output
npx ts-node scripts/operational/bytecode-verifier/index.ts -c config.json -v

# Filter to specific contract
npx ts-node scripts/operational/bytecode-verifier/index.ts -c config.json --contract LineaRollup

# Filter to specific chain
npx ts-node scripts/operational/bytecode-verifier/index.ts -c config.json --chain mainnet

# Skip ABI comparison (bytecode only)
npx ts-node scripts/operational/bytecode-verifier/index.ts -c config.json --skip-abi
```

## Configuration

Create a JSON configuration file. See `config.example.json` for a template.

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

Uses standard Hardhat artifact format (`.json`) with `abi`, `bytecode`, and `deployedBytecode` fields.

Example location: `contracts/deployments/bytecode/2026-01-14/LineaRollup.json`

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

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All verifications passed |
| 1 | One or more verifications failed |
| 2 | Configuration or runtime error |

## Limitations

- **Immutables**: Immutable variables are embedded in bytecode and may cause mismatches
- **Constructor arguments**: Not compared (deployment bytecode vs runtime bytecode)
- **Libraries**: Linked libraries may have different addresses per deployment
- **ABI heuristics**: Function selector extraction from bytecode is heuristic-based

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
