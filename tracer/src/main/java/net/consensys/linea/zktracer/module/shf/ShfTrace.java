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

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({"Trace", "Stamp"})
@SuppressWarnings("unused")
public record ShfTrace(@JsonProperty("Trace") Trace trace, @JsonProperty("Stamp") int stamp) {
  @JsonPropertyOrder({
    "ACC_1",
    "ACC_2",
    "ACC_3",
    "ACC_4",
    "ACC_5",
    "ARG_1_HI",
    "ARG_1_LO",
    "ARG_2_HI",
    "ARG_2_LO",
    "BITS",
    "BIT_1",
    "BIT_2",
    "BIT_3",
    "BIT_4",
    "BIT_B_3",
    "BIT_B_4",
    "BIT_B_5",
    "BIT_B_6",
    "BIT_B_7",
    "BYTE_1",
    "BYTE_2",
    "BYTE_3",
    "BYTE_4",
    "BYTE_5",
    "COUNTER",
    "INST",
    "IS_DATA",
    "KNOWN",
    "LEFT_ALIGNED_SUFFIX_HIGH",
    "LEFT_ALIGNED_SUFFIX_LOW",
    "LOW_3",
    "MICRO_SHIFT_PARAMETER",
    "NEG",
    "ONES",
    "ONE_LINE_INSTRUCTION",
    "RES_HI",
    "RES_LO",
    "RIGHT_ALIGNED_PREFIX_HIGH",
    "RIGHT_ALIGNED_PREFIX_LOW",
    "SHB_3_HI",
    "SHB_3_LO",
    "SHB_4_HI",
    "SHB_4_LO",
    "SHB_5_HI",
    "SHB_5_LO",
    "SHB_6_HI",
    "SHB_6_LO",
    "SHB_7_HI",
    "SHB_7_LO",
    "SHIFT_DIRECTION",
    "SHIFT_STAMP",
  })
  @SuppressWarnings("unused")
  public record Trace(
      @JsonProperty("ACC_1") List<BigInteger> ACC_1,
      @JsonProperty("ACC_2") List<BigInteger> ACC_2,
      @JsonProperty("ACC_3") List<BigInteger> ACC_3,
      @JsonProperty("ACC_4") List<BigInteger> ACC_4,
      @JsonProperty("ACC_5") List<BigInteger> ACC_5,
      @JsonProperty("ARG_1_HI") List<BigInteger> ARG_1_HI,
      @JsonProperty("ARG_1_LO") List<BigInteger> ARG_1_LO,
      @JsonProperty("ARG_2_HI") List<BigInteger> ARG_2_HI,
      @JsonProperty("ARG_2_LO") List<BigInteger> ARG_2_LO,
      @JsonProperty("BITS") List<Boolean> BITS,
      @JsonProperty("BIT_1") List<Boolean> BIT_1,
      @JsonProperty("BIT_2") List<Boolean> BIT_2,
      @JsonProperty("BIT_3") List<Boolean> BIT_3,
      @JsonProperty("BIT_4") List<Boolean> BIT_4,
      @JsonProperty("BIT_B_3") List<Boolean> BIT_B_3,
      @JsonProperty("BIT_B_4") List<Boolean> BIT_B_4,
      @JsonProperty("BIT_B_5") List<Boolean> BIT_B_5,
      @JsonProperty("BIT_B_6") List<Boolean> BIT_B_6,
      @JsonProperty("BIT_B_7") List<Boolean> BIT_B_7,
      @JsonProperty("BYTE_1") List<UnsignedByte> BYTE_1,
      @JsonProperty("BYTE_2") List<UnsignedByte> BYTE_2,
      @JsonProperty("BYTE_3") List<UnsignedByte> BYTE_3,
      @JsonProperty("BYTE_4") List<UnsignedByte> BYTE_4,
      @JsonProperty("BYTE_5") List<UnsignedByte> BYTE_5,
      @JsonProperty("COUNTER") List<Integer> COUNTER,
      @JsonProperty("INST") List<UnsignedByte> INST,
      @JsonProperty("IS_DATA") List<Boolean> IS_DATA,
      @JsonProperty("KNOWN") List<Boolean> KNOWN,
      @JsonProperty("LEFT_ALIGNED_SUFFIX_HIGH") List<UnsignedByte> LEFT_ALIGNED_SUFFIX_HIGH,
      @JsonProperty("LEFT_ALIGNED_SUFFIX_LOW") List<UnsignedByte> LEFT_ALIGNED_SUFFIX_LOW,
      @JsonProperty("LOW_3") List<UnsignedByte> LOW_3,
      @JsonProperty("MICRO_SHIFT_PARAMETER") List<UnsignedByte> MICRO_SHIFT_PARAMETER,
      @JsonProperty("NEG") List<Boolean> NEG,
      @JsonProperty("ONES") List<UnsignedByte> ONES,
      @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> ONE_LINE_INSTRUCTION,
      @JsonProperty("RES_HI") List<BigInteger> RES_HI,
      @JsonProperty("RES_LO") List<BigInteger> RES_LO,
      @JsonProperty("RIGHT_ALIGNED_PREFIX_HIGH") List<UnsignedByte> RIGHT_ALIGNED_PREFIX_HIGH,
      @JsonProperty("RIGHT_ALIGNED_PREFIX_LOW") List<UnsignedByte> RIGHT_ALIGNED_PREFIX_LOW,
      @JsonProperty("SHB_3_HI") List<UnsignedByte> SHB_3_HI,
      @JsonProperty("SHB_3_LO") List<UnsignedByte> SHB_3_LO,
      @JsonProperty("SHB_4_HI") List<UnsignedByte> SHB_4_HI,
      @JsonProperty("SHB_4_LO") List<UnsignedByte> SHB_4_LO,
      @JsonProperty("SHB_5_HI") List<UnsignedByte> SHB_5_HI,
      @JsonProperty("SHB_5_LO") List<UnsignedByte> SHB_5_LO,
      @JsonProperty("SHB_6_HI") List<UnsignedByte> SHB_6_HI,
      @JsonProperty("SHB_6_LO") List<UnsignedByte> SHB_6_LO,
      @JsonProperty("SHB_7_HI") List<UnsignedByte> SHB_7_HI,
      @JsonProperty("SHB_7_LO") List<UnsignedByte> SHB_7_LO,
      @JsonProperty("SHIFT_DIRECTION") List<Boolean> SHIFT_DIRECTION,
      @JsonProperty("SHIFT_STAMP") List<Integer> SHIFT_STAMP) {

    public static class Builder {
      private final List<BigInteger> acc1 = new ArrayList<>();
      private final List<BigInteger> acc2 = new ArrayList<>();
      private final List<BigInteger> acc3 = new ArrayList<>();
      private final List<BigInteger> acc4 = new ArrayList<>();
      private final List<BigInteger> acc5 = new ArrayList<>();
      private final List<BigInteger> arg1Hi = new ArrayList<>();
      private final List<BigInteger> arg1Lo = new ArrayList<>();
      private final List<BigInteger> arg2Hi = new ArrayList<>();
      private final List<BigInteger> arg2Lo = new ArrayList<>();
      private final List<Boolean> bits = new ArrayList<>();
      private final List<Boolean> bit1 = new ArrayList<>();
      private final List<Boolean> bit2 = new ArrayList<>();
      private final List<Boolean> bit3 = new ArrayList<>();
      private final List<Boolean> bit4 = new ArrayList<>();
      private final List<Boolean> bitB3 = new ArrayList<>();
      private final List<Boolean> bitB4 = new ArrayList<>();
      private final List<Boolean> bitB5 = new ArrayList<>();
      private final List<Boolean> bitB6 = new ArrayList<>();
      private final List<Boolean> bitB7 = new ArrayList<>();
      private final List<UnsignedByte> byte1 = new ArrayList<>();
      private final List<UnsignedByte> byte2 = new ArrayList<>();
      private final List<UnsignedByte> byte3 = new ArrayList<>();
      private final List<UnsignedByte> byte4 = new ArrayList<>();
      private final List<UnsignedByte> byte5 = new ArrayList<>();
      private final List<Integer> counter = new ArrayList<>();
      private final List<UnsignedByte> inst = new ArrayList<>();
      private final List<Boolean> isData = new ArrayList<>();
      private final List<Boolean> known = new ArrayList<>();
      private final List<UnsignedByte> leftAlignedSuffixHigh = new ArrayList<>();
      private final List<UnsignedByte> leftAlignedSuffixLow = new ArrayList<>();
      private final List<UnsignedByte> low3 = new ArrayList<>();
      private final List<UnsignedByte> microShiftParameter = new ArrayList<>();
      private final List<Boolean> neg = new ArrayList<>();
      private final List<UnsignedByte> ones = new ArrayList<>();
      private final List<Boolean> oneLineInstruction = new ArrayList<>();
      private final List<BigInteger> resHi = new ArrayList<>();
      private final List<BigInteger> resLo = new ArrayList<>();
      private final List<UnsignedByte> rightAlignedPrefixHigh = new ArrayList<>();
      private final List<UnsignedByte> rightAlignedPrefixLow = new ArrayList<>();
      private final List<UnsignedByte> shb3Hi = new ArrayList<>();
      private final List<UnsignedByte> shb3Lo = new ArrayList<>();
      private final List<UnsignedByte> shb4Hi = new ArrayList<>();
      private final List<UnsignedByte> shb4Lo = new ArrayList<>();
      private final List<UnsignedByte> shb5Hi = new ArrayList<>();
      private final List<UnsignedByte> shb5Lo = new ArrayList<>();
      private final List<UnsignedByte> shb6Hi = new ArrayList<>();
      private final List<UnsignedByte> shb6Lo = new ArrayList<>();
      private final List<UnsignedByte> shb7Hi = new ArrayList<>();
      private final List<UnsignedByte> shb7Lo = new ArrayList<>();
      private final List<Boolean> shiftDirection = new ArrayList<>();
      private final List<Integer> shiftStamp = new ArrayList<>();
      private int stamp = 0;

      private Builder() {}

      public static Builder newInstance() {
        return new Builder();
      }

      public Builder appendAcc1(final BigInteger b) {
        acc1.add(b);
        return this;
      }

      public Builder appendAcc2(final BigInteger b) {
        acc2.add(b);
        return this;
      }

      public Builder appendAcc3(final BigInteger b) {
        acc3.add(b);
        return this;
      }

      public Builder appendAcc4(final BigInteger b) {
        acc4.add(b);
        return this;
      }

      public Builder appendAcc5(final BigInteger b) {
        acc5.add(b);
        return this;
      }

      public Builder appendArg1Hi(final BigInteger b) {
        arg1Hi.add(b);
        return this;
      }

      public Builder appendArg1Lo(final BigInteger b) {
        arg1Lo.add(b);
        return this;
      }

      public Builder appendArg2Hi(final BigInteger b) {
        arg2Hi.add(b);
        return this;
      }

      public Builder appendArg2Lo(final BigInteger b) {
        arg2Lo.add(b);
        return this;
      }

      public Builder appendBits(final Boolean b) {
        bits.add(b);
        return this;
      }

      public Builder appendBit1(final Boolean b) {
        bit1.add(b);
        return this;
      }

      public Builder appendBit2(final Boolean b) {
        bit2.add(b);
        return this;
      }

      public Builder appendBit3(final Boolean b) {
        bit3.add(b);
        return this;
      }

      public Builder appendBit4(final Boolean b) {
        bit4.add(b);
        return this;
      }

      public Builder appendBitB3(final Boolean b) {
        bitB3.add(b);
        return this;
      }

      public Builder appendBitB4(final Boolean b) {
        bitB4.add(b);
        return this;
      }

      public Builder appendBitB5(final Boolean b) {
        bitB5.add(b);
        return this;
      }

      public Builder appendBitB6(final Boolean b) {
        bitB6.add(b);
        return this;
      }

      public Builder appendBitB7(final Boolean b) {
        bitB7.add(b);
        return this;
      }

      public Builder appendByte1(final UnsignedByte b) {
        byte1.add(b);
        return this;
      }

      public Builder appendByte2(final UnsignedByte b) {
        byte2.add(b);
        return this;
      }

      public Builder appendByte3(final UnsignedByte b) {
        byte3.add(b);
        return this;
      }

      public Builder appendByte4(final UnsignedByte b) {
        byte4.add(b);
        return this;
      }

      public Builder appendByte5(final UnsignedByte b) {
        byte5.add(b);
        return this;
      }

      public Builder appendCounter(final Integer b) {
        counter.add(b);
        return this;
      }

      public Builder appendInst(final UnsignedByte b) {
        inst.add(b);
        return this;
      }

      public Builder appendIsData(final Boolean b) {
        isData.add(b);
        return this;
      }

      public Builder appendKnown(final Boolean b) {
        known.add(b);
        return this;
      }

      public Builder appendLeftAlignedSuffixHigh(final UnsignedByte b) {
        leftAlignedSuffixHigh.add(b);
        return this;
      }

      public Builder appendLeftAlignedSuffixLow(final UnsignedByte b) {
        leftAlignedSuffixLow.add(b);
        return this;
      }

      public Builder appendLow3(final UnsignedByte b) {
        low3.add(b);
        return this;
      }

      public Builder appendMicroShiftParameter(final UnsignedByte b) {
        microShiftParameter.add(b);
        return this;
      }

      public Builder appendNeg(final Boolean b) {
        neg.add(b);
        return this;
      }

      public Builder appendOnes(final UnsignedByte b) {
        ones.add(b);
        return this;
      }

      public Builder appendOneLineInstruction(final Boolean b) {
        oneLineInstruction.add(b);
        return this;
      }

      public Builder appendResHi(final BigInteger b) {
        resHi.add(b);
        return this;
      }

      public Builder appendResLo(final BigInteger b) {
        resLo.add(b);
        return this;
      }

      public Builder appendRightAlignedPrefixHigh(final UnsignedByte b) {
        rightAlignedPrefixHigh.add(b);
        return this;
      }

      public Builder appendRightAlignedPrefixLow(final UnsignedByte b) {
        rightAlignedPrefixLow.add(b);
        return this;
      }

      public Builder appendShb3Hi(final UnsignedByte b) {
        shb3Hi.add(b);
        return this;
      }

      public Builder appendShb3Lo(final UnsignedByte b) {
        shb3Lo.add(b);
        return this;
      }

      public Builder appendShb4Hi(final UnsignedByte b) {
        shb4Hi.add(b);
        return this;
      }

      public Builder appendShb4Lo(final UnsignedByte b) {
        shb4Lo.add(b);
        return this;
      }

      public Builder appendShb5Hi(final UnsignedByte b) {
        shb5Hi.add(b);
        return this;
      }

      public Builder appendShb5Lo(final UnsignedByte b) {
        shb5Lo.add(b);
        return this;
      }

      public Builder appendShb6Hi(final UnsignedByte b) {
        shb6Hi.add(b);
        return this;
      }

      public Builder appendShb6Lo(final UnsignedByte b) {
        shb6Lo.add(b);
        return this;
      }

      public Builder appendShb7Hi(final UnsignedByte b) {
        shb7Hi.add(b);
        return this;
      }

      public Builder appendShb7Lo(final UnsignedByte b) {
        shb7Lo.add(b);
        return this;
      }

      public Builder appendShiftDirection(final Boolean b) {
        shiftDirection.add(b);
        return this;
      }

      public Builder appendShiftStamp(final Integer b) {
        shiftStamp.add(b);
        return this;
      }

      public Builder setStamp(final int stamp) {
        this.stamp = stamp;
        return this;
      }

      public ShfTrace build() {
        return new ShfTrace(
            new Trace(
                acc1,
                acc2,
                acc3,
                acc4,
                acc5,
                arg1Hi,
                arg1Lo,
                arg2Hi,
                arg2Lo,
                bits,
                bit1,
                bit2,
                bit3,
                bit4,
                bitB3,
                bitB4,
                bitB5,
                bitB6,
                bitB7,
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
                ones,
                oneLineInstruction,
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
                shiftStamp),
            stamp);
      }
    }
  }
}
