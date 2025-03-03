/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.blake2fmodexpdata;

import java.util.List;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.OperationListModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedList;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.limits.precompiles.BlakeEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.BlakeRounds;
import net.consensys.linea.zktracer.module.limits.precompiles.ModexpEffectiveCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;

@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public class BlakeModexpData implements OperationListModule<BlakeModexpDataOperation> {
  private final Wcp wcp;
  private final ModexpEffectiveCall modexpEffectiveCall;
  private final BlakeEffectiveCall blakeEffectiveCall;
  private final BlakeRounds blakeRounds;

  private final ModuleOperationStackedList<BlakeModexpDataOperation> operations =
      new ModuleOperationStackedList<>();

  private long previousID = 0;

  @Override
  public String moduleKey() {
    return "BLAKE_MODEXP_DATA";
  }

  public void callModexp(final ModexpMetadata modexpMetaData, final int operationID) {
    operations.add(new BlakeModexpDataOperation(modexpMetaData, operationID));
    modexpEffectiveCall.updateTally(1);
    callWcpForIdCheck(operationID);
  }

  public void callBlake(final BlakeComponents blakeComponents, final int operationID) {
    operations.add(new BlakeModexpDataOperation(blakeComponents, operationID));
    blakeEffectiveCall.updateTally(1);
    blakeRounds.addPrecompileLimit(blakeComponents.r());
    callWcpForIdCheck(operationID);
  }

  private void callWcpForIdCheck(final int operationID) {
    wcp.callLT(previousID, operationID);
    previousID = operationID;
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders() {
    return Trace.Blake2fmodexpdata.headers(this.lineCount());
  }

  @Override
  public void commit(Trace trace) {
    int stamp = 0;
    for (BlakeModexpDataOperation o : operations.getAll()) {
      o.trace(trace.blake2fmodexpdata, ++stamp);
    }
  }
}
