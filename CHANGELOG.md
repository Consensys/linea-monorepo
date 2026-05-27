## [1.1.0-rc1] - 2026-05-27

### [coordinator] - [v0.1.1-rc1](https://github.com/Consensys/linea-monorepo/releases/tag/releases%2Fcoordinator%2Fv0.1.1-rc1)

#### 🐛 Bug Fixes

- *(coordinator, jvm-libs, e2e, state-recovery, prover, docker, misc)* Remove state manager request version (#3099)
- *(coordinator)* Remove traces version from requests (#3110)
- *(coordinator)* Export FTX number metrics (#3165)

#### ⚙️ Miscellaneous Tasks

- *(coordinator)* Make persistence module flat (#3066)
- *(coordinator)* Rename package net.consensys.zkevm.persistence to linea.persistence #3073
- *(2876)* Rename catch variable from it to e in GoBackedBlobShnarfCalculator (#2889)
- *(2876)* Coordinator review fixes — dead code, null safety, exception handling, dedup (#2882)
- *(coordinator)* Move Web3SignerTxSignService into web3j-extensions lib (#3091)
- *(coordinator)* Remove "build" prefix from package names
- *(coordinator)* Rename packages net.consensys.zkevm.* -> linea.* (#3105)
- *(coordinator)* Log and message error improvements (#3193)
### [maru] - [v1.0.1-rc1](https://github.com/Consensys/linea-monorepo/releases/tag/releases%2Fmaru%2Fv1.0.1-rc1)

#### 🚀 Features

- Update dockerfile, makefile, and added workflow to build image (#103)
- Support using maru container in e2e tests (#116)
- Assign prevRandao in each block (#199)
- Added error handling for P2P RPC messages (#265)
- Added compression on gossip and P2P response messages (#291)
- Add configs to control TargetChainHeadCalculator and DownloadPe… (#312)
- Update maru genesis for chaos testing (#365)
- Update erigon version to v3.0.2 (#361)
- Added size limit check when constructing beacon blocks by range… (#410)
- Added builtin network configuration for Linea sepolia and Linea… (#453)
- Add test for multiple overrides config files (#457)
- Update genesis file for Osaka on sepolia (#477)
- El block number validation (#490)
- Decouple EL sync from CL sync for validator nodes (#499)
- *(maru)* Import maru beacon chain client into linea-monorepo

#### 🐛 Bug Fixes

- Reference parentBeaconBlock.beaconBlockBody (#213)
- Cache SignatureAlgorithmFactory and fix FCU async race in block … (#498)
- #506, NoClassDefFoundError: org/web3j/abi/datatypes/CustomError (#507)
- *(qbft)* Ignore benign duplicate gossip failures (#514)
- *(maru)* Address PR 3126 workflow follow-ups (#3164)

#### ⚙️ Miscellaneous Tasks

- Try fix concurrency issues (#332)
- Revert change that broke ci (#333)
- Enable security code scanner (#484)
- Changelog (#492)
- Changelog update (#504)
- *(ci)* Update github runners (#512)
- *(deps)* Align stack with linea-monorepo (Vert.x 5, Besu 26.5, Teku 25.12) (#510)
- *(misc)* Align spotless with linea-monorepo (#516)
### [postman] - [v1.0.1-rc1](https://github.com/Consensys/linea-monorepo/releases/tag/releases%2Fpostman%2Fv1.0.1-rc1)

#### ⚙️ Miscellaneous Tasks

- *(deps)* Refresh monorepo dependencies (#3061)
- *(deps)* Update Jest to 30.4 (#3077)
### [prover] - [v1.0.1-rc1](https://github.com/Consensys/linea-monorepo/releases/tag/releases%2Fprover%2Fv1.0.1-rc1)

#### 🐛 Bug Fixes

- *(coordinator, jvm-libs, e2e, state-recovery, prover, docker, misc)* Remove state manager request version (#3099)
- *(ci)* Provide correct path to rlp_blocks.bin (#3125)
- *(prover)* Update rlp_blocks.bin path in shnarf_calculator tests (#3129)
- *(prover)* Stronger soundness binding for euclidean division and crumb decomposition (#2910)
- *(prover)* Valid-nonce-ftx (#3179)

#### ⚙️ Miscellaneous Tasks

- Update gnark (#3089)
- Update to latest gnark and gnark-crypto (#3142)
- Update gnark dependency (#3215)
### [tx-exclusion-api] - [v1.0.1-rc1](https://github.com/Consensys/linea-monorepo/releases/tag/releases%2Ftx-exclusion-api%2Fv1.0.1-rc1)

#### ⚙️ Miscellaneous Tasks

- *(coordinator)* Rename package net.consensys.zkevm.persistence to linea.persistence #3073
### [linea-besu-package] - [v1.0.1-rc1](https://github.com/Consensys/linea-monorepo/releases/tag/releases%2Flinea-besu-package%2Fv1.0.1-rc1)

#### 🐛 Bug Fixes

- *(sequencer)* Bypass background scheduler collision in buildNewBlockAndWait(Long) (#3072)
- *(tracer)* No parallelism (#2957)

#### ⚙️ Miscellaneous Tasks

- *(linea-besu)* Move besu project under nested directory (#3063)
- *(linea-besu)* Move linea-besu-package into linea-besu/package (#3069)
- *(linea-besu)* Move besu-plugins into linea-besu/plugins (#3075)
- *(misc)* Besu-plugin acceptance test cleanup of deadcode (#3152)
- *(maru)* Remove spring dependency management (#3205)
