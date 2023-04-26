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
package net.consensys.linea.zktracer.module.alu.mul;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.AbstractModuleTracerCorsetTest;
import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class MulTracerTest extends AbstractModuleTracerCorsetTest {
  private static final Random rand = new Random();

  private static final int TEST_MUL_REPETITIONS = 16;

  @ParameterizedTest()
  @MethodSource("provideRandomAluMulArguments")
  void aluModTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("singleTinyExponentiation")
  void testSingleTinyExponentiation(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("provideTinyArguments")
  void tinyArgsTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("provideSpecificNonTinyArguments")
  void nonTinyArgsTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("provideRandomNonTinyArguments")
  void randomNonTinyArgsTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  @ParameterizedTest()
  @MethodSource("multiplyByZero")
  void zerosArgsTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, List.of(arg1, arg2));
  }

  public Stream<Arguments> singleTinyExponentiation() {
    List<Arguments> arguments = new ArrayList<>();

    Bytes32 bytes1 = Bytes32.leftPad(Bytes.fromHexString("0x13"));
    Bytes32 bytes2 = Bytes32.leftPad(Bytes.fromHexString("0x02"));
    arguments.add(Arguments.of(OpCode.EXP, bytes1, bytes2));
    return arguments.stream();
  }

  public Stream<Arguments> provideRandomAluMulArguments() {
    List<Arguments> arguments = new ArrayList<>();

    for (int i = 0; i < TEST_MUL_REPETITIONS; i++) {
      arguments.add(getRandomAluMulInstruction(rand.nextInt(32) + 1, rand.nextInt(32) + 1));
    }
    return arguments.stream();
  }

  private Arguments getRandomAluMulInstruction(int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    OpCode opCode = getRandomSupportedOpcode();
    return Arguments.of(opCode, bytes1, bytes2);
  }

  public Stream<Arguments> provideSpecificNonTinyArguments() {
    //    these values are used in Go module test
    //    0x8a, 0x48, 0xaa, 0x20, 0xe2, 0x00, 0xce, 0x3f, 0xee, 0x16, 0xb5, 0xdc, 0xde, 0xc5, 0xc4,
    // 0xfa,
    //            0xff, 0x61, 0x3b, 0xc9, 0x14, 0xd4, 0x7c, 0xd6, 0xca, 0x69, 0x55, 0x3f, 0x8e,
    // 0xb2, 0xb3, 0x77,
    //		byte(vm.PUSH32),
    //            0x59, 0xb6, 0x35, 0xfe, 0xc8, 0x94, 0xca, 0xa3, 0xed, 0x68, 0x17, 0xb1, 0xe6,
    // 0x7b, 0x3c, 0xba,
    //            0xeb, 0x87, 0x57, 0xfd, 0x6c, 0x7b, 0x03, 0x11, 0x9b, 0x79, 0x53, 0x03, 0xb7,
    // 0xcd, 0x72, 0xc1,
    final Bytes32[] payload = new Bytes32[2];
    payload[0] =
        Bytes32.fromHexString("0x8a48aa20e200ce3fee16b5dcdec5c4faff613bc914d47cd6ca69553f8eb2b377");
    payload[1] =
        Bytes32.fromHexString("0x59b635fec894caa3ed6817b1e67b3cbaeb8757fd6c7b03119b795303b7cd72c1");
    return Stream.of(Arguments.of(OpCode.MUL, payload[0], payload[1]));
  }

  public Stream<Arguments> provideRandomNonTinyArguments() {
    List<Arguments> arguments = new ArrayList<>();

    for (int i = 0; i < TEST_MUL_REPETITIONS; i++) {
      arguments.add(getRandomAluMulInstruction(rand.nextInt(32) + 1, rand.nextInt(32) + 1));
    }
    return arguments.stream();
  }

  public Stream<Arguments> provideTinyArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (int i = 0; i < 4; i++) {
      arguments.add(getRandomAluMulInstruction(i, i + 1));
    }
    return arguments.stream();
  }

  @Override
  protected Stream<Arguments> provideNonRandomArguments() {
    List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int k = 0; k <= 3; k++) {
        for (int i = 0; i <= 3; i++) {
          arguments.add(Arguments.of(opCode, List.of(UInt256.valueOf(i), UInt256.valueOf(k))));
        }
      }
    }
    return arguments.stream();
  }

  protected Stream<Arguments> multiplyByZero() {
    List<Arguments> arguments = new ArrayList<>();
    for (int i = 0; i < 2; i++) {
      Bytes32 bytes1 = Bytes32.ZERO;
      Bytes32 bytes2 = UInt256.valueOf(i);
      arguments.add(Arguments.of(OpCode.MUL, bytes1, bytes2));
      arguments.add(Arguments.of(OpCode.MUL, bytes2, bytes1));
    }
    return arguments.stream();
  }

  @Override
  protected ModuleTracer getModuleTracer() {
    return new MulTracer();
  }
}
