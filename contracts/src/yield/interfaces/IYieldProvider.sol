// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { IYieldManager } from "./IYieldManager.sol";

/**
 * @title Contract that will the YieldManager will delegatecall, to handle provider-specific yield operations.
 * @author ConsenSys Software Inc.
 * @dev YieldProvider will handle only the external protocol interactions, YieldManager will handle the remainder (storage update, input validation, etc).
 * @custom:security-contact security-report@linea.build
 */
interface IYieldProvider {
  enum OperationType {
      ReportYield
  }

  error IncorrectYieldProviderType();

  error OperationNotSupportedDuringOssification(OperationType operationType);

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev Since the YieldManager is unaware of donations received via the L1MessageService or L2MessageService,
   *      the `_reserveDonations` parameter is required to ensure accurate yield accounting.
   * @param _totalReserveDonations   Total amount of donations received on the L1MessageService or L2MessageService.
   */
  function reportYield(address _yieldProvider, uint256 _totalReserveDonations) external returns (uint256);

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external;

  /**
   * @notice Permissionlessly request beacon chain withdrawal.
   * @dev    Callable only when the withdrawal reserve is in deficit. 
   * @dev    The permissionless unstake amount is capped to the remaining reserve deficit that 
   *         cannot be covered by other liquidity sources:
   *
   *         PERMISSIONLESS_UNSTAKE_AMOUNT ≤
   *           RESERVE_DEFICIT
   *         - YIELD_PROVIDER_BALANCE
   *         - YIELD_MANAGER_BALANCE
   *         - PENDING_PERMISSIONLESS_UNSTAKE
   *
   * @dev Validates (validatorPubkey, validatorBalance, validatorWithdrawalCredential) against EIP-4788 beacon chain root.
   * @param _withdrawalParams       Provider-specific withdrawal parameters.
   * @param _withdrawalParamsProof  Merkle proof of _withdrawalParams to be verified against EIP-4788 beacon chain root.
   */
  function unstakePermissionless(
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external returns (uint256);

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If fund remaining, will settle any outstanding LST liabilities.
   * @param _amount                 Amount to withdraw.
   */
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount, uint256 _targetReserveDeficit, address _recipient) external returns (uint256);

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   */
  function pauseStaking(address _yieldProvider) external;

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   */
  function unpauseStaking(address _yieldProvider) external;

  function validateAdditionToYieldManager(IYieldManager.YieldProviderRegistration calldata _yieldProviderRegistration) external;

  // Get current ETH balance on the YieldProvider available for withdraw
  function getAvailableBalanceForWithdraw(address _yieldProvider) external view returns (uint256);

  function mintLST(address _yieldProvider, uint256 _amount, address _recipient) external;

  function initiateOssification(address _yieldProvider) external;

  function processPendingOssification(address _yieldProvider) external returns (bool);
}
