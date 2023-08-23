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
package net.consensys.linea.zktracer.testutils;

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Optional;

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
import org.junit.jupiter.api.Test;

public abstract class TestCodeExecutor {
  private final BlockValues blockValues = new FakeBlockValues(13);
  private static final Address SENDER_ADDRESS = Address.fromHexString("0xe8f1b89");
  private final EVM evm;
  private final ZkTracer tracer;
  private final ToyWorld world = new ToyWorld();

  public void setupSenderAccount(MutableAccount senderAccount) {}

  public void setupFrame(MessageFrame frame) {}

  public void postTest(MessageFrame frame) {}

  public abstract Bytes getBytecode();

  public Address getSenderAddress() {
    return SENDER_ADDRESS;
  }

  //  public List<Transaction> getTransactions() {
  //    return List.of(
  //        new Transaction(
  //            123,
  //            Wei.of(1500),
  //            this.getGasLimit(),
  //            Optional.of(Address.fromHexString("0x1234567890")),
  //            this.getValue(),
  //            SECPSignature.create(),
  //            this.getInputData(),
  //            this.getSenderAddress(),
  //            Optional.of(23),
  //            Optional.empty()));
  //  }

  public Wei getValue() {
    return Wei.ZERO;
  }

  public Bytes getInputData() {
    return Bytes.EMPTY;
  }

  public long getGasLimit() {
    return 1_000_000;
  }

  public TestCodeExecutor() {
    this.evm = MainnetEVMs.paris(EvmConfiguration.DEFAULT);
    this.tracer = new ZkTracer();
  }

  public TestCodeExecutor(final EVM evm) {
    this.evm = evm;
    this.tracer = new ZkTracer();
  }

  public TestCodeExecutor(final EVM evm, ZkTracer tracer) {
    this.evm = evm;
    this.tracer = tracer;
  }

  private MessageFrame prepareFrame() {
    final Code code = evm.getCode(Hash.hash(this.getBytecode()), this.getBytecode());

    return new TestMessageFrameBuilder()
        .worldUpdater(this.world.updater())
        .initialGas(this.getGasLimit())
        .address(this.getSenderAddress())
        .originator(this.getSenderAddress())
        .contract(this.getSenderAddress())
        .gasPrice(Wei.ZERO)
        .inputData(this.getInputData())
        .sender(this.getSenderAddress())
        .value(this.getValue())
        .code(code)
        .blockValues(blockValues)
        .build();
  }

  protected MessageFrame executeCode() {
    final MessageCallProcessor messageCallProcessor =
        new MessageCallProcessor(evm, new PrecompileContractRegistry());

    final MessageFrame frame = this.prepareFrame();
    setupFrame(frame);
    BlockHeader mockBlockHeader = BlockHeaderBuilder.createDefault().buildBlockHeader();

    Transaction tx =
        new Transaction(
            123L,
            Wei.of(1500),
            this.getGasLimit(),
            Optional.of(Address.fromHexString("0x1234567890")),
            this.getValue(),
            null, // TODO
            this.getInputData(),
            this.getSenderAddress(),
            Optional.of(BigInteger.valueOf(23)),
            Optional.empty());

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

  @Test
  public void testCode() {
    MessageFrame frame = executeCode();
    this.postTest(frame);
    assertThat(CorsetValidator.isValid(tracer.getTrace().toJson())).isTrue();
  }

  public String traceCode() {
    MessageFrame frame = executeCode();
    this.postTest(frame);
    return tracer.getTrace().toJson();
  }

  public void deployContract(final Address contractAddress, final Bytes codeBytes) {
    var updater = this.world.updater();
    final MutableAccount contract = updater.getOrCreate(contractAddress).getMutable();

    contract.setNonce(0);
    contract.clearStorage();
    contract.setCode(codeBytes);
    updater.commit();
  }

  public void createInitialWorldState() {
    final WorldUpdater worldState = this.world.updater();
    final MutableAccount senderAccount =
        worldState.getOrCreate(this.getSenderAddress()).getMutable();

    setupSenderAccount(senderAccount);
    worldState.commit();
  }
}
