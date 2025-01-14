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

package net.consensys.linea.zktracer.module.wcp;

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
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
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
  private final MappedByteBuffer isEq;
  private final MappedByteBuffer isGeq;
  private final MappedByteBuffer isGt;
  private final MappedByteBuffer isIszero;
  private final MappedByteBuffer isLeq;
  private final MappedByteBuffer isLt;
  private final MappedByteBuffer isSgt;
  private final MappedByteBuffer isSlt;
  private final MappedByteBuffer neg1;
  private final MappedByteBuffer neg2;
  private final MappedByteBuffer oneLineInstruction;
  private final MappedByteBuffer result;
  private final MappedByteBuffer variableLengthInstruction;
  private final MappedByteBuffer wordComparisonStamp;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("wcp.ACC_1", 16, length));
      headers.add(new ColumnHeader("wcp.ACC_2", 16, length));
      headers.add(new ColumnHeader("wcp.ACC_3", 16, length));
      headers.add(new ColumnHeader("wcp.ACC_4", 16, length));
      headers.add(new ColumnHeader("wcp.ACC_5", 16, length));
      headers.add(new ColumnHeader("wcp.ACC_6", 16, length));
      headers.add(new ColumnHeader("wcp.ARGUMENT_1_HI", 16, length));
      headers.add(new ColumnHeader("wcp.ARGUMENT_1_LO", 16, length));
      headers.add(new ColumnHeader("wcp.ARGUMENT_2_HI", 16, length));
      headers.add(new ColumnHeader("wcp.ARGUMENT_2_LO", 16, length));
      headers.add(new ColumnHeader("wcp.BIT_1", 1, length));
      headers.add(new ColumnHeader("wcp.BIT_2", 1, length));
      headers.add(new ColumnHeader("wcp.BIT_3", 1, length));
      headers.add(new ColumnHeader("wcp.BIT_4", 1, length));
      headers.add(new ColumnHeader("wcp.BITS", 1, length));
      headers.add(new ColumnHeader("wcp.BYTE_1", 1, length));
      headers.add(new ColumnHeader("wcp.BYTE_2", 1, length));
      headers.add(new ColumnHeader("wcp.BYTE_3", 1, length));
      headers.add(new ColumnHeader("wcp.BYTE_4", 1, length));
      headers.add(new ColumnHeader("wcp.BYTE_5", 1, length));
      headers.add(new ColumnHeader("wcp.BYTE_6", 1, length));
      headers.add(new ColumnHeader("wcp.COUNTER", 1, length));
      headers.add(new ColumnHeader("wcp.CT_MAX", 1, length));
      headers.add(new ColumnHeader("wcp.INST", 1, length));
      headers.add(new ColumnHeader("wcp.IS_EQ", 1, length));
      headers.add(new ColumnHeader("wcp.IS_GEQ", 1, length));
      headers.add(new ColumnHeader("wcp.IS_GT", 1, length));
      headers.add(new ColumnHeader("wcp.IS_ISZERO", 1, length));
      headers.add(new ColumnHeader("wcp.IS_LEQ", 1, length));
      headers.add(new ColumnHeader("wcp.IS_LT", 1, length));
      headers.add(new ColumnHeader("wcp.IS_SGT", 1, length));
      headers.add(new ColumnHeader("wcp.IS_SLT", 1, length));
      headers.add(new ColumnHeader("wcp.NEG_1", 1, length));
      headers.add(new ColumnHeader("wcp.NEG_2", 1, length));
      headers.add(new ColumnHeader("wcp.ONE_LINE_INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("wcp.RESULT", 1, length));
      headers.add(new ColumnHeader("wcp.VARIABLE_LENGTH_INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("wcp.WORD_COMPARISON_STAMP", 4, length));
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
    this.bit2 = buffers.get(11);
    this.bit3 = buffers.get(12);
    this.bit4 = buffers.get(13);
    this.bits = buffers.get(14);
    this.byte1 = buffers.get(15);
    this.byte2 = buffers.get(16);
    this.byte3 = buffers.get(17);
    this.byte4 = buffers.get(18);
    this.byte5 = buffers.get(19);
    this.byte6 = buffers.get(20);
    this.counter = buffers.get(21);
    this.ctMax = buffers.get(22);
    this.inst = buffers.get(23);
    this.isEq = buffers.get(24);
    this.isGeq = buffers.get(25);
    this.isGt = buffers.get(26);
    this.isIszero = buffers.get(27);
    this.isLeq = buffers.get(28);
    this.isLt = buffers.get(29);
    this.isSgt = buffers.get(30);
    this.isSlt = buffers.get(31);
    this.neg1 = buffers.get(32);
    this.neg2 = buffers.get(33);
    this.oneLineInstruction = buffers.get(34);
    this.result = buffers.get(35);
    this.variableLengthInstruction = buffers.get(36);
    this.wordComparisonStamp = buffers.get(37);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("wcp.ACC_1 already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ACC_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc1.put(bs.get(j)); }

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("wcp.ACC_2 already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ACC_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc2.put(bs.get(j)); }

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("wcp.ACC_3 already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ACC_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc3.put(bs.get(j)); }

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("wcp.ACC_4 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ACC_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc4.put(bs.get(j)); }

    return this;
  }

  public Trace acc5(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("wcp.ACC_5 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ACC_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc5.put(bs.get(j)); }

    return this;
  }

  public Trace acc6(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("wcp.ACC_6 already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ACC_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc6.put(bs.get(j)); }

    return this;
  }

  public Trace argument1Hi(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("wcp.ARGUMENT_1_HI already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ARGUMENT_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace argument1Lo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("wcp.ARGUMENT_1_LO already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ARGUMENT_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace argument2Hi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("wcp.ARGUMENT_2_HI already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ARGUMENT_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace argument2Lo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("wcp.ARGUMENT_2_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("wcp.ARGUMENT_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argument2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argument2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("wcp.BIT_1 already set");
    } else {
      filled.set(11);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("wcp.BIT_2 already set");
    } else {
      filled.set(12);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("wcp.BIT_3 already set");
    } else {
      filled.set(13);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("wcp.BIT_4 already set");
    } else {
      filled.set(14);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("wcp.BITS already set");
    } else {
      filled.set(10);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("wcp.BYTE_1 already set");
    } else {
      filled.set(15);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("wcp.BYTE_2 already set");
    } else {
      filled.set(16);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("wcp.BYTE_3 already set");
    } else {
      filled.set(17);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("wcp.BYTE_4 already set");
    } else {
      filled.set(18);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("wcp.BYTE_5 already set");
    } else {
      filled.set(19);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace byte6(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("wcp.BYTE_6 already set");
    } else {
      filled.set(20);
    }

    byte6.put(b.toByte());

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("wcp.COUNTER already set");
    } else {
      filled.set(21);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("wcp.CT_MAX already set");
    } else {
      filled.set(22);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("wcp.INST already set");
    } else {
      filled.set(23);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isEq(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("wcp.IS_EQ already set");
    } else {
      filled.set(24);
    }

    isEq.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isGeq(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("wcp.IS_GEQ already set");
    } else {
      filled.set(25);
    }

    isGeq.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isGt(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("wcp.IS_GT already set");
    } else {
      filled.set(26);
    }

    isGt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isIszero(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("wcp.IS_ISZERO already set");
    } else {
      filled.set(27);
    }

    isIszero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLeq(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("wcp.IS_LEQ already set");
    } else {
      filled.set(28);
    }

    isLeq.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLt(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("wcp.IS_LT already set");
    } else {
      filled.set(29);
    }

    isLt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSgt(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("wcp.IS_SGT already set");
    } else {
      filled.set(30);
    }

    isSgt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSlt(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("wcp.IS_SLT already set");
    } else {
      filled.set(31);
    }

    isSlt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace neg1(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("wcp.NEG_1 already set");
    } else {
      filled.set(32);
    }

    neg1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace neg2(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("wcp.NEG_2 already set");
    } else {
      filled.set(33);
    }

    neg2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oneLineInstruction(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("wcp.ONE_LINE_INSTRUCTION already set");
    } else {
      filled.set(34);
    }

    oneLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace result(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("wcp.RESULT already set");
    } else {
      filled.set(35);
    }

    result.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace variableLengthInstruction(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("wcp.VARIABLE_LENGTH_INSTRUCTION already set");
    } else {
      filled.set(36);
    }

    variableLengthInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace wordComparisonStamp(final long b) {
    if (filled.get(37)) {
      throw new IllegalStateException("wcp.WORD_COMPARISON_STAMP already set");
    } else {
      filled.set(37);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("wcp.WORD_COMPARISON_STAMP has invalid value (" + b + ")"); }
    wordComparisonStamp.put((byte) (b >> 24));
    wordComparisonStamp.put((byte) (b >> 16));
    wordComparisonStamp.put((byte) (b >> 8));
    wordComparisonStamp.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("wcp.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("wcp.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("wcp.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("wcp.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("wcp.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("wcp.ACC_6 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("wcp.ARGUMENT_1_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("wcp.ARGUMENT_1_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("wcp.ARGUMENT_2_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("wcp.ARGUMENT_2_LO has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("wcp.BIT_1 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("wcp.BIT_2 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("wcp.BIT_3 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("wcp.BIT_4 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("wcp.BITS has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("wcp.BYTE_1 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("wcp.BYTE_2 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("wcp.BYTE_3 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("wcp.BYTE_4 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("wcp.BYTE_5 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("wcp.BYTE_6 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("wcp.COUNTER has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("wcp.CT_MAX has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("wcp.INST has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("wcp.IS_EQ has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("wcp.IS_GEQ has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("wcp.IS_GT has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("wcp.IS_ISZERO has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("wcp.IS_LEQ has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("wcp.IS_LT has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("wcp.IS_SGT has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("wcp.IS_SLT has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("wcp.NEG_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("wcp.NEG_2 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("wcp.ONE_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("wcp.RESULT has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("wcp.VARIABLE_LENGTH_INSTRUCTION has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("wcp.WORD_COMPARISON_STAMP has not been filled");
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
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(13)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(14)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(10)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(15)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(16)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(17)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(18)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(19)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(20)) {
      byte6.position(byte6.position() + 1);
    }

    if (!filled.get(21)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(22)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(23)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(24)) {
      isEq.position(isEq.position() + 1);
    }

    if (!filled.get(25)) {
      isGeq.position(isGeq.position() + 1);
    }

    if (!filled.get(26)) {
      isGt.position(isGt.position() + 1);
    }

    if (!filled.get(27)) {
      isIszero.position(isIszero.position() + 1);
    }

    if (!filled.get(28)) {
      isLeq.position(isLeq.position() + 1);
    }

    if (!filled.get(29)) {
      isLt.position(isLt.position() + 1);
    }

    if (!filled.get(30)) {
      isSgt.position(isSgt.position() + 1);
    }

    if (!filled.get(31)) {
      isSlt.position(isSlt.position() + 1);
    }

    if (!filled.get(32)) {
      neg1.position(neg1.position() + 1);
    }

    if (!filled.get(33)) {
      neg2.position(neg2.position() + 1);
    }

    if (!filled.get(34)) {
      oneLineInstruction.position(oneLineInstruction.position() + 1);
    }

    if (!filled.get(35)) {
      result.position(result.position() + 1);
    }

    if (!filled.get(36)) {
      variableLengthInstruction.position(variableLengthInstruction.position() + 1);
    }

    if (!filled.get(37)) {
      wordComparisonStamp.position(wordComparisonStamp.position() + 4);
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
