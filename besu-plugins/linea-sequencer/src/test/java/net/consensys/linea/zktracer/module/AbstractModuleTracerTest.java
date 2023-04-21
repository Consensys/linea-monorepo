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
package net.consensys.linea.zktracer.module;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.CorsetValidator;
import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.ZkTraceBuilder;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestInstance.Lifecycle;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.Mock;

@TestInstance(Lifecycle.PER_CLASS)
public abstract class AbstractModuleTracerTest {
  static final Random rand = new Random();
  private static final int TEST_REPETITIONS = 8;

  private ZkTracer zkTracer;
  private ZkTraceBuilder zkTraceBuilder;
  @Mock MessageFrame mockFrame;
  @Mock Operation mockOperation;
  static ModuleTracer moduleTracer;

  @ParameterizedTest()
  @MethodSource("provideRandomArguments")
  void randomArgumentsTest(OpCode opCode, List<Bytes32> args) {
    runTest(opCode, args);
  }

  @ParameterizedTest()
  @MethodSource("provideNonRandomArguments")
  void nonRandomArgumentsTest(OpCode opCode, List<Bytes32> arguments) {
    runTest(opCode, arguments);
  }

  protected void runTest(OpCode opCode, List<Bytes32> arguments) {
    when(mockOperation.getOpcode()).thenReturn((int) opCode.value);
    for (int i = 0; i < arguments.size(); i++) {
      when(mockFrame.getStackItem(i)).thenReturn(arguments.get(i));
    }
    zkTracer.tracePreExecution(mockFrame);
    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @BeforeEach
  void setUp() {
    zkTraceBuilder = new ZkTraceBuilder();
    moduleTracer = getModuleTracer();
    zkTracer = new ZkTracer(zkTraceBuilder, List.of(moduleTracer));
    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }

  protected abstract Stream<Arguments> provideNonRandomArguments();

  public Stream<Arguments> provideRandomArguments() {
    final List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int i = 0; i <= TEST_REPETITIONS; i++) {
        arguments.add(Arguments.of(opCode, List.of(Bytes32.random(rand), Bytes32.random(rand))));
      }
    }
    return arguments.stream();
  }

  protected abstract ModuleTracer getModuleTracer();

  protected OpCode getRandomSupportedOpcode() {
    var supportedOpCodes = getModuleTracer().supportedOpCodes();
    int index = rand.nextInt(supportedOpCodes.size());
    return supportedOpCodes.get(index);
  }
}
