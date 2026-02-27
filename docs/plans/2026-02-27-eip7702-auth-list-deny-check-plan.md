# EIP-7702 Authorization List Deny Check - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend `DeniedAddressValidator` to check EIP-7702 authorization list `authority` and `address` fields against the deny list, preventing Puppet and Parasite bypass attacks.

**Architecture:** Inline extension of the existing `DeniedAddressValidator.validateTransaction()` method. After the sender/recipient deny list checks, iterate the transaction's code delegation list and check each tuple's ecrecovered authority and delegation target address. TDD approach - write failing tests first, then implement.

**Tech Stack:** Java 21, Besu plugin-api (Transaction, CodeDelegation interfaces), JUnit 5, AssertJ, Mockito

---

### Task 1: Write failing tests for authorization list deny check

**Files:**
- Modify: `besu-plugins/linea-sequencer/sequencer/src/test/java/net/consensys/linea/sequencer/txpoolvalidation/validators/DeniedAddressValidatorTest.java`

**Context:**
- The existing test class uses `Transaction.builder()` from `org.hyperledger.besu.ethereum.core.Transaction` (not the plugin-api interface)
- `CodeDelegation` is an interface in `org.hyperledger.besu.datatypes` with methods: `authorizer()` (returns `Optional<Address>`), `address()` (returns `Address`), `chainId()`, `nonce()`, `signature()`, `v()`, `r()`, `s()`
- The `Transaction.builder()` may or may not support setting code delegations directly. If not, we'll need to mock the Transaction or use Besu's core `CodeDelegation` implementation.
- Besu core `Transaction.builder()` should have a method like `.codeDelegations(List<CodeDelegation>)` for type DELEGATE_CODE transactions.

**Step 1: Investigate how to construct DELEGATE_CODE transactions in tests**

Check the Besu core `Transaction.Builder` class for code delegation support. Look for `codeDelegation` or `setCodeTransactionPayloads` in `tmp/besu/ethereum/core/src/main/java/org/hyperledger/besu/ethereum/core/Transaction.java`. Also check existing EIP-7702 tests in `tmp/besu` for patterns of constructing test transactions with code delegations.

If direct builder support exists, use it. If not, mock the `Transaction` interface using Mockito (the validator receives `org.hyperledger.besu.datatypes.Transaction` which is an interface).

Run: `grep -r "codeDelegat" tmp/besu/ethereum/core/src/main/java/org/hyperledger/besu/ethereum/core/Transaction.java | head -20`

**Step 2: Write the 7 failing test methods**

Add these test methods to `DeniedAddressValidatorTest.java`. The existing tests use `Transaction.builder()` with real objects. For the new tests, since we need to control `CodeDelegation.authorizer()` return values (including `Optional.empty()` for the ecrecover failure case), use Mockito mocks for both `Transaction` and `CodeDelegation`.

Add imports:
```java
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;
import java.util.List;
import org.hyperledger.besu.datatypes.CodeDelegation;
```

Tests to add:

```java
@Test
void delegateCodeTxPassesWhenNoAuthEntriesOnDenyList() {
    final CodeDelegation delegation = mock(CodeDelegation.class);
    when(delegation.authorizer()).thenReturn(Optional.of(NOT_DENIED));
    when(delegation.address()).thenReturn(NOT_DENIED);

    final org.hyperledger.besu.datatypes.Transaction transaction =
        mock(org.hyperledger.besu.datatypes.Transaction.class);
    when(transaction.getSender()).thenReturn(NOT_DENIED);
    when(transaction.getTo()).thenReturn(Optional.of(NOT_DENIED));
    when(transaction.getCodeDelegationList()).thenReturn(Optional.of(List.of(delegation)));

    assertThat(validator.validateTransaction(transaction, false, false)).isEmpty();
}

@Test
void deniedIfAuthorizationAuthorityOnDenyList() {
    final CodeDelegation delegation = mock(CodeDelegation.class);
    when(delegation.authorizer()).thenReturn(Optional.of(DENIED));
    when(delegation.address()).thenReturn(NOT_DENIED);

    final org.hyperledger.besu.datatypes.Transaction transaction =
        mock(org.hyperledger.besu.datatypes.Transaction.class);
    when(transaction.getSender()).thenReturn(NOT_DENIED);
    when(transaction.getTo()).thenReturn(Optional.of(NOT_DENIED));
    when(transaction.getCodeDelegationList()).thenReturn(Optional.of(List.of(delegation)));

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result).isPresent();
    assertThat(result.get()).contains("authorization authority").contains(DENIED.toHexString());
}

@Test
void deniedIfAuthorizationAddressOnDenyList() {
    final CodeDelegation delegation = mock(CodeDelegation.class);
    when(delegation.authorizer()).thenReturn(Optional.of(NOT_DENIED));
    when(delegation.address()).thenReturn(DENIED);

    final org.hyperledger.besu.datatypes.Transaction transaction =
        mock(org.hyperledger.besu.datatypes.Transaction.class);
    when(transaction.getSender()).thenReturn(NOT_DENIED);
    when(transaction.getTo()).thenReturn(Optional.of(NOT_DENIED));
    when(transaction.getCodeDelegationList()).thenReturn(Optional.of(List.of(delegation)));

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result).isPresent();
    assertThat(result.get()).contains("authorization address").contains(DENIED.toHexString());
}

@Test
void deniedAuthorityBeforeAddress() {
    final CodeDelegation delegation = mock(CodeDelegation.class);
    when(delegation.authorizer()).thenReturn(Optional.of(DENIED));
    when(delegation.address()).thenReturn(DENIED);

    final org.hyperledger.besu.datatypes.Transaction transaction =
        mock(org.hyperledger.besu.datatypes.Transaction.class);
    when(transaction.getSender()).thenReturn(NOT_DENIED);
    when(transaction.getTo()).thenReturn(Optional.of(NOT_DENIED));
    when(transaction.getCodeDelegationList()).thenReturn(Optional.of(List.of(delegation)));

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result).isPresent();
    assertThat(result.get()).contains("authorization authority");
}

@Test
void skipsUnrecoverableAuthorityButStillChecksAddress() {
    final CodeDelegation delegation = mock(CodeDelegation.class);
    when(delegation.authorizer()).thenReturn(Optional.empty());
    when(delegation.address()).thenReturn(DENIED);

    final org.hyperledger.besu.datatypes.Transaction transaction =
        mock(org.hyperledger.besu.datatypes.Transaction.class);
    when(transaction.getSender()).thenReturn(NOT_DENIED);
    when(transaction.getTo()).thenReturn(Optional.of(NOT_DENIED));
    when(transaction.getCodeDelegationList()).thenReturn(Optional.of(List.of(delegation)));

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result).isPresent();
    assertThat(result.get()).contains("authorization address").contains(DENIED.toHexString());
}

@Test
void checksAllTuplesNotJustFirst() {
    final CodeDelegation cleanDelegation = mock(CodeDelegation.class);
    when(cleanDelegation.authorizer()).thenReturn(Optional.of(NOT_DENIED));
    when(cleanDelegation.address()).thenReturn(NOT_DENIED);

    final CodeDelegation deniedDelegation = mock(CodeDelegation.class);
    when(deniedDelegation.authorizer()).thenReturn(Optional.of(DENIED));
    when(deniedDelegation.address()).thenReturn(NOT_DENIED);

    final org.hyperledger.besu.datatypes.Transaction transaction =
        mock(org.hyperledger.besu.datatypes.Transaction.class);
    when(transaction.getSender()).thenReturn(NOT_DENIED);
    when(transaction.getTo()).thenReturn(Optional.of(NOT_DENIED));
    when(transaction.getCodeDelegationList())
        .thenReturn(Optional.of(List.of(cleanDelegation, deniedDelegation)));

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result).isPresent();
    assertThat(result.get()).contains("authorization authority").contains(DENIED.toHexString());
}

@Test
void nonDelegateCodeTxUnaffected() {
    final Transaction transaction =
        Transaction.builder()
            .sender(NOT_DENIED)
            .to(NOT_DENIED)
            .gasPrice(Wei.ZERO)
            .payload(Bytes.EMPTY)
            .build();

    assertThat(validator.validateTransaction(transaction, false, false)).isEmpty();
}
```

**Step 3: Verify tests fail**

Run: `./gradlew :besu-plugins:linea-sequencer:sequencer:test --tests "net.consensys.linea.sequencer.txpoolvalidation.validators.DeniedAddressValidatorTest" -i`

Expected: The 6 new tests that check authorization list behavior should fail (the logic doesn't exist yet). The `nonDelegateCodeTxUnaffected` test should pass (existing behavior). Compilation may fail first if Mockito is not in the test dependencies - check `build.gradle` for mockito.

**Step 4: Commit failing tests**

```bash
git add besu-plugins/linea-sequencer/sequencer/src/test/java/net/consensys/linea/sequencer/txpoolvalidation/validators/DeniedAddressValidatorTest.java
git commit -m "test: add failing tests for EIP-7702 authorization list deny check"
```

---

### Task 2: Implement authorization list deny check

**Files:**
- Modify: `besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/validators/DeniedAddressValidator.java`

**Step 1: Add imports**

Add to `DeniedAddressValidator.java`:
```java
import java.util.List;
import org.hyperledger.besu.datatypes.CodeDelegation;
```

**Step 2: Add authorization list check method**

Add a new private method to the class:
```java
private Optional<String> checkCodeDelegationList(
    final List<CodeDelegation> codeDelegations) {
  for (final CodeDelegation delegation : codeDelegations) {
    final Optional<Address> maybeAuthority = delegation.authorizer();
    if (maybeAuthority.isEmpty()) {
      log.warn("Could not recover authority from code delegation targeting {}", delegation.address());
    } else {
      final Optional<String> authorityResult =
          checkDenied(maybeAuthority.get(), "authorization authority");
      if (authorityResult.isPresent()) {
        return authorityResult;
      }
    }

    final Optional<String> addressResult =
        checkDenied(delegation.address(), "authorization address");
    if (addressResult.isPresent()) {
      return addressResult;
    }
  }
  return Optional.empty();
}
```

**Step 3: Wire into validateTransaction**

Update `validateTransaction()` to call the new method after sender/recipient checks:

```java
@Override
public Optional<String> validateTransaction(
    final Transaction transaction, final boolean isLocal, final boolean hasPriority) {
  return checkDenied(transaction.getSender(), "sender")
      .or(() -> transaction.getTo().flatMap(to -> checkDenied(to, "recipient")))
      .or(
          () ->
              transaction
                  .getCodeDelegationList()
                  .flatMap(this::checkCodeDelegationList));
}
```

**Step 4: Run all tests**

Run: `./gradlew :besu-plugins:linea-sequencer:sequencer:test --tests "net.consensys.linea.sequencer.txpoolvalidation.validators.DeniedAddressValidatorTest" -i`

Expected: All tests pass (existing + new).

**Step 5: Run full test suite**

Run: `./gradlew :besu-plugins:linea-sequencer:sequencer:test -i`

Expected: All tests pass. No regressions.

**Step 6: Run lint**

Run: `./gradlew spotlessCheck`

Expected: No formatting violations. If violations exist, run `./gradlew spotlessApply` and include in commit.

**Step 7: Commit**

```bash
git add besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/validators/DeniedAddressValidator.java
git commit -m "feat: check EIP-7702 authorization list against deny list

Extends DeniedAddressValidator to inspect each CodeDelegation tuple's
authority (ecrecovered signer) and address (delegation target) against
the deny list, preventing Puppet and Parasite bypass attacks."
```
