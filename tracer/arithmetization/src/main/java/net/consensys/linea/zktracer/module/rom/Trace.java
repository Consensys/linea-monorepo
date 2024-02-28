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

package net.consensys.linea.zktracer.module.rom;

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
  public static final int PHASE_BLAKE_DATA = 0x5;
  public static final int PHASE_BLAKE_PARAMS = 0x6;
  public static final int PHASE_BLAKE_RESULT = 0x7;
  public static final int PHASE_KECCAK_DATA = 0x8;
  public static final int PHASE_KECCAK_RESULT = 0x9;
  public static final int PHASE_MODEXP_BASE = 0x1;
  public static final int PHASE_MODEXP_EXPONENT = 0x2;
  public static final int PHASE_MODEXP_MODULUS = 0x3;
  public static final int PHASE_MODEXP_RESULT = 0x4;
  public static final int PHASE_RIPEMD_DATA = 0xc;
  public static final int PHASE_RIPEMD_RESULT = 0xd;
  public static final int PHASE_SHA2_DATA = 0xa;
  public static final int PHASE_SHA2_RESULT = 0xb;
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

  private final MappedByteBuffer acc;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexInfty;
  private final MappedByteBuffer codeSize;
  private final MappedByteBuffer codesizeReached;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer counterMax;
  private final MappedByteBuffer counterPush;
  private final MappedByteBuffer index;
  private final MappedByteBuffer isJumpdest;
  private final MappedByteBuffer isPush;
  private final MappedByteBuffer isPushData;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nBytesAcc;
  private final MappedByteBuffer opcode;
  private final MappedByteBuffer paddedBytecodeByte;
  private final MappedByteBuffer programCounter;
  private final MappedByteBuffer pushFunnelBit;
  private final MappedByteBuffer pushParameter;
  private final MappedByteBuffer pushValueAcc;
  private final MappedByteBuffer pushValueHi;
  private final MappedByteBuffer pushValueLo;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("rom.ACC", 32, length),
        new ColumnHeader("rom.CODE_FRAGMENT_INDEX", 8, length),
        new ColumnHeader("rom.CODE_FRAGMENT_INDEX_INFTY", 8, length),
        new ColumnHeader("rom.CODE_SIZE", 8, length),
        new ColumnHeader("rom.CODESIZE_REACHED", 1, length),
        new ColumnHeader("rom.COUNTER", 1, length),
        new ColumnHeader("rom.COUNTER_MAX", 1, length),
        new ColumnHeader("rom.COUNTER_PUSH", 1, length),
        new ColumnHeader("rom.INDEX", 8, length),
        new ColumnHeader("rom.IS_JUMPDEST", 1, length),
        new ColumnHeader("rom.IS_PUSH", 1, length),
        new ColumnHeader("rom.IS_PUSH_DATA", 1, length),
        new ColumnHeader("rom.LIMB", 32, length),
        new ColumnHeader("rom.nBYTES", 1, length),
        new ColumnHeader("rom.nBYTES_ACC", 1, length),
        new ColumnHeader("rom.OPCODE", 1, length),
        new ColumnHeader("rom.PADDED_BYTECODE_BYTE", 1, length),
        new ColumnHeader("rom.PROGRAM_COUNTER", 8, length),
        new ColumnHeader("rom.PUSH_FUNNEL_BIT", 1, length),
        new ColumnHeader("rom.PUSH_PARAMETER", 1, length),
        new ColumnHeader("rom.PUSH_VALUE_ACC", 32, length),
        new ColumnHeader("rom.PUSH_VALUE_HI", 32, length),
        new ColumnHeader("rom.PUSH_VALUE_LO", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc = buffers.get(0);
    this.codeFragmentIndex = buffers.get(1);
    this.codeFragmentIndexInfty = buffers.get(2);
    this.codeSize = buffers.get(3);
    this.codesizeReached = buffers.get(4);
    this.counter = buffers.get(5);
    this.counterMax = buffers.get(6);
    this.counterPush = buffers.get(7);
    this.index = buffers.get(8);
    this.isJumpdest = buffers.get(9);
    this.isPush = buffers.get(10);
    this.isPushData = buffers.get(11);
    this.limb = buffers.get(12);
    this.nBytes = buffers.get(13);
    this.nBytesAcc = buffers.get(14);
    this.opcode = buffers.get(15);
    this.paddedBytecodeByte = buffers.get(16);
    this.programCounter = buffers.get(17);
    this.pushFunnelBit = buffers.get(18);
    this.pushParameter = buffers.get(19);
    this.pushValueAcc = buffers.get(20);
    this.pushValueHi = buffers.get(21);
    this.pushValueLo = buffers.get(22);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rom.ACC already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc.put((byte) 0);
    }
    acc.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(2);
    }

    codeFragmentIndex.putLong(b);

    return this;
  }

  public Trace codeFragmentIndexInfty(final long b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX_INFTY already set");
    } else {
      filled.set(3);
    }

    codeFragmentIndexInfty.putLong(b);

    return this;
  }

  public Trace codeSize(final long b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rom.CODE_SIZE already set");
    } else {
      filled.set(4);
    }

    codeSize.putLong(b);

    return this;
  }

  public Trace codesizeReached(final Boolean b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rom.CODESIZE_REACHED already set");
    } else {
      filled.set(1);
    }

    codesizeReached.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rom.COUNTER already set");
    } else {
      filled.set(5);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace counterMax(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rom.COUNTER_MAX already set");
    } else {
      filled.set(6);
    }

    counterMax.put(b.toByte());

    return this;
  }

  public Trace counterPush(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rom.COUNTER_PUSH already set");
    } else {
      filled.set(7);
    }

    counterPush.put(b.toByte());

    return this;
  }

  public Trace index(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rom.INDEX already set");
    } else {
      filled.set(8);
    }

    index.putLong(b);

    return this;
  }

  public Trace isJumpdest(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rom.IS_JUMPDEST already set");
    } else {
      filled.set(9);
    }

    isJumpdest.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPush(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rom.IS_PUSH already set");
    } else {
      filled.set(10);
    }

    isPush.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPushData(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rom.IS_PUSH_DATA already set");
    } else {
      filled.set(11);
    }

    isPushData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rom.LIMB already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nBytes(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rom.nBYTES already set");
    } else {
      filled.set(21);
    }

    nBytes.put(b.toByte());

    return this;
  }

  public Trace nBytesAcc(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rom.nBYTES_ACC already set");
    } else {
      filled.set(22);
    }

    nBytesAcc.put(b.toByte());

    return this;
  }

  public Trace opcode(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rom.OPCODE already set");
    } else {
      filled.set(13);
    }

    opcode.put(b.toByte());

    return this;
  }

  public Trace paddedBytecodeByte(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rom.PADDED_BYTECODE_BYTE already set");
    } else {
      filled.set(14);
    }

    paddedBytecodeByte.put(b.toByte());

    return this;
  }

  public Trace programCounter(final long b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rom.PROGRAM_COUNTER already set");
    } else {
      filled.set(15);
    }

    programCounter.putLong(b);

    return this;
  }

  public Trace pushFunnelBit(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rom.PUSH_FUNNEL_BIT already set");
    } else {
      filled.set(16);
    }

    pushFunnelBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pushParameter(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rom.PUSH_PARAMETER already set");
    } else {
      filled.set(17);
    }

    pushParameter.put(b.toByte());

    return this;
  }

  public Trace pushValueAcc(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rom.PUSH_VALUE_ACC already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushValueAcc.put((byte) 0);
    }
    pushValueAcc.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pushValueHi(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rom.PUSH_VALUE_HI already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushValueHi.put((byte) 0);
    }
    pushValueHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pushValueLo(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rom.PUSH_VALUE_LO already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushValueLo.put((byte) 0);
    }
    pushValueLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rom.ACC has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX_INFTY has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rom.CODE_SIZE has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rom.CODESIZE_REACHED has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rom.COUNTER has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rom.COUNTER_MAX has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rom.COUNTER_PUSH has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rom.INDEX has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rom.IS_JUMPDEST has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rom.IS_PUSH has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rom.IS_PUSH_DATA has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rom.LIMB has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rom.nBYTES has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rom.nBYTES_ACC has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rom.OPCODE has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rom.PADDED_BYTECODE_BYTE has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rom.PROGRAM_COUNTER has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rom.PUSH_FUNNEL_BIT has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rom.PUSH_PARAMETER has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rom.PUSH_VALUE_ACC has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rom.PUSH_VALUE_HI has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rom.PUSH_VALUE_LO has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc.position(acc.position() + 32);
    }

    if (!filled.get(2)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 8);
    }

    if (!filled.get(3)) {
      codeFragmentIndexInfty.position(codeFragmentIndexInfty.position() + 8);
    }

    if (!filled.get(4)) {
      codeSize.position(codeSize.position() + 8);
    }

    if (!filled.get(1)) {
      codesizeReached.position(codesizeReached.position() + 1);
    }

    if (!filled.get(5)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(6)) {
      counterMax.position(counterMax.position() + 1);
    }

    if (!filled.get(7)) {
      counterPush.position(counterPush.position() + 1);
    }

    if (!filled.get(8)) {
      index.position(index.position() + 8);
    }

    if (!filled.get(9)) {
      isJumpdest.position(isJumpdest.position() + 1);
    }

    if (!filled.get(10)) {
      isPush.position(isPush.position() + 1);
    }

    if (!filled.get(11)) {
      isPushData.position(isPushData.position() + 1);
    }

    if (!filled.get(12)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(21)) {
      nBytes.position(nBytes.position() + 1);
    }

    if (!filled.get(22)) {
      nBytesAcc.position(nBytesAcc.position() + 1);
    }

    if (!filled.get(13)) {
      opcode.position(opcode.position() + 1);
    }

    if (!filled.get(14)) {
      paddedBytecodeByte.position(paddedBytecodeByte.position() + 1);
    }

    if (!filled.get(15)) {
      programCounter.position(programCounter.position() + 8);
    }

    if (!filled.get(16)) {
      pushFunnelBit.position(pushFunnelBit.position() + 1);
    }

    if (!filled.get(17)) {
      pushParameter.position(pushParameter.position() + 1);
    }

    if (!filled.get(18)) {
      pushValueAcc.position(pushValueAcc.position() + 32);
    }

    if (!filled.get(19)) {
      pushValueHi.position(pushValueHi.position() + 32);
    }

    if (!filled.get(20)) {
      pushValueLo.position(pushValueLo.position() + 32);
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
