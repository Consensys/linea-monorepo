// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { ILineaRollupYieldExtension } from "./interfaces/ILineaRollupYieldExtension.sol";
import { IYieldManager } from "../yield/interfaces/IYieldManager.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { IMessageService } from "../messaging/interfaces/IMessageService.sol";
import { LineaRollupPauseManager } from "../security/pausing/LineaRollupPauseManager.sol";

/**
 * @title Native yield extension module for the Linea L1MessageService.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract LineaRollupYieldExtension is
  LineaRollupPauseManager,
  ILineaRollupYieldExtension,
  IMessageService,
  IGenericErrors
{
  /// @notice The role required to send ETH to the YieldManager.
  bytes32 public constant YIELD_PROVIDER_STAKING_ROLE = keccak256("YIELD_PROVIDER_STAKING_ROLE");

  /// @notice The role required to call fund().
  bytes32 public constant FUNDER_ROLE = keccak256("FUNDER_ROLE");

  /// @notice The role required to set the YieldManager address.
  bytes32 public constant SET_YIELD_MANAGER_ROLE = keccak256("SET_YIELD_MANAGER_ROLE");

  bool transient IS_WITHDRAW_LST_ALLOWED;

  /// @dev To consider - could this be an immutable variable instead? Do we expect YieldManager instance to change at a different cadence than LineaRollup upgrades?
  /// @custom:storage-location erc7201:linea.storage.LineaRollupYieldExtensionStorage
  struct LineaRollupYieldExtensionStorage {
    address _yieldManager;
  }

  // keccak256(abi.encode(uint256(keccak256("linea.storage.LineaRollupYieldExtensionStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant LineaRollupYieldExtensionStorageLocation =
    0x1ca1eef1e96a909fae6702b42f1bcde6999f4e0fc09e0e51d048b197a65a8f00;

  function _storage() private pure returns (LineaRollupYieldExtensionStorage storage $) {
    assembly {
      $.slot := LineaRollupYieldExtensionStorageLocation
    }
  }

  /// @notice The address of the YieldManager.
  function yieldManager() public view returns (address) {
    return _storage()._yieldManager;
  }

  /**
   * @notice Initialises the LineaRollupYieldExtension.
   * @param _yieldManager YieldManager address.
   */
  function __LineaRollupYieldExtension_init(address _yieldManager) internal onlyInitializing {
    emit YieldManagerChanged(_storage()._yieldManager, _yieldManager);
    _storage()._yieldManager = _yieldManager;
  }

  function isWithdrawLSTAllowed() external view returns (bool) {
    return IS_WITHDRAW_LST_ALLOWED;
  }

  /**
   * @notice Transfer ETH to the registered YieldManager.
   * @dev YIELD_PROVIDER_STAKING_ROLE is required to execute.
   * @dev Enforces that, after transfer, the L1MessageService balance remains ≥ the configured effective minimum reserve.
   * @param _amount Amount of ETH to transfer.
   */
  function transferFundsForNativeYield(
    uint256 _amount
  ) external whenTypeAndGeneralNotPaused(PauseType.NATIVE_YIELD_STAKING) onlyRole(YIELD_PROVIDER_STAKING_ROLE) {
    IYieldManager(yieldManager()).receiveFundsFromReserve{ value: _amount }();
  }

  /**
   * @notice Send ETH to this contract.
   * @dev FUNDER_ROLE is required to execute.
   */
  function fund() external payable onlyRole(FUNDER_ROLE) {
    emit FundingReceived(msg.sender, msg.value);
  }

  /**
   * @notice Set YieldManager address.
   * @dev SET_YIELD_MANAGER_ROLE is required to execute.
   * @param _newYieldManager YieldManager address.
   */
  function setYieldManager(address _newYieldManager) public onlyRole(SET_YIELD_MANAGER_ROLE) {
    require(_newYieldManager != address(0), ZeroAddressNotAllowed());
    LineaRollupYieldExtensionStorage storage $ = _storage();
    emit YieldManagerChanged($._yieldManager, _newYieldManager);
    $._yieldManager = _newYieldManager;
  }
}
