// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LidoStVaultYieldProvider } from "../../../yield/LidoStVaultYieldProvider.sol";
import { GIndex } from "../../../yield/libs/vendor/lido/GIndex.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestLidoStVaultYieldProvider is LidoStVaultYieldProvider {
  constructor(
    address _l1MessageService,
    address _yieldManager,
    address _vaultHub,
    address _vaultFactory,
    address _steth,
    address _validatorContainerProofVerifier
  )
    LidoStVaultYieldProvider(
      _l1MessageService,
      _yieldManager,
      _vaultHub,
      _vaultFactory,
      _steth,
      _validatorContainerProofVerifier
    )
  {}

  function getEntrypointContract(address _yieldProvider) external view returns (address entrypointContract) {
    return _getEntrypointContract(_yieldProvider);
  }

  function getDashboard(address _yieldProvider) external view returns (address dashboard) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return address(_getDashboard($$));
  }

  function getVault(address _yieldProvider) external view returns (address vault) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return address(_getVault($$));
  }

  function syncExternalLiabilitySettlement(
    address _yieldProvider,
    uint256 _liabilityShares,
    uint256 _lstLiabilityPrincipalCached
  ) external returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return _syncExternalLiabilitySettlement($$, _liabilityShares, _lstLiabilityPrincipalCached);
  }

  function payObligations(address _yieldProvider) external returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return _payObligations($$);
  }

  function payNodeOperatorFees(address _yieldProvider, uint256 _availableYield) external returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return _payNodeOperatorFees($$, _availableYield);
  }

  function payLSTPrincipalInternal(address _yieldProvider, uint256 _availableFunds) external returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return _payLSTPrincipal($$, _availableFunds);
  }

  function unstakeHarness(
    address _yieldProvider,
    bytes calldata _pubkeys,
    uint64[] calldata _amounts,
    address _refundRecipient
  ) external payable {
    _unstake(_yieldProvider, _pubkeys, _amounts, _refundRecipient);
  }

  function validateUnstakePermissionlessRequestHarness(
    address _yieldProvider,
    bytes calldata _pubkeys,
    uint64[] calldata _amounts,
    bytes calldata _withdrawalParamsProof
  ) external view returns (uint256) {
    return _validateUnstakePermissionlessRequest(_yieldProvider, _pubkeys, _amounts, _withdrawalParamsProof);
  }

  function payMaximumPossibleLSTLiability(address _yieldProvider) external returns (uint256) {
    YieldProviderStorage storage $$ = _getYieldProviderStorage(_yieldProvider);
    return _payMaximumPossibleLSTLiability($$);
  }
}
