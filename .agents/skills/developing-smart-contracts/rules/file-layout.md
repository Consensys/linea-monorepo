# File Layout

**Impact: HIGH (ensures consistent structure across all contracts)**

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

/**
 * @title Brief description.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract SampleContract is ISampleContract {
  // 1. Constants (public, internal, private)
  bytes32 public constant DEFAULT_ADMIN_ROLE = keccak256("DEFAULT_ADMIN_ROLE");
  uint256 internal constant MAX_MESSAGES = 1000;

  // 2. State variables
  uint256 public messageCount;
  mapping(uint256 id => bytes32 hash) public messageHashes;

  // 3. Structs (if not in interface)

  // 4. Enums (if not in interface)

  // 5. Events (with NatSpec) - if not in interface

  // 6. Errors (with NatSpec) - if not in interface

  // 7. Modifiers
  modifier onlyAdmin() {
    // check admin
    _;
  }

  // 8. Functions (public, external, internal, private)
  function sendMessage(address _to) external override {
    // implementation
  }

  function _validateMessage(bytes calldata _data) internal pure returns (bool) {
    // internal helper
  }
}
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
