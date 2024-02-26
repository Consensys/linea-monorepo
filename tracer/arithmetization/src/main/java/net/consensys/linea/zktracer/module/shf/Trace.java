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

package net.consensys.linea.zktracer.module.shf;

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
  private final MappedByteBuffer acc5;
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
  private final MappedByteBuffer bitB3;
  private final MappedByteBuffer bitB4;
  private final MappedByteBuffer bitB5;
  private final MappedByteBuffer bitB6;
  private final MappedByteBuffer bitB7;
  private final MappedByteBuffer bits;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer byte5;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer iomf;
  private final MappedByteBuffer known;
  private final MappedByteBuffer leftAlignedSuffixHigh;
  private final MappedByteBuffer leftAlignedSuffixLow;
  private final MappedByteBuffer low3;
  private final MappedByteBuffer microShiftParameter;
  private final MappedByteBuffer neg;
  private final MappedByteBuffer oneLineInstruction;
  private final MappedByteBuffer ones;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer rightAlignedPrefixHigh;
  private final MappedByteBuffer rightAlignedPrefixLow;
  private final MappedByteBuffer shb3Hi;
  private final MappedByteBuffer shb3Lo;
  private final MappedByteBuffer shb4Hi;
  private final MappedByteBuffer shb4Lo;
  private final MappedByteBuffer shb5Hi;
  private final MappedByteBuffer shb5Lo;
  private final MappedByteBuffer shb6Hi;
  private final MappedByteBuffer shb6Lo;
  private final MappedByteBuffer shb7Hi;
  private final MappedByteBuffer shb7Lo;
  private final MappedByteBuffer shiftDirection;
  private final MappedByteBuffer shiftStamp;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("shf.ACC_1", 32, length),
        new ColumnHeader("shf.ACC_2", 32, length),
        new ColumnHeader("shf.ACC_3", 32, length),
        new ColumnHeader("shf.ACC_4", 32, length),
        new ColumnHeader("shf.ACC_5", 32, length),
        new ColumnHeader("shf.ARG_1_HI", 32, length),
        new ColumnHeader("shf.ARG_1_LO", 32, length),
        new ColumnHeader("shf.ARG_2_HI", 32, length),
        new ColumnHeader("shf.ARG_2_LO", 32, length),
        new ColumnHeader("shf.BIT_1", 1, length),
        new ColumnHeader("shf.BIT_2", 1, length),
        new ColumnHeader("shf.BIT_3", 1, length),
        new ColumnHeader("shf.BIT_4", 1, length),
        new ColumnHeader("shf.BIT_B_3", 1, length),
        new ColumnHeader("shf.BIT_B_4", 1, length),
        new ColumnHeader("shf.BIT_B_5", 1, length),
        new ColumnHeader("shf.BIT_B_6", 1, length),
        new ColumnHeader("shf.BIT_B_7", 1, length),
        new ColumnHeader("shf.BITS", 1, length),
        new ColumnHeader("shf.BYTE_1", 1, length),
        new ColumnHeader("shf.BYTE_2", 1, length),
        new ColumnHeader("shf.BYTE_3", 1, length),
        new ColumnHeader("shf.BYTE_4", 1, length),
        new ColumnHeader("shf.BYTE_5", 1, length),
        new ColumnHeader("shf.COUNTER", 2, length),
        new ColumnHeader("shf.INST", 1, length),
        new ColumnHeader("shf.IOMF", 1, length),
        new ColumnHeader("shf.KNOWN", 1, length),
        new ColumnHeader("shf.LEFT_ALIGNED_SUFFIX_HIGH", 1, length),
        new ColumnHeader("shf.LEFT_ALIGNED_SUFFIX_LOW", 1, length),
        new ColumnHeader("shf.LOW_3", 32, length),
        new ColumnHeader("shf.MICRO_SHIFT_PARAMETER", 2, length),
        new ColumnHeader("shf.NEG", 1, length),
        new ColumnHeader("shf.ONE_LINE_INSTRUCTION", 1, length),
        new ColumnHeader("shf.ONES", 1, length),
        new ColumnHeader("shf.RES_HI", 32, length),
        new ColumnHeader("shf.RES_LO", 32, length),
        new ColumnHeader("shf.RIGHT_ALIGNED_PREFIX_HIGH", 1, length),
        new ColumnHeader("shf.RIGHT_ALIGNED_PREFIX_LOW", 1, length),
        new ColumnHeader("shf.SHB_3_HI", 32, length),
        new ColumnHeader("shf.SHB_3_LO", 32, length),
        new ColumnHeader("shf.SHB_4_HI", 32, length),
        new ColumnHeader("shf.SHB_4_LO", 32, length),
        new ColumnHeader("shf.SHB_5_HI", 32, length),
        new ColumnHeader("shf.SHB_5_LO", 32, length),
        new ColumnHeader("shf.SHB_6_HI", 32, length),
        new ColumnHeader("shf.SHB_6_LO", 32, length),
        new ColumnHeader("shf.SHB_7_HI", 32, length),
        new ColumnHeader("shf.SHB_7_LO", 32, length),
        new ColumnHeader("shf.SHIFT_DIRECTION", 1, length),
        new ColumnHeader("shf.SHIFT_STAMP", 8, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.acc5 = buffers.get(4);
    this.arg1Hi = buffers.get(5);
    this.arg1Lo = buffers.get(6);
    this.arg2Hi = buffers.get(7);
    this.arg2Lo = buffers.get(8);
    this.bit1 = buffers.get(9);
    this.bit2 = buffers.get(10);
    this.bit3 = buffers.get(11);
    this.bit4 = buffers.get(12);
    this.bitB3 = buffers.get(13);
    this.bitB4 = buffers.get(14);
    this.bitB5 = buffers.get(15);
    this.bitB6 = buffers.get(16);
    this.bitB7 = buffers.get(17);
    this.bits = buffers.get(18);
    this.byte1 = buffers.get(19);
    this.byte2 = buffers.get(20);
    this.byte3 = buffers.get(21);
    this.byte4 = buffers.get(22);
    this.byte5 = buffers.get(23);
    this.counter = buffers.get(24);
    this.inst = buffers.get(25);
    this.iomf = buffers.get(26);
    this.known = buffers.get(27);
    this.leftAlignedSuffixHigh = buffers.get(28);
    this.leftAlignedSuffixLow = buffers.get(29);
    this.low3 = buffers.get(30);
    this.microShiftParameter = buffers.get(31);
    this.neg = buffers.get(32);
    this.oneLineInstruction = buffers.get(33);
    this.ones = buffers.get(34);
    this.resHi = buffers.get(35);
    this.resLo = buffers.get(36);
    this.rightAlignedPrefixHigh = buffers.get(37);
    this.rightAlignedPrefixLow = buffers.get(38);
    this.shb3Hi = buffers.get(39);
    this.shb3Lo = buffers.get(40);
    this.shb4Hi = buffers.get(41);
    this.shb4Lo = buffers.get(42);
    this.shb5Hi = buffers.get(43);
    this.shb5Lo = buffers.get(44);
    this.shb6Hi = buffers.get(45);
    this.shb6Lo = buffers.get(46);
    this.shb7Hi = buffers.get(47);
    this.shb7Lo = buffers.get(48);
    this.shiftDirection = buffers.get(49);
    this.shiftStamp = buffers.get(50);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("shf.ACC_1 already set");
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
      throw new IllegalStateException("shf.ACC_2 already set");
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
      throw new IllegalStateException("shf.ACC_3 already set");
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
      throw new IllegalStateException("shf.ACC_4 already set");
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

  public Trace acc5(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("shf.ACC_5 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc5.put((byte) 0);
    }
    acc5.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("shf.ARG_1_HI already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Hi.put((byte) 0);
    }
    arg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("shf.ARG_1_LO already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Lo.put((byte) 0);
    }
    arg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("shf.ARG_2_HI already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Hi.put((byte) 0);
    }
    arg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("shf.ARG_2_LO already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Lo.put((byte) 0);
    }
    arg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("shf.BIT_1 already set");
    } else {
      filled.set(10);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("shf.BIT_2 already set");
    } else {
      filled.set(11);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("shf.BIT_3 already set");
    } else {
      filled.set(12);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("shf.BIT_4 already set");
    } else {
      filled.set(13);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB3(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("shf.BIT_B_3 already set");
    } else {
      filled.set(14);
    }

    bitB3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB4(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("shf.BIT_B_4 already set");
    } else {
      filled.set(15);
    }

    bitB4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB5(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("shf.BIT_B_5 already set");
    } else {
      filled.set(16);
    }

    bitB5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB6(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("shf.BIT_B_6 already set");
    } else {
      filled.set(17);
    }

    bitB6.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB7(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("shf.BIT_B_7 already set");
    } else {
      filled.set(18);
    }

    bitB7.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("shf.BITS already set");
    } else {
      filled.set(9);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("shf.BYTE_1 already set");
    } else {
      filled.set(19);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("shf.BYTE_2 already set");
    } else {
      filled.set(20);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("shf.BYTE_3 already set");
    } else {
      filled.set(21);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("shf.BYTE_4 already set");
    } else {
      filled.set(22);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("shf.BYTE_5 already set");
    } else {
      filled.set(23);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace counter(final short b) {
    if (filled.get(24)) {
      throw new IllegalStateException("shf.COUNTER already set");
    } else {
      filled.set(24);
    }

    counter.putShort(b);

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(25)) {
      throw new IllegalStateException("shf.INST already set");
    } else {
      filled.set(25);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace iomf(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("shf.IOMF already set");
    } else {
      filled.set(26);
    }

    iomf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace known(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("shf.KNOWN already set");
    } else {
      filled.set(27);
    }

    known.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace leftAlignedSuffixHigh(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_HIGH already set");
    } else {
      filled.set(28);
    }

    leftAlignedSuffixHigh.put(b.toByte());

    return this;
  }

  public Trace leftAlignedSuffixLow(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_LOW already set");
    } else {
      filled.set(29);
    }

    leftAlignedSuffixLow.put(b.toByte());

    return this;
  }

  public Trace low3(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("shf.LOW_3 already set");
    } else {
      filled.set(30);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      low3.put((byte) 0);
    }
    low3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace microShiftParameter(final short b) {
    if (filled.get(31)) {
      throw new IllegalStateException("shf.MICRO_SHIFT_PARAMETER already set");
    } else {
      filled.set(31);
    }

    microShiftParameter.putShort(b);

    return this;
  }

  public Trace neg(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("shf.NEG already set");
    } else {
      filled.set(32);
    }

    neg.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oneLineInstruction(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("shf.ONE_LINE_INSTRUCTION already set");
    } else {
      filled.set(34);
    }

    oneLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ones(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("shf.ONES already set");
    } else {
      filled.set(33);
    }

    ones.put(b.toByte());

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("shf.RES_HI already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resHi.put((byte) 0);
    }
    resHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("shf.RES_LO already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resLo.put((byte) 0);
    }
    resLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace rightAlignedPrefixHigh(final UnsignedByte b) {
    if (filled.get(37)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_HIGH already set");
    } else {
      filled.set(37);
    }

    rightAlignedPrefixHigh.put(b.toByte());

    return this;
  }

  public Trace rightAlignedPrefixLow(final UnsignedByte b) {
    if (filled.get(38)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_LOW already set");
    } else {
      filled.set(38);
    }

    rightAlignedPrefixLow.put(b.toByte());

    return this;
  }

  public Trace shb3Hi(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("shf.SHB_3_HI already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb3Hi.put((byte) 0);
    }
    shb3Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb3Lo(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("shf.SHB_3_LO already set");
    } else {
      filled.set(40);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb3Lo.put((byte) 0);
    }
    shb3Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb4Hi(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("shf.SHB_4_HI already set");
    } else {
      filled.set(41);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb4Hi.put((byte) 0);
    }
    shb4Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb4Lo(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("shf.SHB_4_LO already set");
    } else {
      filled.set(42);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb4Lo.put((byte) 0);
    }
    shb4Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb5Hi(final Bytes b) {
    if (filled.get(43)) {
      throw new IllegalStateException("shf.SHB_5_HI already set");
    } else {
      filled.set(43);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb5Hi.put((byte) 0);
    }
    shb5Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb5Lo(final Bytes b) {
    if (filled.get(44)) {
      throw new IllegalStateException("shf.SHB_5_LO already set");
    } else {
      filled.set(44);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb5Lo.put((byte) 0);
    }
    shb5Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb6Hi(final Bytes b) {
    if (filled.get(45)) {
      throw new IllegalStateException("shf.SHB_6_HI already set");
    } else {
      filled.set(45);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb6Hi.put((byte) 0);
    }
    shb6Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb6Lo(final Bytes b) {
    if (filled.get(46)) {
      throw new IllegalStateException("shf.SHB_6_LO already set");
    } else {
      filled.set(46);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb6Lo.put((byte) 0);
    }
    shb6Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb7Hi(final Bytes b) {
    if (filled.get(47)) {
      throw new IllegalStateException("shf.SHB_7_HI already set");
    } else {
      filled.set(47);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb7Hi.put((byte) 0);
    }
    shb7Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shb7Lo(final Bytes b) {
    if (filled.get(48)) {
      throw new IllegalStateException("shf.SHB_7_LO already set");
    } else {
      filled.set(48);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      shb7Lo.put((byte) 0);
    }
    shb7Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace shiftDirection(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("shf.SHIFT_DIRECTION already set");
    } else {
      filled.set(49);
    }

    shiftDirection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace shiftStamp(final long b) {
    if (filled.get(50)) {
      throw new IllegalStateException("shf.SHIFT_STAMP already set");
    } else {
      filled.set(50);
    }

    shiftStamp.putLong(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("shf.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("shf.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("shf.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("shf.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("shf.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("shf.ARG_1_HI has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("shf.ARG_1_LO has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("shf.ARG_2_HI has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("shf.ARG_2_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("shf.BIT_1 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("shf.BIT_2 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("shf.BIT_3 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("shf.BIT_4 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("shf.BIT_B_3 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("shf.BIT_B_4 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("shf.BIT_B_5 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("shf.BIT_B_6 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("shf.BIT_B_7 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("shf.BITS has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("shf.BYTE_1 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("shf.BYTE_2 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("shf.BYTE_3 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("shf.BYTE_4 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("shf.BYTE_5 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("shf.COUNTER has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("shf.INST has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("shf.IOMF has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("shf.KNOWN has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_HIGH has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("shf.LEFT_ALIGNED_SUFFIX_LOW has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("shf.LOW_3 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("shf.MICRO_SHIFT_PARAMETER has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("shf.NEG has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("shf.ONE_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("shf.ONES has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("shf.RES_HI has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("shf.RES_LO has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_HIGH has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("shf.RIGHT_ALIGNED_PREFIX_LOW has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("shf.SHB_3_HI has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("shf.SHB_3_LO has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("shf.SHB_4_HI has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("shf.SHB_4_LO has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("shf.SHB_5_HI has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("shf.SHB_5_LO has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("shf.SHB_6_HI has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("shf.SHB_6_LO has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("shf.SHB_7_HI has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("shf.SHB_7_LO has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("shf.SHIFT_DIRECTION has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("shf.SHIFT_STAMP has not been filled");
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
      acc5.position(acc5.position() + 32);
    }

    if (!filled.get(5)) {
      arg1Hi.position(arg1Hi.position() + 32);
    }

    if (!filled.get(6)) {
      arg1Lo.position(arg1Lo.position() + 32);
    }

    if (!filled.get(7)) {
      arg2Hi.position(arg2Hi.position() + 32);
    }

    if (!filled.get(8)) {
      arg2Lo.position(arg2Lo.position() + 32);
    }

    if (!filled.get(10)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(11)) {
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(12)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(13)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(14)) {
      bitB3.position(bitB3.position() + 1);
    }

    if (!filled.get(15)) {
      bitB4.position(bitB4.position() + 1);
    }

    if (!filled.get(16)) {
      bitB5.position(bitB5.position() + 1);
    }

    if (!filled.get(17)) {
      bitB6.position(bitB6.position() + 1);
    }

    if (!filled.get(18)) {
      bitB7.position(bitB7.position() + 1);
    }

    if (!filled.get(9)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(19)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(20)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(21)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(22)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(23)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(24)) {
      counter.position(counter.position() + 2);
    }

    if (!filled.get(25)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(26)) {
      iomf.position(iomf.position() + 1);
    }

    if (!filled.get(27)) {
      known.position(known.position() + 1);
    }

    if (!filled.get(28)) {
      leftAlignedSuffixHigh.position(leftAlignedSuffixHigh.position() + 1);
    }

    if (!filled.get(29)) {
      leftAlignedSuffixLow.position(leftAlignedSuffixLow.position() + 1);
    }

    if (!filled.get(30)) {
      low3.position(low3.position() + 32);
    }

    if (!filled.get(31)) {
      microShiftParameter.position(microShiftParameter.position() + 2);
    }

    if (!filled.get(32)) {
      neg.position(neg.position() + 1);
    }

    if (!filled.get(34)) {
      oneLineInstruction.position(oneLineInstruction.position() + 1);
    }

    if (!filled.get(33)) {
      ones.position(ones.position() + 1);
    }

    if (!filled.get(35)) {
      resHi.position(resHi.position() + 32);
    }

    if (!filled.get(36)) {
      resLo.position(resLo.position() + 32);
    }

    if (!filled.get(37)) {
      rightAlignedPrefixHigh.position(rightAlignedPrefixHigh.position() + 1);
    }

    if (!filled.get(38)) {
      rightAlignedPrefixLow.position(rightAlignedPrefixLow.position() + 1);
    }

    if (!filled.get(39)) {
      shb3Hi.position(shb3Hi.position() + 32);
    }

    if (!filled.get(40)) {
      shb3Lo.position(shb3Lo.position() + 32);
    }

    if (!filled.get(41)) {
      shb4Hi.position(shb4Hi.position() + 32);
    }

    if (!filled.get(42)) {
      shb4Lo.position(shb4Lo.position() + 32);
    }

    if (!filled.get(43)) {
      shb5Hi.position(shb5Hi.position() + 32);
    }

    if (!filled.get(44)) {
      shb5Lo.position(shb5Lo.position() + 32);
    }

    if (!filled.get(45)) {
      shb6Hi.position(shb6Hi.position() + 32);
    }

    if (!filled.get(46)) {
      shb6Lo.position(shb6Lo.position() + 32);
    }

    if (!filled.get(47)) {
      shb7Hi.position(shb7Hi.position() + 32);
    }

    if (!filled.get(48)) {
      shb7Lo.position(shb7Lo.position() + 32);
    }

    if (!filled.get(49)) {
      shiftDirection.position(shiftDirection.position() + 1);
    }

    if (!filled.get(50)) {
      shiftStamp.position(shiftStamp.position() + 8);
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
