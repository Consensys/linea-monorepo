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
import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { ErrorUtils } from "../libraries/ErrorUtils.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldProviderBase, CLProofVerifier, Initializable, IGenericErrors {
  /// @notice Byte-length of a validator BLS pubkey.
  uint256 private constant PUBLIC_KEY_LENGTH = 48;

  /// @notice Minimum effective balance (in gwei) for a validator with 0x02 withdrawal credentials.
  uint256 private constant MIN_0X02_VALIDATOR_ACTIVATION_BALANCE_GWEI = 32 gwei;

  /// @notice Address of the Lido VaultHub.
  IVaultHub public immutable VAULT_HUB;

  /// @notice Address of the Lido stETH contract.
  IStETH public immutable STETH;

  /// @notice Emitted when a permissionless beacon chain withdrawal is requested.
  /// @param stakingVault The staking vault address.
  /// @param refundRecipient Address designated to receive surplus withdrawal-fee refunds.
  /// @param maxUnstakeAmount Maximum ETH expected to be withdrawn for the request.
  /// @param pubkeys Concatenated validator pubkeys.
  /// @param amounts Withdrawal request amount array (currently length 1).
  event LidoVaultUnstakePermissionlessRequest(
    address indexed stakingVault,
    address indexed refundRecipient,
    uint256 maxUnstakeAmount,
    bytes pubkeys,
    uint64[] amounts
  );

  /// @notice Used to set immutable variables, but not storage.
  /// @param _l1MessageService The Linea L1MessageService, also the withdrawal reserve.
  /// @param _yieldManager The Linea YieldManager.
  /// @param _vaultHub Lido VaultHub contract.
  /// @param _steth Lido stETH contract.
  /// @param _gIFirstValidator Packed generalized index for the first validator before the pivot slot.
  /// @param _gIFirstValidatorAfterChange Packed generalized index after the pivot slot.
  /// @param _changeSlot Beacon chain slot at which the validator generalized index changes.
  constructor(
    address _l1MessageService,
    address _yieldManager,
    address _vaultHub,
    address _steth,
    GIndex _gIFirstValidator,
    GIndex _gIFirstValidatorAfterChange,
    uint64 _changeSlot
  )
    YieldProviderBase(_l1MessageService, _yieldManager)
    CLProofVerifier(_gIFirstValidator, _gIFirstValidatorAfterChange, _changeSlot)
  {
    ErrorUtils.revertIfZeroAddress(_l1MessageService);
    ErrorUtils.revertIfZeroAddress(_yieldManager);
    ErrorUtils.revertIfZeroAddress(_vaultHub);
    ErrorUtils.revertIfZeroAddress(_steth);
    VAULT_HUB = IVaultHub(_vaultHub);
    STETH = IStETH(_steth);
    _disableInitializers();
  }

  /**
   * @dev Storage is expected to be initialized via the permissioned YieldManager.addYieldProvider function
   */
  function initialize() external initializer {}

  /// @notice Helper function to get the Lido contract to interact with.
  function _getEntrypointContract(address _yieldProvider) internal view returns (address entrypointContract) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    entrypointContract = $$.isOssified ? $$.ossifiedEntrypoint : $$.primaryEntrypoint;
  }

  /// @notice Helper function to get the associated Lido Dashboard contract.
  function _getDashboard(YieldProviderStorage storage $$) internal view returns (IDashboard dashboard) {
    dashboard = IDashboard($$.primaryEntrypoint);
  }

  /// @notice Helper function to get the associated Lido StakingVault contract.
  function _getVault(YieldProviderStorage storage $$) internal view returns (IStakingVault vault) {
    vault = IStakingVault($$.ossifiedEntrypoint);
  }

  /**
   * @notice Returns the amount of ETH the provider can immediately remit back to the YieldManager.
   * @dev Called via `delegatecall` from the YieldManager.
   * @param _yieldProvider The yield provider address.
   * @return availableBalance The ETH amount that can be withdrawn.
   */
  function withdrawableValue(address _yieldProvider) external view onlyDelegateCall returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return
      $$.isOssified
        ? IStakingVault($$.ossifiedEntrypoint).availableBalance()
        : IDashboard($$.primaryEntrypoint).withdrawableValue();
  }

  /**
   * @notice Forwards ETH from the YieldManager to the yield provider.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of ETH supplied by the YieldManager.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).fund{ value: _amount }();
  }

  /**
   * @notice Computes and returns earned yield that can be distributed to L2 users.
   * @dev Gross net yield is the difference between recorded user funds dedicated to the YieldProvider,
   *      and current total ETH value of the YieldProvider.
   * @dev Before reporting yield as available for distribution, will first settle the following from earned yield:
   *      - Incurred negative yield
   *      - LST liability
   *      - Obligations (i.e. Lido protocol fees)
   *      - Node operator fees
   * @param _yieldProvider The yield provider address.
   * @return newReportedYield New net yield (denominated in ETH) since the prior report.
   * @return outstandingNegativeYield Amount of outstanding negative yield.
   */
  function reportYield(
    address _yieldProvider
  ) external onlyDelegateCall returns (uint256 newReportedYield, uint256 outstandingNegativeYield) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssified) {
      revert OperationNotSupportedDuringOssification(OperationType.ReportYield);
    }
    // First compute the total yield
    uint256 lastUserFunds = $$.userFunds;
    uint256 totalVaultFunds = _getDashboard($$).totalValue();
    // Gross positive yield
    if (totalVaultFunds > lastUserFunds) {
      newReportedYield = totalVaultFunds - lastUserFunds;
      // 1. Pay liabilities
      uint256 lstLiabilityPayment = _payMaximumPossibleLSTLiability($$);
      newReportedYield = Math256.safeSub(newReportedYield, lstLiabilityPayment);
      // 2. Pay obligations
      uint256 obligationsPaid = _payObligations($$, newReportedYield);
      newReportedYield = Math256.safeSub(newReportedYield, obligationsPaid);
      // 3. Pay node operator fees
      uint256 nodeOperatorFeesPaid = _payNodeOperatorFees($$, newReportedYield);
      newReportedYield = Math256.safeSub(newReportedYield, nodeOperatorFeesPaid);

      outstandingNegativeYield = Math256.safeSub(
        lstLiabilityPayment + obligationsPaid + nodeOperatorFeesPaid,
        totalVaultFunds - lastUserFunds
      );
      // Gross negative yield
    } else {
      newReportedYield = 0;
      outstandingNegativeYield = lastUserFunds - totalVaultFunds;
    }
  }

  /**
   * @notice Helper function to pay the maximum possible outstanding LST liability.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @dev Call _syncExternalLiabilitySettlement() after LST liability payment is done.
   * @return liabilityPaidETH Amount of ETH used to pay liabilities.
   */
  function _payMaximumPossibleLSTLiability(
    YieldProviderStorage storage $$
  ) internal returns (uint256 liabilityPaidETH) {
    if ($$.isOssified) return 0;
    IDashboard dashboard = IDashboard($$.primaryEntrypoint);
    address vault = $$.ossifiedEntrypoint;
    uint256 availableVaultBalanceBeforeRebalance = IStakingVault(vault).availableBalance();
    uint256 availableRebalanceShares = STETH.getSharesByPooledEth(availableVaultBalanceBeforeRebalance);
    uint256 liabilityShares = dashboard.liabilityShares();
    uint256 rebalanceShares = Math256.min(liabilityShares, availableRebalanceShares);
    if (rebalanceShares > 0) {
      // Cheaper lookup for before-after compare than availableBalance()
      uint256 vaultBalanceBeforeRebalance = vault.balance;
      dashboard.rebalanceVaultWithShares(rebalanceShares);
      // Apply consistent accounting treatment that LST interest paid first, then LST principal
      _syncExternalLiabilitySettlement($$, dashboard.liabilityShares(), $$.lstLiabilityPrincipal);
      liabilityPaidETH = vaultBalanceBeforeRebalance - vault.balance;
    }
  }

  /**
   * @notice Helper function to handle liability settlement executed by external actors (i.e. via permissionless VaultHub.settleVaultObligations)
   * @dev Must be called before any function that LST liability by a specified amount. Otherwise we encounter the edge case
   *      that externally settled liabilities will block these operations and hence withdrawal functions that eagerly execute LST principal payment.
   * @dev Greedily assumes that externally settled liability was first allocated to lstLiabilityInterest, then the remainder to lstLiabilityPrincipal
   * @dev User funds removed from the Vault due to external actor settlement, are recognised as negative yield.
   *      See tech spec for more details.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @param _liabilityShares Current outstanding liabilityShares.
   * @param _lstLiabilityPrincipalCached Recorded LST liability principal.
   * @return lstLiabilityPrincipalSynced New LST liability principal.
   * @return isLstLiabilityPrincipalChanged True if LST liability principal was updated.
   */
  function _syncExternalLiabilitySettlement(
    YieldProviderStorage storage $$,
    uint256 _liabilityShares,
    uint256 _lstLiabilityPrincipalCached
  ) internal returns (uint256 lstLiabilityPrincipalSynced, bool isLstLiabilityPrincipalChanged) {
    uint256 liabilityETH = STETH.getPooledEthBySharesRoundUp(_liabilityShares);
    // If true, this means an external actor settled liabilities.
    if (liabilityETH < _lstLiabilityPrincipalCached) {
      uint256 lstLiabilityPrincipalDecrement = _lstLiabilityPrincipalCached - liabilityETH;
      // Any decrement in lstLiabilityPrincipal must be 1:1 matched with decrements in userFunds and _userFundsInYieldProvidersTotal.
      _getYieldManagerStorage().userFundsInYieldProvidersTotal = Math256.safeSub(
        _getYieldManagerStorage().userFundsInYieldProvidersTotal,
        lstLiabilityPrincipalDecrement
      );
      $$.userFunds = Math256.safeSub($$.userFunds, lstLiabilityPrincipalDecrement);
      $$.lstLiabilityPrincipal = liabilityETH;
      return (liabilityETH, true);
    } else {
      return (_lstLiabilityPrincipalCached, false);
    }
  }

  /**
   * @notice Helper function to pay obligations.
   * @dev In Lido's system, a subfield of obligations is 'redemptions'. Redemptions are decremented 1:1 with liabilities.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @param _availableYield Amount of yield available.
   * @return obligationsPaid Amount of ETH used to pay obligations.
   */
  function _payObligations(
    YieldProviderStorage storage $$,
    uint256 _availableYield
  ) internal returns (uint256 obligationsPaid) {
    if (_availableYield == 0) return 0;
    address vault = $$.ossifiedEntrypoint;
    uint256 beforeVaultBalance = vault.balance;
    // Unfortunately, there is no function on VaultHub to specify how much obligation we want to repay.
    VAULT_HUB.settleVaultObligations(vault);
    uint256 afterVaultBalance = vault.balance;
    obligationsPaid = beforeVaultBalance - afterVaultBalance;
  }

  /**
   * @notice Helper function to pay node operator fees.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @param _availableYield Amount of yield available.
   * @return nodeOperatorFeesPaid Amount of ETH used to pay node operator fees.
   */
  function _payNodeOperatorFees(
    YieldProviderStorage storage $$,
    uint256 _availableYield
  ) internal returns (uint256 nodeOperatorFeesPaid) {
    if (_availableYield == 0) return 0;
    IDashboard dashboard = _getDashboard($$);
    IStakingVault vault = _getVault($$);
    uint256 currentFees = dashboard.nodeOperatorDisbursableFee();
    uint256 avalableVaultBalance = vault.availableBalance();
    // Does not allow partial payment of node operator fees, unlike settleVaultObligations
    if (avalableVaultBalance > currentFees) {
      dashboard.disburseNodeOperatorFee();
      nodeOperatorFeesPaid = currentFees;
    }
  }

  /**
   * @notice Reduces the outstanding LST liability principal.
   * @dev Called after the YieldManager has reserved `_availableFunds` for liability
   *      settlement.
   *      - Implementations should update `lstLiabilityPrincipal` in the YieldProvider storage
   *      - Implementations should ensure lstPrincipalPaid <= _availableFunds
   * @param _yieldProvider The yield provider address.
   * @param _availableFunds The maximum amount of ETH that is available to pay LST liability principal.
   * @return lstPrincipalPaid The actual ETH amount paid to reduce LST liability principal.
   */
  function payLSTPrincipal(
    address _yieldProvider,
    uint256 _availableFunds
  ) external onlyDelegateCall returns (uint256 lstPrincipalPaid) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssificationInitiated || $$.isOssified) {
      return 0;
    }
    lstPrincipalPaid = _payLSTPrincipal($$, _availableFunds);
  }

  /**
   * @notice Helper function to pay LST liability principal.
   * @dev Calls _syncExternalLiabilitySettlement before computing the liability payment.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @param _availableFunds The maximum amount of ETH that is available to pay LST liability principal.
   * @return lstPrincipalPaid The actual ETH amount paid to reduce LST liability principal.
   */
  function _payLSTPrincipal(
    YieldProviderStorage storage $$,
    uint256 _availableFunds
  ) internal returns (uint256 lstPrincipalPaid) {
    uint256 lstLiabilityPrincipalCached = $$.lstLiabilityPrincipal;
    if (lstLiabilityPrincipalCached == 0) return 0;
    IDashboard dashboard = _getDashboard($$);
    (uint256 lstLiabilityPrincipalSynced, ) = _syncExternalLiabilitySettlement(
      $$,
      dashboard.liabilityShares(),
      lstLiabilityPrincipalCached
    );
    lstPrincipalPaid = Math256.min(lstLiabilityPrincipalSynced, _availableFunds);
    if (lstPrincipalPaid > 0) {
      dashboard.rebalanceVaultWithEther(lstPrincipalPaid);
      $$.lstLiabilityPrincipal -= lstPrincipalPaid;
    }
  }

  /**
   * @notice Requests beacon chain withdrawal via EIP-7002 withdrawal contract.
   * @dev Parameters are ABI encoded by the YieldManager and understood by the yield provider.
   * @dev Dynamic withdrawal fee is sourced from `msg.value`
   * @param _yieldProvider The yield provider address.
   * @param _withdrawalParams Provider-specific payload describing the withdrawals to trigger.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable onlyDelegateCall {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(
      _withdrawalParams,
      (bytes, uint64[], address)
    );
    _unstake(_yieldProvider, pubkeys, amounts, refundRecipient);
    // Intentional choice to not emit event as downstream StakingVault will emit ValidatorWithdrawalsTriggered event.
  }

  /**
   * @notice Helper function to execute EIP-7002 withdrawal requests via Lido contracts.
   * @param _yieldProvider The yield provider address.
   * @param _pubkeys Concatenated validator public keys (48 bytes each).
   * @param _amounts Withdrawal amounts in gwei for each validator key and must match _pubkeys length.
   *         Set amount to 0 for a full validator exit.
   *         For partial withdrawals, amounts will be trimmed to keep MIN_ACTIVATION_BALANCE on the validator to avoid deactivation
   * @param _refundRecipient Address to receive any fee refunds, if zero, refunds go to msg.sender.
   */
  function _unstake(
    address _yieldProvider,
    bytes memory _pubkeys,
    uint64[] memory _amounts,
    address _refundRecipient
  ) internal {
    // Lido StakingVault.sol will handle the param validation
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).triggerValidatorWithdrawals{ value: msg.value }(
      _pubkeys,
      _amounts,
      _refundRecipient
    );
  }

  /**
   * @notice Permissionlessly requests beacon chain withdrawal via EIP-7002 withdrawal contract when reserve is under minimum threshold.
   * @dev Implementations must verify the calldata proof (for example against EIP-4788 beacon roots)
   *      and enforce any provider-specific safety checks. The returned amount is used by the
   *      YieldManager to cap pending withdrawals tracked on L1.
   * @param _yieldProvider The yield provider address.
   * @param _withdrawalParams ABI encoded provider parameters.
   * @param _withdrawalParamsProof Proof data (typically a beacon chain Merkle proof).
   * @return maxUnstakeAmount Maximum ETH amount expected to be withdrawn as a result of this request.
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
      refundRecipient,
      maxUnstakeAmount,
      pubkeys,
      amounts
    );
  }

  /**
   * @notice Helper function to validate permissionless unstake requests.
   * @dev Validates that the request is for a partial withdrawal from a single validator.
   * @dev Checks guided by consensus specs https://github.com/ethereum/consensus-specs/blob/834e40604ae4411e565bd6540da50b008b2496dc/specs/electra/beacon-chain.md#new-process_withdrawal_request
   * @param _yieldProvider The yield provider address.
   * @param _pubkeys Concatenated validator public keys (48 bytes each).
   * @param _amounts Withdrawal amounts in gwei for each validator key and must match _pubkeys length.
   *         Set amount to 0 for a full validator exit.
   *         For partial withdrawals, amounts will be trimmed to keep MIN_ACTIVATION_BALANCE on the validator to avoid deactivation
   * @param _withdrawalParamsProof Proof data containing a beacon chain Merkle proof against the EIP-4788 beacon chain root.
   * @return maxUnstakeAmount Maximum ETH amount expected to be withdrawn as a result of this request.
   */
  function _validateUnstakePermissionless(
    address _yieldProvider,
    bytes memory _pubkeys,
    uint64[] memory _amounts,
    bytes calldata _withdrawalParamsProof
  ) internal view returns (uint256 maxUnstakeAmount) {
    // Length validator
    if (_pubkeys.length != PUBLIC_KEY_LENGTH || _amounts.length != 1) {
      revert SingleValidatorOnlyForUnstakePermissionless();
    }

    uint256 amount = _amounts[0];
    if (amount == 0) {
      revert NoValidatorExitForUnstakePermissionless();
    }

    ValidatorWitness memory witness = abi.decode(_withdrawalParamsProof, (ValidatorWitness));

    // 0x02 withdrawal credential scheme
    address vault = _getYieldProviderStorage(_yieldProvider).ossifiedEntrypoint;
    bytes32 withdrawalCredentials;
    assembly {
      withdrawalCredentials := or(shl(248, 0x2), vault)
    }

    _validateValidatorContainerForPermissionlessUnstake(witness, withdrawalCredentials);

    // https://github.com/ethereum/consensus-specs/blob/master/specs/electra/beacon-chain.md#modified-get_expected_withdrawals
    uint256 maxUnstakeAmountGwei = Math256.min(
      amount,
      Math256.safeSub(witness.effectiveBalance, MIN_0X02_VALIDATOR_ACTIVATION_BALANCE_GWEI)
    );
    // Convert from Beacon Chain units of 'gwei' to execution layer units of 'wei'
    maxUnstakeAmount = maxUnstakeAmountGwei * 1 gwei;
  }

  /**
   * @notice Withdraws ETH from the provider back into the YieldManager.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of ETH to withdraw to the YieldManager.
   */
  function withdrawFromYieldProvider(address _yieldProvider, uint256 _amount) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).withdraw(address(this), _amount);
  }

  /**
   * @notice Pauses new beacon chain deposits.
   * @param _yieldProvider The yield provider address.
   */
  function pauseStaking(address _yieldProvider) external onlyDelegateCall {
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).pauseBeaconChainDeposits();
  }

  /**
   * @notice Resumes beacon chain deposits for the provider after a pause.
   * @param _yieldProvider The yield provider address.
   * @dev Whether to allow staking during ossification is a vendor-specific detail.
   */
  function unpauseStaking(address _yieldProvider) external onlyDelegateCall {
    if (_getYieldProviderStorage(_yieldProvider).isOssified) revert UnpauseStakingForbiddenWhenOssified();
    ICommonVaultOperations(_getEntrypointContract(_yieldProvider)).resumeBeaconChainDeposits();
  }

  /**
   * @notice Withdraws liquid staking tokens (LST) to a recipient.
   * @dev Implementations must `lstLiabilityPrincipal` state for the yield provider.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of LST (denominated in ETH) to withdraw.
   * @param _recipient Address receiving the LST.
   */
  function withdrawLST(address _yieldProvider, uint256 _amount, address _recipient) external onlyDelegateCall {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if ($$.isOssificationInitiated || $$.isOssified) {
      revert MintLSTDisabledDuringOssification();
    }
    IDashboard($$.primaryEntrypoint).mintStETH(_recipient, _amount);
    $$.lstLiabilityPrincipal += _amount;
  }

  /**
   * @notice Begins the provider-specific ossification workflow.
   * @param _yieldProvider The yield provider address.
   */
  function initiateOssification(address _yieldProvider) external onlyDelegateCall {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    _payMaximumPossibleLSTLiability($$);
    // Lido implementation handles Lido fee payment, and revert on fresh report
    // This will fail if any existing liabilities or obligations
    IDashboard(_getYieldProviderStorage(_yieldProvider).primaryEntrypoint).voluntaryDisconnect();
  }

  /**
   * @notice Reverts a previously initiated ossification request.
   * @dev In Lido's case, this is only available for a limited time after initiateOssification.
   *      If there is subsequent accounting report is applied and no liabilities or obligations
   *      are outstanding, the vault will be disconnected which will be
   * @param _yieldProvider The yield provider address.
   */
  function undoInitiateOssification(address _yieldProvider) external onlyDelegateCall {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    if (!VAULT_HUB.isVaultConnected(address(_getVault($$)))) {
      _getDashboard($$).reconnectToVaultHub();
    }
  }

  /**
   * @notice Process a previously initiated ossification process.
   * @param _yieldProvider The yield provider address.
   * @return isOssificationComplete True if the provider is now in the ossified state.
   */
  function processPendingOssification(
    address _yieldProvider
  ) external onlyDelegateCall returns (bool isOssificationComplete) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    // Give ownership to YieldManager
    IDashboard($$.primaryEntrypoint).abandonDashboard(address(this));
    IStakingVault vault = IStakingVault($$.ossifiedEntrypoint);
    vault.acceptOwnership();
    vault.ossify();
    isOssificationComplete = true;
  }

  /**
   * @notice Performs vendor-specific validation before the provider is registered by the YieldManager.
   * @param _registration Registration payload for the yield provider.
   */
  function validateAdditionToYieldManager(YieldProviderRegistration calldata _registration) external view {
    if (_registration.yieldProviderVendor != YieldProviderVendor.LIDO_STVAULT) {
      revert UnknownYieldProviderVendor();
    }
    IDashboard dashboard = IDashboard(_registration.primaryEntrypoint);
    address expectedVault = address(dashboard.stakingVault());
    if (expectedVault != _registration.ossifiedEntrypoint) {
      revert InvalidYieldProviderRegistration(YieldProviderRegistrationError.LidoDashboardNotLinkedToVault);
    }
    if (_registration.receiveCaller != _registration.ossifiedEntrypoint) {
      revert InvalidYieldProviderRegistration(
        YieldProviderRegistrationError.LidoVaultIsExpectedReceiveCallerAndOssifiedEntrypoint
      );
    }
  }
}
