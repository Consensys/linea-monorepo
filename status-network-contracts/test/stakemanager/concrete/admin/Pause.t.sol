pragma solidity ^0.8.26;

import { StakeManagerTest, IStakeManager } from "../../StakeManagerBase.t.sol";

contract PauseTest is StakeManagerTest {
    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhenNotAdmin() public {
        vm.prank(alice);
        vm.expectRevert(IStakeManager.StakeManager__Unauthorized.selector);
        streamer.pause();
    }

    function test_RevertWhenNotGuardian() public {
        vm.prank(alice);
        vm.expectRevert(IStakeManager.StakeManager__Unauthorized.selector);
        streamer.pause();
    }

    function test_PauseAndUnpause() public {
        // pause the contract
        vm.prank(admin);
        streamer.pause();

        // ensure staking is paused
        vm.expectRevert("Pausable: paused");
        _stake(alice, 10e18, 0);

        // unpause the contract
        vm.prank(admin);
        streamer.unpause();

        // ensure staking works again
        _stake(alice, 10e18, 0);
    }
}