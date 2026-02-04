# Import Conventions

**Impact: MEDIUM (improves code clarity and prevents naming conflicts)**

Always use named imports instead of importing entire files.

## Correct: Named Imports

```solidity
import { IMessageService } from "../interfaces/IMessageService.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { IERC20, SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
```

## Incorrect: Wildcard Imports

```solidity
import "../interfaces/IMessageService.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
```

## Why Named Imports?

1. **Explicit dependencies**: Clear which symbols are used
2. **Prevents naming conflicts**: Only imports what's needed
3. **Better tooling support**: IDEs can track usage
4. **Smaller compile scope**: Compiler processes less code

## Formatting: Blank Line After Imports

Always insert a blank line between the import block and the contract/interface definition (NatSpec docstring comment or contract declaration).

### Why?

**Readability**: Makes it easier to scan file structure at a glance

### Correct

```solidity
// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.33;

import { IMessageService } from "../interfaces/IMessageService.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title MessageService
 * @notice Implementation of the message service.
 */
contract MessageService is IMessageService, Ownable {
```

### Incorrect

```solidity
// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.33;

import { IMessageService } from "../interfaces/IMessageService.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
/**
 * @title MessageService
 * @notice Implementation of the message service.
 */
contract MessageService is IMessageService, Ownable {
```

Note: The blank line should appear after the **last** import statement, before the NatSpec docstring block or contract/interface declaration.
