pragma solidity ^0.8.26;

import { StakeManagerTest } from "../../StakeManagerBase.t.sol";

contract VotesTest is StakeManagerTest {
    uint256 public rewardAmount;

    function setUp() public virtual override {
        super.setUp();

        rewardAmount = 10e18;
        uint256 stakeAmount = 50e18;
        uint256 rewardPeriod = 1 weeks;

        assertEq(karma.balanceOf(alice), 0);
        assertEq(karma.getVotes(alice), 0);

        _stake(alice, stakeAmount, 0);
        _setRewards(rewardAmount, rewardPeriod);
        vm.warp(vm.getBlockTimestamp() + rewardPeriod);

        assertEq(karma.balanceOf(alice), rewardAmount);
        assertEq(karma.getVotes(alice), 0);
    }

    function test_delegateBeforeRedeemingCreatesCheckpointsWithoutVotes() public {
        vm.prank(alice);
        karma.delegate(alice);

        assertEq(karma.balanceOf(alice), rewardAmount);
        // Alice delegated to herself but she only has virtual Karma, which is not counted for votes.
        assertEq(karma.getVotes(alice), 0);

        vm.prank(alice);
        streamer.redeemRewards(alice);

        assertEq(karma.balanceOf(alice), rewardAmount);
        assertEq(karma.getVotes(alice), rewardAmount);

        // there's nothing more to delegate
        vm.prank(alice);
        karma.delegate(alice);

        assertEq(karma.balanceOf(alice), rewardAmount);
        assertEq(karma.getVotes(alice), rewardAmount);
    }

    function test_redeemWithoutDelegatingBeforeDoesntCreateCheckpoints() public {
        assertEq(karma.balanceOf(alice), rewardAmount);
        assertEq(karma.getVotes(alice), 0);

        vm.prank(alice);
        streamer.redeemRewards(alice);

        assertEq(karma.balanceOf(alice), rewardAmount);
        // karma has been redeemed but votes have not been delegated
        assertEq(karma.getVotes(alice), 0);

        vm.prank(alice);
        // now votes are properly delegated
        karma.delegate(alice);

        assertEq(karma.balanceOf(alice), rewardAmount);
        assertEq(karma.getVotes(alice), rewardAmount);
    }
}
