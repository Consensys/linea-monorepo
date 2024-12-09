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
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class RomTest {

  @ParameterizedTest
  @MethodSource("incompletePushRomTestSource")
  void incompletePushRomTest(int j, int k) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    program.incompletePush(k, "ff".repeat(j));
    BytecodeRunner.of(program.compile()).run();
  }

  private static Stream<Arguments> incompletePushRomTestSource() {
    List<Arguments> trailingFFRomTestSourceList = new ArrayList<>();
    for (int k = 1; k <= 32; k++) {
      for (int j = 0; j <= k; j++) {
        trailingFFRomTestSourceList.add(Arguments.of(j, k));
      }
    }
    return trailingFFRomTestSourceList.stream();
  }

  @Test
  void randomConcatenationOfIncompletePushesRomTest() {
    List<Pair<Integer, Integer>> permutationOfKAndJPairs = new ArrayList<>();
    for (int k = 1; k <= 32; k++) {
      for (int j = 0; j <= k; j++) {
        permutationOfKAndJPairs.add(new Pair<>(k, j));
      }
    }
    Collections.shuffle(permutationOfKAndJPairs);

    BytecodeCompiler program = BytecodeCompiler.newProgram();
    for (Pair<Integer, Integer> kAndJPair : permutationOfKAndJPairs) {
      int k = kAndJPair.getFirst();
      int j = kAndJPair.getSecond();
      program.incompletePush(k, "5b".repeat(j));
    }

    BytecodeRunner.of(program.compile()).run();
  }
}
