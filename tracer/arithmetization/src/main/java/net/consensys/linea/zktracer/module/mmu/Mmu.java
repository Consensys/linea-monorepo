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

package net.consensys.linea.zktracer.module.mmu;

import java.nio.MappedByteBuffer;
import java.util.List;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationListModule;
import net.consensys.linea.zktracer.container.stacked.StackedList;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.mmu.values.HubToMmuValues;
import net.consensys.linea.zktracer.module.wcp.Wcp;

@RequiredArgsConstructor
@Accessors(fluent = true)
public class Mmu implements OperationListModule<MmuOperation> {
  @Getter private final StackedList<MmuOperation> operations = new StackedList<>();
  private final Euc euc;
  private final Wcp wcp;

  @Override
  public String moduleKey() {
    return "MMU";
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int mmuStamp = 0;
    int mmioStamp = 0;

    for (MmuOperation mmuOperation : operations.getAll()) {
      Preconditions.checkState(mmuOperation.traceMe(), "Cannot compute if traceMe is false");
      if (mmuOperation.traceMe()) {
        mmuOperation.getCFI();
        mmuOperation.fillLimb();

        mmuStamp += 1;
        mmuOperation.trace(mmuStamp, mmioStamp, trace);
        mmioStamp += mmuOperation.mmuData().numberMmioInstructions();
      }
    }
  }

  public void call(final MmuCall mmuCall) {
    Preconditions.checkState(mmuCall.traceMe(), "Shouldn't compute if traceMe is false");
    MmuData mmuData = new MmuData(mmuCall);
    mmuData.hubToMmuValues(
        HubToMmuValues.fromMmuCall(mmuCall, mmuData.exoLimbIsSource(), mmuData.exoLimbIsTarget()));

    final MmuInstructions mmuInstructions = new MmuInstructions(euc, wcp);
    mmuData = mmuInstructions.compute(mmuData);

    operations.add(new MmuOperation(mmuData));
  }
}
