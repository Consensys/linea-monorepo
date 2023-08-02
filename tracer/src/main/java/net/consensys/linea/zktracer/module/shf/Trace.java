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

package net.consensys.linea.zktracer.module.shf;

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
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ACC_3") List<BigInteger> acc3,
    @JsonProperty("ACC_4") List<BigInteger> acc4,
    @JsonProperty("ACC_5") List<BigInteger> acc5,
    @JsonProperty("ARG_1_HI") List<BigInteger> arg1Hi,
    @JsonProperty("ARG_1_LO") List<BigInteger> arg1Lo,
    @JsonProperty("ARG_2_HI") List<BigInteger> arg2Hi,
    @JsonProperty("ARG_2_LO") List<BigInteger> arg2Lo,
    @JsonProperty("BIT_1") List<Boolean> bit1,
    @JsonProperty("BIT_2") List<Boolean> bit2,
    @JsonProperty("BIT_3") List<Boolean> bit3,
    @JsonProperty("BIT_4") List<Boolean> bit4,
    @JsonProperty("BIT_B_3") List<Boolean> bitB3,
    @JsonProperty("BIT_B_4") List<Boolean> bitB4,
    @JsonProperty("BIT_B_5") List<Boolean> bitB5,
    @JsonProperty("BIT_B_6") List<Boolean> bitB6,
    @JsonProperty("BIT_B_7") List<Boolean> bitB7,
    @JsonProperty("BITS") List<Boolean> bits,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("BYTE_3") List<UnsignedByte> byte3,
    @JsonProperty("BYTE_4") List<UnsignedByte> byte4,
    @JsonProperty("BYTE_5") List<UnsignedByte> byte5,
    @JsonProperty("COUNTER") List<BigInteger> counter,
    @JsonProperty("INST") List<BigInteger> inst,
    @JsonProperty("IS_DATA") List<Boolean> isData,
    @JsonProperty("KNOWN") List<Boolean> known,
    @JsonProperty("LEFT_ALIGNED_SUFFIX_HIGH") List<BigInteger> leftAlignedSuffixHigh,
    @JsonProperty("LEFT_ALIGNED_SUFFIX_LOW") List<BigInteger> leftAlignedSuffixLow,
    @JsonProperty("LOW_3") List<BigInteger> low3,
    @JsonProperty("MICRO_SHIFT_PARAMETER") List<BigInteger> microShiftParameter,
    @JsonProperty("NEG") List<Boolean> neg,
    @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> oneLineInstruction,
    @JsonProperty("ONES") List<BigInteger> ones,
    @JsonProperty("RES_HI") List<BigInteger> resHi,
    @JsonProperty("RES_LO") List<BigInteger> resLo,
    @JsonProperty("RIGHT_ALIGNED_PREFIX_HIGH") List<BigInteger> rightAlignedPrefixHigh,
    @JsonProperty("RIGHT_ALIGNED_PREFIX_LOW") List<BigInteger> rightAlignedPrefixLow,
    @JsonProperty("SHB_3_HI") List<BigInteger> shb3Hi,
    @JsonProperty("SHB_3_LO") List<BigInteger> shb3Lo,
    @JsonProperty("SHB_4_HI") List<BigInteger> shb4Hi,
    @JsonProperty("SHB_4_LO") List<BigInteger> shb4Lo,
    @JsonProperty("SHB_5_HI") List<BigInteger> shb5Hi,
    @JsonProperty("SHB_5_LO") List<BigInteger> shb5Lo,
    @JsonProperty("SHB_6_HI") List<BigInteger> shb6Hi,
    @JsonProperty("SHB_6_LO") List<BigInteger> shb6Lo,
    @JsonProperty("SHB_7_HI") List<BigInteger> shb7Hi,
    @JsonProperty("SHB_7_LO") List<BigInteger> shb7Lo,
    @JsonProperty("SHIFT_DIRECTION") List<Boolean> shiftDirection,
    @JsonProperty("SHIFT_STAMP") List<BigInteger> shiftStamp) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    private final List<BigInteger> acc1 = new ArrayList<>();
    private final List<BigInteger> acc2 = new ArrayList<>();
    private final List<BigInteger> acc3 = new ArrayList<>();
    private final List<BigInteger> acc4 = new ArrayList<>();
    private final List<BigInteger> acc5 = new ArrayList<>();
    private final List<BigInteger> arg1Hi = new ArrayList<>();
    private final List<BigInteger> arg1Lo = new ArrayList<>();
    private final List<BigInteger> arg2Hi = new ArrayList<>();
    private final List<BigInteger> arg2Lo = new ArrayList<>();
    private final List<Boolean> bit1 = new ArrayList<>();
    private final List<Boolean> bit2 = new ArrayList<>();
    private final List<Boolean> bit3 = new ArrayList<>();
    private final List<Boolean> bit4 = new ArrayList<>();
    private final List<Boolean> bitB3 = new ArrayList<>();
    private final List<Boolean> bitB4 = new ArrayList<>();
    private final List<Boolean> bitB5 = new ArrayList<>();
    private final List<Boolean> bitB6 = new ArrayList<>();
    private final List<Boolean> bitB7 = new ArrayList<>();
    private final List<Boolean> bits = new ArrayList<>();
    private final List<UnsignedByte> byte1 = new ArrayList<>();
    private final List<UnsignedByte> byte2 = new ArrayList<>();
    private final List<UnsignedByte> byte3 = new ArrayList<>();
    private final List<UnsignedByte> byte4 = new ArrayList<>();
    private final List<UnsignedByte> byte5 = new ArrayList<>();
    private final List<BigInteger> counter = new ArrayList<>();
    private final List<BigInteger> inst = new ArrayList<>();
    private final List<Boolean> isData = new ArrayList<>();
    private final List<Boolean> known = new ArrayList<>();
    private final List<BigInteger> leftAlignedSuffixHigh = new ArrayList<>();
    private final List<BigInteger> leftAlignedSuffixLow = new ArrayList<>();
    private final List<BigInteger> low3 = new ArrayList<>();
    private final List<BigInteger> microShiftParameter = new ArrayList<>();
    private final List<Boolean> neg = new ArrayList<>();
    private final List<Boolean> oneLineInstruction = new ArrayList<>();
    private final List<BigInteger> ones = new ArrayList<>();
    private final List<BigInteger> resHi = new ArrayList<>();
    private final List<BigInteger> resLo = new ArrayList<>();
    private final List<BigInteger> rightAlignedPrefixHigh = new ArrayList<>();
    private final List<BigInteger> rightAlignedPrefixLow = new ArrayList<>();
    private final List<BigInteger> shb3Hi = new ArrayList<>();
    private final List<BigInteger> shb3Lo = new ArrayList<>();
    private final List<BigInteger> shb4Hi = new ArrayList<>();
    private final List<BigInteger> shb4Lo = new ArrayList<>();
    private final List<BigInteger> shb5Hi = new ArrayList<>();
    private final List<BigInteger> shb5Lo = new ArrayList<>();
    private final List<BigInteger> shb6Hi = new ArrayList<>();
    private final List<BigInteger> shb6Lo = new ArrayList<>();
    private final List<BigInteger> shb7Hi = new ArrayList<>();
    private final List<BigInteger> shb7Lo = new ArrayList<>();
    private final List<Boolean> shiftDirection = new ArrayList<>();
    private final List<BigInteger> shiftStamp = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder acc1(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(47);
      }

      acc1.add(b);

      return this;
    }

    TraceBuilder acc2(final BigInteger b) {
      if (filled.get(19)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(19);
      }

      acc2.add(b);

      return this;
    }

    TraceBuilder acc3(final BigInteger b) {
      if (filled.get(32)) {
        throw new IllegalStateException("ACC_3 already set");
      } else {
        filled.set(32);
      }

      acc3.add(b);

      return this;
    }

    TraceBuilder acc4(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("ACC_4 already set");
      } else {
        filled.set(26);
      }

      acc4.add(b);

      return this;
    }

    TraceBuilder acc5(final BigInteger b) {
      if (filled.get(41)) {
        throw new IllegalStateException("ACC_5 already set");
      } else {
        filled.set(41);
      }

      acc5.add(b);

      return this;
    }

    TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(11);
      }

      arg1Hi.add(b);

      return this;
    }

    TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(45);
      }

      arg1Lo.add(b);

      return this;
    }

    TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(25);
      }

      arg2Hi.add(b);

      return this;
    }

    TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(15);
      }

      arg2Lo.add(b);

      return this;
    }

    TraceBuilder bits(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("BITS already set");
      } else {
        filled.set(8);
      }

      bits.add(b);

      return this;
    }

    TraceBuilder bit1(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(16);
      }

      bit1.add(b);

      return this;
    }

    TraceBuilder bit2(final Boolean b) {
      if (filled.get(7)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(7);
      }

      bit2.add(b);

      return this;
    }

    TraceBuilder bit3(final Boolean b) {
      if (filled.get(28)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(28);
      }

      bit3.add(b);

      return this;
    }

    TraceBuilder bit4(final Boolean b) {
      if (filled.get(37)) {
        throw new IllegalStateException("BIT_4 already set");
      } else {
        filled.set(37);
      }

      bit4.add(b);

      return this;
    }

    TraceBuilder bitB3(final Boolean b) {
      if (filled.get(33)) {
        throw new IllegalStateException("BIT_B_3 already set");
      } else {
        filled.set(33);
      }

      bitB3.add(b);

      return this;
    }

    TraceBuilder bitB4(final Boolean b) {
      if (filled.get(4)) {
        throw new IllegalStateException("BIT_B_4 already set");
      } else {
        filled.set(4);
      }

      bitB4.add(b);

      return this;
    }

    TraceBuilder bitB5(final Boolean b) {
      if (filled.get(27)) {
        throw new IllegalStateException("BIT_B_5 already set");
      } else {
        filled.set(27);
      }

      bitB5.add(b);

      return this;
    }

    TraceBuilder bitB6(final Boolean b) {
      if (filled.get(46)) {
        throw new IllegalStateException("BIT_B_6 already set");
      } else {
        filled.set(46);
      }

      bitB6.add(b);

      return this;
    }

    TraceBuilder bitB7(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("BIT_B_7 already set");
      } else {
        filled.set(18);
      }

      bitB7.add(b);

      return this;
    }

    TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(12)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(12);
      }

      byte1.add(b);

      return this;
    }

    TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(49)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(49);
      }

      byte2.add(b);

      return this;
    }

    TraceBuilder byte3(final UnsignedByte b) {
      if (filled.get(48)) {
        throw new IllegalStateException("BYTE_3 already set");
      } else {
        filled.set(48);
      }

      byte3.add(b);

      return this;
    }

    TraceBuilder byte4(final UnsignedByte b) {
      if (filled.get(35)) {
        throw new IllegalStateException("BYTE_4 already set");
      } else {
        filled.set(35);
      }

      byte4.add(b);

      return this;
    }

    TraceBuilder byte5(final UnsignedByte b) {
      if (filled.get(5)) {
        throw new IllegalStateException("BYTE_5 already set");
      } else {
        filled.set(5);
      }

      byte5.add(b);

      return this;
    }

    TraceBuilder counter(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(44);
      }

      counter.add(b);

      return this;
    }

    TraceBuilder inst(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(6);
      }

      inst.add(b);

      return this;
    }

    TraceBuilder isData(final Boolean b) {
      if (filled.get(38)) {
        throw new IllegalStateException("IS_DATA already set");
      } else {
        filled.set(38);
      }

      isData.add(b);

      return this;
    }

    TraceBuilder known(final Boolean b) {
      if (filled.get(2)) {
        throw new IllegalStateException("KNOWN already set");
      } else {
        filled.set(2);
      }

      known.add(b);

      return this;
    }

    TraceBuilder leftAlignedSuffixHigh(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_HIGH already set");
      } else {
        filled.set(14);
      }

      leftAlignedSuffixHigh.add(b);

      return this;
    }

    TraceBuilder leftAlignedSuffixLow(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_LOW already set");
      } else {
        filled.set(29);
      }

      leftAlignedSuffixLow.add(b);

      return this;
    }

    TraceBuilder low3(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("LOW_3 already set");
      } else {
        filled.set(21);
      }

      low3.add(b);

      return this;
    }

    TraceBuilder microShiftParameter(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("MICRO_SHIFT_PARAMETER already set");
      } else {
        filled.set(23);
      }

      microShiftParameter.add(b);

      return this;
    }

    TraceBuilder neg(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("NEG already set");
      } else {
        filled.set(50);
      }

      neg.add(b);

      return this;
    }

    TraceBuilder ones(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("ONES already set");
      } else {
        filled.set(20);
      }

      ones.add(b);

      return this;
    }

    TraceBuilder oneLineInstruction(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION already set");
      } else {
        filled.set(36);
      }

      oneLineInstruction.add(b);

      return this;
    }

    TraceBuilder resHi(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(42);
      }

      resHi.add(b);

      return this;
    }

    TraceBuilder resLo(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(1);
      }

      resLo.add(b);

      return this;
    }

    TraceBuilder rightAlignedPrefixHigh(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_HIGH already set");
      } else {
        filled.set(40);
      }

      rightAlignedPrefixHigh.add(b);

      return this;
    }

    TraceBuilder rightAlignedPrefixLow(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_LOW already set");
      } else {
        filled.set(22);
      }

      rightAlignedPrefixLow.add(b);

      return this;
    }

    TraceBuilder shb3Hi(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("SHB_3_HI already set");
      } else {
        filled.set(17);
      }

      shb3Hi.add(b);

      return this;
    }

    TraceBuilder shb3Lo(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("SHB_3_LO already set");
      } else {
        filled.set(24);
      }

      shb3Lo.add(b);

      return this;
    }

    TraceBuilder shb4Hi(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("SHB_4_HI already set");
      } else {
        filled.set(10);
      }

      shb4Hi.add(b);

      return this;
    }

    TraceBuilder shb4Lo(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("SHB_4_LO already set");
      } else {
        filled.set(13);
      }

      shb4Lo.add(b);

      return this;
    }

    TraceBuilder shb5Hi(final BigInteger b) {
      if (filled.get(31)) {
        throw new IllegalStateException("SHB_5_HI already set");
      } else {
        filled.set(31);
      }

      shb5Hi.add(b);

      return this;
    }

    TraceBuilder shb5Lo(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("SHB_5_LO already set");
      } else {
        filled.set(30);
      }

      shb5Lo.add(b);

      return this;
    }

    TraceBuilder shb6Hi(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("SHB_6_HI already set");
      } else {
        filled.set(39);
      }

      shb6Hi.add(b);

      return this;
    }

    TraceBuilder shb6Lo(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("SHB_6_LO already set");
      } else {
        filled.set(43);
      }

      shb6Lo.add(b);

      return this;
    }

    TraceBuilder shb7Hi(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("SHB_7_HI already set");
      } else {
        filled.set(0);
      }

      shb7Hi.add(b);

      return this;
    }

    TraceBuilder shb7Lo(final BigInteger b) {
      if (filled.get(34)) {
        throw new IllegalStateException("SHB_7_LO already set");
      } else {
        filled.set(34);
      }

      shb7Lo.add(b);

      return this;
    }

    TraceBuilder shiftDirection(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("SHIFT_DIRECTION already set");
      } else {
        filled.set(9);
      }

      shiftDirection.add(b);

      return this;
    }

    TraceBuilder shiftStamp(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("SHIFT_STAMP already set");
      } else {
        filled.set(3);
      }

      shiftStamp.add(b);

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

    TraceBuilder setAcc5At(final BigInteger b, int i) {
      acc5.set(i, b);

      return this;
    }

    TraceBuilder setArg1HiAt(final BigInteger b, int i) {
      arg1Hi.set(i, b);

      return this;
    }

    TraceBuilder setArg1LoAt(final BigInteger b, int i) {
      arg1Lo.set(i, b);

      return this;
    }

    TraceBuilder setArg2HiAt(final BigInteger b, int i) {
      arg2Hi.set(i, b);

      return this;
    }

    TraceBuilder setArg2LoAt(final BigInteger b, int i) {
      arg2Lo.set(i, b);

      return this;
    }

    TraceBuilder setBitsAt(final Boolean b, int i) {
      bits.set(i, b);

      return this;
    }

    TraceBuilder setBit1At(final Boolean b, int i) {
      bit1.set(i, b);

      return this;
    }

    TraceBuilder setBit2At(final Boolean b, int i) {
      bit2.set(i, b);

      return this;
    }

    TraceBuilder setBit3At(final Boolean b, int i) {
      bit3.set(i, b);

      return this;
    }

    TraceBuilder setBit4At(final Boolean b, int i) {
      bit4.set(i, b);

      return this;
    }

    TraceBuilder setBitB3At(final Boolean b, int i) {
      bitB3.set(i, b);

      return this;
    }

    TraceBuilder setBitB4At(final Boolean b, int i) {
      bitB4.set(i, b);

      return this;
    }

    TraceBuilder setBitB5At(final Boolean b, int i) {
      bitB5.set(i, b);

      return this;
    }

    TraceBuilder setBitB6At(final Boolean b, int i) {
      bitB6.set(i, b);

      return this;
    }

    TraceBuilder setBitB7At(final Boolean b, int i) {
      bitB7.set(i, b);

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

    TraceBuilder setByte5At(final UnsignedByte b, int i) {
      byte5.set(i, b);

      return this;
    }

    TraceBuilder setCounterAt(final BigInteger b, int i) {
      counter.set(i, b);

      return this;
    }

    TraceBuilder setInstAt(final BigInteger b, int i) {
      inst.set(i, b);

      return this;
    }

    TraceBuilder setIsDataAt(final Boolean b, int i) {
      isData.set(i, b);

      return this;
    }

    TraceBuilder setKnownAt(final Boolean b, int i) {
      known.set(i, b);

      return this;
    }

    TraceBuilder setLeftAlignedSuffixHighAt(final BigInteger b, int i) {
      leftAlignedSuffixHigh.set(i, b);

      return this;
    }

    TraceBuilder setLeftAlignedSuffixLowAt(final BigInteger b, int i) {
      leftAlignedSuffixLow.set(i, b);

      return this;
    }

    TraceBuilder setLow3At(final BigInteger b, int i) {
      low3.set(i, b);

      return this;
    }

    TraceBuilder setMicroShiftParameterAt(final BigInteger b, int i) {
      microShiftParameter.set(i, b);

      return this;
    }

    TraceBuilder setNegAt(final Boolean b, int i) {
      neg.set(i, b);

      return this;
    }

    TraceBuilder setOnesAt(final BigInteger b, int i) {
      ones.set(i, b);

      return this;
    }

    TraceBuilder setOneLineInstructionAt(final Boolean b, int i) {
      oneLineInstruction.set(i, b);

      return this;
    }

    TraceBuilder setResHiAt(final BigInteger b, int i) {
      resHi.set(i, b);

      return this;
    }

    TraceBuilder setResLoAt(final BigInteger b, int i) {
      resLo.set(i, b);

      return this;
    }

    TraceBuilder setRightAlignedPrefixHighAt(final BigInteger b, int i) {
      rightAlignedPrefixHigh.set(i, b);

      return this;
    }

    TraceBuilder setRightAlignedPrefixLowAt(final BigInteger b, int i) {
      rightAlignedPrefixLow.set(i, b);

      return this;
    }

    TraceBuilder setShb3HiAt(final BigInteger b, int i) {
      shb3Hi.set(i, b);

      return this;
    }

    TraceBuilder setShb3LoAt(final BigInteger b, int i) {
      shb3Lo.set(i, b);

      return this;
    }

    TraceBuilder setShb4HiAt(final BigInteger b, int i) {
      shb4Hi.set(i, b);

      return this;
    }

    TraceBuilder setShb4LoAt(final BigInteger b, int i) {
      shb4Lo.set(i, b);

      return this;
    }

    TraceBuilder setShb5HiAt(final BigInteger b, int i) {
      shb5Hi.set(i, b);

      return this;
    }

    TraceBuilder setShb5LoAt(final BigInteger b, int i) {
      shb5Lo.set(i, b);

      return this;
    }

    TraceBuilder setShb6HiAt(final BigInteger b, int i) {
      shb6Hi.set(i, b);

      return this;
    }

    TraceBuilder setShb6LoAt(final BigInteger b, int i) {
      shb6Lo.set(i, b);

      return this;
    }

    TraceBuilder setShb7HiAt(final BigInteger b, int i) {
      shb7Hi.set(i, b);

      return this;
    }

    TraceBuilder setShb7LoAt(final BigInteger b, int i) {
      shb7Lo.set(i, b);

      return this;
    }

    TraceBuilder setShiftDirectionAt(final Boolean b, int i) {
      shiftDirection.set(i, b);

      return this;
    }

    TraceBuilder setShiftStampAt(final BigInteger b, int i) {
      shiftStamp.set(i, b);

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

    TraceBuilder setAcc5Relative(final BigInteger b, int i) {
      acc5.set(acc5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg1HiRelative(final BigInteger b, int i) {
      arg1Hi.set(arg1Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg1LoRelative(final BigInteger b, int i) {
      arg1Lo.set(arg1Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg2HiRelative(final BigInteger b, int i) {
      arg2Hi.set(arg2Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg2LoRelative(final BigInteger b, int i) {
      arg2Lo.set(arg2Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitsRelative(final Boolean b, int i) {
      bits.set(bits.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit1Relative(final Boolean b, int i) {
      bit1.set(bit1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit2Relative(final Boolean b, int i) {
      bit2.set(bit2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit3Relative(final Boolean b, int i) {
      bit3.set(bit3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit4Relative(final Boolean b, int i) {
      bit4.set(bit4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitB3Relative(final Boolean b, int i) {
      bitB3.set(bitB3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitB4Relative(final Boolean b, int i) {
      bitB4.set(bitB4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitB5Relative(final Boolean b, int i) {
      bitB5.set(bitB5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitB6Relative(final Boolean b, int i) {
      bitB6.set(bitB6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitB7Relative(final Boolean b, int i) {
      bitB7.set(bitB7.size() - 1 - i, b);

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

    TraceBuilder setByte5Relative(final UnsignedByte b, int i) {
      byte5.set(byte5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterRelative(final BigInteger b, int i) {
      counter.set(counter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInstRelative(final BigInteger b, int i) {
      inst.set(inst.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsDataRelative(final Boolean b, int i) {
      isData.set(isData.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setKnownRelative(final Boolean b, int i) {
      known.set(known.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLeftAlignedSuffixHighRelative(final BigInteger b, int i) {
      leftAlignedSuffixHigh.set(leftAlignedSuffixHigh.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLeftAlignedSuffixLowRelative(final BigInteger b, int i) {
      leftAlignedSuffixLow.set(leftAlignedSuffixLow.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLow3Relative(final BigInteger b, int i) {
      low3.set(low3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setMicroShiftParameterRelative(final BigInteger b, int i) {
      microShiftParameter.set(microShiftParameter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNegRelative(final Boolean b, int i) {
      neg.set(neg.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOnesRelative(final BigInteger b, int i) {
      ones.set(ones.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOneLineInstructionRelative(final Boolean b, int i) {
      oneLineInstruction.set(oneLineInstruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResHiRelative(final BigInteger b, int i) {
      resHi.set(resHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResLoRelative(final BigInteger b, int i) {
      resLo.set(resLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setRightAlignedPrefixHighRelative(final BigInteger b, int i) {
      rightAlignedPrefixHigh.set(rightAlignedPrefixHigh.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setRightAlignedPrefixLowRelative(final BigInteger b, int i) {
      rightAlignedPrefixLow.set(rightAlignedPrefixLow.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb3HiRelative(final BigInteger b, int i) {
      shb3Hi.set(shb3Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb3LoRelative(final BigInteger b, int i) {
      shb3Lo.set(shb3Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb4HiRelative(final BigInteger b, int i) {
      shb4Hi.set(shb4Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb4LoRelative(final BigInteger b, int i) {
      shb4Lo.set(shb4Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb5HiRelative(final BigInteger b, int i) {
      shb5Hi.set(shb5Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb5LoRelative(final BigInteger b, int i) {
      shb5Lo.set(shb5Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb6HiRelative(final BigInteger b, int i) {
      shb6Hi.set(shb6Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb6LoRelative(final BigInteger b, int i) {
      shb6Lo.set(shb6Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb7HiRelative(final BigInteger b, int i) {
      shb7Hi.set(shb7Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShb7LoRelative(final BigInteger b, int i) {
      shb7Lo.set(shb7Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShiftDirectionRelative(final Boolean b, int i) {
      shiftDirection.set(shiftDirection.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setShiftStampRelative(final BigInteger b, int i) {
      shiftStamp.set(shiftStamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(47)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("ACC_3 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("ACC_4 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("ACC_5 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("BIT_4 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("BIT_B_3 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("BIT_B_4 has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("BIT_B_5 has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("BIT_B_6 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("BIT_B_7 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("BITS has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("BYTE_3 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("BYTE_4 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("BYTE_5 has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("IS_DATA has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("KNOWN has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_HIGH has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_LOW has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("LOW_3 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("MICRO_SHIFT_PARAMETER has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("NEG has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("ONES has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_HIGH has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_LOW has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("SHB_3_HI has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("SHB_3_LO has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("SHB_4_HI has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("SHB_4_LO has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("SHB_5_HI has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("SHB_5_LO has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("SHB_6_HI has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("SHB_6_LO has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("SHB_7_HI has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("SHB_7_LO has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("SHIFT_DIRECTION has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("SHIFT_STAMP has not been filled");
      }

      filled.clear();

      return this;
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          acc1,
          acc2,
          acc3,
          acc4,
          acc5,
          arg1Hi,
          arg1Lo,
          arg2Hi,
          arg2Lo,
          bit1,
          bit2,
          bit3,
          bit4,
          bitB3,
          bitB4,
          bitB5,
          bitB6,
          bitB7,
          bits,
          byte1,
          byte2,
          byte3,
          byte4,
          byte5,
          counter,
          inst,
          isData,
          known,
          leftAlignedSuffixHigh,
          leftAlignedSuffixLow,
          low3,
          microShiftParameter,
          neg,
          oneLineInstruction,
          ones,
          resHi,
          resLo,
          rightAlignedPrefixHigh,
          rightAlignedPrefixLow,
          shb3Hi,
          shb3Lo,
          shb4Hi,
          shb4Lo,
          shb5Hi,
          shb5Lo,
          shb6Hi,
          shb6Lo,
          shb7Hi,
          shb7Lo,
          shiftDirection,
          shiftStamp);
    }
  }
}
