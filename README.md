## StakeManager

The rewardIndex is a crucial part of the reward distribution mechanism.

It represents the accumulated rewards per staked token since the beginning of the contract's operation.

Here's how it works:

1 - Initial state: When the contract starts, rewardIndex is 0.
2 - Whenever new rewards are added to the contract (detected in updateRewardIndex()), the rewardIndex increases.
    The increase is calculated as:
    `rewardIndex += (newRewards * SCALE_FACTOR) / totalStaked`
    This calculation distributes the new rewards evenly across all staked tokens.
3 - Each user has their own userRewardIndex, which represents the global rewardIndex at the time
    of their last interaction (stake, unstake, or reward claim).
4 - When a user wants to claim rewards, we calculate the difference between the current rewardIndex
    and the user's userRewardIndex, multiply it by their staked balance, and divide by SCALE_FACTOR
5 - After a user stakes, unstakes, or claims rewards, their userRewardIndex is updated to the current
    global rewardIndex. This "resets" their reward accumulation for the next period.

Instead of updating each user's rewards every time new rewards are added, we only need to update a
single global variable (rewardIndex).

User-specific calculations are done only when a user interacts with the contract.

SCALE_FACTOR is used to maintain precision in calculations.
