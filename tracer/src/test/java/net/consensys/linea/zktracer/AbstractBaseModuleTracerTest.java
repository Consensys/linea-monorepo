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

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

import java.util.List;

import net.consensys.linea.zktracer.corset.CorsetValidator;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestInstance.Lifecycle;

@TestInstance(Lifecycle.PER_CLASS)
public abstract class AbstractBaseModuleTracerTest {
  private ZkTracer zkTracer;
  private ZkTraceBuilder zkTraceBuilder;
  MessageFrame mockFrame;
  Operation mockOperation;
  static ModuleTracer moduleTracer;

  protected void runTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    String trace = generateTrace(opCode, List.of(arg1, arg2));
    assertThat(CorsetValidator.isValid(trace)).isTrue();
  }

  protected String generateTrace(OpCode opCode, List<Bytes32> arguments) {
    when(mockOperation.getOpcode()).thenReturn((int) opCode.value);
    for (int i = 0; i < arguments.size(); i++) {
      when(mockFrame.getStackItem(i)).thenReturn(arguments.get(i));
    }
    zkTracer.tracePreExecution(mockFrame);
    return zkTraceBuilder.build().toJson();
  }

  @BeforeEach
  void setUp() {
    zkTraceBuilder = new ZkTraceBuilder();
    moduleTracer = getModuleTracer();
    zkTracer = new ZkTracer(zkTraceBuilder, List.of(moduleTracer));
    mockFrame = mock(MessageFrame.class);
    mockOperation = mock(Operation.class);
    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }

  protected abstract ModuleTracer getModuleTracer();
}
