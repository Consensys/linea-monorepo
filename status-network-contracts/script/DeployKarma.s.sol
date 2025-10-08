// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import { Karma } from "../src/Karma.sol";

/**
 * @dev This script deploys the Karma contract as an upgradeable proxy using OpenZeppelin's ERC1967Proxy.
 * It provides functions to deploy for production use and for testing purposes.
 * The deploy function handles the deployment of the logic contract and the creation of the proxy.
 */
contract DeployKarmaScript is BaseScript {
    /**
     * @dev Deploys Karma contract for production use and returns the instance
     * along with the address of the logic contract.
     * The deployer/owner of the Karma contract will be set to the broadcaster address.
     * @return karma The deployed Karma contract instance.
     * @return impl The address of the Karma logic contract.
     */
    function run() public returns (Karma karma, address impl) {
        (karma, impl) = deploy(broadcaster);
    }

    /**
     * @dev Deploys Karma contract for testing purposes and returns the instance along with deployment config.
     * The deployer/owner of the Karma contract will be set to the broadcaster address.
     * @return karma The deployed Karma contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest() public returns (Karma, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (Karma karma,) = deploy(broadcaster);
        return (karma, deploymentConfig);
    }

    /**
     * @dev Deploys Karma contract and returns the instance.
     * @param deployer The address that will be set as the deployer/owner of the Karma contract.
     * @return karma The deployed Karma contract instance.
     * @return impl The address of the Karma logic contract.
     */
    function deploy(address deployer) public returns (Karma, address) {
        vm.startBroadcast(deployer);
        // Deploy Karma logic contract
        bytes memory initializeData = abi.encodeCall(Karma.initialize, (deployer));
        address impl = address(new Karma());
        // Create upgradeable proxy
        address proxy = address(new ERC1967Proxy(impl, initializeData));
        vm.stopBroadcast();

        return (Karma(proxy), impl);
    }
}
