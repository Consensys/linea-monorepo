// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";
import { StakeManagerProxy } from "../src/StakeManagerProxy.sol";
import { MockToken } from "./mocks/MockToken.sol";
import { StackOverflowStakeManager } from "./mocks/StackOverflowStakeManager.sol";

contract RewardsStreamerMPTest is Test {
    MockToken stakingToken;
    RewardsStreamerMP public streamer;

    address admin = makeAddr("admin");
    address alice = makeAddr("alice");
    address bob = makeAddr("bob");
    address charlie = makeAddr("charlie");
    address dave = makeAddr("dave");

    mapping(address owner => address vault) public vaults;

    function setUp() public virtual {
        stakingToken = new MockToken("Staking Token", "ST");

        bytes memory initializeData = abi.encodeCall(RewardsStreamerMP.initialize, (admin, address(stakingToken)));
        address impl = address(new RewardsStreamerMP());
        address proxy = address(new StakeManagerProxy(impl, initializeData));
        streamer = RewardsStreamerMP(proxy);

        address[4] memory accounts = [alice, bob, charlie, dave];
        for (uint256 i = 0; i < accounts.length; i++) {
            // ensure user has tokens
            stakingToken.mint(accounts[i], 10_000e18);

            // each user creates a vault
            StakeVault vault = _createTestVault(accounts[i]);
            vaults[accounts[i]] = address(vault);

            vm.prank(accounts[i]);
            stakingToken.approve(address(vault), 10_000e18);
        }
    }

    struct CheckStreamerParams {
        uint256 totalStaked;
        uint256 totalMP;
        uint256 totalMaxMP;
        uint256 stakingBalance;
        uint256 rewardBalance;
        uint256 rewardIndex;
    }

    function checkStreamer(CheckStreamerParams memory p) public view {
        assertEq(streamer.totalStaked(), p.totalStaked, "wrong total staked");
        assertEq(streamer.totalMP(), p.totalMP, "wrong total MP");
        assertEq(streamer.totalMaxMP(), p.totalMaxMP, "wrong totalMaxMP MP");
        // assertEq(rewardToken.balanceOf(address(streamer)), p.rewardBalance, "wrong reward balance");
        // assertEq(streamer.rewardIndex(), p.rewardIndex, "wrong reward index");
    }

    struct CheckAccountParams {
        address account;
        uint256 rewardBalance;
        uint256 stakedBalance;
        uint256 vaultBalance;
        uint256 rewardIndex;
        uint256 accountMP;
        uint256 maxMP;
    }

    function checkAccount(CheckAccountParams memory p) public view {
        // assertEq(rewardToken.balanceOf(p.account), p.rewardBalance, "wrong account reward balance");

        RewardsStreamerMP.Account memory accountInfo = streamer.getAccount(p.account);

        assertEq(accountInfo.stakedBalance, p.stakedBalance, "wrong account staked balance");
        assertEq(stakingToken.balanceOf(p.account), p.vaultBalance, "wrong vault balance");
        // assertEq(accountInfo.accountRewardIndex, p.rewardIndex, "wrong account reward index");
        assertEq(accountInfo.accountMP, p.accountMP, "wrong account MP");
        assertEq(accountInfo.maxMP, p.maxMP, "wrong account max MP");
    }

    function _createTestVault(address owner) internal returns (StakeVault vault) {
        vm.prank(owner);
        vault = new StakeVault(owner, IStakeManagerProxy(address(streamer)));

        if (!streamer.isTrustedCodehash(address(vault).codehash)) {
            vm.prank(admin);
            streamer.setTrustedCodehash(address(vault).codehash, true);
        }
    }

    function _stake(address account, uint256 amount, uint256 lockupTime) public {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.stake(amount, lockupTime);
    }

    function _unstake(address account, uint256 amount) public {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.unstake(amount);
    }

    function _emergencyExit(address account) public {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.emergencyExit(account);
    }

    function _leave(address account) public {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.leave(account);
    }

    function _calculateBonusMP(uint256 amount, uint256 lockupTime) public view returns (uint256) {
        return amount
            * (lockupTime * streamer.MAX_MULTIPLIER() * streamer.SCALE_FACTOR() / streamer.MAX_LOCKUP_PERIOD())
            / streamer.SCALE_FACTOR();
    }

    function _calculeAccuredMP(uint256 totalStaked, uint256 timeDiff) public view returns (uint256) {
        return (timeDiff * totalStaked * streamer.MP_RATE_PER_YEAR()) / (365 days * streamer.SCALE_FACTOR());
    }

    function _calculateTimeToMPLimit(uint256 amount) public view returns (uint256) {
        uint256 maxMP = amount * streamer.MAX_MULTIPLIER();
        uint256 mpPerYear = (amount * streamer.MP_RATE_PER_YEAR()) / streamer.SCALE_FACTOR();
        uint256 timeInSeconds = (maxMP * 365 days) / mpPerYear;
        return timeInSeconds;
    }
}

contract IntegrationTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function testStakeFoo() public {
        streamer.updateGlobalState();

        // T0
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMP: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // T1
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        // T2
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        // T3
        vm.prank(admin);
        // rewardToken.transfer(address(streamer), 1000e18);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17 // 1000 rewards / (40 staked + 40 MP) = 12.5
             })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
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
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                // 6 months passed and more MPs have been accrued
                // so we need to adjust the reward index
                rewardIndex: 10e18
            })
        );

        // T5
        _unstake(alice, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMP: 45e18, // 60 - 15 from Alice (10 + 6 months = 5)
                totalMaxMP: 150e18, // 200e18 - (10e18 * 5) = 150e18
                stakingBalance: 30e18,
                rewardBalance: 750e18,
                rewardIndex: 10e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                accountMP: 0e18,
                maxMP: 0e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        // T5
        _stake(charlie, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMP: 75e18,
                totalMaxMP: 300e18,
                stakingBalance: 60e18,
                rewardBalance: 750e18,
                rewardIndex: 10e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                accountMP: 0e18,
                maxMP: 0e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        // T6
        vm.prank(admin);
        // rewardToken.transfer(address(streamer), 1000e18);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMP: 75e18,
                totalMaxMP: 300e18,
                stakingBalance: 60e18,
                rewardBalance: 1750e18,
                rewardIndex: 17_407_407_407_407_407_407
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                accountMP: 0e18,
                maxMP: 0e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        //T7
        _unstake(bob, 30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMP: 30e18,
                totalMaxMP: 150e18,
                stakingBalance: 30e18,
                // 1750 - (750 + 555.55) = 444.44
                rewardBalance: 444_444_444_444_444_444_475,
                rewardIndex: 17_407_407_407_407_407_407
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                accountMP: 0,
                maxMP: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                // bob had 30 staked + 30 initial MP + 15 MP accrued in 6 months
                // so in the second bucket we have 1000 rewards with
                // bob's weight = 75
                // charlie's weight = 60
                // total weight = 135
                // bobs rewards = 1000 * 75 / 135 = 555.555555555555555555
                // bobs total rewards = 555.55 + 750 of the first bucket = 1305.55
                rewardBalance: 1_305_555_555_555_555_555_525,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 17_407_407_407_407_407_407,
                accountMP: 0,
                maxMP: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                accountMP: 30e18,
                maxMP: 150e18
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
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );
    }

    function test_StakeOneAccountAndRewards() public {
        _stake(alice, 10e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 1000e18,
                rewardIndex: 50e18 // (1000 rewards / (10 staked + 10 MP)) = 50
             })
        );
    }

    function test_StakeOneAccountWithMinLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD();
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10e18 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMP: stakeAmount + expectedBonusMP,
                totalMaxMP: 52_465_753_424_657_534_240,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountWithMaxLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MAX_LOCKUP_PERIOD();
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMP: stakeAmount + expectedBonusMP,
                totalMaxMP: 90e18,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountWithRandomLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD() + 13 days;
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMP: stakeAmount + expectedBonusMP,
                totalMaxMP: 52_821_917_808_219_178_080,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountMPIncreasesMaxMPDoesNotChange() public {
        uint256 stakeAmount = 15e18;
        uint256 totalMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;
        uint256 totalMP = stakeAmount;

        _stake(alice, stakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: stakeAmount,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days));

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        uint256 expectedMPIncrease = stakeAmount; // 1 year passed, 1 MP accrued per token staked
        totalMP = totalMP + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: totalMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                accountMP: totalMP, // accountMP == totalMP because only one account is staking
                maxMP: totalMaxMP
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days / 2));

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        expectedMPIncrease = stakeAmount / 2; // 1/2 year passed, 1/2 MP accrued per token staked
        totalMP = totalMP + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: totalMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                accountMP: totalMP, // accountMP == totalMP because only one account is staking
                maxMP: totalMaxMP
            })
        );
    }

    function test_StakeOneAccountReachingMPLimit() public {
        uint256 stakeAmount = 15e18;
        uint256 totalMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;
        uint256 totalMP = stakeAmount;

        _stake(alice, stakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: stakeAmount,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                accountMP: totalMP, // accountMP == totalMP because only one account is staking
                maxMP: totalMaxMP // maxMP == totalMaxMP because only one account is staking
             })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        uint256 timeToMaxMP = _calculateTimeToMPLimit(stakeAmount);
        vm.warp(currentTime + timeToMaxMP);

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: totalMaxMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                accountMP: totalMaxMP,
                maxMP: totalMaxMP
            })
        );

        // move forward in time to check we're not producing more MP
        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + 1);

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: totalMaxMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
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
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
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
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17 // (1000 rewards / (40 staked + 40 MP)) = 12,5
             })
        );
    }

    function test_StakeMultipleAccountsWithMinLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MIN_LOCKUP_PERIOD();
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
                totalMaxMP: 202_465_753_424_657_534_240,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeMultipleAccountsWithRandomLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MAX_LOCKUP_PERIOD() - 21 days;
        uint256 aliceExpectedBonusMP = _calculateBonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = streamer.MIN_LOCKUP_PERIOD() + 43 days;
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
                totalMaxMP: 250_356_164_383_561_643_820,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeMultipleAccountsMPIncreasesMaxMPDoesNotChange() public {
        uint256 aliceStakeAmount = 15e18;
        uint256 aliceMP = aliceStakeAmount;
        uint256 aliceMaxMP = aliceStakeAmount * streamer.MAX_MULTIPLIER() + aliceMP;

        uint256 bobStakeAmount = 5e18;
        uint256 bobMP = bobStakeAmount;
        uint256 bobMaxMP = bobStakeAmount * streamer.MAX_MULTIPLIER() + bobMP;

        uint256 totalMP = aliceStakeAmount + bobStakeAmount;
        uint256 totalStaked = aliceStakeAmount + bobStakeAmount;
        uint256 totalMaxMP = aliceMaxMP + bobMaxMP;

        _stake(alice, aliceStakeAmount, 0);
        _stake(bob, bobStakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: totalStaked,
                totalMP: totalMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: aliceStakeAmount,
                vaultBalance: aliceStakeAmount,
                rewardIndex: 0,
                accountMP: aliceMP,
                maxMP: aliceMaxMP
            })
        );
        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: bobStakeAmount,
                vaultBalance: bobStakeAmount,
                rewardIndex: 0,
                accountMP: bobMP,
                maxMP: bobMaxMP
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days));

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);
        streamer.updateAccountMP(vaults[bob]);

        uint256 aliceExpectedMPIncrease = aliceStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 bobExpectedMPIncrease = bobStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        aliceMP = aliceMP + aliceExpectedMPIncrease;
        bobMP = bobMP + bobExpectedMPIncrease;
        totalMP = totalMP + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: totalStaked,
                totalMP: totalMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: aliceStakeAmount,
                vaultBalance: aliceStakeAmount,
                rewardIndex: 0,
                accountMP: aliceMP,
                maxMP: aliceMaxMP
            })
        );
        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: bobStakeAmount,
                vaultBalance: bobStakeAmount,
                rewardIndex: 0,
                accountMP: bobMP,
                maxMP: bobMaxMP
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days / 2));

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);
        streamer.updateAccountMP(vaults[bob]);

        aliceExpectedMPIncrease = aliceStakeAmount / 2;
        bobExpectedMPIncrease = bobStakeAmount / 2;
        totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        aliceMP = aliceMP + aliceExpectedMPIncrease;
        bobMP = bobMP + bobExpectedMPIncrease;
        totalMP = totalMP + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: totalStaked,
                totalMP: totalMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: aliceStakeAmount,
                vaultBalance: aliceStakeAmount,
                rewardIndex: 0,
                accountMP: aliceMP,
                maxMP: aliceMaxMP
            })
        );
        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: bobStakeAmount,
                vaultBalance: bobStakeAmount,
                rewardIndex: 0,
                accountMP: bobMP,
                maxMP: bobMaxMP
            })
        );
    }
}

contract UnstakeTest is StakeTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_UnstakeOneAccount() public {
        test_StakeOneAccount();

        _unstake(alice, 8e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 2e18,
                totalMP: 2e18,
                totalMaxMP: 10e18,
                stakingBalance: 2e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 2e18,
                vaultBalance: 2e18,
                rewardIndex: 0,
                accountMP: 2e18,
                maxMP: 10e18
            })
        );

        _unstake(alice, 2e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMP: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_UnstakeOneAccountAndAccruedMP() public {
        test_StakeOneAccount();

        // wait for 1 year
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days));

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 20e18, // total MP must have been doubled
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // unstake half of the tokens
        _unstake(alice, 5e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 5e18, // 10 - 5
                totalMP: 10e18, // 20 - 10 (5 initial + 5 accrued)
                totalMaxMP: 25e18,
                stakingBalance: 5e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_UnstakeOneAccountWithLockUpAndAccruedMP() public {
        test_StakeOneAccountWithMinLockUp();

        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD();
        // 10e18 is what's used in `test_StakeOneAccountWithMinLockUp`
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockUpPeriod);

        // wait for 1 year
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (365 days));

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMP: (stakeAmount + expectedBonusMP) + stakeAmount, // we do `+ stakeAmount` we've accrued
                    // `stakeAmount` after 1 year
                totalMaxMP: 52_465_753_424_657_534_240,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // unstake half of the tokens
        _unstake(alice, 5e18);
        expectedBonusMP = _calculateBonusMP(5e18, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 5e18,
                totalMP: (5e18 + expectedBonusMP) + 5e18,
                totalMaxMP: 26_232_876_712_328_767_120,
                stakingBalance: 5e18,
                rewardBalance: 0,
                rewardIndex: 0
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
                totalMaxMP: 10e18,
                stakingBalance: 2e18,
                rewardBalance: 0, // rewards are all paid out to alice
                rewardIndex: 50e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 1000e18,
                stakedBalance: 2e18,
                vaultBalance: 2e18,
                rewardIndex: 50e18, // alice reward index has been updated
                accountMP: 2e18,
                maxMP: 10e18
            })
        );
    }

    function test_UnstakeBonusMPAndAccuredMP() public {
        // setup variables
        uint256 amountStaked = 10e18;
        uint256 secondsLocked = streamer.MIN_LOCKUP_PERIOD();
        uint256 reducedStake = 5e18;
        uint256 increasedTime = 365 days;

        //initialize memory placehodlders
        uint256[4] memory timestamp;
        uint256[4] memory increasedAccuredMP;
        uint256[4] memory predictedBonusMP;
        uint256[4] memory predictedAccuredMP;
        uint256[4] memory predictedTotalMP;
        uint256[4] memory predictedTotalMaxMP;
        uint256[4] memory totalStaked;

        //stages variables setup
        uint256 stage = 0; // first stage: initialization
        {
            timestamp[stage] = block.timestamp;
            totalStaked[stage] = amountStaked;
            predictedBonusMP[stage] = totalStaked[stage] + _calculateBonusMP(totalStaked[stage], secondsLocked);
            predictedTotalMaxMP[stage] = 52_465_753_424_657_534_240;
            increasedAccuredMP[stage] = 0; //no increased accured MP in first stage
            predictedAccuredMP[stage] = 0; //no accured MP in first stage
            predictedTotalMP[stage] = predictedBonusMP[stage] + predictedAccuredMP[stage];
        }
        stage++; // second stage: progress in time
        {
            timestamp[stage] = timestamp[stage - 1] + increasedTime;
            totalStaked[stage] = totalStaked[stage - 1];
            predictedBonusMP[stage] = predictedBonusMP[stage - 1]; //no change in bonusMP in second stage
            predictedTotalMaxMP[stage] = predictedTotalMaxMP[stage - 1];
            increasedAccuredMP[stage] = _calculeAccuredMP(totalStaked[stage], timestamp[stage] - timestamp[stage - 1]);
            predictedAccuredMP[stage] = predictedAccuredMP[stage - 1] + increasedAccuredMP[stage];
            predictedTotalMP[stage] = predictedBonusMP[stage] + predictedAccuredMP[stage];
        }
        stage++; //third stage: reduced stake
        {
            timestamp[stage] = timestamp[stage - 1]; //no time increased in third stage
            totalStaked[stage] = totalStaked[stage - 1] - reducedStake;
            //bonusMP from this stage is a proportion from the difference of remainingStake and amountStaked
            //if the account reduced 50% of its stake, the bonusMP should be reduced by 50%
            predictedBonusMP[stage] = (totalStaked[stage] * predictedBonusMP[stage - 1]) / totalStaked[stage - 1];
            predictedTotalMaxMP[stage] = (totalStaked[stage] * predictedTotalMaxMP[stage - 1]) / totalStaked[stage - 1];
            increasedAccuredMP[stage] = 0; //no accuredMP in third stage;
            //total accuredMP from this stage is a proportion from the difference of remainingStake and amountStaked
            //if the account reduced 50% of its stake, the accuredMP should be reduced by 50%
            predictedAccuredMP[stage] = (totalStaked[stage] * predictedAccuredMP[stage - 1]) / totalStaked[stage - 1];
            predictedTotalMP[stage] = predictedBonusMP[stage] + predictedAccuredMP[stage];
        }

        // stages execution
        stage = 0; // first stage: initialization
        {
            _stake(alice, amountStaked, secondsLocked);
            {
                RewardsStreamerMP.Account memory accountInfo = streamer.getAccount(vaults[alice]);
                assertEq(accountInfo.stakedBalance, totalStaked[stage], "stage 1: wrong account staked balance");
                assertEq(accountInfo.accountMP, predictedTotalMP[stage], "stage 1: wrong account MP");
                assertEq(accountInfo.maxMP, predictedTotalMaxMP[stage], "stage 1: wrong account max MP");

                assertEq(streamer.totalStaked(), totalStaked[stage], "stage 1: wrong total staked");
                assertEq(streamer.totalMP(), predictedTotalMP[stage], "stage 1: wrong total MP");
                assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 1: wrong totalMaxMP MP");
            }
        }

        stage++; // second stage: progress in time
        vm.warp(timestamp[stage]);
        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);
        {
            RewardsStreamerMP.Account memory accountInfo = streamer.getAccount(vaults[alice]);
            assertEq(accountInfo.stakedBalance, totalStaked[stage], "stage 2: wrong account staked balance");
            assertEq(accountInfo.accountMP, predictedTotalMP[stage], "stage 2: wrong account MP");
            assertEq(accountInfo.maxMP, predictedTotalMaxMP[stage], "stage 2: wrong account max MP");

            assertEq(streamer.totalStaked(), totalStaked[stage], "stage 2: wrong total staked");
            assertEq(streamer.totalMP(), predictedTotalMP[stage], "stage 2: wrong total MP");
            assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 2: wrong totalMaxMP MP");
        }

        stage++; // third stage: reduced stake
        _unstake(alice, reducedStake);
        {
            RewardsStreamerMP.Account memory accountInfo = streamer.getAccount(vaults[alice]);
            assertEq(accountInfo.stakedBalance, totalStaked[stage], "stage 3: wrong account staked balance");
            assertEq(accountInfo.accountMP, predictedTotalMP[stage], "stage 3: wrong account MP");
            assertEq(accountInfo.maxMP, predictedTotalMaxMP[stage], "stage 3: wrong account max MP");

            assertEq(streamer.totalStaked(), totalStaked[stage], "stage 3: wrong total staked");
            assertEq(streamer.totalMP(), predictedTotalMP[stage], "stage 3: wrong total MP");
            assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 3: wrong totalMaxMP MP");
        }
    }

    function test_UnstakeMultipleAccounts() public {
        test_StakeMultipleAccounts();

        _unstake(alice, 10e18);
        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMP: 20e18,
                totalMaxMP: 100e18,
                stakingBalance: 20e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                accountMP: 0,
                maxMP: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 0,
                accountMP: 20e18,
                maxMP: 100e18
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
                totalMaxMP: 150e18,
                stakingBalance: 30e18,
                // alice owned a 25% of the pool, so 25% of the rewards are paid out to alice (250)
                rewardBalance: 750e18,
                rewardIndex: 125e17 // reward index remains unchanged
             })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 125e17,
                accountMP: 0,
                maxMP: 0
            })
        );

        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMP: 20e18,
                totalMaxMP: 100e18,
                stakingBalance: 20e18,
                rewardBalance: 0, // bob should've now gotten the rest of the rewards
                rewardIndex: 125e17
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 750e18,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 125e17,
                accountMP: 20e18,
                maxMP: 100e18
            })
        );

        _unstake(bob, 20e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMP: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 125e17
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 750e18,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 125e17,
                accountMP: 0,
                maxMP: 0
            })
        );
    }
}

contract LockTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function _lock(address account, uint256 lockPeriod) internal {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.lock(lockPeriod);
    }

    function test_LockWithoutPriorLock() public {
        // Setup - alice stakes 10 tokens without lock
        uint256 stakeAmount = 10e18;
        _stake(alice, stakeAmount, 0);

        uint256 initialAccountMP = stakeAmount; // 10e18
        uint256 initialMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount; // 50e18

        // Verify initial state
        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                accountMP: initialAccountMP,
                maxMP: initialMaxMP
            })
        );

        // Lock for 1 year
        uint256 lockPeriod = 365 days;
        uint256 expectedBonusMP = _calculateBonusMP(stakeAmount, lockPeriod);

        _lock(alice, lockPeriod);

        // Check updated state
        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                accountMP: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP
            })
        );
    }

    function test_LockFailsWithNoStake() public {
        vm.expectRevert(RewardsStreamerMP.StakingManager__InsufficientBalance.selector);
        _lock(alice, 365 days);
    }

    function test_LockFailsWithInvalidPeriod() public {
        _stake(alice, 10e18, 0);

        // Test with period = 0
        vm.expectRevert(RewardsStreamerMP.StakingManager__InvalidLockingPeriod.selector);
        _lock(alice, 0);
    }
}

contract EmergencyExitTest is RewardsStreamerMPTest {
    function setUp() public override {
        super.setUp();
    }

    function test_CannotLeaveBeforeEmergencyMode() public {
        _stake(alice, 10e18, 0);
        vm.expectRevert(StakeVault.StakeVault__NotAllowedToExit.selector);
        _emergencyExit(alice);
    }

    function test_OnlyOwnerCanEnableEmergencyMode() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, alice));
        streamer.enableEmergencyMode();
    }

    function test_CannotEnableEmergencyModeTwice() public {
        vm.prank(admin);
        streamer.enableEmergencyMode();

        vm.expectRevert(RewardsStreamerMP.StakingManager__EmergencyModeEnabled.selector);
        vm.prank(admin);
        streamer.enableEmergencyMode();
    }

    function test_EmergencyExitBasic() public {
        uint256 aliceBalance = stakingToken.balanceOf(alice);

        _stake(alice, 10e18, 0);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        // emergency exit will not perform any internal accounting
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        assertEq(stakingToken.balanceOf(alice), aliceBalance, "Alice should get tokens back");
        assertEq(stakingToken.balanceOf(vaults[alice]), 0, "Vault should be empty");
    }

    function test_EmergencyExitWithRewards() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        _stake(alice, 10e18, 0);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 1000e18,
                rewardIndex: 50e18
            })
        );

        // Check Alice staked tokens but no rewards
        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get staked tokens back");
        assertEq(stakingToken.balanceOf(address(vaults[alice])), 0, "Vault should be empty");
    }

    function test_EmergencyExitWithLock() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        _stake(alice, 10e18, 90 days);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        // Check Alice got tokens back despite lock
        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get tokens back despite lock");
        assertEq(stakingToken.balanceOf(address(vaults[alice])), 0, "Vault should be empty");
    }

    function test_EmergencyExitMultipleUsers() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 bobInitialBalance = stakingToken.balanceOf(bob);

        // Setup multiple stakers
        _stake(alice, 10e18, 0);
        _stake(bob, 30e18, 0);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        // Alice exits first
        _emergencyExit(alice);

        // Check intermediate state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17
            })
        );

        // Bob exits
        _emergencyExit(bob);

        // Check final state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMP: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        checkAccount(
            CheckAccountParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 0,
                rewardIndex: 0,
                accountMP: 30e18,
                maxMP: 150e18
            })
        );

        // Verify both users got their tokens back
        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get staked tokens back");
        assertEq(stakingToken.balanceOf(bob), bobInitialBalance, "Bob should get staked tokens back");
        assertEq(stakingToken.balanceOf(vaults[alice]), 0, "Alice vault should have 0 staked tokens");
        assertEq(stakingToken.balanceOf(vaults[bob]), 0, "Bob vault should have 0 staked tokens");
    }

    function test_EmergencyExitToAlternateAddress() public {
        _stake(alice, 10e18, 0);

        address alternateAddress = makeAddr("alternate");
        uint256 alternateInitialBalance = stakingToken.balanceOf(alternateAddress);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        // Alice exits to alternate address
        vm.prank(alice);
        StakeVault aliceVault = StakeVault(vaults[alice]);
        aliceVault.emergencyExit(alternateAddress);

        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                accountMP: 10e18,
                maxMP: 50e18
            })
        );

        // Check alternate address received everything
        assertEq(
            stakingToken.balanceOf(alternateAddress),
            alternateInitialBalance + 10e18,
            "Alternate address should get staked tokens"
        );
    }
}

contract UpgradeTest is RewardsStreamerMPTest {
    function _upgradeStakeManager() internal {
        address newImpl = address(new RewardsStreamerMP());
        bytes memory initializeData;
        vm.prank(admin);
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);
    }

    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhenNotOwner() public {
        address newImpl = address(new RewardsStreamerMP());
        bytes memory initializeData;
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, alice));
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);
    }

    function test_UpgradeStakeManager() public {
        // first, change state of existing stake manager
        _stake(alice, 10e18, 0);

        // check initial state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // next, upgrade the stake manager
        _upgradeStakeManager();

        // ensure state is available in upgraded contract
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMP: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }
}

contract LeaveTest is RewardsStreamerMPTest {
    function _upgradeStakeManager() internal {
        address newImpl = address(new RewardsStreamerMP());
        bytes memory initializeData;
        vm.prank(admin);
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);
    }

    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhenStakeManagerIsTrusted() public {
        _stake(alice, 10e18, 0);
        vm.expectRevert(StakeVault.StakeVault__NotAllowedToLeave.selector);
        _leave(alice);
    }

    function test_LeaveShouldProperlyUpdateAccounting() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        _stake(alice, 100e18, 0);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance - 100e18, "Alice should have staked tokens");

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 100e18,
                totalMP: 100e18,
                totalMaxMP: 500e18,
                stakingBalance: 100e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        _upgradeStakeManager();
        _leave(alice);

        // stake manager properly updates accounting
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMP: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // vault should be empty as funds have been moved out
        checkAccount(
            CheckAccountParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                accountMP: 0,
                maxMP: 0
            })
        );

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice has all her funds back");
    }

    function test_TrustNewStakeManager() public {
        // first, upgrade to new stake manager, marking it as not trusted
        _upgradeStakeManager();

        // ensure vault functions revert if stake manager is not trusted
        vm.expectRevert(StakeVault.StakeVault__StakeManagerImplementationNotTrusted.selector);
        _stake(alice, 100e18, 0);

        // ensure vault functions revert if stake manager is not trusted
        StakeVault vault = StakeVault(vaults[alice]);
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__StakeManagerImplementationNotTrusted.selector);
        vault.lock(365 days);

        // ensure vault functions revert if stake manager is not trusted
        vm.expectRevert(StakeVault.StakeVault__StakeManagerImplementationNotTrusted.selector);
        _unstake(alice, 100e18);

        // now, trust the new stake manager
        address newStakeManagerImpl = IStakeManagerProxy(address(streamer)).implementation();
        vm.prank(alice);
        vault.trustStakeManager(newStakeManagerImpl);

        // stake manager is now trusted, so functions are enabeled again
        _stake(alice, 100e18, 0);

        // however, a trusted manager cannot be left
        vm.expectRevert(StakeVault.StakeVault__NotAllowedToLeave.selector);
        _leave(alice);
    }
}

contract MaliciousUpgradeTest is RewardsStreamerMPTest {
    function _upgradeStakeManager() internal {
        address newImpl = address(new RewardsStreamerMP());
        bytes memory initializeData;
        vm.prank(admin);
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);
    }

    function setUp() public override {
        super.setUp();
    }

    function test_UpgradeStackOverflowStakeManager() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        // first change the existing manager's state
        _stake(alice, 100e18, 0);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 100e18,
                totalMP: 100e18,
                totalMaxMP: 500e18,
                stakingBalance: 100e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // upgrade the manager to a malicious one
        address newImpl = address(new StackOverflowStakeManager());
        bytes memory initializeData;
        vm.prank(admin);
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);

        // alice leaves system and is able to get funds out, despite malicious manager
        _leave(alice);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get her tokens back");
    }
}

contract RewardsStreamerMP_RewardsTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function testSetRewards() public {
        assertEq(streamer.rewardStartTime(), 0);
        assertEq(streamer.rewardEndTime(), 0);
        assertEq(streamer.lastRewardTime(), 0);

        uint256 currentTime = vm.getBlockTimestamp();
        // just to be sure that currentTime is not 0
        // since we are testing that it is used for rewardStartTime
        currentTime += 1 days;
        vm.warp(currentTime);
        vm.prank(admin);
        streamer.setReward(1000, 10);

        assertEq(streamer.rewardStartTime(), currentTime);
        assertEq(streamer.rewardEndTime(), currentTime + 10);
        assertEq(streamer.lastRewardTime(), currentTime);
    }

    function testSetRewards_RevertsNotAuthorized() public {
        vm.prank(alice);
        vm.expectPartialRevert(Ownable.OwnableUnauthorizedAccount.selector);
        streamer.setReward(1000, 10);
    }

    function testSetRewards_RevertsBadDuration() public {
        vm.prank(admin);
        vm.expectRevert(RewardsStreamerMP.StakingManager__DurationCannotBeZero.selector);
        streamer.setReward(1000, 0);
    }

    function testSetRewards_RevertsBadAmount() public {
        vm.prank(admin);
        vm.expectRevert(RewardsStreamerMP.StakingManager__AmountCannotBeZero.selector);
        streamer.setReward(0, 10);
    }

    function testTotalRewardsSupply() public {
        _stake(alice, 100e18, 0);
        assertEq(streamer.totalRewardsSupply(), 0);

        uint256 initialTime = vm.getBlockTimestamp();

        vm.prank(admin);
        streamer.setReward(1000e18, 10 days);
        assertEq(streamer.totalRewardsSupply(), 0);

        for (uint256 i = 0; i <= 10; i++) {
            vm.warp(initialTime + i * 1 days);
            assertEq(streamer.totalRewardsSupply(), 100e18 * i);
        }

        // after the end of the reward period, the total rewards supply does not increase
        vm.warp(initialTime + 11 days);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalRewardsAccrued(), 0);

        uint256 secondRewardTime = initialTime + 20 days;
        vm.warp(secondRewardTime);

        // still the same rewards supply after 20 days
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalRewardsAccrued(), 0);

        // set other 2000 rewards for other 10 days
        vm.prank(admin);
        streamer.setReward(2000e18, 10 days);
        // accrued is 1000 from the previous reward and still 0 for the new one
        assertEq(streamer.totalRewardsSupply(), 1000e18, "totalRewardsSupply should be 1000");
        assertEq(streamer.totalRewardsAccrued(), 1000e18);

        uint256 previousSupply = 1000e18;
        for (uint256 i = 0; i <= 10; i++) {
            vm.warp(secondRewardTime + i * 1 days);
            assertEq(streamer.totalRewardsSupply(), previousSupply + 200e18 * i);
        }
    }

    function testRewardsBalanceOf() public {
        assertEq(streamer.totalRewardsSupply(), 0);

        vm.warp(0);

        uint256 initialTime = vm.getBlockTimestamp();

        _stake(alice, 100e18, 0);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 0);

        vm.prank(admin);
        streamer.setReward(1000e18, 10 days);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 0);

        vm.warp(initialTime + 1 days);

        // FIXME: this is needed to update the global state and account MP
        // Later we should update the functions to use "real-time" values.
        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        uint256 tolerance = 300; // 300 wei

        assertEq(streamer.totalRewardsSupply(), 100e18, "Total rewards supply mismatch");
        assertApproxEqAbs(streamer.rewardsBalanceOf(vaults[alice]), 100e18, tolerance);

        vm.warp(initialTime + 10 days);

        streamer.updateGlobalState();
        streamer.updateAccountMP(vaults[alice]);

        assertEq(streamer.totalRewardsSupply(), 1000e18, "Total rewards supply mismatch");
        assertApproxEqAbs(streamer.rewardsBalanceOf(vaults[alice]), 1000e18, tolerance);
    }
}
