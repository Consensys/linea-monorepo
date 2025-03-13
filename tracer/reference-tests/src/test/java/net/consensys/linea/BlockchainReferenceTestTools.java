/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea;

import static net.consensys.linea.BlockchainReferenceTestJson.readBlockchainReferenceTestsOutput;
import static net.consensys.linea.ReferenceTestOutcomeRecorderTool.JSON_INPUT_FILENAME;
import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ConcurrentSkipListSet;
import java.util.concurrent.ExecutionException;
import java.util.stream.Collectors;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.reporting.TestOutcome;
import net.consensys.linea.reporting.TestOutcomeWriterTool;
import net.consensys.linea.testing.ExecutionEnvironment;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.ethereum.MainnetBlockValidator;
import org.hyperledger.besu.ethereum.ProtocolContext;
import org.hyperledger.besu.ethereum.chain.MutableBlockchain;
import org.hyperledger.besu.ethereum.core.Block;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.MutableWorldState;
import org.hyperledger.besu.ethereum.mainnet.BlockImportResult;
import org.hyperledger.besu.ethereum.mainnet.HeaderValidationMode;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockImporter;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSchedule;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.referencetests.BlockchainReferenceTestCaseSpec;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestProtocolSchedules;
import org.hyperledger.besu.ethereum.rlp.RLPException;
import org.hyperledger.besu.ethereum.trie.diffbased.common.provider.WorldStateQueryParams;
import org.hyperledger.besu.testutil.JsonTestParameters;
import org.junit.jupiter.api.Assumptions;

@Slf4j
public class BlockchainReferenceTestTools {
  private static final ReferenceTestProtocolSchedules REFERENCE_TEST_PROTOCOL_SCHEDULES =
      ReferenceTestProtocolSchedules.create();

  private static final List<String> NETWORKS_TO_RUN = List.of("London");

  public static final JsonTestParameters<?, ?> PARAMS =
      JsonTestParameters.create(BlockchainReferenceTestCaseSpec.class)
          .generator(
              (testName, fullPath, spec, collector) -> {
                final String eip = spec.getNetwork();
                collector.add(
                    testName + "[" + eip + "]", fullPath, spec, NETWORKS_TO_RUN.contains(eip));
              });

  private static final CorsetValidator CORSET_VALIDATOR = new CorsetValidator(ChainConfig.ETHEREUM);

  static {
    if (NETWORKS_TO_RUN.isEmpty()) {
      PARAMS.ignoreAll();
    }
    // ignore tests that are failing in Besu too
    PARAMS.ignore("RevertInCreateInInitCreate2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("RevertInCreateInInit_d0g0v0_London\\[London\\]");
    PARAMS.ignore("create2collisionStorage_d0g0v0_London\\[London\\]");
    PARAMS.ignore("create2collisionStorage_d1g0v0_London\\[London\\]");
    PARAMS.ignore("create2collisionStorage_d2g0v0_London\\[London\\]");
    PARAMS.ignore("dynamicAccountOverwriteEmpty_d0g0v0_London\\[London\\]");

    // ignore tests that are failing because there is an account with nonce 0 and
    // non empty code which can't happen in Linea since we are post LONDON
    PARAMS.ignore("InitCollision_d0g0v0_London\\[London\\]");
    PARAMS.ignore("InitCollision_d1g0v0_London\\[London\\]");
    PARAMS.ignore("InitCollision_d2g0v0_London\\[London\\]");
    PARAMS.ignore("InitCollision_d3g0v0_London\\[London\\]");
    PARAMS.ignore("RevertInCreateInInitCreate2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("RevertInCreateInInit_d0g0v0_London\\[London\\]");

    // Arithmetization restriction: recipient address is a precompile.
    PARAMS.ignore("modexpRandomInput_d0g0v0_London\\[London\\]");
    PARAMS.ignore("modexpRandomInput_d0g1v0_London\\[London\\]");
    PARAMS.ignore("modexpRandomInput_d1g0v0_London\\[London\\]");
    PARAMS.ignore("modexpRandomInput_d1g1v0_London\\[London\\]");
    PARAMS.ignore("modexpRandomInput_d2g0v0_London\\[London\\]");
    PARAMS.ignore("modexpRandomInput_d2g1v0_London\\[London\\]");
    PARAMS.ignore("randomStatetest642_d0g0v0_London\\[London\\]");
    PARAMS.ignore("randomStatetest644_d0g0v0_London\\[London\\]");
    PARAMS.ignore("randomStatetest645_d0g0v0_London\\[London\\]");
    PARAMS.ignore("randomStatetest645_d0g0v1_London\\[London\\]");

    // Consumes a huge amount of memory.
    PARAMS.ignore("static_Call1MB1024Calldepth_d1g0v0_\\w+");
    PARAMS.ignore("ShanghaiLove_.*");
    PARAMS.ignore("/GeneralStateTests/VMTests/vmPerformance/");
    PARAMS.ignore("Call50000");
    PARAMS.ignore("static_LoopCallsDepthThenRevert3");
    PARAMS.ignore("Return50000");

    // Absurd amount of gas, doesn't run in parallel.
    PARAMS.ignore("randomStatetest94_\\w+");

    // Balance is more than 128 bits
    PARAMS.ignore("CALLCODE_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds3_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds4_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds4_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds4_d0g2v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALLCODE_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds2a_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds2a_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds3_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds3_d0g2v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CALL_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds3_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds3_d0g2v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CREATE2_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds3_d0g1v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds3_d0g2v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("CREATE_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("Call1024PreCalls_d0g0v0_London\\[London\\]");
    PARAMS.ignore("Call1024PreCalls_d0g1v0_London\\[London\\]");
    PARAMS.ignore("Call1024PreCalls_d0g2v0_London\\[London\\]");
    PARAMS.ignore("Create2OnDepth1023_d0g0v0_London\\[London\\]");
    PARAMS.ignore("Create2OnDepth1024_d0g0v0_London\\[London\\]");
    PARAMS.ignore("Create2Recursive_d0g0v0_London\\[London\\]");
    PARAMS.ignore("Create2Recursive_d0g1v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds3_d0g1v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds3_d0g2v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("DELEGATECALL_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("DelegateCallSpam_London\\[London\\]");
    PARAMS.ignore("HighGasLimit_d0g0v0_London\\[London\\]");
    PARAMS.ignore("MSTORE_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("MSTORE_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("MSTORE_Bounds2a_d0g0v0_London\\[London\\]");
    PARAMS.ignore("MSTORE_Bounds2a_d0g1v0_London\\[London\\]");
    PARAMS.ignore("MSTORE_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("MSTORE_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("OutOfGasContractCreation_d0g0v0_London\\[London\\]");
    PARAMS.ignore("OutOfGasContractCreation_d0g1v0_London\\[London\\]");
    PARAMS.ignore("OutOfGasContractCreation_d1g0v0_London\\[London\\]");
    PARAMS.ignore("OutOfGasContractCreation_d1g1v0_London\\[London\\]");
    PARAMS.ignore("OverflowGasRequire2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("OverflowGasRequire_London\\[London\\]");
    PARAMS.ignore("RETURN_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("RETURN_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("RETURN_Bounds_d0g2v0_London\\[London\\]");
    PARAMS.ignore("StrangeContractCreation_London\\[London\\]");
    PARAMS.ignore("SuicideIssue_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds2_d0g1v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds2a_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds2a_d0g1v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds3_d0g1v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_CALL_Bounds_d0g1v0_London\\[London\\]");
    PARAMS.ignore("static_Call1024PreCalls2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_Call1024PreCalls2_d1g0v0_London\\[London\\]");
    PARAMS.ignore("static_Call1024PreCalls3_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_Call1024PreCalls3_d1g0v0_London\\[London\\]");
    PARAMS.ignore("static_Call1024PreCalls_d1g0v0_London\\[London\\]");
    PARAMS.ignore("static_RETURN_BoundsOOG_d0g0v0_London\\[London\\]");
    PARAMS.ignore("static_RETURN_BoundsOOG_d1g0v0_London\\[London\\]");
    PARAMS.ignore("static_RETURN_Bounds_d0g0v0_London\\[London\\]");

    // Deployment transaction to an account with nonce / code
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g0v0_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g0v1_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g1v0_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g1v1_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g0v0_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g0v1_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g1v0_London\\[London\\]");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g1v1_London\\[London\\]");
    PARAMS.ignore("createJS_ExampleContract_d0g0v0_London\\[London\\]");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d0g0v0_London\\[London\\]");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d1g0v0_London\\[London\\]");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d2g0v0_London\\[London\\]");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d3g0v0_London\\[London\\]");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d4g0v0_London\\[London\\]");

    // Deployment transaction to an account with zero nonce, empty code (and zero balance) but
    // nonempty storage. Given [EIP-7610](https://github.com/ethereum/EIPs/pull/8161), no Besu
    // execution takes place, which means that no TraceSection's are created beyond the
    // {@link TxInitializationSection}. This triggers a NPE when tracing, as at some point
    // {@link TraceSection#nextSection} is null in {@link TraceSection#computeContextNumberNew()}.
    PARAMS.ignore("FailedCreateRevertsDeletion_d0g0v0_London\\[London\\]");

    // Ignore the following test as it is not supported in Linea.
    // See [issue #1678](https://github.com/Consensys/linea-tracer/issues/1678)
    PARAMS.ignore("suicideStorageCheck_London\\[London\\]");

    // Don't do time-consuming tests.
    PARAMS.ignore("CALLBlake2f_MaxRounds.*");
    PARAMS.ignore("loopMul_*");

    // Inconclusive fork choice rule, since in merge CL should be choosing forks and setting the
    // chain head. Perfectly valid test pre-merge.
    PARAMS.ignore("UncleFromSideChain_(Merge|Shanghai|Cancun|Prague|Osaka|Bogota)");

    // EOF tests are written against an older version of the spec.
    PARAMS.ignore("/stEOF/");

    // We ignore the following tests because they satisfy one of the following:
    // - bbs > 512, bbs ≡ base byte size
    // - ebs > 512, ebs ≡ exponent byte size
    // - mbs > 512, mbs ≡ modulus byte size
    PARAMS.ignore("modexp_d28g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_d28g1v0_London\\[London\\]");
    PARAMS.ignore("modexp_d28g2v0_London\\[London\\]");
    PARAMS.ignore("modexp_d28g3v0_London\\[London\\]");
    PARAMS.ignore("modexp_d29g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_d29g1v0_London\\[London\\]");
    PARAMS.ignore("modexp_d29g2v0_London\\[London\\]");
    PARAMS.ignore("modexp_d29g3v0_London\\[London\\]");
    PARAMS.ignore("modexp_d2g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_d2g1v0_London\\[London\\]");
    PARAMS.ignore("modexp_d2g2v0_London\\[London\\]");
    PARAMS.ignore("modexp_d2g3v0_London\\[London\\]");
    PARAMS.ignore("modexp_d30g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_d30g1v0_London\\[London\\]");
    PARAMS.ignore("modexp_d30g2v0_London\\[London\\]");
    PARAMS.ignore("modexp_d30g3v0_London\\[London\\]");
    PARAMS.ignore("modexp_d36g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_d36g1v0_London\\[London\\]");
    PARAMS.ignore("modexp_d36g2v0_London\\[London\\]");
    PARAMS.ignore("modexp_d36g3v0_London\\[London\\]");
    PARAMS.ignore("modexp_d37g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_d37g1v0_London\\[London\\]");
    PARAMS.ignore("modexp_d37g2v0_London\\[London\\]");
    PARAMS.ignore("modexp_d37g3v0_London\\[London\\]");
    PARAMS.ignore("idPrecomps_d4g0v0_London\\[London\\]");
    PARAMS.ignore("modexp_modsize0_returndatasize_d4g0v0_London\\[London\\]");
    PARAMS.ignore("randomStatetest650_d0g0v0_London\\[London\\]");

    // unsupported behaviour: uncle blocks, re-orgs, forks, side chain (?)
    PARAMS.ignore("ChainAtoChainBCallContractFormA_London\\[London\\]");
    PARAMS.ignore("ChainAtoChainB_London\\[London\\]");
    PARAMS.ignore("ChainAtoChainB_difficultyB_London\\[London\\]");
    PARAMS.ignore("ChainAtoChainBtoChainA_London\\[London\\]");
    PARAMS.ignore("ForkStressTest_London\\[London\\]");
    PARAMS.ignore("newChainFrom4Block_London\\[London\\]");
    PARAMS.ignore("newChainFrom5Block_London\\[London\\]");
    PARAMS.ignore("newChainFrom6Block_London\\[London\\]");
    PARAMS.ignore("sideChainWithMoreTransactions2_London\\[London\\]");
    PARAMS.ignore("sideChainWithMoreTransactions_London\\[London\\]");
    PARAMS.ignore("sideChainWithNewMaxDifficultyStartingFromBlock3AfterBlock4_London\\[London\\]");
    PARAMS.ignore("uncleBlockAtBlock3AfterBlock3_London\\[London\\]");
    PARAMS.ignore("uncleBlockAtBlock3afterBlock4_London\\[London\\]");

    // not sure what these tests are doing, but they blow up BLOCK_DATA, which is the simplest
    // module in existence
    PARAMS.ignore("CallContractFromNotBestBlock_London\\[London\\]");
    PARAMS.ignore("RPC_API_Test_London\\[London\\]");

    // the following tests blow up due monetary creation pre PoS where the COINBASE would get paid 2
    // Eth at the end of every block
    // TODO: re-enable post Paris
    PARAMS.ignore("correct_London\\[London\\]");
    PARAMS.ignore("incorrectUncleTimestamp4_London\\[London\\]");
    PARAMS.ignore("incorrectUncleTimestamp5_London\\[London\\]");
    PARAMS.ignore("timestampTooHigh_London\\[London\\]");
    PARAMS.ignore("timestampTooLow_London\\[London\\]");
    PARAMS.ignore("futureUncleTimestamp3_London\\[London\\]");
    PARAMS.ignore("wrongStateRoot_London\\[London\\]");
    PARAMS.ignore("besuBaseFeeBug_London\\[London\\]");
    PARAMS.ignore("burnVerifyLondon_London\\[London\\]");
    PARAMS.ignore("highDemand_London\\[London\\]");
    PARAMS.ignore("intrinsic_London\\[London\\]");
    PARAMS.ignore("intrinsicTip_London\\[London\\]");
    PARAMS.ignore("medDemand_London\\[London\\]");
    PARAMS.ignore("tipsLondon_London\\[London\\]");
    PARAMS.ignore("transType_London\\[London\\]");
    PARAMS.ignore("highGasUsage_London\\[London\\]");
    PARAMS.ignore("blockhashNonConstArg_London\\[London\\]");
    PARAMS.ignore("blockhashTests_London\\[London\\]");
    PARAMS.ignore("extcodehashEmptySuicide_London\\[London\\]");
    PARAMS.ignore("logRevert_London\\[London\\]");
    PARAMS.ignore("multimpleBalanceInstruction_London\\[London\\]"); // typo intended
    PARAMS.ignore("refundReset_London\\[London\\]");
    PARAMS.ignore("simpleSuicide_London\\[London\\]");
    PARAMS.ignore("futureUncleTimestamp2_London\\[London\\]");
    PARAMS.ignore("futureUncleTimestampDifficultyDrop_London\\[London\\]");
    PARAMS.ignore("futureUncleTimestampDifficultyDrop2_London\\[London\\]");
    PARAMS.ignore("futureUncleTimestampDifficultyDrop3_London\\[London\\]");
    PARAMS.ignore("futureUncleTimestampDifficultyDrop4_London\\[London\\]");
    PARAMS.ignore("uncleBloomNot0_2_London\\[London\\]");
    PARAMS.ignore("uncleBloomNot0_London\\[London\\]");
    PARAMS.ignore("oneUncle_London\\[London\\]");
    PARAMS.ignore("oneUncleGeneration2_London\\[London\\]");
    PARAMS.ignore("oneUncleGeneration3_London\\[London\\]");
    PARAMS.ignore("oneUncleGeneration4_London\\[London\\]");
    PARAMS.ignore("uncleBloomNot0_3_London\\[London\\]");
    PARAMS.ignore("oneUncleGeneration5_London\\[London\\]");
    PARAMS.ignore("oneUncleGeneration6_London\\[London\\]");
    PARAMS.ignore("twoUncle_London\\[London\\]");
    PARAMS.ignore("uncleHeaderAtBlock2_London\\[London\\]");
    PARAMS.ignore("RecallSuicidedContract_London\\[London\\]");
    PARAMS.ignore("RecallSuicidedContractInOneBlock_London\\[London\\]");
    PARAMS.ignore("timeDiff12_London\\[London\\]");
    PARAMS.ignore("timeDiff13_London\\[London\\]");
    PARAMS.ignore("timeDiff14_London\\[London\\]");
    PARAMS.ignore("wallet2outOf3txs_London\\[London\\]");
    PARAMS.ignore("wallet2outOf3txs2_London\\[London\\]");
    PARAMS.ignore("wallet2outOf3txsRevoke_London\\[London\\]");
    PARAMS.ignore("wallet2outOf3txsRevokeAndConfirmAgain_London\\[London\\]");
    PARAMS.ignore("walletReorganizeOwners_London\\[London\\]");
  }

  private BlockchainReferenceTestTools() {
    // utility class
  }

  public static CompletableFuture<Set<String>> getRecordedFailedTestsFromJson(
      String failedModule, String failedConstraint) {
    Set<String> failedTests = new HashSet<>();
    if (failedModule.isEmpty()) {
      return CompletableFuture.completedFuture(failedTests);
    }

    CompletableFuture<TestOutcome> modulesToConstraintsFutures =
        readBlockchainReferenceTestsOutput(JSON_INPUT_FILENAME)
            .thenApply(TestOutcomeWriterTool::parseTestOutcome);

    return modulesToConstraintsFutures.thenApply(
        blockchainReferenceTestOutcome -> {
          ConcurrentMap<String, ConcurrentSkipListSet<String>> filteredFailedTests =
              blockchainReferenceTestOutcome.getModulesToConstraintsToTests().get(failedModule);
          if (filteredFailedTests == null) {
            return failedTests;
          }
          if (!failedConstraint.isEmpty()) {
            return filteredFailedTests.get(failedConstraint);
          }
          return filteredFailedTests.values().stream()
              .flatMap(Set::stream)
              .collect(Collectors.toSet());
        });
  }

  public static Collection<Object[]> generateTestParametersForConfig(final String[] filePath) {
    Arrays.stream(filePath).forEach(f -> log.info("checking file: {}", f));
    return PARAMS.generate(
        Arrays.stream(filePath)
            .map(f -> Paths.get("src/test/resources/ethereum-tests/" + f).toFile())
            .toList());
  }

  public static Collection<Object[]> generateTestParametersForConfigForFailedTests(
      final String[] filePath, String failedModule, String failedConstraint)
      throws ExecutionException, InterruptedException {
    Arrays.stream(filePath).forEach(f -> log.info("checking file: {}", f));
    Collection<Object[]> params =
        PARAMS.generate(
            Arrays.stream(filePath)
                .map(f -> Paths.get("src/test/resources/ethereum-tests/" + f).toFile())
                .toList());

    return getRecordedFailedTestsFromJson(failedModule, failedConstraint)
        .thenApply(
            failedTests -> {
              List<Object[]> modifiedParams = new ArrayList<>();
              for (Object[] param : params) {
                Object[] modifiedParam = markTestToRun(param, failedTests);
                modifiedParams.add(modifiedParam);
              }
              return modifiedParams;
            })
        .get();
  }

  public static Object[] markTestToRun(Object[] param, Set<String> failedTests) {
    String testName = (String) param[0];
    param[2] = failedTests.contains(testName);

    return param;
  }

  public static void executeTest(final BlockchainReferenceTestCaseSpec spec) {
    final BlockHeader genesisBlockHeader = spec.getGenesisBlockHeader();
    final MutableWorldState worldState =
        spec.getWorldStateArchive()
            .getWorldState(
                WorldStateQueryParams.withStateRootAndBlockHashAndUpdateNodeHead(
                    genesisBlockHeader.getStateRoot(), genesisBlockHeader.getHash()))
            .orElseThrow();

    log.info(
        "checking roothash {} is {}", worldState.rootHash(), genesisBlockHeader.getStateRoot());
    assertThat(worldState.rootHash()).isEqualTo(genesisBlockHeader.getStateRoot());

    final ProtocolSchedule schedule =
        REFERENCE_TEST_PROTOCOL_SCHEDULES.getByName(spec.getNetwork());

    final MutableBlockchain blockchain = spec.getBlockchain();
    final ProtocolContext context = spec.getProtocolContext();

    final BigInteger nonnegativeChainId = schedule.getChainId().get().abs();

    final ZkTracer zkTracer = new ZkTracer(ChainConfig.ETHEREUM);
    zkTracer.traceStartConflation(spec.getCandidateBlocks().length);

    for (var candidateBlock : spec.getCandidateBlocks()) {
      Assumptions.assumeTrue(
          candidateBlock.areAllTransactionsValid(),
          "Skipping the test because the block is not executable");
      Assumptions.assumeTrue(
          candidateBlock.isExecutable(), "Skipping the test because the block is not executable");
      Assumptions.assumeTrue(
          candidateBlock.getBlock().getBody().getTransactions().size() > 0,
          "Skipping the test because the block has no transaction");
      Assumptions.assumeTrue(
          Arrays.stream(spec.getCandidateBlocks()).filter(b -> !b.isValid()).count() == 0,
          "Skipping the test because it has invalid blocks");

      try {
        final Block block = candidateBlock.getBlock();

        zkTracer.traceStartBlock(block.getHeader(), block.getHeader().getCoinbase());

        final ProtocolSpec protocolSpec = schedule.getByBlockHeader(block.getHeader());

        final MainnetBlockImporter blockImporter =
            getMainnetBlockImporter(context, protocolSpec, schedule, zkTracer);

        final HeaderValidationMode validationMode =
            "NoProof".equalsIgnoreCase(spec.getSealEngine())
                ? HeaderValidationMode.LIGHT
                : HeaderValidationMode.FULL;

        // Note: somehow this function is calling traceEndBlock through
        // blockValidator.validateAndProcessBlock
        final BlockImportResult importResult =
            blockImporter.importBlock(context, block, validationMode, validationMode);
        log.info(
            "checking block is imported {} equals {}",
            importResult.isImported(),
            candidateBlock.isValid());
        assertThat(importResult.isImported())
            .isEqualTo(candidateBlock.isValid())
            .withFailMessage(
                "checking block is imported {} while expected {}",
                importResult.isImported(),
                candidateBlock.isValid());

      } catch (final RLPException e) {
        log.info("caught RLP exception, checking it's invalid {}", candidateBlock.isValid());
        assertThat(candidateBlock.isValid()).isFalse();
      }
    }

    zkTracer.traceEndConflation(worldState);

    ExecutionEnvironment.checkTracer(
        zkTracer,
        CORSET_VALIDATOR,
        Optional.of(log),
        // NOTE: just use 0 for start and end block here, since this information is not used.
        0,
        0);
    assertThat(blockchain.getChainHeadHash()).isEqualTo(spec.getLastBlockHash());
  }

  private static MainnetBlockImporter getMainnetBlockImporter(
      final ProtocolContext context,
      final ProtocolSpec protocolSpec,
      final ProtocolSchedule schedule,
      final ZkTracer zkTracer) {
    CorsetBlockProcessor corsetBlockProcessor =
        new CorsetBlockProcessor(
            protocolSpec.getTransactionProcessor(),
            protocolSpec.getTransactionReceiptFactory(),
            protocolSpec.getBlockReward(),
            protocolSpec.getMiningBeneficiaryCalculator(),
            protocolSpec.isSkipZeroBlockRewards(),
            schedule,
            zkTracer);

    MainnetBlockValidator blockValidator =
        new MainnetBlockValidator(
            protocolSpec.getBlockHeaderValidator(),
            protocolSpec.getBlockBodyValidator(),
            corsetBlockProcessor,
            context.getBadBlockManager());

    return new MainnetBlockImporter(blockValidator);
  }
}
