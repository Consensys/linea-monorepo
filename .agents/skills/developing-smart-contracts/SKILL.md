---
name: developing-smart-contracts
description: Solidity smart contract development guidelines for Linea blockchain. Covers NatSpec documentation, naming conventions, file layout, and code style. Use when writing, reviewing, or refactoring Solidity contracts, or when the user asks about Solidity best practices, contract structure, or NatSpec.
license: AGPL-3.0
metadata:
  author: linea
  version: '1.0.0'
---

# Linea Smart Contract Development

Best practices for developing Solidity smart contracts on Linea blockchain. Contains
rules covering documentation, naming, structure, and code style.

Reference: [Linea Contract Style Guide](contracts/docs/contract-style-guide.md)

## When to Apply

Reference these guidelines when:

- Writing new Solidity contracts for Linea
- Reviewing smart contract code
- Refactoring existing contracts
- Adding NatSpec documentation
- Setting up contract file structure

## Licenses

- **Interfaces**: `// SPDX-License-Identifier: Apache-2.0`
- **Contracts**: `// SPDX-License-Identifier: AGPL-3.0`

## Rule Categories by Priority

| Priority | Category       | Impact   | Rule File                |
| -------- | -------------- | -------- | ------------------------ |
| 1        | NatSpec        | CRITICAL | `rules/natspec.md`       |
| 2        | File Layout    | HIGH     | `rules/file-layout.md`   |
| 3        | Naming         | HIGH     | `rules/naming-conventions.md` |
| 4        | Imports        | MEDIUM   | `rules/imports.md`       |
| 5        | Visibility     | MEDIUM   | `rules/visibility.md`    |
| 6        | General Rules  | LOW      | `rules/general-rules.md` |

## Quick Reference

### 1. NatSpec Documentation (CRITICAL)

**ALWAYS use NatSpec for all public/external items.**

- Every public/external function MUST have `@notice`
- Every parameter MUST have `@param _paramName` (in signature order)
- Every return value MUST have `@return variableName`
- Events MUST document all parameters
- Errors MUST explain when they are thrown

See [rules/natspec.md](rules/natspec.md) for examples.

### 2. File Layout (HIGH)

**Interface structure:**
1. Structs → 2. Enums → 3. Events → 4. Errors → 5. External Functions

**Contract structure:**
1. Constants → 2. State variables → 3. Structs → 4. Enums → 5. Events → 6. Errors → 7. Modifiers → 8. Functions

See [rules/file-layout.md](rules/file-layout.md) for templates.

### 3. Naming Conventions (HIGH)

| Item                    | Convention       | Example                          |
| ----------------------- | ---------------- | -------------------------------- |
| Public state            | camelCase        | `uint256 messageCount`           |
| Private/internal state  | _camelCase       | `uint256 _internalCounter`       |
| Constants               | UPPER_SNAKE_CASE | `bytes32 DEFAULT_ADMIN_ROLE`     |
| Function params         | _camelCase       | `function send(address _to)`     |
| Return variables        | camelCase        | `returns (bytes32 messageHash)`  |
| Mappings                | descriptive keys | `mapping(uint256 id => bytes32)` |
| Init functions          | __Contract_init  | `__PauseManager_init()`          |

See [rules/naming-conventions.md](rules/naming-conventions.md) for details.

### 4. Imports (MEDIUM)

Always use named imports:

```solidity
// CORRECT
import { IMessageService } from "../interfaces/IMessageService.sol";

// WRONG
import "../interfaces/IMessageService.sol";
```

See [rules/imports.md](rules/imports.md) for details.

### 5. Visibility (MEDIUM)

- Constants: `internal` unless explicitly needed public
- Functions: Minimize `external`/`public` surface area
- Avoid `this.functionCall()` pattern

See [rules/visibility.md](rules/visibility.md) for details.

### 6. General Rules (LOW)

- Avoid magic numbers: Use named constants
- Assembly: Use hex for memory offsets
- Linting: Run `pnpm run lint:fix` before committing

See [rules/general-rules.md](rules/general-rules.md) for deployment, testing, and more.

## PR Checklist

Before submitting a PR:

- [ ] All public items have NatSpec (`@notice`, `@param`, `@return`)
- [ ] Named imports used
- [ ] Naming conventions followed
- [ ] Constants are `internal` unless needed public
- [ ] Linting passes (`pnpm run lint:fix`)
- [ ] No magic numbers

## How to Use

Read individual rule files for detailed explanations and code examples:

```
rules/natspec.md
rules/file-layout.md
rules/naming-conventions.md
```

Each rule file contains:

- Explanation of why the rule matters
- Incorrect code example
- Correct code example

## Full Compiled Document

For the complete guide with all rules expanded: [AGENTS.md](AGENTS.md)
