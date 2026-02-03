# Linea Smart Contracts - Development Guidelines

**Version 1.0.0**  
Engineering

> **Note:**  
> This document is mainly for agents and LLMs to follow when maintaining,  
> generating, or refactoring Linea Solidity smart contracts. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Comprehensive development guidelines for Linea blockchain smart contracts, designed for AI agents and LLMs. Contains rules across 7 categories, prioritized by impact from critical (gas optimization) to medium (general rules). Each rule includes detailed explanations, real-world examples comparing incorrect vs. correct implementations.

Reference: [Linea Contract Style Guide](contracts/docs/contract-style-guide.md)

---

## Table of Contents

1. [Licenses](#1-licenses)
2. [Solidity Pragma](#2-solidity-pragma)
3. [Gas Optimization](#3-gas-optimization) — **CRITICAL**
4. [NatSpec Documentation](#4-natspec-documentation) — **HIGH**
5. [File Layout](#5-file-layout) — **HIGH**
6. [Naming Conventions](#6-naming-conventions) — **HIGH**
7. [Imports](#7-imports) — **MEDIUM**
8. [Visibility](#8-visibility) — **MEDIUM**
9. [General Rules](#9-general-rules) — **MEDIUM**
10. [Commit Checklist](#10-commit-checklist)

---

## 1. Licenses

`// SPDX-License-Identifier: Apache-2.0 OR MIT`

---

## 2. Solidity Pragma

- Contracts: `0.8.33` (exact)
- Interfaces, abstract contracts, libraries: `^0.8.33` (caret)

---

## 3. Gas Optimization

**Impact: CRITICAL**

Gas efficiency is critical for Linea contracts. Apply these rules unless there is a documented safety, audit, or readability reason to deviate.

### Calldata and External Functions

- Use `external` for functions that accept large arrays or structs
- Use `calldata` for read-only dynamic inputs in external functions

```solidity
// Correct: external + calldata
function submit(bytes32[] calldata _proofs) external {
  _verify(_proofs);
}

// Incorrect: public + memory
function submit(bytes32[] memory _proofs) public {
  _verify(_proofs);
}
```

### Minimize Storage Reads and Writes

- Cache storage values used multiple times
- Avoid redundant writes (write only when value changes)

```solidity
// Correct: cache storage read
uint256 current = fee;
if (current == 0) revert FeeNotSet();
_charge(current);

// Incorrect: repeated storage reads
if (fee == 0) revert FeeNotSet();
_charge(fee);
```

### Memory Usage

- Avoid copying `calldata` to memory unless required
- Use storage pointers when updating multiple fields in a struct

```solidity
// Correct: storage pointer for multiple updates
User storage user = users[_id];
user.balance += _amount;
user.lastUpdated = block.timestamp;
```

### Use Custom Errors

Custom errors are cheaper than revert strings.

```solidity
// Correct
error Unauthorized();
if (msg.sender != owner) revert Unauthorized();

// Incorrect
require(msg.sender == owner, "Unauthorized");
```

### Tight Loops and Unchecked Math

Use `unchecked` only when you can prove arithmetic cannot overflow.

```solidity
for (uint256 i; i < items.length; ) {
  _process(items[i]);
  unchecked { ++i; } // i < items.length
}
```

### Short-Circuit Expensive Checks

Order checks to fail early and avoid unnecessary work.

```solidity
// Correct: cheap check first
if (_to == address(0)) revert ZeroAddressNotAllowed();
if (!_isEligible(_to)) revert NotEligible();

// Incorrect: expensive check first
if (!_isEligible(_to)) revert NotEligible();  // reads storage
if (_to == address(0)) revert ZeroAddressNotAllowed();  // cheap comparison
```

### Avoid Unbounded Work

Prefer batching with explicit limits.

```solidity
// Correct: explicit batch limit
function process(uint256[] calldata _ids) external {
  if (_ids.length > MAX_BATCH_SIZE) revert BatchTooLarge();
  // ...
}

// Incorrect: no limit on batch size
function process(uint256[] calldata _ids) external {
  for (uint256 i; i < _ids.length; ) {
    // unbounded loop can exceed block gas limit
  }
}
```

---

## 4. NatSpec Documentation

**Impact: HIGH**

**ALWAYS use NatSpec for all public/external items.** This is critical for:
- Consumer documentation via interfaces
- Block explorer documentation

### Requirements

- Every public/external function MUST have `@notice`
- Every function parameter MUST have `@param _paramName` (in the same order as the function signature)
- Every return value MUST have `@return variableName`
- Events MUST document all parameters (in order)
- Errors MUST explain when they are thrown
- Use `DEPRECATED` in NatSpec for deprecated items

### Example

**Correct:**

```solidity
/**
 * @notice Sends a message to L2.
 * @param _to The recipient address on L2.
 * @param _fee The fee amount in wei.
 * @param _calldata The message calldata.
 * @return messageHash The hash of the sent message.
 */
function sendMessage(
  address _to,
  uint256 _fee,
  bytes calldata _calldata
) external payable returns (bytes32 messageHash);
```

**Incorrect: Missing NatSpec**

```solidity
function sendMessage(
  address _to,
  uint256 _fee,
  bytes calldata _calldata
) external payable returns (bytes32);
```

---

## 5. File Layout

**Impact: HIGH**

### Interface Structure

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.33;

import { ImportType } from "../ImportType.sol";

/**
 * @title Brief description.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface ISampleContract {
  // 1. Structs
  // 2. Enums
  // 3. Events (with NatSpec)
  // 4. Errors (with NatSpec explaining when thrown)
  // 5. External Functions
}
```

### Contract Structure

```solidity
// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.33;

import { ISampleContract } from "./interfaces/ISampleContract.sol";
import { SafeERC20, IERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title Brief description.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract SampleContract is ISampleContract {
  // 1. Using statements (for library extensions)
  using SafeERC20 for IERC20;

  // 2. Constants (public, internal, private)
  // 3. State variables
  // 4. Structs
  // 5. Enums
  // 6. Events (with NatSpec)
  // 7. Errors (with NatSpec)
  // 8. Modifiers
  // 9. Functions (public, external, internal, private)
}
```

---

## 6. Naming Conventions

**Impact: HIGH**

| Item                   | Convention       | Example                              |
| ---------------------- | ---------------- | ------------------------------------ |
| Public state           | camelCase        | `uint256 messageCount`               |
| Private/internal state | _camelCase       | `uint256 _internalCounter`           |
| Constants              | UPPER_SNAKE_CASE | `bytes32 DEFAULT_ADMIN_ROLE`         |
| Function params        | _camelCase       | `function send(address _to)`         |
| Return variables       | camelCase (named)| `returns (bytes32 messageHash)`      |
| Mappings               | descriptive keys | `mapping(uint256 id => bytes32 hash)`|
| Init functions         | __Contract_init  | `__PauseManager_init()`              |

---

## 7. Imports

**Impact: MEDIUM**

Always use named imports:

**Correct:**

```solidity
import { IMessageService } from "../interfaces/IMessageService.sol";
```

**Incorrect:**

```solidity
import "../interfaces/IMessageService.sol";
```

### Blank Line After Imports

Always add a blank line between imports and the contract definition:

```solidity
import { IMessageService } from "../interfaces/IMessageService.sol";

/**
 * @title MessageService
 */
contract MessageService is IMessageService {
```

---

## 8. Visibility

**Impact: MEDIUM**

- **Constants**: `internal` unless explicitly needed public
- **Functions**: Minimize `external`/`public` surface area
- **Avoid**: `this.functionCall()` pattern - refactor instead

---

## 9. General Rules

**Impact: MEDIUM**

### OpenZeppelin Dependencies

Use OpenZeppelin contracts version **4.9.6**:

```json
"@openzeppelin/contracts": "4.9.6",
"@openzeppelin/contracts-upgradeable": "4.9.6"
```

### Inheritance & Customization

- Use `virtual`/`override` keywords
- Override `CONTRACT_VERSION()` for custom versions
- **Note**: Any modifications from audited code should be independently audited

### Avoid Magic Numbers

Use named constants instead of hardcoded values.

### Assembly

Use hex for memory offsets (`mstore(add(mPtr, 0x20), _var)`).

### Linting

Run `pnpm run lint:fix` before committing.

---

## 10. Commit Checklist

Before making a commit, please verify:

- [ ] Rules on licenses and Solidity pragma have been applied
- [ ] All public items have NatSpec (`@notice`, `@param`, `@return`)
- [ ] All rules in `/rules/*.md` have been applied
- [ ] Linting passes (`pnpm run lint:fix`)

---

## References

1. [Linea Contract Style Guide](contracts/docs/contract-style-guide.md)
2. [Linea Deployment Guide](contracts/docs/deployment.md)
3. [Solidity Style Guide](https://docs.soliditylang.org/en/latest/style-guide.html)
