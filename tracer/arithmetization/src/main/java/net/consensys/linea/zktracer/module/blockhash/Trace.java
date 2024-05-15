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

package net.consensys.linea.zktracer.module.blockhash;

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
  public static final int EVM_INST_BLOCKHASH_MAX_HISTORY = 0x100;
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
  public static final int LINEA_BASE_FEE = 0x0;
  public static final int LINEA_BLOCK_GAS_LIMIT = 0x3a2c940;
  public static final int LINEA_CHAIN_ID = 0xe708;
  public static final int LINEA_GOERLI_CHAIN_ID = 0xe704;
  public static final int LINEA_SEPOLIA_CHAIN_ID = 0xe705;
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

  private final MappedByteBuffer absBlock;
  private final MappedByteBuffer blockHashHi;
  private final MappedByteBuffer blockHashLo;
  private final MappedByteBuffer blockNumberHi;
  private final MappedByteBuffer blockNumberLo;
  private final MappedByteBuffer byteHi0;
  private final MappedByteBuffer byteHi1;
  private final MappedByteBuffer byteHi10;
  private final MappedByteBuffer byteHi11;
  private final MappedByteBuffer byteHi12;
  private final MappedByteBuffer byteHi13;
  private final MappedByteBuffer byteHi14;
  private final MappedByteBuffer byteHi15;
  private final MappedByteBuffer byteHi2;
  private final MappedByteBuffer byteHi3;
  private final MappedByteBuffer byteHi4;
  private final MappedByteBuffer byteHi5;
  private final MappedByteBuffer byteHi6;
  private final MappedByteBuffer byteHi7;
  private final MappedByteBuffer byteHi8;
  private final MappedByteBuffer byteHi9;
  private final MappedByteBuffer byteLo0;
  private final MappedByteBuffer byteLo1;
  private final MappedByteBuffer byteLo10;
  private final MappedByteBuffer byteLo11;
  private final MappedByteBuffer byteLo12;
  private final MappedByteBuffer byteLo13;
  private final MappedByteBuffer byteLo14;
  private final MappedByteBuffer byteLo15;
  private final MappedByteBuffer byteLo2;
  private final MappedByteBuffer byteLo3;
  private final MappedByteBuffer byteLo4;
  private final MappedByteBuffer byteLo5;
  private final MappedByteBuffer byteLo6;
  private final MappedByteBuffer byteLo7;
  private final MappedByteBuffer byteLo8;
  private final MappedByteBuffer byteLo9;
  private final MappedByteBuffer inRange;
  private final MappedByteBuffer iomf;
  private final MappedByteBuffer lowerBoundCheck;
  private final MappedByteBuffer relBlock;
  private final MappedByteBuffer resHi;
  private final MappedByteBuffer resLo;
  private final MappedByteBuffer upperBoundCheck;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("blockhash.ABS_BLOCK", 8, length),
        new ColumnHeader("blockhash.BLOCK_HASH_HI", 32, length),
        new ColumnHeader("blockhash.BLOCK_HASH_LO", 32, length),
        new ColumnHeader("blockhash.BLOCK_NUMBER_HI", 32, length),
        new ColumnHeader("blockhash.BLOCK_NUMBER_LO", 32, length),
        new ColumnHeader("blockhash.BYTE_HI_0", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_1", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_10", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_11", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_12", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_13", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_14", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_15", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_2", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_3", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_4", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_5", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_6", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_7", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_8", 1, length),
        new ColumnHeader("blockhash.BYTE_HI_9", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_0", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_1", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_10", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_11", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_12", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_13", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_14", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_15", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_2", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_3", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_4", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_5", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_6", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_7", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_8", 1, length),
        new ColumnHeader("blockhash.BYTE_LO_9", 1, length),
        new ColumnHeader("blockhash.IN_RANGE", 1, length),
        new ColumnHeader("blockhash.IOMF", 1, length),
        new ColumnHeader("blockhash.LOWER_BOUND_CHECK", 1, length),
        new ColumnHeader("blockhash.REL_BLOCK", 2, length),
        new ColumnHeader("blockhash.RES_HI", 32, length),
        new ColumnHeader("blockhash.RES_LO", 32, length),
        new ColumnHeader("blockhash.UPPER_BOUND_CHECK", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absBlock = buffers.get(0);
    this.blockHashHi = buffers.get(1);
    this.blockHashLo = buffers.get(2);
    this.blockNumberHi = buffers.get(3);
    this.blockNumberLo = buffers.get(4);
    this.byteHi0 = buffers.get(5);
    this.byteHi1 = buffers.get(6);
    this.byteHi10 = buffers.get(7);
    this.byteHi11 = buffers.get(8);
    this.byteHi12 = buffers.get(9);
    this.byteHi13 = buffers.get(10);
    this.byteHi14 = buffers.get(11);
    this.byteHi15 = buffers.get(12);
    this.byteHi2 = buffers.get(13);
    this.byteHi3 = buffers.get(14);
    this.byteHi4 = buffers.get(15);
    this.byteHi5 = buffers.get(16);
    this.byteHi6 = buffers.get(17);
    this.byteHi7 = buffers.get(18);
    this.byteHi8 = buffers.get(19);
    this.byteHi9 = buffers.get(20);
    this.byteLo0 = buffers.get(21);
    this.byteLo1 = buffers.get(22);
    this.byteLo10 = buffers.get(23);
    this.byteLo11 = buffers.get(24);
    this.byteLo12 = buffers.get(25);
    this.byteLo13 = buffers.get(26);
    this.byteLo14 = buffers.get(27);
    this.byteLo15 = buffers.get(28);
    this.byteLo2 = buffers.get(29);
    this.byteLo3 = buffers.get(30);
    this.byteLo4 = buffers.get(31);
    this.byteLo5 = buffers.get(32);
    this.byteLo6 = buffers.get(33);
    this.byteLo7 = buffers.get(34);
    this.byteLo8 = buffers.get(35);
    this.byteLo9 = buffers.get(36);
    this.inRange = buffers.get(37);
    this.iomf = buffers.get(38);
    this.lowerBoundCheck = buffers.get(39);
    this.relBlock = buffers.get(40);
    this.resHi = buffers.get(41);
    this.resLo = buffers.get(42);
    this.upperBoundCheck = buffers.get(43);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absBlock(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("blockhash.ABS_BLOCK already set");
    } else {
      filled.set(0);
    }

    absBlock.putLong(b);

    return this;
  }

  public Trace blockHashHi(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_HI already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      blockHashHi.put((byte) 0);
    }
    blockHashHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace blockHashLo(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_LO already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      blockHashLo.put((byte) 0);
    }
    blockHashLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace blockNumberHi(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_HI already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      blockNumberHi.put((byte) 0);
    }
    blockNumberHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace blockNumberLo(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_LO already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      blockNumberLo.put((byte) 0);
    }
    blockNumberLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byteHi0(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("blockhash.BYTE_HI_0 already set");
    } else {
      filled.set(5);
    }

    byteHi0.put(b.toByte());

    return this;
  }

  public Trace byteHi1(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("blockhash.BYTE_HI_1 already set");
    } else {
      filled.set(6);
    }

    byteHi1.put(b.toByte());

    return this;
  }

  public Trace byteHi10(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("blockhash.BYTE_HI_10 already set");
    } else {
      filled.set(7);
    }

    byteHi10.put(b.toByte());

    return this;
  }

  public Trace byteHi11(final UnsignedByte b) {
    if (filled.get(8)) {
      throw new IllegalStateException("blockhash.BYTE_HI_11 already set");
    } else {
      filled.set(8);
    }

    byteHi11.put(b.toByte());

    return this;
  }

  public Trace byteHi12(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blockhash.BYTE_HI_12 already set");
    } else {
      filled.set(9);
    }

    byteHi12.put(b.toByte());

    return this;
  }

  public Trace byteHi13(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blockhash.BYTE_HI_13 already set");
    } else {
      filled.set(10);
    }

    byteHi13.put(b.toByte());

    return this;
  }

  public Trace byteHi14(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blockhash.BYTE_HI_14 already set");
    } else {
      filled.set(11);
    }

    byteHi14.put(b.toByte());

    return this;
  }

  public Trace byteHi15(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blockhash.BYTE_HI_15 already set");
    } else {
      filled.set(12);
    }

    byteHi15.put(b.toByte());

    return this;
  }

  public Trace byteHi2(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("blockhash.BYTE_HI_2 already set");
    } else {
      filled.set(13);
    }

    byteHi2.put(b.toByte());

    return this;
  }

  public Trace byteHi3(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("blockhash.BYTE_HI_3 already set");
    } else {
      filled.set(14);
    }

    byteHi3.put(b.toByte());

    return this;
  }

  public Trace byteHi4(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("blockhash.BYTE_HI_4 already set");
    } else {
      filled.set(15);
    }

    byteHi4.put(b.toByte());

    return this;
  }

  public Trace byteHi5(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("blockhash.BYTE_HI_5 already set");
    } else {
      filled.set(16);
    }

    byteHi5.put(b.toByte());

    return this;
  }

  public Trace byteHi6(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("blockhash.BYTE_HI_6 already set");
    } else {
      filled.set(17);
    }

    byteHi6.put(b.toByte());

    return this;
  }

  public Trace byteHi7(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("blockhash.BYTE_HI_7 already set");
    } else {
      filled.set(18);
    }

    byteHi7.put(b.toByte());

    return this;
  }

  public Trace byteHi8(final UnsignedByte b) {
    if (filled.get(19)) {
      throw new IllegalStateException("blockhash.BYTE_HI_8 already set");
    } else {
      filled.set(19);
    }

    byteHi8.put(b.toByte());

    return this;
  }

  public Trace byteHi9(final UnsignedByte b) {
    if (filled.get(20)) {
      throw new IllegalStateException("blockhash.BYTE_HI_9 already set");
    } else {
      filled.set(20);
    }

    byteHi9.put(b.toByte());

    return this;
  }

  public Trace byteLo0(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("blockhash.BYTE_LO_0 already set");
    } else {
      filled.set(21);
    }

    byteLo0.put(b.toByte());

    return this;
  }

  public Trace byteLo1(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("blockhash.BYTE_LO_1 already set");
    } else {
      filled.set(22);
    }

    byteLo1.put(b.toByte());

    return this;
  }

  public Trace byteLo10(final UnsignedByte b) {
    if (filled.get(23)) {
      throw new IllegalStateException("blockhash.BYTE_LO_10 already set");
    } else {
      filled.set(23);
    }

    byteLo10.put(b.toByte());

    return this;
  }

  public Trace byteLo11(final UnsignedByte b) {
    if (filled.get(24)) {
      throw new IllegalStateException("blockhash.BYTE_LO_11 already set");
    } else {
      filled.set(24);
    }

    byteLo11.put(b.toByte());

    return this;
  }

  public Trace byteLo12(final UnsignedByte b) {
    if (filled.get(25)) {
      throw new IllegalStateException("blockhash.BYTE_LO_12 already set");
    } else {
      filled.set(25);
    }

    byteLo12.put(b.toByte());

    return this;
  }

  public Trace byteLo13(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("blockhash.BYTE_LO_13 already set");
    } else {
      filled.set(26);
    }

    byteLo13.put(b.toByte());

    return this;
  }

  public Trace byteLo14(final UnsignedByte b) {
    if (filled.get(27)) {
      throw new IllegalStateException("blockhash.BYTE_LO_14 already set");
    } else {
      filled.set(27);
    }

    byteLo14.put(b.toByte());

    return this;
  }

  public Trace byteLo15(final UnsignedByte b) {
    if (filled.get(28)) {
      throw new IllegalStateException("blockhash.BYTE_LO_15 already set");
    } else {
      filled.set(28);
    }

    byteLo15.put(b.toByte());

    return this;
  }

  public Trace byteLo2(final UnsignedByte b) {
    if (filled.get(29)) {
      throw new IllegalStateException("blockhash.BYTE_LO_2 already set");
    } else {
      filled.set(29);
    }

    byteLo2.put(b.toByte());

    return this;
  }

  public Trace byteLo3(final UnsignedByte b) {
    if (filled.get(30)) {
      throw new IllegalStateException("blockhash.BYTE_LO_3 already set");
    } else {
      filled.set(30);
    }

    byteLo3.put(b.toByte());

    return this;
  }

  public Trace byteLo4(final UnsignedByte b) {
    if (filled.get(31)) {
      throw new IllegalStateException("blockhash.BYTE_LO_4 already set");
    } else {
      filled.set(31);
    }

    byteLo4.put(b.toByte());

    return this;
  }

  public Trace byteLo5(final UnsignedByte b) {
    if (filled.get(32)) {
      throw new IllegalStateException("blockhash.BYTE_LO_5 already set");
    } else {
      filled.set(32);
    }

    byteLo5.put(b.toByte());

    return this;
  }

  public Trace byteLo6(final UnsignedByte b) {
    if (filled.get(33)) {
      throw new IllegalStateException("blockhash.BYTE_LO_6 already set");
    } else {
      filled.set(33);
    }

    byteLo6.put(b.toByte());

    return this;
  }

  public Trace byteLo7(final UnsignedByte b) {
    if (filled.get(34)) {
      throw new IllegalStateException("blockhash.BYTE_LO_7 already set");
    } else {
      filled.set(34);
    }

    byteLo7.put(b.toByte());

    return this;
  }

  public Trace byteLo8(final UnsignedByte b) {
    if (filled.get(35)) {
      throw new IllegalStateException("blockhash.BYTE_LO_8 already set");
    } else {
      filled.set(35);
    }

    byteLo8.put(b.toByte());

    return this;
  }

  public Trace byteLo9(final UnsignedByte b) {
    if (filled.get(36)) {
      throw new IllegalStateException("blockhash.BYTE_LO_9 already set");
    } else {
      filled.set(36);
    }

    byteLo9.put(b.toByte());

    return this;
  }

  public Trace inRange(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("blockhash.IN_RANGE already set");
    } else {
      filled.set(37);
    }

    inRange.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace iomf(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("blockhash.IOMF already set");
    } else {
      filled.set(38);
    }

    iomf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lowerBoundCheck(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("blockhash.LOWER_BOUND_CHECK already set");
    } else {
      filled.set(39);
    }

    lowerBoundCheck.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace relBlock(final short b) {
    if (filled.get(40)) {
      throw new IllegalStateException("blockhash.REL_BLOCK already set");
    } else {
      filled.set(40);
    }

    relBlock.putShort(b);

    return this;
  }

  public Trace resHi(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("blockhash.RES_HI already set");
    } else {
      filled.set(41);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resHi.put((byte) 0);
    }
    resHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace resLo(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("blockhash.RES_LO already set");
    } else {
      filled.set(42);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resLo.put((byte) 0);
    }
    resLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace upperBoundCheck(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("blockhash.UPPER_BOUND_CHECK already set");
    } else {
      filled.set(43);
    }

    upperBoundCheck.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("blockhash.ABS_BLOCK has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_HI has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("blockhash.BLOCK_HASH_LO has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_HI has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("blockhash.BLOCK_NUMBER_LO has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("blockhash.BYTE_HI_0 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("blockhash.BYTE_HI_1 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("blockhash.BYTE_HI_10 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("blockhash.BYTE_HI_11 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("blockhash.BYTE_HI_12 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("blockhash.BYTE_HI_13 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("blockhash.BYTE_HI_14 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("blockhash.BYTE_HI_15 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("blockhash.BYTE_HI_2 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("blockhash.BYTE_HI_3 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("blockhash.BYTE_HI_4 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("blockhash.BYTE_HI_5 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("blockhash.BYTE_HI_6 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("blockhash.BYTE_HI_7 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("blockhash.BYTE_HI_8 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("blockhash.BYTE_HI_9 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("blockhash.BYTE_LO_0 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("blockhash.BYTE_LO_1 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("blockhash.BYTE_LO_10 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("blockhash.BYTE_LO_11 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("blockhash.BYTE_LO_12 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("blockhash.BYTE_LO_13 has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("blockhash.BYTE_LO_14 has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("blockhash.BYTE_LO_15 has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("blockhash.BYTE_LO_2 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("blockhash.BYTE_LO_3 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("blockhash.BYTE_LO_4 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("blockhash.BYTE_LO_5 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("blockhash.BYTE_LO_6 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("blockhash.BYTE_LO_7 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("blockhash.BYTE_LO_8 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("blockhash.BYTE_LO_9 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("blockhash.IN_RANGE has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("blockhash.IOMF has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("blockhash.LOWER_BOUND_CHECK has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("blockhash.REL_BLOCK has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("blockhash.RES_HI has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("blockhash.RES_LO has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("blockhash.UPPER_BOUND_CHECK has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absBlock.position(absBlock.position() + 8);
    }

    if (!filled.get(1)) {
      blockHashHi.position(blockHashHi.position() + 32);
    }

    if (!filled.get(2)) {
      blockHashLo.position(blockHashLo.position() + 32);
    }

    if (!filled.get(3)) {
      blockNumberHi.position(blockNumberHi.position() + 32);
    }

    if (!filled.get(4)) {
      blockNumberLo.position(blockNumberLo.position() + 32);
    }

    if (!filled.get(5)) {
      byteHi0.position(byteHi0.position() + 1);
    }

    if (!filled.get(6)) {
      byteHi1.position(byteHi1.position() + 1);
    }

    if (!filled.get(7)) {
      byteHi10.position(byteHi10.position() + 1);
    }

    if (!filled.get(8)) {
      byteHi11.position(byteHi11.position() + 1);
    }

    if (!filled.get(9)) {
      byteHi12.position(byteHi12.position() + 1);
    }

    if (!filled.get(10)) {
      byteHi13.position(byteHi13.position() + 1);
    }

    if (!filled.get(11)) {
      byteHi14.position(byteHi14.position() + 1);
    }

    if (!filled.get(12)) {
      byteHi15.position(byteHi15.position() + 1);
    }

    if (!filled.get(13)) {
      byteHi2.position(byteHi2.position() + 1);
    }

    if (!filled.get(14)) {
      byteHi3.position(byteHi3.position() + 1);
    }

    if (!filled.get(15)) {
      byteHi4.position(byteHi4.position() + 1);
    }

    if (!filled.get(16)) {
      byteHi5.position(byteHi5.position() + 1);
    }

    if (!filled.get(17)) {
      byteHi6.position(byteHi6.position() + 1);
    }

    if (!filled.get(18)) {
      byteHi7.position(byteHi7.position() + 1);
    }

    if (!filled.get(19)) {
      byteHi8.position(byteHi8.position() + 1);
    }

    if (!filled.get(20)) {
      byteHi9.position(byteHi9.position() + 1);
    }

    if (!filled.get(21)) {
      byteLo0.position(byteLo0.position() + 1);
    }

    if (!filled.get(22)) {
      byteLo1.position(byteLo1.position() + 1);
    }

    if (!filled.get(23)) {
      byteLo10.position(byteLo10.position() + 1);
    }

    if (!filled.get(24)) {
      byteLo11.position(byteLo11.position() + 1);
    }

    if (!filled.get(25)) {
      byteLo12.position(byteLo12.position() + 1);
    }

    if (!filled.get(26)) {
      byteLo13.position(byteLo13.position() + 1);
    }

    if (!filled.get(27)) {
      byteLo14.position(byteLo14.position() + 1);
    }

    if (!filled.get(28)) {
      byteLo15.position(byteLo15.position() + 1);
    }

    if (!filled.get(29)) {
      byteLo2.position(byteLo2.position() + 1);
    }

    if (!filled.get(30)) {
      byteLo3.position(byteLo3.position() + 1);
    }

    if (!filled.get(31)) {
      byteLo4.position(byteLo4.position() + 1);
    }

    if (!filled.get(32)) {
      byteLo5.position(byteLo5.position() + 1);
    }

    if (!filled.get(33)) {
      byteLo6.position(byteLo6.position() + 1);
    }

    if (!filled.get(34)) {
      byteLo7.position(byteLo7.position() + 1);
    }

    if (!filled.get(35)) {
      byteLo8.position(byteLo8.position() + 1);
    }

    if (!filled.get(36)) {
      byteLo9.position(byteLo9.position() + 1);
    }

    if (!filled.get(37)) {
      inRange.position(inRange.position() + 1);
    }

    if (!filled.get(38)) {
      iomf.position(iomf.position() + 1);
    }

    if (!filled.get(39)) {
      lowerBoundCheck.position(lowerBoundCheck.position() + 1);
    }

    if (!filled.get(40)) {
      relBlock.position(relBlock.position() + 2);
    }

    if (!filled.get(41)) {
      resHi.position(resHi.position() + 32);
    }

    if (!filled.get(42)) {
      resLo.position(resLo.position() + 32);
    }

    if (!filled.get(43)) {
      upperBoundCheck.position(upperBoundCheck.position() + 1);
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
