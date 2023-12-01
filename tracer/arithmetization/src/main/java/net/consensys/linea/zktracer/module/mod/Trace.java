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

package net.consensys.linea.zktracer.module.mod;

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

  private final MappedByteBuffer acc12;
  private final MappedByteBuffer acc13;
  private final MappedByteBuffer acc22;
  private final MappedByteBuffer acc23;
  private final MappedByteBuffer accB0;
  private final MappedByteBuffer accB1;
  private final MappedByteBuffer accB2;
  private final MappedByteBuffer accB3;
  private final MappedByteBuffer accDelta0;
  private final MappedByteBuffer accDelta1;
  private final MappedByteBuffer accDelta2;
  private final MappedByteBuffer accDelta3;
  private final MappedByteBuffer accH0;
  private final MappedByteBuffer accH1;
  private final MappedByteBuffer accH2;
  private final MappedByteBuffer accQ0;
  private final MappedByteBuffer accQ1;
  private final MappedByteBuffer accQ2;
  private final MappedByteBuffer accQ3;
  private final MappedByteBuffer accR0;
  private final MappedByteBuffer accR1;
  private final MappedByteBuffer accR2;
  private final MappedByteBuffer accR3;
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer byte12;
  private final MappedByteBuffer byte13;
  private final MappedByteBuffer byte22;
  private final MappedByteBuffer byte23;
  private final MappedByteBuffer byteB0;
  private final MappedByteBuffer byteB1;
  private final MappedByteBuffer byteB2;
  private final MappedByteBuffer byteB3;
  private final MappedByteBuffer byteDelta0;
  private final MappedByteBuffer byteDelta1;
  private final MappedByteBuffer byteDelta2;
  private final MappedByteBuffer byteDelta3;
  private final MappedByteBuffer byteH0;
  private final MappedByteBuffer byteH1;
  private final MappedByteBuffer byteH2;
  private final MappedByteBuffer byteQ0;
  private final MappedByteBuffer byteQ1;
  private final MappedByteBuffer byteQ2;
  private final MappedByteBuffer byteQ3;
  private final MappedByteBuffer byteR0;
  private final MappedByteBuffer byteR1;
  private final MappedByteBuffer byteR2;
  private final MappedByteBuffer byteR3;
  private final MappedByteBuffer cmp1;
  private final MappedByteBuffer cmp2;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer decOutput;
  private final MappedByteBuffer decSigned;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer msb1;
  private final MappedByteBuffer msb2;
  private final MappedByteBuffer oli;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer stamp;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("mod.ACC_1_2", 32, length),
        new ColumnHeader("mod.ACC_1_3", 32, length),
        new ColumnHeader("mod.ACC_2_2", 32, length),
        new ColumnHeader("mod.ACC_2_3", 32, length),
        new ColumnHeader("mod.ACC_B_0", 32, length),
        new ColumnHeader("mod.ACC_B_1", 32, length),
        new ColumnHeader("mod.ACC_B_2", 32, length),
        new ColumnHeader("mod.ACC_B_3", 32, length),
        new ColumnHeader("mod.ACC_DELTA_0", 32, length),
        new ColumnHeader("mod.ACC_DELTA_1", 32, length),
        new ColumnHeader("mod.ACC_DELTA_2", 32, length),
        new ColumnHeader("mod.ACC_DELTA_3", 32, length),
        new ColumnHeader("mod.ACC_H_0", 32, length),
        new ColumnHeader("mod.ACC_H_1", 32, length),
        new ColumnHeader("mod.ACC_H_2", 32, length),
        new ColumnHeader("mod.ACC_Q_0", 32, length),
        new ColumnHeader("mod.ACC_Q_1", 32, length),
        new ColumnHeader("mod.ACC_Q_2", 32, length),
        new ColumnHeader("mod.ACC_Q_3", 32, length),
        new ColumnHeader("mod.ACC_R_0", 32, length),
        new ColumnHeader("mod.ACC_R_1", 32, length),
        new ColumnHeader("mod.ACC_R_2", 32, length),
        new ColumnHeader("mod.ACC_R_3", 32, length),
        new ColumnHeader("mod.ARG_1_HI", 32, length),
        new ColumnHeader("mod.ARG_1_LO", 32, length),
        new ColumnHeader("mod.ARG_2_HI", 32, length),
        new ColumnHeader("mod.ARG_2_LO", 32, length),
        new ColumnHeader("mod.BYTE_1_2", 1, length),
        new ColumnHeader("mod.BYTE_1_3", 1, length),
        new ColumnHeader("mod.BYTE_2_2", 1, length),
        new ColumnHeader("mod.BYTE_2_3", 1, length),
        new ColumnHeader("mod.BYTE_B_0", 1, length),
        new ColumnHeader("mod.BYTE_B_1", 1, length),
        new ColumnHeader("mod.BYTE_B_2", 1, length),
        new ColumnHeader("mod.BYTE_B_3", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_0", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_1", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_2", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_3", 1, length),
        new ColumnHeader("mod.BYTE_H_0", 1, length),
        new ColumnHeader("mod.BYTE_H_1", 1, length),
        new ColumnHeader("mod.BYTE_H_2", 1, length),
        new ColumnHeader("mod.BYTE_Q_0", 1, length),
        new ColumnHeader("mod.BYTE_Q_1", 1, length),
        new ColumnHeader("mod.BYTE_Q_2", 1, length),
        new ColumnHeader("mod.BYTE_Q_3", 1, length),
        new ColumnHeader("mod.BYTE_R_0", 1, length),
        new ColumnHeader("mod.BYTE_R_1", 1, length),
        new ColumnHeader("mod.BYTE_R_2", 1, length),
        new ColumnHeader("mod.BYTE_R_3", 1, length),
        new ColumnHeader("mod.CMP_1", 1, length),
        new ColumnHeader("mod.CMP_2", 1, length),
        new ColumnHeader("mod.CT", 32, length),
        new ColumnHeader("mod.DEC_OUTPUT", 1, length),
        new ColumnHeader("mod.DEC_SIGNED", 1, length),
        new ColumnHeader("mod.INST", 32, length),
        new ColumnHeader("mod.MSB_1", 1, length),
        new ColumnHeader("mod.MSB_2", 1, length),
        new ColumnHeader("mod.OLI", 1, length),
        new ColumnHeader("mod.RES_HI", 32, length),
        new ColumnHeader("mod.RES_LO", 32, length),
        new ColumnHeader("mod.STAMP", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc12 = buffers.get(0);
    this.acc13 = buffers.get(1);
    this.acc22 = buffers.get(2);
    this.acc23 = buffers.get(3);
    this.accB0 = buffers.get(4);
    this.accB1 = buffers.get(5);
    this.accB2 = buffers.get(6);
    this.accB3 = buffers.get(7);
    this.accDelta0 = buffers.get(8);
    this.accDelta1 = buffers.get(9);
    this.accDelta2 = buffers.get(10);
    this.accDelta3 = buffers.get(11);
    this.accH0 = buffers.get(12);
    this.accH1 = buffers.get(13);
    this.accH2 = buffers.get(14);
    this.accQ0 = buffers.get(15);
    this.accQ1 = buffers.get(16);
    this.accQ2 = buffers.get(17);
    this.accQ3 = buffers.get(18);
    this.accR0 = buffers.get(19);
    this.accR1 = buffers.get(20);
    this.accR2 = buffers.get(21);
    this.accR3 = buffers.get(22);
    this.arg1Hi = buffers.get(23);
    this.arg1Lo = buffers.get(24);
    this.arg2Hi = buffers.get(25);
    this.arg2Lo = buffers.get(26);
    this.byte12 = buffers.get(27);
    this.byte13 = buffers.get(28);
    this.byte22 = buffers.get(29);
    this.byte23 = buffers.get(30);
    this.byteB0 = buffers.get(31);
    this.byteB1 = buffers.get(32);
    this.byteB2 = buffers.get(33);
    this.byteB3 = buffers.get(34);
    this.byteDelta0 = buffers.get(35);
    this.byteDelta1 = buffers.get(36);
    this.byteDelta2 = buffers.get(37);
    this.byteDelta3 = buffers.get(38);
    this.byteH0 = buffers.get(39);
    this.byteH1 = buffers.get(40);
    this.byteH2 = buffers.get(41);
    this.byteQ0 = buffers.get(42);
    this.byteQ1 = buffers.get(43);
    this.byteQ2 = buffers.get(44);
    this.byteQ3 = buffers.get(45);
    this.byteR0 = buffers.get(46);
    this.byteR1 = buffers.get(47);
    this.byteR2 = buffers.get(48);
    this.byteR3 = buffers.get(49);
    this.cmp1 = buffers.get(50);
    this.cmp2 = buffers.get(51);
    this.ct = buffers.get(52);
    this.decOutput = buffers.get(53);
    this.decSigned = buffers.get(54);
    this.inst = buffers.get(55);
    this.msb1 = buffers.get(56);
    this.msb2 = buffers.get(57);
    this.oli = buffers.get(58);
    this.resHi = buffers.get(59);
    this.resLo = buffers.get(60);
    this.stamp = buffers.get(61);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc12(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mod.ACC_1_2 already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc12.put((byte) 0);
    }
    acc12.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc13(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mod.ACC_1_3 already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc13.put((byte) 0);
    }
    acc13.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc22(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mod.ACC_2_2 already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc22.put((byte) 0);
    }
    acc22.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc23(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mod.ACC_2_3 already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc23.put((byte) 0);
    }
    acc23.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB0(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mod.ACC_B_0 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB0.put((byte) 0);
    }
    accB0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB1(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mod.ACC_B_1 already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB1.put((byte) 0);
    }
    accB1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB2(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mod.ACC_B_2 already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB2.put((byte) 0);
    }
    accB2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB3(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mod.ACC_B_3 already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB3.put((byte) 0);
    }
    accB3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta0(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mod.ACC_DELTA_0 already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta0.put((byte) 0);
    }
    accDelta0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta1(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mod.ACC_DELTA_1 already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta1.put((byte) 0);
    }
    accDelta1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta2(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mod.ACC_DELTA_2 already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta2.put((byte) 0);
    }
    accDelta2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta3(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mod.ACC_DELTA_3 already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta3.put((byte) 0);
    }
    accDelta3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accH0(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mod.ACC_H_0 already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accH0.put((byte) 0);
    }
    accH0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accH1(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mod.ACC_H_1 already set");
    } else {
      filled.set(13);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accH1.put((byte) 0);
    }
    accH1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accH2(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mod.ACC_H_2 already set");
    } else {
      filled.set(14);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accH2.put((byte) 0);
    }
    accH2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ0(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mod.ACC_Q_0 already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ0.put((byte) 0);
    }
    accQ0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ1(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mod.ACC_Q_1 already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ1.put((byte) 0);
    }
    accQ1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ2(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mod.ACC_Q_2 already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ2.put((byte) 0);
    }
    accQ2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ3(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mod.ACC_Q_3 already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ3.put((byte) 0);
    }
    accQ3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR0(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mod.ACC_R_0 already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR0.put((byte) 0);
    }
    accR0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR1(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mod.ACC_R_1 already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR1.put((byte) 0);
    }
    accR1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR2(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mod.ACC_R_2 already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR2.put((byte) 0);
    }
    accR2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR3(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mod.ACC_R_3 already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR3.put((byte) 0);
    }
    accR3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mod.ARG_1_HI already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Hi.put((byte) 0);
    }
    arg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mod.ARG_1_LO already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Lo.put((byte) 0);
    }
    arg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mod.ARG_2_HI already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Hi.put((byte) 0);
    }
    arg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mod.ARG_2_LO already set");
    } else {
      filled.set(26);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Lo.put((byte) 0);
    }
    arg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byte12(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mod.BYTE_1_2 already set");
    } else {
      filled.set(27);
    }

    byte12.put(b.toByte());

    return this;
  }

  public Trace byte13(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mod.BYTE_1_3 already set");
    } else {
      filled.set(28);
    }

    byte13.put(b.toByte());

    return this;
  }

  public Trace byte22(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mod.BYTE_2_2 already set");
    } else {
      filled.set(29);
    }

    byte22.put(b.toByte());

    return this;
  }

  public Trace byte23(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mod.BYTE_2_3 already set");
    } else {
      filled.set(30);
    }

    byte23.put(b.toByte());

    return this;
  }

  public Trace byteB0(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mod.BYTE_B_0 already set");
    } else {
      filled.set(31);
    }

    byteB0.put(b.toByte());

    return this;
  }

  public Trace byteB1(final UnsignedByte b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mod.BYTE_B_1 already set");
    } else {
      filled.set(32);
    }

    byteB1.put(b.toByte());

    return this;
  }

  public Trace byteB2(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mod.BYTE_B_2 already set");
    } else {
      filled.set(33);
    }

    byteB2.put(b.toByte());

    return this;
  }

  public Trace byteB3(final UnsignedByte b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mod.BYTE_B_3 already set");
    } else {
      filled.set(34);
    }

    byteB3.put(b.toByte());

    return this;
  }

  public Trace byteDelta0(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mod.BYTE_DELTA_0 already set");
    } else {
      filled.set(35);
    }

    byteDelta0.put(b.toByte());

    return this;
  }

  public Trace byteDelta1(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mod.BYTE_DELTA_1 already set");
    } else {
      filled.set(36);
    }

    byteDelta1.put(b.toByte());

    return this;
  }

  public Trace byteDelta2(final UnsignedByte b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mod.BYTE_DELTA_2 already set");
    } else {
      filled.set(37);
    }

    byteDelta2.put(b.toByte());

    return this;
  }

  public Trace byteDelta3(final UnsignedByte b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mod.BYTE_DELTA_3 already set");
    } else {
      filled.set(38);
    }

    byteDelta3.put(b.toByte());

    return this;
  }

  public Trace byteH0(final UnsignedByte b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mod.BYTE_H_0 already set");
    } else {
      filled.set(39);
    }

    byteH0.put(b.toByte());

    return this;
  }

  public Trace byteH1(final UnsignedByte b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mod.BYTE_H_1 already set");
    } else {
      filled.set(40);
    }

    byteH1.put(b.toByte());

    return this;
  }

  public Trace byteH2(final UnsignedByte b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mod.BYTE_H_2 already set");
    } else {
      filled.set(41);
    }

    byteH2.put(b.toByte());

    return this;
  }

  public Trace byteQ0(final UnsignedByte b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mod.BYTE_Q_0 already set");
    } else {
      filled.set(42);
    }

    byteQ0.put(b.toByte());

    return this;
  }

  public Trace byteQ1(final UnsignedByte b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mod.BYTE_Q_1 already set");
    } else {
      filled.set(43);
    }

    byteQ1.put(b.toByte());

    return this;
  }

  public Trace byteQ2(final UnsignedByte b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mod.BYTE_Q_2 already set");
    } else {
      filled.set(44);
    }

    byteQ2.put(b.toByte());

    return this;
  }

  public Trace byteQ3(final UnsignedByte b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mod.BYTE_Q_3 already set");
    } else {
      filled.set(45);
    }

    byteQ3.put(b.toByte());

    return this;
  }

  public Trace byteR0(final UnsignedByte b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mod.BYTE_R_0 already set");
    } else {
      filled.set(46);
    }

    byteR0.put(b.toByte());

    return this;
  }

  public Trace byteR1(final UnsignedByte b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mod.BYTE_R_1 already set");
    } else {
      filled.set(47);
    }

    byteR1.put(b.toByte());

    return this;
  }

  public Trace byteR2(final UnsignedByte b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mod.BYTE_R_2 already set");
    } else {
      filled.set(48);
    }

    byteR2.put(b.toByte());

    return this;
  }

  public Trace byteR3(final UnsignedByte b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mod.BYTE_R_3 already set");
    } else {
      filled.set(49);
    }

    byteR3.put(b.toByte());

    return this;
  }

  public Trace cmp1(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mod.CMP_1 already set");
    } else {
      filled.set(50);
    }

    cmp1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace cmp2(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mod.CMP_2 already set");
    } else {
      filled.set(51);
    }

    cmp2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final Bytes b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mod.CT already set");
    } else {
      filled.set(52);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      ct.put((byte) 0);
    }
    ct.put(b.toArrayUnsafe());

    return this;
  }

  public Trace decOutput(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mod.DEC_OUTPUT already set");
    } else {
      filled.set(53);
    }

    decOutput.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace decSigned(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mod.DEC_SIGNED already set");
    } else {
      filled.set(54);
    }

    decSigned.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace inst(final Bytes b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mod.INST already set");
    } else {
      filled.set(55);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      inst.put((byte) 0);
    }
    inst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace msb1(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mod.MSB_1 already set");
    } else {
      filled.set(56);
    }

    msb1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace msb2(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mod.MSB_2 already set");
    } else {
      filled.set(57);
    }

    msb2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oli(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mod.OLI already set");
    } else {
      filled.set(58);
    }

    oli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mod.RES_HI already set");
    } else {
      filled.set(59);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resHi.put((byte) 0);
    }
    resHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mod.RES_LO already set");
    } else {
      filled.set(60);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resLo.put((byte) 0);
    }
    resLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final Bytes b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mod.STAMP already set");
    } else {
      filled.set(61);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mod.ACC_1_2 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mod.ACC_1_3 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mod.ACC_2_2 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mod.ACC_2_3 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mod.ACC_B_0 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mod.ACC_B_1 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mod.ACC_B_2 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mod.ACC_B_3 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mod.ACC_DELTA_0 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mod.ACC_DELTA_1 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mod.ACC_DELTA_2 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mod.ACC_DELTA_3 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mod.ACC_H_0 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mod.ACC_H_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mod.ACC_H_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mod.ACC_Q_0 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mod.ACC_Q_1 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mod.ACC_Q_2 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mod.ACC_Q_3 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mod.ACC_R_0 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mod.ACC_R_1 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mod.ACC_R_2 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mod.ACC_R_3 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mod.ARG_1_HI has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mod.ARG_1_LO has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mod.ARG_2_HI has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mod.ARG_2_LO has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mod.BYTE_1_2 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mod.BYTE_1_3 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mod.BYTE_2_2 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mod.BYTE_2_3 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mod.BYTE_B_0 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mod.BYTE_B_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mod.BYTE_B_2 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mod.BYTE_B_3 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mod.BYTE_DELTA_0 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mod.BYTE_DELTA_1 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mod.BYTE_DELTA_2 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mod.BYTE_DELTA_3 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mod.BYTE_H_0 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mod.BYTE_H_1 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mod.BYTE_H_2 has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mod.BYTE_Q_0 has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mod.BYTE_Q_1 has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mod.BYTE_Q_2 has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mod.BYTE_Q_3 has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mod.BYTE_R_0 has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mod.BYTE_R_1 has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mod.BYTE_R_2 has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mod.BYTE_R_3 has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mod.CMP_1 has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mod.CMP_2 has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("mod.CT has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("mod.DEC_OUTPUT has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mod.DEC_SIGNED has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mod.INST has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mod.MSB_1 has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mod.MSB_2 has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mod.OLI has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("mod.RES_HI has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mod.RES_LO has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("mod.STAMP has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc12.position(acc12.position() + 32);
    }

    if (!filled.get(1)) {
      acc13.position(acc13.position() + 32);
    }

    if (!filled.get(2)) {
      acc22.position(acc22.position() + 32);
    }

    if (!filled.get(3)) {
      acc23.position(acc23.position() + 32);
    }

    if (!filled.get(4)) {
      accB0.position(accB0.position() + 32);
    }

    if (!filled.get(5)) {
      accB1.position(accB1.position() + 32);
    }

    if (!filled.get(6)) {
      accB2.position(accB2.position() + 32);
    }

    if (!filled.get(7)) {
      accB3.position(accB3.position() + 32);
    }

    if (!filled.get(8)) {
      accDelta0.position(accDelta0.position() + 32);
    }

    if (!filled.get(9)) {
      accDelta1.position(accDelta1.position() + 32);
    }

    if (!filled.get(10)) {
      accDelta2.position(accDelta2.position() + 32);
    }

    if (!filled.get(11)) {
      accDelta3.position(accDelta3.position() + 32);
    }

    if (!filled.get(12)) {
      accH0.position(accH0.position() + 32);
    }

    if (!filled.get(13)) {
      accH1.position(accH1.position() + 32);
    }

    if (!filled.get(14)) {
      accH2.position(accH2.position() + 32);
    }

    if (!filled.get(15)) {
      accQ0.position(accQ0.position() + 32);
    }

    if (!filled.get(16)) {
      accQ1.position(accQ1.position() + 32);
    }

    if (!filled.get(17)) {
      accQ2.position(accQ2.position() + 32);
    }

    if (!filled.get(18)) {
      accQ3.position(accQ3.position() + 32);
    }

    if (!filled.get(19)) {
      accR0.position(accR0.position() + 32);
    }

    if (!filled.get(20)) {
      accR1.position(accR1.position() + 32);
    }

    if (!filled.get(21)) {
      accR2.position(accR2.position() + 32);
    }

    if (!filled.get(22)) {
      accR3.position(accR3.position() + 32);
    }

    if (!filled.get(23)) {
      arg1Hi.position(arg1Hi.position() + 32);
    }

    if (!filled.get(24)) {
      arg1Lo.position(arg1Lo.position() + 32);
    }

    if (!filled.get(25)) {
      arg2Hi.position(arg2Hi.position() + 32);
    }

    if (!filled.get(26)) {
      arg2Lo.position(arg2Lo.position() + 32);
    }

    if (!filled.get(27)) {
      byte12.position(byte12.position() + 1);
    }

    if (!filled.get(28)) {
      byte13.position(byte13.position() + 1);
    }

    if (!filled.get(29)) {
      byte22.position(byte22.position() + 1);
    }

    if (!filled.get(30)) {
      byte23.position(byte23.position() + 1);
    }

    if (!filled.get(31)) {
      byteB0.position(byteB0.position() + 1);
    }

    if (!filled.get(32)) {
      byteB1.position(byteB1.position() + 1);
    }

    if (!filled.get(33)) {
      byteB2.position(byteB2.position() + 1);
    }

    if (!filled.get(34)) {
      byteB3.position(byteB3.position() + 1);
    }

    if (!filled.get(35)) {
      byteDelta0.position(byteDelta0.position() + 1);
    }

    if (!filled.get(36)) {
      byteDelta1.position(byteDelta1.position() + 1);
    }

    if (!filled.get(37)) {
      byteDelta2.position(byteDelta2.position() + 1);
    }

    if (!filled.get(38)) {
      byteDelta3.position(byteDelta3.position() + 1);
    }

    if (!filled.get(39)) {
      byteH0.position(byteH0.position() + 1);
    }

    if (!filled.get(40)) {
      byteH1.position(byteH1.position() + 1);
    }

    if (!filled.get(41)) {
      byteH2.position(byteH2.position() + 1);
    }

    if (!filled.get(42)) {
      byteQ0.position(byteQ0.position() + 1);
    }

    if (!filled.get(43)) {
      byteQ1.position(byteQ1.position() + 1);
    }

    if (!filled.get(44)) {
      byteQ2.position(byteQ2.position() + 1);
    }

    if (!filled.get(45)) {
      byteQ3.position(byteQ3.position() + 1);
    }

    if (!filled.get(46)) {
      byteR0.position(byteR0.position() + 1);
    }

    if (!filled.get(47)) {
      byteR1.position(byteR1.position() + 1);
    }

    if (!filled.get(48)) {
      byteR2.position(byteR2.position() + 1);
    }

    if (!filled.get(49)) {
      byteR3.position(byteR3.position() + 1);
    }

    if (!filled.get(50)) {
      cmp1.position(cmp1.position() + 1);
    }

    if (!filled.get(51)) {
      cmp2.position(cmp2.position() + 1);
    }

    if (!filled.get(52)) {
      ct.position(ct.position() + 32);
    }

    if (!filled.get(53)) {
      decOutput.position(decOutput.position() + 1);
    }

    if (!filled.get(54)) {
      decSigned.position(decSigned.position() + 1);
    }

    if (!filled.get(55)) {
      inst.position(inst.position() + 32);
    }

    if (!filled.get(56)) {
      msb1.position(msb1.position() + 1);
    }

    if (!filled.get(57)) {
      msb2.position(msb2.position() + 1);
    }

    if (!filled.get(58)) {
      oli.position(oli.position() + 1);
    }

    if (!filled.get(59)) {
      resHi.position(resHi.position() + 32);
    }

    if (!filled.get(60)) {
      resLo.position(resLo.position() + 32);
    }

    if (!filled.get(61)) {
      stamp.position(stamp.position() + 32);
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
