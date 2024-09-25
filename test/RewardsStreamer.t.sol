// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { RewardsStreamer } from "../src/RewardsStreamer.sol";
import { MockToken } from "./mocks/MockToken.sol";

contract RewardsStreamerTest is Test {
    MockToken rewardToken;
    MockToken stakingToken;
    RewardsStreamer public streamer;

    address admin = makeAddr("admin");
    address alice = makeAddr("alice");
    address bob = makeAddr("bob");
    address charlie = makeAddr("charlie");
    address dave = makeAddr("dave");

    function setUp() public {
        rewardToken = new MockToken("Reward Token", "RT");
        stakingToken = new MockToken("Staking Token", "ST");
        streamer = new RewardsStreamer(address(stakingToken), address(rewardToken));

        address[4] memory users = [alice, bob, charlie, dave];
        for (uint256 i = 0; i < users.length; i++) {
            stakingToken.mint(users[i], 10_000e18);
            vm.prank(users[i]);
            stakingToken.approve(address(streamer), 10_000e18);
        }

        rewardToken.mint(admin, 10_000e18);
        vm.prank(admin);
        rewardToken.approve(address(streamer), 10_000e18);
    }

    struct CheckStreamerParams {
        uint256 totalStaked;
        uint256 stakingBalance;
        uint256 rewardBalance;
        uint256 rewardIndex;
        uint256 accountedRewards;
    }

    function checkStreamer(CheckStreamerParams memory p) public view {
        assertEq(streamer.totalStaked(), p.totalStaked);
        assertEq(stakingToken.balanceOf(address(streamer)), p.stakingBalance);
        assertEq(rewardToken.balanceOf(address(streamer)), p.rewardBalance);
        assertEq(streamer.rewardIndex(), p.rewardIndex);
        assertEq(streamer.accountedRewards(), p.accountedRewards);
    }

    struct CheckUserParams {
        address user;
        uint256 rewardBalance;
        uint256 stakedBalance;
        uint256 rewardIndex;
    }

    function checkUser(CheckUserParams memory p) public view {
        assertEq(rewardToken.balanceOf(p.user), p.rewardBalance);

        RewardsStreamer.UserInfo memory userInfo = streamer.getUserInfo(p.user);

        assertEq(userInfo.stakedBalance, p.stakedBalance);
        assertEq(userInfo.userRewardIndex, p.rewardIndex);
    }

    function testStake() public {
        streamer.updateRewardIndex();

        // T0
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        // T1
        vm.prank(alice);
        streamer.stake(10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        // T2
        vm.prank(bob);
        streamer.stake(30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        // T3
        vm.prank(admin);
        rewardToken.transfer(address(streamer), 1000e18);
        streamer.updateRewardIndex();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 25e18,
                accountedRewards: 1000e18
            })
        );

        checkUser(CheckUserParams({ user: alice, rewardBalance: 0, stakedBalance: 10e18, rewardIndex: 0 }));
        checkUser(CheckUserParams({ user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0 }));

        // T4
        vm.prank(alice);
        streamer.unstake(10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                stakingBalance: 30e18,
                rewardBalance: 750e18,
                rewardIndex: 25e18,
                accountedRewards: 750e18
            })
        );

        checkUser(CheckUserParams({ user: alice, rewardBalance: 250e18, stakedBalance: 0e18, rewardIndex: 25e18 }));
        checkUser(CheckUserParams({ user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0 }));

        // T5
        vm.prank(charlie);
        streamer.stake(30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                stakingBalance: 60e18,
                rewardBalance: 750e18,
                rewardIndex: 25e18,
                accountedRewards: 750e18
            })
        );

        checkUser(CheckUserParams({ user: alice, rewardBalance: 250e18, stakedBalance: 0e18, rewardIndex: 25e18 }));
        checkUser(CheckUserParams({ user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0 }));
        checkUser(CheckUserParams({ user: charlie, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 25e18 }));

        // T6
        vm.prank(admin);
        rewardToken.transfer(address(streamer), 1000e18);
        streamer.updateRewardIndex();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                stakingBalance: 60e18,
                rewardBalance: 1750e18,
                rewardIndex: 41_666_666_666_666_666_666,
                accountedRewards: 1750e18
            })
        );

        checkUser(CheckUserParams({ user: alice, rewardBalance: 250e18, stakedBalance: 0, rewardIndex: 25e18 }));
        checkUser(CheckUserParams({ user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0 }));
        checkUser(CheckUserParams({ user: charlie, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 25e18 }));

        //T7
        vm.prank(bob);
        streamer.unstake(30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                stakingBalance: 30e18,
                rewardBalance: 500e18 + 20, // 500e18 (with rounding error of 20 wei)
                rewardIndex: 41_666_666_666_666_666_666,
                accountedRewards: 500e18 + 20
            })
        );

        checkUser(CheckUserParams({ user: alice, rewardBalance: 250e18, stakedBalance: 0, rewardIndex: 25e18 }));
        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 1_249_999_999_999_999_999_980, // 750e18 + 500e18 (with rounding error)
                stakedBalance: 0,
                rewardIndex: 41_666_666_666_666_666_666
            })
        );
    }
}
