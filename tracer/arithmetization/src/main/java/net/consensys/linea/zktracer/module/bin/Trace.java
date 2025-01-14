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
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isAnd;
  private final MappedByteBuffer isByte;
  private final MappedByteBuffer isNot;
  private final MappedByteBuffer isOr;
  private final MappedByteBuffer isSignextend;
  private final MappedByteBuffer isXor;
  private final MappedByteBuffer low4;
  private final MappedByteBuffer neg;
  private final MappedByteBuffer pivot;
  private final MappedByteBuffer resultHi;
  private final MappedByteBuffer resultLo;
  private final MappedByteBuffer small;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer xxxByteHi;
  private final MappedByteBuffer xxxByteLo;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("bin.ACC_1", 16, length));
      headers.add(new ColumnHeader("bin.ACC_2", 16, length));
      headers.add(new ColumnHeader("bin.ACC_3", 16, length));
      headers.add(new ColumnHeader("bin.ACC_4", 16, length));
      headers.add(new ColumnHeader("bin.ACC_5", 16, length));
      headers.add(new ColumnHeader("bin.ACC_6", 16, length));
      headers.add(new ColumnHeader("bin.ARGUMENT_1_HI", 16, length));
      headers.add(new ColumnHeader("bin.ARGUMENT_1_LO", 16, length));
      headers.add(new ColumnHeader("bin.ARGUMENT_2_HI", 16, length));
      headers.add(new ColumnHeader("bin.ARGUMENT_2_LO", 16, length));
      headers.add(new ColumnHeader("bin.BIT_1", 1, length));
      headers.add(new ColumnHeader("bin.BIT_B_4", 1, length));
      headers.add(new ColumnHeader("bin.BITS", 1, length));
      headers.add(new ColumnHeader("bin.BYTE_1", 1, length));
      headers.add(new ColumnHeader("bin.BYTE_2", 1, length));
      headers.add(new ColumnHeader("bin.BYTE_3", 1, length));
      headers.add(new ColumnHeader("bin.BYTE_4", 1, length));
      headers.add(new ColumnHeader("bin.BYTE_5", 1, length));
      headers.add(new ColumnHeader("bin.BYTE_6", 1, length));
      headers.add(new ColumnHeader("bin.COUNTER", 1, length));
      headers.add(new ColumnHeader("bin.CT_MAX", 1, length));
      headers.add(new ColumnHeader("bin.INST", 1, length));
      headers.add(new ColumnHeader("bin.IS_AND", 1, length));
      headers.add(new ColumnHeader("bin.IS_BYTE", 1, length));
      headers.add(new ColumnHeader("bin.IS_NOT", 1, length));
      headers.add(new ColumnHeader("bin.IS_OR", 1, length));
      headers.add(new ColumnHeader("bin.IS_SIGNEXTEND", 1, length));
      headers.add(new ColumnHeader("bin.IS_XOR", 1, length));
      headers.add(new ColumnHeader("bin.LOW_4", 1, length));
      headers.add(new ColumnHeader("bin.NEG", 1, length));
      headers.add(new ColumnHeader("bin.PIVOT", 1, length));
      headers.add(new ColumnHeader("bin.RESULT_HI", 16, length));
      headers.add(new ColumnHeader("bin.RESULT_LO", 16, length));
      headers.add(new ColumnHeader("bin.SMALL", 1, length));
      headers.add(new ColumnHeader("bin.STAMP", 4, length));
      headers.add(new ColumnHeader("bin.XXX_BYTE_HI", 1, length));
      headers.add(new ColumnHeader("bin.XXX_BYTE_LO", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
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
    this.ctMax = buffers.get(20);
    this.inst = buffers.get(21);
    this.isAnd = buffers.get(22);
    this.isByte = buffers.get(23);
    this.isNot = buffers.get(24);
    this.isOr = buffers.get(25);
    this.isSignextend = buffers.get(26);
    this.isXor = buffers.get(27);
    this.low4 = buffers.get(28);
    this.neg = buffers.get(29);
    this.pivot = buffers.get(30);
    this.resultHi = buffers.get(31);
    this.resultLo = buffers.get(32);
    this.small = buffers.get(33);
    this.stamp = buffers.get(34);
    this.xxxByteHi = buffers.get(35);
    this.xxxByteLo = buffers.get(36);
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

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ACC_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc1.put(bs.get(j)); }

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("bin.ACC_2 already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ACC_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc2.put(bs.get(j)); }

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("bin.ACC_3 already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ACC_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc3.put(bs.get(j)); }

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("bin.ACC_4 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ACC_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc4.put(bs.get(j)); }

    return this;
  }

  public Trace acc5(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("bin.ACC_5 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ACC_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc5.put(bs.get(j)); }

    return this;
  }

  public Trace acc6(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("bin.ACC_6 already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ACC_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc6.put(bs.get(j)); }

    return this;
  }

  public Trace argument1Hi(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("bin.ARGUMENT_1_HI already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ARGUMENT_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace argument1Lo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ARGUMENT_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace argument2Hi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ARGUMENT_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace argument2Lo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.ARGUMENT_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument2Lo.put(bs.get(j)); }

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

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("bin.CT_MAX already set");
    } else {
      filled.set(20);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("bin.INST already set");
    } else {
      filled.set(21);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isAnd(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("bin.IS_AND already set");
    } else {
      filled.set(22);
    }

    isAnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isByte(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("bin.IS_BYTE already set");
    } else {
      filled.set(23);
    }

    isByte.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isNot(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("bin.IS_NOT already set");
    } else {
      filled.set(24);
    }

    isNot.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isOr(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("bin.IS_OR already set");
    } else {
      filled.set(25);
    }

    isOr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSignextend(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("bin.IS_SIGNEXTEND already set");
    } else {
      filled.set(26);
    }

    isSignextend.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isXor(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("bin.IS_XOR already set");
    } else {
      filled.set(27);
    }

    isXor.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace low4(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("bin.LOW_4 already set");
    } else {
      filled.set(28);
    }

    low4.put(b.toByte());

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

  public Trace pivot(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("bin.PIVOT already set");
    } else {
      filled.set(30);
    }

    pivot.put(b.toByte());

    return this;
  }

  public Trace resultHi(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("bin.RESULT_HI already set");
    } else {
      filled.set(31);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.RESULT_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resultHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resultHi.put(bs.get(j)); }

    return this;
  }

  public Trace resultLo(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("bin.RESULT_LO already set");
    } else {
      filled.set(32);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("bin.RESULT_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resultLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resultLo.put(bs.get(j)); }

    return this;
  }

  public Trace small(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("bin.SMALL already set");
    } else {
      filled.set(33);
    }

    small.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(34)) {
      throw new IllegalStateException("bin.STAMP already set");
    } else {
      filled.set(34);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("bin.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 24));
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace xxxByteHi(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("bin.XXX_BYTE_HI already set");
    } else {
      filled.set(35);
    }

    xxxByteHi.put(b.toByte());

    return this;
  }

  public Trace xxxByteLo(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("bin.XXX_BYTE_LO already set");
    } else {
      filled.set(36);
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
      throw new IllegalStateException("bin.CT_MAX has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("bin.INST has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("bin.IS_AND has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("bin.IS_BYTE has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("bin.IS_NOT has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("bin.IS_OR has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("bin.IS_SIGNEXTEND has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("bin.IS_XOR has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("bin.LOW_4 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("bin.NEG has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("bin.PIVOT has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("bin.RESULT_HI has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("bin.RESULT_LO has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("bin.SMALL has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("bin.STAMP has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("bin.XXX_BYTE_HI has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("bin.XXX_BYTE_LO has not been filled");
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
      acc3.position(acc3.position() + 16);
    }

    if (!filled.get(3)) {
      acc4.position(acc4.position() + 16);
    }

    if (!filled.get(4)) {
      acc5.position(acc5.position() + 16);
    }

    if (!filled.get(5)) {
      acc6.position(acc6.position() + 16);
    }

    if (!filled.get(6)) {
      argument1Hi.position(argument1Hi.position() + 16);
    }

    if (!filled.get(7)) {
      argument1Lo.position(argument1Lo.position() + 16);
    }

    if (!filled.get(8)) {
      argument2Hi.position(argument2Hi.position() + 16);
    }

    if (!filled.get(9)) {
      argument2Lo.position(argument2Lo.position() + 16);
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
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(21)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(22)) {
      isAnd.position(isAnd.position() + 1);
    }

    if (!filled.get(23)) {
      isByte.position(isByte.position() + 1);
    }

    if (!filled.get(24)) {
      isNot.position(isNot.position() + 1);
    }

    if (!filled.get(25)) {
      isOr.position(isOr.position() + 1);
    }

    if (!filled.get(26)) {
      isSignextend.position(isSignextend.position() + 1);
    }

    if (!filled.get(27)) {
      isXor.position(isXor.position() + 1);
    }

    if (!filled.get(28)) {
      low4.position(low4.position() + 1);
    }

    if (!filled.get(29)) {
      neg.position(neg.position() + 1);
    }

    if (!filled.get(30)) {
      pivot.position(pivot.position() + 1);
    }

    if (!filled.get(31)) {
      resultHi.position(resultHi.position() + 16);
    }

    if (!filled.get(32)) {
      resultLo.position(resultLo.position() + 16);
    }

    if (!filled.get(33)) {
      small.position(small.position() + 1);
    }

    if (!filled.get(34)) {
      stamp.position(stamp.position() + 4);
    }

    if (!filled.get(35)) {
      xxxByteHi.position(xxxByteHi.position() + 1);
    }

    if (!filled.get(36)) {
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
