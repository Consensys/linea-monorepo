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

package net.consensys.linea.zktracer.module.mod;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Mod implements Module {
  @Override
  public String moduleKey() {
    return "MOD";
  }

  private final StackedSet<ModOperation> chunks = new StackedSet<>();

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    final OpCodeData opCodeData = OpCodes.of(frame.getCurrentOperation().getOpcode());
    final Bytes32 arg1 = Bytes32.leftPad(frame.getStackItem(0));
    final Bytes32 arg2 = Bytes32.leftPad(frame.getStackItem(1));

    this.chunks.add(new ModOperation(opCodeData, arg1, arg2));
  }

  @Override
  public void enterTransaction() {
    this.chunks.enter();
  }

  @Override
  public void popTransaction() {
    this.chunks.pop();
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    this.chunks.enter();
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    int stamp = 0;
    for (ModOperation op : this.chunks) {
      stamp++;
      op.trace(trace, stamp);
    }
  }

  @Override
  public int lineCount() {
    return this.chunks.stream().mapToInt(ModOperation::maxCounter).sum();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  /**
   * Performs a context-free call to the DIV opcode in the current trace.
   *
   * @param arg1 the divider
   * @param arg2 the dividend
   */
  public void callDiv(Bytes32 arg1, Bytes32 arg2) {
    this.chunks.add(new ModOperation(OpCode.DIV, arg1, arg2));
  }

  /**
   * Performs a context-free call to the MOD opcode in the current trace.
   *
   * @param arg1 the number
   * @param arg2 the module
   */
  public void callMod(Bytes32 arg1, Bytes32 arg2) {
    this.chunks.add(new ModOperation(OpCode.MOD, arg1, arg2));
  }
}
