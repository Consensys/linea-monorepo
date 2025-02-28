// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { Karma } from "../src/Karma.sol";

contract DeployKarmaScript is BaseScript {
    function run() public returns (Karma) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,,) = deploymentConfig.activeNetworkConfig();

        vm.startBroadcast(deployer);
        address karma = address(new Karma());
        vm.stopBroadcast();

        return Karma(karma);
    }
}
