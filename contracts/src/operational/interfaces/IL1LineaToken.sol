// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

/**
 * @title Simplified L1 Linea Token Interface.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IL1LineaToken {
  /**
   * @notice Synchronizes the total supply of the L1 token to the L2 token.
   * @dev This function sends a message to the L2 token contract to sync the total supply.
   * @dev NB: This function is permissionless on purpose, allowing anyone to trigger the sync.
   */
  function syncTotalSupplyToL2() external;
}
