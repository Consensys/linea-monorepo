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

package net.consensys.linea.zktracer.testing;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.stream.Stream;

import com.google.common.collect.ArrayListMultimap;
import com.google.common.collect.Multimap;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.DynamicTest;

/**
 * Responsible for executing JUnit 5 dynamic tests for modules with the ability to configure custom
 * opcode arguments on a test case basis.
 */
public class DynamicTests {
  private static final Random RAND = new Random();
  private static final int TEST_REPETITIONS = 8;

  private final Map<String, Multimap<OpCode, Bytes32>> testCaseRegistry;

  private final Module module;

  private DynamicTests(Module module) {
    OpCodes.load();
    this.module = module;
    this.testCaseRegistry = new LinkedHashMap<>();
    this.testCaseRegistry.put("random arguments test", provideRandomArguments());
  }

  /**
   * Constructor function for initialization of a suite of dynamic tests per module.
   *
   * @param module module instance for which a dynamic test suite should be run
   * @return an instance of {@link DynamicTests}
   */
  public static DynamicTests forModule(Module module) {
    return new DynamicTests(module);
  }

  /**
   * Constructs a new dynamically generated test case.
   *
   * @param testCaseName name of the test case
   * @param args arguments of the test case
   * @return the current instance
   */
  public DynamicTests testCase(final String testCaseName, final Multimap<OpCode, Bytes32> args) {
    testCaseRegistry.put(testCaseName, args);

    return this;
  }

  /**
   * Runs the suite of dynamic tests per module.
   *
   * @return a {@link Stream} of {@link DynamicTest} ran by a {@link
   *     org.junit.jupiter.api.TestFactory}
   */
  public Stream<DynamicTest> run() {
    return this.testCaseRegistry.entrySet().stream()
        .flatMap(e -> generateTestCases(e.getKey(), e.getValue()));
  }

  /**
   * Get a random {@link OpCode} from the list of supported opcodes per module tracer.
   *
   * @return a random {@link OpCode} from the list of supported opcodes per module tracer.
   */
  public OpCode getRandomSupportedOpcode() {
    List<OpCode> supportedOpCodes = module.supportedOpCodes();
    int index = RAND.nextInt(supportedOpCodes.size());

    return supportedOpCodes.get(index);
  }

  private Stream<DynamicTest> generateTestCases(
      final String testCaseName, final Multimap<OpCode, Bytes32> args) {
    return args.asMap().entrySet().stream()
        .map(
            e -> {
              OpCode opCode = e.getKey();
              List<Bytes32> opArgs = e.getValue().stream().toList();

              String testName =
                  "[%s][%s] Test bytecode for opcode %s with opArgs %s"
                      .formatted(module.jsonKey().toUpperCase(), testCaseName, opCode, opArgs);

              return DynamicTest.dynamicTest(
                  testName, () -> ModuleTests.runTestWithOpCodeArgs(opCode, opArgs));
            });
  }

  /**
   * Converts supported opcodes per module tracer into a {@link Multimap} for dynamic tests.
   *
   * @return a {@link Stream} of {@link DynamicTests} for dynamic test execution via {@link
   *     org.junit.jupiter.api.TestFactory}.
   */
  private Multimap<OpCode, Bytes32> provideRandomArguments() {
    final Multimap<OpCode, Bytes32> arguments = ArrayListMultimap.create();

    for (OpCode opCode : module.supportedOpCodes()) {
      for (int i = 0; i <= TEST_REPETITIONS; i++) {
        for (int j = 0; j < opCode.getData().numberOfArguments(); j++) {
          arguments.put(opCode, Bytes32.random(RAND));
        }
      }
    }

    return arguments;
  }
}
