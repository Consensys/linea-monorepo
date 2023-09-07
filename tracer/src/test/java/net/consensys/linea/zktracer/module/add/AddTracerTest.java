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

package net.consensys.linea.zktracer.module.add;

import java.net.URL;
import java.util.Random;
import java.util.stream.Stream;

import com.google.common.collect.ArrayListMultimap;
import com.google.common.collect.Multimap;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.DynamicTests;
import net.consensys.linea.zktracer.testing.SpecTests;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

class AddTracerTest {
  private static final Random RAND = new Random();
  private static final int TEST_ADD_REPETITIONS = 16;
  private static final Module MODULE = new Add();
  private static final DynamicTests DYN_TESTS = DynamicTests.forModule(MODULE);

  @TestFactory
  Stream<DynamicTest> runDynamicTests() {
    return DYN_TESTS
        .testCase("non random arguments test", provideNonRandomArguments())
        .testCase("simple alu add arguments test", provideSimpleAluAddArguments())
        .testCase("random alu add arguments test", provideRandomAluAddArguments())
        .run();
  }

  @ParameterizedTest(name = "{index} {0}")
  @MethodSource("provideTraceSpecs")
  void traceWithSpecFile(final String ignored, URL specUrl) {
    SpecTests.runSpecTestWithTraceComparison(specUrl, MODULE.jsonKey());
  }

  private static Object[][] provideTraceSpecs() {
    return SpecTests.findSpecFiles(MODULE.jsonKey());
  }

  private Multimap<OpCode, Bytes32> provideNonRandomArguments() {
    return DYN_TESTS.newModuleArgumentsProvider(
        (arguments, opCode) -> {
          for (int k = 1; k <= 4; k++) {
            for (int i = 1; i <= 4; i++) {
              arguments.put(opCode, UInt256.valueOf(i));
              arguments.put(opCode, UInt256.valueOf(k));
            }
          }
        });
  }

  public Multimap<OpCode, Bytes32> provideRandomAluAddArguments() {
    return DYN_TESTS.newModuleArgumentsProvider(
        (arguments, opCode) -> {
          for (int i = 0; i < TEST_ADD_REPETITIONS; i++) {
            addRandomAluAddInstruction(arguments, RAND.nextInt(32) + 1, RAND.nextInt(32) + 1);
          }
        });
  }

  private Multimap<OpCode, Bytes32> provideSimpleAluAddArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    Bytes32 bytes1 = Bytes32.rightPad(Bytes.fromHexString("0x80"));
    Bytes32 bytes2 = Bytes32.leftPad(Bytes.fromHexString("0x01"));

    arguments.put(OpCode.SUB, bytes1);
    arguments.put(OpCode.SUB, bytes2);

    return arguments;
  }

  private void addRandomAluAddInstruction(
      Multimap<OpCode, Bytes32> arguments, int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    OpCode opCode = DYN_TESTS.getRandomSupportedOpcode();

    arguments.put(opCode, bytes1);
    arguments.put(opCode, bytes2);
  }
}
