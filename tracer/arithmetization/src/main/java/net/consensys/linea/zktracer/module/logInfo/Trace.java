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

package net.consensys.linea.zktracer.module.logInfo;

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
  private final MappedByteBuffer absTxnNum;
  private final MappedByteBuffer absTxnNumMax;
  private final MappedByteBuffer addrHi;
  private final MappedByteBuffer addrLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer ctMax;
  private final MappedByteBuffer dataHi;
  private final MappedByteBuffer dataLo;
  private final MappedByteBuffer dataSize;
  private final MappedByteBuffer inst;
  private final MappedByteBuffer isLogX0;
  private final MappedByteBuffer isLogX1;
  private final MappedByteBuffer isLogX2;
  private final MappedByteBuffer isLogX3;
  private final MappedByteBuffer isLogX4;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer topicHi1;
  private final MappedByteBuffer topicHi2;
  private final MappedByteBuffer topicHi3;
  private final MappedByteBuffer topicHi4;
  private final MappedByteBuffer topicLo1;
  private final MappedByteBuffer topicLo2;
  private final MappedByteBuffer topicLo3;
  private final MappedByteBuffer topicLo4;
  private final MappedByteBuffer txnEmitsLogs;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("logInfo.ABS_LOG_NUM", 32, length),
        new ColumnHeader("logInfo.ABS_LOG_NUM_MAX", 32, length),
        new ColumnHeader("logInfo.ABS_TXN_NUM", 32, length),
        new ColumnHeader("logInfo.ABS_TXN_NUM_MAX", 32, length),
        new ColumnHeader("logInfo.ADDR_HI", 32, length),
        new ColumnHeader("logInfo.ADDR_LO", 32, length),
        new ColumnHeader("logInfo.CT", 32, length),
        new ColumnHeader("logInfo.CT_MAX", 32, length),
        new ColumnHeader("logInfo.DATA_HI", 32, length),
        new ColumnHeader("logInfo.DATA_LO", 32, length),
        new ColumnHeader("logInfo.DATA_SIZE", 32, length),
        new ColumnHeader("logInfo.INST", 32, length),
        new ColumnHeader("logInfo.IS_LOG_X_0", 1, length),
        new ColumnHeader("logInfo.IS_LOG_X_1", 1, length),
        new ColumnHeader("logInfo.IS_LOG_X_2", 1, length),
        new ColumnHeader("logInfo.IS_LOG_X_3", 1, length),
        new ColumnHeader("logInfo.IS_LOG_X_4", 1, length),
        new ColumnHeader("logInfo.PHASE", 32, length),
        new ColumnHeader("logInfo.TOPIC_HI_1", 32, length),
        new ColumnHeader("logInfo.TOPIC_HI_2", 32, length),
        new ColumnHeader("logInfo.TOPIC_HI_3", 32, length),
        new ColumnHeader("logInfo.TOPIC_HI_4", 32, length),
        new ColumnHeader("logInfo.TOPIC_LO_1", 32, length),
        new ColumnHeader("logInfo.TOPIC_LO_2", 32, length),
        new ColumnHeader("logInfo.TOPIC_LO_3", 32, length),
        new ColumnHeader("logInfo.TOPIC_LO_4", 32, length),
        new ColumnHeader("logInfo.TXN_EMITS_LOGS", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absLogNum = buffers.get(0);
    this.absLogNumMax = buffers.get(1);
    this.absTxnNum = buffers.get(2);
    this.absTxnNumMax = buffers.get(3);
    this.addrHi = buffers.get(4);
    this.addrLo = buffers.get(5);
    this.ct = buffers.get(6);
    this.ctMax = buffers.get(7);
    this.dataHi = buffers.get(8);
    this.dataLo = buffers.get(9);
    this.dataSize = buffers.get(10);
    this.inst = buffers.get(11);
    this.isLogX0 = buffers.get(12);
    this.isLogX1 = buffers.get(13);
    this.isLogX2 = buffers.get(14);
    this.isLogX3 = buffers.get(15);
    this.isLogX4 = buffers.get(16);
    this.phase = buffers.get(17);
    this.topicHi1 = buffers.get(18);
    this.topicHi2 = buffers.get(19);
    this.topicHi3 = buffers.get(20);
    this.topicHi4 = buffers.get(21);
    this.topicLo1 = buffers.get(22);
    this.topicLo2 = buffers.get(23);
    this.topicLo3 = buffers.get(24);
    this.topicLo4 = buffers.get(25);
    this.txnEmitsLogs = buffers.get(26);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absLogNum(final BigInteger b) {
    if (filled.get(0)) {
      throw new IllegalStateException("logInfo.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    absLogNum.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace absLogNumMax(final BigInteger b) {
    if (filled.get(1)) {
      throw new IllegalStateException("logInfo.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    absLogNumMax.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace absTxnNum(final BigInteger b) {
    if (filled.get(2)) {
      throw new IllegalStateException("logInfo.ABS_TXN_NUM already set");
    } else {
      filled.set(2);
    }

    absTxnNum.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace absTxnNumMax(final BigInteger b) {
    if (filled.get(3)) {
      throw new IllegalStateException("logInfo.ABS_TXN_NUM_MAX already set");
    } else {
      filled.set(3);
    }

    absTxnNumMax.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace addrHi(final BigInteger b) {
    if (filled.get(4)) {
      throw new IllegalStateException("logInfo.ADDR_HI already set");
    } else {
      filled.set(4);
    }

    addrHi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace addrLo(final BigInteger b) {
    if (filled.get(5)) {
      throw new IllegalStateException("logInfo.ADDR_LO already set");
    } else {
      filled.set(5);
    }

    addrLo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace ct(final BigInteger b) {
    if (filled.get(6)) {
      throw new IllegalStateException("logInfo.CT already set");
    } else {
      filled.set(6);
    }

    ct.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace ctMax(final BigInteger b) {
    if (filled.get(7)) {
      throw new IllegalStateException("logInfo.CT_MAX already set");
    } else {
      filled.set(7);
    }

    ctMax.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace dataHi(final BigInteger b) {
    if (filled.get(8)) {
      throw new IllegalStateException("logInfo.DATA_HI already set");
    } else {
      filled.set(8);
    }

    dataHi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace dataLo(final BigInteger b) {
    if (filled.get(9)) {
      throw new IllegalStateException("logInfo.DATA_LO already set");
    } else {
      filled.set(9);
    }

    dataLo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace dataSize(final BigInteger b) {
    if (filled.get(10)) {
      throw new IllegalStateException("logInfo.DATA_SIZE already set");
    } else {
      filled.set(10);
    }

    dataSize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace inst(final BigInteger b) {
    if (filled.get(11)) {
      throw new IllegalStateException("logInfo.INST already set");
    } else {
      filled.set(11);
    }

    inst.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace isLogX0(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_0 already set");
    } else {
      filled.set(12);
    }

    isLogX0.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX1(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_1 already set");
    } else {
      filled.set(13);
    }

    isLogX1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX2(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_2 already set");
    } else {
      filled.set(14);
    }

    isLogX2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX3(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_3 already set");
    } else {
      filled.set(15);
    }

    isLogX3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX4(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_4 already set");
    } else {
      filled.set(16);
    }

    isLogX4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase(final BigInteger b) {
    if (filled.get(17)) {
      throw new IllegalStateException("logInfo.PHASE already set");
    } else {
      filled.set(17);
    }

    phase.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicHi1(final BigInteger b) {
    if (filled.get(18)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_1 already set");
    } else {
      filled.set(18);
    }

    topicHi1.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicHi2(final BigInteger b) {
    if (filled.get(19)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_2 already set");
    } else {
      filled.set(19);
    }

    topicHi2.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicHi3(final BigInteger b) {
    if (filled.get(20)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_3 already set");
    } else {
      filled.set(20);
    }

    topicHi3.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicHi4(final BigInteger b) {
    if (filled.get(21)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_4 already set");
    } else {
      filled.set(21);
    }

    topicHi4.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicLo1(final BigInteger b) {
    if (filled.get(22)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_1 already set");
    } else {
      filled.set(22);
    }

    topicLo1.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicLo2(final BigInteger b) {
    if (filled.get(23)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_2 already set");
    } else {
      filled.set(23);
    }

    topicLo2.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicLo3(final BigInteger b) {
    if (filled.get(24)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_3 already set");
    } else {
      filled.set(24);
    }

    topicLo3.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace topicLo4(final BigInteger b) {
    if (filled.get(25)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_4 already set");
    } else {
      filled.set(25);
    }

    topicLo4.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace txnEmitsLogs(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("logInfo.TXN_EMITS_LOGS already set");
    } else {
      filled.set(26);
    }

    txnEmitsLogs.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("logInfo.ABS_LOG_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("logInfo.ABS_LOG_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("logInfo.ABS_TXN_NUM has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("logInfo.ABS_TXN_NUM_MAX has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("logInfo.ADDR_HI has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("logInfo.ADDR_LO has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("logInfo.CT has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("logInfo.CT_MAX has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("logInfo.DATA_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("logInfo.DATA_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("logInfo.DATA_SIZE has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("logInfo.INST has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_0 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_3 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("logInfo.IS_LOG_X_4 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("logInfo.PHASE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_1 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_2 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_3 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("logInfo.TOPIC_HI_4 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_1 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_2 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_3 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("logInfo.TOPIC_LO_4 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("logInfo.TXN_EMITS_LOGS has not been filled");
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
      absTxnNum.position(absTxnNum.position() + 32);
    }

    if (!filled.get(3)) {
      absTxnNumMax.position(absTxnNumMax.position() + 32);
    }

    if (!filled.get(4)) {
      addrHi.position(addrHi.position() + 32);
    }

    if (!filled.get(5)) {
      addrLo.position(addrLo.position() + 32);
    }

    if (!filled.get(6)) {
      ct.position(ct.position() + 32);
    }

    if (!filled.get(7)) {
      ctMax.position(ctMax.position() + 32);
    }

    if (!filled.get(8)) {
      dataHi.position(dataHi.position() + 32);
    }

    if (!filled.get(9)) {
      dataLo.position(dataLo.position() + 32);
    }

    if (!filled.get(10)) {
      dataSize.position(dataSize.position() + 32);
    }

    if (!filled.get(11)) {
      inst.position(inst.position() + 32);
    }

    if (!filled.get(12)) {
      isLogX0.position(isLogX0.position() + 1);
    }

    if (!filled.get(13)) {
      isLogX1.position(isLogX1.position() + 1);
    }

    if (!filled.get(14)) {
      isLogX2.position(isLogX2.position() + 1);
    }

    if (!filled.get(15)) {
      isLogX3.position(isLogX3.position() + 1);
    }

    if (!filled.get(16)) {
      isLogX4.position(isLogX4.position() + 1);
    }

    if (!filled.get(17)) {
      phase.position(phase.position() + 32);
    }

    if (!filled.get(18)) {
      topicHi1.position(topicHi1.position() + 32);
    }

    if (!filled.get(19)) {
      topicHi2.position(topicHi2.position() + 32);
    }

    if (!filled.get(20)) {
      topicHi3.position(topicHi3.position() + 32);
    }

    if (!filled.get(21)) {
      topicHi4.position(topicHi4.position() + 32);
    }

    if (!filled.get(22)) {
      topicLo1.position(topicLo1.position() + 32);
    }

    if (!filled.get(23)) {
      topicLo2.position(topicLo2.position() + 32);
    }

    if (!filled.get(24)) {
      topicLo3.position(topicLo3.position() + 32);
    }

    if (!filled.get(25)) {
      topicLo4.position(topicLo4.position() + 32);
    }

    if (!filled.get(26)) {
      txnEmitsLogs.position(txnEmitsLogs.position() + 1);
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
