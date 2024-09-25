// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Script, console } from "forge-std/Script.sol";
import { RewardsStreamer } from "../src/RewardsStreamer.sol";

contract RewardsStreamerScript is Script {
    RewardsStreamer public rewardsStreamer;

    function setUp() public { }

    function run() public {
        vm.startBroadcast();

        // stakeManager = new StakeManager();

        vm.stopBroadcast();
    }
}
