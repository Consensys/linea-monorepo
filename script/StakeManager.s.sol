// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {StakeManager} from "../src/StakeManager.sol";

contract StakeManagerScript is Script {
    StakeManager public stakeManager;

    function setUp() public {}

    function run() public {
        vm.startBroadcast();

        stakeManager = new StakeManager();

        vm.stopBroadcast();
    }
}
