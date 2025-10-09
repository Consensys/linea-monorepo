// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "../YieldManagerStorageLayout.sol";
import { IPauseManager } from "../../security/pausing/interfaces/IPauseManager.sol";
import { IPermissionsManager } from "../../security/access/interfaces/IPermissionsManager.sol";

/**
 * @title Contract to handle native yield operations.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IYieldManager {
  /**
   * @notice Initialization data structure for the YieldManager contract.
   * @param pauseTypeRoles The list of pause types to associate with roles.
   * @param unpauseTypeRoles The list of unpause types to associate with roles.
   * @param roleAddresses The list of role address and roles to assign permissions to.
   * @param initialL2YieldRecipients The list of initially accepted L2 yield recipient addresses.
   * @param defaultAdmin The account to be given DEFAULT_ADMIN_ROLE on initialization.
   * @param initialMinimumWithdrawalReservePercentageBps Initial minimum withdrawal reserve percentage in bps.
   * @param initialTargetWithdrawalReservePercentageBps Initial target withdrawal reserve percentage in bps.
   * @param initialMinimumWithdrawalReserveAmount Initial minimum withdrawal reserve in wei.
   * @param initialTargetWithdrawalReserveAmount Initial target withdrawal reserve in wei.
   */
  struct YieldManagerInitializationData {
    IPauseManager.PauseTypeRole[] pauseTypeRoles;
    IPauseManager.PauseTypeRole[] unpauseTypeRoles;
    IPermissionsManager.RoleAddress[] roleAddresses;
    address[] initialL2YieldRecipients;
    address defaultAdmin;
    uint16 initialMinimumWithdrawalReservePercentageBps;
    uint16 initialTargetWithdrawalReservePercentageBps;
    uint256 initialMinimumWithdrawalReserveAmount;
    uint256 initialTargetWithdrawalReserveAmount;
  }

  /**
   * @notice Struct used to represent reserve threshold updates.
   * @param minimumWithdrawalReservePercentageBps Minimum withdrawal reserve percentage in bps.
   * @param targetWithdrawalReservePercentageBps Target withdrawal reserve percentage in bps.
   * @param minimumWithdrawalReserveAmount Minimum withdrawal reserve in wei.
   * @param targetWithdrawalReserveAmount Target withdrawal reserve in wei.
   */
  struct UpdateReserveParametersConfig {
    uint16 minimumWithdrawalReservePercentageBps;
    uint16 targetWithdrawalReservePercentageBps;
    uint256 minimumWithdrawalReserveAmount;
    uint256 targetWithdrawalReserveAmount;
  }

  /**
   * @notice Emitted when ETH is received from the withdrawal reserve.
   * @param amount Amount of ETH received.
   */
  event ReserveFundsReceived(uint256 amount);

  /**
   * @notice Emitted when ETH is sent to a YieldProvider.
   * @param yieldProvider The yield provider address.
   * @param amount Gross amount transferred to the YieldProvider.
   * @param userFundsIncrement Portion of `amount` that is dedicated to staking.
   * @param lstPrincipalRepaid Portion of `amount` used to repay outstanding LST principal.
   */
  event YieldProviderFunded(
    address indexed yieldProvider,
    uint256 amount,
    uint256 userFundsIncrement,
    uint256 lstPrincipalRepaid
  );

  /**
   * @notice Emitted when new earned yield is reported for a YieldProvider.
   * @param yieldProvider The yield provider address.
   * @param l2YieldRecipient The L2 address receiving the yield.
   * @param yieldAmount Reported amount of new earned yield.
   */
  event NativeYieldReported(address indexed yieldProvider, address indexed l2YieldRecipient, uint256 yieldAmount);

  /**
   * @notice Emitted when ETH is requested from a yield provider.
   * @param yieldProvider The yield provider address.
   * @param amountRequested Amount requested to withdraw from the YieldProvider.
   * @param amountWithdrawn Actual amount withdrawn from the YieldProvider.
   * @param reserveIncrementAmount Amount routed to the reserve.
   * @param lstPrincipalPaid Amount of the YieldProvider withdrawal used to repay LST liability principal.
   */
  event YieldProviderWithdrawal(
    address indexed yieldProvider,
    uint256 amountRequested,
    uint256 amountWithdrawn,
    uint256 reserveIncrementAmount,
    uint256 lstPrincipalPaid
  );

  /**
   * @notice Emitted when the withdrawal reserve is augmented by an operator.
   * @param yieldProvider The yield provider address.
   * @param requestedAmount Amount requested to route to the reserve.
   * @param reserveIncrementAmount Total amount actually routed to the reserve.
   * @param fromYieldManager Portion filled  from the YieldManager balance.
   * @param fromYieldProvider Portion filled from the YieldProvider withdrawal.
   * @param lstPrincipalPaid Amount of the YieldProvider withdrawal used to repay LST liability principal.
   */
  event WithdrawalReserveAugmented(
    address indexed yieldProvider,
    uint256 requestedAmount,
    uint256 reserveIncrementAmount,
    uint256 fromYieldManager,
    uint256 fromYieldProvider,
    uint256 lstPrincipalPaid
  );

  /**
   * @notice Emitted when the withdrawal reserve is replenished permissionlessly.
   * @param yieldProvider The yield provider address.
   * @param targetDeficit The deficit from target threshold at the operation start.
   * @param reserveIncrementAmount Total amount routed to the reserve.
   * @param fromYieldManager Portion filled  from the YieldManager balance.
   * @param fromYieldProvider Portion filled from the YieldProvider withdrawal.
   */
  event WithdrawalReserveReplenished(
    address indexed yieldProvider,
    uint256 targetDeficit,
    uint256 reserveIncrementAmount,
    uint256 fromYieldManager,
    uint256 fromYieldProvider
  );

  /**
   * @notice Emitted when staking is paused for a YieldProvider.
   * @param yieldProvider The yield provider address.
   */
  event YieldProviderStakingPaused(address indexed yieldProvider);

  /**
   * @notice Emitted when staking is unpaused for a YieldProvider.
   * @param yieldProvider The yield provider address.
   */
  event YieldProviderStakingUnpaused(address indexed yieldProvider);

  /**
   * @notice Emitted when LST is withdrawn from a YieldProvider.
   * @param yieldProvider The yield provider address.
   * @param recipient Address that received LST.
   * @param amount Amount of LST minted (denominated in ETH).
   */
  event LSTMinted(address indexed yieldProvider, address indexed recipient, uint256 amount);

  /**
   * @notice Emitted when ossification is initiated for a YieldProvider instance.
   * @param yieldProvider The yield provider address.
   */
  event YieldProviderOssificationInitiated(address indexed yieldProvider);

  /**
   * @notice Emitted when ossification initiation is undone.
   * @param yieldProvider The yield provider address.
   */
  event YieldProviderOssificationReverted(address indexed yieldProvider);

  /**
   * @notice Emitted when a previously initiated ossification has progressed to the next stage.
   * @param yieldProvider The yield provider address.
   * @param isOssified Whether ossification has finalized.
   */
  event YieldProviderOssificationProcessed(address indexed yieldProvider, bool isOssified);

  /**
   * @notice Emitted when a donation is received.
   * @param yieldProvider YieldProvider instance whose negative yield was offset.
   * @param amount Amount of ETH donated.
   */
  event DonationProcessed(address indexed yieldProvider, uint256 amount);

  /**
   * @notice Emitted when a yield provider is added.
   * @param yieldProvider YieldProvider instance that was added to the registry.
   * @param yieldProviderVendor Specific type of YieldProvider adaptor.
   * @param primaryEntrypoint Contract used for operations when not-ossified.
   * @param ossifiedEntrypoint Contract used for operations once ossification is finalized.
   * @param receiveCaller Contract which is expected to .call() the YieldManager during withdrawals.
   */
  event YieldProviderAdded(
    address indexed yieldProvider,
    YieldManagerStorageLayout.YieldProviderVendor indexed yieldProviderVendor,
    address primaryEntrypoint,
    address indexed ossifiedEntrypoint,
    address receiveCaller
  );

  /**
   * @notice Emitted when a yield provider is removed.
   * @param yieldProvider YieldProvider instance that was removed from the registry.
   * @param emergencyRemoval True when the removal bypassed the usual requirements of
   *                         0 userFunds and 0 negativeYield.
   */
  event YieldProviderRemoved(address indexed yieldProvider, bool emergencyRemoval);

  /**
   * @notice Emitted when an L2 yield recipient address is added to the allow-list.
   * @param l2YieldRecipient The L2YieldRecipient address to add.
   */
  event L2YieldRecipientAdded(address l2YieldRecipient);

  /**
   * @notice Emitted when an L2 yield recipient address is removed from the allow-list.
   * @param l2YieldRecipient The L2YieldRecipient address to remove.
   */
  event L2YieldRecipientRemoved(address l2YieldRecipient);

  /**
   * @notice Emitted when the minimum withdrawal reserve parameters are updated.
   * @param oldMinimumWithdrawalReservePercentageBps Previous minimum expressed in basis points.
   * @param newMinimumWithdrawalReservePercentageBps New minimum expressed in basis points.
   * @param oldMinimumWithdrawalReserveAmount Previous minimum reserve in wei.
   * @param newMinimumWithdrawalReserveAmount New minimum reserve in wei.
   * @param oldTargetWithdrawalReservePercentageBps Previous target in basis points.
   * @param newTargetWithdrawalReservePercentageBps New target in basis points.
   * @param oldTargetWithdrawalReserveAmount Previous target amount.
   * @param newTargetWithdrawalReserveAmount New target amount.
   */
  event WithdrawalReserveParametersSet(
    uint256 oldMinimumWithdrawalReservePercentageBps,
    uint256 newMinimumWithdrawalReservePercentageBps,
    uint256 oldMinimumWithdrawalReserveAmount,
    uint256 newMinimumWithdrawalReserveAmount,
    uint256 oldTargetWithdrawalReservePercentageBps,
    uint256 newTargetWithdrawalReservePercentageBps,
    uint256 oldTargetWithdrawalReserveAmount,
    uint256 newTargetWithdrawalReserveAmount
  );

  /**
   * @dev Thrown when delegatecall to a YieldProvider instance fails.
   */
  error DelegateCallFailed();

  /**
   * @dev Thrown an unknown YieldProvider address is used.
   */
  error UnknownYieldProvider();

  /**
   * @dev Thrown when querying the yield provider registry with an out-of-bounds index.
   * @param index Supplied registry index.
   * @param count Current number of registered yield providers.
   */
  error YieldProviderIndexOutOfBounds(uint256 index, uint256 count);

  /**
   * @dev Thrown when an unknown L2YieldRecipient address is used.
   */
  error UnknownL2YieldRecipient();

  /**
   * @dev Thrown when sender is not the L1MessageService.
   */
  error SenderNotL1MessageService();

  /**
   * @dev Thrown when an operation will leave the withdrawal reserve below the minimum threshold.
   */
  error InsufficientWithdrawalReserve();

  /**
   * @dev Thrown when caller is missing a required role.
   * @param role1 First accepted role.
   * @param role2 Second acceptable role.
   */
  error CallerMissingRole(bytes32 role1, bytes32 role2);

  /**
   * @dev Thrown when a permissionless rebalance operation is attempted when the withdrawal reserve is not in deficit.
   */
  error WithdrawalReserveNotInDeficit();

  /**
   * @dev Thrown when a permissionless unstake request exceeds the minimum required amount to restore the reserve to the target threshold,
   *      taking into consideration available funds in the system that can be routed to the reserve.
   */
  error PermissionlessUnstakeRequestPlusAvailableFundsExceedsTargetDeficit();

  /**
   * @dev Thrown when pausing staking for a YieldProvider which is currently paused.
   */
  error StakingAlreadyPaused();

  /**
   * @dev Thrown when resuming staking for a YieldProvider which is currently unpaused.
   */
  error StakingAlreadyUnpaused();

  /**
   * @dev Thrown when attempting to unpause staking with an outstanding LST principal liability.
   */
  error UnpauseStakingForbiddenWithCurrentLSTLiability();

  /**
   * @dev Thrown when attempting to unpause staking when ossification has been initiated.
   */
  error UnpauseStakingForbiddenDuringPendingOssification();

  /**
   * @dev Thrown when LST withdrawal is attempted through another route other than L1MessageService.claimMessageWithProofAndWithdrawLST.
   */
  error LSTWithdrawalNotAllowed();

  /**
   * @dev Thrown when LST withdrawal request exceeds the user funds in the YieldProvider.
   */
  error LSTWithdrawalExceedsYieldProviderFunds();

  /**
   * @dev Thrown when attempting to undo or progress an ossification process that was not previously initiated.
   */
  error OssificationNotInitiated();

  /**
   * @dev Thrown when attempting to initiate or progress the ossification process for a YieldProvider that is already ossified.
   */
  error AlreadyOssified();

  /**
   * @dev Thrown when the YieldManager receives ETH from an unexpected sender.
   */
  error UnexpectedReceiveCaller();

  /**
   * @dev Thrown when adding a YieldProvider instance that was previously registered
   */
  error YieldProviderAlreadyAdded();

  /**
   * @dev Thrown when removing a YieldProvider with remaining user funds.
   */
  error YieldProviderHasRemainingFunds();

  /**
   * @dev Thrown when removing a YieldProvider with remaining negative yield.
   */
  error YieldProviderHasRemainingNegativeYield();

  /**
   * @dev Thrown when adding an L2YieldRecipient that has previously been added to the allowlist.
   */
  error L2YieldRecipientAlreadyAdded();

  /**
   * @dev Thrown when >10000 bps is provided.
   */
  error BpsMoreThan10000();

  /**
   * @dev Thrown when the target reserve percentage will be set below the minimum percentage.
   */
  error TargetReservePercentageMustBeAboveMinimum();

  /**
   * @dev Thrown when the target reserve threshold amount will be set below the minimum amount.
   */
  error TargetReserveAmountMustBeAboveMinimum();

  /**
   * @notice Returns the total ETH in the native yield system.
   * @dev Sums the withdrawal reserve, YieldManager balance, and capital deployed into yield providers.
   * @return totalSystemBalance Total system balance in wei.
   */
  function getTotalSystemBalance() external view returns (uint256 totalSystemBalance);

  /**
   * @notice Returns the effective minimum withdrawal reserve considering both percentage and absolute amount configurations.
   * @return minimumWithdrawalReserve Effective minimum reserve in wei.
   */
  function getEffectiveMinimumWithdrawalReserve() external view returns (uint256 minimumWithdrawalReserve);

  /**
   * @notice Returns the effective target withdrawal reserve considering both percentage and absolute amount configurations.
   * @return targetWithdrawalReserve Effective target reserve in wei.
   */
  function getEffectiveTargetWithdrawalReserve() external view returns (uint256 targetWithdrawalReserve);

  /**
   * @notice Returns the shortfall between the minimum reserve threshold and the current reserve balance.
   * @return minimumReserveDeficit Amount of ETH required to meet the minimum reserve, or zero if already satisfied.
   */
  function getMinimumReserveDeficit() external view returns (uint256 minimumReserveDeficit);

  /**
   * @notice Returns the shortfall between the target reserve threshold and the current reserve balance.
   * @return targetReserveDeficit Amount of ETH required to meet the target reserve, or zero if already satisfied.
   */
  function getTargetReserveDeficit() external view returns (uint256 targetReserveDeficit);

  /**
   * @return bool True if the withdrawal reserve balance is below the effective minimum threshold.
   */
  function isWithdrawalReserveBelowMinimum() external view returns (bool);

  /**
   * @param _l2YieldRecipient The L2YieldRecipient address.
   * @return bool True if the L2YieldRecipient is on the allowlist.
   */
  function isL2YieldRecipientKnown(address _l2YieldRecipient) external view returns (bool);

  /**
   * @param _yieldProvider The YieldProvider address.
   * @return bool True if the YieldProvider is registered.
   */
  function isYieldProviderKnown(address _yieldProvider) external view returns (bool);

  /**
   * @notice Returns the number of registered yield provider adaptor contracts.
   * @return count Total number of yield providers tracked by the YieldManager.
   */
  function yieldProviderCount() external view returns (uint256 count);

  /**
   * @notice Returns the yield provider address stored at a specific index in the registry.
   * @dev Uses 1-based indexing: 1 returns the first element.
   * @dev 0 index is the sentinel value for a yield provider not being registered.
   * @param _index Index of the yield provider to query.
   * @return yieldProvider Yield provider adaptor address stored at the supplied index.
   *                       - Zero address if yield provider not registered.
   */
  function yieldProviderByIndex(uint256 _index) external view returns (address yieldProvider);

  /**
   * @notice Returns the full state for a registered yield provider.
   * @param _yieldProvider Yield provider adaptor address to query.
   * @return yieldProviderData Struct containing the yield provider data.
   */
  function getYieldProviderData(
    address _yieldProvider
  ) external view returns (YieldManagerStorageLayout.YieldProviderStorage memory yieldProviderData);

  /**
   * @notice Returns the tracked user funds for a specific yield provider.
   * @param _yieldProvider Yield provider adaptor address to query.
   * @return funds Amount of user funds currently attributed to the yield provider.
   */
  function userFunds(address _yieldProvider) external view returns (uint256 funds);

  /**
   * @notice Returns whether staking is currently paused for a specific yield provider.
   * @param _yieldProvider Yield provider adaptor address to query.
   * @return isPaused True if staking is paused on the yield provider.
   */
  function isStakingPaused(address _yieldProvider) external view returns (bool isPaused);

  /**
   * @notice Returns whether a yield provider has been fully ossified.
   * @param _yieldProvider Yield provider adaptor address to query.
   * @return bool True if the ossification process has completed for the yield provider.
   */
  function isOssified(address _yieldProvider) external view returns (bool);

  /**
   * @notice Returns whether a yield provider has initiated ossification.
   * @param _yieldProvider Yield provider adaptor address to query.
   * @return isInitiated True if ossification has been initiated for the yield provider.
   */
  function isOssificationInitiated(address _yieldProvider) external view returns (bool isInitiated);

  /**
   * @notice Returns the aggregate user funds deployed across all registered yield providers.
   * @return totalUserFunds Aggregate amount of user funds currently deployed across yield providers.
   */
  function userFundsInYieldProvidersTotal() external view returns (uint256 totalUserFunds);

  /**
   * @notice Returns the amount of ETH expected from pending permissionless unstake requests.
   * @return pendingUnstake Amount of ETH pending arrival from the beacon chain via permissionless unstaking.
   */
  function pendingPermissionlessUnstake() external view returns (uint256 pendingUnstake);

  /**
   * @param _yieldProvider The yield provider address.
   * @return withdrawableAmount Amount of ETH that can be instantly withdrawn from the YieldProvider.
   */
  function withdrawableValue(address _yieldProvider) external returns (uint256);

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
   * @notice Send ETH to the specified YieldProvider instance.
   * @dev YIELD_PROVIDER_FUNDER_ROLE is required to execute.
   * @dev Reverts if the withdrawal reserve is below the minimum threshold.
   * @dev ETH sent to the YieldProvider will be eagerly used to settle any outstanding LST liabilities.
   * @param _yieldProvider The target yield provider contract.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Report newly accrued yield for the YieldProvider since the last report.
   * @dev YIELD_REPORTER_ROLE is required to execute.
   * @dev Reported yield excludes amounts reserved to pay system obligations.
   * @param _yieldProvider      The yield provider address.
   * @param _l2YieldRecipient   The L2YieldRecipient address.
   * @return newReportedYield New net yield (denominated in ETH) since the prior report.
   */
  function reportYield(address _yieldProvider, address _l2YieldRecipient) external returns (uint256 newReportedYield);

  /**
   * @notice Request beacon chain withdrawal from specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE or RESERVE_OPERATOR_ROLE is required to execute.
   * @param _yieldProvider      Yield provider address.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable;

  /**
   * @notice Permissionlessly request beacon chain withdrawal from a specified yield provider.
   * @dev    Callable only when the withdrawal reserve is in deficit.
   * @dev    The permissionless unstake amount is capped to the remaining target deficit that
   *         cannot be covered by other liquidity sources:
   *
   *         PERMISSIONLESS_UNSTAKE_AMOUNT ≤
   *           TARGET_DEFICIT
   *         - YIELD_PROVIDER_BALANCE
   *         - YIELD_MANAGER_BALANCE
   *         - PENDING_PERMISSIONLESS_UNSTAKE
   *
   * @dev PENDING_PERMISSIONLESS_UNSTAKE will be greedily reduced with i.) donations or ii.) future withdrawals from the YieldProvider
   * @param _yieldProvider          Yield provider address.
   * @param _withdrawalParams       Provider-specific withdrawal parameters.
   * @param _withdrawalParamsProof  Data containing merkle proof of _withdrawalParams to be verified against EIP-4788 beacon chain root.
   * @return maxUnstakeAmount       Maximum amount expected to be withdrawn from the beacon chain.
   *                                - Cannot efficiently get exact amount as relevant state and computation is located in the consensus client,
   *                                and not the execution layer.
   */
  function unstakePermissionless(
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external payable returns (uint256 maxUnstakeAmount);

  /**
   * @notice Withdraw ETH from a YieldProvider.
   * @dev YIELD_PROVIDER_UNSTAKER_ROLE is required to execute.
   * @dev This function proactively allocates withdrawn funds in the following priority:
   *      1. If the withdrawal reserve is below the target threshold, ETH is routed to the reserve
   *      to restore the deficit.
   *      2. If there is an outstanding LST liability, it will be paid.
   *      3. YieldManager will keep the remainder.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount to withdraw..
   */
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev This function proactively allocates withdrawn funds in the following priority:
   *      1. If the withdrawal reserve is below the target threshold, ETH is routed to the reserve
   *      to restore the deficit.
   *      2. If there is no remaining target deficit and there is an outstanding LST liability, it will be paid.
   *      3. The remainder will be sent to the withdrawal reserve.
   * @param _yieldProvider          Yield provider address.
   * @param _amount                 Amount to rebalance from the YieldManager and specified YieldProvider.
   */
  function addToWithdrawalReserve(address _yieldProvider, uint256 _amount) external;

  /**
   * @notice Permissionlessly top up the withdrawal reserve to the target threshold using available liquidity.
   * @dev Callable only when the reserve is below the effective minimum threshold.
   * @dev The function first spends the YieldManager's balance to clear the target threshold deficit.
   * @dev If a target deficit still remains, then it will withdraw from the specified YieldProvider.
   * @param _yieldProvider The yield provider address.
   */
  function replenishWithdrawalReserve(address _yieldProvider) external;

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   * @dev STAKING_PAUSER_ROLE is required to execute.
   * @param _yieldProvider The yield provider address.
   */
  function pauseStaking(address _yieldProvider) external;

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev STAKING_UNPAUSER_ROLE is required to execute.
   * @dev Will revert if:
   *      - The withdrawal reserve is in deficit, or
   *      - There is an existing LST liability, or
   *      - Ossification has been initiated or finalized.
   * @param _yieldProvider The yield provider address.
   */
  function unpauseStaking(address _yieldProvider) external;

  /**
   * @notice Withdraw LST from a specified YieldProvider instance.
   * @dev Callable only by the L1MessageService
   * @dev Will pause staking to mitigate further reserve deficits.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of LST (ETH-denominated) to withdraw.
   * @param _recipient L1 address to receive the LST.
   */
  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external;

  /**
   * @notice Initiate the ossification sequence for a provider.
   * @dev Will pause beacon chain staking and LST withdrawals.
   * @dev Re-calling this function after a prior initiation is allowed.
   * @param _yieldProvider The yield provider address.
   */
  function initiateOssification(address _yieldProvider) external;

  /**
   * @notice Revert a previously initiated ossification.
   * @param _yieldProvider The yield provider address.
   */
  function undoInitiateOssification(address _yieldProvider) external;

  /**
   * @notice Progress an initiated ossification process.
   * @param _yieldProvider The yield provider address.
   * @return isOssificationComplete True if ossification is finalized.
   */
  function processPendingOssification(address _yieldProvider) external returns (bool isOssificationComplete);

  /**
   * @notice Donate ETH that offsets a specified yield provider's negative yield.
   * @dev Donations are forwarded to the withdrawal reserve.
   * @dev The donate() function is located on the YieldManager because it is otherwise tricky to track donations
   *      to offset negative yield for a specific yield provider.
   * @dev `pendingPermissionlessUnstake` is greedily decremented against incoming donations.
   * @param _yieldProvider The yield provider address.
   */
  function donate(address _yieldProvider) external payable;

  /**
   * @notice Register a new YieldProvider adaptor instance.
   * @param _yieldProvider The yield provider address.
   * @param _registration Struct representing expected information to add a YieldProvider adaptor instance.
   */
  function addYieldProvider(
    address _yieldProvider,
    YieldManagerStorageLayout.YieldProviderRegistration calldata _registration
  ) external;

  /**
   * @notice Remove a YieldProvider instance from the YieldManager.
   * @dev Has safety checks to ensure that there is no remaining user funds or negative yield on the YieldProvider.
   * @param _yieldProvider The yield provider address.
   */
  function removeYieldProvider(address _yieldProvider) external;

  /**
   * @notice Emergency remove a YieldProvider instance from the YieldManager, skipping the regular safety checks.
   * @dev Without this function, newly reported yield can prevent deregistration of the YieldProvider.
   * @param _yieldProvider The yield provider address.
   */
  function emergencyRemoveYieldProvider(address _yieldProvider) external;

  /**
   * @notice Add an address to the allowlist of L2YieldRecipients.
   * @dev SET_L2_YIELD_RECIPIENT_ROLE is required to execute.
   * @param _L2YieldRecipient L2YieldRecipient address.
   */
  function addL2YieldRecipient(address _L2YieldRecipient) external;

  /**
   * @notice Remove an address from the allow-list of L2YieldRecipients.
   * @dev SET_L2_YIELD_RECIPIENT_ROLE is required to execute.
   * @param _L2YieldRecipient L2YieldRecipient address.
   */
  function removeL2YieldRecipient(address _L2YieldRecipient) external;

  /**
   * @notice Update withdrawal reserve parameters
   * @dev WITHDRAWAL_RESERVE_SETTER_ROLE is required to execute.
   * @param _params Data used to update withdrawal reserve parameters.
   */
  function setWithdrawalReserveParameters(UpdateReserveParametersConfig calldata _params) external;
}
