// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "./YieldManagerStorageLayout.sol";
import { IYieldManager } from "./interfaces/IYieldManager.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { ICommonVaultOperations } from "./interfaces/vendor/lido-vault/ICommonVaultOperations.sol";
import { IDashboard } from "./interfaces/vendor/lido-vault/IDashboard.sol";
import { IStETH } from "./interfaces/vendor/lido-vault/IStETH.sol";
import { IVaultHub } from "./interfaces/vendor/lido-vault/IVaultHub.sol";
import { IStakingVault } from "./interfaces/vendor/lido-vault/IStakingVault.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldManagerStorageLayout, IYieldProvider, IGenericErrors {
  // yieldProvider = StakingVault
  address immutable yieldProvider;
  IVaultHub immutable vaultHub;
  IDashboard immutable dashboard;

  // @dev _yieldProvider = stakingVault address
  constructor (address _yieldProvider, address _vaultHub, address _dashboard) {
    // Do checks
    yieldProvider = _yieldProvider;
    vaultHub = IVaultHub(_vaultHub);
    dashboard = IDashboard(_dashboard);
  }

  function _getEntrypointContract() private view returns (address) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(yieldProvider);
    return $$.isOssified ? yieldProvider : address(dashboard);
  }

  function _getStakingVault() private view returns (address) {
    return yieldProvider;
  }

  function _getCurrentLSTLiabilityETH() internal view returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(yieldProvider);
    if ($$.isOssified) return 0;
    uint256 liabilityShares = dashboard.liabilityShares();
    // Use `roundUp` variant to be conservative
    uint256 liabilityEth = IStETH(dashboard.STETH()).getPooledEthBySharesRoundUp(liabilityShares);
    return liabilityEth;
  }

  // Will settle as much LST liability as possible. Will return amount of liabilityEth remaining
  function _payMaximumPossibleLSTLiability() internal returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(yieldProvider);
    if ($$.isOssified) return 0;

    uint256 vaultBalanceEth = yieldProvider.balance;
    IStETH steth = IStETH(dashboard.STETH());
    uint256 vaultBalanceShares = steth.getSharesByPooledEth(vaultBalanceEth);
    uint256 liabilityShares = dashboard.liabilityShares();
    uint256 liabilitySharesPaid = liabilityShares < vaultBalanceShares ? liabilityShares : vaultBalanceShares;
    // Do the payment
    dashboard.rebalanceVaultWithShares(liabilitySharesPaid);
  }

  function _payMaximumPossibleLSTInterest() internal {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(yieldProvider);
    if ($$.isOssified) return;
  }

  // @dev Omit redemptions, because it will overlap with liabilities
  function _getCurrentObligationsMinusRedemptions() internal view returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(yieldProvider);
    // yieldProviderOssificationEntrypoint = StakingVault
    IVaultHub.VaultObligations memory obligations = vaultHub.vaultObligations($$.registration.yieldProviderOssificationEntrypoint);
    return obligations.unsettledLidoFees + dashboard.nodeOperatorDisbursableFee();
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(uint256 _amount) external {
    ICommonVaultOperations(_getEntrypointContract()).fund{value: _amount}();
  }

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev Since the YieldManager is unaware of donations received via the L1MessageService or L2MessageService,
   *      the `_reserveDonations` parameter is required to ensure accurate yield accounting.
   * @param _totalReserveDonations   Total amount of donations received on the L1MessageService or L2MessageService.
   */
  function reportYield(uint256 _totalReserveDonations) external returns (uint256 _newYield) {
    if (_getYieldProviderDataStorage(yieldProvider).isOssified) {
      revert OperationNotSupportedDuringOssification(OperationType.ReportYield);
    }
    _payMaximumPossibleLSTLiability();
    uint256 liabilityEth = _getCurrentLSTLiabilityETH();
    uint256 obligationsEth = _getCurrentObligationsMinusRedemptions();
    // We could cache CONNECT_DEPOSIT = 1 ether, but what if VaulHub contract upgrades and this value changes?
    uint256 totalVaultEth = dashboard.totalValue() - vaultHub.CONNECT_DEPOSIT();
    uint256 fundsAvailableForUserWithdrawal = totalVaultEth - obligationsEth - liabilityEth;
    return fundsAvailableForUserWithdrawal - _getYieldProviderDataStorage(yieldProvider).userFunds;
  }

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(bytes memory _withdrawalParams) external {
    _unstake(_withdrawalParams);
  }

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
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external returns (uint256) {
    // TODO - Verify _withdrawalParamsProof
    uint256 amountUnstaked = _unstake(_withdrawalParams);
    return amountUnstaked;
  }

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function _unstake(bytes memory _withdrawalParams) internal returns (uint256) {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(_withdrawalParams, (bytes, uint64[], address));
    // Lido StakingVault.sol will handle the param validation
    ICommonVaultOperations(_getEntrypointContract()).triggerValidatorWithdrawals(pubkeys, amounts, refundRecipient);
    uint256 amountUnstaked;
    for (uint256 i = 0; i < amounts.length; i++) {
      amountUnstaked += amounts[i];
    }
    return amountUnstaked;
  }

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If fund remaining, will settle any outstanding LST liabilities.
   * @dev This function will first attempt to pay LST liabilities. However it will reserve '_targetReserveDeficit' out of this. So '_targetReserveDeficit' withdrawal is guaranteed.
   * @param _amount                 Amount to withdraw.
   */
  function withdrawWithReserveDeficitPriorityAndLSTLiabilityPrincipalReduction(uint256 _amount, address _recipient, uint256 _targetReserveDeficit) external returns (uint256) {
    uint256 withdrawAmount = _amount;
    if (_targetReserveDeficit >= withdrawAmount) {
      ICommonVaultOperations(_getEntrypointContract()).withdraw(_recipient, withdrawAmount);
      return withdrawAmount;
    } else {
      uint256 amountAvailableToPayLSTLiability = withdrawAmount - _targetReserveDeficit;
      uint256 currentLSTLiabilityETH = _getCurrentLSTLiabilityETH();
      if (currentLSTLiabilityETH > 0) {
        uint256 LSTLiabilityETHPaid = currentLSTLiabilityETH > amountAvailableToPayLSTLiability ? amountAvailableToPayLSTLiability : currentLSTLiabilityETH;
        dashboard.rebalanceVaultWithEther(LSTLiabilityETHPaid);
        withdrawAmount -= LSTLiabilityETHPaid;
      }
      ICommonVaultOperations(_getEntrypointContract()).withdraw(_recipient, withdrawAmount);
      return withdrawAmount;
    }
  }

  function withdrawFromYieldProvider(uint256 _amount, address _recipient) external {
    ICommonVaultOperations(_getEntrypointContract()).withdraw(_recipient, _amount);
  }

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   */
  function pauseStaking() external {
    ICommonVaultOperations(_getEntrypointContract()).pauseBeaconChainDeposits();
  }

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   */
  function unpauseStaking() external {
    ICommonVaultOperations(_getEntrypointContract()).resumeBeaconChainDeposits();
  }

  function validateAdditionToYieldManager(IYieldManager.YieldProviderRegistration calldata _yieldProviderRegistration) external pure {
    if (_yieldProviderRegistration.yieldProviderType != IYieldManager.YieldProviderType.LIDO_STVAULT) {
      revert IncorrectYieldProviderType();
    }
  }

  function getAvailableBalanceForWithdraw() external view returns (uint256) {
    return _getStakingVault().balance;
  }

  function mintLST(uint256 _amount, address _recipient) external {
    dashboard.mintStETH(_recipient, _amount);
  }

  // TODO - Role
  // @dev Requires fresh report
  function initiateOssification() external {
    _payMaximumPossibleLSTLiability();
    
    // Lido implementation handles Lido fee payment, and revert on fresh report
    // This will fail if any existing liabilities or obligations
    dashboard.voluntaryDisconnect();
  }

  // TODO - Role
  // Returns true if ossified after function call is done
  function processPendingOssification() external returns (bool) {
    // Give ownership to YieldManager
    dashboard.abandonDashboard(address(this));
    IStakingVault vault = IStakingVault(yieldProvider);
    vault.acceptOwnership();
    vault.ossify();
    return true;
  }
}