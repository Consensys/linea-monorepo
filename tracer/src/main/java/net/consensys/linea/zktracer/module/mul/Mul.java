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

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Mul implements Module {
  final Trace.TraceBuilder trace = Trace.builder();
  /** A set of the operations to trace */
  private final StackedSet<MulOperation> operations = new StackedSet<>();

  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "mul";
  }

  @Override
  public void tracePreOpcode(MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.leftPad(frame.getStackItem(1));

    operations.add(new MulOperation(opCode, arg1, arg2));
  }

  @Override
  public void enterTransaction() {
    this.operations.enter();
  }

  @Override
  public void popTransaction() {
    this.operations.pop();
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    this.operations.enter();
  }

  @Override
  public ModuleTrace commit() {
    for (var op : this.operations) {
      this.traceMulOperation(op);
    }
    this.traceMulOperation(new MulOperation(OpCode.EXP, Bytes32.ZERO, Bytes32.ZERO));

    return new MulTrace(trace.build());
  }

  private void traceMulOperation(final MulOperation op) {
    this.stamp++;

    switch (op.getRegime()) {
      case EXPONENT_ZERO_RESULT -> traceSubOp(op);

      case EXPONENT_NON_ZERO_RESULT -> {
        while (op.carryOn()) {
          op.update();
          traceSubOp(op);
        }
      }

      case TRIVIAL_MUL, NON_TRIVIAL_MUL -> {
        op.setHsAndBits(UInt256.fromBytes(op.getArg1()), UInt256.fromBytes(op.getArg2()));
        traceSubOp(op);
      }

      default -> throw new RuntimeException("regime not supported");
    }
  }

  private void traceSubOp(final MulOperation data) {
    for (int ct = 0; ct < data.maxCt(); ct++) {
      traceRow(data, ct);
    }
  }

  private void traceRow(final MulOperation op, final int i) {
    trace
        .mulStamp(BigInteger.valueOf(stamp))
        .counter(BigInteger.valueOf(i))
        .oli(op.isOneLineInstruction())
        .tinyBase(op.isTinyBase())
        .tinyExponent(op.isTinyExponent())
        .resultVanishes(op.res.isZero())
        .instruction(BigInteger.valueOf(op.getOpCode().getData().value()))
        .arg1Hi(op.getArg1Hi().toUnsignedBigInteger())
        .arg1Lo(op.getArg1Lo().toUnsignedBigInteger())
        .arg2Hi(op.getArg2Hi().toUnsignedBigInteger())
        .arg2Lo(op.getArg2Lo().toUnsignedBigInteger())
        .resHi(op.res.getHigh().toUnsignedBigInteger())
        .resLo(op.res.getLow().toUnsignedBigInteger())
        .bits(op.bits[i])
        .byteA3(UnsignedByte.of(op.aBytes.get(3, i)))
        .byteA2(UnsignedByte.of(op.aBytes.get(2, i)))
        .byteA1(UnsignedByte.of(op.aBytes.get(1, i)))
        .byteA0(UnsignedByte.of(op.aBytes.get(0, i)))
        .accA3(op.aBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accA2(op.aBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accA1(op.aBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accA0(op.aBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteB3(UnsignedByte.of(op.bBytes.get(3, i)))
        .byteB2(UnsignedByte.of(op.bBytes.get(2, i)))
        .byteB1(UnsignedByte.of(op.bBytes.get(1, i)))
        .byteB0(UnsignedByte.of(op.bBytes.get(0, i)))
        .accB3(op.bBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accB2(op.bBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accB1(op.bBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accB0(op.bBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteC3(UnsignedByte.of(op.cBytes.get(3, i)))
        .byteC2(UnsignedByte.of(op.cBytes.get(2, i)))
        .byteC1(UnsignedByte.of(op.cBytes.get(1, i)))
        .byteC0(UnsignedByte.of(op.cBytes.get(0, i)))
        .accC3(op.cBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accC2(op.cBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accC1(op.cBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accC0(op.cBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .byteH3(UnsignedByte.of(op.hBytes.get(3, i)))
        .byteH2(UnsignedByte.of(op.hBytes.get(2, i)))
        .byteH1(UnsignedByte.of(op.hBytes.get(1, i)))
        .byteH0(UnsignedByte.of(op.hBytes.get(0, i)))
        .accH3(op.hBytes.getRange(3, 0, i + 1).toUnsignedBigInteger())
        .accH2(op.hBytes.getRange(2, 0, i + 1).toUnsignedBigInteger())
        .accH1(op.hBytes.getRange(1, 0, i + 1).toUnsignedBigInteger())
        .accH0(op.hBytes.getRange(0, 0, i + 1).toUnsignedBigInteger())
        .exponentBit(op.isExponentBitSet())
        .exponentBitAccumulator(op.expAcc.toUnsignedBigInteger())
        .exponentBitSource(op.isExponentInSource())
        .squareAndMultiply(op.squareAndMultiply)
        .bitNum(BigInteger.valueOf(op.getBitNum()))
        .validateRow();
  }

  @Override
  public int lineCount() {
    return 1
        + this.operations.stream()
            .map(MulOperation::clone) // The counting operation is destructive, hence the clone
            .mapToInt(
                op ->
                    switch (op.getRegime()) {
                      case EXPONENT_ZERO_RESULT -> op.maxCt();

                      case EXPONENT_NON_ZERO_RESULT -> {
                        int r = 0;
                        while (op.carryOn()) {
                          op.update();
                          r += op.maxCt();
                        }
                        yield r;
                      }

                      case TRIVIAL_MUL, NON_TRIVIAL_MUL -> {
                        op.setHsAndBits(
                            UInt256.fromBytes(op.getArg1()), UInt256.fromBytes(op.getArg2()));
                        yield op.maxCt();
                      }

                      default -> throw new RuntimeException("regime not supported");
                    })
            .sum();
  }
}
