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

package net.consensys.linea.zktracer.exceptions;

import static net.consensys.linea.zktracer.module.hub.signals.TracedException.OUT_OF_GAS_EXCEPTION;
import static net.consensys.linea.zktracer.opcode.OpCodes.opCodeToOpCodeDataMap;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(UnitTestWatcher.class)
public class OutOfGasExceptionTest {

  // TODO: add tests when address is warm. Use constants such as G_WARM_ACCESS etc
  @ParameterizedTest
  @MethodSource("outOfGasExceptionSource")
  void outOfGasExceptionColdTest(
      OpCode opCode, int staticCost, int nPushes, boolean triggersOutOfGasExceptions) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    for (int i = 0; i < nPushes; i++) {
      program.push(0);
    }
    program.op(opCode);
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run(
        (long) 21000 + nPushes * 3L + staticCost - (triggersOutOfGasExceptions ? 1 : 0));
    if (triggersOutOfGasExceptions) {
      assertEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    } else {
      assertNotEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    }
  }

  static Stream<Arguments> outOfGasExceptionSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (OpCodeData opCodeData : opCodeToOpCodeDataMap.values()) {
      OpCode opCode = opCodeData.mnemonic();
      int staticCost = opCodeData.stackSettings().staticGas().cost();
      int delta = opCodeData.stackSettings().delta(); // number of items popped from the stack
      // TODO: some opCodes are excluded for now because they may need to be treated differently
      if (staticCost > 0
          && opCode != OpCode.MLOAD
          && opCode != OpCode.MSTORE8
          && opCode != OpCode.SELFDESTRUCT
          && opCode != OpCode.MSTORE) {
        arguments.add(Arguments.of(opCode, staticCost, delta, true));
        arguments.add(Arguments.of(opCode, staticCost, delta, false));
      }
    }
    return arguments.stream();
  }
}
