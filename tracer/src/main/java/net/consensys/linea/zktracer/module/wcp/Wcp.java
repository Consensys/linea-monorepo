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

package net.consensys.linea.zktracer.module.wcp;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Wcp implements Module {
  final Trace.TraceBuilder builder = Trace.builder();
  private final StackedSet<WcpOperation> operations = new StackedSet<>();
  private int stamp = 0;

  @Override
  public String jsonKey() {
    return "wcp";
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
  public void tracePreOpcode(final MessageFrame frame) {
    final OpCodeData opCode = OpCodes.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
    final Bytes32 arg2 =
        (opCode.mnemonic() != OpCode.ISZERO)
            ? Bytes32.leftPad(frame.getStackItem(1))
            : Bytes32.ZERO;

    this.operations.add(new WcpOperation(opCode, arg1, arg2));
  }

  public void traceWcpOperation(WcpOperation op) {
    this.stamp++;
    for (int i = 0; i < op.maxCt(); i++) {
      builder
          .wordComparisonStamp(BigInteger.valueOf(stamp))
          .oneLineInstruction(op.isOneLineInstruction())
          .counter(BigInteger.valueOf(i))
          .inst(BigInteger.valueOf(op.getOpCode().getData().value()))
          .argument1Hi(op.getArg1Hi().toUnsignedBigInteger())
          .argument1Lo(op.getArg1Lo().toUnsignedBigInteger())
          .argument2Hi(op.getArg2Hi().toUnsignedBigInteger())
          .argument2Lo(op.getArg2Lo().toUnsignedBigInteger())
          .resultHi(op.getResHi() ? BigInteger.ONE : BigInteger.ZERO)
          .resultLo(op.getResLo() ? BigInteger.ONE : BigInteger.ZERO)
          .bits(op.getBits().get(i))
          .neg1(op.getNeg1())
          .neg2(op.getNeg2())
          .byte1(UnsignedByte.of(op.getArg1Hi().get(i)))
          .byte2(UnsignedByte.of(op.getArg1Lo().get(i)))
          .byte3(UnsignedByte.of(op.getArg2Hi().get(i)))
          .byte4(UnsignedByte.of(op.getArg2Lo().get(i)))
          .byte5(UnsignedByte.of(op.getAdjHi().get(i)))
          .byte6(UnsignedByte.of(op.getAdjLo().get(i)))
          .acc1(op.getArg1Hi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc2(op.getArg1Lo().slice(0, 1 + i).toUnsignedBigInteger())
          .acc3(op.getArg2Hi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc4(op.getArg2Lo().slice(0, 1 + i).toUnsignedBigInteger())
          .acc5(op.getAdjHi().slice(0, 1 + i).toUnsignedBigInteger())
          .acc6(op.getAdjLo().slice(0, 1 + i).toUnsignedBigInteger())
          .bit1(op.getBit1())
          .bit2(op.getBit2())
          .bit3(op.getBit3())
          .bit4(op.getBit4())
          .validateRow();
    }
  }

  @Override
  public ModuleTrace commit() {
    for (WcpOperation operation : this.operations) {
      this.traceWcpOperation(operation);
    }

    return new WcpTrace(builder.build());
  }

  @Override
  public int lineCount() {
    return this.operations.stream().mapToInt(WcpOperation::maxCt).sum();
  }

  public void callLT(Bytes32 arg1, Bytes32 arg2) {
    WcpOperation data = new WcpOperation(OpCode.LT, arg1, arg2);
    this.traceWcpOperation(data);
  }

  public void callEQ(Bytes32 arg1, Bytes32 arg2) {
    WcpOperation data = new WcpOperation(OpCode.EQ, arg1, arg2);
    this.traceWcpOperation(data);
  }

  public void callISZERO(Bytes32 arg1) {
    Bytes32 zero = Bytes32.repeat((byte) 0x00);
    WcpOperation data = new WcpOperation(OpCode.ISZERO, arg1, zero);
    this.traceWcpOperation(data);
  }
}
