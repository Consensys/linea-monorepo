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

package net.consensys.linea.zktracer.module.rlptxrcpt;

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
  public static final int SUBPHASE_ID_WEIGHT_DEPTH = 0x30;
  public static final int SUBPHASE_ID_WEIGHT_INDEX_LOCAL = 0x60;
  public static final int SUBPHASE_ID_WEIGHT_IS_OD = 0x18;
  public static final int SUBPHASE_ID_WEIGHT_IS_OT = 0xc;
  public static final int SUBPHASE_ID_WEIGHT_IS_PREFIX = 0x6;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer absLogNum;
  private final MappedByteBuffer absLogNumMax;
  private final MappedByteBuffer absTxNum;
  private final MappedByteBuffer absTxNumMax;
  private final MappedByteBuffer acc1;
  private final MappedByteBuffer acc2;
  private final MappedByteBuffer acc3;
  private final MappedByteBuffer acc4;
  private final MappedByteBuffer accSize;
  private final MappedByteBuffer bit;
  private final MappedByteBuffer bitAcc;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer byte3;
  private final MappedByteBuffer byte4;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer depth1;
  private final MappedByteBuffer done;
  private final MappedByteBuffer index;
  private final MappedByteBuffer indexLocal;
  private final MappedByteBuffer input1;
  private final MappedByteBuffer input2;
  private final MappedByteBuffer input3;
  private final MappedByteBuffer input4;
  private final MappedByteBuffer isData;
  private final MappedByteBuffer isPrefix;
  private final MappedByteBuffer isTopic;
  private final MappedByteBuffer lcCorrection;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer limbConstructed;
  private final MappedByteBuffer localSize;
  private final MappedByteBuffer logEntrySize;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nStep;
  private final MappedByteBuffer phase1;
  private final MappedByteBuffer phase2;
  private final MappedByteBuffer phase3;
  private final MappedByteBuffer phase4;
  private final MappedByteBuffer phase5;
  private final MappedByteBuffer phaseEnd;
  private final MappedByteBuffer phaseId;
  private final MappedByteBuffer phaseSize;
  private final MappedByteBuffer power;
  private final MappedByteBuffer txrcptSize;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("rlptxrcpt.ABS_LOG_NUM", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.ABS_LOG_NUM_MAX", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.ABS_TX_NUM", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.ABS_TX_NUM_MAX", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.ACC_1", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.ACC_2", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.ACC_3", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.ACC_4", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.ACC_SIZE", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.BIT", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.BIT_ACC", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.BYTE_1", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.BYTE_2", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.BYTE_3", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.BYTE_4", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.COUNTER", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.DEPTH_1", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.DONE", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.INDEX", 3, length));
      headers.add(new ColumnHeader("rlptxrcpt.INDEX_LOCAL", 3, length));
      headers.add(new ColumnHeader("rlptxrcpt.INPUT_1", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.INPUT_2", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.INPUT_3", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.INPUT_4", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.IS_DATA", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.IS_PREFIX", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.IS_TOPIC", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.LC_CORRECTION", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.LIMB", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.LIMB_CONSTRUCTED", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.LOCAL_SIZE", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.LOG_ENTRY_SIZE", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.nBYTES", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.nSTEP", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_1", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_2", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_3", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_4", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_5", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_END", 1, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_ID", 2, length));
      headers.add(new ColumnHeader("rlptxrcpt.PHASE_SIZE", 4, length));
      headers.add(new ColumnHeader("rlptxrcpt.POWER", 16, length));
      headers.add(new ColumnHeader("rlptxrcpt.TXRCPT_SIZE", 4, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.absLogNum = buffers.get(0);
    this.absLogNumMax = buffers.get(1);
    this.absTxNum = buffers.get(2);
    this.absTxNumMax = buffers.get(3);
    this.acc1 = buffers.get(4);
    this.acc2 = buffers.get(5);
    this.acc3 = buffers.get(6);
    this.acc4 = buffers.get(7);
    this.accSize = buffers.get(8);
    this.bit = buffers.get(9);
    this.bitAcc = buffers.get(10);
    this.byte1 = buffers.get(11);
    this.byte2 = buffers.get(12);
    this.byte3 = buffers.get(13);
    this.byte4 = buffers.get(14);
    this.counter = buffers.get(15);
    this.depth1 = buffers.get(16);
    this.done = buffers.get(17);
    this.index = buffers.get(18);
    this.indexLocal = buffers.get(19);
    this.input1 = buffers.get(20);
    this.input2 = buffers.get(21);
    this.input3 = buffers.get(22);
    this.input4 = buffers.get(23);
    this.isData = buffers.get(24);
    this.isPrefix = buffers.get(25);
    this.isTopic = buffers.get(26);
    this.lcCorrection = buffers.get(27);
    this.limb = buffers.get(28);
    this.limbConstructed = buffers.get(29);
    this.localSize = buffers.get(30);
    this.logEntrySize = buffers.get(31);
    this.nBytes = buffers.get(32);
    this.nStep = buffers.get(33);
    this.phase1 = buffers.get(34);
    this.phase2 = buffers.get(35);
    this.phase3 = buffers.get(36);
    this.phase4 = buffers.get(37);
    this.phase5 = buffers.get(38);
    this.phaseEnd = buffers.get(39);
    this.phaseId = buffers.get(40);
    this.phaseSize = buffers.get(41);
    this.power = buffers.get(42);
    this.txrcptSize = buffers.get(43);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absLogNum(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlptxrcpt.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.ABS_LOG_NUM has invalid value (" + b + ")"); }
    absLogNum.put((byte) (b >> 24));
    absLogNum.put((byte) (b >> 16));
    absLogNum.put((byte) (b >> 8));
    absLogNum.put((byte) b);


    return this;
  }

  public Trace absLogNumMax(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlptxrcpt.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.ABS_LOG_NUM_MAX has invalid value (" + b + ")"); }
    absLogNumMax.put((byte) (b >> 24));
    absLogNumMax.put((byte) (b >> 16));
    absLogNumMax.put((byte) (b >> 8));
    absLogNumMax.put((byte) b);


    return this;
  }

  public Trace absTxNum(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlptxrcpt.ABS_TX_NUM already set");
    } else {
      filled.set(2);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.ABS_TX_NUM has invalid value (" + b + ")"); }
    absTxNum.put((byte) (b >> 24));
    absTxNum.put((byte) (b >> 16));
    absTxNum.put((byte) (b >> 8));
    absTxNum.put((byte) b);


    return this;
  }

  public Trace absTxNumMax(final long b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlptxrcpt.ABS_TX_NUM_MAX already set");
    } else {
      filled.set(3);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.ABS_TX_NUM_MAX has invalid value (" + b + ")"); }
    absTxNumMax.put((byte) (b >> 24));
    absTxNumMax.put((byte) (b >> 16));
    absTxNumMax.put((byte) (b >> 8));
    absTxNumMax.put((byte) b);


    return this;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlptxrcpt.ACC_1 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.ACC_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc1.put(bs.get(j)); }

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlptxrcpt.ACC_2 already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.ACC_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc2.put(bs.get(j)); }

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlptxrcpt.ACC_3 already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.ACC_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc3.put(bs.get(j)); }

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlptxrcpt.ACC_4 already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.ACC_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc4.put(bs.get(j)); }

    return this;
  }

  public Trace accSize(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlptxrcpt.ACC_SIZE already set");
    } else {
      filled.set(8);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxrcpt.ACC_SIZE has invalid value (" + b + ")"); }
    accSize.put((byte) b);


    return this;
  }

  public Trace bit(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlptxrcpt.BIT already set");
    } else {
      filled.set(9);
    }

    bit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlptxrcpt.BIT_ACC already set");
    } else {
      filled.set(10);
    }

    bitAcc.put(b.toByte());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_1 already set");
    } else {
      filled.set(11);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_2 already set");
    } else {
      filled.set(12);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_3 already set");
    } else {
      filled.set(13);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_4 already set");
    } else {
      filled.set(14);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace counter(final long b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlptxrcpt.COUNTER already set");
    } else {
      filled.set(15);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.COUNTER has invalid value (" + b + ")"); }
    counter.put((byte) (b >> 24));
    counter.put((byte) (b >> 16));
    counter.put((byte) (b >> 8));
    counter.put((byte) b);


    return this;
  }

  public Trace depth1(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlptxrcpt.DEPTH_1 already set");
    } else {
      filled.set(16);
    }

    depth1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace done(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlptxrcpt.DONE already set");
    } else {
      filled.set(17);
    }

    done.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace index(final long b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlptxrcpt.INDEX already set");
    } else {
      filled.set(18);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("rlptxrcpt.INDEX has invalid value (" + b + ")"); }
    index.put((byte) (b >> 16));
    index.put((byte) (b >> 8));
    index.put((byte) b);


    return this;
  }

  public Trace indexLocal(final long b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlptxrcpt.INDEX_LOCAL already set");
    } else {
      filled.set(19);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("rlptxrcpt.INDEX_LOCAL has invalid value (" + b + ")"); }
    indexLocal.put((byte) (b >> 16));
    indexLocal.put((byte) (b >> 8));
    indexLocal.put((byte) b);


    return this;
  }

  public Trace input1(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_1 already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.INPUT_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { input1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { input1.put(bs.get(j)); }

    return this;
  }

  public Trace input2(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_2 already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.INPUT_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { input2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { input2.put(bs.get(j)); }

    return this;
  }

  public Trace input3(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_3 already set");
    } else {
      filled.set(22);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.INPUT_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { input3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { input3.put(bs.get(j)); }

    return this;
  }

  public Trace input4(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_4 already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.INPUT_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { input4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { input4.put(bs.get(j)); }

    return this;
  }

  public Trace isData(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlptxrcpt.IS_DATA already set");
    } else {
      filled.set(24);
    }

    isData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPrefix(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("rlptxrcpt.IS_PREFIX already set");
    } else {
      filled.set(25);
    }

    isPrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isTopic(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("rlptxrcpt.IS_TOPIC already set");
    } else {
      filled.set(26);
    }

    isTopic.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lcCorrection(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("rlptxrcpt.LC_CORRECTION already set");
    } else {
      filled.set(27);
    }

    lcCorrection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("rlptxrcpt.LIMB already set");
    } else {
      filled.set(28);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

    return this;
  }

  public Trace limbConstructed(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("rlptxrcpt.LIMB_CONSTRUCTED already set");
    } else {
      filled.set(29);
    }

    limbConstructed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace localSize(final long b) {
    if (filled.get(30)) {
      throw new IllegalStateException("rlptxrcpt.LOCAL_SIZE already set");
    } else {
      filled.set(30);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.LOCAL_SIZE has invalid value (" + b + ")"); }
    localSize.put((byte) (b >> 24));
    localSize.put((byte) (b >> 16));
    localSize.put((byte) (b >> 8));
    localSize.put((byte) b);


    return this;
  }

  public Trace logEntrySize(final long b) {
    if (filled.get(31)) {
      throw new IllegalStateException("rlptxrcpt.LOG_ENTRY_SIZE already set");
    } else {
      filled.set(31);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.LOG_ENTRY_SIZE has invalid value (" + b + ")"); }
    logEntrySize.put((byte) (b >> 24));
    logEntrySize.put((byte) (b >> 16));
    logEntrySize.put((byte) (b >> 8));
    logEntrySize.put((byte) b);


    return this;
  }

  public Trace nBytes(final long b) {
    if (filled.get(42)) {
      throw new IllegalStateException("rlptxrcpt.nBYTES already set");
    } else {
      filled.set(42);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxrcpt.nBYTES has invalid value (" + b + ")"); }
    nBytes.put((byte) b);


    return this;
  }

  public Trace nStep(final long b) {
    if (filled.get(43)) {
      throw new IllegalStateException("rlptxrcpt.nSTEP already set");
    } else {
      filled.set(43);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.nSTEP has invalid value (" + b + ")"); }
    nStep.put((byte) (b >> 24));
    nStep.put((byte) (b >> 16));
    nStep.put((byte) (b >> 8));
    nStep.put((byte) b);


    return this;
  }

  public Trace phase1(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_1 already set");
    } else {
      filled.set(32);
    }

    phase1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase2(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_2 already set");
    } else {
      filled.set(33);
    }

    phase2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase3(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_3 already set");
    } else {
      filled.set(34);
    }

    phase3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase4(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_4 already set");
    } else {
      filled.set(35);
    }

    phase4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase5(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_5 already set");
    } else {
      filled.set(36);
    }

    phase5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseEnd(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_END already set");
    } else {
      filled.set(37);
    }

    phaseEnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseId(final long b) {
    if (filled.get(38)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_ID already set");
    } else {
      filled.set(38);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("rlptxrcpt.PHASE_ID has invalid value (" + b + ")"); }
    phaseId.put((byte) (b >> 8));
    phaseId.put((byte) b);


    return this;
  }

  public Trace phaseSize(final long b) {
    if (filled.get(39)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_SIZE already set");
    } else {
      filled.set(39);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.PHASE_SIZE has invalid value (" + b + ")"); }
    phaseSize.put((byte) (b >> 24));
    phaseSize.put((byte) (b >> 16));
    phaseSize.put((byte) (b >> 8));
    phaseSize.put((byte) b);


    return this;
  }

  public Trace power(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("rlptxrcpt.POWER already set");
    } else {
      filled.set(40);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxrcpt.POWER has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { power.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { power.put(bs.get(j)); }

    return this;
  }

  public Trace txrcptSize(final long b) {
    if (filled.get(41)) {
      throw new IllegalStateException("rlptxrcpt.TXRCPT_SIZE already set");
    } else {
      filled.set(41);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxrcpt.TXRCPT_SIZE has invalid value (" + b + ")"); }
    txrcptSize.put((byte) (b >> 24));
    txrcptSize.put((byte) (b >> 16));
    txrcptSize.put((byte) (b >> 8));
    txrcptSize.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlptxrcpt.ABS_LOG_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlptxrcpt.ABS_LOG_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlptxrcpt.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlptxrcpt.ABS_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlptxrcpt.ACC_1 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlptxrcpt.ACC_2 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlptxrcpt.ACC_3 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlptxrcpt.ACC_4 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlptxrcpt.ACC_SIZE has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlptxrcpt.BIT has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlptxrcpt.BIT_ACC has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_1 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_2 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_3 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlptxrcpt.BYTE_4 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlptxrcpt.COUNTER has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlptxrcpt.DEPTH_1 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlptxrcpt.DONE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlptxrcpt.INDEX has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlptxrcpt.INDEX_LOCAL has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_1 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_2 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_3 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlptxrcpt.INPUT_4 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlptxrcpt.IS_DATA has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("rlptxrcpt.IS_PREFIX has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("rlptxrcpt.IS_TOPIC has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("rlptxrcpt.LC_CORRECTION has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("rlptxrcpt.LIMB has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("rlptxrcpt.LIMB_CONSTRUCTED has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("rlptxrcpt.LOCAL_SIZE has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("rlptxrcpt.LOG_ENTRY_SIZE has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("rlptxrcpt.nBYTES has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("rlptxrcpt.nSTEP has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_2 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_3 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_4 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_5 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_END has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_ID has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("rlptxrcpt.PHASE_SIZE has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("rlptxrcpt.POWER has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("rlptxrcpt.TXRCPT_SIZE has not been filled");
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
      absTxNum.position(absTxNum.position() + 4);
    }

    if (!filled.get(3)) {
      absTxNumMax.position(absTxNumMax.position() + 4);
    }

    if (!filled.get(4)) {
      acc1.position(acc1.position() + 16);
    }

    if (!filled.get(5)) {
      acc2.position(acc2.position() + 16);
    }

    if (!filled.get(6)) {
      acc3.position(acc3.position() + 16);
    }

    if (!filled.get(7)) {
      acc4.position(acc4.position() + 16);
    }

    if (!filled.get(8)) {
      accSize.position(accSize.position() + 1);
    }

    if (!filled.get(9)) {
      bit.position(bit.position() + 1);
    }

    if (!filled.get(10)) {
      bitAcc.position(bitAcc.position() + 1);
    }

    if (!filled.get(11)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(12)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(13)) {
      byte3.position(byte3.position() + 1);
    }

    if (!filled.get(14)) {
      byte4.position(byte4.position() + 1);
    }

    if (!filled.get(15)) {
      counter.position(counter.position() + 4);
    }

    if (!filled.get(16)) {
      depth1.position(depth1.position() + 1);
    }

    if (!filled.get(17)) {
      done.position(done.position() + 1);
    }

    if (!filled.get(18)) {
      index.position(index.position() + 3);
    }

    if (!filled.get(19)) {
      indexLocal.position(indexLocal.position() + 3);
    }

    if (!filled.get(20)) {
      input1.position(input1.position() + 16);
    }

    if (!filled.get(21)) {
      input2.position(input2.position() + 16);
    }

    if (!filled.get(22)) {
      input3.position(input3.position() + 16);
    }

    if (!filled.get(23)) {
      input4.position(input4.position() + 16);
    }

    if (!filled.get(24)) {
      isData.position(isData.position() + 1);
    }

    if (!filled.get(25)) {
      isPrefix.position(isPrefix.position() + 1);
    }

    if (!filled.get(26)) {
      isTopic.position(isTopic.position() + 1);
    }

    if (!filled.get(27)) {
      lcCorrection.position(lcCorrection.position() + 1);
    }

    if (!filled.get(28)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(29)) {
      limbConstructed.position(limbConstructed.position() + 1);
    }

    if (!filled.get(30)) {
      localSize.position(localSize.position() + 4);
    }

    if (!filled.get(31)) {
      logEntrySize.position(logEntrySize.position() + 4);
    }

    if (!filled.get(42)) {
      nBytes.position(nBytes.position() + 1);
    }

    if (!filled.get(43)) {
      nStep.position(nStep.position() + 4);
    }

    if (!filled.get(32)) {
      phase1.position(phase1.position() + 1);
    }

    if (!filled.get(33)) {
      phase2.position(phase2.position() + 1);
    }

    if (!filled.get(34)) {
      phase3.position(phase3.position() + 1);
    }

    if (!filled.get(35)) {
      phase4.position(phase4.position() + 1);
    }

    if (!filled.get(36)) {
      phase5.position(phase5.position() + 1);
    }

    if (!filled.get(37)) {
      phaseEnd.position(phaseEnd.position() + 1);
    }

    if (!filled.get(38)) {
      phaseId.position(phaseId.position() + 2);
    }

    if (!filled.get(39)) {
      phaseSize.position(phaseSize.position() + 4);
    }

    if (!filled.get(40)) {
      power.position(power.position() + 16);
    }

    if (!filled.get(41)) {
      txrcptSize.position(txrcptSize.position() + 4);
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
