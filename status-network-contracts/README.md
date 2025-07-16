# Staking Protocol [![Github Actions][gha-badge]][gha] [![Codecov][codecov-badge]][codecov] [![Foundry][foundry-badge]][foundry]

[gha]: https://github.com/vacp2p/staking-reward-streamer/actions
[gha-badge]: https://github.com/vacp2p/staking-reward-streamer/actions/workflows/test.yml/badge.svg
[codecov]: https://codecov.io/gh/vacp2p/staking-reward-streamer
[codecov-badge]: https://codecov.io/gh/vacp2p/staking-reward-streamer/graph/badge.svg
[foundry]: https://getfoundry.sh/
[foundry-badge]: https://img.shields.io/badge/Built%20with-Foundry-FFDB1C.svg

## 🧭 Overview

The **Staking Reward Streamer Protocol** enables secure token staking with dynamic reward calculation on Ethereum. Built with modularity and upgradability in mind, the system includes core components to manage stake deposits, reward calculations, time-based locking, and contract migration through user consent.

---

## 🧩 Core Contracts

### 🛠️ `StakeManager`

- Handles staking logic, tracks stakes and reward epochs.
- Calculates APY via **Multiplier Points**, which increase over time.
- Validates vaults using codehash verification for added safety.
- Upgradeable via proxy; users can opt out of migrations.

### 🔐 `StakeVault`

- A vault owned by the user, used to store and manage staked tokens.
- Interacts directly with `StakeManager` for staking and unstaking operations.
- Ensures only the owner can execute critical actions.
- Verifies contract code via codehash to ensure safety.

---

## ✨ Features

- **Secure, user-owned staking vaults**
- **Dynamic APY via Multiplier Points**
- **Stake locking to boost rewards**
- **ERC20-compatible (via OpenZeppelin)**
- **Proxy upgradeability with opt-in/opt-out support**
- **Epoch-based reward streaming**

---

## 🚀 Getting Started

### 📦 Install Dependencies

```bash
pnpm install
```

---

## ⚙️ Usage

### 📄 Deployment Flow

1. **Deploy `StakeManager`**
2. **Deploy a sample `StakeVault` (e.g., on a devnet or testnet)**
3. **Configure codehash** in `StakeManager`:

```solidity
stakeManager.setTrustedCodehash(<vault_codehash>, true);
```

---

### 💰 Staking

1. **Approve** the `StakeVault` to spend your tokens:

```solidity
erc20.approve(stakeVaultAddress, amount);
```

2. **Stake** your tokens:

```solidity
stakeVault.stake(amount, secondsToLock);
```

> ⚠️ Do not transfer tokens directly to the `StakeVault`. Always use `approve` + `stake`.

Minimum stake amount and lock duration are enforced via contract settings. Epochs are automatically processed on stake actions.

---

### 🔓 Unstaking

```solidity
stakeVault.unstake(amount);
```

- Only available for unlocked balances.
- Reduces stake proportionally based on amount and duration.

---

### 🔁 Migration (Opt-In/Out)

Users may opt-in to a new `StakeManager` implementation or leave:

```solidity
stakeVault.acceptMigration(); // opt-in
stakeVault.leave();           // opt-out
```

> Migration triggers automatic reward claiming. Locked balances can still opt out.

---

## 📬 Deployed Contracts

These are the official contract deployments on the **Sepolia testnet** (via [Status Network Explorer](https://sepoliascan.status.network)):

| Contract            | Address                                                                                             |
|---------------------|-----------------------------------------------------------------------------------------------------|
| **StakeManagerProxy** | [0x2C09141e66970A71862beAcCbDb816ec01D6B676](https://sepoliascan.status.network/address/0x2C09141e66970A71862beAcCbDb816ec01D6B676?tab=contract) |
| **StakeManager**      | [0xa2432fB545829f89E172ddE2DeD6D289c7ee125F](https://sepoliascan.status.network/address/0xa2432fB545829f89E172ddE2DeD6D289c7ee125F?tab=contract) |
| **VaultFactory**      | [0xA6300Bd8aF26530D399a1b24B703EEf2c48a71Be](https://sepoliascan.status.network/address/0xA6300Bd8aF26530D399a1b24B703EEf2c48a71Be) |
| **KarmaProxy**        | [0x486Ac0F5Eb7079075dE26739E1192D41F278a8db](https://sepoliascan.status.network/address/0x486Ac0F5Eb7079075dE26739E1192D41F278a8db) |
| **Karma**             | [0xE9413C84eFF6B08E4F614Efe69EB7eb9a1Ca1180](https://sepoliascan.status.network/address/0xE9413C84eFF6B08E4F614Efe69EB7eb9a1Ca1180?tab=contract) |
| **KarmaNFT**          | [0xdE5592e1001f52380f9EDE01aa6725F469A8e46F](https://sepoliascan.status.network/address/0xdE5592e1001f52380f9EDE01aa6725F469A8e46F?tab=contract) |

---

## 🧪 Development

### 🏗️ Build Contracts

```sh
forge build
```

### 🧹 Clean Build Artifacts

```sh
forge clean
```

### 🧪 Run Tests

```sh
forge test
```

### 🧮 Coverage

```sh
forge coverage
```

### 🚀 Deploy Locally (Anvil)

```sh
forge script script/Deploy.s.sol --broadcast --fork-url http://localhost:8545
```

> Requires `MNEMONIC` env variable.

---

## 📊 Gas & Linting

### Gas Reports

```sh
pnpm gas-report
forge snapshot
```

### Linting

```sh
pnpm lint
```

### Formatting

```sh
forge fmt
```

### Commit preparing command

```sh
pnpm adorno
```

