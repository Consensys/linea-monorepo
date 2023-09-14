/*
 * Copyright ConsenSys AG.
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

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.function.Consumer;

import lombok.Builder;
import lombok.RequiredArgsConstructor;
import lombok.Singular;
import net.consensys.linea.zktracer.ZkBlockAwareOperationTracer;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.corset.CorsetValidator;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.BlockBody;
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.evm.Code;
import org.hyperledger.besu.evm.EVM;
import org.hyperledger.besu.evm.MainnetEVMs;
import org.hyperledger.besu.evm.frame.BlockValues;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.precompile.PrecompileContractRegistry;
import org.hyperledger.besu.evm.processor.MessageCallProcessor;
import org.hyperledger.besu.plugin.data.BlockHeader;

/** Fluent API for executing EVM transactions in tests. */
@Builder
@RequiredArgsConstructor
public class ToyExecutionEnvironment {
  private static final Address DEFAULT_SENDER_ADDRESS = Address.fromHexString("0xe8f1b89");
  private static final Wei DEFAULT_VALUE = Wei.ZERO;
  private static final Bytes DEFAULT_INPUT_DATA = Bytes.EMPTY;
  private static final Bytes DEFAULT_BYTECODE = Bytes.EMPTY;
  private static final long DEFAULT_GAS_LIMIT = 1_000_000;
  private static final ToyWorld DEFAULT_TOY_WORLD = ToyWorld.empty();

  private final BlockValues blockValues = ToyBlockValues.builder().number(13L).build();
  private final ToyWorld toyWorld;
  private final EVM evm;
  private final ZkBlockAwareOperationTracer tracer;
  @Singular private final List<Transaction> transactions;
  private final Consumer<MessageFrame> frameAssertions;
  private final Consumer<MessageFrame> customFrameSetup;

  /**
   * Gets the default EVM implementation.
   *
   * @return default EVM implementation
   */
  public static EVM defaultEvm() {
    return MainnetEVMs.london(EvmConfiguration.DEFAULT);
  }

  /**
   * Gets the default tracer implementation.
   *
   * @return the default tracer implementation
   */
  public static ZkBlockAwareOperationTracer defaultTracer() {
    return new ZkTracer();
  }

  /**
   * Execute constructed EVM bytecode and return a JSON trace.
   *
   * @return the generated JSON trace
   */
  public String traceCode() {
    executeCode();

    return tracer.getJsonTrace();
  }

  /** Execute constructed EVM bytecode and perform Corset trace validation. */
  public void run() {
    executeCode();

    assertThat(CorsetValidator.isValid(tracer.getJsonTrace())).isTrue();
  }

  private MessageFrame prepareFrame(final Transaction tx) {
    final Bytes byteCode =
        toyWorld
            .get(
                tx.getTo()
                    .orElseThrow(
                        () ->
                            new IllegalArgumentException(
                                "Cannot fetch receiver account address from transaction")))
            .getCode();

    final Code code = evm.getCode(Hash.hash(byteCode), byteCode);

    return new TestMessageFrameBuilder()
        .worldUpdater(this.toyWorld.updater())
        .initialGas(DEFAULT_GAS_LIMIT)
        .address(DEFAULT_SENDER_ADDRESS)
        .originator(DEFAULT_SENDER_ADDRESS)
        .contract(DEFAULT_SENDER_ADDRESS)
        .gasPrice(Wei.ZERO)
        .inputData(DEFAULT_INPUT_DATA)
        .sender(DEFAULT_SENDER_ADDRESS)
        .value(DEFAULT_VALUE)
        .code(code)
        .blockValues(blockValues)
        .build();
  }

  private void setupFrame(final MessageFrame frame) {
    if (customFrameSetup != null) {
      customFrameSetup.accept(frame);
    }
  }

  private void postTest(final MessageFrame frame) {
    if (frameAssertions != null) {
      frameAssertions.accept(frame);
    }
  }

  private void executeCode() {
    final MessageCallProcessor messageCallProcessor =
        new MessageCallProcessor(evm, new PrecompileContractRegistry());

    BlockHeader mockBlockHeader = BlockHeaderBuilder.createDefault().buildBlockHeader();
    BlockBody mockBlockBody = new BlockBody(transactions, new ArrayList<>());

    tracer.traceStartConflation(1);
    tracer.traceStartBlock(mockBlockHeader, mockBlockBody);

    for (Transaction tx : mockBlockBody.getTransactions()) {
      MessageFrame frame = this.prepareFrame(tx);
      setupFrame(frame);

      tracer.traceStartTransaction(toyWorld.updater(), tx);
      messageCallProcessor.process(frame, this.tracer);
      tracer.traceEndTransaction(
          toyWorld.updater(),
          tx,
          frame.getState() == MessageFrame.State.COMPLETED_SUCCESS,
          frame.getOutputData(),
          frame.getLogs(),
          0, // TODO
          0);

      postTest(frame);
    }

    tracer.traceEndBlock(mockBlockHeader, mockBlockBody);
    tracer.traceEndConflation();
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
        .chainId(BigInteger.valueOf(23))
        //      .versionedHashes(List.of())
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
          Optional.ofNullable(tracer).orElse(defaultTracer()),
          Optional.ofNullable(transactions).orElse(defaultTxList),
          frameAssertions,
          customFrameSetup);
    }
  }
}
