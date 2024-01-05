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

package net.consensys.linea.zktracer.module.mxp;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.hyperledger.besu.evm.frame.MessageFrame;

/** Implementation of a {@link Module} for memory expansion. */
public class Mxp implements Module {
  /** A list of the operations to trace */
  private final StackedList<MxpData> chunks = new StackedList<>();

  private Hub hub;

  @Override
  public String moduleKey() {
    return "MXP";
  }

  public Mxp(Hub hub) {
    this.hub = hub;
  }

  // TODO: update tests and eliminate this constructor
  public Mxp() {}

  @Override
  public void tracePreOpcode(MessageFrame frame) { // This will be renamed to tracePreOp
    this.chunks.add(new MxpData(frame, hub));
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
  public int lineCount() {
    return this.chunks.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    for (int i = 0; i < this.chunks.size(); i++) {
      this.chunks.get(i).trace(i + 1, trace);
    }
  }

  @Override
  public void tracePostOp(MessageFrame frame) { // This is paired with tracePreOp
  }
}
