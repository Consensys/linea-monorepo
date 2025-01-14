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

package net.consensys.linea.zktracer.module.rlptxn;

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
  private final MappedByteBuffer dataGasCost;
  private final MappedByteBuffer dataHi;
  private final MappedByteBuffer dataLo;
  private final MappedByteBuffer depth1;
  private final MappedByteBuffer depth2;
  private final MappedByteBuffer done;
  private final MappedByteBuffer indexData;
  private final MappedByteBuffer indexLt;
  private final MappedByteBuffer indexLx;
  private final MappedByteBuffer input1;
  private final MappedByteBuffer input2;
  private final MappedByteBuffer isPhaseAccessList;
  private final MappedByteBuffer isPhaseBeta;
  private final MappedByteBuffer isPhaseChainId;
  private final MappedByteBuffer isPhaseData;
  private final MappedByteBuffer isPhaseGasLimit;
  private final MappedByteBuffer isPhaseGasPrice;
  private final MappedByteBuffer isPhaseMaxFeePerGas;
  private final MappedByteBuffer isPhaseMaxPriorityFeePerGas;
  private final MappedByteBuffer isPhaseNonce;
  private final MappedByteBuffer isPhaseR;
  private final MappedByteBuffer isPhaseRlpPrefix;
  private final MappedByteBuffer isPhaseS;
  private final MappedByteBuffer isPhaseTo;
  private final MappedByteBuffer isPhaseValue;
  private final MappedByteBuffer isPhaseY;
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
  private final MappedByteBuffer phase;
  private final MappedByteBuffer phaseEnd;
  private final MappedByteBuffer phaseSize;
  private final MappedByteBuffer power;
  private final MappedByteBuffer requiresEvmExecution;
  private final MappedByteBuffer rlpLtBytesize;
  private final MappedByteBuffer rlpLxBytesize;
  private final MappedByteBuffer toHashByProver;
  private final MappedByteBuffer type;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("rlptxn.ABS_TX_NUM", 2, length));
      headers.add(new ColumnHeader("rlptxn.ABS_TX_NUM_INFINY", 2, length));
      headers.add(new ColumnHeader("rlptxn.ACC_1", 16, length));
      headers.add(new ColumnHeader("rlptxn.ACC_2", 16, length));
      headers.add(new ColumnHeader("rlptxn.ACC_BYTESIZE", 1, length));
      headers.add(new ColumnHeader("rlptxn.ACCESS_TUPLE_BYTESIZE", 3, length));
      headers.add(new ColumnHeader("rlptxn.ADDR_HI", 4, length));
      headers.add(new ColumnHeader("rlptxn.ADDR_LO", 16, length));
      headers.add(new ColumnHeader("rlptxn.BIT", 1, length));
      headers.add(new ColumnHeader("rlptxn.BIT_ACC", 1, length));
      headers.add(new ColumnHeader("rlptxn.BYTE_1", 1, length));
      headers.add(new ColumnHeader("rlptxn.BYTE_2", 1, length));
      headers.add(new ColumnHeader("rlptxn.CODE_FRAGMENT_INDEX", 4, length));
      headers.add(new ColumnHeader("rlptxn.COUNTER", 1, length));
      headers.add(new ColumnHeader("rlptxn.DATA_GAS_COST", 4, length));
      headers.add(new ColumnHeader("rlptxn.DATA_HI", 16, length));
      headers.add(new ColumnHeader("rlptxn.DATA_LO", 16, length));
      headers.add(new ColumnHeader("rlptxn.DEPTH_1", 1, length));
      headers.add(new ColumnHeader("rlptxn.DEPTH_2", 1, length));
      headers.add(new ColumnHeader("rlptxn.DONE", 1, length));
      headers.add(new ColumnHeader("rlptxn.INDEX_DATA", 4, length));
      headers.add(new ColumnHeader("rlptxn.INDEX_LT", 4, length));
      headers.add(new ColumnHeader("rlptxn.INDEX_LX", 4, length));
      headers.add(new ColumnHeader("rlptxn.INPUT_1", 16, length));
      headers.add(new ColumnHeader("rlptxn.INPUT_2", 16, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_ACCESS_LIST", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_BETA", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_CHAIN_ID", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_DATA", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_GAS_LIMIT", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_GAS_PRICE", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_MAX_FEE_PER_GAS", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_MAX_PRIORITY_FEE_PER_GAS", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_NONCE", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_R", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_RLP_PREFIX", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_S", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_TO", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_VALUE", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PHASE_Y", 1, length));
      headers.add(new ColumnHeader("rlptxn.IS_PREFIX", 1, length));
      headers.add(new ColumnHeader("rlptxn.LC_CORRECTION", 1, length));
      headers.add(new ColumnHeader("rlptxn.LIMB", 16, length));
      headers.add(new ColumnHeader("rlptxn.LIMB_CONSTRUCTED", 1, length));
      headers.add(new ColumnHeader("rlptxn.LT", 1, length));
      headers.add(new ColumnHeader("rlptxn.LX", 1, length));
      headers.add(new ColumnHeader("rlptxn.nADDR", 2, length));
      headers.add(new ColumnHeader("rlptxn.nBYTES", 1, length));
      headers.add(new ColumnHeader("rlptxn.nKEYS", 2, length));
      headers.add(new ColumnHeader("rlptxn.nKEYS_PER_ADDR", 2, length));
      headers.add(new ColumnHeader("rlptxn.nSTEP", 1, length));
      headers.add(new ColumnHeader("rlptxn.PHASE", 1, length));
      headers.add(new ColumnHeader("rlptxn.PHASE_END", 1, length));
      headers.add(new ColumnHeader("rlptxn.PHASE_SIZE", 4, length));
      headers.add(new ColumnHeader("rlptxn.POWER", 16, length));
      headers.add(new ColumnHeader("rlptxn.REQUIRES_EVM_EXECUTION", 1, length));
      headers.add(new ColumnHeader("rlptxn.RLP_LT_BYTESIZE", 3, length));
      headers.add(new ColumnHeader("rlptxn.RLP_LX_BYTESIZE", 3, length));
      headers.add(new ColumnHeader("rlptxn.TO_HASH_BY_PROVER", 1, length));
      headers.add(new ColumnHeader("rlptxn.TYPE", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
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
    this.dataGasCost = buffers.get(14);
    this.dataHi = buffers.get(15);
    this.dataLo = buffers.get(16);
    this.depth1 = buffers.get(17);
    this.depth2 = buffers.get(18);
    this.done = buffers.get(19);
    this.indexData = buffers.get(20);
    this.indexLt = buffers.get(21);
    this.indexLx = buffers.get(22);
    this.input1 = buffers.get(23);
    this.input2 = buffers.get(24);
    this.isPhaseAccessList = buffers.get(25);
    this.isPhaseBeta = buffers.get(26);
    this.isPhaseChainId = buffers.get(27);
    this.isPhaseData = buffers.get(28);
    this.isPhaseGasLimit = buffers.get(29);
    this.isPhaseGasPrice = buffers.get(30);
    this.isPhaseMaxFeePerGas = buffers.get(31);
    this.isPhaseMaxPriorityFeePerGas = buffers.get(32);
    this.isPhaseNonce = buffers.get(33);
    this.isPhaseR = buffers.get(34);
    this.isPhaseRlpPrefix = buffers.get(35);
    this.isPhaseS = buffers.get(36);
    this.isPhaseTo = buffers.get(37);
    this.isPhaseValue = buffers.get(38);
    this.isPhaseY = buffers.get(39);
    this.isPrefix = buffers.get(40);
    this.lcCorrection = buffers.get(41);
    this.limb = buffers.get(42);
    this.limbConstructed = buffers.get(43);
    this.lt = buffers.get(44);
    this.lx = buffers.get(45);
    this.nAddr = buffers.get(46);
    this.nBytes = buffers.get(47);
    this.nKeys = buffers.get(48);
    this.nKeysPerAddr = buffers.get(49);
    this.nStep = buffers.get(50);
    this.phase = buffers.get(51);
    this.phaseEnd = buffers.get(52);
    this.phaseSize = buffers.get(53);
    this.power = buffers.get(54);
    this.requiresEvmExecution = buffers.get(55);
    this.rlpLtBytesize = buffers.get(56);
    this.rlpLxBytesize = buffers.get(57);
    this.toHashByProver = buffers.get(58);
    this.type = buffers.get(59);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace absTxNum(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlptxn.ABS_TX_NUM already set");
    } else {
      filled.set(0);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("rlptxn.ABS_TX_NUM has invalid value (" + b + ")"); }
    absTxNum.put((byte) (b >> 8));
    absTxNum.put((byte) b);


    return this;
  }

  public Trace absTxNumInfiny(final long b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlptxn.ABS_TX_NUM_INFINY already set");
    } else {
      filled.set(1);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("rlptxn.ABS_TX_NUM_INFINY has invalid value (" + b + ")"); }
    absTxNumInfiny.put((byte) (b >> 8));
    absTxNumInfiny.put((byte) b);


    return this;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlptxn.ACC_1 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.ACC_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc1.put(bs.get(j)); }

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlptxn.ACC_2 already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.ACC_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc2.put(bs.get(j)); }

    return this;
  }

  public Trace accBytesize(final long b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlptxn.ACC_BYTESIZE already set");
    } else {
      filled.set(5);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxn.ACC_BYTESIZE has invalid value (" + b + ")"); }
    accBytesize.put((byte) b);


    return this;
  }

  public Trace accessTupleBytesize(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlptxn.ACCESS_TUPLE_BYTESIZE already set");
    } else {
      filled.set(2);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("rlptxn.ACCESS_TUPLE_BYTESIZE has invalid value (" + b + ")"); }
    accessTupleBytesize.put((byte) (b >> 16));
    accessTupleBytesize.put((byte) (b >> 8));
    accessTupleBytesize.put((byte) b);


    return this;
  }

  public Trace addrHi(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlptxn.ADDR_HI already set");
    } else {
      filled.set(6);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.ADDR_HI has invalid value (" + b + ")"); }
    addrHi.put((byte) (b >> 24));
    addrHi.put((byte) (b >> 16));
    addrHi.put((byte) (b >> 8));
    addrHi.put((byte) b);


    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlptxn.ADDR_LO already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.ADDR_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addrLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addrLo.put(bs.get(j)); }

    return this;
  }

  public Trace bit(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlptxn.BIT already set");
    } else {
      filled.set(8);
    }

    bit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final UnsignedByte b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlptxn.BIT_ACC already set");
    } else {
      filled.set(9);
    }

    bitAcc.put(b.toByte());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlptxn.BYTE_1 already set");
    } else {
      filled.set(10);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace byte2(final UnsignedByte b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlptxn.BYTE_2 already set");
    } else {
      filled.set(11);
    }

    byte2.put(b.toByte());

    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlptxn.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(12);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeFragmentIndex.put((byte) (b >> 24));
    codeFragmentIndex.put((byte) (b >> 16));
    codeFragmentIndex.put((byte) (b >> 8));
    codeFragmentIndex.put((byte) b);


    return this;
  }

  public Trace counter(final long b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlptxn.COUNTER already set");
    } else {
      filled.set(13);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxn.COUNTER has invalid value (" + b + ")"); }
    counter.put((byte) b);


    return this;
  }

  public Trace dataGasCost(final long b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlptxn.DATA_GAS_COST already set");
    } else {
      filled.set(14);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.DATA_GAS_COST has invalid value (" + b + ")"); }
    dataGasCost.put((byte) (b >> 24));
    dataGasCost.put((byte) (b >> 16));
    dataGasCost.put((byte) (b >> 8));
    dataGasCost.put((byte) b);


    return this;
  }

  public Trace dataHi(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlptxn.DATA_HI already set");
    } else {
      filled.set(15);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.DATA_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { dataHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { dataHi.put(bs.get(j)); }

    return this;
  }

  public Trace dataLo(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlptxn.DATA_LO already set");
    } else {
      filled.set(16);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.DATA_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { dataLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { dataLo.put(bs.get(j)); }

    return this;
  }

  public Trace depth1(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlptxn.DEPTH_1 already set");
    } else {
      filled.set(17);
    }

    depth1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace depth2(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlptxn.DEPTH_2 already set");
    } else {
      filled.set(18);
    }

    depth2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace done(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlptxn.DONE already set");
    } else {
      filled.set(19);
    }

    done.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace indexData(final long b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlptxn.INDEX_DATA already set");
    } else {
      filled.set(20);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.INDEX_DATA has invalid value (" + b + ")"); }
    indexData.put((byte) (b >> 24));
    indexData.put((byte) (b >> 16));
    indexData.put((byte) (b >> 8));
    indexData.put((byte) b);


    return this;
  }

  public Trace indexLt(final long b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlptxn.INDEX_LT already set");
    } else {
      filled.set(21);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.INDEX_LT has invalid value (" + b + ")"); }
    indexLt.put((byte) (b >> 24));
    indexLt.put((byte) (b >> 16));
    indexLt.put((byte) (b >> 8));
    indexLt.put((byte) b);


    return this;
  }

  public Trace indexLx(final long b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlptxn.INDEX_LX already set");
    } else {
      filled.set(22);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.INDEX_LX has invalid value (" + b + ")"); }
    indexLx.put((byte) (b >> 24));
    indexLx.put((byte) (b >> 16));
    indexLx.put((byte) (b >> 8));
    indexLx.put((byte) b);


    return this;
  }

  public Trace input1(final Bytes b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlptxn.INPUT_1 already set");
    } else {
      filled.set(23);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.INPUT_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { input1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { input1.put(bs.get(j)); }

    return this;
  }

  public Trace input2(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlptxn.INPUT_2 already set");
    } else {
      filled.set(24);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.INPUT_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { input2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { input2.put(bs.get(j)); }

    return this;
  }

  public Trace isPhaseAccessList(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_ACCESS_LIST already set");
    } else {
      filled.set(25);
    }

    isPhaseAccessList.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseBeta(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_BETA already set");
    } else {
      filled.set(26);
    }

    isPhaseBeta.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseChainId(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_CHAIN_ID already set");
    } else {
      filled.set(27);
    }

    isPhaseChainId.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseData(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_DATA already set");
    } else {
      filled.set(28);
    }

    isPhaseData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseGasLimit(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_GAS_LIMIT already set");
    } else {
      filled.set(29);
    }

    isPhaseGasLimit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseGasPrice(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_GAS_PRICE already set");
    } else {
      filled.set(30);
    }

    isPhaseGasPrice.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseMaxFeePerGas(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_MAX_FEE_PER_GAS already set");
    } else {
      filled.set(31);
    }

    isPhaseMaxFeePerGas.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseMaxPriorityFeePerGas(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_MAX_PRIORITY_FEE_PER_GAS already set");
    } else {
      filled.set(32);
    }

    isPhaseMaxPriorityFeePerGas.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseNonce(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_NONCE already set");
    } else {
      filled.set(33);
    }

    isPhaseNonce.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseR(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_R already set");
    } else {
      filled.set(34);
    }

    isPhaseR.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseRlpPrefix(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_RLP_PREFIX already set");
    } else {
      filled.set(35);
    }

    isPhaseRlpPrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseS(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_S already set");
    } else {
      filled.set(36);
    }

    isPhaseS.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseTo(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_TO already set");
    } else {
      filled.set(37);
    }

    isPhaseTo.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseValue(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_VALUE already set");
    } else {
      filled.set(38);
    }

    isPhaseValue.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPhaseY(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_Y already set");
    } else {
      filled.set(39);
    }

    isPhaseY.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPrefix(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("rlptxn.IS_PREFIX already set");
    } else {
      filled.set(40);
    }

    isPrefix.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lcCorrection(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("rlptxn.LC_CORRECTION already set");
    } else {
      filled.set(41);
    }

    lcCorrection.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(42)) {
      throw new IllegalStateException("rlptxn.LIMB already set");
    } else {
      filled.set(42);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

    return this;
  }

  public Trace limbConstructed(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("rlptxn.LIMB_CONSTRUCTED already set");
    } else {
      filled.set(43);
    }

    limbConstructed.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lt(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("rlptxn.LT already set");
    } else {
      filled.set(44);
    }

    lt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace lx(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("rlptxn.LX already set");
    } else {
      filled.set(45);
    }

    lx.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace nAddr(final long b) {
    if (filled.get(55)) {
      throw new IllegalStateException("rlptxn.nADDR already set");
    } else {
      filled.set(55);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("rlptxn.nADDR has invalid value (" + b + ")"); }
    nAddr.put((byte) (b >> 8));
    nAddr.put((byte) b);


    return this;
  }

  public Trace nBytes(final long b) {
    if (filled.get(56)) {
      throw new IllegalStateException("rlptxn.nBYTES already set");
    } else {
      filled.set(56);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxn.nBYTES has invalid value (" + b + ")"); }
    nBytes.put((byte) b);


    return this;
  }

  public Trace nKeys(final long b) {
    if (filled.get(57)) {
      throw new IllegalStateException("rlptxn.nKEYS already set");
    } else {
      filled.set(57);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("rlptxn.nKEYS has invalid value (" + b + ")"); }
    nKeys.put((byte) (b >> 8));
    nKeys.put((byte) b);


    return this;
  }

  public Trace nKeysPerAddr(final long b) {
    if (filled.get(58)) {
      throw new IllegalStateException("rlptxn.nKEYS_PER_ADDR already set");
    } else {
      filled.set(58);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("rlptxn.nKEYS_PER_ADDR has invalid value (" + b + ")"); }
    nKeysPerAddr.put((byte) (b >> 8));
    nKeysPerAddr.put((byte) b);


    return this;
  }

  public Trace nStep(final long b) {
    if (filled.get(59)) {
      throw new IllegalStateException("rlptxn.nSTEP already set");
    } else {
      filled.set(59);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxn.nSTEP has invalid value (" + b + ")"); }
    nStep.put((byte) b);


    return this;
  }

  public Trace phase(final long b) {
    if (filled.get(46)) {
      throw new IllegalStateException("rlptxn.PHASE already set");
    } else {
      filled.set(46);
    }

    if(b >= 32L) { throw new IllegalArgumentException("rlptxn.PHASE has invalid value (" + b + ")"); }
    phase.put((byte) b);


    return this;
  }

  public Trace phaseEnd(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("rlptxn.PHASE_END already set");
    } else {
      filled.set(47);
    }

    phaseEnd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace phaseSize(final long b) {
    if (filled.get(48)) {
      throw new IllegalStateException("rlptxn.PHASE_SIZE already set");
    } else {
      filled.set(48);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlptxn.PHASE_SIZE has invalid value (" + b + ")"); }
    phaseSize.put((byte) (b >> 24));
    phaseSize.put((byte) (b >> 16));
    phaseSize.put((byte) (b >> 8));
    phaseSize.put((byte) b);


    return this;
  }

  public Trace power(final Bytes b) {
    if (filled.get(49)) {
      throw new IllegalStateException("rlptxn.POWER already set");
    } else {
      filled.set(49);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlptxn.POWER has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { power.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { power.put(bs.get(j)); }

    return this;
  }

  public Trace requiresEvmExecution(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("rlptxn.REQUIRES_EVM_EXECUTION already set");
    } else {
      filled.set(50);
    }

    requiresEvmExecution.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace rlpLtBytesize(final long b) {
    if (filled.get(51)) {
      throw new IllegalStateException("rlptxn.RLP_LT_BYTESIZE already set");
    } else {
      filled.set(51);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("rlptxn.RLP_LT_BYTESIZE has invalid value (" + b + ")"); }
    rlpLtBytesize.put((byte) (b >> 16));
    rlpLtBytesize.put((byte) (b >> 8));
    rlpLtBytesize.put((byte) b);


    return this;
  }

  public Trace rlpLxBytesize(final long b) {
    if (filled.get(52)) {
      throw new IllegalStateException("rlptxn.RLP_LX_BYTESIZE already set");
    } else {
      filled.set(52);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("rlptxn.RLP_LX_BYTESIZE has invalid value (" + b + ")"); }
    rlpLxBytesize.put((byte) (b >> 16));
    rlpLxBytesize.put((byte) (b >> 8));
    rlpLxBytesize.put((byte) b);


    return this;
  }

  public Trace toHashByProver(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("rlptxn.TO_HASH_BY_PROVER already set");
    } else {
      filled.set(53);
    }

    toHashByProver.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace type(final long b) {
    if (filled.get(54)) {
      throw new IllegalStateException("rlptxn.TYPE already set");
    } else {
      filled.set(54);
    }

    if(b >= 8L) { throw new IllegalArgumentException("rlptxn.TYPE has invalid value (" + b + ")"); }
    type.put((byte) b);


    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlptxn.ABS_TX_NUM has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlptxn.ABS_TX_NUM_INFINY has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlptxn.ACC_1 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlptxn.ACC_2 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlptxn.ACC_BYTESIZE has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlptxn.ACCESS_TUPLE_BYTESIZE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlptxn.ADDR_HI has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlptxn.ADDR_LO has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlptxn.BIT has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlptxn.BIT_ACC has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlptxn.BYTE_1 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlptxn.BYTE_2 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlptxn.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlptxn.COUNTER has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlptxn.DATA_GAS_COST has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlptxn.DATA_HI has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlptxn.DATA_LO has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlptxn.DEPTH_1 has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlptxn.DEPTH_2 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlptxn.DONE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlptxn.INDEX_DATA has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlptxn.INDEX_LT has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlptxn.INDEX_LX has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlptxn.INPUT_1 has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlptxn.INPUT_2 has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_ACCESS_LIST has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_BETA has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_CHAIN_ID has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_DATA has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_GAS_LIMIT has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_GAS_PRICE has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_MAX_FEE_PER_GAS has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_MAX_PRIORITY_FEE_PER_GAS has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_NONCE has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_R has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_RLP_PREFIX has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_S has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_TO has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_VALUE has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("rlptxn.IS_PHASE_Y has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("rlptxn.IS_PREFIX has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("rlptxn.LC_CORRECTION has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("rlptxn.LIMB has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("rlptxn.LIMB_CONSTRUCTED has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("rlptxn.LT has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("rlptxn.LX has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("rlptxn.nADDR has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("rlptxn.nBYTES has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("rlptxn.nKEYS has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("rlptxn.nKEYS_PER_ADDR has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("rlptxn.nSTEP has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("rlptxn.PHASE has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("rlptxn.PHASE_END has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("rlptxn.PHASE_SIZE has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("rlptxn.POWER has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("rlptxn.REQUIRES_EVM_EXECUTION has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("rlptxn.RLP_LT_BYTESIZE has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("rlptxn.RLP_LX_BYTESIZE has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("rlptxn.TO_HASH_BY_PROVER has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("rlptxn.TYPE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      absTxNum.position(absTxNum.position() + 2);
    }

    if (!filled.get(1)) {
      absTxNumInfiny.position(absTxNumInfiny.position() + 2);
    }

    if (!filled.get(3)) {
      acc1.position(acc1.position() + 16);
    }

    if (!filled.get(4)) {
      acc2.position(acc2.position() + 16);
    }

    if (!filled.get(5)) {
      accBytesize.position(accBytesize.position() + 1);
    }

    if (!filled.get(2)) {
      accessTupleBytesize.position(accessTupleBytesize.position() + 3);
    }

    if (!filled.get(6)) {
      addrHi.position(addrHi.position() + 4);
    }

    if (!filled.get(7)) {
      addrLo.position(addrLo.position() + 16);
    }

    if (!filled.get(8)) {
      bit.position(bit.position() + 1);
    }

    if (!filled.get(9)) {
      bitAcc.position(bitAcc.position() + 1);
    }

    if (!filled.get(10)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(11)) {
      byte2.position(byte2.position() + 1);
    }

    if (!filled.get(12)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 4);
    }

    if (!filled.get(13)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(14)) {
      dataGasCost.position(dataGasCost.position() + 4);
    }

    if (!filled.get(15)) {
      dataHi.position(dataHi.position() + 16);
    }

    if (!filled.get(16)) {
      dataLo.position(dataLo.position() + 16);
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
      indexData.position(indexData.position() + 4);
    }

    if (!filled.get(21)) {
      indexLt.position(indexLt.position() + 4);
    }

    if (!filled.get(22)) {
      indexLx.position(indexLx.position() + 4);
    }

    if (!filled.get(23)) {
      input1.position(input1.position() + 16);
    }

    if (!filled.get(24)) {
      input2.position(input2.position() + 16);
    }

    if (!filled.get(25)) {
      isPhaseAccessList.position(isPhaseAccessList.position() + 1);
    }

    if (!filled.get(26)) {
      isPhaseBeta.position(isPhaseBeta.position() + 1);
    }

    if (!filled.get(27)) {
      isPhaseChainId.position(isPhaseChainId.position() + 1);
    }

    if (!filled.get(28)) {
      isPhaseData.position(isPhaseData.position() + 1);
    }

    if (!filled.get(29)) {
      isPhaseGasLimit.position(isPhaseGasLimit.position() + 1);
    }

    if (!filled.get(30)) {
      isPhaseGasPrice.position(isPhaseGasPrice.position() + 1);
    }

    if (!filled.get(31)) {
      isPhaseMaxFeePerGas.position(isPhaseMaxFeePerGas.position() + 1);
    }

    if (!filled.get(32)) {
      isPhaseMaxPriorityFeePerGas.position(isPhaseMaxPriorityFeePerGas.position() + 1);
    }

    if (!filled.get(33)) {
      isPhaseNonce.position(isPhaseNonce.position() + 1);
    }

    if (!filled.get(34)) {
      isPhaseR.position(isPhaseR.position() + 1);
    }

    if (!filled.get(35)) {
      isPhaseRlpPrefix.position(isPhaseRlpPrefix.position() + 1);
    }

    if (!filled.get(36)) {
      isPhaseS.position(isPhaseS.position() + 1);
    }

    if (!filled.get(37)) {
      isPhaseTo.position(isPhaseTo.position() + 1);
    }

    if (!filled.get(38)) {
      isPhaseValue.position(isPhaseValue.position() + 1);
    }

    if (!filled.get(39)) {
      isPhaseY.position(isPhaseY.position() + 1);
    }

    if (!filled.get(40)) {
      isPrefix.position(isPrefix.position() + 1);
    }

    if (!filled.get(41)) {
      lcCorrection.position(lcCorrection.position() + 1);
    }

    if (!filled.get(42)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(43)) {
      limbConstructed.position(limbConstructed.position() + 1);
    }

    if (!filled.get(44)) {
      lt.position(lt.position() + 1);
    }

    if (!filled.get(45)) {
      lx.position(lx.position() + 1);
    }

    if (!filled.get(55)) {
      nAddr.position(nAddr.position() + 2);
    }

    if (!filled.get(56)) {
      nBytes.position(nBytes.position() + 1);
    }

    if (!filled.get(57)) {
      nKeys.position(nKeys.position() + 2);
    }

    if (!filled.get(58)) {
      nKeysPerAddr.position(nKeysPerAddr.position() + 2);
    }

    if (!filled.get(59)) {
      nStep.position(nStep.position() + 1);
    }

    if (!filled.get(46)) {
      phase.position(phase.position() + 1);
    }

    if (!filled.get(47)) {
      phaseEnd.position(phaseEnd.position() + 1);
    }

    if (!filled.get(48)) {
      phaseSize.position(phaseSize.position() + 4);
    }

    if (!filled.get(49)) {
      power.position(power.position() + 16);
    }

    if (!filled.get(50)) {
      requiresEvmExecution.position(requiresEvmExecution.position() + 1);
    }

    if (!filled.get(51)) {
      rlpLtBytesize.position(rlpLtBytesize.position() + 3);
    }

    if (!filled.get(52)) {
      rlpLxBytesize.position(rlpLxBytesize.position() + 3);
    }

    if (!filled.get(53)) {
      toHashByProver.position(toHashByProver.position() + 1);
    }

    if (!filled.get(54)) {
      type.position(type.position() + 1);
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
