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

package net.consensys.linea.zktracer.module.blockhash;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.BLOCKHASH_MAX_HISTORY;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class BlockhashTest {

  @Test
  void someBlockhash() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()

                // arg of BlockHash is Blocknumber +1
                .op(OpCode.NUMBER)
                .push(1)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber
                .op(OpCode.NUMBER)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is ridiculously big
                .push(256)
                .op(OpCode.NUMBER)
                .op(OpCode.MUL)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is ridiculously small
                .push(256)
                .op(OpCode.NUMBER)
                .op(OpCode.DIV)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is 0
                .push(0)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is 1 (ie ridiculously small)
                .push(1)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // another arg of BlockHash is ridiculously big
                .push(
                    Bytes.fromHexString(
                        "0x123456789012345678901234567890123456789012345678901234567890"))
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber -256 -2
                .push(BLOCKHASH_MAX_HISTORY + 2)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber -256 -1
                .push(BLOCKHASH_MAX_HISTORY + 1)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber -256
                .push(BLOCKHASH_MAX_HISTORY)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber -256 +1
                .push(BLOCKHASH_MAX_HISTORY - 1)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber -256 +2
                .push(BLOCKHASH_MAX_HISTORY - 2)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber  -1
                .push(1)
                .op(OpCode.NUMBER)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // Duplicate of arg of BlockHash is Blocknumber  -1
                .push(1)
                .op(OpCode.NUMBER)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // Truplicate of arg of BlockHash is Blocknumber  -1
                .push(1)
                .op(OpCode.NUMBER)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is Blocknumber  -2
                .push(2)
                .op(OpCode.NUMBER)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // TODO: add test with different block in the conflated batch

                .compile())
        .run();
  }
}
