// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { BaseScript } from "./Base.s.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";

contract UpgradeRewardsStreamerMPScript is BaseScript {
    function run(address admin, IStakeManagerProxy currentImplProxy) public {
        address deployer = broadcaster;
        if (admin != address(0)) {
            deployer = admin;
        }
        vm.startBroadcast(deployer);
        // Replace this with actual new version of the contract
        address nextImpl = address(new RewardsStreamerMP());
        bytes memory initializeData;
        UUPSUpgradeable(address(currentImplProxy)).upgradeToAndCall(nextImpl, initializeData);
        vm.stopBroadcast();
    }
}
