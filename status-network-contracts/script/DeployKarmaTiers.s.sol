// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { KarmaTiers } from "../src/KarmaTiers.sol";

contract DeployKarmaTiersScript is BaseScript {
    function run() public returns (KarmaTiers, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,) = deploymentConfig.activeNetworkConfig();

        vm.startBroadcast(deployer);

        KarmaTiers karmaTiers = new KarmaTiers();

        vm.stopBroadcast();

        return (karmaTiers, deploymentConfig);
    }
}
