// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { DeployKarmaScript } from "../../script/DeployKarma.s.sol";
import { DeployStakeManagerScript } from "../../script/DeployStakeManager.s.sol";
import { DeployVaultFactoryScript } from "../../script/DeployVaultFactory.s.sol";
import { UpgradeStakeManagerScript } from "../../script/UpgradeStakeManager.s.sol";
import { DeploymentConfig } from "../../script/DeploymentConfig.s.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { Clones } from "@openzeppelin/contracts/proxy/Clones.sol";
import { IStakeManager } from "../../src/interfaces/IStakeManager.sol";
import { ITrustedCodehashAccess } from "../../src/interfaces/ITrustedCodehashAccess.sol";
import { StakeManager } from "../../src/StakeManager.sol";
import { StakeMath } from "../../src/math/StakeMath.sol";
import { StakeVault } from "../../src/StakeVault.sol";
import { VaultFactory } from "../../src/VaultFactory.sol";
import { Karma } from "../../src/Karma.sol";
import { MockToken } from "../mocks/MockToken.sol";
import { StackOverflowStakeManager } from "../mocks/StackOverflowStakeManager.sol";

contract StakeManagerTest is StakeMath, Test {
    MockToken internal stakingToken;
    StakeManager public streamer;
    VaultFactory public vaultFactory;
    Karma public karma;

    address internal admin;
    address internal alice = makeAddr("alice");
    address internal bob = makeAddr("bob");
    address internal charlie = makeAddr("charlie");
    address internal dave = makeAddr("dave");
    address internal guardian = makeAddr("guardian");

    mapping(address owner => address vault) public vaults;

    function setUp() public virtual {
        DeployStakeManagerScript deployment = new DeployStakeManagerScript();
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        DeployVaultFactoryScript vaultFactoryDeployment = new DeployVaultFactoryScript();

        (karma,) = karmaDeployment.runForTest();
        (StakeManager stakeManager, DeploymentConfig deploymentConfig) = deployment.runForTest(address(karma));
        (address _deployer, address _stakingToken) = deploymentConfig.activeNetworkConfig();
        (VaultFactory _vaultFactory,, address vaultProxyClone,) =
            vaultFactoryDeployment.runForTest(address(stakeManager), _stakingToken);

        streamer = stakeManager;
        stakingToken = MockToken(_stakingToken);
        vaultFactory = _vaultFactory;
        admin = _deployer;

        // set up reward distribution
        vm.startPrank(admin);
        karma.addRewardDistributor(address(streamer));
        karma.setAllowedToTransfer(address(streamer), true);
        streamer.setRewardsSupplier(address(karma));
        streamer.grantRole(streamer.GUARDIAN_ROLE(), address(guardian));
        streamer.setTrustedCodehash(vaultProxyClone.codehash, true);
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

    function checkStreamer(string memory text, CheckStreamerParams memory p) public view {
        assertEq(streamer.totalStaked(), p.totalStaked, string(abi.encodePacked(text, "wrong total staked")));
        assertEq(streamer.totalMPStaked(), p.totalMPStaked, string(abi.encodePacked(text, "wrong total staked MP")));
        assertEq(streamer.totalMPAccrued(), p.totalMPAccrued, string(abi.encodePacked(text, "wrong total accrued MP")));
        assertEq(streamer.totalMaxMP(), p.totalMaxMP, string(abi.encodePacked(text, "wrong totalMaxMP MP")));
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
        uint256 rewardsAccrued;
    }

    function checkVault(CheckVaultParams memory p) public view {
        StakeManager.VaultData memory vaultData = streamer.getVault(p.account);

        assertEq(vaultData.stakedBalance, p.stakedBalance, "wrong account staked balance");
        assertEq(stakingToken.balanceOf(p.account), p.vaultBalance, "wrong vault balance");
        assertEq(vaultData.mpAccrued, p.mpAccrued, "wrong account MP accrued");
        assertEq(vaultData.maxMP, p.maxMP, "wrong account max MP");
        assertEq(vaultData.rewardsAccrued, p.rewardsAccrued, "wrong account rewards accrued");
    }

    function checkVault(string memory text, CheckVaultParams memory p) public view {
        // assertEq(rewardToken.balanceOf(p.account), p.rewardBalance, "wrong account reward balance");

        StakeManager.VaultData memory vaultData = streamer.getVault(p.account);

        assertEq(
            vaultData.stakedBalance, p.stakedBalance, string(abi.encodePacked(text, "wrong account staked balance"))
        );
        assertEq(
            stakingToken.balanceOf(p.account), p.vaultBalance, string(abi.encodePacked(text, "wrong vault balance"))
        );
        // assertEq(vaultData.accountRewardIndex, p.rewardIndex, "wrong account reward index");
        assertEq(vaultData.mpAccrued, p.mpAccrued, string(abi.encodePacked(text, "wrong account MP accrued")));
        assertEq(vaultData.maxMP, p.maxMP, string(abi.encodePacked(text, "wrong account max MP")));
        assertEq(
            vaultData.rewardsAccrued, p.rewardsAccrued, string(abi.encodePacked(text, "wrong account rewards accrued"))
        );
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

    function _updateVault(address account) public {
        StakeVault vault = StakeVault(vaults[account]);
        streamer.updateVault(address(vault));
    }

    function _unstake(address account, uint256 amount) public {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.unstake(amount);
    }

    function _lock(address account, uint256 lockPeriod) internal {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        vault.lock(lockPeriod);
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

    function _upgradeStakeManager() internal {
        UpgradeStakeManagerScript upgrade = new UpgradeStakeManagerScript();
        upgrade.runWithAdminAndProxy(admin, address(streamer));
    }

    function _setRewards(uint256 amount, uint256 period) internal {
        vm.prank(admin);
        karma.setReward(address(streamer), amount, period);
    }
    
    function _timeToAccrueMPLimit(uint256 amount) internal view returns (uint256) {
        uint256 maxMP = amount * streamer.MAX_MULTIPLIER();
        uint256 timeInSeconds = _timeToAccrueMP(amount, maxMP);
        return timeInSeconds;
    }

}


