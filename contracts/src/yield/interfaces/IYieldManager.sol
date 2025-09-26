// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Contract to handle native yield operations.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IYieldManager {
  enum YieldProviderType {
      LIDO_STVAULT
  }

  // TODO - YieldProvider and YieldManager share the same storage, so take out?

  /**
   * @notice Supporting data for compressed calldata submission including compressed data.
   * @dev finalStateRootHash is used to set state root at the end of the data.
   */
  struct YieldProviderRegistration {
    YieldProviderType yieldProviderType;
    address yieldProviderEntrypoint;
    address yieldProviderOssificationEntrypoint;
  }

  struct YieldProviderData {
    YieldProviderRegistration registration;
    uint96 yieldProviderIndex;
    bool isStakingPaused;
    bool isOssificationInitiated;
    bool isOssified;
    // Incremented 1:1 with yieldReportedCumulative, because yieldReported becomes user funds
    uint256 userFunds;
    uint256 yieldReportedCumulative;
    uint256 pendingPermissionlessUnstake;
    // Required to socialize losses if permanent
    uint256 currentNegativeYield;
    uint256 lstLiabilityPrincipal;
  }

  /**
   * @notice Emitted when minimumWithdrawalReservePercentageBps is set.
   * @param oldMinimumWithdrawalReservePercentageBps The previous minimumWithdrawalReservePercentageBps.
   * @param newMinimumWithdrawalReservePercentageBps The new minimumWithdrawalReservePercentageBps.
   * @param caller Address which set minimumWithdrawalReservePercentageBps.
   */
  event MinimumWithdrawalReservePercentageBpsSet(
    uint256 oldMinimumWithdrawalReservePercentageBps,
    uint256 newMinimumWithdrawalReservePercentageBps,
    address indexed caller
  );

  /**
   * @notice Emitted when minimumWithdrawalReserveAmountSet is set.
   * @param oldMinimumWithdrawalReserveAmount The previous minimumWithdrawalReserveAmountSet.
   * @param newMinimumWithdrawalReserveAmount The new minimumWithdrawalReserveAmountSet.
   * @param caller Address which set minimumWithdrawalReserveAmountSet.
   */
  event MinimumWithdrawalReserveAmountSet(
    uint256 oldMinimumWithdrawalReserveAmount,
    uint256 newMinimumWithdrawalReserveAmount,
    address indexed caller
  );

  /**
   * @dev Thrown when sender is not the L1MessageService.
   */
  error SenderNotL1MessageService();

  /**
   * @dev Thrown when an operation will leave the withdrawal reserve below the minimum required amount.
   */
  error InsufficientWithdrawalReserve();

  /**
   * @dev Thrown when delegatecall to a YieldProvider fails.
   */
  error DelegateCallFailed();

  /**
   * @dev Thrown when >10000 bps is provided.
   */
  error BpsMoreThan10000();

  /**
   * @dev Thrown when caller is missing a required role.
   * @param role1 First accepted role.
   * @param role2 Second acceptable role.
   */
  error CallerMissingRole(bytes32 role1, bytes32 role2);

  error UnknownYieldProvider();

  error YieldProviderAlreadyAdded();

  error YieldProviderHasRemainingFunds();

  error StakingAlreadyPaused();

  error StakingAlreadyUnpaused();

  error TargetReservePercentageMustBeAboveMinimum();

  error TargetReserveAmountMustBeAboveMinimum();

  error WithdrawalReserveNotInDeficit();

  error SufficientAvailableFundsToCoverDeficit();

  error LSTWithdrawalNotAllowed();

  error AlreadyOssified();

  error OssificationNotInitiated();

  error MintLSTDisabledDuringOssification();

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev YIELD_PROVIDER_FUNDER_ROLE is required to execute.
   * @dev Reverts if the withdrawal reserve is below the minimum threshold.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _yieldProvider The target yield provider contract.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Receive ETH from the withdrawal reserve.
   * @dev Only accepts calls from the withdrawal reserve.
   * @dev Reverts if, after transfer, the withdrawal reserve will be below the minimum threshold.
   */
  function receiveFundsFromReserve() external payable;

  /**
   * @notice Send ETH to the L1MessageService.
   * @dev YIELD_PROVIDER_FUNDER_ROLE or YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @param _amount        The amount of ETH to send.
   */
  function transferFundsToReserve(uint256 _amount) external;

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev YIELD_REPORTER_ROLE is required to execute.
   * @dev Since the YieldManager is unaware of donations received via the L1MessageService or L2MessageService,
   *      the `_reserveDonations` parameter is required to ensure accurate yield accounting.
   * @param _yieldProvider      Yield provider address.
   * @param _totalReserveDonations   Total amount of donations received on the L1MessageService or L2MessageService.
   */
  function reportYield(address _yieldProvider, uint256 _totalReserveDonations) external;

  /**
   * @notice Request beacon chain withdrawal from specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE or RESERVE_OPERATOR_ROLE is required to execute.
   * @param _yieldProvider      Yield provider address.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external;

  /**
   * @notice Permissionlessly request beacon chain withdrawal from a specified yield provider.
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
   * @param _yieldProvider          Yield provider address.
   * @param _withdrawalParams       Provider-specific withdrawal parameters.
   * @param _withdrawalParamsProof  Merkle proof of _withdrawalParams to be verified against EIP-4788 beacon chain root.
   */
  function unstakePermissionless(
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external;

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If fund remaining, will settle any outstanding LST liabilities.
   * @param _yieldProvider          Yield provider address.
   * @param _amount                 Amount to withdraw.
   */
  function withdrawWithReserveDeficitPriorityAndLSTLiabilityPrincipalReduction(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Settles any outstanding LST liabilities, provided this does not leave the withdrawal reserve in deficit.
   * @param _yieldProvider          Yield provider address.
   * @param _amount                 Amount to withdraw.
   */
  function addToWithdrawalReserve(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Permissionlessly rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev Only available when the withdrawal is in deficit.
   * @param _yieldProvider          Yield provider address.
   */
  function replenishWithdrawalReserve(address _yieldProvider) external;

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   * @dev STAKING_PAUSER_ROLE is required to execute.
   * @param _yieldProvider          Yield provider address.
   */
  function pauseStaking(address _yieldProvider) external;

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev STAKING_UNPAUSER_ROLE is required to execute.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   * @param _yieldProvider          Yield provider address.
   */
  function unpauseStaking(address _yieldProvider) external;

  /**
   * @notice Set minimum withdrawal reserve percentage.
   * @dev Units of bps.
   * @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
   * @dev WITHDRAWAL_RESERVE_SETTER_ROLE is required to execute.
   * @param _minimumWithdrawalReservePercentageBps Minimum withdrawal reserve percentage in bps.
   */
  function setMinimumWithdrawalReservePercentageBps(uint256 _minimumWithdrawalReservePercentageBps) external;

  /**
   * @notice Set minimum withdrawal reserve.
   * @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
   * @dev WITHDRAWAL_RESERVE_SETTER_ROLE is required to execute.
   * @param _minimumWithdrawalReserveAmount Minimum withdrawal reserve amount.
   */
  function setMinimumWithdrawalReserveAmount(uint256 _minimumWithdrawalReserveAmount) external;

  function getAvailableBalanceForWithdraw(address _yieldProvider) external returns (uint256);

  function mintLST(address _yieldProvider, uint256 _amount, address _recipient) external;

  function initiateOssification(address _yieldProvider) external;

  function processPendingOssification(address _yieldProvider) external;
}
