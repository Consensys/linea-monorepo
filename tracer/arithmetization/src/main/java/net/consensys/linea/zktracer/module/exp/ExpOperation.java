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

import static com.google.common.base.Preconditions.*;
import static com.google.common.math.BigIntegerMath.log2;
import static java.lang.Math.min;
import static net.consensys.linea.zktracer.Trace.EVM_INST_ISZERO;
import static net.consensys.linea.zktracer.Trace.EVM_INST_LT;
import static net.consensys.linea.zktracer.Trace.EXP_INST_EXPLOG;
import static net.consensys.linea.zktracer.Trace.EXP_INST_MODEXPLOG;
import static net.consensys.linea.zktracer.Trace.Exp.CT_MAX_CMPTN_EXP_LOG;
import static net.consensys.linea.zktracer.Trace.Exp.CT_MAX_CMPTN_MODEXP_LOG;
import static net.consensys.linea.zktracer.Trace.Exp.CT_MAX_PRPRC_EXP_LOG;
import static net.consensys.linea.zktracer.Trace.Exp.CT_MAX_PRPRC_MODEXP_LOG;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_EXP_BYTE;
import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.Trace.LLARGEPO;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.math.BigInteger;
import java.math.RoundingMode;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ExplogExpCall;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ModexpLogExpCall;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
@Getter
@Accessors(fluent = true)
public class ExpOperation extends ModuleOperation {
  @EqualsAndHashCode.Include ExpCall expCall;

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

  boolean isExpLog;

  private final Wcp wcp;
  private final Hub hub;

  public ExpOperation(ExpCall expCall, Wcp wcp, Hub hub) {
    this.expCall = expCall;
    this.wcp = wcp;
    this.hub = hub;
    if (expCall.expInstruction() == EXP_INST_EXPLOG) {
      this.isExpLog = true;
      ExplogExpCall explogExpCall = (ExplogExpCall) expCall;

      // Extract inputs
      EWord exponent = EWord.of(hub.messageFrame().getStackItem(1));
      long dynCost = (long) GAS_CONST_G_EXP_BYTE * exponent.byteLength();

      // Fill expCall
      explogExpCall.exponent(exponent);
      explogExpCall.dynCost(dynCost);

      // Execute preprocessing
      preComputeForExplog(explogExpCall);
    } else if (expCall.expInstruction() == EXP_INST_MODEXPLOG) {
      isExpLog = false;
      ModexpLogExpCall modexplogExpCall = (ModexpLogExpCall) expCall;

      // Extract inputs
      final ModexpMetadata modexpMetadata = modexplogExpCall.getModexpMetadata();
      final int bbsInt = modexpMetadata.bbs().toUnsignedBigInteger().intValueExact();
      final int ebsInt = modexpMetadata.ebs().toUnsignedBigInteger().intValueExact();
      checkArgument(modexpMetadata.callData().size() - 96 - bbsInt >= 0);
      EWord rawLead = modexpMetadata.rawLeadingWord();
      int cdsCutoff = Math.min(modexpMetadata.callData().size() - 96 - bbsInt, 32);
      int ebsCutoff = Math.min(ebsInt, 32);
      BigInteger leadLog =
          BigInteger.valueOf(LeadLogTrimLead.fromArgs(rawLead, cdsCutoff, ebsCutoff).leadLog());

      // Fill expCall
      modexplogExpCall.setRawLeadingWord(rawLead);
      modexplogExpCall.setCdsCutoff(cdsCutoff);
      modexplogExpCall.setEbsCutoff(ebsCutoff);
      modexplogExpCall.setLeadLog(leadLog);

      // Execute preprocessing
      preComputeForModexpLog(modexplogExpCall);
    }
  }

  public void preComputeForExplog(ExplogExpCall explogExpCall) {
    pMacroExpInst = EXP_INST_EXPLOG;
    pMacroData1 = explogExpCall.exponent().hi();
    pMacroData2 = explogExpCall.exponent().lo();
    pMacroData5 = Bytes.ofUnsignedLong(explogExpCall.dynCost());
    initArrays(CT_MAX_PRPRC_EXP_LOG + 1);

    // Preprocessing
    // First row
    pPreprocessingWcpFlag[0] = true;
    pPreprocessingWcpArg1Hi[0] = Bytes.EMPTY;
    pPreprocessingWcpArg1Lo[0] = explogExpCall.exponent().hi();
    pPreprocessingWcpArg2Hi[0] = Bytes.EMPTY;
    pPreprocessingWcpArg2Lo[0] = Bytes.EMPTY;
    pPreprocessingWcpInst[0] = UnsignedByte.of(EVM_INST_ISZERO);
    final boolean expnHiIsZero = wcp.callISZERO(explogExpCall.exponent().hi());
    ;
    pPreprocessingWcpRes[0] = expnHiIsZero;

    // Linking constraints and fill rawAcc
    pComputationPltJmp = 16;
    pComputationRawAcc = explogExpCall.exponent().hi();
    if (expnHiIsZero) {
      pComputationRawAcc = explogExpCall.exponent().lo();
    }

    // Fill trimAcc
    short maxCt = (short) CT_MAX_CMPTN_EXP_LOG;
    for (short i = 0; i < maxCt + 1; i++) {
      boolean pltBit = i >= pComputationPltJmp;
      byte rawByte = pComputationRawAcc.get(i);
      byte trimByte = pltBit ? 0 : rawByte;
      pComputationTrimAcc = Bytes.concatenate(pComputationTrimAcc, Bytes.of(trimByte));
    }
  }

  public void preComputeForModexpLog(ModexpLogExpCall modexplogExpCall) {
    pMacroExpInst = EXP_INST_MODEXPLOG;
    pMacroData1 = modexplogExpCall.getRawLeadingWord().hi();
    pMacroData2 = modexplogExpCall.getRawLeadingWord().lo();
    pMacroData3 = Bytes.of(modexplogExpCall.getCdsCutoff());
    pMacroData4 = Bytes.of(modexplogExpCall.getEbsCutoff());
    pMacroData5 = bigIntegerToBytes(modexplogExpCall.getLeadLog());
    initArrays(CT_MAX_PRPRC_MODEXP_LOG + 1);

    // Preprocessing

    // First row
    pPreprocessingWcpFlag[0] = true;
    pPreprocessingWcpArg1Hi[0] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[0] = Bytes.of(modexplogExpCall.getCdsCutoff());
    pPreprocessingWcpArg2Hi[0] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[0] = Bytes.of(modexplogExpCall.getEbsCutoff());
    pPreprocessingWcpInst[0] = UnsignedByte.of(EVM_INST_LT);
    pPreprocessingWcpRes[0] =
        wcp.callLT(
            Bytes.of(modexplogExpCall.getCdsCutoff()), Bytes.of(modexplogExpCall.getEbsCutoff()));
    final int minCutoff = min(modexplogExpCall.getCdsCutoff(), modexplogExpCall.getEbsCutoff());

    // Second row
    pPreprocessingWcpFlag[1] = true;
    pPreprocessingWcpArg1Hi[1] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[1] = Bytes.of(minCutoff);
    pPreprocessingWcpArg2Hi[1] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[1] = Bytes.of(LLARGEPO);
    pPreprocessingWcpInst[1] = UnsignedByte.of(EVM_INST_LT);
    final boolean minCutoffLeq16 = wcp.callLT(Bytes.of(minCutoff), Bytes.of(LLARGEPO));
    pPreprocessingWcpRes[1] = minCutoffLeq16;

    // Third row
    final EWord rawLead = modexplogExpCall.getRawLeadingWord();
    final Bytes rawLeadHi = rawLead.hi();
    final Bytes rawLeadLo = rawLead.lo();
    pPreprocessingWcpFlag[2] = true;
    pPreprocessingWcpArg1Hi[2] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[2] = rawLeadHi;
    pPreprocessingWcpArg2Hi[2] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[2] = Bytes.of(0);
    pPreprocessingWcpInst[2] = UnsignedByte.of(EVM_INST_ISZERO);
    final boolean rawLeadHiIsZero = wcp.callISZERO(rawLeadHi);
    pPreprocessingWcpRes[2] = rawLeadHiIsZero;

    // Fourth row is filled later since we need pComputationTrimAcc

    // Linking constraints and fill rawAcc and pltJmp
    if (minCutoffLeq16) {
      pComputationRawAcc = leftPadTo(rawLeadHi, LLARGE);
      pComputationPltJmp = (short) minCutoff;
    } else {
      if (!rawLeadHiIsZero) {
        pComputationRawAcc = leftPadTo(rawLeadHi, LLARGE);
        pComputationPltJmp = (short) 16;
      } else {
        pComputationRawAcc = leftPadTo(rawLeadLo, LLARGE);
        pComputationPltJmp = (short) (minCutoff - 16);
      }
    }

    // Fill trimAcc
    final short maxCt = (short) CT_MAX_CMPTN_MODEXP_LOG;
    for (short i = 0; i < maxCt + 1; i++) {
      final boolean pltBit = i >= pComputationPltJmp;
      final byte rawByte = pComputationRawAcc.get(i);
      final byte trimByte = pltBit ? 0 : rawByte;
      pComputationTrimAcc = Bytes.concatenate(pComputationTrimAcc, Bytes.of(trimByte));
      if (trimByte != 0 && pComputationMsb.toInteger() == 0) {
        // Fill msb
        pComputationMsb = UnsignedByte.of(trimByte);
      }
    }

    // Fourth row
    pPreprocessingWcpFlag[3] = true;
    pPreprocessingWcpArg1Hi[3] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[3] = pComputationTrimAcc;
    pPreprocessingWcpArg2Hi[3] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[3] = Bytes.of(0);
    pPreprocessingWcpInst[3] = UnsignedByte.of(EVM_INST_ISZERO);
    final boolean trimAccIsZero = wcp.callISZERO(pComputationTrimAcc);
    pPreprocessingWcpRes[3] = trimAccIsZero;
  }

  final void traceComputation(int stamp, Trace.Exp trace) {
    boolean tanzb;
    short pComputationTanzbAcc = 0; // Paired with Tanzb
    boolean manzb;
    short pComputationManzbAcc = 0; // Paired with Manzb
    short maxCt = (short) (isExpLog() ? CT_MAX_CMPTN_EXP_LOG : CT_MAX_CMPTN_MODEXP_LOG);

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
      tanzb = pComputationTrimAcc.slice(0, i + 1).toUnsignedBigInteger().signum() != 0;
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

  final void traceMacro(int stamp, Trace.Exp trace) {
    // We assume CT_MAX_MACRO_EXP_LOG = CT_MAX_MACRO_MODEXP_LOG = 0;
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

  final void tracePreprocessing(int stamp, Trace.Exp trace) {
    short maxCt = (short) (isExpLog() ? CT_MAX_PRPRC_EXP_LOG : CT_MAX_PRPRC_MODEXP_LOG);
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

  private boolean isExpLog() {
    return isExpLog;
  }

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
      return CT_MAX_CMPTN_EXP_LOG + CT_MAX_PRPRC_EXP_LOG + 3;
    }
    return CT_MAX_CMPTN_MODEXP_LOG + CT_MAX_PRPRC_MODEXP_LOG + 3;
  }

  public record LeadLogTrimLead(int leadLog, BigInteger trim) {
    public static LeadLogTrimLead fromArgs(EWord rawLead, int cdsCutoff, int ebsCutoff) {
      // min_cutoff
      final int minCutoff = min(cdsCutoff, ebsCutoff);

      BigInteger mask = new BigInteger("FF".repeat(minCutoff), 16);
      if (minCutoff < 32) {
        // 32 - minCutoff is the shift distance in bytes, but we need bits
        mask = mask.shiftLeft(8 * (32 - minCutoff));
      }

      // trim (keep only minCutoff bytes of rawLead)
      final BigInteger trim = rawLead.toUnsignedBigInteger().and(mask);

      // lead (keep only minCutoff bytes of rawLead and potentially pad to ebsCutoff with 0's)
      final BigInteger lead = trim.shiftRight(8 * (32 - ebsCutoff));

      // lead_log (same as EYP)
      final int leadLog = lead.signum() == 0 ? 0 : log2(lead, RoundingMode.FLOOR);

      return new LeadLogTrimLead(leadLog, trim);
    }
  }
}
