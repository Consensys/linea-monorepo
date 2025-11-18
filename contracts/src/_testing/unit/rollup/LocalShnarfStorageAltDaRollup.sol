// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { ShnarfDataAcceptor } from "../../../rollup/ShnarfDataAcceptor.sol";
import { LocalShnarfProvider } from "../../../rollup/LocalShnarfProvider.sol";

/**
 * @title Contract to manage alt-DA cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LocalShnarfStorageAltDaRollup is LocalShnarfProvider, ShnarfDataAcceptor {
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
   * @notice Returns the ABI version and not the reinitialize version.
   * @return contractVersion The contract ABI version.
   */
  function CONTRACT_VERSION() external view virtual override returns (string memory contractVersion) {
    contractVersion = "1.0";
  }
}
