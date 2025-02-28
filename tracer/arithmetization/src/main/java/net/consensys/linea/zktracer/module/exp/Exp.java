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

import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.exp.ExpCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;

@Slf4j
@RequiredArgsConstructor
@Accessors(fluent = true)
public class Exp implements OperationSetModule<ExpOperation> {
  private final Hub hub;
  private final Wcp wcp;

  @Getter
  private final ModuleOperationStackedSet<ExpOperation> operations =
      new ModuleOperationStackedSet<>();

  @Override
  public String moduleKey() {
    return "EXP";
  }

  public void call(ExpCall expCall) {
    operations.add(new ExpOperation(expCall, wcp, hub));
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders() {
    return Trace.Exp.headers(this.lineCount());
  }

  @Override
  public void commit(Trace trace) {
    int stamp = 0;
    for (ExpOperation expOp : operations.sortOperations(new ExpOperationComparator())) {
      expOp.traceComputation(++stamp, trace.exp);
      expOp.traceMacro(stamp, trace.exp);
      expOp.tracePreprocessing(stamp, trace.exp);
    }
  }
}
