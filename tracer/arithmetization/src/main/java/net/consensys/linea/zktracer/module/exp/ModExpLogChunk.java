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

import static com.google.common.math.BigIntegerMath.log2;
import static java.lang.Math.max;
import static java.lang.Math.min;
import static net.consensys.linea.zktracer.module.exp.Trace.EXP_MODEXPLOG;
import static net.consensys.linea.zktracer.module.exp.Trace.ISZERO;
import static net.consensys.linea.zktracer.module.exp.Trace.LT;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_CMPTN_MODEXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_PRPRC_MODEXP_LOG;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.math.BigInteger;
import java.math.RoundingMode;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.ModExpLogCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;

@RequiredArgsConstructor
public class ModExpLogChunk extends ExpChunk {
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

  private final EWord rawLead;
  private final int cdsCutoff;
  private final int ebsCutoff;
  private final BigInteger leadLog;
  private final EWord trim;

  @Override
  protected boolean isExpLog() {
    return false;
  }

  public static ModExpLogChunk fromExpLogCall(final Wcp wcp, final ModExpLogCall c) {
    final LeadLogTrimLead leadLogTrimLead =
        LeadLogTrimLead.fromArgs(c.rawLeadingWord(), c.cdsCutoff(), c.ebsCutoff());

    final ModExpLogChunk modExpLogChunk =
        new ModExpLogChunk(
            c.rawLeadingWord(),
            c.cdsCutoff(),
            c.ebsCutoff(),
            BigInteger.valueOf(leadLogTrimLead.leadLog),
            EWord.of(leadLogTrimLead.trim));

    modExpLogChunk.wcp = wcp;
    modExpLogChunk.preCompute();
    return modExpLogChunk;
  }

  public static ModExpLogChunk fromFrame(final Wcp wcp, final MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    // DELEGATECALL, STATICCALL cases
    int cdoIndex = 2;
    int cdsIndex = 3;
    // CALL, CALLCODE cases
    if (opCode == OpCode.CALL || opCode == OpCode.CALLCODE) {
      cdoIndex = 3;
      cdsIndex = 4;
    }

    // TODO: use OperationTransient here
    final BigInteger cdo = frame.getStackItem(cdoIndex).toUnsignedBigInteger();
    final BigInteger cds = frame.getStackItem(cdsIndex).toUnsignedBigInteger();

    // mxp should ensure that the hi part of cds is 0

    if (cds.signum() == 0) {
      return null;
    }
    // Here cds != 0

    final Bytes unpaddedCallData = frame.shadowReadMemory(cdo.longValue(), cds.longValue());

    // pad unpaddedCallData to 96 (this is probably not necessary)
    final Bytes paddedCallData =
        cds.intValue() < 96
            ? Bytes.concatenate(unpaddedCallData, Bytes.repeat((byte) 0, 96 - cds.intValue()))
            : unpaddedCallData;

    final BigInteger bbs = paddedCallData.slice(0, 32).toUnsignedBigInteger();
    final BigInteger ebs = paddedCallData.slice(32, 32).toUnsignedBigInteger();

    // Some other module checks if bbs, ebs and msb are <= 512 (@Francois)

    if (ebs.signum() == 0) {
      return null;
    }
    // Here ebs != 0

    if (cds.compareTo(BigInteger.valueOf(96).add(bbs)) <= 0) {
      return null;
    }

    // pad paddedCallData to 96 + bbs + 32
    final Bytes doublePaddedCallData =
        cds.intValue() < 96 + bbs.intValue() + 32
            ? Bytes.concatenate(
                paddedCallData, Bytes.repeat((byte) 0, 96 + bbs.intValue() + 32 - cds.intValue()))
            : paddedCallData;

    // raw_lead
    final EWord rawLead = EWord.of(doublePaddedCallData.slice(96 + bbs.intValue(), 32));

    // cds_cutoff
    final int cdsCutoff = min(max(cds.intValue() - (96 + bbs.intValue()), 0), 32);
    // ebs_cutoff
    final int ebsCutoff = min(ebs.intValue(), 32);

    final LeadLogTrimLead leadLogTrimLead = LeadLogTrimLead.fromArgs(rawLead, cdsCutoff, ebsCutoff);

    final ModExpLogChunk modExpLogChunk =
        new ModExpLogChunk(
            rawLead,
            cdsCutoff,
            ebsCutoff,
            BigInteger.valueOf(leadLogTrimLead.leadLog),
            EWord.of(leadLogTrimLead.trim));

    modExpLogChunk.wcp = wcp;
    modExpLogChunk.preCompute();
    return modExpLogChunk;
  }

  @Override
  public void preCompute() {
    pMacroExpInst = EXP_MODEXPLOG;
    pMacroData1 = this.rawLead.hi();
    pMacroData2 = this.rawLead.lo();
    pMacroData3 = Bytes.of(this.cdsCutoff);
    pMacroData4 = Bytes.of(this.ebsCutoff);
    pMacroData5 = bigIntegerToBytes(this.leadLog);
    initArrays(MAX_CT_PRPRC_MODEXP_LOG + 1);

    // Preprocessing
    final BigInteger trimLimb =
        this.trim.hi().isZero() ? this.trim.loBigInt() : this.trim.hiBigInt();
    final int trimLog = trimLimb.signum() == 0 ? 0 : log2(trimLimb, RoundingMode.FLOOR);
    final int nBitsOfLeadingByteExcludingLeadingBit = trimLog % 8;
    final int nBytesExcludingLeadingByte = trimLog / 8;

    // First row
    pPreprocessingWcpFlag[0] = true;
    pPreprocessingWcpArg1Hi[0] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[0] = Bytes.of(this.cdsCutoff);
    pPreprocessingWcpArg2Hi[0] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[0] = Bytes.of(this.ebsCutoff);
    pPreprocessingWcpInst[0] = UnsignedByte.of(LT);
    pPreprocessingWcpRes[0] = this.cdsCutoff < this.ebsCutoff;
    final int minCutoff = min(this.cdsCutoff, this.ebsCutoff);

    // Lookup
    wcp.callLT(Bytes.of(this.cdsCutoff), Bytes.of(this.ebsCutoff));

    // Second row
    pPreprocessingWcpFlag[1] = true;
    pPreprocessingWcpArg1Hi[1] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[1] = Bytes.of(minCutoff);
    pPreprocessingWcpArg2Hi[1] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[1] = Bytes.of(17);
    pPreprocessingWcpInst[1] = UnsignedByte.of(LT);
    pPreprocessingWcpRes[1] = minCutoff < 17;
    final boolean minCutoffLeq16 = pPreprocessingWcpRes[1];

    // Lookup
    wcp.callLT(Bytes.of(minCutoff), Bytes.of(17));

    // Third row
    pPreprocessingWcpFlag[2] = true;
    pPreprocessingWcpArg1Hi[2] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[2] = Bytes.of(this.ebsCutoff);
    pPreprocessingWcpArg2Hi[2] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[2] = Bytes.of(17);
    pPreprocessingWcpInst[2] = UnsignedByte.of(LT);
    pPreprocessingWcpRes[2] = this.ebsCutoff < 17;

    // Lookup
    wcp.callLT(Bytes.of(this.ebsCutoff), Bytes.of(17));

    // Fourth row
    pPreprocessingWcpFlag[3] = true;
    pPreprocessingWcpArg1Hi[3] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[3] = this.rawLead.hi();
    pPreprocessingWcpArg2Hi[3] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[3] = Bytes.of(0);
    pPreprocessingWcpInst[3] = UnsignedByte.of(ISZERO);
    pPreprocessingWcpRes[3] = this.rawLead.hi().isZero();
    final boolean rawHiPartIsZero = pPreprocessingWcpRes[3];

    // Lookup
    wcp.callISZERO(this.rawLead.hi());

    // Fifth row
    final int paddedBase2Log =
        8 * nBytesExcludingLeadingByte + nBitsOfLeadingByteExcludingLeadingBit;

    pPreprocessingWcpFlag[4] = true;
    pPreprocessingWcpArg1Hi[4] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[4] = Bytes.of(paddedBase2Log);
    pPreprocessingWcpArg2Hi[4] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[4] = Bytes.of(0);
    pPreprocessingWcpInst[4] = UnsignedByte.of(ISZERO);
    pPreprocessingWcpRes[4] = paddedBase2Log == 0;

    // Lookup
    wcp.callISZERO(Bytes.of(paddedBase2Log));

    // Linking constraints and fill rawAcc
    if (minCutoffLeq16) {
      pComputationRawAcc = leftPadTo(this.rawLead.hi(), 16);
    } else if (!rawHiPartIsZero) {
      pComputationRawAcc = leftPadTo(this.rawLead.hi(), 16);
    } else {
      pComputationRawAcc = leftPadTo(this.rawLead.lo(), 16);
    }

    // Fill pltJmp
    if (minCutoffLeq16) {
      pComputationPltJmp = (short) minCutoff;
    } else {
      if (!rawHiPartIsZero) {
        pComputationPltJmp = (short) 16;
      } else {
        pComputationPltJmp = (short) (minCutoff - 16);
      }
    }

    // Fill trimAcc
    final short maxCt = (short) MAX_CT_CMPTN_MODEXP_LOG;
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
  }
}
