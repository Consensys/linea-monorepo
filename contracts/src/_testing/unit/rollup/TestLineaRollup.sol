// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LineaRollup } from "../../../rollup/LineaRollup.sol";
import { LineaRollupBase } from "../../../rollup/LineaRollupBase.sol";
import { CalldataBlobAcceptor } from "../../../rollup/dataAvailability/CalldataBlobAcceptor.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestLineaRollup is LineaRollup, CalldataBlobAcceptor {
  function setLivenessRecoveryOperatorAddress(address _livenessRecoveryOperator) external {
    livenessRecoveryOperator = _livenessRecoveryOperator;
  }

  function addRollingHash(uint256 _messageNumber, bytes32 _messageHash) external {
    _addRollingHash(_messageNumber, _messageHash);
  }

  function setRollingHash(uint256 _messageNumber, bytes32 _rollingHash) external {
    rollingHashes[_messageNumber] = _rollingHash;
  }

  function validateL2ComputedRollingHash(uint256 _rollingHashMessageNumber, bytes32 _rollingHash) external view {
    _validateL2ComputedRollingHash(_rollingHashMessageNumber, _rollingHash);
  }

  function calculateY(bytes calldata _data, bytes32 _x) external pure returns (bytes32 y) {
    return _calculateY(_data, _x);
  }

  function setupParentShnarf(bytes32 _shnarf) external {
    _blobShnarfExists[_shnarf] = 1;
  }

  function setLastFinalizedBlock(uint256 _blockNumber) external {
    currentL2BlockNumber = _blockNumber;
  }

  function setLastFinalizedShnarf(bytes32 _lastFinalizedShnarf) external {
    currentFinalizedShnarf = _lastFinalizedShnarf;
  }

  function setShnarfFinalBlockNumber(bytes32 _shnarf, uint256 _finalBlockNumber) external {
    _blobShnarfExists[_shnarf] = _finalBlockNumber;
  }

  function setLastFinalizedState(uint256 _messageNumber, bytes32 _rollingHash, uint256 _timestamp) external {
    currentFinalizedState = _computeLastFinalizedState(_messageNumber, _rollingHash, _timestamp);
  }

  /**
   * @notice Revokes `role` from the calling account.
   * @dev Liveness recovery operator cannot renounce role. Reverts with OnlyNonLivenessRecoveryOperator.
   * @param _role The role to renounce.
   * @param _account The account to renounce - can only be the _msgSender().
   */
  function renounceRole(
    bytes32 _role,
    address _account
  ) public virtual override(LineaRollup, AccessControlUpgradeable) {
    super.renounceRole(_role, _account);
  }
}
