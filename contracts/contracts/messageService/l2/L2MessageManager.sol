// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { IL2MessageManager } from "../../interfaces/IL2MessageManager.sol";
import { PauseManager } from "../lib/PauseManager.sol";

/**
 * @title Contract to manage cross-chain message hashes storage and statuses on L2.
 * @author ConsenSys Software Inc.
 */
abstract contract L2MessageManager is Initializable, PauseManager, IL2MessageManager {
  uint8 public constant INBOX_STATUS_UNKNOWN = 0;
  uint8 public constant INBOX_STATUS_RECEIVED = 1;
  uint8 public constant INBOX_STATUS_CLAIMED = 2;

  /// @dev There is a uint232 worth of storage layout here

  bytes32 public constant L1_L2_MESSAGE_SETTER_ROLE = keccak256("L1_L2_MESSAGE_SETTER_ROLE");

  /**
   * @dev Mapping to store L1->L2 message hashes status.
   * @dev messageHash => messageStatus (0: unknown, 1: received, 2: claimed).
   */
  mapping(bytes32 => uint256) public inboxL1L2MessageStatus;

  /// @dev Keep free storage slots for future implementation updates to avoid storage collision.
  // *******************************************************************************************
  // NB: THIS GAP HAS BEEN PUSHED OUT IN FAVOUR OF THE GAP INSIDE THE REENTRANCY CODE
  //uint256[50] private __gap;
  // NB: DO NOT USE THIS GAP
  // *******************************************************************************************

  /**
   * @notice Initialises L2 message manager contract.
   * @param _l1l2MessageSetter The address owning the L1_L2_MESSAGE_SETTER_ROLE role.
   **/
  function __L2MessageManager_init(address _l1l2MessageSetter) internal onlyInitializing {
    _grantRole(L1_L2_MESSAGE_SETTER_ROLE, _l1l2MessageSetter);
  }

  /**
   * @notice Add a cross-chain L1->L2 message hashes in storage.
   * @dev Only address that has the role 'L1_L2_MESSAGE_SETTER_ROLE' are allowed to call this function.
   * @param _messageHashes Message hashes array.
   */
  function addL1L2MessageHashes(bytes32[] calldata _messageHashes) external onlyRole(L1_L2_MESSAGE_SETTER_ROLE) {
    uint256 messageHashesLength = _messageHashes.length;

    if (messageHashesLength > 100) {
      revert MessageHashesListLengthHigherThanOneHundred(messageHashesLength);
    }

    for (uint256 i; i < messageHashesLength; ) {
      bytes32 messageHash = _messageHashes[i];
      if (inboxL1L2MessageStatus[messageHash] == INBOX_STATUS_UNKNOWN) {
        inboxL1L2MessageStatus[messageHash] = INBOX_STATUS_RECEIVED;
      }
      unchecked {
        i++;
      }
    }

    emit L1L2MessageHashesAddedToInbox(_messageHashes);
  }

  /**
   * @notice Update the status of L1->L2 message when a user claim a message on L2.
   * @param _messageHash Hash of the message.
   */
  function _updateL1L2MessageStatusToClaimed(bytes32 _messageHash) internal {
    if (inboxL1L2MessageStatus[_messageHash] != INBOX_STATUS_RECEIVED) {
      revert MessageDoesNotExistOrHasAlreadyBeenClaimed();
    }

    inboxL1L2MessageStatus[_messageHash] = INBOX_STATUS_CLAIMED;
  }
}
