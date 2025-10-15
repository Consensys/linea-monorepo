// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { KarmaTest } from "./Karma.t.sol";

contract OverflowTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhen_MintingCausesOverflow() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), type(uint224).max, 1000);
        vm.stopBroadcast();

        vm.prank(owner);
        vm.expectRevert();
        karma.mint(owner, 1e18);
    }

    function test_RevertWhen_SettingRewardCausesOverflow() public {
        vm.prank(owner);
        karma.mint(owner, type(uint224).max);

        vm.prank(owner);
        vm.expectRevert();
        karma.setReward(address(distributor1), 1e18, 1000);
    }
}
