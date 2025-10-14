// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { KarmaTest } from "./Karma.t.sol";

contract MintTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function test_Mint() public {
        uint256 amount = 1000e18;
        vm.prank(owner);
        karma.mint(alice, amount);
        assertEq(karma.balanceOf(alice), amount);
    }
}
