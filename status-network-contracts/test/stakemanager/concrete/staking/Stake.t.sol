pragma solidity ^0.8.26;

import { StakeManagerTest, StakeMath, StakeVault, StakeManager } from "../../StakeManagerBase.t.sol";

contract StakeTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_StakeOneAccount() public {
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

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
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );
    }

    function test_StakeOneAccountAndRewards() public {
        _stake(alice, 10e18, 0);

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

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
                rewardIndex: 0,
                mpAccrued: 10e18,
                maxMP: 50e18,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 10e18,
                totalMPAccrued: 10e18,
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 1000e18,
                rewardIndex: 50e18 // (1000 rewards / (10 staked + 10 MP)) = 50
             })
        );
    }

    function test_StakeOneAccountWithMinLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD();
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10e18 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountWithMaxLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MAX_LOCKUP_PERIOD();
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: 90e18,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountWithRandomLockUp() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD() + 13 days;
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);

        _stake(alice, stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                // 10 + (amount * (lockPeriod * MAX_MULTIPLIER * SCALE_FACTOR / MAX_LOCKUP_PERIOD) / SCALE_FACTOR)
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeOneAccountMPIncreasesMaxMPDoesNotChange() public {
        uint256 stakeAmount = 15e18;
        uint256 totalMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;
        uint256 totalMPAccrued = stakeAmount;

        _stake(alice, stakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount,
                totalMPAccrued: stakeAmount,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR));

        streamer.updateGlobalState();
        streamer.updateVault(vaults[alice]);

        uint256 expectedMPIncrease = stakeAmount; // 1 year passed, 1 MP accrued per token staked
        totalMPAccrued = totalMPAccrued + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: totalMPAccrued,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP,
                rewardsAccrued: 0
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));

        streamer.updateGlobalState();
        streamer.updateVault(vaults[alice]);

        expectedMPIncrease = stakeAmount / 2; // 1/2 year passed, 1/2 MP accrued per token staked
        totalMPAccrued = totalMPAccrued + expectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: totalMPAccrued,
                totalMPAccrued: totalMPAccrued,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP,
                rewardsAccrued: 0
            })
        );
    }

    function test_StakeOneAccountReachingMPLimit() public {
        uint256 stakeAmount = 15e18;
        uint256 totalMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;
        uint256 totalMPAccrued = stakeAmount;

        _stake(alice, stakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount,
                totalMPAccrued: stakeAmount,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMPAccrued, // accountMP == totalMPAccrued because only one account is staking
                maxMP: totalMaxMP, // maxMP == totalMaxMP because only one account is staking
                rewardsAccrued: 0
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        uint256 timeToMaxMP = _timeToAccrueMP(stakeAmount, totalMaxMP - totalMPAccrued);
        vm.warp(currentTime + timeToMaxMP);

        streamer.updateVault(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: totalMaxMP,
                totalMPAccrued: totalMaxMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: totalMaxMP,
                maxMP: totalMaxMP,
                rewardsAccrued: 0
            })
        );

        // move forward in time to check we're not producing more MP
        currentTime = vm.getBlockTimestamp();
        // increasing time by some big enough time such that MPs are actually generated
        vm.warp(currentTime + 14 days);

        streamer.updateVault(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: totalMaxMP,
                totalMPAccrued: totalMaxMP,
                totalMaxMP: totalMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeMultipleTimesWithLockZeroAfterMaxLock() public {
        uint256 stakeAmount = 10e6;
        uint256 initialTime = vm.getBlockTimestamp();

        // stake and lock 4 years
        _stake(alice, stakeAmount, 4 * YEAR);

        // staking with lock 0 should work even before lock up has expired
        vm.warp(initialTime + 2 * YEAR);
        _stake(alice, stakeAmount, 0);

        // staking with lock 0 should work again when lock up has expired
        vm.warp(initialTime + 4 * YEAR);
        _stake(alice, stakeAmount, 0);

        _stake(alice, stakeAmount, 0);
        // locking up to new limit should render same maxMP as staking initially the whole
        _lock(alice, _lockTimeAvailable(stakeAmount * 4, streamer.getVault(vaults[alice]).maxMP));
        _stake(bob, stakeAmount * 4, MAX_LOCKUP_PERIOD);
        assertEq(streamer.getVault(vaults[bob]).maxMP, streamer.getVault(vaults[alice]).maxMP);
    }

    function test_StakeMultipleTimesWithLockIncreaseAtSameBlock() public {
        uint256 stakeAmount = 10e18;
        uint256 expectedStake = stakeAmount;
        uint256 lockUpIncrease = YEAR;
        uint256 expectedBonus = _bonusMP(stakeAmount, lockUpIncrease);
        uint256 expectedMP = stakeAmount;
        uint256 expectedMaxMP = expectedMP + expectedBonus + (stakeAmount * streamer.MAX_MULTIPLIER());

        // Alice stakes 10 tokens, locks for 1 year
        _stake(alice, stakeAmount, lockUpIncrease);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: expectedStake,
                totalMPStaked: expectedMP + expectedBonus,
                totalMPAccrued: expectedMP + expectedBonus,
                totalMaxMP: expectedMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // Alice stakes again 10 tokens and increases lock by 3 years
        // Since time hasn't passed yet, we essentially have a total lock up
        // of 4 years
        lockUpIncrease = 3 * YEAR;

        // new bonus = old bonus + bonus increase for old stake + bonus for new stake + bonus for new stake
        expectedBonus = expectedBonus + _bonusMP(stakeAmount, lockUpIncrease) + _bonusMP(stakeAmount, lockUpIncrease)
        // This is the bonus for the new stake on the previous lock up
        + _bonusMP(stakeAmount, YEAR);
        expectedMP = expectedMP + stakeAmount;
        expectedMaxMP = expectedMP + expectedBonus + ((stakeAmount * 2) * streamer.MAX_MULTIPLIER());

        _stake(alice, stakeAmount, lockUpIncrease);

        expectedStake = expectedStake + stakeAmount;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: expectedStake,
                totalMPStaked: expectedMP + expectedBonus,
                totalMPAccrued: expectedMP + expectedBonus,
                totalMaxMP: expectedMaxMP,
                stakingBalance: expectedStake,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // any lock up beyond the max lock up period should revert
        vm.expectRevert(StakeMath.StakeMath__InvalidLockingPeriod.selector);
        _stake(alice, 1, MIN_LOCKUP_PERIOD);
    }

    function test_StakeMultipleTimesDoesNotExceedsMaxMP() public {
        // stake and lock 1 year
        uint256 stakeAmount = 10e16;
        uint256 i = 0;
        do {
            i++;
            _stake(alice, stakeAmount, YEAR);
            vm.warp(vm.getBlockTimestamp() + YEAR);
        } while (_lockTimeAvailable(stakeAmount * i, streamer.getVault(vaults[alice]).maxMP) > MIN_LOCKUP_PERIOD);
        _stake(bob, stakeAmount * i, _estimateLockTime(streamer.getVault(vaults[alice]).maxMP, stakeAmount * i));
        assertEq(streamer.getVault(vaults[bob]).maxMP, streamer.getVault(vaults[alice]).maxMP);
    }

    function test_StakeMultipleAccounts() public {
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

        // Bob stakes 30 tokens
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
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
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );
    }

    function test_StakeMultipleAccountsAndRewards() public {
        // Alice stakes 10 tokens
        _stake(alice, 10e18, 0);

        // Bob stakes 30 tokens
        _stake(bob, 30e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 10e18,
                vaultBalance: 10e18,
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
                vaultBalance: 30e18,
                rewardIndex: 0,
                mpAccrued: 30e18,
                maxMP: 150e18,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 40e18,
                totalMPStaked: 40e18,
                totalMPAccrued: 40e18,
                totalMaxMP: 200e18,
                stakingBalance: 40e18,
                rewardBalance: 1000e18,
                rewardIndex: 125e17 // (1000 rewards / (40 staked + 40 MP)) = 12,5
             })
        );
    }

    function test_StakeMultipleAccountsWithMinLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MIN_LOCKUP_PERIOD();
        uint256 aliceExpectedBonusMP = _bonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = 0;
        uint256 bobExpectedBonusMP = _bonusMP(bobStakeAmount, bobLockUpPeriod);

        // alice stakes with lockup period
        _stake(alice, aliceStakeAmount, aliceLockUpPeriod);

        // Bob stakes 30 tokens
        _stake(bob, bobStakeAmount, bobLockUpPeriod);

        uint256 sumOfStakeAmount = aliceStakeAmount + bobStakeAmount;
        uint256 sumOfExpectedBonusMP = aliceExpectedBonusMP + bobExpectedBonusMP;
        uint256 expectedMaxTotalMP =
            _maxTotalMP(aliceStakeAmount, aliceLockUpPeriod) + _maxTotalMP(bobStakeAmount, bobLockUpPeriod);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: sumOfStakeAmount,
                totalMPStaked: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMPAccrued: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_StakeMultipleAccountsWithRandomLockUp() public {
        uint256 aliceStakeAmount = 10e18;
        uint256 aliceLockUpPeriod = streamer.MAX_LOCKUP_PERIOD() - 21 days;
        uint256 aliceExpectedBonusMP = _bonusMP(aliceStakeAmount, aliceLockUpPeriod);

        uint256 bobStakeAmount = 30e18;
        uint256 bobLockUpPeriod = streamer.MIN_LOCKUP_PERIOD() + 43 days;
        uint256 bobExpectedBonusMP = _bonusMP(bobStakeAmount, bobLockUpPeriod);

        // alice stakes with lockup period
        _stake(alice, aliceStakeAmount, aliceLockUpPeriod);

        // Bob stakes 30 tokens
        _stake(bob, bobStakeAmount, bobLockUpPeriod);

        uint256 sumOfStakeAmount = aliceStakeAmount + bobStakeAmount;
        uint256 sumOfExpectedBonusMP = aliceExpectedBonusMP + bobExpectedBonusMP;
        uint256 expectedMaxTotalMP =
            _maxTotalMP(aliceStakeAmount, aliceLockUpPeriod) + _maxTotalMP(bobStakeAmount, bobLockUpPeriod);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: sumOfStakeAmount,
                totalMPStaked: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMPAccrued: sumOfStakeAmount + sumOfExpectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: sumOfStakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    struct TestParams {
        uint256 aliceStakeAmount;
        uint256 bobStakeAmount;
        uint256 totalStaked;
        uint256 totalMPAccrued;
        uint256 totalMaxMP;
    }

    function test_StakeMultipleAccountsMPIncreasesMaxMPDoesNotChange() public {
        TestParams memory params;
        params.aliceStakeAmount = 15e18;
        params.bobStakeAmount = 5e18;
        params.totalStaked = params.aliceStakeAmount + params.bobStakeAmount;
        params.totalMPAccrued = params.totalStaked;
        params.totalMaxMP = (params.aliceStakeAmount * streamer.MAX_MULTIPLIER() + params.aliceStakeAmount)
            + (params.bobStakeAmount * streamer.MAX_MULTIPLIER() + params.bobStakeAmount);

        uint256 aliceMP = params.aliceStakeAmount;
        uint256 aliceMaxMP = params.aliceStakeAmount * streamer.MAX_MULTIPLIER() + aliceMP;

        uint256 bobMP = params.bobStakeAmount;
        uint256 bobMaxMP = params.bobStakeAmount * streamer.MAX_MULTIPLIER() + bobMP;
        _stake(alice, params.aliceStakeAmount, 0);
        _stake(bob, params.bobStakeAmount, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: params.totalStaked,
                totalMPStaked: params.totalMPAccrued,
                totalMPAccrued: params.totalMPAccrued,
                totalMaxMP: params.totalMaxMP,
                stakingBalance: params.totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: params.aliceStakeAmount,
                vaultBalance: params.aliceStakeAmount,
                rewardIndex: 0,
                mpAccrued: aliceMP,
                maxMP: aliceMaxMP,
                rewardsAccrued: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: params.bobStakeAmount,
                vaultBalance: params.bobStakeAmount,
                rewardIndex: 0,
                mpAccrued: bobMP,
                maxMP: bobMaxMP,
                rewardsAccrued: 0
            })
        );

        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR));

        streamer.updateVault(vaults[alice]);
        streamer.updateVault(vaults[bob]);

        uint256 aliceExpectedMPIncrease = params.aliceStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 bobExpectedMPIncrease = params.bobStakeAmount; // 1 year passed, 1 MP accrued per token staked
        uint256 totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        uint256 aliceMPAccrued = aliceMP + aliceExpectedMPIncrease;
        uint256 bobMPAccrued = bobMP + bobExpectedMPIncrease;
        params.totalMPAccrued = params.totalMPAccrued + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: params.totalStaked,
                totalMPStaked: params.totalMPAccrued,
                totalMPAccrued: params.totalMPAccrued,
                totalMaxMP: params.totalMaxMP,
                stakingBalance: params.totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: params.aliceStakeAmount,
                vaultBalance: params.aliceStakeAmount,
                rewardIndex: 0,
                mpAccrued: aliceMP + aliceExpectedMPIncrease,
                maxMP: aliceMaxMP,
                rewardsAccrued: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: params.bobStakeAmount,
                vaultBalance: params.bobStakeAmount,
                rewardIndex: 0,
                mpAccrued: bobMPAccrued,
                maxMP: bobMaxMP,
                rewardsAccrued: 0
            })
        );

        currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR / 2));

        streamer.updateVault(vaults[alice]);
        streamer.updateVault(vaults[bob]);

        aliceExpectedMPIncrease = params.aliceStakeAmount / 2;
        bobExpectedMPIncrease = params.bobStakeAmount / 2;
        totalExpectedMPIncrease = aliceExpectedMPIncrease + bobExpectedMPIncrease;

        aliceMPAccrued = aliceMPAccrued + aliceExpectedMPIncrease;
        bobMPAccrued = bobMPAccrued + bobExpectedMPIncrease;
        params.totalMPAccrued = params.totalMPAccrued + totalExpectedMPIncrease;

        checkStreamer(
            CheckStreamerParams({
                totalStaked: params.totalStaked,
                totalMPStaked: params.totalMPAccrued,
                totalMPAccrued: params.totalMPAccrued,
                totalMaxMP: params.totalMaxMP,
                stakingBalance: params.totalStaked,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: params.aliceStakeAmount,
                vaultBalance: params.aliceStakeAmount,
                rewardIndex: 0,
                mpAccrued: aliceMPAccrued,
                maxMP: aliceMaxMP,
                rewardsAccrued: 0
            })
        );
        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: params.bobStakeAmount,
                vaultBalance: params.bobStakeAmount,
                rewardIndex: 0,
                mpAccrued: bobMPAccrued,
                maxMP: bobMaxMP,
                rewardsAccrued: 0
            })
        );
    }
}