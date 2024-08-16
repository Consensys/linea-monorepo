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

package net.consensys.linea.zktracer.module.ecdata;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;

public class EcAddTest {

  @Test
  void testEcAddGeneric() {
    // TODO: The same inputs in failingMmuModexp return 0x, debug it!
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            // First place the parameters in memory
            .push("070375d4eec4f22aa3ad39cb40ccd73d2dbab6de316e75f81dc2948a996795d5") // pX
            .push(0)
            .op(OpCode.MSTORE)
            .push("041b98f07f44aa55ce8bd97e32cacf55f1e42229d540d5e7a767d1138a5da656") // pY
            .push(0x20)
            .op(OpCode.MSTORE)
            .push("185f6f5cf93c8afa0461a948c2da7c403b6f8477c488155dfa8d2da1c62517b8") // qX
            .push(0x40)
            .op(OpCode.MSTORE)
            .push("13d83d7a51eb18fdb51225873c87d44f883e770ce2ca56c305d02d6cb99ca5b8") // qY
            .push(0x60)
            .op(OpCode.MSTORE)
            // Do the call
            .push(0x40) // retSize
            .push(0x80) // retOffset
            .push(0x80) // argSize
            .push(0) // argOffset
            .push(6) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Test
  void testEcAddWithPointAtInfinityAsResult() {
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            // First place the parameters in memory
            .push(1) // pX
            .push(0)
            .op(OpCode.MSTORE)
            .push(2) // pY
            .push(0x20)
            .op(OpCode.MSTORE)
            .push(1) // qX
            .push(0x40)
            .op(OpCode.MSTORE)
            .push("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd45") // qY
            .push(0x60)
            .op(OpCode.MSTORE)
            // Do the call
            .push(0x40) // retSize
            .push(0x80) // retOffset
            .push(0x80) // argSize
            .push(0) // argOffset
            .push(6) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }
}
