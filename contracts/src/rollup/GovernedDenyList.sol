// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;
import { IGenericErrors } from "../interfaces/IGenericErrors.sol";
import { IGovernedDenyList } from "./interfaces/IGovernedDenyList.sol";
import { AccessControl } from "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title Contract to manage forced transactions on L1.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract GovernedDenyList is AccessControl, IGovernedDenyList {
  bool public useDenyList = true;

  mapping(address => bool) internal deniedAddresses;

  constructor(address _defaultAdmin, address[] memory _initialDeniedAddresses) {
    require(_defaultAdmin != address(0), IGenericErrors.ZeroAddressNotAllowed());

    _grantRole(DEFAULT_ADMIN_ROLE, _defaultAdmin);
    _setAddressesDeniedStatus(_initialDeniedAddresses, true);
  }

  function addressIsDenied(address _addressToCheck) external view returns (bool isDenied) {
    isDenied = deniedAddresses[_addressToCheck];
  }

  function setAddressesDeniedStatus(
    address[] memory _initialDeniedAddresses,
    bool _isDenied
  ) external onlyRole(DEFAULT_ADMIN_ROLE) {
    _setAddressesDeniedStatus(_initialDeniedAddresses, _isDenied);
  }

  function _setAddressesDeniedStatus(address[] memory _initialDeniedAddresses, bool _isDenied) internal {
    uint256 cachedAddressesLength = _initialDeniedAddresses.length;
    for (uint256 i; i < cachedAddressesLength; i++) {
      deniedAddresses[_initialDeniedAddresses[i]] = _isDenied;
    }

    emit DeniedStatusesSet(_initialDeniedAddresses, _isDenied);
  }
}
