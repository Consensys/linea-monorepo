# Understanding Multiplier Points and XP Rewards

## Overview

The staking system uses Multiplier Points (MP) to enhance staking power and influence the rewards participants earn.
This document explains:

1. How MP determines XP rewards.
2. How stake amount and lock-up duration affect MP.
3. The role of Initial MP and Accrued MP.
4. The relationship between XP tokens and the StakeManager.
5. Examples illustrating how MP is calculated and accumulated over time.

## Key Concepts

1. **Initial MP**: Multiplier Points issued immediately based on the stake amount and lock-up duration.
2. **Accrued MP**: MP that accumulate over time as a function of the stake amount, elapsed time, and annual percentage
   yield (APY).
3. **XP Tokens**: The token rewarded by the system.
4. **XP Rewards**: Determined by the total MP a user holds relative to the total MP in the system.

## Formula for Multiplier Points

### Initial MP

The formula for Initial MP is derived as follows:

$$
\text{MP}_ \text{Initial} = \text{Stake} \times \left( 1 + \frac{\text{APY} \times T_ \text{lock}}{100 \times T_ \text{year}} \right)
$$

Where:

- $Stake$: The amount of tokens staked.
- $APY$: Annual Percentage Yield, set at 100%.
- $T_{lock}$: Lock-up duration in seconds.
- $T_{year}$: Total seconds in a year.

This formula calculates the MP issued immediately when tokens are staked with a lock-up period.

### Accrued MP

Accrued MP is calculated for time elapsed as:

$$
\text{MP}_ \text{Accrued} = \text{Stake} \times \frac{\text{APY} \times T_ \text{elapsed}}{100 \times T_ \text{year}}
$$

Where:

- $T_{elapsed}$: Time elapsed since staking began, measured in seconds.

This formula adds MP as a function of time, rewarding users who keep their stake locked.

### Total MP

Total MP combines both Initial MP and Accrued MP:

$$
\text{MP}_ \text{Total} = \text{MP}_ \text{Initial} + \text{MP}_ \text{Accrued}
$$

This total is used to calculate the user’s share of rewards.

## How MP Affects XP Rewards

The rewards distributed in the system are proportional to each user’s MP. The formula for reward share is:

$$
\text{Reward}_ \text{user} = \text{Rewards}_ \text{Total} \times \frac{\text{MP}_ \text{user}}{\text{MP}_ \text{total}}
$$

This ensures rewards are allocated based on the user’s contribution to the total MP.

## Examples

Let’s consider three participants: Alice, Bob, and Charlie. The total reward pool is set at 10,000 XP tokens.

### Example 1: Alice

- **Stake**: 100 tokens
- **Lock-Up Time**: 30 days
- **Elapsed Time**: 15 days

#### Initial MP

Using the formula:

$$
\text{MP}_ \text{Initial} = 100 \times \left( 1 + \frac{100 \times 30}{100 \times 365} \right)
$$

$$
\text{MP}_ \text{Initial} = 100 \times \left( 1 + 0.082 \right) = 108.2
$$

#### Accrued MP

$$
\text{MP}_ \text{Accrued} = 100 \times \frac{100 \times 15}{100 \times 365} = 4.1
$$

#### Total MP

$$
\text{MP}_ \text{Total} = 108.2 + 4.1 = 112.3
$$

#### Reward Share

$$
\text{Reward}_ \text{Alice} = 10,000 \times \frac{112.3}{1,146.7} \approx 978.9
$$

### Example 2: Bob

- **Stake**: 500 tokens
- **Lock-Up Time**: 90 days
- **Elapsed Time**: 45 days

#### Initial MP

$$
\text{MP}_ \text{Initial} = 500 \times \left( 1 + \frac{100 \times 90}{100 \times 365} \right)
$$

$$
\text{MP}_ \text{Initial} = 500 \times \left( 1 + 0.247 \right) = 623.5
$$

#### Accrued MP

$$
\text{MP}_ \text{Accrued} = 500 \times \frac{100 \times 45}{100 \times 365} = 61.6
$$

#### Total MP

$$
\text{MP}_ \text{Total} = 623.5 + 61.6 = 685.1
$$

#### Reward Share

$$
\text{Reward}_ \text{Bob} = 10,000 \times \frac{685.1}{1,146.7} \approx 5,975.2
$$

### Example 3: Charlie

- **Stake**: 300 tokens
- **Lock-Up Time**: 0 days
- **Elapsed Time**: 60 days

#### Initial MP

$$
\text{MP}_ \text{Initial} = 300 \times \left( 1 + \frac{100 \times 0}{100 \times 365} \right) = 300
$$

#### Accrued MP

$$
\text{MP}_ \text{Accrued} = 300 \times \frac{100 \times 60}{100 \times 365} = 49.3
$$

#### Total MP

$$
\text{MP}_ \text{Total} = 300 + 49.3 = 349.3
$$

#### Reward Share

$$
\text{Reward}_ \text{Charlie} = 10,000 \times \frac{349.3}{1,146.7} \approx 3,045.9
$$

### Total MP Calculation

The total MP for all participants is:

$$
\text{MP}_ \text{Total All} = 112.3 + 685.1 + 349.3 = 1,146.7
$$

## Summary

- **Initial MP** is based on the stake amount and lock-up time.
- **Accrued MP** grows over time and adds to the staking power.
- XP tokens rewards are proportional to their MP.
- Total MP determines the share of XP rewards a participant earns.
