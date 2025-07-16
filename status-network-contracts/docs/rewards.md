# Earning Rewards

Accounts participate in the staking protocol to earn rewards. The staking system incorporates a rewards distribution
mechanism that works in conjunction with the [Multiplier Points (MP)](multiplier-points.md) system.

In this document we'll discuss how rewards are calculated and distributed to the participants of the system.

## How Rewards Work

Generally, rewards are set by the admin or owner of the protocol. They decide how many rewards are distributed over a
specific period and the rate at which they are distributed.

For example, if 1000 tokens are to be distributed over the next 10 days, the rate would be 100 tokens per day. These
tokens are then distributed to accounts based on their relative weights in the system.

In other words, if Alice and Bob both own exactly 50% of the total stake in the system (including MP), they would each
receive 50% of the rewards, or 5 tokens per day in this example.

This is a rather trivial example. In practice we'll run into much more complex scenarios. Alice and Bob might have the
same stake, but participated in the system for different amounts of time. Or, one of them might have unstaked some
amount at a later pointer in tame and then re-entered the system by increasing the stake again.

All these factors need to be taken into account when calculating rewards. Let's take a closer look at how this works.

### Reward rate and reward distribution

We've already mentioned that rewards are configured by the admin of the system by specifying the total amount of rewards
and the duration over which they are distributed.

With a total reward of 1000 tokens and a reward period of 10 days, the rate would be 100 reward tokens per day:

$$
\text{Reward Rate} = \frac{\text{Total Reward Amount}}{\text{Reward Duration}}
$$

$$
\text{Reward Rate} = \frac{1000}{\text{10 days}} = 100 \text{ tokens/day}
$$

In fact, just like the underlying [Multiplier Points](multiplier-points.md), rewards are accrued over time and grow by
the second. This is an important characteristic of the system, as it allows for real-time rewards without the need for
accounts to interact with the system.

The next question is, how will these rewards be distributed among the participants?

The system distributes rewards based on tree main factors:

- The amount of tokens staked
- The account's Multiplier Points (MP)
- The accounts reward index (more on that in a bit)

Every account in the system has a "weight" associated with it. This weight is calculated as the sum of the staked tokens
and the MP:

$$
\text{Account Weight} = \text{Staked Balance} + \text{MP Balance}
$$

In addition, the system keeps track of the total weight of all accounts in the system:

$$
\text{Total System Weight} = \sum_{i=1}^{n} \text{Account Weight}_i
$$

One thing to keep in mind here, is that account MP grow linearly with time, which means, account weights increase with
time, which in turn means, the total weight of the system increases with time as well.

With the weights in place, the system can now calculate the rewards for each account. For an individual account, it's as
simple as dividing the account's weight by the total system weight and multiplying it by the total reward amount:

$$
\text{Reward Amount} = \text{Account Weight} \times \frac{\text{Total Reward Amount}}{\text{Total System Weight}}
$$

To plug in some numbers, let's assume Alice has a weight of 150 and Bob has a weight of 350. The total system weight
is 500. If the total reward amount is 1000 tokens, the rewards for Alice and Bob would be:

$$
\text{Alice Reward} = 150 \times \frac{1000}{500} = 300
\text{Bob Reward} = 350 \times \frac{1000}{500} = 700
$$

### Claiming rewards

Because rewards are calculated in real-time, they don't actually need to be claimed by any account. However, whenever an
account performs **state changing** actions, such as staking or unstaking funds, the rewards that have been accrued up
to this point are updated in the account's storage.

We consider any real-time rewards "pending" rewards until the account interacts with the system. These pending rewards
are calculated as:

$$
\text{Pending Rewards} = \text{Account Weight} \times \left( \text{Current Reward Index} \minus \text{Account's Last Reward Index} \right)
$$

The indicies are a new concept we'll cover next.

## Reward Indices

The reward system uses a special accounting mechanism based on indices to track and distribute rewards accurately. This
approach ensures fair distribution even as accounts enter and exit the system or change their stakes over time.

### Global Reward Index

The global reward index represents the cumulative rewards per unit of weight since the system's inception. It increases
whenever new rewards are added to the system:
$
\text{New Index} = \text{Current Index} + \frac{\text{New Rewards} \times \text{Scale Factor}}{\text{Total System Weight}}
$

Where:

- `Scale Factor` is 1e18 (used for precision)
- `Total System Weight` is the sum of all staked tokens and MP in the system

### Account Reward Indices

Each account maintains its own reward index, which represents the point at which they last "claimed" rewards (performed
a state changing action). The difference between the global index and an account's index, multiplied by the account's
weight, determines their unclaimed rewards:

$$
\text{Unclaimed Rewards} = \text{Account Weight} \times (\text{Global Index} - \text{Account Index})
$$

### Reward Index Updates

The system updates indices in the following situations:

- When new rewards are added to the system
- When accounts stake or unstake tokens

This mechanism ensures that, historical rewards are preserved accurately and accounts receive their fair share based on
their weight over time. Also, new accounts don't receive rewards from before they entered the system.

### Index Adjustment Example

Let's say we have a system with:

- Total Weight: 1000 units
- Current Index: 0.5
- New Rewards: 100 tokens

The index would increase by:

```
Increase = (100 × 1e18) / 1000 = 0.1e18
New Index = 0.5e18 + 0.1e18 = 0.6e18
```

If an account has:

- Weight: 200 units
- Last Index: 0.5e18

Their unclaimed rewards would be:

```
Unclaimed = 200 × (0.6e18 - 0.5e18) / 1e18 = 20 tokens
```

## Summary

The reward system provides a fair and transparent method for distributing rewards based on both staked amounts and MP.
It automatically handles:

- Linear distribution of rewards over time
- Fair allocation based on relative weights
- Accurate tracking of individual and global reward states
