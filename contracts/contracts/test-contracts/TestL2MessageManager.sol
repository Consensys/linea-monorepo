// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { L2MessageManager } from "../messageService/l2/L2MessageManager.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";

contract TestL2MessageManager is Initializable, L2MessageManager, IGenericErrors {
  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  function initialize(address _pauserManager, address _l1l2MessageSetter) public initializer {
    if (_pauserManager == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    if (_l1l2MessageSetter == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    __ERC165_init();
    __Context_init();
    __AccessControl_init();
    __L2MessageManager_init(_l1l2MessageSetter);

    _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
    _grantRole(PAUSE_MANAGER_ROLE, _pauserManager);
  }

  function tryInitialize(address _l1l2MessageSetter) external {
    __L2MessageManager_init(_l1l2MessageSetter);
  }

  function updateL1L2MessageStatusToClaimed(bytes32 _messageHash) external {
    _updateL1L2MessageStatusToClaimed(_messageHash);
  }
}
