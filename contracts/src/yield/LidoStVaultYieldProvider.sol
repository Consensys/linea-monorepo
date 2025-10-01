// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldProviderBase } from "./YieldProviderBase.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { ICommonVaultOperations } from "./interfaces/vendor/lido/ICommonVaultOperations.sol";
import { IDashboard } from "./interfaces/vendor/lido/IDashboard.sol";
import { IStETH } from "./interfaces/vendor/lido/IStETH.sol";
import { IVaultHub } from "./interfaces/vendor/lido/IVaultHub.sol";
import { IStakingVault } from "./interfaces/vendor/lido/IStakingVault.sol";
import { Math256 } from "../libraries/Math256.sol";
import { CLProofVerifier } from "./libs/CLProofVerifier.sol";
import { GIndex } from "./libs/vendor/lido/GIndex.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldProviderBase, CLProofVerifier, IGenericErrors {
  uint256 private constant PUBLIC_KEY_LENGTH = 48;
  uint256 private constant MIN_0X02_VALIDATOR_ACTIVATION_BALANCE = 32 ether;
  IVaultHub immutable VAULT_HUB;
  IStETH immutable STETH;

  // yieldProvider = StakingVault
  // address immutable YIELD_PROVIDER;
  // IDashboard immutable DASHBOARD;
  // bytes32 immutable WITHDRAWAL_CREDENTIALS;

  event LidoVaultUnstakePermissionlessRequest(
    address indexed yieldProvider,
    address indexed caller,
    address indexed refundRecipient,
    uint256 maxUnstakeAmount,
    bytes pubkeys,
    uint64[] amounts
  );

  // @dev _yieldProvider = stakingVault address
  constructor(
    address _vaultHub,
    address _steth,
    GIndex _gIFirstValidator,
    GIndex _gIFirstValidatorAfterChange,
    uint64 _changeSlot
  ) CLProofVerifier(_gIFirstValidator, _gIFirstValidatorAfterChange, _changeSlot) {
    // Do checks
    VAULT_HUB = IVaultHub(_vaultHub);
    STETH = IStETH(_steth);
  }

  function _getEntrypointContract(address _yieldProvider) internal view returns (address entrypointContract) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    entrypointContract = $$.isOssified ? $$.ossifiedEntrypoint : $$.primaryEntrypoint;
  }

  function withdrawableValue(address _yieldProvider) external view onlyDelegateCall returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return $$.isOssified ? $$.ossifiedEntrypoint.balance : IDashboard($$.primaryEntrypoint).withdrawableValue();
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @param _amount        The amount of ETH to send.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).fund{ value: _amount }();
  }

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev Note here we have broken our pattern that we handle all state mutation in the YieldManager, we will revisit this
   * @dev Both `redemptions` and `obligations` can both be reduced permissionlessly via VaultHub.settleObligations(). Then this is accounted within negative yield.
   */
  function reportYield(address _yieldProvider) external onlyDelegateCall returns (uint256 newReportedYield) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssified) {
      revert OperationNotSupportedDuringOssification(OperationType.ReportYield);
    }
    // First compute the total yield
    uint256 lastUserFunds = $$.userFunds;
    uint256 totalVaultFunds = IDashboard($$.primaryEntrypoint).totalValue();
    uint256 negativeTotalYield;
    uint256 positiveTotalYield;
    if (totalVaultFunds > lastUserFunds) {
      positiveTotalYield = totalVaultFunds - lastUserFunds;
    } else {
      negativeTotalYield = lastUserFunds - totalVaultFunds;
    }

    if (positiveTotalYield > 0) {
      positiveTotalYield = _handlePostiveYieldAccounting(_yieldProvider, positiveTotalYield);
      // emit event
      return positiveTotalYield;
    } else if (negativeTotalYield > 0) {
      // If prev reportYield had negativeYield, this correctly accounts for both increase and reduction in negativeYield since
      $$.currentNegativeYield = negativeTotalYield;
      // emit event
      return 0;
    }
  }

  function _handlePostiveYieldAccounting(address _yieldProvider, uint256 positiveTotalYield) internal returns (uint256 newReportedYield) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    // First pay negative yield
    newReportedYield = positiveTotalYield;
    uint256 currentNegativeYield = $$.currentNegativeYield;
    if (currentNegativeYield > 0) {
      uint256 negativeYieldReduction = Math256.min(currentNegativeYield, newReportedYield);
      $$.currentNegativeYield -= negativeYieldReduction;
      newReportedYield -= negativeYieldReduction;
    }
    // Then pay liability interest
    newReportedYield -= _payLSTInterest(_yieldProvider, newReportedYield);
    // Then pay obligations
    newReportedYield -= _payObligations(_yieldProvider, newReportedYield);
    // Then pay node operator fee(s)
    newReportedYield -= _payNodeOperatorFees(_yieldProvider, newReportedYield);
  }

  // Returns how much of _maxAvailableRepaymentETH available, after LST interest payment
  // @dev Redemption component of obligations, and liability - are decremented in tandem in Lido VaultHub
  function _payLSTInterest(address _yieldProvider, uint256 _maxAvailableRepaymentETH) internal returns (uint256 lstInterestPaid) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssified) return _maxAvailableRepaymentETH;
    IDashboard dashboard = IDashboard($$.primaryEntrypoint);
    uint256 liabilityTotalShares = dashboard.liabilityShares();
    if (liabilityTotalShares == 0) return _maxAvailableRepaymentETH;
    (uint256 lstLiabilityPrincipalSynced, bool isLstLiabilityPrincipalChanged) = _syncExternalLiabilitySettlement(
      _yieldProvider,
      liabilityTotalShares,
      $$.lstLiabilityPrincipal
    );
    // If lstLiabilityPrincipal was reduced by _syncExternalLiabilitySettlement(), it means all LST interest has been paid
    if (!isLstLiabilityPrincipalChanged) {
      return _maxAvailableRepaymentETH;
    }
    uint256 liabilityInterestETH = STETH.getPooledEthBySharesRoundUp(liabilityTotalShares) -
      lstLiabilityPrincipalSynced;
    lstInterestPaid = Math256.min(liabilityInterestETH, _maxAvailableRepaymentETH);
    // Do the payment
    if (lstInterestPaid > 0) {
      dashboard.rebalanceVaultWithEther(lstInterestPaid);
    }
  }

  // Check if lstPrincipal < ETH value of liabilityShares
  // If true, this means obligations were accrued and settled - settleVaultObligations is permissionless so this could have been us or another entity.
  // This conservatively assumes that interest was settled first, then the leftover is allocated to payment.
  // This is a reasonable approach, because we actually cannot compute the principal/liability split without keeping track of the time that we accrued and reduced lstLiability.
  // @dev May reduce $$.lstLiabilityPrincipal
  // @return New value of lstLiabilityPrincipal
  function _syncExternalLiabilitySettlement(
    address _yieldProvider,
    uint256 _liabilityShares,
    uint256 _lstLiabilityPrincipalCached
  ) internal returns (uint256 lstLiabilityPrincipalSynced, bool isLstLiabilityPrincipalChanged) {
    uint256 liabilityETH = STETH.getPooledEthBySharesRoundUp(_liabilityShares);
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if (liabilityETH < _lstLiabilityPrincipalCached) {
      $$.lstLiabilityPrincipal = liabilityETH;
      return (liabilityETH, true);
    } else {
      return (_lstLiabilityPrincipalCached, false);
    }
  }

  // @dev Obligations can include redemptions, which gets reduced in tandem with liabilities
  // @dev We will not eagerly track redemptions paid through this function, because redemptions can be paid permissionlessly
  // @dev Therefore it is not possible for us to eagerly track all redemption changes
  // @dev From a user funds POV, we isolate permissionless settlement by accounting it as negative yield
  // @dev From an LST liability principal accounting POV, the main issue not eagerly tracking redemptions will cause
  //     is that we will overpay an LST liability payment, and our rebalance() call will fail because the system thinks
  //     it has more debt than it actually has. We handle this by checking if this has happened, and adjusting lstLiabilityPrincipal accordingly via _syncExternalLiabilitySettlement.
  function _payObligations(address _yieldProvider, uint256 _maxAvailableRepaymentETH) internal returns (uint256 obligationsPaid) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    address vault = $$.ossifiedEntrypoint;
    uint256 beforeVaultBalance = vault.balance;
    // Unfortunately, there is no function on VaultHub to specify how much obligation we want to repay.
    VAULT_HUB.settleVaultObligations(vault);
    uint256 afterVaultBalance = vault.balance;
    obligationsPaid = afterVaultBalance - beforeVaultBalance;
    if (obligationsPaid > _maxAvailableRepaymentETH) {
      _getYieldProviderStorage(vault).currentNegativeYield += (obligationsPaid - _maxAvailableRepaymentETH);
    }
  }

  function _payNodeOperatorFees(address _yieldProvider, uint256 _availableAmount) internal returns (uint256 nodeOperatorFeesPaid) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    IDashboard dashboard = IDashboard($$.primaryEntrypoint);
    address vault = $$.ossifiedEntrypoint;
    uint256 currentFees = dashboard.nodeOperatorDisbursableFee();
    uint256 vaultBalance = vault.balance;
    // Does not allow partial payment of node operator fees, unlike settleVaultObligations
    if (vaultBalance > currentFees) {
      dashboard.disburseNodeOperatorFee();
      nodeOperatorFeesPaid = currentFees;
      if (nodeOperatorFeesPaid >= _availableAmount) {
        $$.currentNegativeYield += (nodeOperatorFeesPaid - _availableAmount);
      }
    }
  }

  // @dev LST Principal reduction from discovered external sync, does not count as payment
  // @dev Guard to validate against ossification is done on the YieldManager
  function payLSTPrincipal(
    address _yieldProvider,
    uint256 _maxAvailableRepaymentETH
  ) external onlyDelegateCall returns (uint256 lstPrincipalPaid) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssificationInitiated || $$.isOssified) {
      return 0;
    }
    lstPrincipalPaid = _payLSTPrincipal(_yieldProvider, _maxAvailableRepaymentETH);
  }

  function _payLSTPrincipal(address _yieldProvider, uint256 _maxAvailableRepaymentETH) internal returns (uint256 lstPrincipalPaid) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssified) return 0;
    uint256 lstLiabilityPrincipalCached = $$.lstLiabilityPrincipal;
    if (lstLiabilityPrincipalCached == 0) return 0;
    IDashboard dashboard = IDashboard($$.primaryEntrypoint);
    (uint256 lstLiabilityPrincipalSynced, ) = _syncExternalLiabilitySettlement(
      _yieldProvider,
      dashboard.liabilityShares(),
      lstLiabilityPrincipalCached
    );
    uint256 rebalanceAmount = Math256.min(lstLiabilityPrincipalSynced, _maxAvailableRepaymentETH);
    if (rebalanceAmount > 0) {
      dashboard.rebalanceVaultWithShares(rebalanceAmount);
    }
    $$.lstLiabilityPrincipal -= rebalanceAmount;
    lstPrincipalPaid = rebalanceAmount;
  }

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable onlyDelegateCall {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(
      _withdrawalParams,
      (bytes, uint64[], address)
    );
    _unstake(_yieldProvider, pubkeys, amounts, refundRecipient);
    // Intentional choice to not emit event as downstream StakingVault will emit ValidatorWithdrawalsTriggered event.
  }

  function _unstake(address _yieldProvider, bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) internal {
    // Lido StakingVault.sol will handle the param validation
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).triggerValidatorWithdrawals{ value: msg.value }(
      pubkeys,
      amounts,
      refundRecipient
    );
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
  ) external payable onlyDelegateCall returns (uint256 maxUnstakeAmount) {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(
      _withdrawalParams,
      (bytes, uint64[], address)
    );
    maxUnstakeAmount = _validateUnstakePermissionless(_yieldProvider, pubkeys, amounts, _withdrawalParamsProof);
    _unstake(_yieldProvider, pubkeys, amounts, refundRecipient);

    emit LidoVaultUnstakePermissionlessRequest(
       _getYieldProviderStorage(_yieldProvider).ossifiedEntrypoint,
      msg.sender,
      refundRecipient,
      maxUnstakeAmount,
      pubkeys,
      amounts
    );
  }

  // @dev Checks guided by https://github.com/ethereum/consensus-specs/blob/834e40604ae4411e565bd6540da50b008b2496dc/specs/electra/beacon-chain.md#new-process_withdrawal_request
  function _validateUnstakePermissionless(
    address _yieldProvider,
    bytes memory pubkeys,
    uint64[] memory amounts,
    bytes calldata _withdrawalParamsProof
  ) internal view returns (uint256 maxUnstakeAmount) {
    // Length validator
    if (pubkeys.length != PUBLIC_KEY_LENGTH || amounts.length != 1) {
      revert SingleValidatorOnlyForUnstakePermissionless();
    }

    uint256 amount = amounts[0];
    if (amount == 0) {
      revert NoValidatorExitForUnstakePermissionless();
    }

    ValidatorWitness memory witness = abi.decode(_withdrawalParamsProof, (ValidatorWitness));

    // 0x02 withdrawal credential scheme
    address vault = _getYieldProviderStorage(_yieldProvider).ossifiedEntrypoint;
    bytes32 withdrawalCredentials;
    assembly {
      withdrawalCredentials := or(shl(248, 0x2), _yieldProvider)
    }

    _validateValidatorContainerForPermissionlessUnstake(witness, withdrawalCredentials);

    /** 
      The consensus specs specify this as 
    
       to_withdraw = min(
            state.balances[index] - MIN_ACTIVATION_BALANCE - pending_balance_to_withdraw,
            amount
        )    
      
      We will not keep track of 'pending_balance_to_withdraw'.
      It is enough that $.pendingPermissionlessWithdrawal is decremented on every ETH transfer to L1MessageService.
    */
    maxUnstakeAmount = Math256.min(amount, witness.effectiveBalance - MIN_0X02_VALIDATOR_ACTIVATION_BALANCE);
  }

  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount, address _recipient) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).withdraw(_recipient, _amount);
  }

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   */
  function pauseStaking(address _yieldProvider) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).pauseBeaconChainDeposits();
  }

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   */
  function unpauseStaking(address _yieldProvider) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).resumeBeaconChainDeposits();
  }

  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external onlyDelegateCall {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssificationInitiated || $$.isOssified) {
      revert MintLSTDisabledDuringOssification();
    }
    IDashboard($$.primaryEntrypoint).mintStETH(_recipient, _amount);
    $$.lstLiabilityPrincipal += _amount;
  }

  // TODO - Role
  // @dev Requires fresh report
  function initiateOssification(address _yieldProvider) external onlyDelegateCall {
    _payMaximumPossibleLSTLiability(_yieldProvider);
    // Lido implementation handles Lido fee payment, and revert on fresh report
    // This will fail if any existing liabilities or obligations
    IDashboard(_getYieldProviderStorage(_yieldProvider).primaryEntrypoint).voluntaryDisconnect();
  }

  // Will settle as much LST liability as possible. Will return amount of liabilityEth remaining
  // Settle interest before principal
  function _payMaximumPossibleLSTLiability(address _yieldProvider) internal {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssified) return;
    IDashboard dashboard = IDashboard($$.primaryEntrypoint);
    address vault = $$.ossifiedEntrypoint;
    // Assumption - this is maximum available for rebalance
    uint256 availableRebalanceAmount = vault.balance;
    uint256 availableRebalanceShares = STETH.getSharesByPooledEth(availableRebalanceAmount);
    uint256 liabilityShares = dashboard.liabilityShares();
    uint256 rebalanceShares = Math256.min(liabilityShares, availableRebalanceShares);
    if (rebalanceShares > 0) {
      dashboard.rebalanceVaultWithShares(rebalanceShares);
      // Apply consistent accounting treatment that LST interest paid first, then LST principal
      _syncExternalLiabilitySettlement(_yieldProvider, dashboard.liabilityShares(), $$.lstLiabilityPrincipal);
    }
  }

  /**
   * @notice Start the ossification process for the yield provider.
   */
  function undoInitiateOssification(address _yieldProvider) external onlyDelegateCall {
    // No-op
  }

  // TODO - Role
  // Returns true if ossified after function call is done
  function processPendingOssification(address _yieldProvider) external onlyDelegateCall returns (bool isOssificationComplete) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    // Give ownership to YieldManager
    IDashboard($$.primaryEntrypoint).abandonDashboard(address(this));
    IStakingVault vault = IStakingVault($$.ossifiedEntrypoint);
    vault.acceptOwnership();
    vault.ossify();
    isOssificationComplete = true;
  }

  function validateAdditionToYieldManager(YieldProviderRegistration calldata _registration) external pure {
    if (_registration.yieldProviderVendor != YieldProviderVendor.LIDO_STVAULT) {
      revert UnknownYieldProviderVendor();
    }
  }
}
