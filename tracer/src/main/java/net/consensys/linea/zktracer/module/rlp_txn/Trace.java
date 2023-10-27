/*
 * Copyright ConsenSys AG.
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
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public record Trace(
    @JsonProperty("ABS_TX_NUM") List<BigInteger> absTxNum,
    @JsonProperty("ABS_TX_NUM_INFINY") List<BigInteger> absTxNumInfiny,
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ACC_BYTESIZE") List<BigInteger> accBytesize,
    @JsonProperty("ACCESS_TUPLE_BYTESIZE") List<BigInteger> accessTupleBytesize,
    @JsonProperty("ADDR_HI") List<BigInteger> addrHi,
    @JsonProperty("ADDR_LO") List<BigInteger> addrLo,
    @JsonProperty("BIT") List<Boolean> bit,
    @JsonProperty("BIT_ACC") List<BigInteger> bitAcc,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("CODE_FRAGMENT_INDEX") List<BigInteger> codeFragmentIndex,
    @JsonProperty("COUNTER") List<BigInteger> counter,
    @JsonProperty("DATA_HI") List<BigInteger> dataHi,
    @JsonProperty("DATA_LO") List<BigInteger> dataLo,
    @JsonProperty("DATAGASCOST") List<BigInteger> datagascost,
    @JsonProperty("DEPTH_1") List<Boolean> depth1,
    @JsonProperty("DEPTH_2") List<Boolean> depth2,
    @JsonProperty("DONE") List<Boolean> done,
    @JsonProperty("INDEX_DATA") List<BigInteger> indexData,
    @JsonProperty("INDEX_LT") List<BigInteger> indexLt,
    @JsonProperty("INDEX_LX") List<BigInteger> indexLx,
    @JsonProperty("INPUT_1") List<BigInteger> input1,
    @JsonProperty("INPUT_2") List<BigInteger> input2,
    @JsonProperty("IS_PREFIX") List<Boolean> isPrefix,
    @JsonProperty("LC_CORRECTION") List<Boolean> lcCorrection,
    @JsonProperty("LIMB") List<BigInteger> limb,
    @JsonProperty("LIMB_CONSTRUCTED") List<Boolean> limbConstructed,
    @JsonProperty("LT") List<Boolean> lt,
    @JsonProperty("LX") List<Boolean> lx,
    @JsonProperty("nADDR") List<BigInteger> nAddr,
    @JsonProperty("nBYTES") List<BigInteger> nBytes,
    @JsonProperty("nKEYS") List<BigInteger> nKeys,
    @JsonProperty("nKEYS_PER_ADDR") List<BigInteger> nKeysPerAddr,
    @JsonProperty("nSTEP") List<BigInteger> nStep,
    @JsonProperty("PHASE_0") List<Boolean> phase0,
    @JsonProperty("PHASE_1") List<Boolean> phase1,
    @JsonProperty("PHASE_10") List<Boolean> phase10,
    @JsonProperty("PHASE_11") List<Boolean> phase11,
    @JsonProperty("PHASE_12") List<Boolean> phase12,
    @JsonProperty("PHASE_13") List<Boolean> phase13,
    @JsonProperty("PHASE_14") List<Boolean> phase14,
    @JsonProperty("PHASE_2") List<Boolean> phase2,
    @JsonProperty("PHASE_3") List<Boolean> phase3,
    @JsonProperty("PHASE_4") List<Boolean> phase4,
    @JsonProperty("PHASE_5") List<Boolean> phase5,
    @JsonProperty("PHASE_6") List<Boolean> phase6,
    @JsonProperty("PHASE_7") List<Boolean> phase7,
    @JsonProperty("PHASE_8") List<Boolean> phase8,
    @JsonProperty("PHASE_9") List<Boolean> phase9,
    @JsonProperty("PHASE_END") List<Boolean> phaseEnd,
    @JsonProperty("PHASE_SIZE") List<BigInteger> phaseSize,
    @JsonProperty("POWER") List<BigInteger> power,
    @JsonProperty("REQUIRES_EVM_EXECUTION") List<Boolean> requiresEvmExecution,
    @JsonProperty("RLP_LT_BYTESIZE") List<BigInteger> rlpLtBytesize,
    @JsonProperty("RLP_LX_BYTESIZE") List<BigInteger> rlpLxBytesize,
    @JsonProperty("TYPE") List<BigInteger> type) {
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.absTxNum.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ABS_TX_NUM")
    private final List<BigInteger> absTxNum;

    @JsonProperty("ABS_TX_NUM_INFINY")
    private final List<BigInteger> absTxNumInfiny;

    @JsonProperty("ACC_1")
    private final List<BigInteger> acc1;

    @JsonProperty("ACC_2")
    private final List<BigInteger> acc2;

    @JsonProperty("ACC_BYTESIZE")
    private final List<BigInteger> accBytesize;

    @JsonProperty("ACCESS_TUPLE_BYTESIZE")
    private final List<BigInteger> accessTupleBytesize;

    @JsonProperty("ADDR_HI")
    private final List<BigInteger> addrHi;

    @JsonProperty("ADDR_LO")
    private final List<BigInteger> addrLo;

    @JsonProperty("BIT")
    private final List<Boolean> bit;

    @JsonProperty("BIT_ACC")
    private final List<BigInteger> bitAcc;

    @JsonProperty("BYTE_1")
    private final List<UnsignedByte> byte1;

    @JsonProperty("BYTE_2")
    private final List<UnsignedByte> byte2;

    @JsonProperty("CODE_FRAGMENT_INDEX")
    private final List<BigInteger> codeFragmentIndex;

    @JsonProperty("COUNTER")
    private final List<BigInteger> counter;

    @JsonProperty("DATA_HI")
    private final List<BigInteger> dataHi;

    @JsonProperty("DATA_LO")
    private final List<BigInteger> dataLo;

    @JsonProperty("DATAGASCOST")
    private final List<BigInteger> datagascost;

    @JsonProperty("DEPTH_1")
    private final List<Boolean> depth1;

    @JsonProperty("DEPTH_2")
    private final List<Boolean> depth2;

    @JsonProperty("DONE")
    private final List<Boolean> done;

    @JsonProperty("INDEX_DATA")
    private final List<BigInteger> indexData;

    @JsonProperty("INDEX_LT")
    private final List<BigInteger> indexLt;

    @JsonProperty("INDEX_LX")
    private final List<BigInteger> indexLx;

    @JsonProperty("INPUT_1")
    private final List<BigInteger> input1;

    @JsonProperty("INPUT_2")
    private final List<BigInteger> input2;

    @JsonProperty("IS_PREFIX")
    private final List<Boolean> isPrefix;

    @JsonProperty("LC_CORRECTION")
    private final List<Boolean> lcCorrection;

    @JsonProperty("LIMB")
    private final List<BigInteger> limb;

    @JsonProperty("LIMB_CONSTRUCTED")
    private final List<Boolean> limbConstructed;

    @JsonProperty("LT")
    private final List<Boolean> lt;

    @JsonProperty("LX")
    private final List<Boolean> lx;

    @JsonProperty("nADDR")
    private final List<BigInteger> nAddr;

    @JsonProperty("nBYTES")
    private final List<BigInteger> nBytes;

    @JsonProperty("nKEYS")
    private final List<BigInteger> nKeys;

    @JsonProperty("nKEYS_PER_ADDR")
    private final List<BigInteger> nKeysPerAddr;

    @JsonProperty("nSTEP")
    private final List<BigInteger> nStep;

    @JsonProperty("PHASE_0")
    private final List<Boolean> phase0;

    @JsonProperty("PHASE_1")
    private final List<Boolean> phase1;

    @JsonProperty("PHASE_10")
    private final List<Boolean> phase10;

    @JsonProperty("PHASE_11")
    private final List<Boolean> phase11;

    @JsonProperty("PHASE_12")
    private final List<Boolean> phase12;

    @JsonProperty("PHASE_13")
    private final List<Boolean> phase13;

    @JsonProperty("PHASE_14")
    private final List<Boolean> phase14;

    @JsonProperty("PHASE_2")
    private final List<Boolean> phase2;

    @JsonProperty("PHASE_3")
    private final List<Boolean> phase3;

    @JsonProperty("PHASE_4")
    private final List<Boolean> phase4;

    @JsonProperty("PHASE_5")
    private final List<Boolean> phase5;

    @JsonProperty("PHASE_6")
    private final List<Boolean> phase6;

    @JsonProperty("PHASE_7")
    private final List<Boolean> phase7;

    @JsonProperty("PHASE_8")
    private final List<Boolean> phase8;

    @JsonProperty("PHASE_9")
    private final List<Boolean> phase9;

    @JsonProperty("PHASE_END")
    private final List<Boolean> phaseEnd;

    @JsonProperty("PHASE_SIZE")
    private final List<BigInteger> phaseSize;

    @JsonProperty("POWER")
    private final List<BigInteger> power;

    @JsonProperty("REQUIRES_EVM_EXECUTION")
    private final List<Boolean> requiresEvmExecution;

    @JsonProperty("RLP_LT_BYTESIZE")
    private final List<BigInteger> rlpLtBytesize;

    @JsonProperty("RLP_LX_BYTESIZE")
    private final List<BigInteger> rlpLxBytesize;

    @JsonProperty("TYPE")
    private final List<BigInteger> type;

    private TraceBuilder(int length) {
      this.absTxNum = new ArrayList<>(length);
      this.absTxNumInfiny = new ArrayList<>(length);
      this.acc1 = new ArrayList<>(length);
      this.acc2 = new ArrayList<>(length);
      this.accBytesize = new ArrayList<>(length);
      this.accessTupleBytesize = new ArrayList<>(length);
      this.addrHi = new ArrayList<>(length);
      this.addrLo = new ArrayList<>(length);
      this.bit = new ArrayList<>(length);
      this.bitAcc = new ArrayList<>(length);
      this.byte1 = new ArrayList<>(length);
      this.byte2 = new ArrayList<>(length);
      this.codeFragmentIndex = new ArrayList<>(length);
      this.counter = new ArrayList<>(length);
      this.dataHi = new ArrayList<>(length);
      this.dataLo = new ArrayList<>(length);
      this.datagascost = new ArrayList<>(length);
      this.depth1 = new ArrayList<>(length);
      this.depth2 = new ArrayList<>(length);
      this.done = new ArrayList<>(length);
      this.indexData = new ArrayList<>(length);
      this.indexLt = new ArrayList<>(length);
      this.indexLx = new ArrayList<>(length);
      this.input1 = new ArrayList<>(length);
      this.input2 = new ArrayList<>(length);
      this.isPrefix = new ArrayList<>(length);
      this.lcCorrection = new ArrayList<>(length);
      this.limb = new ArrayList<>(length);
      this.limbConstructed = new ArrayList<>(length);
      this.lt = new ArrayList<>(length);
      this.lx = new ArrayList<>(length);
      this.nAddr = new ArrayList<>(length);
      this.nBytes = new ArrayList<>(length);
      this.nKeys = new ArrayList<>(length);
      this.nKeysPerAddr = new ArrayList<>(length);
      this.nStep = new ArrayList<>(length);
      this.phase0 = new ArrayList<>(length);
      this.phase1 = new ArrayList<>(length);
      this.phase10 = new ArrayList<>(length);
      this.phase11 = new ArrayList<>(length);
      this.phase12 = new ArrayList<>(length);
      this.phase13 = new ArrayList<>(length);
      this.phase14 = new ArrayList<>(length);
      this.phase2 = new ArrayList<>(length);
      this.phase3 = new ArrayList<>(length);
      this.phase4 = new ArrayList<>(length);
      this.phase5 = new ArrayList<>(length);
      this.phase6 = new ArrayList<>(length);
      this.phase7 = new ArrayList<>(length);
      this.phase8 = new ArrayList<>(length);
      this.phase9 = new ArrayList<>(length);
      this.phaseEnd = new ArrayList<>(length);
      this.phaseSize = new ArrayList<>(length);
      this.power = new ArrayList<>(length);
      this.requiresEvmExecution = new ArrayList<>(length);
      this.rlpLtBytesize = new ArrayList<>(length);
      this.rlpLxBytesize = new ArrayList<>(length);
      this.type = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.absTxNum.size();
    }

    public TraceBuilder absTxNum(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ABS_TX_NUM already set");
      } else {
        filled.set(0);
      }

      absTxNum.add(b);

      return this;
    }

    public TraceBuilder absTxNumInfiny(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ABS_TX_NUM_INFINY already set");
      } else {
        filled.set(1);
      }

      absTxNumInfiny.add(b);

      return this;
    }

    public TraceBuilder acc1(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(3);
      }

      acc1.add(b);

      return this;
    }

    public TraceBuilder acc2(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(4);
      }

      acc2.add(b);

      return this;
    }

    public TraceBuilder accBytesize(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("ACC_BYTESIZE already set");
      } else {
        filled.set(5);
      }

      accBytesize.add(b);

      return this;
    }

    public TraceBuilder accessTupleBytesize(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ACCESS_TUPLE_BYTESIZE already set");
      } else {
        filled.set(2);
      }

      accessTupleBytesize.add(b);

      return this;
    }

    public TraceBuilder addrHi(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ADDR_HI already set");
      } else {
        filled.set(6);
      }

      addrHi.add(b);

      return this;
    }

    public TraceBuilder addrLo(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ADDR_LO already set");
      } else {
        filled.set(7);
      }

      addrLo.add(b);

      return this;
    }

    public TraceBuilder bit(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("BIT already set");
      } else {
        filled.set(8);
      }

      bit.add(b);

      return this;
    }

    public TraceBuilder bitAcc(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("BIT_ACC already set");
      } else {
        filled.set(9);
      }

      bitAcc.add(b);

      return this;
    }

    public TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(10);
      }

      byte1.add(b);

      return this;
    }

    public TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(11)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(11);
      }

      byte2.add(b);

      return this;
    }

    public TraceBuilder codeFragmentIndex(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX already set");
      } else {
        filled.set(12);
      }

      codeFragmentIndex.add(b);

      return this;
    }

    public TraceBuilder counter(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(13);
      }

      counter.add(b);

      return this;
    }

    public TraceBuilder dataHi(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("DATA_HI already set");
      } else {
        filled.set(15);
      }

      dataHi.add(b);

      return this;
    }

    public TraceBuilder dataLo(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("DATA_LO already set");
      } else {
        filled.set(16);
      }

      dataLo.add(b);

      return this;
    }

    public TraceBuilder datagascost(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("DATAGASCOST already set");
      } else {
        filled.set(14);
      }

      datagascost.add(b);

      return this;
    }

    public TraceBuilder depth1(final Boolean b) {
      if (filled.get(17)) {
        throw new IllegalStateException("DEPTH_1 already set");
      } else {
        filled.set(17);
      }

      depth1.add(b);

      return this;
    }

    public TraceBuilder depth2(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("DEPTH_2 already set");
      } else {
        filled.set(18);
      }

      depth2.add(b);

      return this;
    }

    public TraceBuilder done(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("DONE already set");
      } else {
        filled.set(19);
      }

      done.add(b);

      return this;
    }

    public TraceBuilder indexData(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("INDEX_DATA already set");
      } else {
        filled.set(20);
      }

      indexData.add(b);

      return this;
    }

    public TraceBuilder indexLt(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("INDEX_LT already set");
      } else {
        filled.set(21);
      }

      indexLt.add(b);

      return this;
    }

    public TraceBuilder indexLx(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("INDEX_LX already set");
      } else {
        filled.set(22);
      }

      indexLx.add(b);

      return this;
    }

    public TraceBuilder input1(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("INPUT_1 already set");
      } else {
        filled.set(23);
      }

      input1.add(b);

      return this;
    }

    public TraceBuilder input2(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("INPUT_2 already set");
      } else {
        filled.set(24);
      }

      input2.add(b);

      return this;
    }

    public TraceBuilder isPrefix(final Boolean b) {
      if (filled.get(25)) {
        throw new IllegalStateException("IS_PREFIX already set");
      } else {
        filled.set(25);
      }

      isPrefix.add(b);

      return this;
    }

    public TraceBuilder lcCorrection(final Boolean b) {
      if (filled.get(26)) {
        throw new IllegalStateException("LC_CORRECTION already set");
      } else {
        filled.set(26);
      }

      lcCorrection.add(b);

      return this;
    }

    public TraceBuilder limb(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("LIMB already set");
      } else {
        filled.set(27);
      }

      limb.add(b);

      return this;
    }

    public TraceBuilder limbConstructed(final Boolean b) {
      if (filled.get(28)) {
        throw new IllegalStateException("LIMB_CONSTRUCTED already set");
      } else {
        filled.set(28);
      }

      limbConstructed.add(b);

      return this;
    }

    public TraceBuilder lt(final Boolean b) {
      if (filled.get(29)) {
        throw new IllegalStateException("LT already set");
      } else {
        filled.set(29);
      }

      lt.add(b);

      return this;
    }

    public TraceBuilder lx(final Boolean b) {
      if (filled.get(30)) {
        throw new IllegalStateException("LX already set");
      } else {
        filled.set(30);
      }

      lx.add(b);

      return this;
    }

    public TraceBuilder nAddr(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("nADDR already set");
      } else {
        filled.set(53);
      }

      nAddr.add(b);

      return this;
    }

    public TraceBuilder nBytes(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("nBYTES already set");
      } else {
        filled.set(54);
      }

      nBytes.add(b);

      return this;
    }

    public TraceBuilder nKeys(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("nKEYS already set");
      } else {
        filled.set(55);
      }

      nKeys.add(b);

      return this;
    }

    public TraceBuilder nKeysPerAddr(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("nKEYS_PER_ADDR already set");
      } else {
        filled.set(56);
      }

      nKeysPerAddr.add(b);

      return this;
    }

    public TraceBuilder nStep(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("nSTEP already set");
      } else {
        filled.set(57);
      }

      nStep.add(b);

      return this;
    }

    public TraceBuilder phase0(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("PHASE_0 already set");
      } else {
        filled.set(31);
      }

      phase0.add(b);

      return this;
    }

    public TraceBuilder phase1(final Boolean b) {
      if (filled.get(32)) {
        throw new IllegalStateException("PHASE_1 already set");
      } else {
        filled.set(32);
      }

      phase1.add(b);

      return this;
    }

    public TraceBuilder phase10(final Boolean b) {
      if (filled.get(33)) {
        throw new IllegalStateException("PHASE_10 already set");
      } else {
        filled.set(33);
      }

      phase10.add(b);

      return this;
    }

    public TraceBuilder phase11(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("PHASE_11 already set");
      } else {
        filled.set(34);
      }

      phase11.add(b);

      return this;
    }

    public TraceBuilder phase12(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("PHASE_12 already set");
      } else {
        filled.set(35);
      }

      phase12.add(b);

      return this;
    }

    public TraceBuilder phase13(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("PHASE_13 already set");
      } else {
        filled.set(36);
      }

      phase13.add(b);

      return this;
    }

    public TraceBuilder phase14(final Boolean b) {
      if (filled.get(37)) {
        throw new IllegalStateException("PHASE_14 already set");
      } else {
        filled.set(37);
      }

      phase14.add(b);

      return this;
    }

    public TraceBuilder phase2(final Boolean b) {
      if (filled.get(38)) {
        throw new IllegalStateException("PHASE_2 already set");
      } else {
        filled.set(38);
      }

      phase2.add(b);

      return this;
    }

    public TraceBuilder phase3(final Boolean b) {
      if (filled.get(39)) {
        throw new IllegalStateException("PHASE_3 already set");
      } else {
        filled.set(39);
      }

      phase3.add(b);

      return this;
    }

    public TraceBuilder phase4(final Boolean b) {
      if (filled.get(40)) {
        throw new IllegalStateException("PHASE_4 already set");
      } else {
        filled.set(40);
      }

      phase4.add(b);

      return this;
    }

    public TraceBuilder phase5(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("PHASE_5 already set");
      } else {
        filled.set(41);
      }

      phase5.add(b);

      return this;
    }

    public TraceBuilder phase6(final Boolean b) {
      if (filled.get(42)) {
        throw new IllegalStateException("PHASE_6 already set");
      } else {
        filled.set(42);
      }

      phase6.add(b);

      return this;
    }

    public TraceBuilder phase7(final Boolean b) {
      if (filled.get(43)) {
        throw new IllegalStateException("PHASE_7 already set");
      } else {
        filled.set(43);
      }

      phase7.add(b);

      return this;
    }

    public TraceBuilder phase8(final Boolean b) {
      if (filled.get(44)) {
        throw new IllegalStateException("PHASE_8 already set");
      } else {
        filled.set(44);
      }

      phase8.add(b);

      return this;
    }

    public TraceBuilder phase9(final Boolean b) {
      if (filled.get(45)) {
        throw new IllegalStateException("PHASE_9 already set");
      } else {
        filled.set(45);
      }

      phase9.add(b);

      return this;
    }

    public TraceBuilder phaseEnd(final Boolean b) {
      if (filled.get(46)) {
        throw new IllegalStateException("PHASE_END already set");
      } else {
        filled.set(46);
      }

      phaseEnd.add(b);

      return this;
    }

    public TraceBuilder phaseSize(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("PHASE_SIZE already set");
      } else {
        filled.set(47);
      }

      phaseSize.add(b);

      return this;
    }

    public TraceBuilder power(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("POWER already set");
      } else {
        filled.set(48);
      }

      power.add(b);

      return this;
    }

    public TraceBuilder requiresEvmExecution(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("REQUIRES_EVM_EXECUTION already set");
      } else {
        filled.set(49);
      }

      requiresEvmExecution.add(b);

      return this;
    }

    public TraceBuilder rlpLtBytesize(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("RLP_LT_BYTESIZE already set");
      } else {
        filled.set(50);
      }

      rlpLtBytesize.add(b);

      return this;
    }

    public TraceBuilder rlpLxBytesize(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("RLP_LX_BYTESIZE already set");
      } else {
        filled.set(51);
      }

      rlpLxBytesize.add(b);

      return this;
    }

    public TraceBuilder type(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("TYPE already set");
      } else {
        filled.set(52);
      }

      type.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ABS_TX_NUM has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ABS_TX_NUM_INFINY has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("ACC_BYTESIZE has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ACCESS_TUPLE_BYTESIZE has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("ADDR_HI has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ADDR_LO has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("BIT has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("BIT_ACC has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("DATA_HI has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("DATA_LO has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("DATAGASCOST has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("DEPTH_1 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("DEPTH_2 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("DONE has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("INDEX_DATA has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("INDEX_LT has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("INDEX_LX has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("INPUT_1 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("INPUT_2 has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("IS_PREFIX has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("LC_CORRECTION has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("LIMB has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("LIMB_CONSTRUCTED has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("LT has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("LX has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("nADDR has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("nBYTES has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("nKEYS has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("nKEYS_PER_ADDR has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("nSTEP has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("PHASE_0 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("PHASE_1 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("PHASE_10 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("PHASE_11 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("PHASE_12 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("PHASE_13 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("PHASE_14 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("PHASE_2 has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("PHASE_3 has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("PHASE_4 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("PHASE_5 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("PHASE_6 has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("PHASE_7 has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("PHASE_8 has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("PHASE_9 has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("PHASE_END has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("PHASE_SIZE has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("POWER has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("REQUIRES_EVM_EXECUTION has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("RLP_LT_BYTESIZE has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("RLP_LX_BYTESIZE has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("TYPE has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        absTxNum.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        absTxNumInfiny.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(3)) {
        acc1.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        acc2.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        accBytesize.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(2)) {
        accessTupleBytesize.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(6)) {
        addrHi.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        addrLo.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        bit.add(false);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        bitAcc.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        byte2.add(UnsignedByte.of(0));
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        codeFragmentIndex.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        counter.add(BigInteger.ZERO);
        this.filled.set(13);
      }
      if (!filled.get(15)) {
        dataHi.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        dataLo.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(14)) {
        datagascost.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(17)) {
        depth1.add(false);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        depth2.add(false);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        done.add(false);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        indexData.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        indexLt.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        indexLx.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        input1.add(BigInteger.ZERO);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        input2.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        isPrefix.add(false);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        lcCorrection.add(false);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        limb.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        limbConstructed.add(false);
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        lt.add(false);
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        lx.add(false);
        this.filled.set(30);
      }
      if (!filled.get(53)) {
        nAddr.add(BigInteger.ZERO);
        this.filled.set(53);
      }
      if (!filled.get(54)) {
        nBytes.add(BigInteger.ZERO);
        this.filled.set(54);
      }
      if (!filled.get(55)) {
        nKeys.add(BigInteger.ZERO);
        this.filled.set(55);
      }
      if (!filled.get(56)) {
        nKeysPerAddr.add(BigInteger.ZERO);
        this.filled.set(56);
      }
      if (!filled.get(57)) {
        nStep.add(BigInteger.ZERO);
        this.filled.set(57);
      }
      if (!filled.get(31)) {
        phase0.add(false);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        phase1.add(false);
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        phase10.add(false);
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        phase11.add(false);
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        phase12.add(false);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        phase13.add(false);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        phase14.add(false);
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        phase2.add(false);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        phase3.add(false);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        phase4.add(false);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        phase5.add(false);
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        phase6.add(false);
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        phase7.add(false);
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        phase8.add(false);
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        phase9.add(false);
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        phaseEnd.add(false);
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        phaseSize.add(BigInteger.ZERO);
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        power.add(BigInteger.ZERO);
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        requiresEvmExecution.add(false);
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        rlpLtBytesize.add(BigInteger.ZERO);
        this.filled.set(50);
      }
      if (!filled.get(51)) {
        rlpLxBytesize.add(BigInteger.ZERO);
        this.filled.set(51);
      }
      if (!filled.get(52)) {
        type.add(BigInteger.ZERO);
        this.filled.set(52);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          absTxNum,
          absTxNumInfiny,
          acc1,
          acc2,
          accBytesize,
          accessTupleBytesize,
          addrHi,
          addrLo,
          bit,
          bitAcc,
          byte1,
          byte2,
          codeFragmentIndex,
          counter,
          dataHi,
          dataLo,
          datagascost,
          depth1,
          depth2,
          done,
          indexData,
          indexLt,
          indexLx,
          input1,
          input2,
          isPrefix,
          lcCorrection,
          limb,
          limbConstructed,
          lt,
          lx,
          nAddr,
          nBytes,
          nKeys,
          nKeysPerAddr,
          nStep,
          phase0,
          phase1,
          phase10,
          phase11,
          phase12,
          phase13,
          phase14,
          phase2,
          phase3,
          phase4,
          phase5,
          phase6,
          phase7,
          phase8,
          phase9,
          phaseEnd,
          phaseSize,
          power,
          requiresEvmExecution,
          rlpLtBytesize,
          rlpLxBytesize,
          type);
    }
  }
}
