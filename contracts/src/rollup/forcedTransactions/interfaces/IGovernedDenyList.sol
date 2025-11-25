// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Interface to manage denied addresses.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IGovernedDenyList {
  event DeniedStatusesSet(address[] deniedAddresses, bool isDenied);

  /**
   * @dev Thrown when the address is already denied.
   */
  error AddressAlreadyDenied(address deniedAddress);

  /**
   * @notice Function returning denied status for an address.
   * @param _addressToCheck The address to check.
   * @return isDenied The bool indicating if the address is on the deny list.
   */
  function addressIsDenied(address _addressToCheck) external view returns (bool isDenied);
}
