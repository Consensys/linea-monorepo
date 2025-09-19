// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Native yield extension module for the Linea L1MessageService.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface ILineaNativeYieldExtension {
  /**
   * @notice Transfer ETH to the registered YieldManager.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Enforces that, after transfer, the L1MessageService balance remains ≥ the configured effective minimum reserve.
   * @param _amount Amount of ETH to transfer.
   */
  function transferFundsForNativeYield(uint256 _amount) external;

  /**
   * @notice Send ETH to this contract.
   * @dev FUNDER_ROLE is required to execute.
   */
  function fund() external payable;

  /**
   * @notice Permissionlessly donate ETH to this contract.
   * @dev Keeps track of ETH sent via this function for donation reporting purposes.
   */
  function fundPermissionless() external payable;

  /**
   * @notice Report native yield earned for L2 distribution by emitting a synthetic `MessageSent` event.
   * @dev Callable only by the registered YieldManager.
   * @param _amount The net earned yield.
   */
  function reportNativeYield(uint256 _amount) external;

  /**
   * @notice Set YieldManager address.
   * @dev YIELD_MANAGER_SETTER_ROLE is required to execute.
   * @param _yieldManager YieldManager address.
   */
  function setYieldManager(address _yieldManager) external;
}
