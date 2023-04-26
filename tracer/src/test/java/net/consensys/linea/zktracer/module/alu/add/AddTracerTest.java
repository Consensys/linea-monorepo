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
package net.consensys.linea.zktracer.module.alu.add;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.AbstractModuleTracerCorsetTest;
import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.module.ModuleTracer;
import net.consensys.linea.zktracer.module.alu.add.AddTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class AddTracerTest extends AbstractModuleTracerCorsetTest {
  private static final Random rand = new Random();

  private static final int TEST_ADD_REPETITIONS = 16;

  @ParameterizedTest()
  @MethodSource("provideRandomAluAddArguments")
  void aluAddTest(OpCode opCode, List<Bytes32> args) {
    runTest(opCode, args);
  }

  @ParameterizedTest()
  @MethodSource("provideSimpleAluAddArguments")
  void simpleAddTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @Override
  protected Stream<Arguments> provideNonRandomArguments() {
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

  public Stream<Arguments> provideSimpleAluAddArguments() {
    List<Arguments> arguments = new ArrayList<>();

    Bytes32 bytes1 = Bytes32.rightPad(Bytes.fromHexString("0x80"));
    Bytes32 bytes2 = Bytes32.leftPad(Bytes.fromHexString("0x01"));
    OpCode opCode = OpCode.SUB;
    arguments.add(Arguments.of(opCode, bytes1, bytes2));
    return arguments.stream();
  }

  public Stream<Arguments> provideRandomAluAddArguments() {
    List<Arguments> arguments = new ArrayList<>();

    for (int i = 0; i < TEST_ADD_REPETITIONS; i++) {
      arguments.add(getRandomAluAddInstruction(rand.nextInt(32) + 1, rand.nextInt(32) + 1));
    }
    return arguments.stream();
  }

  private Arguments getRandomAluAddInstruction(int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    OpCode opCode = getRandomSupportedOpcode();
    return Arguments.of(opCode, List.of(bytes1, bytes2));
  }

  @Override
  protected ModuleTracer getModuleTracer() {
    return new AddTracer();
  }
}
