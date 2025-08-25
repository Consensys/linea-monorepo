# Status Network Contract Deployment

This document explains how to deploy Status Network contracts alongside Linea contracts in the monorepo.

## Overview

The Status Network deployment includes the following contracts:

1. **StakeManager** - Manages staking logic and rewards
2. **VaultFactory** - Creates user staking vaults
3. **Karma** - Soulbound token for gasless transaction quotas
4. **RLN (Rate Limiting Nullifier)** - Zero-knowledge rate limiting system
5. **KarmaNFT** - NFT representation of Karma tokens

## Quick Start

### Deploy with RLN and Status Network Contracts

```bash
# Start the full stack with RLN and Status Network contracts
make start-env-with-rln
```

This will:
- Start L1 and L2 services with RLN support
- Deploy Linea protocol contracts
- Deploy Status Network contracts (Karma, StakeManager, RLN, etc.)
- Configure services for gasless transactions

### Deploy Status Network Contracts Only

```bash
# Deploy only Status Network contracts (requires L2 node running)
make deploy-status-network-contracts
```

### Deploy Using Hardhat

```bash
# Deploy using Hardhat deployment scripts
make deploy-status-network-contracts-hardhat
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STATUS_NETWORK_DEPLOYER` | `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266` | Deployer address for Status Network contracts |
| `STATUS_NETWORK_STAKING_TOKEN` | `0x0000000000000000000000000000000000000001` | SNT token address for staking |
| `STATUS_NETWORK_RLN_DEPTH` | `20` | RLN tree depth (supports ~1M users) |
| `STATUS_NETWORK_CONTRACTS_ENABLED` | `false` | Enable Status Network contract deployment |

### Example with Custom Configuration

```bash
# Deploy with custom configuration
STATUS_NETWORK_DEPLOYER=0x123... \
STATUS_NETWORK_STAKING_TOKEN=0x456... \
STATUS_NETWORK_RLN_DEPTH=24 \
make start-env-with-rln
```

## Contract Deployment Order

The contracts are deployed in the following order to respect dependencies:

1. **StakeManager** (with proxy)
2. **VaultFactory** (depends on StakeManager)
3. **Karma** (with proxy)
4. **RLN** (depends on Karma)
5. **KarmaNFT** (depends on Karma)

## Integration with RLN Services

When deploying with RLN enabled (`start-env-with-rln`), the following services are also started:

- **RLN Prover Service** - Generates zero-knowledge proofs
- **Karma Service** - Manages transaction quotas
- **Custom Sequencer** - Validates gasless transactions
- **Modified RPC Node** - Forwards transactions and handles gas estimation

## Post-Deployment Configuration

After deployment, the following manual steps may be required:

1. **Set Karma as Reward Supplier**:
   ```bash
   cast send $STAKE_MANAGER_ADDRESS "setRewardsSupplier(address)" $KARMA_ADDRESS
   ```

2. **Add StakeManager as Reward Distributor**:
   ```bash
   cast send $KARMA_ADDRESS "addRewardDistributor(address)" $STAKE_MANAGER_ADDRESS
   ```

3. **Whitelist Vault Implementation**:
   ```bash
   cast send $STAKE_MANAGER_ADDRESS "setTrustedCodehash(bytes32,bool)" $VAULT_CODEHASH true
   ```

## Troubleshooting

### Missing Status Network Contracts

If you see errors about missing contract artifacts, ensure that:

1. The `status-network-contracts` branch contains the compiled contracts
2. The contracts are properly compiled with compatible Solidity versions
3. The contract paths in deployment scripts are correct

### Gas Estimation Issues

For gasless transactions to work properly:

1. Ensure RLN services are running and healthy
2. Check that the Karma contract has users with positive balances
3. Verify that the deny list is properly configured

## Contract Verification

After deployment, contracts can be verified on block explorers:

```bash
# Verify StakeManager implementation
forge verify-contract $STAKE_MANAGER_IMPL_ADDRESS src/StakeManager.sol:StakeManager

# Verify Karma implementation  
forge verify-contract $KARMA_IMPL_ADDRESS src/Karma.sol:Karma

# Verify RLN implementation
forge verify-contract $RLN_IMPL_ADDRESS src/rln/RLN.sol:RLN
```

## Development

### Adding New Status Network Contracts

1. Add the contract to the `status-network-contracts` repository
2. Create a new deployment script in `contracts/deploy/`
3. Add the deployment target to `makefile-contracts.mk`
4. Update dependency chains in deployment scripts

### Testing

```bash
# Test Status Network contracts deployment
cd contracts
npx hardhat test --grep "StatusNetwork"
```

For more information about the individual contracts, see the [Status Network Contracts Documentation](../status-network-contracts/README.md).
