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

package net.consensys.linea.zktracer.module.exp;

import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_CMPTN_EXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_CMPTN_MODEXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_PRPRC_EXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_PRPRC_MODEXP_LOG;

import lombok.Getter;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@Getter
public abstract class ExpChunk extends ModuleOperation {
  protected short pComputationPltJmp = 0;
  protected Bytes pComputationRawAcc; // (last row) paired with RawByte
  protected Bytes pComputationTrimAcc = Bytes.EMPTY; // (last row) paired with TrimByte
  protected UnsignedByte pComputationMsb = UnsignedByte.ZERO;

  protected int pMacroExpInst;
  protected Bytes pMacroData1 = Bytes.EMPTY;
  protected Bytes pMacroData2 = Bytes.EMPTY;
  protected Bytes pMacroData3 = Bytes.EMPTY;
  protected Bytes pMacroData4 = Bytes.EMPTY;
  protected Bytes pMacroData5 = Bytes.EMPTY;

  protected boolean[] pPreprocessingWcpFlag;
  protected Bytes[] pPreprocessingWcpArg1Hi;
  protected Bytes[] pPreprocessingWcpArg1Lo;
  protected Bytes[] pPreprocessingWcpArg2Hi;
  protected Bytes[] pPreprocessingWcpArg2Lo;
  protected UnsignedByte[] pPreprocessingWcpInst;
  protected boolean[] pPreprocessingWcpRes;

  protected Wcp wcp;

  protected abstract boolean isExpLog();

  protected void initArrays(int pPreprocessingLen) {
    pPreprocessingWcpFlag = new boolean[pPreprocessingLen];
    pPreprocessingWcpArg1Hi = new Bytes[pPreprocessingLen];
    pPreprocessingWcpArg1Lo = new Bytes[pPreprocessingLen];
    pPreprocessingWcpArg2Hi = new Bytes[pPreprocessingLen];
    pPreprocessingWcpArg2Lo = new Bytes[pPreprocessingLen];
    pPreprocessingWcpInst = new UnsignedByte[pPreprocessingLen];
    pPreprocessingWcpRes = new boolean[pPreprocessingLen];
  }

  @Override
  protected int computeLineCount() {
    // We assume MAX_CT_MACRO_EXP_LOG = MAX_CT_MACRO_MODEXP_LOG = 0;
    if (this.isExpLog()) {
      return MAX_CT_CMPTN_EXP_LOG + MAX_CT_PRPRC_EXP_LOG + 3;
    }

    return MAX_CT_CMPTN_MODEXP_LOG + MAX_CT_PRPRC_MODEXP_LOG + 3;
  }

  public abstract void preCompute();

  final void traceComputation(int stamp, Trace trace) {
    boolean tanzb;
    short pComputationTanzbAcc = 0; // Paired with Tanzb
    boolean manzb;
    short pComputationManzbAcc = 0; // Paired with Manzb
    short maxCt = (short) (isExpLog() ? MAX_CT_CMPTN_EXP_LOG : MAX_CT_CMPTN_MODEXP_LOG);

    for (short i = 0; i < maxCt + 1; i++) {
      /*
      All the values are derived from
      isExpLog
      pComputationPltJmp
      pComputationRawAcc
      pComputationTrimAcc
      pComputationMsb
      */
      // tanzb turns to 1 iff trimAcc is nonzero
      tanzb = pComputationTrimAcc.slice(0, i + 1).toBigInteger().signum() != 0;
      pComputationTanzbAcc += (short) (tanzb ? 1 : 0);
      // manzb turns to 1 iff msbAcc is nonzero
      manzb = i > maxCt - 8 && pComputationMsb.slice(0, i % 8 + 1) != 0;
      pComputationManzbAcc += (short) (manzb ? 1 : 0);
      trace
          .cmptn(true)
          .stamp(stamp)
          .ct(i)
          .ctMax(maxCt)
          .isExpLog(isExpLog())
          .isModexpLog(!isExpLog())
          .pComputationPltBit(i >= pComputationPltJmp)
          .pComputationPltJmp(pComputationPltJmp)
          .pComputationRawByte(UnsignedByte.of(pComputationRawAcc.get(i)))
          .pComputationRawAcc(pComputationRawAcc.slice(0, i + 1))
          .pComputationTrimByte(UnsignedByte.of(pComputationTrimAcc.get(i)))
          .pComputationTrimAcc(pComputationTrimAcc.slice(0, i + 1))
          .pComputationTanzb(tanzb)
          .pComputationTanzbAcc(pComputationTanzbAcc)
          .pComputationMsb(pComputationMsb)
          .pComputationMsbBit(i > maxCt - 8 && pComputationMsb.get(i % 8))
          .pComputationMsbAcc(
              UnsignedByte.of(i > maxCt - 8 ? pComputationMsb.slice(0, i % 8 + 1) : 0))
          .pComputationManzb(manzb)
          .pComputationManzbAcc(pComputationManzbAcc)
          .fillAndValidateRow();
    }
  }

  final void traceMacro(int stamp, Trace trace) {
    // We assume MAX_CT_MACRO_EXP_LOG = MAX_CT_MACRO_MODEXP_LOG = 0;
    trace
        .macro(true)
        .stamp(stamp)
        .ct((short) 0)
        .ctMax((short) 0)
        .isExpLog(isExpLog())
        .isModexpLog(!isExpLog())
        .pMacroExpInst(pMacroExpInst)
        .pMacroData1(pMacroData1)
        .pMacroData2(pMacroData2)
        .pMacroData3(pMacroData3)
        .pMacroData4(pMacroData4)
        .pMacroData5(pMacroData5)
        .fillAndValidateRow();
  }

  final void tracePreprocessing(int stamp, Trace trace) {
    short maxCt = (short) (isExpLog() ? MAX_CT_PRPRC_EXP_LOG : MAX_CT_PRPRC_MODEXP_LOG);
    for (short i = 0; i < maxCt + 1; i++) {
      trace
          .prprc(true)
          .stamp(stamp)
          .ct(i)
          .ctMax(maxCt)
          .isExpLog(isExpLog())
          .isModexpLog(!isExpLog())
          .pPreprocessingWcpFlag(pPreprocessingWcpFlag[i])
          .pPreprocessingWcpArg1Hi(pPreprocessingWcpArg1Hi[i])
          .pPreprocessingWcpArg1Lo(pPreprocessingWcpArg1Lo[i])
          .pPreprocessingWcpArg2Hi(pPreprocessingWcpArg2Hi[i])
          .pPreprocessingWcpArg2Lo(pPreprocessingWcpArg2Lo[i])
          .pPreprocessingWcpInst(pPreprocessingWcpInst[i])
          .pPreprocessingWcpRes(pPreprocessingWcpRes[i])
          .fillAndValidateRow();
    }
  }
}
