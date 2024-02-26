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

package net.consensys.linea.zktracer.module.bin;

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
  private final MappedByteBuffer acc6;
  private final MappedByteBuffer argument1Hi;
  private final MappedByteBuffer argument1Lo;
  private final MappedByteBuffer argument2Hi;
  private final MappedByteBuffer argument2Lo;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bitB4;
  private final MappedByteBuffer bits;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer byte5;
  private final MappedByteBuffer byte6;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isAnd;
  private final MappedByteBuffer isByte;
  private final MappedByteBuffer isNot;
  private final MappedByteBuffer isOr;
  private final MappedByteBuffer isSignextend;
  private final MappedByteBuffer isXor;
  private final MappedByteBuffer low4;
  private final MappedByteBuffer neg;
  private final MappedByteBuffer pivot;
  private final MappedByteBuffer resultHi;
  private final MappedByteBuffer resultLo;
  private final MappedByteBuffer small;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer xxxByteHi;
  private final MappedByteBuffer xxxByteLo;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("bin.ACC_1", 32, length),
        new ColumnHeader("bin.ACC_2", 32, length),
        new ColumnHeader("bin.ACC_3", 32, length),
        new ColumnHeader("bin.ACC_4", 32, length),
        new ColumnHeader("bin.ACC_5", 32, length),
        new ColumnHeader("bin.ACC_6", 32, length),
        new ColumnHeader("bin.ARGUMENT_1_HI", 32, length),
        new ColumnHeader("bin.ARGUMENT_1_LO", 32, length),
        new ColumnHeader("bin.ARGUMENT_2_HI", 32, length),
        new ColumnHeader("bin.ARGUMENT_2_LO", 32, length),
        new ColumnHeader("bin.BIT_1", 1, length),
        new ColumnHeader("bin.BIT_B_4", 1, length),
        new ColumnHeader("bin.BITS", 1, length),
        new ColumnHeader("bin.BYTE_1", 1, length),
        new ColumnHeader("bin.BYTE_2", 1, length),
        new ColumnHeader("bin.BYTE_3", 1, length),
        new ColumnHeader("bin.BYTE_4", 1, length),
        new ColumnHeader("bin.BYTE_5", 1, length),
        new ColumnHeader("bin.BYTE_6", 1, length),
        new ColumnHeader("bin.COUNTER", 1, length),
        new ColumnHeader("bin.CT_MAX", 1, length),
        new ColumnHeader("bin.INST", 1, length),
        new ColumnHeader("bin.IS_AND", 1, length),
        new ColumnHeader("bin.IS_BYTE", 1, length),
        new ColumnHeader("bin.IS_NOT", 1, length),
        new ColumnHeader("bin.IS_OR", 1, length),
        new ColumnHeader("bin.IS_SIGNEXTEND", 1, length),
        new ColumnHeader("bin.IS_XOR", 1, length),
        new ColumnHeader("bin.LOW_4", 1, length),
        new ColumnHeader("bin.NEG", 1, length),
        new ColumnHeader("bin.PIVOT", 1, length),
        new ColumnHeader("bin.RESULT_HI", 32, length),
        new ColumnHeader("bin.RESULT_LO", 32, length),
        new ColumnHeader("bin.SMALL", 1, length),
        new ColumnHeader("bin.STAMP", 8, length),
        new ColumnHeader("bin.XXX_BYTE_HI", 1, length),
        new ColumnHeader("bin.XXX_BYTE_LO", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.acc5 = buffers.get(4);
    this.acc6 = buffers.get(5);
    this.argument1Hi = buffers.get(6);
    this.argument1Lo = buffers.get(7);
    this.argument2Hi = buffers.get(8);
    this.argument2Lo = buffers.get(9);
    this.bit1 = buffers.get(10);
    this.bitB4 = buffers.get(11);
    this.bits = buffers.get(12);
    this.byte1 = buffers.get(13);
    this.byte2 = buffers.get(14);
    this.byte3 = buffers.get(15);
    this.byte4 = buffers.get(16);
    this.byte5 = buffers.get(17);
    this.byte6 = buffers.get(18);
    this.counter = buffers.get(19);
    this.ctMax = buffers.get(20);
    this.inst = buffers.get(21);
    this.isAnd = buffers.get(22);
    this.isByte = buffers.get(23);
    this.isNot = buffers.get(24);
    this.isOr = buffers.get(25);
    this.isSignextend = buffers.get(26);
    this.isXor = buffers.get(27);
    this.low4 = buffers.get(28);
    this.neg = buffers.get(29);
    this.pivot = buffers.get(30);
    this.resultHi = buffers.get(31);
    this.resultLo = buffers.get(32);
    this.small = buffers.get(33);
    this.stamp = buffers.get(34);
    this.xxxByteHi = buffers.get(35);
    this.xxxByteLo = buffers.get(36);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("bin.ACC_1 already set");
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
      throw new IllegalStateException("bin.ACC_2 already set");
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
      throw new IllegalStateException("bin.ACC_3 already set");
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
      throw new IllegalStateException("bin.ACC_4 already set");
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
      throw new IllegalStateException("bin.ACC_5 already set");
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

  public Trace acc6(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("bin.ACC_6 already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc6.put((byte) 0);
    }
    acc6.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument1Hi(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("bin.ARGUMENT_1_HI already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument1Hi.put((byte) 0);
    }
    argument1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument1Lo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument1Lo.put((byte) 0);
    }
    argument1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument2Hi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument2Hi.put((byte) 0);
    }
    argument2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace argument2Lo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      argument2Lo.put((byte) 0);
    }
    argument2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("bin.BIT_1 already set");
    } else {
      filled.set(11);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitB4(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("bin.BIT_B_4 already set");
    } else {
      filled.set(12);
    }

    bitB4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("bin.BITS already set");
    } else {
      filled.set(10);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("bin.BYTE_1 already set");
    } else {
      filled.set(13);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("bin.BYTE_2 already set");
    } else {
      filled.set(14);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("bin.BYTE_3 already set");
    } else {
      filled.set(15);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("bin.BYTE_4 already set");
    } else {
      filled.set(16);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("bin.BYTE_5 already set");
    } else {
      filled.set(17);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace byte6(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("bin.BYTE_6 already set");
    } else {
      filled.set(18);
    }

    byte6.put(b.toByte());

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("bin.COUNTER already set");
    } else {
      filled.set(19);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("bin.CT_MAX already set");
    } else {
      filled.set(20);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("bin.INST already set");
    } else {
      filled.set(21);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isAnd(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("bin.IS_AND already set");
    } else {
      filled.set(22);
    }

    isAnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isByte(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("bin.IS_BYTE already set");
    } else {
      filled.set(23);
    }

    isByte.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isNot(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("bin.IS_NOT already set");
    } else {
      filled.set(24);
    }

    isNot.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isOr(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("bin.IS_OR already set");
    } else {
      filled.set(25);
    }

    isOr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSignextend(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("bin.IS_SIGNEXTEND already set");
    } else {
      filled.set(26);
    }

    isSignextend.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isXor(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("bin.IS_XOR already set");
    } else {
      filled.set(27);
    }

    isXor.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace low4(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("bin.LOW_4 already set");
    } else {
      filled.set(28);
    }

    low4.put(b.toByte());

    return this;
  }

  public Trace neg(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("bin.NEG already set");
    } else {
      filled.set(29);
    }

    neg.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pivot(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("bin.PIVOT already set");
    } else {
      filled.set(30);
    }

    pivot.put(b.toByte());

    return this;
  }

  public Trace resultHi(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("bin.RESULT_HI already set");
    } else {
      filled.set(31);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultHi.put((byte) 0);
    }
    resultHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resultLo(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("bin.RESULT_LO already set");
    } else {
      filled.set(32);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultLo.put((byte) 0);
    }
    resultLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace small(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("bin.SMALL already set");
    } else {
      filled.set(33);
    }

    small.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(34)) {
      throw new IllegalStateException("bin.STAMP already set");
    } else {
      filled.set(34);
    }

    stamp.putLong(b);

    return this;
  }

  public Trace xxxByteHi(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("bin.XXX_BYTE_HI already set");
    } else {
      filled.set(35);
    }

    xxxByteHi.put(b.toByte());

    return this;
  }

  public Trace xxxByteLo(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("bin.XXX_BYTE_LO already set");
    } else {
      filled.set(36);
    }

    xxxByteLo.put(b.toByte());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("bin.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("bin.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("bin.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("bin.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("bin.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("bin.ACC_6 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("bin.ARGUMENT_1_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("bin.ARGUMENT_1_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("bin.ARGUMENT_2_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("bin.ARGUMENT_2_LO has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("bin.BIT_1 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("bin.BIT_B_4 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("bin.BITS has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("bin.BYTE_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("bin.BYTE_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("bin.BYTE_3 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("bin.BYTE_4 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("bin.BYTE_5 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("bin.BYTE_6 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("bin.COUNTER has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("bin.CT_MAX has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("bin.INST has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("bin.IS_AND has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("bin.IS_BYTE has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("bin.IS_NOT has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("bin.IS_OR has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("bin.IS_SIGNEXTEND has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("bin.IS_XOR has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("bin.LOW_4 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("bin.NEG has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("bin.PIVOT has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("bin.RESULT_HI has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("bin.RESULT_LO has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("bin.SMALL has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("bin.STAMP has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("bin.XXX_BYTE_HI has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("bin.XXX_BYTE_LO has not been filled");
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
      acc6.position(acc6.position() + 32);
    }

    if (!filled.get(6)) {
      argument1Hi.position(argument1Hi.position() + 32);
    }

    if (!filled.get(7)) {
      argument1Lo.position(argument1Lo.position() + 32);
    }

    if (!filled.get(8)) {
      argument2Hi.position(argument2Hi.position() + 32);
    }

    if (!filled.get(9)) {
      argument2Lo.position(argument2Lo.position() + 32);
    }

    if (!filled.get(11)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(12)) {
      bitB4.position(bitB4.position() + 1);
    }

    if (!filled.get(10)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(13)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(14)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(15)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(16)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(17)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(18)) {
      byte6.position(byte6.position() + 1);
    }

    if (!filled.get(19)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(20)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(21)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(22)) {
      isAnd.position(isAnd.position() + 1);
    }

    if (!filled.get(23)) {
      isByte.position(isByte.position() + 1);
    }

    if (!filled.get(24)) {
      isNot.position(isNot.position() + 1);
    }

    if (!filled.get(25)) {
      isOr.position(isOr.position() + 1);
    }

    if (!filled.get(26)) {
      isSignextend.position(isSignextend.position() + 1);
    }

    if (!filled.get(27)) {
      isXor.position(isXor.position() + 1);
    }

    if (!filled.get(28)) {
      low4.position(low4.position() + 1);
    }

    if (!filled.get(29)) {
      neg.position(neg.position() + 1);
    }

    if (!filled.get(30)) {
      pivot.position(pivot.position() + 1);
    }

    if (!filled.get(31)) {
      resultHi.position(resultHi.position() + 32);
    }

    if (!filled.get(32)) {
      resultLo.position(resultLo.position() + 32);
    }

    if (!filled.get(33)) {
      small.position(small.position() + 1);
    }

    if (!filled.get(34)) {
      stamp.position(stamp.position() + 8);
    }

    if (!filled.get(35)) {
      xxxByteHi.position(xxxByteHi.position() + 1);
    }

    if (!filled.get(36)) {
      xxxByteLo.position(xxxByteLo.position() + 1);
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
