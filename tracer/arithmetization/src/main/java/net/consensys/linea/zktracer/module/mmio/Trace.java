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

package net.consensys.linea.zktracer.module.mmio;

import java.math.BigInteger;
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
  public static final BigInteger EMPTY_KECCAK_HI =
      new BigInteger("16434357337474432580558001204043214908");
  public static final BigInteger EMPTY_KECCAK_LO =
      new BigInteger("19024806816994025362060938983270537799");
  public static final int EMPTY_RIPEMD_HI = 0x9c1185a;
  public static final BigInteger EMPTY_RIPEMD_LO =
      new BigInteger("16442052386882578548602430796343695571");
  public static final BigInteger EMPTY_SHA2_HI =
      new BigInteger("18915786244935348617899154533661473682");
  public static final BigInteger EMPTY_SHA2_LO =
      new BigInteger("3296542996298665609207448061432114053");
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
  public static final int EXO_SUM_INDEX_BLAKEMODEXP = 0x6;
  public static final int EXO_SUM_INDEX_ECDATA = 0x4;
  public static final int EXO_SUM_INDEX_KEC = 0x1;
  public static final int EXO_SUM_INDEX_LOG = 0x2;
  public static final int EXO_SUM_INDEX_RIPSHA = 0x5;
  public static final int EXO_SUM_INDEX_ROM = 0x0;
  public static final int EXO_SUM_INDEX_TXCD = 0x3;
  public static final int EXO_SUM_WEIGHT_BLAKEMODEXP = 0x40;
  public static final int EXO_SUM_WEIGHT_ECDATA = 0x10;
  public static final int EXO_SUM_WEIGHT_KEC = 0x2;
  public static final int EXO_SUM_WEIGHT_LOG = 0x4;
  public static final int EXO_SUM_WEIGHT_RIPSHA = 0x20;
  public static final int EXO_SUM_WEIGHT_ROM = 0x1;
  public static final int EXO_SUM_WEIGHT_TXCD = 0x8;
  public static final int EXP_INST_EXPLOG = 0xee0a;
  public static final int EXP_INST_MODEXPLOG = 0xee05;
  public static final int GAS_CONST_G_ACCESS_LIST_ADRESS = 0x960;
  public static final int GAS_CONST_G_ACCESS_LIST_STORAGE = 0x76c;
  public static final int GAS_CONST_G_BASE = 0x2;
  public static final int GAS_CONST_G_BLOCKHASH = 0x14;
  public static final int GAS_CONST_G_CALL_STIPEND = 0x8fc;
  public static final int GAS_CONST_G_CALL_VALUE = 0x2328;
  public static final int GAS_CONST_G_CODE_DEPOSIT = 0xc8;
  public static final int GAS_CONST_G_COLD_ACCOUNT_ACCESS = 0xa28;
  public static final int GAS_CONST_G_COLD_SLOAD = 0x834;
  public static final int GAS_CONST_G_COPY = 0x3;
  public static final int GAS_CONST_G_CREATE = 0x7d00;
  public static final int GAS_CONST_G_EXP = 0xa;
  public static final int GAS_CONST_G_EXP_BYTE = 0x32;
  public static final int GAS_CONST_G_HIGH = 0xa;
  public static final int GAS_CONST_G_JUMPDEST = 0x1;
  public static final int GAS_CONST_G_KECCAK_256 = 0x1e;
  public static final int GAS_CONST_G_KECCAK_256_WORD = 0x6;
  public static final int GAS_CONST_G_LOG = 0x177;
  public static final int GAS_CONST_G_LOG_DATA = 0x8;
  public static final int GAS_CONST_G_LOG_TOPIC = 0x177;
  public static final int GAS_CONST_G_LOW = 0x5;
  public static final int GAS_CONST_G_MEMORY = 0x3;
  public static final int GAS_CONST_G_MID = 0x8;
  public static final int GAS_CONST_G_NEW_ACCOUNT = 0x61a8;
  public static final int GAS_CONST_G_SELFDESTRUCT = 0x1388;
  public static final int GAS_CONST_G_SRESET = 0xb54;
  public static final int GAS_CONST_G_SSET = 0x4e20;
  public static final int GAS_CONST_G_TRANSACTION = 0x5208;
  public static final int GAS_CONST_G_TX_CREATE = 0x7d00;
  public static final int GAS_CONST_G_TX_DATA_NONZERO = 0x10;
  public static final int GAS_CONST_G_TX_DATA_ZERO = 0x4;
  public static final int GAS_CONST_G_VERY_LOW = 0x3;
  public static final int GAS_CONST_G_WARM_ACCESS = 0x64;
  public static final int GAS_CONST_G_ZERO = 0x0;
  public static final int INVALID_CODE_PREFIX_VALUE = 0xef;
  public static final int LLARGE = 0x10;
  public static final int LLARGEMO = 0xf;
  public static final int LLARGEPO = 0x11;
  public static final int MISC_EXP_WEIGHT = 0x1;
  public static final int MISC_MMU_WEIGHT = 0x2;
  public static final int MISC_MXP_WEIGHT = 0x4;
  public static final int MISC_OOB_WEIGHT = 0x8;
  public static final int MISC_STP_WEIGHT = 0x10;
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
  public static final int OOB_INST_blake_cds = 0xfa09;
  public static final int OOB_INST_blake_params = 0xfb09;
  public static final int OOB_INST_call = 0xca;
  public static final int OOB_INST_cdl = 0x35;
  public static final int OOB_INST_create = 0xce;
  public static final int OOB_INST_deployment = 0xf3;
  public static final int OOB_INST_ecadd = 0xff06;
  public static final int OOB_INST_ecmul = 0xff07;
  public static final int OOB_INST_ecpairing = 0xff08;
  public static final int OOB_INST_ecrecover = 0xff01;
  public static final int OOB_INST_identity = 0xff04;
  public static final int OOB_INST_jump = 0x56;
  public static final int OOB_INST_jumpi = 0x57;
  public static final int OOB_INST_modexp_cds = 0xfa05;
  public static final int OOB_INST_modexp_extract = 0xfe05;
  public static final int OOB_INST_modexp_lead = 0xfc05;
  public static final int OOB_INST_modexp_pricing = 0xfd05;
  public static final int OOB_INST_modexp_xbs = 0xfb05;
  public static final int OOB_INST_rdc = 0x3e;
  public static final int OOB_INST_ripemd = 0xff03;
  public static final int OOB_INST_sha2 = 0xff02;
  public static final int OOB_INST_sstore = 0x55;
  public static final int OOB_INST_xcall = 0xcc;
  public static final int PHASE_BLAKE_DATA = 0x5;
  public static final int PHASE_BLAKE_PARAMS = 0x6;
  public static final int PHASE_BLAKE_RESULT = 0x7;
  public static final int PHASE_KECCAK_DATA = 0x5;
  public static final int PHASE_KECCAK_RESULT = 0x6;
  public static final int PHASE_MODEXP_BASE = 0x1;
  public static final int PHASE_MODEXP_EXPONENT = 0x2;
  public static final int PHASE_MODEXP_MODULUS = 0x3;
  public static final int PHASE_MODEXP_RESULT = 0x4;
  public static final int PHASE_RIPEMD_DATA = 0x3;
  public static final int PHASE_RIPEMD_RESULT = 0x4;
  public static final int PHASE_SHA2_DATA = 0x1;
  public static final int PHASE_SHA2_RESULT = 0x2;
  public static final int REFUND_CONST_R_SCLEAR = 0x3a98;
  public static final int REFUND_CONST_R_SELFDESTRUCT = 0x5dc0;
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
  private final MappedByteBuffer accB;
  private final MappedByteBuffer accC;
  private final MappedByteBuffer accLimb;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
  private final MappedByteBuffer bit5;
  private final MappedByteBuffer byteA;
  private final MappedByteBuffer byteB;
  private final MappedByteBuffer byteC;
  private final MappedByteBuffer byteLimb;
  private final MappedByteBuffer cnA;
  private final MappedByteBuffer cnB;
  private final MappedByteBuffer cnC;
  private final MappedByteBuffer contextSource;
  private final MappedByteBuffer contextTarget;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer exoId;
  private final MappedByteBuffer exoIsBlakemodexp;
  private final MappedByteBuffer exoIsEcdata;
  private final MappedByteBuffer exoIsKec;
  private final MappedByteBuffer exoIsLog;
  private final MappedByteBuffer exoIsRipsha;
  private final MappedByteBuffer exoIsRom;
  private final MappedByteBuffer exoIsTxcd;
  private final MappedByteBuffer exoSum;
  private final MappedByteBuffer fast;
  private final MappedByteBuffer indexA;
  private final MappedByteBuffer indexB;
  private final MappedByteBuffer indexC;
  private final MappedByteBuffer indexX;
  private final MappedByteBuffer isLimbToRamOneTarget;
  private final MappedByteBuffer isLimbToRamTransplant;
  private final MappedByteBuffer isLimbToRamTwoTarget;
  private final MappedByteBuffer isLimbVanishes;
  private final MappedByteBuffer isRamExcision;
  private final MappedByteBuffer isRamToLimbOneSource;
  private final MappedByteBuffer isRamToLimbTransplant;
  private final MappedByteBuffer isRamToLimbTwoSource;
  private final MappedByteBuffer isRamToRamPartial;
  private final MappedByteBuffer isRamToRamTransplant;
  private final MappedByteBuffer isRamToRamTwoSource;
  private final MappedByteBuffer isRamToRamTwoTarget;
  private final MappedByteBuffer isRamVanishes;
  private final MappedByteBuffer kecId;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer mmioInstruction;
  private final MappedByteBuffer mmioStamp;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer pow2561;
  private final MappedByteBuffer pow2562;
  private final MappedByteBuffer size;
  private final MappedByteBuffer slow;
  private final MappedByteBuffer sourceByteOffset;
  private final MappedByteBuffer sourceLimbOffset;
  private final MappedByteBuffer successBit;
  private final MappedByteBuffer targetByteOffset;
  private final MappedByteBuffer targetLimbOffset;
  private final MappedByteBuffer totalSize;
  private final MappedByteBuffer valA;
  private final MappedByteBuffer valANew;
  private final MappedByteBuffer valB;
  private final MappedByteBuffer valBNew;
  private final MappedByteBuffer valC;
  private final MappedByteBuffer valCNew;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("mmio.ACC_1", 32, length),
        new ColumnHeader("mmio.ACC_2", 32, length),
        new ColumnHeader("mmio.ACC_3", 32, length),
        new ColumnHeader("mmio.ACC_4", 32, length),
        new ColumnHeader("mmio.ACC_A", 32, length),
        new ColumnHeader("mmio.ACC_B", 32, length),
        new ColumnHeader("mmio.ACC_C", 32, length),
        new ColumnHeader("mmio.ACC_LIMB", 32, length),
        new ColumnHeader("mmio.BIT_1", 1, length),
        new ColumnHeader("mmio.BIT_2", 1, length),
        new ColumnHeader("mmio.BIT_3", 1, length),
        new ColumnHeader("mmio.BIT_4", 1, length),
        new ColumnHeader("mmio.BIT_5", 1, length),
        new ColumnHeader("mmio.BYTE_A", 1, length),
        new ColumnHeader("mmio.BYTE_B", 1, length),
        new ColumnHeader("mmio.BYTE_C", 1, length),
        new ColumnHeader("mmio.BYTE_LIMB", 1, length),
        new ColumnHeader("mmio.CN_A", 32, length),
        new ColumnHeader("mmio.CN_B", 32, length),
        new ColumnHeader("mmio.CN_C", 32, length),
        new ColumnHeader("mmio.CONTEXT_SOURCE", 32, length),
        new ColumnHeader("mmio.CONTEXT_TARGET", 32, length),
        new ColumnHeader("mmio.COUNTER", 2, length),
        new ColumnHeader("mmio.EXO_ID", 8, length),
        new ColumnHeader("mmio.EXO_IS_BLAKEMODEXP", 1, length),
        new ColumnHeader("mmio.EXO_IS_ECDATA", 1, length),
        new ColumnHeader("mmio.EXO_IS_KEC", 1, length),
        new ColumnHeader("mmio.EXO_IS_LOG", 1, length),
        new ColumnHeader("mmio.EXO_IS_RIPSHA", 1, length),
        new ColumnHeader("mmio.EXO_IS_ROM", 1, length),
        new ColumnHeader("mmio.EXO_IS_TXCD", 1, length),
        new ColumnHeader("mmio.EXO_SUM", 8, length),
        new ColumnHeader("mmio.FAST", 1, length),
        new ColumnHeader("mmio.INDEX_A", 32, length),
        new ColumnHeader("mmio.INDEX_B", 32, length),
        new ColumnHeader("mmio.INDEX_C", 32, length),
        new ColumnHeader("mmio.INDEX_X", 32, length),
        new ColumnHeader("mmio.IS_LIMB_TO_RAM_ONE_TARGET", 1, length),
        new ColumnHeader("mmio.IS_LIMB_TO_RAM_TRANSPLANT", 1, length),
        new ColumnHeader("mmio.IS_LIMB_TO_RAM_TWO_TARGET", 1, length),
        new ColumnHeader("mmio.IS_LIMB_VANISHES", 1, length),
        new ColumnHeader("mmio.IS_RAM_EXCISION", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_LIMB_ONE_SOURCE", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_LIMB_TRANSPLANT", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_LIMB_TWO_SOURCE", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_RAM_PARTIAL", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_RAM_TRANSPLANT", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_RAM_TWO_SOURCE", 1, length),
        new ColumnHeader("mmio.IS_RAM_TO_RAM_TWO_TARGET", 1, length),
        new ColumnHeader("mmio.IS_RAM_VANISHES", 1, length),
        new ColumnHeader("mmio.KEC_ID", 8, length),
        new ColumnHeader("mmio.LIMB", 32, length),
        new ColumnHeader("mmio.MMIO_INSTRUCTION", 4, length),
        new ColumnHeader("mmio.MMIO_STAMP", 8, length),
        new ColumnHeader("mmio.PHASE", 8, length),
        new ColumnHeader("mmio.POW_256_1", 32, length),
        new ColumnHeader("mmio.POW_256_2", 32, length),
        new ColumnHeader("mmio.SIZE", 32, length),
        new ColumnHeader("mmio.SLOW", 1, length),
        new ColumnHeader("mmio.SOURCE_BYTE_OFFSET", 2, length),
        new ColumnHeader("mmio.SOURCE_LIMB_OFFSET", 32, length),
        new ColumnHeader("mmio.SUCCESS_BIT", 1, length),
        new ColumnHeader("mmio.TARGET_BYTE_OFFSET", 2, length),
        new ColumnHeader("mmio.TARGET_LIMB_OFFSET", 32, length),
        new ColumnHeader("mmio.TOTAL_SIZE", 32, length),
        new ColumnHeader("mmio.VAL_A", 32, length),
        new ColumnHeader("mmio.VAL_A_NEW", 32, length),
        new ColumnHeader("mmio.VAL_B", 32, length),
        new ColumnHeader("mmio.VAL_B_NEW", 32, length),
        new ColumnHeader("mmio.VAL_C", 32, length),
        new ColumnHeader("mmio.VAL_C_NEW", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.accA = buffers.get(4);
    this.accB = buffers.get(5);
    this.accC = buffers.get(6);
    this.accLimb = buffers.get(7);
    this.bit1 = buffers.get(8);
    this.bit2 = buffers.get(9);
    this.bit3 = buffers.get(10);
    this.bit4 = buffers.get(11);
    this.bit5 = buffers.get(12);
    this.byteA = buffers.get(13);
    this.byteB = buffers.get(14);
    this.byteC = buffers.get(15);
    this.byteLimb = buffers.get(16);
    this.cnA = buffers.get(17);
    this.cnB = buffers.get(18);
    this.cnC = buffers.get(19);
    this.contextSource = buffers.get(20);
    this.contextTarget = buffers.get(21);
    this.counter = buffers.get(22);
    this.exoId = buffers.get(23);
    this.exoIsBlakemodexp = buffers.get(24);
    this.exoIsEcdata = buffers.get(25);
    this.exoIsKec = buffers.get(26);
    this.exoIsLog = buffers.get(27);
    this.exoIsRipsha = buffers.get(28);
    this.exoIsRom = buffers.get(29);
    this.exoIsTxcd = buffers.get(30);
    this.exoSum = buffers.get(31);
    this.fast = buffers.get(32);
    this.indexA = buffers.get(33);
    this.indexB = buffers.get(34);
    this.indexC = buffers.get(35);
    this.indexX = buffers.get(36);
    this.isLimbToRamOneTarget = buffers.get(37);
    this.isLimbToRamTransplant = buffers.get(38);
    this.isLimbToRamTwoTarget = buffers.get(39);
    this.isLimbVanishes = buffers.get(40);
    this.isRamExcision = buffers.get(41);
    this.isRamToLimbOneSource = buffers.get(42);
    this.isRamToLimbTransplant = buffers.get(43);
    this.isRamToLimbTwoSource = buffers.get(44);
    this.isRamToRamPartial = buffers.get(45);
    this.isRamToRamTransplant = buffers.get(46);
    this.isRamToRamTwoSource = buffers.get(47);
    this.isRamToRamTwoTarget = buffers.get(48);
    this.isRamVanishes = buffers.get(49);
    this.kecId = buffers.get(50);
    this.limb = buffers.get(51);
    this.mmioInstruction = buffers.get(52);
    this.mmioStamp = buffers.get(53);
    this.phase = buffers.get(54);
    this.pow2561 = buffers.get(55);
    this.pow2562 = buffers.get(56);
    this.size = buffers.get(57);
    this.slow = buffers.get(58);
    this.sourceByteOffset = buffers.get(59);
    this.sourceLimbOffset = buffers.get(60);
    this.successBit = buffers.get(61);
    this.targetByteOffset = buffers.get(62);
    this.targetLimbOffset = buffers.get(63);
    this.totalSize = buffers.get(64);
    this.valA = buffers.get(65);
    this.valANew = buffers.get(66);
    this.valB = buffers.get(67);
    this.valBNew = buffers.get(68);
    this.valC = buffers.get(69);
    this.valCNew = buffers.get(70);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mmio.ACC_1 already set");
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
      throw new IllegalStateException("mmio.ACC_2 already set");
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
      throw new IllegalStateException("mmio.ACC_3 already set");
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
      throw new IllegalStateException("mmio.ACC_4 already set");
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
      throw new IllegalStateException("mmio.ACC_A already set");
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

  public Trace accB(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mmio.ACC_B already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB.put((byte) 0);
    }
    accB.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accC(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mmio.ACC_C already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accC.put((byte) 0);
    }
    accC.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accLimb(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mmio.ACC_LIMB already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accLimb.put((byte) 0);
    }
    accLimb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mmio.BIT_1 already set");
    } else {
      filled.set(8);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mmio.BIT_2 already set");
    } else {
      filled.set(9);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mmio.BIT_3 already set");
    } else {
      filled.set(10);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mmio.BIT_4 already set");
    } else {
      filled.set(11);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit5(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mmio.BIT_5 already set");
    } else {
      filled.set(12);
    }

    bit5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byteA(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mmio.BYTE_A already set");
    } else {
      filled.set(13);
    }

    byteA.put(b.toByte());

    return this;
  }

  public Trace byteB(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mmio.BYTE_B already set");
    } else {
      filled.set(14);
    }

    byteB.put(b.toByte());

    return this;
  }

  public Trace byteC(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mmio.BYTE_C already set");
    } else {
      filled.set(15);
    }

    byteC.put(b.toByte());

    return this;
  }

  public Trace byteLimb(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mmio.BYTE_LIMB already set");
    } else {
      filled.set(16);
    }

    byteLimb.put(b.toByte());

    return this;
  }

  public Trace cnA(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mmio.CN_A already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cnA.put((byte) 0);
    }
    cnA.put(b.toArrayUnsafe());

    return this;
  }

  public Trace cnB(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mmio.CN_B already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cnB.put((byte) 0);
    }
    cnB.put(b.toArrayUnsafe());

    return this;
  }

  public Trace cnC(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mmio.CN_C already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      cnC.put((byte) 0);
    }
    cnC.put(b.toArrayUnsafe());

    return this;
  }

  public Trace contextSource(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mmio.CONTEXT_SOURCE already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      contextSource.put((byte) 0);
    }
    contextSource.put(b.toArrayUnsafe());

    return this;
  }

  public Trace contextTarget(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mmio.CONTEXT_TARGET already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      contextTarget.put((byte) 0);
    }
    contextTarget.put(b.toArrayUnsafe());

    return this;
  }

  public Trace counter(final short b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mmio.COUNTER already set");
    } else {
      filled.set(22);
    }

    counter.putShort(b);

    return this;
  }

  public Trace exoId(final long b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mmio.EXO_ID already set");
    } else {
      filled.set(23);
    }

    exoId.putLong(b);

    return this;
  }

  public Trace exoIsBlakemodexp(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mmio.EXO_IS_BLAKEMODEXP already set");
    } else {
      filled.set(24);
    }

    exoIsBlakemodexp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsEcdata(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mmio.EXO_IS_ECDATA already set");
    } else {
      filled.set(25);
    }

    exoIsEcdata.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsKec(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mmio.EXO_IS_KEC already set");
    } else {
      filled.set(26);
    }

    exoIsKec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsLog(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mmio.EXO_IS_LOG already set");
    } else {
      filled.set(27);
    }

    exoIsLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsRipsha(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mmio.EXO_IS_RIPSHA already set");
    } else {
      filled.set(28);
    }

    exoIsRipsha.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsRom(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mmio.EXO_IS_ROM already set");
    } else {
      filled.set(29);
    }

    exoIsRom.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsTxcd(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mmio.EXO_IS_TXCD already set");
    } else {
      filled.set(30);
    }

    exoIsTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoSum(final long b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mmio.EXO_SUM already set");
    } else {
      filled.set(31);
    }

    exoSum.putLong(b);

    return this;
  }

  public Trace fast(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mmio.FAST already set");
    } else {
      filled.set(32);
    }

    fast.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace indexA(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mmio.INDEX_A already set");
    } else {
      filled.set(33);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      indexA.put((byte) 0);
    }
    indexA.put(b.toArrayUnsafe());

    return this;
  }

  public Trace indexB(final Bytes b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mmio.INDEX_B already set");
    } else {
      filled.set(34);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      indexB.put((byte) 0);
    }
    indexB.put(b.toArrayUnsafe());

    return this;
  }

  public Trace indexC(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mmio.INDEX_C already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      indexC.put((byte) 0);
    }
    indexC.put(b.toArrayUnsafe());

    return this;
  }

  public Trace indexX(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mmio.INDEX_X already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      indexX.put((byte) 0);
    }
    indexX.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isLimbToRamOneTarget(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_ONE_TARGET already set");
    } else {
      filled.set(37);
    }

    isLimbToRamOneTarget.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLimbToRamTransplant(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TRANSPLANT already set");
    } else {
      filled.set(38);
    }

    isLimbToRamTransplant.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLimbToRamTwoTarget(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TWO_TARGET already set");
    } else {
      filled.set(39);
    }

    isLimbToRamTwoTarget.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLimbVanishes(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mmio.IS_LIMB_VANISHES already set");
    } else {
      filled.set(40);
    }

    isLimbVanishes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamExcision(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mmio.IS_RAM_EXCISION already set");
    } else {
      filled.set(41);
    }

    isRamExcision.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToLimbOneSource(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_ONE_SOURCE already set");
    } else {
      filled.set(42);
    }

    isRamToLimbOneSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToLimbTransplant(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TRANSPLANT already set");
    } else {
      filled.set(43);
    }

    isRamToLimbTransplant.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToLimbTwoSource(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TWO_SOURCE already set");
    } else {
      filled.set(44);
    }

    isRamToLimbTwoSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamPartial(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_PARTIAL already set");
    } else {
      filled.set(45);
    }

    isRamToRamPartial.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamTransplant(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TRANSPLANT already set");
    } else {
      filled.set(46);
    }

    isRamToRamTransplant.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamTwoSource(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_SOURCE already set");
    } else {
      filled.set(47);
    }

    isRamToRamTwoSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamTwoTarget(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_TARGET already set");
    } else {
      filled.set(48);
    }

    isRamToRamTwoTarget.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamVanishes(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmio.IS_RAM_VANISHES already set");
    } else {
      filled.set(49);
    }

    isRamVanishes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace kecId(final long b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mmio.KEC_ID already set");
    } else {
      filled.set(50);
    }

    kecId.putLong(b);

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mmio.LIMB already set");
    } else {
      filled.set(51);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace mmioInstruction(final int b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmio.MMIO_INSTRUCTION already set");
    } else {
      filled.set(52);
    }

    mmioInstruction.putInt(b);

    return this;
  }

  public Trace mmioStamp(final long b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmio.MMIO_STAMP already set");
    } else {
      filled.set(53);
    }

    mmioStamp.putLong(b);

    return this;
  }

  public Trace phase(final long b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mmio.PHASE already set");
    } else {
      filled.set(54);
    }

    phase.putLong(b);

    return this;
  }

  public Trace pow2561(final Bytes b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mmio.POW_256_1 already set");
    } else {
      filled.set(55);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pow2561.put((byte) 0);
    }
    pow2561.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pow2562(final Bytes b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmio.POW_256_2 already set");
    } else {
      filled.set(56);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pow2562.put((byte) 0);
    }
    pow2562.put(b.toArrayUnsafe());

    return this;
  }

  public Trace size(final Bytes b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mmio.SIZE already set");
    } else {
      filled.set(57);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      size.put((byte) 0);
    }
    size.put(b.toArrayUnsafe());

    return this;
  }

  public Trace slow(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mmio.SLOW already set");
    } else {
      filled.set(58);
    }

    slow.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace sourceByteOffset(final short b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmio.SOURCE_BYTE_OFFSET already set");
    } else {
      filled.set(59);
    }

    sourceByteOffset.putShort(b);

    return this;
  }

  public Trace sourceLimbOffset(final Bytes b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mmio.SOURCE_LIMB_OFFSET already set");
    } else {
      filled.set(60);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      sourceLimbOffset.put((byte) 0);
    }
    sourceLimbOffset.put(b.toArrayUnsafe());

    return this;
  }

  public Trace successBit(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mmio.SUCCESS_BIT already set");
    } else {
      filled.set(61);
    }

    successBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace targetByteOffset(final short b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmio.TARGET_BYTE_OFFSET already set");
    } else {
      filled.set(62);
    }

    targetByteOffset.putShort(b);

    return this;
  }

  public Trace targetLimbOffset(final Bytes b) {
    if (filled.get(63)) {
      throw new IllegalStateException("mmio.TARGET_LIMB_OFFSET already set");
    } else {
      filled.set(63);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      targetLimbOffset.put((byte) 0);
    }
    targetLimbOffset.put(b.toArrayUnsafe());

    return this;
  }

  public Trace totalSize(final Bytes b) {
    if (filled.get(64)) {
      throw new IllegalStateException("mmio.TOTAL_SIZE already set");
    } else {
      filled.set(64);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      totalSize.put((byte) 0);
    }
    totalSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valA(final Bytes b) {
    if (filled.get(65)) {
      throw new IllegalStateException("mmio.VAL_A already set");
    } else {
      filled.set(65);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valA.put((byte) 0);
    }
    valA.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valANew(final Bytes b) {
    if (filled.get(66)) {
      throw new IllegalStateException("mmio.VAL_A_NEW already set");
    } else {
      filled.set(66);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valANew.put((byte) 0);
    }
    valANew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valB(final Bytes b) {
    if (filled.get(67)) {
      throw new IllegalStateException("mmio.VAL_B already set");
    } else {
      filled.set(67);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valB.put((byte) 0);
    }
    valB.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valBNew(final Bytes b) {
    if (filled.get(68)) {
      throw new IllegalStateException("mmio.VAL_B_NEW already set");
    } else {
      filled.set(68);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valBNew.put((byte) 0);
    }
    valBNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valC(final Bytes b) {
    if (filled.get(69)) {
      throw new IllegalStateException("mmio.VAL_C already set");
    } else {
      filled.set(69);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valC.put((byte) 0);
    }
    valC.put(b.toArrayUnsafe());

    return this;
  }

  public Trace valCNew(final Bytes b) {
    if (filled.get(70)) {
      throw new IllegalStateException("mmio.VAL_C_NEW already set");
    } else {
      filled.set(70);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      valCNew.put((byte) 0);
    }
    valCNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mmio.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mmio.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mmio.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mmio.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mmio.ACC_A has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mmio.ACC_B has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mmio.ACC_C has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mmio.ACC_LIMB has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mmio.BIT_1 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mmio.BIT_2 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mmio.BIT_3 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mmio.BIT_4 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mmio.BIT_5 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mmio.BYTE_A has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mmio.BYTE_B has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mmio.BYTE_C has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mmio.BYTE_LIMB has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mmio.CN_A has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mmio.CN_B has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mmio.CN_C has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mmio.CONTEXT_SOURCE has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mmio.CONTEXT_TARGET has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mmio.COUNTER has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mmio.EXO_ID has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mmio.EXO_IS_BLAKEMODEXP has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mmio.EXO_IS_ECDATA has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mmio.EXO_IS_KEC has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mmio.EXO_IS_LOG has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mmio.EXO_IS_RIPSHA has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mmio.EXO_IS_ROM has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mmio.EXO_IS_TXCD has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mmio.EXO_SUM has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mmio.FAST has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mmio.INDEX_A has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mmio.INDEX_B has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mmio.INDEX_C has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mmio.INDEX_X has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_ONE_TARGET has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TRANSPLANT has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TWO_TARGET has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mmio.IS_LIMB_VANISHES has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mmio.IS_RAM_EXCISION has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_ONE_SOURCE has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TRANSPLANT has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TWO_SOURCE has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_PARTIAL has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TRANSPLANT has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_SOURCE has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_TARGET has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mmio.IS_RAM_VANISHES has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mmio.KEC_ID has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mmio.LIMB has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("mmio.MMIO_INSTRUCTION has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("mmio.MMIO_STAMP has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mmio.PHASE has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mmio.POW_256_1 has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mmio.POW_256_2 has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mmio.SIZE has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mmio.SLOW has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("mmio.SOURCE_BYTE_OFFSET has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mmio.SOURCE_LIMB_OFFSET has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("mmio.SUCCESS_BIT has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("mmio.TARGET_BYTE_OFFSET has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("mmio.TARGET_LIMB_OFFSET has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("mmio.TOTAL_SIZE has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("mmio.VAL_A has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException("mmio.VAL_A_NEW has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException("mmio.VAL_B has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException("mmio.VAL_B_NEW has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException("mmio.VAL_C has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException("mmio.VAL_C_NEW has not been filled");
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
      accB.position(accB.position() + 32);
    }

    if (!filled.get(6)) {
      accC.position(accC.position() + 32);
    }

    if (!filled.get(7)) {
      accLimb.position(accLimb.position() + 32);
    }

    if (!filled.get(8)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(9)) {
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(10)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(11)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(12)) {
      bit5.position(bit5.position() + 1);
    }

    if (!filled.get(13)) {
      byteA.position(byteA.position() + 1);
    }

    if (!filled.get(14)) {
      byteB.position(byteB.position() + 1);
    }

    if (!filled.get(15)) {
      byteC.position(byteC.position() + 1);
    }

    if (!filled.get(16)) {
      byteLimb.position(byteLimb.position() + 1);
    }

    if (!filled.get(17)) {
      cnA.position(cnA.position() + 32);
    }

    if (!filled.get(18)) {
      cnB.position(cnB.position() + 32);
    }

    if (!filled.get(19)) {
      cnC.position(cnC.position() + 32);
    }

    if (!filled.get(20)) {
      contextSource.position(contextSource.position() + 32);
    }

    if (!filled.get(21)) {
      contextTarget.position(contextTarget.position() + 32);
    }

    if (!filled.get(22)) {
      counter.position(counter.position() + 2);
    }

    if (!filled.get(23)) {
      exoId.position(exoId.position() + 8);
    }

    if (!filled.get(24)) {
      exoIsBlakemodexp.position(exoIsBlakemodexp.position() + 1);
    }

    if (!filled.get(25)) {
      exoIsEcdata.position(exoIsEcdata.position() + 1);
    }

    if (!filled.get(26)) {
      exoIsKec.position(exoIsKec.position() + 1);
    }

    if (!filled.get(27)) {
      exoIsLog.position(exoIsLog.position() + 1);
    }

    if (!filled.get(28)) {
      exoIsRipsha.position(exoIsRipsha.position() + 1);
    }

    if (!filled.get(29)) {
      exoIsRom.position(exoIsRom.position() + 1);
    }

    if (!filled.get(30)) {
      exoIsTxcd.position(exoIsTxcd.position() + 1);
    }

    if (!filled.get(31)) {
      exoSum.position(exoSum.position() + 8);
    }

    if (!filled.get(32)) {
      fast.position(fast.position() + 1);
    }

    if (!filled.get(33)) {
      indexA.position(indexA.position() + 32);
    }

    if (!filled.get(34)) {
      indexB.position(indexB.position() + 32);
    }

    if (!filled.get(35)) {
      indexC.position(indexC.position() + 32);
    }

    if (!filled.get(36)) {
      indexX.position(indexX.position() + 32);
    }

    if (!filled.get(37)) {
      isLimbToRamOneTarget.position(isLimbToRamOneTarget.position() + 1);
    }

    if (!filled.get(38)) {
      isLimbToRamTransplant.position(isLimbToRamTransplant.position() + 1);
    }

    if (!filled.get(39)) {
      isLimbToRamTwoTarget.position(isLimbToRamTwoTarget.position() + 1);
    }

    if (!filled.get(40)) {
      isLimbVanishes.position(isLimbVanishes.position() + 1);
    }

    if (!filled.get(41)) {
      isRamExcision.position(isRamExcision.position() + 1);
    }

    if (!filled.get(42)) {
      isRamToLimbOneSource.position(isRamToLimbOneSource.position() + 1);
    }

    if (!filled.get(43)) {
      isRamToLimbTransplant.position(isRamToLimbTransplant.position() + 1);
    }

    if (!filled.get(44)) {
      isRamToLimbTwoSource.position(isRamToLimbTwoSource.position() + 1);
    }

    if (!filled.get(45)) {
      isRamToRamPartial.position(isRamToRamPartial.position() + 1);
    }

    if (!filled.get(46)) {
      isRamToRamTransplant.position(isRamToRamTransplant.position() + 1);
    }

    if (!filled.get(47)) {
      isRamToRamTwoSource.position(isRamToRamTwoSource.position() + 1);
    }

    if (!filled.get(48)) {
      isRamToRamTwoTarget.position(isRamToRamTwoTarget.position() + 1);
    }

    if (!filled.get(49)) {
      isRamVanishes.position(isRamVanishes.position() + 1);
    }

    if (!filled.get(50)) {
      kecId.position(kecId.position() + 8);
    }

    if (!filled.get(51)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(52)) {
      mmioInstruction.position(mmioInstruction.position() + 4);
    }

    if (!filled.get(53)) {
      mmioStamp.position(mmioStamp.position() + 8);
    }

    if (!filled.get(54)) {
      phase.position(phase.position() + 8);
    }

    if (!filled.get(55)) {
      pow2561.position(pow2561.position() + 32);
    }

    if (!filled.get(56)) {
      pow2562.position(pow2562.position() + 32);
    }

    if (!filled.get(57)) {
      size.position(size.position() + 32);
    }

    if (!filled.get(58)) {
      slow.position(slow.position() + 1);
    }

    if (!filled.get(59)) {
      sourceByteOffset.position(sourceByteOffset.position() + 2);
    }

    if (!filled.get(60)) {
      sourceLimbOffset.position(sourceLimbOffset.position() + 32);
    }

    if (!filled.get(61)) {
      successBit.position(successBit.position() + 1);
    }

    if (!filled.get(62)) {
      targetByteOffset.position(targetByteOffset.position() + 2);
    }

    if (!filled.get(63)) {
      targetLimbOffset.position(targetLimbOffset.position() + 32);
    }

    if (!filled.get(64)) {
      totalSize.position(totalSize.position() + 32);
    }

    if (!filled.get(65)) {
      valA.position(valA.position() + 32);
    }

    if (!filled.get(66)) {
      valANew.position(valANew.position() + 32);
    }

    if (!filled.get(67)) {
      valB.position(valB.position() + 32);
    }

    if (!filled.get(68)) {
      valBNew.position(valBNew.position() + 32);
    }

    if (!filled.get(69)) {
      valC.position(valC.position() + 32);
    }

    if (!filled.get(70)) {
      valCNew.position(valCNew.position() + 32);
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
