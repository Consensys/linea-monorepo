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
import { Math256 } from "../libraries/Math256.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldManagerStorageLayout, IYieldProvider, IGenericErrors {
  // yieldProvider = StakingVault
  address immutable YIELD_PROVIDER;
  IVaultHub immutable VAULT_HUB;
  IDashboard immutable DASHBOARD;
  IStETH immutable STETH;

  // @dev _yieldProvider = stakingVault address
  constructor (address _yieldProvider, address _vaultHub, address _dashboard, address _steth) {
    // Do checks
    YIELD_PROVIDER = _yieldProvider;
    VAULT_HUB = IVaultHub(_vaultHub);
    DASHBOARD = IDashboard(_dashboard);
    STETH = IStETH(_steth);
  }

  function _getEntrypointContract() private view returns (address) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(YIELD_PROVIDER);
    return $$.isOssified ? YIELD_PROVIDER : address(DASHBOARD);
  }

  function _getStakingVault() private view returns (address) {
    return YIELD_PROVIDER;
  }

  // Will settle as much LST liability as possible. Will return amount of liabilityEth remaining
  // @dev - we consider principal to be paid before interest. Reason being that Lido doesn't distinguish between the two, so we can choose which is more convenient for us as long as we remain consistent.
  function _payMaximumPossibleLSTLiability() internal {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(YIELD_PROVIDER);
    if ($$.isOssified) return;

    uint256 vaultBalanceEth = DASHBOARD.totalValue();
    uint256 vaultBalanceShares = STETH.getSharesByPooledEth(vaultBalanceEth);
    uint256 liabilityShares = DASHBOARD.liabilityShares();
    // Do the payment
    uint256 beforeVaultBalance = YIELD_PROVIDER.balance;
    DASHBOARD.rebalanceVaultWithShares(Math256.min(liabilityShares, vaultBalanceShares));
    uint256 afterVaultBalance = YIELD_PROVIDER.balance;
    uint256 rebalanceAmountETH = beforeVaultBalance - afterVaultBalance;
    // Adjust LST principal
    // Break rule to handle $$ in YieldProvider, because LST liability being paid in ossification is a Lido-specific concept
    uint256 lstLiabilityPrincipal = $$.lstLiabilityPrincipal;
    $$.lstLiabilityPrincipal = Math256.safeSub(lstLiabilityPrincipal, rebalanceAmountETH);
  }

  function payLSTPrincipal(uint256 _maxAvailableRepaymentETH) external returns (uint256) {
    return _payLSTPrincipal(_maxAvailableRepaymentETH);
  }

  function _payLSTPrincipal(uint256 _maxAvailableRepaymentETH) internal returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(YIELD_PROVIDER);
    if ($$.isOssified) return 0;
    uint256 lstLiabilityPrincipal = $$.lstLiabilityPrincipal;
    if (lstLiabilityPrincipal == 0) return 0;
    uint256 rebalanceAmount = Math256.min(lstLiabilityPrincipal, _maxAvailableRepaymentETH);
    DASHBOARD.rebalanceVaultWithShares(rebalanceAmount);
    return rebalanceAmount;
  }

  // Returns how much of _maxAvailableRepaymentETH available, after LST interest payment
  // @dev Redemption component of obligations, and liability - are decremented in tandem in Lido VaultHub
  function _payLSTInterest(uint256 _maxAvailableRepaymentETH) internal returns (uint256) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(YIELD_PROVIDER);
    if ($$.isOssified) return _maxAvailableRepaymentETH;
    uint256 liabilityTotalShares = DASHBOARD.liabilityShares();
    if (liabilityTotalShares == 0) return _maxAvailableRepaymentETH;
    uint256 liabilityInterestETH = STETH.getPooledEthBySharesRoundUp(liabilityTotalShares) - $$.lstLiabilityPrincipal;
    uint256 lstInterestRepaymentETH = Math256.min(liabilityInterestETH, _maxAvailableRepaymentETH);
    // Do the payment
    DASHBOARD.rebalanceVaultWithEther(lstInterestRepaymentETH);
    return _maxAvailableRepaymentETH - lstInterestRepaymentETH;
  }
  
  // @dev Must consider that this can pay redemptions
  function _payObligations(uint256 _maxAvailableRepaymentETH) internal returns (uint256) {
    uint256 beforeVaultBalance = YIELD_PROVIDER.balance;
    // Unfortunately, there is no function on VaultHub to specify how much obligation we want to repay.
    VAULT_HUB.settleVaultObligations(YIELD_PROVIDER);
    uint256 afterVaultBalance = YIELD_PROVIDER.balance;
    uint256 obligationsPaid = afterVaultBalance - beforeVaultBalance;

    if (obligationsPaid > _maxAvailableRepaymentETH) {
      _getYieldProviderDataStorage(YIELD_PROVIDER).currentNegativeYield += (obligationsPaid - _maxAvailableRepaymentETH);
      return 0;
    } else {
      return _maxAvailableRepaymentETH - obligationsPaid;
    }
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(uint256 _amount) external {
    ICommonVaultOperations(_getEntrypointContract()).fund{value: _amount}();
  }

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev Note here we have broken our pattern that we handle all state mutation in the YieldManager, we will revisit this
   * @dev Both `redemptions` and `obligations` can both be reduced permissionlessly via VaultHub.settleObligations(). Then this is accounted within negative yield.
   */
  function reportYield() external returns (uint256 _newYield) {
    IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(YIELD_PROVIDER);
    if ($$.isOssified) {
      revert OperationNotSupportedDuringOssification(OperationType.ReportYield);
    }
    // First compute the total yield
    uint256 lastUserFunds = $$.userFunds;
    uint256 totalVaultFunds = DASHBOARD.totalValue();
    uint256 negativeTotalYield;
    uint256 positiveTotalYield;
    if (totalVaultFunds > lastUserFunds) {
      positiveTotalYield = totalVaultFunds - lastUserFunds;
    } else {
      negativeTotalYield = lastUserFunds - totalVaultFunds;
    }
    if (positiveTotalYield > 0) {
      positiveTotalYield = _handlePostiveYieldAccounting(positiveTotalYield);
      // emit event
      return positiveTotalYield;
    } else if (negativeTotalYield > 0) {
      // If prev reportYield had negativeYield, this correctly accounts for both increase and reduction in negativeYield since
      $$.currentNegativeYield = negativeTotalYield;
      // emit event
      return 0;
    }
      
  }

  // TODO - What if we incur redemption, and someone forces the settlement of this...

  function _handlePostiveYieldAccounting(uint256 positiveTotalYield) internal returns (uint256) {
      IYieldManager.YieldProviderData storage $$ = _getYieldProviderDataStorage(YIELD_PROVIDER);
      // First pay negative yield
      uint256 positiveRemainingYield = positiveTotalYield;
      uint256 currentNegativeYield = $$.currentNegativeYield;
      if (currentNegativeYield > 0) {
        uint256 negativeYieldReduction = Math256.min(currentNegativeYield, positiveRemainingYield);
        $$.currentNegativeYield -= negativeYieldReduction;
        positiveRemainingYield -= negativeYieldReduction;
      }
      // Then pay liability interest
      positiveRemainingYield = _payLSTInterest(positiveRemainingYield);
      positiveRemainingYield = _payObligations(positiveRemainingYield);
      $$.userFunds += positiveRemainingYield;
      $$.yieldReportedCumulative += positiveRemainingYield;
      return positiveRemainingYield;
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
    return _getYieldProviderDataStorage(YIELD_PROVIDER).isOssified ? YIELD_PROVIDER.balance : DASHBOARD.withdrawableValue();
  }

  function mintLST(uint256 _amount, address _recipient) external {
    DASHBOARD.mintStETH(_recipient, _amount);
  }

  // TODO - Role
  // @dev Requires fresh report
  function initiateOssification() external {
    _payMaximumPossibleLSTLiability();
    
    // Lido implementation handles Lido fee payment, and revert on fresh report
    // This will fail if any existing liabilities or obligations
    DASHBOARD.voluntaryDisconnect();
  }

  // TODO - Role
  // Returns true if ossified after function call is done
  function processPendingOssification() external returns (bool) {
    // Give ownership to YieldManager
    DASHBOARD.abandonDashboard(address(this));
    IStakingVault vault = IStakingVault(YIELD_PROVIDER);
    vault.acceptOwnership();
    vault.ossify();
    return true;
  }
}
