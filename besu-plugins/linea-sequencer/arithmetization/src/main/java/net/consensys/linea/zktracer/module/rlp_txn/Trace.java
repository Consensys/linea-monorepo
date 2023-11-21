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

package net.consensys.linea.zktracer.module.rlp_txn;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
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

  private final MappedByteBuffer absTxNum;
  private final MappedByteBuffer absTxNumInfiny;
  private final MappedByteBuffer acc1;
  private final MappedByteBuffer acc2;
  private final MappedByteBuffer accBytesize;
  private final MappedByteBuffer accessTupleBytesize;
  private final MappedByteBuffer addrHi;
  private final MappedByteBuffer addrLo;
  private final MappedByteBuffer bit;
  private final MappedByteBuffer bitAcc;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer byte2;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer dataHi;
  private final MappedByteBuffer dataLo;
  private final MappedByteBuffer datagascost;
  private final MappedByteBuffer depth1;
  private final MappedByteBuffer depth2;
  private final MappedByteBuffer done;
  private final MappedByteBuffer indexData;
  private final MappedByteBuffer indexLt;
  private final MappedByteBuffer indexLx;
  private final MappedByteBuffer input1;
  private final MappedByteBuffer input2;
  private final MappedByteBuffer isPrefix;
  private final MappedByteBuffer lcCorrection;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer limbConstructed;
  private final MappedByteBuffer lt;
  private final MappedByteBuffer lx;
  private final MappedByteBuffer nAddr;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nKeys;
  private final MappedByteBuffer nKeysPerAddr;
  private final MappedByteBuffer nStep;
  private final MappedByteBuffer phase0;
  private final MappedByteBuffer phase1;
  private final MappedByteBuffer phase10;
  private final MappedByteBuffer phase11;
  private final MappedByteBuffer phase12;
  private final MappedByteBuffer phase13;
  private final MappedByteBuffer phase14;
  private final MappedByteBuffer phase2;
  private final MappedByteBuffer phase3;
  private final MappedByteBuffer phase4;
  private final MappedByteBuffer phase5;
  private final MappedByteBuffer phase6;
  private final MappedByteBuffer phase7;
  private final MappedByteBuffer phase8;
  private final MappedByteBuffer phase9;
  private final MappedByteBuffer phaseEnd;
  private final MappedByteBuffer phaseSize;
  private final MappedByteBuffer power;
  private final MappedByteBuffer requiresEvmExecution;
  private final MappedByteBuffer rlpLtBytesize;
  private final MappedByteBuffer rlpLxBytesize;
  private final MappedByteBuffer type;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("rlpTxn.ABS_TX_NUM", 32, length),
        new ColumnHeader("rlpTxn.ABS_TX_NUM_INFINY", 32, length),
        new ColumnHeader("rlpTxn.ACC_1", 32, length),
        new ColumnHeader("rlpTxn.ACC_2", 32, length),
        new ColumnHeader("rlpTxn.ACC_BYTESIZE", 32, length),
        new ColumnHeader("rlpTxn.ACCESS_TUPLE_BYTESIZE", 32, length),
        new ColumnHeader("rlpTxn.ADDR_HI", 32, length),
        new ColumnHeader("rlpTxn.ADDR_LO", 32, length),
        new ColumnHeader("rlpTxn.BIT", 1, length),
        new ColumnHeader("rlpTxn.BIT_ACC", 32, length),
        new ColumnHeader("rlpTxn.BYTE_1", 1, length),
        new ColumnHeader("rlpTxn.BYTE_2", 1, length),
        new ColumnHeader("rlpTxn.CODE_FRAGMENT_INDEX", 32, length),
        new ColumnHeader("rlpTxn.COUNTER", 32, length),
        new ColumnHeader("rlpTxn.DATA_HI", 32, length),
        new ColumnHeader("rlpTxn.DATA_LO", 32, length),
        new ColumnHeader("rlpTxn.DATAGASCOST", 32, length),
        new ColumnHeader("rlpTxn.DEPTH_1", 1, length),
        new ColumnHeader("rlpTxn.DEPTH_2", 1, length),
        new ColumnHeader("rlpTxn.DONE", 1, length),
        new ColumnHeader("rlpTxn.INDEX_DATA", 32, length),
        new ColumnHeader("rlpTxn.INDEX_LT", 32, length),
        new ColumnHeader("rlpTxn.INDEX_LX", 32, length),
        new ColumnHeader("rlpTxn.INPUT_1", 32, length),
        new ColumnHeader("rlpTxn.INPUT_2", 32, length),
        new ColumnHeader("rlpTxn.IS_PREFIX", 1, length),
        new ColumnHeader("rlpTxn.LC_CORRECTION", 1, length),
        new ColumnHeader("rlpTxn.LIMB", 32, length),
        new ColumnHeader("rlpTxn.LIMB_CONSTRUCTED", 1, length),
        new ColumnHeader("rlpTxn.LT", 1, length),
        new ColumnHeader("rlpTxn.LX", 1, length),
        new ColumnHeader("rlpTxn.nADDR", 32, length),
        new ColumnHeader("rlpTxn.nBYTES", 32, length),
        new ColumnHeader("rlpTxn.nKEYS", 32, length),
        new ColumnHeader("rlpTxn.nKEYS_PER_ADDR", 32, length),
        new ColumnHeader("rlpTxn.nSTEP", 32, length),
        new ColumnHeader("rlpTxn.PHASE_0", 1, length),
        new ColumnHeader("rlpTxn.PHASE_1", 1, length),
        new ColumnHeader("rlpTxn.PHASE_10", 1, length),
        new ColumnHeader("rlpTxn.PHASE_11", 1, length),
        new ColumnHeader("rlpTxn.PHASE_12", 1, length),
        new ColumnHeader("rlpTxn.PHASE_13", 1, length),
        new ColumnHeader("rlpTxn.PHASE_14", 1, length),
        new ColumnHeader("rlpTxn.PHASE_2", 1, length),
        new ColumnHeader("rlpTxn.PHASE_3", 1, length),
        new ColumnHeader("rlpTxn.PHASE_4", 1, length),
        new ColumnHeader("rlpTxn.PHASE_5", 1, length),
        new ColumnHeader("rlpTxn.PHASE_6", 1, length),
        new ColumnHeader("rlpTxn.PHASE_7", 1, length),
        new ColumnHeader("rlpTxn.PHASE_8", 1, length),
        new ColumnHeader("rlpTxn.PHASE_9", 1, length),
        new ColumnHeader("rlpTxn.PHASE_END", 1, length),
        new ColumnHeader("rlpTxn.PHASE_SIZE", 32, length),
        new ColumnHeader("rlpTxn.POWER", 32, length),
        new ColumnHeader("rlpTxn.REQUIRES_EVM_EXECUTION", 1, length),
        new ColumnHeader("rlpTxn.RLP_LT_BYTESIZE", 32, length),
        new ColumnHeader("rlpTxn.RLP_LX_BYTESIZE", 32, length),
        new ColumnHeader("rlpTxn.TYPE", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.absTxNum = buffers.get(0);
    this.absTxNumInfiny = buffers.get(1);
    this.acc1 = buffers.get(2);
    this.acc2 = buffers.get(3);
    this.accBytesize = buffers.get(4);
    this.accessTupleBytesize = buffers.get(5);
    this.addrHi = buffers.get(6);
    this.addrLo = buffers.get(7);
    this.bit = buffers.get(8);
    this.bitAcc = buffers.get(9);
    this.byte1 = buffers.get(10);
    this.byte2 = buffers.get(11);
    this.codeFragmentIndex = buffers.get(12);
    this.counter = buffers.get(13);
    this.dataHi = buffers.get(14);
    this.dataLo = buffers.get(15);
    this.datagascost = buffers.get(16);
    this.depth1 = buffers.get(17);
    this.depth2 = buffers.get(18);
    this.done = buffers.get(19);
    this.indexData = buffers.get(20);
    this.indexLt = buffers.get(21);
    this.indexLx = buffers.get(22);
    this.input1 = buffers.get(23);
    this.input2 = buffers.get(24);
    this.isPrefix = buffers.get(25);
    this.lcCorrection = buffers.get(26);
    this.limb = buffers.get(27);
    this.limbConstructed = buffers.get(28);
    this.lt = buffers.get(29);
    this.lx = buffers.get(30);
    this.nAddr = buffers.get(31);
    this.nBytes = buffers.get(32);
    this.nKeys = buffers.get(33);
    this.nKeysPerAddr = buffers.get(34);
    this.nStep = buffers.get(35);
    this.phase0 = buffers.get(36);
    this.phase1 = buffers.get(37);
    this.phase10 = buffers.get(38);
    this.phase11 = buffers.get(39);
    this.phase12 = buffers.get(40);
    this.phase13 = buffers.get(41);
    this.phase14 = buffers.get(42);
    this.phase2 = buffers.get(43);
    this.phase3 = buffers.get(44);
    this.phase4 = buffers.get(45);
    this.phase5 = buffers.get(46);
    this.phase6 = buffers.get(47);
    this.phase7 = buffers.get(48);
    this.phase8 = buffers.get(49);
    this.phase9 = buffers.get(50);
    this.phaseEnd = buffers.get(51);
    this.phaseSize = buffers.get(52);
    this.power = buffers.get(53);
    this.requiresEvmExecution = buffers.get(54);
    this.rlpLtBytesize = buffers.get(55);
    this.rlpLxBytesize = buffers.get(56);
    this.type = buffers.get(57);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absTxNum(final BigInteger b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM already set");
    } else {
      filled.set(0);
    }

    absTxNum.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace absTxNumInfiny(final BigInteger b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM_INFINY already set");
    } else {
      filled.set(1);
    }

    absTxNumInfiny.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc1(final BigInteger b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlpTxn.ACC_1 already set");
    } else {
      filled.set(3);
    }

    acc1.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace acc2(final BigInteger b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlpTxn.ACC_2 already set");
    } else {
      filled.set(4);
    }

    acc2.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace accBytesize(final BigInteger b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlpTxn.ACC_BYTESIZE already set");
    } else {
      filled.set(5);
    }

    accBytesize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace accessTupleBytesize(final BigInteger b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlpTxn.ACCESS_TUPLE_BYTESIZE already set");
    } else {
      filled.set(2);
    }

    accessTupleBytesize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace addrHi(final BigInteger b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlpTxn.ADDR_HI already set");
    } else {
      filled.set(6);
    }

    addrHi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace addrLo(final BigInteger b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlpTxn.ADDR_LO already set");
    } else {
      filled.set(7);
    }

    addrLo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace bit(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlpTxn.BIT already set");
    } else {
      filled.set(8);
    }

    bit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final BigInteger b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlpTxn.BIT_ACC already set");
    } else {
      filled.set(9);
    }

    bitAcc.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlpTxn.BYTE_1 already set");
    } else {
      filled.set(10);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlpTxn.BYTE_2 already set");
    } else {
      filled.set(11);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace codeFragmentIndex(final BigInteger b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlpTxn.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(12);
    }

    codeFragmentIndex.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace counter(final BigInteger b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlpTxn.COUNTER already set");
    } else {
      filled.set(13);
    }

    counter.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace dataHi(final BigInteger b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlpTxn.DATA_HI already set");
    } else {
      filled.set(15);
    }

    dataHi.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace dataLo(final BigInteger b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlpTxn.DATA_LO already set");
    } else {
      filled.set(16);
    }

    dataLo.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace datagascost(final BigInteger b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlpTxn.DATAGASCOST already set");
    } else {
      filled.set(14);
    }

    datagascost.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace depth1(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlpTxn.DEPTH_1 already set");
    } else {
      filled.set(17);
    }

    depth1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace depth2(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlpTxn.DEPTH_2 already set");
    } else {
      filled.set(18);
    }

    depth2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace done(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlpTxn.DONE already set");
    } else {
      filled.set(19);
    }

    done.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace indexData(final BigInteger b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlpTxn.INDEX_DATA already set");
    } else {
      filled.set(20);
    }

    indexData.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace indexLt(final BigInteger b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlpTxn.INDEX_LT already set");
    } else {
      filled.set(21);
    }

    indexLt.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace indexLx(final BigInteger b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlpTxn.INDEX_LX already set");
    } else {
      filled.set(22);
    }

    indexLx.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace input1(final BigInteger b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlpTxn.INPUT_1 already set");
    } else {
      filled.set(23);
    }

    input1.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace input2(final BigInteger b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlpTxn.INPUT_2 already set");
    } else {
      filled.set(24);
    }

    input2.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace isPrefix(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("rlpTxn.IS_PREFIX already set");
    } else {
      filled.set(25);
    }

    isPrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lcCorrection(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("rlpTxn.LC_CORRECTION already set");
    } else {
      filled.set(26);
    }

    lcCorrection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final BigInteger b) {
    if (filled.get(27)) {
      throw new IllegalStateException("rlpTxn.LIMB already set");
    } else {
      filled.set(27);
    }

    limb.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace limbConstructed(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("rlpTxn.LIMB_CONSTRUCTED already set");
    } else {
      filled.set(28);
    }

    limbConstructed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lt(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("rlpTxn.LT already set");
    } else {
      filled.set(29);
    }

    lt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lx(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("rlpTxn.LX already set");
    } else {
      filled.set(30);
    }

    lx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace nAddr(final BigInteger b) {
    if (filled.get(53)) {
      throw new IllegalStateException("rlpTxn.nADDR already set");
    } else {
      filled.set(53);
    }

    nAddr.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace nBytes(final BigInteger b) {
    if (filled.get(54)) {
      throw new IllegalStateException("rlpTxn.nBYTES already set");
    } else {
      filled.set(54);
    }

    nBytes.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace nKeys(final BigInteger b) {
    if (filled.get(55)) {
      throw new IllegalStateException("rlpTxn.nKEYS already set");
    } else {
      filled.set(55);
    }

    nKeys.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace nKeysPerAddr(final BigInteger b) {
    if (filled.get(56)) {
      throw new IllegalStateException("rlpTxn.nKEYS_PER_ADDR already set");
    } else {
      filled.set(56);
    }

    nKeysPerAddr.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace nStep(final BigInteger b) {
    if (filled.get(57)) {
      throw new IllegalStateException("rlpTxn.nSTEP already set");
    } else {
      filled.set(57);
    }

    nStep.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace phase0(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("rlpTxn.PHASE_0 already set");
    } else {
      filled.set(31);
    }

    phase0.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase1(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("rlpTxn.PHASE_1 already set");
    } else {
      filled.set(32);
    }

    phase1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase10(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("rlpTxn.PHASE_10 already set");
    } else {
      filled.set(33);
    }

    phase10.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase11(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("rlpTxn.PHASE_11 already set");
    } else {
      filled.set(34);
    }

    phase11.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase12(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("rlpTxn.PHASE_12 already set");
    } else {
      filled.set(35);
    }

    phase12.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase13(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("rlpTxn.PHASE_13 already set");
    } else {
      filled.set(36);
    }

    phase13.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase14(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("rlpTxn.PHASE_14 already set");
    } else {
      filled.set(37);
    }

    phase14.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase2(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("rlpTxn.PHASE_2 already set");
    } else {
      filled.set(38);
    }

    phase2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase3(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("rlpTxn.PHASE_3 already set");
    } else {
      filled.set(39);
    }

    phase3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase4(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("rlpTxn.PHASE_4 already set");
    } else {
      filled.set(40);
    }

    phase4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase5(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("rlpTxn.PHASE_5 already set");
    } else {
      filled.set(41);
    }

    phase5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase6(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("rlpTxn.PHASE_6 already set");
    } else {
      filled.set(42);
    }

    phase6.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase7(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("rlpTxn.PHASE_7 already set");
    } else {
      filled.set(43);
    }

    phase7.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase8(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("rlpTxn.PHASE_8 already set");
    } else {
      filled.set(44);
    }

    phase8.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phase9(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("rlpTxn.PHASE_9 already set");
    } else {
      filled.set(45);
    }

    phase9.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseEnd(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("rlpTxn.PHASE_END already set");
    } else {
      filled.set(46);
    }

    phaseEnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseSize(final BigInteger b) {
    if (filled.get(47)) {
      throw new IllegalStateException("rlpTxn.PHASE_SIZE already set");
    } else {
      filled.set(47);
    }

    phaseSize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace power(final BigInteger b) {
    if (filled.get(48)) {
      throw new IllegalStateException("rlpTxn.POWER already set");
    } else {
      filled.set(48);
    }

    power.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace requiresEvmExecution(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("rlpTxn.REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(49);
    }

    requiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rlpLtBytesize(final BigInteger b) {
    if (filled.get(50)) {
      throw new IllegalStateException("rlpTxn.RLP_LT_BYTESIZE already set");
    } else {
      filled.set(50);
    }

    rlpLtBytesize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace rlpLxBytesize(final BigInteger b) {
    if (filled.get(51)) {
      throw new IllegalStateException("rlpTxn.RLP_LX_BYTESIZE already set");
    } else {
      filled.set(51);
    }

    rlpLxBytesize.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace type(final BigInteger b) {
    if (filled.get(52)) {
      throw new IllegalStateException("rlpTxn.TYPE already set");
    } else {
      filled.set(52);
    }

    type.put(UInt256.valueOf(b).toBytes().toArray());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlpTxn.ABS_TX_NUM_INFINY has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlpTxn.ACC_1 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlpTxn.ACC_2 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlpTxn.ACC_BYTESIZE has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlpTxn.ACCESS_TUPLE_BYTESIZE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlpTxn.ADDR_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlpTxn.ADDR_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlpTxn.BIT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlpTxn.BIT_ACC has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlpTxn.BYTE_1 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlpTxn.BYTE_2 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlpTxn.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlpTxn.COUNTER has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlpTxn.DATA_HI has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlpTxn.DATA_LO has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlpTxn.DATAGASCOST has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlpTxn.DEPTH_1 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlpTxn.DEPTH_2 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlpTxn.DONE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlpTxn.INDEX_DATA has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlpTxn.INDEX_LT has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlpTxn.INDEX_LX has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlpTxn.INPUT_1 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlpTxn.INPUT_2 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("rlpTxn.IS_PREFIX has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("rlpTxn.LC_CORRECTION has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("rlpTxn.LIMB has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("rlpTxn.LIMB_CONSTRUCTED has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("rlpTxn.LT has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("rlpTxn.LX has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("rlpTxn.nADDR has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("rlpTxn.nBYTES has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("rlpTxn.nKEYS has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("rlpTxn.nKEYS_PER_ADDR has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("rlpTxn.nSTEP has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("rlpTxn.PHASE_0 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("rlpTxn.PHASE_1 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("rlpTxn.PHASE_10 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("rlpTxn.PHASE_11 has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("rlpTxn.PHASE_12 has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("rlpTxn.PHASE_13 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("rlpTxn.PHASE_14 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("rlpTxn.PHASE_2 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("rlpTxn.PHASE_3 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("rlpTxn.PHASE_4 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("rlpTxn.PHASE_5 has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("rlpTxn.PHASE_6 has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("rlpTxn.PHASE_7 has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("rlpTxn.PHASE_8 has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("rlpTxn.PHASE_9 has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("rlpTxn.PHASE_END has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("rlpTxn.PHASE_SIZE has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("rlpTxn.POWER has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("rlpTxn.REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("rlpTxn.RLP_LT_BYTESIZE has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("rlpTxn.RLP_LX_BYTESIZE has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("rlpTxn.TYPE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absTxNum.position(absTxNum.position() + 32);
    }

    if (!filled.get(1)) {
      absTxNumInfiny.position(absTxNumInfiny.position() + 32);
    }

    if (!filled.get(3)) {
      acc1.position(acc1.position() + 32);
    }

    if (!filled.get(4)) {
      acc2.position(acc2.position() + 32);
    }

    if (!filled.get(5)) {
      accBytesize.position(accBytesize.position() + 32);
    }

    if (!filled.get(2)) {
      accessTupleBytesize.position(accessTupleBytesize.position() + 32);
    }

    if (!filled.get(6)) {
      addrHi.position(addrHi.position() + 32);
    }

    if (!filled.get(7)) {
      addrLo.position(addrLo.position() + 32);
    }

    if (!filled.get(8)) {
      bit.position(bit.position() + 1);
    }

    if (!filled.get(9)) {
      bitAcc.position(bitAcc.position() + 32);
    }

    if (!filled.get(10)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(11)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(12)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 32);
    }

    if (!filled.get(13)) {
      counter.position(counter.position() + 32);
    }

    if (!filled.get(15)) {
      dataHi.position(dataHi.position() + 32);
    }

    if (!filled.get(16)) {
      dataLo.position(dataLo.position() + 32);
    }

    if (!filled.get(14)) {
      datagascost.position(datagascost.position() + 32);
    }

    if (!filled.get(17)) {
      depth1.position(depth1.position() + 1);
    }

    if (!filled.get(18)) {
      depth2.position(depth2.position() + 1);
    }

    if (!filled.get(19)) {
      done.position(done.position() + 1);
    }

    if (!filled.get(20)) {
      indexData.position(indexData.position() + 32);
    }

    if (!filled.get(21)) {
      indexLt.position(indexLt.position() + 32);
    }

    if (!filled.get(22)) {
      indexLx.position(indexLx.position() + 32);
    }

    if (!filled.get(23)) {
      input1.position(input1.position() + 32);
    }

    if (!filled.get(24)) {
      input2.position(input2.position() + 32);
    }

    if (!filled.get(25)) {
      isPrefix.position(isPrefix.position() + 1);
    }

    if (!filled.get(26)) {
      lcCorrection.position(lcCorrection.position() + 1);
    }

    if (!filled.get(27)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(28)) {
      limbConstructed.position(limbConstructed.position() + 1);
    }

    if (!filled.get(29)) {
      lt.position(lt.position() + 1);
    }

    if (!filled.get(30)) {
      lx.position(lx.position() + 1);
    }

    if (!filled.get(53)) {
      nAddr.position(nAddr.position() + 32);
    }

    if (!filled.get(54)) {
      nBytes.position(nBytes.position() + 32);
    }

    if (!filled.get(55)) {
      nKeys.position(nKeys.position() + 32);
    }

    if (!filled.get(56)) {
      nKeysPerAddr.position(nKeysPerAddr.position() + 32);
    }

    if (!filled.get(57)) {
      nStep.position(nStep.position() + 32);
    }

    if (!filled.get(31)) {
      phase0.position(phase0.position() + 1);
    }

    if (!filled.get(32)) {
      phase1.position(phase1.position() + 1);
    }

    if (!filled.get(33)) {
      phase10.position(phase10.position() + 1);
    }

    if (!filled.get(34)) {
      phase11.position(phase11.position() + 1);
    }

    if (!filled.get(35)) {
      phase12.position(phase12.position() + 1);
    }

    if (!filled.get(36)) {
      phase13.position(phase13.position() + 1);
    }

    if (!filled.get(37)) {
      phase14.position(phase14.position() + 1);
    }

    if (!filled.get(38)) {
      phase2.position(phase2.position() + 1);
    }

    if (!filled.get(39)) {
      phase3.position(phase3.position() + 1);
    }

    if (!filled.get(40)) {
      phase4.position(phase4.position() + 1);
    }

    if (!filled.get(41)) {
      phase5.position(phase5.position() + 1);
    }

    if (!filled.get(42)) {
      phase6.position(phase6.position() + 1);
    }

    if (!filled.get(43)) {
      phase7.position(phase7.position() + 1);
    }

    if (!filled.get(44)) {
      phase8.position(phase8.position() + 1);
    }

    if (!filled.get(45)) {
      phase9.position(phase9.position() + 1);
    }

    if (!filled.get(46)) {
      phaseEnd.position(phaseEnd.position() + 1);
    }

    if (!filled.get(47)) {
      phaseSize.position(phaseSize.position() + 32);
    }

    if (!filled.get(48)) {
      power.position(power.position() + 32);
    }

    if (!filled.get(49)) {
      requiresEvmExecution.position(requiresEvmExecution.position() + 1);
    }

    if (!filled.get(50)) {
      rlpLtBytesize.position(rlpLtBytesize.position() + 32);
    }

    if (!filled.get(51)) {
      rlpLxBytesize.position(rlpLxBytesize.position() + 32);
    }

    if (!filled.get(52)) {
      type.position(type.position() + 32);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
    return null;
  }
}
