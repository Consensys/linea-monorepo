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

package net.consensys.linea.zktracer.module.mod;

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
    @JsonProperty("ACC_1_2") List<BigInteger> acc12,
    @JsonProperty("ACC_1_3") List<BigInteger> acc13,
    @JsonProperty("ACC_2_2") List<BigInteger> acc22,
    @JsonProperty("ACC_2_3") List<BigInteger> acc23,
    @JsonProperty("ACC_B_0") List<BigInteger> accB0,
    @JsonProperty("ACC_B_1") List<BigInteger> accB1,
    @JsonProperty("ACC_B_2") List<BigInteger> accB2,
    @JsonProperty("ACC_B_3") List<BigInteger> accB3,
    @JsonProperty("ACC_DELTA_0") List<BigInteger> accDelta0,
    @JsonProperty("ACC_DELTA_1") List<BigInteger> accDelta1,
    @JsonProperty("ACC_DELTA_2") List<BigInteger> accDelta2,
    @JsonProperty("ACC_DELTA_3") List<BigInteger> accDelta3,
    @JsonProperty("ACC_H_0") List<BigInteger> accH0,
    @JsonProperty("ACC_H_1") List<BigInteger> accH1,
    @JsonProperty("ACC_H_2") List<BigInteger> accH2,
    @JsonProperty("ACC_Q_0") List<BigInteger> accQ0,
    @JsonProperty("ACC_Q_1") List<BigInteger> accQ1,
    @JsonProperty("ACC_Q_2") List<BigInteger> accQ2,
    @JsonProperty("ACC_Q_3") List<BigInteger> accQ3,
    @JsonProperty("ACC_R_0") List<BigInteger> accR0,
    @JsonProperty("ACC_R_1") List<BigInteger> accR1,
    @JsonProperty("ACC_R_2") List<BigInteger> accR2,
    @JsonProperty("ACC_R_3") List<BigInteger> accR3,
    @JsonProperty("ARG_1_HI") List<BigInteger> arg1Hi,
    @JsonProperty("ARG_1_LO") List<BigInteger> arg1Lo,
    @JsonProperty("ARG_2_HI") List<BigInteger> arg2Hi,
    @JsonProperty("ARG_2_LO") List<BigInteger> arg2Lo,
    @JsonProperty("BYTE_1_2") List<UnsignedByte> byte12,
    @JsonProperty("BYTE_1_3") List<UnsignedByte> byte13,
    @JsonProperty("BYTE_2_2") List<UnsignedByte> byte22,
    @JsonProperty("BYTE_2_3") List<UnsignedByte> byte23,
    @JsonProperty("BYTE_B_0") List<UnsignedByte> byteB0,
    @JsonProperty("BYTE_B_1") List<UnsignedByte> byteB1,
    @JsonProperty("BYTE_B_2") List<UnsignedByte> byteB2,
    @JsonProperty("BYTE_B_3") List<UnsignedByte> byteB3,
    @JsonProperty("BYTE_DELTA_0") List<UnsignedByte> byteDelta0,
    @JsonProperty("BYTE_DELTA_1") List<UnsignedByte> byteDelta1,
    @JsonProperty("BYTE_DELTA_2") List<UnsignedByte> byteDelta2,
    @JsonProperty("BYTE_DELTA_3") List<UnsignedByte> byteDelta3,
    @JsonProperty("BYTE_H_0") List<UnsignedByte> byteH0,
    @JsonProperty("BYTE_H_1") List<UnsignedByte> byteH1,
    @JsonProperty("BYTE_H_2") List<UnsignedByte> byteH2,
    @JsonProperty("BYTE_Q_0") List<UnsignedByte> byteQ0,
    @JsonProperty("BYTE_Q_1") List<UnsignedByte> byteQ1,
    @JsonProperty("BYTE_Q_2") List<UnsignedByte> byteQ2,
    @JsonProperty("BYTE_Q_3") List<UnsignedByte> byteQ3,
    @JsonProperty("BYTE_R_0") List<UnsignedByte> byteR0,
    @JsonProperty("BYTE_R_1") List<UnsignedByte> byteR1,
    @JsonProperty("BYTE_R_2") List<UnsignedByte> byteR2,
    @JsonProperty("BYTE_R_3") List<UnsignedByte> byteR3,
    @JsonProperty("CMP_1") List<Boolean> cmp1,
    @JsonProperty("CMP_2") List<Boolean> cmp2,
    @JsonProperty("CT") List<BigInteger> ct,
    @JsonProperty("DEC_OUTPUT") List<Boolean> decOutput,
    @JsonProperty("DEC_SIGNED") List<Boolean> decSigned,
    @JsonProperty("INST") List<BigInteger> inst,
    @JsonProperty("MSB_1") List<Boolean> msb1,
    @JsonProperty("MSB_2") List<Boolean> msb2,
    @JsonProperty("OLI") List<Boolean> oli,
    @JsonProperty("RES_HI") List<BigInteger> resHi,
    @JsonProperty("RES_LO") List<BigInteger> resLo,
    @JsonProperty("STAMP") List<BigInteger> stamp) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    private final List<BigInteger> acc12 = new ArrayList<>();
    private final List<BigInteger> acc13 = new ArrayList<>();
    private final List<BigInteger> acc22 = new ArrayList<>();
    private final List<BigInteger> acc23 = new ArrayList<>();
    private final List<BigInteger> accB0 = new ArrayList<>();
    private final List<BigInteger> accB1 = new ArrayList<>();
    private final List<BigInteger> accB2 = new ArrayList<>();
    private final List<BigInteger> accB3 = new ArrayList<>();
    private final List<BigInteger> accDelta0 = new ArrayList<>();
    private final List<BigInteger> accDelta1 = new ArrayList<>();
    private final List<BigInteger> accDelta2 = new ArrayList<>();
    private final List<BigInteger> accDelta3 = new ArrayList<>();
    private final List<BigInteger> accH0 = new ArrayList<>();
    private final List<BigInteger> accH1 = new ArrayList<>();
    private final List<BigInteger> accH2 = new ArrayList<>();
    private final List<BigInteger> accQ0 = new ArrayList<>();
    private final List<BigInteger> accQ1 = new ArrayList<>();
    private final List<BigInteger> accQ2 = new ArrayList<>();
    private final List<BigInteger> accQ3 = new ArrayList<>();
    private final List<BigInteger> accR0 = new ArrayList<>();
    private final List<BigInteger> accR1 = new ArrayList<>();
    private final List<BigInteger> accR2 = new ArrayList<>();
    private final List<BigInteger> accR3 = new ArrayList<>();
    private final List<BigInteger> arg1Hi = new ArrayList<>();
    private final List<BigInteger> arg1Lo = new ArrayList<>();
    private final List<BigInteger> arg2Hi = new ArrayList<>();
    private final List<BigInteger> arg2Lo = new ArrayList<>();
    private final List<UnsignedByte> byte12 = new ArrayList<>();
    private final List<UnsignedByte> byte13 = new ArrayList<>();
    private final List<UnsignedByte> byte22 = new ArrayList<>();
    private final List<UnsignedByte> byte23 = new ArrayList<>();
    private final List<UnsignedByte> byteB0 = new ArrayList<>();
    private final List<UnsignedByte> byteB1 = new ArrayList<>();
    private final List<UnsignedByte> byteB2 = new ArrayList<>();
    private final List<UnsignedByte> byteB3 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta0 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta1 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta2 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta3 = new ArrayList<>();
    private final List<UnsignedByte> byteH0 = new ArrayList<>();
    private final List<UnsignedByte> byteH1 = new ArrayList<>();
    private final List<UnsignedByte> byteH2 = new ArrayList<>();
    private final List<UnsignedByte> byteQ0 = new ArrayList<>();
    private final List<UnsignedByte> byteQ1 = new ArrayList<>();
    private final List<UnsignedByte> byteQ2 = new ArrayList<>();
    private final List<UnsignedByte> byteQ3 = new ArrayList<>();
    private final List<UnsignedByte> byteR0 = new ArrayList<>();
    private final List<UnsignedByte> byteR1 = new ArrayList<>();
    private final List<UnsignedByte> byteR2 = new ArrayList<>();
    private final List<UnsignedByte> byteR3 = new ArrayList<>();
    private final List<Boolean> cmp1 = new ArrayList<>();
    private final List<Boolean> cmp2 = new ArrayList<>();
    private final List<BigInteger> ct = new ArrayList<>();
    private final List<Boolean> decOutput = new ArrayList<>();
    private final List<Boolean> decSigned = new ArrayList<>();
    private final List<BigInteger> inst = new ArrayList<>();
    private final List<Boolean> msb1 = new ArrayList<>();
    private final List<Boolean> msb2 = new ArrayList<>();
    private final List<Boolean> oli = new ArrayList<>();
    private final List<BigInteger> resHi = new ArrayList<>();
    private final List<BigInteger> resLo = new ArrayList<>();
    private final List<BigInteger> stamp = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder acc12(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("ACC_1_2 already set");
      } else {
        filled.set(47);
      }

      acc12.add(b);

      return this;
    }

    TraceBuilder acc13(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_1_3 already set");
      } else {
        filled.set(1);
      }

      acc13.add(b);

      return this;
    }

    TraceBuilder acc22(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("ACC_2_2 already set");
      } else {
        filled.set(30);
      }

      acc22.add(b);

      return this;
    }

    TraceBuilder acc23(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("ACC_2_3 already set");
      } else {
        filled.set(43);
      }

      acc23.add(b);

      return this;
    }

    TraceBuilder accB0(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("ACC_B_0 already set");
      } else {
        filled.set(16);
      }

      accB0.add(b);

      return this;
    }

    TraceBuilder accB1(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("ACC_B_1 already set");
      } else {
        filled.set(39);
      }

      accB1.add(b);

      return this;
    }

    TraceBuilder accB2(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("ACC_B_2 already set");
      } else {
        filled.set(23);
      }

      accB2.add(b);

      return this;
    }

    TraceBuilder accB3(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ACC_B_3 already set");
      } else {
        filled.set(7);
      }

      accB3.add(b);

      return this;
    }

    TraceBuilder accDelta0(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("ACC_DELTA_0 already set");
      } else {
        filled.set(54);
      }

      accDelta0.add(b);

      return this;
    }

    TraceBuilder accDelta1(final BigInteger b) {
      if (filled.get(60)) {
        throw new IllegalStateException("ACC_DELTA_1 already set");
      } else {
        filled.set(60);
      }

      accDelta1.add(b);

      return this;
    }

    TraceBuilder accDelta2(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("ACC_DELTA_2 already set");
      } else {
        filled.set(36);
      }

      accDelta2.add(b);

      return this;
    }

    TraceBuilder accDelta3(final BigInteger b) {
      if (filled.get(61)) {
        throw new IllegalStateException("ACC_DELTA_3 already set");
      } else {
        filled.set(61);
      }

      accDelta3.add(b);

      return this;
    }

    TraceBuilder accH0(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("ACC_H_0 already set");
      } else {
        filled.set(17);
      }

      accH0.add(b);

      return this;
    }

    TraceBuilder accH1(final BigInteger b) {
      if (filled.get(41)) {
        throw new IllegalStateException("ACC_H_1 already set");
      } else {
        filled.set(41);
      }

      accH1.add(b);

      return this;
    }

    TraceBuilder accH2(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("ACC_H_2 already set");
      } else {
        filled.set(55);
      }

      accH2.add(b);

      return this;
    }

    TraceBuilder accQ0(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("ACC_Q_0 already set");
      } else {
        filled.set(26);
      }

      accQ0.add(b);

      return this;
    }

    TraceBuilder accQ1(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("ACC_Q_1 already set");
      } else {
        filled.set(57);
      }

      accQ1.add(b);

      return this;
    }

    TraceBuilder accQ2(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("ACC_Q_2 already set");
      } else {
        filled.set(21);
      }

      accQ2.add(b);

      return this;
    }

    TraceBuilder accQ3(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC_Q_3 already set");
      } else {
        filled.set(0);
      }

      accQ3.add(b);

      return this;
    }

    TraceBuilder accR0(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("ACC_R_0 already set");
      } else {
        filled.set(13);
      }

      accR0.add(b);

      return this;
    }

    TraceBuilder accR1(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("ACC_R_1 already set");
      } else {
        filled.set(40);
      }

      accR1.add(b);

      return this;
    }

    TraceBuilder accR2(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("ACC_R_2 already set");
      } else {
        filled.set(37);
      }

      accR2.add(b);

      return this;
    }

    TraceBuilder accR3(final BigInteger b) {
      if (filled.get(58)) {
        throw new IllegalStateException("ACC_R_3 already set");
      } else {
        filled.set(58);
      }

      accR3.add(b);

      return this;
    }

    TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(34)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(34);
      }

      arg1Hi.add(b);

      return this;
    }

    TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(35)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(35);
      }

      arg1Lo.add(b);

      return this;
    }

    TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(3);
      }

      arg2Hi.add(b);

      return this;
    }

    TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(32)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(32);
      }

      arg2Lo.add(b);

      return this;
    }

    TraceBuilder byte12(final UnsignedByte b) {
      if (filled.get(14)) {
        throw new IllegalStateException("BYTE_1_2 already set");
      } else {
        filled.set(14);
      }

      byte12.add(b);

      return this;
    }

    TraceBuilder byte13(final UnsignedByte b) {
      if (filled.get(11)) {
        throw new IllegalStateException("BYTE_1_3 already set");
      } else {
        filled.set(11);
      }

      byte13.add(b);

      return this;
    }

    TraceBuilder byte22(final UnsignedByte b) {
      if (filled.get(2)) {
        throw new IllegalStateException("BYTE_2_2 already set");
      } else {
        filled.set(2);
      }

      byte22.add(b);

      return this;
    }

    TraceBuilder byte23(final UnsignedByte b) {
      if (filled.get(33)) {
        throw new IllegalStateException("BYTE_2_3 already set");
      } else {
        filled.set(33);
      }

      byte23.add(b);

      return this;
    }

    TraceBuilder byteB0(final UnsignedByte b) {
      if (filled.get(27)) {
        throw new IllegalStateException("BYTE_B_0 already set");
      } else {
        filled.set(27);
      }

      byteB0.add(b);

      return this;
    }

    TraceBuilder byteB1(final UnsignedByte b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BYTE_B_1 already set");
      } else {
        filled.set(20);
      }

      byteB1.add(b);

      return this;
    }

    TraceBuilder byteB2(final UnsignedByte b) {
      if (filled.get(8)) {
        throw new IllegalStateException("BYTE_B_2 already set");
      } else {
        filled.set(8);
      }

      byteB2.add(b);

      return this;
    }

    TraceBuilder byteB3(final UnsignedByte b) {
      if (filled.get(12)) {
        throw new IllegalStateException("BYTE_B_3 already set");
      } else {
        filled.set(12);
      }

      byteB3.add(b);

      return this;
    }

    TraceBuilder byteDelta0(final UnsignedByte b) {
      if (filled.get(22)) {
        throw new IllegalStateException("BYTE_DELTA_0 already set");
      } else {
        filled.set(22);
      }

      byteDelta0.add(b);

      return this;
    }

    TraceBuilder byteDelta1(final UnsignedByte b) {
      if (filled.get(38)) {
        throw new IllegalStateException("BYTE_DELTA_1 already set");
      } else {
        filled.set(38);
      }

      byteDelta1.add(b);

      return this;
    }

    TraceBuilder byteDelta2(final UnsignedByte b) {
      if (filled.get(45)) {
        throw new IllegalStateException("BYTE_DELTA_2 already set");
      } else {
        filled.set(45);
      }

      byteDelta2.add(b);

      return this;
    }

    TraceBuilder byteDelta3(final UnsignedByte b) {
      if (filled.get(53)) {
        throw new IllegalStateException("BYTE_DELTA_3 already set");
      } else {
        filled.set(53);
      }

      byteDelta3.add(b);

      return this;
    }

    TraceBuilder byteH0(final UnsignedByte b) {
      if (filled.get(28)) {
        throw new IllegalStateException("BYTE_H_0 already set");
      } else {
        filled.set(28);
      }

      byteH0.add(b);

      return this;
    }

    TraceBuilder byteH1(final UnsignedByte b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BYTE_H_1 already set");
      } else {
        filled.set(15);
      }

      byteH1.add(b);

      return this;
    }

    TraceBuilder byteH2(final UnsignedByte b) {
      if (filled.get(5)) {
        throw new IllegalStateException("BYTE_H_2 already set");
      } else {
        filled.set(5);
      }

      byteH2.add(b);

      return this;
    }

    TraceBuilder byteQ0(final UnsignedByte b) {
      if (filled.get(4)) {
        throw new IllegalStateException("BYTE_Q_0 already set");
      } else {
        filled.set(4);
      }

      byteQ0.add(b);

      return this;
    }

    TraceBuilder byteQ1(final UnsignedByte b) {
      if (filled.get(48)) {
        throw new IllegalStateException("BYTE_Q_1 already set");
      } else {
        filled.set(48);
      }

      byteQ1.add(b);

      return this;
    }

    TraceBuilder byteQ2(final UnsignedByte b) {
      if (filled.get(44)) {
        throw new IllegalStateException("BYTE_Q_2 already set");
      } else {
        filled.set(44);
      }

      byteQ2.add(b);

      return this;
    }

    TraceBuilder byteQ3(final UnsignedByte b) {
      if (filled.get(9)) {
        throw new IllegalStateException("BYTE_Q_3 already set");
      } else {
        filled.set(9);
      }

      byteQ3.add(b);

      return this;
    }

    TraceBuilder byteR0(final UnsignedByte b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BYTE_R_0 already set");
      } else {
        filled.set(10);
      }

      byteR0.add(b);

      return this;
    }

    TraceBuilder byteR1(final UnsignedByte b) {
      if (filled.get(46)) {
        throw new IllegalStateException("BYTE_R_1 already set");
      } else {
        filled.set(46);
      }

      byteR1.add(b);

      return this;
    }

    TraceBuilder byteR2(final UnsignedByte b) {
      if (filled.get(49)) {
        throw new IllegalStateException("BYTE_R_2 already set");
      } else {
        filled.set(49);
      }

      byteR2.add(b);

      return this;
    }

    TraceBuilder byteR3(final UnsignedByte b) {
      if (filled.get(50)) {
        throw new IllegalStateException("BYTE_R_3 already set");
      } else {
        filled.set(50);
      }

      byteR3.add(b);

      return this;
    }

    TraceBuilder cmp1(final Boolean b) {
      if (filled.get(59)) {
        throw new IllegalStateException("CMP_1 already set");
      } else {
        filled.set(59);
      }

      cmp1.add(b);

      return this;
    }

    TraceBuilder cmp2(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("CMP_2 already set");
      } else {
        filled.set(31);
      }

      cmp2.add(b);

      return this;
    }

    TraceBuilder ct(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(51);
      }

      ct.add(b);

      return this;
    }

    TraceBuilder decOutput(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("DEC_OUTPUT already set");
      } else {
        filled.set(19);
      }

      decOutput.add(b);

      return this;
    }

    TraceBuilder decSigned(final Boolean b) {
      if (filled.get(25)) {
        throw new IllegalStateException("DEC_SIGNED already set");
      } else {
        filled.set(25);
      }

      decSigned.add(b);

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

    TraceBuilder msb1(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("MSB_1 already set");
      } else {
        filled.set(24);
      }

      msb1.add(b);

      return this;
    }

    TraceBuilder msb2(final Boolean b) {
      if (filled.get(42)) {
        throw new IllegalStateException("MSB_2 already set");
      } else {
        filled.set(42);
      }

      msb2.add(b);

      return this;
    }

    TraceBuilder oli(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("OLI already set");
      } else {
        filled.set(18);
      }

      oli.add(b);

      return this;
    }

    TraceBuilder resHi(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(56);
      }

      resHi.add(b);

      return this;
    }

    TraceBuilder resLo(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(52);
      }

      resLo.add(b);

      return this;
    }

    TraceBuilder stamp(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(29);
      }

      stamp.add(b);

      return this;
    }

    TraceBuilder setAcc12At(final BigInteger b, int i) {
      acc12.set(i, b);

      return this;
    }

    TraceBuilder setAcc13At(final BigInteger b, int i) {
      acc13.set(i, b);

      return this;
    }

    TraceBuilder setAcc22At(final BigInteger b, int i) {
      acc22.set(i, b);

      return this;
    }

    TraceBuilder setAcc23At(final BigInteger b, int i) {
      acc23.set(i, b);

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

    TraceBuilder setAccDelta0At(final BigInteger b, int i) {
      accDelta0.set(i, b);

      return this;
    }

    TraceBuilder setAccDelta1At(final BigInteger b, int i) {
      accDelta1.set(i, b);

      return this;
    }

    TraceBuilder setAccDelta2At(final BigInteger b, int i) {
      accDelta2.set(i, b);

      return this;
    }

    TraceBuilder setAccDelta3At(final BigInteger b, int i) {
      accDelta3.set(i, b);

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

    TraceBuilder setAccQ0At(final BigInteger b, int i) {
      accQ0.set(i, b);

      return this;
    }

    TraceBuilder setAccQ1At(final BigInteger b, int i) {
      accQ1.set(i, b);

      return this;
    }

    TraceBuilder setAccQ2At(final BigInteger b, int i) {
      accQ2.set(i, b);

      return this;
    }

    TraceBuilder setAccQ3At(final BigInteger b, int i) {
      accQ3.set(i, b);

      return this;
    }

    TraceBuilder setAccR0At(final BigInteger b, int i) {
      accR0.set(i, b);

      return this;
    }

    TraceBuilder setAccR1At(final BigInteger b, int i) {
      accR1.set(i, b);

      return this;
    }

    TraceBuilder setAccR2At(final BigInteger b, int i) {
      accR2.set(i, b);

      return this;
    }

    TraceBuilder setAccR3At(final BigInteger b, int i) {
      accR3.set(i, b);

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

    TraceBuilder setByte12At(final UnsignedByte b, int i) {
      byte12.set(i, b);

      return this;
    }

    TraceBuilder setByte13At(final UnsignedByte b, int i) {
      byte13.set(i, b);

      return this;
    }

    TraceBuilder setByte22At(final UnsignedByte b, int i) {
      byte22.set(i, b);

      return this;
    }

    TraceBuilder setByte23At(final UnsignedByte b, int i) {
      byte23.set(i, b);

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

    TraceBuilder setByteDelta0At(final UnsignedByte b, int i) {
      byteDelta0.set(i, b);

      return this;
    }

    TraceBuilder setByteDelta1At(final UnsignedByte b, int i) {
      byteDelta1.set(i, b);

      return this;
    }

    TraceBuilder setByteDelta2At(final UnsignedByte b, int i) {
      byteDelta2.set(i, b);

      return this;
    }

    TraceBuilder setByteDelta3At(final UnsignedByte b, int i) {
      byteDelta3.set(i, b);

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

    TraceBuilder setByteQ0At(final UnsignedByte b, int i) {
      byteQ0.set(i, b);

      return this;
    }

    TraceBuilder setByteQ1At(final UnsignedByte b, int i) {
      byteQ1.set(i, b);

      return this;
    }

    TraceBuilder setByteQ2At(final UnsignedByte b, int i) {
      byteQ2.set(i, b);

      return this;
    }

    TraceBuilder setByteQ3At(final UnsignedByte b, int i) {
      byteQ3.set(i, b);

      return this;
    }

    TraceBuilder setByteR0At(final UnsignedByte b, int i) {
      byteR0.set(i, b);

      return this;
    }

    TraceBuilder setByteR1At(final UnsignedByte b, int i) {
      byteR1.set(i, b);

      return this;
    }

    TraceBuilder setByteR2At(final UnsignedByte b, int i) {
      byteR2.set(i, b);

      return this;
    }

    TraceBuilder setByteR3At(final UnsignedByte b, int i) {
      byteR3.set(i, b);

      return this;
    }

    TraceBuilder setCmp1At(final Boolean b, int i) {
      cmp1.set(i, b);

      return this;
    }

    TraceBuilder setCmp2At(final Boolean b, int i) {
      cmp2.set(i, b);

      return this;
    }

    TraceBuilder setCtAt(final BigInteger b, int i) {
      ct.set(i, b);

      return this;
    }

    TraceBuilder setDecOutputAt(final Boolean b, int i) {
      decOutput.set(i, b);

      return this;
    }

    TraceBuilder setDecSignedAt(final Boolean b, int i) {
      decSigned.set(i, b);

      return this;
    }

    TraceBuilder setInstAt(final BigInteger b, int i) {
      inst.set(i, b);

      return this;
    }

    TraceBuilder setMsb1At(final Boolean b, int i) {
      msb1.set(i, b);

      return this;
    }

    TraceBuilder setMsb2At(final Boolean b, int i) {
      msb2.set(i, b);

      return this;
    }

    TraceBuilder setOliAt(final Boolean b, int i) {
      oli.set(i, b);

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

    TraceBuilder setStampAt(final BigInteger b, int i) {
      stamp.set(i, b);

      return this;
    }

    TraceBuilder setAcc12Relative(final BigInteger b, int i) {
      acc12.set(acc12.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc13Relative(final BigInteger b, int i) {
      acc13.set(acc13.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc22Relative(final BigInteger b, int i) {
      acc22.set(acc22.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc23Relative(final BigInteger b, int i) {
      acc23.set(acc23.size() - 1 - i, b);

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

    TraceBuilder setAccDelta0Relative(final BigInteger b, int i) {
      accDelta0.set(accDelta0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccDelta1Relative(final BigInteger b, int i) {
      accDelta1.set(accDelta1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccDelta2Relative(final BigInteger b, int i) {
      accDelta2.set(accDelta2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccDelta3Relative(final BigInteger b, int i) {
      accDelta3.set(accDelta3.size() - 1 - i, b);

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

    TraceBuilder setAccQ0Relative(final BigInteger b, int i) {
      accQ0.set(accQ0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccQ1Relative(final BigInteger b, int i) {
      accQ1.set(accQ1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccQ2Relative(final BigInteger b, int i) {
      accQ2.set(accQ2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccQ3Relative(final BigInteger b, int i) {
      accQ3.set(accQ3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccR0Relative(final BigInteger b, int i) {
      accR0.set(accR0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccR1Relative(final BigInteger b, int i) {
      accR1.set(accR1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccR2Relative(final BigInteger b, int i) {
      accR2.set(accR2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccR3Relative(final BigInteger b, int i) {
      accR3.set(accR3.size() - 1 - i, b);

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

    TraceBuilder setByte12Relative(final UnsignedByte b, int i) {
      byte12.set(byte12.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte13Relative(final UnsignedByte b, int i) {
      byte13.set(byte13.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte22Relative(final UnsignedByte b, int i) {
      byte22.set(byte22.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte23Relative(final UnsignedByte b, int i) {
      byte23.set(byte23.size() - 1 - i, b);

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

    TraceBuilder setByteDelta0Relative(final UnsignedByte b, int i) {
      byteDelta0.set(byteDelta0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteDelta1Relative(final UnsignedByte b, int i) {
      byteDelta1.set(byteDelta1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteDelta2Relative(final UnsignedByte b, int i) {
      byteDelta2.set(byteDelta2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteDelta3Relative(final UnsignedByte b, int i) {
      byteDelta3.set(byteDelta3.size() - 1 - i, b);

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

    TraceBuilder setByteQ0Relative(final UnsignedByte b, int i) {
      byteQ0.set(byteQ0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteQ1Relative(final UnsignedByte b, int i) {
      byteQ1.set(byteQ1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteQ2Relative(final UnsignedByte b, int i) {
      byteQ2.set(byteQ2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteQ3Relative(final UnsignedByte b, int i) {
      byteQ3.set(byteQ3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteR0Relative(final UnsignedByte b, int i) {
      byteR0.set(byteR0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteR1Relative(final UnsignedByte b, int i) {
      byteR1.set(byteR1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteR2Relative(final UnsignedByte b, int i) {
      byteR2.set(byteR2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteR3Relative(final UnsignedByte b, int i) {
      byteR3.set(byteR3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCmp1Relative(final Boolean b, int i) {
      cmp1.set(cmp1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCmp2Relative(final Boolean b, int i) {
      cmp2.set(cmp2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCtRelative(final BigInteger b, int i) {
      ct.set(ct.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDecOutputRelative(final Boolean b, int i) {
      decOutput.set(decOutput.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDecSignedRelative(final Boolean b, int i) {
      decSigned.set(decSigned.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInstRelative(final BigInteger b, int i) {
      inst.set(inst.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setMsb1Relative(final Boolean b, int i) {
      msb1.set(msb1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setMsb2Relative(final Boolean b, int i) {
      msb2.set(msb2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOliRelative(final Boolean b, int i) {
      oli.set(oli.size() - 1 - i, b);

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

    TraceBuilder setStampRelative(final BigInteger b, int i) {
      stamp.set(stamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(47)) {
        throw new IllegalStateException("ACC_1_2 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_1_3 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("ACC_2_2 has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("ACC_2_3 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("ACC_B_0 has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("ACC_B_1 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("ACC_B_2 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ACC_B_3 has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("ACC_DELTA_0 has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("ACC_DELTA_1 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("ACC_DELTA_2 has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException("ACC_DELTA_3 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("ACC_H_0 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("ACC_H_1 has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("ACC_H_2 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("ACC_Q_0 has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("ACC_Q_1 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("ACC_Q_2 has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("ACC_Q_3 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("ACC_R_0 has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("ACC_R_1 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("ACC_R_2 has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("ACC_R_3 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("BYTE_1_2 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("BYTE_1_3 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("BYTE_2_2 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("BYTE_2_3 has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("BYTE_B_0 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BYTE_B_1 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("BYTE_B_2 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("BYTE_B_3 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("BYTE_DELTA_0 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("BYTE_DELTA_1 has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("BYTE_DELTA_2 has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("BYTE_DELTA_3 has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("BYTE_H_0 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BYTE_H_1 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("BYTE_H_2 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("BYTE_Q_0 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("BYTE_Q_1 has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("BYTE_Q_2 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("BYTE_Q_3 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BYTE_R_0 has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("BYTE_R_1 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("BYTE_R_2 has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("BYTE_R_3 has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("CMP_1 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("CMP_2 has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("DEC_OUTPUT has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("DEC_SIGNED has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("MSB_1 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("MSB_2 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("OLI has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("STAMP has not been filled");
      }

      filled.clear();

      return this;
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          acc12,
          acc13,
          acc22,
          acc23,
          accB0,
          accB1,
          accB2,
          accB3,
          accDelta0,
          accDelta1,
          accDelta2,
          accDelta3,
          accH0,
          accH1,
          accH2,
          accQ0,
          accQ1,
          accQ2,
          accQ3,
          accR0,
          accR1,
          accR2,
          accR3,
          arg1Hi,
          arg1Lo,
          arg2Hi,
          arg2Lo,
          byte12,
          byte13,
          byte22,
          byte23,
          byteB0,
          byteB1,
          byteB2,
          byteB3,
          byteDelta0,
          byteDelta1,
          byteDelta2,
          byteDelta3,
          byteH0,
          byteH1,
          byteH2,
          byteQ0,
          byteQ1,
          byteQ2,
          byteQ3,
          byteR0,
          byteR1,
          byteR2,
          byteR3,
          cmp1,
          cmp2,
          ct,
          decOutput,
          decSigned,
          inst,
          msb1,
          msb2,
          oli,
          resHi,
          resLo,
          stamp);
    }
  }
}
