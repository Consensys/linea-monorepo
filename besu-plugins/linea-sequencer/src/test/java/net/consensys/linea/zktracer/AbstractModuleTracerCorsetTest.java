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

import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.TestInstance;
import org.junit.jupiter.api.TestInstance.Lifecycle;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@TestInstance(Lifecycle.PER_CLASS)
public abstract class AbstractModuleTracerCorsetTest extends AbstractBaseModuleTracerTest {
  static final Random rand = new Random();
  private static final int TEST_REPETITIONS = 8;

  @ParameterizedTest()
  @MethodSource("provideRandomArguments")
  void randomArgumentsTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, arg1, arg2);
  }

  @ParameterizedTest()
  @MethodSource("provideNonRandomArguments")
  void nonRandomArgumentsTest(OpCode opCode, final Bytes32 arg1, Bytes32 arg2) {
    runTest(opCode, arg1, arg2);
  }

  protected abstract Stream<Arguments> provideNonRandomArguments();

  public Stream<Arguments> provideRandomArguments() {
    final List<Arguments> arguments = new ArrayList<>();
    for (OpCode opCode : getModuleTracer().supportedOpCodes()) {
      for (int i = 0; i <= TEST_REPETITIONS; i++) {
        arguments.add(Arguments.of(opCode, Bytes32.random(rand), Bytes32.random(rand)));
      }
    }
    return arguments.stream();
  }

  public OpCode getRandomSupportedOpcode() {
    var supportedOpCodes = getModuleTracer().supportedOpCodes();
    int index = rand.nextInt(supportedOpCodes.size());
    return supportedOpCodes.get(index);
  }
}
