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

import static net.consensys.linea.zktracer.module.wcp.WcpOperation.GEQbv;
import static net.consensys.linea.zktracer.module.wcp.WcpOperation.LEQbv;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

@RequiredArgsConstructor
public class Wcp implements Module {
  private final StackedSet<WcpOperation> operations = new StackedSet<>();

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

    this.operations.add(new WcpOperation(opcode.byteValue(), arg1, arg2));
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;
    for (WcpOperation operation : this.operations) {
      stamp++;
      operation.trace(trace, stamp);
    }
  }

  @Override
  public int lineCount() {
    return this.operations.lineCount();
  }

  public boolean callLT(Bytes32 arg1, Bytes32 arg2) {
    this.operations.add(new WcpOperation(OpCode.LT.byteValue(), arg1, arg2));
    return arg1.compareTo(arg2) < 0;
  }

  public boolean callLT(Bytes arg1, Bytes arg2) {
    return this.callLT(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callEQ(Bytes32 arg1, Bytes32 arg2) {
    this.operations.add(new WcpOperation(OpCode.EQ.byteValue(), arg1, arg2));
    return arg1.compareTo(arg2) == 0;
  }

  public boolean callEQ(Bytes arg1, Bytes arg2) {
    return this.callEQ(Bytes32.leftPad(arg1), Bytes32.leftPad(arg2));
  }

  public boolean callISZERO(Bytes32 arg1) {
    this.operations.add(new WcpOperation(OpCode.ISZERO.byteValue(), arg1, Bytes32.ZERO));
    return arg1.isZero();
  }

  public boolean callISZERO(Bytes arg1) {
    return this.callISZERO(Bytes32.leftPad(arg1));
  }

  public boolean callLEQ(Bytes32 arg1, Bytes32 arg2) {
    this.operations.add(new WcpOperation(LEQbv, arg1, arg2));
    return arg1.compareTo(arg2) <= 0;
  }

  public boolean callGEQ(Bytes32 arg1, Bytes32 arg2) {
    this.operations.add(new WcpOperation(GEQbv, arg1, arg2));
    return arg1.compareTo(arg2) >= 0;
  }
}
