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

package net.consensys.linea.zktracer.module.shf;

import static net.consensys.linea.zktracer.opcode.OpCode.*;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;

@Getter
@Accessors(fluent = true)
public class Shf implements OperationSetModule<ShfOperation> {

  private final ModuleOperationStackedSet<ShfOperation> operations =
      new ModuleOperationStackedSet<>();

  @Override
  public String moduleKey() {
    return "SHF";
  }

  @Override
  public void tracePreOpcode(MessageFrame frame, OpCode opcode) {
    if (opcode == SHL || opcode == SHR || opcode == SAR) {

      final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
      final Bytes32 arg2 = Bytes32.leftPad(frame.getStackItem(1));
      operations.add(new ShfOperation(opcode, arg1, arg2));
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
    for (ShfOperation op : operations.sortOperations(new ShfOperationComparator())) {
      op.trace(trace, ++stamp);
    }
  }
}
