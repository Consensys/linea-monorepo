# EIP-7702 Authorization List Deny Check - Acceptance Tests Design

## Problem

The unit tests for the EIP-7702 authorization list deny check use mocks. We need acceptance tests that verify the behavior against a real Besu node with Prague fork configuration, real crypto signatures, and the actual plugin validation pipeline.

## Decision Summary

| Decision | Choice |
|----------|--------|
| Scope | Authority deny (Path 1), address deny (Path 2), positive case, hot-reload |
| Test class | New `EIP7702AuthListDenyTest` extending `LineaPluginPoSTestBase` |
| Block import | Pool rejection only (no engine_newPayloadV4 tests) |
| Tx construction | Parameterize existing `sendRawEIP7702Transaction` pattern with separate authority/sender key pairs |

## Design

### Test Class

New file: `EIP7702AuthListDenyTest.kt` in `besu-plugins/linea-sequencer/acceptance-tests/src/test/kotlin/linea/plugin/acc/test/`

Extends `LineaPluginPoSTestBase()`. Overrides `getTestCliOptions()` to:
- Enable DELEGATE_CODE transactions: `--plugin-linea-delegate-code-tx-enabled=true`
- Set deny list path: `--plugin-linea-deny-list-path=<resource>`

### Key Pair Mapping

- `GENESIS_ACCOUNT_ONE` (`0xfe3b557e...`) = NOT on deny list (clean sender/authority)
- `GENESIS_ACCOUNT_TWO` (`0x627306090a...`) = ON deny list (denied authority/address target)

Existing `denyList.txt` already contains ACCOUNT_TWO's address.

### Helper Method

```kotlin
private fun sendRawEIP7702TransactionWithAuth(
    web3j: Web3j,
    senderCredentials: Credentials,
    authSigner: Credentials,
    delegationAddress: Address,
): EthSendTransaction
```

Constructs a `CodeDelegation` signed by `authSigner` delegating to `delegationAddress`, wrapped in a DELEGATE_CODE transaction signed by `senderCredentials`. Returns `EthSendTransaction` (does not throw) so tests can assert on error messages.

### Test Cases

1. **`eip7702TxWithDeniedAuthorityIsRejected`** (Path 1 - Puppet)
   - Sender: ACCOUNT_ONE (clean), Auth signer: ACCOUNT_TWO (denied), Address: clean
   - Assert: error contains `"authorization authority 0x627306... is blocked"`

2. **`eip7702TxWithDeniedDelegationAddressIsRejected`** (Path 2 - Parasite)
   - Sender: ACCOUNT_ONE (clean), Auth signer: ACCOUNT_ONE (clean), Address: denied address
   - Assert: error contains `"authorization address 0x627306... is blocked"`

3. **`eip7702TxWithCleanAuthListIsAccepted`** (Positive)
   - Sender: ACCOUNT_ONE (clean), Auth signer: ACCOUNT_ONE (clean), Address: clean
   - Assert: transaction accepted, no error

4. **`eip7702AuthListDenyCheckWorksAfterReload`** (Hot-reload)
   - Start with empty deny list (temp file)
   - Send tx with ACCOUNT_TWO as authority - passes
   - Write ACCOUNT_TWO to deny list, reload config
   - Send same tx pattern - rejected
   - Note: This test needs its own `getTestCliOptions()` override with a temp deny list, so it may be a separate inner test or a second test class

### Files Changed

| File | Change |
|------|--------|
| `EIP7702AuthListDenyTest.kt` | New file - 3 tests (deny authority, deny address, clean passes) |
| `EIP7702AuthListDenyReloadTest.kt` | New file - 1 test (hot-reload). Separate class because it needs different CLI options (temp deny list) |

### What We Don't Change

- `LineaPluginPoSTestBase.kt` - the helper stays in the test classes, not the base
- `sendRawEIP7702Transaction()` - existing method remains unchanged
- `denyList.txt` - reuse existing resource
- No new Solidity contracts needed
