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
package net.consensys.linea.zktracer.module.alu.mod;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.module.AbstractModuleTracerTest;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class ModTracerTest extends AbstractModuleTracerTest {
  static final Random rand = new Random();

  @ParameterizedTest()
  @MethodSource("provideNonRandomArguments")
  void testNonRandomArguments(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, arg1, arg2);
  }

  @Test
  void testDivModTest() {
    var t = getRandomAluModInstruction(rand.nextInt(8 + 1), rand.nextInt(8 + 1));
    t.forEach(
        test -> runTest((OpCode) test.get()[0], (Bytes32) test.get()[1], (Bytes32) test.get()[2]));
  }

  @Test
  protected void testExactlyValue() {
    runTest(
        OpCode.SDIV,
        Bytes32.fromHexString("0x9e292fc6050e85e755a609ce0e1cc672ca9ceec4ce68faca2ac61101881aa954"),
        Bytes32.fromHexString(
            "0x111e6c0771f901efae83da54a41a10728c9865398aa12bf16eb3bb9e9a1f8cd0"));
  }

  @Override
  protected Stream<Arguments> provideNonRandomArguments() {
    List<Arguments> argumentsList = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int k = 1; k <= 4; k++) {
        for (int i = 1; i <= 4; i++) {
          argumentsList.add(Arguments.of(opCode, UInt256.valueOf(i), UInt256.valueOf(k)));
        }
      }
    }
    return argumentsList.stream();
  }

  private Stream<Arguments> getRandomAluModInstruction(int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    int remainder = rand.nextInt(4) % 4;
    return switch (remainder) {
      case 0 -> Stream.of(Arguments.of(OpCode.DIV, bytes1, bytes2));
      case 1 -> Stream.of(Arguments.of(OpCode.SDIV, bytes1, bytes2));
      case 2 -> Stream.of(Arguments.of(OpCode.MOD, bytes1, bytes2));
      default -> Stream.of(Arguments.of(OpCode.SMOD, bytes1, bytes2));
    };
  }

  @Override
  protected ModuleTracer getModuleTracer() {
    return new ModTracer();
  }
}
