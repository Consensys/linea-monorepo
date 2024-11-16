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
package net.consensys.linea.zktracer.instructionprocessing.callTests;

import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.*;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

public class DoubleCall {

  /** Same address */
  @Test
  void doubleCallToSameAddressWontRevert() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 1, 0, 0, 0, 0);

    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 2, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  @Test
  void doubleCallToSameAddressWillRevert() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 1, 0, 0, 0, 0);

    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 2, 0, 0, 0, 0);

    program.op(REVERT);

    BytecodeRunner.of(program.compile()).run();
  }

  /** Different address */
  @Test
  void doubleCallTodifferentAddressesWontRevert() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 1, 0, 0, 0, 0);

    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress2), 2, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  @Test
  void doubleCallTodifferentAddressesWillRevert() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 1, 0, 0, 0, 0);

    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress2), 2, 0, 0, 0, 0);

    program.push(13).push(71); // the stack already contains two items but why not ...
    program.op(REVERT);

    BytecodeRunner.of(program.compile()).run();
  }
}
