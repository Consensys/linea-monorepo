// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { IL2MessageManagerV1 } from "../../../interfaces/l2/IL2MessageManagerV1.sol";
import { L2MessageServicePauseManager } from "../../../lib/L2MessageServicePauseManager.sol";

/**
 * @title Contract to manage cross-chain message hashes storage and statuses on L2.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract L2MessageManagerV1 is Initializable, L2MessageServicePauseManager, IL2MessageManagerV1 {
  uint8 public constant INBOX_STATUS_UNKNOWN = 0;
  uint8 public constant INBOX_STATUS_RECEIVED = 1;
  uint8 public constant INBOX_STATUS_CLAIMED = 2;

  bytes32 public constant L1_L2_MESSAGE_SETTER_ROLE = keccak256("L1_L2_MESSAGE_SETTER_ROLE");

  /**
   * @dev Mapping to store L1->L2 message hashes status.
   * @dev messageHash => messageStatus (0: unknown, 1: received, 2: claimed).
   */
  mapping(bytes32 messageHash => uint256 messageStatus) public inboxL1L2MessageStatus;

  /// @dev Keep free storage slots for future implementation updates to avoid storage collision.
  // *******************************************************************************************
  // NB: THIS GAP HAS BEEN PUSHED OUT IN FAVOUR OF THE GAP INSIDE THE REENTRANCY CODE
  //uint256[50] private __gap;
  // NB: DO NOT USE THIS GAP
  // *******************************************************************************************

  /// @dev Total contract storage is 1 slot.

  /**
   * @notice Initialises L2 message manager contract.
   * @param _l1l2MessageSetter The address owning the L1_L2_MESSAGE_SETTER_ROLE role.
   */
  function __L2MessageManager_init(address _l1l2MessageSetter) internal onlyInitializing {
    _grantRole(L1_L2_MESSAGE_SETTER_ROLE, _l1l2MessageSetter);
  }

  /**
   * @notice Update the status of L1->L2 message when a user claims a message on L2.
   * @param _messageHash Hash of the message.
   */
  function _updateL1L2MessageStatusToClaimed(bytes32 _messageHash) internal {
    if (inboxL1L2MessageStatus[_messageHash] != INBOX_STATUS_RECEIVED) {
      revert MessageDoesNotExistOrHasAlreadyBeenClaimed(_messageHash);
    }

    inboxL1L2MessageStatus[_messageHash] = INBOX_STATUS_CLAIMED;
  }
}
