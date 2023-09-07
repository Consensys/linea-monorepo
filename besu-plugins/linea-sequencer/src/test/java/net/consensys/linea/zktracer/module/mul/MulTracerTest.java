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

package net.consensys.linea.zktracer.module.mul;

import java.util.Random;
import java.util.stream.Stream;

import com.google.common.collect.ArrayListMultimap;
import com.google.common.collect.Multimap;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.DynamicTests;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;

class MulTracerTest {
  private static final Random RAND = new Random();

  private static final int TEST_MUL_REPETITIONS = 16;

  private static final Module MODULE = new Mul();

  private static final DynamicTests DYN_TESTS = DynamicTests.forModule(MODULE);

  @TestFactory
  Stream<DynamicTest> runDynamicTests() {
    return DYN_TESTS
        .testCase("non random arguments test", provideNonRandomArguments())
        .testCase(
            "single tiny exponentiation arguments test", provideSingleTinyExponentiationArguments())
        .testCase("random alu mul arguments test", provideRandomAluMulArguments())
        .testCase("random non tiny arguments test", provideRandomNonTinyArguments())
        .testCase("specific non tiny arguments test", provideSpecificNonTinyArguments())
        .testCase("tiny arguments test", provideTinyArguments())
        .testCase("multiplication by zero arguments test", provideMultiplicationByZeroArguments())
        .run();
  }

  private Multimap<OpCode, Bytes32> provideSingleTinyExponentiationArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    Bytes32 bytes1 = Bytes32.leftPad(Bytes.fromHexString("0x13"));
    Bytes32 bytes2 = Bytes32.leftPad(Bytes.fromHexString("0x02"));
    arguments.put(OpCode.EXP, bytes1);
    arguments.put(OpCode.EXP, bytes2);

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideRandomAluMulArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (int i = 0; i < TEST_MUL_REPETITIONS; i++) {
      addRandomAluMulInstruction(arguments, RAND.nextInt(32) + 1, RAND.nextInt(32) + 1);
    }
    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideSpecificNonTinyArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();
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

    arguments.put(OpCode.MUL, payload[0]);
    arguments.put(OpCode.MUL, payload[1]);

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideRandomNonTinyArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (int i = 0; i < TEST_MUL_REPETITIONS; i++) {
      addRandomAluMulInstruction(arguments, RAND.nextInt(32) + 1, RAND.nextInt(32) + 1);
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideTinyArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (int i = 0; i < 4; i++) {
      addRandomAluMulInstruction(arguments, i, i + 1);
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideNonRandomArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : MODULE.supportedOpCodes()) {
      for (int k = 0; k <= 3; k++) {
        for (int i = 0; i <= 3; i++) {
          arguments.put(opCode, UInt256.valueOf(i));
          arguments.put(opCode, UInt256.valueOf(k));
        }
      }
    }

    return arguments;
  }

  private Multimap<OpCode, Bytes32> provideMultiplicationByZeroArguments() {
    Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (int i = 0; i < 2; i++) {
      Bytes32 bytes1 = Bytes32.ZERO;
      Bytes32 bytes2 = UInt256.valueOf(i);

      arguments.put(OpCode.MUL, bytes1);
      arguments.put(OpCode.MUL, bytes2);
      arguments.put(OpCode.MUL, bytes2);
      arguments.put(OpCode.MUL, bytes1);
    }

    return arguments;
  }

  private void addRandomAluMulInstruction(
      Multimap<OpCode, Bytes32> arguments, int sizeArg1MinusOne, int sizeArg2MinusOne) {
    Bytes32 bytes1 = UInt256.valueOf(sizeArg1MinusOne);
    Bytes32 bytes2 = UInt256.valueOf(sizeArg2MinusOne);
    OpCode opCode = DYN_TESTS.getRandomSupportedOpcode();

    arguments.put(opCode, bytes1);
    arguments.put(opCode, bytes2);
  }
}
