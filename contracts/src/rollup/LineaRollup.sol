// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LineaRollupBase } from "./LineaRollupBase.sol";
/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LineaRollup is LineaRollupBase {
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
   * @notice Sets forced transaction gateway and reinitializes the last finalized state including forced tx data.
   * @dev This function is a reinitializer and can only be called once per version. Should be called using an upgradeAndCall transaction to the ProxyAdmin.
   * @param _forcedTransactionGateway The address of the forced transaction gateway.
   */
  function reinitializeLineaRollupV7(address _forcedTransactionGateway) external reinitializer(7) {
    if (_forcedTransactionGateway == address(0)) {
      revert ZeroAddressNotAllowed();
    }

    nextForcedTransactionNumber = 1;

    grantRole(FORCED_TRANSACTION_SENDER_ROLE, _forcedTransactionGateway);

    /// @dev using the constants requires string memory and more complex code.
    emit LineaRollupVersionChanged(bytes8("6.0"), bytes8("7.0"));
  }
}
