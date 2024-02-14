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
  public static final int EQ_ = 20;
  public static final int INVALID_CODE_PREFIX_VALUE = 239;
  public static final int ISZERO = 21;
  public static final int LLARGE = 16;
  public static final int LLARGEMO = 15;
  public static final int LLARGEPO = 17;
  public static final int LT = 16;
  public static final int MMIO_INST_EXO_LIMB_VANISHES = 65280;
  public static final int MMIO_INST_EXO_TO_RAM_LIMB_TRANSPLANT = 65312;
  public static final int MMIO_INST_EXO_TO_RAM_SLIDE_CHUNK = 65313;
  public static final int MMIO_INST_EXO_TO_RAM_SLIDE_OVERLAPPING_CHUNK = 65314;
  public static final int MMIO_INST_LIMB_TO_RAM_OVERLAP = 65330;
  public static final int MMIO_INST_LIMB_TO_RAM_TRANSPLANT = 65329;
  public static final int MMIO_INST_LIMB_TO_RAM_WRITE_LSB = 65328;
  public static final int MMIO_INST_PADDED_EXO_FROM_ONE_RAM = 65298;
  public static final int MMIO_INST_PADDED_EXO_FROM_TWO_RAM = 65299;
  public static final int MMIO_INST_PADDED_LIMB_FROM_ONE_RAM = 65345;
  public static final int MMIO_INST_PADDED_LIMB_FROM_TWO_RAM = 65346;
  public static final int MMIO_INST_RAM_EXCISION = 65376;
  public static final int MMIO_INST_RAM_LIMB_VANISHES = 65377;
  public static final int MMIO_INST_RAM_TO_EXO_LIMB_TRANSPLANT = 65296;
  public static final int MMIO_INST_RAM_TO_LIMB_TRANSPLANT = 65344;
  public static final int MMIO_INST_RAM_TO_RAM_LIMB_TRANSPLANT = 65360;
  public static final int MMIO_INST_RAM_TO_RAM_SLIDE_CHUNK = 65361;
  public static final int MMIO_INST_RAM_TO_RAM_SLIDE_OVERLAPPING_CHUNK = 65362;
  public static final int MMIO_INST_TWO_RAM_TO_EXO_FULL = 65297;
  public static final int MMU_INST_ANY_TO_RAM_WITH_PADDING = 65104;
  public static final int MMU_INST_ANY_TO_RAM_WITH_PADDING_PURE_PADDING = 65106;
  public static final int MMU_INST_ANY_TO_RAM_WITH_PADDING_SOME_DATA = 65105;
  public static final int MMU_INST_BLAKE = 65152;
  public static final int MMU_INST_EXO_TO_RAM_TRANSPLANTS = 65072;
  public static final int MMU_INST_INVALID_CODE_PREFIX = 65024;
  public static final int MMU_INST_MLOAD = 65025;
  public static final int MMU_INST_MODEXP_DATA = 65136;
  public static final int MMU_INST_MODEXP_ZERO = 65120;
  public static final int MMU_INST_MSTORE = 65026;
  public static final int MMU_INST_MSTORE8 = 83;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_BLAKE_PARAM = 2;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_INVALID_CODE_PREFIX = 1;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_MLOAD = 2;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_MODEXP_DATA = 32;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_MODEXP_ZERO = 32;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_MSTORE = 2;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_MSTORE_EIGHT = 1;
  public static final int MMU_INST_NB_MICRO_ROWS_TOT_RIGHT_PADDED_WORD_EXTRACTION = 2;
  public static final int MMU_INST_NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING = 4;
  public static final int MMU_INST_NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING_PO = 5;
  public static final int MMU_INST_NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA = 1;
  public static final int MMU_INST_NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_BLAKE_PARAM = 2;
  public static final int MMU_INST_NB_PP_ROWS_BLAKE_PARAM_PO = 3;
  public static final int MMU_INST_NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS = 1;
  public static final int MMU_INST_NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_INVALID_CODE_PREFIX = 1;
  public static final int MMU_INST_NB_PP_ROWS_INVALID_CODE_PREFIX_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_MLOAD = 1;
  public static final int MMU_INST_NB_PP_ROWS_MLOAD_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_MLOAD_PT = 3;
  public static final int MMU_INST_NB_PP_ROWS_MODEXP_DATA = 6;
  public static final int MMU_INST_NB_PP_ROWS_MODEXP_DATA_PO = 7;
  public static final int MMU_INST_NB_PP_ROWS_MODEXP_ZERO = 1;
  public static final int MMU_INST_NB_PP_ROWS_MODEXP_ZERO_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_MSTORE = 1;
  public static final int MMU_INST_NB_PP_ROWS_MSTORE8 = 1;
  public static final int MMU_INST_NB_PP_ROWS_MSTORE8_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_MSTORE_PO = 2;
  public static final int MMU_INST_NB_PP_ROWS_MSTORE_PT = 3;
  public static final int MMU_INST_NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING = 4;
  public static final int MMU_INST_NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING_PO = 5;
  public static final int MMU_INST_NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING = 5;
  public static final int MMU_INST_NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING_PO = 6;
  public static final int MMU_INST_NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION = 5;
  public static final int MMU_INST_NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION_PO = 6;
  public static final int MMU_INST_NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION_PT = 7;
  public static final int MMU_INST_RAM_TO_EXO_WITH_PADDING = 65056;
  public static final int MMU_INST_RAM_TO_RAM_SANS_PADDING = 65088;
  public static final int MMU_INST_RIGHT_PADDED_WORD_EXTRACTION = 65040;
  public static final int WORD_SIZE = 32;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer auxIdXorCnS;
  private final MappedByteBuffer bin1;
  private final MappedByteBuffer bin2;
  private final MappedByteBuffer bin3;
  private final MappedByteBuffer bin4;
  private final MappedByteBuffer bin5;
  private final MappedByteBuffer eucRem;
  private final MappedByteBuffer exoSumXorCnT;
  private final MappedByteBuffer instXorInstXorCt;
  private final MappedByteBuffer isAnyToRamWithPaddingPurePadding;
  private final MappedByteBuffer isAnyToRamWithPaddingSomeData;
  private final MappedByteBuffer isBlakeParam;
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
  private final MappedByteBuffer refOffsetXorEucA;
  private final MappedByteBuffer refSizeXorEucB;
  private final MappedByteBuffer rzFirst;
  private final MappedByteBuffer rzLast;
  private final MappedByteBuffer rzMddl;
  private final MappedByteBuffer rzOnly;
  private final MappedByteBuffer sboXorWcpInst;
  private final MappedByteBuffer size;
  private final MappedByteBuffer sizeXorEucCeil;
  private final MappedByteBuffer slo;
  private final MappedByteBuffer srcIdXorId1;
  private final MappedByteBuffer srcOffsetHiXorWcpArg2Hi;
  private final MappedByteBuffer srcOffsetLoXorWcpArg2Lo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer successBitXorSuccessBitXorEucFlag;
  private final MappedByteBuffer tbo;
  private final MappedByteBuffer tgtIdXorId2;
  private final MappedByteBuffer tgtOffsetLoXorEucQuot;
  private final MappedByteBuffer tlo;
  private final MappedByteBuffer tot;
  private final MappedByteBuffer totalSize;
  private final MappedByteBuffer totlz;
  private final MappedByteBuffer totnt;
  private final MappedByteBuffer totrz;
  private final MappedByteBuffer wcpFlag;
  private final MappedByteBuffer wcpRes;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("mmu.AUX_ID_xor_CN_S", 32, length),
        new ColumnHeader("mmu.BIN_1", 1, length),
        new ColumnHeader("mmu.BIN_2", 1, length),
        new ColumnHeader("mmu.BIN_3", 1, length),
        new ColumnHeader("mmu.BIN_4", 1, length),
        new ColumnHeader("mmu.BIN_5", 1, length),
        new ColumnHeader("mmu.EUC_REM", 32, length),
        new ColumnHeader("mmu.EXO_SUM_xor_CN_T", 32, length),
        new ColumnHeader("mmu.INST_xor_INST_xor_CT", 32, length),
        new ColumnHeader("mmu.IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING", 1, length),
        new ColumnHeader("mmu.IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA", 1, length),
        new ColumnHeader("mmu.IS_BLAKE_PARAM", 1, length),
        new ColumnHeader("mmu.IS_EXO_TO_RAM_TRANSPLANTS", 1, length),
        new ColumnHeader("mmu.IS_INVALID_CODE_PREFIX", 1, length),
        new ColumnHeader("mmu.IS_MLOAD", 1, length),
        new ColumnHeader("mmu.IS_MODEXP_DATA", 1, length),
        new ColumnHeader("mmu.IS_MODEXP_ZERO", 1, length),
        new ColumnHeader("mmu.IS_MSTORE", 1, length),
        new ColumnHeader("mmu.IS_MSTORE8", 1, length),
        new ColumnHeader("mmu.IS_RAM_TO_EXO_WITH_PADDING", 1, length),
        new ColumnHeader("mmu.IS_RAM_TO_RAM_SANS_PADDING", 1, length),
        new ColumnHeader("mmu.IS_RIGHT_PADDED_WORD_EXTRACTION", 1, length),
        new ColumnHeader("mmu.LIMB_1_xor_LIMB_xor_WCP_ARG_1_HI", 32, length),
        new ColumnHeader("mmu.LIMB_2_xor_WCP_ARG_1_LO", 32, length),
        new ColumnHeader("mmu.LZRO", 1, length),
        new ColumnHeader("mmu.MACRO", 1, length),
        new ColumnHeader("mmu.MICRO", 1, length),
        new ColumnHeader("mmu.MMIO_STAMP", 32, length),
        new ColumnHeader("mmu.NT_FIRST", 1, length),
        new ColumnHeader("mmu.NT_LAST", 1, length),
        new ColumnHeader("mmu.NT_MDDL", 1, length),
        new ColumnHeader("mmu.NT_ONLY", 1, length),
        new ColumnHeader("mmu.OUT_1", 32, length),
        new ColumnHeader("mmu.OUT_2", 32, length),
        new ColumnHeader("mmu.OUT_3", 32, length),
        new ColumnHeader("mmu.OUT_4", 32, length),
        new ColumnHeader("mmu.OUT_5", 32, length),
        new ColumnHeader("mmu.PHASE", 32, length),
        new ColumnHeader("mmu.PHASE_xor_EXO_SUM", 32, length),
        new ColumnHeader("mmu.PRPRC", 1, length),
        new ColumnHeader("mmu.REF_OFFSET_xor_EUC_A", 32, length),
        new ColumnHeader("mmu.REF_SIZE_xor_EUC_B", 32, length),
        new ColumnHeader("mmu.RZ_FIRST", 1, length),
        new ColumnHeader("mmu.RZ_LAST", 1, length),
        new ColumnHeader("mmu.RZ_MDDL", 1, length),
        new ColumnHeader("mmu.RZ_ONLY", 1, length),
        new ColumnHeader("mmu.SBO_xor_WCP_INST", 1, length),
        new ColumnHeader("mmu.SIZE", 1, length),
        new ColumnHeader("mmu.SIZE_xor_EUC_CEIL", 32, length),
        new ColumnHeader("mmu.SLO", 32, length),
        new ColumnHeader("mmu.SRC_ID_xor_ID_1", 32, length),
        new ColumnHeader("mmu.SRC_OFFSET_HI_xor_WCP_ARG_2_HI", 32, length),
        new ColumnHeader("mmu.SRC_OFFSET_LO_xor_WCP_ARG_2_LO", 32, length),
        new ColumnHeader("mmu.STAMP", 32, length),
        new ColumnHeader("mmu.SUCCESS_BIT_xor_SUCCESS_BIT_xor_EUC_FLAG", 1, length),
        new ColumnHeader("mmu.TBO", 1, length),
        new ColumnHeader("mmu.TGT_ID_xor_ID_2", 32, length),
        new ColumnHeader("mmu.TGT_OFFSET_LO_xor_EUC_QUOT", 32, length),
        new ColumnHeader("mmu.TLO", 32, length),
        new ColumnHeader("mmu.TOT", 32, length),
        new ColumnHeader("mmu.TOTAL_SIZE", 32, length),
        new ColumnHeader("mmu.TOTLZ", 32, length),
        new ColumnHeader("mmu.TOTNT", 32, length),
        new ColumnHeader("mmu.TOTRZ", 32, length),
        new ColumnHeader("mmu.WCP_FLAG", 1, length),
        new ColumnHeader("mmu.WCP_RES", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.auxIdXorCnS = buffers.get(0);
    this.bin1 = buffers.get(1);
    this.bin2 = buffers.get(2);
    this.bin3 = buffers.get(3);
    this.bin4 = buffers.get(4);
    this.bin5 = buffers.get(5);
    this.eucRem = buffers.get(6);
    this.exoSumXorCnT = buffers.get(7);
    this.instXorInstXorCt = buffers.get(8);
    this.isAnyToRamWithPaddingPurePadding = buffers.get(9);
    this.isAnyToRamWithPaddingSomeData = buffers.get(10);
    this.isBlakeParam = buffers.get(11);
    this.isExoToRamTransplants = buffers.get(12);
    this.isInvalidCodePrefix = buffers.get(13);
    this.isMload = buffers.get(14);
    this.isModexpData = buffers.get(15);
    this.isModexpZero = buffers.get(16);
    this.isMstore = buffers.get(17);
    this.isMstore8 = buffers.get(18);
    this.isRamToExoWithPadding = buffers.get(19);
    this.isRamToRamSansPadding = buffers.get(20);
    this.isRightPaddedWordExtraction = buffers.get(21);
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
    this.refOffsetXorEucA = buffers.get(40);
    this.refSizeXorEucB = buffers.get(41);
    this.rzFirst = buffers.get(42);
    this.rzLast = buffers.get(43);
    this.rzMddl = buffers.get(44);
    this.rzOnly = buffers.get(45);
    this.sboXorWcpInst = buffers.get(46);
    this.size = buffers.get(47);
    this.sizeXorEucCeil = buffers.get(48);
    this.slo = buffers.get(49);
    this.srcIdXorId1 = buffers.get(50);
    this.srcOffsetHiXorWcpArg2Hi = buffers.get(51);
    this.srcOffsetLoXorWcpArg2Lo = buffers.get(52);
    this.stamp = buffers.get(53);
    this.successBitXorSuccessBitXorEucFlag = buffers.get(54);
    this.tbo = buffers.get(55);
    this.tgtIdXorId2 = buffers.get(56);
    this.tgtOffsetLoXorEucQuot = buffers.get(57);
    this.tlo = buffers.get(58);
    this.tot = buffers.get(59);
    this.totalSize = buffers.get(60);
    this.totlz = buffers.get(61);
    this.totnt = buffers.get(62);
    this.totrz = buffers.get(63);
    this.wcpFlag = buffers.get(64);
    this.wcpRes = buffers.get(65);
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

  public Trace isBlakeParam(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mmu.IS_BLAKE_PARAM already set");
    } else {
      filled.set(7);
    }

    isBlakeParam.put((byte) (b ? 1 : 0));

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

  public Trace mmioStamp(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mmu.MMIO_STAMP already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmioStamp.put((byte) 0);
    }
    mmioStamp.put(b.toArrayUnsafe());

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

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      out1.put((byte) 0);
    }
    out1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace out2(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mmu.OUT_2 already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      out2.put((byte) 0);
    }
    out2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace out3(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mmu.OUT_3 already set");
    } else {
      filled.set(28);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      out3.put((byte) 0);
    }
    out3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace out4(final Bytes b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mmu.OUT_4 already set");
    } else {
      filled.set(29);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      out4.put((byte) 0);
    }
    out4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace out5(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mmu.OUT_5 already set");
    } else {
      filled.set(30);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      out5.put((byte) 0);
    }
    out5.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroAuxId(final Bytes b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmu.macro/AUX_ID already set");
    } else {
      filled.set(48);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      auxIdXorCnS.put((byte) 0);
    }
    auxIdXorCnS.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroExoSum(final Bytes b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmu.macro/EXO_SUM already set");
    } else {
      filled.set(49);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      exoSumXorCnT.put((byte) 0);
    }
    exoSumXorCnT.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroInst(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.macro/INST already set");
    } else {
      filled.set(47);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      instXorInstXorCt.put((byte) 0);
    }
    instXorInstXorCt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroLimb1(final Bytes b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmu.macro/LIMB_1 already set");
    } else {
      filled.set(62);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb1XorLimbXorWcpArg1Hi.put((byte) 0);
    }
    limb1XorLimbXorWcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroLimb2(final Bytes b) {
    if (filled.get(63)) {
      throw new IllegalStateException("mmu.macro/LIMB_2 already set");
    } else {
      filled.set(63);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb2XorWcpArg1Lo.put((byte) 0);
    }
    limb2XorWcpArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroPhase(final Bytes b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mmu.macro/PHASE already set");
    } else {
      filled.set(50);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      phaseXorExoSum.put((byte) 0);
    }
    phaseXorExoSum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroRefOffset(final Bytes b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mmu.macro/REF_OFFSET already set");
    } else {
      filled.set(57);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refOffsetXorEucA.put((byte) 0);
    }
    refOffsetXorEucA.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroRefSize(final Bytes b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mmu.macro/REF_SIZE already set");
    } else {
      filled.set(58);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refSizeXorEucB.put((byte) 0);
    }
    refSizeXorEucB.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroSize(final Bytes b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmu.macro/SIZE already set");
    } else {
      filled.set(59);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      sizeXorEucCeil.put((byte) 0);
    }
    sizeXorEucCeil.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroSrcId(final Bytes b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mmu.macro/SRC_ID already set");
    } else {
      filled.set(51);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      srcIdXorId1.put((byte) 0);
    }
    srcIdXorId1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroSrcOffsetHi(final Bytes b) {
    if (filled.get(64)) {
      throw new IllegalStateException("mmu.macro/SRC_OFFSET_HI already set");
    } else {
      filled.set(64);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      srcOffsetHiXorWcpArg2Hi.put((byte) 0);
    }
    srcOffsetHiXorWcpArg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroSrcOffsetLo(final Bytes b) {
    if (filled.get(65)) {
      throw new IllegalStateException("mmu.macro/SRC_OFFSET_LO already set");
    } else {
      filled.set(65);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      srcOffsetLoXorWcpArg2Lo.put((byte) 0);
    }
    srcOffsetLoXorWcpArg2Lo.put(b.toArrayUnsafe());

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
    if (filled.get(52)) {
      throw new IllegalStateException("mmu.macro/TGT_ID already set");
    } else {
      filled.set(52);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      tgtIdXorId2.put((byte) 0);
    }
    tgtIdXorId2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroTgtOffsetLo(final Bytes b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mmu.macro/TGT_OFFSET_LO already set");
    } else {
      filled.set(60);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      tgtOffsetLoXorEucQuot.put((byte) 0);
    }
    tgtOffsetLoXorEucQuot.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroCnS(final Bytes b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmu.micro/CN_S already set");
    } else {
      filled.set(48);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      auxIdXorCnS.put((byte) 0);
    }
    auxIdXorCnS.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroCnT(final Bytes b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmu.micro/CN_T already set");
    } else {
      filled.set(49);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      exoSumXorCnT.put((byte) 0);
    }
    exoSumXorCnT.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroExoSum(final Bytes b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mmu.micro/EXO_SUM already set");
    } else {
      filled.set(50);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      phaseXorExoSum.put((byte) 0);
    }
    phaseXorExoSum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroId1(final Bytes b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mmu.micro/ID_1 already set");
    } else {
      filled.set(51);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      srcIdXorId1.put((byte) 0);
    }
    srcIdXorId1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroId2(final Bytes b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmu.micro/ID_2 already set");
    } else {
      filled.set(52);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      tgtIdXorId2.put((byte) 0);
    }
    tgtIdXorId2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroInst(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.micro/INST already set");
    } else {
      filled.set(47);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      instXorInstXorCt.put((byte) 0);
    }
    instXorInstXorCt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroLimb(final Bytes b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmu.micro/LIMB already set");
    } else {
      filled.set(62);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb1XorLimbXorWcpArg1Hi.put((byte) 0);
    }
    limb1XorLimbXorWcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroPhase(final Bytes b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmu.micro/PHASE already set");
    } else {
      filled.set(53);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      phase.put((byte) 0);
    }
    phase.put(b.toArrayUnsafe());

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

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      slo.put((byte) 0);
    }
    slo.put(b.toArrayUnsafe());

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

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      tlo.put((byte) 0);
    }
    tlo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMicroTotalSize(final Bytes b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmu.micro/TOTAL_SIZE already set");
    } else {
      filled.set(56);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      totalSize.put((byte) 0);
    }
    totalSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcCt(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmu.prprc/CT already set");
    } else {
      filled.set(47);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      instXorInstXorCt.put((byte) 0);
    }
    instXorInstXorCt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcEucA(final Bytes b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mmu.prprc/EUC_A already set");
    } else {
      filled.set(57);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refOffsetXorEucA.put((byte) 0);
    }
    refOffsetXorEucA.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcEucB(final Bytes b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mmu.prprc/EUC_B already set");
    } else {
      filled.set(58);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refSizeXorEucB.put((byte) 0);
    }
    refSizeXorEucB.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcEucCeil(final Bytes b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmu.prprc/EUC_CEIL already set");
    } else {
      filled.set(59);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      sizeXorEucCeil.put((byte) 0);
    }
    sizeXorEucCeil.put(b.toArrayUnsafe());

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
    if (filled.get(60)) {
      throw new IllegalStateException("mmu.prprc/EUC_QUOT already set");
    } else {
      filled.set(60);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      tgtOffsetLoXorEucQuot.put((byte) 0);
    }
    tgtOffsetLoXorEucQuot.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcEucRem(final Bytes b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mmu.prprc/EUC_REM already set");
    } else {
      filled.set(61);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      eucRem.put((byte) 0);
    }
    eucRem.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcWcpArg1Hi(final Bytes b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_1_HI already set");
    } else {
      filled.set(62);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb1XorLimbXorWcpArg1Hi.put((byte) 0);
    }
    limb1XorLimbXorWcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcWcpArg1Lo(final Bytes b) {
    if (filled.get(63)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_1_LO already set");
    } else {
      filled.set(63);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb2XorWcpArg1Lo.put((byte) 0);
    }
    limb2XorWcpArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcWcpArg2Hi(final Bytes b) {
    if (filled.get(64)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_2_HI already set");
    } else {
      filled.set(64);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      srcOffsetHiXorWcpArg2Hi.put((byte) 0);
    }
    srcOffsetHiXorWcpArg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPrprcWcpArg2Lo(final Bytes b) {
    if (filled.get(65)) {
      throw new IllegalStateException("mmu.prprc/WCP_ARG_2_LO already set");
    } else {
      filled.set(65);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      srcOffsetLoXorWcpArg2Lo.put((byte) 0);
    }
    srcOffsetLoXorWcpArg2Lo.put(b.toArrayUnsafe());

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

  public Trace stamp(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mmu.STAMP already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace tot(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mmu.TOT already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      tot.put((byte) 0);
    }
    tot.put(b.toArrayUnsafe());

    return this;
  }

  public Trace totlz(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mmu.TOTLZ already set");
    } else {
      filled.set(38);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      totlz.put((byte) 0);
    }
    totlz.put(b.toArrayUnsafe());

    return this;
  }

  public Trace totnt(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mmu.TOTNT already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      totnt.put((byte) 0);
    }
    totnt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace totrz(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mmu.TOTRZ already set");
    } else {
      filled.set(40);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      totrz.put((byte) 0);
    }
    totrz.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(48)) {
      throw new IllegalStateException("mmu.AUX_ID_xor_CN_S has not been filled");
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

    if (!filled.get(61)) {
      throw new IllegalStateException("mmu.EUC_REM has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mmu.EXO_SUM_xor_CN_T has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mmu.INST_xor_INST_xor_CT has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException(
          "mmu.IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException(
          "mmu.IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mmu.IS_BLAKE_PARAM has not been filled");
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

    if (!filled.get(62)) {
      throw new IllegalStateException("mmu.LIMB_1_xor_LIMB_xor_WCP_ARG_1_HI has not been filled");
    }

    if (!filled.get(63)) {
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

    if (!filled.get(53)) {
      throw new IllegalStateException("mmu.PHASE has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mmu.PHASE_xor_EXO_SUM has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mmu.PRPRC has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mmu.REF_OFFSET_xor_EUC_A has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mmu.REF_SIZE_xor_EUC_B has not been filled");
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

    if (!filled.get(59)) {
      throw new IllegalStateException("mmu.SIZE_xor_EUC_CEIL has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mmu.SLO has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mmu.SRC_ID_xor_ID_1 has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("mmu.SRC_OFFSET_HI_xor_WCP_ARG_2_HI has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("mmu.SRC_OFFSET_LO_xor_WCP_ARG_2_LO has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mmu.STAMP has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException(
          "mmu.SUCCESS_BIT_xor_SUCCESS_BIT_xor_EUC_FLAG has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mmu.TBO has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("mmu.TGT_ID_xor_ID_2 has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mmu.TGT_OFFSET_LO_xor_EUC_QUOT has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mmu.TLO has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mmu.TOT has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mmu.TOTAL_SIZE has not been filled");
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
    if (!filled.get(48)) {
      auxIdXorCnS.position(auxIdXorCnS.position() + 32);
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

    if (!filled.get(61)) {
      eucRem.position(eucRem.position() + 32);
    }

    if (!filled.get(49)) {
      exoSumXorCnT.position(exoSumXorCnT.position() + 32);
    }

    if (!filled.get(47)) {
      instXorInstXorCt.position(instXorInstXorCt.position() + 32);
    }

    if (!filled.get(5)) {
      isAnyToRamWithPaddingPurePadding.position(isAnyToRamWithPaddingPurePadding.position() + 1);
    }

    if (!filled.get(6)) {
      isAnyToRamWithPaddingSomeData.position(isAnyToRamWithPaddingSomeData.position() + 1);
    }

    if (!filled.get(7)) {
      isBlakeParam.position(isBlakeParam.position() + 1);
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

    if (!filled.get(62)) {
      limb1XorLimbXorWcpArg1Hi.position(limb1XorLimbXorWcpArg1Hi.position() + 32);
    }

    if (!filled.get(63)) {
      limb2XorWcpArg1Lo.position(limb2XorWcpArg1Lo.position() + 32);
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
      mmioStamp.position(mmioStamp.position() + 32);
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
      out1.position(out1.position() + 32);
    }

    if (!filled.get(27)) {
      out2.position(out2.position() + 32);
    }

    if (!filled.get(28)) {
      out3.position(out3.position() + 32);
    }

    if (!filled.get(29)) {
      out4.position(out4.position() + 32);
    }

    if (!filled.get(30)) {
      out5.position(out5.position() + 32);
    }

    if (!filled.get(53)) {
      phase.position(phase.position() + 32);
    }

    if (!filled.get(50)) {
      phaseXorExoSum.position(phaseXorExoSum.position() + 32);
    }

    if (!filled.get(31)) {
      prprc.position(prprc.position() + 1);
    }

    if (!filled.get(57)) {
      refOffsetXorEucA.position(refOffsetXorEucA.position() + 32);
    }

    if (!filled.get(58)) {
      refSizeXorEucB.position(refSizeXorEucB.position() + 32);
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

    if (!filled.get(59)) {
      sizeXorEucCeil.position(sizeXorEucCeil.position() + 32);
    }

    if (!filled.get(54)) {
      slo.position(slo.position() + 32);
    }

    if (!filled.get(51)) {
      srcIdXorId1.position(srcIdXorId1.position() + 32);
    }

    if (!filled.get(64)) {
      srcOffsetHiXorWcpArg2Hi.position(srcOffsetHiXorWcpArg2Hi.position() + 32);
    }

    if (!filled.get(65)) {
      srcOffsetLoXorWcpArg2Lo.position(srcOffsetLoXorWcpArg2Lo.position() + 32);
    }

    if (!filled.get(36)) {
      stamp.position(stamp.position() + 32);
    }

    if (!filled.get(41)) {
      successBitXorSuccessBitXorEucFlag.position(successBitXorSuccessBitXorEucFlag.position() + 1);
    }

    if (!filled.get(46)) {
      tbo.position(tbo.position() + 1);
    }

    if (!filled.get(52)) {
      tgtIdXorId2.position(tgtIdXorId2.position() + 32);
    }

    if (!filled.get(60)) {
      tgtOffsetLoXorEucQuot.position(tgtOffsetLoXorEucQuot.position() + 32);
    }

    if (!filled.get(55)) {
      tlo.position(tlo.position() + 32);
    }

    if (!filled.get(37)) {
      tot.position(tot.position() + 32);
    }

    if (!filled.get(56)) {
      totalSize.position(totalSize.position() + 32);
    }

    if (!filled.get(38)) {
      totlz.position(totlz.position() + 32);
    }

    if (!filled.get(39)) {
      totnt.position(totnt.position() + 32);
    }

    if (!filled.get(40)) {
      totrz.position(totrz.position() + 32);
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
