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

import static net.consensys.linea.zktracer.MultiBlockUtils.multiBlocksTest;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.BLOCKHASH_MAX_HISTORY;

import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

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
  void severalBlockhash() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()

                // arg is NUMBER - 1
                .push(1)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is NUMBER
                .op(OpCode.NUMBER)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is NUMBER + 1
                .op(OpCode.NUMBER)
                .push(1)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is ridiculously big
                .push(256)
                .op(OpCode.NUMBER)
                .op(OpCode.MUL)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is NUMBER / 256 << NUMBER
                .push(256)
                .op(OpCode.NUMBER)
                .op(OpCode.DIV)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is 0 << NUMBER
                .push(0)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is 1 << NUMBER
                .push(1)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg is ridiculously big
                .push(
                    Bytes.fromHexString(
                        "0x123456789012345678901234567890123456789012345678901234567890ffff"))
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is NUMBER - (256 + 2)
                .push(BLOCKHASH_MAX_HISTORY + 2)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is NUMBER - (256 + 1)
                .push(BLOCKHASH_MAX_HISTORY + 1)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is NUMBER - 256
                .push(BLOCKHASH_MAX_HISTORY)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is NUMBER - (256 - 1)
                .push(BLOCKHASH_MAX_HISTORY - 1)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)

                // arg of BlockHash is NUMBER - (256 - 2)
                .push(BLOCKHASH_MAX_HISTORY - 2)
                .op(OpCode.NUMBER)
                .op(OpCode.SUB)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)
                .compile())
        .run();
  }

  @Test
  void singleBlockhash() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram()

                // arg of BlockHash is Blocknumber +1
                .op(OpCode.NUMBER)
                .push(1)
                .op(OpCode.ADD)
                .op(OpCode.BLOCKHASH)
                .op(OpCode.POP)
                .compile())
        .run();
  }

  @Test
  void blockhashArgumentUpperRangeCheckMultiBlockTest() {
    // Block 1
    Bytes program1 = BytecodeCompiler.newProgram().op(OpCode.NUMBER).op(OpCode.BLOCKHASH).compile();

    // Block 2
    Bytes program2 =
        BytecodeCompiler.newProgram()
            .push(1)
            .op(OpCode.NUMBER)
            .op(OpCode.SUB)
            .op(OpCode.BLOCKHASH)
            .compile();

    multiBlocksTest(List.of(program1, program2));
  }

  @Test
  void blockhashArgumentLowerRangeCheckMultiBlockTest() {
    Bytes fillerProgram = BytecodeCompiler.newProgram().op(OpCode.COINBASE).compile();

    Bytes program0 =
        BytecodeCompiler.newProgram()
            .push(1)
            .op(OpCode.NUMBER)
            .op(OpCode.SUB)
            .op(OpCode.BLOCKHASH)
            .compile();

    // Block no longer available
    // Block 1
    Bytes program1 =
        BytecodeCompiler.newProgram()
            .push(256)
            .op(OpCode.NUMBER)
            .op(OpCode.SUB)
            .op(OpCode.BLOCKHASH)
            .compile();

    // Block 2
    Bytes program2 =
        BytecodeCompiler.newProgram()
            .push(257)
            .op(OpCode.NUMBER)
            .op(OpCode.SUB)
            .op(OpCode.BLOCKHASH)
            .compile();

    Bytes program3 =
        BytecodeCompiler.newProgram()
            .push(16)
            .op(OpCode.NUMBER)
            .op(OpCode.SUB)
            .op(OpCode.BLOCKHASH)
            .compile();

    multiBlocksTest(
        Stream.concat(
                Collections.nCopies(256, fillerProgram).stream(),
                List.of(program0, program1, program2, program3).stream())
            .collect(Collectors.toList()));
  }
}
