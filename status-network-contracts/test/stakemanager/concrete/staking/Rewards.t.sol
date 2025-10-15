pragma solidity ^0.8.26;

import { StakeManagerTest, IStakeManager } from "../../StakeManagerBase.t.sol";

contract StakeManager_RewardsTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function testSetRewardsAccumulatesRemainingRewards() public {
        // Setup initial reward period
        uint256 initialReward = 1000e18;
        uint256 duration = 10 days;
        uint256 startTime = vm.getBlockTimestamp();

        _setRewards(initialReward, duration);

        // Verify initial setup
        assertEq(streamer.rewardAmount(), initialReward);
        assertEq(streamer.rewardStartTime(), startTime);
        assertEq(streamer.rewardEndTime(), startTime + duration);

        // Advance time to halfway through the reward period
        uint256 halfwayTime = startTime + duration / 2;
        vm.warp(halfwayTime);

        // Set new rewards before the current period ends
        uint256 newRewardAmount = 500e18;
        uint256 newDuration = 5 days;

        _setRewards(newRewardAmount, newDuration);

        // Calculate expected remaining rewards from first period
        // At halfway point, 50% of rewards should remain
        uint256 expectedRemainingRewards = initialReward / 2;
        uint256 expectedTotalRewards = expectedRemainingRewards + newRewardAmount;

        // Verify the reward amount is accumulated correctly
        assertEq(streamer.rewardAmount(), expectedTotalRewards, "Total reward amount should be remaining + new");
        assertEq(streamer.rewardStartTime(), halfwayTime, "Start time should be updated");
        assertEq(streamer.rewardEndTime(), halfwayTime + newDuration, "End time should be updated");
        assertEq(streamer.lastRewardTime(), halfwayTime, "Last reward time should be updated");
    }

    function testSetRewardsWithPartialElapsedTime() public {
        // Setup initial reward period
        uint256 initialReward = 1200e18;
        uint256 duration = 12 days;
        uint256 startTime = vm.getBlockTimestamp();

        _setRewards(initialReward, duration);

        // Advance time to 1/3 through the reward period (4 days out of 12)
        uint256 oneThirdTime = startTime + duration / 3;
        vm.warp(oneThirdTime);

        // Set new rewards
        uint256 newRewardAmount = 800e18;
        uint256 newDuration = 8 days;

        _setRewards(newRewardAmount, newDuration);

        // Calculate expected remaining rewards
        // After 1/3 of time, 2/3 of rewards should remain
        uint256 expectedRemainingRewards = (initialReward * 2) / 3; // 800e18
        uint256 expectedTotalRewards = expectedRemainingRewards + newRewardAmount; // 1600e18

        assertEq(streamer.rewardAmount(), expectedTotalRewards, "Should accumulate remaining rewards correctly");
    }

    function testSetRewardsAfterPeriodEndedNoAccumulation() public {
        // Setup initial reward period
        uint256 initialReward = 1000e18;
        uint256 duration = 10 days;
        uint256 startTime = vm.getBlockTimestamp();

        _setRewards(initialReward, duration);

        // Advance time past the end of the reward period
        uint256 afterEndTime = startTime + duration + 1 days;
        vm.warp(afterEndTime);

        // Set new rewards after the period has ended
        uint256 newRewardAmount = 500e18;
        uint256 newDuration = 5 days;

        _setRewards(newRewardAmount, newDuration);

        // Should not accumulate any remaining rewards since period ended
        assertEq(streamer.rewardAmount(), newRewardAmount, "Should only use new reward amount when period has ended");
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
        vm.expectPartialRevert(IStakeManager.StakeManager__Unauthorized.selector);
        streamer.setReward(1000, 10);
    }

    function testSetRewards_RevertsBadDuration() public {
        vm.prank(admin);
        vm.expectRevert(IStakeManager.StakeManager__DurationCannotBeZero.selector);
        karma.setReward(address(streamer), 1000, 0);
    }

    function testSetRewards_RevertsBadAmount() public {
        vm.prank(admin);
        vm.expectRevert(IStakeManager.StakeManager__AmountCannotBeZero.selector);
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
        assertEq(streamer.mpAccruedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 0);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 0);
        assertEq(streamer.mpAccruedOf(vaults[bob]), 0);
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
        assertEq(streamer.mpAccruedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 500e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 100e18);
        assertEq(streamer.mpAccruedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 0);

        vm.warp(initialTime + year);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 200e18);
        assertEq(streamer.totalShares(), 400e18);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalMP(), 350e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 200e18);
        assertEq(streamer.mpAccruedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 150e18);
        assertEq(streamer.mpAccruedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 250e18);

        vm.warp(initialTime + year * 2);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 200e18);
        assertEq(streamer.totalShares(), 400e18);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalMP(), 550e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 300e18);
        assertEq(streamer.mpAccruedOf(vaults[alice]), 100e18);
        assertEq(streamer.vaultShares(vaults[alice]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 250e18);
        assertEq(streamer.mpAccruedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 250e18);

        _updateVault(alice);

        assertEq(streamer.totalStaked(), 200e18);
        assertEq(streamer.totalMPStaked(), 400e18);
        assertEq(streamer.totalShares(), 600e18);
        assertEq(streamer.totalRewardsSupply(), 1000e18);
        assertEq(streamer.totalMP(), 550e18);
        assertEq(streamer.mpBalanceOf(vaults[alice]), 300e18);
        assertEq(streamer.mpAccruedOf(vaults[alice]), 300e18);
        assertEq(streamer.vaultShares(vaults[alice]), 400e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 750e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 250e18);
        assertEq(streamer.mpAccruedOf(vaults[bob]), 100e18);
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
        assertEq(streamer.mpAccruedOf(vaults[alice]), 300e18);
        assertEq(streamer.vaultShares(vaults[alice]), 400e18);
        assertEq(streamer.rewardsBalanceOf(vaults[alice]), 1150e18);
        assertEq(streamer.mpBalanceOf(vaults[bob]), 350e18);
        assertEq(streamer.mpAccruedOf(vaults[bob]), 100e18);
        assertEq(streamer.vaultShares(vaults[bob]), 200e18);
        assertEq(streamer.rewardsBalanceOf(vaults[bob]), 450e18);
    }
}