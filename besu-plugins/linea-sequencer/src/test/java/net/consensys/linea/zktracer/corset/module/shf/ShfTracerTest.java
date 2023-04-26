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
package net.consensys.linea.zktracer.corset.module.shf;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;
import static org.mockito.Mockito.when;

import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;

import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.ZkTraceBuilder;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.corset.CorsetValidator;
import net.consensys.linea.zktracer.module.shf.ShfTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
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

@ExtendWith(MockitoExtension.class)
class ShfTracerTest {
  private static final Logger LOG = LoggerFactory.getLogger(ShfTracerTest.class);

  private static final Random rand = new Random();
  private static final int TEST_REPETITIONS = 4;

  private ZkTracer zkTracer;
  private ZkTraceBuilder zkTraceBuilder;

  @Mock MessageFrame mockFrame;
  @Mock Operation mockOperation;

  @BeforeEach
  void setUp() {
    zkTraceBuilder = new ZkTraceBuilder();
    zkTracer = new ZkTracer(zkTraceBuilder, List.of(new ShfTracer()));

    when(mockFrame.getCurrentOperation()).thenReturn(mockOperation);
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("provideShiftOperators")
  void testFailingBlockchainBlock(final int opCodeValue) {
    when(mockOperation.getOpcode()).thenReturn(opCodeValue);

    when(mockFrame.getStackItem(0)).thenReturn(Bytes32.rightPad(Bytes.fromHexString("0x08")));
    when(mockFrame.getStackItem(1)).thenReturn(Bytes32.fromHexString("0x01"));

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("provideRandomSarArguments")
  void testRandomSar(final Bytes32[] payload) {
    LOG.info(
        "value: " + payload[0].toShortHexString() + ", shift by: " + payload[1].toShortHexString());
    when(mockOperation.getOpcode()).thenReturn((int) OpCode.SAR.value);

    when(mockFrame.getStackItem(0)).thenReturn(payload[0]);
    when(mockFrame.getStackItem(1)).thenReturn(payload[1]);

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  @Test
  void testTmp() {
    when(mockOperation.getOpcode()).thenReturn((int) OpCode.SAR.value);

    when(mockFrame.getStackItem(0))
        .thenReturn(Bytes32.fromHexStringLenient("0x54fda4f3c1452c8c58df4fb1e9d6de"));
    when(mockFrame.getStackItem(1)).thenReturn(Bytes32.fromHexStringLenient("0xb5"));

    zkTracer.tracePreExecution(mockFrame);

    assertThat(CorsetValidator.isValid(zkTraceBuilder.build().toJson())).isTrue();
  }

  public static Stream<Arguments> provideRandomSarArguments() {
    final Arguments[] arguments = new Arguments[TEST_REPETITIONS];

    for (int i = 0; i < TEST_REPETITIONS; i++) {
      final boolean signBit = rand.nextInt(2) == 1;

      // leave the first byte untouched
      final int k = 1 + rand.nextInt(31);

      final byte[] randomBytes = new byte[k];
      rand.nextBytes(randomBytes);

      final byte[] signBytes = new byte[32 - k];
      if (signBit) {
        signBytes[0] = (byte) 0x80; // 0b1000_0000, i.e. sign bit == 1
      }

      final byte[] bytes = concatenateArrays(signBytes, randomBytes);
      byte shiftBy = (byte) rand.nextInt(256);

      Bytes32[] payload = new Bytes32[2];
      payload[0] = Bytes32.wrap(bytes);
      payload[1] = Bytes32.leftPad(Bytes.of(shiftBy));

      arguments[i] =
          Arguments.of(
              Named.of(
                  "value: "
                      + payload[0].toHexString()
                      + ", shiftBy: "
                      + payload[1].toShortHexString(),
                  payload));
    }

    return Stream.of(arguments);
  }

  public static Stream<Arguments> provideShiftOperators() {
    return Stream.of(
        Arguments.of(Named.of("SAR", (int) OpCode.SAR.value)),
        Arguments.of(Named.of("SHL", (int) OpCode.SHL.value)),
        Arguments.of(Named.of("SHR", (int) OpCode.SHR.value)));
  }

  private static byte[] concatenateArrays(byte[] a, byte[] b) {
    int length = a.length + b.length;
    byte[] result = new byte[length];
    System.arraycopy(a, 0, result, 0, a.length);
    System.arraycopy(b, 0, result, a.length, b.length);
    return result;
  }
}
