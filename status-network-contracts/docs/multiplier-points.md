# Understanding Multiplier Points

This document details the the internal accounting for what are called "Multiplier Points" ("MP"), which are used to
ensure that participants in the staking system are rewarded fairly based on their stake amount and time staked.

We'll cover how MP are calculated, how the stake amount and lock-up duration affect MP, and how MP influences the
rewards participants earn.

## What and why

Generally, accounts can freely participate in the staking system by depositing funds into a staking vault. The longer
they keep their funds staked, the more rewards they earn. In addition, accounts can lock up their funds for a certain
period which significantly increases the amount of rewards they receive.

To ensure rewards are distributed fairly, the system uses MP for internal accounting. Accounts that stake longer, stake
more, or lock up their stake in the vault, accrue more MP than accounts that participate not as long, with less stake,
or a shorter lock-up period.

This enables use cases where accounts can have a significant increase in voting power, even though they might not have
as much stake as others, but they have still committed to the system for a longer period of time.

That being said, MP are not transferable and are only used for internal accounting purposes. End users will only see the
effect of MP in the rewards they receive.

## Initial MP and Accrued MP

First of all, it's important to understand that the MP, that vaults accumulate, are divided into two parts: **Initial
MP** and **Accrued MP**.

- **Initial MP** is the MP that are issued immediately when tokens are staked (with or without a lock-up period). It is
  based on the stake amount and lock-up duration.
- **Accrued MP** are the MP that accumulate over time as a function of the stake amount, elapsed time, and annual
  percentage yield (APY).

Furthermore, the initial MP could be divided into "initial" and "bonus" MP, as locking up the stake is rewarded with a
bonus on top of the initial MP. For simplicity's sake, we treat it as one value in this document.

### Initial MP formula

The formula for Initial MP is derived as follows:

$$
\text{MP}_ \text{Initial} = \text{Stake} \times \left( 1 + \frac{\text{APY} \times T_ \text{lock}}{100 \times T_ \text{year}} \right)
$$

Where:

- $Stake$: The amount of tokens staked.
- $APY$: Annual Percentage Yield, set at 100%.
- $T_{lock}$: Lock-up duration in seconds.
- $T_{year}$: Total seconds in a year.

This formula calculates the MP issued immediately when tokens are staked with a lock-up period. The longer the lock-up
period, the more MP are issued.

Here are some examples.

#### Alice stakes 100 tokens with no lock-up time

$$
\text{MP}_ \text{Initial} = 100 \times \left( 1 + \frac{100 \times 0}{100 \times 365} \right)
$$

$$
\text{MP}_ \text{Initial} = 100 \times \left( 1 + 0\right) = 100
$$

Alice receives 100 MP.

#### Alice stakes 100 tokens with a 30 days lock-up period

$$
\text{MP}_ \text{Initial} = 100 \times \left( 1 + \frac{100 \times 30}{100 \times 365} \right)
$$

$$
\text{MP}_ \text{Initial} = 100 \times \left( 1 + 0.082 \right) = 108.2
$$

Alice receives 108.2 MP. Notice how, just by locking up the stake for 30 days, Alice receives an additional 8.2 MP right
away. In return, she cannot access her funds until the lock-up period has passed.

### Accrued MP formula

Accrued MP is calculated for time elapsed as:

$$
\text{MP}_ \text{Accrued} = \text{Stake} \times \frac{\text{APY} \times T_ \text{elapsed}}{100 \times T_ \text{year}}
$$

Where:

- $T_{elapsed}$: Time elapsed since staking began, measured in seconds.

This formula adds MP as a function of time, rewarding users who keep their stake locked. Notice that the accrued MP are
always calculated based on the stake amount. Already accrued MP do not affect the calculation of new accrued MP.

Here are some examples.

#### Alice stakes 100 tokens for 15 days

$$
\text{MP}_ \text{Accrued} = 100 \times \frac{100 \times 15}{100 \times 365} = 4.1
$$

Alice receives 4.1 MP for the 15 days she has staked. This is exactly half of the MP she would have received if she had
locked her stake for 30 days as in the previous example.

#### Alice stakes 100 tokens for 30 days

$$
\text{MP}_ \text{Accrued} = 100 \times \frac{100 \times 30}{100 \times 365} = 8.2
$$

Alice receives 8.2 MP for the 30 days she has staked.

### On real-time MP accrual

As mentioned a couple of times above, the more time has elapsed, the more MP are accrued. In fact, MP are increase every
second and can be monitored in real-time via the smart contracts. Users have to "claim" their accrued MP by calling a
function on the stake manager contract.

This means that, unless the maximum amount of MP has been reached (more on that below), the MP amount in storage will
likely be different from the real-time value.

### Total MP

Total MP combines both accrued MP and pending MP. The accrued MP contain the initial MP and the MP accrued over time.
Pending MP are the ones that have yet to be "claimed" by the user:

$$
\text{MP}_ \text{Total} = \text{MP}_ \text{Accrued} + \text{MP}_ \text{Pending}
$$

This total is used to calculate the user’s share of rewards, which we'll cover in another chapter.

## Maximum MP

One additional concept we need to be aware of is the total maximum amount of MP an account can accrue. This is important
to prevent accounts to accumulated massive amounts of MP over time, making it impossible for other participants to catch
up.

Generally, the maximum amount of MP an account can accrue is capped at:

$$
\text{MP}_\text{Maximum} = \text{MP}_ \text{Initial} + \text{MP}_ \text{Potential}
$$

- $\text{MP}_ \text{Initial}$: The initial MP an account receives when staking, \*_including the bonus MP_.
- $\text{MP}_ \text{Potential}$: The initial MP amount multiplied by a $MAX\_MULTIPLIER$ factor.
- $MAX\_MULTIPLIER$: A constant that determines the multiplier for the maximum amount of MP in the system.

For example, assuming a $MAX\_MULTIPLIER$ of `4`, an account that stakes 100 tokens would have a maximum of:

$$
\text{MP}_\text{Maximum} = 100 + {100 \times 4} = 500
$$

This means that the account can never have more than 500 MP, no matter how long they stake. Also notice that the

Another interesting characteristic is that, if the $MAX\_MULTIPLIER$ is equal to the maximum amount of years an account
can lock up their stake, the account will reach the maximum amount of MP right after staking if they lock up their stake
for the maximum amount of time.

## Summary

In this document, we've covered the concept of Multiplier Points (MP) and how they are used to reward participants in
the staking system. Here's a quick recap:

- **Initial MP** is based on the stake amount and lock-up time. Having a lock-up time increases the amount of MP, which
  can also be considered "bonus" MP.
- **Accrued MP** grows over time and adds to the staking power.
- **Total MP** is the sum of Initial MP and Accrued MP.
- **Maximum MP** is the maximum amount of MP an account can accrue.
