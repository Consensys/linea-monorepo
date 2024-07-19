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

package net.consensys.linea.zktracer.module.bin;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.BytecodeRunner;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.Test;

public class BinTest {
  @Test
  public void edgeCase() {
    BytecodeRunner.of(BytecodeCompiler.newProgram().push(0xf0).push(0xf0).op(OpCode.AND).compile())
        .run();
  }

  @Test
  void testSignedSignextend() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .immediate(UInt256.MAX_VALUE)
                .immediate(UInt256.MAX_VALUE)
                .op(OpCode.SIGNEXTEND)
                .compile())
        .run();

    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .immediate(UInt256.valueOf(31))
                .immediate(UInt256.MAX_VALUE)
                .op(OpCode.SIGNEXTEND)
                .compile())
        .run();

    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .immediate(UInt256.valueOf(32))
                .immediate(UInt256.MAX_VALUE)
                .op(OpCode.SIGNEXTEND)
                .compile())
        .run();
  }
}
