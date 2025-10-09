// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { YieldManager } from "../../../yield/YieldManager.sol";
import { MockYieldProviderStorageLayout } from "../../mocks/yield/MockYieldProviderStorageLayout.sol";
import { TestLidoStVaultYieldProvider } from "./TestLidoStVaultYieldProvider.sol";
import { IYieldProvider } from "../../../yield/interfaces/IYieldProvider.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestYieldManager is YieldManager, MockYieldProviderStorageLayout {
  constructor(address _l1MessageService) YieldManager(_l1MessageService) {}

  function setTransientReceiveCaller(address _caller) external {
    TRANSIENT_RECEIVE_CALLER = _caller;
  }

  function getTransientReceiveCaller() external view returns (address) {
    return TRANSIENT_RECEIVE_CALLER;
  }

  function getL1MessageService() external view returns (address) {
    return L1_MESSAGE_SERVICE;
  }

  function getMinimumWithdrawalReservePercentageBps() external view returns (uint16) {
    return _getYieldManagerStorage()._minimumWithdrawalReservePercentageBps;
  }

  function setMinimumWithdrawalReservePercentageBps(uint16 _minimumWithdrawalReservePercentageBps) external {
    _getYieldManagerStorage()._minimumWithdrawalReservePercentageBps = _minimumWithdrawalReservePercentageBps;
  }

  function getTargetWithdrawalReservePercentageBps() external view returns (uint16) {
    return _getYieldManagerStorage()._targetWithdrawalReservePercentageBps;
  }

  function setTargetWithdrawalReservePercentageBps(uint16 _targetWithdrawalReservePercentageBps) external {
    _getYieldManagerStorage()._targetWithdrawalReservePercentageBps = _targetWithdrawalReservePercentageBps;
  }

  function getMinimumWithdrawalReserveAmount() external view returns (uint256) {
    return _getYieldManagerStorage()._minimumWithdrawalReserveAmount;
  }

  function setMinimumWithdrawalReserveAmount(uint256 _minimumWithdrawalReserveAmount) external {
    _getYieldManagerStorage()._minimumWithdrawalReserveAmount = _minimumWithdrawalReserveAmount;
  }

  function getTargetWithdrawalReserveAmount() external view returns (uint256) {
    return _getYieldManagerStorage()._targetWithdrawalReserveAmount;
  }

  function setTargetWithdrawalReserveAmount(uint256 _targetWithdrawalReserveAmount) external {
    _getYieldManagerStorage()._targetWithdrawalReserveAmount = _targetWithdrawalReserveAmount;
  }

  function getUserFundsInYieldProvidersTotal() external view returns (uint256) {
    return _getYieldManagerStorage()._userFundsInYieldProvidersTotal;
  }

  function setUserFundsInYieldProvidersTotal(uint256 _userFundsInYieldProvidersTotal) external {
    _getYieldManagerStorage()._userFundsInYieldProvidersTotal = _userFundsInYieldProvidersTotal;
  }

  function getPendingPermissionlessUnstake() external view returns (uint256) {
    return _getYieldManagerStorage()._pendingPermissionlessUnstake;
  }

  function setPendingPermissionlessUnstake(uint256 _pendingPermissionlessUnstake) external {
    _getYieldManagerStorage()._pendingPermissionlessUnstake = _pendingPermissionlessUnstake;
  }

  function decrementPendingPermissionlessUnstake(uint256 _amount) external {
    _decrementPendingPermissionlessUnstake(_amount);
  }

  function getYieldProviders() external view returns (address[] memory) {
    return _getYieldManagerStorage()._yieldProviders;
  }

  function getIsL2YieldRecipientKnown(address _l2YieldRecipient) external view returns (bool) {
    return _getYieldManagerStorage()._isL2YieldRecipientKnown[_l2YieldRecipient];
  }

  function setIsL2YieldRecipientKnown(address _l2YieldRecipient, bool _isKnown) external {
    _getYieldManagerStorage()._isL2YieldRecipientKnown[_l2YieldRecipient] = _isKnown;
  }

  function getYieldProviderVendor(address _yieldProvider) external view returns (YieldProviderVendor) {
    return _getYieldProviderStorage(_yieldProvider).yieldProviderVendor;
  }

  function setYieldProviderVendor(address _yieldProvider, YieldProviderVendor _yieldProviderVendor) external {
    _getYieldProviderStorage(_yieldProvider).yieldProviderVendor = _yieldProviderVendor;
  }

  function getYieldProviderIsStakingPaused(address _yieldProvider) external view returns (bool) {
    return _getYieldProviderStorage(_yieldProvider).isStakingPaused;
  }

  function setYieldProviderIsStakingPaused(address _yieldProvider, bool _isPaused) external {
    _getYieldProviderStorage(_yieldProvider).isStakingPaused = _isPaused;
  }

  function getYieldProviderIsOssificationInitiated(address _yieldProvider) external view returns (bool) {
    return _getYieldProviderStorage(_yieldProvider).isOssificationInitiated;
  }

  function setYieldProviderIsOssificationInitiated(address _yieldProvider, bool _isInitiated) external {
    _getYieldProviderStorage(_yieldProvider).isOssificationInitiated = _isInitiated;
  }

  function getYieldProviderIsOssified(address _yieldProvider) external view returns (bool) {
    return _getYieldProviderStorage(_yieldProvider).isOssified;
  }

  function setYieldProviderIsOssified(address _yieldProvider, bool _isOssified) external {
    _getYieldProviderStorage(_yieldProvider).isOssified = _isOssified;
  }

  function getYieldProviderPrimaryEntrypoint(address _yieldProvider) external view returns (address) {
    return _getYieldProviderStorage(_yieldProvider).primaryEntrypoint;
  }

  function setYieldProviderPrimaryEntrypoint(address _yieldProvider, address _primaryEntrypoint) external {
    _getYieldProviderStorage(_yieldProvider).primaryEntrypoint = _primaryEntrypoint;
  }

  function getYieldProviderOssifiedEntrypoint(address _yieldProvider) external view returns (address) {
    return _getYieldProviderStorage(_yieldProvider).ossifiedEntrypoint;
  }

  function setYieldProviderOssifiedEntrypoint(address _yieldProvider, address _ossifiedEntrypoint) external {
    _getYieldProviderStorage(_yieldProvider).ossifiedEntrypoint = _ossifiedEntrypoint;
  }

  function getYieldProviderReceiveCaller(address _yieldProvider) external view returns (address) {
    return _getYieldProviderStorage(_yieldProvider).receiveCaller;
  }

  function setYieldProviderReceiveCaller(address _yieldProvider, address _receiveCaller) external {
    _getYieldProviderStorage(_yieldProvider).receiveCaller = _receiveCaller;
  }

  function getYieldProviderIndex(address _yieldProvider) external view returns (uint96) {
    return _getYieldProviderStorage(_yieldProvider).yieldProviderIndex;
  }

  function setYieldProviderIndex(address _yieldProvider, uint96 _yieldProviderIndex) external {
    _getYieldProviderStorage(_yieldProvider).yieldProviderIndex = _yieldProviderIndex;
  }

  function getYieldProviderUserFunds(address _yieldProvider) external view returns (uint256) {
    return _getYieldProviderStorage(_yieldProvider).userFunds;
  }

  function setYieldProviderUserFunds(address _yieldProvider, uint256 _userFunds) external {
    _getYieldProviderStorage(_yieldProvider).userFunds = _userFunds;
  }

  function getYieldProviderYieldReportedCumulative(address _yieldProvider) external view returns (uint256) {
    return _getYieldProviderStorage(_yieldProvider).yieldReportedCumulative;
  }

  function setYieldProviderYieldReportedCumulative(address _yieldProvider, uint256 _yieldReportedCumulative) external {
    _getYieldProviderStorage(_yieldProvider).yieldReportedCumulative = _yieldReportedCumulative;
  }

  function getYieldProviderCurrentNegativeYield(address _yieldProvider) external view returns (uint256) {
    return _getYieldProviderStorage(_yieldProvider).currentNegativeYield;
  }

  function setYieldProviderCurrentNegativeYield(address _yieldProvider, uint256 _currentNegativeYield) external {
    _getYieldProviderStorage(_yieldProvider).currentNegativeYield = _currentNegativeYield;
  }

  function getYieldProviderLstLiabilityPrincipal(address _yieldProvider) external view returns (uint256) {
    return _getYieldProviderStorage(_yieldProvider).lstLiabilityPrincipal;
  }

  function setYieldProviderLstLiabilityPrincipal(address _yieldProvider, uint256 _lstLiabilityPrincipal) external {
    _getYieldProviderStorage(_yieldProvider).lstLiabilityPrincipal = _lstLiabilityPrincipal;
  }

  function delegatecallWithdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external {
    _delegatecallWithdrawFromYieldProvider(_yieldProvider, _amount);
  }

  function getEntrypointContract(address _yieldProvider) external returns (address) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.getEntrypointContract, (_yieldProvider))
    );
    return abi.decode(data, (address));
  }

  function getDashboard(address _yieldProvider) external returns (address) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.getDashboard, (_yieldProvider))
    );
    return abi.decode(data, (address));
  }

  function getVault(address _yieldProvider) external returns (address) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.getVault, (_yieldProvider))
    );
    return abi.decode(data, (address));
  }

  function handlePostiveYieldAccounting(address _yieldProvider, uint256 _availableYield) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.handlePostiveYieldAccounting, (_yieldProvider, _availableYield))
    );
    return abi.decode(data, (uint256));
  }

  function syncExternalLiabilitySettlement(
    address _yieldProvider,
    uint256 _liabilityShares,
    uint256 _lstLiabilityPrincipalCached
  ) external returns (uint256, bool) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(
        TestLidoStVaultYieldProvider.syncExternalLiabilitySettlement,
        (_yieldProvider, _liabilityShares, _lstLiabilityPrincipalCached)
      )
    );
    return abi.decode(data, (uint256, bool));
  }

  function payObligations(address _yieldProvider, uint256 _availableYield) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.payObligations, (_yieldProvider, _availableYield))
    );
    return abi.decode(data, (uint256));
  }

  function payNodeOperatorFees(address _yieldProvider, uint256 _availableYield) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.payNodeOperatorFees, (_yieldProvider, _availableYield))
    );
    return abi.decode(data, (uint256));
  }

  function payLSTPrincipalExternal(address _yieldProvider, uint256 _availableFunds) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(IYieldProvider.payLSTPrincipal, (_yieldProvider, _availableFunds))
    );
    return abi.decode(data, (uint256));
  }

  function payLSTPrincipalInternal(address _yieldProvider, uint256 _availableFunds) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.payLSTPrincipalInternal, (_yieldProvider, _availableFunds))
    );
    return abi.decode(data, (uint256));
  }

  function unstakeHarness(
    address _yieldProvider,
    bytes memory _pubkeys,
    uint64[] memory _amounts,
    address _refundRecipient
  ) external payable {
    _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(
        TestLidoStVaultYieldProvider.unstakeHarness,
        (_yieldProvider, _pubkeys, _amounts, _refundRecipient)
      )
    );
  }

  function validateUnstakePermissionlessHarness(
    address _yieldProvider,
    bytes memory _pubkeys,
    uint64[] memory _amounts,
    bytes memory _withdrawalParamsProof
  ) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(
        TestLidoStVaultYieldProvider.validateUnstakePermissionlessHarness,
        (_yieldProvider, _pubkeys, _amounts, _withdrawalParamsProof)
      )
    );
    return abi.decode(data, (uint256));
  }

  function payMaximumPossibleLSTLiability(address _yieldProvider) external returns (uint256) {
    bytes memory data = _delegatecallYieldProvider(
      _yieldProvider,
      abi.encodeCall(TestLidoStVaultYieldProvider.payMaximumPossibleLSTLiability, (_yieldProvider))
    );
    return abi.decode(data, (uint256));
  }

  function withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
    address _yieldProvider,
    uint256 _amount,
    uint256 _targetDeficit
  ) external returns (uint256 withdrawAmount, uint256 lstPrincipalPaid) {
    return _withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(_yieldProvider, _amount, _targetDeficit);
  }

  function pauseStakingIfNotAlready(address _yieldProvider) external {
    _pauseStakingIfNotAlready(_yieldProvider);
  }
}
