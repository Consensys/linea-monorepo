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

package net.consensys.linea.zktracer.module.trm;

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

  private final MappedByteBuffer accHi;
  private final MappedByteBuffer accLo;
  private final MappedByteBuffer accT;
  private final MappedByteBuffer byteHi;
  private final MappedByteBuffer byteLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer isPrecompile;
  private final MappedByteBuffer one;
  private final MappedByteBuffer pbit;
  private final MappedByteBuffer rawAddressHi;
  private final MappedByteBuffer rawAddressLo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer trmAddressHi;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("trm.ACC_HI", 16, length),
        new ColumnHeader("trm.ACC_LO", 16, length),
        new ColumnHeader("trm.ACC_T", 4, length),
        new ColumnHeader("trm.BYTE_HI", 1, length),
        new ColumnHeader("trm.BYTE_LO", 1, length),
        new ColumnHeader("trm.CT", 1, length),
        new ColumnHeader("trm.IS_PRECOMPILE", 1, length),
        new ColumnHeader("trm.ONE", 1, length),
        new ColumnHeader("trm.PBIT", 1, length),
        new ColumnHeader("trm.RAW_ADDRESS_HI", 16, length),
        new ColumnHeader("trm.RAW_ADDRESS_LO", 16, length),
        new ColumnHeader("trm.STAMP", 3, length),
        new ColumnHeader("trm.TRM_ADDRESS_HI", 4, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.accHi = buffers.get(0);
    this.accLo = buffers.get(1);
    this.accT = buffers.get(2);
    this.byteHi = buffers.get(3);
    this.byteLo = buffers.get(4);
    this.ct = buffers.get(5);
    this.isPrecompile = buffers.get(6);
    this.one = buffers.get(7);
    this.pbit = buffers.get(8);
    this.rawAddressHi = buffers.get(9);
    this.rawAddressLo = buffers.get(10);
    this.stamp = buffers.get(11);
    this.trmAddressHi = buffers.get(12);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace accHi(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("trm.ACC_HI already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException("accHi has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      accHi.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      accHi.put(bs.get(j));
    }

    return this;
  }

  public Trace accLo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("trm.ACC_LO already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException("accLo has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      accLo.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      accLo.put(bs.get(j));
    }

    return this;
  }

  public Trace accT(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("trm.ACC_T already set");
    } else {
      filled.set(2);
    }

    if (b >= 4294967296L) {
      throw new IllegalArgumentException("accT has invalid value (" + b + ")");
    }
    accT.put((byte) (b >> 24));
    accT.put((byte) (b >> 16));
    accT.put((byte) (b >> 8));
    accT.put((byte) b);

    return this;
  }

  public Trace byteHi(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("trm.BYTE_HI already set");
    } else {
      filled.set(3);
    }

    byteHi.put(b.toByte());

    return this;
  }

  public Trace byteLo(final UnsignedByte b) {
    if (filled.get(4)) {
      throw new IllegalStateException("trm.BYTE_LO already set");
    } else {
      filled.set(4);
    }

    byteLo.put(b.toByte());

    return this;
  }

  public Trace ct(final long b) {
    if (filled.get(5)) {
      throw new IllegalStateException("trm.CT already set");
    } else {
      filled.set(5);
    }

    if (b >= 16L) {
      throw new IllegalArgumentException("ct has invalid value (" + b + ")");
    }
    ct.put((byte) b);

    return this;
  }

  public Trace isPrecompile(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("trm.IS_PRECOMPILE already set");
    } else {
      filled.set(6);
    }

    isPrecompile.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace one(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("trm.ONE already set");
    } else {
      filled.set(7);
    }

    one.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pbit(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("trm.PBIT already set");
    } else {
      filled.set(8);
    }

    pbit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rawAddressHi(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("trm.RAW_ADDRESS_HI already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException(
          "rawAddressHi has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      rawAddressHi.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      rawAddressHi.put(bs.get(j));
    }

    return this;
  }

  public Trace rawAddressLo(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("trm.RAW_ADDRESS_LO already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if (bs.bitLength() > 128) {
      throw new IllegalArgumentException(
          "rawAddressLo has invalid width (" + bs.bitLength() + "bits)");
    }
    // Write padding (if necessary)
    for (int i = bs.size(); i < 16; i++) {
      rawAddressLo.put((byte) 0);
    }
    // Write bytes
    for (int j = 0; j < bs.size(); j++) {
      rawAddressLo.put(bs.get(j));
    }

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(11)) {
      throw new IllegalStateException("trm.STAMP already set");
    } else {
      filled.set(11);
    }

    if (b >= 16777216L) {
      throw new IllegalArgumentException("stamp has invalid value (" + b + ")");
    }
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);

    return this;
  }

  public Trace trmAddressHi(final long b) {
    if (filled.get(12)) {
      throw new IllegalStateException("trm.TRM_ADDRESS_HI already set");
    } else {
      filled.set(12);
    }

    if (b >= 4294967296L) {
      throw new IllegalArgumentException("trmAddressHi has invalid value (" + b + ")");
    }
    trmAddressHi.put((byte) (b >> 24));
    trmAddressHi.put((byte) (b >> 16));
    trmAddressHi.put((byte) (b >> 8));
    trmAddressHi.put((byte) b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("trm.ACC_HI has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("trm.ACC_LO has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("trm.ACC_T has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("trm.BYTE_HI has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("trm.BYTE_LO has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("trm.CT has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("trm.IS_PRECOMPILE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("trm.ONE has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("trm.PBIT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("trm.RAW_ADDRESS_HI has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("trm.RAW_ADDRESS_LO has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("trm.STAMP has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("trm.TRM_ADDRESS_HI has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      accHi.position(accHi.position() + 16);
    }

    if (!filled.get(1)) {
      accLo.position(accLo.position() + 16);
    }

    if (!filled.get(2)) {
      accT.position(accT.position() + 4);
    }

    if (!filled.get(3)) {
      byteHi.position(byteHi.position() + 1);
    }

    if (!filled.get(4)) {
      byteLo.position(byteLo.position() + 1);
    }

    if (!filled.get(5)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(6)) {
      isPrecompile.position(isPrecompile.position() + 1);
    }

    if (!filled.get(7)) {
      one.position(one.position() + 1);
    }

    if (!filled.get(8)) {
      pbit.position(pbit.position() + 1);
    }

    if (!filled.get(9)) {
      rawAddressHi.position(rawAddressHi.position() + 16);
    }

    if (!filled.get(10)) {
      rawAddressLo.position(rawAddressLo.position() + 16);
    }

    if (!filled.get(11)) {
      stamp.position(stamp.position() + 3);
    }

    if (!filled.get(12)) {
      trmAddressHi.position(trmAddressHi.position() + 4);
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
