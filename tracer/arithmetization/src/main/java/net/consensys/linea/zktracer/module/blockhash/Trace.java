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

package net.consensys.linea.zktracer.module.blockhash;

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
  public static final int BLOCKHASH_DEPTH = 0x6;
  public static final int NEGATIVE_OF_BLOCKHASH_DEPTH = -0x6;
  public static final int ROFF___ABS___comparison_to_256 = 0x3;
  public static final int ROFF___BLOCKHASH_arguments___equality_test = 0x2;
  public static final int ROFF___BLOCKHASH_arguments___monotony = 0x1;
  public static final int ROFF___curr_BLOCKHASH_argument___comparison_to_max = 0x4;
  public static final int ROFF___curr_BLOCKHASH_argument___comparison_to_min = 0x5;
  public static final int nROWS_MACRO = 0x1;
  public static final int nROWS_PRPRC = 0x5;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absBlock;
  private final MappedByteBuffer blockhashArgHiXorExoArg1Hi;
  private final MappedByteBuffer blockhashArgLoXorExoArg1Lo;
  private final MappedByteBuffer blockhashResHiXorExoArg2Hi;
  private final MappedByteBuffer blockhashResLoXorExoArg2Lo;
  private final MappedByteBuffer blockhashValHi;
  private final MappedByteBuffer blockhashValLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer exoInst;
  private final MappedByteBuffer exoRes;
  private final MappedByteBuffer iomf;
  private final MappedByteBuffer macro;
  private final MappedByteBuffer prprc;
  private final MappedByteBuffer relBlock;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("blockhash.ABS_BLOCK", 6, length));
      headers.add(new ColumnHeader("blockhash.BLOCKHASH_ARG_HI_xor_EXO_ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("blockhash.BLOCKHASH_ARG_LO_xor_EXO_ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("blockhash.BLOCKHASH_RES_HI_xor_EXO_ARG_2_HI", 16, length));
      headers.add(new ColumnHeader("blockhash.BLOCKHASH_RES_LO_xor_EXO_ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("blockhash.BLOCKHASH_VAL_HI", 16, length));
      headers.add(new ColumnHeader("blockhash.BLOCKHASH_VAL_LO", 16, length));
      headers.add(new ColumnHeader("blockhash.CT", 1, length));
      headers.add(new ColumnHeader("blockhash.CT_MAX", 1, length));
      headers.add(new ColumnHeader("blockhash.EXO_INST", 1, length));
      headers.add(new ColumnHeader("blockhash.EXO_RES", 1, length));
      headers.add(new ColumnHeader("blockhash.IOMF", 1, length));
      headers.add(new ColumnHeader("blockhash.MACRO", 1, length));
      headers.add(new ColumnHeader("blockhash.PRPRC", 1, length));
      headers.add(new ColumnHeader("blockhash.REL_BLOCK", 2, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.absBlock = buffers.get(0);
    this.blockhashArgHiXorExoArg1Hi = buffers.get(1);
    this.blockhashArgLoXorExoArg1Lo = buffers.get(2);
    this.blockhashResHiXorExoArg2Hi = buffers.get(3);
    this.blockhashResLoXorExoArg2Lo = buffers.get(4);
    this.blockhashValHi = buffers.get(5);
    this.blockhashValLo = buffers.get(6);
    this.ct = buffers.get(7);
    this.ctMax = buffers.get(8);
    this.exoInst = buffers.get(9);
    this.exoRes = buffers.get(10);
    this.iomf = buffers.get(11);
    this.macro = buffers.get(12);
    this.prprc = buffers.get(13);
    this.relBlock = buffers.get(14);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace ct(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("blockhash.CT already set");
    } else {
      filled.set(0);
    }

    if(b >= 256L) { throw new IllegalArgumentException("blockhash.CT has invalid value (" + b + ")"); }
    ct.put((byte) b);


    return this;
  }

  public Trace ctMax(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("blockhash.CT_MAX already set");
    } else {
      filled.set(1);
    }

    if(b >= 256L) { throw new IllegalArgumentException("blockhash.CT_MAX has invalid value (" + b + ")"); }
    ctMax.put((byte) b);


    return this;
  }

  public Trace iomf(final Boolean b) {
    if (filled.get(2)) {
      throw new IllegalStateException("blockhash.IOMF already set");
    } else {
      filled.set(2);
    }

    iomf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace macro(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("blockhash.MACRO already set");
    } else {
      filled.set(3);
    }

    macro.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMacroAbsBlock(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("blockhash.macro/ABS_BLOCK already set");
    } else {
      filled.set(8);
    }

    if(b >= 281474976710656L) { throw new IllegalArgumentException("blockhash.macro/ABS_BLOCK has invalid value (" + b + ")"); }
    absBlock.put((byte) (b >> 40));
    absBlock.put((byte) (b >> 32));
    absBlock.put((byte) (b >> 24));
    absBlock.put((byte) (b >> 16));
    absBlock.put((byte) (b >> 8));
    absBlock.put((byte) b);


    return this;
  }

  public Trace pMacroBlockhashArgHi(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blockhash.macro/BLOCKHASH_ARG_HI already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.macro/BLOCKHASH_ARG_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashArgHiXorExoArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashArgHiXorExoArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroBlockhashArgLo(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blockhash.macro/BLOCKHASH_ARG_LO already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.macro/BLOCKHASH_ARG_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashArgLoXorExoArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashArgLoXorExoArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroBlockhashResHi(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blockhash.macro/BLOCKHASH_RES_HI already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.macro/BLOCKHASH_RES_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashResHiXorExoArg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashResHiXorExoArg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroBlockhashResLo(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blockhash.macro/BLOCKHASH_RES_LO already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.macro/BLOCKHASH_RES_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashResLoXorExoArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashResLoXorExoArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroBlockhashValHi(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("blockhash.macro/BLOCKHASH_VAL_HI already set");
    } else {
      filled.set(13);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.macro/BLOCKHASH_VAL_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashValHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashValHi.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroBlockhashValLo(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("blockhash.macro/BLOCKHASH_VAL_LO already set");
    } else {
      filled.set(14);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.macro/BLOCKHASH_VAL_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashValLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashValLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroRelBlock(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("blockhash.macro/REL_BLOCK already set");
    } else {
      filled.set(7);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("blockhash.macro/REL_BLOCK has invalid value (" + b + ")"); }
    relBlock.put((byte) (b >> 8));
    relBlock.put((byte) b);


    return this;
  }

  public Trace pPreprocessingExoArg1Hi(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blockhash.preprocessing/EXO_ARG_1_HI already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.preprocessing/EXO_ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashArgHiXorExoArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashArgHiXorExoArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingExoArg1Lo(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blockhash.preprocessing/EXO_ARG_1_LO already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.preprocessing/EXO_ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashArgLoXorExoArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashArgLoXorExoArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingExoArg2Hi(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blockhash.preprocessing/EXO_ARG_2_HI already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.preprocessing/EXO_ARG_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashResHiXorExoArg2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashResHiXorExoArg2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingExoArg2Lo(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blockhash.preprocessing/EXO_ARG_2_LO already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blockhash.preprocessing/EXO_ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { blockhashResLoXorExoArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockhashResLoXorExoArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pPreprocessingExoInst(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("blockhash.preprocessing/EXO_INST already set");
    } else {
      filled.set(6);
    }

    if(b >= 256L) { throw new IllegalArgumentException("blockhash.preprocessing/EXO_INST has invalid value (" + b + ")"); }
    exoInst.put((byte) b);


    return this;
  }

  public Trace pPreprocessingExoRes(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("blockhash.preprocessing/EXO_RES already set");
    } else {
      filled.set(5);
    }

    exoRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prprc(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("blockhash.PRPRC already set");
    } else {
      filled.set(4);
    }

    prprc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(8)) {
      throw new IllegalStateException("blockhash.ABS_BLOCK has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("blockhash.BLOCKHASH_ARG_HI_xor_EXO_ARG_1_HI has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("blockhash.BLOCKHASH_ARG_LO_xor_EXO_ARG_1_LO has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("blockhash.BLOCKHASH_RES_HI_xor_EXO_ARG_2_HI has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("blockhash.BLOCKHASH_RES_LO_xor_EXO_ARG_2_LO has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("blockhash.BLOCKHASH_VAL_HI has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("blockhash.BLOCKHASH_VAL_LO has not been filled");
    }

    if (!filled.get(0)) {
      throw new IllegalStateException("blockhash.CT has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("blockhash.CT_MAX has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("blockhash.EXO_INST has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("blockhash.EXO_RES has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("blockhash.IOMF has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("blockhash.MACRO has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("blockhash.PRPRC has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("blockhash.REL_BLOCK has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(8)) {
      absBlock.position(absBlock.position() + 6);
    }

    if (!filled.get(9)) {
      blockhashArgHiXorExoArg1Hi.position(blockhashArgHiXorExoArg1Hi.position() + 16);
    }

    if (!filled.get(10)) {
      blockhashArgLoXorExoArg1Lo.position(blockhashArgLoXorExoArg1Lo.position() + 16);
    }

    if (!filled.get(11)) {
      blockhashResHiXorExoArg2Hi.position(blockhashResHiXorExoArg2Hi.position() + 16);
    }

    if (!filled.get(12)) {
      blockhashResLoXorExoArg2Lo.position(blockhashResLoXorExoArg2Lo.position() + 16);
    }

    if (!filled.get(13)) {
      blockhashValHi.position(blockhashValHi.position() + 16);
    }

    if (!filled.get(14)) {
      blockhashValLo.position(blockhashValLo.position() + 16);
    }

    if (!filled.get(0)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(1)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(6)) {
      exoInst.position(exoInst.position() + 1);
    }

    if (!filled.get(5)) {
      exoRes.position(exoRes.position() + 1);
    }

    if (!filled.get(2)) {
      iomf.position(iomf.position() + 1);
    }

    if (!filled.get(3)) {
      macro.position(macro.position() + 1);
    }

    if (!filled.get(4)) {
      prprc.position(prprc.position() + 1);
    }

    if (!filled.get(7)) {
      relBlock.position(relBlock.position() + 2);
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
