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

package net.consensys.linea.zktracer.module.ec_data;

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

  private final MappedByteBuffer accDelta;
  private final MappedByteBuffer accPairings;
  private final MappedByteBuffer allChecksPassed;
  private final MappedByteBuffer byteDelta;
  private final MappedByteBuffer comparisons;
  private final MappedByteBuffer ctMin;
  private final MappedByteBuffer cube;
  private final MappedByteBuffer ecAdd;
  private final MappedByteBuffer ecMul;
  private final MappedByteBuffer ecPairing;
  private final MappedByteBuffer ecRecover;
  private final MappedByteBuffer equalities;
  private final MappedByteBuffer extArg1Hi;
  private final MappedByteBuffer extArg1Lo;
  private final MappedByteBuffer extArg2Hi;
  private final MappedByteBuffer extArg2Lo;
  private final MappedByteBuffer extArg3Hi;
  private final MappedByteBuffer extArg3Lo;
  private final MappedByteBuffer extInst;
  private final MappedByteBuffer extResHi;
  private final MappedByteBuffer extResLo;
  private final MappedByteBuffer hurdle;
  private final MappedByteBuffer index;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer preliminaryChecksPassed;
  private final MappedByteBuffer somethingWasntOnG2;
  private final MappedByteBuffer square;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer thisIsNotOnG2;
  private final MappedByteBuffer thisIsNotOnG2Acc;
  private final MappedByteBuffer totalPairings;
  private final MappedByteBuffer type;
  private final MappedByteBuffer wcpArg1Hi;
  private final MappedByteBuffer wcpArg1Lo;
  private final MappedByteBuffer wcpArg2Hi;
  private final MappedByteBuffer wcpArg2Lo;
  private final MappedByteBuffer wcpInst;
  private final MappedByteBuffer wcpRes;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("ec_data.ACC_DELTA", 32, length),
        new ColumnHeader("ec_data.ACC_PAIRINGS", 32, length),
        new ColumnHeader("ec_data.ALL_CHECKS_PASSED", 1, length),
        new ColumnHeader("ec_data.BYTE_DELTA", 1, length),
        new ColumnHeader("ec_data.COMPARISONS", 1, length),
        new ColumnHeader("ec_data.CT_MIN", 32, length),
        new ColumnHeader("ec_data.CUBE", 32, length),
        new ColumnHeader("ec_data.EC_ADD", 1, length),
        new ColumnHeader("ec_data.EC_MUL", 1, length),
        new ColumnHeader("ec_data.EC_PAIRING", 1, length),
        new ColumnHeader("ec_data.EC_RECOVER", 1, length),
        new ColumnHeader("ec_data.EQUALITIES", 1, length),
        new ColumnHeader("ec_data.EXT_ARG1_HI", 32, length),
        new ColumnHeader("ec_data.EXT_ARG1_LO", 32, length),
        new ColumnHeader("ec_data.EXT_ARG2_HI", 32, length),
        new ColumnHeader("ec_data.EXT_ARG2_LO", 32, length),
        new ColumnHeader("ec_data.EXT_ARG3_HI", 32, length),
        new ColumnHeader("ec_data.EXT_ARG3_LO", 32, length),
        new ColumnHeader("ec_data.EXT_INST", 32, length),
        new ColumnHeader("ec_data.EXT_RES_HI", 32, length),
        new ColumnHeader("ec_data.EXT_RES_LO", 32, length),
        new ColumnHeader("ec_data.HURDLE", 1, length),
        new ColumnHeader("ec_data.INDEX", 32, length),
        new ColumnHeader("ec_data.LIMB", 32, length),
        new ColumnHeader("ec_data.PRELIMINARY_CHECKS_PASSED", 1, length),
        new ColumnHeader("ec_data.SOMETHING_WASNT_ON_G2", 1, length),
        new ColumnHeader("ec_data.SQUARE", 32, length),
        new ColumnHeader("ec_data.STAMP", 32, length),
        new ColumnHeader("ec_data.THIS_IS_NOT_ON_G2", 1, length),
        new ColumnHeader("ec_data.THIS_IS_NOT_ON_G2_ACC", 1, length),
        new ColumnHeader("ec_data.TOTAL_PAIRINGS", 32, length),
        new ColumnHeader("ec_data.TYPE", 32, length),
        new ColumnHeader("ec_data.WCP_ARG1_HI", 32, length),
        new ColumnHeader("ec_data.WCP_ARG1_LO", 32, length),
        new ColumnHeader("ec_data.WCP_ARG2_HI", 32, length),
        new ColumnHeader("ec_data.WCP_ARG2_LO", 32, length),
        new ColumnHeader("ec_data.WCP_INST", 32, length),
        new ColumnHeader("ec_data.WCP_RES", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.accDelta = buffers.get(0);
    this.accPairings = buffers.get(1);
    this.allChecksPassed = buffers.get(2);
    this.byteDelta = buffers.get(3);
    this.comparisons = buffers.get(4);
    this.ctMin = buffers.get(5);
    this.cube = buffers.get(6);
    this.ecAdd = buffers.get(7);
    this.ecMul = buffers.get(8);
    this.ecPairing = buffers.get(9);
    this.ecRecover = buffers.get(10);
    this.equalities = buffers.get(11);
    this.extArg1Hi = buffers.get(12);
    this.extArg1Lo = buffers.get(13);
    this.extArg2Hi = buffers.get(14);
    this.extArg2Lo = buffers.get(15);
    this.extArg3Hi = buffers.get(16);
    this.extArg3Lo = buffers.get(17);
    this.extInst = buffers.get(18);
    this.extResHi = buffers.get(19);
    this.extResLo = buffers.get(20);
    this.hurdle = buffers.get(21);
    this.index = buffers.get(22);
    this.limb = buffers.get(23);
    this.preliminaryChecksPassed = buffers.get(24);
    this.somethingWasntOnG2 = buffers.get(25);
    this.square = buffers.get(26);
    this.stamp = buffers.get(27);
    this.thisIsNotOnG2 = buffers.get(28);
    this.thisIsNotOnG2Acc = buffers.get(29);
    this.totalPairings = buffers.get(30);
    this.type = buffers.get(31);
    this.wcpArg1Hi = buffers.get(32);
    this.wcpArg1Lo = buffers.get(33);
    this.wcpArg2Hi = buffers.get(34);
    this.wcpArg2Lo = buffers.get(35);
    this.wcpInst = buffers.get(36);
    this.wcpRes = buffers.get(37);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace accDelta(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("ec_data.ACC_DELTA already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta.put((byte) 0);
    }
    accDelta.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accPairings(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("ec_data.ACC_PAIRINGS already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accPairings.put((byte) 0);
    }
    accPairings.put(b.toArrayUnsafe());

    return this;
  }

  public Trace allChecksPassed(final Boolean b) {
    if (filled.get(2)) {
      throw new IllegalStateException("ec_data.ALL_CHECKS_PASSED already set");
    } else {
      filled.set(2);
    }

    allChecksPassed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byteDelta(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("ec_data.BYTE_DELTA already set");
    } else {
      filled.set(3);
    }

    byteDelta.put(b.toByte());

    return this;
  }

  public Trace comparisons(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("ec_data.COMPARISONS already set");
    } else {
      filled.set(4);
    }

    comparisons.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ctMin(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("ec_data.CT_MIN already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      ctMin.put((byte) 0);
    }
    ctMin.put(b.toArrayUnsafe());

    return this;
  }

  public Trace cube(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("ec_data.CUBE already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cube.put((byte) 0);
    }
    cube.put(b.toArrayUnsafe());

    return this;
  }

  public Trace ecAdd(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("ec_data.EC_ADD already set");
    } else {
      filled.set(7);
    }

    ecAdd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ecMul(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("ec_data.EC_MUL already set");
    } else {
      filled.set(8);
    }

    ecMul.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ecPairing(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("ec_data.EC_PAIRING already set");
    } else {
      filled.set(9);
    }

    ecPairing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ecRecover(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("ec_data.EC_RECOVER already set");
    } else {
      filled.set(10);
    }

    ecRecover.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace equalities(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("ec_data.EQUALITIES already set");
    } else {
      filled.set(11);
    }

    equalities.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace extArg1Hi(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("ec_data.EXT_ARG1_HI already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extArg1Hi.put((byte) 0);
    }
    extArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extArg1Lo(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("ec_data.EXT_ARG1_LO already set");
    } else {
      filled.set(13);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extArg1Lo.put((byte) 0);
    }
    extArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extArg2Hi(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("ec_data.EXT_ARG2_HI already set");
    } else {
      filled.set(14);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extArg2Hi.put((byte) 0);
    }
    extArg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extArg2Lo(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("ec_data.EXT_ARG2_LO already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extArg2Lo.put((byte) 0);
    }
    extArg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extArg3Hi(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("ec_data.EXT_ARG3_HI already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extArg3Hi.put((byte) 0);
    }
    extArg3Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extArg3Lo(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("ec_data.EXT_ARG3_LO already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extArg3Lo.put((byte) 0);
    }
    extArg3Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extInst(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("ec_data.EXT_INST already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extInst.put((byte) 0);
    }
    extInst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extResHi(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("ec_data.EXT_RES_HI already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extResHi.put((byte) 0);
    }
    extResHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace extResLo(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("ec_data.EXT_RES_LO already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      extResLo.put((byte) 0);
    }
    extResLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace hurdle(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("ec_data.HURDLE already set");
    } else {
      filled.set(21);
    }

    hurdle.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace index(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("ec_data.INDEX already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      index.put((byte) 0);
    }
    index.put(b.toArrayUnsafe());

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("ec_data.LIMB already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace preliminaryChecksPassed(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("ec_data.PRELIMINARY_CHECKS_PASSED already set");
    } else {
      filled.set(24);
    }

    preliminaryChecksPassed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace somethingWasntOnG2(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("ec_data.SOMETHING_WASNT_ON_G2 already set");
    } else {
      filled.set(25);
    }

    somethingWasntOnG2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace square(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("ec_data.SQUARE already set");
    } else {
      filled.set(26);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      square.put((byte) 0);
    }
    square.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("ec_data.STAMP already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace thisIsNotOnG2(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("ec_data.THIS_IS_NOT_ON_G2 already set");
    } else {
      filled.set(28);
    }

    thisIsNotOnG2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace thisIsNotOnG2Acc(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("ec_data.THIS_IS_NOT_ON_G2_ACC already set");
    } else {
      filled.set(29);
    }

    thisIsNotOnG2Acc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace totalPairings(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("ec_data.TOTAL_PAIRINGS already set");
    } else {
      filled.set(30);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      totalPairings.put((byte) 0);
    }
    totalPairings.put(b.toArrayUnsafe());

    return this;
  }

  public Trace type(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("ec_data.TYPE already set");
    } else {
      filled.set(31);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      type.put((byte) 0);
    }
    type.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpArg1Hi(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("ec_data.WCP_ARG1_HI already set");
    } else {
      filled.set(32);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpArg1Hi.put((byte) 0);
    }
    wcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpArg1Lo(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("ec_data.WCP_ARG1_LO already set");
    } else {
      filled.set(33);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpArg1Lo.put((byte) 0);
    }
    wcpArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpArg2Hi(final Bytes b) {
    if (filled.get(34)) {
      throw new IllegalStateException("ec_data.WCP_ARG2_HI already set");
    } else {
      filled.set(34);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpArg2Hi.put((byte) 0);
    }
    wcpArg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpArg2Lo(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("ec_data.WCP_ARG2_LO already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpArg2Lo.put((byte) 0);
    }
    wcpArg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpInst(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("ec_data.WCP_INST already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpInst.put((byte) 0);
    }
    wcpInst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wcpRes(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("ec_data.WCP_RES already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wcpRes.put((byte) 0);
    }
    wcpRes.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("ec_data.ACC_DELTA has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("ec_data.ACC_PAIRINGS has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("ec_data.ALL_CHECKS_PASSED has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("ec_data.BYTE_DELTA has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("ec_data.COMPARISONS has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("ec_data.CT_MIN has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("ec_data.CUBE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("ec_data.EC_ADD has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("ec_data.EC_MUL has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("ec_data.EC_PAIRING has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("ec_data.EC_RECOVER has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("ec_data.EQUALITIES has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("ec_data.EXT_ARG1_HI has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("ec_data.EXT_ARG1_LO has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("ec_data.EXT_ARG2_HI has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("ec_data.EXT_ARG2_LO has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("ec_data.EXT_ARG3_HI has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("ec_data.EXT_ARG3_LO has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("ec_data.EXT_INST has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("ec_data.EXT_RES_HI has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("ec_data.EXT_RES_LO has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("ec_data.HURDLE has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("ec_data.INDEX has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("ec_data.LIMB has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("ec_data.PRELIMINARY_CHECKS_PASSED has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("ec_data.SOMETHING_WASNT_ON_G2 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("ec_data.SQUARE has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("ec_data.STAMP has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("ec_data.THIS_IS_NOT_ON_G2 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("ec_data.THIS_IS_NOT_ON_G2_ACC has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("ec_data.TOTAL_PAIRINGS has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("ec_data.TYPE has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("ec_data.WCP_ARG1_HI has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("ec_data.WCP_ARG1_LO has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("ec_data.WCP_ARG2_HI has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("ec_data.WCP_ARG2_LO has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("ec_data.WCP_INST has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("ec_data.WCP_RES has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      accDelta.position(accDelta.position() + 32);
    }

    if (!filled.get(1)) {
      accPairings.position(accPairings.position() + 32);
    }

    if (!filled.get(2)) {
      allChecksPassed.position(allChecksPassed.position() + 1);
    }

    if (!filled.get(3)) {
      byteDelta.position(byteDelta.position() + 1);
    }

    if (!filled.get(4)) {
      comparisons.position(comparisons.position() + 1);
    }

    if (!filled.get(5)) {
      ctMin.position(ctMin.position() + 32);
    }

    if (!filled.get(6)) {
      cube.position(cube.position() + 32);
    }

    if (!filled.get(7)) {
      ecAdd.position(ecAdd.position() + 1);
    }

    if (!filled.get(8)) {
      ecMul.position(ecMul.position() + 1);
    }

    if (!filled.get(9)) {
      ecPairing.position(ecPairing.position() + 1);
    }

    if (!filled.get(10)) {
      ecRecover.position(ecRecover.position() + 1);
    }

    if (!filled.get(11)) {
      equalities.position(equalities.position() + 1);
    }

    if (!filled.get(12)) {
      extArg1Hi.position(extArg1Hi.position() + 32);
    }

    if (!filled.get(13)) {
      extArg1Lo.position(extArg1Lo.position() + 32);
    }

    if (!filled.get(14)) {
      extArg2Hi.position(extArg2Hi.position() + 32);
    }

    if (!filled.get(15)) {
      extArg2Lo.position(extArg2Lo.position() + 32);
    }

    if (!filled.get(16)) {
      extArg3Hi.position(extArg3Hi.position() + 32);
    }

    if (!filled.get(17)) {
      extArg3Lo.position(extArg3Lo.position() + 32);
    }

    if (!filled.get(18)) {
      extInst.position(extInst.position() + 32);
    }

    if (!filled.get(19)) {
      extResHi.position(extResHi.position() + 32);
    }

    if (!filled.get(20)) {
      extResLo.position(extResLo.position() + 32);
    }

    if (!filled.get(21)) {
      hurdle.position(hurdle.position() + 1);
    }

    if (!filled.get(22)) {
      index.position(index.position() + 32);
    }

    if (!filled.get(23)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(24)) {
      preliminaryChecksPassed.position(preliminaryChecksPassed.position() + 1);
    }

    if (!filled.get(25)) {
      somethingWasntOnG2.position(somethingWasntOnG2.position() + 1);
    }

    if (!filled.get(26)) {
      square.position(square.position() + 32);
    }

    if (!filled.get(27)) {
      stamp.position(stamp.position() + 32);
    }

    if (!filled.get(28)) {
      thisIsNotOnG2.position(thisIsNotOnG2.position() + 1);
    }

    if (!filled.get(29)) {
      thisIsNotOnG2Acc.position(thisIsNotOnG2Acc.position() + 1);
    }

    if (!filled.get(30)) {
      totalPairings.position(totalPairings.position() + 32);
    }

    if (!filled.get(31)) {
      type.position(type.position() + 32);
    }

    if (!filled.get(32)) {
      wcpArg1Hi.position(wcpArg1Hi.position() + 32);
    }

    if (!filled.get(33)) {
      wcpArg1Lo.position(wcpArg1Lo.position() + 32);
    }

    if (!filled.get(34)) {
      wcpArg2Hi.position(wcpArg2Hi.position() + 32);
    }

    if (!filled.get(35)) {
      wcpArg2Lo.position(wcpArg2Lo.position() + 32);
    }

    if (!filled.get(36)) {
      wcpInst.position(wcpInst.position() + 32);
    }

    if (!filled.get(37)) {
      wcpRes.position(wcpRes.position() + 32);
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
