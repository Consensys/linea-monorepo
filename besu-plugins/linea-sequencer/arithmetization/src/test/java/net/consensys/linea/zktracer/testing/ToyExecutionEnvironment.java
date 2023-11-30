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

package net.consensys.linea.zktracer.testing;

import static net.consensys.linea.zktracer.runtime.stack.Stack.MAX_STACK_SIZE;
import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.function.Consumer;

import lombok.Builder;
import lombok.RequiredArgsConstructor;
import lombok.Singular;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.*;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.core.feemarket.CoinbaseFeePriceCalculator;
import org.hyperledger.besu.ethereum.mainnet.LondonTargetingGasLimitCalculator;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.mainnet.TransactionValidatorFactory;
import org.hyperledger.besu.ethereum.mainnet.feemarket.FeeMarket;
import org.hyperledger.besu.ethereum.mainnet.feemarket.LondonFeeMarket;
import org.hyperledger.besu.ethereum.processing.TransactionProcessingResult;
import org.hyperledger.besu.evm.EVM;
import org.hyperledger.besu.evm.MainnetEVMs;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.precompile.PrecompileContractRegistry;
import org.hyperledger.besu.evm.processor.ContractCreationProcessor;
import org.hyperledger.besu.evm.processor.MessageCallProcessor;
import org.hyperledger.besu.plugin.data.BlockHeader;

/** Fluent API for executing EVM transactions in tests. */
@Builder
@RequiredArgsConstructor
@Slf4j
public class ToyExecutionEnvironment {
  public static final BigInteger CHAIN_ID = BigInteger.valueOf(1337);
  private static final CorsetValidator corsetValidator = new CorsetValidator();

  private static final Address DEFAULT_SENDER_ADDRESS = Address.fromHexString("0xe8f1b89");
  private static final Wei DEFAULT_VALUE = Wei.ZERO;
  private static final Bytes DEFAULT_INPUT_DATA = Bytes.EMPTY;
  private static final Bytes DEFAULT_BYTECODE = Bytes.EMPTY;
  private static final long DEFAULT_GAS_LIMIT = 1_000_000;
  private static final ToyWorld DEFAULT_TOY_WORLD = ToyWorld.empty();
  private static final Wei DEFAULT_BASE_FEE = Wei.of(1_000_000L);

  private static final GasCalculator gasCalculator = ZkTracer.gasCalculator;
  private static final Address minerAddress = Address.fromHexString("0x1234532342");

  private final ToyWorld toyWorld;
  private final EVM evm;
  @Singular private final List<Transaction> transactions;

  /**
   * A function applied to the {@link TransactionProcessingResult} of each transaction; by default,
   * asserts that the transaction is successful.
   */
  private final Consumer<TransactionProcessingResult> testValidator;

  private final Consumer<ZkTracer> zkTracerValidator;

  private static final FeeMarket feeMarket = FeeMarket.london(-1);
  private final ZkTracer tracer = new ZkTracer();

  /**
   * Gets the default EVM implementation, i.e. London.
   *
   * @return default EVM implementation
   */
  public static EVM defaultEvm() {
    return MainnetEVMs.london(EvmConfiguration.DEFAULT);
  }

  public static void checkTracer(ZkTracer tracer) {
    try {
      final Path traceFile = Files.createTempFile(null, ".lt");
      tracer.writeToFile(traceFile);
      log.info("trace written to `{}`", traceFile);
      assertThat(corsetValidator.validate(traceFile).isValid()).isTrue();
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  public void run() {
    execute();
    checkTracer(this.tracer);
  }

  private void execute() {
    BlockHeader header =
        BlockHeaderBuilder.createDefault().baseFee(DEFAULT_BASE_FEE).buildBlockHeader();
    BlockBody mockBlockBody = new BlockBody(transactions, new ArrayList<>());

    final MainnetTransactionProcessor transactionProcessor = getMainnetTransactionProcessor();

    tracer.traceStartConflation(1);
    tracer.traceStartBlock(header, mockBlockBody);

    for (Transaction tx : mockBlockBody.getTransactions()) {
      tracer.traceStartTransaction(toyWorld.updater(), tx);

      final TransactionProcessingResult result =
          transactionProcessor.processTransaction(
              null,
              toyWorld.updater(),
              (ProcessableBlockHeader) header,
              tx,
              minerAddress,
              tracer,
              blockId -> {
                throw new RuntimeException("Block hash lookup not yet supported");
              },
              false,
              Wei.ZERO);

      long transactionGasUsed = tx.getGasLimit() - result.getGasRemaining();

      tracer.traceEndTransaction(
          toyWorld.updater(),
          tx,
          result.isSuccessful(),
          result.getOutput(),
          result.getLogs(),
          transactionGasUsed,
          0);

      this.testValidator.accept(result);
      this.zkTracerValidator.accept(tracer);
    }

    tracer.traceEndBlock(header, mockBlockBody);
    tracer.traceEndConflation();
  }

  private MainnetTransactionProcessor getMainnetTransactionProcessor() {
    final MessageCallProcessor messageCallProcessor =
        new MessageCallProcessor(evm, new PrecompileContractRegistry());

    final ContractCreationProcessor contractCreationProcessor =
        new ContractCreationProcessor(evm.getGasCalculator(), evm, false, List.of(), 0);

    return new MainnetTransactionProcessor(
        gasCalculator,
        new TransactionValidatorFactory(
            gasCalculator,
            new LondonTargetingGasLimitCalculator(0L, new LondonFeeMarket(0, Optional.empty())),
            false,
            Optional.of(CHAIN_ID),
            Set.of(TransactionType.FRONTIER, TransactionType.ACCESS_LIST, TransactionType.EIP1559)),
        contractCreationProcessor,
        messageCallProcessor,
        true,
        true,
        MAX_STACK_SIZE,
        feeMarket,
        CoinbaseFeePriceCalculator.eip1559());
  }

  private static Transaction defaultTransaction() {
    return Transaction.builder()
        .nonce(123L)
        .type(TransactionType.FRONTIER)
        .gasPrice(Wei.of(1500))
        .gasLimit(DEFAULT_GAS_LIMIT)
        .to(Address.fromHexString("0x1234567890"))
        .value(DEFAULT_VALUE)
        .payload(DEFAULT_INPUT_DATA)
        .sender(DEFAULT_SENDER_ADDRESS)
        .chainId(CHAIN_ID)
        .signAndBuild(new SECP256K1().generateKeyPair());
  }

  /** Customizations applied to the Lombok generated builder. */
  public static class ToyExecutionEnvironmentBuilder {
    /**
     * Builder method returning an instance of {@link ToyExecutionEnvironment}.
     *
     * @return an instance of {@link ToyExecutionEnvironment}
     */
    public ToyExecutionEnvironment build() {
      var defaultTxList = new ArrayList<>(List.of(defaultTransaction()));

      return new ToyExecutionEnvironment(
          Optional.ofNullable(toyWorld).orElse(DEFAULT_TOY_WORLD),
          Optional.ofNullable(evm).orElse(defaultEvm()),
          Optional.ofNullable(transactions).orElse(defaultTxList),
          Optional.ofNullable(testValidator)
              .orElse(result -> assertThat(result.isSuccessful()).isTrue()),
          Optional.ofNullable(zkTracerValidator).orElse(zkTracer -> {}));
    }
  }
}
