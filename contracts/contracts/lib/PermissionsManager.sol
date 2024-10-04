// SPDX-License-Identifier: AGPL-3.0
pragma solidity >=0.8.19 <=0.8.26;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { IPermissionsManager } from "../interfaces/IPermissionsManager.sol";

/**
 * @title Contract to manage cross-chain function pausing.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract PermissionsManager is Initializable, AccessControlUpgradeable, IPermissionsManager, IGenericErrors {
  /**
   * @notice Sets permissions for a list of addresses and their roles.
   * @dev This function is a reinitializer and can only be called once per version.
   * @param _roleAddresses The list of addresses and their roles.
   */
  function __Permissions_init(RoleAddress[] calldata _roleAddresses) internal onlyInitializing {
    uint256 roleAddressesLength = _roleAddresses.length;

    for (uint256 i; i < roleAddressesLength; i++) {
      if (_roleAddresses[i].addressWithRole == address(0)) {
        revert ZeroAddressNotAllowed();
      }

      if (_roleAddresses[i].role == 0x0) {
        revert ZeroHashNotAllowed();
      }

      _grantRole(_roleAddresses[i].role, _roleAddresses[i].addressWithRole);
    }
  }
}
