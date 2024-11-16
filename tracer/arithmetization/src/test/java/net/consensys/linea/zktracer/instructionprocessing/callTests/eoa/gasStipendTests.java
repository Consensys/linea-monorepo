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
package net.consensys.linea.zktracer.instructionprocessing.callTests.eoa;

import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.*;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

/**
 * Transferring nonzero value provides a gas stipend to the callee. This stipend will immediately be
 * restituted to the caller in case of an EOA call.
 */
public class gasStipendTests {

  @Test
  void zeroValueEoaCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 0, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }

  // The caller should recover and extra G_stipend = 2300 gas.
  @Test
  void nonzeroValueEoaCallTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    simpleCall(program, CALL, 0, Address.fromHexString(eoaAddress), 1, 0, 0, 0, 0);

    BytecodeRunner.of(program.compile()).run();
  }
}
