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

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.consensys.linea.zktracer.module.ModuleTracer;
import net.consensys.linea.zktracer.module.alu.add.AddTracer;
import net.consensys.linea.zktracer.module.alu.mod.ModTracer;
import net.consensys.linea.zktracer.module.alu.mul.MulTracer;
import net.consensys.linea.zktracer.module.hub.HubTracer;
import net.consensys.linea.zktracer.module.shf.ShfTracer;
import net.consensys.linea.zktracer.module.wcp.WcpTracer;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.plugin.services.tracer.BlockAwareOperationTracer;

public class ZkTracer implements BlockAwareOperationTracer {
  private final List<ModuleTracer> tracers;
  private final Map<OpCode, List<ModuleTracer>> opCodeTracerMap = new HashMap<>();

  private final ZkTraceBuilder zkTraceBuilder;

  public ZkTracer(final ZkTraceBuilder zkTraceBuilder, final List<ModuleTracer> tracers) {
    this.tracers = tracers;
    this.zkTraceBuilder = zkTraceBuilder;
    setupTracers();
  }

  public ZkTracer(final ZkTraceBuilder zkTraceBuilder) {
    this(
        zkTraceBuilder,
        List.of(
            new HubTracer(),
            new MulTracer(),
            new ShfTracer(),
            new WcpTracer(),
            new AddTracer(),
            new ModTracer()));
  }

  @Override
  public void tracePreExecution(final MessageFrame frame) {
    for (ModuleTracer tracer :
        opCodeTracerMap.get(OpCode.of(frame.getCurrentOperation().getOpcode()))) {
      if (tracer != null) {
        zkTraceBuilder.addTrace(tracer.jsonKey(), tracer.trace(frame));
      }
    }
  }

  private void setupTracers() {
    for (ModuleTracer tracer : tracers) {
      for (OpCode opCode : tracer.supportedOpCodes()) {
        List<ModuleTracer> moduleTracers = opCodeTracerMap.get(opCode);
        if (moduleTracers == null) {
          moduleTracers = List.of(tracer);
        } else {
          moduleTracers.add(tracer);
        }

        opCodeTracerMap.put(opCode, moduleTracers);
      }
    }
  }
}
