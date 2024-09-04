/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.module.hub.fragment.imc.mmu;

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.module.Util.slice;
import static net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeModexpDataOperation.MODEXP_COMPONENT_BYTE_SIZE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EMPTY_RIPEMD_HI;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EMPTY_RIPEMD_LO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EMPTY_SHA2_HI;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EMPTY_SHA2_LO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_BLAKEMODEXP;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_ECDATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_KEC;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_LOG;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_RIPSHA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_ROM;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXO_SUM_WEIGHT_TXCD;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_ANY_TO_RAM_WITH_PADDING;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_BLAKE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_EXO_TO_RAM_TRANSPLANTS;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_INVALID_CODE_PREFIX;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_MLOAD;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_MODEXP_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_MODEXP_ZERO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_MSTORE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_RAM_TO_EXO_WITH_PADDING;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_RAM_TO_RAM_SANS_PADDING;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_RIGHT_PADDED_WORD_EXTRACTION;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_BLAKE_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_BLAKE_PARAMS;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_BLAKE_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECADD_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECADD_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECMUL_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECMUL_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECPAIRING_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECPAIRING_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECRECOVER_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_ECRECOVER_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_BASE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_EXPONENT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_MODULUS;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.PHASE_MODEXP_RESULT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.RLP_TXN_PHASE_DATA;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WORD_SIZE;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileFlag.PRC_RIPEMD_160;
import static net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment.PrecompileFlag.PRC_SHA2_256;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.BASE_MIN_OFFSET;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.EBS_MIN_OFFSET;
import static net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata.MBS_MIN_OFFSET;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static org.apache.tuweni.bytes.Bytes.minimalBytes;

import java.util.Optional;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.State;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.opcode.CodeCopy;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.opcode.Create;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.opcode.Create2;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.opcode.ExtCodeCopy;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.opcode.ReturnFromDeployment;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection.EllipticCurvePrecompileSubsection;
import net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection.ModexpSubsection;
import net.consensys.linea.zktracer.module.hub.section.call.precompileSubsection.PrecompileSubsection;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.runtime.LogData;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.MemorySpan;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * This class represents a call to the MMU. However, some MMU calls may have their actual content
 * only definitely defined at tracing time, post conflation. In these cases, subclasses of this
 * class implement this defer mechanism.
 */
@RequiredArgsConstructor
@Setter
@Getter
@Accessors(fluent = true)
public class MmuCall implements TraceSubFragment, PostTransactionDefer {
  protected boolean traceMe = true;

  protected int instruction = 0;
  protected int sourceId = 0;
  protected int targetId = 0;
  protected int auxId = 0;
  protected EWord sourceOffset = EWord.ZERO;
  protected EWord targetOffset = EWord.ZERO;
  protected long size = 0;
  protected long referenceOffset = 0;
  protected long referenceSize = 0;
  protected boolean successBit = false;
  protected Bytes limb1 = Bytes.EMPTY;
  protected Bytes limb2 = Bytes.EMPTY;
  protected long phase = 0;

  private Optional<Bytes> sourceRamBytes = Optional.empty();
  private Optional<Bytes> targetRamBytes = Optional.empty();
  private Optional<Bytes> exoBytes = Optional.empty();

  protected boolean exoIsRlpTxn = false;
  protected boolean exoIsLog = false;
  protected boolean exoIsRom = false;
  protected boolean exoIsKec = false;
  protected boolean exoIsRipSha = false;
  protected boolean exoIsBlakeModexp = false;
  protected boolean exoIsEcData = false;
  private int exoSum = 0;

  public void dontTraceMe() {
    this.traceMe = false;
  }

  private MmuCall updateExoSum(final int exoValue) {
    this.exoSum += exoValue;
    return this;
  }

  final MmuCall setRlpTxn() {
    return this.exoIsRlpTxn(true).updateExoSum(EXO_SUM_WEIGHT_TXCD);
  }

  public final MmuCall setLog() {
    return this.exoIsLog(true).updateExoSum(EXO_SUM_WEIGHT_LOG);
  }

  public final void setRom() {
    this.exoIsRom(true).updateExoSum(EXO_SUM_WEIGHT_ROM);
  }

  public final MmuCall setKec() {
    return this.exoIsKec(true).updateExoSum(EXO_SUM_WEIGHT_KEC);
  }

  final MmuCall setRipSha() {
    return this.exoIsRipSha(true).updateExoSum(EXO_SUM_WEIGHT_RIPSHA);
  }

  final MmuCall setBlakeModexp() {
    return this.exoIsBlakeModexp(true).updateExoSum(EXO_SUM_WEIGHT_BLAKEMODEXP);
  }

  final MmuCall setEcData() {
    return this.exoIsEcData(true).updateExoSum(EXO_SUM_WEIGHT_ECDATA);
  }

  // TODO: make the instruction an enum
  public MmuCall(final Hub hub, final int instruction) {
    hub.defers().scheduleForPostTransaction(this);
    this.instruction = instruction;
  }

  public static MmuCall sha3(final Hub hub, final Bytes hashInput) {
    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .auxId(hub.state().stamps().hub())
        .exoBytes(Optional.of(hashInput))
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .referenceSize(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .setKec();
  }

  public static MmuCall callDataCopy(final Hub hub) {
    final MemorySpan callDataSegment = hub.currentFrame().callDataInfo().memorySpan();

    final int callDataContextNumber = callDataContextNumber(hub);
    final CallFrame callFrame = hub.callStack().getByContextNumber(callDataContextNumber);

    return new MmuCall(hub, MMU_INST_ANY_TO_RAM_WITH_PADDING)
        .sourceId(callDataContextNumber)
        .sourceRamBytes(Optional.of(callFrame.callDataInfo().data()))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(1)))
        .targetOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(2)))
        .referenceOffset(callDataSegment.offset())
        .referenceSize(callDataSegment.length());
  }

  public static int callDataContextNumber(final Hub hub) {
    final CallFrame currentFrame = hub.callStack().current();

    return currentFrame.isRoot()
        ? currentFrame.contextNumber() - 1
        : hub.callStack().parent().contextNumber();
  }

  public static MmuCall LogX(final Hub hub, final LogData logData) {
    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(logData.callFrame.contextNumber())
        .sourceRamBytes(Optional.of(logData.ramSourceBytes))
        .exoBytes(
            Optional.of(
                slice(
                    logData.ramSourceBytes,
                    (int) Words.clampedToLong(logData.offset),
                    (int) logData.size)))
        .sourceOffset(logData.offset)
        .size(logData.size)
        .referenceSize(logData.size)
        .setLog();

    // Note: The targetId() is set at the end of the transaction. We don't know now if the LOG will
    // be reverted or not

  }

  public static MmuCall codeCopy(final Hub hub) {
    return new CodeCopy(hub);
  }

  public static MmuCall extCodeCopy(final Hub hub) {
    return new ExtCodeCopy(hub);
  }

  public static MmuCall returnDataCopy(final Hub hub) {
    final MemorySpan returnDataSegment = hub.currentFrame().returnDataSpan();
    final CallFrame returnerFrame =
        hub.callStack().getByContextNumber(hub.currentFrame().returnDataContextNumber());

    return new MmuCall(hub, MMU_INST_ANY_TO_RAM_WITH_PADDING)
        .sourceId(returnerFrame.contextNumber())
        .sourceRamBytes(
            Optional.of(
                returnerFrame.frame().shadowReadMemory(0, returnerFrame.frame().memoryByteSize())))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(1)))
        .targetOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(2)))
        .referenceOffset(returnDataSegment.offset())
        .referenceSize(returnDataSegment.length());
  }

  public static MmuCall create(final Hub hub) {
    return new Create(hub);
  }

  public static MmuCall returnFromDeployment(final Hub hub) {
    return new ReturnFromDeployment(hub);
  }

  public static MmuCall returnFromMessageCall(final Hub hub) {
    return MmuCall.revert(hub);
  }

  public static MmuCall create2(final Hub hub, boolean failureCondition) {
    return new Create2(hub, failureCondition);
  }

  public static MmuCall invalidCodePrefix(final Hub hub) {
    final short currentExceptions = hub.pch().exceptions();

    return new MmuCall(hub, MMU_INST_INVALID_CODE_PREFIX)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .successBit(!Exceptions.invalidCodePrefix(currentExceptions));
  }

  public static MmuCall revert(final Hub hub) {
    final CallFrame parentFrame = hub.callStack().parent();

    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .targetId(parentFrame.contextNumber())
        .targetRamBytes(
            Optional.of(
                parentFrame
                    .frame()
                    .shadowReadMemory(0, hub.callStack().parent().frame().memoryByteSize())))
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .referenceOffset(hub.currentFrame().returnDataTargetInCaller().offset())
        .referenceSize(hub.currentFrame().returnDataTargetInCaller().length());
  }

  public static MmuCall txInit(final Hub hub) {
    return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(hub.txStack().current().getAbsoluteTransactionNumber())
        .exoBytes(Optional.of(hub.txStack().current().getBesuTransaction().getPayload()))
        .targetId(hub.stamp())
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .size(hub.txStack().current().getBesuTransaction().getData().map(Bytes::size).orElse(0))
        .phase(RLP_TXN_PHASE_DATA)
        .setRlpTxn();
  }

  public static MmuCall callDataExtractionForEcrecover(
      final Hub hub, EllipticCurvePrecompileSubsection subsection, boolean successfulRecovery) {

    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber()) // called at ContextReEntry
        .sourceRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetId(subsection.exoModuleOperationId())
        .exoBytes(Optional.of(subsection.callData))
        .sourceOffset(EWord.of(subsection.callDataMemorySpan.offset()))
        .size(subsection.callDataMemorySpan.length())
        .referenceSize(128)
        .successBit(successfulRecovery)
        .phase(PHASE_ECRECOVER_DATA)
        .setEcData();
  }

  public static MmuCall fullReturnDataTransferForEcrecover(
      final Hub hub, EllipticCurvePrecompileSubsection subsection) {

    final int precompileContextNumber = subsection.exoModuleOperationId();

    return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(precompileContextNumber)
        .exoBytes(Optional.of(subsection.returnData()))
        .targetId(precompileContextNumber)
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .size(WORD_SIZE)
        .phase(PHASE_ECRECOVER_RESULT)
        .setEcData();
  }

  public static MmuCall partialReturnDataCopyForEcrecover(
      final Hub hub, EllipticCurvePrecompileSubsection subsection) {

    final int precompileContextNumber = subsection.exoModuleOperationId();

    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(precompileContextNumber)
        .sourceRamBytes(Optional.of(subsection.returnData))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .sourceOffset(EWord.ZERO)
        .size(WORD_SIZE)
        .referenceOffset(subsection.parentReturnDataTarget.offset())
        .referenceSize(subsection.parentReturnDataTarget.length());
  }

  public static MmuCall callDataExtractionForShaTwoAndRipemd(
      final Hub hub, PrecompileSubsection precompileSubsection) {

    final PrecompileScenarioFragment.PrecompileFlag flag =
        precompileSubsection.precompileScenarioFragment().flag;
    checkArgument(flag.isAnyOf(PRC_SHA2_256, PRC_RIPEMD_160));

    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot()))
        .targetId(precompileSubsection.exoModuleOperationId())
        .exoBytes(Optional.of(precompileSubsection.callData))
        .sourceOffset(EWord.of(precompileSubsection.callDataMemorySpan.offset()))
        .size(precompileSubsection.callDataMemorySpan.length())
        .referenceSize(precompileSubsection.callDataMemorySpan.length())
        .phase(flag.dataPhase())
        .setRipSha();
  }

  public static MmuCall fullResultTransferForShaTwoAndRipemd(
      final Hub hub, PrecompileSubsection precompileSubsection) {

    final PrecompileScenarioFragment.PrecompileFlag flag =
        precompileSubsection.precompileScenarioFragment().flag;
    checkArgument(flag.isAnyOf(PRC_SHA2_256, PRC_RIPEMD_160));

    final boolean isShaTwo = flag == PRC_SHA2_256;

    if (precompileSubsection.callDataMemorySpan.isEmpty()) {
      return new MmuCall(hub, MMU_INST_MSTORE)
          .targetId(precompileSubsection.exoModuleOperationId())
          .targetOffset(EWord.ZERO)
          .limb1(isShaTwo ? bigIntegerToBytes(EMPTY_SHA2_HI) : minimalBytes(EMPTY_RIPEMD_HI))
          .limb2(isShaTwo ? bigIntegerToBytes(EMPTY_SHA2_LO) : bigIntegerToBytes(EMPTY_RIPEMD_LO));
    } else {
      return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
          .sourceId(precompileSubsection.exoModuleOperationId())
          .exoBytes(Optional.of(precompileSubsection.returnData))
          .targetId(precompileSubsection.returnDataContextNumber())
          .targetRamBytes(Optional.of(Bytes.EMPTY))
          .size(WORD_SIZE)
          .phase(flag.resultPhase())
          .setRipSha();
    }
  }

  public static MmuCall partialReturnDataCopyForShaTwoAndRipemd(
      final Hub hub, PrecompileSubsection precompileSubsection) {

    final PrecompileScenarioFragment.PrecompileFlag flag =
        precompileSubsection.precompileScenarioFragment().flag;

    checkArgument(flag.isAnyOf(PRC_SHA2_256, PRC_RIPEMD_160));
    checkArgument(!precompileSubsection.parentReturnDataTarget.isEmpty());

    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(precompileSubsection.returnDataContextNumber())
        .sourceRamBytes(Optional.of(precompileSubsection.returnData))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot))
        .sourceOffset(EWord.ZERO)
        .size(WORD_SIZE)
        .referenceOffset(precompileSubsection.parentReturnDataTarget.offset())
        .referenceSize(precompileSubsection.parentReturnDataTarget.length());
  }

  public static MmuCall forIdentityExtractCallData(
      final Hub hub, PrecompileSubsection precompileSubsection) {

    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(precompileSubsection.callSection.hubStamp())
        .sourceRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot))
        .targetId(precompileSubsection.exoModuleOperationId())
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .sourceOffset(EWord.of(precompileSubsection.callDataMemorySpan.offset()))
        .size(precompileSubsection.callDataMemorySpan.length())
        .referenceSize(precompileSubsection.callDataMemorySpan.length());
  }

  public static MmuCall forIdentityReturnData(
      final Hub hub, final PrecompileSubsection precompileSubsection) {

    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(precompileSubsection.exoModuleOperationId())
        .sourceRamBytes(Optional.of(precompileSubsection.returnData()))
        .targetId(precompileSubsection.callSection.hubStamp())
        .targetRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot))
        .sourceOffset(EWord.ZERO)
        .targetOffset(EWord.of(precompileSubsection.parentReturnDataTarget.offset()))
        .size(precompileSubsection.parentReturnDataTarget.length())
        .referenceSize(precompileSubsection.parentReturnDataTarget.length());
  }

  public static MmuCall callDataExtractionForEcadd(
      final Hub hub, PrecompileSubsection subsection, boolean failureKnownToRam) {
    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetId(subsection.exoModuleOperationId())
        .exoBytes(Optional.of(subsection.callData))
        .sourceOffset(EWord.of(subsection.callDataMemorySpan.offset()))
        .size(subsection.callDataMemorySpan.length())
        .referenceSize(128)
        .successBit(!failureKnownToRam)
        .setEcData()
        .phase(PHASE_ECADD_DATA);
  }

  public static MmuCall fullReturnDataTransferForEcadd(
      final Hub hub, PrecompileSubsection subsection) {
    return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(subsection.exoModuleOperationId())
        .exoBytes(Optional.of(subsection.returnData()))
        .targetId(subsection.exoModuleOperationId())
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .size(64)
        .setEcData()
        .phase(PHASE_ECADD_RESULT);
  }

  public static MmuCall partialCopyOfReturnDataForEcadd(
      final Hub hub, PrecompileSubsection subsection) {
    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(subsection.exoModuleOperationId())
        .sourceRamBytes(Optional.of(subsection.returnData()))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetOffset(EWord.of(subsection.parentReturnDataTarget.offset()))
        .size(subsection.parentReturnDataTarget.length())
        .referenceSize(64);
  }

  public static MmuCall callDataExtractionForEcmul(
      final Hub hub, final PrecompileSubsection subsection, boolean failureKnownToRam) {
    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetId(subsection.exoModuleOperationId())
        .exoBytes(Optional.of(subsection.callData))
        .sourceOffset(EWord.of(subsection.callDataMemorySpan.offset()))
        .size(subsection.callDataMemorySpan.length())
        .referenceSize(96)
        .successBit(!failureKnownToRam)
        .setEcData()
        .phase(PHASE_ECMUL_DATA);
  }

  public static MmuCall fullReturnDataTransferForEcmul(
      final Hub hub, final PrecompileSubsection subsection) {
    return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(subsection.exoModuleOperationId())
        .exoBytes(Optional.of(subsection.returnData()))
        .targetId(subsection.exoModuleOperationId())
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .size(64)
        .setEcData()
        .phase(PHASE_ECMUL_RESULT);
  }

  public static MmuCall partialCopyOfReturnDataForEcmul(
      final Hub hub, final PrecompileSubsection subsection) {
    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(subsection.exoModuleOperationId())
        .sourceRamBytes(Optional.of(subsection.returnData()))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetOffset(EWord.of(subsection.parentReturnDataTarget().offset()))
        .size(subsection.parentReturnDataTarget().length())
        .referenceSize(64);
  }

  public static MmuCall callDataExtractionForEcpairing(
      final Hub hub, PrecompileSubsection subsection, boolean failureKnownToRam) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetId(precompileContextNumber)
        .exoBytes(Optional.of(subsection.callData))
        .sourceOffset(EWord.of(subsection.callDataMemorySpan.offset()))
        .size(subsection.callDataMemorySpan.length())
        .referenceSize(subsection.callDataMemorySpan.length())
        .successBit(!failureKnownToRam)
        .setEcData()
        .phase(PHASE_ECPAIRING_DATA);
  }

  /**
   * Note that {@link MmuCall#fullReturnDataTransferForEcpairing} handles both cases of interest:
   * empty call data and nonempty call data.
   */
  public static MmuCall fullReturnDataTransferForEcpairing(
      final Hub hub, PrecompileSubsection subsection) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    if (subsection.callDataMemorySpan.isEmpty()) {
      return new MmuCall(hub, MMU_INST_MSTORE).targetId(precompileContextNumber).limb2(Bytes.of(1));
    } else {
      return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
          .sourceId(precompileContextNumber)
          .exoBytes(Optional.of(subsection.returnData()))
          .targetId(precompileContextNumber)
          .targetRamBytes(Optional.of(Bytes.EMPTY))
          .size(WORD_SIZE)
          .setEcData()
          .phase(PHASE_ECPAIRING_RESULT);
    }
  }

  public static MmuCall partialCopyOfReturnDataForEcpairing(
      final Hub hub, PrecompileSubsection subsection) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(precompileContextNumber)
        .sourceRamBytes(Optional.of(subsection.returnData()))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetOffset(EWord.of(subsection.parentReturnDataTarget.offset()))
        .size(subsection.parentReturnDataTarget.length())
        .referenceSize(32);
  }

  public static MmuCall parameterExtractionForBlake(
      final Hub hub,
      PrecompileSubsection subsection,
      boolean blakeSuccess,
      Bytes blakeR,
      Bytes blakeF) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    return new MmuCall(hub, MMU_INST_BLAKE)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetId(precompileContextNumber)
        .sourceOffset(EWord.of(subsection.callDataMemorySpan.offset()))
        .successBit(blakeSuccess)
        .limb1(blakeR)
        .limb2(blakeF)
        .setBlakeModexp()
        .phase(PHASE_BLAKE_PARAMS);
  }

  public static MmuCall callDataExtractionforBlake(final Hub hub, PrecompileSubsection subsection) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    return new MmuCall(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .targetId(precompileContextNumber)
        .exoBytes(Optional.of(subsection.callData))
        .sourceOffset(EWord.of(subsection.callDataMemorySpan.offset() + 4))
        .size(208)
        .referenceSize(208)
        .setBlakeModexp()
        .phase(PHASE_BLAKE_DATA);
  }

  public static MmuCall fullReturnDataTransferForBlake(
      final Hub hub, PrecompileSubsection subsection) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(precompileContextNumber)
        .exoBytes(Optional.of(subsection.returnData()))
        .targetId(precompileContextNumber)
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .size(64)
        .setBlakeModexp()
        .phase(PHASE_BLAKE_RESULT);
  }

  public static MmuCall partialCopyOfReturnDataforBlake(
      final Hub hub, PrecompileSubsection subsection) {
    final int precompileContextNumber = subsection.exoModuleOperationId();
    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(precompileContextNumber)
        .sourceRamBytes(Optional.of(subsection.returnData()))
        .targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(Optional.of(subsection.callerMemorySnapshot()))
        .size(64)
        .referenceOffset(subsection.parentReturnDataTarget.offset())
        .referenceSize(subsection.parentReturnDataTarget.length());
  }

  public static MmuCall forModexpExtractBbs(
      final Hub hub, final ModexpSubsection precompileSubsection, final ModexpMetadata metaData) {
    return new MmuCall(hub, MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot()))
        .referenceOffset(precompileSubsection.callDataMemorySpan.offset())
        .referenceSize(precompileSubsection.callDataMemorySpan.length())
        .limb1(metaData.bbs().hi())
        .limb2(metaData.bbs().lo());
  }

  public static MmuCall forModexpExtractEbs(
      final Hub hub, final ModexpSubsection precompileSubsection, final ModexpMetadata metaData) {
    return new MmuCall(hub, MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot()))
        .sourceOffset(EWord.of(EBS_MIN_OFFSET))
        .referenceOffset(precompileSubsection.callDataMemorySpan.offset())
        .referenceSize(precompileSubsection.callDataMemorySpan.length())
        .limb1(metaData.ebs().hi())
        .limb2(metaData.ebs().lo());
  }

  public static MmuCall forModexpExtractMbs(
      final Hub hub, final ModexpSubsection precompileSubsection, final ModexpMetadata metaData) {
    return new MmuCall(hub, MMU_INST_RIGHT_PADDED_WORD_EXTRACTION)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot()))
        .sourceOffset(EWord.of(MBS_MIN_OFFSET))
        .referenceOffset(precompileSubsection.callDataMemorySpan.offset())
        .referenceSize(precompileSubsection.callDataMemorySpan.length())
        .limb1(metaData.mbs().hi())
        .limb2(metaData.mbs().lo());
  }

  public static MmuCall forModexpLoadLead(
      final Hub hub, final ModexpSubsection precompileSubsection, final ModexpMetadata metaData) {
    return new MmuCall(hub, MMU_INST_MLOAD)
        .sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(Optional.of(precompileSubsection.callerMemorySnapshot()))
        .sourceOffset(
            EWord.of(
                precompileSubsection.callDataMemorySpan.offset() + 96 + metaData.bbs().toInt()))
        .limb1(metaData.rawLeadingWord().hi())
        .limb2(metaData.rawLeadingWord().lo());
  }

  public static MmuCall forModexpExtractBase(
      final Hub hub, final ModexpSubsection modexpSubsection, final ModexpMetadata modExpMetadata) {
    if (modExpMetadata.extractBase()) {
      return new MmuCall(hub, MMU_INST_MODEXP_DATA)
          .sourceId(modexpSubsection.callSection.hubStamp())
          .sourceRamBytes(Optional.of(modexpSubsection.callerMemorySnapshot))
          .targetId(modexpSubsection.exoModuleOperationId())
          .exoBytes(Optional.of(modExpMetadata.base()))
          .sourceOffset(EWord.of(BASE_MIN_OFFSET))
          .size(modExpMetadata.bbs().toInt())
          .referenceOffset(modexpSubsection.callDataMemorySpan.offset())
          .referenceSize(modexpSubsection.callDataMemorySpan.length())
          .phase(PHASE_MODEXP_BASE)
          .setBlakeModexp();
    } else {
      return new MmuCall(hub, MMU_INST_MODEXP_ZERO)
          .targetId(modexpSubsection.exoModuleOperationId())
          .phase(PHASE_MODEXP_BASE)
          .setBlakeModexp();
    }
  }

  public static MmuCall forModexpExtractExponent(
      final Hub hub, final ModexpSubsection modexpSubsection, final ModexpMetadata modExpMetadata) {
    if (modExpMetadata.extractExponent()) {
      return new MmuCall(hub, MMU_INST_MODEXP_DATA)
          .sourceId(modexpSubsection.callSection.hubStamp())
          .sourceRamBytes(Optional.of(modexpSubsection.callerMemorySnapshot))
          .targetId(modexpSubsection.exoModuleOperationId())
          .exoBytes(Optional.of(modExpMetadata.exp()))
          .sourceOffset(EWord.of(BASE_MIN_OFFSET + modExpMetadata.bbs().toInt()))
          .size(modExpMetadata.ebs().toInt())
          .referenceOffset(modexpSubsection.callDataMemorySpan.offset())
          .referenceSize(modexpSubsection.callDataMemorySpan.length())
          .phase(PHASE_MODEXP_EXPONENT)
          .setBlakeModexp();
    } else {
      return new MmuCall(hub, MMU_INST_MODEXP_ZERO)
          .targetId(modexpSubsection.exoModuleOperationId())
          .phase(PHASE_MODEXP_EXPONENT)
          .setBlakeModexp();
    }
  }

  public static MmuCall forModexpExtractModulus(
      final Hub hub, final ModexpSubsection modexpSubsection, final ModexpMetadata modExpMetadata) {
    return new MmuCall(hub, MMU_INST_MODEXP_DATA)
        .sourceId(modexpSubsection.callSection.hubStamp())
        .sourceRamBytes(Optional.of(modexpSubsection.callerMemorySnapshot))
        .targetId(modexpSubsection.exoModuleOperationId())
        .exoBytes(Optional.of(modExpMetadata.mod()))
        .sourceOffset(
            EWord.of(BASE_MIN_OFFSET + modExpMetadata.bbs().toInt() + modExpMetadata.ebs().toInt()))
        .size(modExpMetadata.mbs().toInt())
        .referenceOffset(modexpSubsection.callDataMemorySpan.offset())
        .referenceSize(modexpSubsection.callDataMemorySpan.length())
        .phase(PHASE_MODEXP_MODULUS)
        .setBlakeModexp();
  }

  public static MmuCall forModexpFullResultCopy(
      final Hub hub, final ModexpSubsection modexpSubsection, final ModexpMetadata modExpMetadata) {
    return new MmuCall(hub, MMU_INST_EXO_TO_RAM_TRANSPLANTS)
        .sourceId(modexpSubsection.exoModuleOperationId())
        .exoBytes(Optional.of(modexpSubsection.returnData()))
        .targetId(modexpSubsection.returnDataContextNumber())
        .targetRamBytes(Optional.of(Bytes.EMPTY))
        .size(MODEXP_COMPONENT_BYTE_SIZE)
        .phase(PHASE_MODEXP_RESULT)
        .setBlakeModexp();
  }

  public static MmuCall forModexpPartialResultCopy(
      final Hub hub, final ModexpSubsection modexpSubsection, final ModexpMetadata modExpMetadata) {
    return new MmuCall(hub, MMU_INST_RAM_TO_RAM_SANS_PADDING)
        .sourceId(modexpSubsection.exoModuleOperationId())
        .sourceRamBytes(Optional.of(modexpSubsection.returnData()))
        .targetId(modexpSubsection.callSection.hubStamp())
        .targetRamBytes(Optional.of(modexpSubsection.callerMemorySnapshot))
        .sourceOffset(EWord.of(MODEXP_COMPONENT_BYTE_SIZE - modExpMetadata.mbs().toInt()))
        .size(modExpMetadata.mbs().toInt())
        .referenceOffset(modexpSubsection.parentReturnDataTarget.offset())
        .referenceSize(modexpSubsection.parentReturnDataTarget.length());
  }

  @Override
  public Trace trace(Trace trace, State.TxState.Stamps stamps) {
    if (traceMe) {
      stamps.incrementMmuStamp();
      return trace
          .pMiscMmuFlag(true)
          .pMiscMmuInst(instruction)
          .pMiscMmuTgtId(targetId)
          .pMiscMmuSrcId(sourceId)
          .pMiscMmuAuxId(auxId)
          .pMiscMmuSrcOffsetHi(sourceOffset.hi())
          .pMiscMmuSrcOffsetLo(sourceOffset.lo())
          .pMiscMmuTgtOffsetLo(targetOffset.lo())
          .pMiscMmuSize(size)
          .pMiscMmuRefOffset(referenceOffset)
          .pMiscMmuRefSize(referenceSize)
          .pMiscMmuSuccessBit(successBit)
          .pMiscMmuLimb1(limb1)
          .pMiscMmuLimb2(limb2)
          .pMiscMmuExoSum(exoSum)
          .pMiscMmuPhase(phase);
    } else {
      return trace;
    }
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    if (traceMe) {
      hub.mmu().call(this);
    }
  }
}
