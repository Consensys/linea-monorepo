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
package net.consensys.linea.zktracer.instructionprocessing.callTests.abort;

import static net.consensys.linea.zktracer.opcode.OpCode.*;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;

/**
 * Attempt to trigger the maximum call stack depth abort. We put everything to 0 to avoid memory
 * expansion costs. We will want to revert so we transfer value to see the effect of reverting.
 */
public class CallStackDepthAbortTests {
  @Test
  void attemptAtCallStackDepthAbortWillRevert() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(0)
            .push(0)
            .push(0)
            .push(0) //
            .push(5) // value
            .op(ADDRESS) // current address
            .op(GAS) // providing all available gas
            .op(CALL) // self-call
            .push(6)
            .push(7)
            .op(REVERT)
            .compile();

    BytecodeRunner.of(bytecode).run();
  }

  @Test
  void attemptAtCallStackDepthAbortWontRevert() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(0)
            .push(0)
            .push(0)
            .push(0)
            .op(ADDRESS)
            .op(GAS) // providing as much gas as possible
            .op(STATICCALL) // self-call
            .compile();

    BytecodeRunner.of(bytecode).run();
  }
}
