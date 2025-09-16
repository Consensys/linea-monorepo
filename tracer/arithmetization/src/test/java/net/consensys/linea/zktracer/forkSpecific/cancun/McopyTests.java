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

package net.consensys.linea.zktracer.forkSpecific.cancun;

import static net.consensys.linea.zktracer.Trace.*;
import static net.consensys.linea.zktracer.TraceCancun.Mxp.CANCUN_MXPX_THRESHOLD;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

public class McopyTests extends TracerTestBase {

  @ParameterizedTest
  @MethodSource("inputParamsUnit")
  void McopyLight(Bytes targetOffset, Bytes sourceOffset, Bytes size, TestInfo testInfo) {
    singleMcopy(targetOffset, sourceOffset, size, testInfo);
  }

  @Tag("nightly")
  @ParameterizedTest
  @MethodSource("inputParamsNightly")
  void McopyExtensive(Bytes targetOffset, Bytes sourceOffset, Bytes size, TestInfo testInfo) {
    singleMcopy(targetOffset, sourceOffset, size, testInfo);
  }

  // Parameterized method
  private static Stream<Arguments> inputs(List<Bytes32> inputsValues) {
    final List<Arguments> arguments = new ArrayList<>();
    for (Bytes32 targetOffset : inputsValues) {
      for (Bytes32 sourceOffset : inputsValues) {
        for (Bytes32 size : inputsValues) {
          arguments.add(Arguments.of(targetOffset, sourceOffset, size));
        }
      }
    }
    return arguments.stream();
  }

  private static Stream<Arguments> inputParamsNightly() {
    return inputs(inputsValuesNightly);
  }

  private static Stream<Arguments> inputParamsUnit() {
    return inputs(inputsValuesUnit);
  }

  private static final List<Bytes32> inputsValuesUnit =
      List.of(
          Bytes32.ZERO,
          Bytes32.leftPad(Bytes.ofUnsignedInt(1)),
          Bytes32.leftPad(Bytes.ofUnsignedLong(CANCUN_MXPX_THRESHOLD - 1)),
          Bytes32.leftPad(Bytes.ofUnsignedLong(CANCUN_MXPX_THRESHOLD)),
          Bytes32.repeat((byte) 0xff));

  private static final List<Bytes32> inputsValuesNightly =
      Stream.concat(
              inputsValuesUnit.stream(),
              Stream.of(
                  Bytes32.leftPad(Bytes.ofUnsignedInt(LLARGEMO)),
                  Bytes32.leftPad(Bytes.ofUnsignedInt(LLARGE)),
                  Bytes32.leftPad(Bytes.ofUnsignedInt(LLARGEPO)),
                  Bytes32.leftPad(Bytes.ofUnsignedInt(WORD_SIZE_MO)),
                  Bytes32.leftPad(Bytes.ofUnsignedInt(WORD_SIZE)),
                  Bytes32.leftPad(Bytes.ofUnsignedInt(33)),
                  Bytes32.leftPad(Bytes.ofUnsignedLong(Long.MAX_VALUE)),
                  Bytes32.leftPad(Bytes.ofUnsignedLong(CANCUN_MXPX_THRESHOLD + 1))))
          .toList();

  // Main test
  private void singleMcopy(Bytes targetOffset, Bytes sourceOffset, Bytes size, TestInfo testInfo) {

    final Bytes FILL_MEMORY =
        BytecodeCompiler.newProgram(chainConfig)
            .push(
                Bytes32.fromHexString(
                    "0x11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff")) // value
            .push(0) // offset
            .op(MSTORE)
            .compile();

    final Bytes MLOADS =
        BytecodeCompiler.newProgram(chainConfig)
            .push(0)
            .op(MLOAD)
            .op(POP)
            .push(WORD_SIZE)
            .op(MLOAD)
            .op(POP)
            .push(2 * WORD_SIZE)
            .op(MLOAD)
            .op(POP)
            .compile();

    BytecodeRunner.of(
            Bytes.concatenate(
                FILL_MEMORY, // We fill the first 32 bytes of memory with non-trivial value
                pushAndMcopy(targetOffset, sourceOffset, size), // We perform the MCOPY
                MLOADS // We load the first 3 words of memory to check the result
                ))
        .run(chainConfig, testInfo);
  }

  private Bytes pushAndMcopy(Bytes targetOffset, Bytes sourceOffset, Bytes size) {
    return Bytes.concatenate(
        Bytes.of(PUSH32.byteValue()),
        Bytes32.leftPad(size),
        Bytes.of(PUSH32.byteValue()),
        Bytes32.leftPad(sourceOffset),
        Bytes.of(PUSH32.byteValue()),
        Bytes32.leftPad(targetOffset),
        Bytes.fromHexString("0x5E") // MCOPY opcode
        );
  }
}
