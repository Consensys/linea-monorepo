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

package net.consensys.linea.zktracer.module.hub;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class TestTwoPlusTwo {
  @Test
  void testAssembler() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .assemble(
                    """
      ; Perform some addition
      PUSH8  02
      PUSH32 0x1234
      ADD
      """)
                .compile())
        .run();
  }

  @Test
  void ensureCorrectReturnDataInModexp() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .assemble(
                    """
; Call MODEXP
PUSH1 0     ; retSize
PUSH1 0     ; retOffset
PUSH1 0     ; argsSize
PUSH1 0     ; argsOffset
PUSH1 5     ; MODEXP precompile
PUSH4 20000 ; gas
STATICCALL
""")
                .compile())
        .run();
  }

  @Test
  void testBreakingCall() {
    // TODO: This test is disabled because it will throw an exception
    BytecodeRunner.of(BytecodeCompiler.newProgram().push(32).op(OpCode.CALL).compile()).run();
  }
}
