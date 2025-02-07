/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.shakira;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.WORD_SIZE;
import static net.consensys.linea.zktracer.opcode.OpCode.*;
import static net.consensys.linea.zktracer.types.Utils.rightPadTo;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@ExtendWith(UnitTestWatcher.class)
public class ShakiraInputsExtensiveTests {

  private final Random SEED = new Random(666);
  private final List<Integer> SIZE =
      List.of(0, 1, 2, 8, 15, 16, 17, 31, 32, 33, 254, 255, 256, 257, 258, 259);
  private final List<Integer> OFFSET =
      List.of(0, 1, 2, 15, 16, 17, 23, 31, 32, 33, 255, 256, 257, 65535, 65535, 65537);
  private final List<OpCode> INSTRUCTION =
      List.of(CALL, CALLCODE, STATICCALL, DELEGATECALL, SHA3, CREATE2, RETURN);

  private static final short CREATE_OPCODE_LENGTH = 21;

  @Tag("Weekly")
  @ParameterizedTest
  @MethodSource("inputs")
  void shakiraInputTesting(final int size, final int offset, final OpCode instruction) {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .op(CALLDATASIZE)
                .push(0)
                .push(0)
                .op(CALLDATACOPY)
                .immediate(instructionSpecificBytecode(size, offset, instruction))
                .compile())
        .run();
  }

  private Stream<Arguments> inputs() {
    final List<Arguments> inputs = new ArrayList<>();
    for (int size : SIZE) {
      for (int offset : OFFSET) {
        inputs.add(
            Arguments.of(size, offset, INSTRUCTION.get(SEED.nextInt(0, INSTRUCTION.size()))));
      }
    }
    return inputs.stream();
  }

  private Bytes instructionSpecificBytecode(int size, int offset, OpCode instruction) {
    switch (instruction) {
      case CALL, CALLCODE:
        {
          return BytecodeCompiler.newProgram()
              .push(SEED.nextInt(WORD_SIZE))
              .push(0)
              .push(size)
              .push(offset)
              .push(0)
              .push(Address.SHA256)
              .push(1000000)
              .op(instruction)
              .compile();
        }
      case STATICCALL, DELEGATECALL:
        {
          return BytecodeCompiler.newProgram()
              .push(SEED.nextInt(WORD_SIZE))
              .push(0)
              .push(size)
              .push(offset)
              .push(Address.SHA256)
              .push(1000000)
              .op(instruction)
              .compile();
        }
      case SHA3:
        {
          return BytecodeCompiler.newProgram().push(size).push(offset).op(SHA3).compile();
        }
      case RETURN:
        {
          return BytecodeCompiler.newProgram()
              .push(
                  rightPadTo(
                      Bytes.concatenate(
                          // size
                          Bytes.of(CALLDATASIZE.byteValue()),
                          Bytes.of(PUSH2.byteValue()),
                          Bytes.ofUnsignedShort(CREATE_OPCODE_LENGTH),
                          Bytes.of(SUB.byteValue()),
                          // offset
                          Bytes.of(PUSH2.byteValue()),
                          Bytes.ofUnsignedShort(CREATE_OPCODE_LENGTH),
                          // destOffset
                          Bytes.of(PUSH2.byteValue()),
                          Bytes.ofUnsignedShort(0),
                          Bytes.of(CALLDATACOPY.byteValue()),
                          Bytes.of(PUSH2.byteValue()),
                          Bytes.ofUnsignedShort(size),
                          Bytes.of(PUSH4.byteValue()),
                          Bytes.ofUnsignedInt(offset),
                          Bytes.of(RETURN.byteValue())),
                      WORD_SIZE))
              .push(0)
              .op(MSTORE)
              .op(CALLDATASIZE)
              .push(0)
              .push(CREATE_OPCODE_LENGTH)
              .op(CALLDATACOPY)
              .op(MSIZE) // size
              .push(0) // offset
              .push(0) // value
              .op(CREATE)
              .compile();
        }
      case CREATE2:
        {
          return BytecodeCompiler.newProgram()
              .push(Bytes32.random(SEED))
              .push(size)
              .push(offset)
              .push(12) // value
              .op(CREATE2)
              .compile();
        }
      default:
        throw new IllegalArgumentException("Unsupported instruction: " + instruction);
    }
  }
}
