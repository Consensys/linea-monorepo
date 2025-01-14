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

package net.consensys.linea.zktracer.module.mul;

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

  private final MappedByteBuffer accA0;
  private final MappedByteBuffer accA1;
  private final MappedByteBuffer accA2;
  private final MappedByteBuffer accA3;
  private final MappedByteBuffer accB0;
  private final MappedByteBuffer accB1;
  private final MappedByteBuffer accB2;
  private final MappedByteBuffer accB3;
  private final MappedByteBuffer accC0;
  private final MappedByteBuffer accC1;
  private final MappedByteBuffer accC2;
  private final MappedByteBuffer accC3;
  private final MappedByteBuffer accH0;
  private final MappedByteBuffer accH1;
  private final MappedByteBuffer accH2;
  private final MappedByteBuffer accH3;
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer bitNum;
  private final MappedByteBuffer bits;
  private final MappedByteBuffer byteA0;
  private final MappedByteBuffer byteA1;
  private final MappedByteBuffer byteA2;
  private final MappedByteBuffer byteA3;
  private final MappedByteBuffer byteB0;
  private final MappedByteBuffer byteB1;
  private final MappedByteBuffer byteB2;
  private final MappedByteBuffer byteB3;
  private final MappedByteBuffer byteC0;
  private final MappedByteBuffer byteC1;
  private final MappedByteBuffer byteC2;
  private final MappedByteBuffer byteC3;
  private final MappedByteBuffer byteH0;
  private final MappedByteBuffer byteH1;
  private final MappedByteBuffer byteH2;
  private final MappedByteBuffer byteH3;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer exponentBit;
  private final MappedByteBuffer exponentBitAccumulator;
  private final MappedByteBuffer exponentBitSource;
  private final MappedByteBuffer instruction;
  private final MappedByteBuffer mulStamp;
  private final MappedByteBuffer oli;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer resultVanishes;
  private final MappedByteBuffer squareAndMultiply;
  private final MappedByteBuffer tinyBase;
  private final MappedByteBuffer tinyExponent;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("mul.ACC_A_0", 8, length));
      headers.add(new ColumnHeader("mul.ACC_A_1", 8, length));
      headers.add(new ColumnHeader("mul.ACC_A_2", 8, length));
      headers.add(new ColumnHeader("mul.ACC_A_3", 8, length));
      headers.add(new ColumnHeader("mul.ACC_B_0", 8, length));
      headers.add(new ColumnHeader("mul.ACC_B_1", 8, length));
      headers.add(new ColumnHeader("mul.ACC_B_2", 8, length));
      headers.add(new ColumnHeader("mul.ACC_B_3", 8, length));
      headers.add(new ColumnHeader("mul.ACC_C_0", 8, length));
      headers.add(new ColumnHeader("mul.ACC_C_1", 8, length));
      headers.add(new ColumnHeader("mul.ACC_C_2", 8, length));
      headers.add(new ColumnHeader("mul.ACC_C_3", 8, length));
      headers.add(new ColumnHeader("mul.ACC_H_0", 8, length));
      headers.add(new ColumnHeader("mul.ACC_H_1", 8, length));
      headers.add(new ColumnHeader("mul.ACC_H_2", 8, length));
      headers.add(new ColumnHeader("mul.ACC_H_3", 8, length));
      headers.add(new ColumnHeader("mul.ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("mul.ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("mul.ARG_2_HI", 16, length));
      headers.add(new ColumnHeader("mul.ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("mul.BIT_NUM", 1, length));
      headers.add(new ColumnHeader("mul.BITS", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_A_0", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_A_1", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_A_2", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_A_3", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_B_0", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_B_1", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_B_2", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_B_3", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_C_0", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_C_1", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_C_2", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_C_3", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_H_0", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_H_1", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_H_2", 1, length));
      headers.add(new ColumnHeader("mul.BYTE_H_3", 1, length));
      headers.add(new ColumnHeader("mul.COUNTER", 1, length));
      headers.add(new ColumnHeader("mul.EXPONENT_BIT", 1, length));
      headers.add(new ColumnHeader("mul.EXPONENT_BIT_ACCUMULATOR", 16, length));
      headers.add(new ColumnHeader("mul.EXPONENT_BIT_SOURCE", 1, length));
      headers.add(new ColumnHeader("mul.INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("mul.MUL_STAMP", 4, length));
      headers.add(new ColumnHeader("mul.OLI", 1, length));
      headers.add(new ColumnHeader("mul.RES_HI", 16, length));
      headers.add(new ColumnHeader("mul.RES_LO", 16, length));
      headers.add(new ColumnHeader("mul.RESULT_VANISHES", 1, length));
      headers.add(new ColumnHeader("mul.SQUARE_AND_MULTIPLY", 1, length));
      headers.add(new ColumnHeader("mul.TINY_BASE", 1, length));
      headers.add(new ColumnHeader("mul.TINY_EXPONENT", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.accA0 = buffers.get(0);
    this.accA1 = buffers.get(1);
    this.accA2 = buffers.get(2);
    this.accA3 = buffers.get(3);
    this.accB0 = buffers.get(4);
    this.accB1 = buffers.get(5);
    this.accB2 = buffers.get(6);
    this.accB3 = buffers.get(7);
    this.accC0 = buffers.get(8);
    this.accC1 = buffers.get(9);
    this.accC2 = buffers.get(10);
    this.accC3 = buffers.get(11);
    this.accH0 = buffers.get(12);
    this.accH1 = buffers.get(13);
    this.accH2 = buffers.get(14);
    this.accH3 = buffers.get(15);
    this.arg1Hi = buffers.get(16);
    this.arg1Lo = buffers.get(17);
    this.arg2Hi = buffers.get(18);
    this.arg2Lo = buffers.get(19);
    this.bitNum = buffers.get(20);
    this.bits = buffers.get(21);
    this.byteA0 = buffers.get(22);
    this.byteA1 = buffers.get(23);
    this.byteA2 = buffers.get(24);
    this.byteA3 = buffers.get(25);
    this.byteB0 = buffers.get(26);
    this.byteB1 = buffers.get(27);
    this.byteB2 = buffers.get(28);
    this.byteB3 = buffers.get(29);
    this.byteC0 = buffers.get(30);
    this.byteC1 = buffers.get(31);
    this.byteC2 = buffers.get(32);
    this.byteC3 = buffers.get(33);
    this.byteH0 = buffers.get(34);
    this.byteH1 = buffers.get(35);
    this.byteH2 = buffers.get(36);
    this.byteH3 = buffers.get(37);
    this.counter = buffers.get(38);
    this.exponentBit = buffers.get(39);
    this.exponentBitAccumulator = buffers.get(40);
    this.exponentBitSource = buffers.get(41);
    this.instruction = buffers.get(42);
    this.mulStamp = buffers.get(43);
    this.oli = buffers.get(44);
    this.resHi = buffers.get(45);
    this.resLo = buffers.get(46);
    this.resultVanishes = buffers.get(47);
    this.squareAndMultiply = buffers.get(48);
    this.tinyBase = buffers.get(49);
    this.tinyExponent = buffers.get(50);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace accA0(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mul.ACC_A_0 already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_A_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA0.put(bs.get(j)); }

    return this;
  }

  public Trace accA1(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mul.ACC_A_1 already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_A_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA1.put(bs.get(j)); }

    return this;
  }

  public Trace accA2(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mul.ACC_A_2 already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_A_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA2.put(bs.get(j)); }

    return this;
  }

  public Trace accA3(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mul.ACC_A_3 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_A_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA3.put(bs.get(j)); }

    return this;
  }

  public Trace accB0(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mul.ACC_B_0 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_B_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB0.put(bs.get(j)); }

    return this;
  }

  public Trace accB1(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mul.ACC_B_1 already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_B_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB1.put(bs.get(j)); }

    return this;
  }

  public Trace accB2(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mul.ACC_B_2 already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_B_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB2.put(bs.get(j)); }

    return this;
  }

  public Trace accB3(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mul.ACC_B_3 already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_B_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB3.put(bs.get(j)); }

    return this;
  }

  public Trace accC0(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mul.ACC_C_0 already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_C_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC0.put(bs.get(j)); }

    return this;
  }

  public Trace accC1(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mul.ACC_C_1 already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_C_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC1.put(bs.get(j)); }

    return this;
  }

  public Trace accC2(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mul.ACC_C_2 already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_C_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC2.put(bs.get(j)); }

    return this;
  }

  public Trace accC3(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mul.ACC_C_3 already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_C_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC3.put(bs.get(j)); }

    return this;
  }

  public Trace accH0(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mul.ACC_H_0 already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_H_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH0.put(bs.get(j)); }

    return this;
  }

  public Trace accH1(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mul.ACC_H_1 already set");
    } else {
      filled.set(13);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_H_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH1.put(bs.get(j)); }

    return this;
  }

  public Trace accH2(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mul.ACC_H_2 already set");
    } else {
      filled.set(14);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_H_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH2.put(bs.get(j)); }

    return this;
  }

  public Trace accH3(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mul.ACC_H_3 already set");
    } else {
      filled.set(15);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mul.ACC_H_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH3.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mul.ARG_1_HI already set");
    } else {
      filled.set(16);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mul.ARG_1_LO already set");
    } else {
      filled.set(17);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mul.ARG_2_HI already set");
    } else {
      filled.set(18);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.ARG_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mul.ARG_2_LO already set");
    } else {
      filled.set(19);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace bitNum(final long b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mul.BIT_NUM already set");
    } else {
      filled.set(21);
    }

    if(b >= 128L) { throw new IllegalArgumentException("mul.BIT_NUM has invalid value (" + b + ")"); }
    bitNum.put((byte) b);


    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mul.BITS already set");
    } else {
      filled.set(20);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byteA0(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mul.BYTE_A_0 already set");
    } else {
      filled.set(22);
    }

    byteA0.put(b.toByte());

    return this;
  }

  public Trace byteA1(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mul.BYTE_A_1 already set");
    } else {
      filled.set(23);
    }

    byteA1.put(b.toByte());

    return this;
  }

  public Trace byteA2(final UnsignedByte b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mul.BYTE_A_2 already set");
    } else {
      filled.set(24);
    }

    byteA2.put(b.toByte());

    return this;
  }

  public Trace byteA3(final UnsignedByte b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mul.BYTE_A_3 already set");
    } else {
      filled.set(25);
    }

    byteA3.put(b.toByte());

    return this;
  }

  public Trace byteB0(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mul.BYTE_B_0 already set");
    } else {
      filled.set(26);
    }

    byteB0.put(b.toByte());

    return this;
  }

  public Trace byteB1(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mul.BYTE_B_1 already set");
    } else {
      filled.set(27);
    }

    byteB1.put(b.toByte());

    return this;
  }

  public Trace byteB2(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mul.BYTE_B_2 already set");
    } else {
      filled.set(28);
    }

    byteB2.put(b.toByte());

    return this;
  }

  public Trace byteB3(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mul.BYTE_B_3 already set");
    } else {
      filled.set(29);
    }

    byteB3.put(b.toByte());

    return this;
  }

  public Trace byteC0(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mul.BYTE_C_0 already set");
    } else {
      filled.set(30);
    }

    byteC0.put(b.toByte());

    return this;
  }

  public Trace byteC1(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mul.BYTE_C_1 already set");
    } else {
      filled.set(31);
    }

    byteC1.put(b.toByte());

    return this;
  }

  public Trace byteC2(final UnsignedByte b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mul.BYTE_C_2 already set");
    } else {
      filled.set(32);
    }

    byteC2.put(b.toByte());

    return this;
  }

  public Trace byteC3(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mul.BYTE_C_3 already set");
    } else {
      filled.set(33);
    }

    byteC3.put(b.toByte());

    return this;
  }

  public Trace byteH0(final UnsignedByte b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mul.BYTE_H_0 already set");
    } else {
      filled.set(34);
    }

    byteH0.put(b.toByte());

    return this;
  }

  public Trace byteH1(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mul.BYTE_H_1 already set");
    } else {
      filled.set(35);
    }

    byteH1.put(b.toByte());

    return this;
  }

  public Trace byteH2(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mul.BYTE_H_2 already set");
    } else {
      filled.set(36);
    }

    byteH2.put(b.toByte());

    return this;
  }

  public Trace byteH3(final UnsignedByte b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mul.BYTE_H_3 already set");
    } else {
      filled.set(37);
    }

    byteH3.put(b.toByte());

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mul.COUNTER already set");
    } else {
      filled.set(38);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace exponentBit(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mul.EXPONENT_BIT already set");
    } else {
      filled.set(39);
    }

    exponentBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exponentBitAccumulator(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mul.EXPONENT_BIT_ACCUMULATOR already set");
    } else {
      filled.set(40);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.EXPONENT_BIT_ACCUMULATOR has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { exponentBitAccumulator.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { exponentBitAccumulator.put(bs.get(j)); }

    return this;
  }

  public Trace exponentBitSource(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mul.EXPONENT_BIT_SOURCE already set");
    } else {
      filled.set(41);
    }

    exponentBitSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace instruction(final UnsignedByte b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mul.INSTRUCTION already set");
    } else {
      filled.set(42);
    }

    instruction.put(b.toByte());

    return this;
  }

  public Trace mulStamp(final long b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mul.MUL_STAMP already set");
    } else {
      filled.set(43);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mul.MUL_STAMP has invalid value (" + b + ")"); }
    mulStamp.put((byte) (b >> 24));
    mulStamp.put((byte) (b >> 16));
    mulStamp.put((byte) (b >> 8));
    mulStamp.put((byte) b);


    return this;
  }

  public Trace oli(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mul.OLI already set");
    } else {
      filled.set(44);
    }

    oli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mul.RES_HI already set");
    } else {
      filled.set(46);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.RES_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resHi.put(bs.get(j)); }

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mul.RES_LO already set");
    } else {
      filled.set(47);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mul.RES_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resLo.put(bs.get(j)); }

    return this;
  }

  public Trace resultVanishes(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mul.RESULT_VANISHES already set");
    } else {
      filled.set(45);
    }

    resultVanishes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace squareAndMultiply(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mul.SQUARE_AND_MULTIPLY already set");
    } else {
      filled.set(48);
    }

    squareAndMultiply.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace tinyBase(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mul.TINY_BASE already set");
    } else {
      filled.set(49);
    }

    tinyBase.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace tinyExponent(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mul.TINY_EXPONENT already set");
    } else {
      filled.set(50);
    }

    tinyExponent.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mul.ACC_A_0 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mul.ACC_A_1 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mul.ACC_A_2 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mul.ACC_A_3 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mul.ACC_B_0 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mul.ACC_B_1 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mul.ACC_B_2 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mul.ACC_B_3 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mul.ACC_C_0 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mul.ACC_C_1 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mul.ACC_C_2 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mul.ACC_C_3 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mul.ACC_H_0 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mul.ACC_H_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mul.ACC_H_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mul.ACC_H_3 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mul.ARG_1_HI has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mul.ARG_1_LO has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mul.ARG_2_HI has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mul.ARG_2_LO has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mul.BIT_NUM has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mul.BITS has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mul.BYTE_A_0 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mul.BYTE_A_1 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mul.BYTE_A_2 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mul.BYTE_A_3 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mul.BYTE_B_0 has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mul.BYTE_B_1 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mul.BYTE_B_2 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mul.BYTE_B_3 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mul.BYTE_C_0 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mul.BYTE_C_1 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mul.BYTE_C_2 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mul.BYTE_C_3 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mul.BYTE_H_0 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mul.BYTE_H_1 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mul.BYTE_H_2 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mul.BYTE_H_3 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mul.COUNTER has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mul.EXPONENT_BIT has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mul.EXPONENT_BIT_ACCUMULATOR has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mul.EXPONENT_BIT_SOURCE has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mul.INSTRUCTION has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mul.MUL_STAMP has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mul.OLI has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mul.RES_HI has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mul.RES_LO has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mul.RESULT_VANISHES has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mul.SQUARE_AND_MULTIPLY has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mul.TINY_BASE has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mul.TINY_EXPONENT has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      accA0.position(accA0.position() + 8);
    }

    if (!filled.get(1)) {
      accA1.position(accA1.position() + 8);
    }

    if (!filled.get(2)) {
      accA2.position(accA2.position() + 8);
    }

    if (!filled.get(3)) {
      accA3.position(accA3.position() + 8);
    }

    if (!filled.get(4)) {
      accB0.position(accB0.position() + 8);
    }

    if (!filled.get(5)) {
      accB1.position(accB1.position() + 8);
    }

    if (!filled.get(6)) {
      accB2.position(accB2.position() + 8);
    }

    if (!filled.get(7)) {
      accB3.position(accB3.position() + 8);
    }

    if (!filled.get(8)) {
      accC0.position(accC0.position() + 8);
    }

    if (!filled.get(9)) {
      accC1.position(accC1.position() + 8);
    }

    if (!filled.get(10)) {
      accC2.position(accC2.position() + 8);
    }

    if (!filled.get(11)) {
      accC3.position(accC3.position() + 8);
    }

    if (!filled.get(12)) {
      accH0.position(accH0.position() + 8);
    }

    if (!filled.get(13)) {
      accH1.position(accH1.position() + 8);
    }

    if (!filled.get(14)) {
      accH2.position(accH2.position() + 8);
    }

    if (!filled.get(15)) {
      accH3.position(accH3.position() + 8);
    }

    if (!filled.get(16)) {
      arg1Hi.position(arg1Hi.position() + 16);
    }

    if (!filled.get(17)) {
      arg1Lo.position(arg1Lo.position() + 16);
    }

    if (!filled.get(18)) {
      arg2Hi.position(arg2Hi.position() + 16);
    }

    if (!filled.get(19)) {
      arg2Lo.position(arg2Lo.position() + 16);
    }

    if (!filled.get(21)) {
      bitNum.position(bitNum.position() + 1);
    }

    if (!filled.get(20)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(22)) {
      byteA0.position(byteA0.position() + 1);
    }

    if (!filled.get(23)) {
      byteA1.position(byteA1.position() + 1);
    }

    if (!filled.get(24)) {
      byteA2.position(byteA2.position() + 1);
    }

    if (!filled.get(25)) {
      byteA3.position(byteA3.position() + 1);
    }

    if (!filled.get(26)) {
      byteB0.position(byteB0.position() + 1);
    }

    if (!filled.get(27)) {
      byteB1.position(byteB1.position() + 1);
    }

    if (!filled.get(28)) {
      byteB2.position(byteB2.position() + 1);
    }

    if (!filled.get(29)) {
      byteB3.position(byteB3.position() + 1);
    }

    if (!filled.get(30)) {
      byteC0.position(byteC0.position() + 1);
    }

    if (!filled.get(31)) {
      byteC1.position(byteC1.position() + 1);
    }

    if (!filled.get(32)) {
      byteC2.position(byteC2.position() + 1);
    }

    if (!filled.get(33)) {
      byteC3.position(byteC3.position() + 1);
    }

    if (!filled.get(34)) {
      byteH0.position(byteH0.position() + 1);
    }

    if (!filled.get(35)) {
      byteH1.position(byteH1.position() + 1);
    }

    if (!filled.get(36)) {
      byteH2.position(byteH2.position() + 1);
    }

    if (!filled.get(37)) {
      byteH3.position(byteH3.position() + 1);
    }

    if (!filled.get(38)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(39)) {
      exponentBit.position(exponentBit.position() + 1);
    }

    if (!filled.get(40)) {
      exponentBitAccumulator.position(exponentBitAccumulator.position() + 16);
    }

    if (!filled.get(41)) {
      exponentBitSource.position(exponentBitSource.position() + 1);
    }

    if (!filled.get(42)) {
      instruction.position(instruction.position() + 1);
    }

    if (!filled.get(43)) {
      mulStamp.position(mulStamp.position() + 4);
    }

    if (!filled.get(44)) {
      oli.position(oli.position() + 1);
    }

    if (!filled.get(46)) {
      resHi.position(resHi.position() + 16);
    }

    if (!filled.get(47)) {
      resLo.position(resLo.position() + 16);
    }

    if (!filled.get(45)) {
      resultVanishes.position(resultVanishes.position() + 1);
    }

    if (!filled.get(48)) {
      squareAndMultiply.position(squareAndMultiply.position() + 1);
    }

    if (!filled.get(49)) {
      tinyBase.position(tinyBase.position() + 1);
    }

    if (!filled.get(50)) {
      tinyExponent.position(tinyExponent.position() + 1);
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
