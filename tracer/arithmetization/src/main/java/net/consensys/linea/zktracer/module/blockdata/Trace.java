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

package net.consensys.linea.zktracer.module.blockdata;

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
  public static final int MAX_CT = 0x6;
  public static final int ROW_SHIFT_BASEFEE = 0x6;
  public static final int ROW_SHIFT_CHAINID = 0x5;
  public static final int ROW_SHIFT_COINBASE = 0x0;
  public static final int ROW_SHIFT_DIFFICULTY = 0x3;
  public static final int ROW_SHIFT_GASLIMIT = 0x4;
  public static final int ROW_SHIFT_NUMBER = 0x2;
  public static final int ROW_SHIFT_TIMESTAMP = 0x1;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer basefee;
  private final MappedByteBuffer blockGasLimit;
  private final MappedByteBuffer byteHi0;
  private final MappedByteBuffer byteHi1;
  private final MappedByteBuffer byteHi10;
  private final MappedByteBuffer byteHi11;
  private final MappedByteBuffer byteHi12;
  private final MappedByteBuffer byteHi13;
  private final MappedByteBuffer byteHi14;
  private final MappedByteBuffer byteHi15;
  private final MappedByteBuffer byteHi2;
  private final MappedByteBuffer byteHi3;
  private final MappedByteBuffer byteHi4;
  private final MappedByteBuffer byteHi5;
  private final MappedByteBuffer byteHi6;
  private final MappedByteBuffer byteHi7;
  private final MappedByteBuffer byteHi8;
  private final MappedByteBuffer byteHi9;
  private final MappedByteBuffer byteLo0;
  private final MappedByteBuffer byteLo1;
  private final MappedByteBuffer byteLo10;
  private final MappedByteBuffer byteLo11;
  private final MappedByteBuffer byteLo12;
  private final MappedByteBuffer byteLo13;
  private final MappedByteBuffer byteLo14;
  private final MappedByteBuffer byteLo15;
  private final MappedByteBuffer byteLo2;
  private final MappedByteBuffer byteLo3;
  private final MappedByteBuffer byteLo4;
  private final MappedByteBuffer byteLo5;
  private final MappedByteBuffer byteLo6;
  private final MappedByteBuffer byteLo7;
  private final MappedByteBuffer byteLo8;
  private final MappedByteBuffer byteLo9;
  private final MappedByteBuffer coinbaseHi;
  private final MappedByteBuffer coinbaseLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer dataHi;
  private final MappedByteBuffer dataLo;
  private final MappedByteBuffer firstBlockNumber;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer relBlock;
  private final MappedByteBuffer relTxNumMax;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("blockdata.BASEFEE", 8, length),
        new ColumnHeader("blockdata.BLOCK_GAS_LIMIT", 8, length),
        new ColumnHeader("blockdata.BYTE_HI_0", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_1", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_10", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_11", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_12", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_13", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_14", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_15", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_2", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_3", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_4", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_5", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_6", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_7", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_8", 1, length),
        new ColumnHeader("blockdata.BYTE_HI_9", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_0", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_1", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_10", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_11", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_12", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_13", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_14", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_15", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_2", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_3", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_4", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_5", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_6", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_7", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_8", 1, length),
        new ColumnHeader("blockdata.BYTE_LO_9", 1, length),
        new ColumnHeader("blockdata.COINBASE_HI", 8, length),
        new ColumnHeader("blockdata.COINBASE_LO", 32, length),
        new ColumnHeader("blockdata.CT", 2, length),
        new ColumnHeader("blockdata.DATA_HI", 32, length),
        new ColumnHeader("blockdata.DATA_LO", 32, length),
        new ColumnHeader("blockdata.FIRST_BLOCK_NUMBER", 8, length),
        new ColumnHeader("blockdata.INST", 1, length),
        new ColumnHeader("blockdata.REL_BLOCK", 2, length),
        new ColumnHeader("blockdata.REL_TX_NUM_MAX", 2, length),
        new ColumnHeader("blockdata.WCP_FLAG", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.basefee = buffers.get(0);
    this.blockGasLimit = buffers.get(1);
    this.byteHi0 = buffers.get(2);
    this.byteHi1 = buffers.get(3);
    this.byteHi10 = buffers.get(4);
    this.byteHi11 = buffers.get(5);
    this.byteHi12 = buffers.get(6);
    this.byteHi13 = buffers.get(7);
    this.byteHi14 = buffers.get(8);
    this.byteHi15 = buffers.get(9);
    this.byteHi2 = buffers.get(10);
    this.byteHi3 = buffers.get(11);
    this.byteHi4 = buffers.get(12);
    this.byteHi5 = buffers.get(13);
    this.byteHi6 = buffers.get(14);
    this.byteHi7 = buffers.get(15);
    this.byteHi8 = buffers.get(16);
    this.byteHi9 = buffers.get(17);
    this.byteLo0 = buffers.get(18);
    this.byteLo1 = buffers.get(19);
    this.byteLo10 = buffers.get(20);
    this.byteLo11 = buffers.get(21);
    this.byteLo12 = buffers.get(22);
    this.byteLo13 = buffers.get(23);
    this.byteLo14 = buffers.get(24);
    this.byteLo15 = buffers.get(25);
    this.byteLo2 = buffers.get(26);
    this.byteLo3 = buffers.get(27);
    this.byteLo4 = buffers.get(28);
    this.byteLo5 = buffers.get(29);
    this.byteLo6 = buffers.get(30);
    this.byteLo7 = buffers.get(31);
    this.byteLo8 = buffers.get(32);
    this.byteLo9 = buffers.get(33);
    this.coinbaseHi = buffers.get(34);
    this.coinbaseLo = buffers.get(35);
    this.ct = buffers.get(36);
    this.dataHi = buffers.get(37);
    this.dataLo = buffers.get(38);
    this.firstBlockNumber = buffers.get(39);
    this.inst = buffers.get(40);
    this.relBlock = buffers.get(41);
    this.relTxNumMax = buffers.get(42);
    this.wcpFlag = buffers.get(43);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace basefee(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("blockdata.BASEFEE already set");
    } else {
      filled.set(0);
    }

    basefee.putLong(b);

    return this;
  }

  public Trace blockGasLimit(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("blockdata.BLOCK_GAS_LIMIT already set");
    } else {
      filled.set(1);
    }

    blockGasLimit.putLong(b);

    return this;
  }

  public Trace byteHi0(final UnsignedByte b) {
    if (filled.get(2)) {
      throw new IllegalStateException("blockdata.BYTE_HI_0 already set");
    } else {
      filled.set(2);
    }

    byteHi0.put(b.toByte());

    return this;
  }

  public Trace byteHi1(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("blockdata.BYTE_HI_1 already set");
    } else {
      filled.set(3);
    }

    byteHi1.put(b.toByte());

    return this;
  }

  public Trace byteHi10(final UnsignedByte b) {
    if (filled.get(4)) {
      throw new IllegalStateException("blockdata.BYTE_HI_10 already set");
    } else {
      filled.set(4);
    }

    byteHi10.put(b.toByte());

    return this;
  }

  public Trace byteHi11(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("blockdata.BYTE_HI_11 already set");
    } else {
      filled.set(5);
    }

    byteHi11.put(b.toByte());

    return this;
  }

  public Trace byteHi12(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("blockdata.BYTE_HI_12 already set");
    } else {
      filled.set(6);
    }

    byteHi12.put(b.toByte());

    return this;
  }

  public Trace byteHi13(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("blockdata.BYTE_HI_13 already set");
    } else {
      filled.set(7);
    }

    byteHi13.put(b.toByte());

    return this;
  }

  public Trace byteHi14(final UnsignedByte b) {
    if (filled.get(8)) {
      throw new IllegalStateException("blockdata.BYTE_HI_14 already set");
    } else {
      filled.set(8);
    }

    byteHi14.put(b.toByte());

    return this;
  }

  public Trace byteHi15(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blockdata.BYTE_HI_15 already set");
    } else {
      filled.set(9);
    }

    byteHi15.put(b.toByte());

    return this;
  }

  public Trace byteHi2(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blockdata.BYTE_HI_2 already set");
    } else {
      filled.set(10);
    }

    byteHi2.put(b.toByte());

    return this;
  }

  public Trace byteHi3(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blockdata.BYTE_HI_3 already set");
    } else {
      filled.set(11);
    }

    byteHi3.put(b.toByte());

    return this;
  }

  public Trace byteHi4(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blockdata.BYTE_HI_4 already set");
    } else {
      filled.set(12);
    }

    byteHi4.put(b.toByte());

    return this;
  }

  public Trace byteHi5(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("blockdata.BYTE_HI_5 already set");
    } else {
      filled.set(13);
    }

    byteHi5.put(b.toByte());

    return this;
  }

  public Trace byteHi6(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("blockdata.BYTE_HI_6 already set");
    } else {
      filled.set(14);
    }

    byteHi6.put(b.toByte());

    return this;
  }

  public Trace byteHi7(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("blockdata.BYTE_HI_7 already set");
    } else {
      filled.set(15);
    }

    byteHi7.put(b.toByte());

    return this;
  }

  public Trace byteHi8(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("blockdata.BYTE_HI_8 already set");
    } else {
      filled.set(16);
    }

    byteHi8.put(b.toByte());

    return this;
  }

  public Trace byteHi9(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("blockdata.BYTE_HI_9 already set");
    } else {
      filled.set(17);
    }

    byteHi9.put(b.toByte());

    return this;
  }

  public Trace byteLo0(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("blockdata.BYTE_LO_0 already set");
    } else {
      filled.set(18);
    }

    byteLo0.put(b.toByte());

    return this;
  }

  public Trace byteLo1(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("blockdata.BYTE_LO_1 already set");
    } else {
      filled.set(19);
    }

    byteLo1.put(b.toByte());

    return this;
  }

  public Trace byteLo10(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("blockdata.BYTE_LO_10 already set");
    } else {
      filled.set(20);
    }

    byteLo10.put(b.toByte());

    return this;
  }

  public Trace byteLo11(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("blockdata.BYTE_LO_11 already set");
    } else {
      filled.set(21);
    }

    byteLo11.put(b.toByte());

    return this;
  }

  public Trace byteLo12(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("blockdata.BYTE_LO_12 already set");
    } else {
      filled.set(22);
    }

    byteLo12.put(b.toByte());

    return this;
  }

  public Trace byteLo13(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("blockdata.BYTE_LO_13 already set");
    } else {
      filled.set(23);
    }

    byteLo13.put(b.toByte());

    return this;
  }

  public Trace byteLo14(final UnsignedByte b) {
    if (filled.get(24)) {
      throw new IllegalStateException("blockdata.BYTE_LO_14 already set");
    } else {
      filled.set(24);
    }

    byteLo14.put(b.toByte());

    return this;
  }

  public Trace byteLo15(final UnsignedByte b) {
    if (filled.get(25)) {
      throw new IllegalStateException("blockdata.BYTE_LO_15 already set");
    } else {
      filled.set(25);
    }

    byteLo15.put(b.toByte());

    return this;
  }

  public Trace byteLo2(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("blockdata.BYTE_LO_2 already set");
    } else {
      filled.set(26);
    }

    byteLo2.put(b.toByte());

    return this;
  }

  public Trace byteLo3(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("blockdata.BYTE_LO_3 already set");
    } else {
      filled.set(27);
    }

    byteLo3.put(b.toByte());

    return this;
  }

  public Trace byteLo4(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("blockdata.BYTE_LO_4 already set");
    } else {
      filled.set(28);
    }

    byteLo4.put(b.toByte());

    return this;
  }

  public Trace byteLo5(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("blockdata.BYTE_LO_5 already set");
    } else {
      filled.set(29);
    }

    byteLo5.put(b.toByte());

    return this;
  }

  public Trace byteLo6(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("blockdata.BYTE_LO_6 already set");
    } else {
      filled.set(30);
    }

    byteLo6.put(b.toByte());

    return this;
  }

  public Trace byteLo7(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("blockdata.BYTE_LO_7 already set");
    } else {
      filled.set(31);
    }

    byteLo7.put(b.toByte());

    return this;
  }

  public Trace byteLo8(final UnsignedByte b) {
    if (filled.get(32)) {
      throw new IllegalStateException("blockdata.BYTE_LO_8 already set");
    } else {
      filled.set(32);
    }

    byteLo8.put(b.toByte());

    return this;
  }

  public Trace byteLo9(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("blockdata.BYTE_LO_9 already set");
    } else {
      filled.set(33);
    }

    byteLo9.put(b.toByte());

    return this;
  }

  public Trace coinbaseHi(final long b) {
    if (filled.get(34)) {
      throw new IllegalStateException("blockdata.COINBASE_HI already set");
    } else {
      filled.set(34);
    }

    coinbaseHi.putLong(b);

    return this;
  }

  public Trace coinbaseLo(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("blockdata.COINBASE_LO already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      coinbaseLo.put((byte) 0);
    }
    coinbaseLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace ct(final short b) {
    if (filled.get(36)) {
      throw new IllegalStateException("blockdata.CT already set");
    } else {
      filled.set(36);
    }

    ct.putShort(b);

    return this;
  }

  public Trace dataHi(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("blockdata.DATA_HI already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      dataHi.put((byte) 0);
    }
    dataHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace dataLo(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("blockdata.DATA_LO already set");
    } else {
      filled.set(38);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      dataLo.put((byte) 0);
    }
    dataLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace firstBlockNumber(final long b) {
    if (filled.get(39)) {
      throw new IllegalStateException("blockdata.FIRST_BLOCK_NUMBER already set");
    } else {
      filled.set(39);
    }

    firstBlockNumber.putLong(b);

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(40)) {
      throw new IllegalStateException("blockdata.INST already set");
    } else {
      filled.set(40);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace relBlock(final short b) {
    if (filled.get(41)) {
      throw new IllegalStateException("blockdata.REL_BLOCK already set");
    } else {
      filled.set(41);
    }

    relBlock.putShort(b);

    return this;
  }

  public Trace relTxNumMax(final short b) {
    if (filled.get(42)) {
      throw new IllegalStateException("blockdata.REL_TX_NUM_MAX already set");
    } else {
      filled.set(42);
    }

    relTxNumMax.putShort(b);

    return this;
  }

  public Trace wcpFlag(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("blockdata.WCP_FLAG already set");
    } else {
      filled.set(43);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("blockdata.BASEFEE has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("blockdata.BLOCK_GAS_LIMIT has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("blockdata.BYTE_HI_0 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("blockdata.BYTE_HI_1 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("blockdata.BYTE_HI_10 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("blockdata.BYTE_HI_11 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("blockdata.BYTE_HI_12 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("blockdata.BYTE_HI_13 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("blockdata.BYTE_HI_14 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("blockdata.BYTE_HI_15 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("blockdata.BYTE_HI_2 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("blockdata.BYTE_HI_3 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("blockdata.BYTE_HI_4 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("blockdata.BYTE_HI_5 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("blockdata.BYTE_HI_6 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("blockdata.BYTE_HI_7 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("blockdata.BYTE_HI_8 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("blockdata.BYTE_HI_9 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("blockdata.BYTE_LO_0 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("blockdata.BYTE_LO_1 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("blockdata.BYTE_LO_10 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("blockdata.BYTE_LO_11 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("blockdata.BYTE_LO_12 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("blockdata.BYTE_LO_13 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("blockdata.BYTE_LO_14 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("blockdata.BYTE_LO_15 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("blockdata.BYTE_LO_2 has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("blockdata.BYTE_LO_3 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("blockdata.BYTE_LO_4 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("blockdata.BYTE_LO_5 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("blockdata.BYTE_LO_6 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("blockdata.BYTE_LO_7 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("blockdata.BYTE_LO_8 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("blockdata.BYTE_LO_9 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("blockdata.COINBASE_HI has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("blockdata.COINBASE_LO has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("blockdata.CT has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("blockdata.DATA_HI has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("blockdata.DATA_LO has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("blockdata.FIRST_BLOCK_NUMBER has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("blockdata.INST has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("blockdata.REL_BLOCK has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("blockdata.REL_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("blockdata.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      basefee.position(basefee.position() + 8);
    }

    if (!filled.get(1)) {
      blockGasLimit.position(blockGasLimit.position() + 8);
    }

    if (!filled.get(2)) {
      byteHi0.position(byteHi0.position() + 1);
    }

    if (!filled.get(3)) {
      byteHi1.position(byteHi1.position() + 1);
    }

    if (!filled.get(4)) {
      byteHi10.position(byteHi10.position() + 1);
    }

    if (!filled.get(5)) {
      byteHi11.position(byteHi11.position() + 1);
    }

    if (!filled.get(6)) {
      byteHi12.position(byteHi12.position() + 1);
    }

    if (!filled.get(7)) {
      byteHi13.position(byteHi13.position() + 1);
    }

    if (!filled.get(8)) {
      byteHi14.position(byteHi14.position() + 1);
    }

    if (!filled.get(9)) {
      byteHi15.position(byteHi15.position() + 1);
    }

    if (!filled.get(10)) {
      byteHi2.position(byteHi2.position() + 1);
    }

    if (!filled.get(11)) {
      byteHi3.position(byteHi3.position() + 1);
    }

    if (!filled.get(12)) {
      byteHi4.position(byteHi4.position() + 1);
    }

    if (!filled.get(13)) {
      byteHi5.position(byteHi5.position() + 1);
    }

    if (!filled.get(14)) {
      byteHi6.position(byteHi6.position() + 1);
    }

    if (!filled.get(15)) {
      byteHi7.position(byteHi7.position() + 1);
    }

    if (!filled.get(16)) {
      byteHi8.position(byteHi8.position() + 1);
    }

    if (!filled.get(17)) {
      byteHi9.position(byteHi9.position() + 1);
    }

    if (!filled.get(18)) {
      byteLo0.position(byteLo0.position() + 1);
    }

    if (!filled.get(19)) {
      byteLo1.position(byteLo1.position() + 1);
    }

    if (!filled.get(20)) {
      byteLo10.position(byteLo10.position() + 1);
    }

    if (!filled.get(21)) {
      byteLo11.position(byteLo11.position() + 1);
    }

    if (!filled.get(22)) {
      byteLo12.position(byteLo12.position() + 1);
    }

    if (!filled.get(23)) {
      byteLo13.position(byteLo13.position() + 1);
    }

    if (!filled.get(24)) {
      byteLo14.position(byteLo14.position() + 1);
    }

    if (!filled.get(25)) {
      byteLo15.position(byteLo15.position() + 1);
    }

    if (!filled.get(26)) {
      byteLo2.position(byteLo2.position() + 1);
    }

    if (!filled.get(27)) {
      byteLo3.position(byteLo3.position() + 1);
    }

    if (!filled.get(28)) {
      byteLo4.position(byteLo4.position() + 1);
    }

    if (!filled.get(29)) {
      byteLo5.position(byteLo5.position() + 1);
    }

    if (!filled.get(30)) {
      byteLo6.position(byteLo6.position() + 1);
    }

    if (!filled.get(31)) {
      byteLo7.position(byteLo7.position() + 1);
    }

    if (!filled.get(32)) {
      byteLo8.position(byteLo8.position() + 1);
    }

    if (!filled.get(33)) {
      byteLo9.position(byteLo9.position() + 1);
    }

    if (!filled.get(34)) {
      coinbaseHi.position(coinbaseHi.position() + 8);
    }

    if (!filled.get(35)) {
      coinbaseLo.position(coinbaseLo.position() + 32);
    }

    if (!filled.get(36)) {
      ct.position(ct.position() + 2);
    }

    if (!filled.get(37)) {
      dataHi.position(dataHi.position() + 32);
    }

    if (!filled.get(38)) {
      dataLo.position(dataLo.position() + 32);
    }

    if (!filled.get(39)) {
      firstBlockNumber.position(firstBlockNumber.position() + 8);
    }

    if (!filled.get(40)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(41)) {
      relBlock.position(relBlock.position() + 2);
    }

    if (!filled.get(42)) {
      relTxNumMax.position(relTxNumMax.position() + 2);
    }

    if (!filled.get(43)) {
      wcpFlag.position(wcpFlag.position() + 1);
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
