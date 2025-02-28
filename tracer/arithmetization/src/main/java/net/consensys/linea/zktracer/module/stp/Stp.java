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

package net.consensys.linea.zktracer.module.stp;

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.types.Conversions.longToBytes32;

import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.OperationSetModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedSet;
import net.consensys.linea.zktracer.module.hub.fragment.imc.StpCall;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes32;

@RequiredArgsConstructor
@Accessors(fluent = true)
public class Stp implements OperationSetModule<StpOperation> {

  private final Wcp wcp;
  private final Mod mod;

  @Getter
  private final ModuleOperationStackedSet<StpOperation> operations =
      new ModuleOperationStackedSet<>();

  public void call(StpCall stpCall) {
    final StpOperation stpOperation = new StpOperation(stpCall);
    operations.add(stpOperation);

    checkArgument(
        stpCall.opCode().isCall() || stpCall.opCode().isCreate(),
        "STP handles only Calls and CREATEs");

    if (stpCall.opCode().isCreate()) {
      wcp.callLT(longToBytes32(stpCall.gasActual()), Bytes32.ZERO);
      wcp.callLT(longToBytes32(stpCall.gasActual()), longToBytes32(stpCall.upfrontGasCost()));
      if (!stpCall.outOfGasException()) {
        mod.callDIV(longToBytes32(stpOperation.getGDiff()), longToBytes32(64L));
      }
    }

    if (stpCall.opCode().isCall()) {
      wcp.callLT(longToBytes32(stpCall.gasActual()), Bytes32.ZERO);
      if (stpCall.opCode().callHasValueArgument()) {
        wcp.callISZERO(Bytes32.leftPad(stpCall.value()));
      }
      wcp.callLT(longToBytes32(stpCall.gasActual()), longToBytes32(stpCall.upfrontGasCost()));
      if (!stpCall.outOfGasException()) {
        mod.callDIV(longToBytes32(stpOperation.getGDiff()), longToBytes32(64L));
        wcp.callLT(stpCall.gas(), longToBytes32(stpOperation.get63of64GDiff()));
      }
    }
  }

  @Override
  public String moduleKey() {
    return "STP";
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders() {
    return Trace.Stp.headers(this.lineCount());
  }

  @Override
  public void commit(Trace trace) {
    int stamp = 0;
    for (StpOperation operation : operations.sortOperations(new StpOperationComparator())) {
      operation.trace(trace.stp, ++stamp);
    }
  }
}
