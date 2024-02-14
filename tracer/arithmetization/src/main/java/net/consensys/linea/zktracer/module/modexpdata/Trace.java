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

package net.consensys.linea.zktracer.module.modexpdata;

import java.nio.MappedByteBuffer;
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
  public static final int PHASE_BASE = 0x1;
  public static final int PHASE_EXPONENT = 0x2;
  public static final int PHASE_MODULUS = 0x3;
  public static final int PHASE_RESULT = 0x4;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer bemr;
  private final MappedByteBuffer bytes;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer index;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer resultDataContext;
  private final MappedByteBuffer stamp;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("modexpdata.BEMR", 1, length),
        new ColumnHeader("modexpdata.BYTES", 1, length),
        new ColumnHeader("modexpdata.CT", 1, length),
        new ColumnHeader("modexpdata.INDEX", 1, length),
        new ColumnHeader("modexpdata.LIMB", 32, length),
        new ColumnHeader("modexpdata.PHASE", 1, length),
        new ColumnHeader("modexpdata.RESULT_DATA_CONTEXT", 32, length),
        new ColumnHeader("modexpdata.STAMP", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.bemr = buffers.get(0);
    this.bytes = buffers.get(1);
    this.ct = buffers.get(2);
    this.index = buffers.get(3);
    this.limb = buffers.get(4);
    this.phase = buffers.get(5);
    this.resultDataContext = buffers.get(6);
    this.stamp = buffers.get(7);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace bemr(final UnsignedByte b) {
    if (filled.get(0)) {
      throw new IllegalStateException("modexpdata.BEMR already set");
    } else {
      filled.set(0);
    }

    bemr.put(b.toByte());

    return this;
  }

  public Trace bytes(final UnsignedByte b) {
    if (filled.get(1)) {
      throw new IllegalStateException("modexpdata.BYTES already set");
    } else {
      filled.set(1);
    }

    bytes.put(b.toByte());

    return this;
  }

  public Trace ct(final UnsignedByte b) {
    if (filled.get(2)) {
      throw new IllegalStateException("modexpdata.CT already set");
    } else {
      filled.set(2);
    }

    ct.put(b.toByte());

    return this;
  }

  public Trace index(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("modexpdata.INDEX already set");
    } else {
      filled.set(3);
    }

    index.put(b.toByte());

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("modexpdata.LIMB already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace phase(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("modexpdata.PHASE already set");
    } else {
      filled.set(5);
    }

    phase.put(b.toByte());

    return this;
  }

  public Trace resultDataContext(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("modexpdata.RESULT_DATA_CONTEXT already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      resultDataContext.put((byte) 0);
    }
    resultDataContext.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("modexpdata.STAMP already set");
    } else {
      filled.set(7);
    }

    stamp.put(b.toByte());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("modexpdata.BEMR has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("modexpdata.BYTES has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("modexpdata.CT has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("modexpdata.INDEX has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("modexpdata.LIMB has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("modexpdata.PHASE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("modexpdata.RESULT_DATA_CONTEXT has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("modexpdata.STAMP has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      bemr.position(bemr.position() + 1);
    }

    if (!filled.get(1)) {
      bytes.position(bytes.position() + 1);
    }

    if (!filled.get(2)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(3)) {
      index.position(index.position() + 1);
    }

    if (!filled.get(4)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(5)) {
      phase.position(phase.position() + 1);
    }

    if (!filled.get(6)) {
      resultDataContext.position(resultDataContext.position() + 32);
    }

    if (!filled.get(7)) {
      stamp.position(stamp.position() + 1);
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
