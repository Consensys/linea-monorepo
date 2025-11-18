// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { LineaRollupBase } from "../../../rollup/LineaRollupBase.sol";
import { IProvideShnarf } from "../../../rollup/interfaces/IProvideShnarf.sol";

/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract NonBlobRollup is LineaRollupBase {
  IProvideShnarf public shnarfProvider;

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

  function blobShnarfExists(bytes32 _shnarf) public view returns (uint256 shnarfExists) {
    shnarfExists = shnarfProvider.blobShnarfExists(_shnarf);
  }

  // ADD OTHER CODE HERE
}

contract SwiftRollup is LineaRollupBase {
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

  function blobShnarfExists(bytes32 _shnarf) public view returns (uint256 shnarfExists) {
    shnarfExists = _blobShnarfExists[_shnarf];
  }

  function submitShnarf(bytes32 _shnarf) external {
    _blobShnarfExists[_shnarf] = 1;
  }
}
