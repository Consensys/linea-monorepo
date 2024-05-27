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

package net.consensys.linea.zktracer.module.shakiradata;

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
  public static final int INDEX_MAX_RESULT = 0x1;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer id;
  private final MappedByteBuffer index;
  private final MappedByteBuffer indexMax;
  private final MappedByteBuffer isKeccakData;
  private final MappedByteBuffer isKeccakResult;
  private final MappedByteBuffer isRipemdData;
  private final MappedByteBuffer isRipemdResult;
  private final MappedByteBuffer isSha2Data;
  private final MappedByteBuffer isSha2Result;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nBytesAcc;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer ripshaStamp;
  private final MappedByteBuffer selectorKeccakResHi;
  private final MappedByteBuffer selectorRipemdResHi;
  private final MappedByteBuffer selectorSha2ResHi;
  private final MappedByteBuffer totalSize;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("shakiradata.ID", 8, length),
        new ColumnHeader("shakiradata.INDEX", 8, length),
        new ColumnHeader("shakiradata.INDEX_MAX", 8, length),
        new ColumnHeader("shakiradata.IS_KECCAK_DATA", 1, length),
        new ColumnHeader("shakiradata.IS_KECCAK_RESULT", 1, length),
        new ColumnHeader("shakiradata.IS_RIPEMD_DATA", 1, length),
        new ColumnHeader("shakiradata.IS_RIPEMD_RESULT", 1, length),
        new ColumnHeader("shakiradata.IS_SHA2_DATA", 1, length),
        new ColumnHeader("shakiradata.IS_SHA2_RESULT", 1, length),
        new ColumnHeader("shakiradata.LIMB", 32, length),
        new ColumnHeader("shakiradata.nBYTES", 2, length),
        new ColumnHeader("shakiradata.nBYTES_ACC", 8, length),
        new ColumnHeader("shakiradata.PHASE", 1, length),
        new ColumnHeader("shakiradata.RIPSHA_STAMP", 8, length),
        new ColumnHeader("shakiradata.SELECTOR_KECCAK_RES_HI", 1, length),
        new ColumnHeader("shakiradata.SELECTOR_RIPEMD_RES_HI", 1, length),
        new ColumnHeader("shakiradata.SELECTOR_SHA2_RES_HI", 1, length),
        new ColumnHeader("shakiradata.TOTAL_SIZE", 8, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.id = buffers.get(0);
    this.index = buffers.get(1);
    this.indexMax = buffers.get(2);
    this.isKeccakData = buffers.get(3);
    this.isKeccakResult = buffers.get(4);
    this.isRipemdData = buffers.get(5);
    this.isRipemdResult = buffers.get(6);
    this.isSha2Data = buffers.get(7);
    this.isSha2Result = buffers.get(8);
    this.limb = buffers.get(9);
    this.nBytes = buffers.get(10);
    this.nBytesAcc = buffers.get(11);
    this.phase = buffers.get(12);
    this.ripshaStamp = buffers.get(13);
    this.selectorKeccakResHi = buffers.get(14);
    this.selectorRipemdResHi = buffers.get(15);
    this.selectorSha2ResHi = buffers.get(16);
    this.totalSize = buffers.get(17);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace id(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("shakiradata.ID already set");
    } else {
      filled.set(0);
    }

    id.putLong(b);

    return this;
  }

  public Trace index(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("shakiradata.INDEX already set");
    } else {
      filled.set(1);
    }

    index.putLong(b);

    return this;
  }

  public Trace indexMax(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("shakiradata.INDEX_MAX already set");
    } else {
      filled.set(2);
    }

    indexMax.putLong(b);

    return this;
  }

  public Trace isKeccakData(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("shakiradata.IS_KECCAK_DATA already set");
    } else {
      filled.set(3);
    }

    isKeccakData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isKeccakResult(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("shakiradata.IS_KECCAK_RESULT already set");
    } else {
      filled.set(4);
    }

    isKeccakResult.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRipemdData(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("shakiradata.IS_RIPEMD_DATA already set");
    } else {
      filled.set(5);
    }

    isRipemdData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRipemdResult(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("shakiradata.IS_RIPEMD_RESULT already set");
    } else {
      filled.set(6);
    }

    isRipemdResult.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSha2Data(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("shakiradata.IS_SHA2_DATA already set");
    } else {
      filled.set(7);
    }

    isSha2Data.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isSha2Result(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("shakiradata.IS_SHA2_RESULT already set");
    } else {
      filled.set(8);
    }

    isSha2Result.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("shakiradata.LIMB already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nBytes(final short b) {
    if (filled.get(16)) {
      throw new IllegalStateException("shakiradata.nBYTES already set");
    } else {
      filled.set(16);
    }

    nBytes.putShort(b);

    return this;
  }

  public Trace nBytesAcc(final long b) {
    if (filled.get(17)) {
      throw new IllegalStateException("shakiradata.nBYTES_ACC already set");
    } else {
      filled.set(17);
    }

    nBytesAcc.putLong(b);

    return this;
  }

  public Trace phase(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("shakiradata.PHASE already set");
    } else {
      filled.set(10);
    }

    phase.put(b.toByte());

    return this;
  }

  public Trace ripshaStamp(final long b) {
    if (filled.get(11)) {
      throw new IllegalStateException("shakiradata.RIPSHA_STAMP already set");
    } else {
      filled.set(11);
    }

    ripshaStamp.putLong(b);

    return this;
  }

  public Trace selectorKeccakResHi(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("shakiradata.SELECTOR_KECCAK_RES_HI already set");
    } else {
      filled.set(12);
    }

    selectorKeccakResHi.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace selectorRipemdResHi(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("shakiradata.SELECTOR_RIPEMD_RES_HI already set");
    } else {
      filled.set(13);
    }

    selectorRipemdResHi.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace selectorSha2ResHi(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("shakiradata.SELECTOR_SHA2_RES_HI already set");
    } else {
      filled.set(14);
    }

    selectorSha2ResHi.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace totalSize(final long b) {
    if (filled.get(15)) {
      throw new IllegalStateException("shakiradata.TOTAL_SIZE already set");
    } else {
      filled.set(15);
    }

    totalSize.putLong(b);

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("shakiradata.ID has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("shakiradata.INDEX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("shakiradata.INDEX_MAX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("shakiradata.IS_KECCAK_DATA has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("shakiradata.IS_KECCAK_RESULT has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("shakiradata.IS_RIPEMD_DATA has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("shakiradata.IS_RIPEMD_RESULT has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("shakiradata.IS_SHA2_DATA has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("shakiradata.IS_SHA2_RESULT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("shakiradata.LIMB has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("shakiradata.nBYTES has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("shakiradata.nBYTES_ACC has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("shakiradata.PHASE has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("shakiradata.RIPSHA_STAMP has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("shakiradata.SELECTOR_KECCAK_RES_HI has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("shakiradata.SELECTOR_RIPEMD_RES_HI has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("shakiradata.SELECTOR_SHA2_RES_HI has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("shakiradata.TOTAL_SIZE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      id.position(id.position() + 8);
    }

    if (!filled.get(1)) {
      index.position(index.position() + 8);
    }

    if (!filled.get(2)) {
      indexMax.position(indexMax.position() + 8);
    }

    if (!filled.get(3)) {
      isKeccakData.position(isKeccakData.position() + 1);
    }

    if (!filled.get(4)) {
      isKeccakResult.position(isKeccakResult.position() + 1);
    }

    if (!filled.get(5)) {
      isRipemdData.position(isRipemdData.position() + 1);
    }

    if (!filled.get(6)) {
      isRipemdResult.position(isRipemdResult.position() + 1);
    }

    if (!filled.get(7)) {
      isSha2Data.position(isSha2Data.position() + 1);
    }

    if (!filled.get(8)) {
      isSha2Result.position(isSha2Result.position() + 1);
    }

    if (!filled.get(9)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(16)) {
      nBytes.position(nBytes.position() + 2);
    }

    if (!filled.get(17)) {
      nBytesAcc.position(nBytesAcc.position() + 8);
    }

    if (!filled.get(10)) {
      phase.position(phase.position() + 1);
    }

    if (!filled.get(11)) {
      ripshaStamp.position(ripshaStamp.position() + 8);
    }

    if (!filled.get(12)) {
      selectorKeccakResHi.position(selectorKeccakResHi.position() + 1);
    }

    if (!filled.get(13)) {
      selectorRipemdResHi.position(selectorRipemdResHi.position() + 1);
    }

    if (!filled.get(14)) {
      selectorSha2ResHi.position(selectorSha2ResHi.position() + 1);
    }

    if (!filled.get(15)) {
      totalSize.position(totalSize.position() + 8);
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
