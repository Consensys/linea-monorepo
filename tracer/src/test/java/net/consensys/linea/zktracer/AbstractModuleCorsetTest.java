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

package net.consensys.linea.zktracer;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestInstance.Lifecycle;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

/**
 * Base class used for running tests suffixed with `CorsetTest`. Provides opcode arguments per
 * module as input for the parameterized tests.
 */
@TestInstance(Lifecycle.PER_CLASS)
@Tag("CorsetTest")
public abstract class AbstractModuleCorsetTest extends AbstractBaseModuleTest {
  static final Random rand = new Random();
  private static final int TEST_REPETITIONS = 8;

  @ParameterizedTest()
  @MethodSource("provideRandomArguments")
  void randomArgumentsTest(OpCodeData opCodeData, List<Bytes32> params) {
    runTest(opCodeData, params);
  }

  @ParameterizedTest()
  @MethodSource("provideNonRandomArguments")
  void nonRandomArgumentsTest(OpCodeData opCodeData, List<Bytes32> params) {
    runTest(opCodeData, params);
  }

  protected abstract Stream<Arguments> provideNonRandomArguments();

  /**
   * Converts supported opcodes per module tracer into {@link Arguments} for parameterized tests.
   *
   * @return a {@link Stream} of {@link Arguments} for parameterized tests.
   */
  public Stream<Arguments> provideRandomArguments() {
    final List<Arguments> arguments = new ArrayList<>();

    for (OpCodeData opCode : getModuleTracer().supportedOpCodes()) {
      for (int i = 0; i <= TEST_REPETITIONS; i++) {
        List<Bytes32> args = new ArrayList<>();
        for (int j = 0; j < opCode.numberOfArguments(); j++) {
          args.add(Bytes32.random(rand));
        }
        arguments.add(Arguments.of(opCode, args));
      }
    }

    return arguments.stream();
  }

  /**
   * Get a random {@link OpCode} from the list of supported opcodes per module tracer.
   *
   * @return a random {@link OpCode} from the list of supported opcodes per module tracer.
   */
  public OpCodeData getRandomSupportedOpcode() {
    List<OpCodeData> supportedOpCodes = getModuleTracer().supportedOpCodes();
    int index = rand.nextInt(supportedOpCodes.size());

    return supportedOpCodes.get(index);
  }
}
