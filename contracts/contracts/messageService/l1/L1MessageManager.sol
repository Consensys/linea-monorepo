// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { IL1MessageManager } from "../../interfaces/IL1MessageManager.sol";

/**
 * @title Contract to manage cross-chain message hashes storage and status on L1.
 * @author ConsenSys Software Inc.
 */
abstract contract L1MessageManager is IL1MessageManager {
  uint8 public constant INBOX_STATUS_UNKNOWN = 0;
  uint8 public constant INBOX_STATUS_RECEIVED = 1;

  uint8 public constant OUTBOX_STATUS_UNKNOWN = 0;
  uint8 public constant OUTBOX_STATUS_SENT = 1;
  uint8 public constant OUTBOX_STATUS_RECEIVED = 2;

  /// @dev There is a uint216 worth of storage layout here.

  /// @dev Mapping to store L1->L2 message hashes status.
  /// @dev messageHash => messageStatus (0: unknown, 1: sent, 2: received).
  mapping(bytes32 => uint256) public outboxL1L2MessageStatus;

  /// @dev Mapping to store L2->L1 message hashes status.
  /// @dev messageHash => messageStatus (0: unknown, 1: received).
  mapping(bytes32 => uint256) public inboxL2L1MessageStatus;

  /// @dev Keep free storage slots for future implementation updates to avoid storage collision.
  // *******************************************************************************************
  // NB: THIS GAP HAS BEEN PUSHED OUT IN FAVOUR OF THE GAP INSIDE THE REENTRANCY CODE
  //uint256[50] private __gap;
  // NB: DO NOT USE THIS GAP
  // *******************************************************************************************

  /**
   * @notice Add a cross-chain L2->L1 message hash in storage.
   * @dev Once the event is emitted, it should be ready for claiming (post block finalization).
   * @param  _messageHash Hash of the message.
   */
  function _addL2L1MessageHash(bytes32 _messageHash) internal {
    if (inboxL2L1MessageStatus[_messageHash] != INBOX_STATUS_UNKNOWN) {
      revert MessageAlreadyReceived(_messageHash);
    }

    inboxL2L1MessageStatus[_messageHash] = INBOX_STATUS_RECEIVED;

    emit L2L1MessageHashAddedToInbox(_messageHash);
  }

  /**
   * @notice Update the status of L2->L1 message when a user claims a message on L1.
   * @dev The L2->L1 message is removed from storage.
   * @dev Due to the nature of the rollup, we should not get a second entry of this.
   * @param  _messageHash Hash of the message.
   */
  function _updateL2L1MessageStatusToClaimed(bytes32 _messageHash) internal {
    if (inboxL2L1MessageStatus[_messageHash] != INBOX_STATUS_RECEIVED) {
      revert MessageDoesNotExistOrHasAlreadyBeenClaimed();
    }

    delete inboxL2L1MessageStatus[_messageHash];
  }

  /**
   * @notice Add L1->L2 message hash in storage when a message is sent on L1.
   * @param  _messageHash Hash of the message.
   */
  function _addL1L2MessageHash(bytes32 _messageHash) internal {
    outboxL1L2MessageStatus[_messageHash] = OUTBOX_STATUS_SENT;
  }

  /**
   * @notice Update the status of L1->L2 messages as received when messages has been stored on L2.
   * @dev The expectation here is that the rollup is limited to 100 hashes being added here - array is not open ended.
   * @param  _messageHashes List of message hashes.
   */
  function _updateL1L2MessageStatusToReceived(bytes32[] memory _messageHashes) internal {
    uint256 messageHashArrayLength = _messageHashes.length;

    for (uint256 i; i < messageHashArrayLength; ) {
      bytes32 messageHash = _messageHashes[i];
      uint256 existingStatus = outboxL1L2MessageStatus[messageHash];

      if (existingStatus == OUTBOX_STATUS_UNKNOWN) {
        revert L1L2MessageNotSent(messageHash);
      }

      if (existingStatus != OUTBOX_STATUS_RECEIVED) {
        outboxL1L2MessageStatus[messageHash] = OUTBOX_STATUS_RECEIVED;
      }

      unchecked {
        i++;
      }
    }

    emit L1L2MessagesReceivedOnL2(_messageHashes);
  }
}
