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

package net.consensys.linea.zktracer.module.tables.bin;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer inputByte1;
  private final MappedByteBuffer inputByte2;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer resultByte;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("binreftable.INPUT_BYTE_1", 1, length));
      headers.add(new ColumnHeader("binreftable.INPUT_BYTE_2", 1, length));
      headers.add(new ColumnHeader("binreftable.INST", 1, length));
      headers.add(new ColumnHeader("binreftable.RESULT_BYTE", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.inputByte1 = buffers.get(0);
    this.inputByte2 = buffers.get(1);
    this.inst = buffers.get(2);
    this.resultByte = buffers.get(3);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace inputByte1(final UnsignedByte b) {
    if (filled.get(0)) {
      throw new IllegalStateException("binreftable.INPUT_BYTE_1 already set");
    } else {
      filled.set(0);
    }

    inputByte1.put(b.toByte());

    return this;
  }

  public Trace inputByte2(final UnsignedByte b) {
    if (filled.get(1)) {
      throw new IllegalStateException("binreftable.INPUT_BYTE_2 already set");
    } else {
      filled.set(1);
    }

    inputByte2.put(b.toByte());

    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(2)) {
      throw new IllegalStateException("binreftable.INST already set");
    } else {
      filled.set(2);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace resultByte(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("binreftable.RESULT_BYTE already set");
    } else {
      filled.set(3);
    }

    resultByte.put(b.toByte());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("binreftable.INPUT_BYTE_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("binreftable.INPUT_BYTE_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("binreftable.INST has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("binreftable.RESULT_BYTE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      inputByte1.position(inputByte1.position() + 1);
    }

    if (!filled.get(1)) {
      inputByte2.position(inputByte2.position() + 1);
    }

    if (!filled.get(2)) {
      inst.position(inst.position() + 1);
    }

    if (!filled.get(3)) {
      resultByte.position(resultByte.position() + 1);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public void build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
  }
}
