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

package net.consensys.linea.zktracer.module.rlp.txrcpt;

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
  private final MappedByteBuffer phaseSize;
  private final MappedByteBuffer power;
  private final MappedByteBuffer txrcptSize;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("rlpTxRcpt.ABS_LOG_NUM", 32, length),
        new ColumnHeader("rlpTxRcpt.ABS_LOG_NUM_MAX", 32, length),
        new ColumnHeader("rlpTxRcpt.ABS_TX_NUM", 32, length),
        new ColumnHeader("rlpTxRcpt.ABS_TX_NUM_MAX", 32, length),
        new ColumnHeader("rlpTxRcpt.ACC_1", 32, length),
        new ColumnHeader("rlpTxRcpt.ACC_2", 32, length),
        new ColumnHeader("rlpTxRcpt.ACC_3", 32, length),
        new ColumnHeader("rlpTxRcpt.ACC_4", 32, length),
        new ColumnHeader("rlpTxRcpt.ACC_SIZE", 32, length),
        new ColumnHeader("rlpTxRcpt.BIT", 1, length),
        new ColumnHeader("rlpTxRcpt.BIT_ACC", 1, length),
        new ColumnHeader("rlpTxRcpt.BYTE_1", 1, length),
        new ColumnHeader("rlpTxRcpt.BYTE_2", 1, length),
        new ColumnHeader("rlpTxRcpt.BYTE_3", 1, length),
        new ColumnHeader("rlpTxRcpt.BYTE_4", 1, length),
        new ColumnHeader("rlpTxRcpt.COUNTER", 32, length),
        new ColumnHeader("rlpTxRcpt.DEPTH_1", 1, length),
        new ColumnHeader("rlpTxRcpt.DONE", 1, length),
        new ColumnHeader("rlpTxRcpt.INDEX", 32, length),
        new ColumnHeader("rlpTxRcpt.INDEX_LOCAL", 32, length),
        new ColumnHeader("rlpTxRcpt.INPUT_1", 32, length),
        new ColumnHeader("rlpTxRcpt.INPUT_2", 32, length),
        new ColumnHeader("rlpTxRcpt.INPUT_3", 32, length),
        new ColumnHeader("rlpTxRcpt.INPUT_4", 32, length),
        new ColumnHeader("rlpTxRcpt.IS_DATA", 1, length),
        new ColumnHeader("rlpTxRcpt.IS_PREFIX", 1, length),
        new ColumnHeader("rlpTxRcpt.IS_TOPIC", 1, length),
        new ColumnHeader("rlpTxRcpt.LC_CORRECTION", 1, length),
        new ColumnHeader("rlpTxRcpt.LIMB", 32, length),
        new ColumnHeader("rlpTxRcpt.LIMB_CONSTRUCTED", 1, length),
        new ColumnHeader("rlpTxRcpt.LOCAL_SIZE", 32, length),
        new ColumnHeader("rlpTxRcpt.LOG_ENTRY_SIZE", 32, length),
        new ColumnHeader("rlpTxRcpt.nBYTES", 1, length),
        new ColumnHeader("rlpTxRcpt.nSTEP", 32, length),
        new ColumnHeader("rlpTxRcpt.PHASE_1", 1, length),
        new ColumnHeader("rlpTxRcpt.PHASE_2", 1, length),
        new ColumnHeader("rlpTxRcpt.PHASE_3", 1, length),
        new ColumnHeader("rlpTxRcpt.PHASE_4", 1, length),
        new ColumnHeader("rlpTxRcpt.PHASE_5", 1, length),
        new ColumnHeader("rlpTxRcpt.PHASE_END", 1, length),
        new ColumnHeader("rlpTxRcpt.PHASE_SIZE", 32, length),
        new ColumnHeader("rlpTxRcpt.POWER", 32, length),
        new ColumnHeader("rlpTxRcpt.TXRCPT_SIZE", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
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
    this.phaseSize = buffers.get(40);
    this.power = buffers.get(41);
    this.txrcptSize = buffers.get(42);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absLogNum(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_LOG_NUM already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absLogNum.put((byte) 0);
    }
    absLogNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace absLogNumMax(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_LOG_NUM_MAX already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absLogNumMax.put((byte) 0);
    }
    absLogNumMax.put(b.toArrayUnsafe());

    return this;
  }

  public Trace absTxNum(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_TX_NUM already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absTxNum.put((byte) 0);
    }
    absTxNum.put(b.toArrayUnsafe());

    return this;
  }

  public Trace absTxNumMax(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_TX_NUM_MAX already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      absTxNumMax.put((byte) 0);
    }
    absTxNumMax.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_1 already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc1.put((byte) 0);
    }
    acc1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_2 already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc2.put((byte) 0);
    }
    acc2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_3 already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc3.put((byte) 0);
    }
    acc3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_4 already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc4.put((byte) 0);
    }
    acc4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accSize(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_SIZE already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accSize.put((byte) 0);
    }
    accSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlpTxRcpt.BIT already set");
    } else {
      filled.set(9);
    }

    bit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlpTxRcpt.BIT_ACC already set");
    } else {
      filled.set(10);
    }

    bitAcc.put(b.toByte());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_1 already set");
    } else {
      filled.set(11);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_2 already set");
    } else {
      filled.set(12);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace byte3(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_3 already set");
    } else {
      filled.set(13);
    }

    byte3.put(b.toByte());

    return this;
  }

  public Trace byte4(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_4 already set");
    } else {
      filled.set(14);
    }

    byte4.put(b.toByte());

    return this;
  }

  public Trace counter(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlpTxRcpt.COUNTER already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counter.put((byte) 0);
    }
    counter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace depth1(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlpTxRcpt.DEPTH_1 already set");
    } else {
      filled.set(16);
    }

    depth1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace done(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlpTxRcpt.DONE already set");
    } else {
      filled.set(17);
    }

    done.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace index(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlpTxRcpt.INDEX already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      index.put((byte) 0);
    }
    index.put(b.toArrayUnsafe());

    return this;
  }

  public Trace indexLocal(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlpTxRcpt.INDEX_LOCAL already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      indexLocal.put((byte) 0);
    }
    indexLocal.put(b.toArrayUnsafe());

    return this;
  }

  public Trace input1(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_1 already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      input1.put((byte) 0);
    }
    input1.put(b.toArrayUnsafe());

    return this;
  }

  public Trace input2(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_2 already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      input2.put((byte) 0);
    }
    input2.put(b.toArrayUnsafe());

    return this;
  }

  public Trace input3(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_3 already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      input3.put((byte) 0);
    }
    input3.put(b.toArrayUnsafe());

    return this;
  }

  public Trace input4(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_4 already set");
    } else {
      filled.set(23);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      input4.put((byte) 0);
    }
    input4.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isData(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlpTxRcpt.IS_DATA already set");
    } else {
      filled.set(24);
    }

    isData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPrefix(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("rlpTxRcpt.IS_PREFIX already set");
    } else {
      filled.set(25);
    }

    isPrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isTopic(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("rlpTxRcpt.IS_TOPIC already set");
    } else {
      filled.set(26);
    }

    isTopic.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lcCorrection(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("rlpTxRcpt.LC_CORRECTION already set");
    } else {
      filled.set(27);
    }

    lcCorrection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(28)) {
      throw new IllegalStateException("rlpTxRcpt.LIMB already set");
    } else {
      filled.set(28);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace limbConstructed(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("rlpTxRcpt.LIMB_CONSTRUCTED already set");
    } else {
      filled.set(29);
    }

    limbConstructed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace localSize(final Bytes b) {
    if (filled.get(30)) {
      throw new IllegalStateException("rlpTxRcpt.LOCAL_SIZE already set");
    } else {
      filled.set(30);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      localSize.put((byte) 0);
    }
    localSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace logEntrySize(final Bytes b) {
    if (filled.get(31)) {
      throw new IllegalStateException("rlpTxRcpt.LOG_ENTRY_SIZE already set");
    } else {
      filled.set(31);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      logEntrySize.put((byte) 0);
    }
    logEntrySize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nBytes(final UnsignedByte b) {
    if (filled.get(41)) {
      throw new IllegalStateException("rlpTxRcpt.nBYTES already set");
    } else {
      filled.set(41);
    }

    nBytes.put(b.toByte());

    return this;
  }

  public Trace nStep(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("rlpTxRcpt.nSTEP already set");
    } else {
      filled.set(42);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nStep.put((byte) 0);
    }
    nStep.put(b.toArrayUnsafe());

    return this;
  }

  public Trace phase1(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_1 already set");
    } else {
      filled.set(32);
    }

    phase1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase2(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_2 already set");
    } else {
      filled.set(33);
    }

    phase2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase3(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_3 already set");
    } else {
      filled.set(34);
    }

    phase3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase4(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_4 already set");
    } else {
      filled.set(35);
    }

    phase4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase5(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_5 already set");
    } else {
      filled.set(36);
    }

    phase5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseEnd(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_END already set");
    } else {
      filled.set(37);
    }

    phaseEnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseSize(final Bytes b) {
    if (filled.get(38)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_SIZE already set");
    } else {
      filled.set(38);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      phaseSize.put((byte) 0);
    }
    phaseSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace power(final Bytes b) {
    if (filled.get(39)) {
      throw new IllegalStateException("rlpTxRcpt.POWER already set");
    } else {
      filled.set(39);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      power.put((byte) 0);
    }
    power.put(b.toArrayUnsafe());

    return this;
  }

  public Trace txrcptSize(final Bytes b) {
    if (filled.get(40)) {
      throw new IllegalStateException("rlpTxRcpt.TXRCPT_SIZE already set");
    } else {
      filled.set(40);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      txrcptSize.put((byte) 0);
    }
    txrcptSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_LOG_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_LOG_NUM_MAX has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlpTxRcpt.ABS_TX_NUM_MAX has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_1 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_2 has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_3 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_4 has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlpTxRcpt.ACC_SIZE has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlpTxRcpt.BIT has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlpTxRcpt.BIT_ACC has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_1 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_2 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_3 has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlpTxRcpt.BYTE_4 has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlpTxRcpt.COUNTER has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlpTxRcpt.DEPTH_1 has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlpTxRcpt.DONE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlpTxRcpt.INDEX has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlpTxRcpt.INDEX_LOCAL has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_1 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_2 has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_3 has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlpTxRcpt.INPUT_4 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlpTxRcpt.IS_DATA has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("rlpTxRcpt.IS_PREFIX has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("rlpTxRcpt.IS_TOPIC has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("rlpTxRcpt.LC_CORRECTION has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("rlpTxRcpt.LIMB has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("rlpTxRcpt.LIMB_CONSTRUCTED has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("rlpTxRcpt.LOCAL_SIZE has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("rlpTxRcpt.LOG_ENTRY_SIZE has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("rlpTxRcpt.nBYTES has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("rlpTxRcpt.nSTEP has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_2 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_3 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_4 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_5 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_END has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("rlpTxRcpt.PHASE_SIZE has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("rlpTxRcpt.POWER has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("rlpTxRcpt.TXRCPT_SIZE has not been filled");
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
      absTxNum.position(absTxNum.position() + 32);
    }

    if (!filled.get(3)) {
      absTxNumMax.position(absTxNumMax.position() + 32);
    }

    if (!filled.get(4)) {
      acc1.position(acc1.position() + 32);
    }

    if (!filled.get(5)) {
      acc2.position(acc2.position() + 32);
    }

    if (!filled.get(6)) {
      acc3.position(acc3.position() + 32);
    }

    if (!filled.get(7)) {
      acc4.position(acc4.position() + 32);
    }

    if (!filled.get(8)) {
      accSize.position(accSize.position() + 32);
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
      counter.position(counter.position() + 32);
    }

    if (!filled.get(16)) {
      depth1.position(depth1.position() + 1);
    }

    if (!filled.get(17)) {
      done.position(done.position() + 1);
    }

    if (!filled.get(18)) {
      index.position(index.position() + 32);
    }

    if (!filled.get(19)) {
      indexLocal.position(indexLocal.position() + 32);
    }

    if (!filled.get(20)) {
      input1.position(input1.position() + 32);
    }

    if (!filled.get(21)) {
      input2.position(input2.position() + 32);
    }

    if (!filled.get(22)) {
      input3.position(input3.position() + 32);
    }

    if (!filled.get(23)) {
      input4.position(input4.position() + 32);
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
      limb.position(limb.position() + 32);
    }

    if (!filled.get(29)) {
      limbConstructed.position(limbConstructed.position() + 1);
    }

    if (!filled.get(30)) {
      localSize.position(localSize.position() + 32);
    }

    if (!filled.get(31)) {
      logEntrySize.position(logEntrySize.position() + 32);
    }

    if (!filled.get(41)) {
      nBytes.position(nBytes.position() + 1);
    }

    if (!filled.get(42)) {
      nStep.position(nStep.position() + 32);
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
      phaseSize.position(phaseSize.position() + 32);
    }

    if (!filled.get(39)) {
      power.position(power.position() + 32);
    }

    if (!filled.get(40)) {
      txrcptSize.position(txrcptSize.position() + 32);
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
