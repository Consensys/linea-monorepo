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

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.AbstractModuleCorsetTest;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class ModTracerTest extends AbstractModuleCorsetTest {
  private static final Random rand = new Random();

  private static final int TEST_MOD_REPETITIONS = 16;

  @ParameterizedTest()
  @MethodSource("provideRandomAluModArguments")
  void aluModTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("provideRandomDivisionsByZeroArguments")
  void aluModRandomDivisionsByZeroTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("provideDivisibleArguments")
  void aluModDivisibleTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("provideNegativeDivisibleArguments")
  void aluModNegativeDivisibleTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  private Stream<Arguments> provideRandomAluModArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      arguments.add(getRandomAluModInstruction(rand.nextInt(32) + 1, rand.nextInt(32) + 1));
    }
    return arguments.stream();
  }

  @Override
  public Stream<Arguments> provideNonRandomArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int k = 1; k <= 4; k++) {
        for (int i = 1; i <= 4; i++) {
          arguments.add(Arguments.of(opCode, List.of(UInt256.valueOf(i), UInt256.valueOf(k))));
        }
      }
    }
    return arguments.stream();
  }

  private Stream<Arguments> provideDivisibleArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      long b = rand.nextInt(Integer.MAX_VALUE);
      long q = rand.nextInt(Integer.MAX_VALUE);
      long a = b * q;
      OpCode opCode = getRandomSupportedOpcode();
      arguments.add(Arguments.of(opCode, UInt256.valueOf(a), UInt256.valueOf(b)));
    }
    return arguments.stream();
  }

  private Stream<Arguments> provideNegativeDivisibleArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      long b = rand.nextInt(Integer.MAX_VALUE) + 1L;
      long q = rand.nextInt(Integer.MAX_VALUE) + 1L;
      long a = b * q;
      OpCode opCode = getRandomSupportedOpcode();
      arguments.add(
          Arguments.of(opCode, convertToComplementBytes32(a), convertToComplementBytes32(b)));
    }
    return arguments.stream();
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

  private Stream<Arguments> provideRandomDivisionsByZeroArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (int i = 0; i < TEST_MOD_REPETITIONS; i++) {
      int a = rand.nextInt(256) + 1;
      arguments.add(getRandomAluModInstruction(a, 0));
    }
    return arguments.stream();
  }

  private Arguments getRandomAluModInstruction(int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    OpCode opCode = getRandomSupportedOpcode();
    return Arguments.of(opCode, bytes1, bytes2);
  }

  @Override
  protected Module getModuleTracer() {
    return new Mod();
  }
}
