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

package net.consensys.linea.zktracer.module.mxp;

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
  public static final int BLAKE_PHASE_DATA = 0x5;
  public static final int BLAKE_PHASE_PARAMS = 0x6;
  public static final int BLAKE_PHASE_RESULT = 0x7;
  public static final int EVM_INST_ADD = 0x1;
  public static final int EVM_INST_ADDMOD = 0x8;
  public static final int EVM_INST_ADDRESS = 0x30;
  public static final int EVM_INST_AND = 0x16;
  public static final int EVM_INST_BALANCE = 0x31;
  public static final int EVM_INST_BASEFEE = 0x48;
  public static final int EVM_INST_BLOCKHASH = 0x40;
  public static final int EVM_INST_BYTE = 0x1a;
  public static final int EVM_INST_CALL = 0xf1;
  public static final int EVM_INST_CALLCODE = 0xf2;
  public static final int EVM_INST_CALLDATACOPY = 0x37;
  public static final int EVM_INST_CALLDATALOAD = 0x35;
  public static final int EVM_INST_CALLDATASIZE = 0x36;
  public static final int EVM_INST_CALLER = 0x33;
  public static final int EVM_INST_CALLVALUE = 0x34;
  public static final int EVM_INST_CHAINID = 0x46;
  public static final int EVM_INST_CODECOPY = 0x39;
  public static final int EVM_INST_CODESIZE = 0x38;
  public static final int EVM_INST_COINBASE = 0x41;
  public static final int EVM_INST_CREATE = 0xf0;
  public static final int EVM_INST_CREATE2 = 0xf5;
  public static final int EVM_INST_DELEGATECALL = 0xf4;
  public static final int EVM_INST_DIFFICULTY = 0x44;
  public static final int EVM_INST_DIV = 0x4;
  public static final int EVM_INST_DUP1 = 0x80;
  public static final int EVM_INST_DUP10 = 0x89;
  public static final int EVM_INST_DUP11 = 0x8a;
  public static final int EVM_INST_DUP12 = 0x8b;
  public static final int EVM_INST_DUP13 = 0x8c;
  public static final int EVM_INST_DUP14 = 0x8d;
  public static final int EVM_INST_DUP15 = 0x8e;
  public static final int EVM_INST_DUP16 = 0x8f;
  public static final int EVM_INST_DUP2 = 0x81;
  public static final int EVM_INST_DUP3 = 0x82;
  public static final int EVM_INST_DUP4 = 0x83;
  public static final int EVM_INST_DUP5 = 0x84;
  public static final int EVM_INST_DUP6 = 0x85;
  public static final int EVM_INST_DUP7 = 0x86;
  public static final int EVM_INST_DUP8 = 0x87;
  public static final int EVM_INST_DUP9 = 0x88;
  public static final int EVM_INST_EQ = 0x14;
  public static final int EVM_INST_EXP = 0xa;
  public static final int EVM_INST_EXTCODECOPY = 0x3c;
  public static final int EVM_INST_EXTCODEHASH = 0x3f;
  public static final int EVM_INST_EXTCODESIZE = 0x3b;
  public static final int EVM_INST_GAS = 0x5a;
  public static final int EVM_INST_GASLIMIT = 0x45;
  public static final int EVM_INST_GASPRICE = 0x3a;
  public static final int EVM_INST_GT = 0x11;
  public static final int EVM_INST_INVALID = 0xfe;
  public static final int EVM_INST_ISZERO = 0x15;
  public static final int EVM_INST_JUMP = 0x56;
  public static final int EVM_INST_JUMPDEST = 0x5b;
  public static final int EVM_INST_JUMPI = 0x57;
  public static final int EVM_INST_LOG0 = 0xa0;
  public static final int EVM_INST_LOG1 = 0xa1;
  public static final int EVM_INST_LOG2 = 0xa2;
  public static final int EVM_INST_LOG3 = 0xa3;
  public static final int EVM_INST_LOG4 = 0xa4;
  public static final int EVM_INST_LT = 0x10;
  public static final int EVM_INST_MLOAD = 0x51;
  public static final int EVM_INST_MOD = 0x6;
  public static final int EVM_INST_MSIZE = 0x59;
  public static final int EVM_INST_MSTORE = 0x52;
  public static final int EVM_INST_MSTORE8 = 0x53;
  public static final int EVM_INST_MUL = 0x2;
  public static final int EVM_INST_MULMOD = 0x9;
  public static final int EVM_INST_NOT = 0x19;
  public static final int EVM_INST_NUMBER = 0x43;
  public static final int EVM_INST_OR = 0x17;
  public static final int EVM_INST_ORIGIN = 0x32;
  public static final int EVM_INST_PC = 0x58;
  public static final int EVM_INST_POP = 0x50;
  public static final int EVM_INST_PUSH1 = 0x60;
  public static final int EVM_INST_PUSH10 = 0x69;
  public static final int EVM_INST_PUSH11 = 0x6a;
  public static final int EVM_INST_PUSH12 = 0x6b;
  public static final int EVM_INST_PUSH13 = 0x6c;
  public static final int EVM_INST_PUSH14 = 0x6d;
  public static final int EVM_INST_PUSH15 = 0x6e;
  public static final int EVM_INST_PUSH16 = 0x6f;
  public static final int EVM_INST_PUSH17 = 0x70;
  public static final int EVM_INST_PUSH18 = 0x71;
  public static final int EVM_INST_PUSH19 = 0x72;
  public static final int EVM_INST_PUSH2 = 0x61;
  public static final int EVM_INST_PUSH20 = 0x73;
  public static final int EVM_INST_PUSH21 = 0x74;
  public static final int EVM_INST_PUSH22 = 0x75;
  public static final int EVM_INST_PUSH23 = 0x76;
  public static final int EVM_INST_PUSH24 = 0x77;
  public static final int EVM_INST_PUSH25 = 0x78;
  public static final int EVM_INST_PUSH26 = 0x79;
  public static final int EVM_INST_PUSH27 = 0x7a;
  public static final int EVM_INST_PUSH28 = 0x7b;
  public static final int EVM_INST_PUSH29 = 0x7c;
  public static final int EVM_INST_PUSH3 = 0x62;
  public static final int EVM_INST_PUSH30 = 0x7d;
  public static final int EVM_INST_PUSH31 = 0x7e;
  public static final int EVM_INST_PUSH32 = 0x7f;
  public static final int EVM_INST_PUSH4 = 0x63;
  public static final int EVM_INST_PUSH5 = 0x64;
  public static final int EVM_INST_PUSH6 = 0x65;
  public static final int EVM_INST_PUSH7 = 0x66;
  public static final int EVM_INST_PUSH8 = 0x67;
  public static final int EVM_INST_PUSH9 = 0x68;
  public static final int EVM_INST_RETURN = 0xf3;
  public static final int EVM_INST_RETURNDATACOPY = 0x3e;
  public static final int EVM_INST_RETURNDATASIZE = 0x3d;
  public static final int EVM_INST_REVERT = 0xfd;
  public static final int EVM_INST_SAR = 0x1d;
  public static final int EVM_INST_SDIV = 0x5;
  public static final int EVM_INST_SELFBALANCE = 0x47;
  public static final int EVM_INST_SELFDESTRUCT = 0xff;
  public static final int EVM_INST_SGT = 0x13;
  public static final int EVM_INST_SHA3 = 0x20;
  public static final int EVM_INST_SHL = 0x1b;
  public static final int EVM_INST_SHR = 0x1c;
  public static final int EVM_INST_SIGNEXTEND = 0xb;
  public static final int EVM_INST_SLOAD = 0x54;
  public static final int EVM_INST_SLT = 0x12;
  public static final int EVM_INST_SMOD = 0x7;
  public static final int EVM_INST_SSTORE = 0x55;
  public static final int EVM_INST_STATICCALL = 0xfa;
  public static final int EVM_INST_STOP = 0x0;
  public static final int EVM_INST_SUB = 0x3;
  public static final int EVM_INST_SWAP1 = 0x90;
  public static final int EVM_INST_SWAP10 = 0x99;
  public static final int EVM_INST_SWAP11 = 0x9a;
  public static final int EVM_INST_SWAP12 = 0x9b;
  public static final int EVM_INST_SWAP13 = 0x9c;
  public static final int EVM_INST_SWAP14 = 0x9d;
  public static final int EVM_INST_SWAP15 = 0x9e;
  public static final int EVM_INST_SWAP16 = 0x9f;
  public static final int EVM_INST_SWAP2 = 0x91;
  public static final int EVM_INST_SWAP3 = 0x92;
  public static final int EVM_INST_SWAP4 = 0x93;
  public static final int EVM_INST_SWAP5 = 0x94;
  public static final int EVM_INST_SWAP6 = 0x95;
  public static final int EVM_INST_SWAP7 = 0x96;
  public static final int EVM_INST_SWAP8 = 0x97;
  public static final int EVM_INST_SWAP9 = 0x98;
  public static final int EVM_INST_TIMESTAMP = 0x42;
  public static final int EVM_INST_XOR = 0x18;
  public static final int INVALID_CODE_PREFIX_VALUE = 0xef;
  public static final int LLARGE = 0x10;
  public static final int LLARGEMO = 0xf;
  public static final int LLARGEPO = 0x11;
  public static final int MMEDIUM = 0x8;
  public static final int MMEDIUMMO = 0x7;
  public static final int MMIO_INST_LIMB_TO_RAM_ONE_TARGET = 0xfe12;
  public static final int MMIO_INST_LIMB_TO_RAM_TRANSPLANT = 0xfe11;
  public static final int MMIO_INST_LIMB_TO_RAM_TWO_TARGET = 0xfe13;
  public static final int MMIO_INST_LIMB_VANISHES = 0xfe01;
  public static final int MMIO_INST_RAM_EXCISION = 0xfe41;
  public static final int MMIO_INST_RAM_TO_LIMB_ONE_SOURCE = 0xfe22;
  public static final int MMIO_INST_RAM_TO_LIMB_TRANSPLANT = 0xfe21;
  public static final int MMIO_INST_RAM_TO_LIMB_TWO_SOURCE = 0xfe23;
  public static final int MMIO_INST_RAM_TO_RAM_PARTIAL = 0xfe32;
  public static final int MMIO_INST_RAM_TO_RAM_TRANSPLANT = 0xfe31;
  public static final int MMIO_INST_RAM_TO_RAM_TWO_SOURCE = 0xfe34;
  public static final int MMIO_INST_RAM_TO_RAM_TWO_TARGET = 0xfe33;
  public static final int MMIO_INST_RAM_VANISHES = 0xfe42;
  public static final int MMU_INST_ANY_TO_RAM_WITH_PADDING = 0xfe50;
  public static final int MMU_INST_ANY_TO_RAM_WITH_PADDING_PURE_PADDING = 0xfe52;
  public static final int MMU_INST_ANY_TO_RAM_WITH_PADDING_SOME_DATA = 0xfe51;
  public static final int MMU_INST_BLAKE = 0xfe80;
  public static final int MMU_INST_EXO_TO_RAM_TRANSPLANTS = 0xfe30;
  public static final int MMU_INST_INVALID_CODE_PREFIX = 0xfe00;
  public static final int MMU_INST_MLOAD = 0xfe01;
  public static final int MMU_INST_MODEXP_DATA = 0xfe70;
  public static final int MMU_INST_MODEXP_ZERO = 0xfe60;
  public static final int MMU_INST_MSTORE = 0xfe02;
  public static final int MMU_INST_MSTORE8 = 0x53;
  public static final int MMU_INST_RAM_TO_EXO_WITH_PADDING = 0xfe20;
  public static final int MMU_INST_RAM_TO_RAM_SANS_PADDING = 0xfe40;
  public static final int MMU_INST_RIGHT_PADDED_WORD_EXTRACTION = 0xfe10;
  public static final int MODEXP_PHASE_BASE = 0x1;
  public static final int MODEXP_PHASE_EXPONENT = 0x2;
  public static final int MODEXP_PHASE_MODULUS = 0x3;
  public static final int MODEXP_PHASE_RESULT = 0x4;
  public static final int RLP_ADDR_RECIPE_1 = 0x1;
  public static final int RLP_ADDR_RECIPE_2 = 0x2;
  public static final int RLP_PREFIX_INT_LONG = 0xb7;
  public static final int RLP_PREFIX_INT_SHORT = 0x80;
  public static final int RLP_PREFIX_LIST_LONG = 0xf7;
  public static final int RLP_PREFIX_LIST_SHORT = 0xc0;
  public static final int RLP_RCPT_SUBPHASE_ID_ADDR = 0x35;
  public static final int RLP_RCPT_SUBPHASE_ID_CUMUL_GAS = 0x3;
  public static final int RLP_RCPT_SUBPHASE_ID_DATA_LIMB = 0x4d;
  public static final int RLP_RCPT_SUBPHASE_ID_DATA_SIZE = 0x53;
  public static final int RLP_RCPT_SUBPHASE_ID_NO_LOG_ENTRY = 0xb;
  public static final int RLP_RCPT_SUBPHASE_ID_STATUS_CODE = 0x2;
  public static final int RLP_RCPT_SUBPHASE_ID_TOPIC_BASE = 0x41;
  public static final int RLP_RCPT_SUBPHASE_ID_TOPIC_DELTA = 0x60;
  public static final int RLP_RCPT_SUBPHASE_ID_TYPE = 0x7;
  public static final int RLP_TXN_PHASE_ACCESS_LIST_VALUE = 0xb;
  public static final int RLP_TXN_PHASE_BETA_VALUE = 0xc;
  public static final int RLP_TXN_PHASE_CHAIN_ID_VALUE = 0x2;
  public static final int RLP_TXN_PHASE_DATA_VALUE = 0xa;
  public static final int RLP_TXN_PHASE_GAS_LIMIT_VALUE = 0x7;
  public static final int RLP_TXN_PHASE_GAS_PRICE_VALUE = 0x4;
  public static final int RLP_TXN_PHASE_MAX_FEE_PER_GAS_VALUE = 0x6;
  public static final int RLP_TXN_PHASE_MAX_PRIORITY_FEE_PER_GAS_VALUE = 0x5;
  public static final int RLP_TXN_PHASE_NONCE_VALUE = 0x3;
  public static final int RLP_TXN_PHASE_RLP_PREFIX_VALUE = 0x1;
  public static final int RLP_TXN_PHASE_R_VALUE = 0xe;
  public static final int RLP_TXN_PHASE_S_VALUE = 0xf;
  public static final int RLP_TXN_PHASE_TO_VALUE = 0x8;
  public static final int RLP_TXN_PHASE_VALUE_VALUE = 0x9;
  public static final int RLP_TXN_PHASE_Y_VALUE = 0xd;
  public static final int WCP_INST_GEQ = 0xe;
  public static final int WCP_INST_LEQ = 0xf;
  public static final int WORD_SIZE = 0x20;
  public static final int WORD_SIZE_MO = 0x1f;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer acc1;
  private final MappedByteBuffer acc2;
  private final MappedByteBuffer acc3;
  private final MappedByteBuffer acc4;
  private final MappedByteBuffer accA;
  private final MappedByteBuffer accQ;
  private final MappedByteBuffer accW;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer byteA;
  private final MappedByteBuffer byteQ;
  private final MappedByteBuffer byteQq;
  private final MappedByteBuffer byteR;
  private final MappedByteBuffer byteW;
  private final MappedByteBuffer cMem;
  private final MappedByteBuffer cMemNew;
  private final MappedByteBuffer cn;
  private final MappedByteBuffer comp;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer deploys;
  private final MappedByteBuffer expands;
  private final MappedByteBuffer gasMxp;
  private final MappedByteBuffer gbyte;
  private final MappedByteBuffer gword;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer linCost;
  private final MappedByteBuffer maxOffset;
  private final MappedByteBuffer maxOffset1;
  private final MappedByteBuffer maxOffset2;
  private final MappedByteBuffer mtntop;
  private final MappedByteBuffer mxpType1;
  private final MappedByteBuffer mxpType2;
  private final MappedByteBuffer mxpType3;
  private final MappedByteBuffer mxpType4;
  private final MappedByteBuffer mxpType5;
  private final MappedByteBuffer mxpx;
  private final MappedByteBuffer noop;
  private final MappedByteBuffer offset1Hi;
  private final MappedByteBuffer offset1Lo;
  private final MappedByteBuffer offset2Hi;
  private final MappedByteBuffer offset2Lo;
  private final MappedByteBuffer quadCost;
  private final MappedByteBuffer roob;
  private final MappedByteBuffer size1Hi;
  private final MappedByteBuffer size1Lo;
  private final MappedByteBuffer size2Hi;
  private final MappedByteBuffer size2Lo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer words;
  private final MappedByteBuffer wordsNew;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("mxp.ACC_1", 32, length),
        new ColumnHeader("mxp.ACC_2", 32, length),
        new ColumnHeader("mxp.ACC_3", 32, length),
        new ColumnHeader("mxp.ACC_4", 32, length),
        new ColumnHeader("mxp.ACC_A", 32, length),
        new ColumnHeader("mxp.ACC_Q", 32, length),
        new ColumnHeader("mxp.ACC_W", 32, length),
        new ColumnHeader("mxp.BYTE_1", 1, length),
        new ColumnHeader("mxp.BYTE_2", 1, length),
        new ColumnHeader("mxp.BYTE_3", 1, length),
        new ColumnHeader("mxp.BYTE_4", 1, length),
        new ColumnHeader("mxp.BYTE_A", 1, length),
        new ColumnHeader("mxp.BYTE_Q", 1, length),
        new ColumnHeader("mxp.BYTE_QQ", 1, length),
        new ColumnHeader("mxp.BYTE_R", 1, length),
        new ColumnHeader("mxp.BYTE_W", 1, length),
        new ColumnHeader("mxp.C_MEM", 32, length),
        new ColumnHeader("mxp.C_MEM_NEW", 32, length),
        new ColumnHeader("mxp.CN", 32, length),
        new ColumnHeader("mxp.COMP", 1, length),
        new ColumnHeader("mxp.CT", 2, length),
        new ColumnHeader("mxp.DEPLOYS", 1, length),
        new ColumnHeader("mxp.EXPANDS", 1, length),
        new ColumnHeader("mxp.GAS_MXP", 32, length),
        new ColumnHeader("mxp.GBYTE", 32, length),
        new ColumnHeader("mxp.GWORD", 32, length),
        new ColumnHeader("mxp.INST", 1, length),
        new ColumnHeader("mxp.LIN_COST", 32, length),
        new ColumnHeader("mxp.MAX_OFFSET", 32, length),
        new ColumnHeader("mxp.MAX_OFFSET_1", 32, length),
        new ColumnHeader("mxp.MAX_OFFSET_2", 32, length),
        new ColumnHeader("mxp.MTNTOP", 1, length),
        new ColumnHeader("mxp.MXP_TYPE_1", 1, length),
        new ColumnHeader("mxp.MXP_TYPE_2", 1, length),
        new ColumnHeader("mxp.MXP_TYPE_3", 1, length),
        new ColumnHeader("mxp.MXP_TYPE_4", 1, length),
        new ColumnHeader("mxp.MXP_TYPE_5", 1, length),
        new ColumnHeader("mxp.MXPX", 1, length),
        new ColumnHeader("mxp.NOOP", 1, length),
        new ColumnHeader("mxp.OFFSET_1_HI", 32, length),
        new ColumnHeader("mxp.OFFSET_1_LO", 32, length),
        new ColumnHeader("mxp.OFFSET_2_HI", 32, length),
        new ColumnHeader("mxp.OFFSET_2_LO", 32, length),
        new ColumnHeader("mxp.QUAD_COST", 32, length),
        new ColumnHeader("mxp.ROOB", 1, length),
        new ColumnHeader("mxp.SIZE_1_HI", 32, length),
        new ColumnHeader("mxp.SIZE_1_LO", 32, length),
        new ColumnHeader("mxp.SIZE_2_HI", 32, length),
        new ColumnHeader("mxp.SIZE_2_LO", 32, length),
        new ColumnHeader("mxp.STAMP", 8, length),
        new ColumnHeader("mxp.WORDS", 32, length),
        new ColumnHeader("mxp.WORDS_NEW", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.accA = buffers.get(4);
    this.accQ = buffers.get(5);
    this.accW = buffers.get(6);
    this.byte1 = buffers.get(7);
    this.byte2 = buffers.get(8);
    this.byte3 = buffers.get(9);
    this.byte4 = buffers.get(10);
    this.byteA = buffers.get(11);
    this.byteQ = buffers.get(12);
    this.byteQq = buffers.get(13);
    this.byteR = buffers.get(14);
    this.byteW = buffers.get(15);
    this.cMem = buffers.get(16);
    this.cMemNew = buffers.get(17);
    this.cn = buffers.get(18);
    this.comp = buffers.get(19);
    this.ct = buffers.get(20);
    this.deploys = buffers.get(21);
    this.expands = buffers.get(22);
    this.gasMxp = buffers.get(23);
    this.gbyte = buffers.get(24);
    this.gword = buffers.get(25);
    this.inst = buffers.get(26);
    this.linCost = buffers.get(27);
    this.maxOffset = buffers.get(28);
    this.maxOffset1 = buffers.get(29);
    this.maxOffset2 = buffers.get(30);
    this.mtntop = buffers.get(31);
    this.mxpType1 = buffers.get(32);
    this.mxpType2 = buffers.get(33);
    this.mxpType3 = buffers.get(34);
    this.mxpType4 = buffers.get(35);
    this.mxpType5 = buffers.get(36);
    this.mxpx = buffers.get(37);
    this.noop = buffers.get(38);
    this.offset1Hi = buffers.get(39);
    this.offset1Lo = buffers.get(40);
    this.offset2Hi = buffers.get(41);
    this.offset2Lo = buffers.get(42);
    this.quadCost = buffers.get(43);
    this.roob = buffers.get(44);
    this.size1Hi = buffers.get(45);
    this.size1Lo = buffers.get(46);
    this.size2Hi = buffers.get(47);
    this.size2Lo = buffers.get(48);
    this.stamp = buffers.get(49);
    this.words = buffers.get(50);
    this.wordsNew = buffers.get(51);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mxp.ACC_1 already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc1.put((byte) 0);
    }
    acc1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mxp.ACC_2 already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc2.put((byte) 0);
    }
    acc2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mxp.ACC_3 already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc3.put((byte) 0);
    }
    acc3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mxp.ACC_4 already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc4.put((byte) 0);
    }
    acc4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accA(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mxp.ACC_A already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accA.put((byte) 0);
    }
    accA.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mxp.ACC_Q already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ.put((byte) 0);
    }
    accQ.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accW(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mxp.ACC_W already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accW.put((byte) 0);
    }
    accW.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mxp.BYTE_1 already set");
    } else {
      filled.set(7);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mxp.BYTE_2 already set");
    } else {
      filled.set(8);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mxp.BYTE_3 already set");
    } else {
      filled.set(9);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mxp.BYTE_4 already set");
    } else {
      filled.set(10);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byteA(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mxp.BYTE_A already set");
    } else {
      filled.set(11);
    }

    byteA.put(b.toByte());

    return this;
  }

  public Trace byteQ(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mxp.BYTE_Q already set");
    } else {
      filled.set(12);
    }

    byteQ.put(b.toByte());

    return this;
  }

  public Trace byteQq(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mxp.BYTE_QQ already set");
    } else {
      filled.set(13);
    }

    byteQq.put(b.toByte());

    return this;
  }

  public Trace byteR(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mxp.BYTE_R already set");
    } else {
      filled.set(14);
    }

    byteR.put(b.toByte());

    return this;
  }

  public Trace byteW(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mxp.BYTE_W already set");
    } else {
      filled.set(15);
    }

    byteW.put(b.toByte());

    return this;
  }

  public Trace cMem(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mxp.C_MEM already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cMem.put((byte) 0);
    }
    cMem.put(b.toArrayUnsafe());

    return this;
  }

  public Trace cMemNew(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mxp.C_MEM_NEW already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cMemNew.put((byte) 0);
    }
    cMemNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace cn(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mxp.CN already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cn.put((byte) 0);
    }
    cn.put(b.toArrayUnsafe());

    return this;
  }

  public Trace comp(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mxp.COMP already set");
    } else {
      filled.set(17);
    }

    comp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final short b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mxp.CT already set");
    } else {
      filled.set(18);
    }

    ct.putShort(b);

    return this;
  }

  public Trace deploys(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mxp.DEPLOYS already set");
    } else {
      filled.set(21);
    }

    deploys.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace expands(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mxp.EXPANDS already set");
    } else {
      filled.set(22);
    }

    expands.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace gasMxp(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mxp.GAS_MXP already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasMxp.put((byte) 0);
    }
    gasMxp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gbyte(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mxp.GBYTE already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gbyte.put((byte) 0);
    }
    gbyte.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gword(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mxp.GWORD already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gword.put((byte) 0);
    }
    gword.put(b.toArrayUnsafe());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mxp.INST already set");
    } else {
      filled.set(26);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace linCost(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mxp.LIN_COST already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      linCost.put((byte) 0);
    }
    linCost.put(b.toArrayUnsafe());

    return this;
  }

  public Trace maxOffset(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mxp.MAX_OFFSET already set");
    } else {
      filled.set(28);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      maxOffset.put((byte) 0);
    }
    maxOffset.put(b.toArrayUnsafe());

    return this;
  }

  public Trace maxOffset1(final Bytes b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mxp.MAX_OFFSET_1 already set");
    } else {
      filled.set(29);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      maxOffset1.put((byte) 0);
    }
    maxOffset1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace maxOffset2(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mxp.MAX_OFFSET_2 already set");
    } else {
      filled.set(30);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      maxOffset2.put((byte) 0);
    }
    maxOffset2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace mtntop(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mxp.MTNTOP already set");
    } else {
      filled.set(31);
    }

    mtntop.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType1(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mxp.MXP_TYPE_1 already set");
    } else {
      filled.set(33);
    }

    mxpType1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType2(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mxp.MXP_TYPE_2 already set");
    } else {
      filled.set(34);
    }

    mxpType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType3(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mxp.MXP_TYPE_3 already set");
    } else {
      filled.set(35);
    }

    mxpType3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType4(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mxp.MXP_TYPE_4 already set");
    } else {
      filled.set(36);
    }

    mxpType4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType5(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mxp.MXP_TYPE_5 already set");
    } else {
      filled.set(37);
    }

    mxpType5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpx(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mxp.MXPX already set");
    } else {
      filled.set(32);
    }

    mxpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace noop(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mxp.NOOP already set");
    } else {
      filled.set(38);
    }

    noop.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace offset1Hi(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mxp.OFFSET_1_HI already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      offset1Hi.put((byte) 0);
    }
    offset1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace offset1Lo(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mxp.OFFSET_1_LO already set");
    } else {
      filled.set(40);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      offset1Lo.put((byte) 0);
    }
    offset1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace offset2Hi(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mxp.OFFSET_2_HI already set");
    } else {
      filled.set(41);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      offset2Hi.put((byte) 0);
    }
    offset2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace offset2Lo(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mxp.OFFSET_2_LO already set");
    } else {
      filled.set(42);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      offset2Lo.put((byte) 0);
    }
    offset2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace quadCost(final Bytes b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mxp.QUAD_COST already set");
    } else {
      filled.set(43);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      quadCost.put((byte) 0);
    }
    quadCost.put(b.toArrayUnsafe());

    return this;
  }

  public Trace roob(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mxp.ROOB already set");
    } else {
      filled.set(44);
    }

    roob.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace size1Hi(final Bytes b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mxp.SIZE_1_HI already set");
    } else {
      filled.set(45);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      size1Hi.put((byte) 0);
    }
    size1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace size1Lo(final Bytes b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mxp.SIZE_1_LO already set");
    } else {
      filled.set(46);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      size1Lo.put((byte) 0);
    }
    size1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace size2Hi(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mxp.SIZE_2_HI already set");
    } else {
      filled.set(47);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      size2Hi.put((byte) 0);
    }
    size2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace size2Lo(final Bytes b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mxp.SIZE_2_LO already set");
    } else {
      filled.set(48);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      size2Lo.put((byte) 0);
    }
    size2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mxp.STAMP already set");
    } else {
      filled.set(49);
    }

    stamp.putLong(b);

    return this;
  }

  public Trace words(final Bytes b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mxp.WORDS already set");
    } else {
      filled.set(50);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      words.put((byte) 0);
    }
    words.put(b.toArrayUnsafe());

    return this;
  }

  public Trace wordsNew(final Bytes b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mxp.WORDS_NEW already set");
    } else {
      filled.set(51);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      wordsNew.put((byte) 0);
    }
    wordsNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mxp.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mxp.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mxp.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mxp.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mxp.ACC_A has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mxp.ACC_Q has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mxp.ACC_W has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mxp.BYTE_1 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mxp.BYTE_2 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mxp.BYTE_3 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mxp.BYTE_4 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mxp.BYTE_A has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mxp.BYTE_Q has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mxp.BYTE_QQ has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mxp.BYTE_R has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mxp.BYTE_W has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mxp.C_MEM has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mxp.C_MEM_NEW has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mxp.CN has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mxp.COMP has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mxp.CT has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mxp.DEPLOYS has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mxp.EXPANDS has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mxp.GAS_MXP has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mxp.GBYTE has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mxp.GWORD has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mxp.INST has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mxp.LIN_COST has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mxp.MAX_OFFSET has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mxp.MAX_OFFSET_1 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mxp.MAX_OFFSET_2 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mxp.MTNTOP has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mxp.MXP_TYPE_1 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mxp.MXP_TYPE_2 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mxp.MXP_TYPE_3 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mxp.MXP_TYPE_4 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mxp.MXP_TYPE_5 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mxp.MXPX has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mxp.NOOP has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mxp.OFFSET_1_HI has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mxp.OFFSET_1_LO has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mxp.OFFSET_2_HI has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mxp.OFFSET_2_LO has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mxp.QUAD_COST has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mxp.ROOB has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mxp.SIZE_1_HI has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mxp.SIZE_1_LO has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mxp.SIZE_2_HI has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mxp.SIZE_2_LO has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mxp.STAMP has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mxp.WORDS has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mxp.WORDS_NEW has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc1.position(acc1.position() + 32);
    }

    if (!filled.get(1)) {
      acc2.position(acc2.position() + 32);
    }

    if (!filled.get(2)) {
      acc3.position(acc3.position() + 32);
    }

    if (!filled.get(3)) {
      acc4.position(acc4.position() + 32);
    }

    if (!filled.get(4)) {
      accA.position(accA.position() + 32);
    }

    if (!filled.get(5)) {
      accQ.position(accQ.position() + 32);
    }

    if (!filled.get(6)) {
      accW.position(accW.position() + 32);
    }

    if (!filled.get(7)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(8)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(9)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(10)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(11)) {
      byteA.position(byteA.position() + 1);
    }

    if (!filled.get(12)) {
      byteQ.position(byteQ.position() + 1);
    }

    if (!filled.get(13)) {
      byteQq.position(byteQq.position() + 1);
    }

    if (!filled.get(14)) {
      byteR.position(byteR.position() + 1);
    }

    if (!filled.get(15)) {
      byteW.position(byteW.position() + 1);
    }

    if (!filled.get(19)) {
      cMem.position(cMem.position() + 32);
    }

    if (!filled.get(20)) {
      cMemNew.position(cMemNew.position() + 32);
    }

    if (!filled.get(16)) {
      cn.position(cn.position() + 32);
    }

    if (!filled.get(17)) {
      comp.position(comp.position() + 1);
    }

    if (!filled.get(18)) {
      ct.position(ct.position() + 2);
    }

    if (!filled.get(21)) {
      deploys.position(deploys.position() + 1);
    }

    if (!filled.get(22)) {
      expands.position(expands.position() + 1);
    }

    if (!filled.get(23)) {
      gasMxp.position(gasMxp.position() + 32);
    }

    if (!filled.get(24)) {
      gbyte.position(gbyte.position() + 32);
    }

    if (!filled.get(25)) {
      gword.position(gword.position() + 32);
    }

    if (!filled.get(26)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(27)) {
      linCost.position(linCost.position() + 32);
    }

    if (!filled.get(28)) {
      maxOffset.position(maxOffset.position() + 32);
    }

    if (!filled.get(29)) {
      maxOffset1.position(maxOffset1.position() + 32);
    }

    if (!filled.get(30)) {
      maxOffset2.position(maxOffset2.position() + 32);
    }

    if (!filled.get(31)) {
      mtntop.position(mtntop.position() + 1);
    }

    if (!filled.get(33)) {
      mxpType1.position(mxpType1.position() + 1);
    }

    if (!filled.get(34)) {
      mxpType2.position(mxpType2.position() + 1);
    }

    if (!filled.get(35)) {
      mxpType3.position(mxpType3.position() + 1);
    }

    if (!filled.get(36)) {
      mxpType4.position(mxpType4.position() + 1);
    }

    if (!filled.get(37)) {
      mxpType5.position(mxpType5.position() + 1);
    }

    if (!filled.get(32)) {
      mxpx.position(mxpx.position() + 1);
    }

    if (!filled.get(38)) {
      noop.position(noop.position() + 1);
    }

    if (!filled.get(39)) {
      offset1Hi.position(offset1Hi.position() + 32);
    }

    if (!filled.get(40)) {
      offset1Lo.position(offset1Lo.position() + 32);
    }

    if (!filled.get(41)) {
      offset2Hi.position(offset2Hi.position() + 32);
    }

    if (!filled.get(42)) {
      offset2Lo.position(offset2Lo.position() + 32);
    }

    if (!filled.get(43)) {
      quadCost.position(quadCost.position() + 32);
    }

    if (!filled.get(44)) {
      roob.position(roob.position() + 1);
    }

    if (!filled.get(45)) {
      size1Hi.position(size1Hi.position() + 32);
    }

    if (!filled.get(46)) {
      size1Lo.position(size1Lo.position() + 32);
    }

    if (!filled.get(47)) {
      size2Hi.position(size2Hi.position() + 32);
    }

    if (!filled.get(48)) {
      size2Lo.position(size2Lo.position() + 32);
    }

    if (!filled.get(49)) {
      stamp.position(stamp.position() + 8);
    }

    if (!filled.get(50)) {
      words.position(words.position() + 32);
    }

    if (!filled.get(51)) {
      wordsNew.position(wordsNew.position() + 32);
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
