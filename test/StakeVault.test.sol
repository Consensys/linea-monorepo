// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";

import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { DeployRewardsStreamerMPScript } from "../script/DeployRewardsStreamerMP.s.sol";
import { VaultFactory } from "../src/VaultFactory.sol";
import { RewardsStreamerMP } from "../src/RewardsStreamerMP.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { MockToken } from "./mocks/MockToken.sol";

contract StakeVaultTest is Test {
    VaultFactory internal vaultFactory;

    RewardsStreamerMP internal streamer;

    StakeVault internal stakeVault;

    MockToken internal rewardToken;

    MockToken internal stakingToken;

    address internal alice = makeAddr("alice");

    function _createTestVault(address owner) internal returns (StakeVault vault) {
        vm.prank(owner);
        vault = vaultFactory.createVault();
    }

    function setUp() public virtual {
        rewardToken = new MockToken("Reward Token", "RT");
        stakingToken = new MockToken("Staking Token", "ST");

        DeployRewardsStreamerMPScript deployment = new DeployRewardsStreamerMPScript();
        (RewardsStreamerMP stakeManager, VaultFactory _vaultFactory, DeploymentConfig deploymentConfig) =
            deployment.run();
        (, address _stakingToken,) = deploymentConfig.activeNetworkConfig();

        streamer = stakeManager;
        stakingToken = MockToken(_stakingToken);
        vaultFactory = _vaultFactory;

        stakingToken.mint(alice, 10_000e18);

        stakeVault = _createTestVault(alice);

        vm.prank(alice);
        stakingToken.approve(address(stakeVault), 10_000e18);
    }
}

contract StakingTokenTest is StakeVaultTest {
    function setUp() public override {
        super.setUp();
    }

    function testStakeToken() public view {
        assertEq(address(stakeVault.STAKING_TOKEN()), address(stakingToken));
    }
}

contract WithdrawTest is StakeVaultTest {
    function setUp() public override {
        super.setUp();
    }

    function test_CannotWithdrawStakedFunds() public {
        // first, stake some funds
        vm.prank(alice);
        stakeVault.stake(10e18, 0);

        assertEq(stakingToken.balanceOf(address(stakeVault)), 10e18);
        assertEq(streamer.totalStaked(), 10e18);

        // try withdrawing funds without unstaking
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__NotEnoughAvailableBalance.selector);
        stakeVault.withdraw(stakingToken, 10e18);
    }
}
