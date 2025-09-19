// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { ILineaNativeYieldExtension } from "./interfaces/ILineaNativeYieldExtension.sol";
import { IYieldManager } from "./interfaces/IYieldManager.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { IMessageService } from "../messaging/interfaces/IMessageService.sol";

/**
 * @title Native yield extension module for the Linea L1MessageService.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract LineaNativeYieldExtension is AccessControlUpgradeable, ILineaNativeYieldExtension, IMessageService, IGenericErrors {
  /// @notice The role required to send ETH to the YieldManager.
  bytes32 public constant RESERVE_OPERATOR_ROLE = keccak256("RESERVE_OPERATOR_ROLE");

  /// @notice The role required to call fund().
  bytes32 public constant FUNDER_ROLE = keccak256("FUNDER_ROLE");

  /// @notice The role required to set the YieldManager address.
  bytes32 public constant YIELD_MANAGER_SETTER_ROLE = keccak256("YIELD_MANAGER_SETTER_ROLE");

  /// @notice The role required to set the L2YieldRecipient address.
  bytes32 public constant L2_YIELD_RECIPIENT_SETTER_ROLE = keccak256("L2_YIELD_RECIPIENT_SETTER_ROLE");

  /// @notice The address of the YieldManager.
  address public yieldManager;

  /// @notice The address of the L2YieldRecipient.
  address public l2YieldRecipient;

  /// @notice The total ETH received through fundPermissionless().
  uint256 public permissionlessDonationTotal;

  /**
   * @notice Transfer ETH to the registered YieldManager.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Enforces that, after transfer, the L1MessageService balance remains ≥ the configured effective minimum reserve.
   * @param _amount Amount of ETH to transfer.
   */
  function transferFundsForNativeYield(uint256 _amount) external onlyRole(RESERVE_OPERATOR_ROLE) {
    IYieldManager(yieldManager).receiveFundsFromReserve{ value: _amount }();
  }

  /**
   * @notice Send ETH to this contract.
   * @dev FUNDER_ROLE is required to execute.
   */
  function fund() external payable onlyRole(YIELD_MANAGER_SETTER_ROLE) {
    emit FundingReceived(msg.sender, msg.value);
  }

  /**
   * @notice Permissionlessly donate ETH to this contract.
   * @dev Keeps track of ETH sent via this function for donation reporting purposes.
   */
  function fundPermissionless() external payable {
    emit PermissionlessDonationReceived(msg.sender, msg.value);
    permissionlessDonationTotal += msg.value;
  }

  /**
   * @notice Set YieldManager address.
   * @dev YIELD_MANAGER_SETTER_ROLE is required to execute.
   * @param _newYieldManager YieldManager address.
   */
  function setYieldManager(address _newYieldManager) external onlyRole(YIELD_MANAGER_SETTER_ROLE) {
    if (_newYieldManager == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    emit YieldManagerChanged(yieldManager, _newYieldManager, msg.sender);
    yieldManager = _newYieldManager;
  }

  /**
   * @notice Set l2YieldRecipient address.
   * @dev L2_YIELD_RECIPIENT_SETTER_ROLE is required to execute.
   * @param _newL2YieldRecipient L2YieldRecipient address.
   */
  function setL2YieldRecipient(address _newL2YieldRecipient) external onlyRole(L2_YIELD_RECIPIENT_SETTER_ROLE) {
    if (_newL2YieldRecipient == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    emit L2YieldRecipientChanged(l2YieldRecipient, _newL2YieldRecipient, msg.sender);
    l2YieldRecipient = _newL2YieldRecipient;
  }
}
