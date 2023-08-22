// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

interface IGenericErrors {
  /**
   * @dev Thrown when a parameter is the zero address.
   */
  error ZeroAddressNotAllowed();
}
