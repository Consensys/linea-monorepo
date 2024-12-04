// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { BaseScript } from "./Base.s.sol";
import { DeploymentConfig } from "./DeploymentConfig.s.sol";
import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";
import { StakeManagerProxy } from "../src/StakeManagerProxy.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { StakeVault } from "../src/StakeVault.sol";

contract DeployRewardsStreamerMPScript is BaseScript {
    function run() public returns (RewardsStreamerMP, DeploymentConfig) {
        DeploymentConfig deploymentConfig = new DeploymentConfig(broadcaster);
        (address deployer, address stakingToken) = deploymentConfig.activeNetworkConfig();

        bytes memory initializeData = abi.encodeCall(RewardsStreamerMP.initialize, (deployer, stakingToken));

        vm.startBroadcast(deployer);
        address impl = address(new RewardsStreamerMP());
        address proxy = address(new StakeManagerProxy(impl, initializeData));
        vm.stopBroadcast();

        RewardsStreamerMP stakeManager = RewardsStreamerMP(proxy);
        StakeVault tempVault = new StakeVault(address(this), IStakeManagerProxy(proxy));
        bytes32 vaultCodeHash = address(tempVault).codehash;

        vm.startBroadcast(deployer);
        stakeManager.setTrustedCodehash(vaultCodeHash, true);
        vm.stopBroadcast();

        return (stakeManager, deploymentConfig);
    }
}
