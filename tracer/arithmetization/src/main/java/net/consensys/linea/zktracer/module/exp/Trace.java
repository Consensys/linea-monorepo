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

package net.consensys.linea.zktracer.module.exp;

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
  public static final int CT_MAX_CMPTN_EXP_LOG = 0xf;
  public static final int CT_MAX_CMPTN_MODEXP_LOG = 0xf;
  public static final int CT_MAX_MACRO_EXP_LOG = 0x0;
  public static final int CT_MAX_MACRO_MODEXP_LOG = 0x0;
  public static final int CT_MAX_PRPRC_EXP_LOG = 0x0;
  public static final int CT_MAX_PRPRC_MODEXP_LOG = 0x3;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer cmptn;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer data3XorWcpArg2Hi;
  private final MappedByteBuffer data4XorWcpArg2Lo;
  private final MappedByteBuffer data5;
  private final MappedByteBuffer expInst;
  private final MappedByteBuffer isExpLog;
  private final MappedByteBuffer isModexpLog;
  private final MappedByteBuffer macro;
  private final MappedByteBuffer manzbAcc;
  private final MappedByteBuffer manzbXorWcpFlag;
  private final MappedByteBuffer msbAcc;
  private final MappedByteBuffer msbBitXorWcpRes;
  private final MappedByteBuffer msbXorWcpInst;
  private final MappedByteBuffer pltBit;
  private final MappedByteBuffer pltJmp;
  private final MappedByteBuffer prprc;
  private final MappedByteBuffer rawAccXorData1XorWcpArg1Hi;
  private final MappedByteBuffer rawByte;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer tanzb;
  private final MappedByteBuffer tanzbAcc;
  private final MappedByteBuffer trimAccXorData2XorWcpArg1Lo;
  private final MappedByteBuffer trimByte;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("exp.CMPTN", 1, length));
      headers.add(new ColumnHeader("exp.CT", 1, length));
      headers.add(new ColumnHeader("exp.CT_MAX", 1, length));
      headers.add(new ColumnHeader("exp.DATA_3_xor_WCP_ARG_2_HI", 16, length));
      headers.add(new ColumnHeader("exp.DATA_4_xor_WCP_ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("exp.DATA_5", 16, length));
      headers.add(new ColumnHeader("exp.EXP_INST", 2, length));
      headers.add(new ColumnHeader("exp.IS_EXP_LOG", 1, length));
      headers.add(new ColumnHeader("exp.IS_MODEXP_LOG", 1, length));
      headers.add(new ColumnHeader("exp.MACRO", 1, length));
      headers.add(new ColumnHeader("exp.MANZB_ACC", 1, length));
      headers.add(new ColumnHeader("exp.MANZB_xor_WCP_FLAG", 1, length));
      headers.add(new ColumnHeader("exp.MSB_ACC", 1, length));
      headers.add(new ColumnHeader("exp.MSB_BIT_xor_WCP_RES", 1, length));
      headers.add(new ColumnHeader("exp.MSB_xor_WCP_INST", 1, length));
      headers.add(new ColumnHeader("exp.PLT_BIT", 1, length));
      headers.add(new ColumnHeader("exp.PLT_JMP", 1, length));
      headers.add(new ColumnHeader("exp.PRPRC", 1, length));
      headers.add(new ColumnHeader("exp.RAW_ACC_xor_DATA_1_xor_WCP_ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("exp.RAW_BYTE", 1, length));
      headers.add(new ColumnHeader("exp.STAMP", 4, length));
      headers.add(new ColumnHeader("exp.TANZB", 1, length));
      headers.add(new ColumnHeader("exp.TANZB_ACC", 1, length));
      headers.add(new ColumnHeader("exp.TRIM_ACC_xor_DATA_2_xor_WCP_ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("exp.TRIM_BYTE", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.cmptn = buffers.get(0);
    this.ct = buffers.get(1);
    this.ctMax = buffers.get(2);
    this.data3XorWcpArg2Hi = buffers.get(3);
    this.data4XorWcpArg2Lo = buffers.get(4);
    this.data5 = buffers.get(5);
    this.expInst = buffers.get(6);
    this.isExpLog = buffers.get(7);
    this.isModexpLog = buffers.get(8);
    this.macro = buffers.get(9);
    this.manzbAcc = buffers.get(10);
    this.manzbXorWcpFlag = buffers.get(11);
    this.msbAcc = buffers.get(12);
    this.msbBitXorWcpRes = buffers.get(13);
    this.msbXorWcpInst = buffers.get(14);
    this.pltBit = buffers.get(15);
    this.pltJmp = buffers.get(16);
    this.prprc = buffers.get(17);
    this.rawAccXorData1XorWcpArg1Hi = buffers.get(18);
    this.rawByte = buffers.get(19);
    this.stamp = buffers.get(20);
    this.tanzb = buffers.get(21);
    this.tanzbAcc = buffers.get(22);
    this.trimAccXorData2XorWcpArg1Lo = buffers.get(23);
    this.trimByte = buffers.get(24);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace cmptn(final Boolean b) {
    if (filled.get(0)) {
      throw new IllegalStateException("exp.CMPTN already set");
    } else {
      filled.set(0);
    }

    cmptn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("exp.CT already set");
    } else {
      filled.set(1);
    }

    if(b >= 16L) { throw new IllegalArgumentException("exp.CT has invalid value (" + b + ")"); }
    ct.put((byte) b);


    return this;
  }

  public Trace ctMax(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("exp.CT_MAX already set");
    } else {
      filled.set(2);
    }

    if(b >= 16L) { throw new IllegalArgumentException("exp.CT_MAX has invalid value (" + b + ")"); }
    ctMax.put((byte) b);


    return this;
  }

  public Trace isExpLog(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("exp.IS_EXP_LOG already set");
    } else {
      filled.set(3);
    }

    isExpLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpLog(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("exp.IS_MODEXP_LOG already set");
    } else {
      filled.set(4);
    }

    isModexpLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace macro(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("exp.MACRO already set");
    } else {
      filled.set(5);
    }

    macro.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationManzb(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("exp.computation/MANZB already set");
    } else {
      filled.set(8);
    }

    manzbXorWcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationManzbAcc(final long b) {
    if (filled.get(16)) {
      throw new IllegalStateException("exp.computation/MANZB_ACC already set");
    } else {
      filled.set(16);
    }

    if(b >= 16L) { throw new IllegalArgumentException("exp.computation/MANZB_ACC has invalid value (" + b + ")"); }
    manzbAcc.put((byte) b);


    return this;
  }

  public Trace pComputationMsb(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("exp.computation/MSB already set");
    } else {
      filled.set(12);
    }

    msbXorWcpInst.put(b.toByte());

    return this;
  }

  public Trace pComputationMsbAcc(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("exp.computation/MSB_ACC already set");
    } else {
      filled.set(13);
    }

    msbAcc.put(b.toByte());

    return this;
  }

  public Trace pComputationMsbBit(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("exp.computation/MSB_BIT already set");
    } else {
      filled.set(9);
    }

    msbBitXorWcpRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationPltBit(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("exp.computation/PLT_BIT already set");
    } else {
      filled.set(10);
    }

    pltBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationPltJmp(final long b) {
    if (filled.get(18)) {
      throw new IllegalStateException("exp.computation/PLT_JMP already set");
    } else {
      filled.set(18);
    }

    if(b >= 64L) { throw new IllegalArgumentException("exp.computation/PLT_JMP has invalid value (" + b + ")"); }
    pltJmp.put((byte) b);


    return this;
  }

  public Trace pComputationRawAcc(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("exp.computation/RAW_ACC already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.computation/RAW_ACC has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rawAccXorData1XorWcpArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rawAccXorData1XorWcpArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pComputationRawByte(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("exp.computation/RAW_BYTE already set");
    } else {
      filled.set(14);
    }

    rawByte.put(b.toByte());

    return this;
  }

  public Trace pComputationTanzb(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("exp.computation/TANZB already set");
    } else {
      filled.set(11);
    }

    tanzb.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationTanzbAcc(final long b) {
    if (filled.get(17)) {
      throw new IllegalStateException("exp.computation/TANZB_ACC already set");
    } else {
      filled.set(17);
    }

    if(b >= 32L) { throw new IllegalArgumentException("exp.computation/TANZB_ACC has invalid value (" + b + ")"); }
    tanzbAcc.put((byte) b);


    return this;
  }

  public Trace pComputationTrimAcc(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("exp.computation/TRIM_ACC already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.computation/TRIM_ACC has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { trimAccXorData2XorWcpArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { trimAccXorData2XorWcpArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pComputationTrimByte(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("exp.computation/TRIM_BYTE already set");
    } else {
      filled.set(15);
    }

    trimByte.put(b.toByte());

    return this;
  }

  public Trace pMacroData1(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("exp.macro/DATA_1 already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.macro/DATA_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rawAccXorData1XorWcpArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rawAccXorData1XorWcpArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroData2(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("exp.macro/DATA_2 already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.macro/DATA_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { trimAccXorData2XorWcpArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { trimAccXorData2XorWcpArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroData3(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("exp.macro/DATA_3 already set");
    } else {
      filled.set(22);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.macro/DATA_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data3XorWcpArg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data3XorWcpArg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroData4(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("exp.macro/DATA_4 already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.macro/DATA_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data4XorWcpArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data4XorWcpArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroData5(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("exp.macro/DATA_5 already set");
    } else {
      filled.set(24);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.macro/DATA_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data5.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroExpInst(final long b) {
    if (filled.get(19)) {
      throw new IllegalStateException("exp.macro/EXP_INST already set");
    } else {
      filled.set(19);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("exp.macro/EXP_INST has invalid value (" + b + ")"); }
    expInst.put((byte) (b >> 8));
    expInst.put((byte) b);


    return this;
  }

  public Trace pPreprocessingWcpArg1Hi(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_1_HI already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.preprocessing/WCP_ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rawAccXorData1XorWcpArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rawAccXorData1XorWcpArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingWcpArg1Lo(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_1_LO already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.preprocessing/WCP_ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { trimAccXorData2XorWcpArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { trimAccXorData2XorWcpArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingWcpArg2Hi(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_2_HI already set");
    } else {
      filled.set(22);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.preprocessing/WCP_ARG_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data3XorWcpArg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data3XorWcpArg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingWcpArg2Lo(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_2_LO already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("exp.preprocessing/WCP_ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { data4XorWcpArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { data4XorWcpArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingWcpFlag(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("exp.preprocessing/WCP_FLAG already set");
    } else {
      filled.set(8);
    }

    manzbXorWcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pPreprocessingWcpInst(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("exp.preprocessing/WCP_INST already set");
    } else {
      filled.set(12);
    }

    msbXorWcpInst.put(b.toByte());

    return this;
  }

  public Trace pPreprocessingWcpRes(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("exp.preprocessing/WCP_RES already set");
    } else {
      filled.set(9);
    }

    msbBitXorWcpRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prprc(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("exp.PRPRC already set");
    } else {
      filled.set(6);
    }

    prprc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("exp.STAMP already set");
    } else {
      filled.set(7);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("exp.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 24));
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("exp.CMPTN has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("exp.CT has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("exp.CT_MAX has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("exp.DATA_3_xor_WCP_ARG_2_HI has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("exp.DATA_4_xor_WCP_ARG_2_LO has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("exp.DATA_5 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("exp.EXP_INST has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("exp.IS_EXP_LOG has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("exp.IS_MODEXP_LOG has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("exp.MACRO has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("exp.MANZB_ACC has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("exp.MANZB_xor_WCP_FLAG has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("exp.MSB_ACC has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("exp.MSB_BIT_xor_WCP_RES has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("exp.MSB_xor_WCP_INST has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("exp.PLT_BIT has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("exp.PLT_JMP has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("exp.PRPRC has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("exp.RAW_ACC_xor_DATA_1_xor_WCP_ARG_1_HI has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("exp.RAW_BYTE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("exp.STAMP has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("exp.TANZB has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("exp.TANZB_ACC has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("exp.TRIM_ACC_xor_DATA_2_xor_WCP_ARG_1_LO has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("exp.TRIM_BYTE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      cmptn.position(cmptn.position() + 1);
    }

    if (!filled.get(1)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(2)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(22)) {
      data3XorWcpArg2Hi.position(data3XorWcpArg2Hi.position() + 16);
    }

    if (!filled.get(23)) {
      data4XorWcpArg2Lo.position(data4XorWcpArg2Lo.position() + 16);
    }

    if (!filled.get(24)) {
      data5.position(data5.position() + 16);
    }

    if (!filled.get(19)) {
      expInst.position(expInst.position() + 2);
    }

    if (!filled.get(3)) {
      isExpLog.position(isExpLog.position() + 1);
    }

    if (!filled.get(4)) {
      isModexpLog.position(isModexpLog.position() + 1);
    }

    if (!filled.get(5)) {
      macro.position(macro.position() + 1);
    }

    if (!filled.get(16)) {
      manzbAcc.position(manzbAcc.position() + 1);
    }

    if (!filled.get(8)) {
      manzbXorWcpFlag.position(manzbXorWcpFlag.position() + 1);
    }

    if (!filled.get(13)) {
      msbAcc.position(msbAcc.position() + 1);
    }

    if (!filled.get(9)) {
      msbBitXorWcpRes.position(msbBitXorWcpRes.position() + 1);
    }

    if (!filled.get(12)) {
      msbXorWcpInst.position(msbXorWcpInst.position() + 1);
    }

    if (!filled.get(10)) {
      pltBit.position(pltBit.position() + 1);
    }

    if (!filled.get(18)) {
      pltJmp.position(pltJmp.position() + 1);
    }

    if (!filled.get(6)) {
      prprc.position(prprc.position() + 1);
    }

    if (!filled.get(20)) {
      rawAccXorData1XorWcpArg1Hi.position(rawAccXorData1XorWcpArg1Hi.position() + 16);
    }

    if (!filled.get(14)) {
      rawByte.position(rawByte.position() + 1);
    }

    if (!filled.get(7)) {
      stamp.position(stamp.position() + 4);
    }

    if (!filled.get(11)) {
      tanzb.position(tanzb.position() + 1);
    }

    if (!filled.get(17)) {
      tanzbAcc.position(tanzbAcc.position() + 1);
    }

    if (!filled.get(21)) {
      trimAccXorData2XorWcpArg1Lo.position(trimAccXorData2XorWcpArg1Lo.position() + 16);
    }

    if (!filled.get(15)) {
      trimByte.position(trimByte.position() + 1);
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
