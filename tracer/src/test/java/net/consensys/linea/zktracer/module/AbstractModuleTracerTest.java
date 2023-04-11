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

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.CorsetValidator;
import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.ZkTraceBuilder;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
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
  private static final int TEST_REPETITIONS = 4;

  private ZkTracer zkTracer;
  private ZkTraceBuilder zkTraceBuilder;
  @Mock
  MessageFrame mockFrame;
  @Mock
  Operation mockOperation;
  static ModuleTracer moduleTracer;

  @ParameterizedTest()
  @MethodSource("provideRandomArguments")
  void testRandomArguments(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, arg1, arg2);
  }

  @ParameterizedTest()
  @MethodSource("provideNonRandomArguments")
  void testNonRandomArguments(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, arg1, arg2);
  }

  protected void runTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    when(mockOperation.getOpcode()).thenReturn((int) opCode.value);
    when(mockFrame.getStackItem(0)).thenReturn(arg1);
    when(mockFrame.getStackItem(1)).thenReturn(arg2);
    zkTracer.tracePreExecution(mockFrame);
    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @BeforeEach
  void setUp() {
    zkTraceBuilder = new ZkTraceBuilder();
    moduleTracer =  getModuleTracer();
    zkTracer = new ZkTracer(zkTraceBuilder, List.of(moduleTracer));
    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }


   protected abstract Stream<Arguments> provideNonRandomArguments();

  public Stream<Arguments> provideRandomArguments() {
    final List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int i = 0; i <= getTestRepetitionsCount(); i++) {
        Bytes32[] payload = new Bytes32[2];
        payload[0] = getFirstArgument();
        payload[1] = getSecondArgument();
        arguments.add(Arguments.of(opCode, payload[0], payload[1]));
      }
    }
    return arguments.stream();
  }

  private static Bytes32 getFirstArgument(){
    return Bytes32.random(rand);
  }

  private static Bytes32 getSecondArgument(){
    return Bytes32.random(rand);
  }

  private static int getTestRepetitionsCount(){
      return TEST_REPETITIONS;
  }

  protected abstract ModuleTracer getModuleTracer() ;
}
