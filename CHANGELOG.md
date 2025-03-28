# Changelog

All notable changes to this project will be documented in this file. See [commit-and-tag-version](https://github.com/absolute-version/commit-and-tag-version) for commit guidelines.

## [0.2.0](https://github.com/vacp2p/staking-reward-streamer/compare/v0.1.1...v0.2.0) (2025-03-28)


### ⚠ BREAKING CHANGES

* **

- `RewardsStreamerMP` is now `StakeManager`
- `StakingManager__*` error selectors are now `StakeManager__*`
  selectors
* - `VaultData.mpStaked` no longer exists, use `VaultData.mpAccrued`
instead.

- `Compound(address,uint256)` is now `VaultUpdated(address,uint256,uint256)`

- `AccountLeft(address)` is now `VaultLeft(address)`
* `mintAllowance()` no longer exists.
* `getStakedBalance(address)` is now `stakedBalanceOf(address)`
* The previous public `rewardIndex` field is now called `lastRewardIndex`.
* **RewardsStreamerMP:** A couple of APIs have been removed or replaced.

-> inline _compound into _updateVault()
-> remove compond() in favor of updateVaultMP()
-> rename updateVaultMP() to updateVault()
-> rename compoundAccount() to updateAccount()

* !refactor(RewardsStreamerMP): rename `RewardsStreamerMP` to `StakeManager` ([801740f](https://github.com/vacp2p/staking-reward-streamer/commit/801740f74fea983782302d65e5c9c34770b93ae3))
* !refactor(RewardsStreamerMP): remove `VaultData.mpStaked` ([7bd0c16](https://github.com/vacp2p/staking-reward-streamer/commit/7bd0c16872edb44a51e3432528cea5be332ff2bc))
* !refactor(Karma): remove `mintAllowance()` ([ce982b9](https://github.com/vacp2p/staking-reward-streamer/commit/ce982b9ce5033e22bc2e087f8146b7514287cb5c)), closes [#192](https://github.com/vacp2p/staking-reward-streamer/issues/192)
* !refactor: rename `getStakedBalance()` -> `stakedBalanceOf()` ([695a208](https://github.com/vacp2p/staking-reward-streamer/commit/695a2088041a53457e7809d00eddb65b858d072c)), closes [#188](https://github.com/vacp2p/staking-reward-streamer/issues/188) [#188](https://github.com/vacp2p/staking-reward-streamer/issues/188)
* !refactor(RewardsStreamerMP): rename internal functions and `rewardIndex` ([8e4aa68](https://github.com/vacp2p/staking-reward-streamer/commit/8e4aa682c118abb0c715a7f9540d78ea66503970)), closes [#189](https://github.com/vacp2p/staking-reward-streamer/issues/189)
* **RewardsStreamerMP:** merge `compound()` with `_updateVault()` ([2e01e0d](https://github.com/vacp2p/staking-reward-streamer/commit/2e01e0d03bc84660ea0ad643649ad6fa6496c27d)), closes [#187](https://github.com/vacp2p/staking-reward-streamer/issues/187) [#187](https://github.com/vacp2p/staking-reward-streamer/issues/187)


### Features

* **RewardsStreamerMP:** allow for staking multiple times with lock ([4fa3eb0](https://github.com/vacp2p/staking-reward-streamer/commit/4fa3eb06e08f7a8f64bfc18623ddcb430e024b67)), closes [#152](https://github.com/vacp2p/staking-reward-streamer/issues/152)


### Bug Fixes

* **RewardsStreamerMP:** remove double totalMPStaked substraction ([dabcf5c](https://github.com/vacp2p/staking-reward-streamer/commit/dabcf5c9908e70c3d57f645f2618ab39a02888ef))
* **StakeManager:** Allow extending the lock after increasing stake to allow account reaching absolute max MP ([8df475a](https://github.com/vacp2p/staking-reward-streamer/commit/8df475aab8c8d1935d48e9589c528d2ff15dc31a))

## 0.1.1 (2025-03-18)


### Features

* add `TrustedCodehashAccess` contract and interface ([6fea58b](https://github.com/vacp2p/staking-reward-streamer/commit/6fea58b334701a0f1d6c41e18434b70a05f47985)), closes [#39](https://github.com/vacp2p/staking-reward-streamer/issues/39) [#15](https://github.com/vacp2p/staking-reward-streamer/issues/15)
* introduce `StakeVault` ([0e14b7b](https://github.com/vacp2p/staking-reward-streamer/commit/0e14b7b3a6979bdcf9721bd6a4ccc9848590f24b)), closes [#14](https://github.com/vacp2p/staking-reward-streamer/issues/14) [#14](https://github.com/vacp2p/staking-reward-streamer/issues/14)
* introduce deployment script for `RewardsStreamerMP` ([a565dbb](https://github.com/vacp2p/staking-reward-streamer/commit/a565dbbac4bef5ba2a75103913039f7037b50110)), closes [#88](https://github.com/vacp2p/staking-reward-streamer/issues/88)
* introduce proxy clones ([70a7f30](https://github.com/vacp2p/staking-reward-streamer/commit/70a7f30d2a5aebc99dc231eecb438c2945827906)), closes [#101](https://github.com/vacp2p/staking-reward-streamer/issues/101)
* **Karma:** allocate external rewards using the Karma contract ([ed3577f](https://github.com/vacp2p/staking-reward-streamer/commit/ed3577f8c4bc52bf85c508d692e2088f50e2a9e7))
* **Karma:** make karma upgradeable ([aa3442b](https://github.com/vacp2p/staking-reward-streamer/commit/aa3442b577e2a1a3287b47080d4daa4d532da152))
* **RewardsStreamerMP.t.sol:** Make tests use calc functions and test those functions ([6afc760](https://github.com/vacp2p/staking-reward-streamer/commit/6afc760974bb58f9680da048d694e5f341b8811c))
* **RewardsStreamerMP:** add `lock(uint256)` function ([5bc7ebf](https://github.com/vacp2p/staking-reward-streamer/commit/5bc7ebf963c3000f2dee723e074921f68209efc5)), closes [#40](https://github.com/vacp2p/staking-reward-streamer/issues/40)
* **RewardsStreamerMP:** add function to compound all MPs for an account ([5e2dcba](https://github.com/vacp2p/staking-reward-streamer/commit/5e2dcbabd1bdd20a9f6a9961a64aae65966b20c5)), closes [#175](https://github.com/vacp2p/staking-reward-streamer/issues/175)
* **RewardsStreamerMP:** allow vaults to migrate to other vaults ([6f19931](https://github.com/vacp2p/staking-reward-streamer/commit/6f199313ecaacdd9ca634b5078a93a66045ecc22)), closes [#127](https://github.com/vacp2p/staking-reward-streamer/issues/127)
* **RewardsStreamerMP:** enable extending lock period ([e4d21b6](https://github.com/vacp2p/staking-reward-streamer/commit/e4d21b6caf386e8b7bfacc64b030575d5ae48906))
* **RewardsStreamerMP:** introduce `leave()` function ([fcfe72d](https://github.com/vacp2p/staking-reward-streamer/commit/fcfe72d050eae3ff5dab68fbe0e8ffac0b4c18e4)), closes [#66](https://github.com/vacp2p/staking-reward-streamer/issues/66)
* **RewardsStreamerMP:** make `RewardsStreamerMP` upgradeable ([8561a68](https://github.com/vacp2p/staking-reward-streamer/commit/8561a68ffd468d5d6251c9a4d0df65d78e7b302a)), closes [#22](https://github.com/vacp2p/staking-reward-streamer/issues/22)
* **RewardsStreamerMP:** rewards are streamed and dynamically "minted" ([56e9244](https://github.com/vacp2p/staking-reward-streamer/commit/56e92444e04043787180080eeb71e26272b4a4a1))
* **RewardsStreamerMP:** rewardsBalanceOf uses account's accrued + pending MPs ([a413f4c](https://github.com/vacp2p/staking-reward-streamer/commit/a413f4cbbb8938101d02cca5476e083781b78022))
* **RewardsStreamerMP:** stream rewards for a period without checking a real reward token balance ([dffaea2](https://github.com/vacp2p/staking-reward-streamer/commit/dffaea2a7395ae4a7e04babf979fe46a0d27ff48))
* **RewardsStreamerMP:** vaults shares are stakedBalance + mpStaked ([6b31d39](https://github.com/vacp2p/staking-reward-streamer/commit/6b31d3944f60aea66f108b5323ac4c6d7e9807e1))
* **RewardStreamerMP:** add emergency mode so users can exit the system ([1e703e3](https://github.com/vacp2p/staking-reward-streamer/commit/1e703e3f7106c9667847619f6028aff4484b4c4a)), closes [#66](https://github.com/vacp2p/staking-reward-streamer/issues/66)
* **scripts:** add upgrade script ([b5ce251](https://github.com/vacp2p/staking-reward-streamer/commit/b5ce251b8ed4632105a1fc5372c910b146b8afd4))
* **StakeManager:** add capabilities to register vaults ([9374025](https://github.com/vacp2p/staking-reward-streamer/commit/93740259240d3a09b401a12bb784c2408f318dd3)), closes [#70](https://github.com/vacp2p/staking-reward-streamer/issues/70)
* **XPNFTToken:** add base XPNFTToken ([7ed87fa](https://github.com/vacp2p/staking-reward-streamer/commit/7ed87fada9f88278ed29eddfe879f326fe8f9605))
* **XPNFTToken:** add NFTMetadataGeneratorURL ([43536a4](https://github.com/vacp2p/staking-reward-streamer/commit/43536a4dca9f8f97ffacd4804fd748c38032b380))
* **XPNFTToken:** add XPNFTMetadataGenerator ([7352f88](https://github.com/vacp2p/staking-reward-streamer/commit/7352f8837a182ec78317cd0a8d869326d260aed0))
* **XPToken:** add base XPToken with IXPProvider interface ([f816755](https://github.com/vacp2p/staking-reward-streamer/commit/f816755340a086c3c9e8cfa9ae2593da6d6200de))
* **XPToken:** external supply and balances are 1:1 to providers numbers ([c6623c3](https://github.com/vacp2p/staking-reward-streamer/commit/c6623c3d34524c1ab09eaeeefec4a8ec9425e270))
* **XPToken:** introduce mintAllowance. cannot mint if totalSupply is >= external supply * 3 ([c54ad8f](https://github.com/vacp2p/staking-reward-streamer/commit/c54ad8f361a40074ab84218b7be89ae699b92f98))
* **XPToken:** XPToken extends ERC20 and use its balanceOf/totalSupply adding external supply/balances ([7e7c513](https://github.com/vacp2p/staking-reward-streamer/commit/7e7c51302cb5eeb89402690e3c35d5818cc317ae))


### Bug Fixes

* **certora:** add prover args to config to prevent timeouts ([b4b9187](https://github.com/vacp2p/staking-reward-streamer/commit/b4b91873ed217a551c23b330e9eb76ea20d4f9c9))
* **certora:** fix timeout on certora with specific config ([6c89793](https://github.com/vacp2p/staking-reward-streamer/commit/6c897938540b36c187a8a473c0801eccb57c474c))
* **ci:** run MPLessEqualMaxMP spec on CI ([#171](https://github.com/vacp2p/staking-reward-streamer/issues/171)) ([fed4446](https://github.com/vacp2p/staking-reward-streamer/commit/fed444674919894f49f129b871c7a415ea763d2a))
* **EmergencyMode.spec:** add YEAR() to isViewFunction ([106ec98](https://github.com/vacp2p/staking-reward-streamer/commit/106ec9883956c80ce6ce3606ebb3fdcef376f84c))
* improve precision loss when unstaking and add testso ([0af58f9](https://github.com/vacp2p/staking-reward-streamer/commit/0af58f90fbbf609a9d268a22900bb3eceec26f98))
* **RewardsStreamerMP:** change account mp update time only if necessary ([aa15954](https://github.com/vacp2p/staking-reward-streamer/commit/aa15954d0c7a3bb62aee348fc5078ee708eb7f05)), closes [#52](https://github.com/vacp2p/staking-reward-streamer/issues/52)
* **RewardsStreamerMP:** ensure `registerVault` reverts in emergency mode ([fb79e24](https://github.com/vacp2p/staking-reward-streamer/commit/fb79e249fb3028c9adc36d67dabe920b2a3bc8a1))
* **RewardsStreamerMP:** prevent attack causes accounts to not accrue MP ([619b541](https://github.com/vacp2p/staking-reward-streamer/commit/619b541d2a4e586bd8d0487dade57fb55e39615b)), closes [#176](https://github.com/vacp2p/staking-reward-streamer/issues/176)
* **RewardsStreamerMP:** rename _calculateAccruedRewards to _calculatePendingRewards and fix specs ([b1a4e5a](https://github.com/vacp2p/staking-reward-streamer/commit/b1a4e5ad37b537e315415387a6cde1b017385cbf))
* **RewardStreamerMP.t:** minor fix on use of variables ([eff15a8](https://github.com/vacp2p/staking-reward-streamer/commit/eff15a8ade1d8a807386308d3444f3d1198bc603))
