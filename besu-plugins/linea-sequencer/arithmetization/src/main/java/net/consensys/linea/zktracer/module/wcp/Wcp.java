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

package net.consensys.linea.zktracer.module.wcp;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

@RequiredArgsConstructor
public class Wcp implements Module {

  private final StackedSet<WcpOperation> operations = new StackedSet<>();
  private int stamp = 0;

  private final Hub hub;

  @Override
  public String moduleKey() {
    return "WCP";
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
    final OpCode opcode = this.hub.opCode();
    final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
    final Bytes32 arg2 =
        (opcode != OpCode.ISZERO) ? Bytes32.leftPad(frame.getStackItem(1)) : Bytes32.ZERO;

    this.operations.add(new WcpOperation(opcode, arg1, arg2));
  }

  public void traceWcpOperation(WcpOperation op, Trace trace) {
    this.stamp++;
    final Bytes resHi = op.getResHi() ? Bytes.of(1) : Bytes.EMPTY;
    final Bytes resLo = op.getResLo() ? Bytes.of(1) : Bytes.EMPTY;
    final List<Boolean> bits = op.getBits();
    final boolean neg1 = op.getNeg1();
    final boolean neg2 = op.getNeg2();
    final Bytes16 adjHi = op.getAdjHi();
    final Bytes16 adjLo = op.getAdjLo();
    final boolean bit1 = op.getBit1();
    final boolean bit2 = op.getBit2();
    final boolean bit3 = op.getBit3();
    final boolean bit4 = op.getBit4();
    for (int i = 0; i < op.maxCt(); i++) {
      trace
          .wordComparisonStamp(Bytes.ofUnsignedInt(stamp))
          .oneLineInstruction(op.isOneLineInstruction())
          .counter(Bytes.of(i))
          .inst(Bytes.of(op.getOpCode().byteValue()))
          .argument1Hi(op.getArg1Hi())
          .argument1Lo(op.getArg1Lo())
          .argument2Hi(op.getArg2Hi())
          .argument2Lo(op.getArg2Lo())
          .resultHi(resHi)
          .resultLo(resLo)
          .bits(bits.get(i))
          .neg1(neg1)
          .neg2(neg2)
          .byte1(UnsignedByte.of(op.getArg1Hi().get(i)))
          .byte2(UnsignedByte.of(op.getArg1Lo().get(i)))
          .byte3(UnsignedByte.of(op.getArg2Hi().get(i)))
          .byte4(UnsignedByte.of(op.getArg2Lo().get(i)))
          .byte5(UnsignedByte.of(adjHi.get(i)))
          .byte6(UnsignedByte.of(adjLo.get(i)))
          .acc1(op.getArg1Hi().slice(0, 1 + i))
          .acc2(op.getArg1Lo().slice(0, 1 + i))
          .acc3(op.getArg2Hi().slice(0, 1 + i))
          .acc4(op.getArg2Lo().slice(0, 1 + i))
          .acc5(adjHi.slice(0, 1 + i))
          .acc6(adjLo.slice(0, 1 + i))
          .bit1(bit1)
          .bit2(bit2)
          .bit3(bit3)
          .bit4(bit4)
          .validateRow();
    }
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    for (WcpOperation operation : this.operations) {
      this.traceWcpOperation(operation, trace);
    }
  }

  @Override
  public int lineCount() {
    return this.operations.stream().mapToInt(WcpOperation::maxCt).sum();
  }

  public boolean callLT(Bytes32 arg1, Bytes32 arg2) {
    this.operations.add(new WcpOperation(OpCode.LT, arg1, arg2));
    return arg1.compareTo(arg2) < 0;
  }

  public boolean callLT(Bytes arg1, Bytes arg2) {
    return this.callLT(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callEQ(Bytes32 arg1, Bytes32 arg2) {
    this.operations.add(new WcpOperation(OpCode.EQ, arg1, arg2));
    return arg1.compareTo(arg2) == 0;
  }

  public boolean callEQ(Bytes arg1, Bytes arg2) {
    return this.callEQ(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callISZERO(Bytes32 arg1) {
    this.operations.add(new WcpOperation(OpCode.ISZERO, arg1, Bytes32.ZERO));
    return arg1.isZero();
  }
}
