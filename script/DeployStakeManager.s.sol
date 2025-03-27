// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { Clones } from "@openzeppelin/contracts/proxy/Clones.sol";

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";

import { TransparentProxy } from "../src/TransparentProxy.sol";
import { StakeManager } from "../src/StakeManager.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { VaultFactory } from "../src/VaultFactory.sol";

contract DeployStakeManagerScript is BaseScript {
    function run() public returns (StakeManager, VaultFactory, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer, address stakingToken,) = deploymentConfig.activeNetworkConfig();

        bytes memory initializeData = abi.encodeCall(StakeManager.initialize, (deployer, stakingToken));

        vm.startBroadcast(deployer);

        // Deploy StakeManager logic contract
        address impl = address(new StakeManager());
        // Create upgradeable proxy
        address proxy = address(new TransparentProxy(impl, initializeData));

        // Create vault implementation for proxy clones
        address vaultImplementation = address(new StakeVault(IERC20(stakingToken)));
        address proxyClone = Clones.clone(vaultImplementation);

        // Whitelist vault implementation codehash
        StakeManager(proxy).setTrustedCodehash(proxyClone.codehash, true);

        // Create vault factory
        VaultFactory vaultFactory = new VaultFactory(deployer, proxy, vaultImplementation);

        vm.stopBroadcast();

        return (StakeManager(proxy), vaultFactory, deploymentConfig);
    }
}
