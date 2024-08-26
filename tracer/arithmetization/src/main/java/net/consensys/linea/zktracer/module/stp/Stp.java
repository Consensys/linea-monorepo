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

import static net.consensys.linea.zktracer.types.Conversions.longToBytes32;

import java.nio.MappedByteBuffer;
import java.util.List;

import com.google.common.base.Preconditions;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.fragment.imc.StpCall;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes32;

@RequiredArgsConstructor
public class Stp implements Module {

  private final StackedSet<StpOperation> operations = new StackedSet<>();

  private final Wcp wcp;
  private final Mod mod;

  public void call(StpCall stpCall) {
    final StpOperation stpOperation = new StpOperation(stpCall);
    this.operations.add(stpOperation);

    Preconditions.checkArgument(
        stpCall.opCode().isCall() || stpCall.opCode().isCreate(),
        "STP handles only Calls and CREATEs");

    if (stpCall.opCode().isCreate()) {
      this.wcp.callLT(longToBytes32(stpCall.gasActual()), Bytes32.ZERO);
      this.wcp.callLT(longToBytes32(stpCall.gasActual()), longToBytes32(stpCall.upfrontGasCost()));
      if (!stpCall.outOfGasException()) {
        this.mod.callDIV(longToBytes32(stpOperation.getGDiff()), longToBytes32(64L));
      }
    }

    if (stpCall.opCode().isCall()) {
      this.wcp.callLT(longToBytes32(stpCall.gasActual()), Bytes32.ZERO);
      if (stpCall.opCode().callCanTransferValue()) {
        this.wcp.callISZERO(Bytes32.leftPad(stpCall.value()));
      }
      this.wcp.callLT(longToBytes32(stpCall.gasActual()), longToBytes32(stpCall.upfrontGasCost()));
      if (!stpCall.outOfGasException()) {
        this.mod.callDIV(longToBytes32(stpOperation.getGDiff()), longToBytes32(64L));
        this.wcp.callLT(stpCall.gas(), longToBytes32(stpOperation.get63of64GDiff()));
      }
    }
  }

  @Override
  public String moduleKey() {
    return "STP";
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
  public int lineCount() {
    return this.operations.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;
    for (StpOperation chunk : operations) {
      stamp++;
      chunk.trace(trace, stamp);
    }
  }
}
