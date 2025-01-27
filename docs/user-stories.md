# User Flows

This document outlines various scenarios and flows that users might encounter when interacting with the staking system.
We'll explore different staking strategies, lock-up periods, and their impact on rewards and Multiplier Points (MP).

## Basic Staking Flow (No Lock-up)

The simplest way to participate in the system is to stake tokens without a lock-up period.

### What happens when you stake without lock-up:

1. Initial stake:

   - You transfer tokens to your stake vault
   - You receive 1:1 initial MP (e.g., staking 100 tokens gives you 100 MP)
   - Your maximum MP is set to 5x your stake (e.g., for 100 tokens, max MP = 500)

2. Over time:

   - You earn additional MP linearly at a rate of 100% APY
   - For example, after 6 months, you'll have earned ~50% more MP
   - MP continue accruing until reaching your maximum MP

3. Rewards:

   - You earn rewards based on your total weight (stake + MP)
   - Rewards accrue in real-time
   - No action needed to "claim" rewards; they're tracked automatically

4. Unstaking:
   - You can unstake at any time
   - MP are reduced proportionally to the amount unstaked
   - Accrued rewards are preserved

### Example Scenario:

Alice stakes 1000 tokens with no lock-up:

```
Initial state:
- Stake: 1000 tokens
- Initial MP: 1000
- Max MP: 5000
- Total Weight: 2000 (1000 stake + 1000 MP)

After 6 months:
- Stake: 1000 tokens
- MP: ~1500 (initial 1000 + ~500 accrued)
- Total Weight: ~2500
```

## Staking with Lock-up

Locking up tokens provides additional MP bonuses, increasing your earning potential.

### What happens when you stake with lock-up:

1. Initial stake with lock:

   - You transfer tokens to your stake vault
   - You receive 1:1 initial MP
   - You receive bonus MP based on lock duration
   - Your maximum MP includes the lock-up bonus

2. During lock period:

   - Tokens cannot be unstaked
   - MP continue accruing as normal
   - Rewards continue based on total weight

3. After lock period:
   - Tokens become available for unstaking
   - Bonus MP are retained
   - Normal MP accrual continues

### Example Scenario:

Bob stakes 1000 tokens with a 1-year lock:

```
Initial state:
- Stake: 1000 tokens
- Initial MP: 1000
- Lock bonus MP: 1000 (100% for 1-year lock)
- Max MP: 6000 (5000 base + 1000 lock bonus)
- Total Weight: 3000 (1000 stake + 2000 MP)

After 6 months:
- Stake: 1000 tokens (still locked)
- MP: ~2500 (1000 initial + 1000 lock bonus + ~500 accrued)
- Total Weight: ~3500
```

## Multi-vault Strategy

Users can create multiple vaults with different configurations.

### What happens with multiple vaults:

1. Creating vaults:

   - Each vault is independent
   - Different lock-up periods possible
   - Separate MP tracking per vault

2. Rewards calculation:
   - System aggregates weights across all vaults
   - Rewards distributed based on total weight
   - Each vault's rewards tracked separately

### Example Scenario:

Charlie creates two vaults:

```
Vault 1 (no lock):
- Stake: 500 tokens
- Initial MP: 500
- Max MP: 2500

Vault 2 (1-year lock):
- Stake: 500 tokens
- Initial MP: 500
- Lock bonus MP: 500
- Max MP: 3000

Total position:
- Total Stake: 1000 tokens
- Total Initial MP: 1500
- Total Max MP: 5500
```

## Reward Distribution Scenarios

Understanding how rewards are distributed in different situations.

### Scenario 1: Single Staker

When you're the only staker in the system:

- You receive 100% of rewards
- Your reward rate is constant if your weight is constant
- Reward rate increases as your MP accumulate

### Scenario 2: Multiple Stakers

With multiple stakers:

- Rewards are proportional to relative weights
- Your share changes as others enter/exit
- Your share increases as your MP accumulate

### Example:

System with 1000 tokens/day rewards:

```
Your position:
- Stake: 1000 tokens
- MP: 1500
- Weight: 2500 (40% of system)

Other stakers:
- Combined Weight: 3750 (60% of system)

Your rewards:
- Daily reward: 400 tokens (40% of 1000)
```

## Emergency Scenarios

The system includes safety mechanisms for unexpected situations.

### Emergency Exit:

If emergency mode is enabled:

1. You can withdraw immediately
2. Lock-up periods are ignored
3. Funds are returned to your specified address

### System Upgrade:

During contract upgrades:

1. You can choose to exit if you don't trust new implementation
2. Your vault remains yours
3. You can rejoin later if desired

## Best Practices

1. Lock-up Strategy:

   - Consider lock-up for long-term positions
   - Balance higher rewards vs flexibility
   - Use multiple vaults for different strategies

2. Monitoring:

   - Track your MP growth
   - Monitor your reward share
   - Watch total system weight

3. Risk Management:
   - Understand lock-up implications
   - Keep some portion liquid if needed
   - Monitor system upgrades

## Common Scenarios and Outcomes

Here's a quick reference for common actions and their outcomes:

| Action           | MP Impact              | Reward Impact     | Lock Status      |
| ---------------- | ---------------------- | ----------------- | ---------------- |
| Stake (no lock)  | 1:1 initial            | Immediate start   | Unstake anytime  |
| Stake (1yr lock) | 2x initial             | Higher share      | Locked for 1yr   |
| Partial unstake  | Proportional reduction | Reduced share     | If not locked    |
| Add to existing  | Proportional increase  | Increased share   | Follows original |
| Create new vault | Independent MP         | Aggregate rewards | Independent lock |

Remember that all these scenarios work together with the MP system described in
[Multiplier Points](multiplier-points.md) and the reward mechanics detailed in [Rewards](rewards.md).
