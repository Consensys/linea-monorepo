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

package net.consensys.linea.testing;

import static net.consensys.linea.zktracer.Trace.LINEA_BLOCK_GAS_LIMIT;
import static org.assertj.core.api.Assertions.assertThat;

import java.util.*;
import java.util.function.Consumer;

import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.BlobGas;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.mainnet.TransactionValidationParams;
import org.hyperledger.besu.ethereum.processing.TransactionProcessingResult;
import org.hyperledger.besu.ethereum.referencetests.GeneralStateTestCaseEipSpec;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestBlockchain;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState;
import org.hyperledger.besu.ethereum.rlp.RLP;
import org.hyperledger.besu.ethereum.vm.BlockchainBasedBlockHashLookup;
import org.hyperledger.besu.evm.EVM;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.fluent.SimpleBlockValues;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;

@Slf4j
public class ToyExecutionTools {
  private static final List<String> SPECS_PRIOR_TO_DELETING_EMPTY_ACCOUNTS =
      Arrays.asList("Frontier", "Homestead", "EIP150");
  private static final CorsetValidator CORSET_VALIDATOR =
      new CorsetValidator(ChainConfig.MAINNET_TESTCONFIG);

  private ToyExecutionTools() {
    // utility class
  }

  @SneakyThrows
  public static void executeTest(
      final GeneralStateTestCaseEipSpec spec,
      final ProtocolSpec protocolSpec,
      final ZkTracer tracer,
      final TransactionProcessingResultValidator transactionProcessingResultValidator,
      final Consumer<ZkTracer> zkTracerValidator) {
    final BlockHeader blockHeader = spec.getBlockHeader();
    final ReferenceTestWorldState initialWorldState = spec.getInitialWorldState();
    final List<Transaction> transactions = new ArrayList<>();
    for (int i = 0; i < spec.getTransactionsCount(); i++) {
      Transaction transaction = spec.getTransaction(i);

      // Sometimes the tests ask us assemble an invalid transaction.  If we have
      // no valid transaction then there is no test.  GeneralBlockChain tests
      // will handle the case where we receive the TXs in a serialized form.
      if (transaction == null) {
        assertThat(spec.getExpectException())
            .withFailMessage("Transaction was not assembled, but no exception was expected")
            .isNotNull();
        return;
      }

      transactions.add(transaction);
    }

    final BlockBody blockBody = new BlockBody(transactions, new ArrayList<>());
    final MutableWorldState worldState = initialWorldState;
    //     final MutableWorldState worldState = initialWorldState.copy();
    final WorldUpdater worldStateUpdater = worldState.updater();
    final MainnetTransactionProcessor processor = protocolSpec.getTransactionProcessor();
    final ReferenceTestBlockchain blockchain = new ReferenceTestBlockchain(blockHeader.getNumber());
    final Wei blobGasPrice =
        protocolSpec
            .getFeeMarket()
            .blobGasPricePerGas(blockHeader.getExcessBlobGas().orElse(BlobGas.ZERO));

    tracer.traceStartConflation(1);
    tracer.traceStartBlock(blockHeader, blockBody, blockHeader.getCoinbase());
    TransactionProcessingResult result = null;
    for (Transaction transaction : blockBody.getTransactions()) {
      // Several of the GeneralStateTests check if the transaction could potentially
      // consume more gas than is left for the block it's attempted to be included in.
      // This check is performed within the `BlockImporter` rather than inside the
      // `TransactionProcessor`, so these tests are skipped.
      if (transaction.getGasLimit() > blockHeader.getGasLimit() - blockHeader.getGasUsed()) {
        throw new IllegalArgumentException("Transaction gas limit higher that available in block");
      }

      result =
          processor.processTransaction(
              worldStateUpdater,
              blockHeader,
              transaction,
              blockHeader.getCoinbase(),
              tracer,
              new BlockchainBasedBlockHashLookup(blockHeader, blockchain),
              false,
              TransactionValidationParams.processingBlock(),
              blobGasPrice);

      if (result.isInvalid()) {
        final TransactionProcessingResult finalResult = result;
        assertThat(spec.getExpectException())
            .withFailMessage(() -> finalResult.getValidationResult().getErrorMessage())
            .isNotNull();
        return;
      }

      transactionProcessingResultValidator.accept(transaction, result);
      zkTracerValidator.accept(tracer);
      worldStateUpdater.commit();
    }

    tracer.traceEndBlock(blockHeader, blockBody);
    tracer.traceEndConflation(worldStateUpdater);

    assertThat(spec.getExpectException())
        .withFailMessage("Exception was expected - " + spec.getExpectException())
        .isNull();

    final Account coinbase = worldStateUpdater.getOrCreate(spec.getBlockHeader().getCoinbase());
    if (coinbase != null && coinbase.isEmpty() && shouldClearEmptyAccounts(spec.getFork())) {
      worldStateUpdater.deleteAccount(coinbase.getAddress());
    }
    worldState.persist(blockHeader);

    // Check the world state root hash.
    final Hash expectedRootHash = spec.getExpectedRootHash();
    Optional.ofNullable(expectedRootHash)
        .ifPresent(
            expected -> {
              assertThat(worldState.rootHash())
                  .withFailMessage(
                      "Unexpected world state root hash; expected state: %s, computed state: %s",
                      spec.getExpectedRootHash(), worldState.rootHash())
                  .isEqualTo(expected);
            });

    // Check the logs.
    final Hash expectedLogsHash = spec.getExpectedLogsHash();
    final TransactionProcessingResult finalResult = result;
    Optional.ofNullable(expectedLogsHash)
        .ifPresent(
            expected -> {
              assert finalResult != null;
              final List<Log> logs = finalResult.getLogs();

              assertThat(Hash.hash(RLP.encode(out -> out.writeList(logs, Log::writeTo))))
                  .withFailMessage("Unmatched logs hash. Generated logs: %s", logs)
                  .isEqualTo(expected);
            });

    ExecutionEnvironment.checkTracer(
        tracer,
        CORSET_VALIDATOR,
        Optional.of(log),
        // block number for first block
        blockHeader.getNumber(),
        // block number for last block
        blockHeader.getNumber());
  }

  @SneakyThrows
  public static long executeTestOnlyForGasCost(
      final GeneralStateTestCaseEipSpec spec,
      final ProtocolSpec protocolSpec,
      final List<ToyAccount> accounts) {
    final MutableWorldState worldState = spec.getInitialWorldState();
    final WorldUpdater worldStateUpdater = worldState.updater();
    final MainnetTransactionProcessor processor = protocolSpec.getTransactionProcessor();
    final EVM evm = protocolSpec.getEvm();

    SimpleBlockValues blockValues = new SimpleBlockValues();
    blockValues.setBaseFee(Optional.of(Wei.of(1)));
    Account senderAccount = accounts.get(0);
    Account receiverAccount = accounts.get(1);
    Transaction tx = spec.getTransaction(0);
    Bytes txPayload = tx.getPayload();
    Wei txValue = tx.getValue();

    MessageFrame initialMessageFrame =
        MessageFrame.builder()
            .worldUpdater(worldStateUpdater)
            .gasPrice(Wei.ONE)
            .blobGasPrice(Wei.ONE)
            .blockValues(blockValues)
            .miningBeneficiary(Address.ZERO)
            .blockHashLookup((__, ___) -> Hash.ZERO)
            .completer(messageFrame -> {})
            .apparentValue(Wei.ZERO)
            .value(txValue)
            .inputData(txPayload)
            .originator(senderAccount.getAddress())
            .address(receiverAccount.getAddress())
            .contract(receiverAccount.getAddress())
            .sender(senderAccount.getAddress())
            // For gas cost purposes, we don't care about the Type of the message frame
            .type(MessageFrame.Type.MESSAGE_CALL)
            .initialGas(LINEA_BLOCK_GAS_LIMIT)
            .code(evm.getCodeUncached(receiverAccount.getCode()))
            .build();

    Deque<MessageFrame> messageFrameStack = initialMessageFrame.getMessageFrameStack();
    while (!messageFrameStack.isEmpty()) {
      processor.process(
          messageFrameStack.peekFirst(), new ZkTracer(ChainConfig.MAINNET_TESTCONFIG));
    }

    final long intrinsicTxCostWithNoAccessOrDelegationCost =
        evm.getGasCalculator().transactionIntrinsicGasCost(tx, 0);

    return LINEA_BLOCK_GAS_LIMIT
        - initialMessageFrame.getRemainingGas()
        + intrinsicTxCostWithNoAccessOrDelegationCost;
  }

  private static boolean shouldClearEmptyAccounts(final String eip) {
    return !SPECS_PRIOR_TO_DELETING_EMPTY_ACCOUNTS.contains(eip);
  }
}
