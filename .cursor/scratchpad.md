# Hardhat v3 Migration Plan - Linea Smart Contracts

## Problem

Migrate `contracts/` from Hardhat v2.28.3 to Hardhat v3.x (currently 3.1.8). Hardhat v3 is a ground-up rewrite with a fundamentally different plugin system, config format, and runtime architecture.

## Current State Inventory

### Dependencies (contracts/package.json)
| Package | Current Version | v3-Compatible Version | Status |
|---|---|---|---|
| `hardhat` | 2.28.3 | 3.1.8 | Core upgrade |
| `@nomicfoundation/hardhat-ethers` | 3.0.5 | **4.0.4** (requires HH ^3.0.7) | Has v3 version |
| `@nomicfoundation/hardhat-network-helpers` | 1.0.10 | **3.0.3** (requires HH ^3.0.0) | Has v3 version |
| `@nomicfoundation/hardhat-verify` | 2.1.1 | **3.0.10** (requires HH ^3.1.6) | Has v3 version |
| `@nomicfoundation/hardhat-toolbox` | 4.0.0 | 6.1.2 (still requires HH ^2.28) | **NOT v3 compatible** |
| `@nomicfoundation/hardhat-foundry` | 1.1.3 | 1.2.0 (still requires HH ^2.26) | **NOT v3 compatible** |
| `@openzeppelin/hardhat-upgrades` | 2.5.1 | 3.9.1 (still requires HH ^2.24) | **NOT v3 compatible** |
| `@typechain/hardhat` | 9.1.0 | 9.1.0 (requires HH ^2.9) | **NOT v3 compatible** |
| `hardhat-deploy` | 0.12.4 | 2.0.0-next.78 (requires HH ^3.1.8) | Has v3-next version |
| `hardhat-gas-reporter` | 2.3.0 | 2.3.0 (requires HH ^2.16) | **NOT v3 compatible** |
| `hardhat-storage-layout` | 0.1.7 | 0.1.7 (requires HH ^2.0.3) | **NOT v3 compatible** |
| `hardhat-tracer` | 2.8.2 | 3.4.0 (requires HH <3.x) | **NOT v3 compatible** |
| `solidity-coverage` | 0.8.17 | **Built-in** to HH v3 | Built-in |
| `solidity-docgen` | 0.6.0-beta.36 | Unknown | **Likely NOT v3 compatible** |

### Codebase Scope
- **81 test files** (`test/hardhat/`)
- **31 deploy scripts** (`deploy/`)
- **19 operational task/script files** (`scripts/operational/`)
- **~15 local deployment scripts** (`local-deployments-artifacts/`)
- **1 hardhat.config.ts** - core config file
- **Numerous helper/utility files** in `common/`, `scripts/hardhat/`, `test/hardhat/common/`

### Key Integration Patterns That Must Change
1. `import { ethers, upgrades } from "hardhat"` (25+ files)
2. `import { ethers } from "hardhat"` (48+ files)
3. `import { loadFixture } from "@nomicfoundation/hardhat-network-helpers"` (43 files)
4. `import { task } from "hardhat/config"` (11 task files)
5. `import { DeployFunction } from "hardhat-deploy/types"` (31 deploy files)
6. `import { HardhatUserConfig } from "hardhat/config"` (1 file)
7. `import { HardhatRuntimeEnvironment } from "hardhat/types"` (10+ files)
8. `from "contracts/typechain-types"` (55+ files)
9. `hre.ethers`, `hre.deployments`, `hre.getNamedAccounts()` (in tasks/deploy scripts)

## Assumptions

1. We can wait for OZ hardhat-upgrades v3 support (critical blocker if not available)
2. Foundry integration can be handled separately or with a workaround
3. The hardhat-deploy migration path (v2.0.0-next) is viable but experimental
4. We may need to replace hardhat-deploy with Hardhat Ignition as the long-term strategy
5. TypeChain will either get v3 support or we'll need Hardhat's built-in artifact types

## Key v3 Architecture Changes

### Config Format
- Network config requires explicit `type: "http"` or `type: "edr-simulated"`
- Solidity config supports build profiles
- `etherscan` config replaced with `chainDescriptors.*.blockExplorers.etherscan`
- `mocha` config section removed (testing is built-in)
- `namedAccounts` (from hardhat-deploy) gone unless using hardhat-deploy@next
- `gasReporter` config changes
- Plugins declared in `plugins` array, not via side-effect imports
- Private keys use `SensitiveString` / `ConfigurationVariable` pattern

### Plugin System
- No more `import "plugin-name"` side-effect imports
- Plugins use a hook-based architecture with `HardhatPlugin` interface
- Tasks defined via `TaskDefinition` objects, not `task()` builder function

### Testing
- Coverage and gas analytics are now built-in
- `loadFixture` moves to different import location
- Test runner integration changes

### HRE (Hardhat Runtime Environment)
- Much slimmer base HRE: only `config`, `userConfig`, `globalOptions`, `interruptions`, `versions`, `tasks`
- `ethers`, `network`, `deployments` etc. are added by plugins

---

## Migration Strategy: Phased PRs

### BLOCKER ASSESSMENT

Before starting, we must validate:
1. **`@openzeppelin/hardhat-upgrades`** - This is used extensively for proxy deployments in both deploy scripts AND test fixtures. If OZ doesn't release a v3-compatible version, we need a custom wrapper. **Check their GitHub for v3 progress.**
2. **`hardhat-deploy` vs Hardhat Ignition** - hardhat-deploy@next (2.0.0-next.78) supports HH3 but is experimental and uses `@rocketh/node`. Alternatively, migrate to Hardhat Ignition (3.0.7, HH3-native). **Decision needed.**
3. **TypeChain** - No v3 support. Hardhat v3 may have built-in artifact typing. Need to investigate `@nomicfoundation/hardhat-ethers@4` for type generation.

---

### PR 1: Pre-Migration Cleanup & Compatibility Assessment
**Goal**: Remove dead code, fix already-broken things, reduce surface area.
**Risk**: Low

Changes:
- Remove the commented-out `hardhat-tracer` import (already broken)
- Remove duplicate `@nomicfoundation/hardhat-foundry` import in `hardhat.config.ts` (line 2 and line 4 are identical)
- Audit and document which deploy scripts are actively used vs. legacy
- Document which `hardhat-deploy` features are actually used (namedAccounts, tags, dependencies)
- Create a compatibility matrix document

---

### PR 2: Decouple OpenZeppelin Upgrades Usage
**Goal**: Abstract OZ upgrades behind a wrapper so we can swap implementations.
**Risk**: Medium
**Depends on**: PR 1

Changes:
- Create an abstraction layer in `scripts/hardhat/utils.ts` and `test/hardhat/common/deployment.ts` that wraps all `upgrades.deployProxy()`, `upgrades.upgradeProxy()` calls
- All call sites already go through `deployUpgradableFromFactory()` and similar helpers, so this is mostly about making the wrapper pluggable
- Ensure the wrapper can be backed by either OZ plugin (v2 or future v3) or a manual proxy deployment

---

### PR 3: Migrate Deploy Scripts from hardhat-deploy to Hardhat Ignition (or hardhat-deploy@next)
**Goal**: Replace the `hardhat-deploy` DeployFunction pattern.
**Risk**: High (31 deploy scripts, touches deployment infrastructure)
**Depends on**: PR 2

**Option A: Hardhat Ignition** (recommended long-term)
- Convert each `DeployFunction` to an Ignition module
- Replace `namedAccounts` with explicit signer resolution
- Replace deployment tags with Ignition's dependency system
- Update `makefile-contracts.mk` deployment targets

**Option B: hardhat-deploy@next (2.0.0-next.78)**
- Upgrade hardhat-deploy to next version
- Adapt to new `@rocketh/node` dependency
- Less work but experimental dependency

Changes for either option:
- Convert all 31 `deploy/*.ts` files
- Update `common/helpers/deployments.ts`
- Update `common/helpers/verifyContract.ts` (uses `hre.run("verify")`)
- Update all `local-deployments-artifacts/*.ts` scripts
- Update `makefile-contracts.mk`
- Update `scripts/hardhat/utils.ts`

---

### PR 4: Migrate Custom Tasks to v3 Task Definitions
**Goal**: Convert all `task()` definitions to v3 `TaskDefinition` format.
**Risk**: Medium (11 task files)
**Depends on**: PR 1

Files to convert:
- `scripts/operational/tasks/getCurrentFinalizedBlockNumberTask.ts`
- `scripts/operational/tasks/grantContractRolesTask.ts`
- `scripts/operational/tasks/renounceContractRolesTask.ts`
- `scripts/operational/tasks/setRateLimitTask.ts`
- `scripts/operational/tasks/setVerifierAddressTask.ts`
- `scripts/operational/tasks/setMessageServiceOnTokenBridgeTask.ts`
- `scripts/operational/yieldBoost/addLidoStVaultYieldProvider.ts`
- `scripts/operational/yieldBoost/prepareInitiateOssification.ts`
- `scripts/operational/yieldBoost/testing/addAndClaimMessage.ts`
- `scripts/operational/yieldBoost/testing/addAndClaimMessageForLST.ts`
- `scripts/operational/yieldBoost/testing/unstakePermissionless.ts`

v2 pattern:
```typescript
import { task } from "hardhat/config";
task("myTask", "description")
  .addOptionalParam("param1")
  .setAction(async (taskArgs, hre) => { ... });
```

v3 pattern:
```typescript
import { task } from "hardhat/config";
const myTask = task("myTask", "description")
  .addOption({ name: "param1", ... })
  .setAction(async (args, hre) => { ... })
  .build();
export default myTask;
```

Also:
- Update `hre.ethers` and `hre.deployments` references inside task actions
- Remove task imports from `hardhat.config.ts` (tasks registered differently in v3)

---

### PR 5: Core Config Migration (hardhat.config.ts)
**Goal**: Rewrite `hardhat.config.ts` for v3 format.
**Risk**: High (central config, everything depends on this)
**Depends on**: PR 3, PR 4

Changes to `hardhat.config.ts`:
- Replace side-effect plugin imports with `plugins` array
- Convert `solidity.compilers` to v3 format (build profiles)
- Convert `solidity.overrides` to v3 format
- Convert `networks.hardhat` to `edr-simulated` type with `hardfork: "osaka"`
- Convert all HTTP networks to `type: "http"` format
- Replace `etherscan` config with `chainDescriptors` + `blockExplorers`
- Remove `mocha` config (now built-in or configured differently)
- Remove `namedAccounts` (replaced in PR 3)
- Handle `gasReporter` config changes
- Handle `docgen` config (may need removal if plugin incompatible)
- Update `hardhat_overrides.ts` if overrides format changes
- Convert private key handling to `ConfigurationVariable` pattern (or keep as strings)

---

### PR 6: Migrate Test Infrastructure
**Goal**: Update all 81 test files for v3 compatibility.
**Risk**: High (largest PR by file count)
**Depends on**: PR 5

Sub-tasks:
1. **Update imports**: Replace `import { ethers } from "hardhat"` with v3 equivalent via `@nomicfoundation/hardhat-ethers@4`
2. **Update `loadFixture`**: Migrate from `@nomicfoundation/hardhat-network-helpers` v1 to v3
3. **Update TypeChain imports**: Replace `from "contracts/typechain-types"` - either with new artifact type system or updated TypeChain
4. **Update `@nomicfoundation/hardhat-ethers/signers`**: `SignerWithAddress` may move or change in v4
5. **Update test deployment helpers**: `test/hardhat/common/deployment.ts` uses `ethers` and `upgrades` from hardhat
6. **Update network helpers**: `time`, `mine`, `setBalance` etc. from `@nomicfoundation/hardhat-network-helpers@3`
7. **Update `chai` assertions**: `@nomicfoundation/hardhat-chai-matchers` is NOT v3 compatible yet - may need workarounds

Could be broken into sub-PRs:
- PR 6a: Test infrastructure/helpers update
- PR 6b: Rollup tests
- PR 6c: Messaging tests
- PR 6d: Bridging/Token tests
- PR 6e: Yield tests
- PR 6f: Security/Governance/Libraries/Operational tests

---

### PR 7: Replace hardhat-toolbox Dependencies
**Goal**: Replace the monolithic `@nomicfoundation/hardhat-toolbox` with individual v3-compatible packages.
**Risk**: Medium
**Depends on**: PR 5

Since `hardhat-toolbox@6` is NOT v3 compatible, install individual packages:
- `@nomicfoundation/hardhat-ethers@4` (v3 compatible)
- `@nomicfoundation/hardhat-network-helpers@3` (v3 compatible)
- `@nomicfoundation/hardhat-verify@3` (v3 compatible)
- Built-in coverage (replaces `solidity-coverage`)
- Built-in gas analytics (replaces `hardhat-gas-reporter` partially)
- Handle TypeChain replacement

Remove:
- `@nomicfoundation/hardhat-toolbox`
- `solidity-coverage` (built-in)
- `@typechain/hardhat` (if v3 provides alternatives)
- `hardhat-storage-layout` (find alternative or remove)
- `hardhat-tracer` (already broken)
- `solidity-docgen` (handle separately)

---

### PR 8: Foundry Integration
**Goal**: Restore Foundry + Hardhat interop.
**Risk**: Medium
**Depends on**: PR 5

`@nomicfoundation/hardhat-foundry@1.2` is NOT v3 compatible. Options:
- Wait for a v3-compatible release
- Fork and port the plugin
- Use Foundry independently (the `foundry.toml` already exists and Foundry tests are in `test/foundry/`)
- Remove the Hardhat-Foundry integration if Foundry can be used standalone

---

### PR 9: Documentation & Verification Update
**Goal**: Update docgen, verification, and CI.
**Risk**: Low
**Depends on**: PR 5, PR 7

Changes:
- Update or remove `solidity-docgen` integration
- Update verification scripts to use `@nomicfoundation/hardhat-verify@3`
- Update `common/helpers/verifyContract.ts`
- Update CI/CD pipelines if any reference hardhat commands
- Update `contracts/README.md`
- Update `package.json` scripts (`build`, `test`, `coverage`, etc.)

---

### PR 10: E2E & Integration Validation
**Goal**: Ensure end-to-end deployment and test workflows function.
**Risk**: Medium
**Depends on**: All previous PRs

Changes:
- Run full test suite
- Run full local deployment via `makefile-contracts.mk`
- Test Etherscan verification flow
- Test upgrade scripts
- Validate `local-deployments-artifacts` scripts work

---

## Recommended Execution Order

```
PR 1 (Cleanup) ──────────────────────────────────────────┐
  │                                                       │
  ├── PR 2 (Abstract OZ Upgrades)                        │
  │     │                                                 │
  │     └── PR 3 (Deploy Script Migration) ──┐            │
  │                                          │            │
  ├── PR 4 (Task Migration) ─────────────────┤            │
  │                                          │            │
  └── PR 7 (Replace Toolbox) ───────────────>│            │
                                             │            │
                                       PR 5 (Core Config) │
                                             │            │
                                       PR 6 (Tests) ─────┤
                                             │            │
                                       PR 8 (Foundry) ───┤
                                             │            │
                                       PR 9 (Docs/CI) ───┤
                                                          │
                                       PR 10 (Validation) ┘
```

## Critical Blockers / Risks

1. **`@openzeppelin/hardhat-upgrades` v3 support** - Currently the biggest blocker. The entire proxy deployment infrastructure depends on this. Without it, we need a custom proxy deployment solution. OZ may release v3 support soon given HH3 is now stable.

2. **`@nomicfoundation/hardhat-chai-matchers`** - Still at v2.1.2 (requires HH ^2.26). Test assertions like `expect(...).to.be.revertedWith(...)` depend on this. Without it, tests need manual assertion rewrites.

3. **TypeChain removal** - `@typechain/hardhat` has no v3 version. 55+ test files import from `contracts/typechain-types`. May need to use Hardhat v3's built-in artifact types or `ethers.getContractFactory()` typing.

4. **`hardhat-deploy` stability** - The `2.0.0-next.78` version is pre-release. If we go the Ignition route instead, it's a bigger rewrite of 31 deploy scripts but more future-proof.

5. **Gas reporter** - `hardhat-gas-reporter` has no v3 version, but HH3 has built-in gas analytics. Need to verify feature parity.

6. **Foundry integration** - `hardhat-foundry` plugin not v3 compatible. Since Foundry tests run independently, this may be lower priority.

## Rollback Strategy

Each PR should be independently revertable. The feature branch structure allows us to test the full migration before merging to main. If critical blockers emerge (especially OZ upgrades), we can pause the migration and wait for ecosystem maturity.

## Decision Needed

**Before starting execution:**
1. Should we wait for `@openzeppelin/hardhat-upgrades` v3 support, or build a custom wrapper now?
2. Hardhat Ignition vs. `hardhat-deploy@next` for deployment scripts?
3. What is the timeline pressure - can we wait 1-3 months for ecosystem to mature?
