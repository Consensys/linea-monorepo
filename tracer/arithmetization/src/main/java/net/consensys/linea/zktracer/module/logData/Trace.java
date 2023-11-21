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

package net.consensys.linea.zktracer.module.logData;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import org.apache.tuweni.units.bigints.UInt256;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {
  static final int CREATE2_SHIFT = 255;
  static final int G_TXDATA_NONZERO = 16;
  static final int G_TXDATA_ZERO = 4;
  static final int INT_LONG = 183;
  static final int INT_SHORT = 128;
  static final int LIST_LONG = 247;
  static final int LIST_SHORT = 192;
  static final int LLARGE = 16;
  static final int LLARGEMO = 15;
  static final int RLPADDR_CONST_RECIPE_1 = 1;
  static final int RLPADDR_CONST_RECIPE_2 = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_ADDR = 53;
  static final int RLPRECEIPT_SUBPHASE_ID_CUMUL_GAS = 3;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_LIMB = 77;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_SIZE = 83;
  static final int RLPRECEIPT_SUBPHASE_ID_NO_LOG_ENTRY = 11;
  static final int RLPRECEIPT_SUBPHASE_ID_STATUS_CODE = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE = 65;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA = 96;
  static final int RLPRECEIPT_SUBPHASE_ID_TYPE = 7;

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
        new ColumnHeader("logData.ABS_LOG_NUM", 32, length),
        new ColumnHeader("logData.ABS_LOG_NUM_MAX", 32, length),
        new ColumnHeader("logData.INDEX", 32, length),
        new ColumnHeader("logData.LIMB", 32, length),
        new ColumnHeader("logData.LOGS_DATA", 1, length),
        new ColumnHeader("logData.SIZE_ACC", 32, length),
        new ColumnHeader("logData.SIZE_LIMB", 32, length),
        new ColumnHeader("logData.SIZE_TOTAL", 32, length));
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

  public Trace absLogNum(final BigInteger b) {
    if (filled.get(0)) {
      throw new IllegalStateException("logData.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    absLogNum.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace absLogNumMax(final BigInteger b) {
    if (filled.get(1)) {
      throw new IllegalStateException("logData.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    absLogNumMax.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace index(final BigInteger b) {
    if (filled.get(2)) {
      throw new IllegalStateException("logData.INDEX already set");
    } else {
      filled.set(2);
    }

    index.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace limb(final BigInteger b) {
    if (filled.get(3)) {
      throw new IllegalStateException("logData.LIMB already set");
    } else {
      filled.set(3);
    }

    limb.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace logsData(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("logData.LOGS_DATA already set");
    } else {
      filled.set(4);
    }

    logsData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace sizeAcc(final BigInteger b) {
    if (filled.get(5)) {
      throw new IllegalStateException("logData.SIZE_ACC already set");
    } else {
      filled.set(5);
    }

    sizeAcc.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace sizeLimb(final BigInteger b) {
    if (filled.get(6)) {
      throw new IllegalStateException("logData.SIZE_LIMB already set");
    } else {
      filled.set(6);
    }

    sizeLimb.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace sizeTotal(final BigInteger b) {
    if (filled.get(7)) {
      throw new IllegalStateException("logData.SIZE_TOTAL already set");
    } else {
      filled.set(7);
    }

    sizeTotal.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("logData.ABS_LOG_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("logData.ABS_LOG_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("logData.INDEX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("logData.LIMB has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("logData.LOGS_DATA has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("logData.SIZE_ACC has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("logData.SIZE_LIMB has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("logData.SIZE_TOTAL has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absLogNum.position(absLogNum.position() + 32);
    }

    if (!filled.get(1)) {
      absLogNumMax.position(absLogNumMax.position() + 32);
    }

    if (!filled.get(2)) {
      index.position(index.position() + 32);
    }

    if (!filled.get(3)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(4)) {
      logsData.position(logsData.position() + 1);
    }

    if (!filled.get(5)) {
      sizeAcc.position(sizeAcc.position() + 32);
    }

    if (!filled.get(6)) {
      sizeLimb.position(sizeLimb.position() + 32);
    }

    if (!filled.get(7)) {
      sizeTotal.position(sizeTotal.position() + 32);
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
