# File Layout

**Impact: HIGH (ensures consistent structure across all contracts)**

## Pragma Version

Choose **one** pragma style and use it consistently across all contracts in the project:

- **Exact version**: `pragma solidity 0.8.33;` — pins to a specific compiler version
- **Caret version**: `pragma solidity ^0.8.33;` — allows compatible patch updates

**Rule**: Once a style is chosen for a project, all contracts MUST use the same style. Do not mix exact and caret versions.

```solidity
// Consistent: all files use caret
pragma solidity ^0.8.33;

// Inconsistent (WRONG): mixing styles across files
// File A: pragma solidity 0.8.33;
// File B: pragma solidity ^0.8.33;
```

## Interface Structure

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
  struct MessageInfo {
    address sender;
    uint256 timestamp;
  }

  // 2. Enums
  enum Status {
    Pending,
    Completed,
    Failed
  }

  // 3. Events (with NatSpec)
  /**
   * @notice Emitted when a message is sent.
   * @param sender The message sender.
   * @param messageHash The hash of the message.
   */
  event MessageSent(address indexed sender, bytes32 messageHash);

  // 4. Errors (with NatSpec explaining when thrown)
  /**
   * @notice Thrown when the caller is not authorized.
   */
  error UnauthorizedCaller();

  // 5. External Functions
  /**
   * @notice Sends a message.
   * @param _to The recipient address.
   */
  function sendMessage(address _to) external;
}
```

## Contract Structure

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
  bytes32 public constant DEFAULT_ADMIN_ROLE = keccak256("DEFAULT_ADMIN_ROLE");
  uint256 internal constant MAX_MESSAGES = 1000;

  // 3. State variables
  uint256 public messageCount;
  mapping(uint256 id => bytes32 hash) public messageHashes;

  // 4. Structs (if not in interface)

  // 5. Enums (if not in interface)

  // 6. Events (with NatSpec) - if not in interface

  // 7. Errors (with NatSpec) - if not in interface

  // 8. Modifiers
  modifier onlyAdmin() {
    // check admin
    _;
  }

  // 9. Functions (public, external, internal, private)
  function sendMessage(address _to) external override {
    // implementation
  }

  function _validateMessage(bytes calldata _data) internal pure returns (bool) {
    // internal helper
  }
}
```

## Using for Library Extensions

The `using A for B` directive attaches library functions to a type. Place these directives immediately after the contract declaration, before any other declarations.

### Why?

1. **Visibility**: Library extensions are immediately visible when reading the contract
2. **Consistency**: Establishes a predictable location for finding type extensions
3. **Scope clarity**: Makes it clear which types have extended functionality

### Correct

```solidity
contract PlonkVerifierForDataAggregation is IPlonkVerifier {
  using Mimc for *;

  // Constants come after using statements
  uint256 public constant VERSION = 1;
```

```solidity
contract TokenVault is ITokenVault {
  using SafeERC20 for IERC20;
  using Address for address;

  // State variables come after using statements
  mapping(address => uint256) public balances;
```

### Incorrect

```solidity
contract TokenVault is ITokenVault {
  uint256 public constant VERSION = 1;
  
  using SafeERC20 for IERC20;  // WRONG: using should be first
  
  mapping(address => uint256) public balances;
```

## Header Template

Every file should have this header structure:

```solidity
// SPDX-License-Identifier: [Apache-2.0 for interfaces | AGPL-3.0 for contracts]
pragma solidity ^0.8.33;

// Named imports
import { Dependency } from "./path/to/Dependency.sol";

/**
 * @title Brief, one-line description.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
```
