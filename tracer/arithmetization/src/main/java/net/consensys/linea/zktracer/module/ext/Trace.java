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

package net.consensys.linea.zktracer.module.ext;

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
  private final MappedByteBuffer accDelta0;
  private final MappedByteBuffer accDelta1;
  private final MappedByteBuffer accDelta2;
  private final MappedByteBuffer accDelta3;
  private final MappedByteBuffer accH0;
  private final MappedByteBuffer accH1;
  private final MappedByteBuffer accH2;
  private final MappedByteBuffer accH3;
  private final MappedByteBuffer accH4;
  private final MappedByteBuffer accH5;
  private final MappedByteBuffer accI0;
  private final MappedByteBuffer accI1;
  private final MappedByteBuffer accI2;
  private final MappedByteBuffer accI3;
  private final MappedByteBuffer accI4;
  private final MappedByteBuffer accI5;
  private final MappedByteBuffer accI6;
  private final MappedByteBuffer accJ0;
  private final MappedByteBuffer accJ1;
  private final MappedByteBuffer accJ2;
  private final MappedByteBuffer accJ3;
  private final MappedByteBuffer accJ4;
  private final MappedByteBuffer accJ5;
  private final MappedByteBuffer accJ6;
  private final MappedByteBuffer accJ7;
  private final MappedByteBuffer accQ0;
  private final MappedByteBuffer accQ1;
  private final MappedByteBuffer accQ2;
  private final MappedByteBuffer accQ3;
  private final MappedByteBuffer accQ4;
  private final MappedByteBuffer accQ5;
  private final MappedByteBuffer accQ6;
  private final MappedByteBuffer accQ7;
  private final MappedByteBuffer accR0;
  private final MappedByteBuffer accR1;
  private final MappedByteBuffer accR2;
  private final MappedByteBuffer accR3;
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer arg3Hi;
  private final MappedByteBuffer arg3Lo;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
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
  private final MappedByteBuffer byteDelta0;
  private final MappedByteBuffer byteDelta1;
  private final MappedByteBuffer byteDelta2;
  private final MappedByteBuffer byteDelta3;
  private final MappedByteBuffer byteH0;
  private final MappedByteBuffer byteH1;
  private final MappedByteBuffer byteH2;
  private final MappedByteBuffer byteH3;
  private final MappedByteBuffer byteH4;
  private final MappedByteBuffer byteH5;
  private final MappedByteBuffer byteI0;
  private final MappedByteBuffer byteI1;
  private final MappedByteBuffer byteI2;
  private final MappedByteBuffer byteI3;
  private final MappedByteBuffer byteI4;
  private final MappedByteBuffer byteI5;
  private final MappedByteBuffer byteI6;
  private final MappedByteBuffer byteJ0;
  private final MappedByteBuffer byteJ1;
  private final MappedByteBuffer byteJ2;
  private final MappedByteBuffer byteJ3;
  private final MappedByteBuffer byteJ4;
  private final MappedByteBuffer byteJ5;
  private final MappedByteBuffer byteJ6;
  private final MappedByteBuffer byteJ7;
  private final MappedByteBuffer byteQ0;
  private final MappedByteBuffer byteQ1;
  private final MappedByteBuffer byteQ2;
  private final MappedByteBuffer byteQ3;
  private final MappedByteBuffer byteQ4;
  private final MappedByteBuffer byteQ5;
  private final MappedByteBuffer byteQ6;
  private final MappedByteBuffer byteQ7;
  private final MappedByteBuffer byteR0;
  private final MappedByteBuffer byteR1;
  private final MappedByteBuffer byteR2;
  private final MappedByteBuffer byteR3;
  private final MappedByteBuffer cmp;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer ofH;
  private final MappedByteBuffer ofI;
  private final MappedByteBuffer ofJ;
  private final MappedByteBuffer ofRes;
  private final MappedByteBuffer oli;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer stamp;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("ext.ACC_A_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_A_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_A_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_A_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_B_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_B_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_B_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_B_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_C_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_C_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_C_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_C_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_DELTA_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_DELTA_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_DELTA_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_DELTA_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_H_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_H_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_H_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_H_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_H_4", 8, length));
      headers.add(new ColumnHeader("ext.ACC_H_5", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_4", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_5", 8, length));
      headers.add(new ColumnHeader("ext.ACC_I_6", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_4", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_5", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_6", 8, length));
      headers.add(new ColumnHeader("ext.ACC_J_7", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_3", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_4", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_5", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_6", 8, length));
      headers.add(new ColumnHeader("ext.ACC_Q_7", 8, length));
      headers.add(new ColumnHeader("ext.ACC_R_0", 8, length));
      headers.add(new ColumnHeader("ext.ACC_R_1", 8, length));
      headers.add(new ColumnHeader("ext.ACC_R_2", 8, length));
      headers.add(new ColumnHeader("ext.ACC_R_3", 8, length));
      headers.add(new ColumnHeader("ext.ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("ext.ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("ext.ARG_2_HI", 16, length));
      headers.add(new ColumnHeader("ext.ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("ext.ARG_3_HI", 16, length));
      headers.add(new ColumnHeader("ext.ARG_3_LO", 16, length));
      headers.add(new ColumnHeader("ext.BIT_1", 1, length));
      headers.add(new ColumnHeader("ext.BIT_2", 1, length));
      headers.add(new ColumnHeader("ext.BIT_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_A_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_A_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_A_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_A_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_B_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_B_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_B_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_B_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_C_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_C_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_C_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_C_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_DELTA_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_DELTA_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_DELTA_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_DELTA_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_H_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_H_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_H_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_H_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_H_4", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_H_5", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_4", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_5", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_I_6", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_4", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_5", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_6", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_J_7", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_3", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_4", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_5", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_6", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_Q_7", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_R_0", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_R_1", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_R_2", 1, length));
      headers.add(new ColumnHeader("ext.BYTE_R_3", 1, length));
      headers.add(new ColumnHeader("ext.CMP", 1, length));
      headers.add(new ColumnHeader("ext.CT", 1, length));
      headers.add(new ColumnHeader("ext.INST", 1, length));
      headers.add(new ColumnHeader("ext.OF_H", 1, length));
      headers.add(new ColumnHeader("ext.OF_I", 1, length));
      headers.add(new ColumnHeader("ext.OF_J", 1, length));
      headers.add(new ColumnHeader("ext.OF_RES", 1, length));
      headers.add(new ColumnHeader("ext.OLI", 1, length));
      headers.add(new ColumnHeader("ext.RES_HI", 16, length));
      headers.add(new ColumnHeader("ext.RES_LO", 16, length));
      headers.add(new ColumnHeader("ext.STAMP", 4, length));
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
    this.accDelta0 = buffers.get(12);
    this.accDelta1 = buffers.get(13);
    this.accDelta2 = buffers.get(14);
    this.accDelta3 = buffers.get(15);
    this.accH0 = buffers.get(16);
    this.accH1 = buffers.get(17);
    this.accH2 = buffers.get(18);
    this.accH3 = buffers.get(19);
    this.accH4 = buffers.get(20);
    this.accH5 = buffers.get(21);
    this.accI0 = buffers.get(22);
    this.accI1 = buffers.get(23);
    this.accI2 = buffers.get(24);
    this.accI3 = buffers.get(25);
    this.accI4 = buffers.get(26);
    this.accI5 = buffers.get(27);
    this.accI6 = buffers.get(28);
    this.accJ0 = buffers.get(29);
    this.accJ1 = buffers.get(30);
    this.accJ2 = buffers.get(31);
    this.accJ3 = buffers.get(32);
    this.accJ4 = buffers.get(33);
    this.accJ5 = buffers.get(34);
    this.accJ6 = buffers.get(35);
    this.accJ7 = buffers.get(36);
    this.accQ0 = buffers.get(37);
    this.accQ1 = buffers.get(38);
    this.accQ2 = buffers.get(39);
    this.accQ3 = buffers.get(40);
    this.accQ4 = buffers.get(41);
    this.accQ5 = buffers.get(42);
    this.accQ6 = buffers.get(43);
    this.accQ7 = buffers.get(44);
    this.accR0 = buffers.get(45);
    this.accR1 = buffers.get(46);
    this.accR2 = buffers.get(47);
    this.accR3 = buffers.get(48);
    this.arg1Hi = buffers.get(49);
    this.arg1Lo = buffers.get(50);
    this.arg2Hi = buffers.get(51);
    this.arg2Lo = buffers.get(52);
    this.arg3Hi = buffers.get(53);
    this.arg3Lo = buffers.get(54);
    this.bit1 = buffers.get(55);
    this.bit2 = buffers.get(56);
    this.bit3 = buffers.get(57);
    this.byteA0 = buffers.get(58);
    this.byteA1 = buffers.get(59);
    this.byteA2 = buffers.get(60);
    this.byteA3 = buffers.get(61);
    this.byteB0 = buffers.get(62);
    this.byteB1 = buffers.get(63);
    this.byteB2 = buffers.get(64);
    this.byteB3 = buffers.get(65);
    this.byteC0 = buffers.get(66);
    this.byteC1 = buffers.get(67);
    this.byteC2 = buffers.get(68);
    this.byteC3 = buffers.get(69);
    this.byteDelta0 = buffers.get(70);
    this.byteDelta1 = buffers.get(71);
    this.byteDelta2 = buffers.get(72);
    this.byteDelta3 = buffers.get(73);
    this.byteH0 = buffers.get(74);
    this.byteH1 = buffers.get(75);
    this.byteH2 = buffers.get(76);
    this.byteH3 = buffers.get(77);
    this.byteH4 = buffers.get(78);
    this.byteH5 = buffers.get(79);
    this.byteI0 = buffers.get(80);
    this.byteI1 = buffers.get(81);
    this.byteI2 = buffers.get(82);
    this.byteI3 = buffers.get(83);
    this.byteI4 = buffers.get(84);
    this.byteI5 = buffers.get(85);
    this.byteI6 = buffers.get(86);
    this.byteJ0 = buffers.get(87);
    this.byteJ1 = buffers.get(88);
    this.byteJ2 = buffers.get(89);
    this.byteJ3 = buffers.get(90);
    this.byteJ4 = buffers.get(91);
    this.byteJ5 = buffers.get(92);
    this.byteJ6 = buffers.get(93);
    this.byteJ7 = buffers.get(94);
    this.byteQ0 = buffers.get(95);
    this.byteQ1 = buffers.get(96);
    this.byteQ2 = buffers.get(97);
    this.byteQ3 = buffers.get(98);
    this.byteQ4 = buffers.get(99);
    this.byteQ5 = buffers.get(100);
    this.byteQ6 = buffers.get(101);
    this.byteQ7 = buffers.get(102);
    this.byteR0 = buffers.get(103);
    this.byteR1 = buffers.get(104);
    this.byteR2 = buffers.get(105);
    this.byteR3 = buffers.get(106);
    this.cmp = buffers.get(107);
    this.ct = buffers.get(108);
    this.inst = buffers.get(109);
    this.ofH = buffers.get(110);
    this.ofI = buffers.get(111);
    this.ofJ = buffers.get(112);
    this.ofRes = buffers.get(113);
    this.oli = buffers.get(114);
    this.resHi = buffers.get(115);
    this.resLo = buffers.get(116);
    this.stamp = buffers.get(117);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace accA0(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("ext.ACC_A_0 already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_A_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA0.put(bs.get(j)); }

    return this;
  }

  public Trace accA1(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("ext.ACC_A_1 already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_A_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA1.put(bs.get(j)); }

    return this;
  }

  public Trace accA2(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("ext.ACC_A_2 already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_A_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA2.put(bs.get(j)); }

    return this;
  }

  public Trace accA3(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("ext.ACC_A_3 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_A_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accA3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA3.put(bs.get(j)); }

    return this;
  }

  public Trace accB0(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("ext.ACC_B_0 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_B_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB0.put(bs.get(j)); }

    return this;
  }

  public Trace accB1(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("ext.ACC_B_1 already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_B_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB1.put(bs.get(j)); }

    return this;
  }

  public Trace accB2(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("ext.ACC_B_2 already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_B_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB2.put(bs.get(j)); }

    return this;
  }

  public Trace accB3(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("ext.ACC_B_3 already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_B_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accB3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB3.put(bs.get(j)); }

    return this;
  }

  public Trace accC0(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("ext.ACC_C_0 already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_C_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC0.put(bs.get(j)); }

    return this;
  }

  public Trace accC1(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("ext.ACC_C_1 already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_C_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC1.put(bs.get(j)); }

    return this;
  }

  public Trace accC2(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("ext.ACC_C_2 already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_C_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC2.put(bs.get(j)); }

    return this;
  }

  public Trace accC3(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("ext.ACC_C_3 already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_C_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accC3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC3.put(bs.get(j)); }

    return this;
  }

  public Trace accDelta0(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("ext.ACC_DELTA_0 already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_DELTA_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accDelta0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accDelta0.put(bs.get(j)); }

    return this;
  }

  public Trace accDelta1(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("ext.ACC_DELTA_1 already set");
    } else {
      filled.set(13);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_DELTA_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accDelta1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accDelta1.put(bs.get(j)); }

    return this;
  }

  public Trace accDelta2(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("ext.ACC_DELTA_2 already set");
    } else {
      filled.set(14);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_DELTA_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accDelta2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accDelta2.put(bs.get(j)); }

    return this;
  }

  public Trace accDelta3(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("ext.ACC_DELTA_3 already set");
    } else {
      filled.set(15);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_DELTA_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accDelta3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accDelta3.put(bs.get(j)); }

    return this;
  }

  public Trace accH0(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("ext.ACC_H_0 already set");
    } else {
      filled.set(16);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_H_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH0.put(bs.get(j)); }

    return this;
  }

  public Trace accH1(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("ext.ACC_H_1 already set");
    } else {
      filled.set(17);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_H_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH1.put(bs.get(j)); }

    return this;
  }

  public Trace accH2(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("ext.ACC_H_2 already set");
    } else {
      filled.set(18);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_H_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH2.put(bs.get(j)); }

    return this;
  }

  public Trace accH3(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("ext.ACC_H_3 already set");
    } else {
      filled.set(19);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_H_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH3.put(bs.get(j)); }

    return this;
  }

  public Trace accH4(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("ext.ACC_H_4 already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_H_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH4.put(bs.get(j)); }

    return this;
  }

  public Trace accH5(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("ext.ACC_H_5 already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_H_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accH5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accH5.put(bs.get(j)); }

    return this;
  }

  public Trace accI0(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("ext.ACC_I_0 already set");
    } else {
      filled.set(22);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI0.put(bs.get(j)); }

    return this;
  }

  public Trace accI1(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("ext.ACC_I_1 already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI1.put(bs.get(j)); }

    return this;
  }

  public Trace accI2(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("ext.ACC_I_2 already set");
    } else {
      filled.set(24);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI2.put(bs.get(j)); }

    return this;
  }

  public Trace accI3(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("ext.ACC_I_3 already set");
    } else {
      filled.set(25);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI3.put(bs.get(j)); }

    return this;
  }

  public Trace accI4(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("ext.ACC_I_4 already set");
    } else {
      filled.set(26);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI4.put(bs.get(j)); }

    return this;
  }

  public Trace accI5(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("ext.ACC_I_5 already set");
    } else {
      filled.set(27);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI5.put(bs.get(j)); }

    return this;
  }

  public Trace accI6(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("ext.ACC_I_6 already set");
    } else {
      filled.set(28);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_I_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accI6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accI6.put(bs.get(j)); }

    return this;
  }

  public Trace accJ0(final Bytes b) {
    if (filled.get(29)) {
      throw new IllegalStateException("ext.ACC_J_0 already set");
    } else {
      filled.set(29);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ0.put(bs.get(j)); }

    return this;
  }

  public Trace accJ1(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("ext.ACC_J_1 already set");
    } else {
      filled.set(30);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ1.put(bs.get(j)); }

    return this;
  }

  public Trace accJ2(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("ext.ACC_J_2 already set");
    } else {
      filled.set(31);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ2.put(bs.get(j)); }

    return this;
  }

  public Trace accJ3(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("ext.ACC_J_3 already set");
    } else {
      filled.set(32);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ3.put(bs.get(j)); }

    return this;
  }

  public Trace accJ4(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("ext.ACC_J_4 already set");
    } else {
      filled.set(33);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ4.put(bs.get(j)); }

    return this;
  }

  public Trace accJ5(final Bytes b) {
    if (filled.get(34)) {
      throw new IllegalStateException("ext.ACC_J_5 already set");
    } else {
      filled.set(34);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ5.put(bs.get(j)); }

    return this;
  }

  public Trace accJ6(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("ext.ACC_J_6 already set");
    } else {
      filled.set(35);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ6.put(bs.get(j)); }

    return this;
  }

  public Trace accJ7(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("ext.ACC_J_7 already set");
    } else {
      filled.set(36);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_J_7 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accJ7.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accJ7.put(bs.get(j)); }

    return this;
  }

  public Trace accQ0(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("ext.ACC_Q_0 already set");
    } else {
      filled.set(37);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ0.put(bs.get(j)); }

    return this;
  }

  public Trace accQ1(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("ext.ACC_Q_1 already set");
    } else {
      filled.set(38);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ1.put(bs.get(j)); }

    return this;
  }

  public Trace accQ2(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("ext.ACC_Q_2 already set");
    } else {
      filled.set(39);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ2.put(bs.get(j)); }

    return this;
  }

  public Trace accQ3(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("ext.ACC_Q_3 already set");
    } else {
      filled.set(40);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ3.put(bs.get(j)); }

    return this;
  }

  public Trace accQ4(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("ext.ACC_Q_4 already set");
    } else {
      filled.set(41);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ4.put(bs.get(j)); }

    return this;
  }

  public Trace accQ5(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("ext.ACC_Q_5 already set");
    } else {
      filled.set(42);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ5.put(bs.get(j)); }

    return this;
  }

  public Trace accQ6(final Bytes b) {
    if (filled.get(43)) {
      throw new IllegalStateException("ext.ACC_Q_6 already set");
    } else {
      filled.set(43);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ6.put(bs.get(j)); }

    return this;
  }

  public Trace accQ7(final Bytes b) {
    if (filled.get(44)) {
      throw new IllegalStateException("ext.ACC_Q_7 already set");
    } else {
      filled.set(44);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_Q_7 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accQ7.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accQ7.put(bs.get(j)); }

    return this;
  }

  public Trace accR0(final Bytes b) {
    if (filled.get(45)) {
      throw new IllegalStateException("ext.ACC_R_0 already set");
    } else {
      filled.set(45);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_R_0 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accR0.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accR0.put(bs.get(j)); }

    return this;
  }

  public Trace accR1(final Bytes b) {
    if (filled.get(46)) {
      throw new IllegalStateException("ext.ACC_R_1 already set");
    } else {
      filled.set(46);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_R_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accR1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accR1.put(bs.get(j)); }

    return this;
  }

  public Trace accR2(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("ext.ACC_R_2 already set");
    } else {
      filled.set(47);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_R_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accR2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accR2.put(bs.get(j)); }

    return this;
  }

  public Trace accR3(final Bytes b) {
    if (filled.get(48)) {
      throw new IllegalStateException("ext.ACC_R_3 already set");
    } else {
      filled.set(48);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("ext.ACC_R_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { accR3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accR3.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(49)) {
      throw new IllegalStateException("ext.ARG_1_HI already set");
    } else {
      filled.set(49);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(50)) {
      throw new IllegalStateException("ext.ARG_1_LO already set");
    } else {
      filled.set(50);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(51)) {
      throw new IllegalStateException("ext.ARG_2_HI already set");
    } else {
      filled.set(51);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.ARG_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(52)) {
      throw new IllegalStateException("ext.ARG_2_LO already set");
    } else {
      filled.set(52);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace arg3Hi(final Bytes b) {
    if (filled.get(53)) {
      throw new IllegalStateException("ext.ARG_3_HI already set");
    } else {
      filled.set(53);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.ARG_3_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg3Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg3Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg3Lo(final Bytes b) {
    if (filled.get(54)) {
      throw new IllegalStateException("ext.ARG_3_LO already set");
    } else {
      filled.set(54);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.ARG_3_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg3Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg3Lo.put(bs.get(j)); }

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("ext.BIT_1 already set");
    } else {
      filled.set(55);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("ext.BIT_2 already set");
    } else {
      filled.set(56);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("ext.BIT_3 already set");
    } else {
      filled.set(57);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byteA0(final UnsignedByte b) {
    if (filled.get(58)) {
      throw new IllegalStateException("ext.BYTE_A_0 already set");
    } else {
      filled.set(58);
    }

    byteA0.put(b.toByte());

    return this;
  }

  public Trace byteA1(final UnsignedByte b) {
    if (filled.get(59)) {
      throw new IllegalStateException("ext.BYTE_A_1 already set");
    } else {
      filled.set(59);
    }

    byteA1.put(b.toByte());

    return this;
  }

  public Trace byteA2(final UnsignedByte b) {
    if (filled.get(60)) {
      throw new IllegalStateException("ext.BYTE_A_2 already set");
    } else {
      filled.set(60);
    }

    byteA2.put(b.toByte());

    return this;
  }

  public Trace byteA3(final UnsignedByte b) {
    if (filled.get(61)) {
      throw new IllegalStateException("ext.BYTE_A_3 already set");
    } else {
      filled.set(61);
    }

    byteA3.put(b.toByte());

    return this;
  }

  public Trace byteB0(final UnsignedByte b) {
    if (filled.get(62)) {
      throw new IllegalStateException("ext.BYTE_B_0 already set");
    } else {
      filled.set(62);
    }

    byteB0.put(b.toByte());

    return this;
  }

  public Trace byteB1(final UnsignedByte b) {
    if (filled.get(63)) {
      throw new IllegalStateException("ext.BYTE_B_1 already set");
    } else {
      filled.set(63);
    }

    byteB1.put(b.toByte());

    return this;
  }

  public Trace byteB2(final UnsignedByte b) {
    if (filled.get(64)) {
      throw new IllegalStateException("ext.BYTE_B_2 already set");
    } else {
      filled.set(64);
    }

    byteB2.put(b.toByte());

    return this;
  }

  public Trace byteB3(final UnsignedByte b) {
    if (filled.get(65)) {
      throw new IllegalStateException("ext.BYTE_B_3 already set");
    } else {
      filled.set(65);
    }

    byteB3.put(b.toByte());

    return this;
  }

  public Trace byteC0(final UnsignedByte b) {
    if (filled.get(66)) {
      throw new IllegalStateException("ext.BYTE_C_0 already set");
    } else {
      filled.set(66);
    }

    byteC0.put(b.toByte());

    return this;
  }

  public Trace byteC1(final UnsignedByte b) {
    if (filled.get(67)) {
      throw new IllegalStateException("ext.BYTE_C_1 already set");
    } else {
      filled.set(67);
    }

    byteC1.put(b.toByte());

    return this;
  }

  public Trace byteC2(final UnsignedByte b) {
    if (filled.get(68)) {
      throw new IllegalStateException("ext.BYTE_C_2 already set");
    } else {
      filled.set(68);
    }

    byteC2.put(b.toByte());

    return this;
  }

  public Trace byteC3(final UnsignedByte b) {
    if (filled.get(69)) {
      throw new IllegalStateException("ext.BYTE_C_3 already set");
    } else {
      filled.set(69);
    }

    byteC3.put(b.toByte());

    return this;
  }

  public Trace byteDelta0(final UnsignedByte b) {
    if (filled.get(70)) {
      throw new IllegalStateException("ext.BYTE_DELTA_0 already set");
    } else {
      filled.set(70);
    }

    byteDelta0.put(b.toByte());

    return this;
  }

  public Trace byteDelta1(final UnsignedByte b) {
    if (filled.get(71)) {
      throw new IllegalStateException("ext.BYTE_DELTA_1 already set");
    } else {
      filled.set(71);
    }

    byteDelta1.put(b.toByte());

    return this;
  }

  public Trace byteDelta2(final UnsignedByte b) {
    if (filled.get(72)) {
      throw new IllegalStateException("ext.BYTE_DELTA_2 already set");
    } else {
      filled.set(72);
    }

    byteDelta2.put(b.toByte());

    return this;
  }

  public Trace byteDelta3(final UnsignedByte b) {
    if (filled.get(73)) {
      throw new IllegalStateException("ext.BYTE_DELTA_3 already set");
    } else {
      filled.set(73);
    }

    byteDelta3.put(b.toByte());

    return this;
  }

  public Trace byteH0(final UnsignedByte b) {
    if (filled.get(74)) {
      throw new IllegalStateException("ext.BYTE_H_0 already set");
    } else {
      filled.set(74);
    }

    byteH0.put(b.toByte());

    return this;
  }

  public Trace byteH1(final UnsignedByte b) {
    if (filled.get(75)) {
      throw new IllegalStateException("ext.BYTE_H_1 already set");
    } else {
      filled.set(75);
    }

    byteH1.put(b.toByte());

    return this;
  }

  public Trace byteH2(final UnsignedByte b) {
    if (filled.get(76)) {
      throw new IllegalStateException("ext.BYTE_H_2 already set");
    } else {
      filled.set(76);
    }

    byteH2.put(b.toByte());

    return this;
  }

  public Trace byteH3(final UnsignedByte b) {
    if (filled.get(77)) {
      throw new IllegalStateException("ext.BYTE_H_3 already set");
    } else {
      filled.set(77);
    }

    byteH3.put(b.toByte());

    return this;
  }

  public Trace byteH4(final UnsignedByte b) {
    if (filled.get(78)) {
      throw new IllegalStateException("ext.BYTE_H_4 already set");
    } else {
      filled.set(78);
    }

    byteH4.put(b.toByte());

    return this;
  }

  public Trace byteH5(final UnsignedByte b) {
    if (filled.get(79)) {
      throw new IllegalStateException("ext.BYTE_H_5 already set");
    } else {
      filled.set(79);
    }

    byteH5.put(b.toByte());

    return this;
  }

  public Trace byteI0(final UnsignedByte b) {
    if (filled.get(80)) {
      throw new IllegalStateException("ext.BYTE_I_0 already set");
    } else {
      filled.set(80);
    }

    byteI0.put(b.toByte());

    return this;
  }

  public Trace byteI1(final UnsignedByte b) {
    if (filled.get(81)) {
      throw new IllegalStateException("ext.BYTE_I_1 already set");
    } else {
      filled.set(81);
    }

    byteI1.put(b.toByte());

    return this;
  }

  public Trace byteI2(final UnsignedByte b) {
    if (filled.get(82)) {
      throw new IllegalStateException("ext.BYTE_I_2 already set");
    } else {
      filled.set(82);
    }

    byteI2.put(b.toByte());

    return this;
  }

  public Trace byteI3(final UnsignedByte b) {
    if (filled.get(83)) {
      throw new IllegalStateException("ext.BYTE_I_3 already set");
    } else {
      filled.set(83);
    }

    byteI3.put(b.toByte());

    return this;
  }

  public Trace byteI4(final UnsignedByte b) {
    if (filled.get(84)) {
      throw new IllegalStateException("ext.BYTE_I_4 already set");
    } else {
      filled.set(84);
    }

    byteI4.put(b.toByte());

    return this;
  }

  public Trace byteI5(final UnsignedByte b) {
    if (filled.get(85)) {
      throw new IllegalStateException("ext.BYTE_I_5 already set");
    } else {
      filled.set(85);
    }

    byteI5.put(b.toByte());

    return this;
  }

  public Trace byteI6(final UnsignedByte b) {
    if (filled.get(86)) {
      throw new IllegalStateException("ext.BYTE_I_6 already set");
    } else {
      filled.set(86);
    }

    byteI6.put(b.toByte());

    return this;
  }

  public Trace byteJ0(final UnsignedByte b) {
    if (filled.get(87)) {
      throw new IllegalStateException("ext.BYTE_J_0 already set");
    } else {
      filled.set(87);
    }

    byteJ0.put(b.toByte());

    return this;
  }

  public Trace byteJ1(final UnsignedByte b) {
    if (filled.get(88)) {
      throw new IllegalStateException("ext.BYTE_J_1 already set");
    } else {
      filled.set(88);
    }

    byteJ1.put(b.toByte());

    return this;
  }

  public Trace byteJ2(final UnsignedByte b) {
    if (filled.get(89)) {
      throw new IllegalStateException("ext.BYTE_J_2 already set");
    } else {
      filled.set(89);
    }

    byteJ2.put(b.toByte());

    return this;
  }

  public Trace byteJ3(final UnsignedByte b) {
    if (filled.get(90)) {
      throw new IllegalStateException("ext.BYTE_J_3 already set");
    } else {
      filled.set(90);
    }

    byteJ3.put(b.toByte());

    return this;
  }

  public Trace byteJ4(final UnsignedByte b) {
    if (filled.get(91)) {
      throw new IllegalStateException("ext.BYTE_J_4 already set");
    } else {
      filled.set(91);
    }

    byteJ4.put(b.toByte());

    return this;
  }

  public Trace byteJ5(final UnsignedByte b) {
    if (filled.get(92)) {
      throw new IllegalStateException("ext.BYTE_J_5 already set");
    } else {
      filled.set(92);
    }

    byteJ5.put(b.toByte());

    return this;
  }

  public Trace byteJ6(final UnsignedByte b) {
    if (filled.get(93)) {
      throw new IllegalStateException("ext.BYTE_J_6 already set");
    } else {
      filled.set(93);
    }

    byteJ6.put(b.toByte());

    return this;
  }

  public Trace byteJ7(final UnsignedByte b) {
    if (filled.get(94)) {
      throw new IllegalStateException("ext.BYTE_J_7 already set");
    } else {
      filled.set(94);
    }

    byteJ7.put(b.toByte());

    return this;
  }

  public Trace byteQ0(final UnsignedByte b) {
    if (filled.get(95)) {
      throw new IllegalStateException("ext.BYTE_Q_0 already set");
    } else {
      filled.set(95);
    }

    byteQ0.put(b.toByte());

    return this;
  }

  public Trace byteQ1(final UnsignedByte b) {
    if (filled.get(96)) {
      throw new IllegalStateException("ext.BYTE_Q_1 already set");
    } else {
      filled.set(96);
    }

    byteQ1.put(b.toByte());

    return this;
  }

  public Trace byteQ2(final UnsignedByte b) {
    if (filled.get(97)) {
      throw new IllegalStateException("ext.BYTE_Q_2 already set");
    } else {
      filled.set(97);
    }

    byteQ2.put(b.toByte());

    return this;
  }

  public Trace byteQ3(final UnsignedByte b) {
    if (filled.get(98)) {
      throw new IllegalStateException("ext.BYTE_Q_3 already set");
    } else {
      filled.set(98);
    }

    byteQ3.put(b.toByte());

    return this;
  }

  public Trace byteQ4(final UnsignedByte b) {
    if (filled.get(99)) {
      throw new IllegalStateException("ext.BYTE_Q_4 already set");
    } else {
      filled.set(99);
    }

    byteQ4.put(b.toByte());

    return this;
  }

  public Trace byteQ5(final UnsignedByte b) {
    if (filled.get(100)) {
      throw new IllegalStateException("ext.BYTE_Q_5 already set");
    } else {
      filled.set(100);
    }

    byteQ5.put(b.toByte());

    return this;
  }

  public Trace byteQ6(final UnsignedByte b) {
    if (filled.get(101)) {
      throw new IllegalStateException("ext.BYTE_Q_6 already set");
    } else {
      filled.set(101);
    }

    byteQ6.put(b.toByte());

    return this;
  }

  public Trace byteQ7(final UnsignedByte b) {
    if (filled.get(102)) {
      throw new IllegalStateException("ext.BYTE_Q_7 already set");
    } else {
      filled.set(102);
    }

    byteQ7.put(b.toByte());

    return this;
  }

  public Trace byteR0(final UnsignedByte b) {
    if (filled.get(103)) {
      throw new IllegalStateException("ext.BYTE_R_0 already set");
    } else {
      filled.set(103);
    }

    byteR0.put(b.toByte());

    return this;
  }

  public Trace byteR1(final UnsignedByte b) {
    if (filled.get(104)) {
      throw new IllegalStateException("ext.BYTE_R_1 already set");
    } else {
      filled.set(104);
    }

    byteR1.put(b.toByte());

    return this;
  }

  public Trace byteR2(final UnsignedByte b) {
    if (filled.get(105)) {
      throw new IllegalStateException("ext.BYTE_R_2 already set");
    } else {
      filled.set(105);
    }

    byteR2.put(b.toByte());

    return this;
  }

  public Trace byteR3(final UnsignedByte b) {
    if (filled.get(106)) {
      throw new IllegalStateException("ext.BYTE_R_3 already set");
    } else {
      filled.set(106);
    }

    byteR3.put(b.toByte());

    return this;
  }

  public Trace cmp(final Boolean b) {
    if (filled.get(107)) {
      throw new IllegalStateException("ext.CMP already set");
    } else {
      filled.set(107);
    }

    cmp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("ext.CT already set");
    } else {
      filled.set(108);
    }

    if(b >= 8L) { throw new IllegalArgumentException("ext.CT has invalid value (" + b + ")"); }
    ct.put((byte) b);


    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(109)) {
      throw new IllegalStateException("ext.INST already set");
    } else {
      filled.set(109);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace ofH(final Boolean b) {
    if (filled.get(110)) {
      throw new IllegalStateException("ext.OF_H already set");
    } else {
      filled.set(110);
    }

    ofH.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ofI(final Boolean b) {
    if (filled.get(111)) {
      throw new IllegalStateException("ext.OF_I already set");
    } else {
      filled.set(111);
    }

    ofI.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ofJ(final Boolean b) {
    if (filled.get(112)) {
      throw new IllegalStateException("ext.OF_J already set");
    } else {
      filled.set(112);
    }

    ofJ.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ofRes(final Boolean b) {
    if (filled.get(113)) {
      throw new IllegalStateException("ext.OF_RES already set");
    } else {
      filled.set(113);
    }

    ofRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oli(final Boolean b) {
    if (filled.get(114)) {
      throw new IllegalStateException("ext.OLI already set");
    } else {
      filled.set(114);
    }

    oli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(115)) {
      throw new IllegalStateException("ext.RES_HI already set");
    } else {
      filled.set(115);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.RES_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resHi.put(bs.get(j)); }

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(116)) {
      throw new IllegalStateException("ext.RES_LO already set");
    } else {
      filled.set(116);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("ext.RES_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resLo.put(bs.get(j)); }

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(117)) {
      throw new IllegalStateException("ext.STAMP already set");
    } else {
      filled.set(117);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("ext.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 24));
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("ext.ACC_A_0 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("ext.ACC_A_1 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("ext.ACC_A_2 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("ext.ACC_A_3 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("ext.ACC_B_0 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("ext.ACC_B_1 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("ext.ACC_B_2 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("ext.ACC_B_3 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("ext.ACC_C_0 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("ext.ACC_C_1 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("ext.ACC_C_2 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("ext.ACC_C_3 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("ext.ACC_DELTA_0 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("ext.ACC_DELTA_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("ext.ACC_DELTA_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("ext.ACC_DELTA_3 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("ext.ACC_H_0 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("ext.ACC_H_1 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("ext.ACC_H_2 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("ext.ACC_H_3 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("ext.ACC_H_4 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("ext.ACC_H_5 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("ext.ACC_I_0 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("ext.ACC_I_1 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("ext.ACC_I_2 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("ext.ACC_I_3 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("ext.ACC_I_4 has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("ext.ACC_I_5 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("ext.ACC_I_6 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("ext.ACC_J_0 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("ext.ACC_J_1 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("ext.ACC_J_2 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("ext.ACC_J_3 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("ext.ACC_J_4 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("ext.ACC_J_5 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("ext.ACC_J_6 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("ext.ACC_J_7 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("ext.ACC_Q_0 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("ext.ACC_Q_1 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("ext.ACC_Q_2 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("ext.ACC_Q_3 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("ext.ACC_Q_4 has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("ext.ACC_Q_5 has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("ext.ACC_Q_6 has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("ext.ACC_Q_7 has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("ext.ACC_R_0 has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("ext.ACC_R_1 has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("ext.ACC_R_2 has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("ext.ACC_R_3 has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("ext.ARG_1_HI has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("ext.ARG_1_LO has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("ext.ARG_2_HI has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("ext.ARG_2_LO has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("ext.ARG_3_HI has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("ext.ARG_3_LO has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("ext.BIT_1 has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("ext.BIT_2 has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("ext.BIT_3 has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("ext.BYTE_A_0 has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("ext.BYTE_A_1 has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("ext.BYTE_A_2 has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("ext.BYTE_A_3 has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("ext.BYTE_B_0 has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("ext.BYTE_B_1 has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("ext.BYTE_B_2 has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("ext.BYTE_B_3 has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException("ext.BYTE_C_0 has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException("ext.BYTE_C_1 has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException("ext.BYTE_C_2 has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException("ext.BYTE_C_3 has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException("ext.BYTE_DELTA_0 has not been filled");
    }

    if (!filled.get(71)) {
      throw new IllegalStateException("ext.BYTE_DELTA_1 has not been filled");
    }

    if (!filled.get(72)) {
      throw new IllegalStateException("ext.BYTE_DELTA_2 has not been filled");
    }

    if (!filled.get(73)) {
      throw new IllegalStateException("ext.BYTE_DELTA_3 has not been filled");
    }

    if (!filled.get(74)) {
      throw new IllegalStateException("ext.BYTE_H_0 has not been filled");
    }

    if (!filled.get(75)) {
      throw new IllegalStateException("ext.BYTE_H_1 has not been filled");
    }

    if (!filled.get(76)) {
      throw new IllegalStateException("ext.BYTE_H_2 has not been filled");
    }

    if (!filled.get(77)) {
      throw new IllegalStateException("ext.BYTE_H_3 has not been filled");
    }

    if (!filled.get(78)) {
      throw new IllegalStateException("ext.BYTE_H_4 has not been filled");
    }

    if (!filled.get(79)) {
      throw new IllegalStateException("ext.BYTE_H_5 has not been filled");
    }

    if (!filled.get(80)) {
      throw new IllegalStateException("ext.BYTE_I_0 has not been filled");
    }

    if (!filled.get(81)) {
      throw new IllegalStateException("ext.BYTE_I_1 has not been filled");
    }

    if (!filled.get(82)) {
      throw new IllegalStateException("ext.BYTE_I_2 has not been filled");
    }

    if (!filled.get(83)) {
      throw new IllegalStateException("ext.BYTE_I_3 has not been filled");
    }

    if (!filled.get(84)) {
      throw new IllegalStateException("ext.BYTE_I_4 has not been filled");
    }

    if (!filled.get(85)) {
      throw new IllegalStateException("ext.BYTE_I_5 has not been filled");
    }

    if (!filled.get(86)) {
      throw new IllegalStateException("ext.BYTE_I_6 has not been filled");
    }

    if (!filled.get(87)) {
      throw new IllegalStateException("ext.BYTE_J_0 has not been filled");
    }

    if (!filled.get(88)) {
      throw new IllegalStateException("ext.BYTE_J_1 has not been filled");
    }

    if (!filled.get(89)) {
      throw new IllegalStateException("ext.BYTE_J_2 has not been filled");
    }

    if (!filled.get(90)) {
      throw new IllegalStateException("ext.BYTE_J_3 has not been filled");
    }

    if (!filled.get(91)) {
      throw new IllegalStateException("ext.BYTE_J_4 has not been filled");
    }

    if (!filled.get(92)) {
      throw new IllegalStateException("ext.BYTE_J_5 has not been filled");
    }

    if (!filled.get(93)) {
      throw new IllegalStateException("ext.BYTE_J_6 has not been filled");
    }

    if (!filled.get(94)) {
      throw new IllegalStateException("ext.BYTE_J_7 has not been filled");
    }

    if (!filled.get(95)) {
      throw new IllegalStateException("ext.BYTE_Q_0 has not been filled");
    }

    if (!filled.get(96)) {
      throw new IllegalStateException("ext.BYTE_Q_1 has not been filled");
    }

    if (!filled.get(97)) {
      throw new IllegalStateException("ext.BYTE_Q_2 has not been filled");
    }

    if (!filled.get(98)) {
      throw new IllegalStateException("ext.BYTE_Q_3 has not been filled");
    }

    if (!filled.get(99)) {
      throw new IllegalStateException("ext.BYTE_Q_4 has not been filled");
    }

    if (!filled.get(100)) {
      throw new IllegalStateException("ext.BYTE_Q_5 has not been filled");
    }

    if (!filled.get(101)) {
      throw new IllegalStateException("ext.BYTE_Q_6 has not been filled");
    }

    if (!filled.get(102)) {
      throw new IllegalStateException("ext.BYTE_Q_7 has not been filled");
    }

    if (!filled.get(103)) {
      throw new IllegalStateException("ext.BYTE_R_0 has not been filled");
    }

    if (!filled.get(104)) {
      throw new IllegalStateException("ext.BYTE_R_1 has not been filled");
    }

    if (!filled.get(105)) {
      throw new IllegalStateException("ext.BYTE_R_2 has not been filled");
    }

    if (!filled.get(106)) {
      throw new IllegalStateException("ext.BYTE_R_3 has not been filled");
    }

    if (!filled.get(107)) {
      throw new IllegalStateException("ext.CMP has not been filled");
    }

    if (!filled.get(108)) {
      throw new IllegalStateException("ext.CT has not been filled");
    }

    if (!filled.get(109)) {
      throw new IllegalStateException("ext.INST has not been filled");
    }

    if (!filled.get(110)) {
      throw new IllegalStateException("ext.OF_H has not been filled");
    }

    if (!filled.get(111)) {
      throw new IllegalStateException("ext.OF_I has not been filled");
    }

    if (!filled.get(112)) {
      throw new IllegalStateException("ext.OF_J has not been filled");
    }

    if (!filled.get(113)) {
      throw new IllegalStateException("ext.OF_RES has not been filled");
    }

    if (!filled.get(114)) {
      throw new IllegalStateException("ext.OLI has not been filled");
    }

    if (!filled.get(115)) {
      throw new IllegalStateException("ext.RES_HI has not been filled");
    }

    if (!filled.get(116)) {
      throw new IllegalStateException("ext.RES_LO has not been filled");
    }

    if (!filled.get(117)) {
      throw new IllegalStateException("ext.STAMP has not been filled");
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
      accDelta0.position(accDelta0.position() + 8);
    }

    if (!filled.get(13)) {
      accDelta1.position(accDelta1.position() + 8);
    }

    if (!filled.get(14)) {
      accDelta2.position(accDelta2.position() + 8);
    }

    if (!filled.get(15)) {
      accDelta3.position(accDelta3.position() + 8);
    }

    if (!filled.get(16)) {
      accH0.position(accH0.position() + 8);
    }

    if (!filled.get(17)) {
      accH1.position(accH1.position() + 8);
    }

    if (!filled.get(18)) {
      accH2.position(accH2.position() + 8);
    }

    if (!filled.get(19)) {
      accH3.position(accH3.position() + 8);
    }

    if (!filled.get(20)) {
      accH4.position(accH4.position() + 8);
    }

    if (!filled.get(21)) {
      accH5.position(accH5.position() + 8);
    }

    if (!filled.get(22)) {
      accI0.position(accI0.position() + 8);
    }

    if (!filled.get(23)) {
      accI1.position(accI1.position() + 8);
    }

    if (!filled.get(24)) {
      accI2.position(accI2.position() + 8);
    }

    if (!filled.get(25)) {
      accI3.position(accI3.position() + 8);
    }

    if (!filled.get(26)) {
      accI4.position(accI4.position() + 8);
    }

    if (!filled.get(27)) {
      accI5.position(accI5.position() + 8);
    }

    if (!filled.get(28)) {
      accI6.position(accI6.position() + 8);
    }

    if (!filled.get(29)) {
      accJ0.position(accJ0.position() + 8);
    }

    if (!filled.get(30)) {
      accJ1.position(accJ1.position() + 8);
    }

    if (!filled.get(31)) {
      accJ2.position(accJ2.position() + 8);
    }

    if (!filled.get(32)) {
      accJ3.position(accJ3.position() + 8);
    }

    if (!filled.get(33)) {
      accJ4.position(accJ4.position() + 8);
    }

    if (!filled.get(34)) {
      accJ5.position(accJ5.position() + 8);
    }

    if (!filled.get(35)) {
      accJ6.position(accJ6.position() + 8);
    }

    if (!filled.get(36)) {
      accJ7.position(accJ7.position() + 8);
    }

    if (!filled.get(37)) {
      accQ0.position(accQ0.position() + 8);
    }

    if (!filled.get(38)) {
      accQ1.position(accQ1.position() + 8);
    }

    if (!filled.get(39)) {
      accQ2.position(accQ2.position() + 8);
    }

    if (!filled.get(40)) {
      accQ3.position(accQ3.position() + 8);
    }

    if (!filled.get(41)) {
      accQ4.position(accQ4.position() + 8);
    }

    if (!filled.get(42)) {
      accQ5.position(accQ5.position() + 8);
    }

    if (!filled.get(43)) {
      accQ6.position(accQ6.position() + 8);
    }

    if (!filled.get(44)) {
      accQ7.position(accQ7.position() + 8);
    }

    if (!filled.get(45)) {
      accR0.position(accR0.position() + 8);
    }

    if (!filled.get(46)) {
      accR1.position(accR1.position() + 8);
    }

    if (!filled.get(47)) {
      accR2.position(accR2.position() + 8);
    }

    if (!filled.get(48)) {
      accR3.position(accR3.position() + 8);
    }

    if (!filled.get(49)) {
      arg1Hi.position(arg1Hi.position() + 16);
    }

    if (!filled.get(50)) {
      arg1Lo.position(arg1Lo.position() + 16);
    }

    if (!filled.get(51)) {
      arg2Hi.position(arg2Hi.position() + 16);
    }

    if (!filled.get(52)) {
      arg2Lo.position(arg2Lo.position() + 16);
    }

    if (!filled.get(53)) {
      arg3Hi.position(arg3Hi.position() + 16);
    }

    if (!filled.get(54)) {
      arg3Lo.position(arg3Lo.position() + 16);
    }

    if (!filled.get(55)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(56)) {
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(57)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(58)) {
      byteA0.position(byteA0.position() + 1);
    }

    if (!filled.get(59)) {
      byteA1.position(byteA1.position() + 1);
    }

    if (!filled.get(60)) {
      byteA2.position(byteA2.position() + 1);
    }

    if (!filled.get(61)) {
      byteA3.position(byteA3.position() + 1);
    }

    if (!filled.get(62)) {
      byteB0.position(byteB0.position() + 1);
    }

    if (!filled.get(63)) {
      byteB1.position(byteB1.position() + 1);
    }

    if (!filled.get(64)) {
      byteB2.position(byteB2.position() + 1);
    }

    if (!filled.get(65)) {
      byteB3.position(byteB3.position() + 1);
    }

    if (!filled.get(66)) {
      byteC0.position(byteC0.position() + 1);
    }

    if (!filled.get(67)) {
      byteC1.position(byteC1.position() + 1);
    }

    if (!filled.get(68)) {
      byteC2.position(byteC2.position() + 1);
    }

    if (!filled.get(69)) {
      byteC3.position(byteC3.position() + 1);
    }

    if (!filled.get(70)) {
      byteDelta0.position(byteDelta0.position() + 1);
    }

    if (!filled.get(71)) {
      byteDelta1.position(byteDelta1.position() + 1);
    }

    if (!filled.get(72)) {
      byteDelta2.position(byteDelta2.position() + 1);
    }

    if (!filled.get(73)) {
      byteDelta3.position(byteDelta3.position() + 1);
    }

    if (!filled.get(74)) {
      byteH0.position(byteH0.position() + 1);
    }

    if (!filled.get(75)) {
      byteH1.position(byteH1.position() + 1);
    }

    if (!filled.get(76)) {
      byteH2.position(byteH2.position() + 1);
    }

    if (!filled.get(77)) {
      byteH3.position(byteH3.position() + 1);
    }

    if (!filled.get(78)) {
      byteH4.position(byteH4.position() + 1);
    }

    if (!filled.get(79)) {
      byteH5.position(byteH5.position() + 1);
    }

    if (!filled.get(80)) {
      byteI0.position(byteI0.position() + 1);
    }

    if (!filled.get(81)) {
      byteI1.position(byteI1.position() + 1);
    }

    if (!filled.get(82)) {
      byteI2.position(byteI2.position() + 1);
    }

    if (!filled.get(83)) {
      byteI3.position(byteI3.position() + 1);
    }

    if (!filled.get(84)) {
      byteI4.position(byteI4.position() + 1);
    }

    if (!filled.get(85)) {
      byteI5.position(byteI5.position() + 1);
    }

    if (!filled.get(86)) {
      byteI6.position(byteI6.position() + 1);
    }

    if (!filled.get(87)) {
      byteJ0.position(byteJ0.position() + 1);
    }

    if (!filled.get(88)) {
      byteJ1.position(byteJ1.position() + 1);
    }

    if (!filled.get(89)) {
      byteJ2.position(byteJ2.position() + 1);
    }

    if (!filled.get(90)) {
      byteJ3.position(byteJ3.position() + 1);
    }

    if (!filled.get(91)) {
      byteJ4.position(byteJ4.position() + 1);
    }

    if (!filled.get(92)) {
      byteJ5.position(byteJ5.position() + 1);
    }

    if (!filled.get(93)) {
      byteJ6.position(byteJ6.position() + 1);
    }

    if (!filled.get(94)) {
      byteJ7.position(byteJ7.position() + 1);
    }

    if (!filled.get(95)) {
      byteQ0.position(byteQ0.position() + 1);
    }

    if (!filled.get(96)) {
      byteQ1.position(byteQ1.position() + 1);
    }

    if (!filled.get(97)) {
      byteQ2.position(byteQ2.position() + 1);
    }

    if (!filled.get(98)) {
      byteQ3.position(byteQ3.position() + 1);
    }

    if (!filled.get(99)) {
      byteQ4.position(byteQ4.position() + 1);
    }

    if (!filled.get(100)) {
      byteQ5.position(byteQ5.position() + 1);
    }

    if (!filled.get(101)) {
      byteQ6.position(byteQ6.position() + 1);
    }

    if (!filled.get(102)) {
      byteQ7.position(byteQ7.position() + 1);
    }

    if (!filled.get(103)) {
      byteR0.position(byteR0.position() + 1);
    }

    if (!filled.get(104)) {
      byteR1.position(byteR1.position() + 1);
    }

    if (!filled.get(105)) {
      byteR2.position(byteR2.position() + 1);
    }

    if (!filled.get(106)) {
      byteR3.position(byteR3.position() + 1);
    }

    if (!filled.get(107)) {
      cmp.position(cmp.position() + 1);
    }

    if (!filled.get(108)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(109)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(110)) {
      ofH.position(ofH.position() + 1);
    }

    if (!filled.get(111)) {
      ofI.position(ofI.position() + 1);
    }

    if (!filled.get(112)) {
      ofJ.position(ofJ.position() + 1);
    }

    if (!filled.get(113)) {
      ofRes.position(ofRes.position() + 1);
    }

    if (!filled.get(114)) {
      oli.position(oli.position() + 1);
    }

    if (!filled.get(115)) {
      resHi.position(resHi.position() + 16);
    }

    if (!filled.get(116)) {
      resLo.position(resLo.position() + 16);
    }

    if (!filled.get(117)) {
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
