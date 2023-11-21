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

package net.consensys.linea.zktracer.module.shf;

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
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
  private final MappedByteBuffer bitB3;
  private final MappedByteBuffer bitB4;
  private final MappedByteBuffer bitB5;
  private final MappedByteBuffer bitB6;
  private final MappedByteBuffer bitB7;
  private final MappedByteBuffer bits;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer byte5;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isData;
  private final MappedByteBuffer known;
  private final MappedByteBuffer leftAlignedSuffixHigh;
  private final MappedByteBuffer leftAlignedSuffixLow;
  private final MappedByteBuffer low3;
  private final MappedByteBuffer microShiftParameter;
  private final MappedByteBuffer neg;
  private final MappedByteBuffer oneLineInstruction;
  private final MappedByteBuffer ones;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer rightAlignedPrefixHigh;
  private final MappedByteBuffer rightAlignedPrefixLow;
  private final MappedByteBuffer shb3Hi;
  private final MappedByteBuffer shb3Lo;
  private final MappedByteBuffer shb4Hi;
  private final MappedByteBuffer shb4Lo;
  private final MappedByteBuffer shb5Hi;
  private final MappedByteBuffer shb5Lo;
  private final MappedByteBuffer shb6Hi;
  private final MappedByteBuffer shb6Lo;
  private final MappedByteBuffer shb7Hi;
  private final MappedByteBuffer shb7Lo;
  private final MappedByteBuffer shiftDirection;
  private final MappedByteBuffer shiftStamp;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("shf.ACC_1", 32, length),
        new ColumnHeader("shf.ACC_2", 32, length),
        new ColumnHeader("shf.ACC_3", 32, length),
        new ColumnHeader("shf.ACC_4", 32, length),
        new ColumnHeader("shf.ACC_5", 32, length),
        new ColumnHeader("shf.ARG_1_HI", 32, length),
        new ColumnHeader("shf.ARG_1_LO", 32, length),
        new ColumnHeader("shf.ARG_2_HI", 32, length),
        new ColumnHeader("shf.ARG_2_LO", 32, length),
        new ColumnHeader("shf.BIT_1", 1, length),
        new ColumnHeader("shf.BIT_2", 1, length),
        new ColumnHeader("shf.BIT_3", 1, length),
        new ColumnHeader("shf.BIT_4", 1, length),
        new ColumnHeader("shf.BIT_B_3", 1, length),
        new ColumnHeader("shf.BIT_B_4", 1, length),
        new ColumnHeader("shf.BIT_B_5", 1, length),
        new ColumnHeader("shf.BIT_B_6", 1, length),
        new ColumnHeader("shf.BIT_B_7", 1, length),
        new ColumnHeader("shf.BITS", 1, length),
        new ColumnHeader("shf.BYTE_1", 1, length),
        new ColumnHeader("shf.BYTE_2", 1, length),
        new ColumnHeader("shf.BYTE_3", 1, length),
        new ColumnHeader("shf.BYTE_4", 1, length),
        new ColumnHeader("shf.BYTE_5", 1, length),
        new ColumnHeader("shf.COUNTER", 32, length),
        new ColumnHeader("shf.INST", 32, length),
        new ColumnHeader("shf.IS_DATA", 1, length),
        new ColumnHeader("shf.KNOWN", 1, length),
        new ColumnHeader("shf.LEFT_ALIGNED_SUFFIX_HIGH", 32, length),
        new ColumnHeader("shf.LEFT_ALIGNED_SUFFIX_LOW", 32, length),
        new ColumnHeader("shf.LOW_3", 32, length),
        new ColumnHeader("shf.MICRO_SHIFT_PARAMETER", 32, length),
        new ColumnHeader("shf.NEG", 1, length),
        new ColumnHeader("shf.ONE_LINE_INSTRUCTION", 1, length),
        new ColumnHeader("shf.ONES", 32, length),
        new ColumnHeader("shf.RES_HI", 32, length),
        new ColumnHeader("shf.RES_LO", 32, length),
        new ColumnHeader("shf.RIGHT_ALIGNED_PREFIX_HIGH", 32, length),
        new ColumnHeader("shf.RIGHT_ALIGNED_PREFIX_LOW", 32, length),
        new ColumnHeader("shf.SHB_3_HI", 32, length),
        new ColumnHeader("shf.SHB_3_LO", 32, length),
        new ColumnHeader("shf.SHB_4_HI", 32, length),
        new ColumnHeader("shf.SHB_4_LO", 32, length),
        new ColumnHeader("shf.SHB_5_HI", 32, length),
        new ColumnHeader("shf.SHB_5_LO", 32, length),
        new ColumnHeader("shf.SHB_6_HI", 32, length),
        new ColumnHeader("shf.SHB_6_LO", 32, length),
        new ColumnHeader("shf.SHB_7_HI", 32, length),
        new ColumnHeader("shf.SHB_7_LO", 32, length),
        new ColumnHeader("shf.SHIFT_DIRECTION", 1, length),
        new ColumnHeader("shf.SHIFT_STAMP", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.acc5 = buffers.get(4);
    this.arg1Hi = buffers.get(5);
    this.arg1Lo = buffers.get(6);
    this.arg2Hi = buffers.get(7);
    this.arg2Lo = buffers.get(8);
    this.bit1 = buffers.get(9);
    this.bit2 = buffers.get(10);
    this.bit3 = buffers.get(11);
    this.bit4 = buffers.get(12);
    this.bitB3 = buffers.get(13);
    this.bitB4 = buffers.get(14);
    this.bitB5 = buffers.get(15);
    this.bitB6 = buffers.get(16);
    this.bitB7 = buffers.get(17);
    this.bits = buffers.get(18);
    this.byte1 = buffers.get(19);
    this.byte2 = buffers.get(20);
    this.byte3 = buffers.get(21);
    this.byte4 = buffers.get(22);
    this.byte5 = buffers.get(23);
    this.counter = buffers.get(24);
    this.inst = buffers.get(25);
    this.isData = buffers.get(26);
    this.known = buffers.get(27);
    this.leftAlignedSuffixHigh = buffers.get(28);
    this.leftAlignedSuffixLow = buffers.get(29);
    this.low3 = buffers.get(30);
    this.microShiftParameter = buffers.get(31);
    this.neg = buffers.get(32);
    this.oneLineInstruction = buffers.get(33);
    this.ones = buffers.get(34);
    this.resHi = buffers.get(35);
    this.resLo = buffers.get(36);
    this.rightAlignedPrefixHigh = buffers.get(37);
    this.rightAlignedPrefixLow = buffers.get(38);
    this.shb3Hi = buffers.get(39);
    this.shb3Lo = buffers.get(40);
    this.shb4Hi = buffers.get(41);
    this.shb4Lo = buffers.get(42);
    this.shb5Hi = buffers.get(43);
    this.shb5Lo = buffers.get(44);
    this.shb6Hi = buffers.get(45);
    this.shb6Lo = buffers.get(46);
    this.shb7Hi = buffers.get(47);
    this.shb7Lo = buffers.get(48);
    this.shiftDirection = buffers.get(49);
    this.shiftStamp = buffers.get(50);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final BigInteger b) {
    if (filled.get(0)) {
      throw new IllegalStateException("shf.ACC_1 already set");
    } else {
      filled.set(0);
    }

    acc1.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc2(final BigInteger b) {
    if (filled.get(1)) {
      throw new IllegalStateException("shf.ACC_2 already set");
    } else {
      filled.set(1);
    }

    acc2.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc3(final BigInteger b) {
    if (filled.get(2)) {
      throw new IllegalStateException("shf.ACC_3 already set");
    } else {
      filled.set(2);
    }

    acc3.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc4(final BigInteger b) {
    if (filled.get(3)) {
      throw new IllegalStateException("shf.ACC_4 already set");
    } else {
      filled.set(3);
    }

    acc4.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc5(final BigInteger b) {
    if (filled.get(4)) {
      throw new IllegalStateException("shf.ACC_5 already set");
    } else {
      filled.set(4);
    }

    acc5.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace arg1Hi(final BigInteger b) {
    if (filled.get(5)) {
      throw new IllegalStateException("shf.ARG_1_HI already set");
    } else {
      filled.set(5);
    }

    arg1Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace arg1Lo(final BigInteger b) {
    if (filled.get(6)) {
      throw new IllegalStateException("shf.ARG_1_LO already set");
    } else {
      filled.set(6);
    }

    arg1Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace arg2Hi(final BigInteger b) {
    if (filled.get(7)) {
      throw new IllegalStateException("shf.ARG_2_HI already set");
    } else {
      filled.set(7);
    }

    arg2Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace arg2Lo(final BigInteger b) {
    if (filled.get(8)) {
      throw new IllegalStateException("shf.ARG_2_LO already set");
    } else {
      filled.set(8);
    }

    arg2Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("shf.BIT_1 already set");
    } else {
      filled.set(10);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("shf.BIT_2 already set");
    } else {
      filled.set(11);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("shf.BIT_3 already set");
    } else {
      filled.set(12);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("shf.BIT_4 already set");
    } else {
      filled.set(13);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB3(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("shf.BIT_B_3 already set");
    } else {
      filled.set(14);
    }

    bitB3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB4(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("shf.BIT_B_4 already set");
    } else {
      filled.set(15);
    }

    bitB4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB5(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("shf.BIT_B_5 already set");
    } else {
      filled.set(16);
    }

    bitB5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB6(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("shf.BIT_B_6 already set");
    } else {
      filled.set(17);
    }

    bitB6.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB7(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("shf.BIT_B_7 already set");
    } else {
      filled.set(18);
    }

    bitB7.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("shf.BITS already set");
    } else {
      filled.set(9);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("shf.BYTE_1 already set");
    } else {
      filled.set(19);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("shf.BYTE_2 already set");
    } else {
      filled.set(20);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("shf.BYTE_3 already set");
    } else {
      filled.set(21);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("shf.BYTE_4 already set");
    } else {
      filled.set(22);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("shf.BYTE_5 already set");
    } else {
      filled.set(23);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace counter(final BigInteger b) {
    if (filled.get(24)) {
      throw new IllegalStateException("shf.COUNTER already set");
    } else {
      filled.set(24);
    }

    counter.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace inst(final BigInteger b) {
    if (filled.get(25)) {
      throw new IllegalStateException("shf.INST already set");
    } else {
      filled.set(25);
    }

    inst.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace isData(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("shf.IS_DATA already set");
    } else {
      filled.set(26);
    }

    isData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace known(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("shf.KNOWN already set");
    } else {
      filled.set(27);
    }

    known.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace leftAlignedSuffixHigh(final BigInteger b) {
    if (filled.get(28)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_HIGH already set");
    } else {
      filled.set(28);
    }

    leftAlignedSuffixHigh.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace leftAlignedSuffixLow(final BigInteger b) {
    if (filled.get(29)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_LOW already set");
    } else {
      filled.set(29);
    }

    leftAlignedSuffixLow.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace low3(final BigInteger b) {
    if (filled.get(30)) {
      throw new IllegalStateException("shf.LOW_3 already set");
    } else {
      filled.set(30);
    }

    low3.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace microShiftParameter(final BigInteger b) {
    if (filled.get(31)) {
      throw new IllegalStateException("shf.MICRO_SHIFT_PARAMETER already set");
    } else {
      filled.set(31);
    }

    microShiftParameter.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace neg(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("shf.NEG already set");
    } else {
      filled.set(32);
    }

    neg.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oneLineInstruction(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("shf.ONE_LINE_INSTRUCTION already set");
    } else {
      filled.set(34);
    }

    oneLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ones(final BigInteger b) {
    if (filled.get(33)) {
      throw new IllegalStateException("shf.ONES already set");
    } else {
      filled.set(33);
    }

    ones.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace resHi(final BigInteger b) {
    if (filled.get(35)) {
      throw new IllegalStateException("shf.RES_HI already set");
    } else {
      filled.set(35);
    }

    resHi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace resLo(final BigInteger b) {
    if (filled.get(36)) {
      throw new IllegalStateException("shf.RES_LO already set");
    } else {
      filled.set(36);
    }

    resLo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace rightAlignedPrefixHigh(final BigInteger b) {
    if (filled.get(37)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_HIGH already set");
    } else {
      filled.set(37);
    }

    rightAlignedPrefixHigh.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace rightAlignedPrefixLow(final BigInteger b) {
    if (filled.get(38)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_LOW already set");
    } else {
      filled.set(38);
    }

    rightAlignedPrefixLow.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb3Hi(final BigInteger b) {
    if (filled.get(39)) {
      throw new IllegalStateException("shf.SHB_3_HI already set");
    } else {
      filled.set(39);
    }

    shb3Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb3Lo(final BigInteger b) {
    if (filled.get(40)) {
      throw new IllegalStateException("shf.SHB_3_LO already set");
    } else {
      filled.set(40);
    }

    shb3Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb4Hi(final BigInteger b) {
    if (filled.get(41)) {
      throw new IllegalStateException("shf.SHB_4_HI already set");
    } else {
      filled.set(41);
    }

    shb4Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb4Lo(final BigInteger b) {
    if (filled.get(42)) {
      throw new IllegalStateException("shf.SHB_4_LO already set");
    } else {
      filled.set(42);
    }

    shb4Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb5Hi(final BigInteger b) {
    if (filled.get(43)) {
      throw new IllegalStateException("shf.SHB_5_HI already set");
    } else {
      filled.set(43);
    }

    shb5Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb5Lo(final BigInteger b) {
    if (filled.get(44)) {
      throw new IllegalStateException("shf.SHB_5_LO already set");
    } else {
      filled.set(44);
    }

    shb5Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb6Hi(final BigInteger b) {
    if (filled.get(45)) {
      throw new IllegalStateException("shf.SHB_6_HI already set");
    } else {
      filled.set(45);
    }

    shb6Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb6Lo(final BigInteger b) {
    if (filled.get(46)) {
      throw new IllegalStateException("shf.SHB_6_LO already set");
    } else {
      filled.set(46);
    }

    shb6Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb7Hi(final BigInteger b) {
    if (filled.get(47)) {
      throw new IllegalStateException("shf.SHB_7_HI already set");
    } else {
      filled.set(47);
    }

    shb7Hi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shb7Lo(final BigInteger b) {
    if (filled.get(48)) {
      throw new IllegalStateException("shf.SHB_7_LO already set");
    } else {
      filled.set(48);
    }

    shb7Lo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace shiftDirection(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("shf.SHIFT_DIRECTION already set");
    } else {
      filled.set(49);
    }

    shiftDirection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace shiftStamp(final BigInteger b) {
    if (filled.get(50)) {
      throw new IllegalStateException("shf.SHIFT_STAMP already set");
    } else {
      filled.set(50);
    }

    shiftStamp.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("shf.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("shf.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("shf.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("shf.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("shf.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("shf.ARG_1_HI has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("shf.ARG_1_LO has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("shf.ARG_2_HI has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("shf.ARG_2_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("shf.BIT_1 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("shf.BIT_2 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("shf.BIT_3 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("shf.BIT_4 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("shf.BIT_B_3 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("shf.BIT_B_4 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("shf.BIT_B_5 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("shf.BIT_B_6 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("shf.BIT_B_7 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("shf.BITS has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("shf.BYTE_1 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("shf.BYTE_2 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("shf.BYTE_3 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("shf.BYTE_4 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("shf.BYTE_5 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("shf.COUNTER has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("shf.INST has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("shf.IS_DATA has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("shf.KNOWN has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_HIGH has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_LOW has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("shf.LOW_3 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("shf.MICRO_SHIFT_PARAMETER has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("shf.NEG has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("shf.ONE_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("shf.ONES has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("shf.RES_HI has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("shf.RES_LO has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_HIGH has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_LOW has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("shf.SHB_3_HI has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("shf.SHB_3_LO has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("shf.SHB_4_HI has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("shf.SHB_4_LO has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("shf.SHB_5_HI has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("shf.SHB_5_LO has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("shf.SHB_6_HI has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("shf.SHB_6_LO has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("shf.SHB_7_HI has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("shf.SHB_7_LO has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("shf.SHIFT_DIRECTION has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("shf.SHIFT_STAMP has not been filled");
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
      arg1Hi.position(arg1Hi.position() + 32);
    }

    if (!filled.get(6)) {
      arg1Lo.position(arg1Lo.position() + 32);
    }

    if (!filled.get(7)) {
      arg2Hi.position(arg2Hi.position() + 32);
    }

    if (!filled.get(8)) {
      arg2Lo.position(arg2Lo.position() + 32);
    }

    if (!filled.get(10)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(11)) {
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(12)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(13)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(14)) {
      bitB3.position(bitB3.position() + 1);
    }

    if (!filled.get(15)) {
      bitB4.position(bitB4.position() + 1);
    }

    if (!filled.get(16)) {
      bitB5.position(bitB5.position() + 1);
    }

    if (!filled.get(17)) {
      bitB6.position(bitB6.position() + 1);
    }

    if (!filled.get(18)) {
      bitB7.position(bitB7.position() + 1);
    }

    if (!filled.get(9)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(19)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(20)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(21)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(22)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(23)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(24)) {
      counter.position(counter.position() + 32);
    }

    if (!filled.get(25)) {
      inst.position(inst.position() + 32);
    }

    if (!filled.get(26)) {
      isData.position(isData.position() + 1);
    }

    if (!filled.get(27)) {
      known.position(known.position() + 1);
    }

    if (!filled.get(28)) {
      leftAlignedSuffixHigh.position(leftAlignedSuffixHigh.position() + 32);
    }

    if (!filled.get(29)) {
      leftAlignedSuffixLow.position(leftAlignedSuffixLow.position() + 32);
    }

    if (!filled.get(30)) {
      low3.position(low3.position() + 32);
    }

    if (!filled.get(31)) {
      microShiftParameter.position(microShiftParameter.position() + 32);
    }

    if (!filled.get(32)) {
      neg.position(neg.position() + 1);
    }

    if (!filled.get(34)) {
      oneLineInstruction.position(oneLineInstruction.position() + 1);
    }

    if (!filled.get(33)) {
      ones.position(ones.position() + 32);
    }

    if (!filled.get(35)) {
      resHi.position(resHi.position() + 32);
    }

    if (!filled.get(36)) {
      resLo.position(resLo.position() + 32);
    }

    if (!filled.get(37)) {
      rightAlignedPrefixHigh.position(rightAlignedPrefixHigh.position() + 32);
    }

    if (!filled.get(38)) {
      rightAlignedPrefixLow.position(rightAlignedPrefixLow.position() + 32);
    }

    if (!filled.get(39)) {
      shb3Hi.position(shb3Hi.position() + 32);
    }

    if (!filled.get(40)) {
      shb3Lo.position(shb3Lo.position() + 32);
    }

    if (!filled.get(41)) {
      shb4Hi.position(shb4Hi.position() + 32);
    }

    if (!filled.get(42)) {
      shb4Lo.position(shb4Lo.position() + 32);
    }

    if (!filled.get(43)) {
      shb5Hi.position(shb5Hi.position() + 32);
    }

    if (!filled.get(44)) {
      shb5Lo.position(shb5Lo.position() + 32);
    }

    if (!filled.get(45)) {
      shb6Hi.position(shb6Hi.position() + 32);
    }

    if (!filled.get(46)) {
      shb6Lo.position(shb6Lo.position() + 32);
    }

    if (!filled.get(47)) {
      shb7Hi.position(shb7Hi.position() + 32);
    }

    if (!filled.get(48)) {
      shb7Lo.position(shb7Lo.position() + 32);
    }

    if (!filled.get(49)) {
      shiftDirection.position(shiftDirection.position() + 1);
    }

    if (!filled.get(50)) {
      shiftStamp.position(shiftStamp.position() + 32);
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
