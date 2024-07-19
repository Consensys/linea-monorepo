/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.mmu;

import java.util.Random;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.BytecodeRunner;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;

class MemoryTests {
  private final Random rnd = new Random(666);

  @Test
  void successionOverlappingMstore() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .push(Bytes.repeat((byte) 1, 32))
                .push(0)
                .op(OpCode.MSTORE)
                .push(Bytes.repeat((byte) 2, 32))
                .push(15)
                .op(OpCode.MSTORE)
                .push(Bytes.repeat((byte) 3, 32))
                .push(2)
                .op(OpCode.MSTORE)
                .push(6)
                .op(OpCode.MLOAD)
                .compile())
        .run();
  }

  @Test
  void fastMload() {
    BytecodeRunner.of(BytecodeCompiler.newProgram().push(34).push(0).op(OpCode.MLOAD).compile())
        .run();
  }

  @Test
  void alignedMstore8() {
    BytecodeRunner.of(BytecodeCompiler.newProgram().push(12).push(0).op(OpCode.MSTORE8).compile())
        .run();
  }

  @Test
  void nonAlignedMstore8() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram().push(66872).push(35).op(OpCode.MSTORE8).compile())
        .run();
  }
}
