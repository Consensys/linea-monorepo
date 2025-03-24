# Staking Protocol [![Github Actions][gha-badge]][gha] [![Codecov][codecov-badge]][codecov] [![Foundry][foundry-badge]][foundry]

[gha]: https://github.com/vacp2p/staking-reward-streamer/actions
[gha-badge]: https://github.com/vacp2p/staking-reward-streamer/actions/workflows/test.yml/badge.svg
[codecov]: https://codecov.io/gh/vacp2p/staking-reward-streamer
[codecov-badge]: https://codecov.io/gh/vacp2p/staking-reward-streamer/graph/badge.svg
[foundry]: https://getfoundry.sh/
[foundry-badge]: https://img.shields.io/badge/Built%20with-Foundry-FFDB1C.svg

---

This project is a smart contract system built for secure token staking and reward management on the Ethereum blockchain. It provides essential components for handling staking operations, managing rewards, and ensuring the security of staked assets. The system is designed to be robust, leveraging mathematical precision for reward calculations and secure vaults for asset protection.

### Key Components

- **StakeManager**: This contract handles the core staking operations. It is responsible for managing the staking process, calculating rewards, and tracking expired staking periods using advanced mathematical calculations. The contract utilizes OpenZeppelin’s ERC20 and math utilities to ensure accuracy and security.
- **StakeVault**: This contract is responsible for securing user stakes. It ensures the security of user deposits by allowing only the owner to control the vault. The vault communicates with the StakeManager to handle staking and unstaking operations, while ensuring both the safety of funds and the validity of the vault’s code.

## Features

- **Secure Staking**: Users can securely stake tokens, with all interactions managed by the StakeVault and governed by the StakeManager. The StakeVault is owned by the user, and only the owner is permitted to perform actions on it. It stores the staked tokens and StakeManager verifies the StakeVault codehash to ensure that the code managing the tokens is valid and secure.
- **Multiplier Points**: The Annual Percentage Yield (APY) is achieved through an internal balance of Multiplier Points, which grow over time based on the amount of staked tokens. Rewards are proportional to the user's balance of Multiplier Points and stake.
- **Stake Locking**: Users can lock their stake to start with an increased balance of Multiplier Points, thereby enhancing their potential rewards.
- **Integration with ERC20 Tokens**: Built on top of the OpenZeppelin ERC20 token standard, the system ensures compatibility with any ERC20-compliant token.
- **Opt-in Upgrade**: StakeManager is a Proxy upgradable contract, however, users can choose to disagree with an upgrade and opt-out without penalities, and simply withdraw their stake, bypassing any lock defined. 

## Installation

To install the dependencies for this project, run the following command:

```bash
pnpm install
```

## Usage

### Contract Setup

To install the system on the blockchain, follow these steps:

1. **Deploy StakeManager**: Begin by deploying the `StakeManager` contract.
2. **Deploy Sample VaultManager**: Next, deploy a sample `VaultManager` contract on any chain (this can be on a development network or a testnet).
3. **Configure Codehash**: Once the `VaultManager` is deployed, retrieve its codehash and configure it in the `StakeManager` using the `setTrustedCodehash(bytes32, bool)` function.

### Staking Tokens

To stake tokens, the `StakeVault` contract interacts with the `StakeManager`. You can call the function `StakeVault.stake(uint256 _amount, uint256 _secondsToLock)` with the desired amount and lock duration. Before staking, it is required to call `approve` on the `StakeToken` to authorize the `StakeVault` address to manage your tokens.

The minimum `_secondsToLock` is defined by the contract settings, and the minimum `_amount` required to stake is set to ensure it generates Multiplier Points in every epoch.

Tokens are never deposited directly or transferred to the `StakeManager`. Additionally, tokens should not be sent directly to the `StakeVault` address. Always use the `approve` method followed by the `stake` function to properly stake tokens.

Note that staking will automatically process epochs. Refer to the **Manually Processing Epochs** section for more information.tion.

### Unstaking

To unstake tokens, users call `StakeVault.unstake(uint256 amount)`. Unstaking reduces the user's balance on the `StakeManager` in proportion to the staked amount and time spent staking. Users can only unstake if their balance is not locked.

### Opt-In or Opt-Out Migration

Users can accept or reject the new contract by calling `StakeVault.acceptMigration()` or `StakeVault.leave()`. If users have pending rewards in the old contract, those rewards will be claimed before opting in or out. Users with locked balances will have the option to leave, even if their balance is still locked.

Note that opting in or out of migration will automatically claim rewards. Refer to the **Claiming Rewards** section for more information.

## Development

This is a list of the most frequently needed commands.

### Build

Build the contracts:

```sh
$ forge build
```

### Clean

Delete the build artifacts and cache directories:

```sh
$ forge clean
```

### Compile

Compile the contracts:

```sh
$ forge build
```

### Coverage

Get a test coverage report:

```sh
$ forge coverage
```

### Deploy

Deploy to Anvil:

```sh
$ forge script script/Deploy.s.sol --broadcast --fork-url http://localhost:8545
```

For this script to work, you need to have a `MNEMONIC` environment variable set to a valid
[BIP39 mnemonic](https://iancoleman.io/bip39/).

For instructions on how to deploy to a testnet or mainnet, check out the
[Solidity Scripting](https://book.getfoundry.sh/tutorials/solidity-scripting.html) tutorial.

### Format

Format the contracts:

```sh
$ forge fmt
```

### Gas Usage

Get a gas report:

```sh
pnpm gas-report
```

Get a gas snapshot:

```bash
forge snapshot
```

### Lint

Lint the contracts:

```sh
$ pnpm lint
```

### Test

Run the tests:

```sh
$ forge test
```

#### Prepare to commit

Formats, generates snapshot and gas report:

```sh
pnpm adorno
```

## License

This project is licensed under MIT.
