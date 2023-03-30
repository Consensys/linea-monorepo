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
package net.consensys.zktracer.module.wcp;

import static net.consensys.zktracer.OpCode.SGT;
import static org.assertj.core.api.AssertionsForClassTypes.assertThat;
import static org.mockito.Mockito.when;

import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.zktracer.CorsetValidator;
import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.ZkTraceBuilder;
import net.consensys.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class WcpTracerTest {
  private ZkTracer zkTracer;
  private ZkTraceBuilder zkTraceBuilder;
  @Mock MessageFrame mockFrame;
  @Mock Operation mockOperation;

  private static final Random rand = new Random();
  private static final int TEST_REPETITIONS = 4;

  @BeforeEach
  void setUp() {
    zkTraceBuilder = new ZkTraceBuilder();
    zkTracer = new ZkTracer(zkTraceBuilder, List.of(new WcpTracer()));
    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }

  @ParameterizedTest()
  @MethodSource("provideNonRandomAddArguments")
  void testRandomWcp(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    when(mockOperation.getOpcode()).thenReturn((int) opCode.value);
    when(mockFrame.getStackItem(0)).thenReturn(arg1);
    when(mockFrame.getStackItem(1)).thenReturn(arg2);
    zkTracer.tracePreExecution(mockFrame);
    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @Test
  public void testNonRandomWcp() {
    Bytes32 arg1 =
        Bytes32.fromHexString("0xdcd5cf52e4daec5389587d0d0e996e6ce2d0546b63d3ea0a0dc48ad984d180a9");
    Bytes32 arg2 =
        Bytes32.fromHexString("0x0479484af4a59464a48818b3980174687661bafb13d06f49537995fa6c02159e");
    traceOperation(SGT, arg1, arg2);
  }

  public static Stream<Arguments> provideNonRandomArguments() {
    final List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : new WcpTracer().supportedOpCodes()) {
      for (int i = 0; i <= TEST_REPETITIONS; i++) {
        Bytes32[] payload = new Bytes32[2];
        payload[0] = Bytes32.random(rand);
        payload[1] = Bytes32.random(rand);
        arguments.add(Arguments.of(opCode, payload[0], payload[1]));
      }
    }
    return arguments.stream();
  }

  private void traceOperation(OpCode opcode, Bytes32 arg1, Bytes32 arg2) {
    when(mockOperation.getOpcode()).thenReturn((int) opcode.value);
    when(mockFrame.getStackItem(0)).thenReturn(arg1);
    when(mockFrame.getStackItem(1)).thenReturn(arg2);
    zkTracer.tracePreExecution(mockFrame);
    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }
}
