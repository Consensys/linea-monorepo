# Ethers to Viem Migration Design - Contracts Package

## Problem Statement

The `contracts` package currently uses ethers.js v6 for all testing, deployment scripts, and blockchain interactions. The goal is to migrate to viem to align with modern practices and potentially share code with the existing `sdk-viem` package.

## Current State Analysis

### Dependencies in `contracts/package.json`
```json
{
  "@nomicfoundation/hardhat-ethers": "3.0.5",
  "@nomicfoundation/hardhat-toolbox": "4.0.0",
  "@typechain/hardhat": "9.1.0",
  "ethers": "6.14.3"
}
```

### Ethers Usage Scope
- **55+ test files** using ethers directly
- **25+ script files** for deployment and operations
- **Key patterns across ~200+ ethers.* calls**

### Usage Categories

| Category | Ethers Pattern | Count | Complexity |
|----------|---------------|-------|------------|
| Contract Deployment | `ethers.getContractFactory()` | High | Medium |
| Provider/Signers | `ethers.provider`, `SignerWithAddress` | High | Medium |
| ABI Encoding | `solidityPacked`, `AbiCoder.encode` | Medium | Low |
| Hashing | `keccak256`, `id`, `hashMessage` | Medium | Low |
| Utilities | `parseEther`, `hexlify`, `ZeroHash` | High | Low |
| Blob Transactions | `Transaction.from({ type: 3, kzg, blobs })` | Low | High |
| Events | `interface.parseLog`, `queryFilter` | Medium | Medium |
| Typechain Types | `SignerWithAddress`, Contract types | High | High |

---

## Options Considered

### Option A: Full Migration with hardhat-viem

**Approach:** Replace `@nomicfoundation/hardhat-ethers` with `@nomicfoundation/hardhat-viem` and migrate all code.

**Pros:**
- Clean, modern codebase
- Better TypeScript inference with viem
- Aligns with sdk-viem patterns
- Smaller bundle sizes for scripts

**Cons:**
- Large upfront effort
- OpenZeppelin hardhat-upgrades still requires ethers (blocker)
- Learning curve for team

**Estimated Files:** 80+ files to modify

### Option B: Hybrid Approach (Pragmatic) ✅ Recommended

**Approach:** Keep ethers for hardhat-specific integrations (deployment, upgrades) but use viem for:
- Test utilities (encoding, hashing)
- New test code
- Shared helpers

**Pros:**
- Incremental migration possible
- OpenZeppelin upgrades continue to work
- Lower risk

**Cons:**
- Two libraries to maintain
- Potential confusion
- Increased bundle size

### Option C: Keep Ethers, Extract Shared Utilities

**Approach:** Keep ethers in contracts, but extract common utilities to a shared package that can be used by both ethers and viem SDKs.

**Pros:**
- Minimal changes
- No migration risk
- Focus effort on sdk-viem instead

**Cons:**
- Doesn't achieve the goal
- Still two patterns in monorepo

---

## Decision: Option A or B Depending on Appetite

### If Full Migration (Option A) is Desired

The `@openzeppelin/hardhat-upgrades` dependency can be removed by implementing **manual proxy deployments**. This is more work upfront but enables a complete ethers-free codebase.

#### Manual Proxy Deployment Pattern (viem)

```typescript
import { 
  getContract, 
  deployContract, 
  encodeDeployData,
  encodeFunctionData,
  type Address,
  type Hex 
} from 'viem';
import { getWalletClients, getPublicClient } from '@nomicfoundation/hardhat-viem/types';

// ABIs for OpenZeppelin contracts (import from @openzeppelin/contracts)
import TransparentUpgradeableProxyABI from '@openzeppelin/contracts/build/contracts/TransparentUpgradeableProxy.json';
import ProxyAdminABI from '@openzeppelin/contracts/build/contracts/ProxyAdmin.json';

async function deployUpgradeableContract(
  implementationBytecode: Hex,
  implementationAbi: any,
  initializerArgs: any[],
  initializerFn: string = 'initialize'
): Promise<{ proxy: Address; implementation: Address; proxyAdmin: Address }> {
  const publicClient = await hre.viem.getPublicClient();
  const [deployer] = await hre.viem.getWalletClients();

  // 1. Deploy implementation
  const implementationHash = await deployer.deployContract({
    abi: implementationAbi,
    bytecode: implementationBytecode,
  });
  const implementationReceipt = await publicClient.waitForTransactionReceipt({ hash: implementationHash });
  const implementation = implementationReceipt.contractAddress!;

  // 2. Encode initializer call
  const initData = encodeFunctionData({
    abi: implementationAbi,
    functionName: initializerFn,
    args: initializerArgs,
  });

  // 3. Deploy TransparentUpgradeableProxy
  const proxyHash = await deployer.deployContract({
    abi: TransparentUpgradeableProxyABI.abi,
    bytecode: TransparentUpgradeableProxyABI.bytecode as Hex,
    args: [implementation, deployer.account.address, initData],
  });
  const proxyReceipt = await publicClient.waitForTransactionReceipt({ hash: proxyHash });
  const proxy = proxyReceipt.contractAddress!;

  // 4. Get ProxyAdmin address (created by TransparentUpgradeableProxy)
  const proxyAdmin = await publicClient.readContract({
    address: proxy,
    abi: [{ 
      name: 'admin', 
      type: 'function', 
      inputs: [], 
      outputs: [{ type: 'address' }],
      stateMutability: 'view'
    }],
    functionName: 'admin',
  });

  return { proxy, implementation, proxyAdmin };
}

async function upgradeProxy(
  proxyAddress: Address,
  proxyAdminAddress: Address,
  newImplementationBytecode: Hex,
  newImplementationAbi: any,
  reinitializerFn?: string,
  reinitializerArgs?: any[]
): Promise<Address> {
  const publicClient = await hre.viem.getPublicClient();
  const [deployer] = await hre.viem.getWalletClients();

  // 1. Deploy new implementation
  const newImplHash = await deployer.deployContract({
    abi: newImplementationAbi,
    bytecode: newImplementationBytecode,
  });
  const newImplReceipt = await publicClient.waitForTransactionReceipt({ hash: newImplHash });
  const newImplementation = newImplReceipt.contractAddress!;

  // 2. Upgrade via ProxyAdmin
  if (reinitializerFn && reinitializerArgs) {
    const reinitData = encodeFunctionData({
      abi: newImplementationAbi,
      functionName: reinitializerFn,
      args: reinitializerArgs,
    });
    
    await deployer.writeContract({
      address: proxyAdminAddress,
      abi: ProxyAdminABI.abi,
      functionName: 'upgradeAndCall',
      args: [proxyAddress, newImplementation, reinitData],
    });
  } else {
    await deployer.writeContract({
      address: proxyAdminAddress,
      abi: ProxyAdminABI.abi,
      functionName: 'upgrade',
      args: [proxyAddress, newImplementation],
    });
  }

  return newImplementation;
}
```

#### Trade-offs of Manual Proxy Deployment

| Aspect | OZ hardhat-upgrades | Manual viem |
|--------|---------------------|-------------|
| Safety checks (storage layout) | Automatic | Must implement or skip |
| Convenience | High | Lower |
| Flexibility | Limited by plugin | Full control |
| ethers dependency | Required | None |
| Code complexity | Low | Medium |

**Recommendation:** If the goal is full viem migration, manual proxy deployment is viable. The `unsafeAllow` flags already used in tests suggest storage layout validation isn't critical for this codebase.

### If Pragmatic Approach (Option B) is Preferred

Keep ethers for the ~15 files using `upgrades.deployProxy`/`upgrades.upgradeProxy` and migrate everything else to viem.

---

## Implementation Plan

### Phase 1: Foundation (Low Risk)

#### 1.1 Add viem as a dependency

Update `contracts/package.json`:
```json
{
  "devDependencies": {
    "@nomicfoundation/hardhat-viem": "3.0.2",
    "viem": "^2.22.0"
  }
}
```

#### 1.2 Create viem-based utility wrappers

New file: `contracts/test/hardhat/common/helpers/viem-utils.ts`

```typescript
import {
  keccak256,
  encodeAbiParameters,
  encodePacked,
  toHex,
  hexToBytes,
  bytesToHex,
  parseEther,
  parseGwei,
  zeroAddress,
  zeroHash,
  type Address,
  type Hash,
  type Hex,
} from 'viem';

// Encoding helpers
export const encodeData = (types: readonly string[], values: readonly unknown[], packed?: boolean): Hex => {
  if (packed) {
    return encodePacked(types as any, values as any);
  }
  return encodeAbiParameters(
    types.map((type, i) => ({ type, name: `arg${i}` })) as any,
    values as any
  );
};

// Hashing helpers  
export const generateKeccak256 = (types: string[], values: unknown[], packed?: boolean): Hash =>
  keccak256(encodeData(types, values, packed));

export const generateKeccak256Hash = (str: string): Hash => 
  generateKeccak256(['string'], [str], true);

// Constants
export const HASH_ZERO = zeroHash;
export const ADDRESS_ZERO = zeroAddress;

// Random data generation
export const generateRandomBytes = (length: number): Hex => {
  const bytes = crypto.getRandomValues(new Uint8Array(length));
  return bytesToHex(bytes);
};

// String conversion
export function convertStringToPaddedHexBytes(strVal: string, paddedSize: number): Hex {
  if (strVal.length > paddedSize) {
    throw new Error("Length is longer than padded size!");
  }
  const encoder = new TextEncoder();
  const strBytes = encoder.encode(strVal);
  const paddedBytes = new Uint8Array(paddedSize);
  paddedBytes.set(strBytes);
  return bytesToHex(paddedBytes);
}
```

---

### Phase 2: Test Utilities Migration

#### 2.1 Migrate encoding.ts

```typescript
// Before (ethers)
import { ethers, AbiCoder } from "ethers";
export const encodeData = (types: string[], values: unknown[], packed?: boolean) => {
  if (packed) return ethers.solidityPacked(types, values);
  return AbiCoder.defaultAbiCoder().encode(types, values);
};

// After (viem)
import { encodePacked, encodeAbiParameters, type Hex } from 'viem';
export const encodeData = (types: readonly string[], values: readonly unknown[], packed?: boolean): Hex => {
  if (packed) return encodePacked(types as any, values as any);
  return encodeAbiParameters(types.map((t, i) => ({ type: t, name: `v${i}` })) as any, values as any);
};
```

#### 2.2 Migrate hashing.ts

```typescript
// Before (ethers)
import { ethers } from "ethers";
export const generateKeccak256 = (types: string[], values: unknown[], packed?: boolean) =>
  ethers.keccak256(encodeData(types, values, packed));

// After (viem)
import { keccak256 } from 'viem';
export const generateKeccak256 = (types: string[], values: unknown[], packed?: boolean) =>
  keccak256(encodeData(types, values, packed));
```

#### 2.3 Migrate general.ts

```typescript
// Before
import { ethers } from "ethers";
export const generateRandomBytes = (length: number): string => ethers.hexlify(ethers.randomBytes(length));

// After
import { bytesToHex } from 'viem';
export const generateRandomBytes = (length: number): `0x${string}` => {
  const bytes = crypto.getRandomValues(new Uint8Array(length));
  return bytesToHex(bytes);
};
```

---

### Phase 3: Constants Migration

#### 3.1 Migrate `contracts/test/hardhat/common/constants/general.ts`

```typescript
// Before
import { ethers } from "hardhat";
export const HASH_ZERO = ethers.ZeroHash;
export const ADDRESS_ZERO = ethers.ZeroAddress;
export const ONE_GWEI = ethers.parseUnits("1", "gwei");
export const ONE_ETHER = ethers.parseEther("1");

// After (can keep hardhat import for provider access only)
import { zeroHash, zeroAddress, parseGwei, parseEther } from 'viem';
export const HASH_ZERO = zeroHash;
export const ADDRESS_ZERO = zeroAddress;
export const ONE_GWEI = parseGwei('1');
export const ONE_ETHER = parseEther('1');
```

---

### Phase 4: Blob Transaction Support (Complex)

Blob transactions require special handling. Viem has native blob support:

```typescript
// Before (ethers)
import { Transaction } from "ethers";
import * as kzg from "c-kzg";

const transaction = Transaction.from({
  type: 3,
  kzg,
  maxFeePerBlobGas: 1n,
  blobs: compressedBlobs,
  // ...
});
const signedTx = await wallet.signTransaction(transaction);
const txResponse = await ethers.provider.broadcastTransaction(signedTx);

// After (viem) - using walletClient
import { createWalletClient, custom } from 'viem';
import { mainnet } from 'viem/chains';
import { toBlobs, toBlobSidecars } from 'viem';

const client = createWalletClient({
  chain: mainnet,
  transport: custom(hre.network.provider),
});

const blobSidecars = toBlobSidecars({ blobs: compressedBlobs, kzg });
const hash = await client.sendTransaction({
  account,
  to: lineaRollupAddress,
  data: encodedCall,
  blobs: blobSidecars,
  maxFeePerBlobGas: 1n,
});
```

---

### Phase 5: Contract Deployment & Interactions (Two Paths)

#### Path A: Full viem (Manual Proxies)

```typescript
// Contract deployment with hardhat-viem
import hre from 'hardhat';

// Get clients
const publicClient = await hre.viem.getPublicClient();
const [deployer, operator] = await hre.viem.getWalletClients();

// Deploy contract
const contract = await hre.viem.deployContract('MyContract', [arg1, arg2]);

// Or deploy with explicit bytecode
const hash = await deployer.deployContract({
  abi: MyContractABI,
  bytecode: MyContractBytecode,
  args: [arg1, arg2],
});

// Get contract instance
const myContract = await hre.viem.getContractAt('MyContract', contractAddress);

// Interact as different account
await myContract.write.someMethod([args], { account: operator.account });
```

#### Path B: Keep ethers for deployment (Hybrid)

```typescript
// Keep ethers for deployment - works well with typechain
const factory = await ethers.getContractFactory("MyContract");
const contract = await factory.deploy(...args);

// Keep ethers for upgrades - avoids reimplementing safety checks
import { upgrades } from "hardhat";
await upgrades.deployProxy(factory, args, opts);
await upgrades.upgradeProxy(proxy, newFactory);

// Keep ethers signers for contract interactions
const [admin, operator] = await ethers.getSigners();
await contract.connect(operator).someMethod();
```

**Note:** Path B files (~15 files using upgrades) can be migrated later if manual proxy deployment proves stable.

---

### Migration Order (by file type)

#### Hybrid Approach (Path B)

| Priority | Files | Approach |
|----------|-------|----------|
| 1 | `common/helpers/encoding.ts` | Full viem |
| 2 | `common/helpers/hashing.ts` | Full viem |
| 3 | `common/helpers/general.ts` | Full viem |
| 4 | `common/constants/*.ts` | Full viem |
| 5 | `common/helpers/dataGeneration.ts` | Partial viem (encoding only) |
| 6 | Test files (`*.ts`) | Keep ethers for contract/signer, viem for utils |
| 7 | `common/deployment.ts` | Keep ethers |
| 8 | `scripts/*.ts` | Keep ethers (deployment focus) |

#### Full Migration (Path A)

| Priority | Files | Approach |
|----------|-------|----------|
| 1-5 | Same as above | Full viem |
| 6 | `common/deployment.ts` | Rewrite with manual proxy deployment |
| 7 | Test files using `upgrades.*` | Use manual proxy helpers |
| 8 | All remaining test files | Full viem with hardhat-viem |
| 9 | `scripts/*.ts` | Full viem |
| 10 | Remove ethers dependency | Clean up package.json |

---

## API Mapping Reference

| Ethers | Viem | Notes |
|--------|------|-------|
| `ethers.keccak256(data)` | `keccak256(data)` | Direct |
| `ethers.solidityPacked(types, values)` | `encodePacked(types, values)` | Direct |
| `AbiCoder.encode(types, values)` | `encodeAbiParameters(params, values)` | Requires param objects |
| `ethers.parseEther('1')` | `parseEther('1')` | Direct |
| `ethers.parseUnits('1', 'gwei')` | `parseGwei('1')` | Direct |
| `ethers.ZeroHash` | `zeroHash` | Direct |
| `ethers.ZeroAddress` | `zeroAddress` | Direct |
| `ethers.hexlify(bytes)` | `bytesToHex(bytes)` | Direct |
| `ethers.randomBytes(n)` | `crypto.getRandomValues(new Uint8Array(n))` | Native |
| `ethers.id(str)` | `keccak256(toBytes(str))` | Two steps |
| `ethers.toUtf8Bytes(str)` | `new TextEncoder().encode(str)` | Native |
| `ethers.zeroPadBytes(bytes, len)` | `pad(bytes, { size: len })` | Direct |
| `ethers.decodeBase64(str)` | `hexToBytes(atob(str))` | Native + viem |
| `Transaction.from({ type: 3, ... })` | `toBlobSidecars({ blobs, kzg })` | Different pattern |

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| OpenZeppelin upgrades breaks | Manual proxy deployment OR keep ethers for upgrades only |
| Storage layout validation lost | Review OZ plugin checks, implement critical ones manually if needed |
| Type incompatibilities | Create adapter types where needed |
| Typechain types not generated | Use `@wagmi/cli` or `viem`'s `getContract` with ABIs directly |
| Runtime behavior differences | Comprehensive test coverage |
| Team unfamiliarity | Documentation + pair programming |
| Performance regression | Benchmark critical paths |

### Typechain Consideration

Currently using `@typechain/hardhat` with `ethers-v6` target. For full viem migration:

**Option 1:** Use ABIs directly (viem's preferred pattern)
```typescript
// Instead of typechain-generated types
import { getContract } from 'viem';
import { lineaRollupAbi } from './abis/LineaRollup';

const contract = getContract({
  address: '0x...',
  abi: lineaRollupAbi,
  client: { public: publicClient, wallet: walletClient },
});
```

**Option 2:** Use `@wagmi/cli` for type generation
```typescript
// wagmi.config.ts
import { defineConfig } from '@wagmi/cli';
import { hardhat } from '@wagmi/cli/plugins';

export default defineConfig({
  out: 'src/generated.ts',
  plugins: [hardhat({ project: '.' })],
});
```

**Option 3:** Keep typechain for ethers in hybrid approach

---

## Definition of Done

- [ ] All helper utilities migrated to viem
- [ ] All constants use viem types
- [ ] All tests pass
- [ ] No runtime errors
- [ ] TypeScript types correctly inferred
- [ ] Documentation updated
- [ ] Performance benchmarks show no regression

---

## Rollback Plan

If issues arise:
1. Git revert the migration PR
2. Keep ethers-only version on a branch
3. Address issues incrementally

---

## Open Questions

1. Should we maintain backward compatibility exports from utilities?
2. Do we need to update CI to test both ethers and viem patterns?
3. Should we create a shared utility package for the monorepo?

---

## References

- [viem documentation](https://viem.sh)
- [@nomicfoundation/hardhat-viem](https://hardhat.org/hardhat-runner/docs/advanced/using-viem)
- [Existing sdk-viem in monorepo](/sdk/sdk-viem)
- [viem blob transaction support](https://viem.sh/docs/guides/blob-transactions)
