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

package net.consensys.linea.zktracer.module.mmu;

import static net.consensys.linea.zktracer.types.Conversions.booleanToBigInteger;
import static net.consensys.linea.zktracer.types.Conversions.unsignedBytesToUnsignedBigInteger;

import java.math.BigInteger;
import java.util.Map;

import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.module.mmio.Mmio;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.StackOperation;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.commons.lang3.ArrayUtils;

public class Mmu implements Module {
  private final StackedList<MicroData> state = new StackedList<>();
  private Mmio mmio;
  private int ramStamp;
  private boolean isMicro;
  private final MicroDataProcessor microDataProcessor;

  private final CallStack callStack;

  public Mmu(final CallStack callStack) {
    this.callStack = callStack;
    this.mmio = new Mmio();
    this.microDataProcessor = new MicroDataProcessor();
  }

  @Override
  public String jsonKey() {
    return "mmu";
  }

  @Override
  public void enterTransaction() {
    this.state.enter();
  }

  @Override
  public void popTransaction() {
    this.state.pop();
  }

  @Override
  public int lineCount() {
    return this.state.stream().mapToInt(m -> maxCounter(m.pointers().oob())).sum();
  }

  @Override
  public ModuleTrace commit() {
    final Trace.TraceBuilder trace = Trace.builder(this.lineCount());

    for (MicroData m : this.state) {
      traceMicroData(m, callStack, trace);
    }

    return new MmuTrace(trace.build());
  }

  /**
   * TODO: should be called from the hub.
   *
   * @param opCode
   * @param stackOps
   * @param callStack
   */
  public void handleRam(
      final OpCode opCode, final Map<Integer, StackOperation> stackOps, final CallStack callStack) {
    MicroData microData = microDataProcessor.dispatchOpCode(opCode, stackOps, callStack);

    this.state.add(microData);
  }

  private void traceMicroData(
      MicroData microData, final CallStack callStack, Trace.TraceBuilder trace) {
    if (microData.skip()) {
      return;
    }

    this.ramStamp++;
    this.isMicro = false;
    int maxCounter = maxCounter(microData.pointers().oob());

    microData.processingRow(-1);

    while (microData.counter() < maxCounter) {
      microDataProcessor.initializePreProcessing(callStack);
      trace(microData, trace);
      microData.incrementCounter(1);
    }

    this.isMicro = true;
    microData.counter(0);
    microData.processingRow(0);

    while (microData.processingRow() < microData.readPad().totalNumber()) {
      microDataProcessor.initializeProcessing(callStack, microData);
      //      self.Mmio.handleRam(&uop, self.MicroStamp, callStack, moduleStamp)
      trace(microData, trace);
      microData.incrementProcessingRow(1);
    }
  }

  private void trace(MicroData microData, Trace.TraceBuilder trace) {
    Pointers pointers = microData.pointers();

    BigInteger value = microData.value().toUnsignedBigInteger();

    InstructionContext stackFrames = microData.instructionContext();

    UnsignedByte[] nibbles = microData.nibbles();

    boolean[] bits = microData.bits();

    trace
        .ramStamp(BigInteger.valueOf(this.ramStamp))
        .microInstructionStamp(BigInteger.ZERO)
        .isMicroInstruction(this.isMicro)
        .off1Lo(pointers.stack1().lo().toUnsignedBigInteger())
        .off2Hi(pointers.stack2().toUnsignedBigInteger())
        .sizeImported(BigInteger.valueOf(microData.sizeImported()))
        .valHi(value)
        .valLo(value)
        .contextNumber(BigInteger.valueOf(stackFrames.self()))
        .caller(BigInteger.valueOf(stackFrames.caller()))
        .returner(BigInteger.valueOf(stackFrames.returner()))
        .contextSource(BigInteger.valueOf(microData.sourceContext()))
        .contextTarget(BigInteger.valueOf(microData.targetContext()))
        .counter(BigInteger.valueOf(microData.counter()))
        .offsetOutOfBounds(pointers.oob())
        .precomputation(BigInteger.valueOf(microData.precomputation()))
        .ternary(BigInteger.valueOf(microData.ternary()))
        .microInstruction(BigInteger.valueOf(microData.microOp()))
        .exoIsRom(microData.exoIsRom())
        .exoIsLog(microData.exoIsLog())
        .exoIsHash(microData.exoIsHash())
        .exoIsTxcd(microData.exoIsTxcd())
        .sourceLimbOffset(microData.sourceLimbOffset().toUnsignedBigInteger())
        .sourceByteOffset(microData.sourceByteOffset().toBigInteger())
        .targetLimbOffset(microData.targetLimbOffset().toUnsignedBigInteger())
        .targetByteOffset(microData.targetByteOffset().toBigInteger())
        .size(BigInteger.valueOf(microData.size()))
        .nib1(nibbles[0])
        .nib2(nibbles[1])
        .nib3(nibbles[2])
        .nib4(nibbles[3])
        .nib5(nibbles[4])
        .nib6(nibbles[5])
        .nib7(nibbles[6])
        .nib8(nibbles[7])
        .nib9(nibbles[8])
        .acc1(acc(0, microData))
        .byte1(accByte(0, microData))
        .acc2(acc(1, microData))
        .byte2(accByte(1, microData))
        .acc3(acc(2, microData))
        .byte3(accByte(2, microData))
        .acc4(acc(3, microData))
        .byte4(accByte(3, microData))
        .acc5(acc(4, microData))
        .byte5(accByte(4, microData))
        .acc6(acc(5, microData))
        .byte6(accByte(5, microData))
        .acc7(acc(6, microData))
        .byte7(accByte(6, microData))
        .acc8(acc(7, microData))
        .byte8(accByte(7, microData))
        .bit1(bits[0])
        .bit2(bits[1])
        .bit3(bits[2])
        .bit4(bits[3])
        .bit5(bits[4])
        .bit6(bits[5])
        .bit7(bits[6])
        .bit8(bits[7])
        .aligned(booleanToBigInteger(microData.aligned()))
        .fast(booleanToBigInteger(microData.isFast()))
        .min(BigInteger.valueOf(microData.min()))
        .callStackDepth(BigInteger.valueOf(microData.callStackDepth()))
        .callDataSize(BigInteger.valueOf(microData.callDataOffset()))
        .instruction(BigInteger.valueOf(microData.opCode().getData().value()))
        .totalNumberOfMicroInstructions(
            BigInteger.valueOf(this.ramStamp == 0 ? 0 : microData.remainingMicroInstructions()))
        .totalNumberOfReads(BigInteger.valueOf(microData.remainingReads()))
        .totalNumberOfPaddings(BigInteger.valueOf(microData.remainingPads()))
        .toRam(microData.toRam())
        .erf(isMicro && microData.isErf())
        .returnOffset(
            stackFrames.returnOffset().isUInt64() || stackFrames.returnCapacity() == 0
                ? BigInteger.ZERO
                : stackFrames.returnOffset().toUnsignedBigInteger())
        .returnCapacity(BigInteger.valueOf(stackFrames.returnCapacity()))
        .refs(BigInteger.valueOf(microData.referenceSize()))
        .refo(BigInteger.valueOf(microData.referenceOffset()))
        .info(booleanToBigInteger(microData.info()))
        .isData(this.ramStamp != 0)
        .validateRow();
  }

  private BigInteger acc(final int accIndex, final MicroData microData) {
    int maxCounter = maxCounter(microData.pointers().oob());

    return unsignedBytesToUnsignedBigInteger(
        ArrayUtils.subarray(
            microData.accs()[accIndex],
            32 - maxCounter,
            32 - maxCounter + microData.counter() + 1));
  }

  private UnsignedByte accByte(final int accIndex, final MicroData microData) {
    int maxCounter = maxCounter(microData.pointers().oob());

    return microData.accs()[accIndex][32 - maxCounter + microData.counter()];
  }

  private int maxCounter(final boolean oob) {
    return oob ? 16 : 3;
  }
}
