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

package net.consensys.linea.zktracer.module.mmu;

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
  public static final int NB_MICRO_ROWS_TOT_BLAKE = 0x2;
  public static final int NB_MICRO_ROWS_TOT_INVALID_CODE_PREFIX = 0x1;
  public static final int NB_MICRO_ROWS_TOT_MLOAD = 0x2;
  public static final int NB_MICRO_ROWS_TOT_MODEXP_DATA = 0x20;
  public static final int NB_MICRO_ROWS_TOT_MODEXP_ZERO = 0x20;
  public static final int NB_MICRO_ROWS_TOT_MSTORE = 0x2;
  public static final int NB_MICRO_ROWS_TOT_MSTORE_EIGHT = 0x1;
  public static final int NB_MICRO_ROWS_TOT_RIGHT_PADDED_WORD_EXTRACTION = 0x2;
  public static final int NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING = 0x4;
  public static final int NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING_PO = 0x5;
  public static final int NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA = 0xa;
  public static final int NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA_PO = 0xb;
  public static final int NB_PP_ROWS_BLAKE = 0x2;
  public static final int NB_PP_ROWS_BLAKE_PO = 0x3;
  public static final int NB_PP_ROWS_BLAKE_PT = 0x4;
  public static final int NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS = 0x1;
  public static final int NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS_PO = 0x2;
  public static final int NB_PP_ROWS_INVALID_CODE_PREFIX = 0x1;
  public static final int NB_PP_ROWS_INVALID_CODE_PREFIX_PO = 0x2;
  public static final int NB_PP_ROWS_MLOAD = 0x1;
  public static final int NB_PP_ROWS_MLOAD_PO = 0x2;
  public static final int NB_PP_ROWS_MLOAD_PT = 0x3;
  public static final int NB_PP_ROWS_MODEXP_DATA = 0x6;
  public static final int NB_PP_ROWS_MODEXP_DATA_PO = 0x7;
  public static final int NB_PP_ROWS_MODEXP_ZERO = 0x1;
  public static final int NB_PP_ROWS_MODEXP_ZERO_PO = 0x2;
  public static final int NB_PP_ROWS_MSTORE = 0x1;
  public static final int NB_PP_ROWS_MSTORE8 = 0x1;
  public static final int NB_PP_ROWS_MSTORE8_PO = 0x2;
  public static final int NB_PP_ROWS_MSTORE_PO = 0x2;
  public static final int NB_PP_ROWS_MSTORE_PT = 0x3;
  public static final int NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING = 0x4;
  public static final int NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING_PO = 0x5;
  public static final int NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING = 0x5;
  public static final int NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING_PO = 0x6;
  public static final int NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION = 0x5;
  public static final int NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION_PO = 0x6;
  public static final int NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION_PT = 0x7;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer auxIdXorCnSXorEucA;
  private final MappedByteBuffer bin1;
  private final MappedByteBuffer bin2;
  private final MappedByteBuffer bin3;
  private final MappedByteBuffer bin4;
  private final MappedByteBuffer bin5;
  private final MappedByteBuffer exoSumXorExoId;
  private final MappedByteBuffer instXorInstXorCt;
  private final MappedByteBuffer isAnyToRamWithPaddingPurePadding;
  private final MappedByteBuffer isAnyToRamWithPaddingSomeData;
  private final MappedByteBuffer isBlake;
  private final MappedByteBuffer isExoToRamTransplants;
  private final MappedByteBuffer isInvalidCodePrefix;
  private final MappedByteBuffer isMload;
  private final MappedByteBuffer isModexpData;
  private final MappedByteBuffer isModexpZero;
  private final MappedByteBuffer isMstore;
  private final MappedByteBuffer isMstore8;
  private final MappedByteBuffer isRamToExoWithPadding;
  private final MappedByteBuffer isRamToRamSansPadding;
  private final MappedByteBuffer isRightPaddedWordExtraction;
  private final MappedByteBuffer kecId;
  private final MappedByteBuffer limb1XorLimbXorWcpArg1Hi;
  private final MappedByteBuffer limb2XorWcpArg1Lo;
  private final MappedByteBuffer lzro;
  private final MappedByteBuffer macro;
  private final MappedByteBuffer micro;
  private final MappedByteBuffer mmioStamp;
  private final MappedByteBuffer ntFirst;
  private final MappedByteBuffer ntLast;
  private final MappedByteBuffer ntMddl;
  private final MappedByteBuffer ntOnly;
  private final MappedByteBuffer out1;
  private final MappedByteBuffer out2;
  private final MappedByteBuffer out3;
  private final MappedByteBuffer out4;
  private final MappedByteBuffer out5;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer phaseXorExoSum;
  private final MappedByteBuffer prprc;
  private final MappedByteBuffer refOffsetXorCnTXorEucB;
  private final MappedByteBuffer refSizeXorSloXorEucCeil;
  private final MappedByteBuffer rzFirst;
  private final MappedByteBuffer rzLast;
  private final MappedByteBuffer rzMddl;
  private final MappedByteBuffer rzOnly;
  private final MappedByteBuffer sboXorWcpInst;
  private final MappedByteBuffer size;
  private final MappedByteBuffer sizeXorTloXorEucQuot;
  private final MappedByteBuffer srcIdXorTotalSizeXorEucRem;
  private final MappedByteBuffer srcOffsetHiXorWcpArg2Lo;
  private final MappedByteBuffer srcOffsetLo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer successBitXorSuccessBitXorEucFlag;
  private final MappedByteBuffer tbo;
  private final MappedByteBuffer tgtId;
  private final MappedByteBuffer tgtOffsetLo;
  private final MappedByteBuffer tot;
  private final MappedByteBuffer totlz;
  private final MappedByteBuffer totnt;
  private final MappedByteBuffer totrz;
  private final MappedByteBuffer wcpFlag;
  private final MappedByteBuffer wcpRes;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("mmu.AUX_ID_xor_CN_S_xor_EUC_A", 8, length));
      headers.add(new ColumnHeader("mmu.BIN_1", 1, length));
      headers.add(new ColumnHeader("mmu.BIN_2", 1, length));
      headers.add(new ColumnHeader("mmu.BIN_3", 1, length));
      headers.add(new ColumnHeader("mmu.BIN_4", 1, length));
      headers.add(new ColumnHeader("mmu.BIN_5", 1, length));
      headers.add(new ColumnHeader("mmu.EXO_SUM_xor_EXO_ID", 4, length));
      headers.add(new ColumnHeader("mmu.INST_xor_INST_xor_CT", 2, length));
      headers.add(new ColumnHeader("mmu.IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING", 1, length));
      headers.add(new ColumnHeader("mmu.IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA", 1, length));
      headers.add(new ColumnHeader("mmu.IS_BLAKE", 1, length));
      headers.add(new ColumnHeader("mmu.IS_EXO_TO_RAM_TRANSPLANTS", 1, length));
      headers.add(new ColumnHeader("mmu.IS_INVALID_CODE_PREFIX", 1, length));
      headers.add(new ColumnHeader("mmu.IS_MLOAD", 1, length));
      headers.add(new ColumnHeader("mmu.IS_MODEXP_DATA", 1, length));
      headers.add(new ColumnHeader("mmu.IS_MODEXP_ZERO", 1, length));
      headers.add(new ColumnHeader("mmu.IS_MSTORE", 1, length));
      headers.add(new ColumnHeader("mmu.IS_MSTORE8", 1, length));
      headers.add(new ColumnHeader("mmu.IS_RAM_TO_EXO_WITH_PADDING", 1, length));
      headers.add(new ColumnHeader("mmu.IS_RAM_TO_RAM_SANS_PADDING", 1, length));
      headers.add(new ColumnHeader("mmu.IS_RIGHT_PADDED_WORD_EXTRACTION", 1, length));
      headers.add(new ColumnHeader("mmu.KEC_ID", 4, length));
      headers.add(new ColumnHeader("mmu.LIMB_1_xor_LIMB_xor_WCP_ARG_1_HI", 16, length));
      headers.add(new ColumnHeader("mmu.LIMB_2_xor_WCP_ARG_1_LO", 16, length));
      headers.add(new ColumnHeader("mmu.LZRO", 1, length));
      headers.add(new ColumnHeader("mmu.MACRO", 1, length));
      headers.add(new ColumnHeader("mmu.MICRO", 1, length));
      headers.add(new ColumnHeader("mmu.MMIO_STAMP", 4, length));
      headers.add(new ColumnHeader("mmu.NT_FIRST", 1, length));
      headers.add(new ColumnHeader("mmu.NT_LAST", 1, length));
      headers.add(new ColumnHeader("mmu.NT_MDDL", 1, length));
      headers.add(new ColumnHeader("mmu.NT_ONLY", 1, length));
      headers.add(new ColumnHeader("mmu.OUT_1", 8, length));
      headers.add(new ColumnHeader("mmu.OUT_2", 8, length));
      headers.add(new ColumnHeader("mmu.OUT_3", 8, length));
      headers.add(new ColumnHeader("mmu.OUT_4", 8, length));
      headers.add(new ColumnHeader("mmu.OUT_5", 8, length));
      headers.add(new ColumnHeader("mmu.PHASE", 4, length));
      headers.add(new ColumnHeader("mmu.PHASE_xor_EXO_SUM", 4, length));
      headers.add(new ColumnHeader("mmu.PRPRC", 1, length));
      headers.add(new ColumnHeader("mmu.REF_OFFSET_xor_CN_T_xor_EUC_B", 8, length));
      headers.add(new ColumnHeader("mmu.REF_SIZE_xor_SLO_xor_EUC_CEIL", 8, length));
      headers.add(new ColumnHeader("mmu.RZ_FIRST", 1, length));
      headers.add(new ColumnHeader("mmu.RZ_LAST", 1, length));
      headers.add(new ColumnHeader("mmu.RZ_MDDL", 1, length));
      headers.add(new ColumnHeader("mmu.RZ_ONLY", 1, length));
      headers.add(new ColumnHeader("mmu.SBO_xor_WCP_INST", 1, length));
      headers.add(new ColumnHeader("mmu.SIZE", 1, length));
      headers.add(new ColumnHeader("mmu.SIZE_xor_TLO_xor_EUC_QUOT", 8, length));
      headers.add(new ColumnHeader("mmu.SRC_ID_xor_TOTAL_SIZE_xor_EUC_REM", 8, length));
      headers.add(new ColumnHeader("mmu.SRC_OFFSET_HI_xor_WCP_ARG_2_LO", 16, length));
      headers.add(new ColumnHeader("mmu.SRC_OFFSET_LO", 16, length));
      headers.add(new ColumnHeader("mmu.STAMP", 4, length));
      headers.add(new ColumnHeader("mmu.SUCCESS_BIT_xor_SUCCESS_BIT_xor_EUC_FLAG", 1, length));
      headers.add(new ColumnHeader("mmu.TBO", 1, length));
      headers.add(new ColumnHeader("mmu.TGT_ID", 8, length));
      headers.add(new ColumnHeader("mmu.TGT_OFFSET_LO", 8, length));
      headers.add(new ColumnHeader("mmu.TOT", 4, length));
      headers.add(new ColumnHeader("mmu.TOTLZ", 4, length));
      headers.add(new ColumnHeader("mmu.TOTNT", 4, length));
      headers.add(new ColumnHeader("mmu.TOTRZ", 4, length));
      headers.add(new ColumnHeader("mmu.WCP_FLAG", 1, length));
      headers.add(new ColumnHeader("mmu.WCP_RES", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.auxIdXorCnSXorEucA = buffers.get(0);
    this.bin1 = buffers.get(1);
    this.bin2 = buffers.get(2);
    this.bin3 = buffers.get(3);
    this.bin4 = buffers.get(4);
    this.bin5 = buffers.get(5);
    this.exoSumXorExoId = buffers.get(6);
    this.instXorInstXorCt = buffers.get(7);
    this.isAnyToRamWithPaddingPurePadding = buffers.get(8);
    this.isAnyToRamWithPaddingSomeData = buffers.get(9);
    this.isBlake = buffers.get(10);
    this.isExoToRamTransplants = buffers.get(11);
    this.isInvalidCodePrefix = buffers.get(12);
    this.isMload = buffers.get(13);
    this.isModexpData = buffers.get(14);
    this.isModexpZero = buffers.get(15);
    this.isMstore = buffers.get(16);
    this.isMstore8 = buffers.get(17);
    this.isRamToExoWithPadding = buffers.get(18);
    this.isRamToRamSansPadding = buffers.get(19);
    this.isRightPaddedWordExtraction = buffers.get(20);
    this.kecId = buffers.get(21);
    this.limb1XorLimbXorWcpArg1Hi = buffers.get(22);
    this.limb2XorWcpArg1Lo = buffers.get(23);
    this.lzro = buffers.get(24);
    this.macro = buffers.get(25);
    this.micro = buffers.get(26);
    this.mmioStamp = buffers.get(27);
    this.ntFirst = buffers.get(28);
    this.ntLast = buffers.get(29);
    this.ntMddl = buffers.get(30);
    this.ntOnly = buffers.get(31);
    this.out1 = buffers.get(32);
    this.out2 = buffers.get(33);
    this.out3 = buffers.get(34);
    this.out4 = buffers.get(35);
    this.out5 = buffers.get(36);
    this.phase = buffers.get(37);
    this.phaseXorExoSum = buffers.get(38);
    this.prprc = buffers.get(39);
    this.refOffsetXorCnTXorEucB = buffers.get(40);
    this.refSizeXorSloXorEucCeil = buffers.get(41);
    this.rzFirst = buffers.get(42);
    this.rzLast = buffers.get(43);
    this.rzMddl = buffers.get(44);
    this.rzOnly = buffers.get(45);
    this.sboXorWcpInst = buffers.get(46);
    this.size = buffers.get(47);
    this.sizeXorTloXorEucQuot = buffers.get(48);
    this.srcIdXorTotalSizeXorEucRem = buffers.get(49);
    this.srcOffsetHiXorWcpArg2Lo = buffers.get(50);
    this.srcOffsetLo = buffers.get(51);
    this.stamp = buffers.get(52);
    this.successBitXorSuccessBitXorEucFlag = buffers.get(53);
    this.tbo = buffers.get(54);
    this.tgtId = buffers.get(55);
    this.tgtOffsetLo = buffers.get(56);
    this.tot = buffers.get(57);
    this.totlz = buffers.get(58);
    this.totnt = buffers.get(59);
    this.totrz = buffers.get(60);
    this.wcpFlag = buffers.get(61);
    this.wcpRes = buffers.get(62);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace bin1(final Boolean b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mmu.BIN_1 already set");
    } else {
      filled.set(0);
    }

    bin1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bin2(final Boolean b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mmu.BIN_2 already set");
    } else {
      filled.set(1);
    }

    bin2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bin3(final Boolean b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mmu.BIN_3 already set");
    } else {
      filled.set(2);
    }

    bin3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bin4(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mmu.BIN_4 already set");
    } else {
      filled.set(3);
    }

    bin4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bin5(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mmu.BIN_5 already set");
    } else {
      filled.set(4);
    }

    bin5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isAnyToRamWithPaddingPurePadding(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mmu.IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING already set");
    } else {
      filled.set(5);
    }

    isAnyToRamWithPaddingPurePadding.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isAnyToRamWithPaddingSomeData(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mmu.IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA already set");
    } else {
      filled.set(6);
    }

    isAnyToRamWithPaddingSomeData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isBlake(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mmu.IS_BLAKE already set");
    } else {
      filled.set(7);
    }

    isBlake.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isExoToRamTransplants(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mmu.IS_EXO_TO_RAM_TRANSPLANTS already set");
    } else {
      filled.set(8);
    }

    isExoToRamTransplants.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isInvalidCodePrefix(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mmu.IS_INVALID_CODE_PREFIX already set");
    } else {
      filled.set(9);
    }

    isInvalidCodePrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isMload(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mmu.IS_MLOAD already set");
    } else {
      filled.set(10);
    }

    isMload.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpData(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mmu.IS_MODEXP_DATA already set");
    } else {
      filled.set(11);
    }

    isModexpData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpZero(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mmu.IS_MODEXP_ZERO already set");
    } else {
      filled.set(12);
    }

    isModexpZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isMstore(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mmu.IS_MSTORE already set");
    } else {
      filled.set(13);
    }

    isMstore.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isMstore8(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mmu.IS_MSTORE8 already set");
    } else {
      filled.set(14);
    }

    isMstore8.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToExoWithPadding(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mmu.IS_RAM_TO_EXO_WITH_PADDING already set");
    } else {
      filled.set(15);
    }

    isRamToExoWithPadding.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamSansPadding(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mmu.IS_RAM_TO_RAM_SANS_PADDING already set");
    } else {
      filled.set(16);
    }

    isRamToRamSansPadding.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRightPaddedWordExtraction(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mmu.IS_RIGHT_PADDED_WORD_EXTRACTION already set");
    } else {
      filled.set(17);
    }

    isRightPaddedWordExtraction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lzro(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mmu.LZRO already set");
    } else {
      filled.set(18);
    }

    lzro.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace macro(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mmu.MACRO already set");
    } else {
      filled.set(19);
    }

    macro.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace micro(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mmu.MICRO already set");
    } else {
      filled.set(20);
    }

    micro.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mmioStamp(final long b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mmu.MMIO_STAMP already set");
    } else {
      filled.set(21);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.MMIO_STAMP has invalid value (" + b + ")"); }
    mmioStamp.put((byte) (b >> 24));
    mmioStamp.put((byte) (b >> 16));
    mmioStamp.put((byte) (b >> 8));
    mmioStamp.put((byte) b);


    return this;
  }

  public Trace ntFirst(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mmu.NT_FIRST already set");
    } else {
      filled.set(22);
    }

    ntFirst.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ntLast(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mmu.NT_LAST already set");
    } else {
      filled.set(23);
    }

    ntLast.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ntMddl(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mmu.NT_MDDL already set");
    } else {
      filled.set(24);
    }

    ntMddl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ntOnly(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mmu.NT_ONLY already set");
    } else {
      filled.set(25);
    }

    ntOnly.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace out1(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mmu.OUT_1 already set");
    } else {
      filled.set(26);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.OUT_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { out1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { out1.put(bs.get(j)); }

    return this;
  }

  public Trace out2(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mmu.OUT_2 already set");
    } else {
      filled.set(27);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.OUT_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { out2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { out2.put(bs.get(j)); }

    return this;
  }

  public Trace out3(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mmu.OUT_3 already set");
    } else {
      filled.set(28);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.OUT_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { out3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { out3.put(bs.get(j)); }

    return this;
  }

  public Trace out4(final Bytes b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mmu.OUT_4 already set");
    } else {
      filled.set(29);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.OUT_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { out4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { out4.put(bs.get(j)); }

    return this;
  }

  public Trace out5(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mmu.OUT_5 already set");
    } else {
      filled.set(30);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.OUT_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { out5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { out5.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroAuxId(final Bytes b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmu.macro/AUX_ID already set");
    } else {
      filled.set(52);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/AUX_ID has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { auxIdXorCnSXorEucA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { auxIdXorCnSXorEucA.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroExoSum(final long b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmu.macro/EXO_SUM already set");
    } else {
      filled.set(48);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.macro/EXO_SUM has invalid value (" + b + ")"); }
    exoSumXorExoId.put((byte) (b >> 24));
    exoSumXorExoId.put((byte) (b >> 16));
    exoSumXorExoId.put((byte) (b >> 8));
    exoSumXorExoId.put((byte) b);


    return this;
  }

  public Trace pMacroInst(final long b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.macro/INST already set");
    } else {
      filled.set(47);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("mmu.macro/INST has invalid value (" + b + ")"); }
    instXorInstXorCt.put((byte) (b >> 8));
    instXorInstXorCt.put((byte) b);


    return this;
  }

  public Trace pMacroLimb1(final Bytes b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmu.macro/LIMB_1 already set");
    } else {
      filled.set(59);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.macro/LIMB_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb1XorLimbXorWcpArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb1XorLimbXorWcpArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroLimb2(final Bytes b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mmu.macro/LIMB_2 already set");
    } else {
      filled.set(60);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.macro/LIMB_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb2XorWcpArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb2XorWcpArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroPhase(final long b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmu.macro/PHASE already set");
    } else {
      filled.set(49);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.macro/PHASE has invalid value (" + b + ")"); }
    phaseXorExoSum.put((byte) (b >> 24));
    phaseXorExoSum.put((byte) (b >> 16));
    phaseXorExoSum.put((byte) (b >> 8));
    phaseXorExoSum.put((byte) b);


    return this;
  }

  public Trace pMacroRefOffset(final Bytes b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmu.macro/REF_OFFSET already set");
    } else {
      filled.set(53);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/REF_OFFSET has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refOffsetXorCnTXorEucB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refOffsetXorCnTXorEucB.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroRefSize(final Bytes b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mmu.macro/REF_SIZE already set");
    } else {
      filled.set(54);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/REF_SIZE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refSizeXorSloXorEucCeil.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refSizeXorSloXorEucCeil.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroSize(final Bytes b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mmu.macro/SIZE already set");
    } else {
      filled.set(55);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/SIZE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { sizeXorTloXorEucQuot.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { sizeXorTloXorEucQuot.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroSrcId(final Bytes b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmu.macro/SRC_ID already set");
    } else {
      filled.set(56);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/SRC_ID has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { srcIdXorTotalSizeXorEucRem.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { srcIdXorTotalSizeXorEucRem.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroSrcOffsetHi(final Bytes b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mmu.macro/SRC_OFFSET_HI already set");
    } else {
      filled.set(61);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.macro/SRC_OFFSET_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { srcOffsetHiXorWcpArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { srcOffsetHiXorWcpArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroSrcOffsetLo(final Bytes b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmu.macro/SRC_OFFSET_LO already set");
    } else {
      filled.set(62);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.macro/SRC_OFFSET_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { srcOffsetLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { srcOffsetLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroSuccessBit(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mmu.macro/SUCCESS_BIT already set");
    } else {
      filled.set(41);
    }

    successBitXorSuccessBitXorEucFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMacroTgtId(final Bytes b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mmu.macro/TGT_ID already set");
    } else {
      filled.set(57);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/TGT_ID has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { tgtId.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { tgtId.put(bs.get(j)); }

    return this;
  }

  public Trace pMacroTgtOffsetLo(final Bytes b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mmu.macro/TGT_OFFSET_LO already set");
    } else {
      filled.set(58);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.macro/TGT_OFFSET_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { tgtOffsetLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { tgtOffsetLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMicroCnS(final Bytes b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmu.micro/CN_S already set");
    } else {
      filled.set(52);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.micro/CN_S has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { auxIdXorCnSXorEucA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { auxIdXorCnSXorEucA.put(bs.get(j)); }

    return this;
  }

  public Trace pMicroCnT(final Bytes b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmu.micro/CN_T already set");
    } else {
      filled.set(53);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.micro/CN_T has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refOffsetXorCnTXorEucB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refOffsetXorCnTXorEucB.put(bs.get(j)); }

    return this;
  }

  public Trace pMicroExoId(final long b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmu.micro/EXO_ID already set");
    } else {
      filled.set(48);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.micro/EXO_ID has invalid value (" + b + ")"); }
    exoSumXorExoId.put((byte) (b >> 24));
    exoSumXorExoId.put((byte) (b >> 16));
    exoSumXorExoId.put((byte) (b >> 8));
    exoSumXorExoId.put((byte) b);


    return this;
  }

  public Trace pMicroExoSum(final long b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmu.micro/EXO_SUM already set");
    } else {
      filled.set(49);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.micro/EXO_SUM has invalid value (" + b + ")"); }
    phaseXorExoSum.put((byte) (b >> 24));
    phaseXorExoSum.put((byte) (b >> 16));
    phaseXorExoSum.put((byte) (b >> 8));
    phaseXorExoSum.put((byte) b);


    return this;
  }

  public Trace pMicroInst(final long b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.micro/INST already set");
    } else {
      filled.set(47);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("mmu.micro/INST has invalid value (" + b + ")"); }
    instXorInstXorCt.put((byte) (b >> 8));
    instXorInstXorCt.put((byte) b);


    return this;
  }

  public Trace pMicroKecId(final long b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mmu.micro/KEC_ID already set");
    } else {
      filled.set(50);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.micro/KEC_ID has invalid value (" + b + ")"); }
    kecId.put((byte) (b >> 24));
    kecId.put((byte) (b >> 16));
    kecId.put((byte) (b >> 8));
    kecId.put((byte) b);


    return this;
  }

  public Trace pMicroLimb(final Bytes b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmu.micro/LIMB already set");
    } else {
      filled.set(59);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.micro/LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb1XorLimbXorWcpArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb1XorLimbXorWcpArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMicroPhase(final long b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mmu.micro/PHASE already set");
    } else {
      filled.set(51);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.micro/PHASE has invalid value (" + b + ")"); }
    phase.put((byte) (b >> 24));
    phase.put((byte) (b >> 16));
    phase.put((byte) (b >> 8));
    phase.put((byte) b);


    return this;
  }

  public Trace pMicroSbo(final UnsignedByte b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mmu.micro/SBO already set");
    } else {
      filled.set(44);
    }

    sboXorWcpInst.put(b.toByte());

    return this;
  }

  public Trace pMicroSize(final UnsignedByte b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mmu.micro/SIZE already set");
    } else {
      filled.set(45);
    }

    size.put(b.toByte());

    return this;
  }

  public Trace pMicroSlo(final Bytes b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mmu.micro/SLO already set");
    } else {
      filled.set(54);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.micro/SLO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refSizeXorSloXorEucCeil.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refSizeXorSloXorEucCeil.put(bs.get(j)); }

    return this;
  }

  public Trace pMicroSuccessBit(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mmu.micro/SUCCESS_BIT already set");
    } else {
      filled.set(41);
    }

    successBitXorSuccessBitXorEucFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMicroTbo(final UnsignedByte b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mmu.micro/TBO already set");
    } else {
      filled.set(46);
    }

    tbo.put(b.toByte());

    return this;
  }

  public Trace pMicroTlo(final Bytes b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mmu.micro/TLO already set");
    } else {
      filled.set(55);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.micro/TLO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { sizeXorTloXorEucQuot.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { sizeXorTloXorEucQuot.put(bs.get(j)); }

    return this;
  }

  public Trace pMicroTotalSize(final Bytes b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmu.micro/TOTAL_SIZE already set");
    } else {
      filled.set(56);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.micro/TOTAL_SIZE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { srcIdXorTotalSizeXorEucRem.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { srcIdXorTotalSizeXorEucRem.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcCt(final long b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.prprc/CT already set");
    } else {
      filled.set(47);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("mmu.prprc/CT has invalid value (" + b + ")"); }
    instXorInstXorCt.put((byte) (b >> 8));
    instXorInstXorCt.put((byte) b);


    return this;
  }

  public Trace pPrprcEucA(final Bytes b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmu.prprc/EUC_A already set");
    } else {
      filled.set(52);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.prprc/EUC_A has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { auxIdXorCnSXorEucA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { auxIdXorCnSXorEucA.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcEucB(final Bytes b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmu.prprc/EUC_B already set");
    } else {
      filled.set(53);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.prprc/EUC_B has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refOffsetXorCnTXorEucB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refOffsetXorCnTXorEucB.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcEucCeil(final Bytes b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mmu.prprc/EUC_CEIL already set");
    } else {
      filled.set(54);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.prprc/EUC_CEIL has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refSizeXorSloXorEucCeil.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refSizeXorSloXorEucCeil.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcEucFlag(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mmu.prprc/EUC_FLAG already set");
    } else {
      filled.set(41);
    }

    successBitXorSuccessBitXorEucFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pPrprcEucQuot(final Bytes b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mmu.prprc/EUC_QUOT already set");
    } else {
      filled.set(55);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.prprc/EUC_QUOT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { sizeXorTloXorEucQuot.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { sizeXorTloXorEucQuot.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcEucRem(final Bytes b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmu.prprc/EUC_REM already set");
    } else {
      filled.set(56);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmu.prprc/EUC_REM has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { srcIdXorTotalSizeXorEucRem.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { srcIdXorTotalSizeXorEucRem.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcWcpArg1Hi(final Bytes b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_1_HI already set");
    } else {
      filled.set(59);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.prprc/WCP_ARG_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb1XorLimbXorWcpArg1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb1XorLimbXorWcpArg1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcWcpArg1Lo(final Bytes b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_1_LO already set");
    } else {
      filled.set(60);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.prprc/WCP_ARG_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb2XorWcpArg1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb2XorWcpArg1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcWcpArg2Lo(final Bytes b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_2_LO already set");
    } else {
      filled.set(61);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmu.prprc/WCP_ARG_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { srcOffsetHiXorWcpArg2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { srcOffsetHiXorWcpArg2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pPrprcWcpFlag(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mmu.prprc/WCP_FLAG already set");
    } else {
      filled.set(42);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pPrprcWcpInst(final UnsignedByte b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mmu.prprc/WCP_INST already set");
    } else {
      filled.set(44);
    }

    sboXorWcpInst.put(b.toByte());

    return this;
  }

  public Trace pPrprcWcpRes(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mmu.prprc/WCP_RES already set");
    } else {
      filled.set(43);
    }

    wcpRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prprc(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mmu.PRPRC already set");
    } else {
      filled.set(31);
    }

    prprc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rzFirst(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mmu.RZ_FIRST already set");
    } else {
      filled.set(32);
    }

    rzFirst.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rzLast(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mmu.RZ_LAST already set");
    } else {
      filled.set(33);
    }

    rzLast.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rzMddl(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mmu.RZ_MDDL already set");
    } else {
      filled.set(34);
    }

    rzMddl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rzOnly(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mmu.RZ_ONLY already set");
    } else {
      filled.set(35);
    }

    rzOnly.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mmu.STAMP already set");
    } else {
      filled.set(36);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 24));
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace tot(final long b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mmu.TOT already set");
    } else {
      filled.set(37);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.TOT has invalid value (" + b + ")"); }
    tot.put((byte) (b >> 24));
    tot.put((byte) (b >> 16));
    tot.put((byte) (b >> 8));
    tot.put((byte) b);


    return this;
  }

  public Trace totlz(final long b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mmu.TOTLZ already set");
    } else {
      filled.set(38);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.TOTLZ has invalid value (" + b + ")"); }
    totlz.put((byte) (b >> 24));
    totlz.put((byte) (b >> 16));
    totlz.put((byte) (b >> 8));
    totlz.put((byte) b);


    return this;
  }

  public Trace totnt(final long b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mmu.TOTNT already set");
    } else {
      filled.set(39);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.TOTNT has invalid value (" + b + ")"); }
    totnt.put((byte) (b >> 24));
    totnt.put((byte) (b >> 16));
    totnt.put((byte) (b >> 8));
    totnt.put((byte) b);


    return this;
  }

  public Trace totrz(final long b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mmu.TOTRZ already set");
    } else {
      filled.set(40);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmu.TOTRZ has invalid value (" + b + ")"); }
    totrz.put((byte) (b >> 24));
    totrz.put((byte) (b >> 16));
    totrz.put((byte) (b >> 8));
    totrz.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(52)) {
      throw new IllegalStateException("mmu.AUX_ID_xor_CN_S_xor_EUC_A has not been filled");
    }

    if (!filled.get(0)) {
      throw new IllegalStateException("mmu.BIN_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mmu.BIN_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mmu.BIN_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mmu.BIN_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mmu.BIN_5 has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mmu.EXO_SUM_xor_EXO_ID has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mmu.INST_xor_INST_xor_CT has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mmu.IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mmu.IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mmu.IS_BLAKE has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mmu.IS_EXO_TO_RAM_TRANSPLANTS has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mmu.IS_INVALID_CODE_PREFIX has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mmu.IS_MLOAD has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mmu.IS_MODEXP_DATA has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mmu.IS_MODEXP_ZERO has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mmu.IS_MSTORE has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mmu.IS_MSTORE8 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mmu.IS_RAM_TO_EXO_WITH_PADDING has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mmu.IS_RAM_TO_RAM_SANS_PADDING has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mmu.IS_RIGHT_PADDED_WORD_EXTRACTION has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mmu.KEC_ID has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("mmu.LIMB_1_xor_LIMB_xor_WCP_ARG_1_HI has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mmu.LIMB_2_xor_WCP_ARG_1_LO has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mmu.LZRO has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mmu.MACRO has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mmu.MICRO has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mmu.MMIO_STAMP has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mmu.NT_FIRST has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mmu.NT_LAST has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mmu.NT_MDDL has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mmu.NT_ONLY has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mmu.OUT_1 has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mmu.OUT_2 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mmu.OUT_3 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mmu.OUT_4 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mmu.OUT_5 has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mmu.PHASE has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mmu.PHASE_xor_EXO_SUM has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mmu.PRPRC has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("mmu.REF_OFFSET_xor_CN_T_xor_EUC_B has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mmu.REF_SIZE_xor_SLO_xor_EUC_CEIL has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mmu.RZ_FIRST has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mmu.RZ_LAST has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mmu.RZ_MDDL has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mmu.RZ_ONLY has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mmu.SBO_xor_WCP_INST has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mmu.SIZE has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mmu.SIZE_xor_TLO_xor_EUC_QUOT has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mmu.SRC_ID_xor_TOTAL_SIZE_xor_EUC_REM has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("mmu.SRC_OFFSET_HI_xor_WCP_ARG_2_LO has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("mmu.SRC_OFFSET_LO has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mmu.STAMP has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mmu.SUCCESS_BIT_xor_SUCCESS_BIT_xor_EUC_FLAG has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mmu.TBO has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mmu.TGT_ID has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mmu.TGT_OFFSET_LO has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mmu.TOT has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mmu.TOTLZ has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mmu.TOTNT has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mmu.TOTRZ has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mmu.WCP_FLAG has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mmu.WCP_RES has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(52)) {
      auxIdXorCnSXorEucA.position(auxIdXorCnSXorEucA.position() + 8);
    }

    if (!filled.get(0)) {
      bin1.position(bin1.position() + 1);
    }

    if (!filled.get(1)) {
      bin2.position(bin2.position() + 1);
    }

    if (!filled.get(2)) {
      bin3.position(bin3.position() + 1);
    }

    if (!filled.get(3)) {
      bin4.position(bin4.position() + 1);
    }

    if (!filled.get(4)) {
      bin5.position(bin5.position() + 1);
    }

    if (!filled.get(48)) {
      exoSumXorExoId.position(exoSumXorExoId.position() + 4);
    }

    if (!filled.get(47)) {
      instXorInstXorCt.position(instXorInstXorCt.position() + 2);
    }

    if (!filled.get(5)) {
      isAnyToRamWithPaddingPurePadding.position(isAnyToRamWithPaddingPurePadding.position() + 1);
    }

    if (!filled.get(6)) {
      isAnyToRamWithPaddingSomeData.position(isAnyToRamWithPaddingSomeData.position() + 1);
    }

    if (!filled.get(7)) {
      isBlake.position(isBlake.position() + 1);
    }

    if (!filled.get(8)) {
      isExoToRamTransplants.position(isExoToRamTransplants.position() + 1);
    }

    if (!filled.get(9)) {
      isInvalidCodePrefix.position(isInvalidCodePrefix.position() + 1);
    }

    if (!filled.get(10)) {
      isMload.position(isMload.position() + 1);
    }

    if (!filled.get(11)) {
      isModexpData.position(isModexpData.position() + 1);
    }

    if (!filled.get(12)) {
      isModexpZero.position(isModexpZero.position() + 1);
    }

    if (!filled.get(13)) {
      isMstore.position(isMstore.position() + 1);
    }

    if (!filled.get(14)) {
      isMstore8.position(isMstore8.position() + 1);
    }

    if (!filled.get(15)) {
      isRamToExoWithPadding.position(isRamToExoWithPadding.position() + 1);
    }

    if (!filled.get(16)) {
      isRamToRamSansPadding.position(isRamToRamSansPadding.position() + 1);
    }

    if (!filled.get(17)) {
      isRightPaddedWordExtraction.position(isRightPaddedWordExtraction.position() + 1);
    }

    if (!filled.get(50)) {
      kecId.position(kecId.position() + 4);
    }

    if (!filled.get(59)) {
      limb1XorLimbXorWcpArg1Hi.position(limb1XorLimbXorWcpArg1Hi.position() + 16);
    }

    if (!filled.get(60)) {
      limb2XorWcpArg1Lo.position(limb2XorWcpArg1Lo.position() + 16);
    }

    if (!filled.get(18)) {
      lzro.position(lzro.position() + 1);
    }

    if (!filled.get(19)) {
      macro.position(macro.position() + 1);
    }

    if (!filled.get(20)) {
      micro.position(micro.position() + 1);
    }

    if (!filled.get(21)) {
      mmioStamp.position(mmioStamp.position() + 4);
    }

    if (!filled.get(22)) {
      ntFirst.position(ntFirst.position() + 1);
    }

    if (!filled.get(23)) {
      ntLast.position(ntLast.position() + 1);
    }

    if (!filled.get(24)) {
      ntMddl.position(ntMddl.position() + 1);
    }

    if (!filled.get(25)) {
      ntOnly.position(ntOnly.position() + 1);
    }

    if (!filled.get(26)) {
      out1.position(out1.position() + 8);
    }

    if (!filled.get(27)) {
      out2.position(out2.position() + 8);
    }

    if (!filled.get(28)) {
      out3.position(out3.position() + 8);
    }

    if (!filled.get(29)) {
      out4.position(out4.position() + 8);
    }

    if (!filled.get(30)) {
      out5.position(out5.position() + 8);
    }

    if (!filled.get(51)) {
      phase.position(phase.position() + 4);
    }

    if (!filled.get(49)) {
      phaseXorExoSum.position(phaseXorExoSum.position() + 4);
    }

    if (!filled.get(31)) {
      prprc.position(prprc.position() + 1);
    }

    if (!filled.get(53)) {
      refOffsetXorCnTXorEucB.position(refOffsetXorCnTXorEucB.position() + 8);
    }

    if (!filled.get(54)) {
      refSizeXorSloXorEucCeil.position(refSizeXorSloXorEucCeil.position() + 8);
    }

    if (!filled.get(32)) {
      rzFirst.position(rzFirst.position() + 1);
    }

    if (!filled.get(33)) {
      rzLast.position(rzLast.position() + 1);
    }

    if (!filled.get(34)) {
      rzMddl.position(rzMddl.position() + 1);
    }

    if (!filled.get(35)) {
      rzOnly.position(rzOnly.position() + 1);
    }

    if (!filled.get(44)) {
      sboXorWcpInst.position(sboXorWcpInst.position() + 1);
    }

    if (!filled.get(45)) {
      size.position(size.position() + 1);
    }

    if (!filled.get(55)) {
      sizeXorTloXorEucQuot.position(sizeXorTloXorEucQuot.position() + 8);
    }

    if (!filled.get(56)) {
      srcIdXorTotalSizeXorEucRem.position(srcIdXorTotalSizeXorEucRem.position() + 8);
    }

    if (!filled.get(61)) {
      srcOffsetHiXorWcpArg2Lo.position(srcOffsetHiXorWcpArg2Lo.position() + 16);
    }

    if (!filled.get(62)) {
      srcOffsetLo.position(srcOffsetLo.position() + 16);
    }

    if (!filled.get(36)) {
      stamp.position(stamp.position() + 4);
    }

    if (!filled.get(41)) {
      successBitXorSuccessBitXorEucFlag.position(successBitXorSuccessBitXorEucFlag.position() + 1);
    }

    if (!filled.get(46)) {
      tbo.position(tbo.position() + 1);
    }

    if (!filled.get(57)) {
      tgtId.position(tgtId.position() + 8);
    }

    if (!filled.get(58)) {
      tgtOffsetLo.position(tgtOffsetLo.position() + 8);
    }

    if (!filled.get(37)) {
      tot.position(tot.position() + 4);
    }

    if (!filled.get(38)) {
      totlz.position(totlz.position() + 4);
    }

    if (!filled.get(39)) {
      totnt.position(totnt.position() + 4);
    }

    if (!filled.get(40)) {
      totrz.position(totrz.position() + 4);
    }

    if (!filled.get(42)) {
      wcpFlag.position(wcpFlag.position() + 1);
    }

    if (!filled.get(43)) {
      wcpRes.position(wcpRes.position() + 1);
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
