// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title L1 Message manager V1 interface for pre-existing functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IL1MessageManagerV1 {
  /**
   * @notice Emitted when L2->L1 message hashes have been added to L1 storage.
   * @param messageHash The indexed hash of the message parameters.
   * @dev DEPRECATED - This is kept for backwards compatability for external consumers.
   */
  event L2L1MessageHashAddedToInbox(bytes32 indexed messageHash);

  /**
   * @notice Emitted when L1->L2 messages have been anchored on L2 and updated on L1.
   * @param messageHashes The collection of hashes indicating which messages were added on L2. of the message parameters.
   * @dev DEPRECATED - This is kept for backwards compatability for external consumers.
   */
  event L1L2MessagesReceivedOnL2(bytes32[] messageHashes);

  /**
   * @dev Thrown when the message has already been claimed.
   */
  error MessageDoesNotExistOrHasAlreadyBeenClaimed(bytes32 messageHash);
}
