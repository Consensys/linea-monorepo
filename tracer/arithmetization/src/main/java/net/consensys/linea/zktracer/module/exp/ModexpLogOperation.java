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
import static java.lang.Math.min;
import static net.consensys.linea.zktracer.module.exp.Trace.EVM_INST_ISZERO;
import static net.consensys.linea.zktracer.module.exp.Trace.EVM_INST_LT;
import static net.consensys.linea.zktracer.module.exp.Trace.EXP_INST_MODEXPLOG;
import static net.consensys.linea.zktracer.module.exp.Trace.LLARGE;
import static net.consensys.linea.zktracer.module.exp.Trace.LLARGEPO;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_CMPTN_MODEXP_LOG;
import static net.consensys.linea.zktracer.module.exp.Trace.MAX_CT_PRPRC_MODEXP_LOG;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.math.BigInteger;
import java.math.RoundingMode;

import lombok.EqualsAndHashCode;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.exp.ExpCallForModexpLogComputation;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
@RequiredArgsConstructor
public class ModexpLogOperation extends ExpOperation {
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

  @EqualsAndHashCode.Include private final EWord rawLead;
  @EqualsAndHashCode.Include private final int cdsCutoff;
  @EqualsAndHashCode.Include private final int ebsCutoff;
  private final BigInteger leadLog;
  private final EWord trim;

  @Override
  protected boolean isExpLog() {
    return false;
  }

  public static ModexpLogOperation fromExpLogCall(
      final Wcp wcp, final ExpCallForModexpLogComputation c) {
    final LeadLogTrimLead leadLogTrimLead =
        LeadLogTrimLead.fromArgs(c.rawLeadingWord(), c.cdsCutoff(), c.ebsCutoff());

    final ModexpLogOperation modExpLogOperation =
        new ModexpLogOperation(
            c.rawLeadingWord(),
            c.cdsCutoff(),
            c.ebsCutoff(),
            BigInteger.valueOf(leadLogTrimLead.leadLog),
            EWord.of(leadLogTrimLead.trim));

    modExpLogOperation.wcp = wcp;
    modExpLogOperation.preCompute();
    return modExpLogOperation;
  }

  @Override
  public void preCompute() {
    pMacroExpInst = EXP_INST_MODEXPLOG;
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
    pPreprocessingWcpInst[0] = UnsignedByte.of(EVM_INST_LT);
    pPreprocessingWcpRes[0] = wcp.callLT(Bytes.of(this.cdsCutoff), Bytes.of(this.ebsCutoff));
    final int minCutoff = min(this.cdsCutoff, this.ebsCutoff);

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
    pPreprocessingWcpFlag[2] = true;
    pPreprocessingWcpArg1Hi[2] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[2] = Bytes.of(this.ebsCutoff);
    pPreprocessingWcpArg2Hi[2] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[2] = Bytes.of(LLARGEPO);
    pPreprocessingWcpInst[2] = UnsignedByte.of(EVM_INST_LT);
    pPreprocessingWcpRes[2] = wcp.callLT(Bytes.of(this.ebsCutoff), Bytes.of(LLARGEPO));

    // Fourth row
    pPreprocessingWcpFlag[3] = true;
    pPreprocessingWcpArg1Hi[3] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[3] = this.rawLead.hi();
    pPreprocessingWcpArg2Hi[3] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[3] = Bytes.of(0);
    pPreprocessingWcpInst[3] = UnsignedByte.of(EVM_INST_ISZERO);
    final boolean rawHiPartIsZero = wcp.callISZERO(this.rawLead.hi());
    pPreprocessingWcpRes[3] = rawHiPartIsZero;

    // Fifth row
    final int paddedBase2Log =
        8 * nBytesExcludingLeadingByte + nBitsOfLeadingByteExcludingLeadingBit;

    pPreprocessingWcpFlag[4] = true;
    pPreprocessingWcpArg1Hi[4] = Bytes.of(0);
    pPreprocessingWcpArg1Lo[4] = Bytes.of(paddedBase2Log);
    pPreprocessingWcpArg2Hi[4] = Bytes.of(0);
    pPreprocessingWcpArg2Lo[4] = Bytes.of(0);
    pPreprocessingWcpInst[4] = UnsignedByte.of(EVM_INST_ISZERO);
    pPreprocessingWcpRes[4] = wcp.callISZERO(Bytes.of(paddedBase2Log));

    // Linking constraints and fill rawAcc
    if (minCutoffLeq16) {
      pComputationRawAcc = leftPadTo(this.rawLead.hi(), LLARGE);
    } else if (!rawHiPartIsZero) {
      pComputationRawAcc = leftPadTo(this.rawLead.hi(), LLARGE);
    } else {
      pComputationRawAcc = leftPadTo(this.rawLead.lo(), LLARGE);
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
