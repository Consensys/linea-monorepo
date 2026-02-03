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
