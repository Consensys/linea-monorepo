// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

interface IPauseManager {
  /**
   * @dev Thrown when a specific pause type is paused.
   */
  error IsPaused(bytes32 pauseType);

  /**
   * @dev Thrown when a specific pause type is not paused and expected to be.
   */
  error IsNotPaused(bytes32 pauseType);

  /**
   * @dev Emitted when a pause type is paused.
   */
  event Paused(address messageSender, bytes32 pauseType);

  /**
   * @dev Emitted when a pause type is unpaused.
   */
  event UnPaused(address messageSender, bytes32 pauseType);
}
