# Ethers-to-Viem Migration Design -- contracts package

## Problem

The `contracts/` package uses ethers v6 (`6.14.3`) as its core EVM library, both directly and via the Hardhat ethers plugin (`@nomicfoundation/hardhat-ethers`). Viem is the modern, type-safe, lightweight alternative with better tree-shaking and first-class TypeScript support. The goal is to replace ethers with viem across the contracts package.

## Assumptions

- Hardhat remains the compilation/test framework (not migrating to Foundry for TS tooling).
- Hardhat has a first-party viem plugin: `@nomicfoundation/hardhat-viem`.
- OpenZeppelin's `hardhat-upgrades` plugin currently depends on `@nomicfoundation/hardhat-ethers`. There is beta/experimental viem support via `@openzeppelin/hardhat-upgrades/viem`.
- TypeChain is ethers-specific; the viem Hardhat plugin generates its own typed contract bindings.
- The Solidity contracts (.sol) are not affected at all -- this is a pure TS/JS tooling migration.

## Scope Inventory

### Surface area (counts are file-level, not line-level)

| Category | Files touching `ethers` | Files touching `hre.ethers` | Notes |
|---|---|---|---|
| **Test helpers** (`test/hardhat/common/`) | ~12 | ~5 | Constants, encoding, hashing, deployment, expectations |
| **Test suites** (`test/hardhat/`) | ~45 | ~40 | getContractFactory, getSigners, provider |
| **Deploy scripts** (`deploy/`) | ~20 | ~18 | getContractFactory, upgrades.deployProxy |
| **Operational scripts** (`scripts/`) | ~25 | ~8 | Mixed: some use hardhat runtime, some use ethers directly |
| **Local deployment artifacts** | ~15 | 0 | Use ethers directly (ContractFactory, Wallet, etc.) |
| **Common helpers** (`common/helpers/`) | ~5 | 0 | Pure ethers utilities (encoding, hashing, deployments) |

### Ethers APIs in use (high to low frequency)

1. **`ethers.getContractFactory` / `ethers.getSigners` / `ethers.provider`** -- via `hardhat` import (~80+ call sites). These are the Hardhat-ethers plugin APIs.
2. **`ethers.keccak256` / `ethers.solidityPacked` / `ethers.AbiCoder`** -- encoding/hashing (~60+ uses). Viem equivalents: `keccak256`, `encodePacked`, `encodeAbiParameters`.
3. **`ethers.parseEther` / `ethers.parseUnits` / `ethers.formatEther`** -- value conversions (~30 uses). Viem: `parseEther`, `parseGwei`, `formatEther`.
4. **`ethers.ZeroHash` / `ethers.ZeroAddress` / `ethers.MaxUint256`** -- constants (~15 uses). Viem: `zeroHash`, `zeroAddress`, `maxUint256`.
5. **`ethers.hexlify` / `ethers.getBytes` / `ethers.concat` / `ethers.toBeHex` / `ethers.randomBytes`** -- byte manipulation (~40 uses). Viem: `toHex`, `hexToBytes`, `concat`, `toHex(n)`, etc.
6. **`ethers.Wallet` / `ethers.HDNodeWallet` / `ethers.SigningKey`** -- key management (~10 uses). Viem: `privateKeyToAccount`, `mnemonicToAccount`.
7. **`ethers.Interface` / `ethers.ContractFactory`** -- ABI encoding (~15 uses). Viem: `encodeFunctionData`, `decodeFunctionResult`, `getContract`.
8. **`ethers.Transaction`** -- raw tx construction for blob txs (~5 uses). Viem: `serializeTransaction`.
9. **`ethers.Signature`** -- signature parsing (~2 uses). Viem: `parseSignature`.
10. **`ethers.id` / `ethers.encodeBytes32String`** -- misc (~5 uses). Viem: `keccak256(toHex(str))` / `stringToHex(str, { size: 32 })`.
11. **TypeChain types** (`contracts/typechain-types`) -- used across all test files. Will be replaced by hardhat-viem generated types.
12. **`@nomicfoundation/hardhat-ethers/signers` (`SignerWithAddress`)** -- used in ~35 files. Viem equivalent: `WalletClient` from hardhat-viem.
13. **`@openzeppelin/hardhat-upgrades`** -- `upgrades.deployProxy` / `upgrades.deployImplementation` (~25 uses). Must use `/viem` subpath.

## Options and Tradeoffs

### Option A: Big-bang migration
- **Pros**: Clean cut, no dual-dependency period, consistent codebase.
- **Cons**: Massive PR (100+ files), high review burden, risky merge conflicts, long period of broken tests during migration.

### Option B: Incremental migration with compatibility layer (RECOMMENDED)
- **Pros**: Smaller PRs, testable at each step, can ship intermediate states.
- **Cons**: Temporary dual dependencies, need a shim/adapter period.

### Option C: Thin wrapper that abstracts ethers/viem
- **Pros**: Can swap underlying lib without touching consumers.
- **Cons**: Adds indirection, defeats the purpose of viem's type safety, maintenance burden of the wrapper.

## Decision: Option B -- Incremental migration

### Phase 0: Preparation & Foundation (1 PR)
1. Add `viem` and `@nomicfoundation/hardhat-viem` as dependencies.
2. Update `hardhat.config.ts` to import `@nomicfoundation/hardhat-viem`.
3. Keep `@nomicfoundation/hardhat-ethers` in place (both plugins can coexist).
4. Verify the project compiles and tests pass with both plugins loaded.

### Phase 1: Migrate pure utility functions -- no Hardhat runtime dependency (1-2 PRs)
Target: `common/helpers/`, `test/hardhat/common/helpers/`, `test/hardhat/common/constants/`

These files use ethers as a pure library (encoding, hashing, hex manipulation). They don't depend on the Hardhat runtime ethers object. Straightforward 1:1 API swaps:

| ethers v6 | viem |
|---|---|
| `ethers.keccak256(bytes)` | `keccak256(bytes)` |
| `ethers.solidityPacked(types, values)` | `encodePacked(types, values)` |
| `AbiCoder.defaultAbiCoder().encode(types, values)` | `encodeAbiParameters(parseAbiParameters(types), values)` |
| `ethers.parseEther("1")` | `parseEther("1")` |
| `ethers.parseUnits("1", "gwei")` | `parseGwei("1")` |
| `ethers.ZeroHash` | `zeroHash` (from `viem`) |
| `ethers.ZeroAddress` | `zeroAddress` |
| `ethers.hexlify(bytes)` | `toHex(bytes)` |
| `ethers.getBytes(hex)` | `hexToBytes(hex)` |
| `ethers.concat([a, b])` | `concat([a, b])` |
| `ethers.toBeHex(n, size)` | `toHex(n, { size })` |
| `ethers.randomBytes(n)` | `crypto.getRandomValues(new Uint8Array(n))` or a small helper |
| `ethers.decodeBase64(str)` | `Buffer.from(str, 'base64')` (Node built-in) |
| `ethers.id(str)` | `keccak256(toHex(str))` |
| `ethers.toUtf8Bytes(str)` | `toBytes(str)` or `stringToBytes(str)` |
| `ethers.zeroPadBytes(b, len)` | `pad(b, { size: len })` |
| `ethers.isAddress(addr)` | `isAddress(addr)` |
| `ethers.getAddress(addr)` | `getAddress(addr)` |
| `ethers.Signature.from(sig)` | `parseSignature(sig)` |
| `ethers.Interface` | Use `encodeFunctionData` / `decodeFunctionResult` from viem |
| `ethers.ContractFactory` | No direct equivalent; use `deployContract` from hardhat-viem or viem's `walletClient.deployContract` |

### Phase 2: Migrate test infrastructure -- Hardhat runtime swap (2-3 PRs)

#### 2a: Deployment helpers (`test/hardhat/common/deployment.ts`)
Replace:
```ts
// Before
import { ethers, upgrades } from "hardhat";
const factory = await ethers.getContractFactory(name);
const contract = await factory.deploy(...args);

// After
import { viem } from "hardhat";
const contract = await viem.deployContract(name, args);
```

For upgradeable:
```ts
// Before
import { upgrades } from "hardhat";
const contract = await upgrades.deployProxy(factory, args, opts);

// After
import { upgrades } from "hardhat";
// Use the viem-compatible API from @openzeppelin/hardhat-upgrades
const contract = await upgrades.deployProxy(factory, args, opts);
// Note: this requires @openzeppelin/hardhat-upgrades with viem support
```

**Key risk**: `@openzeppelin/hardhat-upgrades` viem support maturity. Need to verify API parity for `deployProxy`, `deployImplementation`, `upgradeProxy`. If not available, keep ethers shim for OZ upgrades only.

#### 2b: Signer/account pattern
Replace `SignerWithAddress` with viem's `WalletClient`:
```ts
// Before
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
const [admin, operator] = await ethers.getSigners();

// After
const [admin, operator] = await viem.getWalletClients();
// Or for public client (read-only):
const publicClient = await viem.getPublicClient();
```

#### 2c: Contract interaction pattern
Replace TypeChain-generated contract objects with viem's `getContractAt`:
```ts
// Before (typechain)
const lineaRollup: TestLineaRollup = ...;
await lineaRollup.connect(operator).submitBlobs(data);

// After (hardhat-viem)
const lineaRollup = await viem.getContractAt("TestLineaRollup", address);
await lineaRollup.write.submitBlobs([data], { account: operator.account });
// Read:
const value = await lineaRollup.read.currentL2BlockNumber();
```

### Phase 3: Migrate test suites (5-8 PRs, by domain)
Migrate in order of increasing complexity / dependency:
1. `test/hardhat/libraries/` (Mimc, SparseMerkleProof, etc.) -- simplest
2. `test/hardhat/security/` (PauseManager, RateLimiter)
3. `test/hardhat/governance/` (Timelock)
4. `test/hardhat/messaging/` (L1/L2 MessageService, MessageManager)
5. `test/hardhat/bridging/` (TokenBridge, BridgedToken, MockedE2E)
6. `test/hardhat/operational/` (RollupRevenueVault, etc.)
7. `test/hardhat/rollup/` (LineaRollup, BlobSubmission, Finalization, Validium) -- most complex due to blob tx construction
8. `test/hardhat/yield/` (YieldManager, LidoStVault, etc.) -- large surface area

### Phase 4: Migrate deploy and operational scripts (2-3 PRs)
1. `deploy/` scripts -- swap `ethers.getContractFactory` to viem equivalents
2. `scripts/operational/` -- tasks, yield boost scripts
3. `scripts/upgrades/` -- upgrade scripts
4. `local-deployments-artifacts/` -- local deployment helpers

### Phase 5: Migrate blob transaction construction (1 PR)
The `sendBlobTransaction` pattern in `test/hardhat/rollup/helpers/blob.ts` manually constructs and signs Type 3 (EIP-4844) transactions using `ethers.Transaction`. Viem equivalent:
```ts
import { serializeTransaction } from "viem";
// Construct and sign with walletClient.sendTransaction or manual serialization
```

### Phase 6: Remove ethers dependency (1 PR)
1. Remove `ethers` from `package.json`
2. Remove `@nomicfoundation/hardhat-ethers`
3. Remove `@typechain/hardhat` (replaced by hardhat-viem bindings)
4. Remove `hardhat.config.ts` ethers plugin import
5. Delete `typechain-types/` directory
6. Clean up any remaining ethers imports

## Risks

1. **`@openzeppelin/hardhat-upgrades` viem compatibility**: If the OZ viem API is immature or missing features (e.g., `unsafeAllow`, `constructorArgs`), we may need to keep ethers as a dev dependency solely for OZ upgrades, or write a thin adapter.
   - *Mitigation*: Check OZ plugin version and viem support status before starting Phase 2. If insufficient, isolate OZ usage behind a single helper file that retains ethers.

2. **Blob transaction support**: Viem's blob tx support has different ergonomics than ethers' `Transaction.from()`. Need to verify `c-kzg` integration.
   - *Mitigation*: Spike/prototype blob tx construction with viem early (during Phase 0).

3. **Chai matchers**: `@nomicfoundation/hardhat-chai-matchers` (`.to.be.revertedWithCustomError`, `.to.emit`) works with ethers contract types. Need to verify compatibility with viem contract types via `@nomicfoundation/hardhat-toolbox-viem`.
   - *Mitigation*: Replace `@nomicfoundation/hardhat-toolbox` with `@nomicfoundation/hardhat-toolbox-viem` in Phase 0 and test a single test file.

4. **`hardhat-deploy` plugin**: Uses ethers under the hood. Need to verify it works with viem or find alternative.
   - *Mitigation*: The `deploy/` scripts are run less frequently; they can be migrated last or kept on ethers if `hardhat-deploy` doesn't support viem.

5. **Merge conflict risk**: With ~100 files to touch, parallel development on `main` will cause conflicts.
   - *Mitigation*: Incremental PRs, coordinate with team, merge to main frequently.

## Rollout

| Phase | PRs | Est. Effort | Risk |
|---|---|---|---|
| 0: Foundation | 1 | 1 day | Low |
| 1: Pure utilities | 1-2 | 2 days | Low |
| 2: Test infrastructure | 2-3 | 3 days | Medium (OZ upgrades) |
| 3: Test suites | 5-8 | 5-8 days | Medium |
| 4: Scripts & deploy | 2-3 | 2-3 days | Medium (hardhat-deploy) |
| 5: Blob txs | 1 | 1 day | Medium |
| 6: Cleanup | 1 | 0.5 day | Low |
| **Total** | **~15-19 PRs** | **~15-18 days** | |

## Rollback

Each phase is independently revertible since ethers and viem coexist. If a phase introduces test failures, revert the PR. The final cleanup phase (removing ethers) is the only irreversible step -- only execute after full green CI.
