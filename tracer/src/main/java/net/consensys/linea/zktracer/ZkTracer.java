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
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.gascalculator.LondonGasCalculator;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@RequiredArgsConstructor
public class ZkTracer implements ZkBlockAwareOperationTracer {
  /** The {@link GasCalculator} used in this version of the arithmetization */
  public static final GasCalculator gasCalculator = new LondonGasCalculator();

  private final ZkTraceBuilder zkTraceBuilder = new ZkTraceBuilder();
  private final Hub hub;

  private final List<Module> modules;

  public ZkTracer() {
    this.hub = new Hub();
    this.modules = hub.getSelfStandingModules();

    // Load opcodes configured in src/main/resources/opcodes.yml.
    OpCodes.load();
  }

  public ZkTrace getTrace() {
    zkTraceBuilder.addTrace(this.hub);
    for (Module module : this.modules) {
      zkTraceBuilder.addTrace(module);
    }

    return zkTraceBuilder.build();
  }

  @Override
  public void traceStartConflation(final long numBlocksInConflation) {
    hub.traceStartConflation(numBlocksInConflation);
    for (Module module : this.modules) {
      module.traceStartConflation(numBlocksInConflation);
    }
  }

  @Override
  public void traceEndConflation() {
    for (Module module : this.modules) {
      module.traceEndConflation();
    }
  }

  @Override
  public String getJsonTrace() {
    return getTrace().toJson();
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.hub.traceStartBlock(blockHeader, blockBody);
    for (Module module : this.modules) {
      module.traceStartBlock(blockHeader, blockBody);
    }
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.hub.traceEndBlock(blockHeader, blockBody);
    for (Module module : this.modules) {
      module.traceEndBlock(blockHeader, blockBody);
    }
  }

  @Override
  public void traceStartTransaction(WorldView worldView, Transaction transaction) {
    this.hub.traceStartTx(worldView, transaction);
    for (Module module : this.modules) {
      module.traceStartTx(worldView, transaction);
    }
  }

  @Override
  public void traceEndTransaction(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      long timeNs) {
    this.hub.traceEndTx(worldView, tx, status, output, logs, gasUsed);
    for (Module module : this.modules) {
      module.traceEndTx(worldView, tx, status, output, logs, gasUsed);
    }
  }

  @Override
  public void tracePreExecution(final MessageFrame frame) {
    this.hub.trace(frame);
  }

  @Override
  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    this.hub.tracePostExecution(frame, operationResult);
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    this.hub.traceContextEnter(frame);
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    this.hub.traceContextExit(frame);
  }
}
