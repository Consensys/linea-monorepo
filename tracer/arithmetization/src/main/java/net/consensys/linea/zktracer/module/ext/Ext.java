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

package net.consensys.linea.zktracer.module.ext;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

@RequiredArgsConstructor
public class Ext implements Module {
  private final Hub hub;

  /** A set of the operations to trace */
  private final StackedSet<ExtOperation> operations = new StackedSet<>();

  @Override
  public String moduleKey() {
    return "EXT";
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
    final OpCodeData opCode = hub.opCodeData();
    this.operations.add(
        new ExtOperation(
            opCode.mnemonic(),
            Bytes32.leftPad(frame.getStackItem(0)),
            Bytes32.leftPad(frame.getStackItem(1)),
            Bytes32.leftPad(frame.getStackItem(2))));
  }

  public Bytes call(OpCode opCode, Bytes _arg1, Bytes _arg2, Bytes _arg3) {
    final Bytes32 arg1 = Bytes32.leftPad(_arg1);
    final Bytes32 arg2 = Bytes32.leftPad(_arg2);
    final Bytes32 arg3 = Bytes32.leftPad(_arg3);
    final ExtOperation op = new ExtOperation(opCode, arg1, arg2, arg3);
    final Bytes result = op.compute();
    this.operations.add(op);
    return result;
  }

  public Bytes callADDMOD(Bytes _arg1, Bytes _arg2, Bytes _arg3) {
    return this.call(OpCode.ADDMOD, _arg1, _arg2, _arg3);
  }

  public Bytes callMULMOD(Bytes _arg1, Bytes _arg2, Bytes _arg3) {
    return this.call(OpCode.MULMOD, _arg1, _arg2, _arg3);
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;
    for (ExtOperation operation : this.operations) {
      stamp++;
      operation.trace(trace, stamp);
    }
  }

  @Override
  public int lineCount() {
    return this.operations.lineCount();
  }
}
