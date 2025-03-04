# Staking Protocol [![Github Actions][gha-badge]][gha] [![Codecov][codecov-badge]][codecov] [![Foundry][foundry-badge]][foundry]

[gha]: https://github.com/vacp2p/staking-reward-streamer/actions
[gha-badge]: https://github.com/vacp2p/staking-reward-streamer/actions/workflows/test.yml/badge.svg
[codecov]: https://codecov.io/gh/vacp2p/staking-reward-streamer
[codecov-badge]: https://codecov.io/gh/vacp2p/staking-reward-streamer/graph/badge.svg
[foundry]: https://getfoundry.sh/
[foundry-badge]: https://img.shields.io/badge/Built%20with-Foundry-FFDB1C.svg

Smart contracts for staking and rewards distribution using Karma token.

This projects implements smart contracts that are used by Status to distribute rewards to users who stake their tokens.

# Deployments

| **Contract**                | **Address**                                                                                                                                   | **Snapshot**                                                                                                      |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| **Sepolia**                 |                                                                                                                                               |                                                                                                                   |
| StakeManager      | [`0x223532449d4cceBD432043aDb1CA0af642A2b3e0`](https://sepolia.etherscan.io/address/0x223532449d4cceBD432043aDb1CA0af642A2b3e0#code)                       | [`aa3442b`](https://github.com/status-im/communities-contracts/commit/aa3442b)   |
| StakeManagerProxy | [`0x2B862e47E4743D929Da90998f1Ec2465DA184Dad`](https://sepolia.etherscan.io/address/0x2B862e47E4743D929Da90998f1Ec2465DA184Dad)                       | [`aa3442b`](https://github.com/status-im/communities-contracts/commit/aa3442b)   |
| VaultFactory | [`0x899da2e9f6C8fbA95d9F1dD5a0C984F2435ab8e0`](https://sepolia.etherscan.io/address/0x899da2e9f6C8fbA95d9F1dD5a0C984F2435ab8e0)                       | [`aa3442b`](https://github.com/status-im/communities-contracts/commit/aa3442b)   |
| Karma | [`0x1e1Be9175AA9f135Fe986Ef9b43421F6685c65FA`](https://sepolia.etherscan.io/address/0x1e1be9175aa9f135fe986ef9b43421f6685c65fa#readContract)                       | [`aa3442b`](https://github.com/status-im/communities-contracts/commit/aa3442b)   |
| KarmaProxy | [`0x2eE435C111C1c04d1698870f3300B77F5c7f30Eb`](https://sepolia.etherscan.io/address/0x2eE435C111C1c04d1698870f3300B77F5c7f30Eb)                       | [`aa3442b`](https://github.com/status-im/communities-contracts/commit/aa3442b)   |
| **Status Network Sepolia**                 |                                                                                                                                               |                                                                                                                   |
| StakeManager      | [`0xE452027cdEF746c7Cd3DB31CB700428b16cD8E51`](https://sepoliascan.status.network/address/0xE452027cdEF746c7Cd3DB31CB700428b16cD8E51)                       | [`aa1addb`](https://github.com/vacp2p/staking-reward-streamer/commit/aa1addbfcd240f7e64050ffc4eba8399e40617a5)   |
| StakeManagerProxy | [`0x785e6c5af58FB26F4a0E43e0cF254af10EaEe0f1`](https://sepoliascan.status.network/address/0x785e6c5af58FB26F4a0E43e0cF254af10EaEe0f1?tab=txs)                       | [`aa1addb`](https://github.com/vacp2p/staking-reward-streamer/commit/aa1addbfcd240f7e64050ffc4eba8399e40617a5)   |
| VaultFactory | [`0xf7b6EC76aCa97b395dc48f7A2861aD810B34b52e`](https://sepoliascan.status.network/address/0xf7b6EC76aCa97b395dc48f7A2861aD810B34b52e)                       | [`aa1addb`](https://github.com/vacp2p/staking-reward-streamer/commit/aa1addbfcd240f7e64050ffc4eba8399e40617a5)   |
| Karma | [`0x0936792b0efa243a5Ddff7035E84749E5a54FA9c`](https://sepoliascan.status.network/address/0x0936792b0efa243a5Ddff7035E84749E5a54FA9c)                       | [`aa1addb`](https://github.com/vacp2p/staking-reward-streamer/commit/aa1addbfcd240f7e64050ffc4eba8399e40617a5)   |
| KarmaProxy | [`0x59510D0b235c75d7bCAEb66A420e9bb0edC976AE`](https://sepoliascan.status.network/address/0x59510D0b235c75d7bCAEb66A420e9bb0edC976AE)                       | [`aa1addb`](https://github.com/vacp2p/staking-reward-streamer/commit/aa1addbfcd240f7e64050ffc4eba8399e40617a5)   |
| KarmaNFT | [`0x1E9E85e91deF9a9aCf1d6F2888033180e4673d57`](https://sepoliascan.status.network/address/0x1E9E85e91deF9a9aCf1d6F2888033180e4673d57?tab=contract)                       | [`aa1addb`](https://github.com/vacp2p/staking-reward-streamer/commit/aa1addbfcd240f7e64050ffc4eba8399e40617a5)   |

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
