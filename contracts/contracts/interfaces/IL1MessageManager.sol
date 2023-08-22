// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

interface IL1MessageManager {
  /**
   * @dev Emitted when L2->L1 message hashes have been added to L1 storage.
   */
  event L2L1MessageHashAddedToInbox(bytes32 indexed messageHash);

  /**
   * @dev Emitted when L1->L2 messages have been anchored on L2 and updated on L1.
   */
  event L1L2MessagesReceivedOnL2(bytes32[] messageHashes);

  /**
   * @dev Thrown when the message has been already sent.
   */
  error MessageAlreadySent();

  /**
   * @dev Thrown when the message has already been claimed.
   */
  error MessageDoesNotExistOrHasAlreadyBeenClaimed();

  /**
   * @dev Thrown when the message has already been received.
   */
  error MessageAlreadyReceived(bytes32 messageHash);

  /**
   * @dev Thrown when the L1->L2 message has not been sent.
   */
  error L1L2MessageNotSent(bytes32 messageHash);
}
