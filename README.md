# Rewards Streaming

- [Rewards Streamer](#rewards-streamer)
- [Rewards Streamer with Multiplier Points](#rewards-streamer-with-multiplier-points)

## Rewards Streamer

The rewardIndex is a crucial part of the reward distribution mechanism.

It represents the accumulated rewards per staked token since the beginning of the contract's operation.

Here's how it works:

1. Initial state: When the contract starts, rewardIndex is 0.
2. Whenever new rewards are added to the contract (detected in updateRewardIndex()), the rewardIndex increases. The
   increase is calculated as: `rewardIndex += (newRewards * SCALE_FACTOR) / totalStaked` This calculation distributes
   the new rewards evenly across all staked tokens.
3. Each user has their own userRewardIndex, which represents the global rewardIndex at the time of their last
   interaction (stake, unstake, or reward claim).
4. When a user wants to claim rewards, we calculate the difference between the current rewardIndex and the user's
   userRewardIndex, multiply it by their staked balance, and divide by SCALE_FACTOR
5. After a user stakes, unstakes, or claims rewards, their userRewardIndex is updated to the current global rewardIndex.
   This "resets" their reward accumulation for the next period.

Instead of updating each user's rewards every time new rewards are added, we only need to update a single global
variable (rewardIndex).

We don't need to assign Rewards to epochs, so we don't need to finalize Rewords for each epoch and each user.

User-specific calculations are done only when a user interacts with the contract.

`SCALE_FACTOR` is used to maintain precision in calculations.

### Rewards Streamer Example

![example](https://github.com/user-attachments/assets/970dbb89-6163-494e-8276-358c5c405566)

**Initial setup:**

- rewardIndex: 0
- accountedRewards: 0
- Rewards in contract: 0

**T1: Alice stakes 10 tokens**

- Alice's userRewardIndex: 0
- Alice's staked tokens: 10
- totalStaked: 10 tokens

**T2: Bob stakes 30 tokens**

- Alice's userRewardIndex: 0
- Bob's userRewardIndex: 0
- Alice's staked tokens: 10
- Bob's staked tokens: 30
- totalStaked: 40 tokens

**T3: 1000 Rewards arrive**

New rewardIndex calculation:

- newRewards = 1000
- rewardIndex increase = 1000 / 40 = 25
- rewardIndex = 0 + 25 = 25
- accountedRewards: 1000
- Rewards in contract: 1000

Potential Rewards for Alice and Bob:

For Alice:

- Staked amount: 10 tokens
- Potential Rewards: 10 \* (25 - 0) = 250

For Bob:

- Staked amount: 30 tokens
- Potential Rewards: 30 \* (25 - 0) = 750

**T4: Alice withdraws her stake and Rewards**

Alice's withdrawal:

- tokens returned: 10
- Rewards: 250

Update state:

- totalStaked = 40 - 10 = 30 tokens
- Rewards in contract = 1000 - 250 = 750

**T5: Charlie stakes 30 tokens**

- Charlie's userRewardIndex: 25
- totalStaked = 30 + 30 = 60 tokens

**T6: Another 1000 Rewards arrive**

New rewardIndex calculation:

- newRewards = 1000
- rewardIndex increase = 1000 / 60 = 16.67
- new rewardIndex = 25 + 16.67 = 41.67
- accountedRewards: 1000 + 1000 = 2000
- Rewards in contract = 750 + 1000 = 1750

Rewards for Bob and Charlie:

For Bob:

- Staked amount: 30 tokens
- Potential Rewards: 30 \* (41.67 - 0) = 1250.1 // rounding error
  - In bucket 1: 30 \* (25 - 0) = 750
  - In bucket 2: 30 \* (16.67 - 0) = 500.1
  - Total of b1 + b2: 750 + 500.1 = 1250.1
  - Which is equal to
    - 30 \* ( (25 - 0) + (41.67 - 25) )

For Charlie:

- Staked amount: 30 tokens
- Potential Rewards: 30 \* (41.67 - 25) = 500.1 // rounding error

If Bob and Charlie were to withdraw now:

Bob's withdrawal:

- tokens returned: 30
- Rewards: 1250.1
- Rewards in contract after Bob's withdrawal: 1750 - 1250.1 = 499.9

Charlie's withdrawal:

- tokens returned: 30
- Rewards: 499.9
- Rewards in contract after Charlie's withdrawal: 499.9 - 499.9 = 0

**T7: Final state:**

- Alice received: 10 tokens and 250 Rewards
- Bob received: 30 tokens and 1250.1 Rewards
- Charlie received: 30 tokens and 499.9 Rewards
- Total Rewards distributed: 2000 Rewards
- Rewards remaining in contract: 0

## Rewards Streamer with Multiplier Points

TODO
