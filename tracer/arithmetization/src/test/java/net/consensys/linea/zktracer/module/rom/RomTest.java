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

package net.consensys.linea.zktracer.module.rom;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.stream.Stream;

import kotlin.Pair;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class RomTest extends TracerTestBase {

  @Test
  void oneIncompletePushTest(TestInfo testInfo) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    program.incompletePush(12, "ff".repeat(4));
    BytecodeRunner.of(program.compile()).run(chainConfig, testInfo);
  }

  @Tag("nightly")
  @ParameterizedTest
  @MethodSource("incompletePushTestSource")
  void extensiveIncompletePushTest(int j, int k, TestInfo testInfo) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    // Bytes taken by a PUSHX do not have a specific purpose here
    program.incompletePush(k, "ff".repeat(j));
    BytecodeRunner.of(program.compile()).run(chainConfig, testInfo);
  }

  private static Stream<Arguments> incompletePushTestSource() {
    List<Arguments> trailingFFRomTestSourceList = new ArrayList<>();
    for (int k = 1; k <= 32; k++) {
      for (int j = 0; j <= k; j++) {
        trailingFFRomTestSourceList.add(Arguments.of(j, k));
      }
    }
    return trailingFFRomTestSourceList.stream();
  }

  /**
   * The bytecode constructed in the following test is a random concatenation of incomplete pushes
   * where every "incomplete push" is made up of some <b>PUSHX</b> opcode follwed by
   *
   * <pre> l := 0, 1, ..., X </pre>
   *
   * bytes with value "5b", i.e. the byte value of <b>JUMPDEST</b>.
   *
   * <p>The execution of this code is therefore a mixture of <b>PUSHX</b>'s and <b>JUMPDEST</b>'s.
   *
   * <p>The purpose is to test ROM module's ability to correctly perform "jump destination analysis"
   * i.e. its ability to distinguish between valid <b>JUMPDEST</b>'s and invalid ones, i.e. "5b"'s
   * claimed by some <b>PUSHX</b> opcode.
   */
  @Test
  void jumpDestinationAnalysisTest(TestInfo testInfo) {
    List<Pair<Integer, Integer>> permutationOfKAndJPairs = new ArrayList<>();
    for (int k = 1; k <= 32; k++) {
      for (int j = 0; j <= k; j++) {
        permutationOfKAndJPairs.add(new Pair<>(k, j));
      }
    }
    Collections.shuffle(permutationOfKAndJPairs);

    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    for (Pair<Integer, Integer> kAndJPair : permutationOfKAndJPairs) {
      int k = kAndJPair.getFirst();
      int j = kAndJPair.getSecond();
      program.incompletePush(k, "5b".repeat(j)); // invalid JUMPDEST
    }

    BytecodeRunner.of(program.compile()).run(chainConfig, testInfo);
  }
}
