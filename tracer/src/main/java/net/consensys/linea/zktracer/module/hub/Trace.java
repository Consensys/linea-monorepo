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
record Trace(
    @JsonProperty("ABORT_FLAG") List<Boolean> abortFlag,
    @JsonProperty("ABSOLUTE_TRANSACTION_NUMBER") List<BigInteger> absoluteTransactionNumber,
    @JsonProperty("ACC_FLAG_xor_VAL_NEXT_IS_CURR_xor_WARM")
        List<Boolean> accFlagXorValNextIsCurrXorWarm,
    @JsonProperty("ADD_FLAG") List<Boolean> addFlag,
    @JsonProperty("ADDRESS_HI_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_STACK_ITEM_VALUE_LO_1")
        List<BigInteger> addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1,
    @JsonProperty(
            "ADDRESS_LO_xor_VAL_CURR_LO_xor_CALLER_ADDRESS_HI_xor_STACK_ITEM_HEIGHT_1_xor_FROM_ADDRESS_HI")
        List<BigInteger> addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi,
    @JsonProperty(
            "BALANCE_NEW_xor_VAL_CURR_HI_xor_CALL_STACK_DEPTH_xor_STACK_ITEM_VALUE_LO_2_xor_GAS_TIP")
        List<BigInteger> balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip,
    @JsonProperty(
            "BALANCE_xor_STORAGE_KEY_HI_xor_BYTE_CODE_ADDRESS_HI_xor_STACK_ITEM_VALUE_LO_4_xor_GAS_FEE")
        List<BigInteger> balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee,
    @JsonProperty("BATCH_NUMBER") List<BigInteger> batchNumber,
    @JsonProperty("BIN_FLAG") List<Boolean> binFlag,
    @JsonProperty("BTC_FLAG") List<Boolean> btcFlag,
    @JsonProperty("BYTE_CODE_ADDRESS_LO_xor_STACK_ITEM_STAMP_4")
        List<BigInteger> byteCodeAddressLoXorStackItemStamp4,
    @JsonProperty("CALL_FLAG") List<Boolean> callFlag,
    @JsonProperty("CALLER_CONTEXT_NUMBER") List<BigInteger> callerContextNumber,
    @JsonProperty("CALLER_CONTEXT_NUMBER_xor_PUSH_VALUE_LO")
        List<BigInteger> callerContextNumberXorPushValueLo,
    @JsonProperty("CODE_ADDRESS_HI") List<BigInteger> codeAddressHi,
    @JsonProperty("CODE_ADDRESS_LO") List<BigInteger> codeAddressLo,
    @JsonProperty("CODE_DEPLOYMENT_NUMBER") List<BigInteger> codeDeploymentNumber,
    @JsonProperty("CODE_DEPLOYMENT_STATUS") List<Boolean> codeDeploymentStatus,
    @JsonProperty(
            "CODE_HASH_HI_NEW_xor_ADDRESS_LO_xor_CALLER_ADDRESS_LO_xor_STACK_ITEM_VALUE_LO_3_xor_ABSOLUTE_TRANSACTION_NUMBER")
        List<BigInteger>
            codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber,
    @JsonProperty("CODE_HASH_HI_xor_ACCOUNT_ADDRESS_HI_xor_STACK_ITEM_HEIGHT_3_xor_NONCE")
        List<BigInteger> codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce,
    @JsonProperty("CODE_HASH_LO_NEW_xor_CALL_VALUE_xor_STACK_ITEM_STAMP_1")
        List<BigInteger> codeHashLoNewXorCallValueXorStackItemStamp1,
    @JsonProperty(
            "CODE_HASH_LO_xor_VAL_ORIG_HI_xor_RETURN_DATA_SIZE_xor_HEIGHT_UNDER_xor_BATCH_NUMBER")
        List<BigInteger> codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber,
    @JsonProperty(
            "CODE_SIZE_NEW_xor_VAL_NEXT_LO_xor_RETURNER_CONTEXT_NUMBER_xor_STATIC_GAS_xor_VALUE")
        List<BigInteger> codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue,
    @JsonProperty("CODE_SIZE_xor_CALL_DATA_OFFSET_xor_STACK_ITEM_STAMP_2")
        List<BigInteger> codeSizeXorCallDataOffsetXorStackItemStamp2,
    @JsonProperty("CON_FLAG") List<Boolean> conFlag,
    @JsonProperty("CONTEXT_GETS_REVRTD_FLAG") List<Boolean> contextGetsRevrtdFlag,
    @JsonProperty("CONTEXT_MAY_CHANGE_FLAG") List<Boolean> contextMayChangeFlag,
    @JsonProperty("CONTEXT_NUMBER") List<BigInteger> contextNumber,
    @JsonProperty("CONTEXT_NUMBER_NEW") List<BigInteger> contextNumberNew,
    @JsonProperty("CONTEXT_REVERT_STAMP") List<BigInteger> contextRevertStamp,
    @JsonProperty("CONTEXT_SELF_REVRTS_FLAG") List<Boolean> contextSelfRevrtsFlag,
    @JsonProperty("CONTEXT_WILL_REVERT_FLAG") List<Boolean> contextWillRevertFlag,
    @JsonProperty("COPY_FLAG") List<Boolean> copyFlag,
    @JsonProperty("COUNTER_NSR") List<BigInteger> counterNsr,
    @JsonProperty("COUNTER_TLI") List<Boolean> counterTli,
    @JsonProperty("CREATE_FLAG") List<Boolean> createFlag,
    @JsonProperty("DECODED_FLAG_1") List<Boolean> decodedFlag1,
    @JsonProperty("DECODED_FLAG_2_xor_VAL_ORIG_IS_ZERO_xor_HAS_CODE")
        List<Boolean> decodedFlag2XorValOrigIsZeroXorHasCode,
    @JsonProperty("DECODED_FLAG_3") List<Boolean> decodedFlag3,
    @JsonProperty("DECODED_FLAG_4") List<Boolean> decodedFlag4,
    @JsonProperty(
            "DEPLOYMENT_NUMBER_INFTY_xor_DEPLOYMENT_NUMBER_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_STACK_ITEM_VALUE_HI_4_xor_FROM_ADDRESS_LO")
        List<BigInteger>
            deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo,
    @JsonProperty(
            "DEPLOYMENT_NUMBER_NEW_xor_STORAGE_KEY_LO_xor_ACCOUNT_ADDRESS_LO_xor_STACK_ITEM_VALUE_HI_1_xor_TO_ADDRESS_HI")
        List<BigInteger>
            deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi,
    @JsonProperty(
            "DEPLOYMENT_NUMBER_xor_ADDRESS_HI_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_HEIGHT_OVER_xor_INIT_GAS")
        List<BigInteger>
            deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas,
    @JsonProperty("DEPLOYMENT_STATUS_NEW_xor_IS_STATIC_xor_INSTRUCTION")
        List<BigInteger> deploymentStatusNewXorIsStaticXorInstruction,
    @JsonProperty("DEPLOYMENT_STATUS_xor_RETURN_AT_SIZE_xor_HEIGHT")
        List<BigInteger> deploymentStatusXorReturnAtSizeXorHeight,
    @JsonProperty("DUP_FLAG_xor_VAL_NEXT_IS_ORIG_xor_IS_PRECOMPILE")
        List<Boolean> dupFlagXorValNextIsOrigXorIsPrecompile,
    @JsonProperty("EXCEPTION_AHOY_FLAG") List<Boolean> exceptionAhoyFlag,
    @JsonProperty("EXT_FLAG") List<Boolean> extFlag,
    @JsonProperty("FAILURE_CONDITION_FLAG") List<Boolean> failureConditionFlag,
    @JsonProperty("GAS_ACTUAL") List<BigInteger> gasActual,
    @JsonProperty("GAS_COST") List<BigInteger> gasCost,
    @JsonProperty("GAS_EXPECTED") List<BigInteger> gasExpected,
    @JsonProperty("GAS_MEMORY_EXPANSION") List<BigInteger> gasMemoryExpansion,
    @JsonProperty("GAS_NEXT") List<BigInteger> gasNext,
    @JsonProperty("GAS_REFUND") List<BigInteger> gasRefund,
    @JsonProperty("HALT_FLAG") List<Boolean> haltFlag,
    @JsonProperty("HUB_STAMP") List<BigInteger> hubStamp,
    @JsonProperty("INVALID_FLAG") List<Boolean> invalidFlag,
    @JsonProperty("INVPREX") List<Boolean> invprex,
    @JsonProperty("JUMP_FLAG") List<Boolean> jumpFlag,
    @JsonProperty("JUMPX") List<Boolean> jumpx,
    @JsonProperty("KEC_FLAG") List<Boolean> kecFlag,
    @JsonProperty("LOG_FLAG") List<Boolean> logFlag,
    @JsonProperty("MAXCSX_xor_WARM_NEW_xor_DEPLOYMENT_STATUS_INFTY")
        List<Boolean> maxcsxXorWarmNewXorDeploymentStatusInfty,
    @JsonProperty("MOD_FLAG") List<Boolean> modFlag,
    @JsonProperty("MUL_FLAG") List<Boolean> mulFlag,
    @JsonProperty("MXP_FLAG") List<Boolean> mxpFlag,
    @JsonProperty("MXPX") List<Boolean> mxpx,
    @JsonProperty(
            "NONCE_NEW_xor_VAL_NEXT_HI_xor_CONTEXT_NUMBER_xor_STACK_ITEM_HEIGHT_2_xor_TO_ADDRESS_LO")
        List<BigInteger> nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo,
    @JsonProperty("NONCE_xor_VAL_ORIG_LO_xor_CALL_DATA_SIZE_xor_HEIGHT_NEW_xor_GAS_MAXFEE")
        List<BigInteger> nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee,
    @JsonProperty("NUMBER_OF_NON_STACK_ROWS") List<BigInteger> numberOfNonStackRows,
    @JsonProperty("OOB_FLAG") List<Boolean> oobFlag,
    @JsonProperty("OOGX_xor_VAL_NEXT_IS_ZERO_xor_EXISTS")
        List<Boolean> oogxXorValNextIsZeroXorExists,
    @JsonProperty("OPCX") List<Boolean> opcx,
    @JsonProperty("PEEK_AT_ACCOUNT") List<Boolean> peekAtAccount,
    @JsonProperty("PEEK_AT_CONTEXT") List<Boolean> peekAtContext,
    @JsonProperty("PEEK_AT_STACK") List<Boolean> peekAtStack,
    @JsonProperty("PEEK_AT_STORAGE") List<Boolean> peekAtStorage,
    @JsonProperty("PEEK_AT_TRANSACTION") List<Boolean> peekAtTransaction,
    @JsonProperty("PROGRAM_COUNTER") List<BigInteger> programCounter,
    @JsonProperty("PROGRAM_COUNTER_NEW") List<BigInteger> programCounterNew,
    @JsonProperty("PUSHPOP_FLAG_xor_VAL_CURR_IS_ORIG_xor_WARM_NEW")
        List<Boolean> pushpopFlagXorValCurrIsOrigXorWarmNew,
    @JsonProperty("RDCX") List<Boolean> rdcx,
    @JsonProperty("RETURN_AT_OFFSET_xor_PUSH_VALUE_HI")
        List<BigInteger> returnAtOffsetXorPushValueHi,
    @JsonProperty("RETURN_DATA_OFFSET_xor_STACK_ITEM_STAMP_3")
        List<BigInteger> returnDataOffsetXorStackItemStamp3,
    @JsonProperty("RETURNER_IS_PRECOMPILE_xor_STACK_ITEM_HEIGHT_4")
        List<BigInteger> returnerIsPrecompileXorStackItemHeight4,
    @JsonProperty("SHF_FLAG") List<Boolean> shfFlag,
    @JsonProperty("SOX") List<Boolean> sox,
    @JsonProperty("SSTOREX") List<Boolean> sstorex,
    @JsonProperty("STACK_ITEM_POP_1") List<Boolean> stackItemPop1,
    @JsonProperty("STACK_ITEM_POP_2") List<Boolean> stackItemPop2,
    @JsonProperty("STACK_ITEM_POP_3") List<Boolean> stackItemPop3,
    @JsonProperty("STACK_ITEM_POP_4") List<Boolean> stackItemPop4,
    @JsonProperty("STACK_ITEM_VALUE_HI_2") List<BigInteger> stackItemValueHi2,
    @JsonProperty("STACK_ITEM_VALUE_HI_3") List<BigInteger> stackItemValueHi3,
    @JsonProperty("STACKRAM_FLAG") List<Boolean> stackramFlag,
    @JsonProperty("STATIC_FLAG_xor_WARM_xor_HAS_CODE_NEW")
        List<Boolean> staticFlagXorWarmXorHasCodeNew,
    @JsonProperty("STATICX_xor_UPDATE_xor_VAL_CURR_CHANGES_xor_IS_DEPLOYMENT_xor_EXISTS_NEW")
        List<Boolean> staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew,
    @JsonProperty("STO_FLAG") List<Boolean> stoFlag,
    @JsonProperty("SUX") List<Boolean> sux,
    @JsonProperty("SWAP_FLAG") List<Boolean> swapFlag,
    @JsonProperty("TRANSACTION_END_STAMP") List<BigInteger> transactionEndStamp,
    @JsonProperty("TRANSACTION_REVERTS") List<BigInteger> transactionReverts,
    @JsonProperty("TRM_FLAG_xor_VAL_CURR_IS_ZERO_xor_SUFFICIENT_BALANCE")
        List<Boolean> trmFlagXorValCurrIsZeroXorSufficientBalance,
    @JsonProperty("TWO_LINE_INSTRUCTION") List<Boolean> twoLineInstruction,
    @JsonProperty("TX_EXEC") List<Boolean> txExec,
    @JsonProperty("TX_FINL") List<Boolean> txFinl,
    @JsonProperty("TX_INIT") List<Boolean> txInit,
    @JsonProperty("TX_SKIP") List<Boolean> txSkip,
    @JsonProperty("TX_WARM") List<Boolean> txWarm,
    @JsonProperty("TXN_FLAG") List<Boolean> txnFlag,
    @JsonProperty("WCP_FLAG") List<Boolean> wcpFlag) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ABORT_FLAG")
    private final List<Boolean> abortFlag = new ArrayList<>();

    @JsonProperty("ABSOLUTE_TRANSACTION_NUMBER")
    private final List<BigInteger> absoluteTransactionNumber = new ArrayList<>();

    @JsonProperty("ACC_FLAG_xor_VAL_NEXT_IS_CURR_xor_WARM")
    private final List<Boolean> accFlagXorValNextIsCurrXorWarm = new ArrayList<>();

    @JsonProperty("ADD_FLAG")
    private final List<Boolean> addFlag = new ArrayList<>();

    @JsonProperty("ADDRESS_HI_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_STACK_ITEM_VALUE_LO_1")
    private final List<BigInteger> addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1 =
        new ArrayList<>();

    @JsonProperty(
        "ADDRESS_LO_xor_VAL_CURR_LO_xor_CALLER_ADDRESS_HI_xor_STACK_ITEM_HEIGHT_1_xor_FROM_ADDRESS_HI")
    private final List<BigInteger>
        addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi =
            new ArrayList<>();

    @JsonProperty(
        "BALANCE_NEW_xor_VAL_CURR_HI_xor_CALL_STACK_DEPTH_xor_STACK_ITEM_VALUE_LO_2_xor_GAS_TIP")
    private final List<BigInteger>
        balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip = new ArrayList<>();

    @JsonProperty(
        "BALANCE_xor_STORAGE_KEY_HI_xor_BYTE_CODE_ADDRESS_HI_xor_STACK_ITEM_VALUE_LO_4_xor_GAS_FEE")
    private final List<BigInteger>
        balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee = new ArrayList<>();

    @JsonProperty("BATCH_NUMBER")
    private final List<BigInteger> batchNumber = new ArrayList<>();

    @JsonProperty("BIN_FLAG")
    private final List<Boolean> binFlag = new ArrayList<>();

    @JsonProperty("BTC_FLAG")
    private final List<Boolean> btcFlag = new ArrayList<>();

    @JsonProperty("BYTE_CODE_ADDRESS_LO_xor_STACK_ITEM_STAMP_4")
    private final List<BigInteger> byteCodeAddressLoXorStackItemStamp4 = new ArrayList<>();

    @JsonProperty("CALL_FLAG")
    private final List<Boolean> callFlag = new ArrayList<>();

    @JsonProperty("CALLER_CONTEXT_NUMBER")
    private final List<BigInteger> callerContextNumber = new ArrayList<>();

    @JsonProperty("CALLER_CONTEXT_NUMBER_xor_PUSH_VALUE_LO")
    private final List<BigInteger> callerContextNumberXorPushValueLo = new ArrayList<>();

    @JsonProperty("CODE_ADDRESS_HI")
    private final List<BigInteger> codeAddressHi = new ArrayList<>();

    @JsonProperty("CODE_ADDRESS_LO")
    private final List<BigInteger> codeAddressLo = new ArrayList<>();

    @JsonProperty("CODE_DEPLOYMENT_NUMBER")
    private final List<BigInteger> codeDeploymentNumber = new ArrayList<>();

    @JsonProperty("CODE_DEPLOYMENT_STATUS")
    private final List<Boolean> codeDeploymentStatus = new ArrayList<>();

    @JsonProperty(
        "CODE_HASH_HI_NEW_xor_ADDRESS_LO_xor_CALLER_ADDRESS_LO_xor_STACK_ITEM_VALUE_LO_3_xor_ABSOLUTE_TRANSACTION_NUMBER")
    private final List<BigInteger>
        codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber =
            new ArrayList<>();

    @JsonProperty("CODE_HASH_HI_xor_ACCOUNT_ADDRESS_HI_xor_STACK_ITEM_HEIGHT_3_xor_NONCE")
    private final List<BigInteger> codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce =
        new ArrayList<>();

    @JsonProperty("CODE_HASH_LO_NEW_xor_CALL_VALUE_xor_STACK_ITEM_STAMP_1")
    private final List<BigInteger> codeHashLoNewXorCallValueXorStackItemStamp1 = new ArrayList<>();

    @JsonProperty(
        "CODE_HASH_LO_xor_VAL_ORIG_HI_xor_RETURN_DATA_SIZE_xor_HEIGHT_UNDER_xor_BATCH_NUMBER")
    private final List<BigInteger>
        codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber = new ArrayList<>();

    @JsonProperty(
        "CODE_SIZE_NEW_xor_VAL_NEXT_LO_xor_RETURNER_CONTEXT_NUMBER_xor_STATIC_GAS_xor_VALUE")
    private final List<BigInteger>
        codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue = new ArrayList<>();

    @JsonProperty("CODE_SIZE_xor_CALL_DATA_OFFSET_xor_STACK_ITEM_STAMP_2")
    private final List<BigInteger> codeSizeXorCallDataOffsetXorStackItemStamp2 = new ArrayList<>();

    @JsonProperty("CON_FLAG")
    private final List<Boolean> conFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_GETS_REVRTD_FLAG")
    private final List<Boolean> contextGetsRevrtdFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_MAY_CHANGE_FLAG")
    private final List<Boolean> contextMayChangeFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_NUMBER")
    private final List<BigInteger> contextNumber = new ArrayList<>();

    @JsonProperty("CONTEXT_NUMBER_NEW")
    private final List<BigInteger> contextNumberNew = new ArrayList<>();

    @JsonProperty("CONTEXT_REVERT_STAMP")
    private final List<BigInteger> contextRevertStamp = new ArrayList<>();

    @JsonProperty("CONTEXT_SELF_REVRTS_FLAG")
    private final List<Boolean> contextSelfRevrtsFlag = new ArrayList<>();

    @JsonProperty("CONTEXT_WILL_REVERT_FLAG")
    private final List<Boolean> contextWillRevertFlag = new ArrayList<>();

    @JsonProperty("COPY_FLAG")
    private final List<Boolean> copyFlag = new ArrayList<>();

    @JsonProperty("COUNTER_NSR")
    private final List<BigInteger> counterNsr = new ArrayList<>();

    @JsonProperty("COUNTER_TLI")
    private final List<Boolean> counterTli = new ArrayList<>();

    @JsonProperty("CREATE_FLAG")
    private final List<Boolean> createFlag = new ArrayList<>();

    @JsonProperty("DECODED_FLAG_1")
    private final List<Boolean> decodedFlag1 = new ArrayList<>();

    @JsonProperty("DECODED_FLAG_2_xor_VAL_ORIG_IS_ZERO_xor_HAS_CODE")
    private final List<Boolean> decodedFlag2XorValOrigIsZeroXorHasCode = new ArrayList<>();

    @JsonProperty("DECODED_FLAG_3")
    private final List<Boolean> decodedFlag3 = new ArrayList<>();

    @JsonProperty("DECODED_FLAG_4")
    private final List<Boolean> decodedFlag4 = new ArrayList<>();

    @JsonProperty(
        "DEPLOYMENT_NUMBER_INFTY_xor_DEPLOYMENT_NUMBER_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_STACK_ITEM_VALUE_HI_4_xor_FROM_ADDRESS_LO")
    private final List<BigInteger>
        deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo =
            new ArrayList<>();

    @JsonProperty(
        "DEPLOYMENT_NUMBER_NEW_xor_STORAGE_KEY_LO_xor_ACCOUNT_ADDRESS_LO_xor_STACK_ITEM_VALUE_HI_1_xor_TO_ADDRESS_HI")
    private final List<BigInteger>
        deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi =
            new ArrayList<>();

    @JsonProperty(
        "DEPLOYMENT_NUMBER_xor_ADDRESS_HI_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_HEIGHT_OVER_xor_INIT_GAS")
    private final List<BigInteger>
        deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas =
            new ArrayList<>();

    @JsonProperty("DEPLOYMENT_STATUS_NEW_xor_IS_STATIC_xor_INSTRUCTION")
    private final List<BigInteger> deploymentStatusNewXorIsStaticXorInstruction = new ArrayList<>();

    @JsonProperty("DEPLOYMENT_STATUS_xor_RETURN_AT_SIZE_xor_HEIGHT")
    private final List<BigInteger> deploymentStatusXorReturnAtSizeXorHeight = new ArrayList<>();

    @JsonProperty("DUP_FLAG_xor_VAL_NEXT_IS_ORIG_xor_IS_PRECOMPILE")
    private final List<Boolean> dupFlagXorValNextIsOrigXorIsPrecompile = new ArrayList<>();

    @JsonProperty("EXCEPTION_AHOY_FLAG")
    private final List<Boolean> exceptionAhoyFlag = new ArrayList<>();

    @JsonProperty("EXT_FLAG")
    private final List<Boolean> extFlag = new ArrayList<>();

    @JsonProperty("FAILURE_CONDITION_FLAG")
    private final List<Boolean> failureConditionFlag = new ArrayList<>();

    @JsonProperty("GAS_ACTUAL")
    private final List<BigInteger> gasActual = new ArrayList<>();

    @JsonProperty("GAS_COST")
    private final List<BigInteger> gasCost = new ArrayList<>();

    @JsonProperty("GAS_EXPECTED")
    private final List<BigInteger> gasExpected = new ArrayList<>();

    @JsonProperty("GAS_MEMORY_EXPANSION")
    private final List<BigInteger> gasMemoryExpansion = new ArrayList<>();

    @JsonProperty("GAS_NEXT")
    private final List<BigInteger> gasNext = new ArrayList<>();

    @JsonProperty("GAS_REFUND")
    private final List<BigInteger> gasRefund = new ArrayList<>();

    @JsonProperty("HALT_FLAG")
    private final List<Boolean> haltFlag = new ArrayList<>();

    @JsonProperty("HUB_STAMP")
    private final List<BigInteger> hubStamp = new ArrayList<>();

    @JsonProperty("INVALID_FLAG")
    private final List<Boolean> invalidFlag = new ArrayList<>();

    @JsonProperty("INVPREX")
    private final List<Boolean> invprex = new ArrayList<>();

    @JsonProperty("JUMP_FLAG")
    private final List<Boolean> jumpFlag = new ArrayList<>();

    @JsonProperty("JUMPX")
    private final List<Boolean> jumpx = new ArrayList<>();

    @JsonProperty("KEC_FLAG")
    private final List<Boolean> kecFlag = new ArrayList<>();

    @JsonProperty("LOG_FLAG")
    private final List<Boolean> logFlag = new ArrayList<>();

    @JsonProperty("MAXCSX_xor_WARM_NEW_xor_DEPLOYMENT_STATUS_INFTY")
    private final List<Boolean> maxcsxXorWarmNewXorDeploymentStatusInfty = new ArrayList<>();

    @JsonProperty("MOD_FLAG")
    private final List<Boolean> modFlag = new ArrayList<>();

    @JsonProperty("MUL_FLAG")
    private final List<Boolean> mulFlag = new ArrayList<>();

    @JsonProperty("MXP_FLAG")
    private final List<Boolean> mxpFlag = new ArrayList<>();

    @JsonProperty("MXPX")
    private final List<Boolean> mxpx = new ArrayList<>();

    @JsonProperty(
        "NONCE_NEW_xor_VAL_NEXT_HI_xor_CONTEXT_NUMBER_xor_STACK_ITEM_HEIGHT_2_xor_TO_ADDRESS_LO")
    private final List<BigInteger>
        nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo = new ArrayList<>();

    @JsonProperty("NONCE_xor_VAL_ORIG_LO_xor_CALL_DATA_SIZE_xor_HEIGHT_NEW_xor_GAS_MAXFEE")
    private final List<BigInteger> nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee =
        new ArrayList<>();

    @JsonProperty("NUMBER_OF_NON_STACK_ROWS")
    private final List<BigInteger> numberOfNonStackRows = new ArrayList<>();

    @JsonProperty("OOB_FLAG")
    private final List<Boolean> oobFlag = new ArrayList<>();

    @JsonProperty("OOGX_xor_VAL_NEXT_IS_ZERO_xor_EXISTS")
    private final List<Boolean> oogxXorValNextIsZeroXorExists = new ArrayList<>();

    @JsonProperty("OPCX")
    private final List<Boolean> opcx = new ArrayList<>();

    @JsonProperty("PEEK_AT_ACCOUNT")
    private final List<Boolean> peekAtAccount = new ArrayList<>();

    @JsonProperty("PEEK_AT_CONTEXT")
    private final List<Boolean> peekAtContext = new ArrayList<>();

    @JsonProperty("PEEK_AT_STACK")
    private final List<Boolean> peekAtStack = new ArrayList<>();

    @JsonProperty("PEEK_AT_STORAGE")
    private final List<Boolean> peekAtStorage = new ArrayList<>();

    @JsonProperty("PEEK_AT_TRANSACTION")
    private final List<Boolean> peekAtTransaction = new ArrayList<>();

    @JsonProperty("PROGRAM_COUNTER")
    private final List<BigInteger> programCounter = new ArrayList<>();

    @JsonProperty("PROGRAM_COUNTER_NEW")
    private final List<BigInteger> programCounterNew = new ArrayList<>();

    @JsonProperty("PUSHPOP_FLAG_xor_VAL_CURR_IS_ORIG_xor_WARM_NEW")
    private final List<Boolean> pushpopFlagXorValCurrIsOrigXorWarmNew = new ArrayList<>();

    @JsonProperty("RDCX")
    private final List<Boolean> rdcx = new ArrayList<>();

    @JsonProperty("RETURN_AT_OFFSET_xor_PUSH_VALUE_HI")
    private final List<BigInteger> returnAtOffsetXorPushValueHi = new ArrayList<>();

    @JsonProperty("RETURN_DATA_OFFSET_xor_STACK_ITEM_STAMP_3")
    private final List<BigInteger> returnDataOffsetXorStackItemStamp3 = new ArrayList<>();

    @JsonProperty("RETURNER_IS_PRECOMPILE_xor_STACK_ITEM_HEIGHT_4")
    private final List<BigInteger> returnerIsPrecompileXorStackItemHeight4 = new ArrayList<>();

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

    @JsonProperty("STACK_ITEM_VALUE_HI_2")
    private final List<BigInteger> stackItemValueHi2 = new ArrayList<>();

    @JsonProperty("STACK_ITEM_VALUE_HI_3")
    private final List<BigInteger> stackItemValueHi3 = new ArrayList<>();

    @JsonProperty("STACKRAM_FLAG")
    private final List<Boolean> stackramFlag = new ArrayList<>();

    @JsonProperty("STATIC_FLAG_xor_WARM_xor_HAS_CODE_NEW")
    private final List<Boolean> staticFlagXorWarmXorHasCodeNew = new ArrayList<>();

    @JsonProperty("STATICX_xor_UPDATE_xor_VAL_CURR_CHANGES_xor_IS_DEPLOYMENT_xor_EXISTS_NEW")
    private final List<Boolean> staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew =
        new ArrayList<>();

    @JsonProperty("STO_FLAG")
    private final List<Boolean> stoFlag = new ArrayList<>();

    @JsonProperty("SUX")
    private final List<Boolean> sux = new ArrayList<>();

    @JsonProperty("SWAP_FLAG")
    private final List<Boolean> swapFlag = new ArrayList<>();

    @JsonProperty("TRANSACTION_END_STAMP")
    private final List<BigInteger> transactionEndStamp = new ArrayList<>();

    @JsonProperty("TRANSACTION_REVERTS")
    private final List<BigInteger> transactionReverts = new ArrayList<>();

    @JsonProperty("TRM_FLAG_xor_VAL_CURR_IS_ZERO_xor_SUFFICIENT_BALANCE")
    private final List<Boolean> trmFlagXorValCurrIsZeroXorSufficientBalance = new ArrayList<>();

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

    @JsonProperty("WCP_FLAG")
    private final List<Boolean> wcpFlag = new ArrayList<>();

    private TraceBuilder() {}

    int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.abortFlag.size();
    }

    TraceBuilder abortFlag(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("ABORT_FLAG already set");
      } else {
        filled.set(41);
      }

      abortFlag.add(b);

      return this;
    }

    TraceBuilder absoluteTransactionNumber(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ABSOLUTE_TRANSACTION_NUMBER already set");
      } else {
        filled.set(4);
      }

      absoluteTransactionNumber.add(b);

      return this;
    }

    TraceBuilder batchNumber(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("BATCH_NUMBER already set");
      } else {
        filled.set(14);
      }

      batchNumber.add(b);

      return this;
    }

    TraceBuilder callerContextNumber(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("CALLER_CONTEXT_NUMBER already set");
      } else {
        filled.set(23);
      }

      callerContextNumber.add(b);

      return this;
    }

    TraceBuilder codeAddressHi(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("CODE_ADDRESS_HI already set");
      } else {
        filled.set(9);
      }

      codeAddressHi.add(b);

      return this;
    }

    TraceBuilder codeAddressLo(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("CODE_ADDRESS_LO already set");
      } else {
        filled.set(40);
      }

      codeAddressLo.add(b);

      return this;
    }

    TraceBuilder codeDeploymentNumber(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(28);
      }

      codeDeploymentNumber.add(b);

      return this;
    }

    TraceBuilder codeDeploymentStatus(final Boolean b) {
      if (filled.get(26)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_STATUS already set");
      } else {
        filled.set(26);
      }

      codeDeploymentStatus.add(b);

      return this;
    }

    TraceBuilder contextGetsRevrtdFlag(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("CONTEXT_GETS_REVRTD_FLAG already set");
      } else {
        filled.set(13);
      }

      contextGetsRevrtdFlag.add(b);

      return this;
    }

    TraceBuilder contextMayChangeFlag(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("CONTEXT_MAY_CHANGE_FLAG already set");
      } else {
        filled.set(24);
      }

      contextMayChangeFlag.add(b);

      return this;
    }

    TraceBuilder contextNumber(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("CONTEXT_NUMBER already set");
      } else {
        filled.set(20);
      }

      contextNumber.add(b);

      return this;
    }

    TraceBuilder contextNumberNew(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("CONTEXT_NUMBER_NEW already set");
      } else {
        filled.set(3);
      }

      contextNumberNew.add(b);

      return this;
    }

    TraceBuilder contextRevertStamp(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("CONTEXT_REVERT_STAMP already set");
      } else {
        filled.set(30);
      }

      contextRevertStamp.add(b);

      return this;
    }

    TraceBuilder contextSelfRevrtsFlag(final Boolean b) {
      if (filled.get(0)) {
        throw new IllegalStateException("CONTEXT_SELF_REVRTS_FLAG already set");
      } else {
        filled.set(0);
      }

      contextSelfRevrtsFlag.add(b);

      return this;
    }

    TraceBuilder contextWillRevertFlag(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("CONTEXT_WILL_REVERT_FLAG already set");
      } else {
        filled.set(12);
      }

      contextWillRevertFlag.add(b);

      return this;
    }

    TraceBuilder counterNsr(final BigInteger b) {
      if (filled.get(34)) {
        throw new IllegalStateException("COUNTER_NSR already set");
      } else {
        filled.set(34);
      }

      counterNsr.add(b);

      return this;
    }

    TraceBuilder counterTli(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("COUNTER_TLI already set");
      } else {
        filled.set(18);
      }

      counterTli.add(b);

      return this;
    }

    TraceBuilder exceptionAhoyFlag(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("EXCEPTION_AHOY_FLAG already set");
      } else {
        filled.set(36);
      }

      exceptionAhoyFlag.add(b);

      return this;
    }

    TraceBuilder failureConditionFlag(final Boolean b) {
      if (filled.get(37)) {
        throw new IllegalStateException("FAILURE_CONDITION_FLAG already set");
      } else {
        filled.set(37);
      }

      failureConditionFlag.add(b);

      return this;
    }

    TraceBuilder gasActual(final BigInteger b) {
      if (filled.get(31)) {
        throw new IllegalStateException("GAS_ACTUAL already set");
      } else {
        filled.set(31);
      }

      gasActual.add(b);

      return this;
    }

    TraceBuilder gasCost(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("GAS_COST already set");
      } else {
        filled.set(22);
      }

      gasCost.add(b);

      return this;
    }

    TraceBuilder gasExpected(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("GAS_EXPECTED already set");
      } else {
        filled.set(15);
      }

      gasExpected.add(b);

      return this;
    }

    TraceBuilder gasMemoryExpansion(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("GAS_MEMORY_EXPANSION already set");
      } else {
        filled.set(8);
      }

      gasMemoryExpansion.add(b);

      return this;
    }

    TraceBuilder gasNext(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("GAS_NEXT already set");
      } else {
        filled.set(1);
      }

      gasNext.add(b);

      return this;
    }

    TraceBuilder gasRefund(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("GAS_REFUND already set");
      } else {
        filled.set(21);
      }

      gasRefund.add(b);

      return this;
    }

    TraceBuilder hubStamp(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("HUB_STAMP already set");
      } else {
        filled.set(2);
      }

      hubStamp.add(b);

      return this;
    }

    TraceBuilder numberOfNonStackRows(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("NUMBER_OF_NON_STACK_ROWS already set");
      } else {
        filled.set(29);
      }

      numberOfNonStackRows.add(b);

      return this;
    }

    TraceBuilder pAccountAddressHi(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("ADDRESS_HI already set");
      } else {
        filled.set(54);
      }

      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.add(b);

      return this;
    }

    TraceBuilder pAccountAddressLo(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("ADDRESS_LO already set");
      } else {
        filled.set(43);
      }

      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.add(b);

      return this;
    }

    TraceBuilder pAccountBalance(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("BALANCE already set");
      } else {
        filled.set(52);
      }

      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.add(b);

      return this;
    }

    TraceBuilder pAccountBalanceNew(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("BALANCE_NEW already set");
      } else {
        filled.set(45);
      }

      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.add(b);

      return this;
    }

    TraceBuilder pAccountCodeHashHi(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("CODE_HASH_HI already set");
      } else {
        filled.set(53);
      }

      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.add(b);

      return this;
    }

    TraceBuilder pAccountCodeHashHiNew(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("CODE_HASH_HI_NEW already set");
      } else {
        filled.set(49);
      }

      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .add(b);

      return this;
    }

    TraceBuilder pAccountCodeHashLo(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("CODE_HASH_LO already set");
      } else {
        filled.set(48);
      }

      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.add(b);

      return this;
    }

    TraceBuilder pAccountCodeHashLoNew(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("CODE_HASH_LO_NEW already set");
      } else {
        filled.set(55);
      }

      codeHashLoNewXorCallValueXorStackItemStamp1.add(b);

      return this;
    }

    TraceBuilder pAccountCodeSize(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("CODE_SIZE already set");
      } else {
        filled.set(57);
      }

      codeSizeXorCallDataOffsetXorStackItemStamp2.add(b);

      return this;
    }

    TraceBuilder pAccountCodeSizeNew(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("CODE_SIZE_NEW already set");
      } else {
        filled.set(42);
      }

      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.add(b);

      return this;
    }

    TraceBuilder pAccountDeploymentNumber(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(46);
      }

      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.add(b);

      return this;
    }

    TraceBuilder pAccountDeploymentNumberInfty(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("DEPLOYMENT_NUMBER_INFTY already set");
      } else {
        filled.set(47);
      }

      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .add(b);

      return this;
    }

    TraceBuilder pAccountDeploymentNumberNew(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("DEPLOYMENT_NUMBER_NEW already set");
      } else {
        filled.set(51);
      }

      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.add(
          b);

      return this;
    }

    TraceBuilder pAccountDeploymentStatus(final BigInteger b) {
      if (filled.get(58)) {
        throw new IllegalStateException("DEPLOYMENT_STATUS already set");
      } else {
        filled.set(58);
      }

      deploymentStatusXorReturnAtSizeXorHeight.add(b);

      return this;
    }

    TraceBuilder pAccountDeploymentStatusInfty(final Boolean b) {
      if (filled.get(70)) {
        throw new IllegalStateException("DEPLOYMENT_STATUS_INFTY already set");
      } else {
        filled.set(70);
      }

      maxcsxXorWarmNewXorDeploymentStatusInfty.add(b);

      return this;
    }

    TraceBuilder pAccountDeploymentStatusNew(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("DEPLOYMENT_STATUS_NEW already set");
      } else {
        filled.set(56);
      }

      deploymentStatusNewXorIsStaticXorInstruction.add(b);

      return this;
    }

    TraceBuilder pAccountExists(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("EXISTS already set");
      } else {
        filled.set(67);
      }

      oogxXorValNextIsZeroXorExists.add(b);

      return this;
    }

    TraceBuilder pAccountExistsNew(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("EXISTS_NEW already set");
      } else {
        filled.set(66);
      }

      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.add(b);

      return this;
    }

    TraceBuilder pAccountHasCode(final Boolean b) {
      if (filled.get(69)) {
        throw new IllegalStateException("HAS_CODE already set");
      } else {
        filled.set(69);
      }

      decodedFlag2XorValOrigIsZeroXorHasCode.add(b);

      return this;
    }

    TraceBuilder pAccountHasCodeNew(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("HAS_CODE_NEW already set");
      } else {
        filled.set(68);
      }

      staticFlagXorWarmXorHasCodeNew.add(b);

      return this;
    }

    TraceBuilder pAccountIsPrecompile(final Boolean b) {
      if (filled.get(73)) {
        throw new IllegalStateException("IS_PRECOMPILE already set");
      } else {
        filled.set(73);
      }

      dupFlagXorValNextIsOrigXorIsPrecompile.add(b);

      return this;
    }

    TraceBuilder pAccountNonce(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("NONCE already set");
      } else {
        filled.set(44);
      }

      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.add(b);

      return this;
    }

    TraceBuilder pAccountNonceNew(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("NONCE_NEW already set");
      } else {
        filled.set(50);
      }

      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.add(b);

      return this;
    }

    TraceBuilder pAccountSufficientBalance(final Boolean b) {
      if (filled.get(74)) {
        throw new IllegalStateException("SUFFICIENT_BALANCE already set");
      } else {
        filled.set(74);
      }

      trmFlagXorValCurrIsZeroXorSufficientBalance.add(b);

      return this;
    }

    TraceBuilder pAccountWarm(final Boolean b) {
      if (filled.get(71)) {
        throw new IllegalStateException("WARM already set");
      } else {
        filled.set(71);
      }

      accFlagXorValNextIsCurrXorWarm.add(b);

      return this;
    }

    TraceBuilder pAccountWarmNew(final Boolean b) {
      if (filled.get(72)) {
        throw new IllegalStateException("WARM_NEW already set");
      } else {
        filled.set(72);
      }

      pushpopFlagXorValCurrIsOrigXorWarmNew.add(b);

      return this;
    }

    TraceBuilder pContextAccountAddressHi(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("ACCOUNT_ADDRESS_HI already set");
      } else {
        filled.set(53);
      }

      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.add(b);

      return this;
    }

    TraceBuilder pContextAccountAddressLo(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("ACCOUNT_ADDRESS_LO already set");
      } else {
        filled.set(51);
      }

      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.add(
          b);

      return this;
    }

    TraceBuilder pContextAccountDeploymentNumber(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("ACCOUNT_DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(46);
      }

      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.add(b);

      return this;
    }

    TraceBuilder pContextByteCodeAddressHi(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("BYTE_CODE_ADDRESS_HI already set");
      } else {
        filled.set(52);
      }

      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.add(b);

      return this;
    }

    TraceBuilder pContextByteCodeAddressLo(final BigInteger b) {
      if (filled.get(63)) {
        throw new IllegalStateException("BYTE_CODE_ADDRESS_LO already set");
      } else {
        filled.set(63);
      }

      byteCodeAddressLoXorStackItemStamp4.add(b);

      return this;
    }

    TraceBuilder pContextByteCodeDeploymentNumber(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("BYTE_CODE_DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(54);
      }

      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.add(b);

      return this;
    }

    TraceBuilder pContextByteCodeDeploymentStatus(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("BYTE_CODE_DEPLOYMENT_STATUS already set");
      } else {
        filled.set(47);
      }

      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .add(b);

      return this;
    }

    TraceBuilder pContextCallDataOffset(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("CALL_DATA_OFFSET already set");
      } else {
        filled.set(57);
      }

      codeSizeXorCallDataOffsetXorStackItemStamp2.add(b);

      return this;
    }

    TraceBuilder pContextCallDataSize(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("CALL_DATA_SIZE already set");
      } else {
        filled.set(44);
      }

      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.add(b);

      return this;
    }

    TraceBuilder pContextCallStackDepth(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("CALL_STACK_DEPTH already set");
      } else {
        filled.set(45);
      }

      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.add(b);

      return this;
    }

    TraceBuilder pContextCallValue(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("CALL_VALUE already set");
      } else {
        filled.set(55);
      }

      codeHashLoNewXorCallValueXorStackItemStamp1.add(b);

      return this;
    }

    TraceBuilder pContextCallerAddressHi(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("CALLER_ADDRESS_HI already set");
      } else {
        filled.set(43);
      }

      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.add(b);

      return this;
    }

    TraceBuilder pContextCallerAddressLo(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("CALLER_ADDRESS_LO already set");
      } else {
        filled.set(49);
      }

      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .add(b);

      return this;
    }

    TraceBuilder pContextCallerContextNumber(final BigInteger b) {
      if (filled.get(61)) {
        throw new IllegalStateException("CALLER_CONTEXT_NUMBER already set");
      } else {
        filled.set(61);
      }

      callerContextNumberXorPushValueLo.add(b);

      return this;
    }

    TraceBuilder pContextContextNumber(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("CONTEXT_NUMBER already set");
      } else {
        filled.set(50);
      }

      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.add(b);

      return this;
    }

    TraceBuilder pContextIsStatic(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("IS_STATIC already set");
      } else {
        filled.set(56);
      }

      deploymentStatusNewXorIsStaticXorInstruction.add(b);

      return this;
    }

    TraceBuilder pContextReturnAtOffset(final BigInteger b) {
      if (filled.get(59)) {
        throw new IllegalStateException("RETURN_AT_OFFSET already set");
      } else {
        filled.set(59);
      }

      returnAtOffsetXorPushValueHi.add(b);

      return this;
    }

    TraceBuilder pContextReturnAtSize(final BigInteger b) {
      if (filled.get(58)) {
        throw new IllegalStateException("RETURN_AT_SIZE already set");
      } else {
        filled.set(58);
      }

      deploymentStatusXorReturnAtSizeXorHeight.add(b);

      return this;
    }

    TraceBuilder pContextReturnDataOffset(final BigInteger b) {
      if (filled.get(62)) {
        throw new IllegalStateException("RETURN_DATA_OFFSET already set");
      } else {
        filled.set(62);
      }

      returnDataOffsetXorStackItemStamp3.add(b);

      return this;
    }

    TraceBuilder pContextReturnDataSize(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("RETURN_DATA_SIZE already set");
      } else {
        filled.set(48);
      }

      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.add(b);

      return this;
    }

    TraceBuilder pContextReturnerContextNumber(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("RETURNER_CONTEXT_NUMBER already set");
      } else {
        filled.set(42);
      }

      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.add(b);

      return this;
    }

    TraceBuilder pContextReturnerIsPrecompile(final BigInteger b) {
      if (filled.get(60)) {
        throw new IllegalStateException("RETURNER_IS_PRECOMPILE already set");
      } else {
        filled.set(60);
      }

      returnerIsPrecompileXorStackItemHeight4.add(b);

      return this;
    }

    TraceBuilder pContextUpdate(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("UPDATE already set");
      } else {
        filled.set(66);
      }

      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.add(b);

      return this;
    }

    TraceBuilder pStackAccFlag(final Boolean b) {
      if (filled.get(71)) {
        throw new IllegalStateException("ACC_FLAG already set");
      } else {
        filled.set(71);
      }

      accFlagXorValNextIsCurrXorWarm.add(b);

      return this;
    }

    TraceBuilder pStackAddFlag(final Boolean b) {
      if (filled.get(88)) {
        throw new IllegalStateException("ADD_FLAG already set");
      } else {
        filled.set(88);
      }

      addFlag.add(b);

      return this;
    }

    TraceBuilder pStackBinFlag(final Boolean b) {
      if (filled.get(91)) {
        throw new IllegalStateException("BIN_FLAG already set");
      } else {
        filled.set(91);
      }

      binFlag.add(b);

      return this;
    }

    TraceBuilder pStackBtcFlag(final Boolean b) {
      if (filled.get(94)) {
        throw new IllegalStateException("BTC_FLAG already set");
      } else {
        filled.set(94);
      }

      btcFlag.add(b);

      return this;
    }

    TraceBuilder pStackCallFlag(final Boolean b) {
      if (filled.get(90)) {
        throw new IllegalStateException("CALL_FLAG already set");
      } else {
        filled.set(90);
      }

      callFlag.add(b);

      return this;
    }

    TraceBuilder pStackConFlag(final Boolean b) {
      if (filled.get(112)) {
        throw new IllegalStateException("CON_FLAG already set");
      } else {
        filled.set(112);
      }

      conFlag.add(b);

      return this;
    }

    TraceBuilder pStackCopyFlag(final Boolean b) {
      if (filled.get(92)) {
        throw new IllegalStateException("COPY_FLAG already set");
      } else {
        filled.set(92);
      }

      copyFlag.add(b);

      return this;
    }

    TraceBuilder pStackCreateFlag(final Boolean b) {
      if (filled.get(105)) {
        throw new IllegalStateException("CREATE_FLAG already set");
      } else {
        filled.set(105);
      }

      createFlag.add(b);

      return this;
    }

    TraceBuilder pStackDecodedFlag1(final Boolean b) {
      if (filled.get(81)) {
        throw new IllegalStateException("DECODED_FLAG_1 already set");
      } else {
        filled.set(81);
      }

      decodedFlag1.add(b);

      return this;
    }

    TraceBuilder pStackDecodedFlag2(final Boolean b) {
      if (filled.get(69)) {
        throw new IllegalStateException("DECODED_FLAG_2 already set");
      } else {
        filled.set(69);
      }

      decodedFlag2XorValOrigIsZeroXorHasCode.add(b);

      return this;
    }

    TraceBuilder pStackDecodedFlag3(final Boolean b) {
      if (filled.get(103)) {
        throw new IllegalStateException("DECODED_FLAG_3 already set");
      } else {
        filled.set(103);
      }

      decodedFlag3.add(b);

      return this;
    }

    TraceBuilder pStackDecodedFlag4(final Boolean b) {
      if (filled.get(108)) {
        throw new IllegalStateException("DECODED_FLAG_4 already set");
      } else {
        filled.set(108);
      }

      decodedFlag4.add(b);

      return this;
    }

    TraceBuilder pStackDupFlag(final Boolean b) {
      if (filled.get(73)) {
        throw new IllegalStateException("DUP_FLAG already set");
      } else {
        filled.set(73);
      }

      dupFlagXorValNextIsOrigXorIsPrecompile.add(b);

      return this;
    }

    TraceBuilder pStackExtFlag(final Boolean b) {
      if (filled.get(86)) {
        throw new IllegalStateException("EXT_FLAG already set");
      } else {
        filled.set(86);
      }

      extFlag.add(b);

      return this;
    }

    TraceBuilder pStackHaltFlag(final Boolean b) {
      if (filled.get(99)) {
        throw new IllegalStateException("HALT_FLAG already set");
      } else {
        filled.set(99);
      }

      haltFlag.add(b);

      return this;
    }

    TraceBuilder pStackHeight(final BigInteger b) {
      if (filled.get(58)) {
        throw new IllegalStateException("HEIGHT already set");
      } else {
        filled.set(58);
      }

      deploymentStatusXorReturnAtSizeXorHeight.add(b);

      return this;
    }

    TraceBuilder pStackHeightNew(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("HEIGHT_NEW already set");
      } else {
        filled.set(44);
      }

      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.add(b);

      return this;
    }

    TraceBuilder pStackHeightOver(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("HEIGHT_OVER already set");
      } else {
        filled.set(46);
      }

      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.add(b);

      return this;
    }

    TraceBuilder pStackHeightUnder(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("HEIGHT_UNDER already set");
      } else {
        filled.set(48);
      }

      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.add(b);

      return this;
    }

    TraceBuilder pStackInstruction(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("INSTRUCTION already set");
      } else {
        filled.set(56);
      }

      deploymentStatusNewXorIsStaticXorInstruction.add(b);

      return this;
    }

    TraceBuilder pStackInvalidFlag(final Boolean b) {
      if (filled.get(93)) {
        throw new IllegalStateException("INVALID_FLAG already set");
      } else {
        filled.set(93);
      }

      invalidFlag.add(b);

      return this;
    }

    TraceBuilder pStackInvprex(final Boolean b) {
      if (filled.get(109)) {
        throw new IllegalStateException("INVPREX already set");
      } else {
        filled.set(109);
      }

      invprex.add(b);

      return this;
    }

    TraceBuilder pStackJumpFlag(final Boolean b) {
      if (filled.get(95)) {
        throw new IllegalStateException("JUMP_FLAG already set");
      } else {
        filled.set(95);
      }

      jumpFlag.add(b);

      return this;
    }

    TraceBuilder pStackJumpx(final Boolean b) {
      if (filled.get(80)) {
        throw new IllegalStateException("JUMPX already set");
      } else {
        filled.set(80);
      }

      jumpx.add(b);

      return this;
    }

    TraceBuilder pStackKecFlag(final Boolean b) {
      if (filled.get(75)) {
        throw new IllegalStateException("KEC_FLAG already set");
      } else {
        filled.set(75);
      }

      kecFlag.add(b);

      return this;
    }

    TraceBuilder pStackLogFlag(final Boolean b) {
      if (filled.get(104)) {
        throw new IllegalStateException("LOG_FLAG already set");
      } else {
        filled.set(104);
      }

      logFlag.add(b);

      return this;
    }

    TraceBuilder pStackMaxcsx(final Boolean b) {
      if (filled.get(70)) {
        throw new IllegalStateException("MAXCSX already set");
      } else {
        filled.set(70);
      }

      maxcsxXorWarmNewXorDeploymentStatusInfty.add(b);

      return this;
    }

    TraceBuilder pStackModFlag(final Boolean b) {
      if (filled.get(111)) {
        throw new IllegalStateException("MOD_FLAG already set");
      } else {
        filled.set(111);
      }

      modFlag.add(b);

      return this;
    }

    TraceBuilder pStackMulFlag(final Boolean b) {
      if (filled.get(84)) {
        throw new IllegalStateException("MUL_FLAG already set");
      } else {
        filled.set(84);
      }

      mulFlag.add(b);

      return this;
    }

    TraceBuilder pStackMxpFlag(final Boolean b) {
      if (filled.get(107)) {
        throw new IllegalStateException("MXP_FLAG already set");
      } else {
        filled.set(107);
      }

      mxpFlag.add(b);

      return this;
    }

    TraceBuilder pStackMxpx(final Boolean b) {
      if (filled.get(79)) {
        throw new IllegalStateException("MXPX already set");
      } else {
        filled.set(79);
      }

      mxpx.add(b);

      return this;
    }

    TraceBuilder pStackOobFlag(final Boolean b) {
      if (filled.get(76)) {
        throw new IllegalStateException("OOB_FLAG already set");
      } else {
        filled.set(76);
      }

      oobFlag.add(b);

      return this;
    }

    TraceBuilder pStackOogx(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("OOGX already set");
      } else {
        filled.set(67);
      }

      oogxXorValNextIsZeroXorExists.add(b);

      return this;
    }

    TraceBuilder pStackOpcx(final Boolean b) {
      if (filled.get(82)) {
        throw new IllegalStateException("OPCX already set");
      } else {
        filled.set(82);
      }

      opcx.add(b);

      return this;
    }

    TraceBuilder pStackPushValueHi(final BigInteger b) {
      if (filled.get(59)) {
        throw new IllegalStateException("PUSH_VALUE_HI already set");
      } else {
        filled.set(59);
      }

      returnAtOffsetXorPushValueHi.add(b);

      return this;
    }

    TraceBuilder pStackPushValueLo(final BigInteger b) {
      if (filled.get(61)) {
        throw new IllegalStateException("PUSH_VALUE_LO already set");
      } else {
        filled.set(61);
      }

      callerContextNumberXorPushValueLo.add(b);

      return this;
    }

    TraceBuilder pStackPushpopFlag(final Boolean b) {
      if (filled.get(72)) {
        throw new IllegalStateException("PUSHPOP_FLAG already set");
      } else {
        filled.set(72);
      }

      pushpopFlagXorValCurrIsOrigXorWarmNew.add(b);

      return this;
    }

    TraceBuilder pStackRdcx(final Boolean b) {
      if (filled.get(77)) {
        throw new IllegalStateException("RDCX already set");
      } else {
        filled.set(77);
      }

      rdcx.add(b);

      return this;
    }

    TraceBuilder pStackShfFlag(final Boolean b) {
      if (filled.get(98)) {
        throw new IllegalStateException("SHF_FLAG already set");
      } else {
        filled.set(98);
      }

      shfFlag.add(b);

      return this;
    }

    TraceBuilder pStackSox(final Boolean b) {
      if (filled.get(78)) {
        throw new IllegalStateException("SOX already set");
      } else {
        filled.set(78);
      }

      sox.add(b);

      return this;
    }

    TraceBuilder pStackSstorex(final Boolean b) {
      if (filled.get(96)) {
        throw new IllegalStateException("SSTOREX already set");
      } else {
        filled.set(96);
      }

      sstorex.add(b);

      return this;
    }

    TraceBuilder pStackStackItemHeight1(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_1 already set");
      } else {
        filled.set(43);
      }

      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.add(b);

      return this;
    }

    TraceBuilder pStackStackItemHeight2(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_2 already set");
      } else {
        filled.set(50);
      }

      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.add(b);

      return this;
    }

    TraceBuilder pStackStackItemHeight3(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_3 already set");
      } else {
        filled.set(53);
      }

      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.add(b);

      return this;
    }

    TraceBuilder pStackStackItemHeight4(final BigInteger b) {
      if (filled.get(60)) {
        throw new IllegalStateException("STACK_ITEM_HEIGHT_4 already set");
      } else {
        filled.set(60);
      }

      returnerIsPrecompileXorStackItemHeight4.add(b);

      return this;
    }

    TraceBuilder pStackStackItemPop1(final Boolean b) {
      if (filled.get(100)) {
        throw new IllegalStateException("STACK_ITEM_POP_1 already set");
      } else {
        filled.set(100);
      }

      stackItemPop1.add(b);

      return this;
    }

    TraceBuilder pStackStackItemPop2(final Boolean b) {
      if (filled.get(110)) {
        throw new IllegalStateException("STACK_ITEM_POP_2 already set");
      } else {
        filled.set(110);
      }

      stackItemPop2.add(b);

      return this;
    }

    TraceBuilder pStackStackItemPop3(final Boolean b) {
      if (filled.get(106)) {
        throw new IllegalStateException("STACK_ITEM_POP_3 already set");
      } else {
        filled.set(106);
      }

      stackItemPop3.add(b);

      return this;
    }

    TraceBuilder pStackStackItemPop4(final Boolean b) {
      if (filled.get(97)) {
        throw new IllegalStateException("STACK_ITEM_POP_4 already set");
      } else {
        filled.set(97);
      }

      stackItemPop4.add(b);

      return this;
    }

    TraceBuilder pStackStackItemStamp1(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_1 already set");
      } else {
        filled.set(55);
      }

      codeHashLoNewXorCallValueXorStackItemStamp1.add(b);

      return this;
    }

    TraceBuilder pStackStackItemStamp2(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_2 already set");
      } else {
        filled.set(57);
      }

      codeSizeXorCallDataOffsetXorStackItemStamp2.add(b);

      return this;
    }

    TraceBuilder pStackStackItemStamp3(final BigInteger b) {
      if (filled.get(62)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_3 already set");
      } else {
        filled.set(62);
      }

      returnDataOffsetXorStackItemStamp3.add(b);

      return this;
    }

    TraceBuilder pStackStackItemStamp4(final BigInteger b) {
      if (filled.get(63)) {
        throw new IllegalStateException("STACK_ITEM_STAMP_4 already set");
      } else {
        filled.set(63);
      }

      byteCodeAddressLoXorStackItemStamp4.add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueHi1(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_1 already set");
      } else {
        filled.set(51);
      }

      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.add(
          b);

      return this;
    }

    TraceBuilder pStackStackItemValueHi2(final BigInteger b) {
      if (filled.get(65)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_2 already set");
      } else {
        filled.set(65);
      }

      stackItemValueHi2.add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueHi3(final BigInteger b) {
      if (filled.get(64)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_3 already set");
      } else {
        filled.set(64);
      }

      stackItemValueHi3.add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueHi4(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_4 already set");
      } else {
        filled.set(47);
      }

      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueLo1(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_1 already set");
      } else {
        filled.set(54);
      }

      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueLo2(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_2 already set");
      } else {
        filled.set(45);
      }

      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueLo3(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_3 already set");
      } else {
        filled.set(49);
      }

      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .add(b);

      return this;
    }

    TraceBuilder pStackStackItemValueLo4(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_LO_4 already set");
      } else {
        filled.set(52);
      }

      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.add(b);

      return this;
    }

    TraceBuilder pStackStackramFlag(final Boolean b) {
      if (filled.get(87)) {
        throw new IllegalStateException("STACKRAM_FLAG already set");
      } else {
        filled.set(87);
      }

      stackramFlag.add(b);

      return this;
    }

    TraceBuilder pStackStaticFlag(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("STATIC_FLAG already set");
      } else {
        filled.set(68);
      }

      staticFlagXorWarmXorHasCodeNew.add(b);

      return this;
    }

    TraceBuilder pStackStaticGas(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("STATIC_GAS already set");
      } else {
        filled.set(42);
      }

      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.add(b);

      return this;
    }

    TraceBuilder pStackStaticx(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("STATICX already set");
      } else {
        filled.set(66);
      }

      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.add(b);

      return this;
    }

    TraceBuilder pStackStoFlag(final Boolean b) {
      if (filled.get(85)) {
        throw new IllegalStateException("STO_FLAG already set");
      } else {
        filled.set(85);
      }

      stoFlag.add(b);

      return this;
    }

    TraceBuilder pStackSux(final Boolean b) {
      if (filled.get(83)) {
        throw new IllegalStateException("SUX already set");
      } else {
        filled.set(83);
      }

      sux.add(b);

      return this;
    }

    TraceBuilder pStackSwapFlag(final Boolean b) {
      if (filled.get(89)) {
        throw new IllegalStateException("SWAP_FLAG already set");
      } else {
        filled.set(89);
      }

      swapFlag.add(b);

      return this;
    }

    TraceBuilder pStackTrmFlag(final Boolean b) {
      if (filled.get(74)) {
        throw new IllegalStateException("TRM_FLAG already set");
      } else {
        filled.set(74);
      }

      trmFlagXorValCurrIsZeroXorSufficientBalance.add(b);

      return this;
    }

    TraceBuilder pStackTxnFlag(final Boolean b) {
      if (filled.get(101)) {
        throw new IllegalStateException("TXN_FLAG already set");
      } else {
        filled.set(101);
      }

      txnFlag.add(b);

      return this;
    }

    TraceBuilder pStackWcpFlag(final Boolean b) {
      if (filled.get(102)) {
        throw new IllegalStateException("WCP_FLAG already set");
      } else {
        filled.set(102);
      }

      wcpFlag.add(b);

      return this;
    }

    TraceBuilder pStorageAddressHi(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("ADDRESS_HI already set");
      } else {
        filled.set(46);
      }

      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.add(b);

      return this;
    }

    TraceBuilder pStorageAddressLo(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("ADDRESS_LO already set");
      } else {
        filled.set(49);
      }

      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .add(b);

      return this;
    }

    TraceBuilder pStorageDeploymentNumber(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("DEPLOYMENT_NUMBER already set");
      } else {
        filled.set(47);
      }

      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .add(b);

      return this;
    }

    TraceBuilder pStorageStorageKeyHi(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("STORAGE_KEY_HI already set");
      } else {
        filled.set(52);
      }

      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.add(b);

      return this;
    }

    TraceBuilder pStorageStorageKeyLo(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("STORAGE_KEY_LO already set");
      } else {
        filled.set(51);
      }

      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.add(
          b);

      return this;
    }

    TraceBuilder pStorageValCurrChanges(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("VAL_CURR_CHANGES already set");
      } else {
        filled.set(66);
      }

      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.add(b);

      return this;
    }

    TraceBuilder pStorageValCurrHi(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("VAL_CURR_HI already set");
      } else {
        filled.set(45);
      }

      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.add(b);

      return this;
    }

    TraceBuilder pStorageValCurrIsOrig(final Boolean b) {
      if (filled.get(72)) {
        throw new IllegalStateException("VAL_CURR_IS_ORIG already set");
      } else {
        filled.set(72);
      }

      pushpopFlagXorValCurrIsOrigXorWarmNew.add(b);

      return this;
    }

    TraceBuilder pStorageValCurrIsZero(final Boolean b) {
      if (filled.get(74)) {
        throw new IllegalStateException("VAL_CURR_IS_ZERO already set");
      } else {
        filled.set(74);
      }

      trmFlagXorValCurrIsZeroXorSufficientBalance.add(b);

      return this;
    }

    TraceBuilder pStorageValCurrLo(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("VAL_CURR_LO already set");
      } else {
        filled.set(43);
      }

      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.add(b);

      return this;
    }

    TraceBuilder pStorageValNextHi(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("VAL_NEXT_HI already set");
      } else {
        filled.set(50);
      }

      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.add(b);

      return this;
    }

    TraceBuilder pStorageValNextIsCurr(final Boolean b) {
      if (filled.get(71)) {
        throw new IllegalStateException("VAL_NEXT_IS_CURR already set");
      } else {
        filled.set(71);
      }

      accFlagXorValNextIsCurrXorWarm.add(b);

      return this;
    }

    TraceBuilder pStorageValNextIsOrig(final Boolean b) {
      if (filled.get(73)) {
        throw new IllegalStateException("VAL_NEXT_IS_ORIG already set");
      } else {
        filled.set(73);
      }

      dupFlagXorValNextIsOrigXorIsPrecompile.add(b);

      return this;
    }

    TraceBuilder pStorageValNextIsZero(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("VAL_NEXT_IS_ZERO already set");
      } else {
        filled.set(67);
      }

      oogxXorValNextIsZeroXorExists.add(b);

      return this;
    }

    TraceBuilder pStorageValNextLo(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("VAL_NEXT_LO already set");
      } else {
        filled.set(42);
      }

      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.add(b);

      return this;
    }

    TraceBuilder pStorageValOrigHi(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("VAL_ORIG_HI already set");
      } else {
        filled.set(48);
      }

      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.add(b);

      return this;
    }

    TraceBuilder pStorageValOrigIsZero(final Boolean b) {
      if (filled.get(69)) {
        throw new IllegalStateException("VAL_ORIG_IS_ZERO already set");
      } else {
        filled.set(69);
      }

      decodedFlag2XorValOrigIsZeroXorHasCode.add(b);

      return this;
    }

    TraceBuilder pStorageValOrigLo(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("VAL_ORIG_LO already set");
      } else {
        filled.set(44);
      }

      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.add(b);

      return this;
    }

    TraceBuilder pStorageWarm(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("WARM already set");
      } else {
        filled.set(68);
      }

      staticFlagXorWarmXorHasCodeNew.add(b);

      return this;
    }

    TraceBuilder pStorageWarmNew(final Boolean b) {
      if (filled.get(70)) {
        throw new IllegalStateException("WARM_NEW already set");
      } else {
        filled.set(70);
      }

      maxcsxXorWarmNewXorDeploymentStatusInfty.add(b);

      return this;
    }

    TraceBuilder pTransactionAbsoluteTransactionNumber(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("ABSOLUTE_TRANSACTION_NUMBER already set");
      } else {
        filled.set(49);
      }

      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .add(b);

      return this;
    }

    TraceBuilder pTransactionBatchNumber(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("BATCH_NUMBER already set");
      } else {
        filled.set(48);
      }

      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.add(b);

      return this;
    }

    TraceBuilder pTransactionFromAddressHi(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("FROM_ADDRESS_HI already set");
      } else {
        filled.set(43);
      }

      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.add(b);

      return this;
    }

    TraceBuilder pTransactionFromAddressLo(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("FROM_ADDRESS_LO already set");
      } else {
        filled.set(47);
      }

      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .add(b);

      return this;
    }

    TraceBuilder pTransactionGasFee(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("GAS_FEE already set");
      } else {
        filled.set(52);
      }

      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.add(b);

      return this;
    }

    TraceBuilder pTransactionGasMaxfee(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("GAS_MAXFEE already set");
      } else {
        filled.set(44);
      }

      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.add(b);

      return this;
    }

    TraceBuilder pTransactionGasTip(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("GAS_TIP already set");
      } else {
        filled.set(45);
      }

      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.add(b);

      return this;
    }

    TraceBuilder pTransactionInitGas(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("INIT_GAS already set");
      } else {
        filled.set(46);
      }

      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.add(b);

      return this;
    }

    TraceBuilder pTransactionIsDeployment(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("IS_DEPLOYMENT already set");
      } else {
        filled.set(66);
      }

      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.add(b);

      return this;
    }

    TraceBuilder pTransactionNonce(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("NONCE already set");
      } else {
        filled.set(53);
      }

      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.add(b);

      return this;
    }

    TraceBuilder pTransactionToAddressHi(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("TO_ADDRESS_HI already set");
      } else {
        filled.set(51);
      }

      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.add(
          b);

      return this;
    }

    TraceBuilder pTransactionToAddressLo(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("TO_ADDRESS_LO already set");
      } else {
        filled.set(50);
      }

      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.add(b);

      return this;
    }

    TraceBuilder pTransactionValue(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("VALUE already set");
      } else {
        filled.set(42);
      }

      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.add(b);

      return this;
    }

    TraceBuilder peekAtAccount(final Boolean b) {
      if (filled.get(32)) {
        throw new IllegalStateException("PEEK_AT_ACCOUNT already set");
      } else {
        filled.set(32);
      }

      peekAtAccount.add(b);

      return this;
    }

    TraceBuilder peekAtContext(final Boolean b) {
      if (filled.get(39)) {
        throw new IllegalStateException("PEEK_AT_CONTEXT already set");
      } else {
        filled.set(39);
      }

      peekAtContext.add(b);

      return this;
    }

    TraceBuilder peekAtStack(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("PEEK_AT_STACK already set");
      } else {
        filled.set(35);
      }

      peekAtStack.add(b);

      return this;
    }

    TraceBuilder peekAtStorage(final Boolean b) {
      if (filled.get(7)) {
        throw new IllegalStateException("PEEK_AT_STORAGE already set");
      } else {
        filled.set(7);
      }

      peekAtStorage.add(b);

      return this;
    }

    TraceBuilder peekAtTransaction(final Boolean b) {
      if (filled.get(17)) {
        throw new IllegalStateException("PEEK_AT_TRANSACTION already set");
      } else {
        filled.set(17);
      }

      peekAtTransaction.add(b);

      return this;
    }

    TraceBuilder programCounter(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("PROGRAM_COUNTER already set");
      } else {
        filled.set(38);
      }

      programCounter.add(b);

      return this;
    }

    TraceBuilder programCounterNew(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("PROGRAM_COUNTER_NEW already set");
      } else {
        filled.set(10);
      }

      programCounterNew.add(b);

      return this;
    }

    TraceBuilder transactionEndStamp(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("TRANSACTION_END_STAMP already set");
      } else {
        filled.set(11);
      }

      transactionEndStamp.add(b);

      return this;
    }

    TraceBuilder transactionReverts(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("TRANSACTION_REVERTS already set");
      } else {
        filled.set(33);
      }

      transactionReverts.add(b);

      return this;
    }

    TraceBuilder twoLineInstruction(final Boolean b) {
      if (filled.get(27)) {
        throw new IllegalStateException("TWO_LINE_INSTRUCTION already set");
      } else {
        filled.set(27);
      }

      twoLineInstruction.add(b);

      return this;
    }

    TraceBuilder txExec(final Boolean b) {
      if (filled.get(6)) {
        throw new IllegalStateException("TX_EXEC already set");
      } else {
        filled.set(6);
      }

      txExec.add(b);

      return this;
    }

    TraceBuilder txFinl(final Boolean b) {
      if (filled.get(5)) {
        throw new IllegalStateException("TX_FINL already set");
      } else {
        filled.set(5);
      }

      txFinl.add(b);

      return this;
    }

    TraceBuilder txInit(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("TX_INIT already set");
      } else {
        filled.set(16);
      }

      txInit.add(b);

      return this;
    }

    TraceBuilder txSkip(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("TX_SKIP already set");
      } else {
        filled.set(19);
      }

      txSkip.add(b);

      return this;
    }

    TraceBuilder txWarm(final Boolean b) {
      if (filled.get(25)) {
        throw new IllegalStateException("TX_WARM already set");
      } else {
        filled.set(25);
      }

      txWarm.add(b);

      return this;
    }

    TraceBuilder setAbortFlagAt(final Boolean b, int i) {
      abortFlag.set(i, b);

      return this;
    }

    TraceBuilder setAbsoluteTransactionNumberAt(final BigInteger b, int i) {
      absoluteTransactionNumber.set(i, b);

      return this;
    }

    TraceBuilder setBatchNumberAt(final BigInteger b, int i) {
      batchNumber.set(i, b);

      return this;
    }

    TraceBuilder setCallerContextNumberAt(final BigInteger b, int i) {
      callerContextNumber.set(i, b);

      return this;
    }

    TraceBuilder setCodeAddressHiAt(final BigInteger b, int i) {
      codeAddressHi.set(i, b);

      return this;
    }

    TraceBuilder setCodeAddressLoAt(final BigInteger b, int i) {
      codeAddressLo.set(i, b);

      return this;
    }

    TraceBuilder setCodeDeploymentNumberAt(final BigInteger b, int i) {
      codeDeploymentNumber.set(i, b);

      return this;
    }

    TraceBuilder setCodeDeploymentStatusAt(final Boolean b, int i) {
      codeDeploymentStatus.set(i, b);

      return this;
    }

    TraceBuilder setContextGetsRevrtdFlagAt(final Boolean b, int i) {
      contextGetsRevrtdFlag.set(i, b);

      return this;
    }

    TraceBuilder setContextMayChangeFlagAt(final Boolean b, int i) {
      contextMayChangeFlag.set(i, b);

      return this;
    }

    TraceBuilder setContextNumberAt(final BigInteger b, int i) {
      contextNumber.set(i, b);

      return this;
    }

    TraceBuilder setContextNumberNewAt(final BigInteger b, int i) {
      contextNumberNew.set(i, b);

      return this;
    }

    TraceBuilder setContextRevertStampAt(final BigInteger b, int i) {
      contextRevertStamp.set(i, b);

      return this;
    }

    TraceBuilder setContextSelfRevrtsFlagAt(final Boolean b, int i) {
      contextSelfRevrtsFlag.set(i, b);

      return this;
    }

    TraceBuilder setContextWillRevertFlagAt(final Boolean b, int i) {
      contextWillRevertFlag.set(i, b);

      return this;
    }

    TraceBuilder setCounterNsrAt(final BigInteger b, int i) {
      counterNsr.set(i, b);

      return this;
    }

    TraceBuilder setCounterTliAt(final Boolean b, int i) {
      counterTli.set(i, b);

      return this;
    }

    TraceBuilder setExceptionAhoyFlagAt(final Boolean b, int i) {
      exceptionAhoyFlag.set(i, b);

      return this;
    }

    TraceBuilder setFailureConditionFlagAt(final Boolean b, int i) {
      failureConditionFlag.set(i, b);

      return this;
    }

    TraceBuilder setGasActualAt(final BigInteger b, int i) {
      gasActual.set(i, b);

      return this;
    }

    TraceBuilder setGasCostAt(final BigInteger b, int i) {
      gasCost.set(i, b);

      return this;
    }

    TraceBuilder setGasExpectedAt(final BigInteger b, int i) {
      gasExpected.set(i, b);

      return this;
    }

    TraceBuilder setGasMemoryExpansionAt(final BigInteger b, int i) {
      gasMemoryExpansion.set(i, b);

      return this;
    }

    TraceBuilder setGasNextAt(final BigInteger b, int i) {
      gasNext.set(i, b);

      return this;
    }

    TraceBuilder setGasRefundAt(final BigInteger b, int i) {
      gasRefund.set(i, b);

      return this;
    }

    TraceBuilder setHubStampAt(final BigInteger b, int i) {
      hubStamp.set(i, b);

      return this;
    }

    TraceBuilder setNumberOfNonStackRowsAt(final BigInteger b, int i) {
      numberOfNonStackRows.set(i, b);

      return this;
    }

    TraceBuilder setPAccountAddressHiAt(final BigInteger b, int i) {
      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.set(i, b);

      return this;
    }

    TraceBuilder setPAccountAddressLoAt(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(i, b);

      return this;
    }

    TraceBuilder setPAccountBalanceAt(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(i, b);

      return this;
    }

    TraceBuilder setPAccountBalanceNewAt(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashHiAt(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashHiNewAt(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashLoAt(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashLoNewAt(final BigInteger b, int i) {
      codeHashLoNewXorCallValueXorStackItemStamp1.set(i, b);

      return this;
    }

    TraceBuilder setPAccountCodeSizeAt(final BigInteger b, int i) {
      codeSizeXorCallDataOffsetXorStackItemStamp2.set(i, b);

      return this;
    }

    TraceBuilder setPAccountCodeSizeNewAt(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentNumberAt(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentNumberInftyAt(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentNumberNewAt(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentStatusAt(final BigInteger b, int i) {
      deploymentStatusXorReturnAtSizeXorHeight.set(i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentStatusInftyAt(final Boolean b, int i) {
      maxcsxXorWarmNewXorDeploymentStatusInfty.set(i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentStatusNewAt(final BigInteger b, int i) {
      deploymentStatusNewXorIsStaticXorInstruction.set(i, b);

      return this;
    }

    TraceBuilder setPAccountExistsAt(final Boolean b, int i) {
      oogxXorValNextIsZeroXorExists.set(i, b);

      return this;
    }

    TraceBuilder setPAccountExistsNewAt(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(i, b);

      return this;
    }

    TraceBuilder setPAccountHasCodeAt(final Boolean b, int i) {
      decodedFlag2XorValOrigIsZeroXorHasCode.set(i, b);

      return this;
    }

    TraceBuilder setPAccountHasCodeNewAt(final Boolean b, int i) {
      staticFlagXorWarmXorHasCodeNew.set(i, b);

      return this;
    }

    TraceBuilder setPAccountIsPrecompileAt(final Boolean b, int i) {
      dupFlagXorValNextIsOrigXorIsPrecompile.set(i, b);

      return this;
    }

    TraceBuilder setPAccountNonceAt(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(i, b);

      return this;
    }

    TraceBuilder setPAccountNonceNewAt(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(i, b);

      return this;
    }

    TraceBuilder setPAccountSufficientBalanceAt(final Boolean b, int i) {
      trmFlagXorValCurrIsZeroXorSufficientBalance.set(i, b);

      return this;
    }

    TraceBuilder setPAccountWarmAt(final Boolean b, int i) {
      accFlagXorValNextIsCurrXorWarm.set(i, b);

      return this;
    }

    TraceBuilder setPAccountWarmNewAt(final Boolean b, int i) {
      pushpopFlagXorValCurrIsOrigXorWarmNew.set(i, b);

      return this;
    }

    TraceBuilder setPContextAccountAddressHiAt(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(i, b);

      return this;
    }

    TraceBuilder setPContextAccountAddressLoAt(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          i, b);

      return this;
    }

    TraceBuilder setPContextAccountDeploymentNumberAt(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(i, b);

      return this;
    }

    TraceBuilder setPContextByteCodeAddressHiAt(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(i, b);

      return this;
    }

    TraceBuilder setPContextByteCodeAddressLoAt(final BigInteger b, int i) {
      byteCodeAddressLoXorStackItemStamp4.set(i, b);

      return this;
    }

    TraceBuilder setPContextByteCodeDeploymentNumberAt(final BigInteger b, int i) {
      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.set(i, b);

      return this;
    }

    TraceBuilder setPContextByteCodeDeploymentStatusAt(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(i, b);

      return this;
    }

    TraceBuilder setPContextCallDataOffsetAt(final BigInteger b, int i) {
      codeSizeXorCallDataOffsetXorStackItemStamp2.set(i, b);

      return this;
    }

    TraceBuilder setPContextCallDataSizeAt(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(i, b);

      return this;
    }

    TraceBuilder setPContextCallStackDepthAt(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(i, b);

      return this;
    }

    TraceBuilder setPContextCallValueAt(final BigInteger b, int i) {
      codeHashLoNewXorCallValueXorStackItemStamp1.set(i, b);

      return this;
    }

    TraceBuilder setPContextCallerAddressHiAt(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(i, b);

      return this;
    }

    TraceBuilder setPContextCallerAddressLoAt(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(i, b);

      return this;
    }

    TraceBuilder setPContextCallerContextNumberAt(final BigInteger b, int i) {
      callerContextNumberXorPushValueLo.set(i, b);

      return this;
    }

    TraceBuilder setPContextContextNumberAt(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(i, b);

      return this;
    }

    TraceBuilder setPContextIsStaticAt(final BigInteger b, int i) {
      deploymentStatusNewXorIsStaticXorInstruction.set(i, b);

      return this;
    }

    TraceBuilder setPContextReturnAtOffsetAt(final BigInteger b, int i) {
      returnAtOffsetXorPushValueHi.set(i, b);

      return this;
    }

    TraceBuilder setPContextReturnAtSizeAt(final BigInteger b, int i) {
      deploymentStatusXorReturnAtSizeXorHeight.set(i, b);

      return this;
    }

    TraceBuilder setPContextReturnDataOffsetAt(final BigInteger b, int i) {
      returnDataOffsetXorStackItemStamp3.set(i, b);

      return this;
    }

    TraceBuilder setPContextReturnDataSizeAt(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(i, b);

      return this;
    }

    TraceBuilder setPContextReturnerContextNumberAt(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(i, b);

      return this;
    }

    TraceBuilder setPContextReturnerIsPrecompileAt(final BigInteger b, int i) {
      returnerIsPrecompileXorStackItemHeight4.set(i, b);

      return this;
    }

    TraceBuilder setPContextUpdateAt(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(i, b);

      return this;
    }

    TraceBuilder setPStackAccFlagAt(final Boolean b, int i) {
      accFlagXorValNextIsCurrXorWarm.set(i, b);

      return this;
    }

    TraceBuilder setPStackAddFlagAt(final Boolean b, int i) {
      addFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackBinFlagAt(final Boolean b, int i) {
      binFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackBtcFlagAt(final Boolean b, int i) {
      btcFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackCallFlagAt(final Boolean b, int i) {
      callFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackConFlagAt(final Boolean b, int i) {
      conFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackCopyFlagAt(final Boolean b, int i) {
      copyFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackCreateFlagAt(final Boolean b, int i) {
      createFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag1At(final Boolean b, int i) {
      decodedFlag1.set(i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag2At(final Boolean b, int i) {
      decodedFlag2XorValOrigIsZeroXorHasCode.set(i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag3At(final Boolean b, int i) {
      decodedFlag3.set(i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag4At(final Boolean b, int i) {
      decodedFlag4.set(i, b);

      return this;
    }

    TraceBuilder setPStackDupFlagAt(final Boolean b, int i) {
      dupFlagXorValNextIsOrigXorIsPrecompile.set(i, b);

      return this;
    }

    TraceBuilder setPStackExtFlagAt(final Boolean b, int i) {
      extFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackHaltFlagAt(final Boolean b, int i) {
      haltFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackHeightAt(final BigInteger b, int i) {
      deploymentStatusXorReturnAtSizeXorHeight.set(i, b);

      return this;
    }

    TraceBuilder setPStackHeightNewAt(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(i, b);

      return this;
    }

    TraceBuilder setPStackHeightOverAt(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(i, b);

      return this;
    }

    TraceBuilder setPStackHeightUnderAt(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(i, b);

      return this;
    }

    TraceBuilder setPStackInstructionAt(final BigInteger b, int i) {
      deploymentStatusNewXorIsStaticXorInstruction.set(i, b);

      return this;
    }

    TraceBuilder setPStackInvalidFlagAt(final Boolean b, int i) {
      invalidFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackInvprexAt(final Boolean b, int i) {
      invprex.set(i, b);

      return this;
    }

    TraceBuilder setPStackJumpFlagAt(final Boolean b, int i) {
      jumpFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackJumpxAt(final Boolean b, int i) {
      jumpx.set(i, b);

      return this;
    }

    TraceBuilder setPStackKecFlagAt(final Boolean b, int i) {
      kecFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackLogFlagAt(final Boolean b, int i) {
      logFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackMaxcsxAt(final Boolean b, int i) {
      maxcsxXorWarmNewXorDeploymentStatusInfty.set(i, b);

      return this;
    }

    TraceBuilder setPStackModFlagAt(final Boolean b, int i) {
      modFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackMulFlagAt(final Boolean b, int i) {
      mulFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackMxpFlagAt(final Boolean b, int i) {
      mxpFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackMxpxAt(final Boolean b, int i) {
      mxpx.set(i, b);

      return this;
    }

    TraceBuilder setPStackOobFlagAt(final Boolean b, int i) {
      oobFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackOogxAt(final Boolean b, int i) {
      oogxXorValNextIsZeroXorExists.set(i, b);

      return this;
    }

    TraceBuilder setPStackOpcxAt(final Boolean b, int i) {
      opcx.set(i, b);

      return this;
    }

    TraceBuilder setPStackPushValueHiAt(final BigInteger b, int i) {
      returnAtOffsetXorPushValueHi.set(i, b);

      return this;
    }

    TraceBuilder setPStackPushValueLoAt(final BigInteger b, int i) {
      callerContextNumberXorPushValueLo.set(i, b);

      return this;
    }

    TraceBuilder setPStackPushpopFlagAt(final Boolean b, int i) {
      pushpopFlagXorValCurrIsOrigXorWarmNew.set(i, b);

      return this;
    }

    TraceBuilder setPStackRdcxAt(final Boolean b, int i) {
      rdcx.set(i, b);

      return this;
    }

    TraceBuilder setPStackShfFlagAt(final Boolean b, int i) {
      shfFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackSoxAt(final Boolean b, int i) {
      sox.set(i, b);

      return this;
    }

    TraceBuilder setPStackSstorexAt(final Boolean b, int i) {
      sstorex.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight1At(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight2At(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight3At(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight4At(final BigInteger b, int i) {
      returnerIsPrecompileXorStackItemHeight4.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop1At(final Boolean b, int i) {
      stackItemPop1.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop2At(final Boolean b, int i) {
      stackItemPop2.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop3At(final Boolean b, int i) {
      stackItemPop3.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop4At(final Boolean b, int i) {
      stackItemPop4.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp1At(final BigInteger b, int i) {
      codeHashLoNewXorCallValueXorStackItemStamp1.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp2At(final BigInteger b, int i) {
      codeSizeXorCallDataOffsetXorStackItemStamp2.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp3At(final BigInteger b, int i) {
      returnDataOffsetXorStackItemStamp3.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp4At(final BigInteger b, int i) {
      byteCodeAddressLoXorStackItemStamp4.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi1At(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi2At(final BigInteger b, int i) {
      stackItemValueHi2.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi3At(final BigInteger b, int i) {
      stackItemValueHi3.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi4At(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo1At(final BigInteger b, int i) {
      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo2At(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo3At(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo4At(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(i, b);

      return this;
    }

    TraceBuilder setPStackStackramFlagAt(final Boolean b, int i) {
      stackramFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackStaticFlagAt(final Boolean b, int i) {
      staticFlagXorWarmXorHasCodeNew.set(i, b);

      return this;
    }

    TraceBuilder setPStackStaticGasAt(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(i, b);

      return this;
    }

    TraceBuilder setPStackStaticxAt(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(i, b);

      return this;
    }

    TraceBuilder setPStackStoFlagAt(final Boolean b, int i) {
      stoFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackSuxAt(final Boolean b, int i) {
      sux.set(i, b);

      return this;
    }

    TraceBuilder setPStackSwapFlagAt(final Boolean b, int i) {
      swapFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackTrmFlagAt(final Boolean b, int i) {
      trmFlagXorValCurrIsZeroXorSufficientBalance.set(i, b);

      return this;
    }

    TraceBuilder setPStackTxnFlagAt(final Boolean b, int i) {
      txnFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStackWcpFlagAt(final Boolean b, int i) {
      wcpFlag.set(i, b);

      return this;
    }

    TraceBuilder setPStorageAddressHiAt(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(i, b);

      return this;
    }

    TraceBuilder setPStorageAddressLoAt(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(i, b);

      return this;
    }

    TraceBuilder setPStorageDeploymentNumberAt(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(i, b);

      return this;
    }

    TraceBuilder setPStorageStorageKeyHiAt(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(i, b);

      return this;
    }

    TraceBuilder setPStorageStorageKeyLoAt(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrChangesAt(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrHiAt(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrIsOrigAt(final Boolean b, int i) {
      pushpopFlagXorValCurrIsOrigXorWarmNew.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrIsZeroAt(final Boolean b, int i) {
      trmFlagXorValCurrIsZeroXorSufficientBalance.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrLoAt(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValNextHiAt(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValNextIsCurrAt(final Boolean b, int i) {
      accFlagXorValNextIsCurrXorWarm.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValNextIsOrigAt(final Boolean b, int i) {
      dupFlagXorValNextIsOrigXorIsPrecompile.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValNextIsZeroAt(final Boolean b, int i) {
      oogxXorValNextIsZeroXorExists.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValNextLoAt(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValOrigHiAt(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValOrigIsZeroAt(final Boolean b, int i) {
      decodedFlag2XorValOrigIsZeroXorHasCode.set(i, b);

      return this;
    }

    TraceBuilder setPStorageValOrigLoAt(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(i, b);

      return this;
    }

    TraceBuilder setPStorageWarmAt(final Boolean b, int i) {
      staticFlagXorWarmXorHasCodeNew.set(i, b);

      return this;
    }

    TraceBuilder setPStorageWarmNewAt(final Boolean b, int i) {
      maxcsxXorWarmNewXorDeploymentStatusInfty.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionAbsoluteTransactionNumberAt(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(i, b);

      return this;
    }

    TraceBuilder setPTransactionBatchNumberAt(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionFromAddressHiAt(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionFromAddressLoAt(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(i, b);

      return this;
    }

    TraceBuilder setPTransactionGasFeeAt(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionGasMaxfeeAt(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionGasTipAt(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionInitGasAt(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionIsDeploymentAt(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionNonceAt(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionToAddressHiAt(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          i, b);

      return this;
    }

    TraceBuilder setPTransactionToAddressLoAt(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(i, b);

      return this;
    }

    TraceBuilder setPTransactionValueAt(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(i, b);

      return this;
    }

    TraceBuilder setPeekAtAccountAt(final Boolean b, int i) {
      peekAtAccount.set(i, b);

      return this;
    }

    TraceBuilder setPeekAtContextAt(final Boolean b, int i) {
      peekAtContext.set(i, b);

      return this;
    }

    TraceBuilder setPeekAtStackAt(final Boolean b, int i) {
      peekAtStack.set(i, b);

      return this;
    }

    TraceBuilder setPeekAtStorageAt(final Boolean b, int i) {
      peekAtStorage.set(i, b);

      return this;
    }

    TraceBuilder setPeekAtTransactionAt(final Boolean b, int i) {
      peekAtTransaction.set(i, b);

      return this;
    }

    TraceBuilder setProgramCounterAt(final BigInteger b, int i) {
      programCounter.set(i, b);

      return this;
    }

    TraceBuilder setProgramCounterNewAt(final BigInteger b, int i) {
      programCounterNew.set(i, b);

      return this;
    }

    TraceBuilder setTransactionEndStampAt(final BigInteger b, int i) {
      transactionEndStamp.set(i, b);

      return this;
    }

    TraceBuilder setTransactionRevertsAt(final BigInteger b, int i) {
      transactionReverts.set(i, b);

      return this;
    }

    TraceBuilder setTwoLineInstructionAt(final Boolean b, int i) {
      twoLineInstruction.set(i, b);

      return this;
    }

    TraceBuilder setTxExecAt(final Boolean b, int i) {
      txExec.set(i, b);

      return this;
    }

    TraceBuilder setTxFinlAt(final Boolean b, int i) {
      txFinl.set(i, b);

      return this;
    }

    TraceBuilder setTxInitAt(final Boolean b, int i) {
      txInit.set(i, b);

      return this;
    }

    TraceBuilder setTxSkipAt(final Boolean b, int i) {
      txSkip.set(i, b);

      return this;
    }

    TraceBuilder setTxWarmAt(final Boolean b, int i) {
      txWarm.set(i, b);

      return this;
    }

    TraceBuilder setAbortFlagRelative(final Boolean b, int i) {
      abortFlag.set(abortFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAbsoluteTransactionNumberRelative(final BigInteger b, int i) {
      absoluteTransactionNumber.set(absoluteTransactionNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBatchNumberRelative(final BigInteger b, int i) {
      batchNumber.set(batchNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCallerContextNumberRelative(final BigInteger b, int i) {
      callerContextNumber.set(callerContextNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCodeAddressHiRelative(final BigInteger b, int i) {
      codeAddressHi.set(codeAddressHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCodeAddressLoRelative(final BigInteger b, int i) {
      codeAddressLo.set(codeAddressLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCodeDeploymentNumberRelative(final BigInteger b, int i) {
      codeDeploymentNumber.set(codeDeploymentNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCodeDeploymentStatusRelative(final Boolean b, int i) {
      codeDeploymentStatus.set(codeDeploymentStatus.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextGetsRevrtdFlagRelative(final Boolean b, int i) {
      contextGetsRevrtdFlag.set(contextGetsRevrtdFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextMayChangeFlagRelative(final Boolean b, int i) {
      contextMayChangeFlag.set(contextMayChangeFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextNumberRelative(final BigInteger b, int i) {
      contextNumber.set(contextNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextNumberNewRelative(final BigInteger b, int i) {
      contextNumberNew.set(contextNumberNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextRevertStampRelative(final BigInteger b, int i) {
      contextRevertStamp.set(contextRevertStamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextSelfRevrtsFlagRelative(final Boolean b, int i) {
      contextSelfRevrtsFlag.set(contextSelfRevrtsFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setContextWillRevertFlagRelative(final Boolean b, int i) {
      contextWillRevertFlag.set(contextWillRevertFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterNsrRelative(final BigInteger b, int i) {
      counterNsr.set(counterNsr.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterTliRelative(final Boolean b, int i) {
      counterTli.set(counterTli.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setExceptionAhoyFlagRelative(final Boolean b, int i) {
      exceptionAhoyFlag.set(exceptionAhoyFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setFailureConditionFlagRelative(final Boolean b, int i) {
      failureConditionFlag.set(failureConditionFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setGasActualRelative(final BigInteger b, int i) {
      gasActual.set(gasActual.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setGasCostRelative(final BigInteger b, int i) {
      gasCost.set(gasCost.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setGasExpectedRelative(final BigInteger b, int i) {
      gasExpected.set(gasExpected.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setGasMemoryExpansionRelative(final BigInteger b, int i) {
      gasMemoryExpansion.set(gasMemoryExpansion.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setGasNextRelative(final BigInteger b, int i) {
      gasNext.set(gasNext.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setGasRefundRelative(final BigInteger b, int i) {
      gasRefund.set(gasRefund.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setHubStampRelative(final BigInteger b, int i) {
      hubStamp.set(hubStamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNumberOfNonStackRowsRelative(final BigInteger b, int i) {
      numberOfNonStackRows.set(numberOfNonStackRows.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountAddressHiRelative(final BigInteger b, int i) {
      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.set(
          addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountAddressLoRelative(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(
          addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPAccountBalanceRelative(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(
          balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPAccountBalanceNewRelative(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(
          balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashHiRelative(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(
          codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashHiNewRelative(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(
              codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPAccountCodeHashLoRelative(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(
          codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountCodeHashLoNewRelative(final BigInteger b, int i) {
      codeHashLoNewXorCallValueXorStackItemStamp1.set(
          codeHashLoNewXorCallValueXorStackItemStamp1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountCodeSizeRelative(final BigInteger b, int i) {
      codeSizeXorCallDataOffsetXorStackItemStamp2.set(
          codeSizeXorCallDataOffsetXorStackItemStamp2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountCodeSizeNewRelative(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(
          codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentNumberRelative(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(
          deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPAccountDeploymentNumberInftyRelative(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(
              deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPAccountDeploymentNumberNewRelative(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi
                  .size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPAccountDeploymentStatusRelative(final BigInteger b, int i) {
      deploymentStatusXorReturnAtSizeXorHeight.set(
          deploymentStatusXorReturnAtSizeXorHeight.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentStatusInftyRelative(final Boolean b, int i) {
      maxcsxXorWarmNewXorDeploymentStatusInfty.set(
          maxcsxXorWarmNewXorDeploymentStatusInfty.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountDeploymentStatusNewRelative(final BigInteger b, int i) {
      deploymentStatusNewXorIsStaticXorInstruction.set(
          deploymentStatusNewXorIsStaticXorInstruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountExistsRelative(final Boolean b, int i) {
      oogxXorValNextIsZeroXorExists.set(oogxXorValNextIsZeroXorExists.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountExistsNewRelative(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(
          staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountHasCodeRelative(final Boolean b, int i) {
      decodedFlag2XorValOrigIsZeroXorHasCode.set(
          decodedFlag2XorValOrigIsZeroXorHasCode.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountHasCodeNewRelative(final Boolean b, int i) {
      staticFlagXorWarmXorHasCodeNew.set(staticFlagXorWarmXorHasCodeNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountIsPrecompileRelative(final Boolean b, int i) {
      dupFlagXorValNextIsOrigXorIsPrecompile.set(
          dupFlagXorValNextIsOrigXorIsPrecompile.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountNonceRelative(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(
          nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountNonceNewRelative(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(
          nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountSufficientBalanceRelative(final Boolean b, int i) {
      trmFlagXorValCurrIsZeroXorSufficientBalance.set(
          trmFlagXorValCurrIsZeroXorSufficientBalance.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountWarmRelative(final Boolean b, int i) {
      accFlagXorValNextIsCurrXorWarm.set(accFlagXorValNextIsCurrXorWarm.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPAccountWarmNewRelative(final Boolean b, int i) {
      pushpopFlagXorValCurrIsOrigXorWarmNew.set(
          pushpopFlagXorValCurrIsOrigXorWarmNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextAccountAddressHiRelative(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(
          codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextAccountAddressLoRelative(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi
                  .size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPContextAccountDeploymentNumberRelative(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(
          deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPContextByteCodeAddressHiRelative(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(
          balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPContextByteCodeAddressLoRelative(final BigInteger b, int i) {
      byteCodeAddressLoXorStackItemStamp4.set(
          byteCodeAddressLoXorStackItemStamp4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextByteCodeDeploymentNumberRelative(final BigInteger b, int i) {
      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.set(
          addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextByteCodeDeploymentStatusRelative(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(
              deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPContextCallDataOffsetRelative(final BigInteger b, int i) {
      codeSizeXorCallDataOffsetXorStackItemStamp2.set(
          codeSizeXorCallDataOffsetXorStackItemStamp2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextCallDataSizeRelative(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(
          nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextCallStackDepthRelative(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(
          balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextCallValueRelative(final BigInteger b, int i) {
      codeHashLoNewXorCallValueXorStackItemStamp1.set(
          codeHashLoNewXorCallValueXorStackItemStamp1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextCallerAddressHiRelative(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(
          addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPContextCallerAddressLoRelative(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(
              codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPContextCallerContextNumberRelative(final BigInteger b, int i) {
      callerContextNumberXorPushValueLo.set(callerContextNumberXorPushValueLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextContextNumberRelative(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(
          nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextIsStaticRelative(final BigInteger b, int i) {
      deploymentStatusNewXorIsStaticXorInstruction.set(
          deploymentStatusNewXorIsStaticXorInstruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextReturnAtOffsetRelative(final BigInteger b, int i) {
      returnAtOffsetXorPushValueHi.set(returnAtOffsetXorPushValueHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextReturnAtSizeRelative(final BigInteger b, int i) {
      deploymentStatusXorReturnAtSizeXorHeight.set(
          deploymentStatusXorReturnAtSizeXorHeight.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextReturnDataOffsetRelative(final BigInteger b, int i) {
      returnDataOffsetXorStackItemStamp3.set(returnDataOffsetXorStackItemStamp3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextReturnDataSizeRelative(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(
          codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextReturnerContextNumberRelative(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(
          codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextReturnerIsPrecompileRelative(final BigInteger b, int i) {
      returnerIsPrecompileXorStackItemHeight4.set(
          returnerIsPrecompileXorStackItemHeight4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPContextUpdateRelative(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(
          staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackAccFlagRelative(final Boolean b, int i) {
      accFlagXorValNextIsCurrXorWarm.set(accFlagXorValNextIsCurrXorWarm.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackAddFlagRelative(final Boolean b, int i) {
      addFlag.set(addFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackBinFlagRelative(final Boolean b, int i) {
      binFlag.set(binFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackBtcFlagRelative(final Boolean b, int i) {
      btcFlag.set(btcFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackCallFlagRelative(final Boolean b, int i) {
      callFlag.set(callFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackConFlagRelative(final Boolean b, int i) {
      conFlag.set(conFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackCopyFlagRelative(final Boolean b, int i) {
      copyFlag.set(copyFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackCreateFlagRelative(final Boolean b, int i) {
      createFlag.set(createFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag1Relative(final Boolean b, int i) {
      decodedFlag1.set(decodedFlag1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag2Relative(final Boolean b, int i) {
      decodedFlag2XorValOrigIsZeroXorHasCode.set(
          decodedFlag2XorValOrigIsZeroXorHasCode.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag3Relative(final Boolean b, int i) {
      decodedFlag3.set(decodedFlag3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackDecodedFlag4Relative(final Boolean b, int i) {
      decodedFlag4.set(decodedFlag4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackDupFlagRelative(final Boolean b, int i) {
      dupFlagXorValNextIsOrigXorIsPrecompile.set(
          dupFlagXorValNextIsOrigXorIsPrecompile.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackExtFlagRelative(final Boolean b, int i) {
      extFlag.set(extFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackHaltFlagRelative(final Boolean b, int i) {
      haltFlag.set(haltFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackHeightRelative(final BigInteger b, int i) {
      deploymentStatusXorReturnAtSizeXorHeight.set(
          deploymentStatusXorReturnAtSizeXorHeight.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackHeightNewRelative(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(
          nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackHeightOverRelative(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(
          deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPStackHeightUnderRelative(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(
          codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackInstructionRelative(final BigInteger b, int i) {
      deploymentStatusNewXorIsStaticXorInstruction.set(
          deploymentStatusNewXorIsStaticXorInstruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackInvalidFlagRelative(final Boolean b, int i) {
      invalidFlag.set(invalidFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackInvprexRelative(final Boolean b, int i) {
      invprex.set(invprex.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackJumpFlagRelative(final Boolean b, int i) {
      jumpFlag.set(jumpFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackJumpxRelative(final Boolean b, int i) {
      jumpx.set(jumpx.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackKecFlagRelative(final Boolean b, int i) {
      kecFlag.set(kecFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackLogFlagRelative(final Boolean b, int i) {
      logFlag.set(logFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackMaxcsxRelative(final Boolean b, int i) {
      maxcsxXorWarmNewXorDeploymentStatusInfty.set(
          maxcsxXorWarmNewXorDeploymentStatusInfty.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackModFlagRelative(final Boolean b, int i) {
      modFlag.set(modFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackMulFlagRelative(final Boolean b, int i) {
      mulFlag.set(mulFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackMxpFlagRelative(final Boolean b, int i) {
      mxpFlag.set(mxpFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackMxpxRelative(final Boolean b, int i) {
      mxpx.set(mxpx.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackOobFlagRelative(final Boolean b, int i) {
      oobFlag.set(oobFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackOogxRelative(final Boolean b, int i) {
      oogxXorValNextIsZeroXorExists.set(oogxXorValNextIsZeroXorExists.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackOpcxRelative(final Boolean b, int i) {
      opcx.set(opcx.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackPushValueHiRelative(final BigInteger b, int i) {
      returnAtOffsetXorPushValueHi.set(returnAtOffsetXorPushValueHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackPushValueLoRelative(final BigInteger b, int i) {
      callerContextNumberXorPushValueLo.set(callerContextNumberXorPushValueLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackPushpopFlagRelative(final Boolean b, int i) {
      pushpopFlagXorValCurrIsOrigXorWarmNew.set(
          pushpopFlagXorValCurrIsOrigXorWarmNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackRdcxRelative(final Boolean b, int i) {
      rdcx.set(rdcx.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackShfFlagRelative(final Boolean b, int i) {
      shfFlag.set(shfFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackSoxRelative(final Boolean b, int i) {
      sox.set(sox.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackSstorexRelative(final Boolean b, int i) {
      sstorex.set(sstorex.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight1Relative(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(
          addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight2Relative(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(
          nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight3Relative(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(
          codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemHeight4Relative(final BigInteger b, int i) {
      returnerIsPrecompileXorStackItemHeight4.set(
          returnerIsPrecompileXorStackItemHeight4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop1Relative(final Boolean b, int i) {
      stackItemPop1.set(stackItemPop1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop2Relative(final Boolean b, int i) {
      stackItemPop2.set(stackItemPop2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop3Relative(final Boolean b, int i) {
      stackItemPop3.set(stackItemPop3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemPop4Relative(final Boolean b, int i) {
      stackItemPop4.set(stackItemPop4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp1Relative(final BigInteger b, int i) {
      codeHashLoNewXorCallValueXorStackItemStamp1.set(
          codeHashLoNewXorCallValueXorStackItemStamp1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp2Relative(final BigInteger b, int i) {
      codeSizeXorCallDataOffsetXorStackItemStamp2.set(
          codeSizeXorCallDataOffsetXorStackItemStamp2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp3Relative(final BigInteger b, int i) {
      returnDataOffsetXorStackItemStamp3.set(returnDataOffsetXorStackItemStamp3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemStamp4Relative(final BigInteger b, int i) {
      byteCodeAddressLoXorStackItemStamp4.set(
          byteCodeAddressLoXorStackItemStamp4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi1Relative(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi
                  .size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi2Relative(final BigInteger b, int i) {
      stackItemValueHi2.set(stackItemValueHi2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi3Relative(final BigInteger b, int i) {
      stackItemValueHi3.set(stackItemValueHi3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueHi4Relative(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(
              deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo1Relative(final BigInteger b, int i) {
      addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.set(
          addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo2Relative(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(
          balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo3Relative(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(
              codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPStackStackItemValueLo4Relative(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(
          balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPStackStackramFlagRelative(final Boolean b, int i) {
      stackramFlag.set(stackramFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStaticFlagRelative(final Boolean b, int i) {
      staticFlagXorWarmXorHasCodeNew.set(staticFlagXorWarmXorHasCodeNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStaticGasRelative(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(
          codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStaticxRelative(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(
          staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackStoFlagRelative(final Boolean b, int i) {
      stoFlag.set(stoFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackSuxRelative(final Boolean b, int i) {
      sux.set(sux.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackSwapFlagRelative(final Boolean b, int i) {
      swapFlag.set(swapFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackTrmFlagRelative(final Boolean b, int i) {
      trmFlagXorValCurrIsZeroXorSufficientBalance.set(
          trmFlagXorValCurrIsZeroXorSufficientBalance.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackTxnFlagRelative(final Boolean b, int i) {
      txnFlag.set(txnFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStackWcpFlagRelative(final Boolean b, int i) {
      wcpFlag.set(wcpFlag.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageAddressHiRelative(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(
          deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPStorageAddressLoRelative(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(
              codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPStorageDeploymentNumberRelative(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(
              deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPStorageStorageKeyHiRelative(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(
          balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPStorageStorageKeyLoRelative(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi
                  .size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPStorageValCurrChangesRelative(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(
          staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrHiRelative(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(
          balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrIsOrigRelative(final Boolean b, int i) {
      pushpopFlagXorValCurrIsOrigXorWarmNew.set(
          pushpopFlagXorValCurrIsOrigXorWarmNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrIsZeroRelative(final Boolean b, int i) {
      trmFlagXorValCurrIsZeroXorSufficientBalance.set(
          trmFlagXorValCurrIsZeroXorSufficientBalance.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValCurrLoRelative(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(
          addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPStorageValNextHiRelative(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(
          nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValNextIsCurrRelative(final Boolean b, int i) {
      accFlagXorValNextIsCurrXorWarm.set(accFlagXorValNextIsCurrXorWarm.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValNextIsOrigRelative(final Boolean b, int i) {
      dupFlagXorValNextIsOrigXorIsPrecompile.set(
          dupFlagXorValNextIsOrigXorIsPrecompile.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValNextIsZeroRelative(final Boolean b, int i) {
      oogxXorValNextIsZeroXorExists.set(oogxXorValNextIsZeroXorExists.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValNextLoRelative(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(
          codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValOrigHiRelative(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(
          codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValOrigIsZeroRelative(final Boolean b, int i) {
      decodedFlag2XorValOrigIsZeroXorHasCode.set(
          decodedFlag2XorValOrigIsZeroXorHasCode.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageValOrigLoRelative(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(
          nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageWarmRelative(final Boolean b, int i) {
      staticFlagXorWarmXorHasCodeNew.set(staticFlagXorWarmXorHasCodeNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPStorageWarmNewRelative(final Boolean b, int i) {
      maxcsxXorWarmNewXorDeploymentStatusInfty.set(
          maxcsxXorWarmNewXorDeploymentStatusInfty.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionAbsoluteTransactionNumberRelative(final BigInteger b, int i) {
      codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
          .set(
              codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPTransactionBatchNumberRelative(final BigInteger b, int i) {
      codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.set(
          codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionFromAddressHiRelative(final BigInteger b, int i) {
      addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.set(
          addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPTransactionFromAddressLoRelative(final BigInteger b, int i) {
      deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
          .set(
              deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
                      .size()
                  - 1
                  - i,
              b);

      return this;
    }

    TraceBuilder setPTransactionGasFeeRelative(final BigInteger b, int i) {
      balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.set(
          balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.size() - 1 - i,
          b);

      return this;
    }

    TraceBuilder setPTransactionGasMaxfeeRelative(final BigInteger b, int i) {
      nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.set(
          nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionGasTipRelative(final BigInteger b, int i) {
      balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.set(
          balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionInitGasRelative(final BigInteger b, int i) {
      deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.set(
          deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPTransactionIsDeploymentRelative(final Boolean b, int i) {
      staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.set(
          staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionNonceRelative(final BigInteger b, int i) {
      codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.set(
          codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionToAddressHiRelative(final BigInteger b, int i) {
      deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.set(
          deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi
                  .size()
              - 1
              - i,
          b);

      return this;
    }

    TraceBuilder setPTransactionToAddressLoRelative(final BigInteger b, int i) {
      nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.set(
          nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPTransactionValueRelative(final BigInteger b, int i) {
      codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.set(
          codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPeekAtAccountRelative(final Boolean b, int i) {
      peekAtAccount.set(peekAtAccount.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPeekAtContextRelative(final Boolean b, int i) {
      peekAtContext.set(peekAtContext.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPeekAtStackRelative(final Boolean b, int i) {
      peekAtStack.set(peekAtStack.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPeekAtStorageRelative(final Boolean b, int i) {
      peekAtStorage.set(peekAtStorage.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPeekAtTransactionRelative(final Boolean b, int i) {
      peekAtTransaction.set(peekAtTransaction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setProgramCounterRelative(final BigInteger b, int i) {
      programCounter.set(programCounter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setProgramCounterNewRelative(final BigInteger b, int i) {
      programCounterNew.set(programCounterNew.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTransactionEndStampRelative(final BigInteger b, int i) {
      transactionEndStamp.set(transactionEndStamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTransactionRevertsRelative(final BigInteger b, int i) {
      transactionReverts.set(transactionReverts.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTwoLineInstructionRelative(final Boolean b, int i) {
      twoLineInstruction.set(twoLineInstruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTxExecRelative(final Boolean b, int i) {
      txExec.set(txExec.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTxFinlRelative(final Boolean b, int i) {
      txFinl.set(txFinl.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTxInitRelative(final Boolean b, int i) {
      txInit.set(txInit.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTxSkipRelative(final Boolean b, int i) {
      txSkip.set(txSkip.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTxWarmRelative(final Boolean b, int i) {
      txWarm.set(txWarm.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(41)) {
        throw new IllegalStateException("ABORT_FLAG has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ABSOLUTE_TRANSACTION_NUMBER has not been filled");
      }

      if (!filled.get(71)) {
        throw new IllegalStateException(
            "ACC_FLAG_xor_VAL_NEXT_IS_CURR_xor_WARM has not been filled");
      }

      if (!filled.get(88)) {
        throw new IllegalStateException("ADD_FLAG has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException(
            "ADDRESS_HI_xor_BYTE_CODE_DEPLOYMENT_NUMBER_xor_STACK_ITEM_VALUE_LO_1 has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException(
            "ADDRESS_LO_xor_VAL_CURR_LO_xor_CALLER_ADDRESS_HI_xor_STACK_ITEM_HEIGHT_1_xor_FROM_ADDRESS_HI has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException(
            "BALANCE_NEW_xor_VAL_CURR_HI_xor_CALL_STACK_DEPTH_xor_STACK_ITEM_VALUE_LO_2_xor_GAS_TIP has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException(
            "BALANCE_xor_STORAGE_KEY_HI_xor_BYTE_CODE_ADDRESS_HI_xor_STACK_ITEM_VALUE_LO_4_xor_GAS_FEE has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("BATCH_NUMBER has not been filled");
      }

      if (!filled.get(91)) {
        throw new IllegalStateException("BIN_FLAG has not been filled");
      }

      if (!filled.get(94)) {
        throw new IllegalStateException("BTC_FLAG has not been filled");
      }

      if (!filled.get(63)) {
        throw new IllegalStateException(
            "BYTE_CODE_ADDRESS_LO_xor_STACK_ITEM_STAMP_4 has not been filled");
      }

      if (!filled.get(90)) {
        throw new IllegalStateException("CALL_FLAG has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("CALLER_CONTEXT_NUMBER has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException(
            "CALLER_CONTEXT_NUMBER_xor_PUSH_VALUE_LO has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("CODE_ADDRESS_HI has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("CODE_ADDRESS_LO has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_NUMBER has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("CODE_DEPLOYMENT_STATUS has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException(
            "CODE_HASH_HI_NEW_xor_ADDRESS_LO_xor_CALLER_ADDRESS_LO_xor_STACK_ITEM_VALUE_LO_3_xor_ABSOLUTE_TRANSACTION_NUMBER has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException(
            "CODE_HASH_HI_xor_ACCOUNT_ADDRESS_HI_xor_STACK_ITEM_HEIGHT_3_xor_NONCE has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException(
            "CODE_HASH_LO_NEW_xor_CALL_VALUE_xor_STACK_ITEM_STAMP_1 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException(
            "CODE_HASH_LO_xor_VAL_ORIG_HI_xor_RETURN_DATA_SIZE_xor_HEIGHT_UNDER_xor_BATCH_NUMBER has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException(
            "CODE_SIZE_NEW_xor_VAL_NEXT_LO_xor_RETURNER_CONTEXT_NUMBER_xor_STATIC_GAS_xor_VALUE has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException(
            "CODE_SIZE_xor_CALL_DATA_OFFSET_xor_STACK_ITEM_STAMP_2 has not been filled");
      }

      if (!filled.get(112)) {
        throw new IllegalStateException("CON_FLAG has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("CONTEXT_GETS_REVRTD_FLAG has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("CONTEXT_MAY_CHANGE_FLAG has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("CONTEXT_NUMBER has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("CONTEXT_NUMBER_NEW has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("CONTEXT_REVERT_STAMP has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("CONTEXT_SELF_REVRTS_FLAG has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("CONTEXT_WILL_REVERT_FLAG has not been filled");
      }

      if (!filled.get(92)) {
        throw new IllegalStateException("COPY_FLAG has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("COUNTER_NSR has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("COUNTER_TLI has not been filled");
      }

      if (!filled.get(105)) {
        throw new IllegalStateException("CREATE_FLAG has not been filled");
      }

      if (!filled.get(81)) {
        throw new IllegalStateException("DECODED_FLAG_1 has not been filled");
      }

      if (!filled.get(69)) {
        throw new IllegalStateException(
            "DECODED_FLAG_2_xor_VAL_ORIG_IS_ZERO_xor_HAS_CODE has not been filled");
      }

      if (!filled.get(103)) {
        throw new IllegalStateException("DECODED_FLAG_3 has not been filled");
      }

      if (!filled.get(108)) {
        throw new IllegalStateException("DECODED_FLAG_4 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException(
            "DEPLOYMENT_NUMBER_INFTY_xor_DEPLOYMENT_NUMBER_xor_BYTE_CODE_DEPLOYMENT_STATUS_xor_STACK_ITEM_VALUE_HI_4_xor_FROM_ADDRESS_LO has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException(
            "DEPLOYMENT_NUMBER_NEW_xor_STORAGE_KEY_LO_xor_ACCOUNT_ADDRESS_LO_xor_STACK_ITEM_VALUE_HI_1_xor_TO_ADDRESS_HI has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException(
            "DEPLOYMENT_NUMBER_xor_ADDRESS_HI_xor_ACCOUNT_DEPLOYMENT_NUMBER_xor_HEIGHT_OVER_xor_INIT_GAS has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException(
            "DEPLOYMENT_STATUS_NEW_xor_IS_STATIC_xor_INSTRUCTION has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException(
            "DEPLOYMENT_STATUS_xor_RETURN_AT_SIZE_xor_HEIGHT has not been filled");
      }

      if (!filled.get(73)) {
        throw new IllegalStateException(
            "DUP_FLAG_xor_VAL_NEXT_IS_ORIG_xor_IS_PRECOMPILE has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("EXCEPTION_AHOY_FLAG has not been filled");
      }

      if (!filled.get(86)) {
        throw new IllegalStateException("EXT_FLAG has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("FAILURE_CONDITION_FLAG has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("GAS_ACTUAL has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("GAS_COST has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("GAS_EXPECTED has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("GAS_MEMORY_EXPANSION has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("GAS_NEXT has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("GAS_REFUND has not been filled");
      }

      if (!filled.get(99)) {
        throw new IllegalStateException("HALT_FLAG has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("HUB_STAMP has not been filled");
      }

      if (!filled.get(93)) {
        throw new IllegalStateException("INVALID_FLAG has not been filled");
      }

      if (!filled.get(109)) {
        throw new IllegalStateException("INVPREX has not been filled");
      }

      if (!filled.get(95)) {
        throw new IllegalStateException("JUMP_FLAG has not been filled");
      }

      if (!filled.get(80)) {
        throw new IllegalStateException("JUMPX has not been filled");
      }

      if (!filled.get(75)) {
        throw new IllegalStateException("KEC_FLAG has not been filled");
      }

      if (!filled.get(104)) {
        throw new IllegalStateException("LOG_FLAG has not been filled");
      }

      if (!filled.get(70)) {
        throw new IllegalStateException(
            "MAXCSX_xor_WARM_NEW_xor_DEPLOYMENT_STATUS_INFTY has not been filled");
      }

      if (!filled.get(111)) {
        throw new IllegalStateException("MOD_FLAG has not been filled");
      }

      if (!filled.get(84)) {
        throw new IllegalStateException("MUL_FLAG has not been filled");
      }

      if (!filled.get(107)) {
        throw new IllegalStateException("MXP_FLAG has not been filled");
      }

      if (!filled.get(79)) {
        throw new IllegalStateException("MXPX has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException(
            "NONCE_NEW_xor_VAL_NEXT_HI_xor_CONTEXT_NUMBER_xor_STACK_ITEM_HEIGHT_2_xor_TO_ADDRESS_LO has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException(
            "NONCE_xor_VAL_ORIG_LO_xor_CALL_DATA_SIZE_xor_HEIGHT_NEW_xor_GAS_MAXFEE has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("NUMBER_OF_NON_STACK_ROWS has not been filled");
      }

      if (!filled.get(76)) {
        throw new IllegalStateException("OOB_FLAG has not been filled");
      }

      if (!filled.get(67)) {
        throw new IllegalStateException("OOGX_xor_VAL_NEXT_IS_ZERO_xor_EXISTS has not been filled");
      }

      if (!filled.get(82)) {
        throw new IllegalStateException("OPCX has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("PEEK_AT_ACCOUNT has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("PEEK_AT_CONTEXT has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("PEEK_AT_STACK has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("PEEK_AT_STORAGE has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("PEEK_AT_TRANSACTION has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("PROGRAM_COUNTER has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("PROGRAM_COUNTER_NEW has not been filled");
      }

      if (!filled.get(72)) {
        throw new IllegalStateException(
            "PUSHPOP_FLAG_xor_VAL_CURR_IS_ORIG_xor_WARM_NEW has not been filled");
      }

      if (!filled.get(77)) {
        throw new IllegalStateException("RDCX has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("RETURN_AT_OFFSET_xor_PUSH_VALUE_HI has not been filled");
      }

      if (!filled.get(62)) {
        throw new IllegalStateException(
            "RETURN_DATA_OFFSET_xor_STACK_ITEM_STAMP_3 has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException(
            "RETURNER_IS_PRECOMPILE_xor_STACK_ITEM_HEIGHT_4 has not been filled");
      }

      if (!filled.get(98)) {
        throw new IllegalStateException("SHF_FLAG has not been filled");
      }

      if (!filled.get(78)) {
        throw new IllegalStateException("SOX has not been filled");
      }

      if (!filled.get(96)) {
        throw new IllegalStateException("SSTOREX has not been filled");
      }

      if (!filled.get(100)) {
        throw new IllegalStateException("STACK_ITEM_POP_1 has not been filled");
      }

      if (!filled.get(110)) {
        throw new IllegalStateException("STACK_ITEM_POP_2 has not been filled");
      }

      if (!filled.get(106)) {
        throw new IllegalStateException("STACK_ITEM_POP_3 has not been filled");
      }

      if (!filled.get(97)) {
        throw new IllegalStateException("STACK_ITEM_POP_4 has not been filled");
      }

      if (!filled.get(65)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_2 has not been filled");
      }

      if (!filled.get(64)) {
        throw new IllegalStateException("STACK_ITEM_VALUE_HI_3 has not been filled");
      }

      if (!filled.get(87)) {
        throw new IllegalStateException("STACKRAM_FLAG has not been filled");
      }

      if (!filled.get(68)) {
        throw new IllegalStateException(
            "STATIC_FLAG_xor_WARM_xor_HAS_CODE_NEW has not been filled");
      }

      if (!filled.get(66)) {
        throw new IllegalStateException(
            "STATICX_xor_UPDATE_xor_VAL_CURR_CHANGES_xor_IS_DEPLOYMENT_xor_EXISTS_NEW has not been filled");
      }

      if (!filled.get(85)) {
        throw new IllegalStateException("STO_FLAG has not been filled");
      }

      if (!filled.get(83)) {
        throw new IllegalStateException("SUX has not been filled");
      }

      if (!filled.get(89)) {
        throw new IllegalStateException("SWAP_FLAG has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("TRANSACTION_END_STAMP has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("TRANSACTION_REVERTS has not been filled");
      }

      if (!filled.get(74)) {
        throw new IllegalStateException(
            "TRM_FLAG_xor_VAL_CURR_IS_ZERO_xor_SUFFICIENT_BALANCE has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("TWO_LINE_INSTRUCTION has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("TX_EXEC has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("TX_FINL has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("TX_INIT has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("TX_SKIP has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("TX_WARM has not been filled");
      }

      if (!filled.get(101)) {
        throw new IllegalStateException("TXN_FLAG has not been filled");
      }

      if (!filled.get(102)) {
        throw new IllegalStateException("WCP_FLAG has not been filled");
      }

      filled.clear();

      return this;
    }

    TraceBuilder fillAndValidateRow() {
      if (!filled.get(41)) {
        abortFlag.add(false);
        this.filled.set(41);
      }
      if (!filled.get(4)) {
        absoluteTransactionNumber.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(71)) {
        accFlagXorValNextIsCurrXorWarm.add(false);
        this.filled.set(71);
      }
      if (!filled.get(88)) {
        addFlag.add(false);
        this.filled.set(88);
      }
      if (!filled.get(54)) {
        addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1.add(BigInteger.ZERO);
        this.filled.set(54);
      }
      if (!filled.get(43)) {
        addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi.add(
            BigInteger.ZERO);
        this.filled.set(43);
      }
      if (!filled.get(45)) {
        balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip.add(BigInteger.ZERO);
        this.filled.set(45);
      }
      if (!filled.get(52)) {
        balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee.add(
            BigInteger.ZERO);
        this.filled.set(52);
      }
      if (!filled.get(14)) {
        batchNumber.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(91)) {
        binFlag.add(false);
        this.filled.set(91);
      }
      if (!filled.get(94)) {
        btcFlag.add(false);
        this.filled.set(94);
      }
      if (!filled.get(63)) {
        byteCodeAddressLoXorStackItemStamp4.add(BigInteger.ZERO);
        this.filled.set(63);
      }
      if (!filled.get(90)) {
        callFlag.add(false);
        this.filled.set(90);
      }
      if (!filled.get(23)) {
        callerContextNumber.add(BigInteger.ZERO);
        this.filled.set(23);
      }
      if (!filled.get(61)) {
        callerContextNumberXorPushValueLo.add(BigInteger.ZERO);
        this.filled.set(61);
      }
      if (!filled.get(9)) {
        codeAddressHi.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(40)) {
        codeAddressLo.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(28)) {
        codeDeploymentNumber.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(26)) {
        codeDeploymentStatus.add(false);
        this.filled.set(26);
      }
      if (!filled.get(49)) {
        codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber
            .add(BigInteger.ZERO);
        this.filled.set(49);
      }
      if (!filled.get(53)) {
        codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce.add(BigInteger.ZERO);
        this.filled.set(53);
      }
      if (!filled.get(55)) {
        codeHashLoNewXorCallValueXorStackItemStamp1.add(BigInteger.ZERO);
        this.filled.set(55);
      }
      if (!filled.get(48)) {
        codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber.add(BigInteger.ZERO);
        this.filled.set(48);
      }
      if (!filled.get(42)) {
        codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue.add(BigInteger.ZERO);
        this.filled.set(42);
      }
      if (!filled.get(57)) {
        codeSizeXorCallDataOffsetXorStackItemStamp2.add(BigInteger.ZERO);
        this.filled.set(57);
      }
      if (!filled.get(112)) {
        conFlag.add(false);
        this.filled.set(112);
      }
      if (!filled.get(13)) {
        contextGetsRevrtdFlag.add(false);
        this.filled.set(13);
      }
      if (!filled.get(24)) {
        contextMayChangeFlag.add(false);
        this.filled.set(24);
      }
      if (!filled.get(20)) {
        contextNumber.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(3)) {
        contextNumberNew.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(30)) {
        contextRevertStamp.add(BigInteger.ZERO);
        this.filled.set(30);
      }
      if (!filled.get(0)) {
        contextSelfRevrtsFlag.add(false);
        this.filled.set(0);
      }
      if (!filled.get(12)) {
        contextWillRevertFlag.add(false);
        this.filled.set(12);
      }
      if (!filled.get(92)) {
        copyFlag.add(false);
        this.filled.set(92);
      }
      if (!filled.get(34)) {
        counterNsr.add(BigInteger.ZERO);
        this.filled.set(34);
      }
      if (!filled.get(18)) {
        counterTli.add(false);
        this.filled.set(18);
      }
      if (!filled.get(105)) {
        createFlag.add(false);
        this.filled.set(105);
      }
      if (!filled.get(81)) {
        decodedFlag1.add(false);
        this.filled.set(81);
      }
      if (!filled.get(69)) {
        decodedFlag2XorValOrigIsZeroXorHasCode.add(false);
        this.filled.set(69);
      }
      if (!filled.get(103)) {
        decodedFlag3.add(false);
        this.filled.set(103);
      }
      if (!filled.get(108)) {
        decodedFlag4.add(false);
        this.filled.set(108);
      }
      if (!filled.get(47)) {
        deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo
            .add(BigInteger.ZERO);
        this.filled.set(47);
      }
      if (!filled.get(51)) {
        deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi.add(
            BigInteger.ZERO);
        this.filled.set(51);
      }
      if (!filled.get(46)) {
        deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas.add(
            BigInteger.ZERO);
        this.filled.set(46);
      }
      if (!filled.get(56)) {
        deploymentStatusNewXorIsStaticXorInstruction.add(BigInteger.ZERO);
        this.filled.set(56);
      }
      if (!filled.get(58)) {
        deploymentStatusXorReturnAtSizeXorHeight.add(BigInteger.ZERO);
        this.filled.set(58);
      }
      if (!filled.get(73)) {
        dupFlagXorValNextIsOrigXorIsPrecompile.add(false);
        this.filled.set(73);
      }
      if (!filled.get(36)) {
        exceptionAhoyFlag.add(false);
        this.filled.set(36);
      }
      if (!filled.get(86)) {
        extFlag.add(false);
        this.filled.set(86);
      }
      if (!filled.get(37)) {
        failureConditionFlag.add(false);
        this.filled.set(37);
      }
      if (!filled.get(31)) {
        gasActual.add(BigInteger.ZERO);
        this.filled.set(31);
      }
      if (!filled.get(22)) {
        gasCost.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(15)) {
        gasExpected.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(8)) {
        gasMemoryExpansion.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(1)) {
        gasNext.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(21)) {
        gasRefund.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(99)) {
        haltFlag.add(false);
        this.filled.set(99);
      }
      if (!filled.get(2)) {
        hubStamp.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(93)) {
        invalidFlag.add(false);
        this.filled.set(93);
      }
      if (!filled.get(109)) {
        invprex.add(false);
        this.filled.set(109);
      }
      if (!filled.get(95)) {
        jumpFlag.add(false);
        this.filled.set(95);
      }
      if (!filled.get(80)) {
        jumpx.add(false);
        this.filled.set(80);
      }
      if (!filled.get(75)) {
        kecFlag.add(false);
        this.filled.set(75);
      }
      if (!filled.get(104)) {
        logFlag.add(false);
        this.filled.set(104);
      }
      if (!filled.get(70)) {
        maxcsxXorWarmNewXorDeploymentStatusInfty.add(false);
        this.filled.set(70);
      }
      if (!filled.get(111)) {
        modFlag.add(false);
        this.filled.set(111);
      }
      if (!filled.get(84)) {
        mulFlag.add(false);
        this.filled.set(84);
      }
      if (!filled.get(107)) {
        mxpFlag.add(false);
        this.filled.set(107);
      }
      if (!filled.get(79)) {
        mxpx.add(false);
        this.filled.set(79);
      }
      if (!filled.get(50)) {
        nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo.add(BigInteger.ZERO);
        this.filled.set(50);
      }
      if (!filled.get(44)) {
        nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee.add(BigInteger.ZERO);
        this.filled.set(44);
      }
      if (!filled.get(29)) {
        numberOfNonStackRows.add(BigInteger.ZERO);
        this.filled.set(29);
      }
      if (!filled.get(76)) {
        oobFlag.add(false);
        this.filled.set(76);
      }
      if (!filled.get(67)) {
        oogxXorValNextIsZeroXorExists.add(false);
        this.filled.set(67);
      }
      if (!filled.get(82)) {
        opcx.add(false);
        this.filled.set(82);
      }
      if (!filled.get(32)) {
        peekAtAccount.add(false);
        this.filled.set(32);
      }
      if (!filled.get(39)) {
        peekAtContext.add(false);
        this.filled.set(39);
      }
      if (!filled.get(35)) {
        peekAtStack.add(false);
        this.filled.set(35);
      }
      if (!filled.get(7)) {
        peekAtStorage.add(false);
        this.filled.set(7);
      }
      if (!filled.get(17)) {
        peekAtTransaction.add(false);
        this.filled.set(17);
      }
      if (!filled.get(38)) {
        programCounter.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(10)) {
        programCounterNew.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(72)) {
        pushpopFlagXorValCurrIsOrigXorWarmNew.add(false);
        this.filled.set(72);
      }
      if (!filled.get(77)) {
        rdcx.add(false);
        this.filled.set(77);
      }
      if (!filled.get(59)) {
        returnAtOffsetXorPushValueHi.add(BigInteger.ZERO);
        this.filled.set(59);
      }
      if (!filled.get(62)) {
        returnDataOffsetXorStackItemStamp3.add(BigInteger.ZERO);
        this.filled.set(62);
      }
      if (!filled.get(60)) {
        returnerIsPrecompileXorStackItemHeight4.add(BigInteger.ZERO);
        this.filled.set(60);
      }
      if (!filled.get(98)) {
        shfFlag.add(false);
        this.filled.set(98);
      }
      if (!filled.get(78)) {
        sox.add(false);
        this.filled.set(78);
      }
      if (!filled.get(96)) {
        sstorex.add(false);
        this.filled.set(96);
      }
      if (!filled.get(100)) {
        stackItemPop1.add(false);
        this.filled.set(100);
      }
      if (!filled.get(110)) {
        stackItemPop2.add(false);
        this.filled.set(110);
      }
      if (!filled.get(106)) {
        stackItemPop3.add(false);
        this.filled.set(106);
      }
      if (!filled.get(97)) {
        stackItemPop4.add(false);
        this.filled.set(97);
      }
      if (!filled.get(65)) {
        stackItemValueHi2.add(BigInteger.ZERO);
        this.filled.set(65);
      }
      if (!filled.get(64)) {
        stackItemValueHi3.add(BigInteger.ZERO);
        this.filled.set(64);
      }
      if (!filled.get(87)) {
        stackramFlag.add(false);
        this.filled.set(87);
      }
      if (!filled.get(68)) {
        staticFlagXorWarmXorHasCodeNew.add(false);
        this.filled.set(68);
      }
      if (!filled.get(66)) {
        staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew.add(false);
        this.filled.set(66);
      }
      if (!filled.get(85)) {
        stoFlag.add(false);
        this.filled.set(85);
      }
      if (!filled.get(83)) {
        sux.add(false);
        this.filled.set(83);
      }
      if (!filled.get(89)) {
        swapFlag.add(false);
        this.filled.set(89);
      }
      if (!filled.get(11)) {
        transactionEndStamp.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(33)) {
        transactionReverts.add(BigInteger.ZERO);
        this.filled.set(33);
      }
      if (!filled.get(74)) {
        trmFlagXorValCurrIsZeroXorSufficientBalance.add(false);
        this.filled.set(74);
      }
      if (!filled.get(27)) {
        twoLineInstruction.add(false);
        this.filled.set(27);
      }
      if (!filled.get(6)) {
        txExec.add(false);
        this.filled.set(6);
      }
      if (!filled.get(5)) {
        txFinl.add(false);
        this.filled.set(5);
      }
      if (!filled.get(16)) {
        txInit.add(false);
        this.filled.set(16);
      }
      if (!filled.get(19)) {
        txSkip.add(false);
        this.filled.set(19);
      }
      if (!filled.get(25)) {
        txWarm.add(false);
        this.filled.set(25);
      }
      if (!filled.get(101)) {
        txnFlag.add(false);
        this.filled.set(101);
      }
      if (!filled.get(102)) {
        wcpFlag.add(false);
        this.filled.set(102);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          abortFlag,
          absoluteTransactionNumber,
          accFlagXorValNextIsCurrXorWarm,
          addFlag,
          addressHiXorByteCodeDeploymentNumberXorStackItemValueLo1,
          addressLoXorValCurrLoXorCallerAddressHiXorStackItemHeight1XorFromAddressHi,
          balanceNewXorValCurrHiXorCallStackDepthXorStackItemValueLo2XorGasTip,
          balanceXorStorageKeyHiXorByteCodeAddressHiXorStackItemValueLo4XorGasFee,
          batchNumber,
          binFlag,
          btcFlag,
          byteCodeAddressLoXorStackItemStamp4,
          callFlag,
          callerContextNumber,
          callerContextNumberXorPushValueLo,
          codeAddressHi,
          codeAddressLo,
          codeDeploymentNumber,
          codeDeploymentStatus,
          codeHashHiNewXorAddressLoXorCallerAddressLoXorStackItemValueLo3XorAbsoluteTransactionNumber,
          codeHashHiXorAccountAddressHiXorStackItemHeight3XorNonce,
          codeHashLoNewXorCallValueXorStackItemStamp1,
          codeHashLoXorValOrigHiXorReturnDataSizeXorHeightUnderXorBatchNumber,
          codeSizeNewXorValNextLoXorReturnerContextNumberXorStaticGasXorValue,
          codeSizeXorCallDataOffsetXorStackItemStamp2,
          conFlag,
          contextGetsRevrtdFlag,
          contextMayChangeFlag,
          contextNumber,
          contextNumberNew,
          contextRevertStamp,
          contextSelfRevrtsFlag,
          contextWillRevertFlag,
          copyFlag,
          counterNsr,
          counterTli,
          createFlag,
          decodedFlag1,
          decodedFlag2XorValOrigIsZeroXorHasCode,
          decodedFlag3,
          decodedFlag4,
          deploymentNumberInftyXorDeploymentNumberXorByteCodeDeploymentStatusXorStackItemValueHi4XorFromAddressLo,
          deploymentNumberNewXorStorageKeyLoXorAccountAddressLoXorStackItemValueHi1XorToAddressHi,
          deploymentNumberXorAddressHiXorAccountDeploymentNumberXorHeightOverXorInitGas,
          deploymentStatusNewXorIsStaticXorInstruction,
          deploymentStatusXorReturnAtSizeXorHeight,
          dupFlagXorValNextIsOrigXorIsPrecompile,
          exceptionAhoyFlag,
          extFlag,
          failureConditionFlag,
          gasActual,
          gasCost,
          gasExpected,
          gasMemoryExpansion,
          gasNext,
          gasRefund,
          haltFlag,
          hubStamp,
          invalidFlag,
          invprex,
          jumpFlag,
          jumpx,
          kecFlag,
          logFlag,
          maxcsxXorWarmNewXorDeploymentStatusInfty,
          modFlag,
          mulFlag,
          mxpFlag,
          mxpx,
          nonceNewXorValNextHiXorContextNumberXorStackItemHeight2XorToAddressLo,
          nonceXorValOrigLoXorCallDataSizeXorHeightNewXorGasMaxfee,
          numberOfNonStackRows,
          oobFlag,
          oogxXorValNextIsZeroXorExists,
          opcx,
          peekAtAccount,
          peekAtContext,
          peekAtStack,
          peekAtStorage,
          peekAtTransaction,
          programCounter,
          programCounterNew,
          pushpopFlagXorValCurrIsOrigXorWarmNew,
          rdcx,
          returnAtOffsetXorPushValueHi,
          returnDataOffsetXorStackItemStamp3,
          returnerIsPrecompileXorStackItemHeight4,
          shfFlag,
          sox,
          sstorex,
          stackItemPop1,
          stackItemPop2,
          stackItemPop3,
          stackItemPop4,
          stackItemValueHi2,
          stackItemValueHi3,
          stackramFlag,
          staticFlagXorWarmXorHasCodeNew,
          staticxXorUpdateXorValCurrChangesXorIsDeploymentXorExistsNew,
          stoFlag,
          sux,
          swapFlag,
          transactionEndStamp,
          transactionReverts,
          trmFlagXorValCurrIsZeroXorSufficientBalance,
          twoLineInstruction,
          txExec,
          txFinl,
          txInit,
          txSkip,
          txWarm,
          txnFlag,
          wcpFlag);
    }
  }
}
