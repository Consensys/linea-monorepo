# Contract Address Registry

This directory contains the manually-maintained, per-network deployed address registry for the Linea protocol.

## Files

| File | Network | Chain ID |
|------|---------|----------|
| `mainnet.json` | Ethereum mainnet | 1 |
| `sepolia.json` | Ethereum Sepolia testnet | 11155111 |
| `hoodi.json` | Ethereum Hoodi testnet | 560048 |
| `linea_mainnet.json` | Linea mainnet (L2) | 59144 |
| `linea_sepolia.json` | Linea Sepolia (L2 testnet) | 59141 |
| `schema.json` | JSON Schema for all registry files | — |

## How Addresses Are Used

Deploy scripts call `requireAddressOrRegistry(networkName, contractKey, envVarName)`. The resolution order is:

| Registry entry | Env var set | Outcome |
|---|---|---|
| Present (non-zero) | Not set | Registry address used |
| Present (non-zero) | **Matches** registry | Registry address used |
| Present (non-zero) | **Conflicts** with registry | **Hard fail before any transaction** |
| Absent or zero address | Set | Env var used (format-validated) |
| Absent or zero address | Not set | Hard fail — no source available |

Networks without a registry file (`custom`, `zkevm_dev`, `l2`) skip the registry entirely
and fall back to requiring the env var.

## How to Update an Address

1. Edit the appropriate network JSON file with the new EIP-55 checksummed address.
2. Open a pull request — the registry change is the auditable record that the address changed.
3. Get a review from a team member familiar with the deployment.
4. Remove any corresponding env var from `.env` files to avoid a conflict error on the next deploy.

## Address Entry Format

```json
{
  "ContractKey": {
    "address": "0x...",
    "notes": "Optional human-readable notes"
  }
}
```

- `address` must be a valid EIP-55 checksummed Ethereum address.
- Zero address (`0x0000...0000`) is treated as a placeholder meaning "not yet populated".
  Registry entries initialised with the zero address are ignored and env vars are used instead.
- `notes` is free-form text for context (proxy type, multisig info, etc.).

## Contract Key Mapping

The following keys are recognised by deploy scripts:

| Key | Env var equivalent | Notes |
|-----|--------------------|-------|
| `LineaRollup` | `LINEA_ROLLUP_ADDRESS` | Transparent upgradeable proxy (L1) |
| `Validium` | — | Transparent upgradeable proxy (L1) |
| `L2MessageService` | `L2_MESSAGE_SERVICE_ADDRESS` | Transparent upgradeable proxy (L2) |
| `TokenBridge_L1` | `TOKEN_BRIDGE_ADDRESS` (when `TOKEN_BRIDGE_L1=true`) | L1 token bridge proxy |
| `TokenBridge_L2` | `TOKEN_BRIDGE_ADDRESS` (when `TOKEN_BRIDGE_L1` unset) | L2 token bridge proxy |
| `CallForwardingProxy` | — | CallForwardingProxy (L1) |
| `YieldManager` | `YIELD_MANAGER_ADDRESS` | YieldManager proxy (L1) |
| `AddressFilter` | `LINEA_ROLLUP_ADDRESS_FILTER` | AddressFilter (L1) |
| `RollupRevenueVault` | `ROLLUP_REVENUE_VAULT_ADDRESS` | RollupRevenueVault proxy (L2) |
| `L1_SECURITY_COUNCIL` | `L1_SECURITY_COUNCIL` | L1 security council multisig |
| `L2_SECURITY_COUNCIL` | `L2_SECURITY_COUNCIL` | L2 security council multisig |
| `NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS` | `NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS` | Yield automation service operator (L1) |
| `VAULT_HUB` | `VAULT_HUB` | Lido VaultHub proxy (L1) |
| `VAULT_FACTORY` | `VAULT_FACTORY` | Lido Staking Vault Factory (L1) |
| `STETH` | `STETH` | Lido stETH token proxy (L1) |

> **Note:** `PLONKVERIFIER_ADDRESS` is env-var-only — it rotates with each proof system upgrade and is not tracked in the registry.
