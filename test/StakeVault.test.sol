// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";

import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { DeployStakeManagerScript } from "../script/DeployStakeManager.s.sol";
import { VaultFactory } from "../src/VaultFactory.sol";
import { StakeManager } from "../src/StakeManager.sol";
import { StakeVault } from "../src/StakeVault.sol";
import { MockToken } from "./mocks/MockToken.sol";

contract StakeVaultTest is Test {
    VaultFactory internal vaultFactory;
    StakeManager internal streamer;
    StakeVault internal stakeVault;
    MockToken internal rewardToken;
    MockToken internal stakingToken;
    MockToken internal otherToken;
    address internal alice = makeAddr("alice");
    address internal bob = makeAddr("bob");

    function _createTestVault(address owner) internal returns (StakeVault stakeVault) {
        vm.prank(owner);
        stakeVault = vaultFactory.createVault();
    }

    function setUp() public virtual {
        rewardToken = new MockToken("Reward Token", "RT");
        stakingToken = new MockToken("Staking Token", "ST");
        otherToken = new MockToken("Other Token", "OT");

        DeployStakeManagerScript deployment = new DeployStakeManagerScript();
        (StakeManager stakeManager, VaultFactory _vaultFactory, DeploymentConfig deploymentConfig) = deployment.run();
        (, address _stakingToken) = deploymentConfig.activeNetworkConfig();

        streamer = stakeManager;
        stakingToken = MockToken(_stakingToken);
        vaultFactory = _vaultFactory;

        stakingToken.mint(alice, 10_000e18);

        stakeVault = _createTestVault(alice);

        vm.prank(alice);
        stakingToken.approve(address(stakeVault), 10_000e18);
    }

    function testOwner() public view {
        assertEq(stakeVault.owner(), alice);
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

contract StakeVaultCoverageTest is StakeVaultTest {
    /*////////////////////////////////////////////////////////////
                        TESTES PARA stake()
    ////////////////////////////////////////////////////////////*/

    function test_StakeTransfersTokensToVault() public {
        vm.prank(alice);
        stakeVault.stake(1e18, 90 days);
        assertEq(stakingToken.balanceOf(address(stakeVault)), 1e18);
        assertEq(streamer.stakedBalanceOf(address(stakeVault)), 1e18);
    }

    function test_StakeRevertsIfNotOwner() public {
        vm.prank(bob);
        vm.expectRevert("Ownable: caller is not the owner");
        stakeVault.stake(1e18, 90 days);
    }

    function test_StakeRevertsIfManagerNotTrusted() public {
        vm.prank(alice);
        stakeVault.trustStakeManager(address(0xDEAD));
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__StakeManagerImplementationNotTrusted.selector);
        stakeVault.stake(1e18, 3600);
    }

    /*////////////////////////////////////////////////////////////
                           TESTES PARA lock()
    ////////////////////////////////////////////////////////////*/

    function test_LockSetsLockUntilTimestamp() public {
        uint256 delta = 90 days;
        vm.startPrank(alice);
        stakeVault.stake(1e18, 0); // Stake some tokens firstq
        stakeVault.lock(delta);
        uint256 expected = block.timestamp + delta;
        assertEq(stakeVault.lockUntil(), expected);
    }

    function test_LockRevertsIfManagerNotTrusted() public {
        vm.prank(alice);
        stakeVault.trustStakeManager(address(0xBEEF));
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__StakeManagerImplementationNotTrusted.selector);
        stakeVault.lock(3600);
    }

    /*////////////////////////////////////////////////////////////
                       TESTES PARA unstake()
    ////////////////////////////////////////////////////////////*/

    function test_UnstakeTransfersTokensBackToOwner() public {
        uint256 startBalance = stakingToken.balanceOf(alice);
        vm.prank(alice);
        stakeVault.stake(5e18, 0);
        vm.prank(alice);
        stakeVault.unstake(5e18);
        assertEq(stakingToken.balanceOf(alice), startBalance);
    }

    function test_UnstakeRevertsWithInvalidDestination() public {
        vm.prank(alice);
        stakeVault.stake(1e18, 0);
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__InvalidDestinationAddress.selector);
        stakeVault.unstake(1e18, address(0));
    }

    /*////////////////////////////////////////////////////////////
                          TESTES PARA leave()
    ////////////////////////////////////////////////////////////*/

    function test_LeaveRevertsWhenManagerTrusted() public {
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__NotAllowedToLeave.selector);
        stakeVault.leave(alice);
    }

    function test_LeaveTransfersAllFundsAfterUntrustingManager() public {
        vm.prank(alice);
        stakeVault.stake(2e18, 0);
        vm.prank(alice);
        stakeVault.trustStakeManager(address(1));
        vm.prank(alice);
        stakeVault.leave(bob);
        assertEq(stakingToken.balanceOf(bob), 2e18);
    }

    /*////////////////////////////////////////////////////////////
                       TESTES PARA withdraw()
    ////////////////////////////////////////////////////////////*/

    function test_WithdrawOtherTokenTransfersToDestination() public {
        otherToken.mint(address(stakeVault), 1e18);
        vm.prank(alice);
        stakeVault.withdraw(otherToken, 1e18, bob);
        assertEq(otherToken.balanceOf(bob), 1e18);
    }

    function test_WithdrawRevertsIfInsufficientAvailableBalance() public {
        vm.prank(alice);
        stakeVault.stake(3e18, 0);
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__NotEnoughAvailableBalance.selector);
        stakeVault.withdraw(stakingToken, 3e18);
    }

    function test_WithdrawTransfersGenericTokenToOwner() public {
        otherToken.mint(address(stakeVault), 5e17);
        vm.prank(alice);
        stakeVault.withdraw(otherToken, 5e17);
        assertEq(otherToken.balanceOf(alice), 5e17);
    }

    function test_WithdrawRevertsIfInvalidDestination() public {
        otherToken.mint(address(stakeVault), 1e18);
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__InvalidDestinationAddress.selector);
        stakeVault.withdraw(otherToken, 1e18, address(0));
    }
}
