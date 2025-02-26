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

package net.consensys.linea.zktracer.module.add;

import static net.consensys.linea.zktracer.opcode.OpCode.ADD;
import static net.consensys.linea.zktracer.opcode.OpCode.SUB;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

/** Implementation of a {@link Module} for addition/subtraction. */
@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public class Add implements OperationSetModule<AddOperation> {

  private final ModuleOperationStackedSet<AddOperation> operations =
      new ModuleOperationStackedSet<>();

  @Override
  public String moduleKey() {
    return "ADD";
  }

  @Override
  public void tracePreOpcode(MessageFrame frame, OpCode opcode) {
    if ((opcode == ADD || opcode == SUB)) {
      operations.add(
          new AddOperation(
              opcode,
              Bytes32.leftPad(frame.getStackItem(0)),
              Bytes32.leftPad(frame.getStackItem(1))));
    }
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    int stamp = 0;
    for (AddOperation op : sortOperations(new AddOperationComparator())) {
      op.trace(++stamp, trace);
    }
  }

  public BigInteger callADD(Bytes32 arg1, Bytes32 arg2) {
    operations.add(new AddOperation(ADD, arg1, arg2));
    return arg1.toUnsignedBigInteger().add(arg2.toUnsignedBigInteger());
  }
}
