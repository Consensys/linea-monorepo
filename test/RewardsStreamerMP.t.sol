// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { DeployRewardsStreamerMPScript } from "../script/DeployRewardsStreamerMP.s.sol";
import { UpgradeRewardsStreamerMPScript } from "../script/UpgradeRewardsStreamerMP.s.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { Clones } from "@openzeppelin/contracts/proxy/Clones.sol";
import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";
import { ITrustedCodehashAccess } from "../src/interfaces/ITrustedCodehashAccess.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { StakeMath } from "../src/math/StakeMath.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { VaultFactory } from "../src/VaultFactory.sol";
import { MockToken } from "./mocks/MockToken.sol";
import { StackOverflowStakeManager } from "./mocks/StackOverflowStakeManager.sol";

contract RewardsStreamerMPTest is StakeMath, Test {
    MockToken internal stakingToken;
    RewardsStreamerMP public streamer;
    VaultFactory public vaultFactory;

    address internal admin;
    address internal alice = makeAddr("alice");
    address internal bob = makeAddr("bob");
    address internal charlie = makeAddr("charlie");
    address internal dave = makeAddr("dave");

    mapping(address owner => address vault) public vaults;

    function setUp() public virtual {
        DeployRewardsStreamerMPScript deployment = new DeployRewardsStreamerMPScript();
        (RewardsStreamerMP stakeManager, VaultFactory _vaultFactory, DeploymentConfig deploymentConfig) =
            deployment.run();

        (address _deployer, address _stakingToken) = deploymentConfig.activeNetworkConfig();

        streamer = stakeManager;
        stakingToken = MockToken(_stakingToken);
        vaultFactory = _vaultFactory;
        admin = _deployer;

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
        uint256 totalMPAccrued;
        uint256 totalMaxMP;
        uint256 stakingBalance;
        uint256 rewardBalance;
        uint256 rewardIndex;
    }

    function checkStreamer(CheckStreamerParams memory p) public view {
        assertEq(streamer.totalStaked(), p.totalStaked, "wrong total staked");
        assertEq(streamer.totalMPAccrued(), p.totalMPAccrued, "wrong total MP");
        assertEq(streamer.totalMaxMP(), p.totalMaxMP, "wrong totalMaxMP MP");
        // assertEq(rewardToken.balanceOf(address(streamer)), p.rewardBalance, "wrong reward balance");
        // assertEq(streamer.rewardIndex(), p.rewardIndex, "wrong reward index");
    }

    struct CheckVaultParams {
        address account;
        uint256 rewardBalance;
        uint256 stakedBalance;
        uint256 vaultBalance;
        uint256 rewardIndex;
        uint256 mpAccrued;
        uint256 maxMP;
    }

    function checkVault(CheckVaultParams memory p) public view {
        // assertEq(rewardToken.balanceOf(p.account), p.rewardBalance, "wrong account reward balance");

        RewardsStreamerMP.VaultData memory vaultData = streamer.getVault(p.account);

        assertEq(vaultData.stakedBalance, p.stakedBalance, "wrong account staked balance");
        assertEq(stakingToken.balanceOf(p.account), p.vaultBalance, "wrong vault balance");
        // assertEq(vaultData.accountRewardIndex, p.rewardIndex, "wrong account reward index");
        assertEq(vaultData.mpAccrued, p.mpAccrued, "wrong account MP");
        assertEq(vaultData.maxMP, p.maxMP, "wrong account max MP");
    }

    struct CheckUserTotalsParams {
        address user;
        uint256 totalStakedBalance;
        uint256 totalMPAccrued;
        uint256 totalMaxMP;
    }

    function checkUserTotals(CheckUserTotalsParams memory p) public view {
        assertEq(streamer.getAccountTotalStakedBalance(p.user), p.totalStakedBalance, "wrong user total stake balance");
        assertEq(streamer.mpBalanceOfAccount(p.user), p.totalMPAccrued, "wrong user total MP");
        assertEq(streamer.getAccountTotalMaxMP(p.user), p.totalMaxMP, "wrong user total MP");
    }

    function _createTestVault(address owner) internal returns (StakeVault vault) {
        vm.prank(owner);
        vault = vaultFactory.createVault();
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

    function _timeToAccrueMPLimit(uint256 amount) internal view returns (uint256) {
        uint256 maxMP = amount * streamer.MAX_MULTIPLIER();
        uint256 timeInSeconds = _timeToAccrueMP(amount, maxMP);
        return timeInSeconds;
    }

    function _upgradeStakeManager() internal {
        UpgradeRewardsStreamerMPScript upgrade = new UpgradeRewardsStreamerMPScript();
        upgrade.run(admin, IStakeManagerProxy(address(streamer)));
    }
}

contract MathTest is RewardsStreamerMPTest {
    function test_CalcInitialMP() public pure {
        assertEq(_initialMP(1), 1, "wrong initial MP");
        assertEq(_initialMP(10e18), 10e18, "wrong initial MP");
        assertEq(_initialMP(20e18), 20e18, "wrong initial MP");
        assertEq(_initialMP(30e18), 30e18, "wrong initial MP");
    }

    function test_CalcAccrueMP() public pure {
        assertEq(_accrueMP(10e18, 0), 0, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days / 2), 5e18, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days), 10e18, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days * 2), 20e18, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days * 3), 30e18, "wrong accrued MP");
    }

    function test_CalcBonusMP() public view {
        assertEq(_bonusMP(10e18, 0), 0, "wrong bonus MP");
        assertEq(_bonusMP(10e18, streamer.MIN_LOCKUP_PERIOD()), 2_465_753_424_657_534_246, "wrong bonus MP");
        assertEq(_bonusMP(10e18, streamer.MIN_LOCKUP_PERIOD() + 13 days), 2_821_917_808_219_178_082, "wrong bonus MP");
        assertEq(_bonusMP(100e18, 0), 0, "wrong bonus MP");
    }

    function test_CalcMaxTotalMP() public view {
        assertEq(_maxTotalMP(10e18, 0), 50e18, "wrong max total MP");
        assertEq(_maxTotalMP(10e18, streamer.MIN_LOCKUP_PERIOD()), 52_465_753_424_657_534_246, "wrong max total MP");
        assertEq(
            _maxTotalMP(10e18, streamer.MIN_LOCKUP_PERIOD() + 13 days), 52_821_917_808_219_178_082, "wrong max total MP"
        );
        assertEq(_maxTotalMP(100e18, 0), 500e18, "wrong max total MP");
    }

    function test_CalcAbsoluteMaxTotalMP() public pure {
        assertEq(_maxAbsoluteTotalMP(10e18), 90e18, "wrong absolute max total MP");
        assertEq(_maxAbsoluteTotalMP(100e18), 900e18, "wrong absolute max total MP");
    }

    function test_CalcMaxAccruedMP() public pure {
        assertEq(_maxAccrueMP(10e18), 40e18, "wrong max accrued MP");
        assertEq(_maxAccrueMP(100e18), 400e18, "wrong max accrued MP");
    }
}

contract VaultRegistrationTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_VaultRegistration() public view {
        address[4] memory accounts = [alice, bob, charlie, dave];
        for (uint256 i = 0; i < accounts.length; i++) {
            address[] memory userVaults = streamer.getAccountVaults(accounts[i]);
            assertEq(userVaults.length, 1, "wrong number of vaults");
            assertEq(userVaults[0], vaults[accounts[i]], "wrong vault address");
        }
    }
}

contract TrustedCodehashAccessTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_RevertWhenProxyCloneCodehashNotTrusted() public {
        // create independent (possibly malicious) StakeVault
        address vaultTpl = address(new StakeVault(stakingToken));
        StakeVault proxyClone = StakeVault(Clones.clone(vaultTpl));
        proxyClone.initialize(address(this), address(streamer));

        // registering already fails as codehash is not trusted
        vm.expectRevert(ITrustedCodehashAccess.TrustedCodehashAccess__UnauthorizedCodehash.selector);
        proxyClone.register();

        // staking fails as codehash is not trusted
        vm.expectRevert(ITrustedCodehashAccess.TrustedCodehashAccess__UnauthorizedCodehash.selector);
        proxyClone.stake(10e10, 0);
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
                totalMPAccrued: 0,
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
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        // T2
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        // T3
        vm.prank(admin);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17 // 1000 rewards / (40 staked + 40 MP) = 12.5
             })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        // T4
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPAccrued: 60e18, // 6 months passed, 20 MP accrued
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
                totalMPAccrued: 45e18, // 60 - 15 from Alice (10 + 6 months = 5)
                totalMaxMP: 150e18, // 200e18 - (10e18 * 5) = 150e18
                stakingBalance: 30e18,
                rewardBalance: 750e18,
                rewardIndex: 10e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                mpAccrued: 0e18,
                maxMP: 0e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        // T5
        _stake(charlie, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMPAccrued: 75e18,
                totalMaxMP: 300e18,
                stakingBalance: 60e18,
                rewardBalance: 750e18,
                rewardIndex: 10e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                mpAccrued: 0e18,
                maxMP: 0e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        // T6
        vm.prank(admin);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMPAccrued: 75e18,
                totalMaxMP: 300e18,
                stakingBalance: 60e18,
                rewardBalance: 1750e18,
                rewardIndex: 17_407_407_407_407_407_407
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                mpAccrued: 0e18,
                maxMP: 0e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        //T7
        _unstake(bob, 30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMPAccrued: 30e18,
                totalMaxMP: 150e18,
                stakingBalance: 30e18,
                // 1750 - (750 + 555.55) = 444.44
                rewardBalance: 444_444_444_444_444_444_475,
                rewardIndex: 17_407_407_407_407_407_407
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0e18,
                vaultBalance: 0e18,
                rewardIndex: 10e18,
                mpAccrued: 0,
                maxMP: 0
            })
        );

        checkVault(
            CheckVaultParams({
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
                mpAccrued: 0,
                maxMP: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                mpAccrued: 30e18,
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
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );
    }

    function test_StakeOneAccountAndRewards() public {
        _stake(alice, 10e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPAccrued: 10e18,
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
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10e18 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountWithMaxLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MAX_LOCKUP_PERIOD();
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMPAccrued: stakeAmount + expectedBonusMP,
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
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountMPIncreasesMaxMPDoesNotChange() public {
        uint256 stakeAmount = 15e18;
        uint256 totalMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;
        uint256 totalMPAccrued = stakeAmount;

        _stake(alice, stakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: stakeAmount,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        uint256 expectedMPIncrease = stakeAmount; // 1 year passed, 1 MP accrued per token staked
        totalMPAccrued = totalMPAccrued + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        expectedMPIncrease = stakeAmount / 2; // 1/2 year passed, 1/2 MP accrued per token staked
        totalMPAccrued = totalMPAccrued + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP
            })
        );
    }

    function test_StakeOneAccountReachingMPLimit() public {
        uint256 stakeAmount = 15e18;
        uint256 totalMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;
        uint256 totalMPAccrued = stakeAmount;

        _stake(alice, stakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: stakeAmount,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP // maxMP == totalMaxMP because only one account is staking
             })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        uint256 timeToMaxMP = _timeToAccrueMP(stakeAmount, totalMaxMP - totalMPAccrued);
        vm.warp(currentTime + timeToMaxMP);

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: totalMaxMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMaxMP,
                maxMP: totalMaxMP
            })
        );

        // move forward in time to check we're not producing more MP
        currentTime = vm.getBlockTimestamp();
        // increasing time by some big enough time such that MPs are actually generated
        vm.warp(currentTime + 14 days);

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: totalMaxMP,
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
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
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
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPAccrued: 40e18,
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
        uint256 aliceExpectedBonusMP = _bonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = 0;
        uint256 bobExpectedBonusMP = _bonusMP(bobStakeAmount, bobLockUpPeriod);

        // alice stakes with lockup period
        _stake(alice, aliceStakeAmount, aliceLockUpPeriod);

        // Bob stakes 30 tokens
        _stake(bob, bobStakeAmount, bobLockUpPeriod);

        uint256 sumOfStakeAmount = aliceStakeAmount + bobStakeAmount;
        uint256 sumOfExpectedBonusMP = aliceExpectedBonusMP + bobExpectedBonusMP;
        uint256 expectedMaxTotalMP =
            _maxTotalMP(aliceStakeAmount, aliceLockUpPeriod) + _maxTotalMP(bobStakeAmount, bobLockUpPeriod);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: sumOfStakeAmount,
                totalMPAccrued: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeMultipleAccountsWithRandomLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MAX_LOCKUP_PERIOD() - 21 days;
        uint256 aliceExpectedBonusMP = _bonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = streamer.MIN_LOCKUP_PERIOD() + 43 days;
        uint256 bobExpectedBonusMP = _bonusMP(bobStakeAmount, bobLockUpPeriod);

        // alice stakes with lockup period
        _stake(alice, aliceStakeAmount, aliceLockUpPeriod);

        // Bob stakes 30 tokens
        _stake(bob, bobStakeAmount, bobLockUpPeriod);

        uint256 sumOfStakeAmount = aliceStakeAmount + bobStakeAmount;
        uint256 sumOfExpectedBonusMP = aliceExpectedBonusMP + bobExpectedBonusMP;
        uint256 expectedMaxTotalMP =
            _maxTotalMP(aliceStakeAmount, aliceLockUpPeriod) + _maxTotalMP(bobStakeAmount, bobLockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: sumOfStakeAmount,
                totalMPAccrued: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
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

        uint256 totalMPAccrued = aliceStakeAmount + bobStakeAmount;
        uint256 totalStaked = aliceStakeAmount + bobStakeAmount;
        uint256 totalMaxMP = aliceMaxMP + bobMaxMP;

        _stake(alice, aliceStakeAmount, 0);
        _stake(bob, bobStakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: totalStaked,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: aliceStakeAmount,
                vaultBalance: aliceStakeAmount,
                rewardIndex: 0,
                mpAccrued: aliceMP,
                maxMP: aliceMaxMP
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: bobStakeAmount,
                vaultBalance: bobStakeAmount,
                rewardIndex: 0,
                mpAccrued: bobMP,
                maxMP: bobMaxMP
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);
        streamer.updateVaultMP(vaults[bob]);

        uint256 aliceExpectedMPIncrease = aliceStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 bobExpectedMPIncrease = bobStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        aliceMP = aliceMP + aliceExpectedMPIncrease;
        bobMP = bobMP + bobExpectedMPIncrease;
        totalMPAccrued = totalMPAccrued + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: totalStaked,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: aliceStakeAmount,
                vaultBalance: aliceStakeAmount,
                rewardIndex: 0,
                mpAccrued: aliceMP,
                maxMP: aliceMaxMP
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: bobStakeAmount,
                vaultBalance: bobStakeAmount,
                rewardIndex: 0,
                mpAccrued: bobMP,
                maxMP: bobMaxMP
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);
        streamer.updateVaultMP(vaults[bob]);

        aliceExpectedMPIncrease = aliceStakeAmount / 2;
        bobExpectedMPIncrease = bobStakeAmount / 2;
        totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        aliceMP = aliceMP + aliceExpectedMPIncrease;
        bobMP = bobMP + bobExpectedMPIncrease;
        totalMPAccrued = totalMPAccrued + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: totalStaked,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: aliceStakeAmount,
                vaultBalance: aliceStakeAmount,
                rewardIndex: 0,
                mpAccrued: aliceMP,
                maxMP: aliceMaxMP
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: bobStakeAmount,
                vaultBalance: bobStakeAmount,
                rewardIndex: 0,
                mpAccrued: bobMP,
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
                totalMPAccrued: 2e18,
                totalMaxMP: 10e18,
                stakingBalance: 2e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 2e18,
                vaultBalance: 2e18,
                rewardIndex: 0,
                mpAccrued: 2e18,
                maxMP: 10e18
            })
        );

        _unstake(alice, 2e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMPAccrued: 0,
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
        vm.warp(currentTime + (YEAR));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPAccrued: 20e18, // total MP must have been doubled
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
                totalMPAccrued: 10e18, // 20 - 10 (5 initial + 5 accrued)
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
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 unstakeAmount = 5e18;
        uint256 warpLength = (365 days);
        // wait for 1 year
        uint256 currentTime = vm.getBlockTimestamp();

        vm.warp(currentTime + (warpLength));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPAccrued: (stakeAmount + expectedBonusMP) + stakeAmount, // we do `+ stakeAmount` we've accrued
                    // `stakeAmount` after 1 year
                totalMaxMP: _maxTotalMP(stakeAmount, lockUpPeriod),
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
        uint256 newBalance = stakeAmount - unstakeAmount;
        // unstake half of the tokens
        _unstake(alice, unstakeAmount);

        uint256 expectedTotalMP =
            _initialMP(newBalance) + _bonusMP(newBalance, lockUpPeriod) + _accrueMP(newBalance, warpLength);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: newBalance,
                totalMPAccrued: expectedTotalMP,
                totalMaxMP: _maxTotalMP(newBalance, lockUpPeriod),
                stakingBalance: newBalance,
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
                totalMPAccrued: 2e18,
                totalMaxMP: 10e18,
                stakingBalance: 2e18,
                rewardBalance: 0, // rewards are all paid out to alice
                rewardIndex: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 1000e18,
                stakedBalance: 2e18,
                vaultBalance: 2e18,
                rewardIndex: 50e18, // alice reward index has been updated
                mpAccrued: 2e18,
                maxMP: 10e18
            })
        );
    }

    function test_UnstakeBonusMPAndAccuredMP() public {
        // setup variables
        uint256 amountStaked = 10e18;
        uint256 secondsLocked = streamer.MIN_LOCKUP_PERIOD();
        uint256 reducedStake = 5e18;
        uint256 increasedTime = YEAR;

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
            predictedBonusMP[stage] = totalStaked[stage] + _bonusMP(totalStaked[stage], secondsLocked);
            predictedTotalMaxMP[stage] = _maxTotalMP(totalStaked[stage], secondsLocked);
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
            // solhint-disable-next-line max-line-length
            increasedAccuredMP[stage] = _accrueMP(totalStaked[stage], timestamp[stage] - timestamp[stage - 1]);
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
                RewardsStreamerMP.VaultData memory vaultData = streamer.getVault(vaults[alice]);
                assertEq(vaultData.stakedBalance, totalStaked[stage], "stage 1: wrong account staked balance");
                assertEq(vaultData.mpAccrued, predictedTotalMP[stage], "stage 1: wrong account MP");
                assertEq(vaultData.maxMP, predictedTotalMaxMP[stage], "stage 1: wrong account max MP");

                assertEq(streamer.totalStaked(), totalStaked[stage], "stage 1: wrong total staked");
                assertEq(streamer.totalMPAccrued(), predictedTotalMP[stage], "stage 1: wrong total MP");
                assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 1: wrong totalMaxMP MP");
            }
        }

        stage++; // second stage: progress in time
        vm.warp(timestamp[stage]);
        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);
        {
            RewardsStreamerMP.VaultData memory vaultData = streamer.getVault(vaults[alice]);
            assertEq(vaultData.stakedBalance, totalStaked[stage], "stage 2: wrong account staked balance");
            assertEq(vaultData.mpAccrued, predictedTotalMP[stage], "stage 2: wrong account MP");
            assertEq(vaultData.maxMP, predictedTotalMaxMP[stage], "stage 2: wrong account max MP");

            assertEq(streamer.totalStaked(), totalStaked[stage], "stage 2: wrong total staked");
            assertEq(streamer.totalMPAccrued(), predictedTotalMP[stage], "stage 2: wrong total MP");
            assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 2: wrong totalMaxMP MP");
        }

        stage++; // third stage: reduced stake
        _unstake(alice, reducedStake);
        {
            RewardsStreamerMP.VaultData memory vaultData = streamer.getVault(vaults[alice]);
            assertEq(vaultData.stakedBalance, totalStaked[stage], "stage 3: wrong account staked balance");
            assertEq(vaultData.mpAccrued, predictedTotalMP[stage], "stage 3: wrong account MP");
            assertEq(vaultData.maxMP, predictedTotalMaxMP[stage], "stage 3: wrong account max MP");

            assertEq(streamer.totalStaked(), totalStaked[stage], "stage 3: wrong total staked");
            assertEq(streamer.totalMPAccrued(), predictedTotalMP[stage], "stage 3: wrong total MP");
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
                totalMPAccrued: 20e18,
                totalMaxMP: 100e18,
                stakingBalance: 20e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 0,
                maxMP: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 0,
                mpAccrued: 20e18,
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
                totalMPAccrued: 30e18,
                totalMaxMP: 150e18,
                stakingBalance: 30e18,
                // alice owned a 25% of the pool, so 25% of the rewards are paid out to alice (250)
                rewardBalance: 750e18,
                rewardIndex: 125e17 // reward index remains unchanged
             })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 125e17,
                mpAccrued: 0,
                maxMP: 0
            })
        );

        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMPAccrued: 20e18,
                totalMaxMP: 100e18,
                stakingBalance: 20e18,
                rewardBalance: 0, // bob should've now gotten the rest of the rewards
                rewardIndex: 125e17
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 750e18,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 125e17,
                mpAccrued: 20e18,
                maxMP: 100e18
            })
        );

        _unstake(bob, 20e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMPAccrued: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 125e17
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 750e18,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 125e17,
                mpAccrued: 0,
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
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP
            })
        );

        // Lock for 1 year
        uint256 lockPeriod = YEAR;
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockPeriod);

        _lock(alice, lockPeriod);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP
            })
        );
    }

    function test_LockFailsWithNoStake() public {
        vm.expectRevert(StakeMath.StakeMath__InsufficientBalance.selector);
        _lock(alice, YEAR);
    }

    function test_LockFailsWithZero() public {
        _stake(alice, 10e18, 0);

        // Test with period = 0
        vm.expectRevert(RewardsStreamerMP.StakingManager__LockingPeriodCannotBeZero.selector);
        _lock(alice, 0);
    }

    function test_LockFailsWithInvalidPeriod(uint256 _lockPeriod) public {
        vm.assume(_lockPeriod > 0);
        vm.assume(_lockPeriod < MIN_LOCKUP_PERIOD || _lockPeriod > MAX_LOCKUP_PERIOD);
        vm.assume(_lockPeriod < (type(uint256).max - block.timestamp)); //prevents arithmetic overflow

        _stake(alice, 10e18, 0);

        vm.expectRevert(StakeMath.StakeMath__InvalidLockingPeriod.selector);
        _lock(alice, _lockPeriod);
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
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 10e18,
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
                totalMPAccrued: 10e18,
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
                totalMPAccrued: 40e18,
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
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 30e18,
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

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 10e18,
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
                totalMPAccrued: 10e18,
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
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }
}

contract LeaveTest is RewardsStreamerMPTest {
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
                totalMPAccrued: 100e18,
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
                totalMPAccrued: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // vault should be empty as funds have been moved out
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 0,
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
                totalMPAccrued: 100e18,
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

        uint256 initialTime = vm.getBlockTimestamp();

        _stake(alice, 100e18, 0);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 0);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.prank(admin);
        streamer.setReward(1000e18, 10 days);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 0);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.warp(initialTime + 1 days);

        assertEq(streamer.totalRewardsSupply(), 100e18, "Total rewards supply mismatch");
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 100e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.warp(initialTime + 5 days);
        _stake(bob, 100e18, 0);

        assertEq(streamer.totalRewardsSupply(), 500e18, "Total rewards supply mismatch");
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 500e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.warp(initialTime + 10 days);

        assertEq(streamer.totalRewardsSupply(), 1000e18, "Total rewards supply mismatch");
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 250e18);
    }
}

contract MultipleVaultsStakeTest is RewardsStreamerMPTest {
    StakeVault public vault1;
    StakeVault public vault2;
    StakeVault public vault3;

    function setUp() public override {
        super.setUp();

        vault1 = _createTestVault(alice);
        vault2 = _createTestVault(alice);
        vault3 = _createTestVault(alice);

        vm.startPrank(alice);
        stakingToken.approve(address(vault1), 10_000e18);
        stakingToken.approve(address(vault2), 10_000e18);
        stakingToken.approve(address(vault3), 10_000e18);
        vm.stopPrank();
    }

    function _stakeWithVault(address account, StakeVault vault, uint256 amount, uint256 lockupTime) public {
        vm.prank(account);
        vault.stake(amount, lockupTime);
    }

    function test_StakeMultipleVaults() public {
        // Alice vault1 stakes 10 tokens
        _stakeWithVault(alice, vault1, 10e18, 0);

        // Alice vault2 stakes 20 tokens
        _stakeWithVault(alice, vault2, 20e18, 0);

        // Alice vault3 stakes 30 tokens
        _stakeWithVault(alice, vault3, 60e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 90e18,
                totalMPAccrued: 90e18,
                totalMaxMP: 450e18,
                stakingBalance: 90e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkUserTotals(
            CheckUserTotalsParams({ user: alice, totalStakedBalance: 90e18, totalMPAccrued: 90e18, totalMaxMP: 450e18 })
        );
    }
}
