pragma solidity ^0.8.26;

import { StakeTest, StakeVault, StakeManager } from "./Stake.t.sol";

contract UnstakeTest is StakeTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_RevertWhen_FundsLocked() public {
        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD();

        _stake(alice, stakeAmount, lockUpPeriod);

        // Alice tries to unstake before lock up period has expired
        vm.expectRevert(StakeVault.StakeVault__FundsLocked.selector);
        _unstake(alice, stakeAmount);

        vm.warp(vm.getBlockTimestamp() + lockUpPeriod);

        // Alice unstake after lock up period has expired
        _unstake(alice, stakeAmount);
    }

    function test_UnstakeOneAccount() public {
        test_StakeOneAccount();

        _unstake(alice, 8e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 2e18,
                totalMPStaked: 2e18,
                totalMPAccrued: 2e18,
                totalMaxMP: 10e18,
                stakingBalance: 2e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 2e18,
                vaultBalance: 2e18,
                rewardIndex: 0,
                mpAccrued: 2e18,
                maxMP: 10e18,
                rewardsAccrued: 0
            })
        );

        _unstake(alice, 2e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMPStaked: 0,
                totalMPAccrued: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_UnstakeOneAccountAndAccruedMP() public {
        test_StakeOneAccount();

        // wait for 1 year
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + (YEAR));

        streamer.updateVault(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 10e18,
                totalMPStaked: 20e18,
                totalMPAccrued: 20e18, // total MP must have been doubled
                totalMaxMP: 50e18,
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // unstake half of the tokens
        _unstake(alice, 5e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 5e18, // 10 - 5
                totalMPStaked: 10e18,
                totalMPAccrued: 10e18, // 20 - 10 (5 initial + 5 accrued)
                totalMaxMP: 25e18,
                stakingBalance: 5e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_UnstakeOneAccountWithLockUpAndAccruedMP() public {
        test_StakeOneAccountWithMinLockUp();

        uint256 stakeAmount = 10e18;
        uint256 lockUpPeriod = streamer.MIN_LOCKUP_PERIOD();
        // 10e18 is what's used in `test_StakeOneAccountWithMinLockUp`
        uint256 expectedBonusMP = _bonusMP(stakeAmount, lockUpPeriod);
        uint256 unstakeAmount = 5e18;
        uint256 warpLength = (365 days);
        // wait for 1 year
        uint256 currentTime = vm.getBlockTimestamp();

        vm.warp(currentTime + (warpLength));

        streamer.updateVault(vaults[alice]);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: (stakeAmount + expectedBonusMP) + stakeAmount,
                totalMPAccrued: (stakeAmount + expectedBonusMP) + stakeAmount, // we do `+ stakeAmount` we've accrued
                // `stakeAmount` after 1 year
                totalMaxMP: _maxTotalMP(stakeAmount, lockUpPeriod),
                stakingBalance: 10e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
        uint256 newBalance = stakeAmount - unstakeAmount;
        // unstake half of the tokens
        _unstake(alice, unstakeAmount);

        uint256 expectedTotalMP =
            _initialMP(newBalance) + _bonusMP(newBalance, lockUpPeriod) + _accrueMP(newBalance, warpLength);
        checkStreamer(
            CheckStreamerParams({
                totalStaked: newBalance,
                totalMPStaked: expectedTotalMP,
                totalMPAccrued: expectedTotalMP,
                totalMaxMP: _maxTotalMP(newBalance, lockUpPeriod),
                stakingBalance: newBalance,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );
    }

    function test_UnstakeOneAccountAndRewards() public {
        test_StakeOneAccountAndRewards();

        _unstake(alice, 8e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 2e18,
                totalMPStaked: 2e18,
                totalMPAccrued: 2e18,
                totalMaxMP: 10e18,
                stakingBalance: 2e18,
                rewardBalance: 0, // rewards are all paid out to alice
                rewardIndex: 50e18
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 1000e18,
                stakedBalance: 2e18,
                vaultBalance: 2e18,
                rewardIndex: 50e18, // alice reward index has been updated
                mpAccrued: 2e18,
                maxMP: 10e18,
                rewardsAccrued: 0
            })
        );
    }

    function test_UnstakeBonusMPAndAccuredMP() public {
        // setup variables
        uint256 amountStaked = 10e18;
        uint256 secondsLocked = streamer.MIN_LOCKUP_PERIOD();
        uint256 reducedStake = 5e18;
        uint256 increasedTime = YEAR;

        //initialize memory placehodlders
        uint256[4] memory timestamp;
        uint256[4] memory increasedAccuredMP;
        uint256[4] memory predictedBonusMP;
        uint256[4] memory predictedAccuredMP;
        uint256[4] memory predictedTotalMP;
        uint256[4] memory predictedTotalMaxMP;
        uint256[4] memory totalStaked;

        //stages variables setup
        uint256 stage = 0; // first stage: initialization
        {
            timestamp[stage] = block.timestamp;
            totalStaked[stage] = amountStaked;
            predictedBonusMP[stage] = totalStaked[stage] + _bonusMP(totalStaked[stage], secondsLocked);
            predictedTotalMaxMP[stage] = _maxTotalMP(totalStaked[stage], secondsLocked);
            increasedAccuredMP[stage] = 0; //no increased accured MP in first stage
            predictedAccuredMP[stage] = 0; //no accured MP in first stage
            predictedTotalMP[stage] = predictedBonusMP[stage] + predictedAccuredMP[stage];
        }
        stage++; // second stage: progress in time
        {
            timestamp[stage] = timestamp[stage - 1] + increasedTime;
            totalStaked[stage] = totalStaked[stage - 1];
            predictedBonusMP[stage] = predictedBonusMP[stage - 1]; //no change in bonusMP in second stage
            predictedTotalMaxMP[stage] = predictedTotalMaxMP[stage - 1];
            // solhint-disable-next-line max-line-length
            increasedAccuredMP[stage] = _accrueMP(totalStaked[stage], timestamp[stage] - timestamp[stage - 1]);
            predictedAccuredMP[stage] = predictedAccuredMP[stage - 1] + increasedAccuredMP[stage];
            predictedTotalMP[stage] = predictedBonusMP[stage] + predictedAccuredMP[stage];
        }
        stage++; //third stage: reduced stake
        {
            timestamp[stage] = timestamp[stage - 1]; //no time increased in third stage
            totalStaked[stage] = totalStaked[stage - 1] - reducedStake;
            //bonusMP from this stage is a proportion from the difference of remainingStake and amountStaked
            //if the account reduced 50% of its stake, the bonusMP should be reduced by 50%
            predictedBonusMP[stage] = (totalStaked[stage] * predictedBonusMP[stage - 1]) / totalStaked[stage - 1];
            predictedTotalMaxMP[stage] = (totalStaked[stage] * predictedTotalMaxMP[stage - 1]) / totalStaked[stage - 1];
            increasedAccuredMP[stage] = 0; //no accuredMP in third stage;
            //total accuredMP from this stage is a proportion from the difference of remainingStake and amountStaked
            //if the account reduced 50% of its stake, the accuredMP should be reduced by 50%
            predictedAccuredMP[stage] = (totalStaked[stage] * predictedAccuredMP[stage - 1]) / totalStaked[stage - 1];
            predictedTotalMP[stage] = predictedBonusMP[stage] + predictedAccuredMP[stage];
        }

        // stages execution
        stage = 0; // first stage: initialization
        {
            _stake(alice, amountStaked, secondsLocked);
            {
                StakeManager.VaultData memory vaultData = streamer.getVault(vaults[alice]);
                assertEq(vaultData.stakedBalance, totalStaked[stage], "stage 1: wrong account staked balance");
                assertEq(vaultData.mpAccrued, predictedTotalMP[stage], "stage 1: wrong account MP");
                assertEq(vaultData.maxMP, predictedTotalMaxMP[stage], "stage 1: wrong account max MP");

                assertEq(streamer.totalStaked(), totalStaked[stage], "stage 1: wrong total staked");
                assertEq(streamer.totalMPAccrued(), predictedTotalMP[stage], "stage 1: wrong total MP");
                assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 1: wrong totalMaxMP MP");
            }
        }

        stage++; // second stage: progress in time
        vm.warp(timestamp[stage]);
        streamer.updateVault(vaults[alice]);
        {
            StakeManager.VaultData memory vaultData = streamer.getVault(vaults[alice]);
            assertEq(vaultData.stakedBalance, totalStaked[stage], "stage 2: wrong account staked balance");
            assertEq(vaultData.mpAccrued, predictedTotalMP[stage], "stage 2: wrong account MP");
            assertEq(vaultData.maxMP, predictedTotalMaxMP[stage], "stage 2: wrong account max MP");

            assertEq(streamer.totalStaked(), totalStaked[stage], "stage 2: wrong total staked");
            assertEq(streamer.totalMPAccrued(), predictedTotalMP[stage], "stage 2: wrong total MP");
            assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 2: wrong totalMaxMP MP");
        }

        stage++; // third stage: reduced stake
        _unstake(alice, reducedStake);
        {
            StakeManager.VaultData memory vaultData = streamer.getVault(vaults[alice]);
            assertEq(vaultData.stakedBalance, totalStaked[stage], "stage 3: wrong account staked balance");
            assertEq(vaultData.mpAccrued, predictedTotalMP[stage], "stage 3: wrong account MP");
            assertEq(vaultData.maxMP, predictedTotalMaxMP[stage], "stage 3: wrong account max MP");

            assertEq(streamer.totalStaked(), totalStaked[stage], "stage 3: wrong total staked");
            assertEq(streamer.totalMPAccrued(), predictedTotalMP[stage], "stage 3: wrong total MP");
            assertEq(streamer.totalMaxMP(), predictedTotalMaxMP[stage], "stage 3: wrong totalMaxMP MP");
        }
    }

    function test_UnstakeMultipleAccounts() public {
        test_StakeMultipleAccounts();

        _unstake(alice, 10e18);
        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMPStaked: 20e18,
                totalMPAccrued: 20e18,
                totalMaxMP: 100e18,
                stakingBalance: 20e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 0,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 0,
                mpAccrued: 20e18,
                maxMP: 100e18,
                rewardsAccrued: 0
            })
        );
    }

    function test_UnstakeMultipleAccountsAndRewards() public {
        test_StakeMultipleAccountsAndRewards();

        _unstake(alice, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 30e18,
                totalMPStaked: 30e18,
                totalMPAccrued: 30e18,
                totalMaxMP: 150e18,
                stakingBalance: 30e18,
                // alice owned a 25% of the pool, so 25% of the rewards are paid out to alice (250)
                rewardBalance: 750e18,
                rewardIndex: 125e17 // reward index remains unchanged
             })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 250e18,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 125e17,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        _unstake(bob, 10e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 20e18,
                totalMPStaked: 20e18,
                totalMPAccrued: 20e18,
                totalMaxMP: 100e18,
                stakingBalance: 20e18,
                rewardBalance: 0, // bob should've now gotten the rest of the rewards
                rewardIndex: 125e17
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 750e18,
                stakedBalance: 20e18,
                vaultBalance: 20e18,
                rewardIndex: 125e17,
                mpAccrued: 20e18,
                maxMP: 100e18,
                rewardsAccrued: 0
            })
        );

        _unstake(bob, 20e18);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 0,
                totalMPStaked: 0,
                totalMPAccrued: 0,
                totalMaxMP: 0,
                stakingBalance: 0,
                rewardBalance: 0,
                rewardIndex: 125e17
            })
        );

        checkVault(
            CheckVaultParams({
                account: vaults[bob],
                rewardBalance: 750e18,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 125e17,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );
    }
}