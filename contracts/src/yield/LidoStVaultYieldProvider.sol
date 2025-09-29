// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldManagerStorageLayout } from "./YieldManagerStorageLayout.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { ICommonVaultOperations } from "./interfaces/vendor/lido/ICommonVaultOperations.sol";
import { IDashboard } from "./interfaces/vendor/lido/IDashboard.sol";
import { IStETH } from "./interfaces/vendor/lido/IStETH.sol";
import { IVaultHub } from "./interfaces/vendor/lido/IVaultHub.sol";
import { IStakingVault } from "./interfaces/vendor/lido/IStakingVault.sol";
import { Math256 } from "../libraries/Math256.sol";
import { CLProofVerifier } from "./libs/CLProofVerifier.sol";
import {GIndex} from "./libs/vendor/lido/GIndex.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldManagerStorageLayout, CLProofVerifier, IYieldProvider, IGenericErrors {
  // yieldProvider = StakingVault
  address immutable YIELD_PROVIDER;
  IVaultHub immutable VAULT_HUB;
  IDashboard immutable DASHBOARD;
  IStETH immutable STETH;
  bytes32 immutable WITHDRAWAL_CREDENTIALS;

  uint256 private constant PUBLIC_KEY_LENGTH = 48;
  uint256 private constant MIN_0X02_VALIDATOR_ACTIVATION_BALANCE = 32 ether;
  // Validator must be active for this many epochs before it is eligible for withdrawals
  uint256 private constant SHARD_COMMITTEE_PERIOD = 256;
  uint256 private constant SLOTS_PER_EPOCH = 32;

  // @dev _yieldProvider = stakingVault address
  constructor (
    address _yieldProvider, 
    address _vaultHub, 
    address _dashboard, 
    address _steth,
    GIndex _gIFirstValidator,
    GIndex _gIFirstValidatorAfterChange,
    uint64 _changeSlot
  ) CLProofVerifier(_gIFirstValidator, _gIFirstValidatorAfterChange, _changeSlot) {
    // Do checks
    YIELD_PROVIDER = _yieldProvider;
    VAULT_HUB = IVaultHub(_vaultHub);
    DASHBOARD = IDashboard(_dashboard);
    STETH = IStETH(_steth);
    // 0x02 withdrawal credential scheme
    bytes32 withdrawalCredentials;
    assembly {
      withdrawalCredentials := or(shl(248, 0x2), _yieldProvider)
    }
    WITHDRAWAL_CREDENTIALS = withdrawalCredentials;
  }

  function _getEntrypointContract() private view returns (address) {
    return _getYieldProviderStorage(YIELD_PROVIDER).isOssified ? YIELD_PROVIDER : address(DASHBOARD);
  }

  function _getStakingVault() private view returns (address) {
    return YIELD_PROVIDER;
  }

  // Will settle as much LST liability as possible. Will return amount of liabilityEth remaining
  // Settle interest before principal
  function _payMaximumPossibleLSTLiability() internal {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
    if ($$.isOssified) return;
    // Assumption - this is maximum available for rebalance
    uint256 availableRebalanceAmount = YIELD_PROVIDER.balance;
    uint256 availableRebalanceShares = STETH.getSharesByPooledEth(availableRebalanceAmount);
    uint256 liabilityShares = DASHBOARD.liabilityShares();
    uint256 rebalanceShares = Math256.min(liabilityShares, availableRebalanceShares);
    if (rebalanceShares > 0) {
      DASHBOARD.rebalanceVaultWithShares(rebalanceShares);
      // Apply consistent accounting treatment that LST interest paid first, then LST principal
      _syncExternalLiabilitySettlement(DASHBOARD.liabilityShares(), $$.lstLiabilityPrincipal);
    }
  }

  // @dev LST Principal reduction from discovered external sync, does not count as payment
  // @dev Guard to validate against ossification is done on the YieldManager
  function payLSTPrincipal(uint256 _maxAvailableRepaymentETH) external returns (uint256 lstPrincipalPayment) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
    if ($$.isOssificationInitiated || $$.isOssified) {
      return 0;
    }
    return _payLSTPrincipal(_maxAvailableRepaymentETH);
  }

  function _payLSTPrincipal(uint256 _maxAvailableRepaymentETH) internal returns (uint256 lstPrincipalPayment) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
    if ($$.isOssified) return 0;
    uint256 lstLiabilityPrincipalCached = $$.lstLiabilityPrincipal;
    if (lstLiabilityPrincipalCached == 0) return 0;
    (uint256 lstLiabilityPrincipalSynced, ) = _syncExternalLiabilitySettlement(DASHBOARD.liabilityShares(), lstLiabilityPrincipalCached);
    uint256 rebalanceAmount = Math256.min(lstLiabilityPrincipalSynced, _maxAvailableRepaymentETH);
    if (rebalanceAmount > 0) {
      DASHBOARD.rebalanceVaultWithShares(rebalanceAmount);
    }
    $$.lstLiabilityPrincipal -= rebalanceAmount;
    lstPrincipalPayment = rebalanceAmount;
  }

  // Check if lstPrincipal < ETH value of liabilityShares
  // If true, this means obligations were accrued and settled - settleVaultObligations is permissionless so this could have been us or another entity.
  // This conservatively assumes that interest was settled first, then the leftover is allocated to payment.
  // This is a reasonable approach, because we actually cannot compute the principal/liability split without keeping track of the time that we accrued and reduced lstLiability.
  // @dev May reduce $$.lstLiabilityPrincipal 
  // @return New value of lstLiabilityPrincipal
  function _syncExternalLiabilitySettlement(uint256 liabilityShares, uint256 lstLiabilityPrincipalCached) internal returns (uint256 lstLiabilityPrincipal, bool isLstLiabilityPrincipalChanged) {
    // uint256 liabilityShares = DASHBOARD.liabilityShares();
    uint256 liabilityETH = STETH.getPooledEthBySharesRoundUp(liabilityShares);
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
    // uint256 lstLiabilityPrincipalCached = $$.lstLiabilityPrincipal;
    if (liabilityETH < lstLiabilityPrincipalCached) {
      $$.lstLiabilityPrincipal = liabilityETH;
      return (liabilityETH, true);
    } else {
      return (lstLiabilityPrincipalCached, false);
    }
  }

  // Returns how much of _maxAvailableRepaymentETH available, after LST interest payment
  // @dev Redemption component of obligations, and liability - are decremented in tandem in Lido VaultHub
  function _payLSTInterest(uint256 _maxAvailableRepaymentETH) internal returns (uint256 payment) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
    if ($$.isOssified) return _maxAvailableRepaymentETH;
    uint256 liabilityTotalShares = DASHBOARD.liabilityShares();
    if (liabilityTotalShares == 0) return _maxAvailableRepaymentETH;
    (uint256 lstLiabilityPrincipalSynced, bool isLstLiabilityPrincipalChanged) = _syncExternalLiabilitySettlement(liabilityTotalShares, $$.lstLiabilityPrincipal);
    // If lstLiabilityPrincipal was reduced by _syncExternalLiabilitySettlement(), it means all LST interest has been paid
    if (!isLstLiabilityPrincipalChanged) {
      return _maxAvailableRepaymentETH;
    }
    uint256 liabilityInterestETH = STETH.getPooledEthBySharesRoundUp(liabilityTotalShares) - lstLiabilityPrincipalSynced;
    uint256 lstInterestRepaymentETH = Math256.min(liabilityInterestETH, _maxAvailableRepaymentETH);
    // Do the payment
    if (lstInterestRepaymentETH > 0) {
      DASHBOARD.rebalanceVaultWithEther(lstInterestRepaymentETH);
    }
    payment = lstInterestRepaymentETH;
  }
  
  // @dev Obligations can include redemptions, which gets reduced in tandem with liabilities
  // @dev We will not eagerly track redemptions paid through this function, because redemptions can be paid permissionlessly
  // @dev Therefore it is not possible for us to eagerly track all redemption changes
  // @dev From a user funds POV, we isolate permissionless settlement by accounting it as negative yield
  // @dev From an LST liability principal accounting POV, the main issue not eagerly tracking redemptions will cause
  //     is that we will overpay an LST liability payment, and our rebalance() call will fail because the system thinks
  //     it has more debt than it actually has. We handle this by checking if this has happened, and adjusting lstLiabilityPrincipal accordingly via _syncExternalLiabilitySettlement.
  function _payObligations(uint256 _maxAvailableRepaymentETH) internal returns (uint256 obligationsPaid) {
    uint256 beforeVaultBalance = YIELD_PROVIDER.balance;
    // Unfortunately, there is no function on VaultHub to specify how much obligation we want to repay.
    VAULT_HUB.settleVaultObligations(YIELD_PROVIDER);
    uint256 afterVaultBalance = YIELD_PROVIDER.balance;
    obligationsPaid = afterVaultBalance - beforeVaultBalance;
    if (obligationsPaid > _maxAvailableRepaymentETH) {
      _getYieldProviderStorage(YIELD_PROVIDER).currentNegativeYield += (obligationsPaid - _maxAvailableRepaymentETH);
    }
  }

  function _payNodeOperatorFees(uint256 _availableAmount) internal returns (uint256 payment) {
    uint256 currentFees = DASHBOARD.nodeOperatorDisbursableFee();
    uint256 vaultBalance = YIELD_PROVIDER.balance;
    // Does not allow partial payment of node operator fees, unlike settleVaultObligations
    if (vaultBalance > currentFees) {
      DASHBOARD.disburseNodeOperatorFee();
      payment = currentFees;
      if (payment >= _availableAmount) {
        _getYieldProviderStorage(YIELD_PROVIDER).currentNegativeYield += (payment - _availableAmount);
      }
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
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
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

  function _handlePostiveYieldAccounting(uint256 positiveTotalYield) internal returns (uint256) {
      YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
      // First pay negative yield
      uint256 positiveRemainingYield = positiveTotalYield;
      uint256 currentNegativeYield = $$.currentNegativeYield;
      if (currentNegativeYield > 0) {
        uint256 negativeYieldReduction = Math256.min(currentNegativeYield, positiveRemainingYield);
        $$.currentNegativeYield -= negativeYieldReduction;
        positiveRemainingYield -= negativeYieldReduction;
      }
      // Then pay liability interest
      positiveRemainingYield -= _payLSTInterest(positiveRemainingYield);
      // Then pay obligations
      positiveRemainingYield -= _payObligations(positiveRemainingYield);
      // Then pay node operator fee(s)
      positiveRemainingYield -= _payNodeOperatorFees(positiveRemainingYield);
      return positiveRemainingYield;
  }

  /**
   * @notice Request beacon chain withdrawal.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   */
  function unstake(bytes memory _withdrawalParams) external payable {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(_withdrawalParams, (bytes, uint64[], address));
    _unstake(pubkeys, amounts, refundRecipient);
    // TODO - Emit event
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
  ) external payable returns (uint256) {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(_withdrawalParams, (bytes, uint64[], address));
    uint256 maxPossibleUnstakeAmount = _validateUnstakePermissionless(pubkeys, amounts, _withdrawalParamsProof);
    _unstake(pubkeys, amounts, refundRecipient);

    // TODO - Emit event
    return maxPossibleUnstakeAmount;
  }

  function _unstake(bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) internal {
    // Lido StakingVault.sol will handle the param validation
    ICommonVaultOperations(_getEntrypointContract()).triggerValidatorWithdrawals{value: msg.value}(pubkeys, amounts, refundRecipient);
  }

  // @dev Checks guided by https://github.com/ethereum/consensus-specs/blob/834e40604ae4411e565bd6540da50b008b2496dc/specs/electra/beacon-chain.md#new-process_withdrawal_request
  function _validateUnstakePermissionless(bytes memory pubkeys, uint64[] memory amounts, bytes calldata _withdrawalParamsProof) internal view returns (uint256 maxPossibleUnstake) {
    // Length validator
    if (pubkeys.length != PUBLIC_KEY_LENGTH || amounts.length != 1 ) {
      revert SingleValidatorOnlyForUnstakePermissionless();
    }
    
    uint256 amount = amounts[0];
    if (amount == 0) {
      revert NoValidatorExitForUnstakePermissionless();
    }
    
    (ValidatorWitness memory witness) = abi.decode(_withdrawalParamsProof, (ValidatorWitness));
    uint256 epoch = witness.slot / SLOTS_PER_EPOCH;
    if (epoch < witness.activationEpoch + SHARD_COMMITTEE_PERIOD) {
      revert ValidatorNotActiveForLongEnough();
    }

    _validateValidatorContainerForPermissionlessUnstake(witness, WITHDRAWAL_CREDENTIALS);

    /** 
      The consensus specs specify this as 
    
       to_withdraw = min(
            state.balances[index] - MIN_ACTIVATION_BALANCE - pending_balance_to_withdraw,
            amount
        )    
      
      We will not keep track of 'pending_balance_to_withdraw'.
      It is enough that $.pendingPermissionlessWithdrawal is decremented on every ETH transfer to L1MessageService.
    */ 
    maxPossibleUnstake = Math256.min(amount, witness.effectiveBalance - MIN_0X02_VALIDATOR_ACTIVATION_BALANCE);
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

  function validateAdditionToYieldManager(YieldProviderRegistration calldata _yieldProviderRegistration) external pure {
    if (_yieldProviderRegistration.yieldProviderType != YieldProviderType.LIDO_STVAULT) {
      revert IncorrectYieldProviderType();
    }
  }

  function withdrawableValue() external view returns (uint256) {
    return _getYieldProviderStorage(YIELD_PROVIDER).isOssified ? YIELD_PROVIDER.balance : DASHBOARD.withdrawableValue();
  }

  function mintLST(uint256 _amount, address _recipient) external {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(YIELD_PROVIDER);
    if ($$.isOssificationInitiated || $$.isOssified) {
      revert MintLSTDisabledDuringOssification();
    }
    DASHBOARD.mintStETH(_recipient, _amount);
    $$.lstLiabilityPrincipal += _amount;
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
