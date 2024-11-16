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
package net.consensys.linea.zktracer.instructionprocessing.callTests.prc.ecrecover;

import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.*;
import static net.consensys.linea.zktracer.opcode.OpCode.CALL;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

/**
 * For these tests to work as expected, the transaction should start out with sufficient gas. At
 * least 21k plus gas to cover the PUSH's and the (warm) CALL with potentially value transfer and
 * potentially account creation costs. Also there should be enough gas in the end for the 63/64
 * business not diminish the gas we provide the callee.
 *
 * <p>Something like 60k gas should cover all costs (21k transaction costs + 9k for potential value
 * transfer + 25k if value transfer leads to a precompile starting to exist in the state etc ... +
 * 3k for the callee + opcode costs on the order of 130 or so)
 */
public class gasStipendTests {

  // sufficient gas for PRC execution
  @Test
  void zeroValueEcrecoverCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    validEcrecoverData(program);
    simpleCall(
        program, CALL, 3000, Address.fromHexString(Address.ECREC.toHexString()), 0, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  @Test
  void nonzeroValueEcrecoverCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    validEcrecoverData(program);
    simpleCall(
        program, CALL, 3000, Address.fromHexString(Address.ECREC.toHexString()), 1, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  @Test
  void nonzeroValueStipendCompletesGasEcrecoverCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    validEcrecoverData(program);
    simpleCall(
        program, CALL, 700, Address.fromHexString(Address.ECREC.toHexString()), 1, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  // insufficient gas for PRC execution
  @Test
  void nonzeroValueShortOnGasEcrecoverCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    validEcrecoverData(program);
    simpleCall(
        program, CALL, 2999, Address.fromHexString(Address.ECREC.toHexString()), 1, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  @Test
  void nonzeroValueStipendFallsShortOfCompletingGasEcrecoverCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    validEcrecoverData(program);
    simpleCall(
        program, CALL, 699, Address.fromHexString(Address.ECREC.toHexString()), 1, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }
}
