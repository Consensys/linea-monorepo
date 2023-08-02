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

package net.consensys.linea.zktracer.module.mul;

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
    @JsonProperty("ACC_A_0") List<BigInteger> accA0,
    @JsonProperty("ACC_A_1") List<BigInteger> accA1,
    @JsonProperty("ACC_A_2") List<BigInteger> accA2,
    @JsonProperty("ACC_A_3") List<BigInteger> accA3,
    @JsonProperty("ACC_B_0") List<BigInteger> accB0,
    @JsonProperty("ACC_B_1") List<BigInteger> accB1,
    @JsonProperty("ACC_B_2") List<BigInteger> accB2,
    @JsonProperty("ACC_B_3") List<BigInteger> accB3,
    @JsonProperty("ACC_C_0") List<BigInteger> accC0,
    @JsonProperty("ACC_C_1") List<BigInteger> accC1,
    @JsonProperty("ACC_C_2") List<BigInteger> accC2,
    @JsonProperty("ACC_C_3") List<BigInteger> accC3,
    @JsonProperty("ACC_H_0") List<BigInteger> accH0,
    @JsonProperty("ACC_H_1") List<BigInteger> accH1,
    @JsonProperty("ACC_H_2") List<BigInteger> accH2,
    @JsonProperty("ACC_H_3") List<BigInteger> accH3,
    @JsonProperty("ARG_1_HI") List<BigInteger> arg1Hi,
    @JsonProperty("ARG_1_LO") List<BigInteger> arg1Lo,
    @JsonProperty("ARG_2_HI") List<BigInteger> arg2Hi,
    @JsonProperty("ARG_2_LO") List<BigInteger> arg2Lo,
    @JsonProperty("BIT_NUM") List<BigInteger> bitNum,
    @JsonProperty("BITS") List<Boolean> bits,
    @JsonProperty("BYTE_A_0") List<UnsignedByte> byteA0,
    @JsonProperty("BYTE_A_1") List<UnsignedByte> byteA1,
    @JsonProperty("BYTE_A_2") List<UnsignedByte> byteA2,
    @JsonProperty("BYTE_A_3") List<UnsignedByte> byteA3,
    @JsonProperty("BYTE_B_0") List<UnsignedByte> byteB0,
    @JsonProperty("BYTE_B_1") List<UnsignedByte> byteB1,
    @JsonProperty("BYTE_B_2") List<UnsignedByte> byteB2,
    @JsonProperty("BYTE_B_3") List<UnsignedByte> byteB3,
    @JsonProperty("BYTE_C_0") List<UnsignedByte> byteC0,
    @JsonProperty("BYTE_C_1") List<UnsignedByte> byteC1,
    @JsonProperty("BYTE_C_2") List<UnsignedByte> byteC2,
    @JsonProperty("BYTE_C_3") List<UnsignedByte> byteC3,
    @JsonProperty("BYTE_H_0") List<UnsignedByte> byteH0,
    @JsonProperty("BYTE_H_1") List<UnsignedByte> byteH1,
    @JsonProperty("BYTE_H_2") List<UnsignedByte> byteH2,
    @JsonProperty("BYTE_H_3") List<UnsignedByte> byteH3,
    @JsonProperty("COUNTER") List<BigInteger> counter,
    @JsonProperty("EXPONENT_BIT") List<Boolean> exponentBit,
    @JsonProperty("EXPONENT_BIT_ACCUMULATOR") List<BigInteger> exponentBitAccumulator,
    @JsonProperty("EXPONENT_BIT_SOURCE") List<Boolean> exponentBitSource,
    @JsonProperty("INSTRUCTION") List<BigInteger> instruction,
    @JsonProperty("MUL_STAMP") List<BigInteger> mulStamp,
    @JsonProperty("OLI") List<Boolean> oli,
    @JsonProperty("RES_HI") List<BigInteger> resHi,
    @JsonProperty("RES_LO") List<BigInteger> resLo,
    @JsonProperty("RESULT_VANISHES") List<Boolean> resultVanishes,
    @JsonProperty("SQUARE_AND_MULTIPLY") List<Boolean> squareAndMultiply,
    @JsonProperty("TINY_BASE") List<Boolean> tinyBase,
    @JsonProperty("TINY_EXPONENT") List<Boolean> tinyExponent) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    private final List<BigInteger> accA0 = new ArrayList<>();
    private final List<BigInteger> accA1 = new ArrayList<>();
    private final List<BigInteger> accA2 = new ArrayList<>();
    private final List<BigInteger> accA3 = new ArrayList<>();
    private final List<BigInteger> accB0 = new ArrayList<>();
    private final List<BigInteger> accB1 = new ArrayList<>();
    private final List<BigInteger> accB2 = new ArrayList<>();
    private final List<BigInteger> accB3 = new ArrayList<>();
    private final List<BigInteger> accC0 = new ArrayList<>();
    private final List<BigInteger> accC1 = new ArrayList<>();
    private final List<BigInteger> accC2 = new ArrayList<>();
    private final List<BigInteger> accC3 = new ArrayList<>();
    private final List<BigInteger> accH0 = new ArrayList<>();
    private final List<BigInteger> accH1 = new ArrayList<>();
    private final List<BigInteger> accH2 = new ArrayList<>();
    private final List<BigInteger> accH3 = new ArrayList<>();
    private final List<BigInteger> arg1Hi = new ArrayList<>();
    private final List<BigInteger> arg1Lo = new ArrayList<>();
    private final List<BigInteger> arg2Hi = new ArrayList<>();
    private final List<BigInteger> arg2Lo = new ArrayList<>();
    private final List<BigInteger> bitNum = new ArrayList<>();
    private final List<Boolean> bits = new ArrayList<>();
    private final List<UnsignedByte> byteA0 = new ArrayList<>();
    private final List<UnsignedByte> byteA1 = new ArrayList<>();
    private final List<UnsignedByte> byteA2 = new ArrayList<>();
    private final List<UnsignedByte> byteA3 = new ArrayList<>();
    private final List<UnsignedByte> byteB0 = new ArrayList<>();
    private final List<UnsignedByte> byteB1 = new ArrayList<>();
    private final List<UnsignedByte> byteB2 = new ArrayList<>();
    private final List<UnsignedByte> byteB3 = new ArrayList<>();
    private final List<UnsignedByte> byteC0 = new ArrayList<>();
    private final List<UnsignedByte> byteC1 = new ArrayList<>();
    private final List<UnsignedByte> byteC2 = new ArrayList<>();
    private final List<UnsignedByte> byteC3 = new ArrayList<>();
    private final List<UnsignedByte> byteH0 = new ArrayList<>();
    private final List<UnsignedByte> byteH1 = new ArrayList<>();
    private final List<UnsignedByte> byteH2 = new ArrayList<>();
    private final List<UnsignedByte> byteH3 = new ArrayList<>();
    private final List<BigInteger> counter = new ArrayList<>();
    private final List<Boolean> exponentBit = new ArrayList<>();
    private final List<BigInteger> exponentBitAccumulator = new ArrayList<>();
    private final List<Boolean> exponentBitSource = new ArrayList<>();
    private final List<BigInteger> instruction = new ArrayList<>();
    private final List<BigInteger> mulStamp = new ArrayList<>();
    private final List<Boolean> oli = new ArrayList<>();
    private final List<BigInteger> resHi = new ArrayList<>();
    private final List<BigInteger> resLo = new ArrayList<>();
    private final List<Boolean> resultVanishes = new ArrayList<>();
    private final List<Boolean> squareAndMultiply = new ArrayList<>();
    private final List<Boolean> tinyBase = new ArrayList<>();
    private final List<Boolean> tinyExponent = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder accA0(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_A_0 already set");
      } else {
        filled.set(3);
      }

      accA0.add(b);

      return this;
    }

    TraceBuilder accA1(final BigInteger b) {
      if (filled.get(41)) {
        throw new IllegalStateException("ACC_A_1 already set");
      } else {
        filled.set(41);
      }

      accA1.add(b);

      return this;
    }

    TraceBuilder accA2(final BigInteger b) {
      if (filled.get(34)) {
        throw new IllegalStateException("ACC_A_2 already set");
      } else {
        filled.set(34);
      }

      accA2.add(b);

      return this;
    }

    TraceBuilder accA3(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("ACC_A_3 already set");
      } else {
        filled.set(45);
      }

      accA3.add(b);

      return this;
    }

    TraceBuilder accB0(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("ACC_B_0 already set");
      } else {
        filled.set(29);
      }

      accB0.add(b);

      return this;
    }

    TraceBuilder accB1(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("ACC_B_1 already set");
      } else {
        filled.set(36);
      }

      accB1.add(b);

      return this;
    }

    TraceBuilder accB2(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("ACC_B_2 already set");
      } else {
        filled.set(49);
      }

      accB2.add(b);

      return this;
    }

    TraceBuilder accB3(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("ACC_B_3 already set");
      } else {
        filled.set(37);
      }

      accB3.add(b);

      return this;
    }

    TraceBuilder accC0(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ACC_C_0 already set");
      } else {
        filled.set(9);
      }

      accC0.add(b);

      return this;
    }

    TraceBuilder accC1(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("ACC_C_1 already set");
      } else {
        filled.set(21);
      }

      accC1.add(b);

      return this;
    }

    TraceBuilder accC2(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("ACC_C_2 already set");
      } else {
        filled.set(48);
      }

      accC2.add(b);

      return this;
    }

    TraceBuilder accC3(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("ACC_C_3 already set");
      } else {
        filled.set(5);
      }

      accC3.add(b);

      return this;
    }

    TraceBuilder accH0(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("ACC_H_0 already set");
      } else {
        filled.set(38);
      }

      accH0.add(b);

      return this;
    }

    TraceBuilder accH1(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("ACC_H_1 already set");
      } else {
        filled.set(30);
      }

      accH1.add(b);

      return this;
    }

    TraceBuilder accH2(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ACC_H_2 already set");
      } else {
        filled.set(8);
      }

      accH2.add(b);

      return this;
    }

    TraceBuilder accH3(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("ACC_H_3 already set");
      } else {
        filled.set(47);
      }

      accH3.add(b);

      return this;
    }

    TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(50);
      }

      arg1Hi.add(b);

      return this;
    }

    TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(28);
      }

      arg1Lo.add(b);

      return this;
    }

    TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(0);
      }

      arg2Hi.add(b);

      return this;
    }

    TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(20);
      }

      arg2Lo.add(b);

      return this;
    }

    TraceBuilder bits(final Boolean b) {
      if (filled.get(44)) {
        throw new IllegalStateException("BITS already set");
      } else {
        filled.set(44);
      }

      bits.add(b);

      return this;
    }

    TraceBuilder bitNum(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("BIT_NUM already set");
      } else {
        filled.set(13);
      }

      bitNum.add(b);

      return this;
    }

    TraceBuilder byteA0(final UnsignedByte b) {
      if (filled.get(42)) {
        throw new IllegalStateException("BYTE_A_0 already set");
      } else {
        filled.set(42);
      }

      byteA0.add(b);

      return this;
    }

    TraceBuilder byteA1(final UnsignedByte b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BYTE_A_1 already set");
      } else {
        filled.set(10);
      }

      byteA1.add(b);

      return this;
    }

    TraceBuilder byteA2(final UnsignedByte b) {
      if (filled.get(7)) {
        throw new IllegalStateException("BYTE_A_2 already set");
      } else {
        filled.set(7);
      }

      byteA2.add(b);

      return this;
    }

    TraceBuilder byteA3(final UnsignedByte b) {
      if (filled.get(39)) {
        throw new IllegalStateException("BYTE_A_3 already set");
      } else {
        filled.set(39);
      }

      byteA3.add(b);

      return this;
    }

    TraceBuilder byteB0(final UnsignedByte b) {
      if (filled.get(40)) {
        throw new IllegalStateException("BYTE_B_0 already set");
      } else {
        filled.set(40);
      }

      byteB0.add(b);

      return this;
    }

    TraceBuilder byteB1(final UnsignedByte b) {
      if (filled.get(4)) {
        throw new IllegalStateException("BYTE_B_1 already set");
      } else {
        filled.set(4);
      }

      byteB1.add(b);

      return this;
    }

    TraceBuilder byteB2(final UnsignedByte b) {
      if (filled.get(6)) {
        throw new IllegalStateException("BYTE_B_2 already set");
      } else {
        filled.set(6);
      }

      byteB2.add(b);

      return this;
    }

    TraceBuilder byteB3(final UnsignedByte b) {
      if (filled.get(14)) {
        throw new IllegalStateException("BYTE_B_3 already set");
      } else {
        filled.set(14);
      }

      byteB3.add(b);

      return this;
    }

    TraceBuilder byteC0(final UnsignedByte b) {
      if (filled.get(35)) {
        throw new IllegalStateException("BYTE_C_0 already set");
      } else {
        filled.set(35);
      }

      byteC0.add(b);

      return this;
    }

    TraceBuilder byteC1(final UnsignedByte b) {
      if (filled.get(2)) {
        throw new IllegalStateException("BYTE_C_1 already set");
      } else {
        filled.set(2);
      }

      byteC1.add(b);

      return this;
    }

    TraceBuilder byteC2(final UnsignedByte b) {
      if (filled.get(32)) {
        throw new IllegalStateException("BYTE_C_2 already set");
      } else {
        filled.set(32);
      }

      byteC2.add(b);

      return this;
    }

    TraceBuilder byteC3(final UnsignedByte b) {
      if (filled.get(1)) {
        throw new IllegalStateException("BYTE_C_3 already set");
      } else {
        filled.set(1);
      }

      byteC3.add(b);

      return this;
    }

    TraceBuilder byteH0(final UnsignedByte b) {
      if (filled.get(19)) {
        throw new IllegalStateException("BYTE_H_0 already set");
      } else {
        filled.set(19);
      }

      byteH0.add(b);

      return this;
    }

    TraceBuilder byteH1(final UnsignedByte b) {
      if (filled.get(17)) {
        throw new IllegalStateException("BYTE_H_1 already set");
      } else {
        filled.set(17);
      }

      byteH1.add(b);

      return this;
    }

    TraceBuilder byteH2(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BYTE_H_2 already set");
      } else {
        filled.set(23);
      }

      byteH2.add(b);

      return this;
    }

    TraceBuilder byteH3(final UnsignedByte b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BYTE_H_3 already set");
      } else {
        filled.set(15);
      }

      byteH3.add(b);

      return this;
    }

    TraceBuilder counter(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(11);
      }

      counter.add(b);

      return this;
    }

    TraceBuilder exponentBit(final Boolean b) {
      if (filled.get(46)) {
        throw new IllegalStateException("EXPONENT_BIT already set");
      } else {
        filled.set(46);
      }

      exponentBit.add(b);

      return this;
    }

    TraceBuilder exponentBitAccumulator(final BigInteger b) {
      if (filled.get(31)) {
        throw new IllegalStateException("EXPONENT_BIT_ACCUMULATOR already set");
      } else {
        filled.set(31);
      }

      exponentBitAccumulator.add(b);

      return this;
    }

    TraceBuilder exponentBitSource(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("EXPONENT_BIT_SOURCE already set");
      } else {
        filled.set(24);
      }

      exponentBitSource.add(b);

      return this;
    }

    TraceBuilder instruction(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("INSTRUCTION already set");
      } else {
        filled.set(27);
      }

      instruction.add(b);

      return this;
    }

    TraceBuilder mulStamp(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("MUL_STAMP already set");
      } else {
        filled.set(16);
      }

      mulStamp.add(b);

      return this;
    }

    TraceBuilder oli(final Boolean b) {
      if (filled.get(26)) {
        throw new IllegalStateException("OLI already set");
      } else {
        filled.set(26);
      }

      oli.add(b);

      return this;
    }

    TraceBuilder resultVanishes(final Boolean b) {
      if (filled.get(22)) {
        throw new IllegalStateException("RESULT_VANISHES already set");
      } else {
        filled.set(22);
      }

      resultVanishes.add(b);

      return this;
    }

    TraceBuilder resHi(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(33);
      }

      resHi.add(b);

      return this;
    }

    TraceBuilder resLo(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(43);
      }

      resLo.add(b);

      return this;
    }

    TraceBuilder squareAndMultiply(final Boolean b) {
      if (filled.get(25)) {
        throw new IllegalStateException("SQUARE_AND_MULTIPLY already set");
      } else {
        filled.set(25);
      }

      squareAndMultiply.add(b);

      return this;
    }

    TraceBuilder tinyBase(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("TINY_BASE already set");
      } else {
        filled.set(12);
      }

      tinyBase.add(b);

      return this;
    }

    TraceBuilder tinyExponent(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("TINY_EXPONENT already set");
      } else {
        filled.set(18);
      }

      tinyExponent.add(b);

      return this;
    }

    TraceBuilder setAccA0At(final BigInteger b, int i) {
      accA0.set(i, b);

      return this;
    }

    TraceBuilder setAccA1At(final BigInteger b, int i) {
      accA1.set(i, b);

      return this;
    }

    TraceBuilder setAccA2At(final BigInteger b, int i) {
      accA2.set(i, b);

      return this;
    }

    TraceBuilder setAccA3At(final BigInteger b, int i) {
      accA3.set(i, b);

      return this;
    }

    TraceBuilder setAccB0At(final BigInteger b, int i) {
      accB0.set(i, b);

      return this;
    }

    TraceBuilder setAccB1At(final BigInteger b, int i) {
      accB1.set(i, b);

      return this;
    }

    TraceBuilder setAccB2At(final BigInteger b, int i) {
      accB2.set(i, b);

      return this;
    }

    TraceBuilder setAccB3At(final BigInteger b, int i) {
      accB3.set(i, b);

      return this;
    }

    TraceBuilder setAccC0At(final BigInteger b, int i) {
      accC0.set(i, b);

      return this;
    }

    TraceBuilder setAccC1At(final BigInteger b, int i) {
      accC1.set(i, b);

      return this;
    }

    TraceBuilder setAccC2At(final BigInteger b, int i) {
      accC2.set(i, b);

      return this;
    }

    TraceBuilder setAccC3At(final BigInteger b, int i) {
      accC3.set(i, b);

      return this;
    }

    TraceBuilder setAccH0At(final BigInteger b, int i) {
      accH0.set(i, b);

      return this;
    }

    TraceBuilder setAccH1At(final BigInteger b, int i) {
      accH1.set(i, b);

      return this;
    }

    TraceBuilder setAccH2At(final BigInteger b, int i) {
      accH2.set(i, b);

      return this;
    }

    TraceBuilder setAccH3At(final BigInteger b, int i) {
      accH3.set(i, b);

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

    TraceBuilder setBitNumAt(final BigInteger b, int i) {
      bitNum.set(i, b);

      return this;
    }

    TraceBuilder setByteA0At(final UnsignedByte b, int i) {
      byteA0.set(i, b);

      return this;
    }

    TraceBuilder setByteA1At(final UnsignedByte b, int i) {
      byteA1.set(i, b);

      return this;
    }

    TraceBuilder setByteA2At(final UnsignedByte b, int i) {
      byteA2.set(i, b);

      return this;
    }

    TraceBuilder setByteA3At(final UnsignedByte b, int i) {
      byteA3.set(i, b);

      return this;
    }

    TraceBuilder setByteB0At(final UnsignedByte b, int i) {
      byteB0.set(i, b);

      return this;
    }

    TraceBuilder setByteB1At(final UnsignedByte b, int i) {
      byteB1.set(i, b);

      return this;
    }

    TraceBuilder setByteB2At(final UnsignedByte b, int i) {
      byteB2.set(i, b);

      return this;
    }

    TraceBuilder setByteB3At(final UnsignedByte b, int i) {
      byteB3.set(i, b);

      return this;
    }

    TraceBuilder setByteC0At(final UnsignedByte b, int i) {
      byteC0.set(i, b);

      return this;
    }

    TraceBuilder setByteC1At(final UnsignedByte b, int i) {
      byteC1.set(i, b);

      return this;
    }

    TraceBuilder setByteC2At(final UnsignedByte b, int i) {
      byteC2.set(i, b);

      return this;
    }

    TraceBuilder setByteC3At(final UnsignedByte b, int i) {
      byteC3.set(i, b);

      return this;
    }

    TraceBuilder setByteH0At(final UnsignedByte b, int i) {
      byteH0.set(i, b);

      return this;
    }

    TraceBuilder setByteH1At(final UnsignedByte b, int i) {
      byteH1.set(i, b);

      return this;
    }

    TraceBuilder setByteH2At(final UnsignedByte b, int i) {
      byteH2.set(i, b);

      return this;
    }

    TraceBuilder setByteH3At(final UnsignedByte b, int i) {
      byteH3.set(i, b);

      return this;
    }

    TraceBuilder setCounterAt(final BigInteger b, int i) {
      counter.set(i, b);

      return this;
    }

    TraceBuilder setExponentBitAt(final Boolean b, int i) {
      exponentBit.set(i, b);

      return this;
    }

    TraceBuilder setExponentBitAccumulatorAt(final BigInteger b, int i) {
      exponentBitAccumulator.set(i, b);

      return this;
    }

    TraceBuilder setExponentBitSourceAt(final Boolean b, int i) {
      exponentBitSource.set(i, b);

      return this;
    }

    TraceBuilder setInstructionAt(final BigInteger b, int i) {
      instruction.set(i, b);

      return this;
    }

    TraceBuilder setMulStampAt(final BigInteger b, int i) {
      mulStamp.set(i, b);

      return this;
    }

    TraceBuilder setOliAt(final Boolean b, int i) {
      oli.set(i, b);

      return this;
    }

    TraceBuilder setResultVanishesAt(final Boolean b, int i) {
      resultVanishes.set(i, b);

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

    TraceBuilder setSquareAndMultiplyAt(final Boolean b, int i) {
      squareAndMultiply.set(i, b);

      return this;
    }

    TraceBuilder setTinyBaseAt(final Boolean b, int i) {
      tinyBase.set(i, b);

      return this;
    }

    TraceBuilder setTinyExponentAt(final Boolean b, int i) {
      tinyExponent.set(i, b);

      return this;
    }

    TraceBuilder setAccA0Relative(final BigInteger b, int i) {
      accA0.set(accA0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccA1Relative(final BigInteger b, int i) {
      accA1.set(accA1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccA2Relative(final BigInteger b, int i) {
      accA2.set(accA2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccA3Relative(final BigInteger b, int i) {
      accA3.set(accA3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccB0Relative(final BigInteger b, int i) {
      accB0.set(accB0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccB1Relative(final BigInteger b, int i) {
      accB1.set(accB1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccB2Relative(final BigInteger b, int i) {
      accB2.set(accB2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccB3Relative(final BigInteger b, int i) {
      accB3.set(accB3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccC0Relative(final BigInteger b, int i) {
      accC0.set(accC0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccC1Relative(final BigInteger b, int i) {
      accC1.set(accC1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccC2Relative(final BigInteger b, int i) {
      accC2.set(accC2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccC3Relative(final BigInteger b, int i) {
      accC3.set(accC3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccH0Relative(final BigInteger b, int i) {
      accH0.set(accH0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccH1Relative(final BigInteger b, int i) {
      accH1.set(accH1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccH2Relative(final BigInteger b, int i) {
      accH2.set(accH2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccH3Relative(final BigInteger b, int i) {
      accH3.set(accH3.size() - 1 - i, b);

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

    TraceBuilder setBitNumRelative(final BigInteger b, int i) {
      bitNum.set(bitNum.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteA0Relative(final UnsignedByte b, int i) {
      byteA0.set(byteA0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteA1Relative(final UnsignedByte b, int i) {
      byteA1.set(byteA1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteA2Relative(final UnsignedByte b, int i) {
      byteA2.set(byteA2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteA3Relative(final UnsignedByte b, int i) {
      byteA3.set(byteA3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteB0Relative(final UnsignedByte b, int i) {
      byteB0.set(byteB0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteB1Relative(final UnsignedByte b, int i) {
      byteB1.set(byteB1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteB2Relative(final UnsignedByte b, int i) {
      byteB2.set(byteB2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteB3Relative(final UnsignedByte b, int i) {
      byteB3.set(byteB3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteC0Relative(final UnsignedByte b, int i) {
      byteC0.set(byteC0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteC1Relative(final UnsignedByte b, int i) {
      byteC1.set(byteC1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteC2Relative(final UnsignedByte b, int i) {
      byteC2.set(byteC2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteC3Relative(final UnsignedByte b, int i) {
      byteC3.set(byteC3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteH0Relative(final UnsignedByte b, int i) {
      byteH0.set(byteH0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteH1Relative(final UnsignedByte b, int i) {
      byteH1.set(byteH1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteH2Relative(final UnsignedByte b, int i) {
      byteH2.set(byteH2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteH3Relative(final UnsignedByte b, int i) {
      byteH3.set(byteH3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterRelative(final BigInteger b, int i) {
      counter.set(counter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setExponentBitRelative(final Boolean b, int i) {
      exponentBit.set(exponentBit.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setExponentBitAccumulatorRelative(final BigInteger b, int i) {
      exponentBitAccumulator.set(exponentBitAccumulator.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setExponentBitSourceRelative(final Boolean b, int i) {
      exponentBitSource.set(exponentBitSource.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInstructionRelative(final BigInteger b, int i) {
      instruction.set(instruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setMulStampRelative(final BigInteger b, int i) {
      mulStamp.set(mulStamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOliRelative(final Boolean b, int i) {
      oli.set(oli.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResultVanishesRelative(final Boolean b, int i) {
      resultVanishes.set(resultVanishes.size() - 1 - i, b);

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

    TraceBuilder setSquareAndMultiplyRelative(final Boolean b, int i) {
      squareAndMultiply.set(squareAndMultiply.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTinyBaseRelative(final Boolean b, int i) {
      tinyBase.set(tinyBase.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTinyExponentRelative(final Boolean b, int i) {
      tinyExponent.set(tinyExponent.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_A_0 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("ACC_A_1 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("ACC_A_2 has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("ACC_A_3 has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("ACC_B_0 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("ACC_B_1 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("ACC_B_2 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("ACC_B_3 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ACC_C_0 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("ACC_C_1 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("ACC_C_2 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("ACC_C_3 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("ACC_H_0 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("ACC_H_1 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ACC_H_2 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("ACC_H_3 has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("BIT_NUM has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("BITS has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("BYTE_A_0 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BYTE_A_1 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("BYTE_A_2 has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("BYTE_A_3 has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("BYTE_B_0 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("BYTE_B_1 has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("BYTE_B_2 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("BYTE_B_3 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("BYTE_C_0 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("BYTE_C_1 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("BYTE_C_2 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("BYTE_C_3 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("BYTE_H_0 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("BYTE_H_1 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BYTE_H_2 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BYTE_H_3 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("EXPONENT_BIT has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("EXPONENT_BIT_ACCUMULATOR has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("EXPONENT_BIT_SOURCE has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("INSTRUCTION has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("MUL_STAMP has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("OLI has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("RESULT_VANISHES has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("SQUARE_AND_MULTIPLY has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("TINY_BASE has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("TINY_EXPONENT has not been filled");
      }

      filled.clear();

      return this;
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
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
          bitNum,
          bits,
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
          instruction,
          mulStamp,
          oli,
          resHi,
          resLo,
          resultVanishes,
          squareAndMultiply,
          tinyBase,
          tinyExponent);
    }
  }
}
