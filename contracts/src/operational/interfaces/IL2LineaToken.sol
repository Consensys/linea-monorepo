// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

/**
 * @title Simplified L2 Linea Token Interface.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IL2LineaToken {
  /**
   * @dev Returns the amount of tokens owned by `account`.
   * @param account The address of the account to query.
   */
  function balanceOf(address account) external view returns (uint256);
}
