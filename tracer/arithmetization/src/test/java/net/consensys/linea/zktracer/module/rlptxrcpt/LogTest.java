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

package net.consensys.linea.zktracer.module.rlptxrcpt;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class LogTest {
  @Test
  void log2Test() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .push(
                    Bytes32.fromHexString(
                        "0x00112233445566778899AABBCCDDEE0000112233445566778899AABBCCDDEE00"))
                .push(1)
                .op(OpCode.MSTORE)
                .push(Bytes32.fromHexString("0x00070b1c100000070b1c100000070b1c100000070b1c1000"))
                .push(Bytes32.fromHexString("0x00070b1c200000070b1c200000070b1c200000070b1c2000"))
                .push(33) // size
                .push(4) // offset
                .op(OpCode.LOG2)
                .push(1)
                .push(0)
                .op(OpCode.MSTORE)
                .compile())
        .run();
  }
}
