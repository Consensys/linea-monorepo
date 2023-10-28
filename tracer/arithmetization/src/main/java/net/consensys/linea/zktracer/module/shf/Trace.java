/*
 * Copyright Consensys Software Inc.
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
public record Trace(
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
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.acc1.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ACC_1")
    private final List<BigInteger> acc1;

    @JsonProperty("ACC_2")
    private final List<BigInteger> acc2;

    @JsonProperty("ACC_3")
    private final List<BigInteger> acc3;

    @JsonProperty("ACC_4")
    private final List<BigInteger> acc4;

    @JsonProperty("ACC_5")
    private final List<BigInteger> acc5;

    @JsonProperty("ARG_1_HI")
    private final List<BigInteger> arg1Hi;

    @JsonProperty("ARG_1_LO")
    private final List<BigInteger> arg1Lo;

    @JsonProperty("ARG_2_HI")
    private final List<BigInteger> arg2Hi;

    @JsonProperty("ARG_2_LO")
    private final List<BigInteger> arg2Lo;

    @JsonProperty("BIT_1")
    private final List<Boolean> bit1;

    @JsonProperty("BIT_2")
    private final List<Boolean> bit2;

    @JsonProperty("BIT_3")
    private final List<Boolean> bit3;

    @JsonProperty("BIT_4")
    private final List<Boolean> bit4;

    @JsonProperty("BIT_B_3")
    private final List<Boolean> bitB3;

    @JsonProperty("BIT_B_4")
    private final List<Boolean> bitB4;

    @JsonProperty("BIT_B_5")
    private final List<Boolean> bitB5;

    @JsonProperty("BIT_B_6")
    private final List<Boolean> bitB6;

    @JsonProperty("BIT_B_7")
    private final List<Boolean> bitB7;

    @JsonProperty("BITS")
    private final List<Boolean> bits;

    @JsonProperty("BYTE_1")
    private final List<UnsignedByte> byte1;

    @JsonProperty("BYTE_2")
    private final List<UnsignedByte> byte2;

    @JsonProperty("BYTE_3")
    private final List<UnsignedByte> byte3;

    @JsonProperty("BYTE_4")
    private final List<UnsignedByte> byte4;

    @JsonProperty("BYTE_5")
    private final List<UnsignedByte> byte5;

    @JsonProperty("COUNTER")
    private final List<BigInteger> counter;

    @JsonProperty("INST")
    private final List<BigInteger> inst;

    @JsonProperty("IS_DATA")
    private final List<Boolean> isData;

    @JsonProperty("KNOWN")
    private final List<Boolean> known;

    @JsonProperty("LEFT_ALIGNED_SUFFIX_HIGH")
    private final List<BigInteger> leftAlignedSuffixHigh;

    @JsonProperty("LEFT_ALIGNED_SUFFIX_LOW")
    private final List<BigInteger> leftAlignedSuffixLow;

    @JsonProperty("LOW_3")
    private final List<BigInteger> low3;

    @JsonProperty("MICRO_SHIFT_PARAMETER")
    private final List<BigInteger> microShiftParameter;

    @JsonProperty("NEG")
    private final List<Boolean> neg;

    @JsonProperty("ONE_LINE_INSTRUCTION")
    private final List<Boolean> oneLineInstruction;

    @JsonProperty("ONES")
    private final List<BigInteger> ones;

    @JsonProperty("RES_HI")
    private final List<BigInteger> resHi;

    @JsonProperty("RES_LO")
    private final List<BigInteger> resLo;

    @JsonProperty("RIGHT_ALIGNED_PREFIX_HIGH")
    private final List<BigInteger> rightAlignedPrefixHigh;

    @JsonProperty("RIGHT_ALIGNED_PREFIX_LOW")
    private final List<BigInteger> rightAlignedPrefixLow;

    @JsonProperty("SHB_3_HI")
    private final List<BigInteger> shb3Hi;

    @JsonProperty("SHB_3_LO")
    private final List<BigInteger> shb3Lo;

    @JsonProperty("SHB_4_HI")
    private final List<BigInteger> shb4Hi;

    @JsonProperty("SHB_4_LO")
    private final List<BigInteger> shb4Lo;

    @JsonProperty("SHB_5_HI")
    private final List<BigInteger> shb5Hi;

    @JsonProperty("SHB_5_LO")
    private final List<BigInteger> shb5Lo;

    @JsonProperty("SHB_6_HI")
    private final List<BigInteger> shb6Hi;

    @JsonProperty("SHB_6_LO")
    private final List<BigInteger> shb6Lo;

    @JsonProperty("SHB_7_HI")
    private final List<BigInteger> shb7Hi;

    @JsonProperty("SHB_7_LO")
    private final List<BigInteger> shb7Lo;

    @JsonProperty("SHIFT_DIRECTION")
    private final List<Boolean> shiftDirection;

    @JsonProperty("SHIFT_STAMP")
    private final List<BigInteger> shiftStamp;

    private TraceBuilder(int length) {
      this.acc1 = new ArrayList<>(length);
      this.acc2 = new ArrayList<>(length);
      this.acc3 = new ArrayList<>(length);
      this.acc4 = new ArrayList<>(length);
      this.acc5 = new ArrayList<>(length);
      this.arg1Hi = new ArrayList<>(length);
      this.arg1Lo = new ArrayList<>(length);
      this.arg2Hi = new ArrayList<>(length);
      this.arg2Lo = new ArrayList<>(length);
      this.bit1 = new ArrayList<>(length);
      this.bit2 = new ArrayList<>(length);
      this.bit3 = new ArrayList<>(length);
      this.bit4 = new ArrayList<>(length);
      this.bitB3 = new ArrayList<>(length);
      this.bitB4 = new ArrayList<>(length);
      this.bitB5 = new ArrayList<>(length);
      this.bitB6 = new ArrayList<>(length);
      this.bitB7 = new ArrayList<>(length);
      this.bits = new ArrayList<>(length);
      this.byte1 = new ArrayList<>(length);
      this.byte2 = new ArrayList<>(length);
      this.byte3 = new ArrayList<>(length);
      this.byte4 = new ArrayList<>(length);
      this.byte5 = new ArrayList<>(length);
      this.counter = new ArrayList<>(length);
      this.inst = new ArrayList<>(length);
      this.isData = new ArrayList<>(length);
      this.known = new ArrayList<>(length);
      this.leftAlignedSuffixHigh = new ArrayList<>(length);
      this.leftAlignedSuffixLow = new ArrayList<>(length);
      this.low3 = new ArrayList<>(length);
      this.microShiftParameter = new ArrayList<>(length);
      this.neg = new ArrayList<>(length);
      this.oneLineInstruction = new ArrayList<>(length);
      this.ones = new ArrayList<>(length);
      this.resHi = new ArrayList<>(length);
      this.resLo = new ArrayList<>(length);
      this.rightAlignedPrefixHigh = new ArrayList<>(length);
      this.rightAlignedPrefixLow = new ArrayList<>(length);
      this.shb3Hi = new ArrayList<>(length);
      this.shb3Lo = new ArrayList<>(length);
      this.shb4Hi = new ArrayList<>(length);
      this.shb4Lo = new ArrayList<>(length);
      this.shb5Hi = new ArrayList<>(length);
      this.shb5Lo = new ArrayList<>(length);
      this.shb6Hi = new ArrayList<>(length);
      this.shb6Lo = new ArrayList<>(length);
      this.shb7Hi = new ArrayList<>(length);
      this.shb7Lo = new ArrayList<>(length);
      this.shiftDirection = new ArrayList<>(length);
      this.shiftStamp = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.acc1.size();
    }

    public TraceBuilder acc1(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(0);
      }

      acc1.add(b);

      return this;
    }

    public TraceBuilder acc2(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(1);
      }

      acc2.add(b);

      return this;
    }

    public TraceBuilder acc3(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ACC_3 already set");
      } else {
        filled.set(2);
      }

      acc3.add(b);

      return this;
    }

    public TraceBuilder acc4(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_4 already set");
      } else {
        filled.set(3);
      }

      acc4.add(b);

      return this;
    }

    public TraceBuilder acc5(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_5 already set");
      } else {
        filled.set(4);
      }

      acc5.add(b);

      return this;
    }

    public TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(5);
      }

      arg1Hi.add(b);

      return this;
    }

    public TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(6);
      }

      arg1Lo.add(b);

      return this;
    }

    public TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(7);
      }

      arg2Hi.add(b);

      return this;
    }

    public TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(8);
      }

      arg2Lo.add(b);

      return this;
    }

    public TraceBuilder bit1(final Boolean b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(10);
      }

      bit1.add(b);

      return this;
    }

    public TraceBuilder bit2(final Boolean b) {
      if (filled.get(11)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(11);
      }

      bit2.add(b);

      return this;
    }

    public TraceBuilder bit3(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(12);
      }

      bit3.add(b);

      return this;
    }

    public TraceBuilder bit4(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("BIT_4 already set");
      } else {
        filled.set(13);
      }

      bit4.add(b);

      return this;
    }

    public TraceBuilder bitB3(final Boolean b) {
      if (filled.get(14)) {
        throw new IllegalStateException("BIT_B_3 already set");
      } else {
        filled.set(14);
      }

      bitB3.add(b);

      return this;
    }

    public TraceBuilder bitB4(final Boolean b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BIT_B_4 already set");
      } else {
        filled.set(15);
      }

      bitB4.add(b);

      return this;
    }

    public TraceBuilder bitB5(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("BIT_B_5 already set");
      } else {
        filled.set(16);
      }

      bitB5.add(b);

      return this;
    }

    public TraceBuilder bitB6(final Boolean b) {
      if (filled.get(17)) {
        throw new IllegalStateException("BIT_B_6 already set");
      } else {
        filled.set(17);
      }

      bitB6.add(b);

      return this;
    }

    public TraceBuilder bitB7(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("BIT_B_7 already set");
      } else {
        filled.set(18);
      }

      bitB7.add(b);

      return this;
    }

    public TraceBuilder bits(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("BITS already set");
      } else {
        filled.set(9);
      }

      bits.add(b);

      return this;
    }

    public TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(19)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(19);
      }

      byte1.add(b);

      return this;
    }

    public TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(20);
      }

      byte2.add(b);

      return this;
    }

    public TraceBuilder byte3(final UnsignedByte b) {
      if (filled.get(21)) {
        throw new IllegalStateException("BYTE_3 already set");
      } else {
        filled.set(21);
      }

      byte3.add(b);

      return this;
    }

    public TraceBuilder byte4(final UnsignedByte b) {
      if (filled.get(22)) {
        throw new IllegalStateException("BYTE_4 already set");
      } else {
        filled.set(22);
      }

      byte4.add(b);

      return this;
    }

    public TraceBuilder byte5(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BYTE_5 already set");
      } else {
        filled.set(23);
      }

      byte5.add(b);

      return this;
    }

    public TraceBuilder counter(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(24);
      }

      counter.add(b);

      return this;
    }

    public TraceBuilder inst(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(25);
      }

      inst.add(b);

      return this;
    }

    public TraceBuilder isData(final Boolean b) {
      if (filled.get(26)) {
        throw new IllegalStateException("IS_DATA already set");
      } else {
        filled.set(26);
      }

      isData.add(b);

      return this;
    }

    public TraceBuilder known(final Boolean b) {
      if (filled.get(27)) {
        throw new IllegalStateException("KNOWN already set");
      } else {
        filled.set(27);
      }

      known.add(b);

      return this;
    }

    public TraceBuilder leftAlignedSuffixHigh(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_HIGH already set");
      } else {
        filled.set(28);
      }

      leftAlignedSuffixHigh.add(b);

      return this;
    }

    public TraceBuilder leftAlignedSuffixLow(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_LOW already set");
      } else {
        filled.set(29);
      }

      leftAlignedSuffixLow.add(b);

      return this;
    }

    public TraceBuilder low3(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("LOW_3 already set");
      } else {
        filled.set(30);
      }

      low3.add(b);

      return this;
    }

    public TraceBuilder microShiftParameter(final BigInteger b) {
      if (filled.get(31)) {
        throw new IllegalStateException("MICRO_SHIFT_PARAMETER already set");
      } else {
        filled.set(31);
      }

      microShiftParameter.add(b);

      return this;
    }

    public TraceBuilder neg(final Boolean b) {
      if (filled.get(32)) {
        throw new IllegalStateException("NEG already set");
      } else {
        filled.set(32);
      }

      neg.add(b);

      return this;
    }

    public TraceBuilder oneLineInstruction(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION already set");
      } else {
        filled.set(34);
      }

      oneLineInstruction.add(b);

      return this;
    }

    public TraceBuilder ones(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("ONES already set");
      } else {
        filled.set(33);
      }

      ones.add(b);

      return this;
    }

    public TraceBuilder resHi(final BigInteger b) {
      if (filled.get(35)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(35);
      }

      resHi.add(b);

      return this;
    }

    public TraceBuilder resLo(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(36);
      }

      resLo.add(b);

      return this;
    }

    public TraceBuilder rightAlignedPrefixHigh(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_HIGH already set");
      } else {
        filled.set(37);
      }

      rightAlignedPrefixHigh.add(b);

      return this;
    }

    public TraceBuilder rightAlignedPrefixLow(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_LOW already set");
      } else {
        filled.set(38);
      }

      rightAlignedPrefixLow.add(b);

      return this;
    }

    public TraceBuilder shb3Hi(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("SHB_3_HI already set");
      } else {
        filled.set(39);
      }

      shb3Hi.add(b);

      return this;
    }

    public TraceBuilder shb3Lo(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("SHB_3_LO already set");
      } else {
        filled.set(40);
      }

      shb3Lo.add(b);

      return this;
    }

    public TraceBuilder shb4Hi(final BigInteger b) {
      if (filled.get(41)) {
        throw new IllegalStateException("SHB_4_HI already set");
      } else {
        filled.set(41);
      }

      shb4Hi.add(b);

      return this;
    }

    public TraceBuilder shb4Lo(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("SHB_4_LO already set");
      } else {
        filled.set(42);
      }

      shb4Lo.add(b);

      return this;
    }

    public TraceBuilder shb5Hi(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("SHB_5_HI already set");
      } else {
        filled.set(43);
      }

      shb5Hi.add(b);

      return this;
    }

    public TraceBuilder shb5Lo(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("SHB_5_LO already set");
      } else {
        filled.set(44);
      }

      shb5Lo.add(b);

      return this;
    }

    public TraceBuilder shb6Hi(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("SHB_6_HI already set");
      } else {
        filled.set(45);
      }

      shb6Hi.add(b);

      return this;
    }

    public TraceBuilder shb6Lo(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("SHB_6_LO already set");
      } else {
        filled.set(46);
      }

      shb6Lo.add(b);

      return this;
    }

    public TraceBuilder shb7Hi(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("SHB_7_HI already set");
      } else {
        filled.set(47);
      }

      shb7Hi.add(b);

      return this;
    }

    public TraceBuilder shb7Lo(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("SHB_7_LO already set");
      } else {
        filled.set(48);
      }

      shb7Lo.add(b);

      return this;
    }

    public TraceBuilder shiftDirection(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("SHIFT_DIRECTION already set");
      } else {
        filled.set(49);
      }

      shiftDirection.add(b);

      return this;
    }

    public TraceBuilder shiftStamp(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("SHIFT_STAMP already set");
      } else {
        filled.set(50);
      }

      shiftStamp.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ACC_3 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_4 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_5 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("BIT_4 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("BIT_B_3 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BIT_B_4 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("BIT_B_5 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("BIT_B_6 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("BIT_B_7 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("BITS has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("BYTE_3 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("BYTE_4 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BYTE_5 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("IS_DATA has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("KNOWN has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_HIGH has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("LEFT_ALIGNED_SUFFIX_LOW has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("LOW_3 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("MICRO_SHIFT_PARAMETER has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("NEG has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("ONES has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_HIGH has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("RIGHT_ALIGNED_PREFIX_LOW has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("SHB_3_HI has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("SHB_3_LO has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("SHB_4_HI has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("SHB_4_LO has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("SHB_5_HI has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("SHB_5_LO has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("SHB_6_HI has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("SHB_6_LO has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("SHB_7_HI has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("SHB_7_LO has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("SHIFT_DIRECTION has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("SHIFT_STAMP has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        acc1.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        acc2.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        acc3.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        acc4.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        acc5.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        arg1Hi.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        arg1Lo.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        arg2Hi.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        arg2Lo.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(10)) {
        bit1.add(false);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        bit2.add(false);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        bit3.add(false);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        bit4.add(false);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        bitB3.add(false);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        bitB4.add(false);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        bitB5.add(false);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        bitB6.add(false);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        bitB7.add(false);
        this.filled.set(18);
      }
      if (!filled.get(9)) {
        bits.add(false);
        this.filled.set(9);
      }
      if (!filled.get(19)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        byte2.add(UnsignedByte.of(0));
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        byte3.add(UnsignedByte.of(0));
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        byte4.add(UnsignedByte.of(0));
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        byte5.add(UnsignedByte.of(0));
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        counter.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        inst.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        isData.add(false);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        known.add(false);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        leftAlignedSuffixHigh.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        leftAlignedSuffixLow.add(BigInteger.ZERO);
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        low3.add(BigInteger.ZERO);
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        microShiftParameter.add(BigInteger.ZERO);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        neg.add(false);
        this.filled.set(32);
      }
      if (!filled.get(34)) {
        oneLineInstruction.add(false);
        this.filled.set(34);
      }
      if (!filled.get(33)) {
        ones.add(BigInteger.ZERO);
        this.filled.set(33);
      }
      if (!filled.get(35)) {
        resHi.add(BigInteger.ZERO);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        resLo.add(BigInteger.ZERO);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        rightAlignedPrefixHigh.add(BigInteger.ZERO);
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        rightAlignedPrefixLow.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        shb3Hi.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        shb3Lo.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        shb4Hi.add(BigInteger.ZERO);
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        shb4Lo.add(BigInteger.ZERO);
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        shb5Hi.add(BigInteger.ZERO);
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        shb5Lo.add(BigInteger.ZERO);
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        shb6Hi.add(BigInteger.ZERO);
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        shb6Lo.add(BigInteger.ZERO);
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        shb7Hi.add(BigInteger.ZERO);
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        shb7Lo.add(BigInteger.ZERO);
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        shiftDirection.add(false);
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        shiftStamp.add(BigInteger.ZERO);
        this.filled.set(50);
      }

      return this.validateRow();
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
