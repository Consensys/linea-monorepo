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

import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absoluteTransactionNumber;
  private final MappedByteBuffer accFinal;
  private final MappedByteBuffer accFirst;
  private final MappedByteBuffer
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee;
  private final MappedByteBuffer
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum;
  private final MappedByteBuffer
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi;
  private final MappedByteBuffer
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize;
  private final MappedByteBuffer batchNumber;
  private final MappedByteBuffer callerContextNumber;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo;
  private final MappedByteBuffer
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo;
  private final MappedByteBuffer
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi;
  private final MappedByteBuffer
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover;
  private final MappedByteBuffer
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable;
  private final MappedByteBuffer
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice;
  private final MappedByteBuffer
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit;
  private final MappedByteBuffer conAgain;
  private final MappedByteBuffer conFirst;
  private final MappedByteBuffer contextGetsReverted;
  private final MappedByteBuffer contextMayChange;
  private final MappedByteBuffer contextNumber;
  private final MappedByteBuffer contextNumberNew;
  private final MappedByteBuffer contextRevertStamp;
  private final MappedByteBuffer contextSelfReverts;
  private final MappedByteBuffer contextWillRevert;
  private final MappedByteBuffer counterNsr;
  private final MappedByteBuffer counterTli;
  private final MappedByteBuffer createFailureConditionWillRevertXorHashInfoFlag;
  private final MappedByteBuffer createFailureConditionWontRevertXorIcpx;
  private final MappedByteBuffer createNonemptyInitCodeFailureWillRevertXorInvalidFlag;
  private final MappedByteBuffer createNonemptyInitCodeFailureWontRevertXorJumpx;
  private final MappedByteBuffer
      createNonemptyInitCodeSuccessWillRevertXorJumpDestinationVettingRequired;
  private final MappedByteBuffer createNonemptyInitCodeSuccessWontRevertXorJumpFlag;
  private final MappedByteBuffer
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize;
  private final MappedByteBuffer
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce;
  private final MappedByteBuffer
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance;
  private final MappedByteBuffer
      deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment;
  private final MappedByteBuffer
      deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2;
  private final MappedByteBuffer
      deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd;
  private final MappedByteBuffer domStamp;
  private final MappedByteBuffer exceptionAhoy;
  private final MappedByteBuffer
      existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode;
  private final MappedByteBuffer
      existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution;
  private final MappedByteBuffer expInstXorPrcCalleeGas;
  private final MappedByteBuffer gasActual;
  private final MappedByteBuffer gasCost;
  private final MappedByteBuffer gasExpected;
  private final MappedByteBuffer gasNext;
  private final MappedByteBuffer
      hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero;
  private final MappedByteBuffer
      hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero;
  private final MappedByteBuffer hashInfoStamp;
  private final MappedByteBuffer height;
  private final MappedByteBuffer heightNew;
  private final MappedByteBuffer hubStamp;
  private final MappedByteBuffer hubStampTransactionEnd;
  private final MappedByteBuffer
      isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth;
  private final MappedByteBuffer logInfoStamp;
  private final MappedByteBuffer
      markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2;
  private final MappedByteBuffer
      markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew;
  private final MappedByteBuffer mmuAuxIdXorPrcCallerGas;
  private final MappedByteBuffer mmuExoSumXorPrcCdo;
  private final MappedByteBuffer mmuInstXorPrcCds;
  private final MappedByteBuffer mmuLimb1;
  private final MappedByteBuffer mmuLimb2;
  private final MappedByteBuffer mmuPhaseXorPrcRac;
  private final MappedByteBuffer mmuRefOffsetXorPrcRao;
  private final MappedByteBuffer mmuRefSizeXorPrcReturnGas;
  private final MappedByteBuffer mmuSize;
  private final MappedByteBuffer mmuSrcId;
  private final MappedByteBuffer mmuSrcOffsetHi;
  private final MappedByteBuffer mmuSrcOffsetLo;
  private final MappedByteBuffer mmuStamp;
  private final MappedByteBuffer mmuTgtId;
  private final MappedByteBuffer mmuTgtOffsetLo;
  private final MappedByteBuffer mxpStamp;
  private final MappedByteBuffer nonStackRows;
  private final MappedByteBuffer
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity;
  private final MappedByteBuffer
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas;
  private final MappedByteBuffer oobData7XorStackItemValueLo3;
  private final MappedByteBuffer oobData8XorStackItemValueLo4;
  private final MappedByteBuffer oobInst;
  private final MappedByteBuffer peekAtAccount;
  private final MappedByteBuffer peekAtContext;
  private final MappedByteBuffer peekAtMiscellaneous;
  private final MappedByteBuffer peekAtScenario;
  private final MappedByteBuffer peekAtStack;
  private final MappedByteBuffer peekAtStorage;
  private final MappedByteBuffer peekAtTransaction;
  private final MappedByteBuffer prcBlake2FXorKecFlag;
  private final MappedByteBuffer prcEcaddXorLogFlag;
  private final MappedByteBuffer prcEcmulXorLogInfoFlag;
  private final MappedByteBuffer prcEcpairingXorMachineStateFlag;
  private final MappedByteBuffer prcEcrecoverXorMaxcsx;
  private final MappedByteBuffer prcFailureKnownToHubXorModFlag;
  private final MappedByteBuffer prcFailureKnownToRamXorMulFlag;
  private final MappedByteBuffer prcIdentityXorMxpx;
  private final MappedByteBuffer prcModexpXorMxpFlag;
  private final MappedByteBuffer prcRipemd160XorOogx;
  private final MappedByteBuffer prcSha2256XorOpcx;
  private final MappedByteBuffer prcSuccessWillRevertXorPushpopFlag;
  private final MappedByteBuffer prcSuccessWontRevertXorRdcx;
  private final MappedByteBuffer programCounter;
  private final MappedByteBuffer programCounterNew;
  private final MappedByteBuffer refundCounter;
  private final MappedByteBuffer refundCounterNew;
  private final MappedByteBuffer returnExceptionXorShfFlag;
  private final MappedByteBuffer returnFromDeploymentEmptyCodeWillRevertXorSox;
  private final MappedByteBuffer returnFromDeploymentEmptyCodeWontRevertXorSstorex;
  private final MappedByteBuffer returnFromDeploymentNonemptyCodeWillRevertXorStackramFlag;
  private final MappedByteBuffer returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop1;
  private final MappedByteBuffer returnFromMessageCallWillTouchRamXorStackItemPop2;
  private final MappedByteBuffer returnFromMessageCallWontTouchRamXorStackItemPop3;
  private final MappedByteBuffer
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective;
  private final MappedByteBuffer
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi;
  private final MappedByteBuffer rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3;
  private final MappedByteBuffer
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo;
  private final MappedByteBuffer
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue;
  private final MappedByteBuffer rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3;
  private final MappedByteBuffer rlpaddrSaltHiXorOobData4XorStackItemValueHi4;
  private final MappedByteBuffer rlpaddrSaltLoXorOobData5XorStackItemValueLo1;
  private final MappedByteBuffer romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4;
  private final MappedByteBuffer selfdestructExceptionXorStackItemPop4;
  private final MappedByteBuffer selfdestructWillRevertXorStaticx;
  private final MappedByteBuffer selfdestructWontRevertAlreadyMarkedXorStaticFlag;
  private final MappedByteBuffer selfdestructWontRevertNotYetMarkedXorStoFlag;
  private final MappedByteBuffer stoFinal;
  private final MappedByteBuffer stoFirst;
  private final MappedByteBuffer stpGasHiXorStaticGas;
  private final MappedByteBuffer stpGasLo;
  private final MappedByteBuffer stpGasMxp;
  private final MappedByteBuffer stpGasPaidOutOfPocket;
  private final MappedByteBuffer stpGasStipend;
  private final MappedByteBuffer stpGasUpfrontGasCost;
  private final MappedByteBuffer stpInstruction;
  private final MappedByteBuffer stpValHi;
  private final MappedByteBuffer stpValLo;
  private final MappedByteBuffer subStamp;
  private final MappedByteBuffer sux;
  private final MappedByteBuffer swapFlag;
  private final MappedByteBuffer transactionReverts;
  private final MappedByteBuffer trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag;
  private final MappedByteBuffer trmRawAddressHiXorOobData6XorStackItemValueLo2;
  private final MappedByteBuffer twoLineInstruction;
  private final MappedByteBuffer txExec;
  private final MappedByteBuffer txFinl;
  private final MappedByteBuffer txInit;
  private final MappedByteBuffer txSkip;
  private final MappedByteBuffer txWarm;
  private final MappedByteBuffer txnFlag;
  private final MappedByteBuffer warmthNewXorCreateExceptionXorHaltFlag;
  private final MappedByteBuffer warmthXorCreateEmptyInitCodeWontRevertXorExtFlag;
  private final MappedByteBuffer wcpFlag;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("hub.ABSOLUTE_TRANSACTION_NUMBER", 32, length),
        new ColumnHeader("hub.acc_FINAL", 1, length),
        new ColumnHeader("hub.acc_FIRST", 1, length),
        new ColumnHeader(
            "hub.ADDRESS_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_ALPHA_xor_ADDRESS_HI_xor_BASEFEE",
            32,
            length),
        new ColumnHeader(
            "hub.ADDRESS_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP_DATA_1_xor_DELTA_xor_ADDRESS_LO_xor_BATCH_NUM",
            32,
            length),
        new ColumnHeader(
            "hub.BALANCE_NEW_xor_BYTE_CODE_ADDRESS_HI_xor_EXP_DATA_3_xor_HASH_INFO_KECCAK_LO_xor_DEPLOYMENT_NUMBER_INFTY_xor_COINBASE_ADDRESS_HI",
            32,
            length),
        new ColumnHeader(
            "hub.BALANCE_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP_DATA_2_xor_HASH_INFO_KECCAK_HI_xor_DEPLOYMENT_NUMBER_xor_CALL_DATA_SIZE",
            32,
            length),
        new ColumnHeader("hub.BATCH_NUMBER", 32, length),
        new ColumnHeader("hub.CALLER_CONTEXT_NUMBER", 32, length),
        new ColumnHeader("hub.CODE_FRAGMENT_INDEX", 32, length),
        new ColumnHeader(
            "hub.CODE_FRAGMENT_INDEX_xor_BYTE_CODE_ADDRESS_LO_xor_EXP_DATA_4_xor_HASH_INFO_SIZE_xor_STORAGE_KEY_HI_xor_COINBASE_ADDRESS_LO",
            32,
            length),
        new ColumnHeader(
            "hub.CODE_HASH_HI_NEW_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MXP_GAS_MXP_xor_NB_ADDED_xor_VALUE_CURR_HI_xor_FROM_ADDRESS_LO",
            32,
            length),
        new ColumnHeader(
            "hub.CODE_HASH_HI_xor_BYTE_CODE_CODE_FRAGMENT_INDEX_xor_EXP_DATA_5_xor_INSTRUCTION_xor_STORAGE_KEY_LO_xor_FROM_ADDRESS_HI",
            32,
            length),
        new ColumnHeader(
            "hub.CODE_HASH_LO_NEW_xor_CALLER_ADDRESS_HI_xor_MXP_MTNTOP_xor_PUSH_VALUE_HI_xor_VALUE_NEXT_HI_xor_GAS_LEFTOVER",
            32,
            length),
        new ColumnHeader(
            "hub.CODE_HASH_LO_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MXP_INST_xor_NB_REMOVED_xor_VALUE_CURR_LO_xor_GAS_INITIALLY_AVAILABLE",
            32,
            length),
        new ColumnHeader(
            "hub.CODE_SIZE_NEW_xor_CALL_DATA_CONTEXT_NUMBER_xor_MXP_OFFSET_1_LO_xor_STACK_ITEM_HEIGHT_1_xor_VALUE_ORIG_HI_xor_GAS_PRICE",
            32,
            length),
        new ColumnHeader(
            "hub.CODE_SIZE_xor_CALLER_ADDRESS_LO_xor_MXP_OFFSET_1_HI_xor_PUSH_VALUE_LO_xor_VALUE_NEXT_LO_xor_GAS_LIMIT",
            32,
            length),
        new ColumnHeader("hub.con_AGAIN", 1, length),
        new ColumnHeader("hub.con_FIRST", 1, length),
        new ColumnHeader("hub.CONTEXT_GETS_REVERTED", 1, length),
        new ColumnHeader("hub.CONTEXT_MAY_CHANGE", 1, length),
        new ColumnHeader("hub.CONTEXT_NUMBER", 32, length),
        new ColumnHeader("hub.CONTEXT_NUMBER_NEW", 32, length),
        new ColumnHeader("hub.CONTEXT_REVERT_STAMP", 32, length),
        new ColumnHeader("hub.CONTEXT_SELF_REVERTS", 1, length),
        new ColumnHeader("hub.CONTEXT_WILL_REVERT", 1, length),
        new ColumnHeader("hub.COUNTER_NSR", 32, length),
        new ColumnHeader("hub.COUNTER_TLI", 1, length),
        new ColumnHeader("hub.CREATE_FAILURE_CONDITION_WILL_REVERT_xor_HASH_INFO_FLAG", 1, length),
        new ColumnHeader("hub.CREATE_FAILURE_CONDITION_WONT_REVERT_xor_ICPX", 1, length),
        new ColumnHeader(
            "hub.CREATE_NONEMPTY_INIT_CODE_FAILURE_WILL_REVERT_xor_INVALID_FLAG", 1, length),
        new ColumnHeader("hub.CREATE_NONEMPTY_INIT_CODE_FAILURE_WONT_REVERT_xor_JUMPX", 1, length),
        new ColumnHeader(
            "hub.CREATE_NONEMPTY_INIT_CODE_SUCCESS_WILL_REVERT_xor_JUMP_DESTINATION_VETTING_REQUIRED",
            1,
            length),
        new ColumnHeader(
            "hub.CREATE_NONEMPTY_INIT_CODE_SUCCESS_WONT_REVERT_xor_JUMP_FLAG", 1, length),
        new ColumnHeader(
            "hub.DEPLOYMENT_NUMBER_INFTY_xor_CALL_DATA_SIZE_xor_MXP_OFFSET_2_LO_xor_STACK_ITEM_HEIGHT_3_xor_INIT_CODE_SIZE",
            32,
            length),
        new ColumnHeader(
            "hub.DEPLOYMENT_NUMBER_NEW_xor_CALL_STACK_DEPTH_xor_MXP_SIZE_1_HI_xor_STACK_ITEM_HEIGHT_4_xor_NONCE",
            32,
            length),
        new ColumnHeader(
            "hub.DEPLOYMENT_NUMBER_xor_CALL_DATA_OFFSET_xor_MXP_OFFSET_2_HI_xor_STACK_ITEM_HEIGHT_2_xor_VALUE_ORIG_LO_xor_INITIAL_BALANCE",
            32,
            length),
        new ColumnHeader(
            "hub.DEPLOYMENT_STATUS_INFTY_xor_IS_STATIC_xor_EXP_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_ADD_FLAG_xor_VALUE_CURR_IS_ORIG_xor_IS_DEPLOYMENT",
            1,
            length),
        new ColumnHeader(
            "hub.DEPLOYMENT_STATUS_NEW_xor_UPDATE_xor_MMU_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BIN_FLAG_xor_VALUE_CURR_IS_ZERO_xor_IS_TYPE2",
            1,
            length),
        new ColumnHeader(
            "hub.DEPLOYMENT_STATUS_xor_IS_ROOT_xor_CCSR_FLAG_xor_CALL_ABORT_xor_ACC_FLAG_xor_VALUE_CURR_CHANGES_xor_COPY_TXCD",
            1,
            length),
        new ColumnHeader("hub.DOM_STAMP", 32, length),
        new ColumnHeader("hub.EXCEPTION_AHOY", 1, length),
        new ColumnHeader(
            "hub.EXISTS_NEW_xor_MXP_DEPLOYS_xor_CALL_PRC_FAILURE_xor_CALL_FLAG_xor_VALUE_NEXT_IS_ORIG_xor_STATUS_CODE",
            1,
            length),
        new ColumnHeader(
            "hub.EXISTS_xor_MMU_SUCCESS_BIT_xor_CALL_EXCEPTION_xor_BTC_FLAG_xor_VALUE_NEXT_IS_CURR_xor_REQUIRES_EVM_EXECUTION",
            1,
            length),
        new ColumnHeader("hub.EXP_INST_xor_PRC_CALLEE_GAS", 8, length),
        new ColumnHeader("hub.GAS_ACTUAL", 32, length),
        new ColumnHeader("hub.GAS_COST", 32, length),
        new ColumnHeader("hub.GAS_EXPECTED", 32, length),
        new ColumnHeader("hub.GAS_NEXT", 32, length),
        new ColumnHeader(
            "hub.HAS_CODE_NEW_xor_MXP_MXPX_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_COPY_FLAG_xor_VALUE_ORIG_IS_ZERO",
            1,
            length),
        new ColumnHeader(
            "hub.HAS_CODE_xor_MXP_FLAG_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_CON_FLAG_xor_VALUE_NEXT_IS_ZERO",
            1,
            length),
        new ColumnHeader("hub.HASH_INFO_STAMP", 32, length),
        new ColumnHeader("hub.HEIGHT", 32, length),
        new ColumnHeader("hub.HEIGHT_NEW", 32, length),
        new ColumnHeader("hub.HUB_STAMP", 32, length),
        new ColumnHeader("hub.HUB_STAMP_TRANSACTION_END", 32, length),
        new ColumnHeader(
            "hub.IS_PRECOMPILE_xor_OOB_FLAG_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_CREATE_FLAG_xor_WARMTH",
            1,
            length),
        new ColumnHeader("hub.LOG_INFO_STAMP", 32, length),
        new ColumnHeader(
            "hub.MARKED_FOR_SELFDESTRUCT_NEW_xor_STP_FLAG_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DEC_FLAG_2",
            1,
            length),
        new ColumnHeader(
            "hub.MARKED_FOR_SELFDESTRUCT_xor_STP_EXISTS_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DEC_FLAG_1_xor_WARMTH_NEW",
            1,
            length),
        new ColumnHeader("hub.MMU_AUX_ID_xor_PRC_CALLER_GAS", 8, length),
        new ColumnHeader("hub.MMU_EXO_SUM_xor_PRC_CDO", 8, length),
        new ColumnHeader("hub.MMU_INST_xor_PRC_CDS", 8, length),
        new ColumnHeader("hub.MMU_LIMB_1", 32, length),
        new ColumnHeader("hub.MMU_LIMB_2", 32, length),
        new ColumnHeader("hub.MMU_PHASE_xor_PRC_RAC", 8, length),
        new ColumnHeader("hub.MMU_REF_OFFSET_xor_PRC_RAO", 8, length),
        new ColumnHeader("hub.MMU_REF_SIZE_xor_PRC_RETURN_GAS", 8, length),
        new ColumnHeader("hub.MMU_SIZE", 8, length),
        new ColumnHeader("hub.MMU_SRC_ID", 8, length),
        new ColumnHeader("hub.MMU_SRC_OFFSET_HI", 32, length),
        new ColumnHeader("hub.MMU_SRC_OFFSET_LO", 32, length),
        new ColumnHeader("hub.MMU_STAMP", 32, length),
        new ColumnHeader("hub.MMU_TGT_ID", 8, length),
        new ColumnHeader("hub.MMU_TGT_OFFSET_LO", 32, length),
        new ColumnHeader("hub.MXP_STAMP", 32, length),
        new ColumnHeader("hub.NON_STACK_ROWS", 32, length),
        new ColumnHeader(
            "hub.NONCE_NEW_xor_CONTEXT_NUMBER_xor_MXP_SIZE_2_HI_xor_STACK_ITEM_STAMP_2_xor_REFUND_COUNTER_INFINITY",
            32,
            length),
        new ColumnHeader(
            "hub.NONCE_xor_CALL_VALUE_xor_MXP_SIZE_1_LO_xor_STACK_ITEM_STAMP_1_xor_PRIORITY_FEE_PER_GAS",
            32,
            length),
        new ColumnHeader("hub.OOB_DATA_7_xor_STACK_ITEM_VALUE_LO_3", 32, length),
        new ColumnHeader("hub.OOB_DATA_8_xor_STACK_ITEM_VALUE_LO_4", 32, length),
        new ColumnHeader("hub.OOB_INST", 8, length),
        new ColumnHeader("hub.PEEK_AT_ACCOUNT", 1, length),
        new ColumnHeader("hub.PEEK_AT_CONTEXT", 1, length),
        new ColumnHeader("hub.PEEK_AT_MISCELLANEOUS", 1, length),
        new ColumnHeader("hub.PEEK_AT_SCENARIO", 1, length),
        new ColumnHeader("hub.PEEK_AT_STACK", 1, length),
        new ColumnHeader("hub.PEEK_AT_STORAGE", 1, length),
        new ColumnHeader("hub.PEEK_AT_TRANSACTION", 1, length),
        new ColumnHeader("hub.PRC_BLAKE2f_xor_KEC_FLAG", 1, length),
        new ColumnHeader("hub.PRC_ECADD_xor_LOG_FLAG", 1, length),
        new ColumnHeader("hub.PRC_ECMUL_xor_LOG_INFO_FLAG", 1, length),
        new ColumnHeader("hub.PRC_ECPAIRING_xor_MACHINE_STATE_FLAG", 1, length),
        new ColumnHeader("hub.PRC_ECRECOVER_xor_MAXCSX", 1, length),
        new ColumnHeader("hub.PRC_FAILURE_KNOWN_TO_HUB_xor_MOD_FLAG", 1, length),
        new ColumnHeader("hub.PRC_FAILURE_KNOWN_TO_RAM_xor_MUL_FLAG", 1, length),
        new ColumnHeader("hub.PRC_IDENTITY_xor_MXPX", 1, length),
        new ColumnHeader("hub.PRC_MODEXP_xor_MXP_FLAG", 1, length),
        new ColumnHeader("hub.PRC_RIPEMD-160_xor_OOGX", 1, length),
        new ColumnHeader("hub.PRC_SHA2-256_xor_OPCX", 1, length),
        new ColumnHeader("hub.PRC_SUCCESS_WILL_REVERT_xor_PUSHPOP_FLAG", 1, length),
        new ColumnHeader("hub.PRC_SUCCESS_WONT_REVERT_xor_RDCX", 1, length),
        new ColumnHeader("hub.PROGRAM_COUNTER", 32, length),
        new ColumnHeader("hub.PROGRAM_COUNTER_NEW", 32, length),
        new ColumnHeader("hub.REFUND_COUNTER", 32, length),
        new ColumnHeader("hub.REFUND_COUNTER_NEW", 32, length),
        new ColumnHeader("hub.RETURN_EXCEPTION_xor_SHF_FLAG", 1, length),
        new ColumnHeader("hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT_xor_SOX", 1, length),
        new ColumnHeader(
            "hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT_xor_SSTOREX", 1, length),
        new ColumnHeader(
            "hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT_xor_STACKRAM_FLAG", 1, length),
        new ColumnHeader(
            "hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT_xor_STACK_ITEM_POP_1", 1, length),
        new ColumnHeader(
            "hub.RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM_xor_STACK_ITEM_POP_2", 1, length),
        new ColumnHeader(
            "hub.RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM_xor_STACK_ITEM_POP_3", 1, length),
        new ColumnHeader(
            "hub.RLPADDR_DEP_ADDR_HI_xor_RETURN_AT_CAPACITY_xor_MXP_SIZE_2_LO_xor_STACK_ITEM_STAMP_3_xor_REFUND_EFFECTIVE",
            32,
            length),
        new ColumnHeader(
            "hub.RLPADDR_DEP_ADDR_LO_xor_RETURN_AT_OFFSET_xor_MXP_WORDS_xor_STACK_ITEM_STAMP_4_xor_TO_ADDRESS_HI",
            32,
            length),
        new ColumnHeader(
            "hub.RLPADDR_FLAG_xor_STP_OOGX_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DEC_FLAG_3",
            1,
            length),
        new ColumnHeader(
            "hub.RLPADDR_KEC_HI_xor_RETURN_DATA_CONTEXT_NUMBER_xor_OOB_DATA_1_xor_STACK_ITEM_VALUE_HI_1_xor_TO_ADDRESS_LO",
            32,
            length),
        new ColumnHeader(
            "hub.RLPADDR_KEC_LO_xor_RETURN_DATA_OFFSET_xor_OOB_DATA_2_xor_STACK_ITEM_VALUE_HI_2_xor_VALUE",
            32,
            length),
        new ColumnHeader(
            "hub.RLPADDR_RECIPE_xor_RETURN_DATA_SIZE_xor_OOB_DATA_3_xor_STACK_ITEM_VALUE_HI_3",
            32,
            length),
        new ColumnHeader(
            "hub.RLPADDR_SALT_HI_xor_OOB_DATA_4_xor_STACK_ITEM_VALUE_HI_4", 32, length),
        new ColumnHeader(
            "hub.RLPADDR_SALT_LO_xor_OOB_DATA_5_xor_STACK_ITEM_VALUE_LO_1", 32, length),
        new ColumnHeader(
            "hub.ROM_LEX_FLAG_xor_STP_WARMTH_xor_CREATE_ABORT_xor_DEC_FLAG_4", 1, length),
        new ColumnHeader("hub.SELFDESTRUCT_EXCEPTION_xor_STACK_ITEM_POP_4", 1, length),
        new ColumnHeader("hub.SELFDESTRUCT_WILL_REVERT_xor_STATICX", 1, length),
        new ColumnHeader("hub.SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED_xor_STATIC_FLAG", 1, length),
        new ColumnHeader("hub.SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED_xor_STO_FLAG", 1, length),
        new ColumnHeader("hub.sto_FINAL", 1, length),
        new ColumnHeader("hub.sto_FIRST", 1, length),
        new ColumnHeader("hub.STP_GAS_HI_xor_STATIC_GAS", 32, length),
        new ColumnHeader("hub.STP_GAS_LO", 32, length),
        new ColumnHeader("hub.STP_GAS_MXP", 32, length),
        new ColumnHeader("hub.STP_GAS_PAID_OUT_OF_POCKET", 32, length),
        new ColumnHeader("hub.STP_GAS_STIPEND", 32, length),
        new ColumnHeader("hub.STP_GAS_UPFRONT_GAS_COST", 32, length),
        new ColumnHeader("hub.STP_INSTRUCTION", 32, length),
        new ColumnHeader("hub.STP_VAL_HI", 32, length),
        new ColumnHeader("hub.STP_VAL_LO", 32, length),
        new ColumnHeader("hub.SUB_STAMP", 32, length),
        new ColumnHeader("hub.SUX", 1, length),
        new ColumnHeader("hub.SWAP_FLAG", 1, length),
        new ColumnHeader("hub.TRANSACTION_REVERTS", 1, length),
        new ColumnHeader(
            "hub.TRM_FLAG_xor_CREATE_EMPTY_INIT_CODE_WILL_REVERT_xor_DUP_FLAG", 1, length),
        new ColumnHeader(
            "hub.TRM_RAW_ADDRESS_HI_xor_OOB_DATA_6_xor_STACK_ITEM_VALUE_LO_2", 32, length),
        new ColumnHeader("hub.TWO_LINE_INSTRUCTION", 1, length),
        new ColumnHeader("hub.TX_EXEC", 1, length),
        new ColumnHeader("hub.TX_FINL", 1, length),
        new ColumnHeader("hub.TX_INIT", 1, length),
        new ColumnHeader("hub.TX_SKIP", 1, length),
        new ColumnHeader("hub.TX_WARM", 1, length),
        new ColumnHeader("hub.TXN_FLAG", 1, length),
        new ColumnHeader("hub.WARMTH_NEW_xor_CREATE_EXCEPTION_xor_HALT_FLAG", 1, length),
        new ColumnHeader(
            "hub.WARMTH_xor_CREATE_EMPTY_INIT_CODE_WONT_REVERT_xor_EXT_FLAG", 1, length),
        new ColumnHeader("hub.WCP_FLAG", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absoluteTransactionNumber = buffers.get(0);
    this.accFinal = buffers.get(1);
    this.accFirst = buffers.get(2);
    this.addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee = buffers.get(3);
    this.addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum = buffers.get(4);
    this
            .balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi =
        buffers.get(5);
    this
            .balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize =
        buffers.get(6);
    this.batchNumber = buffers.get(7);
    this.callerContextNumber = buffers.get(8);
    this.codeFragmentIndex = buffers.get(9);
    this
            .codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo =
        buffers.get(10);
    this
            .codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo =
        buffers.get(11);
    this
            .codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi =
        buffers.get(12);
    this.codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover =
        buffers.get(13);
    this
            .codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable =
        buffers.get(14);
    this
            .codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice =
        buffers.get(15);
    this.codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit =
        buffers.get(16);
    this.conAgain = buffers.get(17);
    this.conFirst = buffers.get(18);
    this.contextGetsReverted = buffers.get(19);
    this.contextMayChange = buffers.get(20);
    this.contextNumber = buffers.get(21);
    this.contextNumberNew = buffers.get(22);
    this.contextRevertStamp = buffers.get(23);
    this.contextSelfReverts = buffers.get(24);
    this.contextWillRevert = buffers.get(25);
    this.counterNsr = buffers.get(26);
    this.counterTli = buffers.get(27);
    this.createFailureConditionWillRevertXorHashInfoFlag = buffers.get(28);
    this.createFailureConditionWontRevertXorIcpx = buffers.get(29);
    this.createNonemptyInitCodeFailureWillRevertXorInvalidFlag = buffers.get(30);
    this.createNonemptyInitCodeFailureWontRevertXorJumpx = buffers.get(31);
    this.createNonemptyInitCodeSuccessWillRevertXorJumpDestinationVettingRequired = buffers.get(32);
    this.createNonemptyInitCodeSuccessWontRevertXorJumpFlag = buffers.get(33);
    this.deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize =
        buffers.get(34);
    this.deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce =
        buffers.get(35);
    this
            .deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance =
        buffers.get(36);
    this
            .deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment =
        buffers.get(37);
    this
            .deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2 =
        buffers.get(38);
    this.deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd =
        buffers.get(39);
    this.domStamp = buffers.get(40);
    this.exceptionAhoy = buffers.get(41);
    this.existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode =
        buffers.get(42);
    this.existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution =
        buffers.get(43);
    this.expInstXorPrcCalleeGas = buffers.get(44);
    this.gasActual = buffers.get(45);
    this.gasCost = buffers.get(46);
    this.gasExpected = buffers.get(47);
    this.gasNext = buffers.get(48);
    this.hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero =
        buffers.get(49);
    this.hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero =
        buffers.get(50);
    this.hashInfoStamp = buffers.get(51);
    this.height = buffers.get(52);
    this.heightNew = buffers.get(53);
    this.hubStamp = buffers.get(54);
    this.hubStampTransactionEnd = buffers.get(55);
    this.isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth =
        buffers.get(56);
    this.logInfoStamp = buffers.get(57);
    this.markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2 =
        buffers.get(58);
    this.markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew =
        buffers.get(59);
    this.mmuAuxIdXorPrcCallerGas = buffers.get(60);
    this.mmuExoSumXorPrcCdo = buffers.get(61);
    this.mmuInstXorPrcCds = buffers.get(62);
    this.mmuLimb1 = buffers.get(63);
    this.mmuLimb2 = buffers.get(64);
    this.mmuPhaseXorPrcRac = buffers.get(65);
    this.mmuRefOffsetXorPrcRao = buffers.get(66);
    this.mmuRefSizeXorPrcReturnGas = buffers.get(67);
    this.mmuSize = buffers.get(68);
    this.mmuSrcId = buffers.get(69);
    this.mmuSrcOffsetHi = buffers.get(70);
    this.mmuSrcOffsetLo = buffers.get(71);
    this.mmuStamp = buffers.get(72);
    this.mmuTgtId = buffers.get(73);
    this.mmuTgtOffsetLo = buffers.get(74);
    this.mxpStamp = buffers.get(75);
    this.nonStackRows = buffers.get(76);
    this.nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity =
        buffers.get(77);
    this.nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas = buffers.get(78);
    this.oobData7XorStackItemValueLo3 = buffers.get(79);
    this.oobData8XorStackItemValueLo4 = buffers.get(80);
    this.oobInst = buffers.get(81);
    this.peekAtAccount = buffers.get(82);
    this.peekAtContext = buffers.get(83);
    this.peekAtMiscellaneous = buffers.get(84);
    this.peekAtScenario = buffers.get(85);
    this.peekAtStack = buffers.get(86);
    this.peekAtStorage = buffers.get(87);
    this.peekAtTransaction = buffers.get(88);
    this.prcBlake2FXorKecFlag = buffers.get(89);
    this.prcEcaddXorLogFlag = buffers.get(90);
    this.prcEcmulXorLogInfoFlag = buffers.get(91);
    this.prcEcpairingXorMachineStateFlag = buffers.get(92);
    this.prcEcrecoverXorMaxcsx = buffers.get(93);
    this.prcFailureKnownToHubXorModFlag = buffers.get(94);
    this.prcFailureKnownToRamXorMulFlag = buffers.get(95);
    this.prcIdentityXorMxpx = buffers.get(96);
    this.prcModexpXorMxpFlag = buffers.get(97);
    this.prcRipemd160XorOogx = buffers.get(98);
    this.prcSha2256XorOpcx = buffers.get(99);
    this.prcSuccessWillRevertXorPushpopFlag = buffers.get(100);
    this.prcSuccessWontRevertXorRdcx = buffers.get(101);
    this.programCounter = buffers.get(102);
    this.programCounterNew = buffers.get(103);
    this.refundCounter = buffers.get(104);
    this.refundCounterNew = buffers.get(105);
    this.returnExceptionXorShfFlag = buffers.get(106);
    this.returnFromDeploymentEmptyCodeWillRevertXorSox = buffers.get(107);
    this.returnFromDeploymentEmptyCodeWontRevertXorSstorex = buffers.get(108);
    this.returnFromDeploymentNonemptyCodeWillRevertXorStackramFlag = buffers.get(109);
    this.returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop1 = buffers.get(110);
    this.returnFromMessageCallWillTouchRamXorStackItemPop2 = buffers.get(111);
    this.returnFromMessageCallWontTouchRamXorStackItemPop3 = buffers.get(112);
    this.rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective =
        buffers.get(113);
    this.rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi =
        buffers.get(114);
    this.rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3 = buffers.get(115);
    this.rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo =
        buffers.get(116);
    this.rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue = buffers.get(117);
    this.rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3 = buffers.get(118);
    this.rlpaddrSaltHiXorOobData4XorStackItemValueHi4 = buffers.get(119);
    this.rlpaddrSaltLoXorOobData5XorStackItemValueLo1 = buffers.get(120);
    this.romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4 = buffers.get(121);
    this.selfdestructExceptionXorStackItemPop4 = buffers.get(122);
    this.selfdestructWillRevertXorStaticx = buffers.get(123);
    this.selfdestructWontRevertAlreadyMarkedXorStaticFlag = buffers.get(124);
    this.selfdestructWontRevertNotYetMarkedXorStoFlag = buffers.get(125);
    this.stoFinal = buffers.get(126);
    this.stoFirst = buffers.get(127);
    this.stpGasHiXorStaticGas = buffers.get(128);
    this.stpGasLo = buffers.get(129);
    this.stpGasMxp = buffers.get(130);
    this.stpGasPaidOutOfPocket = buffers.get(131);
    this.stpGasStipend = buffers.get(132);
    this.stpGasUpfrontGasCost = buffers.get(133);
    this.stpInstruction = buffers.get(134);
    this.stpValHi = buffers.get(135);
    this.stpValLo = buffers.get(136);
    this.subStamp = buffers.get(137);
    this.sux = buffers.get(138);
    this.swapFlag = buffers.get(139);
    this.transactionReverts = buffers.get(140);
    this.trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag = buffers.get(141);
    this.trmRawAddressHiXorOobData6XorStackItemValueLo2 = buffers.get(142);
    this.twoLineInstruction = buffers.get(143);
    this.txExec = buffers.get(144);
    this.txFinl = buffers.get(145);
    this.txInit = buffers.get(146);
    this.txSkip = buffers.get(147);
    this.txWarm = buffers.get(148);
    this.txnFlag = buffers.get(149);
    this.warmthNewXorCreateExceptionXorHaltFlag = buffers.get(150);
    this.warmthXorCreateEmptyInitCodeWontRevertXorExtFlag = buffers.get(151);
    this.wcpFlag = buffers.get(152);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absoluteTransactionNumber(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("hub.ABSOLUTE_TRANSACTION_NUMBER already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absoluteTransactionNumber.put((byte) 0);
    }
    absoluteTransactionNumber.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accFinal(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("hub.acc_FINAL already set");
    } else {
      filled.set(47);
    }

    accFinal.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace accFirst(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("hub.acc_FIRST already set");
    } else {
      filled.set(48);
    }

    accFirst.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace batchNumber(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("hub.BATCH_NUMBER already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      batchNumber.put((byte) 0);
    }
    batchNumber.put(b.toArrayUnsafe());

    return this;
  }

  public Trace callerContextNumber(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("hub.CALLER_CONTEXT_NUMBER already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      callerContextNumber.put((byte) 0);
    }
    callerContextNumber.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndex(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("hub.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndex.put((byte) 0);
    }
    codeFragmentIndex.put(b.toArrayUnsafe());

    return this;
  }

  public Trace conAgain(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("hub.con_AGAIN already set");
    } else {
      filled.set(49);
    }

    conAgain.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace conFirst(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("hub.con_FIRST already set");
    } else {
      filled.set(50);
    }

    conFirst.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextGetsReverted(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("hub.CONTEXT_GETS_REVERTED already set");
    } else {
      filled.set(4);
    }

    contextGetsReverted.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextMayChange(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("hub.CONTEXT_MAY_CHANGE already set");
    } else {
      filled.set(5);
    }

    contextMayChange.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextNumber(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      contextNumber.put((byte) 0);
    }
    contextNumber.put(b.toArrayUnsafe());

    return this;
  }

  public Trace contextNumberNew(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER_NEW already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      contextNumberNew.put((byte) 0);
    }
    contextNumberNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace contextRevertStamp(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("hub.CONTEXT_REVERT_STAMP already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      contextRevertStamp.put((byte) 0);
    }
    contextRevertStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace contextSelfReverts(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("hub.CONTEXT_SELF_REVERTS already set");
    } else {
      filled.set(9);
    }

    contextSelfReverts.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace contextWillRevert(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("hub.CONTEXT_WILL_REVERT already set");
    } else {
      filled.set(10);
    }

    contextWillRevert.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace counterNsr(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("hub.COUNTER_NSR already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counterNsr.put((byte) 0);
    }
    counterNsr.put(b.toArrayUnsafe());

    return this;
  }

  public Trace counterTli(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("hub.COUNTER_TLI already set");
    } else {
      filled.set(12);
    }

    counterTli.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace domStamp(final Bytes b) {
    if (filled.get(13)) {
      throw new IllegalStateException("hub.DOM_STAMP already set");
    } else {
      filled.set(13);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      domStamp.put((byte) 0);
    }
    domStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace exceptionAhoy(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("hub.EXCEPTION_AHOY already set");
    } else {
      filled.set(14);
    }

    exceptionAhoy.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace gasActual(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("hub.GAS_ACTUAL already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasActual.put((byte) 0);
    }
    gasActual.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasCost(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("hub.GAS_COST already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasCost.put((byte) 0);
    }
    gasCost.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasExpected(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("hub.GAS_EXPECTED already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasExpected.put((byte) 0);
    }
    gasExpected.put(b.toArrayUnsafe());

    return this;
  }

  public Trace gasNext(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("hub.GAS_NEXT already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      gasNext.put((byte) 0);
    }
    gasNext.put(b.toArrayUnsafe());

    return this;
  }

  public Trace hashInfoStamp(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("hub.HASH_INFO_STAMP already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      hashInfoStamp.put((byte) 0);
    }
    hashInfoStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace height(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("hub.HEIGHT already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      height.put((byte) 0);
    }
    height.put(b.toArrayUnsafe());

    return this;
  }

  public Trace heightNew(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("hub.HEIGHT_NEW already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      heightNew.put((byte) 0);
    }
    heightNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace hubStamp(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("hub.HUB_STAMP already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      hubStamp.put((byte) 0);
    }
    hubStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace hubStampTransactionEnd(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("hub.HUB_STAMP_TRANSACTION_END already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      hubStampTransactionEnd.put((byte) 0);
    }
    hubStampTransactionEnd.put(b.toArrayUnsafe());

    return this;
  }

  public Trace logInfoStamp(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("hub.LOG_INFO_STAMP already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      logInfoStamp.put((byte) 0);
    }
    logInfoStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace mmuStamp(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("hub.MMU_STAMP already set");
    } else {
      filled.set(25);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmuStamp.put((byte) 0);
    }
    mmuStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace mxpStamp(final Bytes b) {
    if (filled.get(26)) {
      throw new IllegalStateException("hub.MXP_STAMP already set");
    } else {
      filled.set(26);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mxpStamp.put((byte) 0);
    }
    mxpStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nonStackRows(final Bytes b) {
    if (filled.get(27)) {
      throw new IllegalStateException("hub.NON_STACK_ROWS already set");
    } else {
      filled.set(27);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonStackRows.put((byte) 0);
    }
    nonStackRows.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountAddressHi(final Bytes b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.account/ADDRESS_HI already set");
    } else {
      filled.set(118);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put((byte) 0);
    }
    addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountAddressLo(final Bytes b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.account/ADDRESS_LO already set");
    } else {
      filled.set(119);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put((byte) 0);
    }
    addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountBalance(final Bytes b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.account/BALANCE already set");
    } else {
      filled.set(120);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .put((byte) 0);
    }
    balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountBalanceNew(final Bytes b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.account/BALANCE_NEW already set");
    } else {
      filled.set(121);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .put((byte) 0);
    }
    balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeFragmentIndex(final Bytes b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.account/CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(122);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .put((byte) 0);
    }
    codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeHashHi(final Bytes b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.account/CODE_HASH_HI already set");
    } else {
      filled.set(123);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .put((byte) 0);
    }
    codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeHashHiNew(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.account/CODE_HASH_HI_NEW already set");
    } else {
      filled.set(124);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .put((byte) 0);
    }
    codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeHashLo(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.account/CODE_HASH_LO already set");
    } else {
      filled.set(125);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .put((byte) 0);
    }
    codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeHashLoNew(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.account/CODE_HASH_LO_NEW already set");
    } else {
      filled.set(126);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
          (byte) 0);
    }
    codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeSize(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.account/CODE_SIZE already set");
    } else {
      filled.set(127);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
          (byte) 0);
    }
    codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountCodeSizeNew(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.account/CODE_SIZE_NEW already set");
    } else {
      filled.set(128);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .put((byte) 0);
    }
    codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountDeploymentNumber(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(129);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .put((byte) 0);
    }
    deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountDeploymentNumberInfty(final Bytes b) {
    if (filled.get(130)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER_INFTY already set");
    } else {
      filled.set(130);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
          (byte) 0);
    }
    deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountDeploymentNumberNew(final Bytes b) {
    if (filled.get(131)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_NUMBER_NEW already set");
    } else {
      filled.set(131);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put((byte) 0);
    }
    deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountDeploymentStatus(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_STATUS already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountDeploymentStatusInfty(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_STATUS_INFTY already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountDeploymentStatusNew(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.account/DEPLOYMENT_STATUS_NEW already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountExists(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.account/EXISTS already set");
    } else {
      filled.set(56);
    }

    existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountExistsNew(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.account/EXISTS_NEW already set");
    } else {
      filled.set(57);
    }

    existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountHasCode(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.account/HAS_CODE already set");
    } else {
      filled.set(58);
    }

    hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountHasCodeNew(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.account/HAS_CODE_NEW already set");
    } else {
      filled.set(59);
    }

    hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountIsPrecompile(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.account/IS_PRECOMPILE already set");
    } else {
      filled.set(60);
    }

    isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountMarkedForSelfdestruct(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.account/MARKED_FOR_SELFDESTRUCT already set");
    } else {
      filled.set(61);
    }

    markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountMarkedForSelfdestructNew(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.account/MARKED_FOR_SELFDESTRUCT_NEW already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountNonce(final Bytes b) {
    if (filled.get(132)) {
      throw new IllegalStateException("hub.account/NONCE already set");
    } else {
      filled.set(132);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put((byte) 0);
    }
    nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountNonceNew(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.account/NONCE_NEW already set");
    } else {
      filled.set(133);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put((byte) 0);
    }
    nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrDepAddrHi(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.account/RLPADDR_DEP_ADDR_HI already set");
    } else {
      filled.set(134);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
          (byte) 0);
    }
    rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrDepAddrLo(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.account/RLPADDR_DEP_ADDR_LO already set");
    } else {
      filled.set(135);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put((byte) 0);
    }
    rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrFlag(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("hub.account/RLPADDR_FLAG already set");
    } else {
      filled.set(63);
    }

    rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountRlpaddrKecHi(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.account/RLPADDR_KEC_HI already set");
    } else {
      filled.set(136);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
          (byte) 0);
    }
    rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrKecLo(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.account/RLPADDR_KEC_LO already set");
    } else {
      filled.set(137);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put((byte) 0);
    }
    rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrRecipe(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.account/RLPADDR_RECIPE already set");
    } else {
      filled.set(138);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put((byte) 0);
    }
    rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrSaltHi(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.account/RLPADDR_SALT_HI already set");
    } else {
      filled.set(139);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrSaltHiXorOobData4XorStackItemValueHi4.put((byte) 0);
    }
    rlpaddrSaltHiXorOobData4XorStackItemValueHi4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRlpaddrSaltLo(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.account/RLPADDR_SALT_LO already set");
    } else {
      filled.set(140);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrSaltLoXorOobData5XorStackItemValueLo1.put((byte) 0);
    }
    rlpaddrSaltLoXorOobData5XorStackItemValueLo1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountRomLexFlag(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.account/ROM_LEX_FLAG already set");
    } else {
      filled.set(64);
    }

    romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountTrmFlag(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException("hub.account/TRM_FLAG already set");
    } else {
      filled.set(65);
    }

    trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountTrmRawAddressHi(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.account/TRM_RAW_ADDRESS_HI already set");
    } else {
      filled.set(141);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trmRawAddressHiXorOobData6XorStackItemValueLo2.put((byte) 0);
    }
    trmRawAddressHiXorOobData6XorStackItemValueLo2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pAccountWarmth(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException("hub.account/WARMTH already set");
    } else {
      filled.set(66);
    }

    warmthXorCreateEmptyInitCodeWontRevertXorExtFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pAccountWarmthNew(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("hub.account/WARMTH_NEW already set");
    } else {
      filled.set(67);
    }

    warmthNewXorCreateExceptionXorHaltFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pContextAccountAddressHi(final Bytes b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.context/ACCOUNT_ADDRESS_HI already set");
    } else {
      filled.set(118);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put((byte) 0);
    }
    addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextAccountAddressLo(final Bytes b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.context/ACCOUNT_ADDRESS_LO already set");
    } else {
      filled.set(119);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put((byte) 0);
    }
    addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextAccountDeploymentNumber(final Bytes b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.context/ACCOUNT_DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(120);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .put((byte) 0);
    }
    balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextByteCodeAddressHi(final Bytes b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_ADDRESS_HI already set");
    } else {
      filled.set(121);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .put((byte) 0);
    }
    balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextByteCodeAddressLo(final Bytes b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_ADDRESS_LO already set");
    } else {
      filled.set(122);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .put((byte) 0);
    }
    codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextByteCodeCodeFragmentIndex(final Bytes b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(123);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .put((byte) 0);
    }
    codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextByteCodeDeploymentNumber(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(124);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .put((byte) 0);
    }
    codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextByteCodeDeploymentStatus(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.context/BYTE_CODE_DEPLOYMENT_STATUS already set");
    } else {
      filled.set(125);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .put((byte) 0);
    }
    codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallDataContextNumber(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.context/CALL_DATA_CONTEXT_NUMBER already set");
    } else {
      filled.set(128);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .put((byte) 0);
    }
    codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallDataOffset(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.context/CALL_DATA_OFFSET already set");
    } else {
      filled.set(129);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .put((byte) 0);
    }
    deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallDataSize(final Bytes b) {
    if (filled.get(130)) {
      throw new IllegalStateException("hub.context/CALL_DATA_SIZE already set");
    } else {
      filled.set(130);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
          (byte) 0);
    }
    deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallStackDepth(final Bytes b) {
    if (filled.get(131)) {
      throw new IllegalStateException("hub.context/CALL_STACK_DEPTH already set");
    } else {
      filled.set(131);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put((byte) 0);
    }
    deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallValue(final Bytes b) {
    if (filled.get(132)) {
      throw new IllegalStateException("hub.context/CALL_VALUE already set");
    } else {
      filled.set(132);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put((byte) 0);
    }
    nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallerAddressHi(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.context/CALLER_ADDRESS_HI already set");
    } else {
      filled.set(126);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
          (byte) 0);
    }
    codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextCallerAddressLo(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.context/CALLER_ADDRESS_LO already set");
    } else {
      filled.set(127);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
          (byte) 0);
    }
    codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextContextNumber(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.context/CONTEXT_NUMBER already set");
    } else {
      filled.set(133);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put((byte) 0);
    }
    nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextIsRoot(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.context/IS_ROOT already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pContextIsStatic(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.context/IS_STATIC already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pContextReturnAtCapacity(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.context/RETURN_AT_CAPACITY already set");
    } else {
      filled.set(134);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
          (byte) 0);
    }
    rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextReturnAtOffset(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.context/RETURN_AT_OFFSET already set");
    } else {
      filled.set(135);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put((byte) 0);
    }
    rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextReturnDataContextNumber(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.context/RETURN_DATA_CONTEXT_NUMBER already set");
    } else {
      filled.set(136);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
          (byte) 0);
    }
    rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pContextReturnDataOffset(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.context/RETURN_DATA_OFFSET already set");
    } else {
      filled.set(137);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put((byte) 0);
    }
    rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextReturnDataSize(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.context/RETURN_DATA_SIZE already set");
    } else {
      filled.set(138);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put((byte) 0);
    }
    rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pContextUpdate(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.context/UPDATE already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscCcrsStamp(final Bytes b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.misc/CCRS_STAMP already set");
    } else {
      filled.set(118);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put((byte) 0);
    }
    addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscCcsrFlag(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.misc/CCSR_FLAG already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscExpData1(final Bytes b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_1 already set");
    } else {
      filled.set(119);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put((byte) 0);
    }
    addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscExpData2(final Bytes b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_2 already set");
    } else {
      filled.set(120);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .put((byte) 0);
    }
    balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscExpData3(final Bytes b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_3 already set");
    } else {
      filled.set(121);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .put((byte) 0);
    }
    balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscExpData4(final Bytes b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_4 already set");
    } else {
      filled.set(122);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .put((byte) 0);
    }
    codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscExpData5(final Bytes b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.misc/EXP_DATA_5 already set");
    } else {
      filled.set(123);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .put((byte) 0);
    }
    codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscExpFlag(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.misc/EXP_FLAG already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscExpInst(final long b) {
    if (filled.get(102)) {
      throw new IllegalStateException("hub.misc/EXP_INST already set");
    } else {
      filled.set(102);
    }

    expInstXorPrcCalleeGas.putLong(b);

    return this;
  }

  public Trace pMiscMmuAuxId(final long b) {
    if (filled.get(103)) {
      throw new IllegalStateException("hub.misc/MMU_AUX_ID already set");
    } else {
      filled.set(103);
    }

    mmuAuxIdXorPrcCallerGas.putLong(b);

    return this;
  }

  public Trace pMiscMmuExoSum(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.misc/MMU_EXO_SUM already set");
    } else {
      filled.set(104);
    }

    mmuExoSumXorPrcCdo.putLong(b);

    return this;
  }

  public Trace pMiscMmuFlag(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.misc/MMU_FLAG already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMmuInst(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.misc/MMU_INST already set");
    } else {
      filled.set(105);
    }

    mmuInstXorPrcCds.putLong(b);

    return this;
  }

  public Trace pMiscMmuLimb1(final Bytes b) {
    if (filled.get(113)) {
      throw new IllegalStateException("hub.misc/MMU_LIMB_1 already set");
    } else {
      filled.set(113);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmuLimb1.put((byte) 0);
    }
    mmuLimb1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMmuLimb2(final Bytes b) {
    if (filled.get(114)) {
      throw new IllegalStateException("hub.misc/MMU_LIMB_2 already set");
    } else {
      filled.set(114);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmuLimb2.put((byte) 0);
    }
    mmuLimb2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMmuPhase(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.misc/MMU_PHASE already set");
    } else {
      filled.set(106);
    }

    mmuPhaseXorPrcRac.putLong(b);

    return this;
  }

  public Trace pMiscMmuRefOffset(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.misc/MMU_REF_OFFSET already set");
    } else {
      filled.set(107);
    }

    mmuRefOffsetXorPrcRao.putLong(b);

    return this;
  }

  public Trace pMiscMmuRefSize(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("hub.misc/MMU_REF_SIZE already set");
    } else {
      filled.set(108);
    }

    mmuRefSizeXorPrcReturnGas.putLong(b);

    return this;
  }

  public Trace pMiscMmuSize(final long b) {
    if (filled.get(109)) {
      throw new IllegalStateException("hub.misc/MMU_SIZE already set");
    } else {
      filled.set(109);
    }

    mmuSize.putLong(b);

    return this;
  }

  public Trace pMiscMmuSrcId(final long b) {
    if (filled.get(110)) {
      throw new IllegalStateException("hub.misc/MMU_SRC_ID already set");
    } else {
      filled.set(110);
    }

    mmuSrcId.putLong(b);

    return this;
  }

  public Trace pMiscMmuSrcOffsetHi(final Bytes b) {
    if (filled.get(115)) {
      throw new IllegalStateException("hub.misc/MMU_SRC_OFFSET_HI already set");
    } else {
      filled.set(115);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmuSrcOffsetHi.put((byte) 0);
    }
    mmuSrcOffsetHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMmuSrcOffsetLo(final Bytes b) {
    if (filled.get(116)) {
      throw new IllegalStateException("hub.misc/MMU_SRC_OFFSET_LO already set");
    } else {
      filled.set(116);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmuSrcOffsetLo.put((byte) 0);
    }
    mmuSrcOffsetLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMmuSuccessBit(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.misc/MMU_SUCCESS_BIT already set");
    } else {
      filled.set(56);
    }

    existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMmuTgtId(final long b) {
    if (filled.get(111)) {
      throw new IllegalStateException("hub.misc/MMU_TGT_ID already set");
    } else {
      filled.set(111);
    }

    mmuTgtId.putLong(b);

    return this;
  }

  public Trace pMiscMmuTgtOffsetLo(final Bytes b) {
    if (filled.get(117)) {
      throw new IllegalStateException("hub.misc/MMU_TGT_OFFSET_LO already set");
    } else {
      filled.set(117);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      mmuTgtOffsetLo.put((byte) 0);
    }
    mmuTgtOffsetLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpDeploys(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.misc/MXP_DEPLOYS already set");
    } else {
      filled.set(57);
    }

    existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpFlag(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.misc/MXP_FLAG already set");
    } else {
      filled.set(58);
    }

    hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpGasMxp(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.misc/MXP_GAS_MXP already set");
    } else {
      filled.set(124);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .put((byte) 0);
    }
    codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpInst(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.misc/MXP_INST already set");
    } else {
      filled.set(125);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .put((byte) 0);
    }
    codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpMtntop(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.misc/MXP_MTNTOP already set");
    } else {
      filled.set(126);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
          (byte) 0);
    }
    codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpMxpx(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.misc/MXP_MXPX already set");
    } else {
      filled.set(59);
    }

    hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscMxpOffset1Hi(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_1_HI already set");
    } else {
      filled.set(127);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
          (byte) 0);
    }
    codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpOffset1Lo(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_1_LO already set");
    } else {
      filled.set(128);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .put((byte) 0);
    }
    codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpOffset2Hi(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_2_HI already set");
    } else {
      filled.set(129);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .put((byte) 0);
    }
    deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpOffset2Lo(final Bytes b) {
    if (filled.get(130)) {
      throw new IllegalStateException("hub.misc/MXP_OFFSET_2_LO already set");
    } else {
      filled.set(130);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
          (byte) 0);
    }
    deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpSize1Hi(final Bytes b) {
    if (filled.get(131)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_1_HI already set");
    } else {
      filled.set(131);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put((byte) 0);
    }
    deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpSize1Lo(final Bytes b) {
    if (filled.get(132)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_1_LO already set");
    } else {
      filled.set(132);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put((byte) 0);
    }
    nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpSize2Hi(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_2_HI already set");
    } else {
      filled.set(133);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put((byte) 0);
    }
    nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpSize2Lo(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.misc/MXP_SIZE_2_LO already set");
    } else {
      filled.set(134);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
          (byte) 0);
    }
    rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscMxpWords(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.misc/MXP_WORDS already set");
    } else {
      filled.set(135);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put((byte) 0);
    }
    rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData1(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_1 already set");
    } else {
      filled.set(136);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
          (byte) 0);
    }
    rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData2(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_2 already set");
    } else {
      filled.set(137);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put((byte) 0);
    }
    rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData3(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_3 already set");
    } else {
      filled.set(138);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put((byte) 0);
    }
    rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData4(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_4 already set");
    } else {
      filled.set(139);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrSaltHiXorOobData4XorStackItemValueHi4.put((byte) 0);
    }
    rlpaddrSaltHiXorOobData4XorStackItemValueHi4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData5(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_5 already set");
    } else {
      filled.set(140);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrSaltLoXorOobData5XorStackItemValueLo1.put((byte) 0);
    }
    rlpaddrSaltLoXorOobData5XorStackItemValueLo1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData6(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_6 already set");
    } else {
      filled.set(141);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trmRawAddressHiXorOobData6XorStackItemValueLo2.put((byte) 0);
    }
    trmRawAddressHiXorOobData6XorStackItemValueLo2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData7(final Bytes b) {
    if (filled.get(142)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_7 already set");
    } else {
      filled.set(142);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      oobData7XorStackItemValueLo3.put((byte) 0);
    }
    oobData7XorStackItemValueLo3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobData8(final Bytes b) {
    if (filled.get(143)) {
      throw new IllegalStateException("hub.misc/OOB_DATA_8 already set");
    } else {
      filled.set(143);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      oobData8XorStackItemValueLo4.put((byte) 0);
    }
    oobData8XorStackItemValueLo4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscOobFlag(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.misc/OOB_FLAG already set");
    } else {
      filled.set(60);
    }

    isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscOobInst(final long b) {
    if (filled.get(112)) {
      throw new IllegalStateException("hub.misc/OOB_INST already set");
    } else {
      filled.set(112);
    }

    oobInst.putLong(b);

    return this;
  }

  public Trace pMiscStpExists(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.misc/STP_EXISTS already set");
    } else {
      filled.set(61);
    }

    markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscStpFlag(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.misc/STP_FLAG already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscStpGasHi(final Bytes b) {
    if (filled.get(144)) {
      throw new IllegalStateException("hub.misc/STP_GAS_HI already set");
    } else {
      filled.set(144);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasHiXorStaticGas.put((byte) 0);
    }
    stpGasHiXorStaticGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpGasLo(final Bytes b) {
    if (filled.get(145)) {
      throw new IllegalStateException("hub.misc/STP_GAS_LO already set");
    } else {
      filled.set(145);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasLo.put((byte) 0);
    }
    stpGasLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpGasMxp(final Bytes b) {
    if (filled.get(146)) {
      throw new IllegalStateException("hub.misc/STP_GAS_MXP already set");
    } else {
      filled.set(146);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasMxp.put((byte) 0);
    }
    stpGasMxp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpGasPaidOutOfPocket(final Bytes b) {
    if (filled.get(147)) {
      throw new IllegalStateException("hub.misc/STP_GAS_PAID_OUT_OF_POCKET already set");
    } else {
      filled.set(147);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasPaidOutOfPocket.put((byte) 0);
    }
    stpGasPaidOutOfPocket.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpGasStipend(final Bytes b) {
    if (filled.get(148)) {
      throw new IllegalStateException("hub.misc/STP_GAS_STIPEND already set");
    } else {
      filled.set(148);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasStipend.put((byte) 0);
    }
    stpGasStipend.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpGasUpfrontGasCost(final Bytes b) {
    if (filled.get(149)) {
      throw new IllegalStateException("hub.misc/STP_GAS_UPFRONT_GAS_COST already set");
    } else {
      filled.set(149);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasUpfrontGasCost.put((byte) 0);
    }
    stpGasUpfrontGasCost.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpInstruction(final Bytes b) {
    if (filled.get(150)) {
      throw new IllegalStateException("hub.misc/STP_INSTRUCTION already set");
    } else {
      filled.set(150);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpInstruction.put((byte) 0);
    }
    stpInstruction.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpOogx(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("hub.misc/STP_OOGX already set");
    } else {
      filled.set(63);
    }

    rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pMiscStpValHi(final Bytes b) {
    if (filled.get(151)) {
      throw new IllegalStateException("hub.misc/STP_VAL_HI already set");
    } else {
      filled.set(151);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpValHi.put((byte) 0);
    }
    stpValHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpValLo(final Bytes b) {
    if (filled.get(152)) {
      throw new IllegalStateException("hub.misc/STP_VAL_LO already set");
    } else {
      filled.set(152);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpValLo.put((byte) 0);
    }
    stpValLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pMiscStpWarmth(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.misc/STP_WARMTH already set");
    } else {
      filled.set(64);
    }

    romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallAbort(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.scenario/CALL_ABORT already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallEoaSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_EOA_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallEoaSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_EOA_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallException(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.scenario/CALL_EXCEPTION already set");
    } else {
      filled.set(56);
    }

    existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallPrcFailure(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.scenario/CALL_PRC_FAILURE already set");
    } else {
      filled.set(57);
    }

    existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallPrcSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_PRC_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(58);
    }

    hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallPrcSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_PRC_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(59);
    }

    hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcFailureCallerWillRevert(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_SMC_FAILURE_CALLER_WILL_REVERT already set");
    } else {
      filled.set(60);
    }

    isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcFailureCallerWontRevert(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_SMC_FAILURE_CALLER_WONT_REVERT already set");
    } else {
      filled.set(61);
    }

    markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcSuccessCallerWillRevert(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_SMC_SUCCESS_CALLER_WILL_REVERT already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCallSmcSuccessCallerWontRevert(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException(
          "hub.scenario/CALL_SMC_SUCCESS_CALLER_WONT_REVERT already set");
    } else {
      filled.set(63);
    }

    rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateAbort(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.scenario/CREATE_ABORT already set");
    } else {
      filled.set(64);
    }

    romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateEmptyInitCodeWillRevert(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_EMPTY_INIT_CODE_WILL_REVERT already set");
    } else {
      filled.set(65);
    }

    trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateEmptyInitCodeWontRevert(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_EMPTY_INIT_CODE_WONT_REVERT already set");
    } else {
      filled.set(66);
    }

    warmthXorCreateEmptyInitCodeWontRevertXorExtFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateException(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("hub.scenario/CREATE_EXCEPTION already set");
    } else {
      filled.set(67);
    }

    warmthNewXorCreateExceptionXorHaltFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateFailureConditionWillRevert(final Boolean b) {
    if (filled.get(68)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_FAILURE_CONDITION_WILL_REVERT already set");
    } else {
      filled.set(68);
    }

    createFailureConditionWillRevertXorHashInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateFailureConditionWontRevert(final Boolean b) {
    if (filled.get(69)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_FAILURE_CONDITION_WONT_REVERT already set");
    } else {
      filled.set(69);
    }

    createFailureConditionWontRevertXorIcpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeFailureWillRevert(final Boolean b) {
    if (filled.get(70)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_NONEMPTY_INIT_CODE_FAILURE_WILL_REVERT already set");
    } else {
      filled.set(70);
    }

    createNonemptyInitCodeFailureWillRevertXorInvalidFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeFailureWontRevert(final Boolean b) {
    if (filled.get(71)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_NONEMPTY_INIT_CODE_FAILURE_WONT_REVERT already set");
    } else {
      filled.set(71);
    }

    createNonemptyInitCodeFailureWontRevertXorJumpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeSuccessWillRevert(final Boolean b) {
    if (filled.get(72)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_NONEMPTY_INIT_CODE_SUCCESS_WILL_REVERT already set");
    } else {
      filled.set(72);
    }

    createNonemptyInitCodeSuccessWillRevertXorJumpDestinationVettingRequired.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioCreateNonemptyInitCodeSuccessWontRevert(final Boolean b) {
    if (filled.get(73)) {
      throw new IllegalStateException(
          "hub.scenario/CREATE_NONEMPTY_INIT_CODE_SUCCESS_WONT_REVERT already set");
    } else {
      filled.set(73);
    }

    createNonemptyInitCodeSuccessWontRevertXorJumpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcBlake2F(final Boolean b) {
    if (filled.get(74)) {
      throw new IllegalStateException("hub.scenario/PRC_BLAKE2f already set");
    } else {
      filled.set(74);
    }

    prcBlake2FXorKecFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcCalleeGas(final long b) {
    if (filled.get(102)) {
      throw new IllegalStateException("hub.scenario/PRC_CALLEE_GAS already set");
    } else {
      filled.set(102);
    }

    expInstXorPrcCalleeGas.putLong(b);

    return this;
  }

  public Trace pScenarioPrcCallerGas(final long b) {
    if (filled.get(103)) {
      throw new IllegalStateException("hub.scenario/PRC_CALLER_GAS already set");
    } else {
      filled.set(103);
    }

    mmuAuxIdXorPrcCallerGas.putLong(b);

    return this;
  }

  public Trace pScenarioPrcCdo(final long b) {
    if (filled.get(104)) {
      throw new IllegalStateException("hub.scenario/PRC_CDO already set");
    } else {
      filled.set(104);
    }

    mmuExoSumXorPrcCdo.putLong(b);

    return this;
  }

  public Trace pScenarioPrcCds(final long b) {
    if (filled.get(105)) {
      throw new IllegalStateException("hub.scenario/PRC_CDS already set");
    } else {
      filled.set(105);
    }

    mmuInstXorPrcCds.putLong(b);

    return this;
  }

  public Trace pScenarioPrcEcadd(final Boolean b) {
    if (filled.get(75)) {
      throw new IllegalStateException("hub.scenario/PRC_ECADD already set");
    } else {
      filled.set(75);
    }

    prcEcaddXorLogFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcEcmul(final Boolean b) {
    if (filled.get(76)) {
      throw new IllegalStateException("hub.scenario/PRC_ECMUL already set");
    } else {
      filled.set(76);
    }

    prcEcmulXorLogInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcEcpairing(final Boolean b) {
    if (filled.get(77)) {
      throw new IllegalStateException("hub.scenario/PRC_ECPAIRING already set");
    } else {
      filled.set(77);
    }

    prcEcpairingXorMachineStateFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcEcrecover(final Boolean b) {
    if (filled.get(78)) {
      throw new IllegalStateException("hub.scenario/PRC_ECRECOVER already set");
    } else {
      filled.set(78);
    }

    prcEcrecoverXorMaxcsx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcFailureKnownToHub(final Boolean b) {
    if (filled.get(79)) {
      throw new IllegalStateException("hub.scenario/PRC_FAILURE_KNOWN_TO_HUB already set");
    } else {
      filled.set(79);
    }

    prcFailureKnownToHubXorModFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcFailureKnownToRam(final Boolean b) {
    if (filled.get(80)) {
      throw new IllegalStateException("hub.scenario/PRC_FAILURE_KNOWN_TO_RAM already set");
    } else {
      filled.set(80);
    }

    prcFailureKnownToRamXorMulFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcIdentity(final Boolean b) {
    if (filled.get(81)) {
      throw new IllegalStateException("hub.scenario/PRC_IDENTITY already set");
    } else {
      filled.set(81);
    }

    prcIdentityXorMxpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcModexp(final Boolean b) {
    if (filled.get(82)) {
      throw new IllegalStateException("hub.scenario/PRC_MODEXP already set");
    } else {
      filled.set(82);
    }

    prcModexpXorMxpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcRac(final long b) {
    if (filled.get(106)) {
      throw new IllegalStateException("hub.scenario/PRC_RAC already set");
    } else {
      filled.set(106);
    }

    mmuPhaseXorPrcRac.putLong(b);

    return this;
  }

  public Trace pScenarioPrcRao(final long b) {
    if (filled.get(107)) {
      throw new IllegalStateException("hub.scenario/PRC_RAO already set");
    } else {
      filled.set(107);
    }

    mmuRefOffsetXorPrcRao.putLong(b);

    return this;
  }

  public Trace pScenarioPrcReturnGas(final long b) {
    if (filled.get(108)) {
      throw new IllegalStateException("hub.scenario/PRC_RETURN_GAS already set");
    } else {
      filled.set(108);
    }

    mmuRefSizeXorPrcReturnGas.putLong(b);

    return this;
  }

  public Trace pScenarioPrcRipemd160(final Boolean b) {
    if (filled.get(83)) {
      throw new IllegalStateException("hub.scenario/PRC_RIPEMD-160 already set");
    } else {
      filled.set(83);
    }

    prcRipemd160XorOogx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcSha2256(final Boolean b) {
    if (filled.get(84)) {
      throw new IllegalStateException("hub.scenario/PRC_SHA2-256 already set");
    } else {
      filled.set(84);
    }

    prcSha2256XorOpcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcSuccessWillRevert(final Boolean b) {
    if (filled.get(85)) {
      throw new IllegalStateException("hub.scenario/PRC_SUCCESS_WILL_REVERT already set");
    } else {
      filled.set(85);
    }

    prcSuccessWillRevertXorPushpopFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioPrcSuccessWontRevert(final Boolean b) {
    if (filled.get(86)) {
      throw new IllegalStateException("hub.scenario/PRC_SUCCESS_WONT_REVERT already set");
    } else {
      filled.set(86);
    }

    prcSuccessWontRevertXorRdcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnException(final Boolean b) {
    if (filled.get(87)) {
      throw new IllegalStateException("hub.scenario/RETURN_EXCEPTION already set");
    } else {
      filled.set(87);
    }

    returnExceptionXorShfFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentEmptyCodeWillRevert(final Boolean b) {
    if (filled.get(88)) {
      throw new IllegalStateException(
          "hub.scenario/RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT already set");
    } else {
      filled.set(88);
    }

    returnFromDeploymentEmptyCodeWillRevertXorSox.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentEmptyCodeWontRevert(final Boolean b) {
    if (filled.get(89)) {
      throw new IllegalStateException(
          "hub.scenario/RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT already set");
    } else {
      filled.set(89);
    }

    returnFromDeploymentEmptyCodeWontRevertXorSstorex.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentNonemptyCodeWillRevert(final Boolean b) {
    if (filled.get(90)) {
      throw new IllegalStateException(
          "hub.scenario/RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT already set");
    } else {
      filled.set(90);
    }

    returnFromDeploymentNonemptyCodeWillRevertXorStackramFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromDeploymentNonemptyCodeWontRevert(final Boolean b) {
    if (filled.get(91)) {
      throw new IllegalStateException(
          "hub.scenario/RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT already set");
    } else {
      filled.set(91);
    }

    returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromMessageCallWillTouchRam(final Boolean b) {
    if (filled.get(92)) {
      throw new IllegalStateException(
          "hub.scenario/RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM already set");
    } else {
      filled.set(92);
    }

    returnFromMessageCallWillTouchRamXorStackItemPop2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioReturnFromMessageCallWontTouchRam(final Boolean b) {
    if (filled.get(93)) {
      throw new IllegalStateException(
          "hub.scenario/RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM already set");
    } else {
      filled.set(93);
    }

    returnFromMessageCallWontTouchRamXorStackItemPop3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructException(final Boolean b) {
    if (filled.get(94)) {
      throw new IllegalStateException("hub.scenario/SELFDESTRUCT_EXCEPTION already set");
    } else {
      filled.set(94);
    }

    selfdestructExceptionXorStackItemPop4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructWillRevert(final Boolean b) {
    if (filled.get(95)) {
      throw new IllegalStateException("hub.scenario/SELFDESTRUCT_WILL_REVERT already set");
    } else {
      filled.set(95);
    }

    selfdestructWillRevertXorStaticx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructWontRevertAlreadyMarked(final Boolean b) {
    if (filled.get(96)) {
      throw new IllegalStateException(
          "hub.scenario/SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED already set");
    } else {
      filled.set(96);
    }

    selfdestructWontRevertAlreadyMarkedXorStaticFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pScenarioSelfdestructWontRevertNotYetMarked(final Boolean b) {
    if (filled.get(97)) {
      throw new IllegalStateException(
          "hub.scenario/SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED already set");
    } else {
      filled.set(97);
    }

    selfdestructWontRevertNotYetMarkedXorStoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackAccFlag(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.stack/ACC_FLAG already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackAddFlag(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.stack/ADD_FLAG already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackAlpha(final Bytes b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.stack/ALPHA already set");
    } else {
      filled.set(118);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put((byte) 0);
    }
    addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackBinFlag(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.stack/BIN_FLAG already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackBtcFlag(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.stack/BTC_FLAG already set");
    } else {
      filled.set(56);
    }

    existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackCallFlag(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.stack/CALL_FLAG already set");
    } else {
      filled.set(57);
    }

    existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackConFlag(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.stack/CON_FLAG already set");
    } else {
      filled.set(58);
    }

    hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackCopyFlag(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.stack/COPY_FLAG already set");
    } else {
      filled.set(59);
    }

    hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackCreateFlag(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.stack/CREATE_FLAG already set");
    } else {
      filled.set(60);
    }

    isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag1(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_1 already set");
    } else {
      filled.set(61);
    }

    markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag2(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_2 already set");
    } else {
      filled.set(62);
    }

    markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag3(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_3 already set");
    } else {
      filled.set(63);
    }

    rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDecFlag4(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("hub.stack/DEC_FLAG_4 already set");
    } else {
      filled.set(64);
    }

    romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackDelta(final Bytes b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.stack/DELTA already set");
    } else {
      filled.set(119);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put((byte) 0);
    }
    addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackDupFlag(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException("hub.stack/DUP_FLAG already set");
    } else {
      filled.set(65);
    }

    trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackExtFlag(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException("hub.stack/EXT_FLAG already set");
    } else {
      filled.set(66);
    }

    warmthXorCreateEmptyInitCodeWontRevertXorExtFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackHaltFlag(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("hub.stack/HALT_FLAG already set");
    } else {
      filled.set(67);
    }

    warmthNewXorCreateExceptionXorHaltFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackHashInfoFlag(final Boolean b) {
    if (filled.get(68)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_FLAG already set");
    } else {
      filled.set(68);
    }

    createFailureConditionWillRevertXorHashInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackHashInfoKeccakHi(final Bytes b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_KECCAK_HI already set");
    } else {
      filled.set(120);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .put((byte) 0);
    }
    balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackHashInfoKeccakLo(final Bytes b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_KECCAK_LO already set");
    } else {
      filled.set(121);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .put((byte) 0);
    }
    balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackHashInfoSize(final Bytes b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.stack/HASH_INFO_SIZE already set");
    } else {
      filled.set(122);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .put((byte) 0);
    }
    codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackIcpx(final Boolean b) {
    if (filled.get(69)) {
      throw new IllegalStateException("hub.stack/ICPX already set");
    } else {
      filled.set(69);
    }

    createFailureConditionWontRevertXorIcpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackInstruction(final Bytes b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.stack/INSTRUCTION already set");
    } else {
      filled.set(123);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .put((byte) 0);
    }
    codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackInvalidFlag(final Boolean b) {
    if (filled.get(70)) {
      throw new IllegalStateException("hub.stack/INVALID_FLAG already set");
    } else {
      filled.set(70);
    }

    createNonemptyInitCodeFailureWillRevertXorInvalidFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackJumpDestinationVettingRequired(final Boolean b) {
    if (filled.get(72)) {
      throw new IllegalStateException("hub.stack/JUMP_DESTINATION_VETTING_REQUIRED already set");
    } else {
      filled.set(72);
    }

    createNonemptyInitCodeSuccessWillRevertXorJumpDestinationVettingRequired.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackJumpFlag(final Boolean b) {
    if (filled.get(73)) {
      throw new IllegalStateException("hub.stack/JUMP_FLAG already set");
    } else {
      filled.set(73);
    }

    createNonemptyInitCodeSuccessWontRevertXorJumpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackJumpx(final Boolean b) {
    if (filled.get(71)) {
      throw new IllegalStateException("hub.stack/JUMPX already set");
    } else {
      filled.set(71);
    }

    createNonemptyInitCodeFailureWontRevertXorJumpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackKecFlag(final Boolean b) {
    if (filled.get(74)) {
      throw new IllegalStateException("hub.stack/KEC_FLAG already set");
    } else {
      filled.set(74);
    }

    prcBlake2FXorKecFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackLogFlag(final Boolean b) {
    if (filled.get(75)) {
      throw new IllegalStateException("hub.stack/LOG_FLAG already set");
    } else {
      filled.set(75);
    }

    prcEcaddXorLogFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackLogInfoFlag(final Boolean b) {
    if (filled.get(76)) {
      throw new IllegalStateException("hub.stack/LOG_INFO_FLAG already set");
    } else {
      filled.set(76);
    }

    prcEcmulXorLogInfoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMachineStateFlag(final Boolean b) {
    if (filled.get(77)) {
      throw new IllegalStateException("hub.stack/MACHINE_STATE_FLAG already set");
    } else {
      filled.set(77);
    }

    prcEcpairingXorMachineStateFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMaxcsx(final Boolean b) {
    if (filled.get(78)) {
      throw new IllegalStateException("hub.stack/MAXCSX already set");
    } else {
      filled.set(78);
    }

    prcEcrecoverXorMaxcsx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackModFlag(final Boolean b) {
    if (filled.get(79)) {
      throw new IllegalStateException("hub.stack/MOD_FLAG already set");
    } else {
      filled.set(79);
    }

    prcFailureKnownToHubXorModFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMulFlag(final Boolean b) {
    if (filled.get(80)) {
      throw new IllegalStateException("hub.stack/MUL_FLAG already set");
    } else {
      filled.set(80);
    }

    prcFailureKnownToRamXorMulFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMxpFlag(final Boolean b) {
    if (filled.get(82)) {
      throw new IllegalStateException("hub.stack/MXP_FLAG already set");
    } else {
      filled.set(82);
    }

    prcModexpXorMxpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackMxpx(final Boolean b) {
    if (filled.get(81)) {
      throw new IllegalStateException("hub.stack/MXPX already set");
    } else {
      filled.set(81);
    }

    prcIdentityXorMxpx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackNbAdded(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.stack/NB_ADDED already set");
    } else {
      filled.set(124);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .put((byte) 0);
    }
    codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackNbRemoved(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.stack/NB_REMOVED already set");
    } else {
      filled.set(125);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .put((byte) 0);
    }
    codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackOogx(final Boolean b) {
    if (filled.get(83)) {
      throw new IllegalStateException("hub.stack/OOGX already set");
    } else {
      filled.set(83);
    }

    prcRipemd160XorOogx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackOpcx(final Boolean b) {
    if (filled.get(84)) {
      throw new IllegalStateException("hub.stack/OPCX already set");
    } else {
      filled.set(84);
    }

    prcSha2256XorOpcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackPushValueHi(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.stack/PUSH_VALUE_HI already set");
    } else {
      filled.set(126);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
          (byte) 0);
    }
    codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackPushValueLo(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.stack/PUSH_VALUE_LO already set");
    } else {
      filled.set(127);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
          (byte) 0);
    }
    codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackPushpopFlag(final Boolean b) {
    if (filled.get(85)) {
      throw new IllegalStateException("hub.stack/PUSHPOP_FLAG already set");
    } else {
      filled.set(85);
    }

    prcSuccessWillRevertXorPushpopFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackRdcx(final Boolean b) {
    if (filled.get(86)) {
      throw new IllegalStateException("hub.stack/RDCX already set");
    } else {
      filled.set(86);
    }

    prcSuccessWontRevertXorRdcx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackShfFlag(final Boolean b) {
    if (filled.get(87)) {
      throw new IllegalStateException("hub.stack/SHF_FLAG already set");
    } else {
      filled.set(87);
    }

    returnExceptionXorShfFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSox(final Boolean b) {
    if (filled.get(88)) {
      throw new IllegalStateException("hub.stack/SOX already set");
    } else {
      filled.set(88);
    }

    returnFromDeploymentEmptyCodeWillRevertXorSox.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSstorex(final Boolean b) {
    if (filled.get(89)) {
      throw new IllegalStateException("hub.stack/SSTOREX already set");
    } else {
      filled.set(89);
    }

    returnFromDeploymentEmptyCodeWontRevertXorSstorex.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemHeight1(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_1 already set");
    } else {
      filled.set(128);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .put((byte) 0);
    }
    codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemHeight2(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_2 already set");
    } else {
      filled.set(129);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .put((byte) 0);
    }
    deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemHeight3(final Bytes b) {
    if (filled.get(130)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_3 already set");
    } else {
      filled.set(130);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
          (byte) 0);
    }
    deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemHeight4(final Bytes b) {
    if (filled.get(131)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_HEIGHT_4 already set");
    } else {
      filled.set(131);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put((byte) 0);
    }
    deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemPop1(final Boolean b) {
    if (filled.get(91)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_1 already set");
    } else {
      filled.set(91);
    }

    returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemPop2(final Boolean b) {
    if (filled.get(92)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_2 already set");
    } else {
      filled.set(92);
    }

    returnFromMessageCallWillTouchRamXorStackItemPop2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemPop3(final Boolean b) {
    if (filled.get(93)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_3 already set");
    } else {
      filled.set(93);
    }

    returnFromMessageCallWontTouchRamXorStackItemPop3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemPop4(final Boolean b) {
    if (filled.get(94)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_POP_4 already set");
    } else {
      filled.set(94);
    }

    selfdestructExceptionXorStackItemPop4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStackItemStamp1(final Bytes b) {
    if (filled.get(132)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_1 already set");
    } else {
      filled.set(132);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put((byte) 0);
    }
    nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemStamp2(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_2 already set");
    } else {
      filled.set(133);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put((byte) 0);
    }
    nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemStamp3(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_3 already set");
    } else {
      filled.set(134);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
          (byte) 0);
    }
    rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemStamp4(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_STAMP_4 already set");
    } else {
      filled.set(135);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put((byte) 0);
    }
    rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueHi1(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_1 already set");
    } else {
      filled.set(136);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
          (byte) 0);
    }
    rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueHi2(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_2 already set");
    } else {
      filled.set(137);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put((byte) 0);
    }
    rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueHi3(final Bytes b) {
    if (filled.get(138)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_3 already set");
    } else {
      filled.set(138);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put((byte) 0);
    }
    rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueHi4(final Bytes b) {
    if (filled.get(139)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_HI_4 already set");
    } else {
      filled.set(139);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrSaltHiXorOobData4XorStackItemValueHi4.put((byte) 0);
    }
    rlpaddrSaltHiXorOobData4XorStackItemValueHi4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueLo1(final Bytes b) {
    if (filled.get(140)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_1 already set");
    } else {
      filled.set(140);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrSaltLoXorOobData5XorStackItemValueLo1.put((byte) 0);
    }
    rlpaddrSaltLoXorOobData5XorStackItemValueLo1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueLo2(final Bytes b) {
    if (filled.get(141)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_2 already set");
    } else {
      filled.set(141);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trmRawAddressHiXorOobData6XorStackItemValueLo2.put((byte) 0);
    }
    trmRawAddressHiXorOobData6XorStackItemValueLo2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueLo3(final Bytes b) {
    if (filled.get(142)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_3 already set");
    } else {
      filled.set(142);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      oobData7XorStackItemValueLo3.put((byte) 0);
    }
    oobData7XorStackItemValueLo3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackItemValueLo4(final Bytes b) {
    if (filled.get(143)) {
      throw new IllegalStateException("hub.stack/STACK_ITEM_VALUE_LO_4 already set");
    } else {
      filled.set(143);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      oobData8XorStackItemValueLo4.put((byte) 0);
    }
    oobData8XorStackItemValueLo4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStackramFlag(final Boolean b) {
    if (filled.get(90)) {
      throw new IllegalStateException("hub.stack/STACKRAM_FLAG already set");
    } else {
      filled.set(90);
    }

    returnFromDeploymentNonemptyCodeWillRevertXorStackramFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStaticFlag(final Boolean b) {
    if (filled.get(96)) {
      throw new IllegalStateException("hub.stack/STATIC_FLAG already set");
    } else {
      filled.set(96);
    }

    selfdestructWontRevertAlreadyMarkedXorStaticFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStaticGas(final Bytes b) {
    if (filled.get(144)) {
      throw new IllegalStateException("hub.stack/STATIC_GAS already set");
    } else {
      filled.set(144);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stpGasHiXorStaticGas.put((byte) 0);
    }
    stpGasHiXorStaticGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStackStaticx(final Boolean b) {
    if (filled.get(95)) {
      throw new IllegalStateException("hub.stack/STATICX already set");
    } else {
      filled.set(95);
    }

    selfdestructWillRevertXorStaticx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackStoFlag(final Boolean b) {
    if (filled.get(97)) {
      throw new IllegalStateException("hub.stack/STO_FLAG already set");
    } else {
      filled.set(97);
    }

    selfdestructWontRevertNotYetMarkedXorStoFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSux(final Boolean b) {
    if (filled.get(98)) {
      throw new IllegalStateException("hub.stack/SUX already set");
    } else {
      filled.set(98);
    }

    sux.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackSwapFlag(final Boolean b) {
    if (filled.get(99)) {
      throw new IllegalStateException("hub.stack/SWAP_FLAG already set");
    } else {
      filled.set(99);
    }

    swapFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackTxnFlag(final Boolean b) {
    if (filled.get(100)) {
      throw new IllegalStateException("hub.stack/TXN_FLAG already set");
    } else {
      filled.set(100);
    }

    txnFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStackWcpFlag(final Boolean b) {
    if (filled.get(101)) {
      throw new IllegalStateException("hub.stack/WCP_FLAG already set");
    } else {
      filled.set(101);
    }

    wcpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageAddressHi(final Bytes b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.storage/ADDRESS_HI already set");
    } else {
      filled.set(118);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put((byte) 0);
    }
    addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageAddressLo(final Bytes b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.storage/ADDRESS_LO already set");
    } else {
      filled.set(119);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put((byte) 0);
    }
    addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageDeploymentNumber(final Bytes b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.storage/DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(120);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .put((byte) 0);
    }
    balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageDeploymentNumberInfty(final Bytes b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.storage/DEPLOYMENT_NUMBER_INFTY already set");
    } else {
      filled.set(121);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .put((byte) 0);
    }
    balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageStorageKeyHi(final Bytes b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.storage/STORAGE_KEY_HI already set");
    } else {
      filled.set(122);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .put((byte) 0);
    }
    codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageStorageKeyLo(final Bytes b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.storage/STORAGE_KEY_LO already set");
    } else {
      filled.set(123);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .put((byte) 0);
    }
    codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageValueCurrChanges(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_CHANGES already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueCurrHi(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_HI already set");
    } else {
      filled.set(124);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .put((byte) 0);
    }
    codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageValueCurrIsOrig(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_IS_ORIG already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueCurrIsZero(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_IS_ZERO already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueCurrLo(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.storage/VALUE_CURR_LO already set");
    } else {
      filled.set(125);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .put((byte) 0);
    }
    codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageValueNextHi(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_HI already set");
    } else {
      filled.set(126);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
          (byte) 0);
    }
    codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageValueNextIsCurr(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_IS_CURR already set");
    } else {
      filled.set(56);
    }

    existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueNextIsOrig(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_IS_ORIG already set");
    } else {
      filled.set(57);
    }

    existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueNextIsZero(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_IS_ZERO already set");
    } else {
      filled.set(58);
    }

    hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueNextLo(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.storage/VALUE_NEXT_LO already set");
    } else {
      filled.set(127);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
          (byte) 0);
    }
    codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageValueOrigHi(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.storage/VALUE_ORIG_HI already set");
    } else {
      filled.set(128);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .put((byte) 0);
    }
    codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageValueOrigIsZero(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("hub.storage/VALUE_ORIG_IS_ZERO already set");
    } else {
      filled.set(59);
    }

    hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageValueOrigLo(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.storage/VALUE_ORIG_LO already set");
    } else {
      filled.set(129);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .put((byte) 0);
    }
    deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pStorageWarmth(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("hub.storage/WARMTH already set");
    } else {
      filled.set(60);
    }

    isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pStorageWarmthNew(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("hub.storage/WARMTH_NEW already set");
    } else {
      filled.set(61);
    }

    markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionBasefee(final Bytes b) {
    if (filled.get(118)) {
      throw new IllegalStateException("hub.transaction/BASEFEE already set");
    } else {
      filled.set(118);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put((byte) 0);
    }
    addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionBatchNum(final Bytes b) {
    if (filled.get(119)) {
      throw new IllegalStateException("hub.transaction/BATCH_NUM already set");
    } else {
      filled.set(119);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put((byte) 0);
    }
    addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionCallDataSize(final Bytes b) {
    if (filled.get(120)) {
      throw new IllegalStateException("hub.transaction/CALL_DATA_SIZE already set");
    } else {
      filled.set(120);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .put((byte) 0);
    }
    balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionCoinbaseAddressHi(final Bytes b) {
    if (filled.get(121)) {
      throw new IllegalStateException("hub.transaction/COINBASE_ADDRESS_HI already set");
    } else {
      filled.set(121);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .put((byte) 0);
    }
    balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionCoinbaseAddressLo(final Bytes b) {
    if (filled.get(122)) {
      throw new IllegalStateException("hub.transaction/COINBASE_ADDRESS_LO already set");
    } else {
      filled.set(122);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .put((byte) 0);
    }
    codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionCopyTxcd(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("hub.transaction/COPY_TXCD already set");
    } else {
      filled.set(53);
    }

    deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionFromAddressHi(final Bytes b) {
    if (filled.get(123)) {
      throw new IllegalStateException("hub.transaction/FROM_ADDRESS_HI already set");
    } else {
      filled.set(123);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .put((byte) 0);
    }
    codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionFromAddressLo(final Bytes b) {
    if (filled.get(124)) {
      throw new IllegalStateException("hub.transaction/FROM_ADDRESS_LO already set");
    } else {
      filled.set(124);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .put((byte) 0);
    }
    codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionGasInitiallyAvailable(final Bytes b) {
    if (filled.get(125)) {
      throw new IllegalStateException("hub.transaction/GAS_INITIALLY_AVAILABLE already set");
    } else {
      filled.set(125);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .put((byte) 0);
    }
    codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionGasLeftover(final Bytes b) {
    if (filled.get(126)) {
      throw new IllegalStateException("hub.transaction/GAS_LEFTOVER already set");
    } else {
      filled.set(126);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
          (byte) 0);
    }
    codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionGasLimit(final Bytes b) {
    if (filled.get(127)) {
      throw new IllegalStateException("hub.transaction/GAS_LIMIT already set");
    } else {
      filled.set(127);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
          (byte) 0);
    }
    codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionGasPrice(final Bytes b) {
    if (filled.get(128)) {
      throw new IllegalStateException("hub.transaction/GAS_PRICE already set");
    } else {
      filled.set(128);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .put((byte) 0);
    }
    codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionInitCodeSize(final Bytes b) {
    if (filled.get(130)) {
      throw new IllegalStateException("hub.transaction/INIT_CODE_SIZE already set");
    } else {
      filled.set(130);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
          (byte) 0);
    }
    deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionInitialBalance(final Bytes b) {
    if (filled.get(129)) {
      throw new IllegalStateException("hub.transaction/INITIAL_BALANCE already set");
    } else {
      filled.set(129);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .put((byte) 0);
    }
    deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
        .put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionIsDeployment(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("hub.transaction/IS_DEPLOYMENT already set");
    } else {
      filled.set(54);
    }

    deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionIsType2(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("hub.transaction/IS_TYPE2 already set");
    } else {
      filled.set(55);
    }

    deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
        .put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionNonce(final Bytes b) {
    if (filled.get(131)) {
      throw new IllegalStateException("hub.transaction/NONCE already set");
    } else {
      filled.set(131);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put((byte) 0);
    }
    deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionPriorityFeePerGas(final Bytes b) {
    if (filled.get(132)) {
      throw new IllegalStateException("hub.transaction/PRIORITY_FEE_PER_GAS already set");
    } else {
      filled.set(132);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put((byte) 0);
    }
    nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionRefundCounterInfinity(final Bytes b) {
    if (filled.get(133)) {
      throw new IllegalStateException("hub.transaction/REFUND_COUNTER_INFINITY already set");
    } else {
      filled.set(133);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put((byte) 0);
    }
    nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionRefundEffective(final Bytes b) {
    if (filled.get(134)) {
      throw new IllegalStateException("hub.transaction/REFUND_EFFECTIVE already set");
    } else {
      filled.set(134);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
          (byte) 0);
    }
    rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionRequiresEvmExecution(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("hub.transaction/REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(56);
    }

    existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionStatusCode(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("hub.transaction/STATUS_CODE already set");
    } else {
      filled.set(57);
    }

    existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.put(
        (byte) (b ? 1 : 0));

    return this;
  }

  public Trace pTransactionToAddressHi(final Bytes b) {
    if (filled.get(135)) {
      throw new IllegalStateException("hub.transaction/TO_ADDRESS_HI already set");
    } else {
      filled.set(135);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put((byte) 0);
    }
    rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionToAddressLo(final Bytes b) {
    if (filled.get(136)) {
      throw new IllegalStateException("hub.transaction/TO_ADDRESS_LO already set");
    } else {
      filled.set(136);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
          (byte) 0);
    }
    rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.put(
        b.toArrayUnsafe());

    return this;
  }

  public Trace pTransactionValue(final Bytes b) {
    if (filled.get(137)) {
      throw new IllegalStateException("hub.transaction/VALUE already set");
    } else {
      filled.set(137);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put((byte) 0);
    }
    rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.put(b.toArrayUnsafe());

    return this;
  }

  public Trace peekAtAccount(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("hub.PEEK_AT_ACCOUNT already set");
    } else {
      filled.set(28);
    }

    peekAtAccount.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtContext(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("hub.PEEK_AT_CONTEXT already set");
    } else {
      filled.set(29);
    }

    peekAtContext.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtMiscellaneous(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("hub.PEEK_AT_MISCELLANEOUS already set");
    } else {
      filled.set(30);
    }

    peekAtMiscellaneous.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtScenario(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("hub.PEEK_AT_SCENARIO already set");
    } else {
      filled.set(31);
    }

    peekAtScenario.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtStack(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("hub.PEEK_AT_STACK already set");
    } else {
      filled.set(32);
    }

    peekAtStack.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtStorage(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("hub.PEEK_AT_STORAGE already set");
    } else {
      filled.set(33);
    }

    peekAtStorage.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace peekAtTransaction(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("hub.PEEK_AT_TRANSACTION already set");
    } else {
      filled.set(34);
    }

    peekAtTransaction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace programCounter(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER already set");
    } else {
      filled.set(35);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      programCounter.put((byte) 0);
    }
    programCounter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace programCounterNew(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER_NEW already set");
    } else {
      filled.set(36);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      programCounterNew.put((byte) 0);
    }
    programCounterNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace refundCounter(final Bytes b) {
    if (filled.get(37)) {
      throw new IllegalStateException("hub.REFUND_COUNTER already set");
    } else {
      filled.set(37);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refundCounter.put((byte) 0);
    }
    refundCounter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace refundCounterNew(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("hub.REFUND_COUNTER_NEW already set");
    } else {
      filled.set(38);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      refundCounterNew.put((byte) 0);
    }
    refundCounterNew.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stoFinal(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("hub.sto_FINAL already set");
    } else {
      filled.set(51);
    }

    stoFinal.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stoFirst(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("hub.sto_FIRST already set");
    } else {
      filled.set(52);
    }

    stoFirst.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace subStamp(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("hub.SUB_STAMP already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      subStamp.put((byte) 0);
    }
    subStamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace transactionReverts(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("hub.TRANSACTION_REVERTS already set");
    } else {
      filled.set(40);
    }

    transactionReverts.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace twoLineInstruction(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("hub.TWO_LINE_INSTRUCTION already set");
    } else {
      filled.set(41);
    }

    twoLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txExec(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("hub.TX_EXEC already set");
    } else {
      filled.set(42);
    }

    txExec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txFinl(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("hub.TX_FINL already set");
    } else {
      filled.set(43);
    }

    txFinl.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txInit(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("hub.TX_INIT already set");
    } else {
      filled.set(44);
    }

    txInit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txSkip(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("hub.TX_SKIP already set");
    } else {
      filled.set(45);
    }

    txSkip.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace txWarm(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("hub.TX_WARM already set");
    } else {
      filled.set(46);
    }

    txWarm.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("hub.ABSOLUTE_TRANSACTION_NUMBER has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("hub.acc_FINAL has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("hub.acc_FIRST has not been filled");
    }

    if (!filled.get(118)) {
      throw new IllegalStateException(
          "hub.ADDRESS_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_ALPHA_xor_ADDRESS_HI_xor_BASEFEE has not been filled");
    }

    if (!filled.get(119)) {
      throw new IllegalStateException(
          "hub.ADDRESS_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP_DATA_1_xor_DELTA_xor_ADDRESS_LO_xor_BATCH_NUM has not been filled");
    }

    if (!filled.get(121)) {
      throw new IllegalStateException(
          "hub.BALANCE_NEW_xor_BYTE_CODE_ADDRESS_HI_xor_EXP_DATA_3_xor_HASH_INFO_KECCAK_LO_xor_DEPLOYMENT_NUMBER_INFTY_xor_COINBASE_ADDRESS_HI has not been filled");
    }

    if (!filled.get(120)) {
      throw new IllegalStateException(
          "hub.BALANCE_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP_DATA_2_xor_HASH_INFO_KECCAK_HI_xor_DEPLOYMENT_NUMBER_xor_CALL_DATA_SIZE has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("hub.BATCH_NUMBER has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("hub.CALLER_CONTEXT_NUMBER has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("hub.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(122)) {
      throw new IllegalStateException(
          "hub.CODE_FRAGMENT_INDEX_xor_BYTE_CODE_ADDRESS_LO_xor_EXP_DATA_4_xor_HASH_INFO_SIZE_xor_STORAGE_KEY_HI_xor_COINBASE_ADDRESS_LO has not been filled");
    }

    if (!filled.get(124)) {
      throw new IllegalStateException(
          "hub.CODE_HASH_HI_NEW_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MXP_GAS_MXP_xor_NB_ADDED_xor_VALUE_CURR_HI_xor_FROM_ADDRESS_LO has not been filled");
    }

    if (!filled.get(123)) {
      throw new IllegalStateException(
          "hub.CODE_HASH_HI_xor_BYTE_CODE_CODE_FRAGMENT_INDEX_xor_EXP_DATA_5_xor_INSTRUCTION_xor_STORAGE_KEY_LO_xor_FROM_ADDRESS_HI has not been filled");
    }

    if (!filled.get(126)) {
      throw new IllegalStateException(
          "hub.CODE_HASH_LO_NEW_xor_CALLER_ADDRESS_HI_xor_MXP_MTNTOP_xor_PUSH_VALUE_HI_xor_VALUE_NEXT_HI_xor_GAS_LEFTOVER has not been filled");
    }

    if (!filled.get(125)) {
      throw new IllegalStateException(
          "hub.CODE_HASH_LO_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MXP_INST_xor_NB_REMOVED_xor_VALUE_CURR_LO_xor_GAS_INITIALLY_AVAILABLE has not been filled");
    }

    if (!filled.get(128)) {
      throw new IllegalStateException(
          "hub.CODE_SIZE_NEW_xor_CALL_DATA_CONTEXT_NUMBER_xor_MXP_OFFSET_1_LO_xor_STACK_ITEM_HEIGHT_1_xor_VALUE_ORIG_HI_xor_GAS_PRICE has not been filled");
    }

    if (!filled.get(127)) {
      throw new IllegalStateException(
          "hub.CODE_SIZE_xor_CALLER_ADDRESS_LO_xor_MXP_OFFSET_1_HI_xor_PUSH_VALUE_LO_xor_VALUE_NEXT_LO_xor_GAS_LIMIT has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("hub.con_AGAIN has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("hub.con_FIRST has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("hub.CONTEXT_GETS_REVERTED has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("hub.CONTEXT_MAY_CHANGE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("hub.CONTEXT_NUMBER_NEW has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("hub.CONTEXT_REVERT_STAMP has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("hub.CONTEXT_SELF_REVERTS has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("hub.CONTEXT_WILL_REVERT has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("hub.COUNTER_NSR has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("hub.COUNTER_TLI has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException(
          "hub.CREATE_FAILURE_CONDITION_WILL_REVERT_xor_HASH_INFO_FLAG has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException(
          "hub.CREATE_FAILURE_CONDITION_WONT_REVERT_xor_ICPX has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException(
          "hub.CREATE_NONEMPTY_INIT_CODE_FAILURE_WILL_REVERT_xor_INVALID_FLAG has not been filled");
    }

    if (!filled.get(71)) {
      throw new IllegalStateException(
          "hub.CREATE_NONEMPTY_INIT_CODE_FAILURE_WONT_REVERT_xor_JUMPX has not been filled");
    }

    if (!filled.get(72)) {
      throw new IllegalStateException(
          "hub.CREATE_NONEMPTY_INIT_CODE_SUCCESS_WILL_REVERT_xor_JUMP_DESTINATION_VETTING_REQUIRED has not been filled");
    }

    if (!filled.get(73)) {
      throw new IllegalStateException(
          "hub.CREATE_NONEMPTY_INIT_CODE_SUCCESS_WONT_REVERT_xor_JUMP_FLAG has not been filled");
    }

    if (!filled.get(130)) {
      throw new IllegalStateException(
          "hub.DEPLOYMENT_NUMBER_INFTY_xor_CALL_DATA_SIZE_xor_MXP_OFFSET_2_LO_xor_STACK_ITEM_HEIGHT_3_xor_INIT_CODE_SIZE has not been filled");
    }

    if (!filled.get(131)) {
      throw new IllegalStateException(
          "hub.DEPLOYMENT_NUMBER_NEW_xor_CALL_STACK_DEPTH_xor_MXP_SIZE_1_HI_xor_STACK_ITEM_HEIGHT_4_xor_NONCE has not been filled");
    }

    if (!filled.get(129)) {
      throw new IllegalStateException(
          "hub.DEPLOYMENT_NUMBER_xor_CALL_DATA_OFFSET_xor_MXP_OFFSET_2_HI_xor_STACK_ITEM_HEIGHT_2_xor_VALUE_ORIG_LO_xor_INITIAL_BALANCE has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException(
          "hub.DEPLOYMENT_STATUS_INFTY_xor_IS_STATIC_xor_EXP_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_ADD_FLAG_xor_VALUE_CURR_IS_ORIG_xor_IS_DEPLOYMENT has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException(
          "hub.DEPLOYMENT_STATUS_NEW_xor_UPDATE_xor_MMU_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BIN_FLAG_xor_VALUE_CURR_IS_ZERO_xor_IS_TYPE2 has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException(
          "hub.DEPLOYMENT_STATUS_xor_IS_ROOT_xor_CCSR_FLAG_xor_CALL_ABORT_xor_ACC_FLAG_xor_VALUE_CURR_CHANGES_xor_COPY_TXCD has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("hub.DOM_STAMP has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("hub.EXCEPTION_AHOY has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException(
          "hub.EXISTS_NEW_xor_MXP_DEPLOYS_xor_CALL_PRC_FAILURE_xor_CALL_FLAG_xor_VALUE_NEXT_IS_ORIG_xor_STATUS_CODE has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException(
          "hub.EXISTS_xor_MMU_SUCCESS_BIT_xor_CALL_EXCEPTION_xor_BTC_FLAG_xor_VALUE_NEXT_IS_CURR_xor_REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(102)) {
      throw new IllegalStateException("hub.EXP_INST_xor_PRC_CALLEE_GAS has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("hub.GAS_ACTUAL has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("hub.GAS_COST has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("hub.GAS_EXPECTED has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("hub.GAS_NEXT has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException(
          "hub.HAS_CODE_NEW_xor_MXP_MXPX_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_COPY_FLAG_xor_VALUE_ORIG_IS_ZERO has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException(
          "hub.HAS_CODE_xor_MXP_FLAG_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_CON_FLAG_xor_VALUE_NEXT_IS_ZERO has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("hub.HASH_INFO_STAMP has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("hub.HEIGHT has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("hub.HEIGHT_NEW has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("hub.HUB_STAMP has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("hub.HUB_STAMP_TRANSACTION_END has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException(
          "hub.IS_PRECOMPILE_xor_OOB_FLAG_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_CREATE_FLAG_xor_WARMTH has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("hub.LOG_INFO_STAMP has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException(
          "hub.MARKED_FOR_SELFDESTRUCT_NEW_xor_STP_FLAG_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DEC_FLAG_2 has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException(
          "hub.MARKED_FOR_SELFDESTRUCT_xor_STP_EXISTS_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DEC_FLAG_1_xor_WARMTH_NEW has not been filled");
    }

    if (!filled.get(103)) {
      throw new IllegalStateException("hub.MMU_AUX_ID_xor_PRC_CALLER_GAS has not been filled");
    }

    if (!filled.get(104)) {
      throw new IllegalStateException("hub.MMU_EXO_SUM_xor_PRC_CDO has not been filled");
    }

    if (!filled.get(105)) {
      throw new IllegalStateException("hub.MMU_INST_xor_PRC_CDS has not been filled");
    }

    if (!filled.get(113)) {
      throw new IllegalStateException("hub.MMU_LIMB_1 has not been filled");
    }

    if (!filled.get(114)) {
      throw new IllegalStateException("hub.MMU_LIMB_2 has not been filled");
    }

    if (!filled.get(106)) {
      throw new IllegalStateException("hub.MMU_PHASE_xor_PRC_RAC has not been filled");
    }

    if (!filled.get(107)) {
      throw new IllegalStateException("hub.MMU_REF_OFFSET_xor_PRC_RAO has not been filled");
    }

    if (!filled.get(108)) {
      throw new IllegalStateException("hub.MMU_REF_SIZE_xor_PRC_RETURN_GAS has not been filled");
    }

    if (!filled.get(109)) {
      throw new IllegalStateException("hub.MMU_SIZE has not been filled");
    }

    if (!filled.get(110)) {
      throw new IllegalStateException("hub.MMU_SRC_ID has not been filled");
    }

    if (!filled.get(115)) {
      throw new IllegalStateException("hub.MMU_SRC_OFFSET_HI has not been filled");
    }

    if (!filled.get(116)) {
      throw new IllegalStateException("hub.MMU_SRC_OFFSET_LO has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("hub.MMU_STAMP has not been filled");
    }

    if (!filled.get(111)) {
      throw new IllegalStateException("hub.MMU_TGT_ID has not been filled");
    }

    if (!filled.get(117)) {
      throw new IllegalStateException("hub.MMU_TGT_OFFSET_LO has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("hub.MXP_STAMP has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("hub.NON_STACK_ROWS has not been filled");
    }

    if (!filled.get(133)) {
      throw new IllegalStateException(
          "hub.NONCE_NEW_xor_CONTEXT_NUMBER_xor_MXP_SIZE_2_HI_xor_STACK_ITEM_STAMP_2_xor_REFUND_COUNTER_INFINITY has not been filled");
    }

    if (!filled.get(132)) {
      throw new IllegalStateException(
          "hub.NONCE_xor_CALL_VALUE_xor_MXP_SIZE_1_LO_xor_STACK_ITEM_STAMP_1_xor_PRIORITY_FEE_PER_GAS has not been filled");
    }

    if (!filled.get(142)) {
      throw new IllegalStateException(
          "hub.OOB_DATA_7_xor_STACK_ITEM_VALUE_LO_3 has not been filled");
    }

    if (!filled.get(143)) {
      throw new IllegalStateException(
          "hub.OOB_DATA_8_xor_STACK_ITEM_VALUE_LO_4 has not been filled");
    }

    if (!filled.get(112)) {
      throw new IllegalStateException("hub.OOB_INST has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("hub.PEEK_AT_ACCOUNT has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("hub.PEEK_AT_CONTEXT has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("hub.PEEK_AT_MISCELLANEOUS has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("hub.PEEK_AT_SCENARIO has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("hub.PEEK_AT_STACK has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("hub.PEEK_AT_STORAGE has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("hub.PEEK_AT_TRANSACTION has not been filled");
    }

    if (!filled.get(74)) {
      throw new IllegalStateException("hub.PRC_BLAKE2f_xor_KEC_FLAG has not been filled");
    }

    if (!filled.get(75)) {
      throw new IllegalStateException("hub.PRC_ECADD_xor_LOG_FLAG has not been filled");
    }

    if (!filled.get(76)) {
      throw new IllegalStateException("hub.PRC_ECMUL_xor_LOG_INFO_FLAG has not been filled");
    }

    if (!filled.get(77)) {
      throw new IllegalStateException(
          "hub.PRC_ECPAIRING_xor_MACHINE_STATE_FLAG has not been filled");
    }

    if (!filled.get(78)) {
      throw new IllegalStateException("hub.PRC_ECRECOVER_xor_MAXCSX has not been filled");
    }

    if (!filled.get(79)) {
      throw new IllegalStateException(
          "hub.PRC_FAILURE_KNOWN_TO_HUB_xor_MOD_FLAG has not been filled");
    }

    if (!filled.get(80)) {
      throw new IllegalStateException(
          "hub.PRC_FAILURE_KNOWN_TO_RAM_xor_MUL_FLAG has not been filled");
    }

    if (!filled.get(81)) {
      throw new IllegalStateException("hub.PRC_IDENTITY_xor_MXPX has not been filled");
    }

    if (!filled.get(82)) {
      throw new IllegalStateException("hub.PRC_MODEXP_xor_MXP_FLAG has not been filled");
    }

    if (!filled.get(83)) {
      throw new IllegalStateException("hub.PRC_RIPEMD-160_xor_OOGX has not been filled");
    }

    if (!filled.get(84)) {
      throw new IllegalStateException("hub.PRC_SHA2-256_xor_OPCX has not been filled");
    }

    if (!filled.get(85)) {
      throw new IllegalStateException(
          "hub.PRC_SUCCESS_WILL_REVERT_xor_PUSHPOP_FLAG has not been filled");
    }

    if (!filled.get(86)) {
      throw new IllegalStateException("hub.PRC_SUCCESS_WONT_REVERT_xor_RDCX has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("hub.PROGRAM_COUNTER_NEW has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("hub.REFUND_COUNTER has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("hub.REFUND_COUNTER_NEW has not been filled");
    }

    if (!filled.get(87)) {
      throw new IllegalStateException("hub.RETURN_EXCEPTION_xor_SHF_FLAG has not been filled");
    }

    if (!filled.get(88)) {
      throw new IllegalStateException(
          "hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WILL_REVERT_xor_SOX has not been filled");
    }

    if (!filled.get(89)) {
      throw new IllegalStateException(
          "hub.RETURN_FROM_DEPLOYMENT_EMPTY_CODE_WONT_REVERT_xor_SSTOREX has not been filled");
    }

    if (!filled.get(90)) {
      throw new IllegalStateException(
          "hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WILL_REVERT_xor_STACKRAM_FLAG has not been filled");
    }

    if (!filled.get(91)) {
      throw new IllegalStateException(
          "hub.RETURN_FROM_DEPLOYMENT_NONEMPTY_CODE_WONT_REVERT_xor_STACK_ITEM_POP_1 has not been filled");
    }

    if (!filled.get(92)) {
      throw new IllegalStateException(
          "hub.RETURN_FROM_MESSAGE_CALL_WILL_TOUCH_RAM_xor_STACK_ITEM_POP_2 has not been filled");
    }

    if (!filled.get(93)) {
      throw new IllegalStateException(
          "hub.RETURN_FROM_MESSAGE_CALL_WONT_TOUCH_RAM_xor_STACK_ITEM_POP_3 has not been filled");
    }

    if (!filled.get(134)) {
      throw new IllegalStateException(
          "hub.RLPADDR_DEP_ADDR_HI_xor_RETURN_AT_CAPACITY_xor_MXP_SIZE_2_LO_xor_STACK_ITEM_STAMP_3_xor_REFUND_EFFECTIVE has not been filled");
    }

    if (!filled.get(135)) {
      throw new IllegalStateException(
          "hub.RLPADDR_DEP_ADDR_LO_xor_RETURN_AT_OFFSET_xor_MXP_WORDS_xor_STACK_ITEM_STAMP_4_xor_TO_ADDRESS_HI has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException(
          "hub.RLPADDR_FLAG_xor_STP_OOGX_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DEC_FLAG_3 has not been filled");
    }

    if (!filled.get(136)) {
      throw new IllegalStateException(
          "hub.RLPADDR_KEC_HI_xor_RETURN_DATA_CONTEXT_NUMBER_xor_OOB_DATA_1_xor_STACK_ITEM_VALUE_HI_1_xor_TO_ADDRESS_LO has not been filled");
    }

    if (!filled.get(137)) {
      throw new IllegalStateException(
          "hub.RLPADDR_KEC_LO_xor_RETURN_DATA_OFFSET_xor_OOB_DATA_2_xor_STACK_ITEM_VALUE_HI_2_xor_VALUE has not been filled");
    }

    if (!filled.get(138)) {
      throw new IllegalStateException(
          "hub.RLPADDR_RECIPE_xor_RETURN_DATA_SIZE_xor_OOB_DATA_3_xor_STACK_ITEM_VALUE_HI_3 has not been filled");
    }

    if (!filled.get(139)) {
      throw new IllegalStateException(
          "hub.RLPADDR_SALT_HI_xor_OOB_DATA_4_xor_STACK_ITEM_VALUE_HI_4 has not been filled");
    }

    if (!filled.get(140)) {
      throw new IllegalStateException(
          "hub.RLPADDR_SALT_LO_xor_OOB_DATA_5_xor_STACK_ITEM_VALUE_LO_1 has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException(
          "hub.ROM_LEX_FLAG_xor_STP_WARMTH_xor_CREATE_ABORT_xor_DEC_FLAG_4 has not been filled");
    }

    if (!filled.get(94)) {
      throw new IllegalStateException(
          "hub.SELFDESTRUCT_EXCEPTION_xor_STACK_ITEM_POP_4 has not been filled");
    }

    if (!filled.get(95)) {
      throw new IllegalStateException(
          "hub.SELFDESTRUCT_WILL_REVERT_xor_STATICX has not been filled");
    }

    if (!filled.get(96)) {
      throw new IllegalStateException(
          "hub.SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED_xor_STATIC_FLAG has not been filled");
    }

    if (!filled.get(97)) {
      throw new IllegalStateException(
          "hub.SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED_xor_STO_FLAG has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("hub.sto_FINAL has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("hub.sto_FIRST has not been filled");
    }

    if (!filled.get(144)) {
      throw new IllegalStateException("hub.STP_GAS_HI_xor_STATIC_GAS has not been filled");
    }

    if (!filled.get(145)) {
      throw new IllegalStateException("hub.STP_GAS_LO has not been filled");
    }

    if (!filled.get(146)) {
      throw new IllegalStateException("hub.STP_GAS_MXP has not been filled");
    }

    if (!filled.get(147)) {
      throw new IllegalStateException("hub.STP_GAS_PAID_OUT_OF_POCKET has not been filled");
    }

    if (!filled.get(148)) {
      throw new IllegalStateException("hub.STP_GAS_STIPEND has not been filled");
    }

    if (!filled.get(149)) {
      throw new IllegalStateException("hub.STP_GAS_UPFRONT_GAS_COST has not been filled");
    }

    if (!filled.get(150)) {
      throw new IllegalStateException("hub.STP_INSTRUCTION has not been filled");
    }

    if (!filled.get(151)) {
      throw new IllegalStateException("hub.STP_VAL_HI has not been filled");
    }

    if (!filled.get(152)) {
      throw new IllegalStateException("hub.STP_VAL_LO has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("hub.SUB_STAMP has not been filled");
    }

    if (!filled.get(98)) {
      throw new IllegalStateException("hub.SUX has not been filled");
    }

    if (!filled.get(99)) {
      throw new IllegalStateException("hub.SWAP_FLAG has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("hub.TRANSACTION_REVERTS has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException(
          "hub.TRM_FLAG_xor_CREATE_EMPTY_INIT_CODE_WILL_REVERT_xor_DUP_FLAG has not been filled");
    }

    if (!filled.get(141)) {
      throw new IllegalStateException(
          "hub.TRM_RAW_ADDRESS_HI_xor_OOB_DATA_6_xor_STACK_ITEM_VALUE_LO_2 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("hub.TWO_LINE_INSTRUCTION has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("hub.TX_EXEC has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("hub.TX_FINL has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("hub.TX_INIT has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("hub.TX_SKIP has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("hub.TX_WARM has not been filled");
    }

    if (!filled.get(100)) {
      throw new IllegalStateException("hub.TXN_FLAG has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException(
          "hub.WARMTH_NEW_xor_CREATE_EXCEPTION_xor_HALT_FLAG has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException(
          "hub.WARMTH_xor_CREATE_EMPTY_INIT_CODE_WONT_REVERT_xor_EXT_FLAG has not been filled");
    }

    if (!filled.get(101)) {
      throw new IllegalStateException("hub.WCP_FLAG has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absoluteTransactionNumber.position(absoluteTransactionNumber.position() + 32);
    }

    if (!filled.get(47)) {
      accFinal.position(accFinal.position() + 1);
    }

    if (!filled.get(48)) {
      accFirst.position(accFirst.position() + 1);
    }

    if (!filled.get(118)) {
      addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.position(
          addressHiXorAccountAddressHiXorCcrsStampXorAlphaXorAddressHiXorBasefee.position() + 32);
    }

    if (!filled.get(119)) {
      addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.position(
          addressLoXorAccountAddressLoXorExpData1XorDeltaXorAddressLoXorBatchNum.position() + 32);
    }

    if (!filled.get(121)) {
      balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
          .position(
              balanceNewXorByteCodeAddressHiXorExpData3XorHashInfoKeccakLoXorDeploymentNumberInftyXorCoinbaseAddressHi
                      .position()
                  + 32);
    }

    if (!filled.get(120)) {
      balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
          .position(
              balanceXorAccountDeploymentNumberXorExpData2XorHashInfoKeccakHiXorDeploymentNumberXorCallDataSize
                      .position()
                  + 32);
    }

    if (!filled.get(1)) {
      batchNumber.position(batchNumber.position() + 32);
    }

    if (!filled.get(2)) {
      callerContextNumber.position(callerContextNumber.position() + 32);
    }

    if (!filled.get(3)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 32);
    }

    if (!filled.get(122)) {
      codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
          .position(
              codeFragmentIndexXorByteCodeAddressLoXorExpData4XorHashInfoSizeXorStorageKeyHiXorCoinbaseAddressLo
                      .position()
                  + 32);
    }

    if (!filled.get(124)) {
      codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
          .position(
              codeHashHiNewXorByteCodeDeploymentNumberXorMxpGasMxpXorNbAddedXorValueCurrHiXorFromAddressLo
                      .position()
                  + 32);
    }

    if (!filled.get(123)) {
      codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
          .position(
              codeHashHiXorByteCodeCodeFragmentIndexXorExpData5XorInstructionXorStorageKeyLoXorFromAddressHi
                      .position()
                  + 32);
    }

    if (!filled.get(126)) {
      codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover
          .position(
              codeHashLoNewXorCallerAddressHiXorMxpMtntopXorPushValueHiXorValueNextHiXorGasLeftover
                      .position()
                  + 32);
    }

    if (!filled.get(125)) {
      codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
          .position(
              codeHashLoXorByteCodeDeploymentStatusXorMxpInstXorNbRemovedXorValueCurrLoXorGasInitiallyAvailable
                      .position()
                  + 32);
    }

    if (!filled.get(128)) {
      codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
          .position(
              codeSizeNewXorCallDataContextNumberXorMxpOffset1LoXorStackItemHeight1XorValueOrigHiXorGasPrice
                      .position()
                  + 32);
    }

    if (!filled.get(127)) {
      codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit.position(
          codeSizeXorCallerAddressLoXorMxpOffset1HiXorPushValueLoXorValueNextLoXorGasLimit
                  .position()
              + 32);
    }

    if (!filled.get(49)) {
      conAgain.position(conAgain.position() + 1);
    }

    if (!filled.get(50)) {
      conFirst.position(conFirst.position() + 1);
    }

    if (!filled.get(4)) {
      contextGetsReverted.position(contextGetsReverted.position() + 1);
    }

    if (!filled.get(5)) {
      contextMayChange.position(contextMayChange.position() + 1);
    }

    if (!filled.get(6)) {
      contextNumber.position(contextNumber.position() + 32);
    }

    if (!filled.get(7)) {
      contextNumberNew.position(contextNumberNew.position() + 32);
    }

    if (!filled.get(8)) {
      contextRevertStamp.position(contextRevertStamp.position() + 32);
    }

    if (!filled.get(9)) {
      contextSelfReverts.position(contextSelfReverts.position() + 1);
    }

    if (!filled.get(10)) {
      contextWillRevert.position(contextWillRevert.position() + 1);
    }

    if (!filled.get(11)) {
      counterNsr.position(counterNsr.position() + 32);
    }

    if (!filled.get(12)) {
      counterTli.position(counterTli.position() + 1);
    }

    if (!filled.get(68)) {
      createFailureConditionWillRevertXorHashInfoFlag.position(
          createFailureConditionWillRevertXorHashInfoFlag.position() + 1);
    }

    if (!filled.get(69)) {
      createFailureConditionWontRevertXorIcpx.position(
          createFailureConditionWontRevertXorIcpx.position() + 1);
    }

    if (!filled.get(70)) {
      createNonemptyInitCodeFailureWillRevertXorInvalidFlag.position(
          createNonemptyInitCodeFailureWillRevertXorInvalidFlag.position() + 1);
    }

    if (!filled.get(71)) {
      createNonemptyInitCodeFailureWontRevertXorJumpx.position(
          createNonemptyInitCodeFailureWontRevertXorJumpx.position() + 1);
    }

    if (!filled.get(72)) {
      createNonemptyInitCodeSuccessWillRevertXorJumpDestinationVettingRequired.position(
          createNonemptyInitCodeSuccessWillRevertXorJumpDestinationVettingRequired.position() + 1);
    }

    if (!filled.get(73)) {
      createNonemptyInitCodeSuccessWontRevertXorJumpFlag.position(
          createNonemptyInitCodeSuccessWontRevertXorJumpFlag.position() + 1);
    }

    if (!filled.get(130)) {
      deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize
          .position(
              deploymentNumberInftyXorCallDataSizeXorMxpOffset2LoXorStackItemHeight3XorInitCodeSize
                      .position()
                  + 32);
    }

    if (!filled.get(131)) {
      deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.position(
          deploymentNumberNewXorCallStackDepthXorMxpSize1HiXorStackItemHeight4XorNonce.position()
              + 32);
    }

    if (!filled.get(129)) {
      deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
          .position(
              deploymentNumberXorCallDataOffsetXorMxpOffset2HiXorStackItemHeight2XorValueOrigLoXorInitialBalance
                      .position()
                  + 32);
    }

    if (!filled.get(54)) {
      deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
          .position(
              deploymentStatusInftyXorIsStaticXorExpFlagXorCallEoaSuccessCallerWillRevertXorAddFlagXorValueCurrIsOrigXorIsDeployment
                      .position()
                  + 1);
    }

    if (!filled.get(55)) {
      deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
          .position(
              deploymentStatusNewXorUpdateXorMmuFlagXorCallEoaSuccessCallerWontRevertXorBinFlagXorValueCurrIsZeroXorIsType2
                      .position()
                  + 1);
    }

    if (!filled.get(53)) {
      deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd
          .position(
              deploymentStatusXorIsRootXorCcsrFlagXorCallAbortXorAccFlagXorValueCurrChangesXorCopyTxcd
                      .position()
                  + 1);
    }

    if (!filled.get(13)) {
      domStamp.position(domStamp.position() + 32);
    }

    if (!filled.get(14)) {
      exceptionAhoy.position(exceptionAhoy.position() + 1);
    }

    if (!filled.get(57)) {
      existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode.position(
          existsNewXorMxpDeploysXorCallPrcFailureXorCallFlagXorValueNextIsOrigXorStatusCode
                  .position()
              + 1);
    }

    if (!filled.get(56)) {
      existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution
          .position(
              existsXorMmuSuccessBitXorCallExceptionXorBtcFlagXorValueNextIsCurrXorRequiresEvmExecution
                      .position()
                  + 1);
    }

    if (!filled.get(102)) {
      expInstXorPrcCalleeGas.position(expInstXorPrcCalleeGas.position() + 8);
    }

    if (!filled.get(15)) {
      gasActual.position(gasActual.position() + 32);
    }

    if (!filled.get(16)) {
      gasCost.position(gasCost.position() + 32);
    }

    if (!filled.get(17)) {
      gasExpected.position(gasExpected.position() + 32);
    }

    if (!filled.get(18)) {
      gasNext.position(gasNext.position() + 32);
    }

    if (!filled.get(59)) {
      hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero.position(
          hasCodeNewXorMxpMxpxXorCallPrcSuccessCallerWontRevertXorCopyFlagXorValueOrigIsZero
                  .position()
              + 1);
    }

    if (!filled.get(58)) {
      hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.position(
          hasCodeXorMxpFlagXorCallPrcSuccessCallerWillRevertXorConFlagXorValueNextIsZero.position()
              + 1);
    }

    if (!filled.get(19)) {
      hashInfoStamp.position(hashInfoStamp.position() + 32);
    }

    if (!filled.get(20)) {
      height.position(height.position() + 32);
    }

    if (!filled.get(21)) {
      heightNew.position(heightNew.position() + 32);
    }

    if (!filled.get(22)) {
      hubStamp.position(hubStamp.position() + 32);
    }

    if (!filled.get(23)) {
      hubStampTransactionEnd.position(hubStampTransactionEnd.position() + 32);
    }

    if (!filled.get(60)) {
      isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.position(
          isPrecompileXorOobFlagXorCallSmcFailureCallerWillRevertXorCreateFlagXorWarmth.position()
              + 1);
    }

    if (!filled.get(24)) {
      logInfoStamp.position(logInfoStamp.position() + 32);
    }

    if (!filled.get(62)) {
      markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2.position(
          markedForSelfdestructNewXorStpFlagXorCallSmcSuccessCallerWillRevertXorDecFlag2.position()
              + 1);
    }

    if (!filled.get(61)) {
      markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew
          .position(
              markedForSelfdestructXorStpExistsXorCallSmcFailureCallerWontRevertXorDecFlag1XorWarmthNew
                      .position()
                  + 1);
    }

    if (!filled.get(103)) {
      mmuAuxIdXorPrcCallerGas.position(mmuAuxIdXorPrcCallerGas.position() + 8);
    }

    if (!filled.get(104)) {
      mmuExoSumXorPrcCdo.position(mmuExoSumXorPrcCdo.position() + 8);
    }

    if (!filled.get(105)) {
      mmuInstXorPrcCds.position(mmuInstXorPrcCds.position() + 8);
    }

    if (!filled.get(113)) {
      mmuLimb1.position(mmuLimb1.position() + 32);
    }

    if (!filled.get(114)) {
      mmuLimb2.position(mmuLimb2.position() + 32);
    }

    if (!filled.get(106)) {
      mmuPhaseXorPrcRac.position(mmuPhaseXorPrcRac.position() + 8);
    }

    if (!filled.get(107)) {
      mmuRefOffsetXorPrcRao.position(mmuRefOffsetXorPrcRao.position() + 8);
    }

    if (!filled.get(108)) {
      mmuRefSizeXorPrcReturnGas.position(mmuRefSizeXorPrcReturnGas.position() + 8);
    }

    if (!filled.get(109)) {
      mmuSize.position(mmuSize.position() + 8);
    }

    if (!filled.get(110)) {
      mmuSrcId.position(mmuSrcId.position() + 8);
    }

    if (!filled.get(115)) {
      mmuSrcOffsetHi.position(mmuSrcOffsetHi.position() + 32);
    }

    if (!filled.get(116)) {
      mmuSrcOffsetLo.position(mmuSrcOffsetLo.position() + 32);
    }

    if (!filled.get(25)) {
      mmuStamp.position(mmuStamp.position() + 32);
    }

    if (!filled.get(111)) {
      mmuTgtId.position(mmuTgtId.position() + 8);
    }

    if (!filled.get(117)) {
      mmuTgtOffsetLo.position(mmuTgtOffsetLo.position() + 32);
    }

    if (!filled.get(26)) {
      mxpStamp.position(mxpStamp.position() + 32);
    }

    if (!filled.get(27)) {
      nonStackRows.position(nonStackRows.position() + 32);
    }

    if (!filled.get(133)) {
      nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.position(
          nonceNewXorContextNumberXorMxpSize2HiXorStackItemStamp2XorRefundCounterInfinity.position()
              + 32);
    }

    if (!filled.get(132)) {
      nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.position(
          nonceXorCallValueXorMxpSize1LoXorStackItemStamp1XorPriorityFeePerGas.position() + 32);
    }

    if (!filled.get(142)) {
      oobData7XorStackItemValueLo3.position(oobData7XorStackItemValueLo3.position() + 32);
    }

    if (!filled.get(143)) {
      oobData8XorStackItemValueLo4.position(oobData8XorStackItemValueLo4.position() + 32);
    }

    if (!filled.get(112)) {
      oobInst.position(oobInst.position() + 8);
    }

    if (!filled.get(28)) {
      peekAtAccount.position(peekAtAccount.position() + 1);
    }

    if (!filled.get(29)) {
      peekAtContext.position(peekAtContext.position() + 1);
    }

    if (!filled.get(30)) {
      peekAtMiscellaneous.position(peekAtMiscellaneous.position() + 1);
    }

    if (!filled.get(31)) {
      peekAtScenario.position(peekAtScenario.position() + 1);
    }

    if (!filled.get(32)) {
      peekAtStack.position(peekAtStack.position() + 1);
    }

    if (!filled.get(33)) {
      peekAtStorage.position(peekAtStorage.position() + 1);
    }

    if (!filled.get(34)) {
      peekAtTransaction.position(peekAtTransaction.position() + 1);
    }

    if (!filled.get(74)) {
      prcBlake2FXorKecFlag.position(prcBlake2FXorKecFlag.position() + 1);
    }

    if (!filled.get(75)) {
      prcEcaddXorLogFlag.position(prcEcaddXorLogFlag.position() + 1);
    }

    if (!filled.get(76)) {
      prcEcmulXorLogInfoFlag.position(prcEcmulXorLogInfoFlag.position() + 1);
    }

    if (!filled.get(77)) {
      prcEcpairingXorMachineStateFlag.position(prcEcpairingXorMachineStateFlag.position() + 1);
    }

    if (!filled.get(78)) {
      prcEcrecoverXorMaxcsx.position(prcEcrecoverXorMaxcsx.position() + 1);
    }

    if (!filled.get(79)) {
      prcFailureKnownToHubXorModFlag.position(prcFailureKnownToHubXorModFlag.position() + 1);
    }

    if (!filled.get(80)) {
      prcFailureKnownToRamXorMulFlag.position(prcFailureKnownToRamXorMulFlag.position() + 1);
    }

    if (!filled.get(81)) {
      prcIdentityXorMxpx.position(prcIdentityXorMxpx.position() + 1);
    }

    if (!filled.get(82)) {
      prcModexpXorMxpFlag.position(prcModexpXorMxpFlag.position() + 1);
    }

    if (!filled.get(83)) {
      prcRipemd160XorOogx.position(prcRipemd160XorOogx.position() + 1);
    }

    if (!filled.get(84)) {
      prcSha2256XorOpcx.position(prcSha2256XorOpcx.position() + 1);
    }

    if (!filled.get(85)) {
      prcSuccessWillRevertXorPushpopFlag.position(
          prcSuccessWillRevertXorPushpopFlag.position() + 1);
    }

    if (!filled.get(86)) {
      prcSuccessWontRevertXorRdcx.position(prcSuccessWontRevertXorRdcx.position() + 1);
    }

    if (!filled.get(35)) {
      programCounter.position(programCounter.position() + 32);
    }

    if (!filled.get(36)) {
      programCounterNew.position(programCounterNew.position() + 32);
    }

    if (!filled.get(37)) {
      refundCounter.position(refundCounter.position() + 32);
    }

    if (!filled.get(38)) {
      refundCounterNew.position(refundCounterNew.position() + 32);
    }

    if (!filled.get(87)) {
      returnExceptionXorShfFlag.position(returnExceptionXorShfFlag.position() + 1);
    }

    if (!filled.get(88)) {
      returnFromDeploymentEmptyCodeWillRevertXorSox.position(
          returnFromDeploymentEmptyCodeWillRevertXorSox.position() + 1);
    }

    if (!filled.get(89)) {
      returnFromDeploymentEmptyCodeWontRevertXorSstorex.position(
          returnFromDeploymentEmptyCodeWontRevertXorSstorex.position() + 1);
    }

    if (!filled.get(90)) {
      returnFromDeploymentNonemptyCodeWillRevertXorStackramFlag.position(
          returnFromDeploymentNonemptyCodeWillRevertXorStackramFlag.position() + 1);
    }

    if (!filled.get(91)) {
      returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop1.position(
          returnFromDeploymentNonemptyCodeWontRevertXorStackItemPop1.position() + 1);
    }

    if (!filled.get(92)) {
      returnFromMessageCallWillTouchRamXorStackItemPop2.position(
          returnFromMessageCallWillTouchRamXorStackItemPop2.position() + 1);
    }

    if (!filled.get(93)) {
      returnFromMessageCallWontTouchRamXorStackItemPop3.position(
          returnFromMessageCallWontTouchRamXorStackItemPop3.position() + 1);
    }

    if (!filled.get(134)) {
      rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective.position(
          rlpaddrDepAddrHiXorReturnAtCapacityXorMxpSize2LoXorStackItemStamp3XorRefundEffective
                  .position()
              + 32);
    }

    if (!filled.get(135)) {
      rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.position(
          rlpaddrDepAddrLoXorReturnAtOffsetXorMxpWordsXorStackItemStamp4XorToAddressHi.position()
              + 32);
    }

    if (!filled.get(63)) {
      rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3.position(
          rlpaddrFlagXorStpOogxXorCallSmcSuccessCallerWontRevertXorDecFlag3.position() + 1);
    }

    if (!filled.get(136)) {
      rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo.position(
          rlpaddrKecHiXorReturnDataContextNumberXorOobData1XorStackItemValueHi1XorToAddressLo
                  .position()
              + 32);
    }

    if (!filled.get(137)) {
      rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.position(
          rlpaddrKecLoXorReturnDataOffsetXorOobData2XorStackItemValueHi2XorValue.position() + 32);
    }

    if (!filled.get(138)) {
      rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.position(
          rlpaddrRecipeXorReturnDataSizeXorOobData3XorStackItemValueHi3.position() + 32);
    }

    if (!filled.get(139)) {
      rlpaddrSaltHiXorOobData4XorStackItemValueHi4.position(
          rlpaddrSaltHiXorOobData4XorStackItemValueHi4.position() + 32);
    }

    if (!filled.get(140)) {
      rlpaddrSaltLoXorOobData5XorStackItemValueLo1.position(
          rlpaddrSaltLoXorOobData5XorStackItemValueLo1.position() + 32);
    }

    if (!filled.get(64)) {
      romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4.position(
          romLexFlagXorStpWarmthXorCreateAbortXorDecFlag4.position() + 1);
    }

    if (!filled.get(94)) {
      selfdestructExceptionXorStackItemPop4.position(
          selfdestructExceptionXorStackItemPop4.position() + 1);
    }

    if (!filled.get(95)) {
      selfdestructWillRevertXorStaticx.position(selfdestructWillRevertXorStaticx.position() + 1);
    }

    if (!filled.get(96)) {
      selfdestructWontRevertAlreadyMarkedXorStaticFlag.position(
          selfdestructWontRevertAlreadyMarkedXorStaticFlag.position() + 1);
    }

    if (!filled.get(97)) {
      selfdestructWontRevertNotYetMarkedXorStoFlag.position(
          selfdestructWontRevertNotYetMarkedXorStoFlag.position() + 1);
    }

    if (!filled.get(51)) {
      stoFinal.position(stoFinal.position() + 1);
    }

    if (!filled.get(52)) {
      stoFirst.position(stoFirst.position() + 1);
    }

    if (!filled.get(144)) {
      stpGasHiXorStaticGas.position(stpGasHiXorStaticGas.position() + 32);
    }

    if (!filled.get(145)) {
      stpGasLo.position(stpGasLo.position() + 32);
    }

    if (!filled.get(146)) {
      stpGasMxp.position(stpGasMxp.position() + 32);
    }

    if (!filled.get(147)) {
      stpGasPaidOutOfPocket.position(stpGasPaidOutOfPocket.position() + 32);
    }

    if (!filled.get(148)) {
      stpGasStipend.position(stpGasStipend.position() + 32);
    }

    if (!filled.get(149)) {
      stpGasUpfrontGasCost.position(stpGasUpfrontGasCost.position() + 32);
    }

    if (!filled.get(150)) {
      stpInstruction.position(stpInstruction.position() + 32);
    }

    if (!filled.get(151)) {
      stpValHi.position(stpValHi.position() + 32);
    }

    if (!filled.get(152)) {
      stpValLo.position(stpValLo.position() + 32);
    }

    if (!filled.get(39)) {
      subStamp.position(subStamp.position() + 32);
    }

    if (!filled.get(98)) {
      sux.position(sux.position() + 1);
    }

    if (!filled.get(99)) {
      swapFlag.position(swapFlag.position() + 1);
    }

    if (!filled.get(40)) {
      transactionReverts.position(transactionReverts.position() + 1);
    }

    if (!filled.get(65)) {
      trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag.position(
          trmFlagXorCreateEmptyInitCodeWillRevertXorDupFlag.position() + 1);
    }

    if (!filled.get(141)) {
      trmRawAddressHiXorOobData6XorStackItemValueLo2.position(
          trmRawAddressHiXorOobData6XorStackItemValueLo2.position() + 32);
    }

    if (!filled.get(41)) {
      twoLineInstruction.position(twoLineInstruction.position() + 1);
    }

    if (!filled.get(42)) {
      txExec.position(txExec.position() + 1);
    }

    if (!filled.get(43)) {
      txFinl.position(txFinl.position() + 1);
    }

    if (!filled.get(44)) {
      txInit.position(txInit.position() + 1);
    }

    if (!filled.get(45)) {
      txSkip.position(txSkip.position() + 1);
    }

    if (!filled.get(46)) {
      txWarm.position(txWarm.position() + 1);
    }

    if (!filled.get(100)) {
      txnFlag.position(txnFlag.position() + 1);
    }

    if (!filled.get(67)) {
      warmthNewXorCreateExceptionXorHaltFlag.position(
          warmthNewXorCreateExceptionXorHaltFlag.position() + 1);
    }

    if (!filled.get(66)) {
      warmthXorCreateEmptyInitCodeWontRevertXorExtFlag.position(
          warmthXorCreateEmptyInitCodeWontRevertXorExtFlag.position() + 1);
    }

    if (!filled.get(101)) {
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
