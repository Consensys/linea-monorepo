# Changelog

## 0.5.3-beta
* fix(return): OOGX for return for deployment + call SHAKIRA for return + cleaning ([#1105](https://github.com/Consensys/linea-tracer/pull/1105))
* Failing tests file ([#1098](https://github.com/Consensys/linea-tracer/pull/1098))
* fix(oob): inst modexp pircing f of max ([#1107](https://github.com/Consensys/linea-tracer/pull/1107))
* chore: add validation and usage for bulk capture ([#1096](https://github.com/Consensys/linea-tracer/pull/1096))
* TX_SKIP: canonical snapshots + preconditions ([#1094](https://github.com/Consensys/linea-tracer/pull/1094))
* feat: update replay tests with tx outcomes ([#1091](https://github.com/Consensys/linea-tracer/pull/1091))
* Tweak Memory Settings for Go Corset
* fix(selfdestruct): invoke selfDestructAtTransactionEnd ([#1022](https://github.com/Consensys/linea-tracer/pull/1022))
* Renaming zkevm-constraints to linea-constraints + .gitmodules update ([#1083](https://github.com/Consensys/linea-tracer/pull/1083))

## 0.5.2-beta
* fix(callDataCopy): source ID ([#1081](https://github.com/Consensys/linea-tracer/pull/1081))
* add a binary search script ([#1073](https://github.com/Consensys/linea-tracer/pull/1073))
* chore: add more description for bulkReplay ([#1058](https://github.com/Consensys/linea-tracer/pull/1058))
* Leo's new ranges ([#1051](https://github.com/Consensys/linea-tracer/pull/1051))
* fix: precompiles are considered COLD when using BESU frame method ([#1079](https://github.com/Consensys/linea-tracer/pull/1079))
* perf: more efficient data storage structure ([#1034](https://github.com/Consensys/linea-tracer/pull/1034))
* [997] Migrate most of the tests to ToyExecutionEnviromentV2 ([#1066](https://github.com/Consensys/linea-tracer/pull/1066))
* Feat/issue 975/ec data add relevant limits for ecpairing ([#993](https://github.com/Consensys/linea-tracer/pull/993))
* fix(mmu call): need CFI for failed CREATE2 that doesn't trigger the ROM ([#1070](https://github.com/Consensys/linea-tracer/pull/1070))
* fix(selfdestruct): wrap ([#1074](https://github.com/Consensys/linea-tracer/pull/1074))
* Feat: bulk capture and bulk replay ([#1001](https://github.com/Consensys/linea-tracer/pull/1001))
* Fix initial contract nonce ([#1056](https://github.com/Consensys/linea-tracer/pull/1056))
* fix(precompile): defers with wrong id ([#1055](https://github.com/Consensys/linea-tracer/pull/1055))
* Use `ReferenceTestWorldState` instead of `ToyWorld` ([#1045](https://github.com/Consensys/linea-tracer/pull/1045))
* fix: many things ... ([#1027](https://github.com/Consensys/linea-tracer/pull/1027))
* [934] Use ToyExecutionEnvironmentV2 for blockhash test ([#1040](https://github.com/Consensys/linea-tracer/pull/1040))
* fix: update `Trace.java` file for Hub ([#1037](https://github.com/Consensys/linea-tracer/pull/1037))
* Tests/ethereum ([#1012](https://github.com/Consensys/linea-tracer/pull/1012))
* Expose true output from tx result failure ([#1030](https://github.com/Consensys/linea-tracer/pull/1030))
* format(txndata): ras ([#1032](https://github.com/Consensys/linea-tracer/pull/1032))
* feat: update gradle test config to show better output ([#1021](https://github.com/Consensys/linea-tracer/pull/1021))
* fix(mmu): anyToRamWithPadding mixed subcase ([#1016](https://github.com/Consensys/linea-tracer/pull/1016))
* Fix for `BLOCKHASH` conversion error ([#1025](https://github.com/Consensys/linea-tracer/pull/1025))
* perf(MMU): discard CallStackReader ([#988](https://github.com/Consensys/linea-tracer/pull/988))
* feat: capture block hashes in `BlockCapturer` for replay ([#1019](https://github.com/Consensys/linea-tracer/pull/1019))
* add tests for exceptions. ([#1007](https://github.com/Consensys/linea-tracer/pull/1007))
* Fix #998 ([#1015](https://github.com/Consensys/linea-tracer/pull/1015))
* Update Linea-Besu dependency to 24.8-develop-915fcb01 (delivery-29) ([#1011](https://github.com/Consensys/linea-tracer/pull/1011))
* feat: update blockcapturer to capture outcomes ([#982](https://github.com/Consensys/linea-tracer/pull/982))
* Fix stp fix ([#1003](https://github.com/Consensys/linea-tracer/pull/1003))
* Make `ReturnSection`'s `resolveAtContextReEntry` method use the createe frame ([#992](https://github.com/Consensys/linea-tracer/pull/992))
* feat(corset): enable `go-corset` in `gradle` action ([#938](https://github.com/Consensys/linea-tracer/pull/938))
* fix(STP): fixed exists
* fix(StpCall): made STP fillers private
* fix(CallSection): removed redundant stpCallForCalls call
* fix(CreateSection): removed redundant stpCallForCreates call

## 0.5.1-beta
* fix: removed old documentation files ([#995](https://github.com/Consensys/linea-tracer/pull/995))
* feat(txndata): implement txn data update handling the max nonce ([#987](https://github.com/Consensys/linea-tracer/pull/987))
* feat(hub): fill pMiscOobData9 ([#985](https://github.com/Consensys/linea-tracer/pull/985))
* feat(ecdata): use constants for call data sizes ([#983](https://github.com/Consensys/linea-tracer/pull/983))
* fix(hub): static exceptions that should not be thrown. ([#986](https://github.com/Consensys/linea-tracer/pull/986))
* feat(ecdata): switch from stacked set to stacked list ([#981](https://github.com/Consensys/linea-tracer/pull/981))
* fix(Makefile): updated file path to AccountFragment.java

## 0.5.0-beta
* Merged the HUB branch ([#748](https://github.com/Consensys/linea-tracer/pull/748))

## 0.4.0-rc2
* fix: change implementation of gasAvailableForChildCall due to having side effects in TangerineWhistleGasCalculator ([#950](https://github.com/Consensys/linea-tracer/pull/950))
* fix: `ToyWorld.commit()` ([#966](https://github.com/Consensys/linea-tracer/pull/966))
* feat(ecadd): add test ([#956](https://github.com/Consensys/linea-tracer/pull/956))
* fix: Use correct CHAINID in EVM ([#947](https://github.com/Consensys/linea-tracer/pull/947))
* fix: swap ordering of r/s fields in tx snapshot ([#946](https://github.com/Consensys/linea-tracer/pull/946))

## 0.4.0-rc1
* feat(toy-exec-env-v2): add new ToyExecutionEnvironment that builds the General State Test Case spec to run tests ([#842](https://github.com/Consensys/linea-tracer/pull/842))
* feat: initial Integration of Go Corset ([#907](https://github.com/Consensys/linea-tracer/pull/907))
* feat(exp): update ([#937](https://github.com/Consensys/linea-tracer/pull/937))
* fix(blockCapturer): missing handling of selfdestruct ([#936](https://github.com/Consensys/linea-tracer/pull/936))
* docs: retires zk-EVM ([#903](https://github.com/Consensys/linea-tracer/pull/903))
* fix: add replay test for incident 777 on zkGeth mainnet ([#927](https://github.com/Consensys/linea-tracer/pull/927))
* test(ecpairing): implement extensive test for ecpairing ([#822](https://github.com/Consensys/linea-tracer/pull/822))([#909](https://github.com/Consensys/linea-tracer/pull/909))

## 0.3.0-rc2
* feat: Update Linea-Besu to 24.7-develop-c0029e6 ([#905](https://github.com/Consensys/linea-tracer/pull/905))

## 0.3.0-rc1
* feat: upgrade besu version to 24.7-develop-f812936 ([#880](https://github.com/Consensys/linea-arithmetization/pull/880))

## 0.2.0-rc5
* fix(ecData): ugly hack to discard unsuccessful EcRecover call ([#891](https://github.com/Consensys/linea-arithmetization/pull/891))

## 0.2.0-rc4
* fix: init config object only once ([#873](https://github.com/Consensys/linea-arithmetization/pull/873))
* feat: improve design of shared and private CLI options ([#864](https://github.com/Consensys/linea-arithmetization/pull/864))

## 0.2.0-rc3
* fix: make --plugin-linea-conflated-trace-generation-traces-output-path option required to avoid faulty registration of the trace generation RPC endpoint ([#858](https://github.com/Consensys/linea-arithmetization/pull/858))
* feat: separate shared and private CLI options ([#856](https://github.com/Consensys/linea-arithmetization/pull/856))

## 0.2.0-rc2
* feat: improve ZkTracer initialization time by doing only once Opcodes and spillings loading from disk resources ([#720](https://github.com/Consensys/linea-arithmetization/pull/720))
* perf: parallelize refundedGas for big transactions ([#793](https://github.com/Consensys/linea-arithmetization/pull/793))

## 0.2.0-rc1
* feat: add PRECOMPILE_ECPAIRING_G2_MEMBERSHIP_CALLS in spillings.toml and did some renaming ([#819](https://github.com/Consensys/linea-arithmetization/pull/819))
* feat: optimise trace generation (except hub) ([#838](https://github.com/Consensys/linea-arithmetization/pull/838))

## 0.1.5-rc6
* Migrating of `TRACES_DIR` env var to 
`plugin-linea-conflated-trace-generation-traces-output-path` CLI option that can be included in the toml config files.
The path specified in `plugin-linea-conflated-trace-generation-traces-output-path` will be created automatically if it does not exist.
This time this has nothing to do with the `ContinuousTracingPlugin` [#830](https://github.com/Consensys/linea-arithmetization/pull/830).

## 0.1.5-rc5
* Migrating of `TRACES_DIR` env var to `plugin-linea-continuous-tracing-traces-dir` CLI option that can be included in the 
toml config files. The path specified in `plugin-linea-continuous-tracing-traces-dir` will be created automatically 
  if it does not exist [#825](https://github.com/Consensys/linea-arithmetization/pull/825).

## 0.1.4-test21
Test pre-release 21 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix: capture SSTORE-touched storage slots for correct gas computations [#606](https://github.com/Consensys/linea-arithmetization/pull/606)
* build: make the build script portable, explicit dependency on Go & GCC, test libcompress build [#621](https://github.com/Consensys/linea-arithmetization/pull/621)
* Update after the refactor of transaction selection service [#626](https://github.com/Consensys/linea-arithmetization/pull/626)
* Use the right classloader to load the native library [#628](https://github.com/Consensys/linea-arithmetization/pull/628)

## 0.1.4-test20
Test pre-release 20 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Get L2L1 settings from CLI options [#591](https://github.com/Consensys/linea-arithmetization/pull/591)
* feat: add a replay capture script [#600](https://github.com/Consensys/linea-arithmetization/pull/600)
* move compress native into plugin repo [#604](https://github.com/Consensys/linea-arithmetization/pull/604)
* Add compression [#605](https://github.com/Consensys/linea-arithmetization/pull/605)
* Update for the new bad block manager [#607](https://github.com/Consensys/linea-arithmetization/pull/607)

## 0.1.4-test19
Test pre-release 19 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Avoid returning an estimated priority fee that is less than the min gas price [#598](https://github.com/Consensys/linea-arithmetization/pull/598)

## 0.1.4-test18
Test pre-release 18 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix: check that spilling and limits file contain all counted modules [#592](https://github.com/Consensys/linea-arithmetization/pull/592)

## 0.1.4-test18-RC3
Test pre-release 18-RC3 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
*  Use compressed tx size also when selecting txs from block creation [#590](https://github.com/Consensys/linea-arithmetization/pull/590)

## 0.1.4-test18-RC2
Test pre-release 18-RC2 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
*  Fix linea_estimateGas reports Internal error when value or gas price is missing [#587](https://github.com/Consensys/linea-arithmetization/pull/587)

## 0.1.4-test18-RC1
Test pre-release 18-RC1 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Linea estimate gas endpoint [#585](https://github.com/Consensys/linea-arithmetization/pull/585)

## 0.1.4-test17
Test pre-release 17 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* tests: drop huge random tests [#563](https://github.com/Consensys/linea-arithmetization/pull/563)
* feat(modexp-data): implement MODEXP_DATA module [#547](https://github.com/Consensys/linea-arithmetization/pull/547)
* feat: mechanics to capture conflations & replay them as test cases [#561](https://github.com/Consensys/linea-arithmetization/pull/561)
* perf(EUC): one less column [#570](https://github.com/Consensys/linea-arithmetization/pull/570)
* docs: Add basic plugins doc [#509](https://github.com/Consensys/linea-arithmetization/pull/509)
* Check upfront profitability + Unprofitable txs cache and retry limit [#565](https://github.com/Consensys/linea-arithmetization/pull/565)
* Avoid reprocessing txs that go over line count limit [#571](https://github.com/Consensys/linea-arithmetization/pull/571)

## 0.1.4-test16
Test pre-release 16 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix: bug-compatibility with Geth
* fix: PubHash 16 factor

Full changeset https://github.com/Consensys/linea-arithmetization/compare/v0.1.4-test15...v0.1.4-test16

## 0.1.4-test15
release rebase off of main
* add option to adjust the tx size used to calculate the profitability of a tx during block creation(#562)[https://github.com/Consensys/linea-arithmetization/pull/562]

## 0.1.4-test14
release rebase off of main
Test pre-release 14 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Fix log of line counts in case of block limit reached + minor changes [#555](https://github.com/ConsenSys/linea-arithmetization/pull/555)
* build: update Corset to 9.3.0 [#554](https://github.com/ConsenSys/linea-arithmetization/pull/554)

## 0.1.4-test13
Test pre-release 13 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix stackedSet [c3f226775f24508b93a758e4226a51ae386d76a5](https://github.com/Consensys/linea-arithmetization/commit/c3f226775f24508b93a758e4226a51ae386d76a5)

## 0.1.4-test12
Test pre-release 12 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix: stacked set multiple insertions in a single transaction (#548)

## 0.1.4-test11
Test pre-release 11 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* same as 0.1.4-test10

## 0.1.4-test10
Test pre-release 10 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix: semantics of LinkedList (#544)
* refactor: add @EqualsAndHashCode annotations and remove corresponding methods (#541)

## 0.1.4-test9
Test pre-release 9 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Bump Linea Besu to 24.1.1-SNAPSHOT

## 0.1.4-test8
Test pre-release 8 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Add profitable transaction selector [#530](https://github.com/Consensys/linea-arithmetization/pull/530)
* temp: geth-compatibily hacks [820918a](https://github.com/Consensys/linea-arithmetization/commit/820918a39e8d394e73b8de85a46391ffe7d314b1)

## 0.1.4-test7
Test pre-release 7 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* fix: invalid SStore gas computation [#532](https://github.com/Consensys/linea-arithmetization/pull/532)

## 0.1.4-test6
Test pre-release 6, fix: [make precompile counters work](https://github.com/Consensys/linea-arithmetization/commit/10f03ead5207746f253703a328f13988ed9b9305)
* feat: implement fake hashdata/info [Franklin Delehelle]
* temp: geth-compatibily hacks [Franklin Delehelle]
* refactor: group RLPs modules, use retro-compatible module keys [#508](https://github.com/ConsenSys/linea-arithmetization/pull/508)
* [MINOR] Add javadoc [#507](https://github.com/ConsenSys/linea-arithmetization/pull/507)
* style: update name of prec limits to avoid confusion with old geth name [#506](https://github.com/ConsenSys/linea-arithmetization/pull/506)
* perf: cache tx-specific line counter [#497](https://github.com/ConsenSys/linea-arithmetization/pull/497)
* fix: continuous tracing plugin start check [#500](https://github.com/ConsenSys/linea-arithmetization/pull/500)
* fix: lookup txndata <-> wcp [#488](https://github.com/ConsenSys/linea-arithmetization/pull/488)
* fix(romLex): wrong stack arg for extcodecopy address [#498](https://github.com/ConsenSys/linea-arithmetization/pull/498)

## 0.1.4-test3
Test pre-release 3 from [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
* Log ZkTracer counters for every produced block [#485](https://github.com/ConsenSys/linea-arithmetization/pull/485)
* fix: overflow for modexp arg [#489](https://github.com/ConsenSys/linea-arithmetization/pull/489)
* bin reimplementation [#473](https://github.com/ConsenSys/linea-arithmetization/pull/473)
* applyMavenExclusions=false [#477](https://github.com/ConsenSys/linea-arithmetization/pull/477)

## 0.1.4-test2
Testing pre-release from branch test-release/v0.1.4-test2

* revert make loginfo counts closer to Geth
* head: disable stp & txndata

## 0.1.4-test
Temporary line counting release for testnet.

* count stack temporary impl: make loginfo counts closer to Geth [temp/issue-248/count-stack-only](https://github.com/Consensys/linea-arithmetization/tree/temp/issue-248/count-stack-only)
  --
* fix: `Bytes.toUnsignedInteger` [#484](https://github.com/ConsenSys/linea-arithmetization/pull/484)
* perf: delay computations at trace time [#483](https://github.com/ConsenSys/linea-arithmetization/pull/483)

## 0.1.3
- perf: improve `StackedSet` performances  [#466](https://github.com/ConsenSys/linea-arithmetization/pull/466)
- feat: implement L1 block & Keccak limits [#445](https://github.com/ConsenSys/linea-arithmetization/pull/445)
- feat: partially implement EC_DATA [#475](https://github.com/ConsenSys/linea-arithmetization/pull/475)
- fix: ensure trace files are always deleted [#462](https://github.com/ConsenSys/linea-arithmetization/pull/462)


## 0.1.2
Release 8 for 23.10.4-SNAPSHOT of linea-besu
- changed default file name to toml [#476](https://github.com/ConsenSys/linea-arithmetization/pull/476)
- feat: implement `BIN` counting [#471](https://github.com/ConsenSys/linea-arithmetization/pull/471)
- Upgrade Linea Besu to 23.10.4-SNAPSHOT [#469](https://github.com/ConsenSys/linea-arithmetization/pull/469)
- fix: incorrect address comparison [#470](https://github.com/ConsenSys/linea-arithmetization/pull/470)
- fix: line count discrepancy [#468](https://github.com/ConsenSys/linea-arithmetization/pull/468)

## 0.1.1
Release for 23.10.3-SNAPSHOT of linea-besu

## 0.1.0
- Initial build of linea-arithmetization
- uses 23.10.3-SNAPSHOT as linea-besu version
