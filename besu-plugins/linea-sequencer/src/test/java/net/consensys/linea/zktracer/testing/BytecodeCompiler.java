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

package net.consensys.linea.zktracer.testing;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.platform.commons.util.Preconditions;

/** Fluent API for constructing custom sequences of EVM bytecode. */
public class BytecodeCompiler {
  private final List<Bytes> byteCode = new ArrayList<>();

  private BytecodeCompiler() {}

  /**
   * Create a new program instance that will contain a new bytecode sequence.
   *
   * @return an instance of {@link BytecodeCompiler}
   */
  public static BytecodeCompiler newProgram() {
    return new BytecodeCompiler();
  }

  private static Bytes toBytes(final int x) {
    return Bytes.ofUnsignedShort(x).trimLeadingZeros();
  }

  /**
   * Add an {@link OpCode} to the bytecode sequence.
   *
   * @param opCode opcode to be added
   * @return current instance
   */
  public BytecodeCompiler op(final OpCode opCode) {
    byteCode.add(Bytes.of(opCode.byteValue()));

    return this;
  }

  /**
   * Add an opcode and a list of {@link Bytes32} opcode arguments to the bytecode sequence.
   *
   * @param opCode opcode to be added
   * @param arguments list of arguments related to the opcode to be added
   * @return current instance
   */
  public BytecodeCompiler opAnd32ByteArgs(final OpCode opCode, final List<Bytes32> arguments) {
    for (Bytes32 argument : arguments) {
      push(argument);
    }

    op(opCode);

    return this;
  }

  /**
   * Add a byte array as is to the bytecode sequence.
   *
   * @param bs byte array to be added
   * @return current instance
   */
  public BytecodeCompiler immediate(final byte[] bs) {
    for (byte b : bs) {
      this.byteCode.add(Bytes.of(b));
    }

    return this;
  }

  /**
   * Add a {@link Bytes} instance as is to the bytecode sequence.
   *
   * @param bytes {@link Bytes} to be added
   * @return current instance
   */
  public BytecodeCompiler immediate(final Bytes bytes) {
    return this.immediate(bytes.toArray());
  }

  /**
   * Add an int as is to the bytecode sequence.
   *
   * @param x integer number to be added
   * @return current instance
   */
  public BytecodeCompiler immediate(final int x) {
    return this.immediate(toBytes(x));
  }

  /**
   * Add a {@link UInt256} number as is to the bytecode sequence.
   *
   * @param x {@link UInt256} number to be added
   * @return current instance
   */
  public BytecodeCompiler immediate(final UInt256 x) {
    return this.immediate(x.toArray());
  }

  /**
   * Add {@link OpCode#PUSH1} and byte array arguments.
   *
   * @param xs byte array arguments
   * @return current instance
   */
  public BytecodeCompiler push(final Bytes xs) {
    Preconditions.condition(
        !xs.isEmpty() && xs.size() <= 32, "Provided byte array is empty or exceeds 32 bytes");

    int pushNOpCode = OpCode.PUSH1.byteValue() + xs.size() - 1;

    return this.immediate(pushNOpCode).immediate(xs);
  }

  /**
   * Add {@link OpCode#PUSH1} and int number argument.
   *
   * @param x int number argument
   * @return current instance
   */
  public BytecodeCompiler push(final int x) {
    return this.push(toBytes(x));
  }

  /**
   * Compile bytecode sequence to a single {@link Bytes} instance.
   *
   * @return a {@link Bytes} instance containing a pre-defined sequence of bytes and {@link OpCode}s
   */
  public Bytes compile() {
    return Bytes.concatenate(byteCode);
  }
}
