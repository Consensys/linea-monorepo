// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { BeaconProxy } from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";

/**
 * @notice Deploys upgradeable beacon proxy clones for the LidoStVaultYieldProvider.
 * @custom:security-contact security-report@linea.build
 */
contract LidoStVaultYieldProviderFactory {
  /**
   * @notice Emitted whenever a new LidoStVaultYieldProvider is deployed.
   * @param providerAddress The newly created LidoStVaultYieldProvider address.
   */
  event LidoStVaultYieldProviderCreated(address indexed providerAddress);

  /// @notice Beacon that points to the current LidoStVaultYieldProvider implementation.
  address public immutable BEACON;

  /**
   * @param _beacon Address of the upgradeable beacon shared by all LidoStVaultYieldProvider beacon proxies.
   */
  constructor(address _beacon) {
    BEACON = _beacon;
  }

  /**
   * @notice Creates LidoStVaultYieldProvider instance.
   * @dev LidoStVaultYieldProvider initialization is handled via permissioned YieldManager.addYieldProvider().
   * @return yieldProviderAddress The address of the deployed LidoStVaultYieldProvider beacon proxy.
   */
  function createLidoStVaultYieldProvider() external returns (address yieldProviderAddress) {
    yieldProviderAddress = address(new BeaconProxy(BEACON, ""));
    emit LidoStVaultYieldProviderCreated(yieldProviderAddress);
  }
}
