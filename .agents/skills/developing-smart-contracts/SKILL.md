---
name: developing-smart-contracts
description: Solidity smart contract development guidelines for Linea blockchain. Use when writing, reviewing, or refactoring Solidity contracts, or when the user asks about Solidity best practices, contract structure, or NatSpec. Covers NatSpec documentation, naming conventions, file layout, and code style.
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

// SPDX-License-Identifier: Apache-2.0 OR MIT

## Solidity Pragma

- Contracts: `0.8.33` (exact)
- Interfaces, abstract contracts, libraries: `^0.8.33` (caret)

## Rule Categories by Priority

| Priority | Category       | Impact   | Rule File                |
| -------- | -------------- | -------- | ------------------------ |
| 1        | NatSpec        | CRITICAL | `rules/natspec.md`       |
| 2        | File Layout    | HIGH     | `rules/file-layout.md`   |
| 3        | Naming         | HIGH     | `rules/naming-conventions.md` |
| 4        | Imports        | MEDIUM   | `rules/imports.md`       |
| 5        | Visibility     | MEDIUM   | `rules/visibility.md`    |
| 6        | General Rules  | MEDIUM      | `rules/general-rules.md` |

Read individual rule files for detailed explanations and code examples (correct and incorrect).

## Quick Reference

### 1. NatSpec Documentation (CRITICAL)

**ALWAYS use NatSpec for all public/external items.**

Consult [rules/natspec.md](rules/natspec.md) for NatSpec rules and examples

### 2. File Layout (HIGH)

**Interface structure:**
1. Structs → 2. Enums → 3. Events → 4. Errors → 5. External Functions

**Contract structure:**
1. Using statements → 2. Constants → 3. State variables → 4. Structs → 5. Enums → 6. Events → 7. Errors → 8. Modifiers → 9. Functions

See [rules/file-layout.md](rules/file-layout.md) for templates.

### 3. Naming Conventions (HIGH)

See [rules/naming-conventions.md](rules/naming-conventions.md) for symbol naming conventions.

### 4. Imports (MEDIUM)

Always use named imports.
Always insert a blank line after imports.

See [rules/imports.md](rules/imports.md) for details.

### 5. Visibility (MEDIUM)

See [rules/visibility.md](rules/visibility.md) for rules on applying visibility modifiers (`internal`, `external`, `public`, etc).

### 6. General Rules (MEDIUM)

See [rules/general-rules.md](rules/general-rules.md) for inheritance and general style rules

## Commit Checklist

Before making a commit, please verify:

- [ ] Rules on licenses and Solidity pragma have been applied
- [ ] All public items have NatSpec (`@notice`, `@param`, `@return`)
- [ ] All rules in `/rules/*.md` have been applied
- [ ] Linting passes (`pnpm run lint:fix`)

## Commands

For commands for testing, linting and deployment, refer to [rules/commands.md](rules/commands.md)

## Full Compiled Document

For the complete guide with all rules expanded: [AGENTS.md](AGENTS.md)
