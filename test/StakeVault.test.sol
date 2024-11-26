// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";

import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { MockToken } from "./mocks/MockToken.sol";

contract StakeVaultTest is Test {
    RewardsStreamerMP internal streamer;

    StakeVault internal stakeVault;

    MockToken internal rewardToken;

    MockToken internal stakingToken;

    address internal alice = makeAddr("alice");

    function _createTestVault(address owner) internal returns (StakeVault vault) {
        vm.prank(owner);
        vault = new StakeVault(owner, streamer);

        if (!streamer.isTrustedCodehash(address(vault).codehash)) {
            streamer.setTrustedCodehash(address(vault).codehash, true);
        }
    }

    function setUp() public virtual {
        rewardToken = new MockToken("Reward Token", "RT");
        stakingToken = new MockToken("Staking Token", "ST");
        streamer = new RewardsStreamerMP(address(this), address(stakingToken), address(rewardToken));

        stakingToken.mint(alice, 10_000e18);
        stakeVault = _createTestVault(alice);

        vm.prank(alice);
        stakingToken.approve(address(stakeVault), 10_000e18);
    }
}

contract StakingTokenTest is StakeVaultTest {
    function setUp() public override {
        StakeVaultTest.setUp();
    }

    function testStakeToken() public view {
        assertEq(address(stakeVault.STAKING_TOKEN()), address(stakingToken));
    }
}
