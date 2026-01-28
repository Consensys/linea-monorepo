# Test Migration Patterns - Quick Reference

This document provides copy-paste patterns for migrating tests to use the new consolidated helpers.

---

## 1. Access Control Migrations

### Before: Inline Role Verification

```typescript
// Old pattern
expect(await contract.hasRole(OPERATOR_ROLE, operator.address)).to.be.true;
```

### After: Using Helper

```typescript
import { expectHasRole } from "../common/helpers";

// New pattern
await expectHasRole(contract, OPERATOR_ROLE, operator);
```

---

### Before: Access Error Assertion

```typescript
// Old pattern
await expectRevertWithReason(
  contract.connect(nonAuthorizedAccount).someFunction(),
  buildAccessErrorMessage(nonAuthorizedAccount, REQUIRED_ROLE)
);
```

### After: Using Helper

```typescript
import { expectAccessControlRevert } from "../common/helpers";

// New pattern
await expectAccessControlRevert(
  contract.connect(nonAuthorizedAccount).someFunction(),
  nonAuthorizedAccount,
  REQUIRED_ROLE
);
```

---

### Before: Multiple Role Checks

```typescript
// Old pattern
expect(await contract.hasRole(OPERATOR_ROLE, operator.address)).to.be.true;
expect(await contract.hasRole(ADMIN_ROLE, admin.address)).to.be.true;
expect(await contract.hasRole(PAUSER_ROLE, pauser.address)).to.be.true;
```

### After: Using Helper

```typescript
import { expectHasRoles } from "../common/helpers";

// New pattern
await expectHasRoles(contract, [
  { role: OPERATOR_ROLE, account: operator },
  { role: ADMIN_ROLE, account: admin },
  { role: PAUSER_ROLE, account: pauser },
]);
```

---

## 2. Initialization Migrations

### Before: Zero Address Revert Check

```typescript
// Old pattern
const deployCall = deployUpgradableFromFactory("ContractName", [
  { ...initData, defaultVerifier: ADDRESS_ZERO }
]);
await expectRevertWithCustomError(contract, deployCall, "ZeroAddressNotAllowed");
```

### After: Using Helper

```typescript
import { expectZeroAddressRevert } from "../common/helpers";

// New pattern
const deployCall = deployUpgradableFromFactory("ContractName", [
  { ...initData, defaultVerifier: ADDRESS_ZERO }
]);
await expectZeroAddressRevert({
  contract,
  deployOrInitCall: deployCall,
});
```

---

### Before: Double Initialization Check

```typescript
// Old pattern
const reinitCall = contract.initialize(initParams);
await expectRevertWithReason(reinitCall, "Initializable: contract is already initialized");
```

### After: Using Helper

```typescript
import { expectDoubleInitRevert } from "../common/helpers";

// New pattern
await expectDoubleInitRevert({
  contract,
  initCall: contract.initialize(initParams),
});
```

---

### Before: Multiple Zero Value Checks

```typescript
// Old pattern
it("Should revert if admin is zero", async () => {
  const call = deploy({ admin: ADDRESS_ZERO });
  await expectRevertWithCustomError(contract, call, "ZeroAddressNotAllowed");
});

it("Should revert if fee is zero", async () => {
  const call = deploy({ fee: 0n });
  await expectRevertWithCustomError(contract, call, "ZeroValueNotAllowed");
});
```

### After: Using Helper (Batch)

```typescript
import { expectZeroValueReverts, ZERO_ADDRESS } from "../common/helpers";

// New pattern - single test with multiple scenarios
it("Should revert with zero values", async () => {
  await expectZeroValueReverts({
    contract,
    scenarios: [
      {
        description: "admin address",
        deployOrInitCall: deploy({ admin: ZERO_ADDRESS }),
        expectedError: "ZeroAddressNotAllowed",
      },
      {
        description: "fee amount",
        deployOrInitCall: deploy({ fee: 0n }),
        expectedError: "ZeroValueNotAllowed",
      },
    ],
  });
});
```

---

## 3. Pause Manager Migrations

### Before: Pause State Check

```typescript
// Old pattern
expect(await contract.isPaused(GENERAL_PAUSE_TYPE)).to.be.true;
```

### After: Using Helper

```typescript
import { expectPaused } from "../common/helpers";

// New pattern
await expectPaused(contract, GENERAL_PAUSE_TYPE);
```

---

### Before: Paused Revert Check

```typescript
// Old pattern
await lineaRollup.connect(securityCouncil).pauseByType(STATE_DATA_SUBMISSION_PAUSE_TYPE);
await expectRevertWithCustomError(
  lineaRollup,
  lineaRollup.connect(operator).submitBlobs(...args),
  "IsPaused",
  [STATE_DATA_SUBMISSION_PAUSE_TYPE]
);
```

### After: Using Helper

```typescript
import { pauseAndVerify, expectRevertWhenPaused } from "../common/helpers";

// New pattern
await pauseAndVerify(
  lineaRollup.connect(securityCouncil),
  STATE_DATA_SUBMISSION_PAUSE_TYPE
);
await expectRevertWhenPaused(
  lineaRollup,
  lineaRollup.connect(operator).submitBlobs(...args),
  STATE_DATA_SUBMISSION_PAUSE_TYPE
);
```

---

### Before: Pause Event Check

```typescript
// Old pattern
await expectEvent(
  contract,
  contract.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE),
  "Paused",
  [pauser.address, GENERAL_PAUSE_TYPE]
);
```

### After: Using Helper

```typescript
import { expectPauseEvent } from "../common/helpers";

// New pattern
await expectPauseEvent(
  contract,
  contract.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE),
  pauser.address,
  GENERAL_PAUSE_TYPE
);
```

---

### Before: Full Pause/Unpause Cycle

```typescript
// Old pattern
expect(await contract.isPaused(pauseType)).to.be.false;

await expectEvent(contract, contract.pauseByType(pauseType), "Paused", [pauser, pauseType]);
expect(await contract.isPaused(pauseType)).to.be.true;

await expectEvent(contract, contract.unPauseByType(pauseType), "UnPaused", [pauser, pauseType]);
expect(await contract.isPaused(pauseType)).to.be.false;
```

### After: Using Helper

```typescript
import { testPauseUnpauseCycle } from "../common/helpers";

// New pattern
await testPauseUnpauseCycle(contract, pauseType, pauser.address);
```

---

## 4. Rolling Hash Migrations

### Before: Rolling Hash Computation

```typescript
// Old pattern
const expectedRollingHash = calculateRollingHash(ethers.ZeroHash, messageHash);
expect(await contract.rollingHashes(1n)).to.equal(expectedRollingHash);
```

### After: Using Helper

```typescript
import { computeAndValidateRollingHash, INITIAL_ROLLING_HASH } from "../common/helpers";

// New pattern
await computeAndValidateRollingHash(
  contract,
  INITIAL_ROLLING_HASH,
  messageHash,
  1n
);
```

---

### Before: Rolling Hash Chain

```typescript
// Old pattern
let rollingHash = ethers.ZeroHash;
for (const hash of messageHashes) {
  rollingHash = calculateRollingHash(rollingHash, hash);
}
const finalExpectedHash = rollingHash;
```

### After: Using Helper

```typescript
import { computeRollingHashChain, INITIAL_ROLLING_HASH } from "../common/helpers";

// New pattern
const result = computeRollingHashChain(INITIAL_ROLLING_HASH, messageHashes);
const finalExpectedHash = result.finalHash;
// Also available: result.intermediateHashes
```

---

### Before: Rolling Hash Mismatch Revert

```typescript
// Old pattern
await expectRevertWithCustomError(
  contract,
  contract.anchorL1L2MessageHashes(hashes, messageNumber, wrongRollingHash),
  "RollingHashMismatch",
  [expectedHash, wrongRollingHash]
);
```

### After: Using Helper

```typescript
import { expectRollingHashMismatchRevert } from "../common/helpers";

// New pattern
await expectRollingHashMismatchRevert({
  contract,
  call: contract.anchorL1L2MessageHashes(hashes, messageNumber, wrongRollingHash),
  errorArgs: [expectedHash, wrongRollingHash],
});
```

---

### Before: Rolling Hash Event

```typescript
// Old pattern
await expectEvent(
  contract,
  contract.sendMessage(...args),
  "RollingHashUpdated",
  [messageNumber, expectedRollingHash, messageHash]
);
```

### After: Using Helper

```typescript
import { expectRollingHashUpdatedEvent } from "../common/helpers";

// New pattern
await expectRollingHashUpdatedEvent({
  contract,
  updateCall: contract.sendMessage(...args),
  messageNumber,
  expectedRollingHash,
  messageHash,
});
```

---

## 5. Import Organization

### Recommended Import Structure

```typescript
// Test file imports

// External dependencies
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

// Contract types
import { TestLineaRollup } from "contracts/typechain-types";

// Test helpers - consolidated
import {
  // Access Control
  expectHasRole,
  expectAccessControlRevert,
  expectHasRoles,
  
  // Initialization
  expectZeroAddressRevert,
  expectDoubleInitRevert,
  ZERO_ADDRESS,
  
  // Pause Manager
  expectPaused,
  expectNotPaused,
  expectRevertWhenPaused,
  testPauseUnpauseCycle,
  
  // Rolling Hash
  computeRollingHash,
  computeRollingHashChain,
  validateRollingHashStorage,
  INITIAL_ROLLING_HASH,
  
  // Other helpers
  expectEvent,
  expectRevertWithCustomError,
  generateRandomBytes,
} from "../common/helpers";

// Test fixtures
import { deployLineaRollupFixture, getAccountsFixture } from "./helpers";

// Constants
import { OPERATOR_ROLE, GENERAL_PAUSE_TYPE } from "../common/constants";
```

---

## 6. Common Migration Mistakes

### Mistake 1: Forgetting to await

```typescript
// Wrong - missing await
expectHasRole(contract, ROLE, account);

// Correct
await expectHasRole(contract, ROLE, account);
```

### Mistake 2: Wrong account type

```typescript
// Wrong - passing address string instead of SignerWithAddress
await expectAccessControlRevert(call, "0x123...", ROLE);

// Correct - pass the signer
await expectAccessControlRevert(call, nonAuthorizedAccount, ROLE);
```

### Mistake 3: Not connecting contract for pause helpers

```typescript
// Wrong - contract not connected to authorized account
await pauseAndVerify(contract, pauseType);

// Correct - connect to account with pause role
await pauseAndVerify(contract.connect(securityCouncil), pauseType);
```

### Mistake 4: Using wrong helper for L1 vs standard rolling hashes

```typescript
// Wrong - using standard method for L1 hashes
await validateRollingHashStorage({
  contract,
  messageNumber: 1n,
  expectedHash,
  // Missing isL1RollingHash flag
});

// Correct - specify L1 context
await validateRollingHashStorage({
  contract,
  messageNumber: 1n,
  expectedHash,
  isL1RollingHash: true,
});
```

---

## 7. Testing Your Migration

After migrating a test file:

```bash
# Run specific test file
npx hardhat test test/hardhat/rollup/LineaRollup.ts

# Run with verbose output
npx hardhat test test/hardhat/rollup/LineaRollup.ts --verbose

# Run specific test by name
npx hardhat test --grep "Should revert if verifier address is zero"

# Run with coverage
npx hardhat coverage --testfiles "test/hardhat/rollup/LineaRollup.ts"
```

---

*Use this reference alongside the main TEST_CONSOLIDATION_CHECKLIST.md*
