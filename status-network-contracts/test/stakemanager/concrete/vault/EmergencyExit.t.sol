pragma solidity ^0.8.26;

import { StakeManagerTest, StakeVault, IStakeManager } from "../../StakeManagerBase.t.sol";

contract EmergencyExitTest is StakeManagerTest {
    function setUp() public override {
        super.setUp();
    }

    function test_CannotLeaveBeforeEmergencyMode() public {
        _stake(alice, 10e18, 0);
        vm.expectRevert(StakeVault.StakeVault__NotAllowedToExit.selector);
        _emergencyExit(alice);
    }

    function test_OwnerCanEnableEmergencyMode() public {
        vm.prank(admin);
        streamer.enableEmergencyMode();
    }

    function test_GuardianCanEnableEmergencyMode() public {
        vm.prank(guardian);
        streamer.enableEmergencyMode();
    }

    function test_OnlyOwnerOrGuardianCanEnableEmergencyMode() public {
        vm.prank(alice);
        vm.expectRevert(IStakeManager.StakeManager__Unauthorized.selector);
        streamer.enableEmergencyMode();
    }

    function test_CannotEnableEmergencyModeTwice() public {
        vm.prank(admin);
        streamer.enableEmergencyMode();

        vm.expectRevert(IStakeManager.StakeManager__EmergencyModeEnabled.selector);
        vm.prank(admin);
        streamer.enableEmergencyMode();
    }

    function test_EmergencyExitBasic() public {
        uint256 aliceBalance = stakingToken.balanceOf(alice);

        _stake(alice, 10e18, 0);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        // emergency exit will not perform any internal accounting
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        assertEq(stakingToken.balanceOf(alice), aliceBalance, "Alice should get tokens back");
        assertEq(stakingToken.balanceOf(vaults[alice]), 0, "Vault should be empty");
    }

    function test_EmergencyExitWithRewards() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        _stake(alice, 10e18, 0);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 1000e18,
                rewardIndex: 50e18
            })
        );

        // Check Alice staked tokens but no rewards
        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get staked tokens back");
        assertEq(stakingToken.balanceOf(address(vaults[alice])), 0, "Vault should be empty");
    }

    function test_EmergencyExitWithLock() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);

        _stake(alice, 10e18, 90 days);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        // Check Alice got tokens back despite lock
        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get tokens back despite lock");
        assertEq(stakingToken.balanceOf(address(vaults[alice])), 0, "Vault should be empty");
    }

    function test_EmergencyExitMultipleUsers() public {
        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 bobInitialBalance = stakingToken.balanceOf(bob);

        // Setup multiple stakers
        _stake(alice, 10e18, 0);
        _stake(bob, 30e18, 0);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        // Alice exits first
        _emergencyExit(alice);

        // Check intermediate state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17
            })
        );

        // Bob exits
        _emergencyExit(bob);

        // Check final state
        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 30e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        // Verify both users got their tokens back
        assertEq(stakingToken.balanceOf(alice), aliceInitialBalance, "Alice should get staked tokens back");
        assertEq(stakingToken.balanceOf(bob), bobInitialBalance, "Bob should get staked tokens back");
        assertEq(stakingToken.balanceOf(vaults[alice]), 0, "Alice vault should have 0 staked tokens");
        assertEq(stakingToken.balanceOf(vaults[bob]), 0, "Bob vault should have 0 staked tokens");
    }

    function test_EmergencyExitToAlternateAddress() public {
        _stake(alice, 10e18, 0);

        address alternateAddress = makeAddr("alternate");
        uint256 alternateInitialBalance = stakingToken.balanceOf(alternateAddress);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        // Alice exits to alternate address
        vm.prank(alice);
        StakeVault aliceVault = StakeVault(vaults[alice]);
        aliceVault.emergencyExit(alternateAddress);

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        // Check alternate address received everything
        assertEq(
            stakingToken.balanceOf(alternateAddress),
            alternateInitialBalance + 10e18,
            "Alternate address should get staked tokens"
        );
    }
}