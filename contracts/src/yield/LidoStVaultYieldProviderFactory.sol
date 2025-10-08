// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { GIndex } from "./libs/vendor/lido/GIndex.sol";
import { LidoStVaultYieldProvider } from "./LidoStVaultYieldProvider.sol";
import { ErrorUtils } from "../libraries/ErrorUtils.sol";

/**
 * @notice Deploys LidoStVaultYieldProvider contract.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProviderFactory {
  /**
   * @notice Emitted whenever a new LidoStVaultYieldProvider is deployed.
   * @param providerAddress The newly created LidoStVaultYieldProvider address.
   */
  event LidoStVaultYieldProviderCreated(address indexed providerAddress);

  /// @notice The Linea L1MessageService, also the withdrawal reserve.
  address public immutable L1_MESSAGE_SERVICE;

  /// @notice The Linea YieldManager.
  address public immutable YIELD_MANAGER;

  /// @notice Lido VaultHub contract.
  address public immutable VAULT_HUB;

  /// @notice Lido stETH contract.
  address public immutable STETH;

  /// @notice Packed generalized index for the first validator before the pivot slot.
  GIndex public immutable GI_FIRST_VALIDATOR;

  /// @notice Packed generalized index after the pivot slot.
  GIndex public immutable GI_FIRST_VALIDATOR_AFTER_CHANGE;

  /// @notice Beacon chain slot at which the validator generalized index changes.
  uint64 public immutable CHANGE_SLOT;

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
  ) {
    ErrorUtils.revertIfZeroAddress(_l1MessageService);
    ErrorUtils.revertIfZeroAddress(_yieldManager);
    ErrorUtils.revertIfZeroAddress(_vaultHub);
    ErrorUtils.revertIfZeroAddress(_steth);
    L1_MESSAGE_SERVICE = _l1MessageService;
    YIELD_MANAGER = _yieldManager;
    VAULT_HUB = _vaultHub;
    STETH = _steth;
    GI_FIRST_VALIDATOR = _gIFirstValidator;
    GI_FIRST_VALIDATOR_AFTER_CHANGE = _gIFirstValidatorAfterChange;
    CHANGE_SLOT = _changeSlot;
  }

  /**
   * @notice Creates LidoStVaultYieldProvider instance.
   * @dev LidoStVaultYieldProvider initialization is handled via permissioned YieldManager.addYieldProvider().
   * @return yieldProviderAddress The address of the deployed LidoStVaultYieldProvider beacon proxy.
   */
  function createLidoStVaultYieldProvider() external returns (address yieldProviderAddress) {
    yieldProviderAddress = address(
      new LidoStVaultYieldProvider(
        L1_MESSAGE_SERVICE,
        YIELD_MANAGER,
        VAULT_HUB,
        STETH,
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT
      )
    );
    emit LidoStVaultYieldProviderCreated(yieldProviderAddress);
  }
}
