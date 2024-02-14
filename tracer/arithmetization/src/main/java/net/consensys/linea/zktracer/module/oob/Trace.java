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
  static final int ADD = 1;
  static final int CT_MAX_CALL = 2;
  static final int CT_MAX_CDL = 0;
  static final int CT_MAX_CREATE = 2;
  static final int CT_MAX_JUMP = 0;
  static final int CT_MAX_JUMPI = 1;
  static final int CT_MAX_PRC_BLAKE2F_a = 0;
  static final int CT_MAX_PRC_BLAKE2F_b = 2;
  static final int CT_MAX_PRC_ECADD = 2;
  static final int CT_MAX_PRC_ECMUL = 2;
  static final int CT_MAX_PRC_ECPAIRING = 4;
  static final int CT_MAX_PRC_ECRECOVER = 2;
  static final int CT_MAX_PRC_IDENTITY = 3;
  static final int CT_MAX_PRC_MODEXP_BASE = 3;
  static final int CT_MAX_PRC_MODEXP_CDS = 3;
  static final int CT_MAX_PRC_MODEXP_EXPONENT = 2;
  static final int CT_MAX_PRC_MODEXP_MODULUS = 2;
  static final int CT_MAX_PRC_MODEXP_PRICING = 5;
  static final int CT_MAX_PRC_RIPEMD = 3;
  static final int CT_MAX_PRC_SHA2 = 3;
  static final int CT_MAX_RDC = 2;
  static final int CT_MAX_RETURN = 0;
  static final int CT_MAX_SSTORE = 0;
  static final int CT_MAX_XCALL = 0;
  static final int DIV = 4;
  static final int EQ = 20;
  static final int GT = 17;
  public static final int G_CALLSTIPEND = 2300;
  static final int G_QUADDIVISOR = 3;
  static final int ISZERO = 21;
  static final int LT = 16;
  static final int MOD = 6;
  public static final int OOB_INST_blake2f_cds = 64009;
  public static final int OOB_INST_blake2f_params = 64265;
  public static final int OOB_INST_call = 202;
  public static final int OOB_INST_cdl = 53;
  public static final int OOB_INST_create = 206;
  public static final int OOB_INST_ecadd = 65286;
  public static final int OOB_INST_ecmul = 65287;
  public static final int OOB_INST_ecpairing = 65288;
  public static final int OOB_INST_ecrecover = 65281;
  public static final int OOB_INST_identity = 65284;
  public static final int OOB_INST_jump = 86;
  public static final int OOB_INST_jumpi = 87;
  public static final int OOB_INST_modexp_cds = 64005;
  public static final int OOB_INST_modexp_extract = 65029;
  public static final int OOB_INST_modexp_lead = 64517;
  public static final int OOB_INST_modexp_pricing = 64773;
  public static final int OOB_INST_modexp_xbs = 64261;
  public static final int OOB_INST_rdc = 62;
  public static final int OOB_INST_return = 243;
  public static final int OOB_INST_ripemd = 65283;
  public static final int OOB_INST_sha2 = 65282;
  public static final int OOB_INST_sstore = 85;
  public static final int OOB_INST_xcall = 204;

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
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isCall;
  private final MappedByteBuffer isCdl;
  private final MappedByteBuffer isCreate;
  private final MappedByteBuffer isJump;
  private final MappedByteBuffer isJumpi;
  private final MappedByteBuffer isRdc;
  private final MappedByteBuffer isReturn;
  private final MappedByteBuffer isSstore;
  private final MappedByteBuffer isXcall;
  private final MappedByteBuffer modFlag;
  private final MappedByteBuffer outgoingData1;
  private final MappedByteBuffer outgoingData2;
  private final MappedByteBuffer outgoingData3;
  private final MappedByteBuffer outgoingData4;
  private final MappedByteBuffer outgoingInst;
  private final MappedByteBuffer outgoingResLo;
  private final MappedByteBuffer prcBlake2FCds;
  private final MappedByteBuffer prcBlake2FParams;
  private final MappedByteBuffer prcEcadd;
  private final MappedByteBuffer prcEcmul;
  private final MappedByteBuffer prcEcpairing;
  private final MappedByteBuffer prcEcrecover;
  private final MappedByteBuffer prcIdentity;
  private final MappedByteBuffer prcModexpBase;
  private final MappedByteBuffer prcModexpCds;
  private final MappedByteBuffer prcModexpExponent;
  private final MappedByteBuffer prcModexpModulus;
  private final MappedByteBuffer prcModexpPricing;
  private final MappedByteBuffer prcRipemd;
  private final MappedByteBuffer prcSha2;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("oob.ADD_FLAG", 1, length),
        new ColumnHeader("oob.CT", 32, length),
        new ColumnHeader("oob.CT_MAX", 32, length),
        new ColumnHeader("oob.DATA_1", 32, length),
        new ColumnHeader("oob.DATA_2", 32, length),
        new ColumnHeader("oob.DATA_3", 32, length),
        new ColumnHeader("oob.DATA_4", 32, length),
        new ColumnHeader("oob.DATA_5", 32, length),
        new ColumnHeader("oob.DATA_6", 32, length),
        new ColumnHeader("oob.DATA_7", 32, length),
        new ColumnHeader("oob.DATA_8", 32, length),
        new ColumnHeader("oob.INST", 32, length),
        new ColumnHeader("oob.IS_CALL", 1, length),
        new ColumnHeader("oob.IS_CDL", 1, length),
        new ColumnHeader("oob.IS_CREATE", 1, length),
        new ColumnHeader("oob.IS_JUMP", 1, length),
        new ColumnHeader("oob.IS_JUMPI", 1, length),
        new ColumnHeader("oob.IS_RDC", 1, length),
        new ColumnHeader("oob.IS_RETURN", 1, length),
        new ColumnHeader("oob.IS_SSTORE", 1, length),
        new ColumnHeader("oob.IS_XCALL", 1, length),
        new ColumnHeader("oob.MOD_FLAG", 1, length),
        new ColumnHeader("oob.OUTGOING_DATA_1", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_2", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_3", 32, length),
        new ColumnHeader("oob.OUTGOING_DATA_4", 32, length),
        new ColumnHeader("oob.OUTGOING_INST", 1, length),
        new ColumnHeader("oob.OUTGOING_RES_LO", 32, length),
        new ColumnHeader("oob.PRC_BLAKE2F_cds", 1, length),
        new ColumnHeader("oob.PRC_BLAKE2F_params", 1, length),
        new ColumnHeader("oob.PRC_ECADD", 1, length),
        new ColumnHeader("oob.PRC_ECMUL", 1, length),
        new ColumnHeader("oob.PRC_ECPAIRING", 1, length),
        new ColumnHeader("oob.PRC_ECRECOVER", 1, length),
        new ColumnHeader("oob.PRC_IDENTITY", 1, length),
        new ColumnHeader("oob.PRC_MODEXP_base", 1, length),
        new ColumnHeader("oob.PRC_MODEXP_cds", 1, length),
        new ColumnHeader("oob.PRC_MODEXP_exponent", 1, length),
        new ColumnHeader("oob.PRC_MODEXP_modulus", 1, length),
        new ColumnHeader("oob.PRC_MODEXP_pricing", 1, length),
        new ColumnHeader("oob.PRC_RIPEMD", 1, length),
        new ColumnHeader("oob.PRC_SHA2", 1, length),
        new ColumnHeader("oob.STAMP", 32, length),
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
    this.inst = buffers.get(11);
    this.isCall = buffers.get(12);
    this.isCdl = buffers.get(13);
    this.isCreate = buffers.get(14);
    this.isJump = buffers.get(15);
    this.isJumpi = buffers.get(16);
    this.isRdc = buffers.get(17);
    this.isReturn = buffers.get(18);
    this.isSstore = buffers.get(19);
    this.isXcall = buffers.get(20);
    this.modFlag = buffers.get(21);
    this.outgoingData1 = buffers.get(22);
    this.outgoingData2 = buffers.get(23);
    this.outgoingData3 = buffers.get(24);
    this.outgoingData4 = buffers.get(25);
    this.outgoingInst = buffers.get(26);
    this.outgoingResLo = buffers.get(27);
    this.prcBlake2FCds = buffers.get(28);
    this.prcBlake2FParams = buffers.get(29);
    this.prcEcadd = buffers.get(30);
    this.prcEcmul = buffers.get(31);
    this.prcEcpairing = buffers.get(32);
    this.prcEcrecover = buffers.get(33);
    this.prcIdentity = buffers.get(34);
    this.prcModexpBase = buffers.get(35);
    this.prcModexpCds = buffers.get(36);
    this.prcModexpExponent = buffers.get(37);
    this.prcModexpModulus = buffers.get(38);
    this.prcModexpPricing = buffers.get(39);
    this.prcRipemd = buffers.get(40);
    this.prcSha2 = buffers.get(41);
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

  public Trace ct(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("oob.CT already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      ct.put((byte) 0);
    }
    ct.put(b.toArrayUnsafe());

    return this;
  }

  public Trace ctMax(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("oob.CT_MAX already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      ctMax.put((byte) 0);
    }
    ctMax.put(b.toArrayUnsafe());

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

  public Trace inst(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("oob.INST already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      inst.put((byte) 0);
    }
    inst.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isCall(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("oob.IS_CALL already set");
    } else {
      filled.set(12);
    }

    isCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCdl(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("oob.IS_CDL already set");
    } else {
      filled.set(13);
    }

    isCdl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCreate(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("oob.IS_CREATE already set");
    } else {
      filled.set(14);
    }

    isCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJump(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("oob.IS_JUMP already set");
    } else {
      filled.set(15);
    }

    isJump.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJumpi(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("oob.IS_JUMPI already set");
    } else {
      filled.set(16);
    }

    isJumpi.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRdc(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("oob.IS_RDC already set");
    } else {
      filled.set(17);
    }

    isRdc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isReturn(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("oob.IS_RETURN already set");
    } else {
      filled.set(18);
    }

    isReturn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSstore(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("oob.IS_SSTORE already set");
    } else {
      filled.set(19);
    }

    isSstore.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isXcall(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("oob.IS_XCALL already set");
    } else {
      filled.set(20);
    }

    isXcall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace modFlag(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("oob.MOD_FLAG already set");
    } else {
      filled.set(21);
    }

    modFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace outgoingData1(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_1 already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData1.put((byte) 0);
    }
    outgoingData1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData2(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_2 already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData2.put((byte) 0);
    }
    outgoingData2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData3(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_3 already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData3.put((byte) 0);
    }
    outgoingData3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingData4(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_4 already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingData4.put((byte) 0);
    }
    outgoingData4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace outgoingInst(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("oob.OUTGOING_INST already set");
    } else {
      filled.set(26);
    }

    outgoingInst.put(b.toByte());

    return this;
  }

  public Trace outgoingResLo(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("oob.OUTGOING_RES_LO already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      outgoingResLo.put((byte) 0);
    }
    outgoingResLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace prcBlake2FCds(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("oob.PRC_BLAKE2F_cds already set");
    } else {
      filled.set(28);
    }

    prcBlake2FCds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcBlake2FParams(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("oob.PRC_BLAKE2F_params already set");
    } else {
      filled.set(29);
    }

    prcBlake2FParams.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcEcadd(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("oob.PRC_ECADD already set");
    } else {
      filled.set(30);
    }

    prcEcadd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcEcmul(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("oob.PRC_ECMUL already set");
    } else {
      filled.set(31);
    }

    prcEcmul.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcEcpairing(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("oob.PRC_ECPAIRING already set");
    } else {
      filled.set(32);
    }

    prcEcpairing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcEcrecover(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("oob.PRC_ECRECOVER already set");
    } else {
      filled.set(33);
    }

    prcEcrecover.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcIdentity(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("oob.PRC_IDENTITY already set");
    } else {
      filled.set(34);
    }

    prcIdentity.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcModexpBase(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("oob.PRC_MODEXP_base already set");
    } else {
      filled.set(35);
    }

    prcModexpBase.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcModexpCds(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("oob.PRC_MODEXP_cds already set");
    } else {
      filled.set(36);
    }

    prcModexpCds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcModexpExponent(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("oob.PRC_MODEXP_exponent already set");
    } else {
      filled.set(37);
    }

    prcModexpExponent.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcModexpModulus(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("oob.PRC_MODEXP_modulus already set");
    } else {
      filled.set(38);
    }

    prcModexpModulus.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcModexpPricing(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("oob.PRC_MODEXP_pricing already set");
    } else {
      filled.set(39);
    }

    prcModexpPricing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcRipemd(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("oob.PRC_RIPEMD already set");
    } else {
      filled.set(40);
    }

    prcRipemd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prcSha2(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("oob.PRC_SHA2 already set");
    } else {
      filled.set(41);
    }

    prcSha2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("oob.STAMP already set");
    } else {
      filled.set(42);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

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
      throw new IllegalStateException("oob.INST has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("oob.IS_CALL has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("oob.IS_CDL has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("oob.IS_CREATE has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("oob.IS_JUMP has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("oob.IS_JUMPI has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("oob.IS_RDC has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("oob.IS_RETURN has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("oob.IS_SSTORE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("oob.IS_XCALL has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("oob.MOD_FLAG has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_1 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_2 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_3 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_4 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("oob.OUTGOING_INST has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("oob.OUTGOING_RES_LO has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("oob.PRC_BLAKE2F_cds has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("oob.PRC_BLAKE2F_params has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("oob.PRC_ECADD has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("oob.PRC_ECMUL has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("oob.PRC_ECPAIRING has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("oob.PRC_ECRECOVER has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("oob.PRC_IDENTITY has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("oob.PRC_MODEXP_base has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("oob.PRC_MODEXP_cds has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("oob.PRC_MODEXP_exponent has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("oob.PRC_MODEXP_modulus has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("oob.PRC_MODEXP_pricing has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("oob.PRC_RIPEMD has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("oob.PRC_SHA2 has not been filled");
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
      ct.position(ct.position() + 32);
    }

    if (!filled.get(2)) {
      ctMax.position(ctMax.position() + 32);
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
      inst.position(inst.position() + 32);
    }

    if (!filled.get(12)) {
      isCall.position(isCall.position() + 1);
    }

    if (!filled.get(13)) {
      isCdl.position(isCdl.position() + 1);
    }

    if (!filled.get(14)) {
      isCreate.position(isCreate.position() + 1);
    }

    if (!filled.get(15)) {
      isJump.position(isJump.position() + 1);
    }

    if (!filled.get(16)) {
      isJumpi.position(isJumpi.position() + 1);
    }

    if (!filled.get(17)) {
      isRdc.position(isRdc.position() + 1);
    }

    if (!filled.get(18)) {
      isReturn.position(isReturn.position() + 1);
    }

    if (!filled.get(19)) {
      isSstore.position(isSstore.position() + 1);
    }

    if (!filled.get(20)) {
      isXcall.position(isXcall.position() + 1);
    }

    if (!filled.get(21)) {
      modFlag.position(modFlag.position() + 1);
    }

    if (!filled.get(22)) {
      outgoingData1.position(outgoingData1.position() + 32);
    }

    if (!filled.get(23)) {
      outgoingData2.position(outgoingData2.position() + 32);
    }

    if (!filled.get(24)) {
      outgoingData3.position(outgoingData3.position() + 32);
    }

    if (!filled.get(25)) {
      outgoingData4.position(outgoingData4.position() + 32);
    }

    if (!filled.get(26)) {
      outgoingInst.position(outgoingInst.position() + 1);
    }

    if (!filled.get(27)) {
      outgoingResLo.position(outgoingResLo.position() + 32);
    }

    if (!filled.get(28)) {
      prcBlake2FCds.position(prcBlake2FCds.position() + 1);
    }

    if (!filled.get(29)) {
      prcBlake2FParams.position(prcBlake2FParams.position() + 1);
    }

    if (!filled.get(30)) {
      prcEcadd.position(prcEcadd.position() + 1);
    }

    if (!filled.get(31)) {
      prcEcmul.position(prcEcmul.position() + 1);
    }

    if (!filled.get(32)) {
      prcEcpairing.position(prcEcpairing.position() + 1);
    }

    if (!filled.get(33)) {
      prcEcrecover.position(prcEcrecover.position() + 1);
    }

    if (!filled.get(34)) {
      prcIdentity.position(prcIdentity.position() + 1);
    }

    if (!filled.get(35)) {
      prcModexpBase.position(prcModexpBase.position() + 1);
    }

    if (!filled.get(36)) {
      prcModexpCds.position(prcModexpCds.position() + 1);
    }

    if (!filled.get(37)) {
      prcModexpExponent.position(prcModexpExponent.position() + 1);
    }

    if (!filled.get(38)) {
      prcModexpModulus.position(prcModexpModulus.position() + 1);
    }

    if (!filled.get(39)) {
      prcModexpPricing.position(prcModexpPricing.position() + 1);
    }

    if (!filled.get(40)) {
      prcRipemd.position(prcRipemd.position() + 1);
    }

    if (!filled.get(41)) {
      prcSha2.position(prcSha2.position() + 1);
    }

    if (!filled.get(42)) {
      stamp.position(stamp.position() + 32);
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
