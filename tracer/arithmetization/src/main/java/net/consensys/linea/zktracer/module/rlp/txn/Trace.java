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

package net.consensys.linea.zktracer.module.rlp.txn;

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

  private final MappedByteBuffer absTxNum;
  private final MappedByteBuffer absTxNumInfiny;
  private final MappedByteBuffer acc1;
  private final MappedByteBuffer acc2;
  private final MappedByteBuffer accBytesize;
  private final MappedByteBuffer accessTupleBytesize;
  private final MappedByteBuffer addrHi;
  private final MappedByteBuffer addrLo;
  private final MappedByteBuffer bit;
  private final MappedByteBuffer bitAcc;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer dataGasCost;
  private final MappedByteBuffer dataHi;
  private final MappedByteBuffer dataLo;
  private final MappedByteBuffer depth1;
  private final MappedByteBuffer depth2;
  private final MappedByteBuffer done;
  private final MappedByteBuffer indexData;
  private final MappedByteBuffer indexLt;
  private final MappedByteBuffer indexLx;
  private final MappedByteBuffer input1;
  private final MappedByteBuffer input2;
  private final MappedByteBuffer isPrefix;
  private final MappedByteBuffer lcCorrection;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer limbConstructed;
  private final MappedByteBuffer lt;
  private final MappedByteBuffer lx;
  private final MappedByteBuffer nAddr;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nKeys;
  private final MappedByteBuffer nKeysPerAddr;
  private final MappedByteBuffer nStep;
  private final MappedByteBuffer phase1;
  private final MappedByteBuffer phase10;
  private final MappedByteBuffer phase11;
  private final MappedByteBuffer phase12;
  private final MappedByteBuffer phase13;
  private final MappedByteBuffer phase14;
  private final MappedByteBuffer phase15;
  private final MappedByteBuffer phase2;
  private final MappedByteBuffer phase3;
  private final MappedByteBuffer phase4;
  private final MappedByteBuffer phase5;
  private final MappedByteBuffer phase6;
  private final MappedByteBuffer phase7;
  private final MappedByteBuffer phase8;
  private final MappedByteBuffer phase9;
  private final MappedByteBuffer phaseEnd;
  private final MappedByteBuffer phaseId;
  private final MappedByteBuffer phaseSize;
  private final MappedByteBuffer power;
  private final MappedByteBuffer requiresEvmExecution;
  private final MappedByteBuffer rlpLtBytesize;
  private final MappedByteBuffer rlpLxBytesize;
  private final MappedByteBuffer type;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("rlpTxn.ABS_TX_NUM", 8, length),
        new ColumnHeader("rlpTxn.ABS_TX_NUM_INFINY", 4, length),
        new ColumnHeader("rlpTxn.ACC_1", 32, length),
        new ColumnHeader("rlpTxn.ACC_2", 32, length),
        new ColumnHeader("rlpTxn.ACC_BYTESIZE", 2, length),
        new ColumnHeader("rlpTxn.ACCESS_TUPLE_BYTESIZE", 4, length),
        new ColumnHeader("rlpTxn.ADDR_HI", 8, length),
        new ColumnHeader("rlpTxn.ADDR_LO", 32, length),
        new ColumnHeader("rlpTxn.BIT", 1, length),
        new ColumnHeader("rlpTxn.BIT_ACC", 1, length),
        new ColumnHeader("rlpTxn.BYTE_1", 1, length),
        new ColumnHeader("rlpTxn.BYTE_2", 1, length),
        new ColumnHeader("rlpTxn.CODE_FRAGMENT_INDEX", 8, length),
        new ColumnHeader("rlpTxn.COUNTER", 2, length),
        new ColumnHeader("rlpTxn.DATA_GAS_COST", 8, length),
        new ColumnHeader("rlpTxn.DATA_HI", 32, length),
        new ColumnHeader("rlpTxn.DATA_LO", 32, length),
        new ColumnHeader("rlpTxn.DEPTH_1", 1, length),
        new ColumnHeader("rlpTxn.DEPTH_2", 1, length),
        new ColumnHeader("rlpTxn.DONE", 1, length),
        new ColumnHeader("rlpTxn.INDEX_DATA", 8, length),
        new ColumnHeader("rlpTxn.INDEX_LT", 8, length),
        new ColumnHeader("rlpTxn.INDEX_LX", 8, length),
        new ColumnHeader("rlpTxn.INPUT_1", 32, length),
        new ColumnHeader("rlpTxn.INPUT_2", 32, length),
        new ColumnHeader("rlpTxn.IS_PREFIX", 1, length),
        new ColumnHeader("rlpTxn.LC_CORRECTION", 1, length),
        new ColumnHeader("rlpTxn.LIMB", 32, length),
        new ColumnHeader("rlpTxn.LIMB_CONSTRUCTED", 1, length),
        new ColumnHeader("rlpTxn.LT", 1, length),
        new ColumnHeader("rlpTxn.LX", 1, length),
        new ColumnHeader("rlpTxn.nADDR", 4, length),
        new ColumnHeader("rlpTxn.nBYTES", 2, length),
        new ColumnHeader("rlpTxn.nKEYS", 4, length),
        new ColumnHeader("rlpTxn.nKEYS_PER_ADDR", 4, length),
        new ColumnHeader("rlpTxn.nSTEP", 2, length),
        new ColumnHeader("rlpTxn.PHASE_1", 1, length),
        new ColumnHeader("rlpTxn.PHASE_10", 1, length),
        new ColumnHeader("rlpTxn.PHASE_11", 1, length),
        new ColumnHeader("rlpTxn.PHASE_12", 1, length),
        new ColumnHeader("rlpTxn.PHASE_13", 1, length),
        new ColumnHeader("rlpTxn.PHASE_14", 1, length),
        new ColumnHeader("rlpTxn.PHASE_15", 1, length),
        new ColumnHeader("rlpTxn.PHASE_2", 1, length),
        new ColumnHeader("rlpTxn.PHASE_3", 1, length),
        new ColumnHeader("rlpTxn.PHASE_4", 1, length),
        new ColumnHeader("rlpTxn.PHASE_5", 1, length),
        new ColumnHeader("rlpTxn.PHASE_6", 1, length),
        new ColumnHeader("rlpTxn.PHASE_7", 1, length),
        new ColumnHeader("rlpTxn.PHASE_8", 1, length),
        new ColumnHeader("rlpTxn.PHASE_9", 1, length),
        new ColumnHeader("rlpTxn.PHASE_END", 1, length),
        new ColumnHeader("rlpTxn.PHASE_ID", 2, length),
        new ColumnHeader("rlpTxn.PHASE_SIZE", 8, length),
        new ColumnHeader("rlpTxn.POWER", 32, length),
        new ColumnHeader("rlpTxn.REQUIRES_EVM_EXECUTION", 1, length),
        new ColumnHeader("rlpTxn.RLP_LT_BYTESIZE", 4, length),
        new ColumnHeader("rlpTxn.RLP_LX_BYTESIZE", 4, length),
        new ColumnHeader("rlpTxn.TYPE", 2, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absTxNum = buffers.get(0);
    this.absTxNumInfiny = buffers.get(1);
    this.acc1 = buffers.get(2);
    this.acc2 = buffers.get(3);
    this.accBytesize = buffers.get(4);
    this.accessTupleBytesize = buffers.get(5);
    this.addrHi = buffers.get(6);
    this.addrLo = buffers.get(7);
    this.bit = buffers.get(8);
    this.bitAcc = buffers.get(9);
    this.byte1 = buffers.get(10);
    this.byte2 = buffers.get(11);
    this.codeFragmentIndex = buffers.get(12);
    this.counter = buffers.get(13);
    this.dataGasCost = buffers.get(14);
    this.dataHi = buffers.get(15);
    this.dataLo = buffers.get(16);
    this.depth1 = buffers.get(17);
    this.depth2 = buffers.get(18);
    this.done = buffers.get(19);
    this.indexData = buffers.get(20);
    this.indexLt = buffers.get(21);
    this.indexLx = buffers.get(22);
    this.input1 = buffers.get(23);
    this.input2 = buffers.get(24);
    this.isPrefix = buffers.get(25);
    this.lcCorrection = buffers.get(26);
    this.limb = buffers.get(27);
    this.limbConstructed = buffers.get(28);
    this.lt = buffers.get(29);
    this.lx = buffers.get(30);
    this.nAddr = buffers.get(31);
    this.nBytes = buffers.get(32);
    this.nKeys = buffers.get(33);
    this.nKeysPerAddr = buffers.get(34);
    this.nStep = buffers.get(35);
    this.phase1 = buffers.get(36);
    this.phase10 = buffers.get(37);
    this.phase11 = buffers.get(38);
    this.phase12 = buffers.get(39);
    this.phase13 = buffers.get(40);
    this.phase14 = buffers.get(41);
    this.phase15 = buffers.get(42);
    this.phase2 = buffers.get(43);
    this.phase3 = buffers.get(44);
    this.phase4 = buffers.get(45);
    this.phase5 = buffers.get(46);
    this.phase6 = buffers.get(47);
    this.phase7 = buffers.get(48);
    this.phase8 = buffers.get(49);
    this.phase9 = buffers.get(50);
    this.phaseEnd = buffers.get(51);
    this.phaseId = buffers.get(52);
    this.phaseSize = buffers.get(53);
    this.power = buffers.get(54);
    this.requiresEvmExecution = buffers.get(55);
    this.rlpLtBytesize = buffers.get(56);
    this.rlpLxBytesize = buffers.get(57);
    this.type = buffers.get(58);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absTxNum(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM already set");
    } else {
      filled.set(0);
    }

    absTxNum.putLong(b);

    return this;
  }

  public Trace absTxNumInfiny(final int b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM_INFINY already set");
    } else {
      filled.set(1);
    }

    absTxNumInfiny.putInt(b);

    return this;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlpTxn.ACC_1 already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc1.put((byte) 0);
    }
    acc1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlpTxn.ACC_2 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc2.put((byte) 0);
    }
    acc2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accBytesize(final short b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlpTxn.ACC_BYTESIZE already set");
    } else {
      filled.set(5);
    }

    accBytesize.putShort(b);

    return this;
  }

  public Trace accessTupleBytesize(final int b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlpTxn.ACCESS_TUPLE_BYTESIZE already set");
    } else {
      filled.set(2);
    }

    accessTupleBytesize.putInt(b);

    return this;
  }

  public Trace addrHi(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlpTxn.ADDR_HI already set");
    } else {
      filled.set(6);
    }

    addrHi.putLong(b);

    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlpTxn.ADDR_LO already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrLo.put((byte) 0);
    }
    addrLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlpTxn.BIT already set");
    } else {
      filled.set(8);
    }

    bit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlpTxn.BIT_ACC already set");
    } else {
      filled.set(9);
    }

    bitAcc.put(b.toByte());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlpTxn.BYTE_1 already set");
    } else {
      filled.set(10);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlpTxn.BYTE_2 already set");
    } else {
      filled.set(11);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlpTxn.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(12);
    }

    codeFragmentIndex.putLong(b);

    return this;
  }

  public Trace counter(final short b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlpTxn.COUNTER already set");
    } else {
      filled.set(13);
    }

    counter.putShort(b);

    return this;
  }

  public Trace dataGasCost(final long b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlpTxn.DATA_GAS_COST already set");
    } else {
      filled.set(14);
    }

    dataGasCost.putLong(b);

    return this;
  }

  public Trace dataHi(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlpTxn.DATA_HI already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      dataHi.put((byte) 0);
    }
    dataHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace dataLo(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlpTxn.DATA_LO already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      dataLo.put((byte) 0);
    }
    dataLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace depth1(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlpTxn.DEPTH_1 already set");
    } else {
      filled.set(17);
    }

    depth1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace depth2(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlpTxn.DEPTH_2 already set");
    } else {
      filled.set(18);
    }

    depth2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace done(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlpTxn.DONE already set");
    } else {
      filled.set(19);
    }

    done.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace indexData(final long b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlpTxn.INDEX_DATA already set");
    } else {
      filled.set(20);
    }

    indexData.putLong(b);

    return this;
  }

  public Trace indexLt(final long b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlpTxn.INDEX_LT already set");
    } else {
      filled.set(21);
    }

    indexLt.putLong(b);

    return this;
  }

  public Trace indexLx(final long b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlpTxn.INDEX_LX already set");
    } else {
      filled.set(22);
    }

    indexLx.putLong(b);

    return this;
  }

  public Trace input1(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlpTxn.INPUT_1 already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      input1.put((byte) 0);
    }
    input1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace input2(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlpTxn.INPUT_2 already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      input2.put((byte) 0);
    }
    input2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isPrefix(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("rlpTxn.IS_PREFIX already set");
    } else {
      filled.set(25);
    }

    isPrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lcCorrection(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("rlpTxn.LC_CORRECTION already set");
    } else {
      filled.set(26);
    }

    lcCorrection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("rlpTxn.LIMB already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace limbConstructed(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("rlpTxn.LIMB_CONSTRUCTED already set");
    } else {
      filled.set(28);
    }

    limbConstructed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lt(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("rlpTxn.LT already set");
    } else {
      filled.set(29);
    }

    lt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lx(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("rlpTxn.LX already set");
    } else {
      filled.set(30);
    }

    lx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace nAddr(final int b) {
    if (filled.get(54)) {
      throw new IllegalStateException("rlpTxn.nADDR already set");
    } else {
      filled.set(54);
    }

    nAddr.putInt(b);

    return this;
  }

  public Trace nBytes(final short b) {
    if (filled.get(55)) {
      throw new IllegalStateException("rlpTxn.nBYTES already set");
    } else {
      filled.set(55);
    }

    nBytes.putShort(b);

    return this;
  }

  public Trace nKeys(final int b) {
    if (filled.get(56)) {
      throw new IllegalStateException("rlpTxn.nKEYS already set");
    } else {
      filled.set(56);
    }

    nKeys.putInt(b);

    return this;
  }

  public Trace nKeysPerAddr(final int b) {
    if (filled.get(57)) {
      throw new IllegalStateException("rlpTxn.nKEYS_PER_ADDR already set");
    } else {
      filled.set(57);
    }

    nKeysPerAddr.putInt(b);

    return this;
  }

  public Trace nStep(final short b) {
    if (filled.get(58)) {
      throw new IllegalStateException("rlpTxn.nSTEP already set");
    } else {
      filled.set(58);
    }

    nStep.putShort(b);

    return this;
  }

  public Trace phase1(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("rlpTxn.PHASE_1 already set");
    } else {
      filled.set(31);
    }

    phase1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase10(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("rlpTxn.PHASE_10 already set");
    } else {
      filled.set(32);
    }

    phase10.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase11(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("rlpTxn.PHASE_11 already set");
    } else {
      filled.set(33);
    }

    phase11.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase12(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("rlpTxn.PHASE_12 already set");
    } else {
      filled.set(34);
    }

    phase12.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase13(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("rlpTxn.PHASE_13 already set");
    } else {
      filled.set(35);
    }

    phase13.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase14(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("rlpTxn.PHASE_14 already set");
    } else {
      filled.set(36);
    }

    phase14.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase15(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("rlpTxn.PHASE_15 already set");
    } else {
      filled.set(37);
    }

    phase15.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase2(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("rlpTxn.PHASE_2 already set");
    } else {
      filled.set(38);
    }

    phase2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase3(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("rlpTxn.PHASE_3 already set");
    } else {
      filled.set(39);
    }

    phase3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase4(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("rlpTxn.PHASE_4 already set");
    } else {
      filled.set(40);
    }

    phase4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase5(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("rlpTxn.PHASE_5 already set");
    } else {
      filled.set(41);
    }

    phase5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase6(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("rlpTxn.PHASE_6 already set");
    } else {
      filled.set(42);
    }

    phase6.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase7(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("rlpTxn.PHASE_7 already set");
    } else {
      filled.set(43);
    }

    phase7.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase8(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("rlpTxn.PHASE_8 already set");
    } else {
      filled.set(44);
    }

    phase8.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase9(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("rlpTxn.PHASE_9 already set");
    } else {
      filled.set(45);
    }

    phase9.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseEnd(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("rlpTxn.PHASE_END already set");
    } else {
      filled.set(46);
    }

    phaseEnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseId(final short b) {
    if (filled.get(47)) {
      throw new IllegalStateException("rlpTxn.PHASE_ID already set");
    } else {
      filled.set(47);
    }

    phaseId.putShort(b);

    return this;
  }

  public Trace phaseSize(final long b) {
    if (filled.get(48)) {
      throw new IllegalStateException("rlpTxn.PHASE_SIZE already set");
    } else {
      filled.set(48);
    }

    phaseSize.putLong(b);

    return this;
  }

  public Trace power(final Bytes b) {
    if (filled.get(49)) {
      throw new IllegalStateException("rlpTxn.POWER already set");
    } else {
      filled.set(49);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      power.put((byte) 0);
    }
    power.put(b.toArrayUnsafe());

    return this;
  }

  public Trace requiresEvmExecution(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("rlpTxn.REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(50);
    }

    requiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rlpLtBytesize(final int b) {
    if (filled.get(51)) {
      throw new IllegalStateException("rlpTxn.RLP_LT_BYTESIZE already set");
    } else {
      filled.set(51);
    }

    rlpLtBytesize.putInt(b);

    return this;
  }

  public Trace rlpLxBytesize(final int b) {
    if (filled.get(52)) {
      throw new IllegalStateException("rlpTxn.RLP_LX_BYTESIZE already set");
    } else {
      filled.set(52);
    }

    rlpLxBytesize.putInt(b);

    return this;
  }

  public Trace type(final short b) {
    if (filled.get(53)) {
      throw new IllegalStateException("rlpTxn.TYPE already set");
    } else {
      filled.set(53);
    }

    type.putShort(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM_INFINY has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlpTxn.ACC_1 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlpTxn.ACC_2 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlpTxn.ACC_BYTESIZE has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlpTxn.ACCESS_TUPLE_BYTESIZE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlpTxn.ADDR_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlpTxn.ADDR_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlpTxn.BIT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlpTxn.BIT_ACC has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlpTxn.BYTE_1 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlpTxn.BYTE_2 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlpTxn.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlpTxn.COUNTER has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlpTxn.DATA_GAS_COST has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlpTxn.DATA_HI has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlpTxn.DATA_LO has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlpTxn.DEPTH_1 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlpTxn.DEPTH_2 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlpTxn.DONE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlpTxn.INDEX_DATA has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlpTxn.INDEX_LT has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlpTxn.INDEX_LX has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlpTxn.INPUT_1 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlpTxn.INPUT_2 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("rlpTxn.IS_PREFIX has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("rlpTxn.LC_CORRECTION has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("rlpTxn.LIMB has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("rlpTxn.LIMB_CONSTRUCTED has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("rlpTxn.LT has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("rlpTxn.LX has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("rlpTxn.nADDR has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("rlpTxn.nBYTES has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("rlpTxn.nKEYS has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("rlpTxn.nKEYS_PER_ADDR has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("rlpTxn.nSTEP has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("rlpTxn.PHASE_1 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("rlpTxn.PHASE_10 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("rlpTxn.PHASE_11 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("rlpTxn.PHASE_12 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("rlpTxn.PHASE_13 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("rlpTxn.PHASE_14 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("rlpTxn.PHASE_15 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("rlpTxn.PHASE_2 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("rlpTxn.PHASE_3 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("rlpTxn.PHASE_4 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("rlpTxn.PHASE_5 has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("rlpTxn.PHASE_6 has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("rlpTxn.PHASE_7 has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("rlpTxn.PHASE_8 has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("rlpTxn.PHASE_9 has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("rlpTxn.PHASE_END has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("rlpTxn.PHASE_ID has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("rlpTxn.PHASE_SIZE has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("rlpTxn.POWER has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("rlpTxn.REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("rlpTxn.RLP_LT_BYTESIZE has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("rlpTxn.RLP_LX_BYTESIZE has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("rlpTxn.TYPE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absTxNum.position(absTxNum.position() + 8);
    }

    if (!filled.get(1)) {
      absTxNumInfiny.position(absTxNumInfiny.position() + 4);
    }

    if (!filled.get(3)) {
      acc1.position(acc1.position() + 32);
    }

    if (!filled.get(4)) {
      acc2.position(acc2.position() + 32);
    }

    if (!filled.get(5)) {
      accBytesize.position(accBytesize.position() + 2);
    }

    if (!filled.get(2)) {
      accessTupleBytesize.position(accessTupleBytesize.position() + 4);
    }

    if (!filled.get(6)) {
      addrHi.position(addrHi.position() + 8);
    }

    if (!filled.get(7)) {
      addrLo.position(addrLo.position() + 32);
    }

    if (!filled.get(8)) {
      bit.position(bit.position() + 1);
    }

    if (!filled.get(9)) {
      bitAcc.position(bitAcc.position() + 1);
    }

    if (!filled.get(10)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(11)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(12)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 8);
    }

    if (!filled.get(13)) {
      counter.position(counter.position() + 2);
    }

    if (!filled.get(14)) {
      dataGasCost.position(dataGasCost.position() + 8);
    }

    if (!filled.get(15)) {
      dataHi.position(dataHi.position() + 32);
    }

    if (!filled.get(16)) {
      dataLo.position(dataLo.position() + 32);
    }

    if (!filled.get(17)) {
      depth1.position(depth1.position() + 1);
    }

    if (!filled.get(18)) {
      depth2.position(depth2.position() + 1);
    }

    if (!filled.get(19)) {
      done.position(done.position() + 1);
    }

    if (!filled.get(20)) {
      indexData.position(indexData.position() + 8);
    }

    if (!filled.get(21)) {
      indexLt.position(indexLt.position() + 8);
    }

    if (!filled.get(22)) {
      indexLx.position(indexLx.position() + 8);
    }

    if (!filled.get(23)) {
      input1.position(input1.position() + 32);
    }

    if (!filled.get(24)) {
      input2.position(input2.position() + 32);
    }

    if (!filled.get(25)) {
      isPrefix.position(isPrefix.position() + 1);
    }

    if (!filled.get(26)) {
      lcCorrection.position(lcCorrection.position() + 1);
    }

    if (!filled.get(27)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(28)) {
      limbConstructed.position(limbConstructed.position() + 1);
    }

    if (!filled.get(29)) {
      lt.position(lt.position() + 1);
    }

    if (!filled.get(30)) {
      lx.position(lx.position() + 1);
    }

    if (!filled.get(54)) {
      nAddr.position(nAddr.position() + 4);
    }

    if (!filled.get(55)) {
      nBytes.position(nBytes.position() + 2);
    }

    if (!filled.get(56)) {
      nKeys.position(nKeys.position() + 4);
    }

    if (!filled.get(57)) {
      nKeysPerAddr.position(nKeysPerAddr.position() + 4);
    }

    if (!filled.get(58)) {
      nStep.position(nStep.position() + 2);
    }

    if (!filled.get(31)) {
      phase1.position(phase1.position() + 1);
    }

    if (!filled.get(32)) {
      phase10.position(phase10.position() + 1);
    }

    if (!filled.get(33)) {
      phase11.position(phase11.position() + 1);
    }

    if (!filled.get(34)) {
      phase12.position(phase12.position() + 1);
    }

    if (!filled.get(35)) {
      phase13.position(phase13.position() + 1);
    }

    if (!filled.get(36)) {
      phase14.position(phase14.position() + 1);
    }

    if (!filled.get(37)) {
      phase15.position(phase15.position() + 1);
    }

    if (!filled.get(38)) {
      phase2.position(phase2.position() + 1);
    }

    if (!filled.get(39)) {
      phase3.position(phase3.position() + 1);
    }

    if (!filled.get(40)) {
      phase4.position(phase4.position() + 1);
    }

    if (!filled.get(41)) {
      phase5.position(phase5.position() + 1);
    }

    if (!filled.get(42)) {
      phase6.position(phase6.position() + 1);
    }

    if (!filled.get(43)) {
      phase7.position(phase7.position() + 1);
    }

    if (!filled.get(44)) {
      phase8.position(phase8.position() + 1);
    }

    if (!filled.get(45)) {
      phase9.position(phase9.position() + 1);
    }

    if (!filled.get(46)) {
      phaseEnd.position(phaseEnd.position() + 1);
    }

    if (!filled.get(47)) {
      phaseId.position(phaseId.position() + 2);
    }

    if (!filled.get(48)) {
      phaseSize.position(phaseSize.position() + 8);
    }

    if (!filled.get(49)) {
      power.position(power.position() + 32);
    }

    if (!filled.get(50)) {
      requiresEvmExecution.position(requiresEvmExecution.position() + 1);
    }

    if (!filled.get(51)) {
      rlpLtBytesize.position(rlpLtBytesize.position() + 4);
    }

    if (!filled.get(52)) {
      rlpLxBytesize.position(rlpLxBytesize.position() + 4);
    }

    if (!filled.get(53)) {
      type.position(type.position() + 2);
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
