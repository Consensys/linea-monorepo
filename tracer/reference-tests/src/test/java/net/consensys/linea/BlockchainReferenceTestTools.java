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
import static net.consensys.linea.reporting.TracerTestBase.getForkOrDefault;
import static net.consensys.linea.testing.ToyExecutionTools.addSystemAccountsIfRequired;
import static org.assertj.core.api.Assertions.assertThat;

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
import net.consensys.linea.zktracer.Fork;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.ethereum.BlockValidator;
import org.hyperledger.besu.ethereum.MainnetBlockValidatorBuilder;
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
import org.hyperledger.besu.ethereum.trie.pathbased.common.provider.WorldStateQueryParams;
import org.hyperledger.besu.testutil.JsonTestParameters;
import org.junit.jupiter.api.Assumptions;

@Slf4j
public class BlockchainReferenceTestTools {
  // Keep the forkName and the zkevm_fork in github worklow in PascalCase
  private static final String forkName = getForkOrDefault("London");
  private static final ReferenceTestProtocolSchedules REFERENCE_TEST_PROTOCOL_SCHEDULES =
      ReferenceTestProtocolSchedules.create();

  private static final List<String> NETWORKS_TO_RUN = List.of(forkName);

  public static final JsonTestParameters<?, ?> PARAMS =
      JsonTestParameters.create(BlockchainReferenceTestCaseSpec.class)
          .generator(
              (testName, fullPath, spec, collector) -> {
                final String eip = spec.getNetwork();
                collector.add(
                    testName + "[" + eip + "]", fullPath, spec, NETWORKS_TO_RUN.contains(eip));
              });

  static {
    if (NETWORKS_TO_RUN.isEmpty()) {
      PARAMS.ignoreAll();
    }
    // ignore tests that are failing in Besu too
    PARAMS.ignore("RevertInCreateInInitCreate2_d0g0v0_*");
    PARAMS.ignore("RevertInCreateInInit_d0g0v0_*");
    PARAMS.ignore("create2collisionStorage_d0g0v0_*");
    PARAMS.ignore("create2collisionStorage_d1g0v0_*");
    PARAMS.ignore("create2collisionStorage_d2g0v0_*");
    PARAMS.ignore("dynamicAccountOverwriteEmpty_d0g0v0_*");

    // ignore tests that are failing because there is an account with nonce 0 and
    // non-empty code which can't happen in Linea since we are post LONDON
    PARAMS.ignore("InitCollision_d0g0v0_*");
    PARAMS.ignore("InitCollision_d1g0v0_*");
    PARAMS.ignore("InitCollision_d2g0v0_*");
    PARAMS.ignore("InitCollision_d3g0v0_*");
    PARAMS.ignore("RevertInCreateInInitCreate2_d0g0v0_London\\[London\\]");
    PARAMS.ignore("RevertInCreateInInit_d0g0v0_London\\[London\\]");

    // Arithmetization restriction: recipient address is a precompile.
    PARAMS.ignore("modexpRandomInput_d0g0v0_*");
    PARAMS.ignore("modexpRandomInput_d0g1v0_*");
    PARAMS.ignore("modexpRandomInput_d1g0v0_*");
    PARAMS.ignore("modexpRandomInput_d1g1v0_*");
    PARAMS.ignore("modexpRandomInput_d2g0v0_*");
    PARAMS.ignore("modexpRandomInput_d2g1v0_*");
    PARAMS.ignore("randomStatetest642_d0g0v0_*");
    PARAMS.ignore("randomStatetest644_d0g0v0_*");
    PARAMS.ignore("randomStatetest645_d0g0v0_*");
    PARAMS.ignore("randomStatetest645_d0g0v1_*");

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
    PARAMS.ignore("CALLCODE_Bounds2_d0g0v0_*");
    PARAMS.ignore("CALLCODE_Bounds2_d0g1v0_*");
    PARAMS.ignore("CALLCODE_Bounds3_d0g0v0_*");
    PARAMS.ignore("CALLCODE_Bounds3_d0g1v0_*");
    PARAMS.ignore("CALLCODE_Bounds4_d0g0v0_*");
    PARAMS.ignore("CALLCODE_Bounds4_d0g1v0_*");
    PARAMS.ignore("CALLCODE_Bounds4_d0g2v0_*");
    PARAMS.ignore("CALLCODE_Bounds_d0g0v0_*");
    PARAMS.ignore("CALLCODE_Bounds_d0g1v0_*");
    PARAMS.ignore("CALL_Bounds2_d0g0v0_*");
    PARAMS.ignore("CALL_Bounds2_d0g1v0_*");
    PARAMS.ignore("CALL_Bounds2a_d0g0v0_*");
    PARAMS.ignore("CALL_Bounds2a_d0g1v0_*");
    PARAMS.ignore("CALL_Bounds3_d0g0v0_*");
    PARAMS.ignore("CALL_Bounds3_d0g1v0_*");
    PARAMS.ignore("CALL_Bounds3_d0g2v0_*");
    PARAMS.ignore("CALL_Bounds_d0g0v0_*");
    PARAMS.ignore("CALL_Bounds_d0g1v0_*");
    PARAMS.ignore("CREATE2_Bounds2_d0g0v0_*");
    PARAMS.ignore("CREATE2_Bounds2_d0g1v0_*");
    PARAMS.ignore("CREATE2_Bounds3_d0g0v0_*");
    PARAMS.ignore("CREATE2_Bounds3_d0g1v0_*");
    PARAMS.ignore("CREATE2_Bounds3_d0g2v0_*");
    PARAMS.ignore("CREATE2_Bounds_d0g0v0_*");
    PARAMS.ignore("CREATE2_Bounds_d0g1v0_*");
    PARAMS.ignore("CREATE_Bounds2_d0g0v0_*");
    PARAMS.ignore("CREATE_Bounds2_d0g1v0_*");
    PARAMS.ignore("CREATE_Bounds3_d0g0v0_*");
    PARAMS.ignore("CREATE_Bounds3_d0g1v0_*");
    PARAMS.ignore("CREATE_Bounds3_d0g2v0_*");
    PARAMS.ignore("CREATE_Bounds_d0g0v0_*");
    PARAMS.ignore("CREATE_Bounds_d0g1v0_*");
    PARAMS.ignore("Call1024PreCalls_d0g0v0_*");
    PARAMS.ignore("Call1024PreCalls_d0g1v0_*");
    PARAMS.ignore("Call1024PreCalls_d0g2v0_*");
    PARAMS.ignore("Create2OnDepth1023_d0g0v0_*");
    PARAMS.ignore("Create2OnDepth1024_d0g0v0_*");
    PARAMS.ignore("Create2Recursive_d0g0v0_*");
    PARAMS.ignore("Create2Recursive_d0g1v0_*");
    PARAMS.ignore("Create2Recursive_d0g2v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds2_d0g0v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds2_d0g1v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds3_d0g0v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds3_d0g1v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds3_d0g2v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds_d0g0v0_*");
    PARAMS.ignore("DELEGATECALL_Bounds_d0g1v0_*");
    PARAMS.ignore("DelegateCallSpam_*");
    PARAMS.ignore("HighGasLimit_d0g0v0_*");
    PARAMS.ignore("MSTORE_Bounds2_d0g0v0_*");
    PARAMS.ignore("MSTORE_Bounds2_d0g1v0_*");
    PARAMS.ignore("MSTORE_Bounds2a_d0g0v0_*");
    PARAMS.ignore("MSTORE_Bounds2a_d0g1v0_*");
    PARAMS.ignore("MSTORE_Bounds_d0g0v0_*");
    PARAMS.ignore("MSTORE_Bounds_d0g1v0_*");
    PARAMS.ignore("OutOfGasContractCreation_d0g0v0_*");
    PARAMS.ignore("OutOfGasContractCreation_d0g1v0_*");
    PARAMS.ignore("OutOfGasContractCreation_d1g0v0_*");
    PARAMS.ignore("OutOfGasContractCreation_d1g1v0_*");
    PARAMS.ignore("OverflowGasRequire2_d0g0v0_*");
    PARAMS.ignore("OverflowGasRequire_*");
    PARAMS.ignore("RETURN_Bounds_d0g0v0_*");
    PARAMS.ignore("RETURN_Bounds_d0g1v0_*");
    PARAMS.ignore("RETURN_Bounds_d0g2v0_*");
    PARAMS.ignore("StrangeContractCreation_*");
    PARAMS.ignore("SuicideIssue_*");
    PARAMS.ignore("static_CALL_Bounds2_d0g0v0_*");
    PARAMS.ignore("static_CALL_Bounds2_d0g1v0_*");
    PARAMS.ignore("static_CALL_Bounds2a_d0g0v0_*");
    PARAMS.ignore("static_CALL_Bounds2a_d0g1v0_*");
    PARAMS.ignore("static_CALL_Bounds3_d0g0v0_*");
    PARAMS.ignore("static_CALL_Bounds3_d0g1v0_*");
    PARAMS.ignore("static_CALL_Bounds_d0g0v0_*");
    PARAMS.ignore("static_CALL_Bounds_d0g1v0_*");
    PARAMS.ignore("static_Call1024PreCalls2_d0g0v0_*");
    PARAMS.ignore("static_Call1024PreCalls2_d1g0v0_*");
    PARAMS.ignore("static_Call1024PreCalls3_d0g0v0_*");
    PARAMS.ignore("static_Call1024PreCalls3_d1g0v0_*");
    PARAMS.ignore("static_Call1024PreCalls_d1g0v0_*");
    PARAMS.ignore("static_RETURN_BoundsOOG_d0g0v0_*");
    PARAMS.ignore("static_RETURN_BoundsOOG_d1g0v0_*");
    PARAMS.ignore("static_RETURN_Bounds_d0g0v0_London\\[London\\]");
    PARAMS.ignore("Cancun-enough_gas*");
    PARAMS.ignore("Cancun-out_of_gas*");
    PARAMS.ignore("Cancun-no_stack_overflow*");
    PARAMS.ignore("Cancun-stack_overflow*");
    PARAMS.ignore("Cancun-zero_inputs*");
    PARAMS.ignore("Cancun-zero_length_out_of_bounds_destination*");
    PARAMS.ignore("Cancun-single_byte_rewrite*");
    PARAMS.ignore("Cancun-full_word_rewrite*");
    PARAMS.ignore("Cancun-single_byte_forward_overwrite*");
    PARAMS.ignore("Cancun-full_word_forward_overwrite*");
    PARAMS.ignore("Cancun-mid_word_single_byte_rewrite*");
    PARAMS.ignore("Cancun-mid_word_single_word_rewrite*");
    PARAMS.ignore("Cancun-mid_word_multi_word_rewrite*");
    PARAMS.ignore("Cancun-two_words_forward_overwrite*");
    PARAMS.ignore("Cancun-two_words_backward_overwrite*");
    PARAMS.ignore("Cancun-two_words_backward_overwrite_single_byte_offset*");
    PARAMS.ignore("Cancun-single_byte_memory_extension*");
    PARAMS.ignore("Cancun-single_word_memory_extension*");
    PARAMS.ignore("Cancun-single_word_minus_one_byte_memory_extension*");
    PARAMS.ignore("Cancun-single_word_plus_one_byte_memory_extension*");
    PARAMS.ignore("Cancun-full_memory_rewrite*");
    PARAMS.ignore("Cancun-full_memory_copy*");
    PARAMS.ignore("Cancun-full_memory_copy_offset*");
    PARAMS.ignore("Cancun-full_memory_clean*");
    PARAMS.ignore("Cancun-empty_memory-length=0-src=0-dest=0*");
    PARAMS.ignore("Cancun-empty_memory-length=0-src=0-dest=32*");
    PARAMS.ignore("Cancun-empty_memory-length=0-src=32-dest=0*");
    PARAMS.ignore("Cancun-empty_memory-length=0-src=32-dest=32*");
    PARAMS.ignore("Cancun-empty_memory-length=1-src=0-dest=0*");
    PARAMS.ignore("Cancun-empty_memory-length=1-src=0-dest=32*");
    PARAMS.ignore("Cancun-empty_memory-length=1-src=32-dest=0*");
    PARAMS.ignore("Cancun-empty_memory-length=1-src=32-dest=32*");
    PARAMS.ignore("Cancun-call");
    PARAMS.ignore("Cancun-staticcall_cant_call_tstore");
    PARAMS.ignore("Cancun-staticcall_cant_call_tstore_with_stack_underflow");
    PARAMS.ignore("Cancun-staticcalled_can_call_tstore");
    PARAMS.ignore("Cancun-staticcalled_context_can_call_tload");
    PARAMS.ignore("Cancun-callcode");
    PARAMS.ignore("Cancun-delegatecall");
    PARAMS.ignore("Cancun-call_with_revert");
    PARAMS.ignore("Cancun-call_with_invalid");
    PARAMS.ignore("Cancun-call_with_stack_underflow");
    PARAMS.ignore("Cancun-call_with_tstore_stack_underflow");
    PARAMS.ignore("Cancun-call_with_tstore_stack_underflow_2");
    PARAMS.ignore("Cancun-call_with_tload_stack_underflow");
    PARAMS.ignore("Cancun-call_with_out_of_gas");
    PARAMS.ignore("Cancun-call_with_out_of_gas_2");
    PARAMS.ignore("Cancun-callcode_with_revert");
    PARAMS.ignore("Cancun-callcode_with_invalid");
    PARAMS.ignore("Cancun-callcode_with_stack_underflow");
    PARAMS.ignore("Cancun-callcode_with_tstore_stack_underflow");
    PARAMS.ignore("Cancun-callcode_with_tstore_stack_underflow_2");
    PARAMS.ignore("Cancun-callcode_with_tload_stack_underflow");
    PARAMS.ignore("Cancun-callcode_with_out_of_gas");
    PARAMS.ignore("Cancun-callcode_with_out_of_gas_2");
    PARAMS.ignore("Cancun-delegatecall_with_revert");
    PARAMS.ignore("Cancun-delegatecall_with_invalid");
    PARAMS.ignore("Cancun-delegatecall_with_stack_underflow");
    PARAMS.ignore("Cancun-delegatecall_with_tstore_stack_underflow");
    PARAMS.ignore("Cancun-delegatecall_with_tstore_stack_underflow_2");
    PARAMS.ignore("Cancun-delegatecall_with_tload_stack_underflow");
    PARAMS.ignore("Cancun-delegatecall_with_out_of_gas");
    PARAMS.ignore("Cancun-delegatecall_with_out_of_gas_2");
    PARAMS.ignore("Cancun-tstore_in_reentrant_call");
    PARAMS.ignore("Cancun-tload_after_reentrant_tstore");
    PARAMS.ignore("Cancun-manipulate_in_reentrant_call");
    PARAMS.ignore("Cancun-tstore_in_call_then_tload_return_in_staticcall");
    PARAMS.ignore("Cancun-tstore_before_revert_has_no_effect");
    PARAMS.ignore("Cancun-revert_undoes_all");
    PARAMS.ignore("Cancun-revert_undoes_tstorage_after_successful_call");
    PARAMS.ignore("Cancun-tstore_before_invalid_has_no_effect");
    PARAMS.ignore("Cancun-revert_undoes_all");
    PARAMS.ignore("Cancun-invalid_undoes_all");
    PARAMS.ignore("Cancun-invalid_undoes_tstorage_after_successful_call");
    PARAMS.ignore("Cancun-tload_after_selfdestruct_pre_existing_contract");
    PARAMS.ignore("Cancun-tload_after_selfdestruct_new_contract");
    PARAMS.ignore("Cancun-tload_after_inner_selfdestruct_pre_existing_contract");
    PARAMS.ignore("Cancun-tload_after_inner_selfdestruct_new_contract");
    PARAMS.ignore("Cancun-tstore_after_selfdestruct_pre_existing_contract");
    PARAMS.ignore("Cancun-tstore_after_selfdestruct_new_contract");
    PARAMS.ignore("Cancun-out_of_bounds_memory_extension*");
    PARAMS.ignore("Cancun-opcode=CALL");
    PARAMS.ignore("Cancun-opcode=DELEGATECALL");
    PARAMS.ignore("Cancun-opcode=STATICCALL");
    PARAMS.ignore("Cancun-opcode=CALLCODE");
    PARAMS.ignore("Cancun-opcode=CREATE");
    PARAMS.ignore("Cancun-opcode=CREATE2");

    // Deployment transaction to an account with nonce / code
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g0v0_*");
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g0v1_*");
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g1v0_*");
    PARAMS.ignore("TransactionCollisionToEmptyButCode_d0g1v1_*");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g0v0_*");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g0v1_*");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g1v0_*");
    PARAMS.ignore("TransactionCollisionToEmptyButNonce_d0g1v1_*");
    PARAMS.ignore("createJS_ExampleContract_d0g0v0_*");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d0g0v0_*");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d1g0v0_*");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d2g0v0_*");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d3g0v0_*");
    PARAMS.ignore("initCollidingWithNonEmptyAccount_d4g0v0_*");

    // Deployment transaction to an account with zero nonce, empty code (and zero balance) but
    // nonempty storage. Given [EIP-7610](https://github.com/ethereum/EIPs/pull/8161), no Besu
    // execution takes place, which means that no TraceSection's are created beyond the
    // {@link TxInitializationSection}. This triggers a NPE when tracing, as at some point
    // {@link TraceSection#nextSection} is null in {@link TraceSection#computeContextNumberNew()}.
    PARAMS.ignore("FailedCreateRevertsDeletion_d0g0v0_*");

    // Ignore the following test as it is not supported in Linea.
    // See [issue #1678](https://github.com/Consensys/linea-tracer/issues/1678)
    PARAMS.ignore("suicideStorageCheck_*");

    // Don't do time-consuming tests.
    PARAMS.ignore("CALLBlake2f_MaxRounds.*");
    PARAMS.ignore("loopMul_*");
    PARAMS.ignore("randomStatetest177_d0g0v0_*");
    PARAMS.ignore("15_tstoreCannotBeDosd_d0g0v0*");
    PARAMS.ignore("21_tstoreCannotBeDosdOOO_d0g0v0*");

    // Inconclusive fork choice rule, since in merge CL should be choosing forks and setting the
    // chain head. Perfectly valid test pre-merge.
    PARAMS.ignore("UncleFromSideChain_(Cancun|Prague|Osaka|Bogota)");

    // EOF tests are written against an older version of the spec.
    PARAMS.ignore("/stEOF/");

    // We ignore the following tests because they satisfy one of the following:
    // - bbs > 512, bbs ≡ base byte size
    // - ebs > 512, ebs ≡ exponent byte size
    // - mbs > 512, mbs ≡ modulus byte size
    PARAMS.ignore("modexp_d28g0v0_*");
    PARAMS.ignore("modexp_d28g1v0_*");
    PARAMS.ignore("modexp_d28g2v0_*");
    PARAMS.ignore("modexp_d28g3v0_*");
    PARAMS.ignore("modexp_d29g0v0_*");
    PARAMS.ignore("modexp_d29g1v0_*");
    PARAMS.ignore("modexp_d29g2v0_*");
    PARAMS.ignore("modexp_d29g3v0_*");
    PARAMS.ignore("modexp_d2g0v0_*");
    PARAMS.ignore("modexp_d2g1v0_*");
    PARAMS.ignore("modexp_d2g2v0_*");
    PARAMS.ignore("modexp_d2g3v0_*");
    PARAMS.ignore("modexp_d30g0v0_*");
    PARAMS.ignore("modexp_d30g1v0_*");
    PARAMS.ignore("modexp_d30g2v0_*");
    PARAMS.ignore("modexp_d30g3v0_*");
    PARAMS.ignore("modexp_d36g0v0_*");
    PARAMS.ignore("modexp_d36g1v0_*");
    PARAMS.ignore("modexp_d36g2v0_*");
    PARAMS.ignore("modexp_d36g3v0_*");
    PARAMS.ignore("modexp_d37g0v0_*");
    PARAMS.ignore("modexp_d37g1v0_*");
    PARAMS.ignore("modexp_d37g2v0_*");
    PARAMS.ignore("modexp_d37g3v0_*");
    PARAMS.ignore("idPrecomps_d4g0v0_*");
    PARAMS.ignore("modexp_modsize0_returndatasize_d4g0v0_*");
    PARAMS.ignore("randomStatetest650_d0g0v0_*");

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

    // Tests have a root hash mismatch
    // They have been removed from legacy ethereum tests repo
    // - all the other ecmul tests for point 1,3 factor 0 are run for Byzantium and
    // Contantinople+Fix
    // - in state tests, they run on the 3 forks above only
    // - coinbase is in pre and not post and has no balance
    PARAMS.ignore("ecmul_1-3_0_28000_80_d0g0v0_*");
    PARAMS.ignore("ecmul_1-3_0_28000_80_d0g1v0_*");
    PARAMS.ignore("ecmul_1-3_0_28000_80_d0g2v0_*");
    PARAMS.ignore("ecmul_1-3_0_28000_80_d0g3v0_*");
    // - all the other ecmul tests for point 0,3 factor
    // 21888242871839275222246405745257275088548364400416034343698204186575808495616 have coinbase
    // pre and post with balance
    // - coinbase is in pre and not post and has no balance
    PARAMS.ignore("ecmul_0-3_5616_28000_96_d0g0v0_*");
    PARAMS.ignore("ecmul_0-3_5616_28000_96_d0g1v0_*");
    PARAMS.ignore("ecmul_0-3_5616_28000_96_d0g2v0_*");
    // - all the other ecadd tests for points (0,0) and (0,0) have coinbase pre and post with
    // balance
    // - coinbase is in pre and not post and has no balance
    PARAMS.ignore("ecadd_0-0_0-0_21000_80_d0g0v0_*");
    PARAMS.ignore("ecadd_0-0_0-0_21000_80_d0g1v0_*");
    PARAMS.ignore("ecadd_0-0_0-0_21000_80_d0g2v0_*");
    PARAMS.ignore("ecadd_0-0_0-0_21000_80_d0g3v0_*");
    // - all the other ecadd tests for points (1,3) and (0,0) have coinbase pre and post with
    // balance
    // - coinbase is in pre and not post and has no balance
    PARAMS.ignore("ecadd_1-3_0-0_25000_80_d0g0v0_*");
    PARAMS.ignore("ecadd_1-3_0-0_25000_80_d0g1v0_*");
    PARAMS.ignore("ecadd_1-3_0-0_25000_80_d0g2v0_*");
    PARAMS.ignore("ecadd_1-3_0-0_25000_80_d0g3v0_*");

    // System transactions Withdrawals are not supported
    // Breaks hub.account-consistency---linking---conflation-level---balance as the transition
    // for account 0x0000000000000000000000000000000000000200
    PARAMS.ignore("BlockchainTests/Pyspecs/shanghai/eip4895_withdrawals/balance_within_block.json");
    PARAMS.ignore(
        "BlockchainTests/Pyspecs/shanghai/eip4895_withdrawals/use_value_in_contract.json");
    // for account EIP4788_BEACONROOT_ADDRESS
    PARAMS.ignore("Cancun-block_count=10-buffer_wraparound");
    PARAMS.ignore("Cancun-block_count=10-buffer_wraparound_overwrite");
    PARAMS.ignore("Cancun-block_count=10-buffer_wraparound_overwrite_high_timestamp");
    PARAMS.ignore("Cancun-block_count=10-buffer_wraparound_no_overwrite");
    PARAMS.ignore("Cancun-block_count=10-buffer_wraparound_no_overwrite_2");

    // Pending deployment number fix
    // Issue #https://github.com/Consensys/linea-specification/issues/191
    PARAMS.ignore("create2collisionwithSelfdestructSameBlock.json");

    // Transaction Type not supported at the moment
    PARAMS.ignore("opcodeBlobhBounds*");
    PARAMS.ignore("opcodeBlobhashOutOfRange*");
    PARAMS.ignore("blockWithAllTransactionTypes*");
    PARAMS.ignore("Cancun-tx_type=3*");
    PARAMS.ignore("blobhashListBounds3_d0g0v0_*");
    PARAMS.ignore("blobhashListBounds4_d0g0v0_*");
    PARAMS.ignore("blobhashListBounds5_d0g0v0_*");
    PARAMS.ignore("blobhashListBounds6_d0g0v0_*");
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
    final Fork fork = getForkFromNetwork(spec.getNetwork());
    final ChainConfig chain = ChainConfig.ETHEREUM_CHAIN(fork);
    final MutableBlockchain blockchain = spec.getBlockchain();
    final ProtocolContext context = spec.getProtocolContext();

    // Add system accounts if the fork requires it.
    addSystemAccountsIfRequired(worldState.updater(), chain.fork);

    final CorsetValidator corsetValidator = new CorsetValidator(chain);
    final ZkTracer zkTracer = new ZkTracer(chain);
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

        zkTracer.traceStartBlock(worldState, block.getHeader(), block.getHeader().getCoinbase());

        final ProtocolSpec protocolSpec = schedule.getByBlockHeader(block.getHeader());

        final MainnetBlockImporter blockImporter =
            getMainnetBlockImporter(protocolSpec, schedule, zkTracer);

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
        corsetValidator,
        Optional.of(log),
        // NOTE: just use 0 for start and end block here, since this information is not used.
        0,
        0,
        null);
    assertThat(blockchain.getChainHeadHash()).isEqualTo(spec.getLastBlockHash());
  }

  private static MainnetBlockImporter getMainnetBlockImporter(
      final ProtocolSpec protocolSpec, final ProtocolSchedule schedule, final ZkTracer zkTracer) {
    final CorsetBlockProcessor corsetBlockProcessor =
        new CorsetBlockProcessor(
            protocolSpec.getTransactionProcessor(),
            protocolSpec.getTransactionReceiptFactory(),
            protocolSpec.getBlockReward(),
            protocolSpec.getMiningBeneficiaryCalculator(),
            protocolSpec.isSkipZeroBlockRewards(),
            schedule,
            zkTracer);

    final BlockValidator blockValidator =
        MainnetBlockValidatorBuilder.frontier(
            protocolSpec.getBlockHeaderValidator(),
            protocolSpec.getBlockBodyValidator(),
            corsetBlockProcessor);

    return new MainnetBlockImporter(blockValidator);
  }

  private static Fork getForkFromNetwork(String string) {
    if (string.equals("Merge")) {
      return Fork.PARIS;
    }
    return Fork.fromString(string);
  }
}
