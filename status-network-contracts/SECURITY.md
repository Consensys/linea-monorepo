# Security Reference

This document provides critical security information for auditors and security researchers about the Status Network
contracts.

## System Architecture Overview

The Status Network contracts is a multi-contract system centered around the Karma token and its reward distribution
mechanisms. The staking system is one of several reward distributors in the broader ecosystem. Key components include:

- **Karma**: Reputation and governance token with voting capabilities. Generally non-transferable, but administrators
  can whitelist specific accounts for transfers
- **StakeManager**: Core staking logic with upgradeable proxy pattern (functions as a reward distributor for Karma)
- **StakeVault**: User-owned vaults for secure stake management
- **VaultFactory**: Factory for creating user stake vaults
- **KarmaNFT**: Non-transferable NFT representing user karma levels
- **KarmaRLN**: Spam prevention mechanism for onchain actions

For a detailed system overview, please refer to the [System Overview](docs/system-overview.md) document.

## Formal Verification and Invariants

The Status Network contracts have been formally verified using Certora. All verified properties and invariants are
documented in [PROPERTIES.md](PROPERTIES.md). Auditors should review these invariants to understand the formally
verified security properties of the system.

## System Actors

The Status Network operates with several distinct actor types, each with specific roles and permissions:

- **Administrators**: Privileged accounts that can manage contract upgrades, whitelist transfers, and configure system
  parameters
- **Governance Participants**: Karma token holders who can participate in voting and governance decisions
- **Reward Distributors**: Authorized contracts (like `StakeManager`) that distributes Karma rewards for various
  activities
- **Vault Users**: Own and operate `StakeVault`s, can deposit/withdraw stakes, claim rewards, and participate in staking
  activities

## Target chain for deployment

The Status Network contracts are going to be deployed on the **Status L2**, which is a fork of Linea.

## Integrated tokens

The Status Network contracts integrate with the following tokens:

- **SNT**: The native staking token used for staking in the `StakeManager` and `StakeVault`s. This token is a standard
  [`BridgedToken`](https://github.com/status-im/status-network-monorepo/blob/develop/contracts/src/bridging/token/BridgedToken.sol)
- **Karma**: The reputation and governance token distributed as rewards for staking and other activities

## Trust Assumptions

The system relies on several trust assumptions:

- **Administrator Trust**: Administrators are trusted to act in the best interest of the system, manage upgrades
  responsibly, and not abuse their privileges.
- **Governance Trust**: Karma token holders are trusted to participate in governance decisions responsibly.
- **Reward Distributor Trust**: We expect reward distributors to be reviewed and audited and are therefore trusted to
  distribute rewards fairly and behave according to the system's rules.
- **Sequencer Trust**: The sequencer is trusted to include transactions in a timely manner and not censor specific users
  or transactions.

## Known issues and limitations

### Vault Creation DDOS

Users can create an unlimited number of `StakeVault` instances through the `VaultFactory`, which could potentially
overwhelm operations that iterate over or interact with multiple vaults, causing denial of service for certain system
functions. In practice, we expect users to create only a few vaults, but there are no onchain limits to enforce this. If
a user _does_ create so many vaults that functions like `distributeRewards` run out of gas, it will only affect that
user, not the entire system. In other words, users need to willingly create many vaults to affect their own ability to
use the system, but they cannot affect other users.

### Excessive Reward Distributors

Administrators can add an unlimited number of reward distributors to the Karma token contract. If too many reward
distributors are added, operations that iterate over all distributors could run out of gas, potentially causing
system-wide denial of service. Unlike the vault creation issue, this affects the entire system since it's controlled by
trusted administrators rather than individual users.

### Slashing DDOS via Reward Distributor Revert

When `Karma.slash()` is called, it attempts to call `rewardDistributor.rewardsBalanceOfAccount()` and
`rewardDistributor.redeemRewards()` on all registered reward distributors. If any of these reward distributor functions
revert, the entire slashing operation fails, effectively creating a denial of service for the slashing mechanism. A
malicious or buggy reward distributor could intentionally or unintentionally prevent all slashing operations from
succeeding.

Since we're aiming to review and audit every reward distributor that's added to the system, we expect them to behave
correctly, which means not reverting in these functions.

### Total Supply Calculation Issues

The `Karma.totalSupply()` function iterates over all registered reward distributors to calculate the total token supply.
This creates similar vulnerabilities: if too many reward distributors are added, the function could run out of gas, and
if any reward distributor's balance calculation reverts, the entire `totalSupply()` call fails. This could break
integrations and external contracts that rely on this standard ERC20 function.

### Cumulative Lock-up Time Beyond Maximum

While the system enforces a maximum lock-up time of 4 years per individual lock period, vaults do not track cumulative
lock-up duration across multiple lock cycles. A vault can theoretically lock for extended periods beyond 4 years total
by repeatedly locking for the maximum duration. The only practical limitation is the vault's maximum multiplier points
(maxMP), which caps the rewards a vault can accumulate regardless of total lock duration.
