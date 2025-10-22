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
   * @notice Emitted whenever LidoStVaultYieldProviderFactoryDeployed is deployed.
   * @param l1MessageService The Linea L1MessageService, also the withdrawal reserve.
   * @param yieldManager The Linea YieldManager.
   * @param vaultHub Lido VaultHub contract.
   * @param vaultFactory Lido VaultFactory contract.
   * @param steth Lido stETH contract.
   * @param gIFirstValidator Packed generalized index for the first validator before the pivot slot.
   * @param gIFirstValidatorAfterChange Packed generalized index after the pivot slot.
   * @param changeSlot Beacon chain slot at which the validator generalized index changes.
   */
  event LidoStVaultYieldProviderFactoryDeployed(
    address l1MessageService,
    address yieldManager,
    address vaultHub,
    address vaultFactory,
    address steth,
    GIndex gIFirstValidator,
    GIndex gIFirstValidatorAfterChange,
    uint64 changeSlot
  );

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

  /// @notice Lido VaultFactory contract.
  address public immutable VAULT_FACTORY;

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
  /// @param _vaultFactory Lido VaultFactory contract.
  /// @param _steth Lido stETH contract.
  /// @param _gIFirstValidator Packed generalized index for the first validator before the pivot slot.
  /// @param _gIFirstValidatorAfterChange Packed generalized index after the pivot slot.
  /// @param _changeSlot Beacon chain slot at which the validator generalized index changes.
  constructor(
    address _l1MessageService,
    address _yieldManager,
    address _vaultHub,
    address _vaultFactory,
    address _steth,
    GIndex _gIFirstValidator,
    GIndex _gIFirstValidatorAfterChange,
    uint64 _changeSlot
  ) {
    ErrorUtils.revertIfZeroAddress(_l1MessageService);
    ErrorUtils.revertIfZeroAddress(_yieldManager);
    ErrorUtils.revertIfZeroAddress(_vaultHub);
    ErrorUtils.revertIfZeroAddress(_vaultFactory);
    ErrorUtils.revertIfZeroAddress(_steth);
    L1_MESSAGE_SERVICE = _l1MessageService;
    YIELD_MANAGER = _yieldManager;
    VAULT_HUB = _vaultHub;
    VAULT_FACTORY = _vaultFactory;
    STETH = _steth;
    GI_FIRST_VALIDATOR = _gIFirstValidator;
    GI_FIRST_VALIDATOR_AFTER_CHANGE = _gIFirstValidatorAfterChange;
    CHANGE_SLOT = _changeSlot;

    emit LidoStVaultYieldProviderFactoryDeployed(
      _l1MessageService,
      _yieldManager,
      _vaultHub,
      _vaultFactory,
      _steth,
      _gIFirstValidator,
      _gIFirstValidatorAfterChange,
      _changeSlot
    );
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
        VAULT_FACTORY,
        STETH,
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT
      )
    );
    emit LidoStVaultYieldProviderCreated(yieldProviderAddress);
  }
}
