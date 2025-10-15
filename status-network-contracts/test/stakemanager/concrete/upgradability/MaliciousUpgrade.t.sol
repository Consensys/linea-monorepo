pragma solidity ^0.8.26;

import { StakeManagerTest, StackOverflowStakeManager, UUPSUpgradeable } from "../../StakeManagerBase.t.sol";

contract MaliciousUpgradeTest is StakeManagerTest {
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