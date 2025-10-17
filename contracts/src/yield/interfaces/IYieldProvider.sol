// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "../YieldManagerStorageLayout.sol";

/**
 * @title Interface for a YieldProvider adaptor contract to handle vendor-specific interactions.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IYieldProvider {
  /// @notice Enumerates operations that can be paused during ossification depending on the yield provider vendor.
  enum OperationType {
    ReportYield
  }

  /**
   * @notice Enum defining the specific type of YieldProvider registration error.
   */
  enum YieldProviderRegistrationError {
    LidoDashboardNotLinkedToVault,
    LidoVaultIsExpectedReceiveCallerAndOssifiedEntrypoint
  }

  /// @notice Thrown when an operation is blocked because ossification is either pending or complete.
  /// @param operationType The operation that was attempted.
  error OperationNotSupportedDuringOssification(OperationType operationType);

  /// @notice Raised when the YieldManager attempts to add a yield provider with an unexpected vendor identifier.
  error UnknownYieldProviderVendor();

  /// @notice Raised when unpause staking is attempted when ossification is complete.
  error UnpauseStakingForbiddenWhenOssified();

  /// @notice Raised when LST minting is attempted while ossification is either pending or complete.
  error MintLSTDisabledDuringOssification();

  /// @notice Raised when a permissionless unstake request references more than one validator.
  error SingleValidatorOnlyForUnstakePermissionless();

  /// @notice Raised when a permissionless unstake request specifies a zero exit amount.
  error NoValidatorExitForUnstakePermissionless();

  /// @notice Raised when a function is called outside of a `delegatecall` from the YieldManager.
  error ContextIsNotYieldManager();

  /// @notice Thrown when a YieldProvider registration is invalid.
  /// @param error Specific error details.
  error InvalidYieldProviderRegistration(YieldProviderRegistrationError error);

  /**
   * @notice Returns the amount of ETH the provider can immediately remit back to the YieldManager.
   * @dev Called via `delegatecall` from the YieldManager.
   * @param _yieldProvider The yield provider address.
   * @return availableBalance The ETH amount that can be withdrawn.
   */
  function withdrawableValue(address _yieldProvider) external view returns (uint256 availableBalance);

  /**
   * @notice Forwards ETH from the YieldManager to the yield provider.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of ETH supplied by the YieldManager.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Computes and returns earned yield that can be distributed to L2 users.
   * @dev Implementations should apply provider-specific adjustments (obligations, fees, negative
   *      yield) and mutate only `lstLiabilityPrincipal` and `currentNegativeYield` in the
   *      YieldProvider state.
   * @param _yieldProvider The yield provider address.
   * @return newReportedYield New net yield (denominated in ETH) since the prior report.
   * @return outstandingNegativeYield Amount of outstanding negative yield.
   */
  function reportYield(
    address _yieldProvider
  ) external returns (uint256 newReportedYield, uint256 outstandingNegativeYield);

  /**
   * @notice Reduces the outstanding LST liability principal.
   * @dev Called after the YieldManager has reserved `_availableFunds` for liability
   *      settlement.
   *      - Implementations should update `lstLiabilityPrincipal` in the YieldProvider storage
   *      - Implementations should ensure lstPrincipalPaid <= _availableFunds
   * @param _yieldProvider The yield provider address.
   * @param _availableFunds The maximum amount of ETH that is available to pay LST liability principal.
   * @return lstPrincipalPaid The actual ETH amount paid to reduce LST liability principal.
   */
  function payLSTPrincipal(address _yieldProvider, uint256 _availableFunds) external returns (uint256 lstPrincipalPaid);

  /**
   * @notice Requests beacon chain withdrawal via EIP-7002 withdrawal contract.
   * @dev Parameters are ABI encoded by the YieldManager and understood by the yield provider.
   * @dev Dynamic withdrawal fee is sourced from `msg.value`
   * @param _yieldProvider The yield provider address.
   * @param _withdrawalParams Provider-specific payload describing the withdrawals to trigger.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable;

  /**
   * @notice Permissionlessly requests beacon chain withdrawal via EIP-7002 withdrawal contract when reserve is under minimum threshold.
   * @dev Implementations must verify the calldata proof (for example against EIP-4788 beacon roots)
   *      and enforce any provider-specific safety checks. The returned amount is used by the
   *      YieldManager to cap pending withdrawals tracked on L1.
   * @param _yieldProvider The yield provider address.
   * @param _withdrawalParams ABI encoded provider parameters.
   * @param _withdrawalParamsProof Proof data (typically a beacon chain Merkle proof).
   * @return maxUnstakeAmount Maximum ETH amount expected to be withdrawn as a result of this request.
   */
  function unstakePermissionless(
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external payable returns (uint256 maxUnstakeAmount);

  /**
   * @notice Withdraws ETH from the provider back into the YieldManager.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of ETH to withdraw to the YieldManager.
   */
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Pauses new beacon chain deposits.
   * @param _yieldProvider The yield provider address.
   */
  function pauseStaking(address _yieldProvider) external;

  /**
   * @notice Resumes beacon chain deposits for the provider after a pause.
   * @param _yieldProvider The yield provider address.
   * @dev Whether to allow staking during ossification is a vendor-specific detail.
   */
  function unpauseStaking(address _yieldProvider) external;

  /**
   * @notice Withdraws liquid staking tokens (LST) to a recipient.
   * @dev Implementations must `lstLiabilityPrincipal` state for the yield provider.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of LST (denominated in ETH) to withdraw.
   * @param _recipient Address receiving the LST.
   */
  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external;

  /**
   * @notice Begins the provider-specific ossification workflow.
   * @param _yieldProvider The yield provider address.
   */
  function initiateOssification(address _yieldProvider) external;

  /**
   * @notice Process a previously initiated ossification process.
   * @param _yieldProvider The yield provider address.
   * @return isOssificationComplete True if the provider is now in the ossified state.
   */
  function progressPendingOssification(address _yieldProvider) external returns (bool isOssificationComplete);

  /**
   * @notice Performs vendor-specific validation before the provider is registered by the YieldManager.
   * @param _registration Registration payload for the yield provider.
   */
  function validateAdditionToYieldManager(
    YieldManagerStorageLayout.YieldProviderRegistration calldata _registration
  ) external view;
}
