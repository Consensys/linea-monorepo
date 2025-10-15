pragma solidity ^0.8.26;

import { StakeManagerTest, StakeVault } from "../../StakeManagerBase.t.sol";

contract WithdrawTest is StakeManagerTest {
    function setUp() public override {
        super.setUp();
    }

    function test_RewertWhen_WithdrawingStakedFundsWithoutCallingLeaveFirst() public {
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

        // Alice tries to withdraw staked funds, use `unstake()` instead
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__MustLeaveFirst.selector);
        StakeVault(vaults[alice]).withdraw(stakingToken, 100e18);
    }

    function test_WithdrawStakedTokensAfterLeave() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        _stake(alice, 100e18, MIN_LOCKUP_PERIOD);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance - 100e18, "Alice should have staked tokens");

        // Alice tries to withdraw staked funds, use `unstake()` instead
        vm.prank(alice);
        StakeVault(vaults[alice]).leave(alice);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance - 100e18, "Alice's tokens are still deposited");

        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__FundsLocked.selector);
        StakeVault(vaults[alice]).withdraw(stakingToken, 100e18);

        vm.warp(block.timestamp + MIN_LOCKUP_PERIOD);

        vm.prank(alice);
        StakeVault(vaults[alice]).withdraw(stakingToken, 100e18);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice has all her funds back");
    }
}