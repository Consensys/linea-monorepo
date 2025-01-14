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

package net.consensys.linea.zktracer.module.txndata;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
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
  public static final int BLOCKHASH_MAX_HISTORY = 0x100;
  public static final int COMMON_RLP_TXN_PHASE_NUMBER_0 = 0x1;
  public static final int COMMON_RLP_TXN_PHASE_NUMBER_1 = 0x8;
  public static final int COMMON_RLP_TXN_PHASE_NUMBER_2 = 0x3;
  public static final int COMMON_RLP_TXN_PHASE_NUMBER_3 = 0x9;
  public static final int COMMON_RLP_TXN_PHASE_NUMBER_4 = 0xa;
  public static final int COMMON_RLP_TXN_PHASE_NUMBER_5 = 0x7;
  public static final int CREATE2_SHIFT = 0xff;
  public static final int CT_MAX_TYPE_0 = 0x7;
  public static final int CT_MAX_TYPE_1 = 0x8;
  public static final int CT_MAX_TYPE_2 = 0x8;
  public static final long EIP2681_MAX_NONCE = 0xffffffffffffffffL;
  public static final int EIP_3541_MARKER = 0xef;
  public static final BigInteger EMPTY_KECCAK_HI = new BigInteger("262949717399590921288928019264691438528");
  public static final BigInteger EMPTY_KECCAK_LO = new BigInteger("304396909071904405792975023732328604784");
  public static final int EMPTY_RIPEMD_HI = 0x9c1185a5;
  public static final BigInteger EMPTY_RIPEMD_LO = new BigInteger("263072838190121256777638892741499129137");
  public static final BigInteger EMPTY_SHA2_HI = new BigInteger("302652579918965577886386472538583578916");
  public static final BigInteger EMPTY_SHA2_LO = new BigInteger("52744687940778649747319168982913824853");
  public static final long ETHEREUM_GAS_LIMIT_MAXIMUM = 0xffffffffffffffffL;
  public static final int ETHEREUM_GAS_LIMIT_MINIMUM = 0x1388;
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
  public static final int GAS_LIMIT_ADJUSTMENT_FACTOR = 0x400;
  public static final int LINEA_BASE_FEE = 0x7;
  public static final int LINEA_BLOCK_GAS_LIMIT = 0x3a2c940;
  public static final int LINEA_CHAIN_ID = 0xe708;
  public static final int LINEA_DIFFICULTY = 0x2;
  public static final int LINEA_GAS_LIMIT_MAXIMUM = 0x77359400;
  public static final int LINEA_GAS_LIMIT_MINIMUM = 0x3a2c940;
  public static final int LINEA_GOERLI_CHAIN_ID = 0xe704;
  public static final int LINEA_MAX_NUMBER_OF_TRANSACTIONS_IN_BATCH = 0xc8;
  public static final int LINEA_SEPOLIA_CHAIN_ID = 0xe705;
  public static final int LLARGE = 0x10;
  public static final int LLARGEMO = 0xf;
  public static final int LLARGEPO = 0x11;
  public static final int MAX_CODE_SIZE = 0x6000;
  public static final int MAX_REFUND_QUOTIENT = 0x5;
  public static final int MISC_WEIGHT_EXP = 0x1;
  public static final int MISC_WEIGHT_MMU = 0x2;
  public static final int MISC_WEIGHT_MXP = 0x4;
  public static final int MISC_WEIGHT_OOB = 0x8;
  public static final int MISC_WEIGHT_STP = 0x10;
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
  public static final int MMU_INST_MSTORE8 = 0xfe03;
  public static final int MMU_INST_RAM_TO_EXO_WITH_PADDING = 0xfe20;
  public static final int MMU_INST_RAM_TO_RAM_SANS_PADDING = 0xfe40;
  public static final int MMU_INST_RIGHT_PADDED_WORD_EXTRACTION = 0xfe10;
  public static final int NB_ROWS_TYPE_0 = 0x8;
  public static final int NB_ROWS_TYPE_1 = 0x9;
  public static final int NB_ROWS_TYPE_2 = 0x9;
  public static final int OOB_INST_BLAKE_CDS = 0xfa09;
  public static final int OOB_INST_BLAKE_PARAMS = 0xfb09;
  public static final int OOB_INST_CALL = 0xca;
  public static final int OOB_INST_CDL = 0x35;
  public static final int OOB_INST_CREATE = 0xce;
  public static final int OOB_INST_DEPLOYMENT = 0xf3;
  public static final int OOB_INST_ECADD = 0xff06;
  public static final int OOB_INST_ECMUL = 0xff07;
  public static final int OOB_INST_ECPAIRING = 0xff08;
  public static final int OOB_INST_ECRECOVER = 0xff01;
  public static final int OOB_INST_IDENTITY = 0xff04;
  public static final int OOB_INST_JUMP = 0x56;
  public static final int OOB_INST_JUMPI = 0x57;
  public static final int OOB_INST_MODEXP_CDS = 0xfa05;
  public static final int OOB_INST_MODEXP_EXTRACT = 0xfe05;
  public static final int OOB_INST_MODEXP_LEAD = 0xfc05;
  public static final int OOB_INST_MODEXP_PRICING = 0xfd05;
  public static final int OOB_INST_MODEXP_XBS = 0xfb05;
  public static final int OOB_INST_RDC = 0x3e;
  public static final int OOB_INST_RIPEMD = 0xff03;
  public static final int OOB_INST_SHA2 = 0xff02;
  public static final int OOB_INST_SSTORE = 0x55;
  public static final int OOB_INST_XCALL = 0xcc;
  public static final int PHASE_BLAKE_DATA = 0x5;
  public static final int PHASE_BLAKE_PARAMS = 0x6;
  public static final int PHASE_BLAKE_RESULT = 0x7;
  public static final int PHASE_ECADD_DATA = 0x60a;
  public static final int PHASE_ECADD_RESULT = 0x60b;
  public static final int PHASE_ECMUL_DATA = 0x70a;
  public static final int PHASE_ECMUL_RESULT = 0x70b;
  public static final int PHASE_ECPAIRING_DATA = 0x80a;
  public static final int PHASE_ECPAIRING_RESULT = 0x80b;
  public static final int PHASE_ECRECOVER_DATA = 0x10a;
  public static final int PHASE_ECRECOVER_RESULT = 0x10b;
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
  public static final int REFUND_CONST_R_SCLEAR = 0x12c0;
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
  public static final int TYPE_0_RLP_TXN_PHASE_NUMBER_6 = 0x4;
  public static final int TYPE_1_RLP_TXN_PHASE_NUMBER_6 = 0x4;
  public static final int TYPE_1_RLP_TXN_PHASE_NUMBER_7 = 0xb;
  public static final int TYPE_2_RLP_TXN_PHASE_NUMBER_6 = 0x6;
  public static final int TYPE_2_RLP_TXN_PHASE_NUMBER_7 = 0xb;
  public static final int WCP_INST_GEQ = 0xe;
  public static final int WCP_INST_LEQ = 0xf;
  public static final int WORD_SIZE = 0x20;
  public static final int WORD_SIZE_MO = 0x1f;
  public static final int row_offset___computing_effective_gas_price_comparison = 0x8;
  public static final int row_offset___detecting_empty_call_data_comparison = 0x5;
  public static final int row_offset___effective_refund_comparison = 0x4;
  public static final int row_offset___initial_balance_comparison = 0x1;
  public static final int row_offset___max_fee_and_basefee_comparison = 0x6;
  public static final int row_offset___max_fee_and_max_priority_fee_comparison = 0x7;
  public static final int row_offset___nonce_comparison = 0x0;
  public static final int row_offset___sufficient_gas_comparison = 0x2;
  public static final int row_offset___upper_limit_refunds_comparison = 0x3;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absTxNum;
  private final MappedByteBuffer absTxNumMax;
  private final MappedByteBuffer argOneLo;
  private final MappedByteBuffer argTwoLo;
  private final MappedByteBuffer basefee;
  private final MappedByteBuffer blockGasLimit;
  private final MappedByteBuffer callDataSize;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer coinbaseHi;
  private final MappedByteBuffer coinbaseLo;
  private final MappedByteBuffer copyTxcd;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer eucFlag;
  private final MappedByteBuffer fromHi;
  private final MappedByteBuffer fromLo;
  private final MappedByteBuffer gasCumulative;
  private final MappedByteBuffer gasInitiallyAvailable;
  private final MappedByteBuffer gasLeftover;
  private final MappedByteBuffer gasLimit;
  private final MappedByteBuffer gasPrice;
  private final MappedByteBuffer initCodeSize;
  private final MappedByteBuffer initialBalance;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isDep;
  private final MappedByteBuffer isLastTxOfBlock;
  private final MappedByteBuffer nonce;
  private final MappedByteBuffer outgoingHi;
  private final MappedByteBuffer outgoingLo;
  private final MappedByteBuffer outgoingRlpTxnrcpt;
  private final MappedByteBuffer phaseRlpTxn;
  private final MappedByteBuffer phaseRlpTxnrcpt;
  private final MappedByteBuffer priorityFeePerGas;
  private final MappedByteBuffer refundCounter;
  private final MappedByteBuffer refundEffective;
  private final MappedByteBuffer relBlock;
  private final MappedByteBuffer relTxNum;
  private final MappedByteBuffer relTxNumMax;
  private final MappedByteBuffer requiresEvmExecution;
  private final MappedByteBuffer res;
  private final MappedByteBuffer statusCode;
  private final MappedByteBuffer toHi;
  private final MappedByteBuffer toLo;
  private final MappedByteBuffer type0;
  private final MappedByteBuffer type1;
  private final MappedByteBuffer type2;
  private final MappedByteBuffer value;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("txndata.ABS_TX_NUM", 2, length));
      headers.add(new ColumnHeader("txndata.ABS_TX_NUM_MAX", 2, length));
      headers.add(new ColumnHeader("txndata.ARG_ONE_LO", 16, length));
      headers.add(new ColumnHeader("txndata.ARG_TWO_LO", 16, length));
      headers.add(new ColumnHeader("txndata.BASEFEE", 16, length));
      headers.add(new ColumnHeader("txndata.BLOCK_GAS_LIMIT", 8, length));
      headers.add(new ColumnHeader("txndata.CALL_DATA_SIZE", 4, length));
      headers.add(new ColumnHeader("txndata.CODE_FRAGMENT_INDEX", 4, length));
      headers.add(new ColumnHeader("txndata.COINBASE_HI", 4, length));
      headers.add(new ColumnHeader("txndata.COINBASE_LO", 16, length));
      headers.add(new ColumnHeader("txndata.COPY_TXCD", 1, length));
      headers.add(new ColumnHeader("txndata.CT", 1, length));
      headers.add(new ColumnHeader("txndata.EUC_FLAG", 1, length));
      headers.add(new ColumnHeader("txndata.FROM_HI", 4, length));
      headers.add(new ColumnHeader("txndata.FROM_LO", 16, length));
      headers.add(new ColumnHeader("txndata.GAS_CUMULATIVE", 16, length));
      headers.add(new ColumnHeader("txndata.GAS_INITIALLY_AVAILABLE", 16, length));
      headers.add(new ColumnHeader("txndata.GAS_LEFTOVER", 16, length));
      headers.add(new ColumnHeader("txndata.GAS_LIMIT", 8, length));
      headers.add(new ColumnHeader("txndata.GAS_PRICE", 8, length));
      headers.add(new ColumnHeader("txndata.INIT_CODE_SIZE", 4, length));
      headers.add(new ColumnHeader("txndata.INITIAL_BALANCE", 16, length));
      headers.add(new ColumnHeader("txndata.INST", 1, length));
      headers.add(new ColumnHeader("txndata.IS_DEP", 1, length));
      headers.add(new ColumnHeader("txndata.IS_LAST_TX_OF_BLOCK", 1, length));
      headers.add(new ColumnHeader("txndata.NONCE", 8, length));
      headers.add(new ColumnHeader("txndata.OUTGOING_HI", 8, length));
      headers.add(new ColumnHeader("txndata.OUTGOING_LO", 16, length));
      headers.add(new ColumnHeader("txndata.OUTGOING_RLP_TXNRCPT", 16, length));
      headers.add(new ColumnHeader("txndata.PHASE_RLP_TXN", 1, length));
      headers.add(new ColumnHeader("txndata.PHASE_RLP_TXNRCPT", 1, length));
      headers.add(new ColumnHeader("txndata.PRIORITY_FEE_PER_GAS", 16, length));
      headers.add(new ColumnHeader("txndata.REFUND_COUNTER", 16, length));
      headers.add(new ColumnHeader("txndata.REFUND_EFFECTIVE", 16, length));
      headers.add(new ColumnHeader("txndata.REL_BLOCK", 2, length));
      headers.add(new ColumnHeader("txndata.REL_TX_NUM", 2, length));
      headers.add(new ColumnHeader("txndata.REL_TX_NUM_MAX", 2, length));
      headers.add(new ColumnHeader("txndata.REQUIRES_EVM_EXECUTION", 1, length));
      headers.add(new ColumnHeader("txndata.RES", 8, length));
      headers.add(new ColumnHeader("txndata.STATUS_CODE", 1, length));
      headers.add(new ColumnHeader("txndata.TO_HI", 4, length));
      headers.add(new ColumnHeader("txndata.TO_LO", 16, length));
      headers.add(new ColumnHeader("txndata.TYPE0", 1, length));
      headers.add(new ColumnHeader("txndata.TYPE1", 1, length));
      headers.add(new ColumnHeader("txndata.TYPE2", 1, length));
      headers.add(new ColumnHeader("txndata.VALUE", 16, length));
      headers.add(new ColumnHeader("txndata.WCP_FLAG", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.absTxNum = buffers.get(0);
    this.absTxNumMax = buffers.get(1);
    this.argOneLo = buffers.get(2);
    this.argTwoLo = buffers.get(3);
    this.basefee = buffers.get(4);
    this.blockGasLimit = buffers.get(5);
    this.callDataSize = buffers.get(6);
    this.codeFragmentIndex = buffers.get(7);
    this.coinbaseHi = buffers.get(8);
    this.coinbaseLo = buffers.get(9);
    this.copyTxcd = buffers.get(10);
    this.ct = buffers.get(11);
    this.eucFlag = buffers.get(12);
    this.fromHi = buffers.get(13);
    this.fromLo = buffers.get(14);
    this.gasCumulative = buffers.get(15);
    this.gasInitiallyAvailable = buffers.get(16);
    this.gasLeftover = buffers.get(17);
    this.gasLimit = buffers.get(18);
    this.gasPrice = buffers.get(19);
    this.initCodeSize = buffers.get(20);
    this.initialBalance = buffers.get(21);
    this.inst = buffers.get(22);
    this.isDep = buffers.get(23);
    this.isLastTxOfBlock = buffers.get(24);
    this.nonce = buffers.get(25);
    this.outgoingHi = buffers.get(26);
    this.outgoingLo = buffers.get(27);
    this.outgoingRlpTxnrcpt = buffers.get(28);
    this.phaseRlpTxn = buffers.get(29);
    this.phaseRlpTxnrcpt = buffers.get(30);
    this.priorityFeePerGas = buffers.get(31);
    this.refundCounter = buffers.get(32);
    this.refundEffective = buffers.get(33);
    this.relBlock = buffers.get(34);
    this.relTxNum = buffers.get(35);
    this.relTxNumMax = buffers.get(36);
    this.requiresEvmExecution = buffers.get(37);
    this.res = buffers.get(38);
    this.statusCode = buffers.get(39);
    this.toHi = buffers.get(40);
    this.toLo = buffers.get(41);
    this.type0 = buffers.get(42);
    this.type1 = buffers.get(43);
    this.type2 = buffers.get(44);
    this.value = buffers.get(45);
    this.wcpFlag = buffers.get(46);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absTxNum(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("txndata.ABS_TX_NUM already set");
    } else {
      filled.set(0);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("txndata.ABS_TX_NUM has invalid value (" + b + ")"); }
    absTxNum.put((byte) (b >> 8));
    absTxNum.put((byte) b);


    return this;
  }

  public Trace absTxNumMax(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("txndata.ABS_TX_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("txndata.ABS_TX_NUM_MAX has invalid value (" + b + ")"); }
    absTxNumMax.put((byte) (b >> 8));
    absTxNumMax.put((byte) b);


    return this;
  }

  public Trace argOneLo(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("txndata.ARG_ONE_LO already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.ARG_ONE_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argOneLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argOneLo.put(bs.get(j)); }

    return this;
  }

  public Trace argTwoLo(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("txndata.ARG_TWO_LO already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.ARG_TWO_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { argTwoLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { argTwoLo.put(bs.get(j)); }

    return this;
  }

  public Trace basefee(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("txndata.BASEFEE already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.BASEFEE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { basefee.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { basefee.put(bs.get(j)); }

    return this;
  }

  public Trace blockGasLimit(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("txndata.BLOCK_GAS_LIMIT already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("txndata.BLOCK_GAS_LIMIT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { blockGasLimit.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { blockGasLimit.put(bs.get(j)); }

    return this;
  }

  public Trace callDataSize(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("txndata.CALL_DATA_SIZE already set");
    } else {
      filled.set(6);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("txndata.CALL_DATA_SIZE has invalid value (" + b + ")"); }
    callDataSize.put((byte) (b >> 24));
    callDataSize.put((byte) (b >> 16));
    callDataSize.put((byte) (b >> 8));
    callDataSize.put((byte) b);


    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("txndata.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(7);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("txndata.CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeFragmentIndex.put((byte) (b >> 24));
    codeFragmentIndex.put((byte) (b >> 16));
    codeFragmentIndex.put((byte) (b >> 8));
    codeFragmentIndex.put((byte) b);


    return this;
  }

  public Trace coinbaseHi(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("txndata.COINBASE_HI already set");
    } else {
      filled.set(8);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("txndata.COINBASE_HI has invalid value (" + b + ")"); }
    coinbaseHi.put((byte) (b >> 24));
    coinbaseHi.put((byte) (b >> 16));
    coinbaseHi.put((byte) (b >> 8));
    coinbaseHi.put((byte) b);


    return this;
  }

  public Trace coinbaseLo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("txndata.COINBASE_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.COINBASE_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { coinbaseLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { coinbaseLo.put(bs.get(j)); }

    return this;
  }

  public Trace copyTxcd(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("txndata.COPY_TXCD already set");
    } else {
      filled.set(10);
    }

    copyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ct(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("txndata.CT already set");
    } else {
      filled.set(11);
    }

    ct.put(b.toByte());

    return this;
  }

  public Trace eucFlag(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("txndata.EUC_FLAG already set");
    } else {
      filled.set(12);
    }

    eucFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace fromHi(final long b) {
    if (filled.get(13)) {
      throw new IllegalStateException("txndata.FROM_HI already set");
    } else {
      filled.set(13);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("txndata.FROM_HI has invalid value (" + b + ")"); }
    fromHi.put((byte) (b >> 24));
    fromHi.put((byte) (b >> 16));
    fromHi.put((byte) (b >> 8));
    fromHi.put((byte) b);


    return this;
  }

  public Trace fromLo(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("txndata.FROM_LO already set");
    } else {
      filled.set(14);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.FROM_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { fromLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { fromLo.put(bs.get(j)); }

    return this;
  }

  public Trace gasCumulative(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("txndata.GAS_CUMULATIVE already set");
    } else {
      filled.set(15);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.GAS_CUMULATIVE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { gasCumulative.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasCumulative.put(bs.get(j)); }

    return this;
  }

  public Trace gasInitiallyAvailable(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("txndata.GAS_INITIALLY_AVAILABLE already set");
    } else {
      filled.set(16);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.GAS_INITIALLY_AVAILABLE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { gasInitiallyAvailable.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasInitiallyAvailable.put(bs.get(j)); }

    return this;
  }

  public Trace gasLeftover(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("txndata.GAS_LEFTOVER already set");
    } else {
      filled.set(17);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.GAS_LEFTOVER has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { gasLeftover.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasLeftover.put(bs.get(j)); }

    return this;
  }

  public Trace gasLimit(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("txndata.GAS_LIMIT already set");
    } else {
      filled.set(18);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("txndata.GAS_LIMIT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasLimit.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasLimit.put(bs.get(j)); }

    return this;
  }

  public Trace gasPrice(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("txndata.GAS_PRICE already set");
    } else {
      filled.set(19);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("txndata.GAS_PRICE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasPrice.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasPrice.put(bs.get(j)); }

    return this;
  }

  public Trace initCodeSize(final long b) {
    if (filled.get(21)) {
      throw new IllegalStateException("txndata.INIT_CODE_SIZE already set");
    } else {
      filled.set(21);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("txndata.INIT_CODE_SIZE has invalid value (" + b + ")"); }
    initCodeSize.put((byte) (b >> 24));
    initCodeSize.put((byte) (b >> 16));
    initCodeSize.put((byte) (b >> 8));
    initCodeSize.put((byte) b);


    return this;
  }

  public Trace initialBalance(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("txndata.INITIAL_BALANCE already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.INITIAL_BALANCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { initialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { initialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("txndata.INST already set");
    } else {
      filled.set(22);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isDep(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("txndata.IS_DEP already set");
    } else {
      filled.set(23);
    }

    isDep.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLastTxOfBlock(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("txndata.IS_LAST_TX_OF_BLOCK already set");
    } else {
      filled.set(24);
    }

    isLastTxOfBlock.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace nonce(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("txndata.NONCE already set");
    } else {
      filled.set(25);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("txndata.NONCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonce.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonce.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingHi(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("txndata.OUTGOING_HI already set");
    } else {
      filled.set(26);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("txndata.OUTGOING_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { outgoingHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingHi.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingLo(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("txndata.OUTGOING_LO already set");
    } else {
      filled.set(27);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.OUTGOING_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingLo.put(bs.get(j)); }

    return this;
  }

  public Trace outgoingRlpTxnrcpt(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("txndata.OUTGOING_RLP_TXNRCPT already set");
    } else {
      filled.set(28);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.OUTGOING_RLP_TXNRCPT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { outgoingRlpTxnrcpt.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { outgoingRlpTxnrcpt.put(bs.get(j)); }

    return this;
  }

  public Trace phaseRlpTxn(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("txndata.PHASE_RLP_TXN already set");
    } else {
      filled.set(29);
    }

    phaseRlpTxn.put(b.toByte());

    return this;
  }

  public Trace phaseRlpTxnrcpt(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("txndata.PHASE_RLP_TXNRCPT already set");
    } else {
      filled.set(30);
    }

    phaseRlpTxnrcpt.put(b.toByte());

    return this;
  }

  public Trace priorityFeePerGas(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("txndata.PRIORITY_FEE_PER_GAS already set");
    } else {
      filled.set(31);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.PRIORITY_FEE_PER_GAS has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { priorityFeePerGas.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { priorityFeePerGas.put(bs.get(j)); }

    return this;
  }

  public Trace refundCounter(final Bytes b) {
    if (filled.get(32)) {
      throw new IllegalStateException("txndata.REFUND_COUNTER already set");
    } else {
      filled.set(32);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.REFUND_COUNTER has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { refundCounter.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refundCounter.put(bs.get(j)); }

    return this;
  }

  public Trace refundEffective(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("txndata.REFUND_EFFECTIVE already set");
    } else {
      filled.set(33);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.REFUND_EFFECTIVE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { refundEffective.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refundEffective.put(bs.get(j)); }

    return this;
  }

  public Trace relBlock(final long b) {
    if (filled.get(34)) {
      throw new IllegalStateException("txndata.REL_BLOCK already set");
    } else {
      filled.set(34);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("txndata.REL_BLOCK has invalid value (" + b + ")"); }
    relBlock.put((byte) (b >> 8));
    relBlock.put((byte) b);


    return this;
  }

  public Trace relTxNum(final long b) {
    if (filled.get(35)) {
      throw new IllegalStateException("txndata.REL_TX_NUM already set");
    } else {
      filled.set(35);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("txndata.REL_TX_NUM has invalid value (" + b + ")"); }
    relTxNum.put((byte) (b >> 8));
    relTxNum.put((byte) b);


    return this;
  }

  public Trace relTxNumMax(final long b) {
    if (filled.get(36)) {
      throw new IllegalStateException("txndata.REL_TX_NUM_MAX already set");
    } else {
      filled.set(36);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("txndata.REL_TX_NUM_MAX has invalid value (" + b + ")"); }
    relTxNumMax.put((byte) (b >> 8));
    relTxNumMax.put((byte) b);


    return this;
  }

  public Trace requiresEvmExecution(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("txndata.REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(37);
    }

    requiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace res(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("txndata.RES already set");
    } else {
      filled.set(38);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("txndata.RES has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { res.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { res.put(bs.get(j)); }

    return this;
  }

  public Trace statusCode(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("txndata.STATUS_CODE already set");
    } else {
      filled.set(39);
    }

    statusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace toHi(final long b) {
    if (filled.get(40)) {
      throw new IllegalStateException("txndata.TO_HI already set");
    } else {
      filled.set(40);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("txndata.TO_HI has invalid value (" + b + ")"); }
    toHi.put((byte) (b >> 24));
    toHi.put((byte) (b >> 16));
    toHi.put((byte) (b >> 8));
    toHi.put((byte) b);


    return this;
  }

  public Trace toLo(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("txndata.TO_LO already set");
    } else {
      filled.set(41);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.TO_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { toLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { toLo.put(bs.get(j)); }

    return this;
  }

  public Trace type0(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("txndata.TYPE0 already set");
    } else {
      filled.set(42);
    }

    type0.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace type1(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("txndata.TYPE1 already set");
    } else {
      filled.set(43);
    }

    type1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace type2(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("txndata.TYPE2 already set");
    } else {
      filled.set(44);
    }

    type2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace value(final Bytes b) {
    if (filled.get(45)) {
      throw new IllegalStateException("txndata.VALUE already set");
    } else {
      filled.set(45);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("txndata.VALUE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { value.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { value.put(bs.get(j)); }

    return this;
  }

  public Trace wcpFlag(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("txndata.WCP_FLAG already set");
    } else {
      filled.set(46);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("txndata.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("txndata.ABS_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("txndata.ARG_ONE_LO has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("txndata.ARG_TWO_LO has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("txndata.BASEFEE has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("txndata.BLOCK_GAS_LIMIT has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("txndata.CALL_DATA_SIZE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("txndata.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("txndata.COINBASE_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("txndata.COINBASE_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("txndata.COPY_TXCD has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("txndata.CT has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("txndata.EUC_FLAG has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("txndata.FROM_HI has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("txndata.FROM_LO has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("txndata.GAS_CUMULATIVE has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("txndata.GAS_INITIALLY_AVAILABLE has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("txndata.GAS_LEFTOVER has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("txndata.GAS_LIMIT has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("txndata.GAS_PRICE has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("txndata.INIT_CODE_SIZE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("txndata.INITIAL_BALANCE has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("txndata.INST has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("txndata.IS_DEP has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("txndata.IS_LAST_TX_OF_BLOCK has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("txndata.NONCE has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("txndata.OUTGOING_HI has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("txndata.OUTGOING_LO has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("txndata.OUTGOING_RLP_TXNRCPT has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("txndata.PHASE_RLP_TXN has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("txndata.PHASE_RLP_TXNRCPT has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("txndata.PRIORITY_FEE_PER_GAS has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("txndata.REFUND_COUNTER has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("txndata.REFUND_EFFECTIVE has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("txndata.REL_BLOCK has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("txndata.REL_TX_NUM has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("txndata.REL_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("txndata.REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("txndata.RES has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("txndata.STATUS_CODE has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("txndata.TO_HI has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("txndata.TO_LO has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("txndata.TYPE0 has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("txndata.TYPE1 has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("txndata.TYPE2 has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("txndata.VALUE has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("txndata.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absTxNum.position(absTxNum.position() + 2);
    }

    if (!filled.get(1)) {
      absTxNumMax.position(absTxNumMax.position() + 2);
    }

    if (!filled.get(2)) {
      argOneLo.position(argOneLo.position() + 16);
    }

    if (!filled.get(3)) {
      argTwoLo.position(argTwoLo.position() + 16);
    }

    if (!filled.get(4)) {
      basefee.position(basefee.position() + 16);
    }

    if (!filled.get(5)) {
      blockGasLimit.position(blockGasLimit.position() + 8);
    }

    if (!filled.get(6)) {
      callDataSize.position(callDataSize.position() + 4);
    }

    if (!filled.get(7)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 4);
    }

    if (!filled.get(8)) {
      coinbaseHi.position(coinbaseHi.position() + 4);
    }

    if (!filled.get(9)) {
      coinbaseLo.position(coinbaseLo.position() + 16);
    }

    if (!filled.get(10)) {
      copyTxcd.position(copyTxcd.position() + 1);
    }

    if (!filled.get(11)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(12)) {
      eucFlag.position(eucFlag.position() + 1);
    }

    if (!filled.get(13)) {
      fromHi.position(fromHi.position() + 4);
    }

    if (!filled.get(14)) {
      fromLo.position(fromLo.position() + 16);
    }

    if (!filled.get(15)) {
      gasCumulative.position(gasCumulative.position() + 16);
    }

    if (!filled.get(16)) {
      gasInitiallyAvailable.position(gasInitiallyAvailable.position() + 16);
    }

    if (!filled.get(17)) {
      gasLeftover.position(gasLeftover.position() + 16);
    }

    if (!filled.get(18)) {
      gasLimit.position(gasLimit.position() + 8);
    }

    if (!filled.get(19)) {
      gasPrice.position(gasPrice.position() + 8);
    }

    if (!filled.get(21)) {
      initCodeSize.position(initCodeSize.position() + 4);
    }

    if (!filled.get(20)) {
      initialBalance.position(initialBalance.position() + 16);
    }

    if (!filled.get(22)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(23)) {
      isDep.position(isDep.position() + 1);
    }

    if (!filled.get(24)) {
      isLastTxOfBlock.position(isLastTxOfBlock.position() + 1);
    }

    if (!filled.get(25)) {
      nonce.position(nonce.position() + 8);
    }

    if (!filled.get(26)) {
      outgoingHi.position(outgoingHi.position() + 8);
    }

    if (!filled.get(27)) {
      outgoingLo.position(outgoingLo.position() + 16);
    }

    if (!filled.get(28)) {
      outgoingRlpTxnrcpt.position(outgoingRlpTxnrcpt.position() + 16);
    }

    if (!filled.get(29)) {
      phaseRlpTxn.position(phaseRlpTxn.position() + 1);
    }

    if (!filled.get(30)) {
      phaseRlpTxnrcpt.position(phaseRlpTxnrcpt.position() + 1);
    }

    if (!filled.get(31)) {
      priorityFeePerGas.position(priorityFeePerGas.position() + 16);
    }

    if (!filled.get(32)) {
      refundCounter.position(refundCounter.position() + 16);
    }

    if (!filled.get(33)) {
      refundEffective.position(refundEffective.position() + 16);
    }

    if (!filled.get(34)) {
      relBlock.position(relBlock.position() + 2);
    }

    if (!filled.get(35)) {
      relTxNum.position(relTxNum.position() + 2);
    }

    if (!filled.get(36)) {
      relTxNumMax.position(relTxNumMax.position() + 2);
    }

    if (!filled.get(37)) {
      requiresEvmExecution.position(requiresEvmExecution.position() + 1);
    }

    if (!filled.get(38)) {
      res.position(res.position() + 8);
    }

    if (!filled.get(39)) {
      statusCode.position(statusCode.position() + 1);
    }

    if (!filled.get(40)) {
      toHi.position(toHi.position() + 4);
    }

    if (!filled.get(41)) {
      toLo.position(toLo.position() + 16);
    }

    if (!filled.get(42)) {
      type0.position(type0.position() + 1);
    }

    if (!filled.get(43)) {
      type1.position(type1.position() + 1);
    }

    if (!filled.get(44)) {
      type2.position(type2.position() + 1);
    }

    if (!filled.get(45)) {
      value.position(value.position() + 16);
    }

    if (!filled.get(46)) {
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
