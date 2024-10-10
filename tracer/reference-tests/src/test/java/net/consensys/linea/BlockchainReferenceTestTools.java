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
import static net.consensys.linea.ReferenceTestOutcomeRecorderTool.getModule;
import static net.consensys.linea.ReferenceTestWatcher.JSON_INPUT_FILENAME;
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
import java.util.concurrent.ExecutionException;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.testing.ExecutionEnvironment;
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
import org.hyperledger.besu.testutil.JsonTestParameters;

@Slf4j
public class BlockchainReferenceTestTools {
  private static final ReferenceTestProtocolSchedules REFERENCE_TEST_PROTOCOL_SCHEDULES =
      ReferenceTestProtocolSchedules.create();

  private static final List<String> NETWORKS_TO_RUN = List.of("London");

  private static final JsonTestParameters<?, ?> PARAMS =
      JsonTestParameters.create(BlockchainReferenceTestCaseSpec.class)
          .generator(
              (testName, fullPath, spec, collector) -> {
                final String eip = spec.getNetwork();
                collector.add(
                    testName + "[" + eip + "]", fullPath, spec, NETWORKS_TO_RUN.contains(eip));
              });

  private static final CorsetValidator CORSET_VALIDATOR = new CorsetValidator();

  static {
    if (NETWORKS_TO_RUN.isEmpty()) {
      PARAMS.ignoreAll();
    }

    // Consumes a huge amount of memory.
    PARAMS.ignore("static_Call1MB1024Calldepth_d1g0v0_\\w+");
    PARAMS.ignore("ShanghaiLove_.*");
    PARAMS.ignore("/GeneralStateTests/VMTests/vmPerformance/");

    // Absurd amount of gas, doesn't run in parallel.
    PARAMS.ignore("randomStatetest94_\\w+");

    // Don't do time-consuming tests.
    PARAMS.ignore("CALLBlake2f_MaxRounds.*");
    PARAMS.ignore("loopMul_*");

    // Inconclusive fork choice rule, since in merge CL should be choosing forks and setting the
    // chain head.
    // Perfectly valid test pre-merge.
    PARAMS.ignore("UncleFromSideChain_(Merge|Shanghai|Cancun|Prague|Osaka|Bogota)");

    // EOF tests are written against an older version of the spec.
    PARAMS.ignore("/stEOF/");
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

    CompletableFuture<BlockchainReferenceTestOutcome> modulesToConstraintsFutures =
        readBlockchainReferenceTestsOutput(JSON_INPUT_FILENAME)
            .thenApply(ReferenceTestOutcomeRecorderTool::getBlockchainReferenceTestOutcome);

    return modulesToConstraintsFutures.thenApply(
        blockchainReferenceTestOutcome -> {
          ModuleToConstraints filteredFailedTests =
              getModule(blockchainReferenceTestOutcome.modulesToConstraints(), failedModule);
          if (filteredFailedTests == null) {
            return failedTests;
          }
          if (!failedConstraint.isEmpty()) {
            return filteredFailedTests.getFailedTests(failedConstraint);
          }
          return filteredFailedTests.getFailedTests();
        });
  }

  public static Collection<Object[]> generateTestParametersForConfig(final String[] filePath) {
    Arrays.stream(filePath).forEach(f -> log.info("checking file: {}", f));
    return PARAMS.generate(
        Arrays.stream(filePath)
            .map(f -> Paths.get("src/test/resources/ethereum-tests/" + f).toFile())
            .toList());
  }

  public static Collection<Object[]> generateTestParametersForConfig(
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
            .getMutable(genesisBlockHeader.getStateRoot(), genesisBlockHeader.getHash())
            .get();
    log.info(
        "checking roothash {} is {}", worldState.rootHash(), genesisBlockHeader.getStateRoot());
    assertThat(worldState.rootHash()).isEqualTo(genesisBlockHeader.getStateRoot());

    final ProtocolSchedule schedule =
        REFERENCE_TEST_PROTOCOL_SCHEDULES.getByName(spec.getNetwork());

    final MutableBlockchain blockchain = spec.getBlockchain();
    final ProtocolContext context = spec.getProtocolContext();

    final BigInteger nonnegativeChainId = schedule.getChainId().get().abs();

    final ZkTracer zkTracer = new ZkTracer(nonnegativeChainId);
    zkTracer.traceStartConflation(spec.getCandidateBlocks().length);

    for (var candidateBlock : spec.getCandidateBlocks()) {
      if (!candidateBlock.isExecutable()
          || candidateBlock.getBlock().getBody().getTransactions().isEmpty()) {
        return;
      }

      try {
        final Block block = candidateBlock.getBlock();

        zkTracer.traceStartBlock(block.getHeader());

        final ProtocolSpec protocolSpec = schedule.getByBlockHeader(block.getHeader());

        final MainnetBlockImporter blockImporter =
            getMainnetBlockImporter(context, protocolSpec, schedule, zkTracer);

        final HeaderValidationMode validationMode =
            "NoProof".equalsIgnoreCase(spec.getSealEngine())
                ? HeaderValidationMode.LIGHT
                : HeaderValidationMode.FULL;

        final BlockImportResult importResult =
            blockImporter.importBlock(context, block, validationMode, validationMode);
        log.info(
            "checking block is imported {} equals {}",
            importResult.isImported(),
            candidateBlock.isValid());
        assertThat(importResult.isImported()).isEqualTo(candidateBlock.isValid());

        zkTracer.traceEndBlock(block.getHeader(), block.getBody());
      } catch (final RLPException e) {
        log.info("caught RLP exception, checking it's invalid {}", candidateBlock.isValid());
        assertThat(candidateBlock.isValid()).isFalse();
      }
    }

    zkTracer.traceEndConflation(worldState);

    ExecutionEnvironment.checkTracer(zkTracer, CORSET_VALIDATOR, Optional.of(log));

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
