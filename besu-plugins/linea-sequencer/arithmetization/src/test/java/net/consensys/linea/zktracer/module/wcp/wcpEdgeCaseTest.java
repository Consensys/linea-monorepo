/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.wcp;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.BytecodeRunner;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;

public class wcpEdgeCaseTest {
  @Test
  void testZeroAndHugeArgs() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .push(Bytes.repeat((byte) 0xff, 32))
                .push(Bytes.EMPTY)
                .op(OpCode.SLT)
                .compile())
        .run();
  }

  @Test
  void testHugeAndZeroArgs() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .push(Bytes.EMPTY)
                .push(Bytes.repeat((byte) 0xff, 32))
                .op(OpCode.SLT)
                .compile())
        .run();
  }

  @Test
  void failingOnShadowNodeBlock916394() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .push(Bytes.EMPTY)
                .push(
                    Bytes.concatenate(
                        Bytes.repeat((byte) 0xff, 29),
                        Bytes.of(0xfe),
                        Bytes.of(0x18),
                        Bytes.of(0x59)))
                .op(OpCode.SLT)
                .compile())
        .run();
  }

  @Test
  void failingOnShadowNodeBlockWhatever() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()
                .push(Bytes.EMPTY)
                .push(
                    Bytes.fromHexString(
                        "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe1859"))
                .op(OpCode.SLT)
                .compile())
        .run();
  }
}
