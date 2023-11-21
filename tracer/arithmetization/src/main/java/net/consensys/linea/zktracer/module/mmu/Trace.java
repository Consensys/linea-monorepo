/*
 * Copyright ConsenSys Inc.
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

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.units.bigints.UInt256;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer acc1;
  private final MappedByteBuffer acc2;
  private final MappedByteBuffer acc3;
  private final MappedByteBuffer acc4;
  private final MappedByteBuffer acc5;
  private final MappedByteBuffer acc6;
  private final MappedByteBuffer acc7;
  private final MappedByteBuffer acc8;
  private final MappedByteBuffer aligned;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
  private final MappedByteBuffer bit5;
  private final MappedByteBuffer bit6;
  private final MappedByteBuffer bit7;
  private final MappedByteBuffer bit8;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer byte5;
  private final MappedByteBuffer byte6;
  private final MappedByteBuffer byte7;
  private final MappedByteBuffer byte8;
  private final MappedByteBuffer callDataOffset;
  private final MappedByteBuffer callDataSize;
  private final MappedByteBuffer callStackDepth;
  private final MappedByteBuffer caller;
  private final MappedByteBuffer contextNumber;
  private final MappedByteBuffer contextSource;
  private final MappedByteBuffer contextTarget;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer erf;
  private final MappedByteBuffer exoIsHash;
  private final MappedByteBuffer exoIsLog;
  private final MappedByteBuffer exoIsRom;
  private final MappedByteBuffer exoIsTxcd;
  private final MappedByteBuffer fast;
  private final MappedByteBuffer info;
  private final MappedByteBuffer instruction;
  private final MappedByteBuffer isData;
  private final MappedByteBuffer isMicroInstruction;
  private final MappedByteBuffer microInstruction;
  private final MappedByteBuffer microInstructionStamp;
  private final MappedByteBuffer min;
  private final MappedByteBuffer nib1;
  private final MappedByteBuffer nib2;
  private final MappedByteBuffer nib3;
  private final MappedByteBuffer nib4;
  private final MappedByteBuffer nib5;
  private final MappedByteBuffer nib6;
  private final MappedByteBuffer nib7;
  private final MappedByteBuffer nib8;
  private final MappedByteBuffer nib9;
  private final MappedByteBuffer off1Lo;
  private final MappedByteBuffer off2Hi;
  private final MappedByteBuffer off2Lo;
  private final MappedByteBuffer offsetOutOfBounds;
  private final MappedByteBuffer precomputation;
  private final MappedByteBuffer ramStamp;
  private final MappedByteBuffer refo;
  private final MappedByteBuffer refs;
  private final MappedByteBuffer returnCapacity;
  private final MappedByteBuffer returnOffset;
  private final MappedByteBuffer returner;
  private final MappedByteBuffer size;
  private final MappedByteBuffer sizeImported;
  private final MappedByteBuffer sourceByteOffset;
  private final MappedByteBuffer sourceLimbOffset;
  private final MappedByteBuffer targetByteOffset;
  private final MappedByteBuffer targetLimbOffset;
  private final MappedByteBuffer ternary;
  private final MappedByteBuffer toRam;
  private final MappedByteBuffer totalNumberOfMicroInstructions;
  private final MappedByteBuffer totalNumberOfPaddings;
  private final MappedByteBuffer totalNumberOfReads;
  private final MappedByteBuffer valHi;
  private final MappedByteBuffer valLo;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("mmu.ACC_1", 32, length),
        new ColumnHeader("mmu.ACC_2", 32, length),
        new ColumnHeader("mmu.ACC_3", 32, length),
        new ColumnHeader("mmu.ACC_4", 32, length),
        new ColumnHeader("mmu.ACC_5", 32, length),
        new ColumnHeader("mmu.ACC_6", 32, length),
        new ColumnHeader("mmu.ACC_7", 32, length),
        new ColumnHeader("mmu.ACC_8", 32, length),
        new ColumnHeader("mmu.ALIGNED", 32, length),
        new ColumnHeader("mmu.BIT_1", 1, length),
        new ColumnHeader("mmu.BIT_2", 1, length),
        new ColumnHeader("mmu.BIT_3", 1, length),
        new ColumnHeader("mmu.BIT_4", 1, length),
        new ColumnHeader("mmu.BIT_5", 1, length),
        new ColumnHeader("mmu.BIT_6", 1, length),
        new ColumnHeader("mmu.BIT_7", 1, length),
        new ColumnHeader("mmu.BIT_8", 1, length),
        new ColumnHeader("mmu.BYTE_1", 1, length),
        new ColumnHeader("mmu.BYTE_2", 1, length),
        new ColumnHeader("mmu.BYTE_3", 1, length),
        new ColumnHeader("mmu.BYTE_4", 1, length),
        new ColumnHeader("mmu.BYTE_5", 1, length),
        new ColumnHeader("mmu.BYTE_6", 1, length),
        new ColumnHeader("mmu.BYTE_7", 1, length),
        new ColumnHeader("mmu.BYTE_8", 1, length),
        new ColumnHeader("mmu.CALL_DATA_OFFSET", 32, length),
        new ColumnHeader("mmu.CALL_DATA_SIZE", 32, length),
        new ColumnHeader("mmu.CALL_STACK_DEPTH", 32, length),
        new ColumnHeader("mmu.CALLER", 32, length),
        new ColumnHeader("mmu.CONTEXT_NUMBER", 32, length),
        new ColumnHeader("mmu.CONTEXT_SOURCE", 32, length),
        new ColumnHeader("mmu.CONTEXT_TARGET", 32, length),
        new ColumnHeader("mmu.COUNTER", 32, length),
        new ColumnHeader("mmu.ERF", 1, length),
        new ColumnHeader("mmu.EXO_IS_HASH", 1, length),
        new ColumnHeader("mmu.EXO_IS_LOG", 1, length),
        new ColumnHeader("mmu.EXO_IS_ROM", 1, length),
        new ColumnHeader("mmu.EXO_IS_TXCD", 1, length),
        new ColumnHeader("mmu.FAST", 32, length),
        new ColumnHeader("mmu.INFO", 32, length),
        new ColumnHeader("mmu.INSTRUCTION", 32, length),
        new ColumnHeader("mmu.IS_DATA", 1, length),
        new ColumnHeader("mmu.IS_MICRO_INSTRUCTION", 1, length),
        new ColumnHeader("mmu.MICRO_INSTRUCTION", 32, length),
        new ColumnHeader("mmu.MICRO_INSTRUCTION_STAMP", 32, length),
        new ColumnHeader("mmu.MIN", 32, length),
        new ColumnHeader("mmu.NIB_1", 1, length),
        new ColumnHeader("mmu.NIB_2", 1, length),
        new ColumnHeader("mmu.NIB_3", 1, length),
        new ColumnHeader("mmu.NIB_4", 1, length),
        new ColumnHeader("mmu.NIB_5", 1, length),
        new ColumnHeader("mmu.NIB_6", 1, length),
        new ColumnHeader("mmu.NIB_7", 1, length),
        new ColumnHeader("mmu.NIB_8", 1, length),
        new ColumnHeader("mmu.NIB_9", 1, length),
        new ColumnHeader("mmu.OFF_1_LO", 32, length),
        new ColumnHeader("mmu.OFF_2_HI", 32, length),
        new ColumnHeader("mmu.OFF_2_LO", 32, length),
        new ColumnHeader("mmu.OFFSET_OUT_OF_BOUNDS", 1, length),
        new ColumnHeader("mmu.PRECOMPUTATION", 32, length),
        new ColumnHeader("mmu.RAM_STAMP", 32, length),
        new ColumnHeader("mmu.REFO", 32, length),
        new ColumnHeader("mmu.REFS", 32, length),
        new ColumnHeader("mmu.RETURN_CAPACITY", 32, length),
        new ColumnHeader("mmu.RETURN_OFFSET", 32, length),
        new ColumnHeader("mmu.RETURNER", 32, length),
        new ColumnHeader("mmu.SIZE", 32, length),
        new ColumnHeader("mmu.SIZE_IMPORTED", 32, length),
        new ColumnHeader("mmu.SOURCE_BYTE_OFFSET", 32, length),
        new ColumnHeader("mmu.SOURCE_LIMB_OFFSET", 32, length),
        new ColumnHeader("mmu.TARGET_BYTE_OFFSET", 32, length),
        new ColumnHeader("mmu.TARGET_LIMB_OFFSET", 32, length),
        new ColumnHeader("mmu.TERNARY", 32, length),
        new ColumnHeader("mmu.TO_RAM", 1, length),
        new ColumnHeader("mmu.TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS", 32, length),
        new ColumnHeader("mmu.TOTAL_NUMBER_OF_PADDINGS", 32, length),
        new ColumnHeader("mmu.TOTAL_NUMBER_OF_READS", 32, length),
        new ColumnHeader("mmu.VAL_HI", 32, length),
        new ColumnHeader("mmu.VAL_LO", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.acc5 = buffers.get(4);
    this.acc6 = buffers.get(5);
    this.acc7 = buffers.get(6);
    this.acc8 = buffers.get(7);
    this.aligned = buffers.get(8);
    this.bit1 = buffers.get(9);
    this.bit2 = buffers.get(10);
    this.bit3 = buffers.get(11);
    this.bit4 = buffers.get(12);
    this.bit5 = buffers.get(13);
    this.bit6 = buffers.get(14);
    this.bit7 = buffers.get(15);
    this.bit8 = buffers.get(16);
    this.byte1 = buffers.get(17);
    this.byte2 = buffers.get(18);
    this.byte3 = buffers.get(19);
    this.byte4 = buffers.get(20);
    this.byte5 = buffers.get(21);
    this.byte6 = buffers.get(22);
    this.byte7 = buffers.get(23);
    this.byte8 = buffers.get(24);
    this.callDataOffset = buffers.get(25);
    this.callDataSize = buffers.get(26);
    this.callStackDepth = buffers.get(27);
    this.caller = buffers.get(28);
    this.contextNumber = buffers.get(29);
    this.contextSource = buffers.get(30);
    this.contextTarget = buffers.get(31);
    this.counter = buffers.get(32);
    this.erf = buffers.get(33);
    this.exoIsHash = buffers.get(34);
    this.exoIsLog = buffers.get(35);
    this.exoIsRom = buffers.get(36);
    this.exoIsTxcd = buffers.get(37);
    this.fast = buffers.get(38);
    this.info = buffers.get(39);
    this.instruction = buffers.get(40);
    this.isData = buffers.get(41);
    this.isMicroInstruction = buffers.get(42);
    this.microInstruction = buffers.get(43);
    this.microInstructionStamp = buffers.get(44);
    this.min = buffers.get(45);
    this.nib1 = buffers.get(46);
    this.nib2 = buffers.get(47);
    this.nib3 = buffers.get(48);
    this.nib4 = buffers.get(49);
    this.nib5 = buffers.get(50);
    this.nib6 = buffers.get(51);
    this.nib7 = buffers.get(52);
    this.nib8 = buffers.get(53);
    this.nib9 = buffers.get(54);
    this.off1Lo = buffers.get(55);
    this.off2Hi = buffers.get(56);
    this.off2Lo = buffers.get(57);
    this.offsetOutOfBounds = buffers.get(58);
    this.precomputation = buffers.get(59);
    this.ramStamp = buffers.get(60);
    this.refo = buffers.get(61);
    this.refs = buffers.get(62);
    this.returnCapacity = buffers.get(63);
    this.returnOffset = buffers.get(64);
    this.returner = buffers.get(65);
    this.size = buffers.get(66);
    this.sizeImported = buffers.get(67);
    this.sourceByteOffset = buffers.get(68);
    this.sourceLimbOffset = buffers.get(69);
    this.targetByteOffset = buffers.get(70);
    this.targetLimbOffset = buffers.get(71);
    this.ternary = buffers.get(72);
    this.toRam = buffers.get(73);
    this.totalNumberOfMicroInstructions = buffers.get(74);
    this.totalNumberOfPaddings = buffers.get(75);
    this.totalNumberOfReads = buffers.get(76);
    this.valHi = buffers.get(77);
    this.valLo = buffers.get(78);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final BigInteger b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mmu.ACC_1 already set");
    } else {
      filled.set(0);
    }

    acc1.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc2(final BigInteger b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mmu.ACC_2 already set");
    } else {
      filled.set(1);
    }

    acc2.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc3(final BigInteger b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mmu.ACC_3 already set");
    } else {
      filled.set(2);
    }

    acc3.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc4(final BigInteger b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mmu.ACC_4 already set");
    } else {
      filled.set(3);
    }

    acc4.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc5(final BigInteger b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mmu.ACC_5 already set");
    } else {
      filled.set(4);
    }

    acc5.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc6(final BigInteger b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mmu.ACC_6 already set");
    } else {
      filled.set(5);
    }

    acc6.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc7(final BigInteger b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mmu.ACC_7 already set");
    } else {
      filled.set(6);
    }

    acc7.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc8(final BigInteger b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mmu.ACC_8 already set");
    } else {
      filled.set(7);
    }

    acc8.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace aligned(final BigInteger b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mmu.ALIGNED already set");
    } else {
      filled.set(8);
    }

    aligned.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mmu.BIT_1 already set");
    } else {
      filled.set(9);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mmu.BIT_2 already set");
    } else {
      filled.set(10);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mmu.BIT_3 already set");
    } else {
      filled.set(11);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mmu.BIT_4 already set");
    } else {
      filled.set(12);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit5(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mmu.BIT_5 already set");
    } else {
      filled.set(13);
    }

    bit5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit6(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mmu.BIT_6 already set");
    } else {
      filled.set(14);
    }

    bit6.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit7(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mmu.BIT_7 already set");
    } else {
      filled.set(15);
    }

    bit7.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit8(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mmu.BIT_8 already set");
    } else {
      filled.set(16);
    }

    bit8.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mmu.BYTE_1 already set");
    } else {
      filled.set(17);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mmu.BYTE_2 already set");
    } else {
      filled.set(18);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mmu.BYTE_3 already set");
    } else {
      filled.set(19);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mmu.BYTE_4 already set");
    } else {
      filled.set(20);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mmu.BYTE_5 already set");
    } else {
      filled.set(21);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace byte6(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mmu.BYTE_6 already set");
    } else {
      filled.set(22);
    }

    byte6.put(b.toByte());

    return this;
  }

  public Trace byte7(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mmu.BYTE_7 already set");
    } else {
      filled.set(23);
    }

    byte7.put(b.toByte());

    return this;
  }

  public Trace byte8(final UnsignedByte b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mmu.BYTE_8 already set");
    } else {
      filled.set(24);
    }

    byte8.put(b.toByte());

    return this;
  }

  public Trace callDataOffset(final BigInteger b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mmu.CALL_DATA_OFFSET already set");
    } else {
      filled.set(26);
    }

    callDataOffset.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace callDataSize(final BigInteger b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mmu.CALL_DATA_SIZE already set");
    } else {
      filled.set(27);
    }

    callDataSize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace callStackDepth(final BigInteger b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mmu.CALL_STACK_DEPTH already set");
    } else {
      filled.set(28);
    }

    callStackDepth.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace caller(final BigInteger b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mmu.CALLER already set");
    } else {
      filled.set(25);
    }

    caller.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace contextNumber(final BigInteger b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mmu.CONTEXT_NUMBER already set");
    } else {
      filled.set(29);
    }

    contextNumber.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace contextSource(final BigInteger b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mmu.CONTEXT_SOURCE already set");
    } else {
      filled.set(30);
    }

    contextSource.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace contextTarget(final BigInteger b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mmu.CONTEXT_TARGET already set");
    } else {
      filled.set(31);
    }

    contextTarget.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace counter(final BigInteger b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mmu.COUNTER already set");
    } else {
      filled.set(32);
    }

    counter.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace erf(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mmu.ERF already set");
    } else {
      filled.set(33);
    }

    erf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsHash(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mmu.EXO_IS_HASH already set");
    } else {
      filled.set(34);
    }

    exoIsHash.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsLog(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mmu.EXO_IS_LOG already set");
    } else {
      filled.set(35);
    }

    exoIsLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsRom(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mmu.EXO_IS_ROM already set");
    } else {
      filled.set(36);
    }

    exoIsRom.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsTxcd(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mmu.EXO_IS_TXCD already set");
    } else {
      filled.set(37);
    }

    exoIsTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace fast(final BigInteger b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mmu.FAST already set");
    } else {
      filled.set(38);
    }

    fast.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace info(final BigInteger b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mmu.INFO already set");
    } else {
      filled.set(39);
    }

    info.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace instruction(final BigInteger b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mmu.INSTRUCTION already set");
    } else {
      filled.set(40);
    }

    instruction.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace isData(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mmu.IS_DATA already set");
    } else {
      filled.set(41);
    }

    isData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isMicroInstruction(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mmu.IS_MICRO_INSTRUCTION already set");
    } else {
      filled.set(42);
    }

    isMicroInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace microInstruction(final BigInteger b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mmu.MICRO_INSTRUCTION already set");
    } else {
      filled.set(43);
    }

    microInstruction.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace microInstructionStamp(final BigInteger b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mmu.MICRO_INSTRUCTION_STAMP already set");
    } else {
      filled.set(44);
    }

    microInstructionStamp.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace min(final BigInteger b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mmu.MIN already set");
    } else {
      filled.set(45);
    }

    min.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace nib1(final UnsignedByte b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mmu.NIB_1 already set");
    } else {
      filled.set(46);
    }

    nib1.put(b.toByte());

    return this;
  }

  public Trace nib2(final UnsignedByte b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.NIB_2 already set");
    } else {
      filled.set(47);
    }

    nib2.put(b.toByte());

    return this;
  }

  public Trace nib3(final UnsignedByte b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmu.NIB_3 already set");
    } else {
      filled.set(48);
    }

    nib3.put(b.toByte());

    return this;
  }

  public Trace nib4(final UnsignedByte b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmu.NIB_4 already set");
    } else {
      filled.set(49);
    }

    nib4.put(b.toByte());

    return this;
  }

  public Trace nib5(final UnsignedByte b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mmu.NIB_5 already set");
    } else {
      filled.set(50);
    }

    nib5.put(b.toByte());

    return this;
  }

  public Trace nib6(final UnsignedByte b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mmu.NIB_6 already set");
    } else {
      filled.set(51);
    }

    nib6.put(b.toByte());

    return this;
  }

  public Trace nib7(final UnsignedByte b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmu.NIB_7 already set");
    } else {
      filled.set(52);
    }

    nib7.put(b.toByte());

    return this;
  }

  public Trace nib8(final UnsignedByte b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmu.NIB_8 already set");
    } else {
      filled.set(53);
    }

    nib8.put(b.toByte());

    return this;
  }

  public Trace nib9(final UnsignedByte b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mmu.NIB_9 already set");
    } else {
      filled.set(54);
    }

    nib9.put(b.toByte());

    return this;
  }

  public Trace off1Lo(final BigInteger b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmu.OFF_1_LO already set");
    } else {
      filled.set(56);
    }

    off1Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace off2Hi(final BigInteger b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mmu.OFF_2_HI already set");
    } else {
      filled.set(57);
    }

    off2Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace off2Lo(final BigInteger b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mmu.OFF_2_LO already set");
    } else {
      filled.set(58);
    }

    off2Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace offsetOutOfBounds(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mmu.OFFSET_OUT_OF_BOUNDS already set");
    } else {
      filled.set(55);
    }

    offsetOutOfBounds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace precomputation(final BigInteger b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmu.PRECOMPUTATION already set");
    } else {
      filled.set(59);
    }

    precomputation.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace ramStamp(final BigInteger b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mmu.RAM_STAMP already set");
    } else {
      filled.set(60);
    }

    ramStamp.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace refo(final BigInteger b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mmu.REFO already set");
    } else {
      filled.set(61);
    }

    refo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace refs(final BigInteger b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmu.REFS already set");
    } else {
      filled.set(62);
    }

    refs.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace returnCapacity(final BigInteger b) {
    if (filled.get(64)) {
      throw new IllegalStateException("mmu.RETURN_CAPACITY already set");
    } else {
      filled.set(64);
    }

    returnCapacity.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace returnOffset(final BigInteger b) {
    if (filled.get(65)) {
      throw new IllegalStateException("mmu.RETURN_OFFSET already set");
    } else {
      filled.set(65);
    }

    returnOffset.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace returner(final BigInteger b) {
    if (filled.get(63)) {
      throw new IllegalStateException("mmu.RETURNER already set");
    } else {
      filled.set(63);
    }

    returner.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace size(final BigInteger b) {
    if (filled.get(66)) {
      throw new IllegalStateException("mmu.SIZE already set");
    } else {
      filled.set(66);
    }

    size.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace sizeImported(final BigInteger b) {
    if (filled.get(67)) {
      throw new IllegalStateException("mmu.SIZE_IMPORTED already set");
    } else {
      filled.set(67);
    }

    sizeImported.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace sourceByteOffset(final BigInteger b) {
    if (filled.get(68)) {
      throw new IllegalStateException("mmu.SOURCE_BYTE_OFFSET already set");
    } else {
      filled.set(68);
    }

    sourceByteOffset.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace sourceLimbOffset(final BigInteger b) {
    if (filled.get(69)) {
      throw new IllegalStateException("mmu.SOURCE_LIMB_OFFSET already set");
    } else {
      filled.set(69);
    }

    sourceLimbOffset.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace targetByteOffset(final BigInteger b) {
    if (filled.get(70)) {
      throw new IllegalStateException("mmu.TARGET_BYTE_OFFSET already set");
    } else {
      filled.set(70);
    }

    targetByteOffset.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace targetLimbOffset(final BigInteger b) {
    if (filled.get(71)) {
      throw new IllegalStateException("mmu.TARGET_LIMB_OFFSET already set");
    } else {
      filled.set(71);
    }

    targetLimbOffset.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace ternary(final BigInteger b) {
    if (filled.get(72)) {
      throw new IllegalStateException("mmu.TERNARY already set");
    } else {
      filled.set(72);
    }

    ternary.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace toRam(final Boolean b) {
    if (filled.get(76)) {
      throw new IllegalStateException("mmu.TO_RAM already set");
    } else {
      filled.set(76);
    }

    toRam.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace totalNumberOfMicroInstructions(final BigInteger b) {
    if (filled.get(73)) {
      throw new IllegalStateException("mmu.TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS already set");
    } else {
      filled.set(73);
    }

    totalNumberOfMicroInstructions.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace totalNumberOfPaddings(final BigInteger b) {
    if (filled.get(74)) {
      throw new IllegalStateException("mmu.TOTAL_NUMBER_OF_PADDINGS already set");
    } else {
      filled.set(74);
    }

    totalNumberOfPaddings.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace totalNumberOfReads(final BigInteger b) {
    if (filled.get(75)) {
      throw new IllegalStateException("mmu.TOTAL_NUMBER_OF_READS already set");
    } else {
      filled.set(75);
    }

    totalNumberOfReads.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace valHi(final BigInteger b) {
    if (filled.get(77)) {
      throw new IllegalStateException("mmu.VAL_HI already set");
    } else {
      filled.set(77);
    }

    valHi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace valLo(final BigInteger b) {
    if (filled.get(78)) {
      throw new IllegalStateException("mmu.VAL_LO already set");
    } else {
      filled.set(78);
    }

    valLo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mmu.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mmu.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mmu.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mmu.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mmu.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mmu.ACC_6 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mmu.ACC_7 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mmu.ACC_8 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mmu.ALIGNED has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mmu.BIT_1 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mmu.BIT_2 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mmu.BIT_3 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mmu.BIT_4 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mmu.BIT_5 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mmu.BIT_6 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mmu.BIT_7 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mmu.BIT_8 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mmu.BYTE_1 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mmu.BYTE_2 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mmu.BYTE_3 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mmu.BYTE_4 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mmu.BYTE_5 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mmu.BYTE_6 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mmu.BYTE_7 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mmu.BYTE_8 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mmu.CALL_DATA_OFFSET has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mmu.CALL_DATA_SIZE has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mmu.CALL_STACK_DEPTH has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mmu.CALLER has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mmu.CONTEXT_NUMBER has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mmu.CONTEXT_SOURCE has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mmu.CONTEXT_TARGET has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mmu.COUNTER has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mmu.ERF has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mmu.EXO_IS_HASH has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mmu.EXO_IS_LOG has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mmu.EXO_IS_ROM has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mmu.EXO_IS_TXCD has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mmu.FAST has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mmu.INFO has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mmu.INSTRUCTION has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mmu.IS_DATA has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mmu.IS_MICRO_INSTRUCTION has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mmu.MICRO_INSTRUCTION has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mmu.MICRO_INSTRUCTION_STAMP has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mmu.MIN has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mmu.NIB_1 has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mmu.NIB_2 has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mmu.NIB_3 has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mmu.NIB_4 has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mmu.NIB_5 has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mmu.NIB_6 has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("mmu.NIB_7 has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("mmu.NIB_8 has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mmu.NIB_9 has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mmu.OFF_1_LO has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mmu.OFF_2_HI has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mmu.OFF_2_LO has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mmu.OFFSET_OUT_OF_BOUNDS has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("mmu.PRECOMPUTATION has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mmu.RAM_STAMP has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("mmu.REFO has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("mmu.REFS has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("mmu.RETURN_CAPACITY has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("mmu.RETURN_OFFSET has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("mmu.RETURNER has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException("mmu.SIZE has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException("mmu.SIZE_IMPORTED has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException("mmu.SOURCE_BYTE_OFFSET has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException("mmu.SOURCE_LIMB_OFFSET has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException("mmu.TARGET_BYTE_OFFSET has not been filled");
    }

    if (!filled.get(71)) {
      throw new IllegalStateException("mmu.TARGET_LIMB_OFFSET has not been filled");
    }

    if (!filled.get(72)) {
      throw new IllegalStateException("mmu.TERNARY has not been filled");
    }

    if (!filled.get(76)) {
      throw new IllegalStateException("mmu.TO_RAM has not been filled");
    }

    if (!filled.get(73)) {
      throw new IllegalStateException("mmu.TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS has not been filled");
    }

    if (!filled.get(74)) {
      throw new IllegalStateException("mmu.TOTAL_NUMBER_OF_PADDINGS has not been filled");
    }

    if (!filled.get(75)) {
      throw new IllegalStateException("mmu.TOTAL_NUMBER_OF_READS has not been filled");
    }

    if (!filled.get(77)) {
      throw new IllegalStateException("mmu.VAL_HI has not been filled");
    }

    if (!filled.get(78)) {
      throw new IllegalStateException("mmu.VAL_LO has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc1.position(acc1.position() + 32);
    }

    if (!filled.get(1)) {
      acc2.position(acc2.position() + 32);
    }

    if (!filled.get(2)) {
      acc3.position(acc3.position() + 32);
    }

    if (!filled.get(3)) {
      acc4.position(acc4.position() + 32);
    }

    if (!filled.get(4)) {
      acc5.position(acc5.position() + 32);
    }

    if (!filled.get(5)) {
      acc6.position(acc6.position() + 32);
    }

    if (!filled.get(6)) {
      acc7.position(acc7.position() + 32);
    }

    if (!filled.get(7)) {
      acc8.position(acc8.position() + 32);
    }

    if (!filled.get(8)) {
      aligned.position(aligned.position() + 32);
    }

    if (!filled.get(9)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(10)) {
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(11)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(12)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(13)) {
      bit5.position(bit5.position() + 1);
    }

    if (!filled.get(14)) {
      bit6.position(bit6.position() + 1);
    }

    if (!filled.get(15)) {
      bit7.position(bit7.position() + 1);
    }

    if (!filled.get(16)) {
      bit8.position(bit8.position() + 1);
    }

    if (!filled.get(17)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(18)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(19)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(20)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(21)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(22)) {
      byte6.position(byte6.position() + 1);
    }

    if (!filled.get(23)) {
      byte7.position(byte7.position() + 1);
    }

    if (!filled.get(24)) {
      byte8.position(byte8.position() + 1);
    }

    if (!filled.get(26)) {
      callDataOffset.position(callDataOffset.position() + 32);
    }

    if (!filled.get(27)) {
      callDataSize.position(callDataSize.position() + 32);
    }

    if (!filled.get(28)) {
      callStackDepth.position(callStackDepth.position() + 32);
    }

    if (!filled.get(25)) {
      caller.position(caller.position() + 32);
    }

    if (!filled.get(29)) {
      contextNumber.position(contextNumber.position() + 32);
    }

    if (!filled.get(30)) {
      contextSource.position(contextSource.position() + 32);
    }

    if (!filled.get(31)) {
      contextTarget.position(contextTarget.position() + 32);
    }

    if (!filled.get(32)) {
      counter.position(counter.position() + 32);
    }

    if (!filled.get(33)) {
      erf.position(erf.position() + 1);
    }

    if (!filled.get(34)) {
      exoIsHash.position(exoIsHash.position() + 1);
    }

    if (!filled.get(35)) {
      exoIsLog.position(exoIsLog.position() + 1);
    }

    if (!filled.get(36)) {
      exoIsRom.position(exoIsRom.position() + 1);
    }

    if (!filled.get(37)) {
      exoIsTxcd.position(exoIsTxcd.position() + 1);
    }

    if (!filled.get(38)) {
      fast.position(fast.position() + 32);
    }

    if (!filled.get(39)) {
      info.position(info.position() + 32);
    }

    if (!filled.get(40)) {
      instruction.position(instruction.position() + 32);
    }

    if (!filled.get(41)) {
      isData.position(isData.position() + 1);
    }

    if (!filled.get(42)) {
      isMicroInstruction.position(isMicroInstruction.position() + 1);
    }

    if (!filled.get(43)) {
      microInstruction.position(microInstruction.position() + 32);
    }

    if (!filled.get(44)) {
      microInstructionStamp.position(microInstructionStamp.position() + 32);
    }

    if (!filled.get(45)) {
      min.position(min.position() + 32);
    }

    if (!filled.get(46)) {
      nib1.position(nib1.position() + 1);
    }

    if (!filled.get(47)) {
      nib2.position(nib2.position() + 1);
    }

    if (!filled.get(48)) {
      nib3.position(nib3.position() + 1);
    }

    if (!filled.get(49)) {
      nib4.position(nib4.position() + 1);
    }

    if (!filled.get(50)) {
      nib5.position(nib5.position() + 1);
    }

    if (!filled.get(51)) {
      nib6.position(nib6.position() + 1);
    }

    if (!filled.get(52)) {
      nib7.position(nib7.position() + 1);
    }

    if (!filled.get(53)) {
      nib8.position(nib8.position() + 1);
    }

    if (!filled.get(54)) {
      nib9.position(nib9.position() + 1);
    }

    if (!filled.get(56)) {
      off1Lo.position(off1Lo.position() + 32);
    }

    if (!filled.get(57)) {
      off2Hi.position(off2Hi.position() + 32);
    }

    if (!filled.get(58)) {
      off2Lo.position(off2Lo.position() + 32);
    }

    if (!filled.get(55)) {
      offsetOutOfBounds.position(offsetOutOfBounds.position() + 1);
    }

    if (!filled.get(59)) {
      precomputation.position(precomputation.position() + 32);
    }

    if (!filled.get(60)) {
      ramStamp.position(ramStamp.position() + 32);
    }

    if (!filled.get(61)) {
      refo.position(refo.position() + 32);
    }

    if (!filled.get(62)) {
      refs.position(refs.position() + 32);
    }

    if (!filled.get(64)) {
      returnCapacity.position(returnCapacity.position() + 32);
    }

    if (!filled.get(65)) {
      returnOffset.position(returnOffset.position() + 32);
    }

    if (!filled.get(63)) {
      returner.position(returner.position() + 32);
    }

    if (!filled.get(66)) {
      size.position(size.position() + 32);
    }

    if (!filled.get(67)) {
      sizeImported.position(sizeImported.position() + 32);
    }

    if (!filled.get(68)) {
      sourceByteOffset.position(sourceByteOffset.position() + 32);
    }

    if (!filled.get(69)) {
      sourceLimbOffset.position(sourceLimbOffset.position() + 32);
    }

    if (!filled.get(70)) {
      targetByteOffset.position(targetByteOffset.position() + 32);
    }

    if (!filled.get(71)) {
      targetLimbOffset.position(targetLimbOffset.position() + 32);
    }

    if (!filled.get(72)) {
      ternary.position(ternary.position() + 32);
    }

    if (!filled.get(76)) {
      toRam.position(toRam.position() + 1);
    }

    if (!filled.get(73)) {
      totalNumberOfMicroInstructions.position(totalNumberOfMicroInstructions.position() + 32);
    }

    if (!filled.get(74)) {
      totalNumberOfPaddings.position(totalNumberOfPaddings.position() + 32);
    }

    if (!filled.get(75)) {
      totalNumberOfReads.position(totalNumberOfReads.position() + 32);
    }

    if (!filled.get(77)) {
      valHi.position(valHi.position() + 32);
    }

    if (!filled.get(78)) {
      valLo.position(valLo.position() + 32);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
    return null;
  }
}
