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
package net.consensys.zktracer.module.alu.mul;

import net.consensys.zktracer.CorsetValidator;
import net.consensys.zktracer.OpCode;
import net.consensys.zktracer.ZkTraceBuilder;
import net.consensys.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Named;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.math.BigInteger;
import java.util.Random;
import java.util.stream.Stream;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class MulTracerTest {
  private static final Logger LOG = LoggerFactory.getLogger(MulTracerTest.class);

  private static final Random rand = new Random();
  private static final int TEST_REPETITIONS = 4;

  private ZkTracer zkTracer;
  private ZkTraceBuilder zkTraceBuilder;

  @Mock MessageFrame mockFrame;
  @Mock Operation mockOperation;

  @BeforeEach
  void setUp() {
    zkTraceBuilder = new ZkTraceBuilder();
    zkTracer = new ZkTracer(zkTraceBuilder);

    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("provideMulOperators")
  void testFailingBlockchainBlock(final int opCodeValue) {
    when(mockOperation.getOpcode()).thenReturn(opCodeValue);

    when(mockFrame.getStackItem(0)).thenReturn(Bytes32.rightPad(Bytes.fromHexString("0x08")));
    when(mockFrame.getStackItem(1)).thenReturn(Bytes32.fromHexString("0x0128"));

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("provideRandomArguments")
  void testRandomExp(final Bytes32[] payload) {
    LOG.info(
            "arg1: " + payload[0].toShortHexString() + ", arg2: " + payload[1].toShortHexString());
    when(mockOperation.getOpcode()).thenReturn((int) OpCode.EXP.value);

    when(mockFrame.getStackItem(0)).thenReturn(payload[0]);
    when(mockFrame.getStackItem(1)).thenReturn(payload[1]);

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("provideNonRandomArguments")
  void testNonRandomMul(final Bytes32[] payload) {
    LOG.info(
            "arg1: " + payload[0].toShortHexString() + ", arg2: " + payload[1].toShortHexString());
    when(mockOperation.getOpcode()).thenReturn((int) OpCode.EXP.value);

    when(mockFrame.getStackItem(0)).thenReturn(payload[0]);
    when(mockFrame.getStackItem(1)).thenReturn(payload[1]);

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @Test
  void testTmp() {
    when(mockOperation.getOpcode()).thenReturn((int) OpCode.MUL.value);

    when(mockFrame.getStackItem(0))
        .thenReturn(Bytes32.fromHexStringLenient("0x54fda4f3c1452c8c58df4fb1e9d6de"));
    when(mockFrame.getStackItem(1)).thenReturn(Bytes32.fromHexStringLenient("0xb5"));

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }


  public static Stream<Arguments> provideNonRandomArguments() {
    final Arguments[] arguments = new Arguments[TEST_REPETITIONS];

    for (int i = 0; i < TEST_REPETITIONS; i++) {
      Bytes32[] payload = new Bytes32[2];
      payload[0] = Bytes32.leftPad(Bytes.of(1+i));
      payload[1] = Bytes32.leftPad(Bytes.of(i));
      arguments[i] =
              Arguments.of(
                      Named.of(
                              "arg1: "
                                      + payload[0]
                                      + ", arg2: "
                                      + payload[1],
                              payload));
    }

    return Stream.of(arguments);
  }

  public static Stream<Arguments> provideRandomArguments() {
      final Arguments[] arguments = new Arguments[TEST_REPETITIONS];

      for (int i = 0; i < TEST_REPETITIONS; i++) {

        final byte[] randomBytes1 = new byte[32];
        rand.nextBytes(randomBytes1);
        final byte[] randomBytes2 = new byte[32];
        rand.nextBytes(randomBytes2);

        Bytes32[] payload = new Bytes32[2];
        payload[0] = Bytes32.wrap(randomBytes1);
        payload[1] = Bytes32.wrap(randomBytes2);

        arguments[i] =
                Arguments.of(
                        Named.of(
                                "arg1: "
                                        + payload[0].toHexString()
                                        + ", arg2: "
                                        + payload[1].toHexString(),
                                payload));
      }

      return Stream.of(arguments);
    }
  public static Stream<Arguments> provideMulOperators() {
    return Stream.of(
        Arguments.of(Named.of("MUL", (int) OpCode.MUL.value)),
        Arguments.of(Named.of("EXP", (int) OpCode.EXP.value)));
  }

}
