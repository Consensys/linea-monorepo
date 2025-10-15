pragma solidity ^0.8.26;

import { StakeManagerTest, StakeManager, StakeVault, Math, StakeMath, IStakeManager } from "./StakeManagerBase.t.sol";

contract FuzzTests is StakeManagerTest {
    struct CheckVaultLockParams {
        uint256 lockEnd;
        uint256 totalLockUp;
    }

    bytes4 constant NO_REVERT = 0x00000000;

    error FuzzTests__UndefinedError();

    bytes4 expectedRevert = FuzzTests__UndefinedError.selector;
    CheckStreamerParams expectedSystemState = CheckStreamerParams({
        totalStaked: 0,
        totalMPStaked: 0,
        totalMPAccrued: 0,
        totalMaxMP: 0,
        stakingBalance: 0,
        rewardBalance: 0,
        rewardIndex: 0
    });
    mapping(address userAddress => CheckVaultParams params) public expectedAccountState;
    mapping(address vaultAddress => CheckVaultLockParams params) public expectedVaultLockState;

    function check(string memory test) internal view {
        check(test, expectedSystemState);
    }

    function check(string memory test, address vaultOwner) internal view {
        check(test, expectedAccountState[vaultOwner]);
        check(test, expectedSystemState);
    }

    function check(string memory text, CheckStreamerParams storage p) internal view {
        assertEq(streamer.totalStaked(), p.totalStaked, string(abi.encodePacked(text, "wrong total staked")));
        assertEq(streamer.totalMPStaked(), p.totalMPStaked, string(abi.encodePacked(text, "wrong total staked MP")));
        assertEq(streamer.totalMPAccrued(), p.totalMPAccrued, string(abi.encodePacked(text, "wrong total accrued MP")));
        assertEq(streamer.totalMaxMP(), p.totalMaxMP, string(abi.encodePacked(text, "wrong totalMaxMP MP")));
        // assertEq(rewardToken.balanceOf(address(streamer)), p.rewardBalance, "wrong reward balance");
        // assertEq(streamer.rewardIndex(), p.rewardIndex, "wrong reward index");
    }

    function check(string memory text, CheckVaultParams storage p) internal view {
        // assertEq(rewardToken.balanceOf(p.account), p.rewardBalance, "wrong account reward balance");

        StakeManager.VaultData memory vaultData = streamer.getVault(p.account);

        assertEq(
            vaultData.stakedBalance, p.stakedBalance, string(abi.encodePacked(text, "wrong account staked balance"))
        );
        assertEq(
            stakingToken.balanceOf(p.account), p.vaultBalance, string(abi.encodePacked(text, "wrong vault balance"))
        );
        // assertEq(vaultData.accountRewardIndex, p.rewardIndex, "wrong account reward index");
        assertEq(vaultData.mpAccrued, p.mpAccrued, string(abi.encodePacked(text, "wrong account MP accrued")));
        assertEq(vaultData.maxMP, p.maxMP, string(abi.encodePacked(text, "wrong account max MP")));
        assertEq(
            vaultData.rewardsAccrued, p.rewardsAccrued, string(abi.encodePacked(text, "wrong account rewards accrued"))
        );
        assertEq(
            StakeVault(p.account).lockUntil(),
            expectedVaultLockState[p.account].lockEnd,
            string(abi.encodePacked(text, "wrong account lock end"))
        );
    }

    function _stake(address account, uint256 amount, uint256 lockPeriod, bytes4 _expectedRevert) internal {
        stakingToken.mint(account, amount);
        StakeVault vault = StakeVault(vaults[account]);
        vm.startPrank(account);
        stakingToken.approve(vaults[account], amount);
        _expectRevert(_expectedRevert);
        vault.stake(amount, lockPeriod);
        vm.stopPrank();
        expectedRevert = FuzzTests__UndefinedError.selector;
    }

    function _lock(address account, uint256 lockPeriod, bytes4 _expectedRevert) internal {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        _expectRevert(_expectedRevert);
        vault.lock(lockPeriod);
        expectedRevert = FuzzTests__UndefinedError.selector;
    }

    function _expectRevert(bytes4 _expectedRevert) internal {
        if (_expectedRevert != NO_REVERT) {
            if (_expectedRevert == FuzzTests__UndefinedError.selector) {
                vm.expectRevert();
            } else {
                vm.expectRevert(_expectedRevert);
            }
        }
    }

    function _updateVault(address account, bytes4 _expectedRevert) internal {
        StakeVault vault = StakeVault(vaults[account]);
        _expectRevert(_expectedRevert);
        streamer.updateVault(address(vault));
    }

    function _accrue(address account, uint256 accruedTime) internal {
        if (accruedTime > 0) {
            vm.warp(vm.getBlockTimestamp() + accruedTime);
        }
        streamer.updateVault(vaults[account]);
    }

    function _unstake(address account, uint256 amount, bytes4 _expectedRevert) internal {
        StakeVault vault = StakeVault(vaults[account]);
        vm.prank(account);
        _expectRevert(_expectedRevert);
        vault.unstake(amount);
        expectedRevert = FuzzTests__UndefinedError.selector;
    }

    function _expectUnstake(address account, uint256 amount) internal {
        CheckVaultParams storage expectedAccountParams = expectedAccountState[account];
        expectedAccountParams.account = vaults[account];
        if (expectedVaultLockState[expectedAccountParams.account].lockEnd > vm.getBlockTimestamp()) {
            expectedRevert = StakeVault.StakeVault__FundsLocked.selector;
            return;
        }
        if (amount == 0) {
            expectedRevert = StakeMath.StakeMath__InvalidAmount.selector;
            return;
        }
        if (amount > expectedAccountParams.stakedBalance) {
            expectedRevert = StakeVault.StakeVault__NotEnoughAvailableBalance.selector;
            return;
        }
        expectedRevert = NO_REVERT;
        uint256 expectedReducedMP =
            _reduceMP(expectedAccountParams.stakedBalance, expectedAccountParams.mpAccrued, amount);
        uint256 expectedReducedMaxMP =
            _reduceMP(expectedAccountParams.stakedBalance, expectedAccountParams.maxMP, amount);
        expectedAccountParams.stakedBalance -= amount;
        expectedAccountParams.vaultBalance -= amount;
        expectedSystemState.stakingBalance -= amount;
        expectedSystemState.totalStaked -= amount;
        expectedSystemState.totalMPStaked -= expectedReducedMP;
        expectedAccountParams.mpAccrued -= expectedReducedMP;
        expectedSystemState.totalMPAccrued -= expectedReducedMP;
        expectedAccountParams.maxMP -= expectedReducedMaxMP;
        expectedSystemState.totalMaxMP -= expectedReducedMaxMP;
    }

    function _expectAccrue(address account, uint256 accruedTime) internal {
        CheckVaultParams storage expectedAccountParams = expectedAccountState[account];
        expectedAccountParams.account = vaults[account];
        if (expectedAccountParams.vaultBalance > 0) {
            uint256 rawAccruedMP = _accrueMP(expectedAccountParams.vaultBalance, accruedTime);
            expectedAccountParams.mpAccrued =
                Math.min(expectedAccountParams.mpAccrued + rawAccruedMP, expectedAccountParams.maxMP);
            expectedSystemState.totalMPStaked =
                Math.min(expectedSystemState.totalMPStaked + rawAccruedMP, expectedSystemState.totalMaxMP);

            expectedSystemState.totalMPAccrued =
                Math.min(expectedSystemState.totalMPAccrued + rawAccruedMP, expectedSystemState.totalMaxMP);
        }
    }

    function _expectStake(address account, uint256 stakeAmount, uint256 lockUpPeriod) internal {
        CheckVaultParams storage expectedAccountParams = expectedAccountState[account];
        expectedAccountParams.account = vaults[account];
        uint256 calcLockEnd = Math.max(
            expectedVaultLockState[expectedAccountParams.account].lockEnd, vm.getBlockTimestamp()
        ) + lockUpPeriod;
        uint256 calcLockUpPeriod = calcLockEnd - vm.getBlockTimestamp(); //increased lock + remaining current lock
        if (lockUpPeriod == 0 || (lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD)) {
            //valid raw input
            if (expectedVaultLockState[expectedAccountParams.account].totalLockUp + lockUpPeriod > MAX_LOCKUP_PERIOD) {
                // but total lock time surpassed the maximum allowed
                expectedRevert = StakeMath.StakeMath__AbsoluteMaxMPOverflow.selector;
            } else {
                expectedRevert = NO_REVERT;
                uint256 expectedBonusMP = _bonusMP(stakeAmount, calcLockUpPeriod);
                uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, calcLockUpPeriod);
                if (expectedVaultLockState[expectedAccountParams.account].lockEnd > vm.getBlockTimestamp()) {
                    // in case stake increased for a locked vault
                    // increases MP for the previous balance
                    expectedBonusMP += _bonusMP(expectedAccountParams.stakedBalance, lockUpPeriod);
                    expectedMaxTotalMP += _maxTotalMP(expectedAccountParams.stakedBalance, lockUpPeriod);
                }

                if (lockUpPeriod > 0) {
                    //update lockup end
                    expectedVaultLockState[expectedAccountParams.account].totalLockUp += lockUpPeriod;
                }
                expectedVaultLockState[expectedAccountParams.account].lockEnd = calcLockEnd;
                expectedAccountParams.stakedBalance = stakeAmount;
                expectedAccountParams.vaultBalance = stakeAmount;
                expectedSystemState.stakingBalance += stakeAmount;
                expectedSystemState.totalStaked += stakeAmount;
                expectedSystemState.totalMPStaked += stakeAmount + expectedBonusMP;
                expectedAccountParams.mpAccrued = stakeAmount + expectedBonusMP;
                expectedSystemState.totalMPAccrued += stakeAmount + expectedBonusMP;
                expectedAccountParams.maxMP = expectedMaxTotalMP;
                expectedSystemState.totalMaxMP += expectedMaxTotalMP;
            }
        } else {
            expectedRevert = FuzzTests__UndefinedError.selector;
            return;
        }
    }

    function _expectLock(address account, uint256 lockUpPeriod) internal {
        if (lockUpPeriod == 0) {
            expectedRevert = IStakeManager.StakeManager__DurationCannotBeZero.selector;
            return;
        }

        CheckVaultParams storage expectedAccountParams = expectedAccountState[account];
        if (expectedAccountParams.vaultBalance == 0) {
            expectedRevert = StakeMath.StakeMath__InsufficientBalance.selector;
            return;
        }

        if (lockUpPeriod > MAX_LOCKUP_PERIOD) {
            expectedRevert = StakeMath.StakeMath__InvalidLockingPeriod.selector;
            return;
        }

        uint256 calcLockEnd = Math.max(
            expectedVaultLockState[expectedAccountParams.account].lockEnd, vm.getBlockTimestamp()
        ) + lockUpPeriod;
        uint256 calcLockUpPeriod = calcLockEnd - vm.getBlockTimestamp();
        if (!(calcLockUpPeriod >= MIN_LOCKUP_PERIOD && calcLockUpPeriod <= MAX_LOCKUP_PERIOD)) {
            expectedRevert = FuzzTests__UndefinedError.selector;
            return;
        }
        if (expectedVaultLockState[expectedAccountParams.account].totalLockUp + lockUpPeriod > MAX_LOCKUP_PERIOD) {
            // total lock time surpassed the maximum allowed
            expectedRevert = StakeMath.StakeMath__AbsoluteMaxMPOverflow.selector;
        } else {
            expectedRevert = NO_REVERT;
            uint256 additionalBonusMP = _bonusMP(expectedAccountParams.vaultBalance, lockUpPeriod);
            expectedVaultLockState[expectedAccountParams.account].totalLockUp += lockUpPeriod;
            expectedVaultLockState[expectedAccountParams.account].lockEnd = calcLockEnd;
            expectedSystemState.totalMPStaked += additionalBonusMP;
            expectedSystemState.totalMPAccrued += additionalBonusMP;
            expectedSystemState.totalMaxMP += additionalBonusMP;
            expectedAccountParams.mpAccrued += additionalBonusMP;
            expectedAccountParams.maxMP += additionalBonusMP;
        }
    }

    function testFuzz_Stake(uint256 stakeAmount, uint64 lockUpPeriod) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        _expectStake(alice, stakeAmount, lockUpPeriod);
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        check("Stake: ", alice);
    }

    function testFuzz_Lock(uint256 stakeAmount, uint64 lockUpPeriod) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);

        _expectStake(alice, stakeAmount, 0);
        _stake(alice, stakeAmount, 0, expectedRevert);
        check("Stake:", alice);

        _expectLock(alice, lockUpPeriod);
        _lock(alice, lockUpPeriod, expectedRevert);
        check("Lock: ", alice);
    }

    function testFuzz_Relock(uint256 stakeAmount, uint64 lockUpPeriod, uint64 lockUpPeriod2) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);

        _expectStake(alice, stakeAmount, lockUpPeriod);
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        check("Stake: ", alice);

        _expectLock(alice, lockUpPeriod2);
        _lock(alice, lockUpPeriod2, expectedRevert);
        check("Lock: ", alice);
    }

    function testFuzz_AccrueMP(uint128 stakeAmount, uint64 lockUpPeriod, uint64 accruedTime) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);

        _expectStake(alice, stakeAmount, lockUpPeriod);
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        check("Stake: ", alice);

        _expectAccrue(alice, accruedTime);
        _accrue(alice, accruedTime);
        check("Accrue: ", alice);
    }

    function testFuzz_UpdateVault(uint128 stakeAmount, uint64 lockUpPeriod, uint64 accruedTime) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        _expectStake(alice, stakeAmount, lockUpPeriod);
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        check("Stake: ", alice);

        _expectAccrue(alice, accruedTime);
        _accrue(alice, accruedTime);
        check("Accrue: ", alice);
    }

    /**
     * uint256 public constant MIN_LOCKUP_PERIOD = 90 days; //7776000 seconds
     * uint256 public constant MAX_LOCKUP_PERIOD = MAX_MULTIPLIER * YEAR; // 126230400 seconds
     */
    function testFuzz_AccrueMP_Relock(
        uint128 stakeAmount,
        uint64 lockUpPeriod,
        uint64 lockUpPeriod2,
        uint64 accruedTime
    )
        public
    {
        // we're assuming stakeAmount to be > 10 to avoid cases where
        // balance and accrueAmount cause a deltaMP of 0,
        // which causes the test to emit false negatives
        vm.assume(stakeAmount > 1e1 && stakeAmount <= MAX_BALANCE);

        _expectStake(alice, stakeAmount, lockUpPeriod);
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        check("Stake: ", alice);

        _expectAccrue(alice, accruedTime);
        _accrue(alice, accruedTime);
        check("Accrue: ", alice);

        _expectLock(alice, lockUpPeriod2);
        _lock(alice, lockUpPeriod2, expectedRevert);
        check("Lock: ", alice);
    }

    function testFuzz_Unstake(
        uint128 stakeAmount,
        uint64 lockUpPeriod,
        uint16 accruedTime,
        uint128 unstakeAmount
    )
        public
    {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);

        _expectStake(alice, stakeAmount, lockUpPeriod);
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        check("Stake: ", alice);

        if (accruedTime > 0) {
            _expectAccrue(alice, accruedTime);
            _accrue(alice, accruedTime);
            check("Accrue: ", alice);
        }

        _expectUnstake(alice, unstakeAmount);
        _unstake(alice, unstakeAmount, expectedRevert);
        check("Unstake: ", alice);
    }

    function testFuzz_Rewards(
        uint256 stakeAmount,
        uint256 lockUpPeriod,
        uint256 rewardAmount,
        uint16 rewardPeriod,
        uint16 accountRewardPeriod
    )
        public
    {
        stakeAmount = bound(stakeAmount, 1e18, 20_000_000e18);
        lockUpPeriod = lockUpPeriod == 0 ? 0 : bound(lockUpPeriod, MIN_LOCKUP_PERIOD, MAX_LOCKUP_PERIOD);
        vm.assume(rewardPeriod > 0 && rewardPeriod <= 12 weeks); // assuming max 3 months
        vm.assume(rewardAmount > 1e18 && rewardAmount <= 1_000_000e18); // assuming max 1_000_000 Karma
        vm.assume(accountRewardPeriod <= rewardPeriod); // Ensure accountRewardPeriod doesn't exceed rewardPeriod

        uint256 initialTime = vm.getBlockTimestamp();
        uint256 tolerance = 1000;

        // Calculate expected reward using safe math operations
        uint256 expectedReward = accountRewardPeriod < rewardPeriod
            ? Math.mulDiv(accountRewardPeriod, rewardAmount, rewardPeriod)
            : rewardAmount;

        expectedRevert = NO_REVERT;

        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);

        _setRewards(rewardAmount, rewardPeriod);

        vm.warp(initialTime + accountRewardPeriod);

        assertEq(streamer.totalRewardsSupply(), expectedReward, "Total rewards supply mismatch");
        assertApproxEqAbs(
            streamer.rewardsBalanceOf(vaults[alice]), expectedReward, tolerance, "Reward balance mismatch"
        );
    }

    function testFuzz_EmergencyExit(uint256 stakeAmount, uint256 lockUpPeriod) public {
        vm.assume(stakeAmount > 0 && stakeAmount <= MAX_BALANCE);
        vm.assume(lockUpPeriod == 0 || (lockUpPeriod >= MIN_LOCKUP_PERIOD && lockUpPeriod <= MAX_LOCKUP_PERIOD));

        uint256 aliceInitialBalance = stakingToken.balanceOf(alice);
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 expectedMaxTotalMP = _maxTotalMP(stakeAmount, lockUpPeriod);
        expectedRevert = NO_REVERT;
        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);

        vm.prank(admin);
        streamer.enableEmergencyMode();

        _emergencyExit(alice);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount + expectedBonusMP,
                totalMPAccrued: stakeAmount + expectedBonusMP,
                totalMaxMP: expectedMaxTotalMP,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: stakeAmount + expectedBonusMP,
                maxMP: expectedMaxTotalMP,
                rewardsAccrued: 0
            })
        );

        assertEq(
            stakingToken.balanceOf(alice), aliceInitialBalance + stakeAmount, "Alice should get staked tokens back"
        );
    }

    function testFuzz_RedeemRewards(
        uint256 stakeAmount,
        uint256 lockUpPeriod,
        uint256 rewardAmount,
        uint16 rewardPeriod,
        uint16 accountRewardPeriod
    )
        public
    {
        stakeAmount = bound(stakeAmount, 1e18, 20_000_000e18);
        vm.assume(rewardPeriod > 0 && rewardPeriod <= 12 weeks); // assuming max 3 months
        vm.assume(rewardAmount > 1e18 && rewardAmount <= 1_000_000e18); // assuming max 1_000_000 Karma
        vm.assume(accountRewardPeriod <= rewardPeriod); // Ensure accountRewardPeriod doesn't exceed rewardPeriod
        lockUpPeriod = lockUpPeriod == 0 ? 0 : bound(lockUpPeriod, MIN_LOCKUP_PERIOD, MAX_LOCKUP_PERIOD);
        uint256 initialTime = vm.getBlockTimestamp();
        uint256 tolerance = 1000;

        // Calculate expected reward using safe math operations
        uint256 expectedReward = accountRewardPeriod < rewardPeriod
            ? Math.mulDiv(accountRewardPeriod, rewardAmount, rewardPeriod)
            : rewardAmount;

        expectedRevert = NO_REVERT;

        _stake(alice, stakeAmount, lockUpPeriod, expectedRevert);
        _setRewards(rewardAmount, rewardPeriod);

        uint256 streamerKarmaBalanceBefore = karma.balanceOfRewardDistributor(address(streamer));
        assertEq(streamerKarmaBalanceBefore, rewardAmount);

        vm.warp(initialTime + accountRewardPeriod);

        uint256 redeemed = streamer.redeemRewards(vaults[alice]);

        assertEq(streamer.totalRewardsSupply(), expectedReward, "Total rewards supply mismatch");
        assertApproxEqAbs(karma.balanceOf(alice), expectedReward, tolerance, "Reward balance mismatch");
        assertEq(karma.balanceOfRewardDistributor(address(streamer)), streamerKarmaBalanceBefore - redeemed);
    }
}