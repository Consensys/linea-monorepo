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



/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldManagerStorageLayout, IYieldProvider, IGenericErrors {
  function _getEntrypointContract(address _yieldProvider) private view returns (address) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(_yieldProvider);
    return $$.isOssified ? $$.registration.yieldProviderOssificationEntrypoint : $$.registration.yieldProviderEntrypoint;
  }

  function _getStakingVault(address _yieldProvider) private view returns (address) {
    return _getYieldProviderDataStorage(_yieldProvider).registration.yieldProviderOssificationEntrypoint;
  }

  function _getDashboard(address _yieldProvider) private view returns (address) {
    return _getYieldProviderDataStorage(_yieldProvider).registration.yieldProviderEntrypoint;
  }

  function _getCurrentLSTLiabilityETH(address _yieldProvider) internal view returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(_yieldProvider);
    if ($$.isOssified) return 0;
    IDashboard dashboard = IDashboard($$.registration.yieldProviderEntrypoint);
    uint256 liabilityShares = dashboard.liabilityShares();
    // Use `roundUp` variant to be conservative
    uint256 liabilityEth = IStETH(dashboard.STETH()).getPooledEthBySharesRoundUp(liabilityShares);
    return liabilityEth;
  }

  // Will settle as much LST liability as possible. Will return amount of liabilityEth remaining
  function _payMaximumPossibleLSTLiability(address _yieldProvider) internal returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(_yieldProvider);
    if ($$.isOssified) return 0;

    uint256 vaultBalanceEth = _getStakingVault(_yieldProvider).balance;
    IDashboard dashboard = IDashboard(_getDashboard(_yieldProvider));
    IStETH steth = IStETH(dashboard.STETH());
    uint256 vaultBalanceShares = steth.getSharesByPooledEth(vaultBalanceEth);
    uint256 liabilityShares = dashboard.liabilityShares();
    uint256 liabilitySharesPaid = liabilityShares < vaultBalanceShares ? liabilityShares : vaultBalanceShares;
    // Do the payment
    dashboard.rebalanceVaultWithShares(liabilitySharesPaid);
  }


  // @dev Omit redemptions, because it will overlap with liabilities
  function _getCurrentObligationsMinusRedemptions(address _yieldProvider) internal view returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(_yieldProvider);
    IDashboard dashboard = IDashboard($$.registration.yieldProviderEntrypoint);
    IVaultHub vaultHub = IVaultHub(dashboard.VAULT_HUB());
    // yieldProviderOssificationEntrypoint = StakingVault
    IVaultHub.VaultObligations memory obligations = vaultHub.vaultObligations($$.registration.yieldProviderOssificationEntrypoint);
    return obligations.unsettledLidoFees + dashboard.nodeOperatorDisbursableFee();
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).fund{value: _amount}();
  }

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev Since the YieldManager is unaware of donations received via the L1MessageService or L2MessageService,
   *      the `_reserveDonations` parameter is required to ensure accurate yield accounting.
   * @param _totalReserveDonations   Total amount of donations received on the L1MessageService or L2MessageService.
   */
  function reportYield(address _yieldProvider, uint256 _totalReserveDonations) external returns (uint256 _newYield) {
    if (_getYieldProviderDataStorage(_yieldProvider).isOssified) {
      revert OperationNotSupportedDuringOssification(OperationType.ReportYield);
    }
    _payMaximumPossibleLSTLiability(_yieldProvider);
    uint256 liabilityEth = _getCurrentLSTLiabilityETH(_yieldProvider);
    uint256 obligationsEth = _getCurrentObligationsMinusRedemptions(_yieldProvider);
    uint256 totalVaultEth = IDashboard(_getDashboard(_yieldProvider)).totalValue();
    uint256 fundsAvailableForUserWithdrawal = totalVaultEth - obligationsEth - liabilityEth;
    return fundsAvailableForUserWithdrawal - _getYieldProviderDataStorage(_yieldProvider).userFunds;
  }

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external {
    _unstake(_yieldProvider, _withdrawalParams);
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
    address _yieldProvider,
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof
  ) external returns (uint256) {
    // TODO - Verify _withdrawalParamsProof
    uint256 amountUnstaked = _unstake(_yieldProvider, _withdrawalParams);
    return amountUnstaked;
  }

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function _unstake(address _yieldProvider, bytes memory _withdrawalParams) internal returns (uint256) {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(_withdrawalParams, (bytes, uint64[], address));
    // Lido StakingVault.sol will handle the param validation
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).triggerValidatorWithdrawals(pubkeys, amounts, refundRecipient);
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
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount, uint256 _targetReserveDeficit, address _recipient) external returns (uint256) {
    uint256 withdrawAmount = _amount;
    if (_targetReserveDeficit >= withdrawAmount) {
      ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).withdraw(_recipient, withdrawAmount);
      return withdrawAmount;
    } else {
      uint256 amountAvailableToPayLSTLiability = withdrawAmount - _targetReserveDeficit;
      uint256 currentLSTLiabilityETH = _getCurrentLSTLiabilityETH(_yieldProvider);
      if (currentLSTLiabilityETH > 0) {
        uint256 LSTLiabilityETHPaid = currentLSTLiabilityETH > amountAvailableToPayLSTLiability ? amountAvailableToPayLSTLiability : currentLSTLiabilityETH;
        IDashboard(_getDashboard(_yieldProvider)).rebalanceVaultWithEther(LSTLiabilityETHPaid);
        withdrawAmount -= LSTLiabilityETHPaid;
      }
      ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).withdraw(_recipient, withdrawAmount);
      return withdrawAmount;
    }
  }

  /**
   * @notice Rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev Settles any outstanding LST liabilities, provided this does not leave the withdrawal reserve in deficit.
   * @param _amount                 Amount to withdraw.
   */
  function addToWithdrawalReserve(address _yieldProvider, uint256 _amount) external {

  }

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   */
  function pauseStaking(address _yieldProvider) external {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).pauseBeaconChainDeposits();
  }

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   */
  function unpauseStaking(address _yieldProvider) external {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).resumeBeaconChainDeposits();
  }

  function validateAdditionToYieldManager(IYieldManager.YieldProviderRegistration calldata _yieldProviderRegistration) external pure {
    if (_yieldProviderRegistration.yieldProviderType != IYieldManager.YieldProviderType.LIDO_STVAULT) {
      revert IncorrectYieldProviderType();
    }
  }

  function getAvailableBalanceForWithdraw(address _yieldProvider) external view returns (uint256) {
    return _getStakingVault(_yieldProvider).balance;
  }

  function mintLST(address _yieldProvider, uint256 _amount, address _recipient) external {
    if (_getYieldProviderDataStorage(_yieldProvider).isOssified) {
      revert OperationNotSupportedDuringOssification(OperationType.MintLST);
    }
    IDashboard(_getDashboard(_yieldProvider)).mintStETH(_recipient, _amount);
  }
}