pragma solidity ^0.8.26;

import { StakeManagerTest, StakeVault } from "../../StakeManagerBase.t.sol";

contract UpdateVaultTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_UpdateAccount() public {
        uint256 stakeAmount = 1000e18;

        // alice stakes 1000 tokens
        _stake(alice, stakeAmount, 0);

        // alice creates new vaults
        vm.startPrank(alice);
        address vault2 = address(vaultFactory.createVault());
        address vault3 = address(vaultFactory.createVault());
        address vault4 = address(vaultFactory.createVault());

        stakingToken.approve(vault2, stakeAmount);
        stakingToken.approve(vault3, stakeAmount);
        stakingToken.approve(vault4, stakeAmount);

        // alice stakes 1000 tokens in each vault
        StakeVault(vault2).stake(stakeAmount, 0);
        StakeVault(vault3).stake(stakeAmount, 0);
        StakeVault(vault4).stake(stakeAmount, 0);
        vm.stopPrank();

        // ensure alice has expected MP balance
        assertEq(streamer.mpBalanceOfAccount(alice), stakeAmount * 4); // 4 vaults, 1000e18 staked each

        // distribute rewards
        uint256 rewards = 10_000e18;
        uint256 rewardPeriod = YEAR;
        _setRewards(rewards, rewardPeriod);

        vm.warp(vm.getBlockTimestamp() + rewardPeriod);

        // ensure staked MP haven't changed for alice (yet!)
        assertEq(streamer.mpAccruedOf(vaults[alice]), stakeAmount);
        assertEq(streamer.mpAccruedOf(vault2), stakeAmount);
        assertEq(streamer.mpAccruedOf(vault3), stakeAmount);
        assertEq(streamer.mpAccruedOf(vault4), stakeAmount);

        // compound alice's MP
        streamer.updateAccount(alice);

        uint256 expectedMPIncreasePerVault = _accrueMP(stakeAmount, rewardPeriod);

        // ensure alice's staked MP have been compounded
        assertEq(streamer.mpAccruedOf(vaults[alice]), stakeAmount + expectedMPIncreasePerVault);
        assertEq(streamer.mpAccruedOf(vault2), stakeAmount + expectedMPIncreasePerVault);
        assertEq(streamer.mpAccruedOf(vault3), stakeAmount + expectedMPIncreasePerVault);
        assertEq(streamer.mpAccruedOf(vault4), stakeAmount + expectedMPIncreasePerVault);

        uint256 tolerance = 1000;
        assertApproxEqAbs(streamer.rewardsBalanceOfAccount(alice), rewards, tolerance, "Reward balance mismatch");
    }
}