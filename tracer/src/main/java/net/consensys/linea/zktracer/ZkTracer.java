/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer;

import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.tracer.BlockAwareOperationTracer;

@RequiredArgsConstructor
public class ZkTracer implements BlockAwareOperationTracer {
  private final ZkTraceBuilder zkTraceBuilder = new ZkTraceBuilder();
  private final List<Module> modules;

  public ZkTracer() {
    this(List.of(new Hub(), new Mul(), new Shf(), new Wcp(), new Add(), new Mod()));
  }

  public ZkTrace getTrace() {
    for (Module module : this.modules) {
      zkTraceBuilder.addTrace(module);
    }
    return zkTraceBuilder.build();
  }

  public void traceStartConflation(final long numBlocksInConflation) {
    for (Module module : this.modules) {
      module.traceStartConflation(numBlocksInConflation);
    }
  }

  public void traceEndConflation() {
    for (Module module : this.modules) {
      module.traceEndConflation();
    }
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    for (Module module : this.modules) {
      module.traceStartBlock(blockHeader, blockBody);
    }
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    for (Module module : this.modules) {
      module.traceEndBlock(blockHeader, blockBody);
    }
  }

  @Override
  public void traceStartTransaction(final Transaction transaction) {
    for (Module module : this.modules) {
      module.traceStartTx(transaction);
    }
  }

  @Override
  public void traceEndTransaction(final Bytes output, final long gasUsed, final long timeNs) {
    for (Module module : this.modules) {
      module.traceEndTx();
    }
  }

  // TODO: missing ContextEnter/Exit

  @Override
  public void tracePreExecution(final MessageFrame frame) {
    OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    for (Module module : this.modules) {
      if (module.supportedOpCodes().contains(opCode)) {
        module.trace(frame);
      }
    }
  }
}
