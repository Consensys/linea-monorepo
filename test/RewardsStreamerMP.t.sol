// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { MockToken } from "./mocks/MockToken.sol";
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

    function setUp() public virtual {
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
        assertEq(streamer.totalStaked(), p.totalStaked, "wrong total staked");
        assertEq(streamer.totalMP(), p.totalMP, "wrong total MP");
        assertEq(streamer.potentialMP(), p.potentialMP, "wrong potential MP");
        assertEq(stakingToken.balanceOf(address(streamer)), p.stakingBalance, "wrong staking balance");
        assertEq(rewardToken.balanceOf(address(streamer)), p.rewardBalance, "wrong reward balance");
        assertEq(streamer.rewardIndex(), p.rewardIndex, "wrong reward index");
        assertEq(streamer.accountedRewards(), p.accountedRewards, "wrong accounted rewards");
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

        RewardsStreamerMP.UserInfo memory userInfo = streamer.getUserInfo(p.user);

        assertEq(userInfo.stakedBalance, p.stakedBalance, "wrong user staked balance");
        assertEq(userInfo.userRewardIndex, p.rewardIndex, "wrong user reward index");
        assertEq(userInfo.userMP, p.userMP, "wrong user MP");
        assertEq(userInfo.userPotentialMP, p.userPotentialMP, "wrong user potential MP");
    }

    function _stake(address account, uint256 amount, uint256 lockupTime) public {
        vm.prank(account);
        streamer.stake(amount, lockupTime);
    }

    function _unstake(address account, uint256 amount) public {
        vm.prank(account);
        streamer.unstake(amount);
    }

    function _addReward(uint256 amount) public {
        vm.prank(admin);
        rewardToken.transfer(address(streamer), amount);
        streamer.updateGlobalState();
    }

    function _calculateBonusMP(uint256 amount, uint256 lockupTime) public view returns (uint256) {
        return amount
            * (lockupTime * streamer.MAX_MULTIPLIER() * streamer.SCALE_FACTOR() / streamer.MAX_LOCKING_PERIOD())
            / streamer.SCALE_FACTOR();
    }
}

contract IntegrationTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
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

        // T2
        vm.prank(bob);
        streamer.stake(30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                potentialMP: 160e18,
                stakingBalance: 40e18,
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

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        // T3
        vm.prank(admin);
        rewardToken.transfer(address(streamer), 1000e18);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                potentialMP: 160e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17, // 1000 rewards / (40 staked + 40 MP) = 12.5
                accountedRewards: 1000e18
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

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        // T4
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days / 2));
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 60e18, // 6 months passed, 20 MP accrued
                potentialMP: 140e18, // 160 - 20
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                // 6 months passed and more MPs have been accrued
                // so we need to adjust the reward index
                rewardIndex: 10e18,
                accountedRewards: 1000e18
            })
        );

        // T5
        vm.prank(alice);
        streamer.unstake(10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMP: 45e18, // 60 - 15 from Alice (10 + 6 months = 5)
                potentialMP: 105e18, // Alice's initial potential MP: 40. 5 already accrued in 6 months. new potentialMP
                    // = 140 - 35 = 105
                stakingBalance: 30e18,
                rewardBalance: 750e18,
                rewardIndex: 10e18,
                accountedRewards: 750e18
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                rewardIndex: 10e18,
                userMP: 0e18,
                userPotentialMP: 0e18
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        // T5
        vm.prank(charlie);
        streamer.stake(30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMP: 75e18,
                potentialMP: 225e18,
                stakingBalance: 60e18,
                rewardBalance: 750e18,
                rewardIndex: 10e18,
                accountedRewards: 750e18
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                rewardIndex: 10e18,
                userMP: 0e18,
                userPotentialMP: 0e18
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        checkUser(
            CheckUserParams({
                user: charlie,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 10e18,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        // T6
        vm.prank(admin);
        rewardToken.transfer(address(streamer), 1000e18);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMP: 75e18,
                potentialMP: 225e18,
                stakingBalance: 60e18,
                rewardBalance: 1750e18,
                rewardIndex: 17_407_407_407_407_407_407,
                accountedRewards: 1750e18
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                rewardIndex: 10e18,
                userMP: 0e18,
                userPotentialMP: 0e18
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        checkUser(
            CheckUserParams({
                user: charlie,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 10e18,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );

        //T7
        vm.prank(bob);
        streamer.unstake(30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMP: 30e18,
                // 225 - 105 from bob who had 120 potential MP and had accrued 15
                potentialMP: 120e18,
                stakingBalance: 30e18,
                // 1750 - (750 + 555.55) = 444.44
                rewardBalance: 444_444_444_444_444_444_475,
                rewardIndex: 17_407_407_407_407_407_407,
                accountedRewards: 444_444_444_444_444_444_475
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                rewardIndex: 10e18,
                userMP: 0,
                userPotentialMP: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                // bob had 30 staked + 30 initial MP + 15 MP accrued in 6 months
                // so in the second bucket we have 1000 rewards with
                // bob's weight = 75
                // charlie's weight = 60
                // total weight = 135
                // bobs rewards = 1000 * 75 / 135 = 555.555555555555555555
                // bobs total rewards = 555.55 + 750 of the first bucket = 1305.55
                rewardBalance: 1_305_555_555_555_555_555_525,
                stakedBalance: 0e18,
                rewardIndex: 17_407_407_407_407_407_407,
                userMP: 0,
                userPotentialMP: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: charlie,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 10e18,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );
    }
}

contract StakeTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_StakeOneAccount() public {
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

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
    }

    function test_StakeOneAccountAndRewards() public {
        _stake(alice, 10e18, 0);

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

        // 1000 rewards generated
        _addReward(1000e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                potentialMP: 40e18,
                stakingBalance: 10e18,
                rewardBalance: 1000e18,
                rewardIndex: 50e18, // (1000 rewards / (10 staked + 10 MP)) = 50
                accountedRewards: 1000e18
            })
        );
    }

    function test_StakeOneAccountWithMinLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKING_PERIOD();
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKING_PERIOD) / SCALE_FACTOR)
                totalMP: stakeAmount + expectedBonusMP,
                potentialMP: stakeAmount * streamer.MAX_MULTIPLIER() + expectedBonusMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );
    }

    function test_StakeOneAccountWithMaxLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MAX_LOCKING_PERIOD();
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, 10e18, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKING_PERIOD) / SCALE_FACTOR)
                totalMP: stakeAmount + expectedBonusMP,
                potentialMP: stakeAmount * streamer.MAX_MULTIPLIER() + expectedBonusMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );
    }

    function test_StakeOneAccountWithRandomLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKING_PERIOD() + 13 days;
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKING_PERIOD) / SCALE_FACTOR)
                totalMP: stakeAmount + expectedBonusMP,
                potentialMP: stakeAmount * streamer.MAX_MULTIPLIER() + expectedBonusMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );
    }

    function test_StakeMultipleAccounts() public {
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

        // Bob stakes 30 tokens
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                potentialMP: 160e18,
                stakingBalance: 40e18,
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

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );
    }

    function test_StakeMultipleAccountsAndRewards() public {
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

        // Bob stakes 30 tokens
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                potentialMP: 160e18,
                stakingBalance: 40e18,
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

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 30e18,
                rewardIndex: 0,
                userMP: 30e18,
                userPotentialMP: 120e18
            })
        );
        // 1000 rewards generated
        _addReward(1000e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                potentialMP: 160e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17, // (1000 rewards / (40 staked + 40 MP)) = 12,5
                accountedRewards: 1000e18
            })
        );
    }

    function test_StakeMultipleAccountsWithMinLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MIN_LOCKING_PERIOD();
        uint256 aliceExpectedBonusMP = _calculateBonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = 0;
        uint256 bobExpectedBonusMP = _calculateBonusMP(bobStakeAmount, bobLockUpPeriod);

        // alice stakes with lockup period
        _stake(alice, aliceStakeAmount, aliceLockUpPeriod);

        // Bob stakes 30 tokens
        _stake(bob, bobStakeAmount, bobLockUpPeriod);

        uint256 sumOfStakeAmount = aliceStakeAmount + bobStakeAmount;
        uint256 sumOfExpectedBonusMP = aliceExpectedBonusMP + bobExpectedBonusMP;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: sumOfStakeAmount,
                totalMP: sumOfStakeAmount + sumOfExpectedBonusMP,
                potentialMP: sumOfStakeAmount * streamer.MAX_MULTIPLIER() + sumOfExpectedBonusMP,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );
    }

    function test_StakeMultipleAccountsWithRandomLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MAX_LOCKING_PERIOD() - 21 days;
        uint256 aliceExpectedBonusMP = _calculateBonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = streamer.MIN_LOCKING_PERIOD() + 43 days;
        uint256 bobExpectedBonusMP = _calculateBonusMP(bobStakeAmount, bobLockUpPeriod);

        // alice stakes with lockup period
        _stake(alice, aliceStakeAmount, aliceLockUpPeriod);

        // Bob stakes 30 tokens
        _stake(bob, bobStakeAmount, bobLockUpPeriod);

        uint256 sumOfStakeAmount = aliceStakeAmount + bobStakeAmount;
        uint256 sumOfExpectedBonusMP = aliceExpectedBonusMP + bobExpectedBonusMP;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: sumOfStakeAmount,
                totalMP: sumOfStakeAmount + sumOfExpectedBonusMP,
                potentialMP: sumOfStakeAmount * streamer.MAX_MULTIPLIER() + sumOfExpectedBonusMP,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );
    }
}

contract UnstakeTest is StakeTest {
    function setUp() public override {
        super.setUp();
    }

    function test_UnstakeOneAccount() public {
        test_StakeOneAccount();

        _unstake(alice, 8e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 2e18,
                totalMP: 2e18,
                potentialMP: 8e18,
                stakingBalance: 2e18,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 0,
                stakedBalance: 2e18,
                rewardIndex: 0,
                userMP: 2e18,
                userPotentialMP: 8e18
            })
        );

        _unstake(alice, 2e18);

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
    }

    function test_UnstakeOneAccountAndRewards() public {
        test_StakeOneAccountAndRewards();

        _unstake(alice, 8e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 2e18,
                totalMP: 2e18,
                potentialMP: 8e18,
                stakingBalance: 2e18,
                rewardBalance: 0, // rewards are all paid out to alice
                rewardIndex: 50e18,
                accountedRewards: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 1000e18,
                stakedBalance: 2e18,
                rewardIndex: 50e18, // alice reward index has been updated
                userMP: 2e18,
                userPotentialMP: 8e18
            })
        );
    }

    function test_UnstakeMultipleAccounts() public {
        test_StakeMultipleAccounts();

        _unstake(alice, 10e18);
        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMP: 20e18,
                potentialMP: 80e18,
                stakingBalance: 20e18,
                rewardBalance: 0,
                rewardIndex: 0,
                accountedRewards: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 0,
                stakedBalance: 0,
                rewardIndex: 0,
                userMP: 0,
                userPotentialMP: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 0,
                stakedBalance: 20e18,
                rewardIndex: 0,
                userMP: 20e18,
                userPotentialMP: 80e18
            })
        );
    }

    function test_UnstakeMultipleAccountsAndRewards() public {
        test_StakeMultipleAccountsAndRewards();

        _unstake(alice, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMP: 30e18,
                potentialMP: 120e18,
                stakingBalance: 30e18,
                // alice owned a 25% of the pool, so 25% of the rewards are paid out to alice (250)
                rewardBalance: 750e18,
                rewardIndex: 125e17, // reward index remains unchanged
                accountedRewards: 750e18
            })
        );

        checkUser(
            CheckUserParams({
                user: alice,
                rewardBalance: 250e18,
                stakedBalance: 0,
                rewardIndex: 125e17,
                userMP: 0,
                userPotentialMP: 0
            })
        );

        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMP: 20e18,
                potentialMP: 80e18,
                stakingBalance: 20e18,
                rewardBalance: 0, // bob should've now gotten the rest of the rewards
                rewardIndex: 125e17,
                accountedRewards: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 750e18,
                stakedBalance: 20e18,
                rewardIndex: 125e17,
                userMP: 20e18,
                userPotentialMP: 80e18
            })
        );

        _unstake(bob, 20e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMP: 0,
                potentialMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 125e17,
                accountedRewards: 0
            })
        );

        checkUser(
            CheckUserParams({
                user: bob,
                rewardBalance: 750e18,
                stakedBalance: 0,
                rewardIndex: 125e17,
                userMP: 0,
                userPotentialMP: 0
            })
        );
    }
}
