// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { BaseScript } from "./Base.s.sol";
import { Karma } from "../src/Karma.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

/**
 * @dev This script upgrades the Karma contract to a new implementation.
 * It uses the UUPSUpgradeable pattern to perform the upgrade.
 * The address of the current Karma proxy must be provided via the "KARMA_PROXY_ADDRESS" environment variable.
 * The deployer/admin of the upgrade transaction can be specified via an optional parameter; if not provided, it
 * defaults to the broadcaster address.
 */
contract UpgradeKarmaScript is BaseScript {
    /// @dev Error thrown when the KARMA_PROXY_ADDRESS environment variable is not set.
    error KarmaProxyAddressNotSet();

    /**
     * @dev Upgrades the Karma contract to a new implementation and returns the address of the new
     * implementation.
     * The deployer/admin of the upgrade transaction will be set to the broadcaster address or can be overridden by
     * providing an admin address.
     * @return nextImpl The address of the new Karma implementation contract.
     */
    function run() public returns (address) {
        address currentImplProxy = vm.envAddress("KARMA_PROXY_ADDRESS");
        if (currentImplProxy == address(0)) {
            revert KarmaProxyAddressNotSet();
        }
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,) = deploymentConfig.activeNetworkConfig();
        return runWithAdminAndProxy(deployer, currentImplProxy);
    }

    /**
     * @dev Upgrades the Karma contract to a new implementation using the specified admin address and current
     * proxy instance.
     * @param admin The address to be used as the deployer/admin for the upgrade transaction. If set to address(0), it
     * defaults to the broadcaster address.
     * @param currentImplProxy The instance of the current Karma proxy contract.
     * @return nextImpl The address of the new Karma implementation contract.
     */
    function runWithAdminAndProxy(address admin, address currentImplProxy) public returns (address) {
        address deployer = broadcaster;
        if (admin != address(0)) {
            deployer = admin;
        }
        vm.startBroadcast(deployer);
        // Replace this with actual new version of the contract
        address nextImpl = address(new Karma());
        UUPSUpgradeable(address(currentImplProxy)).upgradeTo(nextImpl);
        vm.stopBroadcast();
        return nextImpl;
    }
}
