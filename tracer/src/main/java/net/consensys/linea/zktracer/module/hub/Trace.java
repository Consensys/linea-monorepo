/*
 * Copyright ConsenSys AG.
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
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public record Trace(
    @JsonProperty("ABSOLUTE_TRANSACTION_NUMBER") List<BigInteger> absoluteTransactionNumber,
    @JsonProperty(
            "ADDR_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_HASH_INFO___KEC_HI_xor_ADDRESS_HI_xor_BASEFEE")
        List<BigInteger>
            addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee,
    @JsonProperty(
            "ADDR_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP___DYNCOST_xor_HASH_INFO___KEC_LO_xor_ADDRESS_LO_xor_CALL_DATA_SIZE")
        List<BigInteger>
            addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize,
    @JsonProperty(
            "BALANCE_NEW_xor_BYTE_CODE_ADDRESS_HI_xor_EXP___EXPONENT_LO_xor_HEIGHT_xor_STORAGE_KEY_HI_xor_COINBASE_ADDRESS_LO")
        List<BigInteger>
            balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo,
    @JsonProperty(
            "BALANCE_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP___EXPONENT_HI_xor_HASH_INFO___SIZE_xor_DEPLOYMENT_NUMBER_xor_COINBASE_ADDRESS_HI")
        List<BigInteger>
            balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi,
    @JsonProperty("BATCH_NUMBER") List<BigInteger> batchNumber,
    @JsonProperty("CALLER_CONTEXT_NUMBER") List<BigInteger> callerContextNumber,
    @JsonProperty("CODE_ADDRESS_HI") List<BigInteger> codeAddressHi,
    @JsonProperty("CODE_ADDRESS_LO") List<BigInteger> codeAddressLo,
    @JsonProperty("CODE_DEPLOYMENT_NUMBER") List<BigInteger> codeDeploymentNumber,
    @JsonProperty("CODE_DEPLOYMENT_STATUS") List<Boolean> codeDeploymentStatus,
    @JsonProperty("CODE_FRAGMENT_INDEX") List<BigInteger> codeFragmentIndex,
    @JsonProperty(
            "CODE_HASH_HI_NEW_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MMU___INST_xor_HEIGHT_OVER_xor_VAL_CURR_HI_xor_FROM_ADDRESS_LO")
        List<BigInteger>
            codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo,
    @JsonProperty(
            "CODE_HASH_HI_xor_BYTE_CODE_ADDRESS_LO_xor_MMU___EXO_SUM_xor_HEIGHT_NEW_xor_STORAGE_KEY_LO_xor_FROM_ADDRESS_HI")
        List<BigInteger>
            codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi,
    @JsonProperty(
            "CODE_HASH_LO_NEW_xor_CALLER_ADDRESS_HI_xor_MMU___OFFSET_2_HI_xor_INST_xor_VAL_NEXT_HI_xor_GAS_PRICE")
        List<BigInteger>
            codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice,
    @JsonProperty(
            "CODE_HASH_LO_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MMU___OFFSET_1_LO_xor_HEIGHT_UNDER_xor_VAL_CURR_LO_xor_GAS_LIMIT")
        List<BigInteger>
            codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit,
    @JsonProperty(
            "CODE_SIZE_NEW_xor_CALLER_CONTEXT_NUMBER_xor_MMU___PARAM_1_xor_PUSH_VALUE_LO_xor_VAL_ORIG_HI_xor_GAS_REFUND_COUNTER_FINAL")
        List<BigInteger>
            codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal,
    @JsonProperty(
            "CODE_SIZE_xor_CALLER_ADDRESS_LO_xor_MMU___OFFSET_2_LO_xor_PUSH_VALUE_HI_xor_VAL_NEXT_LO_xor_GAS_REFUND_AMOUNT")
        List<BigInteger>
            codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount,
    @JsonProperty("CONTEXT_GETS_REVERTED_FLAG") List<Boolean> contextGetsRevertedFlag,
    @JsonProperty("CONTEXT_MAY_CHANGE_FLAG") List<Boolean> contextMayChangeFlag,
    @JsonProperty("CONTEXT_NUMBER") List<BigInteger> contextNumber,
    @JsonProperty("CONTEXT_NUMBER_NEW") List<BigInteger> contextNumberNew,
    @JsonProperty("CONTEXT_REVERT_STAMP") List<BigInteger> contextRevertStamp,
    @JsonProperty("CONTEXT_SELF_REVERTS_FLAG") List<Boolean> contextSelfRevertsFlag,
    @JsonProperty("CONTEXT_WILL_REVERT_FLAG") List<Boolean> contextWillRevertFlag,
    @JsonProperty("COUNTER_NSR") List<BigInteger> counterNsr,
    @JsonProperty("COUNTER_TLI") List<Boolean> counterTli,
    @JsonProperty(
            "DEP_NUM_NEW_xor_CALL_STACK_DEPTH_xor_MMU___REF_SIZE_xor_STACK_ITEM_HEIGHT_3_xor_INIT_GAS")
        List<BigInteger> depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas,
    @JsonProperty(
            "DEP_NUM_xor_CALL_DATA_SIZE_xor_MMU___REF_OFFSET_xor_STACK_ITEM_HEIGHT_2_xor_INIT_CODE_SIZE")
        List<BigInteger> depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize,
    @JsonProperty(
            "DEP_STATUS_NEW_xor_EXP___FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_BIN_FLAG_xor_VAL_CURR_IS_ZERO_xor_STATUS_CODE")
        List<Boolean>
            depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode,
    @JsonProperty(
            "DEP_STATUS_xor_CCSR_FLAG_xor_CALL_ABORT_xor_ADD_FLAG_xor_VAL_CURR_IS_ORIG_xor_IS_EIP1559")
        List<Boolean> depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559,
    @JsonProperty(
            "DEPLOYMENT_NUMBER_INFTY_xor_CALL_DATA_OFFSET_xor_MMU___PARAM_2_xor_STACK_ITEM_HEIGHT_1_xor_VAL_ORIG_LO_xor_INITIAL_BALANCE")
        List<BigInteger>
            deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance,
    @JsonProperty(
            "DEPLOYMENT_STATUS_INFTY_xor_UPDATE_xor_ABORT_FLAG_xor_BLAKE2f_xor_ACC_FLAG_xor_VAL_CURR_CHANGES_xor_IS_DEPLOYMENT")
        List<Boolean>
            deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment,
    @JsonProperty("DOM_STAMP") List<BigInteger> domStamp,
    @JsonProperty("EXCEPTION_AHOY_FLAG") List<Boolean> exceptionAhoyFlag,
    @JsonProperty(
            "EXISTS_NEW_xor_MMU___FLAG_xor_CALL_PRC_FAILURE_CALLER_WILL_REVERT_xor_CALL_FLAG_xor_VAL_NEXT_IS_ORIG")
        List<Boolean>
            existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig,
    @JsonProperty(
            "EXISTS_xor_FCOND_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BTC_FLAG_xor_VAL_NEXT_IS_CURR_xor_TXN_REQUIRES_EVM_EXECUTION")
        List<Boolean>
            existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution,
    @JsonProperty("GAS_ACTUAL") List<BigInteger> gasActual,
    @JsonProperty("GAS_COST") List<BigInteger> gasCost,
    @JsonProperty("GAS_EXPECTED") List<BigInteger> gasExpected,
    @JsonProperty("GAS_NEXT") List<BigInteger> gasNext,
    @JsonProperty("GAS_REFUND") List<BigInteger> gasRefund,
    @JsonProperty("GAS_REFUND_NEW") List<BigInteger> gasRefundNew,
    @JsonProperty(
            "HAS_CODE_NEW_xor_MXP___DEPLOYS_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_COPY_FLAG_xor_VAL_ORIG_IS_ZERO")
        List<Boolean>
            hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero,
    @JsonProperty(
            "HAS_CODE_xor_MMU___INFO_xor_CALL_PRC_FAILURE_CALLER_WONT_REVERT_xor_CON_FLAG_xor_VAL_NEXT_IS_ZERO")
        List<Boolean> hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero,
    @JsonProperty("HASH_INFO_STAMP") List<BigInteger> hashInfoStamp,
    @JsonProperty("HUB_STAMP") List<BigInteger> hubStamp,
    @JsonProperty("HUB_STAMP_TRANSACTION_END") List<BigInteger> hubStampTransactionEnd,
    @JsonProperty(
            "IS_BLAKE2f_xor_MXP___FLAG_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_CREATE_FLAG_xor_WARM")
        List<Boolean> isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm,
    @JsonProperty(
            "IS_ECADD_xor_MXP___MXPX_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_DECODED_FLAG_1_xor_WARM_NEW")
        List<Boolean> isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew,
    @JsonProperty(
            "IS_ECMUL_xor_OOB___EVENT_1_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DECODED_FLAG_2")
        List<Boolean> isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2,
    @JsonProperty(
            "IS_ECPAIRING_xor_OOB___EVENT_2_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DECODED_FLAG_3")
        List<Boolean> isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3,
    @JsonProperty(
            "IS_ECRECOVER_xor_OOB___FLAG_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DECODED_FLAG_4")
        List<Boolean> isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4,
    @JsonProperty("IS_IDENTITY_xor_PRECINFO___FLAG_xor_CODEDEPOSIT_xor_DUP_FLAG")
        List<Boolean> isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag,
    @JsonProperty("IS_MODEXP_xor_STP___EXISTS_xor_CODEDEPOSIT_INVALID_CODE_PREFIX_xor_EXT_FLAG")
        List<Boolean> isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag,
    @JsonProperty("IS_PRECOMPILE_xor_STP___FLAG_xor_CODEDEPOSIT_VALID_CODE_PREFIX_xor_HALT_FLAG")
        List<Boolean> isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag,
    @JsonProperty("IS_RIPEMDsub160_xor_STP___OOGX_xor_ECADD_xor_HASH_INFO_FLAG")
        List<Boolean> isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag,
    @JsonProperty("IS_SHA2sub256_xor_STP___WARM_xor_ECMUL_xor_INVALID_FLAG")
        List<Boolean> isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag,
    @JsonProperty("MMU_STAMP") List<BigInteger> mmuStamp,
    @JsonProperty("MXP___SIZE_1_HI_xor_STACK_ITEM_VALUE_LO_2")
        List<BigInteger> mxpSize1HiXorStackItemValueLo2,
    @JsonProperty("MXP___SIZE_1_LO_xor_STACK_ITEM_VALUE_LO_3")
        List<BigInteger> mxpSize1LoXorStackItemValueLo3,
    @JsonProperty("MXP___SIZE_2_HI_xor_STACK_ITEM_VALUE_LO_4")
        List<BigInteger> mxpSize2HiXorStackItemValueLo4,
    @JsonProperty("MXP___SIZE_2_LO_xor_STATIC_GAS") List<BigInteger> mxpSize2LoXorStaticGas,
    @JsonProperty("MXP_STAMP") List<BigInteger> mxpStamp,
    @JsonProperty("MXP___WORDS") List<BigInteger> mxpWords,
    @JsonProperty("NONCE_NEW_xor_CONTEXT_NUMBER_xor_MMU___SIZE_xor_STACK_ITEM_STAMP_1_xor_NONCE")
        List<BigInteger> nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce,
    @JsonProperty(
            "NONCE_xor_CALL_VALUE_xor_MMU___RETURNER_xor_STACK_ITEM_HEIGHT_4_xor_LEFTOVER_GAS")
        List<BigInteger> nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas,
    @JsonProperty("NUMBER_OF_NON_STACK_ROWS") List<BigInteger> numberOfNonStackRows,
    @JsonProperty("OOB___INST") List<BigInteger> oobInst,
    @JsonProperty("OOB___OUTGOING_DATA_1") List<BigInteger> oobOutgoingData1,
    @JsonProperty("OOB___OUTGOING_DATA_2") List<BigInteger> oobOutgoingData2,
    @JsonProperty("OOB___OUTGOING_DATA_3") List<BigInteger> oobOutgoingData3,
    @JsonProperty("OOB___OUTGOING_DATA_4") List<BigInteger> oobOutgoingData4,
    @JsonProperty("OOB___OUTGOING_DATA_5") List<BigInteger> oobOutgoingData5,
    @JsonProperty("OOB___OUTGOING_DATA_6") List<BigInteger> oobOutgoingData6,
    @JsonProperty("PEEK_AT_ACCOUNT") List<Boolean> peekAtAccount,
    @JsonProperty("PEEK_AT_CONTEXT") List<Boolean> peekAtContext,
    @JsonProperty("PEEK_AT_MISCELLANEOUS") List<Boolean> peekAtMiscellaneous,
    @JsonProperty("PEEK_AT_SCENARIO") List<Boolean> peekAtScenario,
    @JsonProperty("PEEK_AT_STACK") List<Boolean> peekAtStack,
    @JsonProperty("PEEK_AT_STORAGE") List<Boolean> peekAtStorage,
    @JsonProperty("PEEK_AT_TRANSACTION") List<Boolean> peekAtTransaction,
    @JsonProperty("PRECINFO___ADDR_LO") List<BigInteger> precinfoAddrLo,
    @JsonProperty("PRECINFO___CDS") List<BigInteger> precinfoCds,
    @JsonProperty("PRECINFO___EXEC_COST") List<BigInteger> precinfoExecCost,
    @JsonProperty("PRECINFO___PROVIDES_RETURN_DATA") List<BigInteger> precinfoProvidesReturnData,
    @JsonProperty("PRECINFO___RDS") List<BigInteger> precinfoRds,
    @JsonProperty("PRECINFO___SUCCESS") List<BigInteger> precinfoSuccess,
    @JsonProperty("PRECINFO___TOUCHES_RAM") List<BigInteger> precinfoTouchesRam,
    @JsonProperty("PROGRAM_COUNTER") List<BigInteger> programCounter,
    @JsonProperty("PROGRAM_COUNTER_NEW") List<BigInteger> programCounterNew,
    @JsonProperty("PUSHPOP_FLAG") List<Boolean> pushpopFlag,
    @JsonProperty("RDCX") List<Boolean> rdcx,
    @JsonProperty("RIPEMDsub160_xor_KEC_FLAG") List<Boolean> ripemDsub160XorKecFlag,
    @JsonProperty(
            "RLPADDR___DEP_ADDR_HI_xor_IS_STATIC_xor_MMU___STACK_VAL_HI_xor_STACK_ITEM_STAMP_2_xor_TO_ADDRESS_HI")
        List<BigInteger>
            rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi,
    @JsonProperty(
            "RLPADDR___DEP_ADDR_LO_xor_RETURNER_CONTEXT_NUMBER_xor_MMU___STACK_VAL_LO_xor_STACK_ITEM_STAMP_3_xor_TO_ADDRESS_LO")
        List<BigInteger>
            rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo,
    @JsonProperty("RLPADDR___FLAG_xor_ECPAIRING_xor_INVPREX")
        List<Boolean> rlpaddrFlagXorEcpairingXorInvprex,
    @JsonProperty(
            "RLPADDR___KEC_HI_xor_RETURNER_IS_PRECOMPILE_xor_MXP___GAS_MXP_xor_STACK_ITEM_STAMP_4_xor_VALUE")
        List<BigInteger> rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue,
    @JsonProperty("RLPADDR___KEC_LO_xor_RETURN_AT_OFFSET_xor_MXP___INST_xor_STACK_ITEM_VALUE_HI_1")
        List<BigInteger> rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1,
    @JsonProperty(
            "RLPADDR___RECIPE_xor_RETURN_AT_SIZE_xor_MXP___OFFSET_1_HI_xor_STACK_ITEM_VALUE_HI_2")
        List<BigInteger> rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2,
    @JsonProperty(
            "RLPADDR___SALT_HI_xor_RETURN_DATA_OFFSET_xor_MXP___OFFSET_1_LO_xor_STACK_ITEM_VALUE_HI_3")
        List<BigInteger> rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3,
    @JsonProperty(
            "RLPADDR___SALT_LO_xor_RETURN_DATA_SIZE_xor_MXP___OFFSET_2_HI_xor_STACK_ITEM_VALUE_HI_4")
        List<BigInteger> rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4,
    @JsonProperty("SCN_FAILURE_1_xor_LOG_FLAG") List<Boolean> scnFailure1XorLogFlag,
    @JsonProperty("SCN_FAILURE_2_xor_MACHINE_STATE_FLAG")
        List<Boolean> scnFailure2XorMachineStateFlag,
    @JsonProperty("SCN_FAILURE_3_xor_MAXCSX") List<Boolean> scnFailure3XorMaxcsx,
    @JsonProperty("SCN_FAILURE_4_xor_MOD_FLAG") List<Boolean> scnFailure4XorModFlag,
    @JsonProperty("SCN_SUCCESS_1_xor_MUL_FLAG") List<Boolean> scnSuccess1XorMulFlag,
    @JsonProperty("SCN_SUCCESS_2_xor_MXPX") List<Boolean> scnSuccess2XorMxpx,
    @JsonProperty("SCN_SUCCESS_3_xor_MXP_FLAG") List<Boolean> scnSuccess3XorMxpFlag,
    @JsonProperty("SCN_SUCCESS_4_xor_OOB_FLAG") List<Boolean> scnSuccess4XorOobFlag,
    @JsonProperty("SELFDESTRUCT_xor_OOGX") List<Boolean> selfdestructXorOogx,
    @JsonProperty("SHA2sub256_xor_OPCX") List<Boolean> sha2Sub256XorOpcx,
    @JsonProperty("SHF_FLAG") List<Boolean> shfFlag,
    @JsonProperty("SOX") List<Boolean> sox,
    @JsonProperty("SSTOREX") List<Boolean> sstorex,
    @JsonProperty("STACK_ITEM_POP_1") List<Boolean> stackItemPop1,
    @JsonProperty("STACK_ITEM_POP_2") List<Boolean> stackItemPop2,
    @JsonProperty("STACK_ITEM_POP_3") List<Boolean> stackItemPop3,
    @JsonProperty("STACK_ITEM_POP_4") List<Boolean> stackItemPop4,
    @JsonProperty("STACKRAM_FLAG") List<Boolean> stackramFlag,
    @JsonProperty("STATIC_FLAG") List<Boolean> staticFlag,
    @JsonProperty("STATICX") List<Boolean> staticx,
    @JsonProperty("STO_FLAG") List<Boolean> stoFlag,
    @JsonProperty("STP___GAS_HI") List<BigInteger> stpGasHi,
    @JsonProperty("STP___GAS_LO") List<BigInteger> stpGasLo,
    @JsonProperty("STP___GAS_OOPKT") List<BigInteger> stpGasOopkt,
    @JsonProperty("STP___GAS_STPD") List<BigInteger> stpGasStpd,
    @JsonProperty("STP___INST") List<BigInteger> stpInst,
    @JsonProperty("STP___VAL_HI") List<BigInteger> stpValHi,
    @JsonProperty("STP___VAL_LO") List<BigInteger> stpValLo,
    @JsonProperty("SUB_STAMP") List<BigInteger> subStamp,
    @JsonProperty("SUX") List<Boolean> sux,
    @JsonProperty("SWAP_FLAG") List<Boolean> swapFlag,
    @JsonProperty("TRANSACTION_REVERTS") List<Boolean> transactionReverts,
    @JsonProperty("TRM_FLAG") List<Boolean> trmFlag,
    @JsonProperty("TRM___FLAG_xor_ECRECOVER_xor_JUMPX") List<Boolean> trmFlagXorEcrecoverXorJumpx,
    @JsonProperty("TRM___RAW_ADDR_HI_xor_MXP___OFFSET_2_LO_xor_STACK_ITEM_VALUE_LO_1")
        List<BigInteger> trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1,
    @JsonProperty("TWO_LINE_INSTRUCTION") List<Boolean> twoLineInstruction,
    @JsonProperty("TX_EXEC") List<Boolean> txExec,
    @JsonProperty("TX_FINL") List<Boolean> txFinl,
    @JsonProperty("TX_INIT") List<Boolean> txInit,
    @JsonProperty("TX_SKIP") List<Boolean> txSkip,
    @JsonProperty("TX_WARM") List<Boolean> txWarm,
    @JsonProperty("TXN_FLAG") List<Boolean> txnFlag,
    @JsonProperty("WARM_NEW_xor_MODEXP_xor_JUMP_FLAG") List<Boolean> warmNewXorModexpXorJumpFlag,
    @JsonProperty("WARM_xor_IDENTITY_xor_JUMP_DESTINATION_VETTING_REQUIRED")
        List<Boolean> warmXorIdentityXorJumpDestinationVettingRequired,
    @JsonProperty("WCP_FLAG") List<Boolean> wcpFlag) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  public int size() {
    return this.absoluteTransactionNumber.size();
  }

  public static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ABSOLUTE_TRANSACTION_NUMBER")
    private final List<BigInteger> absoluteTransactionNumber = new ArrayList<>();

    @JsonProperty(
        "ADDR_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_HASH_INFO___KEC_HI_xor_ADDRESS_HI_xor_BASEFEE")
    private final List<BigInteger>
        addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee =
            new ArrayList<>();

    @JsonProperty(
        "ADDR_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP___DYNCOST_xor_HASH_INFO___KEC_LO_xor_ADDRESS_LO_xor_CALL_DATA_SIZE")
    private final List<BigInteger>
        addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize =
            new ArrayList<>();

    @JsonProperty(
        "BALANCE_NEW_xor_BYTE_CODE_ADDRESS_HI_xor_EXP___EXPONENT_LO_xor_HEIGHT_xor_STORAGE_KEY_HI_xor_COINBASE_ADDRESS_LO")
    private final List<BigInteger>
        balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo =
            new ArrayList<>();

    @JsonProperty(
        "BALANCE_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP___EXPONENT_HI_xor_HASH_INFO___SIZE_xor_DEPLOYMENT_NUMBER_xor_COINBASE_ADDRESS_HI")
    private final List<BigInteger>
        balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi =
            new ArrayList<>();

    @JsonProperty("BATCH_NUMBER")
    private final List<BigInteger> batchNumber = new ArrayList<>();

    @JsonProperty("CALLER_CONTEXT_NUMBER")
    private final List<BigInteger> callerContextNumber = new ArrayList<>();

    @JsonProperty("CODE_ADDRESS_HI")
    private final List<BigInteger> codeAddressHi = new ArrayList<>();

    @JsonProperty("CODE_ADDRESS_LO")
    private final List<BigInteger> codeAddressLo = new ArrayList<>();

    @JsonProperty("CODE_DEPLOYMENT_NUMBER")
    private final List<BigInteger> codeDeploymentNumber = new ArrayList<>();

    @JsonProperty("CODE_DEPLOYMENT_STATUS")
    private final List<Boolean> codeDeploymentStatus = new ArrayList<>();

    @JsonProperty("CODE_FRAGMENT_INDEX")
    private final List<BigInteger> codeFragmentIndex = new ArrayList<>();

    @JsonProperty(
        "CODE_HASH_HI_NEW_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MMU___INST_xor_HEIGHT_OVER_xor_VAL_CURR_HI_xor_FROM_ADDRESS_LO")
    private final List<BigInteger>
        codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo =
            new ArrayList<>();

    @JsonProperty(
        "CODE_HASH_HI_xor_BYTE_CODE_ADDRESS_LO_xor_MMU___EXO_SUM_xor_HEIGHT_NEW_xor_STORAGE_KEY_LO_xor_FROM_ADDRESS_HI")
    private final List<BigInteger>
        codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi =
            new ArrayList<>();

    @JsonProperty(
        "CODE_HASH_LO_NEW_xor_CALLER_ADDRESS_HI_xor_MMU___OFFSET_2_HI_xor_INST_xor_VAL_NEXT_HI_xor_GAS_PRICE")
    private final List<BigInteger>
        codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice =
            new ArrayList<>();

    @JsonProperty(
        "CODE_HASH_LO_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MMU___OFFSET_1_LO_xor_HEIGHT_UNDER_xor_VAL_CURR_LO_xor_GAS_LIMIT")
    private final List<BigInteger>
        codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit =
            new ArrayList<>();

    @JsonProperty(
        "CODE_SIZE_NEW_xor_CALLER_CONTEXT_NUMBER_xor_MMU___PARAM_1_xor_PUSH_VALUE_LO_xor_VAL_ORIG_HI_xor_GAS_REFUND_COUNTER_FINAL")
    private final List<BigInteger>
        codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal =
            new ArrayList<>();

    @JsonProperty(
        "CODE_SIZE_xor_CALLER_ADDRESS_LO_xor_MMU___OFFSET_2_LO_xor_PUSH_VALUE_HI_xor_VAL_NEXT_LO_xor_GAS_REFUND_AMOUNT")
    private final List<BigInteger>
        codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount =
            new ArrayList<>();

    @JsonProperty("CONTEXT_GETS_REVERTED_FLAG")
    private final List<Boolean> contextGetsRevertedFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_MAY_CHANGE_FLAG")
    private final List<Boolean> contextMayChangeFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_NUMBER")
    private final List<BigInteger> contextNumber = new ArrayList<>();

    @JsonProperty("CONTEXT_NUMBER_NEW")
    private final List<BigInteger> contextNumberNew = new ArrayList<>();

    @JsonProperty("CONTEXT_REVERT_STAMP")
    private final List<BigInteger> contextRevertStamp = new ArrayList<>();

    @JsonProperty("CONTEXT_SELF_REVERTS_FLAG")
    private final List<Boolean> contextSelfRevertsFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_WILL_REVERT_FLAG")
    private final List<Boolean> contextWillRevertFlag = new ArrayList<>();

    @JsonProperty("COUNTER_NSR")
    private final List<BigInteger> counterNsr = new ArrayList<>();

    @JsonProperty("COUNTER_TLI")
    private final List<Boolean> counterTli = new ArrayList<>();

    @JsonProperty(
        "DEP_NUM_NEW_xor_CALL_STACK_DEPTH_xor_MMU___REF_SIZE_xor_STACK_ITEM_HEIGHT_3_xor_INIT_GAS")
    private final List<BigInteger>
        depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas = new ArrayList<>();

    @JsonProperty(
        "DEP_NUM_xor_CALL_DATA_SIZE_xor_MMU___REF_OFFSET_xor_STACK_ITEM_HEIGHT_2_xor_INIT_CODE_SIZE")
    private final List<BigInteger>
        depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize = new ArrayList<>();

    @JsonProperty(
        "DEP_STATUS_NEW_xor_EXP___FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_BIN_FLAG_xor_VAL_CURR_IS_ZERO_xor_STATUS_CODE")
    private final List<Boolean>
        depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode =
            new ArrayList<>();

    @JsonProperty(
        "DEP_STATUS_xor_CCSR_FLAG_xor_CALL_ABORT_xor_ADD_FLAG_xor_VAL_CURR_IS_ORIG_xor_IS_EIP1559")
    private final List<Boolean>
        depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559 = new ArrayList<>();

    @JsonProperty(
        "DEPLOYMENT_NUMBER_INFTY_xor_CALL_DATA_OFFSET_xor_MMU___PARAM_2_xor_STACK_ITEM_HEIGHT_1_xor_VAL_ORIG_LO_xor_INITIAL_BALANCE")
    private final List<BigInteger>
        deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance =
            new ArrayList<>();

    @JsonProperty(
        "DEPLOYMENT_STATUS_INFTY_xor_UPDATE_xor_ABORT_FLAG_xor_BLAKE2f_xor_ACC_FLAG_xor_VAL_CURR_CHANGES_xor_IS_DEPLOYMENT")
    private final List<Boolean>
        deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment =
            new ArrayList<>();

    @JsonProperty("DOM_STAMP")
    private final List<BigInteger> domStamp = new ArrayList<>();

    @JsonProperty("EXCEPTION_AHOY_FLAG")
    private final List<Boolean> exceptionAhoyFlag = new ArrayList<>();

    @JsonProperty(
        "EXISTS_NEW_xor_MMU___FLAG_xor_CALL_PRC_FAILURE_CALLER_WILL_REVERT_xor_CALL_FLAG_xor_VAL_NEXT_IS_ORIG")
    private final List<Boolean>
        existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig =
            new ArrayList<>();

    @JsonProperty(
        "EXISTS_xor_FCOND_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BTC_FLAG_xor_VAL_NEXT_IS_CURR_xor_TXN_REQUIRES_EVM_EXECUTION")
    private final List<Boolean>
        existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution =
            new ArrayList<>();

    @JsonProperty("GAS_ACTUAL")
    private final List<BigInteger> gasActual = new ArrayList<>();

    @JsonProperty("GAS_COST")
    private final List<BigInteger> gasCost = new ArrayList<>();

    @JsonProperty("GAS_EXPECTED")
    private final List<BigInteger> gasExpected = new ArrayList<>();

    @JsonProperty("GAS_NEXT")
    private final List<BigInteger> gasNext = new ArrayList<>();

    @JsonProperty("GAS_REFUND")
    private final List<BigInteger> gasRefund = new ArrayList<>();

    @JsonProperty("GAS_REFUND_NEW")
    private final List<BigInteger> gasRefundNew = new ArrayList<>();

    @JsonProperty(
        "HAS_CODE_NEW_xor_MXP___DEPLOYS_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_COPY_FLAG_xor_VAL_ORIG_IS_ZERO")
    private final List<Boolean>
        hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero =
            new ArrayList<>();

    @JsonProperty(
        "HAS_CODE_xor_MMU___INFO_xor_CALL_PRC_FAILURE_CALLER_WONT_REVERT_xor_CON_FLAG_xor_VAL_NEXT_IS_ZERO")
    private final List<Boolean>
        hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero =
            new ArrayList<>();

    @JsonProperty("HASH_INFO_STAMP")
    private final List<BigInteger> hashInfoStamp = new ArrayList<>();

    @JsonProperty("HUB_STAMP")
    private final List<BigInteger> hubStamp = new ArrayList<>();

    @JsonProperty("HUB_STAMP_TRANSACTION_END")
    private final List<BigInteger> hubStampTransactionEnd = new ArrayList<>();

    @JsonProperty(
        "IS_BLAKE2f_xor_MXP___FLAG_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_CREATE_FLAG_xor_WARM")
    private final List<Boolean>
        isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm =
            new ArrayList<>();

    @JsonProperty(
        "IS_ECADD_xor_MXP___MXPX_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_DECODED_FLAG_1_xor_WARM_NEW")
    private final List<Boolean>
        isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew =
            new ArrayList<>();

    @JsonProperty(
        "IS_ECMUL_xor_OOB___EVENT_1_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DECODED_FLAG_2")
    private final List<Boolean>
        isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2 = new ArrayList<>();

    @JsonProperty(
        "IS_ECPAIRING_xor_OOB___EVENT_2_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DECODED_FLAG_3")
    private final List<Boolean>
        isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3 = new ArrayList<>();

    @JsonProperty(
        "IS_ECRECOVER_xor_OOB___FLAG_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DECODED_FLAG_4")
    private final List<Boolean>
        isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4 = new ArrayList<>();

    @JsonProperty("IS_IDENTITY_xor_PRECINFO___FLAG_xor_CODEDEPOSIT_xor_DUP_FLAG")
    private final List<Boolean> isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag =
        new ArrayList<>();

    @JsonProperty("IS_MODEXP_xor_STP___EXISTS_xor_CODEDEPOSIT_INVALID_CODE_PREFIX_xor_EXT_FLAG")
    private final List<Boolean> isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag =
        new ArrayList<>();

    @JsonProperty("IS_PRECOMPILE_xor_STP___FLAG_xor_CODEDEPOSIT_VALID_CODE_PREFIX_xor_HALT_FLAG")
    private final List<Boolean> isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag =
        new ArrayList<>();

    @JsonProperty("IS_RIPEMDsub160_xor_STP___OOGX_xor_ECADD_xor_HASH_INFO_FLAG")
    private final List<Boolean> isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag = new ArrayList<>();

    @JsonProperty("IS_SHA2sub256_xor_STP___WARM_xor_ECMUL_xor_INVALID_FLAG")
    private final List<Boolean> isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag = new ArrayList<>();

    @JsonProperty("MMU_STAMP")
    private final List<BigInteger> mmuStamp = new ArrayList<>();

    @JsonProperty("MXP___SIZE_1_HI_xor_STACK_ITEM_VALUE_LO_2")
    private final List<BigInteger> mxpSize1HiXorStackItemValueLo2 = new ArrayList<>();

    @JsonProperty("MXP___SIZE_1_LO_xor_STACK_ITEM_VALUE_LO_3")
    private final List<BigInteger> mxpSize1LoXorStackItemValueLo3 = new ArrayList<>();

    @JsonProperty("MXP___SIZE_2_HI_xor_STACK_ITEM_VALUE_LO_4")
    private final List<BigInteger> mxpSize2HiXorStackItemValueLo4 = new ArrayList<>();

    @JsonProperty("MXP___SIZE_2_LO_xor_STATIC_GAS")
    private final List<BigInteger> mxpSize2LoXorStaticGas = new ArrayList<>();

    @JsonProperty("MXP_STAMP")
    private final List<BigInteger> mxpStamp = new ArrayList<>();

    @JsonProperty("MXP___WORDS")
    private final List<BigInteger> mxpWords = new ArrayList<>();

    @JsonProperty("NONCE_NEW_xor_CONTEXT_NUMBER_xor_MMU___SIZE_xor_STACK_ITEM_STAMP_1_xor_NONCE")
    private final List<BigInteger> nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce =
        new ArrayList<>();

    @JsonProperty(
        "NONCE_xor_CALL_VALUE_xor_MMU___RETURNER_xor_STACK_ITEM_HEIGHT_4_xor_LEFTOVER_GAS")
    private final List<BigInteger>
        nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas = new ArrayList<>();

    @JsonProperty("NUMBER_OF_NON_STACK_ROWS")
    private final List<BigInteger> numberOfNonStackRows = new ArrayList<>();

    @JsonProperty("OOB___INST")
    private final List<BigInteger> oobInst = new ArrayList<>();

    @JsonProperty("OOB___OUTGOING_DATA_1")
    private final List<BigInteger> oobOutgoingData1 = new ArrayList<>();

    @JsonProperty("OOB___OUTGOING_DATA_2")
    private final List<BigInteger> oobOutgoingData2 = new ArrayList<>();

    @JsonProperty("OOB___OUTGOING_DATA_3")
    private final List<BigInteger> oobOutgoingData3 = new ArrayList<>();

    @JsonProperty("OOB___OUTGOING_DATA_4")
    private final List<BigInteger> oobOutgoingData4 = new ArrayList<>();

    @JsonProperty("OOB___OUTGOING_DATA_5")
    private final List<BigInteger> oobOutgoingData5 = new ArrayList<>();

    @JsonProperty("OOB___OUTGOING_DATA_6")
    private final List<BigInteger> oobOutgoingData6 = new ArrayList<>();

    @JsonProperty("PEEK_AT_ACCOUNT")
    private final List<Boolean> peekAtAccount = new ArrayList<>();

    @JsonProperty("PEEK_AT_CONTEXT")
    private final List<Boolean> peekAtContext = new ArrayList<>();

    @JsonProperty("PEEK_AT_MISCELLANEOUS")
    private final List<Boolean> peekAtMiscellaneous = new ArrayList<>();

    @JsonProperty("PEEK_AT_SCENARIO")
    private final List<Boolean> peekAtScenario = new ArrayList<>();

    @JsonProperty("PEEK_AT_STACK")
    private final List<Boolean> peekAtStack = new ArrayList<>();

    @JsonProperty("PEEK_AT_STORAGE")
    private final List<Boolean> peekAtStorage = new ArrayList<>();

    @JsonProperty("PEEK_AT_TRANSACTION")
    private final List<Boolean> peekAtTransaction = new ArrayList<>();

    @JsonProperty("PRECINFO___ADDR_LO")
    private final List<BigInteger> precinfoAddrLo = new ArrayList<>();

    @JsonProperty("PRECINFO___CDS")
    private final List<BigInteger> precinfoCds = new ArrayList<>();

    @JsonProperty("PRECINFO___EXEC_COST")
    private final List<BigInteger> precinfoExecCost = new ArrayList<>();

    @JsonProperty("PRECINFO___PROVIDES_RETURN_DATA")
    private final List<BigInteger> precinfoProvidesReturnData = new ArrayList<>();

    @JsonProperty("PRECINFO___RDS")
    private final List<BigInteger> precinfoRds = new ArrayList<>();

    @JsonProperty("PRECINFO___SUCCESS")
    private final List<BigInteger> precinfoSuccess = new ArrayList<>();

    @JsonProperty("PRECINFO___TOUCHES_RAM")
    private final List<BigInteger> precinfoTouchesRam = new ArrayList<>();

    @JsonProperty("PROGRAM_COUNTER")
    private final List<BigInteger> programCounter = new ArrayList<>();

    @JsonProperty("PROGRAM_COUNTER_NEW")
    private final List<BigInteger> programCounterNew = new ArrayList<>();

    @JsonProperty("PUSHPOP_FLAG")
    private final List<Boolean> pushpopFlag = new ArrayList<>();

    @JsonProperty("RDCX")
    private final List<Boolean> rdcx = new ArrayList<>();

    @JsonProperty("RIPEMDsub160_xor_KEC_FLAG")
    private final List<Boolean> ripemDsub160XorKecFlag = new ArrayList<>();

    @JsonProperty(
        "RLPADDR___DEP_ADDR_HI_xor_IS_STATIC_xor_MMU___STACK_VAL_HI_xor_STACK_ITEM_STAMP_2_xor_TO_ADDRESS_HI")
    private final List<BigInteger>
        rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi =
            new ArrayList<>();

    @JsonProperty(
        "RLPADDR___DEP_ADDR_LO_xor_RETURNER_CONTEXT_NUMBER_xor_MMU___STACK_VAL_LO_xor_STACK_ITEM_STAMP_3_xor_TO_ADDRESS_LO")
    private final List<BigInteger>
        rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo =
            new ArrayList<>();

    @JsonProperty("RLPADDR___FLAG_xor_ECPAIRING_xor_INVPREX")
    private final List<Boolean> rlpaddrFlagXorEcpairingXorInvprex = new ArrayList<>();

    @JsonProperty(
        "RLPADDR___KEC_HI_xor_RETURNER_IS_PRECOMPILE_xor_MXP___GAS_MXP_xor_STACK_ITEM_STAMP_4_xor_VALUE")
    private final List<BigInteger>
        rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue =
            new ArrayList<>();

    @JsonProperty("RLPADDR___KEC_LO_xor_RETURN_AT_OFFSET_xor_MXP___INST_xor_STACK_ITEM_VALUE_HI_1")
    private final List<BigInteger> rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1 =
        new ArrayList<>();

    @JsonProperty(
        "RLPADDR___RECIPE_xor_RETURN_AT_SIZE_xor_MXP___OFFSET_1_HI_xor_STACK_ITEM_VALUE_HI_2")
    private final List<BigInteger> rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2 =
        new ArrayList<>();

    @JsonProperty(
        "RLPADDR___SALT_HI_xor_RETURN_DATA_OFFSET_xor_MXP___OFFSET_1_LO_xor_STACK_ITEM_VALUE_HI_3")
    private final List<BigInteger>
        rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3 = new ArrayList<>();

    @JsonProperty(
        "RLPADDR___SALT_LO_xor_RETURN_DATA_SIZE_xor_MXP___OFFSET_2_HI_xor_STACK_ITEM_VALUE_HI_4")
    private final List<BigInteger>
        rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4 = new ArrayList<>();

    @JsonProperty("SCN_FAILURE_1_xor_LOG_FLAG")
    private final List<Boolean> scnFailure1XorLogFlag = new ArrayList<>();

    @JsonProperty("SCN_FAILURE_2_xor_MACHINE_STATE_FLAG")
    private final List<Boolean> scnFailure2XorMachineStateFlag = new ArrayList<>();

    @JsonProperty("SCN_FAILURE_3_xor_MAXCSX")
    private final List<Boolean> scnFailure3XorMaxcsx = new ArrayList<>();

    @JsonProperty("SCN_FAILURE_4_xor_MOD_FLAG")
    private final List<Boolean> scnFailure4XorModFlag = new ArrayList<>();

    @JsonProperty("SCN_SUCCESS_1_xor_MUL_FLAG")
    private final List<Boolean> scnSuccess1XorMulFlag = new ArrayList<>();

    @JsonProperty("SCN_SUCCESS_2_xor_MXPX")
    private final List<Boolean> scnSuccess2XorMxpx = new ArrayList<>();

    @JsonProperty("SCN_SUCCESS_3_xor_MXP_FLAG")
    private final List<Boolean> scnSuccess3XorMxpFlag = new ArrayList<>();

    @JsonProperty("SCN_SUCCESS_4_xor_OOB_FLAG")
    private final List<Boolean> scnSuccess4XorOobFlag = new ArrayList<>();

    @JsonProperty("SELFDESTRUCT_xor_OOGX")
    private final List<Boolean> selfdestructXorOogx = new ArrayList<>();

    @JsonProperty("SHA2sub256_xor_OPCX")
    private final List<Boolean> sha2Sub256XorOpcx = new ArrayList<>();

    @JsonProperty("SHF_FLAG")
    private final List<Boolean> shfFlag = new ArrayList<>();

    @JsonProperty("SOX")
    private final List<Boolean> sox = new ArrayList<>();

    @JsonProperty("SSTOREX")
    private final List<Boolean> sstorex = new ArrayList<>();

    @JsonProperty("STACK_ITEM_POP_1")
    private final List<Boolean> stackItemPop1 = new ArrayList<>();

    @JsonProperty("STACK_ITEM_POP_2")
    private final List<Boolean> stackItemPop2 = new ArrayList<>();

    @JsonProperty("STACK_ITEM_POP_3")
    private final List<Boolean> stackItemPop3 = new ArrayList<>();

    @JsonProperty("STACK_ITEM_POP_4")
    private final List<Boolean> stackItemPop4 = new ArrayList<>();

    @JsonProperty("STACKRAM_FLAG")
    private final List<Boolean> stackramFlag = new ArrayList<>();

    @JsonProperty("STATIC_FLAG")
    private final List<Boolean> staticFlag = new ArrayList<>();

    @JsonProperty("STATICX")
    private final List<Boolean> staticx = new ArrayList<>();

    @JsonProperty("STO_FLAG")
    private final List<Boolean> stoFlag = new ArrayList<>();

    @JsonProperty("STP___GAS_HI")
    private final List<BigInteger> stpGasHi = new ArrayList<>();

    @JsonProperty("STP___GAS_LO")
    private final List<BigInteger> stpGasLo = new ArrayList<>();

    @JsonProperty("STP___GAS_OOPKT")
    private final List<BigInteger> stpGasOopkt = new ArrayList<>();

    @JsonProperty("STP___GAS_STPD")
    private final List<BigInteger> stpGasStpd = new ArrayList<>();

    @JsonProperty("STP___INST")
    private final List<BigInteger> stpInst = new ArrayList<>();

    @JsonProperty("STP___VAL_HI")
    private final List<BigInteger> stpValHi = new ArrayList<>();

    @JsonProperty("STP___VAL_LO")
    private final List<BigInteger> stpValLo = new ArrayList<>();

    @JsonProperty("SUB_STAMP")
    private final List<BigInteger> subStamp = new ArrayList<>();

    @JsonProperty("SUX")
    private final List<Boolean> sux = new ArrayList<>();

    @JsonProperty("SWAP_FLAG")
    private final List<Boolean> swapFlag = new ArrayList<>();

    @JsonProperty("TRANSACTION_REVERTS")
    private final List<Boolean> transactionReverts = new ArrayList<>();

    @JsonProperty("TRM_FLAG")
    private final List<Boolean> trmFlag = new ArrayList<>();

    @JsonProperty("TRM___FLAG_xor_ECRECOVER_xor_JUMPX")
    private final List<Boolean> trmFlagXorEcrecoverXorJumpx = new ArrayList<>();

    @JsonProperty("TRM___RAW_ADDR_HI_xor_MXP___OFFSET_2_LO_xor_STACK_ITEM_VALUE_LO_1")
    private final List<BigInteger> trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1 =
        new ArrayList<>();

    @JsonProperty("TWO_LINE_INSTRUCTION")
    private final List<Boolean> twoLineInstruction = new ArrayList<>();

    @JsonProperty("TX_EXEC")
    private final List<Boolean> txExec = new ArrayList<>();

    @JsonProperty("TX_FINL")
    private final List<Boolean> txFinl = new ArrayList<>();

    @JsonProperty("TX_INIT")
    private final List<Boolean> txInit = new ArrayList<>();

    @JsonProperty("TX_SKIP")
    private final List<Boolean> txSkip = new ArrayList<>();

    @JsonProperty("TX_WARM")
    private final List<Boolean> txWarm = new ArrayList<>();

    @JsonProperty("TXN_FLAG")
    private final List<Boolean> txnFlag = new ArrayList<>();

    @JsonProperty("WARM_NEW_xor_MODEXP_xor_JUMP_FLAG")
    private final List<Boolean> warmNewXorModexpXorJumpFlag = new ArrayList<>();

    @JsonProperty("WARM_xor_IDENTITY_xor_JUMP_DESTINATION_VETTING_REQUIRED")
    private final List<Boolean> warmXorIdentityXorJumpDestinationVettingRequired =
        new ArrayList<>();

    @JsonProperty("WCP_FLAG")
    private final List<Boolean> wcpFlag = new ArrayList<>();

    private TraceBuilder() {}

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.absoluteTransactionNumber.size();
    }

    public TraceBuilder absoluteTransactionNumber(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ABSOLUTE_TRANSACTION_NUMBER already set");
      } else {
        filled.set(0);
      }

      absoluteTransactionNumber.add(b);

      return this;
    }

    public TraceBuilder batchNumber(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("BATCH_NUMBER already set");
      } else {
        filled.set(1);
      }

      batchNumber.add(b);

      return this;
    }

    public TraceBuilder callerContextNumber(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("CALLER_CONTEXT_NUMBER already set");
      } else {
        filled.set(2);
      }

      callerContextNumber.add(b);

      return this;
    }

    public TraceBuilder codeAddressHi(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("CODE_ADDRESS_HI already set");
      } else {
        filled.set(3);
      }

      codeAddressHi.add(b);

      return this;
    }

    public TraceBuilder codeAddressLo(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("CODE_ADDRESS_LO already set");
      } else {
        filled.set(4);
      }

      codeAddressLo.add(b);

      return this;
    }

    public TraceBuilder codeDeploymentNumber(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(5);
      }

      codeDeploymentNumber.add(b);

      return this;
    }

    public TraceBuilder codeDeploymentStatus(final Boolean b) {
      if (filled.get(6)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_STATUS already set");
      } else {
        filled.set(6);
      }

      codeDeploymentStatus.add(b);

      return this;
    }

    public TraceBuilder codeFragmentIndex(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX already set");
      } else {
        filled.set(7);
      }

      codeFragmentIndex.add(b);

      return this;
    }

    public TraceBuilder contextGetsRevertedFlag(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("CONTEXT_GETS_REVERTED_FLAG already set");
      } else {
        filled.set(8);
      }

      contextGetsRevertedFlag.add(b);

      return this;
    }

    public TraceBuilder contextMayChangeFlag(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("CONTEXT_MAY_CHANGE_FLAG already set");
      } else {
        filled.set(9);
      }

      contextMayChangeFlag.add(b);

      return this;
    }

    public TraceBuilder contextNumber(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("CONTEXT_NUMBER already set");
      } else {
        filled.set(10);
      }

      contextNumber.add(b);

      return this;
    }

    public TraceBuilder contextNumberNew(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("CONTEXT_NUMBER_NEW already set");
      } else {
        filled.set(11);
      }

      contextNumberNew.add(b);

      return this;
    }

    public TraceBuilder contextRevertStamp(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("CONTEXT_REVERT_STAMP already set");
      } else {
        filled.set(12);
      }

      contextRevertStamp.add(b);

      return this;
    }

    public TraceBuilder contextSelfRevertsFlag(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("CONTEXT_SELF_REVERTS_FLAG already set");
      } else {
        filled.set(13);
      }

      contextSelfRevertsFlag.add(b);

      return this;
    }

    public TraceBuilder contextWillRevertFlag(final Boolean b) {
      if (filled.get(14)) {
        throw new IllegalStateException("CONTEXT_WILL_REVERT_FLAG already set");
      } else {
        filled.set(14);
      }

      contextWillRevertFlag.add(b);

      return this;
    }

    public TraceBuilder counterNsr(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("COUNTER_NSR already set");
      } else {
        filled.set(15);
      }

      counterNsr.add(b);

      return this;
    }

    public TraceBuilder counterTli(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("COUNTER_TLI already set");
      } else {
        filled.set(16);
      }

      counterTli.add(b);

      return this;
    }

    public TraceBuilder domStamp(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("DOM_STAMP already set");
      } else {
        filled.set(17);
      }

      domStamp.add(b);

      return this;
    }

    public TraceBuilder exceptionAhoyFlag(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("EXCEPTION_AHOY_FLAG already set");
      } else {
        filled.set(18);
      }

      exceptionAhoyFlag.add(b);

      return this;
    }

    public TraceBuilder gasActual(final BigInteger b) {
      if (filled.get(19)) {
        throw new IllegalStateException("GAS_ACTUAL already set");
      } else {
        filled.set(19);
      }

      gasActual.add(b);

      return this;
    }

    public TraceBuilder gasCost(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("GAS_COST already set");
      } else {
        filled.set(20);
      }

      gasCost.add(b);

      return this;
    }

    public TraceBuilder gasExpected(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("GAS_EXPECTED already set");
      } else {
        filled.set(21);
      }

      gasExpected.add(b);

      return this;
    }

    public TraceBuilder gasNext(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("GAS_NEXT already set");
      } else {
        filled.set(22);
      }

      gasNext.add(b);

      return this;
    }

    public TraceBuilder gasRefund(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("GAS_REFUND already set");
      } else {
        filled.set(23);
      }

      gasRefund.add(b);

      return this;
    }

    public TraceBuilder gasRefundNew(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("GAS_REFUND_NEW already set");
      } else {
        filled.set(24);
      }

      gasRefundNew.add(b);

      return this;
    }

    public TraceBuilder hashInfoStamp(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("HASH_INFO_STAMP already set");
      } else {
        filled.set(25);
      }

      hashInfoStamp.add(b);

      return this;
    }

    public TraceBuilder hubStamp(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("HUB_STAMP already set");
      } else {
        filled.set(26);
      }

      hubStamp.add(b);

      return this;
    }

    public TraceBuilder hubStampTransactionEnd(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("HUB_STAMP_TRANSACTION_END already set");
      } else {
        filled.set(27);
      }

      hubStampTransactionEnd.add(b);

      return this;
    }

    public TraceBuilder mmuStamp(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("MMU_STAMP already set");
      } else {
        filled.set(28);
      }

      mmuStamp.add(b);

      return this;
    }

    public TraceBuilder mxpStamp(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("MXP_STAMP already set");
      } else {
        filled.set(29);
      }

      mxpStamp.add(b);

      return this;
    }

    public TraceBuilder numberOfNonStackRows(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("NUMBER_OF_NON_STACK_ROWS already set");
      } else {
        filled.set(30);
      }

      numberOfNonStackRows.add(b);

      return this;
    }

    public TraceBuilder pAccountAddrHi(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("ADDR_HI already set");
      } else {
        filled.set(98);
      }

      addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(b);

      return this;
    }

    public TraceBuilder pAccountAddrLo(final BigInteger b) {
      if (filled.get(99)) {
        throw new IllegalStateException("ADDR_LO already set");
      } else {
        filled.set(99);
      }

      addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(b);

      return this;
    }

    public TraceBuilder pAccountBalance(final BigInteger b) {
      if (filled.get(100)) {
        throw new IllegalStateException("BALANCE already set");
      } else {
        filled.set(100);
      }

      balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
          .add(b);

      return this;
    }

    public TraceBuilder pAccountBalanceNew(final BigInteger b) {
      if (filled.get(101)) {
        throw new IllegalStateException("BALANCE_NEW already set");
      } else {
        filled.set(101);
      }

      balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pAccountCodeHashHi(final BigInteger b) {
      if (filled.get(102)) {
        throw new IllegalStateException("CODE_HASH_HI already set");
      } else {
        filled.set(102);
      }

      codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(b);

      return this;
    }

    public TraceBuilder pAccountCodeHashHiNew(final BigInteger b) {
      if (filled.get(103)) {
        throw new IllegalStateException("CODE_HASH_HI_NEW already set");
      } else {
        filled.set(103);
      }

      codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pAccountCodeHashLo(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("CODE_HASH_LO already set");
      } else {
        filled.set(104);
      }

      codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit.add(
          b);

      return this;
    }

    public TraceBuilder pAccountCodeHashLoNew(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("CODE_HASH_LO_NEW already set");
      } else {
        filled.set(105);
      }

      codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(b);

      return this;
    }

    public TraceBuilder pAccountCodeSize(final BigInteger b) {
      if (filled.get(106)) {
        throw new IllegalStateException("CODE_SIZE already set");
      } else {
        filled.set(106);
      }

      codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(b);

      return this;
    }

    public TraceBuilder pAccountCodeSizeNew(final BigInteger b) {
      if (filled.get(107)) {
        throw new IllegalStateException("CODE_SIZE_NEW already set");
      } else {
        filled.set(107);
      }

      codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
          .add(b);

      return this;
    }

    public TraceBuilder pAccountDepNum(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("DEP_NUM already set");
      } else {
        filled.set(109);
      }

      depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize.add(b);

      return this;
    }

    public TraceBuilder pAccountDepNumNew(final BigInteger b) {
      if (filled.get(110)) {
        throw new IllegalStateException("DEP_NUM_NEW already set");
      } else {
        filled.set(110);
      }

      depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas.add(b);

      return this;
    }

    public TraceBuilder pAccountDepStatus(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("DEP_STATUS already set");
      } else {
        filled.set(49);
      }

      depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(b);

      return this;
    }

    public TraceBuilder pAccountDepStatusNew(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("DEP_STATUS_NEW already set");
      } else {
        filled.set(50);
      }

      depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
          .add(b);

      return this;
    }

    public TraceBuilder pAccountDeploymentNumberInfty(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("DEPLOYMENT_NUMBER_INFTY already set");
      } else {
        filled.set(108);
      }

      deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
          .add(b);

      return this;
    }

    public TraceBuilder pAccountDeploymentStatusInfty(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("DEPLOYMENT_STATUS_INFTY already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pAccountExists(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("EXISTS already set");
      } else {
        filled.set(51);
      }

      existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
          .add(b);

      return this;
    }

    public TraceBuilder pAccountExistsNew(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("EXISTS_NEW already set");
      } else {
        filled.set(52);
      }

      existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig.add(b);

      return this;
    }

    public TraceBuilder pAccountHasCode(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("HAS_CODE already set");
      } else {
        filled.set(53);
      }

      hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero.add(b);

      return this;
    }

    public TraceBuilder pAccountHasCodeNew(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("HAS_CODE_NEW already set");
      } else {
        filled.set(54);
      }

      hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero.add(b);

      return this;
    }

    public TraceBuilder pAccountIsBlake2F(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("IS_BLAKE2f already set");
      } else {
        filled.set(55);
      }

      isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm.add(b);

      return this;
    }

    public TraceBuilder pAccountIsEcadd(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("IS_ECADD already set");
      } else {
        filled.set(56);
      }

      isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew.add(b);

      return this;
    }

    public TraceBuilder pAccountIsEcmul(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("IS_ECMUL already set");
      } else {
        filled.set(57);
      }

      isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2.add(b);

      return this;
    }

    public TraceBuilder pAccountIsEcpairing(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("IS_ECPAIRING already set");
      } else {
        filled.set(58);
      }

      isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3.add(b);

      return this;
    }

    public TraceBuilder pAccountIsEcrecover(final Boolean b) {
      if (filled.get(59)) {
        throw new IllegalStateException("IS_ECRECOVER already set");
      } else {
        filled.set(59);
      }

      isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4.add(b);

      return this;
    }

    public TraceBuilder pAccountIsIdentity(final Boolean b) {
      if (filled.get(60)) {
        throw new IllegalStateException("IS_IDENTITY already set");
      } else {
        filled.set(60);
      }

      isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag.add(b);

      return this;
    }

    public TraceBuilder pAccountIsModexp(final Boolean b) {
      if (filled.get(61)) {
        throw new IllegalStateException("IS_MODEXP already set");
      } else {
        filled.set(61);
      }

      isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag.add(b);

      return this;
    }

    public TraceBuilder pAccountIsPrecompile(final Boolean b) {
      if (filled.get(62)) {
        throw new IllegalStateException("IS_PRECOMPILE already set");
      } else {
        filled.set(62);
      }

      isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag.add(b);

      return this;
    }

    public TraceBuilder pAccountIsRipemd160(final Boolean b) {
      if (filled.get(63)) {
        throw new IllegalStateException("IS_RIPEMD-160 already set");
      } else {
        filled.set(63);
      }

      isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag.add(b);

      return this;
    }

    public TraceBuilder pAccountIsSha2256(final Boolean b) {
      if (filled.get(64)) {
        throw new IllegalStateException("IS_SHA2-256 already set");
      } else {
        filled.set(64);
      }

      isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag.add(b);

      return this;
    }

    public TraceBuilder pAccountNonce(final BigInteger b) {
      if (filled.get(111)) {
        throw new IllegalStateException("NONCE already set");
      } else {
        filled.set(111);
      }

      nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas.add(b);

      return this;
    }

    public TraceBuilder pAccountNonceNew(final BigInteger b) {
      if (filled.get(112)) {
        throw new IllegalStateException("NONCE_NEW already set");
      } else {
        filled.set(112);
      }

      nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrDepAddrHi(final BigInteger b) {
      if (filled.get(113)) {
        throw new IllegalStateException("RLPADDR___DEP_ADDR_HI already set");
      } else {
        filled.set(113);
      }

      rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrDepAddrLo(final BigInteger b) {
      if (filled.get(114)) {
        throw new IllegalStateException("RLPADDR___DEP_ADDR_LO already set");
      } else {
        filled.set(114);
      }

      rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo.add(
          b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrFlag(final Boolean b) {
      if (filled.get(65)) {
        throw new IllegalStateException("RLPADDR___FLAG already set");
      } else {
        filled.set(65);
      }

      rlpaddrFlagXorEcpairingXorInvprex.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrKecHi(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("RLPADDR___KEC_HI already set");
      } else {
        filled.set(115);
      }

      rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrKecLo(final BigInteger b) {
      if (filled.get(116)) {
        throw new IllegalStateException("RLPADDR___KEC_LO already set");
      } else {
        filled.set(116);
      }

      rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrRecipe(final BigInteger b) {
      if (filled.get(117)) {
        throw new IllegalStateException("RLPADDR___RECIPE already set");
      } else {
        filled.set(117);
      }

      rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrSaltHi(final BigInteger b) {
      if (filled.get(118)) {
        throw new IllegalStateException("RLPADDR___SALT_HI already set");
      } else {
        filled.set(118);
      }

      rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3.add(b);

      return this;
    }

    public TraceBuilder pAccountRlpaddrSaltLo(final BigInteger b) {
      if (filled.get(119)) {
        throw new IllegalStateException("RLPADDR___SALT_LO already set");
      } else {
        filled.set(119);
      }

      rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4.add(b);

      return this;
    }

    public TraceBuilder pAccountTrmFlag(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("TRM___FLAG already set");
      } else {
        filled.set(66);
      }

      trmFlagXorEcrecoverXorJumpx.add(b);

      return this;
    }

    public TraceBuilder pAccountTrmRawAddrHi(final BigInteger b) {
      if (filled.get(120)) {
        throw new IllegalStateException("TRM___RAW_ADDR_HI already set");
      } else {
        filled.set(120);
      }

      trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1.add(b);

      return this;
    }

    public TraceBuilder pAccountWarm(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("WARM already set");
      } else {
        filled.set(67);
      }

      warmXorIdentityXorJumpDestinationVettingRequired.add(b);

      return this;
    }

    public TraceBuilder pAccountWarmNew(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("WARM_NEW already set");
      } else {
        filled.set(68);
      }

      warmNewXorModexpXorJumpFlag.add(b);

      return this;
    }

    public TraceBuilder pContextAccountAddressHi(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("ACCOUNT_ADDRESS_HI already set");
      } else {
        filled.set(98);
      }

      addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(b);

      return this;
    }

    public TraceBuilder pContextAccountAddressLo(final BigInteger b) {
      if (filled.get(99)) {
        throw new IllegalStateException("ACCOUNT_ADDRESS_LO already set");
      } else {
        filled.set(99);
      }

      addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(b);

      return this;
    }

    public TraceBuilder pContextAccountDeploymentNumber(final BigInteger b) {
      if (filled.get(100)) {
        throw new IllegalStateException("ACCOUNT_DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(100);
      }

      balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
          .add(b);

      return this;
    }

    public TraceBuilder pContextByteCodeAddressHi(final BigInteger b) {
      if (filled.get(101)) {
        throw new IllegalStateException("BYTE_CODE_ADDRESS_HI already set");
      } else {
        filled.set(101);
      }

      balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pContextByteCodeAddressLo(final BigInteger b) {
      if (filled.get(102)) {
        throw new IllegalStateException("BYTE_CODE_ADDRESS_LO already set");
      } else {
        filled.set(102);
      }

      codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(b);

      return this;
    }

    public TraceBuilder pContextByteCodeDeploymentNumber(final BigInteger b) {
      if (filled.get(103)) {
        throw new IllegalStateException("BYTE_CODE_DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(103);
      }

      codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pContextByteCodeDeploymentStatus(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("BYTE_CODE_DEPLOYMENT_STATUS already set");
      } else {
        filled.set(104);
      }

      codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit.add(
          b);

      return this;
    }

    public TraceBuilder pContextCallDataOffset(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("CALL_DATA_OFFSET already set");
      } else {
        filled.set(108);
      }

      deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
          .add(b);

      return this;
    }

    public TraceBuilder pContextCallDataSize(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("CALL_DATA_SIZE already set");
      } else {
        filled.set(109);
      }

      depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize.add(b);

      return this;
    }

    public TraceBuilder pContextCallStackDepth(final BigInteger b) {
      if (filled.get(110)) {
        throw new IllegalStateException("CALL_STACK_DEPTH already set");
      } else {
        filled.set(110);
      }

      depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas.add(b);

      return this;
    }

    public TraceBuilder pContextCallValue(final BigInteger b) {
      if (filled.get(111)) {
        throw new IllegalStateException("CALL_VALUE already set");
      } else {
        filled.set(111);
      }

      nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas.add(b);

      return this;
    }

    public TraceBuilder pContextCallerAddressHi(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("CALLER_ADDRESS_HI already set");
      } else {
        filled.set(105);
      }

      codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(b);

      return this;
    }

    public TraceBuilder pContextCallerAddressLo(final BigInteger b) {
      if (filled.get(106)) {
        throw new IllegalStateException("CALLER_ADDRESS_LO already set");
      } else {
        filled.set(106);
      }

      codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(b);

      return this;
    }

    public TraceBuilder pContextCallerContextNumber(final BigInteger b) {
      if (filled.get(107)) {
        throw new IllegalStateException("CALLER_CONTEXT_NUMBER already set");
      } else {
        filled.set(107);
      }

      codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
          .add(b);

      return this;
    }

    public TraceBuilder pContextContextNumber(final BigInteger b) {
      if (filled.get(112)) {
        throw new IllegalStateException("CONTEXT_NUMBER already set");
      } else {
        filled.set(112);
      }

      nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce.add(b);

      return this;
    }

    public TraceBuilder pContextIsStatic(final BigInteger b) {
      if (filled.get(113)) {
        throw new IllegalStateException("IS_STATIC already set");
      } else {
        filled.set(113);
      }

      rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi.add(b);

      return this;
    }

    public TraceBuilder pContextReturnAtOffset(final BigInteger b) {
      if (filled.get(116)) {
        throw new IllegalStateException("RETURN_AT_OFFSET already set");
      } else {
        filled.set(116);
      }

      rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1.add(b);

      return this;
    }

    public TraceBuilder pContextReturnAtSize(final BigInteger b) {
      if (filled.get(117)) {
        throw new IllegalStateException("RETURN_AT_SIZE already set");
      } else {
        filled.set(117);
      }

      rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2.add(b);

      return this;
    }

    public TraceBuilder pContextReturnDataOffset(final BigInteger b) {
      if (filled.get(118)) {
        throw new IllegalStateException("RETURN_DATA_OFFSET already set");
      } else {
        filled.set(118);
      }

      rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3.add(b);

      return this;
    }

    public TraceBuilder pContextReturnDataSize(final BigInteger b) {
      if (filled.get(119)) {
        throw new IllegalStateException("RETURN_DATA_SIZE already set");
      } else {
        filled.set(119);
      }

      rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4.add(b);

      return this;
    }

    public TraceBuilder pContextReturnerContextNumber(final BigInteger b) {
      if (filled.get(114)) {
        throw new IllegalStateException("RETURNER_CONTEXT_NUMBER already set");
      } else {
        filled.set(114);
      }

      rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo.add(
          b);

      return this;
    }

    public TraceBuilder pContextReturnerIsPrecompile(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("RETURNER_IS_PRECOMPILE already set");
      } else {
        filled.set(115);
      }

      rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue.add(b);

      return this;
    }

    public TraceBuilder pContextUpdate(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("UPDATE already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousAbortFlag(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("ABORT_FLAG already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousCcrsStamp(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("CCRS_STAMP already set");
      } else {
        filled.set(98);
      }

      addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousCcsrFlag(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("CCSR_FLAG already set");
      } else {
        filled.set(49);
      }

      depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousExpDyncost(final BigInteger b) {
      if (filled.get(99)) {
        throw new IllegalStateException("EXP___DYNCOST already set");
      } else {
        filled.set(99);
      }

      addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousExpExponentHi(final BigInteger b) {
      if (filled.get(100)) {
        throw new IllegalStateException("EXP___EXPONENT_HI already set");
      } else {
        filled.set(100);
      }

      balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousExpExponentLo(final BigInteger b) {
      if (filled.get(101)) {
        throw new IllegalStateException("EXP___EXPONENT_LO already set");
      } else {
        filled.set(101);
      }

      balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousExpFlag(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("EXP___FLAG already set");
      } else {
        filled.set(50);
      }

      depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousFcondFlag(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("FCOND_FLAG already set");
      } else {
        filled.set(51);
      }

      existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuExoSum(final BigInteger b) {
      if (filled.get(102)) {
        throw new IllegalStateException("MMU___EXO_SUM already set");
      } else {
        filled.set(102);
      }

      codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuFlag(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("MMU___FLAG already set");
      } else {
        filled.set(52);
      }

      existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuInfo(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("MMU___INFO already set");
      } else {
        filled.set(53);
      }

      hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuInst(final BigInteger b) {
      if (filled.get(103)) {
        throw new IllegalStateException("MMU___INST already set");
      } else {
        filled.set(103);
      }

      codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuOffset1Lo(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("MMU___OFFSET_1_LO already set");
      } else {
        filled.set(104);
      }

      codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit.add(
          b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuOffset2Hi(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("MMU___OFFSET_2_HI already set");
      } else {
        filled.set(105);
      }

      codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuOffset2Lo(final BigInteger b) {
      if (filled.get(106)) {
        throw new IllegalStateException("MMU___OFFSET_2_LO already set");
      } else {
        filled.set(106);
      }

      codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuParam1(final BigInteger b) {
      if (filled.get(107)) {
        throw new IllegalStateException("MMU___PARAM_1 already set");
      } else {
        filled.set(107);
      }

      codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuParam2(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("MMU___PARAM_2 already set");
      } else {
        filled.set(108);
      }

      deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
          .add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuRefOffset(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("MMU___REF_OFFSET already set");
      } else {
        filled.set(109);
      }

      depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuRefSize(final BigInteger b) {
      if (filled.get(110)) {
        throw new IllegalStateException("MMU___REF_SIZE already set");
      } else {
        filled.set(110);
      }

      depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuReturner(final BigInteger b) {
      if (filled.get(111)) {
        throw new IllegalStateException("MMU___RETURNER already set");
      } else {
        filled.set(111);
      }

      nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuSize(final BigInteger b) {
      if (filled.get(112)) {
        throw new IllegalStateException("MMU___SIZE already set");
      } else {
        filled.set(112);
      }

      nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuStackValHi(final BigInteger b) {
      if (filled.get(113)) {
        throw new IllegalStateException("MMU___STACK_VAL_HI already set");
      } else {
        filled.set(113);
      }

      rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMmuStackValLo(final BigInteger b) {
      if (filled.get(114)) {
        throw new IllegalStateException("MMU___STACK_VAL_LO already set");
      } else {
        filled.set(114);
      }

      rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo.add(
          b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpDeploys(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("MXP___DEPLOYS already set");
      } else {
        filled.set(54);
      }

      hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpFlag(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("MXP___FLAG already set");
      } else {
        filled.set(55);
      }

      isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpGasMxp(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("MXP___GAS_MXP already set");
      } else {
        filled.set(115);
      }

      rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpInst(final BigInteger b) {
      if (filled.get(116)) {
        throw new IllegalStateException("MXP___INST already set");
      } else {
        filled.set(116);
      }

      rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpMxpx(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("MXP___MXPX already set");
      } else {
        filled.set(56);
      }

      isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpOffset1Hi(final BigInteger b) {
      if (filled.get(117)) {
        throw new IllegalStateException("MXP___OFFSET_1_HI already set");
      } else {
        filled.set(117);
      }

      rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpOffset1Lo(final BigInteger b) {
      if (filled.get(118)) {
        throw new IllegalStateException("MXP___OFFSET_1_LO already set");
      } else {
        filled.set(118);
      }

      rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpOffset2Hi(final BigInteger b) {
      if (filled.get(119)) {
        throw new IllegalStateException("MXP___OFFSET_2_HI already set");
      } else {
        filled.set(119);
      }

      rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpOffset2Lo(final BigInteger b) {
      if (filled.get(120)) {
        throw new IllegalStateException("MXP___OFFSET_2_LO already set");
      } else {
        filled.set(120);
      }

      trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpSize1Hi(final BigInteger b) {
      if (filled.get(121)) {
        throw new IllegalStateException("MXP___SIZE_1_HI already set");
      } else {
        filled.set(121);
      }

      mxpSize1HiXorStackItemValueLo2.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpSize1Lo(final BigInteger b) {
      if (filled.get(122)) {
        throw new IllegalStateException("MXP___SIZE_1_LO already set");
      } else {
        filled.set(122);
      }

      mxpSize1LoXorStackItemValueLo3.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpSize2Hi(final BigInteger b) {
      if (filled.get(123)) {
        throw new IllegalStateException("MXP___SIZE_2_HI already set");
      } else {
        filled.set(123);
      }

      mxpSize2HiXorStackItemValueLo4.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpSize2Lo(final BigInteger b) {
      if (filled.get(124)) {
        throw new IllegalStateException("MXP___SIZE_2_LO already set");
      } else {
        filled.set(124);
      }

      mxpSize2LoXorStaticGas.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousMxpWords(final BigInteger b) {
      if (filled.get(125)) {
        throw new IllegalStateException("MXP___WORDS already set");
      } else {
        filled.set(125);
      }

      mxpWords.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobEvent1(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("OOB___EVENT_1 already set");
      } else {
        filled.set(57);
      }

      isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobEvent2(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("OOB___EVENT_2 already set");
      } else {
        filled.set(58);
      }

      isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobFlag(final Boolean b) {
      if (filled.get(59)) {
        throw new IllegalStateException("OOB___FLAG already set");
      } else {
        filled.set(59);
      }

      isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobInst(final BigInteger b) {
      if (filled.get(126)) {
        throw new IllegalStateException("OOB___INST already set");
      } else {
        filled.set(126);
      }

      oobInst.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobOutgoingData1(final BigInteger b) {
      if (filled.get(127)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_1 already set");
      } else {
        filled.set(127);
      }

      oobOutgoingData1.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobOutgoingData2(final BigInteger b) {
      if (filled.get(128)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_2 already set");
      } else {
        filled.set(128);
      }

      oobOutgoingData2.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobOutgoingData3(final BigInteger b) {
      if (filled.get(129)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_3 already set");
      } else {
        filled.set(129);
      }

      oobOutgoingData3.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobOutgoingData4(final BigInteger b) {
      if (filled.get(130)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_4 already set");
      } else {
        filled.set(130);
      }

      oobOutgoingData4.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobOutgoingData5(final BigInteger b) {
      if (filled.get(131)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_5 already set");
      } else {
        filled.set(131);
      }

      oobOutgoingData5.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousOobOutgoingData6(final BigInteger b) {
      if (filled.get(132)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_6 already set");
      } else {
        filled.set(132);
      }

      oobOutgoingData6.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoAddrLo(final BigInteger b) {
      if (filled.get(133)) {
        throw new IllegalStateException("PRECINFO___ADDR_LO already set");
      } else {
        filled.set(133);
      }

      precinfoAddrLo.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoCds(final BigInteger b) {
      if (filled.get(134)) {
        throw new IllegalStateException("PRECINFO___CDS already set");
      } else {
        filled.set(134);
      }

      precinfoCds.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoExecCost(final BigInteger b) {
      if (filled.get(135)) {
        throw new IllegalStateException("PRECINFO___EXEC_COST already set");
      } else {
        filled.set(135);
      }

      precinfoExecCost.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoFlag(final Boolean b) {
      if (filled.get(60)) {
        throw new IllegalStateException("PRECINFO___FLAG already set");
      } else {
        filled.set(60);
      }

      isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoProvidesReturnData(final BigInteger b) {
      if (filled.get(136)) {
        throw new IllegalStateException("PRECINFO___PROVIDES_RETURN_DATA already set");
      } else {
        filled.set(136);
      }

      precinfoProvidesReturnData.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoRds(final BigInteger b) {
      if (filled.get(137)) {
        throw new IllegalStateException("PRECINFO___RDS already set");
      } else {
        filled.set(137);
      }

      precinfoRds.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoSuccess(final BigInteger b) {
      if (filled.get(138)) {
        throw new IllegalStateException("PRECINFO___SUCCESS already set");
      } else {
        filled.set(138);
      }

      precinfoSuccess.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousPrecinfoTouchesRam(final BigInteger b) {
      if (filled.get(139)) {
        throw new IllegalStateException("PRECINFO___TOUCHES_RAM already set");
      } else {
        filled.set(139);
      }

      precinfoTouchesRam.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpExists(final Boolean b) {
      if (filled.get(61)) {
        throw new IllegalStateException("STP___EXISTS already set");
      } else {
        filled.set(61);
      }

      isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpFlag(final Boolean b) {
      if (filled.get(62)) {
        throw new IllegalStateException("STP___FLAG already set");
      } else {
        filled.set(62);
      }

      isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpGasHi(final BigInteger b) {
      if (filled.get(140)) {
        throw new IllegalStateException("STP___GAS_HI already set");
      } else {
        filled.set(140);
      }

      stpGasHi.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpGasLo(final BigInteger b) {
      if (filled.get(141)) {
        throw new IllegalStateException("STP___GAS_LO already set");
      } else {
        filled.set(141);
      }

      stpGasLo.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpGasOopkt(final BigInteger b) {
      if (filled.get(142)) {
        throw new IllegalStateException("STP___GAS_OOPKT already set");
      } else {
        filled.set(142);
      }

      stpGasOopkt.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpGasStpd(final BigInteger b) {
      if (filled.get(143)) {
        throw new IllegalStateException("STP___GAS_STPD already set");
      } else {
        filled.set(143);
      }

      stpGasStpd.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpInst(final BigInteger b) {
      if (filled.get(144)) {
        throw new IllegalStateException("STP___INST already set");
      } else {
        filled.set(144);
      }

      stpInst.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpOogx(final Boolean b) {
      if (filled.get(63)) {
        throw new IllegalStateException("STP___OOGX already set");
      } else {
        filled.set(63);
      }

      isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpValHi(final BigInteger b) {
      if (filled.get(145)) {
        throw new IllegalStateException("STP___VAL_HI already set");
      } else {
        filled.set(145);
      }

      stpValHi.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpValLo(final BigInteger b) {
      if (filled.get(146)) {
        throw new IllegalStateException("STP___VAL_LO already set");
      } else {
        filled.set(146);
      }

      stpValLo.add(b);

      return this;
    }

    public TraceBuilder pMiscellaneousStpWarm(final Boolean b) {
      if (filled.get(64)) {
        throw new IllegalStateException("STP___WARM already set");
      } else {
        filled.set(64);
      }

      isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioBlake2F(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("BLAKE2f already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pScenarioCallAbort(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("CALL_ABORT already set");
      } else {
        filled.set(49);
      }

      depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallEoaSuccessCallerWillRevert(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("CALL_EOA_SUCCESS_CALLER_WILL_REVERT already set");
      } else {
        filled.set(50);
      }

      depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
          .add(b);

      return this;
    }

    public TraceBuilder pScenarioCallEoaSuccessCallerWontRevert(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("CALL_EOA_SUCCESS_CALLER_WONT_REVERT already set");
      } else {
        filled.set(51);
      }

      existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
          .add(b);

      return this;
    }

    public TraceBuilder pScenarioCallPrcFailureCallerWillRevert(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("CALL_PRC_FAILURE_CALLER_WILL_REVERT already set");
      } else {
        filled.set(52);
      }

      existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallPrcFailureCallerWontRevert(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("CALL_PRC_FAILURE_CALLER_WONT_REVERT already set");
      } else {
        filled.set(53);
      }

      hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallPrcSuccessCallerWillRevert(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("CALL_PRC_SUCCESS_CALLER_WILL_REVERT already set");
      } else {
        filled.set(54);
      }

      hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallPrcSuccessCallerWontRevert(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("CALL_PRC_SUCCESS_CALLER_WONT_REVERT already set");
      } else {
        filled.set(55);
      }

      isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallSmcFailureCallerWillRevert(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("CALL_SMC_FAILURE_CALLER_WILL_REVERT already set");
      } else {
        filled.set(56);
      }

      isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallSmcFailureCallerWontRevert(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("CALL_SMC_FAILURE_CALLER_WONT_REVERT already set");
      } else {
        filled.set(57);
      }

      isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallSmcSuccessCallerWillRevert(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("CALL_SMC_SUCCESS_CALLER_WILL_REVERT already set");
      } else {
        filled.set(58);
      }

      isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3.add(b);

      return this;
    }

    public TraceBuilder pScenarioCallSmcSuccessCallerWontRevert(final Boolean b) {
      if (filled.get(59)) {
        throw new IllegalStateException("CALL_SMC_SUCCESS_CALLER_WONT_REVERT already set");
      } else {
        filled.set(59);
      }

      isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4.add(b);

      return this;
    }

    public TraceBuilder pScenarioCodedeposit(final Boolean b) {
      if (filled.get(60)) {
        throw new IllegalStateException("CODEDEPOSIT already set");
      } else {
        filled.set(60);
      }

      isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioCodedepositInvalidCodePrefix(final Boolean b) {
      if (filled.get(61)) {
        throw new IllegalStateException("CODEDEPOSIT_INVALID_CODE_PREFIX already set");
      } else {
        filled.set(61);
      }

      isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioCodedepositValidCodePrefix(final Boolean b) {
      if (filled.get(62)) {
        throw new IllegalStateException("CODEDEPOSIT_VALID_CODE_PREFIX already set");
      } else {
        filled.set(62);
      }

      isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioEcadd(final Boolean b) {
      if (filled.get(63)) {
        throw new IllegalStateException("ECADD already set");
      } else {
        filled.set(63);
      }

      isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioEcmul(final Boolean b) {
      if (filled.get(64)) {
        throw new IllegalStateException("ECMUL already set");
      } else {
        filled.set(64);
      }

      isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioEcpairing(final Boolean b) {
      if (filled.get(65)) {
        throw new IllegalStateException("ECPAIRING already set");
      } else {
        filled.set(65);
      }

      rlpaddrFlagXorEcpairingXorInvprex.add(b);

      return this;
    }

    public TraceBuilder pScenarioEcrecover(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("ECRECOVER already set");
      } else {
        filled.set(66);
      }

      trmFlagXorEcrecoverXorJumpx.add(b);

      return this;
    }

    public TraceBuilder pScenarioIdentity(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("IDENTITY already set");
      } else {
        filled.set(67);
      }

      warmXorIdentityXorJumpDestinationVettingRequired.add(b);

      return this;
    }

    public TraceBuilder pScenarioModexp(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("MODEXP already set");
      } else {
        filled.set(68);
      }

      warmNewXorModexpXorJumpFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioRipemd160(final Boolean b) {
      if (filled.get(69)) {
        throw new IllegalStateException("RIPEMD-160 already set");
      } else {
        filled.set(69);
      }

      ripemDsub160XorKecFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnFailure1(final Boolean b) {
      if (filled.get(70)) {
        throw new IllegalStateException("SCN_FAILURE_1 already set");
      } else {
        filled.set(70);
      }

      scnFailure1XorLogFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnFailure2(final Boolean b) {
      if (filled.get(71)) {
        throw new IllegalStateException("SCN_FAILURE_2 already set");
      } else {
        filled.set(71);
      }

      scnFailure2XorMachineStateFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnFailure3(final Boolean b) {
      if (filled.get(72)) {
        throw new IllegalStateException("SCN_FAILURE_3 already set");
      } else {
        filled.set(72);
      }

      scnFailure3XorMaxcsx.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnFailure4(final Boolean b) {
      if (filled.get(73)) {
        throw new IllegalStateException("SCN_FAILURE_4 already set");
      } else {
        filled.set(73);
      }

      scnFailure4XorModFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnSuccess1(final Boolean b) {
      if (filled.get(74)) {
        throw new IllegalStateException("SCN_SUCCESS_1 already set");
      } else {
        filled.set(74);
      }

      scnSuccess1XorMulFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnSuccess2(final Boolean b) {
      if (filled.get(75)) {
        throw new IllegalStateException("SCN_SUCCESS_2 already set");
      } else {
        filled.set(75);
      }

      scnSuccess2XorMxpx.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnSuccess3(final Boolean b) {
      if (filled.get(76)) {
        throw new IllegalStateException("SCN_SUCCESS_3 already set");
      } else {
        filled.set(76);
      }

      scnSuccess3XorMxpFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioScnSuccess4(final Boolean b) {
      if (filled.get(77)) {
        throw new IllegalStateException("SCN_SUCCESS_4 already set");
      } else {
        filled.set(77);
      }

      scnSuccess4XorOobFlag.add(b);

      return this;
    }

    public TraceBuilder pScenarioSelfdestruct(final Boolean b) {
      if (filled.get(78)) {
        throw new IllegalStateException("SELFDESTRUCT already set");
      } else {
        filled.set(78);
      }

      selfdestructXorOogx.add(b);

      return this;
    }

    public TraceBuilder pScenarioSha2256(final Boolean b) {
      if (filled.get(79)) {
        throw new IllegalStateException("SHA2-256 already set");
      } else {
        filled.set(79);
      }

      sha2Sub256XorOpcx.add(b);

      return this;
    }

    public TraceBuilder pStackAccFlag(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("ACC_FLAG already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pStackAddFlag(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("ADD_FLAG already set");
      } else {
        filled.set(49);
      }

      depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(b);

      return this;
    }

    public TraceBuilder pStackBinFlag(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("BIN_FLAG already set");
      } else {
        filled.set(50);
      }

      depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
          .add(b);

      return this;
    }

    public TraceBuilder pStackBtcFlag(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("BTC_FLAG already set");
      } else {
        filled.set(51);
      }

      existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
          .add(b);

      return this;
    }

    public TraceBuilder pStackCallFlag(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("CALL_FLAG already set");
      } else {
        filled.set(52);
      }

      existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig.add(b);

      return this;
    }

    public TraceBuilder pStackConFlag(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("CON_FLAG already set");
      } else {
        filled.set(53);
      }

      hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero.add(b);

      return this;
    }

    public TraceBuilder pStackCopyFlag(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("COPY_FLAG already set");
      } else {
        filled.set(54);
      }

      hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero.add(b);

      return this;
    }

    public TraceBuilder pStackCreateFlag(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("CREATE_FLAG already set");
      } else {
        filled.set(55);
      }

      isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm.add(b);

      return this;
    }

    public TraceBuilder pStackDecodedFlag1(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("DECODED_FLAG_1 already set");
      } else {
        filled.set(56);
      }

      isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew.add(b);

      return this;
    }

    public TraceBuilder pStackDecodedFlag2(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("DECODED_FLAG_2 already set");
      } else {
        filled.set(57);
      }

      isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2.add(b);

      return this;
    }

    public TraceBuilder pStackDecodedFlag3(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("DECODED_FLAG_3 already set");
      } else {
        filled.set(58);
      }

      isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3.add(b);

      return this;
    }

    public TraceBuilder pStackDecodedFlag4(final Boolean b) {
      if (filled.get(59)) {
        throw new IllegalStateException("DECODED_FLAG_4 already set");
      } else {
        filled.set(59);
      }

      isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4.add(b);

      return this;
    }

    public TraceBuilder pStackDupFlag(final Boolean b) {
      if (filled.get(60)) {
        throw new IllegalStateException("DUP_FLAG already set");
      } else {
        filled.set(60);
      }

      isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag.add(b);

      return this;
    }

    public TraceBuilder pStackExtFlag(final Boolean b) {
      if (filled.get(61)) {
        throw new IllegalStateException("EXT_FLAG already set");
      } else {
        filled.set(61);
      }

      isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag.add(b);

      return this;
    }

    public TraceBuilder pStackHaltFlag(final Boolean b) {
      if (filled.get(62)) {
        throw new IllegalStateException("HALT_FLAG already set");
      } else {
        filled.set(62);
      }

      isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag.add(b);

      return this;
    }

    public TraceBuilder pStackHashInfoFlag(final Boolean b) {
      if (filled.get(63)) {
        throw new IllegalStateException("HASH_INFO_FLAG already set");
      } else {
        filled.set(63);
      }

      isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag.add(b);

      return this;
    }

    public TraceBuilder pStackHashInfoKecHi(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("HASH_INFO___KEC_HI already set");
      } else {
        filled.set(98);
      }

      addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(b);

      return this;
    }

    public TraceBuilder pStackHashInfoKecLo(final BigInteger b) {
      if (filled.get(99)) {
        throw new IllegalStateException("HASH_INFO___KEC_LO already set");
      } else {
        filled.set(99);
      }

      addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(b);

      return this;
    }

    public TraceBuilder pStackHashInfoSize(final BigInteger b) {
      if (filled.get(100)) {
        throw new IllegalStateException("HASH_INFO___SIZE already set");
      } else {
        filled.set(100);
      }

      balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
          .add(b);

      return this;
    }

    public TraceBuilder pStackHeight(final BigInteger b) {
      if (filled.get(101)) {
        throw new IllegalStateException("HEIGHT already set");
      } else {
        filled.set(101);
      }

      balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pStackHeightNew(final BigInteger b) {
      if (filled.get(102)) {
        throw new IllegalStateException("HEIGHT_NEW already set");
      } else {
        filled.set(102);
      }

      codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(b);

      return this;
    }

    public TraceBuilder pStackHeightOver(final BigInteger b) {
      if (filled.get(103)) {
        throw new IllegalStateException("HEIGHT_OVER already set");
      } else {
        filled.set(103);
      }

      codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pStackHeightUnder(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("HEIGHT_UNDER already set");
      } else {
        filled.set(104);
      }

      codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit.add(
          b);

      return this;
    }

    public TraceBuilder pStackInst(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(105);
      }

      codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(b);

      return this;
    }

    public TraceBuilder pStackInvalidFlag(final Boolean b) {
      if (filled.get(64)) {
        throw new IllegalStateException("INVALID_FLAG already set");
      } else {
        filled.set(64);
      }

      isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag.add(b);

      return this;
    }

    public TraceBuilder pStackInvprex(final Boolean b) {
      if (filled.get(65)) {
        throw new IllegalStateException("INVPREX already set");
      } else {
        filled.set(65);
      }

      rlpaddrFlagXorEcpairingXorInvprex.add(b);

      return this;
    }

    public TraceBuilder pStackJumpDestinationVettingRequired(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("JUMP_DESTINATION_VETTING_REQUIRED already set");
      } else {
        filled.set(67);
      }

      warmXorIdentityXorJumpDestinationVettingRequired.add(b);

      return this;
    }

    public TraceBuilder pStackJumpFlag(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("JUMP_FLAG already set");
      } else {
        filled.set(68);
      }

      warmNewXorModexpXorJumpFlag.add(b);

      return this;
    }

    public TraceBuilder pStackJumpx(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("JUMPX already set");
      } else {
        filled.set(66);
      }

      trmFlagXorEcrecoverXorJumpx.add(b);

      return this;
    }

    public TraceBuilder pStackKecFlag(final Boolean b) {
      if (filled.get(69)) {
        throw new IllegalStateException("KEC_FLAG already set");
      } else {
        filled.set(69);
      }

      ripemDsub160XorKecFlag.add(b);

      return this;
    }

    public TraceBuilder pStackLogFlag(final Boolean b) {
      if (filled.get(70)) {
        throw new IllegalStateException("LOG_FLAG already set");
      } else {
        filled.set(70);
      }

      scnFailure1XorLogFlag.add(b);

      return this;
    }

    public TraceBuilder pStackMachineStateFlag(final Boolean b) {
      if (filled.get(71)) {
        throw new IllegalStateException("MACHINE_STATE_FLAG already set");
      } else {
        filled.set(71);
      }

      scnFailure2XorMachineStateFlag.add(b);

      return this;
    }

    public TraceBuilder pStackMaxcsx(final Boolean b) {
      if (filled.get(72)) {
        throw new IllegalStateException("MAXCSX already set");
      } else {
        filled.set(72);
      }

      scnFailure3XorMaxcsx.add(b);

      return this;
    }

    public TraceBuilder pStackModFlag(final Boolean b) {
      if (filled.get(73)) {
        throw new IllegalStateException("MOD_FLAG already set");
      } else {
        filled.set(73);
      }

      scnFailure4XorModFlag.add(b);

      return this;
    }

    public TraceBuilder pStackMulFlag(final Boolean b) {
      if (filled.get(74)) {
        throw new IllegalStateException("MUL_FLAG already set");
      } else {
        filled.set(74);
      }

      scnSuccess1XorMulFlag.add(b);

      return this;
    }

    public TraceBuilder pStackMxpFlag(final Boolean b) {
      if (filled.get(76)) {
        throw new IllegalStateException("MXP_FLAG already set");
      } else {
        filled.set(76);
      }

      scnSuccess3XorMxpFlag.add(b);

      return this;
    }

    public TraceBuilder pStackMxpx(final Boolean b) {
      if (filled.get(75)) {
        throw new IllegalStateException("MXPX already set");
      } else {
        filled.set(75);
      }

      scnSuccess2XorMxpx.add(b);

      return this;
    }

    public TraceBuilder pStackOobFlag(final Boolean b) {
      if (filled.get(77)) {
        throw new IllegalStateException("OOB_FLAG already set");
      } else {
        filled.set(77);
      }

      scnSuccess4XorOobFlag.add(b);

      return this;
    }

    public TraceBuilder pStackOogx(final Boolean b) {
      if (filled.get(78)) {
        throw new IllegalStateException("OOGX already set");
      } else {
        filled.set(78);
      }

      selfdestructXorOogx.add(b);

      return this;
    }

    public TraceBuilder pStackOpcx(final Boolean b) {
      if (filled.get(79)) {
        throw new IllegalStateException("OPCX already set");
      } else {
        filled.set(79);
      }

      sha2Sub256XorOpcx.add(b);

      return this;
    }

    public TraceBuilder pStackPushValueHi(final BigInteger b) {
      if (filled.get(106)) {
        throw new IllegalStateException("PUSH_VALUE_HI already set");
      } else {
        filled.set(106);
      }

      codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(b);

      return this;
    }

    public TraceBuilder pStackPushValueLo(final BigInteger b) {
      if (filled.get(107)) {
        throw new IllegalStateException("PUSH_VALUE_LO already set");
      } else {
        filled.set(107);
      }

      codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
          .add(b);

      return this;
    }

    public TraceBuilder pStackPushpopFlag(final Boolean b) {
      if (filled.get(80)) {
        throw new IllegalStateException("PUSHPOP_FLAG already set");
      } else {
        filled.set(80);
      }

      pushpopFlag.add(b);

      return this;
    }

    public TraceBuilder pStackRdcx(final Boolean b) {
      if (filled.get(81)) {
        throw new IllegalStateException("RDCX already set");
      } else {
        filled.set(81);
      }

      rdcx.add(b);

      return this;
    }

    public TraceBuilder pStackShfFlag(final Boolean b) {
      if (filled.get(82)) {
        throw new IllegalStateException("SHF_FLAG already set");
      } else {
        filled.set(82);
      }

      shfFlag.add(b);

      return this;
    }

    public TraceBuilder pStackSox(final Boolean b) {
      if (filled.get(83)) {
        throw new IllegalStateException("SOX already set");
      } else {
        filled.set(83);
      }

      sox.add(b);

      return this;
    }

    public TraceBuilder pStackSstorex(final Boolean b) {
      if (filled.get(84)) {
        throw new IllegalStateException("SSTOREX already set");
      } else {
        filled.set(84);
      }

      sstorex.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemHeight1(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_1 already set");
      } else {
        filled.set(108);
      }

      deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
          .add(b);

      return this;
    }

    public TraceBuilder pStackStackItemHeight2(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_2 already set");
      } else {
        filled.set(109);
      }

      depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemHeight3(final BigInteger b) {
      if (filled.get(110)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_3 already set");
      } else {
        filled.set(110);
      }

      depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemHeight4(final BigInteger b) {
      if (filled.get(111)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_4 already set");
      } else {
        filled.set(111);
      }

      nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemPop1(final Boolean b) {
      if (filled.get(86)) {
        throw new IllegalStateException("STACK_ITEM_POP_1 already set");
      } else {
        filled.set(86);
      }

      stackItemPop1.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemPop2(final Boolean b) {
      if (filled.get(87)) {
        throw new IllegalStateException("STACK_ITEM_POP_2 already set");
      } else {
        filled.set(87);
      }

      stackItemPop2.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemPop3(final Boolean b) {
      if (filled.get(88)) {
        throw new IllegalStateException("STACK_ITEM_POP_3 already set");
      } else {
        filled.set(88);
      }

      stackItemPop3.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemPop4(final Boolean b) {
      if (filled.get(89)) {
        throw new IllegalStateException("STACK_ITEM_POP_4 already set");
      } else {
        filled.set(89);
      }

      stackItemPop4.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemStamp1(final BigInteger b) {
      if (filled.get(112)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_1 already set");
      } else {
        filled.set(112);
      }

      nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemStamp2(final BigInteger b) {
      if (filled.get(113)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_2 already set");
      } else {
        filled.set(113);
      }

      rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemStamp3(final BigInteger b) {
      if (filled.get(114)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_3 already set");
      } else {
        filled.set(114);
      }

      rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo.add(
          b);

      return this;
    }

    public TraceBuilder pStackStackItemStamp4(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_4 already set");
      } else {
        filled.set(115);
      }

      rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueHi1(final BigInteger b) {
      if (filled.get(116)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_1 already set");
      } else {
        filled.set(116);
      }

      rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueHi2(final BigInteger b) {
      if (filled.get(117)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_2 already set");
      } else {
        filled.set(117);
      }

      rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueHi3(final BigInteger b) {
      if (filled.get(118)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_3 already set");
      } else {
        filled.set(118);
      }

      rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueHi4(final BigInteger b) {
      if (filled.get(119)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_4 already set");
      } else {
        filled.set(119);
      }

      rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueLo1(final BigInteger b) {
      if (filled.get(120)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_1 already set");
      } else {
        filled.set(120);
      }

      trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueLo2(final BigInteger b) {
      if (filled.get(121)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_2 already set");
      } else {
        filled.set(121);
      }

      mxpSize1HiXorStackItemValueLo2.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueLo3(final BigInteger b) {
      if (filled.get(122)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_3 already set");
      } else {
        filled.set(122);
      }

      mxpSize1LoXorStackItemValueLo3.add(b);

      return this;
    }

    public TraceBuilder pStackStackItemValueLo4(final BigInteger b) {
      if (filled.get(123)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_4 already set");
      } else {
        filled.set(123);
      }

      mxpSize2HiXorStackItemValueLo4.add(b);

      return this;
    }

    public TraceBuilder pStackStackramFlag(final Boolean b) {
      if (filled.get(85)) {
        throw new IllegalStateException("STACKRAM_FLAG already set");
      } else {
        filled.set(85);
      }

      stackramFlag.add(b);

      return this;
    }

    public TraceBuilder pStackStaticFlag(final Boolean b) {
      if (filled.get(91)) {
        throw new IllegalStateException("STATIC_FLAG already set");
      } else {
        filled.set(91);
      }

      staticFlag.add(b);

      return this;
    }

    public TraceBuilder pStackStaticGas(final BigInteger b) {
      if (filled.get(124)) {
        throw new IllegalStateException("STATIC_GAS already set");
      } else {
        filled.set(124);
      }

      mxpSize2LoXorStaticGas.add(b);

      return this;
    }

    public TraceBuilder pStackStaticx(final Boolean b) {
      if (filled.get(90)) {
        throw new IllegalStateException("STATICX already set");
      } else {
        filled.set(90);
      }

      staticx.add(b);

      return this;
    }

    public TraceBuilder pStackStoFlag(final Boolean b) {
      if (filled.get(92)) {
        throw new IllegalStateException("STO_FLAG already set");
      } else {
        filled.set(92);
      }

      stoFlag.add(b);

      return this;
    }

    public TraceBuilder pStackSux(final Boolean b) {
      if (filled.get(93)) {
        throw new IllegalStateException("SUX already set");
      } else {
        filled.set(93);
      }

      sux.add(b);

      return this;
    }

    public TraceBuilder pStackSwapFlag(final Boolean b) {
      if (filled.get(94)) {
        throw new IllegalStateException("SWAP_FLAG already set");
      } else {
        filled.set(94);
      }

      swapFlag.add(b);

      return this;
    }

    public TraceBuilder pStackTrmFlag(final Boolean b) {
      if (filled.get(95)) {
        throw new IllegalStateException("TRM_FLAG already set");
      } else {
        filled.set(95);
      }

      trmFlag.add(b);

      return this;
    }

    public TraceBuilder pStackTxnFlag(final Boolean b) {
      if (filled.get(96)) {
        throw new IllegalStateException("TXN_FLAG already set");
      } else {
        filled.set(96);
      }

      txnFlag.add(b);

      return this;
    }

    public TraceBuilder pStackWcpFlag(final Boolean b) {
      if (filled.get(97)) {
        throw new IllegalStateException("WCP_FLAG already set");
      } else {
        filled.set(97);
      }

      wcpFlag.add(b);

      return this;
    }

    public TraceBuilder pStorageAddressHi(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("ADDRESS_HI already set");
      } else {
        filled.set(98);
      }

      addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(b);

      return this;
    }

    public TraceBuilder pStorageAddressLo(final BigInteger b) {
      if (filled.get(99)) {
        throw new IllegalStateException("ADDRESS_LO already set");
      } else {
        filled.set(99);
      }

      addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(b);

      return this;
    }

    public TraceBuilder pStorageDeploymentNumber(final BigInteger b) {
      if (filled.get(100)) {
        throw new IllegalStateException("DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(100);
      }

      balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
          .add(b);

      return this;
    }

    public TraceBuilder pStorageStorageKeyHi(final BigInteger b) {
      if (filled.get(101)) {
        throw new IllegalStateException("STORAGE_KEY_HI already set");
      } else {
        filled.set(101);
      }

      balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pStorageStorageKeyLo(final BigInteger b) {
      if (filled.get(102)) {
        throw new IllegalStateException("STORAGE_KEY_LO already set");
      } else {
        filled.set(102);
      }

      codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(b);

      return this;
    }

    public TraceBuilder pStorageValCurrChanges(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("VAL_CURR_CHANGES already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pStorageValCurrHi(final BigInteger b) {
      if (filled.get(103)) {
        throw new IllegalStateException("VAL_CURR_HI already set");
      } else {
        filled.set(103);
      }

      codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pStorageValCurrIsOrig(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("VAL_CURR_IS_ORIG already set");
      } else {
        filled.set(49);
      }

      depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(b);

      return this;
    }

    public TraceBuilder pStorageValCurrIsZero(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("VAL_CURR_IS_ZERO already set");
      } else {
        filled.set(50);
      }

      depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
          .add(b);

      return this;
    }

    public TraceBuilder pStorageValCurrLo(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("VAL_CURR_LO already set");
      } else {
        filled.set(104);
      }

      codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit.add(
          b);

      return this;
    }

    public TraceBuilder pStorageValNextHi(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("VAL_NEXT_HI already set");
      } else {
        filled.set(105);
      }

      codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(b);

      return this;
    }

    public TraceBuilder pStorageValNextIsCurr(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("VAL_NEXT_IS_CURR already set");
      } else {
        filled.set(51);
      }

      existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
          .add(b);

      return this;
    }

    public TraceBuilder pStorageValNextIsOrig(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("VAL_NEXT_IS_ORIG already set");
      } else {
        filled.set(52);
      }

      existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig.add(b);

      return this;
    }

    public TraceBuilder pStorageValNextIsZero(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("VAL_NEXT_IS_ZERO already set");
      } else {
        filled.set(53);
      }

      hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero.add(b);

      return this;
    }

    public TraceBuilder pStorageValNextLo(final BigInteger b) {
      if (filled.get(106)) {
        throw new IllegalStateException("VAL_NEXT_LO already set");
      } else {
        filled.set(106);
      }

      codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(b);

      return this;
    }

    public TraceBuilder pStorageValOrigHi(final BigInteger b) {
      if (filled.get(107)) {
        throw new IllegalStateException("VAL_ORIG_HI already set");
      } else {
        filled.set(107);
      }

      codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
          .add(b);

      return this;
    }

    public TraceBuilder pStorageValOrigIsZero(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("VAL_ORIG_IS_ZERO already set");
      } else {
        filled.set(54);
      }

      hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero.add(b);

      return this;
    }

    public TraceBuilder pStorageValOrigLo(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("VAL_ORIG_LO already set");
      } else {
        filled.set(108);
      }

      deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
          .add(b);

      return this;
    }

    public TraceBuilder pStorageWarm(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("WARM already set");
      } else {
        filled.set(55);
      }

      isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm.add(b);

      return this;
    }

    public TraceBuilder pStorageWarmNew(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("WARM_NEW already set");
      } else {
        filled.set(56);
      }

      isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew.add(b);

      return this;
    }

    public TraceBuilder pTransactionBasefee(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("BASEFEE already set");
      } else {
        filled.set(98);
      }

      addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(b);

      return this;
    }

    public TraceBuilder pTransactionCallDataSize(final BigInteger b) {
      if (filled.get(99)) {
        throw new IllegalStateException("CALL_DATA_SIZE already set");
      } else {
        filled.set(99);
      }

      addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(b);

      return this;
    }

    public TraceBuilder pTransactionCoinbaseAddressHi(final BigInteger b) {
      if (filled.get(100)) {
        throw new IllegalStateException("COINBASE_ADDRESS_HI already set");
      } else {
        filled.set(100);
      }

      balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionCoinbaseAddressLo(final BigInteger b) {
      if (filled.get(101)) {
        throw new IllegalStateException("COINBASE_ADDRESS_LO already set");
      } else {
        filled.set(101);
      }

      balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionFromAddressHi(final BigInteger b) {
      if (filled.get(102)) {
        throw new IllegalStateException("FROM_ADDRESS_HI already set");
      } else {
        filled.set(102);
      }

      codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(b);

      return this;
    }

    public TraceBuilder pTransactionFromAddressLo(final BigInteger b) {
      if (filled.get(103)) {
        throw new IllegalStateException("FROM_ADDRESS_LO already set");
      } else {
        filled.set(103);
      }

      codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionGasLimit(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("GAS_LIMIT already set");
      } else {
        filled.set(104);
      }

      codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit.add(
          b);

      return this;
    }

    public TraceBuilder pTransactionGasPrice(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("GAS_PRICE already set");
      } else {
        filled.set(105);
      }

      codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(b);

      return this;
    }

    public TraceBuilder pTransactionGasRefundAmount(final BigInteger b) {
      if (filled.get(106)) {
        throw new IllegalStateException("GAS_REFUND_AMOUNT already set");
      } else {
        filled.set(106);
      }

      codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(b);

      return this;
    }

    public TraceBuilder pTransactionGasRefundCounterFinal(final BigInteger b) {
      if (filled.get(107)) {
        throw new IllegalStateException("GAS_REFUND_COUNTER_FINAL already set");
      } else {
        filled.set(107);
      }

      codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionInitCodeSize(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("INIT_CODE_SIZE already set");
      } else {
        filled.set(109);
      }

      depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize.add(b);

      return this;
    }

    public TraceBuilder pTransactionInitGas(final BigInteger b) {
      if (filled.get(110)) {
        throw new IllegalStateException("INIT_GAS already set");
      } else {
        filled.set(110);
      }

      depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas.add(b);

      return this;
    }

    public TraceBuilder pTransactionInitialBalance(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("INITIAL_BALANCE already set");
      } else {
        filled.set(108);
      }

      deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionIsDeployment(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("IS_DEPLOYMENT already set");
      } else {
        filled.set(48);
      }

      deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionIsEip1559(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("IS_EIP1559 already set");
      } else {
        filled.set(49);
      }

      depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(b);

      return this;
    }

    public TraceBuilder pTransactionLeftoverGas(final BigInteger b) {
      if (filled.get(111)) {
        throw new IllegalStateException("LEFTOVER_GAS already set");
      } else {
        filled.set(111);
      }

      nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas.add(b);

      return this;
    }

    public TraceBuilder pTransactionNonce(final BigInteger b) {
      if (filled.get(112)) {
        throw new IllegalStateException("NONCE already set");
      } else {
        filled.set(112);
      }

      nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce.add(b);

      return this;
    }

    public TraceBuilder pTransactionStatusCode(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("STATUS_CODE already set");
      } else {
        filled.set(50);
      }

      depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionToAddressHi(final BigInteger b) {
      if (filled.get(113)) {
        throw new IllegalStateException("TO_ADDRESS_HI already set");
      } else {
        filled.set(113);
      }

      rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi.add(b);

      return this;
    }

    public TraceBuilder pTransactionToAddressLo(final BigInteger b) {
      if (filled.get(114)) {
        throw new IllegalStateException("TO_ADDRESS_LO already set");
      } else {
        filled.set(114);
      }

      rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo.add(
          b);

      return this;
    }

    public TraceBuilder pTransactionTxnRequiresEvmExecution(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("TXN_REQUIRES_EVM_EXECUTION already set");
      } else {
        filled.set(51);
      }

      existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
          .add(b);

      return this;
    }

    public TraceBuilder pTransactionValue(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("VALUE already set");
      } else {
        filled.set(115);
      }

      rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue.add(b);

      return this;
    }

    public TraceBuilder peekAtAccount(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("PEEK_AT_ACCOUNT already set");
      } else {
        filled.set(31);
      }

      peekAtAccount.add(b);

      return this;
    }

    public TraceBuilder peekAtContext(final Boolean b) {
      if (filled.get(32)) {
        throw new IllegalStateException("PEEK_AT_CONTEXT already set");
      } else {
        filled.set(32);
      }

      peekAtContext.add(b);

      return this;
    }

    public TraceBuilder peekAtMiscellaneous(final Boolean b) {
      if (filled.get(33)) {
        throw new IllegalStateException("PEEK_AT_MISCELLANEOUS already set");
      } else {
        filled.set(33);
      }

      peekAtMiscellaneous.add(b);

      return this;
    }

    public TraceBuilder peekAtScenario(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("PEEK_AT_SCENARIO already set");
      } else {
        filled.set(34);
      }

      peekAtScenario.add(b);

      return this;
    }

    public TraceBuilder peekAtStack(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("PEEK_AT_STACK already set");
      } else {
        filled.set(35);
      }

      peekAtStack.add(b);

      return this;
    }

    public TraceBuilder peekAtStorage(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("PEEK_AT_STORAGE already set");
      } else {
        filled.set(36);
      }

      peekAtStorage.add(b);

      return this;
    }

    public TraceBuilder peekAtTransaction(final Boolean b) {
      if (filled.get(37)) {
        throw new IllegalStateException("PEEK_AT_TRANSACTION already set");
      } else {
        filled.set(37);
      }

      peekAtTransaction.add(b);

      return this;
    }

    public TraceBuilder programCounter(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("PROGRAM_COUNTER already set");
      } else {
        filled.set(38);
      }

      programCounter.add(b);

      return this;
    }

    public TraceBuilder programCounterNew(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("PROGRAM_COUNTER_NEW already set");
      } else {
        filled.set(39);
      }

      programCounterNew.add(b);

      return this;
    }

    public TraceBuilder subStamp(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("SUB_STAMP already set");
      } else {
        filled.set(40);
      }

      subStamp.add(b);

      return this;
    }

    public TraceBuilder transactionReverts(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("TRANSACTION_REVERTS already set");
      } else {
        filled.set(41);
      }

      transactionReverts.add(b);

      return this;
    }

    public TraceBuilder twoLineInstruction(final Boolean b) {
      if (filled.get(42)) {
        throw new IllegalStateException("TWO_LINE_INSTRUCTION already set");
      } else {
        filled.set(42);
      }

      twoLineInstruction.add(b);

      return this;
    }

    public TraceBuilder txExec(final Boolean b) {
      if (filled.get(43)) {
        throw new IllegalStateException("TX_EXEC already set");
      } else {
        filled.set(43);
      }

      txExec.add(b);

      return this;
    }

    public TraceBuilder txFinl(final Boolean b) {
      if (filled.get(44)) {
        throw new IllegalStateException("TX_FINL already set");
      } else {
        filled.set(44);
      }

      txFinl.add(b);

      return this;
    }

    public TraceBuilder txInit(final Boolean b) {
      if (filled.get(45)) {
        throw new IllegalStateException("TX_INIT already set");
      } else {
        filled.set(45);
      }

      txInit.add(b);

      return this;
    }

    public TraceBuilder txSkip(final Boolean b) {
      if (filled.get(46)) {
        throw new IllegalStateException("TX_SKIP already set");
      } else {
        filled.set(46);
      }

      txSkip.add(b);

      return this;
    }

    public TraceBuilder txWarm(final Boolean b) {
      if (filled.get(47)) {
        throw new IllegalStateException("TX_WARM already set");
      } else {
        filled.set(47);
      }

      txWarm.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ABSOLUTE_TRANSACTION_NUMBER has not been filled");
      }

      if (!filled.get(98)) {
        throw new IllegalStateException(
            "ADDR_HI_xor_ACCOUNT_ADDRESS_HI_xor_CCRS_STAMP_xor_HASH_INFO___KEC_HI_xor_ADDRESS_HI_xor_BASEFEE has not been filled");
      }

      if (!filled.get(99)) {
        throw new IllegalStateException(
            "ADDR_LO_xor_ACCOUNT_ADDRESS_LO_xor_EXP___DYNCOST_xor_HASH_INFO___KEC_LO_xor_ADDRESS_LO_xor_CALL_DATA_SIZE has not been filled");
      }

      if (!filled.get(101)) {
        throw new IllegalStateException(
            "BALANCE_NEW_xor_BYTE_CODE_ADDRESS_HI_xor_EXP___EXPONENT_LO_xor_HEIGHT_xor_STORAGE_KEY_HI_xor_COINBASE_ADDRESS_LO has not been filled");
      }

      if (!filled.get(100)) {
        throw new IllegalStateException(
            "BALANCE_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_EXP___EXPONENT_HI_xor_HASH_INFO___SIZE_xor_DEPLOYMENT_NUMBER_xor_COINBASE_ADDRESS_HI has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("BATCH_NUMBER has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("CALLER_CONTEXT_NUMBER has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("CODE_ADDRESS_HI has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("CODE_ADDRESS_LO has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_NUMBER has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_STATUS has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX has not been filled");
      }

      if (!filled.get(103)) {
        throw new IllegalStateException(
            "CODE_HASH_HI_NEW_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_MMU___INST_xor_HEIGHT_OVER_xor_VAL_CURR_HI_xor_FROM_ADDRESS_LO has not been filled");
      }

      if (!filled.get(102)) {
        throw new IllegalStateException(
            "CODE_HASH_HI_xor_BYTE_CODE_ADDRESS_LO_xor_MMU___EXO_SUM_xor_HEIGHT_NEW_xor_STORAGE_KEY_LO_xor_FROM_ADDRESS_HI has not been filled");
      }

      if (!filled.get(105)) {
        throw new IllegalStateException(
            "CODE_HASH_LO_NEW_xor_CALLER_ADDRESS_HI_xor_MMU___OFFSET_2_HI_xor_INST_xor_VAL_NEXT_HI_xor_GAS_PRICE has not been filled");
      }

      if (!filled.get(104)) {
        throw new IllegalStateException(
            "CODE_HASH_LO_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_MMU___OFFSET_1_LO_xor_HEIGHT_UNDER_xor_VAL_CURR_LO_xor_GAS_LIMIT has not been filled");
      }

      if (!filled.get(107)) {
        throw new IllegalStateException(
            "CODE_SIZE_NEW_xor_CALLER_CONTEXT_NUMBER_xor_MMU___PARAM_1_xor_PUSH_VALUE_LO_xor_VAL_ORIG_HI_xor_GAS_REFUND_COUNTER_FINAL has not been filled");
      }

      if (!filled.get(106)) {
        throw new IllegalStateException(
            "CODE_SIZE_xor_CALLER_ADDRESS_LO_xor_MMU___OFFSET_2_LO_xor_PUSH_VALUE_HI_xor_VAL_NEXT_LO_xor_GAS_REFUND_AMOUNT has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("CONTEXT_GETS_REVERTED_FLAG has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("CONTEXT_MAY_CHANGE_FLAG has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("CONTEXT_NUMBER has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("CONTEXT_NUMBER_NEW has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("CONTEXT_REVERT_STAMP has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("CONTEXT_SELF_REVERTS_FLAG has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("CONTEXT_WILL_REVERT_FLAG has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("COUNTER_NSR has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("COUNTER_TLI has not been filled");
      }

      if (!filled.get(110)) {
        throw new IllegalStateException(
            "DEP_NUM_NEW_xor_CALL_STACK_DEPTH_xor_MMU___REF_SIZE_xor_STACK_ITEM_HEIGHT_3_xor_INIT_GAS has not been filled");
      }

      if (!filled.get(109)) {
        throw new IllegalStateException(
            "DEP_NUM_xor_CALL_DATA_SIZE_xor_MMU___REF_OFFSET_xor_STACK_ITEM_HEIGHT_2_xor_INIT_CODE_SIZE has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException(
            "DEP_STATUS_NEW_xor_EXP___FLAG_xor_CALL_EOA_SUCCESS_CALLER_WILL_REVERT_xor_BIN_FLAG_xor_VAL_CURR_IS_ZERO_xor_STATUS_CODE has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException(
            "DEP_STATUS_xor_CCSR_FLAG_xor_CALL_ABORT_xor_ADD_FLAG_xor_VAL_CURR_IS_ORIG_xor_IS_EIP1559 has not been filled");
      }

      if (!filled.get(108)) {
        throw new IllegalStateException(
            "DEPLOYMENT_NUMBER_INFTY_xor_CALL_DATA_OFFSET_xor_MMU___PARAM_2_xor_STACK_ITEM_HEIGHT_1_xor_VAL_ORIG_LO_xor_INITIAL_BALANCE has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException(
            "DEPLOYMENT_STATUS_INFTY_xor_UPDATE_xor_ABORT_FLAG_xor_BLAKE2f_xor_ACC_FLAG_xor_VAL_CURR_CHANGES_xor_IS_DEPLOYMENT has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("DOM_STAMP has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("EXCEPTION_AHOY_FLAG has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException(
            "EXISTS_NEW_xor_MMU___FLAG_xor_CALL_PRC_FAILURE_CALLER_WILL_REVERT_xor_CALL_FLAG_xor_VAL_NEXT_IS_ORIG has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException(
            "EXISTS_xor_FCOND_FLAG_xor_CALL_EOA_SUCCESS_CALLER_WONT_REVERT_xor_BTC_FLAG_xor_VAL_NEXT_IS_CURR_xor_TXN_REQUIRES_EVM_EXECUTION has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("GAS_ACTUAL has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("GAS_COST has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("GAS_EXPECTED has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("GAS_NEXT has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("GAS_REFUND has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("GAS_REFUND_NEW has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException(
            "HAS_CODE_NEW_xor_MXP___DEPLOYS_xor_CALL_PRC_SUCCESS_CALLER_WILL_REVERT_xor_COPY_FLAG_xor_VAL_ORIG_IS_ZERO has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException(
            "HAS_CODE_xor_MMU___INFO_xor_CALL_PRC_FAILURE_CALLER_WONT_REVERT_xor_CON_FLAG_xor_VAL_NEXT_IS_ZERO has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("HASH_INFO_STAMP has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("HUB_STAMP has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("HUB_STAMP_TRANSACTION_END has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException(
            "IS_BLAKE2f_xor_MXP___FLAG_xor_CALL_PRC_SUCCESS_CALLER_WONT_REVERT_xor_CREATE_FLAG_xor_WARM has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException(
            "IS_ECADD_xor_MXP___MXPX_xor_CALL_SMC_FAILURE_CALLER_WILL_REVERT_xor_DECODED_FLAG_1_xor_WARM_NEW has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException(
            "IS_ECMUL_xor_OOB___EVENT_1_xor_CALL_SMC_FAILURE_CALLER_WONT_REVERT_xor_DECODED_FLAG_2 has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException(
            "IS_ECPAIRING_xor_OOB___EVENT_2_xor_CALL_SMC_SUCCESS_CALLER_WILL_REVERT_xor_DECODED_FLAG_3 has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException(
            "IS_ECRECOVER_xor_OOB___FLAG_xor_CALL_SMC_SUCCESS_CALLER_WONT_REVERT_xor_DECODED_FLAG_4 has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException(
            "IS_IDENTITY_xor_PRECINFO___FLAG_xor_CODEDEPOSIT_xor_DUP_FLAG has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException(
            "IS_MODEXP_xor_STP___EXISTS_xor_CODEDEPOSIT_INVALID_CODE_PREFIX_xor_EXT_FLAG has not been filled");
      }

      if (!filled.get(62)) {
        throw new IllegalStateException(
            "IS_PRECOMPILE_xor_STP___FLAG_xor_CODEDEPOSIT_VALID_CODE_PREFIX_xor_HALT_FLAG has not been filled");
      }

      if (!filled.get(63)) {
        throw new IllegalStateException(
            "IS_RIPEMDsub160_xor_STP___OOGX_xor_ECADD_xor_HASH_INFO_FLAG has not been filled");
      }

      if (!filled.get(64)) {
        throw new IllegalStateException(
            "IS_SHA2sub256_xor_STP___WARM_xor_ECMUL_xor_INVALID_FLAG has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("MMU_STAMP has not been filled");
      }

      if (!filled.get(121)) {
        throw new IllegalStateException(
            "MXP___SIZE_1_HI_xor_STACK_ITEM_VALUE_LO_2 has not been filled");
      }

      if (!filled.get(122)) {
        throw new IllegalStateException(
            "MXP___SIZE_1_LO_xor_STACK_ITEM_VALUE_LO_3 has not been filled");
      }

      if (!filled.get(123)) {
        throw new IllegalStateException(
            "MXP___SIZE_2_HI_xor_STACK_ITEM_VALUE_LO_4 has not been filled");
      }

      if (!filled.get(124)) {
        throw new IllegalStateException("MXP___SIZE_2_LO_xor_STATIC_GAS has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("MXP_STAMP has not been filled");
      }

      if (!filled.get(125)) {
        throw new IllegalStateException("MXP___WORDS has not been filled");
      }

      if (!filled.get(112)) {
        throw new IllegalStateException(
            "NONCE_NEW_xor_CONTEXT_NUMBER_xor_MMU___SIZE_xor_STACK_ITEM_STAMP_1_xor_NONCE has not been filled");
      }

      if (!filled.get(111)) {
        throw new IllegalStateException(
            "NONCE_xor_CALL_VALUE_xor_MMU___RETURNER_xor_STACK_ITEM_HEIGHT_4_xor_LEFTOVER_GAS has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("NUMBER_OF_NON_STACK_ROWS has not been filled");
      }

      if (!filled.get(126)) {
        throw new IllegalStateException("OOB___INST has not been filled");
      }

      if (!filled.get(127)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_1 has not been filled");
      }

      if (!filled.get(128)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_2 has not been filled");
      }

      if (!filled.get(129)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_3 has not been filled");
      }

      if (!filled.get(130)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_4 has not been filled");
      }

      if (!filled.get(131)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_5 has not been filled");
      }

      if (!filled.get(132)) {
        throw new IllegalStateException("OOB___OUTGOING_DATA_6 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("PEEK_AT_ACCOUNT has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("PEEK_AT_CONTEXT has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("PEEK_AT_MISCELLANEOUS has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("PEEK_AT_SCENARIO has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("PEEK_AT_STACK has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("PEEK_AT_STORAGE has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("PEEK_AT_TRANSACTION has not been filled");
      }

      if (!filled.get(133)) {
        throw new IllegalStateException("PRECINFO___ADDR_LO has not been filled");
      }

      if (!filled.get(134)) {
        throw new IllegalStateException("PRECINFO___CDS has not been filled");
      }

      if (!filled.get(135)) {
        throw new IllegalStateException("PRECINFO___EXEC_COST has not been filled");
      }

      if (!filled.get(136)) {
        throw new IllegalStateException("PRECINFO___PROVIDES_RETURN_DATA has not been filled");
      }

      if (!filled.get(137)) {
        throw new IllegalStateException("PRECINFO___RDS has not been filled");
      }

      if (!filled.get(138)) {
        throw new IllegalStateException("PRECINFO___SUCCESS has not been filled");
      }

      if (!filled.get(139)) {
        throw new IllegalStateException("PRECINFO___TOUCHES_RAM has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("PROGRAM_COUNTER has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("PROGRAM_COUNTER_NEW has not been filled");
      }

      if (!filled.get(80)) {
        throw new IllegalStateException("PUSHPOP_FLAG has not been filled");
      }

      if (!filled.get(81)) {
        throw new IllegalStateException("RDCX has not been filled");
      }

      if (!filled.get(69)) {
        throw new IllegalStateException("RIPEMDsub160_xor_KEC_FLAG has not been filled");
      }

      if (!filled.get(113)) {
        throw new IllegalStateException(
            "RLPADDR___DEP_ADDR_HI_xor_IS_STATIC_xor_MMU___STACK_VAL_HI_xor_STACK_ITEM_STAMP_2_xor_TO_ADDRESS_HI has not been filled");
      }

      if (!filled.get(114)) {
        throw new IllegalStateException(
            "RLPADDR___DEP_ADDR_LO_xor_RETURNER_CONTEXT_NUMBER_xor_MMU___STACK_VAL_LO_xor_STACK_ITEM_STAMP_3_xor_TO_ADDRESS_LO has not been filled");
      }

      if (!filled.get(65)) {
        throw new IllegalStateException(
            "RLPADDR___FLAG_xor_ECPAIRING_xor_INVPREX has not been filled");
      }

      if (!filled.get(115)) {
        throw new IllegalStateException(
            "RLPADDR___KEC_HI_xor_RETURNER_IS_PRECOMPILE_xor_MXP___GAS_MXP_xor_STACK_ITEM_STAMP_4_xor_VALUE has not been filled");
      }

      if (!filled.get(116)) {
        throw new IllegalStateException(
            "RLPADDR___KEC_LO_xor_RETURN_AT_OFFSET_xor_MXP___INST_xor_STACK_ITEM_VALUE_HI_1 has not been filled");
      }

      if (!filled.get(117)) {
        throw new IllegalStateException(
            "RLPADDR___RECIPE_xor_RETURN_AT_SIZE_xor_MXP___OFFSET_1_HI_xor_STACK_ITEM_VALUE_HI_2 has not been filled");
      }

      if (!filled.get(118)) {
        throw new IllegalStateException(
            "RLPADDR___SALT_HI_xor_RETURN_DATA_OFFSET_xor_MXP___OFFSET_1_LO_xor_STACK_ITEM_VALUE_HI_3 has not been filled");
      }

      if (!filled.get(119)) {
        throw new IllegalStateException(
            "RLPADDR___SALT_LO_xor_RETURN_DATA_SIZE_xor_MXP___OFFSET_2_HI_xor_STACK_ITEM_VALUE_HI_4 has not been filled");
      }

      if (!filled.get(70)) {
        throw new IllegalStateException("SCN_FAILURE_1_xor_LOG_FLAG has not been filled");
      }

      if (!filled.get(71)) {
        throw new IllegalStateException("SCN_FAILURE_2_xor_MACHINE_STATE_FLAG has not been filled");
      }

      if (!filled.get(72)) {
        throw new IllegalStateException("SCN_FAILURE_3_xor_MAXCSX has not been filled");
      }

      if (!filled.get(73)) {
        throw new IllegalStateException("SCN_FAILURE_4_xor_MOD_FLAG has not been filled");
      }

      if (!filled.get(74)) {
        throw new IllegalStateException("SCN_SUCCESS_1_xor_MUL_FLAG has not been filled");
      }

      if (!filled.get(75)) {
        throw new IllegalStateException("SCN_SUCCESS_2_xor_MXPX has not been filled");
      }

      if (!filled.get(76)) {
        throw new IllegalStateException("SCN_SUCCESS_3_xor_MXP_FLAG has not been filled");
      }

      if (!filled.get(77)) {
        throw new IllegalStateException("SCN_SUCCESS_4_xor_OOB_FLAG has not been filled");
      }

      if (!filled.get(78)) {
        throw new IllegalStateException("SELFDESTRUCT_xor_OOGX has not been filled");
      }

      if (!filled.get(79)) {
        throw new IllegalStateException("SHA2sub256_xor_OPCX has not been filled");
      }

      if (!filled.get(82)) {
        throw new IllegalStateException("SHF_FLAG has not been filled");
      }

      if (!filled.get(83)) {
        throw new IllegalStateException("SOX has not been filled");
      }

      if (!filled.get(84)) {
        throw new IllegalStateException("SSTOREX has not been filled");
      }

      if (!filled.get(86)) {
        throw new IllegalStateException("STACK_ITEM_POP_1 has not been filled");
      }

      if (!filled.get(87)) {
        throw new IllegalStateException("STACK_ITEM_POP_2 has not been filled");
      }

      if (!filled.get(88)) {
        throw new IllegalStateException("STACK_ITEM_POP_3 has not been filled");
      }

      if (!filled.get(89)) {
        throw new IllegalStateException("STACK_ITEM_POP_4 has not been filled");
      }

      if (!filled.get(85)) {
        throw new IllegalStateException("STACKRAM_FLAG has not been filled");
      }

      if (!filled.get(91)) {
        throw new IllegalStateException("STATIC_FLAG has not been filled");
      }

      if (!filled.get(90)) {
        throw new IllegalStateException("STATICX has not been filled");
      }

      if (!filled.get(92)) {
        throw new IllegalStateException("STO_FLAG has not been filled");
      }

      if (!filled.get(140)) {
        throw new IllegalStateException("STP___GAS_HI has not been filled");
      }

      if (!filled.get(141)) {
        throw new IllegalStateException("STP___GAS_LO has not been filled");
      }

      if (!filled.get(142)) {
        throw new IllegalStateException("STP___GAS_OOPKT has not been filled");
      }

      if (!filled.get(143)) {
        throw new IllegalStateException("STP___GAS_STPD has not been filled");
      }

      if (!filled.get(144)) {
        throw new IllegalStateException("STP___INST has not been filled");
      }

      if (!filled.get(145)) {
        throw new IllegalStateException("STP___VAL_HI has not been filled");
      }

      if (!filled.get(146)) {
        throw new IllegalStateException("STP___VAL_LO has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("SUB_STAMP has not been filled");
      }

      if (!filled.get(93)) {
        throw new IllegalStateException("SUX has not been filled");
      }

      if (!filled.get(94)) {
        throw new IllegalStateException("SWAP_FLAG has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("TRANSACTION_REVERTS has not been filled");
      }

      if (!filled.get(95)) {
        throw new IllegalStateException("TRM_FLAG has not been filled");
      }

      if (!filled.get(66)) {
        throw new IllegalStateException("TRM___FLAG_xor_ECRECOVER_xor_JUMPX has not been filled");
      }

      if (!filled.get(120)) {
        throw new IllegalStateException(
            "TRM___RAW_ADDR_HI_xor_MXP___OFFSET_2_LO_xor_STACK_ITEM_VALUE_LO_1 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("TWO_LINE_INSTRUCTION has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("TX_EXEC has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("TX_FINL has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("TX_INIT has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("TX_SKIP has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("TX_WARM has not been filled");
      }

      if (!filled.get(96)) {
        throw new IllegalStateException("TXN_FLAG has not been filled");
      }

      if (!filled.get(68)) {
        throw new IllegalStateException("WARM_NEW_xor_MODEXP_xor_JUMP_FLAG has not been filled");
      }

      if (!filled.get(67)) {
        throw new IllegalStateException(
            "WARM_xor_IDENTITY_xor_JUMP_DESTINATION_VETTING_REQUIRED has not been filled");
      }

      if (!filled.get(97)) {
        throw new IllegalStateException("WCP_FLAG has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        absoluteTransactionNumber.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(98)) {
        addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee.add(
            BigInteger.ZERO);
        this.filled.set(98);
      }
      if (!filled.get(99)) {
        addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize.add(
            BigInteger.ZERO);
        this.filled.set(99);
      }
      if (!filled.get(101)) {
        balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo
            .add(BigInteger.ZERO);
        this.filled.set(101);
      }
      if (!filled.get(100)) {
        balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi
            .add(BigInteger.ZERO);
        this.filled.set(100);
      }
      if (!filled.get(1)) {
        batchNumber.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        callerContextNumber.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        codeAddressHi.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        codeAddressLo.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        codeDeploymentNumber.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        codeDeploymentStatus.add(false);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        codeFragmentIndex.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(103)) {
        codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo
            .add(BigInteger.ZERO);
        this.filled.set(103);
      }
      if (!filled.get(102)) {
        codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi.add(
            BigInteger.ZERO);
        this.filled.set(102);
      }
      if (!filled.get(105)) {
        codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice.add(
            BigInteger.ZERO);
        this.filled.set(105);
      }
      if (!filled.get(104)) {
        codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit
            .add(BigInteger.ZERO);
        this.filled.set(104);
      }
      if (!filled.get(107)) {
        codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal
            .add(BigInteger.ZERO);
        this.filled.set(107);
      }
      if (!filled.get(106)) {
        codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount.add(
            BigInteger.ZERO);
        this.filled.set(106);
      }
      if (!filled.get(8)) {
        contextGetsRevertedFlag.add(false);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        contextMayChangeFlag.add(false);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        contextNumber.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        contextNumberNew.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        contextRevertStamp.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        contextSelfRevertsFlag.add(false);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        contextWillRevertFlag.add(false);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        counterNsr.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        counterTli.add(false);
        this.filled.set(16);
      }
      if (!filled.get(110)) {
        depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas.add(BigInteger.ZERO);
        this.filled.set(110);
      }
      if (!filled.get(109)) {
        depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize.add(BigInteger.ZERO);
        this.filled.set(109);
      }
      if (!filled.get(50)) {
        depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode
            .add(false);
        this.filled.set(50);
      }
      if (!filled.get(49)) {
        depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559.add(false);
        this.filled.set(49);
      }
      if (!filled.get(108)) {
        deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance
            .add(BigInteger.ZERO);
        this.filled.set(108);
      }
      if (!filled.get(48)) {
        deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment
            .add(false);
        this.filled.set(48);
      }
      if (!filled.get(17)) {
        domStamp.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        exceptionAhoyFlag.add(false);
        this.filled.set(18);
      }
      if (!filled.get(52)) {
        existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig.add(false);
        this.filled.set(52);
      }
      if (!filled.get(51)) {
        existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution
            .add(false);
        this.filled.set(51);
      }
      if (!filled.get(19)) {
        gasActual.add(BigInteger.ZERO);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        gasCost.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        gasExpected.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        gasNext.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        gasRefund.add(BigInteger.ZERO);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        gasRefundNew.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(54)) {
        hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero.add(
            false);
        this.filled.set(54);
      }
      if (!filled.get(53)) {
        hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero.add(false);
        this.filled.set(53);
      }
      if (!filled.get(25)) {
        hashInfoStamp.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        hubStamp.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        hubStampTransactionEnd.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(55)) {
        isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm.add(false);
        this.filled.set(55);
      }
      if (!filled.get(56)) {
        isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew.add(false);
        this.filled.set(56);
      }
      if (!filled.get(57)) {
        isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2.add(false);
        this.filled.set(57);
      }
      if (!filled.get(58)) {
        isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3.add(false);
        this.filled.set(58);
      }
      if (!filled.get(59)) {
        isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4.add(false);
        this.filled.set(59);
      }
      if (!filled.get(60)) {
        isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag.add(false);
        this.filled.set(60);
      }
      if (!filled.get(61)) {
        isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag.add(false);
        this.filled.set(61);
      }
      if (!filled.get(62)) {
        isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag.add(false);
        this.filled.set(62);
      }
      if (!filled.get(63)) {
        isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag.add(false);
        this.filled.set(63);
      }
      if (!filled.get(64)) {
        isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag.add(false);
        this.filled.set(64);
      }
      if (!filled.get(28)) {
        mmuStamp.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(121)) {
        mxpSize1HiXorStackItemValueLo2.add(BigInteger.ZERO);
        this.filled.set(121);
      }
      if (!filled.get(122)) {
        mxpSize1LoXorStackItemValueLo3.add(BigInteger.ZERO);
        this.filled.set(122);
      }
      if (!filled.get(123)) {
        mxpSize2HiXorStackItemValueLo4.add(BigInteger.ZERO);
        this.filled.set(123);
      }
      if (!filled.get(124)) {
        mxpSize2LoXorStaticGas.add(BigInteger.ZERO);
        this.filled.set(124);
      }
      if (!filled.get(29)) {
        mxpStamp.add(BigInteger.ZERO);
        this.filled.set(29);
      }
      if (!filled.get(125)) {
        mxpWords.add(BigInteger.ZERO);
        this.filled.set(125);
      }
      if (!filled.get(112)) {
        nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce.add(BigInteger.ZERO);
        this.filled.set(112);
      }
      if (!filled.get(111)) {
        nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas.add(BigInteger.ZERO);
        this.filled.set(111);
      }
      if (!filled.get(30)) {
        numberOfNonStackRows.add(BigInteger.ZERO);
        this.filled.set(30);
      }
      if (!filled.get(126)) {
        oobInst.add(BigInteger.ZERO);
        this.filled.set(126);
      }
      if (!filled.get(127)) {
        oobOutgoingData1.add(BigInteger.ZERO);
        this.filled.set(127);
      }
      if (!filled.get(128)) {
        oobOutgoingData2.add(BigInteger.ZERO);
        this.filled.set(128);
      }
      if (!filled.get(129)) {
        oobOutgoingData3.add(BigInteger.ZERO);
        this.filled.set(129);
      }
      if (!filled.get(130)) {
        oobOutgoingData4.add(BigInteger.ZERO);
        this.filled.set(130);
      }
      if (!filled.get(131)) {
        oobOutgoingData5.add(BigInteger.ZERO);
        this.filled.set(131);
      }
      if (!filled.get(132)) {
        oobOutgoingData6.add(BigInteger.ZERO);
        this.filled.set(132);
      }
      if (!filled.get(31)) {
        peekAtAccount.add(false);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        peekAtContext.add(false);
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        peekAtMiscellaneous.add(false);
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        peekAtScenario.add(false);
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        peekAtStack.add(false);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        peekAtStorage.add(false);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        peekAtTransaction.add(false);
        this.filled.set(37);
      }
      if (!filled.get(133)) {
        precinfoAddrLo.add(BigInteger.ZERO);
        this.filled.set(133);
      }
      if (!filled.get(134)) {
        precinfoCds.add(BigInteger.ZERO);
        this.filled.set(134);
      }
      if (!filled.get(135)) {
        precinfoExecCost.add(BigInteger.ZERO);
        this.filled.set(135);
      }
      if (!filled.get(136)) {
        precinfoProvidesReturnData.add(BigInteger.ZERO);
        this.filled.set(136);
      }
      if (!filled.get(137)) {
        precinfoRds.add(BigInteger.ZERO);
        this.filled.set(137);
      }
      if (!filled.get(138)) {
        precinfoSuccess.add(BigInteger.ZERO);
        this.filled.set(138);
      }
      if (!filled.get(139)) {
        precinfoTouchesRam.add(BigInteger.ZERO);
        this.filled.set(139);
      }
      if (!filled.get(38)) {
        programCounter.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        programCounterNew.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(80)) {
        pushpopFlag.add(false);
        this.filled.set(80);
      }
      if (!filled.get(81)) {
        rdcx.add(false);
        this.filled.set(81);
      }
      if (!filled.get(69)) {
        ripemDsub160XorKecFlag.add(false);
        this.filled.set(69);
      }
      if (!filled.get(113)) {
        rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi.add(
            BigInteger.ZERO);
        this.filled.set(113);
      }
      if (!filled.get(114)) {
        rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo
            .add(BigInteger.ZERO);
        this.filled.set(114);
      }
      if (!filled.get(65)) {
        rlpaddrFlagXorEcpairingXorInvprex.add(false);
        this.filled.set(65);
      }
      if (!filled.get(115)) {
        rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue.add(
            BigInteger.ZERO);
        this.filled.set(115);
      }
      if (!filled.get(116)) {
        rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1.add(BigInteger.ZERO);
        this.filled.set(116);
      }
      if (!filled.get(117)) {
        rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2.add(BigInteger.ZERO);
        this.filled.set(117);
      }
      if (!filled.get(118)) {
        rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3.add(BigInteger.ZERO);
        this.filled.set(118);
      }
      if (!filled.get(119)) {
        rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4.add(BigInteger.ZERO);
        this.filled.set(119);
      }
      if (!filled.get(70)) {
        scnFailure1XorLogFlag.add(false);
        this.filled.set(70);
      }
      if (!filled.get(71)) {
        scnFailure2XorMachineStateFlag.add(false);
        this.filled.set(71);
      }
      if (!filled.get(72)) {
        scnFailure3XorMaxcsx.add(false);
        this.filled.set(72);
      }
      if (!filled.get(73)) {
        scnFailure4XorModFlag.add(false);
        this.filled.set(73);
      }
      if (!filled.get(74)) {
        scnSuccess1XorMulFlag.add(false);
        this.filled.set(74);
      }
      if (!filled.get(75)) {
        scnSuccess2XorMxpx.add(false);
        this.filled.set(75);
      }
      if (!filled.get(76)) {
        scnSuccess3XorMxpFlag.add(false);
        this.filled.set(76);
      }
      if (!filled.get(77)) {
        scnSuccess4XorOobFlag.add(false);
        this.filled.set(77);
      }
      if (!filled.get(78)) {
        selfdestructXorOogx.add(false);
        this.filled.set(78);
      }
      if (!filled.get(79)) {
        sha2Sub256XorOpcx.add(false);
        this.filled.set(79);
      }
      if (!filled.get(82)) {
        shfFlag.add(false);
        this.filled.set(82);
      }
      if (!filled.get(83)) {
        sox.add(false);
        this.filled.set(83);
      }
      if (!filled.get(84)) {
        sstorex.add(false);
        this.filled.set(84);
      }
      if (!filled.get(86)) {
        stackItemPop1.add(false);
        this.filled.set(86);
      }
      if (!filled.get(87)) {
        stackItemPop2.add(false);
        this.filled.set(87);
      }
      if (!filled.get(88)) {
        stackItemPop3.add(false);
        this.filled.set(88);
      }
      if (!filled.get(89)) {
        stackItemPop4.add(false);
        this.filled.set(89);
      }
      if (!filled.get(85)) {
        stackramFlag.add(false);
        this.filled.set(85);
      }
      if (!filled.get(91)) {
        staticFlag.add(false);
        this.filled.set(91);
      }
      if (!filled.get(90)) {
        staticx.add(false);
        this.filled.set(90);
      }
      if (!filled.get(92)) {
        stoFlag.add(false);
        this.filled.set(92);
      }
      if (!filled.get(140)) {
        stpGasHi.add(BigInteger.ZERO);
        this.filled.set(140);
      }
      if (!filled.get(141)) {
        stpGasLo.add(BigInteger.ZERO);
        this.filled.set(141);
      }
      if (!filled.get(142)) {
        stpGasOopkt.add(BigInteger.ZERO);
        this.filled.set(142);
      }
      if (!filled.get(143)) {
        stpGasStpd.add(BigInteger.ZERO);
        this.filled.set(143);
      }
      if (!filled.get(144)) {
        stpInst.add(BigInteger.ZERO);
        this.filled.set(144);
      }
      if (!filled.get(145)) {
        stpValHi.add(BigInteger.ZERO);
        this.filled.set(145);
      }
      if (!filled.get(146)) {
        stpValLo.add(BigInteger.ZERO);
        this.filled.set(146);
      }
      if (!filled.get(40)) {
        subStamp.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(93)) {
        sux.add(false);
        this.filled.set(93);
      }
      if (!filled.get(94)) {
        swapFlag.add(false);
        this.filled.set(94);
      }
      if (!filled.get(41)) {
        transactionReverts.add(false);
        this.filled.set(41);
      }
      if (!filled.get(95)) {
        trmFlag.add(false);
        this.filled.set(95);
      }
      if (!filled.get(66)) {
        trmFlagXorEcrecoverXorJumpx.add(false);
        this.filled.set(66);
      }
      if (!filled.get(120)) {
        trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1.add(BigInteger.ZERO);
        this.filled.set(120);
      }
      if (!filled.get(42)) {
        twoLineInstruction.add(false);
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        txExec.add(false);
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        txFinl.add(false);
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        txInit.add(false);
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        txSkip.add(false);
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        txWarm.add(false);
        this.filled.set(47);
      }
      if (!filled.get(96)) {
        txnFlag.add(false);
        this.filled.set(96);
      }
      if (!filled.get(68)) {
        warmNewXorModexpXorJumpFlag.add(false);
        this.filled.set(68);
      }
      if (!filled.get(67)) {
        warmXorIdentityXorJumpDestinationVettingRequired.add(false);
        this.filled.set(67);
      }
      if (!filled.get(97)) {
        wcpFlag.add(false);
        this.filled.set(97);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          absoluteTransactionNumber,
          addrHiXorAccountAddressHiXorCcrsStampXorHashInfoKecHiXorAddressHiXorBasefee,
          addrLoXorAccountAddressLoXorExpDyncostXorHashInfoKecLoXorAddressLoXorCallDataSize,
          balanceNewXorByteCodeAddressHiXorExpExponentLoXorHeightXorStorageKeyHiXorCoinbaseAddressLo,
          balanceXorAccountDeploymentNumberXorExpExponentHiXorHashInfoSizeXorDeploymentNumberXorCoinbaseAddressHi,
          batchNumber,
          callerContextNumber,
          codeAddressHi,
          codeAddressLo,
          codeDeploymentNumber,
          codeDeploymentStatus,
          codeFragmentIndex,
          codeHashHiNewXorByteCodeDeploymentNumberXorMmuInstXorHeightOverXorValCurrHiXorFromAddressLo,
          codeHashHiXorByteCodeAddressLoXorMmuExoSumXorHeightNewXorStorageKeyLoXorFromAddressHi,
          codeHashLoNewXorCallerAddressHiXorMmuOffset2HiXorInstXorValNextHiXorGasPrice,
          codeHashLoXorByteCodeDeploymentStatusXorMmuOffset1LoXorHeightUnderXorValCurrLoXorGasLimit,
          codeSizeNewXorCallerContextNumberXorMmuParam1XorPushValueLoXorValOrigHiXorGasRefundCounterFinal,
          codeSizeXorCallerAddressLoXorMmuOffset2LoXorPushValueHiXorValNextLoXorGasRefundAmount,
          contextGetsRevertedFlag,
          contextMayChangeFlag,
          contextNumber,
          contextNumberNew,
          contextRevertStamp,
          contextSelfRevertsFlag,
          contextWillRevertFlag,
          counterNsr,
          counterTli,
          depNumNewXorCallStackDepthXorMmuRefSizeXorStackItemHeight3XorInitGas,
          depNumXorCallDataSizeXorMmuRefOffsetXorStackItemHeight2XorInitCodeSize,
          depStatusNewXorExpFlagXorCallEoaSuccessCallerWillRevertXorBinFlagXorValCurrIsZeroXorStatusCode,
          depStatusXorCcsrFlagXorCallAbortXorAddFlagXorValCurrIsOrigXorIsEip1559,
          deploymentNumberInftyXorCallDataOffsetXorMmuParam2XorStackItemHeight1XorValOrigLoXorInitialBalance,
          deploymentStatusInftyXorUpdateXorAbortFlagXorBlake2FXorAccFlagXorValCurrChangesXorIsDeployment,
          domStamp,
          exceptionAhoyFlag,
          existsNewXorMmuFlagXorCallPrcFailureCallerWillRevertXorCallFlagXorValNextIsOrig,
          existsXorFcondFlagXorCallEoaSuccessCallerWontRevertXorBtcFlagXorValNextIsCurrXorTxnRequiresEvmExecution,
          gasActual,
          gasCost,
          gasExpected,
          gasNext,
          gasRefund,
          gasRefundNew,
          hasCodeNewXorMxpDeploysXorCallPrcSuccessCallerWillRevertXorCopyFlagXorValOrigIsZero,
          hasCodeXorMmuInfoXorCallPrcFailureCallerWontRevertXorConFlagXorValNextIsZero,
          hashInfoStamp,
          hubStamp,
          hubStampTransactionEnd,
          isBlake2FXorMxpFlagXorCallPrcSuccessCallerWontRevertXorCreateFlagXorWarm,
          isEcaddXorMxpMxpxXorCallSmcFailureCallerWillRevertXorDecodedFlag1XorWarmNew,
          isEcmulXorOobEvent1XorCallSmcFailureCallerWontRevertXorDecodedFlag2,
          isEcpairingXorOobEvent2XorCallSmcSuccessCallerWillRevertXorDecodedFlag3,
          isEcrecoverXorOobFlagXorCallSmcSuccessCallerWontRevertXorDecodedFlag4,
          isIdentityXorPrecinfoFlagXorCodedepositXorDupFlag,
          isModexpXorStpExistsXorCodedepositInvalidCodePrefixXorExtFlag,
          isPrecompileXorStpFlagXorCodedepositValidCodePrefixXorHaltFlag,
          isRipemDsub160XorStpOogxXorEcaddXorHashInfoFlag,
          isSha2Sub256XorStpWarmXorEcmulXorInvalidFlag,
          mmuStamp,
          mxpSize1HiXorStackItemValueLo2,
          mxpSize1LoXorStackItemValueLo3,
          mxpSize2HiXorStackItemValueLo4,
          mxpSize2LoXorStaticGas,
          mxpStamp,
          mxpWords,
          nonceNewXorContextNumberXorMmuSizeXorStackItemStamp1XorNonce,
          nonceXorCallValueXorMmuReturnerXorStackItemHeight4XorLeftoverGas,
          numberOfNonStackRows,
          oobInst,
          oobOutgoingData1,
          oobOutgoingData2,
          oobOutgoingData3,
          oobOutgoingData4,
          oobOutgoingData5,
          oobOutgoingData6,
          peekAtAccount,
          peekAtContext,
          peekAtMiscellaneous,
          peekAtScenario,
          peekAtStack,
          peekAtStorage,
          peekAtTransaction,
          precinfoAddrLo,
          precinfoCds,
          precinfoExecCost,
          precinfoProvidesReturnData,
          precinfoRds,
          precinfoSuccess,
          precinfoTouchesRam,
          programCounter,
          programCounterNew,
          pushpopFlag,
          rdcx,
          ripemDsub160XorKecFlag,
          rlpaddrDepAddrHiXorIsStaticXorMmuStackValHiXorStackItemStamp2XorToAddressHi,
          rlpaddrDepAddrLoXorReturnerContextNumberXorMmuStackValLoXorStackItemStamp3XorToAddressLo,
          rlpaddrFlagXorEcpairingXorInvprex,
          rlpaddrKecHiXorReturnerIsPrecompileXorMxpGasMxpXorStackItemStamp4XorValue,
          rlpaddrKecLoXorReturnAtOffsetXorMxpInstXorStackItemValueHi1,
          rlpaddrRecipeXorReturnAtSizeXorMxpOffset1HiXorStackItemValueHi2,
          rlpaddrSaltHiXorReturnDataOffsetXorMxpOffset1LoXorStackItemValueHi3,
          rlpaddrSaltLoXorReturnDataSizeXorMxpOffset2HiXorStackItemValueHi4,
          scnFailure1XorLogFlag,
          scnFailure2XorMachineStateFlag,
          scnFailure3XorMaxcsx,
          scnFailure4XorModFlag,
          scnSuccess1XorMulFlag,
          scnSuccess2XorMxpx,
          scnSuccess3XorMxpFlag,
          scnSuccess4XorOobFlag,
          selfdestructXorOogx,
          sha2Sub256XorOpcx,
          shfFlag,
          sox,
          sstorex,
          stackItemPop1,
          stackItemPop2,
          stackItemPop3,
          stackItemPop4,
          stackramFlag,
          staticFlag,
          staticx,
          stoFlag,
          stpGasHi,
          stpGasLo,
          stpGasOopkt,
          stpGasStpd,
          stpInst,
          stpValHi,
          stpValLo,
          subStamp,
          sux,
          swapFlag,
          transactionReverts,
          trmFlag,
          trmFlagXorEcrecoverXorJumpx,
          trmRawAddrHiXorMxpOffset2LoXorStackItemValueLo1,
          twoLineInstruction,
          txExec,
          txFinl,
          txInit,
          txSkip,
          txWarm,
          txnFlag,
          warmNewXorModexpXorJumpFlag,
          warmXorIdentityXorJumpDestinationVettingRequired,
          wcpFlag);
    }
  }
}
