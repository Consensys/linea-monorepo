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

import static net.consensys.linea.zktracer.module.stp.Stp.callCanTransferValue;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;
import java.util.Optional;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;

@RequiredArgsConstructor
@Accessors(fluent = true)
@Getter
public final class StpChunk extends ModuleOperation {
  private final OpCode opCode;
  private final Long gasActual;
  private final Long gasPrelim;
  private final Boolean oogx;
  private final Long gasMxp;
  private final Wei balance;
  private final Address to;
  private final Bytes32 value;
  private final Optional<Boolean> toExists;
  private final Optional<Boolean> toWarm;
  private final Optional<Bytes32> gas;

  // Used by Create's instruction
  public StpChunk(
      OpCode opcode,
      Long gasActual,
      Long gasPrelim,
      Boolean oogx,
      Long gasMxp,
      Wei balance,
      Address to,
      Bytes32 value) {
    this(
        opcode,
        gasActual,
        gasPrelim,
        oogx,
        gasMxp,
        balance,
        to,
        value,
        Optional.empty(),
        Optional.empty(),
        Optional.empty());
  }

  // Used by Call's instruction
  public StpChunk(
      OpCode opcode,
      Long gasActual,
      Long gasPrelim,
      Boolean oogx,
      Long gasMxp,
      Wei balance,
      Address to,
      Bytes32 value,
      Boolean toExists,
      Boolean toWarm,
      Bytes32 gas) {
    this(
        opcode,
        gasActual,
        gasPrelim,
        oogx,
        gasMxp,
        balance,
        to,
        value,
        Optional.of(toExists),
        Optional.of(toWarm),
        Optional.of(gas));
  }

  long getGDiff() {
    Preconditions.checkArgument(!this.oogx());
    return this.gasActual() - this.gasPrelim();
  }

  long getGDiffOver64() {
    return this.getGDiff() / 64;
  }

  long get63of64GDiff() {
    return this.getGDiff() - this.getGDiffOver64();
  }

  void trace(Trace trace, int stamp) {
    if (this.opCode().isCreate()) {
      this.traceCreate(trace, stamp);
    } else {
      this.traceCall(trace, stamp);
    }
  }

  private void traceCreate(Trace trace, int stamp) {
    final int ctMax = this.maxCt();
    final long gasOopkt = this.oogx() ? 0 : this.get63of64GDiff();

    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .stamp(Bytes.ofUnsignedInt(stamp))
          .ct(Bytes.of(ct))
          .ctMax(Bytes.of(ctMax))
          .instruction(UnsignedByte.of(this.opCode().byteValue()))
          .isCreate(this.opCode() == OpCode.CREATE)
          .isCreate2(this.opCode() == OpCode.CREATE2)
          .isCall(false)
          .isCallcode(false)
          .isDelegatecall(false)
          .isStaticcall(false)
          .gasHi(Bytes.EMPTY)
          .gasLo(Bytes.EMPTY)
          .valHi(this.value().slice(0, 16))
          .valLo(this.value().slice(16, 16))
          .exists(false) // TODO document this
          .warm(false) // TODO document this
          .outOfGasException(this.oogx())
          .gasActual(Bytes.ofUnsignedLong(this.gasActual()))
          .gasMxp(Bytes.ofUnsignedLong(this.gasMxp()))
          .gasUpfront(Bytes.ofUnsignedLong(this.gasPrelim()))
          .gasOopkt(Bytes.ofUnsignedLong(gasOopkt))
          .gasStipend(Bytes.EMPTY)
          .arg1Hi(Bytes.EMPTY);

      switch (ct) {
        case 0 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(this.gasActual()))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.EMPTY) // we REQUIRE that the currently available gas is nonnegative
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        case 1 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(this.gasActual()))
            .arg2Lo(Bytes.ofUnsignedLong(this.gasPrelim()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.of(this.oogx() ? 1 : 0))
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        case 2 -> trace
            .arg1Lo(Bytes.ofUnsignedLong(getGDiff()))
            .arg2Lo(Bytes.of(64))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.DIV.byteValue()))
            .resLo(Bytes.ofUnsignedLong(getGDiffOver64()))
            .wcpFlag(false)
            .modFlag(true)
            .validateRow();
        default -> throw new IllegalArgumentException("counter too big, should be <=" + ctMax);
      }
    }
  }

  private void traceCall(Trace trace, int stamp) {
    final int ctMax = this.maxCt();
    final long gasStipend =
        (!this.oogx() && callCanTransferValue(this.opCode()) && !this.value().isZero())
            ? GasConstants.G_CALL_STIPEND.cost()
            : 0;
    final Bytes gasOopkt =
        this.oogx()
            ? Bytes.EMPTY
            : bigIntegerToBytes(
                this.gas()
                    .orElseThrow()
                    .toUnsignedBigInteger()
                    .min(BigInteger.valueOf(get63of64GDiff())));

    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .stamp(Bytes.ofUnsignedInt(stamp))
          .ct(Bytes.of(ct))
          .ctMax(Bytes.of(ctMax))
          .instruction(UnsignedByte.of(this.opCode().byteValue()))
          .isCreate(false)
          .isCreate2(false)
          .isCall(this.opCode() == OpCode.CALL)
          .isCallcode(this.opCode() == OpCode.CALLCODE)
          .isDelegatecall(this.opCode() == OpCode.DELEGATECALL)
          .isStaticcall(this.opCode() == OpCode.STATICCALL)
          .gasHi(this.gas().orElseThrow().slice(0, 16))
          .gasLo(this.gas().orElseThrow().slice(16))
          .valHi(this.value().slice(0, 16))
          .valLo(this.value().slice(16))
          .exists(this.toExists().orElseThrow())
          .warm(this.toWarm().orElseThrow())
          .outOfGasException(this.oogx())
          .gasActual(Bytes.ofUnsignedLong(this.gasActual()))
          .gasMxp(Bytes.ofUnsignedLong(this.gasMxp()))
          .gasUpfront(Bytes.ofUnsignedLong(this.gasPrelim()))
          .gasOopkt(gasOopkt)
          .gasStipend(Bytes.ofUnsignedLong(gasStipend));

      switch (ct) {
        case 0 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(this.gasActual()))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.EMPTY) // we REQUIRE that the currently available gas is nonnegative
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        case 1 -> trace
            .arg1Hi(this.value().slice(0, 16))
            .arg1Lo(this.value().slice(16, 16))
            .arg2Lo(Bytes.EMPTY)
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.ISZERO.byteValue()))
            .resLo(Bytes.of(this.value().isZero() ? 1 : 0))
            .wcpFlag(callCanTransferValue(this.opCode()))
            .modFlag(false)
            .validateRow();
        case 2 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(this.gasActual()))
            .arg2Lo(Bytes.ofUnsignedLong(this.gasPrelim()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(Bytes.of(this.oogx() ? 1 : 0))
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
          // the following rows are only filled in if no out of gas exception
        case 3 -> trace
            .arg1Hi(Bytes.EMPTY)
            .arg1Lo(Bytes.ofUnsignedLong(getGDiff()))
            .arg2Lo(Bytes.of(64))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.DIV.byteValue()))
            .resLo(Bytes.ofUnsignedLong(getGDiffOver64()))
            .wcpFlag(false)
            .modFlag(true)
            .validateRow();
        case 4 -> trace
            .arg1Hi(this.gas().orElseThrow().slice(0, 16))
            .arg1Lo(this.gas().orElseThrow().slice(16, 16))
            .arg2Lo(Bytes.ofUnsignedLong(getGDiff() - getGDiffOver64()))
            .exogenousModuleInstruction(UnsignedByte.of(OpCode.LT.byteValue()))
            .resLo(
                Bytes.of(
                    this.gas()
                                .orElseThrow()
                                .toUnsignedBigInteger()
                                .compareTo(BigInteger.valueOf(get63of64GDiff()))
                            < 0
                        ? 1
                        : 0))
            .wcpFlag(true)
            .modFlag(false)
            .validateRow();
        default -> throw new IllegalArgumentException("counter too big, should be <=" + ctMax);
      }
    }
  }

  private int maxCt() {
    if (this.oogx) {
      return this.opCode.isCreate() ? 1 : 2;
    } else {
      return this.opCode.isCreate() ? 2 : 4;
    }
  }

  @Override
  protected int computeLineCount() {
    return 1 + this.maxCt();
  }
}
