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

package net.consensys.linea.zktracer.module.add;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
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
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer overflow;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer stamp;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("add.ACC_1", 16, length));
      headers.add(new ColumnHeader("add.ACC_2", 16, length));
      headers.add(new ColumnHeader("add.ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("add.ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("add.ARG_2_HI", 16, length));
      headers.add(new ColumnHeader("add.ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("add.BYTE_1", 1, length));
      headers.add(new ColumnHeader("add.BYTE_2", 1, length));
      headers.add(new ColumnHeader("add.CT", 1, length));
      headers.add(new ColumnHeader("add.CT_MAX", 1, length));
      headers.add(new ColumnHeader("add.INST", 1, length));
      headers.add(new ColumnHeader("add.OVERFLOW", 1, length));
      headers.add(new ColumnHeader("add.RES_HI", 16, length));
      headers.add(new ColumnHeader("add.RES_LO", 16, length));
      headers.add(new ColumnHeader("add.STAMP", 4, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.arg1Hi = buffers.get(2);
    this.arg1Lo = buffers.get(3);
    this.arg2Hi = buffers.get(4);
    this.arg2Lo = buffers.get(5);
    this.byte1 = buffers.get(6);
    this.byte2 = buffers.get(7);
    this.ct = buffers.get(8);
    this.ctMax = buffers.get(9);
    this.inst = buffers.get(10);
    this.overflow = buffers.get(11);
    this.resHi = buffers.get(12);
    this.resLo = buffers.get(13);
    this.stamp = buffers.get(14);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("add.ACC_1 already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.ACC_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc1.put(bs.get(j)); }

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("add.ACC_2 already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.ACC_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc2.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("add.ARG_1_HI already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("add.ARG_1_LO already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("add.ARG_2_HI already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.ARG_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("add.ARG_2_LO already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("add.BYTE_1 already set");
    } else {
      filled.set(6);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("add.BYTE_2 already set");
    } else {
      filled.set(7);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace ct(final UnsignedByte b) {
    if (filled.get(8)) {
      throw new IllegalStateException("add.CT already set");
    } else {
      filled.set(8);
    }

    ct.put(b.toByte());

    return this;
  }

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("add.CT_MAX already set");
    } else {
      filled.set(9);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("add.INST already set");
    } else {
      filled.set(10);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace overflow(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("add.OVERFLOW already set");
    } else {
      filled.set(11);
    }

    overflow.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("add.RES_HI already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.RES_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resHi.put(bs.get(j)); }

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("add.RES_LO already set");
    } else {
      filled.set(13);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("add.RES_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resLo.put(bs.get(j)); }

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(14)) {
      throw new IllegalStateException("add.STAMP already set");
    } else {
      filled.set(14);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("add.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 24));
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("add.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("add.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("add.ARG_1_HI has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("add.ARG_1_LO has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("add.ARG_2_HI has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("add.ARG_2_LO has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("add.BYTE_1 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("add.BYTE_2 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("add.CT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("add.CT_MAX has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("add.INST has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("add.OVERFLOW has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("add.RES_HI has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("add.RES_LO has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("add.STAMP has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc1.position(acc1.position() + 16);
    }

    if (!filled.get(1)) {
      acc2.position(acc2.position() + 16);
    }

    if (!filled.get(2)) {
      arg1Hi.position(arg1Hi.position() + 16);
    }

    if (!filled.get(3)) {
      arg1Lo.position(arg1Lo.position() + 16);
    }

    if (!filled.get(4)) {
      arg2Hi.position(arg2Hi.position() + 16);
    }

    if (!filled.get(5)) {
      arg2Lo.position(arg2Lo.position() + 16);
    }

    if (!filled.get(6)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(7)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(8)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(9)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(10)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(11)) {
      overflow.position(overflow.position() + 1);
    }

    if (!filled.get(12)) {
      resHi.position(resHi.position() + 16);
    }

    if (!filled.get(13)) {
      resLo.position(resLo.position() + 16);
    }

    if (!filled.get(14)) {
      stamp.position(stamp.position() + 4);
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
