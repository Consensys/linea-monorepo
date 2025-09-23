// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "./YieldManagerStorageLayout.sol";
import { IYieldManager } from "./interfaces/IYieldManager.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { ILineaNativeYieldExtension } from "./interfaces/ILineaNativeYieldExtension.sol";
import { YieldManagerPauseManager } from "../security/pausing/YieldManagerPauseManager.sol";

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
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      return $._l1MessageService;
  }

  /// @notice Minimum withdrawal reserve percentage in bps.
  /// @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
  function minimumWithdrawalReservePercentageBps() public view returns (uint256) {
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      return $._minimumWithdrawalReservePercentageBps;
  }

  /// @notice Minimum withdrawal reserve amount.
  /// @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
  function minimumWithdrawalReserveAmount() public view returns (uint256) {
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      return $._minimumWithdrawalReserveAmount;
  }

  /**
   * @notice Initialises the YieldManager.
   */
  function initialize() external initializer {
  }

  modifier onlyKnownYieldProvider(address _yieldProvider) {
      YieldManagerStorage storage $ = _getYieldManagerStorage();
      if ($._yieldProviderData[_yieldProvider].yieldProviderIndex != 0) {
        revert UnknownYieldProvider();
      }
      _;
  }

 function getWithdrawalReserveBalance() public view returns (uint256) {
    return _getYieldManagerStorage()._l1MessageService.balance;
  }

  function getTotalSystemBalance() public view returns (uint256) {
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    return $._l1MessageService.balance + address(this).balance + $._userFundsInYieldProvidersTotal;
  }

  function getMinimumWithdrawalReserveByPercentage() public view returns (uint256) {
    return getTotalSystemBalance() * _getYieldManagerStorage()._minimumWithdrawalReservePercentageBps / MAX_BPS;
  }

  /// @notice Get effective minimum withdrawal reserve
  /// @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
  function getEffectiveMinimumWithdrawalReserve() public view returns (uint256) {
      uint256 minimumWithdrawalReserveAmountCached = _getYieldManagerStorage()._minimumWithdrawalReserveAmount;
      uint256 minWithdrawalReserveByPercentage = getMinimumWithdrawalReserveByPercentage();
      return minWithdrawalReserveByPercentage > minimumWithdrawalReserveAmountCached ? minWithdrawalReserveByPercentage : minimumWithdrawalReserveAmountCached;
  }

  /// @notice Returns true if withdrawal reserve balance is below effective required minimum.
  /// @dev We are doing duplicate BALANCE opcode call, but how to remove duplicate call while maintaining readability?
  function isWithdrawalReserveBelowEffectiveMinimum() public view returns (bool) {
      return _getYieldManagerStorage()._l1MessageService.balance < getEffectiveMinimumWithdrawalReserve();
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev YIELD_PROVIDER_FUNDER_ROLE is required to execute.
   * @dev Reverts if the withdrawal reserve is below the minimum threshold.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _amount        The amount of ETH to send.
   * @param _yieldProvider The target yield provider contract.
   */
  function fundYieldProvider(uint256 _amount, address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    if (!isWithdrawalReserveBelowEffectiveMinimum()) {
        revert InsufficientWithdrawalReserve();
    }
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.fundYieldProvider, (_yieldProvider, _amount)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    _getYieldManagerStorage()._userFundsInYieldProvidersTotal += _amount;
    _getYieldProviderDataStorage(_yieldProvider).userFunds += _amount;
    // emit event?
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
    if (!isWithdrawalReserveBelowEffectiveMinimum()) {
        revert InsufficientWithdrawalReserve();
    }
    // TODO - Emit event
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
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev YIELD_REPORTER_ROLE is required to execute.
   * @dev Since the YieldManager is unaware of donations received via the L1MessageService or L2MessageService,
   *      the `_reserveDonations` parameter is required to ensure accurate yield accounting.
   * @param _totalReserveDonations   Total amount of donations received on the L1MessageService or L2MessageService.
   * @param _yieldProvider      Yield provider address.
   */
  function reportYield(uint256 _totalReserveDonations, address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {

  }

  /**
   * @notice Request beacon chain withdrawal from specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE or RESERVE_OPERATOR_ROLE is required to execute.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   * @param _yieldProvider      Yield provider address.
   */
  function unstake(bytes memory _withdrawalParams, address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {

  }

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
   * @param _withdrawalParams       Provider-specific withdrawal parameters.
   * @param _withdrawalParamsProof  Merkle proof of _withdrawalParams to be verified against EIP-4788 beacon chain root.
   * @param _yieldProvider          Yield provider address.
   */
  function unstakePermissionless(
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof,
    address _yieldProvider
  ) external onlyKnownYieldProvider(_yieldProvider) {

  }

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If fund remaining, will settle any outstanding LST liabilities and protocol obligations.
   * @param _amount                 Amount to withdraw.
   * @param _yieldProvider          Yield provider address.
   */
  function withdrawFromYieldProvider(uint256 _amount, address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {

  }

  /**
   * @notice Rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Settles any outstanding LST liabilities, provided this does not leave the withdrawal reserve in deficit.
   * @param _amount                 Amount to withdraw.
   * @param _yieldProvider          Yield provider address.
   */
  function addToWithdrawalReserve(uint256 _amount, address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {

  }

  /**
   * @notice Permissionlessly rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev Only available when the withdrawal is in deficit.
   * @param _amount                 Amount to withdraw.
   * @param _yieldProvider          Yield provider address.
   */
  function replenishWithdrawalReserve(uint256 _amount, address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {

  }

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   * @dev STAKING_PAUSER_ROLE is required to execute.
   * @param _yieldProvider          Yield provider address.
   */
  function pauseStaking(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    YieldProviderData storage $$ = _getYieldProviderDataStorage(_yieldProvider);
    if ($$.isStakingPaused) {
      revert StakingAlreadyPaused();
    }
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.pauseStaking, (_yieldProvider)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    $$.isStakingPaused = true;
    // Emit event
  }

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev STAKING_UNPAUSER_ROLE is required to execute.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   * @param _yieldProvider          Yield provider address.
   */
  function unpauseStaking(address _yieldProvider) external onlyKnownYieldProvider(_yieldProvider) {
    // Other checks for unstaking
    YieldProviderData storage $$ = _getYieldProviderDataStorage(_yieldProvider);
    if (!$$.isStakingPaused) {
      revert StakingAlreadyUnpaused();
    }
    if (!isWithdrawalReserveBelowEffectiveMinimum()) {
        revert InsufficientWithdrawalReserve();
    }
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.pauseStaking, (_yieldProvider)
    ));
    if (!success) {
      revert DelegateCallFailed();
    }
    $$.isStakingPaused = false;
    // emit Event
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
    
    if ($._yieldProviderData[_yieldProvider].yieldProviderIndex != 0) {
      revert YieldProviderAlreadyAdded();
    }
    // Ensure no added yield provider has index 0
    uint96 yieldProviderIndex = uint96($._yieldProviders.length) + 1;
    // TODO - ? Need to check for uint96 overflow
    $._yieldProviders.push(_yieldProvider);
    // $._yieldProviderRegistration[_yieldProvider] = _yieldProviderRegistration;
    $._yieldProviderData[_yieldProvider] = YieldProviderData({
        registration: _yieldProviderRegistration,
        yieldProviderIndex: yieldProviderIndex,
        isStakingPaused: false,
        isOssified: false,
        userFunds: 0,
        yieldReportedCumulative: 0
    });
    // TODO - Emit event
  }

  function removeYieldProvider(address _yieldProvider) external onlyRole(YIELD_PROVIDER_SETTER)  {
    if (_yieldProvider == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    YieldManagerStorage storage $ = _getYieldManagerStorage();
    // TODO - How to handle remaining yield situation? If we keep earning yield, we can never exit?...Then don't report
    if ($._yieldProviderData[_yieldProvider].userFunds != 0) {
      revert YieldProviderHasRemainingFunds();
    }
    _removeYieldProvider(_yieldProvider);
    // TODO - Emit event
  }

  // @dev Removes the requirement that there is 0 userFunds remaining in the YieldProvder
  // @dev Otherwise newly reported yield can prevent removeYieldProvider
  function emergencyRemoveYieldProvider(address _yieldProvider) external onlyRole(YIELD_PROVIDER_SETTER)  {
    if (_yieldProvider == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    YieldManagerStorage storage $ = _getYieldManagerStorage();
    _removeYieldProvider(_yieldProvider);
    // TODO - Emit event
  }


  function _removeYieldProvider(address _yieldProvider) internal  {
    YieldManagerStorage storage $ = _getYieldManagerStorage();
    uint96 yieldProviderIndex = $._yieldProviderData[_yieldProvider].yieldProviderIndex;
    address lastYieldProvider = $._yieldProviders[$._yieldProviders.length - 1];
    $._yieldProviderData[lastYieldProvider].yieldProviderIndex = yieldProviderIndex;
    $._yieldProviders[yieldProviderIndex] = lastYieldProvider;
    $._yieldProviders.pop();

    delete $._yieldProviderData[_yieldProvider];
  }

  // TODO - Setter for l1MessageService

  /**
   * @notice Set minimum withdrawal reserve percentage.
   * @dev Units of bps.
   * @dev Effective minimum reserve is min(minimumWithdrawalReservePercentageBps, minimumWithdrawalReserveAmount).
   * @dev WITHDRAWAL_RESERVE_SETTER_ROLE is required to execute.
   * @param _minimumWithdrawalReservePercentageBps Minimum withdrawal reserve percentage in bps.
   */
  function setMinimumWithdrawalReservePercentageBps(uint256 _minimumWithdrawalReservePercentageBps) external onlyRole(WITHDRAWAL_RESERVE_SETTER_ROLE) {
      if (_minimumWithdrawalReservePercentageBps > MAX_BPS) {
        revert BpsMoreThan10000();
      }
      YieldManagerStorage storage $ = _getYieldManagerStorage();
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
      emit MinimumWithdrawalReserveAmountSet($._minimumWithdrawalReserveAmount, _minimumWithdrawalReserveAmount, msg.sender);
      $._minimumWithdrawalReserveAmount = _minimumWithdrawalReserveAmount;
  }
}