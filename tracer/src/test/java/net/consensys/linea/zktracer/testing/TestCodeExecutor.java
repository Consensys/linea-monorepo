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
import java.util.Optional;
import java.util.function.Consumer;

import lombok.Builder;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ZkBlockAwareOperationTracer;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.corset.CorsetValidator;
import net.consensys.linea.zktracer.toy.ToyWorld;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.BlockBody;
import org.hyperledger.besu.ethereum.core.BlockHeaderBuilder;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.evm.Code;
import org.hyperledger.besu.evm.EVM;
import org.hyperledger.besu.evm.MainnetEVMs;
import org.hyperledger.besu.evm.account.MutableAccount;
import org.hyperledger.besu.evm.frame.BlockValues;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.precompile.PrecompileContractRegistry;
import org.hyperledger.besu.evm.processor.MessageCallProcessor;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.plugin.data.BlockHeader;

/** Fluent API for executing EVM bytecode in tests. */
@Builder
@RequiredArgsConstructor
public class TestCodeExecutor {
  private static final Address DEFAULT_SENDER_ADDRESS = Address.fromHexString("0xe8f1b89");
  public static final Wei DEFAULT_VALUE = Wei.ZERO;
  public static final Bytes DEFAULT_INPUT_DATA = Bytes.EMPTY;
  public static final long DEFAULT_GAS_LIMIT = 1_000_000;

  private final BlockValues blockValues = new FakeBlockValues(13);
  private final WorldUpdater toyWorld = new ToyWorld();

  private final EVM evm;
  private final ZkBlockAwareOperationTracer tracer;
  @Getter private final Bytes byteCode;
  private final Consumer<MessageFrame> frameAssertions;
  private final Consumer<MessageFrame> customFrameSetup;
  private final Consumer<MutableAccount> senderAccountSetup;

  /**
   * Gets the default EVM implementation.
   *
   * @return default EVM implementation
   */
  public static EVM defaultEvm() {
    return MainnetEVMs.paris(EvmConfiguration.DEFAULT);
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
   * Constructor with a predefined EVM and tracer.
   *
   * @param byteCode constructed bytecode
   * @param frameAssertions a collections of assertions on {@link MessageFrame}
   * @param customFrameSetup optional customization on the {@link MessageFrame}
   * @param senderAccountSetup optional customization on {@link MutableAccount}
   */
  public TestCodeExecutor(
      final Bytes byteCode,
      final Consumer<MessageFrame> frameAssertions,
      final Consumer<MessageFrame> customFrameSetup,
      final Consumer<MutableAccount> senderAccountSetup) {
    this(
        defaultEvm(),
        defaultTracer(),
        byteCode,
        frameAssertions,
        customFrameSetup,
        senderAccountSetup);
  }

  /**
   * Constructor with predefined tracer.
   *
   * @param evm custom EVM type and configuration
   * @param byteCode constructed bytecode
   * @param frameAssertions a collections of assertions on {@link MessageFrame}
   * @param customFrameSetup optional customization on the {@link MessageFrame}
   * @param senderAccountSetup optional customization on {@link MutableAccount}
   */
  public TestCodeExecutor(
      final EVM evm,
      final Bytes byteCode,
      final Consumer<MessageFrame> frameAssertions,
      final Consumer<MessageFrame> customFrameSetup,
      final Consumer<MutableAccount> senderAccountSetup) {
    this(evm, defaultTracer(), byteCode, frameAssertions, customFrameSetup, senderAccountSetup);
  }

  /**
   * Constructor with a predefined EVM.
   *
   * @param tracer custom tracer implementation
   * @param byteCode constructed bytecode
   * @param frameAssertions a collections of assertions on {@link MessageFrame}
   * @param customFrameSetup optional customization on the {@link MessageFrame}
   * @param senderAccountSetup optional customization on {@link MutableAccount}
   */
  public TestCodeExecutor(
      final ZkBlockAwareOperationTracer tracer,
      final Bytes byteCode,
      final Consumer<MessageFrame> frameAssertions,
      final Consumer<MessageFrame> customFrameSetup,
      final Consumer<MutableAccount> senderAccountSetup) {
    this(defaultEvm(), tracer, byteCode, frameAssertions, customFrameSetup, senderAccountSetup);
  }

  /**
   * Execute constructed EVM bytecode and return a JSON trace.
   *
   * @return the generated JSON trace
   */
  public String traceCode() {
    MessageFrame frame = executeCode();
    this.postTest(frame);

    return tracer.getJsonTrace();
  }

  /** Execute constructed EVM bytecode and perform Corset trace validation. */
  public void run() {
    MessageFrame frame = executeCode();
    this.postTest(frame);

    assertThat(CorsetValidator.isValid(tracer.getJsonTrace())).isTrue();
  }

  /**
   * Deploy a contract.
   *
   * @param contractAddress address of the contract
   * @param codeBytes bytecode of the contract
   */
  public void deployContract(final Address contractAddress, final Bytes codeBytes) {
    var updater = this.toyWorld.updater();
    final MutableAccount contract = updater.getOrCreate(contractAddress).getMutable();

    contract.setNonce(0);
    contract.clearStorage();
    contract.setCode(codeBytes);

    updater.commit();
  }

  /** Create the initial world state of the EVM. */
  public void createInitialWorldState() {
    final WorldUpdater worldState = this.toyWorld.updater();
    final MutableAccount senderAccount =
        worldState.getOrCreate(DEFAULT_SENDER_ADDRESS).getMutable();

    setupSenderAccount(senderAccount);

    worldState.commit();
  }

  private MessageFrame prepareFrame() {
    final Code code = evm.getCode(Hash.hash(this.byteCode), this.byteCode);

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

  private void setupSenderAccount(MutableAccount senderAccount) {
    if (senderAccountSetup != null) {
      senderAccountSetup.accept(senderAccount);
    }
  }

  private void setupFrame(MessageFrame frame) {
    if (customFrameSetup != null) {
      customFrameSetup.accept(frame);
    }
  }

  private void postTest(MessageFrame frame) {
    if (frameAssertions != null) {
      frameAssertions.accept(frame);
    }
  }

  private MessageFrame executeCode() {
    final MessageCallProcessor messageCallProcessor =
        new MessageCallProcessor(evm, new PrecompileContractRegistry());

    final MessageFrame frame = this.prepareFrame();
    setupFrame(frame);

    BlockHeader mockBlockHeader = BlockHeaderBuilder.createDefault().buildBlockHeader();

    Transaction tx = defaultTransaction();

    BlockBody mockBlockBody =
        new BlockBody(new ArrayList<>() /* transactions */, new ArrayList<>() /* ommers */);
    tracer.traceStartConflation(1);
    tracer.traceStartBlock(mockBlockHeader, mockBlockBody);
    tracer.traceStartTransaction(tx);

    messageCallProcessor.process(frame, this.tracer);

    tracer.traceEndTransaction(Bytes.EMPTY, 0, 0); // TODO
    tracer.traceEndBlock(mockBlockHeader, mockBlockBody);
    tracer.traceEndConflation();

    return frame;
  }

  private static Transaction defaultTransaction() {
    return new Transaction(
        123L,
        Wei.of(1500),
        DEFAULT_GAS_LIMIT,
        Optional.of(Address.fromHexString("0x1234567890")),
        DEFAULT_VALUE,
        null, // TODO
        DEFAULT_INPUT_DATA,
        DEFAULT_SENDER_ADDRESS,
        Optional.of(BigInteger.valueOf(23)),
        Optional.empty());
  }

  /** Customizations performed on the Lombok generated builder. */
  public static class TestCodeExecutorBuilder {

    /**
     * Builder method returning an instance of {@link TestCodeExecutor}.
     *
     * @return an instance of {@link TestCodeExecutor}
     */
    public TestCodeExecutor build() {
      if (evm != null && tracer != null) {
        return new TestCodeExecutor(
            evm, tracer, byteCode, frameAssertions, customFrameSetup, senderAccountSetup);
      } else if (evm != null) {
        return new TestCodeExecutor(
            evm, byteCode, frameAssertions, customFrameSetup, senderAccountSetup);
      } else if (tracer != null) {
        return new TestCodeExecutor(
            tracer, byteCode, frameAssertions, customFrameSetup, senderAccountSetup);
      }

      return new TestCodeExecutor(byteCode, frameAssertions, customFrameSetup, senderAccountSetup);
    }
  }
}
