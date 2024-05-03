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

package net.consensys.linea.zktracer.module.wcp;

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
  public static final int EC_DATA_PHASE_ECADD_DATA = 0x3;
  public static final int EC_DATA_PHASE_ECADD_RESULT = 0x4;
  public static final int EC_DATA_PHASE_ECMUL_DATA = 0x5;
  public static final int EC_DATA_PHASE_ECMUL_RESULT = 0x6;
  public static final int EC_DATA_PHASE_ECRECOVER_DATA = 0x1;
  public static final int EC_DATA_PHASE_ECRECOVER_RESULT = 0x2;
  public static final int EC_DATA_PHASE_PAIRING_DATA = 0x7;
  public static final int EC_DATA_PHASE_PAIRING_RESULT = 0x8;
  public static final int EIP_3541_MARKER = 0xef;
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
  public static final int LINEA_BLOCK_GAS_LIMIT = 0x1c9c380;
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
  public static final int RLP_TXN_PHASE_ACCESS_LIST = 0xb;
  public static final int RLP_TXN_PHASE_BETA = 0xc;
  public static final int RLP_TXN_PHASE_CHAIN_ID = 0x2;
  public static final int RLP_TXN_PHASE_DATA = 0xa;
  public static final int RLP_TXN_PHASE_GAS_LIMIT = 0x7;
  public static final int RLP_TXN_PHASE_GAS_PRICE = 0x4;
  public static final int RLP_TXN_PHASE_MAX_FEE_PER_GAS = 0x6;
  public static final int RLP_TXN_PHASE_MAX_PRIORITY_FEE_PER_GAS = 0x5;
  public static final int RLP_TXN_PHASE_NONCE = 0x3;
  public static final int RLP_TXN_PHASE_R = 0xe;
  public static final int RLP_TXN_PHASE_RLP_PREFIX = 0x1;
  public static final int RLP_TXN_PHASE_S = 0xf;
  public static final int RLP_TXN_PHASE_TO = 0x8;
  public static final int RLP_TXN_PHASE_VALUE = 0x9;
  public static final int RLP_TXN_PHASE_Y = 0xd;
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
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
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
  private final MappedByteBuffer isEq;
  private final MappedByteBuffer isGeq;
  private final MappedByteBuffer isGt;
  private final MappedByteBuffer isIszero;
  private final MappedByteBuffer isLeq;
  private final MappedByteBuffer isLt;
  private final MappedByteBuffer isSgt;
  private final MappedByteBuffer isSlt;
  private final MappedByteBuffer neg1;
  private final MappedByteBuffer neg2;
  private final MappedByteBuffer oneLineInstruction;
  private final MappedByteBuffer result;
  private final MappedByteBuffer variableLengthInstruction;
  private final MappedByteBuffer wordComparisonStamp;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("wcp.ACC_1", 32, length),
        new ColumnHeader("wcp.ACC_2", 32, length),
        new ColumnHeader("wcp.ACC_3", 32, length),
        new ColumnHeader("wcp.ACC_4", 32, length),
        new ColumnHeader("wcp.ACC_5", 32, length),
        new ColumnHeader("wcp.ACC_6", 32, length),
        new ColumnHeader("wcp.ARGUMENT_1_HI", 32, length),
        new ColumnHeader("wcp.ARGUMENT_1_LO", 32, length),
        new ColumnHeader("wcp.ARGUMENT_2_HI", 32, length),
        new ColumnHeader("wcp.ARGUMENT_2_LO", 32, length),
        new ColumnHeader("wcp.BIT_1", 1, length),
        new ColumnHeader("wcp.BIT_2", 1, length),
        new ColumnHeader("wcp.BIT_3", 1, length),
        new ColumnHeader("wcp.BIT_4", 1, length),
        new ColumnHeader("wcp.BITS", 1, length),
        new ColumnHeader("wcp.BYTE_1", 1, length),
        new ColumnHeader("wcp.BYTE_2", 1, length),
        new ColumnHeader("wcp.BYTE_3", 1, length),
        new ColumnHeader("wcp.BYTE_4", 1, length),
        new ColumnHeader("wcp.BYTE_5", 1, length),
        new ColumnHeader("wcp.BYTE_6", 1, length),
        new ColumnHeader("wcp.COUNTER", 1, length),
        new ColumnHeader("wcp.CT_MAX", 1, length),
        new ColumnHeader("wcp.INST", 1, length),
        new ColumnHeader("wcp.IS_EQ", 1, length),
        new ColumnHeader("wcp.IS_GEQ", 1, length),
        new ColumnHeader("wcp.IS_GT", 1, length),
        new ColumnHeader("wcp.IS_ISZERO", 1, length),
        new ColumnHeader("wcp.IS_LEQ", 1, length),
        new ColumnHeader("wcp.IS_LT", 1, length),
        new ColumnHeader("wcp.IS_SGT", 1, length),
        new ColumnHeader("wcp.IS_SLT", 1, length),
        new ColumnHeader("wcp.NEG_1", 1, length),
        new ColumnHeader("wcp.NEG_2", 1, length),
        new ColumnHeader("wcp.ONE_LINE_INSTRUCTION", 1, length),
        new ColumnHeader("wcp.RESULT", 1, length),
        new ColumnHeader("wcp.VARIABLE_LENGTH_INSTRUCTION", 1, length),
        new ColumnHeader("wcp.WORD_COMPARISON_STAMP", 8, length));
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
    this.bit2 = buffers.get(11);
    this.bit3 = buffers.get(12);
    this.bit4 = buffers.get(13);
    this.bits = buffers.get(14);
    this.byte1 = buffers.get(15);
    this.byte2 = buffers.get(16);
    this.byte3 = buffers.get(17);
    this.byte4 = buffers.get(18);
    this.byte5 = buffers.get(19);
    this.byte6 = buffers.get(20);
    this.counter = buffers.get(21);
    this.ctMax = buffers.get(22);
    this.inst = buffers.get(23);
    this.isEq = buffers.get(24);
    this.isGeq = buffers.get(25);
    this.isGt = buffers.get(26);
    this.isIszero = buffers.get(27);
    this.isLeq = buffers.get(28);
    this.isLt = buffers.get(29);
    this.isSgt = buffers.get(30);
    this.isSlt = buffers.get(31);
    this.neg1 = buffers.get(32);
    this.neg2 = buffers.get(33);
    this.oneLineInstruction = buffers.get(34);
    this.result = buffers.get(35);
    this.variableLengthInstruction = buffers.get(36);
    this.wordComparisonStamp = buffers.get(37);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("wcp.ACC_1 already set");
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
      throw new IllegalStateException("wcp.ACC_2 already set");
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
      throw new IllegalStateException("wcp.ACC_3 already set");
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
      throw new IllegalStateException("wcp.ACC_4 already set");
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
      throw new IllegalStateException("wcp.ACC_5 already set");
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
      throw new IllegalStateException("wcp.ACC_6 already set");
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
      throw new IllegalStateException("wcp.ARGUMENT_1_HI already set");
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
      throw new IllegalStateException("wcp.ARGUMENT_1_LO already set");
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
      throw new IllegalStateException("wcp.ARGUMENT_2_HI already set");
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
      throw new IllegalStateException("wcp.ARGUMENT_2_LO already set");
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
      throw new IllegalStateException("wcp.BIT_1 already set");
    } else {
      filled.set(11);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("wcp.BIT_2 already set");
    } else {
      filled.set(12);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("wcp.BIT_3 already set");
    } else {
      filled.set(13);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("wcp.BIT_4 already set");
    } else {
      filled.set(14);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bits(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("wcp.BITS already set");
    } else {
      filled.set(10);
    }

    bits.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("wcp.BYTE_1 already set");
    } else {
      filled.set(15);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("wcp.BYTE_2 already set");
    } else {
      filled.set(16);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("wcp.BYTE_3 already set");
    } else {
      filled.set(17);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("wcp.BYTE_4 already set");
    } else {
      filled.set(18);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace byte5(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("wcp.BYTE_5 already set");
    } else {
      filled.set(19);
    }

    byte5.put(b.toByte());

    return this;
  }

  public Trace byte6(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("wcp.BYTE_6 already set");
    } else {
      filled.set(20);
    }

    byte6.put(b.toByte());

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("wcp.COUNTER already set");
    } else {
      filled.set(21);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("wcp.CT_MAX already set");
    } else {
      filled.set(22);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("wcp.INST already set");
    } else {
      filled.set(23);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isEq(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("wcp.IS_EQ already set");
    } else {
      filled.set(24);
    }

    isEq.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isGeq(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("wcp.IS_GEQ already set");
    } else {
      filled.set(25);
    }

    isGeq.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isGt(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("wcp.IS_GT already set");
    } else {
      filled.set(26);
    }

    isGt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isIszero(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("wcp.IS_ISZERO already set");
    } else {
      filled.set(27);
    }

    isIszero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLeq(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("wcp.IS_LEQ already set");
    } else {
      filled.set(28);
    }

    isLeq.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLt(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("wcp.IS_LT already set");
    } else {
      filled.set(29);
    }

    isLt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSgt(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("wcp.IS_SGT already set");
    } else {
      filled.set(30);
    }

    isSgt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSlt(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("wcp.IS_SLT already set");
    } else {
      filled.set(31);
    }

    isSlt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace neg1(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("wcp.NEG_1 already set");
    } else {
      filled.set(32);
    }

    neg1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace neg2(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("wcp.NEG_2 already set");
    } else {
      filled.set(33);
    }

    neg2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oneLineInstruction(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("wcp.ONE_LINE_INSTRUCTION already set");
    } else {
      filled.set(34);
    }

    oneLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace result(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("wcp.RESULT already set");
    } else {
      filled.set(35);
    }

    result.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace variableLengthInstruction(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("wcp.VARIABLE_LENGTH_INSTRUCTION already set");
    } else {
      filled.set(36);
    }

    variableLengthInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace wordComparisonStamp(final long b) {
    if (filled.get(37)) {
      throw new IllegalStateException("wcp.WORD_COMPARISON_STAMP already set");
    } else {
      filled.set(37);
    }

    wordComparisonStamp.putLong(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("wcp.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("wcp.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("wcp.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("wcp.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("wcp.ACC_5 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("wcp.ACC_6 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("wcp.ARGUMENT_1_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("wcp.ARGUMENT_1_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("wcp.ARGUMENT_2_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("wcp.ARGUMENT_2_LO has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("wcp.BIT_1 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("wcp.BIT_2 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("wcp.BIT_3 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("wcp.BIT_4 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("wcp.BITS has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("wcp.BYTE_1 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("wcp.BYTE_2 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("wcp.BYTE_3 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("wcp.BYTE_4 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("wcp.BYTE_5 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("wcp.BYTE_6 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("wcp.COUNTER has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("wcp.CT_MAX has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("wcp.INST has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("wcp.IS_EQ has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("wcp.IS_GEQ has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("wcp.IS_GT has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("wcp.IS_ISZERO has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("wcp.IS_LEQ has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("wcp.IS_LT has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("wcp.IS_SGT has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("wcp.IS_SLT has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("wcp.NEG_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("wcp.NEG_2 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("wcp.ONE_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("wcp.RESULT has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("wcp.VARIABLE_LENGTH_INSTRUCTION has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("wcp.WORD_COMPARISON_STAMP has not been filled");
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
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(13)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(14)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(10)) {
      bits.position(bits.position() + 1);
    }

    if (!filled.get(15)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(16)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(17)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(18)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(19)) {
      byte5.position(byte5.position() + 1);
    }

    if (!filled.get(20)) {
      byte6.position(byte6.position() + 1);
    }

    if (!filled.get(21)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(22)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(23)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(24)) {
      isEq.position(isEq.position() + 1);
    }

    if (!filled.get(25)) {
      isGeq.position(isGeq.position() + 1);
    }

    if (!filled.get(26)) {
      isGt.position(isGt.position() + 1);
    }

    if (!filled.get(27)) {
      isIszero.position(isIszero.position() + 1);
    }

    if (!filled.get(28)) {
      isLeq.position(isLeq.position() + 1);
    }

    if (!filled.get(29)) {
      isLt.position(isLt.position() + 1);
    }

    if (!filled.get(30)) {
      isSgt.position(isSgt.position() + 1);
    }

    if (!filled.get(31)) {
      isSlt.position(isSlt.position() + 1);
    }

    if (!filled.get(32)) {
      neg1.position(neg1.position() + 1);
    }

    if (!filled.get(33)) {
      neg2.position(neg2.position() + 1);
    }

    if (!filled.get(34)) {
      oneLineInstruction.position(oneLineInstruction.position() + 1);
    }

    if (!filled.get(35)) {
      result.position(result.position() + 1);
    }

    if (!filled.get(36)) {
      variableLengthInstruction.position(variableLengthInstruction.position() + 1);
    }

    if (!filled.get(37)) {
      wordComparisonStamp.position(wordComparisonStamp.position() + 8);
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
