pragma solidity ^0.8.26;

import { StakeManagerTest, StakeMath, IStakeManager } from "../../StakeManagerBase.t.sol";

contract LockTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_LockWithPriorLock() public {
        // Setup - alice stakes 10 tokens without lock
        uint256 stakeAmount = 10e18;
        _stake(alice, stakeAmount, 0);

        uint256 initialAccountMP = stakeAmount; // 10e18
        uint256 initialMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount; // 50e18

        // Verify initial state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        // Lock for 1 year
        uint256 lockPeriod = YEAR;
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockPeriod);

        _lock(alice, lockPeriod);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );

        expectedBonusMP = _bonusMP(stakeAmount, lockPeriod * 2);
        // Lock for more one 1 year
        _lock(alice, lockPeriod);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );
    }

    function test_LockWithoutPriorLock() public {
        // Setup - alice stakes 10 tokens without lock
        uint256 stakeAmount = 10e18;
        _stake(alice, stakeAmount, 0);

        uint256 initialAccountMP = stakeAmount; // 10e18
        uint256 initialMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount; // 50e18

        // Verify initial state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        // Lock for 1 year
        uint256 lockPeriod = YEAR;
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockPeriod);

        _lock(alice, lockPeriod);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );
    }

    function test_LockMultipleTimesExceedMaxLock() public {
        // Setup - alice stakes 10 tokens without lock
        uint256 stakeAmount = 10e18;

        _stake(alice, stakeAmount, 0);

        uint256 initialAccountMP = stakeAmount; // 10e18
        uint256 initialMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount; // 50e18

        // Verify initial state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        // Lock for 1 year
        uint256 lockPeriod = 4 * YEAR;
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockPeriod);

        _lock(alice, lockPeriod);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP + expectedBonusMP,
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );

        // wait for lock year to be over
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (4 * YEAR));

        streamer.updateVault(vaults[alice]);

        // Check updated state
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP + expectedBonusMP + (initialAccountMP * 4),
                maxMP: initialMaxMP + expectedBonusMP,
                rewardsAccrued: 0
            })
        );

        // lock for another year should fail as 4 years is the maximum of total lock time
        vm.expectRevert(StakeMath.StakeMath__AbsoluteMaxMPOverflow.selector);
        _lock(alice, YEAR);
    }

    function test_LockFailsWithNoStake() public {
        vm.expectRevert(StakeMath.StakeMath__InsufficientBalance.selector);
        _lock(alice, YEAR);
    }

    function test_LockFailsWithZero() public {
        _stake(alice, 10e18, 0);
        vm.expectRevert(IStakeManager.StakeManager__DurationCannotBeZero.selector);
        _lock(alice, 0);
    }

    function test_LockFailsWithInvalidPeriod(uint256 _lockPeriod) public {
        vm.assume(_lockPeriod > 0);
        vm.assume(_lockPeriod < MIN_LOCKUP_PERIOD || _lockPeriod > MAX_LOCKUP_PERIOD);
        vm.assume(_lockPeriod < (type(uint256).max - block.timestamp)); //prevents arithmetic overflow

        _stake(alice, 10e18, 0);

        vm.expectRevert();
        _lock(alice, _lockPeriod);
    }

    function test_RevertWhenVaultToLockIsEmpty() public {
        vm.expectRevert(StakeMath.StakeMath__InsufficientBalance.selector);
        _lock(alice, YEAR);
    }
}