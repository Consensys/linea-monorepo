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
package net.consensys.linea.zktracer.module.alu.mul;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({"Trace", "Stamp"})
@SuppressWarnings("unused")
public record MulTrace(@JsonProperty("Trace") Trace trace, @JsonProperty("Stamp") int stamp) {
  @JsonPropertyOrder({
    "ACC_A_0",
    "ACC_A_1",
    "ACC_A_2",
    "ACC_A_3",
    "ACC_B_0",
    "ACC_B_1",
    "ACC_B_2",
    "ACC_B_3",
    "ACC_C_0",
    "ACC_C_1",
    "ACC_C_2",
    "ACC_C_3",
    "ACC_H_0",
    "ACC_H_1",
    "ACC_H_2",
    "ACC_H_3",
    "ARG_1_HI",
    "ARG_1_LO",
    "ARG_2_HI",
    "ARG_2_LO",
    "BITS",
    "BIT_NUM",
    "BYTE_A_0",
    "BYTE_A_1",
    "BYTE_A_2",
    "BYTE_A_3",
    "BYTE_B_0",
    "BYTE_B_1",
    "BYTE_B_2",
    "BYTE_B_3",
    "BYTE_C_0",
    "BYTE_C_1",
    "BYTE_C_2",
    "BYTE_C_3",
    "BYTE_H_0",
    "BYTE_H_1",
    "BYTE_H_2",
    "BYTE_H_3",
    "COUNTER",
    "EXPONENT_BIT",
    "EXPONENT_BIT_ACCUMULATOR",
    "EXPONENT_BIT_SOURCE",
    "INST", // INSTRUCTION
    "MUL_STAMP",
    "OLI", // "ONE_LINE_INSTRUCTION",
    "RESULT_VANISHES",
    "RES_HI",
    "RES_LO",
    "SQUARE_AND_MULTIPLY",
    "TINY_BASE",
    "TINY_EXPONENT",
  })
  @SuppressWarnings("unused")
  public record Trace(
      @JsonProperty("ACC_A_0") List<BigInteger> ACC_A_0,
      @JsonProperty("ACC_A_1") List<BigInteger> ACC_A_1,
      @JsonProperty("ACC_A_2") List<BigInteger> ACC_A_2,
      @JsonProperty("ACC_A_3") List<BigInteger> ACC_A_3,
      @JsonProperty("ACC_B_0") List<BigInteger> ACC_B_0,
      @JsonProperty("ACC_B_1") List<BigInteger> ACC_B_1,
      @JsonProperty("ACC_B_2") List<BigInteger> ACC_B_2,
      @JsonProperty("ACC_B_3") List<BigInteger> ACC_B_3,
      @JsonProperty("ACC_C_0") List<BigInteger> ACC_C_0,
      @JsonProperty("ACC_C_1") List<BigInteger> ACC_C_1,
      @JsonProperty("ACC_C_2") List<BigInteger> ACC_C_2,
      @JsonProperty("ACC_C_3") List<BigInteger> ACC_C_3,
      @JsonProperty("ACC_H_0") List<BigInteger> ACC_H_0,
      @JsonProperty("ACC_H_1") List<BigInteger> ACC_H_1,
      @JsonProperty("ACC_H_2") List<BigInteger> ACC_H_2,
      @JsonProperty("ACC_H_3") List<BigInteger> ACC_H_3,
      @JsonProperty("ARG_1_HI") List<BigInteger> ARG_1_HI,
      @JsonProperty("ARG_1_LO") List<BigInteger> ARG_1_LO,
      @JsonProperty("ARG_2_HI") List<BigInteger> ARG_2_HI,
      @JsonProperty("ARG_2_LO") List<BigInteger> ARG_2_LO,
      @JsonProperty("BITS") List<Boolean> BITS,
      @JsonProperty("BIT_NUM") List<Integer> BIT_NUM,
      @JsonProperty("BYTE_A_0") List<UnsignedByte> BYTE_A_0,
      @JsonProperty("BYTE_A_1") List<UnsignedByte> BYTE_A_1,
      @JsonProperty("BYTE_A_2") List<UnsignedByte> BYTE_A_2,
      @JsonProperty("BYTE_A_3") List<UnsignedByte> BYTE_A_3,
      @JsonProperty("BYTE_B_0") List<UnsignedByte> BYTE_B_0,
      @JsonProperty("BYTE_B_1") List<UnsignedByte> BYTE_B_1,
      @JsonProperty("BYTE_B_2") List<UnsignedByte> BYTE_B_2,
      @JsonProperty("BYTE_B_3") List<UnsignedByte> BYTE_B_3,
      @JsonProperty("BYTE_C_0") List<UnsignedByte> BYTE_C_0,
      @JsonProperty("BYTE_C_1") List<UnsignedByte> BYTE_C_1,
      @JsonProperty("BYTE_C_2") List<UnsignedByte> BYTE_C_2,
      @JsonProperty("BYTE_C_3") List<UnsignedByte> BYTE_C_3,
      @JsonProperty("BYTE_H_0") List<UnsignedByte> BYTE_H_0,
      @JsonProperty("BYTE_H_1") List<UnsignedByte> BYTE_H_1,
      @JsonProperty("BYTE_H_2") List<UnsignedByte> BYTE_H_2,
      @JsonProperty("BYTE_H_3") List<UnsignedByte> BYTE_H_3,
      @JsonProperty("COUNTER") List<Integer> COUNTER,
      @JsonProperty("EXPONENT_BIT") List<Boolean> EXPONENT_BIT,
      @JsonProperty("EXPONENT_BIT_ACCUMULATOR") List<BigInteger> EXPONENT_BIT_ACCUMULATOR,
      @JsonProperty("EXPONENT_BIT_SOURCE") List<Boolean> EXPONENT_BIT_SOURCE,
      @JsonProperty("INST") List<UnsignedByte> INST,
      @JsonProperty("MUL_STAMP") List<Integer> MUL_STAMP,
      @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> ONE_LINE_INSTRUCTION,
      @JsonProperty("RESULT_VANISHES") List<Boolean> RESULT_VANISHES,
      @JsonProperty("RES_HI") List<BigInteger> RES_HI,
      @JsonProperty("RES_LO") List<BigInteger> RES_LO,
      @JsonProperty("SQUARE_AND_MULTIPLY") List<Boolean> SQUARE_AND_MULTIPLY,
      @JsonProperty("TINY_BASE") List<Boolean> TINY_BASE,
      @JsonProperty("TINY_EXPONENT") List<Boolean> TINY_EXPONENT) {

    public static class Builder {
      private final List<Integer> mulStamp = new ArrayList<>();
      private final List<Integer> counter = new ArrayList<>();
      private final List<Boolean> oneLineInstruction = new ArrayList<>();
      private final List<Boolean> tinyBase = new ArrayList<>();
      private final List<Boolean> tinyExponent = new ArrayList<>();
      private final List<Boolean> resultVanishes = new ArrayList<>();
      private final List<UnsignedByte> inst = new ArrayList<>();
      private final List<BigInteger> arg1Hi = new ArrayList<>();
      private final List<BigInteger> arg1Lo = new ArrayList<>();
      private final List<BigInteger> arg2Hi = new ArrayList<>();
      private final List<BigInteger> arg2Lo = new ArrayList<>();
      private final List<BigInteger> resHi = new ArrayList<>();
      private final List<BigInteger> resLo = new ArrayList<>();
      private final List<Boolean> bits = new ArrayList<>();

      private final List<UnsignedByte> byteA3 = new ArrayList<>();
      private final List<UnsignedByte> byteA2 = new ArrayList<>();
      private final List<UnsignedByte> byteA1 = new ArrayList<>();
      private final List<UnsignedByte> byteA0 = new ArrayList<>();
      private final List<BigInteger> accA3 = new ArrayList<>();
      private final List<BigInteger> accA2 = new ArrayList<>();
      private final List<BigInteger> accA1 = new ArrayList<>();
      private final List<BigInteger> accA0 = new ArrayList<>();

      private final List<UnsignedByte> byteB3 = new ArrayList<>();
      private final List<UnsignedByte> byteB2 = new ArrayList<>();
      private final List<UnsignedByte> byteB1 = new ArrayList<>();
      private final List<UnsignedByte> byteB0 = new ArrayList<>();
      private final List<BigInteger> accB3 = new ArrayList<>();
      private final List<BigInteger> accB2 = new ArrayList<>();
      private final List<BigInteger> accB1 = new ArrayList<>();
      private final List<BigInteger> accB0 = new ArrayList<>();

      private final List<UnsignedByte> byteC3 = new ArrayList<>();
      private final List<UnsignedByte> byteC2 = new ArrayList<>();
      private final List<UnsignedByte> byteC1 = new ArrayList<>();
      private final List<UnsignedByte> byteC0 = new ArrayList<>();
      private final List<BigInteger> accC3 = new ArrayList<>();
      private final List<BigInteger> accC2 = new ArrayList<>();
      private final List<BigInteger> accC1 = new ArrayList<>();
      private final List<BigInteger> accC0 = new ArrayList<>();

      private final List<UnsignedByte> byteH3 = new ArrayList<>();
      private final List<UnsignedByte> byteH2 = new ArrayList<>();
      private final List<UnsignedByte> byteH1 = new ArrayList<>();
      private final List<UnsignedByte> byteH0 = new ArrayList<>();
      private final List<BigInteger> accH3 = new ArrayList<>();
      private final List<BigInteger> accH2 = new ArrayList<>();
      private final List<BigInteger> accH1 = new ArrayList<>();
      private final List<BigInteger> accH0 = new ArrayList<>();

      private final List<Boolean> exponentBit = new ArrayList<>();
      private final List<BigInteger> exponentBitAccumulator = new ArrayList<>();
      private final List<Boolean> exponentBitSource = new ArrayList<>();
      private final List<Boolean> squareAndMultiply = new ArrayList<>();
      private final List<Integer> bitNum = new ArrayList<>();
      private int stamp = 0;

      private Builder() {}

      public static Builder newInstance() {
        return new Builder();
      }

      public Builder appendAccA0(final BigInteger b) {
        accA0.add(b);
        return this;
      }

      public Builder appendAccA1(final BigInteger b) {
        accA1.add(b);
        return this;
      }

      public Builder appendAccA2(final BigInteger b) {
        accA2.add(b);
        return this;
      }

      public Builder appendAccA3(final BigInteger b) {
        accA3.add(b);
        return this;
      }

      public Builder appendAccB0(final BigInteger b) {
        accB0.add(b);
        return this;
      }

      public Builder appendAccB1(final BigInteger b) {
        accB1.add(b);
        return this;
      }

      public Builder appendAccB2(final BigInteger b) {
        accB2.add(b);
        return this;
      }

      public Builder appendAccB3(final BigInteger b) {
        accB3.add(b);
        return this;
      }

      public Builder appendAccC0(final BigInteger b) {
        accC0.add(b);
        return this;
      }

      public Builder appendAccC1(final BigInteger b) {
        accC1.add(b);
        return this;
      }

      public Builder appendAccC2(final BigInteger b) {
        accC2.add(b);
        return this;
      }

      public Builder appendAccC3(final BigInteger b) {
        accC3.add(b);
        return this;
      }

      public Builder appendAccH0(final BigInteger b) {
        accH0.add(b);
        return this;
      }

      public Builder appendAccH1(final BigInteger b) {
        accH1.add(b);
        return this;
      }

      public Builder appendAccH2(final BigInteger b) {
        accH2.add(b);
        return this;
      }

      public Builder appendAccH3(final BigInteger b) {
        accH3.add(b);
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

      public Builder appendByteA0(final UnsignedByte b) {
        byteA0.add(b);
        return this;
      }

      public Builder appendByteA1(final UnsignedByte b) {
        byteA1.add(b);
        return this;
      }

      public Builder appendByteA2(final UnsignedByte b) {
        byteA2.add(b);
        return this;
      }

      public Builder appendByteA3(final UnsignedByte b) {
        byteA3.add(b);
        return this;
      }

      public Builder appendByteB0(final UnsignedByte b) {
        byteB0.add(b);
        return this;
      }

      public Builder appendByteB1(final UnsignedByte b) {
        byteB1.add(b);
        return this;
      }

      public Builder appendByteB2(final UnsignedByte b) {
        byteB2.add(b);
        return this;
      }

      public Builder appendByteB3(final UnsignedByte b) {
        byteB3.add(b);
        return this;
      }

      public Builder appendByteC0(final UnsignedByte b) {
        byteC0.add(b);
        return this;
      }

      public Builder appendByteC1(final UnsignedByte b) {
        byteC1.add(b);
        return this;
      }

      public Builder appendByteC2(final UnsignedByte b) {
        byteC2.add(b);
        return this;
      }

      public Builder appendByteC3(final UnsignedByte b) {
        byteC3.add(b);
        return this;
      }

      public Builder appendByteH0(final UnsignedByte b) {
        byteH0.add(b);
        return this;
      }

      public Builder appendByteH1(final UnsignedByte b) {
        byteH1.add(b);
        return this;
      }

      public Builder appendByteH2(final UnsignedByte b) {
        byteH2.add(b);
        return this;
      }

      public Builder appendByteH3(final UnsignedByte b) {
        byteH3.add(b);
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

      public Builder appendOneLineInstruction(final Boolean b) {
        oneLineInstruction.add(b);
        return this;
      }

      public Builder appendTinyBase(final Boolean b) {
        tinyBase.add(b);
        return this;
      }

      public Builder appendTinyExponent(final Boolean b) {
        tinyExponent.add(b);
        return this;
      }

      public Builder appendResultVanishes(final Boolean b) {
        resultVanishes.add(b);
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

      //

      public Builder appendExponentBit(final Boolean b) {
        exponentBit.add(b);
        return this;
      }

      public Builder appendExponentBitAcc(final BigInteger b) {
        exponentBitAccumulator.add(b);
        return this;
      }

      public Builder appendExponentBitSource(final Boolean b) {
        exponentBitSource.add(b);
        return this;
      }

      public Builder appendSquareAndMultiply(final Boolean b) {
        squareAndMultiply.add(b);
        return this;
      }

      public Builder appendBitNum(final Integer b) {
        bitNum.add(b);
        return this;
      }

      public Builder appendStamp(final Integer b) {
        mulStamp.add(b);
        return this;
      }

      public Builder setStamp(final int stamp) {
        this.stamp = stamp;
        return this;
      }

      public MulTrace build() {
        return new MulTrace(
            new Trace(
                accA0,
                accA1,
                accA2,
                accA3,
                accB0,
                accB1,
                accB2,
                accB3,
                accC0,
                accC1,
                accC2,
                accC3,
                accH0,
                accH1,
                accH2,
                accH3,
                arg1Hi,
                arg1Lo,
                arg2Hi,
                arg2Lo,
                bits,
                bitNum,
                byteA0,
                byteA1,
                byteA2,
                byteA3,
                byteB0,
                byteB1,
                byteB2,
                byteB3,
                byteC0,
                byteC1,
                byteC2,
                byteC3,
                byteH0,
                byteH1,
                byteH2,
                byteH3,
                counter,
                exponentBit,
                exponentBitAccumulator,
                exponentBitSource,
                inst,
                mulStamp,
                oneLineInstruction,
                resultVanishes,
                resHi,
                resLo,
                squareAndMultiply,
                tinyBase,
                tinyExponent),
            stamp);
      }
    }
  }
}
