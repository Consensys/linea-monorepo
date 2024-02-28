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
    return List.of(
        new ColumnHeader("stp.ARG_1_HI", 32, length),
        new ColumnHeader("stp.ARG_1_LO", 32, length),
        new ColumnHeader("stp.ARG_2_LO", 32, length),
        new ColumnHeader("stp.CT", 1, length),
        new ColumnHeader("stp.CT_MAX", 1, length),
        new ColumnHeader("stp.EXISTS", 1, length),
        new ColumnHeader("stp.EXOGENOUS_MODULE_INSTRUCTION", 1, length),
        new ColumnHeader("stp.GAS_ACTUAL", 32, length),
        new ColumnHeader("stp.GAS_HI", 32, length),
        new ColumnHeader("stp.GAS_LO", 32, length),
        new ColumnHeader("stp.GAS_MXP", 32, length),
        new ColumnHeader("stp.GAS_OUT_OF_POCKET", 32, length),
        new ColumnHeader("stp.GAS_STIPEND", 32, length),
        new ColumnHeader("stp.GAS_UPFRONT", 32, length),
        new ColumnHeader("stp.INSTRUCTION", 1, length),
        new ColumnHeader("stp.IS_CALL", 1, length),
        new ColumnHeader("stp.IS_CALLCODE", 1, length),
        new ColumnHeader("stp.IS_CREATE", 1, length),
        new ColumnHeader("stp.IS_CREATE2", 1, length),
        new ColumnHeader("stp.IS_DELEGATECALL", 1, length),
        new ColumnHeader("stp.IS_STATICCALL", 1, length),
        new ColumnHeader("stp.MOD_FLAG", 1, length),
        new ColumnHeader("stp.OUT_OF_GAS_EXCEPTION", 1, length),
        new ColumnHeader("stp.RES_LO", 32, length),
        new ColumnHeader("stp.STAMP", 4, length),
        new ColumnHeader("stp.VAL_HI", 32, length),
        new ColumnHeader("stp.VAL_LO", 32, length),
        new ColumnHeader("stp.WARM", 1, length),
        new ColumnHeader("stp.WCP_FLAG", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
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

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Hi.put((byte) 0);
    }
    arg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("stp.ARG_1_LO already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Lo.put((byte) 0);
    }
    arg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("stp.ARG_2_LO already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Lo.put((byte) 0);
    }
    arg2Lo.put(b.toArrayUnsafe());

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

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasActual.put((byte) 0);
    }
    gasActual.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasHi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("stp.GAS_HI already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasHi.put((byte) 0);
    }
    gasHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasLo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("stp.GAS_LO already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasLo.put((byte) 0);
    }
    gasLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasMxp(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("stp.GAS_MXP already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasMxp.put((byte) 0);
    }
    gasMxp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasOutOfPocket(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("stp.GAS_OUT_OF_POCKET already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasOutOfPocket.put((byte) 0);
    }
    gasOutOfPocket.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasStipend(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("stp.GAS_STIPEND already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasStipend.put((byte) 0);
    }
    gasStipend.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasUpfront(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("stp.GAS_UPFRONT already set");
    } else {
      filled.set(13);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasUpfront.put((byte) 0);
    }
    gasUpfront.put(b.toArrayUnsafe());

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

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resLo.put((byte) 0);
    }
    resLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final int b) {
    if (filled.get(24)) {
      throw new IllegalStateException("stp.STAMP already set");
    } else {
      filled.set(24);
    }

    stamp.putInt(b);

    return this;
  }

  public Trace valHi(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("stp.VAL_HI already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valHi.put((byte) 0);
    }
    valHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valLo(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("stp.VAL_LO already set");
    } else {
      filled.set(26);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valLo.put((byte) 0);
    }
    valLo.put(b.toArrayUnsafe());

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
      arg1Hi.position(arg1Hi.position() + 32);
    }

    if (!filled.get(1)) {
      arg1Lo.position(arg1Lo.position() + 32);
    }

    if (!filled.get(2)) {
      arg2Lo.position(arg2Lo.position() + 32);
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
      gasActual.position(gasActual.position() + 32);
    }

    if (!filled.get(8)) {
      gasHi.position(gasHi.position() + 32);
    }

    if (!filled.get(9)) {
      gasLo.position(gasLo.position() + 32);
    }

    if (!filled.get(10)) {
      gasMxp.position(gasMxp.position() + 32);
    }

    if (!filled.get(11)) {
      gasOutOfPocket.position(gasOutOfPocket.position() + 32);
    }

    if (!filled.get(12)) {
      gasStipend.position(gasStipend.position() + 32);
    }

    if (!filled.get(13)) {
      gasUpfront.position(gasUpfront.position() + 32);
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
      resLo.position(resLo.position() + 32);
    }

    if (!filled.get(24)) {
      stamp.position(stamp.position() + 4);
    }

    if (!filled.get(25)) {
      valHi.position(valHi.position() + 32);
    }

    if (!filled.get(26)) {
      valLo.position(valLo.position() + 32);
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
