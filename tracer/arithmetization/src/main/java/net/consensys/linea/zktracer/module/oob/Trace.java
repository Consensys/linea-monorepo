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

package net.consensys.linea.zktracer.module.oob;

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
  public static final int ADD = 0x1;
  public static final int CT_MAX_BLAKE2F_cds = 0x1;
  public static final int CT_MAX_BLAKE2F_params = 0x1;
  public static final int CT_MAX_CALL = 0x2;
  public static final int CT_MAX_CDL = 0x0;
  public static final int CT_MAX_CREATE = 0x2;
  public static final int CT_MAX_DEPLOYMENT = 0x0;
  public static final int CT_MAX_ECADD = 0x2;
  public static final int CT_MAX_ECMUL = 0x2;
  public static final int CT_MAX_ECPAIRING = 0x4;
  public static final int CT_MAX_ECRECOVER = 0x2;
  public static final int CT_MAX_IDENTITY = 0x3;
  public static final int CT_MAX_JUMP = 0x0;
  public static final int CT_MAX_JUMPI = 0x1;
  public static final int CT_MAX_MODEXP_cds = 0x2;
  public static final int CT_MAX_MODEXP_extract = 0x3;
  public static final int CT_MAX_MODEXP_lead = 0x3;
  public static final int CT_MAX_MODEXP_pricing = 0x5;
  public static final int CT_MAX_MODEXP_xbs = 0x2;
  public static final int CT_MAX_RDC = 0x2;
  public static final int CT_MAX_RIPEMD = 0x3;
  public static final int CT_MAX_SHA2 = 0x3;
  public static final int CT_MAX_SSTORE = 0x0;
  public static final int CT_MAX_XCALL = 0x0;
  public static final int DIV = 0x4;
  public static final int EQ = 0x14;
  public static final int GT = 0x11;
  public static final int G_CALLSTIPEND = 0x8fc;
  public static final int G_QUADDIVISOR = 0x3;
  public static final int ISZERO = 0x15;
  public static final int LT = 0x10;
  public static final int MOD = 0x6;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer addFlag;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer data1;
  private final MappedByteBuffer data2;
  private final MappedByteBuffer data3;
  private final MappedByteBuffer data4;
  private final MappedByteBuffer data5;
  private final MappedByteBuffer data6;
  private final MappedByteBuffer data7;
  private final MappedByteBuffer data8;
  private final MappedByteBuffer isBlake2FCds;
  private final MappedByteBuffer isBlake2FParams;
  private final MappedByteBuffer isCall;
  private final MappedByteBuffer isCdl;
  private final MappedByteBuffer isCreate;
  private final MappedByteBuffer isDeployment;
  private final MappedByteBuffer isEcadd;
  private final MappedByteBuffer isEcmul;
  private final MappedByteBuffer isEcpairing;
  private final MappedByteBuffer isEcrecover;
  private final MappedByteBuffer isIdentity;
  private final MappedByteBuffer isJump;
  private final MappedByteBuffer isJumpi;
  private final MappedByteBuffer isModexpCds;
  private final MappedByteBuffer isModexpExtract;
  private final MappedByteBuffer isModexpLead;
  private final MappedByteBuffer isModexpPricing;
  private final MappedByteBuffer isModexpXbs;
  private final MappedByteBuffer isRdc;
  private final MappedByteBuffer isRipemd;
  private final MappedByteBuffer isSha2;
  private final MappedByteBuffer isSstore;
  private final MappedByteBuffer isXcall;
  private final MappedByteBuffer modFlag;
  private final MappedByteBuffer oobInst;
  private final MappedByteBuffer outgoingData1;
  private final MappedByteBuffer outgoingData2;
  private final MappedByteBuffer outgoingData3;
  private final MappedByteBuffer outgoingData4;
  private final MappedByteBuffer outgoingInst;
  private final MappedByteBuffer outgoingResLo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("oob.ADD_FLAG", 1, length),
        new ColumnHeader("oob.CT", 2, length),
        new ColumnHeader("oob.CT_MAX", 2, length),
        new ColumnHeader("oob.DATA_1", 32, length),
        new ColumnHeader("oob.DATA_2", 32, length),
        new ColumnHeader("oob.DATA_3", 32, length),
        new ColumnHeader("oob.DATA_4", 32, length),
        new ColumnHeader("oob.DATA_5", 32, length),
        new ColumnHeader("oob.DATA_6", 32, length),
        new ColumnHeader("oob.DATA_7", 32, length),
        new ColumnHeader("oob.DATA_8", 32, length),
        new ColumnHeader("oob.IS_BLAKE2F_cds", 1, length),
        new ColumnHeader("oob.IS_BLAKE2F_params", 1, length),
        new ColumnHeader("oob.IS_CALL", 1, length),
        new ColumnHeader("oob.IS_CDL", 1, length),
        new ColumnHeader("oob.IS_CREATE", 1, length),
        new ColumnHeader("oob.IS_DEPLOYMENT", 1, length),
        new ColumnHeader("oob.IS_ECADD", 1, length),
        new ColumnHeader("oob.IS_ECMUL", 1, length),
        new ColumnHeader("oob.IS_ECPAIRING", 1, length),
        new ColumnHeader("oob.IS_ECRECOVER", 1, length),
        new ColumnHeader("oob.IS_IDENTITY", 1, length),
        new ColumnHeader("oob.IS_JUMP", 1, length),
        new ColumnHeader("oob.IS_JUMPI", 1, length),
        new ColumnHeader("oob.IS_MODEXP_cds", 1, length),
        new ColumnHeader("oob.IS_MODEXP_extract", 1, length),
        new ColumnHeader("oob.IS_MODEXP_lead", 1, length),
        new ColumnHeader("oob.IS_MODEXP_pricing", 1, length),
        new ColumnHeader("oob.IS_MODEXP_xbs", 1, length),
        new ColumnHeader("oob.IS_RDC", 1, length),
        new ColumnHeader("oob.IS_RIPEMD", 1, length),
        new ColumnHeader("oob.IS_SHA2", 1, length),
        new ColumnHeader("oob.IS_SSTORE", 1, length),
        new ColumnHeader("oob.IS_XCALL", 1, length),
        new ColumnHeader("oob.MOD_FLAG", 1, length),
        new ColumnHeader("oob.OOB_INST", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_1", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_2", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_3", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_4", 32, length),
        new ColumnHeader("oob.OUTGOING_INST", 1, length),
        new ColumnHeader("oob.OUTGOING_RES_LO", 32, length),
        new ColumnHeader("oob.STAMP", 8, length),
        new ColumnHeader("oob.WCP_FLAG", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.addFlag = buffers.get(0);
    this.ct = buffers.get(1);
    this.ctMax = buffers.get(2);
    this.data1 = buffers.get(3);
    this.data2 = buffers.get(4);
    this.data3 = buffers.get(5);
    this.data4 = buffers.get(6);
    this.data5 = buffers.get(7);
    this.data6 = buffers.get(8);
    this.data7 = buffers.get(9);
    this.data8 = buffers.get(10);
    this.isBlake2FCds = buffers.get(11);
    this.isBlake2FParams = buffers.get(12);
    this.isCall = buffers.get(13);
    this.isCdl = buffers.get(14);
    this.isCreate = buffers.get(15);
    this.isDeployment = buffers.get(16);
    this.isEcadd = buffers.get(17);
    this.isEcmul = buffers.get(18);
    this.isEcpairing = buffers.get(19);
    this.isEcrecover = buffers.get(20);
    this.isIdentity = buffers.get(21);
    this.isJump = buffers.get(22);
    this.isJumpi = buffers.get(23);
    this.isModexpCds = buffers.get(24);
    this.isModexpExtract = buffers.get(25);
    this.isModexpLead = buffers.get(26);
    this.isModexpPricing = buffers.get(27);
    this.isModexpXbs = buffers.get(28);
    this.isRdc = buffers.get(29);
    this.isRipemd = buffers.get(30);
    this.isSha2 = buffers.get(31);
    this.isSstore = buffers.get(32);
    this.isXcall = buffers.get(33);
    this.modFlag = buffers.get(34);
    this.oobInst = buffers.get(35);
    this.outgoingData1 = buffers.get(36);
    this.outgoingData2 = buffers.get(37);
    this.outgoingData3 = buffers.get(38);
    this.outgoingData4 = buffers.get(39);
    this.outgoingInst = buffers.get(40);
    this.outgoingResLo = buffers.get(41);
    this.stamp = buffers.get(42);
    this.wcpFlag = buffers.get(43);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace addFlag(final Boolean b) {
    if (filled.get(0)) {
      throw new IllegalStateException("oob.ADD_FLAG already set");
    } else {
      filled.set(0);
    }

    addFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final short b) {
    if (filled.get(1)) {
      throw new IllegalStateException("oob.CT already set");
    } else {
      filled.set(1);
    }

    ct.putShort(b);

    return this;
  }

  public Trace ctMax(final short b) {
    if (filled.get(2)) {
      throw new IllegalStateException("oob.CT_MAX already set");
    } else {
      filled.set(2);
    }

    ctMax.putShort(b);

    return this;
  }

  public Trace data1(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("oob.DATA_1 already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data1.put((byte) 0);
    }
    data1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data2(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("oob.DATA_2 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data2.put((byte) 0);
    }
    data2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data3(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("oob.DATA_3 already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data3.put((byte) 0);
    }
    data3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data4(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("oob.DATA_4 already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data4.put((byte) 0);
    }
    data4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data5(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("oob.DATA_5 already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data5.put((byte) 0);
    }
    data5.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data6(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("oob.DATA_6 already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data6.put((byte) 0);
    }
    data6.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data7(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("oob.DATA_7 already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data7.put((byte) 0);
    }
    data7.put(b.toArrayUnsafe());

    return this;
  }

  public Trace data8(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("oob.DATA_8 already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data8.put((byte) 0);
    }
    data8.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isBlake2FCds(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_cds already set");
    } else {
      filled.set(11);
    }

    isBlake2FCds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isBlake2FParams(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_params already set");
    } else {
      filled.set(12);
    }

    isBlake2FParams.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCall(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("oob.IS_CALL already set");
    } else {
      filled.set(13);
    }

    isCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCdl(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("oob.IS_CDL already set");
    } else {
      filled.set(14);
    }

    isCdl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCreate(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("oob.IS_CREATE already set");
    } else {
      filled.set(15);
    }

    isCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isDeployment(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("oob.IS_DEPLOYMENT already set");
    } else {
      filled.set(16);
    }

    isDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcadd(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("oob.IS_ECADD already set");
    } else {
      filled.set(17);
    }

    isEcadd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcmul(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("oob.IS_ECMUL already set");
    } else {
      filled.set(18);
    }

    isEcmul.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcpairing(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("oob.IS_ECPAIRING already set");
    } else {
      filled.set(19);
    }

    isEcpairing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcrecover(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("oob.IS_ECRECOVER already set");
    } else {
      filled.set(20);
    }

    isEcrecover.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isIdentity(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("oob.IS_IDENTITY already set");
    } else {
      filled.set(21);
    }

    isIdentity.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJump(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("oob.IS_JUMP already set");
    } else {
      filled.set(22);
    }

    isJump.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJumpi(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("oob.IS_JUMPI already set");
    } else {
      filled.set(23);
    }

    isJumpi.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpCds(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("oob.IS_MODEXP_cds already set");
    } else {
      filled.set(24);
    }

    isModexpCds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpExtract(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("oob.IS_MODEXP_extract already set");
    } else {
      filled.set(25);
    }

    isModexpExtract.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpLead(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("oob.IS_MODEXP_lead already set");
    } else {
      filled.set(26);
    }

    isModexpLead.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpPricing(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("oob.IS_MODEXP_pricing already set");
    } else {
      filled.set(27);
    }

    isModexpPricing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpXbs(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("oob.IS_MODEXP_xbs already set");
    } else {
      filled.set(28);
    }

    isModexpXbs.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRdc(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("oob.IS_RDC already set");
    } else {
      filled.set(29);
    }

    isRdc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRipemd(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("oob.IS_RIPEMD already set");
    } else {
      filled.set(30);
    }

    isRipemd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSha2(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("oob.IS_SHA2 already set");
    } else {
      filled.set(31);
    }

    isSha2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSstore(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("oob.IS_SSTORE already set");
    } else {
      filled.set(32);
    }

    isSstore.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isXcall(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("oob.IS_XCALL already set");
    } else {
      filled.set(33);
    }

    isXcall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace modFlag(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("oob.MOD_FLAG already set");
    } else {
      filled.set(34);
    }

    modFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oobInst(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("oob.OOB_INST already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      oobInst.put((byte) 0);
    }
    oobInst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData1(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_1 already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData1.put((byte) 0);
    }
    outgoingData1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData2(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_2 already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData2.put((byte) 0);
    }
    outgoingData2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData3(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_3 already set");
    } else {
      filled.set(38);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData3.put((byte) 0);
    }
    outgoingData3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData4(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_4 already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData4.put((byte) 0);
    }
    outgoingData4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingInst(final UnsignedByte b) {
    if (filled.get(40)) {
      throw new IllegalStateException("oob.OUTGOING_INST already set");
    } else {
      filled.set(40);
    }

    outgoingInst.put(b.toByte());

    return this;
  }

  public Trace outgoingResLo(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("oob.OUTGOING_RES_LO already set");
    } else {
      filled.set(41);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingResLo.put((byte) 0);
    }
    outgoingResLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(42)) {
      throw new IllegalStateException("oob.STAMP already set");
    } else {
      filled.set(42);
    }

    stamp.putLong(b);

    return this;
  }

  public Trace wcpFlag(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("oob.WCP_FLAG already set");
    } else {
      filled.set(43);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("oob.ADD_FLAG has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("oob.CT has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("oob.CT_MAX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("oob.DATA_1 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("oob.DATA_2 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("oob.DATA_3 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("oob.DATA_4 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("oob.DATA_5 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("oob.DATA_6 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("oob.DATA_7 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("oob.DATA_8 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_cds has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_params has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("oob.IS_CALL has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("oob.IS_CDL has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("oob.IS_CREATE has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("oob.IS_DEPLOYMENT has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("oob.IS_ECADD has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("oob.IS_ECMUL has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("oob.IS_ECPAIRING has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("oob.IS_ECRECOVER has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("oob.IS_IDENTITY has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("oob.IS_JUMP has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("oob.IS_JUMPI has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("oob.IS_MODEXP_cds has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("oob.IS_MODEXP_extract has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("oob.IS_MODEXP_lead has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("oob.IS_MODEXP_pricing has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("oob.IS_MODEXP_xbs has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("oob.IS_RDC has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("oob.IS_RIPEMD has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("oob.IS_SHA2 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("oob.IS_SSTORE has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("oob.IS_XCALL has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("oob.MOD_FLAG has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("oob.OOB_INST has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_1 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_2 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_3 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_4 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("oob.OUTGOING_INST has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("oob.OUTGOING_RES_LO has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("oob.STAMP has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("oob.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      addFlag.position(addFlag.position() + 1);
    }

    if (!filled.get(1)) {
      ct.position(ct.position() + 2);
    }

    if (!filled.get(2)) {
      ctMax.position(ctMax.position() + 2);
    }

    if (!filled.get(3)) {
      data1.position(data1.position() + 32);
    }

    if (!filled.get(4)) {
      data2.position(data2.position() + 32);
    }

    if (!filled.get(5)) {
      data3.position(data3.position() + 32);
    }

    if (!filled.get(6)) {
      data4.position(data4.position() + 32);
    }

    if (!filled.get(7)) {
      data5.position(data5.position() + 32);
    }

    if (!filled.get(8)) {
      data6.position(data6.position() + 32);
    }

    if (!filled.get(9)) {
      data7.position(data7.position() + 32);
    }

    if (!filled.get(10)) {
      data8.position(data8.position() + 32);
    }

    if (!filled.get(11)) {
      isBlake2FCds.position(isBlake2FCds.position() + 1);
    }

    if (!filled.get(12)) {
      isBlake2FParams.position(isBlake2FParams.position() + 1);
    }

    if (!filled.get(13)) {
      isCall.position(isCall.position() + 1);
    }

    if (!filled.get(14)) {
      isCdl.position(isCdl.position() + 1);
    }

    if (!filled.get(15)) {
      isCreate.position(isCreate.position() + 1);
    }

    if (!filled.get(16)) {
      isDeployment.position(isDeployment.position() + 1);
    }

    if (!filled.get(17)) {
      isEcadd.position(isEcadd.position() + 1);
    }

    if (!filled.get(18)) {
      isEcmul.position(isEcmul.position() + 1);
    }

    if (!filled.get(19)) {
      isEcpairing.position(isEcpairing.position() + 1);
    }

    if (!filled.get(20)) {
      isEcrecover.position(isEcrecover.position() + 1);
    }

    if (!filled.get(21)) {
      isIdentity.position(isIdentity.position() + 1);
    }

    if (!filled.get(22)) {
      isJump.position(isJump.position() + 1);
    }

    if (!filled.get(23)) {
      isJumpi.position(isJumpi.position() + 1);
    }

    if (!filled.get(24)) {
      isModexpCds.position(isModexpCds.position() + 1);
    }

    if (!filled.get(25)) {
      isModexpExtract.position(isModexpExtract.position() + 1);
    }

    if (!filled.get(26)) {
      isModexpLead.position(isModexpLead.position() + 1);
    }

    if (!filled.get(27)) {
      isModexpPricing.position(isModexpPricing.position() + 1);
    }

    if (!filled.get(28)) {
      isModexpXbs.position(isModexpXbs.position() + 1);
    }

    if (!filled.get(29)) {
      isRdc.position(isRdc.position() + 1);
    }

    if (!filled.get(30)) {
      isRipemd.position(isRipemd.position() + 1);
    }

    if (!filled.get(31)) {
      isSha2.position(isSha2.position() + 1);
    }

    if (!filled.get(32)) {
      isSstore.position(isSstore.position() + 1);
    }

    if (!filled.get(33)) {
      isXcall.position(isXcall.position() + 1);
    }

    if (!filled.get(34)) {
      modFlag.position(modFlag.position() + 1);
    }

    if (!filled.get(35)) {
      oobInst.position(oobInst.position() + 32);
    }

    if (!filled.get(36)) {
      outgoingData1.position(outgoingData1.position() + 32);
    }

    if (!filled.get(37)) {
      outgoingData2.position(outgoingData2.position() + 32);
    }

    if (!filled.get(38)) {
      outgoingData3.position(outgoingData3.position() + 32);
    }

    if (!filled.get(39)) {
      outgoingData4.position(outgoingData4.position() + 32);
    }

    if (!filled.get(40)) {
      outgoingInst.position(outgoingInst.position() + 1);
    }

    if (!filled.get(41)) {
      outgoingResLo.position(outgoingResLo.position() + 32);
    }

    if (!filled.get(42)) {
      stamp.position(stamp.position() + 8);
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
