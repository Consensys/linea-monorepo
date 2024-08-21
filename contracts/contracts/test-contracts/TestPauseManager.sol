// SPDX-License-Identifier: AGPL-3.0
pragma solidity >=0.8.19 <=0.8.26;

import { PauseManager } from "../messageService/lib/PauseManager.sol";

contract TestPauseManager is PauseManager {
  function initialize() public initializer {
    __AccessControl_init();
    _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
  }
}
