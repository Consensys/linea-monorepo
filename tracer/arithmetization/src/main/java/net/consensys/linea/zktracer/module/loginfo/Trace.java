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

package net.consensys.linea.zktracer.module.loginfo;

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
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("loginfo.ABS_LOG_NUM", 3, length));
      headers.add(new ColumnHeader("loginfo.ABS_LOG_NUM_MAX", 3, length));
      headers.add(new ColumnHeader("loginfo.ABS_TXN_NUM", 3, length));
      headers.add(new ColumnHeader("loginfo.ABS_TXN_NUM_MAX", 3, length));
      headers.add(new ColumnHeader("loginfo.ADDR_HI", 4, length));
      headers.add(new ColumnHeader("loginfo.ADDR_LO", 16, length));
      headers.add(new ColumnHeader("loginfo.CT", 1, length));
      headers.add(new ColumnHeader("loginfo.CT_MAX", 1, length));
      headers.add(new ColumnHeader("loginfo.DATA_HI", 16, length));
      headers.add(new ColumnHeader("loginfo.DATA_LO", 16, length));
      headers.add(new ColumnHeader("loginfo.DATA_SIZE", 4, length));
      headers.add(new ColumnHeader("loginfo.INST", 1, length));
      headers.add(new ColumnHeader("loginfo.IS_LOG_X_0", 1, length));
      headers.add(new ColumnHeader("loginfo.IS_LOG_X_1", 1, length));
      headers.add(new ColumnHeader("loginfo.IS_LOG_X_2", 1, length));
      headers.add(new ColumnHeader("loginfo.IS_LOG_X_3", 1, length));
      headers.add(new ColumnHeader("loginfo.IS_LOG_X_4", 1, length));
      headers.add(new ColumnHeader("loginfo.PHASE", 2, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_HI_1", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_HI_2", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_HI_3", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_HI_4", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_LO_1", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_LO_2", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_LO_3", 32, length));
      headers.add(new ColumnHeader("loginfo.TOPIC_LO_4", 32, length));
      headers.add(new ColumnHeader("loginfo.TXN_EMITS_LOGS", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
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

  public Trace absLogNum(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("loginfo.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("loginfo.ABS_LOG_NUM has invalid value (" + b + ")"); }
    absLogNum.put((byte) (b >> 16));
    absLogNum.put((byte) (b >> 8));
    absLogNum.put((byte) b);


    return this;
  }

  public Trace absLogNumMax(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("loginfo.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("loginfo.ABS_LOG_NUM_MAX has invalid value (" + b + ")"); }
    absLogNumMax.put((byte) (b >> 16));
    absLogNumMax.put((byte) (b >> 8));
    absLogNumMax.put((byte) b);


    return this;
  }

  public Trace absTxnNum(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("loginfo.ABS_TXN_NUM already set");
    } else {
      filled.set(2);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("loginfo.ABS_TXN_NUM has invalid value (" + b + ")"); }
    absTxnNum.put((byte) (b >> 16));
    absTxnNum.put((byte) (b >> 8));
    absTxnNum.put((byte) b);


    return this;
  }

  public Trace absTxnNumMax(final long b) {
    if (filled.get(3)) {
      throw new IllegalStateException("loginfo.ABS_TXN_NUM_MAX already set");
    } else {
      filled.set(3);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("loginfo.ABS_TXN_NUM_MAX has invalid value (" + b + ")"); }
    absTxnNumMax.put((byte) (b >> 16));
    absTxnNumMax.put((byte) (b >> 8));
    absTxnNumMax.put((byte) b);


    return this;
  }

  public Trace addrHi(final long b) {
    if (filled.get(4)) {
      throw new IllegalStateException("loginfo.ADDR_HI already set");
    } else {
      filled.set(4);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("loginfo.ADDR_HI has invalid value (" + b + ")"); }
    addrHi.put((byte) (b >> 24));
    addrHi.put((byte) (b >> 16));
    addrHi.put((byte) (b >> 8));
    addrHi.put((byte) b);


    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("loginfo.ADDR_LO already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("loginfo.ADDR_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addrLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addrLo.put(bs.get(j)); }

    return this;
  }

  public Trace ct(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("loginfo.CT already set");
    } else {
      filled.set(6);
    }

    ct.put(b.toByte());

    return this;
  }

  public Trace ctMax(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("loginfo.CT_MAX already set");
    } else {
      filled.set(7);
    }

    ctMax.put(b.toByte());

    return this;
  }

  public Trace dataHi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("loginfo.DATA_HI already set");
    } else {
      filled.set(8);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("loginfo.DATA_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { dataHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { dataHi.put(bs.get(j)); }

    return this;
  }

  public Trace dataLo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("loginfo.DATA_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("loginfo.DATA_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { dataLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { dataLo.put(bs.get(j)); }

    return this;
  }

  public Trace dataSize(final long b) {
    if (filled.get(10)) {
      throw new IllegalStateException("loginfo.DATA_SIZE already set");
    } else {
      filled.set(10);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("loginfo.DATA_SIZE has invalid value (" + b + ")"); }
    dataSize.put((byte) (b >> 24));
    dataSize.put((byte) (b >> 16));
    dataSize.put((byte) (b >> 8));
    dataSize.put((byte) b);


    return this;
  }

  public Trace inst(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("loginfo.INST already set");
    } else {
      filled.set(11);
    }

    inst.put(b.toByte());

    return this;
  }

  public Trace isLogX0(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_0 already set");
    } else {
      filled.set(12);
    }

    isLogX0.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX1(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_1 already set");
    } else {
      filled.set(13);
    }

    isLogX1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX2(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_2 already set");
    } else {
      filled.set(14);
    }

    isLogX2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX3(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_3 already set");
    } else {
      filled.set(15);
    }

    isLogX3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLogX4(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_4 already set");
    } else {
      filled.set(16);
    }

    isLogX4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase(final long b) {
    if (filled.get(17)) {
      throw new IllegalStateException("loginfo.PHASE already set");
    } else {
      filled.set(17);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("loginfo.PHASE has invalid value (" + b + ")"); }
    phase.put((byte) (b >> 8));
    phase.put((byte) b);


    return this;
  }

  public Trace topicHi1(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_1 already set");
    } else {
      filled.set(18);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_HI_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicHi1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicHi1.put(bs.get(j)); }

    return this;
  }

  public Trace topicHi2(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_2 already set");
    } else {
      filled.set(19);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_HI_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicHi2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicHi2.put(bs.get(j)); }

    return this;
  }

  public Trace topicHi3(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_3 already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_HI_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicHi3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicHi3.put(bs.get(j)); }

    return this;
  }

  public Trace topicHi4(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_4 already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_HI_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicHi4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicHi4.put(bs.get(j)); }

    return this;
  }

  public Trace topicLo1(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_1 already set");
    } else {
      filled.set(22);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_LO_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicLo1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicLo1.put(bs.get(j)); }

    return this;
  }

  public Trace topicLo2(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_2 already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_LO_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicLo2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicLo2.put(bs.get(j)); }

    return this;
  }

  public Trace topicLo3(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_3 already set");
    } else {
      filled.set(24);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_LO_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicLo3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicLo3.put(bs.get(j)); }

    return this;
  }

  public Trace topicLo4(final Bytes b) {
    if (filled.get(25)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_4 already set");
    } else {
      filled.set(25);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("loginfo.TOPIC_LO_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { topicLo4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { topicLo4.put(bs.get(j)); }

    return this;
  }

  public Trace txnEmitsLogs(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("loginfo.TXN_EMITS_LOGS already set");
    } else {
      filled.set(26);
    }

    txnEmitsLogs.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("loginfo.ABS_LOG_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("loginfo.ABS_LOG_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("loginfo.ABS_TXN_NUM has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("loginfo.ABS_TXN_NUM_MAX has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("loginfo.ADDR_HI has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("loginfo.ADDR_LO has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("loginfo.CT has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("loginfo.CT_MAX has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("loginfo.DATA_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("loginfo.DATA_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("loginfo.DATA_SIZE has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("loginfo.INST has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_0 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_1 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_2 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_3 has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("loginfo.IS_LOG_X_4 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("loginfo.PHASE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_1 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_2 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_3 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("loginfo.TOPIC_HI_4 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_1 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_2 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_3 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("loginfo.TOPIC_LO_4 has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("loginfo.TXN_EMITS_LOGS has not been filled");
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
      absTxnNum.position(absTxnNum.position() + 3);
    }

    if (!filled.get(3)) {
      absTxnNumMax.position(absTxnNumMax.position() + 3);
    }

    if (!filled.get(4)) {
      addrHi.position(addrHi.position() + 4);
    }

    if (!filled.get(5)) {
      addrLo.position(addrLo.position() + 16);
    }

    if (!filled.get(6)) {
      ct.position(ct.position() + 1);
    }

    if (!filled.get(7)) {
      ctMax.position(ctMax.position() + 1);
    }

    if (!filled.get(8)) {
      dataHi.position(dataHi.position() + 16);
    }

    if (!filled.get(9)) {
      dataLo.position(dataLo.position() + 16);
    }

    if (!filled.get(10)) {
      dataSize.position(dataSize.position() + 4);
    }

    if (!filled.get(11)) {
      inst.position(inst.position() + 1);
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
      phase.position(phase.position() + 2);
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
