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

package net.consensys.linea.zktracer.module.stp;

import static com.google.common.base.Preconditions.*;

import java.math.BigInteger;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.hub.fragment.imc.StpCall;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@Accessors(fluent = true)
@Getter
@EqualsAndHashCode(callSuper = false)
public final class StpOperation extends ModuleOperation {
  private final StpCall stpCall;

  public StpOperation(StpCall stpCall) {
    this.stpCall = stpCall;
  }

  private boolean isCall() {
    return stpCall.opCode() == OpCode.CALL;
  }

  private boolean isCallCode() {
    return stpCall.opCode() == OpCode.CALLCODE;
  }

  private boolean isDelegateCall() {
    return stpCall.opCode() == OpCode.DELEGATECALL;
  }

  private boolean isStaticCall() {
    return stpCall.opCode() == OpCode.STATICCALL;
  }

  private boolean isCreate() {
    return stpCall.opCode() == OpCode.CREATE;
  }

  private boolean isCreate2() {
    return stpCall.opCode() == OpCode.CREATE2;
  }

  long getGDiff() {
    checkArgument(!stpCall.outOfGasException());
    return stpCall.gasActual() - stpCall.upfrontGasCost();
  }

  long getGDiffOver64() {
    return this.getGDiff() / 64;
  }

  long get63of64GDiff() {
    return this.getGDiff() - this.getGDiffOver64();
  }

  void trace(Trace.Stp trace, int stamp) {
    if (stpCall.opCode().isCreate()) {
      this.traceCreate(trace, stamp);
    } else {
      this.traceCall(trace, stamp);
    }
  }

  private void traceCreate(Trace.Stp trace, int stamp) {
    final int ctMax = this.maxCt();
    final long gasOopkt = stpCall.outOfGasException() ? 0 : this.get63of64GDiff();

    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .stamp(stamp)
          .ct(UnsignedByte.of(ct))
          .ctMax(UnsignedByte.of(ctMax))
          .instruction(UnsignedByte.of(stpCall.opCode().byteValue()))
          .isCreate(isCreate())
          .isCreate2(isCreate2())
          // .gasHi(Bytes.EMPTY) // redundant
          // .gasLo(Bytes.EMPTY) // redundant
          .valHi(stpCall.value().slice(0, 16))
          .valLo(stpCall.value().slice(16, 16))
          // .exists(false) // redundant
          // .warm(false)   // redundant
          .outOfGasException(stpCall.outOfGasException())
          .gasActual(Bytes.ofUnsignedLong(stpCall.gasActual()))
          .gasMxp(Bytes.ofUnsignedLong(stpCall.memoryExpansionGas()))
          .gasUpfront(Bytes.ofUnsignedLong(stpCall.upfrontGasCost()))
          .gasOutOfPocket(Bytes.ofUnsignedLong(gasOopkt))
          .gasStipend(Bytes.ofUnsignedLong(stpCall.stipend()))
          .arg1Hi(Bytes.EMPTY);

      switch (ct) {
        case 0 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(stpCall.gasActual()))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.EMPTY) // we REQUIRE that the currently available gas is nonnegative
            .wcpFlag(true)
            .modFlag(false)
            .fillAndValidateRow();
        case 1 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(stpCall.gasActual()))
            .arg2Lo(Bytes.ofUnsignedLong(stpCall.upfrontGasCost()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.of(stpCall.outOfGasException() ? 1 : 0))
            .wcpFlag(true)
            .modFlag(false)
            .fillAndValidateRow();
        case 2 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(getGDiff()))
            .arg2Lo(Bytes.of(64))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.DIV.byteValue()))
            .resLo(Bytes.ofUnsignedLong(getGDiffOver64()))
            .wcpFlag(false)
            .modFlag(true)
            .fillAndValidateRow();
        default -> throw new IllegalArgumentException("counter too big, should be <=" + ctMax);
      }
    }
  }

  private void traceCall(Trace.Stp trace, int stamp) {
    final int ctMax = this.maxCt();
    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .stamp(stamp)
          .ct(UnsignedByte.of(ct))
          .ctMax(UnsignedByte.of(ctMax))
          .instruction(UnsignedByte.of(stpCall.opCode().byteValue()))
          .isCall(isCall())
          .isCallcode(isCallCode())
          .isDelegatecall(isDelegateCall())
          .isStaticcall(isStaticCall())
          .gasHi(stpCall.gas().slice(0, 16))
          .gasLo(stpCall.gas().slice(16))
          .valHi(stpCall.value().slice(0, 16))
          .valLo(stpCall.value().slice(16))
          .exists(stpCall.exists())
          .warm(stpCall.warm())
          .outOfGasException(stpCall.outOfGasException())
          .gasActual(Bytes.ofUnsignedLong(stpCall.gasActual()))
          .gasMxp(Bytes.ofUnsignedLong(stpCall.memoryExpansionGas()))
          .gasUpfront(Bytes.ofUnsignedLong(stpCall.upfrontGasCost()))
          .gasOutOfPocket(Bytes.ofUnsignedLong(stpCall.gasPaidOutOfPocket()))
          .gasStipend(Bytes.ofUnsignedLong(stpCall.stipend()));

      switch (ct) {
        case 0 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(stpCall.gasActual()))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.EMPTY) // we REQUIRE that the currently available gas is nonnegative
            .wcpFlag(true)
            .modFlag(false)
            .fillAndValidateRow();
        case 1 -> trace
            .arg1Hi(stpCall.value().slice(0, 16))
            .arg1Lo(stpCall.value().slice(16, 16))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.ISZERO.byteValue()))
            .resLo(Bytes.of(stpCall.value().isZero() ? 1 : 0))
            .wcpFlag(stpCall.opCode().callHasValueArgument())
            .modFlag(false)
            .fillAndValidateRow();
        case 2 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(stpCall.gasActual()))
            .arg2Lo(Bytes.ofUnsignedLong(stpCall.upfrontGasCost()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.of(stpCall.outOfGasException() ? 1 : 0))
            .wcpFlag(true)
            .modFlag(false)
            .fillAndValidateRow();
          // the following rows are only filled in if no out of gas exception
        case 3 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(getGDiff()))
            .arg2Lo(Bytes.of(64))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.DIV.byteValue()))
            .resLo(Bytes.ofUnsignedLong(getGDiffOver64()))
            .wcpFlag(false)
            .modFlag(true)
            .fillAndValidateRow();
        case 4 -> trace
            .arg1Hi(stpCall.gas().slice(0, 16))
            .arg1Lo(stpCall.gas().slice(16, 16))
            .arg2Lo(Bytes.ofUnsignedLong(getGDiff() - getGDiffOver64()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(
                Bytes.of(
                    stpCall
                                .gas()
                                .toUnsignedBigInteger()
                                .compareTo(BigInteger.valueOf(get63of64GDiff()))
                            < 0
                        ? 1
                        : 0))
            .wcpFlag(true)
            .modFlag(false)
            .fillAndValidateRow();
        default -> throw new IllegalArgumentException("counter too big, should be <=" + ctMax);
      }
    }
  }

  private int maxCt() {
    if (stpCall.outOfGasException()) {
      return stpCall.opCode().isCreate() ? 1 : 2;
    } else {
      return stpCall.opCode().isCreate() ? 2 : 4;
    }
  }

  @Override
  protected int computeLineCount() {
    return 1 + this.maxCt();
  }
}
