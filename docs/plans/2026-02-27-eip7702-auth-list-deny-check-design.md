# EIP-7702 Authorization List Deny Check

## Problem

The existing `DeniedAddressValidator` only checks `tx.sender` and `tx.to` against the deny list. EIP-7702 Type 4 (DELEGATE_CODE) transactions include an `authorization_list` containing signed tuples that introduce two new actors per tuple:

1. **Authority** - the EOA that signed the tuple (recovered via implicit ecrecover)
2. **Address** - the contract whose code the authority delegates to

These actors bypass deny list enforcement through two paths:

- **Path 1 ("Puppet"):** A denylisted EOA signs an authorization tuple, and a non-denylisted sponsor submits the transaction. The denylisted EOA is only present as a signature in the authorization list.
- **Path 2 ("Parasite"):** A non-denylisted user delegates their EOA to a denylisted contract address. The denylisted contract's code executes but never appears as `tx.to`.

Path 3 ("Sleeper" - pre-delegated accounts called internally) requires execution-layer tracing and is out of scope for this change.

## Decision Summary

| Decision | Choice |
|----------|--------|
| Scope | Paths 1 & 2 only (pre-execution) |
| Placement | Extend existing `DeniedAddressValidator` inline |
| ecrecover failure handling | Skip tuple, log warning |
| Feature gating | Always-on for DELEGATE_CODE txs |
| Approach | Inline extension (~20 lines) |

## Design

### Core Logic

In `DeniedAddressValidator.validateTransaction()`, after the existing sender/recipient checks, if the transaction has a code delegation list:

1. Iterate **all** tuples in the authorization list (not just first/last - prevents hiding denied addresses in overridden tuples)
2. For each `CodeDelegation`:
   - Call `authorizer()` to get the ecrecovered authority address
   - If `Optional.empty()` (ecrecover failed), log a warning and skip to step (b)
   - If the authority is on the deny list, reject the transaction
   - Call `address()` to get the delegation target
   - If the address is on the deny list, reject the transaction
3. Error messages follow existing format: `"{role} {address} is blocked as appearing on the SDN or other legally prohibited list"`

Roles: `"authorization authority"` (for ecrecovered signer), `"authorization address"` (for delegation target).

### Besu API Surface

- `Transaction.getCodeDelegationList()` returns `Optional<List<CodeDelegation>>`
- `CodeDelegation.authorizer()` returns `Optional<Address>` (ecrecover result, memoized)
- `CodeDelegation.address()` returns `Address` (delegation target)

### Files Changed

| File | Change |
|------|--------|
| `DeniedAddressValidator.java` | Add authorization list iteration + deny check |
| `DeniedAddressValidatorTest.java` | Add 6-7 test cases for EIP-7702 scenarios |

### Test Cases

1. DELEGATE_CODE tx with no denied entries passes
2. Rejected when authority is on deny list (Path 1)
3. Rejected when delegation target address is on deny list (Path 2)
4. Rejected when both authority and address are denied (authority error first)
5. Skips tuple with unrecoverable authority, still checks address
6. Checks all tuples - denied address in second tuple caught even if first is clean
7. Non-DELEGATE_CODE tx unaffected

### What We Don't Change

- `LineaTransactionValidatorPlugin.java` - DELEGATE_CODE type gate stays as-is
- `LineaTransactionPoolValidatorFactory.java` - no new validators to wire
- `LineaTransactionPoolValidatorConfiguration.java` - no new config fields
- Deny list file format and reload mechanism unchanged

## Out of Scope

- Path 3 ("Sleeper" bypass) - requires EVM tracer integration
- New CLI flags
- Bundle-specific deny list overrides for authorization lists
