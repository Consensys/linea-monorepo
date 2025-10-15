pragma solidity ^0.8.26;

import { StakeManagerTest, StakeManager, IStakeManager, UUPSUpgradeable } from "../../StakeManagerBase.t.sol";

contract UpgradeTest is StakeManagerTest {
    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhenNotOwner() public {
        address newImpl = address(new StakeManager());
        bytes memory initializeData;
        vm.prank(alice);
        vm.expectRevert(IStakeManager.StakeManager__Unauthorized.selector);
        UUPSUpgradeable(streamer).upgradeToAndCall(newImpl, initializeData);
    }

    function test_UpgradeStakeManager() public {
        // first, change state of existing stake manager
        _stake(alice, 10e18, 0);

        // check initial state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // next, upgrade the stake manager
        _upgradeStakeManager();

        // ensure state is available in upgraded contract
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }
}
