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

import static net.consensys.linea.zktracer.opcode.OpCode.*;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;

/**
 * In the arithmetization there are the following EOA specific scenarios:
 *
 * <p>- <b>scn/CALL_EOA_SUCCESS_WILL_REVERT</b>
 *
 * <p>- <b>scn/CALL_EOA_SUCCESS_WONT_REVERT</b>
 */
public class EoaTests {

  final String eoaAddress = "abcdef0123456789";

  @Test
  void transfersValueWillRevertTest() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .push(5) // value
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL)
            .op(POP)
            .push(6)
            .push(7)
            .op(REVERT)
            .compile();

    BytecodeRunner.of(bytecode).run();
  }

  @Test
  void transfersValueWontRevertTest() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .push(5) // value
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL)
            .op(POP)
            .compile();

    BytecodeRunner.of(bytecode).run();
  }

  @Test
  void transfersAllValueWillRevertTest() {

    Bytes program =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .op(SELFBALANCE) // all our balance
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL)
            .op(POP)
            .push(6)
            .push(7)
            .op(REVERT)
            .compile();

    BytecodeRunner.of(program).run();
  }

  @Test
  void transfersAllValueWontRevertTest() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .op(SELFBALANCE) // all our balance
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL)
            .compile();

    BytecodeRunner.of(bytecode).run();
  }

  @Test
  void transfersNoValueWillRevertTest() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .push(0) // value
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL)
            .op(POP)
            .push(6)
            .push(7)
            .op(REVERT)
            .compile();

    BytecodeRunner.of(bytecode).run();
  }

  @Test
  void transfersNoValueWontRevertTest() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .push(0) // value
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL)
            .op(POP)
            .compile();

    BytecodeRunner.of(bytecode).run();
  }
}
