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

package net.consensys.linea.zktracer.module.mmu;

import static net.consensys.linea.zktracer.types.Conversions.bytesToUnsignedBytes;
import static net.consensys.linea.zktracer.types.Conversions.unsignedBytesToEWord;

import java.math.BigInteger;
import java.util.Arrays;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@AllArgsConstructor
@Accessors(fluent = true)
class MicroData {
  private static final UnsignedByte[] DEFAULT_NIBBLES = new UnsignedByte[9];
  private static final UnsignedByte[][] DEFAULT_ACCS = new UnsignedByte[8][32];
  private static final boolean[] DEFAULT_BITS = new boolean[8];
  private static final UnsignedByte[] DEFAULT_VAL_A_CACHE = new UnsignedByte[16];
  private static final UnsignedByte[] DEFAULT_VAL_C_CACHE = new UnsignedByte[16];
  private static final UnsignedByte[] DEFAULT_VAL_B_CACHE = new UnsignedByte[16];

  @Getter @Setter private int callStackDepth;
  @Getter @Setter private int callStackSize;
  @Getter @Setter private int callDataOffset;
  @Getter @Setter private int callDataSize;
  @Getter @Setter private OpCode opCode;
  @Getter @Setter private boolean skip;
  @Getter @Setter private int processingRow;
  // Same as {@link ReadPad#totalNumberPaddingMicroInstructions()} in size.
  @Getter @Setter private boolean toRam;
  // < 16
  @Getter @Setter private int counter;
  @Getter @Setter private int microOp;
  @Getter @Setter private boolean aligned;
  @Getter @Setter private ReadPad readPad;
  // < 1.000.000
  @Getter @Setter private int sizeImported;
  // < 1.000.000
  @Getter @Setter private int size;
  // stack element <=> uint256
  @Getter @Setter private Bytes value;
  // stack element
  @Getter @Setter private Pointers pointers;
  // stack element
  @Getter @Setter private Offsets offsets;
  // precomputation type
  @Getter @Setter private int precomputation;
  // < 1.000.000
  @Getter @Setter private int min;
  @Getter @Setter private int ternary;
  @Getter @Setter private InstructionContext instructionContext;
  @Getter @Setter private Contexts contexts;
  // every acc is actually on 16 bytes
  @Getter @Setter private UnsignedByte[] nibbles;
  @Getter @Setter private UnsignedByte[][] accs;
  @Getter @Setter private boolean[] bits;
  @Getter @Setter private boolean exoIsHash;
  @Getter @Setter private boolean exoIsLog;
  @Getter @Setter private boolean exoIsRom;
  @Getter @Setter private boolean exoIsTxcd;
  @Getter @Setter private boolean info;
  private UnsignedByte[] valACache;
  private UnsignedByte[] valBCache;
  private UnsignedByte[] valCCache;
  // < 1.000.000
  @Getter @Setter private int referenceOffset;
  // < 1.000.000
  @Getter @Setter private int referenceSize;

  MicroData() {
    this(
        0,
        0,
        0,
        0,
        null,
        false,
        0,
        false,
        0,
        0,
        false,
        null,
        0,
        0,
        null,
        null,
        null,
        0,
        0,
        0,
        null,
        null,
        DEFAULT_NIBBLES,
        DEFAULT_ACCS,
        DEFAULT_BITS,
        false,
        false,
        false,
        false,
        false,
        DEFAULT_VAL_A_CACHE,
        DEFAULT_VAL_B_CACHE,
        DEFAULT_VAL_C_CACHE,
        0,
        0);
  }

  boolean isErf() {
    return microOp == MmuTrace.StoreXInAThreeRequired;
  }

  boolean isFast() {
    return Arrays.asList(
            MmuTrace.RamToRam,
            MmuTrace.ExoToRam,
            MmuTrace.RamIsExo,
            MmuTrace.KillingOne,
            MmuTrace.PushTwoRamToStack,
            MmuTrace.PushOneRamToStack,
            MmuTrace.ExceptionalRamToStack3To2FullFast,
            MmuTrace.PushTwoStackToRam,
            MmuTrace.StoreXInAThreeRequired,
            MmuTrace.StoreXInB,
            MmuTrace.StoreXInC)
        .contains(microOp);
  }

  boolean isType5() {
    return microOp == OpCode.CALLDATALOAD.getData().value();
  }

  int remainingMicroInstructions() {
    return readPad.remainingMicroInstructions(processingRow);
  }

  boolean isRead() {
    return readPad.isRead(processingRow);
  }

  int remainingReads() {
    return readPad.remainingReads(processingRow);
  }

  int remainingPads() {
    return readPad.remainingPads(processingRow);
  }

  boolean isFirstRead() {
    return readPad.isFirstRead(processingRow);
  }

  boolean isFirstPad() {
    return readPad.isFirstPad(processingRow);
  }

  boolean isFirstMicroInstruction() {
    return readPad.isFirstMicroInstruction(processingRow);
  }

  boolean isLastRead() {
    return readPad.isLastRead(processingRow);
  }

  boolean isLastPad() {
    return readPad.isLastPad(processingRow);
  }

  boolean isRootContext() {
    return callStackDepth == 1;
  }

  int sourceContext() {
    return contexts.source();
  }

  void sourceContext(final int value) {
    contexts.source(value);
  }

  int targetContext() {
    return contexts.target();
  }

  void targetContext(final int value) {
    contexts.target(value);
  }

  EWord sourceLimbOffset() {
    return offsets.source().limb();
  }

  void sourceLimbOffset(final EWord value) {
    offsets.source().limb(value);
  }

  UnsignedByte sourceByteOffset() {
    return offsets.source().uByte();
  }

  void sourceByteOffset(final UnsignedByte value) {
    offsets.source().uByte(value);
  }

  EWord targetLimbOffset() {
    return offsets.target().limb();
  }

  void targetLimbOffset(final EWord value) {
    offsets.target().limb(value);
  }

  UnsignedByte targetByteOffset() {
    return offsets.target().uByte();
  }

  void targetByteOffset(final UnsignedByte value) {
    offsets.target().uByte(value);
  }

  EWord getAccsAtIndex(final int index) {
    return unsignedBytesToEWord(accs[index]);
  }

  void setAccsAtIndex(final int index, final EWord value) {
    byte[] rawBytes = value.hiBigInt().toByteArray();

    accs[index] = bytesToUnsignedBytes(rawBytes);
  }

  void setAccsAtIndex(final int index, final BigInteger value) {
    byte[] rawBytes = value.toByteArray();

    accs[index] = bytesToUnsignedBytes(rawBytes);
  }

  void setAccsAndNibblesAtIndex(final int index, final EWord value) {
    EWord div = value.divide(16);
    setAccsAtIndex(index, div);

    UnsignedByte modulus = UnsignedByte.of(value.mod(16).toLong());
    nibbles[index] = modulus;
  }

  void setInfo(final CallStack callStack) {
    if (Arrays.asList(OpCode.CODECOPY, OpCode.RETURN).contains(opCode)) {
      info = callStack.current().type() == CallFrameType.INIT_CODE;
      // TODO: settle EXTCODEDOPY info for CODECOPY
    } else if (opCode == OpCode.CALLDATACOPY) {
      info = callStack.depth() == 1;
    }
  }

  void incrementCounter(final int value) {
    counter += value;
  }

  void incrementProcessingRow(final int value) {
    processingRow += value;
  }
}
