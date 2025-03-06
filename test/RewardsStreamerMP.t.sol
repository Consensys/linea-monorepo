// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { DeployKarmaScript } from "../script/DeployKarma.s.sol";
import { DeployRewardsStreamerMPScript } from "../script/DeployRewardsStreamerMP.s.sol";
import { UpgradeRewardsStreamerMPScript } from "../script/UpgradeRewardsStreamerMP.s.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { Clones } from "@openzeppelin/contracts/proxy/Clones.sol";
import { IStakeManager } from "../src/interfaces/IStakeManager.sol";
import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";
import { ITrustedCodehashAccess } from "../src/interfaces/ITrustedCodehashAccess.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { StakeMath } from "../src/math/StakeMath.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { VaultFactory } from "../src/VaultFactory.sol";
import { Karma } from "../src/Karma.sol";
import { MockToken } from "./mocks/MockToken.sol";
import { StackOverflowStakeManager } from "./mocks/StackOverflowStakeManager.sol";

contract RewardsStreamerMPTest is StakeMath, Test {
    MockToken internal stakingToken;
    RewardsStreamerMP public streamer;
    VaultFactory public vaultFactory;
    Karma public karma;

    address internal admin;
    address internal alice = makeAddr("alice");
    address internal bob = makeAddr("bob");
    address internal charlie = makeAddr("charlie");
    address internal dave = makeAddr("dave");

    mapping(address owner => address vault) public vaults;

    function setUp() public virtual {
        DeployRewardsStreamerMPScript deployment = new DeployRewardsStreamerMPScript();
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        (RewardsStreamerMP stakeManager, VaultFactory _vaultFactory, DeploymentConfig deploymentConfig) =
            deployment.run();

        (address _deployer, address _stakingToken,) = deploymentConfig.activeNetworkConfig();

        streamer = stakeManager;
        stakingToken = MockToken(_stakingToken);
        vaultFactory = _vaultFactory;
        admin = _deployer;
        (karma,) = karmaDeployment.run();

        // set up reward distribution
        vm.startPrank(admin);
        karma.addRewardDistributor(address(streamer));
        streamer.setRewardsSupplier(address(karma));
        vm.stopPrank();

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
        uint256 totalMPStaked;
        uint256 totalMPAccrued;
        uint256 totalMaxMP;
        uint256 stakingBalance;
        uint256 rewardBalance;
        uint256 rewardIndex;
    }

    function checkStreamer(CheckStreamerParams memory p) public view {
        assertEq(streamer.totalStaked(), p.totalStaked, "wrong total staked");
        assertEq(streamer.totalMPStaked(), p.totalMPStaked, "wrong total staked MP");
        assertEq(streamer.totalMPAccrued(), p.totalMPAccrued, "wrong total accrued MP");
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
        uint256 mpStaked;
        uint256 mpAccrued;
        uint256 maxMP;
        uint256 rewardsAccrued;
    }

    function checkVault(CheckVaultParams memory p) public view {
        // assertEq(rewardToken.balanceOf(p.account), p.rewardBalance, "wrong account reward balance");

        RewardsStreamerMP.VaultData memory vaultData = streamer.getVault(p.account);

        assertEq(vaultData.stakedBalance, p.stakedBalance, "wrong account staked balance");
        assertEq(stakingToken.balanceOf(p.account), p.vaultBalance, "wrong vault balance");
        // assertEq(vaultData.accountRewardIndex, p.rewardIndex, "wrong account reward index");
        assertEq(vaultData.mpStaked, p.mpStaked, "wrong account MP staked");
        assertEq(vaultData.mpAccrued, p.mpAccrued, "wrong account MP accrued");
        assertEq(vaultData.maxMP, p.maxMP, "wrong account max MP");
        assertEq(vaultData.rewardsAccrued, p.rewardsAccrued, "wrong account rewards accrued");
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

    function _stake(address account, uint256 amount, uint256 lockupTime) public virtual {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.stake(amount, lockupTime);
    }

    function _compound(address account) public {
        StakeVault vault = StakeVault(vaults[account]);
        streamer.compound(address(vault));
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
        upgrade.runWithAdminAndProxy(admin, IStakeManagerProxy(address(streamer)));
    }

    function _setRewards(uint256 amount, uint256 period) internal {
        vm.prank(admin);
        karma.setReward(address(streamer), amount, period);
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
                totalMPStaked: 0,
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
                totalMPStaked: 10e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        // T2
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        // T3
        vm.prank(admin);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        // T4
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
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
                totalMPStaked: 30e18,
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
                mpStaked: 0e18,
                mpAccrued: 0e18,
                maxMP: 0e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        // T5
        _stake(charlie, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMPStaked: 60e18,
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
                mpStaked: 0e18,
                mpAccrued: 0e18,
                maxMP: 0e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        // T6
        vm.prank(admin);
        streamer.updateGlobalState();

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 60e18,
                totalMPStaked: 60e18,
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
                mpStaked: 0e18,
                mpAccrued: 0e18,
                maxMP: 0e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        //T7
        _unstake(bob, 30e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMPStaked: 30e18,
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
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
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
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[charlie],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 10e18,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
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
                totalMPStaked: 10e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );
    }

    function test_StakeOneAccountAndRewards() public {
        _stake(alice, 10e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
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
                totalMPStaked: stakeAmount + expectedBonusMP,
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
                totalMPStaked: stakeAmount + expectedBonusMP,
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
                totalMPStaked: stakeAmount + expectedBonusMP,
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
                totalMPStaked: stakeAmount,
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
        uint256 totalMPStaked = totalMPAccrued;
        totalMPAccrued = totalMPAccrued + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount,
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
                mpStaked: totalMPStaked,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP,
                rewardsAccrued: 0
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
                totalMPStaked: stakeAmount,
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
                mpStaked: totalMPStaked,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP,
                rewardsAccrued: 0
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
                totalMPStaked: stakeAmount,
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
                mpStaked: stakeAmount,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP, // maxMP == totalMaxMP because only one account is staking
                rewardsAccrued: 0
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
                totalMPStaked: stakeAmount,
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
                mpStaked: stakeAmount,
                mpAccrued: totalMaxMP,
                maxMP: totalMaxMP,
                rewardsAccrued: 0
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
                totalMPStaked: stakeAmount,
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
                totalMPStaked: 40e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
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
                totalMPStaked: 40e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
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
                totalMPStaked: sumOfStakeAmount + sumOfExpectedBonusMP,
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
                totalMPStaked: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMPAccrued: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    struct TestParams {
        uint256 aliceStakeAmount;
        uint256 bobStakeAmount;
        uint256 totalStaked;
        uint256 totalMPAccrued;
        uint256 totalMaxMP;
    }

    function test_StakeMultipleAccountsMPIncreasesMaxMPDoesNotChange() public {
        TestParams memory params;
        params.aliceStakeAmount = 15e18;
        params.bobStakeAmount = 5e18;
        params.totalStaked = params.aliceStakeAmount + params.bobStakeAmount;
        params.totalMPAccrued = params.totalStaked;
        params.totalMaxMP = (params.aliceStakeAmount * streamer.MAX_MULTIPLIER() + params.aliceStakeAmount)
            + (params.bobStakeAmount * streamer.MAX_MULTIPLIER() + params.bobStakeAmount);

        uint256 aliceMP = params.aliceStakeAmount;
        uint256 aliceMaxMP = params.aliceStakeAmount * streamer.MAX_MULTIPLIER() + aliceMP;

        uint256 bobMP = params.bobStakeAmount;
        uint256 bobMaxMP = params.bobStakeAmount * streamer.MAX_MULTIPLIER() + bobMP;
        _stake(alice, params.aliceStakeAmount, 0);
        _stake(bob, params.bobStakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: params.totalStaked,
                totalMPStaked: params.totalStaked,
                totalMPAccrued: params.totalMPAccrued,
                totalMaxMP: params.totalMaxMP,
                stakingBalance: params.totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: params.aliceStakeAmount,
                vaultBalance: params.aliceStakeAmount,
                rewardIndex: 0,
                mpStaked: aliceMP,
                mpAccrued: aliceMP,
                maxMP: aliceMaxMP,
                rewardsAccrued: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: params.bobStakeAmount,
                vaultBalance: params.bobStakeAmount,
                rewardIndex: 0,
                mpStaked: bobMP,
                mpAccrued: bobMP,
                maxMP: bobMaxMP,
                rewardsAccrued: 0
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);
        streamer.updateVaultMP(vaults[bob]);

        uint256 aliceExpectedMPIncrease = params.aliceStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 bobExpectedMPIncrease = params.bobStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        uint256 aliceMPAccrued = aliceMP + aliceExpectedMPIncrease;
        uint256 bobMPAccrued = bobMP + bobExpectedMPIncrease;
        params.totalMPAccrued = params.totalMPAccrued + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: params.totalStaked,
                totalMPStaked: params.totalStaked,
                totalMPAccrued: params.totalMPAccrued,
                totalMaxMP: params.totalMaxMP,
                stakingBalance: params.totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: params.aliceStakeAmount,
                vaultBalance: params.aliceStakeAmount,
                rewardIndex: 0,
                mpStaked: aliceMP,
                mpAccrued: aliceMP + aliceExpectedMPIncrease,
                maxMP: aliceMaxMP,
                rewardsAccrued: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: params.bobStakeAmount,
                vaultBalance: params.bobStakeAmount,
                rewardIndex: 0,
                mpStaked: bobMP,
                mpAccrued: bobMPAccrued,
                maxMP: bobMaxMP,
                rewardsAccrued: 0
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);
        streamer.updateVaultMP(vaults[bob]);

        aliceExpectedMPIncrease = params.aliceStakeAmount / 2;
        bobExpectedMPIncrease = params.bobStakeAmount / 2;
        totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        aliceMPAccrued = aliceMPAccrued + aliceExpectedMPIncrease;
        bobMPAccrued = bobMPAccrued + bobExpectedMPIncrease;
        params.totalMPAccrued = params.totalMPAccrued + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: params.totalStaked,
                totalMPStaked: params.totalStaked,
                totalMPAccrued: params.totalMPAccrued,
                totalMaxMP: params.totalMaxMP,
                stakingBalance: params.totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: params.aliceStakeAmount,
                vaultBalance: params.aliceStakeAmount,
                rewardIndex: 0,
                mpStaked: aliceMP,
                mpAccrued: aliceMPAccrued,
                maxMP: aliceMaxMP,
                rewardsAccrued: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: params.bobStakeAmount,
                vaultBalance: params.bobStakeAmount,
                rewardIndex: 0,
                mpStaked: bobMP,
                mpAccrued: bobMPAccrued,
                maxMP: bobMaxMP,
                rewardsAccrued: 0
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
                totalMPStaked: 2e18,
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
                mpStaked: 2e18,
                mpAccrued: 2e18,
                maxMP: 10e18,
                rewardsAccrued: 0
            })
        );

        _unstake(alice, 2e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMPStaked: 0,
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
                totalMPStaked: 10e18,
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
                totalMPStaked: 10e18,
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
                totalMPStaked: (stakeAmount + expectedBonusMP),
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
                totalMPStaked: expectedTotalMP,
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
                totalMPStaked: 2e18,
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
                mpStaked: 2e18,
                mpAccrued: 2e18,
                maxMP: 10e18,
                rewardsAccrued: 0
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
                totalMPStaked: 20e18,
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
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 0,
                mpStaked: 20e18,
                mpAccrued: 20e18,
                maxMP: 100e18,
                rewardsAccrued: 0
            })
        );
    }

    function test_UnstakeMultipleAccountsAndRewards() public {
        test_StakeMultipleAccountsAndRewards();

        _unstake(alice, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMPStaked: 30e18,
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
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMPStaked: 20e18,
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
                mpStaked: 20e18,
                mpAccrued: 20e18,
                maxMP: 100e18,
                rewardsAccrued: 0
            })
        );

        _unstake(bob, 20e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMPStaked: 0,
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
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
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

    function test_LockWithPriorLock() public {
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
                mpStaked: initialAccountMP,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
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
                mpStaked: initialAccountMP + expectedBonusMP,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );

        expectedBonusMP = _bonusMP(stakeAmount, lockPeriod * 2);
        // Lock for more one 1 year
        _lock(alice, lockPeriod);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpStaked: initialAccountMP + expectedBonusMP,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );
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
                mpStaked: initialAccountMP,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
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
                mpStaked: initialAccountMP + expectedBonusMP,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );
    }

    function test_LockMultipleTimesExceedMaxLock() public {
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
                mpStaked: initialAccountMP,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        // Lock for 1 year
        uint256 lockPeriod = 4 * YEAR;
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
                mpStaked: initialAccountMP + expectedBonusMP,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );

        // wait for lock year to be over
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (4 * YEAR));

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpStaked: initialAccountMP + expectedBonusMP,
                mpAccrued: initialAccountMP + expectedBonusMP + (initialAccountMP * 4),
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );

        // lock for another year should fail as 4 years is the maximum of total lock time
        vm.expectRevert(StakeMath.StakeMath__AbsoluteMaxMPOverflow.selector);
        _lock(alice, YEAR);
    }

    function test_LockFailsWithNoStake() public {
        vm.expectRevert(StakeMath.StakeMath__InsufficientBalance.selector);
        _lock(alice, YEAR);
    }

    function test_LockFailsWithZero() public {
        _stake(alice, 10e18, 0);

        // Test with period = 0
        vm.expectRevert(IStakeManager.StakingManager__DurationCannotBeZero.selector);
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

    function test_RevertWhenVaultToLockIsEmpty() public {
        vm.expectRevert(StakeMath.StakeMath__InsufficientBalance.selector);
        _lock(alice, YEAR);
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
        vm.expectRevert("Ownable: caller is not the owner");
        streamer.enableEmergencyMode();
    }

    function test_CannotEnableEmergencyModeTwice() public {
        vm.prank(admin);
        streamer.enableEmergencyMode();

        vm.expectRevert(IStakeManager.StakingManager__EmergencyModeEnabled.selector);
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
                totalMPStaked: 10e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
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
                totalMPStaked: 10e18,
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
                totalMPStaked: 40e18,
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
                totalMPStaked: 40e18,
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpStaked: 30e18,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
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
                mpStaked: 10e18,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
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
        vm.expectRevert("Ownable: caller is not the owner");
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);
    }

    function test_UpgradeStakeManager() public {
        // first, change state of existing stake manager
        _stake(alice, 10e18, 0);

        // check initial state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
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
                totalMPStaked: 10e18,
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
                totalMPStaked: 100e18,
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
                totalMPStaked: 0,
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
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
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
                totalMPStaked: 100e18,
                totalMPAccrued: 100e18,
                totalMaxMP: 500e18,
                stakingBalance: 100e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // upgrade the manager to a malicious one
        address newImpl = address(new StackOverflowStakeManager());
        vm.prank(admin);
        UUPSUpgradeable(streamer).upgradeTo(newImpl);

        // alice leaves system and is able to get funds out, despite malicious manager
        _leave(alice);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get her tokens back");
    }
}

// solhint-disable-next-line
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
        _setRewards(1000, 10);

        assertEq(streamer.rewardStartTime(), currentTime);
        assertEq(streamer.rewardEndTime(), currentTime + 10);
        assertEq(streamer.lastRewardTime(), currentTime);
    }

    function testSetRewards_RevertsNotAuthorized() public {
        vm.prank(alice);
        vm.expectPartialRevert(IStakeManager.StakingManager__Unauthorized.selector);
        streamer.setReward(1000, 10);
    }

    function testSetRewards_RevertsBadDuration() public {
        vm.prank(admin);
        vm.expectRevert(IStakeManager.StakingManager__DurationCannotBeZero.selector);
        karma.setReward(address(streamer), 1000, 0);
    }

    function testSetRewards_RevertsBadAmount() public {
        vm.prank(admin);
        vm.expectRevert(IStakeManager.StakingManager__AmountCannotBeZero.selector);
        karma.setReward(address(streamer), 0, 10);
    }

    function testTotalRewardsSupply() public {
        _stake(alice, 100e18, 0);
        assertEq(streamer.totalRewardsSupply(), 0);

        uint256 initialTime = vm.getBlockTimestamp();

        _setRewards(1000e18, 10 days);
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
        _setRewards(2000e18, 10 days);

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
        uint256 year = 365 days;
        uint256 initialTime = vm.getBlockTimestamp();

        _stake(alice, 100e18, 0);
        _setRewards(1000e18, year);

        assertEq(streamer.totalStaked(), 100e18);
        assertEq(streamer.totalMPStaked(), 100e18);
        assertEq(streamer.totalShares(), 200e18);
        assertEq(streamer.totalRewardsSupply(), 0);
        assertEq(streamer.totalMP(), 100e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 100e18);
        assertEq(streamer.mpStakedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 0);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 0);
        assertEq(streamer.mpStakedOf(vaults[bob]), 0);
        assertEq(streamer.mpStakedOf(vaults[bob]), 0);
        assertEq(streamer.vaultShares(vaults[bob]), 0);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.warp(initialTime + year / 2);
        _stake(bob, 100e18, 0);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 200e18);
        assertEq(streamer.totalShares(), 400e18);
        assertEq(streamer.totalRewardsSupply(), 500e18);
        // totalMP: 200 + 50 accrued by Alice (not stake yet)
        assertEq(streamer.totalMP(), 250e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 150e18);
        assertEq(streamer.mpStakedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 500e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 100e18);
        assertEq(streamer.mpStakedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.warp(initialTime + year);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 200e18);
        assertEq(streamer.totalShares(), 400e18);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalMP(), 350e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 200e18);
        assertEq(streamer.mpStakedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 150e18);
        assertEq(streamer.mpStakedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 250e18);

        vm.warp(initialTime + year * 2);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 200e18);
        assertEq(streamer.totalShares(), 400e18);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalMP(), 550e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 300e18);
        assertEq(streamer.mpStakedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 250e18);
        assertEq(streamer.mpStakedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 250e18);

        _compound(alice);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 400e18);
        assertEq(streamer.totalShares(), 600e18);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalMP(), 550e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 300e18);
        assertEq(streamer.mpStakedOf(vaults[alice]), 300e18);
        assertEq(streamer.vaultShares(vaults[alice]), 400e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 250e18);
        assertEq(streamer.mpStakedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 250e18);

        _setRewards(600e18, year);

        vm.warp(initialTime + year * 3);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 400e18);
        assertEq(streamer.totalShares(), 600e18);
        assertEq(streamer.totalRewardsSupply(), 1600e18);
        assertEq(streamer.totalMP(), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 400e18);
        assertEq(streamer.mpStakedOf(vaults[alice]), 300e18);
        assertEq(streamer.vaultShares(vaults[alice]), 400e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 1150e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 350e18);
        assertEq(streamer.mpStakedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 450e18);
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
        console.log(MAX_BALANCE);

        // Alice vault1 stakes 10 tokens
        _stakeWithVault(alice, vault1, 10e18, 0);

        // Alice vault2 stakes 20 tokens
        _stakeWithVault(alice, vault2, 20e18, 0);

        // Alice vault3 stakes 30 tokens
        _stakeWithVault(alice, vault3, 60e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 90e18,
                totalMPStaked: 90e18,
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

contract StakeVaultMigrationTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_RevertWhenNotOwnerOfMigrationVault() public {
        // alice tries to migrate to a vault she doesn't own
        vm.prank(alice);
        vm.expectRevert(IStakeManager.StakingManager__Unauthorized.selector);
        StakeVault(vaults[alice]).migrateToVault(vaults[bob]);
    }

    function test_RevertWhenMigrationVaultNotEmpty() public {
        // alice creates new vault
        vm.startPrank(alice);
        StakeVault newVault = vaultFactory.createVault();

        // ensure new vault is in use
        stakingToken.approve(address(newVault), 10e18);
        newVault.stake(10e18, 0);

        // alice tries to migrate to a vault that is not empty
        vm.expectRevert(IStakeManager.StakingManager__MigrationTargetHasFunds.selector);
        StakeVault(vaults[alice]).migrateToVault(address(newVault));
    }

    function testMigrateToVault() public {
        uint256 stakeAmount = 100e18;

        uint256 initialAccountMP = stakeAmount;
        uint256 initialMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;

        // first, ensure alice has a vault with staked funds
        _stake(alice, stakeAmount, 0);

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpStaked: initialAccountMP,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: initialAccountMP,
                totalMPAccrued: initialAccountMP,
                totalMaxMP: initialMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // some time passes
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + 365 days);

        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);

        // ensure vault has accumulated MPs
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpStaked: initialAccountMP,
                mpAccrued: initialAccountMP * 2, // alice now has twice the amount after a year
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount,
                totalMPAccrued: initialAccountMP * 2, // stakemanager has twice the amount after a year
                totalMaxMP: initialMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // alice creates new vault
        vm.prank(alice);
        address newVault = address(vaultFactory.createVault());

        // alice migrates to new vault
        vm.prank(alice);
        StakeVault(vaults[alice]).migrateToVault(newVault);

        // ensure stake manager's total stats have not changed
        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: initialAccountMP,
                totalMPAccrued: initialAccountMP * 2,
                totalMaxMP: initialMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // check that alice's funds are now in the new vault
        checkVault(
            CheckVaultParams({
                account: newVault,
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpStaked: initialAccountMP,
                mpAccrued: initialAccountMP * 2, // alice now has twice the amount after a year
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        // check that alice's old vault is empty
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                mpStaked: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );
    }
}

contract CompoundTest is RewardsStreamerMPTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_RevertWhenInsufficientMPBalance() public {
        _stake(alice, 10e18, 0);
        vm.expectRevert(IStakeManager.StakingManager__InsufficientBalance.selector);
        _compound(alice);
    }
}

contract FuzzTests is RewardsStreamerMPTest {
    function _stake(address account, uint256 amount, uint256 lockPeriod) public override {
        stakingToken.mint(account, amount);
        vm.prank(account);
        stakingToken.approve(vaults[account], amount);
        super._stake(account, amount, lockPeriod);
    }

    function _lock(address account, uint256 lockPeriod) internal {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.lock(lockPeriod);
    }

    function testFuzz_Stake(uint256 stakeAmount, uint256 lockUpPeriod) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        vm.assume(lockUpPeriod == 0 || (lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD));
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
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
                mpStaked: stakeAmount + expectedBonusMP,
                mpAccrued: stakeAmount + expectedBonusMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );
    }

    function testFuzz_Lock(uint256 stakeAmount, uint256 lockUpPeriod) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        vm.assume(lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD);
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, 0);
        _lock(alice, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
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
                mpStaked: stakeAmount + expectedBonusMP,
                mpAccrued: stakeAmount + expectedBonusMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );
    }

    function testFuzz_Relock(uint256 stakeAmount, uint256 lockUpPeriod, uint256 lockUpPeriod2) public {
        stakeAmount = bound(stakeAmount, 1, MAX_BALANCE);
        lockUpPeriod = lockUpPeriod == 0 ? 0 : bound(lockUpPeriod, MIN_LOCKUP_PERIOD, MAX_LOCKUP_PERIOD);
        vm.assume(
            lockUpPeriod2 > 0 && lockUpPeriod2 <= MAX_LOCKUP_PERIOD && lockUpPeriod + lockUpPeriod2 >= MIN_LOCKUP_PERIOD
                && lockUpPeriod + lockUpPeriod2 <= MAX_LOCKUP_PERIOD
        );
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
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
                mpStaked: stakeAmount + expectedBonusMP,
                mpAccrued: stakeAmount + expectedBonusMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );

        _lock(alice, lockUpPeriod2);
        expectedBonusMP += _bonusMP(stakeAmount, lockUpPeriod2);
        expectedMaxTotalMP += _bonusMP(stakeAmount, lockUpPeriod2);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
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
                mpStaked: stakeAmount + expectedBonusMP,
                mpAccrued: stakeAmount + expectedBonusMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );
    }

    function testFuzz_AccrueMP(uint256 stakeAmount, uint256 lockUpPeriod, uint16 accruedTime) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        vm.assume(lockUpPeriod == 0 || (lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD));
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);
        uint256 expectedStakedMP = _initialMP(stakeAmount) + _bonusMP(stakeAmount, lockUpPeriod);
        uint256 rawTotalMP = expectedStakedMP + _accrueMP(stakeAmount, accruedTime);
        uint256 expectedTotalMP = Math.min(rawTotalMP, expectedMaxTotalMP);

        _stake(alice, stakeAmount, lockUpPeriod);

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + accruedTime);
        streamer.updateGlobalState();
        streamer.updateVaultMP(vaults[alice]);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: expectedStakedMP,
                totalMPAccrued: expectedTotalMP,
                totalMaxMP: expectedMaxTotalMP,
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
                mpStaked: expectedStakedMP,
                mpAccrued: expectedTotalMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );
    }

    function testFuzz_AccrueMP_Relock(uint256 stakeAmount, uint256 lockUpPeriod2, uint16 accruedTime) public {
        uint256 lockUpPeriod = MIN_LOCKUP_PERIOD;
        vm.assume(lockUpPeriod2 <= MAX_LOCKUP_PERIOD);
        if (accruedTime <= lockUpPeriod) {
            vm.assume(
                lockUpPeriod2 > 0 && lockUpPeriod + lockUpPeriod2 - accruedTime >= MIN_LOCKUP_PERIOD
                    && lockUpPeriod + lockUpPeriod2 - accruedTime <= MAX_LOCKUP_PERIOD
            );
        } else {
            vm.assume(lockUpPeriod2 >= MIN_LOCKUP_PERIOD);
        }
        stakeAmount = bound(stakeAmount, MIN_BALANCE, MAX_BALANCE);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);
        uint256 expectedStakedMP = _initialMP(stakeAmount) + _bonusMP(stakeAmount, lockUpPeriod);
        uint256 rawTotalMP = expectedStakedMP + _accrueMP(stakeAmount, accruedTime);
        uint256 expectedTotalMP = Math.min(rawTotalMP, expectedMaxTotalMP);

        _stake(alice, stakeAmount, lockUpPeriod);

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + accruedTime);
        _lock(alice, lockUpPeriod2);
        expectedTotalMP += _bonusMP(stakeAmount, lockUpPeriod2);
        expectedStakedMP += _bonusMP(stakeAmount, lockUpPeriod2);
        expectedMaxTotalMP += _bonusMP(stakeAmount, lockUpPeriod2);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: expectedStakedMP,
                totalMPAccrued: expectedTotalMP,
                totalMaxMP: expectedMaxTotalMP,
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
                mpStaked: expectedStakedMP,
                mpAccrued: expectedTotalMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );
    }

    function testFuzz_Unstake(
        uint256 stakeAmount,
        uint256 lockUpPeriod,
        uint16 accruedTime,
        uint256 unstakeAmount
    )
        public
    {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        vm.assume(lockUpPeriod == 0 || (lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD));
        vm.assume(unstakeAmount <= stakeAmount);
        vm.assume(accruedTime >= lockUpPeriod);

        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);
        uint256 expectedStakedMP = _initialMP(stakeAmount) + _bonusMP(stakeAmount, lockUpPeriod);
        uint256 rawTotalMP = expectedStakedMP + _accrueMP(stakeAmount, accruedTime);
        uint256 expectedTotalMP = Math.min(rawTotalMP, expectedMaxTotalMP);

        _stake(alice, stakeAmount, lockUpPeriod);

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + accruedTime);

        _unstake(alice, unstakeAmount);

        uint256 totalMPAccrued = expectedTotalMP - _reduceMP(stakeAmount, expectedTotalMP, unstakeAmount);
        if (totalMPAccrued < expectedStakedMP) {
            expectedStakedMP = totalMPAccrued;
        }

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount - unstakeAmount,
                totalMPStaked: expectedStakedMP,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: expectedMaxTotalMP - _reduceMP(stakeAmount, expectedMaxTotalMP, unstakeAmount),
                stakingBalance: stakeAmount - unstakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount - unstakeAmount,
                vaultBalance: stakeAmount - unstakeAmount,
                rewardIndex: 0,
                mpStaked: expectedStakedMP,
                mpAccrued: expectedTotalMP - _reduceMP(stakeAmount, expectedTotalMP, unstakeAmount),
                maxMP: expectedMaxTotalMP - _reduceMP(stakeAmount, expectedMaxTotalMP, unstakeAmount),
                rewardsAccrued: 0
            })
        );
    }

    function testFuzz_Rewards(
        uint256 stakeAmount,
        uint256 lockUpPeriod,
        uint256 rewardAmount,
        uint16 rewardPeriod,
        uint16 accountRewardPeriod
    )
        public
    {
        stakeAmount = bound(stakeAmount, 1e18, 20_000_000e18);
        lockUpPeriod = lockUpPeriod == 0 ? 0 : bound(lockUpPeriod, MIN_LOCKUP_PERIOD, MAX_LOCKUP_PERIOD);
        vm.assume(rewardPeriod > 0 && rewardPeriod <= 12 weeks); // assuming max 3 months
        vm.assume(rewardAmount > 1e18 && rewardAmount <= 1_000_000e18); // assuming max 1_000_000 Karma
        vm.assume(accountRewardPeriod <= rewardPeriod); // Ensure accountRewardPeriod doesn't exceed rewardPeriod

        uint256 initialTime = vm.getBlockTimestamp();
        uint256 tolerance = 1000;

        // Calculate expected reward using safe math operations
        uint256 expectedReward = accountRewardPeriod < rewardPeriod
            ? Math.mulDiv(accountRewardPeriod, rewardAmount, rewardPeriod)
            : rewardAmount;

        _stake(alice, stakeAmount, lockUpPeriod);

        _setRewards(rewardAmount, rewardPeriod);

        vm.warp(initialTime + accountRewardPeriod);

        assertEq(streamer.totalRewardsSupply(), expectedReward, "Total rewards supply mismatch");
        assertApproxEqAbs(
            streamer.rewardsBalanceOf(vaults[alice]), expectedReward, tolerance, "Reward balance mismatch"
        );
    }

    function testFuzz_EmergencyExit(uint256 stakeAmount, uint256 lockUpPeriod) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        vm.assume(lockUpPeriod == 0 || (lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD));

        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: 0,
                rewardIndex: 0,
                mpStaked: stakeAmount + expectedBonusMP,
                mpAccrued: stakeAmount + expectedBonusMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );

        assertEq(
            stakingToken.balanceOf(alice), aliceInitialBalance + stakeAmount, "Alice should get staked tokens back"
        );
    }
}
