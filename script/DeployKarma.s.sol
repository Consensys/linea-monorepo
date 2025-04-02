// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import { Karma } from "../src/Karma.sol";

contract DeployKarmaScript is BaseScript {
    function run() public returns (Karma, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer,) = deploymentConfig.activeNetworkConfig();

        vm.startBroadcast(deployer);

        // Deploy Karma logic contract
        bytes memory initializeData = abi.encodeCall(Karma.initialize, (deployer));
        address impl = address(new Karma());
        // Create upgradeable proxy
        address proxy = address(new ERC1967Proxy(impl, initializeData));

        vm.stopBroadcast();

        return (Karma(proxy), deploymentConfig);
    }
}
