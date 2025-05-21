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

import static org.assertj.core.api.Assertions.assertThat;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.testing.ExecutionEnvironment;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.datatypes.BlobGas;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.mainnet.TransactionValidationParams;
import org.hyperledger.besu.ethereum.processing.TransactionProcessingResult;
import org.hyperledger.besu.ethereum.referencetests.GeneralStateTestCaseEipSpec;
import org.hyperledger.besu.ethereum.referencetests.GeneralStateTestCaseSpec;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestBlockchain;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestProtocolSchedules;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState;
import org.hyperledger.besu.ethereum.rlp.RLP;
import org.hyperledger.besu.ethereum.vm.BlockchainBasedBlockHashLookup;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.testutil.JsonTestParameters;
import org.junit.jupiter.api.Assumptions;

@Slf4j
public class GeneralStateReferenceTestTools {

  private static final ReferenceTestProtocolSchedules REFERENCE_TEST_PROTOCOL_SCHEDULES =
      ReferenceTestProtocolSchedules.create();
  private static final List<String> SPECS_PRIOR_TO_DELETING_EMPTY_ACCOUNTS =
      Arrays.asList("Frontier", "Homestead", "EIP150");

  private static MainnetTransactionProcessor transactionProcessor(final String name) {
    return protocolSpec(name).getTransactionProcessor();
  }

  private static ProtocolSpec protocolSpec(final String name) {
    return REFERENCE_TEST_PROTOCOL_SCHEDULES
        .getByName(name)
        .getByBlockHeader(BlockHeaderBuilder.createDefault().buildBlockHeader());
  }

  private static final List<String> EIPS_TO_RUN;

  static {
    final String eips =
        System.getProperty(
            "test.ethereum.state.eips",
            "Frontier,Homestead,EIP150,EIP158,Byzantium,Constantinople,ConstantinopleFix,Istanbul,Berlin,"
                + "London" /* + ",Merge,Shanghai,Cancun,Prague,Osaka,Bogota"*/);
    EIPS_TO_RUN = Arrays.asList(eips.split(","));
  }

  private static final JsonTestParameters<?, ?> PARAMS =
      JsonTestParameters.create(GeneralStateTestCaseSpec.class, GeneralStateTestCaseEipSpec.class)
          .generator(
              (testName, fullPath, stateSpec, collector) -> {
                final String prefix = testName + "-";
                for (final Map.Entry<String, List<GeneralStateTestCaseEipSpec>> entry :
                    stateSpec.finalStateSpecs().entrySet()) {
                  final String eip = entry.getKey();
                  final boolean runTest =
                      EIPS_TO_RUN.contains(eip) && eip.equalsIgnoreCase("London");
                  final List<GeneralStateTestCaseEipSpec> eipSpecs = entry.getValue();
                  if (eipSpecs.size() == 1) {
                    collector.add(prefix + eip, fullPath, eipSpecs.get(0), runTest);
                  } else {
                    for (int i = 0; i < eipSpecs.size(); i++) {
                      collector.add(
                          prefix + eip + '[' + i + ']', fullPath, eipSpecs.get(i), runTest);
                    }
                  }
                }
              });

  static {
    if (EIPS_TO_RUN.isEmpty()) {
      PARAMS.ignoreAll();
    }

    // ignore tests that are failing in Besu too
    PARAMS.ignore("stCreate2/RevertInCreateInInitCreate2.json");
    PARAMS.ignore("stRevertTest/RevertInCreateInInit.json");
    PARAMS.ignore("stCreate2/create2collisionStorage.json");
    PARAMS.ignore("stExtCodeHash/dynamicAccountOverwriteEmpty.json");

    // ignore tests that are failing because there is an account with nonce 0 and
    // non empty code which can't happen in Linea since we are post LONDON
    PARAMS.ignore("stSStoreTest/InitCollision.json");

    // Consumes a huge amount of memory
    PARAMS.ignore("static_Call1MB1024Calldepth-\\w");
    PARAMS.ignore("ShanghaiLove_.*");
    PARAMS.ignore("VMTests/vmPerformance/");
    PARAMS.ignore("Call50000");
    PARAMS.ignore("static_LoopCallsDepthThenRevert3");
    PARAMS.ignore("Return50000");
    PARAMS.ignore("Callcode50000");

    // Don't do time consuming tests
    PARAMS.ignore("CALLBlake2f_MaxRounds.*");
    PARAMS.ignore("loopMul-.*");

    // Reference Tests are old.  Max blob count is 6.
    PARAMS.ignore("blobhashListBounds5");

    // EOF tests are written against an older version of the spec
    PARAMS.ignore("/stEOF/");

    // Not compliant with the zkEVM requirements.
    PARAMS.ignore("stPreCompiledContracts2/modexpRandomInput.*");
    PARAMS.ignore("tQuadraticComplexityTest/Call50000_ecrec.*");
    PARAMS.ignore("stStaticCall/static_Call50000_ecrec.*");
    PARAMS.ignore("stRandom2/randomStatetest642.*");
    PARAMS.ignore("stRandom2/randomStatetest644.*");
    PARAMS.ignore("stRandom2/randomStatetest645.*");

    // Balance is more than 128 bits
    PARAMS.ignore("stCreate2/CREATE2_Bounds.json");
    PARAMS.ignore("stCreate2/CREATE2_Bounds2.json");
    PARAMS.ignore("stCreate2/CREATE2_Bounds3.json");
    PARAMS.ignore("stCreate2/Create2OnDepth1023.json");
    PARAMS.ignore("stCreate2/Create2OnDepth1024.json");
    PARAMS.ignore("stCreate2/Create2Recursive.json");
    PARAMS.ignore("stDelegatecallTestHomestead/Call1024PreCalls.json");
    PARAMS.ignore("stInitCodeTest/OutOfGasContractCreation.json");
    PARAMS.ignore("stMemoryStressTest/CALLCODE_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/CALLCODE_Bounds2.json");
    PARAMS.ignore("stMemoryStressTest/CALLCODE_Bounds3.json");
    PARAMS.ignore("stMemoryStressTest/CALLCODE_Bounds4.json");
    PARAMS.ignore("stMemoryStressTest/CALL_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/CALL_Bounds2.json");
    PARAMS.ignore("stMemoryStressTest/CALL_Bounds2a.json");
    PARAMS.ignore("stMemoryStressTest/CALL_Bounds3.json");
    PARAMS.ignore("stMemoryStressTest/CREATE_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/CREATE_Bounds2.json");
    PARAMS.ignore("stMemoryStressTest/CREATE_Bounds3.json");
    PARAMS.ignore("stMemoryStressTest/CREATE_Bounds3.json");
    PARAMS.ignore("stMemoryStressTest/DELEGATECALL_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/DELEGATECALL_Bounds2.json");
    PARAMS.ignore("stMemoryStressTest/DELEGATECALL_Bounds3.json");
    PARAMS.ignore("stMemoryStressTest/MSTORE_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/MSTORE_Bounds2.json");
    PARAMS.ignore("stMemoryStressTest/MSTORE_Bounds2a.json");
    PARAMS.ignore("stMemoryStressTest/RETURN_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/static_CALL_Bounds.json");
    PARAMS.ignore("stMemoryStressTest/static_CALL_Bounds2.json");
    PARAMS.ignore("stMemoryStressTest/static_CALL_Bounds2a.json");
    PARAMS.ignore("stMemoryStressTest/static_CALL_Bounds3.json");
    PARAMS.ignore("stStaticCall/static_Call1024PreCalls.json");
    PARAMS.ignore("stStaticCall/static_Call1024PreCalls2.json");
    ;
    PARAMS.ignore("stStaticCall/static_Call1024PreCalls3.json");
    ;
    PARAMS.ignore("stStaticCall/static_RETURN_Bounds.json");
    ;
    PARAMS.ignore("stStaticCall/static_RETURN_BoundsOOG.json");
    PARAMS.ignore("stTransactionTest/HighGasLimit.json");
    PARAMS.ignore("stTransactionTest/OverflowGasRequire2.json");
    PARAMS.ignore("stCallCreateCallCodeTest/Call1024PreCalls.json");

    // Deployment transaction to an account with nonce / code
    PARAMS.ignore("stCreateTest/TransactionCollisionToEmptyButCode.json");
    PARAMS.ignore("stCreateTest/TransactionCollisionToEmptyButNonce.json");
    PARAMS.ignore("stCallCreateCallCodeTest/createJS_ExampleContract.json");
    PARAMS.ignore("stEIP3607/initCollidingWithNonEmptyAccount.json");

    // Deployment transaction to an account with zero nonce, empty code (and zero balance) but
    // nonempty storage. Given [EIP-7610](https://github.com/ethereum/EIPs/pull/8161), no Besu
    // execution takes place, which means that no TraceSection's are created beyond the
    // {@link TxInitializationSection}. This triggers a NPE when tracing, as at some point
    // {@link TraceSection#nextSection} is null in {@link TraceSection#computeContextNumberNew()}.
    PARAMS.ignore("stSpecialTest/FailedCreateRevertsDeletion.json");

    // We ignore the following tests because they satisfy one of the following:
    // - bbs > 512, bbs ≡ base byte size
    // - ebs > 512, ebs ≡ exponent byte size
    // - mbs > 512, mbs ≡ modulus byte size
    PARAMS.ignore("modexp-London\\[12\\]");
    PARAMS.ignore("modexp-London\\[13\\]");
    PARAMS.ignore("modexp-London\\[14\\]");
    PARAMS.ignore("modexp-London\\[15\\]");
    PARAMS.ignore("modexp-London\\[76\\]");
    PARAMS.ignore("modexp-London\\[77\\]");
    PARAMS.ignore("modexp-London\\[78\\]");
    PARAMS.ignore("modexp-London\\[79\\]");
    PARAMS.ignore("modexp-London\\[80\\]");
    PARAMS.ignore("modexp-London\\[81\\]");
    PARAMS.ignore("modexp-London\\[82\\]");
    PARAMS.ignore("modexp-London\\[83\\]");
    PARAMS.ignore("modexp-London\\[84\\]");
    PARAMS.ignore("modexp-London\\[85\\]");
    PARAMS.ignore("modexp-London\\[86\\]");
    PARAMS.ignore("modexp-London\\[87\\]");
    PARAMS.ignore("modexp-London\\[144\\]");
    PARAMS.ignore("modexp-London\\[145\\]");
    PARAMS.ignore("modexp-London\\[146\\]");
    PARAMS.ignore("modexp-London\\[147\\]");
    PARAMS.ignore("modexp-London\\[148\\]");
    PARAMS.ignore("modexp-London\\[149\\]");
    PARAMS.ignore("modexp-London\\[150\\]");
    PARAMS.ignore("modexp-London\\[151\\]");
    PARAMS.ignore("modexp-London\\[151\\]");
    PARAMS.ignore("idPrecomps-London\\[4\\]");
    PARAMS.ignore("modexp_modsize0_returndatasize-London\\[4\\]");
    PARAMS.ignore("stRandom2/randomStatetest650.json");
  }

  private GeneralStateReferenceTestTools() {
    // utility class
  }

  public static Collection<Object[]> generateTestParametersForConfig(final String[] filePath) {
    return PARAMS.generate(filePath);
  }

  @SneakyThrows
  public static void executeTest(final GeneralStateTestCaseEipSpec spec) {
    final BlockHeader blockHeader = spec.getBlockHeader();
    final ReferenceTestWorldState initialWorldState = spec.getInitialWorldState();
    final Transaction transaction = spec.getTransaction(0);
    Assumptions.assumeTrue(
        transaction != null, "Skipping the test because the block has no transaction");
    final BlockBody blockBody = new BlockBody(List.of(transaction), new ArrayList<>());

    // Sometimes the tests ask us assemble an invalid transaction.  If we have
    // no valid transaction then there is no test.  GeneralBlockChain tests
    // will handle the case where we receive the TXs in a serialized form.
    if (transaction == null) {
      assertThat(spec.getExpectException())
          .withFailMessage("Transaction was not assembled, but no exception was expected")
          .isNotNull();
      return;
    }

    final MutableWorldState worldState = initialWorldState.copy();
    // Several of the GeneralStateTests check if the transaction could potentially
    // consume more gas than is left for the block it's attempted to be included in.
    // This check is performed within the `BlockImporter` rather than inside the
    // `TransactionProcessor`, so these tests are skipped.
    if (transaction.getGasLimit() > blockHeader.getGasLimit() - blockHeader.getGasUsed()) {
      return;
    }

    final MainnetTransactionProcessor processor = transactionProcessor(spec.getFork());
    final WorldUpdater worldStateUpdater = worldState.updater();
    final ReferenceTestBlockchain blockchain = new ReferenceTestBlockchain(blockHeader.getNumber());
    final Wei blobGasPrice =
        protocolSpec(spec.getFork())
            .getFeeMarket()
            .blobGasPricePerGas(blockHeader.getExcessBlobGas().orElse(BlobGas.ZERO));

    final ChainConfig chain = ChainConfig.ETHEREUM_CHAIN(spec.getFork());
    final CorsetValidator corsetValidator = new CorsetValidator(chain);
    final ZkTracer zkTracer = new ZkTracer(chain);
    zkTracer.traceStartConflation(1);
    zkTracer.traceStartBlock(blockHeader, blockBody, blockHeader.getCoinbase());

    final TransactionProcessingResult result =
        processor.processTransaction(
            worldStateUpdater,
            blockHeader,
            transaction,
            blockHeader.getCoinbase(),
            zkTracer,
            new BlockchainBasedBlockHashLookup(blockHeader, blockchain),
            false,
            TransactionValidationParams.processingBlock(),
            blobGasPrice);
    if (result.isInvalid()) {
      assertThat(spec.getExpectException())
          .withFailMessage(() -> result.getValidationResult().getErrorMessage())
          .isNotNull();
      return;
    }

    zkTracer.traceEndBlock(blockHeader, blockBody);
    zkTracer.traceEndConflation(worldStateUpdater);

    assertThat(spec.getExpectException())
        .withFailMessage("Exception was expected - " + spec.getExpectException())
        .isNull();

    final Account coinbase = worldStateUpdater.getOrCreate(spec.getBlockHeader().getCoinbase());
    if (coinbase != null && coinbase.isEmpty() && shouldClearEmptyAccounts(spec.getFork())) {
      worldStateUpdater.deleteAccount(coinbase.getAddress());
    }
    worldStateUpdater.commit();
    worldState.persist(blockHeader);

    // Check the world state root hash.
    final Hash expectedRootHash = spec.getExpectedRootHash();
    assertThat(worldState.rootHash())
        .withFailMessage(
            "Unexpected world state root hash; expected state: %s, computed state: %s",
            spec.getExpectedRootHash(), worldState.rootHash())
        .isEqualTo(expectedRootHash);

    // Check the logs.
    final Hash expectedLogsHash = spec.getExpectedLogsHash();
    Optional.ofNullable(expectedLogsHash)
        .ifPresent(
            expected -> {
              final List<Log> logs = result.getLogs();

              assertThat(Hash.hash(RLP.encode(out -> out.writeList(logs, Log::writeTo))))
                  .withFailMessage("Unmatched logs hash. Generated logs: %s", logs)
                  .isEqualTo(expected);
            });

    ExecutionEnvironment.checkTracer(
        zkTracer,
        corsetValidator,
        Optional.of(log),
        // block number for first block
        blockHeader.getNumber(),
        // block number for last block
        blockHeader.getNumber());
  }

  private static boolean shouldClearEmptyAccounts(final String eip) {
    return !SPECS_PRIOR_TO_DELETING_EMPTY_ACCOUNTS.contains(eip);
  }
}
