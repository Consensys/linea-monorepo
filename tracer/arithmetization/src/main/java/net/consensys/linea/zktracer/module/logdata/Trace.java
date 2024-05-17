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

package net.consensys.linea.zktracer.module.logdata;

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

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absLogNum;
  private final MappedByteBuffer absLogNumMax;
  private final MappedByteBuffer index;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer logsData;
  private final MappedByteBuffer sizeAcc;
  private final MappedByteBuffer sizeLimb;
  private final MappedByteBuffer sizeTotal;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("logdata.ABS_LOG_NUM", 4, length),
        new ColumnHeader("logdata.ABS_LOG_NUM_MAX", 4, length),
        new ColumnHeader("logdata.INDEX", 4, length),
        new ColumnHeader("logdata.LIMB", 32, length),
        new ColumnHeader("logdata.LOGS_DATA", 1, length),
        new ColumnHeader("logdata.SIZE_ACC", 8, length),
        new ColumnHeader("logdata.SIZE_LIMB", 1, length),
        new ColumnHeader("logdata.SIZE_TOTAL", 8, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absLogNum = buffers.get(0);
    this.absLogNumMax = buffers.get(1);
    this.index = buffers.get(2);
    this.limb = buffers.get(3);
    this.logsData = buffers.get(4);
    this.sizeAcc = buffers.get(5);
    this.sizeLimb = buffers.get(6);
    this.sizeTotal = buffers.get(7);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absLogNum(final int b) {
    if (filled.get(0)) {
      throw new IllegalStateException("logdata.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    absLogNum.putInt(b);

    return this;
  }

  public Trace absLogNumMax(final int b) {
    if (filled.get(1)) {
      throw new IllegalStateException("logdata.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    absLogNumMax.putInt(b);

    return this;
  }

  public Trace index(final int b) {
    if (filled.get(2)) {
      throw new IllegalStateException("logdata.INDEX already set");
    } else {
      filled.set(2);
    }

    index.putInt(b);

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("logdata.LIMB already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace logsData(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("logdata.LOGS_DATA already set");
    } else {
      filled.set(4);
    }

    logsData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace sizeAcc(final long b) {
    if (filled.get(5)) {
      throw new IllegalStateException("logdata.SIZE_ACC already set");
    } else {
      filled.set(5);
    }

    sizeAcc.putLong(b);

    return this;
  }

  public Trace sizeLimb(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("logdata.SIZE_LIMB already set");
    } else {
      filled.set(6);
    }

    sizeLimb.put(b.toByte());

    return this;
  }

  public Trace sizeTotal(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("logdata.SIZE_TOTAL already set");
    } else {
      filled.set(7);
    }

    sizeTotal.putLong(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("logdata.ABS_LOG_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("logdata.ABS_LOG_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("logdata.INDEX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("logdata.LIMB has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("logdata.LOGS_DATA has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("logdata.SIZE_ACC has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("logdata.SIZE_LIMB has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("logdata.SIZE_TOTAL has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absLogNum.position(absLogNum.position() + 4);
    }

    if (!filled.get(1)) {
      absLogNumMax.position(absLogNumMax.position() + 4);
    }

    if (!filled.get(2)) {
      index.position(index.position() + 4);
    }

    if (!filled.get(3)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(4)) {
      logsData.position(logsData.position() + 1);
    }

    if (!filled.get(5)) {
      sizeAcc.position(sizeAcc.position() + 8);
    }

    if (!filled.get(6)) {
      sizeLimb.position(sizeLimb.position() + 1);
    }

    if (!filled.get(7)) {
      sizeTotal.position(sizeTotal.position() + 8);
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
