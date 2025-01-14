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

package net.consensys.linea.zktracer.module.blake2fmodexpdata;

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
  public static final int INDEX_MAX_BLAKE_DATA = 0xc;
  public static final int INDEX_MAX_BLAKE_PARAMS = 0x1;
  public static final int INDEX_MAX_BLAKE_RESULT = 0x3;
  public static final int INDEX_MAX_MODEXP = 0x1f;
  public static final int INDEX_MAX_MODEXP_BASE = 0x1f;
  public static final int INDEX_MAX_MODEXP_EXPONENT = 0x1f;
  public static final int INDEX_MAX_MODEXP_MODULUS = 0x1f;
  public static final int INDEX_MAX_MODEXP_RESULT = 0x1f;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer id;
  private final MappedByteBuffer index;
  private final MappedByteBuffer indexMax;
  private final MappedByteBuffer isBlakeData;
  private final MappedByteBuffer isBlakeParams;
  private final MappedByteBuffer isBlakeResult;
  private final MappedByteBuffer isModexpBase;
  private final MappedByteBuffer isModexpExponent;
  private final MappedByteBuffer isModexpModulus;
  private final MappedByteBuffer isModexpResult;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer stamp;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("blake2fmodexpdata.ID", 4, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.INDEX", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.INDEX_MAX", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_BLAKE_DATA", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_BLAKE_PARAMS", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_BLAKE_RESULT", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_MODEXP_BASE", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_MODEXP_EXPONENT", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_MODEXP_MODULUS", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.IS_MODEXP_RESULT", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.LIMB", 16, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.PHASE", 1, length));
      headers.add(new ColumnHeader("blake2fmodexpdata.STAMP", 2, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.id = buffers.get(0);
    this.index = buffers.get(1);
    this.indexMax = buffers.get(2);
    this.isBlakeData = buffers.get(3);
    this.isBlakeParams = buffers.get(4);
    this.isBlakeResult = buffers.get(5);
    this.isModexpBase = buffers.get(6);
    this.isModexpExponent = buffers.get(7);
    this.isModexpModulus = buffers.get(8);
    this.isModexpResult = buffers.get(9);
    this.limb = buffers.get(10);
    this.phase = buffers.get(11);
    this.stamp = buffers.get(12);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace id(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("blake2fmodexpdata.ID already set");
    } else {
      filled.set(0);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("blake2fmodexpdata.ID has invalid value (" + b + ")"); }
    id.put((byte) (b >> 24));
    id.put((byte) (b >> 16));
    id.put((byte) (b >> 8));
    id.put((byte) b);


    return this;
  }

  public Trace index(final UnsignedByte b) {
    if (filled.get(1)) {
      throw new IllegalStateException("blake2fmodexpdata.INDEX already set");
    } else {
      filled.set(1);
    }

    index.put(b.toByte());

    return this;
  }

  public Trace indexMax(final UnsignedByte b) {
    if (filled.get(2)) {
      throw new IllegalStateException("blake2fmodexpdata.INDEX_MAX already set");
    } else {
      filled.set(2);
    }

    indexMax.put(b.toByte());

    return this;
  }

  public Trace isBlakeData(final Boolean b) {
    if (filled.get(3)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_BLAKE_DATA already set");
    } else {
      filled.set(3);
    }

    isBlakeData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isBlakeParams(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_BLAKE_PARAMS already set");
    } else {
      filled.set(4);
    }

    isBlakeParams.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isBlakeResult(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_BLAKE_RESULT already set");
    } else {
      filled.set(5);
    }

    isBlakeResult.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpBase(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_BASE already set");
    } else {
      filled.set(6);
    }

    isModexpBase.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpExponent(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_EXPONENT already set");
    } else {
      filled.set(7);
    }

    isModexpExponent.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpModulus(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_MODULUS already set");
    } else {
      filled.set(8);
    }

    isModexpModulus.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isModexpResult(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_RESULT already set");
    } else {
      filled.set(9);
    }

    isModexpResult.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("blake2fmodexpdata.LIMB already set");
    } else {
      filled.set(10);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("blake2fmodexpdata.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

    return this;
  }

  public Trace phase(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("blake2fmodexpdata.PHASE already set");
    } else {
      filled.set(11);
    }

    phase.put(b.toByte());

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(12)) {
      throw new IllegalStateException("blake2fmodexpdata.STAMP already set");
    } else {
      filled.set(12);
    }

    if(b >= 1024L) { throw new IllegalArgumentException("blake2fmodexpdata.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("blake2fmodexpdata.ID has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("blake2fmodexpdata.INDEX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("blake2fmodexpdata.INDEX_MAX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_BLAKE_DATA has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_BLAKE_PARAMS has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_BLAKE_RESULT has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_BASE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_EXPONENT has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_MODULUS has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("blake2fmodexpdata.IS_MODEXP_RESULT has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("blake2fmodexpdata.LIMB has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("blake2fmodexpdata.PHASE has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("blake2fmodexpdata.STAMP has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      id.position(id.position() + 4);
    }

    if (!filled.get(1)) {
      index.position(index.position() + 1);
    }

    if (!filled.get(2)) {
      indexMax.position(indexMax.position() + 1);
    }

    if (!filled.get(3)) {
      isBlakeData.position(isBlakeData.position() + 1);
    }

    if (!filled.get(4)) {
      isBlakeParams.position(isBlakeParams.position() + 1);
    }

    if (!filled.get(5)) {
      isBlakeResult.position(isBlakeResult.position() + 1);
    }

    if (!filled.get(6)) {
      isModexpBase.position(isModexpBase.position() + 1);
    }

    if (!filled.get(7)) {
      isModexpExponent.position(isModexpExponent.position() + 1);
    }

    if (!filled.get(8)) {
      isModexpModulus.position(isModexpModulus.position() + 1);
    }

    if (!filled.get(9)) {
      isModexpResult.position(isModexpResult.position() + 1);
    }

    if (!filled.get(10)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(11)) {
      phase.position(phase.position() + 1);
    }

    if (!filled.get(12)) {
      stamp.position(stamp.position() + 2);
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
