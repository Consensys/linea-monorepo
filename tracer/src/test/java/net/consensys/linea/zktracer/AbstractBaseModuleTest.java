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

import java.util.List;

import net.consensys.linea.zktracer.corset.CorsetValidator;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestInstance.Lifecycle;

/**
 * Base test class used to set up mocking of a {@link MessageFrame}, {@link OpCode} and trace
 * generation functionality.
 */
@TestInstance(Lifecycle.PER_CLASS)
public abstract class AbstractBaseModuleTest {
  private ZkTracer zkTracer;
  MessageFrame mockFrame;
  Operation mockOperation;
  static Module module;

  @BeforeAll
  static void beforeAll() {
    OpCodes.load();
  }

  @BeforeEach
  void beforeEach() {
    module = getModuleTracer();
    zkTracer = new ZkTracer(List.of(module));
    mockFrame = mock(MessageFrame.class);
    mockOperation = mock(Operation.class);
    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }

  protected void runTest(final OpCodeData opCodeData, final List<Bytes32> arguments) {
    assertThat(CorsetValidator.isValid(generateTrace(opCodeData, arguments))).isTrue();
  }

  protected String generateTrace(OpCodeData opCodeData, List<Bytes32> arguments) {
    when(mockOperation.getOpcode()).thenReturn(opCodeData.value().intValue());

    for (int i = 0; i < arguments.size(); i++) {
      when(mockFrame.getStackItem(i)).thenReturn(arguments.get(i));
    }

    zkTracer.traceStartConflation(1);
    zkTracer.tracePreExecution(mockFrame);
    zkTracer.traceEndConflation();

    return zkTracer.getTrace().toJson();
  }

  protected abstract Module getModuleTracer();
}
