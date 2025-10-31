// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { YieldProviderBase } from "./YieldProviderBase.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { ICommonVaultOperations } from "./interfaces/vendor/lido/ICommonVaultOperations.sol";
import { IDashboard } from "./interfaces/vendor/lido/IDashboard.sol";
import { IStETH } from "./interfaces/vendor/lido/IStETH.sol";
import { IVaultHub } from "./interfaces/vendor/lido/IVaultHub.sol";
import { IVaultFactory } from "./interfaces/vendor/lido/IVaultFactory.sol";
import { IStakingVault } from "./interfaces/vendor/lido/IStakingVault.sol";
import { Math256 } from "../lib/Math256.sol";
import { ErrorUtils } from "../lib/ErrorUtils.sol";
import { IPermissionsManager } from "../interfaces/IPermissionsManager.sol";
import { ProgressOssificationResult, YieldProviderRegistration, YieldProviderVendor } from "./interfaces/YieldTypes.sol";
import { IValidatorContainerProofVerifier } from "./interfaces/IValidatorContainerProofVerifier.sol";

/**
 * @title Contract to handle native yield operations with Lido Staking Vault.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProvider is YieldProviderBase, IGenericErrors {
  /// @notice Byte-length of a validator BLS pubkey.
  uint256 private constant PUBLIC_KEY_LENGTH = 48;

  /// @notice Minimum effective balance (in gwei) for a validator with 0x02 withdrawal credentials.
  uint256 private constant MIN_0X02_VALIDATOR_ACTIVATION_BALANCE_GWEI = 32 gwei;

  /// @notice Address of the Lido VaultHub.
  IVaultHub public immutable VAULT_HUB;

  /// @notice Address of the Lido VaultFactory.
  IVaultFactory public immutable VAULT_FACTORY;

  /// @notice Address of the Lido stETH contract.
  IStETH public immutable STETH;

  /// @notice Linea ValidatorContainerProofVerifier contract.
  IValidatorContainerProofVerifier public immutable VALIDATOR_CONTAINER_PROOF_VERIFIER;

  /// @notice amount of ETH that is locked on the vault on connect and can be withdrawn on disconnect only
  uint256 public constant CONNECT_DEPOSIT = 1 ether;

  /**
   * @notice Emitted whenever LidoStVaultYieldProvider is deployed.
   * @param l1MessageService The Linea L1MessageService, also the withdrawal reserve holding contract.
   * @param yieldManager The Linea YieldManager.
   * @param vaultHub Lido VaultHub contract.
   * @param vaultFactory Lido VaultFactory contract.
   * @param steth Lido stETH contract.
   * @param validatorContainerProofVerifier Linea ValidatorContainerProofVerifier contract.
   */
  event LidoStVaultYieldProviderDeployed(
    address l1MessageService,
    address yieldManager,
    address vaultHub,
    address vaultFactory,
    address steth,
    address validatorContainerProofVerifier
  );

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
  /// @param _l1MessageService The Linea L1MessageService, also the withdrawal reserve holding contract.
  /// @param _yieldManager The Linea YieldManager.
  /// @param _vaultHub Lido VaultHub contract.
  /// @param _vaultFactory Lido VaultFactory contract.
  /// @param _steth Lido stETH contract.
  /// @param _validatorContainerProofVerifier Linea ValidatorContainerProofVerifier contract.
  constructor(
    address _l1MessageService,
    address _yieldManager,
    address _vaultHub,
    address _vaultFactory,
    address _steth,
    address _validatorContainerProofVerifier
  ) YieldProviderBase(_l1MessageService, _yieldManager) {
    ErrorUtils.revertIfZeroAddress(_l1MessageService);
    ErrorUtils.revertIfZeroAddress(_yieldManager);
    ErrorUtils.revertIfZeroAddress(_vaultHub);
    ErrorUtils.revertIfZeroAddress(_vaultFactory);
    ErrorUtils.revertIfZeroAddress(_steth);
    ErrorUtils.revertIfZeroAddress(_validatorContainerProofVerifier);
    VAULT_HUB = IVaultHub(_vaultHub);
    VAULT_FACTORY = IVaultFactory(_vaultFactory);
    STETH = IStETH(_steth);
    VALIDATOR_CONTAINER_PROOF_VERIFIER = IValidatorContainerProofVerifier(_validatorContainerProofVerifier);

    emit LidoStVaultYieldProviderDeployed(
      _l1MessageService,
      _yieldManager,
      _vaultHub,
      _vaultFactory,
      _steth,
      _validatorContainerProofVerifier
    );
  }

  /**
   * @notice Helper function to get the Lido contract to interact with.
   * @dev If the vault has been ossified, the underlying contract switches.
   * @param _yieldProvider The yield provider contract address.
   * @return entrypointContract The Lido contract to interact with.
   */
  function _getEntrypointContract(address _yieldProvider) internal view returns (address entrypointContract) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    entrypointContract = $$.isOssified ? $$.ossifiedEntrypoint : $$.primaryEntrypoint;
  }

  /**
   * @notice Helper function to get the associated Lido Dashboard contract.
   * @param $$ The yield provider storage pointer.
   * @return dashboard The dashboard contract.
   */
  function _getDashboard(YieldProviderStorage storage $$) internal view returns (IDashboard dashboard) {
    dashboard = IDashboard($$.primaryEntrypoint);
  }

  /**
   * @notice Helper function to get the associated Lido StakingVault contract.
   * @param $$ The yield provider storage pointer.
   * @return vault The StakingVault contract.
   */
  function _getVault(YieldProviderStorage storage $$) internal view returns (IStakingVault vault) {
    vault = IStakingVault($$.ossifiedEntrypoint);
  }

  /**
   * @notice Returns the amount of ETH the provider can immediately remit back to the YieldManager.
   * @dev Called via `delegatecall` from the YieldManager.
   * @param _yieldProvider The yield provider address.
   * @return availableBalance The ETH amount that can be withdrawn.
   */
  function withdrawableValue(address _yieldProvider) external view onlyDelegateCall returns (uint256 availableBalance) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    availableBalance = $$.isOssified
      ? IStakingVault($$.ossifiedEntrypoint).availableBalance()
      : IDashboard($$.primaryEntrypoint).withdrawableValue();
  }

  /**
   * @notice Forwards ETH from the YieldManager to the yield provider.
   * @param _yieldProvider The yield provider address.
   * @param _amount Amount of ETH supplied by the YieldManager.
   */
  function fundYieldProvider(address _yieldProvider, uint256 _amount) external onlyDelegateCall {
    // Ossified -> Vault cannot generate yield -> Should fully withdraw
    if (_getYieldProviderStorage(_yieldProvider).isOssified)
      revert OperationNotSupportedDuringOssification(OperationType.FundYieldProvider);
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
    if ($$.isOssified) revert OperationNotSupportedDuringOssification(OperationType.ReportYield);
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
      uint256 obligationsPaid = _payObligations($$);
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
    uint256 rebalanceShares = Math256.min(
      dashboard.liabilityShares(),
      STETH.getSharesByPooledEth(IStakingVault(vault).availableBalance())
    );
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
   * @notice Helper function to handle liability settlement executed by external actors (i.e. via permissionless VaultHub.settleLidoFees)
   * @dev Must be called before any function that LST liability by a specified amount. Otherwise we encounter the edge case
   *      that externally settled liabilities will block these operations and hence withdrawal functions that eagerly execute LST principal payment.
   * @dev Greedily assumes that externally settled liability was first allocated to lstLiabilityInterest, then the remainder to lstLiabilityPrincipal
   * @dev User funds removed from the Vault due to external actor settlement, are recognised as negative yield.
   *      See tech spec for more details.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @param _liabilityShares Current outstanding liabilityShares.
   * @param _lstLiabilityPrincipalCached Recorded LST liability principal.
   * @return lstLiabilityPrincipalSynced New LST liability principal.
   */
  function _syncExternalLiabilitySettlement(
    YieldProviderStorage storage $$,
    uint256 _liabilityShares,
    uint256 _lstLiabilityPrincipalCached
  ) internal returns (uint256 lstLiabilityPrincipalSynced) {
    uint256 liabilityETH = STETH.getPooledEthBySharesRoundUp(_liabilityShares);
    // If true, this means an external actor settled liabilities.
    if (liabilityETH < _lstLiabilityPrincipalCached) {
      $$.lstLiabilityPrincipal = liabilityETH;
      return liabilityETH;
    } else {
      return _lstLiabilityPrincipalCached;
    }
  }

  /**
   * @notice Helper function to pay obligations.
   * @dev Greedily pay Lido fees. `settleLidoFees` is permissionless, and is better settled eagerly.
   * @dev There are multiple revert conditions on the Lido implementation.
   * @dev Settling failures shouldn't block other flows in this scenario.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   * @return obligationsPaid Amount of ETH used to pay obligations.
   */
  function _payObligations(YieldProviderStorage storage $$) internal returns (uint256 obligationsPaid) {
    address vault = $$.ossifiedEntrypoint;
    uint256 beforeVaultBalance = vault.balance;

    try VAULT_HUB.settleLidoFees(vault) {
      obligationsPaid = beforeVaultBalance - vault.balance;
    } catch {
      return 0;
    }
  }

  /**
   * @notice Helper function to pay node operator fees.
   * @dev Does not allow partial payment of node operator fees, unlike settleLidoFees.
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
    uint256 currentFees = dashboard.accruedFee();
    uint256 availableVaultBalance = vault.availableBalance();
    // Does not allow partial payment of node operator fees, unlike settleLidoFees
    if (availableVaultBalance > currentFees) {
      // External call to dashboard may revert for reasons beyond YieldManager control
      try dashboard.disburseFee() {
        nodeOperatorFeesPaid = currentFees;
      } catch {
        return 0;
      }
    }
  }

  /**
   * @notice Reduces the outstanding LST liability principal.
   * @dev Called after the YieldManager has reserved `_availableFunds` for liability settlement.
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
    uint256 lstLiabilityPrincipalSynced = _syncExternalLiabilitySettlement(
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
   * @dev Dynamic withdrawal fee is sourced from `msg.value`.
   * @param _yieldProvider The yield provider address.
   * @param _withdrawalParams Provider-specific payload describing the withdrawals to trigger.
   */
  function unstake(address _yieldProvider, bytes memory _withdrawalParams) external payable onlyDelegateCall {
    (bytes memory pubkeys, uint64[] memory amounts, address refundRecipient) = abi.decode(
      _withdrawalParams,
      (bytes, uint64[], address)
    );
    _unstake(_yieldProvider, pubkeys, amounts, refundRecipient);
    /// @dev Intentional choice to not emit event as downstream StakingVault will emit ValidatorWithdrawalsTriggered event.
  }

  /**
   * @notice Helper function to execute EIP-7002 withdrawal requests via Lido contracts.
   * @param _yieldProvider The yield provider address.
   * @param _pubkeys Concatenated validator public keys (48 bytes each).
   * @param _amounts Withdrawal amounts in gwei for each validator key and must match _pubkeys length.
   *         Set amount to 0 for a full validator exit.
   *         For partial withdrawals, amounts will be trimmed to keep MIN_ACTIVATION_BALANCE on the validator to avoid deactivation.
   * @param _refundRecipient Address to receive any fee refunds, if zero, refunds go to msg.sender.
   */
  function _unstake(
    address _yieldProvider,
    bytes memory _pubkeys,
    uint64[] memory _amounts,
    address _refundRecipient
  ) internal {
    /// @dev Lido StakingVault.sol will handle the param validation
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
    maxUnstakeAmount = _validateUnstakePermissionlessRequest(_yieldProvider, pubkeys, amounts, _withdrawalParamsProof);
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
   *         For partial withdrawals, amounts will be trimmed to keep MIN_ACTIVATION_BALANCE on the validator to avoid deactivation.
   * @param _withdrawalParamsProof Proof data containing a beacon chain Merkle proof against the EIP-4788 beacon chain root.
   * @return maxUnstakeAmount Maximum ETH amount expected to be withdrawn as a result of this request.
   */
  function _validateUnstakePermissionlessRequest(
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

    IValidatorContainerProofVerifier.ValidatorContainerWitness memory witness = abi.decode(
      _withdrawalParamsProof,
      (IValidatorContainerProofVerifier.ValidatorContainerWitness)
    );

    // 0x02 withdrawal credential scheme
    address vault = _getYieldProviderStorage(_yieldProvider).ossifiedEntrypoint;
    bytes32 withdrawalCredentials;
    assembly {
      withdrawalCredentials := or(shl(248, 0x2), vault)
    }

    VALIDATOR_CONTAINER_PROOF_VERIFIER.verifyActiveValidatorContainer(witness, _pubkeys, withdrawalCredentials);

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
   * @dev Caller emits minting event.
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
   * @dev WARNING: This operation irreversibly pauses beacon chain deposits and LST withdrawals.
   */
  function initiateOssification(address _yieldProvider) external onlyDelegateCall {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    _initiateOssification($$);
  }

  /**
   * @notice Internal helper function to initiate ossification.
   * @param $$ Storage pointer for the YieldProvider-scoped storage.
   */
  function _initiateOssification(YieldProviderStorage storage $$) internal {
    _payMaximumPossibleLSTLiability($$);
    // Attempt to disconnect the YieldProvider gracefully.
    // The try/catch block prevents reverts from external factors such as:
    //   i.) Missing or stale reports
    //   ii.) Outstanding liabilities
    //   iii.) Unsettled obligations
    // This function is intended to be executed by the Security Council and may take several days to prepare.
    // Regardless of the post-disconnect state, the automation service running `preparePendingOssification`
    // will progress the ossification process.
    try _getDashboard($$).voluntaryDisconnect() {} catch {}
  }

  /**
   * @notice Process a previously initiated ossification process.
   * @param _yieldProvider The yield provider address.
   * @return progressOssificationResult The operation result.
   */
  function progressPendingOssification(
    address _yieldProvider
  ) external onlyDelegateCall returns (ProgressOssificationResult progressOssificationResult) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    IStakingVault vault = _getVault($$);
    if (!VAULT_HUB.isVaultConnected(address(vault))) {
      // Disconnected, can complete ossification
      // Transfer StakingVault ownership to YieldManager
      IDashboard($$.primaryEntrypoint).abandonDashboard(address(this));
      vault.acceptOwnership();
      // Ossify
      vault.ossify();
      // Unstage all ETH
      vault.setDepositor(address(this));
      vault.unstage(vault.stagedBalance());
      progressOssificationResult = ProgressOssificationResult.COMPLETE;
    } else if (VAULT_HUB.isPendingDisconnect(address(vault))) {
      // No-op, needs accounting report to progress.
      progressOssificationResult = ProgressOssificationResult.NOOP;
    } else {
      // Previous disconnect attempt has aborted, must re-execute.
      _initiateOssification($$);
      progressOssificationResult = ProgressOssificationResult.REINITIATED;
    }
  }

  /**
   * @notice Performs vendor-specific initialization logic.
   * @param _vendorInitializationData Vendor-specific initialization data.
   * @return registrationData Data required to register a new YieldProvider with the YieldManager.
   */
  function initializeVendorContracts(
    bytes memory _vendorInitializationData
  ) external onlyDelegateCall returns (YieldProviderRegistration memory registrationData) {
    (
      address defaultAdmin,
      address nodeOperator,
      address nodeOperatorManager,
      uint256 nodeOperatorFeeBP,
      uint256 confirmExpiry,
      IPermissionsManager.RoleAddress[] memory roleAssignments
    ) = abi.decode(
        _vendorInitializationData,
        (address, address, address, uint256, uint256, IPermissionsManager.RoleAddress[])
      );

    (address vault, address dashboard) = VAULT_FACTORY.createVaultWithDashboard{ value: CONNECT_DEPOSIT }(
      defaultAdmin,
      nodeOperator,
      nodeOperatorManager,
      nodeOperatorFeeBP,
      confirmExpiry,
      roleAssignments
    );

    registrationData = YieldProviderRegistration({
      yieldProviderVendor: YieldProviderVendor.LIDO_STVAULT,
      primaryEntrypoint: dashboard,
      ossifiedEntrypoint: vault
    });
  }

  /**
   * @notice Performs vendor-specific exit logic.
   * @param _vendorExitData Vendor-specific exit data.
   */
  function exitVendorContracts(address _yieldProvider, bytes memory _vendorExitData) external onlyDelegateCall {
    if (_vendorExitData.length == 0) return;
    address newVaultOwner = abi.decode(_vendorExitData, (address));
    ErrorUtils.revertIfZeroAddress(newVaultOwner);
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    $$.isOssified
      ? _getVault($$).transferOwnership(newVaultOwner)
      : _getDashboard($$).transferVaultOwnership(newVaultOwner);
  }
}
