pragma solidity ^0.8.26;

import { StakeManagerTest, StakeVault, IERC20 } from "../../StakeManagerBase.t.sol";

contract LeaveTest is StakeManagerTest {
    function setUp() public override {
        super.setUp();
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
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice has all her funds back");
    }

    function test_LeaveShouldRedeemRewardsToActualKarma() public {
        vm.startPrank(admin);
        karma.setReward(address(streamer), 1000e18, 10 days);
        vm.stopPrank();

        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 streamerInitialKarmaBalance = karma.balanceOfRewardDistributor(address(streamer));

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

        vm.warp(block.timestamp + 10 days);

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
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice has all her funds back");
        assertEq(karma.balanceOfRewardDistributor(address(streamer)), 0, "Streamer has no Karma left");
        assertEq(karma.balanceOf(alice), streamerInitialKarmaBalance, "Alice has all the Karma rewards");
    }

    function test_LeaveShouldKeepFundsLockedInStakeVault() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD();

        _stake(alice, stakeAmount, lockUpPeriod);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance - stakeAmount, "Alice should have staked tokens");

        _upgradeStakeManager();
        _leave(alice);

        assertEq(
            stakingToken.balanceOf(alice), aliceInitialBalance - stakeAmount, "Alice still doesn't have her funds back"
        );

        vm.warp(block.timestamp + lockUpPeriod);

        StakeVault vault = StakeVault(vaults[alice]);
        IERC20 token = vault.STAKING_TOKEN();
        vm.prank(alice);
        vault.withdraw(token, stakeAmount);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice has withdrawn her funds");
    }
}