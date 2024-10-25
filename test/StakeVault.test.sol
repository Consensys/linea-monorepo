// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";

import { IStakeManagerProxy } from "../src/interfaces/IStakeManagerProxy.sol";
import { StakeManagerProxy } from "../src/StakeManagerProxy.sol";
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
        vault = new StakeVault(owner, IStakeManagerProxy(address(streamer)));
        vault.register();
    }

    function setUp() public virtual {
        rewardToken = new MockToken("Reward Token", "RT");
        stakingToken = new MockToken("Staking Token", "ST");
        address impl = address(new RewardsStreamerMP());
        bytes memory initializeData = abi.encodeWithSelector(
            RewardsStreamerMP.initialize.selector, address(this), address(stakingToken), address(rewardToken)
        );
        address proxy = address(new StakeManagerProxy(impl, initializeData));
        streamer = RewardsStreamerMP(proxy);

        stakingToken.mint(alice, 10_000e18);

        // Create a temporary vault just to get the codehash
        StakeVault tempVault = new StakeVault(address(this), IStakeManagerProxy(address(streamer)));
        bytes32 vaultCodeHash = address(tempVault).codehash;

        // Register the codehash before creating any user vaults
        streamer.setTrustedCodehash(vaultCodeHash, true);

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
