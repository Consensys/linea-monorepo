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

  private final MappedByteBuffer absLogNum;
  private final MappedByteBuffer absLogNumMax;
  private final MappedByteBuffer index;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer logsData;
  private final MappedByteBuffer sizeAcc;
  private final MappedByteBuffer sizeLimb;
  private final MappedByteBuffer sizeTotal;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("logdata.ABS_LOG_NUM", 3, length));
      headers.add(new ColumnHeader("logdata.ABS_LOG_NUM_MAX", 3, length));
      headers.add(new ColumnHeader("logdata.INDEX", 3, length));
      headers.add(new ColumnHeader("logdata.LIMB", 16, length));
      headers.add(new ColumnHeader("logdata.LOGS_DATA", 1, length));
      headers.add(new ColumnHeader("logdata.SIZE_ACC", 4, length));
      headers.add(new ColumnHeader("logdata.SIZE_LIMB", 1, length));
      headers.add(new ColumnHeader("logdata.SIZE_TOTAL", 4, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
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

  public Trace absLogNum(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("logdata.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("logdata.ABS_LOG_NUM has invalid value (" + b + ")"); }
    absLogNum.put((byte) (b >> 16));
    absLogNum.put((byte) (b >> 8));
    absLogNum.put((byte) b);


    return this;
  }

  public Trace absLogNumMax(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("logdata.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("logdata.ABS_LOG_NUM_MAX has invalid value (" + b + ")"); }
    absLogNumMax.put((byte) (b >> 16));
    absLogNumMax.put((byte) (b >> 8));
    absLogNumMax.put((byte) b);


    return this;
  }

  public Trace index(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("logdata.INDEX already set");
    } else {
      filled.set(2);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("logdata.INDEX has invalid value (" + b + ")"); }
    index.put((byte) (b >> 16));
    index.put((byte) (b >> 8));
    index.put((byte) b);


    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("logdata.LIMB already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("logdata.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

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

    if(b >= 4294967296L) { throw new IllegalArgumentException("logdata.SIZE_ACC has invalid value (" + b + ")"); }
    sizeAcc.put((byte) (b >> 24));
    sizeAcc.put((byte) (b >> 16));
    sizeAcc.put((byte) (b >> 8));
    sizeAcc.put((byte) b);


    return this;
  }

  public Trace sizeLimb(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("logdata.SIZE_LIMB already set");
    } else {
      filled.set(6);
    }

    if(b >= 32L) { throw new IllegalArgumentException("logdata.SIZE_LIMB has invalid value (" + b + ")"); }
    sizeLimb.put((byte) b);


    return this;
  }

  public Trace sizeTotal(final long b) {
    if (filled.get(7)) {
      throw new IllegalStateException("logdata.SIZE_TOTAL already set");
    } else {
      filled.set(7);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("logdata.SIZE_TOTAL has invalid value (" + b + ")"); }
    sizeTotal.put((byte) (b >> 24));
    sizeTotal.put((byte) (b >> 16));
    sizeTotal.put((byte) (b >> 8));
    sizeTotal.put((byte) b);


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
      absLogNum.position(absLogNum.position() + 3);
    }

    if (!filled.get(1)) {
      absLogNumMax.position(absLogNumMax.position() + 3);
    }

    if (!filled.get(2)) {
      index.position(index.position() + 3);
    }

    if (!filled.get(3)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(4)) {
      logsData.position(logsData.position() + 1);
    }

    if (!filled.get(5)) {
      sizeAcc.position(sizeAcc.position() + 4);
    }

    if (!filled.get(6)) {
      sizeLimb.position(sizeLimb.position() + 1);
    }

    if (!filled.get(7)) {
      sizeTotal.position(sizeTotal.position() + 4);
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
