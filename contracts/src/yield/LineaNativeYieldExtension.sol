// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { ILineaNativeYieldExtension } from "./interfaces/ILineaNativeYieldExtension.sol";
import { IYieldManager } from "./interfaces/IYieldManager.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { IMessageService } from "../messaging/interfaces/IMessageService.sol";
import { LineaRollupPauseManager } from "../security/pausing/LineaRollupPauseManager.sol";

/**
 * @title Native yield extension module for the Linea L1MessageService.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract LineaNativeYieldExtension is LineaRollupPauseManager, ILineaNativeYieldExtension, IMessageService, IGenericErrors {
  /// @notice The role required to send ETH to the YieldManager.
  bytes32 public constant RESERVE_OPERATOR_ROLE = keccak256("RESERVE_OPERATOR_ROLE");

  /// @notice The role required to call fund().
  bytes32 public constant FUNDER_ROLE = keccak256("FUNDER_ROLE");

  /// @notice The role required to set the YieldManager address.
  bytes32 public constant YIELD_MANAGER_SETTER_ROLE = keccak256("YIELD_MANAGER_SETTER_ROLE");

  bool transient IS_WITHDRAW_LST_ALLOWED;

  /// @custom:storage-location erc7201:linea.storage.LineaNativeYieldExtensionStorage
  struct LineaNativeYieldExtensionStorage {    
      address _yieldManager;
  }

  // keccak256(abi.encode(uint256(keccak256("linea.storage.LineaNativeYieldExtensionStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant LineaNativeYieldExtensionStorageLocation = 0x1ca1eef1e96a909fae6702b42f1bcde6999f4e0fc09e0e51d048b197a65a8f00;

  function _storage() private pure returns (LineaNativeYieldExtensionStorage storage $) {
      assembly {
          $.slot := LineaNativeYieldExtensionStorageLocation
      }
  }

  /// @notice The address of the YieldManager.
  function yieldManager() public view returns (address) {
      return _storage()._yieldManager;
  }

  function isWithdrawLSTAllowed() external view returns (bool) {
    return IS_WITHDRAW_LST_ALLOWED;
  }

  /**
   * @notice Initialises the LineaNativeYieldExtension.
   * @param _yieldManager YieldManager address.
   */
  function __LineaNativeYieldExtension_init(address _yieldManager) internal onlyInitializing {
    setYieldManager(_yieldManager);
  }

  /**
   * @notice Transfer ETH to the registered YieldManager.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Enforces that, after transfer, the L1MessageService balance remains ≥ the configured effective minimum reserve.
   * @param _amount Amount of ETH to transfer.
   */
  function transferFundsForNativeYield(uint256 _amount) external whenTypeAndGeneralNotPaused(PauseType.L1_YIELDMANAGER) onlyRole(RESERVE_OPERATOR_ROLE) {
    IYieldManager(yieldManager()).receiveFundsFromReserve{ value: _amount }();
  }

  /**
   * @notice Send ETH to this contract.
   * @dev FUNDER_ROLE is required to execute.
   */
  function fund() external payable whenTypeAndGeneralNotPaused(PauseType.FUNDING) onlyRole(FUNDER_ROLE) {
    emit FundingReceived(msg.sender, msg.value);
  }

  /**
   * @notice Set YieldManager address.
   * @dev YIELD_MANAGER_SETTER_ROLE is required to execute.
   * @param _newYieldManager YieldManager address.
   */
  function setYieldManager(address _newYieldManager) public onlyRole(YIELD_MANAGER_SETTER_ROLE) {
    if (_newYieldManager == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    LineaNativeYieldExtensionStorage storage $ = _storage();
    emit YieldManagerChanged($._yieldManager, _newYieldManager, msg.sender);
    $._yieldManager = _newYieldManager;
  }
}
