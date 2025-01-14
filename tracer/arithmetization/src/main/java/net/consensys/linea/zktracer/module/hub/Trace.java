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

package net.consensys.linea.zktracer.module.hub;

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
  public static final int DOM_SUB_STAMP_OFFSET___FINALIZATION = 0x9;
  public static final int DOM_SUB_STAMP_OFFSET___REVERT = 0x8;
  public static final int DOM_SUB_STAMP_OFFSET___SELFDESTRUCT = 0xa;
  public static final int MULTIPLIER___DOM_SUB_STAMPS = 0x10;
  public static final int MULTIPLIER___STACK_STAMP = 0x8;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absoluteTransactionNumber;
  private final MappedByteBuffer addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize;
  private final MappedByteBuffer addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo;
  private final MappedByteBuffer againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd;
  private final MappedByteBuffer againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment;
  private final MappedByteBuffer againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2;
  private final MappedByteBuffer alpha;
  private final MappedByteBuffer balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance;
  private final MappedByteBuffer balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo;
  private final MappedByteBuffer callDataOffsetXorMmuSize;
  private final MappedByteBuffer callDataSizeXorMmuSrcId;
  private final MappedByteBuffer callStackDepthXorStackItemHeight1;
  private final MappedByteBuffer callerContextNumber;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi;
  private final MappedByteBuffer codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue;
  private final MappedByteBuffer codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo;
  private final MappedByteBuffer codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo;
  private final MappedByteBuffer codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi;
  private final MappedByteBuffer codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize;
  private final MappedByteBuffer codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi;
  private final MappedByteBuffer contextGetsReverted;
  private final MappedByteBuffer contextMayChange;
  private final MappedByteBuffer contextNumber;
  private final MappedByteBuffer contextNumberNew;
  private final MappedByteBuffer contextNumberXorMmuTgtId;
  private final MappedByteBuffer contextRevertStamp;
  private final MappedByteBuffer contextSelfReverts;
  private final MappedByteBuffer contextWillRevert;
  private final MappedByteBuffer counterNsr;
  private final MappedByteBuffer counterTli;
  private final MappedByteBuffer delta;
  private final MappedByteBuffer deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock;
  private final MappedByteBuffer deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock;
  private final MappedByteBuffer deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase;
  private final MappedByteBuffer deploymentNumberNewXorCallerAddressHiXorMmuRefOffset;
  private final MappedByteBuffer deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi;
  private final MappedByteBuffer deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode;
  private final MappedByteBuffer deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn;
  private final MappedByteBuffer deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution;
  private final MappedByteBuffer domStamp;
  private final MappedByteBuffer exceptionAhoy;
  private final MappedByteBuffer existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf;
  private final MappedByteBuffer existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk;
  private final MappedByteBuffer finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn;
  private final MappedByteBuffer finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges;
  private final MappedByteBuffer finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig;
  private final MappedByteBuffer firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero;
  private final MappedByteBuffer firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr;
  private final MappedByteBuffer firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig;
  private final MappedByteBuffer gasActual;
  private final MappedByteBuffer gasCost;
  private final MappedByteBuffer gasExpected;
  private final MappedByteBuffer gasLimit;
  private final MappedByteBuffer gasNext;
  private final MappedByteBuffer gasPrice;
  private final MappedByteBuffer hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero;
  private final MappedByteBuffer hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero;
  private final MappedByteBuffer height;
  private final MappedByteBuffer heightNew;
  private final MappedByteBuffer hubStamp;
  private final MappedByteBuffer hubStampTransactionEnd;
  private final MappedByteBuffer instruction;
  private final MappedByteBuffer isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth;
  private final MappedByteBuffer logInfoStamp;
  private final MappedByteBuffer markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx;
  private final MappedByteBuffer markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew;
  private final MappedByteBuffer mmuStamp;
  private final MappedByteBuffer mxpOffset2Hi;
  private final MappedByteBuffer mxpOffset2Lo;
  private final MappedByteBuffer mxpSize1Hi;
  private final MappedByteBuffer mxpSize1Lo;
  private final MappedByteBuffer mxpSize2Hi;
  private final MappedByteBuffer mxpSize2Lo;
  private final MappedByteBuffer mxpStamp;
  private final MappedByteBuffer mxpWords;
  private final MappedByteBuffer nonStackRows;
  private final MappedByteBuffer nonce;
  private final MappedByteBuffer nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable;
  private final MappedByteBuffer nonceXorStpGasMxpXorPrcCalleeGasXorBasefee;
  private final MappedByteBuffer oobData1;
  private final MappedByteBuffer oobData2;
  private final MappedByteBuffer oobData3;
  private final MappedByteBuffer oobData4;
  private final MappedByteBuffer oobData5;
  private final MappedByteBuffer oobData6;
  private final MappedByteBuffer oobData7;
  private final MappedByteBuffer oobData8;
  private final MappedByteBuffer oobData9;
  private final MappedByteBuffer peekAtAccount;
  private final MappedByteBuffer peekAtContext;
  private final MappedByteBuffer peekAtMiscellaneous;
  private final MappedByteBuffer peekAtScenario;
  private final MappedByteBuffer peekAtStack;
  private final MappedByteBuffer peekAtStorage;
  private final MappedByteBuffer peekAtTransaction;
  private final MappedByteBuffer prcEcmulXorMachineStateFlag;
  private final MappedByteBuffer prcEcpairingXorMaxcsx;
  private final MappedByteBuffer prcEcrecoverXorModFlag;
  private final MappedByteBuffer prcFailureKnownToHubXorMulFlag;
  private final MappedByteBuffer prcFailureKnownToRamXorMxpx;
  private final MappedByteBuffer prcIdentityXorMxpFlag;
  private final MappedByteBuffer prcModexpXorOogx;
  private final MappedByteBuffer prcRipemd160XorOpcx;
  private final MappedByteBuffer prcSha2256XorPushpopFlag;
  private final MappedByteBuffer prcSuccessCallerWillRevertXorRdcx;
  private final MappedByteBuffer prcSuccessCallerWontRevertXorShfFlag;
  private final MappedByteBuffer priorityFeePerGas;
  private final MappedByteBuffer programCounter;
  private final MappedByteBuffer programCounterNew;
  private final MappedByteBuffer refundCounter;
  private final MappedByteBuffer refundCounterInfinity;
  private final MappedByteBuffer refundCounterNew;
  private final MappedByteBuffer refundEffective;
  private final MappedByteBuffer relativeBlockNumber;
  private final MappedByteBuffer returnAtCapacityXorMxpInst;
  private final MappedByteBuffer returnAtOffsetXorOobInst;
  private final MappedByteBuffer returnDataContextNumberXorStpGasStipend;
  private final MappedByteBuffer returnDataOffsetXorStpInstruction;
  private final MappedByteBuffer returnDataSize;
  private final MappedByteBuffer returnExceptionXorSox;
  private final MappedByteBuffer returnFromDeploymentEmptyCodeWillRevertXorSstorex;
  private final MappedByteBuffer returnFromDeploymentEmptyCodeWontRevertXorStackramFlag;
  private final MappedByteBuffer returnFromDeploymentNonemptyCodeWillRevertXorStackItemPop1;
  private final MappedByteBuffer returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop2;
  private final MappedByteBuffer returnFromMessageCallWillTouchRamXorStackItemPop3;
  private final MappedByteBuffer returnFromMessageCallWontTouchRamXorStackItemPop4;
  private final MappedByteBuffer rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize;
  private final MappedByteBuffer rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi;
  private final MappedByteBuffer rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired;
  private final MappedByteBuffer rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo;
  private final MappedByteBuffer rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2;
  private final MappedByteBuffer rlpaddrRecipe;
  private final MappedByteBuffer rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3;
  private final MappedByteBuffer rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4;
  private final MappedByteBuffer romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag;
  private final MappedByteBuffer selfdestructExceptionXorStaticx;
  private final MappedByteBuffer selfdestructWillRevertXorStaticFlag;
  private final MappedByteBuffer selfdestructWontRevertAlreadyMarkedXorStoFlag;
  private final MappedByteBuffer selfdestructWontRevertNotYetMarkedXorSux;
  private final MappedByteBuffer stackItemHeight2;
  private final MappedByteBuffer stackItemHeight3;
  private final MappedByteBuffer stackItemHeight4;
  private final MappedByteBuffer stackItemStamp1;
  private final MappedByteBuffer stackItemStamp2;
  private final MappedByteBuffer stackItemStamp3;
  private final MappedByteBuffer stackItemStamp4;
  private final MappedByteBuffer stpGasHi;
  private final MappedByteBuffer stpGasLo;
  private final MappedByteBuffer stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover;
  private final MappedByteBuffer stpValueHi;
  private final MappedByteBuffer stpValueLo;
  private final MappedByteBuffer subStamp;
  private final MappedByteBuffer swapFlag;
  private final MappedByteBuffer trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag;
  private final MappedByteBuffer trmRawAddressHiXorMxpOffset1Lo;
  private final MappedByteBuffer twoLineInstruction;
  private final MappedByteBuffer txExec;
  private final MappedByteBuffer txFinl;
  private final MappedByteBuffer txInit;
  private final MappedByteBuffer txSkip;
  private final MappedByteBuffer txWarm;
  private final MappedByteBuffer txnFlag;
  private final MappedByteBuffer warmthNewXorPrcEcaddXorLogInfoFlag;
  private final MappedByteBuffer warmthXorPrcBlake2FXorLogFlag;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("hub.ABSOLUTE_TRANSACTION_NUMBER", 2, length));
      headers.add(new ColumnHeader("hub.ADDRESS_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_PRC_CDO_xor_STATIC_GAS_xor_ADDRESS_HI_xor_CALL_DATA_SIZE", 4, length));
      headers.add(new ColumnHeader("hub.ADDRESS_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP_DATA_1_xor_HASH_INFO_KECCAK_HI_xor_ADDRESS_LO_xor_COINBASE_ADDRESS_LO", 16, length));
      headers.add(new ColumnHeader("hub.AGAIN_IN_BLK_xor_IS_ROOT_xor_CCSR_FLAG_xor_CALL_ABORT_WILL_REVERT_xor_ACC_FLAG_xor_AGAIN_IN_BLK_xor_COPY_TXCD", 1, length));
      headers.add(new ColumnHeader("hub.AGAIN_IN_CNF_xor_IS_STATIC_xor_EXP_FLAG_xor_CALL_ABORT_WONT_REVERT_xor_ADD_FLAG_xor_AGAIN_IN_CNF_xor_IS_DEPLOYMENT", 1, length));
      headers.add(new ColumnHeader("hub.AGAIN_IN_TXN_xor_UPDATE_xor_MMU_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_BIN_FLAG_xor_AGAIN_IN_TXN_xor_IS_TYPE2", 1, length));
      headers.add(new ColumnHeader("hub.ALPHA", 1, length));
      headers.add(new ColumnHeader("hub.BALANCE_NEW_xor_CALLER_ADDRESS_LO_xor_EXP_DATA_3_xor_PUSH_VALUE_HI_xor_STORAGE_KEY_LO_xor_INITIAL_BALANCE", 16, length));
      headers.add(new ColumnHeader("hub.BALANCE_xor_BYTE_CODE_ADDRESS_LO_xor_EXP_DATA_2_xor_HASH_INFO_KECCAK_LO_xor_STORAGE_KEY_HI_xor_FROM_ADDRESS_LO", 16, length));
      headers.add(new ColumnHeader("hub.CALL_DATA_OFFSET_xor_MMU_SIZE", 4, length));
      headers.add(new ColumnHeader("hub.CALL_DATA_SIZE_xor_MMU_SRC_ID", 4, length));
      headers.add(new ColumnHeader("hub.CALL_STACK_DEPTH_xor_STACK_ITEM_HEIGHT_1", 2, length));
      headers.add(new ColumnHeader("hub.CALLER_CONTEXT_NUMBER", 4, length));
      headers.add(new ColumnHeader("hub.CODE_FRAGMENT_INDEX", 4, length));
      headers.add(new ColumnHeader("hub.CODE_FRAGMENT_INDEX_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP_INST_xor_PRC_CDS_xor_DEPLOYMENT_NUMBER_xor_COINBASE_ADDRESS_HI", 4, length));
      headers.add(new ColumnHeader("hub.CODE_HASH_HI_NEW_xor_EXP_DATA_5_xor_STACK_ITEM_VALUE_HI_1_xor_VALUE_CURR_LO_xor_VALUE", 16, length));
      headers.add(new ColumnHeader("hub.CODE_HASH_HI_xor_CALL_VALUE_xor_EXP_DATA_4_xor_PUSH_VALUE_LO_xor_VALUE_CURR_HI_xor_TO_ADDRESS_LO", 16, length));
      headers.add(new ColumnHeader("hub.CODE_HASH_LO_NEW_xor_MMU_LIMB_2_xor_STACK_ITEM_VALUE_HI_3_xor_VALUE_NEXT_LO", 16, length));
      headers.add(new ColumnHeader("hub.CODE_HASH_LO_xor_MMU_LIMB_1_xor_STACK_ITEM_VALUE_HI_2_xor_VALUE_NEXT_HI", 16, length));
      headers.add(new ColumnHeader("hub.CODE_SIZE_NEW_xor_BYTE_CODE_CODE_FRAGMENT_INDEX_xor_MMU_EXO_SUM_xor_PRC_RAO_xor_INIT_CODE_SIZE", 4, length));
      headers.add(new ColumnHeader("hub.CODE_SIZE_xor_BYTE_CODE_ADDRESS_HI_xor_MMU_AUX_ID_xor_PRC_RAC_xor_DEPLOYMENT_NUMBER_INFTY_xor_FROM_ADDRESS_HI", 4, length));
      headers.add(new ColumnHeader("hub.CONTEXT_GETS_REVERTED", 1, length));
      headers.add(new ColumnHeader("hub.CONTEXT_MAY_CHANGE", 1, length));
      headers.add(new ColumnHeader("hub.CONTEXT_NUMBER", 4, length));
      headers.add(new ColumnHeader("hub.CONTEXT_NUMBER_NEW", 4, length));
      headers.add(new ColumnHeader("hub.CONTEXT_NUMBER_xor_MMU_TGT_ID", 4, length));
      headers.add(new ColumnHeader("hub.CONTEXT_REVERT_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.CONTEXT_SELF_REVERTS", 1, length));
      headers.add(new ColumnHeader("hub.CONTEXT_WILL_REVERT", 1, length));
      headers.add(new ColumnHeader("hub.COUNTER_NSR", 1, length));
      headers.add(new ColumnHeader("hub.COUNTER_TLI", 1, length));
      headers.add(new ColumnHeader("hub.DELTA", 1, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_NUMBER_FINAL_IN_BLOCK_xor_DEPLOYMENT_NUMBER_FINAL_IN_BLOCK", 2, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_NUMBER_FIRST_IN_BLOCK_xor_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK", 2, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_NUMBER_INFTY_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MMU_PHASE", 4, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_NUMBER_NEW_xor_CALLER_ADDRESS_HI_xor_MMU_REF_OFFSET", 4, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_NUMBER_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MMU_INST_xor_TO_ADDRESS_HI", 4, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_STATUS_INFTY_xor_MXP_DEPLOYS_xor_CALL_EXCEPTION_xor_CALL_FLAG_xor_FINAL_IN_CNF_xor_STATUS_CODE", 1, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_STATUS_NEW_xor_MXP_FLAG_xor_CALL_PRC_FAILURE_xor_CON_FLAG_xor_FINAL_IN_TXN", 1, length));
      headers.add(new ColumnHeader("hub.DEPLOYMENT_STATUS_xor_MMU_SUCCESS_BIT_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BTC_FLAG_xor_FINAL_IN_BLK_xor_REQUIRES_EVM_EXECUTION", 1, length));
      headers.add(new ColumnHeader("hub.DOM_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.EXCEPTION_AHOY", 1, length));
      headers.add(new ColumnHeader("hub.EXISTS_NEW_xor_MXP_MXPX_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_CREATE_FLAG_xor_FIRST_IN_CNF", 1, length));
      headers.add(new ColumnHeader("hub.EXISTS_xor_MXP_MTNTOP_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_COPY_FLAG_xor_FIRST_IN_BLK", 1, length));
      headers.add(new ColumnHeader("hub.FINAL_IN_BLK_xor_MXP_SIZE_1_NONZERO_NO_MXPX_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_DEC_FLAG_1_xor_FIRST_IN_TXN", 1, length));
      headers.add(new ColumnHeader("hub.FINAL_IN_CNF_xor_MXP_SIZE_2_NONZERO_NO_MXPX_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DEC_FLAG_2_xor_VALUE_CURR_CHANGES", 1, length));
      headers.add(new ColumnHeader("hub.FINAL_IN_TXN_xor_OOB_FLAG_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DEC_FLAG_3_xor_VALUE_CURR_IS_ORIG", 1, length));
      headers.add(new ColumnHeader("hub.FIRST_IN_BLK_xor_STP_EXISTS_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DEC_FLAG_4_xor_VALUE_CURR_IS_ZERO", 1, length));
      headers.add(new ColumnHeader("hub.FIRST_IN_CNF_xor_STP_FLAG_xor_CREATE_ABORT_xor_DUP_FLAG_xor_VALUE_NEXT_IS_CURR", 1, length));
      headers.add(new ColumnHeader("hub.FIRST_IN_TXN_xor_STP_OOGX_xor_CREATE_EMPTY_INIT_CODE_WILL_REVERT_xor_EXT_FLAG_xor_VALUE_NEXT_IS_ORIG", 1, length));
      headers.add(new ColumnHeader("hub.GAS_ACTUAL", 8, length));
      headers.add(new ColumnHeader("hub.GAS_COST", 8, length));
      headers.add(new ColumnHeader("hub.GAS_EXPECTED", 8, length));
      headers.add(new ColumnHeader("hub.GAS_LIMIT", 8, length));
      headers.add(new ColumnHeader("hub.GAS_NEXT", 8, length));
      headers.add(new ColumnHeader("hub.GAS_PRICE", 8, length));
      headers.add(new ColumnHeader("hub.HAS_CODE_NEW_xor_CREATE_EXCEPTION_xor_HASH_INFO_FLAG_xor_VALUE_ORIG_IS_ZERO", 1, length));
      headers.add(new ColumnHeader("hub.HAS_CODE_xor_STP_WARMTH_xor_CREATE_EMPTY_INIT_CODE_WONT_REVERT_xor_HALT_FLAG_xor_VALUE_NEXT_IS_ZERO", 1, length));
      headers.add(new ColumnHeader("hub.HEIGHT", 2, length));
      headers.add(new ColumnHeader("hub.HEIGHT_NEW", 2, length));
      headers.add(new ColumnHeader("hub.HUB_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.HUB_STAMP_TRANSACTION_END", 4, length));
      headers.add(new ColumnHeader("hub.INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("hub.IS_PRECOMPILE_xor_CREATE_FAILURE_CONDITION_WILL_REVERT_xor_ICPX_xor_WARMTH", 1, length));
      headers.add(new ColumnHeader("hub.LOG_INFO_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.MARKED_FOR_SELFDESTRUCT_NEW_xor_CREATE_NONEMPTY_INIT_CODE_FAILURE_WILL_REVERT_xor_JUMPX", 1, length));
      headers.add(new ColumnHeader("hub.MARKED_FOR_SELFDESTRUCT_xor_CREATE_FAILURE_CONDITION_WONT_REVERT_xor_INVALID_FLAG_xor_WARMTH_NEW", 1, length));
      headers.add(new ColumnHeader("hub.MMU_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.MXP_OFFSET_2_HI", 16, length));
      headers.add(new ColumnHeader("hub.MXP_OFFSET_2_LO", 16, length));
      headers.add(new ColumnHeader("hub.MXP_SIZE_1_HI", 16, length));
      headers.add(new ColumnHeader("hub.MXP_SIZE_1_LO", 16, length));
      headers.add(new ColumnHeader("hub.MXP_SIZE_2_HI", 16, length));
      headers.add(new ColumnHeader("hub.MXP_SIZE_2_LO", 16, length));
      headers.add(new ColumnHeader("hub.MXP_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.MXP_WORDS", 16, length));
      headers.add(new ColumnHeader("hub.NON_STACK_ROWS", 1, length));
      headers.add(new ColumnHeader("hub.NONCE", 8, length));
      headers.add(new ColumnHeader("hub.NONCE_NEW_xor_STP_GAS_PAID_OUT_OF_POCKET_xor_PRC_CALLER_GAS_xor_GAS_INITIALLY_AVAILABLE", 8, length));
      headers.add(new ColumnHeader("hub.NONCE_xor_STP_GAS_MXP_xor_PRC_CALLEE_GAS_xor_BASEFEE", 8, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_1", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_2", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_3", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_4", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_5", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_6", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_7", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_8", 16, length));
      headers.add(new ColumnHeader("hub.OOB_DATA_9", 16, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_ACCOUNT", 1, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_CONTEXT", 1, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_MISCELLANEOUS", 1, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_SCENARIO", 1, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_STACK", 1, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_STORAGE", 1, length));
      headers.add(new ColumnHeader("hub.PEEK_AT_TRANSACTION", 1, length));
      headers.add(new ColumnHeader("hub.PRC_ECMUL_xor_MACHINE_STATE_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.PRC_ECPAIRING_xor_MAXCSX", 1, length));
      headers.add(new ColumnHeader("hub.PRC_ECRECOVER_xor_MOD_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.PRC_FAILURE_KNOWN_TO_HUB_xor_MUL_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.PRC_FAILURE_KNOWN_TO_RAM_xor_MXPX", 1, length));
      headers.add(new ColumnHeader("hub.PRC_IDENTITY_xor_MXP_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.PRC_MODEXP_xor_OOGX", 1, length));
      headers.add(new ColumnHeader("hub.PRC_RIPEMD-160_xor_OPCX", 1, length));
      headers.add(new ColumnHeader("hub.PRC_SHA2-256_xor_PUSHPOP_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.PRC_SUCCESS_CALLER_WILL_REVERT_xor_RDCX", 1, length));
      headers.add(new ColumnHeader("hub.PRC_SUCCESS_CALLER_WONT_REVERT_xor_SHF_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.PRIORITY_FEE_PER_GAS", 8, length));
      headers.add(new ColumnHeader("hub.PROGRAM_COUNTER", 4, length));
      headers.add(new ColumnHeader("hub.PROGRAM_COUNTER_NEW", 4, length));
      headers.add(new ColumnHeader("hub.REFUND_COUNTER", 4, length));
      headers.add(new ColumnHeader("hub.REFUND_COUNTER_INFINITY", 8, length));
      headers.add(new ColumnHeader("hub.REFUND_COUNTER_NEW", 4, length));
      headers.add(new ColumnHeader("hub.REFUND_EFFECTIVE", 8, length));
      headers.add(new ColumnHeader("hub.RELATIVE_BLOCK_NUMBER", 2, length));
      headers.add(new ColumnHeader("hub.RETURN_AT_CAPACITY_xor_MXP_INST", 4, length));
      headers.add(new ColumnHeader("hub.RETURN_AT_OFFSET_xor_OOB_INST", 4, length));
      headers.add(new ColumnHeader("hub.RETURN_DATA_CONTEXT_NUMBER_xor_STP_GAS_STIPEND", 4, length));
      headers.add(new ColumnHeader("hub.RETURN_DATA_OFFSET_xor_STP_INSTRUCTION", 4, length));
      headers.add(new ColumnHeader("hub.RETURN_DATA_SIZE", 4, length));
      headers.add(new ColumnHeader("hub.RETURN_EXCEPTION_xor_SOX", 1, length));
      headers.add(new ColumnHeader("hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT_xor_SSTOREX", 1, length));
      headers.add(new ColumnHeader("hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT_xor_STACKRAM_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT_xor_STACK_ITEM_POP_1", 1, length));
      headers.add(new ColumnHeader("hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT_xor_STACK_ITEM_POP_2", 1, length));
      headers.add(new ColumnHeader("hub.RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM_xor_STACK_ITEM_POP_3", 1, length));
      headers.add(new ColumnHeader("hub.RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM_xor_STACK_ITEM_POP_4", 1, length));
      headers.add(new ColumnHeader("hub.RLPADDR_DEP_ADDR_HI_xor_CALL_DATA_CONTEXT_NUMBER_xor_MMU_REF_SIZE", 4, length));
      headers.add(new ColumnHeader("hub.RLPADDR_DEP_ADDR_LO_xor_MMU_SRC_OFFSET_HI_xor_STACK_ITEM_VALUE_HI_4_xor_VALUE_ORIG_HI", 16, length));
      headers.add(new ColumnHeader("hub.RLPADDR_FLAG_xor_CREATE_NONEMPTY_INIT_CODE_FAILURE_WONT_REVERT_xor_JUMP_DESTINATION_VETTING_REQUIRED", 1, length));
      headers.add(new ColumnHeader("hub.RLPADDR_KEC_HI_xor_MMU_SRC_OFFSET_LO_xor_STACK_ITEM_VALUE_LO_1_xor_VALUE_ORIG_LO", 16, length));
      headers.add(new ColumnHeader("hub.RLPADDR_KEC_LO_xor_MMU_TGT_OFFSET_LO_xor_STACK_ITEM_VALUE_LO_2", 16, length));
      headers.add(new ColumnHeader("hub.RLPADDR_RECIPE", 1, length));
      headers.add(new ColumnHeader("hub.RLPADDR_SALT_HI_xor_MXP_GAS_MXP_xor_STACK_ITEM_VALUE_LO_3", 16, length));
      headers.add(new ColumnHeader("hub.RLPADDR_SALT_LO_xor_MXP_OFFSET_1_HI_xor_STACK_ITEM_VALUE_LO_4", 16, length));
      headers.add(new ColumnHeader("hub.ROMLEX_FLAG_xor_CREATE_NONEMPTY_INIT_CODE_SUCCESS_WILL_REVERT_xor_JUMP_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.SELFDESTRUCT_EXCEPTION_xor_STATICX", 1, length));
      headers.add(new ColumnHeader("hub.SELFDESTRUCT_WILL_REVERT_xor_STATIC_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED_xor_STO_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED_xor_SUX", 1, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_HEIGHT_2", 2, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_HEIGHT_3", 2, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_HEIGHT_4", 2, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_STAMP_1", 5, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_STAMP_2", 5, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_STAMP_3", 5, length));
      headers.add(new ColumnHeader("hub.STACK_ITEM_STAMP_4", 5, length));
      headers.add(new ColumnHeader("hub.STP_GAS_HI", 16, length));
      headers.add(new ColumnHeader("hub.STP_GAS_LO", 16, length));
      headers.add(new ColumnHeader("hub.STP_GAS_UPFRONT_GAS_COST_xor_PRC_RETURN_GAS_xor_GAS_LEFTOVER", 8, length));
      headers.add(new ColumnHeader("hub.STP_VALUE_HI", 16, length));
      headers.add(new ColumnHeader("hub.STP_VALUE_LO", 16, length));
      headers.add(new ColumnHeader("hub.SUB_STAMP", 4, length));
      headers.add(new ColumnHeader("hub.SWAP_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.TRM_FLAG_xor_CREATE_NONEMPTY_INIT_CODE_SUCCESS_WONT_REVERT_xor_KEC_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.TRM_RAW_ADDRESS_HI_xor_MXP_OFFSET_1_LO", 16, length));
      headers.add(new ColumnHeader("hub.TWO_LINE_INSTRUCTION", 1, length));
      headers.add(new ColumnHeader("hub.TX_EXEC", 1, length));
      headers.add(new ColumnHeader("hub.TX_FINL", 1, length));
      headers.add(new ColumnHeader("hub.TX_INIT", 1, length));
      headers.add(new ColumnHeader("hub.TX_SKIP", 1, length));
      headers.add(new ColumnHeader("hub.TX_WARM", 1, length));
      headers.add(new ColumnHeader("hub.TXN_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.WARMTH_NEW_xor_PRC_ECADD_xor_LOG_INFO_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.WARMTH_xor_PRC_BLAKE2f_xor_LOG_FLAG", 1, length));
      headers.add(new ColumnHeader("hub.WCP_FLAG", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.absoluteTransactionNumber = buffers.get(0);
    this.addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize = buffers.get(1);
    this.addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo = buffers.get(2);
    this.againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd = buffers.get(3);
    this.againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment = buffers.get(4);
    this.againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2 = buffers.get(5);
    this.alpha = buffers.get(6);
    this.balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance = buffers.get(7);
    this.balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo = buffers.get(8);
    this.callDataOffsetXorMmuSize = buffers.get(9);
    this.callDataSizeXorMmuSrcId = buffers.get(10);
    this.callStackDepthXorStackItemHeight1 = buffers.get(11);
    this.callerContextNumber = buffers.get(12);
    this.codeFragmentIndex = buffers.get(13);
    this.codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi = buffers.get(14);
    this.codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue = buffers.get(15);
    this.codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo = buffers.get(16);
    this.codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo = buffers.get(17);
    this.codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi = buffers.get(18);
    this.codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize = buffers.get(19);
    this.codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi = buffers.get(20);
    this.contextGetsReverted = buffers.get(21);
    this.contextMayChange = buffers.get(22);
    this.contextNumber = buffers.get(23);
    this.contextNumberNew = buffers.get(24);
    this.contextNumberXorMmuTgtId = buffers.get(25);
    this.contextRevertStamp = buffers.get(26);
    this.contextSelfReverts = buffers.get(27);
    this.contextWillRevert = buffers.get(28);
    this.counterNsr = buffers.get(29);
    this.counterTli = buffers.get(30);
    this.delta = buffers.get(31);
    this.deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock = buffers.get(32);
    this.deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock = buffers.get(33);
    this.deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase = buffers.get(34);
    this.deploymentNumberNewXorCallerAddressHiXorMmuRefOffset = buffers.get(35);
    this.deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi = buffers.get(36);
    this.deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode = buffers.get(37);
    this.deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn = buffers.get(38);
    this.deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution = buffers.get(39);
    this.domStamp = buffers.get(40);
    this.exceptionAhoy = buffers.get(41);
    this.existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf = buffers.get(42);
    this.existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk = buffers.get(43);
    this.finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn = buffers.get(44);
    this.finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges = buffers.get(45);
    this.finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig = buffers.get(46);
    this.firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero = buffers.get(47);
    this.firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr = buffers.get(48);
    this.firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig = buffers.get(49);
    this.gasActual = buffers.get(50);
    this.gasCost = buffers.get(51);
    this.gasExpected = buffers.get(52);
    this.gasLimit = buffers.get(53);
    this.gasNext = buffers.get(54);
    this.gasPrice = buffers.get(55);
    this.hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero = buffers.get(56);
    this.hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero = buffers.get(57);
    this.height = buffers.get(58);
    this.heightNew = buffers.get(59);
    this.hubStamp = buffers.get(60);
    this.hubStampTransactionEnd = buffers.get(61);
    this.instruction = buffers.get(62);
    this.isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth = buffers.get(63);
    this.logInfoStamp = buffers.get(64);
    this.markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx = buffers.get(65);
    this.markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew = buffers.get(66);
    this.mmuStamp = buffers.get(67);
    this.mxpOffset2Hi = buffers.get(68);
    this.mxpOffset2Lo = buffers.get(69);
    this.mxpSize1Hi = buffers.get(70);
    this.mxpSize1Lo = buffers.get(71);
    this.mxpSize2Hi = buffers.get(72);
    this.mxpSize2Lo = buffers.get(73);
    this.mxpStamp = buffers.get(74);
    this.mxpWords = buffers.get(75);
    this.nonStackRows = buffers.get(76);
    this.nonce = buffers.get(77);
    this.nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable = buffers.get(78);
    this.nonceXorStpGasMxpXorPrcCalleeGasXorBasefee = buffers.get(79);
    this.oobData1 = buffers.get(80);
    this.oobData2 = buffers.get(81);
    this.oobData3 = buffers.get(82);
    this.oobData4 = buffers.get(83);
    this.oobData5 = buffers.get(84);
    this.oobData6 = buffers.get(85);
    this.oobData7 = buffers.get(86);
    this.oobData8 = buffers.get(87);
    this.oobData9 = buffers.get(88);
    this.peekAtAccount = buffers.get(89);
    this.peekAtContext = buffers.get(90);
    this.peekAtMiscellaneous = buffers.get(91);
    this.peekAtScenario = buffers.get(92);
    this.peekAtStack = buffers.get(93);
    this.peekAtStorage = buffers.get(94);
    this.peekAtTransaction = buffers.get(95);
    this.prcEcmulXorMachineStateFlag = buffers.get(96);
    this.prcEcpairingXorMaxcsx = buffers.get(97);
    this.prcEcrecoverXorModFlag = buffers.get(98);
    this.prcFailureKnownToHubXorMulFlag = buffers.get(99);
    this.prcFailureKnownToRamXorMxpx = buffers.get(100);
    this.prcIdentityXorMxpFlag = buffers.get(101);
    this.prcModexpXorOogx = buffers.get(102);
    this.prcRipemd160XorOpcx = buffers.get(103);
    this.prcSha2256XorPushpopFlag = buffers.get(104);
    this.prcSuccessCallerWillRevertXorRdcx = buffers.get(105);
    this.prcSuccessCallerWontRevertXorShfFlag = buffers.get(106);
    this.priorityFeePerGas = buffers.get(107);
    this.programCounter = buffers.get(108);
    this.programCounterNew = buffers.get(109);
    this.refundCounter = buffers.get(110);
    this.refundCounterInfinity = buffers.get(111);
    this.refundCounterNew = buffers.get(112);
    this.refundEffective = buffers.get(113);
    this.relativeBlockNumber = buffers.get(114);
    this.returnAtCapacityXorMxpInst = buffers.get(115);
    this.returnAtOffsetXorOobInst = buffers.get(116);
    this.returnDataContextNumberXorStpGasStipend = buffers.get(117);
    this.returnDataOffsetXorStpInstruction = buffers.get(118);
    this.returnDataSize = buffers.get(119);
    this.returnExceptionXorSox = buffers.get(120);
    this.returnFromDeploymentEmptyCodeWillRevertXorSstorex = buffers.get(121);
    this.returnFromDeploymentEmptyCodeWontRevertXorStackramFlag = buffers.get(122);
    this.returnFromDeploymentNonemptyCodeWillRevertXorStackItemPop1 = buffers.get(123);
    this.returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop2 = buffers.get(124);
    this.returnFromMessageCallWillTouchRamXorStackItemPop3 = buffers.get(125);
    this.returnFromMessageCallWontTouchRamXorStackItemPop4 = buffers.get(126);
    this.rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize = buffers.get(127);
    this.rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi = buffers.get(128);
    this.rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired = buffers.get(129);
    this.rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo = buffers.get(130);
    this.rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2 = buffers.get(131);
    this.rlpaddrRecipe = buffers.get(132);
    this.rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3 = buffers.get(133);
    this.rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4 = buffers.get(134);
    this.romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag = buffers.get(135);
    this.selfdestructExceptionXorStaticx = buffers.get(136);
    this.selfdestructWillRevertXorStaticFlag = buffers.get(137);
    this.selfdestructWontRevertAlreadyMarkedXorStoFlag = buffers.get(138);
    this.selfdestructWontRevertNotYetMarkedXorSux = buffers.get(139);
    this.stackItemHeight2 = buffers.get(140);
    this.stackItemHeight3 = buffers.get(141);
    this.stackItemHeight4 = buffers.get(142);
    this.stackItemStamp1 = buffers.get(143);
    this.stackItemStamp2 = buffers.get(144);
    this.stackItemStamp3 = buffers.get(145);
    this.stackItemStamp4 = buffers.get(146);
    this.stpGasHi = buffers.get(147);
    this.stpGasLo = buffers.get(148);
    this.stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover = buffers.get(149);
    this.stpValueHi = buffers.get(150);
    this.stpValueLo = buffers.get(151);
    this.subStamp = buffers.get(152);
    this.swapFlag = buffers.get(153);
    this.trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag = buffers.get(154);
    this.trmRawAddressHiXorMxpOffset1Lo = buffers.get(155);
    this.twoLineInstruction = buffers.get(156);
    this.txExec = buffers.get(157);
    this.txFinl = buffers.get(158);
    this.txInit = buffers.get(159);
    this.txSkip = buffers.get(160);
    this.txWarm = buffers.get(161);
    this.txnFlag = buffers.get(162);
    this.warmthNewXorPrcEcaddXorLogInfoFlag = buffers.get(163);
    this.warmthXorPrcBlake2FXorLogFlag = buffers.get(164);
    this.wcpFlag = buffers.get(165);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absoluteTransactionNumber(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("hub.ABSOLUTE_TRANSACTION_NUMBER already set");
    } else {
      filled.set(0);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("hub.ABSOLUTE_TRANSACTION_NUMBER has invalid value (" + b + ")"); }
    absoluteTransactionNumber.put((byte) (b >> 8));
    absoluteTransactionNumber.put((byte) b);


    return this;
  }

  public Trace callerContextNumber(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("hub.CALLER_CONTEXT_NUMBER already set");
    } else {
      filled.set(1);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.CALLER_CONTEXT_NUMBER has invalid value (" + b + ")"); }
    callerContextNumber.put((byte) (b >> 24));
    callerContextNumber.put((byte) (b >> 16));
    callerContextNumber.put((byte) (b >> 8));
    callerContextNumber.put((byte) b);


    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("hub.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(2);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeFragmentIndex.put((byte) (b >> 24));
    codeFragmentIndex.put((byte) (b >> 16));
    codeFragmentIndex.put((byte) (b >> 8));
    codeFragmentIndex.put((byte) b);


    return this;
  }

  public Trace contextGetsReverted(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("hub.CONTEXT_GETS_REVERTED already set");
    } else {
      filled.set(3);
    }

    contextGetsReverted.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextMayChange(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("hub.CONTEXT_MAY_CHANGE already set");
    } else {
      filled.set(4);
    }

    contextMayChange.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextNumber(final long b) {
    if (filled.get(5)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER already set");
    } else {
      filled.set(5);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.CONTEXT_NUMBER has invalid value (" + b + ")"); }
    contextNumber.put((byte) (b >> 24));
    contextNumber.put((byte) (b >> 16));
    contextNumber.put((byte) (b >> 8));
    contextNumber.put((byte) b);


    return this;
  }

  public Trace contextNumberNew(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER_NEW already set");
    } else {
      filled.set(6);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.CONTEXT_NUMBER_NEW has invalid value (" + b + ")"); }
    contextNumberNew.put((byte) (b >> 24));
    contextNumberNew.put((byte) (b >> 16));
    contextNumberNew.put((byte) (b >> 8));
    contextNumberNew.put((byte) b);


    return this;
  }

  public Trace contextRevertStamp(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("hub.CONTEXT_REVERT_STAMP already set");
    } else {
      filled.set(7);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.CONTEXT_REVERT_STAMP has invalid value (" + b + ")"); }
    contextRevertStamp.put((byte) (b >> 24));
    contextRevertStamp.put((byte) (b >> 16));
    contextRevertStamp.put((byte) (b >> 8));
    contextRevertStamp.put((byte) b);


    return this;
  }

  public Trace contextSelfReverts(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("hub.CONTEXT_SELF_REVERTS already set");
    } else {
      filled.set(8);
    }

    contextSelfReverts.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextWillRevert(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("hub.CONTEXT_WILL_REVERT already set");
    } else {
      filled.set(9);
    }

    contextWillRevert.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace counterNsr(final long b) {
    if (filled.get(10)) {
      throw new IllegalStateException("hub.COUNTER_NSR already set");
    } else {
      filled.set(10);
    }

    if(b >= 256L) { throw new IllegalArgumentException("hub.COUNTER_NSR has invalid value (" + b + ")"); }
    counterNsr.put((byte) b);


    return this;
  }

  public Trace counterTli(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("hub.COUNTER_TLI already set");
    } else {
      filled.set(11);
    }

    counterTli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace domStamp(final long b) {
    if (filled.get(12)) {
      throw new IllegalStateException("hub.DOM_STAMP already set");
    } else {
      filled.set(12);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.DOM_STAMP has invalid value (" + b + ")"); }
    domStamp.put((byte) (b >> 24));
    domStamp.put((byte) (b >> 16));
    domStamp.put((byte) (b >> 8));
    domStamp.put((byte) b);


    return this;
  }

  public Trace exceptionAhoy(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("hub.EXCEPTION_AHOY already set");
    } else {
      filled.set(13);
    }

    exceptionAhoy.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace gasActual(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("hub.GAS_ACTUAL already set");
    } else {
      filled.set(14);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.GAS_ACTUAL has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasActual.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasActual.put(bs.get(j)); }

    return this;
  }

  public Trace gasCost(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("hub.GAS_COST already set");
    } else {
      filled.set(15);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.GAS_COST has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasCost.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasCost.put(bs.get(j)); }

    return this;
  }

  public Trace gasExpected(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("hub.GAS_EXPECTED already set");
    } else {
      filled.set(16);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.GAS_EXPECTED has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasExpected.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasExpected.put(bs.get(j)); }

    return this;
  }

  public Trace gasNext(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("hub.GAS_NEXT already set");
    } else {
      filled.set(17);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.GAS_NEXT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasNext.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasNext.put(bs.get(j)); }

    return this;
  }

  public Trace height(final long b) {
    if (filled.get(18)) {
      throw new IllegalStateException("hub.HEIGHT already set");
    } else {
      filled.set(18);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.HEIGHT has invalid value (" + b + ")"); }
    height.put((byte) (b >> 8));
    height.put((byte) b);


    return this;
  }

  public Trace heightNew(final long b) {
    if (filled.get(19)) {
      throw new IllegalStateException("hub.HEIGHT_NEW already set");
    } else {
      filled.set(19);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.HEIGHT_NEW has invalid value (" + b + ")"); }
    heightNew.put((byte) (b >> 8));
    heightNew.put((byte) b);


    return this;
  }

  public Trace hubStamp(final long b) {
    if (filled.get(20)) {
      throw new IllegalStateException("hub.HUB_STAMP already set");
    } else {
      filled.set(20);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.HUB_STAMP has invalid value (" + b + ")"); }
    hubStamp.put((byte) (b >> 24));
    hubStamp.put((byte) (b >> 16));
    hubStamp.put((byte) (b >> 8));
    hubStamp.put((byte) b);


    return this;
  }

  public Trace hubStampTransactionEnd(final long b) {
    if (filled.get(21)) {
      throw new IllegalStateException("hub.HUB_STAMP_TRANSACTION_END already set");
    } else {
      filled.set(21);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.HUB_STAMP_TRANSACTION_END has invalid value (" + b + ")"); }
    hubStampTransactionEnd.put((byte) (b >> 24));
    hubStampTransactionEnd.put((byte) (b >> 16));
    hubStampTransactionEnd.put((byte) (b >> 8));
    hubStampTransactionEnd.put((byte) b);


    return this;
  }

  public Trace logInfoStamp(final long b) {
    if (filled.get(22)) {
      throw new IllegalStateException("hub.LOG_INFO_STAMP already set");
    } else {
      filled.set(22);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.LOG_INFO_STAMP has invalid value (" + b + ")"); }
    logInfoStamp.put((byte) (b >> 24));
    logInfoStamp.put((byte) (b >> 16));
    logInfoStamp.put((byte) (b >> 8));
    logInfoStamp.put((byte) b);


    return this;
  }

  public Trace mmuStamp(final long b) {
    if (filled.get(23)) {
      throw new IllegalStateException("hub.MMU_STAMP already set");
    } else {
      filled.set(23);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.MMU_STAMP has invalid value (" + b + ")"); }
    mmuStamp.put((byte) (b >> 24));
    mmuStamp.put((byte) (b >> 16));
    mmuStamp.put((byte) (b >> 8));
    mmuStamp.put((byte) b);


    return this;
  }

  public Trace mxpStamp(final long b) {
    if (filled.get(24)) {
      throw new IllegalStateException("hub.MXP_STAMP already set");
    } else {
      filled.set(24);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.MXP_STAMP has invalid value (" + b + ")"); }
    mxpStamp.put((byte) (b >> 24));
    mxpStamp.put((byte) (b >> 16));
    mxpStamp.put((byte) (b >> 8));
    mxpStamp.put((byte) b);


    return this;
  }

  public Trace nonStackRows(final long b) {
    if (filled.get(25)) {
      throw new IllegalStateException("hub.NON_STACK_ROWS already set");
    } else {
      filled.set(25);
    }

    if(b >= 256L) { throw new IllegalArgumentException("hub.NON_STACK_ROWS has invalid value (" + b + ")"); }
    nonStackRows.put((byte) b);


    return this;
  }

  public Trace pAccountAddressHi(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.account/ADDRESS_HI already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/ADDRESS_HI has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pAccountAddressLo(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.account/ADDRESS_LO already set");
    } else {
      filled.set(133);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountAgainInBlk(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.account/AGAIN_IN_BLK already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountAgainInCnf(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.account/AGAIN_IN_CNF already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountAgainInTxn(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.account/AGAIN_IN_TXN already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountBalance(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.account/BALANCE already set");
    } else {
      filled.set(134);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/BALANCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountBalanceNew(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.account/BALANCE_NEW already set");
    } else {
      filled.set(135);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/BALANCE_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountCodeFragmentIndex(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.account/CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(105);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 24));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 16));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 8));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) b);


    return this;
  }

  public Trace pAccountCodeHashHi(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.account/CODE_HASH_HI already set");
    } else {
      filled.set(136);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/CODE_HASH_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountCodeHashHiNew(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.account/CODE_HASH_HI_NEW already set");
    } else {
      filled.set(137);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/CODE_HASH_HI_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountCodeHashLo(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.account/CODE_HASH_LO already set");
    } else {
      filled.set(138);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/CODE_HASH_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountCodeHashLoNew(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.account/CODE_HASH_LO_NEW already set");
    } else {
      filled.set(139);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/CODE_HASH_LO_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountCodeSize(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.account/CODE_SIZE already set");
    } else {
      filled.set(106);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/CODE_SIZE has invalid value (" + b + ")"); }
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 24));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 16));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 8));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) b);


    return this;
  }

  public Trace pAccountCodeSizeNew(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.account/CODE_SIZE_NEW already set");
    } else {
      filled.set(107);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/CODE_SIZE_NEW has invalid value (" + b + ")"); }
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 24));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 16));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 8));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) b);


    return this;
  }

  public Trace pAccountDeploymentNumber(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(108);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/DEPLOYMENT_NUMBER has invalid value (" + b + ")"); }
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 24));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 16));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 8));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) b);


    return this;
  }

  public Trace pAccountDeploymentNumberFinalInBlock(final long b) {
    if (filled.get(102)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER_FINAL_IN_BLOCK already set");
    } else {
      filled.set(102);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("hub.account/DEPLOYMENT_NUMBER_FINAL_IN_BLOCK has invalid value (" + b + ")"); }
    deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock.put((byte) (b >> 8));
    deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock.put((byte) b);


    return this;
  }

  public Trace pAccountDeploymentNumberFirstInBlock(final long b) {
    if (filled.get(103)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER_FIRST_IN_BLOCK already set");
    } else {
      filled.set(103);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("hub.account/DEPLOYMENT_NUMBER_FIRST_IN_BLOCK has invalid value (" + b + ")"); }
    deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock.put((byte) (b >> 8));
    deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock.put((byte) b);


    return this;
  }

  public Trace pAccountDeploymentNumberInfty(final long b) {
    if (filled.get(109)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER_INFTY already set");
    } else {
      filled.set(109);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/DEPLOYMENT_NUMBER_INFTY has invalid value (" + b + ")"); }
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 24));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 16));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 8));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) b);


    return this;
  }

  public Trace pAccountDeploymentNumberNew(final long b) {
    if (filled.get(110)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER_NEW already set");
    } else {
      filled.set(110);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/DEPLOYMENT_NUMBER_NEW has invalid value (" + b + ")"); }
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 24));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 16));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 8));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) b);


    return this;
  }

  public Trace pAccountDeploymentStatus(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_STATUS already set");
    } else {
      filled.set(48);
    }

    deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountDeploymentStatusInfty(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_STATUS_INFTY already set");
    } else {
      filled.set(49);
    }

    deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountDeploymentStatusNew(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_STATUS_NEW already set");
    } else {
      filled.set(50);
    }

    deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountExists(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("hub.account/EXISTS already set");
    } else {
      filled.set(51);
    }

    existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountExistsNew(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("hub.account/EXISTS_NEW already set");
    } else {
      filled.set(52);
    }

    existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountFinalInBlk(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.account/FINAL_IN_BLK already set");
    } else {
      filled.set(53);
    }

    finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountFinalInCnf(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.account/FINAL_IN_CNF already set");
    } else {
      filled.set(54);
    }

    finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountFinalInTxn(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.account/FINAL_IN_TXN already set");
    } else {
      filled.set(55);
    }

    finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountFirstInBlk(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.account/FIRST_IN_BLK already set");
    } else {
      filled.set(56);
    }

    firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountFirstInCnf(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.account/FIRST_IN_CNF already set");
    } else {
      filled.set(57);
    }

    firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountFirstInTxn(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.account/FIRST_IN_TXN already set");
    } else {
      filled.set(58);
    }

    firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountHasCode(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.account/HAS_CODE already set");
    } else {
      filled.set(59);
    }

    hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountHasCodeNew(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.account/HAS_CODE_NEW already set");
    } else {
      filled.set(60);
    }

    hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountIsPrecompile(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.account/IS_PRECOMPILE already set");
    } else {
      filled.set(61);
    }

    isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountMarkedForSelfdestruct(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.account/MARKED_FOR_SELFDESTRUCT already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountMarkedForSelfdestructNew(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("hub.account/MARKED_FOR_SELFDESTRUCT_NEW already set");
    } else {
      filled.set(63);
    }

    markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountNonce(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.account/NONCE already set");
    } else {
      filled.set(124);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.account/NONCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountNonceNew(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.account/NONCE_NEW already set");
    } else {
      filled.set(125);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.account/NONCE_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountRlpaddrDepAddrHi(final long b) {
    if (filled.get(111)) {
      throw new IllegalStateException("hub.account/RLPADDR_DEP_ADDR_HI already set");
    } else {
      filled.set(111);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.account/RLPADDR_DEP_ADDR_HI has invalid value (" + b + ")"); }
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 24));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 16));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 8));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) b);


    return this;
  }

  public Trace pAccountRlpaddrDepAddrLo(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.account/RLPADDR_DEP_ADDR_LO already set");
    } else {
      filled.set(140);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/RLPADDR_DEP_ADDR_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountRlpaddrFlag(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.account/RLPADDR_FLAG already set");
    } else {
      filled.set(64);
    }

    rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountRlpaddrKecHi(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.account/RLPADDR_KEC_HI already set");
    } else {
      filled.set(141);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/RLPADDR_KEC_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountRlpaddrKecLo(final Bytes b) {
    if (filled.get(142)) {
      throw new IllegalStateException("hub.account/RLPADDR_KEC_LO already set");
    } else {
      filled.set(142);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/RLPADDR_KEC_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountRlpaddrRecipe(final long b) {
    if (filled.get(97)) {
      throw new IllegalStateException("hub.account/RLPADDR_RECIPE already set");
    } else {
      filled.set(97);
    }

    if(b >= 256L) { throw new IllegalArgumentException("hub.account/RLPADDR_RECIPE has invalid value (" + b + ")"); }
    rlpaddrRecipe.put((byte) b);


    return this;
  }

  public Trace pAccountRlpaddrSaltHi(final Bytes b) {
    if (filled.get(143)) {
      throw new IllegalStateException("hub.account/RLPADDR_SALT_HI already set");
    } else {
      filled.set(143);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/RLPADDR_SALT_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountRlpaddrSaltLo(final Bytes b) {
    if (filled.get(144)) {
      throw new IllegalStateException("hub.account/RLPADDR_SALT_LO already set");
    } else {
      filled.set(144);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/RLPADDR_SALT_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountRomlexFlag(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException("hub.account/ROMLEX_FLAG already set");
    } else {
      filled.set(65);
    }

    romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountTrmFlag(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException("hub.account/TRM_FLAG already set");
    } else {
      filled.set(66);
    }

    trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountTrmRawAddressHi(final Bytes b) {
    if (filled.get(145)) {
      throw new IllegalStateException("hub.account/TRM_RAW_ADDRESS_HI already set");
    } else {
      filled.set(145);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.account/TRM_RAW_ADDRESS_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { trmRawAddressHiXorMxpOffset1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { trmRawAddressHiXorMxpOffset1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pAccountWarmth(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("hub.account/WARMTH already set");
    } else {
      filled.set(67);
    }

    warmthXorPrcBlake2FXorLogFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountWarmthNew(final Boolean b) {
    if (filled.get(68)) {
      throw new IllegalStateException("hub.account/WARMTH_NEW already set");
    } else {
      filled.set(68);
    }

    warmthNewXorPrcEcaddXorLogInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pContextAccountAddressHi(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.context/ACCOUNT_ADDRESS_HI already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/ACCOUNT_ADDRESS_HI has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pContextAccountAddressLo(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.context/ACCOUNT_ADDRESS_LO already set");
    } else {
      filled.set(133);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.context/ACCOUNT_ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pContextAccountDeploymentNumber(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.context/ACCOUNT_DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(105);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/ACCOUNT_DEPLOYMENT_NUMBER has invalid value (" + b + ")"); }
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 24));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 16));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 8));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) b);


    return this;
  }

  public Trace pContextByteCodeAddressHi(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_ADDRESS_HI already set");
    } else {
      filled.set(106);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/BYTE_CODE_ADDRESS_HI has invalid value (" + b + ")"); }
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 24));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 16));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 8));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) b);


    return this;
  }

  public Trace pContextByteCodeAddressLo(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_ADDRESS_LO already set");
    } else {
      filled.set(134);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.context/BYTE_CODE_ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pContextByteCodeCodeFragmentIndex(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(107);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/BYTE_CODE_CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 24));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 16));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 8));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) b);


    return this;
  }

  public Trace pContextByteCodeDeploymentNumber(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(108);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/BYTE_CODE_DEPLOYMENT_NUMBER has invalid value (" + b + ")"); }
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 24));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 16));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 8));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) b);


    return this;
  }

  public Trace pContextByteCodeDeploymentStatus(final long b) {
    if (filled.get(109)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_DEPLOYMENT_STATUS already set");
    } else {
      filled.set(109);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/BYTE_CODE_DEPLOYMENT_STATUS has invalid value (" + b + ")"); }
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 24));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 16));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 8));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) b);


    return this;
  }

  public Trace pContextCallDataContextNumber(final long b) {
    if (filled.get(111)) {
      throw new IllegalStateException("hub.context/CALL_DATA_CONTEXT_NUMBER already set");
    } else {
      filled.set(111);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/CALL_DATA_CONTEXT_NUMBER has invalid value (" + b + ")"); }
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 24));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 16));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 8));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) b);


    return this;
  }

  public Trace pContextCallDataOffset(final long b) {
    if (filled.get(112)) {
      throw new IllegalStateException("hub.context/CALL_DATA_OFFSET already set");
    } else {
      filled.set(112);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/CALL_DATA_OFFSET has invalid value (" + b + ")"); }
    callDataOffsetXorMmuSize.put((byte) (b >> 24));
    callDataOffsetXorMmuSize.put((byte) (b >> 16));
    callDataOffsetXorMmuSize.put((byte) (b >> 8));
    callDataOffsetXorMmuSize.put((byte) b);


    return this;
  }

  public Trace pContextCallDataSize(final long b) {
    if (filled.get(113)) {
      throw new IllegalStateException("hub.context/CALL_DATA_SIZE already set");
    } else {
      filled.set(113);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/CALL_DATA_SIZE has invalid value (" + b + ")"); }
    callDataSizeXorMmuSrcId.put((byte) (b >> 24));
    callDataSizeXorMmuSrcId.put((byte) (b >> 16));
    callDataSizeXorMmuSrcId.put((byte) (b >> 8));
    callDataSizeXorMmuSrcId.put((byte) b);


    return this;
  }

  public Trace pContextCallStackDepth(final long b) {
    if (filled.get(98)) {
      throw new IllegalStateException("hub.context/CALL_STACK_DEPTH already set");
    } else {
      filled.set(98);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.context/CALL_STACK_DEPTH has invalid value (" + b + ")"); }
    callStackDepthXorStackItemHeight1.put((byte) (b >> 8));
    callStackDepthXorStackItemHeight1.put((byte) b);


    return this;
  }

  public Trace pContextCallValue(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.context/CALL_VALUE already set");
    } else {
      filled.set(136);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.context/CALL_VALUE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pContextCallerAddressHi(final long b) {
    if (filled.get(110)) {
      throw new IllegalStateException("hub.context/CALLER_ADDRESS_HI already set");
    } else {
      filled.set(110);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/CALLER_ADDRESS_HI has invalid value (" + b + ")"); }
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 24));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 16));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 8));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) b);


    return this;
  }

  public Trace pContextCallerAddressLo(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.context/CALLER_ADDRESS_LO already set");
    } else {
      filled.set(135);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.context/CALLER_ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace pContextContextNumber(final long b) {
    if (filled.get(114)) {
      throw new IllegalStateException("hub.context/CONTEXT_NUMBER already set");
    } else {
      filled.set(114);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/CONTEXT_NUMBER has invalid value (" + b + ")"); }
    contextNumberXorMmuTgtId.put((byte) (b >> 24));
    contextNumberXorMmuTgtId.put((byte) (b >> 16));
    contextNumberXorMmuTgtId.put((byte) (b >> 8));
    contextNumberXorMmuTgtId.put((byte) b);


    return this;
  }

  public Trace pContextIsRoot(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.context/IS_ROOT already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pContextIsStatic(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.context/IS_STATIC already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pContextReturnAtCapacity(final long b) {
    if (filled.get(115)) {
      throw new IllegalStateException("hub.context/RETURN_AT_CAPACITY already set");
    } else {
      filled.set(115);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/RETURN_AT_CAPACITY has invalid value (" + b + ")"); }
    returnAtCapacityXorMxpInst.put((byte) (b >> 24));
    returnAtCapacityXorMxpInst.put((byte) (b >> 16));
    returnAtCapacityXorMxpInst.put((byte) (b >> 8));
    returnAtCapacityXorMxpInst.put((byte) b);


    return this;
  }

  public Trace pContextReturnAtOffset(final long b) {
    if (filled.get(116)) {
      throw new IllegalStateException("hub.context/RETURN_AT_OFFSET already set");
    } else {
      filled.set(116);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/RETURN_AT_OFFSET has invalid value (" + b + ")"); }
    returnAtOffsetXorOobInst.put((byte) (b >> 24));
    returnAtOffsetXorOobInst.put((byte) (b >> 16));
    returnAtOffsetXorOobInst.put((byte) (b >> 8));
    returnAtOffsetXorOobInst.put((byte) b);


    return this;
  }

  public Trace pContextReturnDataContextNumber(final long b) {
    if (filled.get(117)) {
      throw new IllegalStateException("hub.context/RETURN_DATA_CONTEXT_NUMBER already set");
    } else {
      filled.set(117);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/RETURN_DATA_CONTEXT_NUMBER has invalid value (" + b + ")"); }
    returnDataContextNumberXorStpGasStipend.put((byte) (b >> 24));
    returnDataContextNumberXorStpGasStipend.put((byte) (b >> 16));
    returnDataContextNumberXorStpGasStipend.put((byte) (b >> 8));
    returnDataContextNumberXorStpGasStipend.put((byte) b);


    return this;
  }

  public Trace pContextReturnDataOffset(final long b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.context/RETURN_DATA_OFFSET already set");
    } else {
      filled.set(118);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/RETURN_DATA_OFFSET has invalid value (" + b + ")"); }
    returnDataOffsetXorStpInstruction.put((byte) (b >> 24));
    returnDataOffsetXorStpInstruction.put((byte) (b >> 16));
    returnDataOffsetXorStpInstruction.put((byte) (b >> 8));
    returnDataOffsetXorStpInstruction.put((byte) b);


    return this;
  }

  public Trace pContextReturnDataSize(final long b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.context/RETURN_DATA_SIZE already set");
    } else {
      filled.set(119);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.context/RETURN_DATA_SIZE has invalid value (" + b + ")"); }
    returnDataSize.put((byte) (b >> 24));
    returnDataSize.put((byte) (b >> 16));
    returnDataSize.put((byte) (b >> 8));
    returnDataSize.put((byte) b);


    return this;
  }

  public Trace pContextUpdate(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.context/UPDATE already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscCcrsStamp(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.misc/CCRS_STAMP already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/CCRS_STAMP has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pMiscCcsrFlag(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.misc/CCSR_FLAG already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscExpData1(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_1 already set");
    } else {
      filled.set(133);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/EXP_DATA_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscExpData2(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_2 already set");
    } else {
      filled.set(134);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/EXP_DATA_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscExpData3(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_3 already set");
    } else {
      filled.set(135);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/EXP_DATA_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscExpData4(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_4 already set");
    } else {
      filled.set(136);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/EXP_DATA_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscExpData5(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_5 already set");
    } else {
      filled.set(137);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/EXP_DATA_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscExpFlag(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.misc/EXP_FLAG already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscExpInst(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.misc/EXP_INST already set");
    } else {
      filled.set(105);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/EXP_INST has invalid value (" + b + ")"); }
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 24));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 16));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 8));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) b);


    return this;
  }

  public Trace pMiscMmuAuxId(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.misc/MMU_AUX_ID already set");
    } else {
      filled.set(106);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_AUX_ID has invalid value (" + b + ")"); }
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 24));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 16));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 8));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) b);


    return this;
  }

  public Trace pMiscMmuExoSum(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.misc/MMU_EXO_SUM already set");
    } else {
      filled.set(107);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_EXO_SUM has invalid value (" + b + ")"); }
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 24));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 16));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 8));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) b);


    return this;
  }

  public Trace pMiscMmuFlag(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.misc/MMU_FLAG already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMmuInst(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("hub.misc/MMU_INST already set");
    } else {
      filled.set(108);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_INST has invalid value (" + b + ")"); }
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 24));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 16));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 8));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) b);


    return this;
  }

  public Trace pMiscMmuLimb1(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.misc/MMU_LIMB_1 already set");
    } else {
      filled.set(138);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MMU_LIMB_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMmuLimb2(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.misc/MMU_LIMB_2 already set");
    } else {
      filled.set(139);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MMU_LIMB_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMmuPhase(final long b) {
    if (filled.get(109)) {
      throw new IllegalStateException("hub.misc/MMU_PHASE already set");
    } else {
      filled.set(109);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_PHASE has invalid value (" + b + ")"); }
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 24));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 16));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) (b >> 8));
    deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.put((byte) b);


    return this;
  }

  public Trace pMiscMmuRefOffset(final long b) {
    if (filled.get(110)) {
      throw new IllegalStateException("hub.misc/MMU_REF_OFFSET already set");
    } else {
      filled.set(110);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_REF_OFFSET has invalid value (" + b + ")"); }
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 24));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 16));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) (b >> 8));
    deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.put((byte) b);


    return this;
  }

  public Trace pMiscMmuRefSize(final long b) {
    if (filled.get(111)) {
      throw new IllegalStateException("hub.misc/MMU_REF_SIZE already set");
    } else {
      filled.set(111);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_REF_SIZE has invalid value (" + b + ")"); }
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 24));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 16));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) (b >> 8));
    rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.put((byte) b);


    return this;
  }

  public Trace pMiscMmuSize(final long b) {
    if (filled.get(112)) {
      throw new IllegalStateException("hub.misc/MMU_SIZE already set");
    } else {
      filled.set(112);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_SIZE has invalid value (" + b + ")"); }
    callDataOffsetXorMmuSize.put((byte) (b >> 24));
    callDataOffsetXorMmuSize.put((byte) (b >> 16));
    callDataOffsetXorMmuSize.put((byte) (b >> 8));
    callDataOffsetXorMmuSize.put((byte) b);


    return this;
  }

  public Trace pMiscMmuSrcId(final long b) {
    if (filled.get(113)) {
      throw new IllegalStateException("hub.misc/MMU_SRC_ID already set");
    } else {
      filled.set(113);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_SRC_ID has invalid value (" + b + ")"); }
    callDataSizeXorMmuSrcId.put((byte) (b >> 24));
    callDataSizeXorMmuSrcId.put((byte) (b >> 16));
    callDataSizeXorMmuSrcId.put((byte) (b >> 8));
    callDataSizeXorMmuSrcId.put((byte) b);


    return this;
  }

  public Trace pMiscMmuSrcOffsetHi(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.misc/MMU_SRC_OFFSET_HI already set");
    } else {
      filled.set(140);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MMU_SRC_OFFSET_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMmuSrcOffsetLo(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.misc/MMU_SRC_OFFSET_LO already set");
    } else {
      filled.set(141);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MMU_SRC_OFFSET_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMmuSuccessBit(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.misc/MMU_SUCCESS_BIT already set");
    } else {
      filled.set(48);
    }

    deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMmuTgtId(final long b) {
    if (filled.get(114)) {
      throw new IllegalStateException("hub.misc/MMU_TGT_ID already set");
    } else {
      filled.set(114);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MMU_TGT_ID has invalid value (" + b + ")"); }
    contextNumberXorMmuTgtId.put((byte) (b >> 24));
    contextNumberXorMmuTgtId.put((byte) (b >> 16));
    contextNumberXorMmuTgtId.put((byte) (b >> 8));
    contextNumberXorMmuTgtId.put((byte) b);


    return this;
  }

  public Trace pMiscMmuTgtOffsetLo(final Bytes b) {
    if (filled.get(142)) {
      throw new IllegalStateException("hub.misc/MMU_TGT_OFFSET_LO already set");
    } else {
      filled.set(142);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MMU_TGT_OFFSET_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpDeploys(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.misc/MXP_DEPLOYS already set");
    } else {
      filled.set(49);
    }

    deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpFlag(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("hub.misc/MXP_FLAG already set");
    } else {
      filled.set(50);
    }

    deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpGasMxp(final Bytes b) {
    if (filled.get(143)) {
      throw new IllegalStateException("hub.misc/MXP_GAS_MXP already set");
    } else {
      filled.set(143);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_GAS_MXP has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpInst(final long b) {
    if (filled.get(115)) {
      throw new IllegalStateException("hub.misc/MXP_INST already set");
    } else {
      filled.set(115);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/MXP_INST has invalid value (" + b + ")"); }
    returnAtCapacityXorMxpInst.put((byte) (b >> 24));
    returnAtCapacityXorMxpInst.put((byte) (b >> 16));
    returnAtCapacityXorMxpInst.put((byte) (b >> 8));
    returnAtCapacityXorMxpInst.put((byte) b);


    return this;
  }

  public Trace pMiscMxpMtntop(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("hub.misc/MXP_MTNTOP already set");
    } else {
      filled.set(51);
    }

    existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpMxpx(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("hub.misc/MXP_MXPX already set");
    } else {
      filled.set(52);
    }

    existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpOffset1Hi(final Bytes b) {
    if (filled.get(144)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_1_HI already set");
    } else {
      filled.set(144);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_OFFSET_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpOffset1Lo(final Bytes b) {
    if (filled.get(145)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_1_LO already set");
    } else {
      filled.set(145);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_OFFSET_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { trmRawAddressHiXorMxpOffset1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { trmRawAddressHiXorMxpOffset1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpOffset2Hi(final Bytes b) {
    if (filled.get(146)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_2_HI already set");
    } else {
      filled.set(146);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_OFFSET_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpOffset2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpOffset2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpOffset2Lo(final Bytes b) {
    if (filled.get(147)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_2_LO already set");
    } else {
      filled.set(147);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_OFFSET_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpOffset2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpOffset2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpSize1Hi(final Bytes b) {
    if (filled.get(148)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_1_HI already set");
    } else {
      filled.set(148);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_SIZE_1_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpSize1Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpSize1Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpSize1Lo(final Bytes b) {
    if (filled.get(149)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_1_LO already set");
    } else {
      filled.set(149);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_SIZE_1_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpSize1Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpSize1Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpSize1NonzeroNoMxpx(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_1_NONZERO_NO_MXPX already set");
    } else {
      filled.set(53);
    }

    finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpSize2Hi(final Bytes b) {
    if (filled.get(150)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_2_HI already set");
    } else {
      filled.set(150);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_SIZE_2_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpSize2Hi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpSize2Hi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpSize2Lo(final Bytes b) {
    if (filled.get(151)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_2_LO already set");
    } else {
      filled.set(151);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_SIZE_2_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpSize2Lo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpSize2Lo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscMxpSize2NonzeroNoMxpx(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_2_NONZERO_NO_MXPX already set");
    } else {
      filled.set(54);
    }

    finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpWords(final Bytes b) {
    if (filled.get(152)) {
      throw new IllegalStateException("hub.misc/MXP_WORDS already set");
    } else {
      filled.set(152);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/MXP_WORDS has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { mxpWords.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { mxpWords.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData1(final Bytes b) {
    if (filled.get(153)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_1 already set");
    } else {
      filled.set(153);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData1.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData2(final Bytes b) {
    if (filled.get(154)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_2 already set");
    } else {
      filled.set(154);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData2.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData3(final Bytes b) {
    if (filled.get(155)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_3 already set");
    } else {
      filled.set(155);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData3.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData4(final Bytes b) {
    if (filled.get(156)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_4 already set");
    } else {
      filled.set(156);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData4.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData5(final Bytes b) {
    if (filled.get(157)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_5 already set");
    } else {
      filled.set(157);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_5 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData5.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData5.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData6(final Bytes b) {
    if (filled.get(158)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_6 already set");
    } else {
      filled.set(158);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_6 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData6.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData6.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData7(final Bytes b) {
    if (filled.get(159)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_7 already set");
    } else {
      filled.set(159);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_7 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData7.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData7.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData8(final Bytes b) {
    if (filled.get(160)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_8 already set");
    } else {
      filled.set(160);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_8 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData8.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData8.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobData9(final Bytes b) {
    if (filled.get(161)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_9 already set");
    } else {
      filled.set(161);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/OOB_DATA_9 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { oobData9.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { oobData9.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscOobFlag(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.misc/OOB_FLAG already set");
    } else {
      filled.set(55);
    }

    finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscOobInst(final long b) {
    if (filled.get(116)) {
      throw new IllegalStateException("hub.misc/OOB_INST already set");
    } else {
      filled.set(116);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/OOB_INST has invalid value (" + b + ")"); }
    returnAtOffsetXorOobInst.put((byte) (b >> 24));
    returnAtOffsetXorOobInst.put((byte) (b >> 16));
    returnAtOffsetXorOobInst.put((byte) (b >> 8));
    returnAtOffsetXorOobInst.put((byte) b);


    return this;
  }

  public Trace pMiscStpExists(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.misc/STP_EXISTS already set");
    } else {
      filled.set(56);
    }

    firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscStpFlag(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.misc/STP_FLAG already set");
    } else {
      filled.set(57);
    }

    firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscStpGasHi(final Bytes b) {
    if (filled.get(162)) {
      throw new IllegalStateException("hub.misc/STP_GAS_HI already set");
    } else {
      filled.set(162);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/STP_GAS_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { stpGasHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpGasHi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpGasLo(final Bytes b) {
    if (filled.get(163)) {
      throw new IllegalStateException("hub.misc/STP_GAS_LO already set");
    } else {
      filled.set(163);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/STP_GAS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { stpGasLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpGasLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpGasMxp(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.misc/STP_GAS_MXP already set");
    } else {
      filled.set(124);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.misc/STP_GAS_MXP has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpGasPaidOutOfPocket(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.misc/STP_GAS_PAID_OUT_OF_POCKET already set");
    } else {
      filled.set(125);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.misc/STP_GAS_PAID_OUT_OF_POCKET has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpGasStipend(final long b) {
    if (filled.get(117)) {
      throw new IllegalStateException("hub.misc/STP_GAS_STIPEND already set");
    } else {
      filled.set(117);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/STP_GAS_STIPEND has invalid value (" + b + ")"); }
    returnDataContextNumberXorStpGasStipend.put((byte) (b >> 24));
    returnDataContextNumberXorStpGasStipend.put((byte) (b >> 16));
    returnDataContextNumberXorStpGasStipend.put((byte) (b >> 8));
    returnDataContextNumberXorStpGasStipend.put((byte) b);


    return this;
  }

  public Trace pMiscStpGasUpfrontGasCost(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.misc/STP_GAS_UPFRONT_GAS_COST already set");
    } else {
      filled.set(126);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.misc/STP_GAS_UPFRONT_GAS_COST has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpInstruction(final long b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.misc/STP_INSTRUCTION already set");
    } else {
      filled.set(118);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.misc/STP_INSTRUCTION has invalid value (" + b + ")"); }
    returnDataOffsetXorStpInstruction.put((byte) (b >> 24));
    returnDataOffsetXorStpInstruction.put((byte) (b >> 16));
    returnDataOffsetXorStpInstruction.put((byte) (b >> 8));
    returnDataOffsetXorStpInstruction.put((byte) b);


    return this;
  }

  public Trace pMiscStpOogx(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.misc/STP_OOGX already set");
    } else {
      filled.set(58);
    }

    firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscStpValueHi(final Bytes b) {
    if (filled.get(164)) {
      throw new IllegalStateException("hub.misc/STP_VALUE_HI already set");
    } else {
      filled.set(164);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/STP_VALUE_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { stpValueHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpValueHi.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpValueLo(final Bytes b) {
    if (filled.get(165)) {
      throw new IllegalStateException("hub.misc/STP_VALUE_LO already set");
    } else {
      filled.set(165);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.misc/STP_VALUE_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { stpValueLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpValueLo.put(bs.get(j)); }

    return this;
  }

  public Trace pMiscStpWarmth(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.misc/STP_WARMTH already set");
    } else {
      filled.set(59);
    }

    hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallAbortWillRevert(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.scenario/CALL_ABORT_WILL_REVERT already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallAbortWontRevert(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.scenario/CALL_ABORT_WONT_REVERT already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallEoaSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.scenario/CALL_EOA_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallEoaSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.scenario/CALL_EOA_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(48);
    }

    deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallException(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.scenario/CALL_EXCEPTION already set");
    } else {
      filled.set(49);
    }

    deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallPrcFailure(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("hub.scenario/CALL_PRC_FAILURE already set");
    } else {
      filled.set(50);
    }

    deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallPrcSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("hub.scenario/CALL_PRC_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(51);
    }

    existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallPrcSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("hub.scenario/CALL_PRC_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(52);
    }

    existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcFailureCallerWillRevert(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.scenario/CALL_SMC_FAILURE_CALLER_WILL_REVERT already set");
    } else {
      filled.set(53);
    }

    finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcFailureCallerWontRevert(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.scenario/CALL_SMC_FAILURE_CALLER_WONT_REVERT already set");
    } else {
      filled.set(54);
    }

    finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.scenario/CALL_SMC_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(55);
    }

    finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.scenario/CALL_SMC_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(56);
    }

    firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateAbort(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.scenario/CREATE_ABORT already set");
    } else {
      filled.set(57);
    }

    firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateEmptyInitCodeWillRevert(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.scenario/CREATE_EMPTY_INIT_CODE_WILL_REVERT already set");
    } else {
      filled.set(58);
    }

    firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateEmptyInitCodeWontRevert(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.scenario/CREATE_EMPTY_INIT_CODE_WONT_REVERT already set");
    } else {
      filled.set(59);
    }

    hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateException(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.scenario/CREATE_EXCEPTION already set");
    } else {
      filled.set(60);
    }

    hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateFailureConditionWillRevert(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.scenario/CREATE_FAILURE_CONDITION_WILL_REVERT already set");
    } else {
      filled.set(61);
    }

    isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateFailureConditionWontRevert(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.scenario/CREATE_FAILURE_CONDITION_WONT_REVERT already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeFailureWillRevert(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("hub.scenario/CREATE_NONEMPTY_INIT_CODE_FAILURE_WILL_REVERT already set");
    } else {
      filled.set(63);
    }

    markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeFailureWontRevert(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.scenario/CREATE_NONEMPTY_INIT_CODE_FAILURE_WONT_REVERT already set");
    } else {
      filled.set(64);
    }

    rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeSuccessWillRevert(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException("hub.scenario/CREATE_NONEMPTY_INIT_CODE_SUCCESS_WILL_REVERT already set");
    } else {
      filled.set(65);
    }

    romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeSuccessWontRevert(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException("hub.scenario/CREATE_NONEMPTY_INIT_CODE_SUCCESS_WONT_REVERT already set");
    } else {
      filled.set(66);
    }

    trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcBlake2F(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("hub.scenario/PRC_BLAKE2f already set");
    } else {
      filled.set(67);
    }

    warmthXorPrcBlake2FXorLogFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcCalleeGas(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.scenario/PRC_CALLEE_GAS already set");
    } else {
      filled.set(124);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.scenario/PRC_CALLEE_GAS has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put(bs.get(j)); }

    return this;
  }

  public Trace pScenarioPrcCallerGas(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.scenario/PRC_CALLER_GAS already set");
    } else {
      filled.set(125);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.scenario/PRC_CALLER_GAS has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put(bs.get(j)); }

    return this;
  }

  public Trace pScenarioPrcCdo(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.scenario/PRC_CDO already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.scenario/PRC_CDO has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pScenarioPrcCds(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.scenario/PRC_CDS already set");
    } else {
      filled.set(105);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.scenario/PRC_CDS has invalid value (" + b + ")"); }
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 24));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 16));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 8));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) b);


    return this;
  }

  public Trace pScenarioPrcEcadd(final Boolean b) {
    if (filled.get(68)) {
      throw new IllegalStateException("hub.scenario/PRC_ECADD already set");
    } else {
      filled.set(68);
    }

    warmthNewXorPrcEcaddXorLogInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcEcmul(final Boolean b) {
    if (filled.get(69)) {
      throw new IllegalStateException("hub.scenario/PRC_ECMUL already set");
    } else {
      filled.set(69);
    }

    prcEcmulXorMachineStateFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcEcpairing(final Boolean b) {
    if (filled.get(70)) {
      throw new IllegalStateException("hub.scenario/PRC_ECPAIRING already set");
    } else {
      filled.set(70);
    }

    prcEcpairingXorMaxcsx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcEcrecover(final Boolean b) {
    if (filled.get(71)) {
      throw new IllegalStateException("hub.scenario/PRC_ECRECOVER already set");
    } else {
      filled.set(71);
    }

    prcEcrecoverXorModFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcFailureKnownToHub(final Boolean b) {
    if (filled.get(72)) {
      throw new IllegalStateException("hub.scenario/PRC_FAILURE_KNOWN_TO_HUB already set");
    } else {
      filled.set(72);
    }

    prcFailureKnownToHubXorMulFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcFailureKnownToRam(final Boolean b) {
    if (filled.get(73)) {
      throw new IllegalStateException("hub.scenario/PRC_FAILURE_KNOWN_TO_RAM already set");
    } else {
      filled.set(73);
    }

    prcFailureKnownToRamXorMxpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcIdentity(final Boolean b) {
    if (filled.get(74)) {
      throw new IllegalStateException("hub.scenario/PRC_IDENTITY already set");
    } else {
      filled.set(74);
    }

    prcIdentityXorMxpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcModexp(final Boolean b) {
    if (filled.get(75)) {
      throw new IllegalStateException("hub.scenario/PRC_MODEXP already set");
    } else {
      filled.set(75);
    }

    prcModexpXorOogx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcRac(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.scenario/PRC_RAC already set");
    } else {
      filled.set(106);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.scenario/PRC_RAC has invalid value (" + b + ")"); }
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 24));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 16));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 8));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) b);


    return this;
  }

  public Trace pScenarioPrcRao(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.scenario/PRC_RAO already set");
    } else {
      filled.set(107);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.scenario/PRC_RAO has invalid value (" + b + ")"); }
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 24));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 16));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 8));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) b);


    return this;
  }

  public Trace pScenarioPrcReturnGas(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.scenario/PRC_RETURN_GAS already set");
    } else {
      filled.set(126);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.scenario/PRC_RETURN_GAS has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.put(bs.get(j)); }

    return this;
  }

  public Trace pScenarioPrcRipemd160(final Boolean b) {
    if (filled.get(76)) {
      throw new IllegalStateException("hub.scenario/PRC_RIPEMD-160 already set");
    } else {
      filled.set(76);
    }

    prcRipemd160XorOpcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcSha2256(final Boolean b) {
    if (filled.get(77)) {
      throw new IllegalStateException("hub.scenario/PRC_SHA2-256 already set");
    } else {
      filled.set(77);
    }

    prcSha2256XorPushpopFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(78)) {
      throw new IllegalStateException("hub.scenario/PRC_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(78);
    }

    prcSuccessCallerWillRevertXorRdcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(79)) {
      throw new IllegalStateException("hub.scenario/PRC_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(79);
    }

    prcSuccessCallerWontRevertXorShfFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnException(final Boolean b) {
    if (filled.get(80)) {
      throw new IllegalStateException("hub.scenario/RETURN_EXCEPTION already set");
    } else {
      filled.set(80);
    }

    returnExceptionXorSox.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentEmptyCodeWillRevert(final Boolean b) {
    if (filled.get(81)) {
      throw new IllegalStateException("hub.scenario/RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT already set");
    } else {
      filled.set(81);
    }

    returnFromDeploymentEmptyCodeWillRevertXorSstorex.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentEmptyCodeWontRevert(final Boolean b) {
    if (filled.get(82)) {
      throw new IllegalStateException("hub.scenario/RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT already set");
    } else {
      filled.set(82);
    }

    returnFromDeploymentEmptyCodeWontRevertXorStackramFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentNonemptyCodeWillRevert(final Boolean b) {
    if (filled.get(83)) {
      throw new IllegalStateException("hub.scenario/RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT already set");
    } else {
      filled.set(83);
    }

    returnFromDeploymentNonemptyCodeWillRevertXorStackItemPop1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentNonemptyCodeWontRevert(final Boolean b) {
    if (filled.get(84)) {
      throw new IllegalStateException("hub.scenario/RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT already set");
    } else {
      filled.set(84);
    }

    returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromMessageCallWillTouchRam(final Boolean b) {
    if (filled.get(85)) {
      throw new IllegalStateException("hub.scenario/RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM already set");
    } else {
      filled.set(85);
    }

    returnFromMessageCallWillTouchRamXorStackItemPop3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromMessageCallWontTouchRam(final Boolean b) {
    if (filled.get(86)) {
      throw new IllegalStateException("hub.scenario/RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM already set");
    } else {
      filled.set(86);
    }

    returnFromMessageCallWontTouchRamXorStackItemPop4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructException(final Boolean b) {
    if (filled.get(87)) {
      throw new IllegalStateException("hub.scenario/SELFDESTRUCT_EXCEPTION already set");
    } else {
      filled.set(87);
    }

    selfdestructExceptionXorStaticx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructWillRevert(final Boolean b) {
    if (filled.get(88)) {
      throw new IllegalStateException("hub.scenario/SELFDESTRUCT_WILL_REVERT already set");
    } else {
      filled.set(88);
    }

    selfdestructWillRevertXorStaticFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructWontRevertAlreadyMarked(final Boolean b) {
    if (filled.get(89)) {
      throw new IllegalStateException("hub.scenario/SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED already set");
    } else {
      filled.set(89);
    }

    selfdestructWontRevertAlreadyMarkedXorStoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructWontRevertNotYetMarked(final Boolean b) {
    if (filled.get(90)) {
      throw new IllegalStateException("hub.scenario/SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED already set");
    } else {
      filled.set(90);
    }

    selfdestructWontRevertNotYetMarkedXorSux.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackAccFlag(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.stack/ACC_FLAG already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackAddFlag(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.stack/ADD_FLAG already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackAlpha(final UnsignedByte b) {
    if (filled.get(94)) {
      throw new IllegalStateException("hub.stack/ALPHA already set");
    } else {
      filled.set(94);
    }

    alpha.put(b.toByte());

    return this;
  }

  public Trace pStackBinFlag(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.stack/BIN_FLAG already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackBtcFlag(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.stack/BTC_FLAG already set");
    } else {
      filled.set(48);
    }

    deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackCallFlag(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.stack/CALL_FLAG already set");
    } else {
      filled.set(49);
    }

    deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackConFlag(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("hub.stack/CON_FLAG already set");
    } else {
      filled.set(50);
    }

    deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackCopyFlag(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("hub.stack/COPY_FLAG already set");
    } else {
      filled.set(51);
    }

    existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackCreateFlag(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("hub.stack/CREATE_FLAG already set");
    } else {
      filled.set(52);
    }

    existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag1(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_1 already set");
    } else {
      filled.set(53);
    }

    finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag2(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_2 already set");
    } else {
      filled.set(54);
    }

    finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag3(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_3 already set");
    } else {
      filled.set(55);
    }

    finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag4(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_4 already set");
    } else {
      filled.set(56);
    }

    firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDelta(final UnsignedByte b) {
    if (filled.get(95)) {
      throw new IllegalStateException("hub.stack/DELTA already set");
    } else {
      filled.set(95);
    }

    delta.put(b.toByte());

    return this;
  }

  public Trace pStackDupFlag(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.stack/DUP_FLAG already set");
    } else {
      filled.set(57);
    }

    firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackExtFlag(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.stack/EXT_FLAG already set");
    } else {
      filled.set(58);
    }

    firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackHaltFlag(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.stack/HALT_FLAG already set");
    } else {
      filled.set(59);
    }

    hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackHashInfoFlag(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_FLAG already set");
    } else {
      filled.set(60);
    }

    hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackHashInfoKeccakHi(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_KECCAK_HI already set");
    } else {
      filled.set(133);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/HASH_INFO_KECCAK_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStackHashInfoKeccakLo(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_KECCAK_LO already set");
    } else {
      filled.set(134);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/HASH_INFO_KECCAK_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStackIcpx(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.stack/ICPX already set");
    } else {
      filled.set(61);
    }

    isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackInstruction(final UnsignedByte b) {
    if (filled.get(96)) {
      throw new IllegalStateException("hub.stack/INSTRUCTION already set");
    } else {
      filled.set(96);
    }

    instruction.put(b.toByte());

    return this;
  }

  public Trace pStackInvalidFlag(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.stack/INVALID_FLAG already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackJumpDestinationVettingRequired(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.stack/JUMP_DESTINATION_VETTING_REQUIRED already set");
    } else {
      filled.set(64);
    }

    rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackJumpFlag(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException("hub.stack/JUMP_FLAG already set");
    } else {
      filled.set(65);
    }

    romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackJumpx(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("hub.stack/JUMPX already set");
    } else {
      filled.set(63);
    }

    markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackKecFlag(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException("hub.stack/KEC_FLAG already set");
    } else {
      filled.set(66);
    }

    trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackLogFlag(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("hub.stack/LOG_FLAG already set");
    } else {
      filled.set(67);
    }

    warmthXorPrcBlake2FXorLogFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackLogInfoFlag(final Boolean b) {
    if (filled.get(68)) {
      throw new IllegalStateException("hub.stack/LOG_INFO_FLAG already set");
    } else {
      filled.set(68);
    }

    warmthNewXorPrcEcaddXorLogInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMachineStateFlag(final Boolean b) {
    if (filled.get(69)) {
      throw new IllegalStateException("hub.stack/MACHINE_STATE_FLAG already set");
    } else {
      filled.set(69);
    }

    prcEcmulXorMachineStateFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMaxcsx(final Boolean b) {
    if (filled.get(70)) {
      throw new IllegalStateException("hub.stack/MAXCSX already set");
    } else {
      filled.set(70);
    }

    prcEcpairingXorMaxcsx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackModFlag(final Boolean b) {
    if (filled.get(71)) {
      throw new IllegalStateException("hub.stack/MOD_FLAG already set");
    } else {
      filled.set(71);
    }

    prcEcrecoverXorModFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMulFlag(final Boolean b) {
    if (filled.get(72)) {
      throw new IllegalStateException("hub.stack/MUL_FLAG already set");
    } else {
      filled.set(72);
    }

    prcFailureKnownToHubXorMulFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMxpFlag(final Boolean b) {
    if (filled.get(74)) {
      throw new IllegalStateException("hub.stack/MXP_FLAG already set");
    } else {
      filled.set(74);
    }

    prcIdentityXorMxpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMxpx(final Boolean b) {
    if (filled.get(73)) {
      throw new IllegalStateException("hub.stack/MXPX already set");
    } else {
      filled.set(73);
    }

    prcFailureKnownToRamXorMxpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackOogx(final Boolean b) {
    if (filled.get(75)) {
      throw new IllegalStateException("hub.stack/OOGX already set");
    } else {
      filled.set(75);
    }

    prcModexpXorOogx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackOpcx(final Boolean b) {
    if (filled.get(76)) {
      throw new IllegalStateException("hub.stack/OPCX already set");
    } else {
      filled.set(76);
    }

    prcRipemd160XorOpcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackPushValueHi(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.stack/PUSH_VALUE_HI already set");
    } else {
      filled.set(135);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/PUSH_VALUE_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace pStackPushValueLo(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.stack/PUSH_VALUE_LO already set");
    } else {
      filled.set(136);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/PUSH_VALUE_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStackPushpopFlag(final Boolean b) {
    if (filled.get(77)) {
      throw new IllegalStateException("hub.stack/PUSHPOP_FLAG already set");
    } else {
      filled.set(77);
    }

    prcSha2256XorPushpopFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackRdcx(final Boolean b) {
    if (filled.get(78)) {
      throw new IllegalStateException("hub.stack/RDCX already set");
    } else {
      filled.set(78);
    }

    prcSuccessCallerWillRevertXorRdcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackShfFlag(final Boolean b) {
    if (filled.get(79)) {
      throw new IllegalStateException("hub.stack/SHF_FLAG already set");
    } else {
      filled.set(79);
    }

    prcSuccessCallerWontRevertXorShfFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSox(final Boolean b) {
    if (filled.get(80)) {
      throw new IllegalStateException("hub.stack/SOX already set");
    } else {
      filled.set(80);
    }

    returnExceptionXorSox.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSstorex(final Boolean b) {
    if (filled.get(81)) {
      throw new IllegalStateException("hub.stack/SSTOREX already set");
    } else {
      filled.set(81);
    }

    returnFromDeploymentEmptyCodeWillRevertXorSstorex.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemHeight1(final long b) {
    if (filled.get(98)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_1 already set");
    } else {
      filled.set(98);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_HEIGHT_1 has invalid value (" + b + ")"); }
    callStackDepthXorStackItemHeight1.put((byte) (b >> 8));
    callStackDepthXorStackItemHeight1.put((byte) b);


    return this;
  }

  public Trace pStackStackItemHeight2(final long b) {
    if (filled.get(99)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_2 already set");
    } else {
      filled.set(99);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_HEIGHT_2 has invalid value (" + b + ")"); }
    stackItemHeight2.put((byte) (b >> 8));
    stackItemHeight2.put((byte) b);


    return this;
  }

  public Trace pStackStackItemHeight3(final long b) {
    if (filled.get(100)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_3 already set");
    } else {
      filled.set(100);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_HEIGHT_3 has invalid value (" + b + ")"); }
    stackItemHeight3.put((byte) (b >> 8));
    stackItemHeight3.put((byte) b);


    return this;
  }

  public Trace pStackStackItemHeight4(final long b) {
    if (filled.get(101)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_4 already set");
    } else {
      filled.set(101);
    }

    if(b >= 2048L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_HEIGHT_4 has invalid value (" + b + ")"); }
    stackItemHeight4.put((byte) (b >> 8));
    stackItemHeight4.put((byte) b);


    return this;
  }

  public Trace pStackStackItemPop1(final Boolean b) {
    if (filled.get(83)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_1 already set");
    } else {
      filled.set(83);
    }

    returnFromDeploymentNonemptyCodeWillRevertXorStackItemPop1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemPop2(final Boolean b) {
    if (filled.get(84)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_2 already set");
    } else {
      filled.set(84);
    }

    returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemPop3(final Boolean b) {
    if (filled.get(85)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_3 already set");
    } else {
      filled.set(85);
    }

    returnFromMessageCallWillTouchRamXorStackItemPop3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemPop4(final Boolean b) {
    if (filled.get(86)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_4 already set");
    } else {
      filled.set(86);
    }

    returnFromMessageCallWontTouchRamXorStackItemPop4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemStamp1(final long b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_1 already set");
    } else {
      filled.set(120);
    }

    if(b >= 68719476736L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_STAMP_1 has invalid value (" + b + ")"); }
    stackItemStamp1.put((byte) (b >> 32));
    stackItemStamp1.put((byte) (b >> 24));
    stackItemStamp1.put((byte) (b >> 16));
    stackItemStamp1.put((byte) (b >> 8));
    stackItemStamp1.put((byte) b);


    return this;
  }

  public Trace pStackStackItemStamp2(final long b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_2 already set");
    } else {
      filled.set(121);
    }

    if(b >= 68719476736L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_STAMP_2 has invalid value (" + b + ")"); }
    stackItemStamp2.put((byte) (b >> 32));
    stackItemStamp2.put((byte) (b >> 24));
    stackItemStamp2.put((byte) (b >> 16));
    stackItemStamp2.put((byte) (b >> 8));
    stackItemStamp2.put((byte) b);


    return this;
  }

  public Trace pStackStackItemStamp3(final long b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_3 already set");
    } else {
      filled.set(122);
    }

    if(b >= 68719476736L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_STAMP_3 has invalid value (" + b + ")"); }
    stackItemStamp3.put((byte) (b >> 32));
    stackItemStamp3.put((byte) (b >> 24));
    stackItemStamp3.put((byte) (b >> 16));
    stackItemStamp3.put((byte) (b >> 8));
    stackItemStamp3.put((byte) b);


    return this;
  }

  public Trace pStackStackItemStamp4(final long b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_4 already set");
    } else {
      filled.set(123);
    }

    if(b >= 68719476736L) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_STAMP_4 has invalid value (" + b + ")"); }
    stackItemStamp4.put((byte) (b >> 32));
    stackItemStamp4.put((byte) (b >> 24));
    stackItemStamp4.put((byte) (b >> 16));
    stackItemStamp4.put((byte) (b >> 8));
    stackItemStamp4.put((byte) b);


    return this;
  }

  public Trace pStackStackItemValueHi1(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_1 already set");
    } else {
      filled.set(137);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_HI_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueHi2(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_2 already set");
    } else {
      filled.set(138);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_HI_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueHi3(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_3 already set");
    } else {
      filled.set(139);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_HI_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueHi4(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_4 already set");
    } else {
      filled.set(140);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_HI_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueLo1(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_1 already set");
    } else {
      filled.set(141);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_LO_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueLo2(final Bytes b) {
    if (filled.get(142)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_2 already set");
    } else {
      filled.set(142);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_LO_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueLo3(final Bytes b) {
    if (filled.get(143)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_3 already set");
    } else {
      filled.set(143);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_LO_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackItemValueLo4(final Bytes b) {
    if (filled.get(144)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_4 already set");
    } else {
      filled.set(144);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.stack/STACK_ITEM_VALUE_LO_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.put(bs.get(j)); }

    return this;
  }

  public Trace pStackStackramFlag(final Boolean b) {
    if (filled.get(82)) {
      throw new IllegalStateException("hub.stack/STACKRAM_FLAG already set");
    } else {
      filled.set(82);
    }

    returnFromDeploymentEmptyCodeWontRevertXorStackramFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStaticFlag(final Boolean b) {
    if (filled.get(88)) {
      throw new IllegalStateException("hub.stack/STATIC_FLAG already set");
    } else {
      filled.set(88);
    }

    selfdestructWillRevertXorStaticFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStaticGas(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.stack/STATIC_GAS already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.stack/STATIC_GAS has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pStackStaticx(final Boolean b) {
    if (filled.get(87)) {
      throw new IllegalStateException("hub.stack/STATICX already set");
    } else {
      filled.set(87);
    }

    selfdestructExceptionXorStaticx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStoFlag(final Boolean b) {
    if (filled.get(89)) {
      throw new IllegalStateException("hub.stack/STO_FLAG already set");
    } else {
      filled.set(89);
    }

    selfdestructWontRevertAlreadyMarkedXorStoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSux(final Boolean b) {
    if (filled.get(90)) {
      throw new IllegalStateException("hub.stack/SUX already set");
    } else {
      filled.set(90);
    }

    selfdestructWontRevertNotYetMarkedXorSux.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSwapFlag(final Boolean b) {
    if (filled.get(91)) {
      throw new IllegalStateException("hub.stack/SWAP_FLAG already set");
    } else {
      filled.set(91);
    }

    swapFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackTxnFlag(final Boolean b) {
    if (filled.get(92)) {
      throw new IllegalStateException("hub.stack/TXN_FLAG already set");
    } else {
      filled.set(92);
    }

    txnFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackWcpFlag(final Boolean b) {
    if (filled.get(93)) {
      throw new IllegalStateException("hub.stack/WCP_FLAG already set");
    } else {
      filled.set(93);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageAddressHi(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.storage/ADDRESS_HI already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.storage/ADDRESS_HI has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pStorageAddressLo(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.storage/ADDRESS_LO already set");
    } else {
      filled.set(133);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageAgainInBlk(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.storage/AGAIN_IN_BLK already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageAgainInCnf(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.storage/AGAIN_IN_CNF already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageAgainInTxn(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.storage/AGAIN_IN_TXN already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageDeploymentNumber(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.storage/DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(105);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.storage/DEPLOYMENT_NUMBER has invalid value (" + b + ")"); }
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 24));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 16));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 8));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) b);


    return this;
  }

  public Trace pStorageDeploymentNumberFinalInBlock(final long b) {
    if (filled.get(102)) {
      throw new IllegalStateException("hub.storage/DEPLOYMENT_NUMBER_FINAL_IN_BLOCK already set");
    } else {
      filled.set(102);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("hub.storage/DEPLOYMENT_NUMBER_FINAL_IN_BLOCK has invalid value (" + b + ")"); }
    deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock.put((byte) (b >> 8));
    deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock.put((byte) b);


    return this;
  }

  public Trace pStorageDeploymentNumberFirstInBlock(final long b) {
    if (filled.get(103)) {
      throw new IllegalStateException("hub.storage/DEPLOYMENT_NUMBER_FIRST_IN_BLOCK already set");
    } else {
      filled.set(103);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("hub.storage/DEPLOYMENT_NUMBER_FIRST_IN_BLOCK has invalid value (" + b + ")"); }
    deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock.put((byte) (b >> 8));
    deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock.put((byte) b);


    return this;
  }

  public Trace pStorageDeploymentNumberInfty(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.storage/DEPLOYMENT_NUMBER_INFTY already set");
    } else {
      filled.set(106);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.storage/DEPLOYMENT_NUMBER_INFTY has invalid value (" + b + ")"); }
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 24));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 16));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 8));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) b);


    return this;
  }

  public Trace pStorageFinalInBlk(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.storage/FINAL_IN_BLK already set");
    } else {
      filled.set(48);
    }

    deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageFinalInCnf(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.storage/FINAL_IN_CNF already set");
    } else {
      filled.set(49);
    }

    deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageFinalInTxn(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("hub.storage/FINAL_IN_TXN already set");
    } else {
      filled.set(50);
    }

    deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageFirstInBlk(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("hub.storage/FIRST_IN_BLK already set");
    } else {
      filled.set(51);
    }

    existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageFirstInCnf(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("hub.storage/FIRST_IN_CNF already set");
    } else {
      filled.set(52);
    }

    existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageFirstInTxn(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.storage/FIRST_IN_TXN already set");
    } else {
      filled.set(53);
    }

    finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageStorageKeyHi(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.storage/STORAGE_KEY_HI already set");
    } else {
      filled.set(134);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/STORAGE_KEY_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageStorageKeyLo(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.storage/STORAGE_KEY_LO already set");
    } else {
      filled.set(135);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/STORAGE_KEY_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageValueCurrChanges(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_CHANGES already set");
    } else {
      filled.set(54);
    }

    finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueCurrHi(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_HI already set");
    } else {
      filled.set(136);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/VALUE_CURR_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageValueCurrIsOrig(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_IS_ORIG already set");
    } else {
      filled.set(55);
    }

    finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueCurrIsZero(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_IS_ZERO already set");
    } else {
      filled.set(56);
    }

    firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueCurrLo(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_LO already set");
    } else {
      filled.set(137);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/VALUE_CURR_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageValueNextHi(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_HI already set");
    } else {
      filled.set(138);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/VALUE_NEXT_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageValueNextIsCurr(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_IS_CURR already set");
    } else {
      filled.set(57);
    }

    firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueNextIsOrig(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_IS_ORIG already set");
    } else {
      filled.set(58);
    }

    firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueNextIsZero(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_IS_ZERO already set");
    } else {
      filled.set(59);
    }

    hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueNextLo(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_LO already set");
    } else {
      filled.set(139);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/VALUE_NEXT_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageValueOrigHi(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.storage/VALUE_ORIG_HI already set");
    } else {
      filled.set(140);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/VALUE_ORIG_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageValueOrigIsZero(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.storage/VALUE_ORIG_IS_ZERO already set");
    } else {
      filled.set(60);
    }

    hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueOrigLo(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.storage/VALUE_ORIG_LO already set");
    } else {
      filled.set(141);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.storage/VALUE_ORIG_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.put(bs.get(j)); }

    return this;
  }

  public Trace pStorageWarmth(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.storage/WARMTH already set");
    } else {
      filled.set(61);
    }

    isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageWarmthNew(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.storage/WARMTH_NEW already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionBasefee(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.transaction/BASEFEE already set");
    } else {
      filled.set(124);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/BASEFEE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionCallDataSize(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.transaction/CALL_DATA_SIZE already set");
    } else {
      filled.set(104);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.transaction/CALL_DATA_SIZE has invalid value (" + b + ")"); }
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 24));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 16));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) (b >> 8));
    addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.put((byte) b);


    return this;
  }

  public Trace pTransactionCoinbaseAddressHi(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.transaction/COINBASE_ADDRESS_HI already set");
    } else {
      filled.set(105);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.transaction/COINBASE_ADDRESS_HI has invalid value (" + b + ")"); }
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 24));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 16));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) (b >> 8));
    codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.put((byte) b);


    return this;
  }

  public Trace pTransactionCoinbaseAddressLo(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.transaction/COINBASE_ADDRESS_LO already set");
    } else {
      filled.set(133);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.transaction/COINBASE_ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionCopyTxcd(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.transaction/COPY_TXCD already set");
    } else {
      filled.set(45);
    }

    againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionFromAddressHi(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.transaction/FROM_ADDRESS_HI already set");
    } else {
      filled.set(106);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.transaction/FROM_ADDRESS_HI has invalid value (" + b + ")"); }
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 24));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 16));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) (b >> 8));
    codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.put((byte) b);


    return this;
  }

  public Trace pTransactionFromAddressLo(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.transaction/FROM_ADDRESS_LO already set");
    } else {
      filled.set(134);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.transaction/FROM_ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionGasInitiallyAvailable(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.transaction/GAS_INITIALLY_AVAILABLE already set");
    } else {
      filled.set(125);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/GAS_INITIALLY_AVAILABLE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionGasLeftover(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.transaction/GAS_LEFTOVER already set");
    } else {
      filled.set(126);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/GAS_LEFTOVER has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionGasLimit(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.transaction/GAS_LIMIT already set");
    } else {
      filled.set(127);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/GAS_LIMIT has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasLimit.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasLimit.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionGasPrice(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.transaction/GAS_PRICE already set");
    } else {
      filled.set(128);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/GAS_PRICE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { gasPrice.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { gasPrice.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionInitCodeSize(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.transaction/INIT_CODE_SIZE already set");
    } else {
      filled.set(107);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.transaction/INIT_CODE_SIZE has invalid value (" + b + ")"); }
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 24));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 16));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) (b >> 8));
    codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.put((byte) b);


    return this;
  }

  public Trace pTransactionInitialBalance(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.transaction/INITIAL_BALANCE already set");
    } else {
      filled.set(135);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.transaction/INITIAL_BALANCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionIsDeployment(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.transaction/IS_DEPLOYMENT already set");
    } else {
      filled.set(46);
    }

    againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionIsType2(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.transaction/IS_TYPE2 already set");
    } else {
      filled.set(47);
    }

    againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionNonce(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.transaction/NONCE already set");
    } else {
      filled.set(129);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/NONCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonce.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonce.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionPriorityFeePerGas(final Bytes b) {
    if (filled.get(130)) {
      throw new IllegalStateException("hub.transaction/PRIORITY_FEE_PER_GAS already set");
    } else {
      filled.set(130);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/PRIORITY_FEE_PER_GAS has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { priorityFeePerGas.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { priorityFeePerGas.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionRefundCounterInfinity(final Bytes b) {
    if (filled.get(131)) {
      throw new IllegalStateException("hub.transaction/REFUND_COUNTER_INFINITY already set");
    } else {
      filled.set(131);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/REFUND_COUNTER_INFINITY has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refundCounterInfinity.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refundCounterInfinity.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionRefundEffective(final Bytes b) {
    if (filled.get(132)) {
      throw new IllegalStateException("hub.transaction/REFUND_EFFECTIVE already set");
    } else {
      filled.set(132);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("hub.transaction/REFUND_EFFECTIVE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { refundEffective.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { refundEffective.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionRequiresEvmExecution(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.transaction/REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(48);
    }

    deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionStatusCode(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.transaction/STATUS_CODE already set");
    } else {
      filled.set(49);
    }

    deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionToAddressHi(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("hub.transaction/TO_ADDRESS_HI already set");
    } else {
      filled.set(108);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.transaction/TO_ADDRESS_HI has invalid value (" + b + ")"); }
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 24));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 16));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) (b >> 8));
    deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.put((byte) b);


    return this;
  }

  public Trace pTransactionToAddressLo(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.transaction/TO_ADDRESS_LO already set");
    } else {
      filled.set(136);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.transaction/TO_ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.put(bs.get(j)); }

    return this;
  }

  public Trace pTransactionValue(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.transaction/VALUE already set");
    } else {
      filled.set(137);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("hub.transaction/VALUE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.put(bs.get(j)); }

    return this;
  }

  public Trace peekAtAccount(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("hub.PEEK_AT_ACCOUNT already set");
    } else {
      filled.set(26);
    }

    peekAtAccount.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtContext(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("hub.PEEK_AT_CONTEXT already set");
    } else {
      filled.set(27);
    }

    peekAtContext.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtMiscellaneous(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("hub.PEEK_AT_MISCELLANEOUS already set");
    } else {
      filled.set(28);
    }

    peekAtMiscellaneous.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtScenario(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("hub.PEEK_AT_SCENARIO already set");
    } else {
      filled.set(29);
    }

    peekAtScenario.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtStack(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("hub.PEEK_AT_STACK already set");
    } else {
      filled.set(30);
    }

    peekAtStack.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtStorage(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("hub.PEEK_AT_STORAGE already set");
    } else {
      filled.set(31);
    }

    peekAtStorage.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtTransaction(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("hub.PEEK_AT_TRANSACTION already set");
    } else {
      filled.set(32);
    }

    peekAtTransaction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace programCounter(final long b) {
    if (filled.get(33)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER already set");
    } else {
      filled.set(33);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.PROGRAM_COUNTER has invalid value (" + b + ")"); }
    programCounter.put((byte) (b >> 24));
    programCounter.put((byte) (b >> 16));
    programCounter.put((byte) (b >> 8));
    programCounter.put((byte) b);


    return this;
  }

  public Trace programCounterNew(final long b) {
    if (filled.get(34)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER_NEW already set");
    } else {
      filled.set(34);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.PROGRAM_COUNTER_NEW has invalid value (" + b + ")"); }
    programCounterNew.put((byte) (b >> 24));
    programCounterNew.put((byte) (b >> 16));
    programCounterNew.put((byte) (b >> 8));
    programCounterNew.put((byte) b);


    return this;
  }

  public Trace refundCounter(final long b) {
    if (filled.get(35)) {
      throw new IllegalStateException("hub.REFUND_COUNTER already set");
    } else {
      filled.set(35);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.REFUND_COUNTER has invalid value (" + b + ")"); }
    refundCounter.put((byte) (b >> 24));
    refundCounter.put((byte) (b >> 16));
    refundCounter.put((byte) (b >> 8));
    refundCounter.put((byte) b);


    return this;
  }

  public Trace refundCounterNew(final long b) {
    if (filled.get(36)) {
      throw new IllegalStateException("hub.REFUND_COUNTER_NEW already set");
    } else {
      filled.set(36);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.REFUND_COUNTER_NEW has invalid value (" + b + ")"); }
    refundCounterNew.put((byte) (b >> 24));
    refundCounterNew.put((byte) (b >> 16));
    refundCounterNew.put((byte) (b >> 8));
    refundCounterNew.put((byte) b);


    return this;
  }

  public Trace relativeBlockNumber(final long b) {
    if (filled.get(37)) {
      throw new IllegalStateException("hub.RELATIVE_BLOCK_NUMBER already set");
    } else {
      filled.set(37);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("hub.RELATIVE_BLOCK_NUMBER has invalid value (" + b + ")"); }
    relativeBlockNumber.put((byte) (b >> 8));
    relativeBlockNumber.put((byte) b);


    return this;
  }

  public Trace subStamp(final long b) {
    if (filled.get(38)) {
      throw new IllegalStateException("hub.SUB_STAMP already set");
    } else {
      filled.set(38);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("hub.SUB_STAMP has invalid value (" + b + ")"); }
    subStamp.put((byte) (b >> 24));
    subStamp.put((byte) (b >> 16));
    subStamp.put((byte) (b >> 8));
    subStamp.put((byte) b);


    return this;
  }

  public Trace twoLineInstruction(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("hub.TWO_LINE_INSTRUCTION already set");
    } else {
      filled.set(39);
    }

    twoLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txExec(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("hub.TX_EXEC already set");
    } else {
      filled.set(40);
    }

    txExec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txFinl(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("hub.TX_FINL already set");
    } else {
      filled.set(41);
    }

    txFinl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txInit(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("hub.TX_INIT already set");
    } else {
      filled.set(42);
    }

    txInit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txSkip(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("hub.TX_SKIP already set");
    } else {
      filled.set(43);
    }

    txSkip.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txWarm(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("hub.TX_WARM already set");
    } else {
      filled.set(44);
    }

    txWarm.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("hub.ABSOLUTE_TRANSACTION_NUMBER has not been filled");
    }

    if (!filled.get(104)) {
      throw new IllegalStateException("hub.ADDRESS_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_PRC_CDO_xor_STATIC_GAS_xor_ADDRESS_HI_xor_CALL_DATA_SIZE has not been filled");
    }

    if (!filled.get(133)) {
      throw new IllegalStateException("hub.ADDRESS_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP_DATA_1_xor_HASH_INFO_KECCAK_HI_xor_ADDRESS_LO_xor_COINBASE_ADDRESS_LO has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("hub.AGAIN_IN_BLK_xor_IS_ROOT_xor_CCSR_FLAG_xor_CALL_ABORT_WILL_REVERT_xor_ACC_FLAG_xor_AGAIN_IN_BLK_xor_COPY_TXCD has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("hub.AGAIN_IN_CNF_xor_IS_STATIC_xor_EXP_FLAG_xor_CALL_ABORT_WONT_REVERT_xor_ADD_FLAG_xor_AGAIN_IN_CNF_xor_IS_DEPLOYMENT has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("hub.AGAIN_IN_TXN_xor_UPDATE_xor_MMU_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_BIN_FLAG_xor_AGAIN_IN_TXN_xor_IS_TYPE2 has not been filled");
    }

    if (!filled.get(94)) {
      throw new IllegalStateException("hub.ALPHA has not been filled");
    }

    if (!filled.get(135)) {
      throw new IllegalStateException("hub.BALANCE_NEW_xor_CALLER_ADDRESS_LO_xor_EXP_DATA_3_xor_PUSH_VALUE_HI_xor_STORAGE_KEY_LO_xor_INITIAL_BALANCE has not been filled");
    }

    if (!filled.get(134)) {
      throw new IllegalStateException("hub.BALANCE_xor_BYTE_CODE_ADDRESS_LO_xor_EXP_DATA_2_xor_HASH_INFO_KECCAK_LO_xor_STORAGE_KEY_HI_xor_FROM_ADDRESS_LO has not been filled");
    }

    if (!filled.get(112)) {
      throw new IllegalStateException("hub.CALL_DATA_OFFSET_xor_MMU_SIZE has not been filled");
    }

    if (!filled.get(113)) {
      throw new IllegalStateException("hub.CALL_DATA_SIZE_xor_MMU_SRC_ID has not been filled");
    }

    if (!filled.get(98)) {
      throw new IllegalStateException("hub.CALL_STACK_DEPTH_xor_STACK_ITEM_HEIGHT_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("hub.CALLER_CONTEXT_NUMBER has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("hub.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(105)) {
      throw new IllegalStateException("hub.CODE_FRAGMENT_INDEX_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP_INST_xor_PRC_CDS_xor_DEPLOYMENT_NUMBER_xor_COINBASE_ADDRESS_HI has not been filled");
    }

    if (!filled.get(137)) {
      throw new IllegalStateException("hub.CODE_HASH_HI_NEW_xor_EXP_DATA_5_xor_STACK_ITEM_VALUE_HI_1_xor_VALUE_CURR_LO_xor_VALUE has not been filled");
    }

    if (!filled.get(136)) {
      throw new IllegalStateException("hub.CODE_HASH_HI_xor_CALL_VALUE_xor_EXP_DATA_4_xor_PUSH_VALUE_LO_xor_VALUE_CURR_HI_xor_TO_ADDRESS_LO has not been filled");
    }

    if (!filled.get(139)) {
      throw new IllegalStateException("hub.CODE_HASH_LO_NEW_xor_MMU_LIMB_2_xor_STACK_ITEM_VALUE_HI_3_xor_VALUE_NEXT_LO has not been filled");
    }

    if (!filled.get(138)) {
      throw new IllegalStateException("hub.CODE_HASH_LO_xor_MMU_LIMB_1_xor_STACK_ITEM_VALUE_HI_2_xor_VALUE_NEXT_HI has not been filled");
    }

    if (!filled.get(107)) {
      throw new IllegalStateException("hub.CODE_SIZE_NEW_xor_BYTE_CODE_CODE_FRAGMENT_INDEX_xor_MMU_EXO_SUM_xor_PRC_RAO_xor_INIT_CODE_SIZE has not been filled");
    }

    if (!filled.get(106)) {
      throw new IllegalStateException("hub.CODE_SIZE_xor_BYTE_CODE_ADDRESS_HI_xor_MMU_AUX_ID_xor_PRC_RAC_xor_DEPLOYMENT_NUMBER_INFTY_xor_FROM_ADDRESS_HI has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("hub.CONTEXT_GETS_REVERTED has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("hub.CONTEXT_MAY_CHANGE has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER_NEW has not been filled");
    }

    if (!filled.get(114)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER_xor_MMU_TGT_ID has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("hub.CONTEXT_REVERT_STAMP has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("hub.CONTEXT_SELF_REVERTS has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("hub.CONTEXT_WILL_REVERT has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("hub.COUNTER_NSR has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("hub.COUNTER_TLI has not been filled");
    }

    if (!filled.get(95)) {
      throw new IllegalStateException("hub.DELTA has not been filled");
    }

    if (!filled.get(102)) {
      throw new IllegalStateException("hub.DEPLOYMENT_NUMBER_FINAL_IN_BLOCK_xor_DEPLOYMENT_NUMBER_FINAL_IN_BLOCK has not been filled");
    }

    if (!filled.get(103)) {
      throw new IllegalStateException("hub.DEPLOYMENT_NUMBER_FIRST_IN_BLOCK_xor_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK has not been filled");
    }

    if (!filled.get(109)) {
      throw new IllegalStateException("hub.DEPLOYMENT_NUMBER_INFTY_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MMU_PHASE has not been filled");
    }

    if (!filled.get(110)) {
      throw new IllegalStateException("hub.DEPLOYMENT_NUMBER_NEW_xor_CALLER_ADDRESS_HI_xor_MMU_REF_OFFSET has not been filled");
    }

    if (!filled.get(108)) {
      throw new IllegalStateException("hub.DEPLOYMENT_NUMBER_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MMU_INST_xor_TO_ADDRESS_HI has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("hub.DEPLOYMENT_STATUS_INFTY_xor_MXP_DEPLOYS_xor_CALL_EXCEPTION_xor_CALL_FLAG_xor_FINAL_IN_CNF_xor_STATUS_CODE has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("hub.DEPLOYMENT_STATUS_NEW_xor_MXP_FLAG_xor_CALL_PRC_FAILURE_xor_CON_FLAG_xor_FINAL_IN_TXN has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("hub.DEPLOYMENT_STATUS_xor_MMU_SUCCESS_BIT_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BTC_FLAG_xor_FINAL_IN_BLK_xor_REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("hub.DOM_STAMP has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("hub.EXCEPTION_AHOY has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("hub.EXISTS_NEW_xor_MXP_MXPX_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_CREATE_FLAG_xor_FIRST_IN_CNF has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("hub.EXISTS_xor_MXP_MTNTOP_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_COPY_FLAG_xor_FIRST_IN_BLK has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("hub.FINAL_IN_BLK_xor_MXP_SIZE_1_NONZERO_NO_MXPX_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_DEC_FLAG_1_xor_FIRST_IN_TXN has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("hub.FINAL_IN_CNF_xor_MXP_SIZE_2_NONZERO_NO_MXPX_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DEC_FLAG_2_xor_VALUE_CURR_CHANGES has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("hub.FINAL_IN_TXN_xor_OOB_FLAG_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DEC_FLAG_3_xor_VALUE_CURR_IS_ORIG has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("hub.FIRST_IN_BLK_xor_STP_EXISTS_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DEC_FLAG_4_xor_VALUE_CURR_IS_ZERO has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("hub.FIRST_IN_CNF_xor_STP_FLAG_xor_CREATE_ABORT_xor_DUP_FLAG_xor_VALUE_NEXT_IS_CURR has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("hub.FIRST_IN_TXN_xor_STP_OOGX_xor_CREATE_EMPTY_INIT_CODE_WILL_REVERT_xor_EXT_FLAG_xor_VALUE_NEXT_IS_ORIG has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("hub.GAS_ACTUAL has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("hub.GAS_COST has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("hub.GAS_EXPECTED has not been filled");
    }

    if (!filled.get(127)) {
      throw new IllegalStateException("hub.GAS_LIMIT has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("hub.GAS_NEXT has not been filled");
    }

    if (!filled.get(128)) {
      throw new IllegalStateException("hub.GAS_PRICE has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("hub.HAS_CODE_NEW_xor_CREATE_EXCEPTION_xor_HASH_INFO_FLAG_xor_VALUE_ORIG_IS_ZERO has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("hub.HAS_CODE_xor_STP_WARMTH_xor_CREATE_EMPTY_INIT_CODE_WONT_REVERT_xor_HALT_FLAG_xor_VALUE_NEXT_IS_ZERO has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("hub.HEIGHT has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("hub.HEIGHT_NEW has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("hub.HUB_STAMP has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("hub.HUB_STAMP_TRANSACTION_END has not been filled");
    }

    if (!filled.get(96)) {
      throw new IllegalStateException("hub.INSTRUCTION has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("hub.IS_PRECOMPILE_xor_CREATE_FAILURE_CONDITION_WILL_REVERT_xor_ICPX_xor_WARMTH has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("hub.LOG_INFO_STAMP has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("hub.MARKED_FOR_SELFDESTRUCT_NEW_xor_CREATE_NONEMPTY_INIT_CODE_FAILURE_WILL_REVERT_xor_JUMPX has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("hub.MARKED_FOR_SELFDESTRUCT_xor_CREATE_FAILURE_CONDITION_WONT_REVERT_xor_INVALID_FLAG_xor_WARMTH_NEW has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("hub.MMU_STAMP has not been filled");
    }

    if (!filled.get(146)) {
      throw new IllegalStateException("hub.MXP_OFFSET_2_HI has not been filled");
    }

    if (!filled.get(147)) {
      throw new IllegalStateException("hub.MXP_OFFSET_2_LO has not been filled");
    }

    if (!filled.get(148)) {
      throw new IllegalStateException("hub.MXP_SIZE_1_HI has not been filled");
    }

    if (!filled.get(149)) {
      throw new IllegalStateException("hub.MXP_SIZE_1_LO has not been filled");
    }

    if (!filled.get(150)) {
      throw new IllegalStateException("hub.MXP_SIZE_2_HI has not been filled");
    }

    if (!filled.get(151)) {
      throw new IllegalStateException("hub.MXP_SIZE_2_LO has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("hub.MXP_STAMP has not been filled");
    }

    if (!filled.get(152)) {
      throw new IllegalStateException("hub.MXP_WORDS has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("hub.NON_STACK_ROWS has not been filled");
    }

    if (!filled.get(129)) {
      throw new IllegalStateException("hub.NONCE has not been filled");
    }

    if (!filled.get(125)) {
      throw new IllegalStateException("hub.NONCE_NEW_xor_STP_GAS_PAID_OUT_OF_POCKET_xor_PRC_CALLER_GAS_xor_GAS_INITIALLY_AVAILABLE has not been filled");
    }

    if (!filled.get(124)) {
      throw new IllegalStateException("hub.NONCE_xor_STP_GAS_MXP_xor_PRC_CALLEE_GAS_xor_BASEFEE has not been filled");
    }

    if (!filled.get(153)) {
      throw new IllegalStateException("hub.OOB_DATA_1 has not been filled");
    }

    if (!filled.get(154)) {
      throw new IllegalStateException("hub.OOB_DATA_2 has not been filled");
    }

    if (!filled.get(155)) {
      throw new IllegalStateException("hub.OOB_DATA_3 has not been filled");
    }

    if (!filled.get(156)) {
      throw new IllegalStateException("hub.OOB_DATA_4 has not been filled");
    }

    if (!filled.get(157)) {
      throw new IllegalStateException("hub.OOB_DATA_5 has not been filled");
    }

    if (!filled.get(158)) {
      throw new IllegalStateException("hub.OOB_DATA_6 has not been filled");
    }

    if (!filled.get(159)) {
      throw new IllegalStateException("hub.OOB_DATA_7 has not been filled");
    }

    if (!filled.get(160)) {
      throw new IllegalStateException("hub.OOB_DATA_8 has not been filled");
    }

    if (!filled.get(161)) {
      throw new IllegalStateException("hub.OOB_DATA_9 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("hub.PEEK_AT_ACCOUNT has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("hub.PEEK_AT_CONTEXT has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("hub.PEEK_AT_MISCELLANEOUS has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("hub.PEEK_AT_SCENARIO has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("hub.PEEK_AT_STACK has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("hub.PEEK_AT_STORAGE has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("hub.PEEK_AT_TRANSACTION has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException("hub.PRC_ECMUL_xor_MACHINE_STATE_FLAG has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException("hub.PRC_ECPAIRING_xor_MAXCSX has not been filled");
    }

    if (!filled.get(71)) {
      throw new IllegalStateException("hub.PRC_ECRECOVER_xor_MOD_FLAG has not been filled");
    }

    if (!filled.get(72)) {
      throw new IllegalStateException("hub.PRC_FAILURE_KNOWN_TO_HUB_xor_MUL_FLAG has not been filled");
    }

    if (!filled.get(73)) {
      throw new IllegalStateException("hub.PRC_FAILURE_KNOWN_TO_RAM_xor_MXPX has not been filled");
    }

    if (!filled.get(74)) {
      throw new IllegalStateException("hub.PRC_IDENTITY_xor_MXP_FLAG has not been filled");
    }

    if (!filled.get(75)) {
      throw new IllegalStateException("hub.PRC_MODEXP_xor_OOGX has not been filled");
    }

    if (!filled.get(76)) {
      throw new IllegalStateException("hub.PRC_RIPEMD-160_xor_OPCX has not been filled");
    }

    if (!filled.get(77)) {
      throw new IllegalStateException("hub.PRC_SHA2-256_xor_PUSHPOP_FLAG has not been filled");
    }

    if (!filled.get(78)) {
      throw new IllegalStateException("hub.PRC_SUCCESS_CALLER_WILL_REVERT_xor_RDCX has not been filled");
    }

    if (!filled.get(79)) {
      throw new IllegalStateException("hub.PRC_SUCCESS_CALLER_WONT_REVERT_xor_SHF_FLAG has not been filled");
    }

    if (!filled.get(130)) {
      throw new IllegalStateException("hub.PRIORITY_FEE_PER_GAS has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER_NEW has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("hub.REFUND_COUNTER has not been filled");
    }

    if (!filled.get(131)) {
      throw new IllegalStateException("hub.REFUND_COUNTER_INFINITY has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("hub.REFUND_COUNTER_NEW has not been filled");
    }

    if (!filled.get(132)) {
      throw new IllegalStateException("hub.REFUND_EFFECTIVE has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("hub.RELATIVE_BLOCK_NUMBER has not been filled");
    }

    if (!filled.get(115)) {
      throw new IllegalStateException("hub.RETURN_AT_CAPACITY_xor_MXP_INST has not been filled");
    }

    if (!filled.get(116)) {
      throw new IllegalStateException("hub.RETURN_AT_OFFSET_xor_OOB_INST has not been filled");
    }

    if (!filled.get(117)) {
      throw new IllegalStateException("hub.RETURN_DATA_CONTEXT_NUMBER_xor_STP_GAS_STIPEND has not been filled");
    }

    if (!filled.get(118)) {
      throw new IllegalStateException("hub.RETURN_DATA_OFFSET_xor_STP_INSTRUCTION has not been filled");
    }

    if (!filled.get(119)) {
      throw new IllegalStateException("hub.RETURN_DATA_SIZE has not been filled");
    }

    if (!filled.get(80)) {
      throw new IllegalStateException("hub.RETURN_EXCEPTION_xor_SOX has not been filled");
    }

    if (!filled.get(81)) {
      throw new IllegalStateException("hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT_xor_SSTOREX has not been filled");
    }

    if (!filled.get(82)) {
      throw new IllegalStateException("hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT_xor_STACKRAM_FLAG has not been filled");
    }

    if (!filled.get(83)) {
      throw new IllegalStateException("hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT_xor_STACK_ITEM_POP_1 has not been filled");
    }

    if (!filled.get(84)) {
      throw new IllegalStateException("hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT_xor_STACK_ITEM_POP_2 has not been filled");
    }

    if (!filled.get(85)) {
      throw new IllegalStateException("hub.RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM_xor_STACK_ITEM_POP_3 has not been filled");
    }

    if (!filled.get(86)) {
      throw new IllegalStateException("hub.RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM_xor_STACK_ITEM_POP_4 has not been filled");
    }

    if (!filled.get(111)) {
      throw new IllegalStateException("hub.RLPADDR_DEP_ADDR_HI_xor_CALL_DATA_CONTEXT_NUMBER_xor_MMU_REF_SIZE has not been filled");
    }

    if (!filled.get(140)) {
      throw new IllegalStateException("hub.RLPADDR_DEP_ADDR_LO_xor_MMU_SRC_OFFSET_HI_xor_STACK_ITEM_VALUE_HI_4_xor_VALUE_ORIG_HI has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("hub.RLPADDR_FLAG_xor_CREATE_NONEMPTY_INIT_CODE_FAILURE_WONT_REVERT_xor_JUMP_DESTINATION_VETTING_REQUIRED has not been filled");
    }

    if (!filled.get(141)) {
      throw new IllegalStateException("hub.RLPADDR_KEC_HI_xor_MMU_SRC_OFFSET_LO_xor_STACK_ITEM_VALUE_LO_1_xor_VALUE_ORIG_LO has not been filled");
    }

    if (!filled.get(142)) {
      throw new IllegalStateException("hub.RLPADDR_KEC_LO_xor_MMU_TGT_OFFSET_LO_xor_STACK_ITEM_VALUE_LO_2 has not been filled");
    }

    if (!filled.get(97)) {
      throw new IllegalStateException("hub.RLPADDR_RECIPE has not been filled");
    }

    if (!filled.get(143)) {
      throw new IllegalStateException("hub.RLPADDR_SALT_HI_xor_MXP_GAS_MXP_xor_STACK_ITEM_VALUE_LO_3 has not been filled");
    }

    if (!filled.get(144)) {
      throw new IllegalStateException("hub.RLPADDR_SALT_LO_xor_MXP_OFFSET_1_HI_xor_STACK_ITEM_VALUE_LO_4 has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("hub.ROMLEX_FLAG_xor_CREATE_NONEMPTY_INIT_CODE_SUCCESS_WILL_REVERT_xor_JUMP_FLAG has not been filled");
    }

    if (!filled.get(87)) {
      throw new IllegalStateException("hub.SELFDESTRUCT_EXCEPTION_xor_STATICX has not been filled");
    }

    if (!filled.get(88)) {
      throw new IllegalStateException("hub.SELFDESTRUCT_WILL_REVERT_xor_STATIC_FLAG has not been filled");
    }

    if (!filled.get(89)) {
      throw new IllegalStateException("hub.SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED_xor_STO_FLAG has not been filled");
    }

    if (!filled.get(90)) {
      throw new IllegalStateException("hub.SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED_xor_SUX has not been filled");
    }

    if (!filled.get(99)) {
      throw new IllegalStateException("hub.STACK_ITEM_HEIGHT_2 has not been filled");
    }

    if (!filled.get(100)) {
      throw new IllegalStateException("hub.STACK_ITEM_HEIGHT_3 has not been filled");
    }

    if (!filled.get(101)) {
      throw new IllegalStateException("hub.STACK_ITEM_HEIGHT_4 has not been filled");
    }

    if (!filled.get(120)) {
      throw new IllegalStateException("hub.STACK_ITEM_STAMP_1 has not been filled");
    }

    if (!filled.get(121)) {
      throw new IllegalStateException("hub.STACK_ITEM_STAMP_2 has not been filled");
    }

    if (!filled.get(122)) {
      throw new IllegalStateException("hub.STACK_ITEM_STAMP_3 has not been filled");
    }

    if (!filled.get(123)) {
      throw new IllegalStateException("hub.STACK_ITEM_STAMP_4 has not been filled");
    }

    if (!filled.get(162)) {
      throw new IllegalStateException("hub.STP_GAS_HI has not been filled");
    }

    if (!filled.get(163)) {
      throw new IllegalStateException("hub.STP_GAS_LO has not been filled");
    }

    if (!filled.get(126)) {
      throw new IllegalStateException("hub.STP_GAS_UPFRONT_GAS_COST_xor_PRC_RETURN_GAS_xor_GAS_LEFTOVER has not been filled");
    }

    if (!filled.get(164)) {
      throw new IllegalStateException("hub.STP_VALUE_HI has not been filled");
    }

    if (!filled.get(165)) {
      throw new IllegalStateException("hub.STP_VALUE_LO has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("hub.SUB_STAMP has not been filled");
    }

    if (!filled.get(91)) {
      throw new IllegalStateException("hub.SWAP_FLAG has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException("hub.TRM_FLAG_xor_CREATE_NONEMPTY_INIT_CODE_SUCCESS_WONT_REVERT_xor_KEC_FLAG has not been filled");
    }

    if (!filled.get(145)) {
      throw new IllegalStateException("hub.TRM_RAW_ADDRESS_HI_xor_MXP_OFFSET_1_LO has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("hub.TWO_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("hub.TX_EXEC has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("hub.TX_FINL has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("hub.TX_INIT has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("hub.TX_SKIP has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("hub.TX_WARM has not been filled");
    }

    if (!filled.get(92)) {
      throw new IllegalStateException("hub.TXN_FLAG has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException("hub.WARMTH_NEW_xor_PRC_ECADD_xor_LOG_INFO_FLAG has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException("hub.WARMTH_xor_PRC_BLAKE2f_xor_LOG_FLAG has not been filled");
    }

    if (!filled.get(93)) {
      throw new IllegalStateException("hub.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absoluteTransactionNumber.position(absoluteTransactionNumber.position() + 2);
    }

    if (!filled.get(104)) {
      addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.position(addressHiXorAccountAddressHiXorCcrsStampXorPrcCdoXorStaticGasXorAddressHiXorCallDataSize.position() + 4);
    }

    if (!filled.get(133)) {
      addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.position(addressLoXorAccountAddressLoXorExpData1XorHashInfoKeccakHiXorAddressLoXorCoinbaseAddressLo.position() + 16);
    }

    if (!filled.get(45)) {
      againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.position(againInBlkXorIsRootXorCcsrFlagXorCallAbortWillRevertXorAccFlagXorAgainInBlkXorCopyTxcd.position() + 1);
    }

    if (!filled.get(46)) {
      againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.position(againInCnfXorIsStaticXorExpFlagXorCallAbortWontRevertXorAddFlagXorAgainInCnfXorIsDeployment.position() + 1);
    }

    if (!filled.get(47)) {
      againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.position(againInTxnXorUpdateXorMmuFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorAgainInTxnXorIsType2.position() + 1);
    }

    if (!filled.get(94)) {
      alpha.position(alpha.position() + 1);
    }

    if (!filled.get(135)) {
      balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.position(balanceNewXorCallerAddressLoXorExpData3XorPushValueHiXorStorageKeyLoXorInitialBalance.position() + 16);
    }

    if (!filled.get(134)) {
      balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.position(balanceXorByteCodeAddressLoXorExpData2XorHashInfoKeccakLoXorStorageKeyHiXorFromAddressLo.position() + 16);
    }

    if (!filled.get(112)) {
      callDataOffsetXorMmuSize.position(callDataOffsetXorMmuSize.position() + 4);
    }

    if (!filled.get(113)) {
      callDataSizeXorMmuSrcId.position(callDataSizeXorMmuSrcId.position() + 4);
    }

    if (!filled.get(98)) {
      callStackDepthXorStackItemHeight1.position(callStackDepthXorStackItemHeight1.position() + 2);
    }

    if (!filled.get(1)) {
      callerContextNumber.position(callerContextNumber.position() + 4);
    }

    if (!filled.get(2)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 4);
    }

    if (!filled.get(105)) {
      codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.position(codeFragmentIndexXorAccountDeploymentNumberXorExpInstXorPrcCdsXorDeploymentNumberXorCoinbaseAddressHi.position() + 4);
    }

    if (!filled.get(137)) {
      codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.position(codeHashHiNewXorExpData5XorStackItemValueHi1XorValueCurrLoXorValue.position() + 16);
    }

    if (!filled.get(136)) {
      codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.position(codeHashHiXorCallValueXorExpData4XorPushValueLoXorValueCurrHiXorToAddressLo.position() + 16);
    }

    if (!filled.get(139)) {
      codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.position(codeHashLoNewXorMmuLimb2XorStackItemValueHi3XorValueNextLo.position() + 16);
    }

    if (!filled.get(138)) {
      codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.position(codeHashLoXorMmuLimb1XorStackItemValueHi2XorValueNextHi.position() + 16);
    }

    if (!filled.get(107)) {
      codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.position(codeSizeNewXorByteCodeCodeFragmentIndexXorMmuExoSumXorPrcRaoXorInitCodeSize.position() + 4);
    }

    if (!filled.get(106)) {
      codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.position(codeSizeXorByteCodeAddressHiXorMmuAuxIdXorPrcRacXorDeploymentNumberInftyXorFromAddressHi.position() + 4);
    }

    if (!filled.get(3)) {
      contextGetsReverted.position(contextGetsReverted.position() + 1);
    }

    if (!filled.get(4)) {
      contextMayChange.position(contextMayChange.position() + 1);
    }

    if (!filled.get(5)) {
      contextNumber.position(contextNumber.position() + 4);
    }

    if (!filled.get(6)) {
      contextNumberNew.position(contextNumberNew.position() + 4);
    }

    if (!filled.get(114)) {
      contextNumberXorMmuTgtId.position(contextNumberXorMmuTgtId.position() + 4);
    }

    if (!filled.get(7)) {
      contextRevertStamp.position(contextRevertStamp.position() + 4);
    }

    if (!filled.get(8)) {
      contextSelfReverts.position(contextSelfReverts.position() + 1);
    }

    if (!filled.get(9)) {
      contextWillRevert.position(contextWillRevert.position() + 1);
    }

    if (!filled.get(10)) {
      counterNsr.position(counterNsr.position() + 1);
    }

    if (!filled.get(11)) {
      counterTli.position(counterTli.position() + 1);
    }

    if (!filled.get(95)) {
      delta.position(delta.position() + 1);
    }

    if (!filled.get(102)) {
      deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock.position(deploymentNumberFinalInBlockXorDeploymentNumberFinalInBlock.position() + 2);
    }

    if (!filled.get(103)) {
      deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock.position(deploymentNumberFirstInBlockXorDeploymentNumberFirstInBlock.position() + 2);
    }

    if (!filled.get(109)) {
      deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.position(deploymentNumberInftyXorByteCodeDeploymentStatusXorMmuPhase.position() + 4);
    }

    if (!filled.get(110)) {
      deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.position(deploymentNumberNewXorCallerAddressHiXorMmuRefOffset.position() + 4);
    }

    if (!filled.get(108)) {
      deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.position(deploymentNumberXorByteCodeDeploymentNumberXorMmuInstXorToAddressHi.position() + 4);
    }

    if (!filled.get(49)) {
      deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.position(deploymentStatusInftyXorMxpDeploysXorCallExceptionXorCallFlagXorFinalInCnfXorStatusCode.position() + 1);
    }

    if (!filled.get(50)) {
      deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.position(deploymentStatusNewXorMxpFlagXorCallPrcFailureXorConFlagXorFinalInTxn.position() + 1);
    }

    if (!filled.get(48)) {
      deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.position(deploymentStatusXorMmuSuccessBitXorCallEoaSuccessCallerWontRevertXorBtcFlagXorFinalInBlkXorRequiresEvmExecution.position() + 1);
    }

    if (!filled.get(12)) {
      domStamp.position(domStamp.position() + 4);
    }

    if (!filled.get(13)) {
      exceptionAhoy.position(exceptionAhoy.position() + 1);
    }

    if (!filled.get(52)) {
      existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.position(existsNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCreateFlagXorFirstInCnf.position() + 1);
    }

    if (!filled.get(51)) {
      existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.position(existsXorMxpMtntopXorCallPrcSuccessCallerWillRevertXorCopyFlagXorFirstInBlk.position() + 1);
    }

    if (!filled.get(53)) {
      finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.position(finalInBlkXorMxpSize1NonzeroNoMxpxXorCallSmcFailureCallerWillRevertXorDecFlag1XorFirstInTxn.position() + 1);
    }

    if (!filled.get(54)) {
      finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.position(finalInCnfXorMxpSize2NonzeroNoMxpxXorCallSmcFailureCallerWontRevertXorDecFlag2XorValueCurrChanges.position() + 1);
    }

    if (!filled.get(55)) {
      finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.position(finalInTxnXorOobFlagXorCallSmcSuccessCallerWillRevertXorDecFlag3XorValueCurrIsOrig.position() + 1);
    }

    if (!filled.get(56)) {
      firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.position(firstInBlkXorStpExistsXorCallSmcSuccessCallerWontRevertXorDecFlag4XorValueCurrIsZero.position() + 1);
    }

    if (!filled.get(57)) {
      firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.position(firstInCnfXorStpFlagXorCreateAbortXorDupFlagXorValueNextIsCurr.position() + 1);
    }

    if (!filled.get(58)) {
      firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.position(firstInTxnXorStpOogxXorCreateEmptyInitCodeWillRevertXorExtFlagXorValueNextIsOrig.position() + 1);
    }

    if (!filled.get(14)) {
      gasActual.position(gasActual.position() + 8);
    }

    if (!filled.get(15)) {
      gasCost.position(gasCost.position() + 8);
    }

    if (!filled.get(16)) {
      gasExpected.position(gasExpected.position() + 8);
    }

    if (!filled.get(127)) {
      gasLimit.position(gasLimit.position() + 8);
    }

    if (!filled.get(17)) {
      gasNext.position(gasNext.position() + 8);
    }

    if (!filled.get(128)) {
      gasPrice.position(gasPrice.position() + 8);
    }

    if (!filled.get(60)) {
      hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero.position(hasCodeNewXorCreateExceptionXorHashInfoFlagXorValueOrigIsZero.position() + 1);
    }

    if (!filled.get(59)) {
      hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.position(hasCodeXorStpWarmthXorCreateEmptyInitCodeWontRevertXorHaltFlagXorValueNextIsZero.position() + 1);
    }

    if (!filled.get(18)) {
      height.position(height.position() + 2);
    }

    if (!filled.get(19)) {
      heightNew.position(heightNew.position() + 2);
    }

    if (!filled.get(20)) {
      hubStamp.position(hubStamp.position() + 4);
    }

    if (!filled.get(21)) {
      hubStampTransactionEnd.position(hubStampTransactionEnd.position() + 4);
    }

    if (!filled.get(96)) {
      instruction.position(instruction.position() + 1);
    }

    if (!filled.get(61)) {
      isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth.position(isPrecompileXorCreateFailureConditionWillRevertXorIcpxXorWarmth.position() + 1);
    }

    if (!filled.get(22)) {
      logInfoStamp.position(logInfoStamp.position() + 4);
    }

    if (!filled.get(63)) {
      markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx.position(markedForSelfdestructNewXorCreateNonemptyInitCodeFailureWillRevertXorJumpx.position() + 1);
    }

    if (!filled.get(62)) {
      markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew.position(markedForSelfdestructXorCreateFailureConditionWontRevertXorInvalidFlagXorWarmthNew.position() + 1);
    }

    if (!filled.get(23)) {
      mmuStamp.position(mmuStamp.position() + 4);
    }

    if (!filled.get(146)) {
      mxpOffset2Hi.position(mxpOffset2Hi.position() + 16);
    }

    if (!filled.get(147)) {
      mxpOffset2Lo.position(mxpOffset2Lo.position() + 16);
    }

    if (!filled.get(148)) {
      mxpSize1Hi.position(mxpSize1Hi.position() + 16);
    }

    if (!filled.get(149)) {
      mxpSize1Lo.position(mxpSize1Lo.position() + 16);
    }

    if (!filled.get(150)) {
      mxpSize2Hi.position(mxpSize2Hi.position() + 16);
    }

    if (!filled.get(151)) {
      mxpSize2Lo.position(mxpSize2Lo.position() + 16);
    }

    if (!filled.get(24)) {
      mxpStamp.position(mxpStamp.position() + 4);
    }

    if (!filled.get(152)) {
      mxpWords.position(mxpWords.position() + 16);
    }

    if (!filled.get(25)) {
      nonStackRows.position(nonStackRows.position() + 1);
    }

    if (!filled.get(129)) {
      nonce.position(nonce.position() + 8);
    }

    if (!filled.get(125)) {
      nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.position(nonceNewXorStpGasPaidOutOfPocketXorPrcCallerGasXorGasInitiallyAvailable.position() + 8);
    }

    if (!filled.get(124)) {
      nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.position(nonceXorStpGasMxpXorPrcCalleeGasXorBasefee.position() + 8);
    }

    if (!filled.get(153)) {
      oobData1.position(oobData1.position() + 16);
    }

    if (!filled.get(154)) {
      oobData2.position(oobData2.position() + 16);
    }

    if (!filled.get(155)) {
      oobData3.position(oobData3.position() + 16);
    }

    if (!filled.get(156)) {
      oobData4.position(oobData4.position() + 16);
    }

    if (!filled.get(157)) {
      oobData5.position(oobData5.position() + 16);
    }

    if (!filled.get(158)) {
      oobData6.position(oobData6.position() + 16);
    }

    if (!filled.get(159)) {
      oobData7.position(oobData7.position() + 16);
    }

    if (!filled.get(160)) {
      oobData8.position(oobData8.position() + 16);
    }

    if (!filled.get(161)) {
      oobData9.position(oobData9.position() + 16);
    }

    if (!filled.get(26)) {
      peekAtAccount.position(peekAtAccount.position() + 1);
    }

    if (!filled.get(27)) {
      peekAtContext.position(peekAtContext.position() + 1);
    }

    if (!filled.get(28)) {
      peekAtMiscellaneous.position(peekAtMiscellaneous.position() + 1);
    }

    if (!filled.get(29)) {
      peekAtScenario.position(peekAtScenario.position() + 1);
    }

    if (!filled.get(30)) {
      peekAtStack.position(peekAtStack.position() + 1);
    }

    if (!filled.get(31)) {
      peekAtStorage.position(peekAtStorage.position() + 1);
    }

    if (!filled.get(32)) {
      peekAtTransaction.position(peekAtTransaction.position() + 1);
    }

    if (!filled.get(69)) {
      prcEcmulXorMachineStateFlag.position(prcEcmulXorMachineStateFlag.position() + 1);
    }

    if (!filled.get(70)) {
      prcEcpairingXorMaxcsx.position(prcEcpairingXorMaxcsx.position() + 1);
    }

    if (!filled.get(71)) {
      prcEcrecoverXorModFlag.position(prcEcrecoverXorModFlag.position() + 1);
    }

    if (!filled.get(72)) {
      prcFailureKnownToHubXorMulFlag.position(prcFailureKnownToHubXorMulFlag.position() + 1);
    }

    if (!filled.get(73)) {
      prcFailureKnownToRamXorMxpx.position(prcFailureKnownToRamXorMxpx.position() + 1);
    }

    if (!filled.get(74)) {
      prcIdentityXorMxpFlag.position(prcIdentityXorMxpFlag.position() + 1);
    }

    if (!filled.get(75)) {
      prcModexpXorOogx.position(prcModexpXorOogx.position() + 1);
    }

    if (!filled.get(76)) {
      prcRipemd160XorOpcx.position(prcRipemd160XorOpcx.position() + 1);
    }

    if (!filled.get(77)) {
      prcSha2256XorPushpopFlag.position(prcSha2256XorPushpopFlag.position() + 1);
    }

    if (!filled.get(78)) {
      prcSuccessCallerWillRevertXorRdcx.position(prcSuccessCallerWillRevertXorRdcx.position() + 1);
    }

    if (!filled.get(79)) {
      prcSuccessCallerWontRevertXorShfFlag.position(prcSuccessCallerWontRevertXorShfFlag.position() + 1);
    }

    if (!filled.get(130)) {
      priorityFeePerGas.position(priorityFeePerGas.position() + 8);
    }

    if (!filled.get(33)) {
      programCounter.position(programCounter.position() + 4);
    }

    if (!filled.get(34)) {
      programCounterNew.position(programCounterNew.position() + 4);
    }

    if (!filled.get(35)) {
      refundCounter.position(refundCounter.position() + 4);
    }

    if (!filled.get(131)) {
      refundCounterInfinity.position(refundCounterInfinity.position() + 8);
    }

    if (!filled.get(36)) {
      refundCounterNew.position(refundCounterNew.position() + 4);
    }

    if (!filled.get(132)) {
      refundEffective.position(refundEffective.position() + 8);
    }

    if (!filled.get(37)) {
      relativeBlockNumber.position(relativeBlockNumber.position() + 2);
    }

    if (!filled.get(115)) {
      returnAtCapacityXorMxpInst.position(returnAtCapacityXorMxpInst.position() + 4);
    }

    if (!filled.get(116)) {
      returnAtOffsetXorOobInst.position(returnAtOffsetXorOobInst.position() + 4);
    }

    if (!filled.get(117)) {
      returnDataContextNumberXorStpGasStipend.position(returnDataContextNumberXorStpGasStipend.position() + 4);
    }

    if (!filled.get(118)) {
      returnDataOffsetXorStpInstruction.position(returnDataOffsetXorStpInstruction.position() + 4);
    }

    if (!filled.get(119)) {
      returnDataSize.position(returnDataSize.position() + 4);
    }

    if (!filled.get(80)) {
      returnExceptionXorSox.position(returnExceptionXorSox.position() + 1);
    }

    if (!filled.get(81)) {
      returnFromDeploymentEmptyCodeWillRevertXorSstorex.position(returnFromDeploymentEmptyCodeWillRevertXorSstorex.position() + 1);
    }

    if (!filled.get(82)) {
      returnFromDeploymentEmptyCodeWontRevertXorStackramFlag.position(returnFromDeploymentEmptyCodeWontRevertXorStackramFlag.position() + 1);
    }

    if (!filled.get(83)) {
      returnFromDeploymentNonemptyCodeWillRevertXorStackItemPop1.position(returnFromDeploymentNonemptyCodeWillRevertXorStackItemPop1.position() + 1);
    }

    if (!filled.get(84)) {
      returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop2.position(returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop2.position() + 1);
    }

    if (!filled.get(85)) {
      returnFromMessageCallWillTouchRamXorStackItemPop3.position(returnFromMessageCallWillTouchRamXorStackItemPop3.position() + 1);
    }

    if (!filled.get(86)) {
      returnFromMessageCallWontTouchRamXorStackItemPop4.position(returnFromMessageCallWontTouchRamXorStackItemPop4.position() + 1);
    }

    if (!filled.get(111)) {
      rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.position(rlpaddrDepAddrHiXorCallDataContextNumberXorMmuRefSize.position() + 4);
    }

    if (!filled.get(140)) {
      rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.position(rlpaddrDepAddrLoXorMmuSrcOffsetHiXorStackItemValueHi4XorValueOrigHi.position() + 16);
    }

    if (!filled.get(64)) {
      rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired.position(rlpaddrFlagXorCreateNonemptyInitCodeFailureWontRevertXorJumpDestinationVettingRequired.position() + 1);
    }

    if (!filled.get(141)) {
      rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.position(rlpaddrKecHiXorMmuSrcOffsetLoXorStackItemValueLo1XorValueOrigLo.position() + 16);
    }

    if (!filled.get(142)) {
      rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.position(rlpaddrKecLoXorMmuTgtOffsetLoXorStackItemValueLo2.position() + 16);
    }

    if (!filled.get(97)) {
      rlpaddrRecipe.position(rlpaddrRecipe.position() + 1);
    }

    if (!filled.get(143)) {
      rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.position(rlpaddrSaltHiXorMxpGasMxpXorStackItemValueLo3.position() + 16);
    }

    if (!filled.get(144)) {
      rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.position(rlpaddrSaltLoXorMxpOffset1HiXorStackItemValueLo4.position() + 16);
    }

    if (!filled.get(65)) {
      romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag.position(romlexFlagXorCreateNonemptyInitCodeSuccessWillRevertXorJumpFlag.position() + 1);
    }

    if (!filled.get(87)) {
      selfdestructExceptionXorStaticx.position(selfdestructExceptionXorStaticx.position() + 1);
    }

    if (!filled.get(88)) {
      selfdestructWillRevertXorStaticFlag.position(selfdestructWillRevertXorStaticFlag.position() + 1);
    }

    if (!filled.get(89)) {
      selfdestructWontRevertAlreadyMarkedXorStoFlag.position(selfdestructWontRevertAlreadyMarkedXorStoFlag.position() + 1);
    }

    if (!filled.get(90)) {
      selfdestructWontRevertNotYetMarkedXorSux.position(selfdestructWontRevertNotYetMarkedXorSux.position() + 1);
    }

    if (!filled.get(99)) {
      stackItemHeight2.position(stackItemHeight2.position() + 2);
    }

    if (!filled.get(100)) {
      stackItemHeight3.position(stackItemHeight3.position() + 2);
    }

    if (!filled.get(101)) {
      stackItemHeight4.position(stackItemHeight4.position() + 2);
    }

    if (!filled.get(120)) {
      stackItemStamp1.position(stackItemStamp1.position() + 5);
    }

    if (!filled.get(121)) {
      stackItemStamp2.position(stackItemStamp2.position() + 5);
    }

    if (!filled.get(122)) {
      stackItemStamp3.position(stackItemStamp3.position() + 5);
    }

    if (!filled.get(123)) {
      stackItemStamp4.position(stackItemStamp4.position() + 5);
    }

    if (!filled.get(162)) {
      stpGasHi.position(stpGasHi.position() + 16);
    }

    if (!filled.get(163)) {
      stpGasLo.position(stpGasLo.position() + 16);
    }

    if (!filled.get(126)) {
      stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.position(stpGasUpfrontGasCostXorPrcReturnGasXorGasLeftover.position() + 8);
    }

    if (!filled.get(164)) {
      stpValueHi.position(stpValueHi.position() + 16);
    }

    if (!filled.get(165)) {
      stpValueLo.position(stpValueLo.position() + 16);
    }

    if (!filled.get(38)) {
      subStamp.position(subStamp.position() + 4);
    }

    if (!filled.get(91)) {
      swapFlag.position(swapFlag.position() + 1);
    }

    if (!filled.get(66)) {
      trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag.position(trmFlagXorCreateNonemptyInitCodeSuccessWontRevertXorKecFlag.position() + 1);
    }

    if (!filled.get(145)) {
      trmRawAddressHiXorMxpOffset1Lo.position(trmRawAddressHiXorMxpOffset1Lo.position() + 16);
    }

    if (!filled.get(39)) {
      twoLineInstruction.position(twoLineInstruction.position() + 1);
    }

    if (!filled.get(40)) {
      txExec.position(txExec.position() + 1);
    }

    if (!filled.get(41)) {
      txFinl.position(txFinl.position() + 1);
    }

    if (!filled.get(42)) {
      txInit.position(txInit.position() + 1);
    }

    if (!filled.get(43)) {
      txSkip.position(txSkip.position() + 1);
    }

    if (!filled.get(44)) {
      txWarm.position(txWarm.position() + 1);
    }

    if (!filled.get(92)) {
      txnFlag.position(txnFlag.position() + 1);
    }

    if (!filled.get(68)) {
      warmthNewXorPrcEcaddXorLogInfoFlag.position(warmthNewXorPrcEcaddXorLogInfoFlag.position() + 1);
    }

    if (!filled.get(67)) {
      warmthXorPrcBlake2FXorLogFlag.position(warmthXorPrcBlake2FXorLogFlag.position() + 1);
    }

    if (!filled.get(93)) {
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
