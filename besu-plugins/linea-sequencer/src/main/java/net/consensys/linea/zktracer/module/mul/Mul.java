/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.mul;

import java.math.BigInteger;
import java.util.List;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Mul implements Module {
  final Trace.TraceBuilder builder = Trace.builder();

  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "mul";
  }

  @Override
  public final List<OpCode> supportedOpCodes() {
    return List.of(OpCode.MUL, OpCode.EXP);
  }

  @SuppressWarnings("UnusedVariable")
  @Override
  public void trace(MessageFrame frame) {
    final Bytes32 arg1 = Bytes32.wrap(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.wrap(frame.getStackItem(1));

    final OpCodeData opCode = OpCodes.of(frame.getCurrentOperation().getOpcode());

    // argument order is reversed ??
    final MulData data = new MulData(opCode, arg1, arg2);

    stamp++;
    switch (data.getRegime()) {
      case EXPONENT_ZERO_RESULT -> traceSubOp(builder, data);

      case EXPONENT_NON_ZERO_RESULT -> {
        while (data.carryOn()) {
          data.update();
          traceSubOp(builder, data);
        }
      }

      case TRIVIAL_MUL, NON_TRIVIAL_MUL -> {
        data.setHsAndBits(UInt256.fromBytes(arg1), UInt256.fromBytes(arg2));
        traceSubOp(builder, data);
      }

      default -> throw new RuntimeException("regime not supported");
    }
  }

  @Override
  public void traceEndConflation() {
    stamp++;
    MulData finalZeroToTheZero = new MulData(OpCode.EXP, Bytes32.ZERO, Bytes32.ZERO);
    traceSubOp(builder, finalZeroToTheZero);
  }

  @Override
  public Object commit() {
    return new MulTrace(builder.build());
  }

  private void traceSubOp(final Trace.TraceBuilder builder, final MulData data) {
    for (int ct = 0; ct < data.maxCt(); ct++) {
      traceRow(builder, data, ct);
    }
  }

  private void traceRow(final Trace.TraceBuilder builder, final MulData data, final int i) {
    builder
        .mulStamp(BigInteger.valueOf(stamp))
        .counter(BigInteger.valueOf(i))
        .oli(data.isOneLineInstruction())
        .tinyBase(data.isTinyBase())
        .tinyExponent(data.isTinyExponent())
        .resultVanishes(data.res.isZero())
        .instruction(BigInteger.valueOf(data.getOpCode().getData().value()))
        .arg1Hi(data.getArg1Hi().toUnsignedBigInteger())
        .arg1Lo(data.getArg1Lo().toUnsignedBigInteger())
        .arg2Hi(data.getArg2Hi().toUnsignedBigInteger())
        .arg2Lo(data.getArg2Lo().toUnsignedBigInteger())
        .resHi(data.res.getHigh().toUnsignedBigInteger())
        .resLo(data.res.getLow().toUnsignedBigInteger())
        .bits(data.bits[i])
        .byteA3(UnsignedByte.of(data.aBytes.get(3, i)))
        .byteA2(UnsignedByte.of(data.aBytes.get(2, i)))
        .byteA1(UnsignedByte.of(data.aBytes.get(1, i)))
        .byteA0(UnsignedByte.of(data.aBytes.get(0, i)))
        .accA3(data.aBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accA2(data.aBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accA1(data.aBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accA0(data.aBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteB3(UnsignedByte.of(data.bBytes.get(3, i)))
        .byteB2(UnsignedByte.of(data.bBytes.get(2, i)))
        .byteB1(UnsignedByte.of(data.bBytes.get(1, i)))
        .byteB0(UnsignedByte.of(data.bBytes.get(0, i)))
        .accB3(data.bBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accB2(data.bBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accB1(data.bBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accB0(data.bBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteC3(UnsignedByte.of(data.cBytes.get(3, i)))
        .byteC2(UnsignedByte.of(data.cBytes.get(2, i)))
        .byteC1(UnsignedByte.of(data.cBytes.get(1, i)))
        .byteC0(UnsignedByte.of(data.cBytes.get(0, i)))
        .accC3(data.cBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accC2(data.cBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accC1(data.cBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accC0(data.cBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteH3(UnsignedByte.of(data.hBytes.get(3, i)))
        .byteH2(UnsignedByte.of(data.hBytes.get(2, i)))
        .byteH1(UnsignedByte.of(data.hBytes.get(1, i)))
        .byteH0(UnsignedByte.of(data.hBytes.get(0, i)))
        .accH3(data.hBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accH2(data.hBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accH1(data.hBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accH0(data.hBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .exponentBit(data.isExponentBitSet())
        .exponentBitAccumulator(data.expAcc.toUnsignedBigInteger())
        .exponentBitSource(data.isExponentInSource())
        .squareAndMultiply(data.squareAndMultiply)
        .bitNum(BigInteger.valueOf(data.getBitNum()))
        .validateRow();
  }
}
