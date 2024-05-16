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

package net.consensys.linea.zktracer.module.exp;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.exp.ExpCallForExpPricing;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.exp.ExpCallForModexpLogComputation;
import net.consensys.linea.zktracer.module.wcp.Wcp;

@Slf4j
@RequiredArgsConstructor
public class Exp implements Module {
  /** A list of the operations to trace */
  private final StackedSet<ExpOperation> chunks = new StackedSet<>();

  private final Wcp wcp;

  @Override
  public String moduleKey() {
    return "EXP";
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

  public void callExpLogCall(final ExpCallForExpPricing c) {
    this.chunks.add(ExpLogOperation.fromExpLogCall(this.wcp, c));
  }

  public void callModExpLogCall(final ExpCallForModexpLogComputation c) {
    this.chunks.add(ModexpLogOperation.fromExpLogCall(this.wcp, c));
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;

    for (ExpOperation op : this.chunks) {
      stamp += 1;
      op.traceComputation(stamp, trace);
      op.traceMacro(stamp, trace);
      op.tracePreprocessing(stamp, trace);
    }
  }
}
