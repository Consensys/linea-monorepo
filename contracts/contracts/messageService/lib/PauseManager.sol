// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.19;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { IPauseManager } from "../../interfaces/IPauseManager.sol";

/**
 * @title Contract to manage cross-chain function pausing.
 * @author ConsenSys Software Inc.
 */
abstract contract PauseManager is Initializable, IPauseManager, AccessControlUpgradeable {
  bytes32 public constant PAUSE_MANAGER_ROLE = keccak256("PAUSE_MANAGER_ROLE");

  bytes32 public constant GENERAL_PAUSE_TYPE = keccak256("GENERAL_PAUSE_TYPE");
  bytes32 public constant L1_L2_PAUSE_TYPE = keccak256("L1_L2_PAUSE_TYPE");
  bytes32 public constant L2_L1_PAUSE_TYPE = keccak256("L2_L1_PAUSE_TYPE");
  bytes32 public constant PROVING_SYSTEM_PAUSE_TYPE = keccak256("PROVING_SYSTEM_PAUSE_TYPE");

  mapping(bytes32 => bool) public pauseTypeStatuses;

  uint256[10] private _gap;

  /**
   * @dev Modifier to make a function callable only when the type is not paused.
   *
   * Requirements:
   *
   * - The type must not be paused.
   */
  modifier whenTypeNotPaused(bytes32 _pauseType) {
    _requireTypeNotPaused(_pauseType);
    _;
  }

  /**
   * @dev Modifier to make a function callable only when the type is paused.
   *
   * Requirements:
   *
   * - The type must not be paused.
   */
  modifier whenTypePaused(bytes32 _pauseType) {
    _requireTypePaused(_pauseType);
    _;
  }

  /**
   * @dev Throws if the type is not paused.
   * @param _pauseType The keccak256 pause type being checked,
   */
  function _requireTypePaused(bytes32 _pauseType) internal view virtual {
    if (!pauseTypeStatuses[_pauseType]) {
      revert IsNotPaused(_pauseType);
    }
  }

  /**
   * @dev Throws if the type is paused.
   * @param _pauseType The keccak256 pause type being checked,
   */
  function _requireTypeNotPaused(bytes32 _pauseType) internal view virtual {
    if (pauseTypeStatuses[_pauseType]) {
      revert IsPaused(_pauseType);
    }
  }

  /**
   * @notice Pauses functionality by specific type.
   * @dev Requires PAUSE_MANAGER_ROLE.
   * @param _pauseType keccak256 pause type.
   **/
  function pauseByType(bytes32 _pauseType) external whenTypeNotPaused(_pauseType) onlyRole(PAUSE_MANAGER_ROLE) {
    pauseTypeStatuses[_pauseType] = true;
    emit Paused(_msgSender(), _pauseType);
  }

  /**
   * @notice Unpauses functionality by specific type.
   * @dev Requires PAUSE_MANAGER_ROLE.
   * @param _pauseType keccak256 pause type.
   **/
  function unPauseByType(bytes32 _pauseType) external whenTypePaused(_pauseType) onlyRole(PAUSE_MANAGER_ROLE) {
    pauseTypeStatuses[_pauseType] = false;
    emit UnPaused(_msgSender(), _pauseType);
  }
}
