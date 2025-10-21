// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

/**
 * @title Simplified L1 Linea Token Interface.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IL1LineaToken {
  /**
   * @dev Destroys `amount` tokens from the caller.
   * @param amount The amount of tokens to burn.
   */
  function burn(uint256 amount) external;

  /**
   * @dev Returns the amount of tokens owned by `account`.
   * @param account The address of the account to query.
   * @return The amount of tokens owned by `account`.
   */
  function balanceOf(address account) external view returns (uint256);

  /**
   * @notice Synchronizes the total supply of the L1 token to the L2 token.
   * @dev This function sends a message to the L2 token contract to sync the total supply.
   * @dev NB: This function is permissionless on purpose, allowing anyone to trigger the sync.
   */
  function syncTotalSupplyToL2() external;
}
