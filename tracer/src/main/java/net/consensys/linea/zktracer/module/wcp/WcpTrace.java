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
package net.consensys.linea.zktracer.module.wcp;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({"Trace", "Stamp"})
@SuppressWarnings("unused")
public record WcpTrace(@JsonProperty("Trace") Trace trace, @JsonProperty("Stamp") int stamp) {
  @JsonPropertyOrder({
    "ACC_1",
    "ACC_2",
    "ACC_3",
    "ACC_4",
    "ACC_5",
    "ACC_6",
    "ARGUMENT_1_HI",
    "ARGUMENT_1_LO",
    "ARGUMENT_2_HI",
    "ARGUMENT_2_LO",
    "BITS",
    "BIT_1",
    "BIT_2",
    "BIT_3",
    "BIT_4",
    "BYTE_1",
    "BYTE_2",
    "BYTE_3",
    "BYTE_4",
    "BYTE_5",
    "BYTE_6",
    "COUNTER",
    "INST",
    "NEG_1",
    "NEG_2",
    "ONE_LINE_INSTRUCTION",
    "RESULT_HI",
    "RESULT_LO",
    "WORD_COMPARISON_STAMP"
  })
  public record Trace(
      @JsonProperty("ACC_1") List<BigInteger> ACC_1,
      @JsonProperty("ACC_2") List<BigInteger> ACC_2,
      @JsonProperty("ACC_3") List<BigInteger> ACC_3,
      @JsonProperty("ACC_4") List<BigInteger> ACC_4,
      @JsonProperty("ACC_5") List<BigInteger> ACC_5,
      @JsonProperty("ACC_6") List<BigInteger> ACC_6,
      @JsonProperty("ARGUMENT_1_HI") List<BigInteger> ARGUMENT_1_HI,
      @JsonProperty("ARGUMENT_1_LO") List<BigInteger> ARGUMENT_1_LO,
      @JsonProperty("ARGUMENT_2_HI") List<BigInteger> ARGUMENT_2_HI,
      @JsonProperty("ARGUMENT_2_LO") List<BigInteger> ARGUMENT_2_LO,
      @JsonProperty("BITS") List<Boolean> BITS,
      @JsonProperty("BIT_1") List<Boolean> BIT_1,
      @JsonProperty("BIT_2") List<Boolean> BIT_2,
      @JsonProperty("BIT_3") List<Boolean> BIT_3,
      @JsonProperty("BIT_4") List<Boolean> BIT_4,
      @JsonProperty("BYTE_1") List<UnsignedByte> BYTE_1,
      @JsonProperty("BYTE_2") List<UnsignedByte> BYTE_2,
      @JsonProperty("BYTE_3") List<UnsignedByte> BYTE_3,
      @JsonProperty("BYTE_4") List<UnsignedByte> BYTE_4,
      @JsonProperty("BYTE_5") List<UnsignedByte> BYTE_5,
      @JsonProperty("BYTE_6") List<UnsignedByte> BYTE_6,
      @JsonProperty("COUNTER") List<Integer> COUNTER,
      @JsonProperty("INST") List<UnsignedByte> INST,
      @JsonProperty("NEG_1") List<Boolean> NEG_1,
      @JsonProperty("NEG_2") List<Boolean> NEG_2,
      @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> ONE_LINE_INSTRUCTION,
      @JsonProperty("RESULT_HI") List<Boolean> RESULT_HI,
      @JsonProperty("RESULT_LO") List<Boolean> RESULT_LO,
      @JsonProperty("WORD_COMPARISON_STAMP") List<Integer> WORD_COMPARISON_STAMP) {

    public static class Builder {
      private final List<BigInteger> acc1 = new ArrayList<>();
      private final List<BigInteger> acc2 = new ArrayList<>();
      private final List<BigInteger> acc3 = new ArrayList<>();
      private final List<BigInteger> acc4 = new ArrayList<>();
      private final List<BigInteger> acc5 = new ArrayList<>();
      private final List<BigInteger> acc6 = new ArrayList<>();
      private final List<BigInteger> arg1Hi = new ArrayList<>();
      private final List<BigInteger> arg1Lo = new ArrayList<>();
      private final List<BigInteger> arg2Hi = new ArrayList<>();
      private final List<BigInteger> arg2Lo = new ArrayList<>();
      private final List<Boolean> bits = new ArrayList<>();
      private final List<Boolean> bit1 = new ArrayList<>();
      private final List<Boolean> bit2 = new ArrayList<>();
      private final List<Boolean> bit3 = new ArrayList<>();
      private final List<Boolean> bit4 = new ArrayList<>();
      private final List<UnsignedByte> byte1 = new ArrayList<>();
      private final List<UnsignedByte> byte2 = new ArrayList<>();
      private final List<UnsignedByte> byte3 = new ArrayList<>();
      private final List<UnsignedByte> byte4 = new ArrayList<>();
      private final List<UnsignedByte> byte5 = new ArrayList<>();
      private final List<UnsignedByte> byte6 = new ArrayList<>();
      private final List<Integer> counter = new ArrayList<>();
      private final List<UnsignedByte> inst = new ArrayList<>();
      private final List<Boolean> neg1 = new ArrayList<>();
      private final List<Boolean> neg2 = new ArrayList<>();
      private final List<Boolean> oneLineInstruction = new ArrayList<>();
      private final List<Boolean> resHi = new ArrayList<>();
      private final List<Boolean> resLo = new ArrayList<>();
      private final List<Integer> wcpStamp = new ArrayList<>();
      private int stamp = 0;

      private Builder() {}

      public static Builder newInstance() {
        return new Builder();
      }

      public Builder appendOneLineInstruction(final Boolean b) {
        oneLineInstruction.add(b);
        return this;
      }

      public Builder appendInst(final UnsignedByte b) {
        inst.add(b);
        return this;
      }

      public Builder appendCounter(final Integer b) {
        counter.add(b);
        return this;
      }

      public Builder appendWcpStamp(final Integer b) {
        wcpStamp.add(b);
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

      public Builder appendAcc6(final BigInteger b) {
        acc6.add(b);
        return this;
      }

      public Builder appendNeg1(final Boolean b) {
        neg1.add(b);
        return this;
      }

      public Builder appendNeg2(final Boolean b) {
        neg2.add(b);
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

      public Builder appendByte6(final UnsignedByte b) {
        byte6.add(b);
        return this;
      }

      public Builder appendResHi(final Boolean b) {
        resHi.add(b);
        return this;
      }

      public Builder appendResLo(final Boolean b) {
        resLo.add(b);
        return this;
      }

      public Builder appendBits(final Boolean b) {
        bits.add(b);
        return this;
      }

      public Builder setStamp(final int stamp) {
        this.stamp = stamp;
        return this;
      }

      public WcpTrace build() {
        return new WcpTrace(
            new Trace(
                acc1,
                acc2,
                acc3,
                acc4,
                acc5,
                acc6,
                arg1Hi,
                arg1Lo,
                arg2Hi,
                arg2Lo,
                bits,
                bit1,
                bit2,
                bit3,
                bit4,
                byte1,
                byte2,
                byte3,
                byte4,
                byte5,
                byte6,
                counter,
                inst,
                neg1,
                neg2,
                oneLineInstruction,
                resHi,
                resLo,
                wcpStamp),
            stamp);
      }
    }
  }
}
