// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "../YieldManagerStorageLayout.sol";

/**
 * @title Contract to handle native yield operations.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IYieldManager {
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
   * @notice Emitted when funds are sent to a yield provider.
   * @param yieldProvider The yield provider that received funds.
   * @param caller Address that initiated the funding operation.
   * @param amount Total amount forwarded to the provider.
   * @param lstPrincipalRepaid Portion of the amount used to repay LST principal liabilities.
   * @param userFundsIncreased Portion of the amount remaining in the yield provider.
   */
  event YieldProviderFunded(
    address indexed yieldProvider,
    address indexed caller,
    uint256 amount,
    uint256 lstPrincipalRepaid,
    uint256 userFundsIncreased
  );

  /**
   * @notice Emitted when the withdrawal reserve sends funds to the yield manager.
   * @param amount Amount of ETH received.
   */
  event ReserveFundsReceived(uint256 amount);

  /**
   * @notice Emitted when new native yield is reported for a provider.
   * @param yieldProvider The provider that produced yield.
   * @param caller Address that reported the yield.
   * @param yieldAmount Amount of yield accounted for users.
   */
  event NativeYieldReported(address indexed yieldProvider, address indexed caller, uint256 yieldAmount);

  // TODO - Also parameter for lstPrincipalRepayment
  /**
   * @notice Emitted when funds are withdrawn from a yield provider.
   * @param yieldProvider The provider that processed the withdrawal.
   * @param caller Address initiating the withdrawal.
   * @param amountRequested Total amount requested to withdraw.
   * @param amountWithdrawn Amount actually withdrawn from the provider.
   * @param amountSentToReserve Portion sent directly to the withdrawal reserve.
   */
  event YieldProviderWithdrawal(
    address indexed yieldProvider,
    address indexed caller,
    uint256 amountRequested,
    uint256 amountWithdrawn,
    uint256 amountSentToReserve
  );

  /**
   * @notice Emitted when the withdrawal reserve is augmented by an operator.
   * @param yieldProvider The provider supplying additional liquidity.
   * @param caller Address performing the rebalance.
   * @param requestedAmount The targeted increase of the reserve.
   * @param fromYieldManager Amount provided by current YieldManager balance.
   * @param fromYieldProvider Amount withdrawn from the provider.
   */
  event WithdrawalReserveAugmented(
    address indexed yieldProvider,
    address indexed caller,
    uint256 requestedAmount,
    uint256 fromYieldManager,
    uint256 fromYieldProvider
  );

  /**
   * @notice Emitted when the withdrawal reserve is replenished permissionlessly.
   * @param yieldProvider The provider tapped for liquidity.
   * @param caller Address initiating the permissionless replenish.
   * @param fromYieldManager Amount provided by the YieldManager balance.
   * @param fromYieldProvider Amount withdrawn from the provider.
   * @param deficitBefore Initial target deficit.
   * @param deficitAfter Remaining deficit after the operation.
   */
  event WithdrawalReserveReplenished(
    address indexed yieldProvider,
    address indexed caller,
    uint256 fromYieldManager,
    uint256 fromYieldProvider,
    uint256 deficitBefore,
    uint256 deficitAfter
  );

  /**
   * @notice Emitted when a yield provider is added.
   * @param yieldProvider The provider address added to the manager.
   * @param caller Address performing the addition.
   * @param yieldProviderType Provider type identifier.
   * @param yieldProviderEntrypoint Entrypoint used for delegatecalls.
   * @param yieldProviderOssificationEntrypoint Entrypoint used once ossified.
   */
  event YieldProviderAdded(
    address indexed yieldProvider,
    address indexed caller,
    YieldManagerStorageLayout.YieldProviderType yieldProviderType,
    address yieldProviderEntrypoint,
    address yieldProviderOssificationEntrypoint
  );

  /**
   * @notice Emitted when a yield provider is removed.
   * @param yieldProvider The provider being removed.
   * @param caller Address performing the removal.
   * @param emergencyRemoval True if removal bypassed remaining funds checks.
   */
  event YieldProviderRemoved(address indexed yieldProvider, address indexed caller, bool emergencyRemoval);

  /**
   * @notice Emitted when LST principal is minted for a provider.
   * @param yieldProvider The provider on whose behalf LST was minted.
   * @param caller Address initiating the mint.
   * @param recipient LST recipient address.
   * @param amount Amount of LST minted.
   */
  event LSTMinted(address indexed yieldProvider, address indexed caller, address indexed recipient, uint256 amount);

  /**
   * @notice Emitted when the L1 message service address is updated.
   * @param oldL1MessageService Previous address.
   * @param newL1MessageService New address.
   * @param caller Address performing the update.
   */
  event L1MessageServiceUpdated(
    address indexed oldL1MessageService,
    address indexed newL1MessageService,
    address indexed caller
  );

  /**
   * @notice Emitted when the target withdrawal reserve percentage is updated.
   * @param oldTargetWithdrawalReservePercentageBps Previous target in basis points.
   * @param newTargetWithdrawalReservePercentageBps New target in basis points.
   * @param caller Address performing the update.
   */
  event TargetWithdrawalReservePercentageBpsSet(
    uint256 oldTargetWithdrawalReservePercentageBps,
    uint256 newTargetWithdrawalReservePercentageBps,
    address indexed caller
  );

  /**
   * @notice Emitted when the target withdrawal reserve amount is updated.
   * @param oldTargetWithdrawalReserveAmount Previous target amount.
   * @param newTargetWithdrawalReserveAmount New target amount.
   * @param caller Address performing the update.
   */
  event TargetWithdrawalReserveAmountSet(
    uint256 oldTargetWithdrawalReserveAmount,
    uint256 newTargetWithdrawalReserveAmount,
    address indexed caller
  );

  /**
   * @notice Emitted when staking is paused for a provider.
   * @param yieldProvider The provider whose staking was paused.
   * @param caller Address executing the pause.
   */
  event YieldProviderStakingPaused(address indexed yieldProvider, address indexed caller);

  /**
   * @notice Emitted when staking is unpaused for a provider.
   * @param yieldProvider The provider whose staking was unpaused.
   * @param caller Address executing the unpause.
   */
  event YieldProviderStakingUnpaused(address indexed yieldProvider, address indexed caller);

  /**
   * @notice Emitted when ossification is initiated for a provider.
   * @param yieldProvider The provider being ossified.
   * @param caller Address initiating ossification.
   */
  event YieldProviderOssificationInitiated(address indexed yieldProvider, address indexed caller);

  /**
   * @notice Emitted when ossification initiation is undone.
   * @param yieldProvider The provider whose ossification was reverted.
   * @param caller Address executing the revert.
   */
  event YieldProviderOssificationReverted(address indexed yieldProvider, address indexed caller);

  /**
   * @notice Emitted when ossification processing occurs.
   * @param yieldProvider The provider being processed.
   * @param caller Address executing the processing.
   * @param isOssified Whether the provider is now ossified.
   */
  event YieldProviderOssificationProcessed(address indexed yieldProvider, address indexed caller, bool isOssified);

  /**
   * @notice Emitted when a donation is routed.
   * @param yieldProvider Provider associated with the donation accounting.
   * @param caller Address providing the donation.
   * @param destination Chosen destination for the forwarded ETH.
   * @param amount Amount donated.
   */
  event DonationProcessed(address indexed yieldProvider, address indexed caller, address indexed destination, uint256 amount);

  /**
   * @notice Emitted when externally settled LST principal is reconciled.
   * @param yieldProvider Provider whose liability was reduced.
   * @param caller Address triggering the reconciliation.
   * @param amount Amount of LST principal reconciled.
   */
  event ExternalLSTPrincipalReconciled(address indexed yieldProvider, address indexed caller, uint256 amount);

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

  error UnstakeRequestPlusAvailableFundsExceedsTargetDeficit();

  error LSTWithdrawalNotAllowed();

  error AlreadyOssified();

  error OssificationNotInitiated();

  error MintLSTDisabledDuringOssification();

  error IllegalDonationAddress();

  error UnpauseStakingForbiddenWithCurrentLSTPrincipal();

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
   * @param _yieldProvider      Yield provider address.
   */
  function reportYield(address _yieldProvider) external;

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
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external;

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

  function donate(address _yieldProvider, address _destination) external payable;
}
