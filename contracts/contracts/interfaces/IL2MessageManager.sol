// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

interface IL2MessageManager {
  /**
   * @dev Emitted when L1->L2 message hashes have been added to L2 storage.
   */
  event L1L2MessageHashesAddedToInbox(bytes32[] messageHashes);

  /**
   * @dev Thrown when the message hashes list length is higher than one hundred.
   */
  error MessageHashesListLengthHigherThanOneHundred(uint256 length);

  /**
   * @dev Thrown when the message hashes array is empty.
   */
  error EmptyMessageHashesArray();

  /**
   * @dev Thrown when the message does not exist or has already been claimed.
   */
  error MessageDoesNotExistOrHasAlreadyBeenClaimed();

  /**
   * @notice Anchor L1-> L2 message hashes.
   * @param _messageHashes New message hashes to anchor on L2.
   */
  function addL1L2MessageHashes(bytes32[] calldata _messageHashes) external;
}
