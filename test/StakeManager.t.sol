// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Test, console} from "forge-std/Test.sol";
import {StakeManager} from "../src/StakeManager.sol";

contract StakeManagerTest is Test {
    StakeManager public stakeManager;

    function setUp() public {
        stakeManager = new StakeManager();
    }
}
