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

package net.consensys.linea.zktracer.module.mxp.moduleOperation;

import static net.consensys.linea.zktracer.module.mxp.MxpUtils.*;
import static net.consensys.linea.zktracer.types.Conversions.bytesToLong;

import lombok.Getter;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.mxp.moduleCall.*;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@Getter
public class CancunMxpOperation extends MxpOperation {

  private final int contextNumber;
  private final CancunMxpCall cancunMxpCall;

  public CancunMxpOperation(final MxpCall mxpCall) {
    super(mxpCall);
    // Setting of global variables
    this.contextNumber = mxpCall.hub.currentFrame().contextNumber();

    this.cancunMxpCall = (CancunMxpCall) mxpCall;
  }

  private int nRowsComputation() {
    return cancunMxpCall.ctMax() + 1;
  }

  @Override
  protected int computeLineCount() {
    return nRowsComputation() + 3; // 3 for decoder, macro and scenario
  }

  @Override
  public final void trace(int stamp, Trace.Mxp trace) {
    traceDecoder(stamp, trace);
    traceMacro(stamp, trace);
    traceScenario(stamp, trace);
    traceComputation(stamp, trace);
  }

  final void traceDecoder(int stamp, Trace.Mxp trace) {
    OpCodeData opCodeData = cancunMxpCall.getOpCodeData();

    trace
        .mxpStamp(stamp)
        .cn(Bytes.ofUnsignedLong(this.getContextNumber()))
        .decoder(true)
        .pDecoderInst(UnsignedByte.of(opCodeData.mnemonic().byteValue()))
        .pDecoderIsMsize(opCodeData.isMSize())
        .pDecoderIsReturn(opCodeData.isReturn())
        .pDecoderIsMcopy(opCodeData.isMCopy())
        .pDecoderIsFixedSize32(opCodeData.isFixedSize32())
        .pDecoderIsFixedSize1(opCodeData.isFixedSize1())
        .pDecoderIsSingleMaxOffset(opCodeData.isSingleOffset())
        .pDecoderIsDoubleMaxOffset(opCodeData.isDoubleOffset())
        .pDecoderIsWordPricing(opCodeData.isWordPricing())
        .pDecoderIsBytePricing(opCodeData.isBytePricing())
        .pDecoderGword(cancunMxpCall.gWord)
        .pDecoderGbyte(cancunMxpCall.gByte)
        .fillAndValidateRow();
  }

  final void traceMacro(int stamp, Trace.Mxp trace) {
    OpCode opCode = cancunMxpCall.getOpCodeData().mnemonic();

    trace
        .mxpStamp(stamp)
        .cn(Bytes.ofUnsignedLong(this.getContextNumber()))
        .macro(true)
        .pMacroInst(UnsignedByte.of(opCode.byteValue()))
        .pMacroDeploying(cancunMxpCall.isDeploys())
        .pMacroOffset1Hi(cancunMxpCall.getOffset1().hi())
        .pMacroOffset1Lo(cancunMxpCall.getOffset1().lo())
        .pMacroSize1Hi(cancunMxpCall.getSize1().hi())
        .pMacroSize1Lo(cancunMxpCall.getSize1().lo())
        .pMacroOffset2Hi(cancunMxpCall.getOffset2().hi())
        .pMacroOffset2Lo(cancunMxpCall.getOffset2().lo())
        .pMacroSize2Hi(cancunMxpCall.getSize2().hi())
        .pMacroSize2Lo(cancunMxpCall.getSize2().lo())
        .pMacroRes(cancunMxpCall.isMSizeScenario() ? cancunMxpCall.getMemorySizeInWords() : 0L)
        .pMacroMxpx(cancunMxpCall.isMxpx())
        .pMacroGasMxp(Bytes.ofUnsignedLong(cancunMxpCall.getGasMxp()))
        .pMacroS1Nznomxpx(!cancunMxpCall.getSize1().isZero() && !cancunMxpCall.isMxpx())
        .pMacroS2Nznomxpx(!cancunMxpCall.getSize2().isZero() && !cancunMxpCall.isMxpx())
        .fillAndValidateRow();
  }

  final void traceScenario(int stamp, Trace.Mxp trace) {
    trace
        .mxpStamp(stamp)
        .cn(Bytes.ofUnsignedLong(this.getContextNumber()))
        .scenario(true)
        .pScenarioMsize(cancunMxpCall.isMSizeScenario())
        .pScenarioTrivial(cancunMxpCall.isTrivialScenario())
        .pScenarioMxpx(cancunMxpCall.isMxpxScenario())
        .pScenarioStateUpdateWordPricing(cancunMxpCall.isStateUpdateWordPricingScenario())
        .pScenarioStateUpdateBytePricing(cancunMxpCall.isStateUpdateBytePricingScenario())
        .pScenarioWords(cancunMxpCall.words)
        .pScenarioWordsNew(cancunMxpCall.wordsNew)
        .pScenarioCmem(Bytes.ofUnsignedLong(cancunMxpCall.cMem))
        .pScenarioCmemNew(Bytes.ofUnsignedLong(cancunMxpCall.cMemNew))
        .fillAndValidateRow();
  }

  final void traceComputation(int stamp, Trace.Mxp trace) {

    for (int i = 0; i < nRowsComputation(); i++) {
      trace
          .mxpStamp(stamp)
          .cn(Bytes.ofUnsignedLong(this.getContextNumber()))
          .computation(true)
          .ct(i)
          .ctMax(cancunMxpCall.ctMax())
          .pComputationWcpFlag(cancunMxpCall.exoCalls[i].wcpFlag())
          .pComputationEucFlag(cancunMxpCall.exoCalls[i].eucFlag())
          .pComputationExoInst(cancunMxpCall.exoCalls[i].instruction())
          .pComputationArg1Hi(cancunMxpCall.exoCalls[i].arg1Hi())
          .pComputationArg1Lo(cancunMxpCall.exoCalls[i].arg1Lo())
          .pComputationArg2Hi(cancunMxpCall.exoCalls[i].arg2Hi())
          .pComputationArg2Lo(cancunMxpCall.exoCalls[i].arg2Lo())
          .pComputationResA(bytesToLong(cancunMxpCall.exoCalls[i].resultA()))
          .pComputationResB(bytesToLong(cancunMxpCall.exoCalls[i].resultB()))
          .fillAndValidateRow();
    }
  }
}
