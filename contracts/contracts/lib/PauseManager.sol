// SPDX-License-Identifier: AGPL-3.0
pragma solidity >=0.8.19 <=0.8.26;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { IPauseManager } from "../interfaces/IPauseManager.sol";

/**
 * @title Contract to manage cross-chain function pausing.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract PauseManager is Initializable, IPauseManager, AccessControlUpgradeable {
  bytes32 public constant PAUSE_ALL_ROLE = keccak256("PAUSE_ALL_ROLE");
  bytes32 public constant UNPAUSE_ALL_ROLE = keccak256("UNPAUSE_ALL_ROLE");
  bytes32 public constant PAUSE_L1_L2_ROLE = keccak256("PAUSE_L1_L2_ROLE");
  bytes32 public constant UNPAUSE_L1_L2_ROLE = keccak256("UNPAUSE_L1_L2_ROLE");
  bytes32 public constant PAUSE_L2_L1_ROLE = keccak256("PAUSE_L2_L1_ROLE");
  bytes32 public constant UNPAUSE_L2_L1_ROLE = keccak256("UNPAUSE_L2_L1_ROLE");
  bytes32 public constant PAUSE_L2_BLOB_SUBMISSION_ROLE = keccak256("PAUSE_L2_BLOB_SUBMISSION_ROLE");
  bytes32 public constant UNPAUSE_L2_BLOB_SUBMISSION_ROLE = keccak256("UNPAUSE_L2_BLOB_SUBMISSION_ROLE");
  bytes32 public constant PAUSE_FINALIZE_WITHPROOF_ROLE = keccak256("PAUSE_FINALIZE_WITHPROOF_ROLE");
  bytes32 public constant UNPAUSE_FINALIZE_WITHPROOF_ROLE = keccak256("UNPAUSE_FINALIZE_WITHPROOF_ROLE");

  uint8 public constant GENERAL_PAUSE_TYPE = 1;
  uint8 public constant L1_L2_PAUSE_TYPE = 2;
  uint8 public constant L2_L1_PAUSE_TYPE = 3;
  uint8 public constant BLOB_SUBMISSION_PAUSE_TYPE = 4;
  uint8 public constant CALLDATA_SUBMISSION_PAUSE_TYPE = 5;
  uint8 public constant FINALIZATION_PAUSE_TYPE = 6;
  uint8 public constant INITIATE_TOKEN_BRIDGING_PAUSE_TYPE = 7;
  uint8 public constant COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE = 8;

  // @dev DEPRECATED. USE _pauseTypeStatusesBitMap INSTEAD
  mapping(bytes32 pauseType => bool pauseStatus) public pauseTypeStatuses;

  uint256 private _pauseTypeStatusesBitMap;
  mapping(uint8 pauseType => bytes32 role) private pauseTypeRoles;
  mapping(uint8 unPauseType => bytes32 role) private unPauseTypeRoles;

  /// @dev Total contract storage is 11 slots with the gap below.
  /// @dev Keep 7 free storage slots for future implementation updates to avoid storage collision.
  /// @dev Note: This was reduced previously to cater for new functionality.
  uint256[7] private __gap;

  /**
   * @dev Modifier to make a function callable only when the specific and general types are not paused.
   * @param _pauseType The pause type value being checked.
   * Requirements:
   *
   * - The type must not be paused.
   */
  modifier whenTypeAndGeneralNotPaused(uint8 _pauseType) {
    _requireTypeAndGeneralNotPaused(_pauseType);
    _;
  }

  /**
   * @dev Modifier to make a function callable only when the type is not paused.
   * @param _pauseType The pause type value being checked.
   * Requirements:
   *
   * - The type must not be paused.
   */
  modifier whenTypeNotPaused(uint8 _pauseType) {
    _requireTypeNotPaused(_pauseType);
    _;
  }

  /**
   * @notice Initializes the pause manager with the given pause and unpause roles.
   * @dev This function is called during contract initialization to set up the pause and unpause roles.
   * @param _pauseTypeRoles An array of PauseTypeRole structs defining the pause types and their associated roles.
   * @param _unPauseTypeRoles An array of PauseTypeRole structs defining the unpause types and their associated roles.
   */
  function __PauseManager_init(
    PauseTypeRole[] calldata _pauseTypeRoles,
    PauseTypeRole[] calldata _unPauseTypeRoles
  ) internal onlyInitializing {
    uint256 arrayLength = _pauseTypeRoles.length;
    for (uint256 i; i < arrayLength; i++) {
      pauseTypeRoles[_pauseTypeRoles[i].pauseType] = _pauseTypeRoles[i].role;
    }
    arrayLength = _unPauseTypeRoles.length;
    for (uint256 i; i < arrayLength; i++) {
      unPauseTypeRoles[_unPauseTypeRoles[i].pauseType] = _unPauseTypeRoles[i].role;
    }
  }

  /**
   * @dev Throws if the specific or general types are paused.
   * @dev Checks the specific and general pause types.
   * @param _pauseType The pause type value being checked.
   */
  function _requireTypeAndGeneralNotPaused(uint8 _pauseType) internal view virtual {
    uint256 pauseBitMap = _pauseTypeStatusesBitMap;

    if (pauseBitMap & (1 << uint256(_pauseType)) != 0) {
      revert IsPaused(_pauseType);
    }

    if (pauseBitMap & (1 << uint256(GENERAL_PAUSE_TYPE)) != 0) {
      revert IsPaused(GENERAL_PAUSE_TYPE);
    }
  }

  /**
   * @dev Throws if the type is paused.
   * @dev Checks the specific pause type.
   * @param _pauseType The pause type value being checked.
   */
  function _requireTypeNotPaused(uint8 _pauseType) internal view virtual {
    if (isPaused(_pauseType)) {
      revert IsPaused(_pauseType);
    }
  }

  /**
   * @notice Pauses functionality by specific type.
   * @dev Requires the role mapped in pauseTypeRoles for the pauseType.
   * @param _pauseType The pause type value.
   */
  function pauseByType(uint8 _pauseType) external onlyRole(pauseTypeRoles[_pauseType]) {
    if (isPaused(_pauseType)) {
      revert IsPaused(_pauseType);
    }

    _pauseTypeStatusesBitMap |= 1 << uint256(_pauseType);
    emit Paused(_msgSender(), _pauseType);
  }

  /**
   * @notice Unpauses functionality by specific type.
   * @dev Requires the role mapped in unPauseTypeRoles for the pauseType.
   * @param _pauseType The pause type value.
   */
  function unPauseByType(uint8 _pauseType) external onlyRole(unPauseTypeRoles[_pauseType]) {
    if (!isPaused(_pauseType)) {
      revert IsNotPaused(_pauseType);
    }

    _pauseTypeStatusesBitMap &= ~(1 << uint256(_pauseType));
    emit UnPaused(_msgSender(), _pauseType);
  }

  /**
   * @notice Check if a pause type is enabled.
   * @param _pauseType The pause type value.
   * @return boolean True if the pause type if enabled, false otherwise.
   */
  function isPaused(uint8 _pauseType) public view returns (bool) {
    return (_pauseTypeStatusesBitMap & (1 << uint256(_pauseType))) != 0;
  }
}
