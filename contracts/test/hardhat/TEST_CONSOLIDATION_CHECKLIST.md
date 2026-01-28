# Test Consolidation Execution Checklist

This document tracks the execution of the test consolidation plan for the Linea Rollup contracts test suite.

---

## 1. Pre-Consolidation Checklist

### 1.1 Baseline Metrics

- [ ] **Record current test counts**
  - [ ] Run `npx hardhat test --grep "Linea Rollup"` and record total tests
  - [ ] Run `npx hardhat test --grep "Validium"` and record total tests
  - [ ] Run `npx hardhat test --grep "BlobSubmission"` and record total tests
  - [ ] Run `npx hardhat test --grep "Finalization"` and record total tests
  - [ ] Document total test count: `____`

- [ ] **Record current test execution time**
  - [ ] Run full test suite with timing: `time npx hardhat test`
  - [ ] Document baseline execution time: `____`

- [ ] **Record current code coverage**
  - [ ] Run `npx hardhat coverage` (if configured)
  - [ ] Document coverage percentages for key contracts:
    - [ ] LineaRollup.sol: `____%`
    - [ ] Validium.sol: `____%`
    - [ ] LineaRollupBase.sol: `____%`

- [ ] **Document current linting status**
  - [ ] Run `npm run lint` and record any existing issues
  - [ ] Document count of existing lint errors: `____`

### 1.2 Files to Backup/Branch

- [ ] **Create feature branch**
  - [ ] Branch name: `feat/test-consolidation`
  - [ ] Base branch: Current working branch

- [ ] **Key files to track changes**
  - [ ] `contracts/test/hardhat/rollup/LineaRollup.ts` (1118 lines)
  - [ ] `contracts/test/hardhat/rollup/Validium.ts` (580 lines)
  - [ ] `contracts/test/hardhat/rollup/LineaRollup/BlobSubmission.ts` (900 lines)
  - [ ] `contracts/test/hardhat/rollup/LineaRollup/Finalization.ts` (779 lines)
  - [ ] `contracts/test/hardhat/common/helpers/index.ts`

### 1.3 CI/CD Considerations

- [ ] **Verify CI pipeline status**
  - [ ] Current CI is passing on base branch
  - [ ] Understand test timeouts in CI configuration

- [ ] **Plan for incremental PRs**
  - [ ] Decide on PR strategy: single large PR vs. multiple small PRs
  - [ ] Recommended: One PR per helper migration phase

---

## 2. Implementation Checklist - Helper Files

### 2.1 Access Control Helper (`accessControl.ts`)

**Status**: ✅ Created

**Functions Available**:
- `expectHasRole` - Verify account has specific role
- `expectDoesNotHaveRole` - Verify account lacks role
- `expectHasRoles` - Batch role verification
- `grantRoles` - Grant multiple roles in parallel
- `revokeRoles` - Revoke multiple roles in parallel
- `expectAccessControlRevert` - Expect revert due to missing role
- `testRoleBasedAccess` - Test authorized vs unauthorized access
- `testRoleGrantAndRevoke` - Test grant/revoke cycle
- `expectRoleAdmin` - Verify admin role configuration

**Migration Tasks**:

- [ ] **Identify usage patterns in test files**
  - [ ] Search for `hasRole` assertions
  - [ ] Search for `buildAccessErrorMessage` calls
  - [ ] Search for role grant/revoke test patterns

- [ ] **Replace inline patterns**
  - [ ] Replace `expect(await contract.hasRole(role, account)).to.be.true` with `expectHasRole`
  - [ ] Replace `expectRevertWithReason(call, buildAccessErrorMessage(...))` with `expectAccessControlRevert`

- [ ] **Verification**
  - [ ] Run affected tests after changes
  - [ ] Verify no test regressions

---

### 2.2 Initialization Helper (`initialization.ts`)

**Status**: ✅ Created

**Functions Available**:
- `expectZeroAddressRevert` - Test zero address rejection
- `expectZeroValueRevert` - Test zero value rejection
- `expectZeroHashRevert` - Test zero hash rejection
- `expectZeroLengthRevert` - Test zero length rejection
- `expectDoubleInitRevert` - Test re-initialization rejection
- `expectZeroValueReverts` - Batch zero value tests
- `validateInitializedState` - Validate state after init
- `validateNonZeroAddress/Hash/Value` - Individual validations

**Migration Tasks**:

- [ ] **Identify usage patterns in test files**
  - [ ] Search for `ZeroAddressNotAllowed` error expectations
  - [ ] Search for `Initializable: contract is already initialized` assertions
  - [ ] Search for initialization validation patterns

- [ ] **Replace inline patterns**
  - [ ] Replace manual zero address revert checks with `expectZeroAddressRevert`
  - [ ] Replace double-init checks with `expectDoubleInitRevert`
  - [ ] Group related zero-value tests using `expectZeroValueReverts`

- [ ] **Verification**
  - [ ] Run affected tests after changes
  - [ ] Verify no test regressions

---

### 2.3 Pause Manager Helper (`pauseManager.ts`)

**Status**: ✅ Created

**Functions Available**:
- `expectPaused` / `expectNotPaused` - Verify pause state
- `pauseAndVerify` / `unpauseAndVerify` - Pause with verification
- `expectPauseEvent` / `expectUnpauseEvent` - Event assertions
- `expectUnpauseExpiryEvent` - Expiry event assertion
- `expectRevertWhenPaused` / `expectRevertWhenNotPaused` - Pause-related reverts
- `expectPauseRoleRevert` / `expectUnpauseRoleRevert` - Role-based reverts
- `expectPauseTypeNotUsedRevert` - Invalid pause type revert
- `expectCooldownRevert` - Cooldown enforcement revert
- `expectPauseNotExpiredRevert` - Expiry check revert
- `testActionFailsWhenPaused` - Comprehensive pause test
- `testPauseUnpauseCycle` - Full cycle test
- `getExpectedCooldownEnd` - Calculate cooldown timestamp

**Migration Tasks**:

- [ ] **Identify usage patterns in test files**
  - [ ] Search for `isPaused` assertions
  - [ ] Search for `IsPaused` error expectations
  - [ ] Search for `pauseByType` / `unPauseByType` patterns
  - [ ] Search for pause event assertions

- [ ] **Replace inline patterns**
  - [ ] Replace `expect(await contract.isPaused(type)).to.be.true` with `expectPaused`
  - [ ] Replace pause event assertions with `expectPauseEvent` / `expectUnpauseEvent`
  - [ ] Replace `IsPaused` error checks with `expectRevertWhenPaused`

- [ ] **Verification**
  - [ ] Run affected tests after changes
  - [ ] Verify no test regressions

---

### 2.4 Rolling Hash Helper (`rollingHash.ts`)

**Status**: ✅ Created

**Functions Available**:
- `computeRollingHash` - Compute single rolling hash
- `computeRollingHashChain` - Compute chain with intermediates
- `computeFinalRollingHash` - Compute final hash from collection
- `generateMessageHash` - Create test message hash
- `validateRollingHashStorage` - Verify stored hash value
- `validateRollingHashNotZero` / `validateRollingHashIsZero` - Zero checks
- `expectRollingHashUpdatedEvent` - Event assertion
- `expectRollingHashMismatchRevert` - Mismatch error
- `computeAndValidateRollingHash` - Compute and verify in one call
- `createMessageHashSequence` - Generate test hash sequences
- `validateRollingHashChain` - Validate entire chain

**Migration Tasks**:

- [ ] **Identify usage patterns in test files**
  - [ ] Search for `calculateRollingHash` calls
  - [ ] Search for `rollingHashes` storage reads
  - [ ] Search for `RollingHashMismatch` error expectations
  - [ ] Search for `RollingHashUpdated` event assertions

- [ ] **Replace inline patterns**
  - [ ] Use `computeRollingHash` for single hash calculations
  - [ ] Use `validateRollingHashStorage` for storage verification
  - [ ] Use `expectRollingHashMismatchRevert` for mismatch tests

- [ ] **Verification**
  - [ ] Run affected tests after changes
  - [ ] Verify no test regressions

---

## 3. File-by-File Migration Checklist

### 3.1 LineaRollup.ts (Main)

**File**: `contracts/test/hardhat/rollup/LineaRollup.ts`
**Lines**: ~1118

**Applicable Helpers**:
- [x] Access Control - Role verification tests
- [x] Initialization - Zero address/value validation
- [x] Pause Manager - Pause/unpause tests
- [ ] Rolling Hash - (limited usage in main file)

**Migration Tasks**:

- [ ] **Initialisation section** (lines ~113-350)
  - [ ] Replace `expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed")` 
        with `expectZeroAddressRevert({ contract: lineaRollup, deployOrInitCall: deployCall })`
  - [ ] Replace `expectRevertWithReason(reinitCall, INITIALIZED_ALREADY_MESSAGE)` 
        with `expectDoubleInitRevert({ contract: lineaRollup, initCall: reinitCall })`
  - [ ] Replace inline role verification with `expectHasRole`
  - [ ] Estimated test count affected: 15-20 tests

- [ ] **Role-based access tests** (scattered throughout)
  - [ ] Replace `expect(await lineaRollup.hasRole(ROLE, account)).to.be.true` 
        with `expectHasRole(lineaRollup, ROLE, account)`
  - [ ] Replace access error assertions with `expectAccessControlRevert`
  - [ ] Estimated test count affected: 10-15 tests

- [ ] **Pause functionality tests** (lines ~400-500)
  - [ ] Replace manual pause state checks with `expectPaused` / `expectNotPaused`
  - [ ] Replace `IsPaused` error checks with `expectRevertWhenPaused`
  - [ ] Estimated test count affected: 5-10 tests

- [ ] **Keep separate (do not migrate)**:
  - [ ] Fallback/Receive tests - simple, no helper benefit
  - [ ] Complex submission logic tests - domain-specific
  - [ ] Verifier management tests - unique patterns

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/rollup/LineaRollup.ts`
- [ ] Verify all tests pass
- [ ] Compare test count before/after

---

### 3.2 Validium.ts

**File**: `contracts/test/hardhat/rollup/Validium.ts`
**Lines**: ~580

**Applicable Helpers**:
- [x] Access Control - Role verification tests
- [x] Initialization - Zero address validation
- [x] Pause Manager - Pause/unpause tests

**Migration Tasks**:

- [ ] **Initialisation section** (lines ~84-200)
  - [ ] Replace zero address revert checks with `expectZeroAddressRevert`
  - [ ] Replace role verification with `expectHasRole`
  - [ ] Group zero-value tests with `expectZeroValueReverts`
  - [ ] Estimated test count affected: 10-15 tests

- [ ] **Role-based access tests**
  - [ ] Identify and replace all `hasRole` assertions
  - [ ] Replace access error messages with `expectAccessControlRevert`
  - [ ] Estimated test count affected: 5-10 tests

- [ ] **Pause functionality tests**
  - [ ] Replace pause state assertions with helpers
  - [ ] Replace pause-related revert checks
  - [ ] Estimated test count affected: 5-8 tests

- [ ] **Keep separate (do not migrate)**:
  - [ ] Validium-specific submission tests
  - [ ] External storage tests

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/rollup/Validium.ts`
- [ ] Verify all tests pass
- [ ] Compare test count before/after

---

### 3.3 BlobSubmission.ts

**File**: `contracts/test/hardhat/rollup/LineaRollup/BlobSubmission.ts`
**Lines**: ~900

**Applicable Helpers**:
- [x] Access Control - Operator role checks
- [x] Pause Manager - Submission pause tests
- [ ] Initialization - (not applicable)
- [ ] Rolling Hash - (limited in blob context)

**Migration Tasks**:

- [ ] **Operator role tests**
  - [ ] Replace `buildAccessErrorMessage(nonAuthorizedAccount, OPERATOR_ROLE)` patterns
        with `expectAccessControlRevert`
  - [ ] Estimated test count affected: 3-5 tests

- [ ] **Pause-related tests**
  - [ ] Replace `STATE_DATA_SUBMISSION_PAUSE_TYPE` pause checks
  - [ ] Use `expectRevertWhenPaused` for paused submission tests
  - [ ] Estimated test count affected: 2-4 tests

- [ ] **Keep separate (do not migrate)**:
  - [ ] Blob construction and KZG tests - highly specialized
  - [ ] Shnarf calculation tests - domain-specific
  - [ ] Multi-blob submission tests - complex orchestration

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/rollup/LineaRollup/BlobSubmission.ts`
- [ ] Verify all tests pass
- [ ] Compare test count before/after

---

### 3.4 Finalization.ts

**File**: `contracts/test/hardhat/rollup/LineaRollup/Finalization.ts`
**Lines**: ~779

**Applicable Helpers**:
- [x] Access Control - Operator role checks
- [x] Pause Manager - Finalization pause tests
- [x] Rolling Hash - L1 rolling hash verification

**Migration Tasks**:

- [ ] **Rolling hash validation tests** (extensive)
  - [ ] Replace inline rolling hash calculations with `computeRollingHash`
  - [ ] Replace rolling hash storage checks with `validateRollingHashStorage`
  - [ ] Use `expectRollingHashMismatchRevert` for mismatch tests
  - [ ] Estimated test count affected: 10-15 tests

- [ ] **Operator role tests**
  - [ ] Replace access error assertions with `expectAccessControlRevert`
  - [ ] Estimated test count affected: 2-3 tests

- [ ] **Pause-related tests**
  - [ ] Replace `FINALIZATION_PAUSE_TYPE` checks
  - [ ] Use pause helper functions
  - [ ] Estimated test count affected: 2-4 tests

- [ ] **Keep separate (do not migrate)**:
  - [ ] Proof verification tests - complex ZK logic
  - [ ] State transition validation - domain-specific
  - [ ] Multi-proof finalization - orchestration tests

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/rollup/LineaRollup/Finalization.ts`
- [ ] Verify all tests pass
- [ ] Compare test count before/after

---

### 3.5 L1MessageService.ts

**File**: `contracts/test/hardhat/messaging/l1/L1MessageService.ts`

**Applicable Helpers**:
- [x] Rolling Hash - Message hash and rolling hash tests
- [x] Access Control - Role-based message claiming
- [x] Pause Manager - Messaging pause tests

**Migration Tasks**:

- [ ] **Rolling hash tests**
  - [ ] Identify all `calculateRollingHash` usages
  - [ ] Replace with `computeRollingHash` / `computeRollingHashChain`
  - [ ] Use `validateRollingHashStorage` for state checks
  - [ ] Estimated test count affected: 5-10 tests

- [ ] **Access control tests**
  - [ ] Replace role verification patterns
  - [ ] Estimated test count affected: 3-5 tests

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/messaging/l1/L1MessageService.ts`
- [ ] Verify all tests pass

---

### 3.6 L2MessageService.ts

**File**: `contracts/test/hardhat/messaging/l2/L2MessageService.ts`

**Applicable Helpers**:
- [x] Rolling Hash - L2 message rolling hashes
- [x] Access Control - Anchoring permissions
- [x] Pause Manager - L2 messaging pause

**Migration Tasks**:

- [ ] **Rolling hash tests**
  - [ ] Replace inline hash computations
  - [ ] Use `expectRollingHashUpdatedEvent` for event tests
  - [ ] Estimated test count affected: 5-8 tests

- [ ] **Access and pause tests**
  - [ ] Apply standard helper patterns
  - [ ] Estimated test count affected: 3-5 tests

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/messaging/l2/L2MessageService.ts`
- [ ] Verify all tests pass

---

### 3.7 TokenBridge.ts

**File**: `contracts/test/hardhat/bridging/token/TokenBridge.ts`

**Applicable Helpers**:
- [x] Pause Manager - Bridge pause functionality
- [x] Access Control - Bridge admin roles
- [ ] Initialization - Bridge initialization

**Migration Tasks**:

- [ ] **Pause functionality tests**
  - [ ] Replace bridge pause assertions with helpers
  - [ ] Estimated test count affected: 3-5 tests

- [ ] **Role tests**
  - [ ] Apply access control helpers
  - [ ] Estimated test count affected: 2-4 tests

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/bridging/token/TokenBridge.ts`
- [ ] Verify all tests pass

---

### 3.8 PauseManager.ts (Security)

**File**: `contracts/test/hardhat/security/PauseManager.ts`

**Applicable Helpers**:
- [x] Pause Manager - Primary consumer of pause helpers

**Migration Tasks**:

- [ ] **This file should heavily use pause helpers**
  - [ ] Replace all inline pause assertions
  - [ ] Use `testPauseUnpauseCycle` for cycle tests
  - [ ] Use `getExpectedCooldownEnd` for timing tests
  - [ ] Estimated test count affected: 15-25 tests

**Verification**:
- [ ] Run `npx hardhat test test/hardhat/security/PauseManager.ts`
- [ ] Verify all tests pass

---

## 4. Post-Consolidation Checklist

### 4.1 Regression Testing

- [ ] **Run full test suite**
  - [ ] `npx hardhat test`
  - [ ] All tests should pass
  - [ ] Document any failures and investigate

- [ ] **Compare metrics to baseline**
  - [ ] Test count: `____` (expected: same or slightly fewer)
  - [ ] Execution time: `____` (expected: similar or improved)
  - [ ] Coverage: `____%` (expected: same or improved)

- [ ] **Run tests in CI environment**
  - [ ] Push changes and verify CI passes
  - [ ] Check for any timeout issues

### 4.2 Documentation Updates

- [ ] **Update helper index.ts**
  - [ ] Verify all new helpers are exported
  - [ ] Current exports in `common/helpers/index.ts`:
    ```typescript
    export * from "./initialization";
    export * from "./rollingHash";
    export * from "./accessControl";
    export * from "./pauseManager";
    ```

- [ ] **Add JSDoc to new helpers** (if not already present)
  - [ ] Each function should have:
    - [ ] Description
    - [ ] @param documentation
    - [ ] @example usage

- [ ] **Update any internal documentation**
  - [ ] Testing guidelines (if exists)
  - [ ] Contributing guide (if exists)

### 4.3 Cleanup of Old Code

- [ ] **Identify deprecated patterns**
  - [ ] Search for any remaining inline patterns that should use helpers
  - [ ] Document any intentionally kept inline patterns

- [ ] **Remove unused imports**
  - [ ] Run linter to identify unused imports
  - [ ] Remove any helper functions that were fully replaced

- [ ] **Verify no duplicate logic**
  - [ ] Search for patterns that exist both inline and in helpers
  - [ ] Consolidate to helper usage only

### 4.4 Final Validation

- [ ] **Code review checklist**
  - [ ] All new code follows existing style
  - [ ] No commented-out code
  - [ ] No debugging statements left behind
  - [ ] TypeScript strict mode compliance

- [ ] **Linting**
  - [ ] `npm run lint` passes
  - [ ] No new lint warnings introduced

- [ ] **Type checking**
  - [ ] `npx tsc --noEmit` passes
  - [ ] No type errors in helper files

---

## 5. Migration Order Recommendation

Execute migrations in this order to minimize risk:

### Phase 1: Low-Risk Helpers
1. [ ] Initialization helper migration
2. [ ] Access control helper migration

### Phase 2: Medium-Risk Helpers
3. [ ] Pause manager helper migration
4. [ ] Rolling hash helper migration

### Phase 3: File-by-File Migration
5. [ ] Validium.ts (smaller file, good test case)
6. [ ] LineaRollup.ts (main file)
7. [ ] BlobSubmission.ts
8. [ ] Finalization.ts
9. [ ] Messaging files
10. [ ] Remaining files

---

## 6. Rollback Plan

If issues are discovered:

1. **Revert to pre-consolidation branch**
   ```bash
   git checkout <base-branch>
   ```

2. **If partial migration completed**
   - Identify which files were successfully migrated
   - Cherry-pick successful changes only

3. **Document issues**
   - Record what went wrong
   - Update this checklist with lessons learned

---

## 7. Progress Summary

| File | Status | Tests Before | Tests After | Notes |
|------|--------|--------------|-------------|-------|
| accessControl.ts | ✅ Created | - | - | Helper ready |
| initialization.ts | ✅ Created | - | - | Helper ready |
| pauseManager.ts | ✅ Created | - | - | Helper ready |
| rollingHash.ts | ✅ Created | - | - | Helper ready |
| LineaRollup.ts | ⬜ Pending | | | |
| Validium.ts | ⬜ Pending | | | |
| BlobSubmission.ts | ⬜ Pending | | | |
| Finalization.ts | ⬜ Pending | | | |
| L1MessageService.ts | ⬜ Pending | | | |
| L2MessageService.ts | ⬜ Pending | | | |
| TokenBridge.ts | ⬜ Pending | | | |
| PauseManager.ts | ⬜ Pending | | | |

---

## 8. Completion Criteria

The consolidation is complete when:

- [ ] All identified test files have been migrated
- [ ] All tests pass locally and in CI
- [ ] Test execution time has not significantly increased
- [ ] Code coverage is maintained or improved
- [ ] All documentation is updated
- [ ] PR has been reviewed and approved
- [ ] Changes have been merged to main branch

---

*Last Updated: [DATE]*
*Author: [NAME]*
