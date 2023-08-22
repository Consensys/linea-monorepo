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

package net.consensys.linea.zktracer.testutils;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;

public class BytecodeCompiler {
  private final List<Byte> byteCode = new ArrayList<>();

  private static Bytes toBytes(int x) {
    return Bytes.ofUnsignedShort(x).trimLeadingZeros();
  }

  public BytecodeCompiler op(OpCode opCode) {
    this.byteCode.add(opCode.byteValue());
    return this;
  }

  public BytecodeCompiler immediate(byte[] bs) {
    for (Byte b : bs) {
      this.byteCode.add(b);
    }
    return this;
  }

  public BytecodeCompiler immediate(Bytes bytes) {
    return this.immediate(bytes.toArray());
  }

  public BytecodeCompiler immediate(int x) {
    return this.immediate(toBytes(x));
  }

  public BytecodeCompiler immediate(UInt256 x) {
    return this.immediate(x.toArray());
  }

  public BytecodeCompiler push(Bytes xs) {
    assert xs.size() > 0 && xs.size() <= 32;
    return this.immediate(OpCode.PUSH1.byteValue() + xs.size() - 1).immediate(xs);
  }

  public BytecodeCompiler push(byte[] xs) {
    assert xs.length > 0 && xs.length <= 32;
    return this.immediate(OpCode.PUSH1.byteValue() + xs.length - 1).immediate(xs);
  }

  public BytecodeCompiler push(int x) {
    return this.push(toBytes(x));
  }

  public Bytes compile() {
    byte[] ret = new byte[this.byteCode.size()];
    for (int i = 0; i < ret.length; i++) {
      ret[i] = this.byteCode.get(i);
    }

    return Bytes.wrap(ret);
  }
}
