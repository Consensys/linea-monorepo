// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LineaRollupYieldExtension } from "./LineaRollupYieldExtension.sol";
/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LineaRollup is LineaRollupYieldExtension {
  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  /**
   * @notice Initializes LineaRollup and underlying service dependencies - used for new networks only.
   * @dev DEFAULT_ADMIN_ROLE is set for the security council.
   * @dev OPERATOR_ROLE is set for operators.
   * @dev Note: This is used for new testnets and local/CI testing, and will not replace existing proxy based contracts.
   * @param _initializationData The initial data used for proof verification.
   */
  function initialize(InitializationData calldata _initializationData) external initializer {
    __LineaRollup_init(_initializationData);
  }

  /**
   * @notice Sets the roles for a list of addresses, the PauseManager pauseType:role mappings and sets the YieldManager address.
   * @dev This function is a reinitializer and can only be called once per version. Should be called using an upgradeAndCall transaction to the ProxyAdmin.
   * @param _roleAddresses The list of addresses and roles to assign permissions to.
   * @param _pauseTypeRoles The list of pause types to associate with roles.
   * @param _unpauseTypeRoles The list of unpause types to associate with roles.
   * @param _yieldManager The address of the YieldManager contract.
   */
  function reinitializeLineaRollupV7(
    RoleAddress[] calldata _roleAddresses,
    PauseTypeRole[] calldata _pauseTypeRoles,
    PauseTypeRole[] calldata _unpauseTypeRoles,
    address _yieldManager
  ) external reinitializer(7) {
    __Permissions_init(_roleAddresses);
    __PauseManager_init(_pauseTypeRoles, _unpauseTypeRoles);

    if (_yieldManager == address(0)) {
      revert ZeroAddressNotAllowed();
    }
    __LineaRollupYieldExtension_init(_yieldManager);

    /// @dev using the constants requires string memory and more complex code.
    emit LineaRollupVersionChanged(bytes8("6.0"), bytes8("7.0"));
  }
}
