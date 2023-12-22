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

package net.consensys.linea.zktracer.testing;

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.StringReader;
import java.math.BigInteger;
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
    return Bytes.ofUnsignedLong(x).trimLeadingZeros();
  }

  /**
   * Assemble an EVM program into bytecode. Requirement are: one instruction per line; `;` mark a
   * line as a comment; no more than one word per line, safe for PUSHs.
   *
   * @param program
   * @return the bytecode corresponding to the assembly code
   */
  public BytecodeCompiler assemble(final String program) {
    BufferedReader bufReader = new BufferedReader(new StringReader(program));
    String line;
    while (true) {
      try {
        if ((line = bufReader.readLine()) == null) break;
        if (line.isBlank()) continue;
        if (line.startsWith(";")) continue;

        String[] ls = line.split("\\s+");
        final OpCode opCode = OpCode.fromMnemonic(ls[0]);
        this.op(opCode);
        final int pushSize = opCode.byteValue() - (int) OpCode.PUSH1.byteValue() + 1;
        final boolean isPush = pushSize >= 1 && pushSize <= 32;
        if (isPush) {
          this.immediate(Bytes.fromHexString(ls[1], pushSize));
        } else {
          if (ls.length > 1) {
            throw new IllegalArgumentException("expected nothing, found" + ls[1]);
          }
        }
      } catch (IOException e) {
        throw new RuntimeException(e);
      }
    }
    return this;
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
    this.byteCode.add(Bytes.wrap(bs));

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
    this.byteCode.add(x);

    return this;
  }

  /**
   * Add a {@link OpCode#PUSH1} and byte array arguments.
   *
   * @param xs byte array arguments
   * @return current instance
   */
  public BytecodeCompiler push(final Bytes xs) {
    Preconditions.condition(xs.size() <= 32, "Provided byte array is empty or exceeds 32 bytes");

    if (xs.isEmpty()) {
      return this.immediate(OpCode.PUSH1.byteValue()).immediate(Bytes.of(0));
    } else {
      int pushNOpCode = OpCode.PUSH1.byteValue() + xs.size() - 1;
      return this.immediate(pushNOpCode).immediate(xs);
    }
  }

  /**
   * Add a {@link OpCode#PUSH1} and a {@link BigInteger} argument.
   *
   * @param xs BigInteger argument
   * @return current instance
   */
  public BytecodeCompiler push(final BigInteger xs) {
    return this.push(bigIntegerToBytes(xs));
  }

  /**
   * Add a {@link OpCode#PUSH1} and int number argument.
   *
   * @param x int number argument
   * @return current instance
   */
  public BytecodeCompiler push(final int x) {
    return this.push(toBytes(x));
  }

  /**
   * Add a PUSH operation of the given width and its (padded) argument
   *
   * @param w the width to push (in [[1; 32]])
   * @param bs byte array to be added
   * @return current instance
   */
  public BytecodeCompiler push(final int w, final byte[] bs) {
    Preconditions.condition(w > 0 && w <= 32, "Invalid PUSH width");
    Preconditions.condition(bs.length <= w, "PUSH argument must be smaller than the width");

    final int padding = w - bs.length;

    this.op(OpCode.of(0x5f + w));
    this.byteCode.add(Bytes.repeat((byte) 0, padding));
    this.byteCode.add(Bytes.of(bs));

    return this;
  }

  /**
   * Add a PUSH operation of the given width and its (padded) argument
   *
   * @param w the width to push (in [[1; 32]])
   * @param bytes {@link Bytes} to be added
   * @return current instance
   */
  public BytecodeCompiler push(final int w, final Bytes bytes) {

    return this.push(w, bytes.toArray());
  }

  /**
   * Add an int as is to the bytecode sequence.
   *
   * @param w the width to push (in [[1; 32]])
   * @param x integer number to be added
   * @return current instance
   */
  public BytecodeCompiler push(final int w, final int x) {
    return this.push(w, toBytes(x));
  }

  /**
   * Add a PUSH operation of the given width and its (padded) argument
   *
   * @param w the width to push (in [[1; 32]])
   * @param x {@link UInt256} number to be added
   * @return current instance
   */
  public BytecodeCompiler push(final int w, final UInt256 x) {
    return this.push(w, x.toArray());
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
