/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.module.mod;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.DynamicTests;
import net.consensys.linea.zktracer.testing.OpcodeCall;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;

class ModTracerTest {
  private static final Random RAND = new Random();
  private static final int TEST_MOD_REPETITIONS = 16;
  private static final Module MODULE = new Mod();
  private static final DynamicTests DYN_TESTS = DynamicTests.forModule(MODULE);

  @TestFactory
  Stream<DynamicTest> runDynamicTests() {
    return DYN_TESTS
        .testCase("non random arguments test", provideNonRandomArguments())
        .testCase("random alu mod arguments test", provideRandomAluModArguments())
        .testCase(
            "random divisions by zero arguments test", provideRandomDivisionsByZeroArguments())
        .testCase("random divisible arguments test", provideRandomDivisibleArguments())
        .testCase(
            "random negative divisible arguments test", provideRandomNegativeDivisibleArguments())
        .run();
  }

  private List<OpcodeCall> provideRandomAluModArguments() {
    return DYN_TESTS.newModuleArgumentsProvider(
        (arguments, opCode) -> {
          for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
            addRandomAluModInstruction(arguments, RAND.nextInt(32) + 1, RAND.nextInt(32) + 1);
          }
        });
  }

  private List<OpcodeCall> provideNonRandomArguments() {
    return DYN_TESTS.newModuleArgumentsProvider(
        (testCases, opCode) -> {
          for (int k = 1; k <= 4; k++) {
            for (int i = 1; i <= 4; i++) {
              testCases.add(
                  new OpcodeCall(opCode, List.of(UInt256.valueOf(i), UInt256.valueOf(k))));
            }
          }
        });
  }

  private List<OpcodeCall> provideRandomDivisibleArguments() {
    List<OpcodeCall> testCases = new ArrayList<>();

    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      long b = RAND.nextInt(Integer.MAX_VALUE);
      long q = RAND.nextInt(Integer.MAX_VALUE);
      long a = b * q;

      OpCode opCode = DYN_TESTS.getRandomSupportedOpcode();
      testCases.add(new OpcodeCall(opCode, List.of(UInt256.valueOf(a), UInt256.valueOf(b))));
    }

    return testCases;
  }

  private List<OpcodeCall> provideRandomNegativeDivisibleArguments() {
    List<OpcodeCall> testCases = new ArrayList<>();

    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      long b = RAND.nextInt(Integer.MAX_VALUE) + 1L;
      long q = RAND.nextInt(Integer.MAX_VALUE) + 1L;
      long a = b * q;

      OpCode opCode = DYN_TESTS.getRandomSupportedOpcode();

      testCases.add(
          new OpcodeCall(
              opCode, List.of(convertToComplementBytes32(a), convertToComplementBytes32(b))));
    }

    return testCases;
  }

  private void addRandomAluModInstruction(
      List<OpcodeCall> testCases, int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    OpCode opCode = DYN_TESTS.getRandomSupportedOpcode();

    testCases.add(new OpcodeCall(opCode, List.of(bytes1, bytes2)));
  }

  private Bytes32 convertToComplementBytes32(long number) {
    BigInteger bigInteger = BigInteger.valueOf(number);
    if (Math.random() >= 0.5) {
      bigInteger = bigInteger.negate();
    }

    Bytes resultBytes = Bytes.wrap(bigInteger.toByteArray());
    if (resultBytes.size() > 32) {
      resultBytes = resultBytes.slice(resultBytes.size() - 32, 32);
    }

    return Bytes32.leftPad(resultBytes, bigInteger.signum() < 0 ? (byte) 0xFF : 0x00);
  }

  private List<OpcodeCall> provideRandomDivisionsByZeroArguments() {
    List<OpcodeCall> testCases = new ArrayList<>();

    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      int a = RAND.nextInt(256) + 1;
      addRandomAluModInstruction(testCases, a, 0);
    }

    return testCases;
  }
}
