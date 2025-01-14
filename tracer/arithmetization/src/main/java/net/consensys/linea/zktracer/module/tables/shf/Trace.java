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

package net.consensys.linea.zktracer.module.tables.shf;

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

  private final MappedByteBuffer byte1;
  private final MappedByteBuffer iomf;
  private final MappedByteBuffer las;
  private final MappedByteBuffer mshp;
  private final MappedByteBuffer ones;
  private final MappedByteBuffer rap;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("shfreftable.BYTE1", 1, length));
      headers.add(new ColumnHeader("shfreftable.IOMF", 1, length));
      headers.add(new ColumnHeader("shfreftable.LAS", 1, length));
      headers.add(new ColumnHeader("shfreftable.MSHP", 1, length));
      headers.add(new ColumnHeader("shfreftable.ONES", 1, length));
      headers.add(new ColumnHeader("shfreftable.RAP", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.byte1 = buffers.get(0);
    this.iomf = buffers.get(1);
    this.las = buffers.get(2);
    this.mshp = buffers.get(3);
    this.ones = buffers.get(4);
    this.rap = buffers.get(5);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(0)) {
      throw new IllegalStateException("shfreftable.BYTE1 already set");
    } else {
      filled.set(0);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace iomf(final Boolean b) {
    if (filled.get(1)) {
      throw new IllegalStateException("shfreftable.IOMF already set");
    } else {
      filled.set(1);
    }

    iomf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace las(final UnsignedByte b) {
    if (filled.get(2)) {
      throw new IllegalStateException("shfreftable.LAS already set");
    } else {
      filled.set(2);
    }

    las.put(b.toByte());

    return this;
  }

  public Trace mshp(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("shfreftable.MSHP already set");
    } else {
      filled.set(3);
    }

    mshp.put(b.toByte());

    return this;
  }

  public Trace ones(final UnsignedByte b) {
    if (filled.get(4)) {
      throw new IllegalStateException("shfreftable.ONES already set");
    } else {
      filled.set(4);
    }

    ones.put(b.toByte());

    return this;
  }

  public Trace rap(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("shfreftable.RAP already set");
    } else {
      filled.set(5);
    }

    rap.put(b.toByte());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("shfreftable.BYTE1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("shfreftable.IOMF has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("shfreftable.LAS has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("shfreftable.MSHP has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("shfreftable.ONES has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("shfreftable.RAP has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(1)) {
      iomf.position(iomf.position() + 1);
    }

    if (!filled.get(2)) {
      las.position(las.position() + 1);
    }

    if (!filled.get(3)) {
      mshp.position(mshp.position() + 1);
    }

    if (!filled.get(4)) {
      ones.position(ones.position() + 1);
    }

    if (!filled.get(5)) {
      rap.position(rap.position() + 1);
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
