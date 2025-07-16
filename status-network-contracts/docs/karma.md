# Karma

## Overview

The Karma contract is an ERC-20 token implementation with a modified supply mechanism that incorporates external reward
distributors. Karma tokens are not transferrable, but they can be used as voting power in the Status Network.

## Features

- **Minting with Restrictions:**
  - The contract owner (admin) can mint tokens, and their accounting is kept internally.
  - Prevents exceeding a dynamically calculated mint allowance.
- **Reward Distributors Integration:**
  - Tracks balances and supplies from external reward distributors.
  - Allows addition and removal of reward distributor by the owner.
- **Non-Transferrable Tokens:**
  - Transfers, approvals, and allowances are disabled.
  - Users can only receive balances from minting or reward distributions.
- **Supply Calculation:**
  - The total supply is the sum of the internal supply and the external supplies.

## Contract Details

- `NAME`: "Karma"
- `SYMBOL`: "KARMA"

### State Variables

- `rewardDistributors`: A list of addresses implementing the `IRewardDistributor` interface.

### Errors

- `Karma__MintAllowanceExceeded`: Raised when minting exceeds the allowed threshold.
- `Karma__TransfersNotAllowed`: Raised when a transfer, approval, or transferFrom is attempted.

## Supply and Balance Calculation

- **totalSupply()**: Sum of the internal supply and the sum of external supplies from reward distributors.
- **balanceOf(address)**: Internal balance plus the sum of external balances from reward distributors.

## Sources of XP Tokens

One of the sources for the generation of XP tokens is the [staking protocol](overview.md), with more sources planned in
the future.

## Notes

- The contract is designed to work alongside an external reward system.
- Transfers and approvals are explicitly disabled to enforce controlled distribution.
- The contract ensures a dynamic supply mechanism tied to external rewards.
