// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.19 <=0.8.24;

/**
 * @title Interface declaring pre-existing pausing functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IPauseManager {
  /**
   * @notice Emitted when a pause type is paused.
   * @param messageSender The address performing the pause.
   * @param pauseType The indexed pause type that was paused.
   */
  event Paused(address messageSender, uint256 indexed pauseType);

  /**
   * @notice Emitted when a pause type is unpaused.
   * @param messageSender The address performing the unpause.
   * @param pauseType The indexed pause type that was unpaused.
   */
  event UnPaused(address messageSender, uint256 indexed pauseType);

  /**
   * @dev Thrown when a specific pause type is paused.
   */
  error IsPaused(uint256 pauseType);

  /**
   * @dev Thrown when a specific pause type is not paused and expected to be.
   */
  error IsNotPaused(uint256 pauseType);
}
