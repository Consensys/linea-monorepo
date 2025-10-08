// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { KarmaTiers } from "../src/KarmaTiers.sol";

/**
 * @dev Script for deploying KarmaTiers contract.
 */
contract DeployKarmaTiersScript is BaseScript {
    /**
     * @dev Deploys KarmaTiers contract for production use and returns the instance.
     * The deployer/owner of the KarmaTiers contract will be set to the broadcaster address.
     * @return karmaTiers The deployed KarmaTiers contract instance.
     */
    function run() public returns (KarmaTiers) {
        return deploy(broadcaster);
    }

    /**
     * @dev Deploys KarmaTiers contract for testing purposes and returns the instance along with deployment config.
     * The deployer/owner of the KarmaTiers contract will be set to the broadcaster address.
     * @return karmaTiers The deployed KarmaTiers contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest() public returns (KarmaTiers, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        KarmaTiers karmaTiers = deploy(broadcaster);
        return (karmaTiers, deploymentConfig);
    }

    /**
     * @dev Deploys KarmaTiers contract and returns the instance.
     * @param deployer The address that will be set as the deployer/owner of the KarmaTiers contract.
     * @return karmaTiers The deployed KarmaTiers contract instance.
     */
    function deploy(address deployer) public returns (KarmaTiers karmaTiers) {
        vm.startBroadcast(deployer);
        karmaTiers = new KarmaTiers();
        vm.stopBroadcast();
    }
}
