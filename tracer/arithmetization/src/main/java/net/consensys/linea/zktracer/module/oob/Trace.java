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
  public static final int CT_MAX_BLAKE2F_CDS = 0x1;
  public static final int CT_MAX_BLAKE2F_PARAMS = 0x1;
  public static final int CT_MAX_CALL = 0x2;
  public static final int CT_MAX_CDL = 0x0;
  public static final int CT_MAX_CREATE = 0x3;
  public static final int CT_MAX_DEPLOYMENT = 0x0;
  public static final int CT_MAX_ECADD = 0x2;
  public static final int CT_MAX_ECMUL = 0x2;
  public static final int CT_MAX_ECPAIRING = 0x4;
  public static final int CT_MAX_ECRECOVER = 0x2;
  public static final int CT_MAX_IDENTITY = 0x3;
  public static final int CT_MAX_JUMP = 0x0;
  public static final int CT_MAX_JUMPI = 0x1;
  public static final int CT_MAX_MODEXP_CDS = 0x2;
  public static final int CT_MAX_MODEXP_EXTRACT = 0x3;
  public static final int CT_MAX_MODEXP_LEAD = 0x3;
  public static final int CT_MAX_MODEXP_PRICING = 0x5;
  public static final int CT_MAX_MODEXP_XBS = 0x2;
  public static final int CT_MAX_RDC = 0x2;
  public static final int CT_MAX_RIPEMD = 0x3;
  public static final int CT_MAX_SHA2 = 0x3;
  public static final int CT_MAX_SSTORE = 0x0;
  public static final int CT_MAX_XCALL = 0x0;
  public static final int G_QUADDIVISOR = 0x3;

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
  private final MappedByteBuffer data9;
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
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("oob.ADD_FLAG", 1, length));
      headers.add(new ColumnHeader("oob.CT", 1, length));
      headers.add(new ColumnHeader("oob.CT_MAX", 1, length));
      headers.add(new ColumnHeader("oob.DATA_1", 16, length));
      headers.add(new ColumnHeader("oob.DATA_2", 16, length));
      headers.add(new ColumnHeader("oob.DATA_3", 16, length));
      headers.add(new ColumnHeader("oob.DATA_4", 16, length));
      headers.add(new ColumnHeader("oob.DATA_5", 16, length));
      headers.add(new ColumnHeader("oob.DATA_6", 16, length));
      headers.add(new ColumnHeader("oob.DATA_7", 16, length));
      headers.add(new ColumnHeader("oob.DATA_8", 16, length));
      headers.add(new ColumnHeader("oob.DATA_9", 16, length));
      headers.add(new ColumnHeader("oob.IS_BLAKE2F_CDS", 1, length));
      headers.add(new ColumnHeader("oob.IS_BLAKE2F_PARAMS", 1, length));
      headers.add(new ColumnHeader("oob.IS_CALL", 1, length));
      headers.add(new ColumnHeader("oob.IS_CDL", 1, length));
      headers.add(new ColumnHeader("oob.IS_CREATE", 1, length));
      headers.add(new ColumnHeader("oob.IS_DEPLOYMENT", 1, length));
      headers.add(new ColumnHeader("oob.IS_ECADD", 1, length));
      headers.add(new ColumnHeader("oob.IS_ECMUL", 1, length));
      headers.add(new ColumnHeader("oob.IS_ECPAIRING", 1, length));
      headers.add(new ColumnHeader("oob.IS_ECRECOVER", 1, length));
      headers.add(new ColumnHeader("oob.IS_IDENTITY", 1, length));
      headers.add(new ColumnHeader("oob.IS_JUMP", 1, length));
      headers.add(new ColumnHeader("oob.IS_JUMPI", 1, length));
      headers.add(new ColumnHeader("oob.IS_MODEXP_CDS", 1, length));
      headers.add(new ColumnHeader("oob.IS_MODEXP_EXTRACT", 1, length));
      headers.add(new ColumnHeader("oob.IS_MODEXP_LEAD", 1, length));
      headers.add(new ColumnHeader("oob.IS_MODEXP_PRICING", 1, length));
      headers.add(new ColumnHeader("oob.IS_MODEXP_XBS", 1, length));
      headers.add(new ColumnHeader("oob.IS_RDC", 1, length));
      headers.add(new ColumnHeader("oob.IS_RIPEMD", 1, length));
      headers.add(new ColumnHeader("oob.IS_SHA2", 1, length));
      headers.add(new ColumnHeader("oob.IS_SSTORE", 1, length));
      headers.add(new ColumnHeader("oob.IS_XCALL", 1, length));
      headers.add(new ColumnHeader("oob.MOD_FLAG", 1, length));
      headers.add(new ColumnHeader("oob.OOB_INST", 2, length));
      headers.add(new ColumnHeader("oob.OUTGOING_DATA_1", 16, length));
      headers.add(new ColumnHeader("oob.OUTGOING_DATA_2", 16, length));
      headers.add(new ColumnHeader("oob.OUTGOING_DATA_3", 16, length));
      headers.add(new ColumnHeader("oob.OUTGOING_DATA_4", 16, length));
      headers.add(new ColumnHeader("oob.OUTGOING_INST", 1, length));
      headers.add(new ColumnHeader("oob.OUTGOING_RES_LO", 16, length));
      headers.add(new ColumnHeader("oob.STAMP", 4, length));
      headers.add(new ColumnHeader("oob.WCP_FLAG", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
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
    this.data9 = buffers.get(11);
    this.isBlake2FCds = buffers.get(12);
    this.isBlake2FParams = buffers.get(13);
    this.isCall = buffers.get(14);
    this.isCdl = buffers.get(15);
    this.isCreate = buffers.get(16);
    this.isDeployment = buffers.get(17);
    this.isEcadd = buffers.get(18);
    this.isEcmul = buffers.get(19);
    this.isEcpairing = buffers.get(20);
    this.isEcrecover = buffers.get(21);
    this.isIdentity = buffers.get(22);
    this.isJump = buffers.get(23);
    this.isJumpi = buffers.get(24);
    this.isModexpCds = buffers.get(25);
    this.isModexpExtract = buffers.get(26);
    this.isModexpLead = buffers.get(27);
    this.isModexpPricing = buffers.get(28);
    this.isModexpXbs = buffers.get(29);
    this.isRdc = buffers.get(30);
    this.isRipemd = buffers.get(31);
    this.isSha2 = buffers.get(32);
    this.isSstore = buffers.get(33);
    this.isXcall = buffers.get(34);
    this.modFlag = buffers.get(35);
    this.oobInst = buffers.get(36);
    this.outgoingData1 = buffers.get(37);
    this.outgoingData2 = buffers.get(38);
    this.outgoingData3 = buffers.get(39);
    this.outgoingData4 = buffers.get(40);
    this.outgoingInst = buffers.get(41);
    this.outgoingResLo = buffers.get(42);
    this.stamp = buffers.get(43);
    this.wcpFlag = buffers.get(44);
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

  public Trace ct(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("oob.CT already set");
    } else {
      filled.set(1);
    }

    if(b >= 8L) { throw new IllegalArgumentException("oob.CT has invalid value (" + b + ")"); }
    ct.put((byte) b);


    return this;
  }

  public Trace ctMax(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("oob.CT_MAX already set");
    } else {
      filled.set(2);
    }

    if(b >= 8L) { throw new IllegalArgumentException("oob.CT_MAX has invalid value (" + b + ")"); }
    ctMax.put((byte) b);


    return this;
  }

  public Trace data1(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("oob.DATA_1 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data1.put(bs.get(j)); }

    return this;
  }

  public Trace data2(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("oob.DATA_2 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data2.put(bs.get(j)); }

    return this;
  }

  public Trace data3(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("oob.DATA_3 already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data3.put(bs.get(j)); }

    return this;
  }

  public Trace data4(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("oob.DATA_4 already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data4.put(bs.get(j)); }

    return this;
  }

  public Trace data5(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("oob.DATA_5 already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data5.put(bs.get(j)); }

    return this;
  }

  public Trace data6(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("oob.DATA_6 already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data6.put(bs.get(j)); }

    return this;
  }

  public Trace data7(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("oob.DATA_7 already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_7 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data7.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data7.put(bs.get(j)); }

    return this;
  }

  public Trace data8(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("oob.DATA_8 already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_8 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data8.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data8.put(bs.get(j)); }

    return this;
  }

  public Trace data9(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("oob.DATA_9 already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.DATA_9 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data9.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data9.put(bs.get(j)); }

    return this;
  }

  public Trace isBlake2FCds(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_CDS already set");
    } else {
      filled.set(12);
    }

    isBlake2FCds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isBlake2FParams(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_PARAMS already set");
    } else {
      filled.set(13);
    }

    isBlake2FParams.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCall(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("oob.IS_CALL already set");
    } else {
      filled.set(14);
    }

    isCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCdl(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("oob.IS_CDL already set");
    } else {
      filled.set(15);
    }

    isCdl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCreate(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("oob.IS_CREATE already set");
    } else {
      filled.set(16);
    }

    isCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isDeployment(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("oob.IS_DEPLOYMENT already set");
    } else {
      filled.set(17);
    }

    isDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcadd(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("oob.IS_ECADD already set");
    } else {
      filled.set(18);
    }

    isEcadd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcmul(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("oob.IS_ECMUL already set");
    } else {
      filled.set(19);
    }

    isEcmul.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcpairing(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("oob.IS_ECPAIRING already set");
    } else {
      filled.set(20);
    }

    isEcpairing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isEcrecover(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("oob.IS_ECRECOVER already set");
    } else {
      filled.set(21);
    }

    isEcrecover.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isIdentity(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("oob.IS_IDENTITY already set");
    } else {
      filled.set(22);
    }

    isIdentity.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJump(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("oob.IS_JUMP already set");
    } else {
      filled.set(23);
    }

    isJump.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJumpi(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("oob.IS_JUMPI already set");
    } else {
      filled.set(24);
    }

    isJumpi.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpCds(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("oob.IS_MODEXP_CDS already set");
    } else {
      filled.set(25);
    }

    isModexpCds.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpExtract(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("oob.IS_MODEXP_EXTRACT already set");
    } else {
      filled.set(26);
    }

    isModexpExtract.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpLead(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("oob.IS_MODEXP_LEAD already set");
    } else {
      filled.set(27);
    }

    isModexpLead.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpPricing(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("oob.IS_MODEXP_PRICING already set");
    } else {
      filled.set(28);
    }

    isModexpPricing.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpXbs(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("oob.IS_MODEXP_XBS already set");
    } else {
      filled.set(29);
    }

    isModexpXbs.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRdc(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("oob.IS_RDC already set");
    } else {
      filled.set(30);
    }

    isRdc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRipemd(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("oob.IS_RIPEMD already set");
    } else {
      filled.set(31);
    }

    isRipemd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSha2(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("oob.IS_SHA2 already set");
    } else {
      filled.set(32);
    }

    isSha2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSstore(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("oob.IS_SSTORE already set");
    } else {
      filled.set(33);
    }

    isSstore.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isXcall(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("oob.IS_XCALL already set");
    } else {
      filled.set(34);
    }

    isXcall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace modFlag(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("oob.MOD_FLAG already set");
    } else {
      filled.set(35);
    }

    modFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oobInst(final long b) {
    if (filled.get(36)) {
      throw new IllegalStateException("oob.OOB_INST already set");
    } else {
      filled.set(36);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("oob.OOB_INST has invalid value (" + b + ")"); }
    oobInst.put((byte) (b >> 8));
    oobInst.put((byte) b);


    return this;
  }

  public Trace outgoingData1(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_1 already set");
    } else {
      filled.set(37);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.OUTGOING_DATA_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingData1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingData1.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingData2(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_2 already set");
    } else {
      filled.set(38);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.OUTGOING_DATA_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingData2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingData2.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingData3(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_3 already set");
    } else {
      filled.set(39);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.OUTGOING_DATA_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingData3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingData3.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingData4(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_4 already set");
    } else {
      filled.set(40);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.OUTGOING_DATA_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingData4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingData4.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingInst(final UnsignedByte b) {
    if (filled.get(41)) {
      throw new IllegalStateException("oob.OUTGOING_INST already set");
    } else {
      filled.set(41);
    }

    outgoingInst.put(b.toByte());

    return this;
  }

  public Trace outgoingResLo(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("oob.OUTGOING_RES_LO already set");
    } else {
      filled.set(42);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("oob.OUTGOING_RES_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingResLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingResLo.put(bs.get(j)); }

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(43)) {
      throw new IllegalStateException("oob.STAMP already set");
    } else {
      filled.set(43);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("oob.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 24));
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace wcpFlag(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("oob.WCP_FLAG already set");
    } else {
      filled.set(44);
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
      throw new IllegalStateException("oob.DATA_9 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_CDS has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("oob.IS_BLAKE2F_PARAMS has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("oob.IS_CALL has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("oob.IS_CDL has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("oob.IS_CREATE has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("oob.IS_DEPLOYMENT has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("oob.IS_ECADD has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("oob.IS_ECMUL has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("oob.IS_ECPAIRING has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("oob.IS_ECRECOVER has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("oob.IS_IDENTITY has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("oob.IS_JUMP has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("oob.IS_JUMPI has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("oob.IS_MODEXP_CDS has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("oob.IS_MODEXP_EXTRACT has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("oob.IS_MODEXP_LEAD has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("oob.IS_MODEXP_PRICING has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("oob.IS_MODEXP_XBS has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("oob.IS_RDC has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("oob.IS_RIPEMD has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("oob.IS_SHA2 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("oob.IS_SSTORE has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("oob.IS_XCALL has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("oob.MOD_FLAG has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("oob.OOB_INST has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_1 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_2 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_3 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("oob.OUTGOING_DATA_4 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("oob.OUTGOING_INST has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("oob.OUTGOING_RES_LO has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("oob.STAMP has not been filled");
    }

    if (!filled.get(44)) {
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
      ct.position(ct.position() + 1);
    }

    if (!filled.get(2)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(3)) {
      data1.position(data1.position() + 16);
    }

    if (!filled.get(4)) {
      data2.position(data2.position() + 16);
    }

    if (!filled.get(5)) {
      data3.position(data3.position() + 16);
    }

    if (!filled.get(6)) {
      data4.position(data4.position() + 16);
    }

    if (!filled.get(7)) {
      data5.position(data5.position() + 16);
    }

    if (!filled.get(8)) {
      data6.position(data6.position() + 16);
    }

    if (!filled.get(9)) {
      data7.position(data7.position() + 16);
    }

    if (!filled.get(10)) {
      data8.position(data8.position() + 16);
    }

    if (!filled.get(11)) {
      data9.position(data9.position() + 16);
    }

    if (!filled.get(12)) {
      isBlake2FCds.position(isBlake2FCds.position() + 1);
    }

    if (!filled.get(13)) {
      isBlake2FParams.position(isBlake2FParams.position() + 1);
    }

    if (!filled.get(14)) {
      isCall.position(isCall.position() + 1);
    }

    if (!filled.get(15)) {
      isCdl.position(isCdl.position() + 1);
    }

    if (!filled.get(16)) {
      isCreate.position(isCreate.position() + 1);
    }

    if (!filled.get(17)) {
      isDeployment.position(isDeployment.position() + 1);
    }

    if (!filled.get(18)) {
      isEcadd.position(isEcadd.position() + 1);
    }

    if (!filled.get(19)) {
      isEcmul.position(isEcmul.position() + 1);
    }

    if (!filled.get(20)) {
      isEcpairing.position(isEcpairing.position() + 1);
    }

    if (!filled.get(21)) {
      isEcrecover.position(isEcrecover.position() + 1);
    }

    if (!filled.get(22)) {
      isIdentity.position(isIdentity.position() + 1);
    }

    if (!filled.get(23)) {
      isJump.position(isJump.position() + 1);
    }

    if (!filled.get(24)) {
      isJumpi.position(isJumpi.position() + 1);
    }

    if (!filled.get(25)) {
      isModexpCds.position(isModexpCds.position() + 1);
    }

    if (!filled.get(26)) {
      isModexpExtract.position(isModexpExtract.position() + 1);
    }

    if (!filled.get(27)) {
      isModexpLead.position(isModexpLead.position() + 1);
    }

    if (!filled.get(28)) {
      isModexpPricing.position(isModexpPricing.position() + 1);
    }

    if (!filled.get(29)) {
      isModexpXbs.position(isModexpXbs.position() + 1);
    }

    if (!filled.get(30)) {
      isRdc.position(isRdc.position() + 1);
    }

    if (!filled.get(31)) {
      isRipemd.position(isRipemd.position() + 1);
    }

    if (!filled.get(32)) {
      isSha2.position(isSha2.position() + 1);
    }

    if (!filled.get(33)) {
      isSstore.position(isSstore.position() + 1);
    }

    if (!filled.get(34)) {
      isXcall.position(isXcall.position() + 1);
    }

    if (!filled.get(35)) {
      modFlag.position(modFlag.position() + 1);
    }

    if (!filled.get(36)) {
      oobInst.position(oobInst.position() + 2);
    }

    if (!filled.get(37)) {
      outgoingData1.position(outgoingData1.position() + 16);
    }

    if (!filled.get(38)) {
      outgoingData2.position(outgoingData2.position() + 16);
    }

    if (!filled.get(39)) {
      outgoingData3.position(outgoingData3.position() + 16);
    }

    if (!filled.get(40)) {
      outgoingData4.position(outgoingData4.position() + 16);
    }

    if (!filled.get(41)) {
      outgoingInst.position(outgoingInst.position() + 1);
    }

    if (!filled.get(42)) {
      outgoingResLo.position(outgoingResLo.position() + 16);
    }

    if (!filled.get(43)) {
      stamp.position(stamp.position() + 4);
    }

    if (!filled.get(44)) {
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
