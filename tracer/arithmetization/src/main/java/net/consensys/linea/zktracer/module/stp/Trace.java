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

package net.consensys.linea.zktracer.module.stp;

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

  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer exists;
  private final MappedByteBuffer exogenousModuleInstruction;
  private final MappedByteBuffer gasActual;
  private final MappedByteBuffer gasHi;
  private final MappedByteBuffer gasLo;
  private final MappedByteBuffer gasMxp;
  private final MappedByteBuffer gasOutOfPocket;
  private final MappedByteBuffer gasStipend;
  private final MappedByteBuffer gasUpfront;
  private final MappedByteBuffer instruction;
  private final MappedByteBuffer isCall;
  private final MappedByteBuffer isCallcode;
  private final MappedByteBuffer isCreate;
  private final MappedByteBuffer isCreate2;
  private final MappedByteBuffer isDelegatecall;
  private final MappedByteBuffer isStaticcall;
  private final MappedByteBuffer modFlag;
  private final MappedByteBuffer outOfGasException;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer valHi;
  private final MappedByteBuffer valLo;
  private final MappedByteBuffer warm;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("stp.ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("stp.ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("stp.ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("stp.CT", 1, length));
      headers.add(new ColumnHeader("stp.CT_MAX", 1, length));
      headers.add(new ColumnHeader("stp.EXISTS", 1, length));
      headers.add(new ColumnHeader("stp.EXOGENOUS_MODULE_INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("stp.GAS_ACTUAL", 8, length));
      headers.add(new ColumnHeader("stp.GAS_HI", 16, length));
      headers.add(new ColumnHeader("stp.GAS_LO", 16, length));
      headers.add(new ColumnHeader("stp.GAS_MXP", 8, length));
      headers.add(new ColumnHeader("stp.GAS_OUT_OF_POCKET", 8, length));
      headers.add(new ColumnHeader("stp.GAS_STIPEND", 8, length));
      headers.add(new ColumnHeader("stp.GAS_UPFRONT", 8, length));
      headers.add(new ColumnHeader("stp.INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("stp.IS_CALL", 1, length));
      headers.add(new ColumnHeader("stp.IS_CALLCODE", 1, length));
      headers.add(new ColumnHeader("stp.IS_CREATE", 1, length));
      headers.add(new ColumnHeader("stp.IS_CREATE2", 1, length));
      headers.add(new ColumnHeader("stp.IS_DELEGATECALL", 1, length));
      headers.add(new ColumnHeader("stp.IS_STATICCALL", 1, length));
      headers.add(new ColumnHeader("stp.MOD_FLAG", 1, length));
      headers.add(new ColumnHeader("stp.OUT_OF_GAS_EXCEPTION", 1, length));
      headers.add(new ColumnHeader("stp.RES_LO", 16, length));
      headers.add(new ColumnHeader("stp.STAMP", 3, length));
      headers.add(new ColumnHeader("stp.VAL_HI", 16, length));
      headers.add(new ColumnHeader("stp.VAL_LO", 16, length));
      headers.add(new ColumnHeader("stp.WARM", 1, length));
      headers.add(new ColumnHeader("stp.WCP_FLAG", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.arg1Hi = buffers.get(0);
    this.arg1Lo = buffers.get(1);
    this.arg2Lo = buffers.get(2);
    this.ct = buffers.get(3);
    this.ctMax = buffers.get(4);
    this.exists = buffers.get(5);
    this.exogenousModuleInstruction = buffers.get(6);
    this.gasActual = buffers.get(7);
    this.gasHi = buffers.get(8);
    this.gasLo = buffers.get(9);
    this.gasMxp = buffers.get(10);
    this.gasOutOfPocket = buffers.get(11);
    this.gasStipend = buffers.get(12);
    this.gasUpfront = buffers.get(13);
    this.instruction = buffers.get(14);
    this.isCall = buffers.get(15);
    this.isCallcode = buffers.get(16);
    this.isCreate = buffers.get(17);
    this.isCreate2 = buffers.get(18);
    this.isDelegatecall = buffers.get(19);
    this.isStaticcall = buffers.get(20);
    this.modFlag = buffers.get(21);
    this.outOfGasException = buffers.get(22);
    this.resLo = buffers.get(23);
    this.stamp = buffers.get(24);
    this.valHi = buffers.get(25);
    this.valLo = buffers.get(26);
    this.warm = buffers.get(27);
    this.wcpFlag = buffers.get(28);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("stp.ARG_1_HI already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("stp.ARG_1_LO already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("stp.ARG_2_LO already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { arg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { arg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace ct(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("stp.CT already set");
    } else {
      filled.set(3);
    }

    ct.put(b.toByte());

    return this;
  }

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(4)) {
      throw new IllegalStateException("stp.CT_MAX already set");
    } else {
      filled.set(4);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace exists(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("stp.EXISTS already set");
    } else {
      filled.set(5);
    }

    exists.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exogenousModuleInstruction(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("stp.EXOGENOUS_MODULE_INSTRUCTION already set");
    } else {
      filled.set(6);
    }

    exogenousModuleInstruction.put(b.toByte());

    return this;
  }

  public Trace gasActual(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("stp.GAS_ACTUAL already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("stp.GAS_ACTUAL has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasActual.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasActual.put(bs.get(j)); }

    return this;
  }

  public Trace gasHi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("stp.GAS_HI already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.GAS_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { gasHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasHi.put(bs.get(j)); }

    return this;
  }

  public Trace gasLo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("stp.GAS_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.GAS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { gasLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasLo.put(bs.get(j)); }

    return this;
  }

  public Trace gasMxp(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("stp.GAS_MXP already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("stp.GAS_MXP has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasMxp.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasMxp.put(bs.get(j)); }

    return this;
  }

  public Trace gasOutOfPocket(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("stp.GAS_OUT_OF_POCKET already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("stp.GAS_OUT_OF_POCKET has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasOutOfPocket.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasOutOfPocket.put(bs.get(j)); }

    return this;
  }

  public Trace gasStipend(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("stp.GAS_STIPEND already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("stp.GAS_STIPEND has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasStipend.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasStipend.put(bs.get(j)); }

    return this;
  }

  public Trace gasUpfront(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("stp.GAS_UPFRONT already set");
    } else {
      filled.set(13);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("stp.GAS_UPFRONT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasUpfront.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasUpfront.put(bs.get(j)); }

    return this;
  }

  public Trace instruction(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("stp.INSTRUCTION already set");
    } else {
      filled.set(14);
    }

    instruction.put(b.toByte());

    return this;
  }

  public Trace isCall(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("stp.IS_CALL already set");
    } else {
      filled.set(15);
    }

    isCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCallcode(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("stp.IS_CALLCODE already set");
    } else {
      filled.set(16);
    }

    isCallcode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCreate(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("stp.IS_CREATE already set");
    } else {
      filled.set(17);
    }

    isCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isCreate2(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("stp.IS_CREATE2 already set");
    } else {
      filled.set(18);
    }

    isCreate2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isDelegatecall(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("stp.IS_DELEGATECALL already set");
    } else {
      filled.set(19);
    }

    isDelegatecall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isStaticcall(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("stp.IS_STATICCALL already set");
    } else {
      filled.set(20);
    }

    isStaticcall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace modFlag(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("stp.MOD_FLAG already set");
    } else {
      filled.set(21);
    }

    modFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace outOfGasException(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("stp.OUT_OF_GAS_EXCEPTION already set");
    } else {
      filled.set(22);
    }

    outOfGasException.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("stp.RES_LO already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.RES_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { resLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { resLo.put(bs.get(j)); }

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(24)) {
      throw new IllegalStateException("stp.STAMP already set");
    } else {
      filled.set(24);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("stp.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace valHi(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("stp.VAL_HI already set");
    } else {
      filled.set(25);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.VAL_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valHi.put(bs.get(j)); }

    return this;
  }

  public Trace valLo(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("stp.VAL_LO already set");
    } else {
      filled.set(26);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("stp.VAL_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valLo.put(bs.get(j)); }

    return this;
  }

  public Trace warm(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("stp.WARM already set");
    } else {
      filled.set(27);
    }

    warm.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace wcpFlag(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("stp.WCP_FLAG already set");
    } else {
      filled.set(28);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("stp.ARG_1_HI has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("stp.ARG_1_LO has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("stp.ARG_2_LO has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("stp.CT has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("stp.CT_MAX has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("stp.EXISTS has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("stp.EXOGENOUS_MODULE_INSTRUCTION has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("stp.GAS_ACTUAL has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("stp.GAS_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("stp.GAS_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("stp.GAS_MXP has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("stp.GAS_OUT_OF_POCKET has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("stp.GAS_STIPEND has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("stp.GAS_UPFRONT has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("stp.INSTRUCTION has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("stp.IS_CALL has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("stp.IS_CALLCODE has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("stp.IS_CREATE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("stp.IS_CREATE2 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("stp.IS_DELEGATECALL has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("stp.IS_STATICCALL has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("stp.MOD_FLAG has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("stp.OUT_OF_GAS_EXCEPTION has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("stp.RES_LO has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("stp.STAMP has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("stp.VAL_HI has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("stp.VAL_LO has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("stp.WARM has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("stp.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      arg1Hi.position(arg1Hi.position() + 16);
    }

    if (!filled.get(1)) {
      arg1Lo.position(arg1Lo.position() + 16);
    }

    if (!filled.get(2)) {
      arg2Lo.position(arg2Lo.position() + 16);
    }

    if (!filled.get(3)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(4)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(5)) {
      exists.position(exists.position() + 1);
    }

    if (!filled.get(6)) {
      exogenousModuleInstruction.position(exogenousModuleInstruction.position() + 1);
    }

    if (!filled.get(7)) {
      gasActual.position(gasActual.position() + 8);
    }

    if (!filled.get(8)) {
      gasHi.position(gasHi.position() + 16);
    }

    if (!filled.get(9)) {
      gasLo.position(gasLo.position() + 16);
    }

    if (!filled.get(10)) {
      gasMxp.position(gasMxp.position() + 8);
    }

    if (!filled.get(11)) {
      gasOutOfPocket.position(gasOutOfPocket.position() + 8);
    }

    if (!filled.get(12)) {
      gasStipend.position(gasStipend.position() + 8);
    }

    if (!filled.get(13)) {
      gasUpfront.position(gasUpfront.position() + 8);
    }

    if (!filled.get(14)) {
      instruction.position(instruction.position() + 1);
    }

    if (!filled.get(15)) {
      isCall.position(isCall.position() + 1);
    }

    if (!filled.get(16)) {
      isCallcode.position(isCallcode.position() + 1);
    }

    if (!filled.get(17)) {
      isCreate.position(isCreate.position() + 1);
    }

    if (!filled.get(18)) {
      isCreate2.position(isCreate2.position() + 1);
    }

    if (!filled.get(19)) {
      isDelegatecall.position(isDelegatecall.position() + 1);
    }

    if (!filled.get(20)) {
      isStaticcall.position(isStaticcall.position() + 1);
    }

    if (!filled.get(21)) {
      modFlag.position(modFlag.position() + 1);
    }

    if (!filled.get(22)) {
      outOfGasException.position(outOfGasException.position() + 1);
    }

    if (!filled.get(23)) {
      resLo.position(resLo.position() + 16);
    }

    if (!filled.get(24)) {
      stamp.position(stamp.position() + 3);
    }

    if (!filled.get(25)) {
      valHi.position(valHi.position() + 16);
    }

    if (!filled.get(26)) {
      valLo.position(valLo.position() + 16);
    }

    if (!filled.get(27)) {
      warm.position(warm.position() + 1);
    }

    if (!filled.get(28)) {
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
