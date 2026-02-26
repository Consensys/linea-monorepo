# Linea Native Yield Automation Service

## Overview

Automates native yield operations on Linea by monitoring the YieldManager contract and executing the appropriate operation mode (yield reporting, ossification pending, or ossification complete). See [docs/architecture.md](./docs/architecture.md) for design details.

## Folder Structure

```
automation-service/
├── src/
│   ├── application/          # Application bootstrap and configuration
│   │   ├── main/             # Service entry point and config loading
│   │   └── metrics/          # Metrics service and updaters
│   ├── clients/              # External service clients
│   │   ├── contracts/        # Smart contract clients (LazyOracle, YieldManager, VaultHub, etc.)
│   ├── core/                 # Interfaces
│   │   ├── abis/             # Contract ABIs
│   │   ├── clients/          # Client interfaces
│   │   ├── entities/         # Domain entities and data models
│   │   ├── enums/            # Enums
│   │   ├── metrics/          # Metrics interfaces and types
│   │   └── services/         # Service interfaces
│   ├── services/             # Business logic services
│   │   ├── operation-mode-processors/  # Mode-specific processors
│   │   └── OperationModeSelector.ts    # Mode selection orchestrator
│   └── utils/                # Utility functions
└── scripts/                  # Manual integration runner scripts
```

## Configuration

See the [configuration schema file](./src/application/main/config/config.schema.ts) for all available options and their defaults.

### Key Environment Variables

Copy `.env.sample` as a starting point (`cp .env.sample .env`) and fill in real values.

| Category | Variable | Description |
|----------|----------|-------------|
| **RPC & APIs** | `CHAIN_ID` | L1 chain ID (e.g. `1` for mainnet) |
| | `L1_RPC_URL` | L1 Ethereum RPC endpoint |
| | `L1_RPC_URL_FALLBACK` | Optional fallback L1 RPC |
| | `BEACON_CHAIN_RPC_URL` | Beacon chain API endpoint |
| | `STAKING_GRAPHQL_URL` | Consensys Staking GraphQL endpoint |
| | `IPFS_BASE_URL` | IPFS gateway for Lido vault reports |
| **OAuth2** | `CONSENSYS_STAKING_OAUTH2_TOKEN_ENDPOINT` | Token endpoint |
| | `CONSENSYS_STAKING_OAUTH2_CLIENT_ID` | Client ID |
| | `CONSENSYS_STAKING_OAUTH2_CLIENT_SECRET` | Client secret |
| | `CONSENSYS_STAKING_OAUTH2_AUDIENCE` | Audience claim |
| **L1 Contracts** | `LINEA_ROLLUP_ADDRESS` | LineaRollup proxy |
| | `YIELD_MANAGER_ADDRESS` | YieldManager proxy |
| | `LAZY_ORACLE_ADDRESS` | Lido LazyOracle |
| | `VAULT_HUB_ADDRESS` | Lido VaultHub |
| | `LIDO_YIELD_PROVIDER_ADDRESS` | LidoStVaultYieldProvider |
| | `STETH_ADDRESS` | stETH token |
| **L2 Contracts** | `L2_YIELD_RECIPIENT` | Yield distribution recipient on L2 |
| **Timing** | `TRIGGER_EVENT_POLL_INTERVAL_MS` | Event polling interval |
| | `TRIGGER_MAX_INACTION_MS` | Max idle before forced execution |
| | `CONTRACT_READ_RETRY_TIME_MS` | Retry delay on read failures |
| | `GAUGE_METRICS_POLL_INTERVAL_MS` | Metrics gauge refresh interval |
| **Thresholds** | `SHOULD_SUBMIT_VAULT_REPORT` | Enable vault report submission |
| | `SHOULD_REPORT_YIELD` | Enable yield reporting |
| | `IS_UNPAUSE_STAKING_ENABLED` | Enable automatic unpause |
| | `MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI` | Min negative yield delta (wei) |
| | `CYCLES_PER_YIELD_REPORT` | Forced report every N cycles |
| | `REBALANCE_TOLERANCE_AMOUNT_WEI` | Rebalance tolerance band (wei) |
| | `STAKING_REBALANCE_QUOTA_BPS` | Rebalance quota (bps of TSB) |
| | `STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES` | Quota rolling window (0 = disabled) |
| **Unstaking** | `MAX_VALIDATOR_WITHDRAWAL_REQUESTS_PER_TRANSACTION` | Batch size per tx |
| | `MIN_WITHDRAWAL_THRESHOLD_ETH` | Min balance before withdrawal |
| **Web3Signer** | `WEB3SIGNER_URL` | Signer service URL (HTTPS) |
| | `WEB3SIGNER_PUBLIC_KEY` | Uncompressed secp256k1 public key |
| | `WEB3SIGNER_KEYSTORE_PATH` | mTLS keystore file path |
| | `WEB3SIGNER_KEYSTORE_PASSPHRASE` | Keystore passphrase |
| | `WEB3SIGNER_TRUSTSTORE_PATH` | mTLS truststore file path |
| | `WEB3SIGNER_TRUSTSTORE_PASSPHRASE` | Truststore passphrase |
| | `WEB3SIGNER_TLS_ENABLED` | TLS enabled flag |
| **Service** | `API_PORT` | Metrics API HTTP port (1024–49000) |
| | `LOG_LEVEL` | Log verbosity (default: `info`) |

## Development

### Running locally

1. Create `.env` as per `.env.sample` and the [configuration schema](./src/application/main/config/config.schema.ts)

2. `pnpm --filter @consensys/linea-native-yield-automation-service exec tsx run.ts`

### Build

```bash
# Dependency on pnpm --filter @consensys/linea-shared-utils build
pnpm --filter @consensys/linea-native-yield-automation-service build
```

### Unit tests

```bash
pnpm --filter @consensys/linea-native-yield-automation-service test
```

### Manual integration scripts

The `scripts/` directory contains manual runners for exercising individual clients against a live environment. Each script requires environment variables - see the file header for usage.

| Script | Client under test |
|--------|-------------------|
| `test-consensys-staking-graphql-client.ts` | `ConsensysStakingApiClient` (validator data via GraphQL) |
| `test-lazy-oracle-contract-client.ts` | `LazyOracleContractClient` (event watching, oracle reads) |
| `test-lido-accounting-report-client.ts` | `LidoAccountingReportClient` (report fetching/submission) |
| `test-yield-manager-contract-client.ts` | `YieldManagerContractClient` (rebalance, withdraw, state queries) |

Example:

```bash
RPC_URL=https://0xrpc.io/hoodi \
PRIVATE_KEY=0xabc123... \
YIELD_MANAGER_ADDRESS=0x... \
pnpm --filter @consensys/linea-native-yield-automation-service exec tsx scripts/test-yield-manager-contract-client.ts
```

## License

This package is licensed under the [Apache 2.0](../../LICENSE-APACHE) and the [MIT](../../LICENSE-MIT) licenses.
