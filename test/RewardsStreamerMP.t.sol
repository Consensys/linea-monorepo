// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test, console} from "forge-std/Test.sol";
import {RewardsStreamerMP} from "../src/RewardsStreamerMP.sol";
import {MockToken} from "./mocks/MockToken.sol";
import "forge-std/console.sol";

contract RewardsStreamerMPTest is Test {
    MockToken rewardToken;
    MockToken stakingToken;
    RewardsStreamerMP public streamer;

    address admin = makeAddr("admin");
    address alice = makeAddr("alice");
    address bob = makeAddr("bob");
    address charlie = makeAddr("charlie");
    address dave = makeAddr("dave");

    function setUp() public {
        rewardToken = new MockToken("Reward Token", "RT");
        stakingToken = new MockToken("Staking Token", "ST");
        streamer = new RewardsStreamerMP(address(stakingToken), address(rewardToken));

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
        uint256 totalMP;
        uint256 potentialMP;
        uint256 stakingBalance;
        uint256 rewardBalance;
        uint256 rewardIndex;
        uint256 accountedRewards;
    }

    function checkStreamer(CheckStreamerParams memory p) public view {
        // assertEq(streamer.totalStaked(), p.totalStaked, "wrong total staked");
        // assertEq(streamer.totalMP(), p.totalMP, "wrong total MP");
        // assertEq(streamer.potentialMP(), p.potentialMP, "wrong potential MP");
        // assertEq(stakingToken.balanceOf(address(streamer)), p.stakingBalance, "wrong staking balance");
        // // assertEq(rewardToken.balanceOf(address(streamer)), p.rewardBalance, "wrong reward balance");
        // assertEq(streamer.rewardIndex(), p.rewardIndex, "wrong reward index");
        // // assertEq(streamer.accountedRewards(), p.accountedRewards, "wrong accounted rewards");
    }

    struct CheckUserParams {
        address user;
        uint256 rewardBalance;
        uint256 stakedBalance;
        uint256 rewardIndex;
        uint256 userMP;
        uint256 userPotentialMP;
    }

    function checkUser(CheckUserParams memory p) public view {
        assertEq(rewardToken.balanceOf(p.user), p.rewardBalance, "wrong user reward balance");

        // RewardsStreamerMP.UserInfo memory userInfo = streamer.getUserInfo(p.user);

        // assertEq(userInfo.stakedBalance, p.stakedBalance, "wrong user staked balance");
        // assertEq(userInfo.userRewardIndex, p.rewardIndex, "wrong user reward index");
        // assertEq(userInfo.userMP, p.userMP, "wrong user MP");
        // assertEq(userInfo.userPotentialMP, p.userPotentialMP, "wrong user potential MP");
    }

    function testStake() public {
        streamer.updateGlobalState();

        // T0
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMP: 0,
                potentialMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        // T1
        // Alice stakes 10 tokens
        vm.prank(alice);
        streamer.stake(10e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                potentialMP: 40e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 0,
                stakedBalance: 10e18,
                rewardIndex: 0,
                userMP: 10e18,
                userPotentialMP: 40e18
            })
        );

        // // T2
        // vm.prank(bob);
        // streamer.stake(30e18, 0);

        // checkStreamer(
        //     CheckStreamerParams({
        //         totalStaked: 40e18,
        //         stakingBalance: 40e18,
        //         rewardBalance: 0,
        //         rewardIndex: 0,
        //         accountedRewards: 0
        //     })
        // );

        // // T3
        // vm.prank(admin);
        // rewardToken.transfer(address(streamer), 1000e18);
        // streamer.updateGlobalState();

        // checkStreamer(
        //     CheckStreamerParams({
        //         totalStaked: 40e18,
        //         stakingBalance: 40e18,
        //         rewardBalance: 1000e18,
        //         rewardIndex: 25e18,
        //         accountedRewards: 1000e18
        //     })
        // );

        // checkUser(
        //     CheckUserParams({
        //         user: alice,
        //         rewardBalance: 0,
        //         stakedBalance: 10e18,
        //         rewardIndex: 0,
        //         userMP: 10e18,
        //         userPotentialMP: 40e18
        //     })
        // );
        // checkUser(
        //     CheckUserParams({
        //         user: bob,
        //         rewardBalance: 0,
        //         stakedBalance: 30e18,
        //         rewardIndex: 0,
        //         userMP: 0,
        //         userPotentialMP: 120e18
        //     })
        // );

        //// T4
        //vm.prank(alice);
        //streamer.unstake(10e18);

        //checkStreamer(
        //    CheckStreamerParams({
        //        totalStaked: 30e18,
        //        stakingBalance: 30e18,
        //        rewardBalance: 750e18,
        //        rewardIndex: 25e18,
        //        accountedRewards: 750e18
        //    })
        //);

        //checkUser(CheckUserParams({user: alice, rewardBalance: 250e18, stakedBalance: 0e18, rewardIndex: 25e18}));
        //checkUser(CheckUserParams({user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0}));

        //// T5
        //vm.prank(charlie);
        //streamer.stake(30e18, 0);

        //checkStreamer(
        //    CheckStreamerParams({
        //        totalStaked: 60e18,
        //        stakingBalance: 60e18,
        //        rewardBalance: 750e18,
        //        rewardIndex: 25e18,
        //        accountedRewards: 750e18
        //    })
        //);

        //checkUser(CheckUserParams({user: alice, rewardBalance: 250e18, stakedBalance: 0e18, rewardIndex: 25e18}));
        //checkUser(CheckUserParams({user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0}));
        //checkUser(CheckUserParams({user: charlie, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 25e18}));

        //// T6
        //vm.prank(admin);
        //rewardToken.transfer(address(streamer), 1000e18);
        //streamer.updateRewardIndex();

        //checkStreamer(
        //    CheckStreamerParams({
        //        totalStaked: 60e18,
        //        stakingBalance: 60e18,
        //        rewardBalance: 1750e18,
        //        rewardIndex: 41666666666666666666,
        //        accountedRewards: 1750e18
        //    })
        //);

        //checkUser(CheckUserParams({user: alice, rewardBalance: 250e18, stakedBalance: 0, rewardIndex: 25e18}));
        //checkUser(CheckUserParams({user: bob, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 0}));
        //checkUser(CheckUserParams({user: charlie, rewardBalance: 0, stakedBalance: 30e18, rewardIndex: 25e18}));

        ////T7
        //vm.prank(bob);
        //streamer.unstake(30e18);

        //checkStreamer(
        //    CheckStreamerParams({
        //        totalStaked: 30e18,
        //        stakingBalance: 30e18,
        //        rewardBalance: 500e18 + 20, // 500e18 (with rounding error of 20 wei)
        //        rewardIndex: 41666666666666666666,
        //        accountedRewards: 500e18 + 20
        //    })
        //);

        //checkUser(CheckUserParams({user: alice, rewardBalance: 250e18, stakedBalance: 0, rewardIndex: 25e18}));
        //checkUser(
        //    CheckUserParams({
        //        user: bob,
        //        rewardBalance: 1249999999999999999980, // 750e18 + 500e18 (with rounding error)
        //        stakedBalance: 0,
        //        rewardIndex: 41666666666666666666
        //    })
        //);
    }
}
