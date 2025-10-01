// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "../YieldManagerStorageLayout.sol";

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

  /// @notice Thrown when an operation is forbidden while ossification is active.
  /// @param operationType The attempted operation type.
  error OperationNotSupportedDuringOssification(OperationType operationType);

  /// @notice Thrown when the registration yield provider type is not expected.
  error UnknownYieldProviderVendor();

  error MintLSTDisabledDuringOssification();

  error SingleValidatorOnlyForUnstakePermissionless();

  error NoValidatorExitForUnstakePermissionless();

  error ContextIsNotYieldManager();

  /**
   * @notice Get the ETH balance held by the yield provider that can be withdrawn immediately.
   * @return The available ETH balance that may be withdrawn.
   */
  function withdrawableValue(address _yieldProvider) external view returns (uint256);

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   */
  function reportYield(address _yieldProvider) external returns (uint256 newReportedYield);

  /**
   * @notice Repay part or all of the outstanding LST principal liability.
   * @param _maxAvailableRepaymentETH Maximum amount of ETH available to repay the liability.
   * @return lstPrincipalPaid The amount of ETH used to reduce the liability.
   */
  function payLSTPrincipal(address _yieldProvider, uint256 _maxAvailableRepaymentETH) external returns (uint256 lstPrincipalPaid);

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable;

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
  ) external payable returns (uint256 maxUnstakeAmount);

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If fund remaining, will settle any outstanding LST liabilities.
   * @param _amount Amount to withdraw.
   */
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   */
  function pauseStaking(address _yieldProvider) external;

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   */
  function unpauseStaking(address _yieldProvider) external;

  /**
   * @notice Mint LST to a recipient .
   * @param _amount Amount of underlying to convert into LST.
   * @param _recipient Address that receives the minted LST.
   */
  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external;

  /**
   * @notice Start the ossification process for the yield provider.
   */
  function initiateOssification(address _yieldProvider) external;

  /**
   * @notice Start the ossification process for the yield provider.
   */
  function undoInitiateOssification(address _yieldProvider) external;

  /**
   * @notice Process a previously initiated ossification process.
   * @return isOssificationComplete True if ossification is completed.
   */
  function processPendingOssification(address _yieldProvider) external returns (bool isOssificationComplete);

  /**
   * @notice Validate the supplied registration before it is added to the yield manager.
   * @param _registration Supplied registration data for the yield provider.
   */
  function validateAdditionToYieldManager(
    YieldManagerStorageLayout.YieldProviderRegistration calldata _registration
  ) external;
}
