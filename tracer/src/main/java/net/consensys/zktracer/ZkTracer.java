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
package net.consensys.zktracer;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.ZkTraceBuilder;
import net.consensys.zktracer.module.ModuleTracer;
import net.consensys.zktracer.module.shf.ShfTracer;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.tracing.OperationTracer;

public class ZkTracer implements OperationTracer {
  private final List<ModuleTracer> tracers = List.of(new ShfTracer());
  private final Map<OpCode, List<ModuleTracer>> opCodeTracerMap = new HashMap<>();

  private final ZkTraceBuilder zkTraceBuilder;

  public ZkTracer(final ZkTraceBuilder zkTraceBuilder) {
    this.zkTraceBuilder = zkTraceBuilder;

    setupTracers();
  }

  @Override
  public void tracePreExecution(final MessageFrame frame) {

    opCodeTracerMap
        .get(OpCode.of(frame.getCurrentOperation().getOpcode()))
        .forEach(
            tracer -> {
              if (tracer != null) {
                zkTraceBuilder.addTrace(tracer.jsonKey(), tracer.trace(frame));
              }
            });
  }

  private void setupTracers() {
    tracers.forEach(
        tracer ->
            tracer
                .supportedOpCodes()
                .forEach(
                    opCode -> {
                      if (opCodeTracerMap.containsKey(opCode)) {
                        throw new AssertionError(
                            "OpCode "
                                + opCode.name()
                                + " supported by more than one Tracer: "
                                + opCodeTracerMap.get(opCode).getClass().getSimpleName()
                                + " ,"
                                + tracer.getClass().getSimpleName());
                      }

                      List<ModuleTracer> moduleTracers = opCodeTracerMap.get(opCode);

                      if (moduleTracers == null) {
                        moduleTracers = List.of(tracer);
                      } else {
                        moduleTracers.add(tracer);
                      }

                      opCodeTracerMap.put(opCode, moduleTracers);
                    }));
  }
}
