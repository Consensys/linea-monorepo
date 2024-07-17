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

package net.consensys.linea.zktracer.module.blockhash;

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

  private final MappedByteBuffer absBlock;
  private final MappedByteBuffer blockHashHi;
  private final MappedByteBuffer blockHashLo;
  private final MappedByteBuffer blockNumberHi;
  private final MappedByteBuffer blockNumberLo;
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
  private final MappedByteBuffer inRange;
  private final MappedByteBuffer iomf;
  private final MappedByteBuffer lowerBoundCheck;
  private final MappedByteBuffer relBlock;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer upperBoundCheck;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("blockhash.ABS_BLOCK", 6, length),
        new ColumnHeader("blockhash.BLOCK_HASH_HI", 16, length),
        new ColumnHeader("blockhash.BLOCK_HASH_LO", 16, length),
        new ColumnHeader("blockhash.BLOCK_NUMBER_HI", 16, length),
        new ColumnHeader("blockhash.BLOCK_NUMBER_LO", 16, length),
        new ColumnHeader("blockhash.BYTE_HI_0", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_1", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_10", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_11", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_12", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_13", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_14", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_15", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_2", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_3", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_4", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_5", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_6", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_7", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_8", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_9", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_0", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_1", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_10", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_11", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_12", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_13", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_14", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_15", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_2", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_3", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_4", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_5", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_6", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_7", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_8", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_9", 1, length),
        new ColumnHeader("blockhash.IN_RANGE", 1, length),
        new ColumnHeader("blockhash.IOMF", 1, length),
        new ColumnHeader("blockhash.LOWER_BOUND_CHECK", 1, length),
        new ColumnHeader("blockhash.REL_BLOCK", 1, length),
        new ColumnHeader("blockhash.RES_HI", 16, length),
        new ColumnHeader("blockhash.RES_LO", 16, length),
        new ColumnHeader("blockhash.UPPER_BOUND_CHECK", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absBlock = buffers.get(0);
    this.blockHashHi = buffers.get(1);
    this.blockHashLo = buffers.get(2);
    this.blockNumberHi = buffers.get(3);
    this.blockNumberLo = buffers.get(4);
    this.byteHi0 = buffers.get(5);
    this.byteHi1 = buffers.get(6);
    this.byteHi10 = buffers.get(7);
    this.byteHi11 = buffers.get(8);
    this.byteHi12 = buffers.get(9);
    this.byteHi13 = buffers.get(10);
    this.byteHi14 = buffers.get(11);
    this.byteHi15 = buffers.get(12);
    this.byteHi2 = buffers.get(13);
    this.byteHi3 = buffers.get(14);
    this.byteHi4 = buffers.get(15);
    this.byteHi5 = buffers.get(16);
    this.byteHi6 = buffers.get(17);
    this.byteHi7 = buffers.get(18);
    this.byteHi8 = buffers.get(19);
    this.byteHi9 = buffers.get(20);
    this.byteLo0 = buffers.get(21);
    this.byteLo1 = buffers.get(22);
    this.byteLo10 = buffers.get(23);
    this.byteLo11 = buffers.get(24);
    this.byteLo12 = buffers.get(25);
    this.byteLo13 = buffers.get(26);
    this.byteLo14 = buffers.get(27);
    this.byteLo15 = buffers.get(28);
    this.byteLo2 = buffers.get(29);
    this.byteLo3 = buffers.get(30);
    this.byteLo4 = buffers.get(31);
    this.byteLo5 = buffers.get(32);
    this.byteLo6 = buffers.get(33);
    this.byteLo7 = buffers.get(34);
    this.byteLo8 = buffers.get(35);
    this.byteLo9 = buffers.get(36);
    this.inRange = buffers.get(37);
    this.iomf = buffers.get(38);
    this.lowerBoundCheck = buffers.get(39);
    this.relBlock = buffers.get(40);
    this.resHi = buffers.get(41);
    this.resLo = buffers.get(42);
    this.upperBoundCheck = buffers.get(43);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absBlock(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("blockhash.ABS_BLOCK already set");
    } else {
      filled.set(0);
    }

    if (b >= 281474976710656L) {
      throw new IllegalArgumentException("absBlock has invalid value (" + b + ")");
    }
    absBlock.put((byte) (b >> 40));
    absBlock.put((byte) (b >> 32));
    absBlock.put((byte) (b >> 24));
    absBlock.put((byte) (b >> 16));
    absBlock.put((byte) (b >> 8));
    absBlock.put((byte) b);

    return this;
  }

  public Trace blockHashHi(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_HI already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException(
          "blockHashHi has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      blockHashHi.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      blockHashHi.put(bs.get(j));
    }

    return this;
  }

  public Trace blockHashLo(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_LO already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException(
          "blockHashLo has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      blockHashLo.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      blockHashLo.put(bs.get(j));
    }

    return this;
  }

  public Trace blockNumberHi(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_HI already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException(
          "blockNumberHi has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      blockNumberHi.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      blockNumberHi.put(bs.get(j));
    }

    return this;
  }

  public Trace blockNumberLo(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_LO already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException(
          "blockNumberLo has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      blockNumberLo.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      blockNumberLo.put(bs.get(j));
    }

    return this;
  }

  public Trace byteHi0(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("blockhash.BYTE_HI_0 already set");
    } else {
      filled.set(5);
    }

    byteHi0.put(b.toByte());

    return this;
  }

  public Trace byteHi1(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("blockhash.BYTE_HI_1 already set");
    } else {
      filled.set(6);
    }

    byteHi1.put(b.toByte());

    return this;
  }

  public Trace byteHi10(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("blockhash.BYTE_HI_10 already set");
    } else {
      filled.set(7);
    }

    byteHi10.put(b.toByte());

    return this;
  }

  public Trace byteHi11(final UnsignedByte b) {
    if (filled.get(8)) {
      throw new IllegalStateException("blockhash.BYTE_HI_11 already set");
    } else {
      filled.set(8);
    }

    byteHi11.put(b.toByte());

    return this;
  }

  public Trace byteHi12(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blockhash.BYTE_HI_12 already set");
    } else {
      filled.set(9);
    }

    byteHi12.put(b.toByte());

    return this;
  }

  public Trace byteHi13(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blockhash.BYTE_HI_13 already set");
    } else {
      filled.set(10);
    }

    byteHi13.put(b.toByte());

    return this;
  }

  public Trace byteHi14(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blockhash.BYTE_HI_14 already set");
    } else {
      filled.set(11);
    }

    byteHi14.put(b.toByte());

    return this;
  }

  public Trace byteHi15(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blockhash.BYTE_HI_15 already set");
    } else {
      filled.set(12);
    }

    byteHi15.put(b.toByte());

    return this;
  }

  public Trace byteHi2(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("blockhash.BYTE_HI_2 already set");
    } else {
      filled.set(13);
    }

    byteHi2.put(b.toByte());

    return this;
  }

  public Trace byteHi3(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("blockhash.BYTE_HI_3 already set");
    } else {
      filled.set(14);
    }

    byteHi3.put(b.toByte());

    return this;
  }

  public Trace byteHi4(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("blockhash.BYTE_HI_4 already set");
    } else {
      filled.set(15);
    }

    byteHi4.put(b.toByte());

    return this;
  }

  public Trace byteHi5(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("blockhash.BYTE_HI_5 already set");
    } else {
      filled.set(16);
    }

    byteHi5.put(b.toByte());

    return this;
  }

  public Trace byteHi6(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("blockhash.BYTE_HI_6 already set");
    } else {
      filled.set(17);
    }

    byteHi6.put(b.toByte());

    return this;
  }

  public Trace byteHi7(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("blockhash.BYTE_HI_7 already set");
    } else {
      filled.set(18);
    }

    byteHi7.put(b.toByte());

    return this;
  }

  public Trace byteHi8(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("blockhash.BYTE_HI_8 already set");
    } else {
      filled.set(19);
    }

    byteHi8.put(b.toByte());

    return this;
  }

  public Trace byteHi9(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("blockhash.BYTE_HI_9 already set");
    } else {
      filled.set(20);
    }

    byteHi9.put(b.toByte());

    return this;
  }

  public Trace byteLo0(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("blockhash.BYTE_LO_0 already set");
    } else {
      filled.set(21);
    }

    byteLo0.put(b.toByte());

    return this;
  }

  public Trace byteLo1(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("blockhash.BYTE_LO_1 already set");
    } else {
      filled.set(22);
    }

    byteLo1.put(b.toByte());

    return this;
  }

  public Trace byteLo10(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("blockhash.BYTE_LO_10 already set");
    } else {
      filled.set(23);
    }

    byteLo10.put(b.toByte());

    return this;
  }

  public Trace byteLo11(final UnsignedByte b) {
    if (filled.get(24)) {
      throw new IllegalStateException("blockhash.BYTE_LO_11 already set");
    } else {
      filled.set(24);
    }

    byteLo11.put(b.toByte());

    return this;
  }

  public Trace byteLo12(final UnsignedByte b) {
    if (filled.get(25)) {
      throw new IllegalStateException("blockhash.BYTE_LO_12 already set");
    } else {
      filled.set(25);
    }

    byteLo12.put(b.toByte());

    return this;
  }

  public Trace byteLo13(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("blockhash.BYTE_LO_13 already set");
    } else {
      filled.set(26);
    }

    byteLo13.put(b.toByte());

    return this;
  }

  public Trace byteLo14(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("blockhash.BYTE_LO_14 already set");
    } else {
      filled.set(27);
    }

    byteLo14.put(b.toByte());

    return this;
  }

  public Trace byteLo15(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("blockhash.BYTE_LO_15 already set");
    } else {
      filled.set(28);
    }

    byteLo15.put(b.toByte());

    return this;
  }

  public Trace byteLo2(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("blockhash.BYTE_LO_2 already set");
    } else {
      filled.set(29);
    }

    byteLo2.put(b.toByte());

    return this;
  }

  public Trace byteLo3(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("blockhash.BYTE_LO_3 already set");
    } else {
      filled.set(30);
    }

    byteLo3.put(b.toByte());

    return this;
  }

  public Trace byteLo4(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("blockhash.BYTE_LO_4 already set");
    } else {
      filled.set(31);
    }

    byteLo4.put(b.toByte());

    return this;
  }

  public Trace byteLo5(final UnsignedByte b) {
    if (filled.get(32)) {
      throw new IllegalStateException("blockhash.BYTE_LO_5 already set");
    } else {
      filled.set(32);
    }

    byteLo5.put(b.toByte());

    return this;
  }

  public Trace byteLo6(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("blockhash.BYTE_LO_6 already set");
    } else {
      filled.set(33);
    }

    byteLo6.put(b.toByte());

    return this;
  }

  public Trace byteLo7(final UnsignedByte b) {
    if (filled.get(34)) {
      throw new IllegalStateException("blockhash.BYTE_LO_7 already set");
    } else {
      filled.set(34);
    }

    byteLo7.put(b.toByte());

    return this;
  }

  public Trace byteLo8(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("blockhash.BYTE_LO_8 already set");
    } else {
      filled.set(35);
    }

    byteLo8.put(b.toByte());

    return this;
  }

  public Trace byteLo9(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("blockhash.BYTE_LO_9 already set");
    } else {
      filled.set(36);
    }

    byteLo9.put(b.toByte());

    return this;
  }

  public Trace inRange(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("blockhash.IN_RANGE already set");
    } else {
      filled.set(37);
    }

    inRange.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace iomf(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("blockhash.IOMF already set");
    } else {
      filled.set(38);
    }

    iomf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lowerBoundCheck(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("blockhash.LOWER_BOUND_CHECK already set");
    } else {
      filled.set(39);
    }

    lowerBoundCheck.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace relBlock(final long b) {
    if (filled.get(40)) {
      throw new IllegalStateException("blockhash.REL_BLOCK already set");
    } else {
      filled.set(40);
    }

    if (b >= 256L) {
      throw new IllegalArgumentException("relBlock has invalid value (" + b + ")");
    }
    relBlock.put((byte) b);

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("blockhash.RES_HI already set");
    } else {
      filled.set(41);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException("resHi has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      resHi.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      resHi.put(bs.get(j));
    }

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("blockhash.RES_LO already set");
    } else {
      filled.set(42);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException("resLo has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      resLo.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      resLo.put(bs.get(j));
    }

    return this;
  }

  public Trace upperBoundCheck(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("blockhash.UPPER_BOUND_CHECK already set");
    } else {
      filled.set(43);
    }

    upperBoundCheck.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("blockhash.ABS_BLOCK has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_HI has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_LO has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_HI has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_LO has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("blockhash.BYTE_HI_0 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("blockhash.BYTE_HI_1 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("blockhash.BYTE_HI_10 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("blockhash.BYTE_HI_11 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("blockhash.BYTE_HI_12 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("blockhash.BYTE_HI_13 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("blockhash.BYTE_HI_14 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("blockhash.BYTE_HI_15 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("blockhash.BYTE_HI_2 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("blockhash.BYTE_HI_3 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("blockhash.BYTE_HI_4 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("blockhash.BYTE_HI_5 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("blockhash.BYTE_HI_6 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("blockhash.BYTE_HI_7 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("blockhash.BYTE_HI_8 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("blockhash.BYTE_HI_9 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("blockhash.BYTE_LO_0 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("blockhash.BYTE_LO_1 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("blockhash.BYTE_LO_10 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("blockhash.BYTE_LO_11 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("blockhash.BYTE_LO_12 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("blockhash.BYTE_LO_13 has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("blockhash.BYTE_LO_14 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("blockhash.BYTE_LO_15 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("blockhash.BYTE_LO_2 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("blockhash.BYTE_LO_3 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("blockhash.BYTE_LO_4 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("blockhash.BYTE_LO_5 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("blockhash.BYTE_LO_6 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("blockhash.BYTE_LO_7 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("blockhash.BYTE_LO_8 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("blockhash.BYTE_LO_9 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("blockhash.IN_RANGE has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("blockhash.IOMF has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("blockhash.LOWER_BOUND_CHECK has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("blockhash.REL_BLOCK has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("blockhash.RES_HI has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("blockhash.RES_LO has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("blockhash.UPPER_BOUND_CHECK has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absBlock.position(absBlock.position() + 6);
    }

    if (!filled.get(1)) {
      blockHashHi.position(blockHashHi.position() + 16);
    }

    if (!filled.get(2)) {
      blockHashLo.position(blockHashLo.position() + 16);
    }

    if (!filled.get(3)) {
      blockNumberHi.position(blockNumberHi.position() + 16);
    }

    if (!filled.get(4)) {
      blockNumberLo.position(blockNumberLo.position() + 16);
    }

    if (!filled.get(5)) {
      byteHi0.position(byteHi0.position() + 1);
    }

    if (!filled.get(6)) {
      byteHi1.position(byteHi1.position() + 1);
    }

    if (!filled.get(7)) {
      byteHi10.position(byteHi10.position() + 1);
    }

    if (!filled.get(8)) {
      byteHi11.position(byteHi11.position() + 1);
    }

    if (!filled.get(9)) {
      byteHi12.position(byteHi12.position() + 1);
    }

    if (!filled.get(10)) {
      byteHi13.position(byteHi13.position() + 1);
    }

    if (!filled.get(11)) {
      byteHi14.position(byteHi14.position() + 1);
    }

    if (!filled.get(12)) {
      byteHi15.position(byteHi15.position() + 1);
    }

    if (!filled.get(13)) {
      byteHi2.position(byteHi2.position() + 1);
    }

    if (!filled.get(14)) {
      byteHi3.position(byteHi3.position() + 1);
    }

    if (!filled.get(15)) {
      byteHi4.position(byteHi4.position() + 1);
    }

    if (!filled.get(16)) {
      byteHi5.position(byteHi5.position() + 1);
    }

    if (!filled.get(17)) {
      byteHi6.position(byteHi6.position() + 1);
    }

    if (!filled.get(18)) {
      byteHi7.position(byteHi7.position() + 1);
    }

    if (!filled.get(19)) {
      byteHi8.position(byteHi8.position() + 1);
    }

    if (!filled.get(20)) {
      byteHi9.position(byteHi9.position() + 1);
    }

    if (!filled.get(21)) {
      byteLo0.position(byteLo0.position() + 1);
    }

    if (!filled.get(22)) {
      byteLo1.position(byteLo1.position() + 1);
    }

    if (!filled.get(23)) {
      byteLo10.position(byteLo10.position() + 1);
    }

    if (!filled.get(24)) {
      byteLo11.position(byteLo11.position() + 1);
    }

    if (!filled.get(25)) {
      byteLo12.position(byteLo12.position() + 1);
    }

    if (!filled.get(26)) {
      byteLo13.position(byteLo13.position() + 1);
    }

    if (!filled.get(27)) {
      byteLo14.position(byteLo14.position() + 1);
    }

    if (!filled.get(28)) {
      byteLo15.position(byteLo15.position() + 1);
    }

    if (!filled.get(29)) {
      byteLo2.position(byteLo2.position() + 1);
    }

    if (!filled.get(30)) {
      byteLo3.position(byteLo3.position() + 1);
    }

    if (!filled.get(31)) {
      byteLo4.position(byteLo4.position() + 1);
    }

    if (!filled.get(32)) {
      byteLo5.position(byteLo5.position() + 1);
    }

    if (!filled.get(33)) {
      byteLo6.position(byteLo6.position() + 1);
    }

    if (!filled.get(34)) {
      byteLo7.position(byteLo7.position() + 1);
    }

    if (!filled.get(35)) {
      byteLo8.position(byteLo8.position() + 1);
    }

    if (!filled.get(36)) {
      byteLo9.position(byteLo9.position() + 1);
    }

    if (!filled.get(37)) {
      inRange.position(inRange.position() + 1);
    }

    if (!filled.get(38)) {
      iomf.position(iomf.position() + 1);
    }

    if (!filled.get(39)) {
      lowerBoundCheck.position(lowerBoundCheck.position() + 1);
    }

    if (!filled.get(40)) {
      relBlock.position(relBlock.position() + 1);
    }

    if (!filled.get(41)) {
      resHi.position(resHi.position() + 16);
    }

    if (!filled.get(42)) {
      resLo.position(resLo.position() + 16);
    }

    if (!filled.get(43)) {
      upperBoundCheck.position(upperBoundCheck.position() + 1);
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
