# System Overview

This document provides an overview of the system architecture and design of the staking system. We'll learn about the
system components, their interactions, and the data flow between them.

## Components

![System Overview](assets/overview.png)

1. **Account**: EOAs and smart accounts that participate in the staking system.
2. **Stake vault**: A smart contract that maintains the account's stake. Accounts interface with the staking system
   through their stake vaults.
3. **Stake manager proxy**: A smart contract proxy that delegates to the logic/implementation contract of the staking
   system. It maintains its own onchain storage. Stake vaults interact with the staking system through the proxy.
4. **Stake manager implementation**: A smart contract in which the logic of the staking system resides. The stake
   manager proxy uses this contract to execute the staking system's logic. This contract is ugpradeable.
5. **Karma distributor**: Systems that distribute Karma tokens on behalf of the Karma token contract. The staking system
   is one such a Karma distributor.
6. **Karma**: An ERC20 token contract that issues Karma tokens to accounts based on their stake. In addition to its own
   balance accounting, it relies on Karma distributors to aggregate total balances of individual accounts. Karma tokens
   are not transferable.
7. **Karma NFT**: An ERC721 token contract that maintains non-transferable NFTs of accounts that participate in the
   Karma program, either via the staking system or other activities. The Karma NFTs rely on the Karma token contract to
   render an account's Karma balance.
8. **Accounts, DApps, Wallets**: External entities that consume the Karma tokens and NFTs.
9. **DAO/Admin**: Entity that controls the contract and sets XP rewards in the system which will be distributed
   according to every participant's stake.

## How the system works

- Accounts (EOAs or smart accounts) create one or multiple stake vaults to stake their SNT and participate in the Karma
  Programme. In most cases, accounts will have only one stake vault, but nothing prevents them from creating more vaults
  with different configurations.
- The stake vaults interact with the stake manager through the proxy by forwarding calls to the implementation contract.
  When an account stakes funds, their funds are moved into the stake vault and they will stay there until the account
  decides to unstake them.
- While the account is staking, it will accrue Karma points based on the amount of SNT staked and the duration of the
  stake. The longer accounts stake, the more Karma they will earn.
- By locking up their stake, accounts receive multiplier points that increase their initial Karma earnings upon staking.
- At any point in time, accounts can view their Karma token balance in their wallets and how it updates in realtime.
- In addition, every account receives an NFT that represents their participation in the Karma Programme. It will evolve
  as the account gains Karma.
- Under the hood, multiplier points accrue over time. Accounts can compound their multiplier points to increase their
  Karma rewards in the system.
- Eventually, an account will reach the maximum amount of multiplier points they can accrue based on their stake amount,
  at which point their total weight in Karma shares will no longer increase. They still earn Karma rewards set by the
  admin, according to their share.
- Accounts can unstake their SNT at any time (unless locked up). When they do, they will receive their initial stake
  back, along with any Karma rewards they have earned, however, they will loose their multiplier points.
- The Karma in the account's wallet can be used to access exclusive features or services within the Status app or other
  DApps that support the Karma token, like voting in governance proposals or participating in exclusive events.
