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

package net.consensys.linea.zktracer.module.tables.bin;

import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
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

  private final MappedByteBuffer andByte;
  private final MappedByteBuffer byteArg1;
  private final MappedByteBuffer byteArg2;
  private final MappedByteBuffer isInRt;
  private final MappedByteBuffer notByte;
  private final MappedByteBuffer orByte;
  private final MappedByteBuffer xorByte;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("binRT.AND_BYTE", 32, length),
        new ColumnHeader("binRT.BYTE_ARG_1", 32, length),
        new ColumnHeader("binRT.BYTE_ARG_2", 32, length),
        new ColumnHeader("binRT.IS_IN_RT", 32, length),
        new ColumnHeader("binRT.NOT_BYTE", 32, length),
        new ColumnHeader("binRT.OR_BYTE", 32, length),
        new ColumnHeader("binRT.XOR_BYTE", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.andByte = buffers.get(0);
    this.byteArg1 = buffers.get(1);
    this.byteArg2 = buffers.get(2);
    this.isInRt = buffers.get(3);
    this.notByte = buffers.get(4);
    this.orByte = buffers.get(5);
    this.xorByte = buffers.get(6);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace andByte(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("binRT.AND_BYTE already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      andByte.put((byte) 0);
    }
    andByte.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byteArg1(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("binRT.BYTE_ARG_1 already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      byteArg1.put((byte) 0);
    }
    byteArg1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byteArg2(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("binRT.BYTE_ARG_2 already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      byteArg2.put((byte) 0);
    }
    byteArg2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isInRt(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("binRT.IS_IN_RT already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      isInRt.put((byte) 0);
    }
    isInRt.put(b.toArrayUnsafe());

    return this;
  }

  public Trace notByte(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("binRT.NOT_BYTE already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      notByte.put((byte) 0);
    }
    notByte.put(b.toArrayUnsafe());

    return this;
  }

  public Trace orByte(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("binRT.OR_BYTE already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      orByte.put((byte) 0);
    }
    orByte.put(b.toArrayUnsafe());

    return this;
  }

  public Trace xorByte(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("binRT.XOR_BYTE already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      xorByte.put((byte) 0);
    }
    xorByte.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("binRT.AND_BYTE has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("binRT.BYTE_ARG_1 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("binRT.BYTE_ARG_2 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("binRT.IS_IN_RT has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("binRT.NOT_BYTE has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("binRT.OR_BYTE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("binRT.XOR_BYTE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      andByte.position(andByte.position() + 32);
    }

    if (!filled.get(1)) {
      byteArg1.position(byteArg1.position() + 32);
    }

    if (!filled.get(2)) {
      byteArg2.position(byteArg2.position() + 32);
    }

    if (!filled.get(3)) {
      isInRt.position(isInRt.position() + 32);
    }

    if (!filled.get(4)) {
      notByte.position(notByte.position() + 32);
    }

    if (!filled.get(5)) {
      orByte.position(orByte.position() + 32);
    }

    if (!filled.get(6)) {
      xorByte.position(xorByte.position() + 32);
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
