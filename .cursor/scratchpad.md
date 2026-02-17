# Hardhat v2 to v3 Migration Plan

## Problem Statement

Migrate the Linea smart contracts project from Hardhat v2 (currently 2.28.3) to Hardhat v3. This involves significant changes to the deployment infrastructure, testing framework, plugin ecosystem, and configuration.

## Current State Analysis

### Dependencies (contracts/package.json)

| Package | Current Version | Hardhat v3 Status |
|---------|-----------------|-------------------|
| `hardhat` | 2.28.3 | Needs upgrade to 3.x |
| `hardhat-deploy` | 0.12.4 | **DEPRECATED** - Replace with Hardhat Ignition |
| `@nomicfoundation/hardhat-toolbox` | 4.0.0 | Needs v5 for Hardhat v3 |
| `@nomicfoundation/hardhat-ethers` | 3.0.5 | Needs update for v3 |
| `@nomicfoundation/hardhat-network-helpers` | 1.0.10 | Needs update for v3 |
| `@nomicfoundation/hardhat-verify` | 2.1.1 | Needs update for v3 |
| `@nomicfoundation/hardhat-foundry` | 1.1.3 | Needs update for v3 |
| `@openzeppelin/hardhat-upgrades` | 2.5.1 | Needs update for v3 |
| `@typechain/hardhat` | 9.1.0 | Needs update for v3 |
| `hardhat-storage-layout` | 0.1.7 | May need replacement/update |
| `hardhat-tracer` | 2.8.2 | Already broken, needs replacement |
| `hardhat-gas-reporter` | 2.3.0 | Needs update for v3 |
| `solidity-coverage` | 0.8.17 | Needs update for v3 |
| `solidity-docgen` | 0.6.0-beta.36 | Needs update for v3 |

### Files Requiring Changes

#### Deploy Scripts (31 files using `hardhat-deploy`)
All deploy scripts in `contracts/deploy/` use `DeployFunction` from `hardhat-deploy/types`:
- `01_deploy_PlonkVerifier.ts`
- `02_deploy_Timelock.ts`
- `03_deploy_LineaRollup.ts` (and variants)
- `04_deploy_L2MessageService.ts` (and variants)
- `05_deploy_BridgedToken.ts`
- `06_deploy_TokenBridge.ts` (and variants)
- ... and 20+ more

#### Test Files (81 files)
All test files in `contracts/test/hardhat/` use:
- `import { ethers, upgrades } from "hardhat"` - needs migration
- `import { loadFixture } from "@nomicfoundation/hardhat-network-helpers"` - needs update
- Typechain types from `contracts/typechain-types`

#### Task Files (7 files)
Custom Hardhat tasks in `contracts/scripts/operational/tasks/`:
- `grantContractRolesTask.ts`
- `renounceContractRolesTask.ts`
- `setRateLimitTask.ts`
- `setVerifierAddressTask.ts`
- `setMessageServiceOnTokenBridgeTask.ts`
- `getCurrentFinalizedBlockNumberTask.ts`

#### Utility Files
- `contracts/scripts/hardhat/utils.ts` - deployment utilities
- `contracts/test/hardhat/common/deployment.ts` - test deployment helpers

### Configuration File (hardhat.config.ts)
Current config uses:
- `namedAccounts` from hardhat-deploy (not native to v3)
- Custom hardfork settings (`osaka`)
- Solidity compiler overrides (from `hardhat_overrides.ts`)
- Multiple network configurations
- Etherscan/docgen/gas reporter configs

## Key Hardhat v3 Changes

1. **ESM-First Architecture**: Hardhat v3 is ESM-native
2. **New Plugin System**: Plugins are now Hardhat plugins with new registration API
3. **hardhat-deploy Deprecated**: Replace with Hardhat Ignition for deployments
4. **Configuration Format**: Some config options have changed
5. **Network Helpers**: New API for network manipulation
6. **Type Generation**: TypeChain integration may change

## Assumptions

1. We want to maintain feature parity with current functionality
2. Deployment reproducibility is critical
3. Test coverage must remain intact
4. CI/CD pipelines need to continue working
5. Foundry integration must be preserved
6. Local deployment scripts (`local-deployments-artifacts/`) need to work

## Migration Strategy

### Phase 1: Foundation (PR #1)
Prepare the codebase for migration without breaking changes:
- Update to latest Hardhat v2 versions of all plugins
- Fix any deprecation warnings
- Ensure all tests pass
- Document current behavior

### Phase 2: Dependency Updates (PR #2)
Update core dependencies:
- Upgrade `hardhat` to v3
- Update `@nomicfoundation/hardhat-toolbox` to v5
- Update `@nomicfoundation/hardhat-ethers` for v3
- Update `@nomicfoundation/hardhat-network-helpers` for v3
- Update `@typechain/hardhat` for v3

### Phase 3: Configuration Migration (PR #3)
Update configuration:
- Migrate `hardhat.config.ts` to v3 format
- Update `hardhat_overrides.ts` if needed
- Update compiler settings format
- Update network configurations

### Phase 4: Replace hardhat-deploy with Ignition (PR #4-6)
This is the largest change - migrate all deployment scripts:
- Install and configure `@nomicfoundation/hardhat-ignition`
- Create Ignition modules for each deployment
- Migrate `namedAccounts` to Ignition accounts
- Update deployment tracking

Sub-PRs:
- PR #4: Core infrastructure contracts (PlonkVerifier, Timelock, LineaRollup)
- PR #5: Messaging contracts (L2MessageService variants)
- PR #6: Bridge contracts (TokenBridge, BridgedToken)
- Additional PRs as needed for remaining contracts

### Phase 5: Test Migration (PR #7-9)
Update all test files:
- Update imports from `hardhat` package
- Update network helpers usage
- Update fixture patterns if needed
- Update TypeChain usage

Sub-PRs:
- PR #7: Rollup tests
- PR #8: Messaging and bridge tests
- PR #9: Utility and operational tests

### Phase 6: Task Migration (PR #10)
Update custom Hardhat tasks:
- Migrate task definitions to v3 API
- Update `hre` usage patterns
- Test all operational tasks

### Phase 7: Plugin Updates (PR #11-13)
Update remaining plugins:
- PR #11: `@openzeppelin/hardhat-upgrades` for v3
- PR #12: Coverage and gas reporter
- PR #13: Foundry integration, docgen, storage-layout

### Phase 8: Local Deployment Scripts (PR #14)
Update local deployment utilities:
- Update `local-deployments-artifacts/` scripts
- Update `makefile-contracts.mk` if needed
- Test local development workflow

### Phase 9: CI/CD Updates (PR #15)
Update GitHub workflows:
- Update `run-smc-tests.yml`
- Update any other workflows using Hardhat
- Ensure all pipelines pass

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Plugin incompatibility | High | High | Research each plugin early, find alternatives |
| Test failures | Medium | High | Run tests incrementally, maintain v2 branch |
| Deployment script migration errors | High | Critical | Extensive testing, use staging environments |
| CI pipeline breakage | Medium | Medium | Update workflows in parallel branch |
| Performance regression | Low | Medium | Benchmark before/after |

## Rollout Plan

1. Create feature branch from main
2. Implement changes in phases
3. Maintain working state after each PR
4. Test each phase thoroughly
5. Keep v2 branch as fallback
6. Gradual rollout with monitoring

## Rollback Strategy

- Maintain `hardhat-v2` branch with working state
- Each PR should be revertible
- Keep old `package.json` and lockfile for rollback
- Document rollback procedures for each phase

## Open Questions

1. Is Hardhat v3 stable enough for production use?
2. Are all required plugins available for v3?
3. What is the timeline requirement for this migration?
4. Are there any blocked features on v3?
5. Should we consider alternatives like Foundry-only?

## Decision

Proceed with phased migration as outlined above. Each phase produces a working, testable state.
