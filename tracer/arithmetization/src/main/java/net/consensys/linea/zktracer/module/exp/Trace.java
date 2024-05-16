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

package net.consensys.linea.zktracer.module.exp;

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
  public static final int EQ = 0x14;
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
  public static final int G_EXPBYTES = 0x32;
  public static final int ISZERO = 0x15;
  public static final int LINEA_BLOCK_GAS_LIMIT = 0x1c9c380;
  public static final int LLARGE = 0x10;
  public static final int LLARGEMO = 0xf;
  public static final int LLARGEPO = 0x11;
  public static final int LT = 0x10;
  public static final int MAX_CT_CMPTN_EXP_LOG = 0xf;
  public static final int MAX_CT_CMPTN_MODEXP_LOG = 0xf;
  public static final int MAX_CT_MACRO_EXP_LOG = 0x0;
  public static final int MAX_CT_MACRO_MODEXP_LOG = 0x0;
  public static final int MAX_CT_PRPRC_EXP_LOG = 0x0;
  public static final int MAX_CT_PRPRC_MODEXP_LOG = 0x4;
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

  private final MappedByteBuffer cmptn;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer data3XorWcpArg2Hi;
  private final MappedByteBuffer data4XorWcpArg2Lo;
  private final MappedByteBuffer data5;
  private final MappedByteBuffer expInst;
  private final MappedByteBuffer isExpLog;
  private final MappedByteBuffer isModexpLog;
  private final MappedByteBuffer macro;
  private final MappedByteBuffer manzbAcc;
  private final MappedByteBuffer manzbXorWcpFlag;
  private final MappedByteBuffer msbAcc;
  private final MappedByteBuffer msbBitXorWcpRes;
  private final MappedByteBuffer msbXorWcpInst;
  private final MappedByteBuffer pltBit;
  private final MappedByteBuffer pltJmp;
  private final MappedByteBuffer prprc;
  private final MappedByteBuffer rawAccXorData1XorWcpArg1Hi;
  private final MappedByteBuffer rawByte;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer tanzb;
  private final MappedByteBuffer tanzbAcc;
  private final MappedByteBuffer trimAccXorData2XorWcpArg1Lo;
  private final MappedByteBuffer trimByte;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("exp.CMPTN", 1, length),
        new ColumnHeader("exp.CT", 2, length),
        new ColumnHeader("exp.CT_MAX", 2, length),
        new ColumnHeader("exp.DATA_3_xor_WCP_ARG_2_HI", 32, length),
        new ColumnHeader("exp.DATA_4_xor_WCP_ARG_2_LO", 32, length),
        new ColumnHeader("exp.DATA_5", 32, length),
        new ColumnHeader("exp.EXP_INST", 4, length),
        new ColumnHeader("exp.IS_EXP_LOG", 1, length),
        new ColumnHeader("exp.IS_MODEXP_LOG", 1, length),
        new ColumnHeader("exp.MACRO", 1, length),
        new ColumnHeader("exp.MANZB_ACC", 2, length),
        new ColumnHeader("exp.MANZB_xor_WCP_FLAG", 1, length),
        new ColumnHeader("exp.MSB_ACC", 1, length),
        new ColumnHeader("exp.MSB_BIT_xor_WCP_RES", 1, length),
        new ColumnHeader("exp.MSB_xor_WCP_INST", 1, length),
        new ColumnHeader("exp.PLT_BIT", 1, length),
        new ColumnHeader("exp.PLT_JMP", 2, length),
        new ColumnHeader("exp.PRPRC", 1, length),
        new ColumnHeader("exp.RAW_ACC_xor_DATA_1_xor_WCP_ARG_1_HI", 32, length),
        new ColumnHeader("exp.RAW_BYTE", 1, length),
        new ColumnHeader("exp.STAMP", 8, length),
        new ColumnHeader("exp.TANZB", 1, length),
        new ColumnHeader("exp.TANZB_ACC", 2, length),
        new ColumnHeader("exp.TRIM_ACC_xor_DATA_2_xor_WCP_ARG_1_LO", 32, length),
        new ColumnHeader("exp.TRIM_BYTE", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.cmptn = buffers.get(0);
    this.ct = buffers.get(1);
    this.ctMax = buffers.get(2);
    this.data3XorWcpArg2Hi = buffers.get(3);
    this.data4XorWcpArg2Lo = buffers.get(4);
    this.data5 = buffers.get(5);
    this.expInst = buffers.get(6);
    this.isExpLog = buffers.get(7);
    this.isModexpLog = buffers.get(8);
    this.macro = buffers.get(9);
    this.manzbAcc = buffers.get(10);
    this.manzbXorWcpFlag = buffers.get(11);
    this.msbAcc = buffers.get(12);
    this.msbBitXorWcpRes = buffers.get(13);
    this.msbXorWcpInst = buffers.get(14);
    this.pltBit = buffers.get(15);
    this.pltJmp = buffers.get(16);
    this.prprc = buffers.get(17);
    this.rawAccXorData1XorWcpArg1Hi = buffers.get(18);
    this.rawByte = buffers.get(19);
    this.stamp = buffers.get(20);
    this.tanzb = buffers.get(21);
    this.tanzbAcc = buffers.get(22);
    this.trimAccXorData2XorWcpArg1Lo = buffers.get(23);
    this.trimByte = buffers.get(24);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace cmptn(final Boolean b) {
    if (filled.get(0)) {
      throw new IllegalStateException("exp.CMPTN already set");
    } else {
      filled.set(0);
    }

    cmptn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final short b) {
    if (filled.get(1)) {
      throw new IllegalStateException("exp.CT already set");
    } else {
      filled.set(1);
    }

    ct.putShort(b);

    return this;
  }

  public Trace ctMax(final short b) {
    if (filled.get(2)) {
      throw new IllegalStateException("exp.CT_MAX already set");
    } else {
      filled.set(2);
    }

    ctMax.putShort(b);

    return this;
  }

  public Trace isExpLog(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("exp.IS_EXP_LOG already set");
    } else {
      filled.set(3);
    }

    isExpLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpLog(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("exp.IS_MODEXP_LOG already set");
    } else {
      filled.set(4);
    }

    isModexpLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace macro(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("exp.MACRO already set");
    } else {
      filled.set(5);
    }

    macro.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationManzb(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("exp.computation/MANZB already set");
    } else {
      filled.set(8);
    }

    manzbXorWcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationManzbAcc(final short b) {
    if (filled.get(16)) {
      throw new IllegalStateException("exp.computation/MANZB_ACC already set");
    } else {
      filled.set(16);
    }

    manzbAcc.putShort(b);

    return this;
  }

  public Trace pComputationMsb(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("exp.computation/MSB already set");
    } else {
      filled.set(12);
    }

    msbXorWcpInst.put(b.toByte());

    return this;
  }

  public Trace pComputationMsbAcc(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("exp.computation/MSB_ACC already set");
    } else {
      filled.set(13);
    }

    msbAcc.put(b.toByte());

    return this;
  }

  public Trace pComputationMsbBit(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("exp.computation/MSB_BIT already set");
    } else {
      filled.set(9);
    }

    msbBitXorWcpRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationPltBit(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("exp.computation/PLT_BIT already set");
    } else {
      filled.set(10);
    }

    pltBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationPltJmp(final short b) {
    if (filled.get(18)) {
      throw new IllegalStateException("exp.computation/PLT_JMP already set");
    } else {
      filled.set(18);
    }

    pltJmp.putShort(b);

    return this;
  }

  public Trace pComputationRawAcc(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("exp.computation/RAW_ACC already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rawAccXorData1XorWcpArg1Hi.put((byte) 0);
    }
    rawAccXorData1XorWcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pComputationRawByte(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("exp.computation/RAW_BYTE already set");
    } else {
      filled.set(14);
    }

    rawByte.put(b.toByte());

    return this;
  }

  public Trace pComputationTanzb(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("exp.computation/TANZB already set");
    } else {
      filled.set(11);
    }

    tanzb.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pComputationTanzbAcc(final short b) {
    if (filled.get(17)) {
      throw new IllegalStateException("exp.computation/TANZB_ACC already set");
    } else {
      filled.set(17);
    }

    tanzbAcc.putShort(b);

    return this;
  }

  public Trace pComputationTrimAcc(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("exp.computation/TRIM_ACC already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trimAccXorData2XorWcpArg1Lo.put((byte) 0);
    }
    trimAccXorData2XorWcpArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pComputationTrimByte(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("exp.computation/TRIM_BYTE already set");
    } else {
      filled.set(15);
    }

    trimByte.put(b.toByte());

    return this;
  }

  public Trace pMacroData1(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("exp.macro/DATA_1 already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rawAccXorData1XorWcpArg1Hi.put((byte) 0);
    }
    rawAccXorData1XorWcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroData2(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("exp.macro/DATA_2 already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trimAccXorData2XorWcpArg1Lo.put((byte) 0);
    }
    trimAccXorData2XorWcpArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroData3(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("exp.macro/DATA_3 already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data3XorWcpArg2Hi.put((byte) 0);
    }
    data3XorWcpArg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroData4(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("exp.macro/DATA_4 already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data4XorWcpArg2Lo.put((byte) 0);
    }
    data4XorWcpArg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroData5(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("exp.macro/DATA_5 already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data5.put((byte) 0);
    }
    data5.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMacroExpInst(final int b) {
    if (filled.get(19)) {
      throw new IllegalStateException("exp.macro/EXP_INST already set");
    } else {
      filled.set(19);
    }

    expInst.putInt(b);

    return this;
  }

  public Trace pPreprocessingWcpArg1Hi(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_1_HI already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rawAccXorData1XorWcpArg1Hi.put((byte) 0);
    }
    rawAccXorData1XorWcpArg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPreprocessingWcpArg1Lo(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_1_LO already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trimAccXorData2XorWcpArg1Lo.put((byte) 0);
    }
    trimAccXorData2XorWcpArg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPreprocessingWcpArg2Hi(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_2_HI already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data3XorWcpArg2Hi.put((byte) 0);
    }
    data3XorWcpArg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPreprocessingWcpArg2Lo(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("exp.preprocessing/WCP_ARG_2_LO already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      data4XorWcpArg2Lo.put((byte) 0);
    }
    data4XorWcpArg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pPreprocessingWcpFlag(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("exp.preprocessing/WCP_FLAG already set");
    } else {
      filled.set(8);
    }

    manzbXorWcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pPreprocessingWcpInst(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("exp.preprocessing/WCP_INST already set");
    } else {
      filled.set(12);
    }

    msbXorWcpInst.put(b.toByte());

    return this;
  }

  public Trace pPreprocessingWcpRes(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("exp.preprocessing/WCP_RES already set");
    } else {
      filled.set(9);
    }

    msbBitXorWcpRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace prprc(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("exp.PRPRC already set");
    } else {
      filled.set(6);
    }

    prprc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("exp.STAMP already set");
    } else {
      filled.set(7);
    }

    stamp.putLong(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("exp.CMPTN has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("exp.CT has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("exp.CT_MAX has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("exp.DATA_3_xor_WCP_ARG_2_HI has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("exp.DATA_4_xor_WCP_ARG_2_LO has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("exp.DATA_5 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("exp.EXP_INST has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("exp.IS_EXP_LOG has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("exp.IS_MODEXP_LOG has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("exp.MACRO has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("exp.MANZB_ACC has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("exp.MANZB_xor_WCP_FLAG has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("exp.MSB_ACC has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("exp.MSB_BIT_xor_WCP_RES has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("exp.MSB_xor_WCP_INST has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("exp.PLT_BIT has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("exp.PLT_JMP has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("exp.PRPRC has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException(
          "exp.RAW_ACC_xor_DATA_1_xor_WCP_ARG_1_HI has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("exp.RAW_BYTE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("exp.STAMP has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("exp.TANZB has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("exp.TANZB_ACC has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException(
          "exp.TRIM_ACC_xor_DATA_2_xor_WCP_ARG_1_LO has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("exp.TRIM_BYTE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      cmptn.position(cmptn.position() + 1);
    }

    if (!filled.get(1)) {
      ct.position(ct.position() + 2);
    }

    if (!filled.get(2)) {
      ctMax.position(ctMax.position() + 2);
    }

    if (!filled.get(22)) {
      data3XorWcpArg2Hi.position(data3XorWcpArg2Hi.position() + 32);
    }

    if (!filled.get(23)) {
      data4XorWcpArg2Lo.position(data4XorWcpArg2Lo.position() + 32);
    }

    if (!filled.get(24)) {
      data5.position(data5.position() + 32);
    }

    if (!filled.get(19)) {
      expInst.position(expInst.position() + 4);
    }

    if (!filled.get(3)) {
      isExpLog.position(isExpLog.position() + 1);
    }

    if (!filled.get(4)) {
      isModexpLog.position(isModexpLog.position() + 1);
    }

    if (!filled.get(5)) {
      macro.position(macro.position() + 1);
    }

    if (!filled.get(16)) {
      manzbAcc.position(manzbAcc.position() + 2);
    }

    if (!filled.get(8)) {
      manzbXorWcpFlag.position(manzbXorWcpFlag.position() + 1);
    }

    if (!filled.get(13)) {
      msbAcc.position(msbAcc.position() + 1);
    }

    if (!filled.get(9)) {
      msbBitXorWcpRes.position(msbBitXorWcpRes.position() + 1);
    }

    if (!filled.get(12)) {
      msbXorWcpInst.position(msbXorWcpInst.position() + 1);
    }

    if (!filled.get(10)) {
      pltBit.position(pltBit.position() + 1);
    }

    if (!filled.get(18)) {
      pltJmp.position(pltJmp.position() + 2);
    }

    if (!filled.get(6)) {
      prprc.position(prprc.position() + 1);
    }

    if (!filled.get(20)) {
      rawAccXorData1XorWcpArg1Hi.position(rawAccXorData1XorWcpArg1Hi.position() + 32);
    }

    if (!filled.get(14)) {
      rawByte.position(rawByte.position() + 1);
    }

    if (!filled.get(7)) {
      stamp.position(stamp.position() + 8);
    }

    if (!filled.get(11)) {
      tanzb.position(tanzb.position() + 1);
    }

    if (!filled.get(17)) {
      tanzbAcc.position(tanzbAcc.position() + 2);
    }

    if (!filled.get(21)) {
      trimAccXorData2XorWcpArg1Lo.position(trimAccXorData2XorWcpArg1Lo.position() + 32);
    }

    if (!filled.get(15)) {
      trimByte.position(trimByte.position() + 1);
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
