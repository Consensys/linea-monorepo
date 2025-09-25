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
        return _run();
    }

    /**
     * @dev Deploys KarmaTiers contract for testing purposes and returns the instance along with deployment config.
     * The deployer/owner of the KarmaTiers contract will be set to the broadcaster address.
     * @return karmaTiers The deployed KarmaTiers contract instance.
     * @return deploymentConfig The DeploymentConfig instance for the current network.
     */
    function runForTest() public returns (KarmaTiers, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        KarmaTiers karmaTiers = _run();
        return (karmaTiers, deploymentConfig);
    }

    /**
     * @dev Deploys KarmaTiers contract within a broadcast context and returns the instance.
     * @return karmaTiers The deployed KarmaTiers contract instance.
     */
    function _run() internal broadcast returns (KarmaTiers) {
        return deploy();
    }

    /**
     * @dev Deploys KarmaTiers contract and returns the instance.
     * Note: This function does not handle broadcasting; it should be called within a broadcast context.
     * @return karmaTiers The deployed KarmaTiers contract instance.
     */
    function deploy() public returns (KarmaTiers karmaTiers) {
        karmaTiers = new KarmaTiers();
    }
}
