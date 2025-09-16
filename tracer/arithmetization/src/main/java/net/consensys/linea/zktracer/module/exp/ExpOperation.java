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
import static net.consensys.linea.zktracer.Trace.EXP_INST_EXPLOG;
import static net.consensys.linea.zktracer.Trace.EXP_INST_MODEXPLOG;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_G_EXP_BYTE;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

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
import org.apache.tuweni.bytes.Bytes;

@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
@Getter
@Accessors(fluent = true)
public class ExpOperation extends ModuleOperation {
  @EqualsAndHashCode.Include ExpCall expCall;

  public ExpOperation(ExpCall expCall, Wcp wcp, Hub hub) {
    this.expCall = expCall;

    switch (expCall.expInstruction()) {
      case EXP_INST_EXPLOG -> {
        ExplogExpCall explogExpCall = (ExplogExpCall) expCall;
        // Extract inputs
        EWord exponent = EWord.of(hub.messageFrame().getStackItem(1));
        long dynCost = (long) GAS_CONST_G_EXP_BYTE * exponent.byteLength();
        // Fill expCall
        explogExpCall.exponent(exponent);
        explogExpCall.dynCost(dynCost);
      }
      case EXP_INST_MODEXPLOG -> {
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
      }
      default -> throw new IllegalArgumentException(
          "invalid EXP instruction: " + expCall.expInstruction());
    }
  }

  final void trace(Trace.Exp trace) {
    // Handle each case separately
    switch (expCall.expInstruction()) {
      case EXP_INST_EXPLOG -> {
        ExplogExpCall call = (ExplogExpCall) expCall;
        trace
            .inst(EXP_INST_EXPLOG)
            .arg(call.exponent())
            .cds(0)
            .ebs(0)
            .res(Bytes.ofUnsignedLong(call.dynCost()))
            .validateRow();
        break;
      }
      case EXP_INST_MODEXPLOG -> {
        ModexpLogExpCall call = (ModexpLogExpCall) expCall;
        trace
            .inst(EXP_INST_MODEXPLOG)
            .arg(call.getRawLeadingWord())
            .cds(call.getCdsCutoff())
            .ebs(call.getEbsCutoff())
            .res(bigIntegerToBytes(call.getLeadLog()))
            .validateRow();
        break;
      }
      default -> throw new IllegalArgumentException(
          "invalid EXP instruction: " + expCall.expInstruction());
    }
  }

  @Override
  protected int computeLineCount() {
    return 1;
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
