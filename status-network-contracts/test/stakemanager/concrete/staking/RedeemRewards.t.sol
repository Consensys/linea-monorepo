pragma solidity ^0.8.26;

import { StakeManagerTest } from "../../StakeManagerBase.t.sol";

contract RedeemRewardsTest is StakeManagerTest {
    uint256 public rewardAmount = 1000e18;
    uint256 public rewardDuration = 10 days;

    function setUp() public override {
        super.setUp();
    }

    function test_RedeemRewardsZeroRewards() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 aliceInitialKarmaBalance = karma.balanceOf(alice);

        _stake(alice, 100e18, 0);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance - 100e18, "Alice should have staked tokens");

        // let half of the reward duration pass
        vm.warp(block.timestamp + rewardDuration);

        vm.prank(alice);
        uint256 redeemed = streamer.redeemRewards(alice);

        assertEq(redeemed, 0, "Alice redeemed zero Karma rewards");
        assertEq(karma.balanceOf(alice), aliceInitialKarmaBalance, "Alice redeemed half the Karma rewards");
    }

    function test_RedeemRewards() public {
        vm.startPrank(admin);
        karma.setReward(address(streamer), rewardAmount, rewardDuration);
        vm.stopPrank();

        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 aliceInitialKarmaBalance = karma.balanceOf(alice);
        uint256 streamerInitialKarmaBalance = karma.balanceOfRewardDistributor(address(streamer));

        _stake(alice, 100e18, 0);

        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance - 100e18, "Alice should have staked tokens");

        // let half of the reward duration pass
        vm.warp(block.timestamp + rewardDuration);

        streamer.updateGlobalState();
        uint256 totalRewardsAccruedBefore = streamer.totalRewardsAccrued();

        vm.prank(alice);
        uint256 redeemed = streamer.redeemRewards(alice);

        assertEq(redeemed, rewardAmount, "Alice redeemed all the Karma rewards");
        assertEq(
            streamer.totalRewardsAccrued(),
            totalRewardsAccruedBefore - rewardAmount,
            "Streamer totalRewardsAccrued decreased"
        );
        assertEq(karma.balanceOf(alice), aliceInitialKarmaBalance + rewardAmount, "Alice redeemed Karma rewards");
        assertEq(
            karma.balanceOfRewardDistributor(address(streamer)),
            streamerInitialKarmaBalance - rewardAmount,
            "Streamer paid the Karma rewards"
        );
    }
}