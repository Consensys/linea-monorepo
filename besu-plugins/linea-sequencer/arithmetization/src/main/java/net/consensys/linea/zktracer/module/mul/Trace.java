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

package net.consensys.linea.zktracer.module.mul;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import net.consensys.linea.zktracer.types.UnsignedByte;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public record Trace(
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
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.accA0.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ACC_A_0")
    private final List<BigInteger> accA0;

    @JsonProperty("ACC_A_1")
    private final List<BigInteger> accA1;

    @JsonProperty("ACC_A_2")
    private final List<BigInteger> accA2;

    @JsonProperty("ACC_A_3")
    private final List<BigInteger> accA3;

    @JsonProperty("ACC_B_0")
    private final List<BigInteger> accB0;

    @JsonProperty("ACC_B_1")
    private final List<BigInteger> accB1;

    @JsonProperty("ACC_B_2")
    private final List<BigInteger> accB2;

    @JsonProperty("ACC_B_3")
    private final List<BigInteger> accB3;

    @JsonProperty("ACC_C_0")
    private final List<BigInteger> accC0;

    @JsonProperty("ACC_C_1")
    private final List<BigInteger> accC1;

    @JsonProperty("ACC_C_2")
    private final List<BigInteger> accC2;

    @JsonProperty("ACC_C_3")
    private final List<BigInteger> accC3;

    @JsonProperty("ACC_H_0")
    private final List<BigInteger> accH0;

    @JsonProperty("ACC_H_1")
    private final List<BigInteger> accH1;

    @JsonProperty("ACC_H_2")
    private final List<BigInteger> accH2;

    @JsonProperty("ACC_H_3")
    private final List<BigInteger> accH3;

    @JsonProperty("ARG_1_HI")
    private final List<BigInteger> arg1Hi;

    @JsonProperty("ARG_1_LO")
    private final List<BigInteger> arg1Lo;

    @JsonProperty("ARG_2_HI")
    private final List<BigInteger> arg2Hi;

    @JsonProperty("ARG_2_LO")
    private final List<BigInteger> arg2Lo;

    @JsonProperty("BIT_NUM")
    private final List<BigInteger> bitNum;

    @JsonProperty("BITS")
    private final List<Boolean> bits;

    @JsonProperty("BYTE_A_0")
    private final List<UnsignedByte> byteA0;

    @JsonProperty("BYTE_A_1")
    private final List<UnsignedByte> byteA1;

    @JsonProperty("BYTE_A_2")
    private final List<UnsignedByte> byteA2;

    @JsonProperty("BYTE_A_3")
    private final List<UnsignedByte> byteA3;

    @JsonProperty("BYTE_B_0")
    private final List<UnsignedByte> byteB0;

    @JsonProperty("BYTE_B_1")
    private final List<UnsignedByte> byteB1;

    @JsonProperty("BYTE_B_2")
    private final List<UnsignedByte> byteB2;

    @JsonProperty("BYTE_B_3")
    private final List<UnsignedByte> byteB3;

    @JsonProperty("BYTE_C_0")
    private final List<UnsignedByte> byteC0;

    @JsonProperty("BYTE_C_1")
    private final List<UnsignedByte> byteC1;

    @JsonProperty("BYTE_C_2")
    private final List<UnsignedByte> byteC2;

    @JsonProperty("BYTE_C_3")
    private final List<UnsignedByte> byteC3;

    @JsonProperty("BYTE_H_0")
    private final List<UnsignedByte> byteH0;

    @JsonProperty("BYTE_H_1")
    private final List<UnsignedByte> byteH1;

    @JsonProperty("BYTE_H_2")
    private final List<UnsignedByte> byteH2;

    @JsonProperty("BYTE_H_3")
    private final List<UnsignedByte> byteH3;

    @JsonProperty("COUNTER")
    private final List<BigInteger> counter;

    @JsonProperty("EXPONENT_BIT")
    private final List<Boolean> exponentBit;

    @JsonProperty("EXPONENT_BIT_ACCUMULATOR")
    private final List<BigInteger> exponentBitAccumulator;

    @JsonProperty("EXPONENT_BIT_SOURCE")
    private final List<Boolean> exponentBitSource;

    @JsonProperty("INSTRUCTION")
    private final List<BigInteger> instruction;

    @JsonProperty("MUL_STAMP")
    private final List<BigInteger> mulStamp;

    @JsonProperty("OLI")
    private final List<Boolean> oli;

    @JsonProperty("RES_HI")
    private final List<BigInteger> resHi;

    @JsonProperty("RES_LO")
    private final List<BigInteger> resLo;

    @JsonProperty("RESULT_VANISHES")
    private final List<Boolean> resultVanishes;

    @JsonProperty("SQUARE_AND_MULTIPLY")
    private final List<Boolean> squareAndMultiply;

    @JsonProperty("TINY_BASE")
    private final List<Boolean> tinyBase;

    @JsonProperty("TINY_EXPONENT")
    private final List<Boolean> tinyExponent;

    private TraceBuilder(int length) {
      this.accA0 = new ArrayList<>(length);
      this.accA1 = new ArrayList<>(length);
      this.accA2 = new ArrayList<>(length);
      this.accA3 = new ArrayList<>(length);
      this.accB0 = new ArrayList<>(length);
      this.accB1 = new ArrayList<>(length);
      this.accB2 = new ArrayList<>(length);
      this.accB3 = new ArrayList<>(length);
      this.accC0 = new ArrayList<>(length);
      this.accC1 = new ArrayList<>(length);
      this.accC2 = new ArrayList<>(length);
      this.accC3 = new ArrayList<>(length);
      this.accH0 = new ArrayList<>(length);
      this.accH1 = new ArrayList<>(length);
      this.accH2 = new ArrayList<>(length);
      this.accH3 = new ArrayList<>(length);
      this.arg1Hi = new ArrayList<>(length);
      this.arg1Lo = new ArrayList<>(length);
      this.arg2Hi = new ArrayList<>(length);
      this.arg2Lo = new ArrayList<>(length);
      this.bitNum = new ArrayList<>(length);
      this.bits = new ArrayList<>(length);
      this.byteA0 = new ArrayList<>(length);
      this.byteA1 = new ArrayList<>(length);
      this.byteA2 = new ArrayList<>(length);
      this.byteA3 = new ArrayList<>(length);
      this.byteB0 = new ArrayList<>(length);
      this.byteB1 = new ArrayList<>(length);
      this.byteB2 = new ArrayList<>(length);
      this.byteB3 = new ArrayList<>(length);
      this.byteC0 = new ArrayList<>(length);
      this.byteC1 = new ArrayList<>(length);
      this.byteC2 = new ArrayList<>(length);
      this.byteC3 = new ArrayList<>(length);
      this.byteH0 = new ArrayList<>(length);
      this.byteH1 = new ArrayList<>(length);
      this.byteH2 = new ArrayList<>(length);
      this.byteH3 = new ArrayList<>(length);
      this.counter = new ArrayList<>(length);
      this.exponentBit = new ArrayList<>(length);
      this.exponentBitAccumulator = new ArrayList<>(length);
      this.exponentBitSource = new ArrayList<>(length);
      this.instruction = new ArrayList<>(length);
      this.mulStamp = new ArrayList<>(length);
      this.oli = new ArrayList<>(length);
      this.resHi = new ArrayList<>(length);
      this.resLo = new ArrayList<>(length);
      this.resultVanishes = new ArrayList<>(length);
      this.squareAndMultiply = new ArrayList<>(length);
      this.tinyBase = new ArrayList<>(length);
      this.tinyExponent = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.accA0.size();
    }

    public TraceBuilder accA0(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC_A_0 already set");
      } else {
        filled.set(0);
      }

      accA0.add(b);

      return this;
    }

    public TraceBuilder accA1(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_A_1 already set");
      } else {
        filled.set(1);
      }

      accA1.add(b);

      return this;
    }

    public TraceBuilder accA2(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ACC_A_2 already set");
      } else {
        filled.set(2);
      }

      accA2.add(b);

      return this;
    }

    public TraceBuilder accA3(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_A_3 already set");
      } else {
        filled.set(3);
      }

      accA3.add(b);

      return this;
    }

    public TraceBuilder accB0(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_B_0 already set");
      } else {
        filled.set(4);
      }

      accB0.add(b);

      return this;
    }

    public TraceBuilder accB1(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("ACC_B_1 already set");
      } else {
        filled.set(5);
      }

      accB1.add(b);

      return this;
    }

    public TraceBuilder accB2(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ACC_B_2 already set");
      } else {
        filled.set(6);
      }

      accB2.add(b);

      return this;
    }

    public TraceBuilder accB3(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ACC_B_3 already set");
      } else {
        filled.set(7);
      }

      accB3.add(b);

      return this;
    }

    public TraceBuilder accC0(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ACC_C_0 already set");
      } else {
        filled.set(8);
      }

      accC0.add(b);

      return this;
    }

    public TraceBuilder accC1(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ACC_C_1 already set");
      } else {
        filled.set(9);
      }

      accC1.add(b);

      return this;
    }

    public TraceBuilder accC2(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("ACC_C_2 already set");
      } else {
        filled.set(10);
      }

      accC2.add(b);

      return this;
    }

    public TraceBuilder accC3(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("ACC_C_3 already set");
      } else {
        filled.set(11);
      }

      accC3.add(b);

      return this;
    }

    public TraceBuilder accH0(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("ACC_H_0 already set");
      } else {
        filled.set(12);
      }

      accH0.add(b);

      return this;
    }

    public TraceBuilder accH1(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("ACC_H_1 already set");
      } else {
        filled.set(13);
      }

      accH1.add(b);

      return this;
    }

    public TraceBuilder accH2(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("ACC_H_2 already set");
      } else {
        filled.set(14);
      }

      accH2.add(b);

      return this;
    }

    public TraceBuilder accH3(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("ACC_H_3 already set");
      } else {
        filled.set(15);
      }

      accH3.add(b);

      return this;
    }

    public TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(16);
      }

      arg1Hi.add(b);

      return this;
    }

    public TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(17);
      }

      arg1Lo.add(b);

      return this;
    }

    public TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(18)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(18);
      }

      arg2Hi.add(b);

      return this;
    }

    public TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(19)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(19);
      }

      arg2Lo.add(b);

      return this;
    }

    public TraceBuilder bitNum(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("BIT_NUM already set");
      } else {
        filled.set(21);
      }

      bitNum.add(b);

      return this;
    }

    public TraceBuilder bits(final Boolean b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BITS already set");
      } else {
        filled.set(20);
      }

      bits.add(b);

      return this;
    }

    public TraceBuilder byteA0(final UnsignedByte b) {
      if (filled.get(22)) {
        throw new IllegalStateException("BYTE_A_0 already set");
      } else {
        filled.set(22);
      }

      byteA0.add(b);

      return this;
    }

    public TraceBuilder byteA1(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BYTE_A_1 already set");
      } else {
        filled.set(23);
      }

      byteA1.add(b);

      return this;
    }

    public TraceBuilder byteA2(final UnsignedByte b) {
      if (filled.get(24)) {
        throw new IllegalStateException("BYTE_A_2 already set");
      } else {
        filled.set(24);
      }

      byteA2.add(b);

      return this;
    }

    public TraceBuilder byteA3(final UnsignedByte b) {
      if (filled.get(25)) {
        throw new IllegalStateException("BYTE_A_3 already set");
      } else {
        filled.set(25);
      }

      byteA3.add(b);

      return this;
    }

    public TraceBuilder byteB0(final UnsignedByte b) {
      if (filled.get(26)) {
        throw new IllegalStateException("BYTE_B_0 already set");
      } else {
        filled.set(26);
      }

      byteB0.add(b);

      return this;
    }

    public TraceBuilder byteB1(final UnsignedByte b) {
      if (filled.get(27)) {
        throw new IllegalStateException("BYTE_B_1 already set");
      } else {
        filled.set(27);
      }

      byteB1.add(b);

      return this;
    }

    public TraceBuilder byteB2(final UnsignedByte b) {
      if (filled.get(28)) {
        throw new IllegalStateException("BYTE_B_2 already set");
      } else {
        filled.set(28);
      }

      byteB2.add(b);

      return this;
    }

    public TraceBuilder byteB3(final UnsignedByte b) {
      if (filled.get(29)) {
        throw new IllegalStateException("BYTE_B_3 already set");
      } else {
        filled.set(29);
      }

      byteB3.add(b);

      return this;
    }

    public TraceBuilder byteC0(final UnsignedByte b) {
      if (filled.get(30)) {
        throw new IllegalStateException("BYTE_C_0 already set");
      } else {
        filled.set(30);
      }

      byteC0.add(b);

      return this;
    }

    public TraceBuilder byteC1(final UnsignedByte b) {
      if (filled.get(31)) {
        throw new IllegalStateException("BYTE_C_1 already set");
      } else {
        filled.set(31);
      }

      byteC1.add(b);

      return this;
    }

    public TraceBuilder byteC2(final UnsignedByte b) {
      if (filled.get(32)) {
        throw new IllegalStateException("BYTE_C_2 already set");
      } else {
        filled.set(32);
      }

      byteC2.add(b);

      return this;
    }

    public TraceBuilder byteC3(final UnsignedByte b) {
      if (filled.get(33)) {
        throw new IllegalStateException("BYTE_C_3 already set");
      } else {
        filled.set(33);
      }

      byteC3.add(b);

      return this;
    }

    public TraceBuilder byteH0(final UnsignedByte b) {
      if (filled.get(34)) {
        throw new IllegalStateException("BYTE_H_0 already set");
      } else {
        filled.set(34);
      }

      byteH0.add(b);

      return this;
    }

    public TraceBuilder byteH1(final UnsignedByte b) {
      if (filled.get(35)) {
        throw new IllegalStateException("BYTE_H_1 already set");
      } else {
        filled.set(35);
      }

      byteH1.add(b);

      return this;
    }

    public TraceBuilder byteH2(final UnsignedByte b) {
      if (filled.get(36)) {
        throw new IllegalStateException("BYTE_H_2 already set");
      } else {
        filled.set(36);
      }

      byteH2.add(b);

      return this;
    }

    public TraceBuilder byteH3(final UnsignedByte b) {
      if (filled.get(37)) {
        throw new IllegalStateException("BYTE_H_3 already set");
      } else {
        filled.set(37);
      }

      byteH3.add(b);

      return this;
    }

    public TraceBuilder counter(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(38);
      }

      counter.add(b);

      return this;
    }

    public TraceBuilder exponentBit(final Boolean b) {
      if (filled.get(39)) {
        throw new IllegalStateException("EXPONENT_BIT already set");
      } else {
        filled.set(39);
      }

      exponentBit.add(b);

      return this;
    }

    public TraceBuilder exponentBitAccumulator(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("EXPONENT_BIT_ACCUMULATOR already set");
      } else {
        filled.set(40);
      }

      exponentBitAccumulator.add(b);

      return this;
    }

    public TraceBuilder exponentBitSource(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("EXPONENT_BIT_SOURCE already set");
      } else {
        filled.set(41);
      }

      exponentBitSource.add(b);

      return this;
    }

    public TraceBuilder instruction(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("INSTRUCTION already set");
      } else {
        filled.set(42);
      }

      instruction.add(b);

      return this;
    }

    public TraceBuilder mulStamp(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("MUL_STAMP already set");
      } else {
        filled.set(43);
      }

      mulStamp.add(b);

      return this;
    }

    public TraceBuilder oli(final Boolean b) {
      if (filled.get(44)) {
        throw new IllegalStateException("OLI already set");
      } else {
        filled.set(44);
      }

      oli.add(b);

      return this;
    }

    public TraceBuilder resHi(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(46);
      }

      resHi.add(b);

      return this;
    }

    public TraceBuilder resLo(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(47);
      }

      resLo.add(b);

      return this;
    }

    public TraceBuilder resultVanishes(final Boolean b) {
      if (filled.get(45)) {
        throw new IllegalStateException("RESULT_VANISHES already set");
      } else {
        filled.set(45);
      }

      resultVanishes.add(b);

      return this;
    }

    public TraceBuilder squareAndMultiply(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("SQUARE_AND_MULTIPLY already set");
      } else {
        filled.set(48);
      }

      squareAndMultiply.add(b);

      return this;
    }

    public TraceBuilder tinyBase(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("TINY_BASE already set");
      } else {
        filled.set(49);
      }

      tinyBase.add(b);

      return this;
    }

    public TraceBuilder tinyExponent(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("TINY_EXPONENT already set");
      } else {
        filled.set(50);
      }

      tinyExponent.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ACC_A_0 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_A_1 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ACC_A_2 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_A_3 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_B_0 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("ACC_B_1 has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("ACC_B_2 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ACC_B_3 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ACC_C_0 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ACC_C_1 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("ACC_C_2 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("ACC_C_3 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("ACC_H_0 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("ACC_H_1 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("ACC_H_2 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("ACC_H_3 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("BIT_NUM has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BITS has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("BYTE_A_0 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BYTE_A_1 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("BYTE_A_2 has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("BYTE_A_3 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("BYTE_B_0 has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("BYTE_B_1 has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("BYTE_B_2 has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("BYTE_B_3 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("BYTE_C_0 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("BYTE_C_1 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("BYTE_C_2 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("BYTE_C_3 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("BYTE_H_0 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("BYTE_H_1 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("BYTE_H_2 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("BYTE_H_3 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("EXPONENT_BIT has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("EXPONENT_BIT_ACCUMULATOR has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("EXPONENT_BIT_SOURCE has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("INSTRUCTION has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("MUL_STAMP has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("OLI has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("RESULT_VANISHES has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("SQUARE_AND_MULTIPLY has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("TINY_BASE has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("TINY_EXPONENT has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        accA0.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        accA1.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        accA2.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        accA3.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        accB0.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        accB1.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        accB2.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        accB3.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        accC0.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        accC1.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        accC2.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        accC3.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        accH0.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        accH1.add(BigInteger.ZERO);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        accH2.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        accH3.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        arg1Hi.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        arg1Lo.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        arg2Hi.add(BigInteger.ZERO);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        arg2Lo.add(BigInteger.ZERO);
        this.filled.set(19);
      }
      if (!filled.get(21)) {
        bitNum.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(20)) {
        bits.add(false);
        this.filled.set(20);
      }
      if (!filled.get(22)) {
        byteA0.add(UnsignedByte.of(0));
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        byteA1.add(UnsignedByte.of(0));
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        byteA2.add(UnsignedByte.of(0));
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        byteA3.add(UnsignedByte.of(0));
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        byteB0.add(UnsignedByte.of(0));
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        byteB1.add(UnsignedByte.of(0));
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        byteB2.add(UnsignedByte.of(0));
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        byteB3.add(UnsignedByte.of(0));
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        byteC0.add(UnsignedByte.of(0));
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        byteC1.add(UnsignedByte.of(0));
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        byteC2.add(UnsignedByte.of(0));
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        byteC3.add(UnsignedByte.of(0));
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        byteH0.add(UnsignedByte.of(0));
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        byteH1.add(UnsignedByte.of(0));
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        byteH2.add(UnsignedByte.of(0));
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        byteH3.add(UnsignedByte.of(0));
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        counter.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        exponentBit.add(false);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        exponentBitAccumulator.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        exponentBitSource.add(false);
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        instruction.add(BigInteger.ZERO);
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        mulStamp.add(BigInteger.ZERO);
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        oli.add(false);
        this.filled.set(44);
      }
      if (!filled.get(46)) {
        resHi.add(BigInteger.ZERO);
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        resLo.add(BigInteger.ZERO);
        this.filled.set(47);
      }
      if (!filled.get(45)) {
        resultVanishes.add(false);
        this.filled.set(45);
      }
      if (!filled.get(48)) {
        squareAndMultiply.add(false);
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        tinyBase.add(false);
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        tinyExponent.add(false);
        this.filled.set(50);
      }

      return this.validateRow();
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
