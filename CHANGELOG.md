## [1.1.0-mrc6] - 2026-05-18

### [coordinator] - 2026-05-18

#### 🚀 Features

- *(coordinator)* Testing bump version

#### 🐛 Bug Fixes

- *(coordinator, jvm-libs, e2e, state-recovery, prover, docker, misc)* Remove state manager request version (#3099)
- *(coordinator)* Remove traces version from requests (#3110)
- *(coordinator)* Merged conflicts on CHANGELOG.md

#### ⚙️ Miscellaneous Tasks

- *(coordinator)* Make persistence module flat (#3066)
- *(coordinator)* Rename package net.consensys.zkevm.persistence to linea.persistence #3073
- *(2876)* Rename catch variable from it to e in GoBackedBlobShnarfCalculator (#2889)
- *(2876)* Coordinator review fixes — dead code, null safety, exception handling, dedup (#2882)
- *(coordinator)* Move Web3SignerTxSignService into web3j-extensions lib (#3091)
- *(coordinator)* Remove "build" prefix from package names
- *(coordinator)* Bump version tesing
- *(coordinator)* Rename packages net.consensys.zkevm.* -> linea.* (#3105)
### [postman] - 2026-05-18

#### 🐛 Bug Fixes

- *(coordinator)* Merged conflicts on CHANGELOG.md

#### ⚙️ Miscellaneous Tasks

- *(deps)* Refresh monorepo dependencies (#3061)
- *(deps)* Update Jest to 30.4 (#3077)
### [prover] - 2026-05-18

#### 🐛 Bug Fixes

- *(coordinator, jvm-libs, e2e, state-recovery, prover, docker, misc)* Remove state manager request version (#3099)
- *(ci)* Provide correct path to rlp_blocks.bin (#3125)
- *(prover)* Update rlp_blocks.bin path in shnarf_calculator tests (#3129)
- *(coordinator)* Merged conflicts on CHANGELOG.md

#### ⚙️ Miscellaneous Tasks

- Update gnark (#3089)
### [tx-exclusion-api] - 2026-05-18

#### ⚙️ Miscellaneous Tasks

- *(coordinator)* Rename package net.consensys.zkevm.persistence to linea.persistence #3073
### [linea-besu] - 2026-05-18

#### 🐛 Bug Fixes

- *(sequencer)* Bypass background scheduler collision in buildNewBlockAndWait(Long) (#3072)
- *(tracer)* No parallelism (#2957)
- *(coordinator, jvm-libs, e2e, state-recovery, prover, docker, misc)* Remove state manager request version (#3099)
- *(coordinator)* Merged conflicts on CHANGELOG.md

#### ⚙️ Miscellaneous Tasks

- *(linea-besu)* Move besu project under nested directory (#3063)
- *(linea-besu)* Move linea-besu-package into linea-besu/package (#3069)
- *(linea-besu)* Move besu-plugins into linea-besu/plugins (#3075)
- *(coordinator)* Remove "build" prefix from package names
- *(coordinator)* Rename packages net.consensys.zkevm.* -> linea.* (#3105)
