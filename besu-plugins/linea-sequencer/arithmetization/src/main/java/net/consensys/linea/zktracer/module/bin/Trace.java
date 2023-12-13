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
  private final MappedByteBuffer argument1Hi;
  private final MappedByteBuffer argument1Lo;
  private final MappedByteBuffer argument2Hi;
  private final MappedByteBuffer argument2Lo;
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
  private final MappedByteBuffer isAnd;
  private final MappedByteBuffer isByte;
  private final MappedByteBuffer isNot;
  private final MappedByteBuffer isOr;
  private final MappedByteBuffer isSignextend;
  private final MappedByteBuffer isXor;
  private final MappedByteBuffer low4;
  private final MappedByteBuffer mli;
  private final MappedByteBuffer neg;
  private final MappedByteBuffer oneLineInstruction;
  private final MappedByteBuffer pivot;
  private final MappedByteBuffer resultHi;
  private final MappedByteBuffer resultLo;
  private final MappedByteBuffer small;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer xxxByteHi;
  private final MappedByteBuffer xxxByteLo;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("bin.ACC_1", 32, length),
        new ColumnHeader("bin.ACC_2", 32, length),
        new ColumnHeader("bin.ACC_3", 32, length),
        new ColumnHeader("bin.ACC_4", 32, length),
        new ColumnHeader("bin.ACC_5", 32, length),
        new ColumnHeader("bin.ACC_6", 32, length),
        new ColumnHeader("bin.ARGUMENT_1_HI", 32, length),
        new ColumnHeader("bin.ARGUMENT_1_LO", 32, length),
        new ColumnHeader("bin.ARGUMENT_2_HI", 32, length),
        new ColumnHeader("bin.ARGUMENT_2_LO", 32, length),
        new ColumnHeader("bin.BIT_1", 1, length),
        new ColumnHeader("bin.BIT_B_4", 1, length),
        new ColumnHeader("bin.BITS", 1, length),
        new ColumnHeader("bin.BYTE_1", 1, length),
        new ColumnHeader("bin.BYTE_2", 1, length),
        new ColumnHeader("bin.BYTE_3", 1, length),
        new ColumnHeader("bin.BYTE_4", 1, length),
        new ColumnHeader("bin.BYTE_5", 1, length),
        new ColumnHeader("bin.BYTE_6", 1, length),
        new ColumnHeader("bin.COUNTER", 1, length),
        new ColumnHeader("bin.INST", 1, length),
        new ColumnHeader("bin.IS_AND", 1, length),
        new ColumnHeader("bin.IS_BYTE", 1, length),
        new ColumnHeader("bin.IS_NOT", 1, length),
        new ColumnHeader("bin.IS_OR", 1, length),
        new ColumnHeader("bin.IS_SIGNEXTEND", 1, length),
        new ColumnHeader("bin.IS_XOR", 1, length),
        new ColumnHeader("bin.LOW_4", 1, length),
        new ColumnHeader("bin.MLI", 1, length),
        new ColumnHeader("bin.NEG", 1, length),
        new ColumnHeader("bin.ONE_LINE_INSTRUCTION", 1, length),
        new ColumnHeader("bin.PIVOT", 1, length),
        new ColumnHeader("bin.RESULT_HI", 32, length),
        new ColumnHeader("bin.RESULT_LO", 32, length),
        new ColumnHeader("bin.SMALL", 1, length),
        new ColumnHeader("bin.STAMP", 32, length),
        new ColumnHeader("bin.XXX_BYTE_HI", 1, length),
        new ColumnHeader("bin.XXX_BYTE_LO", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.acc5 = buffers.get(4);
    this.acc6 = buffers.get(5);
    this.argument1Hi = buffers.get(6);
    this.argument1Lo = buffers.get(7);
    this.argument2Hi = buffers.get(8);
    this.argument2Lo = buffers.get(9);
    this.bit1 = buffers.get(10);
    this.bitB4 = buffers.get(11);
    this.bits = buffers.get(12);
    this.byte1 = buffers.get(13);
    this.byte2 = buffers.get(14);
    this.byte3 = buffers.get(15);
    this.byte4 = buffers.get(16);
    this.byte5 = buffers.get(17);
    this.byte6 = buffers.get(18);
    this.counter = buffers.get(19);
    this.inst = buffers.get(20);
    this.isAnd = buffers.get(21);
    this.isByte = buffers.get(22);
    this.isNot = buffers.get(23);
    this.isOr = buffers.get(24);
    this.isSignextend = buffers.get(25);
    this.isXor = buffers.get(26);
    this.low4 = buffers.get(27);
    this.mli = buffers.get(28);
    this.neg = buffers.get(29);
    this.oneLineInstruction = buffers.get(30);
    this.pivot = buffers.get(31);
    this.resultHi = buffers.get(32);
    this.resultLo = buffers.get(33);
    this.small = buffers.get(34);
    this.stamp = buffers.get(35);
    this.xxxByteHi = buffers.get(36);
    this.xxxByteLo = buffers.get(37);
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

  public Trace argument1Hi(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("bin.ARGUMENT_1_HI already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument1Hi.put((byte) 0);
    }
    argument1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument1Lo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument1Lo.put((byte) 0);
    }
    argument1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument2Hi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument2Hi.put((byte) 0);
    }
    argument2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument2Lo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument2Lo.put((byte) 0);
    }
    argument2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("bin.BIT_1 already set");
    } else {
      filled.set(11);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB4(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("bin.BIT_B_4 already set");
    } else {
      filled.set(12);
    }

    bitB4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("bin.BITS already set");
    } else {
      filled.set(10);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("bin.BYTE_1 already set");
    } else {
      filled.set(13);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("bin.BYTE_2 already set");
    } else {
      filled.set(14);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("bin.BYTE_3 already set");
    } else {
      filled.set(15);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("bin.BYTE_4 already set");
    } else {
      filled.set(16);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("bin.BYTE_5 already set");
    } else {
      filled.set(17);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace byte6(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("bin.BYTE_6 already set");
    } else {
      filled.set(18);
    }

    byte6.put(b.toByte());

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("bin.COUNTER already set");
    } else {
      filled.set(19);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("bin.INST already set");
    } else {
      filled.set(20);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isAnd(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("bin.IS_AND already set");
    } else {
      filled.set(21);
    }

    isAnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isByte(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("bin.IS_BYTE already set");
    } else {
      filled.set(22);
    }

    isByte.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isNot(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("bin.IS_NOT already set");
    } else {
      filled.set(23);
    }

    isNot.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isOr(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("bin.IS_OR already set");
    } else {
      filled.set(24);
    }

    isOr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSignextend(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("bin.IS_SIGNEXTEND already set");
    } else {
      filled.set(25);
    }

    isSignextend.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isXor(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("bin.IS_XOR already set");
    } else {
      filled.set(26);
    }

    isXor.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace low4(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("bin.LOW_4 already set");
    } else {
      filled.set(27);
    }

    low4.put(b.toByte());

    return this;
  }

  public Trace mli(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("bin.MLI already set");
    } else {
      filled.set(28);
    }

    mli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace neg(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("bin.NEG already set");
    } else {
      filled.set(29);
    }

    neg.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oneLineInstruction(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("bin.ONE_LINE_INSTRUCTION already set");
    } else {
      filled.set(30);
    }

    oneLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pivot(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("bin.PIVOT already set");
    } else {
      filled.set(31);
    }

    pivot.put(b.toByte());

    return this;
  }

  public Trace resultHi(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("bin.RESULT_HI already set");
    } else {
      filled.set(32);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultHi.put((byte) 0);
    }
    resultHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resultLo(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("bin.RESULT_LO already set");
    } else {
      filled.set(33);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultLo.put((byte) 0);
    }
    resultLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace small(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("bin.SMALL already set");
    } else {
      filled.set(34);
    }

    small.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("bin.STAMP already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace xxxByteHi(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("bin.XXX_BYTE_HI already set");
    } else {
      filled.set(36);
    }

    xxxByteHi.put(b.toByte());

    return this;
  }

  public Trace xxxByteLo(final UnsignedByte b) {
    if (filled.get(37)) {
      throw new IllegalStateException("bin.XXX_BYTE_LO already set");
    } else {
      filled.set(37);
    }

    xxxByteLo.put(b.toByte());

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
      throw new IllegalStateException("bin.ARGUMENT_1_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("bin.BIT_1 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("bin.BIT_B_4 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("bin.BITS has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("bin.BYTE_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("bin.BYTE_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("bin.BYTE_3 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("bin.BYTE_4 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("bin.BYTE_5 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("bin.BYTE_6 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("bin.COUNTER has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("bin.INST has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("bin.IS_AND has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("bin.IS_BYTE has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("bin.IS_NOT has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("bin.IS_OR has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("bin.IS_SIGNEXTEND has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("bin.IS_XOR has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("bin.LOW_4 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("bin.MLI has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("bin.NEG has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("bin.ONE_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("bin.PIVOT has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("bin.RESULT_HI has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("bin.RESULT_LO has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("bin.SMALL has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("bin.STAMP has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("bin.XXX_BYTE_HI has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("bin.XXX_BYTE_LO has not been filled");
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
      argument1Hi.position(argument1Hi.position() + 32);
    }

    if (!filled.get(7)) {
      argument1Lo.position(argument1Lo.position() + 32);
    }

    if (!filled.get(8)) {
      argument2Hi.position(argument2Hi.position() + 32);
    }

    if (!filled.get(9)) {
      argument2Lo.position(argument2Lo.position() + 32);
    }

    if (!filled.get(11)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(12)) {
      bitB4.position(bitB4.position() + 1);
    }

    if (!filled.get(10)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(13)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(14)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(15)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(16)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(17)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(18)) {
      byte6.position(byte6.position() + 1);
    }

    if (!filled.get(19)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(20)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(21)) {
      isAnd.position(isAnd.position() + 1);
    }

    if (!filled.get(22)) {
      isByte.position(isByte.position() + 1);
    }

    if (!filled.get(23)) {
      isNot.position(isNot.position() + 1);
    }

    if (!filled.get(24)) {
      isOr.position(isOr.position() + 1);
    }

    if (!filled.get(25)) {
      isSignextend.position(isSignextend.position() + 1);
    }

    if (!filled.get(26)) {
      isXor.position(isXor.position() + 1);
    }

    if (!filled.get(27)) {
      low4.position(low4.position() + 1);
    }

    if (!filled.get(28)) {
      mli.position(mli.position() + 1);
    }

    if (!filled.get(29)) {
      neg.position(neg.position() + 1);
    }

    if (!filled.get(30)) {
      oneLineInstruction.position(oneLineInstruction.position() + 1);
    }

    if (!filled.get(31)) {
      pivot.position(pivot.position() + 1);
    }

    if (!filled.get(32)) {
      resultHi.position(resultHi.position() + 32);
    }

    if (!filled.get(33)) {
      resultLo.position(resultLo.position() + 32);
    }

    if (!filled.get(34)) {
      small.position(small.position() + 1);
    }

    if (!filled.get(35)) {
      stamp.position(stamp.position() + 32);
    }

    if (!filled.get(36)) {
      xxxByteHi.position(xxxByteHi.position() + 1);
    }

    if (!filled.get(37)) {
      xxxByteLo.position(xxxByteLo.position() + 1);
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
