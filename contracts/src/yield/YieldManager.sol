// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "./YieldManagerStorageLayout.sol";
import { IYieldManager } from "./interfaces/IYieldManager.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { ILineaNativeYieldExtension } from "./interfaces/ILineaNativeYieldExtension.sol";
import { YieldManagerPauseManager } from "../security/pausing/YieldManagerPauseManager.sol";
import { Math256 } from "../libraries/Math256.sol";

/**
 * @title Contract to handle native yield operations.
 * @author ConsenSys Software Inc.
 * @dev Sole writer to YieldManagerStorageLayout.
 * @custom:security-contact security-report@linea.build
 */
contract YieldManager is YieldManagerPauseManager, YieldManagerStorageLayout, IYieldManager, IGenericErrors {
  /// @notice The role required to send ETH to a yield provider.
  bytes32 public constant YIELD_PROVIDER_FUNDER_ROLE = keccak256("YIELD_PROVIDER_FUNDER_ROLE");

  /// @notice The role required to unstake ETH from a yield provider.
  bytes32 public constant YIELD_PROVIDER_UNSTAKER_ROLE = keccak256("YIELD_PROVIDER_UNSTAKER_ROLE");

  /// @notice The role required to request a yield report.
  bytes32 public constant YIELD_REPORTER_ROLE = keccak256("YIELD_REPORTER_ROLE");

  /// @notice The role required to rebalance ETH between the withdrawal reserve and yield provider/s.
  bytes32 public constant RESERVE_OPERATOR_ROLE = keccak256("RESERVE_OPERATOR_ROLE");

  /// @notice The role required to pause beacon chain staking for yield provider/s that support this operation.
  bytes32 public constant STAKING_PAUSER_ROLE = keccak256("STAKING_PAUSER_ROLE");

  /// @notice The role required to unpause beacon chain staking for yield provider/s that support this operation.
  bytes32 public constant UNSTAKING_PAUSER_ROLE = keccak256("UNSTAKING_PAUSER_ROLE");

  /// @notice The role required to execute ossification functions.
  bytes32 public constant OSSIFIER_ROLE = keccak256("OSSIFIER_ROLE");

  /// @notice The role required to set withdrawal reserve parameters.
  bytes32 public constant WITHDRAWAL_RESERVE_SETTER_ROLE = keccak256("WITHDRAWAL_RESERVE_SETTER_ROLE");

  /// @notice The role required to add and remove yield providers.
  bytes32 public constant YIELD_PROVIDER_SETTER = keccak256("YIELD_PROVIDER_SETTER");

  /// @notice 100% in BPS.
  uint256 constant MAX_BPS = 10000;

  /// @notice The L1MessageService address.
  function l1MessageService() public view returns (address) {
      return _getYieldManagerStorage()._l1MessageService;
  }

  /// @notice Minimum withdrawal reserve percentage in bps.
  /// @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
  function minimumWithdrawalReservePercentageBps() public view returns (uint256) {
      return _getYieldManagerStorage()._minimumWithdrawalReservePercentageBps;
  }

  /// @notice Minimum withdrawal reserve amount.
  /// @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
  function minimumWithdrawalReserveAmount() public view returns (uint256) {
      return _getYieldManagerStorage()._minimumWithdrawalReserveAmount;
  }

  /**
   * @notice Initialises the YieldManager.
   */
  function initialize() external initializer {
  }

  modifier onlyKnownYieldProvider(address _yieldProvider) {
      if (_getYieldProviderStorage(_yieldProvider).yieldProviderIndex != 0) {
        revert UnknownYieldProvider();
      }
      _;
  }

 function getWithdrawalReserveBalance() external view returns (uint256 withdrawalReserveBalance) {
    withdrawalReserveBalance = l1MessageService().balance;
  }

  function getTotalSystemBalance() external view returns (uint256 totalSystemBalance) {
    (totalSystemBalance,) = _getTotalSystemBalance();
  }

  function _getTotalSystemBalance() internal view returns (uint256 totalSystemBalance, uint256 cachedL1MessageServiceBalance) {
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    cachedL1MessageServiceBalance = $._l1MessageService.balance;
    totalSystemBalance = cachedL1MessageServiceBalance + address(this).balance + $._userFundsInYieldProvidersTotal;
  }

  function getMinimumWithdrawalReserveByPercentage() external view returns (uint256 minimumWithdrawalReserveByPercentage) {
    (minimumWithdrawalReserveByPercentage,) = _getMinimumWithdrawalReserveByPercentage();
  }

  function _getMinimumWithdrawalReserveByPercentage() internal view returns (uint256 minimumWithdrawalReserveByPercentage, uint256 cachedL1MessageServiceBalance) {
    uint256 totalSystemBalance;
    (totalSystemBalance, cachedL1MessageServiceBalance) = _getTotalSystemBalance();
    minimumWithdrawalReserveByPercentage = totalSystemBalance * _getYieldManagerStorage()._minimumWithdrawalReservePercentageBps / MAX_BPS;
  }

  function getTargetWithdrawalReserveByPercentage() external view returns (uint256 targetWithdrawalReserveByPercentage) {
    (targetWithdrawalReserveByPercentage,) = _getTargetWithdrawalReserveByPercentage();
  }

  function _getTargetWithdrawalReserveByPercentage() internal view returns (uint256 targetWithdrawalReserveByPercentage, uint256 cachedL1MessageServiceBalance) {
    uint256 totalSystemBalance;
    (totalSystemBalance, cachedL1MessageServiceBalance) = _getTotalSystemBalance();
    targetWithdrawalReserveByPercentage = totalSystemBalance * _getYieldManagerStorage()._targetWithdrawalReservePercentageBps / MAX_BPS;
  }

  function getMinimumWithdrawalReserve() external view returns (uint256 minimumWithdrawalReserve) {
    (minimumWithdrawalReserve,) = _getMinimumWithdrawalReserve();
  }

  /// @notice Get effective minimum withdrawal reserve
  /// @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
  function _getMinimumWithdrawalReserve() internal view returns (uint256 minimumWithdrawalReserve, uint256 cachedL1MessageServiceBalance) {
      uint256 minimumWithdrawalReserveByPercentage;
      (minimumWithdrawalReserveByPercentage, cachedL1MessageServiceBalance) = _getMinimumWithdrawalReserveByPercentage();
      minimumWithdrawalReserve = Math256.min(minimumWithdrawalReserveByPercentage, _getYieldManagerStorage()._minimumWithdrawalReserveAmount);
  }

  function getTargetWithdrawalReserve() external view returns (uint256 targetWithdrawalReserve) {
    (targetWithdrawalReserve,) = _getTargetWithdrawalReserve();
  }

  function _getTargetWithdrawalReserve() internal view returns (uint256 targetWithdrawalReserve, uint256 cachedL1MessageServiceBalance) {
      uint256 targetWithdrawalReserveByPercentage;
      (targetWithdrawalReserveByPercentage, cachedL1MessageServiceBalance) = _getTargetWithdrawalReserveByPercentage();
      targetWithdrawalReserve = Math256.min(targetWithdrawalReserveByPercentage, _getYieldManagerStorage()._targetWithdrawalReserveAmount);
  }

  function getMinimumReserveDeficit() public view returns (uint256 minimumReserveDeficit) {
    (uint256 minimumWithdrawalReserve, uint256 cachedL1MessageServiceBalance) = _getMinimumWithdrawalReserve();
    minimumReserveDeficit = Math256.safeSub(minimumWithdrawalReserve, cachedL1MessageServiceBalance);
  }

  function getTargetReserveDeficit() public view returns (uint256 targetReserveDeficit) {
    (uint256 targetWithdrawalReserve, uint256 cachedL1MessageServiceBalance) = _getTargetWithdrawalReserve();
    targetReserveDeficit = Math256.safeSub(targetWithdrawalReserve, cachedL1MessageServiceBalance);
  }

  /// @notice Returns true if withdrawal reserve balance is below effective required minimum.
  /// @dev We are doing duplicate BALANCE opcode call, but how to remove duplicate call while maintaining readability?
  function isWithdrawalReserveBelowMinimum() public view returns (bool) {
    (uint256 minimumWithdrawalReserve, uint256 cachedL1MessageServiceBalance) = _getMinimumWithdrawalReserve();
    return cachedL1MessageServiceBalance < minimumWithdrawalReserve;
  }

  function withdrawableValue(address _yieldProvider) public onlyKnownYieldProvider(_yieldProvider) returns (uint256) {
    (bool success, bytes memory data) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.withdrawableValue, ()
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    return abi.decode(data, (uint256));
  }


  /**
   * @notice Receive ETH from the withdrawal reserve.
   * @dev Only accepts calls from the withdrawal reserve.
   * @dev It is possible for an arbitrary user to call this via `L1.claimMessage()`,
   *    this results in a user effectively donating their ETH balance to YieldManager.
   *    This does not violate the safety property of user principal protection, as the user has forfeited their principal.
   * @dev Reverts if, after transfer, the withdrawal reserve will be below the minimum threshold.
   */
  function receiveFundsFromReserve() external payable {
    if (msg.sender != l1MessageService()) {
      revert SenderNotL1MessageService();
    }
    if (isWithdrawalReserveBelowMinimum()) {
        revert InsufficientWithdrawalReserve();
    }
    emit ReserveFundsReceived(msg.value);
  }

  /**
   * @notice Send ETH to the L1MessageService.
   * @dev RESERVE_OPERATOR_ROLE or YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @param _amount        The amount of ETH to send.
   */
  function transferFundsToReserve(uint256 _amount) external {
    if (!hasRole(RESERVE_OPERATOR_ROLE, msg.sender) && !hasRole(YIELD_PROVIDER_FUNDER_ROLE, msg.sender)) {
      revert CallerMissingRole(RESERVE_OPERATOR_ROLE, YIELD_PROVIDER_FUNDER_ROLE);
    }
    ILineaNativeYieldExtension(l1MessageService()).fund{value: _amount}();
    // Destination will emit the event.
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev YIELD_PROVIDER_FUNDER_ROLE is required to execute.
   * @dev Reverts if the withdrawal reserve is below the minimum threshold.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _yieldProvider The target yield provider contract.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external onlyKnownYieldProvider(_yieldProvider) {
    _fundYieldProvider(_yieldProvider, _amount);
    // Do LST repayment
    uint256 lstPrincipalRepayment = _payLSTPrincipal(_yieldProvider, _amount);
    uint256 amountRemaining = _amount - lstPrincipalRepayment;
    _getYieldManagerStorage()._userFundsInYieldProvidersTotal += amountRemaining;
    _getYieldProviderStorage(_yieldProvider).userFunds += amountRemaining;
    emit YieldProviderFunded(_yieldProvider, msg.sender, _amount, lstPrincipalRepayment, amountRemaining);
  }

  function _fundYieldProvider(address _yieldProvider, uint256 _amount) internal {
    if (isWithdrawalReserveBelowMinimum()) {
      revert InsufficientWithdrawalReserve();
    }
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.fundYieldProvider, (_amount)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
  }

  function _payLSTPrincipal(address _yieldProvider, uint256 _maxAvailableRepaymentETH) internal returns (uint256 lstPrincipalPaid) {
    (bool success, bytes memory data) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.payLSTPrincipal, (_maxAvailableRepaymentETH)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    (lstPrincipalPaid) = abi.decode(data, (uint256));
  }

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev YIELD_REPORTER_ROLE is required to execute.
   * @param _yieldProvider      Yield provider address.
   */
  function reportYield(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) returns (uint256 newReportedYield) {
    (bool success, bytes memory data) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.reportYield, ()
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    (newReportedYield) = abi.decode(data, (uint256));
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    $$.userFunds += newReportedYield;
    $$.yieldReportedCumulative += newReportedYield;
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    $._userFundsInYieldProvidersTotal += newReportedYield;
    ILineaNativeYieldExtension(l1MessageService()).reportNativeYield(newReportedYield, $._l2YieldRecipient);
    emit NativeYieldReported(_yieldProvider, msg.sender, newReportedYield);
  }

  /**
   * @notice Request beacon chain withdrawal from specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE or RESERVE_OPERATOR_ROLE is required to execute.
   * @param _yieldProvider      Yield provider address.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable onlyKnownYieldProvider(_yieldProvider) {
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.unstake, (_withdrawalParams)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    // Event emitted by YieldProvider which has provider-specific decoding of _withdrawalParams
  }

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
   * @dev Any future ETH transfer to the L1MessageService will reduce PENDING_PERMISSIONLESS_UNSTAKE.
   * @dev Validates (validatorPubkey, validatorBalance, validatorWithdrawalCredential) against EIP-4788 beacon chain root.
   * @param _yieldProvider          Yield provider address.
   * @param _withdrawalParams       Provider-specific withdrawal parameters.
   * @param _withdrawalParamsProof  Merkle proof of _withdrawalParams to be verified against EIP-4788 beacon chain root.
   */
  function unstakePermissionless(
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external payable onlyKnownYieldProvider(_yieldProvider) returns (uint256 maxUnstakeAmount) {
    if (!isWithdrawalReserveBelowMinimum()) {
      revert WithdrawalReserveNotInDeficit();
    }
    (bool success, bytes memory data) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.unstakePermissionless, (_withdrawalParams, _withdrawalParamsProof)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    (maxUnstakeAmount) = abi.decode(data, (uint256));
   _validateUnstakePermissionlessAmount(_yieldProvider, maxUnstakeAmount);
   _getYieldManagerStorage()._pendingPermissionlessUnstake += maxUnstakeAmount;
    // Event emitted by YieldProvider which has provider-specific decoding of _withdrawalParams
  }

  function _validateUnstakePermissionlessAmount(address _yieldProvider, uint256 _maxUnstakeAmount) internal {
    uint256 targetDeficit = getTargetReserveDeficit();
    uint256 availableFundsToSettleTargetDeficit = address(this).balance + withdrawableValue(_yieldProvider) + _getYieldManagerStorage()._pendingPermissionlessUnstake;
    if (availableFundsToSettleTargetDeficit + _maxUnstakeAmount > targetDeficit) {
      revert UnstakeRequestPlusAvailableFundsExceedsTargetDeficit();
    }
  }

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If funds remaining, will settle any outstanding LST liabilities.
   * @param _yieldProvider          Yield provider address.
   * @param _amount                 Amount to withdraw.
   */
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external onlyKnownYieldProvider(_yieldProvider) {
    uint256 targetDeficit = getTargetReserveDeficit();
    // Withdraw from Vault -> YieldManager
    (uint256 withdrawnFromProvider,) = _withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
      _yieldProvider,
      _amount,
      address(this),
      targetDeficit
    );
    // Send funds to L1MessageService if targetDeficit
    if (targetDeficit > 0) {
      ILineaNativeYieldExtension(l1MessageService()).fund{value: targetDeficit}();
    }
    emit YieldProviderWithdrawal(
      _yieldProvider,
      msg.sender,
      _amount,
      withdrawnFromProvider,
      targetDeficit
    );
  }

  function _withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(address _yieldProvider, uint256 _amount, address _recipient, uint256 _targetDeficit) internal returns (uint256 withdrawAmount, uint256 lstPrincipalPaid) {
    uint256 availableFundsForLSTLiabilityPayment = Math256.safeSub(_amount, _targetDeficit);
    withdrawAmount = _amount;
    if (availableFundsForLSTLiabilityPayment > 0) {
      lstPrincipalPaid -= _payLSTPrincipal(_yieldProvider, availableFundsForLSTLiabilityPayment);
      withdrawAmount -= lstPrincipalPaid;
    // Will remain in target deficit after withdrawal
    } else {
      _pauseStakingIfNotAlready(_yieldProvider);
    }
    _withdrawFromYieldProvider(_yieldProvider, withdrawAmount, _recipient);
  }

  function _withdrawFromYieldProvider(address _yieldProvider, uint256 _amount, address _recipient) internal {
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.withdrawFromYieldProvider, (_amount, _recipient)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    _getYieldProviderStorage(_yieldProvider).userFunds -= _amount;
    _getYieldManagerStorage()._userFundsInYieldProvidersTotal -= _amount;
    // Greedily reduce pendingPermissionlessUnstake with every withdrawal made from the yield provider.
    _decrementPendingPermissionlessUnstake(_amount);
  }

  function _decrementPendingPermissionlessUnstake(uint256 _amount) internal {
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    uint256 pendingPermissionlessUnstake = $._pendingPermissionlessUnstake;
    if (pendingPermissionlessUnstake == 0) return;
    $._pendingPermissionlessUnstake = Math256.safeSub(pendingPermissionlessUnstake, _amount);
  }

  /**
   * @notice Rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Settles any outstanding LST liabilities, provided this does not leave the withdrawal reserve in deficit.
   * @param _yieldProvider          Yield provider address.
   * @param _amount                 Amount to withdraw.
   */
  function addToWithdrawalReserve(address _yieldProvider, uint256 _amount) external onlyKnownYieldProvider(_yieldProvider) {
    uint256 targetRebalanceAmount = _amount;
    address cachedL1MessageService = l1MessageService();
    // Try to settle rebalance amount from yieldManager funds
    uint256 yieldManagerBalance = address(this).balance;
    uint256 fromYieldManager;
    if (yieldManagerBalance > 0) {
      uint256 transferAmount = Math256.min(yieldManagerBalance, targetRebalanceAmount);
      ILineaNativeYieldExtension(cachedL1MessageService).fund{value: transferAmount}();
      targetRebalanceAmount -= transferAmount;
      fromYieldManager = transferAmount;
    }
    // Then meet settle from Vault withdrawal
    uint256 fromYieldProvider;
    if (targetRebalanceAmount > 0) {
      (fromYieldProvider,) = _withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
        _yieldProvider,
        targetRebalanceAmount,
        cachedL1MessageService,
        getTargetReserveDeficit()
      );
    }
    emit WithdrawalReserveAugmented(
      _yieldProvider,
      msg.sender,
      _amount,
      fromYieldManager,
      fromYieldProvider
    );
  }

  /**
   * @notice Permissionlessly rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev Only available when the withdrawal is in deficit.
   * @dev Subtle differences from addToWithdrawalReserve
   *      - permissionless
   *      - does not accept an _amount param and instead routes maximum available funds. Thus not allowed to fail.
   *      - No LST repayments
   * @dev Will rebalance to target
   * @param _yieldProvider          Yield provider address.
   */
  function replenishWithdrawalReserve(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    if (!isWithdrawalReserveBelowMinimum()) {
      revert WithdrawalReserveNotInDeficit();
    }
    uint256 remainingTargetDeficit = getTargetReserveDeficit();
    address cachedL1MessageService = l1MessageService();
    uint256 initialDeficit = remainingTargetDeficit;
    // Try to meet targetDeficit from yieldManager
    uint256 yieldManagerBalance = address(this).balance;
    uint256 fromYieldManager;
    if (yieldManagerBalance > 0) {
      uint256 transferAmount = Math256.min(yieldManagerBalance, remainingTargetDeficit);
      ILineaNativeYieldExtension(cachedL1MessageService).fund{value: transferAmount}();
      remainingTargetDeficit -= transferAmount;
      fromYieldManager = transferAmount;
    }
    // Try to meet remaining targetDeficit by yieldProvider withdraw
    uint256 availableYieldProviderWithdrawBalance = withdrawableValue(_yieldProvider);
    uint256 fromYieldProvider;
    if (remainingTargetDeficit > 0 && availableYieldProviderWithdrawBalance > 0) {
      uint256 withdrawAmount = Math256.min(availableYieldProviderWithdrawBalance, remainingTargetDeficit);
      _withdrawFromYieldProvider(_yieldProvider, withdrawAmount, cachedL1MessageService);
      remainingTargetDeficit -= withdrawAmount;
      fromYieldProvider = withdrawAmount;
    }
    if (remainingTargetDeficit > 0) {
      _pauseStakingIfNotAlready(_yieldProvider);
    }
    emit WithdrawalReserveReplenished(
      _yieldProvider,
      msg.sender,
      fromYieldManager,
      fromYieldProvider,
      initialDeficit,
      remainingTargetDeficit
    );
  }

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   * @dev STAKING_PAUSER_ROLE is required to execute.
   * @param _yieldProvider          Yield provider address.
   */
  function pauseStaking(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    if (_getYieldProviderStorage(_yieldProvider).isStakingPaused) {
      revert StakingAlreadyPaused();
    }
    _pauseStaking(_yieldProvider);
    emit YieldProviderStakingPaused(_yieldProvider, msg.sender);
  }
  
  function _pauseStaking(address _yieldProvider) internal {
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.pauseStaking, ()
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    _getYieldProviderStorage(_yieldProvider).isStakingPaused = true;
  }

  function _pauseStakingIfNotAlready(address _yieldProvider) internal {
    if (!_getYieldProviderStorage(_yieldProvider).isStakingPaused) {
      _pauseStaking(_yieldProvider);
    }
  }

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev STAKING_UNPAUSER_ROLE is required to execute.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   * @param _yieldProvider          Yield provider address.
   */
  function unpauseStaking(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    // Other checks for unstaking
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if (!$$.isStakingPaused) {
      revert StakingAlreadyUnpaused();
    }
    if (isWithdrawalReserveBelowMinimum()) {
      revert InsufficientWithdrawalReserve();
    }
    if ($$.lstLiabilityPrincipal > 0) {
      revert UnpauseStakingForbiddenWithCurrentLSTPrincipal();
    }
    _unpauseStaking(_yieldProvider);
    emit YieldProviderStakingUnpaused(_yieldProvider, msg.sender);
  }

  function _unpauseStaking(address _yieldProvider) internal {
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.pauseStaking, ()
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    _getYieldProviderStorage(_yieldProvider).isStakingPaused = false;
  }

  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external onlyKnownYieldProvider(_yieldProvider) {
    if (!ILineaNativeYieldExtension(l1MessageService()).isWithdrawLSTAllowed()) {
      revert LSTWithdrawalNotAllowed();
    }
    _pauseStakingIfNotAlready(_yieldProvider);
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.withdrawLST, (_amount, _recipient)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    emit LSTMinted(_yieldProvider, msg.sender, _recipient, _amount);
  }

  // TODO - Role
  // @dev Will permanently block LST minting, if there is no undo function.
  function initiateOssification(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssified) {
      revert AlreadyOssified();
    }
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.initiateOssification, ()
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    _pauseStakingIfNotAlready(_yieldProvider);
    $$.isOssificationInitiated = true;
    emit YieldProviderOssificationInitiated(_yieldProvider, msg.sender);
  }

  // TODO - Role
  function undoInitiateOssification(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if (!$$.isOssificationInitiated) {
      revert OssificationNotInitiated();
    }
    if ($$.isOssified) {
      revert AlreadyOssified();
    }
    if (_getYieldProviderStorage(_yieldProvider).isStakingPaused) {
      _unpauseStaking(_yieldProvider);
    }
    $$.isOssificationInitiated = false;
    emit YieldProviderOssificationReverted(_yieldProvider, msg.sender);
  }

  function processPendingOssification(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) returns (bool isOssificationComplete) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if (!$$.isOssificationInitiated) {
      revert OssificationNotInitiated();
    }
    if ($$.isOssified) {
      revert AlreadyOssified();
    }
    (bool success, bytes memory data) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.processPendingOssification, ()
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    (isOssificationComplete) = abi.decode(data, (bool));
    if (isOssificationComplete) {
      $$.isOssificationInitiated = true;
    }
    emit YieldProviderOssificationProcessed(_yieldProvider, msg.sender, isOssificationComplete);
  }

  // Need donate function here, otherwise YieldManager is unable to assign donations for specific yield providers.
  function donate(address _yieldProvider) external payable onlyKnownYieldProvider(_yieldProvider) {
    address l1MessageServiceCached = l1MessageService();
    _decrementNegativeYieldAgainstDonation(_yieldProvider, msg.value);
    _decrementPendingPermissionlessUnstake(msg.value);
    ILineaNativeYieldExtension(l1MessageServiceCached).fund{value: msg.value}();
    emit DonationProcessed(_yieldProvider, msg.sender, l1MessageServiceCached, msg.value);
  }

  // @dev It is not correct to count a donation to the L1MessageService as yield, because reported yield results in newly circulating L2 ETH.
  function _decrementNegativeYieldAgainstDonation(address _yieldProvider, uint256 _amount) internal {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    uint256 currentNegativeYield = $$.currentNegativeYield;
    if (currentNegativeYield > 0) {
      $$.currentNegativeYield -= Math256.min(currentNegativeYield, _amount);
    }
  }

  function addYieldProvider(address _yieldProvider, YieldProviderRegistration calldata _yieldProviderRegistration) external onlyRole(YIELD_PROVIDER_SETTER) {
    if (_yieldProvider == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    if (_yieldProviderRegistration.yieldProviderEntrypoint == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    IYieldProvider(_yieldProvider).validateAdditionToYieldManager(_yieldProviderRegistration);
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    
    if (_getYieldProviderStorage(_yieldProvider).yieldProviderIndex != 0) {
      revert YieldProviderAlreadyAdded();
    }
    // Ensure no added yield provider has index 0
    uint96 yieldProviderIndex = uint96($._yieldProviders.length) + 1;
    $._yieldProviders.push(_yieldProvider);
    $._yieldProviderStorage[_yieldProvider] = YieldProviderStorage({
        yieldProviderType: _yieldProviderRegistration.yieldProviderType,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        yieldProviderEntrypoint: _yieldProviderRegistration.yieldProviderEntrypoint,
        yieldProviderOssificationEntrypoint: _yieldProviderRegistration.yieldProviderOssificationEntrypoint,
        yieldProviderIndex: yieldProviderIndex,
        userFunds: 0,
        yieldReportedCumulative: 0,
        currentNegativeYield: 0,
        lstLiabilityPrincipal: 0
    });
    emit YieldProviderAdded(
      _yieldProvider,
      msg.sender,
      _yieldProviderRegistration.yieldProviderType,
      _yieldProviderRegistration.yieldProviderEntrypoint,
      _yieldProviderRegistration.yieldProviderOssificationEntrypoint
    );
  }

  function removeYieldProvider(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) onlyRole(YIELD_PROVIDER_SETTER)  {
    if (_yieldProvider == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    // We assume that 'pendingPermissionlessUnstake' and 'currentNegativeYield' must be 0, before 'userFunds' can be 0.
    if (_getYieldProviderStorage(_yieldProvider).userFunds != 0) {
      revert YieldProviderHasRemainingFunds();
    }
    _removeYieldProvider(_yieldProvider);
    emit YieldProviderRemoved(_yieldProvider, msg.sender, false);
  }

  // @dev Removes the requirement that there is 0 userFunds remaining in the YieldProvder
  // @dev Otherwise newly reported yield can prevent removeYieldProvider
  function emergencyRemoveYieldProvider(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) onlyRole(YIELD_PROVIDER_SETTER)  {
    if (_yieldProvider == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    _removeYieldProvider(_yieldProvider);
    emit YieldProviderRemoved(_yieldProvider, msg.sender, true);
  }

  function _removeYieldProvider(address _yieldProvider) internal  {
    YieldManagerStorage storage $ = _getYieldManagerStorage();

    uint96 yieldProviderIndex = _getYieldProviderStorage(_yieldProvider).yieldProviderIndex;
    address lastYieldProvider = $._yieldProviders[$._yieldProviders.length - 1];
    $._yieldProviderStorage[lastYieldProvider].yieldProviderIndex = yieldProviderIndex;
    $._yieldProviders[yieldProviderIndex] = lastYieldProvider;
    $._yieldProviders.pop();

    // TODO - Does this actually wipe the whole struct, to delete the storage pointer?
    delete $._yieldProviderStorage[_yieldProvider];
  }

  function setL1MessageService(address _l1MessageService) external {
    if (_l1MessageService == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    address oldL1MessageService = $._l1MessageService;
    emit L1MessageServiceUpdated(oldL1MessageService, _l1MessageService, msg.sender);
    $._l1MessageService = _l1MessageService;
  }

  function setL2YieldRecipient(address _newL2YieldRecipient) external {
    if (_newL2YieldRecipient == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    emit L2YieldRecipientSet($._l2YieldRecipient, _newL2YieldRecipient, msg.sender);
    $._l2YieldRecipient = _newL2YieldRecipient;
  }

  /**
   * @notice Set minimum withdrawal reserve percentage.
   * @dev Units of bps.
   * @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
   * @dev WITHDRAWAL_RESERVE_SETTER_ROLE is required to execute.
   * @param _minimumWithdrawalReservePercentageBps Minimum withdrawal reserve percentage in bps.
   */
  function setMinimumWithdrawalReservePercentageBps(uint16 _minimumWithdrawalReservePercentageBps) external onlyRole(WITHDRAWAL_RESERVE_SETTER_ROLE) {
      if (_minimumWithdrawalReservePercentageBps > MAX_BPS) {
        revert BpsMoreThan10000();
      }
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      if ($._targetWithdrawalReservePercentageBps < _minimumWithdrawalReservePercentageBps) {
        revert TargetReservePercentageMustBeAboveMinimum();
      }
      emit MinimumWithdrawalReservePercentageBpsSet($._minimumWithdrawalReservePercentageBps, _minimumWithdrawalReservePercentageBps, msg.sender);
      $._minimumWithdrawalReservePercentageBps = _minimumWithdrawalReservePercentageBps;
  }

  /**
   * @notice Set minimum withdrawal reserve.
   * @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
   * @dev WITHDRAWAL_RESERVE_SETTER_ROLE is required to execute.
   * @param _minimumWithdrawalReserveAmount Minimum withdrawal reserve amount.
   */
  function setMinimumWithdrawalReserveAmount(uint256 _minimumWithdrawalReserveAmount) external onlyRole(WITHDRAWAL_RESERVE_SETTER_ROLE) {
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      if ($._targetWithdrawalReserveAmount < _minimumWithdrawalReserveAmount) {
        revert TargetReserveAmountMustBeAboveMinimum();
      }
      emit MinimumWithdrawalReserveAmountSet($._minimumWithdrawalReserveAmount, _minimumWithdrawalReserveAmount, msg.sender);
      $._minimumWithdrawalReserveAmount = _minimumWithdrawalReserveAmount;
  }

  function setTargetWithdrawalReservePercentageBps(uint16 _targetWithdrawalReservePercentageBps) external onlyRole(WITHDRAWAL_RESERVE_SETTER_ROLE) {
      if (_targetWithdrawalReservePercentageBps > MAX_BPS) {
        revert BpsMoreThan10000();
      }
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      if (_targetWithdrawalReservePercentageBps < $._minimumWithdrawalReservePercentageBps) {
        revert TargetReservePercentageMustBeAboveMinimum();
      }
      uint16 oldTargetWithdrawalReservePercentageBps = $._targetWithdrawalReservePercentageBps;
      emit TargetWithdrawalReservePercentageBpsSet(
        oldTargetWithdrawalReservePercentageBps,
        _targetWithdrawalReservePercentageBps,
        msg.sender
      );
      $._targetWithdrawalReservePercentageBps = _targetWithdrawalReservePercentageBps;
  }

  function setTargetWithdrawalReserveAmount(uint256 _targetWithdrawalReserveAmount) external onlyRole(WITHDRAWAL_RESERVE_SETTER_ROLE) {
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      if (_targetWithdrawalReserveAmount < $._minimumWithdrawalReserveAmount) {
        revert TargetReserveAmountMustBeAboveMinimum();
      }
      uint256 oldTargetWithdrawalReserveAmount = $._targetWithdrawalReserveAmount;
      emit TargetWithdrawalReserveAmountSet(
        oldTargetWithdrawalReserveAmount,
        _targetWithdrawalReserveAmount,
        msg.sender
      );
      $._targetWithdrawalReserveAmount = _targetWithdrawalReserveAmount;
  }
}
