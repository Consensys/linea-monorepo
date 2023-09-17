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

package net.consensys.linea.zktracer.module.rlptxrcpt;

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
record Trace(
    @JsonProperty("ABS_LOG_NUM") List<BigInteger> absLogNum,
    @JsonProperty("ABS_LOG_NUM_MAX") List<BigInteger> absLogNumMax,
    @JsonProperty("ABS_TX_NUM") List<BigInteger> absTxNum,
    @JsonProperty("ABS_TX_NUM_MAX") List<BigInteger> absTxNumMax,
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ACC_3") List<BigInteger> acc3,
    @JsonProperty("ACC_4") List<BigInteger> acc4,
    @JsonProperty("ACC_SIZE") List<BigInteger> accSize,
    @JsonProperty("BIT") List<Boolean> bit,
    @JsonProperty("BIT_ACC") List<UnsignedByte> bitAcc,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("BYTE_3") List<UnsignedByte> byte3,
    @JsonProperty("BYTE_4") List<UnsignedByte> byte4,
    @JsonProperty("COUNTER") List<UnsignedByte> counter,
    @JsonProperty("DEPTH_1") List<Boolean> depth1,
    @JsonProperty("DONE") List<Boolean> done,
    @JsonProperty("INDEX") List<BigInteger> index,
    @JsonProperty("INDEX_LOCAL") List<BigInteger> indexLocal,
    @JsonProperty("INPUT_1") List<BigInteger> input1,
    @JsonProperty("INPUT_2") List<BigInteger> input2,
    @JsonProperty("INPUT_3") List<BigInteger> input3,
    @JsonProperty("INPUT_4") List<BigInteger> input4,
    @JsonProperty("IS_DATA") List<Boolean> isData,
    @JsonProperty("IS_PREFIX") List<Boolean> isPrefix,
    @JsonProperty("IS_TOPIC") List<Boolean> isTopic,
    @JsonProperty("LC_CORRECTION") List<Boolean> lcCorrection,
    @JsonProperty("LIMB") List<BigInteger> limb,
    @JsonProperty("LIMB_CONSTRUCTED") List<Boolean> limbConstructed,
    @JsonProperty("LOCAL_SIZE") List<BigInteger> localSize,
    @JsonProperty("LOG_ENTRY_SIZE") List<BigInteger> logEntrySize,
    @JsonProperty("nBYTES") List<UnsignedByte> nBytes,
    @JsonProperty("nSTEP") List<UnsignedByte> nStep,
    @JsonProperty("PHASE_0") List<Boolean> phase0,
    @JsonProperty("PHASE_1") List<Boolean> phase1,
    @JsonProperty("PHASE_2") List<Boolean> phase2,
    @JsonProperty("PHASE_3") List<Boolean> phase3,
    @JsonProperty("PHASE_4") List<Boolean> phase4,
    @JsonProperty("PHASE_END") List<Boolean> phaseEnd,
    @JsonProperty("PHASE_SIZE") List<BigInteger> phaseSize,
    @JsonProperty("POWER") List<BigInteger> power,
    @JsonProperty("TXRCPT_SIZE") List<BigInteger> txrcptSize) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ABS_LOG_NUM")
    private final List<BigInteger> absLogNum = new ArrayList<>();

    @JsonProperty("ABS_LOG_NUM_MAX")
    private final List<BigInteger> absLogNumMax = new ArrayList<>();

    @JsonProperty("ABS_TX_NUM")
    private final List<BigInteger> absTxNum = new ArrayList<>();

    @JsonProperty("ABS_TX_NUM_MAX")
    private final List<BigInteger> absTxNumMax = new ArrayList<>();

    @JsonProperty("ACC_1")
    private final List<BigInteger> acc1 = new ArrayList<>();

    @JsonProperty("ACC_2")
    private final List<BigInteger> acc2 = new ArrayList<>();

    @JsonProperty("ACC_3")
    private final List<BigInteger> acc3 = new ArrayList<>();

    @JsonProperty("ACC_4")
    private final List<BigInteger> acc4 = new ArrayList<>();

    @JsonProperty("ACC_SIZE")
    private final List<BigInteger> accSize = new ArrayList<>();

    @JsonProperty("BIT")
    private final List<Boolean> bit = new ArrayList<>();

    @JsonProperty("BIT_ACC")
    private final List<UnsignedByte> bitAcc = new ArrayList<>();

    @JsonProperty("BYTE_1")
    private final List<UnsignedByte> byte1 = new ArrayList<>();

    @JsonProperty("BYTE_2")
    private final List<UnsignedByte> byte2 = new ArrayList<>();

    @JsonProperty("BYTE_3")
    private final List<UnsignedByte> byte3 = new ArrayList<>();

    @JsonProperty("BYTE_4")
    private final List<UnsignedByte> byte4 = new ArrayList<>();

    @JsonProperty("COUNTER")
    private final List<UnsignedByte> counter = new ArrayList<>();

    @JsonProperty("DEPTH_1")
    private final List<Boolean> depth1 = new ArrayList<>();

    @JsonProperty("DONE")
    private final List<Boolean> done = new ArrayList<>();

    @JsonProperty("INDEX")
    private final List<BigInteger> index = new ArrayList<>();

    @JsonProperty("INDEX_LOCAL")
    private final List<BigInteger> indexLocal = new ArrayList<>();

    @JsonProperty("INPUT_1")
    private final List<BigInteger> input1 = new ArrayList<>();

    @JsonProperty("INPUT_2")
    private final List<BigInteger> input2 = new ArrayList<>();

    @JsonProperty("INPUT_3")
    private final List<BigInteger> input3 = new ArrayList<>();

    @JsonProperty("INPUT_4")
    private final List<BigInteger> input4 = new ArrayList<>();

    @JsonProperty("IS_DATA")
    private final List<Boolean> isData = new ArrayList<>();

    @JsonProperty("IS_PREFIX")
    private final List<Boolean> isPrefix = new ArrayList<>();

    @JsonProperty("IS_TOPIC")
    private final List<Boolean> isTopic = new ArrayList<>();

    @JsonProperty("LC_CORRECTION")
    private final List<Boolean> lcCorrection = new ArrayList<>();

    @JsonProperty("LIMB")
    private final List<BigInteger> limb = new ArrayList<>();

    @JsonProperty("LIMB_CONSTRUCTED")
    private final List<Boolean> limbConstructed = new ArrayList<>();

    @JsonProperty("LOCAL_SIZE")
    private final List<BigInteger> localSize = new ArrayList<>();

    @JsonProperty("LOG_ENTRY_SIZE")
    private final List<BigInteger> logEntrySize = new ArrayList<>();

    @JsonProperty("nBYTES")
    private final List<UnsignedByte> nBytes = new ArrayList<>();

    @JsonProperty("nSTEP")
    private final List<UnsignedByte> nStep = new ArrayList<>();

    @JsonProperty("PHASE_0")
    private final List<Boolean> phase0 = new ArrayList<>();

    @JsonProperty("PHASE_1")
    private final List<Boolean> phase1 = new ArrayList<>();

    @JsonProperty("PHASE_2")
    private final List<Boolean> phase2 = new ArrayList<>();

    @JsonProperty("PHASE_3")
    private final List<Boolean> phase3 = new ArrayList<>();

    @JsonProperty("PHASE_4")
    private final List<Boolean> phase4 = new ArrayList<>();

    @JsonProperty("PHASE_END")
    private final List<Boolean> phaseEnd = new ArrayList<>();

    @JsonProperty("PHASE_SIZE")
    private final List<BigInteger> phaseSize = new ArrayList<>();

    @JsonProperty("POWER")
    private final List<BigInteger> power = new ArrayList<>();

    @JsonProperty("TXRCPT_SIZE")
    private final List<BigInteger> txrcptSize = new ArrayList<>();

    private TraceBuilder() {}

    int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.absLogNum.size();
    }

    TraceBuilder absLogNum(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("ABS_LOG_NUM already set");
      } else {
        filled.set(39);
      }

      absLogNum.add(b);

      return this;
    }

    TraceBuilder absLogNumMax(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ABS_LOG_NUM_MAX already set");
      } else {
        filled.set(2);
      }

      absLogNumMax.add(b);

      return this;
    }

    TraceBuilder absTxNum(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("ABS_TX_NUM already set");
      } else {
        filled.set(11);
      }

      absTxNum.add(b);

      return this;
    }

    TraceBuilder absTxNumMax(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ABS_TX_NUM_MAX already set");
      } else {
        filled.set(25);
      }

      absTxNumMax.add(b);

      return this;
    }

    TraceBuilder acc1(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(21);
      }

      acc1.add(b);

      return this;
    }

    TraceBuilder acc2(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(4);
      }

      acc2.add(b);

      return this;
    }

    TraceBuilder acc3(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ACC_3 already set");
      } else {
        filled.set(9);
      }

      acc3.add(b);

      return this;
    }

    TraceBuilder acc4(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("ACC_4 already set");
      } else {
        filled.set(36);
      }

      acc4.add(b);

      return this;
    }

    TraceBuilder accSize(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("ACC_SIZE already set");
      } else {
        filled.set(16);
      }

      accSize.add(b);

      return this;
    }

    TraceBuilder bit(final Boolean b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BIT already set");
      } else {
        filled.set(10);
      }

      bit.add(b);

      return this;
    }

    TraceBuilder bitAcc(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BIT_ACC already set");
      } else {
        filled.set(23);
      }

      bitAcc.add(b);

      return this;
    }

    TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(29)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(29);
      }

      byte1.add(b);

      return this;
    }

    TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(34)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(34);
      }

      byte2.add(b);

      return this;
    }

    TraceBuilder byte3(final UnsignedByte b) {
      if (filled.get(19)) {
        throw new IllegalStateException("BYTE_3 already set");
      } else {
        filled.set(19);
      }

      byte3.add(b);

      return this;
    }

    TraceBuilder byte4(final UnsignedByte b) {
      if (filled.get(22)) {
        throw new IllegalStateException("BYTE_4 already set");
      } else {
        filled.set(22);
      }

      byte4.add(b);

      return this;
    }

    TraceBuilder counter(final UnsignedByte b) {
      if (filled.get(32)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(32);
      }

      counter.add(b);

      return this;
    }

    TraceBuilder depth1(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("DEPTH_1 already set");
      } else {
        filled.set(24);
      }

      depth1.add(b);

      return this;
    }

    TraceBuilder done(final Boolean b) {
      if (filled.get(38)) {
        throw new IllegalStateException("DONE already set");
      } else {
        filled.set(38);
      }

      done.add(b);

      return this;
    }

    TraceBuilder index(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("INDEX already set");
      } else {
        filled.set(27);
      }

      index.add(b);

      return this;
    }

    TraceBuilder indexLocal(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("INDEX_LOCAL already set");
      } else {
        filled.set(37);
      }

      indexLocal.add(b);

      return this;
    }

    TraceBuilder input1(final BigInteger b) {
      if (filled.get(41)) {
        throw new IllegalStateException("INPUT_1 already set");
      } else {
        filled.set(41);
      }

      input1.add(b);

      return this;
    }

    TraceBuilder input2(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("INPUT_2 already set");
      } else {
        filled.set(7);
      }

      input2.add(b);

      return this;
    }

    TraceBuilder input3(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("INPUT_3 already set");
      } else {
        filled.set(3);
      }

      input3.add(b);

      return this;
    }

    TraceBuilder input4(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("INPUT_4 already set");
      } else {
        filled.set(8);
      }

      input4.add(b);

      return this;
    }

    TraceBuilder isData(final Boolean b) {
      if (filled.get(6)) {
        throw new IllegalStateException("IS_DATA already set");
      } else {
        filled.set(6);
      }

      isData.add(b);

      return this;
    }

    TraceBuilder isPrefix(final Boolean b) {
      if (filled.get(0)) {
        throw new IllegalStateException("IS_PREFIX already set");
      } else {
        filled.set(0);
      }

      isPrefix.add(b);

      return this;
    }

    TraceBuilder isTopic(final Boolean b) {
      if (filled.get(5)) {
        throw new IllegalStateException("IS_TOPIC already set");
      } else {
        filled.set(5);
      }

      isTopic.add(b);

      return this;
    }

    TraceBuilder lcCorrection(final Boolean b) {
      if (filled.get(1)) {
        throw new IllegalStateException("LC_CORRECTION already set");
      } else {
        filled.set(1);
      }

      lcCorrection.add(b);

      return this;
    }

    TraceBuilder limb(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("LIMB already set");
      } else {
        filled.set(40);
      }

      limb.add(b);

      return this;
    }

    TraceBuilder limbConstructed(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("LIMB_CONSTRUCTED already set");
      } else {
        filled.set(18);
      }

      limbConstructed.add(b);

      return this;
    }

    TraceBuilder localSize(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("LOCAL_SIZE already set");
      } else {
        filled.set(13);
      }

      localSize.add(b);

      return this;
    }

    TraceBuilder logEntrySize(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("LOG_ENTRY_SIZE already set");
      } else {
        filled.set(28);
      }

      logEntrySize.add(b);

      return this;
    }

    TraceBuilder nBytes(final UnsignedByte b) {
      if (filled.get(12)) {
        throw new IllegalStateException("nBYTES already set");
      } else {
        filled.set(12);
      }

      nBytes.add(b);

      return this;
    }

    TraceBuilder nStep(final UnsignedByte b) {
      if (filled.get(15)) {
        throw new IllegalStateException("nSTEP already set");
      } else {
        filled.set(15);
      }

      nStep.add(b);

      return this;
    }

    TraceBuilder phase0(final Boolean b) {
      if (filled.get(14)) {
        throw new IllegalStateException("PHASE_0 already set");
      } else {
        filled.set(14);
      }

      phase0.add(b);

      return this;
    }

    TraceBuilder phase1(final Boolean b) {
      if (filled.get(17)) {
        throw new IllegalStateException("PHASE_1 already set");
      } else {
        filled.set(17);
      }

      phase1.add(b);

      return this;
    }

    TraceBuilder phase2(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("PHASE_2 already set");
      } else {
        filled.set(35);
      }

      phase2.add(b);

      return this;
    }

    TraceBuilder phase3(final Boolean b) {
      if (filled.get(30)) {
        throw new IllegalStateException("PHASE_3 already set");
      } else {
        filled.set(30);
      }

      phase3.add(b);

      return this;
    }

    TraceBuilder phase4(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("PHASE_4 already set");
      } else {
        filled.set(31);
      }

      phase4.add(b);

      return this;
    }

    TraceBuilder phaseEnd(final Boolean b) {
      if (filled.get(42)) {
        throw new IllegalStateException("PHASE_END already set");
      } else {
        filled.set(42);
      }

      phaseEnd.add(b);

      return this;
    }

    TraceBuilder phaseSize(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("PHASE_SIZE already set");
      } else {
        filled.set(20);
      }

      phaseSize.add(b);

      return this;
    }

    TraceBuilder power(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("POWER already set");
      } else {
        filled.set(26);
      }

      power.add(b);

      return this;
    }

    TraceBuilder txrcptSize(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("TXRCPT_SIZE already set");
      } else {
        filled.set(33);
      }

      txrcptSize.add(b);

      return this;
    }

    TraceBuilder setAbsLogNumAt(final BigInteger b, int i) {
      absLogNum.set(i, b);

      return this;
    }

    TraceBuilder setAbsLogNumMaxAt(final BigInteger b, int i) {
      absLogNumMax.set(i, b);

      return this;
    }

    TraceBuilder setAbsTxNumAt(final BigInteger b, int i) {
      absTxNum.set(i, b);

      return this;
    }

    TraceBuilder setAbsTxNumMaxAt(final BigInteger b, int i) {
      absTxNumMax.set(i, b);

      return this;
    }

    TraceBuilder setAcc1At(final BigInteger b, int i) {
      acc1.set(i, b);

      return this;
    }

    TraceBuilder setAcc2At(final BigInteger b, int i) {
      acc2.set(i, b);

      return this;
    }

    TraceBuilder setAcc3At(final BigInteger b, int i) {
      acc3.set(i, b);

      return this;
    }

    TraceBuilder setAcc4At(final BigInteger b, int i) {
      acc4.set(i, b);

      return this;
    }

    TraceBuilder setAccSizeAt(final BigInteger b, int i) {
      accSize.set(i, b);

      return this;
    }

    TraceBuilder setBitAt(final Boolean b, int i) {
      bit.set(i, b);

      return this;
    }

    TraceBuilder setBitAccAt(final UnsignedByte b, int i) {
      bitAcc.set(i, b);

      return this;
    }

    TraceBuilder setByte1At(final UnsignedByte b, int i) {
      byte1.set(i, b);

      return this;
    }

    TraceBuilder setByte2At(final UnsignedByte b, int i) {
      byte2.set(i, b);

      return this;
    }

    TraceBuilder setByte3At(final UnsignedByte b, int i) {
      byte3.set(i, b);

      return this;
    }

    TraceBuilder setByte4At(final UnsignedByte b, int i) {
      byte4.set(i, b);

      return this;
    }

    TraceBuilder setCounterAt(final UnsignedByte b, int i) {
      counter.set(i, b);

      return this;
    }

    TraceBuilder setDepth1At(final Boolean b, int i) {
      depth1.set(i, b);

      return this;
    }

    TraceBuilder setDoneAt(final Boolean b, int i) {
      done.set(i, b);

      return this;
    }

    TraceBuilder setIndexAt(final BigInteger b, int i) {
      index.set(i, b);

      return this;
    }

    TraceBuilder setIndexLocalAt(final BigInteger b, int i) {
      indexLocal.set(i, b);

      return this;
    }

    TraceBuilder setInput1At(final BigInteger b, int i) {
      input1.set(i, b);

      return this;
    }

    TraceBuilder setInput2At(final BigInteger b, int i) {
      input2.set(i, b);

      return this;
    }

    TraceBuilder setInput3At(final BigInteger b, int i) {
      input3.set(i, b);

      return this;
    }

    TraceBuilder setInput4At(final BigInteger b, int i) {
      input4.set(i, b);

      return this;
    }

    TraceBuilder setIsDataAt(final Boolean b, int i) {
      isData.set(i, b);

      return this;
    }

    TraceBuilder setIsPrefixAt(final Boolean b, int i) {
      isPrefix.set(i, b);

      return this;
    }

    TraceBuilder setIsTopicAt(final Boolean b, int i) {
      isTopic.set(i, b);

      return this;
    }

    TraceBuilder setLcCorrectionAt(final Boolean b, int i) {
      lcCorrection.set(i, b);

      return this;
    }

    TraceBuilder setLimbAt(final BigInteger b, int i) {
      limb.set(i, b);

      return this;
    }

    TraceBuilder setLimbConstructedAt(final Boolean b, int i) {
      limbConstructed.set(i, b);

      return this;
    }

    TraceBuilder setLocalSizeAt(final BigInteger b, int i) {
      localSize.set(i, b);

      return this;
    }

    TraceBuilder setLogEntrySizeAt(final BigInteger b, int i) {
      logEntrySize.set(i, b);

      return this;
    }

    TraceBuilder setNBytesAt(final UnsignedByte b, int i) {
      nBytes.set(i, b);

      return this;
    }

    TraceBuilder setNStepAt(final UnsignedByte b, int i) {
      nStep.set(i, b);

      return this;
    }

    TraceBuilder setPhase0At(final Boolean b, int i) {
      phase0.set(i, b);

      return this;
    }

    TraceBuilder setPhase1At(final Boolean b, int i) {
      phase1.set(i, b);

      return this;
    }

    TraceBuilder setPhase2At(final Boolean b, int i) {
      phase2.set(i, b);

      return this;
    }

    TraceBuilder setPhase3At(final Boolean b, int i) {
      phase3.set(i, b);

      return this;
    }

    TraceBuilder setPhase4At(final Boolean b, int i) {
      phase4.set(i, b);

      return this;
    }

    TraceBuilder setPhaseEndAt(final Boolean b, int i) {
      phaseEnd.set(i, b);

      return this;
    }

    TraceBuilder setPhaseSizeAt(final BigInteger b, int i) {
      phaseSize.set(i, b);

      return this;
    }

    TraceBuilder setPowerAt(final BigInteger b, int i) {
      power.set(i, b);

      return this;
    }

    TraceBuilder setTxrcptSizeAt(final BigInteger b, int i) {
      txrcptSize.set(i, b);

      return this;
    }

    TraceBuilder setAbsLogNumRelative(final BigInteger b, int i) {
      absLogNum.set(absLogNum.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAbsLogNumMaxRelative(final BigInteger b, int i) {
      absLogNumMax.set(absLogNumMax.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAbsTxNumRelative(final BigInteger b, int i) {
      absTxNum.set(absTxNum.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAbsTxNumMaxRelative(final BigInteger b, int i) {
      absTxNumMax.set(absTxNumMax.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc1Relative(final BigInteger b, int i) {
      acc1.set(acc1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc2Relative(final BigInteger b, int i) {
      acc2.set(acc2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc3Relative(final BigInteger b, int i) {
      acc3.set(acc3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc4Relative(final BigInteger b, int i) {
      acc4.set(acc4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccSizeRelative(final BigInteger b, int i) {
      accSize.set(accSize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitRelative(final Boolean b, int i) {
      bit.set(bit.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitAccRelative(final UnsignedByte b, int i) {
      bitAcc.set(bitAcc.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte1Relative(final UnsignedByte b, int i) {
      byte1.set(byte1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte2Relative(final UnsignedByte b, int i) {
      byte2.set(byte2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte3Relative(final UnsignedByte b, int i) {
      byte3.set(byte3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte4Relative(final UnsignedByte b, int i) {
      byte4.set(byte4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterRelative(final UnsignedByte b, int i) {
      counter.set(counter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDepth1Relative(final Boolean b, int i) {
      depth1.set(depth1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDoneRelative(final Boolean b, int i) {
      done.set(done.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIndexRelative(final BigInteger b, int i) {
      index.set(index.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIndexLocalRelative(final BigInteger b, int i) {
      indexLocal.set(indexLocal.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInput1Relative(final BigInteger b, int i) {
      input1.set(input1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInput2Relative(final BigInteger b, int i) {
      input2.set(input2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInput3Relative(final BigInteger b, int i) {
      input3.set(input3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInput4Relative(final BigInteger b, int i) {
      input4.set(input4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsDataRelative(final Boolean b, int i) {
      isData.set(isData.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsPrefixRelative(final Boolean b, int i) {
      isPrefix.set(isPrefix.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsTopicRelative(final Boolean b, int i) {
      isTopic.set(isTopic.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLcCorrectionRelative(final Boolean b, int i) {
      lcCorrection.set(lcCorrection.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLimbRelative(final BigInteger b, int i) {
      limb.set(limb.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLimbConstructedRelative(final Boolean b, int i) {
      limbConstructed.set(limbConstructed.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLocalSizeRelative(final BigInteger b, int i) {
      localSize.set(localSize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLogEntrySizeRelative(final BigInteger b, int i) {
      logEntrySize.set(logEntrySize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNBytesRelative(final UnsignedByte b, int i) {
      nBytes.set(nBytes.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNStepRelative(final UnsignedByte b, int i) {
      nStep.set(nStep.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase0Relative(final Boolean b, int i) {
      phase0.set(phase0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase1Relative(final Boolean b, int i) {
      phase1.set(phase1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase2Relative(final Boolean b, int i) {
      phase2.set(phase2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase3Relative(final Boolean b, int i) {
      phase3.set(phase3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase4Relative(final Boolean b, int i) {
      phase4.set(phase4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhaseEndRelative(final Boolean b, int i) {
      phaseEnd.set(phaseEnd.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhaseSizeRelative(final BigInteger b, int i) {
      phaseSize.set(phaseSize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPowerRelative(final BigInteger b, int i) {
      power.set(power.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTxrcptSizeRelative(final BigInteger b, int i) {
      txrcptSize.set(txrcptSize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(39)) {
        throw new IllegalStateException("ABS_LOG_NUM has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ABS_LOG_NUM_MAX has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("ABS_TX_NUM has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ABS_TX_NUM_MAX has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ACC_3 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("ACC_4 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("ACC_SIZE has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BIT has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BIT_ACC has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("BYTE_3 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("BYTE_4 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("DEPTH_1 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("DONE has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("INDEX has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("INDEX_LOCAL has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("INPUT_1 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("INPUT_2 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("INPUT_3 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("INPUT_4 has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("IS_DATA has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("IS_PREFIX has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("IS_TOPIC has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("LC_CORRECTION has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("LIMB has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("LIMB_CONSTRUCTED has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("LOCAL_SIZE has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("LOG_ENTRY_SIZE has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("nBYTES has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("nSTEP has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("PHASE_0 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("PHASE_1 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("PHASE_2 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("PHASE_3 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("PHASE_4 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("PHASE_END has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("PHASE_SIZE has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("POWER has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("TXRCPT_SIZE has not been filled");
      }

      filled.clear();

      return this;
    }

    TraceBuilder fillAndValidateRow() {
      if (!filled.get(39)) {
        absLogNum.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(2)) {
        absLogNumMax.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(11)) {
        absTxNum.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(25)) {
        absTxNumMax.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(21)) {
        acc1.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(4)) {
        acc2.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(9)) {
        acc3.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(36)) {
        acc4.add(BigInteger.ZERO);
        this.filled.set(36);
      }
      if (!filled.get(16)) {
        accSize.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(10)) {
        bit.add(false);
        this.filled.set(10);
      }
      if (!filled.get(23)) {
        bitAcc.add(UnsignedByte.of(0));
        this.filled.set(23);
      }
      if (!filled.get(29)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(29);
      }
      if (!filled.get(34)) {
        byte2.add(UnsignedByte.of(0));
        this.filled.set(34);
      }
      if (!filled.get(19)) {
        byte3.add(UnsignedByte.of(0));
        this.filled.set(19);
      }
      if (!filled.get(22)) {
        byte4.add(UnsignedByte.of(0));
        this.filled.set(22);
      }
      if (!filled.get(32)) {
        counter.add(UnsignedByte.of(0));
        this.filled.set(32);
      }
      if (!filled.get(24)) {
        depth1.add(false);
        this.filled.set(24);
      }
      if (!filled.get(38)) {
        done.add(false);
        this.filled.set(38);
      }
      if (!filled.get(27)) {
        index.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(37)) {
        indexLocal.add(BigInteger.ZERO);
        this.filled.set(37);
      }
      if (!filled.get(41)) {
        input1.add(BigInteger.ZERO);
        this.filled.set(41);
      }
      if (!filled.get(7)) {
        input2.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(3)) {
        input3.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(8)) {
        input4.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(6)) {
        isData.add(false);
        this.filled.set(6);
      }
      if (!filled.get(0)) {
        isPrefix.add(false);
        this.filled.set(0);
      }
      if (!filled.get(5)) {
        isTopic.add(false);
        this.filled.set(5);
      }
      if (!filled.get(1)) {
        lcCorrection.add(false);
        this.filled.set(1);
      }
      if (!filled.get(40)) {
        limb.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(18)) {
        limbConstructed.add(false);
        this.filled.set(18);
      }
      if (!filled.get(13)) {
        localSize.add(BigInteger.ZERO);
        this.filled.set(13);
      }
      if (!filled.get(28)) {
        logEntrySize.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(12)) {
        nBytes.add(UnsignedByte.of(0));
        this.filled.set(12);
      }
      if (!filled.get(15)) {
        nStep.add(UnsignedByte.of(0));
        this.filled.set(15);
      }
      if (!filled.get(14)) {
        phase0.add(false);
        this.filled.set(14);
      }
      if (!filled.get(17)) {
        phase1.add(false);
        this.filled.set(17);
      }
      if (!filled.get(35)) {
        phase2.add(false);
        this.filled.set(35);
      }
      if (!filled.get(30)) {
        phase3.add(false);
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        phase4.add(false);
        this.filled.set(31);
      }
      if (!filled.get(42)) {
        phaseEnd.add(false);
        this.filled.set(42);
      }
      if (!filled.get(20)) {
        phaseSize.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(26)) {
        power.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(33)) {
        txrcptSize.add(BigInteger.ZERO);
        this.filled.set(33);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          absLogNum,
          absLogNumMax,
          absTxNum,
          absTxNumMax,
          acc1,
          acc2,
          acc3,
          acc4,
          accSize,
          bit,
          bitAcc,
          byte1,
          byte2,
          byte3,
          byte4,
          counter,
          depth1,
          done,
          index,
          indexLocal,
          input1,
          input2,
          input3,
          input4,
          isData,
          isPrefix,
          isTopic,
          lcCorrection,
          limb,
          limbConstructed,
          localSize,
          logEntrySize,
          nBytes,
          nStep,
          phase0,
          phase1,
          phase2,
          phase3,
          phase4,
          phaseEnd,
          phaseSize,
          power,
          txrcptSize);
    }
  }
}
