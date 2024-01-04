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

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;
import java.util.Random;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.BytecodeRunner;
import net.consensys.linea.zktracer.testing.EvmExtension;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(EvmExtension.class)
public class BinTest {
  private static final Random RAND = new Random(666);
  private final int NB_GENERIC_TEST = 500;

  @Test
  public void edgeCase() {
    BytecodeRunner.of(BytecodeCompiler.newProgram().push(0xf0).push(0xf0).op(OpCode.AND).compile())
        .run();
  }

  @Test
  public void randomTest() {
    for (int i = 0; i < NB_GENERIC_TEST; i++) {
      BytecodeRunner.of(
              BytecodeCompiler.newProgram()
                  .push(bigIntegerToBytes(randBigInt()))
                  .push(bigIntegerToBytes(randBigInt()))
                  .op(randOpCode())
                  .compile())
          .run();
    }
  }

  private BigInteger randBigInt() {
    final int selector = RAND.nextInt(0, 5);

    return switch (selector) {
      case 0 -> BigInteger.ZERO;
      case 1 -> BigInteger.valueOf(RAND.nextInt(1, 32));
      case 2 -> BigInteger.valueOf(RAND.nextInt(32, 256));
      case 3 -> new BigInteger(16 * 8, RAND);
      case 4 -> new BigInteger(32 * 8, RAND);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }

  private OpCode randOpCode() {
    final int rand = RAND.nextInt(0, 6);
    return switch (rand) {
      case 0 -> OpCode.AND;
      case 1 -> OpCode.OR;
      case 2 -> OpCode.XOR;
      case 3 -> OpCode.NOT;
      case 4 -> OpCode.BYTE;
      case 5 -> OpCode.SIGNEXTEND;
      default -> throw new IllegalArgumentException("Unexpected value: " + rand);
    };
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
