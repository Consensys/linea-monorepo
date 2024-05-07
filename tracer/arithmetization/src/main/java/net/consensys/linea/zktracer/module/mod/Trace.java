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

package net.consensys.linea.zktracer.module.mod;

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

  private final MappedByteBuffer acc12;
  private final MappedByteBuffer acc13;
  private final MappedByteBuffer acc22;
  private final MappedByteBuffer acc23;
  private final MappedByteBuffer accB0;
  private final MappedByteBuffer accB1;
  private final MappedByteBuffer accB2;
  private final MappedByteBuffer accB3;
  private final MappedByteBuffer accDelta0;
  private final MappedByteBuffer accDelta1;
  private final MappedByteBuffer accDelta2;
  private final MappedByteBuffer accDelta3;
  private final MappedByteBuffer accH0;
  private final MappedByteBuffer accH1;
  private final MappedByteBuffer accH2;
  private final MappedByteBuffer accQ0;
  private final MappedByteBuffer accQ1;
  private final MappedByteBuffer accQ2;
  private final MappedByteBuffer accQ3;
  private final MappedByteBuffer accR0;
  private final MappedByteBuffer accR1;
  private final MappedByteBuffer accR2;
  private final MappedByteBuffer accR3;
  private final MappedByteBuffer arg1Hi;
  private final MappedByteBuffer arg1Lo;
  private final MappedByteBuffer arg2Hi;
  private final MappedByteBuffer arg2Lo;
  private final MappedByteBuffer byte12;
  private final MappedByteBuffer byte13;
  private final MappedByteBuffer byte22;
  private final MappedByteBuffer byte23;
  private final MappedByteBuffer byteB0;
  private final MappedByteBuffer byteB1;
  private final MappedByteBuffer byteB2;
  private final MappedByteBuffer byteB3;
  private final MappedByteBuffer byteDelta0;
  private final MappedByteBuffer byteDelta1;
  private final MappedByteBuffer byteDelta2;
  private final MappedByteBuffer byteDelta3;
  private final MappedByteBuffer byteH0;
  private final MappedByteBuffer byteH1;
  private final MappedByteBuffer byteH2;
  private final MappedByteBuffer byteQ0;
  private final MappedByteBuffer byteQ1;
  private final MappedByteBuffer byteQ2;
  private final MappedByteBuffer byteQ3;
  private final MappedByteBuffer byteR0;
  private final MappedByteBuffer byteR1;
  private final MappedByteBuffer byteR2;
  private final MappedByteBuffer byteR3;
  private final MappedByteBuffer cmp1;
  private final MappedByteBuffer cmp2;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isDiv;
  private final MappedByteBuffer isMod;
  private final MappedByteBuffer isSdiv;
  private final MappedByteBuffer isSmod;
  private final MappedByteBuffer mli;
  private final MappedByteBuffer msb1;
  private final MappedByteBuffer msb2;
  private final MappedByteBuffer oli;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer signed;
  private final MappedByteBuffer stamp;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("mod.ACC_1_2", 32, length),
        new ColumnHeader("mod.ACC_1_3", 32, length),
        new ColumnHeader("mod.ACC_2_2", 32, length),
        new ColumnHeader("mod.ACC_2_3", 32, length),
        new ColumnHeader("mod.ACC_B_0", 32, length),
        new ColumnHeader("mod.ACC_B_1", 32, length),
        new ColumnHeader("mod.ACC_B_2", 32, length),
        new ColumnHeader("mod.ACC_B_3", 32, length),
        new ColumnHeader("mod.ACC_DELTA_0", 32, length),
        new ColumnHeader("mod.ACC_DELTA_1", 32, length),
        new ColumnHeader("mod.ACC_DELTA_2", 32, length),
        new ColumnHeader("mod.ACC_DELTA_3", 32, length),
        new ColumnHeader("mod.ACC_H_0", 32, length),
        new ColumnHeader("mod.ACC_H_1", 32, length),
        new ColumnHeader("mod.ACC_H_2", 32, length),
        new ColumnHeader("mod.ACC_Q_0", 32, length),
        new ColumnHeader("mod.ACC_Q_1", 32, length),
        new ColumnHeader("mod.ACC_Q_2", 32, length),
        new ColumnHeader("mod.ACC_Q_3", 32, length),
        new ColumnHeader("mod.ACC_R_0", 32, length),
        new ColumnHeader("mod.ACC_R_1", 32, length),
        new ColumnHeader("mod.ACC_R_2", 32, length),
        new ColumnHeader("mod.ACC_R_3", 32, length),
        new ColumnHeader("mod.ARG_1_HI", 32, length),
        new ColumnHeader("mod.ARG_1_LO", 32, length),
        new ColumnHeader("mod.ARG_2_HI", 32, length),
        new ColumnHeader("mod.ARG_2_LO", 32, length),
        new ColumnHeader("mod.BYTE_1_2", 1, length),
        new ColumnHeader("mod.BYTE_1_3", 1, length),
        new ColumnHeader("mod.BYTE_2_2", 1, length),
        new ColumnHeader("mod.BYTE_2_3", 1, length),
        new ColumnHeader("mod.BYTE_B_0", 1, length),
        new ColumnHeader("mod.BYTE_B_1", 1, length),
        new ColumnHeader("mod.BYTE_B_2", 1, length),
        new ColumnHeader("mod.BYTE_B_3", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_0", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_1", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_2", 1, length),
        new ColumnHeader("mod.BYTE_DELTA_3", 1, length),
        new ColumnHeader("mod.BYTE_H_0", 1, length),
        new ColumnHeader("mod.BYTE_H_1", 1, length),
        new ColumnHeader("mod.BYTE_H_2", 1, length),
        new ColumnHeader("mod.BYTE_Q_0", 1, length),
        new ColumnHeader("mod.BYTE_Q_1", 1, length),
        new ColumnHeader("mod.BYTE_Q_2", 1, length),
        new ColumnHeader("mod.BYTE_Q_3", 1, length),
        new ColumnHeader("mod.BYTE_R_0", 1, length),
        new ColumnHeader("mod.BYTE_R_1", 1, length),
        new ColumnHeader("mod.BYTE_R_2", 1, length),
        new ColumnHeader("mod.BYTE_R_3", 1, length),
        new ColumnHeader("mod.CMP_1", 1, length),
        new ColumnHeader("mod.CMP_2", 1, length),
        new ColumnHeader("mod.CT", 2, length),
        new ColumnHeader("mod.INST", 1, length),
        new ColumnHeader("mod.IS_DIV", 1, length),
        new ColumnHeader("mod.IS_MOD", 1, length),
        new ColumnHeader("mod.IS_SDIV", 1, length),
        new ColumnHeader("mod.IS_SMOD", 1, length),
        new ColumnHeader("mod.MLI", 1, length),
        new ColumnHeader("mod.MSB_1", 1, length),
        new ColumnHeader("mod.MSB_2", 1, length),
        new ColumnHeader("mod.OLI", 1, length),
        new ColumnHeader("mod.RES_HI", 32, length),
        new ColumnHeader("mod.RES_LO", 32, length),
        new ColumnHeader("mod.SIGNED", 1, length),
        new ColumnHeader("mod.STAMP", 8, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc12 = buffers.get(0);
    this.acc13 = buffers.get(1);
    this.acc22 = buffers.get(2);
    this.acc23 = buffers.get(3);
    this.accB0 = buffers.get(4);
    this.accB1 = buffers.get(5);
    this.accB2 = buffers.get(6);
    this.accB3 = buffers.get(7);
    this.accDelta0 = buffers.get(8);
    this.accDelta1 = buffers.get(9);
    this.accDelta2 = buffers.get(10);
    this.accDelta3 = buffers.get(11);
    this.accH0 = buffers.get(12);
    this.accH1 = buffers.get(13);
    this.accH2 = buffers.get(14);
    this.accQ0 = buffers.get(15);
    this.accQ1 = buffers.get(16);
    this.accQ2 = buffers.get(17);
    this.accQ3 = buffers.get(18);
    this.accR0 = buffers.get(19);
    this.accR1 = buffers.get(20);
    this.accR2 = buffers.get(21);
    this.accR3 = buffers.get(22);
    this.arg1Hi = buffers.get(23);
    this.arg1Lo = buffers.get(24);
    this.arg2Hi = buffers.get(25);
    this.arg2Lo = buffers.get(26);
    this.byte12 = buffers.get(27);
    this.byte13 = buffers.get(28);
    this.byte22 = buffers.get(29);
    this.byte23 = buffers.get(30);
    this.byteB0 = buffers.get(31);
    this.byteB1 = buffers.get(32);
    this.byteB2 = buffers.get(33);
    this.byteB3 = buffers.get(34);
    this.byteDelta0 = buffers.get(35);
    this.byteDelta1 = buffers.get(36);
    this.byteDelta2 = buffers.get(37);
    this.byteDelta3 = buffers.get(38);
    this.byteH0 = buffers.get(39);
    this.byteH1 = buffers.get(40);
    this.byteH2 = buffers.get(41);
    this.byteQ0 = buffers.get(42);
    this.byteQ1 = buffers.get(43);
    this.byteQ2 = buffers.get(44);
    this.byteQ3 = buffers.get(45);
    this.byteR0 = buffers.get(46);
    this.byteR1 = buffers.get(47);
    this.byteR2 = buffers.get(48);
    this.byteR3 = buffers.get(49);
    this.cmp1 = buffers.get(50);
    this.cmp2 = buffers.get(51);
    this.ct = buffers.get(52);
    this.inst = buffers.get(53);
    this.isDiv = buffers.get(54);
    this.isMod = buffers.get(55);
    this.isSdiv = buffers.get(56);
    this.isSmod = buffers.get(57);
    this.mli = buffers.get(58);
    this.msb1 = buffers.get(59);
    this.msb2 = buffers.get(60);
    this.oli = buffers.get(61);
    this.resHi = buffers.get(62);
    this.resLo = buffers.get(63);
    this.signed = buffers.get(64);
    this.stamp = buffers.get(65);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc12(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mod.ACC_1_2 already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc12.put((byte) 0);
    }
    acc12.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc13(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mod.ACC_1_3 already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc13.put((byte) 0);
    }
    acc13.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc22(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mod.ACC_2_2 already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc22.put((byte) 0);
    }
    acc22.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc23(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mod.ACC_2_3 already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc23.put((byte) 0);
    }
    acc23.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB0(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mod.ACC_B_0 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB0.put((byte) 0);
    }
    accB0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB1(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mod.ACC_B_1 already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB1.put((byte) 0);
    }
    accB1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB2(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mod.ACC_B_2 already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB2.put((byte) 0);
    }
    accB2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accB3(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mod.ACC_B_3 already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accB3.put((byte) 0);
    }
    accB3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta0(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mod.ACC_DELTA_0 already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta0.put((byte) 0);
    }
    accDelta0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta1(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mod.ACC_DELTA_1 already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta1.put((byte) 0);
    }
    accDelta1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta2(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mod.ACC_DELTA_2 already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta2.put((byte) 0);
    }
    accDelta2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accDelta3(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mod.ACC_DELTA_3 already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accDelta3.put((byte) 0);
    }
    accDelta3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accH0(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mod.ACC_H_0 already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accH0.put((byte) 0);
    }
    accH0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accH1(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mod.ACC_H_1 already set");
    } else {
      filled.set(13);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accH1.put((byte) 0);
    }
    accH1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accH2(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mod.ACC_H_2 already set");
    } else {
      filled.set(14);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accH2.put((byte) 0);
    }
    accH2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ0(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mod.ACC_Q_0 already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ0.put((byte) 0);
    }
    accQ0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ1(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mod.ACC_Q_1 already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ1.put((byte) 0);
    }
    accQ1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ2(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mod.ACC_Q_2 already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ2.put((byte) 0);
    }
    accQ2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accQ3(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mod.ACC_Q_3 already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accQ3.put((byte) 0);
    }
    accQ3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR0(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mod.ACC_R_0 already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR0.put((byte) 0);
    }
    accR0.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR1(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mod.ACC_R_1 already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR1.put((byte) 0);
    }
    accR1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR2(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mod.ACC_R_2 already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR2.put((byte) 0);
    }
    accR2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accR3(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mod.ACC_R_3 already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accR3.put((byte) 0);
    }
    accR3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Hi(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mod.ARG_1_HI already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Hi.put((byte) 0);
    }
    arg1Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg1Lo(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mod.ARG_1_LO already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg1Lo.put((byte) 0);
    }
    arg1Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Hi(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mod.ARG_2_HI already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Hi.put((byte) 0);
    }
    arg2Hi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace arg2Lo(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mod.ARG_2_LO already set");
    } else {
      filled.set(26);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      arg2Lo.put((byte) 0);
    }
    arg2Lo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byte12(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mod.BYTE_1_2 already set");
    } else {
      filled.set(27);
    }

    byte12.put(b.toByte());

    return this;
  }

  public Trace byte13(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mod.BYTE_1_3 already set");
    } else {
      filled.set(28);
    }

    byte13.put(b.toByte());

    return this;
  }

  public Trace byte22(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mod.BYTE_2_2 already set");
    } else {
      filled.set(29);
    }

    byte22.put(b.toByte());

    return this;
  }

  public Trace byte23(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mod.BYTE_2_3 already set");
    } else {
      filled.set(30);
    }

    byte23.put(b.toByte());

    return this;
  }

  public Trace byteB0(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mod.BYTE_B_0 already set");
    } else {
      filled.set(31);
    }

    byteB0.put(b.toByte());

    return this;
  }

  public Trace byteB1(final UnsignedByte b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mod.BYTE_B_1 already set");
    } else {
      filled.set(32);
    }

    byteB1.put(b.toByte());

    return this;
  }

  public Trace byteB2(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mod.BYTE_B_2 already set");
    } else {
      filled.set(33);
    }

    byteB2.put(b.toByte());

    return this;
  }

  public Trace byteB3(final UnsignedByte b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mod.BYTE_B_3 already set");
    } else {
      filled.set(34);
    }

    byteB3.put(b.toByte());

    return this;
  }

  public Trace byteDelta0(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mod.BYTE_DELTA_0 already set");
    } else {
      filled.set(35);
    }

    byteDelta0.put(b.toByte());

    return this;
  }

  public Trace byteDelta1(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mod.BYTE_DELTA_1 already set");
    } else {
      filled.set(36);
    }

    byteDelta1.put(b.toByte());

    return this;
  }

  public Trace byteDelta2(final UnsignedByte b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mod.BYTE_DELTA_2 already set");
    } else {
      filled.set(37);
    }

    byteDelta2.put(b.toByte());

    return this;
  }

  public Trace byteDelta3(final UnsignedByte b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mod.BYTE_DELTA_3 already set");
    } else {
      filled.set(38);
    }

    byteDelta3.put(b.toByte());

    return this;
  }

  public Trace byteH0(final UnsignedByte b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mod.BYTE_H_0 already set");
    } else {
      filled.set(39);
    }

    byteH0.put(b.toByte());

    return this;
  }

  public Trace byteH1(final UnsignedByte b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mod.BYTE_H_1 already set");
    } else {
      filled.set(40);
    }

    byteH1.put(b.toByte());

    return this;
  }

  public Trace byteH2(final UnsignedByte b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mod.BYTE_H_2 already set");
    } else {
      filled.set(41);
    }

    byteH2.put(b.toByte());

    return this;
  }

  public Trace byteQ0(final UnsignedByte b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mod.BYTE_Q_0 already set");
    } else {
      filled.set(42);
    }

    byteQ0.put(b.toByte());

    return this;
  }

  public Trace byteQ1(final UnsignedByte b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mod.BYTE_Q_1 already set");
    } else {
      filled.set(43);
    }

    byteQ1.put(b.toByte());

    return this;
  }

  public Trace byteQ2(final UnsignedByte b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mod.BYTE_Q_2 already set");
    } else {
      filled.set(44);
    }

    byteQ2.put(b.toByte());

    return this;
  }

  public Trace byteQ3(final UnsignedByte b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mod.BYTE_Q_3 already set");
    } else {
      filled.set(45);
    }

    byteQ3.put(b.toByte());

    return this;
  }

  public Trace byteR0(final UnsignedByte b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mod.BYTE_R_0 already set");
    } else {
      filled.set(46);
    }

    byteR0.put(b.toByte());

    return this;
  }

  public Trace byteR1(final UnsignedByte b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mod.BYTE_R_1 already set");
    } else {
      filled.set(47);
    }

    byteR1.put(b.toByte());

    return this;
  }

  public Trace byteR2(final UnsignedByte b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mod.BYTE_R_2 already set");
    } else {
      filled.set(48);
    }

    byteR2.put(b.toByte());

    return this;
  }

  public Trace byteR3(final UnsignedByte b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mod.BYTE_R_3 already set");
    } else {
      filled.set(49);
    }

    byteR3.put(b.toByte());

    return this;
  }

  public Trace cmp1(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mod.CMP_1 already set");
    } else {
      filled.set(50);
    }

    cmp1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace cmp2(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mod.CMP_2 already set");
    } else {
      filled.set(51);
    }

    cmp2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final short b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mod.CT already set");
    } else {
      filled.set(52);
    }

    ct.putShort(b);

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mod.INST already set");
    } else {
      filled.set(53);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isDiv(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mod.IS_DIV already set");
    } else {
      filled.set(54);
    }

    isDiv.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isMod(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mod.IS_MOD already set");
    } else {
      filled.set(55);
    }

    isMod.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSdiv(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mod.IS_SDIV already set");
    } else {
      filled.set(56);
    }

    isSdiv.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSmod(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mod.IS_SMOD already set");
    } else {
      filled.set(57);
    }

    isSmod.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mli(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mod.MLI already set");
    } else {
      filled.set(58);
    }

    mli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace msb1(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mod.MSB_1 already set");
    } else {
      filled.set(59);
    }

    msb1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace msb2(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mod.MSB_2 already set");
    } else {
      filled.set(60);
    }

    msb2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace oli(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mod.OLI already set");
    } else {
      filled.set(61);
    }

    oli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mod.RES_HI already set");
    } else {
      filled.set(62);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resHi.put((byte) 0);
    }
    resHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(63)) {
      throw new IllegalStateException("mod.RES_LO already set");
    } else {
      filled.set(63);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resLo.put((byte) 0);
    }
    resLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace signed(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("mod.SIGNED already set");
    } else {
      filled.set(64);
    }

    signed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(65)) {
      throw new IllegalStateException("mod.STAMP already set");
    } else {
      filled.set(65);
    }

    stamp.putLong(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mod.ACC_1_2 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mod.ACC_1_3 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mod.ACC_2_2 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mod.ACC_2_3 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mod.ACC_B_0 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mod.ACC_B_1 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mod.ACC_B_2 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mod.ACC_B_3 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mod.ACC_DELTA_0 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mod.ACC_DELTA_1 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mod.ACC_DELTA_2 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mod.ACC_DELTA_3 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mod.ACC_H_0 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mod.ACC_H_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mod.ACC_H_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mod.ACC_Q_0 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mod.ACC_Q_1 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mod.ACC_Q_2 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mod.ACC_Q_3 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mod.ACC_R_0 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mod.ACC_R_1 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mod.ACC_R_2 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mod.ACC_R_3 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mod.ARG_1_HI has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mod.ARG_1_LO has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mod.ARG_2_HI has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mod.ARG_2_LO has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mod.BYTE_1_2 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mod.BYTE_1_3 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mod.BYTE_2_2 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mod.BYTE_2_3 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mod.BYTE_B_0 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mod.BYTE_B_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mod.BYTE_B_2 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mod.BYTE_B_3 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mod.BYTE_DELTA_0 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mod.BYTE_DELTA_1 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mod.BYTE_DELTA_2 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mod.BYTE_DELTA_3 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mod.BYTE_H_0 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mod.BYTE_H_1 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mod.BYTE_H_2 has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mod.BYTE_Q_0 has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mod.BYTE_Q_1 has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mod.BYTE_Q_2 has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mod.BYTE_Q_3 has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mod.BYTE_R_0 has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mod.BYTE_R_1 has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mod.BYTE_R_2 has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mod.BYTE_R_3 has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mod.CMP_1 has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mod.CMP_2 has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("mod.CT has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("mod.INST has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mod.IS_DIV has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mod.IS_MOD has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mod.IS_SDIV has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mod.IS_SMOD has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mod.MLI has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("mod.MSB_1 has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mod.MSB_2 has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("mod.OLI has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("mod.RES_HI has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("mod.RES_LO has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("mod.SIGNED has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("mod.STAMP has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc12.position(acc12.position() + 32);
    }

    if (!filled.get(1)) {
      acc13.position(acc13.position() + 32);
    }

    if (!filled.get(2)) {
      acc22.position(acc22.position() + 32);
    }

    if (!filled.get(3)) {
      acc23.position(acc23.position() + 32);
    }

    if (!filled.get(4)) {
      accB0.position(accB0.position() + 32);
    }

    if (!filled.get(5)) {
      accB1.position(accB1.position() + 32);
    }

    if (!filled.get(6)) {
      accB2.position(accB2.position() + 32);
    }

    if (!filled.get(7)) {
      accB3.position(accB3.position() + 32);
    }

    if (!filled.get(8)) {
      accDelta0.position(accDelta0.position() + 32);
    }

    if (!filled.get(9)) {
      accDelta1.position(accDelta1.position() + 32);
    }

    if (!filled.get(10)) {
      accDelta2.position(accDelta2.position() + 32);
    }

    if (!filled.get(11)) {
      accDelta3.position(accDelta3.position() + 32);
    }

    if (!filled.get(12)) {
      accH0.position(accH0.position() + 32);
    }

    if (!filled.get(13)) {
      accH1.position(accH1.position() + 32);
    }

    if (!filled.get(14)) {
      accH2.position(accH2.position() + 32);
    }

    if (!filled.get(15)) {
      accQ0.position(accQ0.position() + 32);
    }

    if (!filled.get(16)) {
      accQ1.position(accQ1.position() + 32);
    }

    if (!filled.get(17)) {
      accQ2.position(accQ2.position() + 32);
    }

    if (!filled.get(18)) {
      accQ3.position(accQ3.position() + 32);
    }

    if (!filled.get(19)) {
      accR0.position(accR0.position() + 32);
    }

    if (!filled.get(20)) {
      accR1.position(accR1.position() + 32);
    }

    if (!filled.get(21)) {
      accR2.position(accR2.position() + 32);
    }

    if (!filled.get(22)) {
      accR3.position(accR3.position() + 32);
    }

    if (!filled.get(23)) {
      arg1Hi.position(arg1Hi.position() + 32);
    }

    if (!filled.get(24)) {
      arg1Lo.position(arg1Lo.position() + 32);
    }

    if (!filled.get(25)) {
      arg2Hi.position(arg2Hi.position() + 32);
    }

    if (!filled.get(26)) {
      arg2Lo.position(arg2Lo.position() + 32);
    }

    if (!filled.get(27)) {
      byte12.position(byte12.position() + 1);
    }

    if (!filled.get(28)) {
      byte13.position(byte13.position() + 1);
    }

    if (!filled.get(29)) {
      byte22.position(byte22.position() + 1);
    }

    if (!filled.get(30)) {
      byte23.position(byte23.position() + 1);
    }

    if (!filled.get(31)) {
      byteB0.position(byteB0.position() + 1);
    }

    if (!filled.get(32)) {
      byteB1.position(byteB1.position() + 1);
    }

    if (!filled.get(33)) {
      byteB2.position(byteB2.position() + 1);
    }

    if (!filled.get(34)) {
      byteB3.position(byteB3.position() + 1);
    }

    if (!filled.get(35)) {
      byteDelta0.position(byteDelta0.position() + 1);
    }

    if (!filled.get(36)) {
      byteDelta1.position(byteDelta1.position() + 1);
    }

    if (!filled.get(37)) {
      byteDelta2.position(byteDelta2.position() + 1);
    }

    if (!filled.get(38)) {
      byteDelta3.position(byteDelta3.position() + 1);
    }

    if (!filled.get(39)) {
      byteH0.position(byteH0.position() + 1);
    }

    if (!filled.get(40)) {
      byteH1.position(byteH1.position() + 1);
    }

    if (!filled.get(41)) {
      byteH2.position(byteH2.position() + 1);
    }

    if (!filled.get(42)) {
      byteQ0.position(byteQ0.position() + 1);
    }

    if (!filled.get(43)) {
      byteQ1.position(byteQ1.position() + 1);
    }

    if (!filled.get(44)) {
      byteQ2.position(byteQ2.position() + 1);
    }

    if (!filled.get(45)) {
      byteQ3.position(byteQ3.position() + 1);
    }

    if (!filled.get(46)) {
      byteR0.position(byteR0.position() + 1);
    }

    if (!filled.get(47)) {
      byteR1.position(byteR1.position() + 1);
    }

    if (!filled.get(48)) {
      byteR2.position(byteR2.position() + 1);
    }

    if (!filled.get(49)) {
      byteR3.position(byteR3.position() + 1);
    }

    if (!filled.get(50)) {
      cmp1.position(cmp1.position() + 1);
    }

    if (!filled.get(51)) {
      cmp2.position(cmp2.position() + 1);
    }

    if (!filled.get(52)) {
      ct.position(ct.position() + 2);
    }

    if (!filled.get(53)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(54)) {
      isDiv.position(isDiv.position() + 1);
    }

    if (!filled.get(55)) {
      isMod.position(isMod.position() + 1);
    }

    if (!filled.get(56)) {
      isSdiv.position(isSdiv.position() + 1);
    }

    if (!filled.get(57)) {
      isSmod.position(isSmod.position() + 1);
    }

    if (!filled.get(58)) {
      mli.position(mli.position() + 1);
    }

    if (!filled.get(59)) {
      msb1.position(msb1.position() + 1);
    }

    if (!filled.get(60)) {
      msb2.position(msb2.position() + 1);
    }

    if (!filled.get(61)) {
      oli.position(oli.position() + 1);
    }

    if (!filled.get(62)) {
      resHi.position(resHi.position() + 32);
    }

    if (!filled.get(63)) {
      resLo.position(resLo.position() + 32);
    }

    if (!filled.get(64)) {
      signed.position(signed.position() + 1);
    }

    if (!filled.get(65)) {
      stamp.position(stamp.position() + 8);
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
