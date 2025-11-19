// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { Eip4844BlobAcceptor } from "../../../rollup/Eip4844BlobAcceptor.sol";
import { CalldataBlobAcceptor } from "../../../rollup/CalldataBlobAcceptor.sol";
import { IProvideShnarf } from "../../../rollup/interfaces/IProvideShnarf.sol";

/**
 * @title Contract to manage cross-chain messaging on L1, L2 data submission, and rollup proof verification.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract LineaRollupWithoutOldClaiming is Eip4844BlobAcceptor, CalldataBlobAcceptor {
  /// @custom:oz-upgrades-unsafe-allow constructor
  constructor() {
    _disableInitializers();
  }

  /**
   * @notice Initializes LineaRollup and underlying service dependencies - used for new networks only.
   * @dev DEFAULT_ADMIN_ROLE is set for the security council.
   * @dev OPERATOR_ROLE is set for operators.
   * @dev Note: This is used for new testnets and local/CI testing, and will not replace existing proxy based contracts.
   * @param _initializationData The initial data used for contract initialization.
   */
  function initialize(InitializationData calldata _initializationData) external initializer {
    bytes32 genesisShnarf = _computeShnarf(
      EMPTY_HASH,
      EMPTY_HASH,
      _initializationData.initialStateRootHash,
      EMPTY_HASH,
      EMPTY_HASH
    );

    _blobShnarfExists[genesisShnarf] = SHNARF_EXISTS_DEFAULT_VALUE;

    __LineaRollup_init(_initializationData, genesisShnarf);
  }

  /**
   * @notice Reinitializes LineaRollup and sets the _shnarfProvider to itself.
   */
  function reinitializeSettingShnarfProvider() external reinitializer(8) {
    address proxyAdmin;
    assembly {
      proxyAdmin := sload(PROXY_ADMIN_SLOT)
    }
    require(msg.sender == proxyAdmin, CallerNotProxyAdmin());

    _shnarfProvider = IProvideShnarf(address(this));
  }
}
