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

package net.consensys.linea.zktracer.module.alu.mul;

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Module;
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

    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    // argument order is reversed ??
    final MulData data = new MulData(opCode, arg2, arg1);

    switch (data.getRegime()) {
      case EXPONENT_ZERO_RESULT -> trace(builder, data);

      case EXPONENT_NON_ZERO_RESULT -> {
        if (data.carryOn()) {
          data.update();
          trace(builder, data);
        }
      }

      case TRIVIAL_MUL, NON_TRIVIAL_MUL -> {
        data.setHsAndBits(UInt256.fromBytes(arg1), UInt256.fromBytes(arg2));
        trace(builder, data);
      }

      default -> throw new RuntimeException("regime not supported");
    }
    // TODO captureBlockEnd should be called from elsewhere - not within messageFrame
    //    captureBlockEnd();
    MulData finalZeroToTheZero = new MulData(OpCode.EXP, Bytes32.ZERO, Bytes32.ZERO);
    trace(builder, finalZeroToTheZero);
  }

  @Override
  public Object commit() {
    return new MulTrace(builder.build(), stamp);
  }

  private void trace(final Trace.TraceBuilder builder, final MulData data) {
    stamp++;

    for (int ct = 0; ct < data.maxCt(); ct++) {
      trace(builder, data, ct);
    }
  }

  private void trace(final Trace.TraceBuilder builder, final MulData data, final int i) {
    builder
        .mulStampArg(stamp)
        .counterArg(i)
        .oneLineInstructionArg(data.isOneLineInstruction())
        .tinyBaseArg(data.tinyBase)
        .tinyExponentArg(data.tinyExponent)
        .resultVanishesArg(data.res.isZero())
        .instArg(UnsignedByte.of(data.opCode.value))
        .arg1HiArg(data.arg1Hi.toUnsignedBigInteger())
        .arg1LoArg(data.arg1Lo.toUnsignedBigInteger())
        .arg2HiArg(data.arg2Hi.toUnsignedBigInteger())
        .arg2LoArg(data.arg2Lo.toUnsignedBigInteger())
        .resHiArg(data.res.getHigh().toUnsignedBigInteger())
        .resLoArg(data.res.getLow().toUnsignedBigInteger())
        .bitsArg(data.bits[i])
        .byteA3Arg(UnsignedByte.of(data.aBytes.get(3, i)))
        .byteA2Arg(UnsignedByte.of(data.aBytes.get(2, i)))
        .byteA1Arg(UnsignedByte.of(data.aBytes.get(1, i)))
        .byteA0Arg(UnsignedByte.of(data.aBytes.get(0, i)))
        .accA3Arg(data.aBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accA2Arg(data.aBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accA1Arg(data.aBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accA0Arg(data.aBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteB3Arg(UnsignedByte.of(data.bBytes.get(3, i)))
        .byteB2Arg(UnsignedByte.of(data.bBytes.get(2, i)))
        .byteB1Arg(UnsignedByte.of(data.bBytes.get(1, i)))
        .byteB0Arg(UnsignedByte.of(data.bBytes.get(0, i)))
        .accB3Arg(data.bBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accB2Arg(data.bBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accB1Arg(data.bBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accB0Arg(data.bBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteC3Arg(UnsignedByte.of(data.cBytes.get(3, i)))
        .byteC2Arg(UnsignedByte.of(data.cBytes.get(2, i)))
        .byteC1Arg(UnsignedByte.of(data.cBytes.get(1, i)))
        .byteC0Arg(UnsignedByte.of(data.cBytes.get(0, i)))
        .accC3Arg(data.cBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accC2Arg(data.cBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accC1Arg(data.cBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accC0Arg(data.cBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteH3Arg(UnsignedByte.of(data.hBytes.get(3, i)))
        .byteH2Arg(UnsignedByte.of(data.hBytes.get(2, i)))
        .byteH1Arg(UnsignedByte.of(data.hBytes.get(1, i)))
        .byteH0Arg(UnsignedByte.of(data.hBytes.get(0, i)))
        .accH3Arg(data.hBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accH2Arg(data.hBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accH1Arg(data.hBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accH0Arg(data.hBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .exponentBitArg(data.exponentBit())
        .exponentBitAccumulatorArg(data.expAcc.toUnsignedBigInteger())
        .exponentBitSourceArg(data.exponentSource())
        .squareAndMultiplyArg(data.snm)
        .bitNumArg(data.getBitNum());
  }
}
