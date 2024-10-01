// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.19 <=0.8.26;

/**
 * @title Interface declaring pre-existing pausing functions, events and errors.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IPauseManager {
  /**
   * @notice Structure defining a pause type and its associated role.
   * @dev This struct is used for both the _pauseTypeRoles and _unPauseTypeRoles mappings.
   * @param pauseType The type of pause.
   * @param role The role associated with the pause type.
   */
  struct PauseTypeRole {
    uint8 pauseType;
    bytes32 role;
  }

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
   * @notice Emitted when a pause type and its associated role are set in the `_pauseTypeRoles` mapping.
   * @param pauseType The type of pause.
   * @param role The role associated with the pause type.
   */
  event PauseTypeRoleSet(uint8 pauseType, bytes32 role);

  /**
   * @notice Emitted when an unpause type and its associated role are set in the `_unPauseTypeRoles` mapping.
   * @param unPauseType The type of unpause.
   * @param role The role associated with the unpause type.
   */
  event UnPauseTypeRoleSet(uint8 unPauseType, bytes32 role);

  /**
   * @dev Thrown when a specific pause type is paused.
   */
  error IsPaused(uint256 pauseType);

  /**
   * @dev Thrown when a specific pause type is not paused and expected to be.
   */
  error IsNotPaused(uint256 pauseType);

  /**
   * @notice Pauses functionality by specific type.
   * @dev Requires the role mapped in pauseTypeRoles for the pauseType.
   * @param _pauseType The pause type value.
   */
  function pauseByType(uint8 _pauseType) external;

  /**
   * @notice Unpauses functionality by specific type.
   * @dev Requires the role mapped in unPauseTypeRoles for the pauseType.
   * @param _pauseType The pause type value.
   */
  function unPauseByType(uint8 _pauseType) external;

  /**
   * @notice Check if a pause type is enabled.
   * @param _pauseType The pause type value.
   * @return boolean True if the pause type if enabled, false otherwise.
   */
  function isPaused(uint8 _pauseType) external view returns (bool);
}
