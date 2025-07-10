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

package net.consensys.linea.zktracer.module.rlpUtils;

import static net.consensys.linea.zktracer.Trace.*;
import static net.consensys.linea.zktracer.Trace.Rlputils.CT_MAX_INST_INTEGER;
import static net.consensys.linea.zktracer.module.rlpUtils.RlpUtils.BYTES32_PREFIX_SHORT_INT;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.rlptxn.cancun.GenericTracedValue;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.Bytes16;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@Accessors(fluent = true)
public class InstructionInteger extends RlpUtilsCall {
  @Getter private final Bytes32 integer;
  private boolean integerIsZero;
  private boolean integerHiIsNonZero;
  private boolean rlpPrefixRequired;

  public InstructionInteger(Bytes32 integer) {
    super(CT_MAX_INST_INTEGER);
    this.integer = integer;
  }

  @Override
  protected void compute(Wcp wcp) {
    final WcpExoCall firstCall = WcpExoCall.callToIsZero(wcp, integer);
    wcpCalls.add(firstCall);
    integerIsZero = firstCall.result;

    final WcpExoCall secondCall = WcpExoCall.callToGt(wcp, Bytes32.leftPad(data1()), Bytes32.ZERO);
    wcpCalls.add(secondCall);
    integerHiIsNonZero = secondCall.result;

    final WcpExoCall thirdCall =
        WcpExoCall.callToLt(wcp, Bytes32.leftPad(data2()), BYTES32_PREFIX_SHORT_INT);
    wcpCalls.add(thirdCall);
    rlpPrefixRequired = !integerIsZero && !integerHiIsNonZero && thirdCall.result;
  }

  @Override
  public void traceRlpTxn(
      Rlptxn trace,
      GenericTracedValue tracedValues,
      boolean lt,
      boolean lx,
      boolean updateTracedValue,
      int ct) {
    trace.cmp(true).ct(ct).ctMax(2).lt(lt).lx(lx);

    if (ct == 0) {
      trace
          .pCmpRlpUtilsFlag(true)
          .pCmpInst(RLP_UTILS_INST_INTEGER)
          .pCmpExoData1(data1())
          .pCmpExoData2(data2())
          .pCmpExoData3(!integerIsZero)
          .pCmpExoData4(!integerHiIsNonZero)
          .pCmpExoData5(rlpPrefixRequired)
          .pCmpExoData6(rlpPrefix())
          .pCmpExoData7(leadingLimbShifted())
          .pCmpExoData8(leadingLimbBytesize());
    }

    if (ct == 0) {
      if (rlpPrefixRequired) {
        trace.limbConstructed(true).pCmpLimb(rlpPrefix()).pCmpNbytes(1);
        if (lt && updateTracedValue) {
          tracedValues.decrementLtSizeBy(1);
        }
        if (lx && updateTracedValue) {
          tracedValues.decrementLxSizeBy(1);
        }
      }
    }

    if (ct == 1) {
      if (integerHiIsNonZero) {
        trace
            .limbConstructed(true)
            .pCmpLimb(leadingLimbShifted())
            .pCmpNbytes(leadingLimbBytesize());
        if (lt && updateTracedValue) {
          tracedValues.decrementLtSizeBy(leadingLimbBytesize());
        }
        if (lx && updateTracedValue) {
          tracedValues.decrementLxSizeBy(leadingLimbBytesize());
        }
      }
    }

    if (ct == 2) {
      if (!integerIsZero) {
        final int limbLoSize = integerHiIsNonZero ? LLARGE : leadingLimbBytesize();
        trace.limbConstructed(true).pCmpLimb(data2()).pCmpNbytes(limbLoSize);
        if (lt & updateTracedValue) {
          tracedValues.decrementLtSizeBy(limbLoSize);
        }
        if (lx && updateTracedValue) {
          tracedValues.decrementLxSizeBy(limbLoSize);
        }
      }
    }
  }

  @Override
  protected void traceMacro(Trace.Rlputils trace) {
    trace
        .iomf(true)
        .macro(true)
        .isInteger(true)
        .pMacroData1(data1())
        .pMacroData2(data2())
        .pMacroData3(!integerIsZero)
        .pMacroData4(!integerHiIsNonZero)
        .pMacroData5(rlpPrefixRequired)
        .pMacroData6(rlpPrefix())
        .pMacroData7(leadingLimbShifted())
        .pMacroData8(leadingLimbBytesize())
        .fillAndValidateRow();
  }

  @Override
  protected void traceCompt(Trace.Rlputils trace, short ct) {
    final boolean lastRow = ct == CT_MAX_INST_INTEGER;
    trace.iomf(true).macro(true).isInteger(true).ct(ct).ctMax(CT_MAX_INST_INTEGER);
    // related to WCP call
    wcpCalls.get(ct).traceWcpCall(trace);
    // call to POWER ref table for the last row
    trace
        .pComptShfFlag(lastRow)
        .pComptShfArg(lastRow ? 0 : LLARGE - leadingLimbBytesize())
        .pComptShfPower(lastRow ? power(leadingLimbBytesize()) : Bytes.EMPTY)
        .fillAndValidateRow();
  }

  @Override
  protected short instruction() {
    return RLP_UTILS_INST_INTEGER;
  }

  @Override
  protected short compareTo(RlpUtilsCall other) {
    return (short)
        integer
            .toUnsignedBigInteger()
            .compareTo(((InstructionInteger) other).integer.toUnsignedBigInteger());
  }

  @Override
  protected int computeLineCount() {
    return 1 + CT_MAX_INST_INTEGER + 1;
  }

  private Bytes data1() {
    return integer.slice(0, LLARGE);
  }

  private Bytes data2() {
    return integer.slice(LLARGE, LLARGE);
  }

  private Bytes leadingLimbShifted() {
    return integerHiIsNonZero ? data1() : data2();
  }

  private int leadingLimbBytesize() {
    return leadingLimbShifted().trimLeadingZeros().size();
  }

  private Bytes rlpPrefix() {
    return rlpPrefixRequired
        ? Bytes16.leftPad(Bytes.of(RLP_PREFIX_INT_SHORT + integer.trimLeadingZeros().size()))
        : Bytes.EMPTY;
  }
}
