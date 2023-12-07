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

package net.consensys.linea.zktracer.module.bin;

import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

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
  private final MappedByteBuffer andByteHi;
  private final MappedByteBuffer andByteLo;
  private final MappedByteBuffer argument1Hi;
  private final MappedByteBuffer argument1Lo;
  private final MappedByteBuffer argument2Hi;
  private final MappedByteBuffer argument2Lo;
  private final MappedByteBuffer binaryStamp;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bitB4;
  private final MappedByteBuffer bits;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer byte5;
  private final MappedByteBuffer byte6;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isData;
  private final MappedByteBuffer low4;
  private final MappedByteBuffer neg;
  private final MappedByteBuffer notByteHi;
  private final MappedByteBuffer notByteLo;
  private final MappedByteBuffer oneLineInstruction;
  private final MappedByteBuffer orByteHi;
  private final MappedByteBuffer orByteLo;
  private final MappedByteBuffer pivot;
  private final MappedByteBuffer resultHi;
  private final MappedByteBuffer resultLo;
  private final MappedByteBuffer small;
  private final MappedByteBuffer xorByteHi;
  private final MappedByteBuffer xorByteLo;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("bin.ACC_1", 32, length),
        new ColumnHeader("bin.ACC_2", 32, length),
        new ColumnHeader("bin.ACC_3", 32, length),
        new ColumnHeader("bin.ACC_4", 32, length),
        new ColumnHeader("bin.ACC_5", 32, length),
        new ColumnHeader("bin.ACC_6", 32, length),
        new ColumnHeader("bin.AND_BYTE_HI", 32, length),
        new ColumnHeader("bin.AND_BYTE_LO", 32, length),
        new ColumnHeader("bin.ARGUMENT_1_HI", 32, length),
        new ColumnHeader("bin.ARGUMENT_1_LO", 32, length),
        new ColumnHeader("bin.ARGUMENT_2_HI", 32, length),
        new ColumnHeader("bin.ARGUMENT_2_LO", 32, length),
        new ColumnHeader("bin.BINARY_STAMP", 32, length),
        new ColumnHeader("bin.BIT_1", 1, length),
        new ColumnHeader("bin.BIT_B_4", 1, length),
        new ColumnHeader("bin.BITS", 1, length),
        new ColumnHeader("bin.BYTE_1", 1, length),
        new ColumnHeader("bin.BYTE_2", 1, length),
        new ColumnHeader("bin.BYTE_3", 1, length),
        new ColumnHeader("bin.BYTE_4", 1, length),
        new ColumnHeader("bin.BYTE_5", 1, length),
        new ColumnHeader("bin.BYTE_6", 1, length),
        new ColumnHeader("bin.COUNTER", 32, length),
        new ColumnHeader("bin.INST", 32, length),
        new ColumnHeader("bin.IS_DATA", 1, length),
        new ColumnHeader("bin.LOW_4", 32, length),
        new ColumnHeader("bin.NEG", 1, length),
        new ColumnHeader("bin.NOT_BYTE_HI", 32, length),
        new ColumnHeader("bin.NOT_BYTE_LO", 32, length),
        new ColumnHeader("bin.ONE_LINE_INSTRUCTION", 1, length),
        new ColumnHeader("bin.OR_BYTE_HI", 32, length),
        new ColumnHeader("bin.OR_BYTE_LO", 32, length),
        new ColumnHeader("bin.PIVOT", 32, length),
        new ColumnHeader("bin.RESULT_HI", 32, length),
        new ColumnHeader("bin.RESULT_LO", 32, length),
        new ColumnHeader("bin.SMALL", 1, length),
        new ColumnHeader("bin.XOR_BYTE_HI", 32, length),
        new ColumnHeader("bin.XOR_BYTE_LO", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.acc5 = buffers.get(4);
    this.acc6 = buffers.get(5);
    this.andByteHi = buffers.get(6);
    this.andByteLo = buffers.get(7);
    this.argument1Hi = buffers.get(8);
    this.argument1Lo = buffers.get(9);
    this.argument2Hi = buffers.get(10);
    this.argument2Lo = buffers.get(11);
    this.binaryStamp = buffers.get(12);
    this.bit1 = buffers.get(13);
    this.bitB4 = buffers.get(14);
    this.bits = buffers.get(15);
    this.byte1 = buffers.get(16);
    this.byte2 = buffers.get(17);
    this.byte3 = buffers.get(18);
    this.byte4 = buffers.get(19);
    this.byte5 = buffers.get(20);
    this.byte6 = buffers.get(21);
    this.counter = buffers.get(22);
    this.inst = buffers.get(23);
    this.isData = buffers.get(24);
    this.low4 = buffers.get(25);
    this.neg = buffers.get(26);
    this.notByteHi = buffers.get(27);
    this.notByteLo = buffers.get(28);
    this.oneLineInstruction = buffers.get(29);
    this.orByteHi = buffers.get(30);
    this.orByteLo = buffers.get(31);
    this.pivot = buffers.get(32);
    this.resultHi = buffers.get(33);
    this.resultLo = buffers.get(34);
    this.small = buffers.get(35);
    this.xorByteHi = buffers.get(36);
    this.xorByteLo = buffers.get(37);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("bin.ACC_1 already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc1.put((byte) 0);
    }
    acc1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("bin.ACC_2 already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc2.put((byte) 0);
    }
    acc2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("bin.ACC_3 already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc3.put((byte) 0);
    }
    acc3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("bin.ACC_4 already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc4.put((byte) 0);
    }
    acc4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc5(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("bin.ACC_5 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc5.put((byte) 0);
    }
    acc5.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc6(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("bin.ACC_6 already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc6.put((byte) 0);
    }
    acc6.put(b.toArrayUnsafe());

    return this;
  }

  public Trace andByteHi(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("bin.AND_BYTE_HI already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      andByteHi.put((byte) 0);
    }
    andByteHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace andByteLo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("bin.AND_BYTE_LO already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      andByteLo.put((byte) 0);
    }
    andByteLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument1Hi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_1_HI already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument1Hi.put((byte) 0);
    }
    argument1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument1Lo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument1Lo.put((byte) 0);
    }
    argument1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument2Hi(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument2Hi.put((byte) 0);
    }
    argument2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument2Lo(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument2Lo.put((byte) 0);
    }
    argument2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace binaryStamp(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("bin.BINARY_STAMP already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      binaryStamp.put((byte) 0);
    }
    binaryStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("bin.BIT_1 already set");
    } else {
      filled.set(14);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB4(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("bin.BIT_B_4 already set");
    } else {
      filled.set(15);
    }

    bitB4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("bin.BITS already set");
    } else {
      filled.set(13);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("bin.BYTE_1 already set");
    } else {
      filled.set(16);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("bin.BYTE_2 already set");
    } else {
      filled.set(17);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("bin.BYTE_3 already set");
    } else {
      filled.set(18);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("bin.BYTE_4 already set");
    } else {
      filled.set(19);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("bin.BYTE_5 already set");
    } else {
      filled.set(20);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace byte6(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("bin.BYTE_6 already set");
    } else {
      filled.set(21);
    }

    byte6.put(b.toByte());

    return this;
  }

  public Trace counter(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("bin.COUNTER already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counter.put((byte) 0);
    }
    counter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace inst(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("bin.INST already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      inst.put((byte) 0);
    }
    inst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isData(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("bin.IS_DATA already set");
    } else {
      filled.set(24);
    }

    isData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace low4(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("bin.LOW_4 already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      low4.put((byte) 0);
    }
    low4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace neg(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("bin.NEG already set");
    } else {
      filled.set(26);
    }

    neg.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace notByteHi(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("bin.NOT_BYTE_HI already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      notByteHi.put((byte) 0);
    }
    notByteHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace notByteLo(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("bin.NOT_BYTE_LO already set");
    } else {
      filled.set(28);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      notByteLo.put((byte) 0);
    }
    notByteLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace oneLineInstruction(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("bin.ONE_LINE_INSTRUCTION already set");
    } else {
      filled.set(29);
    }

    oneLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace orByteHi(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("bin.OR_BYTE_HI already set");
    } else {
      filled.set(30);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      orByteHi.put((byte) 0);
    }
    orByteHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace orByteLo(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("bin.OR_BYTE_LO already set");
    } else {
      filled.set(31);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      orByteLo.put((byte) 0);
    }
    orByteLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pivot(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("bin.PIVOT already set");
    } else {
      filled.set(32);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pivot.put((byte) 0);
    }
    pivot.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resultHi(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("bin.RESULT_HI already set");
    } else {
      filled.set(33);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultHi.put((byte) 0);
    }
    resultHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resultLo(final Bytes b) {
    if (filled.get(34)) {
      throw new IllegalStateException("bin.RESULT_LO already set");
    } else {
      filled.set(34);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultLo.put((byte) 0);
    }
    resultLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace small(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("bin.SMALL already set");
    } else {
      filled.set(35);
    }

    small.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace xorByteHi(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("bin.XOR_BYTE_HI already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      xorByteHi.put((byte) 0);
    }
    xorByteHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace xorByteLo(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("bin.XOR_BYTE_LO already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      xorByteLo.put((byte) 0);
    }
    xorByteLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("bin.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("bin.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("bin.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("bin.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("bin.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("bin.ACC_6 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("bin.AND_BYTE_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("bin.AND_BYTE_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_1_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("bin.BINARY_STAMP has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("bin.BIT_1 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("bin.BIT_B_4 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("bin.BITS has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("bin.BYTE_1 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("bin.BYTE_2 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("bin.BYTE_3 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("bin.BYTE_4 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("bin.BYTE_5 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("bin.BYTE_6 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("bin.COUNTER has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("bin.INST has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("bin.IS_DATA has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("bin.LOW_4 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("bin.NEG has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("bin.NOT_BYTE_HI has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("bin.NOT_BYTE_LO has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("bin.ONE_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("bin.OR_BYTE_HI has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("bin.OR_BYTE_LO has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("bin.PIVOT has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("bin.RESULT_HI has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("bin.RESULT_LO has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("bin.SMALL has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("bin.XOR_BYTE_HI has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("bin.XOR_BYTE_LO has not been filled");
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
      andByteHi.position(andByteHi.position() + 32);
    }

    if (!filled.get(7)) {
      andByteLo.position(andByteLo.position() + 32);
    }

    if (!filled.get(8)) {
      argument1Hi.position(argument1Hi.position() + 32);
    }

    if (!filled.get(9)) {
      argument1Lo.position(argument1Lo.position() + 32);
    }

    if (!filled.get(10)) {
      argument2Hi.position(argument2Hi.position() + 32);
    }

    if (!filled.get(11)) {
      argument2Lo.position(argument2Lo.position() + 32);
    }

    if (!filled.get(12)) {
      binaryStamp.position(binaryStamp.position() + 32);
    }

    if (!filled.get(14)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(15)) {
      bitB4.position(bitB4.position() + 1);
    }

    if (!filled.get(13)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(16)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(17)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(18)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(19)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(20)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(21)) {
      byte6.position(byte6.position() + 1);
    }

    if (!filled.get(22)) {
      counter.position(counter.position() + 32);
    }

    if (!filled.get(23)) {
      inst.position(inst.position() + 32);
    }

    if (!filled.get(24)) {
      isData.position(isData.position() + 1);
    }

    if (!filled.get(25)) {
      low4.position(low4.position() + 32);
    }

    if (!filled.get(26)) {
      neg.position(neg.position() + 1);
    }

    if (!filled.get(27)) {
      notByteHi.position(notByteHi.position() + 32);
    }

    if (!filled.get(28)) {
      notByteLo.position(notByteLo.position() + 32);
    }

    if (!filled.get(29)) {
      oneLineInstruction.position(oneLineInstruction.position() + 1);
    }

    if (!filled.get(30)) {
      orByteHi.position(orByteHi.position() + 32);
    }

    if (!filled.get(31)) {
      orByteLo.position(orByteLo.position() + 32);
    }

    if (!filled.get(32)) {
      pivot.position(pivot.position() + 32);
    }

    if (!filled.get(33)) {
      resultHi.position(resultHi.position() + 32);
    }

    if (!filled.get(34)) {
      resultLo.position(resultLo.position() + 32);
    }

    if (!filled.get(35)) {
      small.position(small.position() + 1);
    }

    if (!filled.get(36)) {
      xorByteHi.position(xorByteHi.position() + 32);
    }

    if (!filled.get(37)) {
      xorByteLo.position(xorByteLo.position() + 32);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public void build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
  }
}
